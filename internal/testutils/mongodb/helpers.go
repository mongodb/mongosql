package mongodb

import (
	"os"
	"strconv"
	"strings"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/options"
	toolsoptions "github.com/mongodb/mongo-tools/common/options"
)

// VersionAtLeast checks whether the provided version string represents a
// version greater than or equal to the value specified via the -serverVersion
// flag. If the flag was not specified, this function will always return true.
func VersionAtLeast(versionString string) bool {
	if versionString == "" {
		return true
	}

	strServerVersion := strings.Split(*ServerVersion, ".")
	serverVersion := make([]uint8, len(strServerVersion))
	for i, str := range strServerVersion {
		num, err := strconv.ParseInt(str, 0, 0)
		if err != nil {
			panic(err)
		}
		serverVersion[i] = uint8(num)
	}

	strVersion := strings.Split(versionString, ".")
	version := make([]uint8, len(strVersion))
	for i, str := range strVersion {
		num, err := strconv.ParseInt(str, 0, 0)
		if err != nil {
			panic(err)
		}
		version[i] = uint8(num)
	}

	return util.VersionAtLeast(serverVersion, version)
}

// GetToolOptions returns options for connecting to MongoDB via mongo-tools.
// The options are based on the values supplied for the  -mongoHost and
// -mongoPort flags.
func GetToolOptions() *toolsoptions.ToolOptions {
	opts := &toolsoptions.ToolOptions{
		Namespace: &toolsoptions.Namespace{},
		Connection: &toolsoptions.Connection{
			Host: *Host,
			Port: *Port,
		},
		Direct: false,
		URI:    &toolsoptions.URI{},
		SSL:    getSslOpts(),
		Auth:   getAuthOpts(),
	}
	return opts
}

// DrdlTestSSLOpts returns the mongodrdl ssl options to use for testing.
func DrdlTestSSLOpts() *options.DrdlSSL {
	return &options.DrdlSSL{
		UseSSL:              true,
		SSLPEMKeyFile:       "../testdata/resources/x509gen/client.pem",
		SSLAllowInvalidCert: true,
	}
}

func sqldTestSSLOpts() *toolsoptions.SSL {
	return &toolsoptions.SSL{
		UseSSL:              true,
		SSLPEMKeyFile:       "testdata/resources/x509gen/client.pem",
		SSLAllowInvalidCert: true,
	}
}

func getSslOpts() *toolsoptions.SSL {
	sslOpts := &toolsoptions.SSL{}

	if len(os.Getenv(evaluator.SSLTestKey)) > 0 {
		return sqldTestSSLOpts()
	}

	return sslOpts
}

func sqldTestAuthOpts() *toolsoptions.Auth {
	return &toolsoptions.Auth{
		Username: "bob",
		Password: "pwd123",
	}
}

func getAuthOpts() *toolsoptions.Auth {
	authOpts := &toolsoptions.Auth{}
	if len(os.Getenv("SQLPROXY_AUTHTEST")) > 0 {
		return sqldTestAuthOpts()
	}
	return authOpts
}
