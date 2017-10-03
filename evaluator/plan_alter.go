package evaluator

import (
	"github.com/10gen/sqlproxy/schema"
)

type AlterCommand struct {
	Alterations []*schema.Alteration
}

type AlterExecutor struct {
	ctx         *ExecutionCtx
	alterations []*schema.Alteration
}

func (a *AlterCommand) Execute(ctx *ExecutionCtx) Executor {
	return &AlterExecutor{
		ctx:         ctx,
		alterations: a.Alterations,
	}
}

func (a *AlterExecutor) Run() error {
	schema, err := a.ctx.Server().Alter(a.ctx.Context(), a.alterations)
	if err != nil {
		return err
	}

	return a.ctx.UpdateCatalog(schema)
}
