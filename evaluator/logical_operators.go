package evaluator

import (
	"github.com/mongodb/mongo-tools/common/util"
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

//
// Or
//
type Or struct {
	children []Matcher
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

//
// NoopMatch
//

type NoopMatch struct{}

func (no *NoopMatch) Matches(ctx *EvalCtx) (bool, error) {
	return true, nil
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
