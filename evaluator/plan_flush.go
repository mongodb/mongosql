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
	archive, err := log.Rotate()
	if err == nil {
		log.Logf(log.Always, "Rotated logs. Old log file at %s", archive)
	}
	return err
}
