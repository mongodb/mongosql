package mongosql

import (
	"context"
	"testing"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/internal/option"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/drdl"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

const (
	testMongoVersion4 = "4.0.0"
	testMongoVersion3 = "3.0.0"
	testDBName        = "test"
	testSchema        = "../testdata/resources/schema/mongosql.drdl"
	testFormat        = "none"
)

func TestTranslateSQLQuery(t *testing.T) {
	tcases := []struct {
		desc               string
		query              string
		dbName             string
		mongoVersion       string
		schema             string
		format             string
		explain            bool
		expectedError      string
		expectedOutput     string
		expectedCollection string
	}{
		{
			desc:         "query that can't be parsed/explained (command)",
			query:        "drop table foo.t",
			dbName:       testDBName,
			mongoVersion: testMongoVersion4,
			schema:       testSchema,
			format:       testFormat,
			expectedError: `fatal error executing sql "explain drop table foo.t": ERROR 1064 (42000): ` +
				`parse sql 'explain drop table foo.t' error: unexpected DROP at position 14 near drop`,
		},
		{
			desc:          "query that can't be pushed down (char_length with version < 3.4)",
			query:         "select char_length(foo.a) from foo",
			dbName:        testDBName,
			mongoVersion:  testMongoVersion3,
			schema:        testSchema,
			format:        testFormat,
			expectedError: "query not fully pushed down; run with --explain for more details",
		},
		{
			desc:          "query that can't be pushed down (cross join with local db different from foreign db)",
			query:         "select foo.a, bar.b from test.foo, test2.bar",
			dbName:        testDBName,
			mongoVersion:  testMongoVersion4,
			schema:        testSchema,
			format:        testFormat,
			expectedError: "query not fully pushed down; run with --explain for more details",
		},
		{
			desc:               "simple select query (unqualified table) correctly translated to pipeline, using dbName",
			query:              "select foo.a from foo",
			dbName:             testDBName,
			mongoVersion:       testMongoVersion4,
			schema:             testSchema,
			format:             testFormat,
			expectedOutput:     `[{"$project": {"test_DOT_foo_DOT_a": "$a","_id": NumberInt("0")}}]`,
			expectedCollection: "foo",
		},
		{
			desc:               "simple select query (qualified table) correctly translated to pipeline",
			query:              "select foo.a from test.foo",
			dbName:             testDBName,
			mongoVersion:       testMongoVersion4,
			schema:             testSchema,
			format:             testFormat,
			expectedOutput:     `[{"$project": {"test_DOT_foo_DOT_a": "$a","_id": NumberInt("0")}}]`,
			expectedCollection: "foo",
		},
		{
			desc:               "simple where query (qualified table) correctly translated to pipeline",
			query:              "select foo.a from foo where foo.b < foo.c",
			dbName:             testDBName,
			mongoVersion:       testMongoVersion4,
			schema:             testSchema,
			format:             testFormat,
			expectedOutput:     `[{"$match": {"$expr": {"$and": [{"$gt": ["$c",null]},{"$gt": ["$b",null]},{"$lt": ["$b","$c"]}]}}},{"$project": {"test_DOT_foo_DOT_a": "$a","_id": NumberInt("0")}}]`,
			expectedCollection: "foo",
		},
		{
			desc:          "invalid schema parameter - non-drdl file",
			query:         "select foo.a from foo",
			dbName:        testDBName,
			mongoVersion:  testMongoVersion4,
			schema:        "../parser/Makefile",
			format:        testFormat,
			expectedError: "fatal error getting schema from drdl: Invalid timestamp: 'MAKEFLAGS = -s\nsql.go' at line 4, column 0",
		},
		{
			desc:          "invalid schema parameter - no drdl files found in directory",
			query:         "select foo.a from foo",
			dbName:        testDBName,
			mongoVersion:  testMongoVersion4,
			schema:        "../testdata/resources",
			format:        testFormat,
			expectedError: `fatal error executing sql "explain select foo.a from foo": ERROR 1049 (42000): Unknown database 'test'`,
		},
		{
			desc:               "valid directory passed in as schema parameter",
			query:              "select foo.a from foo",
			dbName:             testDBName,
			mongoVersion:       testMongoVersion4,
			schema:             "../testdata/resources/schema",
			format:             testFormat,
			expectedOutput:     `[{"$project": {"test_DOT_foo_DOT_a": "$a","_id": NumberInt("0")}}]`,
			expectedCollection: "foo",
		},
		{
			desc:          "database doesn't exist in schema",
			query:         "select foo.a from foo",
			dbName:        "invalidDatabase",
			mongoVersion:  testMongoVersion4,
			schema:        testSchema,
			format:        testFormat,
			expectedError: `fatal error executing sql "explain select foo.a from foo": ERROR 1049 (42000): Unknown database 'invaliddatabase'`,
		},
		{
			desc:          "table doesn't exist in schema",
			query:         "select a from invalidTable",
			dbName:        testDBName,
			mongoVersion:  testMongoVersion4,
			schema:        testSchema,
			format:        testFormat,
			expectedError: `fatal error executing sql "explain select a from invalidTable": ERROR 1146 (42S02): Table 'test.invalidtable' doesn't exist`,
		},
		{
			desc:          "column doesn't exist in schema",
			query:         "select invalidColumn from foo",
			dbName:        testDBName,
			mongoVersion:  testMongoVersion4,
			schema:        testSchema,
			format:        testFormat,
			expectedError: `fatal error executing sql "explain select invalidColumn from foo": ERROR 1054 (42S22): Unknown column 'invalidColumn' in 'field list'`,
		},
		{
			desc:               "format flag one stage",
			query:              "select a from foo",
			dbName:             testDBName,
			mongoVersion:       testMongoVersion4,
			schema:             testSchema,
			format:             "multiline",
			expectedOutput:     "[\n\t{\"$project\": {\"test_DOT_foo_DOT_a\": \"$a\",\"_id\": NumberInt(\"0\")}},\n]",
			expectedCollection: "foo",
		},
		{
			desc:               "format flag multiple stages",
			query:              "select a, b from foo where a > b group by c order by b desc",
			dbName:             testDBName,
			mongoVersion:       testMongoVersion4,
			schema:             testSchema,
			format:             "multiline",
			expectedOutput:     "[\n\t{\"$match\": {\"$expr\": {\"$and\": [{\"$gt\": [\"$a\",null]},{\"$gt\": [\"$b\",null]},{\"$lt\": [\"$b\",\"$a\"]}]}}},\n\t{\"$group\": {\"_id\": {\"group_key_0\": \"$c\"},\"test_DOT_foo_DOT_b\": {\"$first\": \"$b\"},\"test_DOT_foo_DOT_a\": {\"$first\": \"$a\"}}},\n\t{\"$sort\": {\"test_DOT_foo_DOT_b\": NumberInt(\"-1\")}},\n\t{\"$project\": {\"test_DOT_foo_DOT_a\": \"$test_DOT_foo_DOT_a\",\"test_DOT_foo_DOT_b\": \"$test_DOT_foo_DOT_b\",\"_id\": NumberInt(\"0\")}},\n]",
			expectedCollection: "foo",
		},
		{
			desc:               "format flag, pipeline contains $lookup with pipeline field",
			query:              "select foo.a, baz.b from foo join baz where foo.a > baz.b",
			dbName:             testDBName,
			mongoVersion:       testMongoVersion4,
			schema:             testSchema,
			format:             "multiline",
			expectedOutput:     "[\n\t{\"$lookup\": {\"from\": \"baz\",\"let\": {\"local_table__a\": \"$a\"},\"pipeline\": [{\"$match\": {\"$expr\": {\"$and\": [{\"$gt\": [\"$$local_table__a\",null]},{\"$gt\": [\"$b\",null]},{\"$lt\": [\"$b\",\"$$local_table__a\"]}]}}}],\"as\": \"__joined_baz\"}},\n\t{\"$unwind\": \"$__joined_baz\"},\n\t{\"$project\": {\"test_DOT_foo_DOT_a\": \"$a\",\"test_DOT_baz_DOT_b\": \"$__joined_baz.b\",\"_id\": NumberInt(\"0\")}},\n]",
			expectedCollection: "foo",
		},
		{
			desc:               "explain flag, no formatting, fully pushed down",
			query:              "select a from foo",
			dbName:             testDBName,
			mongoVersion:       testMongoVersion4,
			schema:             testSchema,
			format:             testFormat,
			explain:            true,
			expectedOutput:     `[{"ID":1,"StageType":"MongoSourceStage","Columns":"[{name: test.foo.'a', type: 'float'}]","Sources":null,"Database":{},"Tables":{},"Aliases":{},"Collections":{},"Pipeline":{},"PipelineExplain":{},"PushdownFailures":null}]`,
			expectedCollection: "foo",
		},
		{
			desc:               "explain flag, with formatting, fully pushed down",
			query:              "select a from foo",
			dbName:             testDBName,
			mongoVersion:       testMongoVersion4,
			schema:             testSchema,
			format:             "multiline",
			explain:            true,
			expectedOutput:     "[\n\t{\n\t\t\"ID\": 1,\n\t\t\"StageType\": \"MongoSourceStage\",\n\t\t\"Columns\": \"[{name: test.foo.'a', type: 'float'}]\",\n\t\t\"Sources\": null,\n\t\t\"Database\": {},\n\t\t\"Tables\": {},\n\t\t\"Aliases\": {},\n\t\t\"Collections\": {},\n\t\t\"Pipeline\": {},\n\t\t\"PipelineExplain\": {},\n\t\t\"PushdownFailures\": null\n\t}\n]",
			expectedCollection: "foo",
		},
		{
			desc:               "explain flag, no formatting, not fully pushed down",
			query:              "select adddate(a, 1) from foo",
			dbName:             testDBName,
			mongoVersion:       testMongoVersion3,
			schema:             testSchema,
			format:             testFormat,
			explain:            true,
			expectedOutput:     `[{"ID":1,"StageType":"ProjectStage","Columns":"[{name: 'adddate(a, 1, day)', type: 'timestamp'}]","Sources":[2],"Database":{},"Tables":{},"Aliases":{},"Collections":{},"Pipeline":{},"PipelineExplain":{},"PushdownFailures":[{"name":"SQLConvertExpr","reason":"cannot push down mongosql-mode conversions to MongoDB < 4.0"}]},{"ID":2,"StageType":"MongoSourceStage","Columns":"[{name: test.foo.'a', type: 'float'}]","Sources":null,"Database":{},"Tables":{},"Aliases":{},"Collections":{},"Pipeline":{},"PipelineExplain":{},"PushdownFailures":null}]`,
			expectedCollection: "foo",
		},
		{
			desc:               "explain flag, with formatting, not fully pushed down",
			query:              "select adddate(a, 1) from foo",
			dbName:             testDBName,
			mongoVersion:       testMongoVersion3,
			schema:             testSchema,
			format:             "multiline",
			explain:            true,
			expectedOutput:     "[\n\t{\n\t\t\"ID\": 1,\n\t\t\"StageType\": \"ProjectStage\",\n\t\t\"Columns\": \"[{name: 'adddate(a, 1, day)', type: 'timestamp'}]\",\n\t\t\"Sources\": [\n\t\t\t2\n\t\t],\n\t\t\"Database\": {},\n\t\t\"Tables\": {},\n\t\t\"Aliases\": {},\n\t\t\"Collections\": {},\n\t\t\"Pipeline\": {},\n\t\t\"PipelineExplain\": {},\n\t\t\"PushdownFailures\": [\n\t\t\t{\n\t\t\t\t\"name\": \"SQLConvertExpr\",\n\t\t\t\t\"reason\": \"cannot push down mongosql-mode conversions to MongoDB < 4.0\"\n\t\t\t}\n\t\t]\n\t},\n\t{\n\t\t\"ID\": 2,\n\t\t\"StageType\": \"MongoSourceStage\",\n\t\t\"Columns\": \"[{name: test.foo.'a', type: 'float'}]\",\n\t\t\"Sources\": null,\n\t\t\"Database\": {},\n\t\t\"Tables\": {},\n\t\t\"Aliases\": {},\n\t\t\"Collections\": {},\n\t\t\"Pipeline\": {},\n\t\t\"PipelineExplain\": {},\n\t\t\"PushdownFailures\": null\n\t}\n]",
			expectedCollection: "foo",
		},
		{
			desc:               "mongoVersion = latest",
			query:              "select adddate(a, 1) from foo",
			dbName:             testDBName,
			mongoVersion:       "latest",
			schema:             testSchema,
			format:             testFormat,
			expectedOutput:     `[{"$project": {"adddate(Convert(test_DOT_foo_DOT_a, timestamp),1,day)": {"$let": {"vars": {"date": {"$convert": {"input": "$a","to": {"$literal": "date"},"onError": {"$literal": null},"onNull": {"$literal": null}}}},"in": {"$cond": {"if": {"$lte": ["$$date",{"$literal": null}]},"then": {"$literal": null},"else": {"$add": ["$$date",{"$literal": NumberLong("86400000")}]}}}}},"_id": NumberInt("0")}}]`,
			expectedCollection: "foo",
		},
	}

	for _, tcase := range tcases {
		t.Run(tcase.desc, func(t *testing.T) {
			actualOutput, actualCollection, err := TranslateSQLQuery(tcase.query, tcase.dbName, tcase.mongoVersion, tcase.schema, tcase.format, tcase.explain)

			if tcase.expectedError != "" {
				if err == nil {
					t.Errorf("%s: expected error, but no error was returned", tcase.desc)
				} else if tcase.expectedError != err.Error() {
					t.Errorf(`Error received does not match expected error. Expected "%v", got "%v".`, tcase.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if actualOutput != tcase.expectedOutput {
					t.Fatalf("%s: actual output is not same as expected (+++ actual, --- expected)\n+++ %s\n--- %s\n", tcase.desc, actualOutput, tcase.expectedOutput)
				}
				if actualCollection != tcase.expectedCollection {
					t.Fatalf("%s: actual collection is not same as expected (+++ actual, --- expected)\n+++ %s\n--- %s\n", tcase.desc, actualCollection, tcase.expectedCollection)
				}
			}
		})
	}
}

func TestTranslateSQLQueryFile(t *testing.T) {
	tcases := []struct {
		desc               string
		queryFile          string
		dbName             string
		mongoVersion       string
		schema             string
		format             string
		explain            bool
		expectedError      string
		expectedOutput     string
		expectedCollection string
	}{
		{
			desc:               "query file simple",
			queryFile:          "testdata/simple.txt",
			dbName:             testDBName,
			mongoVersion:       testMongoVersion4,
			schema:             testSchema,
			format:             testFormat,
			expectedOutput:     `[{"$project": {"test_DOT_foo_DOT_a": "$a","_id": NumberInt("0")}}]`,
			expectedCollection: "foo",
		},
		{
			desc:               "query file with backticks",
			queryFile:          "testdata/backticks.txt",
			dbName:             testDBName,
			mongoVersion:       testMongoVersion4,
			schema:             "../testdata/resources/schema/schema_Members.drdl",
			format:             testFormat,
			expectedOutput:     `[{"$unwind": {"path": "$member.MemberAttributeValues.MemberAttributeValue","includeArrayIndex": "member.MemberAttributeValues.MemberAttributeValue_idx"}},{"$match": {"$and": [{"member.MemberAttributeValues.MemberAttributeValue.Void": {"$eq": false}},{"member.MemberAttributeValues.MemberAttributeValue.EndDate": {"$gte": ISODate("2010-09-18T00:00:00")}},{"member.MemberAttributeValues.MemberAttributeValue.StartDate": {"$lte": ISODate("2010-09-18T00:00:00")}}]}},{"$project": {"UMV_DOT_M_DOT__id": "$_id","UMV_DOT_MA_DOT_member_DOT_MemberAttributeValues_DOT_MemberAttributeValue_DOT_Attribute": "$member.MemberAttributeValues.MemberAttributeValue.Attribute","UMV_DOT_MA_DOT_member_DOT_MemberAttributeValues_DOT_MemberAttributeValue_DOT_AttributeCode": "$member.MemberAttributeValues.MemberAttributeValue.AttributeCode","case when (not UMV_DOT_MA_DOT_member_DOT_MemberAttributeValues_DOT_MemberAttributeValue_DOT_CustomValue is NULL) then UMV_DOT_MA_DOT_member_DOT_MemberAttributeValues_DOT_MemberAttributeValue_DOT_CustomValue when (not UMV_DOT_MA_DOT_member_DOT_MemberAttributeValues_DOT_MemberAttributeValue_DOT_DefinedValue is NULL) then UMV_DOT_MA_DOT_member_DOT_MemberAttributeValues_DOT_MemberAttributeValue_DOT_DefinedValue else NULL end": {"$cond": {"if": {"$let": {"vars": {"arg": {"$lte": ["$member.MemberAttributeValues.MemberAttributeValue.CustomValue",{"$literal": null}]}},"in": {"$cond": {"if": {"$lte": ["$$arg",{"$literal": null}]},"then": {"$literal": null},"else": {"$not": ["$$arg"]}}}}},"then": "$member.MemberAttributeValues.MemberAttributeValue.CustomValue","else": {"$cond": {"if": {"$let": {"vars": {"arg": {"$lte": ["$member.MemberAttributeValues.MemberAttributeValue.DefinedValue",{"$literal": null}]}},"in": {"$cond": {"if": {"$lte": ["$$arg",{"$literal": null}]},"then": {"$literal": null},"else": {"$not": ["$$arg"]}}}}},"then": "$member.MemberAttributeValues.MemberAttributeValue.DefinedValue","else": {"$literal": null}}}}},"UMV_DOT_MA_DOT_member_DOT_MemberAttributeValues_DOT_MemberAttributeValue_DOT_CustomValue": "$member.MemberAttributeValues.MemberAttributeValue.CustomValue","UMV_DOT_MA_DOT_member_DOT_MemberAttributeValues_DOT_MemberAttributeValue_DOT_DefinedValue": "$member.MemberAttributeValues.MemberAttributeValue.DefinedValue","UMV_DOT_MA_DOT_member_DOT_MemberAttributeValues_DOT_MemberAttributeValue_DOT_StartDate": "$member.MemberAttributeValues.MemberAttributeValue.StartDate","UMV_DOT_MA_DOT_member_DOT_MemberAttributeValues_DOT_MemberAttributeValue_DOT_EndDate": "$member.MemberAttributeValues.MemberAttributeValue.EndDate","UMV_DOT_MA_DOT_member_DOT_MemberAttributeValues_DOT_MemberAttributeValue_DOT_IsChanged": "$member.MemberAttributeValues.MemberAttributeValue.IsChanged","UMV_DOT_MA_DOT_member_DOT_MemberAttributeValues_DOT_MemberAttributeValue_DOT_LastUpdatedUser": "$member.MemberAttributeValues.MemberAttributeValue.LastUpdatedUser","UMV_DOT_MA_DOT_member_DOT_MemberAttributeValues_DOT_MemberAttributeValue_DOT_Source": "$member.MemberAttributeValues.MemberAttributeValue.Source","UMV_DOT_MA_DOT_member_DOT_MemberAttributeValues_DOT_MemberAttributeValue_DOT_Void": "$member.MemberAttributeValues.MemberAttributeValue.Void","_id": NumberInt("0")}}]`,
			expectedCollection: "Members",
		},
		{
			desc:          "query file doesn't exist",
			queryFile:     "hello_world.txt",
			dbName:        testDBName,
			mongoVersion:  testMongoVersion4,
			schema:        testSchema,
			format:        testFormat,
			expectedError: "could not open file hello_world.txt",
		},
	}

	for _, tcase := range tcases {
		t.Run(tcase.desc, func(t *testing.T) {
			actualOutput, actualCollection, err := TranslateSQLQueryFile(tcase.queryFile, tcase.dbName, tcase.mongoVersion, tcase.schema, tcase.format, tcase.explain)

			if tcase.expectedError != "" {
				if err == nil {
					t.Errorf("%s: expected error, but no error was returned", tcase.desc)
				} else if tcase.expectedError != err.Error() {
					t.Errorf(`Error received does not match expected error. Expected "%v", got "%v".`, tcase.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if actualOutput != tcase.expectedOutput {
					t.Fatalf("%s: actual output is not same as expected (+++ actual, --- expected)\n+++ %s\n--- %s\n", tcase.desc, actualOutput, tcase.expectedOutput)
				}
				if actualCollection != tcase.expectedCollection {
					t.Fatalf("%s: actual collection is not same as expected (+++ actual, --- expected)\n+++ %s\n--- %s\n", tcase.desc, actualCollection, tcase.expectedCollection)
				}
			}
		})
	}
}

func TestTranslateSQLQueryRaw(t *testing.T) {
	drdlSch, err := drdl.NewFromBytes([]byte(`
schema:
- db: test
  tables:
  - table: foo
    collection: foo
    pipeline: []
    columns:
    - Name: a
      MongoType: int
      SqlName: a
      SqlType: int32
    - Name: b
      MongoType: int
      SqlName: b
      SqlType: int32
    - Name: c
      MongoType: int
      SqlName: c
      SqlType: int32
  - table: bar
    collection: bar
    pipeline: []
    columns:
    - Name: a
      MongoType: int
      SQLName: a
      SQLType: int32
    - Name: d
      MongoType: int
      SQLName: d
      SQLType: int32
  - table: uuids
    collection: uuids
    pipeline: []
    columns:
    - Name: _id
      MongoType: bson.UUID
      SQLName: _id
      SQLType: string
- db: test2
  tables:
  - table: bar
    collection: bar
    pipeline: []
    columns:
    - Name: b
      MongoType: int
      SqlName: b
      SqlType: int32
`))
	if err != nil {
		t.Fatalf("error creating DRDL schema: %v", err)
	}

	testSchema, err := schema.NewFromDRDL(nil, drdlSch)
	if err != nil {
		t.Fatalf("error creating test schema: %v", err)
	}

	testCatalogWithoutSharding, err := getCatalog(testSchema)
	if err != nil {
		t.Fatalf("error creating test catalog (no sharding): %v", err)
	}

	testCatalogWithSharding, err := makeTestCatalogWithShardedCollections(testSchema)
	if err != nil {
		t.Fatalf("error creating test catalog (sharding): %v", err)
	}

	tcases := []struct {
		desc               string
		query              string
		dbName             string
		ctlg               catalog.Catalog
		expectedError      string
		expectedOutput     string
		expectedDatabase   string
		expectedCollection string
	}{
		{
			desc:          "query that can't be parsed/explained (command)",
			query:         "drop table foo.t",
			dbName:        testDBName,
			ctlg:          testCatalogWithoutSharding,
			expectedError: `fatal error executing sql "drop table foo.t": command not supported`,
		},
		{
			desc:          "query that can't be pushed down (rand cannot be pushed down)",
			query:         "select rand(foo.a) from foo",
			dbName:        testDBName,
			ctlg:          testCatalogWithoutSharding,
			expectedError: `fatal error executing sql "select rand(foo.a) from foo": query not fully pushed down`,
		},
		{
			desc:          "database doesn't exist in schema",
			query:         "select foo.a from foo",
			dbName:        "invalidDatabase",
			ctlg:          testCatalogWithoutSharding,
			expectedError: `fatal error executing sql "select foo.a from foo": ERROR 1049 (42000): Unknown database 'invaliddatabase'`,
		},
		{
			desc:          "table doesn't exist in schema",
			query:         "select a from invalidTable",
			dbName:        testDBName,
			ctlg:          testCatalogWithoutSharding,
			expectedError: `fatal error executing sql "select a from invalidTable": ERROR 1146 (42S02): Table 'test.invalidtable' doesn't exist`,
		},
		{
			desc:          "column doesn't exist in schema",
			query:         "select invalidColumn from foo",
			dbName:        testDBName,
			ctlg:          testCatalogWithoutSharding,
			expectedError: `fatal error executing sql "select invalidColumn from foo": ERROR 1054 (42S22): Unknown column 'invalidColumn' in 'field list'`,
		},
		{
			desc:               "simple select query (unqualified table) correctly translated to pipeline, using dbName",
			query:              "select foo.a from foo",
			dbName:             testDBName,
			ctlg:               testCatalogWithoutSharding,
			expectedOutput:     `[{"$project": {"values": [{"db": {"$literal": "test"},"table": {"$literal": "foo"},"tableAlias": {"$literal": "foo"},"column": {"$literal": "a"},"columnAlias": {"$literal": "a"},"value": "$a"}],"_id": {"$numberInt":"0"}}}]`,
			expectedDatabase:   testDBName,
			expectedCollection: "foo",
		},
		{
			desc:               "simple select query (qualified table) correctly translated to pipeline",
			query:              "select foo.a from test.foo",
			dbName:             testDBName,
			ctlg:               testCatalogWithoutSharding,
			expectedOutput:     `[{"$project": {"values": [{"db": {"$literal": "test"},"table": {"$literal": "foo"},"tableAlias": {"$literal": "foo"},"column": {"$literal": "a"},"columnAlias": {"$literal": "a"},"value": "$a"}],"_id": {"$numberInt":"0"}}}]`,
			expectedDatabase:   testDBName,
			expectedCollection: "foo",
		},
		{
			desc:               "simple select query (qualified table with different name than default) correctly translated to pipeline",
			query:              "select bar.b from test2.bar",
			dbName:             testDBName,
			ctlg:               testCatalogWithoutSharding,
			expectedOutput:     `[{"$project": {"values": [{"db": {"$literal": "test2"},"table": {"$literal": "bar"},"tableAlias": {"$literal": "bar"},"column": {"$literal": "b"},"columnAlias": {"$literal": "b"},"value": "$b"}],"_id": {"$numberInt":"0"}}}]`,
			expectedDatabase:   "test2",
			expectedCollection: "bar",
		},
		{
			desc:               "simple select query with computed columns correctly translated to pipeline",
			query:              "select a, b, a+b from foo",
			dbName:             testDBName,
			ctlg:               testCatalogWithoutSharding,
			expectedOutput:     `[{"$project": {"values": [{"db": {"$literal": "test"},"table": {"$literal": "foo"},"tableAlias": {"$literal": "foo"},"column": {"$literal": "a"},"columnAlias": {"$literal": "a"},"value": "$a"},{"db": {"$literal": "test"},"table": {"$literal": "foo"},"tableAlias": {"$literal": "foo"},"column": {"$literal": "b"},"columnAlias": {"$literal": "b"},"value": "$b"},{"db": {"$literal": "test"},"table": {"$literal": null},"tableAlias": {"$literal": null},"column": {"$literal": null},"columnAlias": {"$literal": "a+b"},"value": {"$add": ["$a","$b"]}}],"_id": {"$numberInt":"0"}}}]`,
			expectedDatabase:   testDBName,
			expectedCollection: "foo",
		},
		{
			desc:               "query with where clause correctly translated to pipeline",
			query:              "select foo.a from foo where foo.b < foo.c",
			dbName:             testDBName,
			ctlg:               testCatalogWithoutSharding,
			expectedOutput:     `[{"$match": {"$expr": {"$and": [{"$gt": ["$c",null]},{"$gt": ["$b",null]},{"$lt": ["$b","$c"]}]}}},{"$project": {"values": [{"db": {"$literal": "test"},"table": {"$literal": "foo"},"tableAlias": {"$literal": "foo"},"column": {"$literal": "a"},"columnAlias": {"$literal": "a"},"value": "$a"}],"_id": {"$numberInt":"0"}}}]`,
			expectedDatabase:   testDBName,
			expectedCollection: "foo",
		},
		{
			desc:               "query with column and table aliases correctly translated to pipeline",
			query:              "select a as my_a, b as b_cool, a+b as sum from foo t",
			dbName:             testDBName,
			ctlg:               testCatalogWithoutSharding,
			expectedOutput:     `[{"$project": {"values": [{"db": {"$literal": "test"},"table": {"$literal": "foo"},"tableAlias": {"$literal": "t"},"column": {"$literal": "a"},"columnAlias": {"$literal": "my_a"},"value": "$a"},{"db": {"$literal": "test"},"table": {"$literal": "foo"},"tableAlias": {"$literal": "t"},"column": {"$literal": "b"},"columnAlias": {"$literal": "b_cool"},"value": "$b"},{"db": {"$literal": "test"},"table": {"$literal": null},"tableAlias": {"$literal": null},"column": {"$literal": null},"columnAlias": {"$literal": "sum"},"value": {"$add": ["$a","$b"]}}],"_id": {"$numberInt":"0"}}}]`,
			expectedDatabase:   testDBName,
			expectedCollection: "foo",
		},
		{
			desc:               "query with lookup in sharded collection",
			query:              "select * from foo join bar on foo.a = bar.a",
			dbName:             testDBName,
			ctlg:               testCatalogWithSharding,
			expectedOutput:     `[{"$match": {"a": {"$ne": null}}},{"$lookup": {"from": {"db": "test","coll": "bar"},"localField": "a","foreignField": "a","as": "__joined_bar"}},{"$unwind": "$__joined_bar"},{"$project": {"values": [{"db": {"$literal": "test"},"table": {"$literal": "foo"},"tableAlias": {"$literal": "foo"},"column": {"$literal": "a"},"columnAlias": {"$literal": "a"},"value": "$a"},{"db": {"$literal": "test"},"table": {"$literal": "foo"},"tableAlias": {"$literal": "foo"},"column": {"$literal": "b"},"columnAlias": {"$literal": "b"},"value": "$b"},{"db": {"$literal": "test"},"table": {"$literal": "foo"},"tableAlias": {"$literal": "foo"},"column": {"$literal": "c"},"columnAlias": {"$literal": "c"},"value": "$c"},{"db": {"$literal": "test"},"table": {"$literal": "bar"},"tableAlias": {"$literal": "bar"},"column": {"$literal": "a"},"columnAlias": {"$literal": "a"},"value": "$__joined_bar.a"},{"db": {"$literal": "test"},"table": {"$literal": "bar"},"tableAlias": {"$literal": "bar"},"column": {"$literal": "d"},"columnAlias": {"$literal": "d"},"value": "$__joined_bar.d"}],"_id": {"$numberInt":"0"}}}]`,
			expectedDatabase:   testDBName,
			expectedCollection: "foo",
		},
		{
			desc:               "query with cross-database lookup",
			query:              "select * from test.foo join test2.bar",
			dbName:             testDBName,
			ctlg:               testCatalogWithoutSharding,
			expectedOutput:     `[{"$lookup": {"from": {"db": "test2","coll": "bar"},"let": {},"pipeline": [],"as": "__joined_bar"}},{"$unwind": "$__joined_bar"},{"$project": {"values": [{"db": {"$literal": "test"},"table": {"$literal": "foo"},"tableAlias": {"$literal": "foo"},"column": {"$literal": "a"},"columnAlias": {"$literal": "a"},"value": "$a"},{"db": {"$literal": "test"},"table": {"$literal": "foo"},"tableAlias": {"$literal": "foo"},"column": {"$literal": "b"},"columnAlias": {"$literal": "b"},"value": "$b"},{"db": {"$literal": "test"},"table": {"$literal": "foo"},"tableAlias": {"$literal": "foo"},"column": {"$literal": "c"},"columnAlias": {"$literal": "c"},"value": "$c"},{"db": {"$literal": "test2"},"table": {"$literal": "bar"},"tableAlias": {"$literal": "bar"},"column": {"$literal": "b"},"columnAlias": {"$literal": "b"},"value": "$__joined_bar.b"}],"_id": {"$numberInt":"0"}}}]`,
			expectedDatabase:   testDBName,
			expectedCollection: "foo",
		},
		{
			desc:               "no row generator optimization",
			query:              "select 1 + 2 from test.foo",
			dbName:             testDBName,
			ctlg:               testCatalogWithoutSharding,
			expectedOutput:     `[{"$project": {"values": [{"db": {"$literal": null},"table": {"$literal": null},"tableAlias": {"$literal": null},"column": {"$literal": null},"columnAlias": {"$literal": "1+2"},"value": {"$literal": {"$numberLong":"3"}}}],"_id": {"$numberInt":"0"}}}]`,
			expectedDatabase:   testDBName,
			expectedCollection: "foo",
		},
		{
			desc:               "no count optimization",
			query:              "select count(*) from foo",
			dbName:             testDBName,
			ctlg:               testCatalogWithoutSharding,
			expectedOutput:     `[{"$group": {"_id": {},"count(*)": {"$sum": {"$numberInt":"1"}}}},{"$project": {"values": [{"db": {"$literal": null},"table": {"$literal": null},"tableAlias": {"$literal": null},"column": {"$literal": null},"columnAlias": {"$literal": "count(*)"},"value": "$count(*)"}],"_id": {"$numberInt":"0"}}}]`,
			expectedDatabase:   testDBName,
			expectedCollection: "foo",
		},
		{
			desc:               "pushdown dual as information_schema table",
			query:              "select 1+2",
			dbName:             testDBName,
			ctlg:               testCatalogWithoutSharding,
			expectedOutput:     `[{"$project": {"values": [{"db": {"$literal": null},"table": {"$literal": null},"tableAlias": {"$literal": null},"column": {"$literal": null},"columnAlias": {"$literal": "1+2"},"value": {"$literal": {"$numberLong":"3"}}}],"_id": {"$numberInt":"0"}}}]`,
			expectedDatabase:   "information_schema",
			expectedCollection: "dual",
		},
		{
			desc:          "cannot pushdown UUID-to-non-UUID comparison in select clause",
			query:         `select _id = "123" from uuids`,
			dbName:        testDBName,
			ctlg:          testCatalogWithoutSharding,
			expectedError: `fatal error executing sql "select _id = \"123\" from uuids": query not fully pushed down`,
		},
		{
			desc:          "cannot pushdown UUID-to-non-UUID comparison in where clause",
			query:         `select * from uuids where _id = "123"`,
			dbName:        testDBName,
			ctlg:          testCatalogWithoutSharding,
			expectedError: `fatal error executing sql "select * from uuids where _id = \"123\"": query not fully pushed down`,
		},
	}

	for _, tcase := range tcases {
		t.Run(tcase.desc, func(t *testing.T) {
			tCfg := NewTranslationConfig(tcase.ctlg, evaluator.ODBCOutputFormat, 1, tcase.dbName)
			actualOutputRaw, actualDatabase, actualCollection, err := TranslateSQLQueryRaw(context.Background(), tCfg, tcase.query)

			if tcase.expectedError != "" {
				if err == nil {
					t.Errorf("%s: expected error, but no error was returned", tcase.desc)
				} else if tcase.expectedError != err.Error() {
					t.Errorf(`Error received does not match expected error. Expected "%v", got "%v".`, tcase.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				// We create a bsoncore.Value with Type bsontype.Array because (bsoncore.Value).String()
				// checks the Type field and produces an appropriate string.
				// actualOutputRaw.String() would produce a string representation of a document, not an
				// array, because bsoncore.Array is an alias for bsoncore.Document.
				actualOutput := bsoncore.Value{
					Type: bsontype.Array,
					Data: actualOutputRaw,
				}.String()

				if actualOutput != tcase.expectedOutput {
					t.Fatalf("%s: actual output is not same as expected (+++ actual, --- expected)\n+++ %s\n--- %s\n", tcase.desc, actualOutput, tcase.expectedOutput)
				}
				if actualDatabase != tcase.expectedDatabase {
					t.Fatalf("%s: actual database is not same as expected (+++ actual, --- expected)\n+++ %s\n--- %s\n", tcase.desc, actualDatabase, tcase.expectedDatabase)
				}
				if actualCollection != tcase.expectedCollection {
					t.Fatalf("%s: actual collection is not same as expected (+++ actual, --- expected)\n+++ %s\n--- %s\n", tcase.desc, actualCollection, tcase.expectedCollection)
				}
			}
		})
	}
}

func makeTestCatalogWithShardedCollections(s *schema.Schema) (catalog.Catalog, error) {
	databases := s.Databases()
	dbInfos := make(map[mongodb.DatabaseName]*mongodb.DatabaseInfo, len(databases))

	for _, db := range databases {
		tables := db.Tables()
		collectionInfos := make(map[mongodb.CollectionName]*mongodb.CollectionInfo, len(tables))
		for _, t := range tables {
			collectionName := mongodb.CollectionName(t.MongoName())
			collectionInfo := &mongodb.CollectionInfo{
				Name:      collectionName,
				IsSharded: true,
			}

			collectionInfos[collectionName] = collectionInfo
		}

		dbName := mongodb.DatabaseName(db.Name())
		dbInfo := &mongodb.DatabaseInfo{
			Name:        dbName,
			Privileges:  mongodb.AllPrivileges,
			Collections: collectionInfos,
		}

		dbInfos[dbName] = dbInfo
	}

	info := &mongodb.Info{
		Databases: dbInfos,
	}

	return catalog.BuildFromSchema(s, info, false)
}

type ctlgTest struct {
	desc             string
	relationalSchema *schema.Schema
}

func TestGetVariables(t *testing.T) {
	// test that mongoVersion is set as expected
	t.Run("correctly set MongoDB version", func(t *testing.T) {
		tcases := []struct {
			mdbVersion string
		}{
			{"3.2.1"},
			{"3.6.0"},
			{"4.0.2"},
		}

		for _, tc := range tcases {
			vars := getVariables(tc.mdbVersion)

			for _, scope := range []variable.Scope{variable.GlobalScope, variable.SessionScope} {
				actualMdbVersion, err := vars.Get(variable.MongoDBVersion, scope, variable.SystemKind)
				if err != nil {
					t.Fatalf("unexpected error getting MongoDB Version variable for %s scope: %v", scope, err)
				}

				if tc.mdbVersion != actualMdbVersion.Value() {
					t.Fatalf("MongoDB Versions do not match, expected: %s, actual: %s", tc.mdbVersion, actualMdbVersion)
				}
			}
		}
	})
}

func TestGetCatalog(t *testing.T) {

	lg := log.GlobalLogger()

	c1 := schema.NewColumn("c1", schema.SQLInt, "c1", schema.MongoInt, false, option.NoneString())
	c2 := schema.NewColumn("c2", schema.SQLInt, "c2", schema.MongoInt, false, option.NoneString())
	c3 := schema.NewColumn("c3", schema.SQLInt, "c3", schema.MongoInt, false, option.NoneString())
	c4 := schema.NewColumn("c4", schema.SQLInt, "c4", schema.MongoInt, false, option.NoneString())
	c5 := schema.NewColumn("c5", schema.SQLInt, "c5", schema.MongoInt, false, option.NoneString())
	c6 := schema.NewColumn("c6", schema.SQLInt, "c6", schema.MongoInt, false, option.NoneString())
	c7 := schema.NewColumn("c7", schema.SQLInt, "c7", schema.MongoInt, false, option.NoneString())
	c8 := schema.NewColumn("c8", schema.SQLInt, "c8", schema.MongoInt, false, option.NoneString())

	tOneColumn, _ := schema.NewTable(lg, "t1", "t1", []bson.D{}, []*schema.Column{c1}, []schema.Index{}, option.NoneString())
	tManyColumn, _ := schema.NewTable(lg, "t2", "t2", []bson.D{}, []*schema.Column{c1, c2}, []schema.Index{}, option.NoneString())
	tManyColumn2, _ := schema.NewTable(lg, "t3", "t3", []bson.D{}, []*schema.Column{c3, c4}, []schema.Index{}, option.NoneString())
	tManyColumn3, _ := schema.NewTable(lg, "t4", "t4", []bson.D{}, []*schema.Column{c5, c6}, []schema.Index{}, option.NoneString())
	tManyColumn4, _ := schema.NewTable(lg, "t5", "t5", []bson.D{}, []*schema.Column{c7, c8}, []schema.Index{}, option.NoneString())
	tManyColumn5, _ := schema.NewTable(lg, "t6", "t6", []bson.D{}, []*schema.Column{c1, c2}, []schema.Index{}, option.NoneString())
	tManyColumn6, _ := schema.NewTable(lg, "t7", "t7", []bson.D{}, []*schema.Column{c3, c4}, []schema.Index{}, option.NoneString())

	dbOneTableOneColumn := schema.NewDatabase(lg, "db1", []*schema.Table{tOneColumn})
	dbOneTableManyColumn := schema.NewDatabase(lg, "db2", []*schema.Table{tManyColumn})
	dbManyTableManyColumn := schema.NewDatabase(lg, "db3", []*schema.Table{tManyColumn, tManyColumn2})
	dbManyTableManyColumn2 := schema.NewDatabase(lg, "db4", []*schema.Table{tManyColumn3, tManyColumn4})
	dbManyTableManySameColumn := schema.NewDatabase(lg, "db5", []*schema.Table{tManyColumn, tManyColumn5})
	dbManyTableManySameColumn2 := schema.NewDatabase(lg, "db6", []*schema.Table{tManyColumn2, tManyColumn6})

	schema1, _ := schema.New([]*schema.Database{dbOneTableOneColumn})
	schema2, _ := schema.New([]*schema.Database{dbOneTableManyColumn})
	schema3, _ := schema.New([]*schema.Database{dbManyTableManyColumn})
	schema4, _ := schema.New([]*schema.Database{dbManyTableManySameColumn})
	schema5, _ := schema.New([]*schema.Database{dbManyTableManyColumn, dbManyTableManyColumn2})
	schema6, _ := schema.New([]*schema.Database{dbManyTableManySameColumn, dbManyTableManySameColumn2})

	// test that schema is mapped to catalog as expected
	t.Run("correctly map schema to catalog", func(t *testing.T) {

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
			ctlg, err := getCatalog(tcase.relationalSchema)
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
	testCtx := context.Background()
	expectedDBs := tcase.relationalSchema.Databases()
	actualDBs, _ := ctlg.Databases(testCtx)
	if len(expectedDBs) != len(actualDBs) {
		t.Fatalf("Catalog databases don't match expected value: want %v, got %v.", expectedDBs, actualDBs)
	}

	for _, expectedDB := range expectedDBs {
		// ensure all expected databases were found in the catalog.
		if !containsDB(actualDBs, expectedDB) {
			t.Errorf("%s: expected database '%s' not included in catalog", tcase.desc, expectedDB.Name())
		}
		currentDB, err := ctlg.Database(testCtx, expectedDB.Name())
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tcase.desc, err)
		}
		validateDatabase(t, tcase, currentDB)
	}
}

func validateDatabase(t *testing.T, tcase ctlgTest, db catalog.Database) {
	testCtx := context.Background()
	expectedTables := tcase.relationalSchema.Database(string(db.Name())).Tables()
	actualTables, _ := db.Tables(testCtx)
	if len(expectedTables) != len(actualTables) {
		t.Fatalf("Catalog tables don't match expected value: want %v, got %v.", expectedTables, actualTables)
	}

	for _, expectedTable := range expectedTables {
		// ensure all expected tables were found in the catalog.
		if !containsTable(actualTables, expectedTable) {
			t.Errorf("%s: expected table '%s.%s' not included in catalog", tcase.desc, db.Name(), expectedTable.SQLName())
		}
		currentTable, err := db.Table(testCtx, expectedTable.SQLName())
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
