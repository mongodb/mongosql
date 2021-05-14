package desugarer

import (
	"github.com/10gen/mongoast/ast"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

var (
	nullLiteral = ast.NewConstant(bsoncore.Value{
		Type: bsontype.Null,
	})

	zeroLiteral = ast.NewConstant(bsoncore.Value{
		Type: bsontype.Int32,
		Data: bsoncore.AppendInt32(nil, 0),
	})
)
