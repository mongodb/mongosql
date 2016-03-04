package evaluator

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/10gen/sqlproxy/schema"
	"github.com/deafgoat/mixer/sqlparser"
	"github.com/mongodb/mongo-tools/common/log"
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
		case nil, []interface{}:
			return SQLNull, nil
		case bson.ObjectId:
			return SQLString(v.Hex()), nil
		case bool:
			return SQLBool(v), nil
		case string:
			return SQLString(v), nil
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
			return SQLString(strconv.FormatBool(v)), nil
		case string:
			return SQLString(v), nil
		case float32:
			return SQLString(strconv.FormatFloat(float64(v), 'f', -1, 32)), nil
		case float64:
			return SQLString(strconv.FormatFloat(v, 'f', -1, 64)), nil
		case int:
			return SQLString(strconv.FormatInt(int64(v), 10)), nil
		case int8:
			return SQLString(strconv.FormatInt(int64(v), 10)), nil
		case int16:
			return SQLString(strconv.FormatInt(int64(v), 10)), nil
		case int32:
			return SQLString(strconv.FormatInt(int64(v), 10)), nil
		case int64:
			return SQLString(strconv.FormatInt(int64(v), 10)), nil
		case uint8:
			return SQLString(strconv.FormatInt(int64(v), 10)), nil
		case uint16:
			return SQLString(strconv.FormatInt(int64(v), 10)), nil
		case uint32:
			return SQLString(strconv.FormatInt(int64(v), 10)), nil
		case uint64:
			return SQLString(strconv.FormatInt(int64(v), 10)), nil
		case bson.ObjectId:
			return SQLString(v.Hex()), nil
		case time.Time:
			return SQLString(v.String()), nil
		case nil:
			return SQLNull, nil
		default:
			return SQLString(reflect.ValueOf(v).String()), nil
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

	case schema.SQLFloat, schema.SQLNumeric:
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

		return SQLColumnExpr{string(expr.Qualifier), string(expr.Name)}, nil

	case *sqlparser.ComparisonExpr:

		left, err := NewSQLExpr(expr.Left)
		if err != nil {
			return nil, err
		}

		right, err := NewSQLExpr(expr.Right)
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

		return &SQLSubqueryExpr{expr.Select}, nil

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

		if len(expr.Exprs) != 1 {
			return nil, fmt.Errorf("aggregate function can not contain tuples")
		}

		e := expr.Exprs[0]

		switch typedE := e.(type) {
		// TODO: mixture of star and non-star expression acceptable?

		case *sqlparser.StarExpr:

			if name != "count" {
				return nil, fmt.Errorf("%v aggregate function can not contain '*'", name)
			} else {
				if expr.Distinct {
					return nil, fmt.Errorf("count aggregate function can not have distinct '*'")
				}
			}

			exprs = append(exprs, SQLString("*"))

		case *sqlparser.NonStarExpr:

			sqlExpr, err := NewSQLExpr(typedE.Expr)
			if err != nil {
				return nil, err
			}
			exprs = append(exprs, sqlExpr)

		}

		return &SQLAggFunctionExpr{name, expr.Distinct, exprs}, nil
	}

	for _, e := range expr.Exprs {

		switch typedE := e.(type) {

		case *sqlparser.StarExpr:
			switch name {
			case "count":
			default:
				return nil, fmt.Errorf("argument to '%v' function can not contain '*'", name)
			}

		case *sqlparser.NonStarExpr:

			sqlExpr, err := NewSQLExpr(typedE.Expr)
			if err != nil {
				return nil, err
			}

			exprs = append(exprs, sqlExpr)

			switch name {
			case "cast":
				exprs = append(exprs, SQLString(string(typedE.As)))
			}

		}

	}

	return &SQLScalarFunctionExpr{name, exprs}, nil
}
