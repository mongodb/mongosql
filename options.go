package sqlproxy

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"gopkg.in/mgo.v2"
)

type Options struct {
	Addr    string `long:"addr" description:"host address to listen on" default:"127.0.0.1:3307"`
	Verbose []bool `short:"v" long:"verbose" description:"more detailed log output (include multiple times for more verbosity, e.g. -vvvvv)"`

	Schema    string `long:"schema" description:"the path to a schema file"`
	SchemaDir string `long:"schema-dir" description:"the path to a directory containing schema files to load"`

	// Mongo Options
	MongoURI               string `long:"mongo-uri" description:"a mongo URI (https://docs.mongodb.org/manual/reference/connection-string/) to connect to" default:"mongodb://localhost:27017"`
	MongoSSL               bool   `long:"mongo-ssl" description:"use SSL when connecting to mongo instance"`
	MongoPEMFile           string `long:"mongo-ssl-pem-file" description:"path to a file containing the cert and private key for connecting to MongoDB, when using --mongo-ssl"`
	MongoAllowInvalidCerts bool   `long:"mongo-ssl-allow-invalid-certs" description:"don't require the cert presented by the MongoDB server to be valid, when using --mongo-ssl"`
	MongoCAFile            string `long:"mongo-ssl-ca-file" description:"path to a CA certs file to use for authenticating certs from MongoDB, when using --mongo-ssl"`
	MongoTimeout           int64  `long:"mongo-timeout" description:"seconds to wait for a server to respond when connecting or on follow up operations" default:"30" hidden:"true"`

	// SSL Options
	SSLPEMFile           string `long:"ssl-pem-file" description:"path to a file containing the cert and private key estabolishing a connection with a client"`
	SSLAllowInvalidCerts bool   `long:"ssl-allow-invalid-certs" description:"don't require the cert presented by the client to be valid"`
	SSLCAFile            string `long:"ssl-ca-file" description:"path to a CA certs file to use for authenticating certs from a client"`

	// Auth Options
	Auth bool `long:"auth" description:"use authentication/authorization ('ssl-pem-file' is required when using auth)"`
}

func (o Options) Level() int {
	return len(o.Verbose) + 1
}

func (o Options) IsQuiet() bool {
	return false
}

func (o Options) Validate() error {
	if o.Schema == "" && o.SchemaDir == "" {
		return fmt.Errorf("must specify either --schema or --schema-dir")
	}
	if o.Schema != "" && o.SchemaDir != "" {
		return fmt.Errorf("must specify only one of --schema or --schema-dir")
	}
	if !o.MongoSSL && (len(o.MongoPEMFile) > 0 || len(o.MongoCAFile) > 0 || o.MongoAllowInvalidCerts) {
		return fmt.Errorf("must specify --mongo-ssl to use SSL options")
	}
	if o.Auth && o.SSLPEMFile == "" {
		return fmt.Errorf("must specify --ssl-pem-file when using --auth")
	}

	return nil
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
		var certs []tls.Certificate
		var rootCA *x509.CertPool
		if len(opts.MongoPEMFile) > 0 {
			// assume same file includes both private key and cert data.
			cert, err := tls.LoadX509KeyPair(opts.MongoPEMFile, opts.MongoPEMFile)
			if err != nil {
				return nil, fmt.Errorf("failed to load pem key file '%v': %v", opts.MongoPEMFile, err)
			}
			certs = append(certs, cert)
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
				Certificates:       certs,
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
