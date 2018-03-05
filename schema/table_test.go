package schema_test

import (
	"testing"

	"github.com/10gen/sqlproxy/schema"
	"github.com/stretchr/testify/require"
)

func TestTable(t *testing.T) {
	t.Run("AddColumn", testAddColumn)
	t.Run("AddGeoColumn", testAddGeoColumn)
	t.Run("PostProcess", testPostProcessTable)
}

func testAddColumn(t *testing.T) {
	runTest := func(name string, colNames, expected []string) {
		t.Run(name, func(t *testing.T) {
			req := require.New(t)

			table, err := schema.NewTable(lg, "t", "mongo_t1", nil, nil)
			req.NoError(err, "failed to create table")

			for _, colName := range colNames {
				col := schema.NewColumn(colName, "", colName, "")
				table.AddColumn(lg, col, false)
			}

			cols := table.ColumnsSorted()
			req.Len(cols, len(expected), "table should have correct column count")
			for i, col := range cols {
				req.Equalf(expected[i], col.SQLName(), "incorrect SQLName at index %d", i)
			}
		})
	}

	var columns []string
	var expected []string

	columns = []string{"a"}
	expected = []string{"a"}
	runTest("single_column_lowercase", columns, expected)

	columns = []string{"A"}
	expected = []string{"A"}
	runTest("single_column_uppercase", columns, expected)

	columns = []string{"a", "b"}
	expected = []string{"a", "b"}
	runTest("multiple_columns", columns, expected)

	columns = []string{"a", "a"}
	expected = []string{"a", "a_0"}
	runTest("conflict_exact_match", columns, expected)

	columns = []string{"a", "a", "a"}
	expected = []string{"a", "a_0", "a_1"}
	runTest("conflict_exact_match_twice", columns, expected)

	columns = []string{"a", "A"}
	expected = []string{"a", "A_0"}
	runTest("conflict_case_insensitive_0", columns, expected)

	columns = []string{"A", "a"}
	expected = []string{"A", "a_0"}
	runTest("conflict_case_insensitive_1", columns, expected)

	columns = []string{"c", "C", "ca", "cA", "Ca"}
	expected = []string{"c", "C_0", "ca", "cA_0", "Ca_1"}
	runTest("many_conflicts", columns, expected)

	columns = []string{"_z", "_a", "_id", "a", "C", "b", "D"}
	expected = []string{"_id", "_a", "_z", "a", "b", "C", "D"}
	runTest("sort_order", columns, expected)

	// columns with empty string names should not get added
	columns = []string{""}
	expected = []string{}
	runTest("empty", columns, expected)

	// columns composed of spaces should not get added
	columns = []string{"  "}
	expected = []string{}
	runTest("spaces", columns, expected)
}

func testAddGeoColumn(t *testing.T) {
	runTest := func(name string, colNames, geoColumns, expected []string) {
		t.Run(name, func(t *testing.T) {
			req := require.New(t)

			table, err := schema.NewTable(lg, "t", "mongo_t1", nil, nil)
			req.NoError(err, "failed to create table")

			for _, colName := range colNames {
				col := schema.NewColumn(colName, "", colName, "")
				table.AddColumn(lg, col, false)
			}

			for _, colName := range geoColumns {
				col := schema.NewColumn(colName, "", colName, schema.MongoGeo2D)
				table.AddColumn(lg, col, false)
			}

			cols := table.ColumnsSorted()
			req.Len(cols, len(expected), "table should have correct column count")
			for i, col := range cols {
				req.Equalf(expected[i], col.SQLName(), "incorrect SQLName at index %d", i)
			}
		})
	}

	var columns []string
	var geoColumn []string
	var expected []string

	columns = []string{}
	geoColumn = []string{"a"}
	expected = []string{"a_latitude", "a_longitude"}
	runTest("single", columns, geoColumn, expected)

	columns = []string{}
	geoColumn = []string{"a", "b"}
	expected = []string{"a_latitude", "a_longitude", "b_latitude", "b_longitude"}
	runTest("multiple", columns, geoColumn, expected)

	columns = []string{"a"}
	geoColumn = []string{"a"}
	expected = []string{"a", "a_latitude", "a_longitude"}
	runTest("no_conflict_base_name", columns, geoColumn, expected)

	columns = []string{"a_latitude"}
	geoColumn = []string{"a"}
	expected = []string{"a_latitude", "a_latitude_0", "a_longitude"}
	runTest("conflict", columns, geoColumn, expected)

	columns = []string{"a"}
	geoColumn = []string{""}
	expected = []string{"a"}
	runTest("empty", columns, geoColumn, expected)
}

func testPostProcessTable(t *testing.T) {

	t.Run("no_parent", func(t *testing.T) {
		req := require.New(t)

		table, err := schema.NewTable(lg, "tbl", "mongo_tbl", nil, nil)
		req.NoError(err, "failed to create table")

		newTable := table.DeepCopy()
		newTable.PostProcess(lg, false)
		req.NoError(table.Equals(newTable), "table should not change")

		newTable.PostProcess(lg, true)
		req.NoError(table.Equals(newTable), "table should not change")
	})

	t.Run("no_pre_join", func(t *testing.T) {
		req := require.New(t)

		// create some columns
		pk := schema.NewColumn("pk", schema.SQLBoolean, "pk", schema.MongoBool)
		col := schema.NewColumn("col", schema.SQLBoolean, "col", schema.MongoBool)

		// create an empty table
		table, err := schema.NewTable(lg, "tbl", "mongo_tbl", nil, nil)
		req.NoError(err, "failed to create table")

		// create a parent table and add the columns to it
		parent := table.DeepCopy()
		parent.AddColumn(lg, pk.DeepCopy(), true)
		parent.AddColumn(lg, col.DeepCopy(), false)

		// set the parent table
		err = table.SetParent(parent)
		req.NoError(err, "could not set parent")

		err = table.SetParent(parent)
		req.Error(err, "should get an error if setting parent twice")

		// sanity check
		req.Len(parent.Columns(), 2, "parent should have two columns")
		req.Len(table.Columns(), 0, "child should be empty")

		// post-process the table
		table.PostProcess(lg, false)

		// check that parent pk column was copied
		req.Len(parent.Columns(), 2, "parent should have two columns")
		req.Len(table.Columns(), 1, "child should have one column")
		req.NoError(table.Columns()[0].Equals(pk), "child should have pk column")
	})

	t.Run("pre_join", func(t *testing.T) {
		req := require.New(t)

		// create some columns
		pk := schema.NewColumn("pk", schema.SQLBoolean, "pk", schema.MongoBool)
		col := schema.NewColumn("col", schema.SQLBoolean, "col", schema.MongoBool)

		// create an empty table
		table, err := schema.NewTable(lg, "tbl", "mongo_tbl", nil, nil)
		req.NoError(err, "failed to create table")

		// create a parent table and add the columns to it
		parent := table.DeepCopy()
		parent.AddColumn(lg, pk.DeepCopy(), true)
		parent.AddColumn(lg, col.DeepCopy(), false)

		// set the parent table
		err = table.SetParent(parent)
		req.NoError(err, "could not set parent")

		// sanity check
		req.Len(parent.Columns(), 2, "parent should have two columns")
		req.Len(table.Columns(), 0, "child should be empty")

		// post-process the table
		table.PostProcess(lg, true)

		// check that parent columns were copied
		req.Len(parent.Columns(), 2, "parent should have two columns")
		cols := table.ColumnsSorted()
		req.Len(cols, 2, "child should have two columns")
		req.NoError(cols[0].Equals(col), "child should have non-pk column")
		req.NoError(cols[1].Equals(pk), "child should have pk column")
	})

	t.Run("pre_join_conflict", func(t *testing.T) {
		req := require.New(t)

		// create some columns
		pk := schema.NewColumn("pk", schema.SQLBoolean, "pk", schema.MongoBool)
		col := schema.NewColumn("col", schema.SQLBoolean, "col", schema.MongoBool)
		childCol := schema.NewColumn("col", schema.SQLBoolean, "childCol", schema.MongoBool)

		// create an empty table
		table, err := schema.NewTable(lg, "tbl", "mongo_tbl", nil, nil)
		req.NoError(err, "failed to create table")

		// create a parent table and add the columns to it
		parent := table.DeepCopy()
		parent.AddColumn(lg, pk.DeepCopy(), true)
		parent.AddColumn(lg, col.DeepCopy(), false)

		// set the parent table
		err = table.SetParent(parent)
		req.NoError(err, "could not set parent")

		// add a column to the child table
		table.AddColumn(lg, childCol.DeepCopy(), false)

		// sanity check
		req.Len(parent.Columns(), 2, "parent should have two columns")
		req.Len(table.Columns(), 1, "child should have one column")

		// post-process the table
		table.PostProcess(lg, true)

		// check that parent columns were copied
		req.Len(parent.Columns(), 2, "parent should have two columns")
		cols := table.ColumnsSorted()
		req.Len(cols, 3, "child should have three columns")
		req.NoError(cols[0].Equals(childCol), "child should have non-pk column")
		req.Equal(col.SQLName()+"_0", cols[1].SQLName(), "conflicting parent column should be renamed")
		req.Equal(col.MongoName(), cols[1].MongoName(), "conflicting parent column should keep same MongoName")
		req.NoError(cols[2].Equals(pk), "child should have pk column")
	})
}
