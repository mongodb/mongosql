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

		v, err := NewSQLFuncValue(expr)
		if err != nil {
			return nil, err
		}

		return &BoolMatcher{v}, nil

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

	case *sqlparser.CaseExpr:

		val, err := NewSQLCaseValue(expr)
		if err != nil {
			return nil, err
		}

		return &BoolMatcher{val}, nil

	case sqlparser.StrVal:
		val, err := NewSQLValue(expr)
		if err != nil {
			return nil, err
		}

		return &BoolMatcher{val}, nil

	case *sqlparser.Subquery:

		val := &SubqueryValue{expr.Select}

		return &BoolMatcher{val}, nil

	case *sqlparser.ExistsExpr:

		val := &SubqueryValue{expr.Subquery.Select}

		return &BoolMatcher{val}, nil

		/*
			case sqlparser.ValArg:
		*/

	default:
		panic(fmt.Errorf("matcher not yet implemented for %v (%T)", sqlparser.String(expr), expr))
		return nil, nil
	}
}
