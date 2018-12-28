package strutil

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// ByteString takes a count of bytes and creates
// a human readable version.
func ByteString(count uint64) string {
	const (
		kilo = 1024.0
		mega = 1024 * kilo
		giga = 1024 * mega
		tera = 1024 * giga
	)

	unit := ""
	value := float32(count)

	switch {
	case count >= tera:
		unit = "TiB"
		value = value / giga
	case count >= giga:
		unit = "GiB"
		value = value / giga
	case count >= mega:
		unit = "MiB"
		value = value / mega
	case count >= kilo:
		unit = "KiB"
		value = value / kilo
	case count >= 1.0:
		unit = "B"
	case count == 0:
		return "0B"
	}

	s := fmt.Sprintf("%.1f", value)
	s = strings.TrimSuffix(s, ".0")
	return fmt.Sprintf("%s%s", s, unit)
}

// FormatDecimal formats a decimal into a string.
func FormatDecimal(d decimal.Decimal) string {

	exp := int(d.Exponent())
	if exp >= 0 {
		return d.String()
	}

	str := d.String()
	sign := d.Cmp(decimal.Zero) < 0
	if sign {
		str = str[1:]
	}

	var relExp int
	idx := strings.Index(str, ".")
	if idx >= 0 {
		relExp = exp + (len(str) - 1 - idx)
	} else {
		relExp = exp
		str += "."
	}

	if relExp < 0 {
		str += strings.Repeat("0", -relExp)
	}

	if sign {
		return "-" + str
	}

	return str
}
