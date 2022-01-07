package config

import (
	"fmt"
	"io"

	"github.com/10gen/openssl"
)

var (
	// VersionStr represents the version of the binaries.
	VersionStr = "built-without-version-string"

	// Gitspec is the git commit hash the binaries were built from.
	Gitspec = "built-without-git-spec"
)

// PrintVersionInfo prints out the version, gitspec, runtime openssl
// version, and build-time openssl version.
func PrintVersionInfo(toolName string, w io.Writer) {
	_, _ = fmt.Fprintf(w, "%v version: %v\n", toolName, VersionStr)
	_, _ = fmt.Fprintf(w, "git version: %v\n", Gitspec)
	_, _ = fmt.Fprintf(w, "openssl version: %v\n", openssl.Version)
	_, _ = fmt.Fprintf(w, "openssl build version: %v\n", openssl.BuildVersion)
}
