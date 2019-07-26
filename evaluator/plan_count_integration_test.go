//+build integration

package evaluator_test

import (
	"context"
	"testing"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/memory"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/testutil/dbutils"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"github.com/stretchr/testify/require"
)

func TestCountPlanStage(t *testing.T) {
	cfgOne := setupEnv().cfgOne
	infoOne := evaluator.GetMongoDBInfo(nil, cfgOne, mongodb.AllPrivileges)
	variablesOne := evaluator.CreateTestVariables(infoOne)
	catalogOne := evaluator.GetCatalog(cfgOne, variablesOne, infoOne)
	cfg := getConfig(t)
	sessionProvider, err := mongodb.NewSqldSessionProvider(cfg)
	if err != nil {
		t.Fatalf("failed to set up session provider to test server: %v", err)
		return
	}

	session, err := sessionProvider.Session(context.Background())
	if err != nil {
		t.Fatalf("failed to set up session to test server: %v", err)
		return
	}
	defer session.Close()

	rows := bsonutil.NewDArray()
	for i := 0; i < 263; i++ {
		rows = append(rows, bsonutil.NewD(
			bsonutil.NewDocElem("_id", i),
		))
	}

	req := require.New(t)

	var expected []results.RowValues
	vs, err := bsonDToValues(1, "", "", bsonutil.NewD(bsonutil.NewDocElem("count(*)",
		263),
	))
	req.NoError(err, "failed to translate bsonD to values")
	expected = append(expected, vs)

	dbutils.DropCollection(session, dbOne, tableOneName)
	dbutils.InsertDocuments(session, dbOne, tableOneName, rows)
	defer dbutils.DropCollection(session, dbOne, tableOneName)

	bgCtx := context.Background()
	monitor := memory.NewMonitor("evaluator_unit_test_monitor", 0)
	execCfg := createWorkingExecutionCfg(variablesOne, session, monitor)
	execState := evaluator.NewExecutionState()

	db, err := catalogOne.Database(dbOne)
	if err != nil {
		panic("database doesn't exist")
	}
	table, err := db.Table(tableOneName)
	if err != nil {
		panic("table doesn't exist")
	}

	column := results.NewColumn(1, "", "", "", "count(*)", "", "", types.EvalInt64,
		schema.MongoNone, false, true)
	projectedColumn := createProjectedColumnFromColumn(1, column, "", "count(*)", false)

	mongoTable, ok := table.(catalog.MongoDBTable)
	if !ok {
		panic("table is not a MongoDB table")
	}

	mongoSourceStage := evaluator.NewMongoSourceStage(db, mongoTable, 1, "")
	countStage := evaluator.NewCountStage(mongoSourceStage, projectedColumn)

	iter, err := countStage.Open(bgCtx, execCfg, execState)
	req.Nil(err, "error opening count stage")

	row := &results.Row{}

	i := 0

	for iter.Next(bgCtx, row) {
		req.Equal(len(row.Data), len(expected[i]), "number of returned columns does not"+
			" match number of expected columns")
		req.Equal(row.Data, expected[i], "returned data values does not match expected"+
			" data values")
		row = &results.Row{}
		i++
	}
	req.Equal(i, len(expected), "returned number of rows does not match expected number"+
		" of rows")
	req.Nil(iter.Close(), "error closing the iterator")
	req.Nil(iter.Err(), "iterator returned with an error")

	actualAllocated := monitor.Allocated()
	expectedAllocated := valueSize(
		"", "", "count(*)",
		values.NewSQLInt64(values.MySQLValueKind, 0),
	)

	req.Equal(expectedAllocated, actualAllocated)
}
