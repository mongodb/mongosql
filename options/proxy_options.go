package options

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/10gen/sqlproxy/util"
	"github.com/jessevdk/go-flags"
)

var usage = "mongosqld <options>"

type SqldOptions struct {
	*SqldClientConnection
	*SqldGeneral
	*SqldLog
	*SqldMongoConnection
	*SqldSchema
	*SqldSocket
	parser *flags.Parser
}

func (s SqldOptions) String() string {
	var buffer bytes.Buffer
	for i, opts := 0, reflect.ValueOf(s); i < opts.NumField(); i++ {
		// Skip any non-exported fields
		if !opts.Field(i).Elem().CanSet() {
			continue
		}
		for j, subOpts := 0, opts.Field(i).Elem(); j < subOpts.NumField(); j++ {
			// Make sure that the field's Kind is a pointer so that we don't attempt to populate SqldLog.SetVerbosity
			if subOpts.Field(j).Kind() == reflect.Ptr && !subOpts.Field(j).IsNil() {
				flagString := subOpts.Type().Field(j).Tag.Get("long")
				if flagString == "" {
					flagString = subOpts.Type().Field(j).Tag.Get("short")
				}
				var argVal string
				if strings.HasSuffix(flagString, "sslPEMKeyPassword") {
					argVal = "<password>"
				} else {
					argVal = fmt.Sprintf("%v", subOpts.Field(j).Elem().Interface())
				}
				buffer.WriteString("--" + flagString + " " + argVal + " ")
			}
		}

	}
	return strings.TrimSpace(buffer.String())
}

type OptionGroup interface {
	Name() string
}

type SqldClientConnection struct {
	Auth                  *bool   `long:"auth" description:"use authentication/authorization ('sslPEMKeyFile' is required when using auth)"`
	Addr                  *string `long:"addr" description:"host address to listen on" default-mask:"127.0.0.1:3307"`
	SSLAllowInvalidCerts  *bool   `long:"sslAllowInvalidCertificates" description:"don't require the certificate presented by the client to be valid"`
	SSLCAFile             *string `long:"sslCAFile" description:"path to a CA certificate file to use for authenticating client certificate"`
	SSLPEMKeyFile         *string `long:"sslPEMKeyFile" description:"path to a file containing the certificate and private key establishing a connection with a client"`
	SSLPEMKeyFilePassword *string `long:"sslPEMKeyPassword" description:"password to decrypt private key in --sslPEMKeyFile"`
}

func (_ SqldClientConnection) Name() string {
	return "Client Connection"
}

type SqldSocket struct {
	FilePermissions  *string `long:"filePermissions" description:"permissions to set on UNIX domain socket file (default to 0700)" default-mask:"0700"`
	NoUnixSocket     *bool   `long:"noUnixSocket" description:"disable listening on UNIX domain sockets"`
	UnixSocketPrefix *string `long:"unixSocketPrefix" description:"alternative directory for UNIX domain sockets (default to /tmp)" default-mask:"/tmp"`
}

func (_ SqldSocket) Name() string {
	return "Socket"
}

type SqldGeneral struct {
	Fork    *bool `long:"fork" description:"fork mongosqld process" hidden:"true"`
	Help    *bool `short:"h" long:"help" description:"print usage"`
	Version *bool `long:"version" description:"display version information"`
}

func (_ SqldGeneral) Name() string {
	return "General"
}

type SqldLog struct {
	LogAppend    *bool              `long:"logAppend" description:"append new logging output to existing log file"`
	LogPath      *string            `long:"logPath" description:"path to a log file for storing logging output (defaults to stderr)"`
	SetVerbosity func(string) error `short:"v" long:"verbose" value-name:"<level>" description:"more detailed log output (include multiple times for more verbosity, e.g. -vvvvv, or specify a numeric value, e.g. --verbose=N)" optional:"true" optional-value:""`
	Quiet        *bool              `long:"quiet" description:"hide all log output"`
	VLevel       *int               `no-flag:"true" long:"verbose"`
}

func (_ SqldLog) Name() string {
	return "Log"
}

func (lo SqldLog) Level() int {
	if lo.VLevel == nil {
		return 0
	}
	return *lo.VLevel
}

func (lo SqldLog) IsQuiet() bool {
	return !isFalseOrUnset(lo.Quiet)
}

type SqldMongoConnection struct {
	MongoSSL                  *bool   `long:"mongo-ssl" description:"use SSL when connecting to mongo instance"`
	MongoURI                  *string `long:"mongo-uri" description:"a mongo URI (https://docs.mongodb.org/manual/reference/connection-string/) to connect to" default-mask:"mongodb://localhost:27017"`
	MongoAllowInvalidCerts    *bool   `long:"mongo-sslAllowInvalidCertificates" description:"don't require the certificate presented by the MongoDB server to be valid, when using --mongo-ssl"`
	MongoSSLAllowInvalidHost  *bool   `long:"mongo-sslAllowInvalidHostnames" description:"bypass the validation for server name"`
	MongoCAFile               *string `long:"mongo-sslCAFile" value-name:"<filename>" description:"path to a CA certificate file to use for authenticating certificates from MongoDB, when using --mongo-ssl"`
	MongoSSLCRLFile           *string `long:"mongo-sslCRLFile" value-name:"<filename>" description:"the .pem file containing the certificate revocation list"`
	MongoSSLFipsMode          *bool   `long:"mongo-sslFIPSMode" description:"use FIPS mode of the installed openssl library"`
	MongoPEMKeyFile           *string `long:"mongo-sslPEMKeyFile" value-name:"<filename>" description:"path to a file containing the certificate and private key for connecting to MongoDB, when using --mongo-ssl"`
	MongoPEMKeyFilePassword   *string `long:"mongo-sslPEMKeyPassword" description:"password to decrypt private key in mongo-sslPEMKeyFile"`
	MongoVersionCompatibility *string `long:"mongo-versionCompatibility" description:"indicates the mongodb version with which to be compatible (only necessary when used with mixed version replica sets)."`
}

func (_ SqldMongoConnection) Name() string {
	return "Mongo Connection"
}

type SqldSchema struct {
	Schema    *string `long:"schema" description:"the path to a schema file"`
	SchemaDir *string `long:"schemaDirectory" description:"the path to a directory containing schema files to load"`
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
		SqldSocket:           &SqldSocket{},
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

	if !util.IsWindowsOS {
		header := fmt.Sprintf("%s options", opts.SqldSocket.Name())
		if _, err := opts.parser.AddGroup(header, "", opts.SqldSocket); err != nil {
			return SqldOptions{}, err
		}
	}

	return opts, nil
}

func (opts SqldOptions) Parse() error {
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

	if _, err := opts.parser.Parse(); err != nil {
		return err
	}

	return nil
}

func (o SqldOptions) hasSSLOptionsSet() bool {
	return !isEmptyOrUnset(o.MongoCAFile) ||
		!isEmptyOrUnset(o.MongoPEMKeyFile) ||
		!isEmptyOrUnset(o.MongoSSLCRLFile) ||
		!isEmptyOrUnset(o.MongoPEMKeyFilePassword) ||
		!isFalseOrUnset(o.MongoSSLFipsMode) ||
		!isFalseOrUnset(o.MongoAllowInvalidCerts)
}

func (o SqldOptions) Validate() error {
	if !isFalseOrUnset(o.Version) || !isFalseOrUnset(o.Help) {
		return nil
	}
	if isEmptyOrUnset(o.Schema) && isEmptyOrUnset(o.SchemaDir) {
		return fmt.Errorf("must specify either --schema or --schemaDirectory")
	}
	if !isEmptyOrUnset(o.Schema) && !isEmptyOrUnset(o.SchemaDir) {
		return fmt.Errorf("must specify only one of --schema or --schemaDirectory")
	}
	if isFalseOrUnset(o.MongoSSL) && o.hasSSLOptionsSet() {
		return fmt.Errorf("must specify --mongo-ssl to use SSL options")
	}
	if !isFalseOrUnset(o.MongoSSLFipsMode) && runtime.GOOS == "darwin" {
		return fmt.Errorf("this version of mongosqld was not compiled with FIPS support")
	}
	if o.FilePermissions != nil {
		if _, err := strconv.ParseInt(*o.FilePermissions, 8, 64); err != nil {
			return fmt.Errorf("value after --filePermissions must be valid octal")
		}
	}
	if util.IsWindowsOS {
		if o.NoUnixSocket != nil {
			return fmt.Errorf("cannot use Unix-specific option --noUnixSocket on Windows")
		}
		if o.UnixSocketPrefix != nil {
			return fmt.Errorf("cannot use Unix-specific option --unixSocketPrefix on Windows")
		}
		if o.FilePermissions != nil {
			return fmt.Errorf("cannot use Unix-specific option --filePermissions on Windows")
		}
	}
	if !isFalseOrUnset(o.Fork) {
		return fmt.Errorf("--fork is no longer supported")
	}

	return nil
}

// EnsureOptsNotNil initializes all member pointers of all member structs of the SqldOptions object accessible via
// optsPtr to be non-nil. The values stored at the resulting addresses are either the zero value for the type or
// the default-mask value of the member pointer, if present.
func EnsureOptsNotNil(optsPtr *SqldOptions) {
	// We assume optsPtr and all of the members of *optsPtr are non-nil because they are created via NewSqldOpts
	if optsPtr == nil {
		panic("nil SqldOptions pointer")
	}
	for i, opts := 0, reflect.ValueOf(optsPtr).Elem(); i < opts.NumField(); i++ {
		if opts.Field(i).Kind() == reflect.Ptr && opts.Field(i).IsNil() {
			panic("nil SqldOptions field")
		}
		for j, subOpts := 0, opts.Field(i).Elem(); j < subOpts.NumField(); j++ {
			// Allocate space for field and set value to default-mask value, if present,
			// or zero value otherwise.
			if subOpts.Field(j).Kind() == reflect.Ptr && subOpts.Field(j).IsNil() {
				subOpts.Field(j).Set(reflect.New(subOpts.Field(j).Type().Elem()))
				if defaultVal := subOpts.Type().Field(j).Tag.Get("default-mask"); defaultVal != "" {
					if subOpts.Field(j).Elem().Kind() != reflect.String {
						panic("Support for non-string default-mask values has not yet been implemented.")
					}
					subOpts.Field(j).Elem().Set(reflect.ValueOf(defaultVal))
				}
			}
		}
	}

}

func (o SqldOptions) PrintHelp(w io.Writer) bool {
	if !isFalseOrUnset(o.Help) {
		o.parser.WriteHelp(w)
	}

	return !isFalseOrUnset(o.Help)
}

func (o SqldOptions) UseFIPSMode() bool {
	return !isFalseOrUnset(o.SqldMongoConnection.MongoSSLFipsMode)
}

func (o SqldOptions) UseSSL() bool {
	return !isFalseOrUnset(o.SqldMongoConnection.MongoSSL)
}

func parseVal(val string) int {
	idx := strings.Index(val, "=")
	ret, err := strconv.Atoi(val[idx+1:])
	if err != nil {
		panic(fmt.Errorf("value was not a valid integer: %v", err))
	}
	return ret
}

func isFalseOrUnset(b *bool) bool {
	return b == nil || !*b
}

func isEmptyOrUnset(s *string) bool {
	return s == nil || *s == ""
}
