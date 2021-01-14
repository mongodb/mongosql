package evaluator

import (
	"context"
	"fmt"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/astutil"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

type sqlVarargNode struct {
	children []SQLExpr
}

// Children returns a slice of all the Node children of the Node.
func (vn sqlVarargNode) Children() []Node {
	children := make([]Node, len(vn.children))
	for i, child := range vn.children {
		children[i] = child
	}

	return children
}

// sqlValueArgEnum returns all of the child values.SQLValue arguments, if any, and a enum that tells us
// which arguments are values.SQLValues.
func (vn *sqlVarargNode) sqlValueArgEnum() ([]values.SQLValue, valueArgsEnum) {
	childrenVals := make([]values.SQLValue, 0, len(vn.children))

	for _, child := range vn.children {
		if childVal, childIsVal := child.(SQLValueExpr); childIsVal {
			childrenVals = append(childrenVals, childVal.Value)
		}
	}

	if len(childrenVals) == len(vn.children) {
		return childrenVals, allValueArgs
	} else if len(childrenVals) == 0 {
		return nil, noValueArgs
	} else {
		return childrenVals, someValueArgs
	}
}

func (vn *sqlVarargNode) ReplaceChild(i int, n Node) {
	if i >= len(vn.children) {
		panicWithInvalidIndex("sqlVariadicNode", i, len(vn.children)-1)
	}
	vn.children[i] = panicIfNotSQLExpr("sqlVariadicNode", n)
}

func (vn *sqlVarargNode) evaluateArgs(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) ([]values.SQLValue, error) {
	var vals []values.SQLValue
	for _, child := range vn.children {
		val, err := child.Evaluate(ctx, cfg, st)
		if err != nil {
			return nil, err
		}
		vals = append(vals, val)
	}
	return vals, nil
}

// eatChildrenAndFlatten attempts to recursively consume the left and right children of the
// current node and consequently flatten the tree via constant folding.
// Consumption consists of removal of the node and adoption of its children.
// Consumption of a node will succeed only if the consumer and consumed are
// of the same type. The result of the operation is children for the current
// node to adopt.
func eatChildrenAndFlatten(opName string, leftAndRight []SQLExpr) []SQLExpr {

	children := make([]SQLExpr, 0)

	for _, c := range leftAndRight {
		switch t := c.(type) {
		// The only operators supported for eating children are Add, And, Multiply, Or, and Xor.
		case *SQLAddExpr, *SQLAndExpr, *SQLMultiplyExpr, *SQLOrExpr, *SQLXorExpr:
			if c.ExprName() == opName {
				// if the child c is one of the same type as the parent (the opName
				// argument), recursively consume its children.
				children = append(children, eatChildrenAndFlatten(opName, nodesToExprs(t.Children()))...)
				continue
			}
		}

		// if that is not the case, just include c in the list of children
		children = append(children, c)
	}

	return children
}

// SQLAddExpr evaluates to the sum of expressions.
type SQLAddExpr struct{ sqlVarargNode }

// NewSQLAddExpr is a constructor for SQLAddExpr.
func NewSQLAddExpr(leftAndRight ...SQLExpr) *SQLAddExpr {
	children := eatChildrenAndFlatten("SQLAddExpr", leftAndRight)
	return &SQLAddExpr{sqlVarargNode{children}}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLAddExpr) ExprName() string {
	return "SQLAddExpr"
}

// EvalType returns the EvalType associated with SQLAddExpr.
func (add *SQLAddExpr) EvalType() types.EvalType {
	return types.EvalDouble
}

// nolint: unparam
func (add *SQLAddExpr) reconcile() (SQLExpr, error) {
	children := reconcileArithmetic(add.children)
	node := sqlVarargNode{children}
	return &SQLAddExpr{node}, nil
}

func (add *SQLAddExpr) String() string {
	var res string
	for i := 0; i < len(add.children); i++ {
		res += fmt.Sprintf("%v", add.children[i])
		if i != len(add.children)-1 {
			res += fmt.Sprintf("+")
		}
	}
	return res
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (add *SQLAddExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return add.ToAggregationLanguage(t)
}

// doArithmeticHelper runs doArithmetic on all values.SQLValue arguments and returns the cumulative sum.
func (add *SQLAddExpr) doArithmeticHelper(childrenVals []values.SQLValue) (values.SQLValue, error) {
	if len(childrenVals) == 1 {
		return childrenVals[0], nil
	}

	var err error
	var totalSum values.SQLValue = childrenVals[0]

	for _, child := range childrenVals[1:] {
		totalSum, err = doArithmetic(totalSum, child, ADD)
		if err != nil {
			return totalSum, err
		}
	}
	return totalSum, nil
}

// Evaluate evaluates a SQLAddExpr into a values.SQLValue.
func (add *SQLAddExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(add)
	if err != nil {
		return nil, err
	}

	// Iterate over all of the children
	childrenVals, err := add.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if values.HasNullValue(childrenVals...) {
		return values.NewSQLNull(cfg.sqlValueKind), nil
	}

	// totalSum is the cumulative sum of the children's values.
	totalSum, err := add.doArithmeticHelper(childrenVals)
	if err != nil {
		return nil, err
	}
	return totalSum, nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLAddExpr.
func (add *SQLAddExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(add); err != nil {
		return nil, err
	}

	childrenVals, valMask := add.sqlValueArgEnum()
	nonValueChildren := make([]SQLExpr, 0, len(childrenVals))
	allChildren := make([]SQLExpr, 0, len(add.children))

	var totalSum values.SQLValue
	var err error

	switch valMask {
	case noValueArgs:
	case allValueArgs:
		for _, childVal := range childrenVals {
			if childVal.IsNull() {
				return NewSQLValueExpr(childVal), nil
			}
		}
		totalSum, err = add.doArithmeticHelper(childrenVals)
		if err != nil {
			return add, err
		}
		return NewSQLValueExpr(totalSum), nil
	case someValueArgs:
		for _, childVal := range childrenVals {
			if childVal.IsNull() {
				return NewSQLValueExpr(childVal), nil
			}
		}

		// Collect all non- values.SQLValueExpr arguments into nonValueChildren list
		for _, child := range add.children {
			_, childIsVal := child.(SQLValueExpr)
			if !childIsVal {
				nonValueChildren = append(nonValueChildren, child)
			}
		}

		// Sum up all values.SQLValueExpr arguments
		totalSum, err = add.doArithmeticHelper(childrenVals)
		if err != nil {
			return add, err
		}

		if !values.IsZero(totalSum) {
			allChildren = append(allChildren, NewSQLValueExpr(totalSum))
			allChildren = append(allChildren, nonValueChildren...)
			return &SQLAddExpr{sqlVarargNode{allChildren}}, nil
		}
		if len(nonValueChildren) == 1 {
			return nonValueChildren[0], nil
		}
		return &SQLAddExpr{sqlVarargNode{nonValueChildren}}, nil
	}
	return add, nil
}

// ToAggregationLanguage translates SQLAddExpr into something that can
// be used in an aggregation pipeline. If SQLAddExpr cannot be translated,
// it will return nil and error.
func (add *SQLAddExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	ops := make([]ast.Expr, len(add.children))

	var err PushdownFailure
	for i, c := range add.children {
		if ops[i], err = t.ToAggregationLanguage(c); err != nil {
			return nil, err
		}
	}

	return astutil.WrapInOp(bsonutil.OpAdd, ops...), nil
}

// SQLAndExpr evaluates to true if and only if all its children evaluate to true.
type SQLAndExpr struct{ sqlVarargNode }

// NewSQLAndExpr is a constructor for SQLAndExpr.
func NewSQLAndExpr(leftAndRight ...SQLExpr) *SQLAndExpr {
	children := eatChildrenAndFlatten("SQLAndExpr", leftAndRight)
	return &SQLAndExpr{sqlVarargNode{children}}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLAndExpr) ExprName() string {
	return "SQLAndExpr"
}

var _ translatableToMatch = (*SQLAndExpr)(nil)

// EvalType returns the EvalType associated with SQLAndExpr.
func (and *SQLAndExpr) EvalType() types.EvalType {
	return types.EvalBoolean
}

// nolint: unparam
func (and *SQLAndExpr) reconcile() (SQLExpr, error) {
	children := make([]SQLExpr, 0, len(and.children))
	for _, child := range and.children {
		if !isBooleanComparable(child.EvalType()) {
			children = append(children, NewSQLConvertExpr(child, types.EvalBoolean))
		} else {
			children = append(children, child)
		}
	}
	return &SQLAndExpr{sqlVarargNode{children}}, nil
}

func (and *SQLAndExpr) String() string {
	var res string
	for i := 0; i < len(and.children); i++ {
		res += fmt.Sprintf("%v", and.children[i])
		if i != len(and.children)-1 {
			res += fmt.Sprintf(" and ")
		}
	}
	return res
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (and *SQLAndExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return and.ToAggregationLanguage(t)
}

// Evaluate evaluates a SQLAndExpr into a values.SQLValue.
func (and *SQLAndExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(and)
	if err != nil {
		return nil, err
	}
	// Iterate over all of the children
	childrenVals, err := and.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	for _, childVal := range childrenVals {
		if values.IsFalsy(childVal) {
			return values.NewSQLBool(cfg.sqlValueKind, false), nil
		}
	}

	if values.HasNullValue(childrenVals...) {
		return values.NewSQLNull(cfg.sqlValueKind), nil
	}

	return values.NewSQLBool(cfg.sqlValueKind, true), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLAndExpr.
func (and *SQLAndExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(and); err != nil {
		return nil, err
	}

	childrenVals, valMask := and.sqlValueArgEnum()
	allChildren := make([]SQLExpr, 0, len(and.children))
	hasNullChild := false

	switch valMask {
	case noValueArgs:
	case allValueArgs:
		for _, childVal := range childrenVals {
			if values.IsFalsy(childVal) {
				return NewSQLValueExpr(values.NewSQLBool(cfg.sqlValueKind, false)), nil
			}
			hasNullChild = hasNullChild || childVal.IsNull()
		}

		if hasNullChild {
			return NewSQLValueExpr(values.NewSQLNull(cfg.sqlValueKind)), nil
		}

		return NewSQLValueExpr(values.NewSQLBool(cfg.sqlValueKind, true)), nil
	case someValueArgs:
		for _, child := range and.children {
			childVal, childIsVal := child.(SQLValueExpr)
			if childIsVal {
				if values.IsFalsy(childVal.Value) {
					return NewSQLValueExpr(values.NewSQLBool(cfg.sqlValueKind, false)), nil
				}

				hasNullChild = hasNullChild || childVal.Value.IsNull()
			} else {
				allChildren = append(allChildren, child)
			}
		}

		if hasNullChild {
			allChildren = append(allChildren, NewSQLValueExpr(values.NewSQLNull(cfg.sqlValueKind)))
		}

		if len(allChildren) == 1 {
			if allChildren[0].EvalType() == types.EvalBoolean {
				return allChildren[0], nil
			}
			return NewSQLNotExpr(NewSQLNotExpr(allChildren[0])), nil
		}

		return &SQLAndExpr{sqlVarargNode{allChildren}}, nil
	}
	return and, nil
}

// ToAggregationLanguage translates SQLAndExpr into something that can
// be used in an aggregation pipeline. If SQLAndExpr cannot be translated,
// it will return nil and error.
func (and *SQLAndExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	args, err := t.translateArgs(and.children)
	if err != nil {
		return nil, err
	}

	numChildren := len(and.children)

	assignments := make([]*ast.LetVariable, 0, numChildren)
	nullChecks := make([]ast.Expr, 0, numChildren)
	falsyChecks := make([]ast.Expr, 0, numChildren)

	containsNullLiteral := false

	for i, arg := range args {
		switch a := arg.(type) {
		case *ast.Constant:
			// The constant can be Truthy, Falsy, or Null. Falsy constants
			// would be caught by constant folding and would result in the
			// $and evaluating to false. If the constant is null, all we
			// need to do is set the flag. If the constant is not null, it
			// must be truthy; we do not need to assign it and null-check/
			// falsy-check it, we can just continue.
			containsNullLiteral = containsNullLiteral || a.Value.Type == bsontype.Null

		default:
			binding := fmt.Sprintf("expr%d", i)
			ref := ast.NewVariableRef(binding)

			assignments = append(assignments, ast.NewLetVariable(binding, arg))
			nullChecks = append(nullChecks, astutil.WrapInNullCheck(ref))
			falsyChecks = append(falsyChecks,
				ast.NewBinary(bsonutil.OpEq, ref, astutil.ZeroInt32Literal),
				ast.NewBinary(bsonutil.OpEq, ref, astutil.FalseLiteral),
			)
		}

	}

	var evaluation ast.Expr
	if containsNullLiteral {
		// if there is a null literal, return false if any operand is false and null otherwise.
		evaluation = astutil.WrapInCond(astutil.FalseLiteral, astutil.NullLiteral, falsyChecks...)
	} else {
		// contains no literals (or only truthy literals).
		evaluation = astutil.WrapInCond(
			// if any operand is false, return false.
			astutil.FalseLiteral,
			// else if any operand is null, return null; else return true.
			astutil.WrapInCond(astutil.NullLiteral, astutil.TrueLiteral, nullChecks...),
			falsyChecks...,
		)
	}

	return ast.NewLet(assignments, evaluation), nil
}

// ToMatchLanguage translates SQLAndExpr into something that can
// be used in a match expression. If SQLAndExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLAndExpr.
func (and *SQLAndExpr) ToMatchLanguage(t *PushdownTranslator) (ast.Expr, SQLExpr) {
	nonNilChildrenToMatch := make([]ast.Expr, 0, len(and.children))
	nonNilExChildren := make([]SQLExpr, 0, len(and.children))

	for _, child := range and.children {
		childToMatch, exChild := t.ToMatchLanguage(child)

		if childToMatch != nil {
			nonNilChildrenToMatch = append(nonNilChildrenToMatch, childToMatch)
		}
		if exChild != nil {
			nonNilExChildren = append(nonNilExChildren, exChild)
		}
	}

	var match ast.Expr
	if len(nonNilChildrenToMatch) > 0 {
		if len(nonNilChildrenToMatch) == 1 {
			match = nonNilChildrenToMatch[0]
		} else {
			match = astutil.WrapInOp(bsonutil.OpAnd, nonNilChildrenToMatch...)
		}
	} else { // all nil
		return nil, and
	}
	if len(nonNilExChildren) == 0 {
		return match, nil
	} else if len(nonNilExChildren) == 1 {
		return match, nonNilExChildren[0]
	}
	node := SQLAndExpr{sqlVarargNode{nonNilExChildren}}
	return match, &node
}

// SQLMultiplyExpr evaluates to the product of expressions
type SQLMultiplyExpr struct{ sqlVarargNode }

// NewSQLMultiplyExpr is a constructor for SQLMultiplyExpr.
func NewSQLMultiplyExpr(leftAndRight ...SQLExpr) *SQLMultiplyExpr {
	children := eatChildrenAndFlatten("SQLMultiplyExpr", leftAndRight)
	return &SQLMultiplyExpr{sqlVarargNode{children}}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLMultiplyExpr) ExprName() string {
	return "SQLMultiplyExpr"
}

// EvalType returns the EvalType associated with SQLMultiplyExpr.
func (mult *SQLMultiplyExpr) EvalType() types.EvalType {
	return types.EvalDouble
}

// nolint: unparam
func (mult *SQLMultiplyExpr) reconcile() (SQLExpr, error) {
	children := reconcileArithmetic(mult.children)
	node := sqlVarargNode{children}
	return &SQLMultiplyExpr{node}, nil
}

func (mult *SQLMultiplyExpr) String() string {
	var res string
	for i := 0; i < len(mult.children); i++ {
		res += fmt.Sprintf("%v", mult.children[i])
		if i != len(mult.children)-1 {
			res += fmt.Sprintf("*")
		}
	}
	return res
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (mult *SQLMultiplyExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return mult.ToAggregationLanguage(t)
}

// doArithmeticHelper runs doArithmetic on all values.SQLValue arguments and returns the cumulative product.
func (mult *SQLMultiplyExpr) doArithmeticHelper(childrenVals []values.SQLValue) (values.SQLValue, error) {
	if len(childrenVals) == 1 {
		return childrenVals[0], nil
	}

	var err error
	var totalProduct values.SQLValue = childrenVals[0]
	for _, child := range childrenVals[1:] {
		totalProduct, err = doArithmetic(totalProduct, child, MULT)
		if err != nil {
			return totalProduct, err
		}
	}
	return totalProduct, nil
}

// Evaluate evaluates a SQLMultiplyExpr into a values.SQLValue.
func (mult *SQLMultiplyExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(mult)
	if err != nil {
		return nil, err
	}

	// Iterate over all of the children
	childrenVals, err := mult.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if values.HasNullValue(childrenVals...) {
		return values.NewSQLNull(cfg.sqlValueKind), nil
	}

	// totalSum is the cumulative sum of the children's values.
	totalProduct, err := mult.doArithmeticHelper(childrenVals)
	if err != nil {
		return nil, err
	}
	return totalProduct, nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLMultiplyExpr.
func (mult *SQLMultiplyExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(mult); err != nil {
		return nil, err
	}

	childrenVals, valMask := mult.sqlValueArgEnum()
	nonValueChildren := make([]SQLExpr, 0, len(childrenVals))
	allChildren := make([]SQLExpr, 0, len(mult.children))

	var totalProduct values.SQLValue
	var err error

	switch valMask {
	case noValueArgs:
	case allValueArgs:
		for _, childVal := range childrenVals {
			if childVal.IsNull() {
				return NewSQLValueExpr(childVal), nil
			}
		}
		totalProduct, err = mult.doArithmeticHelper(childrenVals)
		if err != nil {
			return mult, err
		}
		return NewSQLValueExpr(totalProduct), nil
	case someValueArgs:
		for _, childVal := range childrenVals {
			if childVal.IsNull() {
				return NewSQLValueExpr(childVal), nil
			}
		}

		// Collect all non- values.SQLValueExpr arguments into nonValueChildren list
		for _, child := range mult.children {
			_, childIsVal := child.(SQLValueExpr)
			if !childIsVal {
				nonValueChildren = append(nonValueChildren, child)
			}
		}

		// Take the product of all of the values.SQLValueExpr arguments
		totalProduct, err = mult.doArithmeticHelper(childrenVals)
		if err != nil {
			return mult, err
		}

		if !values.IsOne(totalProduct) {
			allChildren = append(allChildren, NewSQLValueExpr(totalProduct))
			allChildren = append(allChildren, nonValueChildren...)
			return &SQLMultiplyExpr{sqlVarargNode{allChildren}}, nil
		}
		if len(nonValueChildren) == 1 {
			return nonValueChildren[0], nil
		}
		return &SQLMultiplyExpr{sqlVarargNode{nonValueChildren}}, nil
	}
	return mult, nil
}

// ToAggregationLanguage translates SQLMultiplyExpr into something that can
// be used in an aggregation pipeline. If SQLMultiplyExpr cannot be translated,
// it will return nil and error.
func (mult *SQLMultiplyExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	ops := make([]ast.Expr, len(mult.children))

	var err PushdownFailure
	for i, c := range mult.children {
		if ops[i], err = t.ToAggregationLanguage(c); err != nil {
			return nil, err
		}
	}

	return astutil.WrapInOp(bsonutil.OpMultiply, ops...), nil
}

// SQLOrExpr evaluates to true if any of its children evaluate to true.
type SQLOrExpr struct{ sqlVarargNode }

// NewSQLOrExpr is a constructor for SQLOrExpr.
func NewSQLOrExpr(leftAndRight ...SQLExpr) *SQLOrExpr {
	children := eatChildrenAndFlatten("SQLOrExpr", leftAndRight)
	return &SQLOrExpr{sqlVarargNode{children}}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLOrExpr) ExprName() string {
	return "SQLOrExpr"
}

var _ translatableToMatch = (*SQLOrExpr)(nil)

// EvalType returns the EvalType associated with SQLOrExpr.
func (or *SQLOrExpr) EvalType() types.EvalType {
	return types.EvalBoolean
}

// nolint: unparam
func (or *SQLOrExpr) reconcile() (SQLExpr, error) {
	children := make([]SQLExpr, 0, len(or.children))
	for _, child := range or.children {
		if !isBooleanComparable(child.EvalType()) {
			children = append(children, NewSQLConvertExpr(child, types.EvalBoolean))
		} else {
			children = append(children, child)
		}
	}
	return &SQLOrExpr{sqlVarargNode{children}}, nil
}

func (or *SQLOrExpr) String() string {
	var res string
	for i := 0; i < len(or.children); i++ {
		res += fmt.Sprintf("%v", or.children[i])
		if i != len(or.children)-1 {
			res += fmt.Sprintf(" or ")
		}
	}
	return res
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (or *SQLOrExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return or.ToAggregationLanguage(t)
}

// Evaluate evaluates a SQLOrExpr into a values.SQLValue.
func (or *SQLOrExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(or)
	if err != nil {
		return nil, err
	}
	// Iterate over all of the children
	childrenVals, err := or.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	for _, childVal := range childrenVals {
		if values.Bool(childVal) {
			return values.NewSQLBool(cfg.sqlValueKind, true), nil
		}
	}

	if values.HasNullValue(childrenVals...) {
		return values.NewSQLNull(cfg.sqlValueKind), nil
	}

	return values.NewSQLBool(cfg.sqlValueKind, false), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLOrExpr.
func (or *SQLOrExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(or); err != nil {
		return nil, err
	}

	childrenVals, valMask := or.sqlValueArgEnum()
	allChildren := make([]SQLExpr, 0, len(or.children))
	hasNullChild := false

	switch valMask {
	case noValueArgs:
	case allValueArgs:
		for _, childVal := range childrenVals {
			if values.Bool(childVal) {
				return NewSQLValueExpr(values.NewSQLBool(cfg.sqlValueKind, true)), nil
			}
			hasNullChild = hasNullChild || childVal.IsNull()
		}
		if hasNullChild {
			return NewSQLValueExpr(values.NewSQLNull(cfg.sqlValueKind)), nil
		}
		return NewSQLValueExpr(values.NewSQLBool(cfg.sqlValueKind, false)), nil
	case someValueArgs:
		for _, child := range or.children {
			childVal, childIsVal := child.(SQLValueExpr)
			if childIsVal {
				if values.Bool(childVal.Value) {
					return NewSQLValueExpr(values.NewSQLBool(cfg.sqlValueKind, true)), nil
				}
				hasNullChild = hasNullChild || childVal.Value.IsNull()
			} else {
				allChildren = append(allChildren, child)
			}
		}

		if hasNullChild {
			allChildren = append(allChildren, NewSQLValueExpr(values.NewSQLNull(cfg.sqlValueKind)))
		}

		if len(allChildren) == 1 {
			if allChildren[0].EvalType() == types.EvalBoolean {
				return allChildren[0], nil
			}
			return NewSQLNotExpr(NewSQLNotExpr(allChildren[0])), nil
		}

		return &SQLOrExpr{sqlVarargNode{allChildren}}, nil
	}
	return or, nil
}

// ToAggregationLanguage translates SQLOrExpr into something that can
// be used in an aggregation pipeline. If SQLOrExpr cannot be translated,
// it will return nil and error.
func (or *SQLOrExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	args, err := t.translateArgs(or.children)
	if err != nil {
		return nil, err
	}

	numChildren := len(or.children)
	ops := make([]ast.Expr, 0, numChildren)
	assignments := make([]*ast.LetVariable, 0, numChildren)

	for i, arg := range args {
		if _, ok := arg.(*ast.Constant); ok {
			isTrueLiteral := values.Bool(or.children[i].(SQLValueExpr).Value)
			if isTrueLiteral {
				return astutil.TrueLiteral, nil
			}
		}

		binding := fmt.Sprintf("expr%d", i)
		bindingRef := ast.NewVariableRef(binding)
		assignments = append(assignments, ast.NewLetVariable(binding, arg))
		ops = append(ops, bindingRef)
	}

	evaluation := astutil.WrapInReduce(
		ast.NewArray(ops...),
		astutil.FalseLiteral,
		astutil.WrapInSwitch(
			astutil.FalseLiteral,
			astutil.WrapInCase(
				ast.NewBinary(bsonutil.OpOr, astutil.ThisVarRef, astutil.ValueVarRef),
				astutil.TrueLiteral,
			),
			astutil.WrapInEqCase(
				astutil.ThisVarRef, astutil.NullLiteral,
				astutil.NullLiteral,
			),
			astutil.WrapInEqCase(
				astutil.ValueVarRef, astutil.NullLiteral,
				astutil.NullLiteral,
			),
		),
	)
	return ast.NewLet(assignments, evaluation), nil
}

// ToMatchLanguage translates SQLOrExpr into something that can
// be used in an match expression. If SQLOrExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLOrExpr.
func (or *SQLOrExpr) ToMatchLanguage(t *PushdownTranslator) (ast.Expr, SQLExpr) {
	childrenToMatch := make([]ast.Expr, 0, len(or.children))

	for _, child := range or.children {
		childToMatch, exChild := t.ToMatchLanguage(child)
		childrenToMatch = append(childrenToMatch, childToMatch)

		if exChild != nil {
			return nil, or
		}
	}

	return astutil.WrapInOp(bsonutil.OpOr, childrenToMatch...), nil
}

// SQLXorExpr evaluates to true if and only if one of its children evaluates to true.
type SQLXorExpr struct{ sqlVarargNode }

// NewSQLXorExpr is a constructor for SQLXorExprs.
func NewSQLXorExpr(leftAndRight ...SQLExpr) *SQLXorExpr {
	children := eatChildrenAndFlatten("SQLXorExpr", leftAndRight)
	return &SQLXorExpr{sqlVarargNode{children}}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLXorExpr) ExprName() string {
	return "SQLXorExpr"
}

// EvalType returns the EvalType associated with SQLXorExpr.
func (xor *SQLXorExpr) EvalType() types.EvalType {
	return types.EvalBoolean
}

// nolint: unparam
func (xor *SQLXorExpr) reconcile() (SQLExpr, error) {
	children := make([]SQLExpr, 0, len(xor.children))
	for _, child := range xor.children {
		if !isBooleanComparable(child.EvalType()) {
			children = append(children, NewSQLConvertExpr(child, types.EvalBoolean))
		} else {
			children = append(children, child)
		}
	}
	return &SQLXorExpr{sqlVarargNode{children}}, nil
}

func (xor *SQLXorExpr) String() string {
	var res string
	for i := 0; i < len(xor.children); i++ {
		res += fmt.Sprintf("%v", xor.children[i])
		if i != len(xor.children)-1 {
			res += fmt.Sprintf(" xor ")
		}
	}
	return res
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (xor *SQLXorExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return xor.ToAggregationLanguage(t)
}

// Evaluate evaluates a SQLXorExpr into a values.SQLValue.
func (xor *SQLXorExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(xor)
	if err != nil {
		return nil, err
	}
	// Iterate over all of the children
	childrenVals, err := xor.evaluateArgs(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if values.HasNullValue(childrenVals...) {
		return values.NewSQLNull(cfg.sqlValueKind), nil
	}

	numTrue := 0
	for _, childVal := range childrenVals {
		if !values.IsFalsy(childVal) {
			numTrue++
		}
	}

	value := true
	if numTrue%2 == 0 {
		value = false
	}
	return values.NewSQLBool(cfg.sqlValueKind, value), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLXorExpr.
func (xor *SQLXorExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(xor); err != nil {
		return nil, err
	}

	childrenVals, valMask := xor.sqlValueArgEnum()
	allChildren := make([]SQLExpr, 0, len(xor.children))
	numTrue := 0

	switch valMask {
	case noValueArgs:
	case allValueArgs:
		for _, childVal := range childrenVals {
			if childVal.IsNull() {
				return NewSQLValueExpr(values.NewSQLNull(childVal.Kind())), nil
			}
			if !values.IsFalsy(childVal) {
				numTrue++
			}
		}
		if numTrue%2 == 0 {
			return NewSQLValueExpr(values.NewSQLBool(cfg.sqlValueKind, false)), nil
		}
		return NewSQLValueExpr(values.NewSQLBool(cfg.sqlValueKind, true)), nil

	case someValueArgs:
		for _, child := range xor.children {
			childVal, childIsVal := child.(SQLValueExpr)
			if childIsVal {
				if childVal.Value.IsNull() {
					return NewSQLValueExpr(values.NewSQLNull(childVal.Value.Kind())), nil
				}

				if !values.IsFalsy(childVal.Value) {
					numTrue++
				}
			} else {
				allChildren = append(allChildren, child)
			}
		}
		if numTrue%2 == 0 {
			if len(allChildren) == 1 {
				if allChildren[0].EvalType() == types.EvalBoolean {
					return allChildren[0], nil
				}
				return xor, nil
			}
			return &SQLXorExpr{sqlVarargNode{allChildren}}, nil
		}
		if len(allChildren) == 1 {
			return NewSQLNotExpr(allChildren[0]), nil
		}
		return &SQLNotExpr{sqlUnaryNode{&SQLXorExpr{sqlVarargNode{allChildren}}}}, nil
	}
	return xor, nil
}

// ToAggregationLanguage translates SQLXorExpr into something that can
// be used in an aggregation pipeline. If SQLXorExpr cannot be translated,
// it will return nil and error.
func (xor *SQLXorExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLXorExpr",
			"cannot push down to MongoDB < 3.4",
		)
	}

	args, err := t.translateArgs(xor.children)
	if err != nil {
		return nil, err
	}

	numChildren := len(xor.children)

	ops := make([]ast.Expr, 0, numChildren)
	assignments := make([]*ast.LetVariable, 0, numChildren)
	nullChecks := make([]ast.Expr, 0, numChildren)

	for i, arg := range args {
		switch a := arg.(type) {
		case *ast.Constant:
			if a.Value.Type == bsontype.Null {
				return astutil.NullLiteral, nil
			}

		default:
			binding := fmt.Sprintf("expr%d", i)
			bindingRef := ast.NewVariableRef(binding)

			assignments = append(assignments, ast.NewLetVariable(binding, arg))

			ops = append(ops, bindingRef)
			nullChecks = append(nullChecks, astutil.WrapInNullCheck(bindingRef))
		}
	}

	evaluation := astutil.WrapInCond(
		astutil.NullLiteral,
		astutil.WrapInReduce(
			ast.NewArray(ops...),
			astutil.BooleanValue(false),
			ast.NewBinary(bsonutil.OpAnd,
				ast.NewBinary(bsonutil.OpOr, astutil.ThisVarRef, astutil.ValueVarRef),
				ast.NewFunction(bsonutil.OpNot,
					astutil.WrapInOp(bsonutil.OpAnd, astutil.ThisVarRef, astutil.ValueVarRef),
				),
			),
		),
		nullChecks...,
	)

	return ast.NewLet(assignments, evaluation), nil
}
