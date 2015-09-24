package evaluator

import (
	"gopkg.in/mgo.v2/bson"
)

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
