package translator

import (
	"fmt"
	"github.com/mongodb/mongo-tools/common/log"
	"github.com/mongodb/mongo-tools/common/util"
	"github.com/erh/mixer/sqlparser"
	"gopkg.in/mgo.v2/bson"
)

// translateExpr takes an expression and returns its translated form.
func translateExpr(gExpr sqlparser.Expr) (interface{}, error) {
	log.Logf(log.DebugLow, "expr: %s (type is %T)", sqlparser.String(gExpr), gExpr)

	switch expr := gExpr.(type) {

	case sqlparser.NumVal:
		val, err := getNumVal(expr)
		if err != nil {
			return nil, fmt.Errorf("can't handle NumVal %v: %v", expr, err)
		}
		return val, err

	case sqlparser.ValTuple:
		vals := sqlparser.ValExprs(expr)
		tuples := make([]interface{}, len(vals))
		var err error

		for i, val := range vals {
			tuples[i], err = translateExpr(val)
			if err != nil {
				return nil, fmt.Errorf("can't handle ValExpr %v: %v", val, err)
			}
		}
		return tuples, nil

	case *sqlparser.NullVal:
		return nil, nil

		// TODO: regex lowercased
	case *sqlparser.ColName:
		return sqlparser.String(expr), nil

	case sqlparser.StrVal:
		return string(expr), nil 

	case *sqlparser.BinaryExpr:
		left, right, err := translateLRExpr(expr.Left, expr.Right)
		if err != nil {
			return nil, fmt.Errorf("BinaryExpr LR error: %v", err)
		}

		leftVal, err := util.ToInt(left)
		if err != nil {
			return nil, fmt.Errorf("BinaryExpr leftVal error (%v): %v", left, err)
		}

		rightVal, err := util.ToInt(right)
		if err != nil {
			return nil, fmt.Errorf("BinaryExpr rightVal error (%v): %v", right, err)
		}

		switch expr.Operator {
		case sqlparser.AST_BITAND:
			// integers
			return leftVal & rightVal, nil
		case sqlparser.AST_BITOR:
			// integers
			return leftVal | rightVal, nil
		case sqlparser.AST_BITXOR:
			// integers
			return leftVal ^ rightVal, nil
		case sqlparser.AST_PLUS:
			// TODO ?: floats, complex values, strings
			return leftVal + rightVal, nil
		case sqlparser.AST_MINUS:
			// TODO ?: floats, complex values
			return leftVal - rightVal, nil
		case sqlparser.AST_MULT:
			// TODO ?: floats, complex values
			return leftVal * rightVal, nil
		case sqlparser.AST_DIV:
			// TODO ?: floats, complex values
			return leftVal / rightVal, nil
		case sqlparser.AST_MOD:
			// integers
			return leftVal % rightVal, nil
		default:
			return nil, fmt.Errorf("can't handle BinaryExpr operator: %v", expr.Operator)
		}

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
			// TODO: verify structure
			return bson.M{oprtMap[expr.Operator]: []interface{}{left, right}}, nil
		}

	case *sqlparser.RangeCond:
		from, to, err := translateLRExpr(expr.From, expr.To)
		if err != nil {
			return nil, fmt.Errorf("RangeCond LR error: %v", err)
		}

		left, err := translateExpr(expr.Left)
		if err != nil {
			return nil, fmt.Errorf("RangeCond key error: %v", err)
		}

		switch tLeft := left.(type) {
		case string:
			return bson.M{tLeft: bson.M{MgoGe: from, MgoLe: to}}, nil
		default:
			return nil, fmt.Errorf("RangeCond key type error: %v (%T)", tLeft, tLeft)
		}

		// TODO: how is 'null' interpreted? exists? 'null'?
	case *sqlparser.NullCheck:
		val, err := translateExpr(expr.Expr)
		if err != nil {
			return nil, fmt.Errorf("NullCheck error: %v", err)
		}

		switch tVal := val.(type) {
		case string:
			return bson.M{tVal: bson.M{MgoNe: nil}}, nil
		default:
			// TODO: can node not be a string?
			return nil, fmt.Errorf("NullCheck left type error: %v (%T)", val, tVal)
		}

	case *sqlparser.UnaryExpr:
		val, err := translateExpr(expr.Expr)
		if err != nil {
			return nil, fmt.Errorf("UnaryExpr error: %v", err)
		}

		intVal, err := util.ToInt(val)
		if err != nil {
			return nil, fmt.Errorf("UnaryExpr conversion error (%v): %v", val, err)
		}

		switch expr.Operator {
		case sqlparser.AST_UPLUS:
			return intVal, nil
		case sqlparser.AST_UMINUS:
			return -intVal, nil
		case sqlparser.AST_TILDA:
			return ^intVal, nil
		default:
			return nil, fmt.Errorf("can't handle UnaryExpr operator type %T", expr.Operator)
		}

	case *sqlparser.NotExpr:
		val, err := translateExpr(expr.Expr)
		if err != nil {
			return nil, fmt.Errorf("NotExpr error: %v", err)
		}

		return val, err

	case *sqlparser.ParenBoolExpr:
		val, err := translateExpr(expr.Expr)
		if err != nil {
			return nil, fmt.Errorf("ParenBoolExpr error: %v", err)
		}

		return val, err

		//
		//  some nodes rely on SimpleSelect support
		//

	case *sqlparser.Subquery:
		return nil, fmt.Errorf("can't handle Subquery type %T", expr)

	case sqlparser.ValArg:
		return nil, fmt.Errorf("can't handle ValArg type %T", expr)

	case *sqlparser.FuncExpr:
		return nil, fmt.Errorf("can't handle FuncExpr type %T", expr)

		// TODO: might require resultset post-processing
	case *sqlparser.CaseExpr:
		return nil, fmt.Errorf("can't handle CaseExpr type %T", expr)

	case *sqlparser.ExistsExpr:
		return nil, fmt.Errorf("can't handle ExistsExpr type %T", expr)

	default:
		return nil, fmt.Errorf("can't handle expression type %T", expr)
	}

}

// translateTableExpr takes a table expression and returns its translated form.
func translateTableExpr(tExpr sqlparser.TableExpr) (interface{}, error) {

	log.Logf(log.DebugLow, "table expr: %s (type is %T)", sqlparser.String(tExpr), tExpr)

	switch expr := tExpr.(type) {

	case *sqlparser.AliasedTableExpr:

		// TODO: ignoring index hints for now
		stExpr, err := translateSimpleTableExpr(expr.Expr)
		if err != nil {
			return nil, fmt.Errorf("AliasedTableExpr error: %v", err)
		}

		return []interface{}{stExpr, string(expr.As)}, nil

	case *sqlparser.ParenTableExpr:
		ptExpr, err := translateTableExpr(expr.Expr)
		if err != nil {
			return nil, fmt.Errorf("ParenTableExpr error: %v", err)
		}
		return ptExpr, nil

	case *sqlparser.JoinTableExpr:

		left, right, err := translateLRTableExpr(expr.LeftExpr, expr.RightExpr)
		if err != nil {
			return nil, fmt.Errorf("JoinTableExpr LR error: %v", err)
		}

		criterion, err := translateExpr(expr.On)
		if err != nil {
			return nil, fmt.Errorf("JoinTableExpr On error: %v", err)
		}

		return []interface{}{left, criterion, right}, nil

	default:
		return nil, fmt.Errorf("can't handle table expression type %T", expr)
	}

}

// translateSimpleTableExpr takes a simple table expression and returns its translated form.
func translateSimpleTableExpr(stExpr sqlparser.SimpleTableExpr) (interface{}, error) {

	log.Logf(log.DebugLow, "simple table expr: %s (type is %T)", sqlparser.String(stExpr), stExpr)

	switch expr := stExpr.(type) {

	case *sqlparser.TableName:
		// TODO: ignoring qualifier for now
		return sqlparser.String(expr), nil

	case *sqlparser.Subquery:
		return translateExpr(expr)

	default:
		return nil, fmt.Errorf("can't handle simple table expression type %T", expr)
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

// translateLRTableExpr takes two leaf table expressions and returns their translations.
func translateLRTableExpr(lExpr, rExpr sqlparser.TableExpr) (interface{}, interface{}, error) {

	left, err := translateTableExpr(lExpr)
	if err != nil {
		return nil, nil, fmt.Errorf("lTableExpr error: %v", err)
	}

	right, err := translateTableExpr(rExpr)
	if err != nil {
		return nil, nil, fmt.Errorf("rTableExpr error: %v", err)
	}

	return left, right, nil
}
