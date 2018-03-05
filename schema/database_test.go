package schema_test

import (
	"testing"

	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
	"github.com/stretchr/testify/require"
)

var lg = log.GlobalLogger()

func TestDatabase(t *testing.T) {
	t.Run("AddTable", testAddTables)
	t.Run("PostProcess", testPostProcessDb)
}

func testPostProcessDb(t *testing.T) {
	req := require.New(t)

	db := schema.NewDatabase(lg, "test", nil)

	full, err := schema.NewTable(lg, "full", "mongo_full", nil, nil)
	req.NoError(err, "failed to create table")

	col := schema.NewColumn("col", schema.SQLInt, "mongo_col", schema.MongoInt)
	full.AddColumn(lg, col, false)

	empty, err := schema.NewTable(lg, "empty", "mongo_empty", nil, nil)
	req.NoError(err, "failed to create table")

	db.AddTable(lg, full)
	db.AddTable(lg, empty)

	tbls := db.TablesSorted()
	req.Len(tbls, 2, "wrong number of tables in db")
	req.Equal("empty", tbls[0].SQLName(), "wrong SQLName for first table")
	req.Equal("full", tbls[1].SQLName(), "wrong SQLName for second table")

	db.PostProcess(lg, false)

	tbls = db.TablesSorted()
	req.Len(tbls, 1, "wrong number of tables in db")
	req.Equal("full", tbls[0].SQLName(), "wrong SQLName for table")
}

func testAddTables(t *testing.T) {

	runTest := func(name string, tblNames, expected []string) {
		t.Run(name, func(t *testing.T) {
			req := require.New(t)

			db := schema.NewDatabase(lg, "test", nil)

			for _, tblName := range tblNames {
				tbl, err := schema.NewTable(lg, tblName, tblName, nil, nil)
				req.NoErrorf(err, "failed to create table %q", tblName)
				db.AddTable(lg, tbl)
			}

			tbls := db.TablesSorted()
			req.Len(tbls, len(expected), "database should have correct table count")
			for i, tbl := range tbls {
				req.Equalf(expected[i], tbl.SQLName(), "incorrect SQLName at index %d", i)
			}
		})
	}

	var tables []string
	var expected []string

	tables = []string{"a"}
	expected = []string{"a"}
	runTest("single_table_lowercase", tables, expected)

	tables = []string{"A"}
	expected = []string{"A"}
	runTest("single_table_uppercase", tables, expected)

	tables = []string{"a", "b"}
	expected = []string{"a", "b"}
	runTest("multiple_tables", tables, expected)

	tables = []string{"a", "a"}
	expected = []string{"a", "a_0"}
	runTest("conflict_exact_match", tables, expected)

	tables = []string{"a", "a", "a"}
	expected = []string{"a", "a_0", "a_1"}
	runTest("conflict_exact_match_twice", tables, expected)

	tables = []string{"a", "A"}
	expected = []string{"a", "A_0"}
	runTest("conflict_case_insensitive_0", tables, expected)

	tables = []string{"A", "a"}
	expected = []string{"A", "a_0"}
	runTest("conflict_case_insensitive_1", tables, expected)

	tables = []string{""}
	expected = []string{""}
	runTest("empty_string", tables, expected)

	tables = []string{" "}
	expected = []string{" "}
	runTest("whitespace", tables, expected)
}
