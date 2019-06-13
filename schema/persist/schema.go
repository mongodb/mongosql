package persist

import (
	"time"

	"github.com/10gen/sqlproxy/schema/drdl"

	oldbson "github.com/10gen/mongo-go-driver/bson"
)

type schema struct {
	ID      oldbson.ObjectId `bson:"_id"`
	Created time.Time        `bson:"created"`
	DRDL    *drdl.Schema     `bson:"schema"`
}

func newSchema(ds *drdl.Schema) schema {
	return schema{
		ID:      oldbson.NewObjectId(),
		Created: time.Now(),
		DRDL:    ds,
	}
}
