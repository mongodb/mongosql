package evaluator

import (
	"context"
	"fmt"

	"github.com/10gen/mongoast/ast"

	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/astutil"
	"github.com/10gen/sqlproxy/internal/bsonutil"
)

type sqlUnaryNode struct {
	expr SQLExpr
}

// Children returns a slice of all the Node children of the Node.
func (u sqlUnaryNode) Children() []Node {
	return []Node{u.expr}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (u *sqlUnaryNode) ReplaceChild(i int, n Node) {
	if i != 0 {
		panicWithInvalidIndex("sqlUnaryNode", i, 0)
	}
	u.expr = panicIfNotSQLExpr("sqlUnaryNode", n)
}

// SQLNotExpr evaluates to the inverse of its child.
type SQLNotExpr struct {
	sqlUnaryNode
}

var _ translatableToMatch = (*SQLNotExpr)(nil)

// NewSQLNotExpr is a constructor for SQLNotExpr.
func NewSQLNotExpr(operand SQLExpr) *SQLNotExpr {
	return &SQLNotExpr{sqlUnaryNode{operand}}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLNotExpr) ExprName() string {
	return "SQLNotExpr"
}

// Evaluate evaluates a SQLNotExpr into a values.SQLValue.
func (not *SQLNotExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(not)
	if err != nil {
		return nil, err
	}

	operand, err := not.expr.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if values.HasNullValue(operand) {
		return values.NewSQLNull(cfg.sqlValueKind), nil
	}

	if !values.Bool(operand) {
		return values.NewSQLBool(cfg.sqlValueKind, true), nil
	}

	return values.NewSQLBool(cfg.sqlValueKind, false), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLNotExpr.
func (not *SQLNotExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(not); err != nil {
		return nil, err
	}

	if sqlVal, ok := not.expr.(SQLValueExpr); ok {
		if sqlVal.Value.IsNull() {
			return NewSQLValueExpr(values.NewSQLNull(cfg.sqlValueKind)), nil
		}
		if !values.Bool(sqlVal.Value) {
			return NewSQLValueExpr(values.NewSQLBool(cfg.sqlValueKind, true)), nil
		}
		return NewSQLValueExpr(values.NewSQLBool(cfg.sqlValueKind, false)), nil
	}
	return not, nil
}

// nolint: unparam
func (not *SQLNotExpr) reconcile() (SQLExpr, error) {
	expr := not.expr
	if !isBooleanComparable(expr.EvalType()) {
		expr = NewSQLConvertExpr(expr, types.EvalBoolean)
	}
	return NewSQLNotExpr(expr), nil
}

func (not *SQLNotExpr) String() string {
	return fmt.Sprintf("not %v", not.expr)
}

// ToAggregationLanguage translates SQLNotExpr into something that can
// be used in an aggregation pipeline. If SQLNotExpr cannot be translated,
// it will return nil and error.
func (not *SQLNotExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	args, err := t.translateArgs([]SQLExpr{not.expr})
	if err != nil {
		return nil, err
	}

	assignments, args := minimizeLetAssignments([]string{"op"}, args)

	evaluation := wrapInNullCheckedCond(
		t.ColumnsToNullCheck(),
		astutil.NullLiteral,
		ast.NewFunction(bsonutil.OpNot, args[0]),
		args[0],
	)

	return wrapInLet(assignments, evaluation), nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (not *SQLNotExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return not.ToAggregationLanguage(t)
}

// ToMatchLanguage translates SQLNotExpr into something that can
// be used in an match expression. If SQLNotExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLNotExpr.
func (not *SQLNotExpr) ToMatchLanguage(t *PushdownTranslator) (ast.Expr, SQLExpr) {
	match, ex := t.ToMatchLanguage(not.expr)
	if match == nil {
		return nil, not
	} else if ex == nil {
		return negate(match), nil
	} else {
		// partial translation of Not
		return negate(match), NewSQLNotExpr(ex)
	}

}

// EvalType returns the EvalType associated with SQLNotExpr.
func (*SQLNotExpr) EvalType() types.EvalType {
	return types.EvalBoolean
}

// SQLUnaryMinusExpr evaluates to the negation of the expression.
type SQLUnaryMinusExpr struct {
	sqlUnaryNode
}

// NewSQLUnaryMinusExpr is a constructor for SQLUnaryMinusExpr.
func NewSQLUnaryMinusExpr(operand SQLExpr) *SQLUnaryMinusExpr {
	return &SQLUnaryMinusExpr{sqlUnaryNode{operand}}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLUnaryMinusExpr) ExprName() string {
	return "SQLUnaryMinusExpr"
}

// Evaluate evaluates a SQLUnaryMinusExpr into a values.SQLValue.
func (um *SQLUnaryMinusExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(um)
	if err != nil {
		return nil, err
	}

	if val, err := um.expr.Evaluate(ctx, cfg, st); err == nil {
		if val.IsNull() {
			return values.NewSQLNull(cfg.sqlValueKind), nil
		}
		difference := values.NewSQLFloat(cfg.sqlValueKind, -values.Float64(val))
		converted := values.ConvertTo(difference, um.EvalType())
		return converted, nil
	}
	return nil, fmt.Errorf("UnaryMinus expression does not apply to a %T", um.expr)
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLUnaryMinusExpr.
func (um *SQLUnaryMinusExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(um); err != nil {
		return nil, err
	}

	if val, ok := um.expr.(SQLValueExpr); ok {
		if val.Value.IsNull() {
			return NewSQLValueExpr(values.NewSQLNull(cfg.sqlValueKind)), nil
		}
		return NewSQLValueExpr(values.ConvertTo(values.NewSQLFloat(cfg.sqlValueKind, -values.Float64(val.Value)), um.EvalType())), nil
	}
	return um, nil
}

// nolint: unparam
func (um *SQLUnaryMinusExpr) reconcile() (SQLExpr, error) {
	child := um.expr
	typ := child.EvalType()
	if typ.IsNumeric() || typ == types.EvalNull {
		return um, nil
	}
	if typ == types.EvalString {
		return NewSQLUnaryMinusExpr(NewSQLConvertExpr(child, types.EvalDouble)), nil
	}
	return NewSQLUnaryMinusExpr(NewSQLConvertExpr(child, types.EvalInt64)), nil
}

func (um *SQLUnaryMinusExpr) String() string {
	return fmt.Sprintf("-%v", um.expr)
}

// ToAggregationLanguage translates SQLUnaryMinusExpr into something that can
// be used in an aggregation pipeline. If SQLUnaryMinusExpr cannot be translated,
// it will return nil and error.
func (um *SQLUnaryMinusExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	args, err := t.translateArgs([]SQLExpr{um.expr})
	if err != nil {
		return nil, err
	}

	assignments, args := minimizeLetAssignments([]string{"operand"}, args)
	evaluation := wrapInNullCheckedCond(
		t.ColumnsToNullCheck(),
		astutil.NullLiteral,
		ast.NewBinary(bsonutil.OpMultiply, astutil.Int32Value(-1), args[0]),
		args[0],
	)

	return wrapInLet(assignments, evaluation), nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (um *SQLUnaryMinusExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return um.ToAggregationLanguage(t)
}

// EvalType returns the EvalType associated with SQLUnaryMinusExpr.
func (um *SQLUnaryMinusExpr) EvalType() types.EvalType {
	return um.expr.EvalType()
}

// SQLTildeExpr invert all bits in the operand
// and returns an unsigned 64-bit integer.
type SQLTildeExpr struct {
	sqlUnaryNode
}

// NewSQLTildeExpr is a constructor for SQLTildeExpr.
func NewSQLTildeExpr(operand SQLExpr) *SQLTildeExpr {
	return &SQLTildeExpr{sqlUnaryNode{operand}}
}

// EvalType returns the EvalType associated with SQLTildeExpr.
func (td *SQLTildeExpr) EvalType() types.EvalType {
	return td.expr.EvalType()
}

// Evaluate evaluates a SQLTildeExpr into a values.SQLValue.
func (td *SQLTildeExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(td)
	if err != nil {
		return nil, err
	}

	val, err := td.expr.Evaluate(ctx, cfg, st)
	if err != nil {
		return values.NewSQLBool(cfg.sqlValueKind, false), err
	}

	if val.IsNull() {
		return values.NewSQLNull(cfg.sqlValueKind), nil
	}
	return values.NewSQLUint64(cfg.sqlValueKind, ^uint64(values.Int64(val))), nil
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLTildeExpr) ExprName() string {
	return "SQLTildeExpr"
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLTildeExpr.
func (td *SQLTildeExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(td); err != nil {
		return nil, err
	}

	if val, ok := td.expr.(SQLValueExpr); ok {
		if val.Value.IsNull() {
			return NewSQLValueExpr(values.NewSQLNull(cfg.sqlValueKind)), nil
		}
		return NewSQLValueExpr(values.NewSQLUint64(cfg.sqlValueKind, ^uint64(values.Int64(val.Value)))), nil
	}
	return td, nil
}

// nolint: unparam
func (td *SQLTildeExpr) reconcile() (SQLExpr, error) {
	child := td.expr
	typ := child.EvalType()
	if typ.IsNumeric() || typ == types.EvalNull {
		return td, nil
	}
	return NewSQLTildeExpr(NewSQLConvertExpr(child, types.EvalInt64)), nil
}

func (td *SQLTildeExpr) String() string {
	return fmt.Sprintf("~%v", td.expr)
}

// ToAggregationLanguage translates SQLTildeExpr into something that can
// be used in an aggregation pipeline. If SQLTildeExpr cannot be translated,
// it will return nil and error.
func (*SQLTildeExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return nil, newPushdownFailure(
		"SQLTildeExpr",
		"cannot push down to MongoDB",
	)
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (td *SQLTildeExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return td.ToAggregationLanguage(t)
}
