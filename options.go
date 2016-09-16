package sqlproxy

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
	"gopkg.in/mgo.v2"
)

var usage = "mongosqld <options>"

type Options struct {
	*AuthOpts
	*ConnectionOpts
	*GeneralOpts
	*LogOpts
	*MongoOpts
	*SchemaOpts
	*SSLOpts
	parser *flags.Parser
}

type OptionGroup interface {
	Name() string
}

type AuthOpts struct {
	Auth bool `long:"auth" description:"use authentication/authorization ('sslPEMKeyFile' is required when using auth)"`
}

func (_ AuthOpts) Name() string {
	return "Authentication"
}

type ConnectionOpts struct {
	Addr string `long:"addr" description:"host address to listen on" default:"127.0.0.1:3307"`
}

func (_ ConnectionOpts) Name() string {
	return "Connection"
}

type GeneralOpts struct {
	Help    bool `short:"h" long:"help" description:"print usage"`
	Version bool `long:"version" description:"display version information"`
}

func (_ GeneralOpts) Name() string {
	return "General"
}

type LogOpts struct {
	LogAppend    bool               `long:"logAppend" description:"append new logging output to existing log file"`
	LogPath      string             `long:"logPath" description:"path to a log file for storing logging output (defaults to stderr)"`
	SetVerbosity func(string) error `short:"v" long:"verbose" value-name:"<level>" description:"more detailed log output (include multiple times for more verbosity, e.g. -vvvvv, or specify a numeric value, e.g. --verbose=N)" optional:"true" optional-value:""`
	Quiet        bool               `long:"quiet" description:"hide all log output"`
	VLevel       int                `no-flag:"true"`
}

func (_ LogOpts) Name() string {
	return "Log"
}

func (lo LogOpts) Level() int {
	return lo.VLevel
}

func (lo LogOpts) IsQuiet() bool {
	return lo.Quiet
}

type MongoOpts struct {
	MongoAllowInvalidCerts bool   `long:"mongo-sslAllowInvalidCertificates" description:"don't require the certificate presented by the MongoDB server to be valid, when using --mongo-ssl"`
	MongoCAFile            string `long:"mongo-sslCAFile" description:"path to a CA certificate file to use for authenticating certificates from MongoDB, when using --mongo-ssl"`
	MongoPEMFile           string `long:"mongo-sslPEMKeyFile" description:"path to a file containing the certificate and private key for connecting to MongoDB, when using --mongo-ssl"`
	MongoPEMFilePassword   string `long:"mongo-sslPEMKeyPassword" description:"password to decrypt private key in mongo-sslPEMKeyFile"`
	MongoSSL               bool   `long:"mongo-ssl" description:"use SSL when connecting to mongo instance"`
	MongoTimeout           int64  `long:"mongo-timeout" description:"seconds to wait for a server to respond when connecting or on follow up operations" default:"30" hidden:"true"`
	MongoURI               string `long:"mongo-uri" description:"a mongo URI (https://docs.mongodb.org/manual/reference/connection-string/) to connect to" default:"mongodb://localhost:27017"`
}

func (_ MongoOpts) Name() string {
	return "Mongo"
}

type SchemaOpts struct {
	Schema    string `long:"schema" description:"the path to a schema file"`
	SchemaDir string `long:"schemaDirectory" description:"the path to a directory containing schema files to load"`
}

func (_ SchemaOpts) Name() string {
	return "Schema"
}

type SSLOpts struct {
	SSLAllowInvalidCerts bool   `long:"sslAllowInvalidCertificates" description:"don't require the certificate presented by the client to be valid"`
	SSLCAFile            string `long:"sslCAFile" description:"path to a CA certificate file to use for authenticating client certificate"`
	SSLPEMFile           string `long:"sslPEMKeyFile" description:"path to a file containing the certificate and private key establishing a connection with a client"`
}

func (_ SSLOpts) Name() string {
	return "Client SSL"
}

func NewOptions() (Options, error) {
	opts := Options{
		AuthOpts:       &AuthOpts{},
		ConnectionOpts: &ConnectionOpts{},
		GeneralOpts:    &GeneralOpts{},
		LogOpts:        &LogOpts{},
		MongoOpts:      &MongoOpts{},
		SchemaOpts:     &SchemaOpts{},
		SSLOpts:        &SSLOpts{},
		parser:         flags.NewNamedParser(usage, flags.None),
	}

	groups := []OptionGroup{
		opts.AuthOpts,
		opts.ConnectionOpts,
		opts.GeneralOpts,
		opts.LogOpts,
		opts.MongoOpts,
		opts.SchemaOpts,
		opts.SSLOpts,
	}

	for _, group := range groups {
		header := fmt.Sprintf("%s options", group.Name())
		if _, err := opts.parser.AddGroup(header, "", group); err != nil {
			return Options{}, err
		}
	}

	return opts, nil
}

func (opts Options) Parse() error {
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

func (o Options) Validate() error {
	if o.Schema == "" && o.SchemaDir == "" {
		return fmt.Errorf("must specify either --schema or --schemaDirectory")
	}
	if o.Schema != "" && o.SchemaDir != "" {
		return fmt.Errorf("must specify only one of --schema or --schemaDirectory")
	}
	if !o.MongoSSL && (len(o.MongoPEMFile) > 0 || len(o.MongoCAFile) > 0 || o.MongoAllowInvalidCerts) {
		return fmt.Errorf("must specify --mongo-ssl to use SSL options")
	}
	if o.Auth && o.SSLPEMFile == "" {
		return fmt.Errorf("must specify --sslPEMKeyFile when using --auth")
	}

	return nil
}

func (o Options) PrintHelp(w io.Writer) bool {
	if o.Help {
		o.parser.WriteHelp(w)
	}

	return o.Help
}

// GetDialInfo populates a *mgo.DialInfo object according to
// the settings present in a *schema.Schema object.
// If SSL is enabled, will parse out the relevant SSL config fields
// to construct a tls.Config and use it to replace the DialServer method
// with one that uses tls.Dial.
func GetDialInfo(opts Options) (*mgo.DialInfo, error) {
	dialInfo, err := mgo.ParseURL(opts.MongoURI)
	if err != nil {
		return nil, err
	}

	if dialInfo.Username != "" || dialInfo.Password != "" {
		return nil, fmt.Errorf("--mongo-uri may not contain any authentication information")
	}

	dialInfo.Timeout = time.Duration(opts.MongoTimeout) * time.Second

	if opts.MongoSSL {
		var certificates []tls.Certificate
		var rootCA *x509.CertPool

		if len(opts.MongoPEMFile) > 0 {
			// assume same file includes both private key and certificate data
			if len(opts.MongoPEMFilePassword) == 0 {
				certificate, err := tls.LoadX509KeyPair(opts.MongoPEMFile, opts.MongoPEMFile)
				if err != nil {
					return nil, fmt.Errorf("failed to load PEM file '%v': %v", opts.MongoPEMFile, err)
				}
				certificates = append(certificates, certificate)
			} else {
				pemFile, err := ioutil.ReadFile(opts.MongoPEMFile)
				if err != nil {
					return nil, fmt.Errorf("failed to load PEM file '%v': %v", opts.MongoPEMFile, err)
				}

				var parsedPEMBlock, keyPEMBlock *pem.Block
				var certPEMBlock []byte

				for {
					parsedPEMBlock, pemFile = pem.Decode(pemFile)
					if parsedPEMBlock == nil {
						break
					}

					if (parsedPEMBlock.Type == "PRIVATE KEY" || strings.HasSuffix(parsedPEMBlock.Type, " PRIVATE KEY")) && keyPEMBlock == nil {
						decryptedBlock, err := x509.DecryptPEMBlock(parsedPEMBlock, []byte(opts.MongoPEMFilePassword))
						if err != nil {
							return nil, fmt.Errorf("failed to decrypt PEM file '%v': %v", opts.MongoPEMFile, err)
						}
						keyPEMBlock = &pem.Block{Type: parsedPEMBlock.Type, Bytes: decryptedBlock}
					} else {
						certPEMBlock = append(certPEMBlock, pem.EncodeToMemory(parsedPEMBlock)...)
					}
				}

				certificate, err := tls.X509KeyPair(certPEMBlock, pem.EncodeToMemory(keyPEMBlock))
				if err != nil {
					return nil, fmt.Errorf("failed to load PEM certificate '%v': %v", opts.MongoPEMFile, err)
				}
				certificates = append(certificates, certificate)
			}

		}

		if len(opts.MongoCAFile) > 0 {
			caCert, err := ioutil.ReadFile(opts.MongoCAFile)
			if err != nil {
				return nil, fmt.Errorf("failed to load CA file '%v': %v", opts.MongoCAFile)
			}
			rootCA = x509.NewCertPool()
			ok := rootCA.AppendCertsFromPEM(caCert)
			if !ok {
				return nil, fmt.Errorf("unable to append valid cert from PEM file '%v'", opts.MongoCAFile)
			}
		}

		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			sslConf := &tls.Config{
				// in the future, certificates could be included here to allow x509 auth.
				Certificates:       certificates,
				RootCAs:            rootCA,
				InsecureSkipVerify: opts.MongoAllowInvalidCerts,
			}
			var err error
			sslConf.ServerName, _, err = net.SplitHostPort(addr.String())
			if err != nil {
				return nil, err
			}
			if sslConf.ServerName == "" {
				sslConf.ServerName = "localhost"
			}
			c, err := tls.Dial("tcp", addr.String(), sslConf)
			return c, err
		}
	}

	return dialInfo, nil
}

func parseVal(val string) int {
	idx := strings.Index(val, "=")
	ret, err := strconv.Atoi(val[idx+1:])
	if err != nil {
		panic(fmt.Errorf("value was not a valid integer: %v", err))
	}
	return ret
}
