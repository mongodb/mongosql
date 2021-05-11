package desugarer

import (
	"fmt"
	"os"
	"path/filepath"

	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

type TestCase struct {
	Name     string
	Skip     *string
	Input    bsoncore.Array
	Expected bsoncore.Array
}

func LoadTestCases(fileName string) []TestCase {
	path := filepath.Join(".", "testdata", fileName)
	docs := loadJSONFromFile(path)

	testCases := make([]TestCase, len(docs))
	for i, doc := range docs {
		testCases[i].Name = doc.Lookup("name").StringValue()
		skip, ok := doc.Lookup("skip").StringValueOK()
		if ok {
			testCases[i].Skip = &skip
		} else {
			testCases[i].Skip = nil
		}
		// GODRIVER-1930 (both lines).
		testCases[i].Input = bsoncore.Array(doc.Lookup("input").Array())
		testCases[i].Expected = bsoncore.Array(doc.Lookup("expected").Array())
	}

	return testCases
}

func loadJSONFromFile(path string) []bsoncore.Document {
	r, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("failed to open %s: %w", path, err))
	}
	defer r.Close()

	jr, err := bsonrw.NewExtJSONValueReader(r, false)
	if err != nil {
		panic(fmt.Errorf("failed to read JSON from %s: %w", path, err))
	}

	ar, err := jr.ReadArray()
	if err != nil {
		panic(fmt.Errorf("failed reading top-level JSON array from %s: %w", path, err))
	}

	var result []bsoncore.Document
	c := bsonrw.NewCopier()
	for {
		evr, err := ar.ReadValue()
		if err != nil {
			if err == bsonrw.ErrEOA {
				return result
			}
			panic(fmt.Errorf("failed reading JSON document from %s: %w", path, err))
		}

		if evr.Type() != bsontype.EmbeddedDocument {
			panic(fmt.Sprintf("unexpected data type reading JSON from %s", path))
		}

		doc, err := c.CopyDocumentToBytes(evr)
		if err != nil {
			panic(fmt.Errorf("failed copying document to bytes reading JSON from %s: %w", path, err))
		}

		result = append(result, doc)
	}
}
