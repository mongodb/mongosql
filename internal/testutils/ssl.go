package testutils

import (
	"os"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/options"
	toolsoptions "github.com/mongodb/mongo-tools/common/options"
)

func GetSSLOpts() *toolsoptions.SSL {
	return &toolsoptions.SSL{
		UseSSL:              true,
		SSLPEMKeyFile:       "testdata/resources/client.pem",
		SSLAllowInvalidCert: true,
	}
}

func GetDrdlSSLOpts() *options.DrdlSSL {
	return &options.DrdlSSL{
		UseSSL:              true,
		SSLPEMKeyFile:       "../testdata/resources/client.pem",
		SSLAllowInvalidCert: true,
	}
}

func getSslOpts() *toolsoptions.SSL {
	sslOpts := &toolsoptions.SSL{}

	if len(os.Getenv(evaluator.SSLTestKey)) > 0 {
		return GetSSLOpts()
	}

	return sslOpts
}
