package mongosql

import (
	"encoding/base64"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// Version returns the version of the underlying c translation
// library. The consumer of this library should ensure that the
// version of the go library matches that of the c library.
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
	// relaxSchemaChecking relaxes schema checking for comparisons if it's
	// set to true. This means that schema checking will pass unless a type
	// constraint has been violated.
	relaxSchemaChecking bool
	// ExcludeNamespaces when set to true will return a non-namespaced result set
	ExcludeNamespaces bool
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
}

// TranslationError is an error type that includes additional
// information about whether an error is "internal" or "external"
// (i.e. whether it is safe and useful to expose to end users).
type TranslationError struct {
	internal bool
	err      error
}

// NewInternalError creates a TranslationError from the provided error
// that should not be exposed to end users.
func NewInternalError(err error) TranslationError {
	return TranslationError{internal: true, err: err}
}

// NewExternalError creates a TranslationError from the provided error
// that is okay/useful to expose to end users.
func NewExternalError(err error) TranslationError {
	return TranslationError{internal: false, err: err}
}

// Error implements the error interface by returning the string
// representation of the underlying error.
func (e TranslationError) Error() string {
	return e.err.Error()
}

// IsInternal returns true if this is an "internal" error that should
// not be exposed to end users, and false otherwise.
func (e TranslationError) IsInternal() bool {
	return e.internal
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
		ErrorIsInternal bool              `bson:"error_is_internal"`
		ResultSetSchema bsoncore.Document `bson:"result_set_schema"`
		Namespaces      []Namespace       `bson:"namespaces"`
	}{}

	translationBytes, err := base64.StdEncoding.DecodeString(base64TranslationResult)
	if err != nil {
		return Translation{}, NewInternalError(fmt.Errorf("failed to decode base64 translation result: %w", err))
	}

	err = bson.Unmarshal(translationBytes, &translationResult)
	if err != nil {
		return Translation{}, NewInternalError(fmt.Errorf("failed to unmarshal translation result BSON into struct: %w", err))
	}

	if translationResult.Error != "" {
		if translationResult.ErrorIsInternal {
			err = NewInternalError(errors.New(translationResult.Error))
		} else {
			err = NewExternalError(errors.New(translationResult.Error))
		}
		return Translation{}, err
	}

	typ, pipelineBytes, err := bson.MarshalValue(translationResult.Pipeline)
	if err != nil {
		return Translation{}, NewInternalError(fmt.Errorf("failed to marshal pipeline to bytes: %w", err))
	}
	if typ.String() != "array" {
		// this should never occur, but is here as a sanity check
		panic("didn't marshal to array")
	}

	return Translation{
		TargetDB:         translationResult.DB,
		TargetCollection: translationResult.Collection,
		Pipeline:         pipelineBytes,
		ResultSetSchema:  translationResult.ResultSetSchema,
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
	base64Result := callGetNamespaces(dbName, sqlStatement)

	result := struct {
		Namespaces      []Namespace `bson:"namespaces"`
		Error           string      `bson:"error"`
		ErrorIsInternal bool        `bson:"error_is_internal"`
	}{}

	resultBytes, err := base64.StdEncoding.DecodeString(base64Result)
	if err != nil {
		return nil, NewInternalError(fmt.Errorf("failed to decode base64 translation result: %w", err))
	}

	err = bson.Unmarshal(resultBytes, &result)
	if err != nil {
		return nil, NewInternalError(fmt.Errorf("failed to unmarshal translation result BSON into struct: %w", err))
	}

	if result.Error != "" {
		if result.ErrorIsInternal {
			err = NewInternalError(errors.New(result.Error))
		} else {
			err = NewExternalError(errors.New(result.Error))
		}
		return nil, err
	}

	return result.Namespaces, nil
}
