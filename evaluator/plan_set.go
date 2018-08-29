package evaluator

import (
	"fmt"

	"github.com/10gen/sqlproxy/internal/collation"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/internal/variable"
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

// Authorize authorizes a user to execute this Set Command.
// A user can only set Global variables if they are the admin user.
// If all the variables are Session scope, it does not matter if the
// user is the Admin user or not.
func (s *SetCommand) Authorize(ctx *ExecutionCtx) error {
	isAdminUser := ctx.Server().IsAdminUser(ctx.User(), ctx.AuthenticationDatabase())
	for _, a := range s.assignments {
		if a.variable.Scope == variable.GlobalScope && !isAdminUser {
			return fmt.Errorf("only admin user can set global variables")
		}
	}
	return nil
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
