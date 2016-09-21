package evaluator

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/util"
	"github.com/shopspring/decimal"
	"gopkg.in/mgo.v2/bson"
)

// NewSQLValue is a factory method for creating a SQLValue from
// an in-memory value and column type.
//
// For a full description of its behavior, please see http://bit.ly/2bhWuyw
func NewSQLValue(value interface{}, sqlType schema.SQLType) SQLValue {

	if value == nil {
		return SQLNull
	}

	if v, ok := value.(SQLValue); ok {
		return v
	}

	switch sqlType {

	case schema.SQLBoolean:
		v := NewSQLValue(value, schema.SQLFloat)
		if v.Float64() == 0 {
			return SQLFalse
		}
		return SQLTrue

	case schema.SQLDate:
		switch v := value.(type) {
		case bson.ObjectId:
			return SQLDate{v.Time()}
		case string:
			for _, format := range schema.TimestampCtorFormats {
				d, err := time.Parse(format, v)
				if err == nil {
					date := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, schema.DefaultLocale)
					return SQLDate{date}
				}
			}
			return SQLDate{time.Time{}}
		case time.Time:
			v = v.In(schema.DefaultLocale)
			date := time.Date(v.Year(), v.Month(), v.Day(), 0, 0, 0, 0, schema.DefaultLocale)
			return SQLDate{date}
		default:
			return SQLDate{time.Time{}}
		}

	case schema.SQLDecimal128:
		switch v := value.(type) {
		case bool:
			if v {
				return SQLDecimal128(decimal.NewFromFloat(1))
			}
			return SQLDecimal128(decimal.NewFromFloat(0))
		case bson.ObjectId:
			dec, _ := decimal.NewFromString(v.String())
			return SQLDecimal128(dec)
		case bson.Decimal128:
			dec, _ := decimal.NewFromString(v.String())
			return SQLDecimal128(dec)
		case float32:
			return SQLDecimal128(decimal.NewFromFloat(float64(v)))
		case float64:
			return SQLDecimal128(decimal.NewFromFloat(v))
		case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64:
			flt, _ := util.ToFloat64(v)
			return SQLDecimal128(decimal.NewFromFloat(flt))
		case string:
			dec, err := decimal.NewFromString(v)
			if err == nil {
				return SQLDecimal128(dec)
			}
			value = NewSQLValue(value, schema.SQLFloat)
			flt, _ := value.(SQLFloat)
			return SQLDecimal128(decimal.NewFromFloat(float64(flt)))
		case time.Time:
			h, m, s := v.Clock()

			if h+m+s == 0 {
				flt, _ := strconv.ParseFloat(v.Format("20060102"), 64)
				return SQLDecimal128(decimal.NewFromFloat(flt))
			} else if v.Year() == 0 {
				flt, _ := strconv.ParseFloat(v.Format("150405"), 64)
				return SQLDecimal128(decimal.NewFromFloat(flt))
			}
			flt, _ := strconv.ParseFloat(v.Format("20060102150405"), 64)
			return SQLDecimal128(decimal.NewFromFloat(flt))
		default:
			return SQLDecimal128(decimal.Zero)
		}

	case schema.SQLFloat, schema.SQLNumeric, schema.SQLArrNumeric:
		switch v := value.(type) {
		case bool:
			if v {
				return SQLFloat(1)
			}
			return SQLFloat(0)
		case bson.ObjectId:
			return NewSQLValue(v.Time(), schema.SQLFloat)
		case bson.Decimal128:
			dec, _ := decimal.NewFromString(v.String())
			flt, _ := dec.Float64()
			return SQLFloat(flt)
		case float32:
			return SQLFloat(float64(v))
		case float64:
			return SQLFloat(v)
		case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64:
			flt, _ := util.ToFloat64(v)
			return SQLFloat(flt)
		case string:
			flt, _ := strconv.ParseFloat(v, 64)
			return SQLFloat(flt)
		case time.Time:
			h, m, s := v.Clock()

			if h+m+s == 0 {
				flt, _ := strconv.ParseFloat(v.Format("20060102"), 64)
				return SQLFloat(flt)
			} else if v.Year() == 0 {
				flt, _ := strconv.ParseFloat(v.Format("150405"), 64)
				return SQLFloat(flt)
			}
			flt, _ := strconv.ParseFloat(v.Format("20060102150405"), 64)
			return SQLFloat(flt)
		default:
			return SQLFloat(0.0)
		}

	case schema.SQLInt, schema.SQLInt64:
		flt := NewSQLValue(value, schema.SQLFloat)
		return SQLInt(flt.Int64())

	case schema.SQLUint64:
		flt := NewSQLValue(value, schema.SQLFloat)
		return SQLUint64(flt.Uint64())

	case schema.SQLObjectID:
		switch v := value.(type) {
		case bool:
			if v {
				return SQLObjectID("1")
			}
			return SQLObjectID("0")
		case bson.ObjectId:
			return SQLObjectID(v.Hex())
		case bson.Decimal128:
			return SQLObjectID(v.String())
		case float32, float64:
			flt, _ := util.ToFloat64(v)
			return SQLObjectID(strconv.FormatFloat(flt, 'f', -1, 64))
		case int, int8, int16, int32, uint8, uint16, uint32:
			val, _ := util.ToInt(v)
			return SQLObjectID(strconv.FormatInt(int64(val), 10))
		case int64:
			return SQLObjectID(strconv.FormatInt(int64(v), 10))
		case string:
			return SQLObjectID(v)
		case uint64:
			return SQLObjectID(strconv.FormatUint(v, 10))
		case time.Time:
			return SQLObjectID(bson.NewObjectIdWithTime(v).Hex())
		default:
			return SQLObjectID("")
		}

	case schema.SQLTimestamp:
		switch v := value.(type) {
		case bson.ObjectId:
			return SQLTimestamp{v.Time()}
		case string:
			for _, format := range schema.TimestampCtorFormats {
				d, err := time.Parse(format, v)
				if err == nil {
					ts := time.Date(d.Year(), d.Month(), d.Day(), d.Hour(),
						d.Minute(), d.Second(), d.Nanosecond(), schema.DefaultLocale)
					return SQLTimestamp{ts}
				}
			}
			return SQLTimestamp{time.Time{}}
		case time.Time:
			return SQLTimestamp{v.In(schema.DefaultLocale)}
		default:
			return SQLTimestamp{time.Time{}}
		}

	case schema.SQLUUID:
		v, _ := NewSQLValueFromUUID(value, sqlType, schema.MongoUUID)
		return v

	case schema.SQLVarchar:
		switch v := value.(type) {
		case bool:
			if v {
				return SQLVarchar("1")
			}
			return SQLVarchar("0")
		case bson.ObjectId:
			return SQLObjectID(v.Hex())
		case bson.Decimal128:
			dec, _ := decimal.NewFromString(v.String())
			return SQLDecimal128(dec)
		case float32:
			return SQLVarchar(strconv.FormatFloat(float64(v), 'f', -1, 32))
		case float64:
			return SQLVarchar(strconv.FormatFloat(v, 'f', -1, 64))
		case uint8, uint16, uint32, uint64, int, int8, int16, int32, int64:
			val, _ := util.ToInt(v)
			return SQLVarchar(strconv.FormatInt(int64(val), 10))
		case string:
			return SQLVarchar(v)
		case time.Time:
			return SQLVarchar(v.String())
		default:
			return SQLVarchar(reflect.ValueOf(v).String())
		}
	}

	panic(fmt.Errorf("can't convert this type to a SQLValue: %T", value))

}

// NewSQLValueFromUUID is a factory method for creating a SQLUUID
// from a given value
func NewSQLValueFromUUID(value interface{}, sqlType schema.SQLType, mongoType schema.MongoType) (SQLValue, error) {

	err := fmt.Errorf("unable to convert '%v' (%T) to %v", value, value, sqlType)

	if !isUUID(mongoType) {
		return nil, err
	}

	if sqlType != schema.SQLVarchar {
		return nil, err
	}

	switch v := value.(type) {
	case bson.Binary:
		if err := normalizeUUID(mongoType, v.Data); err != nil {
			return nil, err
		}

		return SQLUUID{mongoType, v.Data}, nil
	case string:
		b, ok := getBinaryFromExpr(mongoType, SQLVarchar(v))
		if !ok {
			return nil, err
		}
		return SQLUUID{mongoType, b.Data}, nil
	}

	return nil, err
}

// NewSQLValueFromSQLColumnExpr is a factory method for creating a SQLValue
// from a given column expression value. It converts the value to the appropriate
// SQLValue using the provided SQLType and MongoType.
func NewSQLValueFromSQLColumnExpr(value interface{}, sqlType schema.SQLType, mongoType schema.MongoType) (SQLValue, error) {
	if value == nil {
		return SQLNull, nil
	}

	if v, ok := value.(SQLValue); ok {
		return v, nil
	}

	if sqlType == schema.SQLNone {
		switch v := value.(type) {
		case nil:
			return SQLNull, nil
		case bson.ObjectId:
			return SQLObjectID(v.Hex()), nil
		case bson.Decimal128:
			d, err := decimal.NewFromString(v.String())
			return SQLDecimal128(d), err
		case bool:
			return SQLBool(v), nil
		case decimal.Decimal:
			return SQLDecimal128(v), nil
		case string:
			return SQLVarchar(v), nil
		case float32, float64:
			eval, _ := util.ToFloat64(v)
			return SQLFloat(eval), nil
		case int, int8, int16, int32, int64:
			eval, _ := util.ToInt(v)
			return SQLInt(eval), nil
		case uint8:
			return SQLUint64(uint64(v)), nil
		case uint16:
			return SQLUint64(uint64(v)), nil
		case uint32:
			return SQLUint64(uint64(v)), nil
		case uint64:
			return SQLUint64(uint64(v)), nil
		case time.Time:
			h := v.Hour()
			mi := v.Minute()
			sd := v.Second()
			ns := v.Nanosecond()

			if ns == 0 && sd == 0 && mi == 0 && h == 0 {
				return NewSQLValueFromSQLColumnExpr(v, schema.SQLDate, mongoType)
			}

			// SQLDateTime and SQLTimestamp can be handled similarly
			return NewSQLValueFromSQLColumnExpr(v, schema.SQLTimestamp, mongoType)
		default:
			panic(fmt.Errorf("can't convert this type to a SQLValue: %T", v))
		}
	}

	switch sqlType {
	case schema.SQLObjectID:
		switch v := value.(type) {
		case string:
			return SQLObjectID(v), nil
		case bson.ObjectId:
			return SQLObjectID(v.Hex()), nil
		case nil:
			return SQLNull, nil
		}
	case schema.SQLInt, schema.SQLInt64:
		switch v := value.(type) {
		case bool:
			if v {
				return SQLInt(1), nil
			}
			return SQLInt(0), nil
		case bson.Decimal128:
			d, err := decimal.NewFromString(v.String())
			return SQLDecimal128(d), err
		case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64:
			eval, err := util.ToInt(v)
			if err == nil {
				return SQLInt(eval), nil
			}
		case time.Time:
			h, m, s := v.Clock()
			// Date, otherwise timestamp
			if h+m+s == 0 {
				eval, _ := strconv.ParseInt(v.Format("20060102"), 10, 64)
				return SQLInt(eval), nil
			} else if v.Year() == 0 {
				eval, _ := strconv.ParseInt(v.Format("150405"), 10, 64)
				return SQLInt(eval), nil
			}
			eval, _ := strconv.ParseInt(v.Format("20060102150405"), 10, 64)
			return SQLInt(eval), nil
		case nil:
			return SQLNull, nil
		}
	case schema.SQLUint64:
		switch v := value.(type) {
		case bool:
			if v {
				return SQLUint64(1), nil
			}
			return SQLUint64(0), nil
		case bson.Decimal128:
			d, err := decimal.NewFromString(v.String())
			return SQLDecimal128(d), err
		case int, int8, int16, int32, int64, float32, float64:
			eval, err := util.ToInt(v)
			if err == nil {
				return SQLUint64(eval), nil
			}
		case uint8:
			return SQLUint64(uint64(v)), nil
		case uint16:
			return SQLUint64(uint64(v)), nil
		case uint32:
			return SQLUint64(uint64(v)), nil
		case uint64:
			return SQLUint64(v), nil
		case time.Time:
			h, m, s := v.Clock()
			// Date, otherwise timestamp
			if h+m+s == 0 {
				eval, _ := strconv.ParseInt(v.Format("20060102"), 10, 64)
				return SQLUint64(eval), nil
			} else if v.Year() == 0 {
				eval, _ := strconv.ParseInt(v.Format("150405"), 10, 64)
				return SQLUint64(eval), nil
			}
			eval, _ := strconv.ParseInt(v.Format("20060102150405"), 10, 64)
			return SQLUint64(eval), nil
		case nil:
			return SQLNull, nil
		}
	case schema.SQLVarchar:
		if isUUID(mongoType) {
			return NewSQLValueFromUUID(value, sqlType, mongoType)
		}

		switch v := value.(type) {
		case bool:
			if v {
				return SQLVarchar("1"), nil
			}
			return SQLVarchar("0"), nil
		case string:
			return SQLVarchar(v), nil
		case float32:
			return SQLVarchar(strconv.FormatFloat(float64(v), 'f', -1, 32)), nil
		case float64:
			return SQLVarchar(strconv.FormatFloat(v, 'f', -1, 64)), nil
		case uint8, uint16, uint32, uint64, int, int8, int16, int32, int64:
			eval, _ := util.ToInt(v)
			return SQLVarchar(strconv.FormatInt(int64(eval), 10)), nil
		case bson.ObjectId:
			return SQLObjectID(v.Hex()), nil
		case bson.Decimal128:
			d, err := decimal.NewFromString(v.String())
			return SQLDecimal128(d), err
		case time.Time:
			return SQLVarchar(v.String()), nil
		case nil:
			return SQLNull, nil
		default:
			return SQLVarchar(reflect.ValueOf(v).String()), nil
		}

	case schema.SQLBoolean:
		switch v := value.(type) {
		case bool:
			if v {
				return SQLTrue, nil
			}
			return SQLFalse, nil
		case string:
			return SQLFalse, nil
		case uint8, uint16, uint32, uint64, int, int8, int16, int32, int64, float32, float64:
			eval, _ := util.ToFloat64(v)
			if eval == 0 {
				return SQLFalse, nil
			}
			return SQLTrue, nil
		case time.Time:
			var eval int64
			h, m, s := v.Clock()
			// Date, time or timestamp.
			if h+m+s == 0 {
				val, _ := strconv.ParseInt(v.Format("20060102"), 10, 64)
				eval = val
			} else if v.Year() == 0 {
				val, _ := strconv.ParseInt(v.Format("150405"), 10, 64)
				eval = val
			} else {
				val, _ := strconv.ParseInt(v.Format("20060102150405"), 10, 64)
				eval = val
			}
			if eval == 0 {
				return SQLFalse, nil
			}
			return SQLTrue, nil
		case nil:
			return SQLNull, nil
		}

	case schema.SQLDecimal128:
		switch v := value.(type) {
		case bool:
			if v {
				return SQLDecimal128(decimal.NewFromFloat(1)), nil
			}
			return SQLDecimal128(decimal.Zero), nil
		case bson.Decimal128:
			d, err := decimal.NewFromString(v.String())
			return SQLDecimal128(d), err
		case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64:
			eval, err := util.ToFloat64(v)
			if err == nil {
				return SQLDecimal128(decimal.NewFromFloat(eval)), nil
			}
		case decimal.Decimal:
			return SQLDecimal128(v), nil
		case time.Time:
			h, m, s := v.Clock()
			// Date, otherwise timestamp
			if h+m+s == 0 {
				eval, _ := strconv.ParseFloat(v.Format("20060102"), 64)
				return SQLDecimal128(decimal.NewFromFloat(eval)), nil
			} else if v.Year() == 0 {
				eval, _ := strconv.ParseFloat(v.Format("150405"), 64)
				return SQLDecimal128(decimal.NewFromFloat(eval)), nil
			}
			eval, _ := strconv.ParseFloat(v.Format("20060102150405"), 64)
			return SQLDecimal128(decimal.NewFromFloat(eval)), nil
		case nil:
			return SQLNull, nil
		}

	case schema.SQLFloat, schema.SQLNumeric, schema.SQLArrNumeric:
		switch v := value.(type) {
		case bool:
			if v {
				return SQLFloat(1), nil
			}
			return SQLFloat(0), nil
		case bson.Decimal128:
			d, err := decimal.NewFromString(v.String())
			return SQLDecimal128(d), err
		case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64:
			eval, err := util.ToFloat64(v)
			if err == nil {
				return SQLFloat(eval), nil
			}
		case time.Time:
			h, m, s := v.Clock()
			// Date, otherwise timestamp
			if h+m+s == 0 {
				eval, _ := strconv.ParseFloat(v.Format("20060102"), 64)
				return SQLFloat(eval), nil
			} else if v.Year() == 0 {
				eval, _ := strconv.ParseFloat(v.Format("150405"), 64)
				return SQLFloat(eval), nil
			}
			eval, _ := strconv.ParseFloat(v.Format("20060102150405"), 64)
			return SQLFloat(eval), nil
		case nil:
			return SQLNull, nil
		}

	case schema.SQLDate:
		var date time.Time
		switch v := value.(type) {
		case string:
			// SQLProxy only allows for converting strings to
			// time objects during pushdown or when performing
			// in-memory processing. Otherwise, string data
			// coming from MongoDB cannot be treated like a
			// time object.
			if mongoType != schema.MongoNone {
				break
			}
			for _, format := range schema.TimestampCtorFormats {
				d, err := time.Parse(format, v)
				if err == nil {
					date = time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, schema.DefaultLocale)
					return SQLDate{date}, nil
				}
			}
			val, _ := strconv.ParseFloat(v, 64)
			return SQLFloat(val), nil
		case time.Time:
			v = v.In(schema.DefaultLocale)
			date := time.Date(v.Year(), v.Month(), v.Day(), 0, 0, 0, 0, schema.DefaultLocale)
			return SQLDate{date}, nil
		default:
			return SQLInt(0), nil
		}

	case schema.SQLTimestamp:
		var ts time.Time
		switch v := value.(type) {
		case string:
			if mongoType != schema.MongoNone {
				break
			}
			for _, format := range schema.TimestampCtorFormats {
				d, err := time.Parse(format, v)
				if err == nil {
					ts = time.Date(d.Year(), d.Month(), d.Day(), d.Hour(),
						d.Minute(), d.Second(), d.Nanosecond(), schema.DefaultLocale)
					return SQLTimestamp{ts}, nil
				}
			}
			val, _ := strconv.ParseFloat(v, 64)
			return SQLFloat(val), nil
		case time.Time:
			return SQLTimestamp{v.In(schema.DefaultLocale)}, nil
		default:
			return SQLInt(0), nil
		}
	}

	return nil, fmt.Errorf("unable to convert '%v' (%T) to %v", value, value, sqlType)
}

//
// SQLBool represents a boolean.
//
type SQLBool bool

// SQLTrue is a constant SQLBool(true).
const SQLTrue = SQLBool(true)

// SQLFalse is a constant SQLBool(false).
const SQLFalse = SQLBool(false)

func (sb SQLBool) Decimal128() decimal.Decimal {
	if bool(sb) {
		return decimal.NewFromFloat(1)
	}
	return decimal.Zero
}

func (sb SQLBool) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return sb, nil
}

func (sb SQLBool) Float64() float64 {
	if bool(sb) {
		return float64(1)
	}
	return float64(0)
}

func (sb SQLBool) Int64() int64 {
	if bool(sb) {
		return int64(1)
	}
	return int64(0)
}

func (sb SQLBool) String() string {
	if sb {
		return "1"
	}
	return "0"
}

func (_ SQLBool) Type() schema.SQLType {
	return schema.SQLBoolean
}

func (sb SQLBool) Uint64() uint64 {
	if bool(sb) {
		return uint64(1)
	}
	return uint64(0)
}

func (sb SQLBool) Value() interface{} {
	return bool(sb)
}

//
// Time related SQL types and helpers.
//
func timeCmpHelper(at1, at2, at3, bt1, bt2, bt3 int) int {
	if at1 > bt1 {
		return 1
	} else if at1 == bt1 {
		if at2 > bt2 {
			return 1
		} else if at2 == bt2 {
			if at3 > bt3 {
				return 1
			} else if at3 < bt3 {
				return -1
			}
		} else if at2 < bt2 {
			return -1
		}
	} else if at1 < bt1 {
		return -1
	}
	return 0
}

//
// SQLDate represents a date.
//
type SQLDate struct {
	Time time.Time
}

func (sd SQLDate) Decimal128() decimal.Decimal {
	return decimal.NewFromFloat(sd.Float64())
}

func (sd SQLDate) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return sd, nil
}

func (sd SQLDate) Float64() float64 {
	val, _ := strconv.ParseFloat(sd.Time.Format("20060102"), 64)
	return val
}

func (sd SQLDate) Int64() int64 {
	val, _ := strconv.ParseInt(sd.Time.Format("20060102"), 10, 64)
	return val
}

func (sd SQLDate) String() string {
	return sd.Time.Format("2006-01-02")
}

func (_ SQLDate) Type() schema.SQLType {
	return schema.SQLDate
}

func (sd SQLDate) Uint64() uint64 {
	val, _ := strconv.ParseUint(sd.Time.Format("20060102"), 10, 64)
	return val
}

func (sd SQLDate) Value() interface{} {
	return sd.Time
}

//
// SQLTimestamp represents a timestamp value.
//
type SQLTimestamp struct {
	Time time.Time
}

func (st SQLTimestamp) Decimal128() decimal.Decimal {
	return decimal.NewFromFloat(st.Float64())
}

func (st SQLTimestamp) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return st, nil
}

func (st SQLTimestamp) Float64() float64 {
	if st.Time.Year() == 0 {
		val, _ := strconv.ParseFloat(st.Time.Format("150405"), 64)
		return val
	}
	val, _ := strconv.ParseFloat(st.Time.Format("20060102150405"), 64)
	return val
}

func (st SQLTimestamp) Int64() int64 {
	if st.Time.Year() == 0 {
		val, _ := strconv.ParseInt(st.Time.Format("150405"), 10, 64)
		return val
	}
	val, _ := strconv.ParseInt(st.Time.Format("20060102150405"), 10, 64)
	return val
}

func (st SQLTimestamp) String() string {
	return st.Time.Format("2006-01-02 15:04:05")
}

func (_ SQLTimestamp) Type() schema.SQLType {
	return schema.SQLTimestamp
}

func (st SQLTimestamp) Uint64() uint64 {
	if st.Time.Year() == 0 {
		val, _ := strconv.ParseUint(st.Time.Format("150405"), 10, 64)
		return val
	}
	val, _ := strconv.ParseUint(st.Time.Format("20060102150405"), 10, 64)
	return val
}

func (st SQLTimestamp) Value() interface{} {
	return st.Time
}

//
// SQLFloat represents a float.
//
type SQLFloat float64

func (sf SQLFloat) Decimal128() decimal.Decimal {
	return decimal.NewFromFloat(float64(sf))
}

func (sf SQLFloat) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return sf, nil
}

func (sf SQLFloat) Float64() float64 {
	return float64(sf)
}

func (sf SQLFloat) Int64() int64 {
	return int64(sf)
}

func (sf SQLFloat) String() string {
	return strconv.FormatFloat(float64(sf), 'f', -1, 64)
}

func (_ SQLFloat) Type() schema.SQLType {
	return schema.SQLFloat
}

func (sf SQLFloat) Uint64() uint64 {
	return uint64(sf)
}

func (sf SQLFloat) Value() interface{} {
	return float64(sf)
}

//
// SQLInt represents a 64-bit integer value.
//
type SQLInt int64

func (si SQLInt) Decimal128() decimal.Decimal {
	d, _ := decimal.NewFromString(si.String())
	return d
}

func (si SQLInt) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return si, nil
}

func (si SQLInt) Float64() float64 {
	return float64(si)
}

func (si SQLInt) Int64() int64 {
	return int64(si)
}

func (si SQLInt) String() string {
	return strconv.FormatInt(si.Int64(), 10)
}

func (_ SQLInt) Type() schema.SQLType {
	return schema.SQLInt
}

func (si SQLInt) Uint64() uint64 {
	return uint64(si)
}

func (si SQLInt) Value() interface{} {
	return int64(si)
}

//
// SQLNullValue represents a null.
//
type SQLNullValue struct{}

// SQLNull is a constant SQLNullValue.
var SQLNull = SQLNullValue{}

func (_ SQLNullValue) Decimal128() decimal.Decimal {
	return decimal.Zero
}

func (nv SQLNullValue) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return nv, nil
}

func (_ SQLNullValue) Float64() float64 {
	return float64(0)
}

func (_ SQLNullValue) Int64() int64 {
	return int64(0)
}

func (nv SQLNullValue) String() string {
	return schema.SQLNull
}

func (_ SQLNullValue) Type() schema.SQLType {
	return schema.SQLNull
}

func (_ SQLNullValue) Uint64() uint64 {
	return uint64(0)
}

func (_ SQLNullValue) Value() interface{} {
	return nil
}

//
// SQLNoValue represents no value.
//
type SQLNoValue struct{}

// SQLNone is a constant SQLNoValue.
var SQLNone = SQLNoValue{}

func (_ SQLNoValue) Decimal128() decimal.Decimal {
	return decimal.Zero
}

func (sn SQLNoValue) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return sn, nil
}

func (_ SQLNoValue) Float64() float64 {
	return float64(0)
}

func (_ SQLNoValue) Int64() int64 {
	return int64(0)
}

func (sn SQLNoValue) String() string {
	return schema.SQLNone
}

func (_ SQLNoValue) Type() schema.SQLType {
	return schema.SQLNone
}

func (_ SQLNoValue) Uint64() uint64 {
	return uint64(0)
}

func (_ SQLNoValue) Value() interface{} {
	return struct{}{}
}

//
// SQLDecimal128 represents a decimal 128 value.
//
type SQLDecimal128 decimal.Decimal

func (sd SQLDecimal128) Decimal128() decimal.Decimal {
	return decimal.Decimal(sd)
}

func (sd SQLDecimal128) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return sd, nil
}

func (sd SQLDecimal128) Float64() float64 {
	// second return value is f represents sd exactly
	f, _ := decimal.Decimal(sd).Float64()
	return f
}

func (sd SQLDecimal128) Int64() int64 {
	return decimal.Decimal(sd).Round(0).IntPart()
}

func (sd SQLDecimal128) String() string {
	return decimal.Decimal(sd).String()
}

func (_ SQLDecimal128) Type() schema.SQLType {
	return schema.SQLDecimal128
}

func (sd SQLDecimal128) Uint64() uint64 {
	return uint64(decimal.Decimal(sd).Round(0).IntPart())
}

func (sd SQLDecimal128) Value() interface{} {
	return decimal.Decimal(sd)
}

//
// SQLUUID represents a MongoDB UUID value.
//
type SQLUUID struct {
	kind  schema.MongoType
	bytes []byte
}

func (_ SQLUUID) Decimal128() decimal.Decimal {
	return decimal.Zero
}

func (uuid SQLUUID) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return uuid, nil
}

func (_ SQLUUID) Float64() float64 {
	return float64(0)
}

func (_ SQLUUID) Int64() int64 {
	return int64(0)
}

func (uuid SQLUUID) String() string {
	str := hex.EncodeToString(uuid.bytes)
	return str[0:8] +
		"-" + str[8:12] +
		"-" + str[12:16] +
		"-" + str[16:20] +
		"-" + str[20:]
}

func (_ SQLUUID) Type() schema.SQLType {
	return schema.SQLUUID
}

func (_ SQLUUID) Uint64() uint64 {
	return uint64(0)
}

func (uuid SQLUUID) Value() interface{} {
	return uuid.bytes
}

//
// SQLObjectID represents a MongoDB ObjectID value.
//
type SQLObjectID string

func (_ SQLObjectID) Decimal128() decimal.Decimal {
	return decimal.Zero
}

func (id SQLObjectID) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return id, nil
}

func (_ SQLObjectID) Float64() float64 {
	return float64(0)
}

func (_ SQLObjectID) Int64() int64 {
	return int64(0)
}

func (id SQLObjectID) String() string {
	return string(id)
}

func (id SQLObjectID) Type() schema.SQLType {
	return schema.SQLObjectID
}

func (_ SQLObjectID) Uint64() uint64 {
	return uint64(0)
}

func (id SQLObjectID) Value() interface{} {
	return bson.ObjectIdHex(string(id))
}

//
// SQLVarchar represents a string value.
//
type SQLVarchar string

func (sv SQLVarchar) Decimal128() decimal.Decimal {
	d, _ := decimal.NewFromString(sv.String())
	return d
}

func (sv SQLVarchar) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return sv, nil
}

func (sv SQLVarchar) Float64() float64 {
	val, _ := strconv.ParseFloat(string(sv), 64)
	return val
}

func (sv SQLVarchar) Int64() int64 {
	val, _ := strconv.ParseInt(string(sv), 10, 64)
	return val
}

func (sv SQLVarchar) String() string {
	return string(sv)
}

func (_ SQLVarchar) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (sv SQLVarchar) Uint64() uint64 {
	val, _ := strconv.ParseUint(string(sv), 10, 64)
	return val
}

func (sv SQLVarchar) Value() interface{} {
	return string(sv)
}

//
// SQLValues represents multiple sql values.
//
type SQLValues struct {
	Values []SQLValue
}

func (sv *SQLValues) Decimal128() decimal.Decimal {
	d, _ := decimal.NewFromString(sv.String())
	return d
}

func (sv *SQLValues) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return sv, nil
}

func (sv *SQLValues) Float64() float64 {
	return float64(sv.Values[0].Float64())
}

func (sv *SQLValues) Int64() int64 {
	return int64(sv.Values[0].Float64())
}

func (sv *SQLValues) normalize() node {
	if len(sv.Values) == 1 {
		return sv.Values[0]
	}

	return sv
}

func (sv *SQLValues) String() string {
	var values []string
	for _, n := range sv.Values {
		values = append(values, n.String())
	}
	return strings.Join(values, ", ")
}

func (v *SQLValues) Type() schema.SQLType {
	if len(v.Values) == 1 {
		return v.Values[0].Type()
	} else if len(v.Values) == 0 {
		return schema.SQLNone
	}

	return schema.SQLTuple
}

func (sv *SQLValues) Uint64() uint64 {
	return sv.Values[0].Uint64()
}

func (sv *SQLValues) Value() interface{} {
	values := []interface{}{}
	for _, v := range sv.Values {
		values = append(values, v.Value())
	}
	return values
}

//
// SQLUint32 represents an unsigned 32-bit integer.
//
type SQLUint32 uint32

func (su SQLUint32) Decimal128() decimal.Decimal {
	d, _ := decimal.NewFromString(su.String())
	return d
}

func (su SQLUint32) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return su, nil
}

func (su SQLUint32) Float64() float64 {
	return float64(su)
}

func (su SQLUint32) Int64() int64 {
	return int64(su)
}

func (su SQLUint32) String() string {
	return strconv.FormatInt(su.Int64(), 10)
}

func (su SQLUint32) Type() schema.SQLType {
	return schema.SQLInt
}

func (su SQLUint32) Uint64() uint64 {
	return uint64(su)
}

func (su SQLUint32) Value() interface{} {
	return uint32(su)
}

//
// SQLUint64 represents an unsigned 64-bit integer.
//
type SQLUint64 uint64

func (su SQLUint64) Decimal128() decimal.Decimal {
	d, _ := decimal.NewFromString(su.String())
	return d
}

func (su SQLUint64) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return su, nil
}

func (su SQLUint64) Float64() float64 {
	return float64(su)
}

func (su SQLUint64) Int64() int64 {
	return int64(su)
}

func (su SQLUint64) String() string {
	return strconv.FormatUint(uint64(su), 10)
}

func (su SQLUint64) Type() schema.SQLType {
	return schema.SQLUint64
}

func (su SQLUint64) Uint64() uint64 {
	return uint64(su)
}

func (su SQLUint64) Value() interface{} {
	return uint64(su)
}

// CompareTo compares two SQLValues. It returns -1 if
// left compares less than right; 1, if left compares
// greater than right; and 0 if left compares equal to
// right.
func CompareTo(left, right SQLValue) (int, error) {
	switch leftVal := left.(type) {
	case *SQLValues:
		err := fmt.Errorf("operand should contain %v columns", len(leftVal.Values))

		rightVal, ok := right.(*SQLValues)
		if !ok {
			// This allows for comparisons such as:
			// `select a, b from foo where (a) < 3`
			if len(leftVal.Values) != 1 {
				return -1, err
			}
			rightVal = &SQLValues{[]SQLValue{right}}
		} else if len(leftVal.Values) != len(rightVal.Values) {
			return -1, err
		}

		for i := 0; i < len(leftVal.Values); i++ {
			_, noLeft := leftVal.Values[i].(SQLNoValue)
			_, noRight := rightVal.Values[i].(SQLNoValue)

			if noLeft && !noRight {
				return -1, nil
			}

			if !noLeft && noRight {
				return 1, nil
			}

			if noLeft && noRight {
				return 0, nil
			}

			c, err := CompareTo(leftVal.Values[i], rightVal.Values[i])
			if err != nil {
				return c, err
			}

			if c != 0 {
				return c, nil
			}
		}
		return 0, nil
	default:
		switch right.(type) {
		case *SQLValues:
			i, err := CompareTo(right, left)
			if err != nil {
				return i, err
			}
			return -i, nil
		}
	}

	if left.Type() == right.Type() {
		switch leftVal := left.(type) {
		case SQLDate, SQLDecimal128, SQLFloat, SQLInt, SQLUint32, SQLUint64, SQLTimestamp:
			return compareDecimal128(left.Decimal128(), right.Decimal128())
		case SQLVarchar:
			rightVal, _ := right.(SQLVarchar)
			s1, s2 := string(leftVal), string(rightVal)
			if s1 < s2 {
				return -1, nil
			} else if s1 > s2 {
				return 1, nil
			}
			return 0, nil
		case SQLObjectID:
			rightVal, _ := right.(SQLObjectID)
			if !bson.IsObjectIdHex(leftVal.String()) {
				return -1, fmt.Errorf("%v is not a valid ObjectID", leftVal.String())
			}
			if !bson.IsObjectIdHex(rightVal.String()) {
				return -1, fmt.Errorf("%v is not a valid ObjectID", rightVal.String())
			}

			s1 := []byte(leftVal.String())
			s2 := []byte(rightVal.String())
			return compareBytes(s1, s2)
		case SQLNullValue:
			return 0, nil
		case SQLUUID:
			rightVal, ok := right.(SQLUUID)
			if !ok {
				return -1, fmt.Errorf("%v is not a valid UUID", right.String())
			}
			return compareBytes(leftVal.bytes, rightVal.bytes)
		}
	}

	// Different types
	switch lVal := left.(type) {
	case SQLNullValue:
		switch right.(type) {
		case *SQLValues:
			i, err := CompareTo(right, left)
			if err != nil {
				return i, err
			}
			return -i, nil
		default:
			return -1, nil
		}
	case SQLVarchar:
		switch right.(type) {
		case SQLDate, SQLTimestamp:
			// MySQL throws an error if you try to compare varchar =,<,> date/timestamp.
			// It works the other way around, however (i.e. date/timestamp =,<,> varchar).
			return -1, fmt.Errorf("Illegal mix of collations %T and %T", left, right)
		case SQLNullValue:
			return 1, nil
		default:
			return compareDecimal128(left.Decimal128(), right.Decimal128())
		}
	case SQLDate:
		switch rVal := right.(type) {
		case SQLVarchar:
			t, ok := parseDateTime(right.String())
			if !ok {
				t, _ = parseDateTime("0001-01-01")
			}
			return compareFloats(left.Float64(), SQLDate{Time: t}.Float64())
		case SQLTimestamp:
			if rVal.Time.Before(lVal.Time) {
				return 1, nil
			} else if rVal.Time.After(lVal.Time) {
				return -1, nil
			}
			return 0, nil
		case SQLNullValue:
			return 1, nil
		default:
			return compareDecimal128(left.Decimal128(), right.Decimal128())
		}
	case SQLTimestamp:
		switch rVal := right.(type) {
		case SQLVarchar:
			t, ok := parseDateTime(right.String())
			if !ok {
				t, _ = parseDateTime("0001-01-01 00:00:00")
			}
			return compareFloats(left.Float64(), SQLTimestamp{Time: t}.Float64())
		case SQLNullValue:
			return 1, nil
		case SQLDate:
			if rVal.Time.Before(lVal.Time) {
				return 1, nil
			} else if rVal.Time.After(lVal.Time) {
				return -1, nil
			}
			return 0, nil
		default:
			return compareDecimal128(left.Decimal128(), right.Decimal128())
		}
	case SQLUUID:
		switch right.(type) {
		case SQLVarchar:
			uuid, _ := getBinaryFromExpr(schema.MongoUUID, right)
			return compareBytes(lVal.bytes, uuid.Data)
		default:
			return compareDecimal128(left.Decimal128(), right.Decimal128())
		}

	default:
		switch right.(type) {
		case SQLNullValue:
			return 1, nil
		default:
			return compareDecimal128(left.Decimal128(), right.Decimal128())
		}
	}

	return -1, fmt.Errorf("comparing failed between %T and %T", left, right)
}
