// +build gssapi,!linux,!darwin

package mongodb

import (
	"context"
	"fmt"
	"runtime"

	"github.com/10gen/mongo-go-driver/mongo/private/conn"
)

// Auth handles authenticating the session.
func (a *GssapiSessionAuthenticator) Auth(ctx context.Context, conns []conn.Connection) error {
	return fmt.Errorf("GSSAPI is not supported on %s", runtime.GOOS)
}
