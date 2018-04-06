package data

import (
	"fmt"

	"github.com/10gen/sqlproxy/internal/testutils/mongodb"
	"github.com/mongodb/mongo-tools/common/bsonutil"
	toolsdb "github.com/mongodb/mongo-tools/common/db"
	toolsoptions "github.com/mongodb/mongo-tools/common/options"
	"gopkg.in/mgo.v2/bson"
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
	sp.SetFlags(toolsdb.DisableSocketTimeout)

	session, err := sp.GetSession()
	if err != nil {
		return err
	}
	defer session.Close()

	ok, err := mongodb.VersionAtLeast(session, i.MinVersion)
	if err != nil {
		return err
	}

	if !ok {
		return nil
	}

	db := session.DB(i.Db)
	c := db.C(i.Collection)

	err = c.DropCollection()
	if err != nil && err.Error() != toolsdb.ErrNsNotFound {
		return err
	}

	if len(i.Collation) > 0 {
		err = db.Run(bson.D{
			{Name: "create", Value: i.Collection},
			{Name: "collation", Value: i.Collation},
		}, &struct{}{})
		if err != nil {
			return err
		}
	}

	for _, idx := range i.Indexes {
		err = db.Run(bson.D{
			{Name: "createIndexes", Value: i.Collection},
			{Name: "indexes", Value: []interface{}{idx}},
		}, &struct{}{})
		if err != nil {
			return err
		}
	}

	bulk := c.Bulk()

	for _, d := range i.Docs {
		doc, convErr := bsonutil.ConvertJSONValueToBSON(d)
		if convErr != nil {
			return fmt.Errorf("unable to parse extended json %v error: %v", d, convErr)
		}
		bulk.Insert(doc)
	}

	_, err = bulk.Run()
	return err
}
