package mongodrdl

import (
	"fmt"
	"github.com/10gen/sqlproxy/mongodrdl/mongo"
	"github.com/10gen/sqlproxy/mongodrdl/relational"
	"github.com/mongodb/mongo-tools/common/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

func (schemaGen *SchemaGenerator) Connect() (*mgo.Session, error) {
	session, err := schemaGen.provider.GetSession()
	if err != nil {
		return nil, fmt.Errorf("can't create session: %v", err)
	}

	return session, nil
}

func (schemaGen *SchemaGenerator) ExportSchemaForDatabase() (*relational.Database, error) {
	session, err := schemaGen.Connect()
	if err != nil {
		return nil, err
	}

	log.Logf(log.Info, "Exporting schema for database %q.", schemaGen.ToolOptions.DB)
	db := session.DB(schemaGen.ToolOptions.DB)
	names, err := db.CollectionNames()
	if err != nil {
		return nil, fmt.Errorf("Can't get the collection names for %s,  session: %v", schemaGen.ToolOptions.DB, err)
	}

	database := relational.NewDatabase(schemaGen.ToolOptions.DB)

	for _, name := range names {
		err := schemaGen.mapCollection(database, db.C(name))
		if err != nil {
			return nil, err
		}
	}

	return database, nil
}

func (schemaGen *SchemaGenerator) ExportSchemaForCollection() (*relational.Database, error) {
	session, err := schemaGen.Connect()
	if err != nil {
		return nil, err
	}

	db := session.DB(schemaGen.ToolOptions.DB)
	database := relational.NewDatabase(schemaGen.ToolOptions.DB)
	err = schemaGen.mapCollection(database, db.C(schemaGen.ToolOptions.Collection))
	if err != nil {
		return nil, err
	}

	return database, nil
}

func (schemaGen *SchemaGenerator) mapCollection(database *relational.Database, collection *mgo.Collection) error {
	if strings.HasPrefix(collection.Name, "system.") {
		log.Logf(log.Info, "Skipping system collection %q.", collection.Name)
		return nil
	}

	log.Logf(log.Info, "Exporting tables for %q.", collection.FullName)
	pipeline := collection.Pipe([]bson.M{{"$sample": bson.M{"size": 1000}}}).AllowDiskUse()
	iter := pipeline.Iter()
	if iter.Err() != nil {
		return iter.Err()
	}

	col := mongo.NewCollection(collection.Name)

	var doc bson.D
	for iter.Next(&doc) {
		// NOTE: Perhaps marshal to json???
		log.Logf(log.DebugHigh, "Including sample: %v", doc)
		err := col.IncludeSample(doc)
		if err != nil {
			log.Logf(log.Always, "Error including sample: %+v", doc)
			return err
		}
	}

	// Indexes are needed in order to determine certain
	// types like geo fields.
	indexes, err := collection.Indexes()
	if err != nil {
		return err
	}
	return database.Map(col, indexes)
}
