package evaluator_test

import (
	"context"
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/catalog"
	"github.com/10gen/sqlproxy/internal/testutils/dbutils"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"github.com/stretchr/testify/require"
)

func TestCountPlanStage(t *testing.T) {
	env := setupEnv(t)
	cfgOne := env.cfgOne
	infoOne := evaluator.GetMongoDBInfo(nil, cfgOne, mongodb.AllPrivileges)
	variablesOne := evaluator.CreateTestVariables(infoOne)
	catalogOne := evaluator.GetCatalogFromSchema(cfgOne, variablesOne)
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

	rows := []bson.D{}
	for i := 0; i < 263; i++ {
		rows = append(rows, bson.D{
			bson.DocElem{Name: "_id", Value: i},
		})
	}

	req := require.New(t)

	var expected []evaluator.Values
	values, err := bsonDToValues(1, "", "", bson.D{bson.DocElem{Name: "count(*)",
		Value: 263}})
	req.NoError(err, "failed to translate bsonD to values")
	expected = append(expected, values)

	dbutils.DropCollection(session, dbOne, tableOneName)
	dbutils.InsertDocuments(session, dbOne, tableOneName, rows)
	defer dbutils.DropCollection(session, dbOne, tableOneName)

	cCtx := &connCtx{
		catalog:   catalogOne,
		session:   session,
		variables: variablesOne,
	}

	ctx := &evaluator.ExecutionCtx{
		ConnectionCtx: cCtx,
	}

	db, err := catalogOne.Database(dbOne)
	if err != nil {
		panic("database doesn't exist")
	}
	table, err := db.Table(tableOneName)
	if err != nil {
		panic("table doesn't exist")
	}

	column := evaluator.NewColumn(1, "", "", "", "count(*)", "", "", evaluator.EvalInt64,
		schema.MongoNone, false)
	projectedColumn := createProjectedColumnFromColumn(1, column, "", "count(*)")

	mongoSourceStage := evaluator.NewMongoSourceStage(db, table.(*catalog.MongoTable), 1,
		"")
	countStage := evaluator.NewCountStage(mongoSourceStage, projectedColumn)

	iter, err := countStage.Open(ctx)
	req.Nil(err, "error opening count stage")

	row := &evaluator.Row{}

	i := 0

	for iter.Next(row) {
		req.Equal(len(row.Data), len(expected[i]), "number of returned columns does not"+
			" match number of expected columns")
		req.Equal(row.Data, expected[i], "returned data values does not match expected"+
			" data values")
		row = &evaluator.Row{}
		i++
	}
	req.Equal(i, len(expected), "returned number of rows does not match expected number"+
		" of rows")
	req.Nil(iter.Close(), "error closing the iterator")
	req.Nil(iter.Err(), "iterator returned with an error")

	actualAllocated := ctx.MemoryMonitor().Allocated()
	expectedAllocated := valueSize(
		"", "", "count(*)",
		evaluator.NewSQLInt64(evaluator.MySQLValueKind, 0),
	)

	req.Equal(expectedAllocated, actualAllocated)
}
