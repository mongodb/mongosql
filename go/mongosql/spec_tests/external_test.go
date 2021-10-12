//go:build spectests

package spec_tests

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/10gen/candiedyaml"
	"github.com/10gen/mongosql-rs/go/mongosql"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

const queryTestsDir = "../../../tests/spec_tests/query_tests"

type YamlTest struct {
	// Tests holds testing information for each test in a given YAML file.
	Tests []struct {
		Description     string   `yaml:"description"`
		CurrentDb       string   `yaml:"current_db"`
		TranslationDB   string   `yaml:"translation_target_db"`
		TranslationColl string   `yaml:"translation_target_coll"`
		Query           string   `yaml:"query"`
		AlgebrizeError  string   `yaml:"algebrize_error"`
		ParseError      string   `yaml:"parse_error"`
		SkipReason      string   `yaml:"skip_reason"`
		Ordered         bool     `yaml:"ordered"`
		Result          []bson.D `yaml:"result"`
	} `yaml:"tests"`
	// CatalogData maps namespaces to slices containing collection documents.
	CatalogData map[string]map[string][]bson.D `yaml:"catalog_data"`
	// CatalogSchema maps namespaces to JSON schema validators.
	CatalogSchema map[string]map[string]bson.D `yaml:"catalog_schema"`
}

// readYamlFile tries to marshal the given YAML file into a YamlTest struct.
func readYamlFile(file string) (*YamlTest, error) {
	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("error reading YAML file: '%s'", file)
	}
	var yamlTest YamlTest
	err = candiedyaml.Unmarshal(yamlFile, &yamlTest)
	if err != nil {
		return nil, fmt.Errorf("error parsing YAML file: %s\n", err)
	}
	return &yamlTest, nil
}

// setUpMongoDB connects to a local MongoDB instance and returns the client.
func setUpMongoDB() (*mongo.Client, error) {
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	// Check the connection
	err = client.Ping(context.Background(), nil)

	return client, err
}

// normalizeBSON converts bson.Ds that represent extended JSON types into
// primitive bson types and converts []interface{} into bson.A. This is
// necessary because the candiedyaml parser does not decode extended JSON values.
func normalizeBSON(docs []bson.D) ([]bson.D, error) {
	newDocs := make([]bson.D, len(docs))

	for i, doc := range docs {
		oldDocBytes, err := bson.MarshalExtJSON(doc, false, false)
		if err != nil {
			return nil, fmt.Errorf("error marshaling input doc %v: %v", i, err)
		}

		newDoc := &bson.D{}
		err = bson.UnmarshalExtJSON(oldDocBytes, false, newDoc)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling input doc %v:\n\tdoc: %v\n\terr: %v", i, string(oldDocBytes), err)
		}

		newDocs[i] = *newDoc
	}

	return newDocs, nil
}

// loadCatalogData loads catalog data into a MongoDB instance via the provided client.
func loadCatalogData(client *mongo.Client, catalogData map[string]map[string][]bson.D) error {
	err := deleteCatalogData(client, catalogData)
	if err != nil {
		return err
	}

	for db, collData := range catalogData {
		clientDB := client.Database(db)
		for coll, documents := range collData {
			clientColl := clientDB.Collection(coll)
			// Normalize each document in case it uses extendedJSON notation.
			normalizedDocs, err := normalizeBSON(documents)
			if err != nil {
				return err
			}
			for _, doc := range normalizedDocs {
				_, err := clientColl.InsertOne(context.Background(), doc)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// deleteCatalogData drops all data from the provided client.
func deleteCatalogData(client *mongo.Client, catalogData map[string]map[string][]bson.D) error {
	for db := range catalogData {
		err := client.Database(db).Drop(context.Background())
		if err != nil {
			return err
		}
	}
	return nil
}

// less returns if the byte slice at index i is less than the byte slice at
// index j.
func less(slice [][]byte, i, j int) bool {
	switch bytes.Compare(slice[i], slice[j]) {
	case -1:
		return true
	default:
		return false
	}
}

// compareUnorderedSets checks if the given sets are equal.
func compareUnorderedSets(expected, actual []bson.D) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("result sets have different sizes: expected %v, found %v", len(expected), len(actual))
	}
	var expectedSlice, actualSlice [][]byte
	for i := 0; i < len(expected); i++ {
		expectedBytes, err := bson.Marshal(expected[i])
		if err != nil {
			return fmt.Errorf("failed to marshal document from expected result set")
		}
		expectedSlice = append(expectedSlice, expectedBytes)

		actualBytes, err := bson.Marshal(actual[i])
		if err != nil {
			return fmt.Errorf("failed to marshal document from actual result set")
		}
		actualSlice = append(actualSlice, actualBytes)
	}
	sort.SliceStable(expectedSlice, func(i, j int) bool {
		return less(expectedSlice, i, j)
	})
	sort.SliceStable(actualSlice, func(i, j int) bool {
		return less(actualSlice, i, j)
	})
	if !reflect.DeepEqual(expectedSlice, actualSlice) {
		return fmt.Errorf("expected unordered result sets to be equal, but they weren't:\nexpected:%v\nactual:%v\n", expected, actual)
	}
	return nil
}

func compareOrderedSets(expected, actual []bson.D) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("result sets have different sizes: expected %v, found %v", len(expected), len(actual))
	}
	for i := 0; i < len(expected); i++ {
		if !reflect.DeepEqual(expected[i], actual[i]) {
			return fmt.Errorf("expected ordered result sets to be equal, but they weren't:\nexpected:%v\nactual:%v\n", expected, actual)
		}
	}
	return nil
}

// toCoreDocument converts a catalog schema from a map of namespaces to bson.Ds to a map
// of namespaces to bsoncore.Documents.
func toCoreDocument(catalogSchema map[string]map[string]bson.D) (map[string]map[string]bsoncore.Document, error) {
	out := make(map[string]map[string]bsoncore.Document, len(catalogSchema))

	for db, collSchema := range catalogSchema {
		out[db] = make(map[string]bsoncore.Document, len(collSchema))
		for coll, schema := range collSchema {
			doc, err := bson.Marshal(&schema)
			if err != nil {
				return nil, err
			}
			out[db][coll] = doc
		}
	}

	return out, nil
}

func TestSpecExecution(t *testing.T) {
	cwd, err := os.Getwd()
	assert.NoError(t, err)

	queryTestsPath := cwd + "/" + queryTestsDir
	queryTests, err := ioutil.ReadDir(queryTestsPath)
	if err != nil {
		t.Fatalf("directory does not exist: '%s'", queryTestsPath)
	}

	client, err := setUpMongoDB()
	assert.NoError(t, err)

	for _, testFile := range queryTests {
		yaml, err := readYamlFile(queryTestsPath + "/" + testFile.Name())
		assert.NoError(t, err)

		assert.NoError(t, loadCatalogData(client, yaml.CatalogData))

		for _, testCase := range yaml.Tests {
			if testCase.SkipReason != "" {
				continue
			}
			catalogSchema, err := toCoreDocument(yaml.CatalogSchema)
			assert.NoError(t, err)

			t.Run(fmt.Sprintf("%s", testCase.Description), func(t *testing.T) {
				args := mongosql.TranslationArgs{
					DB:            testCase.CurrentDb,
					SQL:           testCase.Query,
					CatalogSchema: catalogSchema,
				}

				translation, err := mongosql.Translate(args)

				if testCase.AlgebrizeError != "" {
					if err == nil {
						t.Fatal("expected an algebrize error")
					}
					assert.Contains(t, err.Error(), "algebrize error")
				} else if testCase.ParseError != "" {
					if err == nil {
						t.Fatal("expected a parse error")
					}
					assert.Contains(t, err.Error(), "parse error")
				} else {
					if err == nil {
						var pipelineBsonD []bson.D
						val := bson.RawValue{
							Type:  bsontype.Array,
							Value: translation.Pipeline,
						}
						err := val.Unmarshal(&pipelineBsonD)
						assert.NoError(t, err)

						var actualResultSet []bson.D
						actualCursor, err := client.Database(testCase.TranslationDB).Collection(testCase.TranslationColl).Aggregate(context.Background(), pipelineBsonD)
						assert.NoError(t, err)
						err = actualCursor.All(context.Background(), &actualResultSet)
						assert.NoError(t, err)

						// Normalize the expected result set in case it uses extendedJSON notation.
						normalizedExpectedResultSet, err := normalizeBSON(testCase.Result)
						assert.NoError(t, err)

						if testCase.Ordered {
							assert.NoError(t, compareOrderedSets(normalizedExpectedResultSet, actualResultSet))
						} else {
							assert.NoError(t, compareUnorderedSets(normalizedExpectedResultSet, actualResultSet))
						}
					} else {
						t.Fatalf("translation unexpectedly failed: %v", err.Error())
					}
				}
			})
		}
		assert.NoError(t, deleteCatalogData(client, yaml.CatalogData))
	}
}
