package desugarer

import (
	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/parser"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

const uniqueAsName = "eca58228-b657-498a-b76e-f48a9161a404"

// desugarJoins desugars any $join stages in the pipeline into a sequence
// of equivalent, existing MQL stages.
func desugarJoins(pipeline *ast.Pipeline, _ uint64) *ast.Pipeline {

	// A $join stage is desugared into 4 stages, so the new slice of stages
	// starts with capacity 4 times the length of the pipeline. (The "worst"
	// case is the pipeline consists entirely of $join stages.)
	newStages := make([]ast.Stage, 0, len(pipeline.Stages)*4)

	for _, stage := range pipeline.Stages {
		switch s := stage.(type) {
		case *ast.Unknown:
			if s.StageName() == "$join" {
				newStages = append(newStages, desugarJoinStage(s)...)
				continue
			}
		}

		// this is not a $join, put back in pipeline
		newStages = append(newStages, stage)
	}

	return ast.NewPipeline(newStages...)
}

// desugarJoinStage desugars a $join stage into a sequence of 4 stages that
// are semantically equivalent to $join. Specifically, the desugaring turns
//
//   { $join: {
//       database: <db name>,
//       collection: <coll name>,
//       joinType: <join type>,
//       let: <let doc>,
//       pipeline: <pipeline>,
//       condition: <condition> // note: this is a $match stage
//   }}
//
// into:
//
//   { $lookup: {
//       from: {
//           db: <db name>,
//           coll: <coll name>
//       },
//       let: <let doc>,
//       pipeline: [<pipeline>..., <condition>]
//       as: "<uuid>"
//   }},
//   { $unwind: {
//       path: "$<uuid>",
//       preserveNullAndEmptyArrays: <join type> == "left"
//   }},
//   { $replaceRoot: {
//       newRoot: {
//           $mergeObjects: ["$$ROOT", "$<uuid>"]
//       }
//   }},
//   { $project: { _id: 0, <uuid>: 0 } }
//
// Note that in the $unwind stage, preserveNullAndEmptyArrays is true when
// the joinType is "left", and false when it is "inner".
//
// Also note that the database, collection, let, and condition fields are
// all optional.
func desugarJoinStage(stage *ast.Unknown) []ast.Stage {
	stageDoc := stage.Value.Document().Lookup("$join").Document()

	type joinStage struct {
		Database   *string           `bson:"database"`
		Collection *string           `bson:"collection"`
		JoinType   string            `bson:"joinType"`
		Let        bsoncore.Document `bson:"let"`
		Pipeline   bsoncore.Array    `bson:"pipeline"`
		Condition  bsoncore.Document `bson:"condition"`
	}

	join := joinStage{}
	err := bson.Unmarshal(stageDoc, &join)
	if err != nil {
		panic(err)
	}

	collection := ""
	if join.Collection != nil {
		collection = *join.Collection
	}

	var let []*ast.LookupLetItem
	if join.Let != nil {
		letElements, err := join.Let.Elements()
		if err != nil {
			panic(err)
		}

		let = make([]*ast.LookupLetItem, len(letElements))
		for i, element := range letElements {
			expr, err := parser.ParseExpr(element.Value())
			if err != nil {
				panic(err)
			}
			let[i] = ast.NewLookupLetItem(element.Key(), expr)
		}
	}

	pipeline, err := parser.ParsePipeline(join.Pipeline)
	if err != nil {
		panic(err)
	}

	if join.Condition != nil {
		conditionStage, err := parser.ParseStage(join.Condition)
		if err != nil {
			panic(err)
		}

		pipeline.Stages = append(pipeline.Stages, conditionStage)
	}

	var lookupStage *ast.LookupStage
	if join.Database != nil {
		lookupStage = ast.NewLookupStageWithDB(
			*join.Database,
			collection,
			nil,
			"",
			uniqueAsName,
			let,
			pipeline,
		)
	} else {
		lookupStage = ast.NewLookupStage(
			collection,
			nil,
			"",
			uniqueAsName,
			let,
			pipeline,
		)
	}

	asFieldRef := ast.NewFieldRef(uniqueAsName, nil)

	unwindStage := ast.NewUnwindStage(
		asFieldRef,
		"",
		join.JoinType == "left",
	)

	replaceRootStage := ast.NewReplaceRootStage(
		ast.NewFunction("$mergeObjects", ast.NewArray(
			ast.NewVariableRef("ROOT"),
			asFieldRef,
		)),
	)

	projectStage := ast.NewProjectStage(
		ast.NewExcludeProjectItem(ast.NewFieldRef("_id", nil)),
		ast.NewExcludeProjectItem(asFieldRef),
	)

	return []ast.Stage{lookupStage, unwindStage, replaceRootStage, projectStage}
}
