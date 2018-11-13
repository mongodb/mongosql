package evaluator_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/collation"
	"github.com/10gen/sqlproxy/internal/util/bsonutil"
	"github.com/stretchr/testify/require"
)

var (
	_ fmt.Stringer = nil
)

func TestFilterPlanStage(t *testing.T) {

	bgCtx := context.Background()
	execCfg := createTestExecutionCfg()
	execState := evaluator.NewExecutionState()

	runTest := func(t *testing.T, matcher evaluator.SQLExpr, rows []bson.D, expectedRows []evaluator.Values) {
		req := require.New(t)

		bss := evaluator.NewBSONSourceStage(1, tableTwoName, collation.Default, rows)
		filter := evaluator.NewFilterStage(bss, matcher)

		iter, err := filter.Open(bgCtx, execCfg, execState)
		req.NoError(err)

		row := &evaluator.Row{}

		i := 0

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

	schema := evaluator.MustLoadSchema(testSchema3)

	rows := bsonutil.NewDArray(
		bsonutil.NewD(bsonutil.NewDocElem("a", 6), bsonutil.NewDocElem("b", 7), bsonutil.NewDocElem("_id", 5)),
		bsonutil.NewD(bsonutil.NewDocElem("a", 16), bsonutil.NewDocElem("b", 17), bsonutil.NewDocElem("_id", 15)),
	)

	queries := []string{
		"a = 16",
		"a = 6",
		"a = 99",
		"b > 9",
		"b > 9 or a < 5",
		"b = 7 or a = 6",
	}

	r0, err := bsonDToValues(1, evaluator.BSONSourceDB, tableTwoName, rows[0])
	require.NoError(t, err)
	r1, err := bsonDToValues(1, evaluator.BSONSourceDB, tableTwoName, rows[1])
	require.NoError(t, err)

	expected := [][]evaluator.Values{{r1}, {r0}, nil, {r1}, {r1}, {r0}}

	for i, query := range queries {
		matcher, err := evaluator.GetSQLExpr(schema, evaluator.BSONSourceDB, tableTwoName, query)
		require.NoError(t, err)

		t.Run(query, func(t *testing.T) {
			runTest(t, matcher, rows, expected[i])
		})
	}
}
