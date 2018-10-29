package evaluator

import (
	"context"

	"github.com/10gen/sqlproxy/internal/collation"
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
	cfg         *ExecutionConfig
	called      bool
	count       int
	countColumn *Column

	err error
}

// NewCountStage is a constructor that creates a new count stage for a
// mongoSource and projectedColumn.
func NewCountStage(mongoSource *MongoSourceStage,
	projectedColumn ProjectedColumn) *CountStage {
	return &CountStage{mongoSource, projectedColumn}
}

func (cs *CountStage) getCount(ctx context.Context, cfg *ExecutionConfig) (int, error) {
	errChan := make(chan error, 1)

	var count int
	var err error

	util.PanicSafeGo(func() {
		count, err = cfg.commandHandler.Count(
			ctx,
			cs.mongoSource.dbName,
			cs.mongoSource.collectionNames[0],
		)
		errChan <- err
	}, func(err interface{}) {
		cfg.lg.Errf(log.Admin, "MongoDB data access session closed: %v", err)
	})

	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case err = <-errChan:
	}

	return count, err
}

// Open creates a CountIter which iterates one row containing the count.
func (cs *CountStage) Open(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (Iter, error) {
	count, err := cs.getCount(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &CountIter{
		cfg:         cfg,
		called:      false,
		count:       count,
		countColumn: cs.projectedColumn.Column}, nil
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
func (ci *CountIter) Next(ctx context.Context, row *Row) bool {
	if !ci.called {
		ci.called = true
		row.Data = Values{
			NewValueFromColumn(*ci.countColumn, NewSQLInt64(ci.cfg.sqlValueKind, int64(ci.count))),
		}
		ci.err = ci.cfg.memoryMonitor.Acquire(row.Data.Size())
		return ci.err == nil
	}
	return false
}

// Close closes the iterator.
func (ci *CountIter) Close() error {
	return nil
}

// Err returns any error encountered during iteration.
func (ci *CountIter) Err() error {
	return ci.err
}

func (cs *CountStage) clone() PlanStage {
	return NewCountStage(cs.mongoSource.clone().(*MongoSourceStage), cs.projectedColumn)
}
