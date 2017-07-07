package config

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/jessevdk/go-flags"
)

const usage = `mongosqld [install|uninstall] <options>`

// ParseArgs parses the arguments and overrides values in the cfg.
func ParseArgs(cfg *Config, args []string) error {

	parser := flags.NewNamedParser(usage, flags.None)

	opts := options{
		clientConnectionOptions: &clientConnectionOptions{},
		generalOptions:          &generalOptions{},
		logOptions:              &logOptions{},
		mongoConnectionOptions:  &mongoConnectionOptions{},
		schemaOptions:           &schemaOptions{},
		socketOptions:           &socketOptions{},
		serviceOptions:          &serviceOptions{},
	}

	groups := []optionGroup{
		opts.clientConnectionOptions,
		opts.generalOptions,
		opts.logOptions,
		opts.mongoConnectionOptions,
		opts.schemaOptions,
		opts.serviceOptions,
	}

	if !isWindows {
		groups = append(groups, opts.socketOptions)
	}

	for _, group := range groups {
		header := fmt.Sprintf("%s options", group.name())
		if _, err := parser.AddGroup(header, "", group); err != nil {
			return err
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

	if _, err := parser.ParseArgs(args); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing command line options: %v\n", err)
		fmt.Fprintln(os.Stderr, "try 'mongosqld --help' for more information")
		os.Exit(2)
	}

	for _, group := range groups {
		err := group.mapToConfig(cfg)
		if err != nil {
			return err
		}
	}

	if opts.Version != nil && *opts.Version {
		PrintVersionAndGitspec("mongosqld", os.Stdout)
		os.Exit(0)
	}

	if opts.generalOptions != nil && opts.generalOptions.Help != nil && *opts.generalOptions.Help {
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	return nil
}

type options struct {
	*clientConnectionOptions
	*generalOptions
	*logOptions
	*mongoConnectionOptions
	*schemaOptions
	*socketOptions
	*serviceOptions
}

type clientConnectionOptions struct {
	Auth                 *bool   `long:"auth" description:"use authentication/authorization ('sslPEMKeyFile' is required when using auth)"`
	DefaultAuthMechanism *string `long:"defaultAuthMechanism" description:"the default authentication mechanism (default is SCRAM-SHA-1)"`
	DefaultAuthSource    *string `long:"defaultAuthSource" description:"the default authentication source (default is admin)"`
	Addr                 *string `long:"addr" description:"host address to listen on"`
	SSLMode              *string `long:"sslMode" description:"set the SSL operation mode" choice:"disabled" choice:"allowSSL" choice:"requireSSL"`
	SSLAllowInvalidCerts *bool   `long:"sslAllowInvalidCertificates" description:"don't require the certificate presented by the client to be valid"`
	SSLCAFile            *string `long:"sslCAFile" description:"path to a CA certificate file to use for authenticating client certificate"`
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
	if !isEmptyOrUnset(o.Addr) {
		addr := *o.Addr
		host, portS, err := net.SplitHostPort(addr)
		if err != nil {
			if !strings.Contains(err.Error(), "missing port in address") {
				return err
			}

			host = addr
			portS = "3307"
		}
		port, err := strconv.Atoi(portS)
		if err != nil {
			return err
		}

		cfg.Net.BindIP = []string{host}
		cfg.Net.Port = port
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
	return nil
}

type socketOptions struct {
	FilePermissions  *string `long:"filePermissions" description:"permissions to set on UNIX domain socket file (default to 0700)"`
	NoUnixSocket     *bool   `long:"noUnixSocket" description:"disable listening on UNIX domain sockets"`
	UnixSocketPrefix *string `long:"unixSocketPrefix" description:"alternative directory for UNIX domain sockets (default to /tmp)"`
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
	Fork    *bool   `long:"fork" description:"fork mongosqld process" hidden:"true"`
	Help    *bool   `short:"h" long:"help" description:"print usage"`
	Version *bool   `long:"version" description:"display version information"`
	Config  *string `long:"config" description:"path to a configuration file"`
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

	return nil
}

type logOptions struct {
	LogAppend    *bool              `long:"logAppend" description:"append new logging output to existing log file"`
	LogPath      *string            `long:"logPath" description:"path to a log file for storing logging output"`
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
	if o.Quiet != nil {
		cfg.SystemLog.Quiet = *o.Quiet
	}
	if o.VLevel != nil {
		cfg.SystemLog.Verbosity = *o.VLevel
	}
	return nil
}

type mongoConnectionOptions struct {
	MongoSSL                  *bool   `long:"mongo-ssl" description:"use SSL when connecting to mongo instance"`
	MongoURI                  *string `long:"mongo-uri" description:"a mongo URI (https://docs.mongodb.org/manual/reference/connection-string/) to connect to"`
	MongoUsername             *string `short:"u" value-name:"<username>" long:"mongo-username" description:"authentication username to use for schema discovery (only required if --auth is also enabled)"`
	MongoPassword             *string `short:"p" value-name:"<password>" long:"mongo-password" description:"authentication password to use for schema discovery (only required if --auth is also enabled)"`
	MongoSource               *string `long:"mongo-authenticationSource" value-name:"<authentication source>" description:"database that holds the credentials for the schema discovery user (only required if --auth is also enabled)"`
	MongoMechanism            *string `long:"mongo-authenticationMechanism" description:"authentication mechanism to use for schema discovery (only required if --auth is also enabled)" choice:"SCRAM-SHA-1" choice:"PLAIN"`
	MongoAllowInvalidCerts    *bool   `long:"mongo-sslAllowInvalidCertificates" description:"don't require the certificate presented by the MongoDB server to be valid, when using --mongo-ssl"`
	MongoSSLAllowInvalidHost  *bool   `long:"mongo-sslAllowInvalidHostnames" description:"bypass the validation for server name"`
	MongoCAFile               *string `long:"mongo-sslCAFile" value-name:"<filename>" description:"path to a CA certificate file to use for authenticating certificates from MongoDB, when using --mongo-ssl"`
	MongoSSLCRLFile           *string `long:"mongo-sslCRLFile" value-name:"<filename>" description:"the .pem file containing the certificate revocation list"`
	MongoSSLFipsMode          *bool   `long:"mongo-sslFIPSMode" description:"use FIPS mode of the installed openssl library"`
	MongoPEMKeyFile           *string `long:"mongo-sslPEMKeyFile" value-name:"<filename>" description:"path to a file containing the certificate and private key for connecting to MongoDB, when using --mongo-ssl"`
	MongoPEMKeyPassword       *string `long:"mongo-sslPEMKeyPassword" description:"password to decrypt private key in mongo-sslPEMKeyFile"`
	MongoVersionCompatibility *string `long:"mongo-versionCompatibility" description:"indicates the mongodb version with which to be compatible (only necessary when used with mixed version replica sets)."`
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

	return nil
}

type schemaOptions struct {
	Schema           *string `long:"schema" description:"the path to a schema file"`
	SchemaDir        *string `long:"schemaDirectory" description:"the path to a directory containing schema files to load"`
	MaxVarcharLength *uint16 `long:"maxVarcharLength" description:"the maximum length of a varchar"`
	// Databases will append the database name each time the option is encountered
	// (can be set multiple times, like --sampleDatabase foo --sampleDatabase bar)
	Databases            []string `long:"sampleDatabase" value-name:"<sample database>" description:"database(s) to sample in generating schema (defaults to all databases - except admin and local)"`
	SampleSize           *int64   `long:"sampleSize" short:"s" description:"the number of documents to sample, per database, when generating the schema (default: 1000)"`
	UUIDSubtype3Encoding *string  `long:"uuidSubtype3Encoding" short:"b" description:"encoding used to generate UUID binary subtype 3. old: Old BSON binary subtype representation; csharp: The C#/.NET legacy UUID representation; java: The Java legacy UUID representation" choice:"old" choice:"csharp" choice:"java"`
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

	if o.Databases != nil {
		cfg.Schema.Sample.Databases = o.Databases
	}

	if o.SampleSize != nil {
		cfg.Schema.Sample.SampleSize = *o.SampleSize
	}

	if !isEmptyOrUnset(o.UUIDSubtype3Encoding) {
		cfg.Schema.Sample.UUIDSubtype3Encoding = *o.UUIDSubtype3Encoding
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
