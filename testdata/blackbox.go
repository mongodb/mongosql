package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"text/template"
)

var (
	input  = flag.String("in", "testdata/blackbox_queries.json", "File containing SQL queries")
	output = flag.String("out", "", "File to write SQL benchmark functions (defaults to stdout)")
)

type Query struct {
	Query            string `json:"query"`
	ID               string `json:"id"`
	Database         string `json:"db"`
	Columns          int    `json:"columns"`
	MinServerVersion string `json:"minServerVersion"`
}

type Queries struct {
	Q []Query
}

func main() {
	flag.Parse()

	r, err := os.Open(*input)
	if err != nil {
		log.Fatal(err)
	}

	d := json.NewDecoder(r)

	fo := os.Stdout
	if *output != "" {
		fo, err = os.Create(*output)
		if err != nil {
			panic(err)
		}

		defer func() {
			if err := fo.Close(); err != nil {
				panic(err)
			}
		}()
	}

	var queries Queries
	var q Query

	for {
		q = Query{}
		err = d.Decode(&q)
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		if q.ID == "" {
			panic(fmt.Errorf("empty id"))
		}
		if q.Query == "" {
			panic(fmt.Errorf("empty query"))
		}
		queries.Q = append(queries.Q, q)
	}

	writeBlackBoxTestFile(fo, queries)
}

func templateBlackBox(content string, data interface{}, out io.Writer) {
	tmpl, err := template.New("").Parse(content)
	if err != nil {
		panic(err)
	}

	tmpl.Execute(out, data)
}

func writeBlackBoxTestFile(out io.Writer, queries Queries) {
	content := `package sqlproxy_test

import (
	"database/sql"
	"fmt"
	yaml "github.com/10gen/candiedyaml"
	_ "github.com/go-sql-driver/mysql"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"path/filepath"
	"sync"
	"testing"
)

var (
	once             sync.Once
	db               *sql.DB
)

type resultSchema struct {
	ExpectedNames []string        ` + "`yaml:\"expected_names\"`" + `
	ExpectedTypes []string        ` + "`yaml:\"expected_types\"`" + `
	ExpectedData  [][]interface{} ` + "`yaml:\"expected\"`" + `
}

func setup() {
	_, _, db = prepareForTestCase(filepath.Join("testdata", "blackbox.yml"))
}

func runBlackboxQuery(db *sql.DB, query string, id int) {

	expectedFile := filepath.Join("testdata", "results", fmt.Sprintf("%v.yml", id))

	fileBytes, err := ioutil.ReadFile(expectedFile)
	if err != nil {
		panic(err)
	}

	var conf resultSchema
	err = yaml.Unmarshal(fileBytes, &conf)
	So(err, ShouldBeNil)

	actual, err := runSQL(db, query, conf.ExpectedTypes, conf.ExpectedNames)
	So(err, ShouldBeNil)

	compareResults(conf.ExpectedData, actual)
}

{{range .Q}}
func TestBlackBox{{.ID}}(t *testing.T) {
	t.Parallel()
	query := "{{.Query}}"
	once.Do(setup)
	id := {{.ID}}
	minServerVersion := "{{.MinServerVersion}}"

	if !mongodbVersionAtLeast(minServerVersion) {
		t.Skip(fmt.Sprintf("This test requires mongodb >= %s, but serverVersion is %s", minServerVersion, *serverVersion))
	}

	Convey(fmt.Sprintf("Testing blackbox query %v", id), t, func() {
		runBlackboxQuery(db, query, id)
	})

}
{{end}}
`
	templateBlackBox(content, queries, out)
}
