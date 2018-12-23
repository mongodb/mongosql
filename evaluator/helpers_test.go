package evaluator_test

import (
	"context"
	"fmt"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/mongo-go-driver/mongo/private/ops"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/memory"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/internal/schema"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
)

type mockCmdHandler struct {
	session *mongodb.Session
}

func (c *mockCmdHandler) Aggregate(ctx context.Context, db, col string, pipeline interface{}) (ops.Cursor, error) {
	return c.session.Aggregate(ctx, db, col, pipeline)
}
func (*mockCmdHandler) Alter(context.Context, []*schema.Alteration) error {
	panic("unimplemented")
}
func (c *mockCmdHandler) Count(ctx context.Context, db, col string) (int, error) {
	return c.session.Count(ctx, db, col)
}
func (c *mockCmdHandler) Drop(tbl string) error {
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
func (*mockCmdHandler) Set(variable.Name, variable.Scope, variable.Kind, interface{}) error {
	panic("unimplemented")
}
func (*mockCmdHandler) SetDatabase(db string) error {
	panic("unimplemented")
}
func (*mockCmdHandler) SetScopeAuthorized(variable.Scope) error {
	panic("unimplemented")
}

func createAlgebrizerCfg(dbName string, cat catalog.Catalog) *evaluator.AlgebrizerConfig {
	return evaluator.NewAlgebrizerConfig(log.GlobalLogger(), dbName, cat)
}

func createExecutionCfg(dbName string, maxStageSize uint64, version []uint8) *evaluator.ExecutionConfig {
	return evaluator.CreateTestExecutionCfg(dbName, maxStageSize, version)
}

func createWorkingExecutionCfg(vars *variable.Container, ses *mongodb.Session, mon memory.Monitor) *evaluator.ExecutionConfig {
	return evaluator.NewExecutionConfig(
		log.GlobalLogger(), vars, &mockCmdHandler{ses}, mon,
		dbOne, 42, "evaluator_unit_test_user",
		"evaluator_unit_test_remotehost",
	)
}

func createTestExecutionCfg() *evaluator.ExecutionConfig {
	return createExecutionCfg("evaluator_unit_test_dbname", 0, []uint8{4, 0, 0})
}

func createOptimizerCfg(c *collation.Collation, eCfg *evaluator.ExecutionConfig) *evaluator.OptimizerConfig {
	return evaluator.CreateTestOptimizerCfg(c, eCfg)
}

func createTestPushdownCfg() *evaluator.PushdownConfig {
	return createPushdownCfg([]uint8{4, 0, 0})
}

func createPushdownCfg(version []uint8) *evaluator.PushdownConfig {
	return evaluator.CreateTestPushdownCfg(version)
}

// bsonDToValues takes a bson.D document and returns
// the corresponding values.
// nolint: unparam
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

// nolint: unparam
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
