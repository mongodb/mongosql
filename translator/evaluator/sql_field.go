package evaluator

import (
	"fmt"
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
		return SQLNumeric(float64(v)), nil
	case int32: // NumberInt
		return SQLNumeric(float64(v)), nil
	case float64:
		return SQLNumeric(float64(v)), nil
	case float32:
		return SQLNumeric(float64(v)), nil
	case int64: // NumberLong
		return SQLNumeric(float64(v)), nil
	default:
		panic(fmt.Errorf("can't convert this type to a SQLValue: %T", v))
	}
}

func (sf SQLField) MongoValue() interface{} {
	panic("can't get the mongo value of a field reference.")
}

func (sqlf SQLField) Evaluate(ctx *EvalCtx) SQLValue {
	// TODO how do we report field not existing? do we just treat is a NULL, or something else?
	for _, row := range ctx.Rows {
		for _, data := range row.Data {
			if data.Table == sqlf.tableName {
				if value, hasValue := row.GetField(sqlf.tableName, sqlf.fieldName); hasValue {
					val, err := NewSQLField(value)
					if err != nil {
						panic(err)
					}
					return val
				}
				// field does not exist - return null i guess
				return SQLNull
			}
		}
	}
	return SQLNull
}

func (sqlf SQLField) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	left := sqlf.Evaluate(ctx)
	right := v.Evaluate(ctx)
	return left.CompareTo(ctx, right)
}

//
// SQLNull
//
type SQLNullValue struct{}

var SQLNull = SQLNullValue{}

func (nv SQLNullValue) Evaluate(ctx *EvalCtx) SQLValue {
	return nv
}

func (nv SQLNullValue) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	c := v.Evaluate(ctx)
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
type SQLNumeric float64

func (sn SQLNumeric) Evaluate(_ *EvalCtx) SQLValue {
	return sn
}

func (sn SQLNumeric) MongoValue() interface{} {
	return float64(sn)
}

func (sn SQLNumeric) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	c := v.Evaluate(ctx)
	if n, ok := c.(SQLNumeric); ok {
		return int(float64(sn) - float64(n)), nil
	}
	// can only compare numbers to each other, otherwise treat as error
	return -1, ErrTypeMismatch
}

//
// SQLString
//
type SQLString string

func (ss SQLString) Evaluate(_ *EvalCtx) SQLValue {
	return ss
}

func (ss SQLString) MongoValue() interface{} {
	return string(ss)
}

func (sn SQLString) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	c := v.Evaluate(ctx)
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

func (sb SQLBool) Evaluate(ctx *EvalCtx) SQLValue {
	return sb
}

func (sb SQLBool) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	c := v.Evaluate(ctx)
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
