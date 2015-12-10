package evaluator

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
)

//
// SQLAggFunctionExpr is a wrapper around a sqlparser.FuncExpr designating it
// as an aggregate function. These aggregate functions are avg, sum, count,
// max, and min.
//
type SQLAggFunctionExpr struct {
	*sqlparser.FuncExpr
}

func (f *SQLAggFunctionExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	var distinctMap map[interface{}]bool = nil
	if f.Distinct {
		distinctMap = make(map[interface{}]bool)
	}

	switch string(f.Name) {
	case "avg":
		return f.avgFunc(ctx, distinctMap)
	case "sum":
		return f.sumFunc(ctx, distinctMap)
	case "count":
		return f.countFunc(ctx, distinctMap)
	case "max":
		return f.maxFunc(ctx)
	case "min":
		return f.minFunc(ctx)
	default:
		return nil, fmt.Errorf("aggregate function '%v' is not supported", string(f.Name))
	}
}

func (f *SQLAggFunctionExpr) avgFunc(ctx *EvalCtx, distinctMap map[interface{}]bool) (SQLValue, error) {
	var sum SQLNumeric = SQLInt(0)
	count := 0
	for _, row := range ctx.Rows {
		evalCtx := &EvalCtx{Rows: []Row{row}}
		for _, sExpr := range f.Exprs {
			switch e := sExpr.(type) {
			// mixture of star and non-star expression is acceptable
			case *sqlparser.StarExpr:
				return nil, fmt.Errorf("avg aggregate function can not contain '*'")
			case *sqlparser.NonStarExpr:
				valExpr, err := NewSQLExpr(e.Expr)
				if err != nil {
					return nil, err
				}
				eval, err := valExpr.Evaluate(evalCtx)
				if err != nil {
					return nil, err
				}
				if distinctMap != nil {
					if distinctMap[eval] {
						// already in our distinct map, so we skip this row
						continue
					} else {
						distinctMap[eval] = true
					}
				}
				count += 1
				// TODO: ignoring if we can't convert this to a number
				if n, ok := eval.(SQLNumeric); ok {
					sum = sum.Add(n)
				}
			default:
				return nil, fmt.Errorf("unknown expression in avgFunc: %T", e)
			}
		}
	}

	return SQLFloat(sum.Float64() / float64(count)), nil
}

func (f *SQLAggFunctionExpr) countFunc(ctx *EvalCtx, distinctMap map[interface{}]bool) (SQLValue, error) {
	var count int64
	for _, row := range ctx.Rows {
		evalCtx := &EvalCtx{Rows: []Row{row}}
		for _, sExpr := range f.Exprs {
			switch e := sExpr.(type) {
			// mixture of star and non-star expression is acceptable
			case *sqlparser.StarExpr:
				count += 1

			case *sqlparser.NonStarExpr:
				valExpr, err := NewSQLExpr(e.Expr)
				if err != nil {
					return nil, err
				}
				eval, err := valExpr.Evaluate(evalCtx)
				if err != nil {
					return nil, err
				}
				if distinctMap != nil {
					if distinctMap[eval] {
						// already in our distinct map, so we skip this row
						continue
					} else {
						distinctMap[eval] = true
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
	return SQLInt(count), nil
}

func (f *SQLAggFunctionExpr) maxFunc(ctx *EvalCtx) (SQLValue, error) {
	var max SQLValue
	for _, row := range ctx.Rows {
		evalCtx := &EvalCtx{Rows: []Row{row}}
		for _, sExpr := range f.Exprs {
			switch e := sExpr.(type) {
			// mixture of star and non-star expression is acceptable
			case *sqlparser.StarExpr:
				return nil, fmt.Errorf("max aggregate function can not contain '*'")
			case *sqlparser.NonStarExpr:
				valExpr, err := NewSQLExpr(e.Expr)
				if err != nil {
					return nil, err
				}
				eval, err := valExpr.Evaluate(evalCtx)
				if err != nil {
					return nil, err
				}
				if max == nil {
					max = eval
					continue
				}
				compared, err := max.CompareTo(eval)
				if err != nil {
					return nil, err
				}
				if compared < 0 {
					max = eval
				}
			default:
				return nil, fmt.Errorf("unknown expression in maxFunc: %T", e)
			}
		}
	}
	return max, nil
}

func (f *SQLAggFunctionExpr) minFunc(ctx *EvalCtx) (SQLValue, error) {
	var min SQLValue
	for _, row := range ctx.Rows {
		evalCtx := &EvalCtx{Rows: []Row{row}}
		for _, sExpr := range f.Exprs {
			switch e := sExpr.(type) {
			// mixture of star and non-star expression is acceptable
			case *sqlparser.StarExpr:
				return nil, fmt.Errorf("min aggregate function can not contain '*'")
			case *sqlparser.NonStarExpr:
				valExpr, err := NewSQLExpr(e.Expr)
				if err != nil {
					return nil, err
				}
				eval, err := valExpr.Evaluate(evalCtx)
				if err != nil {
					return nil, err
				}
				if min == nil {
					min = eval
					continue
				}
				compared, err := min.CompareTo(eval)
				if err != nil {
					return nil, err
				}
				if compared > 0 {
					min = eval
				}
			default:
				return nil, fmt.Errorf("unknown expression in minFunc: %T", e)
			}
		}
	}
	return min, nil
}

func (f *SQLAggFunctionExpr) sumFunc(ctx *EvalCtx, distinctMap map[interface{}]bool) (SQLValue, error) {
	var sum SQLNumeric = SQLInt(0)
	for _, row := range ctx.Rows {
		evalCtx := &EvalCtx{Rows: []Row{row}}
		for _, sExpr := range f.Exprs {
			switch e := sExpr.(type) {
			// mixture of star and non-star expression is acceptable
			case *sqlparser.StarExpr:
				return nil, fmt.Errorf("sum aggregate function can not contain '*'")
			case *sqlparser.NonStarExpr:
				valExpr, err := NewSQLExpr(e.Expr)
				if err != nil {
					return nil, err
				}
				eval, err := valExpr.Evaluate(evalCtx)
				if err != nil {
					return nil, err
				}

				if distinctMap != nil {
					if distinctMap[eval] {
						// already in our distinct map, so we skip this row
						continue
					} else {
						distinctMap[eval] = true
					}
				}

				// TODO: ignoring if we can't convert this to a number
				if n, ok := eval.(SQLNumeric); ok {
					sum = sum.Add(n)
				}
			default:
				return nil, fmt.Errorf("unknown expression in sumFunc: %T", e)
			}
		}
	}

	return sum, nil
}
