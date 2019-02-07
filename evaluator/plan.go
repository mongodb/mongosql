package evaluator

import (
	"context"

	"github.com/10gen/mongo-go-driver/bson"

	"github.com/10gen/sqlproxy/collation"
)

// PlanStage represents a single a node in the Plan tree.
type PlanStage interface {
	Node

	// Open returns an iterator that returns results from executing this plan stage with the given
	// ExecutionContext.
	Open(context.Context, *ExecutionConfig, *ExecutionState) (RowIter, error)

	// Columns returns the ordered set of columns that are contained in results from this plan.
	Columns() []*Column

	// Collation returns the collation to use for comparisons.
	Collation() *collation.Collation

	clone() PlanStage
}

// FastPlanStage is a PlanStage that has a FastOpen method.
type FastPlanStage interface {
	PlanStage

	// FastOpen returns a FastIter that streams bson.RawD documents.
	FastOpen(context.Context, *ExecutionConfig, *ExecutionState) (DocIter, error)
}

// Iter is a super interface for the two types of iterators supported by the BIC.
type Iter interface {
	// Close frees up any resources in use by this iterator. Callers should always
	// call the Close method once they are finished with an iterator.
	Close() error
	// Err returns nil if no errors happened during processing, or the actual
	// error otherwise. Callers should always call the Err method to check whether
	// any error was encountered during processing they are finished with an iterator.
	Err() error
}

// RowIter represents an object that can iterate through a set of rows.
type RowIter interface {
	Iter
	// Next retrieves the next row from this iterator. It returns true if it has
	// additional data and false if there is no more data or if an error occurred
	// during processing.
	//
	// When Next returns false, the Err method should be called to verify if
	// there was an error during processing.
	//
	// For example:
	//    iter, err := plan.Open(ctx);
	//
	//    if err != nil {
	//        return err
	//    }
	//
	//    for iter.Next(&row) {
	//        fmt.Printf("Row: %v\n", row)
	//    }
	//
	//    if err := iter.Close(); err != nil {
	//        return err
	//    }
	//
	//    if err := iter.Err(); err != nil {
	//        return err
	//    }
	//
	Next(context.Context, *Row) bool
}

// DocIter is like RowIter, but yields bson.RawD instead of
// *Row on calls to Next. It is used for performance reasons:
// we can copy less data if we handle unmarshalling ourselves
// with respect to the SQL Wire protocol in question.
type DocIter interface {
	Iter
	// Next retrieves the next row from this iterator. It returns true if it has
	// additional data and false if there is no more data or if an error occurred
	// during processing.
	//
	// When Next returns false, the Err method should be called to verify if
	// there was an error during processing.
	//
	// For example:
	//    iter, err := plan.Open(ctx);
	//
	//    if err != nil {
	//        return err
	//    }
	//
	//    for iter.Next(&doc) {
	//        fmt.Printf("Doc: %v\n", doc)
	//    }
	//
	//    if err := iter.Close(); err != nil {
	//        return err
	//    }
	//
	//    if err := iter.Err(); err != nil {
	//        return err
	//    }
	//
	Next(context.Context, *bson.RawD) bool
	// GetColumnInfo returns the slice of ColumnInfo necessary for
	// streaming the results.
	GetColumnInfo() []ColumnInfo
}
