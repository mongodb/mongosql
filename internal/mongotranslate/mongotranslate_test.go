package mongotranslate

import (
	"testing"

	"github.com/10gen/sqlproxy/evaluator/variable"
)

const (
	testMdbVersion4 = "4.0.0"
	testMdbVersion3 = "3.0.0"
)

func TestTranslateSQLQuery(t *testing.T) {
	tcases := []struct {
		desc             string
		query            string
		defaultDB        string
		mdbVersion       string
		expectError      bool
		expectedPipeline string
	}{
		{
			desc:        "query that can't be parsed/explained (command)",
			query:       "drop table foo.t",
			defaultDB:   testDefaultDB,
			mdbVersion:  testMdbVersion4,
			expectError: true,
		},
		{
			desc:        "query that can't be pushed down (char_length with version < 3.4)",
			query:       "select char_length(foo.a) from foo",
			defaultDB:   testDefaultDB,
			mdbVersion:  testMdbVersion3,
			expectError: true,
		},
		{
			desc:        "query that can't be pushed down (cross join with local db different from foreign db)",
			query:       "select foo.a, bar.b from db1.foo, db2.bar",
			defaultDB:   testDefaultDB,
			mdbVersion:  testMdbVersion4,
			expectError: true,
		},
		{
			desc:             "simple select query (unqualified table) correctly translated to pipeline, using defaultDB",
			query:            "select foo.a from foo",
			defaultDB:        testDefaultDB,
			mdbVersion:       testMdbVersion4,
			expectError:      false,
			expectedPipeline: `[{"$project":{"test_DOT_foo_DOT_a":"$a","_id":0}}]`,
		},
		{
			desc:             "simple select query (qualified table) correctly translated to pipeline",
			query:            "select foo.a from db.foo",
			defaultDB:        testDefaultDB,
			mdbVersion:       testMdbVersion4,
			expectError:      false,
			expectedPipeline: `[{"$project":{"db_DOT_foo_DOT_a":"$a","_id":0}}]`,
		},
		{
			desc:             "simple select query (qualified table) correctly translated to pipeline",
			query:            "select foo.a from foo where foo.b < foo.c",
			defaultDB:        testDefaultDB,
			mdbVersion:       testMdbVersion4,
			expectError:      false,
			expectedPipeline: `[{"$match":{"$expr":{"$and":[{"$lt":["$b","$c"]},{"$gt":["$b",null]},{"$gt":["$c",null]}]}}},{"$project":{"test_DOT_foo_DOT_a":"$a","_id":0}}]`,
		},
	}

	for _, tcase := range tcases {
		actualPipeline, err := TranslateSQLQuery(tcase.query, tcase.defaultDB, tcase.mdbVersion, false)

		if tcase.expectError {
			if err == nil {
				t.Fatalf("%s: expected error, but no error was returned", tcase.desc)
			}
			continue
		}

		if actualPipeline != tcase.expectedPipeline {
			t.Fatalf("%s: actual pipeline is not same as expected pipeline (+++ actual, --- expected)\n+++ %s\n--- %s\n", tcase.desc, actualPipeline, tcase.expectedPipeline)
		}
	}
}

func TestGetCatalog(t *testing.T) {

	// test that mdbVersion is set as expected
	t.Run("correctly set MongoDB version", func(t *testing.T) {
		tcases := []string{"3.2.1", "3.6.0", "4.0.2"}

		emptySchema := newInferredSchema()

		for _, expectedMdbVersion := range tcases {
			ctlg, err := getCatalog(expectedMdbVersion, emptySchema)
			if err != nil {
				t.Fatalf("unexpected error getting catalog: %v", err)
			}

			for _, scope := range []variable.Scope{variable.GlobalScope, variable.SessionScope} {
				actualMdbVersion, err := ctlg.Variables().Get(variable.MongoDBVersion, scope, variable.SystemKind)
				if err != nil {
					t.Fatalf("unexpected error getting MongoDB Version variable for %s scope: %v", scope, err)
				}

				if expectedMdbVersion != actualMdbVersion.Value() {
					t.Fatalf("MongoDB Versions do not match, expected: %s, actual: %s", expectedMdbVersion, actualMdbVersion)
				}
			}
		}
	})

	// test that inferred schema is mapped to catalog as expected
	t.Run("correctly map inferred schema to catalog", func(t *testing.T) {
		tcases := []struct {
			desc      string
			databases []string
			schema    map[string]map[string]map[string]struct{}
		}{
			{
				desc:      "one database with one table with one column",
				databases: []string{"db1"},
				schema: map[string]map[string]map[string]struct{}{
					"db1": {
						"t1": {
							"c1": struct{}{},
						},
					},
				},
			},
			{
				desc:      "one database with one table with many columns",
				databases: []string{"db1"},
				schema: map[string]map[string]map[string]struct{}{
					"db1": {
						"t1": {
							"c1": struct{}{},
							"c2": struct{}{},
							"c3": struct{}{},
							"c4": struct{}{},
						},
					},
				},
			},
			{
				desc:      "one database with many tables with many column (all unique)",
				databases: []string{"db1"},
				schema: map[string]map[string]map[string]struct{}{
					"db1": {
						"t1": {
							"c1": struct{}{},
							"c2": struct{}{},
							"c3": struct{}{},
							"c4": struct{}{},
						},
						"t2": {
							"c5": struct{}{},
							"c6": struct{}{},
							"c7": struct{}{},
							"c8": struct{}{},
						},
					},
				},
			},
			{
				desc:      "one database with many tables with many column (same names)",
				databases: []string{"db1"},
				schema: map[string]map[string]map[string]struct{}{
					"db1": {
						"t1": {
							"c1": struct{}{},
							"c2": struct{}{},
							"c3": struct{}{},
							"c4": struct{}{},
						},
						"t2": {
							"c1": struct{}{},
							"c2": struct{}{},
							"c3": struct{}{},
							"c4": struct{}{},
						},
					},
				},
			},
			{
				desc:      "many databases with many tables (all unique) with many column (all unique)",
				databases: []string{"db1"},
				schema: map[string]map[string]map[string]struct{}{
					"db1": {
						"t1": {
							"c1": struct{}{},
							"c2": struct{}{},
						},
						"t2": {
							"c3": struct{}{},
							"c4": struct{}{},
						},
					},
					"db2": {
						"t3": {
							"c5": struct{}{},
							"c6": struct{}{},
						},
						"t4": {
							"c7": struct{}{},
							"c8": struct{}{},
						},
					},
				},
			},
			{
				desc:      "many databases with many tables (same names) with many column (same names)",
				databases: []string{"db1"},
				schema: map[string]map[string]map[string]struct{}{
					"db1": {
						"t1": {
							"c1": struct{}{},
							"c2": struct{}{},
						},
						"t2": {
							"c3": struct{}{},
							"c4": struct{}{},
						},
					},
					"db2": {
						"t1": {
							"c1": struct{}{},
							"c2": struct{}{},
						},
						"t2": {
							"c3": struct{}{},
							"c4": struct{}{},
						},
					},
				},
			},
		}

		for _, tcase := range tcases {
			is := &inferredSchema{tcase.databases, tcase.schema}

			ctlg, err := getCatalog("4.0.0", is)
			if err != nil {
				t.Fatalf("%s: unexpected error: %v", tcase.desc, err)
			}

			expectedDBs := tcase.databases
			actualDBs := ctlg.Databases()
			actualDBNames := make([]string, len(actualDBs))
			for i, actualDB := range actualDBs {
				actualDBNames[i] = string(actualDB.Name())
			}

			// ensure no unexpected databases were included.
			for _, actualDB := range actualDBNames {
				if !contains(expectedDBs, actualDB) {
					t.Errorf("%s: unexpected database '%s' included in catalog", tcase.desc, actualDB)
				}
			}

			for _, expectedDB := range expectedDBs {
				// ensure all expected databases were included.
				if !contains(actualDBNames, expectedDB) {
					t.Errorf("%s: expected database '%s' not included in catalog", tcase.desc, expectedDB)
				}

				expectedTables, _ := is.Tables(expectedDB)
				actualDB, err := ctlg.Database(expectedDB)
				if err != nil {
					t.Fatalf("%s: unexpected error: %v", tcase.desc, err)
				}
				actualTables := actualDB.Tables()
				actualTableNames := make([]string, len(actualTables))
				for i, actualTable := range actualTables {
					actualTableNames[i] = actualTable.Name()
				}

				// ensure no unexpected tables were included.
				for _, actualTable := range actualTableNames {
					if !contains(expectedTables, actualTable) {
						t.Errorf("%s: unexpected table '%s.%s' included in catalog", tcase.desc, expectedDB, actualTable)
					}
				}

				for _, expectedTable := range expectedTables {
					// ensure all expected tables were included.
					if !contains(actualTableNames, expectedTable) {
						t.Errorf("%s: expected table '%s.%s' not included in catalog", tcase.desc, expectedDB, expectedTable)
					}

					expectedColumns, _ := is.Columns(expectedDB, expectedTable)
					actualTable, err := actualDB.Table(expectedTable)
					if err != nil {
						t.Fatalf("%s: unexpected error: %v", tcase.desc, err)
					}
					actualColumns := actualTable.Columns()
					actualColumnNames := make([]string, len(actualColumns))
					for i, actualColumn := range actualColumns {
						actualColumnNames[i] = actualColumn.Name
					}

					// ensure no unexpected columns were included.
					for _, actualColumn := range actualColumnNames {
						if !contains(expectedColumns, actualColumn) {
							t.Errorf("%s: unexpected column '%s.%s.%s' included in catalog", tcase.desc, expectedDB, expectedTable, actualColumn)
						}
					}

					// ensure all expected columns were included.
					for _, expectedColumn := range expectedColumns {
						if !contains(actualColumnNames, expectedColumn) {
							t.Errorf("%s: expected column '%s.%s.%s' not included in catalog", tcase.desc, expectedDB, expectedTable, expectedColumn)
						}
					}
				}
			}
		}
	})
}
