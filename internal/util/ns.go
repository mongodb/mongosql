package util

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	exclusionChar = '~'
)

// There are two types of patterns that we use in this file. though they are
// both strings, we will use two different types to differentiate them more
// easily as they are used in functions and structs below.

// nsPattern is a namespace pattern, i.e. a pattern that uses
// the glob syntax accepted by the --sampleNamespaces option.
type nsPattern string

// rePattern is a pattern that uses golang's regex syntax.
type rePattern string

// Matcher identifies namespaces given user-defined patterns.
type Matcher struct {
	regexpPattern rePattern
	regexpMatcher *regexp.Regexp
	dbMatchers    map[nsPattern]*dbMatcher
	// exclusionMatcher is a matcher that is used to handle
	// exclusion patterns. It is not used in a recursive manner and is
	// here primary as a helper to the original matcher.
	exclusionMatcher *Matcher
}

// NewMatcher creates a matcher that will use the given list
// patterns to match namespaces.
func NewMatcher(patterns []string) (*Matcher, error) {
	var matcher *Matcher
	var err error

	if len(patterns) == 0 {
		return nil, fmt.Errorf("no pattern supplied")
	}

	inclusions, exclusions, err := splitPatterns(patterns)
	if err != nil {
		return nil, err
	}

	// add wildcard db and collection selector if only exclusionary patterns are present.
	if len(inclusions) == 0 {
		inclusions = []string{fmt.Sprintf("%v.%v", wildCardCharacter, wildCardCharacter)}
	}

	matcher, err = newMatcherHelper(inclusions)
	if err != nil {
		return nil, err
	}

	matcher.exclusionMatcher, err = newMatcherHelper(exclusions)
	if err != nil {
		return nil, err
	}

	return matcher, nil
}

func newMatcherHelper(patterns []string) (*Matcher, error) {
	if len(patterns) == 0 {
		return nil, nil
	}

	matcher := &Matcher{
		dbMatchers: make(map[nsPattern]*dbMatcher),
	}

	for _, pattern := range patterns {
		if err := matcher.addPattern(nsPattern(pattern)); err != nil {
			return nil, fmt.Errorf("could not use sample namespace pattern '%s': %v", pattern, err)
		}
	}
	return matcher, nil
}

// CanEnumerateAllCollections returns true if the matcher requires no additional information
// to enumerate all collections in database, db, matching the supplied patterns. This happens
// if the matcher does not use any wildcard character for any collections on the given database.
// Otherwise, this function returns false.
func (m *Matcher) CanEnumerateAllCollections(db string) bool {
	return !m.UsesWildcardCollection(db)
}

// CanEnumerateAllDatabases returns true if the matcher requires no additional information
// to enumerate all databases matching the supplied patterns. This happens if the matcher
// does not use any wildcard database. Otherwise, this function returns false.
func (m *Matcher) CanEnumerateAllDatabases() bool {
	return !m.UsesWildcardDB()
}

// CanEnumerateAllNamespaces returns true if the matcher requires no additional information
// to enumerate all namespaces matching the supplied patterns. This happens if the matcher
// does not use any wildcard character for any database or collection.
// Otherwise, this function returns false.
func (m *Matcher) CanEnumerateAllNamespaces() bool {
	return !m.UsesAnyWildcardCollection() && !m.UsesWildcardDB()
}

// Collections returns all the collections specified for the provided db, as long as only literal
// selectors were used to specify its collections and the collections are not excluded.
// If any wildcards were used, a nil slice is returned.
func (m *Matcher) Collections(db string) []string {
	excludedCollections := map[string]struct{}{}
	if m.exclusionMatcher != nil {
		for _, col := range m.exclusionMatcher.Collections(db) {
			excludedCollections[col] = struct{}{}
		}
	}

	seen := map[string]struct{}{}
	collections := []string{}
	for _, dbm := range m.dbMatchers {
		if dbm.matchesDB(db) {
			for _, col := range dbm.collections() {
				if _, ok := seen[col]; ok {
					continue
				}
				if _, ok := excludedCollections[col]; ok {
					continue
				}
				collections = append(collections, col)
				seen[col] = struct{}{}
			}
		}
	}
	return collections
}

// Databases returns all the database names matched by this matcher, provided that the matcher uses
// only literal database names and isn't excluded. If the matcher uses wildcards in any db selector,
// a nil slice is returned.
func (m *Matcher) Databases() []string {
	if m.UsesWildcardDB() {
		return nil
	}

	var dbs []string
	for _, dbm := range m.dbMatchers {
		dbName := string(dbm.dbPattern)
		if !m.MustExcludeDatabase(dbName) {
			dbs = append(dbs, dbName)
		}
	}
	return dbs
}

// Has returns whether the given namespace matches any of the matcher's pattern.
func (m *Matcher) Has(namespace string) bool {
	if m.regexpMatcher.MatchString(namespace) {
		if m.exclusionMatcher != nil {
			return !m.exclusionMatcher.Has(namespace)
		}
		return true
	}
	return false
}

// HasDatabase returns whether the given database matches any of the matcher's
// patterns.
func (m *Matcher) HasDatabase(db string) bool {
	for _, dbm := range m.dbMatchers {
		if dbm.matchesDB(db) && !m.MustExcludeDatabase(db) {
			return true
		}
	}
	return false
}

// MustExcludeDatabase returns true if a matcher for the given database uses
// a wildcard pattern for its collection and contains an exclusion pattern.
func (m *Matcher) MustExcludeDatabase(db string) bool {
	if m.exclusionMatcher == nil {
		return false
	}

	for _, dbm := range m.exclusionMatcher.dbMatchers {
		// This checks for a pattern like: `test.*`
		if dbm.hasWildcardCollection() && dbm.dbRegexp.MatchString(db) {
			return true
		}
	}
	return false
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
		dbName := string(dbm.dbPattern)
		for _, col := range dbm.collections() {
			ns := fmt.Sprintf("%v.%v", dbName, col)
			if m.exclusionMatcher != nil && m.exclusionMatcher.Has(ns) {
				continue
			}
			namespaces[dbName] = append(namespaces[dbName], col)
		}
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
	var err error

	// unescape tilde and split the pattern into database and collection parts
	pat = tildeUnescaper.Replace(pat)
	components := strings.SplitN(pat, ".", 2)
	dbPattern, colPattern := nsPattern(components[0]), nsPattern(components[1])

	// add this pattern to the db matchers
	dbm, ok := m.dbMatchers[dbPattern]
	if !ok {
		dbm, err = newDBMatcher(dbPattern)
		if err != nil {
			return fmt.Errorf("could not create new db matcher with db pattern %q: %v",
				dbPattern, err)
		}
	}

	if err = dbm.addPattern(colPattern); err != nil {
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

func (m *dbMatcher) hasWildcardCollection() bool {
	for pattern := range m.collectionPatterns {
		if isWildcard(pattern) {
			return true
		}
	}
	return false
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

func splitPatterns(patterns []string) ([]string, []string, error) {
	var inclusions, exclusions []string
	for _, pat := range patterns {
		// normalize the provided pattern
		pat = strings.TrimSpace(pat)

		// validate the provided pattern
		if err := validateMatcherPattern(pat); err != nil {
			return nil, nil, err
		}

		if len(pat) > 0 && pat[0] == exclusionChar {
			pat = pat[1:]
			exclusions = append(exclusions, pat)
		} else {
			inclusions = append(inclusions, pat)
		}
	}
	return inclusions, exclusions, nil
}

// validateMatcherPattern checks whether the provided pattern is a properly
// formatted namespace pattern. If not, an error is returned.
func validateMatcherPattern(pat string) error {
	patternFormat := "pattern must be of the format '<db>.<collection>' or '~<db>.<collection>'"

	if strings.Contains(pat, "$") {
		return fmt.Errorf("'$' is not allowed in pattern")
	}
	if strings.HasPrefix(pat, ".") || strings.HasSuffix(pat, ".") {
		return fmt.Errorf("pattern may not begin or end with a '.'")
	}
	if !strings.Contains(pat, ".") || strings.HasPrefix(pat, "~.") {
		return fmt.Errorf(patternFormat)
	}
	if pat == fmt.Sprintf("%s*.*", string(exclusionChar)) {
		return fmt.Errorf("can not exclude all namespaces")
	}
	return nil
}
