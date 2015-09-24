package planner

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
)

type Expr interface {
	Evaluate(*EvalCtx) (SQLValue, error)
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

func (c *ColName) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	for _, r := range ctx.rows {
		if v, ok := r.GetField(string(c.Qualifier), string(c.Name)); ok {
			return NewSQLField(v)
		}
	}
	return nil, nil
}

func (c *ColName) String() string {
	return fmt.Sprintf("FQNS: '%v.%v'", string(c.Qualifier), string(c.Name))
}

type FuncExpr struct {
	*sqlparser.FuncExpr
}

func (f *FuncExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	var distinctMap map[interface{}]bool = nil
	if f.Distinct {
		distinctMap = make(map[interface{}]bool)
	}
	switch string(f.Name) {
	case "avg":
		return avgFunc(ctx, f.Exprs, distinctMap)
	case "sum":
		return sumFunc(ctx, f.Exprs, distinctMap)
	case "count":
		return countFunc(ctx, f.Exprs, distinctMap)
	case "max":
		return maxFunc(ctx, f.Exprs)
	case "min":
		return minFunc(ctx, f.Exprs)
	default:
		return nil, fmt.Errorf("function '%v' not yet implemented", string(f.Name))
	}
}

func avgFunc(ctx *EvalCtx, sExprs sqlparser.SelectExprs, distinctMap map[interface{}]bool) (SQLValue, error) {
	var sum float64
	count := 0
	for _, row := range ctx.rows {
		evalCtx := &EvalCtx{rows: []Row{row}}
		for _, sExpr := range sExprs {
			switch e := sExpr.(type) {
			// mixture of star and non-star expression is acceptable
			case *sqlparser.StarExpr:
				sum += 1
			case *sqlparser.NonStarExpr:
				val, err := BuildValue(e.Expr)
				if err != nil {
					return nil, err
				}
				eval := val.Evaluate(evalCtx)
				if distinctMap != nil {
					rawVal := eval.MongoValue()
					if distinctMap[rawVal] {
						// already in our distinct map, so we skip this row
						continue
					} else {
						distinctMap[rawVal] = true
					}
				}
				count += 1
				// TODO: ignoring if we can't convert this to a number
				if n, ok := eval.(SQLNumeric); ok {
					sum += float64(n)
				}
			default:
				return nil, fmt.Errorf("unknown expression in sumFunc: %T", e)
			}
		}
	}
	return SQLNumeric(sum / float64(count)), nil
}

func sumFunc(ctx *EvalCtx, sExprs sqlparser.SelectExprs, distinctMap map[interface{}]bool) (SQLValue, error) {
	var sum float64
	for _, row := range ctx.rows {
		evalCtx := &EvalCtx{rows: []Row{row}}
		for _, sExpr := range sExprs {
			switch e := sExpr.(type) {
			// mixture of star and non-star expression is acceptable
			case *sqlparser.StarExpr:
				sum += 1
			case *sqlparser.NonStarExpr:
				val, err := BuildValue(e.Expr)
				if err != nil {
					return nil, err
				}
				eval := val.Evaluate(evalCtx)
				if distinctMap != nil {
					rawVal := eval.MongoValue()
					if distinctMap[rawVal] {
						// already in our distinct map, so we skip this row
						continue
					} else {
						distinctMap[rawVal] = true
					}
				}
				// TODO: ignoring if we can't convert this to a number
				if n, ok := eval.(SQLNumeric); ok {
					sum += float64(n)
				}
			default:
				return nil, fmt.Errorf("unknown expression in sumFunc: %T", e)
			}
		}
	}
	return SQLNumeric(sum), nil
}

func countFunc(ctx *EvalCtx, sExprs sqlparser.SelectExprs, distinctMap map[interface{}]bool) (SQLValue, error) {
	var count int64
	for _, row := range ctx.rows {
		evalCtx := &EvalCtx{rows: []Row{row}}
		for _, sExpr := range sExprs {
			switch e := sExpr.(type) {
			// mixture of star and non-star expression is acceptable
			case *sqlparser.StarExpr:
				count += 1

			case *sqlparser.NonStarExpr:
				val, err := BuildValue(e.Expr)
				if err != nil {
					return nil, err
				}
				eval := val.Evaluate(evalCtx)
				if distinctMap != nil {
					rawVal := eval.MongoValue()
					if distinctMap[rawVal] {
						// already in our distinct map, so we skip this row
						continue
					} else {
						distinctMap[rawVal] = true
					}
				}

				if eval != nil && eval != SQLNull {
					count += 1
				}

			default:
				return nil, fmt.Errorf("unknown expression in countFunc: %T", e)
			}
		}
	}
	return SQLNumeric(count), nil
}

func minFunc(ctx *EvalCtx, sExprs sqlparser.SelectExprs) (SQLValue, error) {
	var min SQLValue
	for _, row := range ctx.rows {
		evalCtx := &EvalCtx{rows: []Row{row}}
		for _, sExpr := range sExprs {
			switch e := sExpr.(type) {
			// mixture of star and non-star expression is acceptable
			case *sqlparser.StarExpr:
				return nil, fmt.Errorf("can't use * as argument to min function.")
			case *sqlparser.NonStarExpr:
				val, err := BuildValue(e.Expr)
				if err != nil {
					return nil, err
				}
				eval := val.Evaluate(evalCtx)
				if min == nil {
					min = eval
					continue
				}
				compared, err := min.CompareTo(evalCtx, eval)
				if err != nil {
					return nil, err
				}
				if compared > 0 {
					min = eval
				}
			default:
				return nil, fmt.Errorf("unknown expression in countFunc: %T", e)
			}
		}
	}
	return min, nil
}

func maxFunc(ctx *EvalCtx, sExprs sqlparser.SelectExprs) (SQLValue, error) {
	var max SQLValue
	for _, row := range ctx.rows {
		evalCtx := &EvalCtx{rows: []Row{row}}
		for _, sExpr := range sExprs {
			switch e := sExpr.(type) {
			// mixture of star and non-star expression is acceptable
			case *sqlparser.StarExpr:
				return nil, fmt.Errorf("can't use * as argument to max function.")
			case *sqlparser.NonStarExpr:
				val, err := BuildValue(e.Expr)
				if err != nil {
					return nil, err
				}
				eval := val.Evaluate(evalCtx)
				if max == nil {
					max = eval
					continue
				}
				compared, err := max.CompareTo(evalCtx, eval)
				if err != nil {
					return nil, err
				}
				if compared < 0 {
					max = eval
				}
			default:
				return nil, fmt.Errorf("unknown expression in countFunc: %T", e)
			}
		}
	}
	return max, nil
}

//---
type BinaryExpr struct {
	*sqlparser.BinaryExpr
}

func (b *BinaryExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return SQLNull, nil
}
