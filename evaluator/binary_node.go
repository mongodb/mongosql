package evaluator

import (
	"fmt"
	"regexp"
	"strconv"
)

//
// LessThan
//
type LessThan BinaryNode
type LessThanOrEqual BinaryNode

func (lt *LessThan) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftEvald, err := lt.left.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	rightEvald, err := lt.right.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	if _, ok := rightEvald.(SQLValues); ok {
		c, err := rightEvald.CompareTo(ctx, leftEvald)
		if err == nil {
			return SQLBool(c > 0), nil
		}
		return SQLBool(false), err
	}

	c, err := leftEvald.CompareTo(ctx, rightEvald)
	if err == nil {
		return SQLBool(c < 0), nil
	}
	return SQLBool(false), err
}

func (lte *LessThanOrEqual) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftEvald, err := lte.left.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	rightEvald, err := lte.right.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	if _, ok := rightEvald.(SQLValues); ok {
		c, err := rightEvald.CompareTo(ctx, leftEvald)
		if err == nil {
			return SQLBool(c >= 0), nil
		}
		return SQLBool(false), err
	}

	c, err := leftEvald.CompareTo(ctx, rightEvald)
	if err == nil {
		return SQLBool(c <= 0), nil
	}
	return SQLBool(false), err
}

//
// GreaterThan
//

type GreaterThan BinaryNode
type GreaterThanOrEqual BinaryNode

func (gt *GreaterThan) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftEvald, err := gt.left.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	rightEvald, err := gt.right.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	if _, ok := rightEvald.(SQLValues); ok {
		c, err := rightEvald.CompareTo(ctx, leftEvald)
		if err == nil {
			return SQLBool(c < 0), nil
		}
		return SQLBool(false), err
	}

	c, err := leftEvald.CompareTo(ctx, rightEvald)
	if err == nil {
		return SQLBool(c > 0), nil
	}
	return SQLBool(false), err
}

func (gte *GreaterThanOrEqual) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftEvald, err := gte.left.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	rightEvald, err := gte.right.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	if _, ok := rightEvald.(SQLValues); ok {
		c, err := rightEvald.CompareTo(ctx, leftEvald)
		if err == nil {
			return SQLBool(c <= 0), nil
		}
		return SQLBool(false), err
	}

	c, err := leftEvald.CompareTo(ctx, rightEvald)
	if err == nil {
		return SQLBool(c >= 0), nil
	}

	return SQLBool(false), err

}

//
// Like
//

type Like BinaryNode

func (l *Like) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	value, err := l.left.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	data, err := SQLValueToString(value)
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

func SQLValueToString(sqlValue SQLValue) (string, error) {
	switch v := sqlValue.(type) {
	case SQLString:
		return string(v), nil
	case SQLInt:
		return string(v), nil
	case SQLUint32:
		return string(v), nil
	case SQLFloat:
		return strconv.FormatFloat(float64(v), 'f', -1, 64), nil
	}

	// TODO: just return empty string with no error?
	return "", fmt.Errorf("unable to convert %v to string", sqlValue)
}

//
// NotEquals
//

type NotEquals BinaryNode

func (neq *NotEquals) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftEvald, err := neq.left.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	rightEvald, err := neq.right.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	if _, ok := rightEvald.(SQLValues); ok {
		c, err := rightEvald.CompareTo(ctx, leftEvald)
		if err == nil {
			return SQLBool(c != 0), nil
		}
		return SQLBool(false), err
	}

	c, err := leftEvald.CompareTo(ctx, rightEvald)
	if err == nil {
		return SQLBool(c != 0), nil
	}

	return SQLBool(false), err

}

//
// Equals
//

type Equals BinaryNode

func (eq *Equals) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftEvald, err := eq.left.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	rightEvald, err := eq.right.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	if _, ok := rightEvald.(SQLValues); ok {
		c, err := rightEvald.CompareTo(ctx, leftEvald)
		if err == nil {
			return SQLBool(c == 0), nil
		}
		return SQLBool(false), err
	}

	c, err := leftEvald.CompareTo(ctx, rightEvald)
	if err == nil {
		return SQLBool(c == 0), nil
	}

	return SQLBool(false), err

}

//
// In
//

type In BinaryNode

func (in *In) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	left, err := in.left.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	right, err := in.right.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), err
	}

	// right child must be of type SQLValues
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
		eq := &Equals{left, right}
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
// NotIn
//

type NotIn BinaryNode

func (nin *NotIn) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	in := &In{nin.left, nin.right}
	m, err := Matches(in, ctx)
	if err != nil {
		return SQLBool(false), err
	}
	return SQLBool(!m), nil
}

//
// SubqueryCmp
//
// INT-911 support ANY and SOME in subquery
type SubqueryCmp struct {
	In    bool
	left  SQLValue
	value *SubqueryValue
}

func (sc *SubqueryCmp) Evaluate(ctx *EvalCtx) (SQLValue, error) {

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
			field, err := NewSQLField(value, "")
			if err != nil {
				return SQLBool(false), err
			}
			right.Values = append(right.Values, field)
		}

		eq := &Equals{left, right}

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
