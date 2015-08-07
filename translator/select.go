package sqleval

import (
	"fmt"
	"github.com/mongodb/mongo-tools/common/log"
	"github.com/siddontang/mixer/sqlparser"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strconv"
)

/*
// ComparisonExpr.Operator
const (
	sqlparser.AST_EQ
	sqlparser.AST_LT
	sqlparser.AST_GT
	sqlparser.AST_LE
	sqlparser.AST_GE
	sqlparser.AST_NE
	sqlparser.AST_NSE
	sqlparser.AST_IN
	sqlparser.AST_NOT_IN
	sqlparser.AST_LIKE
	sqlparser.AST_NOT_LIKE
)

*/
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

/**
* @return (list of expressions that cannot be pushed down, a $where clause to push down, error)
 */
func playWithWhere(where sqlparser.Expr) ([]sqlparser.Expr, bson.M, error) {
	log.Logf(log.DebugLow, "where: %s (type is %T)", sqlparser.String(where), where)

	switch expr := where.(type) {

	case *sqlparser.AndExpr:
		return nil, nil, fmt.Errorf("where can't handle AndExpr type %T", where)

	case *sqlparser.OrExpr:
		return nil, nil, fmt.Errorf("where can't handle OrExpr type %T", where)

	case *sqlparser.NotExpr:
		return nil, nil, fmt.Errorf("where can't handle NotExpr type %T", where)

	case *sqlparser.ParenBoolExpr:
		return nil, nil, fmt.Errorf("where can't handle ParenBoolExpr type %T", where)

	case *sqlparser.ComparisonExpr:

		column, err := getColumnName(expr.Left)
		if err != nil {
			log.Logf(log.DebugLow, "cannot push down (%s) b/c of %s", sqlparser.String(where), err)
			return []sqlparser.Expr{where}, nil, nil
		}

		right, err := getLiteral(expr.Right)
		if err != nil {
			log.Logf(log.DebugLow, "cannot push down (%s) b/c of %s", sqlparser.String(where), err)
			return []sqlparser.Expr{where}, nil, nil
		}

		return nil, bson.M{column: right}, nil

	case *sqlparser.RangeCond:
		return nil, nil, fmt.Errorf("where can't handle RangeCond type %T", where)

	case *sqlparser.NullCheck:
		return nil, nil, fmt.Errorf("where can't handle NullCheck type %T", where)

	case *sqlparser.ExistsExpr:
		return nil, nil, fmt.Errorf("where can't handle ExistsExpr type %T", where)

	default:
		log.Logf(log.DebugLow, "where can't handle expression type %T", where)
		return nil, nil, fmt.Errorf("where can't handle expression type %T", where)
	}

}

func (e *Evalulator) EvalSelect(db string, sql string, stmt *sqlparser.Select) ([]string, [][]interface{}, error) {
	if stmt == nil {
		// we can parse ourselves
		raw, err := sqlparser.Parse(sql)
		if err != nil {
			return nil, nil, err
		}
		stmt = raw.(*sqlparser.Select)
	}

	log.Logf(log.DebugLow, "parsed stmt: %#v", stmt.Where.Expr)

	if len(stmt.From) == 0 {
		return nil, nil, fmt.Errorf("no table selected")
	}
	if len(stmt.From) > 1 {
		return nil, nil, fmt.Errorf("joins not supported yet")
	}

	var whereToEvaluate []sqlparser.Expr
	var whereToPush bson.M = nil

	if stmt.Where != nil {
		toEval, toPush, err := playWithWhere(stmt.Where.Expr)
		if err != nil {
			return nil, nil, err
		}
		whereToEvaluate = toEval
		whereToPush = toPush
		log.Logf(log.DebugLow, "toEval: %v toPush: %v", whereToEvaluate, whereToPush)
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

	var iter *mgo.Iter
	if tableConfig.Pipeline == nil {
		query := collection.Find(whereToPush)
		iter = query.Iter()
	} else {
		thePipe := tableConfig.Pipeline
		if whereToPush != nil {
			thePipe = append(thePipe, bson.M{"$match": whereToPush})
		}
		pipe := collection.Pipe(thePipe)
		iter = pipe.Iter()
	}

	return IterToNamesAndValues(iter)
}
