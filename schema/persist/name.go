package persist

import "github.com/10gen/mongo-go-driver/bson"

// Name represents a stored schema name. The ID field is the display name, and
// the SchemaID field is a reference to a document in the schemasCollection.
type Name struct {
	ID       string        `bson:"_id"`
	SchemaID bson.ObjectId `bson:"schema_id"`
}

func newName(id string, schemaID bson.ObjectId) Name {
	return Name{
		ID:       id,
		SchemaID: schemaID,
	}
}
