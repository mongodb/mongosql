package translator

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"github.com/erh/mongo-sql-temp/config"
	"github.com/erh/mongo-sql-temp/translator/algebrizer"
	"github.com/erh/mongo-sql-temp/translator/mongodb"
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
			if strings.ToLower(tableName) == "tables" ||
				strings.ToLower(tableName) == "txxxables" {
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

// EvalSelect needs to be updated ...
func (e *Evalulator) EvalSelect(db string, sql string, stmt *sqlparser.Select) ([]string, [][]interface{}, error) {
	if stmt == nil {
		// we can parse ourselves
		raw, err := ParseSQL(sql)
		if err != nil {
			return nil, nil, err
		}
		var ok bool
		if stmt, ok = raw.(*sqlparser.Select); !ok {
			return nil, nil, fmt.Errorf("got a non-select statement in EvalSelect: %T", raw)
		}
	}

	ctx, err := algebrizer.NewParseCtx(stmt)
	if err != nil {
		return nil, nil, fmt.Errorf("error constructing new parse context: %v", err)
	}

	if err = algebrizer.AlgebrizeStatement(stmt, ctx); err != nil {
		return nil, nil, fmt.Errorf("error algebrizing select statement: %v", err)
	}

	log.Logf(log.DebugLow, "algebrized tree: %#v", stmt)

	// TODO: full query planner
	var query interface{} = nil

	if stmt.Where != nil {

		log.Logf(log.DebugLow, "where: %s (type is %T)", sqlparser.String(stmt.Where.Expr), stmt.Where.Expr)

		var err error

		query, err = mongodb.TranslateExpr(stmt.Where.Expr)
		if err != nil {
			return nil, nil, err
		}
	}

	tableName, err := ctx.GetCurrentTable("")
	if err != nil {
		return nil, nil, err
	}

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

	outputColumns, includeExtra, err := computeOutputColumns(collection.GetColumns(), stmt.SelectExprs)
	if err != nil {
		return nil, nil, err
	}

	return IterToNamesAndValues(iter, outputColumns, includeExtra)
}
