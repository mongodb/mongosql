package translator

import (
	"fmt"
	"github.com/mongodb/mongo-tools/common/log"
	"github.com/siddontang/mixer/sqlparser"
	"gopkg.in/mgo.v2/bson"
	"strconv"
)

// translateExpr takes an expression and returns its translated form.
func translateExpr(where sqlparser.Expr) (interface{}, error) {
	log.Logf(log.DebugLow, "where: %s (type is %T)", sqlparser.String(where), where)

	switch expr := where.(type) {

	case sqlparser.StrVal:
		return nil, fmt.Errorf("where can't handle StrVal type %T", where)

	case sqlparser.NumVal:
		val, err := getNumVal(expr)
		if err != nil {
			return nil, fmt.Errorf("where can't handle NumVal %v: %v", err)
		}
		return val, err

	case sqlparser.ValArg:
		return nil, fmt.Errorf("where can't handle ValArg type %T", where)

	case sqlparser.ValTuple:
		return nil, fmt.Errorf("where can't handle ValTuple type %T", where)

	case *sqlparser.NullVal:
		return nil, fmt.Errorf("where can't handle NullVal type %T", where)

	case *sqlparser.ColName:
		return sqlparser.String(expr), nil

	case *sqlparser.Subquery:
		return nil, fmt.Errorf("where can't handle Subquery type %T", where)

	case *sqlparser.BinaryExpr:
		return nil, fmt.Errorf("where can't handle BinaryExpr type %T", where)

	case *sqlparser.UnaryExpr:
		return nil, fmt.Errorf("where can't handle UnaryExpr type %T", where)

	case *sqlparser.FuncExpr:
		return nil, fmt.Errorf("where can't handle FuncExpr type %T", where)

	case *sqlparser.CaseExpr:
		return nil, fmt.Errorf("where can't handle CaseExpr type %T", where)

	case *sqlparser.AndExpr:
		left, right, err := translateLRExpr(expr.Left, expr.Right)
		if err != nil {
			return nil, fmt.Errorf("AndExpr error: %v", err)
		}
		return bson.M{"$and": []interface{}{left, right}}, nil
	case *sqlparser.OrExpr:
		return nil, fmt.Errorf("where can't handle OrExpr type %T", where)

	case *sqlparser.NotExpr:
		return nil, fmt.Errorf("where can't handle NotExpr type %T", where)

	case *sqlparser.ParenBoolExpr:
		return nil, fmt.Errorf("where can't handle ParenBoolExpr type %T", where)

	case *sqlparser.ComparisonExpr:
		left, right, err := translateLRExpr(expr.Left, expr.Right)
		if err != nil {
			return nil, fmt.Errorf("ComparisonExpr error: %v", err)
		}
		switch tLeft := left.(type) {
		case string:
			return bson.M{tLeft: bson.M{operators[expr.Operator]: right}}, nil
		default:
			return bson.M{operators[expr.Operator]: []interface{}{left, right}}, nil
		}
	case *sqlparser.RangeCond:
		return nil, fmt.Errorf("where can't handle RangeCond type %T", where)

	case *sqlparser.NullCheck:
		return nil, fmt.Errorf("where can't handle NullCheck type %T", where)

	case *sqlparser.ExistsExpr:
		return nil, fmt.Errorf("where can't handle ExistsExpr type %T", where)

	default:
		return nil, fmt.Errorf("where can't handle expression type %T", where)
	}

}

// translateLRExpr takes two leaf expressions and returns their translations.
func translateLRExpr(lExpr, rExpr sqlparser.Expr) (interface{}, interface{}, error) {
	left, err := translateExpr(lExpr)
	if err != nil {
		return nil, nil, fmt.Errorf("lExpr error: %v", err)
	}
	right, err := translateExpr(rExpr)
	if err != nil {
		return nil, nil, fmt.Errorf("rExpr error: %v", err)
	}
	return left, right, nil
}

// getNumVal takes a number value expression and returns a converted form of it.
func getNumVal(valExpr sqlparser.ValExpr) (interface{}, error) {
	switch val := valExpr.(type) {
	case sqlparser.StrVal:
		return sqlparser.String(val), nil
	case sqlparser.NumVal:
		// TODO add other types
		f, err := strconv.ParseFloat(sqlparser.String(val), 64)
		if err != nil {
			return nil, err
		}
		return f, nil
	default:
		return nil, fmt.Errorf("not a literal type: %T", valExpr)
	}
}
