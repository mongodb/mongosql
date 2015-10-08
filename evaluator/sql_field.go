package evaluator

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"gopkg.in/mgo.v2/bson"
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

func (sf SQLFloat) Float64() float64 {
	return float64(sf)
}

func (si SQLInt) Float64() float64 {
	return float64(si)
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

func (sf SQLFloat) Product(o SQLNumeric) SQLNumeric {
	return SQLFloat(float64(sf) * o.Float64())
}

func (si SQLInt) Product(o SQLNumeric) SQLNumeric {
	if oi, ok := o.(SQLInt); ok {
		return SQLInt(int64(si) * int64(oi))
	}
	return SQLFloat(si.Float64() * o.Float64())
}

func (si SQLInt) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return si, nil
}
func (sf SQLFloat) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return sf, nil
}

func (sf SQLFloat) MongoValue() interface{} {
	return float64(sf)
}

func (si SQLInt) MongoValue() interface{} {
	return int64(si)
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
	// can only compare numbers to each other, otherwise treat as error
	return -1, ErrTypeMismatch
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
	// can only compare numbers to each other, otherwise treat as error
	return -1, ErrTypeMismatch
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
// SQLFuncExpr
//
type SQLFuncExpr struct {
	*sqlparser.FuncExpr
}

func (f *SQLFuncExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
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

func (f *SQLFuncExpr) MongoValue() interface{} {
	return nil
}

func (f *SQLFuncExpr) CompareTo(ctx *EvalCtx, value SQLValue) (int, error) {
	fEval, err := f.Evaluate(ctx)
	if err != nil {
		return 0, err
	}

	return fEval.CompareTo(ctx, value)
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
// SQLParenBoolExpr
//
type SQLParenBoolExpr struct {
	*sqlparser.ParenBoolExpr
}

func (p *SQLParenBoolExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	expr, ok := p.Expr.(sqlparser.Expr)

	if !ok {
		return nil, fmt.Errorf("could not convert ParenBoolExpr Expr to Expr")
	}

	matcher, err := BuildMatcher(expr)
	if err != nil {
		return nil, err
	}

	return SQLBool(matcher.Matches(ctx)), nil
}

func (p *SQLParenBoolExpr) MongoValue() interface{} {
	return nil
}

func (p *SQLParenBoolExpr) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	return 0, nil
}
