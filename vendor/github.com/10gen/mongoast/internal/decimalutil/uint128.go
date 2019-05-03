package decimalutil

import (
	"fmt"
	"regexp"
	"strconv"
)

type uint128 struct {
	High uint64
	Low  uint64
}

func (x uint128) Add(y uint128) uint128 {
	high := x.High + y.High
	low := x.Low + y.Low
	if low < x.Low {
		high++
	}
	return uint128{High: high, Low: low}
}

func (x uint128) CompareTo(y uint128) int {
	if x.High < y.High {
		return -1
	} else if x.High > y.High {
		return 1
	} else if x.Low < y.Low {
		return -1
	} else if x.Low > y.Low {
		return 1
	}
	return 0
}

func (x uint128) Divide(divisor uint32) (quotient uint128, remainder uint32) {
	if x.High == 0 && x.Low == 0 {
		return uint128{0, 0}, 0
	}

	a := x.High >> 32
	b := x.High & 0xFFFFFFFF
	c := x.Low >> 32
	d := x.Low & 0xFFFFFFFF

	temp := a
	a = (temp / uint64(divisor)) & 0xFFFFFFFF
	temp = ((temp % uint64(divisor)) << 32) + b
	b = (temp / uint64(divisor)) & 0xFFFFFFFF
	temp = ((temp % uint64(divisor)) << 32) + c
	c = (temp / uint64(divisor)) & 0xFFFFFFFF
	temp = ((temp % uint64(divisor)) << 32) + d
	d = (temp / uint64(divisor)) & 0xFFFFFFFF

	return uint128{
		High: (a << 32) + b,
		Low:  (c << 32) + d,
	}, uint32(temp % uint64(divisor))
}

func (x uint128) Multiply(y uint32) uint128 {
	a := x.High >> 32
	b := x.High & 0xFFFFFFFF
	c := x.Low >> 32
	d := x.Low & 0xFFFFFFFF

	d = d * uint64(y)
	c = c*uint64(y) + (d >> 32)
	b = b*uint64(y) + (c >> 32)
	a = a*uint64(y) + (b >> 32)

	low := (c << 32) + (d & 0xFFFFFFFF)
	high := (a << 32) + (b & 0xFFFFFFFF)

	return uint128{High: high, Low: low}
}

func parseUint128(s string) uint128 {
	value, ok := tryParseUint128(s)
	if !ok {
		panic(fmt.Sprintf("error parsing uint128 string: %q", s))
	}
	return value
}

func tryParseUint128(s string) (uint128, bool) {
	if len(s) == 0 {
		return uint128{}, false
	}

	// remove leading zeros (and return true if value is zero)
	if s[0] == '0' {
		if len(s) == 1 {
			return uint128{0, 0}, true
		} else {
			s = regexp.MustCompile("^0+").ReplaceAllString(s, "")
			if len(s) == 0 {
				return uint128{0, 0}, true
			}
		}
	}

	// parse 9 or fewer decimal digits at a time
	value := uint128{0, 0}
	for len(s) > 0 {
		fragmentSize := len(s) % 9
		if fragmentSize == 0 {
			fragmentSize = 9
		}
		fragmentString := s[0:fragmentSize]

		fragmentValue, err := strconv.ParseUint(fragmentString, 10, 32)
		if err != nil {
			return uint128{}, false
		}

		combinedValue := value.Multiply(1000000000)
		combinedValue = combinedValue.Add(uint128{High: 0, Low: fragmentValue})
		if combinedValue.CompareTo(value) < 0 {
			// overflow means s represents a value larger than the maximum uint128
			return uint128{}, false
		}
		value = combinedValue

		s = s[fragmentSize:]
	}

	return value, true
}
