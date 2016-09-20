package db

import (
	"crypto/tls"
	"net"
	"strings"
	"time"

	proxyOpts "github.com/10gen/sqlproxy/options"

	"github.com/mongodb/mongo-tools/common/options"
	"github.com/mongodb/mongo-tools/common/util"
	"gopkg.in/mgo.v2"
)

type SessionProvider struct {
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
	timeout := time.Duration(options.DefaultDialTimeoutSeconds) * time.Second
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
		Timeout:        timeout,
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

func NewSessionProvider(opts proxyOpts.Options) (*SessionProvider, error) {
	e := new(SessionProvider)

	info, err := proxyOpts.GetDialInfo(opts)
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
