package evaluator

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"math"
	"time"
)

//
// A function invocation.
//
type SQLBinaryValue struct {
	arguments []SQLValue
	function  func([]SQLValue, *EvalCtx) (SQLValue, error)
}

func (sqlfunc *SQLBinaryValue) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return sqlfunc.function(sqlfunc.arguments, ctx)
}

func (sqlfunc *SQLBinaryValue) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {

	left, err := sqlfunc.Evaluate(ctx)
	if err != nil {
		return 0, err
	}

	right, err := v.Evaluate(ctx)
	if err != nil {
		return 0, err
	}

	return left.CompareTo(ctx, right)
}

type SQLBinaryFunction func([]SQLValue, *EvalCtx) (SQLValue, error)

var binaryFuncMap = map[string]SQLBinaryFunction{

	"+": SQLBinaryFunction(func(args []SQLValue, ctx *EvalCtx) (SQLValue, error) {
		return SQLNumericBinaryOp(args, ctx, "+")
	}),

	"-": SQLBinaryFunction(func(args []SQLValue, ctx *EvalCtx) (SQLValue, error) {
		return SQLNumericBinaryOp(args, ctx, "-")
	}),

	"*": SQLBinaryFunction(func(args []SQLValue, ctx *EvalCtx) (SQLValue, error) {
		return SQLNumericBinaryOp(args, ctx, "*")
	}),

	"/": SQLBinaryFunction(func(args []SQLValue, ctx *EvalCtx) (SQLValue, error) {
		return SQLNumericBinaryOp(args, ctx, "/")
	}),
}

func convertToSQLNumeric(v SQLValue, ctx *EvalCtx) (SQLNumeric, error) {
	eval, err := v.Evaluate(ctx)
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

func SQLNumericBinaryOp(args []SQLValue, ctx *EvalCtx, op string) (SQLValue, error) {
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
				return &SQLNullValue{}, nil
			}
			left = SQLFloat(left.Float64() / right.Float64())
		default:
			return nil, fmt.Errorf("unsupported numeric binary operation: '%v'", op)
		}
	}
	return left, nil
}

//
// A boolean value.
//
type SQLBool bool

func (sb SQLBool) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return sb, nil
}

func (sb SQLBool) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	c, err := v.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	if n, ok := c.(SQLBool); ok {
		s1, s2 := bool(sb), bool(n)
		if s1 == s2 {
			return 0, nil
		} else if !s1 {
			return -1, nil
		}
		return 1, nil
	}
	// TODO: support comparing with SQLInt, SQLFloat, etc
	return 1, fmt.Errorf("type mismatch")
}

//
// SQLCaseValue holds a number of cases to evaluate as well as the value
// to return if any of the cases is matched. If none is matched,
// 'elseValue' is evaluated and returned.
//
type SQLCaseValue struct {
	elseValue      SQLValue
	caseConditions []caseCondition
}

// caseCondition holds a matcher used in evaluating case expressions and
// a value to return if a particular case is matched. If a case is matched,
// the corresponding 'then' value is evaluated and returned ('then'
// corresponds to the 'then' clause in a case expression).
type caseCondition struct {
	matcher SQLExpr
	then    SQLValue
}

func (s SQLCaseValue) Evaluate(ctx *EvalCtx) (SQLValue, error) {

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

func (s SQLCaseValue) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	left, err := s.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	right, err := v.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	return left.CompareTo(ctx, right)
}

//
// A date value.
//
type SQLDate struct {
	Time time.Time
}

func (sd SQLDate) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return sd, nil
}

func (sd SQLDate) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	c, err := v.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	if cmp, ok := c.(SQLDate); ok {
		if sd.Time.After(cmp.Time) {
			return 1, nil
		} else if sd.Time.Before(cmp.Time) {
			return -1, nil
		}
	}
	// TODO: type sort order implementation
	return 0, nil
}

//
// A date time value.
//
type SQLDateTime struct {
	Time time.Time
}

func (sd SQLDateTime) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return sd, nil
}

func (sd SQLDateTime) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	c, err := v.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	if cmp, ok := c.(SQLDate); ok {
		if sd.Time.After(cmp.Time) {
			return 1, nil
		} else if sd.Time.Before(cmp.Time) {
			return -1, nil
		}
	}

	// TODO: type sort order implementation
	return 0, nil
}

//
// A field reference.
//
type SQLField struct {
	tableName string
	fieldName string
}

func (sqlf SQLField) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	// TODO how do we report field not existing? do we just treat is a NULL, or something else?
	for _, row := range ctx.Rows {
		for _, data := range row.Data {
			if data.Table == sqlf.tableName {
				if value, hasValue := row.GetField(sqlf.tableName, sqlf.fieldName); hasValue {
					val, err := NewSQLField(value, "")
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

func (sqlf SQLField) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	left, err := sqlf.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	right, err := v.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	return left.CompareTo(ctx, right)
}

//
// A float value.
//
type SQLFloat float64

func (sf SQLFloat) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return sf, nil
}

func (sf SQLFloat) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	c, err := v.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	if n, ok := c.(SQLNumeric); ok {
		cmp := sf.Float64() - n.Float64()
		if cmp > 0 {
			return 1, nil
		} else if cmp < 0 {
			return -1, nil
		}
		return 0, nil
	}
	return -1, nil
}

func (sf SQLFloat) Add(o SQLNumeric) SQLNumeric {
	return SQLFloat(float64(sf) + o.Float64())
}

func (sf SQLFloat) Float64() float64 {
	return float64(sf)
}

func (sf SQLFloat) Product(o SQLNumeric) SQLNumeric {
	return SQLFloat(float64(sf) * o.Float64())
}

func (sf SQLFloat) Sub(o SQLNumeric) SQLNumeric {
	return SQLFloat(float64(sf) - o.Float64())
}

//
// A 64-bit integer value.
//
type SQLInt int64

func (si SQLInt) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return si, nil
}

func (si SQLInt) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	c, err := v.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	if n, ok := c.(SQLInt); ok {
		cmp := int64(si) - int64(n)
		if cmp > 0 {
			return 1, nil
		} else if cmp < 0 {
			return -1, nil
		}
		return 0, nil
	} else if n, ok := c.(SQLFloat); ok {
		cmp := si.Float64() - n.Float64()
		if cmp > 0 {
			return 1, nil
		} else if cmp < 0 {
			return -1, nil
		}
		return 0, nil
	}
	return -1, nil
}

func (si SQLInt) Add(o SQLNumeric) SQLNumeric {
	if oi, ok := o.(SQLInt); ok {
		return SQLInt(int64(si) + int64(oi))
	}
	return SQLFloat(si.Float64() + o.Float64())
}

func (si SQLInt) Float64() float64 {
	return float64(si)
}

func (si SQLInt) Product(o SQLNumeric) SQLNumeric {
	if oi, ok := o.(SQLInt); ok {
		return SQLInt(int64(si) * int64(oi))
	}
	return SQLFloat(si.Float64() * o.Float64())
}

func (si SQLInt) Sub(o SQLNumeric) SQLNumeric {
	if oi, ok := o.(SQLInt); ok {
		return SQLInt(int64(si) - int64(oi))
	}
	return SQLFloat(si.Float64() - o.Float64())
}

//
// The null value.
//
type SQLNullValue struct{}

var SQLNull = SQLNullValue{}

func (nv SQLNullValue) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return nv, nil
}

func (nv SQLNullValue) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	c, err := v.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	if _, ok := c.(SQLNullValue); ok {
		return 0, nil
	}
	return 1, nil
}

//
// A paren bool value.
//
// TODO: what is this?
type SQLParenBoolValue struct {
	*sqlparser.ParenBoolExpr
}

func (p *SQLParenBoolValue) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	expr, ok := p.Expr.(sqlparser.Expr)

	if !ok {
		return nil, fmt.Errorf("could not convert ParenBoolExpr Expr to Expr")
	}

	matcher, err := NewSQLExpr(expr)
	if err != nil {
		return nil, err
	}

	b, err := Matches(matcher, ctx)
	return SQLBool(b), err
}

func (p *SQLParenBoolValue) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	left, err := p.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	right, err := v.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	return left.CompareTo(ctx, right)
}

//
// A scalar function.
//
type SQLScalarFuncValue struct {
	*sqlparser.FuncExpr
}

func (f *SQLScalarFuncValue) Evaluate(ctx *EvalCtx) (SQLValue, error) {
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

func (f *SQLScalarFuncValue) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	left, err := f.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	right, err := v.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	return left.CompareTo(ctx, right)
}

func (f *SQLScalarFuncValue) connectionIdFunc(ctx *EvalCtx) (SQLValue, error) {
	return SQLUint32(ctx.ExecCtx.ConnectionId()), nil
}

func (f *SQLScalarFuncValue) dbFunc(ctx *EvalCtx) (SQLValue, error) {
	return SQLString(ctx.ExecCtx.DB()), nil
}

func (f *SQLScalarFuncValue) isNullFunc(ctx *EvalCtx) (SQLValue, error) {
	if len(f.Exprs) != 1 {
		return nil, fmt.Errorf("'isnull' function requires exactly one argument")
	}

	var exp sqlparser.Expr

	if v, ok := f.Exprs[0].(*sqlparser.NonStarExpr); ok {
		exp = v.Expr
	} else {
		return nil, fmt.Errorf("argument to 'isnull' function can not contain '*'")
	}

	val, err := NewSQLValue(exp)
	if err != nil {
		return nil, err
	}

	val, err = val.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	matcher := &NullMatcher{val}
	result, err := Matches(matcher, ctx)
	if err != nil {
		return nil, err
	}
	return SQLBool(result), nil
}

func (f *SQLScalarFuncValue) notFunc(ctx *EvalCtx) (SQLValue, error) {
	if len(f.Exprs) != 1 {
		return nil, fmt.Errorf("'not' function requires exactly one argument")
	}
	var notExpr sqlparser.Expr
	if v, ok := f.Exprs[0].(*sqlparser.NonStarExpr); ok {
		notExpr = v.Expr
	} else {
		return nil, fmt.Errorf("argument to 'not' function can not contain '*'")
	}

	notVal, err := NewSQLValue(notExpr)
	if err != nil {
		return nil, err
	}

	matcher := &Not{notVal}
	result, err := Matches(matcher, ctx)
	if err != nil {
		return nil, err
	}
	return SQLBool(result), nil
}

func (f *SQLScalarFuncValue) powFunc(ctx *EvalCtx) (SQLValue, error) {
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

	base, err := NewSQLValue(baseExpr)
	if err != nil {
		return nil, err
	}
	exponent, err := NewSQLValue(expExpr)
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

//
// A string value.
//
type SQLString string

func (ss SQLString) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return ss, nil
}

func (sn SQLString) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	c, err := v.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	if n, ok := c.(SQLString); ok {
		s1, s2 := string(sn), string(n)
		if s1 < s2 {
			return -1, nil
		} else if s1 > s2 {
			return 1, nil
		}
		return 0, nil
	}
	// can only compare numbers to each other, otherwise treat as error
	return 1, fmt.Errorf("type mismatch")
}

// A subquery as a value.
type SQLSubqueryValue struct {
	stmt sqlparser.SelectStatement
}

func (sv *SQLSubqueryValue) Evaluate(ctx *EvalCtx) (value SQLValue, err error) {

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

		field, err := NewSQLField(value, "")
		if err != nil {
			return nil, err
		}

		eval.Values = append(eval.Values, field)

	}

	return eval, nil
}

func (sv *SQLSubqueryValue) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	left, err := sv.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	right, err := v.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	return left.CompareTo(ctx, right)
}

//
// A time value.
//
type SQLTime struct {
	Time time.Time
}

func (st SQLTimestamp) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return st, nil
}

func (st SQLTimestamp) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	c, err := v.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	if cmp, ok := c.(SQLDate); ok {
		if st.Time.After(cmp.Time) {
			return 1, nil
		} else if st.Time.Before(cmp.Time) {
			return -1, nil
		}
	}
	// TODO: type sort order implementation
	return 0, nil
}

//
// A timestamp value
//
type SQLTimestamp struct {
	Time time.Time
}

func (st SQLTime) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return st, nil
}

func (st SQLTime) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	c, err := v.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	if cmp, ok := c.(SQLDate); ok {
		if st.Time.After(cmp.Time) {
			return 1, nil
		} else if st.Time.Before(cmp.Time) {
			return -1, nil
		}
	}
	// TODO: type sort order implementation
	return 0, nil
}

//
// A unary minus expression.
//
type SQLUnaryMinus struct {
	SQLValue // TODO: I think this should maybe be a SQLExpr instead...
}

func (um *SQLUnaryMinus) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	if val, ok := um.SQLValue.(SQLNumeric); ok {
		return SQLInt(-(round(val.Float64()))), nil
	}
	return um.SQLValue, nil
}

func (um *SQLUnaryMinus) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	left, err := um.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	right, err := v.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	return left.CompareTo(ctx, right)
}

//
// A unary plus expression.
//
type SQLUnaryPlus struct {
	SQLValue
}

func (up *SQLUnaryPlus) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	if val, ok := up.SQLValue.(SQLNumeric); ok {
		return SQLInt(round(val.Float64())), nil
	}
	return up.SQLValue, nil
}

func (up *SQLUnaryPlus) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	left, err := up.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	right, err := v.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	return left.CompareTo(ctx, right)
}

//
// A tilde unary expression.
//
type SQLUnaryTilde struct {
	SQLValue
}

func (td *SQLUnaryTilde) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	if val, ok := td.SQLValue.(SQLNumeric); ok {
		return SQLInt(^round(val.Float64())), nil
	}
	return td.SQLValue, nil
}

func (td *SQLUnaryTilde) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	left, err := td.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	right, err := v.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	return left.CompareTo(ctx, right)
}

//
// A tuple value.
//
type SQLValTupleValue SQLValues

func (te SQLValTupleValue) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	var values []SQLValue

	for _, v := range te.Values {
		value, err := v.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}

	return SQLValues{values}, nil
}

func (te SQLValTupleValue) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	left, err := te.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	right, err := v.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	return left.CompareTo(ctx, right)
}

//
// Multiple SQL Values.
//
type SQLValues struct {
	Values []SQLValue
}

func (sv SQLValues) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return sv, nil
}

func (sv SQLValues) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {

	r, ok := v.(SQLValues)
	if !ok {
		//
		// allows for implicit row value comparisons such as:
		//
		// select a, b from foo where (a) < 3;
		//
		if len(sv.Values) != 1 {
			return 1, fmt.Errorf("Operand should contain %v columns", len(sv.Values))
		}
		r.Values = append(r.Values, v)
	} else if len(sv.Values) != len(r.Values) {
		return 1, fmt.Errorf("Operand should contain %v columns", len(sv.Values))
	}

	for i := 0; i < len(sv.Values); i++ {

		c, err := sv.Values[i].CompareTo(ctx, r.Values[i])
		if err != nil {
			return 1, err
		}

		if c != 0 {
			return c, nil
		}

	}

	return 0, nil
}

//
// An unsigned 32-bit integer.
//
type SQLUint32 uint32

func (su SQLUint32) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return su, nil
}

func (su SQLUint32) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	c, err := v.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	if n, ok := c.(SQLUint32); ok {
		cmp := uint32(su) - uint32(n)
		if cmp > 0 {
			return 1, nil
		} else if cmp < 0 {
			return -1, nil
		}
		return 0, nil
	} else if n, ok := c.(SQLFloat); ok {
		cmp := su.Float64() - n.Float64()
		if cmp > 0 {
			return 1, nil
		} else if cmp < 0 {
			return -1, nil
		}
		return 0, nil
	}
	return -1, nil
}

func (su SQLUint32) Add(o SQLNumeric) SQLNumeric {
	if oi, ok := o.(SQLUint32); ok {
		return SQLUint32(uint32(su) + uint32(oi))
	}
	return SQLFloat(su.Float64() + o.Float64())
}

func (su SQLUint32) Float64() float64 {
	return float64(su)
}

func (su SQLUint32) Product(o SQLNumeric) SQLNumeric {
	if oi, ok := o.(SQLUint32); ok {
		return SQLUint32(uint32(su) * uint32(oi))
	}
	return SQLFloat(su.Float64() * o.Float64())
}

func (su SQLUint32) Sub(o SQLNumeric) SQLNumeric {
	if oi, ok := o.(SQLUint32); ok {
		return SQLUint32(uint32(su) - uint32(oi))
	}
	return SQLFloat(su.Float64() - o.Float64())
}

//
// round returns the closest integer value to the float - round half down
// for negative values and round half up otherwise.
func round(f float64) int64 {
	v := f

	if v < 0.0 {
		v += 0.5
	}

	if f < 0 && v == math.Floor(v) {
		return int64(v - 1)
	}

	return int64(math.Floor(v))
}
