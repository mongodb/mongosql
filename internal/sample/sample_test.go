package sample_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/10gen/sqlproxy/internal/config"
	. "github.com/10gen/sqlproxy/internal/sample"
	"github.com/10gen/sqlproxy/internal/testutils/dbutils"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/mongo"

	"github.com/10gen/mongo-go-driver/bson"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/require"
)

const (
	db1, db2, db3 = "sampleTest1", "sampleTest2", "sampleTest3"
	c1, c2        = "c1", "c2"
)

var (
	doc = []bson.D{{}}
	lgr = log.GlobalLogger()
	ctx = context.Background()
	cfg = config.Default()
)

func init() {
	cfg.Schema.Sample.Source = "sampleStore"
	cfg.Schema.Sample.Namespaces = []string{
		"sampleTest*.*", "sampleStore.*",
	}
}

func TestFetchNamespaces(t *testing.T) {
	provider, err := mongodb.NewSqldSessionProvider(cfg)
	if err != nil {
		t.Fatalf("failed to set up session provider to test server: %v", err)
	}

	session, err := provider.Session(context.Background())
	if err != nil {
		t.Fatalf("failed to set up session to test server: %v", err)
	}
	defer session.Close()

	matcher, err := util.NewMatcher([]string{"*.*"})
	if err != nil {
		t.Fatal(err)
	}

	Convey("With namespaces in the database", t, func() {
		cleanupData(session)
		dbutils.InsertDocuments(session, db1, c1, doc)
		dbutils.InsertDocuments(session, db2, c2, doc)
		dbutils.InsertDocuments(session, db2, c1, doc)

		Convey("fetching the namespaces", func() {
			databases, err := FetchNamespaces(session, lgr, matcher)
			So(err, ShouldBeNil)

			Convey("should return namespaces present", func() {
				So(len(databases[db1]), ShouldEqual, 1)
				So(databases[db1][0], ShouldEqual, c1)
				So(len(databases[db2]), ShouldEqual, 2)
			})

			Convey("should exclude namespaces not present", func() {
				dbutils.DropDatabase(session, db2)
				databases, err := FetchNamespaces(session, lgr, matcher)
				So(err, ShouldBeNil)
				_, found := databases[db1]
				So(found, ShouldBeTrue)
				So(len(databases[db1]), ShouldEqual, 1)
				So(databases[db1][0], ShouldEqual, c1)
				_, found = databases[db2]
				So(found, ShouldBeFalse)
			})
		})
	})
}

func TestInsertSampleRecord(t *testing.T) {
	provider, err := mongodb.NewSqldSessionProvider(cfg)
	if err != nil {
		t.Fatalf("failed to set up session provider to test server: %v", err)
	}

	session, err := provider.Session(context.Background())
	if err != nil {
		t.Fatalf("failed to set up session to test server: %v", err)
	}

	defer session.Close()

	Convey("With a given database", t, func() {
		cleanupData(session)

		Convey("inserting a sample record ", func() {
			version := NewVersion("pname")
			startTime := time.Now()
			version.StartSampleTime = startTime
			endTime := startTime.Add(time.Duration(3 * time.Minute))
			version.EndSampleTime = endTime
			namespace := NewNamespace(db1, c1, version.ID)
			version.Databases = []VersionDatabase{
				{Name: db1, Collections: []string{c1}},
			}
			record := &Record{
				Database:   cfg.Schema.Sample.Source,
				Version:    version,
				Namespaces: []*Namespace{namespace},
			}

			err := InsertSampleRecord(record, session, lgr)
			So(err, ShouldBeNil)

			Convey("should match the version supplied", func() {
				cursor := dbutils.Find(session, cfg.Schema.Sample.Source, VersionsCollection, 1000)
				initialBatch := cursor.InitialBatch()
				if l := len(initialBatch); l != 1 {
					t.Fatalf("unexpected version collection document count: %v", l)
				}
				dbVersion := &Version{}
				err = bson.Unmarshal(initialBatch[0].Data, dbVersion)
				So(err, ShouldBeNil)
				So(dbVersion.ProcessName, ShouldResemble, version.ProcessName)
				So(dbVersion.Databases, ShouldResemble, version.Databases)
				So(dbVersion.Protocol, ShouldResemble, version.Protocol)
				So(dbVersion.Generation, ShouldResemble, version.Generation)

				// MongoDB can only store up to millisecond precision
				So(dbVersion.EndSampleTime.Truncate(time.Millisecond).String(),
					ShouldResemble, version.EndSampleTime.Truncate(time.Millisecond).String())
				So(dbVersion.StartSampleTime.Truncate(time.Millisecond).String(),
					ShouldResemble, version.StartSampleTime.Truncate(time.Millisecond).String())
			})

			Convey("should match the namespace(s) supplied", func() {
				Convey("should match the version supplied", func() {
					cursor := dbutils.Find(session, cfg.Schema.Sample.Source, SchemasCollection, 1000)
					initialBatch := cursor.InitialBatch()
					if l := len(initialBatch); l != 1 {
						t.Fatalf("unexpected schemas collection document count: %v", l)
					}
					dbNamespace := &Namespace{}
					err = bson.Unmarshal(initialBatch[0].Data, dbNamespace)
					So(err, ShouldBeNil)
					// MongoDB can only store up to millisecond precision
					So(dbNamespace.EndSampleTime.Truncate(time.Millisecond).String(),
						ShouldResemble, namespace.EndSampleTime.Truncate(time.Millisecond).String())
					So(dbNamespace.StartSampleTime.Truncate(time.Millisecond).String(),
						ShouldResemble, namespace.StartSampleTime.Truncate(time.Millisecond).String())
					So(dbNamespace.Database, ShouldResemble, namespace.Database)
					So(dbNamespace.Collection, ShouldResemble, namespace.Collection)
					So(dbNamespace.SampleSize, ShouldResemble, namespace.SampleSize)
					So(dbNamespace.VersionID, ShouldResemble, namespace.VersionID)
				})
			})
		})
	})
}

func TestReadSchema(t *testing.T) {
	provider, err := mongodb.NewSqldSessionProvider(cfg)
	if err != nil {
		t.Fatalf("failed to set up session provider to test server: %v", err)
	}

	session, err := provider.Session(context.Background())
	if err != nil {
		t.Fatalf("failed to set up session to test server: %v", err)
	}

	defer session.Close()

	Convey("With a given database", t, func() {
		cleanupData(session)

		Convey("after inserting a valid sample record ", func() {
			version := NewVersion("pname")
			startTime := time.Now()
			version.StartSampleTime = startTime
			endTime := startTime.Add(time.Duration(3 * time.Minute))
			version.EndSampleTime = endTime
			mongoSchema, err := mongo.NewObjectSchema(bson.D{
				{Name: "_id", Value: 10},
				{Name: "name", Value: bson.D{
					{Name: "first", Value: "Jack"},
					{Name: "last", Value: "McJack"},
				}},
				{Name: "addresses", Value: []interface{}{"1", "2", "3"}},
			})
			So(err, ShouldBeNil)

			ns1 := NewNamespace(db1, c1, version.ID)
			ns1.Schema = mongoSchema
			ns2 := NewNamespace(db1, c2, version.ID)
			ns2.Schema = mongoSchema
			ns3 := NewNamespace(db2, c1, version.ID)
			ns3.Schema = mongoSchema

			namespaces := []*Namespace{ns1, ns2, ns3}
			version.Databases = []VersionDatabase{
				{Name: db1, Collections: []string{c1, c2}},
				{Name: db2, Collections: []string{c1}},
			}
			record := &Record{
				Database:   cfg.Schema.Sample.Source,
				Version:    version,
				Namespaces: namespaces,
			}

			err = InsertSampleRecord(record, session, lgr)
			So(err, ShouldBeNil)

			Convey("reading the schema should match the inserted schema", func() {

				schema, err := ReadSchema(&cfg.Schema.Sample, session, lgr)
				So(err, ShouldBeNil)

				dbs := schema.DatabasesSorted()
				So(len(dbs), ShouldEqual, 2)
				schemaDB := dbs[0]
				So(schemaDB.Name(), ShouldEqual, db1)
				So(len(schemaDB.Tables()), ShouldEqual, 4)

				schemaTable := schemaDB.TablesSorted()[0]
				So(schemaTable.SQLName(), ShouldEqual, c1)
				So(schemaTable.MongoName(), ShouldEqual, c1)
				So(len(schemaTable.Pipeline()), ShouldEqual, 0)
				So(len(schemaTable.Columns()), ShouldEqual, 3)

				schemaTable = schemaDB.TablesSorted()[1]
				So(schemaTable.SQLName(), ShouldEqual, c1+"_addresses")
				So(schemaTable.MongoName(), ShouldEqual, c1)
				So(len(schemaTable.Pipeline()), ShouldEqual, 1)
				So(len(schemaTable.Columns()), ShouldEqual, 3)

				schemaTable = schemaDB.TablesSorted()[2]
				So(schemaTable.SQLName(), ShouldEqual, c2)
				So(schemaTable.MongoName(), ShouldEqual, c2)
				So(len(schemaTable.Columns()), ShouldEqual, 3)

				schemaTable = schemaDB.TablesSorted()[3]
				So(schemaTable.SQLName(), ShouldEqual, c2+"_addresses")
				So(schemaTable.MongoName(), ShouldEqual, c2)
				So(len(schemaTable.Columns()), ShouldEqual, 3)

				schemaDB = dbs[1]
				So(schemaDB.Name(), ShouldEqual, db2)
				So(len(schemaDB.Tables()), ShouldEqual, 2)

				schemaTable = schemaDB.TablesSorted()[0]
				So(schemaTable.SQLName(), ShouldEqual, c1)
				So(schemaTable.MongoName(), ShouldEqual, c1)
				So(len(schemaTable.Pipeline()), ShouldEqual, 0)
				So(len(schemaTable.Columns()), ShouldEqual, 3)

				schemaTable = schemaDB.TablesSorted()[1]
				So(schemaTable.SQLName(), ShouldEqual, c1+"_addresses")
				So(schemaTable.MongoName(), ShouldEqual, c1)
				So(len(schemaTable.Pipeline()), ShouldEqual, 1)
				So(len(schemaTable.Columns()), ShouldEqual, 3)
			})
		})

		Convey("after inserting an invalid sample record ", func() {
			version := NewVersion("pname")
			startTime := time.Now()
			version.StartSampleTime = startTime
			endTime := startTime.Add(time.Duration(3 * time.Minute))
			version.EndSampleTime = endTime
			mongoSchema, err := mongo.NewObjectSchema(bson.D{
				{Name: "_id", Value: 10},
				{Name: "name", Value: bson.D{
					{Name: "first", Value: "Jack"},
					{Name: "last", Value: "McJack"},
				}},
				{Name: "addresses", Value: []interface{}{"1", "2", "3"}},
			})
			So(err, ShouldBeNil)

			ns1 := NewNamespace(db1, c1, version.ID)
			ns1.Schema = mongoSchema
			ns2 := NewNamespace(db1, c2, version.ID)
			ns2.Schema = mongoSchema
			ns3 := NewNamespace(db2, c1, version.ID)
			ns3.Schema = mongoSchema

			namespaces := []*Namespace{ns1, ns2, ns3}
			version.Databases = []VersionDatabase{
				{Name: db1, Collections: []string{c1, c2}},
				{Name: db2, Collections: []string{c1, c2}}, // c2 shouldn't be here
			}
			record := &Record{
				Database:   cfg.Schema.Sample.Source,
				Version:    version,
				Namespaces: namespaces,
			}

			err = InsertSampleRecord(record, session, lgr)
			So(err, ShouldBeNil)

			Convey("reading the schema should match the inserted schema", func() {
				_, err := ReadSchema(&cfg.Schema.Sample, session, lgr)
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestSample(t *testing.T) {
	provider, err := mongodb.NewSqldSessionProvider(cfg)
	if err != nil {
		t.Fatalf("failed to set up session provider to test server: %v", err)
	}

	session, err := provider.Session(context.Background())
	if err != nil {
		t.Fatalf("failed to set up session to test server: %v", err)
	}
	defer session.Close()

	Convey("When sampling MongoDB", t, func() {
		cleanupData(session)
		dbutils.InsertDocuments(session, db1, c1, doc)
		dbutils.InsertDocuments(session, db2, c2, doc)
		dbutils.InsertDocuments(session, db2, c1, doc)
		dbutils.InsertDocuments(session, cfg.Schema.Sample.Source, c1, doc)
		// enabling profiling should introduce an additional system.profile
		// collection which should not be sampled
		dbutils.RunCmd(session, db2, bson.D{{Name: "profile", Value: 1}}, &struct{}{})

		opts := &cfg.Schema.Sample
		sampleSchema, sampleRecord, err := Schema(opts, "temp", session, lgr)
		So(err, ShouldBeNil)
		So(sampleSchema, ShouldNotBeNil)
		dbutils.RunCmd(session, db2, bson.D{{Name: "profile", Value: 0}}, &struct{}{})

		So(sampleRecord, ShouldNotBeNil)
		So(sampleRecord.Database, ShouldEqual, cfg.Schema.Sample.Source)
		So(sampleRecord.Version, ShouldNotBeNil)
		So(len(sampleRecord.Namespaces), ShouldNotEqual, 0)

		versionID := sampleRecord.Version.ID

		Convey("namespace version ids should match version id", func() {
			for _, ns := range sampleRecord.Namespaces {
				So(ns.VersionID, ShouldEqual, versionID)
			}
		})

		Convey("version sampling process name should be set", func() {
			So(sampleRecord.Version.ProcessName, ShouldEqual, "temp")
		})

		Convey("whitelisted namespaces should be present", func() {
			db1c1 := NewNamespace(db1, c1, versionID)
			db2c2 := NewNamespace(db2, c2, versionID)
			db2c1 := NewNamespace(db2, c1, versionID)
			sampleNS := NewNamespace(cfg.Schema.Sample.Source, c1, versionID)

			_, found := sampleRecord.Version.FindDatabase(db1)
			So(found, ShouldBeTrue)
			_, found = sampleRecord.Version.FindDatabase(db2)
			So(found, ShouldBeTrue)
			_, found = sampleRecord.Version.FindDatabase(cfg.Schema.Sample.Source)
			So(found, ShouldBeTrue)

			So(sampleRecord.Namespaces, shouldContainNS, db1c1)
			So(sampleRecord.Namespaces, shouldContainNS, db2c2)
			So(sampleRecord.Namespaces, shouldContainNS, db2c1)
			So(sampleRecord.Namespaces, shouldContainNS, sampleNS)
		})

		Convey("blacklisted databases should not be present", func() {
			_, found := sampleRecord.Version.FindDatabase("admin")
			So(found, ShouldBeFalse)
			_, found = sampleRecord.Version.FindDatabase("local")
			So(found, ShouldBeFalse)
			_, found = sampleRecord.Version.FindDatabase("system")
			So(found, ShouldBeFalse)
		})

		Convey("non-existent namespaces should not be present", func() {
			db1c2 := NewNamespace(db1, c2, versionID)
			So(sampleRecord.Namespaces, shouldNotContainNS, db1c2)
			db3c2 := NewNamespace(db3, c2, versionID)
			So(sampleRecord.Namespaces, shouldNotContainNS, db3c2)
			db3c1 := NewNamespace(db3, c1, versionID)
			So(sampleRecord.Namespaces, shouldNotContainNS, db3c1)
			profile := NewNamespace(db2, "system.profile", versionID)
			So(sampleRecord.Namespaces, shouldNotContainNS, profile)
		})

		Convey("blacklisted namespaces should not be present", func() {
			ns1 := NewNamespace(cfg.Schema.Sample.Source, SchemasCollection, versionID)
			So(sampleRecord.Namespaces, shouldNotContainNS, ns1)
			ns1 = NewNamespace(cfg.Schema.Sample.Source, LockCollection, versionID)
			So(sampleRecord.Namespaces, shouldNotContainNS, ns1)
			ns1 = NewNamespace(cfg.Schema.Sample.Source, VersionsCollection, versionID)
			So(sampleRecord.Namespaces, shouldNotContainNS, ns1)
		})

		Convey("sample size should be stored with each namespace", func() {
			for _, ns := range sampleRecord.Namespaces {
				So(ns.SampleSize, ShouldResemble, int64(1))
			}
		})

	})
}

func TestSampleTableAndColumnCollisions(t *testing.T) {
	provider, err := mongodb.NewSqldSessionProvider(cfg)
	if err != nil {
		t.Fatalf("failed to set up session provider to test server: %v", err)
	}

	session, err := provider.Session(context.Background())
	if err != nil {
		t.Fatalf("failed to set up session to test server: %v", err)
	}
	defer session.Close()

	cleanupData(session)

	req := require.New(t)

	doc1 := []bson.M{
		{"XX": 2},
		{"xX_0": 4},
		{"xX": []bson.M{{"c": 1}}},
		{"Xx": []bson.M{{"b": 3}}},
	}

	doc2 := []bson.M{{"hello": 2}}

	t1 := "foo"
	t2 := fmt.Sprintf("%v_Xx_0", t1)
	t3, t4 := "X", "x"
	dbutils.InsertDocuments(session, db1, t1, doc1)
	dbutils.InsertDocuments(session, db1, t2, doc2)
	dbutils.InsertDocuments(session, db1, t3, doc)
	dbutils.InsertDocuments(session, db1, t4, doc)

	opts := &cfg.Schema.Sample
	sampleSchema, sampleRecord, err := Schema(opts, "temp", session, lgr)
	req.Nil(err)
	req.NotNil(sampleSchema)
	dbutils.RunCmd(session, db2, bson.D{{Name: "profile", Value: 0}}, &struct{}{})

	req.NotNil(sampleRecord)
	req.Equal(sampleRecord.Database, cfg.Schema.Sample.Source)
	req.NotNil(sampleRecord.Version)
	req.NotEqual(len(sampleRecord.Namespaces), 0)

	versionID := sampleRecord.Version.ID

	db1c1 := NewNamespace(db1, t1, versionID)
	db1c2 := NewNamespace(db1, t2, versionID)
	db1c3 := NewNamespace(db1, t3, versionID)
	db1c4 := NewNamespace(db1, t4, versionID)

	_, found := sampleRecord.Version.FindDatabase(db1)
	req.True(found)
	req.Empty(shouldContainNS(sampleRecord.Namespaces, db1c1))
	req.Empty(shouldContainNS(sampleRecord.Namespaces, db1c2))
	req.Empty(shouldContainNS(sampleRecord.Namespaces, db1c3))
	req.Empty(shouldContainNS(sampleRecord.Namespaces, db1c4))
	req.Equal(len(sampleSchema.Databases()), 1)

	dbs := sampleSchema.DatabasesSorted()
	req.Equal(dbs[0].Name(), db1)
	req.Equal(len(dbs[0].Tables()), 6)

	type sqlTableMapping struct {
		Table, Collection string
	}

	type sqlColumnMapping struct {
		Column, Field string
	}

	expectedTableMappings := []sqlTableMapping{
		{"foo", "foo"},
		{"foo_xX", "foo"},
		{"foo_Xx_0", "foo_Xx_0"},
		{"foo_Xx_1", "foo"},
		{"X", "X"},
		{"x_0", "x"},
	}

	mappings := []sqlTableMapping{}
	for _, table := range dbs[0].TablesSorted() {
		mapping := sqlTableMapping{table.SQLName(), table.MongoName()}
		mappings = append(mappings, mapping)
	}

	req.Empty(ShouldResemble(mappings, expectedTableMappings))

	getColumnMappings := func(t *schema.Table) (mappings []sqlColumnMapping) {
		for _, c := range t.ColumnsSorted() {
			mapping := sqlColumnMapping{c.SQLName(), c.MongoName()}
			mappings = append(mappings, mapping)
		}
		return mappings
	}

	table := dbs[0].Table("foo_Xx_1")
	req.NotNil(table, "did not find table foo_Xx_1")
	expectedColumnMappings := []sqlColumnMapping{
		{"_id", "_id"}, {"Xx.b", "Xx.b"}, {"Xx_idx", "Xx_idx"},
	}
	req.Empty(ShouldResemble(getColumnMappings(table), expectedColumnMappings))

	table = dbs[0].Table("foo_Xx_0")
	req.NotNil(table, "did not find table foo_Xx_0")
	expectedColumnMappings = []sqlColumnMapping{
		{"_id", "_id"}, {"hello", "hello"},
	}
	req.Empty(ShouldResemble(getColumnMappings(table), expectedColumnMappings))

	table = dbs[0].Table("foo_xX")
	req.NotNil(table, "did not find table foo_xX")
	expectedColumnMappings = []sqlColumnMapping{
		{"_id", "_id"}, {"xX.c", "xX.c"}, {"xX_idx", "xX_idx"},
	}
	req.Empty(ShouldResemble(getColumnMappings(table), expectedColumnMappings))

	table = dbs[0].Table("x_0")
	req.NotNil(table, "did not find table x_0")
	expectedColumnMappings = []sqlColumnMapping{{"_id", "_id"}}
	req.Empty(ShouldResemble(getColumnMappings(table), expectedColumnMappings))

	table = dbs[0].Table("X")
	req.NotNil(table, "did not find table X")
	req.Empty(ShouldResemble(getColumnMappings(table), expectedColumnMappings))

	table = dbs[0].Table("foo")
	req.NotNil(table, "did not find table foo")
	expectedColumnMappings = []sqlColumnMapping{
		{"_id", "_id"}, {"XX", "XX"}, {"xX_0", "xX_0"},
	}
	req.Empty(ShouldResemble(getColumnMappings(table), expectedColumnMappings))

}

func cleanupData(session *mongodb.Session) {
	dbutils.DropDatabase(session, cfg.Schema.Sample.Source)
	dbutils.DropDatabase(session, db1)
	dbutils.DropDatabase(session, db2)
	dbutils.DropDatabase(session, db3)
}

func shouldContainNS(actual interface{}, expected ...interface{}) string {
	namespaces, ok := actual.([]*Namespace)
	if !ok {
		return fmt.Sprintf("expected *Namespace, got %T", actual)
	}

	if l := len(expected); l != 1 {
		return fmt.Sprintf("expected 1 namespace, got %v", l)
	}

	ns, ok := expected[0].(*Namespace)
	if !ok {
		return fmt.Sprintf("expected string, got %T", expected)
	}

	for _, namespace := range namespaces {
		if namespace.Equals(ns) {
			return ""
		}
	}

	return fmt.Sprintf("could not find namespace %v", ns)
}

func shouldNotContainNS(actual interface{}, expected ...interface{}) string {
	namespaces, ok := actual.([]*Namespace)
	if !ok {
		return fmt.Sprintf("expected *Namespace, got %T", actual)
	}

	if l := len(expected); l != 1 {
		return fmt.Sprintf("expected 1 namespace, got %v", l)
	}

	ns, ok := expected[0].(*Namespace)
	if !ok {
		return fmt.Sprintf("expected *Namespace, got %T", expected)
	}

	for _, namespace := range namespaces {
		if namespace.Equals(ns) {
			return fmt.Sprintf("found unexpected namespace %v", ns)
		}
	}

	return ""
}
