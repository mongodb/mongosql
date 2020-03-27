package schema_test

import (
	"testing"

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

	runTest := func(name string, isCaseSensitive bool, dbToDrop string, expected []string) {
		t.Run(name, func(t *testing.T) {
			req := require.New(t)

			sch, err := schema.NewFromDRDL(log.GlobalLogger(), drdlSchema, isCaseSensitive)
			setupReq.Nil(err)

			err = sch.DropDatabase(dbToDrop)
			req.Nil(err)

			dbs := sch.DatabasesSorted()
			req.Len(dbs, len(expected), "schema should have correct database count")
			for i, db := range dbs {
				req.Equalf(expected[i], db.Name(), "incorrect SQLName at index %d", i)
			}
		})
	}

	runFailTest := func(name string, isCaseSensitive bool, dbToDrop string, expectedError string) {
		t.Run(name, func(t *testing.T) {
			req := require.New(t)

			sch, err := schema.NewFromDRDL(log.GlobalLogger(), drdlSchema, isCaseSensitive)
			setupReq.Nil(err)

			err = sch.DropDatabase(dbToDrop)
			req.NotNil(err)
			req.Equal(expectedError, err.Error())
		})
	}

	var testCases = []struct {
		name            string
		isCaseSensitive bool
		dbToDrop        string
		expectedDBs     []string
		expectedError   string
	}{
		{
			"drop unknown db",
			false,
			"unknown",
			nil,
			"database 'unknown' cannot be dropped as it does not exist",
		},
		{
			"drop db",
			false,
			"baz",
			[]string{"bar", "foo", "zaz"},
			"",
		},
		{
			"drop db case insensitive",
			false,
			"FOO",
			[]string{"bar", "baz", "zaz"},
			"",
		},
		{
			"drop db case sensitive",
			true,
			"FOO",
			nil,
			"database 'FOO' cannot be dropped as it does not exist",
		},
	}

	for _, test := range testCases {
		if test.expectedError == "" {
			runTest(test.name, test.isCaseSensitive, test.dbToDrop, test.expectedDBs)
		} else {
			runFailTest(test.name, test.isCaseSensitive, test.dbToDrop, test.expectedError)
		}
	}
}
