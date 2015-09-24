package evaluator

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
)

type SQLField struct {
	tableName string
	fieldName string
}

func NewSQLField(value interface{}) (SQLValue, error) {
	switch v := value.(type) {
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

// string
// number
// date
// time
