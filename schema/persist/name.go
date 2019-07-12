package persist

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Name represents a stored schema name. The ID field is the display name, and
// the SchemaID field is a reference to a document in the schemasCollection.
type Name struct {
	ID       string             `bson:"_id"`
	SchemaID primitive.ObjectID `bson:"schema_id"`
}

func newName(id string, schemaID primitive.ObjectID) Name {
	return Name{
		ID:       id,
		SchemaID: schemaID,
	}
}
