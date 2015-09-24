package evaluator

import (
	"gopkg.in/mgo.v2/bson"
)

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
