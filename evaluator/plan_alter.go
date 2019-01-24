package evaluator

import (
	"context"

	"github.com/10gen/sqlproxy/schema"
)

// AlterCommand handles altering the schema.
type AlterCommand struct {
	Alterations []*schema.Alteration
}

// Children returns a slice of all the Node children of the Node.
func (a AlterCommand) Children() []Node {
	return []Node{}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (a *AlterCommand) ReplaceChild(i int, expr Node) {
	panicWithInvalidIndex("AlterCommand", i, -1)
}

// Execute runs this command.
func (a *AlterCommand) Execute(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) error {
	return cfg.commandHandler.Alter(ctx, a.Alterations)
}
