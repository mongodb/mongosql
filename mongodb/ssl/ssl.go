package ssl

import (
	"context"
	"fmt"
	"net"

	"github.com/10gen/openssl"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/procutil"
	"github.com/10gen/sqlproxy/log"

	"go.mongodb.org/mongo-driver/x/mongo/driver/topology"
)

// nolint: golint
var (
	FipsModeSetter func(bool) error
)

// SqldDialer creates a mongosqld dialer.
func SqldDialer(cfg *config.Config) (topology.DialerFunc, error) {
	sslCtx, err := createSqldSSLContext(cfg, true)
	if err != nil {
		return nil, err
	}
	var flags openssl.DialFlags

	if cfg.MongoDB.Net.SSL.AllowInvalidCertificates ||
		cfg.MongoDB.Net.SSL.AllowInvalidHostnames ||
		cfg.MongoDB.Net.SSL.CAFile == "" {
		flags = openssl.InsecureSkipHostVerification
	}

	return Dialer(sslCtx, flags), nil
}

func createSqldSSLContext(cfg *config.Config, isClient bool) (*openssl.Ctx, error) {
	var ctx *openssl.Ctx
	var err error

	if cfg.MongoDB.Net.SSL.FIPSMode {
		if FipsModeSetter == nil {
			return nil, fmt.Errorf("configured to use FIPS mode, but no FIPS mode setter available")
		}
		if err = FipsModeSetter(true); err != nil {
			return nil, err
		}
		log.Infof(log.Admin, "enabled OpenSSL's FIPS mode")
	}

	if ctx, err = openssl.NewCtxWithVersion(openssl.AnyVersion); err != nil {
		return nil, fmt.Errorf("failure creating new openssl context with "+
			"NewCtxWithVersion(AnyVersion): %v", err)
	}

	if isClient {
		SetMinimumTLSProtocolVersion(cfg.MongoDB.Net.SSL.MinimumTLSVersion, ctx)
	} else {
		SetMinimumTLSProtocolVersion(cfg.Net.SSL.MinimumTLSVersion, ctx)
	}

	// HIGH - Enable strong ciphers
	// !EXPORT - Disable export ciphers (40/56 bit)
	// !aNULL - Disable anonymous auth ciphers
	// @STRENGTH - Sort ciphers based on strength
	if err = ctx.SetCipherList("HIGH:!EXPORT:!aNULL@STRENGTH"); err != nil {
		return nil, err
	}

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
		if err := store.SetFlags(openssl.CRLCheck); err != nil {
			return nil, err
		}
		lookup, err := store.AddLookup(openssl.X509LookupFile())
		if err != nil {
			return nil, fmt.Errorf("AddLookup(X509LookupFile()): %v", err)
		}
		if err := lookup.LoadCRLFile(crlFile); err != nil {
			return nil, err
		}
	}

	return ctx, nil
}

// nolint: golint
func Dialer(sslCtx *openssl.Ctx, flags openssl.DialFlags) topology.DialerFunc {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		var c net.Conn
		var err error
		ch := make(chan struct{})
		errChan := make(chan error, 1)

		procutil.PanicSafeGo(func() {
			c, err = openssl.Dial(network, addr, sslCtx, flags)
			ch <- struct{}{}
		}, func(dialErr interface{}) {
			errChan <- fmt.Errorf("openssl dial error: %v", dialErr)
		})

		select {
		case <-ch:
		case <-ctx.Done():
			return nil, ctx.Err()
		case chanErr := <-errChan:
			return nil, chanErr
		}
		return c, err
	}
}

// SetMinimumTLSProtocolVersion sets the minimum TLS version in the OpenSSL context.
func SetMinimumTLSProtocolVersion(minTLS string, ctx *openssl.Ctx) {
	// OpAll - Activate all bug workaround options, to support buggy client SSL's.
	// NoSSLv2 - Disable SSL v2 support
	// NoSSLv3 - Disable SSL v3 support
	defaultOptions := openssl.OpAll | openssl.NoSSLv2 | openssl.NoSSLv3

	switch minTLS {
	case config.TLSv1_0:
		ctx.SetOptions(defaultOptions)
	case config.TLSv1_1:
		ctx.SetOptions(defaultOptions | openssl.NoTLSv1)
	case config.TLSv1_2:
		ctx.SetOptions(defaultOptions | openssl.NoTLSv1 | openssl.NoTLSv1_1)
	default:
		panic(fmt.Sprintf("invalid minimum TLS version: %v", minTLS))
	}
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
