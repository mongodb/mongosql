package translator

import (
	"fmt"
	"github.com/erh/mongo-sql-temp/config"
	"github.com/mongodb/mongo-tools/common/log"
	"github.com/erh/mixer/sqlparser"
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

func (e *Evalulator) getCollection(session *mgo.Session, tableConfig *config.TableConfig) DataSource {
	fullName := tableConfig.Collection
	pcs := strings.SplitN(fullName, ".", 2)
	return MgoDataSource{session.DB(pcs[0]).C(pcs[1]), tableConfig.Columns}
}

func (e *Evalulator) getDataSource(db string, tableName string) (DataSource, error) {

	dbConfig := e.cfg.Schemas[db]
	if dbConfig == nil {
		if strings.ToLower(db) == "information_schema" {
			if strings.ToLower(tableName) == "columns" {
				return ConfigDataSource{e.cfg, true}, nil
			}
			if strings.ToLower(tableName) == "tables" {
				return ConfigDataSource{e.cfg, false}, nil
			}

		}
		
		return nil, fmt.Errorf("db (%s) does not exist", db)
	}

	tableConfig := dbConfig.Tables[tableName]
	if tableConfig == nil {
		return nil, fmt.Errorf("table (%s) does not exist in db(%s)", tableName, db)
	}

	session := e.getSession()
	return e.getCollection(session, tableConfig), nil
}

// EvalSelect needs to be updated ...
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

	ctx, err := NewParseCtx(stmt)
	if err != nil {
		return nil, nil, fmt.Errorf("error constructing new parse context: %v", err)
	}

	if err = algebrizeStatement(stmt, ctx); err != nil {
		return nil, nil, fmt.Errorf("error algebrizing select statement: %v", err)
	}

	log.Logf(log.DebugLow, "algebrized tree: %#v", stmt)

	// TODO: full query planner
	var query interface{} = nil

	if stmt.Where != nil {

		log.Logf(log.DebugLow, "where: %s (type is %T)", sqlparser.String(stmt.Where.Expr), stmt.Where.Expr)

		var err error

		query, err = translateExpr(stmt.Where.Expr)
		if err != nil {
			return nil, nil, err
		}
	}

	alias := ctx.GetDefaultTable()
	tableName := ctx.TableName(alias)

	if strings.Index(tableName, ".") >= 0 {
		split := strings.SplitN(tableName, ".", 2)
		db = split[0]
		tableName = split[1]
	}
	
	collection, err := e.getDataSource(db, tableName)
	if err != nil {
		return nil, nil, err
	}

	result := collection.Find(query)
	iter := result.Iter()

	return IterToNamesAndValues(iter, collection.GetColumns())
}
