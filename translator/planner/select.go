package planner

import (
	"github.com/erh/mongo-sql-temp/config"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

// childData holds a single row information returned from
// a child of a select node.
type childData struct {
	// Table indicates what table (collection) this value was
	// returned from
	Table string
	// Data holds all the fields and values gotten from 'Table'
	// for this row
	Data bson.D
	// TableConfig is a pointer to the global configuration information
	// for a table
	TableConfig *config.TableConfig
}

type Select struct {
	// Columns is used to determine which column information we retrieve
	// from the data returned by a child operator
	Columns []*Column
	// err holds any error that may have occurred during processing
	err error
	// children holds all data sources that are connected to this select
	// operator and is used in conjunction with the Columns to determine
	// what data gets passed up on the query tree.
	children []Operator
	// childData holds data returned from the children operators
	childData []*childData
	// fields holds the set of column display names that is returned
	// to the client in the result set.
	fields []string
	// closed is used to track the state of all operator children of a
	// select node.
	closed map[Operator]bool
	// isStar indicates if one or more of the select expressions in this
	// operator was a star expression - '*'
	isStar bool
}

func (s *Select) Open(ctx *ExecutionCtx) error {
	for _, ns := range s.Columns {
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
	s.childData = make([]*childData, 0)

	// TODO: currently, continue building row data from any child
	// that returns data - even if others are closed. Will need to
	// look into this later.
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
			v := &childData{
				Table:       data.Table,
				Data:        data.Values,
				TableConfig: data.TableConfig,
			}

			s.childData = append(s.childData, v)

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
	if s.Columns == nil || s.isStar {

		if s.isStar && s.Columns == nil {
			// first initialize from column configuration file
			for _, v := range s.childData {
				if v.TableConfig != nil {
					for _, col := range v.TableConfig.Columns {
						column := &Column{v.Table, col.Name, col.Name}
						s.Columns = append(s.Columns, column)
						s.fields = append(s.fields, column.View)
					}
				}
			}
		}

		for _, v := range s.childData {
			for _, f := range v.Data {
				column := &Column{v.Table, f.Name, f.Name}
				// TODO this is n^1230
				found := false
				for _, old := range s.Columns {
					if old.Table == column.Table && old.Name == column.Name {
						found = true
						break
					}
				}

				if !found {
					s.Columns = append(s.Columns, column)
					s.fields = append(s.fields, column.View)
				}
			}
		}

	}

	data := map[string][]bson.DocElem{}
	for _, column := range s.Columns {
		t, v := s.getValue(column)
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

func (s *Select) getValue(column *Column) (string, bson.DocElem) {

	for _, v := range s.childData {
		if column.Table != "" && column.Table != v.Table {
			continue
		}

		for _, c := range v.Data {
			if strings.ToLower(column.Name) == strings.ToLower(c.Name) { // TODO: optimize
				return column.Table, bson.DocElem{column.Name, c.Value}
			}
		}
	}

	return "", bson.DocElem{column.Name, nil}
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
