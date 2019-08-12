package procutil_test

import (
	"fmt"
	"testing"

	. "github.com/10gen/sqlproxy/internal/procutil"

	"github.com/stretchr/testify/require"
)

func TestParseHost(t *testing.T) {

	type test struct {
		description          string
		host                 string
		port                 string
		preferPort           bool
		expectedURI          string
		expectedReplacedPort bool
	}

	runTests := func(tests []test) {
		for _, test := range tests {
			t.Run(test.description, func(t *testing.T) {
				req := require.New(t)
				actualURI, actualReplacedPort := ParseHost(test.host, test.port, test.preferPort)
				req.Equal(test.expectedURI, actualURI, "incorrect URI")
				req.Equal(test.expectedReplacedPort, actualReplacedPort, "incorrect replacedPort")
			})
		}
	}

	tests := []test{
		{"no host or port", "", "", false, "mongodb://localhost:27017", false},
		{"port only", "", "27018", false, "mongodb://localhost:27018", false},
		{"host without port only", "127.0.0.1", "", false, "mongodb://127.0.0.1", false},
		{"host with port only", "8.8.8.8:27000", "", false, "mongodb://8.8.8.8:27000", false},
		{"host without port and port and preferPort false", "localhost", "27019", false, "mongodb://localhost:27019", false},
		{"host without port and port and preferPort true", "localhost", "27019", true, "mongodb://localhost:27019", false},
		{"host with port and port and preferPort false", "localhost:27018", "27019", false, "mongodb://localhost:27018", false},
		{"host with port and port and preferPort true", "localhost:27018", "27019", true, "mongodb://localhost:27019", true},
		{"host with scheme without port only", "mongodb+srv://1.2.3.4", "", false, "mongodb+srv://1.2.3.4", false},
		{"host with scheme with port only", "mongodb+srv://1.2.3.4:5678", "", false, "mongodb+srv://1.2.3.4:5678", false},
		{"host with scheme without port and port and preferPort false", "mongodb://1.2.3.4", "5678", false, "mongodb://1.2.3.4:5678", false},
		{"host with scheme without port and port and preferPort true", "mongodb://1.2.3.4", "5678", true, "mongodb://1.2.3.4:5678", false},
		{"host with scheme with port and port and preferPort false", "mongodb+srv://1.2.3.4:5678", "27017", false, "mongodb+srv://1.2.3.4:5678", false},
		{"host with scheme with port and port and preferPort true", "mongodb+srv://1.2.3.4:5678", "27017", true, "mongodb+srv://1.2.3.4:27017", true},
		{"replset seedlist host with scheme only", "mongodb://replSet/localhost:27017,localhost:27018,localhost:27019", "", false, "mongodb://localhost:27017,localhost:27018,localhost:27019/?replicaSet=replSet", false},
		{"replset seedlist host with scheme and port and preferPort false", "mongodb+srv://replSet/localhost:27017,localhost:27018,localhost:27019", "30000", false, "mongodb+srv://localhost:27017,localhost:27018,localhost:27019/?replicaSet=replSet", false},
		{"replset seedlist host with scheme and port and preferPort true", "mongodb://replSet/localhost:27017,localhost:27018,localhost:27019", "30000", true, "mongodb://localhost:27017,localhost:27018,localhost:27019/?replicaSet=replSet", false},
		{"replset seedlist host without scheme only", "replSet/1.2.3.4:27017,localhost:27018,8.8.8.8:27019", "", false, "mongodb://1.2.3.4:27017,localhost:27018,8.8.8.8:27019/?replicaSet=replSet", false},
		{"replset seedlist host without scheme and port and preferPort false", "replSet/localhost:27017,127.0.0.1:27018,localhost:27019", "30000", false, "mongodb://localhost:27017,127.0.0.1:27018,localhost:27019/?replicaSet=replSet", false},
		{"replset seedlist host without scheme and port and preferPort true", "replSet/localhost:27017,localhost:27018,localhost:27019", "30000", true, "mongodb://localhost:27017,localhost:27018,localhost:27019/?replicaSet=replSet", false},
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
