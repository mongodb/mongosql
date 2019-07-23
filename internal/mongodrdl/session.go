package mongodrdl

import (
	"context"
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/password"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/mongodb/ssl"

	"github.com/10gen/openssl"

	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"go.mongodb.org/mongo-driver/x/mongo/driver/topology"
)

const (
	numDrdlConnsPerSession = 2
)

func createDrdlSSLContext(opts DrdlOptions) (*openssl.Ctx, error) {
	var ctx *openssl.Ctx
	var err error

	if opts.UseFIPSMode() {
		if ssl.FipsModeSetter == nil {
			return nil, fmt.Errorf("configured to use FIPS mode, but no FIPS mode setter available")
		}
		if err = ssl.FipsModeSetter(true); err != nil {
			return nil, err
		}
	}

	if ctx, err = openssl.NewCtxWithVersion(openssl.AnyVersion); err != nil {
		return nil, fmt.Errorf("failure creating new openssl context with "+
			"NewCtxWithVersion(AnyVersion): %v", err)
	}

	ssl.SetMinimumTLSProtocolVersion(opts.MinimumTLSVersion, ctx)

	// HIGH - Enable strong ciphers
	// !EXPORT - Disable export ciphers (40/56 bit)
	// !aNULL - Disable anonymous auth ciphers
	// @STRENGTH - Sort ciphers based on strength
	if err = ctx.SetCipherList("HIGH:!EXPORT:!aNULL@STRENGTH"); err != nil {
		return nil, err
	}

	// add the PEM key file with the cert and private key, if specified
	if opts.SSLPEMKeyFile != "" {
		if err = ctx.UseCertificateChainFile(opts.SSLPEMKeyFile); err != nil {
			return nil, fmt.Errorf("UseCertificateChainFile: %v", err)
		}
		if opts.SSLPEMKeyPassword != "" {
			if err = ctx.UsePrivateKeyFileWithPassword(
				opts.SSLPEMKeyFile, openssl.FiletypePEM, opts.SSLPEMKeyPassword); err != nil {
				return nil, fmt.Errorf("UsePrivateKeyFile: %v", err)
			}
		} else {
			if err = ctx.UsePrivateKeyFile(opts.SSLPEMKeyFile, openssl.FiletypePEM); err != nil {
				return nil, fmt.Errorf("UsePrivateKeyFile: %v", err)
			}
		}
		// Verify that the certificate and the key go together.
		if err = ctx.CheckPrivateKey(); err != nil {
			return nil, fmt.Errorf("CheckPrivateKey: %v", err)
		}
	}

	// If renegotiation is needed, don't return from recv() or send() until it's successful.
	// Note: this is for blocking sockets only.
	ctx.SetMode(openssl.AutoRetry)

	// Disable session caching (see SERVER-10261)
	ctx.SetSessionCacheMode(openssl.SessionCacheOff)

	if opts.SSLCAFile != "" {
		var calist *openssl.StackOfX509Name
		calist, err = openssl.LoadClientCAFile(opts.SSLCAFile)
		if err != nil {
			return nil, fmt.Errorf("LoadClientCAFile: %v", err)
		}
		ctx.SetClientCAList(calist)

		if err = ctx.LoadVerifyLocations(opts.SSLCAFile, ""); err != nil {
			return nil, fmt.Errorf("LoadVerifyLocations: %v", err)
		}

		var verifyOption openssl.VerifyOptions
		if opts.SSLAllowInvalidCert {
			verifyOption = openssl.VerifyNone
		} else {
			verifyOption = openssl.VerifyPeer
		}
		ctx.SetVerify(verifyOption, nil)
	}

	if opts.SSLCRLFile != "" {
		store := ctx.GetCertificateStore()
		if err = store.SetFlags(openssl.CRLCheck); err != nil {
			return nil, err
		}
		lookup, err := store.AddLookup(openssl.X509LookupFile())
		if err != nil {
			return nil, fmt.Errorf("AddLookup(X509LookupFile()): %v", err)
		}
		if err = lookup.LoadCRLFile(opts.SSLCRLFile); err != nil {
			return nil, err
		}
	}

	return ctx, nil
}

// newDrdlSessionProvider creates a new session provider for mongodrdl using the supplied
// DRDL configuration options.
func newDrdlSessionProvider(opts DrdlOptions) (*mongodb.SessionProvider, error) {
	if opts.DrdlAuth.ShouldAskForPassword() {
		opts.DrdlAuth.Password = password.Prompt()
	}

	cs, err := parseDrdlOptions(opts)
	if err != nil {
		return nil, err
	}

	topologyOpts := []topology.Option{
		// Doing this before WithConnString makes these the defaults
		topology.WithServerOptions(
			func(opts ...topology.ServerOption) []topology.ServerOption {
				return append(opts, topology.WithConnectionOptions(
					func(opts ...topology.ConnectionOption) []topology.ConnectionOption {
						return append(opts, topology.WithAppName(func(string) string { return "mongodrdl" }))
					},
				))
			},
		),
		topology.WithConnString(func(connstring.ConnString) connstring.ConnString {
			return cs
		}),
	}

	if opts.UseSSL() {

		var dialer topology.Dialer

		dialer, err = drdlDialer(opts)
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

	readPref, err := mongodb.GetReadPreference(cs)
	if err != nil {
		return nil, err
	}

	t, err := topology.New(topologyOpts...)
	if err != nil {
		return nil, err
	}

	// Open the topology
	err = t.Connect()
	if err != nil {
		return nil, err
	}

	timeout := mongodb.GetConnectTimeout(cs)

	return mongodb.NewDrdlSessionProvider(readPref, t, timeout, numDrdlConnsPerSession), nil
}

func parseDrdlOptions(opts DrdlOptions) (connstring.ConnString, error) {
	uri := opts.Host

	if uri == "" {
		uri = bsonutil.DefaultMongoDBURI
	}

	if !strings.HasPrefix(uri, bsonutil.MongoDBScheme) {
		uri = fmt.Sprintf("%v%v", bsonutil.MongoDBScheme, uri)
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

	if cs.AuthMechanism == "GSSAPI" {
		cs.AuthMechanismProperties = map[string]string{}
		if opts.DrdlKerberos.Service != "" {
			cs.AuthMechanismProperties["SERVICE_NAME"] = opts.DrdlKerberos.Service
		}
		if opts.DrdlKerberos.ServiceHost != "" {
			cs.AuthMechanismProperties["SERVICE_HOST"] = opts.DrdlKerberos.ServiceHost
		}
	}

	if s := opts.GetAuthenticationDatabase(); s != "" {
		cs.AuthSource = s
	}

	return cs, nil
}

// getSession returns a mongodb.Session with the connection options specified
// by the provided DrdlOptions.
func getSession(ctx context.Context, opts DrdlOptions) (*mongodb.Session, error) {
	sp, err := newDrdlSessionProvider(opts)
	if err != nil {
		return nil, err
	}

	session, err := sp.Session(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't create session: %v", err)
	}

	return session, nil
}

// drdlDialer creates a mongodrdl dialer.
func drdlDialer(opts DrdlOptions) (topology.DialerFunc, error) {
	sslCtx, err := createDrdlSSLContext(opts)
	if err != nil {
		return nil, err
	}
	var flags openssl.DialFlags

	if opts.SSLAllowInvalidCert || opts.SSLAllowInvalidHost || opts.SSLCAFile == "" {
		flags = openssl.InsecureSkipHostVerification
	}

	return ssl.Dialer(sslCtx, flags), nil
}
