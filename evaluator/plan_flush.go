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

// Children returns a slice of all the Node children of the Node.
func (FlushCommand) Children() []Node {
	return []Node{}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (FlushCommand) ReplaceChild(i int, n Node) {
	panicWithInvalidIndex("FlushCommand", i, -1)
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
