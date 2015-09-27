package evaluator

import (
	"gopkg.in/mgo.v2/bson"
	"regexp"
)

//
// LessThan
//
type LessThan BinaryNode
type LessThanOrEqual BinaryNode

func (gt *LessThanOrEqual) Transform() (*bson.D, error) {
	return transformComparison(gt.left, gt.right, "$lte", "$gt")
}

func (gt *LessThan) Transform() (*bson.D, error) {
	return transformComparison(gt.left, gt.right, "$lt", "$gte")
}

func (lt *LessThan) Matches(ctx *EvalCtx) bool {
	leftEvald, err := lt.left.Evaluate(ctx)
	if err != nil {
		return false
	}
	rightEvald, err := lt.right.Evaluate(ctx)
	if err != nil {
		return false
	}
	if c, err := leftEvald.CompareTo(ctx, rightEvald); err == nil {
		return c < 0
	}
	return false
}

func (lte *LessThanOrEqual) Matches(ctx *EvalCtx) bool {
	leftEvald, err := lte.left.Evaluate(ctx)
	if err != nil {
		return false
	}
	rightEvald, err := lte.right.Evaluate(ctx)
	if err != nil {
		return false
	}
	if c, err := leftEvald.CompareTo(ctx, rightEvald); err == nil {
		return c <= 0
	}
	return false
}

//
// GreaterThan
//

type GreaterThan BinaryNode
type GreaterThanOrEqual BinaryNode

func (gt *GreaterThan) Matches(ctx *EvalCtx) bool {
	leftEvald, err := gt.left.Evaluate(ctx)
	if err != nil {
		return false
	}
	rightEvald, err := gt.right.Evaluate(ctx)
	if err != nil {
		return false
	}
	if c, err := leftEvald.CompareTo(ctx, rightEvald); err == nil {
		return c > 0
	}
	return false
}

func (gt *GreaterThan) Transform() (*bson.D, error) {
	return transformComparison(gt.left, gt.right, "$gt", "$lte")
}

func (gt *GreaterThanOrEqual) Transform() (*bson.D, error) {
	return transformComparison(gt.left, gt.right, "$gte", "$lt")
}

func (gte *GreaterThanOrEqual) Matches(ctx *EvalCtx) bool {
	leftEvald, err := gte.left.Evaluate(ctx)
	if err != nil {
		return false
	}
	rightEvald, err := gte.right.Evaluate(ctx)
	if err != nil {
		return false
	}
	if c, err := leftEvald.CompareTo(ctx, rightEvald); err == nil {
		return c >= 0
	}
	return false
}

//
// Like
//

type Like BinaryNode

func (l *Like) Transform() (*bson.D, error) {
	return transformComparison(l.left, l.right, "$regex", "no inverse like support")
}

func (l *Like) Matches(ctx *EvalCtx) bool {
	e, err := l.left.Evaluate(ctx)
	if err != nil {
		return false
	}

	left := e.MongoValue().(string)

	e, err = l.right.Evaluate(ctx)
	if err != nil {
		return false
	}

	right := e.MongoValue().(string)

	matches, err := regexp.Match(left, []byte(right))
	if err != nil {
		return false
	}
	return matches
}

//
// NotEquals
//

type NotEquals BinaryNode

func (neq *NotEquals) Matches(ctx *EvalCtx) bool {
	leftEvald, err := neq.left.Evaluate(ctx)
	if err != nil {
		return false
	}
	rightEvald, err := neq.right.Evaluate(ctx)
	if err != nil {
		return false
	}
	if c, err := leftEvald.CompareTo(ctx, rightEvald); err == nil {
		return c != 0
	}
	return false
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

func (eq *Equals) Matches(ctx *EvalCtx) bool {
	leftEvald, err := eq.left.Evaluate(ctx)
	if err != nil {
		return false
	}
	rightEvald, err := eq.right.Evaluate(ctx)
	if err != nil {
		return false
	}
	if c, err := leftEvald.CompareTo(ctx, rightEvald); err == nil {
		return c == 0
	}
	return false
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
