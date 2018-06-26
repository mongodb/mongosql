package evaluator_test

import (
	"testing"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/stretchr/testify/require"

	"github.com/10gen/mongo-go-driver/bson"
)

func TestSubquerySourceStage(t *testing.T) {
	ctx := createTestExecutionCtx(nil)

	runTest := func(t *testing.T, selectID int, aliasName string, optimize bool, rows []bson.D,
		expectedRows []evaluator.Values) {
		ts := evaluator.NewBSONSourceStage(1, tableOneName, collation.Default, rows)

		var plan evaluator.PlanStage
		var err error

		s := evaluator.NewSubquerySourceStage(ts, selectID, "", aliasName, false)
		plan = s
		if optimize {
			plan = evaluator.OptimizePlan(ctx.ConnectionCtx, plan)
		}

		iter, err := plan.Open(ctx)
		require.NoError(t, err)

		i := 0
		row := &evaluator.Row{}

		for iter.Next(row) {
			require.Equal(t, len(row.Data), len(expectedRows[i]))
			require.Equal(t, row.Data, expectedRows[i])
			row = &evaluator.Row{}
			i++
		}

		require.Equal(t, i, len(expectedRows))

		require.NoError(t, iter.Close())
		require.NoError(t, iter.Err())
	}

	rows := []bson.D{
		{{Name: "a", Value: 6}, {Name: "b", Value: 9}},
		{{Name: "a", Value: 3}, {Name: "b", Value: 4}},
	}

	selectID := 42
	aliasName := "funny"

	expected := []evaluator.Values{
		{{SelectID: 42, Database: evaluator.BSONSourceDB, Table: "funny", Name: "a",
			Data: evaluator.SQLInt(6)}, {SelectID: 42, Database: evaluator.BSONSourceDB,
			Table: "funny", Name: "b", Data: evaluator.SQLInt(9)}},
		{{SelectID: 42, Database: evaluator.BSONSourceDB, Table: "funny", Name: "a",
			Data: evaluator.SQLInt(3)}, {SelectID: 42, Database: evaluator.BSONSourceDB,
			Table: "funny", Name: "b", Data: evaluator.SQLInt(4)}},
	}

	runTest(t, selectID, aliasName, false, rows, expected)

	t.Run("and should produce identical results after optimization", func(t *testing.T) {
		runTest(t, selectID, aliasName, true, rows, expected)
	})
}
