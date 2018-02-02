package evaluator_test

import (
	"testing"

	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"github.com/stretchr/testify/require"
)

func TestDynamicSourceStage(t *testing.T) {
	tableName := "foo"
	table := catalog.NewDynamicTable(tableName, catalog.BaseTable, func() []*catalog.DataRow {
		return []*catalog.DataRow{
			catalog.NewDataRow(1, 2),
			catalog.NewDataRow(2, 3),
			catalog.NewDataRow(3, 4),
		}
	})

	testSchema := evaluator.MustLoadSchema(testSchema4)
	testInfo := evaluator.GetMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)

	table.AddColumn("one", schema.SQLInt)
	table.AddColumn("two", schema.SQLInt)

	expected := []evaluator.Values{
		{
			{SelectID: 1, Table: tableName, Name: "one", Data: evaluator.SQLInt(1)},
			{SelectID: 1, Table: tableName, Name: "two", Data: evaluator.SQLInt(2)},
		},
		{
			{SelectID: 1, Table: tableName, Name: "one", Data: evaluator.SQLInt(2)},
			{SelectID: 1, Table: tableName, Name: "two", Data: evaluator.SQLInt(3)},
		},
		{
			{SelectID: 1, Table: tableName, Name: "one", Data: evaluator.SQLInt(3)},
			{SelectID: 1, Table: tableName, Name: "two", Data: evaluator.SQLInt(4)},
		},
	}

	db := &catalog.Database{}

	source := evaluator.NewDynamicSourceStage(db, table, 1, tableName)

	execCtx := createTestExecutionCtx(testInfo)

	iter, err := source.Open(execCtx)
	require.NoError(t, err)

	i := 0

	row := &evaluator.Row{}
	for iter.Next(row) {
		require.Equal(t, len(row.Data), len(expected[i]))
		require.Equal(t, row.Data, expected[i])
		row = &evaluator.Row{}
		i++
	}

	require.NoError(t, iter.Close())
	require.NoError(t, iter.Err())
}

func TestDynamicSourceStageMemoryMonitor(t *testing.T) {
	tableName := "foo"
	table := catalog.NewDynamicTable(tableName, catalog.BaseTable, func() []*catalog.DataRow {
		return []*catalog.DataRow{
			catalog.NewDataRow(1, 2),
			catalog.NewDataRow(2, 3),
			catalog.NewDataRow(3, 4),
		}
	})

	table.AddColumn("one", schema.SQLInt)
	table.AddColumn("two", schema.SQLInt)

	db := &catalog.Database{}

	source := evaluator.NewDynamicSourceStage(db, table, 1, tableName)

	actual := getAllocatedMemorySizeAfterIteration(source)
	expected := (valueSize(string(db.Name), tableName, "one", evaluator.SQLInt(0)) +
		valueSize(string(db.Name), tableName, "two", evaluator.SQLInt(0))) * 3

	require.Equal(t, expected, actual)
}
