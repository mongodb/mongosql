package planner

import (
	"gopkg.in/mgo.v2/bson"
	"github.com/erh/mongo-sql-temp/config"
)

type value struct {
	Table string
	Data  bson.D
	TableConfig *config.TableConfig
}

type Select struct {
	Namespaces []*Namespace // the actual fields to return (TODO WISDOM CHECK COMMENT PLEASE)
	err        error
	children   []Operator
	values     []*value
	fields     []string // field names to return to client (TODO WISDOM CHECK COMMENT PLEASE)
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
				TableConfig: data.TableConfig,
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

	// look for more namespaces to add
	// maybe all of them!
	if s.Namespaces == nil || s.isStar {

		if s.isStar && s.Namespaces == nil {
			// first inialize from select specs
			for _, v := range s.values {
				if v.TableConfig != nil {
					for _, col := range v.TableConfig.Columns {
						ns := &Namespace{v.Table, col.Name, col.Name}
						s.Namespaces = append(s.Namespaces, ns)
						s.fields = append(s.fields, ns.View)
					}
				}
			}
		}
		
		for _, v := range s.values {
			for _, f := range v.Data {
				ns := &Namespace{v.Table, f.Name, f.Name}
				// TODO this is n^1230
				found := false
				for _, old := range s.Namespaces {
					if old.Table == ns.Table &&
						old.Column == ns.Column {
						found = true;
						break;
					}
				}

				if !found {
					s.Namespaces = append(s.Namespaces, ns)
					s.fields = append(s.fields, ns.View)
				}
			}
		}

	}

	data := map[string][]bson.DocElem{}
	for _, ns := range s.Namespaces {
		t, v := s.getValue(ns)
		data[t] = append(data[t], v)
	}

	for k, v := range data {
		r.Data = append(r.Data, TableRow{k, v, nil})
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
