package util_test

import (
	"fmt"
	"testing"

	. "github.com/10gen/sqlproxy/internal/util"

	"github.com/stretchr/testify/require"
)

func TestParseConnectionString(t *testing.T) {

	type test struct {
		connString string
		hosts      []string
		setName    string
	}

	runTests := func(tests []test) {
		for _, test := range tests {
			testCase := fmt.Sprintf("parse_connection_string_%s", test.connString)
			t.Run(testCase, func(t *testing.T) {
				req := require.New(t)
				hosts, setName := ParseConnectionString(test.connString)
				req.ElementsMatch(test.hosts, hosts, "actual hosts did not match expected hosts")
				req.Equal(test.setName, setName, "expected set name did not match actual")
			})
		}
	}

	tests := []test{
		{"", []string{""}, ""},
		{"host1,host2", []string{"host1", "host2"}, ""},
		{"foo/host1,host2", []string{"host1", "host2"}, "foo"},
	}

	runTests(tests)

}

func TestValidateDBName(t *testing.T) {

	type test struct {
		database    string
		shouldError bool
	}

	runTests := func(tests []test) {
		for _, test := range tests {
			testCase := fmt.Sprintf("validate_database_name_%s", test.database)
			t.Run(testCase, func(t *testing.T) {
				req := require.New(t)
				err := ValidateDBName(test.database)
				if test.shouldError {
					req.NotNil(err, "expected error but got no error")
				} else {
					req.Nil(err, "expected no error but got error")
				}
			})
		}
	}

	tests := []test{
		{"test", false},
		{"db/aaa", true},
		{"db spac", true},
		{"db.spac", true},
		{"x$x", true},
		{"\x00", true},
		{" ", true},
		{"", false},
		{"db", false},
	}

	runTests(tests)
}
