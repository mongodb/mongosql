package sqlproxy_test

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"database/sql/driver"

	yaml "github.com/10gen/candiedyaml"
	"github.com/10gen/sqlproxy"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/server"
	_ "github.com/go-sql-driver/mysql"
	toolsdb "github.com/mongodb/mongo-tools/common/db"
	toolsLog "github.com/mongodb/mongo-tools/common/log"
	"github.com/mongodb/mongo-tools/common/options"
	"github.com/mongodb/mongo-tools/mongorestore"
	. "github.com/smartystreets/goconvey/convey"
)

// test flags
var (
	blackbox = flag.Bool("blackbox", false, "Run blackbox tests")
	tableau  = flag.Bool("tableau", false, "Run tableau tests")
)

const (
	testMongoHost     = "127.0.0.1"
	testMongoPort     = "27017"
	testDBAddr        = "127.0.0.1:3308"
	testClientPEMFile = "testdata/client.pem"
)

type testDataSet struct {
	ArchiveFile string `yaml:"archive_file"`
}

type testCase struct {
	SQL           string          `yaml:"sql"`
	Description   string          `yaml:"description"`
	ExpectedTypes []string        `yaml:"expected_types"`
	ExpectedData  [][]interface{} `yaml:"expected"`
}

type testSchema struct {
	DB        string             `yaml:"db"`
	Data      []testDataSet      `yaml:"data"`
	Databases []*schema.Database `yaml:"schema"`
	TestCases []testCase         `yaml:"testcases"`
}

func TestBlackBox(t *testing.T) {
	if !*blackbox {
		t.Skip("skipping blackbox test")
	}

	conf := mustLoadTestSchema(pathify("testdata", "blackbox.yml"))
	mustLoadTestData(testMongoHost, testMongoPort, conf)

	opts := sqlproxy.Options{
		Addr:     testDBAddr,
		MongoURI: fmt.Sprintf("mongodb://%v:%v", testMongoHost, testMongoPort),
	}
	cfg := &schema.Schema{
		RawDatabases: conf.Databases,
	}
	buildSchemaMaps(cfg)
	s, err := testServer(cfg, opts)
	if err != nil {
		panic(err)
	}
	defer s.Close()
	go s.Run()

	Convey("Test blackbox dataset", t, func() {
		err := executeBlackBoxTestCases(t, conf)
		So(err, ShouldBeNil)
	})
}

func TestSimpleQueries(t *testing.T) {
	conf := mustLoadTestSchema("testdata/simple.yml")
	mustLoadTestData(testMongoHost, testMongoPort, conf)

	Convey("Test simple queries", t, func() {
		err := executeTestCase(t, testMongoHost, testMongoPort, conf)
		So(err, ShouldBeNil)
	})
}

func TestShowQueries(t *testing.T) {
	conf := mustLoadTestSchema("testdata/show.yml")
	mustLoadTestData(testMongoHost, testMongoPort, conf)

	Convey("Test show queries", t, func() {
		err := executeTestCase(t, testMongoHost, testMongoPort, conf)
		So(err, ShouldBeNil)
	})
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
		for colnum, col := range row {
			if rownum > len(expected)-1 {
				t.Errorf("expected %v rows but got %v", len(expected), len(actual))
				return
			}
			expectedCol := expected[rownum][colnum]
			// we don't have a good way of representing
			// nil in our CSV test results so we check
			// for an empty string when we expect nil
			if expectedCol == nil && col == "" {
				expectedCol = ""
			}
			// our YAML parser converts numbers in the
			// int range to int but we need them to be
			// int64
			if v, ok := expectedCol.(int); ok {
				expectedCol = int64(v)
			}
			So(col, ShouldResemble, expectedCol)
		}
	}
	So(len(actual), ShouldEqual, len(expected))
}

func executeBlackBoxTestCases(t *testing.T, conf testSchema) error {
	db, err := sql.Open("mysql", fmt.Sprintf("root@tcp(%v)/%v", testDBAddr, conf.Databases[0].Name))
	if err != nil {
		return fmt.Errorf("mysql open: %v", err)
	}
	defer db.Close()

	r, err := os.Open(pathify("testdata", "blackbox_queries.json"))
	if err != nil {
		return fmt.Errorf("Open: %v", err)
	}

	dec := json.NewDecoder(r)

	type queryData struct {
		Id      string `json:"id"`
		Query   string `json:"value"`
		Columns int    `json:"columns"`
		Rows    int    `json:"rows"`
	}

	query := queryData{}

	for {
		err := dec.Decode(&query)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("Decode: %v", err)
		}

		var types []string

		Convey(fmt.Sprintf("Running test query (%v): '%v'", query.Id, query.Query), func() {
			t.Logf("query: %v %v", query.Id, query.Query)

			for j := 0; j < query.Columns; j++ {
				types = append(types, schema.SQLVarchar)
			}

			results, err := runSQL(db, query.Query, types)
			So(err, ShouldBeNil)

			resultFile := pathify("testdata", "results", fmt.Sprintf("%v.csv", query.Id))

			handle, err := os.OpenFile(resultFile, os.O_RDONLY, os.ModeExclusive)
			if err != nil {
				panic(err)
			}

			r := csv.NewReader(handle)

			r.FieldsPerRecord = query.Columns
			data, err := r.ReadAll()
			So(err, ShouldBeNil)

			// TODO: check header as well
			data = data[1:]
			newData := make([][]interface{}, len(data))
			for i, v := range data {
				for _, s := range v {
					newData[i] = append(newData[i], s)
				}
			}

			compareResults(t, results, newData)

		})
	}

	return nil
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
	go s.Run()
	defer s.Close()

	db, err := sql.Open("mysql", fmt.Sprintf("root@tcp(%v)/%v", testDBAddr, conf.Databases[0].Name))
	if err != nil {
		return err
	}
	defer db.Close()

	for _, testCase := range conf.TestCases {
		description := testCase.SQL
		if testCase.Description != "" {
			description = testCase.Description
		}
		Convey(fmt.Sprintf("%v('%v')", description, testCase.SQL), func() {
			results, err := runSQL(db, testCase.SQL, testCase.ExpectedTypes)
			So(err, ShouldBeNil)
			compareResults(t, testCase.ExpectedData, results)
		})

	}
	return nil
}

func mustLoadTestData(dbhost, dbport string, conf testSchema) {
	for _, dataSet := range conf.Data {
		if err := restoreBSON(dbhost, dbport, dataSet.ArchiveFile); err != nil {
			panic(err)
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

func restoreBSON(host, port, file string) error {
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

	connection := &options.Connection{Host: host, Port: port}
	sessionProvider, err := toolsdb.NewSessionProvider(
		options.ToolOptions{
			Auth:       &options.Auth{},
			Connection: connection,
			SSL:        sslOpts,
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

func runSQL(db *sql.DB, query string, types []string) ([][]interface{}, error) {
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
