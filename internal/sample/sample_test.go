package sample_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/config"
	. "github.com/10gen/sqlproxy/internal/sample"
	"github.com/10gen/sqlproxy/internal/testutils/dbutils"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema/mongo"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	db1, db2, db3 = "sampleTest1", "sampleTest2", "sampleTest3"
	c1, c2, c3    = "c1", "c2", "c3"
)

var (
	doc = []bson.D{bson.D{}}
	lgr = log.GlobalLogger()
	ctx = context.Background()
	cfg = config.Default()
)

func init() {
	cfg.Schema.Sample.Source = "sampleStore"
	cfg.Schema.Sample.Namespaces = []string{"sampleTest*.*"}
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

	Convey("With namespaces in the database", t, func() {
		cleanupData(session)
		dbutils.InsertDocuments(session, db1, c1, doc)
		dbutils.InsertDocuments(session, db2, c2, doc)
		dbutils.InsertDocuments(session, db2, c1, doc)

		Convey("fetching the namespaces", func() {
			databases, err := FetchNamespaces(session, &lgr)
			So(err, ShouldBeNil)

			Convey("should return namespaces present", func() {
				So(len(databases[db1]), ShouldEqual, 1)
				So(databases[db1][0], ShouldEqual, c1)
				So(len(databases[db2]), ShouldEqual, 2)
			})

			Convey("should exclude namespaces not present", func() {
				dbutils.DropDatabase(session, db2)
				databases, err := FetchNamespaces(session, &lgr)
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
			namespace := NewNamespace(db1, c1, version.Id)
			version.Databases = []VersionDatabase{
				VersionDatabase{Name: db1, Collections: []string{c1}},
			}
			record := &Record{cfg.Schema.Sample.Source, version, []*Namespace{namespace}}

			err := InsertSampleRecord(record, session, &lgr)
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
					So(dbNamespace.VersionId, ShouldResemble, namespace.VersionId)
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
				{"_id", 10},
				{"name", bson.D{
					{"first", "Jack"},
					{"last", "McJack"},
				}},
				{"addresses", []interface{}{"1", "2", "3"}},
			})
			So(err, ShouldBeNil)

			ns1 := NewNamespace(db1, c1, version.Id)
			ns1.Schema = mongoSchema
			ns2 := NewNamespace(db1, c2, version.Id)
			ns2.Schema = mongoSchema
			ns3 := NewNamespace(db2, c1, version.Id)
			ns3.Schema = mongoSchema

			namespaces := []*Namespace{ns1, ns2, ns3}
			version.Databases = []VersionDatabase{
				VersionDatabase{Name: db1, Collections: []string{c1, c2}},
				VersionDatabase{Name: db2, Collections: []string{c1}},
			}
			record := &Record{cfg.Schema.Sample.Source, version, namespaces}

			err = InsertSampleRecord(record, session, &lgr)
			So(err, ShouldBeNil)

			Convey("reading the schema should match the inserted schema", func() {

				schema, err := ReadSchema(&cfg.Schema.Sample, session, &lgr)
				So(err, ShouldBeNil)

				So(len(schema.Databases), ShouldEqual, 2)

				schemaDB := schema.Databases[0]
				So(schemaDB.Name, ShouldEqual, db1)
				So(len(schemaDB.Tables), ShouldEqual, 4)

				schemaTable := schemaDB.Tables[0]
				So(schemaTable.Name, ShouldEqual, c1)
				So(schemaTable.CollectionName, ShouldEqual, c1)
				So(len(schemaTable.Pipeline), ShouldEqual, 0)
				So(len(schemaTable.Columns), ShouldEqual, 3)

				schemaTable = schemaDB.Tables[1]
				So(schemaTable.Name, ShouldEqual, c1+"_addresses")
				So(schemaTable.CollectionName, ShouldEqual, c1)
				So(len(schemaTable.Pipeline), ShouldEqual, 1)
				So(len(schemaTable.Columns), ShouldEqual, 3)

				schemaTable = schemaDB.Tables[2]
				So(schemaTable.Name, ShouldEqual, c2)
				So(schemaTable.CollectionName, ShouldEqual, c2)
				So(len(schemaTable.Columns), ShouldEqual, 3)

				schemaTable = schemaDB.Tables[3]
				So(schemaTable.Name, ShouldEqual, c2+"_addresses")
				So(schemaTable.CollectionName, ShouldEqual, c2)
				So(len(schemaTable.Columns), ShouldEqual, 3)

				schemaDB = schema.Databases[1]
				So(schemaDB.Name, ShouldEqual, db2)
				So(len(schemaDB.Tables), ShouldEqual, 2)

				schemaTable = schemaDB.Tables[0]
				So(schemaTable.Name, ShouldEqual, c1)
				So(schemaTable.CollectionName, ShouldEqual, c1)
				So(len(schemaTable.Pipeline), ShouldEqual, 0)
				So(len(schemaTable.Columns), ShouldEqual, 3)

				schemaTable = schemaDB.Tables[1]
				So(schemaTable.Name, ShouldEqual, c1+"_addresses")
				So(schemaTable.CollectionName, ShouldEqual, c1)
				So(len(schemaTable.Pipeline), ShouldEqual, 1)
				So(len(schemaTable.Columns), ShouldEqual, 3)
			})
		})

		Convey("after inserting an invalid sample record ", func() {
			version := NewVersion("pname")
			startTime := time.Now()
			version.StartSampleTime = startTime
			endTime := startTime.Add(time.Duration(3 * time.Minute))
			version.EndSampleTime = endTime
			mongoSchema, err := mongo.NewObjectSchema(bson.D{
				{"_id", 10},
				{"name", bson.D{
					{"first", "Jack"},
					{"last", "McJack"},
				}},
				{"addresses", []interface{}{"1", "2", "3"}},
			})
			So(err, ShouldBeNil)

			ns1 := NewNamespace(db1, c1, version.Id)
			ns1.Schema = mongoSchema
			ns2 := NewNamespace(db1, c2, version.Id)
			ns2.Schema = mongoSchema
			ns3 := NewNamespace(db2, c1, version.Id)
			ns3.Schema = mongoSchema

			namespaces := []*Namespace{ns1, ns2, ns3}
			version.Databases = []VersionDatabase{
				VersionDatabase{Name: db1, Collections: []string{c1, c2}},
				VersionDatabase{Name: db2, Collections: []string{c1, c2}}, // c2 shouldn't be here
			}
			record := &Record{cfg.Schema.Sample.Source, version, namespaces}

			err = InsertSampleRecord(record, session, &lgr)
			So(err, ShouldBeNil)

			Convey("reading the schema should match the inserted schema", func() {
				_, err := ReadSchema(&cfg.Schema.Sample, session, &lgr)
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

	shouldContainNS := func(actual interface{}, expected ...interface{}) string {
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

	shouldNotContainNS := func(actual interface{}, expected ...interface{}) string {
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

	Convey("When sampling MongoDB", t, func() {
		cleanupData(session)
		dbutils.InsertDocuments(session, db1, c1, doc)
		dbutils.InsertDocuments(session, db2, c2, doc)
		dbutils.InsertDocuments(session, db2, c1, doc)

		opts := &cfg.Schema.Sample
		schema, sampleRecord, err := SampleSchema(opts, "temp", session, &lgr)
		So(err, ShouldBeNil)
		So(schema, ShouldNotBeNil)

		So(sampleRecord, ShouldNotBeNil)
		So(sampleRecord.Database, ShouldEqual, cfg.Schema.Sample.Source)
		So(sampleRecord.Version, ShouldNotBeNil)
		So(len(sampleRecord.Namespaces), ShouldNotEqual, 0)

		versionId := sampleRecord.Version.Id

		Convey("namespace version ids should match version id", func() {
			for _, ns := range sampleRecord.Namespaces {
				So(ns.VersionId, ShouldEqual, versionId)
			}
		})

		Convey("version sampling process name should be set", func() {
			So(sampleRecord.Version.ProcessName, ShouldEqual, "temp")
		})

		Convey("whitelisted namespaces should be present", func() {
			db1c1 := NewNamespace(db1, c1, versionId)
			db2c2 := NewNamespace(db2, c2, versionId)
			db2c1 := NewNamespace(db2, c1, versionId)

			_, found := sampleRecord.Version.FindDatabase(db1)
			So(found, ShouldBeTrue)
			_, found = sampleRecord.Version.FindDatabase(db2)
			So(found, ShouldBeTrue)
			So(sampleRecord.Namespaces, shouldContainNS, db1c1)
			So(sampleRecord.Namespaces, shouldContainNS, db2c2)
			So(sampleRecord.Namespaces, shouldContainNS, db2c1)
		})

		Convey("blacklisted namespaces should not be present", func() {
			_, found := sampleRecord.Version.FindDatabase("admin")
			So(found, ShouldBeFalse)
			_, found = sampleRecord.Version.FindDatabase("local")
			So(found, ShouldBeFalse)
			_, found = sampleRecord.Version.FindDatabase("system")
			So(found, ShouldBeFalse)
		})

		Convey("non-existent namespaces should not be present", func() {
			db1c2 := NewNamespace(db1, c2, versionId)
			So(sampleRecord.Namespaces, shouldNotContainNS, db1c2)
			db3c2 := NewNamespace(db3, c2, versionId)
			So(sampleRecord.Namespaces, shouldNotContainNS, db3c2)
			db3c1 := NewNamespace(db3, c1, versionId)
			So(sampleRecord.Namespaces, shouldNotContainNS, db3c1)
		})

		Convey("sample size should be stored with each namespace", func() {
			for _, ns := range sampleRecord.Namespaces {
				So(ns.SampleSize, ShouldResemble, int64(1))
			}
		})

	})
}

func cleanupData(session *mongodb.Session) {
	dbutils.DropDatabase(session, cfg.Schema.Sample.Source)
	dbutils.DropDatabase(session, db1)
	dbutils.DropDatabase(session, db2)
	dbutils.DropDatabase(session, db3)
}
