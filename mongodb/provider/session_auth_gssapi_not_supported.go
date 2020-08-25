//+build gssapi,!linux,!darwin,!windows

package provider

import (
	"context"
	"fmt"
	"runtime"

	"go.mongodb.org/mongo-driver/x/mongo/driver"
)

// Auth handles authenticating the session.
func (a *GssapiSessionAuthenticator) Auth(ctx context.Context, conns []driver.Connection) error {
	return fmt.Errorf("GSSAPI is not supported on %s", runtime.GOOS)
}
