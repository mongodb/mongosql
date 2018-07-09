package evaluator_test

import (
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/mongodb"
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
	basicTable1 = []bson.D{
		{
			bson.DocElem{Name: "a", Value: 1},
			bson.DocElem{Name: "b", Value: 2},
		},
		{
			bson.DocElem{Name: "a", Value: 3},
			bson.DocElem{Name: "b", Value: 4},
		},
		{
			bson.DocElem{Name: "a", Value: 5},
			bson.DocElem{Name: "b", Value: 6},
		},
		{
			bson.DocElem{Name: "a", Value: 7},
			bson.DocElem{Name: "b", Value: 8},
		},
	}

	basicTable2 = []bson.D{
		{
			bson.DocElem{Name: "c", Value: 9},
			bson.DocElem{Name: "d", Value: 10},
		},
		{
			bson.DocElem{Name: "c", Value: 11},
			bson.DocElem{Name: "d", Value: 12},
		},
		{
			bson.DocElem{Name: "c", Value: 13},
			bson.DocElem{Name: "d", Value: 14},
		},
		{
			bson.DocElem{Name: "c", Value: 15},
			bson.DocElem{Name: "d", Value: 16},
		},
	}
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

	testSchema := evaluator.MustLoadSchema(testSchema4)
	testInfo := evaluator.GetMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)

	ctx := createTestExecutionCtx(testInfo)

	test := func(t *testing.T, expectedColumns []string, expectedResults []result,
		planStageFactory func() evaluator.PlanStage) {
		row := &evaluator.Row{}

		unionStage := planStageFactory()
		iter, err := unionStage.Open(ctx)
		require.NoError(t, err)

		columns := unionStage.Columns()
		require.Equal(t, len(expectedColumns), len(columns))
		for i, col := range columns {
			require.Equal(t, col.Name, expectedColumns[i])
		}

		for iter.Next(row) {
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

func TestUnionStageMemoryMonitor(t *testing.T) {
	rows := []bson.D{
		{{Name: "a", Value: 6}, {Name: "b", Value: 9}},
		{{Name: "a", Value: 3}, {Name: "b", Value: 4}},
	}
	u := evaluator.NewUnionStage(evaluator.UnionDistinct,
		evaluator.NewBSONSourceStage(1, "foo", collation.Default, rows),
		evaluator.NewBSONSourceStage(2, "bar", collation.Default, rows),
	)

	actual := getAllocatedMemorySizeAfterIteration(u)

	sizeA := valueSize(
		evaluator.BSONSourceDB, tableOneName, "a",
		evaluator.NewSQLInt64(evaluator.MySQLValueKind, 0),
	)
	sizeB := valueSize(
		evaluator.BSONSourceDB, tableOneName, "b",
		evaluator.NewSQLInt64(evaluator.MySQLValueKind, 0),
	)
	expected := uint64(len(rows)*2) * (sizeA + sizeB)

	require.Equal(t, expected, actual)
}
