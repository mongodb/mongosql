package data

import (
	toolsdb "github.com/mongodb/mongo-tools/common/db"
	toolsoptions "github.com/mongodb/mongo-tools/common/options"
	"gopkg.in/mgo.v2/bson"
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
	sp.SetFlags(toolsdb.DisableSocketTimeout)

	session, err := sp.GetSession()
	if err != nil {
		return err
	}

	defer session.Close()

	db := session.DB(i.Database)

	for _, idx := range i.Indexes {
		err = db.Run(bson.D{
			{Name: "createIndexes", Value: i.Collection},
			{Name: "indexes", Value: []interface{}{idx}},
		}, &struct{}{})
		if err != nil {
			return err
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
