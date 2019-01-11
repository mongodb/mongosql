//+build integration

package evaluator_test

import (
	"context"
	"os"
	"testing"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/memory"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/testutil/dbutils"
	mongoutil "github.com/10gen/sqlproxy/internal/testutil/mongodb"
	"github.com/10gen/sqlproxy/mongodb"

	"github.com/smartystreets/goconvey/convey"
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

	var expected []evaluator.Values
	var values []evaluator.Value
	for _, document := range rows {
		values, err = bsonDToValues(1, dbOne, tableThreeName, document)
		require.NoError(t, err)
		expected = append(expected, values)
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

		row := &evaluator.Row{}

		i := 0
		for iter.Next(bgCtx, row) {
			require.Equal(t, len(row.Data), len(expected[i]))
			require.Equal(t, row.Data, expected[i])
			row = &evaluator.Row{}
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

		row := &evaluator.Row{}

		i := 0
		for iter.Next(bgCtx, row) {
			require.Equal(t, len(row.Data), len(expected[i]))
			require.Equal(t, row.Data, expected[i])
			row = &evaluator.Row{}
			i++
		}

		require.NoError(t, iter.Close())
		require.Error(t, iter.Err())
	})
}

func TestExtractField(t *testing.T) {
	req := require.New(t)

	testD := bsonutil.NewD(
		bsonutil.NewDocElem("a", "string"),
		bsonutil.NewDocElem("b", bsonutil.NewArray(
			"inner",
			bsonutil.NewD(bsonutil.NewDocElem("inner2", 1)),
		)),
		bsonutil.NewDocElem("c", bsonutil.NewD(bsonutil.NewDocElem("x", 5))),
		bsonutil.NewDocElem("d", bsonutil.NewD(bsonutil.NewDocElem("z", nil))),
	)

	// regular fields should be extracted by name
	val, ok := bsonutil.ExtractFieldByName("a", testD)
	req.Equal(val, "string")
	req.True(ok)

	// array fields should be extracted by name
	val, ok = bsonutil.ExtractFieldByName("b.1", testD)
	req.Zero(convey.ShouldResemble(val, bsonutil.NewD(bsonutil.NewDocElem("inner2", 1))))
	req.True(ok)
	val, ok = bsonutil.ExtractFieldByName("b.1.inner2", testD)
	req.Equal(val, 1)
	req.True(ok)
	val, ok = bsonutil.ExtractFieldByName("b.0", testD)
	req.Equal(val, "inner")
	req.True(ok)

	// subdocument fields should be extracted by name
	val, ok = bsonutil.ExtractFieldByName("c", testD)
	req.Zero(convey.ShouldResemble(val, bsonutil.NewD(bsonutil.NewDocElem("x", 5))))
	req.True(ok)
	val, ok = bsonutil.ExtractFieldByName("c.x", testD)
	req.Equal(val, 5)
	req.True(ok)

	// even if they contain null values
	val, ok = bsonutil.ExtractFieldByName("d", testD)
	req.Zero(convey.ShouldResemble(val, bsonutil.NewD(bsonutil.NewDocElem("z", nil))))
	req.True(ok)
	val, ok = bsonutil.ExtractFieldByName("d.z", testD)
	req.Equal(val, nil)
	req.True(ok)
	val, ok = bsonutil.ExtractFieldByName("d.z.nope", testD)
	req.Equal(val, nil)
	req.False(ok)

	// non-existing fields should return (nil, false)
	for _, c := range []string{"f", "c.nope", "c.nope.NOPE", "b.1000", "b.1.nada"} {
		val, ok = bsonutil.ExtractFieldByName(c, testD)
		req.Nil(val)
		req.False(ok)
	}

	// bsonutil.Extraction of a non-document should return (nil, false)
	val, ok = bsonutil.ExtractFieldByName("meh", bsonutil.NewArray("meh"))
	req.Nil(val)
	req.False(ok)

	// bsonutil.Extraction of a nil document should return (nil, false)
	val, ok = bsonutil.ExtractFieldByName("a", nil)
	req.Equal(val, nil)
	req.False(ok)
}
