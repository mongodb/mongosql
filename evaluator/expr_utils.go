package evaluator

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/10gen/sqlproxy/schema"
)

func convertToSQLNumeric(expr SQLExpr, ctx *EvalCtx) (SQLNumeric, error) {
	eval, err := expr.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	switch v := eval.(type) {
	case SQLNumeric:
		return v, nil
	case *SQLValues:
		if len(v.Values) != 1 {
			return nil, fmt.Errorf("expected only one SQLValues value - got %v", len(v.Values))
		}
		return convertToSQLNumeric(v.Values[0], ctx)
	default:
		return nil, nil
	}
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

// hasNoSQLValue returns true if any of the value in values
// is of type SQLNoValue.
func hasNoSQLValue(values ...SQLValue) bool {
	for _, value := range values {
		switch v := value.(type) {
		case SQLNoValue:
			return true
		case *SQLValues:
			if hasNoSQLValue(v.Values...) {
				return true
			}
		}
	}
	return false
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

	if leftType == schema.SQLTuple || rightType == schema.SQLTuple {
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

		hasNewLeft := false
		hasNewRight := false

		for i := range rightExprs {
			leftExpr := leftExprs[0]
			if numLeft != 1 {
				leftExpr = leftExprs[i]
			}

			rightExpr := rightExprs[i]

			newLeftExpr, newRightExpr, err := reconcileSQLExprs(leftExpr, rightExpr)
			if err != nil {
				return nil, nil, err

			}

			if leftExpr != newLeftExpr {
				hasNewLeft = true
			}

			if rightExpr != newRightExpr {
				hasNewRight = true
			}

			newLeftExprs = append(newLeftExprs, newLeftExpr)
			newRightExprs = append(newRightExprs, newRightExpr)
		}

		if numLeft == 1 {
			newLeftExprs = newLeftExprs[:1]
		}

		if hasNewLeft {
			left, err = wrapReconciledExprs(left, newLeftExprs)
			if err != nil {
				return nil, nil, err
			}
		}

		if hasNewRight {
			right, err = wrapReconciledExprs(right, newRightExprs)
			if err != nil {
				return nil, nil, err
			}
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

		newLeftExpr, _, err = reconcileSQLExprs(leftExprs[0], right)
		if err != nil {
			return nil, nil, err
		}

		if left != newLeftExpr {
			newLeftExprs = append(newLeftExprs, newLeftExpr)
			left, err = wrapReconciledExprs(left, newLeftExprs)
			if err != nil {
				return nil, nil, err
			}
		}

		return left, right, nil
	}

	// cases here:
	// a = (1)
	// a = (SELECT a FROM foo)
	// a in (1, 2)
	if left.Type() != schema.SQLTuple && right.Type() == schema.SQLTuple {

		hasNewRight := false
		for _, rightExpr := range rightExprs {
			_, newRightExpr, err := reconcileSQLExprs(left, rightExpr)
			if err != nil {
				return nil, nil, err
			}
			if rightExpr != newRightExpr {
				hasNewRight = true
			}
			newRightExprs = append(newRightExprs, newRightExpr)
		}

		if hasNewRight {
			right, err = wrapReconciledExprs(right, newRightExprs)
			if err != nil {
				return nil, nil, err
			}
		}

		return left, right, nil
	}

	return nil, nil, fmt.Errorf("left or right expression must be a tuple")
}

func sqlValueToString(sqlValue SQLValue) (string, error) {
	switch v := sqlValue.(type) {
	case SQLVarchar:
		return string(v), nil
	case SQLNumeric:
		switch t := v.(type) {
		case SQLFloat:
			return strconv.FormatFloat(t.Float64(), 'f', -1, 64), nil
		default:
			return strconv.FormatInt(int64(t.Float64()), 10), nil
		}
	case SQLDate:
		return string(v.String()), nil
	case SQLTimestamp:
		return string(v.String()), nil
	}

	// TODO: just return empty string with no error?
	return "", fmt.Errorf("unable to convert %v to string", sqlValue)
}
