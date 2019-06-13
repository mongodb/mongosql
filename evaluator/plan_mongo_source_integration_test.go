//+build integration

package evaluator_test

import (
	"context"
	"os"
	"testing"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/memory"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/testutil/dbutils"
	mongoutil "github.com/10gen/sqlproxy/internal/testutil/mongodb"
	"github.com/10gen/sqlproxy/mongodb"

	"github.com/stretchr/testify/require"
)

func getConfig(t *testing.T) *config.Config {
	cfg := config.Default()

	// ssl is turned on
	if len(os.Getenv(mongoutil.SSLTestKey)) > 0 {
		t.Logf("Testing with SSL turned on.")
		cfg.MongoDB.Net.SSL.Enabled = true
		cfg.MongoDB.Net.SSL.AllowInvalidCertificates = true
		cfg.MongoDB.Net.SSL.PEMKeyFile = "../testdata/resources/x509gen/client.pem"
	}
	return cfg
}

func TestMongoSourcePlanStage(t *testing.T) {
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

	rows := bsonutil.NewDArray(
		bsonutil.NewD(
			bsonutil.NewDocElem("a", 6),
			bsonutil.NewDocElem("b", 7),
			bsonutil.NewDocElem("d", 8),
		),
		bsonutil.NewD(
			bsonutil.NewDocElem("a", 16),
			bsonutil.NewDocElem("b", 17),
			bsonutil.NewDocElem("d", 18),
		),
	)

	var expected []results.RowValues
	var vs results.RowValues
	for _, document := range rows {
		vs, err = bsonDToValues(1, dbOne, tableThreeName, document)
		require.NoError(t, err)
		expected = append(expected, vs)
	}

	dbutils.DropCollection(session, dbOne, tableThreeName)
	dbutils.InsertDocuments(session, dbOne, tableThreeName, rows)
	defer dbutils.DropCollection(session, dbOne, tableThreeName)

	db, err := catalogOne.Database(dbOne)
	if err != nil {
		panic("database doesn't exist")
	}

	table, err := db.Table(tableThreeName)
	if err != nil {
		panic("table doesn't exist")
	}

	mongoTable, ok := table.(catalog.MongoDBTable)
	if !ok {
		panic("table is not a MongoDB table")
	}

	t.Run("with no memory limit", func(t *testing.T) {
		bgCtx := context.Background()
		monitor := memory.NewMonitor("evaluator_unit_test_monitor", 0)
		execCfg := createWorkingExecutionCfg(variablesOne, session, monitor)
		execState := evaluator.NewExecutionState()

		plan := evaluator.NewMongoSourceStage(db, mongoTable, 1, "")
		iter, err := plan.Open(bgCtx, execCfg, execState)
		require.NoError(t, err)

		row := &results.Row{}

		i := 0
		for iter.Next(bgCtx, row) {
			require.Equal(t, len(row.Data), len(expected[i]))
			require.Equal(t, row.Data, expected[i])
			row = &results.Row{}
			i++
		}

		require.NoError(t, iter.Close())
		require.NoError(t, iter.Err())
	})

	t.Run("with a memory limit", func(t *testing.T) {
		bgCtx := context.Background()
		monitor := memory.NewMonitor("evaluator_unit_test_monitor", 100)
		execCfg := createWorkingExecutionCfg(variablesOne, session, monitor)
		execState := evaluator.NewExecutionState()

		plan := evaluator.NewMongoSourceStage(db, mongoTable, 1, "")
		iter, err := plan.Open(bgCtx, execCfg, execState)
		require.NoError(t, err)

		row := &results.Row{}

		i := 0
		for iter.Next(bgCtx, row) {
			require.Equal(t, len(row.Data), len(expected[i]))
			require.Equal(t, row.Data, expected[i])
			row = &results.Row{}
			i++
		}

		require.NoError(t, iter.Close())
		require.Error(t, iter.Err())
	})
}
