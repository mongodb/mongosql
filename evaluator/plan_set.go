package evaluator

import "github.com/10gen/sqlproxy/collation"

// SetCommand handles setting variables.
type SetCommand struct {
	assignments []*SQLAssignmentExpr
}

type SetExecutor struct {
	assignments []*SQLAssignmentExpr
	ctx         *ExecutionCtx
}

// NewSetCommand creates a new SetCommand.
func NewSetCommand(assignments []*SQLAssignmentExpr) *SetCommand {
	return &SetCommand{assignments}
}

func (s *SetCommand) Execute(ctx *ExecutionCtx) Executor {
	return &SetExecutor{s.assignments, ctx}
}

func (s *SetExecutor) Run() error {

	executorChan := make(chan error)

	var err error

	go func() {
		evalCtx := NewEvalCtx(s.ctx, collation.Default)
		for _, a := range s.assignments {
			_, pErr := a.Evaluate(evalCtx)
			if pErr != nil {
				executorChan <- pErr
			}
		}

		executorChan <- nil
	}()

	select {
	case <-s.ctx.ConnectionCtx.Tomb().Dying():
		err = s.ctx.ConnectionCtx.Tomb().Err()
	case err = <-executorChan:
	}

	return err
}
