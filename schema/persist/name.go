package persist

import (
	oldbson "github.com/10gen/mongo-go-driver/bson"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Name represents a stored schema name. The ID field is the display name, and
// the SchemaID field is a reference to a document in the schemasCollection.
type Name struct {
	ID       string             `bson:"_id"`
	SchemaID primitive.ObjectID `bson:"schema_id"`
}

type oldName struct {
	ID       string           `bson:"_id"`
	SchemaID oldbson.ObjectId `bson:"schema_id"`
}

func newName(id string, schemaID primitive.ObjectID) Name {
	return Name{
		ID:       id,
		SchemaID: schemaID,
	}
}

func (n *oldName) toNew() *Name {
	oid, err := primitive.ObjectIDFromHex(n.SchemaID.Hex())
	if err != nil {
		panic(err)
	}

	return &Name{
		ID:       n.ID,
		SchemaID: oid,
	}
}

func (n *Name) toOld() *oldName {
	return &oldName{
		ID:       n.ID,
		SchemaID: oldbson.ObjectIdHex(n.SchemaID.Hex()),
	}
}
