package evaluator

import (
	"fmt"
	"github.com/10gen/sqlproxy/schema"
	"github.com/deafgoat/mixer/sqlparser"
	"github.com/mongodb/mongo-tools/common/log"
	"github.com/mongodb/mongo-tools/common/util"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"strings"
	"time"
)

//
// NewSQLValue is a factory method for creating a SQLValue from a value and column type.
//
func NewSQLValue(value interface{}, columnType string) (SQLValue, error) {

	if value == nil {
		return SQLNull, nil
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

	// TODO: columnType can - and should - be more expressive. e.g. certain
	// types allow you to specify additional constraints on the value of
	// the type. e.g. unsigned longs, binary, character set for varchars,
	// precision in time values, etc

	// We should have this specified in the DRDL file, and use it when we're
	// formatting the fields for the query response.

	case schema.SQLString:
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
		case bson.ObjectId:
			return SQLString(v.Hex()), nil
		}

	case schema.SQLInt:
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

	case schema.SQLFloat:
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

	//
	// Time related types
	//
	// Note: MongoDB only supports time to millisecond precision when
	// using the UTC datetime type while MySQL supports microsecond
	// precision.
	//
	// See https://dev.mysql.com/doc/refman/5.5/en/date-and-time-types.html
	// for more information on date and time types.
	//

	case schema.SQLDate:

		lower := time.Date(1000, time.January, 1, 0, 0, 0, 0, schema.DefaultLocale)
		upper := time.Date(9999, time.December, 31, 0, 0, 0, 0, schema.DefaultLocale)

		var date time.Time

		switch v := value.(type) {

		case time.Time:

			v = v.In(schema.DefaultLocale)

			date = time.Date(v.Year(), v.Month(), v.Day(), 0, 0, 0, 0, schema.DefaultLocale)

		case string:
			d, err := time.Parse(schema.DateFormat, v)
			if err != nil {
				return SQLDate{schema.DefaultTime}, nil
			}
			date = time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, schema.DefaultLocale)

		case bson.ObjectId:

			return NewSQLValue(v.Time(), columnType)

		default:

			return SQLDate{schema.DefaultTime}, nil

		}

		if date.Before(lower) || date.After(upper) {
			return SQLDate{schema.DefaultTime}, nil
		}

		return SQLDate{date}, nil

	case schema.SQLDateTime:

		lower := time.Date(1000, time.January, 1, 0, 0, 0, 0, schema.DefaultLocale)
		upper := time.Date(9999, time.December, 31, 23, 59, 59, 0, schema.DefaultLocale)

		var dt time.Time

		switch v := value.(type) {

		case time.Time:

			dt = v.In(schema.DefaultLocale)

		case string:

			d, err := time.Parse(schema.TimestampFormat, v)
			if err != nil {
				return SQLDate{schema.DefaultTime}, nil
			}

			dt = time.Date(d.Year(), d.Month(), d.Day(), d.Hour(),
				d.Minute(), d.Second(), d.Nanosecond(), schema.DefaultLocale)

		case bson.ObjectId:

			return NewSQLValue(v.Time(), columnType)

		default:

			return SQLDateTime{schema.DefaultTime}, nil

		}

		if dt.Before(lower) || dt.After(upper) {
			return SQLDateTime{schema.DefaultTime}, nil
		}

		return SQLDateTime{dt}, nil

	case schema.SQLTimestamp:

		lower := time.Date(1970, time.January, 1, 0, 0, 01, 0, schema.DefaultLocale)
		upper := time.Date(2038, time.January, 19, 03, 14, 07, 0, schema.DefaultLocale)

		var ts time.Time

		switch v := value.(type) {

		case time.Time:

			ts = v.In(schema.DefaultLocale)

		case string:

			for _, format := range schema.TimestampCtorFormats {
				d, err := time.Parse(format, v)
				if err == nil {
					ts = time.Date(d.Year(), d.Month(), d.Day(), d.Hour(),
						d.Minute(), d.Second(), d.Nanosecond(), schema.DefaultLocale)
					break
				}
			}

			if ts.Equal(time.Time{}) {
				return SQLTimestamp{schema.DefaultTime}, nil
			}

		case bson.ObjectId:

			return NewSQLValue(v.Time(), columnType)

		default:

			return SQLTimestamp{schema.DefaultTime}, nil

		}

		if ts.Before(lower) || ts.After(upper) {
			return SQLTimestamp{schema.DefaultTime}, nil
		}

		return SQLTimestamp{ts}, nil

	// TODO: not sure if it makes sense to support this type
	// in the config since there isn't really a reasonable
	// mapping between it and any of MongoDB's BSON types.
	case schema.SQLTime:
		v, ok := value.(time.Time)
		if ok {
			return SQLTime{v}, nil
		}

	// See http://dev.mysql.com/doc/refman/5.7/en/year.html
	case schema.SQLYear:
		switch v := value.(type) {

		case time.Time:
			return SQLYear{v.In(schema.DefaultLocale)}, nil

		case int, int32, int64:
			year, err := getYear(value.(int), true)
			if err != nil {
				return nil, err
			}
			return SQLYear{time.Date(year, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)}, nil

		case string:
			if v == "00" || v == "0" {
				return SQLYear{time.Date(2000, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)}, nil
			}

			y, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return SQLYear{schema.DefaultTime}, nil
			}

			year, err := getYear(int(y), false)
			if err != nil {
				return nil, err
			}

			return SQLYear{time.Date(year, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)}, nil

		case bson.ObjectId:

			return NewSQLValue(v.Time(), columnType)

		default:
			return SQLYear{schema.DefaultTime}, nil
		}

	default:
		// programmer error: should never happen
		panic(fmt.Sprintf("unimplemented column type %v", columnType))
	}

	return nil, fmt.Errorf("unable to convert '%v' (%T) to %v", value, value, columnType)
}

func getYear(y int, isNum bool) (int, error) {
	err := fmt.Errorf("invalid year: %v", y)

	// handle error cases
	if y < 0 {
		return 0, err
	}
	if y > 99 && y < 1901 {
		return 0, err
	}
	if y > 2155 {
		return 0, err
	}

	if isNum {
		if y < 70 && y > 0 {
			return 2000 + y, nil
		}
		if y > 69 && y < 100 {
			return 1970 + y, nil
		}
	} else {
		if y < 70 {
			return 2000 + y, nil
		}
		if y > 69 && y < 100 {
			return 1970 + y, nil
		}
	}

	return 0, nil
}

// NewSQLExpr transforms sqlparser expressions into SQLExpr.
func NewSQLExpr(gExpr sqlparser.Expr) (SQLExpr, error) {
	log.Logf(log.DebugLow, "match expr: %#v (type is %T)\n", gExpr, gExpr)

	switch expr := gExpr.(type) {

	case *sqlparser.AndExpr:

		left, err := NewSQLExpr(expr.Left)
		if err != nil {
			return nil, err
		}

		right, err := NewSQLExpr(expr.Right)
		if err != nil {
			return nil, err
		}

		return &SQLAndExpr{left, right}, nil

	case *sqlparser.BinaryExpr:

		left, err := NewSQLExpr(expr.Left)
		if err != nil {
			return nil, err
		}

		right, err := NewSQLExpr(expr.Right)
		if err != nil {
			return nil, err
		}

		switch expr.Operator {
		case '+':
			return &SQLAddExpr{left, right}, nil
		case '-':
			return &SQLSubtractExpr{left, right}, nil
		case '*':
			return &SQLMultiplyExpr{left, right}, nil
		case '/':
			return &SQLDivideExpr{left, right}, nil
		default:
			return nil, fmt.Errorf("can't find implementation for binary operator '%v'", expr.Operator)
		}

	case *sqlparser.CaseExpr:

		return newSQLCaseExpr(expr)

	case *sqlparser.ColName:

		return SQLFieldExpr{string(expr.Qualifier), string(expr.Name)}, nil

	case *sqlparser.ComparisonExpr:

		left, err := NewSQLExpr(expr.Left)
		if err != nil {
			return nil, err
		}

		right, err := newSQLExprForComparison(expr.Right)
		if err != nil {
			return nil, err
		}

		switch expr.Operator {
		case sqlparser.AST_EQ:
			return &SQLEqualsExpr{left, right}, nil
		case sqlparser.AST_LT:
			return &SQLLessThanExpr{left, right}, nil
		case sqlparser.AST_GT:
			return &SQLGreaterThanExpr{left, right}, nil
		case sqlparser.AST_LE:
			return &SQLLessThanOrEqualExpr{left, right}, nil
		case sqlparser.AST_GE:
			return &SQLGreaterThanOrEqualExpr{left, right}, nil
		case sqlparser.AST_NE:
			return &SQLNotEqualsExpr{left, right}, nil
		case sqlparser.AST_LIKE:
			return &SQLLikeExpr{left, right}, nil
		case sqlparser.AST_IN:
			switch eval := right.(type) {
			case *SQLSubqueryExpr:
				return &SQLSubqueryCmpExpr{true, left, eval}, nil
			}
			return &SQLInExpr{left, right}, nil
		case sqlparser.AST_NOT_IN:
			switch eval := right.(type) {
			case *SQLSubqueryExpr:
				return &SQLSubqueryCmpExpr{false, left, eval}, nil
			}
			return &SQLNotExpr{&SQLInExpr{left, right}}, nil
		default:
			return &SQLEqualsExpr{left, right}, fmt.Errorf("sql where clause not implemented: %s", expr.Operator)
		}

	case *sqlparser.CtorExpr:

		return &SQLCtorExpr{Name: expr.Name, Args: expr.Exprs}, nil

	case *sqlparser.ExistsExpr:

		return &SQLExistsExpr{expr.Subquery.Select}, nil

	case *sqlparser.FuncExpr:

		return newSQLFuncExpr(expr)

	case nil:

		return SQLTrue, nil

	case *sqlparser.NotExpr:

		child, err := NewSQLExpr(expr.Expr)
		if err != nil {
			return nil, err
		}

		return &SQLNotExpr{child}, nil

	case *sqlparser.NullCheck:

		val, err := NewSQLExpr(expr.Expr)
		if err != nil {
			return nil, err
		}

		// TODO: Why can't this just be an Equals expression?
		matcher := &SQLNullCmpExpr{val}
		if expr.Operator == sqlparser.AST_IS_NULL {
			return matcher, nil
		}

		return &SQLNotExpr{matcher}, nil

	case *sqlparser.NullVal:

		return SQLNull, nil

	case sqlparser.NumVal:

		// try to parse as int first
		if i, err := strconv.ParseInt(sqlparser.String(expr), 10, 64); err == nil {
			return SQLInt(i), nil
		}

		// if it's not a valid int, try parsing as float instead
		f, err := strconv.ParseFloat(sqlparser.String(expr), 64)
		if err != nil {
			return nil, err
		}

		return SQLFloat(f), nil

	case *sqlparser.OrExpr:

		left, err := NewSQLExpr(expr.Left)
		if err != nil {
			return nil, err
		}

		right, err := NewSQLExpr(expr.Right)
		if err != nil {
			return nil, err
		}

		return &SQLOrExpr{left, right}, nil

	case *sqlparser.ParenBoolExpr:

		return NewSQLExpr(expr.Expr)

	case *sqlparser.RangeCond:

		from, err := NewSQLExpr(expr.From)
		if err != nil {
			return nil, err
		}

		left, err := NewSQLExpr(expr.Left)
		if err != nil {
			return nil, err
		}

		to, err := NewSQLExpr(expr.To)
		if err != nil {
			return nil, err
		}

		lower := &SQLGreaterThanOrEqualExpr{left, from}

		upper := &SQLLessThanOrEqualExpr{left, to}

		m := &SQLAndExpr{lower, upper}

		if expr.Operator == sqlparser.AST_NOT_BETWEEN {
			return &SQLNotExpr{m}, nil
		}

		return m, nil

	case sqlparser.StrVal:

		return SQLString(string([]byte(expr))), nil

	case *sqlparser.Subquery:

		return &SQLExistsExpr{expr.Select}, nil

	case *sqlparser.UnaryExpr:

		val, err := NewSQLExpr(expr.Expr)
		if err != nil {
			return nil, err
		}

		switch expr.Operator {
		case sqlparser.AST_UMINUS:
			return &SQLUnaryMinusExpr{val}, nil
		case sqlparser.AST_TILDA:
			return &SQLUnaryTildeExpr{val}, nil
		}

		return nil, fmt.Errorf("invalid unary operator - '%v'", string(expr.Operator))

	case sqlparser.ValTuple:

		var exprs []SQLExpr

		for _, e := range expr {
			newExpr, err := NewSQLExpr(e)
			if err != nil {
				return nil, err
			}

			exprs = append(exprs, newExpr)
		}

		return &SQLTupleExpr{exprs}, nil

	default:
		panic(fmt.Errorf("NewSQLExpr not yet implemented for %v (%T)", sqlparser.String(expr), expr))
	}
}

func newSQLExprForComparison(gExpr sqlparser.Expr) (SQLExpr, error) {

	switch expr := gExpr.(type) {

	case *sqlparser.Subquery:
		return &SQLSubqueryExpr{expr.Select}, nil

	default:
		return NewSQLExpr(gExpr)
	}
}

// NewSQLCaseExpr returns a SQLValue for SQL case expressions of which there are
// two kinds.
//
// For simple case expressions, we create an equality matcher that compares
// the expression against each value in the list of cases.
//
// For searched case expressions, we create a matcher based on the boolean
// expression in each when condition.
//
func newSQLCaseExpr(expr *sqlparser.CaseExpr) (SQLExpr, error) {

	var e SQLExpr

	var err error

	if expr.Expr != nil {
		e, err = NewSQLExpr(expr.Expr)
		if err != nil {
			return nil, err
		}
	}

	var conditions []caseCondition

	var matcher SQLExpr

	for _, when := range expr.Whens {

		// searched case
		if expr.Expr == nil {
			matcher, err = NewSQLExpr(when.Cond)
			if err != nil {
				return nil, err
			}
		} else {
			// TODO: support simple case in parser
			c, err := NewSQLExpr(when.Cond)
			if err != nil {
				return nil, err
			}

			matcher = &SQLEqualsExpr{e, c}
		}

		then, err := NewSQLExpr(when.Val)
		if err != nil {
			return nil, err
		}

		conditions = append(conditions, caseCondition{matcher, then})
	}

	var elseValue SQLExpr
	if expr.Else == nil {
		elseValue = SQLNull
	} else if elseValue, err = NewSQLExpr(expr.Else); err != nil {
		return nil, err
	}

	value := &SQLCaseExpr{
		elseValue:      elseValue,
		caseConditions: conditions,
	}

	// TODO: You cannot specify the literal NULL for every return expr
	// and the else expr.
	return value, nil
}

func newSQLFuncExpr(expr *sqlparser.FuncExpr) (SQLExpr, error) {

	exprs := []SQLExpr{}

	name := string(expr.Name)

	if isAggFunction(expr.Name) {

		for _, e := range expr.Exprs {

			switch typedE := e.(type) {
			// TODO: mixture of star and non-star expression acceptable?

			case *sqlparser.StarExpr:

				if name != "count" {
					return nil, fmt.Errorf("%v aggregate function can not contain '*'", name)
				}

				exprs = append(exprs, SQLInt(1))

			case *sqlparser.NonStarExpr:

				sqlExpr, err := NewSQLExpr(typedE.Expr)
				if err != nil {
					return nil, err
				}
				exprs = append(exprs, sqlExpr)

			}
		}

		return &SQLAggFunctionExpr{name, expr.Distinct, exprs}, nil
	}

	for _, e := range expr.Exprs {

		switch typedE := e.(type) {

		case *sqlparser.StarExpr:

			switch name {
			case "isnull", "not", "pow":
				return nil, fmt.Errorf("argument to '%v' function can not contain '*'", name)
			}

		case *sqlparser.NonStarExpr:

			sqlExpr, err := NewSQLExpr(typedE.Expr)
			if err != nil {
				return nil, err
			}

			exprs = append(exprs, sqlExpr)

		}

	}

	switch name {
	case "isnull", "not":
		if len(exprs) != 1 {
			return nil, fmt.Errorf("'%v' function requires exactly one argument", name)
		}
	case "pow":
		if len(exprs) != 2 {
			return nil, fmt.Errorf("'%v' function requires exactly two arguments", name)
		}
	}

	return &SQLScalarFunctionExpr{name, exprs}, nil
}
