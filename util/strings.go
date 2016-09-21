package util

import (
	"path/filepath"
)

// Pluralize takes an amount and two strings denoting the singular
// and plural noun the amount represents. If the amount is singular,
// the singular form is returned; otherwise plural is returned. E.g.
//  Pluralize(X, "mouse", "mice") -> 0 mice, 1 mouse, 2 mice, ...
func Pluralize(amount int, singular, plural string) string {
	if amount == 1 {
		return singular
	}
	return plural
}

// ToUniversalPath returns the result of replacing each slash ('/') character
// in "path" with an OS-sepcific separator character. Multiple slashes are
// replaced by multiple separators
func ToUniversalPath(path string) string {
	return filepath.FromSlash(path)
}
