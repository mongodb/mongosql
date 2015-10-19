package evaluator

import (
	"fmt"
	"regexp"
)

//
// LessThan
//
type LessThan BinaryNode
type LessThanOrEqual BinaryNode

func (lt *LessThan) Matches(ctx *EvalCtx) (bool, error) {
	leftEvald, err := lt.left.Evaluate(ctx)
	if err != nil {
		return false, err
	}
	rightEvald, err := lt.right.Evaluate(ctx)
	if err != nil {
		return false, err
	}
	if _, ok := rightEvald.(SQLValues); ok {
		c, err := rightEvald.CompareTo(ctx, leftEvald)
		if err == nil {
			return c > 0, nil
		}
		return false, err
	}
	c, err := leftEvald.CompareTo(ctx, rightEvald)
	if err == nil {
		return c < 0, nil
	}
	return false, err
}

func (lte *LessThanOrEqual) Matches(ctx *EvalCtx) (bool, error) {
	leftEvald, err := lte.left.Evaluate(ctx)
	if err != nil {
		return false, err
	}
	rightEvald, err := lte.right.Evaluate(ctx)
	if err != nil {
		return false, err
	}
	if _, ok := rightEvald.(SQLValues); ok {
		c, err := rightEvald.CompareTo(ctx, leftEvald)
		if err == nil {
			return c >= 0, nil
		}
		return false, err
	}
	c, err := leftEvald.CompareTo(ctx, rightEvald)
	if err == nil {
		return c <= 0, nil
	}
	return false, err
}

//
// GreaterThan
//

type GreaterThan BinaryNode
type GreaterThanOrEqual BinaryNode

func (gt *GreaterThan) Matches(ctx *EvalCtx) (bool, error) {
	leftEvald, err := gt.left.Evaluate(ctx)
	if err != nil {
		return false, err
	}
	rightEvald, err := gt.right.Evaluate(ctx)
	if err != nil {
		return false, err
	}
	if _, ok := rightEvald.(SQLValues); ok {
		c, err := rightEvald.CompareTo(ctx, leftEvald)
		if err == nil {
			return c < 0, nil
		}
		return false, err
	}
	c, err := leftEvald.CompareTo(ctx, rightEvald)
	if err == nil {
		return c > 0, nil
	}
	return false, err
}

func (gte *GreaterThanOrEqual) Matches(ctx *EvalCtx) (bool, error) {
	leftEvald, err := gte.left.Evaluate(ctx)
	if err != nil {
		return false, err
	}
	rightEvald, err := gte.right.Evaluate(ctx)
	if err != nil {
		return false, err
	}
	if _, ok := rightEvald.(SQLValues); ok {
		c, err := rightEvald.CompareTo(ctx, leftEvald)
		if err == nil {
			return c <= 0, nil
		}
		return false, err
	}
	c, err := leftEvald.CompareTo(ctx, rightEvald)
	if err == nil {
		return c >= 0, nil
	}
	return false, err
}

//
// Like
//

type Like BinaryNode

func (l *Like) Matches(ctx *EvalCtx) (bool, error) {
	e, err := l.left.Evaluate(ctx)
	if err != nil {
		return false, err
	}

	value := e.MongoValue()
	left, ok := value.(string)
	if !ok {
		return false, nil
	}

	e, err = l.right.Evaluate(ctx)
	if err != nil {
		return false, err
	}

	value = e.MongoValue()
	right, ok := value.(string)
	if !ok {
		return false, nil
	}

	matches, err := regexp.Match(left, []byte(right))
	if err != nil {
		return false, err
	}
	return matches, nil
}

//
// NotEquals
//

type NotEquals BinaryNode

func (neq *NotEquals) Matches(ctx *EvalCtx) (bool, error) {
	leftEvald, err := neq.left.Evaluate(ctx)
	if err != nil {
		return false, err
	}
	rightEvald, err := neq.right.Evaluate(ctx)
	if err != nil {
		return false, err
	}
	if _, ok := rightEvald.(SQLValues); ok {
		c, err := rightEvald.CompareTo(ctx, leftEvald)
		if err == nil {
			return c != 0, nil
		}
		return false, err
	}
	c, err := leftEvald.CompareTo(ctx, rightEvald)
	if err == nil {
		return c != 0, nil
	}
	return false, err
}

//
// Equals
//

type Equals BinaryNode

func (eq *Equals) Matches(ctx *EvalCtx) (bool, error) {
	leftEvald, err := eq.left.Evaluate(ctx)
	if err != nil {
		return false, err
	}
	rightEvald, err := eq.right.Evaluate(ctx)
	if err != nil {
		return false, err
	}
	if _, ok := rightEvald.(SQLValues); ok {
		c, err := rightEvald.CompareTo(ctx, leftEvald)
		if err == nil {
			return c == 0, nil
		}
		return false, err
	}
	c, err := leftEvald.CompareTo(ctx, rightEvald)
	if err == nil {
		return c == 0, nil
	}
	return false, err
}

//
// In
//

type In BinaryNode

func (in *In) Matches(ctx *EvalCtx) (bool, error) {
	left, err := in.left.Evaluate(ctx)
	if err != nil {
		return false, err
	}

	right, err := in.right.Evaluate(ctx)
	if err != nil {
		return false, err
	}

	// right child must be of type SQLValues
	rightChild, ok := right.(SQLValues)
	if !ok {
		return false, fmt.Errorf("right In expression is %T", right)
	}

	leftChild, ok := left.(SQLValues)
	if ok {
		if len(leftChild.Values) != 1 {
			return false, fmt.Errorf("left operand should contain 1 column")
		}
		left = leftChild.Values[0]
	}

	for _, right := range rightChild.Values {
		eq := &Equals{left, right}
		m, err := eq.Matches(ctx)
		if err != nil {
			return false, err
		}
		if m {
			return true, nil
		}
	}

	return false, nil
}

//
// NotIn
//

type NotIn BinaryNode

func (nin *NotIn) Matches(ctx *EvalCtx) (bool, error) {
	in := &In{nin.left, nin.right}
	m, err := in.Matches(ctx)
	if err != nil {
		return false, err
	}
	return !m, nil
}
