package testutils

import (
	"fmt"
	"io/ioutil"
	"path"

	yaml "github.com/10gen/candiedyaml"
	"github.com/10gen/sqlproxy/schema"

	"gopkg.in/mgo.v2/bson"
)

type TestSuite struct {
	Dirname    string
	Name       string         `yaml:"name"`
	Data       []*TestDataSet `yaml:"data"`
	DefaultDb  string         `yaml:"default_db"`
	Tests      []*TestCase
	Benchmarks []*TestCase
}

type TestDataSet struct {
	ArchiveFile      string         `yaml:"archive_file"`
	Inline           *InlineDataSet `yaml:"inline"`
	MinServerVersion string         `yaml:"min_server_version"`
}

type InlineDataSet struct {
	Db         string   `yaml:"db"`
	Collection string   `yaml:"collection"`
	Collation  bson.D   `yaml:"collation"`
	Docs       []bson.D `yaml:"docs"`
}

type TestFile struct {
	Name       string      `yaml:"name"`
	TestCases  []*TestCase `yaml:"testcases"`
	Benchmarks []*TestCase `yaml:"benchmarks"`
}

type TestCase struct {
	Name             string
	IsBenchmark      bool
	Id               string          `yaml:"id"`
	Database         string          `yaml:"db"`
	Skip             bool            `yaml:"skip"`
	SQL              string          `yaml:"sql"`
	CleanupSQL       string          `yaml:"sql_cleanup"`
	VerificationSQL  string          `yaml:"verify"`
	Description      string          `yaml:"description"`
	MinServerVersion string          `yaml:"min_server_version"`
	ExpectedError    string          `yaml:"expected_error"`
	ExpectedTypes    []string        `yaml:"expected_types"`
	ExpectedNames    []string        `yaml:"expected_names"`
	ExpectedData     [][]interface{} `yaml:"expected"`
	PushDownOnly     bool            `yaml:"pushdown_only"`
}

// LoadTestSuite returns a testSuite struct populated with the
// data from the test suite located at testdata/suites/<name>/
func LoadTestSuite(name string) (*TestSuite, error) {
	suite := loadTestSuiteConfig(name)

	suiteDir := path.Join("testdata", "suites", name)
	files, err := ioutil.ReadDir(suiteDir)
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		if f.Name() == "_suite.yml" {
			continue
		}

		filePath := path.Join(suiteDir, f.Name())
		fileBytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			panic(err)
		}

		tests := new(TestFile)
		err = yaml.Unmarshal(fileBytes, tests)
		if err != nil {
			return nil, fmt.Errorf("Error unmarshalling %s: %s", f.Name(), err.Error())
		}
		if tests.Name == "" {
			tests.Name = f.Name()[0 : len(f.Name())-4]
		}

		err = tests.validate(f.Name())
		if err != nil {
			fmt.Printf("Validation warning (%s): %s\n", filePath, err.Error())
		}

		addTestsToSuite(suite, tests)
	}

	return suite, nil
}

func RestoreSuiteData(name string) error {
	suite := loadTestSuiteConfig(name)
	for i, dataSet := range suite.Data {
		if !MongodbVersionAtLeast(dataSet.MinServerVersion) {
			continue
		}
		if dataSet.ArchiveFile != "" {
			err := restoreBSON(*MongoHost, *MongoPort, dataSet.ArchiveFile)
			if err != nil {
				return err
			}
		} else if dataSet.Inline != nil {
			err := restoreInline(*MongoHost, *MongoPort, dataSet.Inline)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("could not find 'archive_file' or 'inline' key for data set %d", i)
		}
	}
	return nil
}

func addTestsToSuite(suite *TestSuite, tests *TestFile) {
	for _, t := range tests.TestCases {

		// set test name
		testName := tests.Name
		if t.Id != "" {
			testName = fmt.Sprintf("%s_%s", testName, t.Id)
		}
		t.Name = testName

		// set database
		if t.Database == "" {
			t.Database = suite.DefaultDb
		}
	}

	for _, b := range tests.Benchmarks {
		benchName := tests.Name
		if b.Id != "" {
			benchName = fmt.Sprintf("%s_%s", benchName, b.Id)
		}
		b.Name = benchName

		// set database
		if b.Database == "" {
			b.Database = suite.DefaultDb
		}

		b.IsBenchmark = true
	}

	suite.Tests = append(suite.Tests, tests.TestCases...)
	suite.Benchmarks = append(suite.Benchmarks, tests.Benchmarks...)
}

func loadTestSuiteConfig(name string) *TestSuite {
	suiteDir := path.Join("testdata", "suites", name)

	suiteFile := path.Join(suiteDir, "_suite.yml")
	fileBytes, err := ioutil.ReadFile(suiteFile)
	if err != nil {
		panic(err)
	}

	suite := new(TestSuite)
	err = yaml.Unmarshal(fileBytes, suite)
	if err != nil {
		panic(err)
	}
	suite.Dirname = name
	if suite.Name == "" {
		suite.Name = suite.Dirname
	}

	return suite
}

func (f *TestFile) validate(filename string) error {
	for i, t := range f.TestCases {
		if len(f.TestCases) > 1 && t.Id == "" {
			return fmt.Errorf("Id field not provided for testcase %d, but there is more than one testcase in the file", i)
		}

		for _, typ := range t.ExpectedTypes {
			switch typ {
			case schema.SQLVarchar, schema.SQLInt, schema.SQLFloat:
				// this field will be handled as expected
			case schema.SQLInt64, schema.SQLDate, schema.SQLNumeric, "uint", "float", "string":
				// this will be treated as a string, but should still be fine
			default:
				return fmt.Errorf("Expected type '%s' not supported; will be treated as string", typ)
			}
		}
	}

	for i, b := range f.Benchmarks {
		if len(f.Benchmarks) > 1 && b.Id == "" {
			return fmt.Errorf("Id field not provided for benchmark %d, but there is more than one benchmark in the file", i)
		}
	}

	return nil
}
