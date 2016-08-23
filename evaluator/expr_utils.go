package evaluator

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
	"github.com/shopspring/decimal"
)

const (
	regexCharsToEscape    = ".^$*+?()[{\\|"
	likePatternEscapeChar = '\\'
)

func compareDecimal128(left, right decimal.Decimal) (int, error) {
	return left.Cmp(right), nil
}

func compareFloats(left, right float64) (int, error) {
	cmp := left - right
	if cmp < 0 {
		return -1, nil
	} else if cmp > 0 {
		return 1, nil
	}
	return 0, nil
}

func convertSQLValueToPattern(value SQLValue) string {
	pattern := value.String()
	regex := "^"
	escaped := false
	for _, c := range pattern {
		if !escaped && c == likePatternEscapeChar {
			escaped = true
			continue
		}

		switch {
		case c == '_':
			if escaped {
				regex += "_"
			} else {
				regex += "."
			}
		case c == '%':
			if escaped {
				regex += "%"
			} else {
				regex += ".*"
			}
		case strings.Contains(regexCharsToEscape, string(c)):
			regex += "\\" + string(c)
		default:
			regex += string(c)
		}

		escaped = false
	}

	regex += "$"

	return regex
}

// doArithmetic performs the given arithmetic operation using
// leftVal and rightVal as operands.
func doArithmetic(leftVal, rightVal SQLValue, op ArithmeticOperator) (SQLValue, error) {

	preferenceType := preferentialType(leftVal, rightVal)
	useDecimal := preferenceType == schema.SQLDecimal128

	// check if both operands are timestamp or date since
	// arithmetic between time types result in an integer
	if preferenceType == schema.SQLDate || preferenceType == schema.SQLTimestamp {
		preferenceType = schema.SQLInt
	}

	var value interface{}

	switch op {
	case ADD:
		if useDecimal {
			value = leftVal.Decimal128().Add(rightVal.Decimal128())
		} else {
			value = leftVal.Float64() + rightVal.Float64()
		}
	case DIV:
		if useDecimal {
			div := leftVal.Decimal128().Div(rightVal.Decimal128())
			return SQLDecimal128(div), nil
		} else {
			return SQLFloat(leftVal.Float64() / rightVal.Float64()), nil
		}
	case MULT:
		if useDecimal {
			value = leftVal.Decimal128().Mul(rightVal.Decimal128())
		} else {
			value = leftVal.Float64() * rightVal.Float64()
		}
	case SUB:
		if useDecimal {
			value = leftVal.Decimal128().Sub(rightVal.Decimal128())
		} else {
			value = leftVal.Float64() - rightVal.Float64()
		}
	default:
		return nil, fmt.Errorf("unrecognized arithmetic operator: %v", op)
	}

	return NewSQLValueFromSQLColumnExpr(value, preferenceType, schema.MongoNone)
}

// fast2Sum returns the exact unevaluated sum of a and b
// where the first member is the float64 nearest the sum
// (ties to even) and the second member is the remainder
// (assuming |b| <= |a|).
//
// T. J. Dekker. A floating-point technique for extending
// the available precision. Numerische Mathematik,
// 18(3):224–242, 1971.
func fast2Sum(a, b float64) (float64, float64) {
	var s, z, t float64
	s = a + b
	z = s - a
	t = b - z
	return s, t
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

func getSQLTupleExprs(left, right SQLExpr) ([]SQLExpr, []SQLExpr, error) {

	getExprs := func(expr SQLExpr) ([]SQLExpr, error) {
		switch typedE := expr.(type) {
		case *SQLTupleExpr:
			return typedE.Exprs, nil
		case *SQLValues:
			var exprs []SQLExpr
			for _, value := range typedE.Values {
				exprs = append(exprs, value)
			}
			return exprs, nil
		default:
			return nil, fmt.Errorf("invalid SQLTupleExpr type '%T'", expr)
		}
	}

	leftExprs, err := getExprs(left)
	if err != nil {
		return nil, nil, err
	}

	rightExprs, err := getExprs(right)
	if err != nil {
		return nil, nil, err
	}

	return leftExprs, rightExprs, nil
}

// hasNullValue returns true if any of the value in values
// is of type SQLNoValue or SQLNullValue.
func hasNullValue(values ...SQLValue) bool {
	for _, value := range values {
		switch v := value.(type) {
		case SQLNoValue, SQLNullValue:
			return true
		case *SQLValues:
			if hasNullValue(v.Values...) {
				return true
			}
		}
	}
	return false
}

// hasNullExpr returns true if any of the expr in exprs
// is of type SQLNoValue or SQLNullValue.
func hasNullExpr(exprs ...SQLExpr) bool {
	for _, e := range exprs {
		switch typedE := e.(type) {
		case SQLNoValue, SQLNullValue:
			return true
		case *SQLTupleExpr:
			return hasNullExpr(typedE.Exprs...)
		case *SQLValues:
			return hasNullValue(typedE.Values...)
		}
	}

	return false
}

func isFalsy(value SQLValue) bool {
	switch v := value.(type) {
	case SQLInt, SQLFloat, SQLUint32, SQLTimestamp, SQLDate, SQLVarchar, SQLObjectID, SQLBool:
		return v.Float64() == float64(0)
	default:
		return false
	}
}

func isTruthy(value SQLValue) bool {
	switch v := value.(type) {
	case SQLInt, SQLFloat, SQLUint32, SQLTimestamp, SQLDate, SQLVarchar, SQLObjectID, SQLBool:
		return v.Float64() != float64(0)
	default:
		return false
	}
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
			return nil, nil, mysqlerrors.Defaultf(mysqlerrors.ER_OPERAND_COLUMNS, numLeft)
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
			return nil, nil, mysqlerrors.Defaultf(mysqlerrors.ER_OPERAND_COLUMNS, len(leftExprs))
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

// round returns the closest integer value to the
// float - round half down for negative values and
// round half up otherwise.
func round(f float64) int64 {
	v := f

	if v < 0.0 {
		v += 0.5
	}

	if f < 0 && v == math.Floor(v) {
		return int64(v - 1)
	}

	return int64(math.Floor(v))
}

func shouldFlip(n sqlBinaryNode) bool {
	if _, ok := n.left.(SQLValue); ok {
		if _, ok := n.right.(SQLValue); !ok {
			return true
		}
	}

	return false
}

// twoSum returns the exact unevaluated sum of a and b,
// where the first member is the double nearest the sum
// (ties to even) and the second member is the remainder.
//
// O. Møller. Quasi double-precision in floating-point
// addition. BIT, 5:37–50, 1965.
//
// D. Knuth. The Art of Computer Programming, vol 2.
// Addison-Wesley, Reading, MA, 3rd ed, 1998.
func twoSum(a, b float64) (float64, float64) {
	var s, aPrime, bPrime, deltaA, deltaB, t float64
	s = a + b
	aPrime = s - b
	bPrime = s - aPrime
	deltaA = a - aPrime
	deltaB = b - bPrime
	t = deltaA + deltaB
	return s, t
}

func getSQLInExprs(right SQLExpr) []SQLExpr {
	var exprs []SQLExpr

	// The right child could be a non-SQLValues SQLValue
	// if the tuple can be evaluated and/or simplified. For
	// example in these sorts of cases: (1), (8-7), (date "2005-03-22").
	// The right child could be of type *SQLValues when each of the
	// expressions in the tuple are evaluated to a SQLValue.
	// Finally, it could be of type *SQLTupleExpr when
	// OptimizeExpr yielded no change.
	sqlValue, isSQLValue := right.(SQLValue)
	sqlValues, isSQLValues := right.(*SQLValues)
	sqlTupleExpr, isSQLTupleExpr := right.(*SQLTupleExpr)

	if isSQLValues {
		for _, value := range sqlValues.Values {
			exprs = append(exprs, value.(SQLExpr))
		}
	} else if isSQLValue {
		exprs = []SQLExpr{sqlValue.(SQLExpr)}
	} else if isSQLTupleExpr {
		exprs = sqlTupleExpr.Exprs
	}

	return exprs
}

func translateTupleExpr(leftExpr, rightExpr SQLExpr, op string) (SQLExpr, error) {
	left, right, err := getSQLTupleExprs(leftExpr, rightExpr)
	if err != nil {
		return nil, err
	}

	if len(left) != len(right) {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ER_OPERAND_COLUMNS, len(left))
	}

	var constructTupleExpr func(string, []SQLExpr, []SQLExpr, bool) (SQLExpr, error)
	constructTupleExpr = func(op string, left, right []SQLExpr, isEqual bool) (SQLExpr, error) {
		if len(left) == 1 {
			return comparisonExpr(left[0], right[0], op)
		} else {
			rightChild, err := constructTupleExpr(op, left[1:], right[1:], isEqual)
			if !isEqual {
				return &SQLOrExpr{&SQLNotEqualsExpr{left[0], right[0]}, rightChild}, err
			}
			return &SQLAndExpr{&SQLEqualsExpr{left[0], right[0]}, rightChild}, err
		}
	}

	var translationFunc func(int) (SQLExpr, error)
	translationFunc = func(i int) (SQLExpr, error) {
		if len(left[i:]) == 0 {
			return SQLFalse, nil
		} else {
			var leftChild SQLExpr
			var err error

			if i == 0 {
				cmpOp := op
				if op == sqlOpLTE {
					cmpOp = sqlOpLT
				} else if op == sqlOpGTE {
					cmpOp = sqlOpGT
				}
				leftChild, err = comparisonExpr(left[0], right[0], cmpOp)
			} else {
				leftChild, err = constructTupleExpr(op, left[:i+1], right[:i+1], true)
			}

			if err != nil {
				return nil, err
			}

			rightChild, err := translationFunc(i + 1)

			return &SQLOrExpr{leftChild, rightChild}, err
		}
	}

	switch op {
	case sqlOpEQ:
		return constructTupleExpr(op, left, right, true)
	case sqlOpNEQ:
		return constructTupleExpr(op, left, right, false)
	default:
		return translationFunc(0)
	}
}
