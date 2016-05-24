package evaluator

import (
	"github.com/10gen/sqlproxy/schema"
	"gopkg.in/mgo.v2/bson"
)

// BSONSource is the simple interface for SQLProxy to simulate
// data coming from a MongoDB installation.
type BSONSourceStage struct {
	tableName string
	data      []bson.D
}

type BSONSourceIter struct {
	tableName string
	data      []bson.D
	index     int
	err       error
}

/*
func NewBSONSource(ctx *ExecutionCtx, tableName string, data []bson.D) (*BSONSource, error) {

	testSource := &BSONSource{
		data:      data,
		tableName: tableName,
	}

	return testSource, nil
}
*/

func (bs *BSONSourceStage) Open(ctx *ExecutionCtx) (Iter, error) {
	return &BSONSourceIter{data: bs.data, tableName: bs.tableName, index: 0}, nil
}

func (bs *BSONSourceIter) Next(row *Row) bool {

	if bs.index == len(bs.data) || bs.data == nil {
		return false
	}

	var values Values

	for _, docElem := range bs.data[bs.index] {

		var value SQLValue

		value, bs.err = NewSQLValue(docElem.Value, schema.SQLNone, schema.MongoNone)
		if bs.err != nil {
			return false
		}

		values = append(values, Value{
			Table: bs.tableName,
			Name:  docElem.Name,
			Data:  value,
		})
	}

	row.Data = values
	bs.index += 1

	return true
}

func (bs *BSONSourceStage) OpFields() (columns []*Column) {
	return nil
}

func (bs *BSONSourceIter) Close() error {
	return nil
}

func (bs *BSONSourceIter) Err() error {
	return bs.err
}
