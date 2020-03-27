package manager

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/option"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/drdl"
	"github.com/10gen/sqlproxy/schema/sample"
	"github.com/stretchr/testify/require"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestFileBasedSchema(t *testing.T) {

	t.Run("initialization", func(t *testing.T) {
		req := require.New(t)
		mgr := setupFileBased()

		sch := mgr.getSchema()
		req.Nil(sch, "schema should not be available before starting manager")

		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		sch = mgr.getSchema()
		req.NotNil(sch, "schema should be available after starting manager")
	})

	t.Run("resample not allowed", func(t *testing.T) {
		req := require.New(t)
		mgr := setupFileBased()

		mgr.Start()
		time.Sleep(300 * time.Millisecond)

		sch, err := mgr.Resample(context.Background())
		req.EqualError(err, "cannot resample when using a file-based schema")
		req.Nil(sch)
		req.NotNil(mgr.getSchema())
	})

	t.Run("test create database not allowed", func(t *testing.T) {
		req := require.New(t)
		mgr := setupFileBased()
		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.NotNil(mgr.getSchema())

		dbName := "writeFoo"
		_, err := mgr.CreateDatabase(context.Background(), dbName)
		req.EqualError(err, "create database only allowed in write mode")
	})

	t.Run("test drop database not allowed", func(t *testing.T) {
		req := require.New(t)
		mgr := setupFileBased()
		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.NotNil(mgr.getSchema())

		dbName := "testDb"
		session := &testUserSession{}
		_, err := mgr.DropDatabase(context.Background(), dbName, session)
		req.EqualError(err, "drop database only allowed in write mode")
	})

	t.Run("test drop table not allowed", func(t *testing.T) {
		req := require.New(t)
		mgr := setupFileBased()
		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.NotNil(mgr.getSchema())

		tableName := "testTbl"
		dbName := "testDb"
		session := &testUserSession{}
		_, err := mgr.DropTable(context.Background(), dbName, tableName, session)
		req.EqualError(err, "drop table only allowed in write mode")
	})

	t.Run("test create table not allowed", func(t *testing.T) {
		req := require.New(t)
		mgr := setupFileBased()
		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.NotNil(mgr.getSchema())

		tableName := "foo"
		table := testTable(tableName, tableName)
		dbName := "testDb"
		session := &testUserSession{}
		_, err := mgr.CreateTable(context.Background(), dbName, table, session)
		req.EqualError(err, "create table only allowed in write mode")
	})
}

func TestAutoMode(t *testing.T) {

	t.Run("initialization", func(t *testing.T) {
		req := require.New(t)
		mgr, _, prst := setupAutoMode()

		req.Nil(mgr.getSchema(), "schema should not be available before starting manager")

		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		sch := mgr.getSchema()
		req.NotNil(sch, "schema should be available after starting manager")

		oid, ok := prst.idsByName["defaultSchema"]
		req.True(ok, "defaultSchema name should have been persisted")

		persistedSchema, ok := prst.schemasByID[oid]
		req.True(ok, "defaultSchema name should point to valid schema")
		req.Equal(sch.ToDRDL(), persistedSchema, "persisted schema should be equivalent to manager's schema")
	})

	t.Run("auto resample should persist", func(t *testing.T) {
		req := require.New(t)
		mgr, smp, prst := setupWithOpts(AutoSchemaMode, 1*time.Second, nil)

		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		sch := mgr.getSchema()
		req.NotNil(sch, "schema should be available")
		req.Equal(1, smp.completed, "exactly one sampling pass should have completed")
		req.Equal(1, prst.schemaInsertCount, "exactly one schema should have been uploaded")
		req.Equal(1, prst.nameUpdateCount, "a schema name should have been updated exactly once")

		time.Sleep(1 * time.Second)
		req.Equal(2, smp.completed, "another sampling pass should have completed")
		req.Equal(2, prst.schemaInsertCount, "a second schema should have been uploaded")
		req.Equal(2, prst.nameUpdateCount, "a schema name should have been updated again")
	})

	t.Run("manual resample should persist", func(t *testing.T) {
		req := require.New(t)
		mgr, smp, prst := setupAutoMode()

		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		sch := mgr.getSchema()
		req.NotNil(sch, "schema should be available")
		req.Equal(1, smp.completed, "exactly one sampling pass should have completed")
		req.Equal(1, prst.schemaInsertCount, "exactly one schema should have been uploaded")
		req.Equal(1, prst.nameUpdateCount, "a schema name should have been updated exactly once")

		_, err := mgr.Resample(context.Background())
		req.NoError(err)
		req.Equal(2, smp.completed, "another sampling pass should have completed")
		req.Equal(2, prst.schemaInsertCount, "a second schema should have been uploaded")
		req.Equal(2, prst.nameUpdateCount, "a schema name should have been updated again")
	})

	t.Run("public getter no refresh", func(t *testing.T) {
		req := require.New(t)
		mgr, smp, _ := setupAutoMode()

		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		sch := mgr.getSchema()
		req.NotNil(sch, "schema should be available")
		req.Equal(1, smp.started, "schema should have been refreshed exactly once")

		sch = mgr.getSchema()
		req.NotNil(sch)
		req.Equal(1, smp.started, "private getter should not trigger refresh")

		sch = mgr.Schema(context.Background())
		req.NotNil(sch)
		req.Equal(1, smp.started, "public getter should not trigger refresh")
	})

	t.Run("test create database not allowed", func(t *testing.T) {
		req := require.New(t)
		mgr, _, _ := setupAutoMode()
		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.NotNil(mgr.getSchema())

		dbName := "writeFoo"
		_, err := mgr.CreateDatabase(context.Background(), dbName)
		req.EqualError(err, "create database only allowed in write mode")
	})

	t.Run("test drop table not allowed", func(t *testing.T) {
		req := require.New(t)
		mgr, _, _ := setupAutoMode()
		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.NotNil(mgr.getSchema())

		tableName := "testTbl"
		dbName := "testDb"
		session := &testUserSession{}
		_, err := mgr.DropTable(context.Background(), dbName, tableName, session)
		req.EqualError(err, "drop table only allowed in write mode")
	})

	t.Run("test create table not allowed", func(t *testing.T) {
		req := require.New(t)
		mgr, _, _ := setupAutoMode()
		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.NotNil(mgr.getSchema())

		tableName := "foo"
		table := testTable(tableName, tableName)
		dbName := "testDb"
		session := &testUserSession{}
		_, err := mgr.CreateTable(context.Background(), dbName, table, session)
		req.EqualError(err, "create table only allowed in write mode")
	})
}

func TestCustomMode(t *testing.T) {

	t.Run("initialization", func(t *testing.T) {
		req := require.New(t)
		mgr, prst := setupCustomMode()
		prst.populate()

		req.Nil(mgr.getSchema(), "schema should not be available before starting manager")

		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		sch := mgr.getSchema()
		req.NotNil(sch, "schema should be available after starting manager")

		persistedSchema := prst.schemasByID[prst.idsByName["defaultSchema"]]
		req.Equal(sch.ToDRDL(), persistedSchema, "persisted schema should be equivalent to manager's schema")
	})

	t.Run("resample not allowed", func(t *testing.T) {
		req := require.New(t)
		mgr, prst := setupCustomMode()
		prst.populate()

		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		sch := mgr.getSchema()
		req.NotNil(sch, "schema should be initialized")

		sch, err := mgr.Resample(context.Background())
		req.EqualError(err, "cannot resample in custom-schema mode")
		req.Nil(sch)
		req.NotNil(mgr.getSchema())
	})

	t.Run("public getter refresh", func(t *testing.T) {
		req := require.New(t)
		mgr, prst := setupCustomMode()
		prst.populate()

		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		sch := mgr.getSchema()
		req.NotNil(sch, "schema should be available")
		req.Equal(1, prst.fetchCount, "schema should have been fetched exactly once")

		sch = mgr.getSchema()
		req.NotNil(sch)
		req.Equal(1, prst.fetchCount, "private getter should not trigger refresh")

		sch = mgr.Schema(context.Background())
		req.NotNil(sch)
		req.Equal(2, prst.fetchCount, "public getter should trigger refresh")
	})

	t.Run("test create database not allowed", func(t *testing.T) {
		req := require.New(t)
		mgr, prst := setupCustomMode()
		prst.populate()

		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.NotNil(mgr.getSchema())

		dbName := "writeFoo"
		_, err := mgr.CreateDatabase(context.Background(), dbName)
		req.EqualError(err, "create database only allowed in write mode")
	})

	t.Run("test drop database not allowed", func(t *testing.T) {
		req := require.New(t)
		mgr, prst := setupCustomMode()
		prst.populate()

		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.NotNil(mgr.getSchema())

		dbName := "testDb"
		session := &testUserSession{}
		_, err := mgr.DropDatabase(context.Background(), dbName, session)
		req.EqualError(err, "drop database only allowed in write mode")
	})

	t.Run("test drop table not allowed", func(t *testing.T) {
		req := require.New(t)
		mgr, prst := setupCustomMode()
		prst.populate()

		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.NotNil(mgr.getSchema())

		tableName := "testTbl"
		dbName := "testDb"
		session := &testUserSession{}
		_, err := mgr.DropTable(context.Background(), dbName, tableName, session)
		req.EqualError(err, "drop table only allowed in write mode")
	})

	t.Run("test create table not allowed", func(t *testing.T) {
		req := require.New(t)
		mgr, prst := setupCustomMode()
		prst.populate()

		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.NotNil(mgr.getSchema())

		tableName := "foo"
		table := testTable(tableName, tableName)
		dbName := "testDb"
		session := &testUserSession{}
		_, err := mgr.CreateTable(context.Background(), dbName, table, session)
		req.EqualError(err, "create table only allowed in write mode")
	})
}

func TestStandaloneMode(t *testing.T) {
	t.Parallel()

	t.Run("no sample before start", func(t *testing.T) {
		req := require.New(t)
		mgr, smp := setupStandalone()
		req.NotNil(mgr, "ensure manager is actually created and not optimized away")
		req.Zero(smp.started, "no calls to Sample should be made immediately after construction")
		time.Sleep(3 * time.Second)
		req.Zero(smp.started, "no calls to Sample should be made at all before Start() is called")
	})

	t.Run("should attempt to initialize until initialization succeeds", func(t *testing.T) {
		req := require.New(t)
		mgr, smp, _ := setupWithOpts(StandaloneSchemaMode, 1*time.Second, nil)
		smp.fail = true

		smp.latency = 100 * time.Millisecond

		mgr.Start()
		time.Sleep(100 * time.Millisecond)
		req.Equal(1, smp.started, "a call to Sample should be made after starting")
		req.Equal(0, smp.completed, "Sample should not complete")

		time.Sleep(5500 * time.Millisecond)
		req.Equal(2, smp.started, "a second call to Sample should be made since it failed to sample")
		req.Equal(0, smp.completed, "Sample should not complete")

		smp.fail = false

		time.Sleep(5500 * time.Millisecond)
		req.Equal(3, smp.started, "a third call to Sample should be made since it failed to sample")
		req.Equal(1, smp.completed, "Sample should complete")

		mgr.Close()
		time.Sleep(6 * time.Second)
		req.Equal(3, smp.started, "no further calls to Sample should have been made")
		req.Equal(1, smp.completed, "Sample should not complete")
	})

	t.Run("close before start", func(t *testing.T) {
		req := require.New(t)
		mgr, _ := setupStandalone()
		req.Panics(mgr.Close, "closing the manager before starting should result in a panic")
	})

	t.Run("close before initial sample completed", func(t *testing.T) {
		req := require.New(t)
		mgr, smp := setupStandalone()
		smp.latency = 2 * time.Second

		mgr.Start()
		time.Sleep(100 * time.Millisecond)
		req.Equal(1, smp.started, "a call to Sample should be made after starting")

		mgr.Close()
		time.Sleep(2 * time.Second)
		req.Equal(0, smp.completed, "the call to Sample should not have completed")
		req.Equal(1, smp.started, "no further calls to Sample should have been made")
		req.Nil(mgr.getSchema(), "the schema should be nil")
	})

	t.Run("close after initial sample completed with refresh interval", func(t *testing.T) {
		req := require.New(t)
		mgr, smp, _ := setupWithOpts(StandaloneSchemaMode, 1*time.Second, nil)

		mgr.Start()
		time.Sleep(100 * time.Millisecond)
		req.Equal(1, smp.started, "a call to Sample should be made after starting")
		req.Equal(1, smp.completed, "the call to Sample should have completed")

		time.Sleep(1 * time.Second)
		req.Equal(2, smp.started, "a second call to Sample should be made after starting")
		req.Equal(2, smp.completed, "the second call to Sample should have completed")

		mgr.Close()
		time.Sleep(2 * time.Second)
		req.Equal(2, smp.started, "no further calls to Sample should have been made")
	})

	t.Run("initial sample without refresh interval", func(t *testing.T) {
		req := require.New(t)
		mgr, smp := setupStandalone()

		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.Equal(1, smp.started, "a call to Sample should be made after starting")

		time.Sleep(3 * time.Second)
		req.Equal(1, smp.started, "no further calls to Sample should be made")
	})

	t.Run("manual resample without refresh interval", func(t *testing.T) {
		req := require.New(t)
		mgr, smp := setupStandalone()

		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.Equal(1, smp.started, "initial sample should have started")
		req.Equal(1, smp.completed, "initial sample should be complete")

		sch, err := mgr.Resample(context.Background())
		req.NoError(err, "manual resample should succeed")
		req.NotNil(sch, "resample should return a schema")
		req.NotNil(mgr.getSchema(), "resample should set a schema")
		req.Equal(2, smp.started, "resample should have triggered a sample")
		req.Equal(2, smp.completed, "resample should have completed")

		sch, err = mgr.Resample(context.Background())
		req.NoError(err, "another manual resample should succeed")
		req.NotNil(sch, "resample should return a schema")
		req.NotNil(mgr.getSchema(), "resample should set a schema")
		req.Equal(3, smp.started, "resample should have triggered another sample")
		req.Equal(3, smp.completed, "resample should have completed")
	})

	t.Run("initial sample with latency", func(t *testing.T) {
		req := require.New(t)
		mgr, smp := setupStandalone()
		smp.latency = 3 * time.Second

		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.Equal(1, smp.started, "initial sample should have started")
		req.Equal(0, smp.completed, "initial sample should not be complete yet")
		req.Nil(mgr.getSchema(), "schema should not be available yet")
		req.NoError(mgr.GetLastErr(), "there should be no sampling error")

		time.Sleep(3 * time.Second)
		req.Equal(1, smp.started, "no more samples should have been started")
		req.Equal(1, smp.completed, "initial sample should be complete")
		req.NotNil(mgr.getSchema(), "schema should be available")
		req.NoError(mgr.GetLastErr(), "there should be no sampling error")
	})

	t.Run("manual refresh during initial sample", func(t *testing.T) {
		req := require.New(t)
		mgr, smp := setupStandalone()
		smp.latency = 3 * time.Second

		// start the schema manager, and sleep long enough for it to issue its
		// initial sample call
		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.Equal(1, smp.started, "initial sample should have started")
		req.Equal(0, smp.completed, "initial sample should not be complete yet")
		req.Nil(mgr.getSchema(), "schema should not be available yet")
		req.NoError(mgr.GetLastErr(), "there should be no sampling error")

		// attempts to resample before schema initialization should fail
		sch, err := mgr.Resample(context.Background())
		req.Nil(sch)
		req.Nil(mgr.getSchema(), "resample should not set a schema")
		req.EqualError(err, "cannot resample before schema is initialized")
		req.Equal(1, smp.started, "no more samples should have started")
		req.Equal(0, smp.completed, "initial sample should still not be complete yet")

		// ensure initial sample still finishes
		time.Sleep(3 * time.Second)
		req.Equal(1, smp.completed, "initial sample should have run to completion")
	})

	t.Run("public getter no refresh", func(t *testing.T) {
		req := require.New(t)
		mgr, smp := setupStandalone()

		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		sch := mgr.getSchema()
		req.NotNil(sch, "schema should be available")
		req.Equal(1, smp.started, "schema should have been refreshed exactly once")

		sch = mgr.getSchema()
		req.NotNil(sch)
		req.Equal(1, smp.started, "private getter should not trigger refresh")

		sch = mgr.Schema(context.Background())
		req.NotNil(sch)
		req.Equal(1, smp.started, "public getter should not trigger refresh")
	})

	t.Run("test create database not allowed", func(t *testing.T) {
		req := require.New(t)
		mgr, _ := setupStandalone()
		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.NotNil(mgr.getSchema())

		dbName := "writeFoo"
		_, err := mgr.CreateDatabase(context.Background(), dbName)
		req.EqualError(err, "create database only allowed in write mode")
	})

	t.Run("test drop database not allowed", func(t *testing.T) {
		req := require.New(t)
		mgr, _ := setupStandalone()
		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.NotNil(mgr.getSchema())

		dbName := "testDb"
		session := &testUserSession{}
		_, err := mgr.DropDatabase(context.Background(), dbName, session)
		req.EqualError(err, "drop database only allowed in write mode")
	})

	t.Run("test drop table not allowed", func(t *testing.T) {
		req := require.New(t)
		mgr, _ := setupStandalone()
		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.NotNil(mgr.getSchema())

		tableName := "testTbl"
		dbName := "testDb"
		session := &testUserSession{}
		_, err := mgr.DropTable(context.Background(), dbName, tableName, session)
		req.EqualError(err, "drop table only allowed in write mode")
	})

	t.Run("test create table not allowed", func(t *testing.T) {
		req := require.New(t)
		mgr, _ := setupStandalone()
		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.NotNil(mgr.getSchema())

		tableName := "foo"
		table := testTable(tableName, tableName)
		dbName := "testDb"
		session := &testUserSession{}
		_, err := mgr.CreateTable(context.Background(), dbName, table, session)
		req.EqualError(err, "create table only allowed in write mode")
	})
}

func TestWriteMode(t *testing.T) {
	t.Parallel()

	t.Run("initialization", func(t *testing.T) {
		req := require.New(t)
		mgr, _ := setupWriteMode()

		sch := mgr.getSchema()
		req.Nil(sch, "schema should not be available before starting manager")

		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		sch = mgr.getSchema()
		req.NotNil(sch, "schema should be available after starting manager")
	})

	t.Run("resample not allowed", func(t *testing.T) {
		req := require.New(t)
		mgr, _ := setupWriteMode()

		mgr.Start()
		time.Sleep(300 * time.Millisecond)

		sch, err := mgr.Resample(context.Background())
		req.EqualError(err, "cannot resample in write mode")
		req.Nil(sch)
		req.NotNil(mgr.getSchema())
	})

	t.Run("should attempt to initialize until initialization succeeds", func(t *testing.T) {
		req := require.New(t)
		mgr, smp, _ := setupWithOpts(WriteSchemaMode, 1*time.Second, nil)
		smp.fail = true

		smp.latency = 100 * time.Millisecond

		mgr.Start()
		time.Sleep(100 * time.Millisecond)
		req.Equal(1, smp.started, "a call to Sample should be made after starting")
		req.Equal(0, smp.completed, "Sample should not complete")

		time.Sleep(5500 * time.Millisecond)
		req.Equal(2, smp.started, "a second call to Sample should be made since it failed to sample")
		req.Equal(0, smp.completed, "Sample should not complete")

		smp.fail = false

		time.Sleep(5500 * time.Millisecond)
		req.Equal(3, smp.started, "a third call to Sample should be made since it failed to sample")
		req.Equal(1, smp.completed, "Sample should complete")

		mgr.Close()
		time.Sleep(6 * time.Second)
		req.Equal(3, smp.started, "no further calls to Sample should have been made")
		req.Equal(1, smp.completed, "Sample should not complete")
	})

	t.Run("should not resample on refresh interval", func(t *testing.T) {
		req := require.New(t)
		mgr, smp, _ := setupWithOpts(WriteSchemaMode, 1*time.Second, nil)

		mgr.Start()
		time.Sleep(100 * time.Millisecond)
		req.Equal(1, smp.started, "a call to Sample should be made after starting")
		req.Equal(1, smp.completed, "the call to Sample should have completed")

		time.Sleep(1 * time.Second)
		req.Equal(1, smp.started, "a second call to Sample should not be made")

		mgr.Close()
		time.Sleep(2 * time.Second)
		req.Equal(1, smp.started, "no further calls to Sample should have been made")
	})

	t.Run("public getter no refresh", func(t *testing.T) {
		req := require.New(t)
		mgr, smp, _ := setupWithOpts(WriteSchemaMode, 1*time.Second, nil)

		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		sch := mgr.getSchema()
		req.NotNil(sch, "schema should be available")
		req.Equal(1, smp.started, "schema should have been refreshed exactly once")

		sch = mgr.getSchema()
		req.NotNil(sch)
		req.Equal(1, smp.started, "private getter should not trigger refresh")

		sch = mgr.Schema(context.Background())
		req.NotNil(sch)
		req.Equal(1, smp.started, "public getter should not trigger refresh")
	})

	t.Run("close before start", func(t *testing.T) {
		req := require.New(t)
		mgr, _ := setupWriteMode()
		req.Panics(mgr.Close, "closing the manager before starting should result in a panic")
	})

	t.Run("close before initial sample completed", func(t *testing.T) {
		req := require.New(t)
		mgr, smp := setupWriteMode()
		smp.latency = 2 * time.Second

		mgr.Start()
		time.Sleep(100 * time.Millisecond)
		req.Equal(1, smp.started, "a call to Sample should be made after starting")

		mgr.Close()
		time.Sleep(2 * time.Second)
		req.Equal(0, smp.completed, "the call to Sample should not have completed")
		req.Equal(1, smp.started, "no further calls to Sample should have been made")
		req.Nil(mgr.getSchema(), "the schema should be nil")
	})

	t.Run("test create database", func(t *testing.T) {
		req := require.New(t)
		mgr, _ := setupWriteMode()
		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.NotNil(mgr.getSchema())

		dbName := "writeFoo"
		sch, err := mgr.CreateDatabase(context.Background(), dbName)
		req.NoError(err)
		req.Nil(sch.Equals(mgr.getSchema()), "manager's schema should be updated")
		database := sch.Database(dbName)
		req.NotNil(database, "database was not added")
	})

	t.Run("test drop database", func(t *testing.T) {
		req := require.New(t)
		mgr, _ := setupWriteMode()
		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.NotNil(mgr.getSchema())

		dbName := "testDb"
		session := &testUserSession{}
		sch, err := mgr.DropDatabase(context.Background(), dbName, session)
		req.NoError(err)
		req.Nil(sch.Equals(mgr.getSchema()), "manager's schema should be updated")
		req.Equal(1, session.callsToDropDatabase)
		req.Equal(0, session.callsToDropCollection)
		req.Equal(0, session.callsToRun)
		database := sch.Database(dbName)
		req.Nil(database, "database was not dropped")
	})

	t.Run("test create table", func(t *testing.T) {
		req := require.New(t)
		mgr, _ := setupWriteMode()
		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.NotNil(mgr.getSchema())

		tableName := "foo"
		table := testTable(tableName, tableName)
		dbName := "testDb"
		session := &testUserSession{}
		sch, err := mgr.CreateTable(context.Background(), dbName, table, session)
		req.NoError(err)
		req.Nil(sch.Equals(mgr.getSchema()), "manager's schema should be updated")
		req.Equal(0, session.callsToDropDatabase)
		req.Equal(0, session.callsToDropCollection)
		req.Equal(2, session.callsToRun)
		database := sch.Database(dbName)
		req.NotNil(database, "database was not added")
		schTable := database.Table(tableName)
		req.NotNil(schTable, "table was not added")
		req.Nil(table.Equals(schTable), "table was not added correctly")
	})

	t.Run("test create table with failed indexes", func(t *testing.T) {
		req := require.New(t)
		mgr, _ := setupWriteMode()
		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.NotNil(mgr.getSchema())

		tableName := "foo"
		table := testTable(tableName, tableName)
		req.Equal(2, len(table.Indexes()), "table should have 2 indexes before creation")
		dbName := "testDb"
		session := &cannotCreateIndexesUserSession{&testUserSession{}}
		sch, err := mgr.CreateTable(context.Background(), dbName, table, session)
		req.EqualError(err, fmt.Sprintf("failed to create indexes for table '%s.%s'", dbName, tableName))
		req.Nil(sch.Equals(mgr.getSchema()), "manager's schema should be updated")
		req.Equal(0, session.callsToDropDatabase)
		req.Equal(0, session.callsToDropCollection)
		req.Equal(2, session.callsToRun)
		database := sch.Database(dbName)
		req.NotNil(database, "database was not added")
		schTable := database.Table(tableName)
		req.NotNil(schTable, "table was not added")
		req.EqualError(table.Equals(schTable), "this table has indexes:\n[]\nother "+
			"table has indexes:\n[{bar bAr true false "+
			"[{a a 1} {b b -1}]} {b_text_c_text b_text_c_text false true [{b b 1} {c c 1}]}]",
			"CreateTable did not DeepCopy the Table")
		table.DropIndexes()
		req.NoError(table.Equals(schTable), "table was not added correctly")
		req.Zero(len(schTable.Indexes()), "there should be no indexes")
	})

	t.Run("test drop table", func(t *testing.T) {
		req := require.New(t)
		mgr, _ := setupWriteMode()
		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.NotNil(mgr.getSchema())

		tableName := "testTbl"
		dbName := "testDb"
		session := &testUserSession{}
		sch, err := mgr.DropTable(context.Background(), dbName, tableName, session)
		req.NoError(err)
		req.Nil(sch.Equals(mgr.getSchema()), "manager's schema should be updated")
		req.Equal(0, session.callsToDropDatabase)
		req.Equal(1, session.callsToDropCollection)
		req.Equal(0, session.callsToRun)
		database := sch.Database(dbName)
		req.NotNil(database, "database was not added")
		schTable := database.Table(tableName)
		req.Nil(schTable, "table was not dropped")
	})

	t.Run("test create database failure", func(t *testing.T) {
		req := require.New(t)
		mgr, _ := setupWriteMode()
		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.NotNil(mgr.getSchema())

		dbLen := len(mgr.getSchema().Databases())
		dbName := "testDb"
		_, err := mgr.CreateDatabase(context.Background(), dbName)
		req.EqualError(err, fmt.Sprintf("database %q already exists in schema", dbName))
		req.Equal(dbLen, len(mgr.getSchema().Databases()), "no database should have been added")
	})

	t.Run("test drop database failure", func(t *testing.T) {
		req := require.New(t)
		mgr, _ := setupWriteMode()
		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.NotNil(mgr.getSchema())

		dbName := "random"
		session := &testUserSession{}
		sch := mgr.getSchema()
		dbLen := len(sch.Databases())
		_, err := mgr.DropDatabase(context.Background(), dbName, session)
		req.EqualError(err, fmt.Sprintf("database '%s' cannot be dropped as it does not exist", dbName))
		req.Equal(dbLen, len(mgr.getSchema().Databases()), "no database should have been dropped")

		dbName = "testDb"
		badSession := reallyBadSession{}
		_, err = mgr.DropDatabase(context.Background(), dbName, badSession)
		req.EqualError(err, "this is a really bad session that fails at everything")
		req.Equal(dbLen, len(mgr.getSchema().Databases()), "no database should have been dropped")
	})

	t.Run("test drop table failure", func(t *testing.T) {
		req := require.New(t)
		mgr, _ := setupWriteMode()
		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		req.NotNil(mgr.getSchema())

		tableName := "testTbl"
		dbName := "random"
		session := &testUserSession{}
		_, err := mgr.DropTable(context.Background(), dbName, tableName, session)
		req.EqualError(err, "ERROR 1049 (42000): Unknown database 'random'")
		tableName = "random"
		sch := mgr.getSchema()
		dbName = "testDb"
		tblLen := len(sch.Database(dbName).Tables())
		_, err = mgr.DropTable(context.Background(), dbName, tableName, session)
		req.EqualError(err, "ERROR 1051 (42S02): Unknown table 'random'")
		req.Equal(tblLen, len(mgr.getSchema().Database(dbName).Tables()), "no table should have been dropped")

		dbName = "testDb"
		tableName = "testTbl"
		badSession := reallyBadSession{}
		_, err = mgr.DropTable(context.Background(), dbName, tableName, badSession)
		req.EqualError(err, "this is a really bad session that fails at everything")
		req.Equal(tblLen, len(mgr.getSchema().Database(dbName).Tables()), "no table should have been dropped")
	})

	t.Run("test create table failure", func(t *testing.T) {
		req := require.New(t)
		mgr, _ := setupWriteMode()
		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		sch := mgr.getSchema()
		req.NotNil(sch)
		tableName := "testTbl"
		colName := "testCol"
		table := testTable(tableName, colName)
		dbName := "testDb"
		session := &testUserSession{}
		tableLen := len(sch.Database(dbName).Tables())
		_, err := mgr.CreateTable(context.Background(), dbName, table, session)
		req.EqualError(err, "ERROR 1050 (42S01): Table 'testTbl' already exists")
		req.Equal(1, tableLen, "table should not have been added")

		badSession := reallyBadSession{}
		_, err = mgr.CreateTable(context.Background(), dbName, table, badSession)
		req.EqualError(err, "this is a really bad session that fails at everything")
	})

}

type testUserSession struct {
	callsToDropDatabase   int
	callsToDropCollection int
	callsToRun            int
}

func (t *testUserSession) DropDatabase(context.Context, string) error {
	t.callsToDropDatabase++
	return nil
}

func (t *testUserSession) DropCollection(context.Context, string, string) error {
	t.callsToDropCollection++
	return nil

}

func (t *testUserSession) Run(context.Context, string, bson.D, interface{}) error {
	t.callsToRun++
	return nil
}

type reallyBadSession struct{}

func (reallyBadSession) DropDatabase(context.Context, string) error {
	return fmt.Errorf("this is a really bad session that fails at everything")
}

func (reallyBadSession) DropCollection(context.Context, string, string) error {
	return fmt.Errorf("this is a really bad session that fails at everything")
}

func (reallyBadSession) Run(context.Context, string, bson.D, interface{}) error {
	return fmt.Errorf("this is a really bad session that fails at everything")
}

type cannotCreateIndexesUserSession struct {
	*testUserSession
}

func (t *cannotCreateIndexesUserSession) Run(ctx context.Context, db string, cmd bson.D, result interface{}) error {
	t.callsToRun++
	if cmd[0].Key == "createIndexes" {
		return fmt.Errorf("failed to create indexes")
	}
	return nil
}

func setupFileBased() *Manager {
	mgr, _, _ := setupWithOpts(FileBasedSchemaMode, 0, testSchema())
	return mgr
}

func setupStandalone() (*Manager, *testSampler) {
	mgr, smp, _ := setupWithOpts(StandaloneSchemaMode, 0, nil)
	return mgr, smp
}

func setupAutoMode() (*Manager, *testSampler, *testPersistor) {
	return setupWithOpts(AutoSchemaMode, 0, nil)
}

func setupCustomMode() (*Manager, *testPersistor) {
	mgr, _, prst := setupWithOpts(CustomSchemaMode, 0, nil)
	return mgr, prst
}

func setupWriteMode() (*Manager, *testSampler) {
	mgr, smp, _ := setupWithOpts(WriteSchemaMode, 0, nil)
	return mgr, smp
}

func setupWithOpts(mode SchemaMode, refreshInterval time.Duration,
	fileBasedSchema *schema.Schema) (*Manager, *testSampler, *testPersistor) {
	cfg := newTestConfig(mode, refreshInterval, fileBasedSchema)
	lg := log.NoOpLogger()
	sampler := &testSampler{}
	persistor := newTestPersistor()
	mgr := newManager(cfg, lg, persistor, sampler)
	return mgr, sampler, persistor
}

type testConfig struct {
	mode            SchemaMode
	refreshInterval time.Duration
	schemaName      string
	fileBasedSchema *schema.Schema
}

func newTestConfig(mode SchemaMode, refreshInterval time.Duration, fileBasedSchema *schema.Schema) Config {
	return testConfig{
		mode:            mode,
		refreshInterval: refreshInterval,
		schemaName:      "defaultSchema",
		fileBasedSchema: fileBasedSchema,
	}
}

func (tc testConfig) Mode() SchemaMode {
	return tc.mode
}

func (tc testConfig) RefreshInterval() time.Duration {
	return tc.refreshInterval
}

func (tc testConfig) SampleConfig() sample.Config {
	cfg := config.Default()
	return sample.NewMongosqldConfig(&cfg.Schema, nil)
}

func (tc testConfig) SchemaName() string {
	return tc.schemaName
}

func (tc testConfig) FileBasedSchema() *schema.Schema {
	return tc.fileBasedSchema
}

type testSampler struct {
	// started is the number of times sampling has started.
	started int
	// completed is the number of times sampling has successfully completed.
	completed int
	// fail tells the sampler to fail to sample.
	fail bool
	// latency is the length of time before sampling should complete or fail, if fail is true.
	latency time.Duration
}

func (t *testSampler) Sample(ctx context.Context) (*schema.Schema, error) {
	t.started++
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(t.latency):
		// continue below
	}
	if t.fail {
		return nil, fmt.Errorf("failed to sample")
	}
	t.completed++
	return testSchema(), nil
}

type testPersistor struct {
	schemasByID map[primitive.ObjectID]*drdl.Schema
	idsByName   map[string]primitive.ObjectID

	fetchCount        int
	schemaInsertCount int
	nameUpdateCount   int
}

func newTestPersistor() *testPersistor {
	return &testPersistor{
		schemasByID: make(map[primitive.ObjectID]*drdl.Schema),
		idsByName:   make(map[string]primitive.ObjectID),
	}
}

func (t *testPersistor) populate() {
	drdlSchema := testSchema().ToDRDL()

	bgCtx := context.Background()

	oid, _ := t.InsertSchema(bgCtx, drdlSchema)
	t.schemaInsertCount--

	_ = t.UpsertName(bgCtx, "defaultSchema", oid)
	t.nameUpdateCount--
}

func (t *testPersistor) InsertSchema(_ context.Context, ds *drdl.Schema) (primitive.ObjectID, error) {
	oid := primitive.NewObjectID()
	t.schemasByID[oid] = ds
	t.schemaInsertCount++
	return oid, nil
}

func (t *testPersistor) UpsertName(_ context.Context, name string, schemaID primitive.ObjectID) error {
	t.idsByName[name] = schemaID
	t.nameUpdateCount++
	return nil
}

func (t *testPersistor) FindSchemaByName(_ context.Context, name string) (*drdl.Schema, error) {
	oid, ok := t.idsByName[name]
	if !ok {
		panic("invalid schema name")
	}
	sch, ok := t.schemasByID[oid]
	if !ok {
		panic("invalid schema id")
	}
	t.fetchCount++
	return sch, nil
}

func testSchema() *schema.Schema {
	drdlSchema := &drdl.Schema{
		Databases: []*drdl.Database{{
			Name: "testDb",
			Tables: []*drdl.Table{{
				SQLName:   "testTbl",
				MongoName: "testCol",
				Columns: []*drdl.Column{{
					MongoName: "testField",
					MongoType: "bool",
					SQLName:   "testCol",
					SQLType:   "boolean",
				}},
			}},
		}},
	}
	sch, err := schema.NewFromDRDL(nil, drdlSchema, false)
	if err != nil {
		panic(err)
	}
	return sch
}

func testTable(tableName, colName string) *schema.Table {
	ret, err := schema.NewTable(
		log.NoOpLogger(),
		tableName,
		colName,
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
		false,
	)
	if err != nil {
		panic("this table should not error")
	}
	return ret
}
