package mathutil

import (
	"fmt"
	"math"
	"reflect"

	"github.com/shopspring/decimal"
)

// Numeric Conversion Tools

type converterFunc func(interface{}) (interface{}, error)

var intConverter = newNumberConverter(reflect.TypeOf(int(0)))

var float64Converter = newNumberConverter(reflect.TypeOf(float64(0)))

// CompareDecimal128 compares two decimals returning < 0 for lt, 0 for eq, and
// > 0 for gt.
func CompareDecimal128(left, right decimal.Decimal) (int, error) {
	return left.Cmp(right), nil
}

// CompareFloats compares two decimals returning < 0 for lt, 0 for eq, and
// > 0 for gt.
func CompareFloats(left, right float64) (int, error) {
	cmp := left - right
	if cmp < 0 {
		return -1, nil
	} else if cmp > 0 {
		return 1, nil
	}
	return 0, nil
}

// CompareInts compares two decimals returning < 0 for lt, 0 for eq, and
// > 0 for gt.
func CompareInts(left, right int) int {
	if left < right {
		return -1
	} else if left > right {
		return 1
	}
	return 0
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

// Round founds a float64 to an int64 using MySQL Rounding conventions (Round
// ties away from 0). This is the simplest implementation of Round I have found.
// https://github.com/golang/go/issues/4594#issuecomment-66073312.
func Round(x float64) int64 {
	if x < 0 {
		return int64(math.Ceil(x - 0.5))
	}
	return int64(math.Floor(x + 0.5))
}

// RoundToDecimalPlaces rounds base to d number of decimal places.
func RoundToDecimalPlaces(d int64, base float64) float64 {
	var rounded float64
	pow := math.Pow(10, float64(d))
	digit := pow * base
	_, div := math.Modf(digit)
	if base > 0 {
		if div >= 0.5 {
			rounded = math.Ceil(digit) / pow
		} else {
			rounded = math.Floor(digit) / pow
		}
	} else {
		if math.Abs(div) >= 0.5 {
			rounded = math.Floor(digit) / pow
		} else {
			rounded = math.Ceil(digit) / pow
		}
	}
	return rounded
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
