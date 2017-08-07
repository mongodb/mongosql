package config_test

import (
	"bytes"
	"fmt"
	"runtime"
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
  sample:
    databases: ["a", "b"]
    mode: read
    size: 969
    namespaces: ["foo.*", "*.bar"]
    readIntervalSecs: 1005
    writeIntervalSecs: 983
    uuidSubtype3Encoding: java

runtime:
  memory:
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

security:
  enabled: true
  defaultMechanism: "GSSAPI"
  defaultSource: "$external"

mongodb:
  versionCompatibility: "3.2"
  net:
    uri: "mongodb://hostname:27018"
    auth:
      username: user
      password: pass
      source: admin
      mechanism: scram
    ssl:
      enabled: true
      allowInvalidCertificates: true
      allowInvalidHostnames: true
      PEMKeyFile: "mongopemkeyfile"
      PEMKeyPassword: "mongopemkeypassword"
      CAFile: "mongocafile"
      CRLFile: "mongocrlfile"
      FIPSMode: true

processManagement:
  service:
    name: oompa
    displayName: loompa
    description: doompa tee do
`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testBool(t, cfg.SystemLog.LogAppend, true, "cfg.SystemLog.LogAppend")
	testString(t, string(cfg.SystemLog.LogRotate), string(log.Reopen), "cfg.SystemLog.LogRotate")
	testString(t, cfg.SystemLog.Path, "temp", "cfg.SystemLog.Quiet")
	testBool(t, cfg.SystemLog.Quiet, true, "cfg.SystemLog.Quiet")
	testInt(t, cfg.SystemLog.Verbosity, 2, "cfg.SystemLog.Verbosity")

	testString(t, cfg.Schema.Path, "/var/test", "cfg.Schema.Path")
	testUint16(t, cfg.Schema.MaxVarcharLength, 1000, "cfg.Schema.MaxVarcharLength")
	testInt64(t, cfg.Schema.Sample.Size, 969, "cfg.Schema.Sample.Size")
	testString(t, cfg.Schema.Sample.Mode, "read", "cfg.Schema.Sample.Mode")
	testStringSlice(t, cfg.Schema.Sample.Namespaces, []string{"foo.*", "*.bar"}, "cfg.Schema.Sample.Namespaces")
	testInt64(t, cfg.Schema.Sample.ReadIntervalSecs, 1005, "cfg.Schema.Sample.ReadIntervalSecs")
	testInt64(t, cfg.Schema.Sample.WriteIntervalSecs, 983, "cfg.Schema.Sample.WriteIntervalSecs")
	testString(t, cfg.Schema.Sample.UUIDSubtype3Encoding, "java", "cfg.Schema.UUIDSubtype3Encoding")
	testStringSlice(t, cfg.Schema.Sample.Databases, []string{"a", "b"}, "cfg.Schema.Sample.Databases")

	testUint64(t, cfg.Runtime.Memory.MaxPerStage, 102400, "cfg.Runtime.Memory.MaxPerStage")

	testStringSlice(t, cfg.Net.BindIP, []string{"192.168.20.1"}, "cfg.Net.BindIP")
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

	testString(t, cfg.MongoDB.Net.Auth.Username, "user", "cfg.MongoDB.Net.Auth.Username")
	testString(t, cfg.MongoDB.Net.Auth.Password, "pass", "cfg.MongoDB.Net.Auth.Password")
	testString(t, cfg.MongoDB.Net.Auth.Source, "admin", "cfg.MongoDB.Net.Auth.Source")
	testString(t, cfg.MongoDB.Net.Auth.Mechanism, "scram", "cfg.MongoDB.Net.Auth.Mechanism")

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

func TestParseYaml_Valid2(t *testing.T) {
	cfg := Default()
	err := ParseYaml(cfg, bytes.NewBufferString(`
net:
  bindIp: 192.168.20.1,host2
  port: 3306
`))
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
`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testStringSlice(t, cfg.Net.BindIP, []string{"192.168.20.1", "host2"}, "cfg.Net.BindIP")
	testInt(t, cfg.Net.Port, 3306, "cfg.Net.Port")
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
		{err: "invalid value for systemLog.verbosity: strconv.ParseInt: parsing \"funny\": invalid syntax", yaml: `
systemLog:
    verbosity: funny
`},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%v-%v", i, test.err), func(t *testing.T) {
			cfg := Default()
			err := ParseYaml(cfg, bytes.NewBufferString(test.yaml))
			if err == nil {
				t.Fatalf("expected error, but got none")
			}

			if err.Error() != test.err {
				t.Fatalf("expected err '%s' but got '%v'", test.err, err)
			}
		})
	}
}
