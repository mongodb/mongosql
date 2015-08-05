package sqleval

import (
	"github.com/erh/mongo-sql-temp/config"
	"gopkg.in/mgo.v2"
	"strings"
)

type Evalulator struct {
	cfg           *config.Config
	globalSession *mgo.Session
}

func NewEvalulator(cfg *config.Config) (*Evalulator, error) {
	e := new(Evalulator)
	e.cfg = cfg

	session, err := mgo.Dial(cfg.Url)
	if err != nil {
		return nil, err
	}
	e.globalSession = session

	return e, nil
}

func (e *Evalulator) getSession() *mgo.Session {
	return e.globalSession.New()
}

func (e *Evalulator) getCollection(session *mgo.Session, fullName string) *mgo.Collection {
	pcs := strings.SplitN(fullName, ".", 2)
	return session.DB(pcs[0]).C(pcs[1])
}
