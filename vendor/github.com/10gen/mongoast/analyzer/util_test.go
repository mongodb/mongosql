package analyzer_test

import (
	"fmt"
	"testing"

	"github.com/10gen/mongoast/analyzer"
	"github.com/10gen/mongoast/internal/testutil"
	"github.com/google/go-cmp/cmp"
)

func TestIsDotPrefixOf(t *testing.T) {
	testCases := []struct {
		prefix    string
		totalPath string
		expected  bool
	}{
		{"a", "", false},
		{"", "a", false},
		{"a.b", "a.b.c", true},
		{"a.b", "a.bb.c", false},
		{"a.b", "a.b.c.d.ee.ff", true},
		// We want a path to be a prefix of itself to simplify use of this code.
		{"a.b", "a.b", true},
	}

	for _, tc := range testCases {
		callString := fmt.Sprintf("IsDotPrefixString(%s, %s)", tc.prefix, tc.totalPath)
		t.Run(callString, func(t *testing.T) {
			isPrefix := analyzer.IsDotPrefixOfString(tc.prefix, tc.totalPath)
			if isPrefix != tc.expected {
				t.Fatalf("expected %s to be %v", callString, tc.expected)
			}
		})
	}
}

func TestGetDotPrefixesOf(t *testing.T) {
	testCases := []struct {
		path     string
		expected []string
	}{
		{"", []string{""}},
		{"a", []string{"a"}},
		{"a.b", []string{"a", "a.b"}},
		{"a.b.c", []string{"a", "a.b", "a.b.c"}},
		{"a.b.c.d", []string{"a", "a.b", "a.b.c", "a.b.c.d"}},
	}

	for _, tc := range testCases {
		callString := fmt.Sprintf("GetDotPrefixesOfString(%s)", tc.path)
		t.Run(callString, func(t *testing.T) {
			prefixes := analyzer.GetDotPrefixesOfString(tc.path)
			if !cmp.Equal(tc.expected, prefixes) {
				t.Fatalf("expected and actual do not match:  %s", cmp.Diff(tc.expected, prefixes))
			}
		})
	}
}

func TestGetPathRootString(t *testing.T) {
	testCases := []struct {
		path     string
		expected string
	}{
		{"", ""},
		{"a", "a"},
		{"a.b", "a"},
		{"a.b.c", "a"},
		{"a.b.c.d", "a"},
		{"b.b.c.d", "b"},
	}

	for _, tc := range testCases {
		callString := fmt.Sprintf("GetPathRootString(%s)", tc.path)
		t.Run(callString, func(t *testing.T) {
			prefixes := analyzer.GetPathRootString(tc.path)
			if !cmp.Equal(tc.expected, prefixes) {
				t.Fatalf("expected and actual do not match:  %s", cmp.Diff(tc.expected, prefixes))
			}
		})
	}
}

func TestGetPathRootFromRef(t *testing.T) {
	testCases := []struct {
		path     string
		expected string
	}{
		{"", ""},
		{"a", "a"},
		{"a.b", "a"},
		{"a.b.c", "a"},
		{"a.b.c.d", "a"},
		{"b.b.c.d", "b"},
	}

	for _, tc := range testCases {
		callString := fmt.Sprintf("GetPathRootFromRef(%s)", tc.path)
		t.Run(callString, func(t *testing.T) {
			prefixes, _ := analyzer.GetPathRootFromRef(testutil.StringToFieldRef(tc.path))
			if !cmp.Equal(tc.expected, prefixes) {
				t.Fatalf("expected and actual do not match:  %s", cmp.Diff(tc.expected, prefixes))
			}
		})
	}
}
