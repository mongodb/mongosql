package strutil

import (
	"fmt"
	"strings"
)

const (
	// This excludes the special regexp characters that
	// we handle ourselves: `*`, `\`, and `.`.
	specialRegexpChars = "`+?()|[]{}^`"
	wildCardCharacter  = "*"
)

var (
	escapeReplacer = strings.NewReplacer(
		`\~`, `$escapedTilde$`,
		`\\`, `$escapedSlash$`,
		`\*`, `$escapedAsterisk$`,
		`.`, `$dot$`,
		`\`, ``,
	)

	regexpReplacer = strings.NewReplacer(
		`*`, `.*`,
		`$escapedAsterisk$`, `\*`,
		`$escapedTilde$`, `~`,
		`$escapedSlash$`, `\\`,
		`$dot$`, `\.`,
	)

	tildeUnescaper = strings.NewReplacer(
		`\~`, `~`,
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

// isWildcard returns whether the provided pattern is a wildcards.
func isWildcard(pattern nsPattern) bool {
	return escapeReplacer.Replace(string(pattern)) == wildCardCharacter
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
	return strings.Contains(pat, wildCardCharacter)
}
