package mongosql

import (
	"encoding/base64"
	"fmt"

	"github.com/10gen/mongosql-rs/go/internal/desugarer"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// Version returns the version of the underlying c translation
// library. The consumer of this library should ensure that the
// version of the the go library matches that of the c library.
func Version() string {
	return version()
}

// TranslationArgs contains the arguments to the translation engine.
type TranslationArgs struct {
	// DB represents the current database in which the sql query was run
	DB string
	// SQL is a string containing the sql query
	SQL string
	// CatalogSchema maps namespaces to JSON Schemas that describe the
	// shape of the documents in the namespace.
	CatalogSchema map[string]map[string]bsoncore.Document
	// skipDesugaring skips desugaring the translation pipeline if it's set to true
	skipDesugaring bool
}

// Translation represents the result of translating a sql query to
// MQL. The fields of this struct can be used to construct an
// aggregate command equivalent to a SQL query.
type Translation struct {
	// TargetDB is the MongoDB database against which the aggregate
	// command should run
	TargetDB string
	// TargetCollection is the collection to be specified in the
	// aggregate command, or nil if an "aggregate: 1" style aggregate
	// command should be used
	TargetCollection string
	// Pipeline is the array representing the aggregation pipeline
	// serialized to BSON
	Pipeline []byte
	// ResultSetSchema is a JSON Schema document that describes
	// the documents returned by this Translation
	ResultSetSchema bsoncore.Document

	// namespaces is the list of namespaces referenced by the
	// SQL query from which this Translation is built
	namespaces []Namespace
}

// Translate accepts TranslationArgs, returning a Translation and an
// error if the translation failed. If the returned error is non-nil,
// the returned Translation should be disregarded.
func Translate(args TranslationArgs) (Translation, error) {
	base64TranslationResult, err := callTranslate(args)
	if err != nil {
		return Translation{}, err
	}

	translationResult := struct {
		DB              string            `bson:"target_db"`
		Collection      string            `bson:"target_collection"`
		Pipeline        []bson.D          `bson:"pipeline"`
		Error           string            `bson:"error"`
		ResultSetSchema bsoncore.Document `bson:"result_set_schema"`
		Namespaces      []Namespace       `bson:"namespaces"`
	}{}

	translationBytes, err := base64.StdEncoding.DecodeString(base64TranslationResult)
	if err != nil {
		panic(err)
	}

	err = bson.Unmarshal(translationBytes, &translationResult)
	if err != nil {
		return Translation{}, fmt.Errorf("failed to unmarshal translation result BSON into struct: %w", err)
	}

	if translationResult.Error != "" {
		return Translation{}, fmt.Errorf(translationResult.Error)
	}

	typ, pipelineBytes, err := bson.MarshalValue(translationResult.Pipeline)
	if err != nil {
		return Translation{}, fmt.Errorf("failed to marshal pipeline to bytes: %w", err)
	}
	if typ.String() != "array" {
		// this should never occur, but is here as a sanity check
		panic("didn't marshal to array")
	}

	if !args.skipDesugaring {
		pipelineBytes, err = desugarer.Desugar(pipelineBytes, translationResult.DB)
		if err != nil {
			return Translation{}, fmt.Errorf("failed to desugar pipeline: %w", err)
		}
	}

	return Translation{
		TargetDB:         translationResult.DB,
		TargetCollection: translationResult.Collection,
		Pipeline:         pipelineBytes,
		ResultSetSchema:  translationResult.ResultSetSchema,
		namespaces:       translationResult.Namespaces,
	}, nil
}

// Namespace represents a MongoDB collection namespace.
type Namespace struct {
	// Database is the database component of the namespace
	Database string `bson:"database"`
	// Collection is the collection component of the namespace
	Collection string `bson:"collection"`
}

// GetNamespaces returns the Namespaces referenced in the provided
// sqlStatement. Unqualified collections in the statement are assumed
// to be in the provided database.
func GetNamespaces(dbName, sqlStatement string) ([]Namespace, error) {
	translation, err := Translate(TranslationArgs{
		DB:            dbName,
		SQL:           sqlStatement,
		CatalogSchema: nil,
	})
	if err != nil {
		return nil, err
	}

	return translation.namespaces, nil
}
