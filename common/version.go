package common

import (
	"fmt"
	"io"
)

var (
	// VersionStr represents the version of the binaries.
	VersionStr = "v2.0.0-beta4-18-ge256b52"

	// Gitspec is the git commit hash the binaries were built from.
	Gitspec = "e256b525d757958e246c51188b8c0499621175d3"
)

// PrintVersionAndGitspec prints out the version and the gitspec.
func PrintVersionAndGitspec(toolName string, w io.Writer) {
	fmt.Fprintf(w, "%v version %v\n", toolName, VersionStr)
	fmt.Fprintf(w, "git version: %v\n", Gitspec)
}
