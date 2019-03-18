package persist

import (
	"time"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/schema/drdl"
)

type schema struct {
	ID      bson.ObjectId `bson:"_id"`
	Created time.Time     `bson:"created"`
	DRDL    *drdl.Schema  `bson:"schema"`
}

func newSchema(ds *drdl.Schema) schema {
	return schema{
		ID:      bson.NewObjectId(),
		Created: time.Now(),
		DRDL:    ds,
	}
}
