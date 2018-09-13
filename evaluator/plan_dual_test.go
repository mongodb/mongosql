package evaluator_test

import (
	"context"
	"testing"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/stretchr/testify/require"
)

func TestDualOperator(t *testing.T) {
	req := require.New(t)

	bgCtx := context.Background()
	execCfg := createTestExecutionCfg()
	execState := evaluator.NewExecutionState()

	// dual operators should only ever return one row with no data

	operator := &evaluator.DualStage{}

	iter, err := operator.Open(bgCtx, execCfg, execState)
	req.NoError(err, "failed to open DualStage")

	row := &evaluator.Row{}

	i := 0
	for iter.Next(bgCtx, row) {
		req.Len(row.Data, 0, "row should have no data")
		i++
	}

	req.Equal(i, 1, "iter should only return one row")

	req.NoError(iter.Close(), "got error while closing iter")
	req.NoError(iter.Err(), "iterator had an error")
}
