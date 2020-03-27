package schema_test

import (
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/internal/option"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/drdl"
	"github.com/stretchr/testify/require"
)

var lg = log.GlobalLogger()

func TestDatabase(t *testing.T) {
	t.Run("AddTable", testAddTables)
	t.Run("DropTable", testDropTables)
	t.Run("PostProcess", testPostProcessDb)
}

func testPostProcessDb(t *testing.T) {
	req := require.New(t)

	db := schema.NewDatabase(lg, "test", nil, false)

	full, err := schema.NewTable(lg, "full", "mongo_full", nil, nil, []schema.Index{}, option.NoneString(), false)
	req.NoError(err, "failed to create table")

	col := schema.NewColumn("col", schema.SQLInt, "mongo_col", schema.MongoInt, false, option.NoneString())
	full.AddColumn(lg, col, false)

	empty, err := schema.NewTable(lg, "empty", "mongo_empty", nil, nil, []schema.Index{}, option.NoneString(), false)
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

	runTest := func(name string, isCaseSensitive bool, tblNames, expected []string) {
		t.Run(name, func(t *testing.T) {
			req := require.New(t)

			db := schema.NewDatabase(lg, "test", nil, isCaseSensitive)

			for _, tblName := range tblNames {
				tbl, err := schema.NewTable(lg, tblName, tblName, nil, nil, []schema.Index{}, option.NoneString(), false)
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

	var testCases = []struct {
		name            string
		isCaseSensitive bool
		tables          []string
		expected        []string
	}{
		{
			"single table lowercase",
			false,
			[]string{"a"},
			[]string{"a"},
		},
		{
			"single table uppercase",
			false,
			[]string{"A"},
			[]string{"A"},
		},
		{
			"multiple tables",
			false,
			[]string{"a", "b"},
			[]string{"a", "b"},
		},
		{
			"case insensitive database with case sensitive conflict",
			false,
			[]string{"a", "a"},
			[]string{"a", "a_0"},
		},
		{
			"case insensitive database with multiple case sensitive conflicts",
			false,
			[]string{"a", "a", "a"},
			[]string{"a", "a_0", "a_1"},
		},
		{
			"case insensitive database with case insensitive conflict",
			false,
			[]string{"a", "A"},
			[]string{"a", "A_0"},
		},
		{
			"case insensitive database with case insensitive conflict flipped",
			false,
			[]string{"A", "a"},
			[]string{"A", "a_0"},
		},
		{
			"case sensitive database with case insensitive conflicts does NOT rename",
			true,
			[]string{"FOO", "fOo", "foo"},
			[]string{"FOO", "fOo", "foo"},
		},
		{
			"case sensitive database with case sensitive conflict results in rename",
			true,
			[]string{"FOO", "FOO", "FOO"},
			[]string{"FOO", "FOO_0", "FOO_1"},
		},
		{
			"empty string",
			false,
			[]string{""},
			[]string{""},
		},
		{
			"whitespace",
			false,
			[]string{" "},
			[]string{" "},
		},
	}

	for _, test := range testCases {
		runTest(test.name, test.isCaseSensitive, test.tables, test.expected)
	}
}

func testDropTables(t *testing.T) {
	setupReq := require.New(t)
	var testDRDL = []byte(
		`
schema:
-
  db: test1
  tables:
  -
      table: foo
      collection: foo
      columns:
      -
        Name: a
        MongoType: int
        SqlName: a
        SqlType: int

  -
      table: bar
      collection: bar
      columns:
      -
        Name: a
        MongoType: int
        SqlName: a
        SqlType: int

  -
      table: zaz
      collection: zaz
      columns:
      -
        Name: a
        MongoType: int
        SqlName: a
        SqlType: int

  -
      table: baz
      collection: baz
      columns:
      -
        Name: a
        MongoType: int
        SqlName: a
        SqlType: int
`)

	drdlSchema, err := drdl.NewFromBytes(testDRDL)
	setupReq.Nil(err)

	runTest := func(name string, isCaseSensitive bool, table string, expected []string) {
		t.Run(name, func(t *testing.T) {
			req := require.New(t)

			sch, err := schema.NewFromDRDL(log.GlobalLogger(), drdlSchema, isCaseSensitive)
			setupReq.Nil(err)

			db := sch.Database("test1")
			setupReq.NotNil(db)

			err = db.DropTable(table)
			req.Nil(err)

			tbls := db.TablesSorted()
			req.Len(tbls, len(expected), "database should have correct table count")
			for i, tbl := range tbls {
				req.Equalf(expected[i], tbl.SQLName(), "incorrect SQLName at index %d", i)
			}
		})
	}

	runFailTest := func(name string, isCaseSensitive bool, table string, expectedError string) {
		t.Run(name, func(t *testing.T) {
			req := require.New(t)

			sch, err := schema.NewFromDRDL(log.GlobalLogger(), drdlSchema, isCaseSensitive)
			setupReq.Nil(err)

			db := sch.Database("test1")
			setupReq.NotNil(db)

			err = db.DropTable(table)
			req.NotNil(err)
			req.Equal(expectedError, err.Error())
		})
	}

	var testCases = []struct {
		name            string
		isCaseSensitive bool
		tableToDrop     string
		expectedTables  []string
		expectedError   string
	}{
		{
			"drop unknown table",
			false,
			"unknown",
			nil,
			"table 'test1.unknown' cannot be dropped, no such table in database 'test1'",
		},
		{
			"drop table",
			false,
			"baz",
			[]string{"bar", "foo", "zaz"},
			"",
		},
		{
			"drop table case insensitive",
			false,
			"FOO",
			[]string{"bar", "baz", "zaz"},
			"",
		},
		{
			"drop table case sensitive",
			true,
			"FOO",
			nil,
			"table 'test1.FOO' cannot be dropped, no such table in database 'test1'",
		},
	}

	for _, test := range testCases {
		if test.expectedError == "" {
			runTest(test.name, test.isCaseSensitive, test.tableToDrop, test.expectedTables)
		} else {
			runFailTest(test.name, test.isCaseSensitive, test.tableToDrop, test.expectedError)
		}
	}
}

func TestNewDatabaseFromDRDLWithInvalidMongoType(t *testing.T) {
	var testSchemaWithInvalidMongoType = []byte(
		`
schema:
-
  db: test1
  tables:
  -
      table: foo
      collection: foo
      columns:
      -
        Name: a
        MongoType: invalidtype
        SqlName: a
        SqlType: int
`)

	drdlSchema, err := drdl.NewFromBytes(testSchemaWithInvalidMongoType)
	if err != nil {
		t.Fatal(err)
	}

	// Expect an error on mapping a schema with an invalid mongo type.
	_, err = schema.NewFromDRDL(log.GlobalLogger(), drdlSchema, false)
	require.Equal(t, err, fmt.Errorf(`unable to create database "test1" from drdl: `+
		`unable to create table "foo" from drdl: `+
		`unable to create column "a" from drdl: `+
		`unsupported Mongo type: "invalidtype" on column "a"`))
}

func TestNewDatabaseFromDRDLWithInvalidSQLType(t *testing.T) {
	var testSchemaWithInvalidSQLType = []byte(
		`
schema:
-
  db: test2
  tables:
  -
      table: foo
      collection: foo
      columns:
      -
        Name: b
        MongoType: int
        SqlName: b
        SqlType: notrealtype
`)

	drdlSchema, err := drdl.NewFromBytes(testSchemaWithInvalidSQLType)
	if err != nil {
		t.Fatal(err)
	}

	// Expect an error on mapping a schema with an invalid sql type.
	_, err = schema.NewFromDRDL(log.GlobalLogger(), drdlSchema, false)
	require.Equal(t, err, fmt.Errorf(`unable to create database "test2" from drdl: `+
		`unable to create table "foo" from drdl: `+
		`unable to create column "b" from drdl: `+
		`unsupported SQL type: "notrealtype" on column "b"`))
}
