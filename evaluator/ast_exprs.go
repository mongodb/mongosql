package evaluator

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"math"
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

//
// SQLCaseExpr holds a number of cases to evaluate as well as the value
// to return if any of the cases is matched. If none is matched,
// 'elseValue' is evaluated and returned.
//
type SQLCaseExpr struct {
	elseValue      SQLExpr
	caseConditions []caseCondition
}

// caseCondition holds a matcher used in evaluating case expressions and
// a value to return if a particular case is matched. If a case is matched,
// the corresponding 'then' value is evaluated and returned ('then'
// corresponds to the 'then' clause in a case expression).
type caseCondition struct {
	matcher SQLExpr
	then    SQLExpr
}

func (s SQLCaseExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	for _, condition := range s.caseConditions {

		m, err := Matches(condition.matcher, ctx)
		if err != nil {
			return nil, err
		}

		if m {
			return condition.then.Evaluate(ctx)
		}
	}

	return s.elseValue.Evaluate(ctx)

}

//
// SQLFieldExpr represents a field reference.
//
type SQLFieldExpr struct {
	tableName string
	fieldName string
}

func (sqlf SQLFieldExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	// TODO how do we report field not existing? do we just treat is a NULL, or something else?
	for _, row := range ctx.Rows {
		for _, data := range row.Data {
			if data.Table == sqlf.tableName {
				if value, hasValue := row.GetField(sqlf.tableName, sqlf.fieldName); hasValue {
					val, err := NewSQLValue(value, "")
					if err != nil {
						return nil, err
					}
					return val, nil
				}
				// field does not exist - return null i guess
				return SQLNull, nil
			}
		}
	}
	return SQLNull, nil
}

//
// SQLScalarFunctionExpr is a wrapper around a sqlparser.FuncExpr.
//
// TODO: we should just convert the sqlparser.FuncExpr into our own.
//
type SQLScalarFunctionExpr struct {
	*sqlparser.FuncExpr
}

func (f *SQLScalarFunctionExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	switch string(f.Name) {
	// connector functions
	case "connection_id":
		return f.connectionIdFunc(ctx)
	case "database":
		return f.dbFunc(ctx)

		// scalar functions
	case "isnull":
		return f.isNullFunc(ctx)
	case "not":
		return f.notFunc(ctx)
	case "pow":
		return f.powFunc(ctx)

	default:
		return nil, fmt.Errorf("function '%v' is not supported", string(f.Name))
	}
}

func (f *SQLScalarFunctionExpr) connectionIdFunc(ctx *EvalCtx) (SQLValue, error) {
	return SQLUint32(ctx.ExecCtx.ConnectionId()), nil
}

func (f *SQLScalarFunctionExpr) dbFunc(ctx *EvalCtx) (SQLValue, error) {
	return SQLString(ctx.ExecCtx.DB()), nil
}

func (f *SQLScalarFunctionExpr) isNullFunc(ctx *EvalCtx) (SQLValue, error) {
	if len(f.Exprs) != 1 {
		return nil, fmt.Errorf("'isnull' function requires exactly one argument")
	}

	var exp sqlparser.Expr

	if v, ok := f.Exprs[0].(*sqlparser.NonStarExpr); ok {
		exp = v.Expr
	} else {
		return nil, fmt.Errorf("argument to 'isnull' function can not contain '*'")
	}

	sqlExpr, err := NewSQLExpr(exp)
	if err != nil {
		return nil, err
	}

	matcher := &SQLNullCmpExpr{sqlExpr}
	result, err := Matches(matcher, ctx)
	if err != nil {
		return nil, err
	}
	return SQLBool(result), nil
}

func (f *SQLScalarFunctionExpr) notFunc(ctx *EvalCtx) (SQLValue, error) {
	if len(f.Exprs) != 1 {
		return nil, fmt.Errorf("'not' function requires exactly one argument")
	}
	var notExpr sqlparser.Expr
	if v, ok := f.Exprs[0].(*sqlparser.NonStarExpr); ok {
		notExpr = v.Expr
	} else {
		return nil, fmt.Errorf("argument to 'not' function can not contain '*'")
	}

	child, err := NewSQLExpr(notExpr)
	if err != nil {
		return nil, err
	}

	matcher := &SQLNotExpr{child}
	result, err := Matches(matcher, ctx)
	if err != nil {
		return nil, err
	}
	return SQLBool(result), nil
}

func (f *SQLScalarFunctionExpr) powFunc(ctx *EvalCtx) (SQLValue, error) {
	if len(f.Exprs) != 2 {
		return nil, fmt.Errorf("'pow' function requires exactly two arguments")
	}
	var baseExpr, expExpr sqlparser.Expr
	if v, ok := f.Exprs[0].(*sqlparser.NonStarExpr); ok {
		baseExpr = v.Expr
	} else {
		return nil, fmt.Errorf("argument to 'pow' function can not contain '*'")
	}
	if v, ok := f.Exprs[1].(*sqlparser.NonStarExpr); ok {
		expExpr = v.Expr
	} else {
		return nil, fmt.Errorf("argument to 'pow' function can not contain '*'")
	}

	base, err := NewSQLExpr(baseExpr)
	if err != nil {
		return nil, err
	}
	exponent, err := NewSQLExpr(expExpr)
	if err != nil {
		return nil, err
	}

	base, err = base.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	exponent, err = exponent.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	if bNum, ok := base.(SQLNumeric); ok {
		if eNum, ok := exponent.(SQLNumeric); ok {
			return SQLFloat(math.Pow(bNum.Float64(), eNum.Float64())), nil
		}
		return nil, fmt.Errorf("exponent must be a number, but got %t", exponent)
	}
	return nil, fmt.Errorf("base must be a number, but got %T", base)
}

// SQLSubqueryExpr is a wrapper around a sqlparser.SelectStatement representing
// a subquery.
type SQLSubqueryExpr struct {
	stmt sqlparser.SelectStatement
}

func (sv *SQLSubqueryExpr) Evaluate(ctx *EvalCtx) (value SQLValue, err error) {

	ctx.ExecCtx.Depth += 1

	var operator Operator

	operator, err = PlanQuery(ctx.ExecCtx, sv.stmt)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err == nil {
			err = operator.Close()
		} else {
			operator.Close()
		}

		if err == nil {
			err = operator.Err()
		}

		// add context to error
		if err != nil {
			err = fmt.Errorf("SubqueryValue (%v): %v", ctx.ExecCtx.Depth, err)
		}

		ctx.ExecCtx.Depth -= 1

	}()

	err = operator.Open(ctx.ExecCtx)
	if err != nil {
		return nil, err
	}

	row := &Row{}

	hasNext := operator.Next(row)

	// Filter has to check the entire source to return an accurate 'hasNext'
	if hasNext && operator.Next(&Row{}) {
		return nil, fmt.Errorf("Subquery returns more than 1 row")
	}

	values := row.GetValues(operator.OpFields())

	eval := SQLValues{}
	for _, value := range values {

		field, err := NewSQLValue(value, "")
		if err != nil {
			return nil, err
		}

		eval = append(eval, field)

	}

	return eval, nil
}

//
// SQLTupleExpr represents a tuple.
//
type SQLTupleExpr struct {
	Exprs []SQLExpr
}

func (te SQLTupleExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	var values []SQLValue

	for _, v := range te.Exprs {
		value, err := v.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}

	return SQLValues(values), nil
}
