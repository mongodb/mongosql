package mongodb

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/10gen/sqlproxy/options"

	"github.com/10gen/mongo-go-driver/cluster"
	"github.com/10gen/mongo-go-driver/conn"
	"github.com/10gen/mongo-go-driver/connstring"
	"github.com/10gen/mongo-go-driver/ops"
	"github.com/10gen/mongo-go-driver/readpref"
)

const (
	mongoDBScheme = "mongodb://"
)

// DialInfo holds options for establishing a session with a MongoDB cluster.
type DialInfo struct {
	conn.EndpointDialer
	connstring.ConnString
}

// Dial establishes a new session to the cluster
// using the provided monitor choose a server
// that meets the read preference.
func (info *DialInfo) Dial(ctx context.Context, monitor *cluster.Monitor,
	readPreference *readpref.ReadPref) (*Session, error) {
	opts := []conn.Option{
		conn.WithAppName(info.AppName),
		conn.WithEndpointDialer(info.EndpointDialer),
	}

	selector := cluster.ReadPrefSelector(readPreference)

	optionName := "connectTimeoutMS"
	option, ok := info.getUnknownOption(optionName)
	if !ok {
		panic(fmt.Sprintf("%v is unset", optionName))
	}

	intValue, err := strconv.Atoi(option[0])
	if err != nil {
		return nil, fmt.Errorf("invalid %v value: %v", optionName, option[0])
	}

	connectTimeout := time.Duration(intValue) * time.Millisecond
	connectCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()
	servers, err := cluster.SelectServers(connectCtx, monitor, selector)
	if err != nil {
		return nil, err
	}

	if len(servers) == 0 {
		return nil, fmt.Errorf("could not find any servers")
	}

	connectCtx, cancel = context.WithTimeout(ctx, connectTimeout)
	defer cancel()
	connection, err := NewDirectServerConnection(connectCtx, servers[0], opts...)
	if err != nil {
		return nil, err
	}

	selectedServer := &ops.SelectedServer{
		serverImpl{connection},
		readPreference,
	}

	session := &Session{
		appName:    info.AppName,
		connection: connection,
		ctx:        ctx,
		dialer:     info.EndpointDialer,
		server:     selectedServer,
	}

	return session, nil
}

func (info *DialInfo) getUnknownOption(option string) ([]string, bool) {
	values, ok := info.ConnString.UnknownOptions[strings.ToLower(option)]
	return values, ok
}

// setDefaultTimeouts sets the default connection and socket
// timeouts if not specified via the URI options.
func (info *DialInfo) setDefaultTimeouts() {
	if info.UnknownOptions == nil {
		info.UnknownOptions = map[string][]string{}
	}

	// By default, we'll enforce a 5 second connection timeout
	if _, ok := info.UnknownOptions["connecttimeoutms"]; !ok {
		info.UnknownOptions["connecttimeoutms"] = []string{"5000"}
	}

	if _, ok := info.UnknownOptions["sockettimeoutms"]; !ok {
		info.UnknownOptions["sockettimeoutms"] = []string{"0"}
	}
}

func ParseDrdlOptions(opts options.DrdlOptions) (*DialInfo, error) {

	if strings.HasPrefix(opts.Host, mongoDBScheme) {
		opts.Host = opts.Host[len(mongoDBScheme):]
	}

	hosts, replicaSetName := parseMongoDRDLHost(opts.Host)

	if opts.Port != "" {
		for i, host := range hosts {
			if strings.Index(host, ":") == -1 {
				host = fmt.Sprintf("%v:%v", host, opts.Port)
			}
			hosts[i] = host
		}
	}

	addr := strings.Join(hosts, ",")

	info, err := parse(addr)
	if err != nil {
		return nil, err
	}

	if replicaSetName != "" {
		info.ReplicaSet = replicaSetName
	} else {
		info.Connect = connstring.SingleConnect
	}

	if opts.DrdlAuth.Username == "" {
		info.Username = opts.DrdlAuth.Username
	}

	if opts.DrdlAuth.Password != "" {
		info.Password = opts.DrdlAuth.Password
		info.PasswordSet = true
	}

	if opts.DrdlAuth.Mechanism != "" {
		info.AuthMechanism = opts.DrdlAuth.Mechanism
	}

	if authSource := opts.GetAuthenticationDatabase(); authSource != "" {
		info.AuthSource = authSource
	}

	info.setDefaultTimeouts()

	if info.AppName == "" {
		info.AppName = "mongodrdl"
	}

	return info, nil
}

func ParseSqldOptions(opts options.SqldOptions) (*DialInfo, error) {
	info, err := parse(*opts.MongoURI)
	if err != nil {
		return nil, err
	}

	info.setDefaultTimeouts()

	if info.AppName == "" {
		info.AppName = "mongosqld"
	}

	return info, nil
}

// ensureScheme ensures that the uri passed in contains
// the MongoDB scheme and adds it if the uri doesn't.
func ensureScheme(uri string) string {
	if !strings.HasPrefix(uri, mongoDBScheme) {
		return fmt.Sprintf("%v%v", mongoDBScheme, uri)
	}
	return uri
}

// parse parses the MongoDB uri given and returns
// a DialInfo constructed from it.
func parse(uri string) (*DialInfo, error) {
	connectionString, err := connstring.Parse(ensureScheme(uri))
	if err != nil {
		return nil, err
	}
	return &DialInfo{ConnString: connectionString}, nil
}

// parseMongoDRDLHost extract the replica set name and the
// list of hosts from the connection string
func parseMongoDRDLHost(connString string) ([]string, string) {
	slashIndex := strings.Index(connString, "/")
	setName := ""
	if slashIndex != -1 {
		setName = connString[:slashIndex]
		if slashIndex == len(connString)-1 {
			return []string{""}, setName
		}
		connString = connString[slashIndex+1:]
	}
	return strings.Split(connString, ","), setName
}
