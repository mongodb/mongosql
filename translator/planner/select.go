package planner

import (
	"gopkg.in/mgo.v2/bson"
	"strings"
)

type Select struct {
	// Columns is used to determine which column information we retrieve
	// from the data returned by a child operator
	Columns []Column

	// err holds any error that may have occurred during processing
	err error

	// source holds the source for this select statement
	source Operator

	// fields holds the set of column display names that is returned
	// to the client in the result set.
	fields []string

	// ctx is the current execution context
	ctx *ExecutionCtx
}

// SelectColumn holds all columns referenced in select expressions.
type SelectColumn struct {
	Columns []*Column
	Level   int
}

func (s *Select) Open(ctx *ExecutionCtx) error {

	if err := s.source.Open(ctx); err != nil {
		return err
	}

	s.ctx = ctx

	// no columns imply a star expression - in which case we grab all the fields
	if len(s.Columns) == 0 {
		for _, column := range s.source.OpFields() {
			s.fields = append(s.fields, column.View)
			s.Columns = append(s.Columns, *column)
		}
	}

	return nil
}

func (s *Select) Next(r *Row) bool {

	row := &Row{}

	var hasNext bool

	if len(s.Columns) == 0 {
		hasNext = s.source.Next(r)
	} else {
		hasNext = s.source.Next(row)
	}

	if !hasNext {

		if err := s.source.Err(); err != nil {
			s.err = err
		} else if err := s.source.Close(); err != nil {
			s.err = err
		}

	}

	if len(s.Columns) != 0 {
		data := map[string][]bson.DocElem{}
		for _, column := range s.Columns {
			t, v := s.getValue(&column, row)
			data[t] = append(data[t], v)
		}

		for k, v := range data {
			r.Data = append(r.Data, TableRow{k, v, nil})
		}
	}

	return hasNext
}

func (s *Select) getValue(column *Column, row *Row) (string, bson.DocElem) {

	for _, v := range row.Data {

		if column.Table != "" && column.Table != v.Table {
			continue
		}

		for _, c := range v.Values {
			if strings.ToLower(column.Name) == strings.ToLower(c.Name) { // TODO: optimize
				return column.Table, bson.DocElem{column.Name, c.Value}
			}
		}
	}

	return "", bson.DocElem{column.Name, nil}
}

func (s *Select) Fields() (f []string) {
	for _, column := range s.Columns {
		f = append(f, column.View)
	}
	return f
}

func (s *Select) Close() error {
	return s.source.Close()
}

func (s *Select) OpFields() (columns []*Column) {

	for _, column := range s.Columns {
		column := Column(column)
		columns = append(columns, &column)
	}

	return columns
}

func (s *Select) Err() error {
	return s.err
}
