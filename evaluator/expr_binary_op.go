package evaluator

import (
	"context"
	"fmt"
	"math"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/shopspring/decimal"
)

type sqlBinaryNode struct {
	left, right SQLExpr
}

func (bn sqlBinaryNode) reconcileArithmetic() (sqlBinaryNode, error) {
	var err error
	kind := MongoSQLValueKind

	left := bn.left
	right := bn.right

	if left.EvalType() == EvalDatetime || right.EvalType() == EvalDatetime {
		// Arithmetic with Timestamps should treat them as floating points due to fractional seconds.

		left, _, err = ReconcileSQLExprs(left, NewSQLDecimal128(kind, decimal.NewFromFloat(0.0)))
		if err != nil {
			return bn, err
		}
		_, right, err = ReconcileSQLExprs(NewSQLDecimal128(kind, decimal.NewFromFloat(0.0)), right)
		if err != nil {
			return bn, err
		}

	} else if left.EvalType() == EvalDate || right.EvalType() == EvalDate {
		// Arithmetic with Dates should treat them as integers.

		left, _, err = ReconcileSQLExprs(left, NewSQLInt64(kind, 0))
		if err != nil {
			return bn, err
		}
		_, right, err = ReconcileSQLExprs(NewSQLInt64(kind, 0), right)
		if err != nil {
			return bn, err
		}

	} else {
		// otherwise, reconcile left and right side with each other
		left, right, err = ReconcileSQLExprs(left, right)
		if err != nil {
			return bn, err
		}
	}

	// now convert them all to doubles
	reconciled := convertAllExprs([]SQLExpr{left, right}, EvalDouble)

	return sqlBinaryNode{reconciled[0], reconciled[1]}, nil
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
	noValueArgs       valueArgsEnum = 0
	leftOnlyValueArg  valueArgsEnum = iota
	rightOnlyValueArg valueArgsEnum = iota
	bothValueArgs     valueArgsEnum = iota
)

// sqlValueArgEnum returns the left and right SQLValue arguments, if any, and a enum that tells us
// which arguments are SQLValues.
func (bn *sqlBinaryNode) sqlValueArgEnum() (SQLValue, SQLValue, valueArgsEnum) {
	leftVal, leftIsVal := bn.left.(SQLValue)
	rightVal, rightIsVal := bn.right.(SQLValue)
	if leftIsVal && rightIsVal {
		return leftVal, rightVal, bothValueArgs
	}
	if leftIsVal {
		return leftVal, nil, leftOnlyValueArg
	}
	if rightIsVal {
		return nil, rightVal, rightOnlyValueArg
	}
	return nil, nil, noValueArgs
}

func (bn *sqlBinaryNode) evaluateArgs(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, SQLValue, error) {
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

func (bn *sqlBinaryNode) toAggregationLanguageArgs(t *PushdownTranslator) (interface{}, interface{}, PushdownFailure) {

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

// eatChildren attempts to recursively consume the left and right children of the
// current node. Consumption consists of removal of the node and adoption of its
// children. Consumption of a node will succeed only if the consumer and consumed
// are of the same type. The result of the operation is children for the current
// node to adopt.
// Invoking eatChildren with more than 2 SQLExprs will cause a panic.
func eatChildren(opName string, leftAndRight []SQLExpr) []SQLExpr {
	if len(leftAndRight) != 2 {
		panic("eatChildren called with more than 2 children")
	}

	children := make([]SQLExpr, 0)

	for _, c := range leftAndRight {
		switch t := c.(type) {
		// The only operators supported for eating children are Add, And, Multiply, Or, and Xor.
		case *SQLAddExpr, *SQLAndExpr, *SQLMultiplyExpr, *SQLOrExpr, *SQLXorExpr:
			if c.ExprName() == opName {
				// if the child c is one of the same type as the parent (the opName
				// argument), recursively consume its children.
				children = append(children, eatChildren(opName, t.Children())...)
				continue
			}
		}

		// if that is not the case, just include c in the list of children
		children = append(children, c)
	}

	return children
}

// SQLAddExpr evaluates to the sum of two expressions.
type SQLAddExpr struct{ sqlBinaryNode }

// NewSQLAddExpr is a constructor for SQLAddExpr.
func NewSQLAddExpr(left, right SQLExpr) *SQLAddExpr {
	return &SQLAddExpr{sqlBinaryNode{left, right}}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLAddExpr) ExprName() string {
	return "SQLAddExpr"
}

// Evaluate evaluates a SQLAddExpr into a SQLValue.
func (add *SQLAddExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	leftVal, rightVal, err := add.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(cfg.sqlValueKind, add.EvalType()), nil
	}

	return doArithmetic(leftVal, rightVal, ADD)
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLAddExpr.
func (add *SQLAddExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	leftVal, rightVal, valMask := add.sqlValueArgEnum()
	switch valMask {
	case noValueArgs:
	case leftOnlyValueArg:
		if leftVal.IsNull() {
			return leftVal
		}
		if isZero(leftVal) {
			return add.right
		}
	case rightOnlyValueArg:
		if rightVal.IsNull() {
			return rightVal
		}
		if isZero(rightVal) {
			return add.left
		}
	case bothValueArgs:
		if leftVal.IsNull() || rightVal.IsNull() {
			return NewSQLNull(rightVal.Kind(), rightVal.EvalType())
		}
		if out, err := doArithmetic(leftVal, rightVal, ADD); err == nil {
			return out
		}
	}
	return add
}

func (add *SQLAddExpr) reconcile() (SQLExpr, error) {
	bin, err := add.reconcileArithmetic()
	if err != nil {
		return nil, err
	}
	newExpr := SQLAddExpr{bin}
	return &newExpr, nil
}

func (add *SQLAddExpr) String() string {
	return fmt.Sprintf("%v+%v", add.left, add.right)
}

// ToAggregationLanguage translates SQLAddExpr into something that can
// be used in an aggregation pipeline. If SQLAddExpr cannot be translated,
// it will return nil and error.
func (add *SQLAddExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	children := eatChildren(add.ExprName(), add.Children())
	ops := make([]interface{}, len(children))

	var err PushdownFailure
	for i, c := range children {
		if ops[i], err = t.ToAggregationLanguage(c); err != nil {
			return nil, err
		}
	}

	return bsonutil.WrapInOp(bsonutil.OpAdd, ops...), nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (add *SQLAddExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return add.ToAggregationLanguage(t)
}

// EvalType returns the EvalType associated with SQLAddExpr.
func (add *SQLAddExpr) EvalType() EvalType {
	return EvalDouble
}

// SQLAndExpr evaluates to true if and only if all its children evaluate to true.
type SQLAndExpr struct{ sqlBinaryNode }

// ExprName returns a string representing this SQLExpr's name.
func (*SQLAndExpr) ExprName() string {
	return "SQLAndExpr"
}

var _ translatableToMatch = (*SQLAndExpr)(nil)

// NewSQLAndExpr is a constructor for SQLAndExpr.
func NewSQLAndExpr(left, right SQLExpr) *SQLAndExpr {
	return &SQLAndExpr{
		sqlBinaryNode{
			left:  left,
			right: right,
		}}
}

// Evaluate evaluates a SQLAndExpr into a SQLValue.
func (and *SQLAndExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	leftVal, rightVal, err := and.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if IsFalsy(leftVal) || IsFalsy(rightVal) {
		return NewSQLBool(cfg.sqlValueKind, false), nil
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(cfg.sqlValueKind, and.EvalType()), nil
	}

	return NewSQLBool(cfg.sqlValueKind, true), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLAndExpr.
func (and *SQLAndExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	leftVal, rightVal, valMask := and.sqlValueArgEnum()

	switch valMask {
	case noValueArgs:
	case leftOnlyValueArg:
		if IsFalsy(leftVal) {
			return NewSQLBool(cfg.sqlValueKind, false)
		}
		if Bool(leftVal) {
			and.left = NewSQLBool(cfg.sqlValueKind, true)
		}
	case rightOnlyValueArg:
		if IsFalsy(rightVal) {
			return NewSQLBool(cfg.sqlValueKind, false)
		}
		if Bool(rightVal) {
			and.right = NewSQLBool(cfg.sqlValueKind, true)
		}
	case bothValueArgs:
		if IsFalsy(leftVal) || IsFalsy(rightVal) {
			return NewSQLBool(cfg.sqlValueKind, false)
		}
		if leftVal.IsNull() || rightVal.IsNull() {
			return NewSQLNull(rightVal.Kind(), rightVal.EvalType())
		}
		return NewSQLBool(cfg.sqlValueKind, true)
	}
	return and
}

func (and *SQLAndExpr) String() string {
	return fmt.Sprintf("%v and %v", and.left, and.right)
}

// nolint: unparam
func (and *SQLAndExpr) reconcile() (SQLExpr, error) {
	left := and.left
	right := and.right

	if !isBooleanComparable(left.EvalType()) {
		left = NewSQLConvertExpr(left, EvalBoolean)
	}
	if !isBooleanComparable(right.EvalType()) {
		right = NewSQLConvertExpr(right, EvalBoolean)
	}
	return NewSQLAndExpr(left, right), nil
}

// ToAggregationLanguage translates SQLAndExpr into something that can
// be used in an aggregation pipeline. If SQLAndExpr cannot be translated,
// it will return nil and error.
func (and *SQLAndExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	children := eatChildren(and.ExprName(), and.Children())

	args, err := t.typedTranslateArgs(children)
	if err != nil {
		return nil, err
	}

	numChildren := len(children)

	assignments := make([]bson.DocElem, 0, numChildren)
	nullChecks := make([]interface{}, 0, numChildren)
	falsyChecks := make([]interface{}, 0, numChildren)

	containsNullLiteral := false

	columnsToNullCheck := t.ColumnsToNullCheck()

	for i, arg := range args {
		var ref interface{}
		switch arg.t {
		case argLiteralType:
			containsNullLiteral = containsNullLiteral || arg.v == nil
			continue
		case argColumnType:
			columnName := arg.v.(string)
			ref = columnName

			columnsToNullCheck[columnName] = struct{}{}

			nullChecks = append(nullChecks, toNullCheckedLetVarRef(columnName))
		case argOtherType:
			binding := fmt.Sprintf("expr%d", i)
			ref = fmt.Sprintf("$$%s", binding)

			assignments = append(assignments, bsonutil.NewDocElem(binding, arg.v))

			nullChecks = append(nullChecks, bsonutil.WrapInNullCheck(ref))
		}

		falsyChecks = append(falsyChecks, bsonutil.WrapInOp(bsonutil.OpEq, ref, 0), bsonutil.WrapInOp(bsonutil.OpEq, ref, false))
	}

	var evaluation interface{}
	if containsNullLiteral {
		// if there is a null literal, return false if any operand is false and null otherwise.
		evaluation = bsonutil.WrapInCond(false, nil, falsyChecks...)
	} else {
		// contains no literals (or only truthy literals).
		evaluation = bsonutil.WrapInCond(
			// if any operand is false, return false.
			false,
			// else if any operand is null, return null; else return true.
			bsonutil.WrapInCond(nil, true, nullChecks...),
			falsyChecks...,
		)
	}

	return wrapInLet(assignments, evaluation), nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (and *SQLAndExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return and.ToAggregationLanguage(t)
}

// ToMatchLanguage translates SQLAndExpr into something that can
// be used in an match expression. If SQLAndExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLAndExpr.
func (and *SQLAndExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	left, exLeft := t.ToMatchLanguage(and.left)
	right, exRight := t.ToMatchLanguage(and.right)

	var match bson.M
	if left == nil && right == nil {
		return nil, and
	} else if left != nil && right == nil {
		match = left
	} else if left == nil && right != nil {
		match = right
	} else {
		cond := bsonutil.NewArray()
		if v, ok := left[bsonutil.OpAnd]; ok {
			array := v.([]interface{})
			cond = append(cond, array...)
		} else {
			cond = append(cond, left)
		}

		if v, ok := right[bsonutil.OpAnd]; ok {
			array := v.([]interface{})
			cond = append(cond, array...)
		} else {
			cond = append(cond, right)
		}

		match = bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpAnd, cond))
	}

	if exLeft == nil && exRight == nil {
		return match, nil
	} else if exLeft != nil && exRight == nil {
		return match, exLeft
	} else if exLeft == nil && exRight != nil {
		return match, exRight
	}
	return match, NewSQLAndExpr(exLeft, exRight)
}

// EvalType returns the EvalType associated with SQLAndExpr.
func (*SQLAndExpr) EvalType() EvalType {
	return EvalBoolean
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

// NewSQLMultiplyExpr is a constructor for SQLMultiplyExpr.
func NewSQLMultiplyExpr(left, right SQLExpr) *SQLMultiplyExpr {
	return &SQLMultiplyExpr{sqlBinaryNode{left, right}}
}

// NewSQLSubtractExpr is a constructor for SQLSubtractExpr.
func NewSQLSubtractExpr(left, right SQLExpr) *SQLSubtractExpr {
	return &SQLSubtractExpr{sqlBinaryNode{left, right}}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLDivideExpr) ExprName() string {
	return "SQLDivideExpr"
}

// Evaluate evaluates a SQLDivideExpr into a SQLValue.
func (div *SQLDivideExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	leftVal, rightVal, err := div.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if Float64(rightVal) == 0 || hasNullValue(leftVal, rightVal) {
		return NewSQLNull(cfg.sqlValueKind, div.EvalType()), nil
	}

	return doArithmetic(leftVal, rightVal, DIV)
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLDivideExpr.
func (div *SQLDivideExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	leftVal, rightVal, valMask := div.sqlValueArgEnum()
	switch valMask {
	case noValueArgs:
	case leftOnlyValueArg:
		if leftVal.IsNull() {
			return leftVal
		}
	case rightOnlyValueArg:
		if rightVal.IsNull() {
			return rightVal
		}
		frightVal := Float64(rightVal)
		if frightVal == 0.0 {
			return NewSQLNull(rightVal.Kind(), rightVal.EvalType())
		}
		if frightVal == 1.0 {
			return div.left
		}
	case bothValueArgs:
		if leftVal.IsNull() || rightVal.IsNull() {
			return NewSQLNull(rightVal.Kind(), rightVal.EvalType())
		}
		if out, err := doArithmetic(leftVal, rightVal, DIV); err == nil {
			return out
		}
	}
	return div
}

func (div *SQLDivideExpr) reconcile() (SQLExpr, error) {
	bin, err := div.reconcileArithmetic()
	if err != nil {
		return nil, err
	}
	newExpr := SQLDivideExpr{bin}
	return &newExpr, nil
}

func (div *SQLDivideExpr) String() string {
	return fmt.Sprintf("%v/%v", div.left, div.right)
}

// ToAggregationLanguage translates SQLDivideExpr into something that can
// be used in an aggregation pipeline. If SQLDivideExpr cannot be translated,
// it will return nil and error.
func (div *SQLDivideExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	left, right, err := div.toAggregationLanguageArgs(t)
	if err != nil {
		return nil, err
	}

	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("left", left),
		bsonutil.NewDocElem("right", right),
	)

	letEvaluation := bsonutil.WrapInCond(
		nil, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpDivide, bsonutil.NewArray(
			"$$left",
			"$$right",
		))), bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpEq, bsonutil.NewArray(
			"$$right",
			0,
		))),
	)

	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil

}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (div *SQLDivideExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return div.ToAggregationLanguage(t)
}

// EvalType returns the EvalType associated with SQLDivideExpr.
func (div *SQLDivideExpr) EvalType() EvalType {
	return EvalDouble
}

func getStringColumnReference(expr SQLExpr, translation interface{}) (string, bool) {
	_, isColumnExpr := expr.(SQLColumnExpr)
	name, isString := translation.(string)
	return name, isColumnExpr && isString
}

// compExprToAggregationLanguageHelper translates a binary comparison SQLExpr
// into something that can be used in an aggregation pipeline.
// This helper is specifically intended for use with =, <>, <, <=, >, and >=.
// If the expression cannot be translated, it will return nil and error.
func compExprToAggregationLanguageHelper(leftExpr, rightExpr SQLExpr, cmpOp string, t *PushdownTranslator) (interface{}, PushdownFailure) {
	args, err := t.typedTranslateArgs([]SQLExpr{leftExpr, rightExpr})
	if err != nil {
		return nil, err
	}

	assignments, args := minimizeLetAssignments([]string{"left", "right"}, args)

	comparison := bsonutil.WrapInOp(cmpOp, args[0].v, args[1].v)
	evaluation := wrapInNullCheckedCond(t.ColumnsToNullCheck(), bsonutil.MgoNullLiteral, comparison, args...)

	return wrapInLet(assignments, evaluation), nil
}

// SQLEqualsExpr evaluates to true if the left equals the right.
type SQLEqualsExpr struct{ sqlBinaryNode }

// ExprName returns a string representing this SQLExpr's name.
func (*SQLEqualsExpr) ExprName() string {
	return "SQLEqualsExpr"
}

var _ translatableToMatch = (*SQLEqualsExpr)(nil)

// NewSQLEqualsExpr is a constructor for SQLEqualsExpr.
func NewSQLEqualsExpr(left, right SQLExpr) *SQLEqualsExpr {
	return &SQLEqualsExpr{
		sqlBinaryNode{
			left:  left,
			right: right,
		}}
}

// Evaluate evaluates a SQLEqualsExpr into a SQLValue.
func (eq *SQLEqualsExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	leftVal, rightVal, err := eq.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(cfg.sqlValueKind, eq.EvalType()), nil
	}

	c, err := CompareTo(leftVal, rightVal, st.collation)
	if err == nil {
		return NewSQLBool(cfg.sqlValueKind, c == 0), nil
	}

	return NewSQLBool(cfg.sqlValueKind, false), err
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLEqualsExpr.
func (eq *SQLEqualsExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	leftVal, rightVal, valMask := eq.sqlValueArgEnum()
	switch valMask {
	case noValueArgs:
	case leftOnlyValueArg:
		if leftVal.IsNull() {
			return leftVal
		}
	case rightOnlyValueArg:
		if rightVal.IsNull() {
			return rightVal
		}
	case bothValueArgs:
		if leftVal.IsNull() || rightVal.IsNull() {
			return NewSQLNull(rightVal.Kind(), rightVal.EvalType())
		}
		c, err := CompareTo(leftVal, rightVal, cfg.collation)
		if err == nil {
			return NewSQLBool(cfg.sqlValueKind, c == 0)
		}
	}
	if shouldFlip(eq.sqlBinaryNode) {
		left, right := eq.right, eq.left
		eq.left, eq.right = left, right
	}
	return eq
}

func (eq *SQLEqualsExpr) String() string {
	return fmt.Sprintf("%v = %v", eq.left, eq.right)
}

// ToAggregationLanguage translates SQLEqualsExpr into something that can
// be used in an aggregation pipeline. If SQLEqualsExpr cannot be translated,
// it will return nil and error.
func (eq *SQLEqualsExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return compExprToAggregationLanguageHelper(eq.left, eq.right, bsonutil.OpEq, t)
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (eq *SQLEqualsExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	left, right, err := eq.toAggregationLanguageArgs(t)
	if err != nil {
		return nil, err
	}

	// When a SQLEqualsExpr is the top level expression of the $expr in a $match
	// stage, we don't care about the difference between null and false. This
	// allows us to omit $ifNull and $cond, which should improve index usage for
	// some queries.
	return bson.M{
		bsonutil.OpEq: []interface{}{left, right},
	}, nil
}

// ToMatchLanguage translates SQLEqualsExpr into something that can
// be used in an match expression. If SQLEqualsExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLEqualsExpr.
func (eq *SQLEqualsExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	match, ok := t.translateOperator(bsonutil.OpEq, eq.left, eq.right)
	if !ok {
		return nil, eq
	}
	return match, nil
}

// EvalType returns the EvalType associated with SQLEqualsExpr.
func (*SQLEqualsExpr) EvalType() EvalType {
	return EvalBoolean
}

func (eq *SQLEqualsExpr) reconcile() (SQLExpr, error) {
	var reconciled bool

	left := eq.left
	right := eq.right

	if isBooleanColumnAndNumber(left, right) || isBooleanColumnAndNumber(right, left) {
		var col SQLColumnExpr
		var lit SQLNumber

		switch left.EvalType() {
		case EvalBoolean:
			col = left.(SQLColumnExpr)
			lit = right.(SQLNumber)
		default:
			col = right.(SQLColumnExpr)
			lit = left.(SQLNumber)
		}

		if ilit := Int64(lit); ilit == 1 || ilit == 0 {
			left = col
			right = NewSQLConvertExpr(lit.(SQLExpr), EvalBoolean)
			reconciled = true
		}
	}

	if !reconciled {
		var err error
		left, right, err = ReconcileSQLExprs(eq.left, eq.right)
		if err != nil {
			return nil, err
		}
	}

	return NewSQLEqualsExpr(left, right), nil
}

// SQLGreaterThanExpr evaluates to true when the left is greater than the right.
type SQLGreaterThanExpr struct{ sqlBinaryNode }

// ExprName returns a string representing this SQLExpr's name.
func (*SQLGreaterThanExpr) ExprName() string {
	return "SQLGreaterThanExpr"
}

var _ translatableToMatch = (*SQLGreaterThanExpr)(nil)

// NewSQLGreaterThanExpr is a constructor for SQLGreaterThanExpr.
func NewSQLGreaterThanExpr(left, right SQLExpr) *SQLGreaterThanExpr {
	return &SQLGreaterThanExpr{
		sqlBinaryNode{
			left:  left,
			right: right,
		}}
}

// Evaluate evaluates a SQLGreaterThanExpr into a SQLValue.
func (gt *SQLGreaterThanExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	leftVal, rightVal, err := gt.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(cfg.sqlValueKind, gt.EvalType()), nil
	}

	c, err := CompareTo(leftVal, rightVal, st.collation)
	if err == nil {
		return NewSQLBool(cfg.sqlValueKind, c > 0), nil
	}
	return NewSQLBool(cfg.sqlValueKind, false), err
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLGreaterThanExpr.
func (gt *SQLGreaterThanExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	leftVal, rightVal, valMask := gt.sqlValueArgEnum()
	switch valMask {
	case noValueArgs:
	case leftOnlyValueArg:
		if leftVal.IsNull() {
			return leftVal
		}
	case rightOnlyValueArg:
		if rightVal.IsNull() {
			return rightVal
		}
	case bothValueArgs:
		if leftVal.IsNull() || rightVal.IsNull() {
			return NewSQLNull(rightVal.Kind(), rightVal.EvalType())
		}
		c, err := CompareTo(leftVal, rightVal, cfg.collation)
		if err == nil {
			return NewSQLBool(cfg.sqlValueKind, c > 0)
		}
	}
	if shouldFlip(gt.sqlBinaryNode) {
		left, right := gt.right, gt.left
		return NewSQLLessThanExpr(left, right)
	}
	return gt
}

func (gt *SQLGreaterThanExpr) reconcile() (SQLExpr, error) {
	left, right, err := ReconcileSQLExprs(gt.left, gt.right)
	if err != nil {
		return nil, err
	}
	return NewSQLGreaterThanExpr(left, right), nil
}

func (gt *SQLGreaterThanExpr) String() string {
	return fmt.Sprintf("%v>%v", gt.left, gt.right)
}

// ToAggregationLanguage translates SQLGreaterThanExpr into something that can
// be used in an aggregation pipeline. If SQLGreaterThanExpr cannot be translated,
// it will return nil and error.
func (gt *SQLGreaterThanExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return compExprToAggregationLanguageHelper(gt.left, gt.right, bsonutil.OpGt, t)
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (gt *SQLGreaterThanExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return gt.ToAggregationLanguage(t)
}

// ToMatchLanguage translates SQLGreaterThanExpr into something that can
// be used in an match expression. If SQLGreaterThanExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLGreaterThanExpr.
func (gt *SQLGreaterThanExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	match, ok := t.translateOperator(bsonutil.OpGt, gt.left, gt.right)
	if !ok {
		return nil, gt
	}
	return match, nil
}

// EvalType returns the EvalType associated with SQLGreaterThanExpr.
func (*SQLGreaterThanExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLGreaterThanOrEqualExpr evaluates to true when the left is greater than or equal to the right.
type SQLGreaterThanOrEqualExpr struct{ sqlBinaryNode }

// ExprName returns a string representing this SQLExpr's name.
func (*SQLGreaterThanOrEqualExpr) ExprName() string {
	return "SQLGreaterThanOrEqualExpr"
}

// NewSQLGreaterThanOrEqualExpr is a constructor for SQLGreaterThanOrEqualExpr.
func NewSQLGreaterThanOrEqualExpr(left, right SQLExpr) *SQLGreaterThanOrEqualExpr {
	return &SQLGreaterThanOrEqualExpr{
		sqlBinaryNode{
			left:  left,
			right: right,
		}}
}

// Evaluate evaluates a SQLGreaterThanOrEqualExpr into a SQLValue.
func (gte *SQLGreaterThanOrEqualExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	leftVal, rightVal, err := gte.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(cfg.sqlValueKind, gte.EvalType()), nil
	}

	c, err := CompareTo(leftVal, rightVal, st.collation)
	if err == nil {
		return NewSQLBool(cfg.sqlValueKind, c >= 0), nil
	}

	return NewSQLBool(cfg.sqlValueKind, false), err
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLGreaterThanOrEqualExpr.
func (gte *SQLGreaterThanOrEqualExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	leftVal, rightVal, valMask := gte.sqlValueArgEnum()
	switch valMask {
	case noValueArgs:
	case leftOnlyValueArg:
		if leftVal.IsNull() {
			return leftVal
		}
	case rightOnlyValueArg:
		if rightVal.IsNull() {
			return rightVal
		}
	case bothValueArgs:
		if leftVal.IsNull() || rightVal.IsNull() {
			return NewSQLNull(rightVal.Kind(), rightVal.EvalType())
		}
		c, err := CompareTo(leftVal, rightVal, cfg.collation)
		if err == nil {
			return NewSQLBool(cfg.sqlValueKind, c >= 0)
		}
	}
	if shouldFlip(gte.sqlBinaryNode) {
		left, right := gte.right, gte.left
		return NewSQLLessThanOrEqualExpr(left, right)
	}
	return gte
}

func (gte *SQLGreaterThanOrEqualExpr) reconcile() (SQLExpr, error) {
	left, right, err := ReconcileSQLExprs(gte.left, gte.right)
	if err != nil {
		return nil, err
	}
	return NewSQLGreaterThanOrEqualExpr(left, right), nil
}

func (gte *SQLGreaterThanOrEqualExpr) String() string {
	return fmt.Sprintf("%v>=%v", gte.left, gte.right)
}

// ToAggregationLanguage translates SQLGreaterThanOrEqualExpr into something
// that can be used in an aggregation pipeline. If SQLGreaterThanOrEqualExpr
// cannot be translated, it will return nil and error.
func (gte *SQLGreaterThanOrEqualExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return compExprToAggregationLanguageHelper(gte.left, gte.right, bsonutil.OpGte, t)
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (gte *SQLGreaterThanOrEqualExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return gte.ToAggregationLanguage(t)
}

// ToMatchLanguage translates SQLGreaterThanOrEqualExpr into something that can
// be used in an match expression. If SQLGreaterThanOrEqualExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLGreaterThanOrEqualExpr.
func (gte *SQLGreaterThanOrEqualExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	match, ok := t.translateOperator(bsonutil.OpGte, gte.left, gte.right)
	if !ok {
		return nil, gte
	}
	return match, nil
}

// EvalType returns the EvalType associated with SQLGreaterThanOrEqualExpr.
func (*SQLGreaterThanOrEqualExpr) EvalType() EvalType {
	return EvalBoolean
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

// Evaluate evaluates a SQLIDivideExpr into a SQLValue.
func (div *SQLIDivideExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	leftVal, rightVal, err := div.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	frightVal := Float64(rightVal)
	if frightVal == 0.0 || hasNullValue(leftVal, rightVal) {
		// NOTE: this is per the mysql manual.
		return NewSQLNull(cfg.sqlValueKind, div.EvalType()), nil
	}

	return NewSQLInt64(cfg.sqlValueKind, int64(Float64(leftVal)/frightVal)), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLIDivideExpr.
func (div *SQLIDivideExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	leftVal, rightVal, valMask := div.sqlValueArgEnum()
	switch valMask {
	case noValueArgs:
	case leftOnlyValueArg:
		if leftVal.IsNull() {
			return leftVal
		}
	case rightOnlyValueArg:
		if rightVal.IsNull() {
			return rightVal
		}
		irightVal := Int64(rightVal)
		if irightVal == 0 {
			return NewSQLNull(rightVal.Kind(), rightVal.EvalType())
		}
		if irightVal == 1 {
			return div.left
		}
	case bothValueArgs:
		if leftVal.IsNull() || rightVal.IsNull() {
			return NewSQLNull(rightVal.Kind(), rightVal.EvalType())
		}
		frightVal := Float64(rightVal)
		return NewSQLInt64(cfg.sqlValueKind, int64(Float64(leftVal)/frightVal))
	}
	return div
}

func (div *SQLIDivideExpr) reconcile() (SQLExpr, error) {
	bin, err := div.reconcileArithmetic()
	if err != nil {
		return nil, err
	}
	newExpr := SQLIDivideExpr{bin}
	return &newExpr, nil
}

func (div *SQLIDivideExpr) String() string {
	return fmt.Sprintf("%v/%v", div.left, div.right)
}

// ToAggregationLanguage translates SQLIDivideExpr into something that can
// be used in an aggregation pipeline. If SQLIDivideExpr cannot be translated,
// it will return nil and error.
func (div *SQLIDivideExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	left, right, err := div.toAggregationLanguageArgs(t)
	if err != nil {
		return nil, err
	}

	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("left", left),
		bsonutil.NewDocElem("right", right),
	)

	letEvaluation := bsonutil.WrapInCond(
		nil, bsonutil.NewM(
			bsonutil.NewDocElem("$trunc", bsonutil.NewArray(
				bsonutil.NewM(
					bsonutil.NewDocElem(bsonutil.OpDivide, bsonutil.NewArray(
						"$$left",
						"$$right",
					)),
				),
			)),
		), bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpEq, bsonutil.NewArray(
			"$$right",
			0,
		))),
	)

	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil

}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (div *SQLIDivideExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return div.ToAggregationLanguage(t)
}

// EvalType returns the EvalType associated with SQLIDivideExpr.
func (div *SQLIDivideExpr) EvalType() EvalType {
	return preferentialType(div.left, div.right)
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

// Evaluate evaluates a SQLIsExpr into a SQLValue.
func (is *SQLIsExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	leftVal, rightVal, err := is.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if leftVal.IsNull() {
		if _, ok := rightVal.(SQLBool); ok {
			return NewSQLBool(cfg.sqlValueKind, false), nil
		}
		return NewSQLBool(cfg.sqlValueKind, true), nil
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLBool(cfg.sqlValueKind, false), nil
	}

	return NewSQLBool(cfg.sqlValueKind, Bool(leftVal) == Bool(rightVal)), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLIsExpr.
func (is *SQLIsExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	leftVal, rightVal, valMask := is.sqlValueArgEnum()
	switch valMask {
	case noValueArgs, leftOnlyValueArg:
		panic("the right argument to SQLIsExpr should always be a SQLValue")
	case rightOnlyValueArg:
	case bothValueArgs:
		if leftVal.IsNull() {
			if _, ok := rightVal.(SQLBool); ok {
				return NewSQLBool(cfg.sqlValueKind, false)
			}
			return NewSQLBool(cfg.sqlValueKind, true)
		}
		if rightVal.IsNull() {
			return NewSQLBool(cfg.sqlValueKind, false)
		}

		return NewSQLBool(cfg.sqlValueKind, Bool(leftVal) == Bool(rightVal))
	}
	if shouldFlip(is.sqlBinaryNode) {
		left, right := is.right, is.left
		is.left, is.right = left, right
	}
	return is
}

// nolint: unparam
func (is *SQLIsExpr) reconcile() (SQLExpr, error) {
	if is.right.EvalType() == EvalBoolean {
		leftType := is.left.EvalType()
		if !(leftType.IsNumeric() || leftType == EvalBoolean) {
			reconciled := convertAllExprs([]SQLExpr{is.left, is.right}, EvalBoolean)
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
func (is *SQLIsExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	left, right, err := is.toAggregationLanguageArgs(t)
	if err != nil {
		return nil, err
	}

	// if right side is {null,unknown}, it's a simple case
	sqlVal, ok := is.right.(SQLValue)
	if ok && sqlVal.IsNull() {
		return bsonutil.WrapInOp(bsonutil.OpLte,
			left,
			bsonutil.WrapInLiteral(nil),
		), nil
	}

	// if left side is a boolean, this is still simple
	if is.left.EvalType() == EvalBoolean {
		return bsonutil.WrapInOp(bsonutil.OpEq,
			left,
			right,
		), nil
	}

	// otherwise, left side is a number type
	if is.right == NewSQLBool(t.valueKind(), true) {
		return bsonutil.WrapInOp(bsonutil.OpNeq,
			bsonutil.WrapInIfNull(left, 0),
			0,
		), nil
	} else if is.right == NewSQLBool(t.valueKind(), false) {
		return bsonutil.WrapInOp(bsonutil.OpEq,
			left,
			0,
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
func (is *SQLIsExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return is.ToAggregationLanguage(t)
}

// ToMatchLanguage translates SQLIsExpr into something that can
// be used in an match expression. If SQLIsExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLIsExpr.
func (is *SQLIsExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	name, ok := t.getFieldName(is.left)
	if !ok {
		return nil, is
	}

	rightVal, ok := is.right.(SQLValue)
	if !ok {
		return nil, is
	}

	if rightVal.IsNull() {
		return bsonutil.NewM(bsonutil.NewDocElem(name, nil)), nil
	}

	rightBool, ok := rightVal.(SQLBool)
	if !ok {
		return nil, is
	}

	if rightBool.Value().(bool) {
		if is.left.EvalType() == EvalBoolean {
			return bsonutil.NewM(bsonutil.NewDocElem(name, true)), nil
		}
		return bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpAnd, bsonutil.NewArray(
				bsonutil.NewM(bsonutil.NewDocElem(name, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpNeq, 0)))),
				bsonutil.NewM(bsonutil.NewDocElem(name, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpNeq, nil)))),
			)),
		), nil
	}

	if is.left.EvalType() == EvalBoolean {
		return bsonutil.NewM(bsonutil.NewDocElem(name, false)), nil
	}
	return bsonutil.NewM(bsonutil.NewDocElem(name, 0)), nil
}

// EvalType returns the EvalType associated with SQLIsExpr.
func (*SQLIsExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLLessThanExpr evaluates to true when the left is less than the right.
type SQLLessThanExpr struct{ sqlBinaryNode }

// ExprName returns a string representing this SQLExpr's name.
func (*SQLLessThanExpr) ExprName() string {
	return "SQLLessThanExpr"
}

var _ translatableToMatch = (*SQLLessThanExpr)(nil)

// NewSQLLessThanExpr is a constructor for SQLLessThanExpr.
func NewSQLLessThanExpr(left, right SQLExpr) *SQLLessThanExpr {
	return &SQLLessThanExpr{
		sqlBinaryNode{
			left:  left,
			right: right,
		}}
}

// Evaluate evaluates a SQLLessThanExpr into a SQLValue.
func (lt *SQLLessThanExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	leftVal, rightVal, err := lt.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(cfg.sqlValueKind, lt.EvalType()), nil
	}

	c, err := CompareTo(leftVal, rightVal, st.collation)
	if err == nil {
		return NewSQLBool(cfg.sqlValueKind, c < 0), nil
	}
	return NewSQLBool(cfg.sqlValueKind, false), err
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLLessThanExpr.
func (lt *SQLLessThanExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	leftVal, rightVal, valMask := lt.sqlValueArgEnum()
	switch valMask {
	case noValueArgs:
	case leftOnlyValueArg:
		if leftVal.IsNull() {
			return leftVal
		}
	case rightOnlyValueArg:
		if rightVal.IsNull() {
			return rightVal
		}
	case bothValueArgs:
		if leftVal.IsNull() || rightVal.IsNull() {
			return NewSQLNull(rightVal.Kind(), rightVal.EvalType())
		}
		c, err := CompareTo(leftVal, rightVal, cfg.collation)
		if err == nil {
			return NewSQLBool(cfg.sqlValueKind, c < 0)
		}
	}
	if shouldFlip(lt.sqlBinaryNode) {
		left, right := lt.right, lt.left
		return NewSQLGreaterThanExpr(left, right)
	}
	return lt
}

func (lt *SQLLessThanExpr) reconcile() (SQLExpr, error) {
	left, right, err := ReconcileSQLExprs(lt.left, lt.right)
	if err != nil {
		return nil, err
	}
	return NewSQLLessThanExpr(left, right), nil
}

func (lt *SQLLessThanExpr) String() string {
	return fmt.Sprintf("%v<%v", lt.left, lt.right)
}

// ToAggregationLanguage translates SQLLessThanExpr into something that can
// be used in an aggregation pipeline. If SQLLessThanExpr cannot be translated,
// it will return nil and error.
func (lt *SQLLessThanExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return compExprToAggregationLanguageHelper(lt.left, lt.right, bsonutil.OpLt, t)
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (lt *SQLLessThanExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return lt.ToAggregationLanguage(t)
}

// ToMatchLanguage translates SQLLessThanExpr into something that can
// be used in an match expression. If SQLLessThanExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLLessThanExpr.
func (lt *SQLLessThanExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	match, ok := t.translateOperator(bsonutil.OpLt, lt.left, lt.right)
	if !ok {
		return nil, lt
	}
	return match, nil
}

// EvalType returns the EvalType associated with SQLLessThanExpr.
func (*SQLLessThanExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLLessThanOrEqualExpr evaluates to true when the left is less than or equal to the right.
type SQLLessThanOrEqualExpr struct{ sqlBinaryNode }

// ExprName returns a string representing this SQLExpr's name.
func (*SQLLessThanOrEqualExpr) ExprName() string {
	return "SQLLessThanOrEqualExpr"
}

var _ translatableToMatch = (*SQLLessThanOrEqualExpr)(nil)

// NewSQLLessThanOrEqualExpr is a constructor for SQLLessThanOrEqualExpr.
func NewSQLLessThanOrEqualExpr(left, right SQLExpr) *SQLLessThanOrEqualExpr {
	return &SQLLessThanOrEqualExpr{
		sqlBinaryNode{
			left:  left,
			right: right,
		}}
}

// Evaluate evaluates a SQLLessThanOrEqualExpr into a SQLValue.
func (lte *SQLLessThanOrEqualExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	leftVal, rightVal, err := lte.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(cfg.sqlValueKind, lte.EvalType()), nil
	}

	c, err := CompareTo(leftVal, rightVal, st.collation)
	if err == nil {
		return NewSQLBool(cfg.sqlValueKind, c <= 0), nil
	}
	return NewSQLBool(cfg.sqlValueKind, false), err
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLLessThanOrEqualExpr.
func (lte *SQLLessThanOrEqualExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	leftVal, rightVal, valMask := lte.sqlValueArgEnum()
	switch valMask {
	case noValueArgs:
	case leftOnlyValueArg:
		if leftVal.IsNull() {
			return leftVal
		}
	case rightOnlyValueArg:
		if rightVal.IsNull() {
			return rightVal
		}
	case bothValueArgs:
		if leftVal.IsNull() || rightVal.IsNull() {
			return NewSQLNull(rightVal.Kind(), rightVal.EvalType())
		}
		c, err := CompareTo(leftVal, rightVal, cfg.collation)
		if err == nil {
			return NewSQLBool(cfg.sqlValueKind, c <= 0)
		}
	}
	if shouldFlip(lte.sqlBinaryNode) {
		left, right := lte.right, lte.left
		return NewSQLGreaterThanOrEqualExpr(left, right)
	}
	return lte
}

func (lte *SQLLessThanOrEqualExpr) reconcile() (SQLExpr, error) {
	left, right, err := ReconcileSQLExprs(lte.left, lte.right)
	if err != nil {
		return nil, err
	}
	return NewSQLLessThanOrEqualExpr(left, right), nil
}

func (lte *SQLLessThanOrEqualExpr) String() string {
	return fmt.Sprintf("%v<=%v", lte.left, lte.right)
}

// ToAggregationLanguage translates SQLLessThanOrEqualExpr into something that can
// be used in an aggregation pipeline. If SQLLessThanOrEqualExpr cannot be translated,
// it will return nil and error.
func (lte *SQLLessThanOrEqualExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return compExprToAggregationLanguageHelper(lte.left, lte.right, bsonutil.OpLte, t)
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (lte *SQLLessThanOrEqualExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return lte.ToAggregationLanguage(t)
}

// ToMatchLanguage translates SQLLessThanOrEqualExpr into something that can
// be used in an match expression. If SQLLessThanOrEqualExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLLessThanOrEqualExpr.
func (lte *SQLLessThanOrEqualExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	match, ok := t.translateOperator(bsonutil.OpLte, lte.left, lte.right)
	if !ok {
		return nil, lte
	}
	return match, nil
}

// EvalType returns the EvalType associated with SQLLessThanOrEqualExpr.
func (*SQLLessThanOrEqualExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLModExpr evaluates the modulus of two expressions
type SQLModExpr struct{ sqlBinaryNode }

// ExprName returns a string representing this SQLExpr's name.
func (*SQLModExpr) ExprName() string {
	return "SQLModExpr"
}

// Evaluate evaluates a SQLModExpr into a SQLValue.
func (mod *SQLModExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	leftVal, rightVal, err := mod.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	frightVal := Float64(rightVal)
	if math.Abs(frightVal) == 0.0 || hasNullValue(leftVal, rightVal) {
		return NewSQLNull(cfg.sqlValueKind, mod.EvalType()), nil
	}

	modVal := math.Mod(Float64(leftVal), frightVal)
	if modVal == -0 {
		modVal *= -1
	}

	return NewSQLFloat(cfg.sqlValueKind, modVal), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLModExpr.
func (mod *SQLModExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	leftVal, rightVal, valMask := mod.sqlValueArgEnum()
	switch valMask {
	case noValueArgs:
	case leftOnlyValueArg:
		if leftVal.IsNull() {
			return leftVal
		}
	case rightOnlyValueArg:
		if rightVal.IsNull() {
			return rightVal
		}
		frightVal := Float64(rightVal)
		if frightVal == 0.0 {
			return NewSQLNull(rightVal.Kind(), rightVal.EvalType())
		}
		if frightVal == 1.0 {
			return mod.left
		}
	case bothValueArgs:
		if leftVal.IsNull() || rightVal.IsNull() {
			return NewSQLNull(rightVal.Kind(), rightVal.EvalType())
		}
		frightVal := Float64(rightVal)
		if math.Abs(frightVal) == 0.0 || hasNullValue(leftVal, rightVal) {
			return NewSQLNull(cfg.sqlValueKind, mod.EvalType())
		}
		modVal := math.Mod(Float64(leftVal), frightVal)
		if modVal == -0 {
			modVal *= -1
		}
		return NewSQLFloat(cfg.sqlValueKind, modVal)
	}
	return mod
}

func (mod *SQLModExpr) reconcile() (SQLExpr, error) {
	bin, err := mod.reconcileArithmetic()
	if err != nil {
		return nil, err
	}
	newExpr := SQLModExpr{bin}
	return &newExpr, nil
}

func (mod *SQLModExpr) String() string {
	return fmt.Sprintf("%v/%v", mod.left, mod.right)
}

// ToAggregationLanguage translates SQLModExpr into something that can
// be used in an aggregation pipeline. If SQLModExpr cannot be translated,
// it will return nil and error.
func (mod *SQLModExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	left, right, err := mod.toAggregationLanguageArgs(t)
	if err != nil {
		return nil, err
	}

	return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpMod, bsonutil.NewArray(
		left,
		right,
	))), nil

}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (mod *SQLModExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return mod.ToAggregationLanguage(t)
}

// EvalType returns the EvalType associated with SQLModExpr.
func (mod *SQLModExpr) EvalType() EvalType {
	return preferentialType(mod.left, mod.right)
}

// SQLMultiplyExpr evaluates to the product of two expressions
type SQLMultiplyExpr struct{ sqlBinaryNode }

// ExprName returns a string representing this SQLExpr's name.
func (*SQLMultiplyExpr) ExprName() string {
	return "SQLMultiplyExpr"
}

// Evaluate evaluates a SQLMultiplyExpr into a SQLValue.
func (mult *SQLMultiplyExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	leftVal, rightVal, err := mult.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(cfg.sqlValueKind, mult.EvalType()), nil
	}

	return doArithmetic(leftVal, rightVal, MULT)
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLMultiplyExpr.
func (mult *SQLMultiplyExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	leftVal, rightVal, valMask := mult.sqlValueArgEnum()
	switch valMask {
	case noValueArgs:
	case leftOnlyValueArg:
		if leftVal.IsNull() {
			return leftVal
		}
		if isOne(leftVal) {
			return mult.right
		}
	case rightOnlyValueArg:
		if rightVal.IsNull() {
			return rightVal
		}
		if isOne(rightVal) {
			return mult.left
		}
	case bothValueArgs:
		if leftVal.IsNull() || rightVal.IsNull() {
			return NewSQLNull(rightVal.Kind(), rightVal.EvalType())
		}
		if out, err := doArithmetic(leftVal, rightVal, MULT); err == nil {
			return out
		}
	}
	return mult
}

func (mult *SQLMultiplyExpr) reconcile() (SQLExpr, error) {
	bin, err := mult.reconcileArithmetic()
	if err != nil {
		return nil, err
	}
	newExpr := SQLMultiplyExpr{bin}
	return &newExpr, nil

}

func (mult *SQLMultiplyExpr) String() string {
	return fmt.Sprintf("%v*%v", mult.left, mult.right)
}

// ToAggregationLanguage translates SQLMultiplyExpr into something that can
// be used in an aggregation pipeline. If SQLMultiplyExpr cannot be translated,
// it will return nil and error.
func (mult *SQLMultiplyExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	children := eatChildren(mult.ExprName(), mult.Children())
	ops := make([]interface{}, len(children))

	var err PushdownFailure
	for i, c := range children {
		if ops[i], err = t.ToAggregationLanguage(c); err != nil {
			return nil, err
		}
	}

	return bsonutil.WrapInOp(bsonutil.OpMultiply, ops...), nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (mult *SQLMultiplyExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return mult.ToAggregationLanguage(t)
}

// EvalType returns the EvalType associated with SQLMultiplyExpr.
func (mult *SQLMultiplyExpr) EvalType() EvalType {
	return EvalDouble
}

// SQLNotEqualsExpr evaluates to true if the left does not equal the right.
type SQLNotEqualsExpr struct{ sqlBinaryNode }

// ExprName returns a string representing this SQLExpr's name.
func (*SQLNotEqualsExpr) ExprName() string {
	return "SQLNotEqualsExpr"
}

var _ translatableToMatch = (*SQLNotEqualsExpr)(nil)

// NewSQLNotEqualsExpr is a constructor for SQLNotEqualsExpr.
func NewSQLNotEqualsExpr(left, right SQLExpr) *SQLNotEqualsExpr {
	return &SQLNotEqualsExpr{
		sqlBinaryNode{
			left:  left,
			right: right,
		}}
}

// Evaluate evaluates a SQLNotEqualsExpr into a SQLValue.
func (neq *SQLNotEqualsExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	leftVal, rightVal, err := neq.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(cfg.sqlValueKind, neq.EvalType()), nil
	}

	c, err := CompareTo(leftVal, rightVal, st.collation)
	if err == nil {
		return NewSQLBool(cfg.sqlValueKind, c != 0), nil
	}

	return NewSQLBool(cfg.sqlValueKind, false), err
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLNotEqualsExpr.
func (neq *SQLNotEqualsExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	leftVal, rightVal, valMask := neq.sqlValueArgEnum()
	switch valMask {
	case noValueArgs:
	case leftOnlyValueArg:
		if leftVal.IsNull() {
			return leftVal
		}
	case rightOnlyValueArg:
		if rightVal.IsNull() {
			return rightVal
		}
	case bothValueArgs:
		if leftVal.IsNull() || rightVal.IsNull() {
			return NewSQLNull(rightVal.Kind(), rightVal.EvalType())
		}
		c, err := CompareTo(leftVal, rightVal, cfg.collation)
		if err == nil {
			return NewSQLBool(cfg.sqlValueKind, c != 0)
		}
	}
	if shouldFlip(neq.sqlBinaryNode) {
		left, right := neq.right, neq.left
		neq.left, neq.right = left, right
	}
	return neq
}

func (neq *SQLNotEqualsExpr) reconcile() (SQLExpr, error) {
	left, right, err := ReconcileSQLExprs(neq.left, neq.right)
	if err != nil {
		return nil, err
	}
	return NewSQLNotEqualsExpr(left, right), nil
}

func (neq *SQLNotEqualsExpr) String() string {
	return fmt.Sprintf("%v != %v", neq.left, neq.right)
}

// ToAggregationLanguage translates SQLNotEqualsExpr into something that can
// be used in an aggregation pipeline. If SQLNotEqualsExpr cannot be translated,
// it will return nil and error.
func (neq *SQLNotEqualsExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return compExprToAggregationLanguageHelper(neq.left, neq.right, bsonutil.OpNeq, t)
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (neq *SQLNotEqualsExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return neq.ToAggregationLanguage(t)
}

// ToMatchLanguage translates SQLNotEqualsExpr into something that can
// be used in an match expression. If SQLNotEqualsExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLNotEqualsExpr.
func (neq *SQLNotEqualsExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	match, ok := t.translateOperator(bsonutil.OpNeq, neq.left, neq.right)
	if !ok {
		return nil, neq
	}

	value, err := t.getValue(neq.right)
	if err != nil {
		return nil, neq
	}

	if value != nil {
		name, ok := t.getFieldName(neq.left)
		if !ok {
			return nil, neq
		}
		match = bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpAnd, bsonutil.NewArray(
				match,
				bsonutil.NewM(bsonutil.NewDocElem(name, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpNeq, nil)))),
			)),
		)

	}

	return match, nil
}

// EvalType returns the EvalType associated with SQLNotEqualsExpr.
func (*SQLNotEqualsExpr) EvalType() EvalType {
	return EvalBoolean
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

// Evaluate evaluates a SQLNullSafeEqualsExpr into a SQLValue.
func (nse *SQLNullSafeEqualsExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	leftVal, rightVal, err := nse.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	c, err := CompareTo(leftVal, rightVal, st.collation)
	if err == nil {
		return NewSQLBool(cfg.sqlValueKind, c == 0), nil
	}

	return NewSQLBool(cfg.sqlValueKind, false), err
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLNullSafeEqualsExpr.
func (nse *SQLNullSafeEqualsExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	leftVal, rightVal, valMask := nse.sqlValueArgEnum()
	switch valMask {
	// Because constant NULLs do not cause <=> to evaluate to NULL, there is
	// no room for ConstantFolding unless BOTH sides are constants.
	case noValueArgs, leftOnlyValueArg, rightOnlyValueArg:
	case bothValueArgs:
		c, err := CompareTo(leftVal, rightVal, cfg.collation)
		if err == nil {
			return NewSQLBool(cfg.sqlValueKind, c == 0)
		}
	}
	if shouldFlip(nse.sqlBinaryNode) {
		left, right := nse.right, nse.left
		nse.left, nse.right = left, right
	}
	return nse
}

func (nse *SQLNullSafeEqualsExpr) reconcile() (SQLExpr, error) {
	left, right, err := ReconcileSQLExprs(nse.left, nse.right)
	if err != nil {
		return nil, err
	}
	return NewSQLNullSafeEqualsExpr(left, right), nil
}

func (nse *SQLNullSafeEqualsExpr) String() string {
	return fmt.Sprintf("%v <=> %v", nse.left, nse.right)
}

// ToAggregationLanguage translates SQLNullSafeEqualsExpr into something that can
// be used in an aggregation pipeline. If SQLNullSafeEqualsExpr cannot be translated,
// it will return nil and error.
func (nse *SQLNullSafeEqualsExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	left, right, err := nse.toAggregationLanguageArgs(t)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapInBinOp(bsonutil.OpEq,
		bsonutil.WrapInIfNull(left, nil),
		bsonutil.WrapInIfNull(right, nil),
	), nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (nse *SQLNullSafeEqualsExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nse.ToAggregationLanguage(t)
}

// EvalType returns the EvalType associated with SQLNullSafeEqualsExpr.
func (*SQLNullSafeEqualsExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLOrExpr evaluates to true if any of its children evaluate to true.
type SQLOrExpr struct{ sqlBinaryNode }

// ExprName returns a string representing this SQLExpr's name.
func (*SQLOrExpr) ExprName() string {
	return "SQLOrExpr"
}

var _ translatableToMatch = (*SQLOrExpr)(nil)

// NewSQLOrExpr is a constructor for SQLOrExpr.
func NewSQLOrExpr(left, right SQLExpr) *SQLOrExpr {
	return &SQLOrExpr{
		sqlBinaryNode{
			left:  left,
			right: right,
		}}
}

// Evaluate evaluates a SQLOrExpr into a SQLValue.
func (or *SQLOrExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	leftVal, rightVal, err := or.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if Bool(leftVal) || Bool(rightVal) {
		return NewSQLBool(cfg.sqlValueKind, true), nil
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(cfg.sqlValueKind, or.EvalType()), nil
	}

	return NewSQLBool(cfg.sqlValueKind, false), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLOrExpr.
func (or *SQLOrExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	leftVal, rightVal, valMask := or.sqlValueArgEnum()

	switch valMask {
	case noValueArgs:
	case leftOnlyValueArg:
		if Bool(leftVal) {
			return NewSQLBool(cfg.sqlValueKind, true)
		}
		if IsFalsy(leftVal) {
			or.left = NewSQLBool(cfg.sqlValueKind, false)
		}
	case rightOnlyValueArg:
		if Bool(rightVal) {
			return NewSQLBool(cfg.sqlValueKind, true)
		}
		if IsFalsy(rightVal) {
			or.right = NewSQLBool(cfg.sqlValueKind, false)
		}
	case bothValueArgs:
		if Bool(leftVal) || Bool(rightVal) {
			return NewSQLBool(cfg.sqlValueKind, true)
		}
		if leftVal.IsNull() || rightVal.IsNull() {
			return NewSQLNull(rightVal.Kind(), rightVal.EvalType())
		}
		return NewSQLBool(cfg.sqlValueKind, false)
	}
	return or
}

func (or *SQLOrExpr) String() string {
	return fmt.Sprintf("%v or %v", or.left, or.right)
}

// nolint: unparam
func (or *SQLOrExpr) reconcile() (SQLExpr, error) {
	left := or.left
	right := or.right

	if !isBooleanComparable(left.EvalType()) {
		left = NewSQLConvertExpr(left, EvalBoolean)
	}
	if !isBooleanComparable(right.EvalType()) {
		right = NewSQLConvertExpr(right, EvalBoolean)
	}
	return NewSQLOrExpr(left, right), nil
}

// ToAggregationLanguage translates SQLOrExpr into something that can
// be used in an aggregation pipeline. If SQLOrExpr cannot be translated,
// it will return nil and error.
func (or *SQLOrExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	children := eatChildren(or.ExprName(), or.Children())

	args, err := t.typedTranslateArgs(children)
	if err != nil {
		return nil, err
	}

	numChildren := len(children)

	// the operands of the OR
	ops := make([]interface{}, 0, numChildren)

	// the conditions of the OR
	// if any condition is true, the OR will evaluate to null;
	// if none of the conditions are true, the OR will evaluate.
	conds := make([]interface{}, 0)

	assignments := make([]bson.DocElem, 0, numChildren)
	nullChecks := make([]interface{}, 0, numChildren)
	notChecks := make([]interface{}, 0, numChildren)

	containsLiteral := false
	containsNullLiteral := false
	containsFalsyLiteral := false

	columnsToNullCheck := t.ColumnsToNullCheck()

	for i, arg := range args {
		switch arg.t {
		case argLiteralType:
			containsNullLiteral = containsNullLiteral || arg.v == nil
			containsFalsyLiteral = containsFalsyLiteral || (arg.v == false || arg.v == 0)
			containsLiteral = true
		case argColumnType:
			columnName := arg.v.(string)

			columnsToNullCheck[columnName] = struct{}{}

			ops = append(ops, columnName)
			nullChecks = append(nullChecks, toNullCheckedLetVarRef(columnName))
			notChecks = append(notChecks, bsonutil.WrapInOp(bsonutil.OpNot, columnName))
		case argOtherType:
			binding := fmt.Sprintf("expr%d", i)
			bindingRef := fmt.Sprintf("$$%s", binding)

			assignments = append(assignments, bsonutil.NewDocElem(binding, arg.v))

			ops = append(ops, bindingRef)
			nullChecks = append(nullChecks, bsonutil.WrapInNullCheck(bindingRef))
			notChecks = append(notChecks, bsonutil.WrapInOp(bsonutil.OpNot, bindingRef))
		}
	}

	// if there is at least one literal, and there are no null or falsy literals, whole expression evaluates to true.
	if containsLiteral && !containsNullLiteral && !containsFalsyLiteral {
		return bsonutil.WrapInLiteral(true), nil
	}

	// build the conditions
	// if there is a null literal, return null if all other expressions are falsy.
	if containsNullLiteral {
		conds = []interface{}{bsonutil.WrapInOp(bsonutil.OpAnd, notChecks...)}
	} else if containsFalsyLiteral { // if there is a falsy literal, return null if all other expressions are null.
		conds = []interface{}{bsonutil.WrapInOp(bsonutil.OpAnd, nullChecks...)}
	} else { // if there are no literals, return null using the following condition:
		for i := range nullChecks {
			// If the "i"th expression is null, and all of the others are falsy,
			nots := append(append([]interface{}{}, notChecks[:i]...), notChecks[i+1:]...)
			if len(nots) > 0 {
				// need to include the null check along with all the other nots.
				nots = append(nots, nullChecks[i])
				conds = append(conds, bsonutil.WrapInOp(bsonutil.OpAnd, nots...))
			} else {
				conds = append(conds, nullChecks[i])
			}
		}
	}

	// build the expression
	evaluation := bsonutil.WrapInCond(nil, bsonutil.WrapInOp(bsonutil.OpOr, ops...), conds...)

	return wrapInLet(assignments, evaluation), nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (or *SQLOrExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return or.ToAggregationLanguage(t)
}

// ToMatchLanguage translates SQLOrExpr into something that can
// be used in an match expression. If SQLOrExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLOrExpr.
func (or *SQLOrExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	left, exLeft := t.ToMatchLanguage(or.left)
	if exLeft != nil {
		// cannot partially translate an OR
		return nil, or
	}
	right, exRight := t.ToMatchLanguage(or.right)
	if exRight != nil {
		// cannot partially translate an OR
		return nil, or
	}

	cond := bsonutil.NewArray()

	if v, ok := left[bsonutil.OpOr]; ok {
		array := v.([]interface{})
		cond = append(cond, array...)
	} else {
		cond = append(cond, left)
	}

	if v, ok := right[bsonutil.OpOr]; ok {
		array := v.([]interface{})
		cond = append(cond, array...)
	} else {
		cond = append(cond, right)
	}

	return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpOr, cond)), nil
}

// EvalType returns the EvalType associated with SQLOrExpr.
func (*SQLOrExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLSubtractExpr evaluates to the difference of the left expression minus the right expressions.
type SQLSubtractExpr struct{ sqlBinaryNode }

// ExprName returns a string representing this SQLExpr's name.
func (*SQLSubtractExpr) ExprName() string {
	return "SQLSubtractExpr"
}

// Evaluate evaluates a SQLSubtractExpr into a SQLValue.
func (sub *SQLSubtractExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	leftVal, rightVal, err := sub.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(cfg.sqlValueKind, sub.EvalType()), nil
	}

	return doArithmetic(leftVal, rightVal, SUB)
}

// EvalType returns the EvalType associated with SQLSubtractExpr.
func (sub *SQLSubtractExpr) EvalType() EvalType {
	return EvalDouble
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLSubtractExpr.
func (sub *SQLSubtractExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	leftVal, rightVal, valMask := sub.sqlValueArgEnum()
	switch valMask {
	case noValueArgs:
	case leftOnlyValueArg:
		if leftVal.IsNull() {
			return leftVal
		}
		if isZero(leftVal) {
			return sub.right
		}
	case rightOnlyValueArg:
		if rightVal.IsNull() {
			return rightVal
		}
		if isZero(rightVal) {
			return sub.left
		}
	case bothValueArgs:
		if leftVal.IsNull() || rightVal.IsNull() {
			return NewSQLNull(rightVal.Kind(), rightVal.EvalType())
		}
		if out, err := doArithmetic(leftVal, rightVal, SUB); err == nil {
			return out
		}
	}
	return sub
}

func (sub *SQLSubtractExpr) String() string {
	return fmt.Sprintf("%v-%v", sub.left, sub.right)
}

func (sub *SQLSubtractExpr) reconcile() (SQLExpr, error) {
	bin, err := sub.reconcileArithmetic()
	if err != nil {
		return nil, err
	}
	newExpr := SQLSubtractExpr{bin}
	return &newExpr, nil
}

// ToAggregationLanguage translates SQLSubtractExpr into something that can
// be used in an aggregation pipeline. If SQLSubtractExpr cannot be translated,
// it will return nil and error.
func (sub *SQLSubtractExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	left, right, err := sub.toAggregationLanguageArgs(t)
	if err != nil {
		return nil, err
	}

	return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpSubtract, bsonutil.NewArray(
		left,
		right,
	))), nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (sub *SQLSubtractExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return sub.ToAggregationLanguage(t)
}

// SQLXorExpr evaluates to true if and only if one of its children evaluates to true.
type SQLXorExpr struct{ sqlBinaryNode }

// NewSQLXorExpr is a constructor for SQLXorExprs.
func NewSQLXorExpr(left, right SQLExpr) *SQLXorExpr {
	return &SQLXorExpr{sqlBinaryNode{left, right}}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLXorExpr) ExprName() string {
	return "SQLXorExpr"
}

// Evaluate evaluates a SQLXorExpr into a SQLValue.
func (xor *SQLXorExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	leftVal, rightVal, err := xor.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(cfg.sqlValueKind, xor.EvalType()), nil
	}

	if IsFalsy(leftVal) {
		return rightVal.SQLBool(), nil
	}

	return NewSQLBool(cfg.sqlValueKind, !Bool(rightVal)), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLXorExpr.
func (xor *SQLXorExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	leftVal, rightVal, valMask := xor.sqlValueArgEnum()

	switch valMask {
	case noValueArgs:
	case leftOnlyValueArg:
		if leftVal.IsNull() {
			return NewSQLNull(leftVal.Kind(), leftVal.EvalType())
		}
		if IsFalsy(leftVal) {
			// Type reconciliation will ensure that xor.right is converted to boolean already, if
			// necessary.  So this is safe, unlike in the SQLOrExpr and SQLAndExpr cases.
			return xor.right
		}
		return NewSQLNotExpr(xor.right)
	case rightOnlyValueArg:
		if rightVal.IsNull() {
			return NewSQLNull(rightVal.Kind(), rightVal.EvalType())
		}
		if IsFalsy(rightVal) {
			return xor.left
		}
		return NewSQLNotExpr(xor.left)
	case bothValueArgs:
		if leftVal.IsNull() || rightVal.IsNull() {
			return NewSQLNull(rightVal.Kind(), rightVal.EvalType())
		}
		if IsFalsy(leftVal) != IsFalsy(rightVal) {
			return NewSQLBool(cfg.sqlValueKind, true)
		}
		return NewSQLBool(cfg.sqlValueKind, false)
	}
	return xor
}

func (xor *SQLXorExpr) String() string {
	return fmt.Sprintf("%v xor %v", xor.left, xor.right)
}

// nolint: unparam
func (xor *SQLXorExpr) reconcile() (SQLExpr, error) {
	left := xor.left
	right := xor.right

	if !isBooleanComparable(left.EvalType()) {
		left = NewSQLConvertExpr(left, EvalBoolean)
	}
	if !isBooleanComparable(right.EvalType()) {
		right = NewSQLConvertExpr(right, EvalBoolean)
	}
	return NewSQLXorExpr(left, right), nil
}

// ToAggregationLanguage translates SQLXorExpr into something that can
// be used in an aggregation pipeline. If SQLXorExpr cannot be translated,
// it will return nil and error.
func (xor *SQLXorExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	children := eatChildren(xor.ExprName(), xor.Children())

	args, err := t.typedTranslateArgs(children)
	if err != nil {
		return nil, err
	}

	numChildren := len(children)

	ops := make([]interface{}, 0, numChildren)
	assignments := make([]bson.DocElem, 0, numChildren)
	nullChecks := make([]interface{}, 0, numChildren)

	initialValue := false

	columnsToNullCheck := t.ColumnsToNullCheck()

	for i, arg := range args {
		switch arg.t {
		case argLiteralType:
			if arg.v == nil {
				return bsonutil.MgoNullLiteral, nil
			}

			valueIsFalsy := arg.v == 0 || arg.v == false
			initialValue = initialValue == valueIsFalsy
		case argColumnType:
			columnName := arg.v.(string)

			columnsToNullCheck[columnName] = struct{}{}

			ops = append(ops, columnName)
			nullChecks = append(nullChecks, toNullCheckedLetVarRef(columnName))
		case argOtherType:
			binding := fmt.Sprintf("expr%d", i)
			bindingRef := fmt.Sprintf("$$%s", binding)

			assignments = append(assignments, bsonutil.NewDocElem(binding, arg.v))

			ops = append(ops, bindingRef)
			nullChecks = append(nullChecks, bsonutil.WrapInNullCheck(bindingRef))
		}
	}

	evaluation := bsonutil.WrapInCond(
		bsonutil.MgoNullLiteral,
		bsonutil.WrapInReduce(
			ops,
			initialValue,
			bsonutil.WrapInOp(bsonutil.OpAnd,
				bsonutil.WrapInOp(bsonutil.OpOr, "$$this", "$$value"),
				bsonutil.WrapInOp(bsonutil.OpNot, bsonutil.WrapInOp(bsonutil.OpAnd, "$$this", "$$value")),
			),
		),
		nullChecks...,
	)

	return wrapInLet(assignments, evaluation), nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (xor *SQLXorExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return xor.ToAggregationLanguage(t)
}

// EvalType returns the EvalType associated with SQLXorExpr.
func (*SQLXorExpr) EvalType() EvalType {
	return EvalBoolean
}
