package util_test

import (
	"testing"

	"github.com/10gen/sqlproxy/internal/util"
)

func TestByteString(t *testing.T) {

	tests := []struct {
		count int
		s     string
	}{
		{0, "0"},
		{27, "27B"},
		{1023, "1023B"},
		{1024, "1K"},
		{1728, "1.7K"},
		{110592, "108K"},
		{7077888, "6.8M"},
		{45298432, "43.2M"},
		{28991029248, "27G"},
	}

	for _, test := range tests {
		actual := util.ByteString(test.count)
		if actual != test.s {
			t.Fatalf("expected '%s' by got '%s'", test.s, actual)
		}
	}
}
