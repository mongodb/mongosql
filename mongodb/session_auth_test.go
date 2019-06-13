package mongodb_test

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/10gen/sqlproxy/internal/astutil"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	. "github.com/10gen/sqlproxy/mongodb"

	oldbson "github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/mongo-go-driver/mongo/model"
	"github.com/10gen/mongo-go-driver/mongo/private/auth"
	"github.com/10gen/mongo-go-driver/mongo/private/conn"
	"github.com/10gen/mongo-go-driver/mongo/private/msg"
	. "github.com/smartystreets/goconvey/convey"

	"go.mongodb.org/mongo-driver/bson"
)

func TestCleartextSessionAuthenticator(t *testing.T) {
	Convey("Subject: Cleartext Session Authenticator", t, func() {
		var dummy *dummyAuthenticator
		auth.RegisterAuthenticatorFactory(
			"dummy",
			func(cred *auth.Cred) (auth.Authenticator, error) {
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
			Convey(fmt.Sprintf("%d conversation(s)", i), func() {
				var conns []conn.Connection
				for j := 0; j < i; j++ {
					conns = append(conns, &mockConnection{})
				}

				So(subject.Auth(context.Background(), conns), ShouldBeNil)

				So(dummy.Cred.Source, ShouldEqual, "db")
				So(dummy.Cred.Username, ShouldEqual, "user")
				So(dummy.Cred.Password, ShouldEqual, "pencil")
				So(dummy.Cred.PasswordSet, ShouldBeTrue)

				So(dummy.InvokedCount, ShouldEqual, i)
			})
		}
	})
}

func TestSaslSessionAuthenticator(t *testing.T) {
	Convey("Subject: Sasl Session Authenticator", t, func() {

		Convey("Single Step Success", func() {
			subject := &SaslSessionAuthenticator{
				Source:    "db",
				Username:  "user",
				Mechanism: "SINGLE",
			}

			for i := 1; i < 3; i++ {
				Convey(fmt.Sprintf("%d conversation(s)", i), func() {
					var conns []conn.Connection
					for j := 0; j < i; j++ {
						subject.AddConversation([]byte("something"), true)

						saslStartReply := createCommandReply(bsonutil.NewD(
							bsonutil.NewDocElem("ok", 1),
							bsonutil.NewDocElem("conversationId", j+1),
							bsonutil.NewDocElem("payload", []byte{}),
							bsonutil.NewDocElem("done", true),
						))
						conns = append(conns, &mockConnection{
							ResponseQ: []*msg.Reply{saslStartReply},
						})
					}

					err := subject.Auth(context.Background(), conns)
					So(err, ShouldBeNil)

					for j := 0; j < i; j++ {
						conn := conns[j].(*mockConnection)

						So(len(conn.Sent), ShouldEqual, 1)

						saslStartRequest := conn.Sent[0].(*msg.Query)
						expectedCmd := astutil.NewToOldBSOND(bsonutil.NewD(
							bsonutil.NewDocElem("saslStart", 1),
							bsonutil.NewDocElem("mechanism", "SINGLE"),
							bsonutil.NewDocElem("payload", []byte("something")),
						))

						So(saslStartRequest.Query, ShouldResemble, expectedCmd)
					}
				})
			}
		})

		Convey("Multi Step Success", func() {
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

			for i := 1; i < 3; i++ {
				Convey(fmt.Sprintf("%d conversation(s)", i), func() {
					var conns []conn.Connection
					for j := 0; j < i; j++ {
						subject.AddConversation([]byte("first"), false)

						saslStartReply := createCommandReply(bsonutil.NewD(
							bsonutil.NewDocElem("ok", 1),
							bsonutil.NewDocElem("conversationId", j+1),
							bsonutil.NewDocElem("payload", []byte("firstReply")),
							bsonutil.NewDocElem("done", false),
						))
						saslContinueReply := createCommandReply(bsonutil.NewD(
							bsonutil.NewDocElem("ok", 1),
							bsonutil.NewDocElem("conversationId", j+1),
							bsonutil.NewDocElem("payload", []byte("secondReply")),
							bsonutil.NewDocElem("done", true),
						))
						conns = append(conns, &mockConnection{
							ResponseQ: []*msg.Reply{saslStartReply, saslContinueReply},
						})
					}

					err := subject.Auth(context.Background(), conns)
					So(err, ShouldBeNil)

					for j := 0; j < i; j++ {
						conn := conns[j].(*mockConnection)

						So(len(conn.Sent), ShouldEqual, 2)

						saslStartRequest := conn.Sent[0].(*msg.Query)
						expectedCmd := astutil.NewToOldBSOND(bsonutil.NewD(
							bsonutil.NewDocElem("saslStart", 1),
							bsonutil.NewDocElem("mechanism", "MULTI"),
							bsonutil.NewDocElem("payload", []byte("first")),
						))

						So(saslStartRequest.Query, ShouldResemble, expectedCmd)
						saslContinueRequest := conn.Sent[1].(*msg.Query)
						expectedCmd = astutil.NewToOldBSOND(bsonutil.NewD(
							bsonutil.NewDocElem("saslContinue", 1),
							bsonutil.NewDocElem("conversationId", j+1),
							bsonutil.NewDocElem("payload", []byte("second")),
						))

						So(saslContinueRequest.Query, ShouldResemble, expectedCmd)
					}
				})
			}
		})

		Convey("Failure from Server Initial", func() {
			subject := &SaslSessionAuthenticator{
				Source:    "db",
				Username:  "user",
				Mechanism: "MULTI",
			}

			for i := 1; i < 3; i++ {
				Convey(fmt.Sprintf("%d conversation(s)", i), func() {
					var conns []conn.Connection
					for j := 0; j < i; j++ {
						subject.AddConversation([]byte("first"), false)

						saslStartReply := createCommandReply(bsonutil.NewD(
							bsonutil.NewDocElem("ok", 1),
							bsonutil.NewDocElem("conversationId", j+1),
							bsonutil.NewDocElem("payload", []byte{}),
							bsonutil.NewDocElem("code", 143),
							bsonutil.NewDocElem("done", true),
						))
						conns = append(conns, &mockConnection{
							ResponseQ: []*msg.Reply{saslStartReply},
						})
					}

					err := subject.Auth(context.Background(), conns)
					So(err, ShouldNotBeNil)
				})
			}
		})

		Convey("Failure from Client in Callback", func() {
			subject := &SaslSessionAuthenticator{
				Source:    "db",
				Username:  "user",
				Mechanism: "MULTI",

				Callback: func(convos []*SaslConversation) error {
					return fmt.Errorf("error")
				},
			}

			for i := 1; i < 3; i++ {
				Convey(fmt.Sprintf("%d conversation(s)", i), func() {
					var conns []conn.Connection
					for j := 0; j < i; j++ {
						subject.AddConversation([]byte("first"), false)

						saslStartReply := createCommandReply(bsonutil.NewD(
							bsonutil.NewDocElem("ok", 1),
							bsonutil.NewDocElem("conversationId", j+1),
							bsonutil.NewDocElem("payload", []byte("firstReply")),
							bsonutil.NewDocElem("done", false),
						))
						saslContinueReply := createCommandReply(bsonutil.NewD(
							bsonutil.NewDocElem("ok", 1),
							bsonutil.NewDocElem("conversationId", j+1),
							bsonutil.NewDocElem("payload", []byte{}),
							bsonutil.NewDocElem("code", 143),
							bsonutil.NewDocElem("done", true),
						))
						conns = append(conns, &mockConnection{
							ResponseQ: []*msg.Reply{saslStartReply, saslContinueReply},
						})
					}

					err := subject.Auth(context.Background(), conns)
					So(err, ShouldNotBeNil)
				})
			}
		})

		Convey("Failure from Server After Callback", func() {
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

			for i := 1; i < 3; i++ {
				Convey(fmt.Sprintf("%d conversation(s)", i), func() {
					var conns []conn.Connection
					for j := 0; j < i; j++ {
						subject.AddConversation([]byte("first"), false)

						saslStartReply := createCommandReply(bsonutil.NewD(
							bsonutil.NewDocElem("ok", 1),
							bsonutil.NewDocElem("conversationId", j+1),
							bsonutil.NewDocElem("payload", []byte{}),
							bsonutil.NewDocElem("code", 143),
							bsonutil.NewDocElem("done", true),
						))
						conns = append(conns, &mockConnection{
							ResponseQ: []*msg.Reply{saslStartReply},
						})
					}

					err := subject.Auth(context.Background(), conns)
					So(err, ShouldNotBeNil)
				})
			}
		})
	})
}

// dummy auth Authenticator
type dummyAuthenticator struct {
	Cred         *auth.Cred
	InvokedCount int
}

func (a *dummyAuthenticator) Auth(context.Context, conn.Connection) error {
	a.InvokedCount++
	return nil
}

type mockConnection struct {
	Dead      bool
	Sent      []msg.Request
	ResponseQ []*msg.Reply
	WriteErr  error

	SkipResponseToFixup bool
}

func (c *mockConnection) Alive() bool {
	return !c.Dead
}

func (c *mockConnection) Close() error {
	c.Dead = true
	return nil
}

func (c *mockConnection) CloseIgnoreError() {
	c.Close()
}

func (c *mockConnection) Expired() bool {
	return c.Dead
}

func (c *mockConnection) MarkDead() {
	c.Dead = true
}

func (c *mockConnection) LocalAddr() net.Addr {
	return nil
}

func (c *mockConnection) Model() *model.Conn {
	return &model.Conn{}
}

func (c *mockConnection) Read(ctx context.Context, responseTo int32) (msg.Response, error) {
	if len(c.ResponseQ) == 0 {
		return nil, fmt.Errorf("no response queued")
	}
	resp := c.ResponseQ[0]
	c.ResponseQ = c.ResponseQ[1:]
	return resp, nil
}

func (c *mockConnection) Write(ctx context.Context, reqs ...msg.Request) error {
	if c.WriteErr != nil {
		err := c.WriteErr
		c.WriteErr = nil
		return err
	}

	for i, req := range reqs {
		c.Sent = append(c.Sent, req)
		if !c.SkipResponseToFixup && i < len(c.ResponseQ) {
			c.ResponseQ[i].RespTo = req.RequestID()
		}
	}
	return nil
}

func createCommandReply(cmd bson.D) *msg.Reply {
	doc, _ := oldbson.Marshal(astutil.NewToOldBSOND(cmd))
	reply := &msg.Reply{
		NumberReturned: 1,
		DocumentsBytes: doc,
	}

	// encode it, then decode it to handle the internal workings of msg.Reply
	codec := msg.NewWireProtocolCodec()
	var b bytes.Buffer
	err := codec.Encode(&b, reply)
	if err != nil {
		panic(err)
	}
	resp, err := codec.Decode(&b)
	if err != nil {
		panic(err)
	}

	return resp.(*msg.Reply)
}
