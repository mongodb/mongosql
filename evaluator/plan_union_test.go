package evaluator_test

import (
	"context"
	"testing"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/collation"
	"github.com/10gen/sqlproxy/internal/util/bsonutil"
	"github.com/stretchr/testify/require"
)

var (
	SQL1  = evaluator.NewSQLInt64(evaluator.MySQLValueKind, 1)
	SQL2  = evaluator.NewSQLInt64(evaluator.MySQLValueKind, 2)
	SQL3  = evaluator.NewSQLInt64(evaluator.MySQLValueKind, 3)
	SQL4  = evaluator.NewSQLInt64(evaluator.MySQLValueKind, 4)
	SQL5  = evaluator.NewSQLInt64(evaluator.MySQLValueKind, 5)
	SQL6  = evaluator.NewSQLInt64(evaluator.MySQLValueKind, 6)
	SQL7  = evaluator.NewSQLInt64(evaluator.MySQLValueKind, 7)
	SQL8  = evaluator.NewSQLInt64(evaluator.MySQLValueKind, 8)
	SQL9  = evaluator.NewSQLInt64(evaluator.MySQLValueKind, 9)
	SQL10 = evaluator.NewSQLInt64(evaluator.MySQLValueKind, 10)
	SQL11 = evaluator.NewSQLInt64(evaluator.MySQLValueKind, 11)
	SQL12 = evaluator.NewSQLInt64(evaluator.MySQLValueKind, 12)
	SQL13 = evaluator.NewSQLInt64(evaluator.MySQLValueKind, 13)
	SQL14 = evaluator.NewSQLInt64(evaluator.MySQLValueKind, 14)
	SQL15 = evaluator.NewSQLInt64(evaluator.MySQLValueKind, 15)
	SQL16 = evaluator.NewSQLInt64(evaluator.MySQLValueKind, 16)
)

var (
	basicTable1 = bsonutil.NewDArray(
		bsonutil.NewD(
			bsonutil.NewDocElem("a", 1),
			bsonutil.NewDocElem("b", 2),
		),
		bsonutil.NewD(
			bsonutil.NewDocElem("a", 3),
			bsonutil.NewDocElem("b", 4),
		),
		bsonutil.NewD(
			bsonutil.NewDocElem("a", 5),
			bsonutil.NewDocElem("b", 6),
		),
		bsonutil.NewD(
			bsonutil.NewDocElem("a", 7),
			bsonutil.NewDocElem("b", 8),
		),
	)

	basicTable2 = bsonutil.NewDArray(
		bsonutil.NewD(
			bsonutil.NewDocElem("c", 9),
			bsonutil.NewDocElem("d", 10),
		),
		bsonutil.NewD(
			bsonutil.NewDocElem("c", 11),
			bsonutil.NewDocElem("d", 12),
		),
		bsonutil.NewD(
			bsonutil.NewDocElem("c", 13),
			bsonutil.NewDocElem("d", 14),
		),
		bsonutil.NewD(
			bsonutil.NewDocElem("c", 15),
			bsonutil.NewDocElem("d", 16),
		),
	)
)

type result map[string]interface{}

func containsRow(t *testing.T, results []result, row *evaluator.Row) ([]result, bool) {
	toRemove := -1

	contains := false
	for i, result := range results {
		matches := true
		for _, value := range row.Data {
			resultVal, ok := result[value.Name]
			require.True(t, ok)
			if resultVal != value.Data {
				matches = false
				break
			}
		}

		if matches {
			toRemove = i
			contains = true
			break
		}
	}

	// necessary for multiset equality
	if toRemove >= 0 {
		results = append(results[:toRemove], results[toRemove+1:]...)
	}

	return results, contains
}

func TestUnionPlanStage(t *testing.T) {

	bgCtx := context.Background()
	execCfg := createTestExecutionCfg()
	execState := evaluator.NewExecutionState()

	test := func(t *testing.T, expectedColumns []string, expectedResults []result,
		planStageFactory func() evaluator.PlanStage) {

		row := &evaluator.Row{}

		unionStage := planStageFactory()
		iter, err := unionStage.Open(bgCtx, execCfg, execState)
		require.NoError(t, err)

		columns := unionStage.Columns()
		require.Equal(t, len(expectedColumns), len(columns))
		for i, col := range columns {
			require.Equal(t, col.Name, expectedColumns[i])
		}

		for iter.Next(bgCtx, row) {
			trimmed, contains := containsRow(t, expectedResults, row)
			expectedResults = trimmed
			require.True(t, contains)
		}
		require.Empty(t, expectedResults)

		err = iter.Err()
		require.NoError(t, err)

		err = iter.Close()
		require.NoError(t, err)
	}

	t.Run("a b before c d", func(t *testing.T) {
		test(t, []string{"a", "b"},
			[]result{{"a": SQL1, "b": SQL2}, {"a": SQL3, "b": SQL4},
				{"a": SQL5, "b": SQL6}, {"a": SQL7, "b": SQL8},
				{"a": SQL9, "b": SQL10}, {"a": SQL11, "b": SQL12},
				{"a": SQL13, "b": SQL14}, {"a": SQL15, "b": SQL16}},
			func() evaluator.PlanStage {
				return evaluator.NewUnionStage(evaluator.UnionDistinct,
					evaluator.NewBSONSourceStage(1, "foo", collation.Default, basicTable1),
					evaluator.NewBSONSourceStage(2, "bar", collation.Default, basicTable2),
				)
			})
	})

	t.Run("c d before a b", func(t *testing.T) {
		test(t, []string{"c", "d"},
			[]result{{"c": SQL1, "d": SQL2}, {"c": SQL3, "d": SQL4},
				{"c": SQL5, "d": SQL6}, {"c": SQL7, "d": SQL8},
				{"c": SQL9, "d": SQL10}, {"c": SQL11, "d": SQL12},
				{"c": SQL13, "d": SQL14}, {"c": SQL15, "d": SQL16}},
			func() evaluator.PlanStage {
				return evaluator.NewUnionStage(evaluator.UnionDistinct,
					evaluator.NewBSONSourceStage(1, "foo", collation.Default, basicTable2),
					evaluator.NewBSONSourceStage(2, "bar", collation.Default, basicTable1),
				)
			})
	})
}
