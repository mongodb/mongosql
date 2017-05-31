package config_test

import (
	"runtime"
	"testing"

	. "github.com/10gen/sqlproxy/internal/config"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	testBool(t, cfg.SystemLog.LogAppend, false, "cfg.SystemLog.Append")
	testString(t, cfg.SystemLog.Path, "", "cfg.SystemLog.Quiet")
	testBool(t, cfg.SystemLog.Quiet, false, "cfg.SystemLog.Quiet")
	testInt(t, cfg.SystemLog.Verbosity, 0, "cfg.SystemLog.Verbosity")

	testString(t, cfg.Schema.Path, "", "cfg.Schema.Path")
	testUint16(t, cfg.Schema.MaxVarcharLength, 0, "cfg.Schema.MaxVarcharLength")

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

	cfg, err := Load(args)
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

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none", err)
	}

	expected := "unsupported sample authentication mechanism 'GSSAPI'"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}

	cfg.MongoDB.Net.Auth.Mechanism = "foo"

	err = Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none", err)
	}

	expected = "unsupported sample authentication mechanism 'foo'"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}
}

func TestValidate_No_Schema_Path(t *testing.T) {
	cfg := Default()
	cfg.Schema.Path = ""

	err := Validate(cfg)
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}
}

func TestValidate_SampleAuth_options_specified_but_auth_disabled(t *testing.T) {
	cfg := Default()
	cfg.MongoDB.Net.Auth.Username = "foo"

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none", err)
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
		t.Fatalf("expected an error, but got none", err)
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
		t.Fatalf("expected an error, but got none", err)
	}

	expected := "when specifying SSL options, SSL must be enabled with --mongo-ssl or in a config file at 'mongodb.net.ssl.enabled'"
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
		t.Fatalf("expected an error, but got none", err)
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
		t.Fatalf("expected an error, but got none", err)
	}

	expected := "filePermissions must be valid octal"
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
