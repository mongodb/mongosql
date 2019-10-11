package evaluator

import (
	"context"
	"fmt"
	"math"

	"github.com/10gen/mongoast/ast"

	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/astutil"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

type sqlBinaryNode struct {
	left, right SQLExpr
}

// reconcileArithmetic is responsible for type reconciliation for all
// arithmetic operators ($add, $multiply, $subtract, $divide, $mod, etc.).
// The default conversion for an argument of type EvalDatetime is EvalDecimal128, while the
// default conversion for an argument of type EvalDate is EvalInt64. If a child is of
// numeric type, do nothing; otherwise, convert it to EvalDouble, unless it is a
// boolean, in which case it should be converted to EvalInt64.
func reconcileArithmetic(children []SQLExpr) []SQLExpr {
	convertedChildren := make([]SQLExpr, len(children))
	for i, child := range children {
		if child.EvalType().IsNumeric() {
			convertedChildren[i] = child
		} else {
			targetType := types.EvalDouble
			switch child.EvalType() {
			case types.EvalDatetime:
				targetType = types.EvalDecimal128
			case types.EvalDate, types.EvalBoolean:
				targetType = types.EvalInt64
			}
			convertedChildren[i] = NewSQLConvertExpr(child, targetType)
		}

	}

	return convertedChildren
}

// reconcileComparison behaves much as reconcileArithmetic does by checking to see if both arguments are numeric. If so, then there is no
// need for conversion. Otherwise, ReconcileSQLExprs is called to reconcile the two types based on the mySQL hierarchy.
func (bn sqlBinaryNode) reconcileComparison() sqlBinaryNode {
	left := bn.left
	right := bn.right

	leftType := left.EvalType()
	rightType := right.EvalType()

	// If both arguments are numeric, there is no need for
	// type reconciliation for comparison operators.
	if !(leftType.IsNumeric() && rightType.IsNumeric()) {
		left, right = ReconcileSQLExprs(left, right)
	}

	return sqlBinaryNode{left, right}
}

// Children returns a slice of all the Node children of the Node.
func (bn sqlBinaryNode) Children() []Node {
	return []Node{bn.left, bn.right}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (bn *sqlBinaryNode) ReplaceChild(i int, n Node) {
	switch i {
	case 0:
		bn.left = panicIfNotSQLExpr("sqlBinaryNode", n)
	case 1:
		bn.right = panicIfNotSQLExpr("sqlBinaryNode", n)
	default:
		panicWithInvalidIndex("sqlBinaryNode", i, 1)
	}
}

type valueArgsEnum int

const (
	noValueArgs valueArgsEnum = iota
	leftOnlyValueArg
	rightOnlyValueArg
	bothValueArgs
	allValueArgs
	someValueArgs
)

// sqlValueArgEnum returns the left and right values.SQLValue arguments, if any, and a enum that tells us
// which arguments are values.SQLValues.
func (bn *sqlBinaryNode) sqlValueArgEnum() (values.SQLValue, values.SQLValue, valueArgsEnum) {
	leftVal, leftIsVal := bn.left.(SQLValueExpr)
	rightVal, rightIsVal := bn.right.(SQLValueExpr)
	if leftIsVal && rightIsVal {
		return leftVal.Value, rightVal.Value, bothValueArgs
	}
	if leftIsVal {
		return leftVal.Value, nil, leftOnlyValueArg
	}
	if rightIsVal {
		return nil, rightVal.Value, rightOnlyValueArg
	}
	return nil, nil, noValueArgs
}

func (bn *sqlBinaryNode) evaluateArgs(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, values.SQLValue, error) {
	leftVal, err := bn.left.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, nil, err
	}

	rightVal, err := bn.right.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, nil, err
	}
	return leftVal, rightVal, nil
}

func (bn *sqlBinaryNode) toAggregationLanguageArgs(t *PushdownTranslator) (ast.Expr, ast.Expr, PushdownFailure) {

	left, err := t.ToAggregationLanguage(bn.left)
	if err != nil {
		return nil, nil, err
	}

	right, err := t.ToAggregationLanguage(bn.right)
	if err != nil {
		return nil, nil, err
	}
	return left, right, nil
}

// cmpOpToAggregationLanguage translates a binary comparison SQLExpr
// into something that can be used in an aggregation pipeline.
// This helper is specifically intended for use with =, <>, <, <=, >, and >=.
// If the expression cannot be translated, it will return nil and error.
func (bn *sqlBinaryNode) cmpOpToAggregationLanguage(t *PushdownTranslator, cmpOp string) (ast.Expr, PushdownFailure) {
	left, right, err := bn.toAggregationLanguageArgs(t)
	if err != nil {
		return nil, err
	}

	leftRef, rightRef := ast.NewVariableRef("left"), ast.NewVariableRef("right")

	assignments := []*ast.LetVariable{
		ast.NewLetVariable("left", left),
		ast.NewLetVariable("right", right),
	}

	comparison := ast.NewBinary(ast.BinaryOp(cmpOp), leftRef, rightRef)
	evaluation := astutil.WrapInNullCheckedCond(astutil.NullLiteral, comparison, leftRef, rightRef)

	return ast.NewLet(assignments, evaluation), nil
}

// cmpOpToAggregationPredicate translates a binary operation expression to the aggregation
// language to be used as a predicate in a $match stage via $expr. It is used by most comparison
// operators and translates the operation into a conjunctive expression that not only compares the
// left and right children, but also checks if they are null.
func (bn *sqlBinaryNode) cmpOpToAggregationPredicate(t *PushdownTranslator, cmpOp string) (ast.Expr, PushdownFailure) {
	left, right, err := bn.toAggregationLanguageArgs(t)
	if err != nil {
		return nil, err
	}

	return astutil.WrapInOp(bsonutil.OpAnd,
		astutil.WrapInOp(cmpOp, left, right),
		ast.NewBinary(bsonutil.OpGt, left, astutil.NullLiteral),
		ast.NewBinary(bsonutil.OpGt, right, astutil.NullLiteral)), nil
}

// SQLDivideExpr evaluates to the quotient of the left expression divided by the right.
type SQLDivideExpr struct{ sqlBinaryNode }

// NewSQLDivideExpr is a constructor for SQLDivideExpr.
func NewSQLDivideExpr(left, right SQLExpr) *SQLDivideExpr {
	return &SQLDivideExpr{sqlBinaryNode{left, right}}
}

// NewSQLModExpr is a constructor for SQLModExpr.
func NewSQLModExpr(left, right SQLExpr) *SQLModExpr {
	return &SQLModExpr{sqlBinaryNode{left, right}}
}

// NewSQLSubtractExpr is a constructor for SQLSubtractExpr.
func NewSQLSubtractExpr(left, right SQLExpr) *SQLSubtractExpr {
	return &SQLSubtractExpr{sqlBinaryNode{left, right}}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLDivideExpr) ExprName() string {
	return "SQLDivideExpr"
}

// Evaluate evaluates a SQLDivideExpr into a values.SQLValue.
func (div *SQLDivideExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(div)
	if err != nil {
		return nil, err
	}

	leftVal, rightVal, err := div.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if values.Float64(rightVal) == 0 || values.HasNullValue(leftVal, rightVal) {
		return values.NewSQLNull(cfg.sqlValueKind), nil
	}

	return doArithmetic(leftVal, rightVal, DIV)
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLDivideExpr.
func (div *SQLDivideExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(div); err != nil {
		return nil, err
	}

	leftVal, rightVal, valMask := div.sqlValueArgEnum()
	switch valMask {
	case noValueArgs:
	case leftOnlyValueArg:
		if leftVal.IsNull() {
			return NewSQLValueExpr(leftVal), nil
		}
	case rightOnlyValueArg:
		if rightVal.IsNull() {
			return NewSQLValueExpr(rightVal), nil
		}
		frightVal := values.Float64(rightVal)
		if frightVal == 0.0 {
			return NewSQLValueExpr(values.NewSQLNull(rightVal.Kind())), nil
		}
		if frightVal == 1.0 {
			return div.left, nil
		}
	case bothValueArgs:
		if leftVal.IsNull() || rightVal.IsNull() {
			return NewSQLValueExpr(values.NewSQLNull(rightVal.Kind())), nil
		}
		if out, err := doArithmetic(leftVal, rightVal, DIV); err == nil {
			return NewSQLValueExpr(out), nil
		}
	}
	return div, nil
}

// nolint: unparam
func (div *SQLDivideExpr) reconcile() (SQLExpr, error) {
	children := reconcileArithmetic([]SQLExpr{div.left, div.right})
	node := sqlBinaryNode{children[0], children[1]}
	return &SQLDivideExpr{node}, nil
}

func (div *SQLDivideExpr) String() string {
	return fmt.Sprintf("%v/%v", div.left, div.right)
}

// ToAggregationLanguage translates SQLDivideExpr into something that can
// be used in an aggregation pipeline. If SQLDivideExpr cannot be translated,
// it will return nil and error.
func (div *SQLDivideExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	left, right, err := div.toAggregationLanguageArgs(t)
	if err != nil {
		return nil, err
	}

	leftRef, rightRef := ast.NewVariableRef("left"), ast.NewVariableRef("right")

	assignments := []*ast.LetVariable{
		ast.NewLetVariable("left", left),
		ast.NewLetVariable("right", right),
	}

	evaluation := astutil.WrapInCond(
		astutil.NullLiteral,
		ast.NewBinary(bsonutil.OpDivide, leftRef, rightRef),
		ast.NewBinary(bsonutil.OpEq, rightRef, astutil.ZeroInt32Literal),
	)

	return ast.NewLet(assignments, evaluation), nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (div *SQLDivideExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return div.ToAggregationLanguage(t)
}

// EvalType returns the EvalType associated with SQLDivideExpr.
func (div *SQLDivideExpr) EvalType() types.EvalType {
	// the server returns a decimal if, and only if, either side is decimal.
	if div.left.EvalType() == types.EvalDecimal128 || div.right.EvalType() == types.EvalDecimal128 {
		return types.EvalDecimal128
	}
	return types.EvalDouble
}

// SQLIDivideExpr evaluates the integer quotient of the left expression divided by the right.
type SQLIDivideExpr struct{ sqlBinaryNode }

// NewSQLIDivideExpr is a constructor for SQLIDivideExpr.
func NewSQLIDivideExpr(left, right SQLExpr) *SQLIDivideExpr {
	return &SQLIDivideExpr{sqlBinaryNode{left, right}}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLIDivideExpr) ExprName() string {
	return "SQLIDivideExpr"
}

func idivide(knd values.SQLValueKind, leftVal, rightVal values.SQLValue) values.SQLValue {
	dividend := values.Float64(leftVal)
	divisor := values.Float64(rightVal)

	if divisor == 0 || values.HasNullValue(leftVal, rightVal) {
		// NOTE: this is per mysql manual
		return values.NewSQLNull(knd)
	}

	return values.NewSQLInt64(knd, int64(dividend/divisor))
}

// Evaluate evaluates a SQLIDivideExpr into a values.SQLValue.
func (div *SQLIDivideExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(div)
	if err != nil {
		return nil, err
	}

	leftVal, rightVal, err := div.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	return idivide(cfg.sqlValueKind, leftVal, rightVal), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLIDivideExpr.
func (div *SQLIDivideExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(div); err != nil {
		return nil, err
	}

	leftVal, rightVal, valMask := div.sqlValueArgEnum()
	switch valMask {
	case noValueArgs:
	case leftOnlyValueArg:
		if leftVal.IsNull() {
			return NewSQLValueExpr(leftVal), nil
		}
	case rightOnlyValueArg:
		if rightVal.IsNull() {
			return NewSQLValueExpr(rightVal), nil
		}
		rightFloatVal := values.Float64(rightVal)
		if rightFloatVal == 0 {
			return NewSQLValueExpr(values.NewSQLNull(rightVal.Kind())), nil
		}
		if rightFloatVal == 1 {
			return NewSQLScalarFunctionExpr("floor", []SQLExpr{div.left})
		}
	case bothValueArgs:
		return NewSQLValueExpr(idivide(cfg.sqlValueKind, leftVal, rightVal)), nil
	}
	return div, nil
}

// nolint: unparam
func (div *SQLIDivideExpr) reconcile() (SQLExpr, error) {
	children := reconcileArithmetic([]SQLExpr{div.left, div.right})
	node := sqlBinaryNode{children[0], children[1]}
	return &SQLIDivideExpr{node}, nil
}

func (div *SQLIDivideExpr) String() string {
	return fmt.Sprintf("%v/%v", div.left, div.right)
}

// ToAggregationLanguage translates SQLIDivideExpr into something that can
// be used in an aggregation pipeline. If SQLIDivideExpr cannot be translated,
// it will return nil and error.
func (div *SQLIDivideExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	left, right, err := div.toAggregationLanguageArgs(t)
	if err != nil {
		return nil, err
	}

	leftRef, rightRef := ast.NewVariableRef("left"), ast.NewVariableRef("right")

	assignments := []*ast.LetVariable{
		ast.NewLetVariable("left", left),
		ast.NewLetVariable("right", right),
	}

	evaluation := astutil.WrapInCond(
		astutil.NullLiteral,
		astutil.WrapInOp(bsonutil.OpTrunc, ast.NewBinary(bsonutil.OpDivide, leftRef, rightRef)),
		ast.NewBinary(bsonutil.OpEq, rightRef, astutil.ZeroInt32Literal),
	)

	return ast.NewLet(assignments, evaluation), nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (div *SQLIDivideExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return div.ToAggregationLanguage(t)
}

// EvalType returns the EvalType associated with SQLIDivideExpr.
func (div *SQLIDivideExpr) EvalType() types.EvalType {
	return types.EvalInt64
}

// SQLIsExpr evaluates to true if the left is equal to the boolean value on the right.
type SQLIsExpr struct{ sqlBinaryNode }

// NewSQLIsExpr is a constructor for SQLIsExpr.
func NewSQLIsExpr(left, right SQLExpr) *SQLIsExpr {
	return &SQLIsExpr{sqlBinaryNode{left, right}}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLIsExpr) ExprName() string {
	return "SQLIsExpr"
}

var _ translatableToMatch = (*SQLIsExpr)(nil)

// Evaluate evaluates a SQLIsExpr into a values.SQLValue.
func (is *SQLIsExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(is)
	if err != nil {
		return nil, err
	}

	leftVal, rightVal, err := is.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if leftVal.IsNull() {
		if rightVal.IsNull() {
			return values.NewSQLBool(cfg.sqlValueKind, true), nil
		}
		return values.NewSQLBool(cfg.sqlValueKind, false), nil
	}

	if values.HasNullValue(leftVal, rightVal) {
		return values.NewSQLBool(cfg.sqlValueKind, false), nil
	}

	return values.NewSQLBool(cfg.sqlValueKind, values.Bool(leftVal) == values.Bool(rightVal)), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLIsExpr.
func (is *SQLIsExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(is); err != nil {
		return nil, err
	}

	leftVal, rightVal, valMask := is.sqlValueArgEnum()
	switch valMask {
	case noValueArgs, leftOnlyValueArg:
		panic("the right argument to SQLIsExpr should always be a values.SQLValue")
	case rightOnlyValueArg:
	case bothValueArgs:
		if leftVal.IsNull() {
			if rightVal.IsNull() {
				return NewSQLValueExpr(values.NewSQLBool(cfg.sqlValueKind, true)), nil
			}
			return NewSQLValueExpr(values.NewSQLBool(cfg.sqlValueKind, false)), nil
		}
		if rightVal.IsNull() {
			return NewSQLValueExpr(values.NewSQLBool(cfg.sqlValueKind, false)), nil
		}

		return NewSQLValueExpr(values.NewSQLBool(cfg.sqlValueKind, values.Bool(leftVal) == values.Bool(rightVal))), nil
	}
	if shouldFlip(is.sqlBinaryNode) {
		left, right := is.right, is.left
		is.left, is.right = left, right
	}
	return is, nil
}

// nolint: unparam
func (is *SQLIsExpr) reconcile() (SQLExpr, error) {
	if is.right.EvalType() == types.EvalBoolean {
		leftType := is.left.EvalType()
		if !(leftType.IsNumeric() || leftType == types.EvalBoolean) {
			reconciled := convertAllExprs([]SQLExpr{is.left, is.right}, types.EvalBoolean)
			return NewSQLIsExpr(reconciled[0], reconciled[1]), nil
		}
	}
	return is, nil
}

func (is *SQLIsExpr) String() string {
	return fmt.Sprintf("%v is %v", is.left, is.right)
}

// ToAggregationLanguage translates SQLIsExpr into something that can
// be used in an aggregation pipeline. If SQLIsExpr cannot be translated,
// it will return nil and error.
func (is *SQLIsExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	left, right, err := is.toAggregationLanguageArgs(t)
	if err != nil {
		return nil, err
	}

	// if right side is {null,unknown}, it's a simple case
	sqlVal, ok := is.right.(SQLValueExpr)
	if ok && sqlVal.Value.IsNull() {
		return astutil.WrapInOp(bsonutil.OpLte,
			left,
			astutil.NullLiteral,
		), nil
	}

	// if left side is a boolean, this is still simple
	if is.left.EvalType() == types.EvalBoolean {
		return astutil.WrapInOp(bsonutil.OpEq,
			left,
			right,
		), nil
	}

	// otherwise, left side is a number type
	if ok && sqlVal.Value == values.NewSQLBool(t.valueKind(), true) {
		return astutil.WrapInOp(bsonutil.OpNeq,
			astutil.WrapInIfNull(left, astutil.ZeroInt32Literal),
			astutil.ZeroInt32Literal,
		), nil
	} else if ok && sqlVal.Value == values.NewSQLBool(t.valueKind(), false) {
		return astutil.WrapInOp(bsonutil.OpEq,
			left,
			astutil.ZeroInt32Literal,
		), nil
	}

	return nil, newPushdownFailure(
		is.ExprName(),
		"not one of the enumerated translatable forms",
		"expr", is.String(),
	)
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (is *SQLIsExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return is.ToAggregationLanguage(t)
}

// ToMatchLanguage translates SQLIsExpr into something that can
// be used in an match expression. If SQLIsExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLIsExpr.
func (is *SQLIsExpr) ToMatchLanguage(t *PushdownTranslator) (ast.Expr, SQLExpr) {
	ref, ok := t.getFieldRef(is.left)
	if !ok {
		return nil, is
	}

	rightVal, ok := is.right.(SQLValueExpr)
	if !ok {
		return nil, is
	}

	if rightVal.Value.IsNull() {
		return ast.NewBinary(bsonutil.OpEq, ref, astutil.NullLiteral), nil
	}

	rightBool, ok := rightVal.Value.(values.SQLBool)
	if !ok {
		return nil, is
	}

	if rightBool.Value().(bool) {
		if is.left.EvalType() == types.EvalBoolean {
			return ast.NewBinary(bsonutil.OpEq, ref, astutil.TrueLiteral), nil
		}

		return astutil.WrapInOp(bsonutil.OpAnd,
			ast.NewBinary(bsonutil.OpNeq, ref, astutil.ZeroInt32Literal),
			ast.NewBinary(bsonutil.OpNeq, ref, astutil.NullLiteral),
		), nil
	}

	if is.left.EvalType() == types.EvalBoolean {
		return ast.NewBinary(bsonutil.OpEq, ref, astutil.FalseLiteral), nil
	}
	return ast.NewBinary(bsonutil.OpEq, ref, astutil.ZeroInt32Literal), nil
}

// EvalType returns the EvalType associated with SQLIsExpr.
func (*SQLIsExpr) EvalType() types.EvalType {
	return types.EvalBoolean
}

// ComparisonOp is the type of the constants used in SQLComparisonExpr.
type ComparisonOp byte

// These constants are used in SQLComparisonExpr to identify the comparison
// operation being used.
const (
	EQ ComparisonOp = iota
	NEQ
	LT
	LTE
)

// SQLComparisonExpr evaluates 4 different kinds of comparison expressions:
// <, <=, =, and !=. The > and >= operators are desugared in the parser and
// rewritten as > and >=, respectively.
type SQLComparisonExpr struct {
	sqlBinaryNode
	op ComparisonOp
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLComparisonExpr) ExprName() string {
	return "SQLComparisonExpr"
}

var _ translatableToMatch = (*SQLComparisonExpr)(nil)

// NewSQLComparisonExpr is a constructor for SQLComparisonExpr.
func NewSQLComparisonExpr(op ComparisonOp, left, right SQLExpr) *SQLComparisonExpr {
	return &SQLComparisonExpr{
		sqlBinaryNode{
			left:  left,
			right: right,
		}, op}
}

// convertComparisonOpToBool converts the value of the comparison c into a bool and then
// wraps that in a SQLValue.
func convertComparisonOpToBool(op ComparisonOp, c int, kind values.SQLValueKind) values.SQLValue {
	switch op {
	case LT:
		return values.NewSQLBool(kind, c < 0)
	case LTE:
		return values.NewSQLBool(kind, c <= 0)
	case EQ:
		return values.NewSQLBool(kind, c == 0)
	case NEQ:
		return values.NewSQLBool(kind, c != 0)
	}
	panic(fmt.Sprintf("unsupported operation in SQLComparisonExpr: %v\n", op))
}

// Evaluate evaluates a SQLComparisonExpr into a values.SQLValue.
func (comp *SQLComparisonExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(comp)
	if err != nil {
		return nil, err
	}

	leftVal, rightVal, err := comp.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if values.HasNullValue(leftVal, rightVal) {
		return values.NewSQLNull(cfg.sqlValueKind), nil
	}

	c, err := values.CompareTo(leftVal, rightVal, st.collation)
	if err == nil {
		return convertComparisonOpToBool(comp.op, c, cfg.sqlValueKind), nil
	}

	return values.NewSQLBool(cfg.sqlValueKind, false), err
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLComparisonExpr.
func (comp *SQLComparisonExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(comp); err != nil {
		return nil, err
	}

	leftVal, rightVal, valMask := comp.sqlValueArgEnum()
	switch valMask {
	case noValueArgs:
	case leftOnlyValueArg:
		if leftVal.IsNull() {
			return NewSQLValueExpr(leftVal), nil
		}
	case rightOnlyValueArg:
		if rightVal.IsNull() {
			return NewSQLValueExpr(rightVal), nil
		}
	case bothValueArgs:
		if leftVal.IsNull() || rightVal.IsNull() {
			return NewSQLValueExpr(values.NewSQLNull(rightVal.Kind())), nil
		}
		c, err := values.CompareTo(leftVal, rightVal, cfg.collation)
		if err == nil {
			return NewSQLValueExpr(convertComparisonOpToBool(comp.op, c, cfg.sqlValueKind)), nil
		}
	}

	return comp, nil
}

// nolint: unparam
func (comp *SQLComparisonExpr) reconcile() (SQLExpr, error) {
	switch comp.op {
	case EQ:
		var reconciled bool

		left := comp.left
		right := comp.right

		if isBooleanColumnAndNumber(left, right) || isBooleanColumnAndNumber(right, left) {
			var col SQLColumnExpr
			var lit values.SQLNumber

			switch left.EvalType() {
			case types.EvalBoolean:
				col = left.(SQLColumnExpr)
				lit = right.(SQLValueExpr).Value.(values.SQLNumber)
			default:
				col = right.(SQLColumnExpr)
				lit = left.(SQLValueExpr).Value.(values.SQLNumber)
			}

			if ilit := values.Int64(lit); ilit == 1 || ilit == 0 {
				left = col
				right = NewSQLConvertExpr(NewSQLValueExpr(lit), types.EvalBoolean)
				reconciled = true
			}
		}

		if !reconciled {
			return &SQLComparisonExpr{comp.reconcileComparison(), EQ}, nil
		}

		return NewSQLComparisonExpr(EQ, left, right), nil
	default:
		return &SQLComparisonExpr{comp.reconcileComparison(), comp.op}, nil
	}
}

func (comp *SQLComparisonExpr) String() string {
	switch comp.op {
	case LT:
		return fmt.Sprintf("%v < %v", comp.left, comp.right)
	case LTE:
		return fmt.Sprintf("%v <= %v", comp.left, comp.right)
	case EQ:
		return fmt.Sprintf("%v = %v", comp.left, comp.right)
	case NEQ:
		return fmt.Sprintf("%v != %v", comp.left, comp.right)
	}

	panic(fmt.Sprintf("unsupported operation in SQLComparisonExpr: %v\n", comp.op))
}

// flipAndGetOpName checks if a SQLComparisonExpr's children should be flipped,
// flips them if necessary, and returns the appropriate comparison operation
// name. Flipping is only ever necessary in toMatchLanguage.
// If the operator is LT or LTE, we do not physically swap the left and right
// children in the tree; we instead return swapped copies.
func (comp *SQLComparisonExpr) flipAndGetOpName(isMatchLanguage bool) (string, SQLExpr, SQLExpr) {

	flipped := false
	var left, right SQLExpr
	if shouldFlip(comp.sqlBinaryNode) && isMatchLanguage {
		left, right = comp.right, comp.left
		flipped = true
	}

	switch comp.op {
	case LT:
		if flipped {
			return bsonutil.OpGt, left, right
		}
		return bsonutil.OpLt, comp.left, comp.right
	case LTE:
		if flipped {
			return bsonutil.OpGte, left, right
		}
		return bsonutil.OpLte, comp.left, comp.right
	case EQ:
		if flipped {
			comp.left, comp.right = left, right
		}
		return bsonutil.OpEq, comp.left, comp.right
	case NEQ:
		if flipped {
			comp.left, comp.right = left, right
		}
		return bsonutil.OpNeq, comp.left, comp.right
	default:
		panic(fmt.Sprintf("unsupported operation in SQLComparisonExpr: %v\n", comp.op))
	}
}

// ToAggregationLanguage translates SQLComparisonExpr into something that can
// be used in an aggregation pipeline. If SQLComparisonExpr cannot be translated,
// it will return nil and error.
func (comp *SQLComparisonExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	opName, _, _ := comp.flipAndGetOpName(false)
	return comp.cmpOpToAggregationLanguage(t, opName)
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (comp *SQLComparisonExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	opName, _, _ := comp.flipAndGetOpName(false)
	return comp.cmpOpToAggregationPredicate(t, opName)
}

// ToMatchLanguage translates SQLComparisonExpr into something that can
// be used in an match expression. If SQLComparisonExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLComparisonExpr.
func (comp *SQLComparisonExpr) ToMatchLanguage(t *PushdownTranslator) (ast.Expr, SQLExpr) {
	var ok bool
	var match *ast.Binary

	switch comp.op {
	case LT, LTE, EQ:
		opName, left, right := comp.flipAndGetOpName(true)
		match, ok = t.translateOperator(opName, left, right)
		if !ok {
			return nil, comp
		}
		return match, nil
	case NEQ:
		var match ast.Expr
		opName, _, _ := comp.flipAndGetOpName(true)
		match, ok = t.translateOperator(opName, comp.left, comp.right)
		if !ok {
			return nil, comp
		}

		value, err := t.getValue(comp.right)
		if err != nil {
			return nil, comp
		}

		if value.Value.Type != bsontype.Null {
			ref, ok := t.getFieldRef(comp.left)
			if !ok {
				return nil, comp
			}
			match = astutil.WrapInOp(bsonutil.OpAnd,
				match,
				ast.NewBinary(bsonutil.OpNeq, ref, astutil.NullLiteral),
			)
		}

		return match, nil

	}

	panic(fmt.Sprintf("unsupported operation in SQLComparisonExpr: %v\n", comp.op))
}

// EvalType returns the EvalType associated with SQLComparisonExpr.
func (*SQLComparisonExpr) EvalType() types.EvalType {
	return types.EvalBoolean
}

// SQLModExpr evaluates the modulus of two expressions
type SQLModExpr struct{ sqlBinaryNode }

// ExprName returns a string representing this SQLExpr's name.
func (*SQLModExpr) ExprName() string {
	return "SQLModExpr"
}

// Evaluate evaluates a SQLModExpr into a values.SQLValue.
func (mod *SQLModExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(mod)
	if err != nil {
		return nil, err
	}

	leftVal, rightVal, err := mod.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	frightVal := values.Float64(rightVal)
	if math.Abs(frightVal) == 0.0 || values.HasNullValue(leftVal, rightVal) {
		return values.NewSQLNull(cfg.sqlValueKind), nil
	}

	modVal := math.Mod(values.Float64(leftVal), frightVal)
	if modVal == -0 {
		modVal *= -1
	}

	return values.NewSQLFloat(cfg.sqlValueKind, modVal), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLModExpr.
func (mod *SQLModExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(mod); err != nil {
		return nil, err
	}

	leftVal, rightVal, valMask := mod.sqlValueArgEnum()
	switch valMask {
	case noValueArgs:
	case leftOnlyValueArg:
		if leftVal.IsNull() {
			return NewSQLValueExpr(leftVal), nil
		}
	case rightOnlyValueArg:
		if rightVal.IsNull() {
			return NewSQLValueExpr(rightVal), nil
		}
		frightVal := values.Float64(rightVal)
		if frightVal == 0.0 {
			return NewSQLValueExpr(values.NewSQLNull(rightVal.Kind())), nil
		}
		if frightVal == 1.0 {
			return mod.left, nil
		}
	case bothValueArgs:
		if leftVal.IsNull() || rightVal.IsNull() {
			return NewSQLValueExpr(values.NewSQLNull(rightVal.Kind())), nil
		}
		frightVal := values.Float64(rightVal)
		if math.Abs(frightVal) == 0.0 || values.HasNullValue(leftVal, rightVal) {
			return NewSQLValueExpr(values.NewSQLNull(cfg.sqlValueKind)), nil
		}
		modVal := math.Mod(values.Float64(leftVal), frightVal)
		if modVal == -0 {
			modVal *= -1
		}
		return NewSQLValueExpr(values.NewSQLFloat(cfg.sqlValueKind, modVal)), nil
	}
	return mod, nil
}

// nolint: unparam
func (mod *SQLModExpr) reconcile() (SQLExpr, error) {
	children := reconcileArithmetic([]SQLExpr{mod.left, mod.right})
	node := sqlBinaryNode{children[0], children[1]}
	return &SQLModExpr{node}, nil
}

func (mod *SQLModExpr) String() string {
	return fmt.Sprintf("%v/%v", mod.left, mod.right)
}

// ToAggregationLanguage translates SQLModExpr into something that can
// be used in an aggregation pipeline. If SQLModExpr cannot be translated,
// it will return nil and error.
func (mod *SQLModExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	left, right, err := mod.toAggregationLanguageArgs(t)
	if err != nil {
		return nil, err
	}

	return ast.NewBinary(bsonutil.OpMod, left, right), nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (mod *SQLModExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return mod.ToAggregationLanguage(t)
}

// EvalType returns the EvalType associated with SQLModExpr.
func (mod *SQLModExpr) EvalType() types.EvalType {
	return preferentialType(mod.left, mod.right)
}

// SQLNullSafeEqualsExpr behaves like the = operator,
// but returns 1 rather than NULL if both operands are
// NULL, and 0 rather than NULL if one operand is NULL.
type SQLNullSafeEqualsExpr struct{ sqlBinaryNode }

// ExprName returns a string representing this SQLExpr's name.
func (*SQLNullSafeEqualsExpr) ExprName() string {
	return "SQLNullSafeEqualsExpr"
}

// NewSQLNullSafeEqualsExpr is a constructor for SQLNullSafeEqualsExpr.
func NewSQLNullSafeEqualsExpr(left, right SQLExpr) *SQLNullSafeEqualsExpr {
	return &SQLNullSafeEqualsExpr{
		sqlBinaryNode{
			left:  left,
			right: right,
		}}
}

// Evaluate evaluates a SQLNullSafeEqualsExpr into a values.SQLValue.
func (nse *SQLNullSafeEqualsExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(nse)
	if err != nil {
		return nil, err
	}

	leftVal, rightVal, err := nse.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	c, err := values.CompareTo(leftVal, rightVal, st.collation)
	if err == nil {
		return values.NewSQLBool(cfg.sqlValueKind, c == 0), nil
	}

	return values.NewSQLBool(cfg.sqlValueKind, false), err
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLNullSafeEqualsExpr.
func (nse *SQLNullSafeEqualsExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(nse); err != nil {
		return nil, err
	}

	leftVal, rightVal, valMask := nse.sqlValueArgEnum()
	switch valMask {
	// Because constant NULLs do not cause <=> to evaluate to NULL, there is
	// no room for ConstantFolding unless BOTH sides are constants.
	case noValueArgs, leftOnlyValueArg, rightOnlyValueArg:
	case bothValueArgs:
		c, err := values.CompareTo(leftVal, rightVal, cfg.collation)
		if err == nil {
			return NewSQLValueExpr(values.NewSQLBool(cfg.sqlValueKind, c == 0)), nil
		}
	}
	if shouldFlip(nse.sqlBinaryNode) {
		left, right := nse.right, nse.left
		nse.left, nse.right = left, right
	}
	return nse, nil
}

// nolint: unparam
func (nse *SQLNullSafeEqualsExpr) reconcile() (SQLExpr, error) {
	return &SQLNullSafeEqualsExpr{nse.reconcileComparison()}, nil
}

func (nse *SQLNullSafeEqualsExpr) String() string {
	return fmt.Sprintf("%v <=> %v", nse.left, nse.right)
}

// ToAggregationLanguage translates SQLNullSafeEqualsExpr into something that can
// be used in an aggregation pipeline. If SQLNullSafeEqualsExpr cannot be translated,
// it will return nil and error.
func (nse *SQLNullSafeEqualsExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	left, right, err := nse.toAggregationLanguageArgs(t)
	if err != nil {
		return nil, err
	}

	return ast.NewBinary(bsonutil.OpEq,
		astutil.WrapInIfNull(left, astutil.NullLiteral),
		astutil.WrapInIfNull(right, astutil.NullLiteral),
	), nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (nse *SQLNullSafeEqualsExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return nse.ToAggregationLanguage(t)
}

// EvalType returns the EvalType associated with SQLNullSafeEqualsExpr.
func (*SQLNullSafeEqualsExpr) EvalType() types.EvalType {
	return types.EvalBoolean
}

// SQLSubtractExpr evaluates to the difference of the left expression minus the right expressions.
type SQLSubtractExpr struct{ sqlBinaryNode }

// ExprName returns a string representing this SQLExpr's name.
func (*SQLSubtractExpr) ExprName() string {
	return "SQLSubtractExpr"
}

// Evaluate evaluates a SQLSubtractExpr into a values.SQLValue.
func (sub *SQLSubtractExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(sub)
	if err != nil {
		return nil, err
	}

	leftVal, rightVal, err := sub.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if values.HasNullValue(leftVal, rightVal) {
		return values.NewSQLNull(cfg.sqlValueKind), nil
	}

	return doArithmetic(leftVal, rightVal, SUB)
}

// EvalType returns the EvalType associated with SQLSubtractExpr.
func (sub *SQLSubtractExpr) EvalType() types.EvalType {
	// the server determines the result's type based on the input types.
	return arithmeticEvalType(sub.left, sub.right)
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLSubtractExpr.
func (sub *SQLSubtractExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(sub); err != nil {
		return nil, err
	}

	leftVal, rightVal, valMask := sub.sqlValueArgEnum()
	switch valMask {
	case noValueArgs:
	case leftOnlyValueArg:
		if leftVal.IsNull() {
			return NewSQLValueExpr(leftVal), nil
		}
		if values.IsZero(leftVal) {
			return sub.right, nil
		}
	case rightOnlyValueArg:
		if rightVal.IsNull() {
			return NewSQLValueExpr(rightVal), nil
		}
		if values.IsZero(rightVal) {
			return sub.left, nil
		}
	case bothValueArgs:
		if leftVal.IsNull() || rightVal.IsNull() {
			return NewSQLValueExpr(values.NewSQLNull(rightVal.Kind())), nil
		}
		if out, err := doArithmetic(leftVal, rightVal, SUB); err == nil {
			return NewSQLValueExpr(out), nil
		}
	}
	return sub, nil
}

func (sub *SQLSubtractExpr) String() string {
	return fmt.Sprintf("%v-%v", sub.left, sub.right)
}

// nolint: unparam
func (sub *SQLSubtractExpr) reconcile() (SQLExpr, error) {
	children := reconcileArithmetic([]SQLExpr{sub.left, sub.right})
	node := sqlBinaryNode{children[0], children[1]}
	return &SQLSubtractExpr{node}, nil
}

// ToAggregationLanguage translates SQLSubtractExpr into something that can
// be used in an aggregation pipeline. If SQLSubtractExpr cannot be translated,
// it will return nil and error.
func (sub *SQLSubtractExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	left, right, err := sub.toAggregationLanguageArgs(t)
	if err != nil {
		return nil, err
	}

	return ast.NewBinary(bsonutil.OpSubtract, left, right), nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (sub *SQLSubtractExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return sub.ToAggregationLanguage(t)
}
