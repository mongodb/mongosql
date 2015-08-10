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

	case sqlparser.NumVal:
		val, err := getNumVal(expr)
		if err != nil {
			return nil, fmt.Errorf("where can't handle NumVal %v: %v", err)
		}
		return val, err

	case sqlparser.ValArg:
		return nil, fmt.Errorf("where can't handle ValArg type %T", where)

	case sqlparser.ValTuple:
		vals := sqlparser.ValExprs(expr)
		tuples := make([]interface{}, len(vals))
		var err error
		for i, val := range vals {
			tuples[i], err = translateExpr(val)
			if err != nil {
				return nil, fmt.Errorf("where can't handle ValExpr %v: %v", val, err)
			}
		}
		return tuples, nil

	case *sqlparser.NullVal:
		return nil, nil

	case sqlparser.StrVal, *sqlparser.ColName:
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
		return bson.M{MgoAnd: []interface{}{left, right}}, nil

	case *sqlparser.OrExpr:
		left, right, err := translateLRExpr(expr.Left, expr.Right)
		if err != nil {
			return nil, fmt.Errorf("OrExpr error: %v", err)
		}
		return bson.M{MgoOr: []interface{}{left, right}}, nil

	case *sqlparser.ComparisonExpr:
		left, right, err := translateLRExpr(expr.Left, expr.Right)
		if err != nil {
			return nil, fmt.Errorf("ComparisonExpr error: %v", err)
		}
		switch tLeft := left.(type) {
		case string:
			return bson.M{tLeft: bson.M{oprtMap[expr.Operator]: right}}, nil
		default:
			return bson.M{oprtMap[expr.Operator]: []interface{}{left, right}}, nil
		}

	case *sqlparser.RangeCond:
		from, to, err := translateLRExpr(expr.From, expr.To)
		if err != nil {
			return nil, fmt.Errorf("RangeCond lr error: %v", err)
		}
		left, err := translateExpr(expr.Left)
		if err != nil {
			return nil, fmt.Errorf("RangeCond left error: %v", err)
		}
		cond := bson.M{MgoGe: from, MgoLe: to}
		switch tLeft := left.(type) {
		case string:
			return bson.M{tLeft: cond}, nil
		default:
			return nil, fmt.Errorf("RangeCond left type error: %v (%T)", tLeft, tLeft)
		}

		// TODO: how is 'null' interpreted? exists? 'null'?
	case *sqlparser.NullCheck:
		val, err := translateExpr(expr.Expr)
		if err != nil {
			return nil, fmt.Errorf("NullCheck key error: %v", err)
		}
		switch tVal := val.(type) {
		case string:
			return bson.M{tVal: bson.M{MgoNe: nil}}, nil
		default:
			// TODO: can node not be a string?
			return nil, fmt.Errorf("NullCheck left type error: %v (%T)", val, tVal)
		}

	case *sqlparser.ExistsExpr:
		return nil, fmt.Errorf("where can't handle ExistsExpr type %T", where)

	case *sqlparser.NotExpr:
		return nil, fmt.Errorf("where can't handle NotExpr type %T", where)

	case *sqlparser.ParenBoolExpr:
		return nil, fmt.Errorf("where can't handle ParenBoolExpr type %T", where)

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
