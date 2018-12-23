package mongodrdl_test

import (
	"os"
	"strings"
	"testing"

	"github.com/10gen/sqlproxy/internal/mongodrdl"
)

func TestParseArgs_Valid(t *testing.T) {

	test := func(args []string, expected string) {
		os.Args = args
		opts, err := mongodrdl.NewDrdlOptions()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = opts.Parse()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if opts.DrdlConnection.Host != expected {
			t.Fatalf("expected '%s', but got '%s'", expected, opts.DrdlConnection.Host)
		}
	}

	test(
		[]string{
			"mongodrdl",
			"--host", "localhost",
			"--port", "6999",
		},
		"localhost:6999",
	)

	test(
		[]string{
			"mongodrdl",
			"--host", "localhost:34325452",
			"--port", "6999",
		},
		"localhost:6999",
	)

	test(
		[]string{
			"mongodrdl",
			"--host", "localhost:6999",
		},
		"localhost:6999",
	)

	test(
		[]string{
			"mongodrdl",
			"--host", "localhost",
			"--port", "6999",
			"--ssl",
			"--sslCAFile", "hello",
		},
		"localhost:6999",
	)
}

func TestParseSSL_Invalid(t *testing.T) {

	test := func(args []string) {
		os.Args = args
		opts, err := mongodrdl.NewDrdlOptions()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = opts.Parse()
		if err != nil {
			t.Fatalf("expected parse error %v got none", err)
		}

		expectedErr := "when specifying SSL options, SSL must be enabled with --ssl"
		err = opts.Validate()
		if err == nil {
			t.Fatalf("expected SSL related error for %q but got none", strings.Join(args, " "))
		} else if err.Error() != expectedErr {
			t.Fatalf("expected SSL related error for %q got %v", strings.Join(args, " "), err)
		}
	}

	test(
		[]string{
			"mongodrdl",
			"--host", "localhost",
			"--port", "6999",
			"--sslCAFile", "hello",
			"-d", "output",
		},
	)

	test(
		[]string{
			"mongodrdl",
			"--host", "localhost",
			"--port", "6999",
			"--sslPEMKeyFile", "hello",
			"-d", "output",
		},
	)

	test(
		[]string{
			"mongodrdl",
			"--host", "localhost",
			"--port", "6999",
			"--sslPEMKeyPassword", "hello",
			"-d", "output",
		},
	)

	test(
		[]string{
			"mongodrdl",
			"--host", "localhost",
			"--port", "6999",
			"--sslCRLFile", "hello",
			"-d", "output",
		},
	)

	test(
		[]string{
			"mongodrdl",
			"--host", "localhost",
			"--port", "6999",
			"--sslAllowInvalidCertificates", "hello",
			"-d", "output",
		},
	)

}
