package ssl

import (
	"context"
	"fmt"
	"net"

	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/options"
	"github.com/10gen/sqlproxy/util"
	"github.com/spacemonkeygo/openssl"
)

var (
	fipsModeSetter func(bool) error
)

// SqldDialer creates a mongosqld dialer.
func SqldDialer(cfg *config.Config) (func(ctx context.Context, network, address string) (net.Conn, error), error) {
	sslCtx, err := createSqldSSLContext(cfg, true)
	if err != nil {
		return nil, err
	}
	var flags openssl.DialFlags

	if cfg.MongoDB.Net.SSL.AllowInvalidCertificates || cfg.MongoDB.Net.SSL.AllowInvalidHostnames || cfg.MongoDB.Net.SSL.CAFile == "" {
		flags = openssl.InsecureSkipHostVerification
	}

	return dialer(sslCtx, flags), nil
}

// Handshake performs a TLS handshake over the connection.
func Handshake(conn net.Conn, cfg *config.Config) (net.Conn, error) {
	sslCtx, err := createSqldSSLContext(cfg, false)
	if err != nil {
		return nil, err
	}

	tlsConn, err := openssl.Server(conn, sslCtx)
	if err != nil {
		return nil, err
	}

	err = tlsConn.Handshake()
	if err != nil {
		return nil, err
	}

	return tlsConn, nil
}

// DrdlDialer creates a mongodrdl dialer.
func DrdlDialer(opts options.DrdlOptions) (func(ctx context.Context, network, address string) (net.Conn, error), error) {
	sslCtx, err := createDrdlSSLContext(opts)
	if err != nil {
		return nil, err
	}
	var flags openssl.DialFlags

	if opts.SSLAllowInvalidCert || opts.SSLAllowInvalidHost || opts.SSLCAFile == "" {
		flags = openssl.InsecureSkipHostVerification
	}

	return dialer(sslCtx, flags), nil
}

func dialer(sslCtx *openssl.Ctx, flags openssl.DialFlags) func(ctx context.Context, network, address string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		var c net.Conn
		var err error
		ch := make(chan struct{})
		errChan := make(chan error, 1)

		util.PanicSafeGo(func() {
			c, err = openssl.Dial(network, address, sslCtx, flags)
			ch <- struct{}{}
		}, func(err interface{}) {
			errChan <- fmt.Errorf("openssl dial error: %v", err)
		})

		select {
		case <-ch:
		case <-ctx.Done():
			return nil, ctx.Err()
		case err := <-errChan:
			return nil, err
		}
		return c, err
	}
}

func createDrdlSSLContext(opts options.DrdlOptions) (*openssl.Ctx, error) {
	var ctx *openssl.Ctx
	var err error

	if fipsModeSetter != nil {
		fipsModeSetter(opts.UseFIPSMode())
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

func createSqldSSLContext(cfg *config.Config, isClient bool) (*openssl.Ctx, error) {
	var ctx *openssl.Ctx
	var err error

	if fipsModeSetter != nil {
		fipsModeSetter(cfg.MongoDB.Net.SSL.FIPSMode)
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

	if !isClient {
		pemKeyFile = cfg.Net.SSL.PEMKeyFile
		pemFilePassword = cfg.Net.SSL.PEMKeyPassword
		caFile = cfg.Net.SSL.CAFile
		allowInvalidCerts = cfg.Net.SSL.AllowInvalidCertificates
	} else {
		pemKeyFile = cfg.MongoDB.Net.SSL.PEMKeyFile
		pemFilePassword = cfg.MongoDB.Net.SSL.PEMKeyPassword
		caFile = cfg.MongoDB.Net.SSL.CAFile
		allowInvalidCerts = cfg.MongoDB.Net.SSL.AllowInvalidCertificates
		crlFile = cfg.MongoDB.Net.SSL.CRLFile
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
