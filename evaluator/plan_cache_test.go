package evaluator_test

import (
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/stretchr/testify/require"
)

var (
	_ fmt.Stringer = nil
)

func TestCachePlanStage(t *testing.T) {
	ctx := createTestExecutionCtx(nil)

	t.Run("should not open without rows", func(t *testing.T) {
		cs := &evaluator.CacheStage{}
		_, err := cs.Open(ctx)
		require.Error(t, err)
	})

	t.Run("should iterate through all rows contained successfully", func(t *testing.T) {
		testCache := func(cs *evaluator.CacheStage, ctx *evaluator.ExecutionCtx,
			expected []evaluator.Values) {
			iter, err := cs.Open(ctx)
			require.NoError(t, err)

			row := &evaluator.Row{}
			i := 0
			for iter.Next(row) {
				require.Equal(t, len(row.Data), len(expected[i]))
				require.Equal(t, row.Data, expected[i])
				row = &evaluator.Row{}
				i++
			}
			require.Equal(t, i, len(expected))

			require.NoError(t, iter.Close())
			require.NoError(t, iter.Err())
		}

		expected := []evaluator.Values{
			{{SelectID: 1,
				Database: dbOne, Table: tableOneName, Name: "a",
				Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 1)}},
			{{SelectID: 1,
				Database: dbOne, Table: tableOneName, Name: "a",
				Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 2)}},
			{{SelectID: 1,
				Database: dbOne, Table: tableOneName, Name: "a",
				Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 3)}},
			{{SelectID: 1,
				Database: dbOne, Table: tableOneName, Name: "a",
				Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 4)}},
			{{SelectID: 1,
				Database: dbOne, Table: tableOneName, Name: "a",
				Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 5)}},
			{{SelectID: 1,
				Database: dbOne, Table: tableOneName, Name: "a",
				Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 6)}},
			{{SelectID: 1,
				Database: dbOne, Table: tableOneName, Name: "a",
				Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 7)}},
		}

		var rows []evaluator.Row
		for _, values := range expected {
			rows = append(rows, evaluator.Row{Data: values})
		}

		cs := evaluator.NewCacheStage(0, rows, nil, nil)
		// Iterate through the cache twice to ensure the same values are obtained both times
		testCache(cs, ctx, expected)
		testCache(cs, ctx, expected)
	})
}

func TestCachePlanStageMemoryMonitor(t *testing.T) {
	rows := []evaluator.Row{
		{Data: evaluator.Values{
			{SelectID: 1,
				Database: dbOne, Table: tableOneName, Name: "a",
				Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 1)}}}}
	cs := evaluator.NewCacheStage(0, rows, nil, nil)

	actual := getAllocatedMemorySizeAfterIteration(cs) + getAllocatedMemorySizeAfterIteration(cs)
	expected := 2 * valueSize(
		dbOne, tableOneName, "a",
		evaluator.NewSQLInt64(evaluator.MySQLValueKind, 0),
	)

	require.Equal(t, expected, actual)
}
