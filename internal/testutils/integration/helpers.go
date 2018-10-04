package integration

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/testutils/flags"
	"github.com/10gen/sqlproxy/internal/testutils/mongodb"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/mongodb/ssl"
	"github.com/10gen/sqlproxy/schema"
	"github.com/go-sql-driver/mysql"
	"github.com/shopspring/decimal"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/mongo-go-driver/mongo/connstring"
	"github.com/10gen/mongo-go-driver/mongo/private/cluster"
	"github.com/10gen/mongo-go-driver/mongo/private/conn"
	"github.com/10gen/mongo-go-driver/mongo/private/ops"
	"github.com/10gen/mongo-go-driver/mongo/private/server"
	"github.com/10gen/mongo-go-driver/mongo/readpref"
)

const (
	// SetNoPushDown is a query string that can be run to instruct the BI
	// Connector not to push down queries.
	setNoPushDown = "set @@optimize_push_down = false"
	// SetPushDown is a query string that can be run to restore
	// the variable optimize_push_down to true
	setPushDown = "set @@optimize_push_down = true"
)

// RunSQL runs the provided SQL query using the provided database handle.
// It expects the results to have the provided column names and types, and
// returns a list of rows and an error.
func RunSQL(conn *sql.Conn, q string, types []schema.SQLType,
	names []string) ([][]interface{}, error) {
	rows, err := conn.QueryContext(context.Background(), q)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	if len(types) > 0 && len(cols) != len(types) {
		err = fmt.Errorf(
			"Number of columns in result set (%v) does not match columns in expected types (%v)",
			len(cols), len(types),
		)
		return nil, err
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
		switch t {
		case schema.SQLVarchar:
			resultContainer = append(resultContainer, &sql.NullString{})
		case schema.SQLInt:
			resultContainer = append(resultContainer, &sql.NullInt64{})
		case schema.SQLFloat:
			resultContainer = append(resultContainer, &sql.NullFloat64{})
		case schema.SQLDecimal:
			resultContainer = append(resultContainer, &nullDecimal{})
		case schema.SQLDate:
			resultContainer = append(resultContainer, &mysql.NullTime{})
		default:
			return nil, fmt.Errorf("unknown result column type %q", t)
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

type nullDecimal struct {
	sql.NullString
}

func (d *nullDecimal) Value() (driver.Value, error) {
	if !d.Valid {
		return nil, nil
	}
	return decimal.NewFromString(d.String)
}

// CompareResults checks whether a one set of SQL results matches another,
// returning an error if they do not match.
func CompareResults(expected [][]interface{}, actual [][]interface{}) error {
	if len(actual) != len(expected) {
		return fmt.Errorf("Expected %d rows, got %d rows", len(expected), len(actual))
	}

	for rownum, row := range actual {
		err := compareRows(rownum, expected[rownum], row)
		if err != nil {
			return err
		}
	}
	return nil
}

func compareRows(rownum int, expectedRow []interface{}, actualRow []interface{}) error {
	if len(expectedRow) != len(actualRow) {
		return fmt.Errorf("expected %d columns, got %d", len(expectedRow), len(actualRow))
	}

	for i, actualVal := range actualRow {
		expectedVal := expectedRow[i]

		var err error
		switch typedExpected := expectedVal.(type) {
		case string:
			typedActual, ok := actualVal.(string)
			if !ok {
				err = fmt.Errorf("expected string, got a %T", actualVal)
			}
			if typedExpected != typedActual {
				err = fmt.Errorf("expected %v, got %v", typedExpected, typedActual)
			}

		case int64:
			typedActual, ok := actualVal.(int64)
			if !ok {
				err = fmt.Errorf("expected int64, got a %T", actualVal)
			}
			if typedExpected != typedActual {
				err = fmt.Errorf("expected %v, got %v", typedExpected, typedActual)
			}

		case time.Time:
			typedActual, ok := actualVal.(time.Time)
			if !ok {
				err = fmt.Errorf("expected time.Time, got a %T", actualVal)
			}
			if !typedExpected.Equal(typedActual) {
				err = fmt.Errorf("expected %v, got %v", typedExpected, typedActual)
			}

		case decimal.Decimal:
			typedActual, ok := actualVal.(decimal.Decimal)
			if !ok {
				err = fmt.Errorf("expected decimal.Decimal, got a %T", actualVal)
			}
			if !typedExpected.Equals(typedActual) {
				err = fmt.Errorf("expected %v, got %v", typedExpected, typedActual)
			}

		case float64:
			typedActual, ok := actualVal.(float64)
			if !ok {
				err = fmt.Errorf("expected float64, got a %T", actualVal)
			}
			if typedExpected != typedActual {
				err = fuzzyFloatEquals(typedExpected, typedActual)
			}
		case nil:
			if actualVal != nil {
				err = fmt.Errorf("expected NULL, got %v", actualVal)
			}
		}

		if err != nil {
			return fmt.Errorf("%v at row %d column %d", err, rownum, i)
		}
	}
	return nil
}

func fuzzyFloatEquals(expected, actual float64) error {
	if expected == actual {
		return nil
	}

	// account for minute floating point imprecision
	prec, err := getPrecision(expected)
	if err != nil {
		return fmt.Errorf("could not find precision for expected %v", expected)
	}
	// There are several places in the blackbox tests where we are using
	// prec 0 exepected values that end up being something lke 73.99999
	// and the code outputs 74, giving us an off by one error when there
	// is no real error. This will be cleaned up in BI-1743.
	if prec <= 0 && math.Abs(actual-expected) > 1 {
		return fmt.Errorf("expected %v, got %v with precision %d", expected, actual, prec)
	}

	precisionFormatString := fmt.Sprintf("%%.%df", prec)
	actualFormatted := fmt.Sprintf(precisionFormatString, actual)
	actual, err = strconv.ParseFloat(actualFormatted, 64)
	if err != nil {
		return err
	}

	// default tolerance is 0.0000000001
	if math.Abs(actual-expected) > 9.9*math.Pow(10, -float64(prec)) {
		return fmt.Errorf(
			"expected %v, got %v (difference of %v)",
			expected, actual,
			math.Abs(expected-actual),
		)
	}

	return nil
}

// UnorderedCompareResults checks whether one set of SQL results matches another
// modulo row order, returning an error if they do not match.
func UnorderedCompareResults(expected [][]interface{}, actual [][]interface{}) error {
	if len(actual) != len(expected) {
		return fmt.Errorf("expected %d rows, got %d rows", len(expected), len(actual))
	}
Outer:
	for _, row := range actual {
		for i, expectedRow := range expected {
			err := compareRows(i, expectedRow, row)
			if err == nil {
				expected = append(expected[:i], expected[i+1:]...)
				continue Outer
			}
		}
		return fmt.Errorf("unordered matching failed: expected %v, actual %v",
			expected, actual)
	}

	return nil
}

func getPrecision(num float64) (int, error) {
	s := fmt.Sprintf("%v", num)
	// If this is in scientific notation, we need to find precision differently.
	exponentAdjustment := 0
	var err error
	if strings.Contains(s, "e") {
		pieces := strings.Split(s, "e")
		// We need to adjust the precision based on the negation of the exponent.
		// e.g., the precision of 3.1415e3 is actually 1, not 4, and the precision
		// of 3.14e-3 is actually 5, not 2.
		exponentAdjustment, err = strconv.Atoi(pieces[1])
		if err != nil {
			return 0, err
		}
		exponentAdjustment *= -1.0
		// Adjust s to pieces[0] so we don't count e.g., "e-3" as part of our precision.
		s = pieces[0]
	}
	i := strings.Index(s, ".")
	if i == -1 {
		ret := 0 + exponentAdjustment
		// Precision should be at least 0.
		if ret < 0 {
			ret = 0
		}
		return ret, nil
	}

	ret := len(s[i+1:]) + exponentAdjustment
	// Precision should be at least 0.
	if ret < 0 {
		ret = 0
	}
	return ret, nil
}

// RunTest runs the provided integration test using the provided database
// handle, returning any error encountered during query execution. If the query
// returns a result, then the error returned will be nil (regardless of whether
// the test passed).
func RunTest(t *testing.T, test *TestCase, conn *sql.Conn) {
	_, err := conn.ExecContext(context.Background(), "SET @@type_conversion_mode = 'mysql'")
	if err != nil {
		t.Fatalf("failed to set type_conversion_mode to mysql: %v", err)
	}

	for name, value := range test.Variables {
		query := ""
		switch typedV := value.(type) {
		case string:
			query = fmt.Sprintf("SET @@%s = %q", name, typedV)
		case bool:
			b := 0
			if typedV {
				b = 1
			}
			query = fmt.Sprintf("SET @@%s = %v", name, b)
		default:
			query = fmt.Sprintf("SET @@%s = %v", name, typedV)
		}
		_, err = conn.ExecContext(context.Background(), query)
		if err != nil {
			t.Fatalf("failed to set session variable %q: %v", name, err)
		}
	}

	query := test.SQL

	if test.VerificationSQL != "" {
		_, err = conn.ExecContext(context.Background(), query)
		if err != nil {
			t.Fatal(err)
		}

		query = test.VerificationSQL
	}

	results, err := RunSQL(conn, query, test.ExpectedTypes, test.ExpectedNames)
	if test.ExpectedError != "" {
		if err == nil {
			t.Fatal(fmt.Errorf("expected error, but query executed successfully"))
		}
		if err.Error() != test.ExpectedError {
			t.Fatal(fmt.Errorf("expected error '%s', got '%v'", test.ExpectedError, err.Error()))
		}
	} else if err != nil {
		t.Fatal(err)
	}

	if test.CleanupSQL != "" {
		if _, err = conn.ExecContext(context.Background(), test.CleanupSQL); err != nil {
			t.Fatal(err)
		}
	}

	for name := range test.Variables {
		unsetQuery := fmt.Sprintf("SET @@%s = DEFAULT", name)
		_, setErr := conn.ExecContext(context.Background(), unsetQuery)
		if setErr != nil {
			t.Fatalf("failed to unset session variable %q: %v", name, err)
		}
	}

	if test.Unordered {
		err = UnorderedCompareResults(test.ExpectedData, results)
	} else {
		err = CompareResults(test.ExpectedData, results)
	}
	if err != nil {
		t.Fatal(err)
	}
}

// RunIntegrationSuite performs any necessary setup for the suite with the
// provided name, and the runs all the tests in that suite as subtests.
func RunIntegrationSuite(t *testing.T, name string) {
	// we do not run suites in parallel to avoid races
	// when restoring suite data
	suite := setupIntegrationSuite(t, name)
	serverVersion := getServerVersion(t)
	for _, test := range suite.Tests {
		t.Run(test.Name, func(t *testing.T) {
			runIntegrationTest(t, test, serverVersion)
		})
	}
}

func getServerVersion(t *testing.T) []uint8 {
	t.Log(">> Getting MongoDB server version...")
	uri := fmt.Sprintf("mongodb://%v:%v", *flags.Host, *flags.Port)

	// Get connection string from uri.
	cs, err := connstring.Parse(uri)
	if err != nil {
		t.Fatalf("unable to parse connection string: %v", err)
	}

	// Construct cluster options.
	clusterOpts := []cluster.Option{
		cluster.WithConnString(cs),
	}

	// Check if MongoDB has SSL enabled.
	requiresSSL := mongodb.GetToolOptions().SSL.UseSSL

	if requiresSSL {
		// Create dialer.
		var dialer func(ctx context.Context, dialer *net.Dialer,
			network, address string) (net.Conn, error)

		sslCfg := config.MongoDBNetSSL{
			Enabled:                  true,
			AllowInvalidCertificates: true,
			MinimumTLSVersion:        config.TLSv1_1,
			PEMKeyFile:               *flags.ClientPEMKeyFile,
		}

		var cfg config.Config
		cfg.MongoDB.Net.SSL = sslCfg

		dialer, err = ssl.SqldDialer(&cfg)
		if err != nil {
			t.Fatalf("unable to create a sqld dialer: %v", err)
		}

		// Append SSL options to clusterOpts.
		clusterOpts = append(clusterOpts,
			cluster.WithMoreServerOptions(
				server.WithMoreConnectionOptions(
					conn.WithDialer(dialer),
				),
			),
		)
	}

	c, err := cluster.New(clusterOpts...)
	if err != nil {
		t.Fatalf("unable to create a new cluster: %v", err)
	}

	ctx := context.Background()
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	s, err := c.SelectServer(timeoutCtx, cluster.WriteSelector(), readpref.Primary())
	if err != nil {
		t.Fatalf("%v: %v", err, c.Model().Servers[0].LastError)
	}

	// Run command to get buildInfo, which contains the server version.
	dbname := cs.Database
	if dbname == "" {
		dbname = "test"
	}
	cmd := bson.M{"buildInfo": 1}
	result := struct {
		Version string `bson:"version"`
	}{}
	err = ops.Run(
		ctx,
		&ops.SelectedServer{
			Server:   s,
			ReadPref: readpref.Primary(),
		},
		dbname,
		cmd,
		&result)

	if err != nil {
		t.Fatalf("Error querying for mongodb version: %v\n", err)
	}

	version := result.Version

	t.Logf(">> MongoDB server version is %v\n", version)
	serverVersion, err := util.VersionToSlice(version)
	if err != nil {
		t.Fatalf("Error converting version to slice: %v\n", err)
	}

	return serverVersion
}

func runIntegrationTest(t *testing.T, test *TestCase, serverVersion []uint8) {

	if test.Sync {
		t.Log("not running test in parallel: sync is set to true")
	} else {
		t.Parallel()
	}

	if test.Skip {
		if *flags.RunSkipped {
			t.Log("Running test with skip=true")
		} else {
			t.Skip("Skipping test with skip=true")
		}
	}
	{
		// Skip this test if it is not tagged with the command line
		// heuristics flag as one of its flags.
		heuristicFlag := strings.ToLower(*flags.SchemaMappingHeuristic)
		skip := true
		if heuristicFlag == "" {
			heuristicFlag = "drdl"
		}
		// if no heuristic is passed and there are no tags on the case, just don't
		// skip it. This keeps us from having to modify every single blackbox test,
		// for now.
		if heuristicFlag == "drdl" && len(test.SchemaMappingHeuristics) == 0 {
			skip = false
		} else {
			for _, testFlag := range test.SchemaMappingHeuristics {
				if heuristicFlag == strings.ToLower(testFlag) {
					skip = false
					break
				}
			}
		}

		if skip {
			t.Skipf(
				"skipping test as it was not flagged for this schema mapping heuristic: "+
					"mode: %s, test tags: %v",
				*flags.SchemaMappingHeuristic, test.SchemaMappingHeuristics)
		}
	}

	if test.MinServerVersion != "" {
		minRequiredVersion, err := util.VersionToSlice(test.MinServerVersion)
		if err != nil {
			t.Fatalf("error getting test min_server_version: %v", err)
		}

		if !util.VersionAtLeast(serverVersion, minRequiredVersion) {
			t.Skipf("skipping test with min_server_version=%v against MongoDB %v",
				test.MinServerVersion, serverVersion)
		}
	}

	if test.ExactServerVersion != "" {
		requiredVersion, err := util.VersionToSlice(test.ExactServerVersion)
		if err != nil {
			t.Fatalf("error getting test exact_server_version: %v", err)
		}

		if !util.VersionExactly(serverVersion, requiredVersion) {
			t.Skipf("skipping test with exact_server_version=%v against MongoDB %v",
				test.ExactServerVersion, serverVersion)
		}
	}

	dbName := test.Database

	compressionVal := ""
	if *flags.DriverCompression {
		compressionVal = "&compress=1"
	}

	connString := fmt.Sprintf("root@tcp(%v)/%v?allowNativePasswords=1%v",
		*flags.DbAddr, dbName, compressionVal)
	db, err := sql.Open("mysql", connString)
	if err != nil {
		t.Fatal(err.Error())
	}

	conn, err := db.Conn(context.Background())
	if err != nil {
		t.Fatalf("error opening new connection: %v", err)
	}

	noPushDownMode := *flags.NoPushdown

	if noPushDownMode && test.PushDownOnly {
		t.Skip("Skipping pushdown-only test in pushdown mode")
	} else if noPushDownMode {
		_, err := conn.ExecContext(context.Background(), setNoPushDown)
		if err != nil {
			t.Fatalf("error setting optimize_push_down: %v", err)
		}
	}

	defer func() {
		_, err := conn.ExecContext(context.Background(), setPushDown)
		if err != nil {
			t.Fatalf("error setting optimize_push_down: %v", err)
		}

		_ = conn.Close()
		_ = db.Close()
	}()

	RunTest(t, test, conn)
}

func setupIntegrationSuite(t *testing.T, suite string) *TestSuite {
	automate := *flags.Automate
	if automate == "data" {
		t.Logf(">> Restoring %s data...\n", suite)
		err := RestoreSuiteData(suite)
		if err != nil {
			t.Fatalf("error restoring data: %v\n", err)
		}
		t.Logf(">> Done restoring %s data\n", suite)
	}

	t.Logf(">> Loading %s test suite...\n", suite)
	tests, err := LoadTestSuite(suite)
	if err != nil {
		t.Fatalf("error loading test suite: %v\n", err)
	}

	t.Logf(">> Done loading %s test suite\n", suite)

	return tests
}
