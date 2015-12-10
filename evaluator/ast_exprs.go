package evaluator

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"math"
)

//
// SQLBinaryFunctionExpr represents a function invocation containing the
// function as well as the arguments.
//
type SQLBinaryFunctionExpr struct {
	arguments []SQLExpr
	function  func([]SQLExpr, *EvalCtx) (SQLValue, error)
}

func (sqlfunc *SQLBinaryFunctionExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return sqlfunc.function(sqlfunc.arguments, ctx)
}

// SQLBinaryFunction is a function alias expected by SQLBinaryFunctionExpr.
type SQLBinaryFunction func([]SQLExpr, *EvalCtx) (SQLValue, error)

var binaryFuncMap = map[string]SQLBinaryFunction{

	"+": SQLBinaryFunction(func(args []SQLExpr, ctx *EvalCtx) (SQLValue, error) {
		return sqlNumericBinaryOp(args, ctx, "+")
	}),

	"-": SQLBinaryFunction(func(args []SQLExpr, ctx *EvalCtx) (SQLValue, error) {
		return sqlNumericBinaryOp(args, ctx, "-")
	}),

	"*": SQLBinaryFunction(func(args []SQLExpr, ctx *EvalCtx) (SQLValue, error) {
		return sqlNumericBinaryOp(args, ctx, "*")
	}),

	"/": SQLBinaryFunction(func(args []SQLExpr, ctx *EvalCtx) (SQLValue, error) {
		return sqlNumericBinaryOp(args, ctx, "/")
	}),
}

func convertToSQLNumeric(expr SQLExpr, ctx *EvalCtx) (SQLNumeric, error) {
	eval, err := expr.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	switch v := eval.(type) {

	case SQLNumeric:

		return v, nil

	case SQLValues:

		if len(v.Values) != 1 {
			return nil, fmt.Errorf("expected only one SQLValues value - got %v", len(v.Values))
		}

		return convertToSQLNumeric(v.Values[0], ctx)

	default:

		return nil, fmt.Errorf("can not convert %T to SQLNumeric", eval)

	}

}

func sqlNumericBinaryOp(args []SQLExpr, ctx *EvalCtx, op string) (SQLValue, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("%v function needs at least 2 args", op)
	}

	left, err := convertToSQLNumeric(args[0], ctx)
	if err != nil {
		return nil, err
	}

	for _, arg := range args[1:] {
		right, err := convertToSQLNumeric(arg, ctx)
		if err != nil {
			return nil, err
		}
		switch op {
		case "+":
			left = left.Add(right)
		case "-":
			left = left.Sub(right)
		case "*":
			left = left.Product(right)
		case "/":
			if right.Float64() == 0 {
				return SQLNull, nil
			}
			left = SQLFloat(left.Float64() / right.Float64())
		default:
			return nil, fmt.Errorf("unsupported numeric binary operation: '%v'", op)
		}
	}
	return left, nil
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

	eval := SQLValues{}

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

	for _, value := range values {

		field, err := NewSQLValue(value, "")
		if err != nil {
			return nil, err
		}

		eval.Values = append(eval.Values, field)

	}

	return eval, nil
}

//
// SQLUnaryMinusExpr represents a unary minus expression.
//
type SQLUnaryMinusExpr struct {
	SQLExpr
}

func (um *SQLUnaryMinusExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	if val, ok := um.SQLExpr.(SQLNumeric); ok {
		return SQLInt(-(round(val.Float64()))), nil
	}
	return um.SQLExpr.Evaluate(ctx)
}

//
// SQLUnaryPlusExpr represents a unary plus expression.
//
type SQLUnaryPlusExpr struct {
	SQLExpr
}

func (up *SQLUnaryPlusExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	if val, ok := up.SQLExpr.(SQLNumeric); ok {
		return SQLInt(round(val.Float64())), nil
	}
	return up.SQLExpr.Evaluate(ctx)
}

//
// SQLUnaryTildeExpr represents a unary tilde expression.
//
type SQLUnaryTildeExpr struct {
	SQLExpr
}

func (td *SQLUnaryTildeExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	if val, ok := td.SQLExpr.(SQLNumeric); ok {
		return SQLInt(^round(val.Float64())), nil
	}
	return td.SQLExpr.Evaluate(ctx)
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

	return SQLValues{values}, nil
}
