package evaluator_test

import (
	"context"
	"fmt"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/catalog"
	"github.com/10gen/sqlproxy/internal/collation"
	"github.com/10gen/sqlproxy/internal/memory"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/internal/variable"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
)

type fakeConnectionCtx struct {
	memoryMonitor *memory.Monitor
	variables     *variable.Container
	info          *mongodb.Info
	server        evaluator.ServerCtx
	version       []uint8
}

func (*fakeConnectionCtx) LastInsertId() int64 {
	return 11
}
func (*fakeConnectionCtx) Logger(_ ...string) log.Logger {
	lg := log.GlobalLogger()
	return lg
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
func (*fakeConnectionCtx) RemoteHost() string {
	return "localhost"
}
func (f *fakeConnectionCtx) Server() evaluator.ServerCtx {
	return f.server
}
func (*fakeConnectionCtx) Session() *mongodb.Session {
	return nil
}
func (*fakeConnectionCtx) AuthenticationDatabase() string {
	return "test_source"
}
func (*fakeConnectionCtx) User() string {
	return "test user"
}
func (f *fakeConnectionCtx) Variables() *variable.Container {
	if f.variables == nil {
		f.variables = evaluator.CreateTestVariables(f.info)
	}
	return f.variables
}
func (f *fakeConnectionCtx) MemoryMonitor() *memory.Monitor {
	if f.memoryMonitor == nil {
		f.memoryMonitor = memory.NewMonitor("fakeConnectionCtx", 0)
	}
	return f.memoryMonitor
}

// VersionAtLeast here compares user passed in version to the version
// fakeConnectionCtx was created with. Creating with 0,0,0 will result
// in always pushing down.
func (f *fakeConnectionCtx) VersionAtLeast(userVersion ...uint8) bool {
	return util.VersionAtLeast(f.version, userVersion)
}

// bsonDToValues takes a bson.D document and returns
// the corresponding values.
func bsonDToValues(selectID int, databaseName, tableName string, document bson.D) (
	[]evaluator.Value, error) {
	values := []evaluator.Value{}
	for _, v := range document {
		value := evaluator.GoValueToSQLValue(evaluator.MySQLValueKind, v.Value)
		values = append(values, evaluator.NewValue(selectID, databaseName, tableName, v.Name,
			value))
	}
	return values, nil
}

func createAllProjectedColumnsFromSource(selectID int, source evaluator.PlanStage,
	projectedTableName string) evaluator.ProjectedColumns {
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

func createProjectedColumnFromColumn(newSelectID int, column *evaluator.Column, projectedTableName,
	projectedColumnName string) evaluator.ProjectedColumn {
	return evaluator.ProjectedColumn{
		Column: &evaluator.Column{
			SelectID:      newSelectID,
			Name:          projectedColumnName,
			OriginalName:  column.OriginalName,
			Database:      column.Database,
			Table:         projectedTableName,
			OriginalTable: column.OriginalTable,
			ColumnType: evaluator.ColumnType{
				EvalType:    column.EvalType,
				MongoType:   column.MongoType,
				UUIDSubType: evaluator.EvalNone,
			},
			PrimaryKey: column.PrimaryKey,
		},
		Expr: evaluator.NewSQLColumnExpr(column.SelectID, column.Database, column.Table,
			column.Name, column.EvalType, column.MongoType),
	}
}

func createProjectedColumn(selectID int, source evaluator.PlanStage, sourceTableName,
	sourceColumnName, projectedTableName, projectedColumnName string) evaluator.ProjectedColumn {
	for _, c := range source.Columns() {
		if c.MongoType == schema.MongoFilter {
			continue
		}
		if c.Table == sourceTableName && c.Name == sourceColumnName {
			return createProjectedColumnFromColumn(selectID, c, projectedTableName,
				projectedColumnName)
		}
	}

	panic(fmt.Sprintf("no column found with the name %q", sourceColumnName))
}

func createSQLColumnExprFromSource(source evaluator.PlanStage, tableName,
	columnName string) evaluator.SQLColumnExpr {
	for _, c := range source.Columns() {
		if c.MongoType == schema.MongoFilter {
			continue
		}
		if c.Table == tableName && c.Name == columnName {
			return evaluator.NewSQLColumnExpr(c.SelectID, c.Database, c.Table, c.Name, c.EvalType,
				c.MongoType)
		}
	}

	panic("column not found")
}

func createTestConnectionCtx(info *mongodb.Info, version ...uint8) evaluator.ConnectionCtx {
	return &fakeConnectionCtx{info: info,
		version: version,
	}
}

func createTestExecutionCtx(info *mongodb.Info, version ...uint8) *evaluator.ExecutionCtx {
	return &evaluator.ExecutionCtx{
		ConnectionCtx: createTestConnectionCtx(info, version...),
	}
}

func createTestEvalCtx(info *mongodb.Info, version ...uint8) *evaluator.EvalCtx {
	return &evaluator.EvalCtx{
		ExecutionCtx: createTestExecutionCtx(info, version...),
		Collation:    collation.Default,
	}
}

// getMongoDBInfoWithShardedCollection returns Info without looking up the information in MongoDB
//by setting all privileges to the specified privileges and a specific collection to be sharded.
func getMongoDBInfoWithShardedCollection(versionArray []uint8, sch *schema.Schema,
	privileges mongodb.Privilege, shardedCollection string) *mongodb.Info {
	info := evaluator.GetMongoDBInfo(versionArray, sch, privileges)
	for _, db := range sch.Databases() {
		// dbInfo is a pointer.
		dbInfo := info.Databases[mongodb.DatabaseName(db.Name())]
		for _, col := range db.Tables() {
			if string(col.SQLName()) == shardedCollection {
				dbInfo.Collections[mongodb.CollectionName(col.SQLName())].IsSharded = true
			}
		}
	}

	return info
}

func createFieldNameLookup(db *schema.Database) evaluator.FieldNameLookup {

	return func(databaseName, tableName, columnName string) (string, bool) {
		table := db.Table(tableName)
		if table == nil {
			return "", false
		}

		column := table.Column(columnName)
		if column == nil {
			return "", false
		}

		return column.MongoName(), true
	}
}
