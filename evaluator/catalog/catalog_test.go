package catalog_test

import (
	"context"
	"testing"

	"github.com/10gen/sqlproxy/evaluator/catalog"

	"github.com/stretchr/testify/require"
)

func TestCatalog(t *testing.T) {
	type addDatabaseTest struct {
		name          string
		expectedError bool
	}

	type findDatabaseTest struct {
		name          string
		expectedError bool
		expectedName  string
	}

	tests := []struct {
		desc              string
		isCaseSensitive   bool
		addDatabaseTests  []addDatabaseTest
		findDatabaseTests []findDatabaseTest
		expectedDBNames   []string
	}{
		{
			"case insensitive",
			false,
			[]addDatabaseTest{
				{"test", false},
				{"test", true},
				{"TEST", true},
				{"tEsT", true},
				{"test2", false},
			},
			[]findDatabaseTest{
				{"blah", true, ""},
				{"test", false, "test"},
				{"TEST", false, "test"},
				{"tEsT", false, "test"},
				{"test2", false, "test2"},
			},
			[]string{"test", "test2"},
		},
		{
			"case sensitive",
			true,
			[]addDatabaseTest{
				{"test", false},
				{"test", true},
				{"TEST", false},
				{"tEsT", false},
				{"test2", false},
			},
			[]findDatabaseTest{
				{"blah", true, ""},
				{"test", false, "test"},
				{"TEST", false, "TEST"},
				{"tEsT", false, "tEsT"},
				{"test2", false, "test2"},
			},
			[]string{"test", "TEST", "tEsT", "test2"},
		},
	}

	testCtx := context.Background()
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			req := require.New(t)

			// Create the test catalog
			c := catalog.New("def", test.isCaseSensitive)

			// Check that it is initialized with the correct name and no databases
			req.Equal("def", string(c.Name))
			dbs, _ := c.Databases(testCtx)
			req.Empty(dbs)

			// Check that it fails to return a nonexistent database
			_, err := c.Database(testCtx, "test")
			req.Error(err)

			// Add databases
			var db catalog.Database
			for _, addDatabaseTest := range test.addDatabaseTests {
				db, err = c.AddDatabase(addDatabaseTest.name)
				if addDatabaseTest.expectedError {
					req.Error(err)
				} else {
					req.NoError(err)
					req.NotNil(db)
					req.Equal(addDatabaseTest.name, string(db.Name()))
				}
			}

			// Find each database
			for _, findDatabaseTest := range test.findDatabaseTests {
				db, err = c.Database(testCtx, findDatabaseTest.name)
				if findDatabaseTest.expectedError {
					req.Error(err)
				} else {
					req.NoError(err)
					req.NotNil(db)
					req.Equal(findDatabaseTest.expectedName, string(db.Name()))
				}
			}

			// Find all databases
			dbs, err = c.Databases(testCtx)
			req.NoError(err)
			req.Equal(len(test.expectedDBNames), len(dbs))

			for i, db := range dbs {
				req.Equal(test.expectedDBNames[i], string(db.Name()))
			}
		})
	}
}

func TestDatabase(t *testing.T) {
	type addTableTest struct {
		name          string
		expectedError bool
	}

	type findTableTest struct {
		name          string
		expectedError bool
		expectedName  string
	}

	tests := []struct {
		desc               string
		isCaseSensitive    bool
		addTableTests      []addTableTest
		findTableTests     []findTableTest
		expectedTableNames []string
	}{
		{
			"case insensitive",
			false,
			[]addTableTest{
				{"foo", false},
				{"foo", true},
				{"FOO", true},
				{"FoO", true},
				{"foo2", false},
			},
			[]findTableTest{
				{"blah", true, ""},
				{"foo", false, "foo"},
				{"FOO", false, "foo"},
				{"FoO", false, "foo"},
				{"foo2", false, "foo2"},
			},
			[]string{"foo", "foo2"},
		},
		{
			"case sensitive",
			true,
			[]addTableTest{
				{"foo", false},
				{"foo", true},
				{"FOO", false},
				{"FoO", false},
				{"foo2", false},
			},
			[]findTableTest{
				{"blah", true, ""},
				{"foo", false, "foo"},
				{"FOO", false, "FOO"},
				{"FoO", false, "FoO"},
				{"foo2", false, "foo2"},
			},
			[]string{"foo", "FOO", "FoO", "foo2"},
		},
	}

	testCtx := context.Background()
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			req := require.New(t)

			// Create the test database
			d, err := catalog.New("def", test.isCaseSensitive).AddDatabase("test")

			// Check that it is initialized with the correct name and no tables
			req.NoError(err)
			req.NotNil(d)
			req.Equal("test", string(d.Name()))
			tbls, _ := d.Tables(testCtx)
			req.Empty(tbls)

			// Check that it fails to return a nonexistent table
			_, err = d.Table(testCtx, "foo")
			req.Error(err)

			// Add tables
			for _, addTableTest := range test.addTableTests {
				tbl := catalog.NewInMemoryTable(addTableTest.name)
				err = d.AddTable(tbl)
				if addTableTest.expectedError {
					req.Error(err)
				} else {
					req.NoError(err)
				}
			}

			// Find each table
			var tbl catalog.Table
			for _, findTableTest := range test.findTableTests {
				tbl, err = d.Table(testCtx, findTableTest.name)
				if findTableTest.expectedError {
					req.Error(err)
				} else {
					req.NoError(err)
					req.NotNil(tbl)
					req.Equal(findTableTest.expectedName, tbl.Name())
				}
			}

			// Find all tables
			tbls, err = d.Tables(testCtx)
			req.NoError(err)
			req.Equal(len(test.expectedTableNames), len(tbls))

			for i, tbl := range tbls {
				req.Equal(test.expectedTableNames[i], tbl.Name())
			}
		})
	}
}
