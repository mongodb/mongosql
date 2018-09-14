package integration

import (
	"fmt"
	"io/ioutil"
	"path"

	yaml "github.com/10gen/candiedyaml"
	"github.com/10gen/sqlproxy/internal/testutils/mongodb"
	"github.com/10gen/sqlproxy/schema"
)

// TestSuite represents a suite of end-to-end correctness tests.
type TestSuite struct {
	Dirname    string
	Name       string `yaml:"name"`
	DefaultDb  string `yaml:"default_db"`
	Tests      []*TestCase
	Benchmarks []*TestCase
}

// TestFile represents a single yml file in which integration tests are defined.
type TestFile struct {
	Name      string      `yaml:"name"`
	TestCases []*TestCase `yaml:"testcases"`
}

// TestCase represents an individual integration test.
type TestCase struct {
	Name                    string
	ID                      string                 `yaml:"id"`
	Database                string                 `yaml:"db"`
	Skip                    bool                   `yaml:"skip"`
	SQL                     string                 `yaml:"sql"`
	CleanupSQL              string                 `yaml:"sql_cleanup"`
	VerificationSQL         string                 `yaml:"verify"`
	Description             string                 `yaml:"description"`
	MinServerVersion        string                 `yaml:"min_server_version"`
	ExactServerVersion      string                 `yaml:"exact_server_version"`
	ExpectedError           string                 `yaml:"expected_error"`
	ExpectedTypes           []schema.SQLType       `yaml:"expected_types"`
	ExpectedNames           []string               `yaml:"expected_names"`
	ExpectedData            [][]interface{}        `yaml:"expected"`
	PushDownOnly            bool                   `yaml:"pushdown_only"`
	SchemaMappingHeuristics []string               `yaml:"schema_mapping_heuristics"`
	Sync                    bool                   `yaml:"sync"`
	Unordered               bool                   `yaml:"unordered"`
	Variables               map[string]interface{} `yaml:"variables"`
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

		err = tests.validate()
		if err != nil {
			fmt.Printf("Validation warning (%s): %s\n", filePath, err.Error())
		}

		addTestsToSuite(suite, tests)
	}

	return suite, nil
}

// RestoreSuiteData restores the test data needed for the tests in the suite
// with the provided name.
func RestoreSuiteData(name string) error {
	dataset, ok := datasetsByName[name]
	if !ok {
		return fmt.Errorf("No integration dataset with name %s", name)
	}
	return dataset.Restore(mongodb.GetToolOptions())
}

func addTestsToSuite(suite *TestSuite, tests *TestFile) {
	for _, t := range tests.TestCases {

		// set test name
		testName := tests.Name
		if t.ID != "" {
			testName = fmt.Sprintf("%s_%s", testName, t.ID)
		}
		t.Name = testName

		// set database
		if t.Database == "" {
			t.Database = suite.DefaultDb
		}
	}
	suite.Tests = append(suite.Tests, tests.TestCases...)
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

func (f *TestFile) validate() error {
	for i, t := range f.TestCases {
		if len(f.TestCases) > 1 && t.ID == "" {
			return fmt.Errorf("Id field not provided for testcase %d, but there is "+
				"more than one testcase in the file", i)
		}

		for _, typ := range t.ExpectedTypes {
			switch typ {
			case schema.SQLVarchar, schema.SQLInt, "float64", "int64":
				// this field will be handled as expected
			case schema.SQLDate, schema.SQLNumeric, "uint", schema.SQLFloat, "string":
				// this will be treated as a string, but should still be fine
			default:
				return fmt.Errorf(
					"Expected type '%s' not supported; will be treated as string",
					typ,
				)
			}
		}
	}

	return nil
}
