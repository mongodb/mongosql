package dbutils

import (
	"context"
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/internal/bsonutil"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
	"go.mongodb.org/mongo-driver/x/mongo/driver"
	"go.mongodb.org/mongo-driver/x/mongo/driver/operation"
)

// CreateIndex creates an index with the provided keys on
// the specified collection.
func CreateIndex(s driver.Server, db, col string, keys []string) {
	index := bsonutil.NewD()
	var v interface{}
	for _, k := range keys {
		v = 1
		if strings.HasPrefix(k, "$2d:") {
			k, v = k[4:], "2d"
		}
		index = append(index, bsonutil.NewDocElem(k, v))
	}
	name := strings.Join(keys, "_")
	index = bsonutil.NewD(bsonutil.NewDocElem("key", index), bsonutil.NewDocElem("name", name))

	if v != 1 {
		index = append(index, bsonutil.NewDocElem("bits", 26))
	}

	indexesBytes, err := bsonutil.DocSliceToCoreArray(bsonutil.NewDArray(index))
	if err != nil {
		panic(fmt.Errorf("failed to marshal indexes doc: %v", err))
	}

	d := driver.SingleServerDeployment{Server: s}

	c := operation.NewCreateIndexes(indexesBytes).Database(db).Collection(col).Deployment(d)
	err = c.Execute(context.Background())
	if err != nil {
		panic(fmt.Errorf("failed to execute createIndexes: %v", err))
	}
}

// DropCollection drops the specified collection.
func DropCollection(s driver.Server, db, col string) {
	d := driver.SingleServerDeployment{Server: s}
	c := operation.NewDropCollection().Database(db).Collection(col).Deployment(d)
	err := c.Execute(context.Background())
	if err != nil {
		errString := err.Error()
		collectionNotFound :=
			!strings.HasSuffix(errString, "collection not found") ||
				!strings.Contains(errString, "NamespaceNotFound")
		if !collectionNotFound {
			panic(fmt.Errorf("failed to execute dropCollection: %v", err))
		}
	}
}

// DropDatabase drops the specified database.
func DropDatabase(s driver.Server, db string) {
	d := driver.SingleServerDeployment{Server: s}
	c := operation.NewDropDatabase().Database(db).Deployment(d)
	err := c.Execute(context.Background())
	if err != nil && !strings.HasSuffix(err.Error(), "database not found") {
		panic(fmt.Errorf("failed to execute dropDatabase: %v", err))
	}
}

// Find executes a find command against the specified collection.
func Find(s driver.Server, db, col string, batchSize int32) *driver.BatchCursor {
	d := driver.SingleServerDeployment{Server: s}
	c := operation.NewFind(nil).Database(db).Collection(col).Deployment(d).BatchSize(batchSize)
	err := c.Execute(context.Background())
	if err != nil {
		panic(fmt.Errorf("failed to exectute find: %v", err))
	}

	cursor, err := c.Result(driver.CursorOptions{BatchSize: batchSize})
	if err != nil {
		panic(fmt.Errorf("failed to get find result: %v", err))
	}

	return cursor
}

// InsertDocuments inserts the provided documents into the specified collection.
func InsertDocuments(s driver.Server, db, col string, docs []bson.D) {
	documents := make([]bsoncore.Document, len(docs))
	for i, doc := range docs {
		document, err := bson.Marshal(&doc)
		if err != nil {
			panic(fmt.Errorf("failed to marshal insert doc %d: %v", i, err))
		}

		documents[i] = document
	}

	d := driver.SingleServerDeployment{Server: s}
	c := operation.NewInsert(documents...).Database(db).Collection(col).Deployment(d)
	err := c.Execute(context.Background())
	if err != nil {
		panic(fmt.Errorf("failed to execute insert: %v", err))
	}
}

// RunCmd runs the provided command against the specified database.
func RunCmd(s driver.Server, db string, cmd bson.D, result interface{}) {
	cmdBytes, err := bson.Marshal(&cmd)
	if err != nil {
		panic(fmt.Errorf("failed to marshal command: %v", err))
	}

	d := driver.SingleServerDeployment{Server: s}
	c := operation.NewCommand(cmdBytes).Database(db).Deployment(d)
	err = c.Execute(context.Background())
	if err != nil {
		panic(fmt.Errorf("failed to execute command: %v", err))
	}

	err = bson.Unmarshal(c.Result(), &result)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal command result: %v", err))
	}
}
