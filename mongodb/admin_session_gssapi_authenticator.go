//+build gssapi

package mongodb

import (
	"context"

	"github.com/10gen/sqlproxy/internal/config"

	"github.com/10gen/mongo-go-driver/mongo/private/auth"
	"github.com/10gen/mongo-go-driver/mongo/private/conn"
)

type gssapiAuthenticatorWrapper struct {
	authenticator auth.GSSAPIAuthenticator
}

// nolint: unparam
func newAdminSessionGSSAPIAuthenticator(cfg config.MongoDBNetAuth) (SessionAuthenticator, error) {
	return &gssapiAuthenticatorWrapper{
		authenticator: auth.GSSAPIAuthenticator{
			Username:    cfg.Username,
			Password:    cfg.Password,
			PasswordSet: cfg.Password != "",
			Props: map[string]string{
				"SERVICE_NAME": cfg.GSSAPIServiceName,
			},
		},
	}, nil
}

// Auth will attempt to authenticate the provided connections using GSSAPI.
func (g *gssapiAuthenticatorWrapper) Auth(ctx context.Context, conns []conn.Connection) error {
	for _, c := range conns {
		err := g.authenticator.Auth(ctx, c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *gssapiAuthenticatorWrapper) source() string {
	return "$external"
}
