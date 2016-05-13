package evaluator

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
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

// NewSQLExpr transforms sqlparser expressions into SQLExpr.
func NewSQLExpr(sqlExpr sqlparser.Expr, tables map[string]*schema.Table) (SQLExpr, error) {
	log.Logf(log.DebugHigh, "planning expr: %#v (type is %T)\n", sqlExpr, sqlExpr)

	switch expr := sqlExpr.(type) {

	case *sqlparser.AndExpr:

		left, err := NewSQLExpr(expr.Left, tables)
		if err != nil {
			return nil, err
		}

		right, err := NewSQLExpr(expr.Right, tables)
		if err != nil {
			return nil, err
		}

		left, right, err = reconcileSQLExprs(left, right)
		if err != nil {
			return nil, err
		}

		return &SQLAndExpr{left, right}, nil

	case *sqlparser.BinaryExpr:

		left, err := NewSQLExpr(expr.Left, tables)
		if err != nil {
			return nil, err
		}

		right, err := NewSQLExpr(expr.Right, tables)
		if err != nil {
			return nil, err
		}

		left, right, err = reconcileSQLExprs(left, right)
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

		return newSQLCaseExpr(expr, tables)

	case *sqlparser.ColName:

		tableName, columnName := string(expr.Qualifier), strings.ToLower(string(expr.Name))

		columnType := getColumnType(tables, tableName, columnName)

		return SQLColumnExpr{tableName, columnName, *columnType}, nil

	case *sqlparser.ComparisonExpr:

		left, err := NewSQLExpr(expr.Left, tables)
		if err != nil {
			return nil, err
		}

		right, err := NewSQLExpr(expr.Right, tables)
		if err != nil {
			return nil, err
		}

		switch expr.Operator {
		case sqlparser.AST_EQ:
			left, right, err = reconcileSQLExprs(left, right)
			if err != nil {
				return nil, err
			}
			return &SQLEqualsExpr{left, right}, nil
		case sqlparser.AST_LT:
			left, right, err = reconcileSQLExprs(left, right)
			if err != nil {
				return nil, err
			}
			return &SQLLessThanExpr{left, right}, nil
		case sqlparser.AST_GT:
			left, right, err = reconcileSQLExprs(left, right)
			if err != nil {
				return nil, err
			}

			return &SQLGreaterThanExpr{left, right}, nil
		case sqlparser.AST_LE:
			left, right, err = reconcileSQLExprs(left, right)
			if err != nil {
				return nil, err
			}
			return &SQLLessThanOrEqualExpr{left, right}, nil
		case sqlparser.AST_GE:
			left, right, err = reconcileSQLExprs(left, right)
			if err != nil {
				return nil, err
			}
			return &SQLGreaterThanOrEqualExpr{left, right}, nil
		case sqlparser.AST_NE:
			left, right, err = reconcileSQLExprs(left, right)
			if err != nil {
				return nil, err
			}
			return &SQLNotEqualsExpr{left, right}, nil
		case sqlparser.AST_LIKE:
			if err != nil {
				return nil, err
			}
			return &SQLLikeExpr{left, right}, nil
		case sqlparser.AST_IN:
			left, right, err = reconcileSQLExprs(left, right)
			if err != nil {
				return nil, err
			}

			if eval, ok := right.(*SQLSubqueryExpr); ok {
				return &SQLSubqueryCmpExpr{true, left, eval}, nil
			}

			return &SQLInExpr{left, right}, nil
		case sqlparser.AST_NOT_IN:
			left, right, err = reconcileSQLExprs(left, right)
			if err != nil {
				return nil, err
			}

			if eval, ok := right.(*SQLSubqueryExpr); ok {
				return &SQLSubqueryCmpExpr{true, left, eval}, nil
			}

			return &SQLNotExpr{&SQLInExpr{left, right}}, nil
		default:
			return nil, fmt.Errorf("sql where clause not implemented: %s", expr.Operator)
		}

	case *sqlparser.CtorExpr:

		ctor := &SQLCtorExpr{Name: expr.Name, Args: expr.Exprs}
		return ctor.Evaluate(nil)

	case *sqlparser.ExistsExpr:
		return &SQLExistsExpr{expr.Subquery.Select}, nil

	case *sqlparser.FuncExpr:

		return newSQLFuncExpr(expr, tables)

	case nil:

		return SQLTrue, nil

	case *sqlparser.NotExpr:

		child, err := NewSQLExpr(expr.Expr, tables)
		if err != nil {
			return nil, err
		}

		return &SQLNotExpr{child}, nil

	case *sqlparser.NullCheck:

		val, err := NewSQLExpr(expr.Expr, tables)
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

		left, err := NewSQLExpr(expr.Left, tables)
		if err != nil {
			return nil, err
		}

		right, err := NewSQLExpr(expr.Right, tables)
		if err != nil {
			return nil, err
		}

		return &SQLOrExpr{left, right}, nil

	case *sqlparser.ParenBoolExpr:

		return NewSQLExpr(expr.Expr, tables)

	case *sqlparser.RangeCond:

		from, err := NewSQLExpr(expr.From, tables)
		if err != nil {
			return nil, err
		}

		left, err := NewSQLExpr(expr.Left, tables)
		if err != nil {
			return nil, err
		}

		to, err := NewSQLExpr(expr.To, tables)
		if err != nil {
			return nil, err
		}

		left, from, err = reconcileSQLExprs(left, from)
		if err != nil {
			return nil, err
		}

		lower := &SQLGreaterThanOrEqualExpr{left, from}

		left, to, err = reconcileSQLExprs(left, to)
		if err != nil {
			return nil, err
		}

		upper := &SQLLessThanOrEqualExpr{left, to}

		m := &SQLAndExpr{lower, upper}

		if expr.Operator == sqlparser.AST_NOT_BETWEEN {
			return &SQLNotExpr{m}, nil
		}

		return m, nil

	case sqlparser.StrVal:

		return SQLVarchar(string([]byte(expr))), nil

	case *sqlparser.Subquery:

		sExprs, err := referencedSelectExpressions(expr.Select, tables)
		if err != nil {
			return nil, err
		}

		var exprs []SQLExpr

		for _, sExpr := range sExprs {
			exprs = append(exprs, sExpr.Expr)
		}

		return &SQLSubqueryExpr{expr.Select, exprs, nil}, nil

	case *sqlparser.UnaryExpr:

		val, err := NewSQLExpr(expr.Expr, tables)
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
			newExpr, err := NewSQLExpr(e, tables)
			if err != nil {
				return nil, err
			}

			exprs = append(exprs, newExpr)
		}

		if len(exprs) == 1 {
			return exprs[0], nil
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
func newSQLCaseExpr(expr *sqlparser.CaseExpr, tables map[string]*schema.Table) (SQLExpr, error) {

	var e SQLExpr

	var err error

	if expr.Expr != nil {
		e, err = NewSQLExpr(expr.Expr, tables)
		if err != nil {
			return nil, err
		}
	}

	var conditions []caseCondition

	var matcher SQLExpr

	for _, when := range expr.Whens {

		// searched case
		if expr.Expr == nil {
			matcher, err = NewSQLExpr(when.Cond, tables)
			if err != nil {
				return nil, err
			}
		} else {
			// TODO: support simple case in parser
			c, err := NewSQLExpr(when.Cond, tables)
			if err != nil {
				return nil, err
			}

			matcher = &SQLEqualsExpr{e, c}
		}

		then, err := NewSQLExpr(when.Val, tables)
		if err != nil {
			return nil, err
		}

		conditions = append(conditions, caseCondition{matcher, then})
	}

	var elseValue SQLExpr
	if expr.Else == nil {
		elseValue = SQLNull
	} else if elseValue, err = NewSQLExpr(expr.Else, tables); err != nil {
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

func newSQLFuncExpr(expr *sqlparser.FuncExpr, tables map[string]*schema.Table) (SQLExpr, error) {

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

			exprs = append(exprs, SQLVarchar("*"))

		case *sqlparser.NonStarExpr:

			sqlExpr, err := NewSQLExpr(typedE.Expr, tables)
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

			sqlExpr, err := NewSQLExpr(typedE.Expr, tables)
			if err != nil {
				return nil, err
			}

			exprs = append(exprs, sqlExpr)

			switch name {
			case "cast":
				exprs = append(exprs, SQLVarchar(string(typedE.As)))
			}

		}

	}

	return &SQLScalarFunctionExpr{name, exprs}, nil
}
