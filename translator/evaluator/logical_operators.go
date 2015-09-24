package evaluator

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
)

//
// Not
//
type Not struct {
	child Matcher
}

func (not *Not) Matches(ctx *EvalCtx) bool {
	return !not.child.Matches(ctx)
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

func (or *Or) Matches(ctx *EvalCtx) bool {
	for _, c := range or.children {
		if c.Matches(ctx) {
			return true
		}
	}
	return false
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

func (and *And) Matches(ctx *EvalCtx) bool {
	for _, c := range and.children {
		if !c.Matches(ctx) {
			return false
		}
	}
	return true
}
