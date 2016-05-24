package evaluator

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/10gen/sqlproxy/schema"
)

//
// SQLExpr is the base type for a SQL expression.
//
type SQLExpr interface {
	Evaluate(*EvalCtx) (SQLValue, error)
	String() string
	Type() schema.SQLType
}

//
// SQLValue is a SQLExpr with a value.
//
type SQLValue interface {
	SQLExpr
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

	eval, err := expr.Evaluate(ctx)
	if err != nil {
		return false, err
	}

	switch v := eval.(type) {
	case SQLBool:
		return bool(v), nil
	case SQLNumeric:
		return v.Float64() != float64(0), nil
	case SQLVarchar:
		// more info: http://stackoverflow.com/questions/12221211/how-does-string-truthiness-work-in-mysql
		p, err := strconv.ParseFloat(string(v), 64)
		if err == nil {
			return p != float64(0), nil
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
	case *SQLConvertExpr:
		expr, err := v.Visit(typedE.expr)
		if err != nil {
			return nil, err
		}
		if typedE.expr != expr {
			e = &SQLConvertExpr{expr, typedE.convType}
		}
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
	case SQLDate:
		// nothing to do
	case SQLFloat:
		// nothing to do
	case SQLInt:
		// nothing to do
	case SQLNullValue:
		// nothing to do
	case SQLVarchar:
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

// getColumnType accepts a table name and a column name
// and returns the column type for the given column if
// it is found in the tables map. If it is not found, it
// returns a default column type.
func getColumnType(tables map[string]*schema.Table, tableName, columnName string) *schema.ColumnType {

	none := &schema.ColumnType{schema.SQLNone, schema.MongoNone}

	if tables == nil {
		return none
	}

	table, ok := tables[tableName]
	if !ok {
		return none
	}

	column, ok := table.SQLColumns[columnName]
	if !ok {
		return none
	}

	return &schema.ColumnType{column.SqlType, column.MongoType}
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

	_, leftIsTuple := left.(*SQLTupleExpr)
	_, leftIsSubquery := left.(*SQLSubqueryExpr)

	_, rightIsTuple := right.(*SQLTupleExpr)
	_, rightIsSubquery := right.(*SQLSubqueryExpr)

	if leftIsTuple || rightIsTuple || leftIsSubquery || rightIsSubquery {
		return reconcileSQLTuple(left, right)
	}

	if leftType == rightType || schema.IsSimilar(leftType, rightType) {
		return left, right, nil
	}

	if !schema.CanCompare(leftType, rightType) {
		return nil, nil, fmt.Errorf("cannot compare '%v' type against '%v' type", leftType, rightType)
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
		return nil, fmt.Errorf("can not reconcile non-tuple type '%T'", expr)
	}

	wrapReconciledExprs := func(expr SQLExpr, newExprs []SQLExpr) (SQLExpr, error) {
		switch expr.(type) {
		case *SQLTupleExpr:
			return &SQLTupleExpr{newExprs}, nil
		case *SQLSubqueryExpr:
			// TODO: fix this to wrap the subquery in a SQLConvertExpr...
			//return &SQLSubqueryExpr{typedE.stmt, newExprs, false, nil}, nil
		}
		return nil, fmt.Errorf("can not wrap reconciled non-tuple type '%T'", expr)
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
			return nil, nil, fmt.Errorf("tuple comparison mismatch: expected %v got %v", numLeft, numRight)
		}

		for i, _ := range rightExprs {
			leftExpr := leftExprs[0]
			if numLeft != 1 {
				leftExpr = leftExprs[i]
			}

			rightExpr := rightExprs[i]

			newLeftExpr, newRightExpr, err := reconcileSQLExprs(leftExpr, rightExpr)
			if err != nil {
				return nil, nil, err

			}

			newRightExprs = append(newRightExprs, newRightExpr)
			newLeftExprs = append(newLeftExprs, newLeftExpr)
		}

		if numLeft == 1 {
			newLeftExprs = newLeftExprs[:1]
		}

		left, err = wrapReconciledExprs(left, newLeftExprs)
		if err != nil {
			return nil, nil, err
		}

		right, err = wrapReconciledExprs(right, newRightExprs)
		if err != nil {
			return nil, nil, err
		}

		return left, right, nil
	}

	// cases here:
	// (a) = 1
	// (SELECT a FROM foo) = 1
	if left.Type() == schema.SQLTuple && right.Type() != schema.SQLTuple {

		if len(leftExprs) != 1 {
			return nil, nil, fmt.Errorf("left 'in' operand must have only one value - got %v", len(leftExprs))
		}

		var newLeftExpr SQLExpr

		newLeftExpr, right, err = reconcileSQLExprs(leftExprs[0], right)
		if err != nil {
			return nil, nil, err
		}

		newLeftExprs = append(newLeftExprs, newLeftExpr)

		left, err = wrapReconciledExprs(left, newLeftExprs)
		if err != nil {
			return nil, nil, err
		}

		return left, right, nil
	}

	// cases here:
	// a = (1)
	// a = (SELECT a FROM foo)
	// a in (1, 2)
	if left.Type() != schema.SQLTuple && right.Type() == schema.SQLTuple {

		for _, rightExpr := range rightExprs {
			_, newRightExpr, err := reconcileSQLExprs(left, rightExpr)
			if err != nil {
				return nil, nil, err
			}
			newRightExprs = append(newRightExprs, newRightExpr)
		}

		right, err = wrapReconciledExprs(right, newRightExprs)
		if err != nil {
			return nil, nil, err
		}

		return left, right, nil
	}

	return nil, nil, fmt.Errorf("left or right expression must be a tuple")
}
