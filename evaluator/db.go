package evaluator

import (
	"github.com/10gen/sqlproxy/schema"
	"gopkg.in/mgo.v2"
	"strings"
)

type SessionProvider struct {
	cfg           *schema.Schema
	globalSession *mgo.Session
}

func NewSessionProvider(cfg *schema.Schema) (*SessionProvider, error) {
	e := new(SessionProvider)
	e.cfg = cfg

	session, err := mgo.Dial(cfg.Url)
	if err != nil {
		return nil, err
	}
	e.globalSession = session

	return e, nil
}

func (e *SessionProvider) GetSession() *mgo.Session {
	if e.globalSession == nil {
		panic("No global session has been set")
	}
	return e.globalSession.Copy()
}

func (e *SessionProvider) Namespace(session *mgo.Session, fullName string) *mgo.Collection {
	pcs := strings.SplitN(fullName, ".", 2)
	return session.DB(pcs[0]).C(pcs[1])
}
