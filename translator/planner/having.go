package planner

import (
	"github.com/erh/mixer/sqlparser"
	/*
		"fmt"
		"gopkg.in/mgo.v2/bson"
	*/)

type Having struct {
	// source is the operator that provides the data to filter
	source Operator
	// err holds any error encountered during processing
	err error
	// matcher is used to evaluate each row in determining whether
	// it passes the 'HAVING' filter
	matcher Matcher
	// expr is the boolean expression to match
	expr sqlparser.BoolExpr
}

func (hv *Having) Open(ctx *ExecutionCtx) (err error) {
	hv.matcher, err = BuildMatcher(hv.expr)
	if err != nil {
		return err
	}

	return hv.source.Open(ctx)
}

func (hv *Having) Next(row *Row) bool {
	r := &Row{}

	var hasNext bool

	for {
		hasNext = hv.source.Next(r)

		if !hasNext {
			if err := hv.source.Err(); err != nil {
				hv.err = err
			} else if err := hv.source.Close(); err != nil {
				hv.err = err
			}
			return false
		}

		evalCtx := &EvalCtx{[]Row{*r}}

		if hv.matcher.Matches(evalCtx) {
			row.Data = r.Data
			break
		}
	}

	return hasNext
}

func (hv *Having) Close() error {
	return hv.source.Close()
}

func (hv *Having) Err() error {
	return hv.err
}

func (hv *Having) OpFields() (columns []*Column) {
	return hv.source.OpFields()
}
