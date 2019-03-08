package mongotranslate

import (
	"testing"

	"github.com/10gen/mongo-go-driver/bson"

	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
)

const (
	testMdbVersion4 = "4.0.0"
	testMdbVersion3 = "3.0.0"
	testDefaultDB   = "test"
	testSchema      = "../../testdata/resources/schema/mongotranslate.drdl"
)

func TestTranslateSQLQuery(t *testing.T) {
	tcases := []struct {
		desc             string
		query            string
		defaultDB        string
		mdbVersion       string
		schema           string
		expectedError    string
		expectedPipeline string
	}{
		{
			desc:       "query that can't be parsed/explained (command)",
			query:      "drop table foo.t",
			defaultDB:  testDefaultDB,
			mdbVersion: testMdbVersion4,
			schema:     testSchema,
			expectedError: `fatal error executing sql "explain drop table foo.t": ERROR 1064 (42000): ` +
				`parse sql 'explain drop table foo.t' error: syntax error at position 14 near drop`,
		},
		{
			desc:       "query that can't be pushed down (char_length with version < 3.4)",
			query:      "select char_length(foo.a) from foo",
			defaultDB:  testDefaultDB,
			mdbVersion: testMdbVersion3,
			schema:     testSchema,
			expectedError: "query not fully pushed down: failed to push down SQLScalarFunctionExpr(characterLength): " +
				"cannot push down to MongoDB < 3.4.",
		},
		{
			desc:          "query that can't be pushed down (cross join with local db different from foreign db)",
			query:         "select foo.a, bar.b from test.foo, test2.bar",
			defaultDB:     testDefaultDB,
			mdbVersion:    testMdbVersion4,
			schema:        testSchema,
			expectedError: "query not fully pushed down: failed to push down ProjectStage: unable to push down source stage.",
		},
		{
			desc:             "simple select query (unqualified table) correctly translated to pipeline, using defaultDB",
			query:            "select foo.a from foo",
			defaultDB:        testDefaultDB,
			mdbVersion:       testMdbVersion4,
			schema:           testSchema,
			expectedPipeline: `[{"$project":{"test_DOT_foo_DOT_a":"$a","_id":0}}]`,
		},
		{
			desc:             "simple select query (qualified table) correctly translated to pipeline",
			query:            "select foo.a from test.foo",
			defaultDB:        testDefaultDB,
			mdbVersion:       testMdbVersion4,
			schema:           testSchema,
			expectedPipeline: `[{"$project":{"test_DOT_foo_DOT_a":"$a","_id":0}}]`,
		},
		{
			desc:             "simple select query (qualified table) correctly translated to pipeline",
			query:            "select foo.a from foo where foo.b < foo.c",
			defaultDB:        testDefaultDB,
			mdbVersion:       testMdbVersion4,
			schema:           testSchema,
			expectedPipeline: `[{"$match":{"$expr":{"$and":[{"$lt":["$b","$c"]},{"$gt":["$b",null]},{"$gt":["$c",null]}]}}},{"$project":{"test_DOT_foo_DOT_a":"$a","_id":0}}]`,
		},
		{
			desc:          "invalid schema parameter - non-drdl file",
			query:         "select foo.a from foo",
			defaultDB:     testDefaultDB,
			mdbVersion:    testMdbVersion4,
			schema:        "../../parser/Makefile",
			expectedError: "fatal error getting schema from drdl: Invalid timestamp: 'MAKEFLAGS = -s\nsql.go' at line 4, column 0",
		},
		{
			desc:          "invalid schema parameter - no drdl files found in directory",
			query:         "select foo.a from foo",
			defaultDB:     testDefaultDB,
			mdbVersion:    testMdbVersion4,
			schema:        "../../testdata/resources",
			expectedError: `fatal error executing sql "explain select foo.a from foo": ERROR 1049 (42000): Unknown database 'test'`,
		},
		{
			desc:             "valid directory passed in as schema parameter",
			query:            "select foo.a from foo",
			defaultDB:        testDefaultDB,
			mdbVersion:       testMdbVersion4,
			schema:           "../../testdata/resources/schema",
			expectedPipeline: `[{"$project":{"test_DOT_foo_DOT_a":"$a","_id":0}}]`,
		},
		{
			desc:          "database doesn't exist in schema",
			query:         "select foo.a from foo",
			defaultDB:     "invalidDatabase",
			mdbVersion:    testMdbVersion4,
			schema:        testSchema,
			expectedError: `fatal error executing sql "explain select foo.a from foo": ERROR 1049 (42000): Unknown database 'invaliddatabase'`,
		},
		{
			desc:          "table doesn't exist in schema",
			query:         "select a from invalidTable",
			defaultDB:     testDefaultDB,
			mdbVersion:    testMdbVersion4,
			schema:        testSchema,
			expectedError: `fatal error executing sql "explain select a from invalidTable": ERROR 1146 (42S02): Table 'test.invalidtable' doesn't exist`,
		},
		{
			desc:          "column doesn't exist in schema",
			query:         "select invalidColumn from foo",
			defaultDB:     testDefaultDB,
			mdbVersion:    testMdbVersion4,
			schema:        testSchema,
			expectedError: `fatal error executing sql "explain select invalidColumn from foo": ERROR 1054 (42S22): Unknown column 'invalidColumn' in 'field list'`,
		},
	}

	for _, tcase := range tcases {
		actualPipeline, err := TranslateSQLQuery(tcase.query, tcase.defaultDB, tcase.mdbVersion, tcase.schema)

		if tcase.expectedError != "" {
			if err == nil {
				t.Errorf("%s: expected error, but no error was returned", tcase.desc)
			} else if tcase.expectedError != err.Error() {
				t.Errorf(`Error received does not match expected error. Expected "%v", got "%v".`, tcase.expectedError, err)
			}
			continue
		}

		if actualPipeline != tcase.expectedPipeline {
			t.Fatalf("%s: actual pipeline is not same as expected pipeline (+++ actual, --- expected)\n+++ %s\n--- %s\n", tcase.desc, actualPipeline, tcase.expectedPipeline)
		}
	}
}

type ctlgTest struct {
	desc             string
	relationalSchema *schema.Schema
}

func TestGetCatalog(t *testing.T) {

	// test that mdbVersion is set as expected
	t.Run("correctly set MongoDB version", func(t *testing.T) {
		tcases := []string{"3.2.1", "3.6.0", "4.0.2"}

		schema, err := loadSchema(testSchema)
		if err != nil {
			t.Fatalf("unexpected error loading schema: %v", err)
		}

		for _, expectedMdbVersion := range tcases {
			ctlg, err := getCatalog(expectedMdbVersion, schema)
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

	lg := log.GlobalLogger()

	c1 := schema.NewColumn("c1", schema.SQLInt, "c1", schema.MongoInt)
	c2 := schema.NewColumn("c2", schema.SQLInt, "c2", schema.MongoInt)
	c3 := schema.NewColumn("c3", schema.SQLInt, "c3", schema.MongoInt)
	c4 := schema.NewColumn("c4", schema.SQLInt, "c4", schema.MongoInt)
	c5 := schema.NewColumn("c5", schema.SQLInt, "c5", schema.MongoInt)
	c6 := schema.NewColumn("c6", schema.SQLInt, "c6", schema.MongoInt)
	c7 := schema.NewColumn("c7", schema.SQLInt, "c7", schema.MongoInt)
	c8 := schema.NewColumn("c8", schema.SQLInt, "c8", schema.MongoInt)

	tOneColumn, _ := schema.NewTable(lg, "t1", "t1", []bson.D{}, []*schema.Column{c1})
	tManyColumn, _ := schema.NewTable(lg, "t2", "t2", []bson.D{}, []*schema.Column{c1, c2})
	tManyColumn2, _ := schema.NewTable(lg, "t3", "t3", []bson.D{}, []*schema.Column{c3, c4})
	tManyColumn3, _ := schema.NewTable(lg, "t4", "t4", []bson.D{}, []*schema.Column{c5, c6})
	tManyColumn4, _ := schema.NewTable(lg, "t5", "t5", []bson.D{}, []*schema.Column{c7, c8})
	tManyColumn5, _ := schema.NewTable(lg, "t6", "t6", []bson.D{}, []*schema.Column{c1, c2})
	tManyColumn6, _ := schema.NewTable(lg, "t7", "t7", []bson.D{}, []*schema.Column{c3, c4})

	dbOneTableOneColumn := schema.NewDatabase(lg, "db1", []*schema.Table{tOneColumn})
	dbOneTableManyColumn := schema.NewDatabase(lg, "db2", []*schema.Table{tManyColumn})
	dbManyTableManyColumn := schema.NewDatabase(lg, "db3", []*schema.Table{tManyColumn, tManyColumn2})
	dbManyTableManyColumn2 := schema.NewDatabase(lg, "db4", []*schema.Table{tManyColumn3, tManyColumn4})
	dbManyTableManySameColumn := schema.NewDatabase(lg, "db5", []*schema.Table{tManyColumn, tManyColumn5})
	dbManyTableManySameColumn2 := schema.NewDatabase(lg, "db6", []*schema.Table{tManyColumn2, tManyColumn6})

	schema1, _ := schema.New([]*schema.Database{dbOneTableOneColumn}, nil)
	schema2, _ := schema.New([]*schema.Database{dbOneTableManyColumn}, nil)
	schema3, _ := schema.New([]*schema.Database{dbManyTableManyColumn}, nil)
	schema4, _ := schema.New([]*schema.Database{dbManyTableManySameColumn}, nil)
	schema5, _ := schema.New([]*schema.Database{dbManyTableManyColumn, dbManyTableManyColumn2}, nil)
	schema6, _ := schema.New([]*schema.Database{dbManyTableManySameColumn, dbManyTableManySameColumn2}, nil)

	// test that inferred schema is mapped to catalog as expected
	t.Run("correctly map inferred schema to catalog", func(t *testing.T) {

		var tcases = []ctlgTest{
			{
				desc:             "one database with one table with one column",
				relationalSchema: schema1,
			},
			{
				desc:             "one database with one table with many columns",
				relationalSchema: schema2,
			},
			{
				desc:             "one database with many tables with many column (all unique)",
				relationalSchema: schema3,
			},
			{
				desc:             "one database with many tables with many column (same names)",
				relationalSchema: schema4,
			},
			{
				desc:             "many databases with many tables (all unique) with many column (all unique)",
				relationalSchema: schema5,
			},
			{
				desc:             "many databases with many tables (same names) with many column (same names)",
				relationalSchema: schema6,
			},
		}

		for _, tcase := range tcases {
			ctlg, err := getCatalog("4.0.0", tcase.relationalSchema)
			if err != nil {
				t.Fatalf("%s: unexpected error: %v", tcase.desc, err)
			}

			validateCatalog(t, tcase, ctlg)
		}

	})
}

// validateCatalog ensures that the database, table, and collection names in the relational schema
// match those in the catalog.
func validateCatalog(t *testing.T, tcase ctlgTest, ctlg catalog.Catalog) {
	expectedDBs := tcase.relationalSchema.Databases()
	actualDBs := ctlg.Databases()
	if len(expectedDBs) != len(actualDBs) {
		t.Fatalf("Catalog databases don't match expected value: want %v, got %v.", expectedDBs, actualDBs)
	}

	for _, expectedDB := range expectedDBs {
		// ensure all expected databases were found in the catalog.
		if !containsDB(actualDBs, expectedDB) {
			t.Errorf("%s: expected database '%s' not included in catalog", tcase.desc, expectedDB.Name())
		}
		currentDB, err := ctlg.Database(expectedDB.Name())
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tcase.desc, err)
		}
		validateDatabase(t, tcase, currentDB)
	}
}

func validateDatabase(t *testing.T, tcase ctlgTest, db catalog.Database) {
	expectedTables := tcase.relationalSchema.Database(string(db.Name())).Tables()
	actualTables := db.Tables()
	if len(expectedTables) != len(actualTables) {
		t.Fatalf("Catalog tables don't match expected value: want %v, got %v.", expectedTables, actualTables)
	}

	for _, expectedTable := range expectedTables {
		// ensure all expected tables were found in the catalog.
		if !containsTable(actualTables, expectedTable) {
			t.Errorf("%s: expected table '%s.%s' not included in catalog", tcase.desc, db.Name(), expectedTable.SQLName())
		}
		currentTable, err := db.Table(expectedTable.SQLName())
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tcase.desc, err)
		}
		validateTable(t, tcase, db, currentTable)
	}
}

func validateTable(t *testing.T, tcase ctlgTest, db catalog.Database, tbl catalog.Table) {
	expectedColumns := tcase.relationalSchema.Database(string(db.Name())).Table(tbl.Name()).Columns()
	actualColumns := tbl.Columns()
	if len(expectedColumns) != len(actualColumns) {
		t.Fatalf("Catalog columns don't match expected value: want %v, got %v.", expectedColumns, actualColumns)
	}

	for _, expectedColumn := range expectedColumns {
		// ensure all expected columns were found in the catalog.
		if !containsColumn(actualColumns, expectedColumn) {
			t.Errorf("%s: expected column '%s.%s.%s' not included in catalog", tcase.desc, db.Name(), tbl.Name(), expectedColumn.SQLName())
		}
	}
}

func containsDB(actualDBs []catalog.Database, expectedDB *schema.Database) bool {
	for _, db := range actualDBs {
		if string(db.Name()) == expectedDB.Name() {
			return true
		}
	}
	return false
}

func containsTable(actualTables []catalog.Table, expectedTable *schema.Table) bool {
	for _, table := range actualTables {
		if table.Name() == expectedTable.SQLName() {
			return true
		}
	}
	return false
}

func containsColumn(actualColumns results.Columns, expectedColumn *schema.Column) bool {
	for _, column := range actualColumns {
		if column.Name == expectedColumn.SQLName() && column.MongoName == expectedColumn.MongoName() &&
			types.EvalTypeToSQLType(column.EvalType) == expectedColumn.SQLType() && column.MongoType == expectedColumn.MongoType() {
			return true
		}
	}
	return false
}
