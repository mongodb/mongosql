package normalizer_test

import (
	"testing"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/bsonutil"
	"github.com/10gen/mongoast/normalizer"
	"github.com/google/go-cmp/cmp"
)

func TestNormalize(t *testing.T) {
	testCases := []struct {
		name         string
		unnormalized ast.Node
		expected     ast.Node
	}{
		{
			"x < 5",
			ast.NewBinary(
				ast.LessThan,
				ast.NewFieldRef("x", nil),
				ast.NewConstant(bsonutil.Int64(5))),
			ast.NewBinary(
				ast.LessThan,
				ast.NewFieldRef("x", nil),
				ast.NewConstant(bsonutil.Int64(5))),
		},
		{
			"x <= 5",
			ast.NewBinary(
				ast.LessThanOrEquals,
				ast.NewFieldRef("x", nil),
				ast.NewConstant(bsonutil.Int64(5))),
			ast.NewBinary(
				ast.LessThanOrEquals,
				ast.NewFieldRef("x", nil),
				ast.NewConstant(bsonutil.Int64(5))),
		},
		{
			"x > 5",
			ast.NewBinary(
				ast.GreaterThan,
				ast.NewFieldRef("x", nil),
				ast.NewConstant(bsonutil.Int64(5))),
			ast.NewBinary(
				ast.GreaterThan,
				ast.NewFieldRef("x", nil),
				ast.NewConstant(bsonutil.Int64(5))),
		},
		{
			"x >= 5",
			ast.NewBinary(
				ast.GreaterThanOrEquals,
				ast.NewFieldRef("x", nil),
				ast.NewConstant(bsonutil.Int64(5))),
			ast.NewBinary(
				ast.GreaterThanOrEquals,
				ast.NewFieldRef("x", nil),
				ast.NewConstant(bsonutil.Int64(5))),
		},
		{
			"x == 5",
			ast.NewBinary(
				ast.Equals,
				ast.NewFieldRef("x", nil),
				ast.NewConstant(bsonutil.Int64(5))),
			ast.NewBinary(
				ast.Equals,
				ast.NewFieldRef("x", nil),
				ast.NewConstant(bsonutil.Int64(5))),
		},
		{
			"x != 5",
			ast.NewBinary(
				ast.NotEquals,
				ast.NewFieldRef("x", nil),
				ast.NewConstant(bsonutil.Int64(5))),
			ast.NewBinary(
				ast.NotEquals,
				ast.NewFieldRef("x", nil),
				ast.NewConstant(bsonutil.Int64(5))),
		},
		{
			"5 < x",
			ast.NewBinary(
				ast.LessThan,
				ast.NewConstant(bsonutil.Int64(5)),
				ast.NewFieldRef("x", nil)),
			ast.NewBinary(
				ast.GreaterThan,
				ast.NewFieldRef("x", nil),
				ast.NewConstant(bsonutil.Int64(5))),
		},
		{
			"5 <= x",
			ast.NewBinary(
				ast.LessThanOrEquals,
				ast.NewConstant(bsonutil.Int64(5)),
				ast.NewFieldRef("x", nil)),
			ast.NewBinary(
				ast.GreaterThanOrEquals,
				ast.NewFieldRef("x", nil),
				ast.NewConstant(bsonutil.Int64(5))),
		},
		{
			"5 > x",
			ast.NewBinary(
				ast.GreaterThan,
				ast.NewConstant(bsonutil.Int64(5)),
				ast.NewFieldRef("x", nil)),
			ast.NewBinary(
				ast.LessThan,
				ast.NewFieldRef("x", nil),
				ast.NewConstant(bsonutil.Int64(5))),
		},
		{
			"5 >= x",
			ast.NewBinary(
				ast.GreaterThanOrEquals,
				ast.NewConstant(bsonutil.Int64(5)),
				ast.NewFieldRef("x", nil)),
			ast.NewBinary(
				ast.LessThanOrEquals,
				ast.NewFieldRef("x", nil),
				ast.NewConstant(bsonutil.Int64(5))),
		},
		{
			"5 == x",
			ast.NewBinary(
				ast.Equals,
				ast.NewConstant(bsonutil.Int64(5)),
				ast.NewFieldRef("x", nil)),
			ast.NewBinary(
				ast.Equals,
				ast.NewFieldRef("x", nil),
				ast.NewConstant(bsonutil.Int64(5))),
		},
		{
			"5 != x",
			ast.NewBinary(
				ast.NotEquals,
				ast.NewConstant(bsonutil.Int64(5)),
				ast.NewFieldRef("x", nil)),
			ast.NewBinary(
				ast.NotEquals,
				ast.NewFieldRef("x", nil),
				ast.NewConstant(bsonutil.Int64(5))),
		},
		{
			"5 < x && x < 10",
			ast.NewBinary(
				ast.And,
				ast.NewBinary(
					ast.LessThan,
					ast.NewConstant(bsonutil.Int64(5)),
					ast.NewFieldRef("x", nil)),
				ast.NewBinary(
					ast.LessThan,
					ast.NewFieldRef("x", nil),
					ast.NewConstant(bsonutil.Int64(10)))),
			ast.NewBinary(
				ast.And,
				ast.NewBinary(
					ast.GreaterThan,
					ast.NewFieldRef("x", nil),
					ast.NewConstant(bsonutil.Int64(5))),
				ast.NewBinary(
					ast.LessThan,
					ast.NewFieldRef("x", nil),
					ast.NewConstant(bsonutil.Int64(10)))),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := normalizer.Normalize(tc.unnormalized)
			if !cmp.Equal(tc.expected, actual) {
				t.Fatalf("actual did not match expected\n  %s", cmp.Diff(tc.expected, actual))
			}
		})
	}
}
