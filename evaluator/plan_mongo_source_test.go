package evaluator_test

import (
	"context"
	"os"
	"testing"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/catalog"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/memory"
	"github.com/10gen/sqlproxy/internal/testutils/dbutils"
	mongoutil "github.com/10gen/sqlproxy/internal/testutils/mongodb"
	"github.com/10gen/sqlproxy/internal/util/bsonutil"
	"github.com/10gen/sqlproxy/mongodb"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/require"

	"github.com/10gen/mongo-go-driver/bson"
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

	rows := []bson.D{
		{
			bson.DocElem{Name: "_id", Value: "5"},
			bson.DocElem{Name: "a", Value: 6},
			bson.DocElem{Name: "b", Value: 7},
			bson.DocElem{Name: "d", Value: 8},
		},
		{
			bson.DocElem{Name: "_id", Value: "15"},
			bson.DocElem{Name: "a", Value: 16},
			bson.DocElem{Name: "b", Value: 17},
			bson.DocElem{Name: "d", Value: 18},
		},
	}

	var expected []evaluator.Values
	var values []evaluator.Value
	for _, document := range rows {
		values, err = bsonDToValues(1, dbOne, tableTwoName, document)
		require.NoError(t, err)
		expected = append(expected, values)
	}

	dbutils.DropCollection(session, dbOne, tableTwoName)
	dbutils.InsertDocuments(session, dbOne, tableTwoName, rows)
	defer dbutils.DropCollection(session, dbOne, tableTwoName)

	db, err := catalogOne.Database(dbOne)
	if err != nil {
		panic("database doesn't exist")
	}
	table, err := db.Table(tableTwoName)
	if err != nil {
		panic("table doesn't exist")
	}

	t.Run("with no memory limit", func(t *testing.T) {
		bgCtx := context.Background()
		monitor := memory.NewMonitor("evaluator_unit_test_monitor", 0)
		execCfg := createWorkingExecutionCfg(variablesOne, session, monitor, dbOne)
		execState := evaluator.NewExecutionState()

		plan := evaluator.NewMongoSourceStage(db, table.(*catalog.MongoTable), 1, "")
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
		execCfg := createWorkingExecutionCfg(variablesOne, session, monitor, dbOne)
		execState := evaluator.NewExecutionState()

		plan := evaluator.NewMongoSourceStage(db, table.(*catalog.MongoTable), 1, "")
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

	testD := bson.D{
		{Name: "a", Value: "string"},
		{Name: "b", Value: []interface{}{"inner", bson.D{{Name: "inner2", Value: 1}}}},
		{Name: "c", Value: bson.D{{Name: "x", Value: 5}}},
		{Name: "d", Value: bson.D{{Name: "z", Value: nil}}},
	}

	// regular fields should be extracted by name
	val, ok := bsonutil.ExtractFieldByName("a", testD)
	req.Equal(val, "string")
	req.True(ok)

	// array fields should be extracted by name
	val, ok = bsonutil.ExtractFieldByName("b.1", testD)
	req.Zero(convey.ShouldResemble(val, bson.D{{Name: "inner2", Value: 1}}))
	req.True(ok)
	val, ok = bsonutil.ExtractFieldByName("b.1.inner2", testD)
	req.Equal(val, 1)
	req.True(ok)
	val, ok = bsonutil.ExtractFieldByName("b.0", testD)
	req.Equal(val, "inner")
	req.True(ok)

	// subdocument fields should be extracted by name
	val, ok = bsonutil.ExtractFieldByName("c", testD)
	req.Zero(convey.ShouldResemble(val, bson.D{{Name: "x", Value: 5}}))
	req.True(ok)
	val, ok = bsonutil.ExtractFieldByName("c.x", testD)
	req.Equal(val, 5)
	req.True(ok)

	// even if they contain null values
	val, ok = bsonutil.ExtractFieldByName("d", testD)
	req.Zero(convey.ShouldResemble(val, bson.D{{Name: "z", Value: nil}}))
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
	val, ok = bsonutil.ExtractFieldByName("meh", []interface{}{"meh"})
	req.Nil(val)
	req.False(ok)

	// bsonutil.Extraction of a nil document should return (nil, false)
	val, ok = bsonutil.ExtractFieldByName("a", nil)
	req.Equal(val, nil)
	req.False(ok)
}
