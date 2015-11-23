package evaluator

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"strconv"
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

// And is a matcher that matches if and only if
// all of its children match.
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

// NullMatcher matches if the embedded SQLValue evaluates to null.
type NullMatcher struct {
	val SQLValue
}

func (nm *NullMatcher) Matches(ctx *EvalCtx) (bool, error) {
	eval, err := nm.val.Evaluate(ctx)
	if err != nil {
		return false, nil
	}
	reg := eval.MongoValue()
	return reg == nil, nil
}

// NoopMatcher is a matcher that always returns true for any row.
type NoopMatcher struct{}

func (no *NoopMatcher) Matches(ctx *EvalCtx) (bool, error) {
	return true, nil
}

// ExistsMatcher returns true if any result is returned from an exists
// expression referencing a subquery.
type ExistsMatcher struct {
	stmt sqlparser.SelectStatement
}

func (em *ExistsMatcher) Matches(ctx *EvalCtx) (bool, error) {
	ctx.ExecCtx.Depth += 1

	operator, err := PlanQuery(ctx.ExecCtx, em.stmt)
	if err != nil {
		return false, err
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
		return false, err
	}

	if operator.Next(&Row{}) {
		matches = true
	}

	return matches, operator.Close()
}

// BoolMatcher is a matcher that returns true if the embedded SQLValue evaluates
// to something that is "truthy" i.e. a non-zero number, or a string that parses
// as a non-zero number.
type BoolMatcher struct {
	SQLValue
}

func (bm *BoolMatcher) Matches(ctx *EvalCtx) (bool, error) {
	val, err := bm.Evaluate(ctx)
	if err != nil {
		return false, err
	}
	return IsTruthy(val), nil
}

// IsTruthy checks if a given SQLValue is "truthy" or should by coercing it to a boolean value.
// - booleans: the result is simply that same return value
// - numeric values: the result is true if and only if the value is non-zero.
// - strings, the result is true if and only if that string can be parsed as a number,
//   and that number is non-zero.
func IsTruthy(sv SQLValue) bool {
	if asBool, ok := sv.(SQLBool); ok {
		return bool(asBool)
	}
	if asNum, ok := sv.(SQLNumeric); ok {
		return asNum.Float64() != float64(0)
	}
	if asStr, ok := sv.(SQLString); ok {
		// check if the string should be considered "truthy" by trying to convert it to a number and comparing to 0.
		// more info: http://stackoverflow.com/questions/12221211/how-does-string-truthiness-work-in-mysql
		if parsedFloat, err := strconv.ParseFloat(string(asStr), 64); err == nil {
			return parsedFloat != float64(0)
		}
		return false
	}
	// TODO - handle other types with possible values that are "truthy" : dates, etc?
	return false
}
