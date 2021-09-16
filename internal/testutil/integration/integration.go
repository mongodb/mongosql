package integration

import (
	"fmt"
	"io/ioutil"
	"path"
	"time"

	yaml "github.com/10gen/candiedyaml"
	"github.com/10gen/sqlproxy/internal/testutil/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"github.com/shopspring/decimal"
)

// TestSuite represents a suite of end-to-end correctness tests.
type TestSuite struct {
	Dirname   string
	Name      string `yaml:"name"`
	DefaultDb string `yaml:"default_db"`
	Tests     []*TestCase
}

// TestFile represents a single yml file in which integration tests are defined.
type TestFile struct {
	Name      string      `yaml:"name"`
	TestCases []*TestCase `yaml:"testcases"`
}

// TestCase represents an individual integration test.
type TestCase struct {
	Name                  string
	ID                    string                 `yaml:"id"`
	Database              string                 `yaml:"db"`
	Skip                  bool                   `yaml:"skip"`
	SQL                   string                 `yaml:"sql"`
	SQLList               []string               `yaml:"sql_list"`
	CleanupSQL            string                 `yaml:"sql_cleanup"`
	CleanupSQLList        []string               `yaml:"sql_cleanup_list"`
	VerificationSQL       string                 `yaml:"verify"`
	Description           string                 `yaml:"description"`
	MinServerVersion      string                 `yaml:"min_server_version"`
	MaxServerVersion      string                 `yaml:"max_server_version"`
	ExactServerVersion    string                 `yaml:"exact_server_version"`
	ExpectedError         string                 `yaml:"expected_error"`
	ExpectedErrorContains string                 `yaml:"expected_error_contains"`
	ExpectedTypes         []schema.SQLType       `yaml:"expected_types"`
	ExpectedNames         []string               `yaml:"expected_names"`
	ExpectedData          [][]interface{}        `yaml:"expected_results"`
	PushDownOnly          bool                   `yaml:"pushdown_only"`
	Sync                  bool                   `yaml:"sync"`
	Unordered             bool                   `yaml:"unordered"`
	Variables             map[string]interface{} `yaml:"variables"`
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
			return nil, fmt.Errorf("%s: %v", filePath, err)
		}

		err = tests.fixExpectedTypes()
		if err != nil {
			return nil, fmt.Errorf("%s: %v", filePath, err)
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

func (f *TestFile) fixExpectedTypes() error {
	for _, test := range f.TestCases {
		if test.Skip {
			continue
		}
		err := test.fixExpectedTypes()
		if err != nil {
			return fmt.Errorf("%s: %v", test.ID, err)
		}
	}
	return nil
}

func (f *TestCase) fixExpectedTypes() error {
	if f.ExpectedError != "" {
		return nil
	}

	var err error
	for rowNum, row := range f.ExpectedData {
		for colNum, val := range row {
			row[colNum], err = castExpected(val, f.ExpectedTypes[colNum])
			if err != nil {
				return fmt.Errorf("row %d, col %d: %v", rowNum, colNum, err)
			}
		}
	}

	return nil
}

func castExpected(val interface{}, typ schema.SQLType) (interface{}, error) {
	if val == nil {
		return nil, nil
	}

	switch typ {
	case schema.SQLVarchar:
		strVal, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf(
				"value of type %T not valid for column with expected_type %s",
				val, typ,
			)
		}
		return strVal, nil

	case schema.SQLInt:
		intVal, ok := val.(int64)
		if !ok {
			return nil, fmt.Errorf(
				"value of type %T not valid for column with expected_type %s",
				val, typ,
			)
		}
		return intVal, nil

	case schema.SQLFloat:
		switch typedV := val.(type) {
		case float64:
			return typedV, nil
		case int64:
			return float64(typedV), nil
		default:
			return nil, fmt.Errorf(
				"value of type %T not valid for column with expected_type %s",
				val, typ,
			)
		}

	case schema.SQLDecimal:
		switch typedV := val.(type) {
		case float64:
			dec := decimal.NewFromFloat(typedV)
			return dec, nil
		case int64:
			dec := decimal.New(typedV, 0)
			return dec, nil
		case string:
			return decimal.NewFromString(typedV)
		default:
			return nil, fmt.Errorf(
				"value of type %T not valid for column with expected_type %s",
				val, typ,
			)
		}

	case schema.SQLDate:
		timeVal, ok := val.(time.Time)
		if !ok {
			return nil, fmt.Errorf(
				"value of type %T not valid for column with expected_type %s",
				val, typ,
			)
		}
		return timeVal, nil

	default:
		return nil, fmt.Errorf("expected type '%s' not valid", typ)
	}
}

func (f *TestFile) validate() error {
	for i, t := range f.TestCases {
		if t.Skip {
			continue
		}

		if len(f.TestCases) > 1 && t.ID == "" {
			return fmt.Errorf("Id field not provided for testcase %d, but there is "+
				"more than one testcase in the file", i)
		}

		if len(t.ExpectedData) > 0 {
			if len(t.ExpectedTypes) != len(t.ExpectedData[0]) {
				return fmt.Errorf("expected types count does not match expected "+
					"data dimension for test %s", t.ID)
			}
		}
	}

	return nil
}
