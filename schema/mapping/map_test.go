package mapping_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/drdl"
	"github.com/10gen/sqlproxy/schema/mapping"
	"github.com/10gen/sqlproxy/schema/mongo"

	"go.mongodb.org/mongo-driver/bson"
)

func TestMapSchema(t *testing.T) {
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		dir := filepath.Dir(path)
		collection := filepath.Base(dir)

		switch info.Name() {

		case "lattice_schema.yml":
			// test basic json-to-relation schema mapping
			name := collection + "-lattice-map"
			t.Run(name, func(t *testing.T) {
				err := testMapSchemaFromJSON(collection, false, config.LatticeMappingMode)
				if err != nil {
					t.Fatal(err.Error())
				}
			})

		case "majority_schema.yml":
			// test basic json-to-relation schema mapping
			name := collection + "-majority-map"
			t.Run(name, func(t *testing.T) {
				err := testMapSchemaFromJSON(collection, false, config.MajorityMappingMode)
				if err != nil {
					t.Fatal(err.Error())
				}
			})

		case "prejoined.yml":
			// test json-to-relation schema mapping with prejoins
			name := collection + "-map-prejoined"
			t.Run(name, func(t *testing.T) {
				err := testMapSchemaFromJSON(collection, true, config.LatticeMappingMode)
				if err != nil {
					t.Fatal(err.Error())
				}
			})

		case "sample.json":
			// test json schema creation from sample doc
			name := collection + "-sample"
			t.Run(name, func(t *testing.T) {
				err := testMapSchemaFromSample(collection, config.LatticeMappingMode)
				if err != nil {
					t.Fatal(err.Error())
				}
			})

		}

		return nil
	}

	err := filepath.Walk(filepath.Join("testdata", "map-test"), walkFn)
	if err != nil {
		t.Fatalf("Failed to walk test files: %v", err)
	}
}

func testMapSchemaFromJSON(collection string, prejoined bool, mappingMode config.MappingMode) error {
	dir := "testdata/map-test/" + collection + "/"

	expectedFile := dir + mappingMode + "_schema.yml"
	jsonFile := dir + "schema.json"

	if prejoined {
		expectedFile = dir + "prejoined.yml"
	}

	// load the expected relational drdl
	drdlSchema, err := drdl.NewFromFile(expectedFile)
	if err != nil {
		return err
	}

	// convert the expected drdl to an expected schema
	expected, err := schema.NewFromDRDL(log.GlobalLogger(), drdlSchema, false)
	if err != nil {
		return err
	}

	// load the json schema
	jsonSchema, err := mongo.NewSchemaFromFile(jsonFile)
	if err != nil {
		return err
	}

	// test that the json schema maps to the relational schema
	return testMapSchema(collection, prejoined, jsonSchema, mappingMode, expected)
}

func testMapSchemaFromSample(collection string, mode config.MappingMode) error {
	dir := "testdata/map-test/" + collection + "/"

	expectedFile := dir + mode + "_schema.yml"
	sampleFile := dir + "sample.json"

	// load the expected relational drdl
	drdlSchema, err := drdl.NewFromFile(expectedFile)
	if err != nil {
		return err
	}

	// convert the expected drdl to an expected schema
	expected, err := schema.NewFromDRDL(log.GlobalLogger(), drdlSchema, false)
	if err != nil {
		return err
	}

	actual, err := getSchemaFromJSONSample(sampleFile)
	if err != nil {
		return err
	}

	// compare the generated json schema to the expected one
	return testMapSchema(collection, false, actual, mode, expected)
}

func testMapSchema(col string, prejoined bool, js *mongo.Schema,
	mappingMode config.MappingMode, expected *schema.Schema) error {

	// create a test database schema
	db := schema.NewDatabase(log.GlobalLogger(), "test", nil, false)

	numTables := int64(0)

	// map the json schema into the database
	err := mapping.Map(mapping.NewSchemaMappingConfig(
		db,
		js,
		col,
		prejoined,
		"old",
		[]uint8{4, 0, 0},
		log.GlobalLogger(),
		mappingMode,
		&numTables,
		10,
		200,
		1000,
		false,
	))
	if err != nil {
		return err
	}

	// create a full relational schema from the database
	actual, err := schema.New([]*schema.Database{db}, false)
	if err != nil {
		return err
	}

	// compare the generated schema to the expected one
	return actual.Equals(expected)
}

func getSchemaFromJSONSample(path string) (*mongo.Schema, error) {
	// read in the bytes from the file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	sampleBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// unmarshal schema document into bson.D
	bsonSchema := bson.D{}
	err = bson.UnmarshalExtJSON(sampleBytes, false, &bsonSchema)
	if err != nil {
		return nil, err
	}

	// create an empty json schema and include the sample doc
	mongoSchema := mongo.NewCollectionSchema()
	ft := mongo.NewNoopFieldTracker()
	err = mongoSchema.IncludeSample(bsonSchema, ft)
	return mongoSchema, err
}

func TestMapDataLake(t *testing.T) {
	testDB := "testDB"
	testColl := "testColl"

	checkCatalogTable := func(ctlgTable *catalog.MongoTable, expectedColumnNames []string, t *testing.T) {
		if ctlgTable.Collation() != collation.Default {
			t.Errorf("Expected collation to be %v, got %v", collation.Default, ctlgTable.Collation())
		}
		if !ctlgTable.IsSharded() {
			t.Error("Expected table to be marked as sharded, was not")
		}
		if len(ctlgTable.Indexes()) != 0 {
			t.Error("Did not expect indexes to be set, but they were")
		}
		if ctlgTable.Type() != catalog.BaseTable {
			t.Errorf("Expected table type to be %v, got %v", catalog.BaseTable, ctlgTable.Type())
		}
		columns := ctlgTable.Columns()
		if len(columns) != len(expectedColumnNames) {
			t.Fatalf("Expected %v columns, got %v", len(columns), len(expectedColumnNames))
		}
		for _, columnName := range expectedColumnNames {
			if _, err := ctlgTable.Column(columnName); err != nil {
				t.Errorf("Expected column %v, but did not find it", columnName)
			}
		}
	}

	tests := []struct {
		name           string
		docFileName    string
		expectedSchema map[string][]string
	}{
		{
			"one collection maps to one table",
			"simple.json",
			map[string][]string{
				testColl: {"a", "b"},
			},
		},
		{
			"one collection maps to many tables",
			"arrays.json",
			map[string][]string{
				testColl:             {"a", "b"},
				testColl + "_array1": {"array1", "array1_idx"},
				testColl + "_array2": {"array2", "array2_idx"},
			},
		},
		{
			"column mapping is case sensitive",
			"case-sensitive-fields.json",
			map[string][]string{
				testColl: {"a", "foobar", "FoObaR"},
			},
		},
		{
			"array table mapping is case sensitive",
			"case-sensitive-arrays.json",
			map[string][]string{
				testColl:            {"a", "b"},
				testColl + "_ArRAy": {"ArRAy", "ArRAy_idx"},
				testColl + "_array": {"array", "array_idx"},
			},
		},
		{
			"column mapping within array table is case sensitive",
			"case-sensitive-array-subfields.json",
			map[string][]string{
				testColl:            {"a", "b"},
				testColl + "_array": {"array", "array.fOo", "array.FOO", "array_idx"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema, err := getSchemaFromJSONSample(filepath.Join("testdata", "map-data-lake-test", test.docFileName))
			if err != nil {
				t.Fatal(err)
			}

			ctlgTables, err := mapping.MapDataLake(schema, testDB, testColl)
			if err != nil {
				t.Fatal(err)
			}

			if len(ctlgTables) != len(test.expectedSchema) {
				t.Fatalf("Expected %v tables, but mapped %v tables from %v.%v",
					len(test.expectedSchema), len(ctlgTables), testDB, testColl)
			}

			for expectedTable, expectedColumns := range test.expectedSchema {
				var ctlgTable *catalog.MongoTable
				for _, tbl := range ctlgTables {
					if tbl.Name() == expectedTable {
						ctlgTable = tbl
					}
				}

				if ctlgTable == nil {
					t.Fatalf("Expected table '%v', but didn't find it", expectedTable)
				}
				checkCatalogTable(ctlgTable, expectedColumns, t)
			}
		})
	}
}
