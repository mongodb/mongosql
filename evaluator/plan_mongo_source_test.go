package evaluator_test

import (
	"context"
	"os"
	"testing"

	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/testutils/dbutils"
	mongoutil "github.com/10gen/sqlproxy/internal/testutils/mongodb"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/variable"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/schema"
)

type connCtx struct {
	catalog   *catalog.Catalog
	server    evaluator.ServerCtx
	session   *mongodb.Session
	variables *variable.Container
}

func (*connCtx) LastInsertId() int64 {
	return int64(0)
}

func (*connCtx) RowCount() int64 {
	return int64(0)
}

func (*connCtx) ConnectionID() uint32 {
	return uint32(0)
}

func (c *connCtx) Context() context.Context {
	return context.Background()
}

func (*connCtx) DB() string {
	return ""
}

func (*connCtx) Kill(id uint32, scope evaluator.KillScope) error {
	return nil
}

func (*connCtx) Logger(_ ...string) *log.Logger {
	lg := log.GlobalLogger()
	return lg
}

func (c *connCtx) Server() evaluator.ServerCtx {
	return c.server
}

func (c *connCtx) Session() *mongodb.Session {
	return c.session
}

func (*connCtx) User() string {
	return ""
}

func (c *connCtx) Catalog() *catalog.Catalog {
	return c.catalog
}

func (c *connCtx) UpdateCatalog(*schema.Schema) error {
	return nil
}

func (c *connCtx) GetStartupInfo() []string {
	return []string{}
}

func (c *connCtx) Variables() *variable.Container {
	return c.variables
}

func (c *connCtx) VersionAtLeast(version ...uint8) bool {
	return c.Variables().MongoDBInfo.VersionAtLeast(version...)
}

func getConfig(t *testing.T) *config.Config {
	cfg := config.Default()

	// ssl is turned on
	if len(os.Getenv(mongoutil.SSLTestKey)) > 0 {
		t.Logf("Testing with SSL turned on.")
		cfg.MongoDB.Net.SSL.Enabled = true
		cfg.MongoDB.Net.SSL.AllowInvalidCertificates = true
		cfg.MongoDB.Net.SSL.PEMKeyFile = "../testdata/resources/x509gen/client.pem"
	}
	return cfg
}

func TestMongoSourcePlanStage(t *testing.T) {
	env := setupEnv(t)
	cfgOne := env.cfgOne
	infoOne := evaluator.GetMongoDBInfo(nil, cfgOne, mongodb.AllPrivileges)
	variablesOne := evaluator.CreateTestVariables(infoOne)
	catalogOne := evaluator.GetCatalogFromSchema(cfgOne, variablesOne)
	cfg := getConfig(t)
	sessionProvider, err := mongodb.NewSqldSessionProvider(cfg)
	if err != nil {
		t.Fatalf("failed to set up session provider to test server: %v", err)
		return
	}

	session, err := sessionProvider.Session(context.Background())
	if err != nil {
		t.Fatalf("failed to set up session to test server: %v", err)
		return
	}
	defer session.Close()

	Convey("With a simple test configuration...", t, func() {

		Convey("fetching data from a table scan should return correct results in the right order",
			func() {

				rows := []bson.D{
					{
						bson.DocElem{Name: "_id", Value: "5"},
						bson.DocElem{Name: "a", Value: 6},
						bson.DocElem{Name: "b", Value: 7},
						bson.DocElem{Name: "d", Value: 8},
					},
					{
						bson.DocElem{Name: "_id", Value: "15"},
						bson.DocElem{Name: "a", Value: 16},
						bson.DocElem{Name: "b", Value: 17},
						bson.DocElem{Name: "d", Value: 18},
					},
				}

				var expected []evaluator.Values
				for _, document := range rows {
					values, err := bsonDToValues(1, dbOne, tableTwoName, document)
					So(err, ShouldBeNil)
					expected = append(expected, values)
				}

				dbutils.DropCollection(session, dbOne, tableTwoName)
				dbutils.InsertDocuments(session, dbOne, tableTwoName, rows)

				cCtx := &connCtx{
					catalog:   catalogOne,
					session:   session,
					variables: variablesOne,
				}

				ctx := &evaluator.ExecutionCtx{
					ConnectionCtx: cCtx,
				}

				db, err := catalogOne.Database(dbOne)
				if err != nil {
					panic("database doesn't exist")
				}
				table, err := db.Table(tableTwoName)
				if err != nil {
					panic("table doesn't exist")
				}

				plan := evaluator.NewMongoSourceStage(db, table.(*catalog.MongoTable), 1, "")
				So(err, ShouldBeNil)
				iter, err := plan.Open(ctx)
				So(err, ShouldBeNil)

				row := &evaluator.Row{}

				i := 0

				for iter.Next(row) {
					So(len(row.Data), ShouldEqual, len(expected[i]))
					So(row.Data, ShouldResemble, expected[i])
					row = &evaluator.Row{}
					i++
				}

				So(iter.Close(), ShouldBeNil)
				So(iter.Err(), ShouldBeNil)
			})
	})
}

func TestExtractField(t *testing.T) {
	Convey("With a test bson.D", t, func() {
		testD := bson.D{
			{Name: "a", Value: "string"},
			{Name: "b", Value: []interface{}{"inner", bson.D{{Name: "inner2", Value: 1}}}},
			{Name: "c", Value: bson.D{{Name: "x", Value: 5}}},
			{Name: "d", Value: bson.D{{Name: "z", Value: nil}}},
		}

		Convey("regular fields should be extracted by name", func() {
			val, ok := evaluator.ExtractFieldByName("a", testD)
			So(val, ShouldEqual, "string")
			So(ok, ShouldBeTrue)
		})

		Convey("array fields should be extracted by name", func() {
			val, ok := evaluator.ExtractFieldByName("b.1", testD)
			So(val, ShouldResemble, bson.D{{Name: "inner2", Value: 1}})
			So(ok, ShouldBeTrue)
			val, ok = evaluator.ExtractFieldByName("b.1.inner2", testD)
			So(val, ShouldEqual, 1)
			So(ok, ShouldBeTrue)
			val, ok = evaluator.ExtractFieldByName("b.0", testD)
			So(val, ShouldEqual, "inner")
			So(ok, ShouldBeTrue)
		})

		Convey("subdocument fields should be extracted by name", func() {
			val, ok := evaluator.ExtractFieldByName("c", testD)
			So(val, ShouldResemble, bson.D{{Name: "x", Value: 5}})
			So(ok, ShouldBeTrue)
			val, ok = evaluator.ExtractFieldByName("c.x", testD)
			So(val, ShouldEqual, 5)
			So(ok, ShouldBeTrue)

			Convey("even if they contain null values", func() {
				val, ok := evaluator.ExtractFieldByName("d", testD)
				So(val, ShouldResemble, bson.D{{Name: "z", Value: nil}})
				So(ok, ShouldBeTrue)
				val, ok = evaluator.ExtractFieldByName("d.z", testD)
				So(val, ShouldEqual, nil)
				So(ok, ShouldBeTrue)
				val, ok = evaluator.ExtractFieldByName("d.z.nope", testD)
				So(val, ShouldEqual, nil)
				So(ok, ShouldBeFalse)
			})
		})

		Convey(`non-existing fields should return (nil,false)`, func() {
			for _, c := range []string{"f", "c.nope", "c.nope.NOPE", "b.1000", "b.1.nada"} {
				val, ok := evaluator.ExtractFieldByName(c, testD)
				So(val, ShouldBeNil)
				So(ok, ShouldBeFalse)
			}
		})

	})

	Convey(`Extraction of a non-document should return (nil, false)`, t, func() {
		val, ok := evaluator.ExtractFieldByName("meh", []interface{}{"meh"})
		So(val, ShouldBeNil)
		So(ok, ShouldBeFalse)
	})

	Convey(`Extraction of a nil document should return (nil, false)`, t, func() {
		val, ok := evaluator.ExtractFieldByName("a", nil)
		So(val, ShouldEqual, nil)
		So(ok, ShouldBeFalse)
	})
}
