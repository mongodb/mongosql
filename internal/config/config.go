package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"
	"runtime"
	"strconv"

	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
)

var (
	isDarwin                = runtime.GOOS == "darwin"
	isWindows               = runtime.GOOS == "windows"
	supportedAuthMechanisms = []string{
		"SCRAM-SHA-1",
		"PLAIN",
		"GSSAPI",
	}
	supportedTLSVersions = []string{
		"TLS1_0",
		"TLS1_1",
		"TLS1_2",
	}
)

// Constants for the allowed values of NumConnectionsPerSession
const (
	MinConnections = 2
	MaxConnections = 10
)

// Load creates a new configuration from the specified arguments
// and potentially loads from a separate config source as specified
// on the command line.
func Load(args []string) (*Config, []string, error) {
	cfg := Default()
	args, err := ParseArgs(cfg, args)
	if err != nil {
		return nil, nil, err
	}

	if cfg.Config != "" {
		var yaml []byte
		yaml, err = ioutil.ReadFile(cfg.Config)
		if err != nil {
			return nil, nil, err
		}
		if len(yaml) == 0 {
			// we're done. An empty file shouldn't cause an error.
			return cfg, args, nil
		}

		// we'll start over with a new default set and then re-parse
		// the args because they should override anything specified in
		// the yaml file.
		cfg = Default()
		err = ParseYaml(cfg, bytes.NewReader(yaml))
		if err != nil {
			return nil, nil, fmt.Errorf("unable to parse configuration file: %s", err)
		}
		args, err = ParseArgs(cfg, args)
		if err != nil {
			return nil, nil, err
		}

		if cfg.MongoDB.Net.Auth.Mechanism == "" {
			cfg.MongoDB.Net.Auth.Mechanism = cfg.Security.DefaultMechanism
		}

		if cfg.MongoDB.Net.Auth.Source == "" {
			cfg.MongoDB.Net.Auth.Source = cfg.Security.DefaultSource
		}

		if cfg.Runtime.Memory.MaxPerConnection == 0 {
			cfg.Runtime.Memory.MaxPerConnection = cfg.Runtime.Memory.MaxPerServer
		}

		if cfg.Runtime.Memory.MaxPerStage == 0 {
			cfg.Runtime.Memory.MaxPerStage = cfg.Runtime.Memory.MaxPerConnection
		}
	}

	return cfg, args, err
}

// Default returns the default configuration.
func Default() *Config {
	cfg := &Config{}

	cfg.Net.BindIP = []string{"127.0.0.1"}
	cfg.Net.Port = 3307

	cfg.Net.SSL.Mode = "disabled"
	cfg.Net.SSL.MinimumTLSVersion = "TLS1_1"

	if !isWindows {
		cfg.Net.UnixDomainSocket.Enabled = true
		cfg.Net.UnixDomainSocket.PathPrefix = "/tmp"
		cfg.Net.UnixDomainSocket.FilePermissions = "0700"
	}

	cfg.Security.DefaultMechanism = "SCRAM-SHA-1"
	cfg.Security.DefaultSource = "admin"
	cfg.Security.GSSAPI.ServiceName = "mongosql"

	cfg.MongoDB.Net.URI = "mongodb://localhost:27017"
	cfg.MongoDB.Net.NumConnectionsPerSession = 2

	cfg.MongoDB.Net.Auth.GSSAPIServiceName = "mongodb"
	cfg.MongoDB.Net.Auth.Mechanism = "SCRAM-SHA-1"
	cfg.MongoDB.Net.SSL.MinimumTLSVersion = "TLS1_2"

	cfg.ProcessManagement.Service.Name = "mongosql"
	cfg.ProcessManagement.Service.DisplayName = "MongoSQL Service"
	cfg.ProcessManagement.Service.Description = "MongoSQL accesses MongoDB data with SQL"

	cfg.Schema.Sample.Size = 1000
	cfg.Schema.Sample.Mode = "read"
	cfg.Schema.Sample.Namespaces = []string{"*.*"}
	cfg.Schema.Sample.RefreshIntervalSecs = 0
	cfg.Schema.Sample.UUIDSubtype3Encoding = "old"

	cfg.SystemLog.LogRotate = log.Rename

	cfg.Debug.ProfileScope = "queries"

	return cfg
}

// ToJSON converts the config to json, only outputting
// fields that are not the default.
func ToJSON(cfg *Config) string {
	var w bytes.Buffer
	w.WriteString("{")
	toJSON(&w, reflect.ValueOf(*Default()), reflect.ValueOf(*cfg))
	w.WriteString("}")
	return w.String()
}

// Validate ensure that a config is valid.
func Validate(cfg *Config) error {
	switch cfg.Net.SSL.Mode {
	case "disabled":
		if cfg.Net.SSL.AllowInvalidCertificates ||
			cfg.Net.SSL.PEMKeyFile != "" ||
			cfg.Net.SSL.PEMKeyPassword != "" ||
			cfg.Net.SSL.CAFile != "" {
			return fmt.Errorf("when specifying SSL options, SSL must be enabled with --sslMode " +
				"or in a configuration file at 'net.ssl.mode'")
		}
	case "allowSSL", "requireSSL":
		if cfg.Net.SSL.PEMKeyFile == "" {
			return fmt.Errorf("need sslPEMKeyFile when SSL is enabled")
		}
	default:
		return fmt.Errorf("invalid sslMode %s", cfg.Net.SSL.Mode)
	}

	if !cfg.MongoDB.Net.SSL.Enabled && (cfg.MongoDB.Net.SSL.CAFile != "" ||
		cfg.MongoDB.Net.SSL.CRLFile != "" ||
		cfg.MongoDB.Net.SSL.PEMKeyFile != "" ||
		cfg.MongoDB.Net.SSL.PEMKeyPassword != "") {
		return fmt.Errorf("when specifying MongoDB SSL options, SSL must be enabled with " +
			"--mongo-ssl or in a configuration file at 'mongodb.net.ssl.enabled'")
	}

	if cfg.MongoDB.Net.SSL.FIPSMode && isDarwin {
		return fmt.Errorf("this version of mongosqld was not compiled with FIPS support")
	}

	if isWindows {
		if cfg.Net.UnixDomainSocket.Enabled ||
			cfg.Net.UnixDomainSocket.PathPrefix != "" ||
			cfg.Net.UnixDomainSocket.FilePermissions != "" {
			return fmt.Errorf("unix domain sockets are not supported on windows")
		}
	}

	if cfg.Net.UnixDomainSocket.FilePermissions != "" {
		if _, err := strconv.ParseInt(cfg.Net.UnixDomainSocket.FilePermissions, 8, 64); err != nil {
			return fmt.Errorf("filePermissions must be valid octal")
		}
	}

	if !cfg.Security.Enabled && (cfg.MongoDB.Net.Auth.Username != "" ||
		cfg.MongoDB.Net.Auth.Password != "") {
		return fmt.Errorf("when specifying admin authentication " +
			"credentials, auth must be enabled with --auth or in " +
			"a config file at 'security.enabled'")
	}

	isEmptyUserName := cfg.MongoDB.Net.Auth.Username == ""
	isEmptyPassword := cfg.MongoDB.Net.Auth.Password == ""
	isGssapi := cfg.MongoDB.Net.Auth.Mechanism == "GSSAPI"
	if cfg.Security.Enabled {
		if (isEmptyUserName || isEmptyPassword) && !isGssapi {
			return fmt.Errorf("when authentication is enabled, admin credentials " +
				"must be provided with --mongo-username and --mongo-password or " +
				"in a config file at 'mongodb.net.auth'")
		}

		if isEmptyUserName && isGssapi {
			return fmt.Errorf("GSSAPI authentication is enabled and no username was supplied. " +
				"Please provide credentials using --mongo-username and --mongo-password or in a " +
				"config file at 'mongodb.net.auth'. In lieu of a password, you can use a keytab" +
				" or a credentials cache.")
		}

	}

	if cfg.MongoDB.Net.Auth.Mechanism != "" &&
		!util.SliceContains(supportedAuthMechanisms,
			cfg.MongoDB.Net.Auth.Mechanism) {
		return fmt.Errorf("unsupported sample authentication "+
			"mechanism '%v'", cfg.MongoDB.Net.Auth.Mechanism)
	}

	if cfg.Net.SSL.MinimumTLSVersion != "" &&
		!util.SliceContains(supportedTLSVersions, cfg.Net.SSL.MinimumTLSVersion) {
		return fmt.Errorf("unsupported client minimum TLS version '%v'",
			cfg.Net.SSL.MinimumTLSVersion)
	}

	if cfg.MongoDB.Net.SSL.MinimumTLSVersion != "" &&
		!util.SliceContains(supportedTLSVersions, cfg.MongoDB.Net.SSL.MinimumTLSVersion) {
		return fmt.Errorf("unsupported mongo minimum TLS version '%v'",
			cfg.MongoDB.Net.SSL.MinimumTLSVersion)
	}

	if cfg.MongoDB.Net.NumConnectionsPerSession < MinConnections ||
		cfg.MongoDB.Net.NumConnectionsPerSession > MaxConnections {
		return fmt.Errorf("invalid number of MongoDB connections: %d "+
			"(must be between %d and %d)", cfg.MongoDB.Net.NumConnectionsPerSession,
			MinConnections, MaxConnections)
	}

	if cfg.Schema.Path != "" && cfg.Schema.Sample.Source != "" {
		return fmt.Errorf("must specify only one of sample source or schema")
	}

	if cfg.Schema.Sample.Size < 0 {
		return fmt.Errorf("invalid sample size: %d", cfg.Schema.Sample.Size)
	}

	if _, err := util.NewMatcher(cfg.Schema.Sample.Namespaces); err != nil {
		return fmt.Errorf("invalid specification: %v", err)
	}

	if cfg.Schema.Sample.Mode != ReadSampleMode && cfg.Schema.Sample.Mode != WriteSampleMode {
		return fmt.Errorf("invalid schema sample mode: %v", cfg.Schema.Sample.Mode)
	}

	if cfg.Schema.Sample.Source != "" {
		if err := util.ValidateDBName(cfg.Schema.Sample.Source); err != nil {
			return fmt.Errorf("invalid sample source: %v", err)
		}
	}

	if cfg.Schema.Sample.Source == "" && cfg.Schema.Sample.Mode == WriteSampleMode {
		return fmt.Errorf("sample mode 'write' requires a non-empty sample source")
	}

	if cfg.Schema.Sample.Source != "" &&
		cfg.Schema.Sample.Mode == ReadSampleMode &&
		cfg.Schema.Sample.RefreshIntervalSecs > 0 {

		return fmt.Errorf("sample mode 'read' with a non-empty sample source cannot specify a " +
			"sample refresh interval")
	}

	switch cfg.SystemLog.LogRotate {
	case log.Rename:
		// this is valid
	case log.Reopen:
		if !cfg.SystemLog.LogAppend {
			return fmt.Errorf("When using 'reopen' log rotation strategy, " +
				"logAppend must be turned on with --logAppend or in a " +
				"configuration file at systemLog.logAppend")
		}
	default:
		return fmt.Errorf("Unsupported log rotation strategy '%s'", cfg.SystemLog.LogRotate)
	}

	switch cfg.Debug.EnableProfiling {
	case "cpu", "":
		// valid
	default:
		return fmt.Errorf("invalid profiling type %q", cfg.Debug.EnableProfiling)
	}

	switch cfg.Debug.ProfileScope {
	case "queries", "all":
		// valid
	default:
		return fmt.Errorf("invalid profile scope %q", cfg.Debug.ProfileScope)
	}

	if cfg.Runtime.Memory.MaxPerServer > 0 &&
		cfg.Runtime.Memory.MaxPerConnection > cfg.Runtime.Memory.MaxPerServer {
		return fmt.Errorf("runtime.memory.maxPerServer(%d) must be greater than or equal"+
			" to runtime.memory.maxPerConnection(%d)",
			cfg.Runtime.Memory.MaxPerServer,
			cfg.Runtime.Memory.MaxPerConnection)
	}

	if cfg.Runtime.Memory.MaxPerConnection > 0 &&
		cfg.Runtime.Memory.MaxPerStage > cfg.Runtime.Memory.MaxPerConnection {
		return fmt.Errorf("runtime.memory.maxPerConnection(%d) must be greater than or equal"+
			" to runtime.memory.maxPerStage(%d)",
			cfg.Runtime.Memory.MaxPerConnection,
			cfg.Runtime.Memory.MaxPerStage)
	}

	return nil
}

// Config is the root of the configuration tree for mongosqld.
type Config struct {
	// Config is the file to load extra configuration from.
	Config string

	SystemLog         SystemLog
	Schema            Schema
	Runtime           Runtime
	Net               Net
	Security          Security
	MongoDB           MongoDB `config:"mongodb"`
	ProcessManagement ProcessManagement
	SetParameter      SetParameter
	Debug             Debug
}

// SystemLog holds logging configuration.
type SystemLog struct {
	LogAppend bool
	LogRotate log.RotationStrategy
	Path      string
	Quiet     bool
	Verbosity int
}

// Runtime holds runtime configuration.
type Runtime struct {
	Memory RuntimeMemory
}

// RuntimeMemory holds configuration for memory.
type RuntimeMemory struct {
	MaxPerServer     uint64
	MaxPerConnection uint64
	MaxPerStage      uint64
}

// Schema holds schema configuration.
type Schema struct {
	Path             string
	MaxVarcharLength uint16
	Sample           SchemaSampleOptions `config:"sample"`
}

// SampleMode is an enum representing mongosqld's sampling modes.
type SampleMode string

// Values for SampleMode
const (
	ReadSampleMode  = "read"
	WriteSampleMode = "write"
)

// SchemaSampleOptions holds schema sampling configuration.
type SchemaSampleOptions struct {
	Source               string     `config:"source"`
	Mode                 SampleMode `config:"mode"`
	Size                 int64      `config:"size"`
	PreJoined            bool       `config:"-"`
	Namespaces           []string   `config:"namespaces"`
	RefreshIntervalSecs  int64      `config:"refreshIntervalSecs"`
	UUIDSubtype3Encoding string     `config:"uuidSubtype3Encoding"`
}

// Net holds network related configuration.
type Net struct {
	BindIP           []string `config:"bindIp"`
	Port             int
	UnixDomainSocket NetUnixDomainSocket
	SSL              NetSSL `config:"ssl"`
}

// NetUnixDomainSocket holds configuration for unix domain sockets.
type NetUnixDomainSocket struct {
	Enabled         bool
	PathPrefix      string
	FilePermissions string
}

// NetSSL holds configuration for SSL with a client.
type NetSSL struct {
	Mode                     string
	AllowInvalidCertificates bool
	PEMKeyFile               string `config:"PEMKeyFile"`
	PEMKeyPassword           string `config:"PEMKeyPassword,protected"`
	CAFile                   string `config:"CAFile"`
	MinimumTLSVersion        string `config:"minimumTLSVersion"`
}

// Security holds configuration for security with a client.
type Security struct {
	Enabled          bool
	DefaultMechanism string
	DefaultSource    string
	GSSAPI           SecurityGSSAPI `config:"gssapi"`
}

// SecurityGSSAPI holds configuration for hosting GSSAPI authentication.
type SecurityGSSAPI struct {
	Hostname    string
	ServiceName string `config:"serviceName"`
}

// MongoDB holds configuration for connecting to MongoDB.
type MongoDB struct {
	VersionCompatibility string
	Net                  MongoDBNet
}

// MongoDBNet holds confifuration for network communication with MongoDB.
type MongoDBNet struct {
	URI                      string         `config:"uri"`
	SSL                      MongoDBNetSSL  `config:"ssl"`
	Auth                     MongoDBNetAuth `config:"auth"`
	NumConnectionsPerSession int            `config:"numConnectionsPerSession"`
}

// MongoDBNetSSL holds configuration for SSL with MongoDB.
type MongoDBNetSSL struct {
	Enabled                  bool
	AllowInvalidCertificates bool
	AllowInvalidHostnames    bool
	MinimumTLSVersion        string `config:"minimumTLSVersion"`
	PEMKeyFile               string `config:"PEMKeyFile"`
	PEMKeyPassword           string `config:"PEMKeyPassword,protected"`
	CAFile                   string `config:"CAFile"`
	CRLFile                  string `config:"CRLFile"`
	FIPSMode                 bool   `config:"FIPSMode"`
}

// ProcessManagement holds configuration for managing the MongoSQL process.
type ProcessManagement struct {
	Service ProcessManagementService
}

// ProcessManagementService holds configuration for the service.
type ProcessManagementService struct {
	Name        string
	DisplayName string
	Description string
}

// MongoDBNetAuth holds configuration for authenticating with MongoDB.
type MongoDBNetAuth struct {
	Username          string `config:"username"`
	Password          string `config:"password,protected"`
	Source            string
	Mechanism         string `config:"mechanism"`
	GSSAPIServiceName string `config:"gssapiServiceName"`
}

// SetParameter holds miscellaneous configuration options.
type SetParameter struct {
	EnableTableAlterations bool
}

// Debug holds options that are useful when debugging mongosqld.
type Debug struct {
	EnableProfiling string
	ProfileScope    string
}
