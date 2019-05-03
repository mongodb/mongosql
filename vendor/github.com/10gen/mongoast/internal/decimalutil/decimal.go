package decimalutil

import (
	"math"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// This package provides some basic support for decimal operations on a
// temporary basis until they are added to the Go driver. It is adapted from the
// Decimal128 class in the C# driver.

// Decimal128 represents a decimal value internally for performing arithmetic. It must
// be converted to primitive.Decimal128 from the Go driver in order to be serialized to
// the database.
type Decimal128 struct {
	high uint64
	low  uint64
}

// IsZero checks if a decimal value is zero.
func IsZero(x Decimal128) bool {
	if isFirstForm(x.high) {
		sig := getSignificand(x)
		return sig.High == 0 && sig.Low == 0
	} else if isSecondForm(x.high) {
		return true
	}
	return false
}

// Negate reverses the sign of a decimal value.
func Negate(x Decimal128) Decimal128 {
	return Decimal128{high: x.high ^ signBit, low: x.low}
}

// Compare compares two decimal values.
func Compare(x, y Decimal128) int {
	xType := getDecimal128Type(x)
	yType := getDecimal128Type(y)
	if xType < yType {
		return -1
	} else if xType > yType {
		return 1
	} else if xType == decimal128TypeNumber {
		return compareNumbers(x, y)
	}
	return 0
}

// FromIEEEBits creates a new decimal value from the IEEE encoding bits.
func FromIEEEBits(highBits, lowBits uint64) Decimal128 {
	return Decimal128{high: mapIEEEHighBitsToDecimal128HighBits(highBits), low: lowBits}
}

// FromPrimitive creates a new decimal value from the Go driver primitive type.
func FromPrimitive(value primitive.Decimal128) Decimal128 {
	highBits, lowBits := value.GetBytes()
	return FromIEEEBits(highBits, lowBits)
}

// FromInt32 creates a new decimal value from a 32-bit integer.
func FromInt32(value int32) Decimal128 {
	if value >= 0 {
		return Decimal128{high: 0, low: uint64(value)}
	} else if value == math.MinInt32 {
		return Decimal128{high: signBit, low: uint64(math.MaxInt32) + 1}
	}
	return Decimal128{high: signBit, low: uint64(-value)}
}

// FromInt64 creates a new decimal value from a 64-bit integer.
func FromInt64(value int64) Decimal128 {
	if value >= 0 {
		return Decimal128{high: 0, low: uint64(value)}
	} else if value == math.MinInt64 {
		return Decimal128{high: signBit, low: uint64(math.MaxInt64) + 1}
	}
	return Decimal128{high: signBit, low: uint64(-value)}
}

var (
	maxSignificand = parseUint128("9999999999999999999999999999999999")
)

type decimal128Type int

const (
	decimal128TypeNaN decimal128Type = iota
	decimal128TypeNegativeInfinity
	decimal128TypeNumber
	decimal128TypePositiveInfinity
)

type numberClass int

const (
	numberClassNegative numberClass = iota
	numberClassZero
	numberClassPositive
)

func getDecimal128Type(x Decimal128) decimal128Type {
	if isNaN(x) {
		return decimal128TypeNaN
	} else if isNegativeInfinity(x) {
		return decimal128TypeNegativeInfinity
	} else if isPositiveInfinity(x) {
		return decimal128TypePositiveInfinity
	}
	return decimal128TypeNumber
}

func compareNumbers(x, y Decimal128) int {
	xClass := getNumberClass(x)
	yClass := getNumberClass(y)
	if xClass < yClass {
		return -1
	} else if xClass > yClass {
		return 1
	} else if xClass == numberClassNegative {
		return compareNegativeNumbers(x, y)
	} else if xClass == numberClassPositive {
		return comparePositiveNumbers(x, y)
	}
	return 0
}

func getNumberClass(x Decimal128) numberClass {
	if IsZero(x) {
		return numberClassZero
	} else if isNegative(x) {
		return numberClassNegative
	}
	return numberClassPositive
}

func compareNegativeNumbers(x, y Decimal128) int {
	return -comparePositiveNumbers(Negate(x), Negate(y))
}

func comparePositiveNumbers(x, y Decimal128) int {
	xExponent := getExponent(x)
	xSignificand := getSignificand(x)
	yExponent := getExponent(y)
	ySignificand := getSignificand(y)

	exponentDifference := xExponent - yExponent
	if exponentDifference >= -66 && exponentDifference <= 66 {
		// we may or may not be able to make an exponent equal but we we won't know until we try
		// but we know we can't eliminate an exponent difference larger than 66
		if xExponent < yExponent {
			xSignificand, xExponent = tryIncreaseExponent(xSignificand, xExponent, yExponent)
			ySignificand, yExponent = tryDecreaseExponent(ySignificand, yExponent, xExponent)
		} else if xExponent > yExponent {
			xSignificand, xExponent = tryDecreaseExponent(xSignificand, xExponent, yExponent)
			ySignificand, yExponent = tryIncreaseExponent(ySignificand, yExponent, xExponent)
		}
	}

	if xExponent < yExponent {
		return -1
	} else if xExponent > yExponent {
		return 1
	}
	return xSignificand.CompareTo(ySignificand)
}

const (
	signBit                  uint64 = 0x8000000000000000
	firstFormLeadingBits     uint64 = 0x6000000000000000
	firstFormLeadingBitsMax  uint64 = 0x4000000000000000
	firstFormExponentBits    uint64 = 0x7FFE000000000000
	firstFormSignificandBits uint64 = 0x0001FFFFFFFFFFFF
	secondFormLeadingBits    uint64 = 0x7800000000000000
	secondFormLeadingBitsMin uint64 = 0x6000000000000000
	secondFormLeadingBitsMax uint64 = 0x7000000000000000
	secondFormExponentBits   uint64 = 0x1FFF800000000000
	signedInfinityBits       uint64 = 0xFC00000000000000
	positiveInfinity         uint64 = 0x7800000000000000
	negativeInfinity         uint64 = 0xF800000000000000
	partialNaNBits           uint64 = 0x7C00000000000000
	partialNaN               uint64 = 0x7C00000000000000
)

func isFirstForm(highBits uint64) bool {
	return (highBits & firstFormLeadingBits) <= firstFormLeadingBitsMax
}

func isNaN(x Decimal128) bool {
	return (x.high & partialNaNBits) == partialNaN
}

func isNegative(x Decimal128) bool {
	return (x.high & signBit) != 0
}

func isNegativeInfinity(x Decimal128) bool {
	return (x.high & signedInfinityBits) == negativeInfinity
}

func isPositiveInfinity(x Decimal128) bool {
	return (x.high & signedInfinityBits) == positiveInfinity
}

func isSecondForm(highBits uint64) bool {
	bits := highBits & secondFormLeadingBits
	return bits >= secondFormLeadingBitsMin && bits <= secondFormLeadingBitsMax
}

func getExponent(x Decimal128) int16 {
	if isFirstForm(x.high) {
		return mapDecimal128BiasedExponentToExponent(uint16((x.high & firstFormExponentBits) >> 49))
	} else if isSecondForm(x.high) {
		return mapDecimal128BiasedExponentToExponent(uint16((x.high & secondFormExponentBits) >> 47))
	}
	panic("getExponent cannot be called for Infinity or NaN")
}

func getSignificand(x Decimal128) uint128 {
	if isFirstForm(x.high) {
		return uint128{
			High: x.high & firstFormSignificandBits,
			Low:  x.low,
		}
	} else if isSecondForm(x.high) {
		return uint128{0, 0}
	}
	panic("getSignificand cannot be called for Infinity or NaN")
}

func tryIncreaseExponent(significand uint128, exponent int16, goal int16) (uint128, int16) {
	if significand.High == 0 && significand.Low == 0 {
		return significand, goal
	}

	for exponent < goal {
		significandDividedBy10, remainder := significand.Divide(10)
		if remainder != 0 {
			break
		}
		exponent++
		significand = significandDividedBy10
	}
	return significand, exponent
}

func tryDecreaseExponent(significand uint128, exponent int16, goal int16) (uint128, int16) {
	if significand.High == 0 && significand.Low == 0 {
		return significand, goal
	}

	for exponent > goal {
		significandTimes10 := significand.Multiply(10)
		if significandTimes10.CompareTo(maxSignificand) > 0 {
			break
		}
		exponent--
		significand = significandTimes10
	}
	return significand, exponent
}

func mapDecimal128BiasedExponentToExponent(biasedExponent uint16) int16 {
	if biasedExponent < 6111 {
		return int16(biasedExponent)
	}
	return int16(int32(biasedExponent) - 12288)
}

func mapIEEEHighBitsToDecimal128HighBits(highBits uint64) uint64 {
	if isFirstForm(highBits) {
		exponentBits := highBits & firstFormExponentBits
		if exponentBits <= (6175 << 49) {
			return highBits + (6112 << 49)
		}
		return highBits - (6176 << 49)
	} else if isSecondForm(highBits) {
		exponentBits := highBits & secondFormExponentBits
		if exponentBits <= (6175 << 47) {
			return highBits + (6112 << 47)
		}
		return highBits - (6176 << 47)
	}
	return highBits
}
