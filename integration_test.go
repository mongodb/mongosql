package sqlproxy_test

import (
	"flag"
	"io/ioutil"
	"testing"

	"github.com/10gen/sqlproxy/internal/testutils/bench"
	util "github.com/10gen/sqlproxy/internal/testutils/integration"
)

func init() {
	flag.Parse()
}

func TestIntegration(t *testing.T) {
	t.Parallel()

	suiteDirs, err := ioutil.ReadDir("testdata/suites/")
	if err != nil {
		t.Fatal(err)
	}

	for _, dir := range suiteDirs {
		suite := dir.Name()
		t.Run(suite, func(t *testing.T) {
			util.RunIntegrationSuite(t, suite)
		})
	}
}

func BenchmarkIntegration(b *testing.B) {
	benchSuite, err := bench.LoadBenchmarks()
	if err != nil {
		b.Fatal(err)
	}

	b.Run("queries", func(b *testing.B) {
		for _, query := range benchSuite.Queries {
			b.Run(query.Name, func(b *testing.B) {
				bench.BenchmarkQuery(b, query)
			})
		}
	})

	b.Run("overhead", func(b *testing.B) {
		for _, query := range benchSuite.Overhead {
			b.Run(query.Name, func(b *testing.B) {
				b.Run("query", func(b *testing.B) { bench.BenchmarkQuery(b, query) })
				b.Run("pipeline", func(b *testing.B) { bench.BenchmarkQueryPipeline(b, query) })
			})
		}
	})
}
