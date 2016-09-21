package evaluator

import (
	"fmt"
	"math"
	"regexp"
	"strconv"

	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
	"github.com/shopspring/decimal"
	"gopkg.in/mgo.v2/bson"
)

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
// SQLArithmetic is used to do arithmetic on all types.
//
type SQLArithmetic interface {
	Decimal128() decimal.Decimal
	Float64() float64
	Int64() int64
	Uint64() uint64
}

// Arithmetic consts
type ArithmeticOperator byte

const (
	ADD ArithmeticOperator = iota
	DIV
	MULT
	SUB
)

//
// SQLValue is a SQLExpr with a value.
//
type SQLValue interface {
	SQLExpr
	SQLArithmetic
	Value() interface{}
}

// A base type for a binary node.
type sqlBinaryNode struct {
	left, right SQLExpr
}

type sqlUnaryNode struct {
	operand SQLExpr
}

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
		return bool(v), nil
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

func (_ *MongoFilterExpr) Type() schema.SQLType {
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

func (and *SQLAndExpr) normalize() node {
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

func (_ *SQLAndExpr) Type() schema.SQLType {
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

//
// SQLCaseExpr holds a number of cases to evaluate as well as the value
// to return if any of the cases is matched. If none is matched,
// 'elseValue' is evaluated and returned.
//
type SQLCaseExpr struct {
	elseValue      SQLExpr
	caseConditions []caseCondition
}

// caseCondition holds a matcher used in evaluating case expressions and
// a value to return if a particular case is matched. If a case is matched,
// the corresponding 'then' value is evaluated and returned ('then'
// corresponds to the 'then' clause in a case expression).
type caseCondition struct {
	matcher SQLExpr
	then    SQLExpr
}

func (s SQLCaseExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	for _, condition := range s.caseConditions {
		m, err := Matches(condition.matcher, ctx)
		if err != nil {
			return nil, err
		}

		if m {
			return condition.then.Evaluate(ctx)
		}
	}

	return s.elseValue.Evaluate(ctx)
}

func (c *caseCondition) String() string {
	return fmt.Sprintf("when (%v) then %v", c.matcher, c.then)
}

func (s SQLCaseExpr) String() string {
	str := fmt.Sprintf("case ")
	for _, cond := range s.caseConditions {
		str += fmt.Sprintf("%v ", cond.String())
	}
	if s.elseValue != nil {
		str += fmt.Sprintf("else %v ", s.elseValue.String())
	}
	str += fmt.Sprintf("end")
	return str
}

func (s SQLCaseExpr) Type() schema.SQLType {
	conds := []SQLExpr{s.elseValue}
	for _, cond := range s.caseConditions {
		conds = append(conds, cond.then)
	}
	return preferentialType(conds...)
}

//
// SQLColumnExpr represents a column reference.
//
type SQLColumnExpr struct {
	selectID   int
	tableName  string
	columnName string
	columnType schema.ColumnType
}

// NewSQLColumnExpr creates a new SQLColumnExpr with its required fields.
func NewSQLColumnExpr(selectID int, tableName, columnName string, sqlType schema.SQLType, mongoType schema.MongoType) SQLColumnExpr {
	return SQLColumnExpr{selectID, tableName, columnName, schema.ColumnType{sqlType, mongoType}}
}

func (c SQLColumnExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	// first check our immediate rows
	for _, row := range ctx.Rows {
		if value, ok := row.GetField(c.selectID, c.tableName, c.columnName); ok {
			return NewSQLValueFromSQLColumnExpr(value, c.columnType.SQLType, c.columnType.MongoType)
		}
	}

	// If we didn't find it there, search in the src rows, which contain parent
	// information in the case we are evaluating a correlated column.
	if ctx.ExecutionCtx != nil {
		for _, row := range ctx.ExecutionCtx.SrcRows {
			if value, ok := row.GetField(c.selectID, c.tableName, c.columnName); ok {
				return NewSQLValueFromSQLColumnExpr(value, c.columnType.SQLType, c.columnType.MongoType)
			}
		}
	}

	return SQLNull, nil
}

func (c SQLColumnExpr) String() string {
	var str string
	if c.tableName != "" {
		str += c.tableName + "."
	}
	str += c.columnName
	return str
}

func (c SQLColumnExpr) Type() schema.SQLType {
	if c.columnType.MongoType == schema.MongoObjectId && c.columnType.SQLType == schema.SQLVarchar {
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
	expr     SQLExpr
	convType schema.SQLType
}

func (ce SQLConvertExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	v, err := ce.expr.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	return NewSQLValue(v.Value(), ce.convType), nil
}

func (ce SQLConvertExpr) String() string {
	return ce.expr.String()
}

func (ce SQLConvertExpr) Type() schema.SQLType {
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

func (div *SQLDivideExpr) String() string {
	return fmt.Sprintf("%v/%v", div.left, div.right)
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

	c, err := CompareTo(leftVal, rightVal)
	if err == nil {
		return SQLBool(c == 0), nil
	}

	return SQLFalse, err
}

func (eq *SQLEqualsExpr) normalize() node {
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

func (_ *SQLEqualsExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}

//
// SQLExistsExpr evaluates to true if any result is returned from the subquery.
//
type SQLExistsExpr struct {
	expr *SQLSubqueryExpr
}

func (em *SQLExistsExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	var it Iter
	var err error
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

	return SQLBool(matches), it.Close()
}

func (em *SQLExistsExpr) String() string {
	return fmt.Sprintf("exists(%s)", em.expr.String())
}

func (_ *SQLExistsExpr) Type() schema.SQLType {
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

	c, err := CompareTo(leftVal, rightVal)
	if err == nil {
		return SQLBool(c > 0), nil
	}
	return SQLFalse, err
}

func (gt *SQLGreaterThanExpr) normalize() node {
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

func (_ *SQLGreaterThanExpr) Type() schema.SQLType {
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

	c, err := CompareTo(leftVal, rightVal)
	if err == nil {
		return SQLBool(c >= 0), nil
	}

	return SQLFalse, err
}

func (gte *SQLGreaterThanOrEqualExpr) normalize() node {
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

func (_ *SQLGreaterThanOrEqualExpr) Type() schema.SQLType {
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

	if leftVal == nil || rightVal == nil {
		return SQLNull, nil
	}

	if rightVal.Float64() == 0 {
		// NOTE: this is per the mysql manual.
		return SQLNull, nil
	}

	return SQLInt(int(leftVal.Float64() / rightVal.Float64())), nil
}

func (div *SQLIDivideExpr) String() string {
	return fmt.Sprintf("%v/%v", div.left, div.right)
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

func (in *SQLInExpr) normalize() node {
	if hasNullExpr(in.left) {
		return SQLNull
	}

	return in
}

func (in *SQLInExpr) String() string {
	return fmt.Sprintf("%v in %v", in.left, in.right)
}

func (_ *SQLInExpr) Type() schema.SQLType {
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
		} else {
			return SQLTrue, nil
		}
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

func (_ *SQLIsExpr) Type() schema.SQLType {
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

	c, err := CompareTo(leftVal, rightVal)
	if err == nil {
		return SQLBool(c < 0), nil
	}
	return SQLFalse, err
}

func (lt *SQLLessThanExpr) normalize() node {
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

func (_ *SQLLessThanExpr) Type() schema.SQLType {
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

	c, err := CompareTo(leftVal, rightVal)
	if err == nil {
		return SQLBool(c <= 0), nil
	}
	return SQLFalse, err
}

func (lte *SQLLessThanOrEqualExpr) normalize() node {
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

func (_ *SQLLessThanOrEqualExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}

//
// SQLLikeExpr evaluates to true if the left is 'like' the right.
//
type SQLLikeExpr sqlBinaryNode

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

	pattern := "(?i)" + convertSQLValueToPattern(value)

	matches, err := regexp.Match(pattern, []byte(data))
	if err != nil {
		return nil, err
	}

	return SQLBool(matches), nil
}

func (l *SQLLikeExpr) normalize() node {
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

func (_ *SQLLikeExpr) Type() schema.SQLType {
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

func (mult *SQLMultiplyExpr) Type() schema.SQLType {
	return schema.SQLFloat
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

func (not *SQLNotExpr) normalize() node {
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

func (_ *SQLNotExpr) Type() schema.SQLType {
	return schema.SQLBoolean
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

	c, err := CompareTo(leftVal, rightVal)
	if err == nil {
		return SQLBool(c != 0), nil
	}

	return SQLFalse, err
}

func (neq *SQLNotEqualsExpr) normalize() node {
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

func (_ *SQLNotEqualsExpr) Type() schema.SQLType {
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

func (or *SQLOrExpr) normalize() node {
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

func (_ *SQLOrExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}

type subqueryOp int

const (
	subqueryAll = iota
	subqueryAny
	subqueryIn
	subqueryNotIn
	subquerySome
)

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

func (_ *SQLRegexExpr) Type() schema.SQLType {
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

func (sc *SQLSubqueryCmpExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	left, err := sc.left.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	var it Iter
	defer func() {
		if it != nil {
			if err == nil {
				err = it.Close()
			} else {
				it.Close()
			}

			if err == nil {
				err = it.Err()
			}
		}
	}()

	execCtx := ctx.ExecutionCtx

	if sc.subqueryExpr.correlated {
		execCtx = ctx.CreateChildExecutionCtx()
	}

	if it, err = sc.subqueryExpr.plan.Open(execCtx); err != nil {
		return SQLFalse, err
	}

	row := &Row{}

	mismatch, allMatch := true, true

	switch sc.subqueryOp {
	case subqueryAll, subqueryNotIn:
		mismatch = false
	}

	right := &SQLValues{}

	for it.Next(row) {

		values := row.GetValues()

		for _, value := range values {
			field, err := NewSQLValueFromSQLColumnExpr(value, schema.SQLNone, schema.MongoNone)
			if err != nil {
				return SQLFalse, err
			}
			right.Values = append(right.Values, field)
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

	return SQLBool(!mismatch && allMatch), err
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

func (_ *SQLSubqueryCmpExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}

//
// SQLSubqueryExpr is a wrapper around a parser.SelectStatement representing
// a subquery.
//
type SQLSubqueryExpr struct {
	correlated bool
	plan       PlanStage
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

func (se *SQLSubqueryExpr) Evaluate(evalCtx *EvalCtx) (value SQLValue, err error) {

	var it Iter
	defer func() {
		if it != nil {
			if err == nil {
				err = it.Close()
			} else {
				it.Close()
			}

			if err == nil {
				err = it.Err()
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
			panic("replaceColumnWithConstant returns something that is not a PlanStage")
		}
		plan, err = OptimizePlan(execCtx, plan)
		if err != nil {
			return nil, err
		}
	}
	it, err = plan.Open(execCtx)
	if err != nil {
		return nil, err
	}

	row := &Row{}

	hasNext := it.Next(row)

	// Filter has to check the entire source to return an accurate 'hasNext'
	if hasNext && it.Next(&Row{}) {
		return nil, fmt.Errorf("Subquery returns more than 1 row")
	}

	values := row.GetValues()
	if len(values) == 0 {
		return SQLNone, nil
	}

	eval := &SQLValues{}
	for _, value := range values {
		field, err := NewSQLValueFromSQLColumnExpr(value, schema.SQLNone, schema.MongoNone)
		if err != nil {
			return nil, err
		}
		eval.Values = append(eval.Values, field)
	}

	if len(eval.Values) == 1 {
		return eval.Values[0], nil
	}
	return eval, nil
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

// constantColumnReplacer holds the execution context, which has the data
// used to replace the column expressions.
type constantColumnReplacer struct {
	ctx *ExecutionCtx
}

// replaceColumnWithConstant kicks off the replacement of column expressions.
func replaceColumnWithConstant(n node, ctx *ExecutionCtx) (node, error) {
	v := &constantColumnReplacer{ctx}
	n, err := v.visit(n)
	return n, err
}

func (v *constantColumnReplacer) visit(n node) (node, error) {
	switch typedN := n.(type) {
	case SQLColumnExpr:
		for _, row := range v.ctx.SrcRows {
			if val, ok := row.GetField(typedN.selectID, typedN.tableName, typedN.columnName); ok {
				return NewSQLValueFromSQLColumnExpr(val, typedN.columnType.SQLType, typedN.columnType.MongoType)
			}
		}
	}
	return walk(v, n)
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

func (te *SQLTupleExpr) normalize() node {
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
		return NewSQLValue(-val.Float64(), um.Type()), nil
	}
	return nil, fmt.Errorf("UnaryMinus expression does not apply to a %T", um.operand)
}

func (um *SQLUnaryMinusExpr) String() string {
	return fmt.Sprintf("-%v", um.operand)
}

func (um *SQLUnaryMinusExpr) Type() schema.SQLType {
	return um.operand.Type()
}

func (um *SQLUnaryMinusExpr) normalize() node {
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

func (td *SQLUnaryTildeExpr) normalize() node {
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

// VariableKind indicates if the variable is a system variable or a user variable.
type VariableKind string

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

//
// SQLVariableExpr represents a variable lookup.
//
type SQLVariableExpr struct {
	Name  string
	Kind  variable.Kind
	Scope variable.Scope
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
	return schema.MongoNone
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

func (xor *SQLXorExpr) normalize() node {
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

func (_ *SQLXorExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}
