package evaluator_test

import (
	"context"
	"testing"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/schema"
	"github.com/stretchr/testify/require"
)

func TestEmptyOperator(t *testing.T) {
	req := require.New(t)

	bgCtx := context.Background()
	execCfg := createTestExecutionCfg()
	execState := evaluator.NewExecutionState()

	columns := []*evaluator.Column{
		{
			Table: "foo",
			Name:  "a",
			ColumnType: evaluator.ColumnType{
				EvalType:  evaluator.EvalInt64,
				MongoType: schema.MongoInt,
			},
		},
	}

	e := evaluator.NewEmptyStage(columns, collation.Default)

	iter, err := e.Open(bgCtx, execCfg, execState)
	req.NoError(err)

	req.False(iter.Next(bgCtx, nil), "Next() should always return false")

	res := e.Columns()
	req.Len(res, 1)

	col := res[0]
	req.Equal(col.Table, "foo")
	req.Equal(col.Name, "a")
	req.Equal(col.EvalType, evaluator.EvalInt64)
	req.Equal(col.MongoType, schema.MongoInt)

	req.NoError(iter.Close())
	req.NoError(iter.Err())

}
