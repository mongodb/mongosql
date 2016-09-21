// Package openssl implements connection to MongoDB over SSL.
package openssl

import (
	"fmt"
	"net"
	"time"

	"github.com/10gen/sqlproxy/options"
	"github.com/10gen/sqlproxy/util"
	"github.com/spacemonkeygo/openssl"
	"gopkg.in/mgo.v2"
)

type (
	// SSLDBConnector is a connector for dialing the database, with SSL.
	SSLDBConnector struct {
		dialInfo  *mgo.DialInfo
		dialError error
		ctx       *openssl.Ctx
	}

	dialerFunc func(addr *mgo.ServerAddr) (net.Conn, error)

	sslInitializationFunction func(options.Options) error
)

var (
	sslInitializationFunctions []sslInitializationFunction
)

func (s *SSLDBConnector) ConfigureDrdl(opts options.DrdlOptions) error {
	connectionAddrs := util.CreateConnectionAddrs(opts.Host, opts.Port)

	var err error

	s.ctx, err = setupDrdlCtx(opts)
	if err != nil {
		return fmt.Errorf("openssl configuration: %v", err)
	}

	var flags openssl.DialFlags
	flags = 0

	if opts.SSLAllowInvalidCert || opts.SSLAllowInvalidHost || opts.SSLCAFile == "" {
		flags = openssl.InsecureSkipHostVerification
	}

	dialer := func(addr *mgo.ServerAddr) (net.Conn, error) {
		conn, err := openssl.Dial("tcp", addr.String(), s.ctx, flags)
		s.dialError = err
		return conn, err
	}

	timeout := time.Duration(opts.Timeout) * time.Second

	s.dialInfo = &mgo.DialInfo{
		Addrs:          connectionAddrs,
		Timeout:        timeout,
		Direct:         opts.Direct,
		ReplicaSetName: opts.ReplicaSetName,
		DialServer:     dialer,
		Username:       opts.DrdlAuth.Username,
		Password:       opts.DrdlAuth.Password,
		Source:         opts.GetAuthenticationDatabase(),
		Mechanism:      opts.DrdlAuth.Mechanism,
	}

	return nil
}

func (s *SSLDBConnector) ConfigureSqld(opts options.SqldOptions) error {

	dialInfo, err := mgo.ParseURL(opts.MongoURI)
	if err != nil {
		return fmt.Errorf("parse URL: %v", err)
	}

	if dialInfo.Username != "" || dialInfo.Password != "" || dialInfo.Source != "" {
		return fmt.Errorf("--mongo-uri may not contain any authentication information")
	}
	if dialInfo.Database != "" {
		return fmt.Errorf("--mongo-uri may not contain database name")
	}
	if dialInfo.Mechanism != "" {
		return fmt.Errorf("--mongo-uri may not contain any authentication mechanism")
	}
	if dialInfo.Service != "" {
		return fmt.Errorf("unsupported: --mongo-uri may not contain GSSAPI service name")
	}
	if dialInfo.ServiceHost != "" {
		return fmt.Errorf("unsupported: --mongo-uri may not contain GSSAPI hostname")
	}

	dialInfo.Timeout = time.Duration(opts.MongoTimeout) * time.Second

	s.dialInfo = dialInfo

	ctx, err := setupSqldCtx(opts)
	if err != nil {
		return fmt.Errorf("openssl configuration: %v", err)
	}

	s.ctx = ctx

	var flags openssl.DialFlags
	flags = 0

	if opts.MongoAllowInvalidCerts || opts.MongoSSLAllowInvalidHost || opts.MongoCAFile == "" {
		flags = openssl.InsecureSkipHostVerification
	}

	dialer := func(addr *mgo.ServerAddr) (net.Conn, error) {
		conn, err := openssl.Dial("tcp", addr.String(), s.ctx, flags)
		s.dialError = err
		return conn, err
	}

	s.dialInfo.DialServer = dialer

	return nil
}

func (s *SSLDBConnector) GetNewSession() (*mgo.Session, error) {
	session, err := mgo.DialWithInfo(s.dialInfo)
	if err != nil && s.dialError != nil {
		return nil, fmt.Errorf("%v, openssl error: %v", err, s.dialError)
	}
	return session, err
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

func setupSqldCtx(opts options.SqldOptions) (*openssl.Ctx, error) {
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
	if opts.MongoPEMFile != "" {
		if err = ctx.UseCertificateChainFile(opts.MongoPEMFile); err != nil {
			return nil, fmt.Errorf("UseCertificateChainFile: %v", err)
		}
		if opts.MongoPEMFilePassword != "" {
			if err = ctx.UsePrivateKeyFileWithPassword(
				opts.MongoPEMFile, openssl.FiletypePEM, opts.MongoPEMFilePassword); err != nil {
				return nil, fmt.Errorf("UsePrivateKeyFile: %v", err)
			}
		} else {
			if err = ctx.UsePrivateKeyFile(opts.MongoPEMFile, openssl.FiletypePEM); err != nil {
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

	if opts.MongoCAFile != "" {
		calist, err := openssl.LoadClientCAFile(opts.MongoCAFile)
		if err != nil {
			return nil, fmt.Errorf("LoadClientCAFile: %v", err)
		}
		ctx.SetClientCAList(calist)

		if err = ctx.LoadVerifyLocations(opts.MongoCAFile, ""); err != nil {
			return nil, fmt.Errorf("LoadVerifyLocations: %v", err)
		}

		var verifyOption openssl.VerifyOptions
		if opts.MongoAllowInvalidCerts {
			verifyOption = openssl.VerifyNone
		} else {
			verifyOption = openssl.VerifyPeer
		}
		ctx.SetVerify(verifyOption, nil)
	}

	if opts.MongoSSLCRLFile != "" {
		store := ctx.GetCertificateStore()
		store.SetFlags(openssl.CRLCheck)
		lookup, err := store.AddLookup(openssl.X509LookupFile())
		if err != nil {
			return nil, fmt.Errorf("AddLookup(X509LookupFile()): %v", err)
		}
		lookup.LoadCRLFile(opts.MongoSSLCRLFile)
	}

	return ctx, nil
}
