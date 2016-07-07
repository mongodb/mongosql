package evaluator

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/10gen/sqlproxy/schema"
	"github.com/mongodb/mongo-tools/common/util"
	"gopkg.in/mgo.v2/bson"
)

//
// NewSQLValue is a factory method for creating a SQLValue from a value and column type.
//
func NewSQLValue(value interface{}, sqlType schema.SQLType, mongoType schema.MongoType) (SQLValue, error) {
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
		case bool:
			return SQLBool(v), nil
		case string:
			return SQLVarchar(v), nil
		case float32, float64:
			eval, _ := util.ToFloat64(v)
			return SQLFloat(eval), nil
		case uint8, uint16, uint32, uint64, int, int8, int16, int32, int64:
			eval, _ := util.ToInt(v)
			return SQLInt(eval), nil
		case time.Time:
			h := v.Hour()
			mi := v.Minute()
			sd := v.Second()
			ns := v.Nanosecond()

			if ns == 0 && sd == 0 && mi == 0 && h == 0 {
				return NewSQLValue(v, schema.SQLDate, mongoType)
			}

			// SQLDateTime and SQLTimestamp can be handled similarly
			return NewSQLValue(v, schema.SQLTimestamp, mongoType)
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

	case schema.SQLVarchar:
		switch v := value.(type) {
		case bool:
			return SQLVarchar(strconv.FormatBool(v)), nil
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
		case time.Time:
			return SQLVarchar(v.String()), nil
		case nil:
			return SQLNull, nil
		default:
			return SQLVarchar(reflect.ValueOf(v).String()), nil
		}

	// This is going to be eliminated
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

	case schema.SQLFloat, schema.SQLNumeric, schema.SQLArrNumeric:
		switch v := value.(type) {
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
			// coming from MongoDB can not be treated like a
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

func (sb SQLBool) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return sb, nil
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

func (sb SQLBool) Value() interface{} {
	return bool(sb)
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

func (sd SQLDate) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return sd, nil
}

func (sd SQLDate) String() string {
	return sd.Time.Format("2006-01-02")
}

func (_ SQLDate) Type() schema.SQLType {
	return schema.SQLDate
}

func (sd SQLDate) Value() interface{} {
	return sd.Time
}

func (sd SQLDate) Float64() float64 {
	val, _ := strconv.ParseFloat(sd.Time.Format("20060102"), 64)
	return val
}

func (sd SQLDate) Int64() int64 {
	val, _ := strconv.ParseInt(sd.Time.Format("20060102"), 10, 64)
	return val
}

//
// SQLTimestamp represents a timestamp value.
//
type SQLTimestamp struct {
	Time time.Time
}

func (st SQLTimestamp) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return st, nil
}

func (st SQLTimestamp) String() string {
	return st.Time.Format("2006-01-02 15:04:05")
}

func (_ SQLTimestamp) Type() schema.SQLType {
	return schema.SQLTimestamp
}

func (st SQLTimestamp) Value() interface{} {
	return st.Time
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
	val, _ := strconv.ParseInt(st.Time.Format("20060102150405"), 10, 64)
	return val
}

//
// SQLFloat represents a float.
//
type SQLFloat float64

func (sf SQLFloat) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return sf, nil
}

func (sf SQLFloat) String() string {
	return strconv.FormatFloat(float64(sf), 'f', -1, 64)
}

func (_ SQLFloat) Type() schema.SQLType {
	return schema.SQLFloat
}

func (sf SQLFloat) Value() interface{} {
	return float64(sf)
}

func (sf SQLFloat) Float64() float64 {
	return float64(sf)
}

func (sf SQLFloat) Int64() int64 {
	return int64(sf)
}

//
// SQLInt represents a 64-bit integer value.
//
type SQLInt int64

func (si SQLInt) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return si, nil
}

func (si SQLInt) String() string {
	return strconv.FormatInt(si.Int64(), 10)
}

func (_ SQLInt) Type() schema.SQLType {
	return schema.SQLInt
}

func (si SQLInt) Value() interface{} {
	return int64(si)
}

func (si SQLInt) Float64() float64 {
	return float64(si)
}

func (si SQLInt) Int64() int64 {
	return int64(si)
}

//
// SQLNullValue represents a null.
//
type SQLNullValue struct{}

// SQLNull is a constant SQLNullValue.
var SQLNull = SQLNullValue{}

func (nv SQLNullValue) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return nv, nil
}

func (nv SQLNullValue) String() string {
	return schema.SQLNull
}

func (_ SQLNullValue) Type() schema.SQLType {
	return schema.SQLNull
}

func (_ SQLNullValue) Value() interface{} {
	return nil
}

func (_ SQLNullValue) Float64() float64 {
	return float64(0)
}

func (_ SQLNullValue) Int64() int64 {
	return int64(0)
}

//
// SQLNoValue represents no value.
//
type SQLNoValue struct{}

// SQLNone is a constant SQLNoValue.
var SQLNone = SQLNoValue{}

func (sn SQLNoValue) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return sn, nil
}

func (sn SQLNoValue) String() string {
	return schema.SQLNone
}

func (_ SQLNoValue) Type() schema.SQLType {
	return schema.SQLNone
}

func (_ SQLNoValue) Value() interface{} {
	return struct{}{}
}

func (_ SQLNoValue) Float64() float64 {
	return float64(0)
}

func (_ SQLNoValue) Int64() int64 {
	return int64(0)
}

//
// SQLObjectID represents a MongoDB ObjectID value.
//
type SQLObjectID string

func (id SQLObjectID) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return id, nil
}

func (id SQLObjectID) String() string {
	return string(id)
}

func (id SQLObjectID) Type() schema.SQLType {
	return schema.SQLObjectID
}

func (id SQLObjectID) Value() interface{} {
	return bson.ObjectIdHex(string(id))
}

func (_ SQLObjectID) Float64() float64 {
	return float64(0)
}

func (_ SQLObjectID) Int64() int64 {
	return int64(0)
}

//
// SQLVarchar represents a string value.
//
type SQLVarchar string

func (sv SQLVarchar) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return sv, nil
}

func (sv SQLVarchar) String() string {
	return string(sv)
}

func (_ SQLVarchar) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (sv SQLVarchar) Value() interface{} {
	return string(sv)
}

func (sv SQLVarchar) Float64() float64 {
	val, _ := strconv.ParseFloat(string(sv), 64)
	return val
}

func (sv SQLVarchar) Int64() int64 {
	val, _ := strconv.ParseInt(string(sv), 10, 64)
	return val
}

//
// SQLValues represents multiple sql values.
//
type SQLValues struct {
	Values []SQLValue
}

func (sv *SQLValues) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return sv, nil
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

func (sv *SQLValues) Value() interface{} {
	values := []interface{}{}
	for _, v := range sv.Values {
		values = append(values, v.Value())
	}
	return values
}

func (sv *SQLValues) Float64() float64 {
	return float64(sv.Values[0].Float64())
}

func (sv *SQLValues) Int64() int64 {
	return int64(sv.Values[0].Float64())
}

//
// SQLUint32 represents an unsigned 32-bit integer.
//
type SQLUint32 uint32

func (su SQLUint32) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return su, nil
}

func (su SQLUint32) String() string {
	return strconv.FormatInt(su.Int64(), 10)
}

func (su SQLUint32) Type() schema.SQLType {
	return schema.SQLInt
}

func (su SQLUint32) Value() interface{} {
	return uint32(su)
}

func (su SQLUint32) Float64() float64 {
	return float64(su)
}

func (su SQLUint32) Int64() int64 {
	return int64(su)
}

// round returns the closest integer value to the
// float - round half down for negative values and
// round half up otherwise.
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
		case SQLFloat, SQLInt, SQLUint32, SQLDate, SQLTimestamp:
			return compareFloats(left.Float64(), right.Float64())
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

			for i, _ := range s1 {
				if s1[i] < s2[i] {
					return -1, nil
				} else if s1[i] > s2[i] {
					return 1, nil
				}
			}
			return 0, nil
		case SQLNullValue:
			return 0, nil
		}
	}

	// Mix types
	switch left.(type) {
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
			return compareFloats(left.Float64(), right.Float64())
		}
	case SQLDate:
		switch right.(type) {
		case SQLVarchar:
			t, ok := parseDateTime(right.String())
			if !ok {
				t, _ = parseDateTime("0001-01-01")
			}
			return compareFloats(left.Float64(), SQLDate{Time: t}.Float64())
		case SQLNullValue:
			return 1, nil
		default:
			return compareFloats(left.Float64(), right.Float64())
		}
	case SQLTimestamp:
		switch right.(type) {
		case SQLVarchar:
			t, ok := parseDateTime(right.String())
			if !ok {
				t, _ = parseDateTime("0001-01-01 00:00:00")
			}
			return compareFloats(left.Float64(), SQLTimestamp{Time: t}.Float64())
		case SQLNullValue:
			return 1, nil
		default:
			return compareFloats(left.Float64(), right.Float64())
		}
	default:
		switch right.(type) {
		case SQLNullValue:
			return 1, nil
		default:
			return compareFloats(left.Float64(), right.Float64())
		}
	}

	return -1, fmt.Errorf("comparing failed between %T and %T", left, right)
}

func compareFloats(left, right float64) (int, error) {
	cmp := left - right
	if cmp < 0 {
		return -1, nil
	} else if cmp > 0 {
		return 1, nil
	}
	return 0, nil
}
