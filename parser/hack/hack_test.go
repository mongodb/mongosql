package hack_test

import (
	"bytes"
	"testing"

	"github.com/10gen/sqlproxy/parser/hack"
)

func TestString(t *testing.T) {
	b := []byte("hello world")
	a := hack.String(b)

	if a != "hello world" {
		t.Fatal(a)
	}

	b[0] = 'a'

	if a != "aello world" {
		t.Fatal(a)
	}

	b = append(b, "abc"...)
	if a != "aello world" {
		t.Fatal(a)
	}
	_ = b
}

func TestByte(t *testing.T) {
	a := "hello world"

	b := hack.Slice(a)

	if !bytes.Equal(b, []byte("hello world")) {
		t.Fatal(string(b))
	}
}
