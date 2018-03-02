package options_test

import (
	"os"
	"testing"

	"github.com/10gen/sqlproxy/options"
)

func TestParseArgs_Valid(t *testing.T) {

	test := func(args []string, expected string) {
		os.Args = args
		opts, err := options.NewDrdlOptions()
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

}
