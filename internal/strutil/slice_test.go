package strutil_test

import (
	"testing"

	. "github.com/10gen/sqlproxy/internal/strutil"
)

func TestSliceContains(t *testing.T) {
	tests := []struct {
		slice    interface{}
		element  interface{}
		contains bool
	}{
		{[]int{3, 5, 23}, 4, false},
		{[]int{3, 5, 23}, 3, true},
		{[]int{3, 5, 23}, 2, false},
		{[]int{}, 2, false},
		{[]int{0}, 0, true},
		{[]int{1}, 0, false},
		{[]int{0}, 1, false},
		{[]string{"", "def", "ghi"}, "", true},
		{[]string{"abc", "def", "ghi"}, "gh", false},
		{[]string{"abc", "def", "ghi"}, "ghi", true},
		{[]string{"abc", "def", "ghi"}, "", false},
	}

	for _, test := range tests {
		contains := SliceContains(test.slice, test.element)
		if contains != test.contains {
			t.Fatalf("expected '%t' by got '%t'", test.contains, contains)
		}
	}
}

func TestIntSliceContains(t *testing.T) {
	tests := []struct {
		slice    []int
		element  int
		contains bool
	}{
		{[]int{3, 5, 23}, 4, false},
		{[]int{3, 5, 23}, 3, true},
		{[]int{3, 5, 23}, 2, false},
		{[]int{}, 2, false},
		{[]int{0}, 0, true},
		{[]int{1}, 0, false},
		{[]int{0}, 1, false},
	}

	for _, test := range tests {
		contains := IntSliceContains(test.slice, test.element)
		if contains != test.contains {
			t.Fatalf("expected '%t' by got '%t'", test.contains, contains)
		}
	}
}

func TestIntSliceIndex(t *testing.T) {
	tests := []struct {
		slice   []int
		element int
		index   int
	}{
		{[]int{3, 5, 23}, 4, -1},
		{[]int{3, 5, 23}, 3, 0},
		{[]int{3, 5, 23}, 5, 1},
		{[]int{3, 5, 23}, 23, 2},
		{[]int{3, 5, 23, 23}, 23, 2},
		{[]int{3, 5, 23}, 2, -1},
		{[]int{}, 2, -1},
		{[]int{0}, 0, 0},
		{[]int{1}, 0, -1},
		{[]int{0}, 1, -1},
	}

	for _, test := range tests {
		index := IntSliceIndex(test.slice, test.element)
		if index != test.index {
			t.Fatalf("expected '%d' by got '%d'", test.index, index)
		}
	}
}

func TestStringSliceContains(t *testing.T) {
	tests := []struct {
		slice    []string
		element  string
		contains bool
	}{
		{[]string{"abc", "def", "ghi"}, "gh", false},
		{[]string{"abc", "def", "ghi"}, "ghi", true},
		{[]string{"abc", "def", "ghi"}, "", false},
		{[]string{"", "def", "ghi"}, "", true},
		{[]string{"", "", "ghi"}, "", true},
		{[]string{""}, "", true},
	}

	for _, test := range tests {
		contains := StringSliceContains(test.slice, test.element)
		if contains != test.contains {
			t.Fatalf("expected '%t' by got '%t'", test.contains, contains)
		}
	}
}

func TestStringSliceIndex(t *testing.T) {
	tests := []struct {
		slice   []string
		element string
		index   int
	}{
		{[]string{"abc", "def", "ghi"}, "gh", -1},
		{[]string{"abc", "def", "ghi"}, "ab", -1},
		{[]string{"abc", "def", "ghi"}, "ghi", 2},
		{[]string{"abc", "def", "ghi"}, "", -1},
		{[]string{"", "def", "ghi"}, "", 0},
		{[]string{"", "", "ghi"}, "", 0},
		{[]string{""}, "", 0},
	}

	for _, test := range tests {
		index := StringSliceIndex(test.slice, test.element)
		if index != test.index {
			t.Fatalf("expected '%d' by got '%d'", test.index, index)
		}
	}
}

func TestSliceCount(t *testing.T) {
	tests := []struct {
		slice   interface{}
		element interface{}
		count   int
	}{
		{[]string{"abc", "def", "ghi"}, "gh", 0},
		{[]string{"abc", "def", "ghi"}, "ab", 0},
		{[]string{"abc", "def", "ghi"}, "ghi", 1},
		{[]string{"abc", "def", "ghi"}, "", 0},
		{[]string{"", "def", "ghi"}, "", 1},
		{[]string{"", "", "ghi"}, "", 2},
		{[]string{""}, "", 1},
		{[]int{3, 5, 23}, 4, 0},
		{[]int{3, 5, 23}, 3, 1},
		{[]int{3, 5, 23}, 5, 1},
		{[]int{3, 5, 23}, 23, 1},
		{[]int{3, 5, 23, 23}, 23, 2},
		{[]int{3, 5, 23}, 2, 0},
		{[]int{}, 2, 0},
		{[]int{0}, 0, 1},
		{[]int{1}, 0, 0},
		{[]int{0}, 1, 0},
	}

	for _, test := range tests {
		count := SliceCount(test.slice, test.element)
		if count != test.count {
			t.Fatalf("expected '%d' by got '%d'", test.count, count)
		}
	}
}
