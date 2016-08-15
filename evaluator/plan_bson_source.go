package evaluator

import (
	"github.com/10gen/sqlproxy/schema"
	"gopkg.in/mgo.v2/bson"
)

// BSONSource is the simple interface for SQLProxy to simulate
// data coming from a MongoDB installation.
type BSONSourceStage struct {
	selectID  int
	tableName string
	data      []bson.D
}

// NewBSONSourceStage constructs a BSONSourceStage with its required values.
func NewBSONSourceStage(selectID int, tableName string, data []bson.D) *BSONSourceStage {
	return &BSONSourceStage{
		selectID:  selectID,
		tableName: tableName,
		data:      data,
	}
}

type BSONSourceIter struct {
	selectID  int
	tableName string
	data      []bson.D
	index     int
	err       error
}

func (bs *BSONSourceStage) Open(ctx *ExecutionCtx) (Iter, error) {
	return &BSONSourceIter{selectID: bs.selectID, data: bs.data, tableName: bs.tableName, index: 0}, nil
}

func (bs *BSONSourceIter) Next(row *Row) bool {

	if bs.index == len(bs.data) || bs.data == nil {
		return false
	}

	var values Values

	for _, docElem := range bs.data[bs.index] {

		var value SQLValue

		value, bs.err = NewSQLValueFromSQLColumnExpr(docElem.Value, schema.SQLNone, schema.MongoNone)
		if bs.err != nil {
			return false
		}

		values = append(values, Value{
			SelectID: bs.selectID,
			Table:    bs.tableName,
			Name:     docElem.Name,
			Data:     value,
		})
	}

	row.Data = values
	bs.index += 1

	return true
}

func (bs *BSONSourceStage) Columns() []*Column {
	var columns []*Column
	for _, v := range bs.data[0] {
		columns = append(columns, &Column{
			SelectID:  bs.selectID,
			Table:     bs.tableName,
			Name:      v.Name,
			SQLType:   schema.SQLNone,
			MongoType: schema.MongoNone,
		})
	}
	return columns
}

func (bs *BSONSourceIter) Close() error {
	return nil
}

func (bs *BSONSourceIter) Err() error {
	return bs.err
}
