package evaluator

import (
	"sort"
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

		if targetType == EvalNone {
			// EvalNone indicates that we shouldn't convert this argument
			newExprs[i] = expr
		} else if cexpr, ok := expr.(SQLColumnExpr); ok && IsUUID(cexpr.columnType.MongoType) {
			// SQLColumnExpr may have a MongoType of UUID, which should be
			// converted to SQLVarchar before converting to targetType.
			newExprs[i] = NewSQLConvertExpr(NewSQLConvertExpr(expr, EvalString), targetType)
		} else if isSimilar(exprType, targetType) {
			// don't convert if target type is similar to current type
			newExprs[i] = expr
		} else {
			// convert to the target type
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
		panic("ReconcileSQLExprs should never be called for non-scalar SQLExprs")
	}

	if leftType == rightType || isSimilar(leftType, rightType) {
		return left, right, nil
	}

	_, leftIsStr := left.(SQLVarchar)
	_, rightIsStr := right.(SQLVarchar)
	leftIsID := left.EvalType() == EvalObjectID
	rightIsID := right.EvalType() == EvalObjectID

	if leftIsStr && rightIsID {
		newLeft := NewSQLConvertExpr(left, EvalObjectID)
		return newLeft, right, nil
	} else if rightIsStr && leftIsID {
		newRight := NewSQLConvertExpr(right, EvalObjectID)
		return left, newRight, nil
	}

	sorter := &EvalTypeSorter{
		Types: []EvalType{leftType, rightType},
	}

	if len(preferVarchar) > 0 {
		sorter.VarcharHighPriority = preferVarchar[0]
	}

	sort.Sort(sorter)

	if sorter.Types[1] == leftType {
		right = NewSQLConvertExpr(right, sorter.Types[1])
	} else {
		left = NewSQLConvertExpr(left, sorter.Types[1])
	}

	return left, right, nil
}
