package schema_test

import (
	"strings"
	"testing"

	"github.com/10gen/sqlproxy/schema"
	"github.com/stretchr/testify/require"
)

func TestNewIndex(t *testing.T) {
	runTest := func(name string) {
		t.Run(name, func(t *testing.T) {
			req := require.New(t)

			index := schema.NewIndex(name, false, false, []schema.IndexPart{})
			req.Equal(strings.ToLower(index.MongoName()), index.SQLName(), "SQLName should be a lowercased MongoName")
		})
	}

	for _, name := range []string{
		"foo",
		"fOo",
		"FoO",
		"FOO",
	} {
		runTest(name)
	}
}
