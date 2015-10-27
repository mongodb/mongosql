package evaluator

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"github.com/mongodb/mongo-tools/common/log"
)

// Tree nodes for evaluating if a row matches.
type Matcher interface {
	Matches(*EvalCtx) (bool, error)
}

// BuildMatcher rewrites a boolean expression as a matcher.
func BuildMatcher(gExpr sqlparser.Expr) (Matcher, error) {
	log.Logf(log.DebugLow, "match expr: %#v (type is %T)", gExpr, gExpr)

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
		return &And{[]Matcher{left, right}}, nil
	case *sqlparser.OrExpr:
		left, err := BuildMatcher(expr.Left)
		if err != nil {
			return nil, err
		}
		right, err := BuildMatcher(expr.Right)
		if err != nil {
			return nil, err
		}
		return &Or{[]Matcher{left, right}}, nil
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
			return &In{left, right}, nil
		case sqlparser.AST_NOT_IN:
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
		val, err := NewSQLValue(expr)
		if err != nil {
			return nil, err
		}
		return &BoolMatcher{val}, nil
	case sqlparser.NumVal:
		val, err := NewSQLValue(expr)
		if err != nil {
			return nil, err
		}
		return &BoolMatcher{val}, nil
	case *sqlparser.FuncExpr:
		sqlFuncVal := &SQLFuncExpr{expr}
		return &BoolMatcher{sqlFuncVal}, nil
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

		m := &And{[]Matcher{lower, upper}}

		if expr.Operator == sqlparser.AST_NOT_BETWEEN {
			return &Not{m}, nil
		}

		return m, nil

	case *sqlparser.UnaryExpr:
		val, err := NewSQLValue(expr.Expr)
		if err != nil {
			return nil, err
		}
		return &BoolMatcher{val}, nil

		/*
			case *sqlparser.Subquery:
			case sqlparser.ValArg:
			case *sqlparser.CaseExpr:
			case *sqlparser.ExistsExpr:
		*/
	default:
		panic(fmt.Errorf("matcher not yet implemented for %v (%T)", sqlparser.String(expr), expr))
		return nil, nil
	}
}

// BuildFuncMatcher creates a matcher for a function expression.
/*
func BuildFuncMatcher(fExpr *sqlparser.FuncExpr) (Matcher, error) {
	log.Logf(log.DebugLow, "building scalar function matcher '%v'", string(fExpr.Name))

	exprs, err := getFuncExprs(fExpr.Exprs)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(string(fExpr.Name)) {
	//
	case "not":
		if len(exprs) != 1 {
			return nil, fmt.Errorf("%v scalar function must accept exactly 1 argument, got %v: %#v", string(fExpr.Name), len(fExpr.Exprs), exprs)
		}
		m, err := BuildMatcher(exprs[0])
		if err != nil {
			return nil, err
		}
		return &Not{m}, nil
	case "isnull":
		if len(exprs) != 1 {
			return nil, fmt.Errorf("%v scalar function must accept exactly 1 argument, got %v: %#v", string(fExpr.Name), len(fExpr.Exprs), exprs)
		}
		val, err := NewSQLValue(exprs[0])
		if err != nil {
			return nil, err
		}
		return &NullMatcher{true, val}, nil
	default:
		return nil, fmt.Errorf("scalar function '%v' is not supported", string(fExpr.Name))
	}
}
*/

// getFuncExprs parses a slice of SelectExpr - part of scalar function
// expressions - and returns the referenced expressions for each.
func getFuncExprs(sExprs sqlparser.SelectExprs) ([]sqlparser.Expr, error) {
	exprs := make([]sqlparser.Expr, len(sExprs))

	for i, sExpr := range sExprs {
		switch expr := sExpr.(type) {
		case *sqlparser.StarExpr:
			return nil, fmt.Errorf("can not have star expression in scalar function")
		case *sqlparser.NonStarExpr:
			exprs[i] = expr.Expr
		}
	}

	return exprs, nil
}
