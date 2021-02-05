package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"
	"runtime"
	"strconv"

	"github.com/10gen/sqlproxy/internal/procutil"
	"github.com/10gen/sqlproxy/internal/strutil"
	"github.com/10gen/sqlproxy/log"
)

// TLS protocol version constants.
const (
	TLSv1_0 = "TLS1_0"
	TLSv1_1 = "TLS1_1"
	TLSv1_2 = "TLS1_2"
)

var (
	isDarwin                = runtime.GOOS == "darwin"
	isWindows               = runtime.GOOS == "windows"
	supportedAuthMechanisms = []string{
		"SCRAM-SHA-1",
		"SCRAM-SHA-256",
		"PLAIN",
		"GSSAPI",
	}
	supportedTLSVersions = []string{
		TLSv1_0,
		TLSv1_1,
		TLSv1_2,
	}
)

// These are the constants for default values of variables.
const (
	DefaultRefreshIntervalSecs       = 0
	DefaultSampleSize                = 1000
	DefaultMaxNumColumnsPerTable     = 2000
	DefaultMaxNumFieldsPerCollection = 2000
	DefaultMaxNumTablesPerCollection = 200
	DefaultMaxNumGlobalTables        = 4000
	DefaultMaxNestedTableDepth       = 10
	DefaultMaxAllowedPacket          = 1073741824
)

// These are constants for the allowed values of NumConnectionsPerSession.
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

		enabledExpansions := cfg.ConfigExpand

		// we'll start over with a new default set and then re-parse
		// the args because they should override anything specified in
		// the yaml file.
		cfg = Default()
		err = ParseYaml(cfg, bytes.NewReader(yaml), enabledExpansions)
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

	cfg.ProcessManagement.Service.Name = "mongosql"
	cfg.ProcessManagement.Service.DisplayName = "MongoSQL Service"
	cfg.ProcessManagement.Service.Description = "MongoSQL accesses MongoDB data with SQL"

	cfg.Schema.RefreshIntervalSecs = DefaultRefreshIntervalSecs

	cfg.Schema.Sample.Size = DefaultSampleSize
	cfg.Schema.Sample.MaxNumColumnsPerTable = DefaultMaxNumColumnsPerTable
	cfg.Schema.Sample.MaxNumFieldsPerCollection = DefaultMaxNumFieldsPerCollection
	cfg.Schema.Sample.MaxNumTablesPerCollection = DefaultMaxNumTablesPerCollection
	cfg.Schema.Sample.MaxNumGlobalTables = DefaultMaxNumGlobalTables
	cfg.Schema.Sample.MaxNestedTableDepth = DefaultMaxNestedTableDepth
	cfg.Schema.Sample.Namespaces = []string{"*.*"}
	cfg.Schema.Sample.OptimizeViewSampling = true
	cfg.Schema.Sample.RefreshIntervalSecsDeprecated = DefaultRefreshIntervalSecs
	cfg.Schema.Sample.UUIDSubtype3Encoding = "old"
	cfg.Schema.Sample.SchemaMappingMode = LatticeMappingMode

	cfg.Schema.Stored.Mode = ""
	cfg.Schema.Stored.Source = ""
	cfg.Schema.Stored.Name = "defaultSchema"

	cfg.SystemLog.LogRotate = log.Rename

	cfg.Debug.ProfileScope = "queries"
	cfg.Debug.UsageLogInterval = 60

	cfg.SetParameter.AnonymizeMetrics = true
	cfg.ConfigExpand = EnabledExpansions{
		Exec: false,
		Rest: false,
	}
	cfg.SetParameter.EnableTableAlterations = false
	cfg.SetParameter.MetricsBackend = "off"
	cfg.SetParameter.OptimizeCrossJoins = true
	cfg.SetParameter.OptimizeEvaluations = true
	cfg.SetParameter.OptimizeFiltering = true
	cfg.SetParameter.OptimizeInnerJoins = true
	cfg.SetParameter.OptimizeSelfJoins = true
	cfg.SetParameter.Pushdown = true
	cfg.SetParameter.OptimizeViewSampling = true
	cfg.SetParameter.PolymorphicTypeConversionMode = "off"
	cfg.SetParameter.TypeConversionMode = "mongosql"
	cfg.SetParameter.ReconcileArithmeticAggFunctions = true

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
			cfg.Net.SSL.MinimumTLSVersion != "" ||
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
		cfg.MongoDB.Net.SSL.AllowInvalidCertificates ||
		cfg.MongoDB.Net.SSL.MinimumTLSVersion != "" ||
		cfg.MongoDB.Net.SSL.AllowInvalidHostnames ||
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
		!strutil.SliceContains(supportedAuthMechanisms,
			cfg.MongoDB.Net.Auth.Mechanism) {
		return fmt.Errorf("unsupported sample authentication "+
			"mechanism '%v'", cfg.MongoDB.Net.Auth.Mechanism)
	}

	if cfg.Net.SSL.MinimumTLSVersion != "" &&
		!strutil.SliceContains(supportedTLSVersions, cfg.Net.SSL.MinimumTLSVersion) {
		return fmt.Errorf("unsupported client minimum TLS version '%v'",
			cfg.Net.SSL.MinimumTLSVersion)
	}

	if cfg.MongoDB.Net.SSL.MinimumTLSVersion != "" &&
		!strutil.SliceContains(supportedTLSVersions, cfg.MongoDB.Net.SSL.MinimumTLSVersion) {
		return fmt.Errorf("unsupported mongo minimum TLS version '%v'",
			cfg.MongoDB.Net.SSL.MinimumTLSVersion)
	}

	if cfg.MongoDB.Net.NumConnectionsPerSession < MinConnections ||
		cfg.MongoDB.Net.NumConnectionsPerSession > MaxConnections {
		return fmt.Errorf("invalid number of MongoDB connections: %d "+
			"(must be between %d and %d)", cfg.MongoDB.Net.NumConnectionsPerSession,
			MinConnections, MaxConnections)
	}

	if cfg.Schema.Path != "" && cfg.Schema.Stored.Source != "" {
		return fmt.Errorf("cannot specify both a schema file path and a stored schema source")
	}

	if cfg.Schema.Sample.Size < 0 {
		return fmt.Errorf("invalid sample size: %d", cfg.Schema.Sample.Size)
	}

	if cfg.Schema.Sample.MaxNumColumnsPerTable <= 0 {
		return fmt.Errorf("invalid sample max number of columns per table: %d", cfg.Schema.Sample.MaxNumColumnsPerTable)
	}

	if cfg.Schema.Sample.MaxNumFieldsPerCollection <= 0 {
		return fmt.Errorf("invalid sample max number of fields per collection: %d", cfg.Schema.Sample.MaxNumFieldsPerCollection)
	}

	if cfg.Schema.Sample.MaxNumTablesPerCollection <= 0 {
		return fmt.Errorf("invalid sample max number of global tables: %d", cfg.Schema.Sample.MaxNumTablesPerCollection)
	}

	if cfg.Schema.Sample.MaxNumGlobalTables <= 0 {
		return fmt.Errorf("invalid sample max number of global tables: %d", cfg.Schema.Sample.MaxNumGlobalTables)
	}

	if cfg.Schema.Sample.MaxNestedTableDepth < 0 {
		return fmt.Errorf("invalid sample max nested table depth: %d", cfg.Schema.Sample.MaxNestedTableDepth)
	}

	if _, err := strutil.NewMatcher(cfg.Schema.Sample.Namespaces); err != nil {
		return fmt.Errorf("invalid specification: %v", err)
	}

	switch cfg.Schema.Stored.Mode {
	case AutoStoredSchemaMode, CustomStoredSchemaMode, NoStoredSchemaMode:
		// known values
	default:
		return fmt.Errorf("invalid schema mode: %v", cfg.Schema.Stored.Mode)
	}

	if !(cfg.Schema.Sample.SchemaMappingMode == LatticeMappingMode ||
		cfg.Schema.Sample.SchemaMappingMode == MajorityMappingMode) {
		return fmt.Errorf("invalid schema mapping mode: %v",
			cfg.Schema.Sample.SchemaMappingMode)
	}

	if cfg.Schema.Stored.Source != "" {
		if err := procutil.ValidateDBName(cfg.Schema.Stored.Source); err != nil {
			return fmt.Errorf("invalid schema source: %v", err)
		}
	}

	if cfg.Schema.Stored.Source == "" && cfg.Schema.Stored.Mode != NoStoredSchemaMode {
		return fmt.Errorf("stored schema modes require a non-empty schema source")
	}

	if cfg.Schema.WriteMode {
		err := checkForDissallowedWriteModeSettings(cfg)
		if err != nil {
			return err
		}
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

	if cfg.Debug.UsageLogInterval < 0 {
		return fmt.Errorf(
			"debug.usageLogInterval (%d) must be greater than or equal to 0",
			cfg.Debug.UsageLogInterval,
		)
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

	if cfg.SetParameter.MetricsBackend == "stitch" && cfg.Metrics.StitchURL == "" {
		return fmt.Errorf("must provide metrics.stitchURL when default metrics_backend is 'stitch'")
	}

	return nil
}

// This checks for all the possible settings that we cannot allow in --writeMode.
func checkForDissallowedWriteModeSettings(cfg *Config) error {
	defaultCfg := Default()
	if cfg.Schema.Path != defaultCfg.Schema.Path {
		return fmt.Errorf("write mode schema cannot have a (drdl) schema path")
	}
	if cfg.Schema.RefreshIntervalSecs != defaultCfg.Schema.RefreshIntervalSecs {
		return fmt.Errorf("write mode schema cannot have refreshIntervalSecs")
	}
	if cfg.Schema.Sample.MaxNestedTableDepth != defaultCfg.Schema.Sample.MaxNestedTableDepth {
		return fmt.Errorf("write mode schema cannot have sample settings, found maxNestedTableDepth")
	}
	if cfg.Schema.Sample.MaxNumColumnsPerTable != defaultCfg.Schema.Sample.MaxNumColumnsPerTable {
		return fmt.Errorf("write mode schema cannot have sample settings, found maxNumColumnsPerTable")
	}

	if cfg.Schema.Sample.MaxNumFieldsPerCollection != defaultCfg.Schema.Sample.MaxNumFieldsPerCollection {
		return fmt.Errorf("write mode schema cannot have sample settings, found maxNumFieldsPerCollection")
	}

	if cfg.Schema.Sample.MaxNumTablesPerCollection != defaultCfg.Schema.Sample.MaxNumTablesPerCollection {
		return fmt.Errorf("write mode schema cannot have sample settings, found maxNumTablesPerCollection")
	}

	if cfg.Schema.Sample.MaxNumGlobalTables != defaultCfg.Schema.Sample.MaxNumGlobalTables {
		return fmt.Errorf("write mode schema cannot have sample settings, found maxNumGlobalTables")
	}

	for i := range cfg.Schema.Sample.Namespaces {
		if cfg.Schema.Sample.Namespaces[i] != defaultCfg.Schema.Sample.Namespaces[i] {
			return fmt.Errorf("write mode schema cannot have sample settings, found namespaces")
		}
	}
	if cfg.Schema.Sample.OptimizeViewSampling != defaultCfg.Schema.Sample.OptimizeViewSampling {
		return fmt.Errorf("write mode schema cannot have sample settings, found optimizeViewSampling")
	}
	if cfg.Schema.Sample.PreJoin != defaultCfg.Schema.Sample.PreJoin {
		return fmt.Errorf("write mode schema cannot have sample settings, found prejoin")
	}
	if cfg.Schema.Sample.RefreshIntervalSecsDeprecated != defaultCfg.Schema.Sample.RefreshIntervalSecsDeprecated {
		return fmt.Errorf("write mode schema cannot have sample settings, found refreshIntervalSecsDeprecated")
	}
	if cfg.Schema.Sample.SchemaMappingMode != defaultCfg.Schema.Sample.SchemaMappingMode {
		return fmt.Errorf("write mode schema cannot have sample settings, found schemaMappingMode")
	}
	if cfg.Schema.Sample.Size != defaultCfg.Schema.Sample.Size {
		return fmt.Errorf("write mode schema cannot have sample settings, found size")
	}
	if cfg.Schema.Sample.UUIDSubtype3Encoding != defaultCfg.Schema.Sample.UUIDSubtype3Encoding {
		return fmt.Errorf("write mode schema cannot have sample settings, found uuidSubtype3Encoding")
	}
	if cfg.Schema.Stored.Mode != defaultCfg.Schema.Stored.Mode {
		return fmt.Errorf("write mode schema cannot be used with a stored schema mode")
	}
	if cfg.Schema.Stored.Source != defaultCfg.Schema.Stored.Source {
		return fmt.Errorf("write mode schema cannot have a stored schema source")
	}
	if cfg.Schema.Stored.Name != defaultCfg.Schema.Stored.Name {
		return fmt.Errorf("write mode schema cannot have a stored schema name")
	}
	return nil
}

// Config is the root of the configuration tree for mongosqld.
type Config struct {
	// Config is the file to load extra configuration from.
	Config       string
	ConfigExpand EnabledExpansions

	SystemLog         SystemLog
	Schema            Schema
	Runtime           Runtime
	Net               Net
	Security          Security
	MongoDB           MongoDB `config:"mongodb"`
	Metrics           Metrics
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
	Verbosity int64
}

// Level returns the verbosity level of the logging configuration.
func (cfg SystemLog) Level() int64 {
	if cfg.Quiet {
		return -1
	}
	return cfg.Verbosity
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
	Path                string
	WriteMode           bool
	MaxVarcharLength    uint64
	RefreshIntervalSecs int64
	Sample              SchemaSampleOptions  `config:"sample"`
	Stored              SchemaStorageOptions `config:"stored"`
}

// StoredSchemaMode is an enum representing mongosqld's stored-schema-management modes.
type StoredSchemaMode string

// Values for StoredSchemaMode.
const (
	NoStoredSchemaMode     = ""
	AutoStoredSchemaMode   = "auto"
	CustomStoredSchemaMode = "custom"
)

// MappingMode is a name for the sampling mode to use.
type MappingMode = string

// Values for MappingMode.
const (
	// LatticeMappingMode uses type lattice for resolving scalar type conflicts.
	LatticeMappingMode = "lattice"
	// MajorityMappingMode uses the scalar type with the most samples for resolving
	// scalar conflicts.
	MajorityMappingMode = "majority"
)

// SchemaStorageOptions holds stored-schema-management configuration.
type SchemaStorageOptions struct {
	Mode   StoredSchemaMode `config:"mode"`
	Source string           `config:"source"`
	Name   string           `config:"name"`
}

// SchemaSampleOptions holds schema sampling configuration.
type SchemaSampleOptions struct {
	MaxNestedTableDepth           int64       `config:"-"`
	MaxNumColumnsPerTable         int64       `config:"-"`
	MaxNumFieldsPerCollection     int64       `config:"-"`
	MaxNumTablesPerCollection     int64       `config:"-"`
	MaxNumGlobalTables            int64       `config:"-"`
	Namespaces                    []string    `config:"namespaces"`
	OptimizeViewSampling          bool        `config:"optimizeViewSampling"`
	PreJoin                       bool        `config:"prejoin"`
	RefreshIntervalSecsDeprecated int64       `config:"refreshIntervalSecs"`
	SchemaMappingMode             MappingMode `config:"schemaMappingMode"`
	Size                          int64       `config:"size"`
	UUIDSubtype3Encoding          string      `config:"uuidSubtype3Encoding"`
}

// NewSchemaSampleOptions creates a new schema sampling configuration with the given options.
func NewSchemaSampleOptions(maxNestedTableDepth int64,
	maxNumFieldsPerCollection, maxNumColumnsPerTable, maxNumTablesPerCollection, maxNumGlobalTables int64,
	namespaces []string,
	optimizeViewSampling bool,
	preJoin bool,
	schemaMappingMode MappingMode,
	size int64,
	uuidSubtype3Encoding string) SchemaSampleOptions {
	return SchemaSampleOptions{
		MaxNestedTableDepth:           maxNestedTableDepth,
		MaxNumColumnsPerTable:         maxNumColumnsPerTable,
		MaxNumFieldsPerCollection:     maxNumFieldsPerCollection,
		MaxNumTablesPerCollection:     maxNumTablesPerCollection,
		MaxNumGlobalTables:            maxNumGlobalTables,
		Namespaces:                    namespaces,
		OptimizeViewSampling:          optimizeViewSampling,
		PreJoin:                       preJoin,
		RefreshIntervalSecsDeprecated: DefaultRefreshIntervalSecs,
		SchemaMappingMode:             schemaMappingMode,
		Size:                          size,
		UUIDSubtype3Encoding:          uuidSubtype3Encoding,
	}
}

// GetMappingMode creates a MappingMode from a string.
func GetMappingMode(mode string) MappingMode {
	switch mode {
	case "lattice":
		return LatticeMappingMode
	case "majority":
		return MajorityMappingMode
	}
	panic("Mapping mode must be 'lattice' or 'majority'")
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
	ConstrainedDelegation bool `config:"constrainedDelegation"`
	Hostname              string
	ServiceName           string `config:"serviceName"`
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

// Metrics holds configuration for metrics collection.
type Metrics struct {
	StitchURL string `config:"stitchURL,protected"`
}

// EnabledExpansions holds a 2-bit state enabling expansion directives.
type EnabledExpansions struct {
	Exec bool
	Rest bool
}

// SetParameter holds miscellaneous configuration options.
type SetParameter struct {
	AnonymizeMetrics                bool `config:"anonymize_metrics"`
	EnableTableAlterations          bool
	MetricsBackend                  string `config:"metrics_backend"`
	OptimizeCrossJoins              bool   `config:"optimize_cross_joins"`
	OptimizeEvaluations             bool   `config:"optimize_evaluations"`
	OptimizeFiltering               bool   `config:"optimize_filtering"`
	OptimizeInnerJoins              bool   `config:"optimize_inner_joins"`
	OptimizeSelfJoins               bool   `config:"optimize_self_joins"`
	OptimizeViewSampling            bool   `config:"optimize_view_sampling"`
	PolymorphicTypeConversionMode   string `config:"polymorphic_type_conversion_mode"`
	Pushdown                        bool   `config:"pushdown"`
	RewriteDistinctAsGroup          bool   `config:"rewrite_distinct_as_group"`
	TypeConversionMode              string `config:"type_conversion_mode"`
	ReconcileArithmeticAggFunctions bool   `config:"reconcile_arithmetic_agg_functions"`
}

// Debug holds options that are useful when debugging mongosqld.
type Debug struct {
	EnableProfiling  string
	ProfileScope     string
	UsageLogInterval int
}
