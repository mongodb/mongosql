package mongodb

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/procutil"
	"github.com/10gen/sqlproxy/mongodb/ssl"

	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/tag"
	"go.mongodb.org/mongo-driver/x/mongo/driver"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"go.mongodb.org/mongo-driver/x/mongo/driver/description"
	"go.mongodb.org/mongo-driver/x/mongo/driver/topology"
)

// NewDrdlSessionProvider creates a new session provider for mongodrdl.
func NewDrdlSessionProvider(rp *readpref.ReadPref, t *topology.Topology, timeout time.Duration, numConns int) *SessionProvider {
	return &SessionProvider{
		rp:             rp,
		t:              t,
		connectTimeout: timeout,
		numConns:       numConns,
	}
}

// NewSqldSessionProvider creates a new session provider for mongosql.
func NewSqldSessionProvider(cfg *config.Config) (*SessionProvider, error) {
	uri := cfg.MongoDB.Net.URI
	if !strings.HasPrefix(uri, procutil.MongoDBScheme) &&
		!strings.HasPrefix(uri, procutil.MongoDBSRVScheme) {
		uri = fmt.Sprintf("%v://%v", procutil.MongoDBScheme, uri)
	}

	cs, err := connstring.Parse(uri)
	if err != nil {
		return nil, err
	}

	if err = validateConnString(cs); err != nil {
		return nil, err
	}

	// If no compressors are specified in the connection string,
	// we default them here to zlib,snappy. We add these to the
	// connection string (as opposed to adding them via default
	// options below) because topology.WithConnString overwrites
	// the Compressors unconditionally.
	if len(cs.Compressors) == 0 {
		cs.Compressors = []string{"zlib", "snappy"}
	}

	topologyOpts := []topology.Option{
		// Doing this before WithConnString makes these the defaults
		topology.WithServerOptions(func(options ...topology.ServerOption) []topology.ServerOption {
			return append(options,
				topology.WithMaxConnections(func(uint16) uint16 { return 0 }),       // no upper limit per host
				topology.WithMaxIdleConnections(func(uint16) uint16 { return 100 }), // pool 100 connections per host
				topology.WithConnectionOptions(func(options ...topology.ConnectionOption) []topology.ConnectionOption {
					return append(options,
						topology.WithAppName(func(string) string { return "mongosqld" }),
						topology.WithLifeTimeout(func(time.Duration) time.Duration { return 0 }),
						topology.WithIdleTimeout(func(time.Duration) time.Duration { return 0 }),
					)
				}),
			)
		}),
		topology.WithConnString(func(connstring.ConnString) connstring.ConnString {
			return cs
		}),
	}

	if cfg.MongoDB.Net.SSL.Enabled {
		var dialer topology.Dialer
		dialer, err = ssl.SqldDialer(cfg)
		if err != nil {
			return nil, err
		}

		topologyOpts = append(topologyOpts,
			topology.WithServerOptions(
				func(opts ...topology.ServerOption) []topology.ServerOption {
					return append(opts, topology.WithConnectionOptions(
						func(opts ...topology.ConnectionOption) []topology.ConnectionOption {
							return append(opts, topology.WithDialer(func(topology.Dialer) topology.Dialer { return dialer }))
						},
					))
				},
			),
		)
	}

	rp, err := GetReadPreference(cs)
	if err != nil {
		return nil, err
	}

	t, err := topology.New(topologyOpts...)
	if err != nil {
		return nil, err
	}

	// Must call Connect() on the topology to open it.
	if err = t.Connect(); err != nil {
		return nil, err
	}

	sp := &SessionProvider{
		auth:           cfg.Security.Enabled,
		rp:             rp,
		t:              t,
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
	t              *topology.Topology
	connectTimeout time.Duration
	numConns       int

	adminAuthenticator SessionAuthenticator
}

// Close closes the session provider.
func (sp *SessionProvider) Close() {
	_ = sp.t.Disconnect(context.Background())
}

// AuthenticatedAdminSessionPrimary gets a new Session used for handling administration tasks
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

// AuthenticatedAdminSession gets a new Session used for handling tasks which
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

// Session gets a new Session.
func (sp *SessionProvider) Session(ctx context.Context) (*Session, error) {
	return sp.session(ctx, sp.rp)
}

// session creates a new Session. It uses the provided read preference to select
// a server from the SessionProvider's deployment. We use that server to get the
// connections for the Session's connection pool.
func (sp *SessionProvider) session(ctx context.Context, rp *readpref.ReadPref) (*Session, error) {
	// First, select an appropriate driver.Server from sp's topology.
	selector := description.ReadPrefSelector(rp)
	connectCtx, cancel := context.WithTimeout(ctx, sp.connectTimeout)
	defer cancel()
	server, err := sp.t.SelectServer(connectCtx, selector)
	if err != nil {
		return nil, fmt.Errorf("no servers available: %v", err)
	}

	// Create the Session
	session := &Session{
		topologyKind: sp.t.Kind(),
		numConns:     sp.numConns,
		rp:           rp,
	}

	// Create a connection provider for the connection pool. This provider
	// uses the selected server to get driver.Connections and adds support
	// for unauthenticating the connections.
	provider := func(ctx context.Context) (driver.Connection, error) {
		c, connErr := server.Connection(ctx)
		if connErr != nil {
			return nil, connErr
		}

		if sp.auth {
			// We need to ensure we logout any connections that were
			// authenticated before returning them to the underlying
			// pool, otherwise we'll get privilege escalation.
			// If the driver.Connection returned from the server is
			// not Expirable, we cannot use it for an authenticated
			// session.
			e, ok := c.(expirableConnection)
			if !ok {
				return nil, errors.New("invalid connection for use with authentication")
			}
			c = &autoLogoutConnection{
				s:                   session,
				expirableConnection: e,
			}
		}

		l, ok := c.(driver.LocalAddresser)
		if !ok {
			return nil, errors.New("unable to get connection's local address")
		}

		session.clientAddresses = append(session.clientAddresses, l.LocalAddress().String())

		return c, nil
	}

	// The pool keeps the connections checked out of the
	// underlying pool until the session is closed.
	if session.pool, err = newSessionConnPool(ctx, provider, sp.numConns); err != nil {
		return nil, err
	}

	// Finally, to support running commands against mongodb, the session needs
	// a driver.Deployment. Here, we create a SingleServerDeployment using the
	// the session itself! Session implements driver.Server by using the pool
	// to provide driver.Connections. SingleServerDeployment will always return
	// the session as the selected server, and therefore the connections will
	// always come from the pool. This is the necessary behavior because of our
	// auth requirements.
	session.deployment = driver.SingleServerDeployment{
		Server: session,
	}

	return session, nil
}

type expirableConnection interface {
	driver.Connection
	driver.Expirable
	driver.LocalAddresser
}

type autoLogoutConnection struct {
	expirableConnection
	s *Session
}

func (c *autoLogoutConnection) Close() error {
	if c.Alive() && c.s.authSource != "" {
		logoutRequest := bsonutil.NewD(
			bsonutil.NewDocElem("logout", 1),
		)

		res := bsonutil.NewD()
		if err := c.s.Run(context.Background(), c.s.authSource, logoutRequest, &res); err != nil {
			// If the logout command fails, we expire the connection. We can ignore the error
			// returned by Expire() because the method guarantees the connection will not be
			// returned to the underlying pool (at the Go driver level). Not being returned to
			// the pool ensures the authenticated connection will not be reused.
			_ = c.Expire()
		}
	}

	return c.expirableConnection.Close()
}

// CompressWireMessage handles compressing the provided wire message using the
// underlying driver.Connection if it is also a driver.Compressor.
func (c *autoLogoutConnection) CompressWireMessage(src, dst []byte) ([]byte, error) {
	if compressor, ok := c.expirableConnection.(driver.Compressor); ok {
		return compressor.CompressWireMessage(src, dst)
	}

	// Cannot compress if the underlying driver.Connection is not a driver.Compressor.
	return append(dst, src...), nil
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
		tagSets := tag.NewTagSetsFromMaps(cs.ReadPreferenceTagSets)
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
