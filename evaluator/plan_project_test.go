package evaluator_test

import (
	"context"
	"testing"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/schema"

	"github.com/10gen/mongo-go-driver/bson"

	"github.com/stretchr/testify/require"
)

func TestProjectStage(t *testing.T) {
	bgCtx := context.Background()
	execCfg := createTestExecutionCfg()
	execState := evaluator.NewExecutionState()
	oCfg := createOptimizerCfg(collation.Default, execCfg)
	pCfg := createTestPushdownCfg()

	runTest := func(t *testing.T,
		projectedColumns evaluator.ProjectedColumns,
		optimize bool,
		rows []bson.D,
		expectedRows []evaluator.Values) {

		req := require.New(t)
		ts := evaluator.NewBSONSourceStage(1, tableOneName, collation.Default, rows)

		var plan evaluator.PlanStage
		var err error

		project := evaluator.NewProjectStage(ts, projectedColumns...)

		plan = project
		if optimize {
			plan, err = evaluator.OptimizePlan(context.Background(), oCfg, plan)
			req.NoError(err)
			plan, err = evaluator.PushdownPlan(pCfg, plan)
			req.False(err != nil && !evaluator.IsNonFatalPushdownError(err))
		}

		iter, err := plan.Open(bgCtx, execCfg, execState)
		req.NoError(err)

		i := 0
		row := &evaluator.Row{}

		for iter.Next(bgCtx, row) {
			req.Equal(len(row.Data), len(expectedRows[i]))
			req.Equal(row.Data, expectedRows[i])
			row = &evaluator.Row{}
			i++
		}

		req.Equal(i, len(expectedRows))

		req.NoError(iter.Close())
		req.NoError(iter.Err())
	}

	rows := bsonutil.NewDArray(
		bsonutil.NewD(bsonutil.NewDocElem("a", 6), bsonutil.NewDocElem("b", 9)),
		bsonutil.NewD(bsonutil.NewDocElem("a", 3), bsonutil.NewDocElem("b", 4)),
	)

	projectedColumns := evaluator.ProjectedColumns{
		evaluator.ProjectedColumn{
			Column: &evaluator.Column{SelectID: 1, Table: "", OriginalTable: "",
				Database: evaluator.BSONSourceDB, Name: "a", OriginalName: "a",
				MappingRegistryName: "",
				ColumnType: evaluator.ColumnType{
					EvalType: evaluator.EvalInt64, MongoType: schema.MongoInt,
				},
				PrimaryKey: false},
			Expr: evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB, tableOneName, "a",
				evaluator.EvalInt64, schema.MongoInt),
		},
		evaluator.ProjectedColumn{
			Column: &evaluator.Column{SelectID: 1, Table: "", OriginalTable: "",
				Database: evaluator.BSONSourceDB, Name: "b", OriginalName: "b",
				MappingRegistryName: "",
				ColumnType: evaluator.ColumnType{
					EvalType: evaluator.EvalInt64, MongoType: schema.MongoInt,
				},
				PrimaryKey: false},
			Expr: evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB, tableOneName, "b",
				evaluator.EvalInt64, schema.MongoInt),
		},
	}

	kind := evaluator.MySQLValueKind
	expected := []evaluator.Values{
		{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: "", Name: "a",
			Data: evaluator.NewSQLInt64(kind, 6)}, {SelectID: 1, Database: evaluator.BSONSourceDB,
			Table: "", Name: "b", Data: evaluator.NewSQLInt64(kind, 9)}},
		{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: "", Name: "a",
			Data: evaluator.NewSQLInt64(kind, 3)}, {SelectID: 1, Database: evaluator.BSONSourceDB,
			Table: "", Name: "b", Data: evaluator.NewSQLInt64(evaluator.MySQLValueKind, 4)}},
	}

	t.Run("correct results", func(t *testing.T) {
		runTest(t, projectedColumns, false, rows, expected)
	})
	t.Run("correct results after optimizations", func(t *testing.T) {
		runTest(t, projectedColumns, true, rows, expected)
	})
}
