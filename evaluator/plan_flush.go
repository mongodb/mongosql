package evaluator

import (
	"context"
	"fmt"
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

// NewFlushCommand creates a new FlushCommand.
func NewFlushCommand(kind FlushKind) *FlushCommand {
	return &FlushCommand{kind}
}

// Execute runs this command.
func (f *FlushCommand) Execute(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) error {
	switch f.kind {
	case FlushLogs:
		return cfg.commandHandler.RotateLogs()
	case FlushSample:
		return cfg.commandHandler.Resample(ctx)
	default:
		return fmt.Errorf("unknown kind of flush: %v", f.kind)
	}
}
