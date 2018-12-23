package evaluator

import (
	"context"

	"github.com/10gen/sqlproxy/evaluator/variable"
)

// SetCommand handles setting variables.
type SetCommand struct {
	assignments []*SQLAssignmentExpr
}

// NewSetCommand creates a new SetCommand.
func NewSetCommand(assignments []*SQLAssignmentExpr) *SetCommand {
	return &SetCommand{assignments}
}

// Execute runs this command.
func (s *SetCommand) Execute(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) error {
	var maxScope variable.Scope
	for _, a := range s.assignments {
		if maxScope == 0 {
			maxScope = a.variable.Scope
		}
		if a.variable.Scope == variable.GlobalScope {
			maxScope = variable.GlobalScope
		}
	}

	// Check that we are authorized to make all assignments before we
	// start to execute them.
	err := cfg.commandHandler.SetScopeAuthorized(maxScope)
	if err != nil {
		return err
	}

	for _, a := range s.assignments {
		sqlVal, err := a.Evaluate(ctx, cfg, st)
		if err != nil {
			return err
		}

		var literal interface{}
		if !sqlVal.IsNull() {
			literal = sqlVal.Value()
		}

		err = cfg.commandHandler.Set(
			variable.Name(a.variable.Name),
			a.variable.Scope,
			a.variable.Kind,
			literal,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
