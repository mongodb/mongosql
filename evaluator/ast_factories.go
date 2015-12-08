package evaluator

import (
	"fmt"
	"github.com/10gen/sqlproxy/config"
	"github.com/erh/mixer/sqlparser"
	"github.com/mongodb/mongo-tools/common/log"
	"github.com/mongodb/mongo-tools/common/util"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"strings"
	"time"
)

func NewSQLField(value interface{}, columnType config.ColumnType) (SQLValue, error) {

	if value == nil {
		return SQLNullValue{}, nil
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
	case config.SQLString:
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
		}
	case config.SQLInt:
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
	case config.SQLFloat:
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
	case config.SQLDatetime:
		v, ok := value.(time.Time)
		if ok {
			return SQLDateTime{v}, nil
		}
	case config.SQLTimestamp:
		v, ok := value.(time.Time)
		if ok {
			return SQLTimestamp{v}, nil
		}
	case config.SQLTime:
		v, ok := value.(time.Time)
		if ok {
			return SQLTime{v}, nil
		}
	case config.SQLDate:
		v, ok := value.(time.Time)
		if ok {
			return SQLDate{v}, nil
		}
	default:
		return nil, fmt.Errorf("unknown column type %v", columnType)
	}
	return nil, fmt.Errorf("unable to convert '%v' (%T) to %v", value, value, columnType)

}

// NewSQLExpr transforms sqlparser expressions into SQLExpr
func NewSQLExpr(gExpr sqlparser.Expr) (SQLExpr, error) {
	log.Logf(log.DebugLow, "match expr: %#v (type is %T)\n", gExpr, gExpr)

	switch expr := gExpr.(type) {
	case nil:

		return &NoopMatcher{}, nil

	case *sqlparser.AndExpr:

		left, err := NewSQLExpr(expr.Left)
		if err != nil {
			return nil, err
		}

		right, err := NewSQLExpr(expr.Right)
		if err != nil {
			return nil, err
		}

		return &And{[]SQLExpr{left, right}}, nil

	case *sqlparser.CaseExpr:

		return NewSQLCaseValue(expr)

	case *sqlparser.ColName:

		return NewSQLValue(expr)

	case *sqlparser.ComparisonExpr:

		left, err := NewSQLValue(expr.Left)
		if err != nil {
			return nil, err
		}

		right, err := NewSQLValue(expr.Right)
		if err != nil {
			return nil, err
		}

		switch expr.Operator {
		case sqlparser.AST_EQ:
			return &Equals{left, right}, nil
		case sqlparser.AST_LT:
			return &LessThan{left, right}, nil
		case sqlparser.AST_GT:
			return &GreaterThan{left, right}, nil
		case sqlparser.AST_LE:
			return &LessThanOrEqual{left, right}, nil
		case sqlparser.AST_GE:
			return &GreaterThanOrEqual{left, right}, nil
		case sqlparser.AST_NE:
			return &NotEquals{left, right}, nil
		case sqlparser.AST_LIKE:
			return &Like{left, right}, nil
		case sqlparser.AST_IN:
			switch eval := right.(type) {
			case *SQLSubqueryValue:
				return &SQLSubqueryCmp{true, left, eval}, nil
			}
			return &In{left, right}, nil
		case sqlparser.AST_NOT_IN:
			switch eval := right.(type) {
			case *SQLSubqueryValue:
				return &SQLSubqueryCmp{false, left, eval}, nil
			}
			return &NotIn{left, right}, nil
		default:
			return &Equals{left, right}, fmt.Errorf("sql where clause not implemented: %s", expr.Operator)
		}

	case *sqlparser.ExistsExpr:

		return &ExistsMatcher{expr.Subquery.Select}, nil

		/*
			case sqlparser.ValArg:
		*/

	case *sqlparser.FuncExpr:

		return NewSQLFuncValue(expr)

	case *sqlparser.NotExpr:

		child, err := NewSQLExpr(expr.Expr)
		if err != nil {
			return nil, err
		}

		return &Not{child}, nil

	case *sqlparser.NullCheck:

		val, err := NewSQLValue(expr.Expr)
		if err != nil {
			return nil, err
		}

		matcher := &NullMatcher{val}
		if expr.Operator == sqlparser.AST_IS_NULL {
			return matcher, nil
		}

		return &Not{matcher}, nil

	case sqlparser.NumVal:

		return NewSQLValue(expr)

	case *sqlparser.OrExpr:

		left, err := NewSQLExpr(expr.Left)
		if err != nil {
			return nil, err
		}

		right, err := NewSQLExpr(expr.Right)
		if err != nil {
			return nil, err
		}

		return &Or{[]SQLExpr{left, right}}, nil

	case *sqlparser.ParenBoolExpr:

		child, err := NewSQLExpr(expr.Expr)
		if err != nil {
			return nil, err
		}

		return child, nil

	case *sqlparser.RangeCond:

		from, err := NewSQLValue(expr.From)
		if err != nil {
			return nil, err
		}

		left, err := NewSQLValue(expr.Left)
		if err != nil {
			return nil, err
		}

		to, err := NewSQLValue(expr.To)
		if err != nil {
			return nil, err
		}

		lower := &GreaterThanOrEqual{left, from}

		upper := &LessThanOrEqual{left, to}

		m := &And{[]SQLExpr{lower, upper}}

		if expr.Operator == sqlparser.AST_NOT_BETWEEN {
			return &Not{m}, nil
		}

		return m, nil

	case sqlparser.StrVal:

		return NewSQLValue(expr)

	case *sqlparser.Subquery:

		return &ExistsMatcher{expr.Select}, nil

	case *sqlparser.UnaryExpr:

		return NewSQLValue(expr.Expr)

	default:
		panic(fmt.Errorf("matcher not yet implemented for %v (%T)", sqlparser.String(expr), expr))
	}
}

func NewSQLValue(gExpr sqlparser.Expr) (SQLValue, error) {
	switch expr := gExpr.(type) {

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

	case *sqlparser.ColName:

		return SQLField{string(expr.Qualifier), string(expr.Name)}, nil

	case sqlparser.StrVal:

		return SQLString(string([]byte(expr))), nil

	case *sqlparser.BinaryExpr:

		// look up the function in the function map
		funcImpl, ok := binaryFuncMap[string(expr.Operator)]
		if !ok {
			return nil, fmt.Errorf("can't find implementation for binary operator '%v'", expr.Operator)
		}

		left, err := NewSQLValue(expr.Left)
		if err != nil {
			return nil, err
		}

		right, err := NewSQLValue(expr.Right)
		if err != nil {
			return nil, err
		}

		return &SQLBinaryValue{[]SQLValue{left, right}, funcImpl}, nil

	case *sqlparser.FuncExpr:

		return NewSQLFuncValue(expr)

	case *sqlparser.ParenBoolExpr:

		return &SQLParenBoolValue{expr}, nil

	case sqlparser.ValTuple:

		var values []SQLValue

		for _, e := range expr {
			value, err := NewSQLValue(e)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}

		return &SQLValTupleValue{values}, nil

	case *sqlparser.UnaryExpr:

		val, err := NewSQLValue(expr.Expr)
		if err != nil {
			return nil, err
		}

		switch expr.Operator {

		case sqlparser.AST_UMINUS:

			return &SQLUnaryMinus{val}, nil

		case sqlparser.AST_UPLUS:

			return &SQLUnaryPlus{val}, nil

		case sqlparser.AST_TILDA:

			return &SQLUnaryTilde{val}, nil

		}

		return nil, fmt.Errorf("invalid unary operator - '%v'", string(expr.Operator))

	case *sqlparser.CaseExpr:

		return NewSQLCaseValue(expr)

	case *sqlparser.Subquery:

		return &SQLSubqueryValue{expr.Select}, nil

	case *sqlparser.NullVal, nil:

		return &SQLNullValue{}, nil

	default:

		panic(fmt.Errorf("NewSQLValue expr not yet implemented for %T", expr))
		return nil, nil
	}
}

// NewSQLCaseValue returns a SQLValue for SQL case expressions of which there are
// two kinds.
//
// For simple case expressions, we create an equality matcher that compares
// the expression against each value in the list of cases.
//
// For searched case expressions, we create a matcher based on the boolean
// expression in each when condition.
//
func NewSQLCaseValue(expr *sqlparser.CaseExpr) (SQLValue, error) {

	var e SQLValue

	var err error

	if expr.Expr != nil {
		e, err = NewSQLValue(expr.Expr)
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
			c, err := NewSQLValue(when.Cond)
			if err != nil {
				return nil, err
			}

			matcher = &Equals{e, c}
		}

		then, err := NewSQLValue(when.Val)
		if err != nil {
			return nil, err
		}

		conditions = append(conditions, caseCondition{matcher, then})
	}

	elseValue, err := NewSQLValue(expr.Else)
	if err != nil {
		return nil, err
	}

	value := &SQLCaseValue{
		elseValue:      elseValue,
		caseConditions: conditions,
	}

	// TODO: You cannot specify the literal NULL for every return expr
	// and the else expr.
	return value, nil
}

func NewSQLFuncValue(expr *sqlparser.FuncExpr) (SQLValue, error) {
	if isAggFunction(expr.Name) {
		return &SQLAggFuncValue{expr}, nil
	}
	return &SQLScalarFuncValue{expr}, nil
}
