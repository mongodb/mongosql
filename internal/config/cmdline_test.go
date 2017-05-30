package config_test

import (
	"fmt"
	"runtime"
	"testing"

	. "github.com/10gen/sqlproxy/internal/config"
)

func TestParseArgs_Valid(t *testing.T) {
	cfg := Default()
	args := []string{
		// Client Connection
		"--auth",
		"--addr", "host:3306",
		"--defaultAuthMechanism", "GSSAPI",
		"--defaultAuthSource", "$external",
		"--sslAllowInvalidCertificates",
		"--sslCAFile", "cafile",
		"--sslPEMKeyFile", "pemkeyfile",
		"--sslPEMKeyPassword", "pemkeypassword",

		// Log
		"--logAppend",
		"--logPath", "temp",
		"--quiet",
		"-vv",

		// Mongo Connection
		"--mongo-ssl",
		"--mongo-sslAllowInvalidCertificates",
		"--mongo-sslAllowInvalidHostnames",
		"--mongo-sslCAFile", "mongocafile",
		"--mongo-sslCRLFile", "mongocrlfile",
		"--mongo-sslFIPSMode",
		"--mongo-sslPEMKeyFile", "mongopemkeyfile",
		"--mongo-sslPEMKeyPassword", "mongopemkeypassword",
		"--mongo-uri", "mongodb://hostname:27018",
		"--mongo-versionCompatibility", "3.2",

		// Schema
		"--schema", "path-to-file",
	}
	if runtime.GOOS != "windows" {
		args = append(args, []string{
			// Socket
			"--filePermissions", "0600",
			"--noUnixSocket",
			"--unixSocketPrefix", "/var",
		}...)
	}

	err := ParseArgs(cfg, args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testBool(t, cfg.SystemLog.LogAppend, true, "cfg.SystemLog.Append")
	testString(t, cfg.SystemLog.Path, "temp", "cfg.SystemLog.Quiet")
	testBool(t, cfg.SystemLog.Quiet, true, "cfg.SystemLog.Quiet")
	testInt(t, cfg.SystemLog.Verbosity, 2, "cfg.SystemLog.Verbosity")

	testString(t, cfg.Schema.Path, "path-to-file", "cfg.Schema.Path")

	testString(t, cfg.Net.BindIP, "host", "cfg.Net.BindIP")
	testInt(t, cfg.Net.Port, 3306, "cfg.Net.Port")
	if runtime.GOOS != "windows" {
		testBool(t, cfg.Net.UnixDomainSocket.Enabled, false, "cfg.Net.UnixDomainSocket.Enabled")
		testString(t, cfg.Net.UnixDomainSocket.PathPrefix, "/var", "cfg.Net.UnixDomainSocket.PathPrefix")
		testString(t, cfg.Net.UnixDomainSocket.FilePermissions, "0600", "cfg.Net.UnixDomainSocket.FilePermissions")
	}

	testBool(t, cfg.Net.SSL.AllowInvalidCertificates, true, "cfg.Net.SSL.AllowInvalidCertificates")
	testString(t, cfg.Net.SSL.PEMKeyFile, "pemkeyfile", "cfg.Net.SSL.PEMKeyFile")
	testString(t, cfg.Net.SSL.PEMKeyPassword, "pemkeypassword", "cfg.Net.SSL.PEMKeyPassword")
	testString(t, cfg.Net.SSL.CAFile, "cafile", "cfg.Net.SSL.CAFile")

	testBool(t, cfg.Security.Enabled, true, "cfg.Security.Enabled")
	testString(t, cfg.Security.DefaultMechanism, "GSSAPI", "cfg.Security.DefaultMechanism")
	testString(t, cfg.Security.DefaultSource, "$external", "cfg.Security.DefaultSource")

	testString(t, cfg.MongoDB.VersionCompatibility, "3.2", "cfg.MongoDB.VersionCompatibility")
	testString(t, cfg.MongoDB.Net.URI, "mongodb://hostname:27018", "cfg.MongoDB.Net.URI")

	testBool(t, cfg.MongoDB.Net.SSL.Enabled, true, "cfg.MongoDB.Net.SSL.Enabled")
	testBool(t, cfg.MongoDB.Net.SSL.AllowInvalidCertificates, true, "cfg.MongoDB.Net.SSL.AllowInvalidCertificates")
	testBool(t, cfg.MongoDB.Net.SSL.AllowInvalidHostnames, true, "cfg.MongoDB.Net.SSL.AllowInvalidHostnames")
	testString(t, cfg.MongoDB.Net.SSL.PEMKeyFile, "mongopemkeyfile", "cfg.MongoDB.Net.SSL.PEMKeyFile")
	testString(t, cfg.MongoDB.Net.SSL.PEMKeyPassword, "mongopemkeypassword", "cfg.MongoDB.Net.SSL.PEMKeyPassword")
	testString(t, cfg.MongoDB.Net.SSL.CAFile, "mongocafile", "cfg.MongoDB.Net.SSL.CAFile")
	testString(t, cfg.MongoDB.Net.SSL.CRLFile, "mongocrlfile", "cfg.MongoDB.Net.SSL.CRLFile")
	testBool(t, cfg.MongoDB.Net.SSL.FIPSMode, true, "cfg.MongoDB.Net.SSL.FIPSMode")
}

func TestParseArgs_Valid2(t *testing.T) {
	cfg := Default()
	args := []string{
		// Client Connection
		"--auth",
		"--addr", "host",

		// Schema
		"--schemaDirectory", "path-to-directory",
	}

	err := ParseArgs(cfg, args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testString(t, cfg.Net.BindIP, "host", "cfg.Net.BindIP")
	testInt(t, cfg.Net.Port, 3307, "cfg.Net.Port")
	testString(t, cfg.Schema.Path, "path-to-directory", "cfg.Schema.Path")

}

func TestParseArgs_Invalid(t *testing.T) {

	var tests = []struct {
		err  string
		args []string
	}{
		{err: "", args: []string{"--addr", "sdffewg:2134:12344"}},
		{err: "must specify only one of --schema or --schemaDirectory", args: []string{"--schema", "file", "--schemaDirectory", "dir"}},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%v-%v", i, test.err), func(t *testing.T) {
			cfg := Default()
			err := ParseArgs(cfg, test.args)
			if err == nil {
				t.Fatalf("expected error, but got none")
			}

			if test.err != "" && err.Error() != test.err {
				t.Fatalf("expected err '%s' but got '%v'", test.err, err)
			}
		})
	}
}
