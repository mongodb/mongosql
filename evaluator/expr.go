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
	// Children returns the arguments of the SQLExpr.
	Children() []SQLExpr
	// Evaluate evaluates the receiver expression in memory.
	Evaluate(context.Context, *ExecutionConfig, *ExecutionState) (SQLValue, error)
	// FoldConstants performs constant-folding on this SQLExpr, returning a
	// SQLExpr that is simplified as much as possible.
	FoldConstants(cfg *OptimizerConfig) SQLExpr
	// ReplaceChild sets the value of argument i.
	ReplaceChild(int, SQLExpr)
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

// Children returns the arguments.
func (fe *MongoFilterExpr) Children() []SQLExpr {
	return []SQLExpr{fe.column, fe.expr}
}

// ReplaceChild does nothing for this MongoFilterExpr type.
func (fe *MongoFilterExpr) ReplaceChild(i int, expr SQLExpr) {
	switch i {
	case 0:
		ce, ok := expr.(SQLColumnExpr)
		if !ok {
			panic(fmt.Sprintf("child 0 to MongoFilterExpr must be a SQLColumnExpr not %T", expr))
		}
		fe.column = ce
	case 1:
		fe.expr = expr
	default:
		panic(fmt.Sprintf("child number %d is out of range for MongoFilterExpr", i))
	}
}

// ExprName returns a string representing this SQLExpr's name.
func (*MongoFilterExpr) ExprName() string {
	return "MongoFilterExpr"
}

var _ translatableToMatch = (*MongoFilterExpr)(nil)

// Evaluate evaluates a MongoFilterExpr into a SQLValue.
func (fe *MongoFilterExpr) Evaluate(context.Context, *ExecutionConfig, *ExecutionState) (SQLValue, error) {
	return nil, fmt.Errorf("could not evaluate predicate with mongo filter expression")
}

// nolint: unparam
func (fe *MongoFilterExpr) reconcile() (SQLExpr, error) {
	return fe, nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *MongoFilterExpr.
func (fe *MongoFilterExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return fe
}

func (fe *MongoFilterExpr) String() string {
	return fmt.Sprintf("%v=%v", fe.column.String(), fe.expr.String())
}

// ToMatchLanguage translates MongoFilterExpr into something that can
// be used in an match expression. If MongoFilterExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original MongoFilterExpr.
func (fe *MongoFilterExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	return fe.query, nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (fe *MongoFilterExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return fe.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates MongoFilterExpr into something that can
// be used in an aggregation pipeline. If MongoFilterExpr cannot be translated,
// it will return nil and error.
func (fe *MongoFilterExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(fe)
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

// Children returns the arguments.
func (e *SQLAssignmentExpr) Children() []SQLExpr {
	return []SQLExpr{e.variable, e.expr}
}

// ReplaceChild does nothing for this SQLAssignmentExpr type.
func (e *SQLAssignmentExpr) ReplaceChild(i int, expr SQLExpr) {
	switch i {
	case 0:
		ve, ok := expr.(*SQLVariableExpr)
		if !ok {
			panic(fmt.Sprintf("child 0 to SQLAssignmentExpr must be a *SQLVariableExpr not %T", expr))
		}
		e.variable = ve
	case 1:
		e.expr = expr
	default:
		panic(fmt.Sprintf("child number %d is out of range for SQLAssignmentExpr", i))
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

// Children returns the arguments.
func (e *SQLBenchmarkExpr) Children() []SQLExpr {
	return []SQLExpr{e.count, e.expr}
}

// ReplaceChild does nothing for this SQLBenchmarkExpr type.
func (e *SQLBenchmarkExpr) ReplaceChild(i int, expr SQLExpr) {
	switch i {
	case 0:
		e.count = expr
	case 1:
		e.expr = expr
	default:
		panic(fmt.Sprintf("child number %d is out of range for SQLBenchmarkExpr", i))
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

// Children returns the arguments.
func (e *SQLCaseExpr) Children() []SQLExpr {
	ret := make([]SQLExpr, 2*len(e.caseConditions)+1)
	for i := 0; i < 2*len(e.caseConditions); i += 2 {
		caseCond := e.caseConditions[i/2]
		ret[i], ret[i+1] = caseCond.matcher, caseCond.then
	}
	ret[len(ret)-1] = e.elseValue
	return ret
}

// ReplaceChild does nothing for this SQLCaseExpr type.
func (e *SQLCaseExpr) ReplaceChild(i int, expr SQLExpr) {
	if i < 0 || i > 2*len(e.caseConditions) {
		panic(fmt.Sprintf("child number %d is out of range for SQLCaseExpr: %#v", i, e))
	}
	if i == 2*len(e.caseConditions) {
		e.elseValue = expr
		return
	}
	if i%2 == 0 {
		e.caseConditions[i/2].matcher = expr
		return
	}
	e.caseConditions[i/2].then = expr
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

// Children returns the arguments for c.
func (SQLColumnExpr) Children() []SQLExpr {
	return []SQLExpr{}
}

// Evaluate evaluates a SQLColumnExpr into a SQLValue.
func (c SQLColumnExpr) Evaluate(_ context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {

	for _, row := range st.rows {
		if value, ok := row.GetField(c.selectID, c.databaseName, c.tableName, c.columnName); ok {
			return ConvertTo(value, c.EvalType()), nil
		}
	}

	for _, row := range st.correlatedRows {
		if value, ok := row.GetField(c.selectID, c.databaseName, c.tableName, c.columnName); ok {
			return ConvertTo(value, c.EvalType()), nil
		}
	}

	panic(fmt.Sprintf("cannot find column %q", c))
}

// FoldConstants simplifies expressions containing constants when it is able to for SQLColumnExpr.
func (c SQLColumnExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return c
}

// nolint: unparam
func (c SQLColumnExpr) reconcile() (SQLExpr, error) {
	return c, nil
}

// ReplaceChild sets and argument for c.
func (SQLColumnExpr) ReplaceChild(i int, e SQLExpr) {
	panic("SQLColumnExpr has no children")
}

func (c SQLColumnExpr) String() string {
	var str string
	if c.databaseName != "" {
		str += c.databaseName + "."
	}

	if c.tableName != "" {
		str += c.tableName + "."
	}
	str += c.columnName
	return str
}

// ToAggregationLanguage translates SQLColumnExpr into something that can
// be used in an aggregation pipeline. If SQLColumnExpr cannot be translated,
// it will return nil and error.
func (c SQLColumnExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	if c.correlated {
		cc := t.addCorrelatedSubqueryColumnFuture(&c)
		return bsonutil.WrapInLiteral(cc), nil
	}

	name, ok := t.LookupFieldName(c.databaseName, c.tableName, c.columnName)
	if !ok {
		return nil, newPushdownFailure(
			c.ExprName(),
			"failed to find field name",
			"expr", c.String(),
		)
	}

	return getProjectedFieldName(name, c.columnType.EvalType), nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (c SQLColumnExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return c.ToAggregationLanguage(t)
}

// ToMatchLanguage translates SQLColumnExpr into something that can
// be used in an match expression. If SQLColumnExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLColumnExpr.
func (c SQLColumnExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	if c.correlated {
		cc := t.addCorrelatedSubqueryColumnFuture(&c)
		return bsonutil.WrapInLiteral(cc), nil
	}
	name, ok := t.LookupFieldName(c.databaseName, c.tableName, c.columnName)
	if !ok {
		return nil, c
	}

	if c.EvalType() != EvalBoolean {
		return bsonutil.NewM(
			bsonutil.NewDocElem(name, bsonutil.NewM(
				bsonutil.NewDocElem(bsonutil.OpNeq, nil),
			)),
		), c
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
func (c SQLColumnExpr) EvalType() EvalType {
	return c.columnType.EvalType
}

func (c SQLColumnExpr) isAggregateReplacementColumn() bool {
	return c.tableName == ""
}

// SQLConvertExpr represents a conversion
// of the expression expr to the target
// EvalType.
type SQLConvertExpr struct {
	expr       SQLExpr
	targetType EvalType
}

// Children returns the arguments for c.
func (ce *SQLConvertExpr) Children() []SQLExpr {
	return []SQLExpr{ce.expr}
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
func (ce *SQLConvertExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	v, err := ce.expr.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	return ConvertTo(v, ce.targetType), nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLConvertExpr.
func (ce *SQLConvertExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	if exprVal, ok := ce.expr.(SQLValue); ok {
		out := ConvertTo(exprVal, ce.targetType)
		return out
	}
	return ce
}

// nolint: unparam
func (ce *SQLConvertExpr) reconcile() (SQLExpr, error) {
	return ce, nil
}

// ReplaceChild sets argument i to arg.
func (ce *SQLConvertExpr) ReplaceChild(i int, arg SQLExpr) {
	switch i {
	case 0:
		ce.expr = arg
	default:
		panic(fmt.Sprintf("child number %d is out of range for SQLConvertExpr", i))
	}
}

func (ce *SQLConvertExpr) String() string {
	prettyTypeName := string(EvalTypeToSQLType(ce.targetType))
	return "Convert(" + ce.expr.String() + ", " + prettyTypeName + ")"
}

// EvalType returns the EvalType associated with SQLConvertExpr.
func (ce *SQLConvertExpr) EvalType() EvalType {
	return ce.targetType
}

func (ce *SQLConvertExpr) translateMongoSQL(t *PushdownTranslator) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(4, 0, 0) {
		return nil, newPushdownFailure(
			ce.ExprName(),
			"cannot push down mongosql-mode conversions to MongoDB < 4.0",
		)
	}

	expr, err := t.ToAggregationLanguage(ce.expr)
	if err != nil {
		return nil, err
	}

	converted := translateConvert(expr, ce.expr.EvalType(), ce.targetType)
	return converted, nil
}

func (ce *SQLConvertExpr) translateMySQL(t *PushdownTranslator) (interface{}, PushdownFailure) {
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
			ce.ExprName(),
			"cannot push down mysql-mode conversions to MongoDB < 3.6",
		)
	}

	fromType := ce.expr.EvalType()
	toType := ce.targetType

	expr, err := t.ToAggregationLanguage(ce.expr)
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
			return ce.translateMongoSQL(t)
		}

	case EvalDouble:
		switch toType {
		case EvalInt32, EvalInt64,
			EvalUint32, EvalUint64,
			EvalDecimal128, EvalBoolean:
			return ce.translateMongoSQL(t)
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
			return ce.translateMongoSQL(t)
		}

	default:
		// mysql-mode pushdown not yet implemented for conversions from other types
	}

	return nil, newPushdownFailure(
		ce.ExprName(),
		fmt.Sprintf(
			"cannot push down mysql-mode conversion from type '%s'",
			EvalTypeToMongoType(fromType),
		),
	)
}

// ToAggregationLanguage translates SQLConvertExpr into something that can
// be used in an aggregation pipeline. At the moment, SQLConvertExpr cannot be
// translated, so this function will always return nil and error.
func (ce *SQLConvertExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	if ce.targetType == EvalObjectID {
		sv, ok := ce.expr.(SQLVarchar)
		if !ok {
			return nil, newPushdownFailure(
				ce.ExprName(),
				"can only push down SQLVarchar as ObjectId",
			)
		}
		return sv.SQLObjectID().ToAggregationLanguage(t)
	}

	mode := t.Cfg.sqlValueKind
	switch mode {
	case MySQLValueKind:
		return ce.translateMySQL(t)
	case MongoSQLValueKind:
		return ce.translateMongoSQL(t)
	default:
		panic(fmt.Errorf("impossible value %v for cfg.sqlValueKind", mode))
	}
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (ce *SQLConvertExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return ce.ToAggregationLanguage(t)
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

// Children returns the arguments.
func (*SQLExistsExpr) Children() []SQLExpr {
	return []SQLExpr{}
}

// ReplaceChild does nothing for this SQLExistsExpr type.
func (*SQLExistsExpr) ReplaceChild(i int, e SQLExpr) {
	panic("SQLExistsExpr has no children")
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
func (se *SQLExistsExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	if se.correlated {
		return se.evaluateFromPlan(ctx, cfg, st.SubqueryState(), se.plan)
	}
	if se.cache == nil {
		var err error
		// Populate cache.
		se.cache, err = se.evaluateFromPlan(ctx, cfg, st, se.plan)
		if err != nil {
			return nil, err
		}
	}
	// Read from cache.
	return se.cache, nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLExistsExpr.
func (se *SQLExistsExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return se
}

// nolint: unparam
func (se *SQLExistsExpr) reconcile() (SQLExpr, error) {
	return se, nil
}

func (se *SQLExistsExpr) String() string {
	return fmt.Sprintf("exists(%s)", PrettyPrintPlan(se.plan))
}

// EvalType returns the EvalType associated with SQLExistsExpr.
func (*SQLExistsExpr) EvalType() EvalType {
	return EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (se *SQLExistsExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return se.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLExistsExpr into something that can
// be used in an aggregation pipeline. If SQLExistsExpr cannot be translated,
// it will return nil and error.
func (se *SQLExistsExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(se)
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

// Children returns the arguments.
func (l *SQLLikeExpr) Children() []SQLExpr {
	return []SQLExpr{l.left, l.right, l.escape}
}

// Evaluate evaluates a SQLLikeExpr into a SQLValue.
func (l *SQLLikeExpr) Evaluate(ctx context.Context,
	cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {

	value, err := l.left.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if hasNullValue(value) {
		return NewSQLNull(cfg.sqlValueKind, l.EvalType()), nil
	}

	data := value.String()

	value, err = l.right.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if hasNullValue(value) {
		return NewSQLNull(cfg.sqlValueKind, l.EvalType()), nil
	}

	escape, err := l.escape.Evaluate(ctx, cfg, st)
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
	if l.caseSensitive {
		pattern = ""
	}
	pattern += ConvertSQLValueToPattern(value, escapeChar)

	matches, err := regexp.Match(pattern, []byte(data))
	if err != nil {
		return nil, err
	}

	return NewSQLBool(cfg.sqlValueKind, matches), nil
}

// Normalize will attempt to change SQLLikeExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (l *SQLLikeExpr) Normalize(kind SQLValueKind) Node {
	if right, ok := l.right.(SQLValue); ok {
		if hasNullValue(right) {
			return NewSQLNull(kind, l.EvalType())
		}
	}

	return l
}

// nolint: unparam
func (l *SQLLikeExpr) reconcile() (SQLExpr, error) {
	return l, nil
}

// ReplaceChild does nothing for this SQLLikeExpr type.
func (l *SQLLikeExpr) ReplaceChild(i int, e SQLExpr) {
	switch i {
	case 0:
		l.left = e
	case 1:
		l.right = e
	case 2:
		l.escape = e
	default:
		panic(fmt.Sprintf("child number %d is out of range for SQLLikeExpr", i))
	}
}

func (l *SQLLikeExpr) String() string {
	likeType := "like"
	if l.caseSensitive {
		likeType = "like binary"
	}
	return fmt.Sprintf("%v %v %v", l.left, likeType, l.right)
}

// ToMatchLanguage translates SQLLikeExpr into something that can
// be used in an match expression. If SQLLikeExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLLikeExpr.
func (l *SQLLikeExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	// we cannot do a like comparison on an ObjectID in mongodb.
	if c, ok := l.left.(SQLColumnExpr); ok &&
		c.columnType.MongoType == schema.MongoObjectID {
		return nil, l
	}

	name, ok := t.getFieldName(l.left)
	if !ok {
		return nil, l
	}

	value, ok := l.right.(SQLValue)
	if !ok {
		return nil, l
	}

	if hasNullValue(value) {
		return nil, l
	}

	escape, ok := l.escape.(SQLValue)
	if !ok {
		return nil, l
	}

	escapeSeq := []rune(escape.String())
	if len(escapeSeq) > 1 {
		return nil, l
	}

	var escapeChar rune
	if len(escapeSeq) == 1 {
		escapeChar = escapeSeq[0]
	}

	pattern := ConvertSQLValueToPattern(value, escapeChar)
	opts := "i"
	if l.caseSensitive {
		opts = ""
	}

	return bsonutil.NewM(bsonutil.NewDocElem(name, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpRegex, bson.RegEx{Pattern: pattern, Options: opts})))), nil
}

// evaluate performs evaluation given all SQLValues.
func (l *SQLLikeExpr) evaluate(sqlValueKind SQLValueKind, left, right, escape SQLValue) (SQLValue, error) {
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
	if l.caseSensitive {
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
func (l *SQLLikeExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	valCount := 0
	var left, right, escape SQLValue
	var ok bool
	if left, ok = l.left.(SQLValue); ok {
		if left.IsNull() {
			return NewSQLNull(cfg.sqlValueKind, left.EvalType())
		}
		valCount++
	}
	if right, ok = l.right.(SQLValue); ok {
		if right.IsNull() {
			return NewSQLNull(cfg.sqlValueKind, right.EvalType())
		}
		valCount++
	}
	if escape, ok = l.escape.(SQLValue); ok {
		valCount++
	}
	if valCount == 3 {
		val, err := l.evaluate(cfg.sqlValueKind, left, right, escape)
		if err != nil {
			return NewSQLNull(cfg.sqlValueKind, left.EvalType())
		}
		return val
	}
	return l
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (l *SQLLikeExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return l.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLLikeExpr into something that can
// be used in an aggregation pipeline. If SQLLikeExpr cannot be translated,
// it will return nil and error.
func (l *SQLLikeExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(l)
}

// SQLRegexExpr evaluates to true if the operand matches the regex patttern.
type SQLRegexExpr struct {
	operand, pattern SQLExpr
}

// Children returns the arguments.
func (reg *SQLRegexExpr) Children() []SQLExpr {
	return []SQLExpr{reg.operand, reg.pattern}
}

// ReplaceChild does nothing for this SQLRegexExpr type.
func (reg *SQLRegexExpr) ReplaceChild(i int, e SQLExpr) {
	switch i {
	case 0:
		reg.operand = e
	case 1:
		reg.pattern = e
	default:
		panic(fmt.Sprintf("child number %d is out of range for SQLRegexExpr", i))
	}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLRegexExpr) ExprName() string {
	return "SQLRegexExpr"
}

var _ translatableToMatch = (*SQLRegexExpr)(nil)

// Evaluate evaluates a SQLRegexExpr into a SQLValue.
func (reg *SQLRegexExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	operand, err := reg.operand.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	pattern, err := reg.pattern.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	return reg.evaluate(cfg.sqlValueKind, operand.SQLVarchar(), pattern.SQLVarchar())
}

// evaluate performs evaluation given all SQLValues.
func (reg *SQLRegexExpr) evaluate(sqlValueKind SQLValueKind, operand, pattern SQLValue) (SQLValue, error) {
	if hasNullValue(operand, pattern) {
		return NewSQLNull(sqlValueKind, reg.EvalType()), nil
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
func (reg *SQLRegexExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	var operand, pattern SQLValue
	var ok bool
	valCount := 0
	if operand, ok = reg.operand.(SQLValue); ok {
		valCount++
	}
	if pattern, ok = reg.pattern.(SQLValue); ok {
		valCount++
	}
	if valCount == 2 {
		val, err := reg.evaluate(cfg.sqlValueKind, operand.SQLVarchar(), pattern.SQLVarchar())
		if err != nil {
			return NewSQLNull(cfg.sqlValueKind, reg.EvalType())
		}
		return val
	}
	return reg
}

// nolint: unparam
func (reg *SQLRegexExpr) reconcile() (SQLExpr, error) {
	return reg, nil
}

func (reg *SQLRegexExpr) String() string {
	return fmt.Sprintf("%s matches %s", reg.operand.String(), reg.pattern.String())
}

// ToMatchLanguage translates SQLRegexExpr into something that can
// be used in an match expression. If SQLRegexExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLRegexExpr.
func (reg *SQLRegexExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	name, ok := t.getFieldName(reg.operand)
	if !ok {
		return nil, reg
	}

	pattern, ok := reg.pattern.(SQLVarchar)
	if !ok {
		return nil, reg
	}
	// We need to check if the pattern is valid Extended POSIX regex
	// because MongoDB supports a superset of this specification called
	// PCRE.
	_, err := regexp.CompilePOSIX(pattern.String())
	if err != nil {
		return nil, reg
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
func (reg *SQLRegexExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return reg.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLRegexExpr into something that can
// be used in an aggregation pipeline. If SQLRegexExpr cannot be translated,
// it will return nil and error.
func (reg *SQLRegexExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(reg)
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

// Children returns the arguments.
func (si SQLInSubqueryExpr) Children() []SQLExpr {
	return []SQLExpr{si.left}
}

// ReplaceChild does nothing for this SQLInSubqueryExpr type.
func (si *SQLInSubqueryExpr) ReplaceChild(i int, e SQLExpr) {
	if i != 0 {
		panic("SQLInSubqueryExpr has only one child")
	}
	si.left = e
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
func (si *SQLInSubqueryExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var table *SQLValues
	var err error
	if si.correlated {
		table, err = evaluatePlan(ctx, cfg, st.SubqueryState(), si.plan)
		if err != nil {
			return nil, err
		}
	} else {
		if si.cache == nil {
			// Populate cache.
			si.cache, err = evaluatePlan(ctx, cfg, st, si.plan)
			if err != nil {
				return nil, err
			}
			err = cfg.memoryMonitor.AcquireGlobal(si.cache.Size())
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		table = si.cache
	}

	// Determine length of the left expression.
	var leftLen int
	leftValues, isVals := si.left.(*SQLValues)
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

		eq := NewSQLEqualsExpr(si.left, right)
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
func (si *SQLInSubqueryExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return si
}

// nolint: unparam
func (si *SQLInSubqueryExpr) reconcile() (SQLExpr, error) {
	return si, nil
}

func (si *SQLInSubqueryExpr) String() string {
	return fmt.Sprintf("%v in (%s)", si.left, PrettyPrintPlan(si.plan))
}

// EvalType returns the EvalType associated with SQLInSubqueryExpr.
func (*SQLInSubqueryExpr) EvalType() EvalType {
	return EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (si *SQLInSubqueryExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return si.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLInSubqueryExpr into something that can
// be used in an aggregation pipeline. If SQLInSubqueryExpr cannot be translated,
// it will return nil and error.
func (si *SQLInSubqueryExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(si)
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

// Children returns the arguments.
func (ni SQLNotInSubqueryExpr) Children() []SQLExpr {
	return []SQLExpr{ni.left}
}

// ReplaceChild does nothing for this SQLInSubqueryExpr type.
func (ni *SQLNotInSubqueryExpr) ReplaceChild(i int, e SQLExpr) {
	if i != 0 {
		panic("SQLNotInSubqueryExpr has only one child")
	}
	ni.left = e
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
func (ni *SQLNotInSubqueryExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var table *SQLValues
	var err error
	if ni.correlated {
		table, err = evaluatePlan(ctx, cfg, st.SubqueryState(), ni.plan)
		if err != nil {
			return nil, err
		}
	} else {
		if ni.cache == nil {
			// Populate cache.
			ni.cache, err = evaluatePlan(ctx, cfg, st, ni.plan)
			if err != nil {
				return nil, err
			}
			err = cfg.memoryMonitor.AcquireGlobal(ni.cache.Size())
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		table = ni.cache
	}

	// Determine length of the left expression.
	var leftLen int
	leftValues, isVals := ni.left.(*SQLValues)
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

		eq := NewSQLNotEqualsExpr(ni.left, right)
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
func (ni *SQLNotInSubqueryExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return ni
}

// nolint: unparam
func (ni *SQLNotInSubqueryExpr) reconcile() (SQLExpr, error) {
	return ni, nil
}

func (ni *SQLNotInSubqueryExpr) String() string {
	return fmt.Sprintf("%v not in (%s)", ni.left, PrettyPrintPlan(ni.plan))
}

// EvalType returns the EvalType associated with SQLNotInSubqueryExpr.
func (*SQLNotInSubqueryExpr) EvalType() EvalType {
	return EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (ni *SQLNotInSubqueryExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return ni.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLNotInSubqueryExpr into something that can
// be used in an aggregation pipeline. If SQLNotInSubqueryExpr cannot be translated,
// it will return nil and error.
func (ni *SQLNotInSubqueryExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(ni)
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

// Children returns the arguments.
func (sa SQLAnyExpr) Children() []SQLExpr {
	return []SQLExpr{sa.left}
}

// ReplaceChild does nothing for this SQLInSubqueryExpr type.
func (sa *SQLAnyExpr) ReplaceChild(i int, e SQLExpr) {
	if i != 0 {
		panic("SQLAnyExpr has only one child")
	}
	sa.left = e
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
func (sa *SQLAnyExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var table *SQLValues
	var err error
	if sa.correlated {
		table, err = evaluatePlan(ctx, cfg, st.SubqueryState(), sa.plan)
		if err != nil {
			return nil, err
		}
	} else {
		if sa.cache == nil {
			// Populate cache.
			sa.cache, err = evaluatePlan(ctx, cfg, st, sa.plan)
			if err != nil {
				return nil, err
			}
			err = cfg.memoryMonitor.AcquireGlobal(sa.cache.Size())
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		table = sa.cache
	}

	// Determine length of the left expression.
	var leftLen int
	leftValues, isVals := sa.left.(*SQLValues)
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
		comp, err = comparisonExpr(sa.left, right, sa.operator)
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
func (sa *SQLAnyExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return sa
}

// nolint: unparam
func (sa *SQLAnyExpr) reconcile() (SQLExpr, error) {
	return sa, nil
}

func (sa *SQLAnyExpr) String() string {
	return fmt.Sprintf("%v %s any (%s)", sa.left, sa.operator, PrettyPrintPlan(sa.plan))
}

// EvalType returns the EvalType associated with SQLAnyExpr.
func (*SQLAnyExpr) EvalType() EvalType {
	return EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (sa *SQLAnyExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return sa.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLAnyExpr into something that can
// be used in an aggregation pipeline. If SQLAnyExpr cannot be translated,
// it will return nil and error.
func (sa *SQLAnyExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(sa)
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

// Children returns the arguments.
func (sa SQLAllExpr) Children() []SQLExpr {
	return []SQLExpr{sa.left}
}

// ReplaceChild does nothing for this SQLInSubqueryExpr type.
func (sa *SQLAllExpr) ReplaceChild(i int, e SQLExpr) {
	if i != 0 {
		panic("SQLAllExpr has only one child")
	}
	sa.left = e
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
func (sa *SQLAllExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var table *SQLValues
	var err error
	if sa.correlated {
		table, err = evaluatePlan(ctx, cfg, st.SubqueryState(), sa.plan)
		if err != nil {
			return nil, err
		}
	} else {
		if sa.cache == nil {
			// Populate cache.
			sa.cache, err = evaluatePlan(ctx, cfg, st, sa.plan)
			if err != nil {
				return nil, err
			}
			err = cfg.memoryMonitor.AcquireGlobal(sa.cache.Size())
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		table = sa.cache
	}

	// Determine length of the left expression.
	var leftLen int
	leftValues, isVals := sa.left.(*SQLValues)
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
		comp, err = comparisonExpr(sa.left, right, sa.operator)
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
func (sa *SQLAllExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return sa
}

// nolint: unparam
func (sa *SQLAllExpr) reconcile() (SQLExpr, error) {
	return sa, nil
}

func (sa *SQLAllExpr) String() string {
	return fmt.Sprintf("%v %s all (%s)", sa.left, sa.operator, PrettyPrintPlan(sa.plan))
}

// EvalType returns the EvalType associated with SQLAllExpr.
func (*SQLAllExpr) EvalType() EvalType {
	return EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (sa *SQLAllExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return sa.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLAllExpr into something that can
// be used in an aggregation pipeline. If SQLAllExpr cannot be translated,
// it will return nil and error.
func (sa *SQLAllExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(sa)
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

// Children returns the arguments.
func (sr SQLRightSubqueryCmpExpr) Children() []SQLExpr {
	return []SQLExpr{sr.left}
}

// ReplaceChild does nothing for this SQLExpr type.
func (sr *SQLRightSubqueryCmpExpr) ReplaceChild(i int, e SQLExpr) {
	if i != 0 {
		panic("SQLRightSubqueryCmpExpr has only one child")
	}
	sr.left = e
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
func (sr *SQLRightSubqueryCmpExpr) reconcile() (SQLExpr, error) {
	return sr, nil
}

// Evaluate evaluates a SQLRightSubqueryCmpExpr into a SQLValue.
func (sr *SQLRightSubqueryCmpExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var row *SQLValues
	var err error
	if sr.correlated {
		row, err = evaluatePlanToScalar(ctx, cfg, st.SubqueryState(), sr.plan)
		if err != nil {
			return nil, err
		}
	} else {
		if sr.cache == nil {
			// Populate cache.
			sr.cache, err = evaluatePlanToScalar(ctx, cfg, st, sr.plan)
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		row = sr.cache
	}

	// Determine length of the left expression.
	var leftLen int
	leftValues, isVals := sr.left.(*SQLValues)
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
	comp, err = comparisonExpr(sr.left, row, sr.operator)
	if err != nil {
		return nil, err
	}
	return comp.Evaluate(ctx, cfg, st)
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLRightSubqueryCmpExpr.
func (sr *SQLRightSubqueryCmpExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return sr
}

func (sr *SQLRightSubqueryCmpExpr) String() string {
	return fmt.Sprintf("%v %s (%s)", sr.left, sr.operator, PrettyPrintPlan(sr.plan))
}

// EvalType returns the EvalType associated with SQLRightSubqueryCmpExpr.
func (*SQLRightSubqueryCmpExpr) EvalType() EvalType {
	return EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (sr *SQLRightSubqueryCmpExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return sr.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLRightSubqueryCmpExpr into something that can
// be used in an aggregation pipeline. If SQLRightSubqueryCmpExpr cannot be translated,
// it will return nil and error.
func (sr *SQLRightSubqueryCmpExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(sr)
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

// Children returns the arguments.
func (SQLFullSubqueryCmpExpr) Children() []SQLExpr {
	return []SQLExpr{}
}

// ReplaceChild does nothing for this SQLExpr type.
func (SQLFullSubqueryCmpExpr) ReplaceChild(i int, e SQLExpr) {
	panic("SQLFullSubqueryCmpExpr has no children")
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
func (sf *SQLFullSubqueryCmpExpr) reconcile() (SQLExpr, error) {
	return sf, nil
}

// Evaluate evaluates a SQLFullSubqueryCmpExpr into a SQLValue.
func (sf *SQLFullSubqueryCmpExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	if !sf.leftCorrelated && !sf.rightCorrelated && sf.fullCache != nil {
		return sf.fullCache, nil
	}

	var leftRow *SQLValues
	var rightRow *SQLValues
	var err error
	if sf.leftCorrelated && !sf.rightCorrelated {
		leftRow, err = evaluatePlanToScalar(ctx, cfg, st.SubqueryState(), sf.leftPlan)
		if err != nil {
			return nil, err
		}
		if sf.rightCache == nil {
			// Populate cache.
			sf.rightCache, err = evaluatePlanToScalar(ctx, cfg, st, sf.rightPlan)
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		rightRow = sf.rightCache
	} else if sf.rightCorrelated && !sf.leftCorrelated {
		rightRow, err = evaluatePlanToScalar(ctx, cfg, st.SubqueryState(), sf.rightPlan)
		if err != nil {
			return nil, err
		}
		if sf.leftCache == nil {
			// Populate cache.
			sf.leftCache, err = evaluatePlanToScalar(ctx, cfg, st, sf.leftPlan)
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		leftRow = sf.leftCache
		// Either both sides are correlated or neither are.
	} else {
		leftRow, err = evaluatePlanToScalar(ctx, cfg, st.SubqueryState(), sf.leftPlan)
		if err != nil {
			return nil, err
		}
		rightRow, err = evaluatePlanToScalar(ctx, cfg, st.SubqueryState(), sf.rightPlan)
		if err != nil {
			return nil, err
		}
	}

	// Make sure both subqueres return the same number of columns.
	if len(leftRow.Values) != len(rightRow.Values) {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, len(rightRow.Values))
	}

	var comp SQLExpr
	comp, err = comparisonExpr(leftRow, rightRow, sf.operator)
	if err != nil {
		return nil, err
	}
	var result SQLValue
	result, err = comp.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	// Populate full cache.
	if !sf.leftCorrelated && !sf.rightCorrelated {
		sf.fullCache = result.(SQLBool)
	}

	return result, nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLFullSubqueryCmpExpr.
func (sf *SQLFullSubqueryCmpExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return sf
}

func (sf *SQLFullSubqueryCmpExpr) String() string {
	return fmt.Sprintf("(%s) %s (%s)", PrettyPrintPlan(sf.leftPlan),
		sf.operator, PrettyPrintPlan(sf.rightPlan))
}

// EvalType returns the EvalType associated with SQLFullSubqueryCmpExpr.
func (*SQLFullSubqueryCmpExpr) EvalType() EvalType {
	return EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (sf *SQLFullSubqueryCmpExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return sf.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLFullSubqueryCmpExpr into something that can
// be used in an aggregation pipeline. If SQLFullSubqueryCmpExpr cannot be translated,
// it will return nil and error.
func (sf *SQLFullSubqueryCmpExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(sf)
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

// Children returns the arguments.
func (SQLSubqueryAllExpr) Children() []SQLExpr {
	return []SQLExpr{}
}

// ReplaceChild does nothing for this SQLSubqueryAllExpr type.
func (SQLSubqueryAllExpr) ReplaceChild(i int, e SQLExpr) {
	panic("SQLSubqueryAllExpr has no children")
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
func (sa *SQLSubqueryAllExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var leftRow *SQLValues
	var err error
	if sa.leftCorrelated {
		leftRow, err = evaluatePlanToScalar(ctx, cfg, st.SubqueryState(), sa.leftPlan)
		if err != nil {
			return nil, err
		}
	} else {
		if sa.leftCache == nil {
			// Populate cache.
			sa.leftCache, err = evaluatePlanToScalar(ctx, cfg, st, sa.leftPlan)
			if err != nil {
				return nil, err
			}
			err = cfg.memoryMonitor.AcquireGlobal(sa.leftCache.Size())
			if err != nil {
				return nil, err
			}
		}

		// Read from cache.
		leftRow = sa.leftCache
	}

	var rightTable *SQLValues
	if sa.rightCorrelated {
		rightTable, err = evaluatePlan(ctx, cfg, st.SubqueryState(), sa.rightPlan)
		if err != nil {
			return nil, err
		}
	} else {
		if sa.rightCache == nil {
			// Populate cache.
			sa.rightCache, err = evaluatePlan(ctx, cfg, st, sa.rightPlan)
			if err != nil {
				return nil, err
			}
			err = cfg.memoryMonitor.AcquireGlobal(sa.rightCache.Size())
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		rightTable = sa.rightCache
	}

	leftLen := len(leftRow.Values)
	// <> ALL is rewritten in MySQL to NOT IN.
	// This is the only case when ALL will handle multi column expressions.
	if leftLen > 1 && sa.operator != sqlOpNEQ {
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
		comp, err = comparisonExpr(leftRow, right, sa.operator)
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
func (sa *SQLSubqueryAllExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return sa
}

// nolint: unparam
func (sa *SQLSubqueryAllExpr) reconcile() (SQLExpr, error) {
	return sa, nil
}

func (sa *SQLSubqueryAllExpr) String() string {
	return fmt.Sprintf("%s\n%s all\n(%s)",
		PrettyPrintPlan(sa.leftPlan), sa.operator, PrettyPrintPlan(sa.rightPlan))
}

// EvalType returns the EvalType associated with SQLSubqueryAllExpr.
func (*SQLSubqueryAllExpr) EvalType() EvalType {
	return EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (sa *SQLSubqueryAllExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return sa.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLSubqueryAllExpr into something that can
// be used in an aggregation pipeline. If SQLSubqueryAllExpr cannot be translated,
// it will return nil and error.
func (sa *SQLSubqueryAllExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(sa)
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

// Children returns the arguments.
func (SQLSubqueryAnyExpr) Children() []SQLExpr {
	return []SQLExpr{}
}

// ReplaceChild does nothing for this SQLSubqueryAnyExpr type.
func (SQLSubqueryAnyExpr) ReplaceChild(i int, e SQLExpr) {
	panic("SQLSubqueryAnyExpr has no children")
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
func (sa *SQLSubqueryAnyExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var leftRow *SQLValues
	var err error
	if sa.leftCorrelated {
		leftRow, err = evaluatePlanToScalar(ctx, cfg, st.SubqueryState(), sa.leftPlan)
		if err != nil {
			return nil, err
		}
	} else {
		if sa.leftCache == nil {
			// Populate cache.
			sa.leftCache, err = evaluatePlanToScalar(ctx, cfg, st, sa.leftPlan)
			if err != nil {
				return nil, err
			}
			err = cfg.memoryMonitor.AcquireGlobal(sa.leftCache.Size())
			if err != nil {
				return nil, err
			}
		}

		// Read from cache.
		leftRow = sa.leftCache
	}

	var rightTable *SQLValues
	if sa.rightCorrelated {
		rightTable, err = evaluatePlan(ctx, cfg, st.SubqueryState(), sa.rightPlan)
		if err != nil {
			return nil, err
		}
	} else {
		if sa.rightCache == nil {
			// Populate cache.
			sa.rightCache, err = evaluatePlan(ctx, cfg, st, sa.rightPlan)
			if err != nil {
				return nil, err
			}
			err = cfg.memoryMonitor.AcquireGlobal(sa.rightCache.Size())
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		rightTable = sa.rightCache
	}

	leftLen := len(leftRow.Values)
	// = ANY is rewritten in MySQL to IN.
	// This is the only case when ANY will handle multi column expressions.
	if leftLen > 1 && sa.operator != sqlOpEQ {
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
		comp, err = comparisonExpr(leftRow, right, sa.operator)
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
func (sa *SQLSubqueryAnyExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return sa
}

// nolint: unparam
func (sa *SQLSubqueryAnyExpr) reconcile() (SQLExpr, error) {
	return sa, nil
}

func (sa *SQLSubqueryAnyExpr) String() string {
	return fmt.Sprintf("%s\n%s any\n(%s)",
		PrettyPrintPlan(sa.leftPlan), sa.operator, PrettyPrintPlan(sa.rightPlan))
}

// EvalType returns the EvalType associated with SQLSubqueryAnyExpr.
func (*SQLSubqueryAnyExpr) EvalType() EvalType {
	return EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (sa *SQLSubqueryAnyExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return sa.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLSubqueryAnyExpr into something that can
// be used in an aggregation pipeline. If SQLSubqueryAnyExpr cannot be translated,
// it will return nil and error.
func (sa *SQLSubqueryAnyExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(sa)
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

// Children returns the arguments.
func (SQLSubqueryExpr) Children() []SQLExpr {
	return []SQLExpr{}
}

// ReplaceChild does nothing for this SQLExpr type.
func (SQLSubqueryExpr) ReplaceChild(i int, e SQLExpr) {
	panic("SQLSubqueryExpr has no children")
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
func (se *SQLSubqueryExpr) reconcile() (SQLExpr, error) {
	return se, nil
}

// ToAggregationLanguage translates SQLSubqueryExpr into something that can
// be used in an aggregation pipeline. If SQLSubqueryExpr cannot be translated,
// it will return nil and error.
func (se *SQLSubqueryExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	if se.correlated {
		return nil, newPushdownFailure(
			se.ExprName(),
			"cannot push down correlated subqueries",
		)
	}

	piece := t.addNonCorrelatedSubqueryFuture(se.plan)
	return bsonutil.WrapInLiteral(piece), nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (se *SQLSubqueryExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return se.ToAggregationLanguage(t)
}

func (se *SQLSubqueryExpr) evaluateFromPlan(ctx context.Context,
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
		return NewSQLNull(cfg.sqlValueKind, se.EvalType()), iter.Close()
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
func (se *SQLSubqueryExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	if se.correlated {
		return se.evaluateFromPlan(ctx, cfg, st.SubqueryState(), se.plan)
	}

	var err error
	if se.cache == nil {
		// Populate cache.
		se.cache, err = se.evaluateFromPlan(ctx, cfg, st, se.plan)
		if err != nil {
			return nil, err
		}
	}
	// Read from cache.
	return se.cache, nil
}

// Exprs returns all the SQLColumnExprs associated with the columns of SQLSubqueryExpr.
func (se *SQLSubqueryExpr) Exprs() []SQLExpr {
	exprs := []SQLExpr{}
	for _, c := range se.plan.Columns() {
		exprs = append(exprs, NewSQLColumnExpr(c.SelectID,
			c.Database, c.Table, c.Name, c.EvalType, c.MongoType))
	}

	return exprs
}

func (se *SQLSubqueryExpr) String() string {
	return PrettyPrintPlan(se.plan)
}

// EvalType returns the EvalType associated with SQLSubqueryExpr.
func (se *SQLSubqueryExpr) EvalType() EvalType {
	columns := se.plan.Columns()
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

// Children returns the arguments.
func (SQLSubqueryInSubqueryExpr) Children() []SQLExpr {
	return []SQLExpr{}
}

// ReplaceChild does nothing for this SQLInSubqueryExpr type.
func (SQLSubqueryInSubqueryExpr) ReplaceChild(i int, e SQLExpr) {
	panic("SQLSubqueryInSubqueryExpr has no children")
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
func (si *SQLSubqueryInSubqueryExpr) reconcile() (SQLExpr, error) {
	return si, nil
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
func (si *SQLSubqueryInSubqueryExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var leftRow *SQLValues
	var err error
	if si.leftCorrelated {
		leftRow, err = evaluatePlanToScalar(ctx, cfg, st.SubqueryState(), si.leftPlan)
		if err != nil {
			return nil, err
		}
	} else {
		if si.leftCache == nil {
			// Populate cache.
			si.leftCache, err = evaluatePlanToScalar(ctx, cfg, st, si.leftPlan)
			if err != nil {
				return nil, err
			}
			err = cfg.memoryMonitor.AcquireGlobal(si.leftCache.Size())
			if err != nil {
				return nil, err
			}
		}

		// Read from cache.
		leftRow = si.leftCache
	}

	var rightTable *SQLValues
	if si.rightCorrelated {
		rightTable, err = evaluatePlan(ctx, cfg, st.SubqueryState(), si.rightPlan)
		if err != nil {
			return nil, err
		}
	} else {
		if si.rightCache == nil {
			// Populate cache.
			si.rightCache, err = evaluatePlan(ctx, cfg, st, si.rightPlan)
			if err != nil {
				return nil, err
			}
			err = cfg.memoryMonitor.AcquireGlobal(si.rightCache.Size())
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		rightTable = si.rightCache
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
func (si *SQLSubqueryInSubqueryExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return si
}

func (si *SQLSubqueryInSubqueryExpr) String() string {
	return fmt.Sprintf("%s\nin\n(%s)", PrettyPrintPlan(si.leftPlan), PrettyPrintPlan(si.rightPlan))
}

// EvalType returns the EvalType associated with SQLSubqueryInSubqueryExpr.
func (*SQLSubqueryInSubqueryExpr) EvalType() EvalType {
	return EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (si *SQLSubqueryInSubqueryExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return si.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLSubqueryInSubqueryExpr into something that can
// be used in an aggregation pipeline. If SQLSubqueryInSubqueryExpr cannot be translated,
// it will return nil and error.
func (si *SQLSubqueryInSubqueryExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(si)
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

// Children returns the arguments.
func (SQLSubqueryNotInSubqueryExpr) Children() []SQLExpr {
	return []SQLExpr{}
}

// ReplaceChild does nothing for this SQLSubqueryNotInSubqueryExpr type.
func (SQLSubqueryNotInSubqueryExpr) ReplaceChild(i int, e SQLExpr) {
	panic("SQLSubqueryNotInSubqueryExpr has no children")
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
func (ni *SQLSubqueryNotInSubqueryExpr) reconcile() (SQLExpr, error) {
	return ni, nil
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
func (ni *SQLSubqueryNotInSubqueryExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var leftRow *SQLValues
	var err error
	if ni.leftCorrelated {
		leftRow, err = evaluatePlanToScalar(ctx, cfg, st.SubqueryState(), ni.leftPlan)
		if err != nil {
			return nil, err
		}
	} else {
		if ni.leftCache == nil {
			// Populate cache.
			ni.leftCache, err = evaluatePlanToScalar(ctx, cfg, st, ni.leftPlan)
			if err != nil {
				return nil, err
			}
			err = cfg.memoryMonitor.AcquireGlobal(ni.leftCache.Size())
			if err != nil {
				return nil, err
			}
		}

		// Read from cache.
		leftRow = ni.leftCache
	}

	var rightTable *SQLValues
	if ni.rightCorrelated {
		rightTable, err = evaluatePlan(ctx, cfg, st.SubqueryState(), ni.rightPlan)
		if err != nil {
			return nil, err
		}
	} else {
		if ni.rightCache == nil {
			// Populate cache.
			ni.rightCache, err = evaluatePlan(ctx, cfg, st, ni.rightPlan)
			if err != nil {
				return nil, err
			}
			err = cfg.memoryMonitor.AcquireGlobal(ni.rightCache.Size())
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		rightTable = ni.rightCache
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
func (ni *SQLSubqueryNotInSubqueryExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return ni
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLSubqueryExpr.
func (se *SQLSubqueryExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	return se
}

func (ni *SQLSubqueryNotInSubqueryExpr) String() string {
	return fmt.Sprintf("%s\nnot in\n(%s)", PrettyPrintPlan(ni.leftPlan), PrettyPrintPlan(ni.rightPlan))
}

// EvalType returns the EvalType associated with SQLSubqueryNotInSubqueryExpr.
func (*SQLSubqueryNotInSubqueryExpr) EvalType() EvalType {
	return EvalBoolean
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (ni *SQLSubqueryNotInSubqueryExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return ni.ToAggregationLanguage(t)
}

// ToAggregationLanguage translates SQLSubqueryNotInSubqueryExpr into something that can
// be used in an aggregation pipeline. If SQLSubqueryNotInSubqueryExpr cannot be translated,
// it will return nil and error.
func (ni *SQLSubqueryNotInSubqueryExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return nil, newUntranslatableExprFailure(ni)
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

// Children returns the arguments for c.
func (*SQLVariableExpr) Children() []SQLExpr {
	return []SQLExpr{}
}

// ExprName returns a string representing this SQLExpr's name.
func (*SQLVariableExpr) ExprName() string {
	return "SQLVariableExpr"
}

// Evaluate evaluates a SQLVariableExpr into a SQLValue.
func (v *SQLVariableExpr) Evaluate(_ context.Context, cfg *ExecutionConfig, _ *ExecutionState) (SQLValue, error) {
	val := GoValueToSQLValue(cfg.sqlValueKind, v.Value)
	converted := ConvertTo(val, SQLTypeToEvalType(v.SQLType))
	return converted, nil
}

// nolint: unparam
func (v *SQLVariableExpr) reconcile() (SQLExpr, error) {
	return v, nil
}

// FoldConstants simplifies expressions containing constants when it is able to for *SQLVariableExpr..
// Because variable assignments (even to globals) are not allowed to change during a query,
// it can be constant folded as its value.
func (v *SQLVariableExpr) FoldConstants(cfg *OptimizerConfig) SQLExpr {
	val := GoValueToSQLValue(cfg.sqlValueKind, v.Value)
	converted := ConvertTo(val, SQLTypeToEvalType(v.SQLType))
	return converted
}

func (v *SQLVariableExpr) String() string {
	prefix := ""
	switch v.Kind {
	case variable.UserKind:
		prefix = "@"
	default:
		switch v.Scope {
		case variable.GlobalScope:
			prefix = "@@global."
		case variable.SessionScope:
			prefix = "@@session."
		}
	}

	return prefix + v.Name
}

// EvalType returns the EvalType associated with SQLVariableExpr.
func (v *SQLVariableExpr) EvalType() EvalType {
	return SQLTypeToEvalType(v.SQLType)
}

// ReplaceChild sets and argument for c.
func (*SQLVariableExpr) ReplaceChild(i int, e SQLExpr) {
	panic("SQLVariableExpr has no children")
}

// ToAggregationLanguage translates SQLVariableExpr into something that can
// be used in an aggregation pipeline. If SQLVariableExpr cannot be translated,
// it will return nil and error.
func (v *SQLVariableExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {

	e := SQLTypeToEvalType(v.SQLType)
	if e != EvalBoolean {
		return nil, newPushdownFailure(v.ExprName(), "can only push down boolean variables")
	}

	return bsonutil.WrapInLiteral(v.Value), nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (v *SQLVariableExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return v.ToAggregationLanguage(t)
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
