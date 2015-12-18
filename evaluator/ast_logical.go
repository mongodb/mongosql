package evaluator

import (
	"fmt"
	"github.com/deafgoat/mixer/sqlparser"
	"regexp"
	"strconv"
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

//
// SQLEqualsExpr evaluates to true if the left equals the right.
//
type SQLEqualsExpr sqlBinaryNode

func (eq *SQLEqualsExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftEvald, err := eq.left.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	rightEvald, err := eq.right.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	if _, ok := rightEvald.(SQLValues); ok {
		c, err := rightEvald.CompareTo(leftEvald)
		if err == nil {
			return SQLBool(c == 0), nil
		}
		return SQLFalse, err
	}

	c, err := leftEvald.CompareTo(rightEvald)
	if err == nil {
		return SQLBool(c == 0), nil
	}

	return SQLFalse, err

}

//
// SQLExistsExpr evaluates to true if any result is returned from the subquery.
//
type SQLExistsExpr struct {
	stmt sqlparser.SelectStatement
}

func (em *SQLExistsExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	ctx.ExecCtx.Depth += 1

	operator, err := PlanQuery(ctx.ExecCtx, em.stmt)
	if err != nil {
		return SQLFalse, err
	}

	var matches bool

	defer func() {
		if err == nil {
			err = operator.Err()
		}

		// add context to error
		if err != nil {
			err = fmt.Errorf("ExistsMatcher (%v): %v", ctx.ExecCtx.Depth, err)
		}

		ctx.ExecCtx.Depth -= 1

	}()

	if err := operator.Open(ctx.ExecCtx); err != nil {
		return SQLFalse, err
	}

	if operator.Next(&Row{}) {
		matches = true
	}

	return SQLBool(matches), operator.Close()
}

//
// SQLGreaterThanExpr evaluates to true when the left is greater than the right.
//
type SQLGreaterThanExpr sqlBinaryNode

func (gt *SQLGreaterThanExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftEvald, err := gt.left.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	rightEvald, err := gt.right.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	if _, ok := rightEvald.(SQLValues); ok {
		c, err := rightEvald.CompareTo(leftEvald)
		if err == nil {
			return SQLBool(c < 0), nil
		}
		return SQLFalse, err
	}

	c, err := leftEvald.CompareTo(rightEvald)
	if err == nil {
		return SQLBool(c > 0), nil
	}
	return SQLFalse, err
}

//
// SQLGreaterThanOrEqualExpr evaluates to true when the left is greater than or equal to the right.
//
type SQLGreaterThanOrEqualExpr sqlBinaryNode

func (gte *SQLGreaterThanOrEqualExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftEvald, err := gte.left.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	rightEvald, err := gte.right.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	if _, ok := rightEvald.(SQLValues); ok {
		c, err := rightEvald.CompareTo(leftEvald)
		if err == nil {
			return SQLBool(c <= 0), nil
		}
		return SQLFalse, err
	}

	c, err := leftEvald.CompareTo(rightEvald)
	if err == nil {
		return SQLBool(c >= 0), nil
	}

	return SQLFalse, err
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
	rightChild, ok := right.(SQLValues)
	if !ok {
		return SQLFalse, fmt.Errorf("right In expression is %T", right)
	}

	leftChild, ok := left.(SQLValues)
	if ok {
		if len(leftChild) != 1 {
			return SQLFalse, fmt.Errorf("left operand should contain 1 column")
		}
		left = leftChild[0]
	}

	for _, right := range rightChild {
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

//
// SQLLessThanExpr evaluates to true when the left is less than the right.
//
type SQLLessThanExpr sqlBinaryNode

func (lt *SQLLessThanExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftEvald, err := lt.left.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	rightEvald, err := lt.right.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	if _, ok := rightEvald.(SQLValues); ok {
		c, err := rightEvald.CompareTo(leftEvald)
		if err == nil {
			return SQLBool(c > 0), nil
		}
		return SQLFalse, err
	}

	c, err := leftEvald.CompareTo(rightEvald)
	if err == nil {
		return SQLBool(c < 0), nil
	}
	return SQLFalse, err
}

//
// SQLLessThanOrEqualExpr evaluates to true when the left is less than or equal to the right.
//
type SQLLessThanOrEqualExpr sqlBinaryNode

func (lte *SQLLessThanOrEqualExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftEvald, err := lte.left.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	rightEvald, err := lte.right.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	if _, ok := rightEvald.(SQLValues); ok {
		c, err := rightEvald.CompareTo(leftEvald)
		if err == nil {
			return SQLBool(c >= 0), nil
		}
		return SQLFalse, err
	}

	c, err := leftEvald.CompareTo(rightEvald)
	if err == nil {
		return SQLBool(c <= 0), nil
	}
	return SQLFalse, err
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

	pattern, ok := value.(SQLString)
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

func sqlValueToString(sqlValue SQLValue) (string, error) {
	switch v := sqlValue.(type) {
	case SQLString:
		return string(v), nil
	case SQLInt:
		return string(v), nil // TODO: I don't think this works... you need to use strconv.Itoa
	case SQLUint32:
		return string(v), nil // TODO: same here...
	case SQLFloat:
		return strconv.FormatFloat(float64(v), 'f', -1, 64), nil
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

//
// SQLNotEqualsExpr evaluates to true if the left does not equal the right.
//
type SQLNotEqualsExpr sqlBinaryNode

func (neq *SQLNotEqualsExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftEvald, err := neq.left.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	rightEvald, err := neq.right.Evaluate(ctx)
	if err != nil {
		return SQLFalse, err
	}

	if _, ok := rightEvald.(SQLValues); ok {
		c, err := rightEvald.CompareTo(leftEvald)
		if err == nil {
			return SQLBool(c != 0), nil
		}
		return SQLFalse, err
	}

	c, err := leftEvald.CompareTo(rightEvald)
	if err == nil {
		return SQLBool(c != 0), nil
	}

	return SQLFalse, err
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

//
// SQLSubqueryCmpExpr evaluates to true if ...???
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

	ctx.ExecCtx.Depth += 1

	operator, err := PlanQuery(ctx.ExecCtx, sc.value.stmt)
	if err != nil {
		return SQLFalse, err
	}

	defer func() {
		if err == nil {
			err = operator.Close()
		} else {
			operator.Close()
		}

		if err == nil {
			err = operator.Err()
		}

		if err != nil {
			err = fmt.Errorf("SubqueryCmp (%v): %v", ctx.ExecCtx.Depth, err)
		}

		ctx.ExecCtx.Depth -= 1

	}()

	if err := operator.Open(ctx.ExecCtx); err != nil {
		return SQLFalse, err
	}

	row := &Row{}

	matched := false

	right := SQLValues{}
	for operator.Next(row) {

		values := row.GetValues(operator.OpFields())

		for _, value := range values {
			field, err := NewSQLValue(value, "")
			if err != nil {
				return SQLFalse, err
			}
			right = append(right, field)
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

		row, right = &Row{}, SQLValues{}

	}

	if sc.In {
		matched = true
	}

	return SQLBool(!matched), err
}
