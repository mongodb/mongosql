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

// Matcher identifies namespaces given user-defined patterns.
type Matcher struct {
	// regexp to check namespaces against
	matcher *regexp.Regexp
}

// NewMatcher creates a matcher that will use the given list
// patterns to match namespaces.
func NewMatcher(patterns []string) (*Matcher, error) {

	if len(patterns) == 0 {
		return nil, fmt.Errorf("no pattern supplied")
	}

	regexpPattern := ""

	for _, pattern := range patterns {
		if strings.Contains(pattern, "$") {
			return nil, fmt.Errorf("'$' is not allowed in "+
				"sample namespace pattern: '%v'", pattern)
		}

		pattern = escapeReplacer.Replace(pattern)
		pattern = regexpReplacer.Replace(pattern)
		pattern = quoteSpecialRegexpChars(pattern)
		regexpPattern += fmt.Sprintf("^%s$|", pattern)
	}

	regexpPattern = regexpPattern[:len(regexpPattern)-1]
	return &Matcher{
		matcher: regexp.MustCompile(regexpPattern),
	}, nil
}

// Has returns whether the given namespace matches any of the matcher's pattern.
func (m *Matcher) Has(namespace string) bool {
	return m.matcher.MatchString(namespace)
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
