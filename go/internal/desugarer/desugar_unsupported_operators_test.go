package desugarer

import "testing"

func TestConvertSQLPattern(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		escape   rune
		expected string
	}{
		{
			name:     "no special SQL characters",
			input:    "abc",
			escape:   '\\',
			expected: "^abc$",
		},
		{
			name:     "unescaped special SQL characters",
			input:    "a_b%c",
			escape:   '\\',
			expected: "^a.b.*c$",
		},
		{
			name:     "escaped special SQL characters",
			input:    `a\_b\%c`,
			escape:   '\\',
			expected: "^a_b%c$",
		},
		{
			name:     "escaped escape character",
			input:    `a\\_b\%c`,
			escape:   '\\',
			expected: `^a\\.b%c$`,
		},
		{
			name:     "special MQL characters",
			input:    `.^$*+?()[{\|`,
			escape:   'e',
			expected: `^\.\^\$\*\+\?\(\)\[\{\\\|$`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := convertSQLPattern(test.input, test.escape)
			if test.expected != actual {
				t.Fatalf("\n\texpected: %q\n\tbut got:  %q", test.expected, actual)
			}
		})
	}
}
