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
		case float32:
			return SQLFloat(float64(v)), nil
		case float64:
			return SQLFloat(float64(v)), nil
		case uint8:
			return SQLInt(int64(v)), nil
		case uint16:
			return SQLInt(int64(v)), nil
		case uint32:
			return SQLInt(int64(v)), nil
		case uint64:
			return SQLInt(int64(v)), nil
		case int:
			return SQLInt(int64(v)), nil
		case int8:
			return SQLInt(int64(v)), nil
		case int16:
			return SQLInt(int64(v)), nil
		case int32:
			return SQLInt(int64(v)), nil
		case int64:
			return SQLInt(v), nil
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
	case schema.SQLInt:
		switch v := value.(type) {
		case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64:
			eval, err := util.ToInt(v)
			if err == nil {
				return SQLInt(eval), nil
			}
		case nil:
			return SQLNull, nil
		}
	case schema.SQLInt64:
		switch v := value.(type) {
		case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64:
			eval, err := util.ToInt(v)
			if err == nil {
				return SQLInt(eval), nil
			}
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
		case int:
			return SQLVarchar(strconv.FormatInt(int64(v), 10)), nil
		case int8:
			return SQLVarchar(strconv.FormatInt(int64(v), 10)), nil
		case int16:
			return SQLVarchar(strconv.FormatInt(int64(v), 10)), nil
		case int32:
			return SQLVarchar(strconv.FormatInt(int64(v), 10)), nil
		case int64:
			return SQLVarchar(strconv.FormatInt(int64(v), 10)), nil
		case uint8:
			return SQLVarchar(strconv.FormatInt(int64(v), 10)), nil
		case uint16:
			return SQLVarchar(strconv.FormatInt(int64(v), 10)), nil
		case uint32:
			return SQLVarchar(strconv.FormatInt(int64(v), 10)), nil
		case uint64:
			return SQLVarchar(strconv.FormatInt(int64(v), 10)), nil
		case bson.ObjectId:
			return SQLObjectID(v.Hex()), nil
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
		case time.Time:
			v = v.In(schema.DefaultLocale)
			date := time.Date(v.Year(), v.Month(), v.Day(), 0, 0, 0, 0, schema.DefaultLocale)
			return SQLDate{date}, nil
		case nil:
			return SQLNull, nil
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
		case time.Time:
			return SQLTimestamp{v.In(schema.DefaultLocale)}, nil
		case nil:
			return SQLNull, nil
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
		return "true"
	}
	return "false"
}

func (_ SQLBool) Type() schema.SQLType {
	return schema.SQLBoolean
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

func (sf SQLFloat) Value() interface{} {
	return float64(sf)
}

//
// SQLInt represents a 64-bit integer value.
//
type SQLInt int64

func (si SQLInt) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return si, nil
}

func (si SQLInt) String() string {
	return strconv.Itoa(int(si))
}

func (_ SQLInt) Type() schema.SQLType {
	return schema.SQLInt
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

func (si SQLInt) Value() interface{} {
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
	var exprs []SQLExpr

	for _, expr := range v.Values {
		exprs = append(exprs, expr.(SQLExpr))
	}

	return preferentialType(exprs...)
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

func (su SQLUint32) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return su, nil
}

func (su SQLUint32) String() string {
	return strconv.FormatFloat(float64(su), 'f', -1, 32)
}

func (su SQLUint32) Type() schema.SQLType {
	return schema.SQLInt
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

func (su SQLUint32) Value() interface{} {
	return uint32(su)
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
	case SQLBool:

		switch rightVal := right.(type) {
		case SQLBool:
			s1, s2 := bool(leftVal), bool(rightVal)
			if s1 == s2 {
				return 0, nil
			}
			if s1 && !s2 {
				return 1, nil
			}
			return -1, nil
		case SQLDate, SQLTimestamp:
			return -1, nil
		case SQLFloat, SQLNullValue, SQLInt, SQLObjectID, SQLVarchar, SQLUint32:
			return 1, nil
		case *SQLValues:
			i, err := CompareTo(right, left)
			if err != nil {
				return i, err
			}
			return -i, nil
		}

	case SQLDate:

		var t1, t2 time.Time

		switch rightVal := right.(type) {
		case SQLBool, SQLFloat, SQLInt, SQLNullValue, SQLObjectID, SQLUint32:
			return 1, nil
		case SQLDate:
			t1 = leftVal.Time
			t2 = rightVal.Time
		case SQLVarchar:
			t1 = leftVal.Time
			v, err := NewSQLValue(rightVal.String(), schema.SQLDate, schema.MongoDate)
			if err != nil {
				return 1, nil
			}
			t2 = v.(SQLDate).Time
		case SQLTimestamp:
			t1 = leftVal.Time
			t2 = rightVal.Time
		case *SQLValues:
			i, err := CompareTo(right, left)
			if err != nil {
				return i, err
			}
			return -i, nil
		default:
			return -1, fmt.Errorf("cannot compare SQLDate against %T", rightVal)
		}

		if t1.After(t2) {
			return 1, nil
		} else if t1.Before(t2) {
			return -1, nil
		}

		return 0, nil

	case SQLFloat, SQLInt, SQLUint32:

		switch rightVal := right.(type) {
		case SQLNullValue:
			return 1, nil
		case SQLFloat, SQLInt, SQLUint32:
			cmp := leftVal.(SQLNumeric).Float64() - rightVal.(SQLNumeric).Float64()
			if cmp > 0 {
				return 1, nil
			} else if cmp < 0 {
				return -1, nil
			}
			return 0, nil
		case *SQLValues:
			i, err := CompareTo(right, left)
			if err != nil {
				return i, err
			}
			return -i, nil
		case SQLBool, SQLDate, SQLObjectID, SQLVarchar, SQLTimestamp:
			return -1, nil
		default:
			return -1, fmt.Errorf("cannot compare %T against %T", left, right)
		}

	case SQLNullValue:

		switch right.(type) {
		case SQLNullValue:
			return 0, nil
		case *SQLValues:
			i, err := CompareTo(right, left)
			if err != nil {
				return i, err
			}
			return -i, nil
		default:
			return -1, nil
		}

	case SQLObjectID:

		switch rightVal := right.(type) {
		case SQLBool, SQLDate, SQLTimestamp:
			return -1, nil
		case SQLFloat, SQLNullValue, SQLInt, SQLVarchar, SQLUint32:
			return 1, nil
		case SQLObjectID:
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
		case *SQLValues:
			i, err := CompareTo(right, left)
			if err != nil {
				return i, err
			}
			return -i, nil
		default:
			return -1, fmt.Errorf("cannot compare SQLVarchar against %T", right)
		}

	case SQLVarchar:

		switch rightVal := right.(type) {
		case SQLBool, SQLDate, SQLObjectID, SQLTimestamp:
			return -1, nil
		case SQLFloat, SQLInt, SQLNullValue, SQLUint32:
			return 1, nil
		case SQLVarchar:
			s1, s2 := string(leftVal), string(rightVal)
			if s1 < s2 {
				return -1, nil
			} else if s1 > s2 {
				return 1, nil
			}
			return 0, nil
		case *SQLValues:
			i, err := CompareTo(right, left)
			if err != nil {
				return i, err
			}
			return -i, nil
		default:
			return -1, fmt.Errorf("cannot compare SQLVarchar against %T", right)
		}

	case SQLTimestamp:

		var t1, t2 time.Time

		switch rightVal := right.(type) {
		case SQLBool, SQLFloat, SQLInt, SQLNullValue, SQLObjectID, SQLUint32:
			return 1, nil
		case SQLDate:
			t1 = leftVal.Time
			v, err := NewSQLValue(rightVal.Time, schema.SQLDate, schema.MongoDate)
			if err != nil {
				return 1, nil
			}
			t2 = v.(SQLDate).Time
		case SQLVarchar:
			t1 = leftVal.Time
			v, err := NewSQLValue(rightVal.String(), schema.SQLTimestamp, schema.MongoDate)
			if err != nil {
				return 1, nil
			}
			t2 = v.(SQLTimestamp).Time
		case SQLTimestamp:
			t1 = leftVal.Time
			t2 = rightVal.Time
		case *SQLValues:
			i, err := CompareTo(right, left)
			if err != nil {
				return i, err
			}
			return -i, nil
		default:
			return -1, fmt.Errorf("cannot compare SQLTimestamp against %T", right)
		}

		if t1.After(t2) {
			return 1, nil
		} else if t1.Before(t2) {
			return -1, nil
		}
		return 0, nil

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
	}
	return -1, fmt.Errorf("unknown SQLValue: %T", left)
}
