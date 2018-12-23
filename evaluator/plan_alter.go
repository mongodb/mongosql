package evaluator

import (
	"context"

	"github.com/10gen/sqlproxy/internal/schema"
)

// AlterCommand handles altering the schema.
type AlterCommand struct {
	Alterations []*schema.Alteration
}

// Execute runs this command.
func (a *AlterCommand) Execute(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) error {
	return cfg.commandHandler.Alter(ctx, a.Alterations)
}
