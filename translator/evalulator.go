package translator

import (
	"fmt"
	"github.com/erh/mongo-sql-temp/config"
	"github.com/mongodb/mongo-tools/common/log"
	"github.com/siddontang/mixer/sqlparser"
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
	if e.globalSession == nil {
		panic("No global session has been set")
	}
	return e.globalSession.Copy()
}

func (e *Evalulator) getCollection(session *mgo.Session, fullName string) *mgo.Collection {
	pcs := strings.SplitN(fullName, ".", 2)
	return session.DB(pcs[0]).C(pcs[1])
}

// EvalSelect needs to be updated ...
// TODO: handle SelectExprs => []SelectExpr -> StarExpr and NonStarExpr.
func (e *Evalulator) EvalSelect(db string, sql string, stmt *sqlparser.Select) ([]string, [][]interface{}, error) {
	if stmt == nil {
		// we can parse ourselves
		raw, err := sqlparser.Parse(sql)
		if err != nil {
			return nil, nil, err
		}
		stmt = raw.(*sqlparser.Select)
	}

	// handle select expressions like as aliasing
	// e.g. select FirstName as f, LastName as l from foo;
	for i, expr := range stmt.SelectExprs {
		log.Logf(log.DebugLow, "handling parsed select expr %v: %#v", i, expr)

	}

	if stmt.From != nil {
		var from interface{} = nil

		var err error

		for i, expr := range stmt.From {

			log.Logf(log.DebugLow, "from (%d): %s (type is %T)", i, sqlparser.String(expr), expr)

			from, err = translateTableExpr(expr)
			if err != nil {
				return nil, nil, err
			}

			log.Logf(log.DebugLow, "from %v translation: %#v", i, from)

		}

	}

	var query interface{} = nil

	if stmt.Where != nil {

		log.Logf(log.DebugLow, "where: %s (type is %T)", sqlparser.String(stmt.Where.Expr), stmt.Where.Expr)

		var err error

		query, err = translateExpr(stmt.Where.Expr)
		if err != nil {
			return nil, nil, err
		}
	}

	if stmt.Having != nil {
		return nil, nil, fmt.Errorf("'HAVING' statement not yet supported")
	}

	tableName := sqlparser.String(stmt.From[0])
	dbConfig := e.cfg.Schemas[db]
	if dbConfig == nil {
		return nil, nil, fmt.Errorf("db (%s) does not exist", db)
	}

	tableConfig := dbConfig.Tables[tableName]
	if tableConfig == nil {
		return nil, nil, fmt.Errorf("table (%s) does not exist in db(%s)", tableName, db)
	}

	session := e.getSession()
	collection := e.getCollection(session, tableConfig.Collection)

	log.Logf(log.DebugLow, "query: %#v", query)
	result := collection.Find(query)
	iter := result.Iter()

	return IterToNamesAndValues(iter)
}
