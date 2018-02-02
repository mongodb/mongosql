package evaluator_test

import (
	"context"
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/memory"
	"github.com/10gen/sqlproxy/internal/testutils/dbutils"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/parser"
	"github.com/stretchr/testify/require"
)

func TestEvaluateQuery_should_result_in_0_memory(t *testing.T) {
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

	dbutils.DropCollection(session, dbOne, tableTwoName)
	dbutils.InsertDocuments(session, dbOne, tableTwoName, rows)

	test := func(t *testing.T, sql string) {
		stmt, err := parser.Parse(sql)
		require.NoError(t, err)

		ctx := &connCtx{
			catalog:       catalogOne,
			db:            dbOne,
			memoryMonitor: memory.NewMonitor("test", 0),
			session:       session,
			variables:     variablesOne,
		}

		_, iter, err := evaluator.EvaluateQuery(sql, stmt, ctx)
		require.NoError(t, err)

		switch typedIter := iter.(type) {
		case evaluator.Iter:
			row := &evaluator.Row{}
			for typedIter.Next(row) {
			}
		case evaluator.FastIter:
			row := &bson.RawD{}
			for typedIter.Next(row) {
			}
		}
		iter.Close()

		require.Zero(t, ctx.memoryMonitor.Allocated())
	}

	tests := []string{
		"SELECT a + a, (SELECT a FROM bar LIMIT 1) FROM foo ORDER BY a LIMIT 2, 1",
		"SELECT a + a, (SELECT foo.a FROM bar LIMIT 1) FROM foo ORDER BY a LIMIT 2, 1",
		"SELECT a, b FROM foo WHERE a IN(SELECT a FROM bar)",
		"SELECT * FROM foo, bar",
		"SELECT * FROM foo, bar limit 3, 2",
		"SELECT a, b FROM foo GROUP BY a",
		"SELECT a, b FROM foo GROUP BY a limit 1,1",
		"SELECT a, b FROM foo UNION SELECT a, b FROM bar",
		"(SELECT a, b FROM foo UNION SELECT a, b FROM bar) limit 2, 1",
		"SELECT HOUR(null) FROM foo",
	}

	for _, sql := range tests {
		t.Run(sql, func(t *testing.T) {
			test(t, sql)
		})
	}
}
