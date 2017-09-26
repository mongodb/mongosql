package mongodrdl

import (
	"context"
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/mongodrdl/mongo"
	"github.com/10gen/sqlproxy/mongodrdl/relational"

	"github.com/10gen/mongo-go-driver/bson"
)

func (schemaGen *SchemaGenerator) Connect() (*mongodb.Session, error) {
	session, err := schemaGen.Provider.Session(context.Background())
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
	defer session.Close()

	schemaGen.Logger.Infof(log.Admin, "Creating schema for database %q", schemaGen.ToolOptions.DB)

	iter, err := session.ListCollections(schemaGen.ToolOptions.DB)
	if err != nil {
		return nil, fmt.Errorf("Can't get the collection names for %s: %v", schemaGen.ToolOptions.DB, err)
	}

	var colResult struct {
		Name string `bson:"name"`
	}

	database := relational.NewDatabase(schemaGen.ToolOptions.DB, schemaGen.Logger)

	ctx := session.Context()

	for iter.Next(ctx, &colResult) {
		err := schemaGen.mapCollection(database, colResult.Name, session)
		if err != nil {
			return nil, err
		}
	}

	if err := iter.Close(ctx); err != nil {
		return nil, err
	}

	schemaGen.Logger.Infof(log.Admin, "Created schema for database %q", schemaGen.ToolOptions.DB)

	return database, nil
}

func (schemaGen *SchemaGenerator) ExportSchemaForCollection() (*relational.Database, error) {
	session, err := schemaGen.Connect()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	database := relational.NewDatabase(schemaGen.ToolOptions.DB, schemaGen.Logger)
	err = schemaGen.mapCollection(database, schemaGen.ToolOptions.Collection, session)
	if err != nil {
		return nil, err
	}

	return database, nil
}

func (schemaGen *SchemaGenerator) mapCollection(database *relational.Database, collectionName string, session *mongodb.Session) error {
	dbName := schemaGen.ToolOptions.DB
	if strings.HasPrefix(collectionName, "system.") {
		schemaGen.Logger.Infof(log.Admin, "Skipping system collection %q", collectionName)
		return nil
	}

	schemaGen.Logger.Infof(log.Admin, "Creating schema for namespace %q.%q", dbName, collectionName)
	pipeline := []bson.M{{"$sample": bson.M{"size": schemaGen.SampleOptions.Size}}}

	iter, err := session.Aggregate(dbName, collectionName, pipeline)
	if err != nil {
		return err
	}

	col := mongo.NewCollection(collectionName)
	ctx := session.Context()
	doc := &bson.D{}
	var samplePrint string

	for iter.Next(ctx, doc) {
		err = col.IncludeSample(*doc)
		if err != nil {
			schemaGen.Logger.Infof(log.Always, "Error including sample: %#v", samplePrint)
			return err
		}

		doc = &bson.D{}
	}

	if err := iter.Close(ctx); err != nil {
		return err
	}

	if database.Views == nil {
		type colResult struct {
			Name string `bson:"name"`
			Type string `bson:"type"`
		}

		results := colResult{}

		iter, err := session.ListCollections(dbName)
		if err != nil {
			return fmt.Errorf("failed to run listCollections on database '%v': %v", dbName, err)
		}

		database.Views = map[string]struct{}{}

		for iter.Next(ctx, &results) {
			if results.Type == "view" {
				database.Views[results.Name] = struct{}{}
			}
			results = colResult{}
		}

		if err = iter.Close(ctx); err != nil {
			return err
		}

	}

	if _, ok := database.Views[collectionName]; ok {
		return database.Map(col, []mongodb.Index{}, schemaGen.OutputOptions.PreJoined)
	}

	// Indexes are needed in order to determine certain
	// types like geo fields.
	iter, err = session.ListIndexes(dbName, collectionName)
	if err != nil {
		return fmt.Errorf("failed to run listIndexes on database %q: %v", dbName, err)
	}

	indexes, index := []mongodb.Index{}, mongodb.Index{}

	for iter.Next(ctx, &index) {
		indexes = append(indexes, index)
	}

	if err = iter.Close(ctx); err != nil {
		return err
	}

	err = database.Map(col, indexes, schemaGen.OutputOptions.PreJoined)
	if err != nil {
		return err
	}

	schemaGen.Logger.Infof(log.Admin, "Created schema for namespace %q.%q", dbName, collectionName)
	return nil
}
