package astutil

import (
	"time"

	"github.com/10gen/mongo-go-driver/bson"

	"github.com/10gen/mongoast/ast"

	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// BinaryConstant returns an ast.Constant with a Binary bson value.
func BinaryConstant(v bson.Binary) *ast.Constant {
	return ast.NewConstant(bsoncore.Value{
		Type: bsontype.Binary,
		Data: bsoncore.AppendBinary(nil, v.Kind, v.Data),
	})
}

// BooleanConstant returns an ast.Constant with a Boolean bson value.
func BooleanConstant(v bool) *ast.Constant {
	return ast.NewConstant(bsoncore.Value{
		Type: bsontype.Boolean,
		Data: bsoncore.AppendBoolean(nil, v),
	})
}

// DateConstant returns an ast.Constant with a DateTime bson value.
func DateConstant(v time.Time) *ast.Constant {
	return ast.NewConstant(bsoncore.Value{
		Type: bsontype.DateTime,
		Data: bsoncore.AppendDateTime(nil, v.Unix()*1000+int64(v.Nanosecond()/1e6)),
	})
}

// FloatConstant returns an ast.Constant with a Double bson value.
func FloatConstant(v float64) *ast.Constant {
	return ast.NewConstant(bsoncore.Value{
		Type: bsontype.Double,
		Data: bsoncore.AppendDouble(nil, v),
	})
}

// Int32Constant returns an ast.Constant with an Int32 bson value.
func Int32Constant(v int32) *ast.Constant {
	return ast.NewConstant(bsoncore.Value{
		Type: bsontype.Int32,
		Data: bsoncore.AppendInt32(nil, v),
	})
}

// Int64Constant returns an ast.Constant with an Int64 bson value.
func Int64Constant(v int64) *ast.Constant {
	return ast.NewConstant(bsoncore.Value{
		Type: bsontype.Int64,
		Data: bsoncore.AppendInt64(nil, v),
	})
}

// NullConstant returns an ast.Constant with a Null bson value.
func NullConstant() *ast.Constant {
	return ast.NewConstant(bsoncore.Value{
		Type: bsontype.Null,
	})
}

// ObjectIDConstant returns an ast.Constant with an ObjectID bson value.
func ObjectIDConstant(v primitive.ObjectID) *ast.Constant {
	return ast.NewConstant(bsoncore.Value{
		Type: bsontype.ObjectID,
		Data: bsoncore.AppendObjectID(nil, v),
	})
}

// StringConstant returns an ast.Constant with a String bson value.
func StringConstant(v string) *ast.Constant {
	return ast.NewConstant(bsoncore.Value{
		Type: bsontype.String,
		Data: bsoncore.AppendString(nil, v),
	})
}

// BooleanValue returns an ast.Unknown with a Boolean bson value. Unknowns
// are not wrapped in $literal.
func BooleanValue(v bool) *ast.Unknown {
	return ast.NewUnknown(bsoncore.Value{
		Type: bsontype.Boolean,
		Data: bsoncore.AppendBoolean(nil, v),
	})
}

// DateValue returns an ast.Unknown with a DateTime bson value. Unknowns
// are not wrapped in $literal.
func DateValue(v time.Time) *ast.Unknown {
	return ast.NewUnknown(bsoncore.Value{
		Type: bsontype.DateTime,
		Data: bsoncore.AppendDateTime(nil, v.Unix()*1000+int64(v.Nanosecond()/1e6)),
	})
}

// FloatValue returns an ast.Unknown with a Double bson value. Unknowns
// are not wrapped in $literal.
func FloatValue(v float64) *ast.Unknown {
	return ast.NewUnknown(bsoncore.Value{
		Type: bsontype.Double,
		Data: bsoncore.AppendDouble(nil, v),
	})
}

// Int32Value returns an ast.Unknown with an Int32 bson value. Unknowns
// are not wrapped in $literal.
func Int32Value(v int32) *ast.Unknown {
	return ast.NewUnknown(bsoncore.Value{
		Type: bsontype.Int32,
		Data: bsoncore.AppendInt32(nil, v),
	})
}

// Int64Value returns an ast.Unknown with an Int64 bson value. Unknowns
// are not wrapped in $literal.
func Int64Value(v int64) *ast.Unknown {
	return ast.NewUnknown(bsoncore.Value{
		Type: bsontype.Int64,
		Data: bsoncore.AppendInt64(nil, v),
	})
}

// NullValue returns an ast.Unknown with a Null bson value. Unknowns
// are not wrapped in $literal.
func NullValue() *ast.Unknown {
	return ast.NewUnknown(bsoncore.Value{
		Type: bsontype.Null,
	})
}

// ObjectIDValue returns an ast.Unknown with a ObjectID bson value. Unknowns
// are not wrapped in $literal.
func ObjectIDValue(v primitive.ObjectID) *ast.Unknown {
	return ast.NewUnknown(bsoncore.Value{
		Type: bsontype.ObjectID,
		Data: bsoncore.AppendObjectID(nil, v),
	})
}

// StringValue returns an ast.Unknown with a String bson value. Unknowns
// are not wrapped in $literal.
func StringValue(v string) *ast.Unknown {
	return ast.NewUnknown(bsoncore.Value{
		Type: bsontype.String,
		Data: bsoncore.AppendString(nil, v),
	})
}
