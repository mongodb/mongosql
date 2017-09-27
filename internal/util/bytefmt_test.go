package util_test

import (
	"testing"

	"github.com/10gen/sqlproxy/internal/util"
)

func TestByteString(t *testing.T) {

	tests := []struct {
		count uint64
		s     string
	}{
		{0, "0B"},
		{27, "27B"},
		{1023, "1023B"},
		{1024, "1KiB"},
		{1728, "1.7KiB"},
		{110592, "108KiB"},
		{7077888, "6.8MiB"},
		{45298432, "43.2MiB"},
		{28991029248, "27GiB"},
	}

	for _, test := range tests {
		actual := util.ByteString(test.count)
		if actual != test.s {
			t.Fatalf("expected '%s' by got '%s'", test.s, actual)
		}
	}
}
