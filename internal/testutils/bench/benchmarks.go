package bench

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"path"
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/config"
	testutils "github.com/10gen/sqlproxy/internal/testutils/integration"
	mongoutil "github.com/10gen/sqlproxy/internal/testutils/mongodb"
	"github.com/10gen/sqlproxy/internal/testutils/translator"
	"github.com/10gen/sqlproxy/mongodb"
	yaml "gopkg.in/yaml.v2"
)

// Suite represents a set of benchmarks.
type Suite struct {
	Queries  []*Benchmark
	Overhead []*Benchmark
}

// Benchmark represents an individual benchmark, including the query to be run,
// the database against which to run it, and the data that it expectes to run
// against.
type Benchmark struct {
	Name  string `yaml:"name"`
	Db    string `yaml:"db"`
	Query string `yaml:"query"`
	Type  string `yaml:"type"`
}

// BenchmarkQuery restores the data specified in the provided benchmark,
// updates the schema, and benchmarks the execution of the query.
func BenchmarkQuery(b *testing.B, bench *Benchmark) {
	dbName := "benchmark"
	if bench.Db != "" {
		dbName = bench.Db
	}

	err := restoreBenchmarkData(bench.Name)
	if err != nil {
		b.Fatal(err)
	}

	err = flushSample()
	if err != nil {
		b.Fatal(err)
	}

	connString := fmt.Sprintf("root@tcp(%v)/%v?allowNativePasswords=1", *testutils.DbAddr, dbName)
	db, err := sql.Open("mysql", connString)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	runBenchmark(b, db, bench.Query)
}

// BenchmarkQueryPipeline restores the data specified in the provided benchmark,
// gets the MongoDB aggregation pipeline to which the benchmark query is
// translated, and benchmarks the execution of that pipeline.
func BenchmarkQueryPipeline(b *testing.B, bench *Benchmark) {
	dbName := "benchmark"
	if bench.Db != "" {
		dbName = bench.Db
	}

	err := restoreBenchmarkData(bench.Name)
	if err != nil {
		b.Fatal(err)
	}

	err = flushSample()
	if err != nil {
		b.Fatal(err)
	}

	sp, err := mongodb.NewSqldSessionProvider(config.Default())
	if err != nil {
		b.Fatal(err)
	}

	pipeline, coll, err := getPipeline(dbName, bench.Query, sp)
	if err != nil {
		b.Fatalf("could not get pipeline for benchmark query: %v", err)
	}

	s, err := sp.Session(context.Background())
	if err != nil {
		b.Fatal(err)
	}

	runAggBenchmark(b, s, dbName, coll, pipeline)
}

func getPipeline(db, query string, sp *mongodb.SessionProvider) ([]bson.D, string, error) {
	opts := &config.SchemaSampleOptions{
		Source:               "mongosqld_sample_test",
		UUIDSubtype3Encoding: "old",
	}

	tr, err := translator.NewTranslator(opts, sp)
	if err != nil {
		return nil, "", err
	}

	return tr.TranslateQuery(db, query)
}

func restoreBenchmarkData(name string) error {
	dataset := getDatasetForBenchmark(name)
	return dataset.Restore(mongoutil.GetToolOptions())
}

func flushSample() error {
	connString := fmt.Sprintf("root@tcp(%v)/information_schema?allowNativePasswords=1", *testutils.DbAddr)

	db, err := sql.Open("mysql", connString)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Query("flush sample")
	return err
}

func runAggBenchmark(b *testing.B, session *mongodb.Session, db, coll string, pipeline []bson.D) {
	for n := 0; n < b.N; n++ {
		iter, err := session.Aggregate(db, coll, pipeline)
		if err != nil {
			b.Fatal(err)
		}

		doc := &bson.D{}
		for iter.Next(context.Background(), doc) {
			// do nothing
		}

		iter.Close(context.Background())
	}
}

func runBenchmark(b *testing.B, db *sql.DB, query string) {
	for n := 0; n < b.N; n++ {
		rows, err := db.Query(query)
		if err != nil {
			b.Fatal(err)
		}

		for rows.Next() {
		}
		rows.Close()
	}
}

// LoadBenchmarks reads all of the yaml files in testdata/benchmarks/ and
// returns a benchmark suite that inclues all of the benchmarks defined
// therein.
func LoadBenchmarks() (*Suite, error) {
	files, err := ioutil.ReadDir("testdata/benchmarks/")
	if err != nil {
		return nil, err
	}

	suite := &Suite{}

	for _, f := range files {
		more, err := loadFile(f.Name())
		if err != nil {
			return nil, err
		}

		for _, b := range more {
			switch b.Type {
			case "", "query":
				b.Type = "query"
				suite.Queries = append(suite.Queries, b)
			case "overhead":
				suite.Overhead = append(suite.Overhead, b)
			default:
				return nil, fmt.Errorf("Unknown benchmark type %s", b.Type)
			}
		}
	}

	return suite, nil
}

func loadFile(basename string) ([]*Benchmark, error) {
	benchFile := path.Join("testdata", "benchmarks", basename)

	fileBytes, err := ioutil.ReadFile(benchFile)
	if err != nil {
		return nil, err
	}

	data := struct {
		Benchmarks []*Benchmark `yaml:"benchmarks"`
	}{}

	err = yaml.Unmarshal(fileBytes, &data)
	if err != nil {
		return nil, err
	}

	return data.Benchmarks, nil
}
