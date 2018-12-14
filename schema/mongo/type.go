package mongo

// BSONType is an enum representing all the bson types suported by Schema.
type BSONType string

// Possible values of BSONType
// for reference: https://docs.mongodb.com/manual/reference/operator/query/type/#available-types
const (
	// Ordered in bsonspec order.
	Double             BSONType = "double"
	String             BSONType = "string"
	Object             BSONType = "object"
	Array              BSONType = "array"
	BinData            BSONType = "binData"
	UnsupportedBinData BSONType = "unsupportedBinData"
	Undefined          BSONType = "undefined"
	ObjectID           BSONType = "objectId"
	Boolean            BSONType = "bool"
	Date               BSONType = "date"
	Null               BSONType = "null"
	Regex              BSONType = "regex"
	DBPointer          BSONType = "dbpointer"
	JSCode             BSONType = "jscode"
	Symbol             BSONType = "symbol"
	JSCodeWScope       BSONType = "jscode_w_scope"
	Int                BSONType = "int"
	Timestamp          BSONType = "timestamp"
	Long               BSONType = "long"
	Decimal            BSONType = "decimal"
	MinKey             BSONType = "minkey"
	MaxKey             BSONType = "maxkey"
	// Only used when no types are sampled.
	NoBSONType BSONType = "notype"
)

// IsNullType returns true if this type is Null.
func IsNullType(t BSONType) bool {
	return t == Null
}

// IsUnmappableType returns true if this type is currently unmappable.
func IsUnmappableType(t BSONType) bool {
	switch t {
	// We defined this in terms of the mappable types, so that any new types will automatically
	// be classed as unmappable.
	case Double, String, Object, Array, BinData, ObjectID, Boolean, Date, Null, Int, Long, Decimal:
		return false
	default:
		return true
	}
}

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
		{Int, Null, Int},
		{Int, UnsupportedBinData, Int},
		{Int, Undefined, Int},
		{Int, Regex, Int},
		{Int, DBPointer, Int},
		{Int, JSCode, Int},
		{Int, Symbol, Int},
		{Int, JSCodeWScope, Int},
		{Int, Timestamp, Int},
		{Int, MinKey, Int},
		{Int, MaxKey, Int},
	},
	{
		{Long, Double, Decimal},
		{Long, Decimal, Decimal},
		{Long, Boolean, Long},
		{Long, String, String},
		{Long, Date, String},
		{Long, BinData, String},
		{Long, ObjectID, String},
		{Long, Null, Long},
		{Long, UnsupportedBinData, Long},
		{Long, Undefined, Long},
		{Long, Regex, Long},
		{Long, DBPointer, Long},
		{Long, JSCode, Long},
		{Long, Symbol, Long},
		{Long, JSCodeWScope, Long},
		{Long, Timestamp, Long},
		{Long, MinKey, Long},
		{Long, MaxKey, Long},
	},
	{
		{Double, Decimal, Decimal},
		{Double, Boolean, Double},
		{Double, String, String},
		{Double, Date, String},
		{Double, BinData, String},
		{Double, ObjectID, String},
		{Double, Null, Double},
		{Double, UnsupportedBinData, Double},
		{Double, Undefined, Double},
		{Double, Regex, Double},
		{Double, DBPointer, Double},
		{Double, JSCode, Double},
		{Double, Symbol, Double},
		{Double, JSCodeWScope, Double},
		{Double, Timestamp, Double},
		{Double, MinKey, Double},
		{Double, MaxKey, Double},
	},
	{
		{Decimal, Boolean, Decimal},
		{Decimal, String, String},
		{Decimal, Date, String},
		{Decimal, BinData, String},
		{Decimal, ObjectID, String},
		{Decimal, Null, Decimal},
		{Decimal, UnsupportedBinData, Decimal},
		{Decimal, Undefined, Decimal},
		{Decimal, Regex, Decimal},
		{Decimal, DBPointer, Decimal},
		{Decimal, JSCode, Decimal},
		{Decimal, Symbol, Decimal},
		{Decimal, JSCodeWScope, Decimal},
		{Decimal, Timestamp, Decimal},
		{Decimal, MinKey, Decimal},
		{Decimal, MaxKey, Decimal},
	},
	{
		{Boolean, String, String},
		{Boolean, Date, String},
		{Boolean, BinData, String},
		{Boolean, ObjectID, String},
		{Boolean, Null, Boolean},
		{Boolean, UnsupportedBinData, Boolean},
		{Boolean, Undefined, Boolean},
		{Boolean, Regex, Boolean},
		{Boolean, DBPointer, Boolean},
		{Boolean, JSCode, Boolean},
		{Boolean, Symbol, Boolean},
		{Boolean, JSCodeWScope, Boolean},
		{Boolean, Timestamp, Boolean},
		{Boolean, MinKey, Boolean},
		{Boolean, MaxKey, Boolean},
	},
	{
		{String, Date, String},
		{String, BinData, String},
		{String, ObjectID, String},
		{String, Null, String},
		{String, UnsupportedBinData, String},
		{String, Undefined, String},
		{String, Regex, String},
		{String, DBPointer, String},
		{String, JSCode, String},
		{String, Symbol, String},
		{String, JSCodeWScope, String},
		{String, Timestamp, String},
		{String, MinKey, String},
		{String, MaxKey, String},
	},
	{
		{Date, BinData, String},
		{Date, ObjectID, String},
		{Date, Null, Date},
		{Date, UnsupportedBinData, Date},
		{Date, Undefined, Date},
		{Date, Regex, Date},
		{Date, DBPointer, Date},
		{Date, JSCode, Date},
		{Date, Symbol, Date},
		{Date, JSCodeWScope, Date},
		{Date, Timestamp, Date},
		{Date, MinKey, Date},
		{Date, MaxKey, Date},
	},
	{
		{BinData, ObjectID, String},
		{BinData, Null, BinData},
		{BinData, UnsupportedBinData, BinData},
		{BinData, Undefined, BinData},
		{BinData, Regex, BinData},
		{BinData, DBPointer, BinData},
		{BinData, JSCode, BinData},
		{BinData, Symbol, BinData},
		{BinData, JSCodeWScope, BinData},
		{BinData, Timestamp, BinData},
		{BinData, MinKey, BinData},
		{BinData, MaxKey, BinData},
	},
	{
		{ObjectID, Null, ObjectID},
		{ObjectID, UnsupportedBinData, ObjectID},
		{ObjectID, Undefined, ObjectID},
		{ObjectID, Regex, ObjectID},
		{ObjectID, DBPointer, ObjectID},
		{ObjectID, JSCode, ObjectID},
		{ObjectID, Symbol, ObjectID},
		{ObjectID, JSCodeWScope, ObjectID},
		{ObjectID, Timestamp, ObjectID},
		{ObjectID, MinKey, ObjectID},
		{ObjectID, MaxKey, ObjectID},
	},
	{
		{Null, UnsupportedBinData, Null},
		{Null, Undefined, Null},
		{Null, Regex, Null},
		{Null, DBPointer, Null},
		{Null, JSCode, Null},
		{Null, Symbol, Null},
		{Null, JSCodeWScope, Null},
		{Null, Timestamp, Null},
		{Null, MinKey, Null},
		{Null, MaxKey, Null},
	},
	{
		{UnsupportedBinData, Undefined, UnsupportedBinData},
		{UnsupportedBinData, Regex, UnsupportedBinData},
		{UnsupportedBinData, DBPointer, UnsupportedBinData},
		{UnsupportedBinData, JSCode, UnsupportedBinData},
		{UnsupportedBinData, Symbol, UnsupportedBinData},
		{UnsupportedBinData, JSCodeWScope, UnsupportedBinData},
		{UnsupportedBinData, Timestamp, UnsupportedBinData},
		{UnsupportedBinData, MinKey, UnsupportedBinData},
		{UnsupportedBinData, MaxKey, UnsupportedBinData},
	},
	{
		{Undefined, Regex, Undefined},
		{Undefined, DBPointer, Undefined},
		{Undefined, JSCode, Undefined},
		{Undefined, Symbol, Undefined},
		{Undefined, JSCodeWScope, Undefined},
		{Undefined, Timestamp, Undefined},
		{Undefined, MinKey, Undefined},
		{Undefined, MaxKey, Undefined},
	},
	{
		{Regex, DBPointer, Regex},
		{Regex, JSCode, Regex},
		{Regex, Symbol, Regex},
		{Regex, JSCodeWScope, Regex},
		{Regex, Timestamp, Regex},
		{Regex, MinKey, Regex},
		{Regex, MaxKey, Regex},
	},
	{
		{DBPointer, JSCode, DBPointer},
		{DBPointer, Symbol, DBPointer},
		{DBPointer, JSCodeWScope, DBPointer},
		{DBPointer, Timestamp, DBPointer},
		{DBPointer, MinKey, DBPointer},
		{DBPointer, MaxKey, DBPointer},
	},
	{
		{JSCode, Symbol, JSCode},
		{JSCode, JSCodeWScope, JSCode},
		{JSCode, Timestamp, JSCode},
		{JSCode, MinKey, JSCode},
		{JSCode, MaxKey, JSCode},
	},
	{
		{Symbol, JSCodeWScope, Symbol},
		{Symbol, Timestamp, Symbol},
		{Symbol, MinKey, Symbol},
		{Symbol, MaxKey, Symbol},
	},
	{
		{JSCodeWScope, Timestamp, JSCodeWScope},
		{JSCodeWScope, MinKey, JSCodeWScope},
		{JSCodeWScope, MaxKey, JSCodeWScope},
	},
	{
		{Timestamp, MinKey, Timestamp},
		{Timestamp, MaxKey, Timestamp},
	},
	{
		{MinKey, MaxKey, MinKey},
	},
	{
		{MaxKey, MaxKey, MaxKey},
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
	ret := bsonTypeLatticeLeastUpperBound[left][right]
	return ret
}

// GetSpecialType returns the proper SpecialType based on the passed BSONType and
// SpecialType.
func GetSpecialType(bsonType BSONType, specialType SpecialType) SpecialType {
	if bsonType == BinData || specialType == GeoPoint {
		return specialType
	}
	return NoSpecialType
}
