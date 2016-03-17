package sqlproxy

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"gopkg.in/mgo.v2"
	"io/ioutil"
	"net"
)

type Options struct {
	Addr        string `long:"addr" description:"host address to listen on" default:"127.0.0.1:3307"`
	SQLUser     string `long:"sql-user" description:"username to require authentication as from MySQL clients"`
	SQLPassword string `long:"sql-password" description:"password to require from MySQL clients"`
	Verbose     []bool `short:"v" long:"verbose" description:"more detailed log output (include multiple times for more verbosity, e.g. -vvvvv)"`

	MongoURI  string `long:"mongo-uri" description:"a mongo URI (https://docs.mongodb.org/manual/reference/connection-string/) to connect to" default:"mongodb://localhost:27017"`
	Schema    string `long:"schema" description:"the path to a schema file"`
	SchemaDir string `long:"schema-dir" description:"the path to a directory containing schema files to load"`

	// SSL options
	MongoSSL               bool   `long:"mongo-ssl" description:"use SSL when connecting to mongo instance"`
	MongoPEMFile           string `long:"mongo-pem" description:"path to a file containing the cert and private key for connecting to MongoDB, when using --mongo-ssl"`
	MongoAllowInvalidCerts bool   `long:"mongo-allow-invalid-certs" description:"don't require the cert presented by the MongoDB server to be valid, when using --mongo-ssl"`
	MongoCAFile            string `long:"mongo-ca-file" description:"path to a CA certs file to use for authenticating certs from MongoDB, when using --mongo-ssl"`
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
	if !o.MongoSSL && (len(o.MongoPEMFile) > 0 || len(o.MongoCAFile) > 0 || o.MongoAllowInvalidCerts) {
		return fmt.Errorf("must specify --mongo-ssl to use SSL options")
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
