package mongo

// BSONType is an enum representing all the bson types suported by Schema.
type BSONType string

// Possible values of BSONType
// for reference: https://docs.mongodb.com/manual/reference/operator/query/type/#available-types
const (
	Int        BSONType = "int"
	Long       BSONType = "long"
	Double     BSONType = "double"
	Decimal    BSONType = "decimal"
	Boolean    BSONType = "bool"
	String     BSONType = "string"
	Object     BSONType = "object"
	Array      BSONType = "array"
	Date       BSONType = "date"
	BinData    BSONType = "binData"
	ObjectID   BSONType = "objectId"
	NoBSONType BSONType = ""
)

// Less returns true if b should come before other when ordering BSONTypes.
// Less is equivalent to a lexicographical comparison of two BSONTypes' string
// representations.
func (b BSONType) Less(other BSONType) bool {
	return b < other
}

// SpecialType is an enum representing the possible values of the "specialType"
// field as described in the schema management design document. "specialType" is
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

// bsonTypeLatticeLeastUpperBound map implements the least upper bound (LeastUpperBound)
// of the type lattice presented
// in:
// https://docs.google.com/document/d/1FCsQ9ecDhQfamjvcgvfuaCNcW-RHAFNUdBTZpQWns_c/edit#
// Notes:
//    1. A 64bit integer cannot be represented faitfully using
//       a double, so the LeastUpperBound of double and long must be decimal.
//    2. The only BinData we support is UUID, so BinData is under String
//       in the lattice since UUIDs can be represented safely as Strings.

var bsonTypeLatticeDeclaration = [][][]BSONType{
	{
		{Int, Long, Long},
		{Int, Double, Double},
		{Int, Decimal, Decimal},
		{Int, Boolean, Int},
		{Int, String, String},
		{Int, Date, String},
		{Int, BinData, String},
		{Int, ObjectID, String},
		{Int, NoBSONType, Int},
	},
	{
		{Long, Double, Decimal},
		{Long, Decimal, Decimal},
		{Long, Boolean, Long},
		{Long, String, String},
		{Long, Date, String},
		{Long, BinData, String},
		{Long, ObjectID, String},
		{Long, NoBSONType, Long},
	},
	{
		{Double, Decimal, Decimal},
		{Double, Boolean, Double},
		{Double, String, String},
		{Double, Date, String},
		{Double, BinData, String},
		{Double, ObjectID, String},
		{Double, NoBSONType, Double},
	},
	{
		{Decimal, Boolean, Decimal},
		{Decimal, String, String},
		{Decimal, Date, String},
		{Decimal, BinData, String},
		{Decimal, ObjectID, String},
		{Decimal, NoBSONType, Decimal},
	},
	{
		{Boolean, String, String},
		{Boolean, Date, String},
		{Boolean, BinData, String},
		{Boolean, ObjectID, String},
		{Boolean, NoBSONType, Boolean},
	},
	{
		{String, Date, String},
		{String, BinData, String},
		{String, ObjectID, String},
		{String, NoBSONType, String},
	},
	{
		{Date, BinData, String},
		{Date, ObjectID, String},
		{Date, NoBSONType, Date},
	},
	{
		{BinData, ObjectID, String},
		{BinData, NoBSONType, BinData},
	},
	{
		{ObjectID, NoBSONType, ObjectID},
	},
	{
		{NoBSONType, NoBSONType, NoBSONType},
	},
}

// bsonTypeLatticeLeastUpperBound is a map representing
// the least upper bound of any two types in the BsonTypeLattice.
var bsonTypeLatticeLeastUpperBound = initBSONTypeLatticeLeastUpperBound()

// initBSONTypeLatticeLeastUpperBound initializes the BSONTypeLatticeLeastUpperBound
// maps based on the cut down bsonTypeLatticeDeclaration.
func initBSONTypeLatticeLeastUpperBound() map[BSONType]map[BSONType]BSONType {
	ret := make(map[BSONType]map[BSONType]BSONType)
	// First initialize each inner map, and set leastUpperBound(x,x) = x.
	for _, outer := range bsonTypeLatticeDeclaration {
		outerType := outer[0][0]
		ret[outerType] = make(map[BSONType]BSONType)
		ret[outerType][outerType] = outerType
	}

	for _, outer := range bsonTypeLatticeDeclaration {
		for _, inner := range outer {
			ret[inner[0]][inner[1]] = inner[2]
			ret[inner[1]][inner[0]] = inner[2]
		}
	}
	return ret
}

// LeastUpperBound computes the least upper bound of two BSONTypes with respect
// to the BSONTypeLattice.
func LeastUpperBound(left, right BSONType) BSONType {
	return bsonTypeLatticeLeastUpperBound[left][right]
}

// GetSpecialType returns the proper SpecialType based on the passed BSONType and
// SpecialType.
func GetSpecialType(bsonType BSONType, specialType SpecialType) SpecialType {
	if bsonType == BinData || specialType == GeoPoint {
		return specialType
	}
	return NoSpecialType
}
