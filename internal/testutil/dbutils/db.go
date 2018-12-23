package dbutils

import (
	"context"
	"strings"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/mongo-go-driver/mongo/private/conn"
	"github.com/10gen/mongo-go-driver/mongo/private/msg"
	"github.com/10gen/mongo-go-driver/mongo/private/ops"
	"github.com/10gen/sqlproxy/internal/bsonutil"
)

// CreateIndex creates an index with the provided keys on
// the specified collection.
func CreateIndex(s ops.Server, databaseName, collectionName string, keys []string) {
	indexes := bsonutil.NewD()
	var v interface{}
	for _, k := range keys {
		v = 1
		if strings.HasPrefix(k, "$2d:") {
			k, v = k[4:], "2d"
		}
		indexes = append(indexes, bsonutil.NewDocElem(k, v))
	}
	name := strings.Join(keys, "_")
	indexes = bsonutil.NewD(bsonutil.NewDocElem("key", indexes), bsonutil.NewDocElem("name", name))

	if v != 1 {
		indexes = append(indexes, bsonutil.NewDocElem("bits", 26))
	}

	createIndexCommand := bsonutil.NewD(
		bsonutil.NewDocElem("createIndexes", collectionName),
		bsonutil.NewDocElem("indexes", bsonutil.NewDArray(indexes)),
	)

	request := msg.NewCommand(
		msg.NextRequestID(),
		databaseName,
		false,
		createIndexCommand,
	)

	c, err := s.Connection(context.Background())
	if err != nil {
		panic(err)
	}
	defer func() { _ = c.Close() }()

	d := bsonutil.NewD()
	err = conn.ExecuteCommand(context.Background(), c, request, &d)
	if err != nil {
		panic(err)
	}
}

// DropCollection drops the specified collection.
func DropCollection(s ops.Server, databaseName, collectionName string) {
	c, err := s.Connection(context.Background())
	if err != nil {
		panic(err)
	}
	defer func() { _ = c.Close() }()
	d := bsonutil.NewD()

	err = conn.ExecuteCommand(
		context.Background(),
		c,
		msg.NewCommand(
			msg.NextRequestID(),
			databaseName,
			false, bsonutil.NewD(bsonutil.NewDocElem("drop", collectionName)),
		),

		&d,
	)

	if err != nil {
		errString := err.Error()
		collectionNotFound :=
			!strings.HasSuffix(errString, "collection not found") ||
				!strings.Contains(errString, "NamespaceNotFound")
		if !collectionNotFound {
			panic(err)
		}
	}
}

// DropDatabase drops the specified database.
func DropDatabase(s ops.Server, databaseName string) {
	c, err := s.Connection(context.Background())
	if err != nil {
		panic(err)
	}
	defer func() { _ = c.Close() }()

	d := bsonutil.NewD()
	err = conn.ExecuteCommand(
		context.Background(),
		c,
		msg.NewCommand(
			msg.NextRequestID(),
			databaseName,
			false, bsonutil.NewD(bsonutil.NewDocElem("dropDatabase", 1)),
		),
		&d,
	)
	if err != nil && !strings.HasSuffix(err.Error(), "database not found") {
		panic(err)
	}
}

// Exists checkes whether any documents in the specified collection match
// the provided filter.
func Exists(s ops.Server, databaseName, collectionName string, filter bson.D) bool {
	findCommand := bsonutil.NewD(
		bsonutil.NewDocElem("find", collectionName),
		bsonutil.NewDocElem("filter", filter),
		bsonutil.NewDocElem("limit", 1),
	)

	request := msg.NewCommand(
		msg.NextRequestID(),
		databaseName,
		false,
		findCommand,
	)

	c, err := s.Connection(context.Background())
	if err != nil {
		panic(err)
	}
	defer func() { _ = c.Close() }()

	var result cursorReturningResult

	err = conn.ExecuteCommand(context.Background(), c, request, &result)
	if err != nil {
		panic(err)
	}

	return len(result.Cursor.FirstBatch) > 0
}

// Find executes a find command against the specified collection.
func Find(s ops.Server, databaseName, collectionName string, batchSize int32) ops.CursorResult {
	findCommand := bsonutil.NewD(
		bsonutil.NewDocElem("find", collectionName),
	)

	if batchSize != 0 {
		findCommand = append(findCommand, bsonutil.NewDocElem("batchSize", batchSize))
	}
	request := msg.NewCommand(
		msg.NextRequestID(),
		databaseName,
		false,
		findCommand,
	)

	c, err := s.Connection(context.Background())
	if err != nil {
		panic(err)
	}
	defer func() { _ = c.Close() }()

	var result cursorReturningResult

	err = conn.ExecuteCommand(context.Background(), c, request, &result)
	if err != nil {
		panic(err)
	}

	return &result.Cursor
}

// InsertDocuments inserts the provided documents into the specified collection.
func InsertDocuments(s ops.Server, databaseName, collectionName string, documents interface{}) {
	insertCommand := bsonutil.NewD(
		bsonutil.NewDocElem("insert", collectionName),
		bsonutil.NewDocElem("documents", documents),
	)

	request := msg.NewCommand(
		msg.NextRequestID(),
		databaseName,
		false,
		insertCommand,
	)

	c, err := s.Connection(context.Background())
	if err != nil {
		panic(err)
	}
	defer func() { _ = c.Close() }()

	d := bsonutil.NewD()
	err = conn.ExecuteCommand(context.Background(), c, request, &d)
	if err != nil {
		panic(err)
	}
}

// RunCmd runs the provided command against the specified database.
func RunCmd(s ops.Server, databaseName string, cmd interface{}, result interface{}) {
	request := msg.NewCommand(
		msg.NextRequestID(),
		databaseName,
		false,
		cmd,
	)

	c, err := s.Connection(context.Background())
	if err != nil {
		panic(err)
	}
	defer func() { _ = c.Close() }()

	err = conn.ExecuteCommand(context.Background(), c, request, result)
	if err != nil {
		panic(err)
	}
}

type cursorReturningResult struct {
	Cursor firstBatchCursorResult `bson:"cursor"`
}

type firstBatchCursorResult struct {
	FirstBatch []bson.Raw `bson:"firstBatch"`
	NS         string     `bson:"ns"`
	ID         int64      `bson:"id"`
}

func (cursorResult *firstBatchCursorResult) Namespace() ops.Namespace {
	namespace := ops.ParseNamespace(cursorResult.NS)
	return namespace
}

func (cursorResult *firstBatchCursorResult) InitialBatch() []bson.Raw {
	return cursorResult.FirstBatch
}

func (cursorResult *firstBatchCursorResult) CursorID() int64 {
	return cursorResult.ID
}
