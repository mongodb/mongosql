package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/internal/bsonutil"
	. "github.com/10gen/sqlproxy/mongodb"
	"github.com/stretchr/testify/require"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
	"go.mongodb.org/mongo-driver/x/mongo/driver"
	"go.mongodb.org/mongo-driver/x/mongo/driver/address"
	"go.mongodb.org/mongo-driver/x/mongo/driver/auth"
	"go.mongodb.org/mongo-driver/x/mongo/driver/description"
	"go.mongodb.org/mongo-driver/x/mongo/driver/wiremessage"
)

func TestCleartextSessionAuthenticator(t *testing.T) {
	var dummy *dummyAuthenticator
	auth.RegisterAuthenticatorFactory(
		"dummy",
		func(cred *auth.Cred) (authenticator auth.Authenticator, e error) {
			dummy = &dummyAuthenticator{Cred: cred}
			return dummy, nil
		},
	)

	subject := &CleartextSessionAuthenticator{
		Source:    "db",
		Username:  "user",
		Password:  "pencil",
		Mechanism: "dummy",
	}

	for i := 1; i < 3; i++ {
		t.Run(fmt.Sprintf("%d conversation(s)", i), func(t *testing.T) {
			req := require.New(t)
			conns := make([]driver.Connection, i)
			for j := 0; j < i; j++ {
				conns[j] = &mockConnection{}
			}

			err := subject.Auth(context.Background(), conns)
			req.NoError(err, "unexpected Auth error", i)

			req.Equal("db", dummy.Cred.Source)
			req.Equal("user", dummy.Cred.Username)
			req.Equal("pencil", dummy.Cred.Password)
			req.True(dummy.Cred.PasswordSet)

			req.Equal(i, dummy.InvokedCount)
		})
	}
}

func TestSaslSessionAuthenticator(t *testing.T) {
	t.Run("single step success", func(t *testing.T) {
		for i := 1; i < 3; i++ {
			t.Run(fmt.Sprintf("%d conversation(s)", i), func(t *testing.T) {
				subject := &SaslSessionAuthenticator{
					Source:    "db",
					Username:  "user",
					Mechanism: "SINGLE",
				}

				req := require.New(t)

				conns := make([]driver.Connection, i)
				for j := 0; j < i; j++ {
					subject.AddConversation([]byte("something"), true)

					saslStartReply := &SaslResponse{
						ConversationID: j + 1,
						Done:           true,
						Payload:        []byte{},
						Ok:             1,
					}

					conns[j] = &mockConnection{
						ResponseQ: []*SaslResponse{saslStartReply},
					}
				}

				err := subject.Auth(context.Background(), conns)
				req.NoError(err, "unexpected error")

				for j := 0; j < i; j++ {
					c := conns[j].(*mockConnection)

					req.Equal(1, len(c.Sent), "should have sent 1 message")

					saslStartRequest := bsonutil.NewD(
						bsonutil.NewDocElem("saslStart", int32(1)),
						bsonutil.NewDocElem("mechanism", "SINGLE"),
						bsonutil.NewDocElem("payload", primitive.Binary{Subtype: 0, Data: []byte("something")}),
					)

					req.Equal(saslStartRequest, c.Sent[0], "request message incorrect")
				}
			})
		}
	})

	t.Run("multi step success", func(t *testing.T) {
		for i := 1; i < 3; i++ {
			t.Run(fmt.Sprintf("%d conversation(s)", i), func(t *testing.T) {
				subject := &SaslSessionAuthenticator{
					Source:    "db",
					Username:  "user",
					Mechanism: "MULTI",

					Callback: func(convos []*SaslConversation) error {
						for _, convo := range convos {
							convo.ClientDone = true
							convo.Payload = []byte("second")
						}
						return nil
					},
				}

				req := require.New(t)

				conns := make([]driver.Connection, i)
				for j := 0; j < i; j++ {
					subject.AddConversation([]byte("first"), false)

					saslStartReply := &SaslResponse{
						ConversationID: j + 1,
						Done:           false,
						Payload:        []byte("firstReply"),
						Ok:             1,
					}

					saslContinueReply := &SaslResponse{
						ConversationID: j + 1,
						Done:           true,
						Payload:        []byte("secondReply"),
						Ok:             1,
					}

					conns[j] = &mockConnection{
						ResponseQ: []*SaslResponse{saslStartReply, saslContinueReply},
					}
				}

				err := subject.Auth(context.Background(), conns)
				req.NoError(err, "unexpected error")

				for j := 0; j < i; j++ {
					c := conns[j].(*mockConnection)

					req.Equal(2, len(c.Sent), "should have sent 2 messages")

					expectedCmd := bsonutil.NewD(
						bsonutil.NewDocElem("saslStart", int32(1)),
						bsonutil.NewDocElem("mechanism", "MULTI"),
						bsonutil.NewDocElem("payload", primitive.Binary{Subtype: 0, Data: []byte("first")}),
					)
					req.Equal(expectedCmd, c.Sent[0], "start request message incorrect")

					expectedCmd = bsonutil.NewD(
						bsonutil.NewDocElem("saslContinue", int32(1)),
						bsonutil.NewDocElem("conversationId", int32(j+1)),
						bsonutil.NewDocElem("payload", primitive.Binary{Subtype: 0, Data: []byte("second")}),
					)
					req.Equal(expectedCmd, c.Sent[1], "continue request message incorrect")
				}
			})
		}
	})

	t.Run("failure from server initial", func(t *testing.T) {
		for i := 1; i < 3; i++ {
			t.Run(fmt.Sprintf("%d conversation(s)", i), func(t *testing.T) {
				subject := &SaslSessionAuthenticator{
					Source:    "db",
					Username:  "user",
					Mechanism: "MULTI",
				}

				req := require.New(t)

				conns := make([]driver.Connection, i)
				for j := 0; j < i; j++ {
					subject.AddConversation([]byte("first"), false)

					saslStartReply := &SaslResponse{
						ConversationID: j + 1,
						Code:           143,
						Done:           true,
						Payload:        []byte{},
						Ok:             1,
					}

					conns[j] = &mockConnection{
						ResponseQ: []*SaslResponse{saslStartReply},
					}
				}

				err := subject.Auth(context.Background(), conns)
				req.Error(err)
			})
		}
	})

	t.Run("failure from client in callback", func(t *testing.T) {
		expectedErr := fmt.Errorf("callback error")

		for i := 1; i < 3; i++ {
			t.Run(fmt.Sprintf("%d conversation(s)", i), func(t *testing.T) {
				subject := &SaslSessionAuthenticator{
					Source:    "db",
					Username:  "user",
					Mechanism: "MULTI",

					Callback: func(convos []*SaslConversation) error {
						return expectedErr
					},
				}
				req := require.New(t)

				conns := make([]driver.Connection, i)
				for j := 0; j < i; j++ {
					subject.AddConversation([]byte("first"), false)

					saslStartReply := &SaslResponse{
						ConversationID: j + 1,
						Done:           false,
						Payload:        []byte("firstReply"),
						Ok:             1,
					}

					conns[j] = &mockConnection{
						ResponseQ: []*SaslResponse{saslStartReply},
					}
				}

				err := subject.Auth(context.Background(), conns)
				req.Error(err)
				req.Equal(expectedErr, err)

				for j := 0; j < i; j++ {
					c := conns[j].(*mockConnection)

					req.Equal(1, len(c.Sent), "should have sent 1 message")

					saslStartRequest := bsonutil.NewD(
						bsonutil.NewDocElem("saslStart", int32(1)),
						bsonutil.NewDocElem("mechanism", "MULTI"),
						bsonutil.NewDocElem("payload", primitive.Binary{Subtype: 0, Data: []byte("first")}),
					)

					req.Equal(saslStartRequest, c.Sent[0], "request message incorrect")
				}
			})
		}
	})

	t.Run("failure from server after callback", func(t *testing.T) {
		for i := 1; i < 3; i++ {
			t.Run(fmt.Sprintf("%d conversation(s)", i), func(t *testing.T) {
				subject := &SaslSessionAuthenticator{
					Source:    "db",
					Username:  "user",
					Mechanism: "MULTI",

					Callback: func(convos []*SaslConversation) error {
						for _, convo := range convos {
							convo.ClientDone = true
							convo.Payload = []byte("second")
						}
						return nil
					},
				}

				req := require.New(t)

				conns := make([]driver.Connection, i)
				for j := 0; j < i; j++ {
					subject.AddConversation([]byte("first"), false)

					saslStartReply := &SaslResponse{
						ConversationID: j + 1,
						Done:           false,
						Payload:        []byte("firstReply"),
						Ok:             1,
					}

					saslContinueReply := &SaslResponse{
						ConversationID: j + 1,
						Code:           143,
						Done:           true,
						Payload:        []byte{},
						Ok:             1,
					}

					conns[j] = &mockConnection{
						ResponseQ: []*SaslResponse{saslStartReply, saslContinueReply},
					}
				}

				err := subject.Auth(context.Background(), conns)
				req.Error(err)

				for j := 0; j < i; j++ {
					c := conns[j].(*mockConnection)

					expectedNumSent := 2

					// the first connection will get the 143 code and result in Auth returning
					// an error, so only one message was sent on later connections (saslStart)
					if j > 0 {
						expectedNumSent = 1
					}

					req.Equal(expectedNumSent, len(c.Sent), "should have sent 2 messages")

					expectedCmd := bsonutil.NewD(
						bsonutil.NewDocElem("saslStart", int32(1)),
						bsonutil.NewDocElem("mechanism", "MULTI"),
						bsonutil.NewDocElem("payload", primitive.Binary{Subtype: 0, Data: []byte("first")}),
					)
					req.Equal(expectedCmd, c.Sent[0], "start request message incorrect")

					if j == 0 {
						expectedCmd = bsonutil.NewD(
							bsonutil.NewDocElem("saslContinue", int32(1)),
							bsonutil.NewDocElem("conversationId", int32(j+1)),
							bsonutil.NewDocElem("payload", primitive.Binary{Subtype: 0, Data: []byte("second")}),
						)
						req.Equal(expectedCmd, c.Sent[1], "continue request message incorrect")
					}
				}
			})
		}
	})
}

// dummy auth Authenticator
type dummyAuthenticator struct {
	Cred         *auth.Cred
	InvokedCount int
}

func (a *dummyAuthenticator) Auth(context.Context, description.Server, driver.Connection) error {
	a.InvokedCount++
	return nil
}

// mockConnection implements driver.Connection. We use
// these connections in the tests so that we can mock
// server responses to sasl auth commands. We store the
// bodies of wire messages we write to the server in the
// Sent field. We use the ResponseQ field to send mock
// responses from the server back to the client code.
type mockConnection struct {
	Sent      []bson.D
	ResponseQ []*SaslResponse
}

func (c *mockConnection) WriteWireMessage(_ context.Context, wm []byte) error {
	_, _, _, _, rem, ok := wiremessage.ReadHeader(wm)
	if !ok {
		return errors.New("failed to read header")
	}

	_, rem, ok = wiremessage.ReadQueryFlags(rem)
	if !ok {
		return errors.New("failed to read query flags")
	}

	_, rem, ok = wiremessage.ReadQueryFullCollectionName(rem)
	if !ok {
		return errors.New("failed to read query full collection name")
	}

	_, rem, ok = wiremessage.ReadQueryNumberToSkip(rem)
	if !ok {
		return errors.New("failed to read query number to skip")
	}

	_, rem, ok = wiremessage.ReadQueryNumberToReturn(rem)
	if !ok {
		return errors.New("failed to read query number to return")
	}

	query, _, ok := wiremessage.ReadQueryQuery(rem)
	if !ok {
		return errors.New("failed to read query query")
	}

	sent := bson.D{}
	err := bson.Unmarshal(query, &sent)
	if err != nil {
		return err
	}

	c.Sent = append(c.Sent, sent)
	return nil
}

func (c *mockConnection) ReadWireMessage(ctx context.Context, dst []byte) ([]byte, error) {
	if len(c.ResponseQ) == 0 {
		return nil, fmt.Errorf("no response queued")
	}
	resp := c.ResponseQ[0]
	c.ResponseQ = c.ResponseQ[1:]

	respBytes, err := bson.Marshal(&resp)
	if err != nil {
		return nil, err
	}

	// Header + Flags + CursorID + StartingFrom + NumberReturned + Length of Documents
	b := make([]byte, 0, 16+4+8+4+4+len(respBytes))
	_, b = wiremessage.AppendHeaderStart(b, 0, 0, wiremessage.OpReply)
	b = wiremessage.AppendReplyFlags(b, 0)
	b = wiremessage.AppendReplyCursorID(b, 0)
	b = wiremessage.AppendReplyStartingFrom(b, 0)
	b = wiremessage.AppendReplyNumberReturned(b, 1)
	b = bsoncore.AppendDocument(b, respBytes)

	return b, nil
}

func (c *mockConnection) Description() description.Server {
	return description.Server{}
}

func (c *mockConnection) Close() error {
	return nil
}

func (c *mockConnection) ID() string {
	return ""
}

func (c *mockConnection) Address() address.Address {
	return ""
}
