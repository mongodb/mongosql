package evaluator

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"github.com/mongodb/mongo-tools/common/log"
	"strconv"
)

// Matches checks if a given SQLExpr is "truthy" or should by coercing it to a boolean value.
// - booleans: the result is simply that same return value
// - numeric values: the result is true if and only if the value is non-zero.
// - strings, the result is true if and only if that string can be parsed as a number,
//   and that number is non-zero.
func Matches(expr SQLExpr, ctx *EvalCtx) (bool, error) {

	sv, err := expr.Evaluate(ctx)
	if err != nil {
		return false, err
	}

	if asBool, ok := sv.(SQLBool); ok {
		return bool(asBool), nil
	}
	if asNum, ok := sv.(SQLNumeric); ok {
		return asNum.Float64() != float64(0), nil
	}
	if asStr, ok := sv.(SQLString); ok {
		// check if the string should be considered "truthy" by trying to convert it to a number and comparing to 0.
		// more info: http://stackoverflow.com/questions/12221211/how-does-string-truthiness-work-in-mysql
		if parsedFloat, err := strconv.ParseFloat(string(asStr), 64); err == nil {
			return parsedFloat != float64(0), nil
		}
		return false, nil
	}
	// TODO - handle other types with possible values that are "truthy" : dates, etc?
	return false, nil
}

// BuildMatcher transforms sqlparser expressions into SQLExpr
func BuildMatcher(gExpr sqlparser.Expr) (SQLExpr, error) {
	log.Logf(log.DebugLow, "match expr: %#v (type is %T)\n", gExpr, gExpr)

	switch expr := gExpr.(type) {

	case *sqlparser.AndExpr:

		left, err := BuildMatcher(expr.Left)
		if err != nil {
			return nil, err
		}

		right, err := BuildMatcher(expr.Right)
		if err != nil {
			return nil, err
		}

		return &And{[]SQLExpr{left, right}}, nil

	case *sqlparser.OrExpr:

		left, err := BuildMatcher(expr.Left)
		if err != nil {
			return nil, err
		}

		right, err := BuildMatcher(expr.Right)
		if err != nil {
			return nil, err
		}

		return &Or{[]SQLExpr{left, right}}, nil

	case *sqlparser.ComparisonExpr:

		left, err := NewSQLValue(expr.Left)
		if err != nil {
			return nil, err
		}

		right, err := NewSQLValue(expr.Right)
		if err != nil {
			return nil, err
		}

		switch expr.Operator {
		case sqlparser.AST_EQ:
			return &Equals{left, right}, nil
		case sqlparser.AST_LT:
			return &LessThan{left, right}, nil
		case sqlparser.AST_GT:
			return &GreaterThan{left, right}, nil
		case sqlparser.AST_LE:
			return &LessThanOrEqual{left, right}, nil
		case sqlparser.AST_GE:
			return &GreaterThanOrEqual{left, right}, nil
		case sqlparser.AST_NE:
			return &NotEquals{left, right}, nil
		case sqlparser.AST_LIKE:
			return &Like{left, right}, nil
		case sqlparser.AST_IN:
			switch eval := right.(type) {
			case *SubqueryValue:
				return &SubqueryCmp{true, left, eval}, nil
			}
			return &In{left, right}, nil
		case sqlparser.AST_NOT_IN:
			switch eval := right.(type) {
			case *SubqueryValue:
				return &SubqueryCmp{false, left, eval}, nil
			}
			return &NotIn{left, right}, nil
		default:
			return &Equals{left, right}, fmt.Errorf("sql where clause not implemented: %s", expr.Operator)
		}

	case *sqlparser.NullCheck:

		val, err := NewSQLValue(expr.Expr)
		if err != nil {
			return nil, err
		}

		matcher := &NullMatcher{val}
		if expr.Operator == sqlparser.AST_IS_NULL {
			return matcher, nil
		}

		return &Not{matcher}, nil

	case *sqlparser.NotExpr:

		child, err := BuildMatcher(expr.Expr)
		if err != nil {
			return nil, err
		}

		return &Not{child}, nil

	case *sqlparser.ParenBoolExpr:

		child, err := BuildMatcher(expr.Expr)
		if err != nil {
			return nil, err
		}

		return child, nil

	case nil:

		return &NoopMatcher{}, nil

	case *sqlparser.ColName:

		return NewSQLValue(expr)

	case sqlparser.NumVal:

		return NewSQLValue(expr)

	case *sqlparser.FuncExpr:

		return NewSQLFuncValue(expr)

	case *sqlparser.RangeCond:

		from, err := NewSQLValue(expr.From)
		if err != nil {
			return nil, err
		}

		left, err := NewSQLValue(expr.Left)
		if err != nil {
			return nil, err
		}

		to, err := NewSQLValue(expr.To)
		if err != nil {
			return nil, err
		}

		lower := &GreaterThanOrEqual{left, from}

		upper := &LessThanOrEqual{left, to}

		m := &And{[]SQLExpr{lower, upper}}

		if expr.Operator == sqlparser.AST_NOT_BETWEEN {
			return &Not{m}, nil
		}

		return m, nil

	case *sqlparser.UnaryExpr:

		return NewSQLValue(expr.Expr)

	case *sqlparser.CaseExpr:

		return NewSQLCaseValue(expr)

	case sqlparser.StrVal:

		return NewSQLValue(expr)

	case *sqlparser.Subquery:

		return &ExistsMatcher{expr.Select}, nil

	case *sqlparser.ExistsExpr:

		return &ExistsMatcher{expr.Subquery.Select}, nil

		/*
			case sqlparser.ValArg:
		*/

	default:
		panic(fmt.Errorf("matcher not yet implemented for %v (%T)", sqlparser.String(expr), expr))
	}
}
