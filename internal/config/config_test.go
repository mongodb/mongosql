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
	testString(t, cfg.SystemLog.Path, "", "cfg.SystemLog.Path")
	testBool(t, cfg.SystemLog.Quiet, false, "cfg.SystemLog.Quiet")
	testInt(t, cfg.SystemLog.Verbosity, 0, "cfg.SystemLog.Verbosity")

	testString(t, cfg.Schema.Path, "", "cfg.Schema.Path")
	testUint16(t, cfg.Schema.MaxVarcharLength, 0, "cfg.Schema.MaxVarcharLength")
	testSampleMode(t, cfg.Schema.Sample.Mode, "read", "cfg.Schema.Sample.Mode")
	testString(t, string(cfg.Schema.Sample.SchemaMappingHeuristic), "lattice",
		"cfg.Schema.Sample.SchemaMappingHeuristic")
	testString(t, cfg.Schema.Sample.Source, "", "cfg.Schema.Sample.Source")
	testInt64(t, cfg.Schema.Sample.Size, 1000, "cfg.Schema.Sample.Size")
	testBool(t, cfg.Schema.Sample.PreJoin, false, "cfg.Schema.Sample.PreJoin")
	testBool(t, cfg.Schema.Sample.OptimizeViewSampling, true,
		"cfg.Schema.Sample.OptimizeViewSampling")
	testInt64(t, cfg.Schema.Sample.MaxNumColumnsPerTable, 2000,
		"cfg.Schema.Sample.MaxNumColumnsPerTable")
	testInt64(t, cfg.Schema.Sample.MaxNestedTableDepth, 50,
		"cfg.Schema.Sample.MaxNestedTableDepth")
	testStringSlice(t,
		cfg.Schema.Sample.Namespaces,
		[]string{"*.*"},
		"cfg.Schema.Sample.Namespaces",
	)
	testInt64(t, cfg.Schema.Sample.RefreshIntervalSecs, 0, "cfg.Schema.Sample.RefreshIntervalSecs")
	testString(t,
		cfg.Schema.Sample.UUIDSubtype3Encoding,
		"old",
		"cfg.Schema.Sample.UUIDSubtype3Encoding",
	)

	testUint64(t, cfg.Runtime.Memory.MaxPerServer, 0, "cfg.Runtime.Memory.MaxPerServer")
	testUint64(t, cfg.Runtime.Memory.MaxPerConnection, 0, "cfg.Runtime.Memory.MaxPerConnection")
	testUint64(t, cfg.Runtime.Memory.MaxPerStage, 0, "cfg.Runtime.Memory.MaxPerStage")

	testStringSlice(t, cfg.Net.BindIP, []string{"127.0.0.1"}, "cfg.Net.BindIP")
	testInt(t, cfg.Net.Port, 3307, "cfg.Net.Port")
	if runtime.GOOS != "windows" {
		testBool(t, cfg.Net.UnixDomainSocket.Enabled, true, "cfg.Net.UnixDomainSocket.Enabled")
		testString(t,
			cfg.Net.UnixDomainSocket.PathPrefix,
			"/tmp",
			"cfg.Net.UnixDomainSocket.PathPrefix",
		)
		testString(t,
			cfg.Net.UnixDomainSocket.FilePermissions,
			"0700",
			"cfg.Net.UnixDomainSocket.FilePermissions",
		)
	} else {
		testBool(t, cfg.Net.UnixDomainSocket.Enabled, false, "cfg.Net.UnixDomainSocket.Enabled")
		testString(t,
			cfg.Net.UnixDomainSocket.PathPrefix,
			"",
			"cfg.Net.UnixDomainSocket.PathPrefix",
		)
		testString(t,
			cfg.Net.UnixDomainSocket.FilePermissions,
			"",
			"cfg.Net.UnixDomainSocket.FilePermissions",
		)
	}

	testString(t, cfg.Net.SSL.Mode, "disabled", "cfg.Net.SSL.Mode")
	testBool(t, cfg.Net.SSL.AllowInvalidCertificates, false, "cfg.Net.SSL.AllowInvalidCertificates")
	testString(t, cfg.Net.SSL.PEMKeyFile, "", "cfg.Net.SSL.PEMKeyFile")
	testString(t, cfg.Net.SSL.PEMKeyPassword, "", "cfg.Net.SSL.PEMKeyPassword")
	testString(t, cfg.Net.SSL.CAFile, "", "cfg.Net.SSL.CAFile")
	testString(t, cfg.Net.SSL.MinimumTLSVersion, "", "cfg.Net.SSL.MinimumTLSVersion")

	testBool(t, cfg.Security.Enabled, false, "cfg.Security.Enabled")
	testString(t, cfg.Security.DefaultMechanism, "SCRAM-SHA-1", "cfg.Security.DefaultMechanism")
	testString(t, cfg.Security.DefaultSource, "admin", "cfg.Security.DefaultSource")

	testString(t, cfg.Security.GSSAPI.Hostname, "", "cfg.Security.GSSAPI.Hostname")
	testString(t, cfg.Security.GSSAPI.ServiceName, "mongosql", "cfg.Security.GSSAPI.ServiceName")

	testString(t, cfg.MongoDB.VersionCompatibility, "", "cfg.MongoDB.VersionCompatibility")
	testString(t, cfg.MongoDB.Net.URI, "mongodb://localhost:27017", "cfg.MongoDB.Net.URI")
	testInt(t, cfg.MongoDB.Net.NumConnectionsPerSession, 2,
		"cfg.MongoDB.Net.NumConnectionsPerSession")

	testString(t, cfg.MongoDB.Net.Auth.Username, "", "cfg.MongoDB.Net.Auth.Username")
	testString(t, cfg.MongoDB.Net.Auth.Password, "", "cfg.MongoDB.Net.Auth.Password")
	testString(t, cfg.MongoDB.Net.Auth.Source, "", "cfg.MongoDB.Net.Auth.Source")
	testString(t, cfg.MongoDB.Net.Auth.Mechanism, "SCRAM-SHA-1", "cfg.MongoDB.Net.Auth.Mechanism")
	testString(t, cfg.MongoDB.Net.Auth.GSSAPIServiceName, "mongodb",
		"cfg.MongoDB.Net.Auth.GSSAPIServiceName")

	testBool(t, cfg.MongoDB.Net.SSL.Enabled, false, "cfg.MongoDB.Net.SSL.Enabled")
	testBool(t, cfg.MongoDB.Net.SSL.AllowInvalidCertificates, false,
		"cfg.MongoDB.Net.SSL.AllowInvalidCertificates")

	testBool(t, cfg.MongoDB.Net.SSL.AllowInvalidHostnames, false,
		"cfg.MongoDB.Net.SSL.AllowInvalidHostnames")
	testString(t, cfg.MongoDB.Net.SSL.PEMKeyFile, "", "cfg.MongoDB.Net.SSL.PEMKeyFile")
	testString(t, cfg.MongoDB.Net.SSL.PEMKeyPassword, "", "cfg.MongoDB.Net.SSL.PEMKeyPassword")
	testString(t, cfg.MongoDB.Net.SSL.CAFile, "", "cfg.MongoDB.Net.SSL.CAFile")
	testString(t, cfg.MongoDB.Net.SSL.CRLFile, "", "cfg.MongoDB.Net.SSL.CRLFile")
	testBool(t, cfg.MongoDB.Net.SSL.FIPSMode, false, "cfg.MongoDB.Net.SSL.FIPSMode")
	testString(t, cfg.MongoDB.Net.SSL.MinimumTLSVersion, "",
		"cfg.MongoDB.Net.SSL.MinimumTLSVersion")

	testString(t, cfg.ProcessManagement.Service.Name, "mongosql",
		"cfg.ProcessManagement.Service.Name")
	testString(t, cfg.ProcessManagement.Service.DisplayName, "MongoSQL Service",
		"cfg.ProcessManagement.Service.DisplayName")
	testString(t, cfg.ProcessManagement.Service.Description,
		"MongoSQL accesses MongoDB data with SQL", "cfg.ProcessManagement.Service.Description")

	testBool(t, cfg.SetParameter.EnableTableAlterations, false,
		"cfg.SetParameter.EnableTableAlterations")

	testString(t, cfg.Debug.EnableProfiling, "", "cfg.Debug.EnableProfiling")
	testString(t, cfg.Debug.ProfileScope, "queries", "cfg.Debug.ProfileScope")

	testBool(t, cfg.SetParameter.EnableTableAlterations, false, "cfg.SetParameter.EnableTableAlterations")
	testString(t, cfg.SetParameter.MetricsBackend, "off", "cfg.SetParameter.MetricsBackend")
	testBool(t, cfg.SetParameter.OptimizeCrossJoins, true, "cfg.SetParameter.OptimizeCrossJoins")
	testBool(t, cfg.SetParameter.OptimizeEvaluations, true, "cfg.SetParameter.OptimizeEvaluations")
	testBool(t, cfg.SetParameter.OptimizeFiltering, true, "cfg.SetParameter.OptimizeFiltering")
	testBool(t, cfg.SetParameter.OptimizeInnerJoins, true, "cfg.SetParameter.OptimizeInnerJoins")
	testBool(t, cfg.SetParameter.OptimizeSelfJoins, true, "cfg.SetParameter.OptimizeSelfJoins")
	testBool(t, cfg.SetParameter.OptimizeViewSampling, true, "cfg.SetParameter.OptimizeViewSampling")
	testString(t, cfg.SetParameter.PolymorphicTypeConversionMode, "off", "cfg.SetParameter.PolymorphicTypeConversionMode")
	testBool(t, cfg.SetParameter.Pushdown, true, "cfg.SetParameter.Pushdown")
	testString(t, cfg.SetParameter.TypeConversionMode, "mongosql", "cfg.SetParameter.TypeConversionMode")
}

func TestLoad(t *testing.T) {
	args := []string{"--config", "testdata/sample.conf"}

	cfg, _, err := Load(args)
	if err != nil {
		t.Fatalf("expected no error, but got '%v'", err)
	}

	testBool(t, cfg.SystemLog.LogAppend, true, "cfg.SystemLog.LogAppend")
	testString(t, string(cfg.SystemLog.LogRotate), string(log.Reopen), "cfg.SystemLog.LogRotate")
	testString(t, cfg.SystemLog.Path, "temp", "cfg.SystemLog.Path")
	testBool(t, cfg.SystemLog.Quiet, true, "cfg.SystemLog.Quiet")
	testInt(t, cfg.SystemLog.Verbosity, 2, "cfg.SystemLog.Verbosity")

	testString(t, cfg.Schema.Path, "/var/test", "cfg.Schema.Path")
	testUint16(t, cfg.Schema.MaxVarcharLength, 1000, "cfg.Schema.MaxVarcharLength")
	testSampleMode(t, cfg.Schema.Sample.Mode, "write", "cfg.Schema.Sample.Mode")
	testString(t, string(cfg.Schema.Sample.SchemaMappingHeuristic), "majority",
		"cfg.Schema.Sample.SchemaMappingHeuristic")
	testString(t, cfg.Schema.Sample.Source, "sampleDb", "cfg.Schema.Sample.Source")
	testInt64(t, cfg.Schema.Sample.Size, 969, "cfg.Schema.Sample.Size")
	testBool(t, cfg.Schema.Sample.PreJoin, true, "cfg.Schema.Sample.PreJoin")
	testStringSlice(t, cfg.Schema.Sample.Namespaces, []string{"foo.*", "*.bar"},
		"cfg.Schema.Sample.Namespaces")
	testInt64(t, cfg.Schema.Sample.RefreshIntervalSecs, 983,
		"cfg.Schema.Sample.RefreshIntervalSecs")
	testString(t, cfg.Schema.Sample.UUIDSubtype3Encoding, "java",
		"cfg.Schema.Sample.UUIDSubtype3Encoding",
	)

	testUint64(t, cfg.Runtime.Memory.MaxPerServer, 2000000, "cfg.Runtime.Memory.MaxPerServer")
	testUint64(t, cfg.Runtime.Memory.MaxPerConnection, 1000000,
		"cfg.Runtime.Memory.MaxPerConnection")
	testUint64(t, cfg.Runtime.Memory.MaxPerStage, 102400, "cfg.Runtime.Memory.MaxPerStage")

	testStringSlice(t, cfg.Net.BindIP, []string{"192.168.20.1"}, "cfg.Net.BindIP")
	testInt(t, cfg.Net.Port, 3306, "cfg.Net.Port")
	testBool(t, cfg.Net.UnixDomainSocket.Enabled, false, "cfg.Net.UnixDomainSocket.Enabled")
	testString(t, cfg.Net.UnixDomainSocket.PathPrefix, "/var",
		"cfg.Net.UnixDomainSocket.PathPrefix")
	testString(t, cfg.Net.UnixDomainSocket.FilePermissions, "0600",
		"cfg.Net.UnixDomainSocket.FilePermissions")

	testString(t, cfg.Net.SSL.Mode, "requireSSL", "cfg.Net.SSL.Mode")
	testBool(t, cfg.Net.SSL.AllowInvalidCertificates, true, "cfg.Net.SSL.AllowInvalidCertificates")
	testString(t, cfg.Net.SSL.PEMKeyFile, "pemkeyfile", "cfg.Net.SSL.PEMKeyFile")
	testString(t, cfg.Net.SSL.PEMKeyPassword, "pemkeypassword", "cfg.Net.SSL.PEMKeyPassword")
	testString(t, cfg.Net.SSL.CAFile, "cafile", "cfg.Net.SSL.CAFile")
	testString(t, cfg.Net.SSL.MinimumTLSVersion, "TLS1_0", "cfg.Net.SSL.MinimumTLSVersion")

	testBool(t, cfg.Security.Enabled, true, "cfg.Security.Enabled")
	testString(t, cfg.Security.DefaultMechanism, "GSSAPI", "cfg.Security.DefaultMechanism")
	testString(t, cfg.Security.DefaultSource, "$external", "cfg.Security.DefaultSource")

	testString(t, cfg.Security.GSSAPI.Hostname, "something", "cfg.Security.GSSAPI.Hostname")
	testString(t, cfg.Security.GSSAPI.ServiceName, "awesome", "cfg.Security.GSSAPI.ServiceName")

	testString(t, cfg.MongoDB.VersionCompatibility, "3.2", "cfg.MongoDB.VersionCompatibility")
	testString(t, cfg.MongoDB.Net.URI, "mongodb://hostname:27018", "cfg.MongoDB.Net.URI")
	testInt(t, cfg.MongoDB.Net.NumConnectionsPerSession, 3,
		"cfg.MongoDB.Net.NumConnectionsPerSession")

	testString(t, cfg.MongoDB.Net.Auth.Username, "user", "cfg.MongoDB.Net.Auth.Username")
	testString(t, cfg.MongoDB.Net.Auth.Password, "pass", "cfg.MongoDB.Net.Auth.Password")
	testString(t, cfg.MongoDB.Net.Auth.Source, "adminer", "cfg.MongoDB.Net.Auth.Source")
	testString(t, cfg.MongoDB.Net.Auth.Mechanism, "SCRAM-SHA-256", "cfg.MongoDB.Net.Auth.Mechanism")
	testString(t, cfg.MongoDB.Net.Auth.GSSAPIServiceName, "hola",
		"cfg.MongoDB.Net.Auth.GSSAPIServiceName")

	testBool(t, cfg.MongoDB.Net.SSL.Enabled, true, "cfg.MongoDB.Net.SSL.Enabled")
	testBool(t, cfg.MongoDB.Net.SSL.AllowInvalidCertificates, true,
		"cfg.MongoDB.Net.SSL.AllowInvalidCertificates")
	testBool(t, cfg.MongoDB.Net.SSL.AllowInvalidHostnames, true,
		"cfg.MongoDB.Net.SSL.AllowInvalidHostnames")
	testString(t, cfg.MongoDB.Net.SSL.PEMKeyFile, "mongopemkeyfile",
		"cfg.MongoDB.Net.SSL.PEMKeyFile")
	testString(t, cfg.MongoDB.Net.SSL.PEMKeyPassword, "mongopemkeypassword",
		"cfg.MongoDB.Net.SSL.PEMKeyPassword")
	testString(t, cfg.MongoDB.Net.SSL.CAFile, "mongocafile", "cfg.MongoDB.Net.SSL.CAFile")
	testString(t, cfg.MongoDB.Net.SSL.CRLFile, "mongocrlfile", "cfg.MongoDB.Net.SSL.CRLFile")
	testBool(t, cfg.MongoDB.Net.SSL.FIPSMode, true, "cfg.MongoDB.Net.SSL.FIPSMode")
	testString(t, cfg.MongoDB.Net.SSL.MinimumTLSVersion, "TLS1_0",
		"cfg.MongoDB.Net.SSL.MinimumTLSVersion")

	testString(t, cfg.ProcessManagement.Service.Name, "oompa",
		"cfg.ProcessManagement.Service.Name")
	testString(t, cfg.ProcessManagement.Service.DisplayName, "loompa",
		"cfg.ProcessManagement.Service.DisplayName")
	testString(t, cfg.ProcessManagement.Service.Description, "doompa tee do",
		"cfg.ProcessManagement.Service.Description")

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

	testString(t, cfg.Debug.EnableProfiling, "", "cfg.Debug.EnableProfiling")
	testString(t, cfg.Debug.ProfileScope, "queries", "cfg.Debug.ProfileScope")
}

func TestLoadWithCLIArgs(t *testing.T) {
	args := []string{
		"--config", "testdata/sample.conf",
		"-vv",
		"--mongo-minimumTLSVersion", "TLS1_2",
	}

	cfg, _, err := Load(args)
	if err != nil {
		t.Fatalf("expected no error, but got '%v'", err)
	}

	// 1 from the args, as opposed to the 2 from the file
	testInt(t, cfg.SystemLog.Verbosity, 2, "cfg.SystemLog.Verbosity")
	testString(t, cfg.MongoDB.Net.SSL.MinimumTLSVersion, "TLS1_2",
		"cfg.MongoDB.Net.SSL.MinimumTLSVersion")
}

func TestToJSON(t *testing.T) {
	cfg := Default()

	cfg.Config = "funny"
	cfg.SystemLog.LogAppend = true
	cfg.Net.SSL.PEMKeyPassword = "harumph"

	actual := ToJSON(cfg)
	expected := `{config: "funny", systemLog: {logAppend: true}, ` +
		`net: {ssl: {PEMKeyPassword: "<protected>"}}}`
	if actual != expected {
		t.Fatalf("expected '%s', but got '%s'", expected, actual)
	}
}

func TestToJSON_SetParameter(t *testing.T) {
	cfg := Default()

	cfg.Config = "funny"
	cfg.SystemLog.LogAppend = true
	cfg.Net.SSL.PEMKeyPassword = "harumph"
	cfg.SetParameter.EnableTableAlterations = true

	actual := ToJSON(cfg)

	expected := `{config: "funny", systemLog: {logAppend: true}, ` +
		`net: {ssl: {PEMKeyPassword: "<protected>"}}, setParameter: {enableTableAlterations: true}}`
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

func TestValidate_Invalid_Client_MinimumTLSVersion(t *testing.T) {
	cfg := Default()
	cfg.Net.SSL.Mode = "allowSSL"
	cfg.Net.SSL.PEMKeyFile = "harumph"
	cfg.Net.SSL.MinimumTLSVersion = "bi"

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected := "unsupported client minimum TLS version 'bi'"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}
}

func TestValidate_Invalid_Server_MinimumTLSVersion(t *testing.T) {
	cfg := Default()
	cfg.MongoDB.Net.SSL.Enabled = true
	cfg.MongoDB.Net.SSL.PEMKeyFile = "harumph"
	cfg.MongoDB.Net.SSL.MinimumTLSVersion = "connector"

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected := "unsupported mongo minimum TLS version 'connector'"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}
}

func TestValidate_Invalid_SampleAuth_Mechanism(t *testing.T) {
	cfg := Default()
	cfg.Schema.Sample.Source = "test"

	cfg.MongoDB.Net.Auth.Mechanism = "foo"

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected := "unsupported sample authentication mechanism 'foo'"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}
}

func TestValidate_Invalid_Sample_MaxNumColumnsPerTable(t *testing.T) {
	cfg := Default()
	cfg.Schema.Sample.MaxNumColumnsPerTable = 0
	cfg.Schema.Sample.Source = "test"

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected := "invalid sample max number of columns per table: 0"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}
}

func TestValidate_Invalid_Sample_MaxNestedTableDepth(t *testing.T) {
	cfg := Default()
	cfg.Schema.Sample.MaxNestedTableDepth = -1
	cfg.Schema.Sample.Source = "test"

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected := "invalid sample max nested table depth: -1"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}
}

func TestValidate_Valid_SampleAuth_Mechanism(t *testing.T) {
	cfg := Default()
	cfg.Schema.Sample.Source = "test"

	cfg.MongoDB.Net.Auth.Mechanism = "SCRAM-SHA-1"
	err := Validate(cfg)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}

	cfg.MongoDB.Net.Auth.Mechanism = "SCRAM-SHA-256"
	err = Validate(cfg)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}

	cfg.MongoDB.Net.Auth.Mechanism = "PLAIN"
	err = Validate(cfg)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}

	cfg.MongoDB.Net.Auth.Mechanism = "GSSAPI"
	err = Validate(cfg)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
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

func TestValidate_Sample_Invalid_SchemaMappingHeuristic(t *testing.T) {
	tests := []struct {
		schemaMappingHeuristic MappingHeuristic
		valid                  bool
	}{
		{LatticeMappingMode, true},
		{MajorityMappingMode, true},
		{MappingHeuristic("nope"), false},
		{MappingHeuristic(""), false},
	}

	for _, test := range tests {
		t.Run(string(test.schemaMappingHeuristic), func(t *testing.T) {
			cfg := Default()
			cfg.Schema.Sample.SchemaMappingHeuristic = test.schemaMappingHeuristic

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
		{ns: []string{"one"}, valid: false},
		{ns: []string{"one.*"}, valid: true},
		{ns: []string{"*.two"}, valid: true},
		{ns: []string{"one.two"}, valid: true},
		{ns: []string{"one.two", "three"}, valid: false},
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

	expected := "sample mode 'read' with a non-empty sample source cannot specify " +
		"a sample refresh interval"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but go '%s'", expected, err)
	}
}

func TestValidate_auth_specified_without_admin_creds(t *testing.T) {
	cfg := Default()
	cfg.Security.Enabled = true

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected := "when authentication is enabled, admin credentials must be " +
		"provided with --mongo-username and --mongo-password or in a config " +
		"file at 'mongodb.net.auth'"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}

	cfg.MongoDB.Net.Auth.Username = ""
	cfg.MongoDB.Net.Auth.Password = "foo"

	err = Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected = "when authentication is enabled, admin credentials must be " +
		"provided with --mongo-username and --mongo-password or in a config " +
		"file at 'mongodb.net.auth'"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}

	cfg.MongoDB.Net.Auth.Username = "foo"
	cfg.MongoDB.Net.Auth.Password = ""

	err = Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected = "when authentication is enabled, admin credentials must be " +
		"provided with --mongo-username and --mongo-password or in a config " +
		"file at 'mongodb.net.auth'"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}

	cfg.MongoDB.Net.Auth.Username = ""
	cfg.MongoDB.Net.Auth.Password = "blah"
	cfg.MongoDB.Net.Auth.Mechanism = "GSSAPI"

	err = Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected = "GSSAPI authentication is enabled and no username was supplied. " +
		"Please provide credentials using --mongo-username and --mongo-password or in a " +
		"config file at 'mongodb.net.auth'. In lieu of a password, you can use a keytab" +
		" or a credentials cache."
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}

	cfg.MongoDB.Net.Auth.Username = ""
	cfg.MongoDB.Net.Auth.Password = ""
	cfg.MongoDB.Net.Auth.Mechanism = "GSSAPI"

	err = Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected = "GSSAPI authentication is enabled and no username was supplied. " +
		"Please provide credentials using --mongo-username and --mongo-password or in a " +
		"config file at 'mongodb.net.auth'. In lieu of a password, you can use a keytab" +
		" or a credentials cache."
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}

	cfg.MongoDB.Net.Auth.Username = "blah"
	cfg.MongoDB.Net.Auth.Password = ""
	cfg.MongoDB.Net.Auth.Mechanism = "GSSAPI"

	err = Validate(cfg)
	if err != nil {
		t.Fatalf("expected no error, but got one %v", err)
	}

}

func TestValidate_admin_creds_specified_but_auth_disabled(t *testing.T) {
	cfg := Default()
	cfg.MongoDB.Net.Auth.Username = "foo"
	cfg.Schema.Sample.Source = "test"

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected := "when specifying admin authentication credentials, auth must " +
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

	expected = "when specifying admin authentication credentials, auth must " +
		"be enabled with --auth or in a config file at 'security.enabled'"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}

	cfg.MongoDB.Net.Auth.Username = "foo"
	cfg.MongoDB.Net.Auth.Password = ""

	err = Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected = "when specifying admin authentication credentials, auth must " +
		"be enabled with --auth or in a config file at 'security.enabled'"
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}
}

func TestValidate_SSL_options_specified_but_disabled(t *testing.T) {
	getDefaultConfig := func(isClientTest bool) *Config {
		cfg := Default()
		// only setting this to simplify the client and server SSL option testing.
		cfg.MongoDB.Net.SSL.Enabled = isClientTest
		cfg.Schema.Path = "something"
		return cfg
	}

	type cfgMaker func() *Config

	cfgMakers := []cfgMaker{
		// client side
		func() *Config {
			cfg := getDefaultConfig(true)
			cfg.Net.SSL.AllowInvalidCertificates = true
			return cfg
		},
		func() *Config {
			cfg := getDefaultConfig(true)
			cfg.Net.SSL.PEMKeyFile = "hello"
			return cfg
		},
		func() *Config {
			cfg := getDefaultConfig(true)
			cfg.Net.SSL.PEMKeyPassword = "hello"
			return cfg
		},
		func() *Config {
			cfg := getDefaultConfig(true)
			cfg.Net.SSL.CAFile = "hello"
			return cfg
		},
		func() *Config {
			cfg := getDefaultConfig(true)
			cfg.Net.SSL.MinimumTLSVersion = "TLS1_0"
			return cfg
		},

		// server side
		func() *Config {
			cfg := getDefaultConfig(false)
			cfg.MongoDB.Net.SSL.CAFile = "hello"
			return cfg
		},
		func() *Config {
			cfg := getDefaultConfig(false)
			cfg.MongoDB.Net.SSL.CRLFile = "hello"
			return cfg
		},
		func() *Config {
			cfg := getDefaultConfig(false)
			cfg.MongoDB.Net.SSL.PEMKeyFile = "hello"
			return cfg
		},
		func() *Config {
			cfg := getDefaultConfig(false)
			cfg.MongoDB.Net.SSL.PEMKeyPassword = "hello"
			return cfg
		},
		func() *Config {
			cfg := getDefaultConfig(false)
			cfg.MongoDB.Net.SSL.AllowInvalidCertificates = true
			return cfg
		},
		func() *Config {
			cfg := getDefaultConfig(false)
			cfg.MongoDB.Net.SSL.AllowInvalidHostnames = true
			return cfg
		},
		func() *Config {
			cfg := getDefaultConfig(false)
			cfg.MongoDB.Net.SSL.MinimumTLSVersion = "TLS1_0"
			return cfg
		},
	}

	clientExpected := "when specifying SSL options, SSL must be enabled with --sslMode or in a " +
		"configuration file at 'net.ssl.mode'"

	serverExpected := "when specifying MongoDB SSL options, SSL must be enabled with --mongo-ssl " +
		"or in a configuration file at 'mongodb.net.ssl.enabled'"

	for _, cfgMaker := range cfgMakers {
		cfg := cfgMaker()
		err := Validate(cfg)
		if err == nil {
			t.Fatalf("expected an error, but got none")
		}
		expected := clientExpected
		if !cfg.MongoDB.Net.SSL.Enabled {
			expected = serverExpected
		}
		if err.Error() != expected {
			t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
		}
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

	expected := fmt.Sprintf(
		"invalid number of MongoDB connections: 0 (must be between %d and %d)",
		MinConnections, MaxConnections,
	)
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

	expected := fmt.Sprintf(
		"invalid number of MongoDB connections: 1000 (must be between %d and %d)",
		MinConnections, MaxConnections,
	)
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}
}

func TestValidate_Runtime_Memory_MaxPerConnection_Larger_Than_MaxPerServer(t *testing.T) {
	cfg := Default()
	cfg.Runtime.Memory.MaxPerServer = 10
	cfg.Runtime.Memory.MaxPerConnection = 20

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected := fmt.Sprintf("runtime.memory.maxPerServer(10) must be greater than or equal" +
		" to runtime.memory.maxPerConnection(20)")
	if err.Error() != expected {
		t.Fatalf("expected error to be '%s', but got '%s'", expected, err)
	}
}

func TestValidate_Runtime_Memory_MaxPerStage_Larger_Than_MaxPerConnection(t *testing.T) {
	cfg := Default()
	cfg.Runtime.Memory.MaxPerConnection = 10
	cfg.Runtime.Memory.MaxPerStage = 20

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	expected := fmt.Sprintf("runtime.memory.maxPerConnection(10) must be greater than or equal" +
		" to runtime.memory.maxPerStage(20)")
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
