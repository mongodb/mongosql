package strutil_test

import (
	"bytes"
	"testing"

	"github.com/10gen/sqlproxy/internal/strutil"
	"github.com/stretchr/testify/require"
)

func TestString(t *testing.T) {
	b := []byte("hello world")
	a := strutil.String(b)

	if a != "hello world" {
		t.Fatal(a)
	}

	b[0] = 'a'

	if a != "aello world" {
		t.Fatal(a)
	}

	if a != "aello world" {
		t.Fatal(a)
	}
}

func TestByte(t *testing.T) {
	a := "hello world"

	b := strutil.Slice(a)

	if !bytes.Equal(b, []byte("hello world")) {
		t.Fatal(string(b))
	}
}

func TestCleanNumericString(t *testing.T) {
	type test struct {
		input, output string
	}

	runTests := func(tests []test) {
		for _, test := range tests {
			t.Run(test.input, func(t *testing.T) {
				output := strutil.MySQLCleanNumericString(test.input)
				require.Equal(t, test.output, output)
			})
		}
	}
	tests := []test{
		{"     -12345.1234.34xwwyzz   :", "-12345.1234"},
		{"    - 12345.1234.34xwwyzz   :", "0"},
		{"1234", "1234"},
		{"  1234  ", "1234"},
		{"   -3.14159265xyz", "-3.14159265"},
		{" Hello World  ", "0"},
		{"   ", "0"},
		{"", "0"},
		{"1.2.3.4", "1.2"},
	}

	runTests(tests)
}
