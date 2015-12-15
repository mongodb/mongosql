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

//
// NewSQLValue is a factory method for creating a SQLValue from a value and column type.
//
func NewSQLValue(value interface{}, columnType config.ColumnType) (SQLValue, error) {

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

// NewSQLExpr transforms sqlparser expressions into SQLExpr.
func NewSQLExpr(gExpr sqlparser.Expr) (SQLExpr, error) {
	log.Logf(log.DebugLow, "match expr: %#v (type is %T)\n", gExpr, gExpr)

	switch expr := gExpr.(type) {
	case nil:
		return SQLTrue, nil

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

	case *sqlparser.ExistsExpr:

		return &SQLExistsExpr{expr.Subquery.Select}, nil

	case *sqlparser.FuncExpr:

		return NewSQLFuncValue(expr)

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

func NewSQLFuncValue(expr *sqlparser.FuncExpr) (SQLExpr, error) {
	if isAggFunction(expr.Name) {
		return &SQLAggFunctionExpr{expr}, nil
	}
	return &SQLScalarFunctionExpr{expr}, nil
}
