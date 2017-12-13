package evaluator

import (
	"fmt"
	"math"
	"regexp"
	"strconv"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
	"github.com/shopspring/decimal"
)

const (
	ADD ArithmeticOperator = iota
	DIV
	MULT
	SUB
)

const (
	subqueryAll = iota
	subqueryAny
	subqueryIn
	subqueryNotIn
	subquerySome
)

const (
	// GlobalStatus is a global server status variable.
	GlobalStatus VariableKind = "global_status"
	// GlobalVariable is a global system variable.
	GlobalVariable VariableKind = "global"
	// SessionStatus is a session(local) server status variable
	SessionStatus VariableKind = "session_status"
	// SessionVariable is a session(local) variable.
	SessionVariable VariableKind = "session"
	// UserVariable is a custom variable associated with a session(local).
	UserVariable VariableKind = "user"
)

//
// SQLArithmetic is used to do arithmetic on all types.
//
type SQLArithmetic interface {
	Decimal128() decimal.Decimal
	Float64() float64
	Int64() int64
	Uint64() uint64
}

//
// SQLExpr is the base type for a SQL expression.
//
type SQLExpr interface {
	node
	Evaluate(*EvalCtx) (SQLValue, error)
	String() string
	Type() schema.SQLType
}

//
// SQLValue is a SQLExpr with a value.
//
type SQLValue interface {
	SQLExpr
	SQLArithmetic
	Value() interface{}
	Size() uint64
}

type reconcilingSQLExpr interface {
	SQLExpr
	reconcile() (SQLExpr, error)
}

type ArithmeticOperator byte

//
// MongoFilterExpr holds a MongoDB filter expression
//
type MongoFilterExpr struct {
	column SQLColumnExpr
	expr   SQLExpr
	query  bson.M
}

func (fe *MongoFilterExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return nil, fmt.Errorf("could not evaluate predicate with mongo filter expression")
}

func (fe *MongoFilterExpr) String() string {
	return fmt.Sprintf("%v=%v", fe.column.String(), fe.expr.String())
}

func (fe *MongoFilterExpr) ToMatchLanguage(t *pushDownTranslator) (bson.M, SQLExpr) {
	return fe.query, nil
}

func (*MongoFilterExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}

//
// SQLAddExpr evaluates to the sum of two expressions.
//
type SQLAddExpr sqlBinaryNode

func (add *SQLAddExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftVal, err := add.left.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	rightVal, err := add.right.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	if hasNullValue(leftVal, rightVal) {
		return SQLNull, nil
	}

	return doArithmetic(leftVal, rightVal, ADD)
}

func (add *SQLAddExpr) String() string {
	return fmt.Sprintf("%v+%v", add.left, add.right)
}

func (add *SQLAddExpr) ToAggregationLanguage(t *pushDownTranslator) (interface{}, bool) {
	left, ok := t.ToAggregationLanguage(add.left)
	if !ok {
		return nil, false
	}

	right, ok := t.ToAggregationLanguage(add.right)
	if !ok {
		return nil, false
	}

	return bson.M{"$add": []interface{}{left, right}}, true
}

func (add *SQLAddExpr) Type() schema.SQLType {
	return schema.SQLFloat
}

//
// SQLAndExpr evaluates to true if and only if all its children evaluate to true.
//
type SQLAndExpr sqlBinaryNode

func (and *SQLAndExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	left, err := and.left.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	right, err := and.right.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if isFalsy(left) || isFalsy(right) {
		return SQLFalse, nil
	}

	if hasNullValue(left, right) {
		return SQLNull, nil
	}

	return SQLTrue, nil
}

func (and *SQLAndExpr) Normalize() node {
	left, leftOk := and.left.(SQLValue)
	if leftOk && isFalsy(left) {
		return SQLFalse
	} else if leftOk && isTruthy(left) {
		return and.right
	}

	right, rightOk := and.right.(SQLValue)
	if rightOk && isFalsy(right) {
		return SQLFalse
	} else if rightOk && isTruthy(right) {
		return and.left
	}

	return and
}

func (and *SQLAndExpr) String() string {
	return fmt.Sprintf("%v and %v", and.left, and.right)
}

func (and *SQLAndExpr) ToAggregationLanguage(t *pushDownTranslator) (interface{}, bool) {

	left, ok := t.ToAggregationLanguage(and.left)
	if !ok {
		return nil, false
	}

	right, ok := t.ToAggregationLanguage(and.right)
	if !ok {
		return nil, false
	}

	letAssignment := bson.M{
		"left": left, "right": right,
	}

	letEvaluation := bson.M{
		mgoOperatorCond: []interface{}{
			bson.M{
				mgoOperatorOr: []interface{}{
					bson.M{
						mgoOperatorEq: []interface{}{
							bson.M{
								mgoOperatorIfNull: []interface{}{"$$left", nil}},
							nil,
						},
					},
					bson.M{
						mgoOperatorEq: []interface{}{
							bson.M{
								mgoOperatorIfNull: []interface{}{"$$right", nil}},
							nil,
						},
					},
				},
			},
			bson.M{
				mgoOperatorCond: []interface{}{
					bson.M{
						mgoOperatorOr: []interface{}{
							bson.M{
								mgoOperatorEq: []interface{}{"$$left", false}},
							bson.M{
								mgoOperatorEq: []interface{}{"$$right", false}},
							bson.M{
								mgoOperatorEq: []interface{}{"$$left", 0}},
							bson.M{
								mgoOperatorEq: []interface{}{"$$right", 0}},
						},
					},
					bson.M{
						mgoOperatorAnd: []interface{}{"$$left", "$$right"},
					},
					mgoNullLiteral,
				},
			},
			bson.M{
				mgoOperatorAnd: []interface{}{"$$left", "$$right"},
			},
		},
	}

	return wrapInLet(letAssignment, letEvaluation), true

}

func (and *SQLAndExpr) ToMatchLanguage(t *pushDownTranslator) (bson.M, SQLExpr) {
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
		cond := []interface{}{}
		if v, ok := left[mgoOperatorAnd]; ok {
			array := v.([]interface{})
			cond = append(cond, array...)
		} else {
			cond = append(cond, left)
		}

		if v, ok := right[mgoOperatorAnd]; ok {
			array := v.([]interface{})
			cond = append(cond, array...)
		} else {
			cond = append(cond, right)
		}

		match = bson.M{mgoOperatorAnd: cond}
	}

	if exLeft == nil && exRight == nil {
		return match, nil
	} else if exLeft != nil && exRight == nil {
		return match, exLeft
	} else if exLeft == nil && exRight != nil {
		return match, exRight
	}
	return match, &SQLAndExpr{exLeft, exRight}
}

func (*SQLAndExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}

//
// SQLAssignmentExpr handles assigning a value to a variable.
//
type SQLAssignmentExpr struct {
	variable *SQLVariableExpr
	expr     SQLExpr
}

func (e *SQLAssignmentExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	value, err := e.expr.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	err = ctx.Variables().Set(variable.Name(e.variable.Name), e.variable.Scope, e.variable.Kind, value.Value())
	return value, err
}

func (e *SQLAssignmentExpr) String() string {
	return fmt.Sprintf("%s := %s", e.variable.String(), e.expr.String())
}

func (e *SQLAssignmentExpr) Type() schema.SQLType {
	return e.expr.Type()
}

// SQLBenchmarkExpr evaluates expr the number of times given by count.
// https://dev.mysql.com/doc/refman/5.5/en/information-functions.html#function_benchmark
type SQLBenchmarkExpr struct {
	count SQLExpr
	expr  SQLExpr
}

func (e SQLBenchmarkExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	count, err := e.count.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	replaced, err := replaceMongoSourceStages(e.expr, ctx)
	if err != nil {
		return nil, err
	}

	for i := int64(0); i < count.Int64(); i++ {
		_, err := replaced.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
	}

	return SQLInt(0), nil
}

func (e SQLBenchmarkExpr) String() string {
	return fmt.Sprintf("benchmark(%s, %s)", e.count.String(), e.expr.String())
}

func (e SQLBenchmarkExpr) Type() schema.SQLType {
	return schema.SQLInt
}

//
// SQLCaseExpr holds a number of cases to evaluate as well as the value
// to return if any of the cases is matched. If none is matched,
// 'elseValue' is evaluated and returned.
//
type SQLCaseExpr struct {
	elseValue      SQLExpr
	caseConditions []caseCondition
}

func (e SQLCaseExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	for _, condition := range e.caseConditions {
		m, err := Matches(condition.matcher, ctx)
		if err != nil {
			return nil, err
		}

		if m {
			return condition.then.Evaluate(ctx)
		}
	}

	return e.elseValue.Evaluate(ctx)
}

func (e SQLCaseExpr) String() string {
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

func (e *SQLCaseExpr) ToAggregationLanguage(t *pushDownTranslator) (interface{}, bool) {
	elseValue, ok := t.ToAggregationLanguage(e.elseValue)
	if !ok {
		return nil, false
	}

	var conditions []interface{}
	var thens []interface{}
	for _, condition := range e.caseConditions {
		var c interface{}
		if matcher, ok := condition.matcher.(*SQLEqualsExpr); ok {
			newMatcher := &SQLOrExpr{matcher, &SQLEqualsExpr{matcher.left, SQLTrue}}
			c, ok = t.ToAggregationLanguage(newMatcher)
			if !ok {
				return nil, false
			}
		} else {
			c, ok = t.ToAggregationLanguage(condition.matcher)
			if !ok {
				return nil, false
			}
		}

		then, ok := t.ToAggregationLanguage(condition.then)
		if !ok {
			return nil, false
		}

		conditions = append(conditions, c)
		thens = append(thens, then)
	}

	if len(conditions) != len(thens) {
		return nil, false
	}

	cases := elseValue

	for i := len(conditions) - 1; i >= 0; i-- {
		cases = wrapInCond(thens[i], cases, conditions[i])
	}

	return cases, true

}

func (e SQLCaseExpr) Type() schema.SQLType {
	conds := []SQLExpr{e.elseValue}
	for _, cond := range e.caseConditions {
		conds = append(conds, cond.then)
	}
	return preferentialType(conds...)
}

//
// SQLColumnExpr represents a column reference.
//
type SQLColumnExpr struct {
	selectID     int
	databaseName string
	tableName    string
	columnName   string
	columnType   schema.ColumnType
}

func (c SQLColumnExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	// first check our immediate rows
	for _, row := range ctx.Rows {
		if value, ok := row.GetField(c.selectID, c.databaseName, c.tableName, c.columnName); ok {
			return NewSQLValueFromSQLColumnExpr(value, c.columnType.SQLType, c.columnType.MongoType)
		}
	}

	// If we didn't find it there, search in the src rows, which contain parent
	// information in the case we are evaluating a correlated column.
	if ctx.ExecutionCtx != nil {
		for _, row := range ctx.ExecutionCtx.SrcRows {
			if value, ok := row.GetField(c.selectID, c.databaseName, c.tableName, c.columnName); ok {
				return NewSQLValueFromSQLColumnExpr(value, c.columnType.SQLType, c.columnType.MongoType)
			}
		}
	}

	return SQLNull, nil
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

func (c SQLColumnExpr) ToAggregationLanguage(t *pushDownTranslator) (interface{}, bool) {

	name, ok := t.lookupFieldName(c.databaseName, c.tableName, c.columnName)
	if !ok {
		return nil, false
	}

	return getProjectedFieldName(name, c.columnType.SQLType), true

}

func (c SQLColumnExpr) ToMatchLanguage(t *pushDownTranslator) (bson.M, SQLExpr) {
	name, ok := t.lookupFieldName(c.databaseName, c.tableName, c.columnName)
	if !ok {
		return nil, c
	}

	if c.Type() != schema.SQLBoolean {
		return bson.M{
			name: bson.M{
				mgoOperatorNeq: nil,
			},
		}, c
	}

	return bson.M{
		mgoOperatorAnd: []interface{}{
			bson.M{
				name: bson.M{
					mgoOperatorNeq: false,
				},
			},
			bson.M{
				name: bson.M{
					mgoOperatorNeq: nil,
				},
			},
			bson.M{
				name: bson.M{
					mgoOperatorNeq: 0,
				},
			},
		},
	}, nil
}

func (c SQLColumnExpr) Type() schema.SQLType {
	if c.columnType.MongoType == schema.MongoObjectID && c.columnType.SQLType == schema.SQLVarchar {
		return schema.SQLObjectID
	}

	if c.columnType.MongoType == schema.MongoDecimal128 {
		return schema.SQLDecimal128
	}

	return c.columnType.SQLType
}

func (c SQLColumnExpr) isAggregateReplacementColumn() bool {
	return c.tableName == ""
}

//
// SQLConvertExpr wraps a SQLExpr that can be
// converted to another SQLType.
//
type SQLConvertExpr struct {
	expr         SQLExpr
	convType     schema.SQLType
	defaultValue SQLValue
}

func (ce *SQLConvertExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	// collapse nested SQLConvertExprs
	if sce, ok := ce.expr.(*SQLConvertExpr); ok {
		ce.expr = sce.expr
	}

	v, err := ce.expr.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if ce.defaultValue != SQLNone {
		return NewSQLValueWithDefault(v.Value(), ce.convType, ce.expr.Type(), ce.defaultValue), nil
	}

	val, _ := NewSQLValue(v.Value(), ce.convType, ce.expr.Type())
	return val, nil
}

func (ce *SQLConvertExpr) String() string {
	return ce.expr.String()
}

func (ce *SQLConvertExpr) Type() schema.SQLType {
	return ce.convType
}

//
// SQLDivideExpr evaluates to the quotient of the left expression divided by the right.
//
type SQLDivideExpr sqlBinaryNode

func (div *SQLDivideExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftVal, err := div.left.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	rightVal, err := div.right.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	if rightVal.Float64() == 0 || hasNullValue(leftVal, rightVal) {
		return SQLNull, nil
	}

	return doArithmetic(leftVal, rightVal, DIV)
}

func (div *SQLDivideExpr) Normalize() node {
	if hasNullExpr(div.left, div.right) {
		return SQLNull
	}
	return div
}

func (div *SQLDivideExpr) String() string {
	return fmt.Sprintf("%v/%v", div.left, div.right)
}

func (div *SQLDivideExpr) ToAggregationLanguage(t *pushDownTranslator) (interface{}, bool) {
	left, ok := t.ToAggregationLanguage(div.left)
	if !ok {
		return nil, false
	}

	right, ok := t.ToAggregationLanguage(div.right)
	if !ok {
		return nil, false
	}

	letAssignment := bson.M{
		"left": left, "right": right,
	}

	letEvaluation := wrapInCond(
		nil,
		bson.M{"$divide": []interface{}{"$$left", "$$right"}},
		bson.M{mgoOperatorEq: []interface{}{"$$right", 0}},
	)

	return wrapInLet(letAssignment, letEvaluation), true

}

func (div *SQLDivideExpr) Type() schema.SQLType {
	return schema.SQLFloat
}

//
// SQLEqualsExpr evaluates to true if the left equals the right.
//
type SQLEqualsExpr sqlBinaryNode

func (eq *SQLEqualsExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftVal, err := eq.left.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	rightVal, err := eq.right.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	if hasNullValue(leftVal, rightVal) {
		return SQLNull, nil
	}

	c, err := CompareTo(leftVal, rightVal, ctx.Collation)
	if err == nil {
		return NewSQLBool(c == 0), nil
	}

	return SQLFalse, err
}

func (eq *SQLEqualsExpr) Normalize() node {
	if hasNullExpr(eq.left, eq.right) {
		return SQLNull
	}

	if shouldFlip(sqlBinaryNode(*eq)) {
		return &SQLEqualsExpr{eq.right, eq.left}
	}

	return eq
}

func (eq *SQLEqualsExpr) String() string {
	return fmt.Sprintf("%v = %v", eq.left, eq.right)
}

func (eq *SQLEqualsExpr) ToAggregationLanguage(t *pushDownTranslator) (interface{}, bool) {
	left, ok := t.ToAggregationLanguage(eq.left)
	if !ok {
		return nil, false
	}

	right, ok := t.ToAggregationLanguage(eq.right)
	if !ok {
		return nil, false
	}

	letAssignment := bson.M{
		"left": left, "right": right,
	}

	letEvaluation := wrapInNullCheckedCond(
		nil,
		bson.M{
			mgoOperatorEq: []interface{}{"$$left", "$$right"},
		},
		"$$left",
		"$$right",
	)

	return wrapInLet(letAssignment, letEvaluation), true

}

func (eq *SQLEqualsExpr) ToMatchLanguage(t *pushDownTranslator) (bson.M, SQLExpr) {
	match, ok := t.translateOperator(mgoOperatorEq, eq.left, eq.right)
	if !ok {
		return nil, eq
	}
	return match, nil
}

func (*SQLEqualsExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}

func (eq *SQLEqualsExpr) reconcile() (SQLExpr, error) {
	var reconciled bool
	var err error

	left := eq.left
	right := eq.right

	if isBooleanColumnAndArithmetic(left, right) || isBooleanColumnAndArithmetic(right, left) {
		var lit SQLArithmetic
		var col SQLColumnExpr

		switch left.Type() {
		case schema.SQLBoolean:
			col = left.(SQLColumnExpr)
			lit = right.(SQLArithmetic)
		default:
			col = right.(SQLColumnExpr)
			lit = left.(SQLArithmetic)
		}

		if lit.Int64() == 1 || lit.Int64() == 0 {
			left = col
			right = &SQLConvertExpr{lit.(SQLExpr), schema.SQLBoolean, SQLNone}
			reconciled = true
		}
	}

	if !reconciled {
		left, right, err = reconcileSQLExprs(eq.left, eq.right)
	}

	return &SQLEqualsExpr{left, right}, err
}

//
// SQLExistsExpr evaluates to true if any result is returned from the subquery.
//
type SQLExistsExpr struct {
	expr *SQLSubqueryExpr
}

func (em *SQLExistsExpr) Evaluate(ctx *EvalCtx) (value SQLValue, err error) {
	var it Iter
	var matches bool

	defer func() {
		if it != nil && err == nil {
			err = it.Err()
		}
	}()

	execCtx := ctx.ExecutionCtx

	if em.expr.correlated {
		execCtx = ctx.CreateChildExecutionCtx()
	}

	it, err = em.expr.plan.Open(execCtx)
	if err != nil {
		return SQLFalse, err
	}

	if it.Next(&Row{}) {
		matches = true
	}

	return NewSQLBool(matches), it.Close()
}

func (em *SQLExistsExpr) String() string {
	return fmt.Sprintf("exists(%s)", em.expr.String())
}

func (*SQLExistsExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}

//
// SQLGreaterThanExpr evaluates to true when the left is greater than the right.
//
type SQLGreaterThanExpr sqlBinaryNode

func (gt *SQLGreaterThanExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftVal, err := gt.left.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	rightVal, err := gt.right.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	if hasNullValue(leftVal, rightVal) {
		return SQLNull, nil
	}

	c, err := CompareTo(leftVal, rightVal, ctx.Collation)
	if err == nil {
		return NewSQLBool(c > 0), nil
	}
	return SQLFalse, err
}

func (gt *SQLGreaterThanExpr) Normalize() node {
	if hasNullExpr(gt.left, gt.right) {
		return SQLNull
	}

	if shouldFlip(sqlBinaryNode(*gt)) {
		return &SQLLessThanExpr{gt.right, gt.left}
	}

	return gt
}

func (gt *SQLGreaterThanExpr) String() string {
	return fmt.Sprintf("%v>%v", gt.left, gt.right)
}

func (gt *SQLGreaterThanExpr) ToAggregationLanguage(t *pushDownTranslator) (interface{}, bool) {
	left, ok := t.ToAggregationLanguage(gt.left)
	if !ok {
		return nil, false
	}

	right, ok := t.ToAggregationLanguage(gt.right)
	if !ok {
		return nil, false
	}

	letAssignment := bson.M{
		"left": left, "right": right,
	}

	letEvaluation := wrapInNullCheckedCond(
		nil,
		bson.M{
			mgoOperatorGt: []interface{}{"$$left", "$$right"},
		},
		"$$left",
		"$$right",
	)

	return wrapInLet(letAssignment, letEvaluation), true

}

func (gt *SQLGreaterThanExpr) ToMatchLanguage(t *pushDownTranslator) (bson.M, SQLExpr) {
	match, ok := t.translateOperator(mgoOperatorGt, gt.left, gt.right)
	if !ok {
		return nil, gt
	}
	return match, nil
}

func (*SQLGreaterThanExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}

//
// SQLGreaterThanOrEqualExpr evaluates to true when the left is greater than or equal to the right.
//
type SQLGreaterThanOrEqualExpr sqlBinaryNode

func (gte *SQLGreaterThanOrEqualExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftVal, err := gte.left.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	rightVal, err := gte.right.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	if hasNullValue(leftVal, rightVal) {
		return SQLNull, nil
	}

	c, err := CompareTo(leftVal, rightVal, ctx.Collation)
	if err == nil {
		return NewSQLBool(c >= 0), nil
	}

	return SQLFalse, err
}

func (gte *SQLGreaterThanOrEqualExpr) Normalize() node {
	if hasNullExpr(gte.left, gte.right) {
		return SQLNull
	}

	if shouldFlip(sqlBinaryNode(*gte)) {
		return &SQLLessThanOrEqualExpr{gte.right, gte.left}
	}

	return gte
}

func (gte *SQLGreaterThanOrEqualExpr) String() string {
	return fmt.Sprintf("%v>=%v", gte.left, gte.right)
}

func (gte *SQLGreaterThanOrEqualExpr) ToAggregationLanguage(t *pushDownTranslator) (interface{}, bool) {
	left, ok := t.ToAggregationLanguage(gte.left)
	if !ok {
		return nil, false
	}

	right, ok := t.ToAggregationLanguage(gte.right)
	if !ok {
		return nil, false
	}

	letAssignment := bson.M{
		"left": left, "right": right,
	}

	letEvaluation := wrapInNullCheckedCond(
		nil,
		bson.M{
			mgoOperatorGte: []interface{}{"$$left", "$$right"},
		},
		"$$left",
		"$$right",
	)

	return wrapInLet(letAssignment, letEvaluation), true

}

func (gte *SQLGreaterThanOrEqualExpr) ToMatchLanguage(t *pushDownTranslator) (bson.M, SQLExpr) {
	match, ok := t.translateOperator(mgoOperatorGte, gte.left, gte.right)
	if !ok {
		return nil, gte
	}
	return match, nil
}

func (*SQLGreaterThanOrEqualExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}

//
// SQLIDivideExpr evaluates the integer quotient of the left expression divided by the right.
//
type SQLIDivideExpr sqlBinaryNode

func (div *SQLIDivideExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftVal, err := div.left.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	rightVal, err := div.right.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if rightVal.Float64() == 0 || hasNullValue(leftVal, rightVal) {
		// NOTE: this is per the mysql manual.
		return SQLNull, nil
	}

	return SQLInt(int(leftVal.Float64() / rightVal.Float64())), nil
}

func (div *SQLIDivideExpr) Normalize() node {
	if hasNullExpr(div.left, div.right) {
		return SQLNull
	}
	return div
}

func (div *SQLIDivideExpr) String() string {
	return fmt.Sprintf("%v/%v", div.left, div.right)
}

func (div *SQLIDivideExpr) ToAggregationLanguage(t *pushDownTranslator) (interface{}, bool) {
	left, ok := t.ToAggregationLanguage(div.left)
	if !ok {
		return nil, false
	}

	right, ok := t.ToAggregationLanguage(div.right)
	if !ok {
		return nil, false
	}

	letAssignment := bson.M{
		"left": left, "right": right,
	}

	letEvaluation := wrapInCond(
		nil,
		bson.M{
			"$trunc": []interface{}{
				bson.M{
					"$divide": []interface{}{"$$left", "$$right"},
				},
			},
		},
		bson.M{mgoOperatorEq: []interface{}{"$$right", 0}},
	)

	return wrapInLet(letAssignment, letEvaluation), true

}

func (div *SQLIDivideExpr) Type() schema.SQLType {
	return preferentialType(div.left, div.right)
}

//
// SQLInExpr evaluates to true if the left is in any of the values on the right.
//
type SQLInExpr sqlBinaryNode

func (in *SQLInExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	left, err := in.left.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	right, err := in.right.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	// right child must be of type SQLValues
	// TODO: can we not simply require this as part of the node definition?
	rightChild, ok := right.(*SQLValues)
	if !ok {
		child, ok := right.(SQLValue)
		if !ok {
			return SQLFalse, fmt.Errorf("right 'in' expression is type %T - expected tuple", right)
		}
		rightChild = &SQLValues{[]SQLValue{child}}
	}

	leftChild, ok := left.(*SQLValues)
	if ok {
		if len(leftChild.Values) != 1 {
			return SQLFalse, fmt.Errorf("left operand should contain 1 column - got %v", len(leftChild.Values))
		}
		left = leftChild.Values[0]
	} else {
		if _, ok = left.(SQLNullValue); ok {
			return SQLNull, nil
		}
	}

	nullInValues := false
	for _, right := range rightChild.Values {
		if right == SQLNull {
			nullInValues = true
		}
		eq := &SQLEqualsExpr{left, right}
		m, err := Matches(eq, ctx)
		if err != nil {
			return SQLFalse, err
		}
		if m {
			return SQLTrue, nil
		}
	}

	if nullInValues {
		return SQLNull, nil
	}

	return SQLFalse, nil
}

func (in *SQLInExpr) Normalize() node {
	if hasNullExpr(in.left) {
		return SQLNull
	}

	return in
}

func (in *SQLInExpr) String() string {
	return fmt.Sprintf("%v in %v", in.left, in.right)
}

func (in *SQLInExpr) ToAggregationLanguage(t *pushDownTranslator) (interface{}, bool) {
	left, ok := t.ToAggregationLanguage(in.left)
	if !ok {
		return nil, false
	}

	exprs := getSQLInExprs(in.right)
	if exprs == nil {
		return nil, false
	}

	nullInValues := false
	var right []interface{}
	for _, expr := range exprs {
		if expr == SQLNull {
			nullInValues = true
			continue
		}
		val, ok := t.ToAggregationLanguage(expr)
		if !ok {
			return nil, false
		}
		right = append(right, val)
	}

	return wrapInNullCheckedCond(
		nil,
		wrapInCond(
			true,
			wrapInCond(
				nil,
				false,
				bson.M{mgoOperatorEq: []interface{}{nullInValues, true}},
			),
			bson.M{mgoOperatorGt: []interface{}{
				bson.M{"$size": bson.M{"$filter": bson.M{"input": right,
					"as":   "item",
					"cond": bson.M{mgoOperatorEq: []interface{}{"$$item", left}},
				}}},
				bson.M{"$literal": 0},
			}}),
		left,
	), true

}

func (in *SQLInExpr) ToMatchLanguage(t *pushDownTranslator) (bson.M, SQLExpr) {
	name, ok := t.getFieldName(in.left)
	if !ok {
		return nil, in
	}

	exprs := getSQLInExprs(in.right)
	if exprs == nil {
		return nil, in
	}

	values := []interface{}{}

	for _, expr := range exprs {
		value, ok := t.getValue(expr)
		if !ok {
			return nil, in
		}
		values = append(values, value)
	}

	return bson.M{name: bson.M{"$in": values}}, nil
}

func (*SQLInExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}

//
// SQLIsExpr evaluates to true if the left is equal to the boolean value on the right.
//
type SQLIsExpr sqlBinaryNode

func (is *SQLIsExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftVal, err := is.left.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	rightVal, err := is.right.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	if _, ok := leftVal.(SQLNullValue); ok {
		if _, ok := rightVal.(SQLBool); ok {
			return SQLFalse, nil
		}
		return SQLTrue, nil
	}

	if hasNullValue(leftVal, rightVal) {
		return SQLFalse, nil
	}

	if isTruthy(leftVal) && isTruthy(rightVal) || isFalsy(leftVal) && isFalsy(rightVal) {
		return SQLTrue, nil
	}

	return SQLFalse, nil

}

func (is *SQLIsExpr) String() string {
	return fmt.Sprintf("%v is %v", is.left, is.right)
}

func (is *SQLIsExpr) ToAggregationLanguage(t *pushDownTranslator) (interface{}, bool) {
	left, ok := t.ToAggregationLanguage(is.left)
	if !ok {
		return nil, false
	}

	right, ok := t.ToAggregationLanguage(is.right)
	if !ok {
		return nil, false
	}

	// if right side is {null,unknown}, it's a simple case
	if is.right == SQLNull {
		return wrapInOp(mgoOperatorEq,
			wrapInIfNull(left, mgoNullLiteral),
			right,
		), true
	}
	// otherwise, the right side is a boolean

	// if left side is a boolean, this is still simple
	if is.left.Type() == schema.SQLBoolean {
		return wrapInOp(mgoOperatorEq,
			left,
			right,
		), true
	}

	// otherwise, left side is a number type
	if is.right == SQLTrue {
		return wrapInCond(
			false,
			wrapInOp(mgoOperatorNeq,
				left,
				0,
			),
			wrapInNullCheck(left),
		), true
	} else if is.right == SQLFalse {
		return wrapInOp(mgoOperatorEq,
			left,
			0,
		), true
	}

	// SQL Values
	return nil, false
}

func (is *SQLIsExpr) ToMatchLanguage(t *pushDownTranslator) (bson.M, SQLExpr) {
	name, ok := t.getFieldName(is.left)
	if !ok {
		return nil, is
	}
	switch is.right {
	case SQLNull:
		return bson.M{name: nil}, nil
	case SQLFalse:
		if is.left.Type() == schema.SQLBoolean {
			return bson.M{name: false}, nil
		}
		return bson.M{name: 0}, nil
	case SQLTrue:
		if is.left.Type() == schema.SQLBoolean {
			return bson.M{name: true}, nil
		}
		return bson.M{
			mgoOperatorAnd: []interface{}{
				bson.M{name: bson.M{mgoOperatorNeq: 0}},
				bson.M{name: bson.M{mgoOperatorNeq: nil}},
			},
		}, nil
	}
	return nil, is
}

func (*SQLIsExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}

//
// SQLLessThanExpr evaluates to true when the left is less than the right.
//
type SQLLessThanExpr sqlBinaryNode

func (lt *SQLLessThanExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftVal, err := lt.left.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	rightVal, err := lt.right.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	if hasNullValue(leftVal, rightVal) {
		return SQLNull, nil
	}

	c, err := CompareTo(leftVal, rightVal, ctx.Collation)
	if err == nil {
		return NewSQLBool(c < 0), nil
	}
	return SQLFalse, err
}

func (lt *SQLLessThanExpr) Normalize() node {
	if hasNullExpr(lt.left, lt.right) {
		return SQLNull
	}

	if shouldFlip(sqlBinaryNode(*lt)) {
		return &SQLGreaterThanExpr{lt.right, lt.left}
	}

	return lt
}

func (lt *SQLLessThanExpr) String() string {
	return fmt.Sprintf("%v<%v", lt.left, lt.right)
}

func (lt *SQLLessThanExpr) ToAggregationLanguage(t *pushDownTranslator) (interface{}, bool) {
	left, ok := t.ToAggregationLanguage(lt.left)
	if !ok {
		return nil, false
	}

	right, ok := t.ToAggregationLanguage(lt.right)
	if !ok {
		return nil, false
	}

	letAssignment := bson.M{
		"left": left, "right": right,
	}

	letEvaluation := wrapInNullCheckedCond(
		nil,
		bson.M{
			mgoOperatorLt: []interface{}{"$$left", "$$right"},
		},
		"$$left", "$$right",
	)

	return wrapInLet(letAssignment, letEvaluation), true

}

func (lt *SQLLessThanExpr) ToMatchLanguage(t *pushDownTranslator) (bson.M, SQLExpr) {
	match, ok := t.translateOperator(mgoOperatorLt, lt.left, lt.right)
	if !ok {
		return nil, lt
	}
	return match, nil
}

func (*SQLLessThanExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}

//
// SQLLessThanOrEqualExpr evaluates to true when the left is less than or equal to the right.
//
type SQLLessThanOrEqualExpr sqlBinaryNode

func (lte *SQLLessThanOrEqualExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftVal, err := lte.left.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	rightVal, err := lte.right.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	if hasNullValue(leftVal, rightVal) {
		return SQLNull, nil
	}

	c, err := CompareTo(leftVal, rightVal, ctx.Collation)
	if err == nil {
		return NewSQLBool(c <= 0), nil
	}
	return SQLFalse, err
}

func (lte *SQLLessThanOrEqualExpr) Normalize() node {
	if hasNullExpr(lte.left, lte.right) {
		return SQLNull
	}

	if shouldFlip(sqlBinaryNode(*lte)) {
		return &SQLGreaterThanOrEqualExpr{lte.right, lte.left}
	}

	return lte
}

func (lte *SQLLessThanOrEqualExpr) String() string {
	return fmt.Sprintf("%v<=%v", lte.left, lte.right)
}

func (lte *SQLLessThanOrEqualExpr) ToAggregationLanguage(t *pushDownTranslator) (interface{}, bool) {
	left, ok := t.ToAggregationLanguage(lte.left)
	if !ok {
		return nil, false
	}

	right, ok := t.ToAggregationLanguage(lte.right)
	if !ok {
		return nil, false
	}

	letAssignment := bson.M{
		"left": left, "right": right,
	}

	letEvaluation := wrapInNullCheckedCond(
		nil,
		bson.M{
			mgoOperatorLte: []interface{}{"$$left", "$$right"},
		},
		"$$left", "$$right",
	)

	return wrapInLet(letAssignment, letEvaluation), true

}

func (lte *SQLLessThanOrEqualExpr) ToMatchLanguage(t *pushDownTranslator) (bson.M, SQLExpr) {
	match, ok := t.translateOperator(mgoOperatorLte, lte.left, lte.right)
	if !ok {
		return nil, lte
	}
	return match, nil
}

func (*SQLLessThanOrEqualExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}

//
// SQLLikeExpr evaluates to true if the left is 'like' the right.
//
type SQLLikeExpr struct {
	left   SQLExpr
	right  SQLExpr
	escape SQLExpr
}

func (l *SQLLikeExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	value, err := l.left.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if hasNullValue(value) {
		return SQLNull, nil
	}

	data := value.String()

	value, err = l.right.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if hasNullValue(value) {
		return SQLNull, nil
	}

	escape, err := l.escape.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	escapeSeq := []rune(escape.String())
	if len(escapeSeq) > 1 {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_ARGUMENTS, "ESCAPE")
	}

	var escapeChar rune
	if len(escapeSeq) == 1 {
		escapeChar = escapeSeq[0]
	}

	pattern := "(?i)" + convertSQLValueToPattern(value, escapeChar)

	matches, err := regexp.Match(pattern, []byte(data))
	if err != nil {
		return nil, err
	}

	return NewSQLBool(matches), nil
}

func (l *SQLLikeExpr) Normalize() node {
	if right, ok := l.right.(SQLValue); ok {
		if hasNullValue(right) {
			return SQLNull
		}
	}

	return l
}

func (l *SQLLikeExpr) String() string {
	return fmt.Sprintf("%v like %v", l.left, l.right)
}

func (l *SQLLikeExpr) ToMatchLanguage(t *pushDownTranslator) (bson.M, SQLExpr) {
	// we cannot do a like comparison on an ObjectID in mongodb.
	if l.left.Type() == schema.SQLObjectID {
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

	pattern := convertSQLValueToPattern(value, escapeChar)

	return bson.M{name: bson.M{"$regex": bson.RegEx{Pattern: pattern, Options: "i"}}}, nil
}

func (*SQLLikeExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}

//
// SQLModExpr evaluates the modulus of two expressions
//
type SQLModExpr sqlBinaryNode

func (mod *SQLModExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftVal, err := mod.left.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	rightVal, err := mod.right.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	if math.Abs(rightVal.Float64()) == 0 || hasNullValue(leftVal, rightVal) {
		return SQLNull, nil
	}

	modVal := math.Mod(leftVal.Float64(), rightVal.Float64())
	if modVal == -0 {
		modVal *= -1
	}

	return SQLFloat(modVal), nil
}

func (mod *SQLModExpr) String() string {
	return fmt.Sprintf("%v/%v", mod.left, mod.right)
}

func (mod *SQLModExpr) ToAggregationLanguage(t *pushDownTranslator) (interface{}, bool) {
	left, ok := t.ToAggregationLanguage(mod.left)
	if !ok {
		return nil, false
	}

	right, ok := t.ToAggregationLanguage(mod.right)
	if !ok {
		return nil, false
	}

	return bson.M{"$mod": []interface{}{left, right}}, true

}

func (mod *SQLModExpr) Type() schema.SQLType {
	return preferentialType(mod.left, mod.right)
}

//
// SQLMultiplyExpr evaluates to the product of two expressions
//
type SQLMultiplyExpr sqlBinaryNode

func (mult *SQLMultiplyExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftVal, err := mult.left.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	rightVal, err := mult.right.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	if hasNullValue(leftVal, rightVal) {
		return SQLNull, nil
	}

	return doArithmetic(leftVal, rightVal, MULT)
}

func (mult *SQLMultiplyExpr) String() string {
	return fmt.Sprintf("%v*%v", mult.left, mult.right)
}

func (mult *SQLMultiplyExpr) ToAggregationLanguage(t *pushDownTranslator) (interface{}, bool) {
	left, ok := t.ToAggregationLanguage(mult.left)
	if !ok {
		return nil, false
	}

	right, ok := t.ToAggregationLanguage(mult.right)
	if !ok {
		return nil, false
	}

	return bson.M{"$multiply": []interface{}{left, right}}, true

}

func (mult *SQLMultiplyExpr) Type() schema.SQLType {
	return schema.SQLFloat
}

//
// SQLNotEqualsExpr evaluates to true if the left does not equal the right.
//
type SQLNotEqualsExpr sqlBinaryNode

func (neq *SQLNotEqualsExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftVal, err := neq.left.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	rightVal, err := neq.right.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	if hasNullValue(leftVal, rightVal) {
		return SQLNull, nil
	}

	c, err := CompareTo(leftVal, rightVal, ctx.Collation)
	if err == nil {
		return NewSQLBool(c != 0), nil
	}

	return SQLFalse, err
}

func (neq *SQLNotEqualsExpr) Normalize() node {
	if hasNullExpr(neq.left, neq.right) {
		return SQLNull
	}

	if shouldFlip(sqlBinaryNode(*neq)) {
		return &SQLNotEqualsExpr{neq.right, neq.left}
	}

	return neq
}

func (neq *SQLNotEqualsExpr) String() string {
	return fmt.Sprintf("%v != %v", neq.left, neq.right)
}

func (neq *SQLNotEqualsExpr) ToAggregationLanguage(t *pushDownTranslator) (interface{}, bool) {
	left, ok := t.ToAggregationLanguage(neq.left)
	if !ok {
		return nil, false
	}

	right, ok := t.ToAggregationLanguage(neq.right)
	if !ok {
		return nil, false
	}

	letAssignment := bson.M{
		"left": left, "right": right,
	}

	letEvaluation := wrapInNullCheckedCond(
		nil,
		bson.M{
			mgoOperatorNeq: []interface{}{"$$left", "$$right"},
		},
		"$$left", "$$right",
	)

	return wrapInLet(letAssignment, letEvaluation), true

}

func (neq *SQLNotEqualsExpr) ToMatchLanguage(t *pushDownTranslator) (bson.M, SQLExpr) {
	match, ok := t.translateOperator(mgoOperatorNeq, neq.left, neq.right)
	if !ok {
		return nil, neq
	}

	value, ok := t.getValue(neq.right)
	if !ok {
		return nil, neq
	}

	if value != nil {
		name, ok := t.getFieldName(neq.left)
		if !ok {
			return nil, neq
		}
		match = bson.M{
			mgoOperatorAnd: []interface{}{match,
				bson.M{name: bson.M{mgoOperatorNeq: nil}},
			},
		}
	}

	return match, nil
}

func (*SQLNotEqualsExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}

//
// SQLNotExpr evaluates to the inverse of its child.
//
type SQLNotExpr sqlUnaryNode

func (not *SQLNotExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	operand, err := not.operand.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if hasNullValue(operand) {
		return SQLNull, nil
	}

	if !isTruthy(operand) {
		return SQLTrue, nil
	}

	return SQLFalse, nil
}

func (not *SQLNotExpr) Normalize() node {
	if operand, ok := not.operand.(SQLValue); ok {
		if hasNullValue(operand) {
			return SQLNull
		}

		if isTruthy(operand) {
			return SQLFalse
		} else if isFalsy(operand) {
			return SQLTrue
		}
	}

	return not
}

func (not *SQLNotExpr) String() string {
	return fmt.Sprintf("not %v", not.operand)
}

func (not *SQLNotExpr) ToAggregationLanguage(t *pushDownTranslator) (interface{}, bool) {
	op, ok := t.ToAggregationLanguage(not.operand)
	if !ok {
		return nil, false
	}

	return wrapInNullCheckedCond(nil, bson.M{"$not": op}, op), true

}

func (not *SQLNotExpr) ToMatchLanguage(t *pushDownTranslator) (bson.M, SQLExpr) {
	match, ex := t.ToMatchLanguage(not.operand)
	if match == nil {
		return nil, not
	} else if ex == nil {
		return negate(match), nil
	} else {
		// partial translation of Not
		return negate(match), &SQLNotExpr{ex}
	}

}

func (*SQLNotExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}

//
// SQLNullSafeEqualsExpr behaves like the = operator,
// but returns 1 rather than NULL if both operands are
// NULL, and 0 rather than NULL if one operand is NULL.
//
type SQLNullSafeEqualsExpr sqlBinaryNode

func (nse *SQLNullSafeEqualsExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftVal, err := nse.left.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	rightVal, err := nse.right.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	if leftVal == SQLNull {
		if rightVal == SQLNull {
			return SQLTrue, nil
		}
		return SQLFalse, nil
	}

	if rightVal == SQLNull {
		if leftVal == SQLNull {
			return SQLTrue, nil
		}
		return SQLFalse, nil
	}

	c, err := CompareTo(leftVal, rightVal, ctx.Collation)
	if err == nil {
		return NewSQLBool(c == 0), nil
	}

	return SQLFalse, err
}

func (nse *SQLNullSafeEqualsExpr) Normalize() node {
	if nse.left == SQLNull {
		if nse.right == SQLNull {
			return SQLTrue
		}
		return SQLFalse
	}

	if nse.right == SQLNull {
		if nse.left == SQLNull {
			return SQLTrue
		}
		return SQLFalse
	}

	return nse
}

func (nse *SQLNullSafeEqualsExpr) String() string {
	return fmt.Sprintf("%v <=> %v", nse.left, nse.right)
}

func (nse *SQLNullSafeEqualsExpr) ToAggregationLanguage(t *pushDownTranslator) (interface{}, bool) {
	left, ok := t.ToAggregationLanguage(nse.left)
	if !ok {
		return nil, false
	}

	right, ok := t.ToAggregationLanguage(nse.right)
	if !ok {
		return nil, false
	}

	return bson.M{mgoOperatorEq: []interface{}{left, right}}, true

}

func (*SQLNullSafeEqualsExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}

//
// SQLOrExpr evaluates to true if any of its children evaluate to true.
//
type SQLOrExpr sqlBinaryNode

func (or *SQLOrExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	left, err := or.left.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	right, err := or.right.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if isTruthy(left) || isTruthy(right) {
		return SQLTrue, nil
	}

	if hasNullValue(left, right) {
		return SQLNull, nil
	}

	return SQLFalse, nil
}

func (or *SQLOrExpr) Normalize() node {
	left, leftOk := or.left.(SQLValue)

	if leftOk && isTruthy(left) {
		return SQLTrue
	} else if leftOk && isFalsy(left) {
		return or.right
	}

	right, rightOk := or.right.(SQLValue)
	if rightOk && isTruthy(right) {
		return SQLTrue
	} else if rightOk && isFalsy(right) {
		return or.left
	}

	return or
}

func (or *SQLOrExpr) String() string {
	return fmt.Sprintf("%v or %v", or.left, or.right)
}

func (or *SQLOrExpr) ToAggregationLanguage(t *pushDownTranslator) (interface{}, bool) {
	left, ok := t.ToAggregationLanguage(or.left)
	if !ok {
		return nil, false
	}

	right, ok := t.ToAggregationLanguage(or.right)
	if !ok {
		return nil, false
	}

	letAssignment := bson.M{
		"left": left, "right": right,
	}

	leftIsFalse := bson.M{mgoOperatorOr: []interface{}{
		bson.M{mgoOperatorEq: []interface{}{"$$left", false}},
		bson.M{mgoOperatorEq: []interface{}{"$$left", 0}},
	}}

	leftIsTrue := bson.M{mgoOperatorOr: []interface{}{
		bson.M{mgoOperatorNeq: []interface{}{"$$left", false}},
		bson.M{mgoOperatorNeq: []interface{}{"$$left", 0}},
	}}

	rightIsFalse := bson.M{mgoOperatorOr: []interface{}{
		bson.M{mgoOperatorEq: []interface{}{"$$right", false}},
		bson.M{mgoOperatorEq: []interface{}{"$$right", 0}},
	}}

	rightIsTrue := bson.M{mgoOperatorOr: []interface{}{
		bson.M{mgoOperatorNeq: []interface{}{"$$right", false}},
		bson.M{mgoOperatorNeq: []interface{}{"$$right", 0}},
	}}

	leftIsNull := bson.M{mgoOperatorEq: []interface{}{
		bson.M{
			mgoOperatorIfNull: []interface{}{"$$left", nil}},
		nil,
	}}

	rightIsNull := bson.M{mgoOperatorEq: []interface{}{
		bson.M{
			mgoOperatorIfNull: []interface{}{"$$right", nil}},
		nil,
	}}

	nullOrFalse := bson.M{mgoOperatorOr: []interface{}{
		bson.M{mgoOperatorAnd: []interface{}{
			rightIsNull, leftIsFalse,
		}},
		bson.M{mgoOperatorAnd: []interface{}{
			leftIsNull, rightIsFalse,
		}},
	}}

	nullOrTrue := bson.M{mgoOperatorOr: []interface{}{
		bson.M{mgoOperatorAnd: []interface{}{
			rightIsNull, leftIsTrue,
		}},
		bson.M{mgoOperatorAnd: []interface{}{
			leftIsNull, rightIsTrue,
		}},
	}}

	nullOrNull := bson.M{mgoOperatorAnd: []interface{}{
		leftIsNull, rightIsNull,
	}}

	letEvaluation := bson.M{
		mgoOperatorCond: []interface{}{
			nullOrNull,
			mgoNullLiteral,
			wrapInCond(
				mgoNullLiteral,
				wrapInCond(
					true,
					wrapInNullCheckedCond(
						mgoNullLiteral,
						bson.M{
							mgoOperatorOr: []interface{}{"$$left", "$$right"},
						},
						"$$left", "$$right",
					),
					nullOrTrue,
				),
				nullOrFalse,
			),
		},
	}

	return wrapInLet(letAssignment, letEvaluation), true

}

func (or *SQLOrExpr) ToMatchLanguage(t *pushDownTranslator) (bson.M, SQLExpr) {
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

	cond := []interface{}{}

	if v, ok := left[mgoOperatorOr]; ok {
		array := v.([]interface{})
		cond = append(cond, array...)
	} else {
		cond = append(cond, left)
	}

	if v, ok := right[mgoOperatorOr]; ok {
		array := v.([]interface{})
		cond = append(cond, array...)
	} else {
		cond = append(cond, right)
	}

	return bson.M{mgoOperatorOr: cond}, nil
}

func (*SQLOrExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}

//
// SQLRegexExpr evaluates to true if the operand matches the regex patttern.
//
type SQLRegexExpr struct {
	operand, pattern SQLExpr
}

func (reg *SQLRegexExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	operandVal, err := reg.operand.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	patternVal, err := reg.pattern.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	if hasNullValue(operandVal, patternVal) {
		return SQLNull, nil
	}

	pattern, patternOK := patternVal.(SQLVarchar)
	if patternOK {
		matcher, err := regexp.CompilePOSIX(pattern.String())
		if err != nil {
			return SQLFalse, err
		}
		match := matcher.Find([]byte(operandVal.String()))
		if match != nil {
			return SQLTrue, nil
		}
	}
	return SQLFalse, nil
}

func (reg *SQLRegexExpr) String() string {
	return fmt.Sprintf("%s matches %s", reg.operand.String(), reg.pattern.String())
}

func (reg *SQLRegexExpr) ToMatchLanguage(t *pushDownTranslator) (bson.M, SQLExpr) {
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
	return bson.M{name: bson.M{"$regex": bson.RegEx{Pattern: pattern.String(), Options: ""}}}, nil
}

func (*SQLRegexExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}

//
// SQLSubqueryCmpExpr evaluates to true if left is in any of the
// rows returned by the SQLSubqueryExpr expression results.
//
type SQLSubqueryCmpExpr struct {
	subqueryOp   subqueryOp
	left         SQLExpr
	subqueryExpr *SQLSubqueryExpr
	operator     string
}

func (sc *SQLSubqueryCmpExpr) Evaluate(ctx *EvalCtx) (value SQLValue, err error) {

	left, err := sc.left.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	var iter Iter
	defer func() {
		if iter != nil {
			if err == nil {
				err = iter.Close()
			} else {
				iter.Close()
			}

			if err == nil {
				err = iter.Err()
			}
		}
	}()

	execCtx := ctx.ExecutionCtx

	if sc.subqueryExpr.correlated {
		execCtx = ctx.CreateChildExecutionCtx()
	}

	if iter, err = sc.subqueryExpr.plan.Open(execCtx); err != nil {
		return SQLFalse, err
	}

	row := &Row{}

	mismatch, allMatch := true, true

	switch sc.subqueryOp {
	case subqueryAll, subqueryNotIn:
		mismatch = false
	}

	var leftLen int
	leftValues, lvsOk := left.(*SQLValues)
	if lvsOk {
		leftLen = len(leftValues.Values)
	} else {
		leftLen = 1
	}

	right := &SQLValues{}

	for iter.Next(row) {

		values := row.GetValues()

		for _, value := range values {
			right.Values = append(right.Values, value)
		}

		// Make sure the subquery returns the same number of columns as what it's being compared to
		if leftLen != len(values) {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ER_OPERAND_COLUMNS, leftLen)
		}

		switch sc.subqueryOp {
		case subqueryAll:
			expr, err := comparisonExpr(left, right, sc.operator)
			if err != nil {
				return SQLFalse, err
			}
			matches, err := Matches(expr, ctx)
			if err != nil {
				return SQLFalse, err
			}
			if !matches {
				allMatch = false
			}
		case subqueryAny, subquerySome:
			expr, err := comparisonExpr(left, right, sc.operator)
			if err != nil {
				return SQLFalse, err
			}
			matches, err := Matches(expr, ctx)
			if err != nil {
				return SQLFalse, err
			}
			if matches {
				return SQLTrue, nil
			}
		case subqueryIn:
			eq := &SQLEqualsExpr{left, right}
			matches, err := Matches(eq, ctx)
			if err != nil {
				return SQLFalse, err
			}
			if matches {
				return SQLTrue, nil
			}
		case subqueryNotIn:
			neq := &SQLNotEqualsExpr{left, right}
			matches, err := Matches(neq, ctx)
			if err != nil {
				return SQLFalse, err
			}
			if !matches {
				mismatch = true
			}
		}
		row, right = &Row{}, &SQLValues{}
	}

	return NewSQLBool(!mismatch && allMatch), err
}

func (sc *SQLSubqueryCmpExpr) String() string {
	switch sc.subqueryOp {
	case subqueryAll:
		return fmt.Sprintf("%v %v all %v", sc.left, sc.operator, sc.subqueryExpr)
	case subqueryAny:
		return fmt.Sprintf("%v %v any %v", sc.left, sc.operator, sc.subqueryExpr)
	case subqueryIn:
		return fmt.Sprintf("%v in %v", sc.left, sc.subqueryExpr)
	case subqueryNotIn:
		return fmt.Sprintf("%v not in %v", sc.left, sc.subqueryExpr)
	case subquerySome:
		return fmt.Sprintf("%v %v some %v", sc.left, sc.operator, sc.subqueryExpr)
	}
	return ""
}

func (*SQLSubqueryCmpExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}

//
// SQLSubqueryExpr is a wrapper around a parser.SelectStatement representing
// a subquery.
//
type SQLSubqueryExpr struct {
	correlated bool
	allowRows  bool
	plan       PlanStage
}

func (se *SQLSubqueryExpr) Evaluate(evalCtx *EvalCtx) (value SQLValue, err error) {

	var iter Iter
	defer func() {
		if iter != nil {
			if err == nil {
				err = iter.Close()
			} else {
				iter.Close()
			}

			if err == nil {
				err = iter.Err()
			}
		}
	}()

	execCtx := evalCtx.ExecutionCtx
	plan := se.plan

	if se.correlated {
		execCtx = evalCtx.CreateChildExecutionCtx()
		newPlan, err := replaceColumnWithConstant(plan, execCtx)
		if err != nil {
			return nil, err
		}
		plan, ok := newPlan.(PlanStage)
		if !ok {
			return nil, fmt.Errorf("replaceColumnWithConstant returned "+
				" something that is not a PlanStage: %T", newPlan)
		}
		plan = OptimizePlan(execCtx, plan)
	}
	iter, err = plan.Open(execCtx)
	if err != nil {
		return nil, err
	}

	row := &Row{}

	hasNext := iter.Next(row)

	// Filter has to check the entire source to return an accurate 'hasNext'
	if hasNext && iter.Next(&Row{}) {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ER_SUBQUERY_NO_1_ROW)
	}

	values := row.GetValues()
	if len(values) == 0 {
		return SQLNone, nil
	}

	eval := &SQLValues{}
	for _, value := range values {
		eval.Values = append(eval.Values, value)
	}

	if len(eval.Values) == 1 {
		return eval.Values[0], nil
	}
	return eval, nil
}

func (se *SQLSubqueryExpr) Exprs() []SQLExpr {
	exprs := []SQLExpr{}
	for _, c := range se.plan.Columns() {
		exprs = append(exprs, SQLColumnExpr{
			selectID:   c.SelectID,
			tableName:  c.Table,
			columnName: c.Name,
			columnType: schema.ColumnType{
				SQLType:   c.SQLType,
				MongoType: c.MongoType,
			},
		})
	}

	return exprs
}

func (se *SQLSubqueryExpr) String() string {
	return PrettyPrintPlan(se.plan)
}

func (se *SQLSubqueryExpr) Type() schema.SQLType {
	columns := se.plan.Columns()
	if len(columns) == 1 {
		return columns[0].SQLType
	}

	return schema.SQLTuple
}

//
// SQLSubtractExpr evaluates to the difference of the left expression minus the right expressions.
//
type SQLSubtractExpr sqlBinaryNode

func (sub *SQLSubtractExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftVal, err := sub.left.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	rightVal, err := sub.right.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	if hasNullValue(leftVal, rightVal) {
		return SQLNull, nil
	}

	return doArithmetic(leftVal, rightVal, SUB)
}

func (sub *SQLSubtractExpr) String() string {
	return fmt.Sprintf("%v-%v", sub.left, sub.right)
}

func (sub *SQLSubtractExpr) ToAggregationLanguage(t *pushDownTranslator) (interface{}, bool) {
	left, ok := t.ToAggregationLanguage(sub.left)
	if !ok {
		return nil, false
	}

	right, ok := t.ToAggregationLanguage(sub.right)
	if !ok {
		return nil, false
	}

	return bson.M{"$subtract": []interface{}{left, right}}, true
}

func (sub *SQLSubtractExpr) Type() schema.SQLType {
	return schema.SQLFloat
}

//
// SQLTupleExpr represents a tuple.
//
type SQLTupleExpr struct {
	Exprs []SQLExpr
}

func (te SQLTupleExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	var values []SQLValue

	for _, v := range te.Exprs {
		value, err := v.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}

	if len(values) == 1 {
		return values[0], nil
	}

	return &SQLValues{values}, nil
}

func (te *SQLTupleExpr) Normalize() node {
	if len(te.Exprs) == 1 {
		return te.Exprs[0]
	}

	return te
}

func (te SQLTupleExpr) String() string {
	prefix := "("
	for i, expr := range te.Exprs {
		prefix += fmt.Sprintf("%v", expr.String())
		if i != len(te.Exprs)-1 {
			prefix += ", "
		}
	}
	prefix += ")"
	return prefix
}

func (te *SQLTupleExpr) ToAggregationLanguage(t *pushDownTranslator) (interface{}, bool) {
	var transExprs []interface{}

	for _, expr := range te.Exprs {
		transExpr, ok := t.ToAggregationLanguage(expr)
		if !ok {
			return nil, false
		}
		transExprs = append(transExprs, transExpr)
	}

	return transExprs, true

}

func (te SQLTupleExpr) Type() schema.SQLType {
	if len(te.Exprs) == 1 {
		return te.Exprs[0].Type()
	}

	return schema.SQLTuple
}

//
// SQLUnaryMinusExpr evaluates to the negation of the expression.
//
type SQLUnaryMinusExpr sqlUnaryNode

func (um *SQLUnaryMinusExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	if val, err := um.operand.Evaluate(ctx); err == nil {
		if val == SQLNull {
			return SQLNull, nil
		}
		difference, _ := NewSQLValue(-val.Float64(), um.Type(), schema.SQLNone)
		return difference, nil
	}
	return nil, fmt.Errorf("UnaryMinus expression does not apply to a %T", um.operand)
}

func (um *SQLUnaryMinusExpr) Normalize() node {
	if um.operand == SQLNull {
		return SQLNull
	}

	if um.operand == SQLTrue {
		return SQLInt(-1)
	}

	if um.operand == SQLFalse {
		return SQLInt(0)
	}

	return um
}

func (um *SQLUnaryMinusExpr) String() string {
	return fmt.Sprintf("-%v", um.operand)
}

func (um *SQLUnaryMinusExpr) ToAggregationLanguage(t *pushDownTranslator) (interface{}, bool) {
	operand, ok := t.ToAggregationLanguage(um.operand)
	if !ok {
		return nil, false
	}

	letAssignment := bson.M{
		"operand": operand,
	}

	letEvaluation := wrapInNullCheckedCond(
		nil,
		bson.M{"$multiply": []interface{}{-1, "$$operand"}},
		"$$operand",
	)

	return wrapInLet(letAssignment, letEvaluation), true

}

func (um *SQLUnaryMinusExpr) Type() schema.SQLType {
	return um.operand.Type()
}

//
// SQLUnaryTildeExpr invert all bits in the operand
// and returns an unsigned 64-bit integer.
//
type SQLUnaryTildeExpr sqlUnaryNode

func (td *SQLUnaryTildeExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	expr, err := td.operand.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	if v, ok := expr.(SQLValue); ok {
		return SQLUint64(^uint64(v.Int64())), nil
	}

	return SQLUint64(^uint64(0)), nil
}

func (td *SQLUnaryTildeExpr) Normalize() node {
	if v, ok := td.operand.(SQLValue); ok {
		return SQLUint64(^uint64(v.Int64()))
	}
	return td
}

func (td *SQLUnaryTildeExpr) String() string {
	return fmt.Sprintf("~%v", td.operand)
}

func (td *SQLUnaryTildeExpr) Type() schema.SQLType {
	return td.operand.Type()
}

//
// SQLVariableExpr represents a variable lookup.
//
type SQLVariableExpr struct {
	Name    string
	Kind    variable.Kind
	Scope   variable.Scope
	sqlType schema.SQLType
}

func (v *SQLVariableExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	value, err := ctx.Variables().Get(variable.Name(v.Name), v.Scope, v.Kind)
	if err != nil {
		return nil, err
	}

	return NewSQLValueFromSQLColumnExpr(value.Value, value.SQLType, schema.MongoNone)
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

func (v *SQLVariableExpr) Type() schema.SQLType {
	return v.sqlType
}

//
// SQLXorExpr evaluates to true if and only if one of its children evaluates to true.
//
type SQLXorExpr sqlBinaryNode

func (xor *SQLXorExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	left, err := xor.left.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	right, err := xor.right.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if hasNullValue(left, right) {
		return SQLNull, nil
	}

	if (isFalsy(left) && isTruthy(right)) || (isTruthy(left) && isFalsy(right)) {
		return SQLTrue, nil
	}

	return SQLFalse, nil
}

func (xor *SQLXorExpr) Normalize() node {
	left, leftOk := xor.left.(SQLValue)
	if leftOk {
		if isTruthy(left) {
			return &SQLNotExpr{xor.right}
		} else if isFalsy(left) {
			return &SQLOrExpr{SQLFalse, xor.right}
		}
	}

	right, rightOk := xor.right.(SQLValue)
	if rightOk {
		if isTruthy(right) {
			return &SQLNotExpr{xor.left}
		} else if isFalsy(right) {
			return &SQLOrExpr{SQLFalse, xor.left}
		}
	}

	return xor
}

func (xor *SQLXorExpr) String() string {
	return fmt.Sprintf("%v xor %v", xor.left, xor.right)
}

func (xor *SQLXorExpr) ToAggregationLanguage(t *pushDownTranslator) (interface{}, bool) {
	left, ok := t.ToAggregationLanguage(xor.left)
	if !ok {
		return nil, false
	}

	right, ok := t.ToAggregationLanguage(xor.right)
	if !ok {
		return nil, false
	}

	letAssignment := bson.M{
		"left": left, "right": right,
	}

	letEvaluation := bson.M{
		mgoOperatorCond: []interface{}{
			bson.M{
				mgoOperatorOr: []interface{}{
					bson.M{
						mgoOperatorEq: []interface{}{
							bson.M{
								mgoOperatorIfNull: []interface{}{"$$left", nil}},
							nil,
						},
					},
					bson.M{
						mgoOperatorEq: []interface{}{
							bson.M{
								mgoOperatorIfNull: []interface{}{"$$right", nil}},
							nil,
						},
					},
				},
			},
			mgoNullLiteral,
			bson.M{
				mgoOperatorAnd: []interface{}{
					bson.M{
						mgoOperatorOr: []interface{}{"$$left", "$$right"}},
					bson.M{
						mgoOperatorNot: bson.M{
							mgoOperatorAnd: []interface{}{"$$left", "$$right"},
						},
					},
				},
			},
		},
	}

	return wrapInLet(letAssignment, letEvaluation), true

}

func (*SQLXorExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}

// VariableKind indicates if the variable is a system variable or a user variable.
type VariableKind string

func (k VariableKind) scopeAndKind() (variable.Scope, variable.Kind) {
	scope := variable.SessionScope
	kind := variable.SystemKind
	switch k {
	case GlobalStatus:
		kind = variable.StatusKind
		fallthrough
	case GlobalVariable:
		scope = variable.GlobalScope
	case SessionStatus:
		kind = variable.StatusKind
	case UserVariable:
		kind = variable.UserKind
	}

	return scope, kind
}

// caseCondition holds a matcher used in evaluating case expressions and
// a value to return if a particular case is matched. If a case is matched,
// the corresponding 'then' value is evaluated and returned ('then'
// corresponds to the 'then' clause in a case expression).
type caseCondition struct {
	matcher SQLExpr
	then    SQLExpr
}

func (c *caseCondition) String() string {
	return fmt.Sprintf("when (%v) then %v", c.matcher, c.then)
}

type sqlBinaryNode struct {
	left, right SQLExpr
}

type sqlUnaryNode struct {
	operand SQLExpr
}

type subqueryOp int

// Matches checks if a given SQLExpr is "truthy" by coercing it to a boolean value.
// - booleans: the result is simply that same return value
// - numeric values: the result is true if and only if the value is non-zero.
// - strings, the result is true if and only if that string can be parsed as a number,
//   and that number is non-zero.
func Matches(expr SQLExpr, ctx *EvalCtx) (bool, error) {
	eval, err := expr.Evaluate(ctx)
	if err != nil {
		return false, err
	}

	switch v := eval.(type) {
	case SQLBool:
		return v.Bool(), nil
	case SQLInt, SQLFloat, SQLUint32, SQLUint64:
		return v.Float64() != float64(0), nil
	case SQLDecimal128:
		return decimal.Zero.Equals(v.Decimal128()), nil
	case SQLVarchar:
		// more info: http://stackoverflow.com/questions/12221211/how-does-string-truthiness-work-in-mysql
		p, err := strconv.ParseFloat(string(v), 64)
		if err == nil {
			return p != float64(0), nil
		}
		return false, nil
	}

	// TODO - handle other types with possible values that are "truthy" : dates, etc?
	return false, nil
}

func NewSQLAddExpr(left, right SQLExpr) *SQLAddExpr {
	reconciled := convertAllExprs([]SQLExpr{left, right}, schema.SQLFloat, SQLNone)
	return &SQLAddExpr{reconciled[0], reconciled[1]}
}

// NewSQLColumnExpr creates a new SQLColumnExpr with its required fields.
func NewSQLColumnExpr(selectID int, databaseName, tableName, columnName string, sqlType schema.SQLType, mongoType schema.MongoType) SQLColumnExpr {
	return SQLColumnExpr{selectID: selectID,
		databaseName: databaseName,
		tableName:    tableName,
		columnName:   columnName,
		columnType: schema.ColumnType{SQLType: sqlType,
			MongoType: mongoType}}
}

func NewSQLDivideExpr(left, right SQLExpr) *SQLDivideExpr {
	reconciled := convertAllExprs([]SQLExpr{left, right}, schema.SQLFloat, SQLNone)
	return &SQLDivideExpr{reconciled[0], reconciled[1]}
}

func isBooleanColumnAndArithmetic(left, right SQLExpr) bool {
	if _, ok := left.(SQLColumnExpr); !ok {
		return false
	}

	if left.Type() != schema.SQLBoolean {
		return false
	}

	if _, ok := right.(SQLArithmetic); !ok {
		return false
	}

	if _, ok := right.(SQLBool); ok {
		return false
	}

	return true
}

func NewSQLIDivideExpr(left, right SQLExpr) *SQLIDivideExpr {
	reconciled := convertAllExprs([]SQLExpr{left, right}, schema.SQLFloat, SQLNone)
	return &SQLIDivideExpr{reconciled[0], reconciled[1]}
}

func NewSQLIsExpr(left, right SQLExpr) *SQLIsExpr {
	if right.Type() == schema.SQLBoolean {
		switch left.Type() {
		case schema.SQLBoolean, schema.SQLInt, schema.SQLInt64, schema.SQLUint64, schema.SQLNumeric, schema.SQLDecimal128, schema.SQLFloat:
			// don't reconcile the types
		default:
			reconciled := convertAllExprs([]SQLExpr{left, right}, schema.SQLBoolean, SQLNone)
			return &SQLIsExpr{reconciled[0], reconciled[1]}
		}
	}
	return &SQLIsExpr{left, right}
}

func NewSQLModExpr(left, right SQLExpr) *SQLModExpr {
	reconciled := convertAllExprs([]SQLExpr{left, right}, schema.SQLFloat, SQLNone)
	return &SQLModExpr{reconciled[0], reconciled[1]}
}

func NewSQLMultiplyExpr(left, right SQLExpr) *SQLMultiplyExpr {
	reconciled := convertAllExprs([]SQLExpr{left, right}, schema.SQLFloat, SQLNone)
	return &SQLMultiplyExpr{reconciled[0], reconciled[1]}
}

func NewSQLSubtractExpr(left, right SQLExpr) *SQLSubtractExpr {
	reconciled := convertAllExprs([]SQLExpr{left, right}, schema.SQLFloat, SQLNone)
	return &SQLSubtractExpr{reconciled[0], reconciled[1]}
}

func NewSQLVariableExpr(name string, kind variable.Kind,
	scope variable.Scope, sqlType schema.SQLType) *SQLVariableExpr {
	return &SQLVariableExpr{
		Name:    name,
		Kind:    kind,
		Scope:   scope,
		sqlType: sqlType,
	}
}
