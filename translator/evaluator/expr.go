package evaluator

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"github.com/erh/mongo-sql-temp/translator/types"
	"github.com/mongodb/mongo-tools/common/util"
)

type Expr interface {
	Evaluate(*EvalCtx) (interface{}, error)
}

func NewExpr(e sqlparser.Expr) (Expr, error) {
	switch expr := e.(type) {
	case *sqlparser.AndExpr:
		return nil, fmt.Errorf("NewExpr not yet implemented for %#v", e)
	case *sqlparser.OrExpr:
		return nil, fmt.Errorf("NewExpr not yet implemented for %#v", e)
	case *sqlparser.NotExpr:
		return nil, fmt.Errorf("NewExpr not yet implemented for %#v", e)
	case *sqlparser.ParenBoolExpr:
		return nil, fmt.Errorf("NewExpr not yet implemented for %#v", e)
	case *sqlparser.ComparisonExpr:
		return nil, fmt.Errorf("NewExpr not yet implemented for %#v", e)
	case *sqlparser.RangeCond:
		return nil, fmt.Errorf("NewExpr not yet implemented for %#v", e)
	case *sqlparser.NullCheck:
		return nil, fmt.Errorf("NewExpr not yet implemented for %#v", e)
	case *sqlparser.ExistsExpr:
		return nil, fmt.Errorf("NewExpr not yet implemented for %#v", e)
	case *sqlparser.StrVal:
		return nil, fmt.Errorf("NewExpr not yet implemented for %#v", e)
	case *sqlparser.NumVal:
		return nil, fmt.Errorf("NewExpr not yet implemented for %#v", e)
	case *sqlparser.ValArg:
		return nil, fmt.Errorf("NewExpr not yet implemented for %#v", e)
	case *sqlparser.NullVal:
		return nil, fmt.Errorf("NewExpr not yet implemented for %#v", e)
	case *sqlparser.ColName:
		return &ColName{expr}, nil
	case *sqlparser.ValTuple:
		return nil, fmt.Errorf("NewExpr not yet implemented for %#v", e)
	case *sqlparser.Subquery:
		return nil, fmt.Errorf("NewExpr not yet implemented for %#v", e)
	case *sqlparser.BinaryExpr:
		return &BinaryExpr{expr}, nil
	case *sqlparser.UnaryExpr:
		return nil, fmt.Errorf("NewExpr not yet implemented for %#v", e)
	case *sqlparser.FuncExpr:
		return &FuncExpr{expr}, nil
	case *sqlparser.CaseExpr:
		return nil, fmt.Errorf("NewExpr not yet implemented for %#v", e)
	default:
		return nil, fmt.Errorf("NewExpr not yet implemented for %#v", e)
	}
}

//---
type ColName struct {
	*sqlparser.ColName
}

func (c *ColName) Evaluate(ctx *EvalCtx) (interface{}, error) {
	for _, r := range ctx.Rows {
		if v, ok := r.GetField(string(c.Qualifier), string(c.Name)); ok {
			return v, nil
		}
	}
	return nil, nil
}

func (c *ColName) String() string {
	return fmt.Sprintf("FQNS: '%v.%v'", string(c.Qualifier), string(c.Name))
}

//---
type FuncExpr struct {
	*sqlparser.FuncExpr
}

func (f *FuncExpr) Evaluate(ctx *EvalCtx) (interface{}, error) {
	switch string(f.Name) {
	case "sum":
		// TODO: handle distinct
		return sumFunc(ctx, f.Exprs)
	case "count":
		return countFunc(ctx, f.Exprs)
	default:
		return nil, fmt.Errorf("function '%v' not yet implemented", string(f.Name))
	}
}

func sumFunc(ctx *EvalCtx, sExprs sqlparser.SelectExprs) (interface{}, error) {
	var sum int64

	for _, row := range ctx.Rows {

		for _, sExpr := range sExprs {

			switch e := sExpr.(type) {

			// mixture of star and non-star expression is acceptable
			case *sqlparser.StarExpr:
				sum += 1

			case *sqlparser.NonStarExpr:
				expr, err := NewExpr(e.Expr)
				if err != nil {
					panic(err)
				}

				evalCtx := &EvalCtx{Rows: []types.Row{row}}
				eval, err := expr.Evaluate(evalCtx)
				if err != nil {
					return nil, err
				}

				// TODO: ignoring if we can't convert this to an integer
				// we should instead support all summable types
				value, _ := util.ToInt(eval)

				sum += int64(value)

			default:
				return nil, fmt.Errorf("unknown expression in sumFunc: %T", e)
			}
		}
	}
	return sum, nil
}

func countFunc(ctx *EvalCtx, sExprs sqlparser.SelectExprs) (interface{}, error) {
	var count int64

	for _, row := range ctx.Rows {

		for _, sExpr := range sExprs {

			switch e := sExpr.(type) {

			// mixture of star and non-star expression is acceptable
			case *sqlparser.StarExpr:
				count += 1

			case *sqlparser.NonStarExpr:
				expr, err := NewExpr(e.Expr)
				if err != nil {
					panic(err)
				}

				evalCtx := &EvalCtx{Rows: []types.Row{row}}
				eval, err := expr.Evaluate(evalCtx)
				if err != nil {
					return nil, err
				}

				if eval != nil {
					count += 1
				}

			default:
				return nil, fmt.Errorf("unknown expression in countFunc: %T", e)
			}
		}
	}
	return count, nil
}

//---
type BinaryExpr struct {
	*sqlparser.BinaryExpr
}

func (b *BinaryExpr) Evaluate(ctx *EvalCtx) (interface{}, error) {
	return nil, nil
}
