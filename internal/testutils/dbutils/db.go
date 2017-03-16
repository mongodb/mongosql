package dbutils

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/mongo-go-driver/conn"
	"github.com/10gen/mongo-go-driver/msg"
	. "github.com/10gen/mongo-go-driver/ops"
	"github.com/10gen/mongo-go-driver/server"
)

var testServerOnce sync.Once
var testServer Server

func CreateIndex(s Server, databaseName, collectionName string, keys []string) {
	indexes := bson.D{}
	var v interface{}
	for _, k := range keys {
		v = 1
		if strings.HasPrefix(k, "$2d:") {
			k, v = k[4:], "2d"
		}
		indexes = append(indexes, bson.DocElem{k, v})
	}
	name := strings.Join(keys, "_")
	indexes = bson.D{{"key", indexes}, {"name", name}}

	if v != 1 {
		indexes = append(indexes, bson.DocElem{"bits", 26})
	}

	createIndexCommand := bson.D{
		{"createIndexes", collectionName},
		{"indexes", []bson.D{indexes}},
	}

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
	defer c.Close()

	err = conn.ExecuteCommand(context.Background(), c, request, &bson.D{})
	if err != nil {
		panic(err)
	}
}

func DropCollection(s Server, databaseName, collectionName string) {
	c, err := s.Connection(context.Background())
	if err != nil {
		panic(err)
	}
	defer c.Close()

	err = conn.ExecuteCommand(
		context.Background(),
		c,
		msg.NewCommand(
			msg.NextRequestID(),
			databaseName,
			false,
			bson.D{{"drop", collectionName}},
		),
		&bson.D{},
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

func DropDatabase(s Server, databaseName string) {
	c, err := s.Connection(context.Background())
	if err != nil {
		panic(err)
	}
	defer c.Close()

	err = conn.ExecuteCommand(
		context.Background(),
		c,
		msg.NewCommand(
			msg.NextRequestID(),
			databaseName,
			false,
			bson.D{{"dropDatabase", 1}},
		),
		&bson.D{},
	)
	if err != nil && !strings.HasSuffix(err.Error(), "database not found") {
		panic(err)
	}
}

func Find(s Server, databaseName, collectionName string, batchSize int32) CursorResult {
	findCommand := bson.D{
		{"find", collectionName},
	}
	if batchSize != 0 {
		findCommand = append(findCommand, bson.DocElem{"batchSize", batchSize})
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
	defer c.Close()

	var result cursorReturningResult

	err = conn.ExecuteCommand(context.Background(), c, request, &result)
	if err != nil {
		panic(err)
	}

	return &result.Cursor
}

func GetServer(host, port string) *SelectedServer {
	addr := fmt.Sprintf("%v:%v", host, port)
	testServerOnce.Do(func() {
		var err error
		testServer, err = server.New(
			conn.Endpoint(addr),
			server.WithConnectionOptions(
				conn.WithAppName("mongosqld:evaluator"),
			),
		)
		if err != nil {
			panic(fmt.Errorf("failed dialing mongodb server - ensure that one is running at %s: %v", addr, err))
		}
	})

	return &SelectedServer{testServer, nil}
}

func InsertDocuments(s Server, databaseName, collectionName string, documents interface{}) {
	insertCommand := bson.D{
		{"insert", collectionName},
		{"documents", documents},
	}

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
	defer c.Close()

	err = conn.ExecuteCommand(context.Background(), c, request, &bson.D{})
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

func (cursorResult *firstBatchCursorResult) Namespace() Namespace {
	namespace := ParseNamespace(cursorResult.NS)
	return namespace
}

func (cursorResult *firstBatchCursorResult) InitialBatch() []bson.Raw {
	return cursorResult.FirstBatch
}

func (cursorResult *firstBatchCursorResult) CursorID() int64 {
	return cursorResult.ID
}
