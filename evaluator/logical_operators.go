package evaluator

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
)

//
// Not
//
type Not struct {
	child SQLExpr
}

func (not *Not) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	m, err := Matches(not.child, ctx)
	if err != nil {
		return SQLBool(false), err
	}
	return SQLBool(!m), nil
}

//
// Or
//
type Or struct {
	children []SQLExpr
}

func (or *Or) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	for _, c := range or.children {
		m, err := Matches(c, ctx)
		if err != nil {
			return SQLBool(false), err
		}
		if m {
			return SQLBool(true), nil
		}
	}
	return SQLBool(false), nil
}

// And is a matcher that matches if and only if
// all of its children match.
type And struct {
	children []SQLExpr
}

func (and *And) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	for _, c := range and.children {
		m, err := Matches(c, ctx)
		if err != nil {
			return SQLBool(false), err
		}
		if !m {
			return SQLBool(false), nil
		}
	}
	return SQLBool(true), nil
}

// NullMatcher matches if the embedded SQLValue evaluates to null.
type NullMatcher struct {
	val SQLValue
}

func (nm *NullMatcher) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	eval, err := nm.val.Evaluate(ctx)
	if err != nil {
		return SQLBool(false), nil
	}
	_, ok := eval.(SQLNullValue)
	return SQLBool(ok), nil
}

// NoopMatcher is a matcher that always returns true for any row.
type NoopMatcher struct{}

func (no *NoopMatcher) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return SQLBool(true), nil
}

// ExistsMatcher returns true if any result is returned from an exists
// expression referencing a subquery.
type ExistsMatcher struct {
	stmt sqlparser.SelectStatement
}

func (em *ExistsMatcher) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	ctx.ExecCtx.Depth += 1

	operator, err := PlanQuery(ctx.ExecCtx, em.stmt)
	if err != nil {
		return SQLBool(false), err
	}

	var matches bool

	defer func() {
		if err == nil {
			err = operator.Err()
		}

		// add context to error
		if err != nil {
			err = fmt.Errorf("ExistsMatcher (%v): %v", ctx.ExecCtx.Depth, err)
		}

		ctx.ExecCtx.Depth -= 1

	}()

	if err := operator.Open(ctx.ExecCtx); err != nil {
		return SQLBool(false), err
	}

	if operator.Next(&Row{}) {
		matches = true
	}

	return SQLBool(matches), operator.Close()
}
