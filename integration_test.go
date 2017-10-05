package sqlproxy_test

import (
	"flag"
	"io/ioutil"
	"testing"

	"github.com/10gen/sqlproxy/internal/testutils"
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
			testutils.RunIntegrationSuite(t, suite)
		})
	}
}

func BenchmarkIntegration(b *testing.B) {
	suiteDirs, err := ioutil.ReadDir("testdata/suites/")
	if err != nil {
		b.Fatal(err)
	}

	for _, dir := range suiteDirs {
		suite := dir.Name()
		b.Run(suite, func(b *testing.B) {
			testutils.BenchmarkIntegrationSuite(b, suite)
		})
	}
}
