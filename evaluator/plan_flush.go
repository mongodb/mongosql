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
		for _, info := range f.ctx.GetStartupInfo() {
			f.ctx.Logger(log.ControlComponent).Infof(log.Always, info)
		}
	}
	return nil
}
