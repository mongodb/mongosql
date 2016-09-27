package testutils

import (
	"github.com/10gen/sqlproxy/options"
	toolsoptions "github.com/mongodb/mongo-tools/common/options"
)

func GetSSLOpts() *toolsoptions.SSL {
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
