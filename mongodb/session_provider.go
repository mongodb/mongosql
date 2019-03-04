package mongodb

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/10gen/mongo-go-driver/mongo/connstring"
	"github.com/10gen/mongo-go-driver/mongo/model"
	"github.com/10gen/mongo-go-driver/mongo/private/cluster"
	"github.com/10gen/mongo-go-driver/mongo/private/conn"
	"github.com/10gen/mongo-go-driver/mongo/private/msg"
	"github.com/10gen/mongo-go-driver/mongo/private/ops"
	"github.com/10gen/mongo-go-driver/mongo/private/server"
	"github.com/10gen/mongo-go-driver/mongo/readpref"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/mongodb/ssl"
)

// NewDrdlSessionProvider creates a new session provider for mongodrdl.
func NewDrdlSessionProvider(rp *readpref.ReadPref, c *cluster.Cluster, timeout time.Duration,
	numConns int) *SessionProvider {
	return &SessionProvider{
		rp:             rp,
		c:              c,
		connectTimeout: timeout,
		numConns:       numConns,
	}
}

// NewSqldSessionProvider creates a new session provider for mongosql.
func NewSqldSessionProvider(cfg *config.Config) (*SessionProvider, error) {

	uri := cfg.MongoDB.Net.URI
	if !strings.HasPrefix(uri, bsonutil.MongoDBScheme) {
		uri = fmt.Sprintf("%v%v", bsonutil.MongoDBScheme, uri)
	}

	cs, err := connstring.Parse(uri)
	if err != nil {
		return nil, err
	}

	if err = validateConnString(cs); err != nil {
		return nil, err
	}

	clusterOpts := []cluster.Option{
		// before WithConnString makes these
		// the defaults...
		cluster.WithServerOptions(
			server.WithMaxConnections(0),       // no upper limit per host
			server.WithMaxIdleConnections(100), // pool 100 connections per host
			server.WithConnectionOptions(
				conn.WithAppName("mongosqld"),
				conn.WithLifeTimeout(0),
				conn.WithIdleTimeout(0),
			),
		),
		cluster.WithConnString(cs),
	}

	if cfg.MongoDB.Net.SSL.Enabled {
		var dialer func(ctx context.Context, dialer *net.Dialer,
			network, address string) (net.Conn, error)
		dialer, err = ssl.SqldDialer(cfg)
		if err != nil {
			return nil, err
		}
		clusterOpts = append(clusterOpts,
			cluster.WithMoreServerOptions(
				server.WithMoreConnectionOptions(
					conn.WithDialer(dialer),
				),
			),
		)
	}

	rp, err := GetReadPreference(cs)
	if err != nil {
		return nil, err
	}

	c, err := cluster.New(clusterOpts...)
	if err != nil {
		return nil, err
	}

	sp := &SessionProvider{
		auth:           cfg.Security.Enabled,
		rp:             rp,
		c:              c,
		connectTimeout: GetConnectTimeout(cs),
		numConns:       cfg.MongoDB.Net.NumConnectionsPerSession,
	}

	if cfg.MongoDB.Net.Auth.Username != "" {
		switch cfg.MongoDB.Net.Auth.Mechanism {
		case "SCRAM-SHA-1", "SCRAM-SHA-256", "PLAIN":
			sp.adminAuthenticator = &CleartextSessionAuthenticator{
				Source:    cfg.MongoDB.Net.Auth.Source,
				Mechanism: cfg.MongoDB.Net.Auth.Mechanism,
				Username:  cfg.MongoDB.Net.Auth.Username,
				Password:  cfg.MongoDB.Net.Auth.Password,
			}
		case "GSSAPI":
			authenticator, err := newAdminSessionGSSAPIAuthenticator(cfg.MongoDB.Net.Auth)
			if err != nil {
				return nil, err
			}
			sp.adminAuthenticator = authenticator
		}
	}

	return sp, nil
}

// SessionProvider handles creating sessions.
type SessionProvider struct {
	auth           bool
	rp             *readpref.ReadPref
	c              *cluster.Cluster
	connectTimeout time.Duration
	numConns       int

	adminAuthenticator SessionAuthenticator
}

// Close closes the session provider.
func (sp *SessionProvider) Close() {
	_ = sp.c.Close()
}

// AuthenticatedAdminSessionPrimary gets a new session used for handling administration tasks
// that require a primary and need to be authenticated separately from a client.
func (sp *SessionProvider) AuthenticatedAdminSessionPrimary(ctx context.Context) (*Session, error) {
	session, err := sp.session(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	if sp.adminAuthenticator != nil {
		err = session.Login(ctx, sp.adminAuthenticator)
		if err != nil {
			_ = session.Close()
			return nil, err
		}
	}

	return session, nil
}

// AuthenticatedAdminSession gets a new session used for handling tasks which
// require authentication separately from a client. This session honors the
// read preference specified when starting up mongosqld.
func (sp *SessionProvider) AuthenticatedAdminSession(ctx context.Context) (*Session, error) {
	session, err := sp.session(ctx, sp.rp)
	if err != nil {
		return nil, err
	}

	if sp.adminAuthenticator != nil {
		err = session.Login(ctx, sp.adminAuthenticator)
		if err != nil {
			_ = session.Close()
			return nil, err
		}
	}

	return session, nil
}

// Session gets a new session.
func (sp *SessionProvider) Session(ctx context.Context) (*Session, error) {
	return sp.session(ctx, sp.rp)
}

func (sp *SessionProvider) session(ctx context.Context, rp *readpref.ReadPref) (*Session, error) {
	selector := readpref.Selector(rp)
	connectCtx, cancel := context.WithTimeout(ctx, sp.connectTimeout)
	defer cancel()
	server, err := sp.c.SelectServer(connectCtx, selector, rp)
	if err != nil {
		return nil, fmt.Errorf("no servers available: %v", err)
	}

	session := &Session{
		cluster: sp.c,
		server:  server,
		count:   sp.numConns,
	}

	// Create a provider around server.Connection in order to support
	// unauthenticating the connections.
	provider := func(ctx context.Context) (conn.Connection, error) {
		c, connErr := server.Connection(ctx)
		if connErr != nil {
			return nil, connErr
		}

		if sp.auth {
			// we need to ensure we logout any connections that were
			// authenticated before returning them to the underlying pool,
			// otherwise we'll get privilege escalation.
			c = &autoLogoutConnection{
				s:          session,
				Connection: c,
			}
		}

		session.clientAddresses = append(session.clientAddresses, c.LocalAddr().String())

		return c, nil
	}

	// The pool keeps the connections checked out of the underlying pool until
	// the session is closed.
	if session.pool, err = newSessionConnPool(ctx, provider, sp.numConns); err != nil {
		return nil, err
	}

	selectedServer := &ops.SelectedServer{
		Server:      session,
		ClusterKind: sp.c.Model().Kind,
		ReadPref:    sp.rp,
	}

	session.selectedServer = selectedServer
	return session, nil
}

type autoLogoutConnection struct {
	conn.Connection
	s *Session
}

func (c *autoLogoutConnection) Close() error {
	if c.Alive() && c.s.authSource != "" {
		logoutRequest := msg.NewCommand(
			msg.NextRequestID(),
			c.s.authSource,
			true, bsonutil.NewD(bsonutil.NewDocElem("logout", 1)),
		)

		newM := bsonutil.NewM()
		if err := conn.ExecuteCommand(context.Background(), c, logoutRequest, &newM); err != nil {
			c.MarkDead()
		}
	}

	return c.Connection.Close()
}

// GetConnectTimeout returns the connection string's ConnectTimeout.
func GetConnectTimeout(cs connstring.ConnString) time.Duration {
	if cs.ConnectTimeout == 0 {
		return 5000 * time.Millisecond
	}

	return cs.ConnectTimeout
}

// GetReadPreference returns the connection string's ReadPreference.
func GetReadPreference(cs connstring.ConnString) (*readpref.ReadPref, error) {
	var err error
	mode := readpref.PrimaryMode
	if cs.ReadPreference != "" {
		mode, err = readpref.ModeFromString(cs.ReadPreference)
		if err != nil {
			return nil, err
		}
	}

	if len(cs.ReadPreferenceTagSets) > 0 {
		tagSets := model.NewTagSetsFromMaps(cs.ReadPreferenceTagSets)
		return readpref.New(mode, readpref.WithTagSets(tagSets...))
	}

	return readpref.New(mode)
}

func validateConnString(cs connstring.ConnString) error {
	if cs.Username != "" || cs.PasswordSet ||
		cs.AuthSource != "" || cs.AuthMechanism != "" ||
		len(cs.AuthMechanismProperties) != 0 {

		return fmt.Errorf("--mongo-uri may not contain any authentication information")
	}
	if cs.Database != "" {
		return fmt.Errorf("--mongo-uri may not contain database name")
	}

	return nil
}
