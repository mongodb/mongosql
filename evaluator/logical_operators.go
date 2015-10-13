package evaluator

import (
	"fmt"
	"github.com/mongodb/mongo-tools/common/util"
	"gopkg.in/mgo.v2/bson"
)

//
// Not
//
type Not struct {
	child Matcher
}

func (not *Not) Matches(ctx *EvalCtx) (bool, error) {
	m, err := not.child.Matches(ctx)
	if err != nil {
		return false, err
	}
	return !m, nil
}

func (not *Not) Transform() (*bson.D, error) {
	return nil, fmt.Errorf("transformation of 'not' expressions not supported")
}

//
// Or
//
type Or struct {
	children []Matcher
}

func (or *Or) Transform() (*bson.D, error) {
	transformedChildren := make([]*bson.D, 0, len(or.children))
	for _, child := range or.children {
		transformedChild, err := child.Transform()
		if err != nil {
			return nil, err
		}
		transformedChildren = append(transformedChildren, transformedChild)
	}
	return &bson.D{{"$or", transformedChildren}}, nil
}

func (or *Or) Matches(ctx *EvalCtx) (bool, error) {
	for _, c := range or.children {
		m, err := c.Matches(ctx)
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
// And
//

type And struct {
	children []Matcher
}

func (and *And) Transform() (*bson.D, error) {
	transformedChildren := make([]*bson.D, 0, len(and.children))
	for _, child := range and.children {
		transformedChild, err := child.Transform()
		if err != nil {
			return nil, err
		}
		transformedChildren = append(transformedChildren, transformedChild)
	}
	return &bson.D{{"$and", transformedChildren}}, nil
}

func (and *And) Matches(ctx *EvalCtx) (bool, error) {
	for _, c := range and.children {
		m, err := c.Matches(ctx)
		if err != nil {
			return false, err
		}
		if !m {
			return false, nil
		}
	}
	return true, nil
}

//
// NullMatch
//

type NullMatch struct {
	wantsNull bool
	val       SQLValue
}

func (nm *NullMatch) Matches(ctx *EvalCtx) (bool, error) {
	eval, err := nm.val.Evaluate(ctx)
	if err != nil {
		return false, nil
	}
	reg := eval.MongoValue()
	if nm.wantsNull {
		return reg == nil, nil
	}
	return reg != nil, nil
}

func (nm *NullMatch) Transform() (*bson.D, error) {
	field, ok := nm.val.(SQLField)
	if !ok {
		return nil, ErrUntransformableCondition
	}
	if nm.wantsNull {
		return &bson.D{{field.fieldName, bson.D{{"$eq", nil}}}}, nil
	}
	return &bson.D{{field.fieldName, bson.D{{"$ne", nil}}}}, nil
}

//
// NoopMatch
//

type NoopMatch struct{}

func (no *NoopMatch) Matches(ctx *EvalCtx) (bool, error) {
	return true, nil
}

func (nm *NoopMatch) Transform() (*bson.D, error) {
	return nil, nil
}

//
// BoolMatch
//

type BoolMatch struct {
	SQLValue
}

func (bm *BoolMatch) Matches(ctx *EvalCtx) (bool, error) {
	val, err := bm.Evaluate(ctx)
	if err != nil {
		return false, err
	}
	return util.IsTruthy(val), nil
}

func (bm *BoolMatch) Transform() (*bson.D, error) {
	return nil, nil
}
