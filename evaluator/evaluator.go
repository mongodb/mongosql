package evaluator

import (
	"errors"
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"strconv"
)

// BinaryNode holds two SQLValues.
type BinaryNode struct {
	left, right SQLValue
}

// SQLValue used in computation by matchers
type SQLValue interface {
	Evaluate(*EvalCtx) (SQLValue, error)
	MongoValue() interface{}
	Comparable
}

type Comparable interface {
	CompareTo(*EvalCtx, SQLValue) (int, error)
}

var ErrTypeMismatch = errors.New("type mismatch")

// EvalCtx holds a slice of rows used to evaluate a SQLValue.
type EvalCtx struct {
	Rows    []Row
	ExecCtx *ExecutionCtx
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

			return &UMinus{val}, nil

		case sqlparser.AST_UPLUS:

			return &UPlus{val}, nil

		case sqlparser.AST_TILDA:

			return &Tilda{val}, nil

		}

		return nil, fmt.Errorf("invalid unary operator - '%v'", string(expr.Operator))

	case *sqlparser.CaseExpr:

		return NewSQLCaseValue(expr)

	case nil:

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

	var matcher Matcher

	for _, when := range expr.Whens {

		// searched case
		if expr.Expr == nil {
			matcher, err = BuildMatcher(when.Cond)
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
