package evaluator

import "github.com/10gen/sqlproxy/log"

// FlushCommand handles flushing outputs or reloading caches
type FlushCommand struct{}

type FlushExecutor struct {
	ctx *ExecutionCtx
}

// NewFlushCommand creates a new FlushCommand.
func NewFlushCommand() *FlushCommand {
	return &FlushCommand{}
}

func (f *FlushCommand) Execute(ctx *ExecutionCtx) Executor {
	return &FlushExecutor{
		ctx: ctx,
	}
}

func (f *FlushExecutor) Run() error {
	f.ctx.Logger(log.ControlComponent).Logf(log.Always, "Log Rotation Initiated")
	archive, err := log.Rotate()
	if err == nil && archive != "" {
		f.ctx.Logger(log.ControlComponent).Logf(log.Always, "Rotated logs. Old log file at %s", archive)
	}
	for _, info := range f.ctx.GetStartupInfo() {
		f.ctx.Logger(log.ControlComponent).Logf(log.Always, info)
	}
	return err
}
