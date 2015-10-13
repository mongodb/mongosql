package evaluator

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"regexp"
)

// TODO: support SQLValues in all nodes

//
// LessThan
//
type LessThan BinaryNode
type LessThanOrEqual BinaryNode

func (lt *LessThanOrEqual) Transform() (*bson.D, error) {
	return transformComparison(lt.left, lt.right, "$lte", "$gt")
}

func (lt *LessThan) Transform() (*bson.D, error) {
	return transformComparison(lt.left, lt.right, "$lt", "$gte")
}

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

func (gt *GreaterThan) Transform() (*bson.D, error) {
	return transformComparison(gt.left, gt.right, "$gt", "$lte")
}

func (gt *GreaterThanOrEqual) Transform() (*bson.D, error) {
	return transformComparison(gt.left, gt.right, "$gte", "$lt")
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

func (l *Like) Transform() (*bson.D, error) {
	return transformComparison(l.left, l.right, "$regex", "no inverse like support")
}

func (l *Like) Matches(ctx *EvalCtx) (bool, error) {
	e, err := l.left.Evaluate(ctx)
	if err != nil {
		return false, err
	}

	left := e.MongoValue().(string)

	e, err = l.right.Evaluate(ctx)
	if err != nil {
		return false, err
	}

	right := e.MongoValue().(string)

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

func (neq *NotEquals) Transform() (*bson.D, error) {
	tField, tLiteral, _, err := makeMQLQueryPair(neq.left, neq.right)
	if err != nil {
		return nil, err
	}
	return &bson.D{
		{tField.fieldName, bson.D{{"$neq", tLiteral.MongoValue()}}},
	}, nil
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

func (eq *Equals) Transform() (*bson.D, error) {
	tField, tLiteral, _, err := makeMQLQueryPair(eq.left, eq.right)
	if err != nil {
		return nil, err
	}
	return &bson.D{
		{tField.fieldName, bson.D{{"$eq", tLiteral.MongoValue()}}},
	}, nil
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
		if len(leftChild.values) != 1 {
			return false, fmt.Errorf("left operand should contain 1 column")
		}
		left = leftChild.values[0]
	}

	for _, right := range rightChild.values {
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

func (in *In) Transform() (*bson.D, error) {
	return &bson.D{}, nil
}
