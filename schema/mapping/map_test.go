package mapping_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

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
	expected, err := schema.NewFromDRDL(log.GlobalLogger(), drdlSchema)
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
	expected, err := schema.NewFromDRDL(log.GlobalLogger(), drdlSchema)
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
	err = actual.IncludeSample(sample)
	if err != nil {
		return err
	}

	// compare the generated json schema to the expected one
	return testMapSchema(collection, false, actual, mode, expected)
}

func testMapSchema(col string, prejoined bool, js *mongo.Schema,
	mappingMode config.MappingMode, expected *schema.Schema) error {

	// create a test database schema
	db := schema.NewDatabase(log.GlobalLogger(), "test", nil)

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
		1000,
		50))
	if err != nil {
		return err
	}

	// create a full relational schema from the database
	actual, err := schema.New([]*schema.Database{db}, nil)
	if err != nil {
		return err
	}

	// compare the generated schema to the expected one
	return actual.Equals(expected)
}
