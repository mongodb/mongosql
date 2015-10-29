package evaluator

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"gopkg.in/mgo.v2/bson"
	"math"
)

//
// SQLField
//
type SQLField struct {
	tableName string
	fieldName string
}

func NewSQLField(value interface{}) (SQLValue, error) {
	switch v := value.(type) {
	case SQLValue:
		return v, nil
	case nil:
		return SQLNull, nil
	case bson.ObjectId: // ObjectId
		//TODO handle this a special type? just using a string for now.
		return SQLString(v.Hex()), nil
	case bool:
		return SQLBool(v), nil
	case string:
		return SQLString(v), nil

	// TODO - handle overflow/precision of numeric types!
	case int:
		return SQLInt(v), nil
	case int32: // NumberInt
		return SQLInt(int64(v)), nil
	case float64:
		return SQLFloat(float64(v)), nil
	case float32:
		return SQLFloat(float64(v)), nil
	case int64: // NumberLong
		return SQLInt(v), nil
	default:
		panic(fmt.Errorf("can't convert this type to a SQLValue: %T", v))
	}
}

func (sf SQLField) MongoValue() interface{} {
	panic("can't get the mongo value of a field reference.")
}

func (sqlf SQLField) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	// TODO how do we report field not existing? do we just treat is a NULL, or something else?
	for _, row := range ctx.Rows {
		for _, data := range row.Data {
			if data.Table == sqlf.tableName {
				if value, hasValue := row.GetField(sqlf.tableName, sqlf.fieldName); hasValue {
					val, err := NewSQLField(value)
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
// SQLCaseValue holds a number of cases to evaluate as well as the value
// to return if any of the cases is matched. If none is matched,
// 'elseValue' is evaluated and returned.
//
type SQLCaseValue struct {
	elseValue      SQLValue
	caseConditions []caseCondition
}

// caseCondition holds a matcher used in evaluating case expressions and
// a value to return if a particular case is matched. If a case ia matched,
// the corresponding 'then' value is evaluated and returned ('then'
// corresponds to the 'then' clause in a case expression).
type caseCondition struct {
	matcher Matcher
	then    SQLValue
}

func (s SQLCaseValue) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	for _, condition := range s.caseConditions {

		m, err := condition.matcher.Matches(ctx)
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
	c, err := v.Evaluate(ctx)
	if err != nil {
		return 0, err
	}

	return s.CompareTo(ctx, c)
}

func (s SQLCaseValue) MongoValue() interface{} {
	return nil
}

//
// SQLNull
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
	return -1, nil
}

func (sn SQLNullValue) MongoValue() interface{} {
	return nil
}

//
// SQLNumeric
//
type SQLFloat float64
type SQLInt int64
type SQLUint32 uint32

type SQLNumeric interface {
	SQLValue
	Add(o SQLNumeric) SQLNumeric
	Sub(o SQLNumeric) SQLNumeric
	Product(o SQLNumeric) SQLNumeric
	Float64() float64
}

func (sf SQLFloat) Add(o SQLNumeric) SQLNumeric {
	return SQLFloat(float64(sf) + o.Float64())
}

func (si SQLInt) Add(o SQLNumeric) SQLNumeric {
	if oi, ok := o.(SQLInt); ok {
		return SQLInt(int64(si) + int64(oi))
	}
	return SQLFloat(si.Float64() + o.Float64())
}

func (su SQLUint32) Add(o SQLNumeric) SQLNumeric {
	if oi, ok := o.(SQLUint32); ok {
		return SQLUint32(uint32(su) + uint32(oi))
	}
	return SQLFloat(su.Float64() + o.Float64())
}

func (sf SQLFloat) Float64() float64 {
	return float64(sf)
}

func (si SQLInt) Float64() float64 {
	return float64(si)
}

func (su SQLUint32) Float64() float64 {
	return float64(su)
}

func (sf SQLFloat) Sub(o SQLNumeric) SQLNumeric {
	return SQLFloat(float64(sf) - o.Float64())
}

func (si SQLInt) Sub(o SQLNumeric) SQLNumeric {
	if oi, ok := o.(SQLInt); ok {
		return SQLInt(int64(si) - int64(oi))
	}
	return SQLFloat(si.Float64() - o.Float64())
}

func (su SQLUint32) Sub(o SQLNumeric) SQLNumeric {
	if oi, ok := o.(SQLUint32); ok {
		return SQLUint32(uint32(su) - uint32(oi))
	}
	return SQLFloat(su.Float64() - o.Float64())
}

func (sf SQLFloat) Product(o SQLNumeric) SQLNumeric {
	return SQLFloat(float64(sf) * o.Float64())
}

func (si SQLInt) Product(o SQLNumeric) SQLNumeric {
	if oi, ok := o.(SQLInt); ok {
		return SQLInt(int64(si) * int64(oi))
	}
	return SQLFloat(si.Float64() * o.Float64())
}

func (su SQLUint32) Product(o SQLNumeric) SQLNumeric {
	if oi, ok := o.(SQLUint32); ok {
		return SQLUint32(uint32(su) * uint32(oi))
	}
	return SQLFloat(su.Float64() * o.Float64())
}

func (si SQLInt) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return si, nil
}
func (sf SQLFloat) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return sf, nil
}
func (su SQLUint32) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return su, nil
}

func (sf SQLFloat) MongoValue() interface{} {
	return float64(sf)
}

func (si SQLInt) MongoValue() interface{} {
	return int64(si)
}

func (su SQLUint32) MongoValue() interface{} {
	return uint32(su)
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

//
// SQLString
//
type SQLString string

func (ss SQLString) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return ss, nil
}

func (ss SQLString) MongoValue() interface{} {
	return string(ss)
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
	return -1, ErrTypeMismatch
}

//
// SQLBool
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
	// can only compare bool to a bool, otherwise treat as error
	return -1, ErrTypeMismatch
}

func (sb SQLBool) MongoValue() interface{} {
	return bool(sb)
}

//
// SQLScalarFuncValue
//
type SQLScalarFuncValue struct {
	*sqlparser.FuncExpr
}

func (f *SQLScalarFuncValue) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	switch string(f.Name) {
	// connector functions
	case "connection_id":
		return connectionIdFunc(ctx)
	case "database":
		return dbFunc(ctx)

		// scalar functions
	case "isnull":
		return isNullFunc(ctx, f.Exprs)
	case "not":
		return notFunc(ctx, f.Exprs)
	case "pow":
		return powFunc(ctx, f.Exprs)

	default:
		return nil, fmt.Errorf("function '%v' is not supported", string(f.Name))
	}
}

func (f *SQLScalarFuncValue) MongoValue() interface{} {
	return nil
}

func (f *SQLScalarFuncValue) CompareTo(ctx *EvalCtx, value SQLValue) (int, error) {
	fEval, err := f.Evaluate(ctx)
	if err != nil {
		return 0, err
	}

	return fEval.CompareTo(ctx, value)
}

//
// SQLAggFuncValue
//
type SQLAggFuncValue struct {
	*sqlparser.FuncExpr
}

func (f *SQLAggFuncValue) Evaluate(ctx *EvalCtx) (SQLValue, error) {
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
		return nil, fmt.Errorf("function '%v' is not supported", string(f.Name))
	}
}

func (f *SQLAggFuncValue) MongoValue() interface{} {
	return nil
}

func (f *SQLAggFuncValue) CompareTo(ctx *EvalCtx, value SQLValue) (int, error) {
	fEval, err := f.Evaluate(ctx)
	if err != nil {
		return 0, err
	}

	return fEval.CompareTo(ctx, value)
}

func connectionIdFunc(ctx *EvalCtx) (SQLValue, error) {
	return SQLUint32(ctx.ExecCtx.ConnectionId()), nil
}

func dbFunc(ctx *EvalCtx) (SQLValue, error) {
	return SQLString(ctx.ExecCtx.DB()), nil
}

func isNullFunc(ctx *EvalCtx, sExprs sqlparser.SelectExprs) (SQLValue, error) {
	if len(sExprs) != 1 {
		return nil, fmt.Errorf("'isnull' function requires exactly one argument")
	}
	var exp sqlparser.Expr
	if v, ok := sExprs[0].(*sqlparser.NonStarExpr); ok {
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

	matcher := NullMatcher{val}
	result, err := matcher.Matches(ctx)
	if err != nil {
		return nil, err
	}
	return SQLBool(result), nil
}

func notFunc(ctx *EvalCtx, sExprs sqlparser.SelectExprs) (SQLValue, error) {
	if len(sExprs) != 1 {
		return nil, fmt.Errorf("'not' function requires exactly one argument")
	}
	var notExpr sqlparser.Expr
	if v, ok := sExprs[0].(*sqlparser.NonStarExpr); ok {
		notExpr = v.Expr
	} else {
		return nil, fmt.Errorf("argument to 'not' function can not contain '*'")
	}

	notVal, err := NewSQLValue(notExpr)
	if err != nil {
		return nil, err
	}

	matcher := &Not{&BoolMatcher{notVal}}
	result, err := matcher.Matches(ctx)
	if err != nil {
		return nil, err
	}
	return SQLBool(result), nil
}

func powFunc(ctx *EvalCtx, sExprs sqlparser.SelectExprs) (SQLValue, error) {
	if len(sExprs) != 2 {
		return nil, fmt.Errorf("'pow' function requires exactly two arguments")
	}
	var baseExpr, expExpr sqlparser.Expr
	if v, ok := sExprs[0].(*sqlparser.NonStarExpr); ok {
		baseExpr = v.Expr
	} else {
		return nil, fmt.Errorf("argument to 'pow' function can not contain '*'")
	}
	if v, ok := sExprs[1].(*sqlparser.NonStarExpr); ok {
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

func avgFunc(ctx *EvalCtx, sExprs sqlparser.SelectExprs, distinctMap map[interface{}]bool) (SQLValue, error) {
	var sum SQLNumeric = SQLInt(0)
	count := 0
	for _, row := range ctx.Rows {
		evalCtx := &EvalCtx{Rows: []Row{row}}
		for _, sExpr := range sExprs {
			switch e := sExpr.(type) {
			// mixture of star and non-star expression is acceptable
			case *sqlparser.StarExpr:
				return nil, fmt.Errorf("avg aggregate function can not contain '*'")
			case *sqlparser.NonStarExpr:
				val, err := NewSQLValue(e.Expr)
				if err != nil {
					return nil, err
				}
				eval, err := val.Evaluate(evalCtx)
				if err != nil {
					return nil, err
				}
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
					sum = sum.Add(n)
				}
			default:
				return nil, fmt.Errorf("unknown expression in avgFunc: %T", e)
			}
		}
	}

	return SQLFloat(sum.Float64() / float64(count)), nil
}

func sumFunc(ctx *EvalCtx, sExprs sqlparser.SelectExprs, distinctMap map[interface{}]bool) (SQLValue, error) {
	var sum SQLNumeric = SQLInt(0)
	for _, row := range ctx.Rows {
		evalCtx := &EvalCtx{Rows: []Row{row}}
		for _, sExpr := range sExprs {
			switch e := sExpr.(type) {
			// mixture of star and non-star expression is acceptable
			case *sqlparser.StarExpr:
				return nil, fmt.Errorf("sum aggregate function can not contain '*'")
			case *sqlparser.NonStarExpr:
				val, err := NewSQLValue(e.Expr)
				if err != nil {
					return nil, err
				}
				eval, err := val.Evaluate(evalCtx)
				if err != nil {
					return nil, err
				}

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
					sum = sum.Add(n)
				}
			default:
				return nil, fmt.Errorf("unknown expression in sumFunc: %T", e)
			}
		}
	}

	return sum, nil
}

func countFunc(ctx *EvalCtx, sExprs sqlparser.SelectExprs, distinctMap map[interface{}]bool) (SQLValue, error) {
	var count int64
	for _, row := range ctx.Rows {
		evalCtx := &EvalCtx{Rows: []Row{row}}
		for _, sExpr := range sExprs {
			switch e := sExpr.(type) {
			// mixture of star and non-star expression is acceptable
			case *sqlparser.StarExpr:
				count += 1

			case *sqlparser.NonStarExpr:
				val, err := NewSQLValue(e.Expr)
				if err != nil {
					return nil, err
				}
				eval, err := val.Evaluate(evalCtx)
				if err != nil {
					return nil, err
				}
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
	return SQLInt(count), nil
}

func minFunc(ctx *EvalCtx, sExprs sqlparser.SelectExprs) (SQLValue, error) {
	var min SQLValue
	for _, row := range ctx.Rows {
		evalCtx := &EvalCtx{Rows: []Row{row}}
		for _, sExpr := range sExprs {
			switch e := sExpr.(type) {
			// mixture of star and non-star expression is acceptable
			case *sqlparser.StarExpr:
				return nil, fmt.Errorf("min aggregate function can not contain '*'")
			case *sqlparser.NonStarExpr:
				val, err := NewSQLValue(e.Expr)
				if err != nil {
					return nil, err
				}
				eval, err := val.Evaluate(evalCtx)
				if err != nil {
					return nil, err
				}
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
				return nil, fmt.Errorf("unknown expression in minFunc: %T", e)
			}
		}
	}
	return min, nil
}

func maxFunc(ctx *EvalCtx, sExprs sqlparser.SelectExprs) (SQLValue, error) {
	var max SQLValue
	for _, row := range ctx.Rows {
		evalCtx := &EvalCtx{Rows: []Row{row}}
		for _, sExpr := range sExprs {
			switch e := sExpr.(type) {
			// mixture of star and non-star expression is acceptable
			case *sqlparser.StarExpr:
				return nil, fmt.Errorf("max aggregate function can not contain '*'")
			case *sqlparser.NonStarExpr:
				val, err := NewSQLValue(e.Expr)
				if err != nil {
					return nil, err
				}
				eval, err := val.Evaluate(evalCtx)
				if err != nil {
					return nil, err
				}
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
				return nil, fmt.Errorf("unknown expression in maxFunc: %T", e)
			}
		}
	}
	return max, nil
}

//
// SQLParenBoolValue
//
type SQLParenBoolValue struct {
	*sqlparser.ParenBoolExpr
}

func (p *SQLParenBoolValue) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	expr, ok := p.Expr.(sqlparser.Expr)

	if !ok {
		return nil, fmt.Errorf("could not convert ParenBoolExpr Expr to Expr")
	}

	matcher, err := BuildMatcher(expr)
	if err != nil {
		return nil, err
	}

	b, err := matcher.Matches(ctx)
	return SQLBool(b), err
}

func (p *SQLParenBoolValue) MongoValue() interface{} {
	return nil
}

func (p *SQLParenBoolValue) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	return 0, nil
}

//
// SQLValues
//
type SQLValues struct {
	Values []SQLValue
}

func (sv SQLValues) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return sv, nil
}

func (sv SQLValues) MongoValue() interface{} {
	return nil
}

func (sv SQLValues) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {

	r, ok := v.(SQLValues)
	if !ok {
		if len(sv.Values) != 1 {
			return -1, fmt.Errorf("Operand should contain %v columns", len(sv.Values))
		}
		// allows for implicit row value comparisons such as:
		//
		// select a, b from foo where (a) < 3;
		//
		//
		r.Values = append(r.Values, v)
	} else if len(sv.Values) != len(r.Values) {
		return -1, fmt.Errorf("Operand should contain %v columns", len(sv.Values))
	}

	for i := 0; i < len(sv.Values); i++ {
		c, err := sv.Values[i].CompareTo(ctx, r.Values[i])
		if err != nil {
			return -1, err
		}

		if c != 0 {
			return c, nil
		}

	}

	return 0, nil
}

//
// SQLValTupleValue
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

func (te SQLValTupleValue) MongoValue() interface{} {
	return nil
}

func (te SQLValTupleValue) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	return 0, nil
}

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

//
// UMinus
//

type UMinus struct {
	SQLValue
}

func (um *UMinus) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	if val, ok := um.SQLValue.(SQLNumeric); ok {
		return SQLInt(-(round(val.Float64()))), nil
	}
	return um.SQLValue, nil
}

func (um *UMinus) MongoValue() interface{} {
	return um.SQLValue.MongoValue()
}

func (um *UMinus) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	return um.CompareTo(ctx, v)
}

//
// UPlus
//

type UPlus struct {
	SQLValue
}

func (up *UPlus) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	if val, ok := up.SQLValue.(SQLNumeric); ok {
		return SQLInt(round(val.Float64())), nil
	}
	return up.SQLValue, nil
}

func (up *UPlus) MongoValue() interface{} {
	return up.SQLValue.MongoValue()
}

func (up *UPlus) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	return up.CompareTo(ctx, v)
}

//
// Tilda
//

type Tilda struct {
	SQLValue
}

func (td *Tilda) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	if val, ok := td.SQLValue.(SQLNumeric); ok {
		return SQLInt(^round(val.Float64())), nil
	}
	return td.SQLValue, nil
}

func (td *Tilda) MongoValue() interface{} {
	return td.SQLValue.MongoValue()
}

func (td *Tilda) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	return td.CompareTo(ctx, v)
}
