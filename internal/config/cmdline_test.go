package config_test

import (
	"fmt"
	"runtime"
	"testing"

	. "github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/log"
)

func TestParseArgs_Valid(t *testing.T) {
	cfg := Default()
	args := []string{
		// Client Connection
		"--auth",
		"--addr", "host:3306",
		"--defaultAuthMechanism", "GSSAPI",
		"--defaultAuthSource", "$external",
		"--sslMode", "requireSSL",
		"--sslAllowInvalidCertificates",
		"--sslCAFile", "cafile",
		"--sslPEMKeyFile", "pemkeyfile",
		"--sslPEMKeyPassword", "pemkeypassword",

		// Log
		"--logAppend",
		"--logRotate", "reopen",
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
		"--maxVarcharLength", "1000",
		"--sampleNamespaces", "foo.*",
		"--sampleNamespaces", "*.bar",
		"--sampleSize", "500",
		"--sampleMode", "read",
		"--sampleReadIntervalSecs", "1005",
		"--sampleWriteIntervalSecs", "983",
		"--uuidSubtype3Encoding", "java",

		// Service
		"--serviceName", "oompa",
		"--serviceDisplayName", "loompa",
		"--serviceDescription", "doompa tee do",
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

	testBool(t, cfg.SystemLog.LogAppend, true, "cfg.SystemLog.LogAppend")
	testString(t, string(cfg.SystemLog.LogRotate), string(log.Reopen), "cfg.SystemLog.LogRotate")
	testString(t, cfg.SystemLog.Path, "temp", "cfg.SystemLog.Quiet")
	testBool(t, cfg.SystemLog.Quiet, true, "cfg.SystemLog.Quiet")
	testInt(t, cfg.SystemLog.Verbosity, 2, "cfg.SystemLog.Verbosity")

	testString(t, cfg.Schema.Path, "path-to-file", "cfg.Schema.Path")
	testUint16(t, cfg.Schema.MaxVarcharLength, 1000, "cfg.Schema.MaxVarcharLength")

	testStringSlice(t, cfg.Schema.Sample.Namespaces, []string{"foo.*", "*.bar"}, "cfg.Schema.Sample.Namespaces")
	testInt64(t, cfg.Schema.Sample.Size, 500, "cfg.Schema.Sample.Size")
	testString(t, cfg.Schema.Sample.Mode, "read", "cfg.Schema.Sample.Mode")
	testInt64(t, cfg.Schema.Sample.ReadIntervalSecs, 1005, "cfg.Schema.Sample.ReadIntervalSecs")
	testInt64(t, cfg.Schema.Sample.WriteIntervalSecs, 983, "cfg.Schema.Sample.WriteIFntervalSecs")
	testString(t, cfg.Schema.Sample.UUIDSubtype3Encoding, "java", "cfg.Schema.Sample.UUIDSubtype3Encoding")

	testStringSlice(t, cfg.Net.BindIP, []string{"host"}, "cfg.Net.BindIP")
	testInt(t, cfg.Net.Port, 3306, "cfg.Net.Port")
	if runtime.GOOS != "windows" {
		testBool(t, cfg.Net.UnixDomainSocket.Enabled, false, "cfg.Net.UnixDomainSocket.Enabled")
		testString(t, cfg.Net.UnixDomainSocket.PathPrefix, "/var", "cfg.Net.UnixDomainSocket.PathPrefix")
		testString(t, cfg.Net.UnixDomainSocket.FilePermissions, "0600", "cfg.Net.UnixDomainSocket.FilePermissions")
	}

	testString(t, cfg.Net.SSL.Mode, "requireSSL", "cfg.Net.SSL.Mode")
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

	testString(t, cfg.ProcessManagement.Service.Name, "oompa", "cfg.ProcessManagement.Service.Name")
	testString(t, cfg.ProcessManagement.Service.DisplayName, "loompa", "cfg.ProcessManagement.Service.DisplayName")
	testString(t, cfg.ProcessManagement.Service.Description, "doompa tee do", "cfg.ProcessManagement.Service.Description")
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

	testStringSlice(t, cfg.Net.BindIP, []string{"host"}, "cfg.Net.BindIP")
	testInt(t, cfg.Net.Port, 3307, "cfg.Net.Port")
	testString(t, cfg.Schema.Path, "path-to-directory", "cfg.Schema.Path")

}

func TestParseArgs_Invalid(t *testing.T) {
	var shortOptDelim, longOptDelim string
	if runtime.GOOS == "windows" {
		shortOptDelim, longOptDelim = "/", "/"
	} else {
		shortOptDelim = "-"
		longOptDelim = "--"
	}

	var tests = []struct {
		err  string
		args []string
	}{
		{err: "", args: []string{"--addr", "sdffewg:2134:12344"}},
		{err: "must specify only one of --schema or --schemaDirectory", args: []string{"--schema", "file", "--schemaDirectory", "dir"}},
		{err: "error parsing command line options: Unexpected argument(s): [unexpected args]", args: []string{"unexpected", "args"}},
		{err: "error parsing command line options: invalid argument for flag `" + shortOptDelim + "v, " + longOptDelim + "verbose' " +
			"(expected int): invalid verbosity value given",
			args: []string{"--verbose=silly"}},
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

func TestVerbosity_Valid(t *testing.T) {
	var tests = []struct {
		level int
		args  []string
	}{
		{level: 0, args: []string{}},
		{level: 1, args: []string{"-v"}},
		{level: 1, args: []string{"--verbose"}},
		{level: 2, args: []string{"-vv"}},
		{level: 3, args: []string{"--verbose=3"}},
		{level: 3, args: []string{"-v=3"}},
		{level: 4, args: []string{"-v", "4"}},
		{level: 4, args: []string{"--verbose", "4"}},

		{level: 1, args: []string{"--verbose", "4", "-v"}},
		{level: 3, args: []string{"-v=2", "-vvv"}},
		{level: 4, args: []string{"--verbose=3", "-v", "4"}},
		{level: 4, args: []string{"-vvv", "--verbose", "4"}},
		{level: 5, args: []string{"--verbose", "4", "-v=5"}},
		{level: 5, args: []string{"-v", "4", "--verbose", "5"}},
	}

	for _, test := range tests {
		cfg := Default()
		ParseArgs(cfg, test.args)
		testInt(t, cfg.SystemLog.Verbosity, test.level, "cfg.SystemLog.Verbosity")
	}
}
