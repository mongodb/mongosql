package evaluator_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/collation"
	"github.com/10gen/sqlproxy/internal/util/bsonutil"
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

		req := require.New(t)

		bgCtx := context.Background()
		execCfg := createTestExecutionCfg()
		execState := evaluator.NewExecutionState()

		ts := evaluator.NewBSONSourceStage(1, tableOneName, collation.Default, rows)

		l := evaluator.NewLimitStage(ts, offset, limit)

		iter, err := l.Open(bgCtx, execCfg, execState)
		req.NoError(err)

		row := &evaluator.Row{}

		i := 0

		for iter.Next(bgCtx, row) {
			req.Equal(len(row.Data), len(expectedRows[i]))
			req.Equal(row.Data, expectedRows[i])
			row = &evaluator.Row{}
			i++
		}

		req.Equal(i, len(expectedRows))

		req.NoError(iter.Close())
		req.NoError(iter.Err())
	}

	rows := bsonutil.NewDArray(
		bsonutil.NewD(bsonutil.NewDocElem("a", 1)),
		bsonutil.NewD(bsonutil.NewDocElem("a", 2)),
		bsonutil.NewD(bsonutil.NewDocElem("a", 3)),
		bsonutil.NewD(bsonutil.NewDocElem("a", 4)),
		bsonutil.NewD(bsonutil.NewDocElem("a", 5)),
		bsonutil.NewD(bsonutil.NewDocElem("a", 6)),
		bsonutil.NewD(bsonutil.NewDocElem("a", 7)),
	)

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
