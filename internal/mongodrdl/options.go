package mongodrdl

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/10gen/sqlproxy/log"
	"github.com/jessevdk/go-flags"
)

var drdlUsage = "mongodrdl <options>"

// DrdlOptions hold command line options for mongodrdl.
type DrdlOptions struct {
	*DrdlGeneral
	*DrdlAuth
	*DrdlLog
	*DrdlConnection
	*DrdlKerberos
	*DrdlSSL
	*DrdlNamespace
	*DrdlOutput
	*DrdlSample

	// ReplicaSetName, if specified, will prevent the obtained session from
	// communicating with any server which is not part of a replica set
	// with the given name. The default is to communicate with any server
	// specified or discovered via the servers contacted.
	ReplicaSetName string

	parser *flags.Parser
}

// DrdlAuth holds authentication related command line options for mongodrdl.
// nolint: lll
type DrdlAuth struct {
	Username  string `short:"u" value-name:"<username>" long:"username" description:"username for authentication"`
	Password  string `short:"p" value-name:"<password>" long:"password" description:"password for authentication"`
	Source    string `long:"authenticationDatabase" value-name:"<database-name>" description:"database that holds the user's credentials"`
	Mechanism string `long:"authenticationMechanism" value-name:"<mechanism>" description:"authentication mechanism to use"`
}

// Name returns the name for the authentication-related
// command line options section for mongodrdl.
func (*DrdlAuth) Name() string {
	return "Authentication"
}

// RequiresExternalDB returns true if the desired authentication mechanism
// requires an external database in its operation and false otherwise.
func (auth *DrdlAuth) RequiresExternalDB() bool {
	return auth.Mechanism == "GSSAPI" ||
		auth.Mechanism == "PLAIN" ||
		auth.Mechanism == "MONGODB-X509"
}

// ShouldAskForPassword returns true if a user prompt is required to acquire
// the password for authentication and false otherwise.
func (auth *DrdlAuth) ShouldAskForPassword() bool {
	return auth.Username != "" && auth.Password == "" &&
		!(auth.Mechanism == "MONGODB-X509" || auth.Mechanism == "GSSAPI")
}

// DrdlNamespace holds the namespace - database and collection name -
// information to run mongodrdl on.
// nolint: lll
type DrdlNamespace struct {
	DB         string `short:"d" long:"db" value-name:"<database-name>" description:"database to use"`
	Collection string `short:"c" long:"collection" value-name:"<collection-name>" description:"collection to use"`
}

// Name returns the name for the namespace-related
// command line options section for mongodrdl.
func (*DrdlNamespace) Name() string {
	return "Namespace"
}

// DrdlGeneral holds the help and version flags for mongodrdl.
type DrdlGeneral struct {
	Help    bool `long:"help" description:"print usage"`
	Version bool `long:"version" description:"print the tool version and exit"`
}

// Name returns the name for the general command
// line options section for mongodrdl.
func (*DrdlGeneral) Name() string {
	return "General"
}

// DrdlSSL holds the SSL-related command line options
// for mongodrdl.
// nolint: lll
type DrdlSSL struct {
	Enabled             bool   `long:"ssl" description:"connect to a mongod or mongos that has ssl enabled"`
	SSLCAFile           string `long:"sslCAFile" value-name:"<filename>" description:"the .pem file containing the root certificate chain from the certificate authority"`
	SSLPEMKeyFile       string `long:"sslPEMKeyFile" value-name:"<filename>" description:"the .pem file containing the certificate and key"`
	SSLPEMKeyPassword   string `long:"sslPEMKeyPassword" value-name:"<password>" description:"the password to decrypt the sslPEMKeyFile, if necessary"`
	SSLCRLFile          string `long:"sslCRLFile" value-name:"<filename>" description:"the .pem file containing the certificate revocation list"`
	SSLAllowInvalidCert bool   `long:"sslAllowInvalidCertificates" description:"bypass the validation for server certificates"`
	SSLAllowInvalidHost bool   `long:"sslAllowInvalidHostnames" description:"bypass the validation for server name"`
	SSLFipsMode         bool   `long:"sslFIPSMode" description:"use FIPS mode of the installed openssl library"`
	MinimumTLSVersion   string `long:"minimumTLSVersion" description:"the minimum TLS protocol version to connect to MongoDB" default:"TLS1_1" choice:"TLS1_0" choice:"TLS1_1" choice:"TLS1_2"`
}

// Name returns the name for the SSL-related
// command line options section for mongodrdl.
func (*DrdlSSL) Name() string {
	return "SSL"
}

// DrdlLog holds the logging-related command
// line options for mongodrdl.
// nolint: lll
type DrdlLog struct {
	SetVerbosity func(string) error `short:"v" long:"verbose" value-name:"<level>" description:"more detailed log output (include multiple times for more verbosity, e.g. -vvvvv, or specify a numeric value, e.g. --verbose=N)" optional:"true" optional-value:""`
	Quiet        bool               `long:"quiet" description:"hide all log output"`
	VLevel       int                `no-flag:"true"`
}

// Name returns the name for the logging-related
// command line options section for mongodrdl.
func (*DrdlLog) Name() string {
	return "Log"
}

// Level returns the configured verbosity for mongodrdl.
func (v *DrdlLog) Level() log.Verbosity {
	return log.Verbosity(v.VLevel)
}

// IsQuiet returns true if the configured verbosity is set
// to Quiet - and false otherwise.
func (v *DrdlLog) IsQuiet() bool {
	return v.Quiet
}

// DrdlConnection holds the connection-related command
// line options for mongodrdl.
// nolint: lll
type DrdlConnection struct {
	Host string `short:"h" long:"host" value-name:"<hostname>" description:"mongodb host to connect to (setname/host1,host2 for replica sets)"`
	Port string `long:"port" value-name:"<port>" description:"server port (can also use --host hostname:port)"`
}

// Name returns the name for the connection-related
// command line options section for mongodrdl.
func (*DrdlConnection) Name() string {
	return "Connection"

}

// DrdlKerberos holds the kerberos-related command
// line options for mongodrdl.
// nolint: lll
type DrdlKerberos struct {
	Service     string `long:"gssapiServiceName" value-name:"<service-name>" description:"service name to use when authenticating using GSSAPI/Kerberos ('mongodb' by default)"`
	ServiceHost string `long:"gssapiHostName" value-name:"<host-name>" description:"hostname to use when authenticating using GSSAPI/Kerberos (remote server's address by default)"`
}

// Name returns the name for the kerberos-related
// command line options section for mongodrdl.
func (*DrdlKerberos) Name() string {
	return "Kerberos"
}

// DrdlOutput holds the output-related command
// line options for mongodrdl.
// nolint: lll
type DrdlOutput struct {
	CustomFilterField    string `long:"customFilterField" value-name:"<filter-field-name>" short:"f" description:"the name of the field to use with a custom mongo filter field (defaults to no custom filter field)"`
	UUIDSubtype3Encoding string `long:"uuidSubtype3Encoding" short:"b" description:"encoding used to generate UUID binary subtype 3. old: Old BSON binary subtype representation; csharp: The C#/.NET legacy UUID representation; java: The Java legacy UUID representation" choice:"old" choice:"csharp" choice:"java" default:"old"`
	Out                  string `long:"out" short:"o" description:"output file, or '-' for standard out (defaults to standard out)" default-mask:"-"`
	PreJoined            bool   `long:"preJoined" description:"generate unwound tables including parent columns, effectively resulting in a pre-joined table"`
}

// Name returns the name for the output-related
// command line options section for mongodrdl.
func (*DrdlOutput) Name() string {
	return "Output"
}

// DrdlSample holds the sampling-related command
// line options for mongodrdl.
// nolint: lll
type DrdlSample struct {
	Size int64 `long:"sampleSize" short:"s" description:"the number of documents to sample when generating schema" default:"1000"`
}

// Name returns the name for the sampling-related
// command line options section for mongodrdl.
func (*DrdlSample) Name() string {
	return "Sample"
}

// NewDrdlOptions returns a new instance of the
// DrdlOptions struct.
func NewDrdlOptions() (*DrdlOptions, error) {
	opts := &DrdlOptions{
		DrdlGeneral:    &DrdlGeneral{},
		DrdlAuth:       &DrdlAuth{},
		DrdlLog:        &DrdlLog{},
		DrdlConnection: &DrdlConnection{},
		DrdlKerberos:   &DrdlKerberos{},
		DrdlSSL:        &DrdlSSL{},
		DrdlNamespace:  &DrdlNamespace{},
		DrdlOutput:     &DrdlOutput{},
		DrdlSample:     &DrdlSample{},
		parser:         flags.NewNamedParser(drdlUsage, flags.None),
	}

	groups := []OptionGroup{
		opts.DrdlGeneral,
		opts.DrdlAuth,
		opts.DrdlLog,
		opts.DrdlConnection,
		opts.DrdlKerberos,
		opts.DrdlSSL,
		opts.DrdlNamespace,
		opts.DrdlOutput,
		opts.DrdlSample,
	}

	for _, group := range groups {
		header := fmt.Sprintf("%s options", group.Name())
		if _, err := opts.parser.AddGroup(header, "", group); err != nil {
			return nil, err
		}
	}

	return opts, nil
}

// Parse parses the flags passed to the mongodrdl tool and
// returns any additional unparsed arguments back.
func (o DrdlOptions) Parse() ([]string, error) {
	// called when -v or --verbose is parsed
	o.DrdlLog.SetVerbosity = func(val string) error {
		if i, err := strconv.Atoi(val); err == nil {
			o.VLevel = i // -v=N or --verbose=N
		} else if matched, _ := regexp.MatchString(`^v+$`, val); matched {
			o.VLevel = len(val) + 1 // handles the -vvv cases
		} else if matched, _ := regexp.MatchString(`^v+=[0-9]$`, val); matched {
			o.VLevel = parseVal(val) // i.e. -vv=3
		} else if val == "" {
			o.VLevel = o.VLevel + 1 // increment for every occurrence of flag
		} else {
			return fmt.Errorf("invalid verbosity value given")
		}
		return nil
	}

	// use the quiet verbosity level by default
	o.DrdlLog.VLevel = -1
	args, err := o.parser.Parse()

	// Handle Port and Host. If both Host and Port contain a port spec, we assume the user
	// prefers the one coming from Port, but warn just in case.
	if o.DrdlConnection.Port != "" {
		if strings.Contains(o.DrdlConnection.Host, ":") {
			_, _ = fmt.Fprintf(os.Stderr, "WARNING: port specified in both the '--host' and "+
				"'--port' flags, will use '%s' as port\n", o.DrdlConnection.Port)
			o.DrdlConnection.Host = strings.Split(o.DrdlConnection.Host, ":")[0]
		}
		o.DrdlConnection.Host += ":" + o.DrdlConnection.Port
	}
	return args, err
}

// GetAuthenticationDatabase returns the authentication database with
// which mongodrdl was configured.
func (o DrdlOptions) GetAuthenticationDatabase() string {
	if o.DrdlAuth.Source != "" {
		return o.DrdlAuth.Source
	} else if o.DrdlAuth.RequiresExternalDB() {
		return "$external"
	} else if o.DrdlNamespace != nil && o.DrdlNamespace.DB != "" {
		return o.DrdlNamespace.DB
	}
	return ""
}

// PrintHelp prints the usage message for the tool to stdout. Returns whether or not the
// help flag is specified.
func (o DrdlOptions) PrintHelp(force bool) bool {
	if o.Help || force {
		o.parser.WriteHelp(os.Stdout)
	}
	return o.Help
}

// UseSSL returns true if mongodrdl is configured to use SSL and false otherwise.
func (o DrdlOptions) UseSSL() bool {
	return o.DrdlSSL.Enabled
}

// UseFIPSMode returns true if mongodrdl is configured to use FIPS
// mode within SSL and false otherwise.
func (o DrdlOptions) UseFIPSMode() bool {
	return o.DrdlSSL.SSLFipsMode
}

// Validate validates the options passed to the mongodrdl tool.
// It returns any error found during validation.
func (o DrdlOptions) Validate() error {
	switch {
	case o.DrdlNamespace.DB == "":
		return fmt.Errorf("cannot export a schema without a specified database")
	case o.DrdlNamespace.DB == "" && o.DrdlNamespace.Collection != "":
		return fmt.Errorf("cannot export a schema for a collection without a specified database")
	case o.DrdlSSL.SSLFipsMode && runtime.GOOS == "darwin":
		return fmt.Errorf("this version of mongodrdl was not compiled with FIPS support")
	}

	if !o.Enabled && (o.SSLCAFile != "" ||
		o.SSLPEMKeyFile != "" ||
		o.SSLPEMKeyPassword != "" ||
		o.SSLCRLFile != "" ||
		o.SSLAllowInvalidCert ||
		o.SSLAllowInvalidHost ||
		o.SSLFipsMode) {
		return fmt.Errorf("when specifying SSL options, SSL must be enabled with --ssl")
	}
	return nil
}

func parseVal(val string) int {
	idx := strings.Index(val, "=")
	ret, err := strconv.Atoi(val[idx+1:])
	if err != nil {
		panic(fmt.Errorf("value was not a valid integer: %v", err))
	}
	return ret
}

// OptionGroup is an interface for grouping related option
// flags in mongodrdl.
type OptionGroup interface {
	Name() string
}
