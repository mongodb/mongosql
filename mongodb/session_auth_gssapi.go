//+build gssapi
//+build linux darwin windows

package mongodb

import (
	"context"
	"net"
	"runtime"

	"github.com/10gen/sqlproxy/mongodb/internal/gssapi"

	"go.mongodb.org/mongo-driver/x/mongo/driver"
	"go.mongodb.org/mongo-driver/x/mongo/driver/auth"
)

// Auth handles authenticating the session.
func (a *GssapiSessionAuthenticator) Auth(ctx context.Context, conns []driver.Connection) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	server := gssapi.NewServer(a.HostServiceName, getHostname(a.HostAddr), a.ConstrainedDelegation)
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

	for i := 0; i < len(conns); i++ {
		c := conns[i]
		client := gssapi.NewClient(
			getHostname(c.Address().String()),
			server,
			a.RemoteServiceName,
			false,
			"",
		)
		err := auth.ConductSaslConversation(ctx, c, "$external", client)
		client.Close()

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
