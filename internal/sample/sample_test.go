package sample_test

import (
	"context"
	"fmt"
	"strings"
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

	req := require.New(t)

	cleanupData(session)
	dbutils.InsertDocuments(session, db1, c1, doc)
	dbutils.InsertDocuments(session, db2, c2, doc)
	dbutils.InsertDocuments(session, db2, c1, doc)

	mappings, err := FetchNamespaces(session, lgr, matcher)
	req.Nil(err, "error fetching namespaces")

	req.Equal(len(mappings[db1]), 1)
	req.Equal(mappings[db1][0], c1)
	req.Equal(len(mappings[db2]), 2)

	dbutils.DropDatabase(session, db2)
	mappings, err = FetchNamespaces(session, lgr, matcher)
	req.Nil(err, "error fetching namespaces")
	_, found := mappings[db1]

	errFound := "found unexpected database"
	errMissing := "could not find expected database"
	req.True(found, errMissing)

	req.Equal(len(mappings[db1]), 1)
	req.Equal(mappings[db1][0], c1)
	_, found = mappings[db2]
	req.False(found, errFound)
	_, found = mappings["admin"]
	req.False(found, errFound)
	_, found = mappings["config"]
	req.False(found, errFound)
	_, found = mappings["local"]
	req.False(found, errFound)
	_, found = mappings["system"]
	req.False(found, errFound)
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
					cursor := dbutils.Find(
						session,
						cfg.Schema.Sample.Source,
						SchemasCollection,
						1000,
					)
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
					So(
						dbNamespace.StartSampleTime.Truncate(time.Millisecond).String(),
						ShouldResemble,
						namespace.StartSampleTime.Truncate(time.Millisecond).String(),
					)
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

func TestSchema(t *testing.T) {
	provider, err := mongodb.NewSqldSessionProvider(cfg)
	if err != nil {
		t.Fatalf("failed to set up session provider to test server: %v", err)
	}

	session, err := provider.Session(context.Background())
	if err != nil {
		t.Fatalf("failed to set up session to test server: %v", err)
	}
	defer session.Close()

	req := require.New(t)

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
	req.Nilf(err, "did not expect error in sampling")
	req.NotNilf(sampleSchema, "did not expect sample schema to be nil")
	dbutils.RunCmd(session, db2, bson.D{{Name: "profile", Value: 0}}, &struct{}{})

	req.NotNilf(sampleRecord, "did not expect sample record to be nil")
	req.Equalf(sampleRecord.Database, cfg.Schema.Sample.Source, "mismatched sample source")
	req.NotNilf(sampleRecord, "did not expect sample record version to be nil")
	req.NotEqualf(len(sampleRecord.Namespaces), 0, "found no sampled namespaces")

	versionID := sampleRecord.Version.ID

	for _, ns := range sampleRecord.Namespaces {
		req.Equalf(ns.VersionID, versionID, "namespace version ids should match version id")
	}

	req.Equalf(sampleRecord.Version.ProcessName, "temp",
		"version sampling process name should be set")

	db1c1 := NewNamespace(db1, c1, versionID)
	db2c2 := NewNamespace(db2, c2, versionID)
	db2c1 := NewNamespace(db2, c1, versionID)
	sampleNS := NewNamespace(cfg.Schema.Sample.Source, c1, versionID)

	errMsg := "whitelisted namespaces should be present"
	_, found := sampleRecord.Version.FindDatabase(db1)
	req.Truef(found, errMsg)
	_, found = sampleRecord.Version.FindDatabase(db2)
	req.Truef(found, errMsg)
	_, found = sampleRecord.Version.FindDatabase(cfg.Schema.Sample.Source)
	req.Truef(found, errMsg)

	req.Emptyf(shouldContainNS(sampleRecord.Namespaces, db1c1), errMsg)
	req.Emptyf(shouldContainNS(sampleRecord.Namespaces, db2c2), errMsg)
	req.Emptyf(shouldContainNS(sampleRecord.Namespaces, db2c1), errMsg)
	req.Emptyf(shouldContainNS(sampleRecord.Namespaces, sampleNS), errMsg)

	errMsg = "non-existent namespaces should not be present"

	_, found = sampleRecord.Version.FindDatabase("admin")
	req.Falsef(found, errMsg)
	_, found = sampleRecord.Version.FindDatabase("config")
	req.Falsef(found, errMsg)
	_, found = sampleRecord.Version.FindDatabase("local")
	req.Falsef(found, errMsg)
	_, found = sampleRecord.Version.FindDatabase("system")
	req.Falsef(found, errMsg)

	errMsg = "non-existent namespaces should not be present"

	db1c2 := NewNamespace(db1, c2, versionID)
	req.Emptyf(shouldNotContainNS(sampleRecord.Namespaces, db1c2), errMsg)
	db3c2 := NewNamespace(db3, c2, versionID)
	req.Emptyf(shouldNotContainNS(sampleRecord.Namespaces, db3c2), errMsg)
	db3c1 := NewNamespace(db3, c1, versionID)
	req.Emptyf(shouldNotContainNS(sampleRecord.Namespaces, db3c1), errMsg)
	profile := NewNamespace(db2, "system.profile", versionID)
	req.Emptyf(shouldNotContainNS(sampleRecord.Namespaces, profile), errMsg)

	errMsg = "special sampling namespaces should not be present"
	ns := NewNamespace(cfg.Schema.Sample.Source, SchemasCollection, versionID)
	req.Emptyf(shouldNotContainNS(sampleRecord.Namespaces, ns), errMsg)
	ns = NewNamespace(cfg.Schema.Sample.Source, LockCollection, versionID)
	req.Emptyf(shouldNotContainNS(sampleRecord.Namespaces, ns), errMsg)
	ns = NewNamespace(cfg.Schema.Sample.Source, VersionsCollection, versionID)
	req.Emptyf(shouldNotContainNS(sampleRecord.Namespaces, ns), errMsg)

	for _, ns := range sampleRecord.Namespaces {
		req.Equalf(ns.SampleSize, int64(1), "sample size should be stored with each namespace")
	}

	sampleTestDatabases := []string{"bic_test", "bic_blackbox", "bic_functions_test"}
	databaseNamespaces := map[string][]string{
		sampleTestDatabases[0]: {"eleanor", "bar", "hello"},
		sampleTestDatabases[1]: {"bob", "alice", "joe"},
		sampleTestDatabases[2]: {"eleanor", "joe", "bobby"},
	}

	// ideally, we'd delete all databases in the target mongod cluster
	// but this is undesirable when running the bic locally so we
	// enumerate what collections we know exist, instead.
	cleanupData(session, sampleTestDatabases...)
	for db, collections := range databaseNamespaces {
		for _, collection := range collections {
			dbutils.InsertDocuments(session, db, collection, doc)
		}
	}

	namespaceSelectorTests := []struct {
		description        string
		samplePattern      []string
		expectedNamespaces []string
	}{
		{"inclusion",
			[]string{"*.*"},
			[]string{
				"bic_test.eleanor", "bic_test.hello", "bic_test.bar",
				"bic_blackbox.joe", "bic_blackbox.alice", "bic_blackbox.bob",
				"bic_functions_test.eleanor", "bic_functions_test.bobby", "bic_functions_test.joe",
			},
		},
		{"db_inclusion",
			[]string{"bic_test.*"},
			[]string{"bic_test.eleanor", "bic_test.hello", "bic_test.bar"},
		},
		{"combined_db_inclusion",
			[]string{"bic_test.*", "bic_blackbox.*"},
			[]string{
				"bic_test.eleanor", "bic_test.hello", "bic_test.bar",
				"bic_blackbox.joe", "bic_blackbox.alice", "bic_blackbox.bob",
			},
		},
		{"db_exclusion",
			[]string{"~bic_test.*"},
			[]string{
				"bic_blackbox.joe", "bic_blackbox.alice", "bic_blackbox.bob",
				"bic_functions_test.eleanor", "bic_functions_test.bobby", "bic_functions_test.joe",
			},
		},
		{"combined_db_exclusion",
			[]string{"~bic_test.*", "~bic_blackbox.*"},
			[]string{
				"bic_functions_test.eleanor", "bic_functions_test.bobby", "bic_functions_test.joe",
			},
		},
		{"db_inclusion_and_db_exclusion",
			[]string{"bic_blackbox.*", "~bic_test.*"},
			[]string{
				"bic_blackbox.joe", "bic_blackbox.alice", "bic_blackbox.bob",
			},
		},
		{"combined_db_inclusion_and_db_exclusion",
			[]string{"bic_blackbox.*", "bic_functions_test.*", "~bic_test.*"},
			[]string{
				"bic_blackbox.joe", "bic_blackbox.alice", "bic_blackbox.bob",
				"bic_functions_test.eleanor", "bic_functions_test.bobby", "bic_functions_test.joe",
			},
		},
		{"db_inclusion_and_combined_db_exclusion",
			[]string{"bic_blackbox.*", "~bic_test.*", "~bic_functions_test.*"},
			[]string{
				"bic_blackbox.joe", "bic_blackbox.alice", "bic_blackbox.bob",
			},
		},
		{"collection_inclusion",
			[]string{"*.joe"},
			[]string{"bic_functions_test.joe", "bic_blackbox.joe"},
		},
		{"combined_collection_inclusion",
			[]string{"*.joe", "*.eleanor"},
			[]string{"bic_functions_test.joe", "bic_blackbox.joe", "bic_test.eleanor",
				"bic_functions_test.eleanor",
			},
		},
		{"collection_exclusion",
			[]string{"~*.joe"},
			[]string{
				"bic_test.eleanor", "bic_test.hello", "bic_test.bar",
				"bic_blackbox.alice", "bic_blackbox.bob",
				"bic_functions_test.eleanor", "bic_functions_test.bobby",
			},
		},
		{"combined_collection_exclusion",
			[]string{"~*.joe", "~*.hello", "~*.eleanor"},
			[]string{
				"bic_test.bar", "bic_functions_test.bobby",
				"bic_blackbox.alice", "bic_blackbox.bob",
			},
		},
		{"collection_inclusion_and_collection_exclusion",
			[]string{"*.hello", "~*.joe"},
			[]string{"bic_test.hello"},
		},
		{"combined_collection_inclusion_and_collection_exclusion",
			[]string{"~*.joe", "*.hello", "*.bob*"},
			[]string{"bic_test.hello", "bic_blackbox.bob", "bic_functions_test.bobby"},
		},
		{"collection_inclusion_and_combined_collection_exclusion",
			[]string{"*.hello", "~*.joe", "~*.eleanor"},
			[]string{"bic_test.hello"},
		},
		{"collection_inclusion_and_db_inclusion",
			[]string{"bic_blackbox.joe", "bic_test.*"},
			[]string{
				"bic_test.eleanor", "bic_test.hello", "bic_test.bar", "bic_blackbox.joe",
			},
		},
		{"combined_collection_inclusion_and_db_inclusion",
			[]string{"bic_blackbox.joe", "bic_blackbox.alice", "bic_test.*"},
			[]string{
				"bic_test.eleanor", "bic_test.hello", "bic_test.bar",
				"bic_blackbox.joe", "bic_blackbox.alice",
			},
		},
		{"collection_inclusion_and_combined_db_inclusion",
			[]string{"bic_blackbox.joe", "bic_test.*", "bic_functions_test.*"},
			[]string{
				"bic_test.eleanor", "bic_test.hello", "bic_test.bar", "bic_blackbox.joe",
				"bic_functions_test.eleanor", "bic_functions_test.bobby", "bic_functions_test.joe",
			},
		},
		{"collection_inclusion_and_db_exclusion",
			[]string{"bic_blackbox.joe", "~bic_test.*"},
			[]string{
				"bic_blackbox.joe",
			},
		},
		{"combined_collection_inclusion_and_db_exclusion",
			[]string{"bic_blackbox.joe", "bic_blackbox.alice", "~bic_test.*"},
			[]string{
				"bic_blackbox.joe", "bic_blackbox.alice",
			},
		},
		{"collection_inclusion_and_combined_db_exclusion",
			[]string{"bic_blackbox.joe", "~bic_test.*", "~bic_functions_test.*"},
			[]string{
				"bic_blackbox.joe",
			},
		},
		{"ns_inclusion",
			[]string{"bic_blackbox.joe"},
			[]string{"bic_blackbox.joe"},
		},
		{"combined_ns_inclusion",
			[]string{"bic_blackbox.joe", "bic_functions_test.joe"},
			[]string{"bic_blackbox.joe", "bic_functions_test.joe"},
		},
		{"ns_exclusion",
			[]string{"~bic_blackbox.joe"},
			[]string{
				"bic_test.eleanor", "bic_test.hello", "bic_test.bar",
				"bic_blackbox.alice", "bic_blackbox.bob",
				"bic_functions_test.eleanor", "bic_functions_test.bobby", "bic_functions_test.joe",
			},
		},
		{"combined_ns_exclusion",
			[]string{"~bic_blackbox.joe", "~bic_functions_test.bobby"},
			[]string{
				"bic_test.eleanor", "bic_test.hello", "bic_test.bar",
				"bic_blackbox.alice", "bic_blackbox.bob",
				"bic_functions_test.eleanor", "bic_functions_test.joe",
			},
		},
		{"ns_inclusion_and_ns_exclusion",
			[]string{"bic_test.bar", "~bic_blackbox.joe"},
			[]string{"bic_test.bar"},
		},
		{"combined_ns_inclusion_and_ns_exclusion",
			[]string{"bic_test.bar", "bic_blackbox.alice", "~bic_blackbox.joe"},
			[]string{"bic_test.bar", "bic_blackbox.alice"},
		},
		{"ns_inclusion_and_combined_ns_exclusion",
			[]string{"bic_test.bar", "~bic_blackbox.joe", "~bic_functions_test.joe"},
			[]string{"bic_test.bar"},
		},
		{"ns_inclusion_and_db_inclusion",
			[]string{"bic_blackbox.joe", "bic_test.*"},
			[]string{
				"bic_test.eleanor", "bic_test.hello", "bic_test.bar",
				"bic_blackbox.joe",
			},
		},
		{"combined_ns_inclusion_and_db_inclusion",
			[]string{"bic_blackbox.joe", "bic_functions_test.joe", "bic_test.*"},
			[]string{
				"bic_test.eleanor", "bic_test.hello", "bic_test.bar",
				"bic_blackbox.joe", "bic_functions_test.joe",
			},
		},
		{"ns_inclusion_and_combined_db_inclusion",
			[]string{"bic_blackbox.joe", "bic_test.*", "bic_blackbox.*"},
			[]string{
				"bic_test.eleanor", "bic_test.hello", "bic_test.bar",
				"bic_blackbox.joe", "bic_blackbox.alice", "bic_blackbox.bob",
			},
		},
		{"ns_inclusion_and_db_exclusion",
			[]string{"bic_blackbox.joe", "~bic_test.*"},
			[]string{"bic_blackbox.joe"},
		},
		{"combined_ns_inclusion_and_db_exclusion",
			[]string{"bic_blackbox.joe", "bic_test.bar", "~bic_test.*"},
			[]string{"bic_blackbox.joe"},
		},
		{"ns_inclusion_and_combined_db_exclusion",
			[]string{"bic_blackbox.joe", "~bic_blackbox.*", "~bic_test.*"},
			nil,
		},
		{"ns_inclusion_and_collection_inclusion",
			[]string{"bic_blackbox.joe", "*.eleanor"},
			[]string{"bic_blackbox.joe", "bic_functions_test.eleanor", "bic_test.eleanor"},
		},
		{"combined_ns_inclusion_and_collection_inclusion",
			[]string{"bic_blackbox.joe", "bic_functions_test.joe", "*.eleanor"},
			[]string{
				"bic_blackbox.joe", "bic_test.eleanor",
				"bic_functions_test.eleanor", "bic_functions_test.joe",
			},
		},
		{"ns_inclusion_and_combined_collection_inclusion",
			[]string{"bic_blackbox.joe", "bic_functions_test.joe", "*.eleanor"},
			[]string{
				"bic_blackbox.joe", "bic_test.eleanor",
				"bic_functions_test.joe", "bic_functions_test.eleanor",
			},
		},
		{"ns_inclusion_and_collection_exclusion",
			[]string{"bic_blackbox.joe", "~*.eleanor"},
			[]string{"bic_blackbox.joe"},
		},
		{"combined_ns_inclusion_and_collection_exclusion",
			[]string{"bic_blackbox.joe", "bic_functions_test.joe", "~*.eleanor"},
			[]string{"bic_blackbox.joe", "bic_functions_test.joe"},
		},
		{"ns_inclusion_and_combined_collection_exclusion",
			[]string{"bic_blackbox.joe", "~*.eleanor", "~*.joe"},
			nil,
		},
		{"ns_exclusion_and_db_exclusion",
			[]string{"~bic_blackbox.joe", "~bic_test.*"},
			[]string{
				"bic_blackbox.alice", "bic_blackbox.bob",
				"bic_functions_test.eleanor", "bic_functions_test.bobby", "bic_functions_test.joe",
			},
		},
		{"combined_ns_exclusion_and_db_exclusion",
			[]string{"~bic_blackbox.joe", "~bic_functions_test.bobby", "~bic_test.*"},
			[]string{
				"bic_blackbox.alice", "bic_blackbox.bob",
				"bic_functions_test.eleanor", "bic_functions_test.joe",
			},
		},
		{"ns_exclusion_and_combined_db_exclusion",
			[]string{"~bic_blackbox.joe", "~bic_test.*", "~bic_functions_test.*"},
			[]string{"bic_blackbox.alice", "bic_blackbox.bob"},
		},
		{"ns_exclusion_and_db_inclusion",
			[]string{"~bic_blackbox.joe", "bic_test.*"},
			[]string{"bic_test.eleanor", "bic_test.hello", "bic_test.bar"},
		},
		{"combined_ns_exclusion_and_db_inclusion",
			[]string{"~bic_blackbox.joe", "~bic_functions_test.bobby", "bic_functions_test.*"},
			[]string{"bic_functions_test.eleanor", "bic_functions_test.joe"},
		},
		{"ns_exclusion_and_combined_db_inclusion",
			[]string{"~bic_blackbox.joe", "bic_functions_test.*", "bic_blackbox.*"},
			[]string{
				"bic_blackbox.alice", "bic_blackbox.bob",
				"bic_functions_test.eleanor", "bic_functions_test.bobby", "bic_functions_test.joe",
			},
		},
		{"ns_exclusion_and_collection_inclusion",
			[]string{"~bic_blackbox.joe", "*.bobby"},
			[]string{"bic_functions_test.bobby"},
		},
		{"combined_ns_exclusion_and_collection_inclusion",
			[]string{"~bic_blackbox.joe", "~bic_test.bobby", "*.hello"},
			[]string{"bic_test.hello"},
		},
		{"ns_exclusion_and_combined_collection_inclusion",
			[]string{"~bic_blackbox.joe", "*.bobby", "*.eleanor"},
			[]string{"bic_functions_test.bobby", "bic_functions_test.eleanor", "bic_test.eleanor"},
		},
		{"ns_exclusion_and_collection_exclusion",
			[]string{"~bic_blackbox.joe", "~*.eleanor"},
			[]string{
				"bic_test.hello", "bic_test.bar",
				"bic_blackbox.alice", "bic_blackbox.bob",
				"bic_functions_test.bobby", "bic_functions_test.joe",
			},
		},
		{"combined_ns_exclusion_and_collection_exclusion",
			[]string{"~bic_blackbox.joe", "~bic_test.hello", "~*.eleanor"},
			[]string{
				"bic_test.bar",
				"bic_blackbox.alice", "bic_blackbox.bob",
				"bic_functions_test.bobby", "bic_functions_test.joe",
			},
		},
		{"ns_exclusion_and_combined_collection_exclusion",
			[]string{"~bic_blackbox.joe", "~*.eleanor", "~*.bob*"},
			[]string{
				"bic_test.hello", "bic_test.bar",
				"bic_blackbox.alice", "bic_functions_test.joe",
			},
		},
	}

	nsOpts := config.Default().Schema.Sample
	for _, test := range namespaceSelectorTests {
		t.Run(test.description, func(t *testing.T) {
			req = require.New(t)
			nsOpts.Namespaces = test.samplePattern
			sampleSchema, sampleRecord, err := Schema(&nsOpts, "temp", session, lgr)
			req.Nilf(err, "did not expect error in sampling")
			req.NotNilf(sampleSchema, "did not expect sample schema to be nil")
			req.NotNilf(sampleRecord, "did not expect sample record to be nil")
			req.NotNilf(sampleRecord.Version, "did not expect sample record version to be nil")
			req.Equalf(len(test.expectedNamespaces), len(sampleRecord.Namespaces),
				"sample namespaces not equal")

			for _, expectedNamespace := range test.expectedNamespaces {
				nsComponent := strings.SplitN(expectedNamespace, ".", 2)
				req.Equalf(len(nsComponent), 2, "invalid construction of expected namespace")
				expectedNs := NewNamespace(nsComponent[0], nsComponent[1], sampleRecord.Version.ID)
				req.Emptyf(shouldContainNS(sampleRecord.Namespaces, expectedNs),
					"expected namespaces should be present")
			}
		})
	}
	cleanupData(session, sampleTestDatabases...)
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
	req.NotNilf(sampleSchema, "sample schema is nil")
	dbutils.RunCmd(session, db2, bson.D{{Name: "profile", Value: 0}}, &struct{}{})

	req.NotNilf(sampleRecord, "sample record is nil")
	req.Equal(sampleRecord.Database, cfg.Schema.Sample.Source)
	req.NotNilf(sampleRecord.Version, "sample version is nil")
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
	req.NotNilf(table, "did not find table foo_Xx_1")
	expectedColumnMappings := []sqlColumnMapping{
		{"_id", "_id"}, {"Xx.b", "Xx.b"}, {"Xx_idx", "Xx_idx"},
	}
	req.Empty(ShouldResemble(getColumnMappings(table), expectedColumnMappings))

	table = dbs[0].Table("foo_Xx_0")
	req.NotNilf(table, "did not find table foo_Xx_0")
	expectedColumnMappings = []sqlColumnMapping{
		{"_id", "_id"}, {"hello", "hello"},
	}
	req.Empty(ShouldResemble(getColumnMappings(table), expectedColumnMappings))

	table = dbs[0].Table("foo_xX")
	req.NotNilf(table, "did not find table foo_xX")
	expectedColumnMappings = []sqlColumnMapping{
		{"_id", "_id"}, {"xX.c", "xX.c"}, {"xX_idx", "xX_idx"},
	}
	req.Empty(ShouldResemble(getColumnMappings(table), expectedColumnMappings))

	table = dbs[0].Table("x_0")
	req.NotNilf(table, "did not find table x_0")
	expectedColumnMappings = []sqlColumnMapping{{"_id", "_id"}}
	req.Empty(ShouldResemble(getColumnMappings(table), expectedColumnMappings))

	table = dbs[0].Table("X")
	req.NotNilf(table, "did not find table X")
	req.Empty(ShouldResemble(getColumnMappings(table), expectedColumnMappings))

	table = dbs[0].Table("foo")
	req.NotNilf(table, "did not find table foo")
	expectedColumnMappings = []sqlColumnMapping{
		{"_id", "_id"}, {"XX", "XX"}, {"xX_0", "xX_0"},
	}
	req.Empty(ShouldResemble(getColumnMappings(table), expectedColumnMappings))
}

func cleanupData(session *mongodb.Session, databases ...string) {
	dbutils.DropDatabase(session, cfg.Schema.Sample.Source)
	dbutils.DropDatabase(session, db1)
	dbutils.DropDatabase(session, db2)
	dbutils.DropDatabase(session, db3)
	for _, db := range databases {
		dbutils.DropDatabase(session, db)
	}
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
