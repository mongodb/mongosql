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
	"reflect"
	"testing"

	"github.com/10gen/sqlproxy"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/proxy"
	"github.com/10gen/sqlproxy/schema"
	_ "github.com/go-sql-driver/mysql"
	toolsdb "github.com/mongodb/mongo-tools/common/db"
	toolsLog "github.com/mongodb/mongo-tools/common/log"
	"github.com/mongodb/mongo-tools/common/options"
	"github.com/mongodb/mongo-tools/mongorestore"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/yaml.v2"
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
	NS          string `yaml:"ns"`
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

func testServer(cfg *schema.Schema) (*proxy.Server, error) {
	if len(os.Getenv(evaluator.SSLTestKey)) > 0 {
		cfg.SSL = &schema.SSL{
			AllowInvalidCerts: true,
			PEMKeyFile:        testClientPEMFile,
		}

	}
	evaluator, err := sqlproxy.NewEvaluator(cfg)
	if err != nil {
		return nil, err
	}
	return proxy.NewServer(cfg, evaluator)
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

func compareResults(t *testing.T, actual [][]interface{}, expected [][]interface{}) {
	So(len(actual), ShouldEqual, len(expected))
	for rownum, row := range actual {
		for colnum, col := range row {
			expectedCol := expected[rownum][colnum]
			So(col, ShouldResemble, expectedCol)
		}
	}
}

func restoreBSON(host, port, file string) error {
	var sslOpts *options.SSL
	if len(os.Getenv(evaluator.SSLTestKey)) > 0 {
		toolsdb.GetConnectorFuncs = append(toolsdb.GetConnectorFuncs,
			func(opts options.ToolOptions) toolsdb.DBConnector {
				if opts.SSL.UseSSL {
					return &evaluator.SSLDBConnector{}
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
			resultContainer = append(resultContainer, new(string))
		case schema.SQLInt:
			resultContainer = append(resultContainer, new(int))
		case schema.SQLFloat:
			resultContainer = append(resultContainer, new(float64))
		default:
			resultContainer = append(resultContainer, new(string))
		}
	}

	for rows.Next() {
		resultRow := make([]interface{}, 0, len(types))
		if err := rows.Scan(resultContainer...); err != nil {
			return nil, err
		}
		for _, x := range resultContainer {
			p := reflect.ValueOf(x)
			resultRow = append(resultRow, p.Elem().Interface())
		}
		result = append(result, resultRow)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func executeTestCase(t *testing.T, dbhost, dbport string, conf testSchema) error {
	// make a test server using the embedded database
	cfg := &schema.Schema{
		Addr:         testDBAddr,
		Url:          fmt.Sprintf("mongodb://%v:%v", dbhost, dbport),
		RawDatabases: conf.Databases,
	}
	buildSchemaMaps(cfg)
	s, err := testServer(cfg)
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

	for i, testCase := range conf.TestCases {
		description := testCase.SQL
		if testCase.Description != "" {
			description = testCase.Description
		}
		Convey(description, func() {
			t.Logf("Running test query (%v of %v): '%v'", i+1, len(conf.TestCases), testCase.SQL)
			results, err := runSQL(db, testCase.SQL, testCase.ExpectedTypes)
			So(err, ShouldBeNil)
			compareResults(t, results, testCase.ExpectedData)
		})

	}
	return nil
}

func MustLoadTestSchema(path string) testSchema {
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
	}

	return conf
}

func MustLoadTestData(dbhost, dbport string, conf testSchema) {
	for _, dataSet := range conf.Data {
		if err := restoreBSON(dbhost, dbport, dataSet.ArchiveFile); err != nil {
			panic(err)
		}
	}
}

func TestBlackBox(t *testing.T) {
	if !*blackbox {
		t.Skip("skipping blackbox test")
	}

	conf := MustLoadTestSchema(pathify("testdata", "blackbox.yml"))
	MustLoadTestData(testMongoHost, testMongoPort, conf)

	cfg := &schema.Schema{
		Addr:         testDBAddr,
		Url:          fmt.Sprintf("mongodb://%v:%v", testMongoHost, testMongoPort),
		RawDatabases: conf.Databases,
	}
	buildSchemaMaps(cfg)
	s, err := testServer(cfg)
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

func TestSimpleQueries(t *testing.T) {
	conf := MustLoadTestSchema("testdata/simple.yml")
	MustLoadTestData(testMongoHost, testMongoPort, conf)

	Convey("Test simple queries", t, func() {
		err := executeTestCase(t, testMongoHost, testMongoPort, conf)
		So(err, ShouldBeNil)
	})
}

func TestTableauDemo(t *testing.T) {
	if !*tableau {
		t.Skip("skipping tableau test")
	}

	conf := MustLoadTestSchema("testdata/tableau.yml")
	MustLoadTestData(testMongoHost, testMongoPort, conf)

	Convey("Test tableau dataset", t, func() {
		err := executeTestCase(t, testMongoHost, testMongoPort, conf)
		So(err, ShouldBeNil)
	})
}

func pathify(elem ...string) string {
	return filepath.Join(elem...)
}
