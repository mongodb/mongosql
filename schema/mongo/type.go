package mongo

// BsonType is an enum representing all the bson types suported by Schema.
type BsonType string

// Possible values of BsonType
// for reference: https://docs.mongodb.com/manual/reference/operator/query/type/#available-types
const (
	Int        BsonType = "int"
	Long       BsonType = "long"
	Double     BsonType = "double"
	Decimal    BsonType = "decimal"
	Boolean    BsonType = "bool"
	String     BsonType = "string"
	Object     BsonType = "object"
	Array      BsonType = "array"
	Date       BsonType = "date"
	BinData    BsonType = "binData"
	ObjectId   BsonType = "objectId"
	NoBsonType BsonType = ""
)

// Less returns true if b should come before other when ordering BsonTypes.
// Less is equivalent to a lexicographical comparison of two BsonTypes' string
// representations.
func (b BsonType) Less(other BsonType) bool {
	return b < other
}

// SpecialType is an enum representing the possible values of the "specialType"
// field as descibed in the schema management design document. "specialType" is
// a sqlproxy-specific extension to JSON Schema that allows us to indicate that
// we want json-to-relation translation behavior different from the default for
// this schema's bsonType. For more information, see the design doc here:
// https://docs.google.com/document/d/12LWz00vJo_H-tHFv7IHa5L6X5a6Y-eNE4dyb8TXdk7U/edit
type SpecialType string

// Possible values of SpecialType
const (
	GeoPoint      SpecialType = "geoPoint"
	UUID3         SpecialType = "uuid3"
	UUID4         SpecialType = "uuid4"
	NoSpecialType SpecialType = ""
)
