// +build !gssapi

package provider

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/x/mongo/driver"
)

// Auth handles authenticating the session.
func (a *GssapiSessionAuthenticator) Auth(ctx context.Context, conns []driver.Connection) error {
	return fmt.Errorf("GSSAPI support not enabled during build (-tags gssapi)")
}
