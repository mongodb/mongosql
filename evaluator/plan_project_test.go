package evaluator_test

import (
	"context"
	"testing"

	"github.com/10gen/sqlproxy/collation"
	. "github.com/10gen/sqlproxy/evaluator"
	. "github.com/10gen/sqlproxy/evaluator/types"
	. "github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/schema"

	"github.com/10gen/mongo-go-driver/bson"

	"github.com/stretchr/testify/require"
)

func TestProjectStage(t *testing.T) {
	bgCtx := context.Background()
	execCfg := createTestExecutionCfg()
	execState := NewExecutionState()
	oCfg := createOptimizerCfg(collation.Default, execCfg)
	pCfg := createTestPushdownCfg()

	runTest := func(t *testing.T,
		projectedColumns ProjectedColumns,
		optimize bool,
		rows []bson.D,
		expectedRows []Values) {

		req := require.New(t)
		ts := NewBSONSourceStage(1, tableOneName, collation.Default, rows)

		var plan PlanStage
		var err error

		project := NewProjectStage(ts, projectedColumns...)

		plan = project
		if optimize {
			plan, err = OptimizePlan(context.Background(), oCfg, plan)
			req.NoError(err)
			plan, err = PushdownPlan(pCfg, plan)
			req.False(err != nil && !IsNonFatalPushdownError(err))
		}

		iter, err := plan.Open(bgCtx, execCfg, execState)
		req.NoError(err)

		i := 0
		row := &Row{}

		for iter.Next(bgCtx, row) {
			req.Equal(len(row.Data), len(expectedRows[i]))
			req.Equal(row.Data, expectedRows[i])
			row = &Row{}
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

	projectedColumns := ProjectedColumns{
		ProjectedColumn{
			Column: &Column{SelectID: 1, Table: "", OriginalTable: "",
				Database: BSONSourceDB, Name: "a", OriginalName: "a",
				MappingRegistryName: "",
				ColumnType: ColumnType{
					EvalType: EvalInt64, MongoType: schema.MongoInt,
				},
				PrimaryKey: false},
			Expr: NewSQLColumnExpr(1, BSONSourceDB, tableOneName, "a",
				EvalInt64, schema.MongoInt),
		},
		ProjectedColumn{
			Column: &Column{SelectID: 1, Table: "", OriginalTable: "",
				Database: BSONSourceDB, Name: "b", OriginalName: "b",
				MappingRegistryName: "",
				ColumnType: ColumnType{
					EvalType: EvalInt64, MongoType: schema.MongoInt,
				},
				PrimaryKey: false},
			Expr: NewSQLColumnExpr(1, BSONSourceDB, tableOneName, "b",
				EvalInt64, schema.MongoInt),
		},
	}

	kind := MySQLValueKind
	expected := []Values{
		{{SelectID: 1, Database: BSONSourceDB, Table: "", Name: "a",
			Data: NewSQLInt64(kind, 6)}, {SelectID: 1, Database: BSONSourceDB,
			Table: "", Name: "b", Data: NewSQLInt64(kind, 9)}},
		{{SelectID: 1, Database: BSONSourceDB, Table: "", Name: "a",
			Data: NewSQLInt64(kind, 3)}, {SelectID: 1, Database: BSONSourceDB,
			Table: "", Name: "b", Data: NewSQLInt64(MySQLValueKind, 4)}},
	}

	t.Run("correct results", func(t *testing.T) {
		runTest(t, projectedColumns, false, rows, expected)
	})
	t.Run("correct results after optimizations", func(t *testing.T) {
		runTest(t, projectedColumns, true, rows, expected)
	})
}
