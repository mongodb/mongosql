// +build !gssapi

package mongodb

import (
	"context"
	"fmt"

	"github.com/10gen/mongo-go-driver/mongo/private/conn"
)

// Auth handles authenticating the session.
func (a *GssapiSessionAuthenticator) Auth(ctx context.Context, conns []conn.Connection) error {
	return fmt.Errorf("GSSAPI support not enabled during build (-tags gssapi)")
}
