package mongodb

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/mongo-go-driver/mongo/connstring"
	"github.com/10gen/mongo-go-driver/mongo/model"
	"github.com/10gen/mongo-go-driver/mongo/private/cluster"
	"github.com/10gen/mongo-go-driver/mongo/private/conn"
	"github.com/10gen/mongo-go-driver/mongo/private/msg"
	"github.com/10gen/mongo-go-driver/mongo/private/ops"
	"github.com/10gen/mongo-go-driver/mongo/private/server"
	"github.com/10gen/mongo-go-driver/mongo/readpref"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/options"
	"github.com/10gen/sqlproxy/password"
	"github.com/10gen/sqlproxy/ssl"
)

const (
	mongoDBScheme          = "mongodb://"
	drdlNumConnsPerSession = 2
)

// NewDrdlSessionProvider creates a new session provider for mongodrdl.
func NewDrdlSessionProvider(opts options.DrdlOptions) (*SessionProvider, error) {
	if opts.DrdlAuth.ShouldAskForPassword() {
		opts.DrdlAuth.Password = password.Prompt()
	}

	cs, err := parseDrdlOptions(opts)
	if err != nil {
		return nil, err
	}

	clusterOpts := []cluster.Option{
		// before WithConnString makes these
		// the defaults...
		cluster.WithServerOptions(
			server.WithConnectionOptions(
				conn.WithAppName("mongodrdl"),
			),
		),
		cluster.WithConnString(cs),
	}

	if opts.UseSSL() {
		var dialer func(ctx context.Context, dialer *net.Dialer,
			network, address string) (net.Conn, error)
		dialer, err = ssl.DrdlDialer(opts)
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

	rp, err := getReadPreference(cs)
	if err != nil {
		return nil, err
	}

	c, err := cluster.New(clusterOpts...)
	if err != nil {
		return nil, err
	}

	return &SessionProvider{
		rp:             rp,
		c:              c,
		connectTimeout: getConnectTimeout(cs),
		numConns:       drdlNumConnsPerSession,
	}, nil
}

// NewSqldSessionProvider creates a new session provider for mongosql.
func NewSqldSessionProvider(cfg *config.Config) (*SessionProvider, error) {

	uri := cfg.MongoDB.Net.URI
	if !strings.HasPrefix(uri, mongoDBScheme) {
		uri = fmt.Sprintf("%v%v", mongoDBScheme, uri)
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

	rp, err := getReadPreference(cs)
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
		connectTimeout: getConnectTimeout(cs),
		numConns:       cfg.MongoDB.Net.NumConnectionsPerSession,
	}

	if cfg.MongoDB.Net.Auth.Username != "" {
		sp.adminAuthenticator = &CleartextSessionAuthenticator{
			Source:    cfg.MongoDB.Net.Auth.Source,
			Mechanism: cfg.MongoDB.Net.Auth.Mechanism,
			Username:  cfg.MongoDB.Net.Auth.Username,
			Password:  cfg.MongoDB.Net.Auth.Password,
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

// AdminSession gets a new session used for handling administration tasks that
// require a primary and need to be authenticated separately from a client.
func (sp *SessionProvider) AdminSession(ctx context.Context) (*Session, error) {
	session, err := sp.session(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	if sp.adminAuthenticator != nil {
		err = session.Login(sp.adminAuthenticator)
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
		ctx:     ctx,
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
			true,
			bson.D{{Name: "logout", Value: 1}},
		)

		if err := conn.ExecuteCommand(c.s.ctx, c, logoutRequest, &bson.M{}); err != nil {
			c.MarkDead()
		}
	}

	return c.Connection.Close()
}

func parseDrdlOptions(opts options.DrdlOptions) (connstring.ConnString, error) {

	uri := opts.Host

	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	if !strings.HasPrefix(uri, mongoDBScheme) {
		uri = fmt.Sprintf("%v%v", mongoDBScheme, uri)
	}

	cs, err := connstring.Parse(uri)
	if err != nil {
		return cs, err
	}

	if cs.ReplicaSet == "" {
		cs.Connect = connstring.SingleConnect
	}

	cs.Username = opts.DrdlAuth.Username

	if opts.DrdlAuth.Password != "" {
		cs.Password = opts.DrdlAuth.Password
		cs.PasswordSet = true
	}

	cs.AuthMechanism = opts.DrdlAuth.Mechanism

	if s := opts.GetAuthenticationDatabase(); s != "" {
		cs.AuthSource = s
	}

	return cs, nil
}

func getConnectTimeout(cs connstring.ConnString) time.Duration {
	if cs.ConnectTimeout == 0 {
		return 5000 * time.Millisecond
	}

	return cs.ConnectTimeout
}

func getReadPreference(cs connstring.ConnString) (*readpref.ReadPref, error) {
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
