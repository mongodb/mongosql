package strutil

import "strings"

const punctuation = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"

// IsDigit returns true if this byte is a digit in ASCII/utf-8.
func IsDigit(c byte) bool {
	switch c {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	default:
		return false
	}
}

// IsPunct returns true if this byte is ASCII punctuation in ASCII/utf-8.
func IsPunct(c byte) bool {
	return strings.IndexByte(punctuation, c) != -1
}

// IsSpace returns true if this byte is a space in ASCII/utf-8.
func IsSpace(c byte) bool {
	switch c {
	case ' ':
		return true
	default:
		return false
	}
}
