package evaluator

import (
	"github.com/10gen/sqlproxy/schema"
)

// AlterCommand handles altering the schema.
type AlterCommand struct {
	Alterations []*schema.Alteration
}

type alterExecutor struct {
	ctx         *ExecutionCtx
	alterations []*schema.Alteration
}

// Execute returns an Executor for this command.
func (a *AlterCommand) Execute(ctx *ExecutionCtx) Executor {
	return &alterExecutor{
		ctx:         ctx,
		alterations: a.Alterations,
	}
}

func (a *alterExecutor) Run() error {
	schema, err := a.ctx.Server().Alter(a.ctx.Context(), a.alterations)
	if err != nil {
		return err
	}

	return a.ctx.UpdateCatalog(schema)
}
