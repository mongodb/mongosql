package sqlproxy_test

import (
	"flag"
	"io/ioutil"
	"testing"

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
