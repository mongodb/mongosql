package testutil

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

type TestCase struct {
	Name     string
	Input    bsoncore.Document
	Expected bsoncore.Document
}

func LoadTestCases(fileName string) []TestCase {
	path := filepath.Join("..", "testdata", fileName)
	docs := loadJSONFromFile(path)

	testCases := make([]TestCase, len(docs))
	for i, doc := range docs {
		testCases[i].Name = doc.Lookup("name").StringValue()
		testCases[i].Input = doc.Lookup("input").Array()
		testCases[i].Expected = doc.Lookup("expected").Array()
	}

	return testCases
}

func loadJSONFromFile(path string) []bsoncore.Document {
	r, err := os.Open(path)
	if err != nil {
		panic(errors.Wrapf(err, "failed to open %s", path))
	}
	defer r.Close()

	jr, err := bsonrw.NewExtJSONValueReader(r, false)
	if err != nil {
		panic(errors.Wrapf(err, "failed to read JSON from %s", path))
	}

	ar, err := jr.ReadArray()
	if err != nil {
		panic(errors.Wrapf(err, "failed reading top-level JSON array from %s", path))
	}

	var result []bsoncore.Document
	c := bsonrw.NewCopier()
	for {
		evr, err := ar.ReadValue()
		if err != nil {
			if err == bsonrw.ErrEOA {
				return result
			}
			panic(errors.Wrapf(err, "failed reading JSON document from %s", path))
		}

		if evr.Type() != bsontype.EmbeddedDocument {
			panic(fmt.Sprintf("unexpected data type reading JSON from %s", path))
		}

		doc, err := c.CopyDocumentToBytes(evr)
		if err != nil {
			panic(errors.Wrapf(err, "failed copying document to bytes reading JSON from %s", path))
		}

		result = append(result, doc)
	}
}
