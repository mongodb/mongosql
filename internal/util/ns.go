package util

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// This excludes the special regexp characters that
	// we handle ourselves: `*`, `\`, and `.`.
	specialRegexpChars = "`+?()|[]{}^`"
)

var (
	escapeReplacer = strings.NewReplacer(
		`\\`, `$escapedSlash$`,
		`\*`, `$escapedAsterisk$`,
		`.`, `$dot$`,
		`\`, ``,
	)

	regexpReplacer = strings.NewReplacer(
		`*`, `.*`,
		`$escapedAsterisk$`, `\*`,
		`$escapedSlash$`, `\\`,
		`$dot$`, `\.`,
	)
)

// buildRegexpPattern takes a namespaace pattern and turns it
// into a pattern that can be used with go's regex package.
func buildRegexpPattern(pattern nsPattern) rePattern {
	pat := string(pattern)
	pat = escapeReplacer.Replace(pat)
	pat = regexpReplacer.Replace(pat)
	pat = quoteSpecialRegexpChars(pat)
	pat = fmt.Sprintf("^%s$", pat)
	return rePattern(pat)
}

// quoteSpecialRegexpChars escapes special regexp characters.
// Adapted from the go stdlib's regexp.QuoteMeta function.
func quoteSpecialRegexpChars(s string) string {
	var i int
	for i = 0; i < len(s); i++ {
		if strings.Contains(specialRegexpChars, string(s[i])) {
			break
		}
	}

	// No meta characters found, so return original string.
	if i >= len(s) {
		return s
	}

	b := make([]byte, 2*len(s)-i)
	copy(b, s[:i])
	j := i
	for ; i < len(s); i++ {
		if strings.Contains(specialRegexpChars, string(s[i])) {
			b[j] = '\\'
			j++
		}
		b[j] = s[i]
		j++
	}
	return string(b[:j])
}

// usesWildcard returns whether the provided pattern uses any wildcards.
func usesWildcard(pattern nsPattern) bool {
	pat := string(pattern)
	pat = escapeReplacer.Replace(pat)
	return strings.Contains(pat, "*")
}

// validateMatcherPattern checks whether the provided pattern is a properly
// formatted namespace pattern. If not, an error is returned.
func validateMatcherPattern(pattern nsPattern) error {
	pat := string(pattern)
	if strings.Contains(pat, "$") {
		return fmt.Errorf("'$' is not allowed in pattern")
	}
	if strings.HasPrefix(pat, ".") || strings.HasSuffix(pat, ".") {
		return fmt.Errorf("pattern may not begin or end with a '.'")
	}
	if !strings.Contains(pat, ".") {
		return fmt.Errorf("pattern must be of the format '<db>.<collection>'")
	}
	return nil
}

// There are two types of patterns that we use in this file. though they are
// both strings, we will use two different types to differentiate them more
// easily as they are used in functions and structs below.

// nsPattern is a namespace pattern, i.e. a pattern that uses
// the glob syntax accepted by the --sampleNamespaces option.
type nsPattern string

// rePattern is a pattern that uses golang's regex syntax
type rePattern string

// Matcher identifies namespaces given user-defined patterns.
type Matcher struct {
	regexpPattern rePattern
	regexpMatcher *regexp.Regexp
	dbMatchers    map[nsPattern]*dbMatcher
}

// NewMatcher creates a matcher that will use the given list
// patterns to match namespaces.
func NewMatcher(patterns []string) (*Matcher, error) {

	if len(patterns) == 0 {
		return nil, fmt.Errorf("no pattern supplied")
	}

	matcher := &Matcher{
		dbMatchers: make(map[nsPattern]*dbMatcher),
	}

	for _, pattern := range patterns {
		err := matcher.addPattern(nsPattern(pattern))
		if err != nil {
			return nil, fmt.Errorf("could not use sample namespace pattern '%s': %v", pattern, err)
		}
	}

	return matcher, nil
}

// Collections returns all the collections specified for the provided db,
// as long as only literal selectors were used to specify its collections.
// If any wildcards were used, a nil slice is returned.
func (m *Matcher) Collections(db string) []string {
	seen := map[string]struct{}{}
	collections := []string{}
	for _, dbm := range m.dbMatchers {
		if dbm.matchesDB(db) {
			for _, col := range dbm.collections() {
				if _, ok := seen[col]; ok {
					continue
				}
				collections = append(collections, col)
				seen[col] = struct{}{}
			}
		}
	}
	return collections
}

// Has returns whether the given namespace matches any of the matcher's pattern.
func (m *Matcher) Has(namespace string) bool {
	return m.regexpMatcher.MatchString(namespace)
}

// HasDatabase returns whether the given database matches any of the matcher's
// patterns.
func (m *Matcher) HasDatabase(db string) bool {
	for _, dbm := range m.dbMatchers {
		if dbm.matchesDB(db) {
			return true
		}
	}
	return false
}

// Databases returns all the database names matched by this matcher, provided
// that the matcher uses only literal database names. If the matcher uses
// wildcards in any db selector, a nil slice is returned.
func (m *Matcher) Databases() []string {
	if m.UsesWildcardDB() {
		return nil
	}
	var dbs []string
	for _, dbm := range m.dbMatchers {
		dbs = append(dbs, string(dbm.dbPattern))
	}
	return dbs
}

// Namespaces returns all the namespaces matched by this matcher, provided
// that the matcher uses only literal namespaces. If the matcher uses
// wildcards in any db or collection selector, a nil map is returned.
func (m *Matcher) Namespaces() map[string][]string {
	if m.UsesWildcardDB() || m.UsesAnyWildcardCollection() {
		return nil
	}

	namespaces := make(map[string][]string)
	for _, dbm := range m.dbMatchers {
		namespaces[string(dbm.dbPattern)] = dbm.collections()
	}

	return namespaces
}

// UsesWildcardDB returns whether this matcher uses a wildcard pattern for its
// database selector.
func (m *Matcher) UsesWildcardDB() bool {
	for _, dbm := range m.dbMatchers {
		if dbm.usesWildcardDB() {
			return true
		}
	}
	return false
}

// UsesWildcardCollection returns whether the matcher for this db
// uses a wildcard pattern for its collection selector.
func (m *Matcher) UsesWildcardCollection(db string) bool {
	for _, dbm := range m.dbMatchers {
		if dbm.matchesDB(db) && dbm.usesWildcardCollection() {
			return true
		}
	}
	return false
}

// UsesAnyWildcardCollection returns whether the matcher for any db
// uses a wildcard pattern for its collection selector.
func (m *Matcher) UsesAnyWildcardCollection() bool {
	for _, dbm := range m.dbMatchers {
		if dbm.usesWildcardCollection() {
			return true
		}
	}
	return false
}

func (m *Matcher) addPattern(pattern nsPattern) error {
	pat := string(pattern)

	// normalize and validate the provided pattern
	pat = strings.TrimSpace(pat)
	pattern = nsPattern(pat)
	err := validateMatcherPattern(pattern)
	if err != nil {
		return err
	}

	// split the pattern into database and collection parts
	components := strings.SplitN(pat, ".", 2)
	dbPattern, colPattern := nsPattern(components[0]), nsPattern(components[1])

	// add this pattern to the db matchers
	dbm, ok := m.dbMatchers[dbPattern]
	if !ok {
		dbm, err = newDBMatcher(dbPattern)
		if err != nil {
			return fmt.Errorf(
				"could not create new db matcher with db pattern %q: %v",
				dbPattern, err,
			)
		}
	}
	err = dbm.addPattern(colPattern)
	if err != nil {
		return fmt.Errorf("could not add collection pattern %q to db matcher: %v", colPattern, err)
	}
	m.dbMatchers[dbPattern] = dbm

	// Update the regexp pattern and matcher
	if m.regexpPattern == "" {
		m.regexpPattern = buildRegexpPattern(pattern)
	} else {
		pat := rePattern(fmt.Sprintf("%s|%s", buildRegexpPattern(pattern), m.regexpPattern))
		m.regexpPattern = pat
	}
	m.regexpMatcher, err = regexp.Compile(string(m.regexpPattern))

	return err
}

type dbMatcher struct {
	dbPattern          nsPattern
	dbRegexp           *regexp.Regexp
	collectionPatterns map[nsPattern]struct{}
	collectionPattern  rePattern
	collectionMatcher  *regexp.Regexp
}

func newDBMatcher(pattern nsPattern) (*dbMatcher, error) {
	dbRegexp, err := regexp.Compile(string(buildRegexpPattern(pattern)))
	if err != nil {
		return nil, err
	}

	m := &dbMatcher{
		dbPattern:          pattern,
		dbRegexp:           dbRegexp,
		collectionPatterns: make(map[nsPattern]struct{}),
	}

	return m, nil
}

func (m *dbMatcher) addPattern(pattern nsPattern) error {
	if _, ok := m.collectionPatterns[pattern]; ok {
		return nil
	}
	m.collectionPatterns[pattern] = struct{}{}

	if m.collectionPattern == "" {
		m.collectionPattern = buildRegexpPattern(pattern)
	} else {
		pat := rePattern(fmt.Sprintf("%s|%s", buildRegexpPattern(pattern), m.collectionPattern))
		m.collectionPattern = pat
	}

	matcher, err := regexp.Compile(string(m.collectionPattern))
	if err != nil {
		return err
	}

	m.collectionMatcher = matcher
	return nil
}

func (m *dbMatcher) collections() []string {
	if m.usesWildcardCollection() {
		return nil
	}

	collections := []string{}
	for pattern := range m.collectionPatterns {
		collections = append(collections, string(pattern))
	}

	return collections
}

func (m *dbMatcher) matchesDB(db string) bool {
	return m.dbRegexp.MatchString(db)
}

func (m *dbMatcher) usesWildcardDB() bool {
	return usesWildcard(m.dbPattern)
}

func (m *dbMatcher) usesWildcardCollection() bool {
	for pattern := range m.collectionPatterns {
		if usesWildcard(pattern) {
			return true
		}
	}
	return false
}
