//go:build spectests

package mongosql

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/10gen/candiedyaml"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

const queryTestsDir = "../../tests/spec_tests/query_tests"
const translationTestsDir = "../../tests/translation_tests"

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
		Translation     []bson.D `yaml:"translation"`
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

// compareOrderedSets checks if the given sets are equal, and if their
// documents have the same ordering.
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

func TestSpecTranslations(t *testing.T) {
	cwd, err := os.Getwd()
	assert.NoError(t, err)

	type testDir = struct {
		name string
		path string
	}
	testDirs := []testDir{
		{name: "query", path: queryTestsDir},
		{name: "translation", path: translationTestsDir},
	}

	for _, dir := range testDirs {
		t.Run(dir.name, func(t *testing.T) {
			path := cwd + "/" + dir.path
			tests, err := ioutil.ReadDir(path)
			if err != nil {
				t.Fatalf("directory does not exist: '%s'", path)
			}

			for _, testFile := range tests {
				if !strings.HasSuffix(testFile.Name(), ".yml") {
					continue
				}
				t.Run(testFile.Name(), func(t *testing.T) {
					yaml, err := readYamlFile(path + "/" + testFile.Name())
					assert.NoError(t, err)

					for _, testCase := range yaml.Tests {
						t.Run(testCase.Description, func(t *testing.T) {
							if testCase.SkipReason != "" {
								t.Skip(testCase.SkipReason)
							}

							catalogSchema, err := toCoreDocument(yaml.CatalogSchema)
							assert.NoError(t, err)

							args := TranslationArgs{
								DB:             testCase.CurrentDb,
								SQL:            testCase.Query,
								CatalogSchema:  catalogSchema,
								skipDesugaring: true,
							}

							translation, err := Translate(args)

							if testCase.AlgebrizeError != "" {
								if err == nil {
									t.Fatal("expected an algebrize error, but algebrization succeeded")
								}
								assert.Contains(t, err.Error(), "algebrize error")
							} else if testCase.ParseError != "" {
								if err == nil {
									t.Fatal("expected a parse error")
								}
								assert.Contains(t, err.Error(), "parse error")
							} else {
								if err == nil {
									var actualPipeline []bson.D
									val := bson.RawValue{
										Type:  bsontype.Array,
										Value: translation.Pipeline,
									}
									// Candiedyaml's unmarshaller uses int64, while the BSON unmarshaller uses int32.
									// This registry forces the BSON unmarshaller to use int64.
									registry := bson.NewRegistryBuilder().
										RegisterTypeMapEntry(bsontype.Int32, reflect.TypeOf(int64(0))).
										Build()
									err := val.UnmarshalWithRegistry(registry, &actualPipeline)

									// We have to marshal and then unmarshal the expected pipeline
									// because candiedyaml will convert an array that's inside a bson.D
									// into an []interface{}, instead of a bson.A{}.
									var expectedPipeline []bson.D
									_, expectedBytes, err := bson.MarshalValue(testCase.Translation)
									assert.NoError(t, err)
									v := bson.RawValue{
										Type:  bsontype.Array,
										Value: expectedBytes,
									}
									err = v.Unmarshal(&expectedPipeline)
									assert.NoError(t, err)
									if !reflect.DeepEqual(expectedPipeline, actualPipeline) {
										t.Fatalf("expected translations to be equal, but they weren't:\nexpected:%v\nactual:%v\n", expectedPipeline, actualPipeline)
									}
								} else {
									t.Fatalf("translation unexpectedly failed: %v", err.Error())
								}
							}
						})
					}
				})
			}
		})
	}
}

func TestSpecResultSets(t *testing.T) {
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
		t.Run(testFile.Name(), func(t *testing.T) {
			yaml, err := readYamlFile(queryTestsPath + "/" + testFile.Name())
			assert.NoError(t, err)

			assert.NoError(t, loadCatalogData(client, yaml.CatalogData))

			for _, testCase := range yaml.Tests {
				t.Run(testCase.Description, func(t *testing.T) {
					if testCase.SkipReason != "" {
						t.Skip(testCase.SkipReason)
					}

					catalogSchema, err := toCoreDocument(yaml.CatalogSchema)
					assert.NoError(t, err)

					args := TranslationArgs{
						DB:            testCase.CurrentDb,
						SQL:           testCase.Query,
						CatalogSchema: catalogSchema,
					}

					translation, err := Translate(args)

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
							var actualCursor *mongo.Cursor
							if translation.TargetCollection == "" {
								actualCursor, err = client.Database(testCase.TranslationDB).
									Aggregate(context.Background(), pipelineBsonD)
							} else {
								actualCursor, err = client.Database(testCase.TranslationDB).
									Collection(testCase.TranslationColl).
									Aggregate(context.Background(), pipelineBsonD)
							}
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
		})
	}
}
