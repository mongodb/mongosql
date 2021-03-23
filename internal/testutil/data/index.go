package data

import (
	"context"

	toolsdb "github.com/mongodb/mongo-tools/common/db"
	toolsoptions "github.com/mongodb/mongo-tools/common/options"
	"go.mongodb.org/mongo-driver/bson"
)

type indexDataset struct {
	Database   string
	Collection string
	Indexes    []bson.D
}

// Restore restores the indexes for the specified database and collection.
func (i *indexDataset) Restore(opts *toolsoptions.ToolOptions) error {
	sp, err := toolsdb.NewSessionProvider(*opts)
	if err != nil {
		return err
	}

	defer sp.Close()

	db := sp.DB(i.Database)

	for _, idx := range i.Indexes {
		r := db.RunCommand(context.Background(), bson.D{
			{Key: "createIndexes", Value: i.Collection},
			{Key: "indexes", Value: []interface{}{idx}},
		})
		if r.Err() != nil {
			return r.Err()
		}
	}

	return nil
}

// WithIndexes creates a dataset group that restores indexes for a specific collection and database
func WithIndexes(data Dataset, indexes []bson.D, database, collection string) Dataset {
	return DatasetGroup{
		data,
		&indexDataset{
			database,
			collection,
			indexes,
		},
	}
}
