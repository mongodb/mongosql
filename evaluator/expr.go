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
	node
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
		switch typedE := expr.(type) {
		case *SQLTupleExpr:
			return &SQLTupleExpr{newExprs}, nil
		case *SQLSubqueryExpr:
			plan := typedE.plan

			var projectedColumns ProjectedColumns
			for i, c := range plan.Columns() {
				projectedColumns = append(projectedColumns, ProjectedColumn{
					Column: c,
					Expr:   newExprs[i],
				})
			}

			return &SQLSubqueryExpr{
				correlated: typedE.correlated,
				plan:       NewProjectStage(plan, projectedColumns...),
			}, nil
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
