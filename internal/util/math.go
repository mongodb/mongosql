package util

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/shopspring/decimal"
)

// Numeric Conversion Tools

type converterFunc func(interface{}) (interface{}, error)

var intConverter = newNumberConverter(reflect.TypeOf(int(0)))

var float64Converter = newNumberConverter(reflect.TypeOf(float64(0)))

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

// MaxInt returns the maximum of two integers.
func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MaxUint32 returns the maximum of a and b.
func MaxUint32(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

// MinInt returns the minimum of two integers.
func MinInt(a, b int) int {
	if a > b {
		return b
	}
	return a
}

// MinUint32 returns the minimum of a and b.
func MinUint32(a, b uint32) uint32 {
	if a > b {
		return b
	}
	return a
}

// this helper makes it simple to generate new numeric converters,
// be sure to assign them on a package level instead of dynamically
// within a function to avoid low performance
func newNumberConverter(targetType reflect.Type) converterFunc {
	return func(number interface{}) (interface{}, error) {
		// to avoid panics on nil values
		if number == nil {
			return nil, fmt.Errorf("cannot convert nil value")
		}
		v := reflect.ValueOf(number)
		if !v.Type().ConvertibleTo(targetType) {
			return nil, fmt.Errorf("cannot convert %v to %v", v.Type(), targetType)
		}
		converted := v.Convert(targetType)
		return converted.Interface(), nil
	}
}

// ToFloat64 is a function for converting any numeric type
// into a float64.
func ToFloat64(number interface{}) (float64, error) {
	asInterface, err := float64Converter(number)
	if err != nil {
		return 0, err
	}
	// no check for "ok" here, since we know it will work
	return asInterface.(float64), nil
}

// ToInt is a function for converting any numeric type
// into an int. This can easily result in a loss of information
// due to truncation of floats.
func ToInt(number interface{}) (int, error) {
	asInterface, err := intConverter(number)
	if err != nil {
		return 0, err
	}
	// no check for "ok" here, since we know it will work
	return asInterface.(int), nil
}

// ToUInt32 is a function for converting any numeric type
// into a uint32. This can easily result in a loss of information
// due to truncation, so be careful.
func ToUInt32(number interface{}) (uint32, error) {
	asInterface, err := uint32Converter(number)
	if err != nil {
		return 0, err
	}
	// no check for "ok" here, since we know it will work
	return asInterface.(uint32), nil
}

// making this package level so it is only evaluated once
var uint32Converter = newNumberConverter(reflect.TypeOf(uint32(0)))
