package astutil_test

import (
	"testing"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/sqlproxy/internal/astutil"
	"github.com/google/go-cmp/cmp"

	"go.mongodb.org/mongo-driver/bson"
)

func TestDeparsePipeline(t *testing.T) {
	tests := []struct {
		name     string
		pipeline *ast.Pipeline
		expected []bson.D
	}{
		{
			"single stage pipeline",
			ast.NewPipeline(
				ast.NewProjectStage(
					ast.NewAssignProjectItem("x",
						ast.NewBinary("$add",
							ast.NewFieldRef("a", nil),
							ast.NewFieldRef("b", nil),
						),
					),
					ast.NewExcludeProjectItem(ast.NewFieldRef("_id", nil)),
				),
			),
			[]bson.D{
				{{Key: "$project", Value: bson.D{
					{Key: "x", Value: bson.D{{Key: "$add", Value: bson.A{"$a", "$b"}}}},
					{Key: "_id", Value: int32(0)},
				}}},
			},
		},
		{
			"multiple stage pipeline",
			ast.NewPipeline(
				ast.NewMatchStage(ast.NewBinary("$or",
					ast.NewDocument(ast.NewDocumentElement("x", astutil.Int32Value(1))),
					ast.NewDocument(ast.NewDocumentElement("x", astutil.Int32Value(2))),
				)),
				ast.NewSkipStage(5),
				ast.NewAddFieldsStage(
					ast.NewAddFieldsItem("sum",
						ast.NewFunction("$add", ast.NewArray(
							ast.NewFieldRef("x", nil),
							ast.NewFieldRef("y", nil),
							ast.NewFieldRef("z", nil),
						)),
					),
					ast.NewAddFieldsItem("cons", astutil.StringValue("yo")),
				),
				ast.NewLimitStage(3),
			),
			[]bson.D{
				{{Key: "$match", Value: bson.D{
					{Key: "$or", Value: bson.A{
						bson.D{{Key: "x", Value: int32(1)}},
						bson.D{{Key: "x", Value: int32(2)}},
					}},
				}}},
				{{Key: "$skip", Value: int64(5)}},
				{{Key: "$addFields", Value: bson.D{
					{Key: "sum", Value: bson.D{{Key: "$add", Value: bson.A{"$x", "$y", "$z"}}}},
					{Key: "cons", Value: "yo"},
				}}},
				{{Key: "$limit", Value: int64(3)}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := astutil.DeparsePipeline(tc.pipeline)
			if err != nil {
				t.Fatalf("unexpected error deparsing pipeline: %v", err)
			}

			if !cmp.Equal(tc.expected, actual) {
				t.Fatalf("incorrect pipeline (- expected, + actual):\n%s", cmp.Diff(tc.expected, actual))
			}
		})
	}
}

func TestParsePipeline(t *testing.T) {
	tests := []struct {
		name     string
		docs     []bson.D
		expected *ast.Pipeline
	}{
		{
			"single stage pipeline",
			[]bson.D{
				{{Key: "$project", Value: bson.D{
					{Key: "x", Value: bson.D{{Key: "$add", Value: []interface{}{"$a", "$b"}}}},
					{Key: "_id", Value: int32(0)},
				}}},
			},
			ast.NewPipeline(
				ast.NewProjectStage(
					ast.NewAssignProjectItem("x",
						ast.NewBinary(ast.Add,
							ast.NewFieldRef("a", nil),
							ast.NewFieldRef("b", nil),
						),
					),
					ast.NewExcludeProjectItem(ast.NewFieldRef("_id", nil)),
				),
			),
		},
		{
			"multiple stage pipeline",
			[]bson.D{
				{{Key: "$match", Value: bson.D{
					{Key: "$or", Value: []interface{}{
						bson.D{{Key: "x", Value: int32(1)}},
						bson.D{{Key: "x", Value: int32(2)}},
					}},
				}}},
				{{Key: "$skip", Value: int64(5)}},
				{{Key: "$addFields", Value: bson.D{
					{Key: "sum", Value: bson.D{{Key: "$add", Value: []interface{}{"$x", "$y", "$z"}}}},
					{Key: "cons", Value: bson.D{{Key: "$literal", Value: "yo"}}},
				}}},
				{{Key: "$limit", Value: int64(3)}},
			},
			ast.NewPipeline(
				ast.NewMatchStage(ast.NewBinary("$or",
					// the mongoast explicitly turns field equality checks in match
					// expressions into $eq ast.Binaries.
					ast.NewBinary("$eq", ast.NewFieldRef("x", nil), astutil.Int32Constant(1)),
					ast.NewBinary("$eq", ast.NewFieldRef("x", nil), astutil.Int32Constant(2)),
				)),
				ast.NewSkipStage(5),
				ast.NewAddFieldsStage(
					ast.NewAddFieldsItem("sum",
						ast.NewBinary(ast.Add,
							ast.NewBinary(ast.Add,
								ast.NewFieldRef("x", nil),
								ast.NewFieldRef("y", nil),
							),
							ast.NewFieldRef("z", nil),
						),
					),
					ast.NewAddFieldsItem("cons", astutil.StringConstant("yo")),
				),
				ast.NewLimitStage(3),
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := astutil.ParsePipeline(tc.docs)
			if err != nil {
				t.Fatalf("unexpected error deparsing pipeline: %v", err)
			}

			if !cmp.Equal(tc.expected, actual) {
				t.Fatalf("incorrect pipeline (- expected, + actual):\n%s", cmp.Diff(tc.expected, actual))
			}
		})
	}
}
