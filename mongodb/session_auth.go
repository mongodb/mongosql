package mongodb

import (
	"context"

	"fmt"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/mongo-go-driver/yamgo/private/auth"
	"github.com/10gen/mongo-go-driver/yamgo/private/conn"
	"github.com/10gen/mongo-go-driver/yamgo/private/msg"
)

// SessionAuthenticator authenticates a session.
type SessionAuthenticator interface {
	// Auth handles authenticating the session.
	Auth(context.Context, []conn.Connection) error

	source() string
}

// CleartextSessionAuthenticator authentices a session
// using the cleartext protocol.
type CleartextSessionAuthenticator struct {
	Source    string
	Username  string
	Password  string
	Mechanism string
}

func (a *CleartextSessionAuthenticator) source() string {
	return a.Source
}

// Auth handles authenticating the session.
func (a *CleartextSessionAuthenticator) Auth(ctx context.Context, conns []conn.Connection) error {

	authCred := &auth.Cred{
		Source:      a.Source,
		Username:    a.Username,
		Password:    a.Password,
		PasswordSet: a.Password != "",
	}

	authenticator, err := auth.CreateAuthenticator(a.Mechanism, authCred)
	if err != nil {
		return err
	}

	for i := 0; i < len(conns); i++ {
		err := authenticator.Auth(ctx, conns[i])
		if err != nil {
			return fmt.Errorf("unable to authenticate conversation %d: %s", i, err)
		}
	}

	return nil
}

// SaslConversation is a single conversation occuring
// over the sasl protocol.
type SaslConversation struct {
	Payload    []byte
	ClientDone bool

	id         int
	serverDone bool
}

// SaslSessionAuthenticator authenticates a session using
// the sasl protocol.
type SaslSessionAuthenticator struct {
	Source    string
	Username  string
	Mechanism string

	Callback func(convos []*SaslConversation) error

	conversations saslConversations
}

type saslConversations []*SaslConversation

func (sc saslConversations) AllDone() bool {
	for _, c := range sc {
		if !c.ClientDone {
			return false
		}
		if !c.serverDone {
			return false
		}
	}

	return true
}

func (a *SaslSessionAuthenticator) source() string {
	return a.Source
}

// AddConversation adds a new conversation to the SaslSessionAuthenticator.
func (a *SaslSessionAuthenticator) AddConversation(payload []byte, done bool) {
	a.conversations = append(a.conversations, &SaslConversation{
		Payload:    payload,
		ClientDone: done,
	})
}

// Auth handles authenticating the session.
func (a *SaslSessionAuthenticator) Auth(ctx context.Context, conns []conn.Connection) error {
	source := a.Source

	// So, because sasl is a generic protocol, it can be client first or server first and client last
	// or server last. As such, we need to wait until both the client and the server have said they are done
	// in order to finalize the conversation.

	type saslResponse struct {
		ConversationID int    `bson:"conversationId"`
		Code           int    `bson:"code"`
		Done           bool   `bson:"done"`
		Payload        []byte `bson:"payload"`
	}

	var err error
	for i := 0; i < len(a.conversations); i++ {
		saslStartRequest := msg.NewCommand(
			msg.NextRequestID(),
			source,
			true,
			bson.D{
				{"saslStart", 1},
				{"mechanism", a.Mechanism},
				{"payload", a.conversations[i].Payload},
			},
		)

		var saslResp saslResponse
		err = conn.ExecuteCommand(ctx, conns[i], saslStartRequest, &saslResp)
		if err != nil || saslResp.Code != 0 {
			return fmt.Errorf("unable to saslStart conversation %d: %s", i, err)
		}

		a.conversations[i].id = saslResp.ConversationID
		a.conversations[i].serverDone = saslResp.Done
		a.conversations[i].Payload = saslResp.Payload
	}

	for {
		if a.conversations.AllDone() {
			return nil
		}

		err = a.Callback(a.conversations)
		if err != nil {
			return err
		}

		if a.conversations.AllDone() {
			return nil
		}

		for i := 0; i < len(a.conversations); i++ {
			saslContinueRequest := msg.NewCommand(
				msg.NextRequestID(),
				source,
				true,
				bson.D{
					{"saslContinue", 1},
					{"conversationId", a.conversations[i].id},
					{"payload", a.conversations[i].Payload},
				},
			)

			var saslResp saslResponse
			err = conn.ExecuteCommand(ctx, conns[i], saslContinueRequest, &saslResp)
			if err != nil {
				return fmt.Errorf("unable to saslContinue conversation %d: %s", i, err)
			}

			a.conversations[i].serverDone = saslResp.Done
			a.conversations[i].Payload = saslResp.Payload
		}
	}
}
