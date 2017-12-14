package evaluator

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
	"github.com/shopspring/decimal"
)

const (
	DateTimeFormat = "2006-01-02 15:04:05.000000"
)

// IsSimilar returns true if the logical or comparison
// operations can be carried on leftType against rightType
// with no need for type conversion.
func isSimilar(leftType, rightType schema.SQLType) bool {
	if leftType == rightType {
		return true
	}
	if leftType == schema.SQLNull || rightType == schema.SQLNull {
		return true
	}
	if leftType == schema.SQLNone || rightType == schema.SQLNone {
		return true
	}
	switch leftType {
	case schema.SQLArrNumeric, schema.SQLFloat, schema.SQLInt, schema.SQLInt64, schema.SQLNumeric, schema.SQLUint64, schema.SQLDecimal128:
		switch rightType {
		case schema.SQLArrNumeric, schema.SQLFloat, schema.SQLInt, schema.SQLInt64, schema.SQLNumeric, schema.SQLUint64, schema.SQLDecimal128:
			return true
		}
	case schema.SQLDate, schema.SQLTimestamp:
		switch rightType {
		case schema.SQLDate, schema.SQLTimestamp:
			return true
		}
	}
	return false
}

// NewSQLValue is a factory method for creating a SQLValue from
// an in-memory value and column type.
// It returns the converted SQLValue and a bool, which is true if the value
// returned is a default value, or false if it was successfully converted from
// the provided value.
// For a full description of its behavior, please see http://bit.ly/2bhWuyw
func NewSQLValue(value interface{}, sqlType, fromType schema.SQLType) (SQLValue, bool) {

	if value == nil {
		return SQLNull, true
	}

	if sqlType != schema.SQLNone {
		if v, ok := value.(SQLValue); ok {
			value = v.Value()
		}
	}

	switch sqlType {

	case schema.SQLBoolean:
		switch v := value.(type) {
		case bool:
			return NewSQLBool(v), false
		case bson.ObjectId:
			return SQLTrue, true
		case decimal.Decimal:
			flt, _ := v.Float64()
			return SQLBool(flt), false
		case float32, float64, int, int8, int16, int32, int64, uint8, uint16, uint32, uint64:
			flt, _ := util.ToFloat64(v)
			return SQLBool(flt), false
		case string:
			flt, _ := strconv.ParseFloat(v, 64)
			return SQLBool(flt), false
		case time.Time:
			return SQLTrue, true
		default:
			return SQLFalse, true
		}

	case schema.SQLDate:
		switch v := value.(type) {
		case bson.ObjectId:
			return SQLDate{v.Time()}, false
		case string:
			date, _, ok := parseDateTime(v)
			if ok {
				date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
				return SQLDate{date}, false
			}
			return SQLDate{time.Time{}}, true
		case time.Time:
			v = v.In(schema.DefaultLocale)
			date := time.Date(v.Year(), v.Month(), v.Day(), 0, 0, 0, 0, schema.DefaultLocale)
			return SQLDate{date}, false
		default:
			return SQLDate{time.Time{}}, true
		}

	case schema.SQLDecimal128:
		switch v := value.(type) {
		case bool:
			if v {
				return SQLDecimal128(decimal.NewFromFloat(1)), false
			}
			return SQLDecimal128(decimal.NewFromFloat(0)), false
		case bson.ObjectId:
			dec, _ := decimal.NewFromString(v.String())
			return SQLDecimal128(dec), false
		case decimal.Decimal:
			return SQLDecimal128(v), false
		case float32:
			return SQLDecimal128(decimal.NewFromFloat(float64(v))), false
		case float64:
			return SQLDecimal128(decimal.NewFromFloat(v)), false
		case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64:
			flt, _ := util.ToFloat64(v)
			return SQLDecimal128(decimal.NewFromFloat(flt)), false
		case string:
			dec, err := decimal.NewFromString(v)
			if err == nil {
				return SQLDecimal128(dec), false
			}
			fltVal, isDefault := NewSQLValue(value, schema.SQLFloat, fromType)
			flt, _ := fltVal.(SQLFloat)
			return SQLDecimal128(decimal.NewFromFloat(float64(flt))), isDefault
		case time.Time:
			h, m, s := v.Clock()

			if h+m+s == 0 {
				flt, _ := strconv.ParseFloat(v.Format("20060102"), 64)
				return SQLDecimal128(decimal.NewFromFloat(flt)), false
			} else if v.Year() == 0 {
				flt, _ := strconv.ParseFloat(v.Format("150405.000000"), 64)
				return SQLDecimal128(decimal.NewFromFloat(flt)), false
			}
			flt, _ := strconv.ParseFloat(v.Format("20060102150405.000000"), 64)
			return SQLDecimal128(decimal.NewFromFloat(flt)), false
		default:
			return SQLDecimal128(decimal.Zero), true
		}

	case schema.SQLFloat, schema.SQLNumeric, schema.SQLArrNumeric:
		switch v := value.(type) {
		case bool:
			if v {
				return SQLFloat(1), false
			}
			return SQLFloat(0), false
		case bson.ObjectId:
			return NewSQLValue(v.Time(), schema.SQLFloat, fromType)
		case bson.Decimal128:
			flt, _ := strconv.ParseFloat(v.String(), 64)
			return SQLFloat(flt), false
		case decimal.Decimal:
			flt, _ := v.Float64()
			return SQLFloat(flt), false
		case float32:
			return SQLFloat(float64(v)), false
		case float64:
			return SQLFloat(v), false
		case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64:
			flt, _ := util.ToFloat64(v)
			return SQLFloat(flt), false
		case string:
			flt, _ := strconv.ParseFloat(v, 64)
			return SQLFloat(flt), false
		case time.Time:
			h, m, s := v.Clock()

			if h+m+s == 0 {
				flt, _ := strconv.ParseFloat(v.Format("20060102"), 64)
				return SQLFloat(flt), false
			} else if v.Year() == 0 {
				flt, _ := strconv.ParseFloat(v.Format("150405.000000"), 64)
				return SQLFloat(flt), false
			}
			flt, _ := strconv.ParseFloat(v.Format("20060102150405.000000"), 64)
			return SQLFloat(flt), false
		default:
			return SQLFloat(0.0), true
		}

	case schema.SQLInt:
		switch v := value.(type) {
		case bool:
			if v {
				return SQLInt(1), false
			}
			return SQLInt(0), false
		case bson.ObjectId:
			return NewSQLValue(v.Time(), schema.SQLInt, fromType)
		case bson.Decimal128:
			flt, _ := strconv.ParseFloat(v.String(), 64)
			return SQLInt(round(float64(flt))), false
		case decimal.Decimal:
			i := v.Round(0).IntPart()
			return SQLInt(i), false
		case float32:
			return SQLInt(round(float64(v))), false
		case float64:
			return SQLInt(round(v)), false
		case int:
			return SQLInt(v), false
		case int32:
			return SQLInt(int(v)), false
		case int64:
			return SQLInt(int(v)), false
		case string:
			v = CleanNumericString(v)
			// strings are truncated instead of rounded.
			fStrParts := strings.Split(v, ".")
			i, err := strconv.Atoi(fStrParts[0])
			if err != nil {
				return SQLInt(0), false
			}
			return SQLInt(i), false
		case time.Time:
			h, m, s := v.Clock()

			if h+m+s == 0 {
				flt, _ := strconv.ParseFloat(v.Format("20060102"), 64)
				return SQLInt(int(flt)), false
			} else if v.Year() == 0 {
				flt, _ := strconv.ParseFloat(v.Format("150405.000000"), 64)
				return SQLInt(int(flt)), false
			}
			flt, _ := strconv.ParseFloat(v.Format("20060102150405.000000"), 64)
			return SQLInt(int(flt)), false
		case uint32:
			return SQLInt(int(v)), false
		case uint64:
			return SQLInt(int(v)), false
		default:
			return SQLInt(0), true
		}

	case schema.SQLInt64:
		flt, isDefault := NewSQLValue(value, schema.SQLFloat, fromType)
		return SQLInt(flt.Int64()), isDefault

	case schema.SQLUint64:
		flt, isDefault := NewSQLValue(value, schema.SQLFloat, fromType)
		return SQLUint64(flt.Uint64()), isDefault

	case schema.SQLObjectID:
		switch v := value.(type) {
		case bool:
			if v {
				return SQLObjectID("1"), false
			}
			return SQLObjectID("0"), false
		case bson.ObjectId:
			return SQLObjectID(v.Hex()), false
		case decimal.Decimal:
			return SQLObjectID(util.FormatDecimal(v)), false
		case float32, float64:
			flt, _ := util.ToFloat64(v)
			return SQLObjectID(strconv.FormatFloat(flt, 'f', -1, 64)), false
		case int, int8, int16, int32, uint8, uint16, uint32:
			val, _ := util.ToInt(v)
			return SQLObjectID(strconv.FormatInt(int64(val), 10)), false
		case int64:
			return SQLObjectID(strconv.FormatInt(int64(v), 10)), false
		case string:
			return SQLObjectID(v), false
		case uint64:
			return SQLObjectID(strconv.FormatUint(v, 10)), false
		case time.Time:
			return SQLObjectID(bson.NewObjectIdWithTime(v).Hex()), false
		default:
			return SQLObjectID(""), true
		}
	case schema.SQLTimestamp:
		switch v := value.(type) {
		case bson.ObjectId:
			return SQLTimestamp{v.Time()}, false
		case string:
			date, _, ok := parseDateTime(v)
			if ok {
				return SQLTimestamp{date}, false
			}
			return SQLTimestamp{time.Time{}}, true
		case time.Time:
			return SQLTimestamp{v.In(schema.DefaultLocale)}, false
		default:
			return SQLTimestamp{time.Time{}}, true
		}

	case schema.SQLUUID:
		v, err := NewSQLValueFromUUID(value, sqlType, schema.MongoUUID)
		if err != nil {
			return v, true
		}
		return v, false

	case schema.SQLVarchar:
		switch v := value.(type) {
		case bool:
			if v {
				return SQLVarchar("1"), false
			}
			return SQLVarchar("0"), false
		case bson.ObjectId:
			return SQLObjectID(v.Hex()), false
		case decimal.Decimal:
			return SQLVarchar(util.FormatDecimal(v)), false
		case float32:
			return SQLVarchar(strconv.FormatFloat(float64(v), 'f', -1, 32)), false
		case float64:
			return SQLVarchar(strconv.FormatFloat(v, 'f', -1, 64)), false
		case uint8, uint16, uint32, uint64, int, int8, int16, int32, int64:
			val, _ := util.ToInt(v)
			return SQLVarchar(strconv.FormatInt(int64(val), 10)), false
		case string:
			return SQLVarchar(v), false
		case time.Time:
			if fromType == schema.SQLDate {
				return SQLVarchar(v.Format("2006-01-02")), false
			}
			asString := v.Format(DateTimeFormat)
			return SQLVarchar(asString), false
		default:
			return SQLVarchar(reflect.ValueOf(v).String()), true
		}

	case schema.SQLNone:
		if v, ok := value.(SQLValue); ok {
			return v, true
		}

	}

	panic(fmt.Errorf("can't convert value of go type '%T' to a SQLValue of SQLType '%v'", value, sqlType))
}

func NewSQLValueWithDefault(value interface{}, sqlType, fromType schema.SQLType, defaultValue SQLValue) SQLValue {
	val, isDefault := NewSQLValue(value, sqlType, fromType)
	if isDefault {
		return defaultValue
	}
	return val
}

// NewSQLValueFromUUID is a factory method for creating a SQLUUID
// from a given value
func NewSQLValueFromUUID(value interface{}, sqlType schema.SQLType, mongoType schema.MongoType) (SQLValue, error) {

	err := fmt.Errorf("unable to convert '%v' (%T) to %v", value, value, sqlType)

	if !IsUUID(mongoType) {
		return nil, err
	}

	if sqlType != schema.SQLVarchar {
		return nil, err
	}

	switch v := value.(type) {
	case bson.Binary:
		if err := NormalizeUUID(mongoType, v.Data); err != nil {
			return nil, err
		}

		return SQLUUID{mongoType, v.Data}, nil
	case string:
		b, ok := GetBinaryFromExpr(mongoType, SQLVarchar(v))
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
		case decimal.Decimal:
			return SQLDecimal128(v), nil
		case bool:
			return NewSQLBool(v), nil
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
			return SQLInt(d.IntPart()), err
		case decimal.Decimal:
			return SQLInt(v.IntPart()), nil
		case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64:
			eval, err := util.ToInt(v)
			if err == nil {
				return SQLInt(eval), nil
			}
		case string:
			cleaned := CleanNumericString(v)
			eval, err := strconv.ParseInt(strings.Split(cleaned, ".")[0], 10, 64)
			if err == nil {
				return SQLInt(eval), nil
			}
			return SQLInt(0), nil
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
			return SQLUint64(d.IntPart()), err
		case decimal.Decimal:
			return SQLUint64(v.IntPart()), nil
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
		case string:
			cleaned := CleanNumericString(v)
			eval, err := strconv.ParseInt(strings.Split(cleaned, ".")[0], 10, 64)
			if err == nil {
				return SQLUint64(eval), nil
			}
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
		if IsUUID(mongoType) {
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
			return SQLVarchar(v.String()), nil
		case decimal.Decimal:
			return SQLVarchar(util.FormatDecimal(v)), nil
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
			cleaned := CleanNumericString(v)
			f, err := strconv.ParseFloat(cleaned, 64)
			if err != nil {
				return SQLFalse, nil
			}
			return SQLBool(f), nil
		case uint8, uint16, uint32, uint64, int, int8, int16, int32, int64, float32, float64:
			eval, _ := util.ToFloat64(v)
			return SQLBool(eval), nil
		case time.Time:
			var eval int64
			h, m, s := v.Clock()
			// Date, time or timestamp.
			if h+m+s == 0 {
				val, _ := strconv.ParseInt(v.Format("20060102"), 10, 64)
				// TODO: should we handle the error here?
				eval = val
			} else if v.Year() == 0 {
				val, _ := strconv.ParseInt(v.Format("150405"), 10, 64)
				// TODO: should we handle the error here?
				eval = val
			} else {
				val, _ := strconv.ParseInt(v.Format("20060102150405"), 10, 64)
				// TODO: should we handle the error here?
				eval = val
			}
			return SQLBool(eval), nil
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
		case decimal.Decimal:
			return SQLDecimal128(v), nil
		case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64:
			eval, err := util.ToFloat64(v)
			if err == nil {
				return SQLDecimal128(decimal.NewFromFloat(eval)), nil
			}
		case string:
			cleaned := CleanNumericString(v)
			d, err := decimal.NewFromString(cleaned)
			if err == nil {
				return SQLDecimal128(d), err
			}
		case time.Time:
			h, m, s := v.Clock()
			// Date, otherwise timestamp
			if h+m+s == 0 {
				eval, _ := strconv.ParseFloat(v.Format("20060102"), 64)
				return SQLDecimal128(decimal.NewFromFloat(eval)), nil
			} else if v.Year() == 0 {
				eval, _ := strconv.ParseFloat(v.Format("150405.000000"), 64)
				return SQLDecimal128(decimal.NewFromFloat(eval)), nil
			}
			eval, _ := strconv.ParseFloat(v.Format("20060102150405.000000"), 64)
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
		case decimal.Decimal:
			flt, _ := v.Float64()
			return SQLFloat(flt), nil
		case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64:
			eval, err := util.ToFloat64(v)
			if err == nil {
				return SQLFloat(eval), nil
			}
		case string:
			cleaned := CleanNumericString(v)
			eval, err := strconv.ParseFloat(cleaned, 64)
			if err == nil {
				return SQLFloat(eval), nil
			}
			return SQLFloat(0.0), nil
		case time.Time:
			h, m, s := v.Clock()
			// Date, otherwise timestamp
			if h+m+s == 0 {
				eval, _ := strconv.ParseFloat(v.Format("20060102"), 64)
				return SQLFloat(eval), nil
			} else if v.Year() == 0 {
				eval, _ := strconv.ParseFloat(v.Format("150405.000000"), 64)
				return SQLFloat(eval), nil
			}
			eval, _ := strconv.ParseFloat(v.Format("20060102150405.000000"), 64)
			return SQLFloat(eval), nil
		case nil:
			return SQLNull, nil
		}

	case schema.SQLDate:
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
			if v == "0000-00-00" {
				return SQLVarchar("0000-00-00"), nil
			}
			date, _, ok := parseDateTime(v)
			if ok {
				date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
				return SQLDate{date}, nil
			}
			val, _ := strconv.ParseFloat(v, 64)
			return SQLFloat(val), nil
		case time.Time:
			v = v.In(schema.DefaultLocale)
			date := time.Date(v.Year(), v.Month(), v.Day(), 0, 0, 0, 0, schema.DefaultLocale)
			return SQLDate{date}, nil
		default:
			return SQLVarchar("0000-00-00"), nil
		}

	case schema.SQLTimestamp:
		switch v := value.(type) {
		case string:
			if mongoType != schema.MongoNone {
				break
			}
			date, _, ok := parseDateTime(v)
			if ok {
				return SQLTimestamp{date}, nil
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

func convertExprs(exprs []SQLExpr, convTypes []schema.SQLType, defaults []SQLValue) []SQLExpr {
	if len(convTypes) < len(exprs) {
		// There is an error in how this function is being used
		panic("convTypes shorter than exprs")
	} else if len(convTypes) != len(defaults) {
		// There is an error in how this function is being used
		panic("convTypes not same length as defaults")
	}
	newExprs := make([]SQLExpr, len(exprs))
	for i, expr := range exprs {
		convType := convTypes[i]
		defaultValue := defaults[i]
		exprType := expr.Type()
		if isSimilar(exprType, convType) {
			newExprs[i] = expr
		} else {
			newExprs[i] = &SQLConvertExpr{
				expr,
				convType,
				defaultValue,
			}
		}
	}
	return newExprs
}

func convertAllExprs(exprs []SQLExpr, convType schema.SQLType, defaultValue SQLValue) []SQLExpr {
	convTypes := make([]schema.SQLType, len(exprs))
	defaults := make([]SQLValue, len(exprs))
	for i := range exprs {
		convTypes[i] = convType
		defaults[i] = defaultValue
	}
	return convertExprs(exprs, convTypes, defaults)
}

// preferentialType accepts a variable number of
// SQLExprs and returns the type of the SQLExpr
// with the highest preference.
func preferentialType(exprs ...SQLExpr) schema.SQLType {
	s := &schema.SQLTypesSorter{}
	return preferentialTypeWithSorter(s, exprs...)
}

func preferentialTypeWithSorter(s *schema.SQLTypesSorter, exprs ...SQLExpr) schema.SQLType {
	if len(exprs) == 0 {
		return schema.SQLNone
	}

	for _, expr := range exprs {
		s.Types = append(s.Types, expr.Type())
	}

	sort.Sort(s)

	return s.Types[len(s.Types)-1]
}

// ReconcileSQLExprs takes two SQLExpr and ensures that
// they are of the same type. If they are of different
// types but still comparable, it wraps the SQLExpr with
// a lesser precedence in a SQLConvertExpr. If they are
// not comparable, it returns a non-nil error.
func ReconcileSQLExprs(left, right SQLExpr) (SQLExpr, SQLExpr, error) {

	leftType, rightType := left.Type(), right.Type()

	if leftType == schema.SQLTuple || rightType == schema.SQLTuple {
		return reconcileSQLTuple(left, right)
	}

	if leftType == rightType || isSimilar(leftType, rightType) {
		return left, right, nil
	}

	sorter := &schema.SQLTypesSorter{
		Types: []schema.SQLType{leftType, rightType},
	}

	sort.Sort(sorter)

	if sorter.Types[0] == schema.SQLObjectID {
		sorter.Types[0], sorter.Types[1] = sorter.Types[1], sorter.Types[0]
	}

	if sorter.Types[1] == leftType {
		right = &SQLConvertExpr{right, sorter.Types[1], SQLNone}
	} else {
		left = &SQLConvertExpr{left, sorter.Types[1], SQLNone}
	}

	return left, right, nil
}

func reconcileSQLTuple(left, right SQLExpr) (SQLExpr, SQLExpr, error) {

	getSQLExprs := func(expr SQLExpr) ([]SQLExpr, error) {
		switch typedE := expr.(type) {
		case *SQLTupleExpr:
			return typedE.Exprs, nil
		case *SQLSubqueryExpr:
			return typedE.Exprs(), nil
		}
		return nil, fmt.Errorf("cannot reconcile non-tuple type '%T'", expr)
	}

	wrapReconciledExprs := func(expr SQLExpr, newExprs []SQLExpr) (SQLExpr, error) {
		switch typedE := expr.(type) {
		case *SQLTupleExpr:
			return &SQLTupleExpr{newExprs}, nil
		case *SQLSubqueryExpr:
			plan := typedE.plan

			var projectedColumns ProjectedColumns
			for i, c := range plan.Columns() {
				projectedColumns = append(projectedColumns, ProjectedColumn{
					Column: c,
					Expr:   newExprs[i],
				})
			}

			return &SQLSubqueryExpr{
				correlated: typedE.correlated,
				plan:       NewProjectStage(plan, projectedColumns...),
				allowRows:  typedE.allowRows,
			}, nil
		}
		return nil, fmt.Errorf("cannot wrap reconciled non-tuple type '%T'", expr)
	}

	var leftExprs []SQLExpr
	var rightExprs []SQLExpr
	var err error

	if left.Type() == schema.SQLTuple {
		leftExprs, err = getSQLExprs(left)
		if err != nil {
			return nil, nil, err
		}
	}

	if right.Type() == schema.SQLTuple {
		rightExprs, err = getSQLExprs(right)
		if err != nil {
			return nil, nil, err
		}
	}

	var newLeftExprs []SQLExpr
	var newRightExprs []SQLExpr

	// cases here:
	// (a, b) = (1, 2)
	// (a) = (1)
	// (a) in (1, 2)
	// (a) = (SELECT a FROM foo)
	if left.Type() == schema.SQLTuple && right.Type() == schema.SQLTuple {

		numLeft, numRight := len(leftExprs), len(rightExprs)

		if numLeft != numRight && numLeft != 1 {
			return nil, nil, mysqlerrors.Defaultf(mysqlerrors.ER_OPERAND_COLUMNS, numLeft)
		}

		hasNewLeft := false
		hasNewRight := false

		for i := range rightExprs {
			leftExpr := leftExprs[0]
			if numLeft != 1 {
				leftExpr = leftExprs[i]
			}

			rightExpr := rightExprs[i]

			newLeftExpr, newRightExpr, err := ReconcileSQLExprs(leftExpr, rightExpr)
			if err != nil {
				return nil, nil, err

			}

			if leftExpr != newLeftExpr {
				hasNewLeft = true
			}

			if rightExpr != newRightExpr {
				hasNewRight = true
			}

			newLeftExprs = append(newLeftExprs, newLeftExpr)
			newRightExprs = append(newRightExprs, newRightExpr)
		}

		if numLeft == 1 {
			newLeftExprs = newLeftExprs[:1]
		}

		if hasNewLeft {
			left, err = wrapReconciledExprs(left, newLeftExprs)
			if err != nil {
				return nil, nil, err
			}
		}

		if hasNewRight {
			right, err = wrapReconciledExprs(right, newRightExprs)
			if err != nil {
				return nil, nil, err
			}
		}

		return left, right, nil
	}

	// cases here:
	// (a) = 1
	// (SELECT a FROM foo) = 1
	if left.Type() == schema.SQLTuple && right.Type() != schema.SQLTuple {

		if len(leftExprs) != 1 {
			return nil, nil, mysqlerrors.Defaultf(mysqlerrors.ER_OPERAND_COLUMNS, len(leftExprs))
		}

		var newLeftExpr SQLExpr

		newLeftExpr, _, err = ReconcileSQLExprs(leftExprs[0], right)
		if err != nil {
			return nil, nil, err
		}

		if left != newLeftExpr {
			newLeftExprs = append(newLeftExprs, newLeftExpr)
			left, err = wrapReconciledExprs(left, newLeftExprs)
			if err != nil {
				return nil, nil, err
			}
		}

		return left, right, nil
	}

	// cases here:
	// a = (1)
	// a = (SELECT a FROM foo)
	// a in (1, 2)
	if left.Type() != schema.SQLTuple && right.Type() == schema.SQLTuple {

		hasNewRight := false
		for _, rightExpr := range rightExprs {
			_, newRightExpr, err := ReconcileSQLExprs(left, rightExpr)
			if err != nil {
				return nil, nil, err
			}
			if rightExpr != newRightExpr {
				hasNewRight = true
			}
			newRightExprs = append(newRightExprs, newRightExpr)
		}

		if hasNewRight {
			right, err = wrapReconciledExprs(right, newRightExprs)
			if err != nil {
				return nil, nil, err
			}
		}

		return left, right, nil
	}

	return nil, nil, fmt.Errorf("left or right expression must be a tuple")
}
