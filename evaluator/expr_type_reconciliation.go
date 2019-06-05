package evaluator

import (
	"sort"

	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/log"
)

const (
	// DateTimeFormat is the datetime formatting string we use to convert
	// timestamps into strings.
	DateTimeFormat = "2006-01-02 15:04:05.000000"
)

type reconciler struct {
	cfg *OptimizerConfig
}

func (v *reconciler) visit(n Node) (Node, error) {
	n, err := walk(v, n)
	if err != nil {
		panic(err)
	}

	if expr, ok := n.(SQLExpr); ok {
		newN, err := expr.reconcile()
		if err == nil {
			return newN, nil
		}
		v.cfg.lg.Warnf(log.Admin, "error running reconcileExprs: %v", err)
	} else if plan, ok := n.(PlanStage); ok {
		if project, ok := plan.(*ProjectStage); ok {
			for _, c := range project.ProjectedColumns() {
				c.Column.EvalType = c.Expr.EvalType()
			}
		}
	}
	return n, nil
}

func reconcileExprs(cfg *OptimizerConfig, n Node) (Node, error) {
	v := &reconciler{cfg: cfg}
	return v.visit(n)
}

// IsSimilar returns true if the logical or comparison
// operations can be carried on leftType against rightType
// with no need for type conversion.
func isSimilar(leftType, rightType types.EvalType) bool {
	if leftType == rightType {
		return true
	}
	if leftType == types.EvalNull || rightType == types.EvalNull {
		return true
	}
	if leftType == types.EvalPolymorphic || rightType == types.EvalPolymorphic {
		return false
	}
	if leftType.IsNumeric() && rightType.IsNumeric() {
		return true
	}
	return false
}

func convertExprs(exprs []SQLExpr, targetTypes []types.EvalType) []SQLExpr {
	if len(targetTypes) < len(exprs) {
		// There is an error in how this function is being used
		panic("targetTypes shorter than exprs")
	}
	newExprs := make([]SQLExpr, len(exprs))
	for i, expr := range exprs {

		targetType := targetTypes[i]
		exprType := expr.EvalType()

		if targetType == types.EvalPolymorphic {
			// types.EvalPolymorphic indicates that there is no need to convert this argument,
			// as it accepts any argument.
			newExprs[i] = expr
		} else if cexpr, ok := expr.(SQLColumnExpr); ok && values.IsUUID(cexpr.columnType.MongoType) {
			// SQLColumnExpr may have a MongoType of UUID, which should be
			// converted to SQLVarchar before converting to targetType.
			newExprs[i] = NewSQLConvertExpr(NewSQLConvertExpr(expr, types.EvalString), targetType)
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

func convertAllExprs(exprs []SQLExpr, targetType types.EvalType) []SQLExpr {
	targetTypes := make([]types.EvalType, len(exprs))
	for i := range exprs {
		targetTypes[i] = targetType
	}
	return convertExprs(exprs, targetTypes)
}

// preferentialType accepts a variable number of
// SQLExprs and returns the type of the SQLExpr
// with the highest preference.
func preferentialType(typers ...types.EvalTyper) types.EvalType {
	s := &types.EvalTypeSorter{}
	return preferentialTypeWithSorter(s, typers...)
}

func preferentialTypeWithSorter(s *types.EvalTypeSorter, typers ...types.EvalTyper) types.EvalType {
	for _, expr := range typers {
		valExpr, ok := expr.(SQLValueExpr)
		if ok && valExpr.Value.IsNull() {
			continue
		}
		s.Types = append(s.Types, expr.EvalType())
	}

	if len(s.Types) == 0 {
		return types.EvalPolymorphic
	}

	sort.Sort(s)

	return s.Types[len(s.Types)-1]
}

// ReconcileSQLExprs takes two SQLExprs and ensures that they are the same type.
// If they are of different types, it wraps the SQLExpr with a lesser precedence
// in a SQLConvertExpr.
func ReconcileSQLExprs(left, right SQLExpr) (SQLExpr, SQLExpr) {
	leftType, rightType := left.EvalType(), right.EvalType()

	if leftType == rightType || isSimilar(leftType, rightType) {
		return left, right
	}

	_, leftIsLiteral := left.(SQLValueExpr)
	_, rightIsLiteral := right.(SQLValueExpr)
	leftIsLiteralStr := leftIsLiteral && left.EvalType() == types.EvalString
	rightIsLiteralStr := rightIsLiteral && right.EvalType() == types.EvalString
	leftIsID := left.EvalType() == types.EvalObjectID
	rightIsID := right.EvalType() == types.EvalObjectID

	if leftIsLiteralStr && rightIsID {
		newLeft := NewSQLConvertExpr(left, types.EvalObjectID)
		return newLeft, right
	} else if rightIsLiteralStr && leftIsID {
		newRight := NewSQLConvertExpr(right, types.EvalObjectID)
		return left, newRight
	}

	sorter := &types.EvalTypeSorter{
		Types: []types.EvalType{leftType, rightType},
	}

	sort.Sort(sorter)

	if sorter.Types[1] == leftType {
		right = NewSQLConvertExpr(right, sorter.Types[1])
	} else {
		left = NewSQLConvertExpr(left, sorter.Types[1])
	}

	return left, right
}
