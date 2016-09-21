// +build ssl

package client

import (
	"github.com/10gen/sqlproxy/client/openssl"
	"github.com/10gen/sqlproxy/options"
)

func init() {
	GetConnectorFuncs = append(GetConnectorFuncs, getSSLConnector)
}

func getSSLConnector(opts options.Options) DBConnector {
	if opts.UseSSL() {
		return &openssl.SSLDBConnector{}
	}
	return nil
}
