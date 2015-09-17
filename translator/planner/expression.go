package planner

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
)

type Expr interface {
	Evaluate(*ExecutionCtx) (interface{}, error)
}

func NewExpr(e sqlparser.Expr) (Expr, error) {
	switch expr := e.(type) {
	case *sqlparser.AndExpr:
		return nil, fmt.Errorf("NYI for %#v", e)
	case *sqlparser.OrExpr:
		return nil, fmt.Errorf("NYI for %#v", e)
	case *sqlparser.NotExpr:
		return nil, fmt.Errorf("NYI for %#v", e)
	case *sqlparser.ParenBoolExpr:
		return nil, fmt.Errorf("NYI for %#v", e)
	case *sqlparser.ComparisonExpr:
		return nil, fmt.Errorf("NYI for %#v", e)
	case *sqlparser.RangeCond:
		return nil, fmt.Errorf("NYI for %#v", e)
	case *sqlparser.NullCheck:
		return nil, fmt.Errorf("NYI for %#v", e)
	case *sqlparser.ExistsExpr:
		return nil, fmt.Errorf("NYI for %#v", e)
	case *sqlparser.StrVal:
		return nil, fmt.Errorf("NYI for %#v", e)
	case *sqlparser.NumVal:
		return nil, fmt.Errorf("NYI for %#v", e)
	case *sqlparser.ValArg:
		return nil, fmt.Errorf("NYI for %#v", e)
	case *sqlparser.NullVal:
		return nil, fmt.Errorf("NYI for %#v", e)
	case *sqlparser.ColName:
		return &ColName{expr}, nil
	case *sqlparser.ValTuple:
		return nil, fmt.Errorf("NYI for %#v", e)
	case *sqlparser.Subquery:
		return nil, fmt.Errorf("NYI for %#v", e)
	case *sqlparser.BinaryExpr:
		return nil, fmt.Errorf("NYI for %#v", e)
	case *sqlparser.UnaryExpr:
		return nil, fmt.Errorf("NYI for %#v", e)
	case *sqlparser.FuncExpr:
		return nil, fmt.Errorf("NYI for %#v", e)
	case *sqlparser.CaseExpr:
		return nil, fmt.Errorf("NYI for %#v", e)
	default:
		return nil, fmt.Errorf("NYI for %#v", e)
	}
}

type ColName struct {
	*sqlparser.ColName
}

func (c *ColName) Evaluate(ctx *ExecutionCtx) (interface{}, error) {
	if v, ok := ctx.Row.GetField(string(c.Qualifier), string(c.Name)); ok {
		return v, nil
	}
	return nil, nil
}
