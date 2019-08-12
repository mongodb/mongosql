package strutil

import (
	"bytes"
	"path/filepath"
	"reflect"
	"strings"
	"unsafe"
)

// MySQLCleanNumericString cleans up a numeric string using MySQL's rules (trim, then
// take everything before the first character that isn't . or a number). Must
// handle -, and should return "0" if no viable number can be found.
func MySQLCleanNumericString(s string) string {
	var out bytes.Buffer
	firstDecimal := true
	s = strings.TrimLeft(s, " \t\v\n\r")
	if len(s) == 0 {
		return "0"
	}
	firstChar := s[0]
	if firstChar == '-' || firstChar == '+' {
		out.WriteRune(rune(firstChar))
		s = s[1:]
	}
	for _, c := range s {
		if c == '.' {
			if firstDecimal {
				out.WriteRune(c)
				firstDecimal = false
				continue
			}
			break
		}
		if c >= '0' && c <= '9' {
			out.WriteRune(c)
			continue
		}
		break
	}
	ret := out.String()
	if len(ret) == 0 || ret == "-" {
		return "0"
	}
	return ret
}

// MySQLCleanScientificNotationString cleans up a numeric string that may
// contain scientific notation.
func MySQLCleanScientificNotationString(s string) string {
	s = strings.TrimLeft(s, " \t\v\n\r")
	splitted := strings.Split(s, "e")
	base, exponent := s, ""
	if len(splitted) > 1 {
		// Any extra parts can be safely dropped, mysql only
		// considers the first e.
		base, exponent = splitted[0], splitted[1]
	}

	cleanBase := MySQLCleanNumericString(base)
	// If the base part is _changed_ by MySQLCleanNumericString
	// that means it had trailing charaters, and the exponent should be
	// ignored. Unfortunately, we will TrimLeft twice to make this
	// check work.
	if cleanBase != base {
		return cleanBase
	}
	exponent = MySQLCleanNumericString(exponent)
	return base + "e" + exponent
}

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

// Slice implements a no copy change from string to byte slice.
// Use at your own risk.
func Slice(s string) (b []byte) {
	pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pstring := (*reflect.StringHeader)(unsafe.Pointer(&s))
	pbytes.Data = pstring.Data
	pbytes.Len = pstring.Len
	pbytes.Cap = pstring.Len
	return
}

// String implements a no-copy change from byte slice to string.
// Use at your own risk.
func String(b []byte) (s string) {
	pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pstring := (*reflect.StringHeader)(unsafe.Pointer(&s))
	pstring.Data = pbytes.Data
	pstring.Len = pbytes.Len
	return
}

// ToUniversalPath returns the result of replacing each slash ('/') character
// in "path" with an OS-sepcific separator character. Multiple slashes are
// replaced by multiple separators.
func ToUniversalPath(path string) string {
	return filepath.FromSlash(path)
}

// StringSliceToSet converts a []string to a map[string]struct{},
// necessarily dropping any duplicates.
func StringSliceToSet(strs []string) map[string]struct{} {
	ret := make(map[string]struct{})
	for _, str := range strs {
		ret[str] = struct{}{}
	}
	return ret
}

// CaseInsensitiveEquals compares two strings case insensitively.
func CaseInsensitiveEquals(a, b string) bool {
	return strings.ToLower(a) == strings.ToLower(b)
}
