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
	leftEvald := lt.left.Evaluate(ctx)
	rightEvald := lt.right.Evaluate(ctx)
	if c, err := leftEvald.CompareTo(ctx, rightEvald); err == nil {
		return c < 0
	}
	return false
}

func (lte *LessThanOrEqual) Matches(ctx *EvalCtx) bool {
	leftEvald := lte.left.Evaluate(ctx)
	rightEvald := lte.right.Evaluate(ctx)
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
	leftEvald := gt.left.Evaluate(ctx)
	rightEvald := gt.right.Evaluate(ctx)
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
	leftEvald := gte.left.Evaluate(ctx)
	rightEvald := gte.right.Evaluate(ctx)
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
	reg := l.left.Evaluate(ctx).MongoValue().(string)
	res, err := regexp.Match(reg, []byte(l.right.Evaluate(ctx).MongoValue().(string)))
	if err != nil {
		panic(err)
	}
	return res
}

//
// NotEquals
//

type NotEquals BinaryNode

func (neq *NotEquals) Matches(ctx *EvalCtx) bool {
	leftEvald := neq.left.Evaluate(ctx)
	rightEvald := neq.right.Evaluate(ctx)
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
	leftEvald := eq.left.Evaluate(ctx)
	rightEvald := eq.right.Evaluate(ctx)
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
