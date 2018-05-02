package evaluator

import (
	"fmt"

	"github.com/10gen/sqlproxy/mongodb"
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

// Authorize for a ALTER command. The user must have permissions to insert in
// the sample namespace to be authorized.
func (a *AlterCommand) Authorize(ctx *ExecutionCtx) error {
	info := ctx.Variables().MongoDBInfo
	if !(info.IsAllowedSampleSource(mongodb.InsertPrivilege|mongodb.UpdatePrivilege) ||
		ctx.Server().IsAdminUser(ctx.User(), ctx.AuthenticationDatabase())) {
		return fmt.Errorf(
			"must have `insert` and `update` privileges for the " +
				"'sample source' or be admin user in order to alter tables")
	}
	return nil
}

// Execute returns an Executor for this command.
func (a *AlterCommand) Execute(ctx *ExecutionCtx) Executor {
	return &alterExecutor{
		ctx:         ctx,
		alterations: a.Alterations,
	}
}

// Run the alterExecutor.
func (a *alterExecutor) Run() error {
	schema, err := a.ctx.Server().Alter(a.ctx.Context(), a.alterations)
	if err != nil {
		return err
	}

	return a.ctx.UpdateCatalog(schema)
}
