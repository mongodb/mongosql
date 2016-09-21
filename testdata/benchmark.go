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
	input  = flag.String("in", "testdata/benchmark.json", "File containing SQL queries")
	output = flag.String("out", "", "File to write SQL benchmark functions (defaults to stdout)")
)

type Query struct {
	Query    string `json:"query"`
	ID       string `json:"id"`
	Database string `json:"db"`
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
			panic(fmt.Errorf("decode: %v", err))
		}
		if q.ID == "" {
			panic(fmt.Errorf("empty id"))
		}
		if q.Database == "" {
			panic(fmt.Errorf("empty database"))
		}
		if q.Query == "" {
			panic(fmt.Errorf("empty query"))
		}
		queries.Q = append(queries.Q, q)
	}

	writeBenchmarkTestFile(fo, queries)
}

func templateBenchmark(content string, data interface{}, out io.Writer) {
	tmpl, err := template.New("").Parse(content)
	if err != nil {
		panic(err)
	}

	tmpl.Execute(out, data)
}

func writeBenchmarkTestFile(out io.Writer, queries Queries) {
	content := `package sqlproxy_test

import (
	"database/sql"
	"fmt"
	"flag"
	"github.com/10gen/sqlproxy"
	"github.com/10gen/sqlproxy/schema"
	_ "github.com/go-sql-driver/mysql"
	"sync"
	"testing"
)

var (
	db1              *sql.DB
	db2              *sql.DB
	once sync.Once
	testMongoDBHost  = flag.String("testMongoDBHost", "127.0.0.1", "test mongod host")
	testMongoDBPort  = flag.String("testMongoDBPort", "27017", "test mongod port")
	testProxyAddress = flag.String("testProxyAddress", "127.0.0.1", "test proxy host")
	testProxyPort    = flag.String("testProxyPort", "3308", "test proxy address")
	testProxyDB1     = flag.String("testProxyDB1", "fullblackbox", "test proxy database 1")
	testProxyDB2     = flag.String("testProxyDB2", "tableau", "test proxy database 2")
	testConfig       = flag.String("conf", "testdata/benchmark.yml", "test proxy DRDL file")
)


func getDB(dbName string) *sql.DB {
	if dbName == *testProxyDB1 {
		return db1
	}
    return db2
}

func setup() {
	config := mustLoadTestSchema(*testConfig)
	mustLoadTestData(*testMongoDBHost, *testMongoDBPort, config)

	cfg := &schema.Schema{
		RawDatabases: config.Databases,
	}

	opts, _ := sqlproxy.NewSqldOptions()
	opts.Addr = fmt.Sprintf("%v:%v", *testProxyAddress, *testProxyPort)
	opts.MongoURI = fmt.Sprintf("mongodb://%v:%v", *testMongoDBHost, *testMongoDBPort)
 
	buildSchemaMaps(cfg)

	s, err := testServer(cfg, opts)
	if err != nil {
	    panic(err)
	}
	go s.Run()

	db1, err = sql.Open("mysql", fmt.Sprintf("root@tcp(%v:%v)/%v", *testProxyAddress, *testProxyPort, *testProxyDB1))
	if err != nil {
	    panic(err)
	}

	db2, err = sql.Open("mysql", fmt.Sprintf("root@tcp(%v:%v)/%v", *testProxyAddress, *testProxyPort, *testProxyDB2))
	if err != nil {
	    panic(err)
	}
}

func runQuery(db *sql.DB, query string) error {
	rows, err := db.Query(query)
	if err != nil {
	    return err
	}

	for rows.Next() {	}

	rows.Close()

	return rows.Err()
}

{{range .Q}}
func BenchmarkQuery{{.ID}}(b *testing.B) {
    query := "{{.Query}}"
    once.Do(setup)
    db := getDB("{{.Database}}")
    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        if err := runQuery(db, query); err != nil {
            b.Fatalf("query error: %v", err)
            // This still won't result in a non-zero exit
            // code because of this bug:
            // https://github.com/golang/go/issues/14307
        }
    }
}
{{end}}
`
	templateBenchmark(content, queries, out)
}
