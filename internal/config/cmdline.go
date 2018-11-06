package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"

	"path/filepath"

	"github.com/10gen/sqlproxy/log"
	"github.com/jessevdk/go-flags"
)

const usage = `mongosqld [install|uninstall] <options>`

// ErrExitEarly is used to check if mongosqld should exit early with a normal
// exit code (for --help, --version, etc)
var ErrExitEarly = errors.New("exit early")

// MakeAbsPaths takes args and a map representing a set of command line flags that
// expect a path that should be translated to absolute.
//
// pre-condition: can only convert paths if the arg is of the form --flag=arg, positional
// arguments must be converted elsewhere or they will be missed!
func MakeAbsPaths(args []string, flags map[string]struct{}) ([]string, error) {
	ret := make([]string, len(args), cap(args))
	copy(ret, args)
	var err error
	for i, arg := range args {
		argSplit := strings.Split(arg, "=")
		if len(argSplit) != 2 {
			continue
		}
		if _, contains := flags[argSplit[0]]; contains {
			argSplit[1], err = filepath.Abs(argSplit[1])
			if err != nil {
				return nil, err
			}
			ret[i] = strings.Join(argSplit, "=")
		}
	}
	return ret, nil
}

// CapturePositionalArgs converts positional arguments like --flag arg to
// --flag=arg.  This must be done individually for each flag we wish to
// capture, as some, like --verbose, have special requirements of their
// positional arguments
func CapturePositionalArgs(args []string) ([]string, error) {
	ret := make([]string, len(args), cap(args))
	copy(ret, args)
	// capture positional args for -v, --verbose, and --config
	// as well as for --config
	for i, arg := range ret {
		switch arg {
		case "-v":
			fallthrough
		case "--verbose":
			if i+1 < len(ret) {
				_, err := strconv.Atoi(ret[i+1])
				if err == nil {
					ret[i] = arg + "=" + ret[i+1]
					ret = append(ret[:i+1], ret[i+2:]...)
				}
			}
		case "--config":
			if i+1 >= len(ret) {
				return nil, errors.New("--config flag requires a path argument")
			}
			ret[i] = arg + "=" + ret[i+1]
			ret = append(ret[:i+1], ret[i+2:]...)
		}

	}
	return ret, nil
}

// ParseArgs parses the arguments and overrides values in the cfg.
// it returns the modified args for use elsewhere (such as passing to a Service)
func ParseArgs(cfg *Config, args []string) ([]string, error) {
	// Note: ParseArgs is called multiple times, so it is not safe to modify the
	// passed args slice in place
	parser := flags.NewNamedParser(usage, flags.None)

	opts := options{
		clientConnectionOptions: &clientConnectionOptions{},
		debugOptions:            &debugOptions{},
		generalOptions:          &generalOptions{},
		logOptions:              &logOptions{},
		mongoConnectionOptions:  &mongoConnectionOptions{},
		metricsOptions:          &metricsOptions{},
		schemaOptions:           &schemaOptions{},
		serviceOptions:          &serviceOptions{},
		socketOptions:           &socketOptions{},
	}

	groups := []optionGroup{
		opts.clientConnectionOptions,
		opts.debugOptions,
		opts.generalOptions,
		opts.logOptions,
		opts.mongoConnectionOptions,
		opts.metricsOptions,
		opts.schemaOptions,
		opts.serviceOptions,
	}

	if !isWindows {
		groups = append(groups, opts.socketOptions)
	}

	for _, group := range groups {
		header := fmt.Sprintf("%s options", group.name())
		if _, err := parser.AddGroup(header, "", group); err != nil {
			return nil, err
		}
	}

	// called when -v or --verbose is parsed
	opts.SetVerbosity = func(val string) error {
		opts.VLevel = new(int)
		if i, err := strconv.Atoi(val); err == nil {
			*opts.VLevel = *opts.VLevel + i // -v=N or --verbose=N
		} else if matched, _ := regexp.MatchString(`^v+$`, val); matched {
			*opts.VLevel = *opts.VLevel + len(val) + 1 // handles the -vvv cases
		} else if matched, _ := regexp.MatchString(`^v+=[0-9]$`, val); matched {
			*opts.VLevel = parseVal(val) // i.e. -vv=3
		} else if val == "" {
			*opts.VLevel = *opts.VLevel + 1 // increment for every occurrence of flag
		} else {
			return fmt.Errorf("invalid verbosity value given")
		}
		return nil
	}

	// called when --setParameter is parsed
	opts.SetParameter = func(val string) error {
		if opts.Params == nil {
			opts.Params = make(map[string]string)
		}
		split := strings.Split(val, "=")
		if len(split) != 2 {
			return fmt.Errorf("invalid setParameter expression: %s", val)
		}
		opts.Params[split[0]] = split[1]
		return nil
	}

	args, err := CapturePositionalArgs(args)
	if err != nil {
		return nil, err
	}

	args, err = MakeAbsPaths(args, map[string]struct{}{"--config": {}})
	if err != nil {
		return nil, err
	}

	if retargs, err := parser.ParseArgs(args); err != nil {
		// fix error message when go-flags infers verbosity type incorrectly
		containsFuncType := strings.Contains(err.Error(), "(expected func(string) error)")
		verbosityFlag := strings.Contains(err.Error(), "verbose")
		setParameterFlag := strings.Contains(err.Error(), "setParameter")

		var expectedType string
		if verbosityFlag {
			expectedType = "int"
		} else if setParameterFlag {
			expectedType = "<param>=<value>"
		}

		if containsFuncType {
			oldStr := "(expected func(string) error)"
			newStr := fmt.Sprintf("(expected %s)", expectedType)
			newErrStr := strings.Replace(err.Error(), oldStr, newStr, 1)
			err = errors.New(newErrStr)
		}

		return nil, fmt.Errorf("error parsing command line options: %v", err)
	} else if len(retargs) > 0 {
		return nil, fmt.Errorf(
			"error parsing command line options: Unexpected argument(s): %v",
			retargs,
		)
	}

	for _, group := range groups {
		err := group.mapToConfig(cfg)
		if err != nil {
			return nil, err
		}
	}

	// Set defaults for MinimumTLSVersion if neither command line flags nor configuration file
	// options are set.
	clientMinimumTLSVersionUnset := cfg.MongoDB.Net.SSL.MinimumTLSVersion == "" &&
		isEmptyOrUnset(opts.mongoConnectionOptions.MongoMinimumTLSVersion)

	if cfg.MongoDB.Net.SSL.Enabled && clientMinimumTLSVersionUnset {
		cfg.MongoDB.Net.SSL.MinimumTLSVersion = TLSv1_1
	}

	mongoMinimumTLSVersionUnset := cfg.Net.SSL.MinimumTLSVersion == "" &&
		isEmptyOrUnset(opts.clientConnectionOptions.MinimumTLSVersion)

	if cfg.Net.SSL.Mode != "disabled" && mongoMinimumTLSVersionUnset {
		cfg.Net.SSL.MinimumTLSVersion = TLSv1_1
	}

	if opts.Version != nil && *opts.Version {
		PrintVersionAndGitspec("mongosqld", os.Stdout)
		return nil, ErrExitEarly
	}

	if opts.generalOptions != nil && opts.generalOptions.Help != nil && *opts.generalOptions.Help {
		parser.WriteHelp(os.Stdout)
		return nil, ErrExitEarly
	}

	return args, nil
}

type options struct {
	*clientConnectionOptions
	*generalOptions
	*logOptions
	*mongoConnectionOptions
	*metricsOptions
	*schemaOptions
	*socketOptions
	*serviceOptions
	*debugOptions
}

// nolint: lll
type clientConnectionOptions struct {
	Addr                 *string `long:"addr" description:"comma separated list of ip addresses to listen on ('localhost' by default)"`
	Auth                 *bool   `long:"auth" description:"use authentication/authorization ('sslPEMKeyFile' is required when using auth)"`
	DefaultAuthMechanism *string `long:"defaultAuthMechanism" description:"the default authentication mechanism ('SCRAM-SHA-1' by default)"`
	DefaultAuthSource    *string `long:"defaultAuthSource" description:"the default authentication source ('admin' by default)"`
	GSSAPIHostname       *string `long:"gssapiHostname" description:"the hostname to use when hosting using GSSAPI/Kerberos (server's first bind ip address by default)"`
	GSSAPIServiceName    *string `long:"gssapiServiceName" description:"the service name to use when hosting using GSSAPI/Kerberos ('mongosql' by default)"`
	MinimumTLSVersion    *string `long:"minimumTLSVersion" description:"the minimum TLS protocol version accepted by the BI Connector from the client" choice:"TLS1_0" choice:"TLS1_1" choice:"TLS1_2"`
	SSLAllowInvalidCerts *bool   `long:"sslAllowInvalidCertificates" description:"don't require the certificate presented by the client to be valid"`
	SSLCAFile            *string `long:"sslCAFile" description:"path to a CA certificate file to use for authenticating client certificate"`
	SSLMode              *string `long:"sslMode" description:"set the SSL operation mode" choice:"disabled" choice:"allowSSL" choice:"requireSSL"`
	SSLPEMKeyFile        *string `long:"sslPEMKeyFile" description:"path to a file containing the certificate and private key establishing a connection with a client"`
	SSLPEMKeyPassword    *string `long:"sslPEMKeyPassword" description:"password to decrypt private key in --sslPEMKeyFile"`
}

func (o *clientConnectionOptions) name() string {
	return "Client Connection"
}

func (o *clientConnectionOptions) mapToConfig(cfg *Config) error {
	if o.Auth != nil {
		cfg.Security.Enabled = *o.Auth
	}
	if !isEmptyOrUnset(o.DefaultAuthMechanism) {
		cfg.Security.DefaultMechanism = *o.DefaultAuthMechanism
	}
	if !isEmptyOrUnset(o.DefaultAuthSource) {
		cfg.Security.DefaultSource = *o.DefaultAuthSource
	}
	if !isEmptyOrUnset(o.GSSAPIHostname) {
		cfg.Security.GSSAPI.Hostname = *o.GSSAPIHostname
	}
	if !isEmptyOrUnset(o.GSSAPIServiceName) {
		cfg.Security.GSSAPI.ServiceName = *o.GSSAPIServiceName
	}
	if !isEmptyOrUnset(o.Addr) {
		addrs := strings.Split(*o.Addr, ",")
		prevPort := -1
		for idx, addr := range addrs {
			host, portS, err := net.SplitHostPort(addr)
			if err != nil {
				if !strings.Contains(err.Error(), "missing port in address") {
					return err
				}
				host = addr
				addrs[idx] = host
				// port unspecified do nothing
				continue
			}

			currentPort, err := strconv.Atoi(portS)
			if err != nil {
				return err
			}

			if prevPort == -1 {
				prevPort = currentPort
			} else if currentPort != prevPort {
				return fmt.Errorf("the ports are inconsistent in the provided bind ips")
			}

			addrs[idx] = host
		}

		cfg.Net.BindIP = addrs
		if prevPort == -1 {
			// no ports were provided
			cfg.Net.Port = 3307
		} else {
			cfg.Net.Port = prevPort
		}
	}

	if !isEmptyOrUnset(o.SSLMode) {
		cfg.Net.SSL.Mode = *o.SSLMode
	}
	if o.SSLAllowInvalidCerts != nil {
		cfg.Net.SSL.AllowInvalidCertificates = *o.SSLAllowInvalidCerts
	}
	if !isEmptyOrUnset(o.SSLCAFile) {
		cfg.Net.SSL.CAFile = *o.SSLCAFile
	}
	if !isEmptyOrUnset(o.SSLPEMKeyFile) {
		cfg.Net.SSL.PEMKeyFile = *o.SSLPEMKeyFile
	}
	if !isEmptyOrUnset(o.SSLPEMKeyPassword) {
		cfg.Net.SSL.PEMKeyPassword = *o.SSLPEMKeyPassword
	}
	if !isEmptyOrUnset(o.MinimumTLSVersion) {
		cfg.Net.SSL.MinimumTLSVersion = *o.MinimumTLSVersion
	}
	return nil
}

// nolint: lll
type socketOptions struct {
	FilePermissions  *string `long:"filePermissions" description:"permissions to set on UNIX domain socket file (0700 by default)"`
	NoUnixSocket     *bool   `long:"noUnixSocket" description:"disable listening on UNIX domain sockets"`
	UnixSocketPrefix *string `long:"unixSocketPrefix" description:"alternative directory for UNIX domain sockets (/tmp by default)"`
}

func (o *socketOptions) name() string {
	return "Socket"
}

func (o *socketOptions) mapToConfig(cfg *Config) error {
	if !isEmptyOrUnset(o.FilePermissions) {
		cfg.Net.UnixDomainSocket.FilePermissions = *o.FilePermissions
	}
	if o.NoUnixSocket != nil {
		cfg.Net.UnixDomainSocket.Enabled = !*o.NoUnixSocket
	}
	if !isEmptyOrUnset(o.UnixSocketPrefix) {
		cfg.Net.UnixDomainSocket.PathPrefix = *o.UnixSocketPrefix
	}

	return nil
}

type generalOptions struct {
	Fork         *bool              `long:"fork" description:"fork mongosqld process" hidden:"true"`
	Help         *bool              `short:"h" long:"help" description:"print usage"`
	Version      *bool              `long:"version" description:"display version information"`
	Config       *string            `long:"config" description:"path to a configuration file"`
	SetParameter func(string) error `long:"setParameter" hidden:"true"`
	Params       map[string]string  `no-flag:"true"`
}

func (o *generalOptions) name() string {
	return "General"
}

func (o *generalOptions) mapToConfig(cfg *Config) error {
	if o.Fork != nil {
		return fmt.Errorf("--fork is no longer supported")
	}
	if !isEmptyOrUnset(o.Config) {
		cfg.Config = *o.Config
	}

	for key, val := range o.Params {
		invalidValueErr := fmt.Errorf("invalid value for setParameter %s: %s", key, val)
		switch key {
		case "enableTableAlterations", "enable_table_alterations":
			switch val {
			case "true":
				cfg.SetParameter.EnableTableAlterations = true
			case "false":
				cfg.SetParameter.EnableTableAlterations = false
			default:
				return invalidValueErr
			}
		case "metrics_backend":
			switch val {
			case "off", "log", "stitch":
				cfg.SetParameter.MetricsBackend = val
			default:
				return invalidValueErr
			}
		case "optimize_evaluations":
			switch val {
			case "true":
				cfg.SetParameter.OptimizeEvaluations = true
			case "false":
				cfg.SetParameter.OptimizeEvaluations = false
			default:
				return invalidValueErr
			}
		case "optimize_cross_joins":
			switch val {
			case "true":
				cfg.SetParameter.OptimizeCrossJoins = true
			case "false":
				cfg.SetParameter.OptimizeCrossJoins = false
			default:
				return invalidValueErr
			}
		case "optimize_inner_joins":
			switch val {
			case "true":
				cfg.SetParameter.OptimizeInnerJoins = true
			case "false":
				cfg.SetParameter.OptimizeInnerJoins = false
			default:
				return invalidValueErr
			}
		case "optimize_filtering":
			switch val {
			case "true":
				cfg.SetParameter.OptimizeFiltering = true
			case "false":
				cfg.SetParameter.OptimizeFiltering = false
			default:
				return invalidValueErr
			}
		case "optimize_self_joins":
			switch val {
			case "true":
				cfg.SetParameter.OptimizeSelfJoins = true
			case "false":
				cfg.SetParameter.OptimizeSelfJoins = false
			default:
				return invalidValueErr
			}
		case "pushdown":
			switch val {
			case "true":
				cfg.SetParameter.Pushdown = true
			case "false":
				cfg.SetParameter.Pushdown = false
			default:
				return invalidValueErr
			}
		case "optimize_view_sampling":
			switch val {
			case "true":
				cfg.SetParameter.OptimizeViewSampling = true
			case "false":
				cfg.SetParameter.OptimizeViewSampling = false
			default:
				return invalidValueErr
			}
		case "polymorphic_type_conversion_mode":
			switch val {
			case "safe", "fast", "off":
				cfg.SetParameter.PolymorphicTypeConversionMode = val
			default:
				return invalidValueErr
			}
		case "type_conversion_mode":
			switch val {
			case "mongosql", "mysql":
				cfg.SetParameter.TypeConversionMode = val
			default:
				return invalidValueErr
			}
		default:
			return fmt.Errorf("invalid setParameter key: %s", key)
		}
	}

	return nil
}

// nolint: lll
type logOptions struct {
	LogAppend    *bool              `long:"logAppend" description:"append new logging output to existing log file"`
	LogPath      *string            `long:"logPath" description:"path to a log file for storing logging output"`
	LogRotate    *string            `long:"logRotate" description:"set the log rotation behavior" choice:"rename" choice:"reopen"`
	SetVerbosity func(string) error `short:"v" long:"verbose" value-name:"<level>" description:"more detailed log output (include multiple times for more verbosity, e.g. -vvvvv, or specify a numeric value, e.g. --verbose=N)" optional:"true" optional-value:""`
	Quiet        *bool              `long:"quiet" description:"hide all log output"`
	VLevel       *int               `no-flag:"true"`
}

func (o *logOptions) name() string {
	return "Log"
}

func (o *logOptions) mapToConfig(cfg *Config) error {
	if o.LogAppend != nil {
		cfg.SystemLog.LogAppend = *o.LogAppend
	}
	if !isEmptyOrUnset(o.LogPath) {
		cfg.SystemLog.Path = *o.LogPath
	}
	if !isEmptyOrUnset(o.LogRotate) {
		cfg.SystemLog.LogRotate = log.RotationStrategy(*o.LogRotate)
	}
	if o.Quiet != nil {
		cfg.SystemLog.Quiet = *o.Quiet
	}
	if o.VLevel != nil {
		cfg.SystemLog.Verbosity = *o.VLevel
	}
	return nil
}

// nolint: lll
type mongoConnectionOptions struct {
	MongoMechanism            *string `long:"mongo-authenticationMechanism" description:"authentication mechanism to use for schema discovery (only used if --auth is also enabled)"`
	MongoSource               *string `long:"mongo-authenticationSource" value-name:"<authentication source>" description:"database that holds the credentials for the schema discovery user (only used if --auth is also enabled)"`
	MongoGSSAPIServiceName    *string `long:"mongo-gssapiServiceName" description:"the service name MongoDB is using ('mongodb' by default)"`
	MongoMinimumTLSVersion    *string `long:"mongo-minimumTLSVersion" description:"the minimum TLS protocol version used to connect to MongoDB" choice:"TLS1_0" choice:"TLS1_1" choice:"TLS1_2"`
	MongoPassword             *string `short:"p" value-name:"<password>" long:"mongo-password" description:"password for admin username specified in --mongo-username (required if --auth is enabled)"`
	MongoSSL                  *bool   `long:"mongo-ssl" description:"use SSL when connecting to mongo instance"`
	MongoAllowInvalidCerts    *bool   `long:"mongo-sslAllowInvalidCertificates" description:"don't require the certificate presented by the MongoDB server to be valid, when using --mongo-ssl"`
	MongoSSLAllowInvalidHost  *bool   `long:"mongo-sslAllowInvalidHostnames" description:"bypass the validation for server name"`
	MongoCAFile               *string `long:"mongo-sslCAFile" value-name:"<filename>" description:"path to a CA certificate file to use for authenticating certificates from MongoDB, when using --mongo-ssl"`
	MongoSSLCRLFile           *string `long:"mongo-sslCRLFile" value-name:"<filename>" description:"the .pem file containing the certificate revocation list"`
	MongoSSLFipsMode          *bool   `long:"mongo-sslFIPSMode" description:"use FIPS mode of the installed openssl library"`
	MongoPEMKeyFile           *string `long:"mongo-sslPEMKeyFile" value-name:"<filename>" description:"path to a file containing the certificate and private key for connecting to MongoDB, when using --mongo-ssl"`
	MongoPEMKeyPassword       *string `long:"mongo-sslPEMKeyPassword" description:"password to decrypt private key in mongo-sslPEMKeyFile"`
	MongoUsername             *string `short:"u" value-name:"<username>" long:"mongo-username" description:"MongoDB username to use for admin tasks such as metadata loading and schema discovery (required if --auth is enabled)"`
	MongoURI                  *string `long:"mongo-uri" description:"a mongo URI (https://docs.mongodb.org/manual/reference/connection-string/) to connect to"`
	MongoVersionCompatibility *string `long:"mongo-versionCompatibility" description:"indicates the MongoDB version with which to be compatible (only necessary when used with mixed version replica sets)."`
}

func (o *mongoConnectionOptions) name() string {
	return "Mongo Connection"
}

func (o *mongoConnectionOptions) mapToConfig(cfg *Config) error {
	if o.MongoSSL != nil {
		cfg.MongoDB.Net.SSL.Enabled = *o.MongoSSL
	}
	if !isEmptyOrUnset(o.MongoURI) {
		cfg.MongoDB.Net.URI = *o.MongoURI
	}
	if o.MongoAllowInvalidCerts != nil {
		cfg.MongoDB.Net.SSL.AllowInvalidCertificates = *o.MongoAllowInvalidCerts
	}
	if o.MongoSSLAllowInvalidHost != nil {
		cfg.MongoDB.Net.SSL.AllowInvalidHostnames = *o.MongoSSLAllowInvalidHost
	}
	if !isEmptyOrUnset(o.MongoCAFile) {
		cfg.MongoDB.Net.SSL.CAFile = *o.MongoCAFile
	}
	if !isEmptyOrUnset(o.MongoSSLCRLFile) {
		cfg.MongoDB.Net.SSL.CRLFile = *o.MongoSSLCRLFile
	}
	if o.MongoSSLFipsMode != nil {
		cfg.MongoDB.Net.SSL.FIPSMode = *o.MongoSSLFipsMode
	}
	if !isEmptyOrUnset(o.MongoPEMKeyFile) {
		cfg.MongoDB.Net.SSL.PEMKeyFile = *o.MongoPEMKeyFile
	}
	if !isEmptyOrUnset(o.MongoPEMKeyPassword) {
		cfg.MongoDB.Net.SSL.PEMKeyPassword = *o.MongoPEMKeyPassword
	}
	if !isEmptyOrUnset(o.MongoVersionCompatibility) {
		cfg.MongoDB.VersionCompatibility = *o.MongoVersionCompatibility
	}

	if !isEmptyOrUnset(o.MongoUsername) {
		cfg.MongoDB.Net.Auth.Username = *o.MongoUsername
	}
	if !isEmptyOrUnset(o.MongoPassword) {
		cfg.MongoDB.Net.Auth.Password = *o.MongoPassword
	}
	if !isEmptyOrUnset(o.MongoSource) {
		cfg.MongoDB.Net.Auth.Source = *o.MongoSource
	}
	if !isEmptyOrUnset(o.MongoMechanism) {
		cfg.MongoDB.Net.Auth.Mechanism = *o.MongoMechanism
	}
	if !isEmptyOrUnset(o.MongoGSSAPIServiceName) {
		cfg.MongoDB.Net.Auth.GSSAPIServiceName = *o.MongoGSSAPIServiceName
	}
	if !isEmptyOrUnset(o.MongoMinimumTLSVersion) {
		cfg.MongoDB.Net.SSL.MinimumTLSVersion = *o.MongoMinimumTLSVersion
	}

	return nil
}

type metricsOptions struct {
	StitchURL *string `long:"stitch-url" description:"stitch metrics records endpoint" hidden:"true"`
}

func (*metricsOptions) name() string {
	return "Metrics"
}

func (o *metricsOptions) mapToConfig(cfg *Config) error {
	if !isEmptyOrUnset(o.StitchURL) {
		cfg.Metrics.StitchURL = *o.StitchURL
	}
	return nil
}

// nolint: lll
type schemaOptions struct {
	Schema           *string `long:"schema" description:"the path to a schema file"`
	SchemaDir        *string `long:"schemaDirectory" description:"the path to a directory containing schema files to load"`
	MaxVarcharLength *uint16 `long:"maxVarcharLength" description:"the maximum length of a varchar"`

	SampleSource *string `long:"sampleSource" description:"database to use for reading/writing sampled schema"`
	SampleMode   *string `long:"sampleMode" description:"set the mongosqld sampling operation mode ('read' by default)" choice:"read" choice:"write"`
	SampleSize   *int64  `long:"sampleSize" description:"the number of documents to sample, per database, when sampling the schema(s) (1000 by default)"`
	PreJoin      *bool   `long:"prejoin" description:"generate unwound tables including parent columns, effectively resulting in a pre-joined table"`

	// Namespaces will append the namespace every time the option is encountered
	// (can be set multiple times, like --sampleNamespaces foo.* --sampleNamespaces bar.*_dev)
	SampleNamespaces           []string `long:"sampleNamespaces" value-name:"<sample namespaces>" description:"namespace(s) to sample in generating schema (defaults to all namespaces - except admin and local databases)"`
	SampleRefreshIntervalSecs  *int64   `long:"sampleRefreshIntervalSecs" description:"the interval (in seconds) mongosqld waits before re-sampling the schema(s)"`
	SampleUUIDSubtype3Encoding *string  `long:"uuidSubtype3Encoding" short:"b" description:"encoding used to generate UUID binary subtype 3. old: Old BSON binary subtype representation; csharp: The C#/.NET legacy UUID representation; java: The Java legacy UUID representation" choice:"old" choice:"csharp" choice:"java"`
	SchemaMappingHeuristic     *string  `long:"schemaMappingHeuristic" hidden:"true" description:"schema mapping heuristic to use" choice:"lattice" choice:"majority"`
}

func (o *schemaOptions) name() string {
	return "Schema"
}

func (o *schemaOptions) mapToConfig(cfg *Config) error {
	schemaSet := false

	if !isEmptyOrUnset(o.Schema) {
		cfg.Schema.Path = *o.Schema
		schemaSet = true
	}

	if !isEmptyOrUnset(o.SchemaDir) {
		if schemaSet {
			return fmt.Errorf("must specify only one of --schema or --schemaDirectory")
		}
		cfg.Schema.Path = *o.SchemaDir
	}

	if o.MaxVarcharLength != nil {
		cfg.Schema.MaxVarcharLength = *o.MaxVarcharLength
	}

	if !isEmptyOrUnset(o.SampleMode) {
		cfg.Schema.Sample.Mode = SampleMode(*o.SampleMode)
	}

	if !isEmptyOrUnset(o.SampleSource) {
		cfg.Schema.Sample.Source = *o.SampleSource
	}

	if o.SampleSize != nil {
		cfg.Schema.Sample.Size = *o.SampleSize
	}

	if o.PreJoin != nil {
		cfg.Schema.Sample.PreJoin = *o.PreJoin
	}

	if o.SampleNamespaces != nil {
		cfg.Schema.Sample.Namespaces = o.SampleNamespaces
	}

	if o.SampleRefreshIntervalSecs != nil {
		cfg.Schema.Sample.RefreshIntervalSecs = *o.SampleRefreshIntervalSecs
	}

	if !isEmptyOrUnset(o.SampleUUIDSubtype3Encoding) {
		cfg.Schema.Sample.UUIDSubtype3Encoding = *o.SampleUUIDSubtype3Encoding
	}

	if !isEmptyOrUnset(o.SchemaMappingHeuristic) {
		cfg.Schema.Sample.SchemaMappingHeuristic = GetMappingHeuristic(*o.SchemaMappingHeuristic)
	}

	return nil
}

type serviceOptions struct {
	ServiceName        *string `long:"serviceName" description:"the service name"`
	ServiceDisplayName *string `long:"serviceDisplayName" description:"the service display name"`
	ServiceDescription *string `long:"serviceDescription" description:"the service description"`
}

func (o *serviceOptions) name() string {
	return "Service"
}

func (o *serviceOptions) mapToConfig(cfg *Config) error {
	if !isEmptyOrUnset(o.ServiceName) {
		cfg.ProcessManagement.Service.Name = *o.ServiceName
	}
	if !isEmptyOrUnset(o.ServiceDisplayName) {
		cfg.ProcessManagement.Service.DisplayName = *o.ServiceDisplayName
	}
	if !isEmptyOrUnset(o.ServiceDescription) {
		cfg.ProcessManagement.Service.Description = *o.ServiceDescription
	}

	return nil
}

// nolint: lll
type debugOptions struct {
	EnableProfiling *string `long:"enableProfiling" hidden:"true" description:"generate profiling artifacts of the specified type" choice:"cpu" choice:""`
	ProfileScope    *string `long:"profileScope" hidden:"true" description:"the scope for which profiling artifacts should be generated" choice:"queries" choice:"all"`
}

func (o *debugOptions) name() string {
	return "Debug"
}

func (o *debugOptions) mapToConfig(cfg *Config) error {
	if !isEmptyOrUnset(o.EnableProfiling) {
		cfg.Debug.EnableProfiling = *o.EnableProfiling
	}
	if !isEmptyOrUnset(o.ProfileScope) {
		cfg.Debug.ProfileScope = *o.ProfileScope
	}

	return nil
}

type optionGroup interface {
	name() string
	mapToConfig(*Config) error
}

func parseVal(val string) int {
	idx := strings.Index(val, "=")
	ret, err := strconv.Atoi(val[idx+1:])
	if err != nil {
		panic(fmt.Errorf("value was not a valid integer: %v", err))
	}
	return ret
}

func isEmptyOrUnset(s *string) bool {
	return s == nil || *s == ""
}
