package evaluator

import (
	"fmt"

	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/tomb.v2"
)

// bsonDToValues takes a bson.D document and returns
// the corresponding values.
func bsonDToValues(selectID int, tableName string, document bson.D) ([]Value, error) {
	values := []Value{}
	for _, v := range document {
		value, err := NewSQLValueFromSQLColumnExpr(v.Value, schema.SQLNone, schema.MongoNone)
		if err != nil {
			return nil, err
		}
		values = append(values, Value{selectID, tableName, v.Name, value})
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
}

func (_ fakeConnectionCtx) LastInsertId() int64 {
	return 11
}
func (_ fakeConnectionCtx) RowCount() int64 {
	return 21
}
func (_ fakeConnectionCtx) ConnectionId() uint32 {
	return 42
}
func (_ fakeConnectionCtx) DB() string {
	return "test"
}
func (_ fakeConnectionCtx) Kill(id uint32, scope KillScope) error {
	return nil
}
func (_ fakeConnectionCtx) Session() *mgo.Session {
	return nil
}
func (_ fakeConnectionCtx) User() string {
	return "test user"
}
func (_ fakeConnectionCtx) Tomb() *tomb.Tomb {
	return nil
}
func (f fakeConnectionCtx) Variables() *variable.Container {
	if f.variables == nil {
		f.variables = variable.NewSessionContainer(variable.NewGlobalContainer())
	}
	return f.variables
}

func createTestConnectionCtx() ConnectionCtx {
	return &fakeConnectionCtx{}
}

func createTestExecutionCtx() *ExecutionCtx {
	return &ExecutionCtx{
		ConnectionCtx: createTestConnectionCtx(),
	}
}

func createTestEvalCtx() *EvalCtx {
	return &EvalCtx{
		ExecutionCtx: createTestExecutionCtx(),
	}
}

func createSQLColumnExprFromSource(source PlanStage, tableName, columnName string) SQLColumnExpr {
	for _, c := range source.Columns() {
		if c.MongoType == schema.MongoFilter {
			continue
		}
		if c.Table == tableName && c.Name == columnName {
			return NewSQLColumnExpr(c.SelectID, c.Table, c.Name, c.SQLType, c.MongoType)
		}
	}

	panic("column not found")
}

func createProjectedColumnFromColumn(newSelectID int, column *Column, projectedTableName, projectedColumnName string) ProjectedColumn {
	return ProjectedColumn{
		Column: &Column{
			SelectID:  newSelectID,
			Table:     projectedTableName,
			Name:      projectedColumnName,
			SQLType:   column.SQLType,
			MongoType: column.MongoType,
		},
		Expr: NewSQLColumnExpr(column.SelectID, column.Table, column.Name, column.SQLType, column.MongoType),
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

func createAllProjectedColumnsFromSource(selectID int, source PlanStage, projectedTableName string) ProjectedColumns {
	results := ProjectedColumns{}
	for _, c := range source.Columns() {
		if c.MongoType == schema.MongoFilter {
			continue
		}
		results = append(results, createProjectedColumnFromColumn(selectID, c, projectedTableName, c.Name))
	}

	return results
}

func createProjectedColumnFromSQLExpr(selectID int, tableName, columnName string, expr SQLExpr) ProjectedColumn {
	column := &Column{
		SelectID: selectID,
		Table:    tableName,
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
	actualPlan, err := AlgebrizeSelect(selectStatement, dbName, schema)
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
func getMongoDBInfo(sch *schema.Schema, privileges mongodb.Privilege) *mongodb.Info {
	i := &mongodb.Info{
		Privileges: privileges,
		Databases:  make(map[mongodb.DatabaseName]*mongodb.DatabaseInfo),
	}
	for _, db := range sch.RawDatabases {
		dbInfo := &mongodb.DatabaseInfo{
			Privileges:  privileges,
			Name:        mongodb.DatabaseName(db.Name),
			Collections: make(map[mongodb.CollectionName]*mongodb.CollectionInfo),
		}

		i.Databases[dbInfo.Name] = dbInfo

		for _, col := range db.RawTables {
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
