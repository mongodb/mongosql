package evaluator_test

import (
	"context"
	"testing"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/collation"
	"github.com/10gen/sqlproxy/internal/util/bsonutil"
	"github.com/stretchr/testify/require"

	"github.com/10gen/mongo-go-driver/bson"
)

func TestSubquerySourceStage(t *testing.T) {

	bgCtx := context.Background()
	execCfg := createTestExecutionCfg()
	execState := evaluator.NewExecutionState()
	oCfg := createOptimizerCfg(collation.Default, execCfg)
	pCfg := createTestPushdownCfg()

	runTest := func(t *testing.T, selectID int, aliasName string, optimize bool, rows []bson.D,
		expectedRows []evaluator.Values) {
		ts := evaluator.NewBSONSourceStage(1, tableOneName, collation.Default, rows)

		var plan evaluator.PlanStage
		var err error

		s := evaluator.NewSubquerySourceStage(ts, selectID, "", aliasName, false)
		plan = s
		if optimize {
			plan, err = evaluator.OptimizePlan(context.Background(), oCfg, plan)
			require.NoError(t, err)
			plan, err = evaluator.PushdownPlan(pCfg, plan)
			require.False(t, err != nil && !evaluator.IsNonFatalPushdownError(err))
		}

		iter, err := plan.Open(bgCtx, execCfg, execState)
		require.NoError(t, err)

		i := 0
		row := &evaluator.Row{}

		for iter.Next(bgCtx, row) {
			require.Equal(t, len(row.Data), len(expectedRows[i]))
			require.Equal(t, row.Data, expectedRows[i])
			row = &evaluator.Row{}
			i++
		}

		require.Equal(t, i, len(expectedRows))

		require.NoError(t, iter.Close())
		require.NoError(t, iter.Err())
	}

	rows := bsonutil.NewDArray(
		bsonutil.NewD(bsonutil.NewDocElem("a", 6), bsonutil.NewDocElem("b", 9)),
		bsonutil.NewD(bsonutil.NewDocElem("a", 3), bsonutil.NewDocElem("b", 4)),
	)

	selectID := 42
	aliasName := "funny"

	kind := evaluator.MySQLValueKind
	expected := []evaluator.Values{
		{{SelectID: 42, Database: evaluator.BSONSourceDB, Table: "funny", Name: "a",
			Data: evaluator.NewSQLInt64(kind, 6)}, {SelectID: 42, Database: evaluator.BSONSourceDB,
			Table: "funny", Name: "b", Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 9)}},
		{{SelectID: 42, Database: evaluator.BSONSourceDB, Table: "funny", Name: "a",
			Data: evaluator.NewSQLInt64(kind, 3)}, {SelectID: 42, Database: evaluator.BSONSourceDB,
			Table: "funny", Name: "b", Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 4)}},
	}

	runTest(t, selectID, aliasName, false, rows, expected)

	t.Run("and should produce identical results after optimization", func(t *testing.T) {
		runTest(t, selectID, aliasName, true, rows, expected)
	})
}
