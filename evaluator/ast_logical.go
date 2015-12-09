package evaluator

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"regexp"
	"strconv"
)

//
// Evaluates to true if and only if all its children evaluate to true.
//
type SQLAndExpr struct {
	children []SQLExpr
}

func (and *SQLAndExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	for _, c := range and.children {
		m, err := Matches(c, ctx)
		if err != nil {
			return SQLBool(false), err
		}
		if !m {
			return SQLBool(false), nil
		}
	}
	return SQLBool(true), nil
}

//
// Evaluates to true if the left equals the right.
//
type SQLEqualsExpr SQLBinaryNode

func (eq *SQLEqualsExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftEvald, err := eq.left.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	rightEvald, err := eq.right.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	if _, ok := rightEvald.(SQLValues); ok {
		c, err := rightEvald.CompareTo(leftEvald)
		if err == nil {
			return SQLBool(c == 0), nil
		}
		return SQLBool(false), err
	}

	c, err := leftEvald.CompareTo(rightEvald)
	if err == nil {
		return SQLBool(c == 0), nil
	}

	return SQLBool(false), err

}

//
// Evaluates to true if any result is returned from the subquery.
//
type SQLExistsExpr struct {
	stmt sqlparser.SelectStatement
}

func (em *SQLExistsExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	ctx.ExecCtx.Depth += 1

	operator, err := PlanQuery(ctx.ExecCtx, em.stmt)
	if err != nil {
		return SQLBool(false), err
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
		return SQLBool(false), err
	}

	if operator.Next(&Row{}) {
		matches = true
	}

	return SQLBool(matches), operator.Close()
}

//
// Evaluates to true when the left is greater than the right.
//
type SQLGreaterThanExpr SQLBinaryNode

func (gt *SQLGreaterThanExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftEvald, err := gt.left.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	rightEvald, err := gt.right.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	if _, ok := rightEvald.(SQLValues); ok {
		c, err := rightEvald.CompareTo(leftEvald)
		if err == nil {
			return SQLBool(c < 0), nil
		}
		return SQLBool(false), err
	}

	c, err := leftEvald.CompareTo(rightEvald)
	if err == nil {
		return SQLBool(c > 0), nil
	}
	return SQLBool(false), err
}

//
// Evaluates to true when the left is greater than or equal to the right.
//
type SQLGreaterThanOrEqualExpr SQLBinaryNode

func (gte *SQLGreaterThanOrEqualExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftEvald, err := gte.left.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	rightEvald, err := gte.right.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	if _, ok := rightEvald.(SQLValues); ok {
		c, err := rightEvald.CompareTo(leftEvald)
		if err == nil {
			return SQLBool(c <= 0), nil
		}
		return SQLBool(false), err
	}

	c, err := leftEvald.CompareTo(rightEvald)
	if err == nil {
		return SQLBool(c >= 0), nil
	}

	return SQLBool(false), err
}

//
// Evaluates to true if the left is in any of the values on the right.
//
type SQLInExpr SQLBinaryNode

func (in *SQLInExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	left, err := in.left.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	right, err := in.right.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	// right child must be of type SQLValues
	// TODO: can we not simply require this as part of the node definition?
	rightChild, ok := right.(SQLValues)
	if !ok {
		return SQLBool(false), fmt.Errorf("right In expression is %T", right)
	}

	leftChild, ok := left.(SQLValues)
	if ok {
		if len(leftChild.Values) != 1 {
			return SQLBool(false), fmt.Errorf("left operand should contain 1 column")
		}
		left = leftChild.Values[0]
	}

	for _, right := range rightChild.Values {
		eq := &SQLEqualsExpr{left, right}
		m, err := Matches(eq, ctx)
		if err != nil {
			return SQLBool(false), err
		}
		if m {
			return SQLBool(true), nil
		}
	}

	return SQLBool(false), nil
}

//
// Evaluates to true when the left is less than the right.
//
type SQLLessThanExpr SQLBinaryNode

func (lt *SQLLessThanExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftEvald, err := lt.left.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	rightEvald, err := lt.right.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	if _, ok := rightEvald.(SQLValues); ok {
		c, err := rightEvald.CompareTo(leftEvald)
		if err == nil {
			return SQLBool(c > 0), nil
		}
		return SQLBool(false), err
	}

	c, err := leftEvald.CompareTo(rightEvald)
	if err == nil {
		return SQLBool(c < 0), nil
	}
	return SQLBool(false), err
}

//
// Evaluates to true when the left is less than or equal to the right.
//
type SQLLessThanOrEqualExpr SQLBinaryNode

func (lte *SQLLessThanOrEqualExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftEvald, err := lte.left.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	rightEvald, err := lte.right.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	if _, ok := rightEvald.(SQLValues); ok {
		c, err := rightEvald.CompareTo(leftEvald)
		if err == nil {
			return SQLBool(c >= 0), nil
		}
		return SQLBool(false), err
	}

	c, err := leftEvald.CompareTo(rightEvald)
	if err == nil {
		return SQLBool(c <= 0), nil
	}
	return SQLBool(false), err
}

//
// Evaluates to true if the left is 'like' the right.
//
type SQLLikeExpr SQLBinaryNode

func (l *SQLLikeExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	value, err := l.left.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	data, err := sqlValueToString(value)
	if err != nil {
		return SQLBool(false), err
	}

	value, err = l.right.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	pattern, ok := value.(SQLString)
	if !ok {
		return SQLBool(false), nil
	}

	// TODO: Golang's regexp package expects a regex pattern
	// for matching but MySQL's 'LIKE' operator doesn't exactly
	// work the same way.
	matches, err := regexp.Match(string(pattern), []byte(data))
	if err != nil {
		return SQLBool(false), err
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
// Evaluates to the inverse of its child.
//
type SQLNotExpr struct {
	child SQLExpr
}

func (not *SQLNotExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	m, err := Matches(not.child, ctx)
	if err != nil {
		return SQLBool(false), err
	}
	return SQLBool(!m), nil
}

//
// Evaluates to true if the left does not equal the right.
//
type SQLNotEqualsExpr SQLBinaryNode

func (neq *SQLNotEqualsExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftEvald, err := neq.left.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	rightEvald, err := neq.right.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	if _, ok := rightEvald.(SQLValues); ok {
		c, err := rightEvald.CompareTo(leftEvald)
		if err == nil {
			return SQLBool(c != 0), nil
		}
		return SQLBool(false), err
	}

	c, err := leftEvald.CompareTo(rightEvald)
	if err == nil {
		return SQLBool(c != 0), nil
	}

	return SQLBool(false), err
}

//
// Evaluates to true if the left is not in any of the values on the right.
//
type SQLNotInExpr SQLBinaryNode

func (nin *SQLNotInExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	in := &SQLInExpr{nin.left, nin.right}
	m, err := Matches(in, ctx)
	if err != nil {
		return SQLBool(false), err
	}
	return SQLBool(!m), nil
}

//
// Evaluates to true if its value evaluates to null.
//
type SQLNullCmpExpr struct {
	val SQLExpr
}

func (nm *SQLNullCmpExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	eval, err := nm.val.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), nil
	}
	_, ok := eval.(SQLNullValue)
	return SQLBool(ok), nil
}

//
// Evaluates to true if any of its children evaluate to true.
//
type SQLOrExpr struct {
	children []SQLExpr
}

func (or *SQLOrExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	for _, c := range or.children {
		m, err := Matches(c, ctx)
		if err != nil {
			return SQLBool(false), err
		}
		if m {
			return SQLBool(true), nil
		}
	}
	return SQLBool(false), nil
}

//
// Evaluates to true if ...???
//
type SQLSubqueryCmpExpr struct {
	In    bool
	left  SQLExpr
	value *SQLSubqueryExpr
}

func (sc *SQLSubqueryCmpExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	left, err := sc.left.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	ctx.ExecCtx.Depth += 1

	operator, err := PlanQuery(ctx.ExecCtx, sc.value.stmt)
	if err != nil {
		return SQLBool(false), err
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

	right := SQLValues{}

	if err := operator.Open(ctx.ExecCtx); err != nil {
		return SQLBool(false), err
	}

	row := &Row{}

	matched := false

	for operator.Next(row) {

		values := row.GetValues(operator.OpFields())

		for _, value := range values {
			field, err := NewSQLValue(value, "")
			if err != nil {
				return SQLBool(false), err
			}
			right.Values = append(right.Values, field)
		}

		eq := &SQLEqualsExpr{left, right}

		matches, err := Matches(eq, ctx)
		if err != nil {
			return SQLBool(false), err
		}

		if matches {
			matched = true
			if sc.In {
				return SQLBool(true), err
			}
		}

		row, right = &Row{}, SQLValues{}

	}

	if sc.In {
		matched = true
	}

	return SQLBool(!matched), err
}
