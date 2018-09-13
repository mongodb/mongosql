package evaluator_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/collation"
	"github.com/10gen/sqlproxy/schema"
)

func TestOrderByStage(t *testing.T) {

	bgCtx := context.Background()
	execCfg := createTestExecutionCfg()
	execState := evaluator.NewExecutionState()

	runTest := func(t *testing.T,
		terms []*evaluator.OrderByTerm,
		collation *collation.Collation,
		rows []bson.D,
		expectedIds []int) {

		ts := evaluator.NewBSONSourceStage(1, tableOneName, collation, rows)
		orderby := evaluator.NewOrderByStage(ts, terms...)
		iter, err := orderby.Open(bgCtx, execCfg, execState)
		require.NoError(t, err)

		row := &evaluator.Row{}

		i := 0

		for iter.Next(bgCtx, row) {
			require.Equal(t,
				row.Data[0].Data,
				evaluator.NewSQLInt64(evaluator.MySQLValueKind, int64(expectedIds[i])),
			)
			row = &evaluator.Row{}
			i++
		}

		require.NoError(t, iter.Close())
		require.NoError(t, iter.Err())
	}

	t.Run("default", func(t *testing.T) {

		c := collation.Default

		data := []bson.D{
			{{Name: "_id", Value: 1}, {Name: "a", Value: "a"}, {Name: "b", Value: 7}},
			{{Name: "_id", Value: 2}, {Name: "a", Value: "A"}, {Name: "b", Value: 8}},
			{{Name: "_id", Value: 3}, {Name: "a", Value: "b"}, {Name: "b", Value: 8}},
			{{Name: "_id", Value: 4}, {Name: "a", Value: "B"}, {Name: "b", Value: 7}},
		}

		t.Run("single_key", func(t *testing.T) {
			t.Run("asc", func(t *testing.T) {
				terms := []*evaluator.OrderByTerm{
					evaluator.NewOrderByTerm(evaluator.NewSQLColumnExpr(1,
						evaluator.BSONSourceDB, tableOneName, "a", evaluator.EvalString,
						schema.MongoString), true),
				}

				expected := []int{2, 4, 1, 3}
				runTest(t, terms, c, data, expected)
			})

			t.Run("desc", func(t *testing.T) {
				terms := []*evaluator.OrderByTerm{
					evaluator.NewOrderByTerm(evaluator.NewSQLColumnExpr(1,
						evaluator.BSONSourceDB, tableOneName, "a", evaluator.EvalString,
						schema.MongoString), false),
				}

				expected := []int{3, 1, 4, 2}

				runTest(t, terms, c, data, expected)
			})
		})

		t.Run("multiple_keys", func(t *testing.T) {
			t.Run("asc_asc", func(t *testing.T) {
				terms := []*evaluator.OrderByTerm{
					evaluator.NewOrderByTerm(evaluator.NewSQLColumnExpr(1,
						evaluator.BSONSourceDB, tableOneName, "a", evaluator.EvalString,
						schema.MongoString), true),
					evaluator.NewOrderByTerm(evaluator.NewSQLColumnExpr(1,
						evaluator.BSONSourceDB, tableOneName, "b", evaluator.EvalInt64,
						schema.MongoInt), false),
				}

				expected := []int{2, 4, 1, 3}

				runTest(t, terms, c, data, expected)
			})

			t.Run("asc_desc", func(t *testing.T) {
				terms := []*evaluator.OrderByTerm{
					evaluator.NewOrderByTerm(evaluator.NewSQLColumnExpr(1,
						evaluator.BSONSourceDB, tableOneName, "a", evaluator.EvalString,
						schema.MongoString), true),
					evaluator.NewOrderByTerm(evaluator.NewSQLColumnExpr(1,
						evaluator.BSONSourceDB, tableOneName, "b", evaluator.EvalInt64,
						schema.MongoInt), false),
				}

				expected := []int{2, 4, 1, 3}

				runTest(t, terms, c, data, expected)
			})

			t.Run("desc_asc", func(t *testing.T) {
				terms := []*evaluator.OrderByTerm{
					evaluator.NewOrderByTerm(evaluator.NewSQLColumnExpr(1,
						evaluator.BSONSourceDB, tableOneName, "a", evaluator.EvalString,
						schema.MongoString), false),
					evaluator.NewOrderByTerm(evaluator.NewSQLColumnExpr(1,
						evaluator.BSONSourceDB, tableOneName, "b", evaluator.EvalInt64,
						schema.MongoInt), true),
				}

				expected := []int{3, 1, 4, 2}

				runTest(t, terms, c, data, expected)
			})

			t.Run("desc_desc", func(t *testing.T) {
				terms := []*evaluator.OrderByTerm{
					evaluator.NewOrderByTerm(evaluator.NewSQLColumnExpr(1,
						evaluator.BSONSourceDB, tableOneName, "a", evaluator.EvalString,
						schema.MongoString), false),
					evaluator.NewOrderByTerm(evaluator.NewSQLColumnExpr(1,
						evaluator.BSONSourceDB, tableOneName, "b", evaluator.EvalInt64,
						schema.MongoInt), false),
				}

				expected := []int{3, 1, 4, 2}

				runTest(t, terms, c, data, expected)
			})
		})
	})

	t.Run("utf8_general_ci", func(t *testing.T) {
		c := collation.Must(collation.Get("utf8_general_ci"))

		data := []bson.D{
			{{Name: "_id", Value: 1}, {Name: "a", Value: "a"}, {Name: "b", Value: 7}},
			{{Name: "_id", Value: 2}, {Name: "a", Value: "A"}, {Name: "b", Value: 8}},
			{{Name: "_id", Value: 3}, {Name: "a", Value: "b"}, {Name: "b", Value: 8}},
			{{Name: "_id", Value: 4}, {Name: "a", Value: "B"}, {Name: "b", Value: 7}},
		}

		t.Run("single_key", func(t *testing.T) {
			t.Run("asc", func(t *testing.T) {
				terms := []*evaluator.OrderByTerm{
					evaluator.NewOrderByTerm(evaluator.NewSQLColumnExpr(1,
						evaluator.BSONSourceDB, tableOneName, "a", evaluator.EvalString,
						schema.MongoString), true),
				}

				expected := []int{1, 2, 3, 4}
				runTest(t, terms, c, data, expected)

			})

			t.Run("desc", func(t *testing.T) {
				terms := []*evaluator.OrderByTerm{
					evaluator.NewOrderByTerm(evaluator.NewSQLColumnExpr(1,
						evaluator.BSONSourceDB, tableOneName, "a", evaluator.EvalString,
						schema.MongoString), false),
				}

				expected := []int{3, 4, 1, 2}

				runTest(t, terms, c, data, expected)
			})
		})

		t.Run("multiple_keys", func(t *testing.T) {
			t.Run("asc_asc", func(t *testing.T) {
				terms := []*evaluator.OrderByTerm{
					evaluator.NewOrderByTerm(evaluator.NewSQLColumnExpr(1,
						evaluator.BSONSourceDB, tableOneName, "a", evaluator.EvalString,
						schema.MongoString), true),
					evaluator.NewOrderByTerm(evaluator.NewSQLColumnExpr(1,
						evaluator.BSONSourceDB, tableOneName, "b", evaluator.EvalInt64,
						schema.MongoInt), true),
				}

				expected := []int{1, 2, 4, 3}

				runTest(t, terms, c, data, expected)
			})

			t.Run("asc_desc", func(t *testing.T) {
				terms := []*evaluator.OrderByTerm{
					evaluator.NewOrderByTerm(evaluator.NewSQLColumnExpr(1,
						evaluator.BSONSourceDB, tableOneName, "a", evaluator.EvalString,
						schema.MongoString), true),
					evaluator.NewOrderByTerm(evaluator.NewSQLColumnExpr(1,
						evaluator.BSONSourceDB, tableOneName, "b", evaluator.EvalInt64,
						schema.MongoInt), false),
				}

				expected := []int{2, 1, 3, 4}

				runTest(t, terms, c, data, expected)
			})

			t.Run("desc_asc", func(t *testing.T) {
				terms := []*evaluator.OrderByTerm{
					evaluator.NewOrderByTerm(evaluator.NewSQLColumnExpr(1,
						evaluator.BSONSourceDB, tableOneName, "a", evaluator.EvalString,
						schema.MongoString), false),
					evaluator.NewOrderByTerm(evaluator.NewSQLColumnExpr(1,
						evaluator.BSONSourceDB, tableOneName, "b", evaluator.EvalInt64,
						schema.MongoInt), true),
				}

				expected := []int{4, 3, 1, 2}

				runTest(t, terms, c, data, expected)
			})

			t.Run("desc_desc", func(t *testing.T) {
				terms := []*evaluator.OrderByTerm{
					evaluator.NewOrderByTerm(evaluator.NewSQLColumnExpr(1,
						evaluator.BSONSourceDB, tableOneName, "a", evaluator.EvalString,
						schema.MongoString), false),
					evaluator.NewOrderByTerm(evaluator.NewSQLColumnExpr(1,
						evaluator.BSONSourceDB, tableOneName, "b", evaluator.EvalInt64,
						schema.MongoInt), false),
				}

				expected := []int{3, 4, 2, 1}

				runTest(t, terms, c, data, expected)
			})
		})
	})
}
