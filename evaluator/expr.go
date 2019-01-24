package evaluator

import (
	"context"
	"fmt"
	"regexp"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/evaluator/variable"
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
	// Evaluate evaluates the receiver expression in memory.
	Evaluate(context.Context, *ExecutionConfig, *ExecutionState) (SQLValue, error)
	// FoldConstants performs constant-folding on this SQLExpr, returning a
	// SQLExpr that is simplified as much as possible.
	FoldConstants(cfg *OptimizerConfig) SQLExpr
	// String renders a string representation of the receiver expression.
	String() string
	// EvalType returns the EvalType resulting from evaluating the expression
	// (for instance, SQLEqualsExpr.EvalType() returns EvalBoolean).
	EvalType() EvalType
	// ExprName returns a string representing this SQLExpr's name.
	ExprName() string
	// ToAggregationPredicate translates this expression to the aggregation language
	// to be evaluated as a predicate directly in a $match stage via $expr.
	ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure)
	// ToAggregationLanguage translates a SQLExpr to a MongoDB aggregation expression.
	ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure)
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
	query  bson.M
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

// Evaluate evaluates a MongoFilterExpr into a SQLValue.
func (e *MongoFilterExpr) Evaluate(context.Context, *ExecutionConfig, *ExecutionState) (SQLValue, error) {
	return nil, fmt.Errorf("could not evaluate predicate with mongo filter expression")
}

// nolint: unparam
func (e *MongoFilterExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *MongoFilterExpr.
func (e *MongoFilterExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return e
}

func (e *MongoFilterExpr) String() string {
	return fmt.Sprintf("%v=%v", e.column.String(), e.expr.String())
}

// ToMatchLanguage translates MongoFilterExpr into something that can
// be used in an match expression. If MongoFilterExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original MongoFilterExpr.
func (e *MongoFilterExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	return e.query, nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *MongoFilterExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates MongoFilterExpr into something that can
// be used in an aggregation pipeline. If MongoFilterExpr cannot be translated,
// it will return nil and error.
func (e *MongoFilterExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(e)
}

// EvalType returns the EvalType associated with MongoFilterExpr.
func (*MongoFilterExpr) EvalType() EvalType {
	return EvalBoolean
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

// Evaluate evaluates a SQLAssignmentExpr into a SQLValue.
func (e *SQLAssignmentExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	value, err := e.expr.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	return value, nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLAssignmentExpr.
func (e *SQLAssignmentExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return e
}

// nolint: unparam
func (e *SQLAssignmentExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

func (e *SQLAssignmentExpr) String() string {
	return fmt.Sprintf("%s := %s", e.variable.String(), e.expr.String())
}

// EvalType returns the EvalType associated with SQLAssignmentExpr.
func (e *SQLAssignmentExpr) EvalType() EvalType {
	return e.expr.EvalType()
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLAssignmentExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLAssignmentExpr into something that can
// be used in an aggregation pipeline. If SQLAssignmentExpr cannot be translated,
// it will return nil and error.
func (e *SQLAssignmentExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
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

// Evaluate evaluates a SQLBenchmarkExpr into a SQLValue.
func (e *SQLBenchmarkExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {

	count, err := e.count.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	for i := int64(0); i < Int64(count); i++ {
		_, err := e.expr.Evaluate(ctx, cfg, st)
		if err != nil {
			return nil, err
		}
	}

	return NewSQLInt64(cfg.sqlValueKind, 0), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for SQLBenchmarkExpr.
func (e *SQLBenchmarkExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return e
}

// nolint: unparam
func (e *SQLBenchmarkExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

func (e *SQLBenchmarkExpr) String() string {
	return fmt.Sprintf("benchmark(%s, %s)", e.count.String(), e.expr.String())
}

// EvalType returns the EvalType associated with SQLBenchmarkExpr.
func (e *SQLBenchmarkExpr) EvalType() EvalType {
	return EvalInt64
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLBenchmarkExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLBenchmarkExpr into something that can
// be used in an aggregation pipeline. If SQLBenchmarkExpr cannot be translated,
// it will return nil and error.
func (e *SQLBenchmarkExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
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

// Evaluate evaluates a SQLCaseExpr into a SQLValue.
func (e *SQLCaseExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	for _, condition := range e.caseConditions {
		result, err := condition.matcher.Evaluate(ctx, cfg, st)
		if err != nil {
			return nil, err
		}

		if Bool(result) {
			return condition.then.Evaluate(ctx, cfg, st)
		}
	}

	return e.elseValue.Evaluate(ctx, cfg, st)
}

// FoldConstants simplifies expressions containing constants when it is able to for SQLCaseExpr.
func (e *SQLCaseExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	newCases := make([]caseCondition, 0)
	for _, caseCond := range e.caseConditions {
		if matchVal, ok := caseCond.matcher.(SQLValue); ok {
			// If the matchVal is Falsy, we want to remove
			// it from the caseConditions.
			if Bool(matchVal) {
				newCases = append(newCases, newCaseCondition(matchVal, caseCond.then))
			}
		} else {
			newCases = append(newCases, newCaseCondition(caseCond.matcher, caseCond.then))
		}
	}
	if len(newCases) == 0 {
		return e.elseValue
	}
	// If caseConditions[0].match is a SQLValue, it must be true,
	// as we have removed all false SQLValues, in such a case,
	// return the value of the case. If it is not a SQLValue,
	// we cannot simplify any further because it must contain
	// a column value.
	if _, ok := newCases[0].matcher.(SQLValue); ok {
		return newCases[0].then
	}
	e.caseConditions = newCases
	return e
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
func (e *SQLCaseExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	elseValue, err := t.ToAggregationLanguage(e.elseValue)
	if err != nil {
		return nil, err
	}

	var conditions []interface{}
	var thens []interface{}
	for _, condition := range e.caseConditions {
		var c interface{}
		if matcher, ok := condition.matcher.(*SQLEqualsExpr); ok {
			newMatcher := NewSQLOrExpr(
				matcher,
				NewSQLEqualsExpr(matcher.left, NewSQLBool(t.valueKind(), true)))
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

		conditions = append(conditions, c)
		thens = append(thens, then)
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
		cases = bsonutil.WrapInCond(thens[i], cases, conditions[i])
	}

	return cases, nil

}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLCaseExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// EvalType returns the EvalType associated with SQLCaseExpr.
func (e *SQLCaseExpr) EvalType() EvalType {
	conds := []SQLExpr{e.elseValue}
	for _, cond := range e.caseConditions {
		conds = append(conds, cond.then)
	}
	// Verified that Case expressions in MySQL use
	// VarcharHighPriority.
	s := &EvalTypeSorter{VarcharHighPriority: true}
	return preferentialTypeWithSorter(s, conds...)
}

// SQLColumnExpr represents a column reference.
type SQLColumnExpr struct {
	selectID     int
	databaseName string
	tableName    string
	columnName   string
	columnType   ColumnType
	correlated   bool
}

// NewSQLColumnExpr creates a new SQLColumnExpr with its required fields.
// NewSQLColumnExpr is a constructor for SQLColumnExpr.
func NewSQLColumnExpr(selectID int, databaseName, tableName, columnName string, evalType EvalType, mongoType schema.MongoType) SQLColumnExpr {
	return SQLColumnExpr{
		selectID:     selectID,
		databaseName: databaseName,
		tableName:    tableName,
		columnName:   columnName,
		columnType: NewColumnType(
			evalType,
			mongoType,
		),
	}
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

// Evaluate evaluates a SQLColumnExpr into a SQLValue.
func (e SQLColumnExpr) Evaluate(_ context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {

	for _, row := range st.rows {
		if value, ok := row.GetField(e.selectID, e.databaseName, e.tableName, e.columnName); ok {
			return ConvertTo(value, e.EvalType()), nil
		}
	}

	for _, row := range st.correlatedRows {
		if value, ok := row.GetField(e.selectID, e.databaseName, e.tableName, e.columnName); ok {
			return ConvertTo(value, e.EvalType()), nil
		}
	}

	panic(fmt.Sprintf("cannot find column %q", e))
}

// FoldConstants simplifies expressions containing constants when it is able to for SQLColumnExpr.
func (e SQLColumnExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return e
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
func (e SQLColumnExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	if e.correlated {
		cc := t.addCorrelatedSubqueryColumnFuture(&e)
		return bsonutil.WrapInLiteral(cc), nil
	}

	name, ok := t.LookupFieldName(e.databaseName, e.tableName, e.columnName)
	if !ok {
		return nil, newPushdownFailure(
			e.ExprName(),
			"failed to find field name",
			"expr", e.String(),
		)
	}

	return getProjectedFieldName(name, e.columnType.EvalType), nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e SQLColumnExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToMatchLanguage translates SQLColumnExpr into something that can
// be used in an match expression. If SQLColumnExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLColumnExpr.
func (e SQLColumnExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	if e.correlated {
		cc := t.addCorrelatedSubqueryColumnFuture(&e)
		return bsonutil.WrapInLiteral(cc), nil
	}
	name, ok := t.LookupFieldName(e.databaseName, e.tableName, e.columnName)
	if !ok {
		return nil, e
	}

	if e.EvalType() != EvalBoolean {
		return bsonutil.NewM(
			bsonutil.NewDocElem(name, bsonutil.NewM(
				bsonutil.NewDocElem(bsonutil.OpNeq, nil),
			)),
		), e
	}

	return bsonutil.NewM(
		bsonutil.NewDocElem(bsonutil.OpAnd, bsonutil.NewArray(
			bsonutil.NewM(
				bsonutil.NewDocElem(name, bsonutil.NewM(
					bsonutil.NewDocElem(bsonutil.OpNeq, false),
				)),
			),
			bsonutil.NewM(
				bsonutil.NewDocElem(name, bsonutil.NewM(
					bsonutil.NewDocElem(bsonutil.OpNeq, nil),
				)),
			),
			bsonutil.NewM(
				bsonutil.NewDocElem(name, bsonutil.NewM(
					bsonutil.NewDocElem(bsonutil.OpNeq, 0),
				)),
			),
		)),
	), nil
}

// EvalType returns the EvalType associated with SQLColumnExpr.
func (e SQLColumnExpr) EvalType() EvalType {
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
	targetType EvalType
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
func NewSQLConvertExpr(expr SQLExpr, targetType EvalType) *SQLConvertExpr {
	return &SQLConvertExpr{
		expr:       expr,
		targetType: targetType,
	}
}

// Evaluate evaluates a SQLConvertExpr into a SQLValue.
func (e *SQLConvertExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	v, err := e.expr.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	return ConvertTo(v, e.targetType), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLConvertExpr.
func (e *SQLConvertExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	if exprVal, ok := e.expr.(SQLValue); ok {
		out := ConvertTo(exprVal, e.targetType)
		return out
	}
	return e
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
	prettyTypeName := string(EvalTypeToSQLType(e.targetType))
	return "Convert(" + e.expr.String() + ", " + prettyTypeName + ")"
}

// EvalType returns the EvalType associated with SQLConvertExpr.
func (e *SQLConvertExpr) EvalType() EvalType {
	return e.targetType
}

func (e *SQLConvertExpr) translateMongoSQL(t *PushdownTranslator) (interface{}, PushdownFailure) {
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

func (e *SQLConvertExpr) translateMySQL(t *PushdownTranslator) (interface{}, PushdownFailure) {
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
	case EvalInt32, EvalInt64,
		EvalUint32, EvalUint64,
		EvalDecimal128, EvalBoolean:

		switch toType {
		case EvalInt32, EvalInt64,
			EvalUint32, EvalUint64,
			EvalDecimal128, EvalDouble,
			EvalString, EvalBoolean:
			return e.translateMongoSQL(t)
		}

	case EvalDouble:
		switch toType {
		case EvalInt32, EvalInt64,
			EvalUint32, EvalUint64,
			EvalDecimal128, EvalBoolean:
			return e.translateMongoSQL(t)
		}

	case EvalDatetime:
		year := bsonutil.NewM(bsonutil.NewDocElem("$year", expr))
		month := bsonutil.NewM(bsonutil.NewDocElem("$month", expr))
		day := bsonutil.NewM(bsonutil.NewDocElem("$dayOfMonth", expr))
		hour := bsonutil.NewM(bsonutil.NewDocElem("$hour", expr))
		minute := bsonutil.NewM(bsonutil.NewDocElem("$minute", expr))
		second := bsonutil.NewM(bsonutil.NewDocElem("$second", expr))
		millisecond := bsonutil.NewM(bsonutil.NewDocElem("$millisecond", expr))

		switch toType {
		case EvalDate:
			asDate := bsonutil.NewM(bsonutil.NewDocElem("$dateFromParts", bsonutil.NewM(
				bsonutil.NewDocElem("year", year),
				bsonutil.NewDocElem("month", month),
				bsonutil.NewDocElem("day", day),
			)),
			)

			return asDate, nil

		case EvalInt32, EvalInt64, EvalUint32, EvalUint64:
			asNum := bsonutil.NewM(
				bsonutil.NewDocElem("$add", bsonutil.NewArray(
					second,
					bsonutil.NewM(bsonutil.NewDocElem("$multiply", bsonutil.NewArray(
						minute,
						100,
					))),
					bsonutil.NewM(bsonutil.NewDocElem("$multiply", bsonutil.NewArray(
						hour,
						10000,
					))),
					bsonutil.NewM(bsonutil.NewDocElem("$multiply", bsonutil.NewArray(
						day,
						1000000,
					))),
					bsonutil.NewM(bsonutil.NewDocElem("$multiply", bsonutil.NewArray(
						month,
						100000000,
					))),
					bsonutil.NewM(bsonutil.NewDocElem("$multiply", bsonutil.NewArray(
						year,
						10000000000,
					))),
				)),
			)

			return asNum, nil

		case EvalDecimal128, EvalDouble:
			asNum := bsonutil.NewM(bsonutil.NewDocElem("$add", bsonutil.NewArray(
				bsonutil.NewM(bsonutil.NewDocElem("$divide", bsonutil.NewArray(
					millisecond,
					1000,
				))),
				second,
				bsonutil.NewM(bsonutil.NewDocElem("$multiply", bsonutil.NewArray(
					minute,
					100,
				))),
				bsonutil.NewM(bsonutil.NewDocElem("$multiply", bsonutil.NewArray(
					hour,
					10000,
				))),
				bsonutil.NewM(bsonutil.NewDocElem("$multiply", bsonutil.NewArray(
					day,
					1000000,
				))),
				bsonutil.NewM(bsonutil.NewDocElem("$multiply", bsonutil.NewArray(
					month,
					100000000,
				))),
				bsonutil.NewM(bsonutil.NewDocElem("$multiply", bsonutil.NewArray(
					year,
					10000000000,
				))),
			)),
			)

			return asNum, nil

		case EvalString:
			asString := bsonutil.NewM(bsonutil.NewDocElem("$dateToString", bsonutil.NewM(
				bsonutil.NewDocElem("date", expr),
				bsonutil.NewDocElem("format", "%Y-%m-%d %H:%M:%S.%L000"),
			)),
			)

			return asString, nil

		}

	case EvalDate:
		year := bsonutil.NewM(bsonutil.NewDocElem("$year", expr))
		month := bsonutil.NewM(bsonutil.NewDocElem("$month", expr))
		day := bsonutil.NewM(bsonutil.NewDocElem("$dayOfMonth", expr))

		switch toType {
		case EvalDatetime:
			asDate := bsonutil.NewM(bsonutil.NewDocElem("$dateFromParts", bsonutil.NewM(
				bsonutil.NewDocElem("year", year),
				bsonutil.NewDocElem("month", month),
				bsonutil.NewDocElem("day", day),
			)),
			)

			return asDate, nil

		case EvalInt32, EvalInt64,
			EvalUint32, EvalUint64,
			EvalDecimal128, EvalDouble:

			asNum := bsonutil.NewM(
				bsonutil.NewDocElem("$add", bsonutil.NewArray(
					day,
					bsonutil.NewM(bsonutil.NewDocElem("$multiply", bsonutil.NewArray(
						month,
						100,
					))),
					bsonutil.NewM(bsonutil.NewDocElem("$multiply", bsonutil.NewArray(
						year,
						10000,
					))),
				)),
			)

			return asNum, nil

		case EvalString:
			asString := bsonutil.NewM(bsonutil.NewDocElem("$dateToString", bsonutil.NewM(
				bsonutil.NewDocElem("date", expr),
				bsonutil.NewDocElem("format", "%Y-%m-%d"),
			)),
			)

			return asString, nil

		}

	case EvalObjectID:
		switch toType {
		case EvalString:
			return e.translateMongoSQL(t)
		}

	default:
		// mysql-mode pushdown not yet implemented for conversions from other types
	}

	return nil, newPushdownFailure(
		e.ExprName(),
		fmt.Sprintf(
			"cannot push down mysql-mode conversion from type '%s'",
			EvalTypeToMongoType(fromType),
		),
	)
}

// ToAggregationLanguage translates SQLConvertExpr into something that can
// be used in an aggregation pipeline. At the moment, SQLConvertExpr cannot be
// translated, so this function will always return nil and error.
func (e *SQLConvertExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	if e.targetType == EvalObjectID {
		sv, ok := e.expr.(SQLVarchar)
		if !ok {
			return nil, newPushdownFailure(
				e.ExprName(),
				"can only push down SQLVarchar as ObjectId",
			)
		}
		return sv.SQLObjectID().ToAggregationLanguage(t)
	}

	mode := t.Cfg.sqlValueKind
	switch mode {
	case MySQLValueKind:
		return e.translateMySQL(t)
	case MongoSQLValueKind:
		return e.translateMongoSQL(t)
	default:
		panic(fmt.Errorf("impossible value %v for cfg.sqlValueKind", mode))
	}
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLConvertExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return e.ToAggregationLanguage(t)
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
	cache SQLBool
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

func (*SQLExistsExpr) evaluateFromPlan(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, plan PlanStage) (SQLBool, error) {
	iter, err := plan.Open(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	row := &Row{}

	hasNext := iter.Next(ctx, row)
	// release this memory here... it will be re-allocated by a consuming
	// stage
	if err = cfg.memoryMonitor.Release(row.Data.Size()); err != nil {
		_ = iter.Close()
		return nil, err
	}
	if hasNext {
		return NewSQLBool(cfg.sqlValueKind, true), iter.Close()
	}
	return NewSQLBool(cfg.sqlValueKind, false), iter.Close()
}

// Evaluate evaluates a SQLExistsExpr into a SQLValue.
// EXISTS returns true if its subquery returns at least one row.
// False is returned if there are no rows. EXISTS never returns NULL.
func (e *SQLExistsExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
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
func (e *SQLExistsExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return e
}

// nolint: unparam
func (e *SQLExistsExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

func (e *SQLExistsExpr) String() string {
	return fmt.Sprintf("exists(%s)", PrettyPrintPlan(e.plan))
}

// EvalType returns the EvalType associated with SQLExistsExpr.
func (*SQLExistsExpr) EvalType() EvalType {
	return EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLExistsExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLExistsExpr into something that can
// be used in an aggregation pipeline. If SQLExistsExpr cannot be translated,
// it will return nil and error.
func (e *SQLExistsExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(e)
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

// Evaluate evaluates a SQLLikeExpr into a SQLValue.
func (e *SQLLikeExpr) Evaluate(ctx context.Context,
	cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {

	value, err := e.left.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if hasNullValue(value) {
		return NewSQLNull(cfg.sqlValueKind, e.EvalType()), nil
	}

	data := value.String()

	value, err = e.right.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if hasNullValue(value) {
		return NewSQLNull(cfg.sqlValueKind, e.EvalType()), nil
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

	return NewSQLBool(cfg.sqlValueKind, matches), nil
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
func (e *SQLLikeExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	// we cannot do a like comparison on an ObjectID in mongodb.
	if c, ok := e.left.(SQLColumnExpr); ok &&
		c.columnType.MongoType == schema.MongoObjectID {
		return nil, e
	}

	name, ok := t.getFieldName(e.left)
	if !ok {
		return nil, e
	}

	value, ok := e.right.(SQLValue)
	if !ok {
		return nil, e
	}

	if hasNullValue(value) {
		return nil, e
	}

	escape, ok := e.escape.(SQLValue)
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

	pattern := ConvertSQLValueToPattern(value, escapeChar)
	opts := "i"
	if e.caseSensitive {
		opts = ""
	}

	return bsonutil.NewM(bsonutil.NewDocElem(name, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpRegex, bson.RegEx{Pattern: pattern, Options: opts})))), nil
}

// evaluate performs evaluation given all SQLValues.
func (e *SQLLikeExpr) evaluate(sqlValueKind SQLValueKind, left, right, escape SQLValue) (SQLValue, error) {
	if hasNullValue(left) {
		return left, nil
	}

	if hasNullValue(right) {
		return right, nil
	}

	data := String(left)

	escapeSeq := []rune(String(escape))
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

	return NewSQLBool(sqlValueKind, matches), nil
}

// EvalType returns the EvalType associated with SQLLikeExpr.
func (*SQLLikeExpr) EvalType() EvalType {
	return EvalBoolean
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLLikeExpr.
func (e *SQLLikeExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	valCount := 0
	var left, right, escape SQLValue
	var ok bool
	if left, ok = e.left.(SQLValue); ok {
		if left.IsNull() {
			return NewSQLNull(cfg.sqlValueKind, left.EvalType())
		}
		valCount++
	}
	if right, ok = e.right.(SQLValue); ok {
		if right.IsNull() {
			return NewSQLNull(cfg.sqlValueKind, right.EvalType())
		}
		valCount++
	}
	if escape, ok = e.escape.(SQLValue); ok {
		valCount++
	}
	if valCount == 3 {
		val, err := e.evaluate(cfg.sqlValueKind, left, right, escape)
		if err != nil {
			return NewSQLNull(cfg.sqlValueKind, left.EvalType())
		}
		return val
	}
	return e
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLLikeExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLLikeExpr into something that can
// be used in an aggregation pipeline. If SQLLikeExpr cannot be translated,
// it will return nil and error.
func (e *SQLLikeExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
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

// Evaluate evaluates a SQLRegexExpr into a SQLValue.
func (e *SQLRegexExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	operand, err := e.operand.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	pattern, err := e.pattern.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	return e.evaluate(cfg.sqlValueKind, operand.SQLVarchar(), pattern.SQLVarchar())
}

// evaluate performs evaluation given all SQLValues.
func (e *SQLRegexExpr) evaluate(sqlValueKind SQLValueKind, operand, pattern SQLValue) (SQLValue, error) {
	if hasNullValue(operand, pattern) {
		return NewSQLNull(sqlValueKind, e.EvalType()), nil
	}

	matcher, err := regexp.CompilePOSIX(pattern.String())
	if err != nil {
		return nil, err
	}
	match := matcher.Find([]byte(operand.String()))
	if match != nil {
		return NewSQLBool(sqlValueKind, true), nil
	}
	return NewSQLBool(sqlValueKind, false), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLRegexExpr.
func (e *SQLRegexExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	var operand, pattern SQLValue
	var ok bool
	valCount := 0
	if operand, ok = e.operand.(SQLValue); ok {
		valCount++
	}
	if pattern, ok = e.pattern.(SQLValue); ok {
		valCount++
	}
	if valCount == 2 {
		val, err := e.evaluate(cfg.sqlValueKind, operand.SQLVarchar(), pattern.SQLVarchar())
		if err != nil {
			return NewSQLNull(cfg.sqlValueKind, e.EvalType())
		}
		return val
	}
	return e
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
func (e *SQLRegexExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	name, ok := t.getFieldName(e.operand)
	if !ok {
		return nil, e
	}

	pattern, ok := e.pattern.(SQLVarchar)
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
	return bsonutil.NewM(
		bsonutil.NewDocElem(name, bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpRegex, bson.RegEx{
				Pattern: pattern.String(),
				Options: "",
			}),
		)),
	), nil
}

// EvalType returns the EvalType associated with SQLRegexExpr.
func (*SQLRegexExpr) EvalType() EvalType {
	return EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLRegexExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLRegexExpr into something that can
// be used in an aggregation pipeline. If SQLRegexExpr cannot be translated,
// it will return nil and error.
func (e *SQLRegexExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(e)
}

// evaluatePlan converts a PlanStage into a table in memory, represented
// as a SQLValues of SQLValues. This table is used as the runtime value of a
// subquery expression and can be cached. Optimization opportunity:
// this function copies all of its input data, value-by-value.
func evaluatePlan(ctx context.Context, cfg *ExecutionConfig,
	st *ExecutionState, plan PlanStage) (*SQLValues, error) {

	iter, err := plan.Open(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	row := &Row{}
	valueTable := &SQLValues{}

	for iter.Next(ctx, row) {
		valueRow := &SQLValues{}
		// release this memory here... it will be re-allocated by a consuming
		// stage
		if err = cfg.memoryMonitor.Release(row.Data.Size()); err != nil {
			_ = iter.Close()
			return nil, err
		}

		// The full data copy here is unwanted.
		// This is a good place to attempt to improve performance.
		for _, value := range row.Data {
			valueRow.Values = append(valueRow.Values, value.Data)
		}
		valueTable.Values = append(valueTable.Values, valueRow)
	}

	return valueTable, iter.Close()
}

// evaluatePlanToScalar converts a PlanStage into a row in memory, represented
// as a SQLValues. This row is used as the runtime value of a
// subquery expression and can be cached. Optimization opportunity:
// this function copies all of its input data, value-by-value.
// This function implements the MySQL behavior of evaluating an empty input
// into a row of NULLs with the same dimension as the input.
func evaluatePlanToScalar(ctx context.Context, cfg *ExecutionConfig,
	st *ExecutionState, plan PlanStage) (*SQLValues, error) {

	iter, err := plan.Open(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	row := &Row{}

	valueRow := &SQLValues{}
	if iter.Next(ctx, row) {
		// release this memory here... it will be re-allocated by a consuming
		// stage
		if err = cfg.memoryMonitor.Release(row.Data.Size()); err != nil {
			_ = iter.Close()
			return nil, err
		}

		// The full data copy here is unwanted.
		// This is a good place to attempt to improve performance.
		for _, value := range row.Data {
			valueRow.Values = append(valueRow.Values, value.Data)
		}
	} else {
		// MySQL specific behavior here.
		for lcv := 0; lcv < len(plan.Columns()); lcv++ {
			valueRow.Values = append(valueRow.Values,
				NewPolymorphicSQLNull(cfg.sqlValueKind))
		}
	}

	// input must not have cardinality > 1
	if iter.Next(ctx, &Row{}) {
		_ = iter.Close()
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErSubqueryNoOneRow)
	}

	return valueRow, iter.Close()
}

// SQLInSubqueryExpr evaluates to true if the left expression is equal to any
// of the rows returned by the right subquery.
// Multi-column right subqueries are valid if the left is a tuple or
// subquery with the same number of columns.
// Multirow left subqueries are never valid.
// Note: This should not be confused with SQL's IN list construct which uses
// the same keyword.
// Note: This should not be. {A NOT IN (...)} is trivial to rewrite to
// {A = ANY (...)}.
type SQLInSubqueryExpr struct {
	correlated bool
	left       SQLExpr
	plan       PlanStage
	// We always cache non-correlated subquery results in their entirety.
	// This is a fine place to be more clever in the future.
	// SQLInSubqueryExpr can cache a whole table, with each row being compared
	// to the value result of the left expression.
	cache *SQLValues
}

// Children returns a slice of all the Node children of the Node.
func (e SQLInSubqueryExpr) Children() []Node {
	return []Node{e.left, e.plan}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (e *SQLInSubqueryExpr) ReplaceChild(i int, n Node) {
	switch i {
	case 0:
		e.left = panicIfNotSQLExpr(e.ExprName(), n)
	case 1:
		e.plan = panicIfNotPlanStage(e.ExprName(), n)
	default:
		panicWithInvalidIndex(e.ExprName(), i, 1)
	}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLInSubqueryExpr) ExprName() string {
	return "SQLInSubqueryExpr"
}

// NewSQLInSubqueryExpr is a constructor for SQLInSubqueryExpr.
func NewSQLInSubqueryExpr(
	correlated bool,
	left SQLExpr,
	plan PlanStage) *SQLInSubqueryExpr {
	return &SQLInSubqueryExpr{
		correlated: correlated,
		left:       left,
		plan:       plan,
	}
}

// Evaluate evaluates a SQLInSubqueryExpr into a SQLValue.
// IN performs a series of comparisons. IN always performs equality comparisons.
// The resulting comparisons within columns of a row are ANDed together.
// Comparisons from separate rows are ORed together.
// Using SQL three-value boolean logic, the results are as follows:
// If a series of comparisons within any row is all true, the result is true.
// If not, if any of the series returns NULL (the series contains NULL and no falses),
// the result is NULL.
// Else, the result is false.
func (e *SQLInSubqueryExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var table *SQLValues
	var err error
	if e.correlated {
		table, err = evaluatePlan(ctx, cfg, st.SubqueryState(), e.plan)
		if err != nil {
			return nil, err
		}
	} else {
		if e.cache == nil {
			// Populate cache.
			e.cache, err = evaluatePlan(ctx, cfg, st, e.plan)
			if err != nil {
				return nil, err
			}
			err = cfg.memoryMonitor.AcquireGlobal(e.cache.Size())
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		table = e.cache
	}

	// Determine length of the left expression.
	var leftLen int
	leftValues, isVals := e.left.(*SQLValues)
	if isVals {
		leftLen = len(leftValues.Values)
	} else {
		leftLen = 1
	}

	sawNull := false
	for _, row := range table.Values {
		right := row.(*SQLValues)

		// Make sure the subquery returns the same number of columns as what
		// it's being compared to.
		// Note: This is redundant to do for each row.
		if leftLen != len(right.Values) {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, leftLen)
		}

		eq := NewSQLEqualsExpr(e.left, right)
		var result SQLValue
		result, err = eq.Evaluate(ctx, cfg, st)
		if err != nil {
			return nil, err
		}
		if Bool(result) {
			return NewSQLBool(cfg.sqlValueKind, true), nil
		}
		if result.IsNull() {
			sawNull = true
		}
	}

	// left expression not found in right table.
	if sawNull {
		return NewPolymorphicSQLNull(cfg.sqlValueKind), nil
	}
	return NewSQLBool(cfg.sqlValueKind, false), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLInSubqueryExpr.
func (e *SQLInSubqueryExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return e
}

// nolint: unparam
func (e *SQLInSubqueryExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

func (e *SQLInSubqueryExpr) String() string {
	return fmt.Sprintf("%v in (%s)", e.left, PrettyPrintPlan(e.plan))
}

// EvalType returns the EvalType associated with SQLInSubqueryExpr.
func (*SQLInSubqueryExpr) EvalType() EvalType {
	return EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLInSubqueryExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLInSubqueryExpr into something that can
// be used in an aggregation pipeline. If SQLInSubqueryExpr cannot be translated,
// it will return nil and error.
func (e *SQLInSubqueryExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(e)
}

// SQLNotInSubqueryExpr evaluates to true if the left expression is not equal
// to all of the rows returned by the right subquery.
// Multi-column right subqueries are valid if the left is a tuple or
// subquery with the same number of columns.
// Multirow left subqueries are never valid.
// Note: This should not be confused with SQL's NOT IN list construct which uses
// the same keyword.
// Note: This should not be. {A NOT IN (...)} is trivial to rewrite to
// {A <> ALL (...)}.
type SQLNotInSubqueryExpr struct {
	correlated bool
	left       SQLExpr
	plan       PlanStage
	// We always cache non-correlated subquery results in their entirety.
	// This is a fine place to be more clever in the future.
	// SQLNotInSubqueryExpr can cache a whole table, with each row being compared
	// to the value result of the left expression.
	cache *SQLValues
}

// Children returns a slice of all the Node children of the Node.
func (e SQLNotInSubqueryExpr) Children() []Node {
	return []Node{e.left, e.plan}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (e *SQLNotInSubqueryExpr) ReplaceChild(i int, n Node) {
	switch i {
	case 0:
		e.left = panicIfNotSQLExpr(e.ExprName(), n)
	case 1:
		e.plan = panicIfNotPlanStage(e.ExprName(), n)
	default:
		panicWithInvalidIndex(e.ExprName(), i, 1)
	}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLNotInSubqueryExpr) ExprName() string {
	return "SQLNotInSubqueryExpr"
}

// NewSQLNotInSubqueryExpr is a constructor for SQLNotInSubqueryExpr.
func NewSQLNotInSubqueryExpr(
	correlated bool,
	left SQLExpr,
	plan PlanStage) *SQLNotInSubqueryExpr {
	return &SQLNotInSubqueryExpr{
		correlated: correlated,
		left:       left,
		plan:       plan,
	}
}

// Evaluate evaluates a SQLNotInSubqueryExpr into a SQLValue.
// NOT IN performs a series of comparisons. NOT IN always performs not-equals comparisons.
// The resulting comparisons within columns of a row are ANDed together.
// Comparisons from separate rows are ORed together.
// Using SQL three-value boolean logic, the results are as follows:
// If a series of comparisons within any row is all true, the result is false.
// If not, if any of the series returns NULL (the series contains NULL and no falses),
// the result is NULL.
// Else, the result is true.
func (e *SQLNotInSubqueryExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var table *SQLValues
	var err error
	if e.correlated {
		table, err = evaluatePlan(ctx, cfg, st.SubqueryState(), e.plan)
		if err != nil {
			return nil, err
		}
	} else {
		if e.cache == nil {
			// Populate cache.
			e.cache, err = evaluatePlan(ctx, cfg, st, e.plan)
			if err != nil {
				return nil, err
			}
			err = cfg.memoryMonitor.AcquireGlobal(e.cache.Size())
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		table = e.cache
	}

	// Determine length of the left expression.
	var leftLen int
	leftValues, isVals := e.left.(*SQLValues)
	if isVals {
		leftLen = len(leftValues.Values)
	} else {
		leftLen = 1
	}

	sawNull := false
	for _, row := range table.Values {
		right := row.(*SQLValues)

		// Make sure the subquery returns the same number of columns as what
		// it's being compared to.
		// Note: This is redundant to do for each row.
		if leftLen != len(right.Values) {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, leftLen)
		}

		eq := NewSQLNotEqualsExpr(e.left, right)
		var result SQLValue
		result, err = eq.Evaluate(ctx, cfg, st)
		if err != nil {
			return nil, err
		}
		if !Bool(result) {
			return NewSQLBool(cfg.sqlValueKind, false), nil
		}
		if result.IsNull() {
			sawNull = true
		}
	}

	// left expression not found in right table.
	if sawNull {
		return NewPolymorphicSQLNull(cfg.sqlValueKind), nil
	}
	return NewSQLBool(cfg.sqlValueKind, true), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLNotInSubqueryExpr.
func (e *SQLNotInSubqueryExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return e
}

// nolint: unparam
func (e *SQLNotInSubqueryExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

func (e *SQLNotInSubqueryExpr) String() string {
	return fmt.Sprintf("%v not in (%s)", e.left, PrettyPrintPlan(e.plan))
}

// EvalType returns the EvalType associated with SQLNotInSubqueryExpr.
func (*SQLNotInSubqueryExpr) EvalType() EvalType {
	return EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLNotInSubqueryExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLNotInSubqueryExpr into something that can
// be used in an aggregation pipeline. If SQLNotInSubqueryExpr cannot be translated,
// it will return nil and error.
func (e *SQLNotInSubqueryExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(e)
}

// SQLAnyExpr evaluates to true if the left expression compares true to
// any of the rows returned by the right subquery by a provided comparison
// operator.
// Multi-column right subqueries are valid if the left is a tuple or
// subquery with the same number of columns.
// Multirow left subqueries are never valid.
type SQLAnyExpr struct {
	correlated bool
	left       SQLExpr
	plan       PlanStage
	operator   string
	// We always cache non-correlated subquery results in their entirety.
	// This is a fine place to be more clever in the future.
	// SQLAnyExpr can cache a whole table, with each row being compared
	// to the value result of the left expression.
	cache *SQLValues
}

// Children returns a slice of all the Node children of the Node.
func (e SQLAnyExpr) Children() []Node {
	return []Node{e.left, e.plan}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (e *SQLAnyExpr) ReplaceChild(i int, n Node) {
	switch i {
	case 0:
		e.left = panicIfNotSQLExpr(e.ExprName(), n)
	case 1:
		e.plan = panicIfNotPlanStage(e.ExprName(), n)
	default:
		panicWithInvalidIndex(e.ExprName(), i, 1)
	}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLAnyExpr) ExprName() string {
	return "SQLAnyExpr"
}

// NewSQLAnyExpr is a constructor for SQLAnyExpr.
func NewSQLAnyExpr(
	correlated bool,
	left SQLExpr,
	plan PlanStage,
	operator string) *SQLAnyExpr {
	return &SQLAnyExpr{
		correlated: correlated,
		left:       left,
		plan:       plan,
		operator:   operator,
	}
}

// Evaluate evaluates a SQLAnyExpr into a SQLValue.
// ANY performs a series of comparisons. ANY uses the provided comparison operator.
// The resulting comparisons within columns of a row are ANDed together.
// Comparisons from separate rows are ORed together.
// Using SQL three-value boolean logic, the results are as follows:
// If a series of comparisons within any row is all true, the result is false.
// If not, if any of the series returns NULL (the series contains NULL and no falses),
// the result is NULL.
// Else, the result is true.
func (e *SQLAnyExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var table *SQLValues
	var err error
	if e.correlated {
		table, err = evaluatePlan(ctx, cfg, st.SubqueryState(), e.plan)
		if err != nil {
			return nil, err
		}
	} else {
		if e.cache == nil {
			// Populate cache.
			e.cache, err = evaluatePlan(ctx, cfg, st, e.plan)
			if err != nil {
				return nil, err
			}
			err = cfg.memoryMonitor.AcquireGlobal(e.cache.Size())
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		table = e.cache
	}

	// Determine length of the left expression.
	var leftLen int
	leftValues, isVals := e.left.(*SQLValues)
	if isVals {
		leftLen = len(leftValues.Values)
	} else {
		leftLen = 1
	}

	sawNull := false
	for _, row := range table.Values {
		right := row.(*SQLValues)

		// Make sure the subquery returns the same number of columns as what
		// it's being compared to.
		// Note: This is redundant to do for each row.
		if leftLen != len(right.Values) {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, leftLen)
		}

		var comp SQLExpr
		comp, err = comparisonExpr(e.left, right, e.operator)
		if err != nil {
			return nil, err
		}
		var result SQLValue
		result, err = comp.Evaluate(ctx, cfg, st)
		if err != nil {
			return nil, err
		}
		if Bool(result) {
			return NewSQLBool(cfg.sqlValueKind, true), nil
		}
		if result.IsNull() {
			sawNull = true
		}
	}

	// left expression not comparing successfully to any row in the right table
	if sawNull {
		return NewPolymorphicSQLNull(cfg.sqlValueKind), nil
	}
	return NewSQLBool(cfg.sqlValueKind, false), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLAnyExpr.
func (e *SQLAnyExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return e
}

// nolint: unparam
func (e *SQLAnyExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

func (e *SQLAnyExpr) String() string {
	return fmt.Sprintf("%v %s any (%s)", e.left, e.operator, PrettyPrintPlan(e.plan))
}

// EvalType returns the EvalType associated with SQLAnyExpr.
func (*SQLAnyExpr) EvalType() EvalType {
	return EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLAnyExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLAnyExpr into something that can
// be used in an aggregation pipeline. If SQLAnyExpr cannot be translated,
// it will return nil and error.
func (e *SQLAnyExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(e)
}

// SQLAllExpr evaluates to true if the left expression compares true to
// all of the rows returned by the right subquery by a provided comparison
// operator.
// Multi-column right subqueries are valid if the left is a tuple or
// subquery with the same number of columns.
// Multirow left subqueries are never valid.
type SQLAllExpr struct {
	correlated bool
	left       SQLExpr
	plan       PlanStage
	operator   string
	// We always cache non-correlated subquery results in their entirety.
	// This is a fine place to be more clever in the future.
	// SQLAllExpr can cache a whole table, with each row being compared
	// to the value result of the left expression.
	cache *SQLValues
}

// Children returns a slice of all the Node children of the Node.
func (e SQLAllExpr) Children() []Node {
	return []Node{e.left, e.plan}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (e *SQLAllExpr) ReplaceChild(i int, n Node) {
	switch i {
	case 0:
		e.left = panicIfNotSQLExpr(e.ExprName(), n)
	case 1:
		e.plan = panicIfNotPlanStage(e.ExprName(), n)
	default:
		panicWithInvalidIndex(e.ExprName(), i, 1)
	}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLAllExpr) ExprName() string {
	return "SQLAllExpr"
}

// NewSQLAllExpr is a constructor for SQLAllExpr.
func NewSQLAllExpr(
	correlated bool,
	left SQLExpr,
	plan PlanStage,
	operator string) *SQLAllExpr {
	return &SQLAllExpr{
		correlated: correlated,
		left:       left,
		plan:       plan,
		operator:   operator,
	}
}

// Evaluate evaluates a SQLAllExpr into a SQLValue.
func (e *SQLAllExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var table *SQLValues
	var err error
	if e.correlated {
		table, err = evaluatePlan(ctx, cfg, st.SubqueryState(), e.plan)
		if err != nil {
			return nil, err
		}
	} else {
		if e.cache == nil {
			// Populate cache.
			e.cache, err = evaluatePlan(ctx, cfg, st, e.plan)
			if err != nil {
				return nil, err
			}
			err = cfg.memoryMonitor.AcquireGlobal(e.cache.Size())
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		table = e.cache
	}

	// Determine length of the left expression.
	var leftLen int
	leftValues, isVals := e.left.(*SQLValues)
	if isVals {
		leftLen = len(leftValues.Values)
	} else {
		leftLen = 1
	}

	sawNull := false
	for _, row := range table.Values {
		right := row.(*SQLValues)

		// Make sure the subquery returns the same number of columns as what
		// it's being compared to.
		// Note: This is redundant to do for each row.
		if leftLen != len(right.Values) {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, leftLen)
		}

		var comp SQLExpr
		comp, err = comparisonExpr(e.left, right, e.operator)
		if err != nil {
			return nil, err
		}
		var result SQLValue
		result, err = comp.Evaluate(ctx, cfg, st)
		if err != nil {
			return nil, err
		}
		if !Bool(result) {
			return NewSQLBool(cfg.sqlValueKind, false), nil
		}
		if result.IsNull() {
			sawNull = true
		}
	}

	// left expression compared successfully to all rows in the right table
	if sawNull {
		return NewPolymorphicSQLNull(cfg.sqlValueKind), nil
	}
	return NewSQLBool(cfg.sqlValueKind, true), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLAllExpr.
func (e *SQLAllExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return e
}

// nolint: unparam
func (e *SQLAllExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

func (e *SQLAllExpr) String() string {
	return fmt.Sprintf("%v %s all (%s)", e.left, e.operator, PrettyPrintPlan(e.plan))
}

// EvalType returns the EvalType associated with SQLAllExpr.
func (*SQLAllExpr) EvalType() EvalType {
	return EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLAllExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLAllExpr into something that can
// be used in an aggregation pipeline. If SQLAllExpr cannot be translated,
// it will return nil and error.
func (e *SQLAllExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(e)
}

// SQLRightSubqueryCmpExpr evaluates to true if the left expression compares true to
// the single row returned by the right subquery by a provided comparison
// operator. The right subquery must be scalar. The left expression is not a subquery.
// See SQLFullSubqueryCmpExpr for representation of other
// cases.
type SQLRightSubqueryCmpExpr struct {
	correlated bool
	left       SQLExpr
	plan       PlanStage
	operator   string
	// We always cache non-correlated subquery results in their entirety.
	// This is a fine place to be more clever in the future.
	// SQLRightSubqueryCmpExpr caches a scalar but it can be multicolumn.
	cache *SQLValues
}

// Children returns a slice of all the Node children of the Node.
func (e SQLRightSubqueryCmpExpr) Children() []Node {
	return []Node{e.left, e.plan}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (e *SQLRightSubqueryCmpExpr) ReplaceChild(i int, n Node) {
	switch i {
	case 0:
		e.left = panicIfNotSQLExpr(e.ExprName(), n)
	case 1:
		e.plan = panicIfNotPlanStage(e.ExprName(), n)
	default:
		panicWithInvalidIndex(e.ExprName(), i, 1)
	}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLRightSubqueryCmpExpr) ExprName() string {
	return "SQLRightSubqueryCmpExpr"
}

// NewSQLRightSubqueryCmpExpr is a constructor for SQLRightSubqueryCmpExpr.
func NewSQLRightSubqueryCmpExpr(
	correlated bool,
	left SQLExpr,
	plan PlanStage,
	operator string) *SQLRightSubqueryCmpExpr {
	return &SQLRightSubqueryCmpExpr{
		correlated: correlated,
		left:       left,
		plan:       plan,
		operator:   operator,
	}
}

// nolint: unparam
func (e *SQLRightSubqueryCmpExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

// Evaluate evaluates a SQLRightSubqueryCmpExpr into a SQLValue.
func (e *SQLRightSubqueryCmpExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var row *SQLValues
	var err error
	if e.correlated {
		row, err = evaluatePlanToScalar(ctx, cfg, st.SubqueryState(), e.plan)
		if err != nil {
			return nil, err
		}
	} else {
		if e.cache == nil {
			// Populate cache.
			e.cache, err = evaluatePlanToScalar(ctx, cfg, st, e.plan)
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		row = e.cache
	}

	// Determine length of the left expression.
	var leftLen int
	leftValues, isVals := e.left.(*SQLValues)
	if isVals {
		leftLen = len(leftValues.Values)
	} else {
		leftLen = 1
	}

	// Make sure the subquery returns the same number of columns as what
	// it's being compared to.
	if leftLen != len(row.Values) {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, leftLen)
	}

	var comp SQLExpr
	comp, err = comparisonExpr(e.left, row, e.operator)
	if err != nil {
		return nil, err
	}
	return comp.Evaluate(ctx, cfg, st)
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLRightSubqueryCmpExpr.
func (e *SQLRightSubqueryCmpExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return e
}

func (e *SQLRightSubqueryCmpExpr) String() string {
	return fmt.Sprintf("%v %s (%s)", e.left, e.operator, PrettyPrintPlan(e.plan))
}

// EvalType returns the EvalType associated with SQLRightSubqueryCmpExpr.
func (*SQLRightSubqueryCmpExpr) EvalType() EvalType {
	return EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLRightSubqueryCmpExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLRightSubqueryCmpExpr into something that can
// be used in an aggregation pipeline. If SQLRightSubqueryCmpExpr cannot be translated,
// it will return nil and error.
func (e *SQLRightSubqueryCmpExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(e)
}

// SQLFullSubqueryCmpExpr evaluates to true if the right subquery compares true to
// the left subquery by a provided comparison operator.
// The left and right subqueries need not be scalar but must produce the same number of rows.
// See SQLRightSubqueryCmpExpr and SQLFullSubqueryCmpExpr for representation of other
// cases.
type SQLFullSubqueryCmpExpr struct {
	leftCorrelated  bool
	rightCorrelated bool
	leftPlan        PlanStage
	rightPlan       PlanStage
	operator        string
	// We always cache non-correlated subquery results in their entirety.
	// This cache is for the left-hand side.
	// SQLFullSubqueryCmpExpr's left cache is scalar but it can be multicolumn.
	leftCache *SQLValues
	// This cache is for the right-hand side.
	// SQLFullSubqueryCmpExpr's right cache is scalar but it can be multicolumn.
	rightCache *SQLValues
	// This cache is for the result. It is used if both sides are non-correlated.
	// This cache consists of a boolean.
	fullCache SQLBool
}

// Children returns a slice of all the Node children of the Node.
func (e SQLFullSubqueryCmpExpr) Children() []Node {
	return []Node{e.leftPlan, e.rightPlan}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (e *SQLFullSubqueryCmpExpr) ReplaceChild(i int, n Node) {
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
func (*SQLFullSubqueryCmpExpr) ExprName() string {
	return "SQLFullSubqueryCmpExpr"
}

// NewSQLFullSubqueryCmpExpr is a constructor for SQLFullSubqueryCmpExpr.
func NewSQLFullSubqueryCmpExpr(
	leftCorrelated bool,
	rightCorrelated bool,
	leftPlan PlanStage,
	rightPlan PlanStage,
	operator string) *SQLFullSubqueryCmpExpr {
	return &SQLFullSubqueryCmpExpr{
		leftCorrelated:  leftCorrelated,
		rightCorrelated: rightCorrelated,
		leftPlan:        leftPlan,
		rightPlan:       rightPlan,
		operator:        operator,
	}
}

// nolint: unparam
func (e *SQLFullSubqueryCmpExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

// Evaluate evaluates a SQLFullSubqueryCmpExpr into a SQLValue.
func (e *SQLFullSubqueryCmpExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	if !e.leftCorrelated && !e.rightCorrelated && e.fullCache != nil {
		return e.fullCache, nil
	}

	var leftRow *SQLValues
	var rightRow *SQLValues
	var err error
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

	// Make sure both subqueres return the same number of columns.
	if len(leftRow.Values) != len(rightRow.Values) {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, len(rightRow.Values))
	}

	var comp SQLExpr
	comp, err = comparisonExpr(leftRow, rightRow, e.operator)
	if err != nil {
		return nil, err
	}
	var result SQLValue
	result, err = comp.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	// Populate full cache.
	if !e.leftCorrelated && !e.rightCorrelated {
		e.fullCache = result.(SQLBool)
	}

	return result, nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLFullSubqueryCmpExpr.
func (e *SQLFullSubqueryCmpExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return e
}

func (e *SQLFullSubqueryCmpExpr) String() string {
	return fmt.Sprintf("(%s) %s (%s)", PrettyPrintPlan(e.leftPlan),
		e.operator, PrettyPrintPlan(e.rightPlan))
}

// EvalType returns the EvalType associated with SQLFullSubqueryCmpExpr.
func (*SQLFullSubqueryCmpExpr) EvalType() EvalType {
	return EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLFullSubqueryCmpExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLFullSubqueryCmpExpr into something that can
// be used in an aggregation pipeline. If SQLFullSubqueryCmpExpr cannot be translated,
// it will return nil and error.
func (e *SQLFullSubqueryCmpExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(e)
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
	leftCache *SQLValues
	// SQLSubqueryAllExpr can cache a whole table, with each row being compared
	// to the value result of the left expression.
	rightCache *SQLValues
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
func (e *SQLSubqueryAllExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var leftRow *SQLValues
	var err error
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
			err = cfg.memoryMonitor.AcquireGlobal(e.leftCache.Size())
			if err != nil {
				return nil, err
			}
		}

		// Read from cache.
		leftRow = e.leftCache
	}

	var rightTable *SQLValues
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
			err = cfg.memoryMonitor.AcquireGlobal(e.rightCache.Size())
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		rightTable = e.rightCache
	}

	leftLen := len(leftRow.Values)
	// <> ALL is rewritten in MySQL to NOT IN.
	// This is the only case when ALL will handle multi column expressions.
	if leftLen > 1 && e.operator != sqlOpNEQ {
		// https://dev.mysql.com/doc/mysql-reslimits-excerpt/5.7/en/subquery-restrictions.html
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, 1)
	}

	sawNull := false
	for _, row := range rightTable.Values {
		right := row.(*SQLValues)

		// Make sure the right subquery returns the same amount of columns as the left.
		// Note: This is redundant to do for each row.
		if len(right.Values) != leftLen {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, leftLen)
		}

		var comp SQLExpr
		comp, err = comparisonExpr(leftRow, right, e.operator)
		if err != nil {
			return nil, err
		}
		var result SQLValue
		result, err = comp.Evaluate(ctx, cfg, st)
		if err != nil {
			return nil, err
		}
		if !Bool(result) {
			return NewSQLBool(cfg.sqlValueKind, false), nil
		}
		if result.IsNull() {
			sawNull = true
		}
	}

	// left expression compared successfully to all rows in the right table
	if sawNull {
		return NewPolymorphicSQLNull(cfg.sqlValueKind), nil
	}
	return NewSQLBool(cfg.sqlValueKind, true), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLSubqueryAllExpr.
func (e *SQLSubqueryAllExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return e
}

// nolint: unparam
func (e *SQLSubqueryAllExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

func (e *SQLSubqueryAllExpr) String() string {
	return fmt.Sprintf("%s\n%s all\n(%s)",
		PrettyPrintPlan(e.leftPlan), e.operator, PrettyPrintPlan(e.rightPlan))
}

// EvalType returns the EvalType associated with SQLSubqueryAllExpr.
func (*SQLSubqueryAllExpr) EvalType() EvalType {
	return EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLSubqueryAllExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLSubqueryAllExpr into something that can
// be used in an aggregation pipeline. If SQLSubqueryAllExpr cannot be translated,
// it will return nil and error.
func (e *SQLSubqueryAllExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(e)
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
	leftCache *SQLValues
	// SQLSubqueryAnyExpr can cache a whole table, with each row being compared
	// to the value result of the left expression.
	rightCache *SQLValues
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

// Evaluate evaluates a SQLSubqueryAnyExpr into a SQLValue.
// ANY performs a series of comparisons. ANY uses the provided comparison operator.
// The resulting comparisons within columns of a row are ANDed together.
// Comparisons from separate rows are ORed together.
// Using SQL three-value boolean logic, the results are as follows:
// If a series of comparisons within any row is all true, the result is false.
// If not, if any of the series returns NULL (the series contains NULL and no falses),
// the result is NULL.
// Else, the result is true.
func (e *SQLSubqueryAnyExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var leftRow *SQLValues
	var err error
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
			err = cfg.memoryMonitor.AcquireGlobal(e.leftCache.Size())
			if err != nil {
				return nil, err
			}
		}

		// Read from cache.
		leftRow = e.leftCache
	}

	var rightTable *SQLValues
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
			err = cfg.memoryMonitor.AcquireGlobal(e.rightCache.Size())
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		rightTable = e.rightCache
	}

	leftLen := len(leftRow.Values)
	// = ANY is rewritten in MySQL to IN.
	// This is the only case when ANY will handle multi column expressions.
	if leftLen > 1 && e.operator != sqlOpEQ {
		// https://dev.mysql.com/doc/mysql-reslimits-excerpt/5.7/en/subquery-restrictions.html
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, 1)
	}

	sawNull := false
	for _, row := range rightTable.Values {
		right := row.(*SQLValues)

		// Make sure the subquery returns the same amount of columns as the left subquery.
		// Note: This is redundant to do for each row.
		if len(right.Values) != leftLen {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, leftLen)
		}

		var comp SQLExpr
		comp, err = comparisonExpr(leftRow, right, e.operator)
		if err != nil {
			return nil, err
		}
		var result SQLValue
		result, err = comp.Evaluate(ctx, cfg, st)
		if err != nil {
			return nil, err
		}
		if Bool(result) {
			return NewSQLBool(cfg.sqlValueKind, true), nil
		}
		if result.IsNull() {
			sawNull = true
		}
	}

	// The left expression did not compare successfully to any row in the right table.
	if sawNull {
		return NewPolymorphicSQLNull(cfg.sqlValueKind), nil
	}
	return NewSQLBool(cfg.sqlValueKind, false), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLSubqueryAnyExpr.
func (e *SQLSubqueryAnyExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return e
}

// nolint: unparam
func (e *SQLSubqueryAnyExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

func (e *SQLSubqueryAnyExpr) String() string {
	return fmt.Sprintf("%s\n%s any\n(%s)",
		PrettyPrintPlan(e.leftPlan), e.operator, PrettyPrintPlan(e.rightPlan))
}

// EvalType returns the EvalType associated with SQLSubqueryAnyExpr.
func (*SQLSubqueryAnyExpr) EvalType() EvalType {
	return EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLSubqueryAnyExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLSubqueryAnyExpr into something that can
// be used in an aggregation pipeline. If SQLSubqueryAnyExpr cannot be translated,
// it will return nil and error.
func (e *SQLSubqueryAnyExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(e)
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
	cache SQLValue
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
func (e *SQLSubqueryExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	if e.correlated {
		return nil, newPushdownFailure(
			e.ExprName(),
			"cannot push down correlated subqueries",
		)
	}

	piece := t.addNonCorrelatedSubqueryFuture(e.plan)
	return bsonutil.WrapInLiteral(piece), nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLSubqueryExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

func (e *SQLSubqueryExpr) evaluateFromPlan(ctx context.Context,
	cfg *ExecutionConfig, st *ExecutionState, plan PlanStage) (SQLValue, error) {
	var err error
	var iter Iter
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

	row := &Row{}

	hasNext := iter.Next(ctx, row)
	if hasNext {

		// release this memory here... it will be re-allocated by a consuming stage
		if err = cfg.memoryMonitor.Release(row.Data.Size()); err != nil {
			_ = iter.Close()
			return nil, err
		}

		// Filter has to check the entire source to return an accurate 'hasNext'
		if iter.Next(ctx, &Row{}) {
			_ = iter.Close()
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErSubqueryNoOneRow)
		}
	}

	switch len(row.Data) {
	case 0:
		return NewSQLNull(cfg.sqlValueKind, e.EvalType()), iter.Close()
	case 1:
		return row.Data[0].Data, iter.Close()
	default:
		eval := &SQLValues{}
		for _, value := range row.Data {
			eval.Values = append(eval.Values, value.Data)
		}
		return eval, iter.Close()
	}
}

// Evaluate evaluates a SQLSubqueryExpr into a SQLValue.
func (e *SQLSubqueryExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	if e.correlated {
		return e.evaluateFromPlan(ctx, cfg, st.SubqueryState(), e.plan)
	}

	var err error
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
func (e *SQLSubqueryExpr) EvalType() EvalType {
	columns := e.plan.Columns()
	if len(columns) == 1 {
		return columns[0].EvalType
	}

	return EvalTuple
}

// SQLSubqueryInSubqueryExpr evaluates to true if the left subquery expression is equal to any
// of the rows returned by the right subquery.
// Multi-column right subqueries are valid if the left is a tuple or
// subquery with the same number of columns.
// Multirow left subqueries are never valid.
// Note: This should not be confused with SQL's IN list construct which uses
// the same keyword.
// Note: This should not be. {A IN (...)} is trivial to rewrite to
// {A = ANY (...)}.
type SQLSubqueryInSubqueryExpr struct {
	leftCorrelated  bool
	rightCorrelated bool
	leftPlan        PlanStage
	rightPlan       PlanStage
	// We always cache non-correlated subquery results in their entirety.
	// SQLSubqueryInSubqueryExpr can cache a row, which is being compared
	// to the value result of the right expression.
	leftCache *SQLValues
	// SQLSubqueryInSubqueryExpr can cache a whole table, with each row being compared
	// to the value result of the left expression.
	rightCache *SQLValues
}

// Children returns a slice of all the Node children of the Node.
func (e SQLSubqueryInSubqueryExpr) Children() []Node {
	return []Node{e.leftPlan, e.rightPlan}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (e *SQLSubqueryInSubqueryExpr) ReplaceChild(i int, n Node) {
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
func (*SQLSubqueryInSubqueryExpr) ExprName() string {
	return "SQLSubqueryInSubqueryExpr"
}

// NewSQLSubqueryInSubqueryExpr is a constructor for SQLSubqueryInSubqueryExpr.
func NewSQLSubqueryInSubqueryExpr(
	leftCorrelated bool,
	rightCorrelated bool,
	leftPlan PlanStage,
	rightPlan PlanStage) *SQLSubqueryInSubqueryExpr {
	return &SQLSubqueryInSubqueryExpr{
		leftCorrelated:  leftCorrelated,
		rightCorrelated: rightCorrelated,
		leftPlan:        leftPlan,
		rightPlan:       rightPlan,
	}
}

// nolint: unparam
func (e *SQLSubqueryInSubqueryExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

// Evaluate evaluates a SQLSubqueryInSubqueryExpr into a SQLValue.
// IN performs a series of comparisons. IN always performs equality comparisons.
// The resulting comparisons within columns of a row are ANDed together.
// Comparisons from separate rows are ORed together.
// Using SQL three-value boolean logic, the results are as follows:
// If a series of comparisons within any row is all true, the result is true.
// If not, if any of the series returns NULL (the series contains NULL and no falses),
// the result is NULL.
// Else, the result is false.
func (e *SQLSubqueryInSubqueryExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var leftRow *SQLValues
	var err error
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
			err = cfg.memoryMonitor.AcquireGlobal(e.leftCache.Size())
			if err != nil {
				return nil, err
			}
		}

		// Read from cache.
		leftRow = e.leftCache
	}

	var rightTable *SQLValues
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
			err = cfg.memoryMonitor.AcquireGlobal(e.rightCache.Size())
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		rightTable = e.rightCache
	}

	leftLen := len(leftRow.Values)
	sawNull := false
	for _, row := range rightTable.Values {
		right := row.(*SQLValues)

		// Make sure the subquery returns the same number of columns as what
		// it's being compared to.
		// Note: This is redundant to do for each row.
		if leftLen != len(right.Values) {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, leftLen)
		}

		eq := NewSQLEqualsExpr(leftRow, right)
		var result SQLValue
		result, err = eq.Evaluate(ctx, cfg, st)
		if err != nil {
			return nil, err
		}
		if Bool(result) {
			return NewSQLBool(cfg.sqlValueKind, true), nil
		}
		if result.IsNull() {
			sawNull = true
		}
	}

	// The left expression was not found in right table.
	if sawNull {
		return NewPolymorphicSQLNull(cfg.sqlValueKind), nil
	}
	return NewSQLBool(cfg.sqlValueKind, false), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLSubqueryInSubqueryExpr.
func (e *SQLSubqueryInSubqueryExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return e
}

func (e *SQLSubqueryInSubqueryExpr) String() string {
	return fmt.Sprintf("%s\nin\n(%s)", PrettyPrintPlan(e.leftPlan), PrettyPrintPlan(e.rightPlan))
}

// EvalType returns the EvalType associated with SQLSubqueryInSubqueryExpr.
func (*SQLSubqueryInSubqueryExpr) EvalType() EvalType {
	return EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLSubqueryInSubqueryExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLSubqueryInSubqueryExpr into something that can
// be used in an aggregation pipeline. If SQLSubqueryInSubqueryExpr cannot be translated,
// it will return nil and error.
func (e *SQLSubqueryInSubqueryExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(e)
}

// SQLSubqueryNotInSubqueryExpr evaluates to true if the left subquery expression is
// not equal to all of the rows returned by the right subquery.
// Multi-column right subqueries are valid if the left is a subquery
// with the same number of columns.
// Multirow left subqueries are never valid.
// Note: This should not be confused with SQL's NOT IN list construct which uses
// the same keyword.
// Note: This should not be. {A NOT IN (...)} is trivial to rewrite to
// {A <> ALL (...)}.
type SQLSubqueryNotInSubqueryExpr struct {
	leftCorrelated  bool
	rightCorrelated bool
	leftPlan        PlanStage
	rightPlan       PlanStage
	// We always cache non-correlated subquery results in their entirety.
	// SQLSubqueryNotInSubqueryExpr can cache a row, which is being compared
	// to the value result of the right expression.
	leftCache *SQLValues
	// SQLSubqueryNotInSubqueryExpr can cache a whole table, with each row being compared
	// to the value result of the left expression.
	rightCache *SQLValues
}

// Children returns a slice of all the Node children of the Node.
func (e SQLSubqueryNotInSubqueryExpr) Children() []Node {
	return []Node{e.leftPlan, e.rightPlan}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (e *SQLSubqueryNotInSubqueryExpr) ReplaceChild(i int, n Node) {
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
func (*SQLSubqueryNotInSubqueryExpr) ExprName() string {
	return "SQLSubqueryNotInSubqueryExpr"
}

// NewSQLSubqueryNotInSubqueryExpr is a constructor for SQLSubqueryNotInSubqueryExpr.
func NewSQLSubqueryNotInSubqueryExpr(
	leftCorrelated bool,
	rightCorrelated bool,
	leftPlan PlanStage,
	rightPlan PlanStage) *SQLSubqueryNotInSubqueryExpr {
	return &SQLSubqueryNotInSubqueryExpr{
		leftCorrelated:  leftCorrelated,
		rightCorrelated: rightCorrelated,
		leftPlan:        leftPlan,
		rightPlan:       rightPlan,
	}
}

// nolint: unparam
func (e *SQLSubqueryNotInSubqueryExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

// Evaluate evaluates a SQLSubqueryNotInSubqueryExpr into a SQLValue.
// NOT IN performs a series of comparisons. NOT IN always performs not-equals comparisons.
// The resulting comparisons within columns of a row are ANDed together.
// Comparisons from separate rows are ORed together.
// Using SQL three-value boolean logic, the results are as follows:
// If a series of comparisons within any row is all true, the result is false.
// If not, if any of the series returns NULL (the series contains NULL and no falses),
// the result is NULL.
// Else, the result is true.
func (e *SQLSubqueryNotInSubqueryExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var leftRow *SQLValues
	var err error
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
			err = cfg.memoryMonitor.AcquireGlobal(e.leftCache.Size())
			if err != nil {
				return nil, err
			}
		}

		// Read from cache.
		leftRow = e.leftCache
	}

	var rightTable *SQLValues
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
			err = cfg.memoryMonitor.AcquireGlobal(e.rightCache.Size())
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		rightTable = e.rightCache
	}

	leftLen := len(leftRow.Values)

	sawNull := false
	for _, row := range rightTable.Values {
		right := row.(*SQLValues)

		// Make sure the subquery returns the same number of columns as what
		// it's being compared to.
		// Note: This is redundant to do for each row.
		if leftLen != len(right.Values) {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, leftLen)
		}

		eq := NewSQLNotEqualsExpr(leftRow, right)
		var result SQLValue
		result, err = eq.Evaluate(ctx, cfg, st)
		if err != nil {
			return nil, err
		}
		if !Bool(result) {
			return NewSQLBool(cfg.sqlValueKind, false), nil
		}
		if result.IsNull() {
			sawNull = true
		}
	}

	// left expression not found in right table.
	if sawNull {
		return NewPolymorphicSQLNull(cfg.sqlValueKind), nil
	}
	return NewSQLBool(cfg.sqlValueKind, true), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLSubqueryNotInSubqueryExpr.
func (e *SQLSubqueryNotInSubqueryExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return e
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLSubqueryExpr.
func (e *SQLSubqueryExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return e
}

func (e *SQLSubqueryNotInSubqueryExpr) String() string {
	return fmt.Sprintf("%s\nnot in\n(%s)", PrettyPrintPlan(e.leftPlan), PrettyPrintPlan(e.rightPlan))
}

// EvalType returns the EvalType associated with SQLSubqueryNotInSubqueryExpr.
func (*SQLSubqueryNotInSubqueryExpr) EvalType() EvalType {
	return EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLSubqueryNotInSubqueryExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLSubqueryNotInSubqueryExpr into something that can
// be used in an aggregation pipeline. If SQLSubqueryNotInSubqueryExpr cannot be translated,
// it will return nil and error.
func (e *SQLSubqueryNotInSubqueryExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(e)
}

// SQLVariableExpr represents a variable lookup.
type SQLVariableExpr struct {
	Name    string
	Kind    variable.Kind
	Scope   variable.Scope
	Value   interface{}
	SQLType schema.SQLType
}

// NewSQLVariableExpr is a constructor for SQLVariableExpr.
func NewSQLVariableExpr(name string, kind variable.Kind, scope variable.Scope, sqlType schema.SQLType, value interface{}) *SQLVariableExpr {
	return &SQLVariableExpr{
		Name:    name,
		Kind:    kind,
		Scope:   scope,
		SQLType: sqlType,
		Value:   value,
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

// Evaluate evaluates a SQLVariableExpr into a SQLValue.
func (e *SQLVariableExpr) Evaluate(_ context.Context, cfg *ExecutionConfig, _ *ExecutionState) (SQLValue, error) {
	val := GoValueToSQLValue(cfg.sqlValueKind, e.Value)
	converted := ConvertTo(val, SQLTypeToEvalType(e.SQLType))
	return converted, nil
}

// nolint: unparam
func (e *SQLVariableExpr) reconcile() (SQLExpr, error) {
	return e, nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLVariableExpr..
// Because variable assignments (even to globals) are not allowed to change during a query,
// it can be constant folded as its value.
func (e *SQLVariableExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	val := GoValueToSQLValue(cfg.sqlValueKind, e.Value)
	converted := ConvertTo(val, SQLTypeToEvalType(e.SQLType))
	return converted
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
func (e *SQLVariableExpr) EvalType() EvalType {
	return SQLTypeToEvalType(e.SQLType)
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (e *SQLVariableExpr) ReplaceChild(i int, n Node) {
	panicWithInvalidIndex(e.ExprName(), i, -1)
}

// ToAggregationLanguage translates SQLVariableExpr into something that can
// be used in an aggregation pipeline. If SQLVariableExpr cannot be translated,
// it will return nil and error.
func (e *SQLVariableExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {

	eb := SQLTypeToEvalType(e.SQLType)
	if eb != EvalBoolean {
		return nil, newPushdownFailure(e.ExprName(), "can only push down boolean variables")
	}

	return bsonutil.WrapInLiteral(e.Value), nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (e *SQLVariableExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
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
