package evaluator

import (
	"fmt"

	"github.com/10gen/sqlproxy/log"
)

// FlushCommand handles flushing outputs (such as logs) or reloading caches (such as schemas).
type FlushCommand struct {
	kind FlushKind
}

// FlushKind indicates the thing to be flushed.
type FlushKind string

// FlushKind constants
const (
	FlushLogs   = "logs"
	FlushSample = "sample"
)

// FlushExecutor executes a flush statement.
type FlushExecutor struct {
	kind FlushKind
	ctx  *ExecutionCtx
}

// NewFlushCommand creates a new FlushCommand.
func NewFlushCommand(kind FlushKind) *FlushCommand {
	return &FlushCommand{kind}
}

func (f *FlushCommand) Execute(ctx *ExecutionCtx) Executor {
	return &FlushExecutor{
		kind: f.kind,
		ctx:  ctx,
	}
}

func (f *FlushExecutor) Run() error {
	switch f.kind {
	case FlushLogs:
		return f.flushLogs()
	case FlushSample:
		return f.flushSample()
	}

	return fmt.Errorf("unknown kind of flush: %v", f.kind)
}

func (f *FlushExecutor) flushLogs() error {
	f.ctx.Logger(log.ControlComponent).Infof(log.Always, "log rotation initiated")
	log.Flush()
	archive, err := log.Rotate()
	if err != nil {
		return err
	}
	if archive == "" {
		f.ctx.Logger(log.ControlComponent).Infof(log.Always, "rotated logs using 'reopen' strategy")
	} else {
		f.ctx.Logger(log.ControlComponent).Infof(log.Always, "rotated logs; old log file at %s", archive)
		for _, info := range f.ctx.Server().StartupInfo() {
			f.ctx.Logger(log.ControlComponent).Infof(log.Always, info)
		}
	}
	return nil
}

func (f *FlushExecutor) flushSample() error {
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
