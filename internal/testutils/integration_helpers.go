package testutils

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/schema"
)

var maxTime = time.Duration(*MaxTimeSecs) * time.Second

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

func MongodbVersionAtLeast(versionString string) bool {
	if versionString == "" {
		return true
	}

	strServerVersion := strings.Split(*ServerVersion, ".")
	serverVersion := make([]uint8, len(strServerVersion))
	for i, str := range strServerVersion {
		num, err := strconv.ParseInt(str, 0, 0)
		if err != nil {
			panic(err)
		}
		serverVersion[i] = uint8(num)
	}

	strVersion := strings.Split(versionString, ".")
	version := make([]uint8, len(strVersion))
	for i, str := range strVersion {
		num, err := strconv.ParseInt(str, 0, 0)
		if err != nil {
			panic(err)
		}
		version[i] = uint8(num)
	}

	return util.VersionAtLeast(serverVersion, version)
}

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
	if err != nil {
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

func RunBenchmark(t *testing.B, test *TestCase, db *sql.DB) error {
	query := test.SQL

	db.SetMaxOpenConns(1)
	var connId int
	r := db.QueryRow("select connection_id()")
	err := r.Scan(&connId)
	if err != nil {
		return err
	}

	t.ResetTimer()

	for n := 0; n < t.N; n++ {
		timer := time.NewTimer(maxTime)
		done := false
		killed := false

		go func(done, killed *bool) {
			<-timer.C
			if !*done {
				db.SetMaxOpenConns(2)
				t.Logf("killing query (ran longer than %s)", maxTime)
				kill := fmt.Sprintf("kill query %d", connId)
				db.Query(kill)
				*killed = true
			}
		}(&done, &killed)

		rows, err := db.Query(query)
		done = true
		if err != nil {
			if !killed {
				return err
			}
			continue
		}

		for rows.Next() {
		}
		rows.Close()
	}

	return nil
}
