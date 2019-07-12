package persist

import (
	"time"

	"github.com/10gen/sqlproxy/schema/drdl"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type schema struct {
	ID      primitive.ObjectID `bson:"_id"`
	Created time.Time          `bson:"created"`
	DRDL    *drdl.Schema       `bson:"schema"`
}

func newSchema(ds *drdl.Schema) schema {
	return schema{
		ID:      primitive.NewObjectID(),
		Created: time.Now(),
		DRDL:    ds,
	}
}
