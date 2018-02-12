package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatcher(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		testMatchPattern(t)
		testMatchEverything(t)
		testInvalidMatchers(t)
		testHasDatabase(t)
	})
	t.Run("query", func(t *testing.T) {
		testUsesWildcardDB(t)
		testUsesWildcardCollection(t)
		testUsesAnyWildcardCollection(t)
		testDatabases(t)
		testCollections(t)
	})
}

func testMatchPattern(t *testing.T) {
	req := require.New(t)

	matchers := []string{
		`*.user*`,
		`pr\*d.*`,
		`olea.*`,
		`paren]o.*`,
		`carr\ie.*`,
		`dsla\\xp.*`,
		`tmp|.*`,
		`kathy.*bo*x*`,
	}

	m, err := NewMatcher(matchers)
	req.NoError(err, "failed to create new matcher")
	req.NotNil(m, "failed to create new matcher")

	tests := []struct {
		ns       string
		expected bool
	}{
		{"olaa.bobb", false},
		{"olea.bobb", true},
		{`carr\ie.bobb`, false},
		{"carrie.bobb", true},
		{`dsla\xp.bobb`, true},
		{"kathy.box", true},
		{"kathy.borx", true},
		{"paren]o.borx", true},
		{"kathy.boxer", true},
		{"kathy.rocks.boxs", true},
		{"stuff.user", true},
		{"stuf]f.user", true},
		{"stuff.users", true},
		{"stuff.users", true},
		{"pr*d.users", true},
		{"pr*d.magic", true},
		{`pr*d\.magic`, false},
		{"prod.magic", false},
		{"pr*d.turbo.encabulators", true},
		{"st*ging.turbo.encabulators", false},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("pattern_%d", i), func(t *testing.T) {
			req := require.New(t)
			actual := m.Has(test.ns)
			req.Equal(test.expected, actual, "actual match value should equal expected")
		})
	}
}

func testMatchEverything(t *testing.T) {
	req := require.New(t)

	m, err := NewMatcher([]string{"*.*"})
	req.NoError(err, "failed to create new matcher")
	req.NotNil(m, "failed to create new matcher")

	tests := []struct {
		ns       string
		expected bool
	}{
		{"olaa.bobb", true},
		{"olea.bobb", true},
		{`carr\ie.bobb`, true},
		{"carrie.bobb", true},
		{`dsla\xp.bobb`, true},
		{"kathy.box", true},
		{"kathy.borx", true},
		{"paren]o.borx", true},
		{"kathy.boxer", true},
		{"kathy.rocks.boxs", true},
		{"stuff.user", true},
		{"stuf]f.user", true},
		{"stuff.users", true},
		{"stuff.users", true},
		{"pr*d.users", true},
		{"pr*d.magic", true},
		{`pr*d\.magic`, true},
		{"prod.magic", true},
		{"pr*d.turbo.encabulators", true},
		{"st*ging.turbo.encabulators", true},
		{"stuff", false},
		{"stuff.user", true},
		{"stuff.users", true},
		{"prod.turbo.encabulators", true},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("match_everything_%d", i), func(t *testing.T) {
			req := require.New(t)
			actual := m.Has(test.ns)
			req.Equal(test.expected, actual, "actual match value should equal expected")
		})
	}
}

func testInvalidMatchers(t *testing.T) {
	tests := []string{
		"$.user",
		"",
		"*.user$",
		"user",
	}

	for i, pattern := range tests {
		t.Run(fmt.Sprintf("invalid_matcher_%d", i), func(t *testing.T) {
			req := require.New(t)
			m, err := NewMatcher([]string{pattern})
			req.Error(err, "invalid pattern should cause error")
			req.Nil(m, "matcher should be nil when error is returned")
		})
	}

}

func testHasDatabase(t *testing.T) {
	req := require.New(t)

	m, err := NewMatcher([]string{"abc*.*", "def.c", "ghi.*"})
	req.NoError(err, "failed to create new matcher")
	req.NotNil(m, "failed to create new matcher")

	tests := []struct {
		db       string
		expected bool
	}{
		{"abc", true},
		{"abcd", true},
		{"bcd", false},
		{"bc", false},
		{"def", true},
		{"de", false},
		{"defg", false},
		{"ghi", true},
		{"g", false},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("has_database_%d", i), func(t *testing.T) {
			req := require.New(t)
			actual := m.HasDatabase(test.db)
			req.Equal(test.expected, actual, "actual match value should equal expected")
		})
	}
}

func testUsesWildcardDB(t *testing.T) {
	type testCase struct {
		selectors []string
		expected  bool
	}

	newTestCase := func(expected bool, selectors ...string) testCase {
		return testCase{
			selectors,
			expected,
		}
	}

	tests := []testCase{
		newTestCase(true, "*.*"),
		newTestCase(true, "*.c"),
		newTestCase(true, "ab*.c"),
		newTestCase(true, "a*b.c"),
		newTestCase(true, "**.c"),
		newTestCase(true, "*ab.c*"),
		newTestCase(true, "*ab.foo.c*"),
		newTestCase(false, "ab.*.c"),
		newTestCase(false, "ab.f*oo.c*"),
		newTestCase(false, "ab.c"),
		newTestCase(true, "ab.c", "*.*"),
		newTestCase(true, "ab.*", "*.foo"),
		newTestCase(false, "ab.c", "ab.*.foo"),
		newTestCase(false, "ab.c", "ab.*.foo", `ab\*.foo`),
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("uses_wildcard_db_%d", i), func(t *testing.T) {
			req := require.New(t)

			m, err := NewMatcher(test.selectors)
			req.NoError(err, "failed to create new matcher")
			req.NotNil(m, "failed to create new matcher")

			actual := m.UsesWildcardDB()

			req.Equal(test.expected, actual, "actual match value should equal expected")
		})
	}
}

func testUsesWildcardCollection(t *testing.T) {
	type testCase struct {
		selectors []string
		db        string
		expected  bool
	}

	newTestCase := func(expected bool, db string, selectors ...string) testCase {
		return testCase{
			selectors,
			db,
			expected,
		}
	}

	tests := []testCase{
		newTestCase(true, "foo", "*.*"),
		newTestCase(true, "foo*", "*.*"),
		newTestCase(false, "foo", "*.c"),
		newTestCase(false, "bar*", "*.c"),
		newTestCase(false, "abc", "ab*.c"),
		newTestCase(false, "def", "ab*.c"),
		newTestCase(true, "abc", "ab*.c*"),
		newTestCase(false, "def", "ab*.c*"),
		newTestCase(true, "abc", "ab*.c*", "def.c*"),
		newTestCase(true, "def", "ab*.c*", "def.c*"),
		newTestCase(true, "abc", "ab*.c", "ab*.c*"),
		newTestCase(false, "def", "ab*.c", "ab*.c*"),
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("uses_wildcard_collection_%d", i), func(t *testing.T) {
			req := require.New(t)

			m, err := NewMatcher(test.selectors)
			req.NoError(err, "failed to create new matcher")
			req.NotNil(m, "failed to create new matcher")

			actual := m.UsesWildcardCollection(test.db)

			req.Equal(test.expected, actual, "actual match value should equal expected")
		})
	}
}

func testUsesAnyWildcardCollection(t *testing.T) {
	type testCase struct {
		selectors []string
		expected  bool
	}

	newTestCase := func(expected bool, selectors ...string) testCase {
		return testCase{
			selectors,
			expected,
		}
	}

	tests := []testCase{
		newTestCase(true, "*.*"),
		newTestCase(false, "*.c"),
		newTestCase(false, "ab*.c"),
		newTestCase(false, "a*b.c"),
		newTestCase(false, "**.c"),
		newTestCase(true, "*ab.c*"),
		newTestCase(true, "*ab.foo.c*"),
		newTestCase(true, "ab.*.c"),
		newTestCase(true, "ab.f**"),
		newTestCase(false, `ab.f\*c`),
		newTestCase(true, "ab.c", "*.*"),
		newTestCase(true, "ab.*", "*.foo"),
		newTestCase(true, "ab.c", "ab.*.foo"),
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("uses_any_wildcard_collection_%d", i), func(t *testing.T) {
			req := require.New(t)

			m, err := NewMatcher(test.selectors)
			req.NoError(err, "failed to create new matcher")
			req.NotNil(m, "failed to create new matcher")

			actual := m.UsesAnyWildcardCollection()

			req.Equal(test.expected, actual, "actual match value should equal expected")
		})
	}
}

func testDatabases(t *testing.T) {
	tests := []struct {
		selectors []string
		expected  []string
	}{
		{[]string{"*.*"}, nil},
		{[]string{"*.foo"}, nil},
		{[]string{"*.foo", "foo.bar"}, nil},
		{[]string{"foo.bar"}, []string{"foo"}},
		{[]string{"foo.bar", "foo.bar"}, []string{"foo"}},
		{[]string{"foo.bar", "foo.baz"}, []string{"foo"}},
		{[]string{"foo.bar", "foo.baz", "bar.*"}, []string{"foo", "bar"}},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("databases_%d", i), func(t *testing.T) {
			req := require.New(t)

			m, err := NewMatcher(test.selectors)
			req.NoError(err, "failed to create new matcher")
			req.NotNil(m, "failed to create new matcher")

			actual := m.Databases()

			req.ElementsMatch(test.expected, actual, "actual match value should equal expected")
		})
	}
}

func testCollections(t *testing.T) {
	tests := []struct {
		db        string
		selectors []string
		expected  []string
	}{
		{"foo", []string{"*.*"}, nil},
		{"foo", []string{"*.foo"}, []string{"foo"}},
		{"bar", []string{"*.foo"}, []string{"foo"}},
		{"foo", []string{"*.foo", "foo.bar"}, []string{"foo", "bar"}},
		{"bar", []string{"*.foo", "foo.bar"}, []string{"foo"}},
		{"bar", []string{"*.foo", "bar.foo", "foo.bar"}, []string{"foo"}},
		{"foo", []string{"foo.bar"}, []string{"bar"}},
		{"bar", []string{"foo.bar"}, nil},
		{"foo", []string{"foo.bar", "foo.bar"}, []string{"bar"}},
		{"foo", []string{"foo.bar", "foo.baz"}, []string{"bar", "baz"}},
		{"foo", []string{"foo.bar", "bar.baz"}, []string{"bar"}},
		{"bar", []string{"foo.bar", "bar.baz"}, []string{"baz"}},
		{"foo", []string{"foo.bar", "foo.baz", "bar.baz", "bar.*"}, []string{"bar", "baz"}},
		{"bar", []string{"foo.bar", "foo.baz", "bar.baz", "bar.*"}, nil},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("collections_%d", i), func(t *testing.T) {
			req := require.New(t)

			m, err := NewMatcher(test.selectors)
			req.NoError(err, "failed to create new matcher")
			req.NotNil(m, "failed to create new matcher")

			actual := m.Collections(test.db)

			req.ElementsMatch(test.expected, actual, "actual match value should equal expected")
		})
	}
}
