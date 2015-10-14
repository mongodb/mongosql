package evaluator

import (
	"errors"
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"github.com/mongodb/mongo-tools/common/log"
	"gopkg.in/mgo.v2/bson"
)

var ErrUntransformableCondition = errors.New("condition can't be expressed as a field:value pair")

// Tree nodes for evaluating if a row matches
type Matcher interface {
	Matches(*EvalCtx) (bool, error)
	Transform() (*bson.D, error)
}

// BuildMatcher rewrites a boolean expression as a matcher.
func BuildMatcher(gExpr sqlparser.Expr) (Matcher, error) {
	log.Logf(log.DebugLow, "expr: %#v (type is %T)", gExpr, gExpr)

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
	case *sqlparser.RangeCond:
		// BETWEEN
		panic("not implemented: RangeCond")
	case *sqlparser.NullCheck:
		val, err := NewSQLValue(expr.Expr)
		if err != nil {
			return nil, err
		}
		return &NullMatch{expr.Operator == sqlparser.AST_IS_NULL, val}, nil
	case *sqlparser.UnaryExpr:
		panic("not implemented: UnaryExpr")
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
	case *sqlparser.Subquery:
		panic("not implemented: subquery")
		return nil, nil
	case sqlparser.ValArg:
		panic("not implemented: function")
		return nil, fmt.Errorf("can't handle ValArg type %T", expr)
	case *sqlparser.FuncExpr:
		panic("not implemented: function")
		return nil, nil
	case *sqlparser.CaseExpr:
		panic("not implemented: case")
		return nil, nil
	case *sqlparser.ExistsExpr:
		panic("not implemented: exists")
		return nil, nil
	case nil:
		return &NoopMatch{}, nil
	case *sqlparser.ColName:
		val, err := NewSQLValue(expr)
		if err != nil {
			return nil, err
		}
		return &BoolMatch{val}, nil
	default:
		panic(fmt.Errorf("not implemented: %v (%T)", sqlparser.String(expr), expr))
		return nil, nil
	}

}
