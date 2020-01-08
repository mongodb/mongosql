package evaluator_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/option"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestPlanDDLStage(t *testing.T) {
	req := require.New(t)
	dbName := "writes_test_1"
	tableName := "foo"

	bgCtx := context.Background()
	lg := log.NoOpLogger()

	cfg := config.Default()
	cfg.MongoDB.VersionCompatibility = "4.2.0"
	vars := variable.NewGlobalContainer(cfg)
	// set up a catalog with a database writes_test_1 and table foo
	cat := catalog.New("def")
	_, err := cat.AddDatabase(dbName)
	req.NoError(err)
	catDB, err := cat.Database(bgCtx, dbName)
	req.NoError(err)
	err = catDB.AddTable(catalog.NewDynamicTable(tableName, "", results.NewEmptyRowChanIter))
	req.NoError(err)

	exeCfg := evaluator.NewExecutionConfig(lg, vars, newTestCommandHandler(cat), nil, "information_schema")

	// create a database that already exists with 'if not exists' set, should not error.
	dbc := evaluator.NewCreateDatabaseCommand(cat, dbName, true)
	err = dbc.Execute(bgCtx, exeCfg, nil)
	req.NoError(err)

	// Run the command again, this should return an error telling us that the commandHandler is invoked.
	dbc = evaluator.NewCreateDatabaseCommand(cat, dbName, false)
	err = dbc.Execute(bgCtx, exeCfg, nil)
	req.EqualError(err, "ran create database")

	table := getTable(tableName)
	// create a table in a non-existent database with if not exists set.
	tcc := evaluator.NewCreateTableCommand(cat, "random", table, true)
	err = tcc.Execute(bgCtx, exeCfg, nil)
	req.EqualError(err, "ERROR 1049 (42000): Unknown database 'random'")

	// create a table that already exists in a existent database with if not exists set.
	tcc = evaluator.NewCreateTableCommand(cat, dbName, table, true)
	err = tcc.Execute(bgCtx, exeCfg, nil)
	req.NoError(err)

	// create a table if not exists unset, this will
	// delegate to the commandHandler.
	tcc = evaluator.NewCreateTableCommand(cat, dbName, table, false)
	err = tcc.Execute(bgCtx, exeCfg, nil)
	req.EqualError(err, "ran create table")

	// drop a database that does not exist with if not exists set.
	ddc := evaluator.NewDropDatabaseCommand(cat, "random", true)
	err = ddc.Execute(bgCtx, exeCfg, nil)
	req.NoError(err)

	// drop a database with if not exists unset, this will
	// delegate to the commandHandler.
	ddc = evaluator.NewDropDatabaseCommand(cat, dbName, false)
	err = ddc.Execute(bgCtx, exeCfg, nil)
	req.EqualError(err, "ran drop database")

	// drop a table in a non-existent database that does not exist with if not exists set.
	tdc := evaluator.NewDropTableCommand(cat, "random", "random", true)
	err = tdc.Execute(bgCtx, exeCfg, nil)
	req.NoError(err)

	// drop table that does not exist with if not exists set.
	tdc = evaluator.NewDropTableCommand(cat, dbName, "random", true)
	err = tdc.Execute(bgCtx, exeCfg, nil)
	req.NoError(err)

	// drop table with if not exists unset. Should delegate to commandHandler.
	tdc = evaluator.NewDropTableCommand(cat, dbName, tableName, true)
	err = tdc.Execute(bgCtx, exeCfg, nil)
	req.EqualError(err, "ran drop table")

	// Now drop a #Tableau table
	tdc = evaluator.NewDropTableCommand(cat, dbName, "#Tableau", false)
	err = tdc.Execute(bgCtx, exeCfg, nil)
	req.NoError(err)
}

type ddlTestCommandHandler struct {
	catalog *catalog.SQLCatalog
}

func (ddlTestCommandHandler) Aggregate(ctx context.Context, db, col string, pipeline []bson.D) (mongodb.Cursor, error) {
	panic("unimplemented")
}
func (ddlTestCommandHandler) Count(ctx context.Context, db, col string) (int, error) {
	panic("unimplemented")
}
func (ddlTestCommandHandler) DropTable(ctx context.Context, db, tbl string) error {
	return fmt.Errorf("ran drop table")
}
func (ddlTestCommandHandler) DropDatabase(ctx context.Context, db string) error {
	return fmt.Errorf("ran drop database")
}
func (ddlTestCommandHandler) CreateTable(ctx context.Context, db string, table *schema.Table) error {
	return fmt.Errorf("ran create table")
}
func (ddlTestCommandHandler) CreateDatabase(ctx context.Context, db string) error {
	return fmt.Errorf("ran create database")
}
func (ddlTestCommandHandler) Insert(ctx context.Context, db, table string, docs []interface{}) error {
	panic("unimplemented")
}
func (ddlTestCommandHandler) Kill(ctx context.Context, targetConnID uint32, ks evaluator.KillScope) error {
	panic("unimplemented")
}
func (ddlTestCommandHandler) Resample(context.Context) error {
	panic("unimplemented")
}
func (ddlTestCommandHandler) RotateLogs() error {
	panic("unimplemented")
}
func (ddlTestCommandHandler) Set(variable.Name, variable.Scope, variable.Kind, values.SQLValue) error {
	panic("unimplemented")
}
func (ddlTestCommandHandler) SetDatabase(db string) error {
	panic("unimplemented")
}
func (ddlTestCommandHandler) SetScopeAuthorized(variable.Scope) error {
	panic("unimplemented")
}
func (ddlTestCommandHandler) UnsetDatabase() error {
	panic("unimplemented")
}

func newTestCommandHandler(catalog *catalog.SQLCatalog) ddlTestCommandHandler {
	return ddlTestCommandHandler{
		catalog,
	}
}

func getTable(tableName string) *schema.Table {
	table := newTableTestHelper(
		log.NoOpLogger(),
		tableName,
		tableName,
		[]bson.D{},
		[]*schema.Column{
			schema.NewColumn("a", schema.SQLInt, "a", schema.MongoInt64, false, option.NoneString()),
			schema.NewColumn("b", schema.SQLVarchar, "B", schema.MongoString, false, option.SomeString("fooo")),
			schema.NewColumn("c", schema.SQLVarchar, "C", schema.MongoString, true, option.SomeString("HELLO!")),
		},
		[]schema.Index{
			schema.NewIndex("bAr", true, false,
				[]schema.IndexPart{schema.NewIndexPart("a", 1), schema.NewIndexPart("b", -1)},
			),
			schema.NewIndex("", false, true,
				[]schema.IndexPart{schema.NewIndexPart("b", 1), schema.NewIndexPart("c", 1)},
			),
		},
		option.SomeString("WORLD"),
	)

	return table
}

func newTableTestHelper(lg log.Logger, tbl, col string,
	pipeline []bson.D, cols []*schema.Column,
	indexes []schema.Index, comment option.String) *schema.Table {
	out, err := schema.NewTable(lg, tbl, col, pipeline, cols, indexes, comment)
	if err != nil {
		panic("this table should not error")
	}
	return out
}
