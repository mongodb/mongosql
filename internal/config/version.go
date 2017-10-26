package config

import (
	"fmt"
	"io"
)

var (
	// VersionStr represents the version of the binaries.
	VersionStr = "v2.3.0-beta3-12-g65e8f2d5"

	// Gitspec is the git commit hash the binaries were built from.
	Gitspec = "65e8f2d5dbbc108511d79d91693150444edd8ca1"
)

// PrintVersionAndGitspec prints out the version and the gitspec.
func PrintVersionAndGitspec(toolName string, w io.Writer) {
	fmt.Fprintf(w, "%v version %v\n", toolName, VersionStr)
	fmt.Fprintf(w, "git version: %v\n", Gitspec)
}
