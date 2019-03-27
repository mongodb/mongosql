package parser

import (
	"testing"
)

func TestValidateVariableName(t *testing.T) {
	testCases := []struct {
		name  string
		valid bool
	}{
		{
			name:  "test_123",
			valid: true,
		},
		{
			name:  "Test_123",
			valid: true,
		},
		{
			name:  "\xFFtest",
			valid: true,
		},
		{
			name:  "test\xFF",
			valid: true,
		},
		{
			name:  "0",
			valid: false,
		},
		{
			name:  "_",
			valid: false,
		},
		{
			name:  ".",
			valid: false,
		},
		{
			name:  "@",
			valid: false,
		},
		{
			name:  "x.y",
			valid: false,
		},
		{
			name:  "x@y",
			valid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateVariableName(tc.name)
			if tc.valid && err != nil {
				t.Fatalf("expected no error, but got %v", err)
			} else if !tc.valid && err == nil {
				t.Fatalf("expected error but got none")
			}
		})
	}
}
