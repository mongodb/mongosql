package evaluator

import (
	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/schema"
)

const (
	BSONSourceDB = "bson_source_database"
)

// BSONSourceStage is the simple interface for SQLProxy to simulate
// data coming from a MongoDB installation.
type BSONSourceStage struct {
	selectID     int
	databaseName string
	tableName    string
	collation    *collation.Collation
	data         []bson.D
}

// NewBSONSourceStage constructs a BSONSourceStage with its required values.
func NewBSONSourceStage(selectID int, tableName string, collation *collation.Collation, data []bson.D) *BSONSourceStage {
	return &BSONSourceStage{
		selectID:     selectID,
		databaseName: BSONSourceDB,
		tableName:    tableName,
		collation:    collation,
		data:         data,
	}
}

type BSONSourceIter struct {
	selectID     int
	tableName    string
	databaseName string
	data         []bson.D
	index        int
	err          error
}

func (bs *BSONSourceStage) Open(ctx *ExecutionCtx) (Iter, error) {
	return &BSONSourceIter{selectID: bs.selectID, databaseName: bs.databaseName, data: bs.data, tableName: bs.tableName, index: 0}, nil
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

		values = append(values, NewValue(
			bs.selectID,
			bs.databaseName,
			bs.tableName,
			docElem.Name,
			value))
	}

	row.Data = values
	bs.index++

	return true
}

func (bs *BSONSourceStage) Columns() []*Column {

	var columns []*Column
	for _, v := range bs.data[0] {
		column := NewColumn(bs.selectID, bs.tableName, bs.tableName, bs.databaseName, v.Name, v.Name, "",
			schema.SQLNone, schema.MongoNone, false)
		columns = append(columns, column)
	}
	return columns
}

func (bs *BSONSourceStage) Collation() *collation.Collation {
	return bs.collation
}

func (bs *BSONSourceIter) Close() error {
	return nil
}

func (bs *BSONSourceIter) Err() error {
	return bs.err
}
