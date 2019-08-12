//+build integration

package manager

import (
	"context"
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/option"
	"github.com/10gen/sqlproxy/internal/testutil/dbutils"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/x/mongo/driver"
)

func TestWriteModeIntegration(t *testing.T) {
	req := require.New(t)

	cfg := config.Default()
	cfg.Schema.WriteMode = true
	provider, err := mongodb.NewSqldSessionProvider(cfg)
	req.Nil(err)

	session, err := provider.Session(context.Background())
	req.Nil(err)
	defer session.Close()

	dbName := "foo"
	cleanupData(session, dbName)
	defer cleanupData(session, dbName)

	table1Name := "foo"
	table1 := newTableTestHelper(
		log.NoOpLogger(),
		table1Name,
		table1Name,
		[]bson.D{},
		[]*schema.Column{
			schema.NewColumn("a", schema.SQLInt, "a", schema.MongoInt64, false, option.NoneString()),
			schema.NewColumn("b", schema.SQLVarchar, "B", schema.MongoString, false, option.SomeString("fooo")),
			schema.NewColumn("c", schema.SQLVarchar, "C", schema.MongoString, true, option.SomeString("HELLO!")),
		},
		[]schema.Index{
			schema.NewIndex("bAr", true, false,
				[]schema.IndexPart{schema.NewIndexPart("a", 1), schema.NewIndexPart("b", -1)},
			),
			schema.NewIndex("", false, true,
				[]schema.IndexPart{schema.NewIndexPart("b", 1), schema.NewIndexPart("c", 1)},
			),
		},
		option.SomeString("WORLD"),
	)

	table2Name := "bar"
	table2 := newTableTestHelper(
		log.NoOpLogger(),
		table2Name,
		table2Name,
		[]bson.D{},
		[]*schema.Column{
			schema.NewColumn("a", schema.SQLInt, "a", schema.MongoInt64, false, option.NoneString()),
			schema.NewColumn("b", schema.SQLVarchar, "B", schema.MongoString, false, option.SomeString("fooo")),
		},
		[]schema.Index{
			schema.NewIndex("bAr", true, false,
				[]schema.IndexPart{schema.NewIndexPart("a", 1), schema.NewIndexPart("b", -1)},
			),
		},
		option.SomeString("WORLD"),
	)

	mangerCfg := NewMongosqldConfig(&cfg.Schema, variable.NewGlobalContainer(cfg), nil)
	manager := NewManager(mangerCfg, log.NoOpLogger(), provider, "")
	ctx := context.Background()

	manager.Start()
	sch, err := manager.obtainSchema(ctx)
	req.Nil(err)
	manager.setSchema(sch)

	// Test that we fail because 'bar' database does not exist yet.
	_, err = manager.CreateTable(ctx, dbName, table1, session)
	req.NotNil(err)
	req.EqualError(err, "ERROR 1049 (42000): Unknown database 'foo'")

	// Now create database 'foo'.
	sch, err = manager.CreateDatabase(ctx, dbName)
	req.Nil(err)
	req.Nil(sch.Equals(manager.getSchema()), "schema was not properly set in CreateDatabase")

	// Test that we now create the table.
	sch, err = manager.CreateTable(ctx, dbName, table1, session)
	req.Nil(err)
	req.Nil(sch.Equals(manager.getSchema()), "schema was not properly set in CreateTable")

	// Create a second table.
	sch, err = manager.CreateTable(ctx, dbName, table2, session)
	req.Nil(err)
	req.Nil(sch.Equals(manager.getSchema()), "schema was not properly set in CreateDatabase")

	// Check that the table schema is correct.
	schemaDB := sch.Database(dbName)
	schemaTable1 := schemaDB.Table(table1Name)
	// Populate the cache.
	req.Nil(table1.Equals(schemaTable1), "tables should be equal")

	// Check that table2 schema is correct.
	schemaTable2 := schemaDB.Table(table2Name)
	// Populate the cache.
	req.Nil(table2.Equals(schemaTable2), "tables should be equal")

	// Now check that the correct db name exist in MongoDB
	foundDB := findDB(dbName, req, session)
	req.True(foundDB, "did not find db in mongodb")

	// Now check that our both our table names exists, and that they both _have_ a validator.  (that the jsonSchema
	// and indexes written are correct is checked in sample_internal_test).
	foundTable1, hasValidator1, foundTable2, hasValidator2 :=
		findCollectionsInDB(dbName, table1Name, table2Name, req, session)
	req.True(foundTable1, "did not find table1 in mongodb")
	req.True(hasValidator1, "did not find validator1 for table")
	req.True(foundTable2, "did not find table2 in mongodb")
	req.True(hasValidator2, "did not find validator2 for table")

	randomName := "flibbity"

	// Drop a random database that shouldn't exist.
	_, err = manager.DropDatabase(ctx, randomName, session)
	req.NotNil(err)
	req.Equal(fmt.Sprintf("database '%s' cannot be dropped as it does not exist", randomName), err.Error())

	// Drop a random table in a database that doesn't exist.
	_, err = manager.DropTable(ctx, randomName, table1Name, session)
	req.EqualError(err, "ERROR 1049 (42000): Unknown database 'flibbity'")

	// Drop a random table in a database that does exist.
	_, err = manager.DropTable(ctx, dbName, randomName, session)
	req.EqualError(err, "ERROR 1051 (42S02): Unknown table 'flibbity'")

	// Drop our table1 for real this time.
	sch, err = manager.DropTable(ctx, dbName, table1Name, session)
	req.Nil(err)
	req.Nil(sch.Equals(manager.getSchema()), "schema was not properly set in DropTable")

	// Ensure that table1 does not exist but table2 does, in the schema.
	schemaDB = sch.Database(dbName)
	req.NotNil(schemaDB)
	schemaTable1, schemaTable2 = schemaDB.Table(table1Name), schemaDB.Table(table2Name)
	req.Nil(schemaTable1, "table1 should be gone")
	req.NotNil(schemaTable2, "table2 should exist")

	// Now check that table2 name exists, but table1 does not.
	foundTable1, hasValidator1, foundTable2, hasValidator2 =
		findCollectionsInDB(dbName, table1Name, table2Name, req, session)
	req.False(foundTable1, "should not find table1 in mongodb")
	req.False(hasValidator1, "should not find validator1 for table")
	req.True(foundTable2, "did not find table2 in mongodb")
	req.True(hasValidator2, "did not find validator2 for table")

	// Check that the db still exists in MongoDB (though it is implied by
	// the existence of table2 above).
	foundDB = findDB(dbName, req, session)
	req.True(foundDB, "db missing in mongodb")

	// Drop a database that should exist.
	sch, err = manager.DropDatabase(ctx, dbName, session)
	req.Nil(err)
	req.Nil(sch.Equals(manager.getSchema()), "schema was not properly set in DropDatabase")
	schemaDB = sch.Database(dbName)
	req.Nil(schemaDB, "database was not dropped from schema")

	// Now check that the correct db name does not exist in MongoDB.
	foundDB = findDB(dbName, req, session)
	req.False(foundDB, "db was not removed in mongodb")
}

func findDB(dbName string, req *require.Assertions, session *mongodb.Session) bool {
	ctx := context.Background()
	dbs, err := session.ListDatabases(ctx)
	req.Nil(err)
	foundDB := false
	for _, db := range dbs.Databases {
		if db.Name == dbName {
			foundDB = true
		}
	}
	return foundDB
}

func findCollectionsInDB(dbName, table1Name,
	table2Name string,
	req *require.Assertions,
	session *mongodb.Session) (foundTable1, hasValidator1, foundTable2, hasValidator2 bool) {
	ctx := context.Background()
	collectionCursor, err := session.ListCollections(ctx, dbName, driver.CursorOptions{})
	req.Nil(err)
	type collectionResultType struct {
		Name    string `bson:"name"`
		Options bson.D `bson:"options"`
	}
	collectionResult := collectionResultType{}
	foundTable1, foundTable2 = false, false
	hasValidator1, hasValidator2 = false, false
	for collectionCursor.Next(ctx, &collectionResult) {
		if collectionResult.Name == table1Name {
			foundTable1 = true
			for _, field := range collectionResult.Options {
				if field.Key == "validator" {
					hasValidator1 = true
				}
			}
		} else if collectionResult.Name == table2Name {
			foundTable2 = true
			for _, field := range collectionResult.Options {
				if field.Key == "validator" {
					hasValidator2 = true
				}
			}

		}
	}
	return
}

func cleanupData(session *mongodb.Session, databases ...string) {
	for _, db := range databases {
		dbutils.DropDatabase(session, db)
	}
}

func newTableTestHelper(lg log.Logger, tbl, col string,
	pipeline []bson.D, cols []*schema.Column,
	indexes []schema.Index, comment option.String) *schema.Table {
	out, err := schema.NewTable(lg, tbl, col, pipeline, cols, indexes, comment)
	if err != nil {
		panic("this table should not error")
	}
	return out
}
