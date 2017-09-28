package evaluator

import (
	"context"
	"os"
	"testing"

	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/testutils/dbutils"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/variable"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/schema"
)

const (
	testMongoHost = "127.0.0.1"
	testMongoPort = "27017"
)

type connCtx struct {
	catalog   *catalog.Catalog
	server    ServerCtx
	session   *mongodb.Session
	variables *variable.Container
}

func (_ *connCtx) LastInsertId() int64 {
	return int64(0)
}

func (_ *connCtx) RowCount() int64 {
	return int64(0)
}

func (_ *connCtx) ConnectionId() uint32 {
	return uint32(0)
}

func (c *connCtx) Context() context.Context {
	return context.Background()
}

func (_ *connCtx) DB() string {
	return ""
}

func (_ *connCtx) Kill(id uint32, scope KillScope) error {
	return nil
}

func (_ *connCtx) Logger(_ string) *log.Logger {
	lg := log.GlobalLogger()
	return &lg
}

func (c *connCtx) Server() ServerCtx {
	return c.server
}

func (c *connCtx) Session() *mongodb.Session {
	return c.session
}

func (_ *connCtx) User() string {
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

func getConfig(t *testing.T) *config.Config {
	cfg := config.Default()

	// ssl is turned on
	if len(os.Getenv(SSLTestKey)) > 0 {
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
	infoOne := getMongoDBInfo(nil, cfgOne, mongodb.AllPrivileges)
	variablesOne := createTestVariables(infoOne)
	catalogOne := getCatalogFromSchema(cfgOne, variablesOne)
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

		Convey("fetching data from a table scan should return correct results in the right order", func() {

			rows := []bson.D{
				bson.D{
					bson.DocElem{Name: "a", Value: 6},
					bson.DocElem{Name: "b", Value: 7},
					bson.DocElem{Name: "d", Value: 8},
					bson.DocElem{Name: "_id", Value: "5"},
				},
				bson.D{
					bson.DocElem{Name: "a", Value: 16},
					bson.DocElem{Name: "b", Value: 17},
					bson.DocElem{Name: "d", Value: 18},
					bson.DocElem{Name: "_id", Value: "15"},
				},
			}

			var expected []Values
			for _, document := range rows {
				values, err := bsonDToValues(1, tableTwoName, document)
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

			ctx := &ExecutionCtx{
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

			plan := NewMongoSourceStage(db, table.(*catalog.MongoTable), 1, "")
			So(err, ShouldBeNil)
			iter, err := plan.Open(ctx)
			So(err, ShouldBeNil)

			row := &Row{}

			i := 0

			for iter.Next(row) {
				So(len(row.Data), ShouldEqual, len(expected[i]))
				So(row.Data, ShouldResemble, expected[i])
				row = &Row{}
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
			{"a", "string"},
			{"b", []interface{}{"inner", bson.D{{"inner2", 1}}}},
			{"c", bson.D{{"x", 5}}},
			{"d", bson.D{{"z", nil}}},
		}

		Convey("regular fields should be extracted by name", func() {
			val, ok := extractFieldByName("a", testD)
			So(val, ShouldEqual, "string")
			So(ok, ShouldBeTrue)
		})

		Convey("array fields should be extracted by name", func() {
			val, ok := extractFieldByName("b.1", testD)
			So(val, ShouldResemble, bson.D{{"inner2", 1}})
			So(ok, ShouldBeTrue)
			val, ok = extractFieldByName("b.1.inner2", testD)
			So(val, ShouldEqual, 1)
			So(ok, ShouldBeTrue)
			val, ok = extractFieldByName("b.0", testD)
			So(val, ShouldEqual, "inner")
			So(ok, ShouldBeTrue)
		})

		Convey("subdocument fields should be extracted by name", func() {
			val, ok := extractFieldByName("c", testD)
			So(val, ShouldResemble, bson.D{{"x", 5}})
			So(ok, ShouldBeTrue)
			val, ok = extractFieldByName("c.x", testD)
			So(val, ShouldEqual, 5)
			So(ok, ShouldBeTrue)

			Convey("even if they contain null values", func() {
				val, ok := extractFieldByName("d", testD)
				So(val, ShouldResemble, bson.D{{"z", nil}})
				So(ok, ShouldBeTrue)
				val, ok = extractFieldByName("d.z", testD)
				So(val, ShouldEqual, nil)
				So(ok, ShouldBeTrue)
				val, ok = extractFieldByName("d.z.nope", testD)
				So(val, ShouldEqual, nil)
				So(ok, ShouldBeFalse)
			})
		})

		Convey(`non-existing fields should return (nil,false)`, func() {
			for _, c := range []string{"f", "c.nope", "c.nope.NOPE", "b.1000", "b.1.nada"} {
				val, ok := extractFieldByName(c, testD)
				So(val, ShouldBeNil)
				So(ok, ShouldBeFalse)
			}
		})

	})

	Convey(`Extraction of a non-document should return (nil, false)`, t, func() {
		val, ok := extractFieldByName("meh", []interface{}{"meh"})
		So(val, ShouldBeNil)
		So(ok, ShouldBeFalse)
	})

	Convey(`Extraction of a nil document should return (nil, false)`, t, func() {
		val, ok := extractFieldByName("a", nil)
		So(val, ShouldEqual, nil)
		So(ok, ShouldBeFalse)
	})
}
