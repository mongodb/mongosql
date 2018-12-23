package evaluator_test

import (
	"context"
	"testing"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/internal/schema"

	"github.com/stretchr/testify/require"
)

func TestDynamicSourceStage(t *testing.T) {
	tableName := "foo"
	table := catalog.NewDynamicTable(catalog.TableName(tableName), catalog.BaseTable, func() []*catalog.DataRow {
		return []*catalog.DataRow{
			catalog.NewDataRow(1, 2),
			catalog.NewDataRow(2, 3),
			catalog.NewDataRow(3, 4),
		}
	})

	_, err := table.AddColumn("one", catalog.SQLType(schema.SQLInt))
	require.NoError(t, err)
	_, err = table.AddColumn("two", catalog.SQLType(schema.SQLInt))
	require.NoError(t, err)

	expected := []evaluator.Values{
		{
			{SelectID: 1, Table: tableName, Name: "one", Data: evaluator.NewSQLInt64(valKind, 1)},
			{SelectID: 1, Table: tableName, Name: "two", Data: evaluator.NewSQLInt64(valKind, 2)},
		},
		{
			{SelectID: 1, Table: tableName, Name: "one", Data: evaluator.NewSQLInt64(valKind, 2)},
			{SelectID: 1, Table: tableName, Name: "two", Data: evaluator.NewSQLInt64(valKind, 3)},
		},
		{
			{SelectID: 1, Table: tableName, Name: "one", Data: evaluator.NewSQLInt64(valKind, 3)},
			{SelectID: 1, Table: tableName, Name: "two", Data: evaluator.NewSQLInt64(valKind, 4)},
		},
	}

	db, err := catalog.New("def", nil).AddDatabase("")
	require.NoError(t, err)

	source := evaluator.NewDynamicSourceStage(db, table, 1, tableName)

	bgCtx := context.Background()
	execCfg := createTestExecutionCfg()
	execState := evaluator.NewExecutionState()

	iter, err := source.Open(bgCtx, execCfg, execState)
	require.NoError(t, err)

	i := 0

	row := &evaluator.Row{}
	for iter.Next(bgCtx, row) {
		require.Equal(t, len(row.Data), len(expected[i]))
		require.Equal(t, row.Data, expected[i])
		row = &evaluator.Row{}
		i++
	}

	require.NoError(t, iter.Close())
	require.NoError(t, iter.Err())
}
