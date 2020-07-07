package bsonutil

import (
	"math"

	"github.com/10gen/mongoast/internal/mathutil"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// Abs takes the absolute value of a numeric type
func Abs(a bsoncore.Value) (bsoncore.Value, error) {
	switch a.Type {
	case bsontype.Int32:
		aint32 := a.Int32()
		if aint32 < 0 {
			return Int32(aint32 * -1), nil
		}
		return a, nil
	case bsontype.Int64:
		aint64 := a.Int64()
		if aint64 < 0 {
			return Int64(aint64 * -1), nil
		}
		return a, nil
	case bsontype.Double:
		return Double(math.Abs(a.Double())), nil
	case bsontype.Decimal128:
		d, err := NewDecimal(a.Decimal128())
		if err != nil {
			return Null(), err
		}

		return Decimal128(d.Abs()), nil
	case bsontype.Null, bsontype.Undefined:
		return Null(), nil
	default:
		return Null(), errors.Errorf(
			"$abs only supports numeric types, not %s",
			TypeToString(a.Type),
		)
	}
}

// Ceil takes the ceil of a numeric type
func Ceil(a bsoncore.Value) (bsoncore.Value, error) {
	switch a.Type {
	case bsontype.Int32, bsontype.Int64:
		return a, nil
	case bsontype.Double:
		return Double(math.Ceil(a.Double())), nil
	case bsontype.Decimal128:
		d, err := NewDecimal(a.Decimal128())
		if err != nil {
			return Null(), err
		}

		return Decimal128(d.Ceil()), nil
	case bsontype.Null, bsontype.Undefined:
		return Null(), nil
	default:
		return Null(), errors.Errorf(
			"$ceil only supports numeric types, not %s",
			TypeToString(a.Type),
		)
	}
}

// Floor takes the floor of a numeric type
func Floor(a bsoncore.Value) (bsoncore.Value, error) {
	switch a.Type {
	case bsontype.Int32, bsontype.Int64:
		return a, nil
	case bsontype.Double:
		return Double(math.Floor(a.Double())), nil
	case bsontype.Decimal128:
		d, err := NewDecimal(a.Decimal128())
		if err != nil {
			return Null(), err
		}

		return Decimal128(d.Floor()), nil
	case bsontype.Null, bsontype.Undefined:
		return Null(), nil
	default:
		return Null(), errors.Errorf(
			"$floor only supports numeric types, not %s",
			TypeToString(a.Type),
		)
	}
}

// Exp raises e to the numeric argument
func Exp(a bsoncore.Value) (bsoncore.Value, error) {
	switch a.Type {
	case bsontype.Int32, bsontype.Int64, bsontype.Double:
		afloat64 := AsFloat64(a)
		return Double(math.Exp(afloat64)), nil
	case bsontype.Decimal128:
		d, err := NewDecimal(a.Decimal128())
		if err != nil {
			return Null(), err
		}

		return Decimal128(d.Exp()), nil
	case bsontype.Null, bsontype.Undefined:
		return Null(), nil
	default:
		return Null(), errors.Errorf(
			"$exp only supports numeric types, not %s",
			TypeToString(a.Type),
		)
	}
}

// Ln takes the natural logarithm of a numeric type
func Ln(a bsoncore.Value) (bsoncore.Value, error) {
	switch a.Type {
	case bsontype.Int32, bsontype.Int64, bsontype.Double:
		afloat64 := AsFloat64(a)
		if afloat64 <= 0 {
			return Null(), errors.Errorf("$ln's argument must be a positive number, but is %g", afloat64)
		}
		// math.Log returns the natural logarithm
		return Double(math.Log(afloat64)), nil
	case bsontype.Decimal128:
		d, err := NewDecimal(a.Decimal128())
		if err != nil {
			return Null(), err
		}
		if d.Cmp(Decimal0) <= 0 {
			return Null(), errors.Errorf("$ln's argument must be a positive number, but is %s", d.String())
		}

		return Decimal128(d.Log()), nil
	case bsontype.Null, bsontype.Undefined:
		return Null(), nil
	default:
		return Null(), errors.Errorf(
			"$ln only supports numeric types, not %s",
			TypeToString(a.Type),
		)
	}
}

// Log10 takes the base 10 logarithm of a numeric type
func Log10(a bsoncore.Value) (bsoncore.Value, error) {
	switch a.Type {
	case bsontype.Int32, bsontype.Int64, bsontype.Double:
		afloat64 := AsFloat64(a)
		if afloat64 <= 0 {
			return Null(), errors.Errorf("$log10's argument must be a positive number, but is %g", afloat64)
		}
		return Double(math.Log10(afloat64)), nil
	case bsontype.Decimal128:
		d, err := NewDecimal(a.Decimal128())
		if err != nil {
			return Null(), err
		}
		if d.Cmp(Decimal0) <= 0 {
			return Null(), errors.Errorf("$log10's argument must be a positive number, but is %s", d.String())
		}

		return Decimal128(d.Log10()), nil
	case bsontype.Null, bsontype.Undefined:
		return Null(), nil
	default:
		return Null(), errors.Errorf(
			"$log10 only supports numeric types, not %s",
			TypeToString(a.Type),
		)
	}
}

// Sqrt takes the square root of a numeric type
func Sqrt(a bsoncore.Value) (bsoncore.Value, error) {
	switch a.Type {
	case bsontype.Int32, bsontype.Int64, bsontype.Double:
		afloat64 := AsFloat64(a)
		if afloat64 < 0 {
			return Null(), errors.Errorf("$sqrt's argument must be greater than or equal to 0")
		}
		return Double(math.Sqrt(afloat64)), nil
	case bsontype.Decimal128:
		d, err := NewDecimal(a.Decimal128())
		if err != nil {
			return Null(), err
		}

		return Decimal128(d.Sqrt()), nil
	case bsontype.Null, bsontype.Undefined:
		return Null(), nil
	default:
		return Null(), errors.Errorf(
			"$sqrt only supports numeric types, not %s",
			TypeToString(a.Type),
		)
	}
}

// Add adds two numeric types together
func Add(a bsoncore.Value, b bsoncore.Value) (bsoncore.Value, error) {
	if a.Type == bsontype.DateTime || b.Type == bsontype.DateTime {
		return addDate(a, b)
	}
	t := MaxNumberType(a.Type, b.Type)
	switch t {
	case bsontype.Int32:
		aint32 := a.Int32()
		bint32 := b.Int32()
		result32, ok := mathutil.Add32(aint32, bint32)
		if ok {
			return Int32(result32), nil
		}
		result64, ok := mathutil.Add64(int64(aint32), int64(bint32))
		if ok {
			return Int64(result64), nil
		}
		return Double(AsFloat64(a) + AsFloat64(b)), nil
	case bsontype.Int64:
		aint64 := AsInt64(a)
		bint64 := AsInt64(b)
		result64, ok := mathutil.Add64(aint64, bint64)
		if ok {
			return Int64(result64), nil
		}
		return Double(AsFloat64(a) + AsFloat64(b)), nil
	case bsontype.Double:
		return Double(AsFloat64(a) + AsFloat64(b)), nil
	case bsontype.Decimal128:
		adec := AsDecimal(a)
		bdec := AsDecimal(b)
		return Decimal128(adec.Add(bdec)), nil
	case bsontype.Null, bsontype.Undefined:
		return Null(), nil
	default:
		return Null(), errors.Errorf(
			"$add only supports numeric or date types, not %s",
			TypeToString(t),
		)
	}
}

func addDate(a bsoncore.Value, b bsoncore.Value) (bsoncore.Value, error) {
	switch a.Type {
	case bsontype.DateTime:
		aint64 := a.DateTime()
		switch b.Type {
		case bsontype.Int32:
			bint32 := b.Int32()
			return DateTime(aint64 + int64(bint32)), nil
		case bsontype.Int64:
			bint64 := b.Int64()
			return DateTime(aint64 + bint64), nil
		case bsontype.Double:
			bdouble := b.Double()
			return DateTime(aint64 + int64(math.Round(bdouble))), nil
		case bsontype.Decimal128:
			bint64, _ := AsDecimal(b).RoundToInt().Int64()
			return DateTime(aint64 + bint64), nil
		case bsontype.DateTime:
			return Null(), errors.New("only one date allowed in an $add expression")
		default:
			return Null(), errors.Errorf(
				"$add only supports numeric or date types, not %s",
				TypeToString(b.Type),
			)
		}
	case bsontype.Int32:
		aint32 := a.Int32()
		bint64 := b.DateTime()
		return DateTime(int64(aint32) + bint64), nil
	case bsontype.Int64:
		aint64 := a.Int64()
		bint64 := b.DateTime()
		return DateTime(aint64 + bint64), nil
	case bsontype.Double:
		adouble := a.Double()
		bint64 := b.DateTime()
		return DateTime(int64(math.Round(adouble)) + bint64), nil
	case bsontype.Decimal128:
		aint64, _ := AsDecimal(a).RoundToInt().Int64()
		bint64 := b.DateTime()
		return DateTime(aint64 + bint64), nil
	case bsontype.Null, bsontype.Undefined:
		return Null(), nil
	default:
		return Null(), errors.Errorf(
			"$add only supports numeric or date types, not %s",
			TypeToString(a.Type),
		)
	}
}

// Sub subtracts two numeric types
func Sub(a bsoncore.Value, b bsoncore.Value) (bsoncore.Value, error) {
	if a.Type == bsontype.DateTime {
		return subDate(a, b)
	}
	switch MaxNumberType(a.Type, b.Type) {
	case bsontype.Int32:
		aint32 := a.Int32()
		bint32 := b.Int32()
		result32, ok := mathutil.Sub32(aint32, bint32)
		if ok {
			return Int32(result32), nil
		}
		result64, ok := mathutil.Sub64(int64(aint32), int64(bint32))
		if ok {
			return Int64(result64), nil
		}
		return Double(AsFloat64(a) - AsFloat64(b)), nil
	case bsontype.Int64:
		aint64 := AsInt64(a)
		bint64 := AsInt64(b)
		result64, ok := mathutil.Sub64(aint64, bint64)
		if ok {
			return Int64(result64), nil
		}
		return Double(AsFloat64(a) - AsFloat64(b)), nil
	case bsontype.Double:
		return Double(AsFloat64(a) - AsFloat64(b)), nil
	case bsontype.Decimal128:
		adec := AsDecimal(a)
		bdec := AsDecimal(b)
		return Decimal128(adec.Sub(bdec)), nil
	case bsontype.Null, bsontype.Undefined:
		return Null(), nil
	default:
		return Null(), errors.Errorf(
			"can't $subtract a %s from a %s",
			TypeToString(b.Type),
			TypeToString(a.Type),
		)
	}
}

func subDate(a bsoncore.Value, b bsoncore.Value) (bsoncore.Value, error) {
	aint64 := a.DateTime()
	switch b.Type {
	case bsontype.DateTime:
		aint64 := a.DateTime()
		bint64 := b.DateTime()
		return Int64(aint64 - bint64), nil
	case bsontype.Int32:
		bint32 := b.Int32()
		return DateTime(aint64 - int64(bint32)), nil
	case bsontype.Int64:
		bint64 := b.Int64()
		return DateTime(aint64 - bint64), nil
	case bsontype.Double:
		bdouble := b.Double()
		return DateTime(aint64 - int64(math.Round(bdouble))), nil
	case bsontype.Decimal128:
		bint64, _ := AsDecimal(b).RoundToInt().Int64()
		return DateTime(aint64 - bint64), nil
	case bsontype.Null, bsontype.Undefined:
		return Null(), nil
	default:
		return Null(), errors.Errorf(
			"can't $subtract a %s from a %s",
			TypeToString(b.Type),
			TypeToString(a.Type),
		)
	}
}

// Mul multiplies two numeric types
func Mul(a bsoncore.Value, b bsoncore.Value) (bsoncore.Value, error) {
	t := MaxNumberType(a.Type, b.Type)
	switch t {
	case bsontype.Int32:
		aint32 := a.Int32()
		bint32 := b.Int32()
		result32, ok := mathutil.Mul32(aint32, bint32)
		if ok {
			return Int32(result32), nil
		}
		result64, ok := mathutil.Mul64(int64(aint32), int64(bint32))
		if ok {
			return Int64(result64), nil
		}
		return Double(AsFloat64(a) * AsFloat64(b)), nil
	case bsontype.Int64:
		aint64 := AsInt64(a)
		bint64 := AsInt64(b)
		result64, ok := mathutil.Mul64(aint64, bint64)
		if ok {
			return Int64(result64), nil
		}
		return Double(AsFloat64(a) * AsFloat64(b)), nil
	case bsontype.Double:
		return Double(AsFloat64(a) * AsFloat64(b)), nil
	case bsontype.Decimal128:
		adec := AsDecimal(a)
		bdec := AsDecimal(b)
		return Decimal128(adec.Mul(bdec)), nil
	case bsontype.Null, bsontype.Undefined:
		return Null(), nil
	default:
		return Null(), errors.Errorf(
			"$multiply only supports numeric types, not %s",
			TypeToString(t),
		)
	}
}

// Div divides two numeric types
func Div(a bsoncore.Value, b bsoncore.Value) (bsoncore.Value, error) {
	afloat64, ok1 := AsFloat64OK(a)
	bfloat64, ok2 := AsFloat64OK(b)

	if !ok1 || !ok2 {
		switch MaxNumberType(a.Type, b.Type) {
		case bsontype.Null, bsontype.Undefined:
			return Null(), nil
		default:
			return Null(), errors.Errorf(
				"$divide only supports numeric types, not %s and %s",
				TypeToString(a.Type),
				TypeToString(b.Type),
			)
		}
	}

	if bfloat64 == 0 {
		return Null(), errors.Errorf("can't $divide by zero")
	}

	divResult := afloat64 / bfloat64

	if a.Type == bsontype.Decimal128 || b.Type == bsontype.Decimal128 {
		return Decimal128FromFloat64(divResult), nil
	}

	return Double(divResult), nil
}

// Mod takes the mod of two numeric types
func Mod(a bsoncore.Value, b bsoncore.Value) (bsoncore.Value, error) {
	switch MaxNumberType(a.Type, b.Type) {
	case bsontype.Int32:
		aint32 := a.Int32()
		bint32 := b.Int32()
		if bint32 == 0 {
			return Null(), errors.Errorf("can't $mod by zero")
		}
		return Int32(aint32 % bint32), nil
	case bsontype.Int64:
		aint64 := AsInt64(a)
		bint64 := AsInt64(b)
		if bint64 == 0 {
			return Null(), errors.Errorf("can't $mod by zero")
		}
		return Int64(aint64 % bint64), nil
	case bsontype.Double:
		afloat := AsFloat64(a)
		bfloat := AsFloat64(b)
		if bfloat == 0 {
			return Null(), errors.Errorf("can't $mod by zero")
		}
		return Double(math.Mod(afloat, bfloat)), nil
	case bsontype.Decimal128:
		adec := AsDecimal(a)
		bdec := AsDecimal(b)
		return Decimal128(adec.Mod(bdec)), nil
	case bsontype.Null, bsontype.Undefined:
		return Null(), nil
	default:
		return Null(), errors.Errorf(
			"$mod only supports numeric types, not %s and %s",
			TypeToString(a.Type),
			TypeToString(b.Type),
		)
	}
}

// Log takes the log of a base b of two numeric types
func Log(a bsoncore.Value, b bsoncore.Value) (bsoncore.Value, error) {
	t := MaxNumberType(a.Type, b.Type)
	switch t {
	case bsontype.Int32, bsontype.Int64, bsontype.Double:
		afloat := AsFloat64(a)
		bfloat := AsFloat64(b)
		if afloat <= 0 {
			return Null(), errors.Errorf("$log's argument must be a positive number, but is %g", afloat)
		}

		if bfloat <= 0 || bfloat == 1 {
			return Null(), errors.Errorf("$log's base must be a positive number not equal to 1, but is %g", bfloat)
		}
		return Double(math.Log(afloat) / math.Log(bfloat)), nil
	case bsontype.Decimal128:
		adec := AsDecimal(a)
		if adec.Cmp(Decimal0) <= 0 {
			return Null(), errors.Errorf("$log's argument must be a positive number, but is %v", adec.b)
		}
		bdec := AsDecimal(b)
		if bdec.Cmp(Decimal0) <= 0 || bdec.Cmp(Decimal1) == 0 {
			return Null(), errors.Errorf("$log's base must be a positive number not equal to 1, but is %v", bdec)
		}

		return Decimal128(adec.LogBase(bdec)), nil
	case bsontype.Null, bsontype.Undefined:
		return Null(), nil
	default:
		return Null(), errors.Errorf(
			"$log's argument must be numeric, not %s",
			TypeToString(t),
		)
	}
}

// Pow raises a to the power of b for two numeric types
func Pow(a bsoncore.Value, b bsoncore.Value) (bsoncore.Value, error) {
	switch MaxNumberType(a.Type, b.Type) {
	case bsontype.Int32:
		// Although it would be more precise to calculate pow for ints manually rather than converting
		// them to floats, the server does the same thing, so this should not be any less precise.
		afloat := AsFloat64(a)
		bfloat := AsFloat64(b)
		if afloat == 0 && bfloat < 0 {
			return Null(), errors.Errorf("$pow cannot take a base of 0 and a negative exponent")
		}
		floatResult := math.Pow(afloat, bfloat)
		if math.Floor(floatResult) != floatResult {
			return Double(floatResult), nil
		}

		if floatResult > float64(math.MinInt32) && floatResult < float64(math.MaxInt32) {
			return Int32(int32(floatResult)), nil
		}
		if floatResult > float64(math.MinInt64) && floatResult < float64(math.MaxInt64) {
			return Int64(int64(floatResult)), nil
		}
		return Double(floatResult), nil
	case bsontype.Int64:
		afloat := AsFloat64(a)
		bfloat := AsFloat64(b)
		if afloat == 0 && bfloat < 0 {
			return Null(), errors.Errorf("$pow cannot take a base of 0 and a negative exponent")
		}
		floatResult := math.Pow(afloat, bfloat)
		if math.Floor(floatResult) != floatResult || floatResult < float64(math.MinInt64) || floatResult > float64(math.MaxInt64) {
			return Double(floatResult), nil
		}
		return Int64(int64(floatResult)), nil
	case bsontype.Double:
		afloat := AsFloat64(a)
		bfloat := AsFloat64(b)
		if afloat == 0 && bfloat < 0 {
			return Null(), errors.Errorf("$pow cannot take a base of 0 and a negative exponent")
		}
		floatResult := math.Pow(afloat, bfloat)
		return Double(floatResult), nil
	case bsontype.Decimal128:
		adec := AsDecimal(a)
		bdec := AsDecimal(b)
		return Decimal128(adec.Pow(bdec)), nil
	case bsontype.Null, bsontype.Undefined:
		return Null(), nil
	default:
		switch a.Type {
		case bsontype.Int32, bsontype.Int64, bsontype.Double:
			return Null(), errors.Errorf(
				"$pow's exponent must be numeric, not %s",
				TypeToString(b.Type),
			)
		default:
			return Null(), errors.Errorf(
				"$pow's base must be numeric, not %s",
				TypeToString(a.Type),
			)
		}
	}
}

// Trunc truncates a numeric type to a precision of b
func Trunc(a bsoncore.Value, b bsoncore.Value) (bsoncore.Value, error) {
	if a.Type == bsontype.Decimal128 || b.Type == bsontype.Decimal128 {
		adec := AsDecimal(a)
		bint32 := AsInt32(b)
		return Decimal128(adec.Trunc(int(bint32))), nil
	}

	var precision float64
	switch b.Type {
	case bsontype.Int32, bsontype.Int64:
		precision = AsFloat64(b)
	default:
		return Null(), errors.Errorf(
			"precision argument to $trunc must be an integral value, not %s",
			TypeToString(b.Type),
		)
	}

	var n float64
	switch a.Type {
	case bsontype.Int32, bsontype.Int64, bsontype.Double:
		n = AsFloat64(a)
	case bsontype.Null, bsontype.Undefined:
		return Null(), nil
	default:
		return Null(), errors.Errorf(
			"$trunc only supports numeric types, not %s",
			TypeToString(a.Type),
		)
	}

	var truncated float64
	if precision >= 0 {
		pow := math.Pow(10, precision)
		i, _ := math.Modf(n * pow)
		truncated = i / pow
	} else {
		pow := math.Pow(10, math.Abs(precision))
		i, _ := math.Modf(n / pow)
		truncated = i * pow
	}

	return Double(truncated), nil
}
