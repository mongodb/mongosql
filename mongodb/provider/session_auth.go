package provider

import (
	"context"
	"fmt"

	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/mongodb/internal/mongoutil"

	"go.mongodb.org/mongo-driver/x/mongo/driver"
	"go.mongodb.org/mongo-driver/x/mongo/driver/auth"
)

// SessionAuthenticator authenticates a session.
type SessionAuthenticator interface {
	// Auth handles authenticating the session.
	Auth(context.Context, []driver.Connection) error

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
func (a *CleartextSessionAuthenticator) Auth(ctx context.Context, conns []driver.Connection) error {

	authCred := &auth.Cred{
		Source:      a.Source,
		Username:    a.Username,
		Password:    a.Password,
		PasswordSet: a.Password != "",
	}

	var err error
	var authenticator auth.Authenticator

	if a.Mechanism == SCRAMSHA256 {
		authenticator, err = newScramSHA256Authenticator(authCred)
	} else {
		authenticator, err = auth.CreateAuthenticator(a.Mechanism, authCred)
	}
	if err != nil {
		return err
	}

	for i := 0; i < len(conns); i++ {
		err := authenticator.Auth(ctx, conns[i].Description(), conns[i])
		if err != nil {
			return fmt.Errorf("unable to authenticate conversation %d: %s", i, err)
		}
	}

	return nil
}

// SaslConversation is a single conversation occurring
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
	Source                string
	Username              string
	Mechanism             string
	ConstrainedDelegation bool

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

// SaslResponse represents the server response to
// saslStart and saslContinue messages.
type SaslResponse struct {
	ConversationID int    `bson:"conversationId"`
	Code           int    `bson:"code"`
	Done           bool   `bson:"done"`
	Payload        []byte `bson:"payload"`
	Ok             int    `bson:"ok"`
}

// Auth handles authenticating the session.
func (a *SaslSessionAuthenticator) Auth(ctx context.Context, conns []driver.Connection) error {
	source := a.Source

	// Because sasl is a generic protocol, it can be client first or server first and client last
	// or server last. As such, we need to wait until both the client and the server have said they
	// are done in order to finalize the conversation.

	var err error
	for i := 0; i < len(a.conversations); i++ {
		saslStartRequest := bsonutil.NewD(
			bsonutil.NewDocElem("saslStart", 1),
			bsonutil.NewDocElem("mechanism", a.Mechanism),
			bsonutil.NewDocElem("payload", a.conversations[i].Payload),
		)

		var saslResp SaslResponse
		err = mongoutil.ExecuteWithConnection(ctx, source, conns[i], nil, saslStartRequest, &saslResp)
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
			saslContinueRequest := bsonutil.NewD(
				bsonutil.NewDocElem("saslContinue", 1),
				bsonutil.NewDocElem("conversationId", a.conversations[i].id),
				bsonutil.NewDocElem("payload", a.conversations[i].Payload),
			)

			var saslResp SaslResponse
			err = mongoutil.ExecuteWithConnection(ctx, source, conns[i], nil, saslContinueRequest, &saslResp)
			if err != nil || saslResp.Code != 0 {
				return fmt.Errorf("unable to saslContinue conversation %d: %s", i, err)
			}

			a.conversations[i].serverDone = saslResp.Done
			a.conversations[i].Payload = saslResp.Payload
		}
	}
}

// GssapiSessionAuthenticator authenticates a session
// using the GSSAPI protocol.
type GssapiSessionAuthenticator struct {
	InitialPayload []byte
	Callback       func([]byte) ([]byte, error)

	HostServiceName string
	HostAddr        string

	RemoteServiceName string

	ConstrainedDelegation bool
}

func (a *GssapiSessionAuthenticator) source() string {
	return "$external"
}
