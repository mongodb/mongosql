package evaluator

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"time"

	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/util"
	"github.com/shopspring/decimal"
	"gopkg.in/mgo.v2/bson"
)

// IsSimilar returns true if the logical or comparison
// operations can be carried on leftType against rightType
// with no need for type conversion.
func isSimilar(leftType, rightType schema.SQLType) bool {
	switch leftType {
	case schema.SQLArrNumeric, schema.SQLFloat, schema.SQLInt, schema.SQLInt64, schema.SQLNumeric, schema.SQLUint64:
		switch rightType {
		case schema.SQLArrNumeric, schema.SQLFloat, schema.SQLInt, schema.SQLInt64, schema.SQLNumeric, schema.SQLUint64:
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
		switch v := value.(type) {
		case bool:
			return NewSQLBool(v)
		case bson.ObjectId:
			return SQLTrue
		case bson.Decimal128:
			dec, _ := decimal.NewFromString(v.String())
			flt, _ := dec.Float64()
			return SQLBool(flt)
		case float32, float64, int, int8, int16, int32, int64, uint8, uint16, uint32, uint64:
			flt, _ := util.ToFloat64(v)
			return SQLBool(flt)
		case string:
			flt, _ := strconv.ParseFloat(v, 64)
			return SQLBool(flt)
		case time.Time:
			return SQLTrue
		default:
			return SQLFalse
		}

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
			return NewSQLBool(v), nil
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
			f, err := strconv.ParseFloat(v, 64)
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

// preferentialType accepts a variable number of
// SQLExprs and returns the type of the SQLExpr
// with the highest preference.
func preferentialType(exprs ...SQLExpr) schema.SQLType {
	if len(exprs) == 0 {
		return schema.SQLNone
	}
	var types schema.SQLTypes

	for _, expr := range exprs {
		types = append(types, expr.Type())
	}

	sort.Sort(types)

	return types[len(types)-1]
}

// reconcileSQLExprs takes two SQLExpr and ensures that
// they are of the same type. If they are of different
// types but still comparable, it wraps the SQLExpr with
// a lesser precendence in a SQLConvertExpr. If they are
// not comparable, it returns a non-nil error.
func reconcileSQLExprs(left, right SQLExpr) (SQLExpr, SQLExpr, error) {

	leftType, rightType := left.Type(), right.Type()

	if leftType == schema.SQLTuple || rightType == schema.SQLTuple {
		return reconcileSQLTuple(left, right)
	}

	if leftType == rightType || isSimilar(leftType, rightType) {
		return left, right, nil
	}

	types := schema.SQLTypes{leftType, rightType}
	sort.Sort(types)

	if types[0] == schema.SQLObjectID {
		types[0], types[1] = types[1], types[0]
	}

	if types[1] == leftType {
		right = &SQLConvertExpr{right, types[1]}
	} else {
		left = &SQLConvertExpr{left, types[1]}
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

			newLeftExpr, newRightExpr, err := reconcileSQLExprs(leftExpr, rightExpr)
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

		newLeftExpr, _, err = reconcileSQLExprs(leftExprs[0], right)
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
			_, newRightExpr, err := reconcileSQLExprs(left, rightExpr)
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
