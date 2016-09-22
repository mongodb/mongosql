package util

import (
	"runtime"
)

const (
	IsWindowsOS = runtime.GOOS == "windows"
)
