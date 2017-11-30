package testutils

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/testutils/mongodb"
	"github.com/10gen/sqlproxy/schema"

	// for using the mysql driver with database/sql
	_ "github.com/go-sql-driver/mysql"
)

var maxTime = time.Duration(*MaxTimeSecs) * time.Second

// RunSQL runs the provided SQL query using the provided database handle.
// It expects the results to have the provided column names and types, and
// returns a list of rows and an error.
func RunSQL(db *sql.DB, query string, types []string, names []string) ([][]interface{}, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	if len(types) > 0 && len(cols) != len(types) {
		return nil, fmt.Errorf("Number of columns in result set (%v) does not match columns in expected types (%v)", len(cols), len(types))
	}

	for i, n := range names {
		// This is a hack to get around candiedyaml converting "Null" to ""
		if n == "" {
			n = "Null"
		}

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

// CompareResults checks whether a one set of SQL results matches another,
// returning an error if they do not match.
func CompareResults(expected [][]interface{}, actual [][]interface{}) error {
	if len(actual) != len(expected) {
		return fmt.Errorf("Expected %d rows, got %d rows", len(expected), len(actual))
	}

	for rownum, row := range actual {
		for colnum, actualCol := range row {

			expectedCol := expected[rownum][colnum]
			// our YAML parser converts numbers in the
			// int range to int but we need them to be
			// int64
			if v, ok := expectedCol.(int); ok {
				expectedCol = int64(v)
			}

			// Because of the int conversion behavior,
			// if our actual result is a float64, we
			// need to convert it to an int64
			if _, ok := expectedCol.(int64); ok {
				if v, ok := actualCol.(float64); ok {
					actualCol = int64(v)
				}
			}

			if actualCol != expectedCol {
				expectedFloat, err1 := strconv.ParseFloat(fmt.Sprintf("%v", expectedCol), 64)
				actualFloat, err2 := strconv.ParseFloat(fmt.Sprintf("%v", actualCol), 64)

				// account for minute floating point imprecision
				if err1 == nil && err2 == nil {

					prec := getPrecision(expectedFloat)
					fmtS := fmt.Sprintf("%%.%df", prec)
					floatS := fmt.Sprintf(fmtS, actualFloat)
					actualFloat, _ = strconv.ParseFloat(floatS, 64)

					// default tolerance is 0.0000000001
					if math.Abs(actualFloat-expectedFloat) > 0.00000001 {
						return fmt.Errorf("Expected %v, got %v at row %d, column %d", expectedFloat, actualFloat, rownum, colnum)
					}
				} else {
					if fmt.Sprintf("(%d,%d): %v", rownum, colnum, actualCol) !=
						fmt.Sprintf("(%d,%d): %v", rownum, colnum, expectedCol) {
						return fmt.Errorf("Expected %v, got %v at row %d, column %d", expectedCol, actualCol, rownum, colnum)
					}
				}
			}

		}
	}

	return nil
}

func getPrecision(num float64) int {
	s := fmt.Sprintf("%v", num)
	i := strings.Index(s, ".")
	if i == -1 {
		return 0
	}

	return len(s[i+1:])
}

// RunTest runs the provided integration test using the provided database
// handle, returning any error encountered during query execution. If the query
// returns a result, then the error returned will be nil (regardless of whether
// the test passed).
func RunTest(t *testing.T, test *TestCase, db *sql.DB) error {
	query := test.SQL

	if test.VerificationSQL != "" {
		_, err := db.Exec(query)
		if err != nil {
			return err
		}

		query = test.VerificationSQL
	}

	results, err := RunSQL(db, query, test.ExpectedTypes, test.ExpectedNames)
	if test.ExpectedError != "" {
		if err == nil {
			return fmt.Errorf("expected error, but query executed successfully")
		}
		if err.Error() != test.ExpectedError {
			return fmt.Errorf("expected error '%s', got '%v'", test.ExpectedError, err.Error())
		}
	} else if err != nil {
		return err
	}

	if test.CleanupSQL != "" {
		_, err := db.Exec(test.CleanupSQL)
		if err != nil {
			return err
		}
	}

	err = CompareResults(test.ExpectedData, results)
	if err != nil {
		return err
	}

	return nil
}

// RunIntegrationSuite performs any necessary setup for the suite with the
// provided name, and the runs all the tests in that suite as subtests.
func RunIntegrationSuite(t *testing.T, name string) {
	// we do not run suites in parallel to avoid races
	// when restoring suite data
	suite, err := setupIntegrationSuite(name)
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range suite.Tests {
		t.Run(test.Name, func(t *testing.T) {
			runIntegrationTest(t, test)
		})
	}
}

func runIntegrationTest(t *testing.T, test *TestCase) {
	t.Parallel()

	if test.Skip {
		if *RunSkipped {
			t.Log("Running test with skip=true")
		} else {
			t.Skip("Skipping test with skip=true")
		}
	}

	noPushDownMode := os.Getenv(evaluator.NoPushDown) != ""
	if test.PushDownOnly && noPushDownMode {
		t.Skip("Skipping pushdown-only test in pushdown mode")
	}

	if !mongodb.VersionAtLeast(test.MinServerVersion) {
		t.Skipf("Skipping test with min_server_version=%v against MongoDB %v", test.MinServerVersion, *mongodb.ServerVersion)
	}

	dbName := test.Database

	compressionVal := ""
	if *DriverCompression {
		compressionVal = "&compress=1"
	}

	connString := fmt.Sprintf("root@tcp(%v)/%v?allowNativePasswords=1%v", *DbAddr, dbName, compressionVal)
	db, err := sql.Open("mysql", connString)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer db.Close()

	err = RunTest(t, test, db)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func setupIntegrationSuite(suite string) (*TestSuite, error) {
	automate := *Automate
	if automate == "data" {
		fmt.Printf(">> Restoring %s data...\n", suite)
		err := RestoreSuiteData(suite)
		if err != nil {
			return nil, err
		}
		fmt.Printf(">> Done restoring %s data\n", suite)
	}

	fmt.Printf(">> Loading %s test suite...\n", suite)
	tests, err := LoadTestSuite(suite)
	if err != nil {
		return nil, err
	}
	fmt.Printf(">> Done loading %s test suite\n", suite)

	return tests, nil
}
