package common

import (
	"fmt"
	"io"
)

var (
	// VersionStr represents the version of the binaries.
	VersionStr = "v2.0.0-beta4-6-g6825f26"

	// Gitspec is the git commit hash the binaries were built from.
	Gitspec = "6825f26020aff69acf58300badcd8c06177e706a"
)

// PrintVersionAndGitspec prints out the version and the gitspec.
func PrintVersionAndGitspec(toolName string, w io.Writer) {
	fmt.Fprintf(w, "%v version %v\n", toolName, VersionStr)
	fmt.Fprintf(w, "git version: %v\n", Gitspec)
}
