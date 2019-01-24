package evaluator

import (
	"context"
)

// DropCommand handles a Drop Table command.
type DropCommand struct {
	tableName string
}

// NewDropCommand creates a new DropCommand
func NewDropCommand(table string) *DropCommand {
	return &DropCommand{table}
}

// Children returns a slice of all the Node children of the Node.
func (DropCommand) Children() []Node {
	return []Node{}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (DropCommand) ReplaceChild(i int, n Node) {
	panicWithInvalidIndex("DropCommand", i, -1)
}

// Execute runs this command.
func (tbl *DropCommand) Execute(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) error {
	return cfg.commandHandler.Drop(tbl.tableName)
}
