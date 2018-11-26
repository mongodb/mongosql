package config_test

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	"path/filepath"

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
		"--gssapiHostname", "something",
		"--gssapiServiceName", "awesome",
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
		"--mongo-gssapiServiceName", "hola",
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

		// Metrics
		"--stitch-url", "https://mystitchapp.com",

		// Schema
		"--schema", "path-to-file",
		"--maxVarcharLength", "1000",
		"--sampleNamespaces", "foo.*",
		"--sampleNamespaces", "*.bar",
		"--sampleSize", "500",
		"--sampleMode", "write",
		"--sampleRefreshIntervalSecs", "983",
		"--uuidSubtype3Encoding", "java",
		"--prejoin",

		// Service
		"--serviceName", "oompa",
		"--serviceDisplayName", "loompa",
		"--serviceDescription", "doompa tee do",

		// SetParameter
		"--setParameter", "enableTableAlterations=true",
		"--setParameter", "metrics_backend=stitch",
		"--setParameter", "optimize_cross_joins=false",
		"--setParameter", "optimize_evaluations=false",
		"--setParameter", "optimize_filtering=false",
		"--setParameter", "optimize_inner_joins=false",
		"--setParameter", "optimize_self_joins=false",
		"--setParameter", "pushdown=false",
		"--setParameter", "optimize_view_sampling=false",
		"--setParameter", "polymorphic_type_conversion_mode=fast",
		"--setParameter", "type_conversion_mode=mysql",

		// Debug
		"--enableProfiling", "cpu",
		"--profileScope", "all",
	}
	if runtime.GOOS != "windows" {
		args = append(args, []string{
			// Socket
			"--filePermissions", "0600",
			"--noUnixSocket",
			"--unixSocketPrefix", "/var",
		}...)
	}

	_, err := ParseArgs(cfg, args)
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

	testStringSlice(
		t,
		cfg.Schema.Sample.Namespaces,
		[]string{"foo.*", "*.bar"},
		"cfg.Schema.Sample.Namespaces",
	)
	testInt64(t, cfg.Schema.Sample.Size, 500, "cfg.Schema.Sample.Size")
	testSampleMode(t, cfg.Schema.Sample.Mode, "write", "cfg.Schema.Sample.Mode")
	testBool(t, cfg.Schema.Sample.PreJoin, true, "cfg.Schema.Sample.PreJoin")
	testBool(t,
		cfg.Schema.Sample.OptimizeViewSampling,
		true,
		"cfg.Schema.Sample.OptimizeViewSampling",
	)
	testInt64(
		t,
		cfg.Schema.Sample.RefreshIntervalSecs,
		983,
		"cfg.Schema.Sample.RefreshIntervalSecs",
	)
	testString(
		t,
		cfg.Schema.Sample.UUIDSubtype3Encoding,
		"java",
		"cfg.Schema.Sample.UUIDSubtype3Encoding",
	)

	testStringSlice(t, cfg.Net.BindIP, []string{"host"}, "cfg.Net.BindIP")
	testInt(t, cfg.Net.Port, 3306, "cfg.Net.Port")
	if runtime.GOOS != "windows" {
		testBool(t, cfg.Net.UnixDomainSocket.Enabled, false, "cfg.Net.UnixDomainSocket.Enabled")
		testString(
			t,
			cfg.Net.UnixDomainSocket.PathPrefix,
			"/var",
			"cfg.Net.UnixDomainSocket.PathPrefix",
		)
		testString(
			t,
			cfg.Net.UnixDomainSocket.FilePermissions,
			"0600",
			"cfg.Net.UnixDomainSocket.FilePermissions",
		)
	}

	testString(t, cfg.Net.SSL.Mode, "requireSSL", "cfg.Net.SSL.Mode")
	testBool(t, cfg.Net.SSL.AllowInvalidCertificates, true, "cfg.Net.SSL.AllowInvalidCertificates")
	testString(t, cfg.Net.SSL.PEMKeyFile, "pemkeyfile", "cfg.Net.SSL.PEMKeyFile")
	testString(t, cfg.Net.SSL.PEMKeyPassword, "pemkeypassword", "cfg.Net.SSL.PEMKeyPassword")
	testString(t, cfg.Net.SSL.CAFile, "cafile", "cfg.Net.SSL.CAFile")

	testBool(t, cfg.Security.Enabled, true, "cfg.Security.Enabled")
	testString(t, cfg.Security.DefaultMechanism, "GSSAPI", "cfg.Security.DefaultMechanism")
	testString(t, cfg.Security.DefaultSource, "$external", "cfg.Security.DefaultSource")
	testString(t, cfg.Security.GSSAPI.Hostname, "something", "cfg.Security.GSSAPI.Hostname")
	testString(t, cfg.Security.GSSAPI.ServiceName, "awesome", "cfg.Security.GSSAPI.ServiceName")

	testString(t, cfg.MongoDB.VersionCompatibility, "3.2", "cfg.MongoDB.VersionCompatibility")
	testString(t, cfg.MongoDB.Net.URI, "mongodb://hostname:27018", "cfg.MongoDB.Net.URI")

	testString(
		t,
		cfg.MongoDB.Net.Auth.GSSAPIServiceName,
		"hola",
		"cfg.MongoDB.Net.Auth.GSSAPIServiceName",
	)

	testBool(t, cfg.MongoDB.Net.SSL.Enabled, true, "cfg.MongoDB.Net.SSL.Enabled")
	testBool(t,
		cfg.MongoDB.Net.SSL.AllowInvalidCertificates,
		true,
		"cfg.MongoDB.Net.SSL.AllowInvalidCertificates",
	)
	testBool(t,
		cfg.MongoDB.Net.SSL.AllowInvalidHostnames,
		true,
		"cfg.MongoDB.Net.SSL.AllowInvalidHostnames",
	)
	testString(t,
		cfg.MongoDB.Net.SSL.PEMKeyFile,
		"mongopemkeyfile",
		"cfg.MongoDB.Net.SSL.PEMKeyFile",
	)
	testString(t,
		cfg.MongoDB.Net.SSL.PEMKeyPassword,
		"mongopemkeypassword",
		"cfg.MongoDB.Net.SSL.PEMKeyPassword",
	)
	testString(t, cfg.MongoDB.Net.SSL.CAFile, "mongocafile", "cfg.MongoDB.Net.SSL.CAFile")
	testString(t, cfg.MongoDB.Net.SSL.CRLFile, "mongocrlfile", "cfg.MongoDB.Net.SSL.CRLFile")
	testBool(t, cfg.MongoDB.Net.SSL.FIPSMode, true, "cfg.MongoDB.Net.SSL.FIPSMode")

	testString(t, cfg.Metrics.StitchURL, "https://mystitchapp.com", "cfg.Metrics.StitchURL")

	testString(t, cfg.ProcessManagement.Service.Name, "oompa", "cfg.ProcessManagement.Service.Name")
	testString(t,
		cfg.ProcessManagement.Service.DisplayName,
		"loompa",
		"cfg.ProcessManagement.Service.DisplayName",
	)
	testString(t,
		cfg.ProcessManagement.Service.Description,
		"doompa tee do",
		"cfg.ProcessManagement.Service.Description",
	)

	testBool(t, cfg.SetParameter.EnableTableAlterations, true, "cfg.SetParameter.EnableTableAlterations")
	testString(t, cfg.SetParameter.MetricsBackend, "stitch", "cfg.SetParameter.MetricsBackend")
	testBool(t, cfg.SetParameter.OptimizeCrossJoins, false, "cfg.SetParameter.OptimizeCrossJoins")
	testBool(t, cfg.SetParameter.OptimizeEvaluations, false, "cfg.SetParameter.OptimizeEvaluations")
	testBool(t, cfg.SetParameter.OptimizeFiltering, false, "cfg.SetParameter.OptimizeFiltering")
	testBool(t, cfg.SetParameter.OptimizeInnerJoins, false, "cfg.SetParameter.OptimizeInnerJoins")
	testBool(t, cfg.SetParameter.OptimizeSelfJoins, false, "cfg.SetParameter.OptimizeSelfJoins")
	testBool(t, cfg.SetParameter.OptimizeViewSampling, false, "cfg.SetParameter.OptimizeViewSampling")
	testString(t, cfg.SetParameter.PolymorphicTypeConversionMode, "fast", "cfg.SetParameter.PolymorphicTypeConversionMode")
	testBool(t, cfg.SetParameter.Pushdown, false, "cfg.SetParameter.Pushdown")
	testString(t, cfg.SetParameter.TypeConversionMode, "mysql", "cfg.SetParameter.TypeConversionMode")

	testString(t, cfg.Debug.EnableProfiling, "cpu", "cfg.Debug.EnableProfiling")
	testString(t, cfg.Debug.ProfileScope, "all", "cfg.Debug.ProfileScope")
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

	_, err := ParseArgs(cfg, args)
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
		{
			err:  "must specify only one of --schema or --schemaDirectory",
			args: []string{"--schema", "file", "--schemaDirectory", "dir"},
		},
		{
			err:  "error parsing command line options: Unexpected argument(s): [unexpected args]",
			args: []string{"unexpected", "args"},
		},
		{
			err: "error parsing command line options: invalid argument for flag `" +
				shortOptDelim + "v, " +
				longOptDelim + "verbose' " +
				"(expected int): invalid verbosity value given",
			args: []string{"--verbose=silly"},
		},
		{err: "invalid setParameter key: foo", args: []string{"--setParameter", "foo=bar"}},
		{
			err:  "invalid value for setParameter enableTableAlterations: bar",
			args: []string{"--setParameter", "enableTableAlterations=bar"},
		},
		{
			err: "error parsing command line options: invalid argument for flag `" +
				longOptDelim + "setParameter' (expected <param>=<value>): " +
				"invalid setParameter expression: enableTableAlterations",
			args: []string{"--setParameter", "enableTableAlterations"},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%v-%v", i, test.err), func(t *testing.T) {
			cfg := Default()
			_, err := ParseArgs(cfg, test.args)
			if err == nil {
				t.Fatalf("expected error, but got none")
			}

			if test.err != "" && err.Error() != test.err {
				t.Fatalf("expected err '%s' but got '%v'", test.err, err)
			}
		})
	}
}

func TestCapturePositionalArgs_Valid(t *testing.T) {
	var tests = []struct {
		expected []string
		args     []string
	}{
		{expected: []string{}, args: []string{}},
		{expected: []string{"-v"}, args: []string{"-v"}},
		{expected: []string{"--verbose"}, args: []string{"--verbose"}},
		{expected: []string{"-vv"}, args: []string{"-vv"}},
		{expected: []string{"--verbose=3"}, args: []string{"--verbose=3"}},
		{expected: []string{"-v=3"}, args: []string{"-v=3"}},
		{expected: []string{"-v=4"}, args: []string{"-v", "4"}},
		{expected: []string{"--verbose=4"}, args: []string{"--verbose", "4"}},
		{expected: []string{"--config=foo"}, args: []string{"--config", "foo"}},

		{expected: []string{"--verbose=4", "-v"}, args: []string{"--verbose", "4", "-v"}},
		{expected: []string{"-v=2", "-vvv"}, args: []string{"-v=2", "-vvv"}},
		{expected: []string{"--verbose=3", "-v=4"}, args: []string{"--verbose=3", "-v", "4"}},
		{expected: []string{"-vvv", "--verbose=4"}, args: []string{"-vvv", "--verbose", "4"}},
		{expected: []string{"--verbose=4", "-v=5"}, args: []string{"--verbose", "4", "-v=5"}},
		{expected: []string{"-v=4", "--verbose=5"}, args: []string{"-v", "4", "--verbose", "5"}},

		{
			expected: []string{"-v=4", "--config=foo", "--verbose=5"},
			args:     []string{"-v", "4", "--config", "foo", "--verbose", "5"},
		},
	}

	for _, test := range tests {
		convertedArgs, err := CapturePositionalArgs(test.args)
		if err != nil {
			t.Fatalf("got err: \n\t%v\n\tduring call to CapturePositionalArgs", err)
		}
		testStringSlice(t, convertedArgs, test.expected, "cfg.SystemLog.Verbosity")
	}
}

func TestLoadConfigAbsPath_Valid(t *testing.T) {
	var tests = []struct {
		args []string
	}{
		{args: []string{"--config=../../release/distsrc/example-mongosqld-config.yml"}},
		{args: []string{"--config", "../../release/distsrc/example-mongosqld-config.yml"}},
	}
	for _, test := range tests {
		_, convertedArgs, err := Load(test.args)
		if err != nil {
			t.Fatalf("got err: \n\t%v\n\tduring call to config.Load", err)
		}
		path := strings.Replace(convertedArgs[0], "--config=", "", 1)
		if !filepath.IsAbs(path) {
			t.Errorf("err, config path %v should be absolute, but is not\n", path)
		}
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
		_, err := ParseArgs(cfg, test.args)
		if err != nil {
			t.Fatalf("got err: \n\t%v\n\tduring call to ParseArgs", err)
		}
		testInt(t, cfg.SystemLog.Verbosity, test.level, "cfg.SystemLog.Verbosity")
	}
}

func TestSetParameter_Valid(t *testing.T) {
	var tests = []struct {
		alterationsEnabled bool
		args               []string
	}{
		{true, []string{"--setParameter", "enableTableAlterations=true"}},
		{false, []string{"--setParameter", "enableTableAlterations=false"}},
		{true, []string{"--setParameter=enableTableAlterations=true"}},
		{false, []string{"--setParameter=enableTableAlterations=false"}},
		{
			true,
			[]string{
				"--setParameter",
				"enableTableAlterations=false",
				"--setParameter",
				"enableTableAlterations=true",
			},
		},
		{
			false,
			[]string{"--setParameter",
				"enableTableAlterations=true",
				"--setParameter",
				"enableTableAlterations=false"},
		},
	}

	for _, test := range tests {
		cfg := Default()
		_, err := ParseArgs(cfg, test.args)
		if err != nil {
			t.Fatalf("got err: \n\t%v\n\tduring call to ParseArgs", err)
		}
		testBool(t,
			cfg.SetParameter.EnableTableAlterations,
			test.alterationsEnabled,
			"cfg.SetParameter.EnableTableAlterations",
		)
	}
}
