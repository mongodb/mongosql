package evaluator

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/10gen/sqlproxy/schema"
	"github.com/mongodb/mongo-tools/common/util"
	"gopkg.in/mgo.v2/bson"
)

// newSQLTimeValue is a factory method for creating a
// time SQLValue from a time object.
func newSQLTimeValue(t time.Time, mongoType schema.MongoType) (SQLValue, error) {

	// Time objects can be reliably converted into one of the following:
	//
	// 1. SQLTimestamp
	// 2. SQLDate
	//

	h := t.Hour()
	mi := t.Minute()
	sd := t.Second()
	ns := t.Nanosecond()

	if ns == 0 && sd == 0 && mi == 0 && h == 0 {
		return NewSQLValue(t, schema.SQLDate, mongoType)
	}

	// SQLDateTime and SQLTimestamp can be handled similarly
	return NewSQLValue(t, schema.SQLTimestamp, mongoType)
}

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
			return newSQLTimeValue(v, mongoType)
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
