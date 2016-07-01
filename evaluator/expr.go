package evaluator

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/10gen/sqlproxy/schema"
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
// SQLValue is a SQLExpr with a value.
//
type SQLValue interface {
	SQLExpr
	Value() interface{}
}

//
// SQLNumeric is a numeric SQLValue.
//
type SQLNumeric interface {
	SQLValue
	Add(o SQLNumeric) SQLNumeric
	Sub(o SQLNumeric) SQLNumeric
	Product(o SQLNumeric) SQLNumeric
	Float64() float64
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
	case SQLNumeric:
		return v.Float64() != float64(0), nil
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
// SQLAddExpr evaluates to the sum of two expressions.
//
type SQLAddExpr sqlBinaryNode

func (add *SQLAddExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftVal, err := convertToSQLNumeric(add.left, ctx)
	if err != nil {
		return nil, err
	}

	rightVal, err := convertToSQLNumeric(add.right, ctx)
	if err != nil {
		return nil, err
	}

	if leftVal == nil || rightVal == nil {
		return SQLNull, nil
	}

	return leftVal.Add(rightVal), nil
}

func (add *SQLAddExpr) String() string {
	return fmt.Sprintf("%v+%v", add.left, add.right)
}

func (add *SQLAddExpr) Type() schema.SQLType {
	return preferentialType(add.left, add.right)
}

//
// SQLAndExpr evaluates to true if and only if all its children evaluate to true.
//
type SQLAndExpr sqlBinaryNode

func (and *SQLAndExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftMatches, err := Matches(and.left, ctx)
	if err != nil {
		return nil, err
	}

	rightMatches, err := Matches(and.right, ctx)
	if err != nil {
		return nil, err
	}

	if leftMatches && rightMatches {
		return SQLTrue, nil
	}

	return SQLFalse, nil
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

	err = ctx.SetVariable(e.variable.Name, value, e.variable.Kind)
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
			return NewSQLValue(value, c.columnType.SQLType, c.columnType.MongoType)
		}
	}

	// If we didn't find it there, search in the src rows, which contain parent
	// information in the case we are evaluating a correlated column.
	if ctx.ExecutionCtx != nil {
		for _, row := range ctx.ExecutionCtx.SrcRows {
			if value, ok := row.GetField(c.selectID, c.tableName, c.columnName); ok {
				return NewSQLValue(value, c.columnType.SQLType, c.columnType.MongoType)
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
	return c.columnType.SQLType
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
	return NewSQLValue(v.Value(), ce.convType, schema.MongoNone)
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
	leftVal, err := convertToSQLNumeric(div.left, ctx)
	if err != nil {
		return nil, err
	}

	rightVal, err := convertToSQLNumeric(div.right, ctx)
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

	return SQLFloat(leftVal.Float64() / rightVal.Float64()), nil
}

func (div *SQLDivideExpr) String() string {
	return fmt.Sprintf("%v/%v", div.left, div.right)
}

func (div *SQLDivideExpr) Type() schema.SQLType {
	return preferentialType(div.left, div.right)
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

	if hasNoSQLValue(leftVal, rightVal) {
		return SQLFalse, nil
	}

	c, err := CompareTo(leftVal, rightVal)
	if err == nil {
		return SQLBool(c == 0), nil
	}

	return SQLFalse, err
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

	if hasNoSQLValue(leftVal, rightVal) {
		return SQLFalse, nil
	}

	c, err := CompareTo(leftVal, rightVal)
	if err == nil {
		return SQLBool(c > 0), nil
	}
	return SQLFalse, err
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

	if hasNoSQLValue(leftVal, rightVal) {
		return SQLFalse, nil
	}

	c, err := CompareTo(leftVal, rightVal)
	if err == nil {
		return SQLBool(c >= 0), nil
	}

	return SQLFalse, err
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
	leftVal, err := convertToSQLNumeric(div.left, ctx)
	if err != nil {
		return nil, err
	}

	rightVal, err := convertToSQLNumeric(div.right, ctx)
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
	}

	for _, right := range rightChild.Values {
		eq := &SQLEqualsExpr{left, right}
		m, err := Matches(eq, ctx)
		if err != nil {
			return SQLFalse, err
		}
		if m {
			return SQLTrue, nil
		}
	}

	return SQLFalse, nil
}

func (in *SQLInExpr) String() string {
	return fmt.Sprintf("%v in %v", in.left, in.right)
}

func (_ *SQLInExpr) Type() schema.SQLType {
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

	if hasNoSQLValue(leftVal, rightVal) {
		return SQLFalse, nil
	}

	c, err := CompareTo(leftVal, rightVal)
	if err == nil {
		return SQLBool(c < 0), nil
	}
	return SQLFalse, err
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

	if hasNoSQLValue(leftVal, rightVal) {
		return SQLFalse, nil
	}

	c, err := CompareTo(leftVal, rightVal)
	if err == nil {
		return SQLBool(c <= 0), nil
	}
	return SQLFalse, err
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
		return SQLValue(SQLInt(0)), err
	}

	if _, ok := value.(SQLNullValue); ok {
		return SQLNull, nil
	}

	if _, ok := value.(SQLNoValue); ok {
		return SQLNull, nil
	}

	data, err := sqlValueToString(value)
	if err != nil {
		return SQLValue(SQLInt(0)), err
	}

	value, err = l.right.Evaluate(ctx)
	if err != nil {
		return SQLValue(SQLInt(0)), err
	}

	if _, ok := value.(SQLNullValue); ok {
		return SQLNull, nil
	}

	if _, ok := value.(SQLNoValue); ok {
		return SQLNull, nil
	}

	patternStr, err := sqlValueToString(value)
	if err != nil {
		return SQLValue(SQLInt(0)), err
	}

	// check if pattern ends with a whitespace or tab
	if strings.HasSuffix(patternStr, " ") {
		patternStr = patternStr[0 : len(patternStr)-1]
		patternStr += "\\s$"
	} else if strings.HasSuffix(patternStr, "\\t") {
		patternStr = patternStr[0 : len(patternStr)-1]
		patternStr += "\\t$"
	}

	if !strings.HasPrefix(patternStr, "_") && !strings.HasPrefix(patternStr, "%") {
		patternStr = "^" + patternStr
	}

	if !strings.HasSuffix(patternStr, "_") && !strings.HasSuffix(patternStr, "%") {
		patternStr = patternStr + "$"
	}

	patternStr = strings.Replace(patternStr, "_", "[\\w]", -1)
	patternStr = strings.Replace(patternStr, "%", "[\\w]*", -1)

	// (?i) is case insensitive flag
	reg, err := regexp.Compile("(?i)" + patternStr)
	if err != nil {
		return SQLValue(SQLInt(0)), err
	}

	matches := reg.Match([]byte(data))

	if matches {
		return SQLValue(SQLInt(1)), nil
	}
	return SQLValue(SQLInt(0)), nil
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
	leftVal, err := convertToSQLNumeric(mod.left, ctx)
	if err != nil {
		return nil, err
	}

	rightVal, err := convertToSQLNumeric(mod.right, ctx)
	if err != nil {
		return nil, err
	}

	if leftVal == nil || rightVal == nil || rightVal.Float64() == 0 {
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
	leftVal, err := convertToSQLNumeric(mult.left, ctx)
	if err != nil {
		return nil, err
	}

	rightVal, err := convertToSQLNumeric(mult.right, ctx)
	if err != nil {
		return nil, err
	}

	if leftVal == nil || rightVal == nil {
		return SQLNull, nil
	}

	return leftVal.Product(rightVal), nil
}

func (mult *SQLMultiplyExpr) String() string {
	return fmt.Sprintf("%v*%v", mult.left, mult.right)
}

func (mult *SQLMultiplyExpr) Type() schema.SQLType {
	return preferentialType(mult.left, mult.right)
}

//
// SQLNotExpr evaluates to the inverse of its child.
//
type SQLNotExpr sqlUnaryNode

func (not *SQLNotExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	m, err := Matches(not.operand, ctx)
	if err != nil {
		return SQLFalse, err
	}
	return SQLBool(!m), nil
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

	if hasNoSQLValue(leftVal, rightVal) {
		return SQLFalse, nil
	}

	c, err := CompareTo(leftVal, rightVal)
	if err == nil {
		return SQLBool(c != 0), nil
	}

	return SQLFalse, err
}

func (neq *SQLNotEqualsExpr) String() string {
	return fmt.Sprintf("%v != %v", neq.left, neq.right)
}

func (_ *SQLNotEqualsExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}

//
// SQLNullCmpExpr evaluates to true if its value evaluates to null.
//
type SQLNullCmpExpr sqlUnaryNode

func (nm *SQLNullCmpExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	eval, err := nm.operand.Evaluate(ctx)
	if err != nil {
		return SQLFalse, nil
	}
	_, ok := eval.(SQLNullValue)
	return SQLBool(ok), nil
}

func (nm *SQLNullCmpExpr) String() string {
	return fmt.Sprintf("%v is null", nm.operand.String())
}

func (_ *SQLNullCmpExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}

//
// SQLOrExpr evaluates to true if any of its children evaluate to true.
//
type SQLOrExpr sqlBinaryNode

func (or *SQLOrExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftMatches, err := Matches(or.left, ctx)
	if err != nil {
		return nil, err
	}

	rightMatches, err := Matches(or.right, ctx)
	if err != nil {
		return nil, err
	}

	if leftMatches || rightMatches {
		return SQLTrue, nil
	}

	return SQLFalse, nil
}

func (or *SQLOrExpr) String() string {
	return fmt.Sprintf("%v or %v", or.left, or.right)
}

func (_ *SQLOrExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}

//
// SQLSubqueryCmpExpr evaluates to true if left is in any of the
// rows returned by the SQLSubqueryExpr expression results.
//
type SQLSubqueryCmpExpr struct {
	In    bool
	left  SQLExpr
	value *SQLSubqueryExpr
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

	if sc.value.correlated {
		execCtx = ctx.CreateChildExecutionCtx()
	}

	if it, err = sc.value.plan.Open(execCtx); err != nil {
		return SQLFalse, err
	}

	row := &Row{}

	matched := false

	right := &SQLValues{}
	for it.Next(row) {

		values := row.GetValues()

		for _, value := range values {
			field, err := NewSQLValue(value, schema.SQLNone, schema.MongoNone)
			if err != nil {
				return SQLFalse, err
			}
			right.Values = append(right.Values, field)
		}

		eq := &SQLEqualsExpr{left, right}

		matches, err := Matches(eq, ctx)
		if err != nil {
			return SQLFalse, err
		}

		if matches {
			matched = true
			if sc.In {
				return SQLTrue, err
			}
		}

		row, right = &Row{}, &SQLValues{}

	}

	if sc.In {
		matched = true
	}

	return SQLBool(!matched), err
}

func (sc *SQLSubqueryCmpExpr) String() string {
	in := "in"
	if !sc.In {
		in = "not in"
	}
	return fmt.Sprintf("%v %v %v", sc.left, in, sc.value)
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

	if se.correlated {
		execCtx = evalCtx.CreateChildExecutionCtx()
	}

	it, err = se.plan.Open(execCtx)
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
		field, err := NewSQLValue(value, schema.SQLNone, schema.MongoNone)
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

//
// SQLSubtractExpr evaluates to the difference of the left expression minus the right expressions.
//
type SQLSubtractExpr sqlBinaryNode

func (sub *SQLSubtractExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftVal, err := convertToSQLNumeric(sub.left, ctx)
	if err != nil {
		return nil, err
	}

	rightVal, err := convertToSQLNumeric(sub.right, ctx)
	if err != nil {
		return nil, err
	}

	if leftVal == nil || rightVal == nil {
		return SQLNull, nil
	}

	return leftVal.Sub(rightVal), nil
}

func (sub *SQLSubtractExpr) String() string {
	return fmt.Sprintf("%v-%v", sub.left, sub.right)
}

func (sub *SQLSubtractExpr) Type() schema.SQLType {
	return preferentialType(sub.left, sub.right)
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
	val, ok := um.operand.(SQLNumeric)
	if !ok {
		if operand, err := um.operand.Evaluate(ctx); err == nil {
			val, ok = operand.(SQLNumeric)
		}
	}
	if ok {
		return NewSQLValue(-val.Float64(), um.Type(), schema.MongoNone)
	}

	return nil, fmt.Errorf("UnaryMinus expression does not apply to a %T", um.operand)
}

func (um *SQLUnaryMinusExpr) String() string {
	return fmt.Sprintf("-%v", um.operand)
}

func (um *SQLUnaryMinusExpr) Type() schema.SQLType {
	return um.operand.Type()
}

//
// SQLUnaryTildeExpr evaluates to the bitwise complement of the expression.
//
type SQLUnaryTildeExpr sqlUnaryNode

func (td *SQLUnaryTildeExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	if val, ok := td.operand.(SQLNumeric); ok {
		return SQLInt(^round(val.Float64())), nil
	}

	return td.operand.Evaluate(ctx)
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
	// GlobalVariable is a global system variable.
	GlobalVariable VariableKind = "global"
	// SessionVariable is a session(local) variable.
	SessionVariable VariableKind = "session"
	// UserVariable is a custom variable associated with a session(local).
	UserVariable VariableKind = "user"
)

//
// SQLVariableExpr represents a variable lookup.
//
type SQLVariableExpr struct {
	Name string
	Kind VariableKind
}

func (v *SQLVariableExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return ctx.GetVariable(v.Name, v.Kind)
}

func (v *SQLVariableExpr) String() string {
	prefix := ""
	switch v.Kind {
	case GlobalVariable:
		prefix = "@@global."
	case SessionVariable:
		prefix = "@@session."
	case UserVariable:
		prefix = "@"
	}

	return prefix + v.Name
}

func (v *SQLVariableExpr) Type() schema.SQLType {
	return schema.MongoNone
}
