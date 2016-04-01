package evaluator

import (
	"fmt"
	"github.com/deafgoat/mixer/sqlparser"
	"regexp"
	"strconv"

	"github.com/10gen/sqlproxy/schema"
)

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
	stmt sqlparser.SelectStatement
}

func (em *SQLExistsExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	ctx.ExecCtx.Depth += 1

	var plan PlanStage
	var it Iter

	plan, err := PlanQuery(ctx.ExecCtx.PlanCtx, em.stmt)
	if err != nil {
		return SQLFalse, err
	}

	var matches bool

	defer func() {
		if it != nil && err == nil {
			err = it.Err()
		}

		// add context to error
		if err != nil {
			err = fmt.Errorf("ExistsMatcher (%v): %v", ctx.ExecCtx.Depth, err)
		}

		ctx.ExecCtx.Depth -= 1

	}()

	it, err = plan.Open(ctx.ExecCtx)
	if err != nil {
		return SQLFalse, err
	}

	if it.Next(&Row{}) {
		matches = true
	}

	return SQLBool(matches), it.Close()
}

func (em *SQLExistsExpr) String() string {
	buf := sqlparser.NewTrackedBuffer(nil)

	switch stmt := em.stmt.(type) {
	case *sqlparser.Select:
		stmt.Format(buf)
	case *sqlparser.SimpleSelect:
		stmt.Format(buf)
	case *sqlparser.Union:
		stmt.Format(buf)
	}

	return fmt.Sprintf("exists %v", buf.String())
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
			return SQLFalse, fmt.Errorf("right 'in' expression is type %T - expeccted tuple", right)
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
		return SQLFalse, err
	}

	data, err := sqlValueToString(value)
	if err != nil {
		return SQLFalse, err
	}

	value, err = l.right.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	pattern, ok := value.(SQLVarchar)
	if !ok {
		return SQLFalse, nil
	}

	// TODO: Golang's regexp package expects a regex pattern
	// for matching but MySQL's 'LIKE' operator doesn't exactly
	// work the same way.
	matches, err := regexp.Match(string(pattern), []byte(data))
	if err != nil {
		return SQLFalse, err
	}

	return SQLBool(matches), nil
}

func (l *SQLLikeExpr) String() string {
	return fmt.Sprintf("%v like %v", l.left, l.right)
}

func (_ *SQLLikeExpr) Type() schema.SQLType {
	return schema.SQLBoolean
}

func sqlValueToString(sqlValue SQLValue) (string, error) {
	switch v := sqlValue.(type) {
	case SQLVarchar:
		return string(v), nil
	case SQLNumeric:
		switch t := v.(type) {
		case SQLFloat:
			return strconv.FormatFloat(t.Float64(), 'f', -1, 64), nil
		default:
			return strconv.FormatInt(int64(t.Float64()), 10), nil
		}
	}

	// TODO: just return empty string with no error?
	return "", fmt.Errorf("unable to convert %v to string", sqlValue)
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
// rows returne by the SQLSubqueryExpr expression results.
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

	var plan PlanStage
	var it Iter

	ctx.ExecCtx.Depth += 1

	plan, err = PlanQuery(ctx.ExecCtx.PlanCtx, sc.value.stmt)
	if err != nil {
		return SQLFalse, err
	}

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

		if err != nil {
			err = fmt.Errorf("SubqueryCmp (%v): %v", ctx.ExecCtx.Depth, err)
		}

		ctx.ExecCtx.Depth -= 1

	}()

	if it, err = plan.Open(ctx.ExecCtx); err != nil {
		return SQLFalse, err
	}

	row := &Row{}

	matched := false

	right := &SQLValues{}
	for it.Next(row) {

		values := row.GetValues(plan.OpFields())

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

// hasNoSQLValue returns true if any of the value in values
// is of type SQLNoValue.
func hasNoSQLValue(values ...SQLValue) bool {
	for _, value := range values {
		switch v := value.(type) {
		case SQLNoValue:
			return true
		case *SQLValues:
			if hasNoSQLValue(v.Values...) {
				return true
			}
		}
	}
	return false
}
