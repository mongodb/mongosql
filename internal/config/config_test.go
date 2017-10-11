package config_test

import (
	"fmt"
	"runtime"
	"testing"

	"strings"

	. "github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/log"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	testBool(t, cfg.SystemLog.LogAppend, false, "cfg.SystemLog.LogAppend")
	testString(t, string(cfg.SystemLog.LogRotate), string(log.Rename), "cfg.SystemLog.LogRotate")
	testString(t, cfg.SystemLog.Path, "", "cfg.SystemLog.Quiet")
	testBool(t, cfg.SystemLog.Quiet, false, "cfg.SystemLog.Quiet")
	testInt(t, cfg.SystemLog.Verbosity, 0, "cfg.SystemLog.Verbosity")

	testString(t, cfg.Schema.Path, "", "cfg.Schema.Path")
	testUint16(t, cfg.Schema.MaxVarcharLength, 0, "cfg.Schema.MaxVarcharLength")
	testSampleMode(t, cfg.Schema.Sample.Mode, "read", "cfg.Schema.Sample.Mode")
	testString(t, cfg.Schema.Sample.Source, "", "cfg.Schema.Sample.Source")
	testInt64(t, cfg.Schema.Sample.Size, 1000, "cfg.Schema.Sample.Size")
	testStringSlice(t, cfg.Schema.Sample.Namespaces, []string{"*.*"}, "cfg.Schema.Sample.Namespaces")
	testInt64(t, cfg.Schema.Sample.RefreshIntervalSecs, 0, "cfg.Schema.Sample.RefreshIntervalSecs")
	testString(t, cfg.Schema.Sample.UUIDSubtype3Encoding, "old", "cfg.Schema.Sample.UUIDSubtype3Encoding")

	testUint64(t, cfg.Runtime.Memory.MaxPerStage, 0, "cfg.Runtime.Memory.MaxPerStage")

	testStringSlice(t, cfg.Net.BindIP, []string{"127.0.0.1"}, "cfg.Net.BindIP")
	testInt(t, cfg.Net.Port, 3307, "cfg.Net.Port")
	if runtime.GOOS != "windows" {
		testBool(t, cfg.Net.UnixDomainSocket.Enabled, true, "cfg.Net.UnixDomainSocket.Enabled")
		testString(t, cfg.Net.UnixDomainSocket.PathPrefix, "/tmp", "cfg.Net.UnixDomainSocket.PathPrefix")
		testString(t, cfg.Net.UnixDomainSocket.FilePermissions, "0700", "cfg.Net.UnixDomainSocket.FilePermissions")
	} else {
		testBool(t, cfg.Net.UnixDomainSocket.Enabled, false, "cfg.Net.UnixDomainSocket.Enabled")
		testString(t, cfg.Net.UnixDomainSocket.PathPrefix, "", "cfg.Net.UnixDomainSocket.PathPrefix")
		testString(t, cfg.Net.UnixDomainSocket.FilePermissions, "", "cfg.Net.UnixDomainSocket.FilePermissions")
	}

	testString(t, cfg.Net.SSL.Mode, "disabled", "cfg.Net.SSL.Mode")
	testBool(t, cfg.Net.SSL.AllowInvalidCertificates, false, "cfg.Net.SSL.AllowInvalidCertificates")
	testString(t, cfg.Net.SSL.PEMKeyFile, "", "cfg.Net.SSL.PEMKeyFile")
	testString(t, cfg.Net.SSL.PEMKeyPassword, "", "cfg.Net.SSL.PEMKeyPassword")
	testString(t, cfg.Net.SSL.CAFile, "", "cfg.Net.SSL.CAFile")

	testBool(t, cfg.Security.Enabled, false, "cfg.Security.Enabled")
	testString(t, cfg.Security.DefaultMechanism, "SCRAM-SHA-1", "cfg.Security.DefaultMechanism")
	testString(t, cfg.Security.DefaultSource, "admin", "cfg.Security.DefaultSource")

	testString(t, cfg.MongoDB.VersionCompatibility, "", "cfg.MongoDB.VersionCompatibility")
	testString(t, cfg.MongoDB.Net.URI, "mongodb://localhost:27017", "cfg.MongoDB.Net.URI")

	testBool(t, cfg.MongoDB.Net.SSL.Enabled, false, "cfg.MongoDB.Net.SSL.Enabled")
	testBool(t, cfg.MongoDB.Net.SSL.AllowInvalidCertificates, false, "cfg.MongoDB.Net.SSL.AllowInvalidCertificates")
	testBool(t, cfg.MongoDB.Net.SSL.AllowInvalidHostnames, false, "cfg.MongoDB.Net.SSL.AllowInvalidHostnames")
	testString(t, cfg.MongoDB.Net.SSL.PEMKeyFile, "", "cfg.MongoDB.Net.SSL.PEMKeyFile")
	testString(t, cfg.MongoDB.Net.SSL.PEMKeyPassword, "", "cfg.MongoDB.Net.SSL.PEMKeyPassword")
	testString(t, cfg.MongoDB.Net.SSL.CAFile, "", "cfg.MongoDB.Net.SSL.CAFile")
	testString(t, cfg.MongoDB.Net.SSL.CRLFile, "", "cfg.MongoDB.Net.SSL.CRLFile")
	testBool(t, cfg.MongoDB.Net.SSL.FIPSMode, false, "cfg.MongoDB.Net.SSL.FIPSMode")

	testString(t, cfg.ProcessManagement.Service.Name, "mongosql", "cfg.ProcessManagement.Service.Name")
	testString(t, cfg.ProcessManagement.Service.DisplayName, "MongoSQL Service", "cfg.ProcessManagement.Service.DisplayName")
	testString(t, cfg.ProcessManagement.Service.Description, "MongoSQL accesses MongoDB data with SQL", "cfg.ProcessManagement.Service.Description")
}

func TestLoad(t *testing.T) {
	args := []string{
		"--config", "testdata/sample.conf",
		"-vv",
	}

	cfg, _, err := Load(args)
	if err != nil {
		t.Fatalf("expected no error, but got '%v'", err)
	}

	testInt(t, cfg.SystemLog.Verbosity, 2, "cfg.SystemLog.Verbosity") // 2 from the args, as opposed to the 4 from the file
	testString(t, cfg.Schema.Path, "somewhere", "cfg.Schema.Path")
}

func TestToJSON(t *testing.T) {
	cfg := Default()

	cfg.Config = "funny"
	cfg.SystemLog.LogAppend = true
	cfg.Net.SSL.PEMKeyPassword = "harumph"

	actual := ToJSON(cfg)

	expected := `{config: "funny", systemLog: {logAppend: true}, net: {ssl: {PEMKeyPassword: "<protected>"}}}`
	if actual != expected {
		t.Fatalf("expected '%s', but got '%s'", expected, actual)
	}
}

func TestValidate_Valid(t *testing.T) {
	cfg := Default()
	cfg.Schema.Path = "something"

	err := Validate(cfg)
	if err != nil {
		t.Fatalf("expected no error, but got '%v'", err)
	}
}

func TestValidate_Invalid_SampleAuth_Mechanism(t *testing.T) {
	cfg := Default()
	cfg.MongoDB.Net.Auth.Mechanism = "GSSAPI"
	cfg.Schema.Sample.Source = "test"

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected := "unsupported sample authentication mechanism 'GSSAPI'"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}

	cfg.MongoDB.Net.Auth.Mechanism = "foo"

	err = Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected = "unsupported sample authentication mechanism 'foo'"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}
}

func TestValidate_Sample_Invalid_Mode(t *testing.T) {
	tests := []struct {
		mode  SampleMode
		valid bool
	}{
		{ReadSampleMode, true},
		{WriteSampleMode, true},
		{SampleMode("nope"), false},
	}

	for _, test := range tests {
		t.Run(string(test.mode), func(t *testing.T) {
			cfg := Default()
			cfg.Schema.Sample.Source = "temp"
			cfg.Schema.Sample.Mode = test.mode

			err := Validate(cfg)
			if err != nil && test.valid {
				t.Fatalf("expected no error, but got %v", err)
			}

			if err == nil && !test.valid {
				t.Fatalf("expected an error, but got none")
			}
		})
	}
}

func TestValidate_Sample_Invalid_Namespaces(t *testing.T) {
	tests := []struct {
		ns    []string
		valid bool
	}{
		{ns: []string{"one"}, valid: true},
		{ns: []string{"one.*"}, valid: true},
		{ns: []string{"*.two"}, valid: true},
		{ns: []string{"one.two"}, valid: true},
		{ns: []string{"one.two", "three"}, valid: true},
		{ns: []string{".two"}, valid: false},
		{ns: []string{"one."}, valid: false},
		{ns: []string{"three", "one."}, valid: false},
		{ns: []string{".three", "one"}, valid: false},
		{ns: []string{"som$ething"}, valid: false},
	}

	for _, test := range tests {
		t.Run(strings.Join(test.ns, ","), func(t *testing.T) {
			cfg := Default()
			cfg.Schema.Sample.Namespaces = test.ns

			err := Validate(cfg)
			if err != nil && test.valid {
				t.Fatalf("expected no error, but got %v", err)
			}

			if err == nil && !test.valid {
				t.Fatalf("expected an error, but got none")
			}
		})
	}
}

func TestValidate_Sample_Invalid_Source(t *testing.T) {
	tests := []struct {
		source string
		valid  bool
	}{
		{"test", true},
		{"some$where", false},
		{".somewhere", false},
		{"somewhere.", false},
	}

	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			cfg := Default()
			cfg.Schema.Sample.Source = test.source

			err := Validate(cfg)
			if err != nil && test.valid {
				t.Fatalf("expected no error, but got %v", err)
			}

			if err == nil && !test.valid {
				t.Fatalf("expected an error, but got none")
			}
		})
	}
}

func TestValidate_Sample_Source_And_Schema(t *testing.T) {
	tests := []struct {
		source string
		schema string
		valid  bool
	}{
		{"test", "path", false},
		{"test", "", true},
		{"", "path", true},
		{"", "", true},
	}

	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			cfg := Default()
			cfg.Schema.Sample.Source = test.source
			cfg.Schema.Path = test.schema
			err := Validate(cfg)
			if err != nil && test.valid {
				t.Fatalf("expected no error, but got %v", err)
			}

			if err == nil && !test.valid {
				t.Fatalf("expected an error, but got none")
			}
		})
	}
}

func TestValidate_Sample_StandaloneWriter(t *testing.T) {
	cfg := Default()
	cfg.Schema.Sample.Mode = WriteSampleMode

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected := "sample mode 'write' requires a non-empty sample source"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but go '%s'", expected, err)
	}
}

func TestValidate_Sample_ClusteredSamplingReader(t *testing.T) {
	cfg := Default()
	cfg.Schema.Sample.Source = "somewhere"
	cfg.Schema.Sample.RefreshIntervalSecs = 1

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected := "sample mode 'read' with a non-empty sample source cannot specify a sample refresh interval"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but go '%s'", expected, err)
	}
}

func TestValidate_SampleAuth_options_specified_but_auth_disabled(t *testing.T) {
	cfg := Default()
	cfg.MongoDB.Net.Auth.Username = "foo"
	cfg.Schema.Sample.Source = "test"

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected := "when specifying sample authentication options, auth must " +
		"be enabled with --auth or in a config file at 'security.enabled'"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}

	cfg.MongoDB.Net.Auth.Username = ""
	cfg.MongoDB.Net.Auth.Password = "foo"

	err = Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected = "when specifying sample authentication options, auth must " +
		"be enabled with --auth or in a config file at 'security.enabled'"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}
}

func TestValidate_SSL_options_specified_but_disabled(t *testing.T) {
	cfg := Default()
	cfg.Schema.Path = "something"
	cfg.MongoDB.Net.SSL.CRLFile = "lkasdf"

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected := "when specifying MongoDB SSL options, SSL must be enabled with --mongo-ssl or in a configuration file at 'mongodb.net.ssl.enabled'"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}
}

func TestValidate_sqlproxy_SSL_options_specified_but_disabled(t *testing.T) {
	cfg := Default()
	cfg.Schema.Path = "something"
	cfg.Net.SSL.CAFile = "lkasdf"

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected := "when specifying SSL options, SSL must be enabled with --sslMode or in a configuration file at 'net.ssl.mode'"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}
}

func TestValidate_sqlproxy_SSL_options_PEMKeyFile(t *testing.T) {
	cfg := Default()
	cfg.Schema.Path = "something"
	cfg.Net.SSL.Mode = "allowSSL"

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected := "need sslPEMKeyFile when SSL is enabled"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}
}

func TestValidate_sqlproxy_SSL_options_bad_sslMode(t *testing.T) {
	cfg := Default()
	cfg.Schema.Path = "something"
	cfg.Net.SSL.Mode = "abcde"

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected := "invalid sslMode abcde"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}
}

func TestValidate_UnixDomainSocket_on_windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.SkipNow()
	}
	cfg := Default()
	cfg.Schema.Path = "something"
	cfg.Net.UnixDomainSocket.Enabled = true

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected := "unix domain sockets are not supported on windows"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}
}

func TestValidate_UnixDomainSocket(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.SkipNow()
	}
	cfg := Default()
	cfg.Schema.Path = "something"
	cfg.Net.UnixDomainSocket.FilePermissions = "asdfasdf"

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected := "filePermissions must be valid octal"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}
}

func TestValidate_LogRotate_unsupported(t *testing.T) {
	cfg := Default()
	cfg.SystemLog.LogRotate = "asdfasdf"
	cfg.Schema.Path = "something"

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected := "Unsupported log rotation strategy 'asdfasdf'"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}
}

func TestValidate_Invalid_SampleSize(t *testing.T) {
	cfg := Default()
	cfg.Schema.Sample.Size = -1
	cfg.Schema.Sample.Source = "test"

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected := "invalid sample size: -1"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}
}

func TestValidate_Too_Few_NumConnectionsPerSession(t *testing.T) {
	cfg := Default()
	cfg.MongoDB.Net.NumConnectionsPerSession = 0

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected := fmt.Sprintf("invalid number of MongoDB connections: 0 (must be between %d and %d)", MinConnections, MaxConnections)
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}
}

func TestValidate_Too_Many_NumConnectionsPerSession(t *testing.T) {
	cfg := Default()
	cfg.MongoDB.Net.NumConnectionsPerSession = 1000

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected := fmt.Sprintf("invalid number of MongoDB connections: 1000 (must be between %d and %d)", MinConnections, MaxConnections)
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}
}

func testBool(t *testing.T, actual, expected bool, key string) {
	if actual != expected {
		t.Errorf("%s should be %v but was %v", key, expected, actual)
	}
}

func testInt(t *testing.T, actual, expected int, key string) {
	if actual != expected {
		t.Errorf("%s should be %v but was %v", key, expected, actual)
	}
}

func testInt64(t *testing.T, actual, expected int64, key string) {
	if actual != expected {
		t.Errorf("%s should be %v but was %v", key, expected, actual)
	}
}

func testSampleMode(t *testing.T, actual, expected SampleMode, key SampleMode) {
	if actual != expected {
		t.Errorf("%s should be %v but was %v", key, expected, actual)
	}
}

func testString(t *testing.T, actual, expected string, key string) {
	if actual != expected {
		t.Errorf("%s should be %v but was %v", key, expected, actual)
	}
}

func testStringSlice(t *testing.T, actual, expected []string, key string) {
	if len(actual) != len(expected) {
		t.Errorf("%s should be %v but was %v", key, expected, actual)
	}

	for i := 0; i < len(actual); i++ {
		if actual[i] != expected[i] {
			t.Errorf("%s should be %v but was %v", key, expected, actual)
		}
	}
}

func testUint16(t *testing.T, actual, expected uint16, key string) {
	if actual != expected {
		t.Errorf("%s should be %v but was %v", key, expected, actual)
	}
}

func testUint64(t *testing.T, actual, expected uint64, key string) {
	if actual != expected {
		t.Errorf("%s should be %v but was %v", key, expected, actual)
	}
}
