package mongodrdl_test

import (
	"fmt"
	yaml "github.com/10gen/candiedyaml"
	"github.com/10gen/sqlproxy"
	"github.com/10gen/sqlproxy/mongodrdl"
	"github.com/mongodb/mongo-tools/common/db"
	"github.com/mongodb/mongo-tools/common/options"
	"github.com/mongodb/mongo-tools/common/testutil"
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
	DatabaseName = "schema_test"
	SSLTestKey   = "SQLPROXY_SSLTEST"
)

func getSslOpts() *options.SSL {
	var sslOpts *options.SSL

	if len(os.Getenv(SSLTestKey)) > 0 {
		db.GetConnectorFuncs = []db.GetConnectorFunc{
			func(opts options.ToolOptions) db.DBConnector {
				if opts.SSL.UseSSL {
					return &sqlproxy.SSLDBConnector{}
				}
				return nil
			},
		}
		return &options.SSL{
			UseSSL:              true,
			SSLPEMKeyFile:       "../testdata/client.pem",
			SSLAllowInvalidCert: true,
		}
	}

	return sslOpts
}

func TestConfiguration(t *testing.T) {

	sslOptions := getSslOpts()
	testutil.VerifyTestType(t, testutil.UnitTestType)

	Convey("Verify configuration options", t, func() {
		gen := &mongodrdl.SchemaGenerator{
			ToolOptions: &options.ToolOptions{
				Namespace: &options.Namespace{
					DB:         "testdb",
					Collection: "mongoddl",
				},
				SSL: sslOptions,
			},
			OutputOptions: &mongodrdl.OutputOptions{
				Out: "out/testdb.yml",
			},
			SampleOptions: &mongodrdl.SampleOptions{SampleSize: 1000},
		}

		gen.Init()

		Convey("output should be testdb.yaml", func() {
			So(gen.OutputOptions.Out, ShouldEqual, "out/testdb.yml")
		})

		Convey("DB should be mongoddl", func() {
			So(gen.ToolOptions.Namespace.DB, ShouldEqual, "testdb")
		})

		Convey("Collection should be mongoddl", func() {
			So(gen.ToolOptions.Namespace.Collection, ShouldEqual, "mongoddl")
		})

	})

}

func TestRoundtrips(t *testing.T) {

	sslOptions := getSslOpts()

	Convey("Collection filtering", t, func() {
		Convey("Should ignore system.* collections", func() {
			gen := &mongodrdl.SchemaGenerator{
				ToolOptions: &options.ToolOptions{
					Namespace: &options.Namespace{
						DB: "indexed",
					},
					SSL: sslOptions,
				},
				OutputOptions: &mongodrdl.OutputOptions{
					Out: "out/indexed.yml",
				},
				SampleOptions: &mongodrdl.SampleOptions{SampleSize: 1000},
			}

			gen.Init()

			session, err := gen.Connect()
			So(err, ShouldBeNil)
			defer session.Close()

			db := session.DB(gen.ToolOptions.Namespace.DB)
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

		Convey("Should ignore system.* collections in admin", func() {
			gen := &mongodrdl.SchemaGenerator{
				ToolOptions: &options.ToolOptions{
					Namespace: &options.Namespace{
						DB: "admin",
					},
					SSL: sslOptions,
				},
				OutputOptions: &mongodrdl.OutputOptions{
					Out: "out/admin.yml",
				},
				SampleOptions: &mongodrdl.SampleOptions{SampleSize: 1000},
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
				ToolOptions: &options.ToolOptions{
					Namespace: &options.Namespace{
						DB:         DatabaseName,
						Collection: "complete_schema",
					},
					SSL: sslOptions,
				},
				OutputOptions: &mongodrdl.OutputOptions{
					Out:               "out/complete_schema_synthetic.yml",
					CustomFilterField: "__MONGOQUERY",
				},
				SampleOptions: &mongodrdl.SampleOptions{SampleSize: 1000},
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

	opts := options.New("mongoimport", mongoimport.Usage,
		options.EnabledOptions{Auth: true, Connection: true, Namespace: true})

	opts.Namespace = &options.Namespace{
		DB:         schema.ToolOptions.DB,
		Collection: schema.ToolOptions.Collection,
	}
	opts.SSL = getSslOpts()
	opts.Quiet = true
	opts.SetVerbosity("0")

	sessionProvider, err := db.NewSessionProvider(*opts)
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
