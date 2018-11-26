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

// Execute runs this command.
func (use *UseCommand) Execute(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) error {
	return cfg.commandHandler.SetDatabase(use.db)
}
