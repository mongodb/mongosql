// Package client implements generic connection to MongoDB, and contains
// subpackages for specific methods of connection.
package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/client/openssl"
	"github.com/10gen/sqlproxy/client/plain"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/options"
	"github.com/10gen/sqlproxy/password"

	"github.com/10gen/mongo-go-driver/cluster"
	"github.com/10gen/mongo-go-driver/conn"
	"github.com/10gen/mongo-go-driver/connstring"
	"github.com/10gen/mongo-go-driver/readpref"
	"github.com/10gen/mongo-go-driver/server"
)

type SessionProvider struct {
	connector      DBConnector
	readPreference *readpref.ReadPref
	monitor        *cluster.Monitor
}

func (sp *SessionProvider) GetSession(ctx context.Context) (*mongodb.Session, error) {
	session, err := sp.connector.GetNewSession(ctx, sp.monitor, sp.readPreference)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func NewDrdlSessionProvider(opts options.DrdlOptions) (*SessionProvider, error) {
	provider := &SessionProvider{}

	if opts.DrdlAuth.ShouldAskForPassword() {
		opts.DrdlAuth.Password = password.Prompt()
	}

	provider.connector = getConnector(opts)

	dialInfo, err := mongodb.ParseDrdlOptions(opts)
	if err != nil {
		return nil, err
	}

	err = provider.connector.ConfigureDrdl(opts, dialInfo)
	if err != nil {
		return nil, err
	}

	provider.readPreference, err = getReadPreference(dialInfo.ConnString)
	if err != nil {
		return nil, err
	}

	clusterOpts := []cluster.Option{
		cluster.WithConnString(dialInfo.ConnString),
	}
	if dialInfo.NetDialer != nil {
		clusterOpts = append(clusterOpts, cluster.WithMoreServerOptions(
			server.WithMoreConnectionOptions(
				conn.WithDialer(dialInfo.NetDialer),
			),
		))
	}

	provider.monitor, err = cluster.StartMonitor(clusterOpts...)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

func NewSqldSessionProvider(opts options.SqldOptions) (*SessionProvider, error) {
	provider := &SessionProvider{}

	provider.connector = getConnector(opts)

	dialInfo, err := mongodb.ParseSqldOptions(opts)
	if err != nil {
		return nil, err
	}

	err = provider.connector.ConfigureSqld(opts, dialInfo)
	if err != nil {
		return nil, err
	}

	provider.readPreference, err = getReadPreference(dialInfo.ConnString)
	if err != nil {
		return nil, err
	}

	clusterOpts := []cluster.Option{
		cluster.WithConnString(dialInfo.ConnString),
	}
	if dialInfo.NetDialer != nil {
		clusterOpts = append(clusterOpts, cluster.WithMoreServerOptions(
			server.WithMoreConnectionOptions(
				conn.WithDialer(dialInfo.NetDialer),
			),
		))
	}

	provider.monitor, err = cluster.StartMonitor(clusterOpts...)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

// getConnector returns a DBConnection appropriate for
// the given options.
func getConnector(opts options.Options) DBConnector {
	if opts.UseSSL() {
		return &openssl.SSLDBConnector{}
	}
	return &plain.PlainDBConnector{}
}

// getReadPreference returns a read preference from the given connection string.
func getReadPreference(mongoURI connstring.ConnString) (*readpref.ReadPref, error) {
	readPreference := readpref.Primary()
	mode := readpref.PrimaryMode

	// check if a read preference is specified (default to
	// using the primary)
	rp, ok := mongoURI.UnknownOptions["readpreference"]
	if ok {
		parseErr := fmt.Errorf("invalid read preference specified: %v", rp)
		if len(rp) != 1 {
			return nil, parseErr
		}
		var err error
		mode, err = readpref.ModeFromString(rp[0])
		if err != nil {
			return nil, parseErr
		}
	}

	tagsets, ok := mongoURI.UnknownOptions["readpreferencetags"]
	if ok {
		var allTags []string

		for _, tagset := range tagsets {
			for _, tag := range strings.Split(tagset, ",") {
				allTags = append(allTags, strings.Split(tag, ":")...)
			}
		}

		return readpref.New(mode, readpref.WithTags(allTags...))
	}

	return readPreference, nil
}
