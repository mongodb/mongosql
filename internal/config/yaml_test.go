package config_test

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"testing"

	. "github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/log"
)

func TestParseYaml_Valid(t *testing.T) {
	cfg := Default()
	err := ParseYaml(cfg, bytes.NewBufferString(`
systemLog:
  logAppend: true
  logRotate: "reopen"
  path: temp
  quiet: true
  verbosity: 2

schema:
  path: "/var/test"
  maxVarcharLength: 1000
  refreshIntervalSecs: 0
  stored:
    mode: custom
    source: sampleDb
    name: mySchema
  sample:
    size: 969
    prejoin: true
    namespaces: ["foo.*", "*.bar"]
    refreshIntervalSecs: 983
    uuidSubtype3Encoding: java

runtime:
  memory:
    maxPerServer: 2000000
    maxPerConnection: 1000000
    maxPerStage: 102400

net:
  bindIp: 192.168.20.1
  port: 3306
  unixDomainSocket:
    enabled: false
    pathPrefix: "/var"
    filePermissions: "0600"
  ssl:
    mode: requireSSL
    allowInvalidCertificates: true
    PEMKeyFile: "pemkeyfile"
    PEMKeyPassword: "pemkeypassword"
    CAFile: "cafile"
    minimumTLSVersion: "TLS1_0"

security:
  enabled: true
  defaultMechanism: "GSSAPI"
  defaultSource: "$external"
  gssapi:
    hostname: "something"
    serviceName: "awesome"

mongodb:
  versionCompatibility: "3.2"
  net:
    uri: "mongodb://hostname:27018"
    numConnectionsPerSession: 3
    auth:
      username: user
      password: pass
      source: admin
      mechanism: scram
      gssapiServiceName: "hola"
    ssl:
      enabled: true
      allowInvalidCertificates: true
      allowInvalidHostnames: true
      PEMKeyFile: "mongopemkeyfile"
      PEMKeyPassword: "mongopemkeypassword"
      CAFile: "mongocafile"
      CRLFile: "mongocrlfile"
      FIPSMode: true
      minimumTLSVersion: "TLS1_0"

processManagement:
  service:
    name: oompa
    displayName: loompa
    description: doompa tee do

setParameter:
  anonymize_metrics: false
  enableTableAlterations: true
`), cfg.ConfigExpand)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testBool(t, cfg.SystemLog.LogAppend, true, "cfg.SystemLog.LogAppend")
	testString(t, string(cfg.SystemLog.LogRotate), string(log.Reopen), "cfg.SystemLog.LogRotate")
	testString(t, cfg.SystemLog.Path, "temp", "cfg.SystemLog.Quiet")
	testBool(t, cfg.SystemLog.Quiet, true, "cfg.SystemLog.Quiet")
	testInt64(t, cfg.SystemLog.Verbosity, 2, "cfg.SystemLog.Verbosity")

	testString(t, cfg.Schema.Path, "/var/test", "cfg.Schema.Path")
	testUint64(t, cfg.Schema.MaxVarcharLength, 1000, "cfg.Schema.MaxVarcharLength")
	testInt64(t, cfg.Schema.Sample.Size, 969, "cfg.Schema.Sample.Size")
	testBool(t, cfg.Schema.Sample.PreJoin, true, "cfg.Schema.Sample.PreJoin")
	testStoredSchemaMode(t, cfg.Schema.Stored.Mode, CustomStoredSchemaMode, "cfg.Schema.Stored.Mode")
	testString(t, cfg.Schema.Stored.Source, "sampleDb", "cfg.Schema.Stored.Source")
	testStringSlice(t,
		cfg.Schema.Sample.Namespaces,
		[]string{"foo.*",
			"*.bar"}, "cfg.Schema.Sample.Namespaces",
	)
	testInt64(t,
		cfg.Schema.RefreshIntervalSecs,
		983,
		"cfg.Schema.RefreshIntervalSecs",
	)
	testInt64(t,
		cfg.Schema.Sample.RefreshIntervalSecsDeprecated,
		0,
		"cfg.Schema.Sample.RefreshIntervalSecsDeprecated",
	)
	testString(t, cfg.Schema.Sample.UUIDSubtype3Encoding, "java", "cfg.Schema.UUIDSubtype3Encoding")

	testUint64(t, cfg.Runtime.Memory.MaxPerServer, 2000000, "cfg.Runtime.Memory.MaxPerServer")
	testUint64(t,
		cfg.Runtime.Memory.MaxPerConnection,
		1000000,
		"cfg.Runtime.Memory.MaxPerConnection")
	testUint64(t, cfg.Runtime.Memory.MaxPerStage, 102400, "cfg.Runtime.Memory.MaxPerStage")

	testStringSlice(t, cfg.Net.BindIP, []string{"192.168.20.1"}, "cfg.Net.BindIP")
	testInt(t, cfg.Net.Port, 3306, "cfg.Net.Port")
	if runtime.GOOS != "windows" {
		testBool(t, cfg.Net.UnixDomainSocket.Enabled, false, "cfg.Net.UnixDomainSocket.Enabled")
		testString(t,
			cfg.Net.UnixDomainSocket.PathPrefix,
			"/var",
			"cfg.Net.UnixDomainSocket.PathPrefix",
		)
		testString(t,
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
	testString(t, cfg.Net.SSL.MinimumTLSVersion, "TLS1_0",
		"cfg.Net.SSL.MinimumTLSVersion")

	testBool(t, cfg.Security.Enabled, true, "cfg.Security.Enabled")
	testString(t, cfg.Security.DefaultMechanism, "GSSAPI", "cfg.Security.DefaultMechanism")
	testString(t, cfg.Security.DefaultSource, "$external", "cfg.Security.DefaultSource")
	testString(t, cfg.Security.GSSAPI.Hostname, "something", "cfg.Security.GSSAPI.Hostname")
	testString(t, cfg.Security.GSSAPI.ServiceName, "awesome", "cfg.Security.GSSAPI.ServiceName")

	testString(t, cfg.MongoDB.VersionCompatibility, "3.2", "cfg.MongoDB.VersionCompatibility")
	testString(t, cfg.MongoDB.Net.URI, "mongodb://hostname:27018", "cfg.MongoDB.Net.URI")
	testInt(t,
		cfg.MongoDB.Net.NumConnectionsPerSession,
		3,
		"cfg.MongoDB.Net.NumConnectionsPerSession",
	)

	testString(t, cfg.MongoDB.Net.Auth.Username, "user", "cfg.MongoDB.Net.Auth.Username")
	testString(t, cfg.MongoDB.Net.Auth.Password, "pass", "cfg.MongoDB.Net.Auth.Password")
	testString(t, cfg.MongoDB.Net.Auth.Source, "admin", "cfg.MongoDB.Net.Auth.Source")
	testString(t, cfg.MongoDB.Net.Auth.Mechanism, "scram", "cfg.MongoDB.Net.Auth.Mechanism")

	testString(t,
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
	testString(t, cfg.MongoDB.Net.SSL.MinimumTLSVersion, "TLS1_0",
		"cfg.MongoDB.Net.SSL.MinimumTLSVersion")

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

	testBool(t,
		cfg.SetParameter.EnableTableAlterations,
		true,
		"cfg.SetParameter.EnableTableAlterations",
	)

	testBool(t,
		cfg.SetParameter.AnonymizeMetrics,
		false,
		"cfg.SetParameter.AnonymizeMetrics",
	)
}

func TestParseYaml_Valid2(t *testing.T) {
	cfg := Default()
	err := ParseYaml(cfg, bytes.NewBufferString(`
net:
  bindIp: 192.168.20.1,host2
  port: 3306
`), cfg.ConfigExpand)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testStringSlice(t, cfg.Net.BindIP, []string{"192.168.20.1", "host2"}, "cfg.Net.BindIP")
	testInt(t, cfg.Net.Port, 3306, "cfg.Net.Port")
}

func TestParseYaml_Valid3(t *testing.T) {
	cfg := Default()
	err := ParseYaml(cfg, bytes.NewBufferString(`
net:
  bindIp: 
    - 192.168.20.1
    - host2
  port: 3306
`), cfg.ConfigExpand)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testStringSlice(t, cfg.Net.BindIP, []string{"192.168.20.1", "host2"}, "cfg.Net.BindIP")
	testInt(t, cfg.Net.Port, 3306, "cfg.Net.Port")
}

func TestParseYaml_MongodbSRVURI(t *testing.T) {
	cfg := Default()
	err := ParseYaml(cfg, bytes.NewBufferString(`
mongodb:
  net:
    uri: "mongodb+srv://hostname"
`), cfg.ConfigExpand)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testString(t, cfg.MongoDB.Net.URI, "mongodb+srv://hostname", "cfg.MongoDB.Net.URI")
}

func TestParseYaml_RefreshInterval(t *testing.T) {
	cfg := Default()
	err := ParseYaml(cfg, bytes.NewBufferString(`
schema:
  refreshIntervalSecs: 10
`), cfg.ConfigExpand)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testInt64(t, cfg.Schema.RefreshIntervalSecs, 10, "cfg.Schema.RefreshIntervalSecs")
	testInt64(t, cfg.Schema.Sample.RefreshIntervalSecsDeprecated, 0, "cfg.Schema.Sample.RefreshIntervalSecsDeprecated")
}

func TestParseYaml_RefreshInterval_Deprecated(t *testing.T) {
	cfg := Default()
	err := ParseYaml(cfg, bytes.NewBufferString(`
schema:
  sample:
    refreshIntervalSecs: 10
`), cfg.ConfigExpand)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testInt64(t, cfg.Schema.RefreshIntervalSecs, 10, "cfg.Schema.RefreshIntervalSecs")
	testInt64(t, cfg.Schema.Sample.RefreshIntervalSecsDeprecated, 0, "cfg.Schema.Sample.RefreshIntervalSecsDeprecated")
}

func TestParseYaml_RefreshInterval_Override(t *testing.T) {
	cfg := Default()
	err := ParseYaml(cfg, bytes.NewBufferString(`
schema:
  refreshIntervalSecs: 0
  sample:
    refreshIntervalSecs: 10
`), cfg.ConfigExpand)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testInt64(t, cfg.Schema.RefreshIntervalSecs, 10, "cfg.Schema.RefreshIntervalSecs")
	testInt64(t, cfg.Schema.Sample.RefreshIntervalSecsDeprecated, 0, "cfg.Schema.Sample.RefreshIntervalSecsDeprecated")
}

func TestParseYaml_RefreshInterval_No_Override(t *testing.T) {
	cfg := Default()
	err := ParseYaml(cfg, bytes.NewBufferString(`
schema:
  refreshIntervalSecs: 100
  sample:
    refreshIntervalSecs: 10
`), cfg.ConfigExpand)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testInt64(t, cfg.Schema.RefreshIntervalSecs, 100, "cfg.Schema.RefreshIntervalSecs")
	testInt64(t, cfg.Schema.Sample.RefreshIntervalSecsDeprecated, 0, "cfg.Schema.Sample.RefreshIntervalSecsDeprecated")
}

func TestParseYaml_Exec_Config_Expand(t *testing.T) {
	cfg := Default()
	yaml := `
mongodb:
  net:
    auth:
      username: user
      password:
        __exec: "echo pwd123"
`

	// --configExpand=none
	cfg.ConfigExpand = Expansion{
		Exec: false,
		Rest: false,
	}
	err := ParseYaml(cfg, bytes.NewBufferString(yaml), cfg.ConfigExpand)
	if err == nil {
		t.Fatalf("test should have failed")
	}

	if !strings.Contains(err.Error(), "invalid value for mongodb.net.auth.password, expected a string") {
		t.Fatalf("incorrect error string: %v", err.Error())
	}

	// --configExpand=rest
	cfg.ConfigExpand = Expansion{
		Rest: true,
	}
	err = ParseYaml(cfg, bytes.NewBufferString(yaml), cfg.ConfigExpand)
	if err == nil {
		t.Fatalf("test should have failed")
	}

	if !strings.Contains(err.Error(), "__exec has not been enabled via --configExpand") {
		t.Fatalf("incorrect error string: %v", err.Error())
	}

	// --configExpand=exec,rest
	cfg.ConfigExpand = Expansion{
		Exec: true,
		Rest: true,
	}
	err = ParseYaml(cfg, bytes.NewBufferString(yaml), cfg.ConfigExpand)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testString(t, cfg.MongoDB.Net.Auth.Password, "pwd123", "cfg.MongoDB.Net.Auth.Password")
}

// __exec command execution fails
func TestParseYaml_Exec_Command_Failure(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = Expansion{
		Exec: true,
	}
	err := ParseYaml(cfg, bytes.NewBufferString(`
mongodb:
  net:
    auth:
      username: user
      password:
        __exec: "false"
`), cfg.ConfigExpand)

	if err == nil {
		t.Fatalf("test should have failed")
	}
	if !strings.Contains(err.Error(), "error executing '__exec' command") {
		t.Fatalf("incorrect error string: %v", err.Error())
	}
}

// All exec'd values should be of type string. Type conversions are not supported for non-ints.
func TestParseYaml_Exec_Failure_Non_String_Field(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = Expansion{
		Exec: true,
	}
	err := ParseYaml(cfg, bytes.NewBufferString(`
schema:
  path: "/var/test"
  sample:
    prejoin:
      __exec: "echo true"
`), cfg.ConfigExpand)
	if err == nil {
		t.Fatalf("test should have failed")
	}
	if !strings.Contains(err.Error(), "invalid value for schema.sample.prejoin, expected a bool") {
		t.Fatalf("incorrect error string: %v", err.Error())
	}
}

// If multiple '__exec' keys are present, only the last one will be evaluated.
func TestParseYaml_Exec_Failure_Multiple_Exec_Keys(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = Expansion{
		Exec: true,
	}
	err := ParseYaml(cfg, bytes.NewBufferString(`
mongodb:
  net:
    auth:
      username: user
      password:
        __exec: "echo pwd123"
        __exec: "echo another_pwd"
`), cfg.ConfigExpand)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testString(t, cfg.MongoDB.Net.Auth.Password, "another_pwd", "cfg.MongoDB.Net.Auth.Password")
}

// Fail when there is an unknown key in the __exec block.
func TestParseYaml_Exec_Failure_Unknown_Key(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = Expansion{
		Exec: true,
	}
	err := ParseYaml(cfg, bytes.NewBufferString(`
mongodb:
  net:
    auth:
      username: user
      password:
        __exec: "echo pwd123"
        name: "random"
`), cfg.ConfigExpand)
	if err == nil {
		t.Fatalf("test should have failed")
	}
	if !strings.Contains(err.Error(), "invalid value for mongodb.net.auth.password, expected a string") {
		t.Fatalf("incorrect error string: %v", err.Error())
	}
}

func TestParseYaml_Invalid(t *testing.T) {

	var tests = []struct {
		err  string
		yaml string
	}{
		{err: "unrecognized key 'funny'", yaml: `
funny: 10
`},
		{err: "invalid value for systemLog, expected a map: 10(int64)", yaml: `
systemLog: 10
`},
		{err: "unrecognized key 'systemLog.funny'", yaml: `
systemLog:
    funny: 10
`},
		{err: "invalid value for systemLog.logAppend, expected a bool: 4(int64)", yaml: `
systemLog:
    logAppend: 4
`},
		{err: "invalid value for systemLog.path, expected a string: 4(int64)", yaml: `
systemLog:
    path: 4
`},
		{
			err: "invalid value for systemLog.verbosity: strconv.ParseInt: parsing \"funny\": " +
				"invalid syntax",
			yaml: `
systemLog:
    verbosity: funny
`,
		},
		{err: "unrecognized key 'setParameter.funny'", yaml: `
setParameter:
    funny: 10
`},
		{
			err: "invalid value for setParameter.enableTableAlterations, " +
				"expected a bool: abcde(string)",
			yaml: `
setParameter:
    enableTableAlterations: abcde
`,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%v-%v", i, test.err), func(t *testing.T) {
			cfg := Default()
			err := ParseYaml(cfg, bytes.NewBufferString(test.yaml), cfg.ConfigExpand)
			if err == nil {
				t.Fatalf("expected error, but got none")
			}

			if err.Error() != test.err {
				t.Fatalf("expected err '%s' but got '%v'", test.err, err)
			}
		})
	}
}
