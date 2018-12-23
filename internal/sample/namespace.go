package sample

import (
	"fmt"
	"time"

	"github.com/10gen/sqlproxy/internal/schema/mongo"

	"github.com/10gen/mongo-go-driver/bson"
)

// Namespace holds information for a sampled schema namespace.
type Namespace struct {
	id bson.ObjectId `bson:"_id"`

	// an ObjectId referencing the _id of a schema
	// version document from the versions collection.
	VersionID bson.ObjectId `bson:"versionId,omitempty"`

	// a string representing the name of the database.
	Database string `bson:"database,omitempty"`

	// a string representing the name of the collection.
	Collection string `bson:"collection,omitempty"`

	// an int64 representing the number of sampled documents.
	SampleSize int64 `bson:"sampleSize,omitempty"`

	// a time.Time representing the time at which sampling commenced.
	StartSampleTime time.Time `bson:"startSampleTime,omitempty"`

	// an time.Time representing the time at which sampling completed.
	EndSampleTime time.Time `bson:"endSampleTime,omitempty"`

	// JSON schema for this namespace.
	Schema *mongo.Schema `bson:"schema,omitempty"`
}

// NewNamespace returns a new Namespace using the database,
// collection and id supplied.
func NewNamespace(db, c string, id bson.ObjectId) *Namespace {
	return &Namespace{
		VersionID:  id,
		Database:   db,
		Collection: c,
	}
}

// NewNamespaceWithoutID returns a new Namespace using the database and collection supplied.
func NewNamespaceWithoutID(database, collection string) *Namespace {
	return &Namespace{
		Database:   database,
		Collection: collection,
	}
}

// Equals returns true if n is the same as ns and
// false otherwise.
func (n *Namespace) Equals(ns *Namespace) bool {
	return n.Database == ns.Database &&
		n.Collection == ns.Collection &&
		n.VersionID == ns.VersionID
}

// QuotedString returns a quoted version of the namespace.
func (n *Namespace) QuotedString() string {
	return fmt.Sprintf("%q.%q", n.Database, n.Collection)
}

// String returns a stringified version of the namespace.
func (n *Namespace) String() string {
	return fmt.Sprintf("%s.%s", n.Database, n.Collection)
}
