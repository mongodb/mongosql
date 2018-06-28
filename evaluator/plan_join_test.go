package evaluator_test

import (
	"testing"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
	"github.com/stretchr/testify/require"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/10gen/mongo-go-driver/bson"
)

var (
	customers = []bson.D{
		{
			bson.DocElem{Name: "name", Value: "personA"},
			bson.DocElem{Name: "orderid", Value: 1},
			bson.DocElem{Name: "_id", Value: 1},
		},
		{
			bson.DocElem{Name: "name", Value: "personB"},
			bson.DocElem{Name: "orderid", Value: 2},
			bson.DocElem{Name: "_id", Value: 2},
		},
		{
			bson.DocElem{Name: "name", Value: "personC"},
			bson.DocElem{Name: "orderid", Value: 3},
			bson.DocElem{Name: "_id", Value: 3},
		},
		{
			bson.DocElem{Name: "name", Value: "personD"},
			bson.DocElem{Name: "orderid", Value: 4},
			bson.DocElem{Name: "_id", Value: 4},
		},
	}

	orders = []bson.D{
		{
			bson.DocElem{Name: "orderid", Value: 1},
			bson.DocElem{Name: "amount", Value: 1000},
			bson.DocElem{Name: "_id", Value: 1},
		},
		{
			bson.DocElem{Name: "orderid", Value: 1},
			bson.DocElem{Name: "amount", Value: 450},
			bson.DocElem{Name: "_id", Value: 2},
		},
		{
			bson.DocElem{Name: "orderid", Value: 2},
			bson.DocElem{Name: "amount", Value: 1300},
			bson.DocElem{Name: "_id", Value: 3},
		},
		{
			bson.DocElem{Name: "orderid", Value: 4},
			bson.DocElem{Name: "amount", Value: 390},
			bson.DocElem{Name: "_id", Value: 4},
		},
		{
			bson.DocElem{Name: "orderid", Value: 5},
			bson.DocElem{Name: "amount", Value: 760},
			bson.DocElem{Name: "_id", Value: 5},
		},
	}
)

func setupJoinOperator(on evaluator.SQLExpr, kind evaluator.JoinKind) evaluator.PlanStage {
	ms1 := evaluator.NewBSONSourceStage(1, tableOneName, collation.Default, customers)
	ms2 := evaluator.NewBSONSourceStage(1, tableTwoName, collation.Default, orders)

	return evaluator.NewJoinStage(kind, ms1, ms2, on)
}

func TestJoinPlanStage(t *testing.T) {

	testSchema := evaluator.MustLoadSchema(testSchema4)
	testInfo := evaluator.GetMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)

	Convey("Subject: JoinStage", t, func() {
		criteria := evaluator.NewSQLEqualsExpr(
			evaluator.NewSQLColumnExpr(
				1,
				evaluator.BSONSourceDB,
				tableOneName,
				"orderid",
				evaluator.EvalInt64,
				schema.MongoInt,
			),
			evaluator.NewSQLColumnExpr(
				1,
				evaluator.BSONSourceDB,
				tableTwoName,
				"orderid",
				evaluator.EvalInt64,
				schema.MongoInt,
			),
		)

		ctx := createTestExecutionCtx(testInfo)

		row := &evaluator.Row{}

		i := 0

		Convey("an inner join should return correct results", func() {

			operator := setupJoinOperator(criteria, evaluator.InnerJoin)

			iter, err := operator.Open(ctx)
			So(err, ShouldBeNil)

			expectedResults := []struct {
				Name   interface{}
				Amount interface{}
			}{
				{"personA", 1000},
				{"personA", 450},
				{"personB", 1300},
				{"personD", 390},
			}

			for iter.Next(row) {
				So(len(row.Data), ShouldEqual, 6)
				So(row.Data[0].Table, ShouldEqual, tableOneName)
				So(row.Data[4].Table, ShouldEqual, tableTwoName)
				So(row.Data[0].Data, ShouldEqual, expectedResults[i].Name)
				So(row.Data[4].Data, ShouldEqual, expectedResults[i].Amount)
				i++
			}

			So(i, ShouldEqual, 4)

			So(iter.Close(), ShouldBeNil)
			So(iter.Err(), ShouldBeNil)

		})

		Convey("a left join should return correct results", func() {

			operator := setupJoinOperator(criteria, evaluator.LeftJoin)

			iter, err := operator.Open(ctx)
			So(err, ShouldBeNil)

			expectedResults := []struct {
				Name   interface{}
				Amount interface{}
			}{
				{"personA", 1000},
				{"personA", 450},
				{"personB", 1300},
				{"personC", nil},
				{"personD", 390},
			}

			for iter.Next(row) {
				So(len(row.Data), ShouldEqual, 6)
				So(row.Data[0].Table, ShouldEqual, tableOneName)
				So(row.Data[4].Table, ShouldEqual, tableTwoName)
				So(row.Data[0].Data, ShouldEqual, expectedResults[i].Name)
				if expectedResults[i].Amount == nil {
					So(row.Data[4].Data, ShouldHaveSameTypeAs, evaluator.SQLNull)
				} else {
					So(row.Data[4].Data, ShouldEqual, expectedResults[i].Amount)
				}
				i++
			}

			So(i, ShouldEqual, 5)

			So(iter.Close(), ShouldBeNil)
			So(iter.Err(), ShouldBeNil)

		})

		Convey("a right join should return correct results", func() {

			operator := setupJoinOperator(criteria, evaluator.RightJoin)

			iter, err := operator.Open(ctx)
			So(err, ShouldBeNil)

			expectedResults := []struct {
				Name   interface{}
				Amount interface{}
			}{
				{"personA", 1000},
				{"personA", 450},
				{"personB", 1300},
				{"personD", 390},
				{nil, 760},
			}
			for iter.Next(row) {
				So(len(row.Data), ShouldEqual, 6)
				So(row.Data[0].Table, ShouldEqual, tableOneName)
				So(row.Data[4].Table, ShouldEqual, tableTwoName)
				if expectedResults[i].Name == nil {
					So(row.Data[0].Data, ShouldHaveSameTypeAs, evaluator.SQLNull)
				} else {
					So(row.Data[0].Data, ShouldEqual, expectedResults[i].Name)
				}
				So(row.Data[4].Data, ShouldEqual, expectedResults[i].Amount)
				i++
			}

			So(i, ShouldEqual, 5)

			So(iter.Close(), ShouldBeNil)
			So(iter.Err(), ShouldBeNil)

		})

		Convey("a cross join should return correct results", func() {

			operator := setupJoinOperator(criteria, evaluator.CrossJoin)

			iter, err := operator.Open(ctx)
			So(err, ShouldBeNil)

			expectedNames := []string{"personA", "personB", "personC", "personD", "personE"}
			expectedAmounts := []int{1000, 450, 1300, 390, 760}

			for iter.Next(row) {
				So(len(row.Data), ShouldEqual, 6)
				So(row.Data[0].Table, ShouldEqual, tableOneName)
				So(row.Data[4].Table, ShouldEqual, tableTwoName)
				So(row.Data[0].Data, ShouldEqual, expectedNames[i/5])
				So(row.Data[4].Data, ShouldEqual, expectedAmounts[i%5])
				i++
			}

			So(i, ShouldEqual, 20)

			So(iter.Close(), ShouldBeNil)
			So(iter.Err(), ShouldBeNil)

		})

		Convey("a straight join should return correct results", func() {

			operator := setupJoinOperator(criteria, evaluator.StraightJoin)

			iter, err := operator.Open(ctx)
			So(err, ShouldBeNil)

			expectedResults := []struct {
				Name   interface{}
				Amount interface{}
			}{
				{"personA", 1000},
				{"personA", 450},
				{"personB", 1300},
				{"personD", 390},
			}

			for iter.Next(row) {
				So(len(row.Data), ShouldEqual, 6)
				So(row.Data[0].Table, ShouldEqual, tableOneName)
				So(row.Data[4].Table, ShouldEqual, tableTwoName)
				So(row.Data[0].Data, ShouldEqual, expectedResults[i].Name)
				So(row.Data[4].Data, ShouldEqual, expectedResults[i].Amount)
				i++
			}

			So(i, ShouldEqual, 4)

			So(iter.Close(), ShouldBeNil)
			So(iter.Err(), ShouldBeNil)

		})
	})
}

func TestJoinPlanStage_MemoryLimits(t *testing.T) {

	testSchema := evaluator.MustLoadSchema(testSchema4)
	testInfo := evaluator.GetMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)

	criteria := evaluator.NewSQLEqualsExpr(
		evaluator.NewSQLColumnExpr(
			1,
			evaluator.BSONSourceDB,
			tableOneName,
			"orderid",
			evaluator.EvalInt64,
			schema.MongoInt,
		),
		evaluator.NewSQLColumnExpr(
			1,
			evaluator.BSONSourceDB,
			tableTwoName,
			"orderid",
			evaluator.EvalInt64,
			schema.MongoInt,
		),
	)

	row := &evaluator.Row{}

	t.Run("inner join", func(t *testing.T) {
		ctx := createTestExecutionCtx(testInfo)
		ctx.Variables().SetSystemVariable(variable.MongoDBMaxStageSize, 500)

		operator := setupJoinOperator(criteria, evaluator.InnerJoin)

		iter, err := operator.Open(ctx)
		require.NoError(t, err)

		ok := iter.Next(row)
		require.False(t, ok)

		require.NoError(t, iter.Close())
		require.Error(t, iter.Err())
	})

	t.Run("left join", func(t *testing.T) {
		ctx := createTestExecutionCtx(testInfo)
		ctx.Variables().SetSystemVariable(variable.MongoDBMaxStageSize, 500)

		operator := setupJoinOperator(criteria, evaluator.LeftJoin)

		iter, err := operator.Open(ctx)
		require.NoError(t, err)

		ok := iter.Next(row)
		require.False(t, ok)

		require.NoError(t, iter.Close())
		require.Error(t, iter.Err())
	})

	t.Run("right join", func(t *testing.T) {
		ctx := createTestExecutionCtx(testInfo)
		ctx.Variables().SetSystemVariable(variable.MongoDBMaxStageSize, 500)

		operator := setupJoinOperator(criteria, evaluator.RightJoin)

		iter, err := operator.Open(ctx)
		require.NoError(t, err)

		ok := iter.Next(row)
		require.False(t, ok)

		require.NoError(t, iter.Close())
		require.Error(t, iter.Err())
	})

	t.Run("cross join", func(t *testing.T) {

		ctx := createTestExecutionCtx(testInfo)
		ctx.Variables().SetSystemVariable(variable.MongoDBMaxStageSize, 500)
		operator := setupJoinOperator(nil, evaluator.RightJoin)

		iter, err := operator.Open(ctx)
		require.NoError(t, err)

		ok := iter.Next(row)
		require.False(t, ok)

		require.NoError(t, iter.Close())
		require.Error(t, iter.Err())
	})
}

func TestJoinStageMemoryMonitor(t *testing.T) {
	criteria := evaluator.NewSQLEqualsExpr(
		evaluator.NewSQLColumnExpr(
			1,
			evaluator.BSONSourceDB,
			tableOneName,
			"orderid",
			evaluator.EvalInt64,
			schema.MongoInt,
		),
		evaluator.NewSQLColumnExpr(
			1,
			evaluator.BSONSourceDB,
			tableTwoName,
			"orderid",
			evaluator.EvalInt64,
			schema.MongoInt,
		),
	)

	leftSize := valueSize(evaluator.BSONSourceDB,
		tableOneName,
		"name",
		evaluator.SQLVarchar("personA"),
	) + valueSize(evaluator.BSONSourceDB, tableOneName, "orderid", evaluator.SQLInt64(0)) +
		valueSize(evaluator.BSONSourceDB, tableOneName, "_id", evaluator.SQLInt64(0))

	rightSize := valueSize(evaluator.BSONSourceDB, tableTwoName, "orderid", evaluator.SQLInt64(0)) +
		valueSize(evaluator.BSONSourceDB, tableTwoName, "amount", evaluator.SQLInt64(0)) +
		valueSize(evaluator.BSONSourceDB, tableTwoName, "_id", evaluator.SQLInt64(0))

	t.Run("inner join", func(t *testing.T) {
		operator := setupJoinOperator(criteria, evaluator.InnerJoin)

		actual := getAllocatedMemorySizeAfterIteration(operator)
		expected := (leftSize + rightSize) * 4

		require.Equal(t, expected, actual)
	})

	t.Run("left join", func(t *testing.T) {
		operator := setupJoinOperator(criteria, evaluator.LeftJoin)

		actual := getAllocatedMemorySizeAfterIteration(operator)
		expected := (leftSize+rightSize)*4 +
			(leftSize + rightSize - 24) // nil right values

		require.Equal(t, expected, actual)
	})

	t.Run("right join", func(t *testing.T) {
		operator := setupJoinOperator(criteria, evaluator.RightJoin)

		actual := getAllocatedMemorySizeAfterIteration(operator)
		expected := (leftSize+rightSize)*4 +
			(leftSize - 23 + rightSize) // nil left values

		require.Equal(t, expected, actual)
	})

	t.Run("cross join", func(t *testing.T) {
		operator := setupJoinOperator(nil, evaluator.CrossJoin)

		actual := getAllocatedMemorySizeAfterIteration(operator)
		expected := uint64(len(customers)) * uint64(len(orders)) * (leftSize + rightSize)

		require.Equal(t, expected, actual)
	})
}
