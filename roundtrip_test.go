package sqlproxy

import (
	"database/sql"
	"fmt"
	"github.com/10gen/sqlproxy"
	"github.com/10gen/sqlproxy/proxy"
	"github.com/10gen/sqlproxy/schema"
	_ "github.com/go-sql-driver/mysql"
	toolsdb "github.com/mongodb/mongo-tools/common/db"
	toolsLog "github.com/mongodb/mongo-tools/common/log"
	"github.com/mongodb/mongo-tools/common/options"
	"github.com/mongodb/mongo-tools/mongorestore"
	"github.com/siddontang/go-yaml/yaml"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

const (
	testMongoHost = "127.0.0.1"
	testMongoPort = "27017"
	testDBAddr    = "127.0.0.1:3308"
)

type testDataSet struct {
	NS       string `yaml:"ns"`
	BSONFile string `yaml:"bson_file"`
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
		t.Logf("comparing row %v of %v", rownum+1, len(expected))
		for colnum, col := range row {
			expectedCol := expected[rownum][colnum]
			So(col, ShouldResemble, expectedCol)
		}
	}
}

func restoreBSON(host, port, file, db, collection string) error {

	connection := &options.Connection{Host: host, Port: port}
	sessionProvider, err := toolsdb.NewSessionProvider(options.ToolOptions{
		Auth:       &options.Auth{},
		Connection: connection})

	if err != nil {
		return err
	}

	toolsLog.SetVerbosity(&options.Verbosity{Quiet: true})

	restorer := mongorestore.MongoRestore{
		ToolOptions: &options.ToolOptions{
			Connection: connection,
			Namespace: &options.Namespace{
				DB:         db,
				Collection: collection,
			},
			HiddenOptions: &options.HiddenOptions{NumDecodingWorkers: 10},
		},
		InputOptions: &mongorestore.InputOptions{Gzip: true},
		OutputOptions: &mongorestore.OutputOptions{
			Drop:                   true,
			StopOnError:            true,
			NumParallelCollections: 1,
			NumInsertionWorkers:    1,
		},
		SessionProvider: sessionProvider,
		TargetDirectory: file,
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
		switch t {
		case schema.SQLString:
			resultContainer = append(resultContainer, new(string))
		case schema.SQLInt:
			resultContainer = append(resultContainer, new(int))
		case schema.SQLFloat:
			resultContainer = append(resultContainer, new(float64))
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
	// populate the DB with data for all the files in the config's list of json files
	for _, dataSet := range conf.Data {
		ns := strings.SplitN(dataSet.NS, ".", 2)
		if len(ns) != 2 {
			return fmt.Errorf("ns '%v' missing period; namespace should be specified as 'dbname.collection'", dataSet.NS)
		}
		err := restoreBSON(dbhost, dbport, dataSet.BSONFile, ns[0], ns[1])
		if err != nil {
			return err
		}
	}

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

	return conf
}

func TestTableauDemo(t *testing.T) {
	Convey("Test tableau dataset", t, func() {
		conf := MustLoadTestSchema("testdata/tableau.yml")
		err := executeTestCase(t, testMongoHost, testMongoPort, conf)
		So(err, ShouldBeNil)
	})
}

func TestSimpleQueries(t *testing.T) {
	Convey("Test simple queries", t, func() {
		conf := MustLoadTestSchema("testdata/simple.yml")
		err := executeTestCase(t, testMongoHost, testMongoPort, conf)
		So(err, ShouldBeNil)
	})
}
