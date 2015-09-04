package planner

import (
	"github.com/erh/mongo-sql-temp/config"
	"gopkg.in/mgo.v2"
	"strings"
)

type SessionProvider struct {
	cfg           *config.Config
	globalSession *mgo.Session
}

func NewSessionProvider(cfg *config.Config) (*SessionProvider, error) {
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
