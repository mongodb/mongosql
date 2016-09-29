package mongodrdl_test

import (
	"fmt"
	yaml "github.com/10gen/candiedyaml"
	"github.com/10gen/sqlproxy/mongodrdl"
	"github.com/10gen/sqlproxy/options"
	"github.com/10gen/sqlproxy/testutils"
	toolsdb "github.com/mongodb/mongo-tools/common/db"
	toolsoptions "github.com/mongodb/mongo-tools/common/options"
	"github.com/mongodb/mongo-tools/mongoimport"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

const (
	DatabaseName = "test"
	SSLTestKey   = "SQLPROXY_SSLTEST"
)

func getSslOpts() *options.DrdlSSL {
	sslOpts := &options.DrdlSSL{}

	if len(os.Getenv(SSLTestKey)) > 0 {
		return testutils.GetDrdlSSLOpts()
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
				DrdlSSL: sslOptions,
			},
			OutputOptions: &options.DrdlOutput{
				Out: "out/testdb.yml",
			},
			SampleOptions: &options.DrdlSample{SampleSize: 1000},
		}

		gen.Init()

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
					DrdlSSL: sslOptions,
				},
				OutputOptions: &options.DrdlOutput{
					Out: "out/indexed.yml",
				},
				SampleOptions: &options.DrdlSample{SampleSize: 1000},
			}

			gen.Init()

			session, err := gen.Connect()
			So(err, ShouldBeNil)
			defer session.Close()

			db := session.DB(gen.ToolOptions.DrdlNamespace.DB)
			err = db.Run(bson.D{{"profile", 10}, {"slowms", 0}}, bson.M{})
			So(err, ShouldBeNil)
			defer db.DropDatabase()

			collection := db.C("test")
			err = collection.Insert(bson.M{
				"first":  "Who",
				"second": "What",
			})
			So(err, ShouldBeNil)

			collection.Find(bson.M{})

			collection.EnsureIndexKey("first", "second")
			indexes, err := collection.Indexes()
			So(err, ShouldBeNil)
			So(indexes, ShouldNotBeNil)

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
					DrdlSSL: sslOptions,
				},
				OutputOptions: &options.DrdlOutput{
					Out: "out/views-expected.yml",
				},
				SampleOptions: &options.DrdlSample{SampleSize: 1000},
			}

			gen.Init()

			session, err := gen.Connect()
			So(err, ShouldBeNil)
			db := session.DB(gen.ToolOptions.DrdlNamespace.DB)
			defer func() {
				db.DropDatabase()
				session.Close()
			}()

			So(db.C("base").Insert(
				bson.M{"a": 1, "b": 123},
				bson.M{"a": 2, "b": 134},
				bson.M{"a": 3, "b": "s"},
			), ShouldBeNil)

			So(db.Run(bson.D{
				{"create", "view"},
				{"viewOn", "base"},
				{"pipeline", []bson.M{{"$match": bson.M{"a": 3}}}},
			}, &struct{}{}), ShouldBeNil)

			// for views, get indexes should return an error
			_, err = db.C("view").Indexes()
			So(err, ShouldNotBeNil)

			gen.Generate()

			output, err := readYaml(gen.OutputOptions.Out)
			expected, err := readYaml("testdata/view-expected.yml")
			So(err, ShouldBeNil)
			So(toString(output), ShouldEqual, toString(expected))
		})

		Convey("Should work with mongodb views containing geo index", func() {
			gen := &mongodrdl.SchemaGenerator{
				ToolOptions: &options.DrdlOptions{
					DrdlNamespace: &options.DrdlNamespace{
						DB: "viewDB",
					},
					DrdlSSL: sslOptions,
				},
				OutputOptions: &options.DrdlOutput{
					Out: "out/views-geo-expected.yml",
				},
				SampleOptions: &options.DrdlSample{SampleSize: 1000},
			}

			gen.Init()

			session, err := gen.Connect()
			So(err, ShouldBeNil)
			defer session.Close()

			db := session.DB(gen.ToolOptions.DrdlNamespace.DB)
			defer db.DropDatabase()

			base := db.C("base")
			So(base.Insert(bson.M{"loc": []bson.M{
				bson.M{"type": "Point"},
				bson.M{"coordinates": []interface{}{-73.88, 40.78}}}},
			), ShouldBeNil)

			idx := mgo.Index{
				Key:  []string{"$2d:loc.coordinates"},
				Bits: 26,
			}
			So(base.EnsureIndex(idx), ShouldBeNil)
			indexes, err := base.Indexes()
			So(err, ShouldBeNil)
			So(indexes, ShouldNotBeNil)

			So(db.Run(bson.D{
				{"create", "view"},
				{"viewOn", "base"},
				{"pipeline", []bson.M{}},
			}, &struct{}{}), ShouldBeNil)

			// for views, get indexes should return an error
			_, err = db.C("view").Indexes()
			So(err, ShouldNotBeNil)

			gen.Generate()

			output, err := readYaml(gen.OutputOptions.Out)
			expected, err := readYaml("testdata/view-geo-expected.yml")
			So(err, ShouldBeNil)
			So(toString(output), ShouldEqual, toString(expected))
		})

		Convey("Should ignore system.* collections in admin", func() {
			gen := &mongodrdl.SchemaGenerator{
				ToolOptions: &options.DrdlOptions{
					DrdlNamespace: &options.DrdlNamespace{
						DB: "admin",
					},
					DrdlSSL: sslOptions,
				},
				OutputOptions: &options.DrdlOutput{
					Out: "out/admin.yml",
				},
				SampleOptions: &options.DrdlSample{SampleSize: 1000},
			}
			gen.Init()

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
				testJson(f)
			})
		}

		Convey("Test Synthetic Query Field", func() {
			gen := &mongodrdl.SchemaGenerator{
				ToolOptions: &options.DrdlOptions{
					DrdlNamespace: &options.DrdlNamespace{
						DB:         DatabaseName,
						Collection: "complete_schema",
					},
					DrdlSSL: sslOptions,
				},
				OutputOptions: &options.DrdlOutput{
					Out:               "out/complete_schema_synthetic.yml",
					CustomFilterField: "__MONGOQUERY",
				},
				SampleOptions: &options.DrdlSample{SampleSize: 1000},
			}

			compareYaml(gen, "complete_schema", "complete_schema_synthetic-expected")
		})
	})
}

func testJson(collection string) {
	gen := mongodrdl.NewSchemaGenerator(DatabaseName, collection, fmt.Sprintf("out/%s.yml", collection), getSslOpts())
	compareYaml(gen, collection, collection+"-expected")
}

func compareYaml(gen *mongodrdl.SchemaGenerator, collection string, expectedName string) {
	gen.Init()

	session, err := gen.Connect()
	So(err, ShouldBeNil)

	defer session.Close()
	db := session.DB(DatabaseName)
	So(db, ShouldNotBeNil)
	coll := db.C(collection)
	So(coll, ShouldNotBeNil)
	err = coll.DropCollection()

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

func importJson(schema *mongodrdl.SchemaGenerator, dbName string, collName string, fileName string, indexes ...mgo.Index) {
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
			SSLPEMKeyFile:       "../testdata/client.pem",
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

	coll := session.DB(dbName).C(collName)
	for _, index := range indexes {
		err = coll.EnsureIndex(index)
		So(err, ShouldBeNil)
	}
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
