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
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/require"
)

var (
	customers = bsonutil.NewDArray(
		bsonutil.NewD(
			bsonutil.NewDocElem("name", "personA"),
			bsonutil.NewDocElem("orderid", 1),
			bsonutil.NewDocElem("_id", 1),
		),
		bsonutil.NewD(
			bsonutil.NewDocElem("name", "personB"),
			bsonutil.NewDocElem("orderid", 2),
			bsonutil.NewDocElem("_id", 2),
		),
		bsonutil.NewD(
			bsonutil.NewDocElem("name", "personC"),
			bsonutil.NewDocElem("orderid", 3),
			bsonutil.NewDocElem("_id", 3),
		),
		bsonutil.NewD(
			bsonutil.NewDocElem("name", "personD"),
			bsonutil.NewDocElem("orderid", 4),
			bsonutil.NewDocElem("_id", 4),
		),
	)

	orders = bsonutil.NewDArray(
		bsonutil.NewD(
			bsonutil.NewDocElem("orderid", 1),
			bsonutil.NewDocElem("amount", 1000),
			bsonutil.NewDocElem("_id", 1),
		),
		bsonutil.NewD(
			bsonutil.NewDocElem("orderid", 1),
			bsonutil.NewDocElem("amount", 450),
			bsonutil.NewDocElem("_id", 2),
		),
		bsonutil.NewD(
			bsonutil.NewDocElem("orderid", 2),
			bsonutil.NewDocElem("amount", 1300),
			bsonutil.NewDocElem("_id", 3),
		),
		bsonutil.NewD(
			bsonutil.NewDocElem("orderid", 4),
			bsonutil.NewDocElem("amount", 390),
			bsonutil.NewDocElem("_id", 4),
		),
		bsonutil.NewD(
			bsonutil.NewDocElem("orderid", 5),
			bsonutil.NewDocElem("amount", 760),
			bsonutil.NewDocElem("_id", 5),
		),
	)
)

func setupJoinOperator(on SQLExpr, kind JoinKind) PlanStage {
	ms1 := NewBSONSourceStage(1, tableOneName, collation.Default, customers)
	ms2 := NewBSONSourceStage(1, tableTwoName, collation.Default, orders)

	return NewJoinStage(kind, ms1, ms2, on)
}

func TestJoinPlanStage(t *testing.T) {

	criteria := NewSQLEqualsExpr(
		NewSQLColumnExpr(
			1,
			BSONSourceDB,
			tableOneName,
			"orderid",
			EvalInt64,
			schema.MongoInt,
		),
		NewSQLColumnExpr(
			1,
			BSONSourceDB,
			tableTwoName,
			"orderid",
			EvalInt64,
			schema.MongoInt,
		),
	)

	bgCtx := context.Background()
	execCfg := createTestExecutionCfg()
	execState := NewExecutionState()

	row := &Row{}

	t.Run("inner_join", func(t *testing.T) {
		req := require.New(t)

		operator := setupJoinOperator(criteria, InnerJoin)

		iter, err := operator.Open(bgCtx, execCfg, execState)
		req.NoError(err)

		expectedResults := []struct {
			Name   interface{}
			Amount int64
		}{
			{"personA", 1000},
			{"personA", 450},
			{"personB", 1300},
			{"personD", 390},
		}

		i := 0
		for iter.Next(bgCtx, row) {
			req.Len(row.Data, 6)
			req.Equal(row.Data[0].Table, tableOneName)
			req.Equal(row.Data[4].Table, tableTwoName)
			req.Equal(row.Data[0].Data.Value(), expectedResults[i].Name)
			req.Equal(row.Data[4].Data.Value(), expectedResults[i].Amount)
			i++
		}

		req.Equal(i, 4)

		req.NoError(iter.Close())
		req.NoError(iter.Err())
	})

	t.Run("left_join", func(t *testing.T) {
		req := require.New(t)

		operator := setupJoinOperator(criteria, LeftJoin)

		iter, err := operator.Open(bgCtx, execCfg, execState)
		req.NoError(err)

		expectedResults := []struct {
			Name   interface{}
			Amount interface{}
		}{
			{"personA", int64(1000)},
			{"personA", int64(450)},
			{"personB", int64(1300)},
			{"personC", nil},
			{"personD", int64(390)},
		}

		i := 0
		for iter.Next(bgCtx, row) {
			req.Len(row.Data, 6)
			req.Equal(row.Data[0].Table, tableOneName)
			req.Equal(row.Data[4].Table, tableTwoName)
			req.Equal(row.Data[0].Data.Value(), expectedResults[i].Name)
			if expectedResults[i].Amount == nil {
				req.Zero(convey.ShouldHaveSameTypeAs(
					row.Data[4].Data,
					NewSQLNull(MySQLValueKind),
				))
			} else {
				req.Equal(row.Data[4].Data.Value(), expectedResults[i].Amount)
			}
			i++
		}

		req.Equal(i, 5)

		req.NoError(iter.Close())
		req.NoError(iter.Err())
	})

	t.Run("right_join", func(t *testing.T) {
		req := require.New(t)

		operator := setupJoinOperator(criteria, RightJoin)

		iter, err := operator.Open(bgCtx, execCfg, execState)
		req.NoError(err)

		expectedResults := []struct {
			Name   interface{}
			Amount int64
		}{
			{"personA", 1000},
			{"personA", 450},
			{"personB", 1300},
			{"personD", 390},
			{nil, 760},
		}

		i := 0
		for iter.Next(bgCtx, row) {
			req.Len(row.Data, 6)
			req.Equal(row.Data[0].Table, tableOneName)
			req.Equal(row.Data[4].Table, tableTwoName)
			if expectedResults[i].Name == nil {
				req.Zero(convey.ShouldHaveSameTypeAs(
					row.Data[0].Data,
					NewSQLNull(MySQLValueKind),
				))
			} else {
				req.Equal(row.Data[0].Data.Value(), expectedResults[i].Name)
			}
			req.Equal(row.Data[4].Data.Value(), expectedResults[i].Amount)
			i++
		}

		req.Equal(i, 5)

		req.NoError(iter.Close())
		req.NoError(iter.Err())
	})

	t.Run("cross_join", func(t *testing.T) {
		req := require.New(t)

		operator := setupJoinOperator(criteria, CrossJoin)

		iter, err := operator.Open(bgCtx, execCfg, execState)
		req.NoError(err)

		expectedNames := []string{"personA", "personB", "personC", "personD", "personE"}
		expectedAmounts := []int64{1000, 450, 1300, 390, 760}

		i := 0
		for iter.Next(bgCtx, row) {
			req.Len(row.Data, 6)
			req.Equal(row.Data[0].Table, tableOneName)
			req.Equal(row.Data[4].Table, tableTwoName)
			req.Equal(row.Data[0].Data.Value(), expectedNames[i/5])
			req.Equal(row.Data[4].Data.Value(), expectedAmounts[i%5])
			i++
		}

		req.Equal(i, 20)

		req.NoError(iter.Close())
		req.NoError(iter.Err())
	})

	t.Run("straight_join", func(t *testing.T) {
		req := require.New(t)

		operator := setupJoinOperator(criteria, StraightJoin)

		iter, err := operator.Open(bgCtx, execCfg, execState)
		req.NoError(err)

		expectedResults := []struct {
			Name   interface{}
			Amount int64
		}{
			{"personA", 1000},
			{"personA", 450},
			{"personB", 1300},
			{"personD", 390},
		}

		i := 0
		for iter.Next(bgCtx, row) {
			req.Len(row.Data, 6)
			req.Equal(row.Data[0].Table, tableOneName)
			req.Equal(row.Data[4].Table, tableTwoName)
			req.Equal(row.Data[0].Data.Value(), expectedResults[i].Name)
			req.Equal(row.Data[4].Data.Value(), expectedResults[i].Amount)
			i++
		}

		req.Equal(i, 4)

		req.NoError(iter.Close())
		req.NoError(iter.Err())
	})

}
