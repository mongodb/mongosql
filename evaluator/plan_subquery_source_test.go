package evaluator_test

import (
	"context"
	"testing"

	"github.com/10gen/sqlproxy/collation"
	. "github.com/10gen/sqlproxy/evaluator"
	. "github.com/10gen/sqlproxy/evaluator/results"
	. "github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/stretchr/testify/require"

	"github.com/10gen/mongo-go-driver/bson"
)

func TestSubquerySourceStage(t *testing.T) {

	bgCtx := context.Background()
	execCfg := createTestExecutionCfg(MySQLValueKind)
	execState := NewExecutionState()
	oCfg := createOptimizerCfg(collation.Default, execCfg)
	pCfg := createTestPushdownCfg()

	runTest := func(t *testing.T, selectID int, aliasName string, optimize bool, rows []bson.D,
		expectedRows []RowValues) {
		ts := NewBSONSourceStage(1, tableOneName, collation.Default, rows)

		var plan PlanStage
		var err error

		s := NewSubquerySourceStage(ts, selectID, "", aliasName, false)
		plan = s
		if optimize {
			plan, err = OptimizePlan(context.Background(), oCfg, plan)
			require.NoError(t, err)
			plan, err = PushdownPlan(bgCtx, pCfg, plan)
			require.False(t, err != nil && !IsNonFatalPushdownError(err))
		}

		iter, err := plan.Open(bgCtx, execCfg, execState)
		require.NoError(t, err)

		i := 0
		row := &Row{}

		for iter.Next(bgCtx, row) {
			require.Equal(t, len(row.Data), len(expectedRows[i]))
			require.Equal(t, row.Data, expectedRows[i])
			row = &Row{}
			i++
		}

		require.Equal(t, i, len(expectedRows))

		require.NoError(t, iter.Close())
		require.NoError(t, iter.Err())
	}

	rows := bsonutil.NewDArray(
		bsonutil.NewD(bsonutil.NewDocElem("a", 6), bsonutil.NewDocElem("b", 9)),
		bsonutil.NewD(bsonutil.NewDocElem("a", 3), bsonutil.NewDocElem("b", 4)),
	)

	selectID := 42
	aliasName := "funny"

	kind := MySQLValueKind
	expected := []RowValues{
		{{SelectID: 42, Database: BSONSourceDB, Table: "funny", Name: "a",
			Data: NewSQLInt64(kind, 6)}, {SelectID: 42, Database: BSONSourceDB,
			Table: "funny", Name: "b", Data: NewSQLInt64(MySQLValueKind, 9)}},
		{{SelectID: 42, Database: BSONSourceDB, Table: "funny", Name: "a",
			Data: NewSQLInt64(kind, 3)}, {SelectID: 42, Database: BSONSourceDB,
			Table: "funny", Name: "b", Data: NewSQLInt64(MySQLValueKind, 4)}},
	}

	runTest(t, selectID, aliasName, false, rows, expected)

	t.Run("and should produce identical results after optimization", func(t *testing.T) {
		runTest(t, selectID, aliasName, true, rows, expected)
	})
}
