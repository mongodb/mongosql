package main

import (
	"bytes"
	"flag"
	"go/format"
	"io/ioutil"
	"strings"
	"text/template"

	"github.com/10gen/sqlproxy/internal/testutils"
)

var (
	suites = flag.String("suites", "blackbox,integration,tableau", "comma-separated list of suites to generate tests for")
)

func main() {
	flag.Parse()
	suites := strings.Split(*suites, ",")
	generateSuites(suites...)
}

func generateSuites(names ...string) {
	suites := make([]*testutils.TestSuite, 0)
	for _, name := range names {
		suites = append(suites, testutils.LoadTestSuite(name))
	}

	tmpl, err := template.New("").Parse(testTemplate)
	if err != nil {
		panic(err)
	}

	buf := bytes.NewBuffer([]byte{})
	err = tmpl.Execute(buf, suites)
	if err != nil {
		panic(err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile("integration_test.go", formatted, 0644)
}

var testTemplate = `package sqlproxy_test

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/testutils"
	_ "github.com/go-sql-driver/mysql"
)

var testCasesByName map[string]*testutils.TestCase
var maxTime = time.Duration(*testutils.MaxTimeSecs) * time.Second


func setup() {
	fmt.Println("Starting global setup...")

	testCasesByName = make(map[string]*testutils.TestCase)

	if *testutils.RestoreData != "" {
		suitesToRestore := strings.Split(*testutils.RestoreData, ",")
		for _, suite := range suitesToRestore {
			fmt.Printf("Restoring %s data...\n", suite)
			testutils.LoadSuiteData(suite)
			fmt.Printf("Done restoring %s data\n", suite)
		}
	}

	fmt.Println("Global setup done")
}

{{ range . }}
var once{{ .Name }} sync.Once
func setup{{ .Name }}() {
	fmt.Println("Starting {{ .Name }} setup...")
	suite := testutils.LoadTestSuite("{{ .Dirname }}")
	for _, test := range suite.TestCases {
		testCasesByName[test.Name] = test
	}
	fmt.Println("{{ .Name }} setup done\nRunning {{ .Name }} tests...")
}
{{ end }}

func TestMain(m *testing.M) {
	flag.Parse()
	setup()
	exitCode := m.Run()
	os.Exit(exitCode)
}

{{ range . }}
{{ $parent := . }}
{{ range .TestCases }}
func {{ .Name }}(t *testing.{{ if .IsBenchmark }}B{{ else }}T{{ end }}) {
	{{ if not .IsBenchmark }}
	t.Parallel()
	{{ end }}

	once{{ $parent.Name }}.Do(setup{{ $parent.Name }})

	test := testCasesByName["{{ .Name }}"]

	if test.Skip {
		if *testutils.RunSkipped {
			t.Log("Running test with skip=true")
		} else {
			t.Skip("Skipping test with skip=true")
		}
	}

	noPushDownMode := os.Getenv(evaluator.NoPushDown) != ""
	if test.PushDownOnly && noPushDownMode {
		t.Skip("Skipping pushdown-only test in pushdown mode")
	}

	if !testutils.MongodbVersionAtLeast(test.MinServerVersion) {
		t.Skipf("Skipping test with min_server_version=%v against MongoDB %v", test.MinServerVersion, *testutils.ServerVersion)
	}

	dbName := test.Database
	if dbName == "" {
		dbName = "{{ $parent.DefaultDb }}"
	}


	compressionVal := ""
	if *testutils.DriverCompression {
		compressionVal = "&compress=1"
	}

	connString := fmt.Sprintf("root@tcp(%v)/%v?allowNativePasswords=1%v", *testutils.DbAddr, dbName, compressionVal)
	db, err := sql.Open("mysql", connString)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer db.Close()

	{{ if not .IsBenchmark }}

	err = testutils.RunTest(t, test, db)

	{{ else }}

	err = testutils.RunBenchmark(t, test, db)

	{{ end }}

	if err != nil {
		t.Fatal(err.Error())
	}

}
{{ end }}

{{ end }}
`
