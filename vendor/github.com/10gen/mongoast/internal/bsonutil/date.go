package bsonutil

import (
	"time"

	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

const ISODateFormat = "2006-01-02T15:04:05.999Z07:00"

// ISODateString returns the formatted date according to how MongoDB prints
// dates using ISODate. It panics if the specified value is not a DateTime.
func ISODateString(v bsoncore.Value) string {
	return time.Unix(0, v.DateTime()*time.Millisecond.Nanoseconds()).UTC().Format(ISODateFormat)
}
