package evaluator

import (
	"fmt"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/internal/util"
)

// SetCommand handles setting variables.
type SetCommand struct {
	assignments []*SQLAssignmentExpr
}

type setExecutor struct {
	assignments []*SQLAssignmentExpr
	ctx         *ExecutionCtx
}

// NewSetCommand creates a new SetCommand.
func NewSetCommand(assignments []*SQLAssignmentExpr) *SetCommand {
	return &SetCommand{assignments}
}

// Execute returns an Executor for this command.
func (s *SetCommand) Execute(ctx *ExecutionCtx) Executor {
	return &setExecutor{s.assignments, ctx}
}

func (s *setExecutor) Run() error {

	executorChan := make(chan error, 1)

	var err error

	util.PanicSafeGo(func() {
		evalCtx := NewEvalCtx(s.ctx, collation.Default)
		for _, a := range s.assignments {
			_, pErr := a.Evaluate(evalCtx)
			if pErr != nil {
				executorChan <- pErr
			}
		}

		executorChan <- nil
	}, func(err interface{}) {
		executorChan <- fmt.Errorf("%v", err)
	})

	select {
	case <-s.ctx.ConnectionCtx.Context().Done():
		err = s.ctx.ConnectionCtx.Context().Err()
	case err = <-executorChan:
	}

	return err
}
