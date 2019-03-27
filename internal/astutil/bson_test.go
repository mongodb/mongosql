package astutil_test

import (
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/mongoast/ast"
	"github.com/10gen/sqlproxy/internal/astutil"
	"github.com/google/go-cmp/cmp"
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
				{{"$project", bson.D{
					{"x", bson.D{{"$add", []interface{}{"$a", "$b"}}}},
					{"_id", int32(0)},
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
				{{"$match", bson.D{
					{"$or", []interface{}{
						bson.D{{"x", int32(1)}},
						bson.D{{"x", int32(2)}},
					}},
				}}},
				{{"$skip", int64(5)}},
				{{"$addFields", bson.D{
					{"sum", bson.D{{"$add", []interface{}{"$x", "$y", "$z"}}}},
					{"cons", "yo"},
				}}},
				{{"$limit", int64(3)}},
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
				{{"$project", bson.D{
					{"x", bson.D{{"$add", []interface{}{"$a", "$b"}}}},
					{"_id", int32(0)},
				}}},
			},
			ast.NewPipeline(
				ast.NewProjectStage(
					ast.NewAssignProjectItem("x",
						// the mongoast does not recognize "$add" as an ast.Binary, so it will
						// be parsed as an ast.Function.
						ast.NewFunction("$add", ast.NewArray(
							ast.NewFieldRef("a", nil),
							ast.NewFieldRef("b", nil),
						)),
					),
					ast.NewExcludeProjectItem(ast.NewFieldRef("_id", nil)),
				),
			),
		},
		{
			"multiple stage pipeline",
			[]bson.D{
				{{"$match", bson.D{
					{"$or", []interface{}{
						bson.D{{"x", int32(1)}},
						bson.D{{"x", int32(2)}},
					}},
				}}},
				{{"$skip", int64(5)}},
				{{"$addFields", bson.D{
					{"sum", bson.D{{"$add", []interface{}{"$x", "$y", "$z"}}}},
					{"cons", bson.D{{"$literal", "yo"}}},
				}}},
				{{"$limit", int64(3)}},
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
						ast.NewFunction("$add", ast.NewArray(
							ast.NewFieldRef("x", nil),
							ast.NewFieldRef("y", nil),
							ast.NewFieldRef("z", nil),
						)),
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
