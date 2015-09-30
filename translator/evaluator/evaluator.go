package evaluator

import (
	"errors"
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"github.com/erh/mongo-sql-temp/translator/types"
	"gopkg.in/mgo.v2/bson"
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
	Rows []types.Row
}

func NewSQLValue(gExpr sqlparser.Expr) (SQLValue, error) {
	switch expr := gExpr.(type) {
	case sqlparser.NumVal:
		f, err := strconv.ParseFloat(sqlparser.String(expr), 64)
		if err != nil {
			return nil, err
		}
		return SQLNumeric(f), nil
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
	default:
		panic(fmt.Errorf("NewSQLValue expr not yet implemented for %T", expr))
		return nil, nil
	}
}

// makeMQLQueryPair checks that one of the two provided SQLValues is a field, and the other
// is a literal that can be used in a comparison. It returns the field's SQLValue, then the
// literal's, then a bool indicating if the re-ordered pair is the inverse of the order that was
// passed in. An error is returned if both or neither of the values are SQLFields.
func makeMQLQueryPair(left, right SQLValue) (*SQLField, SQLValue, bool, error) {
	//fmt.Printf("left is %#v right is %#v\n", left, right)
	leftField, leftOk := left.(SQLField)
	rightField, rightOk := right.(SQLField)
	//fmt.Println(leftOk, rightOk)
	if leftOk == rightOk {
		return nil, nil, false, ErrUntransformableCondition
	}
	if leftOk {
		//fmt.Println("returning left field first")
		return &leftField, right, false, nil
	}
	//fmt.Println("returning right field")
	return &rightField, left, true, nil
}

func transformComparison(left, right SQLValue, operator, inverse string) (*bson.D, error) {
	tField, tLiteral, inverted, err := makeMQLQueryPair(left, right)
	if err != nil {
		return nil, err
	}

	mongoOperator := operator
	if inverted {
		mongoOperator = inverse
	}
	return &bson.D{{tField.fieldName, bson.D{{mongoOperator, tLiteral.MongoValue()}}}}, nil
}
