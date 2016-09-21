// +build ssl

package testutils

import (
	"github.com/10gen/sqlproxy/client"
	"github.com/10gen/sqlproxy/client/openssl"
	"github.com/10gen/sqlproxy/options"
	toolsoptions "github.com/mongodb/mongo-tools/common/options"
)

func GetSSLOpts() *toolsoptions.SSL {
	db.GetConnectorFuncs = append(db.GetConnectorFuncs,
		func(opts options.Options) db.DBConnector {
			if opts.UseSSL() {
				return &openssl.SSLDBConnector{}
			}
			return nil
		},
	)

	return &toolsoptions.SSL{
		UseSSL:              true,
		SSLPEMKeyFile:       "testdata/client.pem",
		SSLAllowInvalidCert: true,
	}
}

func GetDrdlSSLOpts() *options.DrdlSSL {
	return &options.DrdlSSL{
		UseSSL:              true,
		SSLPEMKeyFile:       "../testdata/client.pem",
		SSLAllowInvalidCert: true,
	}
}
