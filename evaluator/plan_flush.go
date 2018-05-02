package evaluator

import (
	"fmt"

	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
)

// FlushCommand handles flushing outputs (such as logs) or reloading caches (such as schemas).
type FlushCommand struct {
	kind FlushKind
}

// FlushKind indicates the thing to be flushed.
type FlushKind string

// These are the possible values for FlushKind.
const (
	FlushLogs   = "logs"
	FlushSample = "sample"
)

// FlushExecutor executes a flush statement.
type flushExecutor struct {
	kind FlushKind
	ctx  *ExecutionCtx
}

// NewFlushCommand creates a new FlushCommand.
func NewFlushCommand(kind FlushKind) *FlushCommand {
	return &FlushCommand{kind}
}

// Authorize for a flush command. This needs to differ depending on whether
// this is a Flush Sample or Flush Logs command.
func (f *FlushCommand) Authorize(ctx *ExecutionCtx) error {
	switch f.kind {
	case FlushLogs:
		return f.authorizeFlushLogs(ctx)
	case FlushSample:
		return f.authorizeFlushSample(ctx)
	}

	return fmt.Errorf("unknown kind of flush: %v", f.kind)
}

func (*FlushCommand) authorizeFlushLogs(ctx *ExecutionCtx) error {
	if !ctx.Server().IsAdminUser(ctx.User(), ctx.AuthenticationDatabase()) {
		return fmt.Errorf("only admin user can flush logs")
	}
	return nil
}

func (f *FlushCommand) authorizeFlushSample(ctx *ExecutionCtx) error {
	info := ctx.Variables().MongoDBInfo
	// In Clustered Write Mode we check to ensure the user can write and update
	// the sample namespace. In Clustered Read mode, we already do not allow
	// flushing, so there is no reason to check here.
	if !(info.IsAllowedSampleSource(mongodb.UpdatePrivilege|mongodb.InsertPrivilege) ||
		ctx.Server().IsAdminUser(ctx.User(), ctx.AuthenticationDatabase())) {
		return fmt.Errorf("must have " +
			"`insert` and `update` privileges on " +
			"the 'sample source' or be admin user in order to flush sample")
	}
	cat := ctx.Catalog()
	// In Clustered Write Mode and Standalone Mode we ensure that the user
	// can read all collections that will be sampled. This allows a DBA to give
	// privileges to flush sample to a trusted user who is not the single admin user,
	// and the privileges make sense from the perspective that the user is allowed
	// to see all tables. In Clustered Read Mode, flushing is not allowed, but that
	// is caught in the resample implementation.
	if cat.HasAuthRestrictedNamespaces() {
		// Do not print out the namespaces the user does not have access to;
		// that would be a minor security breach when namespace names are
		// sensitive.
		return fmt.Errorf("must have " +
			"`find` privileges on the 'sample source' in order to flush sample")
	}
	return nil
}

// Execute returns an Executor for this command.
func (f *FlushCommand) Execute(ctx *ExecutionCtx) Executor {
	return &flushExecutor{
		kind: f.kind,
		ctx:  ctx,
	}
}

func (f *flushExecutor) Run() error {
	switch f.kind {
	case FlushLogs:
		return f.flushLogs()
	case FlushSample:
		return f.flushSample()
	}

	return fmt.Errorf("unknown kind of flush: %v", f.kind)
}

func (f *flushExecutor) flushLogs() error {
	f.ctx.Logger(log.ControlComponent).Infof(log.Always, "log rotation initiated")
	log.Flush()
	archive, err := log.Rotate()
	if err != nil {
		return err
	}
	if archive == "" {
		f.ctx.Logger(log.ControlComponent).Infof(log.Always,
			"rotated logs using 'reopen' strategy")
	} else {
		f.ctx.Logger(log.ControlComponent).Infof(log.Always,
			"rotated logs; old log file at %s",
			archive)
		for _, info := range f.ctx.Server().StartupInfo() {
			f.ctx.Logger(log.ControlComponent).Infof(log.Always, info)
		}
	}
	return nil
}

func (f *flushExecutor) flushSample() error {
	f.ctx.Logger(log.ControlComponent).Infof(log.Always, "sample refresh initiated")
	schema, err := f.ctx.Server().Resample(f.ctx.Context())
	if err != nil {
		return err
	}

	err = f.ctx.UpdateCatalog(schema)
	if err != nil {
		return err
	}

	f.ctx.Logger(log.ControlComponent).Infof(log.Always, "sample refresh completed")
	return nil
}
