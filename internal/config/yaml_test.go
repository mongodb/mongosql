package config_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/10gen/sqlproxy/internal/httputil"

	. "github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/log"
)

// hash === SHA256HMAC('pwd123', 'secret')
const hash = "f06f043029484a50667d7e00c7a0fe9256310d3c503e7d8ccc090637efc26fc4"
const hashYaml = "9f3efaba72ecff8efe0872d2cae480d623f2859205e521f9e8b28eedf2989048"
const digestKey = "736563726574"

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
	cfg.ConfigExpand = EnabledExpansions{
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
	cfg.ConfigExpand = EnabledExpansions{
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
	cfg.ConfigExpand = EnabledExpansions{
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
	cfg.ConfigExpand = EnabledExpansions{
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
	cfg.ConfigExpand = EnabledExpansions{
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
	cfg.ConfigExpand = EnabledExpansions{
		Exec: true,
	}
	err := ParseYaml(cfg, bytes.NewBufferString(`
mongodb:
  net:
    auth:
      username: "user"
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
	cfg.ConfigExpand = EnabledExpansions{
		Exec: true,
	}
	err := ParseYaml(cfg, bytes.NewBufferString(`
mongodb:
  net:
    auth:
      username: "user"
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

func TestParseYaml_Rest_Success(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = EnabledExpansions{
		Rest: true,
	}

	// Create a new reader with the password string.
	r := ioutil.NopCloser(bytes.NewReader([]byte("pwd123")))
	client := httputil.MockClient{GetFunc: func(_ string) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}}
	httputil.SetClient(&client)

	err := ParseYaml(cfg, bytes.NewBufferString(`
mongodb:
  net:
    auth:
      username: "user"
      password:
        __rest: "fakeurl"
`), cfg.ConfigExpand)
	if err != nil {
		t.Fatalf("unexpected err: %v\n", err)
	}

	testString(t, cfg.MongoDB.Net.Auth.Password, "pwd123", "cfg.MongoDB.Net.Auth.Password")
}

// Fail when the GET request fails (non-200 error code).
func TestParseYaml_Rest_Request_400_Status(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = EnabledExpansions{
		Rest: true,
	}

	// Create a new reader with the password string, and a response object with a non-200 error code.
	r := ioutil.NopCloser(bytes.NewReader([]byte("pwd123")))
	client := httputil.MockClient{GetFunc: func(_ string) (*http.Response, error) {
		return &http.Response{
			StatusCode: 400,
			Body:       r,
		}, nil
	}}
	httputil.SetClient(&client)

	err := ParseYaml(cfg, bytes.NewBufferString(`
mongodb:
  net:
    auth:
      username: "user"
      password:
        __rest: "fakeurl"
`), cfg.ConfigExpand)
	if err == nil {
		t.Fatalf("test should have failed")
	}

	if !strings.Contains(err.Error(), "GET request resulted in status code 400 from url") {
		t.Fatalf("incorrect error string: %v", err.Error())
	}
}

// Test when the GET request fails with an error and no response.
func TestParseYaml_Rest_Request_Failure(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = EnabledExpansions{
		Rest: true,
	}

	errString := "no response error"

	client := httputil.MockClient{GetFunc: func(_ string) (*http.Response, error) {
		return nil, errors.New(errString)
	}}
	httputil.SetClient(&client)

	err := ParseYaml(cfg, bytes.NewBufferString(`
mongodb:
  net:
    auth:
      username: "user"
      password:
        __rest: "fakeurl"
`), cfg.ConfigExpand)
	if err == nil {
		t.Fatalf("test should have failed")
	}

	if !strings.Contains(err.Error(), errString) {
		t.Fatalf("incorrect error string: %v", err.Error())
	}
}

// Cannot have both an '__exec' and a '__rest' in a block.
func TestParseYaml_Failure_Both_Exec_And_Rest(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = EnabledExpansions{
		Exec: true,
		Rest: true,
	}
	err := ParseYaml(cfg, bytes.NewBufferString(`
mongodb:
  net:
    auth:
      username: "user"
      password:
        __exec: "echo pwd123"
        __rest: "fakeurl"
`), cfg.ConfigExpand)
	if err == nil {
		t.Fatalf("test should have failed")
	}
	if !strings.Contains(err.Error(), "can only use __exec or __rest, not both") {
		t.Fatalf("incorrect error string: %v", err.Error())
	}
}

// startTestServer creates a server with `numRoutes` routes to test redirect handling when evaluating
// a '__rest' expansion directive.
func startTestServer(numRoutes int, body string) *httptest.Server {
	mux := http.NewServeMux()

	for i := 0; i < numRoutes; i++ {
		route := fmt.Sprintf("/route%v/", i)
		nextRoute := fmt.Sprintf("/route%v/", i+1)
		mux.HandleFunc(route, func(w http.ResponseWriter, req *http.Request) {
			http.Redirect(w, req, nextRoute, 302)
		})
	}
	route := fmt.Sprintf("/route%v/", numRoutes)
	mux.HandleFunc(route, func(w http.ResponseWriter, req *http.Request) {
		_, _ = fmt.Fprint(w, body)
	})
	return httptest.NewServer(mux)
}

func TestParseYaml_Rest_Handle_Redirects(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = EnabledExpansions{
		Rest: true,
	}

	// The max number of redirects is currently set to 5, so we will create 5 redirects.
	numRoutes := 5
	server := startTestServer(numRoutes, "pwd123")
	defer server.Close()

	server.Client().CheckRedirect = httputil.CheckRedirectFunc
	httputil.SetClient(server.Client())

	// Construct a yaml string with the value of "__rest" set to the server's first route in the chain of redirects.
	yaml := fmt.Sprintf(`
mongodb:
  net:
    auth:
      username: "user"
      password:
        __rest: "%s/route0/"
`, server.URL)
	err := ParseYaml(cfg, bytes.NewBufferString(yaml), cfg.ConfigExpand)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	testString(t, cfg.MongoDB.Net.Auth.Password, "pwd123", "cfg.MongoDB.Net.Auth.Password")
}

// Fail if there are too many redirects in the GET request.
func TestParseYaml_Rest_Too_Many_Redirects(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = EnabledExpansions{
		Rest: true,
	}

	// The max number of redirects is currently set to 5, so to trigger a failure, we will
	// create 6 redirects.
	numRoutes := 6
	server := startTestServer(numRoutes, "pwd123")
	defer server.Close()

	server.Client().CheckRedirect = httputil.CheckRedirectFunc
	httputil.SetClient(server.Client())

	// Construct a yaml string with the value of "__rest" set to the server's first route in the chain of redirects.
	yaml := fmt.Sprintf(`
mongodb:
  net:
    auth:
      username: "user"
      password:
        __rest: "%s/route0/"
`, server.URL)
	err := ParseYaml(cfg, bytes.NewBufferString(yaml), cfg.ConfigExpand)
	if err == nil {
		t.Fatalf("test should have failed")
	}
	if !strings.Contains(err.Error(), "too many redirects (5 max)") {
		t.Fatalf("incorrect error string: %v", err.Error())
	}
}

// Test that a top-level '__exec' that evaluates to another yaml document gets successfully parsed.
func TestParseYaml_Exec_Type_Yaml_Success(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = EnabledExpansions{
		Exec: true,
		Rest: true,
	}

	tmpFile, err := ioutil.TempFile(os.TempDir(), "")
	if err != nil {
		t.Fatalf("failed to create a temporary file: %v", err)
	}

	defer os.Remove(tmpFile.Name())

	tmpFileYaml := `
mongodb:
  net:
    auth:
      username: "user"
      password: "pwd123"
`

	if _, err = tmpFile.WriteString(tmpFileYaml); err != nil {
		t.Fatalf("failed to write to the temporary file: %v", err)
	}
	yaml := fmt.Sprintf(`
__exec: "cat %s"
type: "yaml"
`, tmpFile.Name())
	err = ParseYaml(cfg, bytes.NewBufferString(yaml), cfg.ConfigExpand)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	testString(t, cfg.MongoDB.Net.Auth.Password, "pwd123", "cfg.MongoDB.Net.Auth.Password")
}

// Test that a top-level '__rest' that evaluates to another yaml document gets successfully parsed.
func TestParseYaml_Rest_Type_Yaml_Success(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = EnabledExpansions{
		Rest: true,
	}

	// Create a new reader with a yaml string.
	r := ioutil.NopCloser(bytes.NewReader([]byte(`
mongodb:
  net:
    auth:
      username: "user"
      password: "pwd123"
`)))
	client := httputil.MockClient{GetFunc: func(_ string) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}}
	httputil.SetClient(&client)

	err := ParseYaml(cfg, bytes.NewBufferString(`
__rest: "fakeurl"
type: "yaml"
`), cfg.ConfigExpand)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	testString(t, cfg.MongoDB.Net.Auth.Password, "pwd123", "cfg.MongoDB.Net.Auth.Password")
}

// Fail when a top-level expansion directive evaluates to a yaml document that includes another expansion directive.
func TestParseYaml_Type_Yaml_Recursion(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = EnabledExpansions{
		Rest: true,
		Exec: true,
	}

	// Create a new reader with a yaml string that includes an expansion directive.
	r := ioutil.NopCloser(bytes.NewReader([]byte(`
mongodb:
  net:
    auth:
      username: "user"
      password:
        __exec: "echo pwd123"
`)))
	client := httputil.MockClient{GetFunc: func(_ string) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}}
	httputil.SetClient(&client)

	err := ParseYaml(cfg, bytes.NewBufferString(`
__rest: "fakeurl"
type: "yaml"
`), cfg.ConfigExpand)
	if err == nil {
		t.Fatalf("test should have failed")
	}
	if !strings.Contains(err.Error(), "expansion directive recursion is not supported") {
		t.Fatalf("incorrect error string: %v", err.Error())
	}
}

// Fail when a top-level expansion directive evaluates to a yaml string that also has a top-level expansion directive.
func TestParseYaml_Type_Yaml_Recursion_Top_Level(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = EnabledExpansions{
		Rest: true,
	}

	// Create a new reader with a yaml string that includes an expansion directive.
	r := ioutil.NopCloser(bytes.NewReader([]byte(`
__exec: "echo this command doesn't matter'"
type: "yaml"
`)))
	client := httputil.MockClient{GetFunc: func(_ string) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}}
	httputil.SetClient(&client)

	err := ParseYaml(cfg, bytes.NewBufferString(`
__rest: "fakeurl"
type: "yaml"
`), cfg.ConfigExpand)
	if err == nil {
		t.Fatalf("test should have failed")
	}
	if !strings.Contains(err.Error(), "expansion directive recursion is not supported") {
		t.Fatalf("incorrect error string: %v", err.Error())
	}
}

// Fail when there's a top-level expansion directive but "type" is not set to "yaml".
func TestParseYaml_Type_String_Failure(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = EnabledExpansions{
		Rest: true,
	}

	// Incorrectly set the value of "type" to "string" instead of "yaml"
	err := ParseYaml(cfg, bytes.NewBufferString(`
__rest: "fakeurl"
type: "string"
`), cfg.ConfigExpand)
	if err == nil {
		t.Fatalf("test should have failed")
	}
	if !strings.Contains(err.Error(), "set {type: 'yaml'} if the config has a top-level expansion directive") {
		t.Fatalf("incorrect error string: %v", err.Error())
	}
}

// Fail when a user sets {type: "yaml"} for a non-root expansion directive.
func TestParseYaml_Type_Non_Root_Yaml_Failure(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = EnabledExpansions{
		Rest: true,
	}

	err := ParseYaml(cfg, bytes.NewBufferString(`
mongodb:
  net:
    __rest: "fakeurl"
    type: "yaml"
`), cfg.ConfigExpand)
	if err == nil {
		t.Fatalf("test should have failed")
	}
	if !strings.Contains(err.Error(), "{type: 'yaml'} is only supported for top-level expansion directives") {
		t.Fatalf("incorrect error string: %v", err.Error())
	}
}

// Fail when there are other keys after a top-level expansion directive.
func TestParseYaml_Type_Yaml_Other_Keys_Failure(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = EnabledExpansions{
		Rest: true,
	}

	err := ParseYaml(cfg, bytes.NewBufferString(`
__rest: "fakeurl"
type: "yaml"

systemLog:
  quiet: false
  verbosity: 1
  logRotate: "rename"
`), cfg.ConfigExpand)
	if err == nil {
		t.Fatalf("test should have failed")
	}
	if !strings.Contains(err.Error(), "unrecognized key") {
		t.Fatalf("incorrect error string: %v", err.Error())
	}
}

// Fail when a top-level '__rest' doesn't set {type: "yaml"}.
func TestParseYaml_Rest_Type_Default_String(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = EnabledExpansions{
		Rest: true,
	}

	err := ParseYaml(cfg, bytes.NewBufferString(`
__rest: "fakeurl"
`), cfg.ConfigExpand)
	if err == nil {
		t.Fatalf("test should have failed")
	}

	if !strings.Contains(err.Error(), "set {type: 'yaml'} if the config has a top-level expansion directive") {
		t.Fatalf("incorrect error string: %v", err.Error())
	}
}

// Fail if the value of "type" is not "string" or "yaml".
func TestParseYaml_Type_Invalid_Value(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = EnabledExpansions{
		Rest: true,
	}

	// Create a new reader with the password string.
	r := ioutil.NopCloser(bytes.NewReader([]byte("pwd123")))
	client := httputil.MockClient{GetFunc: func(_ string) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}}
	httputil.SetClient(&client)

	err := ParseYaml(cfg, bytes.NewBufferString(`
mongodb:
  net:
    auth:
      username: "user"
      password:
        __rest: "fakeurl"
        type: "blah"
`), cfg.ConfigExpand)
	if err == nil {
		t.Fatalf("test should have failed")
	}

	if !strings.Contains(err.Error(), "invalid config: {type: \"blah\"}") {
		t.Fatalf("incorrect error string: %v", err.Error())
	}
}

// Test {trim: "whitespace"}.
func TestParseYaml_Trim_Whitespace(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = EnabledExpansions{
		Exec: true,
	}
	err := ParseYaml(cfg, bytes.NewBufferString(`
mongodb:
  net:
    auth:
      username: user
      password:
        __exec: "echo \tx\ny\n"
        trim: "whitespace"
`), cfg.ConfigExpand)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testString(t, cfg.MongoDB.Net.Auth.Password, "x\ny", "cfg.MongoDB.Net.Auth.Password")
}

// Test {trim: "none"} (default).
func TestParseYaml_Trim_None(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = EnabledExpansions{
		Exec: true,
	}
	err := ParseYaml(cfg, bytes.NewBufferString(`
mongodb:
  net:
    auth:
      username: user
      password:
        __exec: "echo \tx\ny\n"
        trim: "none"
`), cfg.ConfigExpand)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testString(t, cfg.MongoDB.Net.Auth.Password, "\tx\ny\n", "cfg.MongoDB.Net.Auth.Password")
}

// Fail when the value of "trim" is not "none" or "whitespace".
func TestParseYaml_Trim_Invalid_Value(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = EnabledExpansions{
		Exec: true,
	}
	err := ParseYaml(cfg, bytes.NewBufferString(`
mongodb:
  net:
    auth:
      username: user
      password:
        __exec: "false"
        trim: "blah"
`), cfg.ConfigExpand)

	if err == nil {
		t.Fatalf("test should have failed")
	}
	if !strings.Contains(err.Error(), "invalid config: {trim: \"blah\"}") {
		t.Fatalf("incorrect error string: %v", err.Error())
	}
}

// Test a successful usage of 'digest' and 'digest_key' with '__exec'.
func TestParseYaml_Exec_Digest_Success(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = EnabledExpansions{
		Exec: true,
	}

	yaml := fmt.Sprintf(`
mongodb:
  net:
    auth:
      username: user
      password:
        __exec: "echo pwd123"
        digest: %q
        digest_key: %q
`, hash, digestKey)
	err := ParseYaml(cfg, bytes.NewBufferString(yaml), cfg.ConfigExpand)

	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	testString(t, cfg.MongoDB.Net.Auth.Password, "pwd123", "cfg.MongoDB.Net.Auth.Password")
}

// Test a successful usage of 'digest' and 'digest_key' with '__rest'.
func TestParseYaml_Rest_Digest_Success(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = EnabledExpansions{
		Rest: true,
	}

	// Create a new reader with the password string.
	r := ioutil.NopCloser(bytes.NewReader([]byte("pwd123")))
	client := httputil.MockClient{GetFunc: func(_ string) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}}
	httputil.SetClient(&client)

	yaml := fmt.Sprintf(`
mongodb:
  net:
    auth:
      username: user
      password:
        __rest: "fakeurl"
        digest: %q
        digest_key: %q
`, hash, digestKey)
	err := ParseYaml(cfg, bytes.NewBufferString(yaml), cfg.ConfigExpand)
	if err != nil {
		t.Fatalf("unexpected err: %v\n", err)
	}

	testString(t, cfg.MongoDB.Net.Auth.Password, "pwd123", "cfg.MongoDB.Net.Auth.Password")
}

// Fail when digest has an invalid length.
func TestParseYaml_Exec_Digest_Invalid_Length(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = EnabledExpansions{
		Exec: true,
	}

	yaml := fmt.Sprintf(`
mongodb:
  net:
    auth:
      username: user
      password:
        __exec: "echo pwd123"
        digest: "123"
        digest_key: %q
`, digestKey)
	err := ParseYaml(cfg, bytes.NewBufferString(yaml), cfg.ConfigExpand)

	if err == nil {
		t.Fatalf("test should have failed")
	}

	if !strings.Contains(err.Error(), "hashed result does not match digest") {
		t.Fatalf("incorrect error string: %v", err.Error())
	}
}

// Fail when digest_key has invalid characters.
func TestParseYaml_Exec_Digest_Key_Invalid_Characters(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = EnabledExpansions{
		Exec: true,
	}

	yaml := fmt.Sprintf(`
mongodb:
  net:
    auth:
      username: user
      password:
        __exec: "echo pwd123"
        digest: %q
        digest_key: "736563X26574"
`, hash)
	err := ParseYaml(cfg, bytes.NewBufferString(yaml), cfg.ConfigExpand)

	if err == nil {
		t.Fatalf("test should have failed")
	}

	if !strings.Contains(err.Error(), "encoding/hex: invalid byte: U+0058 'X'") {
		t.Fatalf("incorrect error string: %v", err.Error())
	}
}

// Fail when the digest_key is not provided.
func TestParseYaml_Exec_No_Digest_Key(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = EnabledExpansions{
		Exec: true,
	}

	yaml := fmt.Sprintf(`
mongodb:
  net:
    auth:
      username: user
      password:
        __exec: "echo pwd123"
        digest: %q
`, hash)
	err := ParseYaml(cfg, bytes.NewBufferString(yaml), cfg.ConfigExpand)

	if err == nil {
		t.Fatalf("test should have failed")
	}

	if !strings.Contains(err.Error(), "must specify 'digest_key' if 'digest' is specified") {
		t.Fatalf("incorrect error string: %v", err.Error())
	}
}

// Fail when the digest is not provided.
func TestParseYaml_Exec_No_Digest(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = EnabledExpansions{
		Exec: true,
	}

	yaml := fmt.Sprintf(`
mongodb:
  net:
    auth:
      username: user
      password:
        __exec: "echo pwd123"
        digest_key: %q
`, digestKey)
	err := ParseYaml(cfg, bytes.NewBufferString(yaml), cfg.ConfigExpand)

	if err == nil {
		t.Fatalf("test should have failed")
	}

	if !strings.Contains(err.Error(), "must specify 'digest' if 'digest_key' is specified") {
		t.Fatalf("incorrect error string: %v", err.Error())
	}
}

// Test that the digest is checked post trimming.
func TestParseYaml_Exec_Digest_Post_Trimming(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = EnabledExpansions{
		Exec: true,
	}

	yaml := fmt.Sprintf(`
mongodb:
  net:
    auth:
      username: user
      password:
        __exec: "echo \tpwd123\n"
        trim: "whitespace"
        digest: %q
        digest_key: %q
`, hash, digestKey)
	err := ParseYaml(cfg, bytes.NewBufferString(yaml), cfg.ConfigExpand)

	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	testString(t, cfg.MongoDB.Net.Auth.Password, "pwd123", "cfg.MongoDB.Net.Auth.Password")
}

// Test that digests are supported for yaml expansions.
func TestParseYaml_Exec_Digest_Yaml(t *testing.T) {
	cfg := Default()
	cfg.ConfigExpand = EnabledExpansions{
		Exec: true,
		Rest: true,
	}

	tmpFile, err := ioutil.TempFile(os.TempDir(), "")
	if err != nil {
		t.Fatalf("failed to create a temporary file: %v", err)
	}

	defer os.Remove(tmpFile.Name())

	tmpFileYaml := `
mongodb:
  net:
    auth:
      username: "user"
      password: "pwd123"
`

	if _, err = tmpFile.WriteString(tmpFileYaml); err != nil {
		t.Fatalf("failed to write to the temporary file: %v", err)
	}
	yaml := fmt.Sprintf(`
__exec: "cat %s"
type: "yaml"
digest: %q
digest_key: %q
`, tmpFile.Name(), hashYaml, digestKey)
	err = ParseYaml(cfg, bytes.NewBufferString(yaml), cfg.ConfigExpand)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	testString(t, cfg.MongoDB.Net.Auth.Password, "pwd123", "cfg.MongoDB.Net.Auth.Password")
}
