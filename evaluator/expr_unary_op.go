package evaluator

import (
	"context"
	"fmt"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/bsonutil"
)

type sqlUnaryNode struct {
	expr SQLExpr
}

// Children returns the arguments.
func (u sqlUnaryNode) Children() []SQLExpr {
	return []SQLExpr{u.expr}
}

// ReplaceChild sets the argument.
func (u *sqlUnaryNode) ReplaceChild(i int, e SQLExpr) {
	if i != 0 {
		panic(fmt.Sprintf("unary nodes only have one child, index %v is out of range", i))
	}
	u.expr = e
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

// Evaluate evaluates a SQLNotExpr into a SQLValue.
func (not *SQLNotExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	operand, err := not.expr.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if hasNullValue(operand) {
		return NewSQLNull(cfg.sqlValueKind, not.EvalType()), nil
	}

	if !Bool(operand) {
		return NewSQLBool(cfg.sqlValueKind, true), nil
	}

	return NewSQLBool(cfg.sqlValueKind, false), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLNotExpr.
func (not *SQLNotExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	if sqlVal, ok := not.expr.(SQLValue); ok {
		if sqlVal.IsNull() {
			return sqlVal
		}
		if !Bool(sqlVal) {
			return NewSQLBool(cfg.sqlValueKind, true)
		}
		return NewSQLBool(cfg.sqlValueKind, false)
	}
	return not
}

// Normalize will attempt to change SQLNotExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (not *SQLNotExpr) Normalize(kind SQLValueKind) Node {
	if operand, ok := not.expr.(SQLValue); ok {
		if hasNullValue(operand) {
			return NewSQLNull(kind, not.EvalType())
		}

		if Bool(operand) {
			return NewSQLBool(kind, false)
		} else if IsFalsy(operand) {
			return NewSQLBool(kind, true)
		}
	}

	return not
}

// nolint: unparam
func (not *SQLNotExpr) reconcile() (SQLExpr, error) {
	expr := not.expr
	if !isBooleanComparable(expr.EvalType()) {
		expr = NewSQLConvertExpr(expr, EvalBoolean)
	}
	return NewSQLNotExpr(expr), nil
}

func (not *SQLNotExpr) String() string {
	return fmt.Sprintf("not %v", not.expr)
}

// ToAggregationLanguage translates SQLNotExpr into something that can
// be used in an aggregation pipeline. If SQLNotExpr cannot be translated,
// it will return nil and error.
func (not *SQLNotExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	args, err := t.typedTranslateArgs([]SQLExpr{not.expr})
	if err != nil {
		return nil, err
	}

	assignments, args := minimizeLetAssignments([]string{"op"}, args)

	if args[0].t == argLiteralType {
		switch args[0].v {
		case nil:
			return bsonutil.MgoNullLiteral, nil
		case false, 0:
			return true, nil
		default:
			return false, nil
		}
	}

	evaluation := wrapInNullCheckedCond(
		t.ColumnsToNullCheck(),
		bsonutil.MgoNullLiteral,
		bsonutil.WrapInOp(bsonutil.OpNot, args[0].v),
		args...,
	)

	return wrapInLet(assignments, evaluation), nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (not *SQLNotExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return not.ToAggregationLanguage(t)
}

// ToMatchLanguage translates SQLNotExpr into something that can
// be used in an match expression. If SQLNotExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLNotExpr.
func (not *SQLNotExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
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
func (*SQLNotExpr) EvalType() EvalType {
	return EvalBoolean
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

// Evaluate evaluates a SQLUnaryMinusExpr into a SQLValue.
func (um *SQLUnaryMinusExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	if val, err := um.expr.Evaluate(ctx, cfg, st); err == nil {
		if val.IsNull() {
			return NewSQLNull(cfg.sqlValueKind, um.EvalType()), nil
		}
		difference := NewSQLFloat(cfg.sqlValueKind, -Float64(val))
		converted := ConvertTo(difference, um.EvalType())
		return converted, nil
	}
	return nil, fmt.Errorf("UnaryMinus expression does not apply to a %T", um.expr)
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLUnaryMinusExpr.
func (um *SQLUnaryMinusExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	if val, ok := um.expr.(SQLValue); ok {
		if val.IsNull() {
			return NewSQLNull(cfg.sqlValueKind, um.EvalType())
		}
		return ConvertTo(NewSQLFloat(cfg.sqlValueKind, -Float64(val)), um.EvalType())
	}
	return um
}

// Normalize will attempt to change SQLUnaryMinusExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (um *SQLUnaryMinusExpr) Normalize(kind SQLValueKind) Node {
	sqlVal, ok := um.expr.(SQLValue)
	if !ok {
		return um
	}

	if sqlVal.IsNull() {
		return NewSQLNull(kind, um.EvalType())
	}

	if sqlVal.EvalType() == EvalBoolean {
		if sqlVal.Value().(bool) {
			return NewSQLInt64(kind, -1)
		}
		return NewSQLInt64(kind, 0)
	}

	return um
}

// nolint: unparam
func (um *SQLUnaryMinusExpr) reconcile() (SQLExpr, error) {
	child := um.expr
	typ := child.EvalType()
	if typ.IsNumeric() || typ == EvalNull {
		return um, nil
	}
	if typ == EvalString {
		return NewSQLUnaryMinusExpr(NewSQLConvertExpr(child, EvalDouble)), nil
	}
	return NewSQLUnaryMinusExpr(NewSQLConvertExpr(child, EvalInt64)), nil
}

func (um *SQLUnaryMinusExpr) String() string {
	return fmt.Sprintf("-%v", um.expr)
}

// ToAggregationLanguage translates SQLUnaryMinusExpr into something that can
// be used in an aggregation pipeline. If SQLUnaryMinusExpr cannot be translated,
// it will return nil and error.
func (um *SQLUnaryMinusExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	args, err := t.typedTranslateArgs([]SQLExpr{um.expr})
	if err != nil {
		return nil, err
	}

	assignments, args := minimizeLetAssignments([]string{"operand"}, args)
	evaluation := wrapInNullCheckedCond(
		t.ColumnsToNullCheck(),
		bsonutil.MgoNullLiteral,
		bsonutil.WrapInOp(bsonutil.OpMultiply, -1, args[0].v),
		args...,
	)

	return wrapInLet(assignments, evaluation), nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (um *SQLUnaryMinusExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return um.ToAggregationLanguage(t)
}

// EvalType returns the EvalType associated with SQLUnaryMinusExpr.
func (um *SQLUnaryMinusExpr) EvalType() EvalType {
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
func (td *SQLTildeExpr) EvalType() EvalType {
	return td.expr.EvalType()
}

// Evaluate evaluates a SQLTildeExpr into a SQLValue.
func (td *SQLTildeExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	expr, err := td.expr.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	if val, ok := expr.(SQLValue); ok {
		if val.IsNull() {
			return NewSQLNull(cfg.sqlValueKind, td.EvalType()), nil
		}
		return NewSQLUint64(cfg.sqlValueKind, ^uint64(Int64(val))), nil
	}

	return NewSQLUint64(cfg.sqlValueKind, ^uint64(0)), nil
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLTildeExpr) ExprName() string {
	return "SQLTildeExpr"
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLTildeExpr.
func (td *SQLTildeExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	if val, ok := td.expr.(SQLValue); ok {
		if val.IsNull() {
			return NewSQLNull(cfg.sqlValueKind, td.EvalType())
		}
		return NewSQLUint64(cfg.sqlValueKind, ^uint64(Int64(val)))
	}
	return td
}

// nolint: unparam
func (td *SQLTildeExpr) reconcile() (SQLExpr, error) {
	child := td.expr
	typ := child.EvalType()
	if typ.IsNumeric() || typ == EvalNull {
		return td, nil
	}
	if typ == EvalString {
		return NewSQLTildeExpr(NewSQLConvertExpr(child, EvalInt64)), nil
	}
	return NewSQLTildeExpr(NewSQLConvertExpr(child, EvalInt64)), nil
}

func (td *SQLTildeExpr) String() string {
	return fmt.Sprintf("~%v", td.expr)
}

// ToAggregationLanguage translates SQLTildeExpr into something that can
// be used in an aggregation pipeline. If SQLTildeExpr cannot be translated,
// it will return nil and error.
func (*SQLTildeExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newPushdownFailure(
		"SQLTildeExpr",
		"cannot push down to MongoDB",
	)
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (td *SQLTildeExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return td.ToAggregationLanguage(t)
}
