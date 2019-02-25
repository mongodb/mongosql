package evaluator

import (
	"context"
	"fmt"

	"github.com/10gen/sqlproxy/evaluator/variable"
)

// SetCommand handles setting variables.
type SetCommand struct {
	assignments []*SQLAssignmentExpr
}

// Children returns a slice of all the Node children of the Node.
func (s SetCommand) Children() []Node {
	out := make([]Node, len(s.assignments))
	for i := range s.assignments {
		out[i] = s.assignments[i]
	}
	return out
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (s *SetCommand) ReplaceChild(i int, n Node) {
	if 0 <= i && i < len(s.assignments) {
		var ok bool
		s.assignments[i], ok = n.(*SQLAssignmentExpr)
		if !ok {
			panic(fmt.Sprintf("attempt to convert Node %v to *SQLAssignmentExpr in ReplaceChild for SetCommand failed", n))
		}
		return
	}
	panicWithInvalidIndex("SetCommand", i, len(s.assignments)-1)
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

		err = cfg.commandHandler.Set(
			variable.Name(a.variable.Name),
			a.variable.Scope,
			a.variable.Kind,
			sqlVal,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
