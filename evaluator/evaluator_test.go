package evaluator_test

import (
	"context"
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/memory"
	"github.com/10gen/sqlproxy/internal/testutils/dbutils"
	"github.com/10gen/sqlproxy/internal/util/bsonutil"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/parser"
	"github.com/stretchr/testify/require"
)

func TestMemoryZeroSum(t *testing.T) {
	cfgOne := setupEnv().cfgOne
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

	rows := bsonutil.NewDArray(
		bsonutil.NewD(
			bsonutil.NewDocElem("_id", "5"),
			bsonutil.NewDocElem("a", 6),
			bsonutil.NewDocElem("b", 7),
			bsonutil.NewDocElem("d", 8),
		),
		bsonutil.NewD(
			bsonutil.NewDocElem("_id", "15"),
			bsonutil.NewDocElem("a", 16),
			bsonutil.NewDocElem("b", 17),
			bsonutil.NewDocElem("d", 18),
		),
	)

	dbutils.DropCollection(session, dbOne, tableTwoName)
	dbutils.InsertDocuments(session, dbOne, tableTwoName, rows)
	defer dbutils.DropCollection(session, dbOne, tableTwoName)

	test := func(t *testing.T, sql string) {
		stmt, err := parser.Parse(sql)
		require.NoError(t, err)

		bgCtx := context.Background()
		lg := log.GlobalLogger()
		monitor := memory.NewMonitor("evaluator_unit_test_monitor", 0)
		aCfg := evaluator.NewAlgebrizerConfig(lg, sql, stmt, dbOne, catalogOne)
		pCfg := evaluator.NewPushdownConfig(lg, variablesOne)
		eCfg := createWorkingExecutionCfg(variablesOne, session, monitor)
		oCfg := evaluator.NewOptimizerConfig(lg, variablesOne, eCfg)

		res, err := evaluator.EvaluateQuery(bgCtx, aCfg, oCfg, pCfg, eCfg)
		require.NoError(t, err)

		switch typedIter := res.Iter.(type) {
		case evaluator.Iter:
			row := &evaluator.Row{}
			for typedIter.Next(bgCtx, row) {
			}
		case evaluator.FastIter:
			row := &bson.RawD{}
			for typedIter.Next(bgCtx, row) {
			}
		}
		res.Iter.Close()

		require.Zero(t, monitor.Allocated())
	}

	tests := []struct {
		name, sql string
	}{
		{"subquery_non_correlated", "SELECT a + a, (SELECT a FROM bar LIMIT 1) FROM foo ORDER BY a LIMIT 2, 1"},
		{"subquery_correlated", "SELECT a + a, (SELECT foo.a FROM bar LIMIT 1) FROM foo ORDER BY a LIMIT 2, 1"},
		{"subquery_where", "SELECT a, b FROM foo WHERE a IN(SELECT a FROM bar)"},
		{"join", "SELECT * FROM foo, bar"},
		{"join_limit", "SELECT * FROM foo, bar limit 3, 2"},
		{"group", "SELECT a, b FROM foo GROUP BY a"},
		{"group_limit", "SELECT a, b FROM foo GROUP BY a limit 1,1"},
		{"union", "SELECT a, b FROM foo UNION SELECT a, b FROM bar"},
		{"union_limit", "(SELECT a, b FROM foo UNION SELECT a, b FROM bar) limit 2, 1"},
		{"rowgenerator", "SELECT HOUR(null) FROM foo"},
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			test(t, tst.sql)
		})
	}
}
