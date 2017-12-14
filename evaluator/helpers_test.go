package evaluator_test

import (
	"context"
	"fmt"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
)

type fakeConnectionCtx struct {
	variables *variable.Container
	info      *mongodb.Info
	server    evaluator.ServerCtx
}

func (*fakeConnectionCtx) LastInsertId() int64 {
	return 11
}
func (*fakeConnectionCtx) Logger(_ string) *log.Logger {
	lg := log.GlobalLogger()
	return &lg
}
func (*fakeConnectionCtx) RowCount() int64 {
	return 21
}
func (*fakeConnectionCtx) Catalog() *catalog.Catalog {
	return nil
}
func (*fakeConnectionCtx) UpdateCatalog(*schema.Schema) error {
	return nil
}
func (*fakeConnectionCtx) ConnectionID() uint32 {
	return 42
}
func (*fakeConnectionCtx) Context() context.Context {
	return context.Background()
}
func (*fakeConnectionCtx) DB() string {
	return "test"
}
func (*fakeConnectionCtx) GetStartupInfo() []string {
	return []string{}
}
func (*fakeConnectionCtx) Kill(id uint32, scope evaluator.KillScope) error {
	return nil
}
func (f *fakeConnectionCtx) Server() evaluator.ServerCtx {
	return f.server
}
func (*fakeConnectionCtx) Session() *mongodb.Session {
	return nil
}
func (*fakeConnectionCtx) User() string {
	return "test user"
}
func (f *fakeConnectionCtx) Variables() *variable.Container {
	if f.variables == nil {
		f.variables = variable.NewSessionContainer(variable.NewGlobalContainer(nil))
	}
	f.variables.MongoDBInfo = f.info
	return f.variables
}

// bsonDToValues takes a bson.D document and returns
// the corresponding values.
func bsonDToValues(selectID int, databaseName, tableName string, document bson.D) ([]evaluator.Value, error) {
	values := []evaluator.Value{}
	for _, v := range document {
		value, err := evaluator.NewSQLValueFromSQLColumnExpr(v.Value, schema.SQLNone, schema.MongoNone)
		if err != nil {
			return nil, err
		}
		values = append(values, evaluator.NewValue(selectID, databaseName, tableName, v.Name, value))
	}
	return values, nil
}

func createAllProjectedColumnsFromSource(selectID int, source evaluator.PlanStage, projectedTableName string) evaluator.ProjectedColumns {
	results := evaluator.ProjectedColumns{}
	for _, c := range source.Columns() {
		if c.MongoType == schema.MongoFilter {
			continue
		}
		results = append(results, createProjectedColumnFromColumn(
			selectID, c, projectedTableName, c.Name))
	}

	return results
}

func createProjectedColumnFromColumn(newSelectID int, column *evaluator.Column, projectedTableName, projectedColumnName string) evaluator.ProjectedColumn {
	return evaluator.ProjectedColumn{
		Column: &evaluator.Column{
			SelectID:      newSelectID,
			Name:          projectedColumnName,
			OriginalName:  column.OriginalName,
			Database:      column.Database,
			Table:         projectedTableName,
			OriginalTable: column.OriginalTable,
			SQLType:       column.SQLType,
			MongoType:     column.MongoType,
			PrimaryKey:    column.PrimaryKey,
		},
		Expr: evaluator.NewSQLColumnExpr(column.SelectID, column.Database, column.Table, column.Name, column.SQLType, column.MongoType),
	}
}

func createProjectedColumn(selectID int, source evaluator.PlanStage, sourceTableName, sourceColumnName, projectedTableName, projectedColumnName string) evaluator.ProjectedColumn {
	for _, c := range source.Columns() {
		if c.MongoType == schema.MongoFilter {
			continue
		}
		if c.Table == sourceTableName && c.Name == sourceColumnName {
			return createProjectedColumnFromColumn(selectID, c, projectedTableName, projectedColumnName)
		}
	}

	panic(fmt.Sprintf("no column found with the name %q", sourceColumnName))
}

func createSQLColumnExprFromSource(source evaluator.PlanStage, tableName, columnName string) evaluator.SQLColumnExpr {
	for _, c := range source.Columns() {
		if c.MongoType == schema.MongoFilter {
			continue
		}
		if c.Table == tableName && c.Name == columnName {
			return evaluator.NewSQLColumnExpr(c.SelectID, c.Database, c.Table, c.Name, c.SQLType, c.MongoType)
		}
	}

	panic("column not found")
}

func createTestConnectionCtx(info *mongodb.Info) evaluator.ConnectionCtx {
	return &fakeConnectionCtx{info: info}
}

func createTestExecutionCtx(info *mongodb.Info) *evaluator.ExecutionCtx {
	return &evaluator.ExecutionCtx{
		ConnectionCtx: createTestConnectionCtx(info),
	}
}

func createTestVariables(info *mongodb.Info) *variable.Container {
	gbl := variable.NewGlobalContainer(nil)
	gbl.MongoDBInfo = info
	ctn := variable.NewSessionContainer(gbl)
	ctn.MongoDBInfo = info
	return ctn
}

func getCatalogFromSchema(schema *schema.Schema, variables *variable.Container) *catalog.Catalog {
	c, err := catalog.Build(schema, variables)
	if err != nil {
		panic(fmt.Sprintf("unable to build catalog: %v", err))
	}
	return c
}

// getMongoDBInfo returns Info without looking up the information in MongoDB by setting
// all privileges to the specified privileges.
func getMongoDBInfo(versionArray []uint8, sch *schema.Schema, privileges mongodb.Privilege) *mongodb.Info {
	if len(versionArray) == 0 {
		versionArray = []uint8{3, 4, 0}
	}

	versionString := ""

	for _, entry := range versionArray {
		versionString = fmt.Sprintf("%v.", entry)
	}

	i := &mongodb.Info{
		Privileges:   privileges,
		Databases:    make(map[mongodb.DatabaseName]*mongodb.DatabaseInfo),
		Version:      versionString[1:],
		VersionArray: versionArray,
	}

	for _, db := range sch.Databases {
		dbInfo := &mongodb.DatabaseInfo{
			Privileges:  privileges,
			Name:        mongodb.DatabaseName(db.Name),
			Collections: make(map[mongodb.CollectionName]*mongodb.CollectionInfo),
		}

		i.Databases[dbInfo.Name] = dbInfo

		for _, col := range db.Tables {
			if _, ok := dbInfo.Collections[mongodb.CollectionName(col.Name)]; ok {
				continue
			}

			colInfo := &mongodb.CollectionInfo{
				Privileges: privileges,
				Name:       mongodb.CollectionName(col.Name),
			}

			dbInfo.Collections[colInfo.Name] = colInfo
		}
	}

	return i
}
