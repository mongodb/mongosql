//+build !gssapi

package mongodb

import (
	"fmt"

	"github.com/10gen/sqlproxy/internal/config"
)

func newAdminSessionGSSAPIAuthenticator(cfg config.MongoDBNetAuth) (SessionAuthenticator, error) {
	return nil, fmt.Errorf("GSSAPI authentication not supported on this platform")
}
