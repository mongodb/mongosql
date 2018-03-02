package evaluator

import (
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
)

// CountStage is a stage for optimizing count(*) against unsharded MongoDB collections.
type CountStage struct {
	mongoSource     *MongoSourceStage
	projectedColumn ProjectedColumn
}

// CountIter is an iter that iterates over one row, which contains the count for a table.
type CountIter struct {
	called      bool
	count       int
	countColumn *Column
}

// NewCountStage is a constructor that creates a new count stage for a
// mongoSource and projectedColumn.
func NewCountStage(mongoSource *MongoSourceStage,
	projectedColumn ProjectedColumn) *CountStage {
	return &CountStage{mongoSource, projectedColumn}
}

func (cs *CountStage) getCount(ctx *ExecutionCtx) (int, error) {
	errChan := make(chan error, 1)

	var count int
	var err error

	util.PanicSafeGo(func() {
		count,
			err = ctx.Session().Count(cs.mongoSource.dbName,
			cs.mongoSource.collectionNames[0])
		errChan <- err
	}, func(err interface{}) {
		ctx.Logger(log.NetworkComponent).Errf(log.Admin,
			"MongoDB data access session closed: %v", err)
	})

	select {
	case <-ctx.Context().Done():
		return 0, ctx.Context().Err()
	case err = <-errChan:
	}

	return count, err
}

// Open creates a CountIter which iterates one row containing the count.
func (cs *CountStage) Open(ctx *ExecutionCtx) (Iter, error) {
	count, err := cs.getCount(ctx)
	if err != nil {
		return nil, err
	}
	return &CountIter{called: false, count: count, countColumn: cs.projectedColumn.Column}, nil
}

// Columns returns the projected column of count.
func (cs *CountStage) Columns() (columns []*Column) {
	return []*Column{cs.projectedColumn.Column}
}

// Collation returns the collation.
func (cs *CountStage) Collation() *collation.Collation {
	return cs.mongoSource.collation
}

// Next generates a row containing the count and passes it to the row pointer.
func (ci *CountIter) Next(row *Row) bool {
	if !ci.called {
		ci.called = true
		row.Data = Values{NewValueFromColumn(*ci.countColumn, SQLInt(ci.count))}
		return true
	}
	return false
}

// Close closes the iterator.
func (ci *CountIter) Close() error {
	return nil
}

// Err returns any error encountered during iteration.
func (ci *CountIter) Err() error {
	return nil
}
