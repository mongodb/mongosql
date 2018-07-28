package mapping_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/drdl"
	"github.com/10gen/sqlproxy/schema/mapping"
	"github.com/10gen/sqlproxy/schema/mongo"
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

		case "schema.yml":
			// test basic json-to-relation schema mapping
			name := collection + "-map"
			t.Run(name, func(t *testing.T) {
				err := testMapSchemaFromJson(collection, false)
				if err != nil {
					t.Fatal(err.Error())
				}
			})

		case "prejoined.yml":
			// test json-to-relation schema mapping with prejoins
			name := collection + "-map-prejoined"
			t.Run(name, func(t *testing.T) {
				err := testMapSchemaFromJson(collection, true)
				if err != nil {
					t.Fatal(err.Error())
				}
			})

		case "sample.json":
			// test json schema creation from sample doc
			name := collection + "-sample"
			t.Run(name, func(t *testing.T) {
				err := testMapSchemaFromSample(collection)
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

func testMapSchemaFromJson(collection string, prejoined bool) error {
	dir := "testdata/" + collection + "/"

	expectedFile := dir + "schema.yml"
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
	return testMapSchema(collection, prejoined, jsonSchema, expected)
}

func testMapSchemaFromSample(collection string) error {
	dir := "testdata/" + collection + "/"

	expectedFile := dir + "schema.yml"
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
	sample, err := bsonFromJson(sampleBytes)
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
	return testMapSchema(collection, false, actual, expected)
}

//func NewSchemaMappingConfig(
//	database *schema.Database,
//	schema *mongo.Schema,
//	collectionName string,
//	preJoin bool,
//	uuidSubtype3Encoding string,
//	version []uint8,
//	logger log.Logger,
//	heuristic config.MappingHeuristic,
func testMapSchema(col string, prejoined bool, js *mongo.Schema,
	expected *schema.Schema) error {

	// create a test database schema
	db := schema.NewDatabase(log.GlobalLogger(), "test", nil)

	// map the json schema into the database
	err := mapping.Map(mapping.NewSchemaMappingConfig(db, js, col, prejoined,
		"old", []uint8{4, 0, 0}, log.GlobalLogger(), config.MajorityMappingMode))

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

func bsonFromJson(jsonBytes []byte) (bson.D, error) {
	dict := map[string]interface{}{}
	err := json.Unmarshal(jsonBytes, &dict)
	if err != nil {
		return nil, err
	}

	doc := bsonFromMap(dict)
	return doc, nil
}

func toBSON(val interface{}) interface{} {
	switch typedV := val.(type) {
	case map[string]interface{}:
		return bsonFromMap(typedV)
	case []interface{}:
		arr := []interface{}{}
		for _, elem := range typedV {
			arr = append(arr, toBSON(elem))
		}
		return arr
	}
	return val
}

func bsonFromMap(dict map[string]interface{}) bson.D {
	var doc bson.D
	for key, val := range dict {
		doc = append(doc, bson.DocElem{Name: key, Value: toBSON(val)})
	}
	return doc
}
