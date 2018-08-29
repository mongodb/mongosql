package evaluator

import (
	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/collation"
	"github.com/10gen/sqlproxy/internal/memory"
	"github.com/10gen/sqlproxy/schema"
)

const (
	// BSONSourceDB is the database name we use for data sourced from BSON documents.
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
func NewBSONSourceStage(selectID int,
	tableName string,
	collation *collation.Collation,
	data []bson.D) *BSONSourceStage {
	return &BSONSourceStage{
		selectID:     selectID,
		databaseName: BSONSourceDB,
		tableName:    tableName,
		collation:    collation,
		data:         data,
	}
}

// BSONSourceIter returns rows from in-memory BSON structs.
type BSONSourceIter struct {
	ctx           *ExecutionCtx
	memoryMonitor *memory.Monitor
	selectID      int
	tableName     string
	databaseName  string
	data          []bson.D
	index         int
	err           error
}

// Open returns an iterator that returns results from executing this plan stage
// with the given ExecutionContext.
func (bs *BSONSourceStage) Open(ctx *ExecutionCtx) (Iter, error) {
	return &BSONSourceIter{
		ctx:           ctx,
		memoryMonitor: ctx.MemoryMonitor(),
		selectID:      bs.selectID,
		databaseName:  bs.databaseName,
		data:          bs.data,
		tableName:     bs.tableName,
		index:         0,
	}, nil
}

// Next populates the provided Row with this iterator's next available row.
// If the iterator has been exhausted or has encountered an error, Next will
// return false, and the value of the provided Row should not be used.
func (bs *BSONSourceIter) Next(row *Row) bool {
	valueKind := GetSQLValueKind(bs.ctx.Variables())

	if bs.index == len(bs.data) || bs.data == nil {
		return false
	}

	var values Values

	for _, docElem := range bs.data[bs.index] {
		value := GoValueToSQLValue(valueKind, docElem.Value)
		values = append(values, NewValue(
			bs.selectID,
			bs.databaseName,
			bs.tableName,
			docElem.Name,
			value))
	}

	row.Data = values
	bs.index++

	bs.err = bs.memoryMonitor.Acquire(row.Data.Size())
	return bs.err == nil
}

// Columns returns the ordered set of columns that are contained in results
// from this plan.
func (bs *BSONSourceStage) Columns() []*Column {

	var columns []*Column
	for _, v := range bs.data[0] {
		column := NewColumn(bs.selectID,
			bs.tableName,
			bs.tableName,
			bs.databaseName,
			v.Name,
			v.Name,
			"",
			EvalNone, schema.MongoNone, false)
		columns = append(columns, column)
	}
	return columns
}

// Collation returns the collation to use for comparisons.
func (bs *BSONSourceStage) Collation() *collation.Collation {
	return bs.collation
}

// Close closes the iterator, returning any error encountered while doing so.
func (bs *BSONSourceIter) Close() error {
	return nil
}

// Err returns any error that has been encountered while iterating. If no error
// was encountered, Err returns nil.
func (bs *BSONSourceIter) Err() error {
	return bs.err
}

func (bs *BSONSourceStage) clone() PlanStage {
	newData := make([]bson.D, len(bs.data))
	return NewBSONSourceStage(bs.selectID, bs.tableName, bs.collation, newData)
}
