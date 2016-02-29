package evaluator

import (
	"fmt"
	"strconv"
)

//
// SQLExpr is the base type for a SQL expression.
//
type SQLExpr interface {
	Evaluate(*EvalCtx) (SQLValue, error)
	String() string
}

//
// SQLValue is a comparable SQLExpr.
//
type SQLValue interface {
	SQLExpr
	CompareTo(SQLValue) (int, error)
	Value() interface{}
}

//
// SQLNumeric is a numeric SQLValue.
//
type SQLNumeric interface {
	SQLValue
	Add(o SQLNumeric) SQLNumeric
	Sub(o SQLNumeric) SQLNumeric
	Product(o SQLNumeric) SQLNumeric
	Float64() float64
}

// A base type for a binary node.
type sqlBinaryNode struct {
	left, right SQLExpr
}

type sqlUnaryNode struct {
	operand SQLExpr
}

//
// EvalCtx holds a slice of rows used to evaluate a SQLValue.
//
type EvalCtx struct {
	Rows    []Row
	ExecCtx *ExecutionCtx
}

// Matches checks if a given SQLExpr is "truthy" by coercing it to a boolean value.
// - booleans: the result is simply that same return value
// - numeric values: the result is true if and only if the value is non-zero.
// - strings, the result is true if and only if that string can be parsed as a number,
//   and that number is non-zero.
func Matches(expr SQLExpr, ctx *EvalCtx) (bool, error) {

	sv, err := expr.Evaluate(ctx)
	if err != nil {
		return false, err
	}

	if asBool, ok := sv.(SQLBool); ok {
		return bool(asBool), nil
	}
	if asNum, ok := sv.(SQLNumeric); ok {
		return asNum.Float64() != float64(0), nil
	}
	if asStr, ok := sv.(SQLString); ok {
		// check if the string should be considered "truthy" by trying to convert it to a number and comparing to 0.
		// more info: http://stackoverflow.com/questions/12221211/how-does-string-truthiness-work-in-mysql
		if parsedFloat, err := strconv.ParseFloat(string(asStr), 64); err == nil {
			return parsedFloat != float64(0), nil
		}
		return false, nil
	}

	// TODO - handle other types with possible values that are "truthy" : dates, etc?
	return false, nil
}

// OptimizeSQLExpr takes a SQLExpr and optimizes it by normalizing
// it into a semantically equivalent tree and partially evaluating
// any subtrees that evaluatable without data.
func OptimizeSQLExpr(e SQLExpr) (SQLExpr, error) {

	newE, err := normalize(e)
	if err != nil {
		return nil, err
	}

	newE, err = partiallyEvaluate(newE)
	if err != nil {
		return nil, err
	}

	if e != newE {
		// normalized and partially evaluated trees might allow for further
		// optimization
		return OptimizeSQLExpr(newE)
	}

	return newE, nil
}

// SQLExprVisitor is an implementation of the Visitor pattern.
type SQLExprVisitor interface {
	// Visit is called with an expression. It returns:
	// - SQLExpr is the expression used to replace the argument.
	// - error
	Visit(SQLExpr) (SQLExpr, error)
}

// walk handles walking the children of the provided expression, calling
// v.Visit on each child. Some visitor implementations may ignore this
// method completely, but most will use it as the default implementation
// for a majority of nodes.
func walk(v SQLExprVisitor, e SQLExpr) (SQLExpr, error) {
	if v == nil || e == nil {
		return nil, nil
	}

	switch typedE := e.(type) {
	case *SQLAggFunctionExpr:
		hasNewChild := false
		newChildren := []SQLExpr{}
		for _, child := range typedE.Exprs {
			newChild, err := v.Visit(child)
			if err != nil {
				return nil, err
			}

			if child != newChild {
				hasNewChild = true
			}

			newChildren = append(newChildren, newChild)
		}

		if hasNewChild {
			e = &SQLAggFunctionExpr{typedE.Name, typedE.Distinct, newChildren}
		}
	case *SQLAddExpr:
		left, err := v.Visit(typedE.left)
		if err != nil {
			return nil, err
		}
		right, err := v.Visit(typedE.right)
		if err != nil {
			return nil, err
		}

		if typedE.left != left || typedE.right != right {
			e = &SQLAddExpr{left, right}
		}

	case *SQLAndExpr:
		left, err := v.Visit(typedE.left)
		if err != nil {
			return nil, err
		}
		right, err := v.Visit(typedE.right)
		if err != nil {
			return nil, err
		}
		if typedE.left != left || typedE.right != right {
			e = &SQLAndExpr{left, right}
		}

	case *SQLCaseExpr:
		hasNewCond := false
		newConds := []caseCondition{}
		for _, cond := range typedE.caseConditions {
			m, err := v.Visit(cond.matcher)
			if err != nil {
				return nil, err
			}
			t, err := v.Visit(cond.then)
			if err != nil {
				return nil, err
			}

			newCond := cond
			if cond.matcher != m || cond.then != t {
				newCond = caseCondition{m, t}
				hasNewCond = true
			}

			newConds = append(newConds, newCond)
		}

		newElse, err := v.Visit(typedE.elseValue)
		if err != nil {
			return nil, err
		}

		if hasNewCond || typedE.elseValue != newElse {
			e = &SQLCaseExpr{newElse, newConds}
		}
	case SQLColumnExpr:
		// no children
	case *SQLDivideExpr:
		left, err := v.Visit(typedE.left)
		if err != nil {
			return nil, err
		}
		right, err := v.Visit(typedE.right)
		if err != nil {
			return nil, err
		}
		if typedE.left != left || typedE.right != right {
			e = &SQLDivideExpr{left, right}
		}

	case *SQLEqualsExpr:
		left, err := v.Visit(typedE.left)
		if err != nil {
			return nil, err
		}
		right, err := v.Visit(typedE.right)
		if err != nil {
			return nil, err
		}
		if typedE.left != left || typedE.right != right {
			e = &SQLEqualsExpr{left, right}
		}

	case *SQLExistsExpr:
		// child isn't visitable
	case *SQLGreaterThanExpr:
		left, err := v.Visit(typedE.left)
		if err != nil {
			return nil, err
		}
		right, err := v.Visit(typedE.right)
		if err != nil {
			return nil, err
		}
		if typedE.left != left || typedE.right != right {
			e = &SQLGreaterThanExpr{left, right}
		}

	case *SQLGreaterThanOrEqualExpr:
		left, err := v.Visit(typedE.left)
		if err != nil {
			return nil, err
		}
		right, err := v.Visit(typedE.right)
		if err != nil {
			return nil, err
		}
		if typedE.left != left || typedE.right != right {
			e = &SQLGreaterThanOrEqualExpr{left, right}
		}

	case *SQLInExpr:

		left, err := v.Visit(typedE.left)
		if err != nil {
			return nil, err
		}
		right, err := v.Visit(typedE.right)
		if err != nil {
			return nil, err
		}

		if typedE.left != left || typedE.right != right {
			e = &SQLInExpr{left, right}
		}

	case *SQLLessThanExpr:
		left, err := v.Visit(typedE.left)
		if err != nil {
			return nil, err
		}
		right, err := v.Visit(typedE.right)
		if err != nil {
			return nil, err
		}
		if typedE.left != left || typedE.right != right {
			e = &SQLLessThanExpr{left, right}
		}

	case *SQLLessThanOrEqualExpr:
		left, err := v.Visit(typedE.left)
		if err != nil {
			return nil, err
		}
		right, err := v.Visit(typedE.right)
		if err != nil {
			return nil, err
		}
		if typedE.left != left || typedE.right != right {
			e = &SQLLessThanOrEqualExpr{left, right}
		}

	case *SQLLikeExpr:
		left, err := v.Visit(typedE.left)
		if err != nil {
			return nil, err
		}
		right, err := v.Visit(typedE.right)
		if err != nil {
			return nil, err
		}
		if typedE.left != left || typedE.right != right {
			e = &SQLLikeExpr{left, right}
		}

	case *SQLMultiplyExpr:
		left, err := v.Visit(typedE.left)
		if err != nil {
			return nil, err
		}
		right, err := v.Visit(typedE.right)
		if err != nil {
			return nil, err
		}
		if typedE.left != left || typedE.right != right {
			e = &SQLMultiplyExpr{left, right}
		}

	case *SQLNotExpr:
		operand, err := v.Visit(typedE.operand)
		if err != nil {
			return nil, err
		}
		if typedE.operand != operand {
			e = &SQLNotExpr{operand}
		}

	case *SQLNotEqualsExpr:
		left, err := v.Visit(typedE.left)
		if err != nil {
			return nil, err
		}
		right, err := v.Visit(typedE.right)
		if err != nil {
			return nil, err
		}
		if typedE.left != left || typedE.right != right {
			e = &SQLNotEqualsExpr{left, right}
		}

	case *SQLNullCmpExpr:
		operand, err := v.Visit(typedE.operand)
		if err != nil {
			return nil, err
		}
		if typedE.operand != operand {
			e = &SQLNullCmpExpr{operand}
		}

	case *SQLOrExpr:
		left, err := v.Visit(typedE.left)
		if err != nil {
			return nil, err
		}
		right, err := v.Visit(typedE.right)
		if err != nil {
			return nil, err
		}
		if typedE.left != left || typedE.right != right {
			e = &SQLOrExpr{left, right}
		}

	case *SQLScalarFunctionExpr:
		hasNewChild := false
		newChildren := []SQLExpr{}
		for _, child := range typedE.Exprs {
			newChild, err := v.Visit(child)
			if err != nil {
				return nil, err
			}

			if child != newChild {
				hasNewChild = true
			}

			newChildren = append(newChildren, newChild)
		}

		if hasNewChild {
			e = &SQLScalarFunctionExpr{typedE.Name, newChildren}
		}
	case *SQLSubqueryCmpExpr:
		left, err := v.Visit(typedE.left)
		if err != nil {
			return nil, err
		}
		sub, err := v.Visit(typedE.value)
		if err != nil {
			return nil, err
		}

		value, ok := sub.(*SQLSubqueryExpr)
		if !ok {
			return nil, fmt.Errorf("SQLSubqueryCmpExpr requires an evaluator.*SQLSubqueryExpr, but got a %T", sub)
		}

		if typedE.left != left || typedE.value != value {
			e = &SQLSubqueryCmpExpr{typedE.In, left, value}
		}

	case *SQLSubqueryExpr:
		// child isn't visitable
	case *SQLSubtractExpr:
		left, err := v.Visit(typedE.left)
		if err != nil {
			return nil, err
		}
		right, err := v.Visit(typedE.right)
		if err != nil {
			return nil, err
		}
		if typedE.left != left || typedE.right != right {
			e = &SQLSubtractExpr{left, right}
		}

	case *SQLUnaryMinusExpr:
		operand, err := v.Visit(typedE.operand)
		if err != nil {
			return nil, err
		}
		if typedE.operand != operand {
			e = &SQLUnaryMinusExpr{operand}
		}

	case *SQLUnaryTildeExpr:
		operand, err := v.Visit(typedE.operand)
		if err != nil {
			return nil, err
		}
		if typedE.operand != operand {
			e = &SQLUnaryTildeExpr{operand}
		}

	case *SQLTupleExpr:
		hasNewChild := false
		newChildren := []SQLExpr{}
		for _, child := range typedE.Exprs {
			newChild, err := v.Visit(child)
			if err != nil {
				return nil, err
			}

			if child != newChild {
				hasNewChild = true
			}

			newChildren = append(newChildren, newChild)
		}

		if hasNewChild {
			e = &SQLTupleExpr{newChildren}
		}

	// values
	case SQLBool:
		// nothing to do
	case *SQLCtorExpr:
		return e.Evaluate(nil)
	case SQLDate:
		// nothing to do
	case SQLDateTime:
		// nothing to do
	case SQLFloat:
		// nothing to do
	case SQLInt:
		// nothing to do
	case SQLNullValue:
		// nothing to do
	case SQLString:
		// nothing to do
	case SQLTime:
		// nothing to do
	case SQLTimestamp:
		// nothing to do
	case *SQLValues:
		hasNewValue := false
		newValues := []SQLValue{}
		for _, value := range typedE.Values {
			newValueExpr, err := v.Visit(value)
			if err != nil {
				return nil, err
			}
			newValue, ok := newValueExpr.(SQLValue)
			if !ok {
				return nil, fmt.Errorf("evaluator.SQLValues requires an evaluator.SQLValue, but got a %T", newValueExpr)
			}

			if value != newValue {
				hasNewValue = true
			}

			newValues = append(newValues, newValue)
		}

		if hasNewValue {
			e = &SQLValues{newValues}
		}

	case SQLUint32:
		// nothing to do
	default:
		return nil, fmt.Errorf("unsupported expression: %T", typedE)
	}

	return e, nil
}
