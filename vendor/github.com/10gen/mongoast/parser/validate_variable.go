package parser

import (
	"unicode"

	"github.com/pkg/errors"
)

func validateVariableName(name string) error {
	runes := []rune(name)
	if !unicode.IsLetter(runes[0]) && runes[0] < 0x80 {
		return errors.Errorf("'%s' starts with an invalid character for a variable name", name)
	}
	for _, r := range runes[1:] {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r < 0x080 {
			return errors.Errorf("'%s' contains an invalid character for a variable name: '%c'", name, r)
		}
	}
	return nil
}
