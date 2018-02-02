package evaluator_test

import (
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/schema"
	"github.com/stretchr/testify/require"
)

func TestRowGeneratorStage(t *testing.T) {
	selectIDs := []int{1}
	newColumn := evaluator.NewColumn(selectIDs[0], "", "", "", "rowCount", "", "rowCount",
		schema.SQLUint64, schema.MongoInt64, false)
	ctx := &evaluator.ExecutionCtx{ConnectionCtx: createTestConnectionCtx(nil)}

	t.Run("should iterate through all rows contained successfully with only empty rows",
		func(t *testing.T) {
			rows := []evaluator.Row{}
			cs := evaluator.NewCacheStage(0, rows, nil, nil)
			rg := evaluator.NewRowGeneratorStage(cs, newColumn)

			iter, err := rg.Open(ctx)
			require.NoError(t, err)

			row := &evaluator.Row{}
			i := 0
			for iter.Next(row) {
				require.Nil(t, row.Data)
				i++
			}

			require.Equal(t, i, 0)
			require.NoError(t, iter.Close())
			require.NoError(t, iter.Err())
		})

	t.Run("should iterate through all rows contained successfully with only one row having "+
		"content of different field name", func(t *testing.T) {
		rows := []evaluator.Row{
			{Data: evaluator.Values{{SelectID: 1, Database: "test1", Table: "test",
				Name: "rowCount1", Data: evaluator.SQLInt(2)}}},
		}
		cs := evaluator.NewCacheStage(0, rows, nil, nil)
		rg := evaluator.NewRowGeneratorStage(cs, newColumn)
		iter, err := rg.Open(ctx)
		require.NoError(t, err)

		row := &evaluator.Row{}
		i := 0
		for iter.Next(row) {
			require.Nil(t, row.Data)
			i++
		}

		require.Equal(t, i, 0)
		require.NoError(t, iter.Close())
		require.Error(t, iter.Err())
	})

	t.Run("should iterate through all rows contained successfully with only one row having"+
		" 0 value", func(t *testing.T) {
		rows := []evaluator.Row{
			{Data: evaluator.Values{{SelectID: 1, Database: "test1", Table: "test",
				Name: "rowCount", Data: evaluator.SQLInt(0)}}},
		}
		cs := evaluator.NewCacheStage(0, rows, nil, nil)
		rg := evaluator.NewRowGeneratorStage(cs, newColumn)
		iter, err := rg.Open(ctx)
		require.NoError(t, err)

		row := &evaluator.Row{}
		i := 0
		for iter.Next(row) {
			require.Nil(t, row.Data)
			i++
		}

		require.Equal(t, i, 0)
		require.NoError(t, iter.Close())
		require.NoError(t, iter.Err())
	})

	t.Run("should iterate through all rows contained successfully with only one row",
		func(t *testing.T) {
			rows := []evaluator.Row{
				{Data: evaluator.Values{{SelectID: 1, Database: "test1", Table: "test",
					Name: "rowCount", Data: evaluator.SQLInt(5)}}},
			}
			cs := evaluator.NewCacheStage(0, rows, nil, nil)
			rg := evaluator.NewRowGeneratorStage(cs, newColumn)
			iter, err := rg.Open(ctx)
			require.NoError(t, err)

			row := &evaluator.Row{}
			i := 0
			for iter.Next(row) {
				require.Nil(t, row.Data)
				i++
			}

			require.Equal(t, i, 5)
			require.NoError(t, iter.Close())
			require.NoError(t, iter.Err())
		})

	t.Run("should iterate through all rows contained successfully with two rows",
		func(t *testing.T) {
			rows := []evaluator.Row{
				{Data: evaluator.Values{{SelectID: 1, Database: "test1", Table: "test",
					Name: "rowCount", Data: evaluator.SQLInt(5)}}},
				{Data: evaluator.Values{{SelectID: 1, Database: "test1", Table: "test",
					Name: "rowCount", Data: evaluator.SQLInt(2)}}},
			}
			cs := evaluator.NewCacheStage(0, rows, nil, nil)
			rg := evaluator.NewRowGeneratorStage(cs, newColumn)
			iter, err := rg.Open(ctx)
			require.NoError(t, err)

			row := &evaluator.Row{}
			i := 0
			for iter.Next(row) {
				require.Nil(t, row.Data)
				i++
			}

			require.Equal(t, i, 7)
			require.NoError(t, iter.Close())
			require.NoError(t, iter.Err())
		})

	t.Run("should iterate through all rows contained successfully with only two rows with"+
		" at least one row with 0 value", func(t *testing.T) {
		rows := []evaluator.Row{
			{Data: evaluator.Values{{SelectID: 1, Database: "test1", Table: "test",
				Name: "rowCount", Data: evaluator.SQLInt(0)}}},
			{Data: evaluator.Values{{SelectID: 1, Database: "test1", Table: "test",
				Name: "rowCount", Data: evaluator.SQLInt(2)}}},
		}
		cs := evaluator.NewCacheStage(0, rows, nil, nil)
		rg := evaluator.NewRowGeneratorStage(cs, newColumn)
		iter, err := rg.Open(ctx)
		require.NoError(t, err)

		row := &evaluator.Row{}
		i := 0
		for iter.Next(row) {
			require.Nil(t, row.Data)
			i++
		}

		require.Equal(t, i, 2)
		require.NoError(t, iter.Close())
		require.NoError(t, iter.Err())
	})

	t.Run("should clear each row's data upon iteration", func(t *testing.T) {
		rows := []evaluator.Row{
			{Data: evaluator.Values{{SelectID: 1, Database: "test1", Table: "test",
				Name: "rowCount", Data: evaluator.SQLInt(3)}}},
		}
		cs := evaluator.NewCacheStage(0, rows, nil, nil)
		rg := evaluator.NewRowGeneratorStage(cs, newColumn)
		iter, err := rg.Open(ctx)
		require.NoError(t, err)

		row := &evaluator.Row{Data: evaluator.Values{
			evaluator.NewValue(0, "test", "foo", "a", evaluator.SQLInt(1))}}
		i := 0
		for iter.Next(row) {
			require.Nil(t, row.Data)
			i++
			row = &evaluator.Row{Data: evaluator.Values{
				evaluator.NewValue(0, "test", "foo", "a", evaluator.SQLInt(1))}}
		}

		require.Equal(t, i, 3)
		require.NoError(t, iter.Close())
		require.NoError(t, iter.Err())
	})
}

func TestRowGeneratorStageMemoryMonitor(t *testing.T) {
	rows := []bson.D{
		{{Name: "a", Value: 6}},
		{{Name: "a", Value: 3}},
	}
	bss := evaluator.NewBSONSourceStage(1, tableOneName, collation.Default, rows)

	newColumn := evaluator.NewColumn(0, "", "", "", "a", "", "a",
		schema.SQLUint64, schema.MongoInt64, false)

	rg := evaluator.NewRowGeneratorStage(bss, newColumn)

	outputMemory := getAllocatedMemorySizeAfterIteration(rg)

	// empty rows cost nothing
	require.Equal(t, uint64(0), outputMemory)
}
