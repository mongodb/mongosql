package evaluator

import (
	"context"
	"fmt"
	"regexp"

	"github.com/10gen/mongoast/ast"

	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/internal/astutil"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
)

// These are the possible values for the ArithmeticOperator enum.
const (
	ADD ArithmeticOperator = iota
	DIV
	MULT
	SUB
)

// VariableKind indicates if the variable is a system variable or a user variable.
type VariableKind string

const (
	// GlobalStatus is a global server status variable.
	GlobalStatus VariableKind = "global_status"
	// GlobalVariable is a global system variable.
	GlobalVariable VariableKind = "global"
	// SessionStatus is a session(local) server status variable.
	SessionStatus VariableKind = "session_status"
	// SessionVariable is a session(local) variable.
	SessionVariable VariableKind = "session"
	// UserVariable is a custom variable associated with a session(local).
	UserVariable VariableKind = "user"
)

// SQLExpr is the base type for a SQL expression.
type SQLExpr interface {
	Node
	types.EvalTyper
	fmt.Stringer
	// Evaluate evaluates the receiver expression in memory.
	Evaluate(context.Context, *ExecutionConfig, *ExecutionState) (values.SQLValue, error)
	// FoldConstants performs constant-folding on this SQLExpr, returning a
	// SQLExpr that is simplified as much as possible.
	FoldConstants(cfg *OptimizerConfig) (SQLExpr, error)
	// ExprName returns a string representing this SQLExpr's name.
	ExprName() string
	// ToAggregationPredicate translates this expression to the aggregation language
	// to be evaluated as a predicate directly in a $match stage via $expr.
	ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure)
	// ToAggregationLanguage translates a SQLExpr to a MongoDB aggregation expression.
	ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure)
	// reconcile returns a transformed version of this SQLExpr that has wrapped
	// its children in SQLConvertExprs to ensure that they are the correct
	// types.
	reconcile() (SQLExpr, error)
}

// ArithmeticOperator is a type that defines the arithmetic operators: add,
// subtract, multiply, divide.
type ArithmeticOperator byte

// MongoFilterExpr holds a MongoDB filter expression.
type MongoFilterExpr struct {
	column SQLColumnExpr
	expr   SQLExpr
	query  ast.Expr
}

// Children returns a slice of all the Node children of the Node.
func (e *MongoFilterExpr) Children() []Node {
	return []Node{e.column, e.expr}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (e *MongoFilterExpr) ReplaceChild(i int, n Node) {
	switch i {
	case 0:
		ce, ok := n.(SQLColumnExpr)
		if !ok {
			panic(fmt.Sprintf("child 0 to MongoFilterExpr must be a SQLColumnExpr not %T", n))
		}
		e.column = ce
	case 1:
		e.expr = panicIfNotSQLExpr(e.ExprName(), n)
	default:
		panicWithInvalidIndex(e.ExprName(), i, 1)
	}
}

// ExprName returns a string representing this SQLExpr's name.
func (*MongoFilterExpr) ExprName() string {
	return "MongoFilterExpr"
}

var _ translatableToMatch = (*MongoFilterExpr)(nil)

// Evaluate evaluates a MongoFilterExpr into a values.SQLValue.
func (e *MongoFilterExpr) Evaluate(context.Context, *ExecutionConfig, *ExecutionState) (values.SQLValue, error) {
	return nil, fmt.Errorf("could not evaluate predicate with mongo filter expression")
}

// nolint: unparam
func (e *MongoFilterExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *MongoFilterExpr.
func (e *MongoFilterExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	return e, nil
}

func (e *MongoFilterExpr) String() string {
	return fmt.Sprintf("%v=%v", e.column.String(), e.expr.String())
}

// ToMatchLanguage translates MongoFilterExpr into something that can
// be used in an match expression. If MongoFilterExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original MongoFilterExpr.
func (e *MongoFilterExpr) ToMatchLanguage(t *PushdownTranslator) (ast.Expr, SQLExpr) {
	return e.query, nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *MongoFilterExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates MongoFilterExpr into something that can
// be used in an aggregation pipeline. If MongoFilterExpr cannot be translated,
// it will return nil and error.
func (e *MongoFilterExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return nil, newUntranslatableExprFailure(e)
}

// EvalType returns the EvalType associated with MongoFilterExpr.
func (*MongoFilterExpr) EvalType() types.EvalType {
	return types.EvalBoolean
}

// SQLAssignmentExpr handles assigning a value to a variable.
type SQLAssignmentExpr struct {
	variable *SQLVariableExpr
	expr     SQLExpr
}

// Children returns a slice of all the Node children of the Node.
func (e *SQLAssignmentExpr) Children() []Node {
	return []Node{e.variable, e.expr}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (e *SQLAssignmentExpr) ReplaceChild(i int, n Node) {
	switch i {
	case 0:
		ve, ok := n.(*SQLVariableExpr)
		if !ok {
			panic(fmt.Sprintf("child 0 to SQLAssignmentExpr must be a *SQLVariableExpr not %T", n))
		}
		e.variable = ve
	case 1:
		e.expr = panicIfNotSQLExpr(e.ExprName(), n)
	default:
		panicWithInvalidIndex(e.ExprName(), i, 1)
	}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLAssignmentExpr) ExprName() string {
	return "SQLAssignmentExpr"
}

// NewSQLAssignmentExpr is a constructor for SQLAssignmentExpr.
func NewSQLAssignmentExpr(variable *SQLVariableExpr, expr SQLExpr) *SQLAssignmentExpr {
	return &SQLAssignmentExpr{
		variable: variable,
		expr:     expr,
	}
}

// Evaluate evaluates a SQLAssignmentExpr into a values.SQLValue.
func (e *SQLAssignmentExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(e)
	if err != nil {
		return nil, err
	}

	value, err := e.expr.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	return value, nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLAssignmentExpr.
func (e *SQLAssignmentExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(e); err != nil {
		return nil, err
	}

	return e, nil
}

// nolint: unparam
func (e *SQLAssignmentExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

func (e *SQLAssignmentExpr) String() string {
	return fmt.Sprintf("%s := %s", e.variable.String(), e.expr.String())
}

// EvalType returns the EvalType associated with SQLAssignmentExpr.
func (e *SQLAssignmentExpr) EvalType() types.EvalType {
	return e.expr.EvalType()
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLAssignmentExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLAssignmentExpr into something that can
// be used in an aggregation pipeline. If SQLAssignmentExpr cannot be translated,
// it will return nil and error.
func (e *SQLAssignmentExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return nil, newUntranslatableExprFailure(e)
}

// SQLBenchmarkExpr evaluates expr the number of times given by count.
// https://dev.mysql.com/doc/refman/5.5/en/information-functions.html#function_benchmark
type SQLBenchmarkExpr struct {
	count SQLExpr
	expr  SQLExpr
}

// Children returns a slice of all the Node children of the Node.
func (e *SQLBenchmarkExpr) Children() []Node {
	return []Node{e.count, e.expr}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (e *SQLBenchmarkExpr) ReplaceChild(i int, n Node) {
	switch i {
	case 0:
		e.count = panicIfNotSQLExpr(e.ExprName(), n)
	case 1:
		e.expr = panicIfNotSQLExpr(e.ExprName(), n)
	default:
		panicWithInvalidIndex(e.ExprName(), i, 1)
	}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLBenchmarkExpr) ExprName() string {
	return "SQLBenchmarkExpr"
}

// NewSQLBenchmarkExpr is a constructor for SQLBenchmarkExpr.
func NewSQLBenchmarkExpr(count, expr SQLExpr) *SQLBenchmarkExpr {
	return &SQLBenchmarkExpr{
		count: count,
		expr:  expr,
	}
}

// Evaluate evaluates a SQLBenchmarkExpr into a values.SQLValue.
func (e *SQLBenchmarkExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(e)
	if err != nil {
		return nil, err
	}

	count, err := e.count.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	for i := int64(0); i < values.Int64(count); i++ {
		_, err := e.expr.Evaluate(ctx, cfg, st)
		if err != nil {
			return nil, err
		}
	}

	return values.NewSQLInt64(cfg.sqlValueKind, 0), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for SQLBenchmarkExpr.
func (e *SQLBenchmarkExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(e); err != nil {
		return nil, err
	}

	return e, nil
}

// nolint: unparam
func (e *SQLBenchmarkExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

func (e *SQLBenchmarkExpr) String() string {
	return fmt.Sprintf("benchmark(%s, %s)", e.count.String(), e.expr.String())
}

// EvalType returns the EvalType associated with SQLBenchmarkExpr.
func (e *SQLBenchmarkExpr) EvalType() types.EvalType {
	return types.EvalInt64
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLBenchmarkExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLBenchmarkExpr into something that can
// be used in an aggregation pipeline. If SQLBenchmarkExpr cannot be translated,
// it will return nil and error.
func (e *SQLBenchmarkExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return nil, newUntranslatableExprFailure(e)
}

// SQLCaseExpr holds a number of cases to evaluate as well as the value
// to return if any of the cases is matched. If none is matched,
// 'elseValue' is evaluated and returned.
type SQLCaseExpr struct {
	elseValue      SQLExpr
	caseConditions []caseCondition
}

// NewSQLCaseExpr is a constructor for SQLCaseExpr.
func NewSQLCaseExpr(elseValue SQLExpr, caseConditions ...caseCondition) *SQLCaseExpr {
	return &SQLCaseExpr{
		elseValue:      elseValue,
		caseConditions: caseConditions,
	}
}

// Children returns a slice of all the Node children of the Node.
func (e *SQLCaseExpr) Children() []Node {
	ret := make([]Node, 2*len(e.caseConditions)+1)
	for i := 0; i < 2*len(e.caseConditions); i += 2 {
		caseCond := e.caseConditions[i/2]
		ret[i], ret[i+1] = caseCond.matcher, caseCond.then
	}
	ret[len(ret)-1] = e.elseValue
	return ret
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (e *SQLCaseExpr) ReplaceChild(i int, n Node) {
	if i < 0 || i > 2*len(e.caseConditions) {
		panicWithInvalidIndex(e.ExprName(), i, 2*len(e.caseConditions))
	}
	if i == 2*len(e.caseConditions) {
		e.elseValue = panicIfNotSQLExpr(e.ExprName(), n)
		return
	}
	if i%2 == 0 {
		e.caseConditions[i/2].matcher = panicIfNotSQLExpr(e.ExprName(), n)
		return
	}
	e.caseConditions[i/2].then = panicIfNotSQLExpr(e.ExprName(), n)
}

// ExprName returns a string representing this SQLExpr's name.
func (SQLCaseExpr) ExprName() string {
	return "SQLCaseExpr"
}

// Evaluate evaluates a SQLCaseExpr into a values.SQLValue.
func (e *SQLCaseExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(e)
	if err != nil {
		return nil, err
	}

	for _, condition := range e.caseConditions {
		result, err := condition.matcher.Evaluate(ctx, cfg, st)
		if err != nil {
			return nil, err
		}

		if values.Bool(result) {
			return condition.then.Evaluate(ctx, cfg, st)
		}
	}

	return e.elseValue.Evaluate(ctx, cfg, st)
}

// FoldConstants simplifies expressions containing constants when it is able to for SQLCaseExpr.
func (e *SQLCaseExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(e); err != nil {
		return nil, err
	}

	newCases := make([]caseCondition, 0)
	for _, caseCond := range e.caseConditions {
		if matchVal, ok := caseCond.matcher.(SQLValueExpr); ok {
			// If the matchVal is Falsy, we want to remove
			// it from the caseConditions.
			if values.Bool(matchVal.Value) {
				newCases = append(newCases, newCaseCondition(matchVal, caseCond.then))
			}
		} else {
			newCases = append(newCases, newCaseCondition(caseCond.matcher, caseCond.then))
		}
	}
	if len(newCases) == 0 {
		return e.elseValue, nil
	}
	// If caseConditions[0].match is a values.SQLValue, it must be true,
	// as we have removed all false values.SQLValues, in such a case,
	// return the value of the case. If it is not a values.SQLValue,
	// we cannot simplify any further because it must contain
	// a column value.
	if _, ok := newCases[0].matcher.(SQLValueExpr); ok {
		return newCases[0].then, nil
	}
	e.caseConditions = newCases
	return e, nil
}

// nolint: unparam
func (e *SQLCaseExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

func (e *SQLCaseExpr) String() string {
	str := fmt.Sprintf("case ")
	for _, cond := range e.caseConditions {
		str += fmt.Sprintf("%v ", cond.String())
	}
	if e.elseValue != nil {
		str += fmt.Sprintf("else %v ", e.elseValue.String())
	}
	str += fmt.Sprintf("end")
	return str
}

// ToAggregationLanguage translates SQLCaseExpr into something that can
// be used in an aggregation pipeline. If SQLCaseExpr cannot be translated,
// it will return nil and error.
func (e *SQLCaseExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	elseValue, err := t.ToAggregationLanguage(e.elseValue)
	if err != nil {
		return nil, err
	}

	conditions := make([]ast.Expr, len(e.caseConditions))
	thens := make([]ast.Expr, len(e.caseConditions))
	for i, condition := range e.caseConditions {
		var c ast.Expr
		if matcher, ok := condition.matcher.(*SQLEqualsExpr); ok {
			newMatcher := NewSQLOrExpr(
				matcher,
				NewSQLEqualsExpr(matcher.left, NewSQLValueExpr(values.NewSQLBool(t.valueKind(), true))))
			c, err = t.ToAggregationLanguage(newMatcher)
			if err != nil {
				return nil, err
			}
		} else {
			c, err = t.ToAggregationLanguage(condition.matcher)
			if err != nil {
				return nil, err
			}
		}

		then, err := t.ToAggregationLanguage(condition.then)
		if err != nil {
			return nil, err
		}

		conditions[i] = c
		thens[i] = then
	}

	if len(conditions) != len(thens) {
		return nil, newPushdownFailure(
			e.ExprName(),
			"number of conditions does not match number of thens",
			"expr", e.String(),
		)
	}

	cases := elseValue

	for i := len(conditions) - 1; i >= 0; i-- {
		cases = astutil.WrapInCond(thens[i], cases, conditions[i])
	}

	return cases, nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLCaseExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// EvalType returns the EvalType associated with SQLCaseExpr.
func (e *SQLCaseExpr) EvalType() types.EvalType {
	conds := []types.EvalTyper{e.elseValue}
	for _, cond := range e.caseConditions {
		conds = append(conds, cond.then)
	}
	// Verified that Case expressions in MySQL use
	// VarcharHighPriority.
	s := &types.EvalTypeSorter{VarcharHighPriority: true}
	return preferentialTypeWithSorter(s, conds...)
}

// SQLColumnExpr represents a column reference.
type SQLColumnExpr struct {
	selectID     int
	databaseName string
	tableName    string
	columnName   string
	columnType   results.ColumnType
	correlated   bool
}

// NewSQLColumnExpr creates a new SQLColumnExpr with its required fields.
// NewSQLColumnExpr is a constructor for SQLColumnExpr.
func NewSQLColumnExpr(selectID int, databaseName, tableName, columnName string, evalType types.EvalType, mongoType schema.MongoType) SQLColumnExpr {
	return SQLColumnExpr{
		selectID:     selectID,
		databaseName: databaseName,
		tableName:    tableName,
		columnName:   columnName,
		columnType: results.NewColumnType(
			evalType,
			mongoType,
		),
	}
}

func newSQLColumnExprFromColumn(c *results.Column) SQLColumnExpr {
	return NewSQLColumnExpr(c.SelectID,
		c.Database,
		c.Table,
		c.Name,
		c.EvalType,
		c.MongoType,
	)
}

// ExprName returns a string representing this SQLExpr's name.
func (SQLColumnExpr) ExprName() string {
	return "SQLColumnExpr"
}

var _ translatableToMatch = (*SQLColumnExpr)(nil)

// Children returns a slice of all the Node children of the Node.
func (SQLColumnExpr) Children() []Node {
	return []Node{}
}

// Evaluate evaluates a SQLColumnExpr into a values.SQLValue.
func (e SQLColumnExpr) Evaluate(_ context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(e)
	if err != nil {
		return nil, err
	}

	for _, row := range st.rows {
		if value, ok := row.GetField(e.selectID, e.databaseName, e.tableName, e.columnName); ok {
			return values.ConvertTo(value, e.EvalType()), nil
		}
	}

	for _, row := range st.correlatedRows {
		if value, ok := row.GetField(e.selectID, e.databaseName, e.tableName, e.columnName); ok {
			return values.ConvertTo(value, e.EvalType()), nil
		}
	}

	panic(fmt.Sprintf("cannot find column %q", e))
}

// FoldConstants simplifies expressions containing constants when it is able to for SQLColumnExpr.
func (e SQLColumnExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(e); err != nil {
		return nil, err
	}

	return e, nil
}

// nolint: unparam
func (e SQLColumnExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (e SQLColumnExpr) ReplaceChild(i int, n Node) {
	panicWithInvalidIndex(e.ExprName(), i, -1)
}

func (e SQLColumnExpr) String() string {
	var str string
	if e.databaseName != "" {
		str += e.databaseName + "."
	}

	if e.tableName != "" {
		str += e.tableName + "."
	}
	str += e.columnName
	return str
}

// ToAggregationLanguage translates SQLColumnExpr into something that can
// be used in an aggregation pipeline. If SQLColumnExpr cannot be translated,
// it will return nil and error.
func (e SQLColumnExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	if e.correlated {
		return nil, newPushdownFailure(e.ExprName(), "cannot push down correlated subquery column")
	}

	ref, ok := t.LookupFieldRef(e.databaseName, e.tableName, e.columnName)
	if !ok {
		return nil, newPushdownFailure(
			e.ExprName(),
			"failed to find field name",
			"expr", e.String(),
		)
	}

	if fieldRef, isFieldRef := ref.(*ast.FieldRef); isFieldRef && astutil.AllParentsAreFieldRefs(fieldRef) {
		return getProjectedFieldName(astutil.FieldRefString(fieldRef), e.columnType.EvalType), nil
	}

	return ref, nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e SQLColumnExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToMatchLanguage translates SQLColumnExpr into something that can
// be used in an match expression. If SQLColumnExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLColumnExpr.
func (e SQLColumnExpr) ToMatchLanguage(t *PushdownTranslator) (ast.Expr, SQLExpr) {
	if e.correlated {
		return nil, e
	}
	ref, ok := t.LookupFieldRef(e.databaseName, e.tableName, e.columnName)
	if !ok {
		return nil, e
	}

	if e.EvalType() != types.EvalBoolean {
		return ast.NewBinary(bsonutil.OpNeq, ref, astutil.NullLiteral), e
	}

	return astutil.WrapInOp(bsonutil.OpAnd,
		ast.NewBinary(bsonutil.OpNeq, ref, astutil.FalseLiteral),
		ast.NewBinary(bsonutil.OpNeq, ref, astutil.NullLiteral),
		ast.NewBinary(bsonutil.OpNeq, ref, astutil.ZeroInt32Literal),
	), nil
}

// EvalType returns the EvalType associated with SQLColumnExpr.
func (e SQLColumnExpr) EvalType() types.EvalType {
	return e.columnType.EvalType
}

func (e SQLColumnExpr) isAggregateReplacementColumn() bool {
	return e.tableName == ""
}

// SQLConvertExpr represents a conversion
// of the expression expr to the target
// EvalType.
type SQLConvertExpr struct {
	expr       SQLExpr
	targetType types.EvalType
}

// Children returns a slice of all the Node children of the Node.
func (e *SQLConvertExpr) Children() []Node {
	return []Node{e.expr}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLConvertExpr) ExprName() string {
	return "SQLConvertExpr"
}

// NewSQLConvertExpr is a constructor for SQLConvertExpr.
func NewSQLConvertExpr(expr SQLExpr, targetType types.EvalType) *SQLConvertExpr {
	return &SQLConvertExpr{
		expr:       expr,
		targetType: targetType,
	}
}

// Evaluate evaluates a SQLConvertExpr into a values.SQLValue.
func (e *SQLConvertExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(e)
	if err != nil {
		return nil, err
	}

	v, err := e.expr.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	return values.ConvertTo(v, e.targetType), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLConvertExpr.
func (e *SQLConvertExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(e); err != nil {
		return nil, err
	}

	if exprVal, ok := e.expr.(SQLValueExpr); ok {
		out := NewSQLValueExpr(values.ConvertTo(exprVal.Value, e.targetType))
		return out, nil
	}
	return e, nil
}

// nolint: unparam
func (e *SQLConvertExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (e *SQLConvertExpr) ReplaceChild(i int, n Node) {
	switch i {
	case 0:
		e.expr = panicIfNotSQLExpr(e.ExprName(), n)
	default:
		panicWithInvalidIndex(e.ExprName(), i, 0)
	}
}

func (e *SQLConvertExpr) String() string {
	prettyTypeName := string(types.EvalTypeToSQLType(e.targetType))
	return "Convert(" + e.expr.String() + ", " + prettyTypeName + ")"
}

// EvalType returns the EvalType associated with SQLConvertExpr.
func (e *SQLConvertExpr) EvalType() types.EvalType {
	return e.targetType
}

func (e *SQLConvertExpr) translateMongoSQL(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	if !t.versionAtLeast(4, 0, 0) {
		return nil, newPushdownFailure(
			e.ExprName(),
			"cannot push down mongosql-mode conversions to MongoDB < 4.0",
		)
	}

	expr, err := t.ToAggregationLanguage(e.expr)
	if err != nil {
		return nil, err
	}

	converted := translateConvert(expr, e.expr.EvalType(), e.targetType)
	return converted, nil
}

func (e *SQLConvertExpr) translateMySQL(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	//
	// The following type conversions are pushed down:
	//
	//     any numeric type -> any numeric type
	//     any numeric type -> string
	//     any numeric type -> bool
	//     datetime         -> date
	//     datetime         -> string
	//     datetime         -> any numeric type
	//     date             -> datetime
	//     date             -> string
	//     date             -> any numeric type
	//     objectid         -> string
	//

	if !t.versionAtLeast(3, 6, 0) {
		return nil, newPushdownFailure(
			e.ExprName(),
			"cannot push down mysql-mode conversions to MongoDB < 3.6",
		)
	}

	fromType := e.expr.EvalType()
	toType := e.targetType

	expr, err := t.ToAggregationLanguage(e.expr)
	if err != nil {
		return nil, err
	}

	switch fromType {
	case types.EvalInt32, types.EvalInt64,
		types.EvalUint32, types.EvalUint64,
		types.EvalDecimal128, types.EvalBoolean:

		switch toType {
		case types.EvalInt32, types.EvalInt64,
			types.EvalUint32, types.EvalUint64,
			types.EvalDecimal128, types.EvalDouble,
			types.EvalString, types.EvalBoolean:
			return e.translateMongoSQL(t)
		}

	case types.EvalDouble:
		switch toType {
		case types.EvalInt32, types.EvalInt64,
			types.EvalUint32, types.EvalUint64,
			types.EvalDecimal128, types.EvalBoolean:
			return e.translateMongoSQL(t)
		}

	case types.EvalDatetime:
		year := ast.NewFunction(bsonutil.OpYear, expr)
		month := ast.NewFunction(bsonutil.OpMonth, expr)
		day := ast.NewFunction(bsonutil.OpDayOfMonth, expr)
		hour := ast.NewFunction(bsonutil.OpHour, expr)
		minute := ast.NewFunction(bsonutil.OpMinute, expr)
		second := ast.NewFunction(bsonutil.OpSecond, expr)
		millisecond := ast.NewFunction(bsonutil.OpMillisecond, expr)

		switch toType {
		case types.EvalDate:
			asDate := ast.NewFunction(bsonutil.OpDateFromParts, ast.NewDocument(
				ast.NewDocumentElement("year", year),
				ast.NewDocumentElement("month", month),
				ast.NewDocumentElement("day", day),
			))

			return asDate, nil

		case types.EvalInt32, types.EvalInt64, types.EvalUint32, types.EvalUint64:
			asNum := astutil.WrapInOp(bsonutil.OpAdd,
				second,
				ast.NewBinary(bsonutil.OpMultiply, minute, astutil.Int32Value(100)),
				ast.NewBinary(bsonutil.OpMultiply, hour, astutil.Int32Value(10000)),
				ast.NewBinary(bsonutil.OpMultiply, day, astutil.Int32Value(1000000)),
				ast.NewBinary(bsonutil.OpMultiply, month, astutil.Int32Value(100000000)),
				ast.NewBinary(bsonutil.OpMultiply, year, astutil.Int64Value(10000000000)),
			)

			return asNum, nil

		case types.EvalDecimal128, types.EvalDouble:
			asNum := astutil.WrapInOp(bsonutil.OpAdd,
				ast.NewBinary(bsonutil.OpDivide, millisecond, astutil.Int32Value(1000)),
				second,
				ast.NewBinary(bsonutil.OpMultiply, minute, astutil.Int32Value(100)),
				ast.NewBinary(bsonutil.OpMultiply, hour, astutil.Int32Value(10000)),
				ast.NewBinary(bsonutil.OpMultiply, day, astutil.Int32Value(1000000)),
				ast.NewBinary(bsonutil.OpMultiply, month, astutil.Int32Value(100000000)),
				ast.NewBinary(bsonutil.OpMultiply, year, astutil.Int64Value(10000000000)),
			)

			return asNum, nil

		case types.EvalString:
			asString := ast.NewFunction(bsonutil.OpDateToString, ast.NewDocument(
				ast.NewDocumentElement("date", expr),
				ast.NewDocumentElement("format", astutil.StringValue("%Y-%m-%d %H:%M:%S.%L000")),
			))

			return asString, nil

		}

	case types.EvalDate:
		year := ast.NewFunction(bsonutil.OpYear, expr)
		month := ast.NewFunction(bsonutil.OpMonth, expr)
		day := ast.NewFunction(bsonutil.OpDayOfMonth, expr)

		switch toType {
		case types.EvalDatetime:
			asDate := ast.NewFunction(bsonutil.OpDateFromParts, ast.NewDocument(
				ast.NewDocumentElement("year", year),
				ast.NewDocumentElement("month", month),
				ast.NewDocumentElement("day", day),
			))

			return asDate, nil

		case types.EvalInt32, types.EvalInt64,
			types.EvalUint32, types.EvalUint64,
			types.EvalDecimal128, types.EvalDouble:

			asNum := astutil.WrapInOp(bsonutil.OpAdd,
				day,
				ast.NewBinary(bsonutil.OpMultiply, month, astutil.Int32Value(100)),
				ast.NewBinary(bsonutil.OpMultiply, year, astutil.Int32Value(10000)),
			)

			return asNum, nil

		case types.EvalString:
			asString := ast.NewFunction(bsonutil.OpDateToString, ast.NewDocument(
				ast.NewDocumentElement("date", expr),
				ast.NewDocumentElement("format", astutil.StringValue("%Y-%m-%d")),
			))

			return asString, nil

		}

	case types.EvalObjectID:
		switch toType {
		case types.EvalString:
			return e.translateMongoSQL(t)
		}

	default:
		// mysql-mode pushdown not yet implemented for conversions from other types
	}

	return nil, newPushdownFailure(
		e.ExprName(),
		fmt.Sprintf(
			"cannot push down mysql-mode conversion from type '%s'",
			types.EvalTypeToMongoType(fromType),
		),
	)
}

// ToAggregationLanguage translates SQLConvertExpr into something that can
// be used in an aggregation pipeline. At the moment, SQLConvertExpr cannot be
// translated, so this function will always return nil and error.
func (e *SQLConvertExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	if e.targetType == types.EvalObjectID {
		svexpr, ok := e.expr.(SQLValueExpr)
		if !ok {
			return nil, newPushdownFailure(
				e.ExprName(),
				"can only push down SQLVarchar as ObjectId",
			)
		}
		sv, ok := svexpr.Value.(values.SQLVarchar)
		if !ok {
			return nil, newPushdownFailure(
				e.ExprName(),
				"can only push down SQLVarchar as ObjectId",
			)
		}
		return NewSQLValueExpr(sv.SQLObjectID()).ToAggregationLanguage(t)
	}

	mode := t.Cfg.sqlValueKind
	switch mode {
	case values.MySQLValueKind:
		return e.translateMySQL(t)
	case values.MongoSQLValueKind:
		return e.translateMongoSQL(t)
	default:
		panic(fmt.Errorf("impossible value %v for cfg.sqlValueKind", mode))
	}
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLConvertExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

func innerSubqueryPushdownFailure(expr SQLExpr) PushdownFailure {
	return newPushdownFailure(
		expr.ExprName(),
		"could not push down subquery plan",
	)
}

func multiRowSubqueryPushdownFailure(expr SQLExpr) PushdownFailure {
	return newPushdownFailure(
		expr.ExprName(),
		mysqlerrors.Defaultf(mysqlerrors.ErSubqueryNoOneRow, 1).Message,
	)
}

func wrapExprErrWithPushdownFailure(expr SQLExpr, err error) PushdownFailure {
	return newPushdownFailure(
		expr.ExprName(),
		"unexpected error during translation",
		"error", err.Error(),
	)
}

// SQLExistsExpr is a wrapper around a PlanStage representing a check for
// results from a subquery. It evaluates to true if any result is returned
// from the subquery. A SQLExistsExpr always evaluates to a boolean.
type SQLExistsExpr struct {
	correlated bool
	plan       PlanStage
	// We always cache non-correlated subquery results in their entirety.
	// This is a fine place to be more clever in the future.
	// SQLExistsExpr is cached as a boolean.
	cache values.SQLBool
}

// Children returns a slice of all the Node children of the Node.
func (e *SQLExistsExpr) Children() []Node {
	return []Node{e.plan}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (e *SQLExistsExpr) ReplaceChild(i int, n Node) {
	switch i {
	case 0:
		e.plan = panicIfNotPlanStage(e.ExprName(), n)
	default:
		panicWithInvalidIndex(e.ExprName(), i, 0)
	}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLExistsExpr) ExprName() string {
	return "SQLExistsExpr"
}

// NewSQLExistsExpr is a constructor for SQLExistsExpr.
func NewSQLExistsExpr(correlated bool, plan PlanStage) *SQLExistsExpr {
	return &SQLExistsExpr{
		correlated: correlated,
		plan:       plan,
	}
}

func (*SQLExistsExpr) evaluateFromPlan(ctx context.Context,
	cfg *ExecutionConfig, st *ExecutionState, plan PlanStage) (values.SQLBool, error) {
	iter, err := plan.Open(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	row := &results.Row{}

	hasNext := iter.Next(ctx, row)
	// release this memory here... it will be re-allocated by a consuming
	// stage
	if err = cfg.memoryMonitor.Release(row.Data.Size()); err != nil {
		_ = iter.Close()
		return nil, err
	}
	if hasNext {
		return values.NewSQLBool(cfg.sqlValueKind, true), iter.Close()
	}
	return values.NewSQLBool(cfg.sqlValueKind, false), iter.Close()
}

// Evaluate evaluates a SQLExistsExpr into a values.SQLValue.
// EXISTS returns true if its subquery returns at least one row.
// False is returned if there are no rows. EXISTS never returns NULL.
func (e *SQLExistsExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(e)
	if err != nil {
		return nil, err
	}

	if e.correlated {
		return e.evaluateFromPlan(ctx, cfg, st.SubqueryState(), e.plan)
	}
	if e.cache == nil {
		var err error
		// Populate cache.
		e.cache, err = e.evaluateFromPlan(ctx, cfg, st, e.plan)
		if err != nil {
			return nil, err
		}
	}
	// Read from cache.
	return e.cache, nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLExistsExpr.
func (e *SQLExistsExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(e); err != nil {
		return nil, err
	}

	return e, nil
}

// nolint: unparam
func (e *SQLExistsExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

func (e *SQLExistsExpr) String() string {
	return fmt.Sprintf("exists(%s)", PrettyPrintPlan(e.plan))
}

// EvalType returns the EvalType associated with SQLExistsExpr.
func (*SQLExistsExpr) EvalType() types.EvalType {
	return types.EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLExistsExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLExistsExpr into something that can
// be used in an aggregation pipeline. If SQLExistsExpr cannot be translated,
// it will return nil and error.
func (e *SQLExistsExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	subPlan := e.plan
	subPlanMs, ok := subPlan.(*MongoSourceStage)
	if !ok {
		return nil, innerSubqueryPushdownFailure(e)
	}

	err := t.addSubqueryCmpLookupStage(subPlanMs)
	if err != nil {
		return nil, wrapExprErrWithPushdownFailure(e, err)
	}

	// We don't actually care _which_ column we get, because since this is the
	// relational world, if there is N values for the first column, there will
	// also be N values for second, third and so on.  So the first column
	// suffices and will always exist.
	ref, err := lookupArrayRef(subPlanMs, subPlanMs.Columns()[0])
	if err != nil {
		return nil, wrapExprErrWithPushdownFailure(e, err)
	}

	return ast.NewBinary(bsonutil.OpNeq,
		astutil.ZeroInt32Literal,
		ast.NewFunction(bsonutil.OpSize, ref),
	), nil
}

// SQLLikeExpr evaluates to true if the left is 'like' the right.
type SQLLikeExpr struct {
	left          SQLExpr
	right         SQLExpr
	escape        SQLExpr
	caseSensitive bool
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLLikeExpr) ExprName() string {
	return "SQLLikeExpr"
}

var _ translatableToMatch = (*SQLLikeExpr)(nil)

// NewSQLLikeExpr is a constructor for SQLLikeExpr.
func NewSQLLikeExpr(left, right, escape SQLExpr, caseSensitive bool) *SQLLikeExpr {
	return &SQLLikeExpr{
		left:          left,
		right:         right,
		escape:        escape,
		caseSensitive: caseSensitive,
	}
}

// Children returns a slice of all the Node children of the Node.
func (e *SQLLikeExpr) Children() []Node {
	return []Node{e.left, e.right, e.escape}
}

// Evaluate evaluates a SQLLikeExpr into a values.SQLValue.
func (e *SQLLikeExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(e)
	if err != nil {
		return nil, err
	}

	value, err := e.left.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if values.HasNullValue(value) {
		return values.NewSQLNull(cfg.sqlValueKind), nil
	}

	data := value.String()

	value, err = e.right.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if values.HasNullValue(value) {
		return values.NewSQLNull(cfg.sqlValueKind), nil
	}

	escape, err := e.escape.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	escapeSeq := []rune(escape.String())
	if len(escapeSeq) > 1 {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongArguments, "ESCAPE")
	}

	var escapeChar rune
	if len(escapeSeq) == 1 {
		escapeChar = escapeSeq[0]
	}

	pattern := "(?i)"
	if e.caseSensitive {
		pattern = ""
	}
	pattern += ConvertSQLValueToPattern(value, escapeChar)

	matches, err := regexp.Match(pattern, []byte(data))
	if err != nil {
		return nil, err
	}

	return values.NewSQLBool(cfg.sqlValueKind, matches), nil
}

// nolint: unparam
func (e *SQLLikeExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (e *SQLLikeExpr) ReplaceChild(i int, n Node) {
	switch i {
	case 0:
		e.left = panicIfNotSQLExpr(e.ExprName(), n)
	case 1:
		e.right = panicIfNotSQLExpr(e.ExprName(), n)
	case 2:
		e.escape = panicIfNotSQLExpr(e.ExprName(), n)
	default:
		panicWithInvalidIndex(e.ExprName(), i, 2)
	}
}

func (e *SQLLikeExpr) String() string {
	likeType := "like"
	if e.caseSensitive {
		likeType = "like binary"
	}
	return fmt.Sprintf("%v %v %v", e.left, likeType, e.right)
}

// ToMatchLanguage translates SQLLikeExpr into something that can
// be used in an match expression. If SQLLikeExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLLikeExpr.
func (e *SQLLikeExpr) ToMatchLanguage(t *PushdownTranslator) (ast.Expr, SQLExpr) {
	// we cannot do a like comparison on an ObjectID in mongodb.
	if c, ok := e.left.(SQLColumnExpr); ok &&
		c.columnType.MongoType == schema.MongoObjectID {
		return nil, e
	}

	ref, ok := t.getFieldOrVariableRef(e.left)
	if !ok {
		return nil, e
	}

	value, ok := e.right.(SQLValueExpr)
	if !ok {
		return nil, e
	}

	if values.HasNullValue(value.Value) {
		return nil, e
	}

	escape, ok := e.escape.(SQLValueExpr)
	if !ok {
		return nil, e
	}

	escapeSeq := []rune(escape.String())
	if len(escapeSeq) > 1 {
		return nil, e
	}

	var escapeChar rune
	if len(escapeSeq) == 1 {
		escapeChar = escapeSeq[0]
	}

	pattern := ConvertSQLValueToPattern(value.Value, escapeChar)
	opts := "i"
	if e.caseSensitive {
		opts = ""
	}

	name, ok := astutil.GetRefName(ref)
	if !ok {
		return nil, e
	}

	return ast.NewDocument(ast.NewDocumentElement(
		name, astutil.WrapInRegex(pattern, opts),
	)), nil
}

// evaluate performs evaluation given all values.SQLValues.
func (e *SQLLikeExpr) evaluate(sqlValueKind values.SQLValueKind, left, right, escape values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(left) {
		return left, nil
	}

	if values.HasNullValue(right) {
		return right, nil
	}

	data := values.String(left)

	escapeSeq := []rune(values.String(escape))
	if len(escapeSeq) > 1 {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongArguments, "ESCAPE")
	}

	var escapeChar rune
	if len(escapeSeq) == 1 {
		escapeChar = escapeSeq[0]
	}

	pattern := "(?i)"
	if e.caseSensitive {
		pattern = ""
	}
	pattern += ConvertSQLValueToPattern(right, escapeChar)

	matches, err := regexp.Match(pattern, []byte(data))
	if err != nil {
		return nil, err
	}

	return values.NewSQLBool(sqlValueKind, matches), nil
}

// EvalType returns the EvalType associated with SQLLikeExpr.
func (*SQLLikeExpr) EvalType() types.EvalType {
	return types.EvalBoolean
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLLikeExpr.
func (e *SQLLikeExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(e); err != nil {
		return nil, err
	}

	valCount := 0
	var left, right, escape SQLValueExpr
	var ok bool
	if left, ok = e.left.(SQLValueExpr); ok {
		if left.Value.IsNull() {
			return NewSQLValueExpr(values.NewSQLNull(cfg.sqlValueKind)), nil
		}
		valCount++
	}
	if right, ok = e.right.(SQLValueExpr); ok {
		if right.Value.IsNull() {
			return NewSQLValueExpr(values.NewSQLNull(cfg.sqlValueKind)), nil
		}
		valCount++
	}
	if escape, ok = e.escape.(SQLValueExpr); ok {
		valCount++
	}
	if valCount == 3 {
		val, err := e.evaluate(cfg.sqlValueKind, left.Value, right.Value, escape.Value)
		if err != nil {
			return NewSQLValueExpr(values.NewSQLNull(cfg.sqlValueKind)), nil
		}
		return NewSQLValueExpr(val), nil
	}
	return e, nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLLikeExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLLikeExpr into something that can
// be used in an aggregation pipeline. If SQLLikeExpr cannot be translated,
// it will return nil and error.
func (e *SQLLikeExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return nil, newUntranslatableExprFailure(e)
}

// SQLRegexExpr evaluates to true if the operand matches the regex patttern.
type SQLRegexExpr struct {
	operand, pattern SQLExpr
}

// Children returns a slice of all the Node children of the Node.
func (e *SQLRegexExpr) Children() []Node {
	return []Node{e.operand, e.pattern}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (e *SQLRegexExpr) ReplaceChild(i int, n Node) {
	switch i {
	case 0:
		e.operand = panicIfNotSQLExpr(e.ExprName(), n)
	case 1:
		e.pattern = panicIfNotSQLExpr(e.ExprName(), n)
	default:
		panicWithInvalidIndex(e.ExprName(), i, 1)
	}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLRegexExpr) ExprName() string {
	return "SQLRegexExpr"
}

var _ translatableToMatch = (*SQLRegexExpr)(nil)

// Evaluate evaluates a SQLRegexExpr into a values.SQLValue.
func (e *SQLRegexExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(e)
	if err != nil {
		return nil, err
	}

	operand, err := e.operand.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	pattern, err := e.pattern.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	return e.evaluate(cfg.sqlValueKind, operand, pattern)
}

// evaluate performs evaluation given all values.SQLValues.
func (e *SQLRegexExpr) evaluate(sqlValueKind values.SQLValueKind, operand, pattern values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(operand, pattern) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	matcher, err := regexp.CompilePOSIX(values.String(pattern))
	if err != nil {
		return nil, err
	}
	match := matcher.Find([]byte(values.String(operand)))
	if match != nil {
		return values.NewSQLBool(sqlValueKind, true), nil
	}
	return values.NewSQLBool(sqlValueKind, false), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLRegexExpr.
func (e *SQLRegexExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(e); err != nil {
		return nil, err
	}

	var operand, pattern SQLValueExpr
	var ok bool
	valCount := 0
	if operand, ok = e.operand.(SQLValueExpr); ok {
		if operand.Value.IsNull() {
			return NewSQLValueExpr(values.NewSQLNull(cfg.sqlValueKind)), nil
		}
		valCount++
	}
	if pattern, ok = e.pattern.(SQLValueExpr); ok {
		if pattern.Value.IsNull() {
			return NewSQLValueExpr(values.NewSQLNull(cfg.sqlValueKind)), nil
		}
		valCount++
	}
	if valCount == 2 {
		val, err := e.evaluate(cfg.sqlValueKind, operand.Value.SQLVarchar(), pattern.Value.SQLVarchar())
		if err != nil {
			return NewSQLValueExpr(values.NewSQLNull(cfg.sqlValueKind)), nil
		}
		return NewSQLValueExpr(val), nil
	}
	return e, nil
}

// nolint: unparam
func (e *SQLRegexExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

func (e *SQLRegexExpr) String() string {
	return fmt.Sprintf("%s matches %s", e.operand.String(), e.pattern.String())
}

// ToMatchLanguage translates SQLRegexExpr into something that can
// be used in an match expression. If SQLRegexExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLRegexExpr.
func (e *SQLRegexExpr) ToMatchLanguage(t *PushdownTranslator) (ast.Expr, SQLExpr) {
	ref, ok := t.getFieldOrVariableRef(e.operand)
	if !ok {
		return nil, e
	}

	patternExpr, ok := e.pattern.(SQLValueExpr)
	if !ok {
		return nil, e
	}
	pattern, ok := patternExpr.Value.(values.SQLVarchar)
	if !ok {
		return nil, e
	}
	// We need to check if the pattern is valid Extended POSIX regex
	// because MongoDB supports a superset of this specification called
	// PCRE.
	_, err := regexp.CompilePOSIX(pattern.String())
	if err != nil {
		return nil, e
	}

	name, ok := astutil.GetRefName(ref)
	if !ok {
		return nil, e
	}

	return ast.NewDocument(ast.NewDocumentElement(
		name, astutil.WrapInRegex(pattern.String(), ""),
	)), nil
}

// EvalType returns the EvalType associated with SQLRegexExpr.
func (*SQLRegexExpr) EvalType() types.EvalType {
	return types.EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLRegexExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLRegexExpr into something that can
// be used in an aggregation pipeline. If SQLRegexExpr cannot be translated,
// it will return nil and error.
func (e *SQLRegexExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return nil, newUntranslatableExprFailure(e)
}

// evaluatePlan converts a PlanStage into a table in memory, represented
// as a slice of slices of SQLValue. This table is used as the runtime value of a
// subquery expression and can be cached. Optimization opportunity:
// this function copies all of its input data, value-by-value.
func evaluatePlan(ctx context.Context, cfg *ExecutionConfig,
	st *ExecutionState, plan PlanStage) ([][]values.SQLValue, error) {

	iter, err := plan.Open(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	row := &results.Row{}
	var valueTable [][]values.SQLValue

	for iter.Next(ctx, row) {
		valueRow := make([]values.SQLValue, len(row.Data))
		// release this memory here... it will be re-allocated by a consuming
		// stage
		if err = cfg.memoryMonitor.Release(row.Data.Size()); err != nil {
			_ = iter.Close()
			return nil, err
		}

		// The full data copy here is unwanted.
		// This is a good place to attempt to improve performance.
		for i, value := range row.Data {
			valueRow[i] = value.Data
		}
		valueTable = append(valueTable, valueRow)
	}

	return valueTable, iter.Close()
}

// evaluatePlanToScalar converts a PlanStage into a row in memory, represented
// as a slice of SQLValue. This row is used as the runtime value of a
// subquery expression and can be cached. Optimization opportunity:
// this function copies all of its input data, value-by-value.
// This function implements the MySQL behavior of evaluating an empty input
// into a row of NULLs with the same dimension as the input.
func evaluatePlanToScalar(ctx context.Context, cfg *ExecutionConfig,
	st *ExecutionState, plan PlanStage) ([]values.SQLValue, error) {

	iter, err := plan.Open(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	row := &results.Row{}

	var valueRow []values.SQLValue
	if iter.Next(ctx, row) {
		// release this memory here... it will be re-allocated by a consuming
		// stage
		if err = cfg.memoryMonitor.Release(row.Data.Size()); err != nil {
			_ = iter.Close()
			return nil, err
		}

		// The full data copy here is unwanted.
		// This is a good place to attempt to improve performance.
		valueRow = make([]values.SQLValue, len(row.Data))
		for i, value := range row.Data {
			valueRow[i] = value.Data
		}
	} else {
		// MySQL specific behavior here.
		valueRow = make([]values.SQLValue, len(plan.Columns()))
		for i := range valueRow {
			valueRow[i] = values.NewSQLNull(cfg.sqlValueKind)
		}
	}

	// input must not have cardinality > 1
	if iter.Next(ctx, &results.Row{}) {
		_ = iter.Close()
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErSubqueryNoOneRow)
	}

	return valueRow, iter.Close()
}

// SQLSubqueryCmpExpr evaluates to true if the right subquery compares true to
// the left subquery by a provided comparison operator.
// The left and right subqueries need not be scalar but must produce only one
// row.
type SQLSubqueryCmpExpr struct {
	leftCorrelated  bool
	rightCorrelated bool
	leftPlan        PlanStage
	rightPlan       PlanStage
	operator        string
	// We always cache non-correlated subquery results in their entirety.
	// This cache is for the left-hand side.
	// SQLSubqueryCmpExpr's left cache is scalar but it can be multicolumn.
	leftCache []values.SQLValue
	// This cache is for the right-hand side.
	// SQLSubqueryCmpExpr's right cache is scalar but it can be multicolumn.
	rightCache []values.SQLValue
	// This cache is for the result. It is used if both sides are non-correlated.
	// This cache consists of a boolean.
	fullCache values.SQLBool
}

// Children returns a slice of all the Node children of the Node.
func (e SQLSubqueryCmpExpr) Children() []Node {
	return []Node{e.leftPlan, e.rightPlan}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (e *SQLSubqueryCmpExpr) ReplaceChild(i int, n Node) {
	switch i {
	case 0:
		e.leftPlan = panicIfNotPlanStage(e.ExprName(), n)
	case 1:
		e.rightPlan = panicIfNotPlanStage(e.ExprName(), n)
	default:
		panicWithInvalidIndex(e.ExprName(), i, 1)
	}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLSubqueryCmpExpr) ExprName() string {
	return "SQLSubqueryCmpExpr"
}

// NewSQLSubqueryCmpExpr is a constructor for SQLSubqueryCmpExpr.
func NewSQLSubqueryCmpExpr(
	leftCorrelated bool,
	rightCorrelated bool,
	leftPlan PlanStage,
	rightPlan PlanStage,
	operator string) *SQLSubqueryCmpExpr {
	return &SQLSubqueryCmpExpr{
		leftCorrelated:  leftCorrelated,
		rightCorrelated: rightCorrelated,
		leftPlan:        leftPlan,
		rightPlan:       rightPlan,
		operator:        operator,
	}
}

func (e *SQLSubqueryCmpExpr) reconcile() (SQLExpr, error) {
	leftPlan, rightPlan := reconcileSubqueryPlans(e.leftPlan, e.rightPlan)
	return NewSQLSubqueryCmpExpr(e.leftCorrelated, e.rightCorrelated, leftPlan, rightPlan, e.operator), nil
}

// Evaluate evaluates a SQLSubqueryCmpExpr into a SQLValue.
func (e *SQLSubqueryCmpExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(e)
	if err != nil {
		return nil, err
	}

	if !e.leftCorrelated && !e.rightCorrelated && e.fullCache != nil {
		return e.fullCache, nil
	}

	var leftRow []values.SQLValue
	var rightRow []values.SQLValue
	if e.leftCorrelated && !e.rightCorrelated {
		leftRow, err = evaluatePlanToScalar(ctx, cfg, st.SubqueryState(), e.leftPlan)
		if err != nil {
			return nil, err
		}
		if e.rightCache == nil {
			// Populate cache.
			e.rightCache, err = evaluatePlanToScalar(ctx, cfg, st, e.rightPlan)
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		rightRow = e.rightCache
	} else if e.rightCorrelated && !e.leftCorrelated {
		rightRow, err = evaluatePlanToScalar(ctx, cfg, st.SubqueryState(), e.rightPlan)
		if err != nil {
			return nil, err
		}
		if e.leftCache == nil {
			// Populate cache.
			e.leftCache, err = evaluatePlanToScalar(ctx, cfg, st, e.leftPlan)
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		leftRow = e.leftCache
		// Either both sides are correlated or neither are.
	} else {
		leftRow, err = evaluatePlanToScalar(ctx, cfg, st.SubqueryState(), e.leftPlan)
		if err != nil {
			return nil, err
		}
		rightRow, err = evaluatePlanToScalar(ctx, cfg, st.SubqueryState(), e.rightPlan)
		if err != nil {
			return nil, err
		}
	}

	// Make sure both subqueries return the same number of columns.
	if len(leftRow) != len(rightRow) {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, len(rightRow))
	}

	result, err := evaluateComparison(leftRow, rightRow, e.operator, cfg.sqlValueKind, st.collation)
	if err != nil {
		return nil, err
	}

	// Populate full cache.
	if !e.leftCorrelated && !e.rightCorrelated {
		e.fullCache = result.(values.SQLBool)
	}

	return result, nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLSubqueryCmpExpr.
func (e *SQLSubqueryCmpExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(e); err != nil {
		return nil, err
	}

	return e, nil
}

func (e *SQLSubqueryCmpExpr) String() string {
	return fmt.Sprintf("(%s) %s (%s)", PrettyPrintPlan(e.leftPlan),
		e.operator, PrettyPrintPlan(e.rightPlan))
}

// EvalType returns the EvalType associated with SQLSubqueryCmpExpr.
func (*SQLSubqueryCmpExpr) EvalType() types.EvalType {
	return types.EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLSubqueryCmpExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLSubqueryCmpExpr into something that can
// be used in an aggregation pipeline. If SQLSubqueryCmpExpr cannot be translated,
// it will return nil and error.
func (e *SQLSubqueryCmpExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	subPlanRight := e.rightPlan
	subPlanMsRight, ok := subPlanRight.(*MongoSourceStage)
	if !ok {
		return nil, innerSubqueryPushdownFailure(e)
	}

	if subPlanMsRight.LimitRowCount != 1 && !subPlanMsRight.IsDual() {
		return nil, multiRowSubqueryPushdownFailure(e)
	}

	subPlanLeft := e.leftPlan
	subPlanMsLeft, ok := subPlanLeft.(*MongoSourceStage)
	if !ok {
		return nil, innerSubqueryPushdownFailure(e)
	}

	if subPlanMsLeft.LimitRowCount != 1 && !subPlanMsLeft.IsDual() {
		return nil, multiRowSubqueryPushdownFailure(e)
	}

	err := t.addSubqueryCmpLookupStage(subPlanMsRight)
	if err != nil {
		return nil, wrapExprErrWithPushdownFailure(e, err)
	}

	err = t.addSubqueryCmpLookupStage(subPlanMsLeft)
	if err != nil {
		return nil, wrapExprErrWithPushdownFailure(e, err)
	}

	cmpOp, err := opFromSQLOpForSubqueryCmp(e.operator)
	if err != nil {
		return nil, wrapExprErrWithPushdownFailure(e, err)
	}

	rightCols := subPlanMsRight.Columns()
	leftCols := subPlanMsLeft.Columns()
	if len(rightCols) != len(leftCols) {
		// This should not be, this should have been caught during algebrization, so
		// if we get here something has gone wrong.
		panic("number of columns of subqueries do not match")
	}

	cmps := make([]ast.Expr, len(leftCols))

	for i := range leftCols {
		leftRef, err := lookupArrayRef(subPlanMsLeft, leftCols[i])
		if err != nil {
			return nil, wrapExprErrWithPushdownFailure(e, err)
		}

		rightRef, err := lookupArrayRef(subPlanMsRight, rightCols[i])
		if err != nil {
			return nil, wrapExprErrWithPushdownFailure(e, err)
		}

		leftSize := ast.NewFunction(bsonutil.OpSize, leftRef)
		rightSize := ast.NewFunction(bsonutil.OpSize, rightRef)
		sizeCheck := ast.NewBinary(bsonutil.OpEq, leftSize, rightSize)

		arrCmp := ast.NewBinary(cmpOp, leftRef, rightRef)
		cmps[i] = ast.NewBinary(bsonutil.OpAnd, sizeCheck, arrCmp)
	}

	return astutil.WrapInOp(bsonutil.OpAnd, cmps...), nil
}

// SQLSubqueryAllExpr evaluates to true if the left subquery expression compares true to
// all of the rows returned by the right subquery by a provided comparison operator.
// Multirow or multi column left subqueries are never valid.
type SQLSubqueryAllExpr struct {
	leftCorrelated  bool
	rightCorrelated bool
	leftPlan        PlanStage
	rightPlan       PlanStage
	operator        string
	// We always cache non-correlated subquery results in their entirety.
	// SQLSubqueryAllExpr can cache a row, which is being compared
	// to the value result of the right expression.
	leftCache []values.SQLValue
	// SQLSubqueryAllExpr can cache a whole table, with each row being compared
	// to the value result of the left expression.
	rightCache [][]values.SQLValue
}

// Children returns a slice of all the Node children of the Node.
func (e SQLSubqueryAllExpr) Children() []Node {
	return []Node{e.leftPlan, e.rightPlan}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (e *SQLSubqueryAllExpr) ReplaceChild(i int, n Node) {
	switch i {
	case 0:
		e.leftPlan = panicIfNotPlanStage(e.ExprName(), n)
	case 1:
		e.rightPlan = panicIfNotPlanStage(e.ExprName(), n)
	default:
		panicWithInvalidIndex(e.ExprName(), i, 1)
	}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLSubqueryAllExpr) ExprName() string {
	return "SQLSubqueryAllExpr"
}

// NewSQLSubqueryAllExpr is a constructor for SQLSubqueryAllExpr.
func NewSQLSubqueryAllExpr(
	leftCorrelated bool,
	rightCorrelated bool,
	leftPlan PlanStage,
	rightPlan PlanStage,
	operator string) *SQLSubqueryAllExpr {
	return &SQLSubqueryAllExpr{
		leftCorrelated:  leftCorrelated,
		rightCorrelated: rightCorrelated,
		leftPlan:        leftPlan,
		rightPlan:       rightPlan,
		operator:        operator,
	}
}

// Evaluate evaluates a SQLSubqueryAllExpr into a SQLValue.
func (e *SQLSubqueryAllExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(e)
	if err != nil {
		return nil, err
	}

	var leftRow []values.SQLValue
	if e.leftCorrelated {
		leftRow, err = evaluatePlanToScalar(ctx, cfg, st.SubqueryState(), e.leftPlan)
		if err != nil {
			return nil, err
		}
	} else {
		if e.leftCache == nil {
			// Populate cache.
			e.leftCache, err = evaluatePlanToScalar(ctx, cfg, st, e.leftPlan)
			if err != nil {
				return nil, err
			}
			err = cfg.memoryMonitor.AcquireGlobal(values.SQLValuesSize(e.leftCache))
			if err != nil {
				return nil, err
			}
		}

		// Read from cache.
		leftRow = e.leftCache
	}

	var rightTable [][]values.SQLValue
	if e.rightCorrelated {
		rightTable, err = evaluatePlan(ctx, cfg, st.SubqueryState(), e.rightPlan)
		if err != nil {
			return nil, err
		}
	} else {
		if e.rightCache == nil {
			// Populate cache.
			e.rightCache, err = evaluatePlan(ctx, cfg, st, e.rightPlan)
			if err != nil {
				return nil, err
			}
			err = cfg.memoryMonitor.AcquireGlobal(values.SQLValuesSize(e.rightCache...))
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		rightTable = e.rightCache
	}

	leftLen := len(leftRow)
	// <> ALL is rewritten in MySQL to NOT IN.
	// This is the only case when ALL will handle multi column expressions.
	if leftLen > 1 && e.operator != sqlOpNEQ {
		// https://dev.mysql.com/doc/mysql-reslimits-excerpt/5.7/en/subquery-restrictions.html
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, 1)
	}

	// Make sure the right subquery returns the same amount of columns as the left.
	if len(rightTable) > 0 && len(rightTable[0]) != leftLen {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, leftLen)
	}

	sawNull := false
	for _, rightRow := range rightTable {
		result, err := evaluateComparison(leftRow, rightRow, e.operator, cfg.sqlValueKind, st.collation)
		if err != nil {
			return nil, err
		}
		if !values.Bool(result) {
			return values.NewSQLBool(cfg.sqlValueKind, false), nil
		}
		if result.IsNull() {
			sawNull = true
		}
	}

	// left expression compared successfully to all rows in the right table
	if sawNull {
		return values.NewSQLNull(cfg.sqlValueKind), nil
	}
	return values.NewSQLBool(cfg.sqlValueKind, true), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLSubqueryAllExpr.
func (e *SQLSubqueryAllExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(e); err != nil {
		return nil, err
	}

	return e, nil
}

func (e *SQLSubqueryAllExpr) reconcile() (SQLExpr, error) {
	leftPlan, rightPlan := reconcileSubqueryPlans(e.leftPlan, e.rightPlan)
	return NewSQLSubqueryAllExpr(e.leftCorrelated, e.rightCorrelated, leftPlan, rightPlan, e.operator), nil
}

func (e *SQLSubqueryAllExpr) String() string {
	return fmt.Sprintf("%s\n%s all\n(%s)",
		PrettyPrintPlan(e.leftPlan), e.operator, PrettyPrintPlan(e.rightPlan))
}

// EvalType returns the EvalType associated with SQLSubqueryAllExpr.
func (*SQLSubqueryAllExpr) EvalType() types.EvalType {
	return types.EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLSubqueryAllExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLSubqueryAllExpr into something that can
// be used in an aggregation pipeline. If SQLSubqueryAllExpr cannot be translated,
// it will return nil and error.
func (e *SQLSubqueryAllExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	leftPlan := e.leftPlan
	rightPlan := e.rightPlan

	mapCmp, err := t.mapCmpForDoubleSubquery(e, leftPlan, rightPlan, e.operator)
	if err != nil {
		return nil, err
	}

	return astutil.WrapInOp(bsonutil.OpAllElementsTrue, mapCmp), nil
}

// SQLSubqueryAnyExpr evaluates to true if the left subquery expression compares true to
// any of the rows returned by the right subquery by a provided comparison operator.
// Multirow or multi column left subqueries are never valid.
type SQLSubqueryAnyExpr struct {
	leftCorrelated  bool
	rightCorrelated bool
	leftPlan        PlanStage
	rightPlan       PlanStage
	operator        string
	// We always cache non-correlated subquery results in their entirety.
	// SQLSubqueryAnyExpr can cache a row, which is being compared
	// to the value result of the right expression.
	leftCache []values.SQLValue
	// SQLSubqueryAnyExpr can cache a whole table, with each row being compared
	// to the value result of the left expression.
	rightCache [][]values.SQLValue
}

// Children returns a slice of all the Node children of the Node.
func (e SQLSubqueryAnyExpr) Children() []Node {
	return []Node{e.leftPlan, e.rightPlan}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (e *SQLSubqueryAnyExpr) ReplaceChild(i int, n Node) {
	switch i {
	case 0:
		e.leftPlan = panicIfNotPlanStage(e.ExprName(), n)
	case 1:
		e.rightPlan = panicIfNotPlanStage(e.ExprName(), n)
	default:
		panicWithInvalidIndex(e.ExprName(), i, 1)
	}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLSubqueryAnyExpr) ExprName() string {
	return "SQLSubqueryAnyExpr"
}

// NewSQLSubqueryAnyExpr is a constructor for SQLSubqueryAnyExpr.
func NewSQLSubqueryAnyExpr(
	leftCorrelated bool,
	rightCorrelated bool,
	leftPlan PlanStage,
	rightPlan PlanStage,
	operator string) *SQLSubqueryAnyExpr {
	return &SQLSubqueryAnyExpr{
		leftCorrelated:  leftCorrelated,
		rightCorrelated: rightCorrelated,
		leftPlan:        leftPlan,
		rightPlan:       rightPlan,
		operator:        operator,
	}
}

// Evaluate evaluates a SQLSubqueryAnyExpr into a values.SQLValue.
// ANY performs a series of comparisons. ANY uses the provided comparison operator.
// The resulting comparisons within columns of a row are ANDed together.
// Comparisons from separate rows are ORed together.
// Using SQL three-value boolean logic, the results are as follows:
// If a series of comparisons within any row is all true, the result is true.
// If not, if any of the series returns NULL (the series contains NULL and no falses),
// the result is NULL.
// Else, the result is false.
func (e *SQLSubqueryAnyExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(e)
	if err != nil {
		return nil, err
	}

	var leftRow []values.SQLValue
	if e.leftCorrelated {
		leftRow, err = evaluatePlanToScalar(ctx, cfg, st.SubqueryState(), e.leftPlan)
		if err != nil {
			return nil, err
		}
	} else {
		if e.leftCache == nil {
			// Populate cache.
			e.leftCache, err = evaluatePlanToScalar(ctx, cfg, st, e.leftPlan)
			if err != nil {
				return nil, err
			}
			err = cfg.memoryMonitor.AcquireGlobal(values.SQLValuesSize(e.leftCache))
			if err != nil {
				return nil, err
			}
		}

		// Read from cache.
		leftRow = e.leftCache
	}

	var rightTable [][]values.SQLValue
	if e.rightCorrelated {
		rightTable, err = evaluatePlan(ctx, cfg, st.SubqueryState(), e.rightPlan)
		if err != nil {
			return nil, err
		}
	} else {
		if e.rightCache == nil {
			// Populate cache.
			e.rightCache, err = evaluatePlan(ctx, cfg, st, e.rightPlan)
			if err != nil {
				return nil, err
			}
			err = cfg.memoryMonitor.AcquireGlobal(values.SQLValuesSize(e.rightCache...))
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		rightTable = e.rightCache
	}

	leftLen := len(leftRow)
	// = ANY is rewritten in MySQL to IN.
	// This is the only case when ANY will handle multi column expressions.
	if leftLen > 1 && e.operator != sqlOpEQ {
		// https://dev.mysql.com/doc/mysql-reslimits-excerpt/5.7/en/subquery-restrictions.html
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, 1)
	}

	// Make sure the subquery returns the same amount of columns as the left subquery.
	if len(rightTable) > 0 && len(rightTable[0]) != leftLen {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, leftLen)
	}

	sawNull := false
	for _, rightRow := range rightTable {
		result, err := evaluateComparison(leftRow, rightRow, e.operator, cfg.sqlValueKind, st.collation)
		if err != nil {
			return nil, err
		}
		if values.Bool(result) {
			return values.NewSQLBool(cfg.sqlValueKind, true), nil
		}
		if result.IsNull() {
			sawNull = true
		}
	}

	// The left expression did not compare successfully to any row in the right table.
	if sawNull {
		return values.NewSQLNull(cfg.sqlValueKind), nil
	}
	return values.NewSQLBool(cfg.sqlValueKind, false), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLSubqueryAnyExpr.
func (e *SQLSubqueryAnyExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(e); err != nil {
		return nil, err
	}

	return e, nil
}

func (e *SQLSubqueryAnyExpr) reconcile() (SQLExpr, error) {
	leftPlan, rightPlan := reconcileSubqueryPlans(e.leftPlan, e.rightPlan)
	return NewSQLSubqueryAnyExpr(e.leftCorrelated, e.rightCorrelated, leftPlan, rightPlan, e.operator), nil
}

func (e *SQLSubqueryAnyExpr) String() string {
	return fmt.Sprintf("%s\n%s any\n(%s)",
		PrettyPrintPlan(e.leftPlan), e.operator, PrettyPrintPlan(e.rightPlan))
}

// EvalType returns the EvalType associated with SQLSubqueryAnyExpr.
func (*SQLSubqueryAnyExpr) EvalType() types.EvalType {
	return types.EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLSubqueryAnyExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLSubqueryAnyExpr into something that can
// be used in an aggregation pipeline. If SQLSubqueryAnyExpr cannot be translated,
// it will return nil and error.
func (e *SQLSubqueryAnyExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	leftPlan := e.leftPlan
	rightPlan := e.rightPlan

	mapCmp, err := t.mapCmpForDoubleSubquery(e, leftPlan, rightPlan, e.operator)
	if err != nil {
		return nil, err
	}

	return astutil.WrapInOp(bsonutil.OpAnyElementTrue, mapCmp), nil
}

// SQLSubqueryExpr is a wrapper around a parser.SelectStatement representing a subquery
// outside of an EXISTS expression. A SQLSubqueryExpr always evaluates to a single-column
// scalar.
type SQLSubqueryExpr struct {
	correlated bool
	allowRows  bool
	plan       PlanStage
	// We always cache non-correlated subquery results in their entirety.
	// This is a fine place to be more clever in the future.
	// SQLSubqueryExpr caches a single-column scalar.
	cache values.SQLValue
}

// Children returns a slice of all the Node children of the Node.
func (e SQLSubqueryExpr) Children() []Node {
	return []Node{e.plan}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (e *SQLSubqueryExpr) ReplaceChild(i int, n Node) {
	switch i {
	case 0:
		e.plan = panicIfNotPlanStage(e.ExprName(), n)
	default:
		panicWithInvalidIndex(e.ExprName(), i, 0)
	}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLSubqueryExpr) ExprName() string {
	return "SQLSubqueryExpr"
}

// NewSQLSubqueryExpr is a constructor for SQLSubqueryExpr.
func NewSQLSubqueryExpr(correlated, allowRows bool, plan PlanStage) *SQLSubqueryExpr {
	return &SQLSubqueryExpr{
		correlated: correlated,
		allowRows:  allowRows,
		plan:       plan,
	}
}

// nolint: unparam
func (e *SQLSubqueryExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

// ToAggregationLanguage translates SQLSubqueryExpr into something that can
// be used in an aggregation pipeline. If SQLSubqueryExpr cannot be translated,
// it will return nil and error.
func (e *SQLSubqueryExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	if e.correlated {
		return nil, newPushdownFailure(
			e.ExprName(),
			"cannot push down correlated subqueries",
		)
	}

	return nil, newPushdownFailure(e.ExprName(), "cannot push down subquery")
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLSubqueryExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

func (e *SQLSubqueryExpr) evaluateFromPlan(ctx context.Context,
	cfg *ExecutionConfig, st *ExecutionState, plan PlanStage) (values.SQLValue, error) {
	var err error
	var iter RowIter
	defer func() {
		if iter != nil {
			if err == nil {
				err = iter.Close()
			} else {
				// If we already have an err, keep the previous err rather
				// than getting a new one.
				_ = iter.Close()
			}

			if err == nil {
				err = iter.Err()
			}
		}
	}()

	iter, err = plan.Open(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	row := &results.Row{}
	if iter.Next(ctx, row) {

		// release this memory here... it will be re-allocated by a consuming stage
		if err = cfg.memoryMonitor.Release(row.Data.Size()); err != nil {
			_ = iter.Close()
			return nil, err
		}

		// Filter has to check the entire source to return an accurate 'hasNext'
		if iter.Next(ctx, &results.Row{}) {
			_ = iter.Close()
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErSubqueryNoOneRow)
		}
	}

	switch len(row.Data) {
	case 0:
		return values.NewSQLNull(cfg.sqlValueKind), iter.Close()
	case 1:
		return row.Data[0].Data, iter.Close()
	default:
		panic(fmt.Sprintf("SQLSubqueryExpr must evaluate to a single column scalar, instead got %d columns", len(row.Data)))
	}
}

// Evaluate evaluates a SQLSubqueryExpr into a values.SQLValue.
func (e *SQLSubqueryExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(e)
	if err != nil {
		return nil, err
	}

	if e.correlated {
		return e.evaluateFromPlan(ctx, cfg, st.SubqueryState(), e.plan)
	}

	if e.cache == nil {
		// Populate cache.
		e.cache, err = e.evaluateFromPlan(ctx, cfg, st, e.plan)
		if err != nil {
			return nil, err
		}
	}
	// Read from cache.
	return e.cache, nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLSubqueryExpr.
func (e *SQLSubqueryExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(e); err != nil {
		return nil, err
	}

	return e, nil
}

// Exprs returns all the SQLColumnExprs associated with the columns of SQLSubqueryExpr.
func (e *SQLSubqueryExpr) Exprs() []SQLExpr {
	exprs := []SQLExpr{}
	for _, c := range e.plan.Columns() {
		exprs = append(exprs, NewSQLColumnExpr(c.SelectID,
			c.Database, c.Table, c.Name, c.EvalType, c.MongoType))
	}

	return exprs
}

func (e *SQLSubqueryExpr) String() string {
	return PrettyPrintPlan(e.plan)
}

// EvalType returns the EvalType associated with SQLSubqueryExpr.
func (e *SQLSubqueryExpr) EvalType() types.EvalType {
	columns := e.plan.Columns()
	if len(columns) == 1 {
		return columns[0].EvalType
	}

	panic(fmt.Sprintf("SQLSubqueryExpr must evaluate to a single column scalar, instead got %d columns", len(columns)))
}

// SQLValueExpr represents a literal SQLValue in a SQLExpr tree.
type SQLValueExpr struct {
	Value values.SQLValue
}

// NewSQLValueExpr is a constructor for SQLValueExpr.
func NewSQLValueExpr(value values.SQLValue) SQLValueExpr {
	return SQLValueExpr{
		Value: value,
	}
}

// Children returns a slice of all the Node children of the Node.
func (SQLValueExpr) Children() []Node {
	return []Node{}
}

// ExprName returns a string representing this SQLExpr's name.
func (e SQLValueExpr) ExprName() string {
	return fmt.Sprintf("SQLValueExpr(%s)", e.Value.String())
}

// Evaluate evaluates a SQLValueExpr into a values.SQLValue.
func (e SQLValueExpr) Evaluate(_ context.Context, cfg *ExecutionConfig, _ *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(e)
	if err != nil {
		return nil, err
	}

	return e.Value, nil
}

// nolint: unparam
func (e SQLValueExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLValueExpr..
// Because variable assignments (even to globals) are not allowed to change during a query,
// it can be constant folded as its value.
func (e SQLValueExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(e); err != nil {
		return nil, err
	}

	return e, nil
}

func (e SQLValueExpr) String() string {
	return e.Value.String()
}

// EvalType returns the EvalType associated with SQLValueExpr.
func (e SQLValueExpr) EvalType() types.EvalType {
	return e.Value.EvalType()
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (e SQLValueExpr) ReplaceChild(i int, n Node) {
	panicWithInvalidIndex(e.ExprName(), i, -1)
}

func (e SQLValueExpr) to34PlusAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	switch e.EvalType() {
	// These types do not require $literal wrapping. We avoid it for the slight performance
	// and readability benefit.
	case types.EvalDate, types.EvalDatetime, types.EvalString, types.EvalObjectID, types.EvalNull:
		v, err := e.Value.BSONValue()
		if err != nil {
			return nil, newPushdownFailure(
				e.ExprName(),
				"failed to get value from SQLValue",
				"error", fmt.Sprintf("%v", err),
			)
		}

		return ast.NewUnknown(v), nil
	}
	return t.getValue(e)
}

// ToAggregationLanguage translates SQLValueExpr into something that can
// be used in an aggregation pipeline. If SQLValueExpr cannot be translated,
// it will return nil and error.
func (e SQLValueExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	if t.versionAtLeast(3, 4, 0) {
		return e.to34PlusAggregationLanguage(t)
	}

	// On server 3.2.0, everything needs to be wrapped in literal, even null, when
	// used as the top level in a project specification. This call returns an
	// ast.Constant, which is always unconditionally wrapped in $literal.
	return t.getValue(e)
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e SQLValueExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// SQLVariableExpr represents a variable lookup.
type SQLVariableExpr struct {
	Name  string
	Kind  variable.Kind
	Scope variable.Scope
	Value values.SQLValue
}

// NewSQLVariableExpr is a constructor for SQLVariableExpr.
func NewSQLVariableExpr(name string, kind variable.Kind, scope variable.Scope, value values.SQLValue) *SQLVariableExpr {
	return &SQLVariableExpr{
		Name:  name,
		Kind:  kind,
		Scope: scope,
		Value: value,
	}
}

// Children returns a slice of all the Node children of the Node.
func (*SQLVariableExpr) Children() []Node {
	return []Node{}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLVariableExpr) ExprName() string {
	return "SQLVariableExpr"
}

// Evaluate evaluates a SQLVariableExpr into a values.SQLValue.
func (e *SQLVariableExpr) Evaluate(_ context.Context, cfg *ExecutionConfig, _ *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(e)
	if err != nil {
		return nil, err
	}

	// e.Value can be nil: if this variable has never been set before.
	if e.Value == nil {
		return values.NewSQLNull(cfg.sqlValueKind), nil
	}
	return e.Value.CloneWithKind(cfg.sqlValueKind), nil
}

// nolint: unparam
func (e *SQLVariableExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLVariableExpr..
// Because variable assignments (even to globals) are not allowed to change during a query,
// it can be constant folded as its value.
func (e *SQLVariableExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(e); err != nil {
		return nil, err
	}
	// e.Value can be nil: if this variable has never been set before.
	if e.Value == nil {
		return NewSQLValueExpr(values.NewSQLNull(cfg.sqlValueKind)), nil
	}
	return NewSQLValueExpr(e.Value.CloneWithKind(cfg.sqlValueKind)), nil
}

func (e *SQLVariableExpr) String() string {
	prefix := ""
	switch e.Kind {
	case variable.UserKind:
		prefix = "@"
	default:
		switch e.Scope {
		case variable.GlobalScope:
			prefix = "@@global."
		case variable.SessionScope:
			prefix = "@@session."
		}
	}

	return prefix + e.Name
}

// EvalType returns the EvalType associated with SQLVariableExpr.
func (e *SQLVariableExpr) EvalType() types.EvalType {
	// e.Value will be nil during the assignment to a user variable,
	// as it has no value at that time. Since it has no value, it is semantically
	// correct for the EvalType to be EvalPolymorphic.
	if e.Value == nil {
		return types.EvalPolymorphic
	}
	return e.Value.EvalType()
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (e *SQLVariableExpr) ReplaceChild(i int, n Node) {
	panicWithInvalidIndex(e.ExprName(), i, -1)
}

// ToAggregationLanguage translates SQLVariableExpr into something that can
// be used in an aggregation pipeline. If SQLVariableExpr cannot be translated,
// it will return nil and error.
func (e *SQLVariableExpr) ToAggregationLanguage(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	// e.Value can be nil: if this variable has never been set before.
	if e.Value == nil {
		return NewSQLValueExpr(values.NewSQLNull(values.VariableSQLValueKind)).ToAggregationLanguage(t)
	}
	return NewSQLValueExpr(e.Value).ToAggregationLanguage(t)
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLVariableExpr) ToAggregationPredicate(t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// caseCondition holds a matcher used in evaluating case expressions and
// a value to return if a particular case is matched. If a case is matched,
// the corresponding 'then' value is evaluated and returned ('then'
// corresponds to the 'then' clause in a case expression).
type caseCondition struct {
	matcher SQLExpr
	then    SQLExpr
}

func newCaseCondition(matcher, then SQLExpr) caseCondition {
	return caseCondition{
		matcher: matcher,
		then:    then,
	}
}

func (c *caseCondition) String() string {
	return fmt.Sprintf("when (%v) then %v", c.matcher, c.then)
}

// reconcileSubqueryPlans reconciles the left and right PlanStages'
// projectedColumns with each other in a pairwise manner.
func reconcileSubqueryPlans(left, right PlanStage) (PlanStage, PlanStage) {
	leftPlan := panicIfNotProjectStage("left", left)
	rightPlan := panicIfNotProjectStage("right", right)

	leftColumns := make([]ProjectedColumn, len(leftPlan.projectedColumns))
	rightColumns := make([]ProjectedColumn, len(rightPlan.projectedColumns))

	// This should return an error during algebrization, so this should never evaluate to true.
	if len(leftColumns) != len(rightColumns) {
		panic("left and right columns of different lengths")
	}

	for i, lc := range leftPlan.projectedColumns {
		rc := rightPlan.projectedColumns[i]
		newLeft, newRight := ReconcileSQLExprs(lc.Expr, rc.Expr)

		leftColumns[i] = ProjectedColumn{lc.Column.Clone(), newLeft}
		rightColumns[i] = ProjectedColumn{rc.Column.Clone(), newRight}
	}

	newLeftPlan := NewProjectStage(leftPlan.source.clone(), leftColumns...)
	newRightPlan := NewProjectStage(rightPlan.source.clone(), rightColumns...)

	return newLeftPlan, newRightPlan
}

// validateArgs ensures that the expr's arguments are valid for evaluation
// (i.e. they have the same type as they do when the SQLExpr is reconciled).
// If validation fails, an error is returned.
func validateArgs(expr SQLExpr) error {
	children := expr.Children()

	// SQLExprs with PlanStages as children instead of SQLExprs as children
	// require different handling.
	if hasAllPlanStageChildren(expr) {
		// If there are two PlanStage Children (for example, SQLSubqueryCmpExpr),
		// validation is delegated to a different helper function.
		if len(children) == 2 {
			return validateSubqueryPlans(children[0].(PlanStage), children[1].(PlanStage))
		}

		// This case is for SQLExprs with just one child (which is a PlanStage).
		// Currently, the only two like this are SQLExistsExpr and SQLSubqueryExpr.
		// SQLExprs of this form are always unconditionally valid.
		return nil
	}

	preReconciliationChildren := nodesToExprs(children)

	reconciled, err := expr.reconcile()
	if err != nil {
		return err
	}

	postReconciliationChildren := nodesToExprs(reconciled.Children())

	for i, pre := range preReconciliationChildren {
		post := postReconciliationChildren[i]
		if !isSimilar(pre.EvalType(), post.EvalType()) {
			return fmt.Errorf("expected EvalType %x at index %d, but got %x",
				post.EvalType(), i, pre.EvalType())
		}
	}

	return nil
}

// validateSubqueryPlans ensures that the left and right PlanStages have
// similar eval types for corresponding columns. If any pair does not have
// similar types, an error is returned.
func validateSubqueryPlans(left, right PlanStage) error {
	rightColumns := right.Columns()

	for i, lc := range left.Columns() {
		if !isSimilar(lc.EvalType, rightColumns[i].EvalType) {
			return fmt.Errorf("expected EvalType %x at index %d, but got %x",
				lc.EvalType, i, rightColumns[i].EvalType)
		}
	}

	return nil
}

// hasAllPlanStageChildren returns true if the expr's Children are all PlanStages.
func hasAllPlanStageChildren(expr SQLExpr) bool {
	all := true
	for _, c := range expr.Children() {
		_, isPlanStage := c.(PlanStage)
		all = all && isPlanStage
	}
	return all
}
