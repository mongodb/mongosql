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

// Uint128 represents an unsigned 128 bit integer.
type Uint128 struct {
	H uint64
	L uint64
}

var (
	prime = Uint128{
		H: 0x0000000001000000,
		L: 0x000000000000013b,
	}
	primePlus uint64 = 0x00000000000000bf
)

// mult computes the multiplication of two uint64s
// with overflow.
func mult(x, y uint64) (overflow uint64, result uint64) {
	// Computing the lower bits of multiplication
	// is simple, we just multiply. The somewhat
	// trickier part is computing the overflow.
	var lMask uint64 = 0x00000000ffffffff
	// Ultimately, when performing multiplication in any base,
	// overflow digits can only come from when we multiply
	// the upper half of the number add with the overflow from
	// multiplying the lower bits by each other, so we begin by
	// obtaining the upper and lower half of each input.
	xL, yL := x&lMask, y&lMask
	xH, yH := x>>32, y>>32
	// Now obtain the lower bit multiplication overflow
	// so that we know how many bits carry over into
	// the upper portion. The right shift by 32 gives
	// us only the overflow without the trailing result bits.
	xyLOverflow := (xL * yL) >> 32
	// Compute the result of xH*yL, which must include overflow
	// from xL*yL.
	xHyL := xH*yL + xyLOverflow
	// Obtain the upper and lower halfs of xH * yL + xyLOverflow
	xHyLL := xHyL & lMask
	xHyLH := xHyL >> 32
	// Add the result of multiplying xL * yH to the lower part
	// of xHyL.
	xHyLL += xL * yH
	// The last multiplication step for the high bits is to multiply
	// both high sides by each other.
	xHyH := xH * yH
	// The final higher bits are computed by adding
	// all the multiplicands. The lower bits, as mentioned above,
	// are simplying x * y. xHyLL must be right shifted to line
	// up with the other values (note that xHyLH was already
	// right shifted previously).
	return xHyH + xHyLH + xHyLL>>32, x * y
}

// Xor computes the Xor of a Uint128 with a byte. This can
// only affect the lowest byte of the Uint128.
func (h *Uint128) Xor(b byte) {
	h.L ^= uint64(b)
}

// Mult produces the multiplication of the receiver Uint128 with
// a passed Uint128.
func (h *Uint128) Mult(multiplier Uint128) {
	// First multiply the lower half of the two Uint128s,
	// we will need to add the overflow to the high part
	// of the output.
	overflow, res := mult(h.L, multiplier.L)
	// The lower bits of the result
	// are simply the multiplication of the lower halfs
	// of each uint128.
	h.L = res
	// Compute the the result of performing multiplication of the upperhalf
	// of each Uint128 by the lower half of the other and adding them
	// together with the overflow from multiplying the lowerhalfs
	// (recall the algorithm for multiplying large numbers from grade
	// school). We do not need to call mult here because any carry from the
	// upper halves can be discarded as overflow, which does not hurt our
	// ultimate goal of producing 128 bit hashes.
	h.H = overflow + h.H*multiplier.L + h.L*multiplier.H
}

// Plus computes the addition of a uint64 to a Uint128.
// The algorithm is simple because addition never produces
// more than one carry bit, and overflow will result in
// lowerbits than are less than what they were before the
// addition.
func (h *Uint128) Plus(M uint64) {
	old := h.L
	h.L += M
	if h.L < old {
		h.H++
	}
}

// AddByteToHash adds one byte to an FNV-1a hash with prime offset.
func (h *Uint128) AddByteToHash(b byte) {
	h.Mult(prime)
	h.Xor(b)
	h.Plus(primePlus)
}

// AddByteSliceToHash adds a byte slice to an FNV-1a hash with prime offset.
func (h *Uint128) AddByteSliceToHash(slice []byte) {
	for _, b := range slice {
		h.AddByteToHash(b)
	}
}
