package planner

import (
	"gopkg.in/mgo.v2/bson"
)

type value struct {
	Table string
	Data  bson.D
}

type Select struct {
	Namespaces []*Namespace
	err        error
	children   []Operator
	values     []*value
	fields     []string
	closed     map[Operator]bool
	isStar     bool
}

func (s *Select) Open(ctx *ExecutionCtx) error {
	for _, ns := range s.Namespaces {
		s.fields = append(s.fields, ns.View)
	}

	for _, opr := range s.children {
		if err := opr.Open(ctx); err != nil {
			return err
		}
	}

	s.closed = map[Operator]bool{}
	return nil
}

func (s *Select) Next(row *Row) bool {
	s.values = make([]*value, 0)

	// an equal number of input data is expected
	// from both children
	for _, opr := range s.children {

		r := &Row{}

		if !opr.Next(r) {

			if err := opr.Err(); err != nil {
				s.err = err
			} else if err := opr.Close(); err != nil {
				s.err = err
			}

			s.closed[opr] = true

			var hasOpen bool

			for _, opr := range s.children {
				if !s.closed[opr] {
					hasOpen = true
				}
			}

			if !hasOpen {
				return false
			}
		}

		for _, data := range r.Data {
			v := &value{
				Table: data.Table,
				Data:  data.Values,
			}

			s.values = append(s.values, v)

		}

	}

	// merge input from children
	err := s.mergeInput(row)
	if err != nil {
		s.err = err
		return false
	}

	return true
}

func (s *Select) mergeInput(r *Row) error {

	// handles StarExpr
	if s.fields == nil {
		if err := s.setNamespaces(); err != nil {
			return err
		}
	}

	data := map[string][]bson.DocElem{}
	for _, ns := range s.Namespaces {
		t, v := s.getValue(ns)
		data[t] = append(data[t], v)
	}

	for k, v := range data {
		r.Data = append(r.Data, TableRow{k, v})
	}

	return nil
}

func (s *Select) Fields() []string {
	return s.fields
}

func (s *Select) getValue(ns *Namespace) (string, bson.DocElem) {

	for _, v := range s.values {

		if ns.Table != "" && ns.Table != v.Table {
			continue
		}

		for _, c := range v.Data {
			if ns.Column == c.Name {
				return ns.Table, bson.DocElem{ns.Column, c.Value}
			}
		}
	}

	return "", bson.DocElem{ns.Column, nil}
}

func (s *Select) setNamespaces() error {
	// first record retrieved determines the set of columns
	// we return for this query

	// TODO: unqualified join selects should have qualified names

	for _, v := range s.values {
		for _, f := range v.Data {
			ns := &Namespace{v.Table, f.Name, f.Name}
			s.Namespaces = append(s.Namespaces, ns)
			s.fields = append(s.fields, ns.View)
		}
	}

	return nil
}

func (s *Select) Close() error {
	for _, c := range s.children {
		if err := c.Close(); err != nil {
			s.err = err
		}
	}
	return nil
}

func (s *Select) Err() error {
	return s.err
}
