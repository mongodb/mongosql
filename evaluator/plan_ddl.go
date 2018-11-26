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

// Execute runs this command.
func (tbl *DropCommand) Execute(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) error {
	return cfg.commandHandler.Drop(tbl.tableName)
}
