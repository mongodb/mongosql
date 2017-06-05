package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"
	"runtime"
	"strconv"
)

var isDarwin = runtime.GOOS == "darwin"
var isWindows = runtime.GOOS == "windows"

// Load creates a new configuration from the specified arguments
// and potentially loads from a separate config source as specified
// on the command line.
func Load(args []string) (*Config, error) {
	cfg := Default()
	err := ParseArgs(cfg, args)
	if err != nil {
		return nil, err
	}

	if cfg.Config != "" {
		yaml, err := ioutil.ReadFile(cfg.Config)
		if err != nil {
			return nil, err
		}
		if len(yaml) == 0 {
			// we're done. An empty file shouldn't cause an error.
			return cfg, nil
		}

		// we'll start over with a new default set and then re-parse
		// the args because they should override anything specified in
		// the yaml file.
		cfg = Default()
		err = ParseYaml(cfg, bytes.NewReader(yaml))
		if err != nil {
			return nil, fmt.Errorf("unable to parse config file: %s", err)
		}
		err = ParseArgs(cfg, args)
		if err != nil {
			return nil, err
		}
	}

	return cfg, err
}

// Default returns the default configuration.
func Default() *Config {
	cfg := &Config{}

	cfg.Net.BindIP = "127.0.0.1"
	cfg.Net.Port = 3307

	if !isWindows {
		cfg.Net.UnixDomainSocket.Enabled = true
		cfg.Net.UnixDomainSocket.PathPrefix = "/tmp"
		cfg.Net.UnixDomainSocket.FilePermissions = "0700"
	}

	cfg.Security.DefaultMechanism = "SCRAM-SHA-1"
	cfg.Security.DefaultSource = "admin"

	cfg.MongoDB.Net.URI = "mongodb://localhost:27017"

	cfg.ProcessManagement.Service.Name = "mongosql"
	cfg.ProcessManagement.Service.DisplayName = "MongoSQL Service"
	cfg.ProcessManagement.Service.Description = "MongoSQL accesses MongoDB data with SQL"

	return cfg
}

// ToJSON converts the config to json, only outputting
// fields that are not the default.
func ToJSON(cfg *Config) string {
	var w bytes.Buffer
	w.WriteString("{")
	toJSON(&w, reflect.ValueOf(*Default()), reflect.ValueOf(*cfg))
	w.WriteString("}")
	// todo: format string...
	return w.String()
}

// Validate ensure that a config is valid.
func Validate(cfg *Config) error {
	if cfg.Schema.Path == "" {
		return fmt.Errorf("a schema is required, either specified by --schema, --schemaDirectory, or in a config file at 'schema.path'")
	}

	if !cfg.MongoDB.Net.SSL.Enabled && (cfg.MongoDB.Net.SSL.CAFile != "" ||
		cfg.MongoDB.Net.SSL.CRLFile != "" ||
		cfg.MongoDB.Net.SSL.PEMKeyFile != "" ||
		cfg.MongoDB.Net.SSL.PEMKeyPassword != "") {
		return fmt.Errorf("when specifying SSL options, SSL must be enabled with --mongo-ssl or in a config file at 'mongodb.net.ssl.enabled'")
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

	return nil
}

// Config is the root of the configuration tree for mongosqld.
type Config struct {
	// Config is the file to load extra configuration from.
	Config string

	SystemLog SystemLog
	Schema    Schema
	Runtime   Runtime
	Net       Net
	Security  Security
	MongoDB   MongoDB `config:"mongodb"`
	ProcessManagement ProcessManagement
}

// SystemLog holds logging configuration.
type SystemLog struct {
	LogAppend bool
	Path      string
	Quiet     bool
	Verbosity int
}

func (c *SystemLog) validate() error {
	return nil
}

// Runtime holds runtime configuration.
type Runtime struct {
	Memory RuntimeMemory
}

// RuntimeMemory holds configuration for memory.
type RuntimeMemory struct {
	MaxPerStage uint64
}

// Schema holds schema configuration.
type Schema struct {
	Path string
}

// Net holds network related configuration.
type Net struct {
	BindIP           string `config:"bindIp"`
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
	AllowInvalidCertificates bool
	PEMKeyFile               string `config:"PEMKeyFile"`
	PEMKeyPassword           string `config:"PEMKeyPassword,protected"`
	CAFile                   string `config:"CAFile"`
}

// Security holds configuration for security with a client.
type Security struct {
	Enabled          bool
	DefaultMechanism string
	DefaultSource    string
}

// MongoDB holds configuration for connecting to MongoDB.
type MongoDB struct {
	VersionCompatibility string
	Net                  MongoDBNet
}

// MongoDBNet holds confifuration for network communication with MongoDB.
type MongoDBNet struct {
	URI string        `config:"uri"`
	SSL MongoDBNetSSL `config:"ssl"`
}

// MongoDBNetSSL holds configuration for SSL with MongoDB.
type MongoDBNetSSL struct {
	Enabled                  bool
	AllowInvalidCertificates bool
	AllowInvalidHostnames    bool
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
