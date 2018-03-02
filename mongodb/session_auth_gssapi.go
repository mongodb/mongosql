// +build gssapi
// +build linux darwin

package mongodb

import (
	"context"
	"net"

	"github.com/10gen/mongo-go-driver/mongo/private/auth"
	"github.com/10gen/mongo-go-driver/mongo/private/conn"
	"github.com/10gen/sqlproxy/mongodb/internal/gssapi"
)

// Auth handles authenticating the session.
func (a *GssapiSessionAuthenticator) Auth(ctx context.Context, conns []conn.Connection) error {
	server := gssapi.NewServer(a.HostServiceName, getHostname(a.HostAddr))
	defer server.Close()

	err := server.Start()
	if err != nil {
		return err
	}

	// Some SASL clients will send an empty first message. This is not right or wrong,
	// but we can cover over the extraneous message by sending nothing back to the client
	// and waiting for the next challenge.
	payload := a.InitialPayload
	if len(payload) == 0 {
		payload, err = a.Callback(nil)
		if err != nil {
			return err
		}
	}

	for {
		payload, err = server.Next(payload)
		if err != nil {
			return err
		}

		if server.Completed() {
			break
		}

		payload, err = a.Callback(payload)
		if err != nil {
			return err
		}
	}

	errs := make(chan error, len(conns))
	for i := 0; i < len(conns); i++ {
		c := conns[i]
		client := gssapi.NewClient(
			getHostname(c.Model().Addr.String()),
			server,
			a.RemoteServiceName,
			false,
			"",
		)

		go func(client *gssapi.SaslClient, c conn.Connection, errs chan<- error) {
			errs <- auth.ConductSaslConversation(ctx, c, "$external", client)
			client.Close()
		}(client, c, errs)
	}

	for i := 0; i < len(conns); i++ {
		err = <-errs
		if err != nil {
			return err
		}
	}

	return nil
}

func getHostname(addr string) string {
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return host
	}

	return addr
}
