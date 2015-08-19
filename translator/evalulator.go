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
		var ok bool
		if stmt, ok = raw.(*sqlparser.Select); !ok {
			return nil, nil, fmt.Errorf("got a non-selecdt statement in EvalSelect")
		}
	}

	algebrizedQuery, err := getAlgebrizedQuery(stmt, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error constructing algebrized select tree: %v", err)
	}

	log.Logf(log.DebugLow, "algebrized tree: %#v", algebrizedQuery)

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

	result := collection.Find(algebrizedQuery)
	iter := result.Iter()

	return IterToNamesAndValues(iter)
}

func getAlgebrizedQuery(stmt *sqlparser.Select, pCtx *ParseCtx) (interface{}, error) {
	// handle select expressions like as aliasing
	// e.g. select FirstName as f, LastName as l from foo;
	columns, err := getColumnInfo(stmt.SelectExprs)
	if err != nil {
		return nil, err
	}

	log.Logf(log.DebugLow, "columns %#v", columns)

	tables, err := getTableInfo(stmt.From)
	if err != nil {
		return nil, err
	}

	log.Logf(log.DebugLow, "tables %#v", columns)

	ctx := &ParseCtx{
		Table:  tables,
		Column: columns,
		Parent: pCtx,
	}

	var query interface{} = nil

	if stmt.Where != nil {

		log.Logf(log.DebugLow, "where: %s (type is %T)", sqlparser.String(stmt.Where.Expr), stmt.Where.Expr)

		var err error

		query, err = translateExpr(stmt.Where.Expr, ctx)
		if err != nil {
			return nil, err
		}

		log.Logf(log.DebugLow, "query tree: %#v", query)

		return query, nil
	}

	if stmt.Having != nil {
		return nil, fmt.Errorf("'HAVING' statement not yet supported")
	}

	return query, nil
}
