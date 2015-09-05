package translator

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"github.com/erh/mongo-sql-temp/config"
	"github.com/erh/mongo-sql-temp/translator/algebrizer"
	"github.com/erh/mongo-sql-temp/translator/evaluator"
	"github.com/erh/mongo-sql-temp/translator/planner"
	"github.com/mongodb/mongo-tools/common/log"
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

func (e *Evalulator) getCollection(session *mgo.Session, tableConfig *config.TableConfig) planner.DataSource {
	fullName := tableConfig.Collection
	pcs := strings.SplitN(fullName, ".", 2)
	return planner.MgoDataSource{session.DB(pcs[0]).C(pcs[1]), tableConfig.Columns}
}

func (e *Evalulator) getDataSource(db string, tableName string) (planner.DataSource, error) {

	dbConfig := e.cfg.Schemas[db]
	if dbConfig == nil {
		if strings.ToLower(db) == "information_schema" {
			if strings.ToLower(tableName) == "columns" {
				return planner.ConfigDataSource{e.cfg, true}, nil
			}
			if strings.ToLower(tableName) == "tables" ||
				strings.ToLower(tableName) == "txxxables" {
				return planner.ConfigDataSource{e.cfg, false}, nil
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

func computeOutputColumns(schema []config.Column, exprs sqlparser.SelectExprs) ([]config.Column, bool, error) {
	output := make([]config.Column, 0)

	hadStar := false

	for _, c := range exprs {
		switch entry := c.(type) {
		case *sqlparser.StarExpr:
			output = append(output, schema...)
			hadStar = true
		case *sqlparser.NonStarExpr:
			if len(entry.As) > 0 {
				return output, false, fmt.Errorf("can't handle AS yet")
			}

			switch e := entry.Expr.(type) {
			case *sqlparser.ColName:
				name := string(e.Name)
				var idx int = -1
				for i, j := range schema {
					if caseInsensitiveEquals(j.Name, name) {
						idx = i
						break
					}
				}
				if idx == -1 {
					output = append(output, config.Column{name, "", ""})
				} else {
					output = append(output, schema[idx])
				}
			default:
				return output, false, fmt.Errorf("weird non-star type %T", e)
			}
		default:
			return output, false, fmt.Errorf("weird select type %T", entry)
		}
	}

	return output, hadStar, nil
}

// EvalSelect returns all rows matching the query.
func (e *Evalulator) EvalSelect(db string, sql string, stmt *sqlparser.Select) ([]string, [][]interface{}, error) {
	log.Logf(log.DebugLow, "Evaluating select: %#v", stmt)

	if stmt == nil {
		// we can parse ourselves
		raw, err := sqlparser.Parse(sql)
		if err != nil {
			return nil, nil, err
		}
		var ok bool
		if stmt, ok = raw.(*sqlparser.Select); !ok {
			return nil, nil, fmt.Errorf("got a non-select statement in EvalSelect")
		}
	}

	// create initial parse context
	pCtx, err := algebrizer.NewParseCtx(stmt)
	if err != nil {
		return nil, nil, fmt.Errorf("error constructing new parse context: %v", err)
	}

	// resolve names
	if err = algebrizer.AlgebrizeStatement(stmt, pCtx); err != nil {
		return nil, nil, fmt.Errorf("error algebrizing select statement: %v", err)
	}
	
	// construct plan
	queryPlan, err := planner.PlanQuery(stmt)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting query plan: %v", err)
	}

	eCtx := &planner.ExecutionCtx{
		Config: e.cfg,
		Db:     db,
	}

	// execute plan
	columns, results, err := evaluator.Execute(eCtx, queryPlan)
	if err != nil {
		return nil, nil, fmt.Errorf("error executing query: %v", err)
	}

	return columns, results, nil
}
