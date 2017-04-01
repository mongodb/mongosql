// Package openssl implements connection to MongoDB over SSL.
package openssl

import (
	"context"
	"fmt"
	"net"

	"github.com/10gen/mongo-go-driver/cluster"
	"github.com/10gen/mongo-go-driver/model"
	"github.com/10gen/mongo-go-driver/readpref"

	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/options"

	"github.com/spacemonkeygo/openssl"
)

type (
	// SSLDBConnector is a connector for dialing the database, with SSL.
	SSLDBConnector struct {
		ctx      *openssl.Ctx
		dialInfo *mongodb.DialInfo
	}

	dialerFunc func(ctx context.Context, addr model.Addr) (net.Conn, error)

	sslInitializationFunction func(options.Options) error
)

var (
	sslInitializationFunctions []sslInitializationFunction
)

func (s *SSLDBConnector) ConfigureDrdl(opts options.DrdlOptions, dialInfo *mongodb.DialInfo) error {
	ctx, err := setupDrdlCtx(opts)
	if err != nil {
		return fmt.Errorf("openssl configuration: %v", err)
	}

	s.ctx = ctx

	var flags openssl.DialFlags
	flags = 0

	if opts.SSLAllowInvalidCert || opts.SSLAllowInvalidHost || opts.SSLCAFile == "" {
		flags = openssl.InsecureSkipHostVerification
	}

	s.dialInfo = dialInfo
	s.setNetDialer(flags)
	return err
}

func (s *SSLDBConnector) ConfigureSqld(opts options.SqldOptions, dialInfo *mongodb.DialInfo) error {
	ctx, err := SetupSqldCtx(opts, false)
	if err != nil {
		return fmt.Errorf("openssl configuration: %v", err)
	}

	s.ctx = ctx

	var flags openssl.DialFlags
	flags = 0

	if *opts.MongoAllowInvalidCerts || *opts.MongoSSLAllowInvalidHost || *opts.MongoCAFile == "" {
		flags = openssl.InsecureSkipHostVerification
	}

	s.dialInfo = dialInfo
	s.setNetDialer(flags)
	return nil
}

func (s *SSLDBConnector) GetNewSession(ctx context.Context, monitor *cluster.Monitor, readPreference *readpref.ReadPref) (*mongodb.Session, error) {

	session, err := s.dialInfo.Dial(ctx, monitor, readPreference)
	if err != nil {
		return nil, err
	}
	return session, err
}

func (s *SSLDBConnector) setNetDialer(flags openssl.DialFlags) {
	s.dialInfo.NetDialer = func(ctx context.Context, network, address string) (net.Conn, error) {
		var conn net.Conn
		var err error
		connChan := make(chan struct{})

		go func() {
			conn, err = openssl.Dial(network, address, s.ctx, flags)
			connChan <- struct{}{}
		}()

		select {
		case <-connChan:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		return conn, err
	}
}

func setupDrdlCtx(opts options.DrdlOptions) (*openssl.Ctx, error) {
	var ctx *openssl.Ctx
	var err error

	for _, sslInitFunc := range sslInitializationFunctions {
		sslInitFunc(opts)
	}

	if ctx, err = openssl.NewCtxWithVersion(openssl.AnyVersion); err != nil {
		return nil, fmt.Errorf("failure creating new openssl context with "+
			"NewCtxWithVersion(AnyVersion): %v", err)
	}

	// OpAll - Activate all bug workaround options, to support buggy client SSL's.
	// NoSSLv2 - Disable SSL v2 support
	ctx.SetOptions(openssl.OpAll | openssl.NoSSLv2)

	// HIGH - Enable strong ciphers
	// !EXPORT - Disable export ciphers (40/56 bit)
	// !aNULL - Disable anonymous auth ciphers
	// @STRENGTH - Sort ciphers based on strength
	ctx.SetCipherList("HIGH:!EXPORT:!aNULL@STRENGTH")

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
		calist, err := openssl.LoadClientCAFile(opts.SSLCAFile)
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
		store.SetFlags(openssl.CRLCheck)
		lookup, err := store.AddLookup(openssl.X509LookupFile())
		if err != nil {
			return nil, fmt.Errorf("AddLookup(X509LookupFile()): %v", err)
		}
		lookup.LoadCRLFile(opts.SSLCRLFile)
	}

	return ctx, nil
}

func SetupSqldCtx(opts options.SqldOptions, isClient bool) (*openssl.Ctx, error) {
	var ctx *openssl.Ctx
	var err error

	for _, sslInitFunc := range sslInitializationFunctions {
		sslInitFunc(opts)
	}

	if ctx, err = openssl.NewCtxWithVersion(openssl.AnyVersion); err != nil {
		return nil, fmt.Errorf("failure creating new openssl context with "+
			"NewCtxWithVersion(AnyVersion): %v", err)
	}

	// OpAll - Activate all bug workaround options, to support buggy client SSL's.
	// NoSSLv2 - Disable SSL v2 support
	ctx.SetOptions(openssl.OpAll | openssl.NoSSLv2)

	// HIGH - Enable strong ciphers
	// !EXPORT - Disable export ciphers (40/56 bit)
	// !aNULL - Disable anonymous auth ciphers
	// @STRENGTH - Sort ciphers based on strength
	ctx.SetCipherList("HIGH:!EXPORT:!aNULL@STRENGTH")

	var pemKeyFile, pemFilePassword, caFile, crlFile string
	var allowInvalidCerts bool

	if isClient {
		pemKeyFile = *opts.SSLPEMKeyFile
		pemFilePassword = *opts.SSLPEMKeyFilePassword
		caFile = *opts.SSLCAFile
		allowInvalidCerts = *opts.SSLAllowInvalidCerts
	} else {
		pemKeyFile = *opts.MongoPEMKeyFile
		pemFilePassword = *opts.MongoPEMKeyFilePassword
		caFile = *opts.MongoCAFile
		allowInvalidCerts = *opts.MongoAllowInvalidCerts
		crlFile = *opts.MongoSSLCRLFile
	}

	// add the PEM key file with the cert and private key, if specified
	if pemKeyFile != "" {
		if err = ctx.UseCertificateChainFile(pemKeyFile); err != nil {
			return nil, fmt.Errorf("UseCertificateChainFile: %v", err)
		}
		if pemFilePassword != "" {
			if err = ctx.UsePrivateKeyFileWithPassword(
				pemKeyFile, openssl.FiletypePEM, pemFilePassword); err != nil {
				return nil, fmt.Errorf("UsePrivateKeyFile: %v", err)
			}
		} else {
			if err = ctx.UsePrivateKeyFile(pemKeyFile, openssl.FiletypePEM); err != nil {
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

	if caFile != "" {
		calist, err := openssl.LoadClientCAFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("LoadClientCAFile: %v", err)
		}
		ctx.SetClientCAList(calist)

		if err = ctx.LoadVerifyLocations(caFile, ""); err != nil {
			return nil, fmt.Errorf("LoadVerifyLocations: %v", err)
		}

		var verifyOption openssl.VerifyOptions
		if allowInvalidCerts {
			verifyOption = openssl.VerifyNone
		} else {
			verifyOption = openssl.VerifyPeer
		}
		ctx.SetVerify(verifyOption, nil)
	}

	if crlFile != "" {
		store := ctx.GetCertificateStore()
		store.SetFlags(openssl.CRLCheck)
		lookup, err := store.AddLookup(openssl.X509LookupFile())
		if err != nil {
			return nil, fmt.Errorf("AddLookup(X509LookupFile()): %v", err)
		}
		lookup.LoadCRLFile(crlFile)
	}

	return ctx, nil
}

// Server wraps an existing stream connection and puts it in the accept state
// for any subsequent handshakes.
func Server(conn net.Conn, ctx *openssl.Ctx) (*openssl.Conn, error) {
	return openssl.Server(conn, ctx)
}
