package schema_test

import (
	"testing"

	"github.com/10gen/sqlproxy/internal/option"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/drdl"
	"github.com/stretchr/testify/require"
)

func TestSchema(t *testing.T) {
	t.Run("DropDatabase", testDropDatabases)
}

func testDropDatabases(t *testing.T) {
	setupReq := require.New(t)
	var testDRDL = []byte(
		`
schema:
-
  db: foo
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
  db: bar
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
  db: baz
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
  db: zaz
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
`)

	drdlSchema, err := drdl.NewFromBytes(testDRDL)
	setupReq.Nil(err)

	sch, err := schema.NewFromDRDL(log.GlobalLogger(), drdlSchema)
	setupReq.Nil(err)

	runTest := func(name string, table option.String, expected []string) {
		t.Run(name, func(t *testing.T) {
			req := require.New(t)

			if table.IsSome() {
				err = sch.DropDatabase(table.Unwrap())
				req.Nil(err)
			}

			dbs := sch.DatabasesSorted()
			req.Len(dbs, len(expected), "database should have correct table count")
			for i, db := range dbs {
				req.Equalf(expected[i], db.Name(), "incorrect SQLName at index %d", i)
			}
		})
	}

	runFailTest := func(name string, table string, expectedError string) {
		t.Run(name, func(t *testing.T) {
			req := require.New(t)

			err = sch.DropDatabase(table)
			req.NotNil(err)
			req.Equal(expectedError, err.Error())
		})
	}

	var expected []string
	cannotDropUnknown := "database 'unknown' cannot be dropped as it does not exist"

	expected = []string{"bar", "baz", "foo", "zaz"}
	runTest("drop nothing", option.NoneString(), expected)

	runFailTest("drop unknown", "unknown", cannotDropUnknown)

	expected = []string{"bar", "foo", "zaz"}
	runTest("drop baz", option.SomeString("baz"), expected)

	expected = []string{"foo", "zaz"}
	runTest("drop bar", option.SomeString("bar"), expected)

	runFailTest("drop unknown again", "unknown", cannotDropUnknown)

	expected = []string{"foo"}
	runTest("drop zaz", option.SomeString("zaz"), expected)

	expected = []string{}
	runTest("drop FOO case insensitive", option.SomeString("FOO"), expected)

	runFailTest("drop unknown one last time", "unknown", cannotDropUnknown)
}
