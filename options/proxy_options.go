package options

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/jessevdk/go-flags"
)

var usage = "mongosqld <options>"

type SqldOptions struct {
	*SqldClientConnection
	*SqldGeneral
	*SqldLog
	*SqldMongoConnection
	*SqldSchema
	parser *flags.Parser
}

type OptionGroup interface {
	Name() string
}

type SqldClientConnection struct {
	Auth                 bool   `long:"auth" description:"use authentication/authorization ('sslPEMKeyFile' is required when using auth)"`
	Addr                 string `long:"addr" description:"host address to listen on" default:"127.0.0.1:3307"`
	SSLAllowInvalidCerts bool   `long:"sslAllowInvalidCertificates" description:"don't require the certificate presented by the client to be valid"`
	SSLCAFile            string `long:"sslCAFile" description:"path to a CA certificate file to use for authenticating client certificate"`
	SSLPEMFile           string `long:"sslPEMKeyFile" description:"path to a file containing the certificate and private key establishing a connection with a client"`
}

func (_ SqldClientConnection) Name() string {
	return "Client Connection"
}

type SqldGeneral struct {
	Help    bool `short:"h" long:"help" description:"print usage"`
	Version bool `long:"version" description:"display version information"`
}

func (_ SqldGeneral) Name() string {
	return "General"
}

type SqldLog struct {
	LogAppend    bool               `long:"logAppend" description:"append new logging output to existing log file"`
	LogPath      string             `long:"logPath" description:"path to a log file for storing logging output (defaults to stderr)"`
	SetVerbosity func(string) error `short:"v" long:"verbose" value-name:"<level>" description:"more detailed log output (include multiple times for more verbosity, e.g. -vvvvv, or specify a numeric value, e.g. --verbose=N)" optional:"true" optional-value:""`
	Quiet        bool               `long:"quiet" description:"hide all log output"`
	VLevel       int                `no-flag:"true"`
}

func (_ SqldLog) Name() string {
	return "Log"
}

func (lo SqldLog) Level() int {
	return lo.VLevel
}

func (lo SqldLog) IsQuiet() bool {
	return lo.Quiet
}

type SqldMongoConnection struct {
	MongoSSL                 bool   `long:"mongo-ssl" description:"use SSL when connecting to mongo instance"`
	MongoURI                 string `long:"mongo-uri" description:"a mongo URI (https://docs.mongodb.org/manual/reference/connection-string/) to connect to" default:"mongodb://localhost:27017"`
	MongoAllowInvalidCerts   bool   `long:"mongo-sslAllowInvalidCertificates" description:"don't require the certificate presented by the MongoDB server to be valid, when using --mongo-ssl"`
	MongoSSLAllowInvalidHost bool   `long:"mongo-sslAllowInvalidHostnames" description:"bypass the validation for server name"`
	MongoCAFile              string `long:"mongo-sslCAFile" value-name:"<filename>" description:"path to a CA certificate file to use for authenticating certificates from MongoDB, when using --mongo-ssl"`
	MongoSSLCRLFile          string `long:"mongo-sslCRLFile" value-name:"<filename>" description:"the .pem file containing the certificate revocation list"`
	MongoSSLFipsMode         bool   `long:"mongo-sslFIPSMode" description:"use FIPS mode of the installed openssl library"`
	MongoPEMFile             string `long:"mongo-sslPEMKeyFile" value-name:"<filename>" description:"path to a file containing the certificate and private key for connecting to MongoDB, when using --mongo-ssl"`
	MongoPEMFilePassword     string `long:"mongo-sslPEMKeyPassword" description:"password to decrypt private key in mongo-sslPEMKeyFile"`
	MongoTimeout             int64  `long:"mongo-timeout" description:"seconds to wait for a server to respond when connecting or on follow up operations" default:"30" hidden:"true"`
}

func (_ SqldMongoConnection) Name() string {
	return "Mongo Connection"
}

type SqldSchema struct {
	Schema    string `long:"schema" description:"the path to a schema file"`
	SchemaDir string `long:"schemaDirectory" description:"the path to a directory containing schema files to load"`
}

func (_ SqldSchema) Name() string {
	return "Schema"
}

func NewSqldOptions() (SqldOptions, error) {
	opts := SqldOptions{
		SqldClientConnection: &SqldClientConnection{},
		SqldGeneral:          &SqldGeneral{},
		SqldLog:              &SqldLog{},
		SqldMongoConnection:  &SqldMongoConnection{},
		SqldSchema:           &SqldSchema{},
		parser:               flags.NewNamedParser(usage, flags.None),
	}

	groups := []OptionGroup{
		opts.SqldClientConnection,
		opts.SqldGeneral,
		opts.SqldLog,
		opts.SqldMongoConnection,
		opts.SqldSchema,
	}

	for _, group := range groups {
		header := fmt.Sprintf("%s options", group.Name())
		if _, err := opts.parser.AddGroup(header, "", group); err != nil {
			return SqldOptions{}, err
		}
	}

	return opts, nil
}

func (opts SqldOptions) Parse() error {
	// called when -v or --verbose is parsed
	opts.SetVerbosity = func(val string) error {
		if i, err := strconv.Atoi(val); err == nil {
			opts.VLevel = opts.VLevel + i // -v=N or --verbose=N
		} else if matched, _ := regexp.MatchString(`^v+$`, val); matched {
			opts.VLevel = opts.VLevel + len(val) + 1 // handles the -vvv cases
		} else if matched, _ := regexp.MatchString(`^v+=[0-9]$`, val); matched {
			opts.VLevel = parseVal(val) // i.e. -vv=3
		} else if val == "" {
			opts.VLevel = opts.VLevel + 1 // increment for every occurrence of flag
		} else {
			return fmt.Errorf("invalid verbosity value given")
		}
		return nil
	}

	if _, err := opts.parser.Parse(); err != nil {
		return err
	}

	return nil
}

func (o SqldOptions) hasSSLOptionsSet() bool {
	return o.MongoCAFile != "" ||
		o.MongoPEMFile != "" ||
		o.MongoCAFile != "" ||
		o.MongoSSLCRLFile != "" ||
		o.MongoPEMFilePassword != "" ||
		o.MongoSSLFipsMode ||
		o.MongoAllowInvalidCerts
}

func (o SqldOptions) Validate() error {
	if o.Schema == "" && o.SchemaDir == "" {
		return fmt.Errorf("must specify either --schema or --schemaDirectory")
	}
	if o.Schema != "" && o.SchemaDir != "" {
		return fmt.Errorf("must specify only one of --schema or --schemaDirectory")
	}
	if !o.MongoSSL && o.hasSSLOptionsSet() {
		return fmt.Errorf("must specify --mongo-ssl to use SSL options")
	}
	if o.Auth && o.SSLPEMFile == "" {
		return fmt.Errorf("must specify --sslPEMKeyFile when using --auth")
	}

	return nil
}

func (o SqldOptions) PrintHelp(w io.Writer) bool {
	if o.Help {
		o.parser.WriteHelp(w)
	}

	return o.Help
}

func (o SqldOptions) UseFIPSMode() bool {
	return o.SqldMongoConnection.MongoSSLFipsMode
}

func (o SqldOptions) UseSSL() bool {
	return o.SqldMongoConnection.MongoSSL
}

func parseVal(val string) int {
	idx := strings.Index(val, "=")
	ret, err := strconv.Atoi(val[idx+1:])
	if err != nil {
		panic(fmt.Errorf("value was not a valid integer: %v", err))
	}
	return ret
}
