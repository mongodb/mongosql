package parser

import (
	"unicode"

	"github.com/pkg/errors"
)

func validateVariableName(name string) error {
	runes := []rune(name)
	// First character must be [a-z] or non-ASCII.
	if !unicode.IsLower(runes[0]) && runes[0] < 0x80 {
		// CURRENT is a valid variable name even though it breaks the rule above.
		if name == "CURRENT" {
			return nil
		}
		return errors.Errorf("'%s' starts with an invalid character for a user variable name", name)
	}
	// Remaining characters must be [_a-zA-Z0-9] or non-ASCII.
	for _, r := range runes[1:] {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r < 0x080 {
			return errors.Errorf("'%s' contains an invalid character for a variable name: '%c'", name, r)
		}
	}
	return nil
}
