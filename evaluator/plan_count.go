package evaluator

import (
	"context"
	"fmt"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/procutil"
	"github.com/10gen/sqlproxy/log"
)

// CountStage is a stage for optimizing count(*) against unsharded MongoDB collections.
type CountStage struct {
	mongoSource     *MongoSourceStage
	projectedColumn ProjectedColumn
}

// Children returns a slice of all the Node children of the Node.
func (cs CountStage) Children() []Node {
	return []Node{cs.mongoSource, cs.projectedColumn.Expr}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (cs *CountStage) ReplaceChild(i int, n Node) {
	switch i {
	case 0:
		var ok bool
		cs.mongoSource, ok = n.(*MongoSourceStage)
		if !ok {
			panic(fmt.Sprintf("attempt to convert %v to *MongoSourceStage in ReplaceChild for CountStage", n))
		}
	case 1:
		cs.projectedColumn.Expr = panicIfNotSQLExpr("CountStage", n)
	default:
		panicWithInvalidIndex("CountStage", i, 1)
	}
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

	procutil.PanicSafeGo(func() {
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
func (cs *CountStage) Open(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (RowIter, error) {
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
			NewValueFromColumn(*ci.countColumn, values.NewSQLInt64(ci.cfg.sqlValueKind, int64(ci.count))),
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
