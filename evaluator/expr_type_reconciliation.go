package evaluator

import (
	"fmt"
	"sort"

	"github.com/10gen/sqlproxy/internal/mysqlerrors"
)

const (
	// DateTimeFormat is the datetime formatting string we use to convert
	// timestamps into strings.
	DateTimeFormat = "2006-01-02 15:04:05.000000"
)

// IsSimilar returns true if the logical or comparison
// operations can be carried on leftType against rightType
// with no need for type conversion.
func isSimilar(leftType, rightType EvalType) bool {
	if leftType == rightType {
		return true
	}
	if leftType == EvalNull || rightType == EvalNull {
		return true
	}
	if leftType == EvalNone || rightType == EvalNone {
		return true
	}
	if leftType.IsNumeric() && rightType.IsNumeric() {
		return true
	}
	if leftType.IsDate() && rightType.IsDate() {
		return true
	}
	return false
}

func convertExprs(exprs []SQLExpr, targetTypes []EvalType) []SQLExpr {
	if len(targetTypes) < len(exprs) {
		// There is an error in how this function is being used
		panic("targetTypes shorter than exprs")
	}
	newExprs := make([]SQLExpr, len(exprs))
	for i, expr := range exprs {
		targetType := targetTypes[i]
		exprType := expr.EvalType()
		// SQLColumnExpr may have a MongoType of UUID or ObjectID, which
		// should be converted to SQLVarchar before converting to targetType.
		if cexpr, ok := expr.(SQLColumnExpr); ok &&
			isIDType(cexpr.columnType.MongoType) {
			newExprs[i] = NewSQLConvertExpr(NewSQLConvertExpr(expr, EvalString),
				targetType)
		} else if isSimilar(exprType, targetType) {
			newExprs[i] = expr
		} else {
			newExprs[i] = NewSQLConvertExpr(expr, targetType)
		}
	}
	return newExprs
}

func convertAllExprs(exprs []SQLExpr, targetType EvalType) []SQLExpr {
	targetTypes := make([]EvalType, len(exprs))
	for i := range exprs {
		targetTypes[i] = targetType
	}
	return convertExprs(exprs, targetTypes)
}

// preferentialType accepts a variable number of
// SQLExprs and returns the type of the SQLExpr
// with the highest preference.
func preferentialType(exprs ...SQLExpr) EvalType {
	s := &EvalTypeSorter{}
	return preferentialTypeWithSorter(s, exprs...)
}

func preferentialTypeWithSorter(s *EvalTypeSorter, exprs ...SQLExpr) EvalType {
	for _, expr := range exprs {
		val, ok := expr.(SQLValue)
		if ok && val.IsNull() {
			continue
		}
		s.Types = append(s.Types, expr.EvalType())
	}

	if len(s.Types) == 0 {
		return EvalNone
	}

	sort.Sort(s)

	return s.Types[len(s.Types)-1]
}

// ReconcileSQLExprs takes two SQLExpr and ensures that
// they are of the same type. If they are of different
// types but still comparable, it wraps the SQLExpr with
// a lesser precedence in a SQLConvertExpr. If they are
// not comparable, it returns a non-nil error. The optional
// argument preferVarchar causes reconilation to varchar
// if any of the types is a varchar/EvalString.
func ReconcileSQLExprs(left, right SQLExpr, preferVarchar ...bool) (SQLExpr, SQLExpr, error) {
	leftType, rightType := left.EvalType(), right.EvalType()

	if leftType == EvalTuple || rightType == EvalTuple {
		return reconcileTuple(left, right)
	}

	if leftType == rightType || isSimilar(leftType, rightType) {
		return left, right, nil
	}

	sorter := &EvalTypeSorter{
		Types: []EvalType{leftType, rightType},
	}

	if len(preferVarchar) > 0 {
		sorter.VarcharHighPriority = preferVarchar[0]
	}

	sort.Sort(sorter)

	if sorter.Types[0] == EvalObjectID {
		sorter.Types[0], sorter.Types[1] = sorter.Types[1], sorter.Types[0]
	}

	if sorter.Types[1] == leftType {
		right = NewSQLConvertExpr(right, sorter.Types[1])
	} else {
		left = NewSQLConvertExpr(left, sorter.Types[1])
	}

	return left, right, nil
}

func reconcileTuple(left, right SQLExpr) (SQLExpr, SQLExpr, error) {

	getSQLExprs := func(expr SQLExpr) ([]SQLExpr, error) {
		switch typedE := expr.(type) {
		case *SQLTupleExpr:
			return typedE.Exprs, nil
		case *SQLSubqueryExpr:
			return typedE.Exprs(), nil
		}
		return nil, fmt.Errorf("cannot reconcile non-tuple type '%T'", expr)
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
				allowRows:  typedE.allowRows,
			}, nil
		}
		return nil, fmt.Errorf("cannot wrap reconciled non-tuple type '%T'", expr)
	}

	var leftExprs []SQLExpr
	var rightExprs []SQLExpr
	var err error

	if left.EvalType() == EvalTuple {
		leftExprs, err = getSQLExprs(left)
		if err != nil {
			return nil, nil, err
		}
	}

	if right.EvalType() == EvalTuple {
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
	if left.EvalType() == EvalTuple && right.EvalType() == EvalTuple {

		numLeft, numRight := len(leftExprs), len(rightExprs)

		if numLeft != numRight && numLeft != 1 {
			return nil, nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, numLeft)
		}

		hasNewLeft := false
		hasNewRight := false

		for i := range rightExprs {
			leftExpr := leftExprs[0]
			if numLeft != 1 {
				leftExpr = leftExprs[i]
			}

			rightExpr := rightExprs[i]

			var newLeftExpr SQLExpr
			var newRightExpr SQLExpr
			newLeftExpr, newRightExpr, err = ReconcileSQLExprs(leftExpr, rightExpr)
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
	if left.EvalType() == EvalTuple && right.EvalType() != EvalTuple {

		if len(leftExprs) != 1 {
			return nil, nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, len(leftExprs))
		}

		var newLeftExpr SQLExpr

		newLeftExpr, _, err = ReconcileSQLExprs(leftExprs[0], right)
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
	if left.EvalType() != EvalTuple && right.EvalType() == EvalTuple {

		hasNewRight := false
		for _, rightExpr := range rightExprs {
			var newRightExpr SQLExpr
			_, newRightExpr, err = ReconcileSQLExprs(left, rightExpr)
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
