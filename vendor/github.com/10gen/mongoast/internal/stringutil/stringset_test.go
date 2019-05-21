package stringutil_test

import (
	"testing"

	"github.com/10gen/mongoast/internal/stringutil"
)

func TestStringSetIntersection(t *testing.T) {
	testCases := []struct {
		name     string
		input1   []string
		input2   []string
		expected []string
	}{
		{"has empty intersection",
			[]string{"a", "b", "bb"},
			[]string{"c", "aa"},
			[]string{},
		},
		{"has non-empty intersection",
			[]string{"a", "b", "bb"},
			[]string{"bb", "a"},
			[]string{"bb", "a"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			left, right, expected := stringutil.NewStringSet(), stringutil.NewStringSet(), stringutil.NewStringSet()
			left.AddSlice(tc.input1)
			right.AddSlice(tc.input2)
			expected.AddSlice(tc.expected)

			out := left.Intersection(right)

			if !out.Equals(expected) {
				t.Fatalf("expected %v, got %v", tc.expected, out.SortedSlice())
			}
			if len(tc.expected) == 0 && left.HasIntersection(right) {
				t.Fatalf("HasIntersection should be false")
			}
			if len(tc.expected) != 0 && !left.HasIntersection(right) {
				t.Fatalf("HasIntersection should be true")
			}
		})
	}
}
