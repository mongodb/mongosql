package desugarer

import (
	"fmt"
	"strings"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/parser"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

type subqueryExprDesugarer func(*ast.Function, string) (ast.Expr, *ast.LookupStage)

var subqueryExprDesugarers = map[string]subqueryExprDesugarer{
	"$subquery":           desugarSubquery,
	"$subqueryExists":     desugarSubqueryExists,
	"$subqueryComparison": desugarSubqueryComparison,
}

// desugarSubqueryExprs desugars any non-nested $subquery expression (including
// $subqueryExists and $subqueryComparison) into a sequence of equivalent,
// existing stages and expression. A "nested" subquery expression is one that
// exists in a sub-pipeline. This desugarer should skip such expressions since
// we do not want to pull those out to the top-level. Since this pass will only
// desugar subquery expressions in the root pipeline, this pass must be run
// separately against all subpipelines if the caller wishes to desugar all
// subquery expressions.
func desugarSubqueryExprs(pipeline *ast.Pipeline, _ uint64) *ast.Pipeline {

	// A subquery expression is desugared into three parts:
	//   1. A $lookup stage that is placed before the containing stage. This
	//      $lookup stage performs the subquery.
	//   2. An expression that replaces the subquery expression in the
	//      containing stage. This replacement is dependent on the type of
	//      subquery expression.
	//   3. A $project stage that is placed after the containing stage. This
	//      $project stage removes the field introduced by the $lookup. If
	//      there are multiple subquery expressions in a stage, there will
	//      be multiple $lookup stages but only one $project stage at the
	//      end.
	//
	// At minimum, we know the pipeline will be at least the same length, so
	// we start with that capacity.
	newStages := make([]ast.Stage, 0, len(pipeline.Stages))

	for _, stage := range pipeline.Stages {
		// When there are multiple subquery expressions in a single stage,
		// the desugaring results in multiple $lookups. Each one needs to
		// use a different name for the "as" field, so a counter is used
		// to make the "as" names unique.
		subqueryCounter := 0

		var asNames []string

		out, _ := ast.Visit(stage, func(v ast.Visitor, n ast.Node) ast.Node {
			switch tn := n.(type) {
			case *ast.Pipeline:
				// Do not walk sub-pipelines.
				return n
			case *ast.Function:
				if desugarerFunc, ok := subqueryExprDesugarers[tn.Name]; ok {
					asName := fmt.Sprintf("__subquery_result_%d", subqueryCounter)
					replacement, lookupStage := desugarerFunc(tn, asName)
					newStages = append(newStages, lookupStage)
					asNames = append(asNames, asName)
					subqueryCounter += 1

					// There is no need to walk the replacement expr.
					return replacement
				}
			}

			return n.Walk(v)
		})

		newStages = append(newStages, out.(ast.Stage))

		// If $lookup stages were introduced, we must exclude those fields at the end
		if len(asNames) > 0 {
			newStages = append(newStages, makeExclusionStage(asNames))
		}
	}

	return ast.NewPipeline(newStages...)
}

// desugarSubquery desugars a $subquery expression into a $lookup and a
// replacement expression that accesses the data specified by the outputPath.
// Specifically, the desugaring turns
//
//	{ $subquery: {
//	    db: <db name>,
//	    collection: <collection name>,
//	    let: <let doc>,
//	    outputPath: [<component_1>, ...],
//	    pipeline: <pipeline>,
//	}}
//
// into the $lookup stage:
//
//	{ $lookup: {
//	    from: {
//	        db: <db name>,
//	        coll: <coll name>
//	    },
//	    let: <let doc>,
//	    pipeline: <pipeline>,
//	    as: "<asName>"
//	}}
//
// and the replacement expression:
//
//	{ $let: {
//	    vars: {
//	        docExpr: {
//	            $arrayElemAt: ["$<asName>", 0]
//	        }
//	    },
//	    in: "$$docExpr.<component_1>..."
//	}}
//
// The $lookup stage is placed before the stage that contains the $subquery
// expression. The replacement expression replaces the $subquery expression
// in the stage that contains it.
func desugarSubquery(f *ast.Function, asName string) (ast.Expr, *ast.LookupStage) {
	subqueryArgs := parseSubqueryArgs(f.Arg, "docExpr")

	expr := ast.NewLet(
		[]*ast.LetVariable{
			ast.NewLetVariable("docExpr",
				ast.NewFunction("$arrayElemAt", ast.NewArray(
					ast.NewFieldRef(asName, nil),
					zeroLiteral,
				)),
			),
		},
		subqueryArgs.outputPath,
	)

	lookup := subqueryArgs.toLookupStage(asName, true)

	return expr, lookup
}

// desugarSubqueryExists desugars a $subqueryExists expression into a $lookup
// and a replacement expression that checks if the output is non-empty.
// Specifically, the desugaring turns
//
//	{ $subqueryExists: {
//	    db: <db name>,
//	    collection: <collection name>,
//	    let: <let doc>,
//	    pipeline: <pipeline>,
//	}}
//
// into the same $lookup stage as for $subquery (above), and
// the replacement expression:
//
//	{ $gt: [{ $size: "$<asName>" }, 0] }
func desugarSubqueryExists(f *ast.Function, asName string) (ast.Expr, *ast.LookupStage) {
	subqueryArgs := parseSubqueryArgs(f.Arg, "")

	expr := ast.NewBinary(ast.GreaterThan,
		ast.NewFunction("$size", ast.NewFieldRef(asName, nil)),
		zeroLiteral,
	)

	lookup := subqueryArgs.toLookupStage(asName, true)

	return expr, lookup
}

// desugarSubqueryComparison desugars a $subqueryComparison expression into a
// $lookup and a replacement expression that performs the comparison on the
// subquery output.
// Specifically, the desugaring turns
//
//	{ $subqueryComparison: {
//	    op: <comp op>,
//	    modifier: <modifier>,
//	    argument: <arg>,
//	    subquery: <subquery>
//	}}
//
// into the same $lookup stage as for $subquery (above) using the <subquery>
// parameter (which has the same fields as $subquery), and the replacement
// expression:
//
//	{ $reduce: {
//	    "input": "$<asName>",
//	    "initialValue": <initial value>,
//	    "in": {
//	        <combinator func>: [
//	            "$$value",
//	            {
//	                "$<comp op>": [
//	                    <arg>,
//	                    "$$this.<component_1>..."
//	                ]
//	            }
//	        ]
//	    }
//	}}
//
// When <modifier> is "any", <initial value> is false and <combinator func>
// is "$sqlOr". When <modifier> is "all", <initial value> is true and
// <combinator func> is "$sqlAnd".
func desugarSubqueryComparison(f *ast.Function, asName string) (ast.Expr, *ast.LookupStage) {
	subqueryComparisonArgs := parseSubqueryComparisonArgs(f.Arg, "this")

	var initialValue *ast.Constant
	var combinatorFunc string
	if subqueryComparisonArgs.modifier == "any" {
		initialValue = falseLiteral
		combinatorFunc = "$sqlOr"
	} else {
		initialValue = trueLiteral
		combinatorFunc = "$sqlAnd"
	}

	compOp := fmt.Sprintf("$sql%v", strings.Title(strings.ToLower(subqueryComparisonArgs.op)))

	expr := ast.NewReduce(
		ast.NewFieldRef(asName, nil),
		initialValue,
		ast.NewFunction(combinatorFunc, ast.NewArray(
			ast.NewVariableRef("value"),
			ast.NewFunction(
				compOp,
				ast.NewArray(
					subqueryComparisonArgs.arg,
					subqueryComparisonArgs.outputPath,
				),
			),
		)),
	)

	lookup := subqueryComparisonArgs.toLookupStage(asName, false)

	return expr, lookup
}

// subqueryArgs represents the arguments for both $subquery and
// $subqueryExists expressions. A $subqueryExists expression does
// not contain an outputPath, but the other fields are the same.
type subqueryArgs struct {
	db         string
	coll       string
	let        []*ast.LookupLetItem
	outputPath ast.Ref
	pipeline   *ast.Pipeline
}

func (sa *subqueryArgs) toLookupStage(asName string, appendLimit bool) *ast.LookupStage {

	// $subquery and $subqueryExists benefit from appending a limit of 1 since
	// for the former only the first result is used and for the latter the
	// existence of at least one result is all that is necessary.
	if appendLimit {
		sa.pipeline.Stages = append(sa.pipeline.Stages, ast.NewLimitStage(1))
	}

	return ast.NewLookupStageWithDB(
		sa.db,
		sa.coll,
		nil,
		"",
		asName,
		sa.let,
		sa.pipeline,
	)
}

// parseSubqueryArgs parses the arguments of a $subquery or $subqueryExists
// function into a subqueryArgs.
func parseSubqueryArgs(rawArg ast.Expr, outputPathRoot string) subqueryArgs {
	args := rawArg.(*ast.Document).FieldsMap()

	parsedArgs := subqueryArgs{}

	dbArg, hasDB := args["db"]
	collArg, hasColl := args["collection"]
	letArg, hasLet := args["let"]
	outputPathArg, hasOutputPath := args["outputPath"]
	pipelineArg := args["pipeline"]

	if hasDB {
		parsedArgs.db = constantToString(dbArg)
	}

	if hasColl {
		parsedArgs.coll = constantToString(collArg)
	}

	if hasLet {
		letVars := letArg.(*ast.Document).Elements
		parsedArgs.let = make([]*ast.LookupLetItem, len(letVars))
		for i, v := range letVars {
			parsedArgs.let[i] = ast.NewLookupLetItem(v.Name, v.Expr)
		}
	}

	if hasOutputPath {
		outputPath := outputPathArg.(*ast.Array)
		parsedArgs.outputPath = ast.NewVariableRef(outputPathRoot)
		for _, outputPathPart := range outputPath.Elements {
			parsedArgs.outputPath = ast.NewFieldRef(constantToString(outputPathPart), parsedArgs.outputPath)
		}
	}

	v := parser.DeparseNode(pipelineArg)

	var err error
	parsedArgs.pipeline, err = parser.ParsePipeline(bsoncore.Array(v.Array()))
	if err != nil {
		panic(err)
	}

	return parsedArgs
}

// subqueryComparisonArgs represents the arguments for $subqueryComparison
// expressions.
type subqueryComparisonArgs struct {
	*subqueryArgs
	op       string
	modifier string
	arg      ast.Expr
}

// parseSubqueryComparisonArgs parses the arguments of a $subqueryComparison
// function into a subqueryComparisonArgs.
func parseSubqueryComparisonArgs(rawArg ast.Expr, outputPathRoot string) subqueryComparisonArgs {
	args := rawArg.(*ast.Document).FieldsMap()

	parsedArgs := subqueryComparisonArgs{}

	parsedArgs.op = constantToString(args["op"])
	parsedArgs.modifier = constantToString(args["modifier"])
	parsedArgs.arg = args["arg"]

	subqueryArgs := parseSubqueryArgs(args["subquery"], outputPathRoot)
	parsedArgs.subqueryArgs = &subqueryArgs

	return parsedArgs
}

func constantToString(expr ast.Expr) string {
	return expr.(*ast.Constant).Value.StringValue()
}

func makeExclusionStage(asNames []string) *ast.ProjectStage {
	exclusions := make([]ast.ProjectItem, len(asNames)+1)
	exclusions[0] = ast.NewExcludeProjectItem(ast.NewFieldRef("_id", nil))

	for i, asName := range asNames {
		exclusions[i+1] = ast.NewExcludeProjectItem(ast.NewFieldRef(asName, nil))
	}

	return ast.NewProjectStage(exclusions...)
}
