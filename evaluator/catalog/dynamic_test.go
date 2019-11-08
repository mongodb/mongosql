package catalog_test

import (
	"context"
	"testing"

	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/stretchr/testify/require"
)

func TestDynamic(t *testing.T) {
	req := require.New(t)

	testTable := catalog.NewDynamicTable("bar", "foo", func(string) results.RowIter {
		rowChan := make(chan results.Row, results.DefaultRowChannelBufSize)
		done := make(chan struct{})
		// Since channel is blocking, it's put in a goroutine
		go func() {
			defer close(rowChan)

			for i := int64(0); i < 10; i += 2 {
				select {
				case rowChan <- results.NewNamedRow("bar", "foo",
					values.NewNamedSQLValue("one", values.NewSQLInt64(values.MongoSQLValueKind, i)),
					values.NewNamedSQLValue("two", values.NewSQLInt64(values.MongoSQLValueKind, i+1)),
				):
				case <-done:
					return
				}
			}
		}()

		return results.NewRowChanIter(rowChan, done)
	})

	_, err := testTable.AddColumn("foo", "one", types.EvalInt64)
	req.Nil(err)
	_, err = testTable.AddColumn("foo", "two", types.EvalInt64)
	req.Nil(err)

	req.Equal(testTable.Columns()[0].Name, "one")
	req.Equal(testTable.Columns()[0].SelectID, 1)
	req.Equal(testTable.Columns()[1].Name, "two")
	req.Equal(testTable.Columns()[1].SelectID, 2)

	iter := testTable.Rows("")

	row := &results.Row{}
	for iter.Next(context.TODO(), row) {
		req.Equal(row.Data[0].SelectID, 1)
		req.Equal(row.Data[1].SelectID, 2)
		req.Equal(row.Data[0].Name, "one")
		req.Equal(row.Data[1].Name, "two")
	}

	require.NoError(t, iter.Close())
	require.NoError(t, iter.Err())
}

func TestDynamicPartialRead(t *testing.T) {
	req := require.New(t)

	testTable := catalog.NewDynamicTable("bar", "foo", func(string) results.RowIter {
		rowChan := make(chan results.Row, results.DefaultRowChannelBufSize)
		done := make(chan struct{})
		// Since channel is blocking, it's put in a goroutine
		go func() {
			defer close(rowChan)

			for i := int64(0); i < 10; i += 2 {
				select {
				case rowChan <- results.NewNamedRow("bar", "foo",
					values.NewNamedSQLValue("one", values.NewSQLInt64(values.MongoSQLValueKind, i)),
					values.NewNamedSQLValue("two", values.NewSQLInt64(values.MongoSQLValueKind, i+1)),
				):
				case <-done:
					return
				}
			}
		}()

		return results.NewRowChanIter(rowChan, done)
	})

	_, err := testTable.AddColumn("foo", "one", types.EvalInt64)
	req.Nil(err)
	_, err = testTable.AddColumn("foo", "two", types.EvalInt64)
	req.Nil(err)

	req.Equal(testTable.Columns()[0].Name, "one")
	req.Equal(testTable.Columns()[0].SelectID, 1)
	req.Equal(testTable.Columns()[1].Name, "two")
	req.Equal(testTable.Columns()[1].SelectID, 2)

	iter := testTable.Rows("")

	row := &results.Row{}
	iter.Next(context.TODO(), row)
	req.Equal(row.Data[0].SelectID, 1)
	req.Equal(row.Data[1].SelectID, 2)
	req.Equal(row.Data[0].Name, "one")
	req.Equal(row.Data[1].Name, "two")
	require.NoError(t, iter.Close())
	require.False(t, iter.Next(context.TODO(), row))
	require.NoError(t, iter.Err())

	// Should be able to recover when close is called again
	require.NoError(t, iter.Close())
	require.NoError(t, iter.Err())
}

func TestEmptyDynamicTable(t *testing.T) {
	req := require.New(t)

	testTable := catalog.NewDynamicTable("bar", "foo", results.NewEmptyRowChanIter)

	iter := testTable.Rows("")
	row := &results.Row{}
	req.False(iter.Next(context.TODO(), row))
}
