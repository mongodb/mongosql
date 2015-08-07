package translator

import (
	"fmt"
	"github.com/mongodb/mongo-tools/common/log"
	"github.com/siddontang/mixer/sqlparser"
	"gopkg.in/mgo.v2/bson"
	"strconv"
)

// EvalSelect returns ...
func (e *Evalulator) EvalSelect(db string, sql string, stmt *sqlparser.Select) ([]string, [][]interface{}, error) {
	if stmt == nil {
		// we can parse ourselves
		raw, err := sqlparser.Parse(sql)
		if err != nil {
			return nil, nil, err
		}
		stmt = raw.(*sqlparser.Select)
	}

	if len(stmt.From) == 0 {
		return nil, nil, fmt.Errorf("no table selected")
	}
	if len(stmt.From) > 1 {
		return nil, nil, fmt.Errorf("joins not supported yet")
	}

	var query bson.M = nil

	if stmt.Where != nil {

		log.Logf(log.DebugLow, "parsed stmt: %#v", stmt.Where.Expr)

		query, err := translateWhere(stmt.Where.Expr)
		if err != nil {
			return nil, nil, err
		}
		log.Logf(log.DebugLow, "query: %v", query)
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

	result := collection.Find(query)
	iter := result.Iter()

	return IterToNamesAndValues(iter)
}

/**
* @return (list of expressions that cannot be pushed down, a $where clause to push down, error)
 */
func translateWhere(where sqlparser.Expr) (bson.M, error) {
	log.Logf(log.DebugLow, "where: %s (type is %T)", sqlparser.String(where), where)

	switch expr := where.(type) {

	case *sqlparser.AndExpr:
		return nil, fmt.Errorf("where can't handle AndExpr type %T", where)

	case *sqlparser.OrExpr:
		return nil, fmt.Errorf("where can't handle OrExpr type %T", where)

	case *sqlparser.NotExpr:
		return nil, fmt.Errorf("where can't handle NotExpr type %T", where)

	case *sqlparser.ParenBoolExpr:
		return nil, fmt.Errorf("where can't handle ParenBoolExpr type %T", where)

	case *sqlparser.ComparisonExpr:

		column, err := getColumnName(expr.Left)
		if err != nil {
			log.Logf(log.DebugLow, "cannot push down (%s) b/c of %s", sqlparser.String(where), err)
			return nil, nil
		}

		right, err := getLiteral(expr.Right)
		if err != nil {
			log.Logf(log.DebugLow, "cannot push down (%s) b/c of %s", sqlparser.String(where), err)
			return nil, nil
		}

		return bson.M{column: right}, nil

	case *sqlparser.RangeCond:
		return nil, fmt.Errorf("where can't handle RangeCond type %T", where)

	case *sqlparser.NullCheck:
		return nil, fmt.Errorf("where can't handle NullCheck type %T", where)

	case *sqlparser.ExistsExpr:
		return nil, fmt.Errorf("where can't handle ExistsExpr type %T", where)

	default:
		log.Logf(log.DebugLow, "where can't handle expression type %T", where)
		return nil, fmt.Errorf("where can't handle expression type %T", where)
	}

}

func getColumnName(valExpr sqlparser.ValExpr) (string, error) {
	switch val := valExpr.(type) {
	case *sqlparser.ColName:
		return sqlparser.String(val), nil
	default:
		return "", fmt.Errorf("not a column name type: %T", valExpr)
	}
}

func getLiteral(valExpr sqlparser.ValExpr) (interface{}, error) {
	switch val := valExpr.(type) {
	case sqlparser.StrVal:
		return sqlparser.String(val), nil
	case sqlparser.NumVal:
		f, err := strconv.ParseFloat(sqlparser.String(val), 64)
		if err != nil {
			return nil, err
		}
		return f, nil
	default:
		return nil, fmt.Errorf("not a literal type: %T", valExpr)
	}
}
