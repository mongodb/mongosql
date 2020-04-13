package evaluator_test

import (
	"context"
	"fmt"

	"github.com/10gen/mongoast/ast"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/memory"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"

	"go.mongodb.org/mongo-driver/bson"
)

func testSQLColumnExpr(selectID int,
	databaseName, tableName, columnName string,
	evalType types.EvalType,
	mongoType schema.MongoType,
	correlated bool) evaluator.SQLColumnExpr {
	return evaluator.NewSQLColumnExpr(selectID, databaseName, tableName, columnName, evalType, mongoType, correlated, true)
}

type mockCmdHandler struct {
	session *mongodb.Session
}

func (c *mockCmdHandler) Aggregate(ctx context.Context, db, col string, pipeline []bson.D) (mongodb.Cursor, error) {
	return c.session.Aggregate(ctx, db, col, pipeline)
}
func (c *mockCmdHandler) Count(ctx context.Context, db, col string) (int, error) {
	return c.session.Count(ctx, db, col)
}
func (*mockCmdHandler) DropTable(ctx context.Context, db, tbl string) error {
	panic("unimplemented")
}
func (*mockCmdHandler) DropDatabase(ctx context.Context, db string) error {
	panic("unimplemented")
}
func (*mockCmdHandler) CreateTable(ctx context.Context, db string, table *schema.Table) error {
	panic("unimplemented")
}
func (*mockCmdHandler) CreateDatabase(ctx context.Context, db string) error {
	panic("unimplemented")
}
func (*mockCmdHandler) Insert(ctx context.Context, db, table string, docs []interface{}) error {
	panic("unimplemented")
}
func (*mockCmdHandler) Kill(context.Context, uint32, evaluator.KillScope) error {
	panic("unimplemented")
}
func (*mockCmdHandler) Resample(context.Context) error {
	panic("unimplemented")
}
func (*mockCmdHandler) RotateLogs() error {
	panic("unimplemented")
}
func (*mockCmdHandler) Set(variable.Name, variable.Scope, variable.Kind, values.SQLValue) error {
	panic("unimplemented")
}
func (*mockCmdHandler) SetDatabase(db string) error {
	panic("unimplemented")
}
func (*mockCmdHandler) SetScopeAuthorized(variable.Scope) error {
	panic("unimplemented")
}
func (*mockCmdHandler) UnsetDatabase() error {
	panic("unimplemented")
}

func createAlgebrizerCfg(dbName string, ctlg catalog.Catalog, vars *variable.Container, isWriteMode bool) *evaluator.AlgebrizerConfig {
	mongoDBToplogy := vars.GetString(variable.MongoDBTopology)
	sqlValueKind := evaluator.GetSQLValueKind(vars)
	sqlSelectLimit := vars.GetUint64(variable.SQLSelectLimit)
	mongoDBMaxVarcharLength := vars.GetUint64(variable.MongoDBMaxVarcharLength)
	groupConcatMaxLen := vars.GetInt64(variable.GroupConcatMaxLen)
	polymorphicTypeConversionMode := vars.GetString(variable.PolymorphicTypeConversionMode)
	mdbVersion := evaluator.GetMongoDBVersion(vars)

	return evaluator.NewAlgebrizerConfig(log.GlobalLogger(), dbName, ctlg, vars, mongoDBToplogy,
		isWriteMode, sqlValueKind, sqlSelectLimit, mongoDBMaxVarcharLength,
		groupConcatMaxLen, polymorphicTypeConversionMode, mdbVersion, true, false, false, false)
}

func createExecutionCfg(dbName string, maxStageSize uint64, version []uint8, sqlValueKind values.SQLValueKind) *evaluator.ExecutionConfig {
	return evaluator.CreateTestExecutionCfg(dbName, maxStageSize, version, sqlValueKind)
}

func createWorkingExecutionCfg(vars *variable.Container, ses *mongodb.Session, mon memory.Monitor) *evaluator.ExecutionConfig {
	mdbVersion := evaluator.GetMongoDBVersion(vars)
	fullPushdownOnly := vars.GetBool(variable.FullPushdownExecMode)
	maxStageSize := vars.GetUint64(variable.MongoDBMaxStageSize)
	sqlValueKind := evaluator.GetSQLValueKind(vars)

	return evaluator.NewExecutionConfig(
		log.GlobalLogger(), dbOne, mdbVersion, fullPushdownOnly,
		maxStageSize, sqlValueKind, &mockCmdHandler{ses}, mon,
	)
}

// nolint: unparam
func createTestExecutionCfg(sqlValueKind values.SQLValueKind) *evaluator.ExecutionConfig {
	return createExecutionCfg("evaluator_unit_test_dbname", 0, []uint8{4, 0, 0}, sqlValueKind)
}

func createOptimizerCfg(c *collation.Collation, eCfg *evaluator.ExecutionConfig) *evaluator.OptimizerConfig {
	return evaluator.CreateTestOptimizerCfg(c, eCfg)
}

func createTestPushdownCfg() *evaluator.PushdownConfig {
	return createPushdownCfg([]uint8{4, 0, 0}, values.MySQLValueKind)
}

func createPushdownCfg(version []uint8, sqlValueKind values.SQLValueKind) *evaluator.PushdownConfig {
	return evaluator.CreateTestPushdownCfg(version, sqlValueKind)
}

// bsonDToValues takes a bson.D document and returns the corresponding values.
// nolint: unparam
func bsonDToValues(selectID int, databaseName, tableName string, document bson.D) ([]results.RowValue, error) {
	vs := make([]results.RowValue, len(document))
	for i, v := range document {
		value := evaluator.GoValueToSQLValue(values.MySQLValueKind, v.Value)
		vs[i] = results.NewRowValue(selectID, databaseName, tableName, v.Key, value)
	}
	return vs, nil
}

// nolint: unparam
func createAllProjectedColumnsFromSource(selectID int, source evaluator.PlanStage,
	projectedTableName string) evaluator.ProjectedColumns {
	results := evaluator.ProjectedColumns{}
	for _, c := range source.Columns() {
		if c.ColumnType.MongoType == schema.MongoFilter {
			continue
		}
		results = append(results, createProjectedColumnFromColumn(
			selectID, c, projectedTableName, c.Name, false))
	}

	return results
}

func createProjectedColumnFromColumn(
	newSelectID int,
	column *results.Column,
	projectedTableName, projectedColumnName string,
	correlated bool,
) evaluator.ProjectedColumn {
	return evaluator.ProjectedColumn{
		Column: &results.Column{
			SelectID:      newSelectID,
			Name:          projectedColumnName,
			OriginalName:  column.OriginalName,
			Database:      column.Database,
			Table:         projectedTableName,
			OriginalTable: column.OriginalTable,
			ColumnType: &results.ColumnType{
				EvalType:    column.ColumnType.EvalType,
				MongoType:   column.ColumnType.MongoType,
				UUIDSubType: types.EvalBinary,
			},
			PrimaryKey: column.PrimaryKey,
			Comments:   column.Comments,
			MongoName:  column.MongoName,
			Nullable:   column.Nullable,
		},
		Expr: testSQLColumnExpr(column.SelectID, column.Database, column.Table,
			column.Name, column.ColumnType.EvalType, column.ColumnType.MongoType, correlated),
	}
}

func createProjectedColumn(selectID int, source evaluator.PlanStage, sourceTableName, sourceColumnName, projectedTableName, projectedColumnName string, correlated bool) evaluator.ProjectedColumn {
	for _, c := range source.Columns() {
		if c.ColumnType.MongoType == schema.MongoFilter {
			continue
		}
		if c.Table == sourceTableName && c.Name == sourceColumnName {
			return createProjectedColumnFromColumn(selectID, c, projectedTableName, projectedColumnName, correlated)
		}
	}

	panic(fmt.Sprintf("no column found with the name %q", sourceColumnName))
}

func createSQLColumnExprFromSource(source evaluator.PlanStage, tableName, columnName string, correlated bool) evaluator.SQLColumnExpr {
	for _, c := range source.Columns() {
		if c.ColumnType.MongoType == schema.MongoFilter {
			continue
		}
		if c.Table == tableName && c.Name == columnName {
			return testSQLColumnExpr(c.SelectID, c.Database, c.Table, c.Name, c.ColumnType.EvalType, c.ColumnType.MongoType, correlated)
		}
	}

	panic("column not found")
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
			if col.SQLName() == shardedCollection {
				dbInfo.Collections[mongodb.CollectionName(col.SQLName())].IsSharded = true
			}
		}
	}

	return info
}

func createFieldRefLookup(db *schema.Database) evaluator.FieldRefLookup {
	return func(databaseName, tableName, columnName string) (ast.Ref, bool) {
		table := db.Table(tableName)
		if table == nil {
			return nil, false
		}

		column := table.Column(columnName)
		if column == nil {
			return nil, false
		}

		return ast.NewFieldRef(column.MongoName(), nil), true
	}
}

// This function sets all unimportant fields to default values so
// that we can make algebrizer equality asserts pass.
func unsetUnimportantFields(p evaluator.PlanStage) evaluator.PlanStage {
	var aux func(n evaluator.Node) evaluator.Node
	aux = func(n evaluator.Node) evaluator.Node {
		switch typedN := n.(type) {
		case *evaluator.ProjectStage:
			for _, projectedColumn := range typedN.ProjectedColumns() {
				projectedColumn.Column.Comments = ""
				projectedColumn.Column.MongoName = ""
				projectedColumn.Column.Nullable = true
			}
		}
		for i, child := range n.Children() {
			n.ReplaceChild(i, aux(child))
		}
		return n
	}
	return aux(p.(evaluator.Node)).(evaluator.PlanStage)
}
