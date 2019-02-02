package values

// MongoSQLBool represents a boolean with MongoDB type conversion semantics.
// MongoSQLBool shares all of its conversion implementations with BaseSQLBool.
type MongoSQLBool struct {
	BaseSQLBool
}

// MongoSQLDate represents a date with MongoDB type conversion semantics.
// MongoSQLDate shares all of its conversion implementations with BaseSQLDate.
type MongoSQLDate struct {
	BaseSQLDate
}

// MongoSQLDecimal128 represents a decimal with MongoDB type conversion semantics.
// MongoSQLDecimal128 shares all of its conversion implementations with BaseSQLDecimal128.
type MongoSQLDecimal128 struct {
	BaseSQLDecimal128
}

// MongoSQLFloat represents a float with MongoDB type conversion semantics.
// MongoSQLFloat shares all of its conversion implementations with BaseSQLFloat.
type MongoSQLFloat struct {
	BaseSQLFloat
}

// MongoSQLInt64 represents an int64 with MongoDB type conversion semantics.
// MongoSQLInt64 shares all of its conversion implementations with BaseSQLInt64.
type MongoSQLInt64 struct {
	BaseSQLInt64
}

// MongoSQLObjectID represents an ObjectID with MongoDB type conversion semantics.
// MongoSQLObjectID shares all of its conversion implementations with BaseSQLObjectID.
type MongoSQLObjectID struct {
	BaseSQLObjectID
}

// MongoSQLTimestamp represents a timestamp with MongoDB type conversion semantics.
// MongoSQLTimestamp shares all of its conversion implementations with BaseSQLTimestamp.
type MongoSQLTimestamp struct {
	BaseSQLTimestamp
}

// MongoSQLUint64 represents a uint64 with MongoDB type conversion semantics.
// MongoSQLUint64 shares all of its conversion implementations with BaseSQLUint64.
type MongoSQLUint64 struct {
	BaseSQLUint64
}

// MongoSQLVarchar represents a varchar with MongoDB type conversion semantics.
// MongoSQLVarchar shares all of its conversion implementations with BaseSQLVarchar.
type MongoSQLVarchar struct {
	BaseSQLVarchar
}

// MongoSQLNull represents a varchar with MongoDB type conversion semantics.
// MongoSQLNull shares all of its conversion implementations with BaseSQLNull.
type MongoSQLNull struct {
	BaseSQLNull
}
