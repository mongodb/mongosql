package evaluator

import (
	"context"
	"fmt"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
)

// bsonDToValues takes a bson.D document and returns
// the corresponding values.
func bsonDToValues(selectID int, databaseName, tableName string, document bson.D) ([]Value, error) {
	values := []Value{}
	for _, v := range document {
		value, err := NewSQLValueFromSQLColumnExpr(v.Value, schema.SQLNone, schema.MongoNone)
		if err != nil {
			return nil, err
		}
		values = append(values, NewValue(selectID, databaseName, tableName, v.Name, value))
	}
	return values, nil
}

func constructProjectedColumns(exprs map[string]SQLExpr, values ...string) (projectedColumns ProjectedColumns) {
	for _, value := range values {

		expr := exprs[value]

		column := &Column{
			Name: value,
		}

		projectedColumns = append(projectedColumns, ProjectedColumn{
			Column: column,
			Expr:   expr,
		})
	}
	return
}

func constructOrderByTerms(exprs map[string]SQLExpr, values ...string) (terms []*orderByTerm) {
	for i, v := range values {

		term := &orderByTerm{
			expr:      exprs[v],
			ascending: i%2 == 0,
		}

		terms = append(terms, term)
	}

	return
}

type fakeConnectionCtx struct {
	variables *variable.Container
	info      *mongodb.Info
	server    ServerCtx
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
func (*fakeConnectionCtx) Kill(id uint32, scope KillScope) error {
	return nil
}
func (f *fakeConnectionCtx) Server() ServerCtx {
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

func createTestConnectionCtx(info *mongodb.Info) ConnectionCtx {
	return &fakeConnectionCtx{info: info}
}

func createTestExecutionCtx(info *mongodb.Info) *ExecutionCtx {
	return &ExecutionCtx{
		ConnectionCtx: createTestConnectionCtx(info),
	}
}

func createTestEvalCtx(info *mongodb.Info) *EvalCtx {
	return &EvalCtx{
		ExecutionCtx: createTestExecutionCtx(info),
	}
}

func createTestVariables(info *mongodb.Info) *variable.Container {
	gbl := variable.NewGlobalContainer(nil)
	gbl.MongoDBInfo = info
	ctn := variable.NewSessionContainer(gbl)
	ctn.MongoDBInfo = info
	return ctn
}

func createSQLColumnExprFromSource(source PlanStage, tableName, columnName string) SQLColumnExpr {
	for _, c := range source.Columns() {
		if c.MongoType == schema.MongoFilter {
			continue
		}
		if c.Table == tableName && c.Name == columnName {
			return NewSQLColumnExpr(c.SelectID, c.Database, c.Table, c.Name, c.SQLType, c.MongoType)
		}
	}

	panic("column not found")
}

func createProjectedColumnFromColumn(newSelectID int, column *Column, projectedTableName, projectedColumnName string) ProjectedColumn {
	return ProjectedColumn{
		Column: &Column{
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
		Expr: NewSQLColumnExpr(column.SelectID, column.Database, column.Table, column.Name, column.SQLType, column.MongoType),
	}
}

func createProjectedColumn(selectID int, source PlanStage, sourceTableName, sourceColumnName, projectedTableName, projectedColumnName string) ProjectedColumn {
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

func createProjectedColumnWithDatabase(selectID int, source PlanStage, sourceDatabaseName, sourceTableName, sourceColumnName, projectedTableName, projectedColumnName string) ProjectedColumn {
	var dbName string
	for _, c := range source.Columns() {
		if c.MongoType == schema.MongoFilter {
			continue
		}
		if c.Table == sourceTableName && c.Name == sourceColumnName && c.Database == sourceDatabaseName {
			return createProjectedColumnFromColumnWithDatabase(selectID, c, sourceDatabaseName, projectedTableName, projectedColumnName)
		}
		dbName = c.Database
	}

	panic(fmt.Sprintf("no column found with the name %q from database: %s", sourceColumnName, dbName))
}

func createProjectedColumnFromColumnWithDatabase(newSelectID int, column *Column, databaseName, projectedTableName, projectedColumnName string) ProjectedColumn {
	return ProjectedColumn{
		Column: &Column{
			SelectID:      newSelectID,
			Name:          projectedColumnName,
			OriginalName:  column.OriginalName,
			Database:      databaseName,
			Table:         projectedTableName,
			OriginalTable: column.OriginalTable,
			SQLType:       column.SQLType,
			MongoType:     column.MongoType,
			PrimaryKey:    column.PrimaryKey,
		},
		Expr: NewSQLColumnExpr(column.SelectID, databaseName, column.Table, column.Name, column.SQLType, column.MongoType),
	}
}

func createAllProjectedColumnsFromSource(selectID int, source PlanStage, projectedTableName string) ProjectedColumns {
	results := ProjectedColumns{}
	for _, c := range source.Columns() {
		if c.MongoType == schema.MongoFilter {
			continue
		}
		results = append(results, createProjectedColumnFromColumn(
			selectID, c, projectedTableName, c.Name))
	}

	return results
}

func createProjectedColumnFromSQLExpr(selectID int, columnName string, expr SQLExpr) ProjectedColumn {
	column := &Column{
		SelectID: selectID,
		Name:     columnName,
		SQLType:  expr.Type(),
	}

	if sqlColExpr, ok := expr.(SQLColumnExpr); ok {
		column.MongoType = sqlColExpr.columnType.MongoType
	}

	return ProjectedColumn{Column: column, Expr: expr}
}

func getBinaryExprLeaves(expr SQLExpr) (SQLExpr, SQLExpr) {
	switch typedE := expr.(type) {
	case *SQLAndExpr:
		return typedE.left, typedE.right
	case *SQLAddExpr:
		return typedE.left, typedE.right
	case *SQLSubtractExpr:
		return typedE.left, typedE.right
	case *SQLMultiplyExpr:
		return typedE.left, typedE.right
	case *SQLDivideExpr:
		return typedE.left, typedE.right
	case *SQLEqualsExpr:
		return typedE.left, typedE.right
	case *SQLLessThanExpr:
		return typedE.left, typedE.right
	case *SQLGreaterThanExpr:
		return typedE.left, typedE.right
	case *SQLLessThanOrEqualExpr:
		return typedE.left, typedE.right
	case *SQLGreaterThanOrEqualExpr:
		return typedE.left, typedE.right
	case *SQLLikeExpr:
		return typedE.left, typedE.right
	case *SQLSubqueryExpr:
		return nil, &SQLTupleExpr{typedE.Exprs()}
	//case *SQLSubqueryCmpExpr:
	// return typedE.left, &SQLTupleExpr{typedE.value.exprs}
	case *SQLInExpr:
		return typedE.left, typedE.right
	}
	return nil, nil
}

func getSQLExpr(schema *schema.Schema, dbName, tableName, sql string) (SQLExpr, error) {
	statement, err := parser.Parse("select " + sql + " from " + tableName)
	if err != nil {
		return nil, err
	}

	selectStatement := statement.(parser.SelectStatement)
	info := getMongoDBInfo(nil, schema, mongodb.AllPrivileges)
	vars := createTestVariables(info)
	catalog := getCatalogFromSchema(schema, vars)
	actualPlan, err := AlgebrizeQuery(selectStatement, dbName, vars, catalog)
	if err != nil {
		return nil, err
	}

	// Depending on the "sql" expression we are getting, the algebrizer could have put it in
	// either the ProjectStage (for non-aggregate expressions) or a GroupByStage (for aggregate
	// expressions). We don't know which one the user is asking for, so we'll assume the
	// GroupByStage if it exists, otherwise the ProjectStage.
	project := actualPlan.(*ProjectStage)
	expr := project.projectedColumns[0].Expr

	group, ok := project.source.(*GroupByStage)
	if ok {
		expr = group.projectedColumns[0].Expr
	}

	if conv, ok := expr.(*SQLConvertExpr); ok {
		expr = conv.expr
	}

	return expr, nil
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

// getMongoDBInfoWithShardedCollection returns Info without looking up the information in MongoDB by setting
// all privileges to the specified privileges and a specific collection to be sharded.
func getMongoDBInfoWithShardedCollection(versionArray []uint8, sch *schema.Schema, privileges mongodb.Privilege, shardedCollection string) *mongodb.Info {
	info := getMongoDBInfo(versionArray, sch, privileges)
	for _, db := range sch.Databases {
		// dbInfo is a pointer.
		dbInfo := info.Databases[mongodb.DatabaseName(db.Name)]
		for _, col := range db.Tables {
			if string(col.Name) == shardedCollection {
				dbInfo.Collections[mongodb.CollectionName(col.Name)].IsSharded = true
			}
		}
	}

	return info
}

func getCatalogFromSchema(schema *schema.Schema, variables *variable.Container) *catalog.Catalog {
	c, err := catalog.Build(schema, variables)
	if err != nil {
		panic(fmt.Sprintf("unable to build catalog: %v", err))
	}
	return c
}

func GetSQLExpr(schema *schema.Schema, dbName, tableName, sql string) (SQLExpr, error) {
	statement, err := parser.Parse("select " + sql + " from " + tableName)
	if err != nil {
		return nil, err
	}

	selectStatement := statement.(parser.SelectStatement)
	info := getMongoDBInfo(nil, schema, mongodb.AllPrivileges)
	vars := createTestVariables(info)
	catalog := getCatalogFromSchema(schema, vars)
	actualPlan, err := AlgebrizeQuery(selectStatement, dbName, vars, catalog)
	if err != nil {
		return nil, err
	}

	// Depending on the "sql" expression we are getting, the algebrizer could have put it in
	// either the ProjectStage (for non-aggregate expressions) or a GroupByStage (for aggregate
	// expressions). We don't know which one the user is asking for, so we'll assume the
	// GroupByStage if it exists, otherwise the ProjectStage.
	project := actualPlan.(*ProjectStage)
	expr := project.projectedColumns[0].Expr

	group, ok := project.source.(*GroupByStage)
	if ok {
		expr = group.projectedColumns[0].Expr
	}

	if conv, ok := expr.(*SQLConvertExpr); ok {
		expr = conv.expr
	}

	return expr, nil
}
