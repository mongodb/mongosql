package desugarer

import (
	"math"

	"github.com/10gen/mongoast/ast"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

var (
	nullLiteral = ast.NewConstant(bsoncore.Value{
		Type: bsontype.Null,
	})

	nanLiteral = ast.NewConstant(bsoncore.Value{
		Type: bsontype.Double,
		Data: bsoncore.AppendDouble(nil, math.NaN()),
	})

	negTwentyLiteral = ast.NewConstant(bsoncore.Value{
		Type: bsontype.Int32,
		Data: bsoncore.AppendInt32(nil, -20),
	})

	zeroLiteral = ast.NewConstant(bsoncore.Value{
		Type: bsontype.Int32,
		Data: bsoncore.AppendInt32(nil, 0),
	})

	oneLiteral = ast.NewConstant(bsoncore.Value{
		Type: bsontype.Int32,
		Data: bsoncore.AppendInt32(nil, 1),
	})

	oneHundredLiteral = ast.NewConstant(bsoncore.Value{
		Type: bsontype.Int32,
		Data: bsoncore.AppendInt32(nil, 100),
	})

	trueLiteral = ast.NewConstant(bsoncore.Value{
		Type: bsontype.Boolean,
		Data: bsoncore.AppendBoolean(nil, true),
	})

	falseLiteral = ast.NewConstant(bsoncore.Value{
		Type: bsontype.Boolean,
		Data: bsoncore.AppendBoolean(nil, false),
	})

	nullStringLiteral = ast.NewConstant(bsoncore.Value{
		Type: bsontype.String,
		Data: bsoncore.AppendString(nil, "null"),
	})

	missingStringLiteral = ast.NewConstant(bsoncore.Value{
		Type: bsontype.String,
		Data: bsoncore.AppendString(nil, "missing"),
	})
)

func stringLiteral(s string) *ast.Constant {
	return ast.NewConstant(bsoncore.Value{
		Type: bsontype.String,
		Data: bsoncore.AppendString(nil, s),
	})
}
