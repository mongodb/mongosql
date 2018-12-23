package evaluator_test

import (
	"context"
	"testing"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/schema"
	"github.com/stretchr/testify/require"
)

func TestRowGeneratorStage(t *testing.T) {
	selectIDs := []int{1}
	newColumn := evaluator.NewColumn(selectIDs[0], "", "", "", "rowCount", "", "rowCount",
		evaluator.EvalUint64, schema.MongoInt64, false)

	bgCtx := context.Background()
	execCfg := createTestExecutionCfg()
	execState := evaluator.NewExecutionState()

	t.Run("should iterate through all rows contained successfully with only empty rows",
		func(t *testing.T) {
			rows := bsonutil.NewDArray()
			bss := evaluator.NewBSONSourceStage(1, "rowCountSource", collation.Default, rows)
			rg := evaluator.NewRowGeneratorStage(bss, newColumn)

			iter, err := rg.Open(bgCtx, execCfg, execState)
			require.NoError(t, err)

			row := &evaluator.Row{}
			i := 0
			for iter.Next(bgCtx, row) {
				require.Nil(t, row.Data)
				i++
			}

			require.Equal(t, i, 0)
			require.NoError(t, iter.Close())
			require.NoError(t, iter.Err())
		})

	t.Run("should iterate through all rows contained successfully with only one row having "+
		"content of different field name", func(t *testing.T) {
		rows := bsonutil.NewDArray(
			bsonutil.NewD(bsonutil.NewDocElem("rowCount1", 2)),
		)

		bss := evaluator.NewBSONSourceStage(1, "rowCountSource", collation.Default, rows)
		rg := evaluator.NewRowGeneratorStage(bss, newColumn)
		iter, err := rg.Open(bgCtx, execCfg, execState)
		require.NoError(t, err)

		row := &evaluator.Row{}
		i := 0
		for iter.Next(bgCtx, row) {
			require.Nil(t, row.Data)
			i++
		}

		require.Equal(t, i, 0)
		require.NoError(t, iter.Close())
		require.Error(t, iter.Err())
	})

	t.Run("should iterate through all rows contained successfully with only one row having"+
		" 0 value", func(t *testing.T) {
		rows := bsonutil.NewDArray(
			bsonutil.NewD(bsonutil.NewDocElem("rowCount", 0)),
		)

		bss := evaluator.NewBSONSourceStage(1, "rowCountSource", collation.Default, rows)
		rg := evaluator.NewRowGeneratorStage(bss, newColumn)
		iter, err := rg.Open(bgCtx, execCfg, execState)
		require.NoError(t, err)

		row := &evaluator.Row{}
		i := 0
		for iter.Next(bgCtx, row) {
			require.Nil(t, row.Data)
			i++
		}

		require.Equal(t, i, 0)
		require.NoError(t, iter.Close())
		require.NoError(t, iter.Err())
	})

	t.Run("should iterate through all rows contained successfully with only one row",
		func(t *testing.T) {
			rows := bsonutil.NewDArray(
				bsonutil.NewD(bsonutil.NewDocElem("rowCount", 5)),
			)

			bss := evaluator.NewBSONSourceStage(1, "rowCountSource", collation.Default, rows)
			rg := evaluator.NewRowGeneratorStage(bss, newColumn)
			iter, err := rg.Open(bgCtx, execCfg, execState)
			require.NoError(t, err)

			row := &evaluator.Row{}
			i := 0
			for iter.Next(bgCtx, row) {
				require.Nil(t, row.Data)
				i++
			}

			require.Equal(t, i, 5)
			require.NoError(t, iter.Close())
			require.NoError(t, iter.Err())
		})

	t.Run("should iterate through all rows contained successfully with two rows",
		func(t *testing.T) {
			rows := bsonutil.NewDArray(
				bsonutil.NewD(bsonutil.NewDocElem("rowCount", 5)),
				bsonutil.NewD(bsonutil.NewDocElem("rowCount", 2)),
			)

			bss := evaluator.NewBSONSourceStage(1, "rowCountSource", collation.Default, rows)
			rg := evaluator.NewRowGeneratorStage(bss, newColumn)
			iter, err := rg.Open(bgCtx, execCfg, execState)
			require.NoError(t, err)

			row := &evaluator.Row{}
			i := 0
			for iter.Next(bgCtx, row) {
				require.Nil(t, row.Data)
				i++
			}

			require.Equal(t, i, 7)
			require.NoError(t, iter.Close())
			require.NoError(t, iter.Err())
		})

	t.Run("should iterate through all rows contained successfully with only two rows with"+
		" at least one row with 0 value", func(t *testing.T) {
		rows := bsonutil.NewDArray(
			bsonutil.NewD(bsonutil.NewDocElem("rowCount", 0)),
			bsonutil.NewD(bsonutil.NewDocElem("rowCount", 2)),
		)

		bss := evaluator.NewBSONSourceStage(1, "rowCountSource", collation.Default, rows)
		rg := evaluator.NewRowGeneratorStage(bss, newColumn)
		iter, err := rg.Open(bgCtx, execCfg, execState)
		require.NoError(t, err)

		row := &evaluator.Row{}
		i := 0
		for iter.Next(bgCtx, row) {
			require.Nil(t, row.Data)
			i++
		}

		require.Equal(t, i, 2)
		require.NoError(t, iter.Close())
		require.NoError(t, iter.Err())
	})

	t.Run("should clear each row's data upon iteration", func(t *testing.T) {
		rows := bsonutil.NewDArray(
			bsonutil.NewD(bsonutil.NewDocElem("rowCount", 3)),
		)

		bss := evaluator.NewBSONSourceStage(1, "rowCountSource", collation.Default, rows)
		rg := evaluator.NewRowGeneratorStage(bss, newColumn)
		iter, err := rg.Open(bgCtx, execCfg, execState)
		require.NoError(t, err)

		kind := evaluator.MySQLValueKind
		row := &evaluator.Row{Data: evaluator.Values{
			evaluator.NewValue(0, "test", "foo", "a", evaluator.NewSQLInt64(kind, 1))}}
		i := 0
		for iter.Next(bgCtx, row) {
			require.Nil(t, row.Data)
			i++
			row = &evaluator.Row{Data: evaluator.Values{
				evaluator.NewValue(0, "test", "foo", "a", evaluator.NewSQLInt64(kind, 1))}}
		}

		require.Equal(t, i, 3)
		require.NoError(t, iter.Close())
		require.NoError(t, iter.Err())
	})
}
