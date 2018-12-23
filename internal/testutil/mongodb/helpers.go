package mongodb

import (
	"os"

	"github.com/10gen/sqlproxy/internal/testutil/flags"
	"github.com/10gen/sqlproxy/internal/util"

	toolsoptions "github.com/mongodb/mongo-tools/common/options"
	"gopkg.in/mgo.v2"
)

const (
	// SSLTestKey is the name of an environment variable that can be set to
	// indicate that sqlproxy tests will need to enable SSL if they want to
	// connect to mongodb.
	SSLTestKey = "SQLPROXY_SSLTEST"

	// AuthTestKey is the name of an environment variable that can be set to
	// indicate that sqlproxy tests will need to enable authentication if they want to
	// connect to mongodb.
	AuthTestKey = "SQLPROXY_AUTHTEST"
)

// GetToolOptions returns options for connecting to MongoDB via mongo-tools. The options are based
// on the values supplied for the  -mongoHost and -mongoPort flags.
func GetToolOptions() *toolsoptions.ToolOptions {
	opts := &toolsoptions.ToolOptions{
		Namespace: &toolsoptions.Namespace{},
		Connection: &toolsoptions.Connection{
			Host: *flags.Host,
			Port: *flags.Port,
		},
		Direct: false,
		URI:    &toolsoptions.URI{},
		Auth:   getAuthOpts(),
		SSL:    getSslOpts(),
	}
	return opts
}

// VersionAtLeast checks if the server this session is connected to has a version greater than or
// equal to the provided minVersion.
func VersionAtLeast(session *mgo.Session, minVersion string) (bool, error) {
	if minVersion == "" {
		return true, nil
	}

	minRequiredVersion, err := util.VersionToSlice(minVersion)
	if err != nil {
		return false, err
	}

	buildInfo, err := session.BuildInfo()
	if err != nil {
		return false, err
	}

	serverVersion, err := util.VersionToSlice(buildInfo.Version)
	if err != nil {
		return false, err
	}

	return util.VersionAtLeast(serverVersion, minRequiredVersion), nil
}

func getAuthOpts() *toolsoptions.Auth {
	if len(os.Getenv(AuthTestKey)) > 0 {
		return &toolsoptions.Auth{
			Username: "bob",
			Password: "pwd123",
		}
	}
	return &toolsoptions.Auth{}
}

func getSslOpts() *toolsoptions.SSL {
	if len(os.Getenv(SSLTestKey)) > 0 {
		return &toolsoptions.SSL{
			UseSSL:              true,
			SSLPEMKeyFile:       *flags.ClientPEMKeyFile,
			SSLAllowInvalidCert: true,
		}
	}
	return &toolsoptions.SSL{}
}
