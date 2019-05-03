package parser_test

import (
	"testing"
	"time"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/bsonutil"
	"github.com/10gen/mongoast/internal/parsertest"
	"github.com/10gen/mongoast/parser"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestParseExpr(t *testing.T) {
	testCases := []struct {
		input    string
		expected ast.Expr
		err      error
	}{
		// Constants
		{
			`null`,
			ast.NewConstant(bsonutil.Null()),
			nil,
		},
		{
			`{"$undefined": true}`,
			ast.NewConstant(bsonutil.Undefined()),
			nil,
		},
		{
			`1`,
			ast.NewConstant(bsonutil.Int32(1)),
			nil,
		},
		{
			`{"$numberLong": "1"}`,
			ast.NewConstant(bsonutil.Int64(1)),
			nil,
		},
		{
			`{"$numberDouble": "1"}`,
			ast.NewConstant(bsonutil.Double(1)),
			nil,
		},
		{
			`"a"`,
			ast.NewConstant(bsonutil.String("a")),
			nil,
		},
		{
			`{"$numberDecimal": "1"}`,
			ast.NewConstant(bsonutil.Decimal128FromInt64(1)),
			nil,
		},
		{
			`{"$date": "2019-01-01T00:00:00Z" }`,
			ast.NewConstant(bsonutil.DateTime(parseISODate("2019-01-01T00:00:00Z"))),
			nil,
		},
		{
			`{"$oid": "123456789012345678901234" }`,
			ast.NewConstant(
				bsonutil.ObjectID(
					parseObjectID("123456789012345678901234"),
				),
			),
			nil,
		},
		{
			`{"$regularExpression": {"pattern": "abc", "options": "g"}}`,
			ast.NewConstant(bsonutil.Regex("abc", "g")),
			nil,
		},
		{
			`{"$binary": {"base64": "dGVzdA==", "subType": "00"}}`,
			ast.NewConstant(bsonutil.Binary(0x00, []byte{'t', 'e', 's', 't'})),
			nil,
		},
		{
			`{"$timestamp": {"t": 1, "i": 2}}`,
			ast.NewConstant(bsonutil.Timestamp(1, 2)),
			nil,
		},
		{
			`{"a": 1}`,
			ast.NewDocument(
				ast.NewDocumentElement("a", ast.NewConstant(bsonutil.Int32(1))),
			),
			nil,
		},
		{
			`{"$literal": {"a": 1}}`,
			ast.NewConstant(
				bsonutil.DocumentFromElements("a", bsonutil.Int32(1)),
			),
			nil,
		},
		{
			`[1,2]`,
			ast.NewArray(
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(2)),
			),
			nil,
		},
		{
			`{"$literal": [1,2]}`,
			ast.NewConstant(
				bsonutil.ArrayFromValues(
					bsonutil.Int32(1),
					bsonutil.Int32(2),
				),
			),
			nil,
		},
		{
			`{"$minKey": 1}`,
			ast.NewConstant(bsonutil.MinKey()),
			nil,
		},
		{
			`{"$maxKey": 1}`,
			ast.NewConstant(bsonutil.MaxKey()),
			nil,
		},
		{
			`{"$symbol": "foo"}`,
			ast.NewConstant(bsonutil.Symbol("foo")),
			nil,
		},
		{
			`{"$dbPointer": {"$ref": "foo", "$id": {"$oid": "123456789012345678901234"}}}`,
			ast.NewConstant(
				bsonutil.DBPointer("foo", parseObjectID("123456789012345678901234"))),
			nil,
		},
		{
			`{"$code": "function() {}"}`,
			ast.NewConstant(bsonutil.JavaScript("function() {}")),
			nil,
		},
		{
			`{"$code": "function() { return x; }", "$scope": {"x": 1}}`,
			ast.NewConstant(
				bsonutil.CodeWithScope(
					"function() { return x; }",
					bsonutil.DocumentFromElements("x", bsonutil.Int32(1)).Document(),
				),
			),
			nil,
		},
		// Variables
		{
			`"$$a"`,
			ast.NewVariableRef("a"),
			nil,
		},
		{
			`"$$a.b"`,
			ast.NewFieldRef("b", ast.NewVariableRef("a")),
			nil,
		},
		{
			`"$$a.b.c"`,
			ast.NewFieldRef("c", ast.NewFieldRef("b", ast.NewVariableRef("a"))),
			nil,
		},
		// FieldRef
		{
			`"$a"`,
			ast.NewFieldRef("a", nil),
			nil,
		},
		{
			`"$a.b"`,
			ast.NewFieldRef("b", ast.NewFieldRef("a", nil)),
			nil,
		},
		{
			`"$a.b.c"`,
			ast.NewFieldRef("c", ast.NewFieldRef("b", ast.NewFieldRef("a", nil))),
			nil,
		},
		// FieldOrArrayIndexRef
		{
			`"$a.1"`,
			ast.NewFieldOrArrayIndexRef(
				1, ast.NewFieldRef("a", nil),
			),
			nil,
		},
		{
			`"$a.1.2"`,
			ast.NewFieldOrArrayIndexRef(
				2, ast.NewFieldOrArrayIndexRef(
					1, ast.NewFieldRef("a", nil),
				),
			),
			nil,
		},
		{
			`"$a.1.b"`,
			ast.NewFieldRef(
				"b", ast.NewFieldOrArrayIndexRef(
					1, ast.NewFieldRef("a", nil),
				),
			),
			nil,
		},
		// Logical
		{
			`{ "$and": [1] }`,
			ast.NewConstant(bsonutil.Int32(1)),
			nil,
		},
		{
			`{ "$and": [1, 0] }`,
			ast.NewBinary(
				ast.And,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(0)),
			),
			nil,
		},
		{
			`{ "$and": [1, 0, 2] }`,
			ast.NewBinary(
				ast.And,
				ast.NewBinary(
					ast.And,
					ast.NewConstant(bsonutil.Int32(1)),
					ast.NewConstant(bsonutil.Int32(0)),
				),
				ast.NewConstant(bsonutil.Int32(2)),
			),
			nil,
		},
		{
			`{ "$and": 1 }`,
			nil,
			errors.New("$and requires an array"),
		},
		{
			`{ "$or": [1] }`,
			ast.NewConstant(bsonutil.Int32(1)),
			nil,
		},
		{
			`{ "$or": [1, 0] }`,
			ast.NewBinary(
				ast.Or,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(0)),
			),
			nil,
		},
		{
			`{ "$or": [1, 0, 2] }`,
			ast.NewBinary(
				ast.Or,
				ast.NewBinary(
					ast.Or,
					ast.NewConstant(bsonutil.Int32(1)),
					ast.NewConstant(bsonutil.Int32(0)),
				),
				ast.NewConstant(bsonutil.Int32(2)),
			),
			nil,
		},
		{
			`{ "$or": 1 }`,
			nil,
			errors.New("$or requires an array"),
		},
		// Comparisons
		{
			`{ "$eq": [1, 0] }`,
			ast.NewBinary(
				ast.Equals,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(0)),
			),
			nil,
		},
		{
			`{ "$eq": 1 }`,
			nil,
			errors.New("$eq requires an array with 2 elements"),
		},
		{
			`{ "$eq": [1] }`,
			nil,
			errors.New("$eq requires an array with 2 elements"),
		},
		{
			`{ "$eq": [1, 2, 3] }`,
			nil,
			errors.New("$eq requires an array with 2 elements"),
		},
		{
			`{ "$gt": [1, 0] }`,
			ast.NewBinary(
				ast.GreaterThan,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(0)),
			),
			nil,
		},
		{
			`{ "$gte": [1, 0] }`,
			ast.NewBinary(
				ast.GreaterThanOrEquals,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(0)),
			),
			nil,
		},
		{
			`{ "$lt": [1, 0] }`,
			ast.NewBinary(
				ast.LessThan,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(0)),
			),
			nil,
		},
		{
			`{ "$lte": [1, 0] }`,
			ast.NewBinary(
				ast.LessThanOrEquals,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(0)),
			),
			nil,
		},
		{
			`{ "$ne": [1, 0] }`,
			ast.NewBinary(
				ast.NotEquals,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(0)),
			),
			nil,
		},
		{
			`{ "$ne": [1, {}] }`,
			ast.NewBinary(
				ast.NotEquals,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewDocument(),
			),
			nil,
		},
		{
			`{ "$cmp": [1, 2] }`,
			ast.NewBinary(
				ast.Compare,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(2)),
			),
			nil,
		},
		{
			`{ "$or": [1] }`,
			ast.NewConstant(bsonutil.Int32(1)),
			nil,
		},
		{
			`{ "$or": [1, 0] }`,
			ast.NewBinary(
				ast.Or,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(0)),
			),
			nil,
		},
		// Function
		{
			`{ "$arrayElemAt": ["$a", 2] }`,
			ast.NewArrayIndexRef(
				ast.NewConstant(bsonutil.Int32(2)),
				ast.NewFieldRef("a", nil),
			),
			nil,
		},
		{
			`{ "$arrayElemAt": ["$a", "$b"] }`,
			ast.NewArrayIndexRef(
				ast.NewFieldRef("b", nil),
				ast.NewFieldRef("a", nil),
			),
			nil,
		},
		{
			`{ "$arrayElemAt": "$a" }`,
			nil,
			errors.New("$arrayElemAt requires an array with 2 elements"),
		},
		{
			`{ "$arrayElemAt": ["$a"] }`,
			nil,
			errors.New("$arrayElemAt requires an array with 2 elements"),
		},
		{
			`{ "$arrayElemAt": ["$a", 1, 2] }`,
			nil,
			errors.New("$arrayElemAt requires an array with 2 elements"),
		},
		{
			`{ "$sum": "$a" }`,
			ast.NewFunction("$sum", ast.NewFieldRef("a", nil)),
			nil,
		},
		{
			`{ "$sum": ["$a"] }`,
			ast.NewFunction("$sum", ast.NewArray(ast.NewFieldRef("a", nil))),
			nil,
		},
		{
			`{ "$ltrim": { "input": "$a", "chars": "abc" } }`,
			ast.NewFunction(
				"$ltrim",
				ast.NewDocument(
					ast.NewDocumentElement("input", ast.NewFieldRef("a", nil)),
					ast.NewDocumentElement("chars", ast.NewConstant(bsonutil.String("abc"))),
				),
			),
			nil,
		},
		// Let
		{
			`{ "$let": { "vars": { "a": 1, "b": "$x" }, "in": { "$sum": ["$$a", "$$b"] } } }`,
			ast.NewLet(
				[]*ast.LetVariable{
					ast.NewLetVariable("a", ast.NewConstant(bsonutil.Int32(1))),
					ast.NewLetVariable("b", ast.NewFieldRef("x", nil)),
				},
				ast.NewFunction(
					"$sum",
					ast.NewArray(
						ast.NewVariableRef("a"),
						ast.NewVariableRef("b"),
					),
				),
			),
			nil,
		},
		{
			`{ "$let": 1 }`,
			nil,
			errors.New("$let requires a document"),
		},
		{
			`{ "$let": { "vars": { "a": 1 }, "in": "$$a", "foo": 1 } }`,
			nil,
			errors.New("unrecognized parameter to $let: foo"),
		},
		{
			`{ "$let": { "in": "$$a" } }`,
			nil,
			errors.New("missing 'vars' parameter to $let"),
		},
		{
			`{ "$let": { "vars": { "a": 1 } } }`,
			nil,
			errors.New("missing 'in' parameter to $let"),
		},
		{
			`{ "$let": { "vars": "a", "in": "$$a" } }`,
			nil,
			errors.New("invalid parameter: expected an object (vars)"),
		},
		// Conditional
		{
			`{ "$cond": { "if": { "$eq": ["$a", 5] }, "then": 1, "else": 0 } }`,
			ast.NewConditional(
				ast.NewBinary(
					ast.Equals,
					ast.NewFieldRef("a", nil),
					ast.NewConstant(bsonutil.Int32(5)),
				),
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(0)),
			),
			nil,
		},
		{
			`{ "$cond": [{ "$eq": ["$a", 5] }, 1, 0] }`,
			ast.NewConditional(
				ast.NewBinary(
					ast.Equals,
					ast.NewFieldRef("a", nil),
					ast.NewConstant(bsonutil.Int32(5)),
				),
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(0)),
			),
			nil,
		},
		{
			`{ "$cond": true }`,
			nil,
			errors.New("$cond requires a document or an array"),
		},
		{
			`{ "$cond": { "if": true, "then": 1, "else": 0, "foo": 1 } }`,
			nil,
			errors.New("unrecognized parameter to $cond: foo"),
		},
		{
			`{ "$cond": { "then": 1, "else": 0 } }`,
			nil,
			errors.New("missing 'if' parameter to $cond"),
		},
		{
			`{ "$cond": { "if": true, "else": 0 } }`,
			nil,
			errors.New("missing 'then' parameter to $cond"),
		},
		{
			`{ "$cond": { "if": true, "then": 1 } }`,
			nil,
			errors.New("missing 'else' parameter to $cond"),
		},
		{
			`{ "$cond": [true, 1] }`,
			nil,
			errors.New("expression $cond takes exactly 3 arguments"),
		},
		{
			`{ "$cond": [true, 1, 2, 3] }`,
			nil,
			errors.New("expression $cond takes exactly 3 arguments"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			actual, err := parsertest.ParseExprErr(tc.input)

			if err != nil && tc.err == nil {
				t.Fatalf("err should be nil, but was %v", err)
			} else if err == nil && tc.err != nil {
				t.Fatalf("err should not be nil, expected %v", tc.err)
			} else if err != nil && tc.err != nil && err.Error() != tc.err.Error() {
				t.Fatalf("expected error %q, but got %q", tc.err.Error(), err.Error())
			}

			if tc.err == nil && !cmp.Equal(tc.expected, actual) {
				t.Fatalf("pipelines are not equal\n  %s", cmp.Diff(tc.expected, actual))
			}
		})
	}
}

func TestParseFieldRef(t *testing.T) {
	testCases := []struct {
		input    string
		expected ast.Expr
	}{
		{
			"a",
			ast.NewFieldRef("a", nil),
		},
		{
			"a.b",
			ast.NewFieldRef(
				"b", ast.NewFieldRef("a", nil),
			),
		},
		{
			"a.b.c",
			ast.NewFieldRef(
				"c", ast.NewFieldRef(
					"b", ast.NewFieldRef("a", nil),
				),
			),
		},
		{
			"a.1",
			ast.NewFieldOrArrayIndexRef(
				1, ast.NewFieldRef("a", nil),
			),
		},
		{
			"a.1.2",
			ast.NewFieldOrArrayIndexRef(
				2, ast.NewFieldOrArrayIndexRef(
					1, ast.NewFieldRef("a", nil),
				),
			),
		},
		{
			"a.1.b",
			ast.NewFieldRef(
				"b", ast.NewFieldOrArrayIndexRef(
					1, ast.NewFieldRef("a", nil),
				),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			actual, err := parser.ParseFieldRef(tc.input)
			if err != nil {
				t.Fatalf("expected no err, but got %v", err)
			}

			if !cmp.Equal(tc.expected, actual) {
				t.Fatalf("field references are not equals\n  %s", cmp.Diff(tc.expected, actual))
			}
		})
	}
}

func parseISODate(dateString string) int64 {
	t, _ := time.Parse(time.RFC3339, dateString)
	return t.Unix() * 1000
}

func parseObjectID(hex string) primitive.ObjectID {
	v, _ := primitive.ObjectIDFromHex(hex)
	return v
}
