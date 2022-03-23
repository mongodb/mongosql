package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/procutil"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/mongodb/ssl"
	"github.com/10gen/sqlproxy/schema"

	"go.mongodb.org/mongo-driver/mongo/description"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/tag"
	"go.mongodb.org/mongo-driver/x/mongo/driver"
	"go.mongodb.org/mongo-driver/x/mongo/driver/auth"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"go.mongodb.org/mongo-driver/x/mongo/driver/topology"
)

// NewDrdlSessionProvider creates a new session provider for mongodrdl.
func NewDrdlSessionProvider(rp *readpref.ReadPref, t *topology.Topology, timeout time.Duration, numConns int) *SessionProvider {
	return &SessionProvider{
		rp:                rp,
		adminConnTopology: t,
		userConnTopology:  t,
		connectTimeout:    timeout,
		numConns:          numConns,
	}
}

// NewSqldSessionProvider creates a new session provider for mongosqld.
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

	if cs.AppName == "" {
		cs.AppName = "mongosqld"
	}

	// If no compressors are specified in the connection string,
	// we default them here to zlib,snappy. We add these to the
	// connection string (as opposed to adding them via default
	// options below) because topology.WithConnString overwrites
	// the Compressors unconditionally.
	if len(cs.Compressors) == 0 {
		cs.Compressors = []string{"zlib", "snappy"}
	}

	// These topology options will be used for both the adminConnTopology and the userConnTopology.
	topologyOpts := []topology.Option{
		// Doing this before WithConnString makes these the defaults
		topology.WithServerOptions(func(options ...topology.ServerOption) []topology.ServerOption {
			return append(options,
				topology.WithMaxConnections(func(uint64) uint64 { return 0 }), // no upper limit per host
				topology.WithServerAppName(func(string) string { return "mongosqld" }),
				topology.WithConnectionOptions(func(options ...topology.ConnectionOption) []topology.ConnectionOption {
					return append(options, topology.WithIdleTimeout(func(time.Duration) time.Duration { return 0 }))
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

	userConnTopology, err := topology.New(topologyOpts...)
	if err != nil {
		return nil, err
	}

	// Must call Connect() on the topology to open it.
	if err = userConnTopology.Connect(); err != nil {
		return nil, err
	}

	// Add the handshaker option to topologyOpts for the adminConnTopology.
	if cfg.MongoDB.Net.Auth.Username != "" {
		cred := &auth.Cred{
			Username:    cfg.MongoDB.Net.Auth.Username,
			Password:    cfg.MongoDB.Net.Auth.Password,
			PasswordSet: cfg.MongoDB.Net.Auth.Password != "",
			Source:      cfg.MongoDB.Net.Auth.Source,
			Props:       make(map[string]string),
		}
		mechanism := cfg.MongoDB.Net.Auth.Mechanism

		if len(cred.Source) == 0 {
			switch strings.ToUpper(mechanism) {
			case auth.MongoDBX509, auth.GSSAPI, auth.PLAIN:
				cred.Source = "$external"
			default:
				cred.Source = "admin"
			}
		}

		if strings.ToUpper(mechanism) == auth.GSSAPI {
			cred.Props["SERVICE_NAME"] = cfg.MongoDB.Net.Auth.GSSAPIServiceName
		}

		var authenticator auth.Authenticator
		authenticator, err = auth.CreateAuthenticator(mechanism, cred)
		if err != nil {
			return nil, err
		}

		handshakeOpts := &auth.HandshakeOptions{
			AppName:       "mongosqld",
			Authenticator: authenticator,
		}
		if mechanism == "" {
			// Required for SASL mechanism negotiation during handshake
			handshakeOpts.DBUser = cred.Source + "." + cred.Username
		}

		topologyOpts = append(topologyOpts,
			topology.WithServerOptions(
				func(opts ...topology.ServerOption) []topology.ServerOption {
					return append(opts, topology.WithConnectionOptions(
						func(opts ...topology.ConnectionOption) []topology.ConnectionOption {
							return append(opts, topology.WithHandshaker(func(driver.Handshaker) driver.Handshaker {
								return auth.Handshaker(nil, handshakeOpts)
							}))
						},
					))
				},
			),
		)
	}

	adminConnTopology, err := topology.New(topologyOpts...)
	if err != nil {
		return nil, err
	}

	// Must call Connect() on the topology to open it.
	if err = adminConnTopology.Connect(); err != nil {
		return nil, err
	}

	rp, err := GetReadPreference(cs)
	if err != nil {
		return nil, err
	}

	sp := &SessionProvider{
		auth:              cfg.Security.Enabled,
		rp:                rp,
		adminConnTopology: adminConnTopology,
		userConnTopology:  userConnTopology,
		connectTimeout:    GetConnectTimeout(cs),
		numConns:          cfg.MongoDB.Net.NumConnectionsPerSession,
	}

	return sp, nil
}

// SessionProvider handles creating sessions. See mongodb/README.md for more details.
type SessionProvider struct {
	auth              bool
	rp                *readpref.ReadPref
	adminConnTopology *topology.Topology
	userConnTopology  *topology.Topology
	connectTimeout    time.Duration
	numConns          int
}

// Close closes the session provider.
func (sp *SessionProvider) Close() {
	_ = sp.adminConnTopology.Disconnect(context.Background())
	_ = sp.userConnTopology.Disconnect(context.Background())
}

// adminTopology wraps the driver.Deployment type. It is used for admin sessions
// where auth info is stored in the deployment. This is different from the
// driver.SingleServerDeployment used for user sessions, which authenticates
// driver.Connections during the MySQL handshake.
// See mongodb/README.md for more details about admin sessions.
type adminTopology struct {
	deployment driver.Deployment
	session    *Session

	// auth flag from the SessionProvider
	auth bool
}

// SelectServer implements the driver.Deployment interface. For adminTopology, it
// returns an adminServer.
func (t *adminTopology) SelectServer(ctx context.Context, ss description.ServerSelector) (driver.Server, error) {
	s, err := t.deployment.SelectServer(ctx, ss)
	if err != nil {
		return nil, err
	}

	return &adminServer{s, t.session, t.auth}, nil
}

// Kind implements the driver.Deployment interface.
func (t *adminTopology) Kind() description.TopologyKind {
	return t.deployment.Kind()
}

// adminServer wraps the driver.Server type. It is used for admin sessions
// where auth info is stored in the deployment. This is different from the
// Session (which implements driver.Server) used for user sessions, which
// pools its own authenticated driver.Connections.
// See mongodb/README.md for more details about admin sessions.
type adminServer struct {
	server  driver.Server
	session *Session

	// auth flag from the SessionProvider
	auth bool
}

// Connection implements the driver.Server interface. For adminServer, it returns
// an autoLogoutConnection if auth is enabled.
func (s *adminServer) Connection(ctx context.Context) (driver.Connection, error) {
	c, err := s.server.Connection(ctx)
	if err != nil {
		return nil, err
	}

	if s.auth {
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
			s:                   s.session,
			expirableConnection: e,
		}
	}

	return c, nil
}

// AuthenticatedAdminSessionPrimary gets a new Session used for handling administration tasks
// that require a primary and need to be authenticated separately from a client.
func (sp *SessionProvider) AuthenticatedAdminSessionPrimary() (*Session, error) {
	return sp.adminSession(readpref.Primary())
}

// AuthenticatedAdminSession gets a new Session used for handling tasks which
// require authentication separately from a client. This session honors the
// read preference specified when starting up mongosqld.
func (sp *SessionProvider) AuthenticatedAdminSession() (*Session, error) {
	return sp.adminSession(sp.rp)
}

// adminSession creates a new Session to be used for admin connections to MongoDB.
// The Session will use the provided read preference when selecting a server to
// execute commands.
func (sp *SessionProvider) adminSession(rp *readpref.ReadPref) (*Session, error) {
	session := &Session{
		topologyKind:   sp.adminConnTopology.Kind(),
		ReadPreference: rp,
	}

	session.Deployment = &adminTopology{
		deployment: sp.adminConnTopology,
		session:    session,
	}

	ctx, cancel := context.WithTimeout(context.Background(), sp.connectTimeout)
	defer cancel()
	// ping to check credentials
	err := session.Run(ctx, "admin", bsonutil.NewD(bsonutil.NewDocElem("ping", 1)), &struct{}{})
	if err != nil {
		return nil, err
	}

	return session, nil
}

// Session creates a new Session. It uses the SessionProvider's read preference
// to select a server from the SessionProvider's deployment. We use that server
// to get the connections for the Session's connection pool.
func (sp *SessionProvider) Session(ctx context.Context) (*Session, error) {
	// First, select an appropriate driver.Server from sp's topology.
	selector := description.ReadPrefSelector(sp.rp)
	connectCtx, cancel := context.WithTimeout(ctx, sp.connectTimeout)
	defer cancel()
	server, err := sp.userConnTopology.SelectServer(connectCtx, selector)
	if err != nil {
		return nil, fmt.Errorf("no servers available: %v", err)
	}

	// Create the Session
	session := &Session{
		topologyKind:   sp.userConnTopology.Kind(),
		NumConns:       sp.numConns,
		ReadPreference: sp.rp,
	}

	// Create a connection provider for the connection pool. This provider
	// uses the selected server to get driver.Connections and adds support
	// for unauthenticating the connections.
	provider := func(ctx context.Context) (driver.Connection, error) {
		c, connErr := server.Connection(ctx)
		if connErr != nil {
			return nil, connErr
		}

		l, ok := c.(driver.LocalAddresser)
		if !ok {
			return nil, errors.New("unable to get connection's local address")
		}

		session.ClientAddresses = append(session.ClientAddresses, l.LocalAddress().String())

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

		return c, nil
	}

	// The pool keeps the connections checked out of the
	// underlying pool until the session is closed.
	if session.Pool, err = NewSessionConnPool(ctx, provider, sp.numConns); err != nil {
		return nil, err
	}

	// Finally, to support running commands against mongodb, the session needs
	// a driver.Deployment. Here, we create a SingleServerDeployment using the
	// the session itself! Session implements driver.Server by using the pool
	// to provide driver.Connections. SingleServerDeployment will always return
	// the session as the selected server, and therefore the connections will
	// always come from the pool. This is the necessary behavior because of our
	// auth requirements.
	session.Deployment = driver.SingleServerDeployment{
		Server: session,
	}

	return session, nil
}

type expirableConnection interface {
	driver.Connection
	driver.Expirable
}

type autoLogoutConnection struct {
	expirableConnection
	s *Session
}

func (c *autoLogoutConnection) Close() error {
	if c.Alive() && c.s.AuthSource != "" {
		logoutRequest := bsonutil.NewD(
			bsonutil.NewDocElem("logout", 1),
		)

		res := bsonutil.NewD()
		if err := c.s.Run(context.Background(), c.s.AuthSource, logoutRequest, &res); err != nil {
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
	if cs.HasAuthParameters() {
		return fmt.Errorf("--mongo-uri may not contain any authentication information")
	}
	if cs.Database != "" {
		return fmt.Errorf("--mongo-uri may not contain database name")
	}

	return nil
}

// LoadInfo looks up information from MongoDB.
func LoadInfo(ctx context.Context, logger log.Logger, sp *SessionProvider, userSession *Session,
	schema *schema.Schema, config *config.Config) (i *mongodb.Info, e error) {

	defer func() {
		if r := recover(); r != nil {
			i = nil
			logger.Warnf(log.Admin, "MongoDB information access session possibly closed: %v", r)
			// Make sure we return the error. Without the next line go just returns nil, nil,
			// which breaks the contract we want (namely that if the returned info is nil,
			// there should be an error).
			switch x := r.(type) {
			case string:
				e = errors.New(x)
			case error:
				e = x
			default:
				e = errors.New("Unknown panic")
			}
		}
	}()

	adminSession, err := sp.AuthenticatedAdminSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create admin session for loading cluster information: %v", err)
	}
	defer func() {
		_ = adminSession.Close()
	}()

	var version *mongodb.VersionInfo
	version, err = userSession.Version(ctx)
	if err != nil {
		return nil, err
	}

	i = &mongodb.Info{
		Config:       config,
		Databases:    createDatabasesFromSchema(schema),
		GitVersion:   version.GitVersion,
		Version:      version.Version,
		VersionArray: version.VersionArray,
	}

	if config.Security.Enabled {
		err = i.LoadAuthInfo(ctx, logger, userSession, config.Schema.Stored.Source)
		if err != nil {
			return nil, err
		}
	} else {
		i.SetAllPrivileges(mongodb.AllPrivileges)
	}

	i.LoadMetadata(ctx, logger, adminSession)

	return i, nil
}

func createDatabasesFromSchema(config *schema.Schema) map[mongodb.DatabaseName]*mongodb.DatabaseInfo {
	dbInfos := make(map[mongodb.DatabaseName]*mongodb.DatabaseInfo, len(config.Databases()))
	for _, dbSchema := range config.Databases() {
		dbName := strings.ToLower(dbSchema.Name())
		dbInfo := &mongodb.DatabaseInfo{
			CaseSensitiveName: dbSchema.Name(),
			Collections:       make(map[mongodb.CollectionName]*mongodb.CollectionInfo),
		}

		dbInfos[mongodb.DatabaseName(dbName)] = dbInfo

		for _, table := range dbSchema.Tables() {
			name := mongodb.CollectionName(table.MongoName())
			if _, ok := dbInfo.Collections[name]; ok {
				// Because multiple tables can be mapped to the same collection,
				// we can skip collections we've already included.
				continue
			}

			dbInfo.Collections[name] = &mongodb.CollectionInfo{
				Name: name,
			}
		}
	}
	return dbInfos
}
