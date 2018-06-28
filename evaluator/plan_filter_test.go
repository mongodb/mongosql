package evaluator_test

import (
	"fmt"
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/stretchr/testify/require"
)

var (
	_ fmt.Stringer = nil
)

func TestFilterPlanStage(t *testing.T) {
	runTest := func(t *testing.T,
		matcher evaluator.SQLExpr,
		rows []bson.D,
		expectedRows []evaluator.Values) {

		ctx := createTestExecutionCtx(nil)

		bss := evaluator.NewBSONSourceStage(1, tableTwoName, collation.Default, rows)
		filter := evaluator.NewFilterStage(bss, matcher)

		iter, err := filter.Open(ctx)

		require.NoError(t, err)

		row := &evaluator.Row{}

		i := 0

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

	schema := evaluator.MustLoadSchema(testSchema3)

	rows := []bson.D{
		{{Name: "a", Value: 6}, {Name: "b", Value: 7}, {Name: "_id", Value: 5}},
		{{Name: "a", Value: 16}, {Name: "b", Value: 17}, {Name: "_id", Value: 15}},
	}

	queries := []string{
		"a = 16",
		"a = 6",
		"a = 99",
		"b > 9",
		"b > 9 or a < 5",
		"b = 7 or a = 6",
	}

	r0, err := bsonDToValues(1, evaluator.BSONSourceDB, tableTwoName, rows[0])
	require.NoError(t, err)
	r1, err := bsonDToValues(1, evaluator.BSONSourceDB, tableTwoName, rows[1])
	require.NoError(t, err)

	expected := [][]evaluator.Values{{r1}, {r0}, nil, {r1}, {r1}, {r0}}

	for i, query := range queries {
		matcher, err := evaluator.GetSQLExpr(schema, evaluator.BSONSourceDB, tableTwoName, query)
		require.NoError(t, err)

		t.Run(query, func(t *testing.T) {
			runTest(t, matcher, rows, expected[i])
		})
	}
}

func TestFilterStageMemoryMonitor(t *testing.T) {
	schema := evaluator.MustLoadSchema(testSchema3)
	rows := []bson.D{
		{{Name: "a", Value: 6}, {Name: "b", Value: 9}},
		{{Name: "a", Value: 3}, {Name: "b", Value: 4}},
	}
	matcher, err := evaluator.GetSQLExpr(schema,
		evaluator.BSONSourceDB,
		tableTwoName,
		"a = 6")
	require.NoError(t, err)

	bss := evaluator.NewBSONSourceStage(1, tableTwoName, collation.Default, rows)
	filter := evaluator.NewFilterStage(bss, matcher)

	actual := getAllocatedMemorySizeAfterIteration(filter)
	expected := valueSize(evaluator.BSONSourceDB, tableTwoName, "a", evaluator.SQLInt64(0)) +
		valueSize(evaluator.BSONSourceDB, tableTwoName, "b", evaluator.SQLInt64(0))

	require.Equal(t, expected, actual)
}
