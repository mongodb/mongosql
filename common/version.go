package common

import (
	"fmt"
	"io"
)

var (
	// VersionStr represents the version of the binaries.
	VersionStr = "developer build"

	// Gitspec is the git commit-ish the binaries were built from.
	Gitspec = "not-built-with-ldflags"
)

// PrintVersionAndGitspec prints out the version and the gitspec.
func PrintVersionAndGitspec(toolName string, w io.Writer) {
	fmt.Fprintf(w, "%v version %v\n", toolName, VersionStr)
	fmt.Fprintf(w, "git version: %v\n", Gitspec)
}
