package evaluator

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/10gen/sqlproxy/schema"
	toolsdb "github.com/mongodb/mongo-tools/common/db"
	"github.com/mongodb/mongo-tools/common/options"
	"github.com/mongodb/mongo-tools/common/util"
	"gopkg.in/mgo.v2"
	"io/ioutil"
	"net"
	"strings"
)

type SessionProvider struct {
	cfg           *schema.Schema
	globalSession *mgo.Session
}

// SSLDBConnector is a simple implementation of the tools DBConnector interface
// that just uses go's built-in native SSL to dial out, replacing the need for OpenSSL
// to do the test harness setup (invoking mongorestore).
type SSLDBConnector struct {
	sslConf  *tls.Config
	dialInfo *mgo.DialInfo
}

// Configure sets up the SSLDBConnector based on connection settings found in a mongo-tools
// options struct.
func (ssldbc *SSLDBConnector) Configure(opts options.ToolOptions) error {
	ssldbc.sslConf = &tls.Config{}
	connectionAddrs := util.CreateConnectionAddrs(opts.Host, opts.Port)
	if opts.SSLAllowInvalidCert || opts.SSLCAFile == "" {
		ssldbc.sslConf.InsecureSkipVerify = true
	}
	if opts.SSLPEMKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(opts.SSLPEMKeyFile, opts.SSLPEMKeyFile)
		if err != nil {
			return err
		}
		ssldbc.sslConf.Certificates = []tls.Certificate{cert}
	}

	// set up the dial info
	ssldbc.dialInfo = &mgo.DialInfo{
		DialServer: func(addr *mgo.ServerAddr) (net.Conn, error) {
			return tls.Dial("tcp", addr.String(), ssldbc.sslConf)
		},
		Addrs:          connectionAddrs,
		Direct:         opts.Direct,
		Mechanism:      opts.Auth.Mechanism,
		Password:       opts.Auth.Password,
		ReplicaSetName: opts.ReplicaSetName,
		Source:         opts.GetAuthenticationDatabase(),
		Timeout:        toolsdb.DefaultDialTimeout,
		Username:       opts.Auth.Username,
	}

	return nil

}

func (ssldbc *SSLDBConnector) GetNewSession() (*mgo.Session, error) {
	session, err := mgo.DialWithInfo(ssldbc.dialInfo)
	if err != nil {
		return nil, err
	}
	return session, err
}

// GetDialInfo populates a *mgo.DialInfo object according to
// the settings present in a *schema.Schema object.
// If SSL is enabled, will parse out the relevant SSL config fields
// to construct a tls.Config and use it to replace the DialServer method
// with one that uses tls.Dial.
func GetDialInfo(cfg *schema.Schema) (*mgo.DialInfo, error) {
	dialInfo, err := mgo.ParseURL(cfg.Url)
	if err != nil {
		return nil, err
	}
	if cfg.SSL != nil {
		var certs []tls.Certificate
		var rootCA *x509.CertPool
		if len(cfg.SSL.PEMKeyFile) > 0 {
			// assume same file includes both private key and cert data.
			cert, err := tls.LoadX509KeyPair(cfg.SSL.PEMKeyFile, cfg.SSL.PEMKeyFile)
			if err != nil {
				return nil, fmt.Errorf("failed to load pem key file '%v': %v", cfg.SSL.PEMKeyFile, err)
			}
			certs = append(certs, cert)
		}

		if len(cfg.SSL.CAFile) > 0 {
			caCert, err := ioutil.ReadFile(cfg.SSL.CAFile)
			if err != nil {
				return nil, fmt.Errorf("failed to load CA file '%v': %v", cfg.SSL.CAFile)
			}
			rootCA = x509.NewCertPool()
			ok := rootCA.AppendCertsFromPEM(caCert)
			if !ok {
				return nil, fmt.Errorf("unable to append valid cert from PEM file '%v'", cfg.SSL.CAFile)
			}
		}

		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {

			sslConf := &tls.Config{
				// in the future, certificates could be included here to allow x509 auth.
				Certificates:       certs,
				RootCAs:            rootCA,
				InsecureSkipVerify: cfg.SSL.AllowInvalidCerts,
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

func NewSessionProvider(cfg *schema.Schema) (*SessionProvider, error) {
	e := new(SessionProvider)
	e.cfg = cfg

	info, err := GetDialInfo(cfg)
	if err != nil {
		return nil, err
	}

	session, err := mgo.DialWithInfo(info)
	if err != nil {
		return nil, err
	}
	e.globalSession = session

	return e, nil
}

func (e *SessionProvider) GetSession() *mgo.Session {
	if e.globalSession == nil {
		panic("No global session has been set")
	}
	return e.globalSession.Copy()
}

func (e *SessionProvider) Namespace(session *mgo.Session, fullName string) *mgo.Collection {
	pcs := strings.SplitN(fullName, ".", 2)
	return session.DB(pcs[0]).C(pcs[1])
}
