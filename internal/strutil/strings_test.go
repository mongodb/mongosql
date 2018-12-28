package strutil_test

import (
	"bytes"
	"testing"

	"github.com/10gen/sqlproxy/internal/strutil"
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
