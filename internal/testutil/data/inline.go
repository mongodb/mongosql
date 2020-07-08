package data

import (
	"context"
	"fmt"

	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/testutil/mongodb"

	toolsdb "github.com/mongodb/mongo-tools-common/db"
	toolsoptions "github.com/mongodb/mongo-tools-common/options"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// InMemoryDataset is a dataset that is stored in a struct.
type InMemoryDataset struct {
	Db         string   `yaml:"db"`
	Collection string   `yaml:"collection"`
	Collation  bson.D   `yaml:"collation"`
	Docs       []bson.D `yaml:"docs"`
	Indexes    []bson.D `yaml:"indexes"`
	MinVersion string   `yaml:"min_server_version"`
}

// Restore restores the in-memory data to the MongoDB deployment specified
// in the provided options.
func (i *InMemoryDataset) Restore(opts *toolsoptions.ToolOptions) error {
	sp, err := toolsdb.NewSessionProvider(*opts)
	if err != nil {
		return err
	}

	defer sp.Close()

	ok, err := mongodb.VersionAtLeast(sp, i.MinVersion)
	if err != nil {
		return err
	}

	if !ok {
		return nil
	}

	db := sp.DB(i.Db)
	c := db.Collection(i.Collection)

	err = c.Drop(context.Background())
	if err != nil {
		return err
	}

	if len(i.Collation) > 0 {
		r := db.RunCommand(context.Background(), bson.D{
			{Key: "create", Value: i.Collection},
			{Key: "collation", Value: i.Collation},
		})
		if r.Err() != nil {
			return r.Err()
		}
	}

	for _, idx := range i.Indexes {
		r := db.RunCommand(context.Background(), bson.D{
			{Key: "createIndexes", Value: i.Collection},
			{Key: "indexes", Value: bson.A{idx}},
		})
		if r.Err() != nil {
			return r.Err()
		}
	}

	// NormalizeBSON will convert extended JSON documents from the
	// yaml file into their corresponding primitive bson types.
	i.Docs, err = bsonutil.NormalizeBSON(i.Docs)
	if err != nil {
		return fmt.Errorf("error normalizing BSON for namespace=%s.%s: %v", i.Db, i.Collection, err)
	}

	docs := make([]mongo.WriteModel, len(i.Docs))
	for j, d := range i.Docs {
		docs[j] = mongo.NewInsertOneModel().SetDocument(d)
	}

	_, err = c.BulkWrite(context.Background(), docs)
	return err
}
