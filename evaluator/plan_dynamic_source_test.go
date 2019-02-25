package evaluator_test

import (
	"context"
	"testing"

	. "github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	. "github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/types"
	. "github.com/10gen/sqlproxy/evaluator/values"

	"github.com/stretchr/testify/require"
)

func TestDynamicSourceStage(t *testing.T) {
	tableName := "foo"
	table := catalog.NewDynamicTable(tableName, catalog.BaseTable, func() Rows {
		return Rows{
			NewNamedRow("", "foo", NewNamedSQLValue("one", NewSQLInt64(valKind, 1)),
				NewNamedSQLValue("two", NewSQLInt64(valKind, 2))),
			NewNamedRow("", "foo", NewNamedSQLValue("one", NewSQLInt64(valKind, 2)),
				NewNamedSQLValue("two", NewSQLInt64(valKind, 3))),
			NewNamedRow("", "foo", NewNamedSQLValue("one", NewSQLInt64(valKind, 3)),
				NewNamedSQLValue("two", NewSQLInt64(valKind, 4))),
		}
	})

	_, err := table.AddColumn(tableName, "one", types.EvalInt64)
	require.NoError(t, err)
	_, err = table.AddColumn(tableName, "two", types.EvalInt64)
	require.NoError(t, err)

	expected := []RowValues{
		{
			{SelectID: 1, Table: tableName, Name: "one", Data: NewSQLInt64(valKind, 1)},
			{SelectID: 2, Table: tableName, Name: "two", Data: NewSQLInt64(valKind, 2)},
		},
		{
			{SelectID: 1, Table: tableName, Name: "one", Data: NewSQLInt64(valKind, 2)},
			{SelectID: 2, Table: tableName, Name: "two", Data: NewSQLInt64(valKind, 3)},
		},
		{
			{SelectID: 1, Table: tableName, Name: "one", Data: NewSQLInt64(valKind, 3)},
			{SelectID: 2, Table: tableName, Name: "two", Data: NewSQLInt64(valKind, 4)},
		},
	}

	db, err := catalog.New("def", nil).AddDatabase("")
	require.NoError(t, err)

	source := NewDynamicSourceStage(db, table, 1, tableName)

	bgCtx := context.Background()
	execCfg := createTestExecutionCfg(MySQLValueKind)
	execState := NewExecutionState()

	iter, err := source.Open(bgCtx, execCfg, execState)
	require.NoError(t, err)

	i := 0

	row := &Row{}
	for iter.Next(bgCtx, row) {
		require.Equal(t, len(row.Data), len(expected[i]))
		require.Equal(t, row.Data, expected[i])
		row = &Row{}
		i++
	}

	require.NoError(t, iter.Close())
	require.NoError(t, iter.Err())
}
