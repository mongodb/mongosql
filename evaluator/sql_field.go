package evaluator

import (
	"fmt"
	"github.com/10gen/sqlproxy/config"
	"github.com/erh/mixer/sqlparser"
	"github.com/mongodb/mongo-tools/common/util"
	"gopkg.in/mgo.v2/bson"
	"math"
	"strconv"
	"strings"
	"time"
)

//
// SQLField
//
type SQLField struct {
	tableName string
	fieldName string
}

func NewSQLField(value interface{}, columnType config.ColumnType) (SQLValue, error) {

	if value == nil {
		return SQLNullValue{}, nil
	}

	if columnType == "" {
		switch v := value.(type) {
		case SQLValue:
			return v, nil
		case nil:
			return SQLNull, nil
		case bson.ObjectId:
			// TODO: handle this a special type? just using a string for now.
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
		case uint32:
			return SQLUint32(v), nil
		case time.Time:
			return SQLTimestamp{v}, nil
		default:
			panic(fmt.Errorf("can't convert this type to a SQLValue: %T", v))
		}
	}

	switch columnType {
	case config.SQLString:
		switch v := value.(type) {
		case bool:
			return SQLString(strconv.FormatBool(v)), nil
		case string:
			return SQLString(v), nil
		case float64:
			return SQLString(strconv.FormatFloat(v, 'f', -1, 64)), nil
		case int:
			return SQLString(strconv.FormatInt(int64(v), 10)), nil
		case int64:
			return SQLString(strconv.FormatInt(v, 10)), nil
		}
	case config.SQLInt:
		switch v := value.(type) {
		case bool:
			if v {
				return SQLInt(1), nil
			}
			return SQLInt(0), nil
		case string:
			eval, err := strconv.Atoi(v)
			if err == nil {
				return SQLInt(eval), nil
			}
			if strings.Trim(v, " ") == "" {
				return SQLNullValue{}, nil
			}
		case int, int32, int64, float64:
			eval, err := util.ToInt(v)
			if err == nil {
				return SQLInt(eval), nil
			}
		}
	case config.SQLFloat:
		switch v := value.(type) {
		case bool:
			if v {
				return SQLFloat(1), nil
			}
			return SQLFloat(0), nil
		case string:
			eval, err := strconv.Atoi(v)
			if err == nil {
				return SQLFloat(float64(eval)), nil
			}
			if strings.Trim(v, " ") == "" {
				return SQLNullValue{}, nil
			}
		case int, int32, int64, float64:
			eval, err := util.ToFloat64(v)
			if err == nil {
				return SQLFloat(eval), nil
			}
		}
	case config.SQLDatetime:
		v, ok := value.(time.Time)
		if ok {
			return SQLDateTime{v}, nil
		}
	case config.SQLTimestamp:
		v, ok := value.(time.Time)
		if ok {
			return SQLTimestamp{v}, nil
		}
	case config.SQLTime:
		v, ok := value.(time.Time)
		if ok {
			return SQLTime{v}, nil
		}
	case config.SQLDate:
		v, ok := value.(time.Time)
		if ok {
			return SQLDate{v}, nil
		}
	default:
		return nil, fmt.Errorf("unknown column type %v", columnType)
	}
	return nil, fmt.Errorf("unable to convert '%v' (%T) to %v", value, value, columnType)

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
	return 1, nil
}

//
// SQLTemporal
//
type SQLTemporal interface {
	SQLValue
}

type SQLDateTime struct {
	Time time.Time
}
type SQLTimestamp struct {
	Time time.Time
}
type SQLDate struct {
	Time time.Time
}

type SQLTime struct {
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
	return 1, ErrTypeMismatch
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
	// TODO: support comparing with SQLInt, SQLFloat, etc
	return 1, ErrTypeMismatch
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

	matcher := NullMatcher{val}
	result, err := matcher.Matches(ctx)
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

	matcher := &Not{&BoolMatcher{notVal}}
	result, err := matcher.Matches(ctx)
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
		return nil, fmt.Errorf("function '%v' is not supported", string(f.Name))
	}
}

func (f *SQLAggFuncValue) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
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

func (f *SQLAggFuncValue) avgFunc(ctx *EvalCtx, distinctMap map[interface{}]bool) (SQLValue, error) {
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
				val, err := NewSQLValue(e.Expr)
				if err != nil {
					return nil, err
				}
				eval, err := val.Evaluate(evalCtx)
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

func (f *SQLAggFuncValue) sumFunc(ctx *EvalCtx, distinctMap map[interface{}]bool) (SQLValue, error) {
	var sum SQLNumeric = SQLInt(0)
	for _, row := range ctx.Rows {
		evalCtx := &EvalCtx{Rows: []Row{row}}
		for _, sExpr := range f.Exprs {
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

func (f *SQLAggFuncValue) countFunc(ctx *EvalCtx, distinctMap map[interface{}]bool) (SQLValue, error) {
	var count int64
	for _, row := range ctx.Rows {
		evalCtx := &EvalCtx{Rows: []Row{row}}
		for _, sExpr := range f.Exprs {
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

func (f *SQLAggFuncValue) minFunc(ctx *EvalCtx) (SQLValue, error) {
	var min SQLValue
	for _, row := range ctx.Rows {
		evalCtx := &EvalCtx{Rows: []Row{row}}
		for _, sExpr := range f.Exprs {
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

func (f *SQLAggFuncValue) maxFunc(ctx *EvalCtx) (SQLValue, error) {
	var max SQLValue
	for _, row := range ctx.Rows {
		evalCtx := &EvalCtx{Rows: []Row{row}}
		for _, sExpr := range f.Exprs {
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
// SQLValues
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

func (um *UMinus) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
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

func (up *UPlus) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
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

func (td *Tilda) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
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

// SubqueryValue returns true if any result is returned from the subquery.
type SubqueryValue struct {
	stmt sqlparser.SelectStatement
}

func (sv *SubqueryValue) Evaluate(ctx *EvalCtx) (value SQLValue, err error) {

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

func (sv *SubqueryValue) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
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
