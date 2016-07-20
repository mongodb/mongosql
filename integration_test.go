package sqlproxy_test

import (
	"database/sql"
	"database/sql/driver"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/mgo.v2/bson"

	yaml "github.com/10gen/candiedyaml"
	"github.com/10gen/sqlproxy"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/server"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mongodb/mongo-tools/common/bsonutil"
	toolsdb "github.com/mongodb/mongo-tools/common/db"
	"github.com/mongodb/mongo-tools/common/log"
	toolsLog "github.com/mongodb/mongo-tools/common/log"
	"github.com/mongodb/mongo-tools/common/options"
	"github.com/mongodb/mongo-tools/mongorestore"
	. "github.com/smartystreets/goconvey/convey"
)

// test flags
var (
	tableau     = flag.Bool("tableau", false, "Run tableau tests")
	ymlfile     = flag.String("ymlfile", "", "The yml test file to run")
	ymltestname = flag.String("ymltestname", "", "The test name in the yml file to run")
)

const (
	testMongoHost     = "127.0.0.1"
	testMongoPort     = "27017"
	testDBAddr        = "127.0.0.1:3308"
	testClientPEMFile = "testdata/client.pem"
)

type testDataSet struct {
	// ArchiveFile represents a mongodump file that should be restored.
	// It will generally be used when a large amount of data is needed
	// for tests.
	ArchiveFile string `yaml:"archive_file"`
	// Inline represents data that is specified in the yaml file.
	// It will generally be used when a small amount of data is needed
	// for tests.
	Inline *inlineDataSet `yaml:"inline"`
}

type inlineDataSet struct {
	Db         string   `yaml:"db"`
	Collection string   `yaml:"collection"`
	Docs       []bson.D `yaml:"docs"`
}

type testCase struct {
	SQL             string          `yaml:"sql"`
	VerificationSQL string          `yaml:"verify"`
	Description     string          `yaml:"description"`
	ExpectedTypes   []string        `yaml:"expected_types"`
	ExpectedNames   []string        `yaml:"expected_names"`
	ExpectedData    [][]interface{} `yaml:"expected"`
}

type testSchema struct {
	DB        string             `yaml:"db"`
	Data      []testDataSet      `yaml:"data"`
	Databases []*schema.Database `yaml:"schema"`
	TestCases []testCase         `yaml:"testcases"`
}

func TestTableauDemo(t *testing.T) {
	if !*tableau {
		t.Skip("skipping tableau test")
	}

	conf := mustLoadTestSchema("testdata/tableau.yml")
	mustLoadTestData(testMongoHost, testMongoPort, conf)

	Convey("Test tableau dataset", t, func() {
		err := executeTestCase(t, testMongoHost, testMongoPort, conf)
		So(err, ShouldBeNil)
	})
}

func TestIntegration(t *testing.T) {
	files, err := ioutil.ReadDir("integration_tests")
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		if *ymlfile != "" && !strings.Contains(f.Name(), *ymlfile) {
			continue
		}
		conf := mustLoadTestSchema(path.Join("integration_tests", f.Name()))
		mustLoadTestData(testMongoHost, testMongoPort, conf)

		Convey(f.Name(), t, func() {
			err := executeTestCase(t, testMongoHost, testMongoPort, conf)
			So(err, ShouldBeNil)
		})
	}
}

func buildSchemaMaps(conf *schema.Schema) {
	conf.Databases = make(map[string]*schema.Database)
	for _, db := range conf.RawDatabases {
		db.Tables = make(map[string]*schema.Table)
		for _, table := range db.RawTables {
			db.Tables[table.Name] = table
		}
		conf.Databases[db.Name] = db
	}
}

func compareResults(t *testing.T, expected, actual [][]interface{}) {
	for rownum, row := range actual {
		for colnum, actualCol := range row {
			So(rownum, ShouldBeLessThan, len(expected))
			expectedCol := expected[rownum][colnum]
			// our YAML parser converts numbers in the
			// int range to int but we need them to be
			// int64
			if v, ok := expectedCol.(int); ok {
				expectedCol = int64(v)
			}

			if actualCol != expectedCol {
				So(fmt.Sprintf("(%d,%d): %v", rownum, colnum, actualCol), ShouldEqual, fmt.Sprintf("(%d,%d): %v", rownum, colnum, expectedCol))
			}
		}
	}

	So(len(actual), ShouldEqual, len(expected))
}

func executeTestCase(t *testing.T, dbhost, dbport string, conf testSchema) error {
	// make a test server using the embedded database
	cfg := &schema.Schema{
		RawDatabases: conf.Databases,
	}
	opts := sqlproxy.Options{
		Addr:     testDBAddr,
		MongoURI: fmt.Sprintf("mongodb://%v:%v", dbhost, dbport),
	}
	buildSchemaMaps(cfg)
	s, err := testServer(cfg, opts)
	if err != nil {
		return err
	}
	log.SetWriter(ioutil.Discard)
	go s.Run()
	defer s.Close()

	var db *sql.DB
	if len(conf.Databases) > 0 {
		db, err = sql.Open("mysql", fmt.Sprintf("root@tcp(%v)/%v", testDBAddr, conf.Databases[0].Name))
	} else {
		db, err = sql.Open("mysql", fmt.Sprintf("root@tcp(%v)/", testDBAddr))
	}

	if err != nil {
		return err
	}
	defer db.Close()

	for _, testCase := range conf.TestCases {
		description := testCase.SQL
		if testCase.Description != "" {
			description = testCase.Description
		}
		if *ymltestname != "" && !strings.Contains(description, *ymltestname) {
			continue
		}

		Convey(fmt.Sprintf("%v('%v')", description, testCase.SQL), func() {
			sql := testCase.SQL
			if testCase.VerificationSQL != "" {
				_, err := db.Exec(sql)
				So(err, ShouldBeNil)

				sql = testCase.VerificationSQL
			}
			results, err := runSQL(db, sql, testCase.ExpectedTypes, testCase.ExpectedNames)
			So(err, ShouldBeNil)
			compareResults(t, testCase.ExpectedData, results)
		})
	}
	return nil
}

func mustLoadTestData(dbhost, dbport string, conf testSchema) {
	for _, dataSet := range conf.Data {
		if dataSet.ArchiveFile != "" {
			if err := restoreBSON(dbhost, dbport, dataSet.ArchiveFile); err != nil {
				panic(err)
			}
		} else if dataSet.Inline != nil {
			if err := restoreInline(dbhost, dbport, dataSet.Inline); err != nil {
				panic(err)
			}
		} else {
			panic("expected 'archive_file' or 'inline'")
		}

	}
}

func mustLoadTestSchema(path string) testSchema {
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var conf testSchema

	err = yaml.Unmarshal(fileBytes, &conf)
	if err != nil {
		panic(err)
	}

	for _, db := range conf.Databases {
		if err := schema.PopulateColumnMaps(db); err != nil {
			panic(err)
		}
		if err := schema.HandlePipeline(db); err != nil {
			panic(err)
		}
	}

	return conf
}

func pathify(elem ...string) string {
	return filepath.Join(elem...)
}

func getSslOpts() *options.SSL {
	var sslOpts *options.SSL
	if len(os.Getenv(evaluator.SSLTestKey)) > 0 {
		toolsdb.GetConnectorFuncs = append(toolsdb.GetConnectorFuncs,
			func(opts options.ToolOptions) toolsdb.DBConnector {
				if opts.SSL.UseSSL {
					return &sqlproxy.SSLDBConnector{}
				}
				return nil
			},
		)
		sslOpts = &options.SSL{
			UseSSL:              true,
			SSLPEMKeyFile:       "testdata/client.pem",
			SSLAllowInvalidCert: true,
		}
	}

	return sslOpts
}

func restoreInline(host, port string, inline *inlineDataSet) error {
	connection := &options.Connection{Host: host, Port: port}
	sessionProvider, err := toolsdb.NewSessionProvider(
		options.ToolOptions{
			Auth:       &options.Auth{},
			Connection: connection,
			SSL:        getSslOpts(),
		},
	)
	if err != nil {
		return err
	}
	sessionProvider.SetFlags(toolsdb.DisableSocketTimeout)

	session, err := sessionProvider.GetSession()
	if err != nil {
		return err
	}

	c := session.DB(inline.Db).C(inline.Collection)
	c.DropCollection() // don't care about the result
	bulk := c.Bulk()
	for _, d := range inline.Docs {
		doc, err := bsonutil.ConvertJSONValueToBSON(d)
		if err != nil {
			panic(fmt.Sprintf("unable to parse extended json %v error: %v", d, err))
		}
		bulk.Insert(doc)
	}
	_, err = bulk.Run()
	return err
}

func restoreBSON(host, port, file string) error {
	connection := &options.Connection{Host: host, Port: port}
	sessionProvider, err := toolsdb.NewSessionProvider(
		options.ToolOptions{
			Auth:       &options.Auth{},
			Connection: connection,
			SSL:        getSslOpts(),
		},
	)
	if err != nil {
		return err
	}
	sessionProvider.SetFlags(toolsdb.DisableSocketTimeout)

	if err != nil {
		return err
	}

	toolsLog.SetVerbosity(&options.Verbosity{Quiet: true})

	restorer := mongorestore.MongoRestore{
		ToolOptions: &options.ToolOptions{
			Connection: connection,
			Namespace:  &options.Namespace{},
			HiddenOptions: &options.HiddenOptions{
				NumDecodingWorkers: 10,
				BulkBufferSize:     1000,
			},
		},
		InputOptions: &mongorestore.InputOptions{Gzip: true, Archive: file},
		OutputOptions: &mongorestore.OutputOptions{
			Drop:                   true,
			StopOnError:            true,
			NumParallelCollections: 1,
			NumInsertionWorkers:    10,
			MaintainInsertionOrder: true,
		},
		SessionProvider: sessionProvider,
	}

	return restorer.Restore()
}

func runSQL(db *sql.DB, query string, types []string, names []string) ([][]interface{}, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	if len(cols) != len(types) {
		return nil, fmt.Errorf("Number of columns in result set (%v) does not match columns in expected types (%v)", len(cols), len(types))
	}

	for i, n := range names {
		if cols[i] != n {
			return nil, fmt.Errorf("Expected name %q at index %d, but found %q", n, i, cols[i])
		}
	}

	result := [][]interface{}{}

	resultContainer := make([]interface{}, 0, len(types))

	for _, t := range types {
		switch schema.SQLType(t) {
		case schema.SQLVarchar:
			resultContainer = append(resultContainer, &sql.NullString{})
		case schema.SQLInt:
			resultContainer = append(resultContainer, &sql.NullInt64{})
		case schema.SQLFloat:
			resultContainer = append(resultContainer, &sql.NullFloat64{})
		default:
			resultContainer = append(resultContainer, &sql.NullString{})
		}
	}

	for rows.Next() {
		resultRow := make([]interface{}, 0, len(types))
		if err := rows.Scan(resultContainer...); err != nil {
			return nil, err
		}
		for _, x := range resultContainer {
			v, err := x.(driver.Valuer).Value()
			if err != nil {
				return nil, err
			}
			resultRow = append(resultRow, v)
		}
		result = append(result, resultRow)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func testServer(cfg *schema.Schema, opts sqlproxy.Options) (*server.Server, error) {
	if len(os.Getenv(evaluator.SSLTestKey)) > 0 {
		opts.MongoSSL = true
		opts.MongoAllowInvalidCerts = true
		opts.MongoPEMFile = testClientPEMFile
	}
	evaluator, err := sqlproxy.NewEvaluator(cfg, opts)
	if err != nil {
		return nil, err
	}
	return server.New(cfg, evaluator, opts)
}
