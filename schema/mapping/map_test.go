package mapping_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/internal/bsonutil"
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

	err := filepath.Walk("testdata/", walkFn)
	if err != nil {
		t.Fatalf("Failed to walk test files: %v", err)
	}
}

func testMapSchemaFromJSON(collection string, prejoined bool, mappingMode config.MappingMode) error {
	dir := "testdata/" + collection + "/"

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
	dir := "testdata/" + collection + "/"

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

	// read in the bytes from the sample file
	file, err := os.Open(sampleFile)
	if err != nil {
		return err
	}
	defer file.Close()
	sampleBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	// unmarshal sample document into bson.D
	sample := bson.D{}
	err = bson.UnmarshalExtJSON(sampleBytes, false, &sample)
	if err != nil {
		return err
	}

	// create an empty json schema and include the sample doc
	actual := mongo.NewCollectionSchema()
	ft := mongo.NewNoopFieldTracker()
	err = actual.IncludeSample(sample, ft)
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

func TestMapDataLake(t *testing.T) {
	testDb := "testDb"
	testColl := "testColl"

	schema, err := mongo.NewSchemaFromValue(bsonutil.NewD(
		bsonutil.NewDocElem("a", int32(1)),
		bsonutil.NewDocElem("b", int32(2)),
	))
	if err != nil {
		t.Fatal(err)
	}

	schemaWithArrays, err := mongo.NewSchemaFromValue(bsonutil.NewD(
		bsonutil.NewDocElem("a", int32(1)),
		bsonutil.NewDocElem("b", int32(2)),
		bsonutil.NewDocElem("arrayField1", bsonutil.NewArray(int32(1), int32(2))),
		bsonutil.NewDocElem("arrayField2", bsonutil.NewArray(int32(1), int32(2))),
	))
	if err != nil {
		t.Fatal(err)
	}

	checkCatalogTable := func(ctlgTable *catalog.MongoTable, expectedTableName string, t *testing.T) {
		if ctlgTable.Name() != expectedTableName {
			t.Errorf("Expected table %v, got table %v", expectedTableName, ctlgTable.Name())
		}
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
	}

	tests := []struct {
		name               string
		schema             *mongo.Schema
		db                 string
		coll               string
		expectedTableNames []string
	}{
		{
			"one collection maps to one table",
			schema,
			testDb,
			testColl,
			[]string{testColl},
		},
		{
			"one collection maps to many tables",
			schemaWithArrays,
			testDb,
			testColl,
			[]string{testColl, testColl + "_arrayField1", testColl + "_arrayField2"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctlgTables, err := mapping.MapDataLake(test.schema, test.db, test.coll)
			if err != nil {
				t.Fatal(err)
			}
			if len(ctlgTables) != len(test.expectedTableNames) {
				t.Fatalf("Expected %v tables, mapped %v tables from %v.%v",
					len(test.expectedTableNames), len(ctlgTables), testDb, testColl)
			}

			for i, ctlgTable := range ctlgTables {
				t.Run(test.expectedTableNames[i], func(t *testing.T) {
					checkCatalogTable(ctlgTable, test.expectedTableNames[i], t)
				})
			}
		})
	}
}
