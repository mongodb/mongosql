package mongodrdl_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	yaml "github.com/10gen/candiedyaml"
	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/testutils"
	"github.com/10gen/sqlproxy/internal/testutils/dbutils"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/mongodrdl"
	"github.com/10gen/sqlproxy/mongodrdl/relational"
	"github.com/10gen/sqlproxy/options"

	toolsdb "github.com/mongodb/mongo-tools/common/db"
	toolsoptions "github.com/mongodb/mongo-tools/common/options"
	"github.com/mongodb/mongo-tools/mongoimport"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	DatabaseName = "test"
	SSLTestKey   = "SQLPROXY_SSLTEST"
	host         = "mongodb://localhost:27017"
)

var (
	logger = log.NewComponentLogger(
		"MONGODRDL", log.GlobalLogger(),
	)
)

func getSslOpts() *options.DrdlSSL {
	sslOpts := &options.DrdlSSL{}

	if len(os.Getenv(SSLTestKey)) > 0 {
		return testutils.DrdlTestSSLOpts()
	}

	return sslOpts
}

func TestConfiguration(t *testing.T) {

	sslOptions := getSslOpts()

	Convey("Verify configuration options", t, func() {
		gen := &mongodrdl.SchemaGenerator{
			ToolOptions: &options.DrdlOptions{
				DrdlNamespace: &options.DrdlNamespace{
					DB:         "testdb",
					Collection: "mongoddl",
				},
				DrdlConnection: &options.DrdlConnection{
					Host: host,
				},
				DrdlSSL: sslOptions,
			},
			OutputOptions: &options.DrdlOutput{
				Out:       "out/testdb.yml",
				PreJoined: true,
			},
			SampleOptions: &options.DrdlSample{Size: 1000},
			Logger:        logger,
		}

		So(gen.Init(), ShouldBeNil)

		Convey("output should be testdb.yaml", func() {
			So(gen.OutputOptions.Out, ShouldEqual, "out/testdb.yml")
		})

		Convey("DB should be mongoddl", func() {
			So(gen.ToolOptions.DrdlNamespace.DB, ShouldEqual, "testdb")
		})

		Convey("Collection should be mongoddl", func() {
			So(gen.ToolOptions.DrdlNamespace.Collection, ShouldEqual, "mongoddl")
		})

	})

}

func TestRoundtrips(t *testing.T) {

	sslOptions := getSslOpts()

	Convey("Collection filtering", t, func() {
		Convey("Should ignore system.* collections", func() {
			gen := &mongodrdl.SchemaGenerator{
				ToolOptions: &options.DrdlOptions{
					DrdlNamespace: &options.DrdlNamespace{
						DB: "indexed",
					},
					DrdlConnection: &options.DrdlConnection{
						Host: host,
					},
					DrdlSSL: sslOptions,
				},
				OutputOptions: &options.DrdlOutput{
					Out:       "out/indexed.yml",
					PreJoined: true,
				},
				SampleOptions: &options.DrdlSample{Size: 1000},
				Logger:        logger,
			}

			So(gen.Init(), ShouldBeNil)

			session, err := gen.Connect()
			So(err, ShouldBeNil)
			defer session.Close()
			db := gen.ToolOptions.DrdlNamespace.DB
			defer dbutils.DropDatabase(session, db)
			dbutils.DropDatabase(session, db)
			documents := []bson.M{
				bson.M{
					"first":  "Who",
					"second": "What",
				},
			}

			dbutils.InsertDocuments(session, db, "test", documents)
			dbutils.CreateIndex(session, db, "test", []string{"first", "second"})

			iter, err := session.ListIndexes(db, "test")
			So(err, ShouldBeNil)

			indexes, index := []mongodb.Index{}, mongodb.Index{}
			ctx := context.Background()
			for iter.Next(ctx, &index) {
				indexes = append(indexes, index)
			}
			So(iter.Close(ctx), ShouldBeNil)
			So(len(indexes), ShouldEqual, 2)

			gen.Generate()
			output, err := readYaml(gen.OutputOptions.Out)
			expected, err := readYaml("testdata/indexed-expected.yml")
			So(err, ShouldBeNil)
			So(toString(output), ShouldEqual, toString(expected))
		})

		Convey("Should work with mongodb views containing no geo index", func() {
			gen := &mongodrdl.SchemaGenerator{
				ToolOptions: &options.DrdlOptions{
					DrdlNamespace: &options.DrdlNamespace{
						DB: "viewDB",
					},
					DrdlConnection: &options.DrdlConnection{
						Host: host,
					},
					DrdlSSL: sslOptions,
				},
				OutputOptions: &options.DrdlOutput{
					Out:       "out/views-expected.yml",
					PreJoined: true,
				},
				SampleOptions: &options.DrdlSample{Size: 1000},
				Logger:        logger,
			}

			So(gen.Init(), ShouldBeNil)

			session, err := gen.Connect()
			So(err, ShouldBeNil)
			defer session.Close()
			db := gen.ToolOptions.DrdlNamespace.DB
			defer func() {
				dbutils.DropDatabase(session, db)
				session.Close()
			}()
			documents := []bson.M{
				bson.M{"a": 1, "b": 123},
				bson.M{"a": 2, "b": 134},
				bson.M{"a": 3, "b": "s"},
			}
			dbutils.InsertDocuments(session, db, "base", documents)

			version := session.Model().Version
			if version.AtLeast(3, 3, 0) {
				So(session.Run(db, bson.D{
					{"create", "view"},
					{"viewOn", "base"},
					{"pipeline", []bson.M{{"$match": bson.M{"a": 3}}}},
				}, &struct{}{}), ShouldBeNil)

				// for views, get indexes should return an error
				_, err := session.ListIndexes(db, "view")
				So(err, ShouldNotBeNil)

				gen.Generate()

				output, err := readYaml(gen.OutputOptions.Out)
				expected, err := readYaml("testdata/view-expected.yml")
				So(err, ShouldBeNil)
				So(toString(output), ShouldEqual, toString(expected))
			}
		})

		Convey("Should work with mongodb views containing geo index", func() {
			gen := &mongodrdl.SchemaGenerator{
				ToolOptions: &options.DrdlOptions{
					DrdlNamespace: &options.DrdlNamespace{
						DB: "viewDB",
					},
					DrdlConnection: &options.DrdlConnection{
						Host: host,
					},
					DrdlSSL: sslOptions,
				},
				OutputOptions: &options.DrdlOutput{
					Out:       "out/views-geo-expected.yml",
					PreJoined: true,
				},
				SampleOptions: &options.DrdlSample{Size: 1000},
				Logger:        logger,
			}

			So(gen.Init(), ShouldBeNil)

			db := gen.ToolOptions.DrdlNamespace.DB
			session, err := gen.Connect()
			So(err, ShouldBeNil)
			defer session.Close()
			defer dbutils.DropDatabase(session, db)
			documents := []bson.M{
				bson.M{
					"loc": []bson.M{
						bson.M{"type": "Point"},
						bson.M{"coordinates": []interface{}{-73.88, 40.78}},
					},
				},
			}
			dbutils.InsertDocuments(session, db, "base", documents)
			dbutils.CreateIndex(session, db, "base", []string{"$2d:loc.coordinates"})

			iter, err := session.ListIndexes(db, "base")
			So(err, ShouldBeNil)
			ctx := context.Background()
			So(iter.Next(ctx, nil), ShouldNotBeNil)

			version := session.Model().Version
			if version.AtLeast(3, 3, 0) {
				So(session.Run(db, bson.D{
					{"create", "view"},
					{"viewOn", "base"},
					{"pipeline", []bson.M{}},
				}, &struct{}{}), ShouldBeNil)

				// for views, get indexes should return an error
				_, err = session.ListIndexes(db, "view")
				So(err, ShouldNotBeNil)

				gen.Generate()

				output, err := readYaml(gen.OutputOptions.Out)
				expected, err := readYaml("testdata/view-geo-expected.yml")
				So(err, ShouldBeNil)
				So(toString(output), ShouldEqual, toString(expected))
			}
		})

		Convey("Should ignore system.* collections in admin", func() {
			gen := &mongodrdl.SchemaGenerator{
				ToolOptions: &options.DrdlOptions{
					DrdlNamespace: &options.DrdlNamespace{
						DB: "admin",
					},
					DrdlConnection: &options.DrdlConnection{
						Host: host,
					},
					DrdlSSL: sslOptions,
				},
				OutputOptions: &options.DrdlOutput{
					Out:       "out/admin.yml",
					PreJoined: true,
				},
				SampleOptions: &options.DrdlSample{Size: 1000},
				Logger:        logger,
			}

			So(gen.Init(), ShouldBeNil)

			session, err := gen.Connect()
			So(err, ShouldBeNil)
			defer session.Close()

			gen.Generate()
			output, err := readYaml(gen.OutputOptions.Out)
			expected, err := readYaml("testdata/admin-expected.yml")
			So(err, ShouldBeNil)
			So(toString(output), ShouldEqual, toString(expected))
		})
	})

	Convey("Roundtrip testing", t, func() {

		files := []string{
			"arrays",
			"arraysDuplicateNamesDifferentPaths",
			"complete_schema",
			"nestedArrays",
			"nestedArraysDocs",
			"roundtrip",
			"sub_documents",
		}

		for _, f := range files {
			Convey(fmt.Sprintf("With %s", f), func() {
				testJson(f, false)
			})
			Convey(fmt.Sprintf("With %s prejoined", f), func() {
				testJson(f, true)
			})
		}

		Convey("Test Synthetic Query Field", func() {
			gen := &mongodrdl.SchemaGenerator{
				ToolOptions: &options.DrdlOptions{
					DrdlNamespace: &options.DrdlNamespace{
						DB:         DatabaseName,
						Collection: "complete_schema",
					},
					DrdlConnection: &options.DrdlConnection{
						Host: host,
					},
					DrdlSSL: sslOptions,
				},
				OutputOptions: &options.DrdlOutput{
					Out:               "out/complete_schema_synthetic.yml",
					CustomFilterField: "__MONGOQUERY",
					PreJoined:         true,
				},
				SampleOptions: &options.DrdlSample{Size: 1000},
				Logger:        logger,
			}

			compareYaml(gen, "complete_schema", "complete_schema_synthetic-expected")
		})
	})
}

func testJson(collection string, prejoined bool) {
	gen := newSchemaGenerator(DatabaseName, collection, fmt.Sprintf("out/%s.yml", collection), getSslOpts())
	gen.OutputOptions.PreJoined = prejoined
	name := collection + "-expected"
	if prejoined {
		name += "-prejoined"
	}
	compareYaml(gen, collection, name)
}

func compareYaml(gen *mongodrdl.SchemaGenerator, collection string, expectedName string) {

	So(gen.Init(), ShouldBeNil)

	session, err := gen.Connect()
	So(err, ShouldBeNil)
	defer session.Close()

	dbutils.DropCollection(session, DatabaseName, collection)

	importJson(gen, DatabaseName, collection, fmt.Sprintf("testdata/%s.json", collection))

	schema, err := gen.Generate()
	So(err, ShouldBeNil)
	So(schema, ShouldNotBeNil)

	output, err := readYaml(gen.OutputOptions.Out)
	So(err, ShouldBeNil)

	expected, err := readYaml(fmt.Sprintf("testdata/%s.yml", expectedName))
	So(err, ShouldBeNil)

	So(toString(output), ShouldEqual, toString(expected))
}

func importJson(schema *mongodrdl.SchemaGenerator, dbName, collName,
	fileName string, indexes ...mongodb.Index) {
	session, err := schema.Connect()
	So(err, ShouldBeNil)
	defer session.Close()

	opts := toolsoptions.New("mongoimport", mongoimport.Usage,
		toolsoptions.EnabledOptions{Auth: true, Connection: true, Namespace: true})

	opts.Namespace = &toolsoptions.Namespace{
		DB:         schema.ToolOptions.DB,
		Collection: schema.ToolOptions.Collection,
	}

	sslOpts := getSslOpts()
	opts.SSL = &toolsoptions.SSL{}
	if sslOpts.UseSSL {
		opts.SSL = &toolsoptions.SSL{
			UseSSL:              true,
			SSLPEMKeyFile:       "../testdata/resources/x509gen/client.pem",
			SSLAllowInvalidCert: true,
		}
	}

	opts.Quiet = true
	opts.SetVerbosity("0")

	sessionProvider, err := toolsdb.NewSessionProvider(*opts)
	So(err, ShouldBeNil)

	m := &mongoimport.MongoImport{
		ToolOptions: opts,
		InputOptions: &mongoimport.InputOptions{
			File: fileName,
		},
		IngestOptions: &mongoimport.IngestOptions{
			Drop:        true,
			StopOnError: false,
		},
		SessionProvider: sessionProvider,
	}

	_, err = m.ImportDocuments()
	So(err, ShouldBeNil)
	time.Sleep(1 * time.Second)

	for _, index := range indexes {
		dbutils.CreateIndex(session, dbName, collName, relational.SimpleIndexKey(index.Key))
	}
}

func newSchemaGenerator(db, collection, outputFile string, sslOptions *options.DrdlSSL) *mongodrdl.SchemaGenerator {
	gen := &mongodrdl.SchemaGenerator{
		ToolOptions: &options.DrdlOptions{
			DrdlNamespace: &options.DrdlNamespace{
				DB:         db,
				Collection: collection,
			},
			DrdlConnection: &options.DrdlConnection{
				Host: host,
			},
			DrdlSSL: sslOptions,
		},
		OutputOptions: &options.DrdlOutput{
			Out: outputFile,
		},
		SampleOptions: &options.DrdlSample{Size: 1000},
		Logger:        logger,
	}

	if err := gen.Init(); err != nil {
		panic(err)
	}

	return gen
}

func readYaml(file string) (*mongodrdl.Schema, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	schema := &mongodrdl.Schema{}
	return schema, yaml.Unmarshal(bytes, schema)
}

func toString(s *mongodrdl.Schema) string {
	bytes, err := yaml.Marshal(s)
	if err != nil {
		panic(err.Error())
	}
	return string(bytes)
}
