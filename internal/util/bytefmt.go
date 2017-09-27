package util

import (
	"fmt"
	"strings"
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
