package sample

import (
	"time"

	"github.com/10gen/sqlproxy/schema/mongo"

	"github.com/10gen/mongo-go-driver/bson"
)

// Namespace holds information for a sampled schema namespace.
type Namespace struct {
	id bson.ObjectId `bson:"_id"`

	// an ObjectId referencing the _id of a schema
	// version document from the versions collection.
	VersionId bson.ObjectId `bson:"versionId,omitempty"`

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

func NewNamespace(db, c string, id bson.ObjectId) *Namespace {
	return &Namespace{
		VersionId:  id,
		Database:   db,
		Collection: c,
	}
}

func (n *Namespace) Equals(ns *Namespace) bool {
	return n.Database == ns.Database &&
		n.Collection == ns.Collection &&
		n.VersionId == ns.VersionId
}
