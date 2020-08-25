package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
	"go.mongodb.org/mongo-driver/x/mongo/driver"
)

// Cursor wraps the mongo.batchCursor interface for
// mongosqld and mongodrdl clients. It exposes a
// simpler interface akin to an older version of the
// Go driver.
type Cursor interface {
	// Get the next result from the cursor.
	// Returns true if there were no errors and there was
	// a result unmarshalled into the provided interface{}.
	Next(context.Context, interface{}) bool

	// Get the next result from the cursor as a bson.Raw.
	// Returns true if there were no errors and there was
	// a result unmarshalled into the provided *bson.Raw.
	NextRaw(context.Context, *bson.Raw) bool

	// Returns the error status of the cursor.
	Err() error

	// Close the cursor.
	Close(context.Context) error
}

// NewBatchCursor returns a new Cursor using the provided
// driver.BatchCursor as the backing cursor.
func NewBatchCursor(cursor *driver.BatchCursor) Cursor {
	return &batchCursor{
		cursor: cursor,
	}
}

// NewListCollectionsCursor returns a new Cursor using the provided
// driver.ListCollectionsBatchCursor as the backing cursor.
func NewListCollectionsCursor(cursor *driver.ListCollectionsBatchCursor) Cursor {
	return &listCollectionsCursor{
		cursor: cursor,
	}
}

// batchCursor implements the Cursor interface using the
// driver.BatchCursor as the backing cursor.
type batchCursor struct {
	cursor *driver.BatchCursor

	currentBatch     []bsoncore.Document
	currentBatchSize int
	index            int

	lastErr error

	closed bool
}

func (c *batchCursor) getNextBatch(ctx context.Context) bool {
	if c.lastErr != nil || c.closed {
		return false
	}

	var err error
	if c.currentBatch == nil || c.index == c.currentBatchSize {
		// need to get next batch
		if c.cursor.Next(ctx) {
			c.currentBatch, err = c.cursor.Batch().Documents()
			if err != nil {
				c.lastErr = fmt.Errorf("error getting next batch: %v", err)
				return false
			}
			c.index = 0
			c.currentBatchSize = len(c.currentBatch)
		} else {
			return false
		}
	}

	return true
}

// Get the next result from the cursor.
// Returns true if there were no errors and there was
// a result unmarshalled into the provided interface{}.
func (c *batchCursor) Next(ctx context.Context, res interface{}) bool {
	if !c.getNextBatch(ctx) {
		return false
	}

	err := bson.Unmarshal(c.currentBatch[c.index], res)
	if err != nil {
		c.lastErr = fmt.Errorf("error unmarshaling result: %v", err)
		return false
	}

	c.index++

	return true
}

// Get the next result from the cursor as a bson.Raw.
// Returns true if there were no errors and there was
// a result unmarshalled into the provided *bson.Raw.
func (c *batchCursor) NextRaw(ctx context.Context, doc *bson.Raw) bool {
	if !c.getNextBatch(ctx) {
		return false
	}

	*doc = bson.Raw(c.currentBatch[c.index])
	c.index++
	return true
}

// Returns the error status of the cursor.
func (c *batchCursor) Err() error {
	return c.lastErr
}

// Closes the cursor.
func (c *batchCursor) Close(ctx context.Context) error {
	c.closed = true
	c.lastErr = c.cursor.Close(ctx)
	return c.lastErr
}

// listCollectionsCursor implements the Cursor interface using the
// driver.ListCollectionsBatchCursor as the backing cursor.
type listCollectionsCursor struct {
	cursor *driver.ListCollectionsBatchCursor

	currentBatch     []bsoncore.Document
	currentBatchSize int
	index            int

	lastErr error

	closed bool
}

func (c *listCollectionsCursor) getNextBatch(ctx context.Context) bool {
	if c.lastErr != nil || c.closed {
		return false
	}

	var err error
	if c.currentBatch == nil || c.index == c.currentBatchSize {
		// need to get next batch
		if c.cursor.Next(ctx) {
			c.currentBatch, err = c.cursor.Batch().Documents()
			if err != nil {
				c.lastErr = fmt.Errorf("error getting next batch: %v", err)
				return false
			}
			c.index = 0
			c.currentBatchSize = len(c.currentBatch)
		} else {
			return false
		}
	}

	return true
}

// Get the next result from the cursor.
// Returns true if there were no errors and there was
// a result unmarshalled into the provided interface{}.
func (c *listCollectionsCursor) Next(ctx context.Context, res interface{}) bool {
	if !c.getNextBatch(ctx) {
		return false
	}

	err := bson.Unmarshal(c.currentBatch[c.index], res)
	if err != nil {
		c.lastErr = fmt.Errorf("error unmarshaling result: %v", err)
		return false
	}

	c.index++

	return true
}

// Get the next result from the cursor as a bson.Raw.
// Returns true if there were no errors and there was
// a result unmarshalled into the provided *bson.Raw.
func (c *listCollectionsCursor) NextRaw(ctx context.Context, doc *bson.Raw) bool {
	if !c.getNextBatch(ctx) {
		return false
	}

	*doc = bson.Raw(c.currentBatch[c.index])
	c.index++
	return true
}

// Returns the error status of the cursor.
func (c *listCollectionsCursor) Err() error {
	return c.lastErr
}

// Closes the cursor.
func (c *listCollectionsCursor) Close(ctx context.Context) error {
	c.closed = true
	c.lastErr = c.cursor.Close(ctx)
	return c.lastErr
}
