package evaluator

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/schema"
	"github.com/shopspring/decimal"
)

//
// SQLBool represents a boolean.
//
type SQLBool float64

// SQLTrue is a constant SQLBool(1).
const SQLTrue = SQLBool(1)

// SQLFalse is a constant SQLBool(0).
const SQLFalse = SQLBool(0)

func NewSQLBool(b bool) SQLBool {
	if b {
		return SQLTrue
	}
	return SQLFalse
}

func (sb SQLBool) Decimal128() decimal.Decimal {
	return decimal.NewFromFloat(sb.Float64())
}

func (sb SQLBool) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return sb, nil
}

func (sb SQLBool) Float64() float64 {
	return float64(sb)
}

func (sb SQLBool) Int64() int64 {
	return int64(sb)
}

func (sb SQLBool) String() string {
	return strconv.FormatFloat(sb.Float64(), 'f', -1, 64)
}

func (_ SQLBool) Type() schema.SQLType {
	return schema.SQLBoolean
}

func (sb SQLBool) Uint64() uint64 {
	return uint64(sb)
}

func (sb SQLBool) Bool() bool {
	return sb > 0
}

func (sb SQLBool) Value() interface{} {
	return sb > 0
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
	ms := st.Time.Round(time.Microsecond)
	if ms.Equal(st.Time) {
		return st.Time.Format("2006-01-02 15:04:05")
	}
	return st.Time.Format("2006-01-02 15:04:05.000000")
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
	d, err := decimal.NewFromString(sv.String())
	if err != nil {
		return decimal.Zero
	}
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
func CompareTo(left, right SQLValue, collation *collation.Collation) (int, error) {
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

			c, err := CompareTo(leftVal.Values[i], rightVal.Values[i], collation)
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
			i, err := CompareTo(right, left, collation)
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
			return collation.CompareString(string(leftVal), string(rightVal)), nil
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
			i, err := CompareTo(right, left, collation)
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
