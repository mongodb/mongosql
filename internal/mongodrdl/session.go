package mongodrdl

import (
	"fmt"

	"github.com/10gen/sqlproxy/mongodb/provider"
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
func newDrdlSessionProvider(opts DrdlOptions) (*provider.SessionProvider, error) {
	var err error

	// Ensure the password is set. If we need to ask the user, this method will
	// prompt for password input and set it appropriately.
	opts.EnsurePassword()
	cs, err := opts.ConnString()
	if err != nil {
		return nil, err
	}

	topologyOpts := []topology.Option{
		// Doing this before WithConnString makes these the defaults
		topology.WithServerOptions(
			func(opts ...topology.ServerOption) []topology.ServerOption {
				return append(opts,
					topology.WithServerAppName(func(string) string { return "mongodrdl" }),
					topology.WithConnectionOptions(func(opts ...topology.ConnectionOption) []topology.ConnectionOption {
						return append(opts, topology.WithConnectionAppName(func(string) string { return "mongodrdl" }))
					}),
				)
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

	readPref, err := provider.GetReadPreference(cs)
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

	timeout := provider.GetConnectTimeout(cs)

	return provider.NewDrdlSessionProvider(readPref, t, timeout, numDrdlConnsPerSession), nil
}

// drdlDialer creates a mongodrdl dialer.
func drdlDialer(opts DrdlOptions) (topology.DialerFunc, error) {
	sslCtx, err := createDrdlSSLContext(opts)
	if err != nil {
		return nil, err
	}
	var flags openssl.DialFlags

	if opts.DrdlSSL.SSLAllowInvalidCert || opts.DrdlSSL.SSLAllowInvalidHost || opts.DrdlSSL.SSLCAFile == "" {
		flags = openssl.InsecureSkipHostVerification
	}

	return ssl.Dialer(sslCtx, flags), nil
}
