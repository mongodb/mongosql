package util

import (
	"errors"
)

// Exit codes to which we have ascribed special meaning
const (
	// ExitError indicates that the process exited in error.
	ExitError int = 1
	// ExitClean indicates that the process exited cleanly.
	ExitClean int = 0
	// ExitClean indicates that the process exited because of
	// invalid configuration options.
	ExitBadOptions int = 3
	// ExitClean indicates that the process exited as a result
	// of being killed.
	ExitKill int = 4
	// Go reserves exit code 2 for its own use.
)

// Errors to use along with some of our custom error codes
var (
	// ErrTerminated indicates that a termination signal was received.
	ErrTerminated = errors.New("received termination signal")
)
