package options

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/jessevdk/go-flags"
)

var drdlUsage = "mongodrdl <options>"

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

	// Force direct connection to the server and disable the
	// drivers automatic repl set discovery logic.
	Direct bool

	// ReplicaSetName, if specified, will prevent the obtained session from
	// communicating with any server which is not part of a replica set
	// with the given name. The default is to communicate with any server
	// specified or discovered via the servers contacted.
	ReplicaSetName string

	parser *flags.Parser
}

type DrdlAuth struct {
	Username  string `short:"u" value-name:"<username>" long:"username" description:"username for authentication"`
	Password  string `short:"p" value-name:"<password>" long:"password" description:"password for authentication"`
	Source    string `long:"authenticationDatabase" value-name:"<database-name>" description:"database that holds the user's credentials"`
	Mechanism string `long:"authenticationMechanism" value-name:"<mechanism>" description:"authentication mechanism to use"`
}

func (*DrdlAuth) Name() string {
	return "Authentication"
}

func (auth *DrdlAuth) RequiresExternalDB() bool {
	return auth.Mechanism == "GSSAPI" || auth.Mechanism == "PLAIN" || auth.Mechanism == "MONGODB-X509"
}

func (auth *DrdlAuth) ShouldAskForPassword() bool {
	return auth.Username != "" && auth.Password == "" &&
		!(auth.Mechanism == "MONGODB-X509" || auth.Mechanism == "GSSAPI")
}

type DrdlNamespace struct {
	DB         string `short:"d" long:"db" value-name:"<database-name>" description:"database to use"`
	Collection string `short:"c" long:"collection" value-name:"<collection-name>" description:"collection to use"`
}

func (*DrdlNamespace) Name() string {
	return "Namespace"
}

type DrdlGeneral struct {
	Help    bool `long:"help" description:"print usage"`
	Version bool `long:"version" description:"print the tool version and exit"`
}

func (*DrdlGeneral) Name() string {
	return "General"
}

type DrdlSSL struct {
	UseSSL              bool   `long:"ssl" description:"connect to a mongod or mongos that has ssl enabled"`
	SSLCAFile           string `long:"sslCAFile" value-name:"<filename>" description:"the .pem file containing the root certificate chain from the certificate authority"`
	SSLPEMKeyFile       string `long:"sslPEMKeyFile" value-name:"<filename>" description:"the .pem file containing the certificate and key"`
	SSLPEMKeyPassword   string `long:"sslPEMKeyPassword" value-name:"<password>" description:"the password to decrypt the sslPEMKeyFile, if necessary"`
	SSLCRLFile          string `long:"sslCRLFile" value-name:"<filename>" description:"the .pem file containing the certificate revocation list"`
	SSLAllowInvalidCert bool   `long:"sslAllowInvalidCertificates" description:"bypass the validation for server certificates"`
	SSLAllowInvalidHost bool   `long:"sslAllowInvalidHostnames" description:"bypass the validation for server name"`
	SSLFipsMode         bool   `long:"sslFIPSMode" description:"use FIPS mode of the installed openssl library"`
}

func (*DrdlSSL) Name() string {
	return "SSL"
}

type DrdlLog struct {
	SetVerbosity func(string) error `short:"v" long:"verbose" value-name:"<level>" description:"more detailed log output (include multiple times for more verbosity, e.g. -vvvvv, or specify a numeric value, e.g. --verbose=N)" optional:"true" optional-value:""`
	Quiet        bool               `long:"quiet" description:"hide all log output"`
	VLevel       int                `no-flag:"true"`
}

func (*DrdlLog) Name() string {
	return "Log"
}

func (v *DrdlLog) Level() int {
	return v.VLevel
}

func (v *DrdlLog) IsQuiet() bool {
	return v.Quiet
}

type DrdlConnection struct {
	Host    string `short:"h" long:"host" value-name:"<hostname>" description:"mongodb host to connect to (setname/host1,host2 for replica sets)"`
	Port    string `long:"port" value-name:"<port>" description:"server port (can also use --host hostname:port)"`
	Timeout int    `long:"dialTimeout" default:"3" hidden:"true" description:"dial timeout in seconds"`
}

func (*DrdlConnection) Name() string {
	return "Connection"
}

type DrdlKerberos struct {
	Service     string `long:"gssapiServiceName" value-name:"<service-name>" description:"service name to use when authenticating using GSSAPI/Kerberos ('mongodb' by default)"`
	ServiceHost string `long:"gssapiHostName" value-name:"<host-name>" description:"hostname to use when authenticating using GSSAPI/Kerberos (remote server's address by default)"`
}

func (*DrdlKerberos) Name() string {
	return "Kerberos"
}

type DrdlOutput struct {
	CustomFilterField    string `long:"customFilterField" value-name:"<filter-field-name>" short:"f" description:"the name of the field to use with a custom mongo filter field (defaults to no custom filter field)"`
	UUIDSubtype3Encoding string `long:"uuidSubtype3Encoding" short:"b" description:"encoding used to generate UUID binary subtype 3. old: Old BSON binary subtype representation; csharp: The C#/.NET legacy UUID representation; java: The Java legacy UUID representation" choice:"old" choice:"csharp" choice:"java"`
	Out                  string `long:"out" short:"o" description:"output file, or '-' for standard out (defaults to standard out)" default-mask:"-"`
}

func (*DrdlOutput) Name() string {
	return "Output"
}

type DrdlSample struct {
	SampleSize int64 `long:"sampleSize" short:"s" description:"the number of documents to sample when generating schema" default:"1000"`
}

func (*DrdlSample) Name() string {
	return "Sample"
}

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

func (opts DrdlOptions) Parse() ([]string, error) {
	opts.SetVerbosity = func(val string) error {
		if i, err := strconv.Atoi(val); err == nil {
			opts.VLevel = opts.VLevel + i
		} else if matched, _ := regexp.MatchString(`^v+$`, val); matched {
			opts.VLevel = opts.VLevel + len(val) + 1
		} else if matched, _ := regexp.MatchString(`^v+=[0-9]$`, val); matched {
			opts.VLevel = parseVal(val)
		} else if val == "" {
			opts.VLevel = opts.VLevel + 1
		} else {
			return fmt.Errorf("invalid verbosity value given")
		}
		return nil
	}

	args, err := opts.parser.Parse()
	return args, err
}

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

// Print the usage message for the tool to stdout. Returns whether or not the
// help flag is specified.
func (o DrdlOptions) PrintHelp(force bool) bool {
	if o.Help || force {
		o.parser.WriteHelp(os.Stdout)
	}
	return o.Help
}

func (o DrdlOptions) UseSSL() bool {
	return o.DrdlSSL.UseSSL
}

func (o DrdlOptions) UseFIPSMode() bool {
	return o.DrdlSSL.SSLFipsMode
}

func (o DrdlOptions) Validate() error {
	switch {
	case o.DrdlNamespace.DB == "":
		return fmt.Errorf("cannot export a schema without a specified database")
	case o.DrdlNamespace.DB == "" && o.DrdlNamespace.Collection != "":
		return fmt.Errorf("cannot export a schema for a collection without a specified database")
	}
	return nil
}
