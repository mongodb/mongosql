package evaluator

import (
	"gopkg.in/mgo.v2/bson"
)

// BSONSource is the simple interface for SQLProxy to simulate
// data coming from a MongoDB installation.
type BSONSource struct {
	index     int
	tableName string
	data      []bson.D
	err       error
}

func NewBSONSource(ctx *ExecutionCtx, tableName string, data []bson.D) (*BSONSource, error) {

	testSource := &BSONSource{
		data:      data,
		tableName: tableName,
	}

	return testSource, nil
}

func (ts *BSONSource) Open(ctx *ExecutionCtx) error {
	return nil
}

func (ts *BSONSource) Next(row *Row) bool {

	if ts.index == len(ts.data) || ts.data == nil {
		return false
	}

	var values Values

	for _, docElem := range ts.data[ts.index] {

		var value SQLValue

		value, ts.err = NewSQLValue(docElem.Value, "")
		if ts.err != nil {
			return false
		}

		values = append(values, Value{docElem.Name, docElem.Name, value})
	}

	row.Data = TableRows{{ts.tableName, values}}
	ts.index += 1

	return true
}

func (ts *BSONSource) OpFields() (columns []*Column) {
	return nil
}

func (ts *BSONSource) Close() error {
	return nil
}

func (ts *BSONSource) Err() error {
	return ts.err
}
