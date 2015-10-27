package sqlproxy

import (
	"database/sql"
	"fmt"
	"github.com/erh/mongo-sql-temp"
	"github.com/erh/mongo-sql-temp/config"
	"github.com/erh/mongo-sql-temp/proxy"
	_ "github.com/go-sql-driver/mysql"
	toolsdb "github.com/mongodb/mongo-tools/common/db"
	toolsLog "github.com/mongodb/mongo-tools/common/log"
	"github.com/mongodb/mongo-tools/common/options"
	"github.com/mongodb/mongo-tools/mongoimport"
	"github.com/siddontang/go-yaml/yaml"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
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
	JSONFile string `yaml:"json_file"`
}

type testCase struct {
	SQL         string `yaml:"sql"`
	Description string `yaml:"description"`
	expected    string `yaml:"expected"`
}

type testConfig struct {
	DB        string           `yaml:"db"`
	Data      []testDataSet    `yaml:"data"`
	Schemas   []*config.Schema `yaml:"schema"`
	TestCases []testCase       `yaml:"testcases"`
}

func testServer(cfg *config.Config) (*proxy.Server, error) {
	evaluator, err := sqlproxy.NewEvaluator(cfg)
	if err != nil {
		return nil, err
	}
	return proxy.NewServer(cfg, evaluator)
}

func buildSchemaMaps(conf *config.Config) {
	conf.Schemas = make(map[string]*config.Schema)
	for _, schema := range conf.RawSchemas {
		schema.Tables = make(map[string]*config.TableConfig)
		for _, table := range schema.RawTables {
			schema.Tables[table.Table] = table
		}
		conf.Schemas[schema.DB] = schema
	}
}

func compareResults(actual [][]interface{}, expected [][]interface{}) {
	So(actual, ShouldResemble, expected)
}

func importJSON(host, port, file, db, collection string) error {
	connection := &options.Connection{Host: host, Port: port}
	sessionProvider, err := toolsdb.NewSessionProvider(options.ToolOptions{
		Auth:       &options.Auth{},
		Connection: connection})
	if err != nil {
		return err
	}

	toolsLog.SetVerbosity(&options.Verbosity{Quiet: true})
	importer := mongoimport.MongoImport{
		ToolOptions: &options.ToolOptions{
			Connection: connection,
			Namespace: &options.Namespace{
				DB:         db,
				Collection: collection,
			},
			HiddenOptions: &options.HiddenOptions{NumDecodingWorkers: 10},
		},
		InputOptions: &mongoimport.InputOptions{File: file},
		IngestOptions: &mongoimport.IngestOptions{
			Drop:        true,
			StopOnError: true,
		},
		SessionProvider: sessionProvider,
	}
	_, err = importer.ImportDocuments()
	return err
}

func runSQL(db *sql.DB, query string) ([][]interface{}, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	result := [][]interface{}{}
	i := 0
	for rows.Next() {
		i += 1
		resultRow := make([]interface{}, len(cols))
		resultRowVals := make([]interface{}, len(cols))
		for i, _ := range resultRow {
			resultRow[i] = &resultRowVals[i]
		}
		if err := rows.Scan(resultRow...); err != nil {
			return nil, err
		}
		result = append(result, resultRowVals)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func executeTestCase(dbhost, dbport string, conf testConfig) error {
	// populate the DB with data for all the files in the config's list of json files
	for _, dataSet := range conf.Data {
		ns := strings.SplitN(dataSet.NS, ".", 2)
		if len(ns) != 2 {
			return fmt.Errorf("ns '%v' missing period; namespace should be specified as 'dbname.collection'", dataSet.NS)
		}
		err := importJSON(dbhost, dbport, dataSet.JSONFile, ns[0], ns[1])
		if err != nil {
			return err
		}
	}

	// make a test server using the embedded schema
	cfg := &config.Config{
		Addr:       testDBAddr,
		Url:        fmt.Sprintf("mongodb://%v:%v", dbhost, dbport),
		RawSchemas: conf.Schemas,
	}
	buildSchemaMaps(cfg)
	s, err := testServer(cfg)
	if err != nil {
		return err
	}
	go s.Run()
	defer s.Close()

	db, err := sql.Open("mysql", fmt.Sprintf("root@tcp(%v)/%v", testDBAddr, conf.Schemas[0].DB))
	if err != nil {
		return err
	}
	defer db.Close()

	for _, testCase := range conf.TestCases {
		description := testCase.SQL
		if testCase.Description != "" {
			description = testCase.Description
		}
		Convey(description, func() {
			_, err := runSQL(db, testCase.SQL)
			So(err, ShouldBeNil)
		})
	}
	return nil
}

func MustLoadTestConfig(path string) testConfig {
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var conf testConfig
	err = yaml.Unmarshal(fileBytes, &conf)
	if err != nil {
		panic(err)
	}
	return conf
}

func TestTableauDemo(t *testing.T) {
	Convey("Test tableau dataset", t, func() {
		conf := MustLoadTestConfig("testdata/tableau.yml")
		err := executeTestCase(testMongoHost, testMongoPort, conf)
		So(err, ShouldBeNil)
	})
}

func TestSimpleQueries(t *testing.T) {
	Convey("Test simple queries", t, func() {
		conf := MustLoadTestConfig("testdata/simple.yml")
		err := executeTestCase(testMongoHost, testMongoPort, conf)
		So(err, ShouldBeNil)
	})
}
