package evaluator

import (
	"context"
)

// UseCommand handles setting the current database.
type UseCommand struct {
	db string
}

// NewUseCommand creates a new UseCommand.
func NewUseCommand(db string) *UseCommand {
	return &UseCommand{db}
}

// Children returns a slice of all Node children of the Node.
func (UseCommand) Children() []Node {
	return []Node{}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (UseCommand) ReplaceChild(i int, e Node) {
	panicWithInvalidIndex("UseCommand", i, -1)
}

// Execute runs this command.
func (use *UseCommand) Execute(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) error {
	return cfg.commandHandler.SetDatabase(use.db)
}
