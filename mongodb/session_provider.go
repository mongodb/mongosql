package mongodb

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/10gen/mongo-go-driver/cluster"
	"github.com/10gen/mongo-go-driver/conn"
	"github.com/10gen/mongo-go-driver/connstring"
	"github.com/10gen/mongo-go-driver/model"
	"github.com/10gen/mongo-go-driver/msg"
	"github.com/10gen/mongo-go-driver/ops"
	"github.com/10gen/mongo-go-driver/readpref"
	"github.com/10gen/mongo-go-driver/server"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/options"
	"github.com/10gen/sqlproxy/password"
	"github.com/10gen/sqlproxy/ssl"
)

const (
	mongoDBScheme = "mongodb://"
)

// NewDrdlSessionProvider creates a new session provider for mongodrdl.
func NewDrdlSessionProvider(opts options.DrdlOptions) (*SessionProvider, error) {
	if opts.DrdlAuth.ShouldAskForPassword() {
		opts.DrdlAuth.Password = password.Prompt()
	}

	cs := parseDrdlOptions(opts)

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
		dialer, err := ssl.DrdlDialer(opts)
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

	if err := validateConnString(cs); err != nil {
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
			),
		),
		cluster.WithConnString(cs),
	}

	if cfg.MongoDB.Net.SSL.Enabled {
		dialer, err := ssl.SqldDialer(cfg)
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
		auth:           cfg.Security.Enabled,
		rp:             rp,
		c:              c,
		connectTimeout: getConnectTimeout(cs),
	}, nil
}

// SessionProvider handles creating sessions.
type SessionProvider struct {
	auth           bool
	rp             *readpref.ReadPref
	c              *cluster.Cluster
	connectTimeout time.Duration
}

// Close closes the session provider.
func (sp *SessionProvider) Close() {
	sp.c.Close()
}

// Session gets a new session.
func (sp *SessionProvider) Session(ctx context.Context) (*Session, error) {
	selector := readpref.Selector(sp.rp)
	connectCtx, cancel := context.WithTimeout(ctx, sp.connectTimeout)
	defer cancel()
	server, err := sp.c.SelectServer(connectCtx, selector)
	if err != nil {
		return nil, fmt.Errorf("no servers available")
	}

	const nConns = 2

	session := &Session{
		ctx:     ctx,
		cluster: sp.c,
		server:  server,
		count:   nConns,
	}

	var numCreated int32
	provider := func(ctx context.Context) (conn.Connection, error) {
		// we can only create up to nConns due to authentication reasons.
		// Therefore, we are going to cap this and refuse to create any more.
		// This should cause us to error the mysql connection.
		new := atomic.AddInt32(&numCreated, 1)
		if new <= nConns {
			c, err := server.Connection(ctx)
			if err != nil {
				return nil, err
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

			return c, nil
		}

		return nil, fmt.Errorf("not enough mongodb connections available for session")
	}

	// The pool keeps the connections checked out of the underlying pool until
	// the session is closed.
	session.pool = conn.NewPool(nConns, provider)
	// This ensures that only nConns connections are allowed per session.
	session.provider = conn.CappedProvider(nConns, session.pool.Get)

	selectedServer := &ops.SelectedServer{
		Server:   session,
		ReadPref: sp.rp,
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

func parseDrdlOptions(opts options.DrdlOptions) connstring.ConnString {
	cs := connstring.ConnString{}

	if strings.HasPrefix(opts.Host, mongoDBScheme) {
		opts.Host = opts.Host[len(mongoDBScheme):]
	}

	hosts, replicaSetName := parseDrdlHost(opts.Host)
	if opts.Port != "" {
		for i, host := range hosts {
			if strings.Index(host, ":") == -1 {
				host = fmt.Sprintf("%v:%v", host, opts.Port)
			}
			hosts[i] = host
		}
	}
	cs.Hosts = hosts
	cs.ReplicaSet = replicaSetName
	if cs.ReplicaSet == "" {
		cs.Connect = connstring.SingleConnect
	}

	if opts.DrdlAuth.Username != "" {
		cs.Username = opts.DrdlAuth.Username
	}

	if opts.DrdlAuth.Password != "" {
		cs.Password = opts.DrdlAuth.Password
		cs.PasswordSet = true
	}

	if opts.DrdlAuth.Mechanism != "" {
		cs.AuthMechanism = opts.DrdlAuth.Mechanism
	}

	if authSource := opts.GetAuthenticationDatabase(); authSource != "" {
		cs.AuthSource = authSource
	}

	return cs
}

func parseDrdlHost(host string) ([]string, string) {
	slashIndex := strings.Index(host, "/")
	setName := ""
	if slashIndex != -1 {
		setName = host[:slashIndex]
		if slashIndex == len(host)-1 {
			return []string{""}, setName
		}
		host = host[slashIndex+1:]
	}
	return strings.Split(host, ","), setName
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
