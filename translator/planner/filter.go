package planner

import (
	"sync"
)

type Filter struct {
	filter   interface{}
	children []Operator
	err      error
	sync.Mutex
}

func (f *Filter) Open(ctx *ExecutionCtx) error {
	// open children as well
	for _, opr := range f.children {
		if err := opr.Open(ctx); err != nil {
			return err
		}
	}

	return nil
}

// Not Yet Functional
func (f *Filter) check(row *Row) bool {
	return true
}

func (f *Filter) Next(row *Row) bool {
	unfiltered := &Row{}

	// filter data from all the children
	for _, child := range f.children {
		for child.Next(unfiltered) {
			if f.check(unfiltered) {
				row.Data = unfiltered.Data
				return true
			}
		}
	}
	return false
}

func (f *Filter) Close() error {
	for _, c := range f.children {
		if err := c.Close(); err != nil {
			f.Lock()
			if f.err != nil {
				f.err = err
			}
			f.Unlock()
		}
	}
	return nil
}

func (f *Filter) Err() error {
	return nil
}
