package util

import (
	"runtime"
)

// Constants that we may want to use in various places.
const (
	// IsWindowsOS is true if the runtime operating system is Windows.
	IsWindowsOS = runtime.GOOS == "windows"
)
