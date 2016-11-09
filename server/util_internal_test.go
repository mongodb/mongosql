package server

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestFormatDecimal(t *testing.T) {

	tests := []string{
		"12300000000000000000000000000000000000000000000000000000000000000000",
		"-12300000000000000000000000000000000000000000000000000000000000000000",
		"1230000",
		"-1230000",
		"123.0000",
		"-123.0000",
		"123.320000",
		"-123.320000",
		"123.032",
		"-123.032",
	}

	for _, e := range tests {
		d, _ := decimal.NewFromString(e)
		if formatDecimal(d) != e {
			t.Fatalf("%s did not roundtrip")
		}
	}
}
