package manager

import (
	"context"
	"testing"
	"time"

	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/drdl"
	"github.com/10gen/sqlproxy/schema/sample"
	"github.com/stretchr/testify/require"

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

	t.Run("alter is allowed", func(t *testing.T) {
		req := require.New(t)
		mgr := setupFileBased()

		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		oldSchema := mgr.getSchema()
		req.NotNil(oldSchema, "schema should be available after starting manager")

		alts := []*schema.Alteration{{
			Type:     schema.RenameTable,
			Db:       "testDb",
			Table:    "testTbl",
			NewTable: "renamedTbl",
		}}
		newSchema, err := mgr.Alter(context.Background(), alts)
		req.NoError(err, "alterations should succeed")
		req.NotEqual(oldSchema, newSchema, "alterations should result in different schema")

		tbl := newSchema.Database("testDb").Tables()[0]
		req.Equal("renamedTbl", tbl.SQLName(), "table should have new SQLName")
		req.Equal("testCol", tbl.MongoName(), "table should have same MongoName")
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

	t.Run("alter not allowed", func(t *testing.T) {
		req := require.New(t)
		mgr, _, _ := setupAutoMode()

		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		sch := mgr.getSchema()
		req.NotNil(sch, "schema should be available after starting manager")

		alts := []*schema.Alteration{}
		newSchema, err := mgr.Alter(context.Background(), alts)
		req.EqualError(err, "alterations not allowed in stored-schema modes")
		req.Nil(newSchema)
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

	t.Run("alter not allowed", func(t *testing.T) {
		req := require.New(t)
		mgr, prst := setupCustomMode()
		prst.populate()

		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		sch := mgr.getSchema()
		req.NotNil(sch, "schema should be available after starting manager")

		alts := []*schema.Alteration{}
		newSchema, err := mgr.Alter(context.Background(), alts)
		req.EqualError(err, "alterations not allowed in stored-schema modes")
		req.Nil(newSchema)
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
}

func TestStandaloneMode(t *testing.T) {
	t.Run("no sample before start", func(t *testing.T) {
		req := require.New(t)
		mgr, smp := setupStandalone()
		req.NotNil(mgr, "ensure manager is actually created and not optimized away")
		req.Zero(smp.started, "no calls to Sample should be made immediately after construction")
		time.Sleep(3 * time.Second)
		req.Zero(smp.started, "no calls to Sample should be made at all before Start() is called")
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

	t.Run("alter schema no alterations", func(t *testing.T) {
		req := require.New(t)
		mgr, _ := setupStandalone()

		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		oldSchema := mgr.getSchema()
		req.NotNil(oldSchema, "schema should be initialized")

		alts := []*schema.Alteration{}
		newSchema, err := mgr.Alter(context.Background(), alts)
		req.NoError(err, "alterations should succeed")
		req.Equal(oldSchema, newSchema, "empty list of alterations should result in same schema")
	})

	t.Run("alter schema before initialization", func(t *testing.T) {
		req := require.New(t)
		mgr, smp := setupStandalone()
		smp.latency = 3 * time.Second

		mgr.Start()
		time.Sleep(300 * time.Millisecond)

		alts := []*schema.Alteration{}
		newSchema, err := mgr.Alter(context.Background(), alts)
		req.EqualError(err, "cannot alter schema before it is initialized")
		req.Nil(newSchema)
	})

	t.Run("alter schema", func(t *testing.T) {
		req := require.New(t)
		mgr, _ := setupStandalone()

		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		oldSchema := mgr.getSchema()
		req.NotNil(oldSchema, "schema should be initialized")

		alts := []*schema.Alteration{{
			Type:     schema.RenameTable,
			Db:       "testDb",
			Table:    "testTbl",
			NewTable: "renamedTbl",
		}}
		newSchema, err := mgr.Alter(context.Background(), alts)
		req.NoError(err, "alterations should succeed")
		req.NotEqual(oldSchema, newSchema, "alterations should result in different schema")

		tbl := newSchema.Database("testDb").Tables()[0]
		req.Equal("renamedTbl", tbl.SQLName(), "table should have new SQLName")
		req.Equal("testCol", tbl.MongoName(), "table should have same MongoName")
	})

	t.Run("alter schema manual resample", func(t *testing.T) {
		req := require.New(t)
		mgr, _ := setupStandalone()

		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		oldSchema := mgr.getSchema()
		req.NotNil(oldSchema, "schema should be initialized")

		alts := []*schema.Alteration{{
			Type:     schema.RenameTable,
			Db:       "testDb",
			Table:    "testTbl",
			NewTable: "renamedTbl",
		}}
		newSchema, err := mgr.Alter(context.Background(), alts)
		req.NoError(err, "alterations should succeed")
		req.NotEqual(oldSchema, newSchema, "alterations should result in different schema")

		_, err = mgr.Resample(context.Background())
		req.NoError(err, "resampling should succeed")
		req.Equal(oldSchema, mgr.getSchema(), "resampling should overwrite the altered schema")
	})

	t.Run("alter schema automatic resample", func(t *testing.T) {
		req := require.New(t)
		mgr, smp, _ := setupWithOpts(StandaloneSchemaMode, 2*time.Second, nil)

		mgr.Start()
		time.Sleep(300 * time.Millisecond)
		oldSchema := mgr.getSchema()
		req.NotNil(oldSchema, "schema should be initialized")

		alts := []*schema.Alteration{{
			Type:     schema.RenameTable,
			Db:       "testDb",
			Table:    "testTbl",
			NewTable: "renamedTbl",
		}}
		newSchema, err := mgr.Alter(context.Background(), alts)
		req.NoError(err, "alterations should succeed")
		req.NotEqual(oldSchema, newSchema, "alterations should result in different schema")

		time.Sleep(2 * time.Second)
		req.Equal(1, smp.started, "no more sampling passes should have started")

		currentSchema := mgr.getSchema()
		req.Equal(newSchema, currentSchema, "automatic resampling should not overwrite altered schema")
		req.EqualError(mgr.GetLastErr(), "skipping automatic resample: schema has been altered")
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

func setupWithOpts(mode SchemaMode, refreshInterval time.Duration, fileBasedSchema *schema.Schema) (*Manager, *testSampler, *testPersistor) {
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
	started   int
	completed int
	latency   time.Duration
}

func (t *testSampler) Sample(ctx context.Context) (*schema.Schema, error) {
	t.started++
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(t.latency):
		// continue below
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
	sch, err := schema.NewFromDRDL(nil, drdlSchema)
	if err != nil {
		panic(err)
	}
	return sch
}
