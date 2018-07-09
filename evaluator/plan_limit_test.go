package evaluator_test

import (
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/schema"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/stretchr/testify/require"
)

var (
	_      fmt.Stringer = nil
	limit  uint64
	offset uint64
)

func TestLimitPlanStage(t *testing.T) {
	runTest := func(t *testing.T,
		limit uint64,
		offset uint64,
		rows []bson.D,
		expectedRows []evaluator.Values) {

		ctx := createTestExecutionCtx(nil)

		ts := evaluator.NewBSONSourceStage(1, tableOneName, collation.Default, rows)

		l := evaluator.NewLimitStage(ts, offset, limit)

		iter, err := l.Open(ctx)
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

	rows := []bson.D{
		{{Name: "a", Value: 1}},
		{{Name: "a", Value: 2}},
		{{Name: "a", Value: 3}},
		{{Name: "a", Value: 4}},
		{{Name: "a", Value: 5}},
		{{Name: "a", Value: 6}},
		{{Name: "a", Value: 7}},
	}

	t.Run("should return only 'limit' records if the limit is less than the total number of"+
		" records", func(t *testing.T) {

		limit = 2
		offset = 0

		expected := []evaluator.Values{
			{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
				Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 1)}},
			{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
				Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 2)}},
		}

		runTest(t, limit, offset, rows, expected)
	})

	t.Run("should return the right slice of the records with an offset leaving less records"+
		" than the limit covers", func(t *testing.T) {

		limit = 2
		offset = 4

		expected := []evaluator.Values{
			{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
				Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 5)}},
			{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
				Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 6)}},
		}

		runTest(t, limit, offset, rows, expected)
	})

	t.Run("should return no records if the offset is greater than the number of records",
		func(t *testing.T) {
			limit = 2
			offset = 40

			expected := []evaluator.Values{}

			runTest(t, limit, offset, rows, expected)
		})

	t.Run("should return no records if the limit and offset are both greater than the "+
		"number of records", func(t *testing.T) {

		limit = 40
		offset = 40
		expected := []evaluator.Values{}

		runTest(t, limit, offset, rows, expected)
	})

	t.Run("should only return the number of records if the limit is greater than the number "+
		"of records", func(t *testing.T) {

		limit = 40
		offset = 0

		expected := []evaluator.Values{
			{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
				Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 1)}},
			{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
				Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 2)}},
			{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
				Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 3)}},
			{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
				Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 4)}},
			{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
				Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 5)}},
			{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
				Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 6)}},
			{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
				Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 7)}},
		}

		runTest(t, limit, offset, rows, expected)
	})

	t.Run("should return the one record if the limit is 1 with an offset of 1", func(t *testing.T) {
		limit = 1
		offset = 1

		expected := []evaluator.Values{{{SelectID: 1, Database: evaluator.BSONSourceDB,
			Table: tableOneName, Name: "a",
			Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 2)}}}

		runTest(t, limit, offset, rows, expected)

	})

	t.Run("should return the one record if all but the last record is skipped", func(t *testing.T) {

		limit = 1
		offset = 6

		expected := []evaluator.Values{{{SelectID: 1, Database: evaluator.BSONSourceDB,
			Table: tableOneName, Name: "a",
			Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 7)}}}

		runTest(t, limit, offset, rows, expected)
	})

	t.Run("should return no records with a limit and offset of 0", func(t *testing.T) {

		limit = 0
		offset = 0

		expected := []evaluator.Values{}

		runTest(t, limit, offset, rows, expected)
	})

	t.Run("should return only 'limit' records if there is no offset", func(t *testing.T) {

		limit = 3
		offset = 0

		expected := []evaluator.Values{
			{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
				Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 1)}},
			{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
				Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 2)}},
			{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
				Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 3)}}}

		runTest(t, limit, offset, rows, expected)
	})

	t.Run("should return no records if the limit is 0", func(t *testing.T) {

		limit = 0
		offset = 4

		expected := []evaluator.Values{}

		runTest(t, limit, offset, rows, expected)
	})
}

func TestLimitStageMemoryMonitor(t *testing.T) {
	rows := []bson.D{
		{{Name: "a", Value: 6}, {Name: "b", Value: 9}},
		{{Name: "a", Value: 3}, {Name: "b", Value: 4}},
		{{Name: "a", Value: 0}, {Name: "b", Value: 13}},
		{{Name: "a", Value: -3}, {Name: "b", Value: 8}},
		{{Name: "a", Value: -6}, {Name: "b", Value: 17}},
		{{Name: "a", Value: -9}, {Name: "b", Value: 12}},
	}

	t.Run("non-blocking source", func(t *testing.T) {
		bss := evaluator.NewBSONSourceStage(1, tableTwoName, collation.Default, rows)
		ls := evaluator.NewLimitStage(bss, 2, 2)

		actual := getAllocatedMemorySizeAfterIteration(ls)

		sizeA := valueSize(
			evaluator.BSONSourceDB, tableTwoName, "a",
			evaluator.NewSQLInt64(evaluator.MySQLValueKind, 0),
		)
		sizeB := valueSize(
			evaluator.BSONSourceDB, tableTwoName, "b",
			evaluator.NewSQLInt64(evaluator.MySQLValueKind, 0),
		)
		expected := 2 * (sizeA + sizeB)

		require.Equal(t, expected, actual)
	})
	t.Run("blocking source", func(t *testing.T) {
		bss := evaluator.NewBSONSourceStage(1, tableTwoName, collation.Default, rows)
		os := evaluator.NewOrderByStage(bss,
			evaluator.NewOrderByTerm(
				evaluator.NewSQLColumnExpr(
					1,
					evaluator.BSONSourceDB,
					tableTwoName,
					"a",
					evaluator.EvalInt64,
					schema.MongoInt),
				true))
		ls := evaluator.NewLimitStage(os, 2, 2)

		actual := getAllocatedMemorySizeAfterIteration(ls)

		sizeA := valueSize(
			evaluator.BSONSourceDB, tableTwoName, "a",
			evaluator.NewSQLInt64(evaluator.MySQLValueKind, 0),
		)
		sizeB := valueSize(
			evaluator.BSONSourceDB, tableTwoName, "b",
			evaluator.NewSQLInt64(evaluator.MySQLValueKind, 0),
		)
		expected := 4 * (sizeA + sizeB)

		require.Equal(t, expected, actual)
	})
}
