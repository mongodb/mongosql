package evaluator

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
)

type Not struct {
	child Matcher
}

func (not *Not) Matches(ctx *EvalCtx) bool {
	return !not.child.Matches(ctx)
}

func (not *Not) Transform() (*bson.D, error) {
	return nil, fmt.Errorf("transformation of 'not' expressions not supported")
}
