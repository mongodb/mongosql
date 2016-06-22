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
	Query    string `json:"query"`
	ID       string `json:"id"`
	Database string `json:"db"`
	Columns  int    `json:"columns"`
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
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/10gen/sqlproxy"
	"github.com/10gen/sqlproxy/schema"
	_ "github.com/go-sql-driver/mysql"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"sync"
	"testing"
)

var (
	once             sync.Once
	db               *sql.DB
	testProxyAddress = flag.String("testProxyAddress", "127.0.0.1", "test proxy host")
	testProxyPort    = flag.String("testProxyPort", "3308", "test proxy address")
)

func setup() {
	conf := mustLoadTestSchema(pathify("testdata", "blackbox.yml"))
	mustLoadTestData(testMongoHost, testMongoPort, conf)

	opts := sqlproxy.Options{
		Addr:     testDBAddr,
		MongoURI: fmt.Sprintf("mongodb://%v:%v", testMongoHost, testMongoPort),
	}
	cfg := &schema.Schema{
		RawDatabases: conf.Databases,
	}
	buildSchemaMaps(cfg)
	s, err := testServer(cfg, opts)
	if err != nil {
		panic(err)
	}
	go s.Run()

	db, err = sql.Open("mysql", fmt.Sprintf("root@tcp(%v:%v)/%v", *testProxyAddress, *testProxyPort, conf.Databases[0].Name))
	if err != nil {
		panic(err)
	}
}

func runBlackboxQuery(t *testing.T, db *sql.DB, query string, columns, id int) {
	var types []string

	for j := 0; j < columns; j++ {
		types = append(types, schema.SQLVarchar)
	}

	actual, err := runSQL(db, query, types, nil)
	So(err, ShouldBeNil)

	expectedFile := pathify("testdata", "results", fmt.Sprintf("%v.csv", id))

	handle, err := os.OpenFile(expectedFile, os.O_RDONLY, os.ModeExclusive)
	if err != nil {
		panic(err)
	}

	r := csv.NewReader(handle)

	r.FieldsPerRecord = columns
	data, err := r.ReadAll()
	So(err, ShouldBeNil)
		
	// TODO: check header as well
	data = data[1:]
	expected := make([][]interface{}, len(data))
	for i, v := range data {
		for _, s := range v {
			expected[i] = append(expected[i], s)
		}
	}

	compareResults(t, expected, actual)
}

{{range .Q}}
func TestBlackBox{{.ID}}(t *testing.T) {
	t.Parallel()
	query := "{{.Query}}"
	once.Do(setup)
	columns := {{.Columns}}
	id := {{.ID}}

	Convey(fmt.Sprintf("Testing blackbox query %v", id), t, func() {
		runBlackboxQuery(t, db, query, columns, id)
	})

}
{{end}}
`
	templateBlackBox(content, queries, out)
}
