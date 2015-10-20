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
		return &SQLBinaryExprValue{[]SQLValue{left, right}, funcImpl}, nil
	case *sqlparser.FuncExpr:
		return &SQLFuncExpr{expr}, nil
	case *sqlparser.ParenBoolExpr:
		return &SQLParenBoolExpr{expr}, nil
	case sqlparser.ValTuple:
		var values []SQLValue
		for _, e := range expr {
			value, err := NewSQLValue(e)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}
		return &SQLValTupleExpr{values}, nil

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
	case nil:
		return &SQLNullValue{}, nil
	default:
		panic(fmt.Errorf("NewSQLValue expr not yet implemented for %T", expr))
		return nil, nil
	}
}
