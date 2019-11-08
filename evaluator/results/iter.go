package results

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

// DefaultRowChannelBufSize is the default buffer size for row generator inside RowChanIter
const DefaultRowChannelBufSize = 5

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
	Next(context.Context, *bson.Raw) bool
	// GetColumnInfo returns the slice of ColumnInfo necessary for
	// streaming the results.
	GetColumnInfo() []ColumnInfo
}

// RowChanIter represents an implementation of RowIter using Channel
type RowChanIter struct {
	// generator is the upstream channel sends the row data
	generator <-chan Row
	// done is the channel to interrupt upstream sender, it's listened by upstream sender and sent by the downstream receiver
	done     chan<- struct{}
	isClosed bool

	err error
}

// NewRowChanIter creates a new RowChanIter
func NewRowChanIter(generator <-chan Row, done chan<- struct{}) *RowChanIter {
	return &RowChanIter{generator: generator, done: done, isClosed: false}
}

// NewEmptyRowChanIter creates an empty Row iterator
func NewEmptyRowChanIter(string) RowIter {
	rowChan := make(chan Row)
	done := make(chan struct{})
	close(rowChan)
	return NewRowChanIter(rowChan, done)
}

// Next gets next Row object from the iterator and passes it to the row pointer.
// Returns true if there were no errors and there was a new row assigned to the result pointer
func (c *RowChanIter) Next(ctx context.Context, result *Row) bool {
	// Iter is already closed, return immediately
	if c.isClosed {
		return false
	}

	var ok bool
	*result, ok = <-c.generator
	return ok
}

// Close function will be called when all the data has been sent, or by the receiver when there is an error.
func (c *RowChanIter) Close() error {
	if c.isClosed {
		return nil
	}

	close(c.done)
	c.isClosed = true
	return nil
}

// Err returns any error encountered during iteration.
func (c *RowChanIter) Err() error {
	return c.err
}
