package evaluator_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/10gen/sqlproxy/collation"
	. "github.com/10gen/sqlproxy/evaluator"
	. "github.com/10gen/sqlproxy/evaluator/results"
	. "github.com/10gen/sqlproxy/evaluator/types"
	. "github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/schema"

	"go.mongodb.org/mongo-driver/bson"
)

func TestOrderByStage(t *testing.T) {

	bgCtx := context.Background()
	execCfg := createTestExecutionCfg(MySQLValueKind)
	execState := NewExecutionState()

	runTest := func(t *testing.T,
		terms []*OrderByTerm,
		collation *collation.Collation,
		rows []bson.D,
		expectedIds []int) {

		ts := NewBSONSourceStage(1, tableOneName, collation, rows)
		orderby := NewOrderByStage(ts, terms...)
		iter, err := orderby.Open(bgCtx, execCfg, execState)
		require.NoError(t, err)

		row := &Row{}

		i := 0

		for iter.Next(bgCtx, row) {
			require.Equal(t,
				row.Data[0].Data,
				NewSQLInt64(MySQLValueKind, int64(expectedIds[i])),
			)
			row = &Row{}
			i++
		}

		require.NoError(t, iter.Close())
		require.NoError(t, iter.Err())
	}

	t.Run("default", func(t *testing.T) {

		c := collation.Default

		data := bsonutil.NewDArray(
			bsonutil.NewD(bsonutil.NewDocElem("_id", 1), bsonutil.NewDocElem("a", "a"), bsonutil.NewDocElem("b", 7)),
			bsonutil.NewD(bsonutil.NewDocElem("_id", 2), bsonutil.NewDocElem("a", "A"), bsonutil.NewDocElem("b", 8)),
			bsonutil.NewD(bsonutil.NewDocElem("_id", 3), bsonutil.NewDocElem("a", "b"), bsonutil.NewDocElem("b", 8)),
			bsonutil.NewD(bsonutil.NewDocElem("_id", 4), bsonutil.NewDocElem("a", "B"), bsonutil.NewDocElem("b", 7)),
		)

		t.Run("single_key", func(t *testing.T) {
			t.Run("asc", func(t *testing.T) {
				terms := []*OrderByTerm{
					NewOrderByTerm(NewSQLColumnExpr(1,
						BSONSourceDB, tableOneName, "a", EvalString,
						schema.MongoString, false), true),
				}

				expected := []int{2, 4, 1, 3}
				runTest(t, terms, c, data, expected)
			})

			t.Run("desc", func(t *testing.T) {
				terms := []*OrderByTerm{
					NewOrderByTerm(NewSQLColumnExpr(1,
						BSONSourceDB, tableOneName, "a", EvalString,
						schema.MongoString, false), false),
				}

				expected := []int{3, 1, 4, 2}

				runTest(t, terms, c, data, expected)
			})
		})

		t.Run("multiple_keys", func(t *testing.T) {
			t.Run("asc_asc", func(t *testing.T) {
				terms := []*OrderByTerm{
					NewOrderByTerm(NewSQLColumnExpr(1,
						BSONSourceDB, tableOneName, "a", EvalString,
						schema.MongoString, false), true),
					NewOrderByTerm(NewSQLColumnExpr(1,
						BSONSourceDB, tableOneName, "b", EvalInt64,
						schema.MongoInt, false), false),
				}

				expected := []int{2, 4, 1, 3}

				runTest(t, terms, c, data, expected)
			})

			t.Run("asc_desc", func(t *testing.T) {
				terms := []*OrderByTerm{
					NewOrderByTerm(NewSQLColumnExpr(1,
						BSONSourceDB, tableOneName, "a", EvalString,
						schema.MongoString, false), true),
					NewOrderByTerm(NewSQLColumnExpr(1,
						BSONSourceDB, tableOneName, "b", EvalInt64,
						schema.MongoInt, false), false),
				}

				expected := []int{2, 4, 1, 3}

				runTest(t, terms, c, data, expected)
			})

			t.Run("desc_asc", func(t *testing.T) {
				terms := []*OrderByTerm{
					NewOrderByTerm(NewSQLColumnExpr(1,
						BSONSourceDB, tableOneName, "a", EvalString,
						schema.MongoString, false), false),
					NewOrderByTerm(NewSQLColumnExpr(1,
						BSONSourceDB, tableOneName, "b", EvalInt64,
						schema.MongoInt, false), true),
				}

				expected := []int{3, 1, 4, 2}

				runTest(t, terms, c, data, expected)
			})

			t.Run("desc_desc", func(t *testing.T) {
				terms := []*OrderByTerm{
					NewOrderByTerm(NewSQLColumnExpr(1,
						BSONSourceDB, tableOneName, "a", EvalString,
						schema.MongoString, false), false),
					NewOrderByTerm(NewSQLColumnExpr(1,
						BSONSourceDB, tableOneName, "b", EvalInt64,
						schema.MongoInt, false), false),
				}

				expected := []int{3, 1, 4, 2}

				runTest(t, terms, c, data, expected)
			})
		})
	})

	t.Run("utf8_general_ci", func(t *testing.T) {
		c := collation.Must(collation.Get("utf8_general_ci"))

		data := bsonutil.NewDArray(
			bsonutil.NewD(bsonutil.NewDocElem("_id", 1), bsonutil.NewDocElem("a", "a"), bsonutil.NewDocElem("b", 7)),
			bsonutil.NewD(bsonutil.NewDocElem("_id", 2), bsonutil.NewDocElem("a", "A"), bsonutil.NewDocElem("b", 8)),
			bsonutil.NewD(bsonutil.NewDocElem("_id", 3), bsonutil.NewDocElem("a", "b"), bsonutil.NewDocElem("b", 8)),
			bsonutil.NewD(bsonutil.NewDocElem("_id", 4), bsonutil.NewDocElem("a", "B"), bsonutil.NewDocElem("b", 7)),
		)

		t.Run("single_key", func(t *testing.T) {
			t.Run("asc", func(t *testing.T) {
				terms := []*OrderByTerm{
					NewOrderByTerm(NewSQLColumnExpr(1,
						BSONSourceDB, tableOneName, "a", EvalString,
						schema.MongoString, false), true),
				}

				expected := []int{1, 2, 3, 4}
				runTest(t, terms, c, data, expected)

			})

			t.Run("desc", func(t *testing.T) {
				terms := []*OrderByTerm{
					NewOrderByTerm(NewSQLColumnExpr(1,
						BSONSourceDB, tableOneName, "a", EvalString,
						schema.MongoString, false), false),
				}

				expected := []int{3, 4, 1, 2}

				runTest(t, terms, c, data, expected)
			})
		})

		t.Run("multiple_keys", func(t *testing.T) {
			t.Run("asc_asc", func(t *testing.T) {
				terms := []*OrderByTerm{
					NewOrderByTerm(NewSQLColumnExpr(1,
						BSONSourceDB, tableOneName, "a", EvalString,
						schema.MongoString, false), true),
					NewOrderByTerm(NewSQLColumnExpr(1,
						BSONSourceDB, tableOneName, "b", EvalInt64,
						schema.MongoInt, false), true),
				}

				expected := []int{1, 2, 4, 3}

				runTest(t, terms, c, data, expected)
			})

			t.Run("asc_desc", func(t *testing.T) {
				terms := []*OrderByTerm{
					NewOrderByTerm(NewSQLColumnExpr(1,
						BSONSourceDB, tableOneName, "a", EvalString,
						schema.MongoString, false), true),
					NewOrderByTerm(NewSQLColumnExpr(1,
						BSONSourceDB, tableOneName, "b", EvalInt64,
						schema.MongoInt, false), false),
				}

				expected := []int{2, 1, 3, 4}

				runTest(t, terms, c, data, expected)
			})

			t.Run("desc_asc", func(t *testing.T) {
				terms := []*OrderByTerm{
					NewOrderByTerm(NewSQLColumnExpr(1,
						BSONSourceDB, tableOneName, "a", EvalString,
						schema.MongoString, false), false),
					NewOrderByTerm(NewSQLColumnExpr(1,
						BSONSourceDB, tableOneName, "b", EvalInt64,
						schema.MongoInt, false), true),
				}

				expected := []int{4, 3, 1, 2}

				runTest(t, terms, c, data, expected)
			})

			t.Run("desc_desc", func(t *testing.T) {
				terms := []*OrderByTerm{
					NewOrderByTerm(NewSQLColumnExpr(1,
						BSONSourceDB, tableOneName, "a", EvalString,
						schema.MongoString, false), false),
					NewOrderByTerm(NewSQLColumnExpr(1,
						BSONSourceDB, tableOneName, "b", EvalInt64,
						schema.MongoInt, false), false),
				}

				expected := []int{3, 4, 2, 1}

				runTest(t, terms, c, data, expected)
			})
		})
	})
}
