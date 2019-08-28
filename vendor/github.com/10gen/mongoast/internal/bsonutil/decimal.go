package bsonutil

import (
	"github.com/ericlagergren/decimal"
	"github.com/ericlagergren/decimal/math"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NewDecimal makes a Decimal.
func NewDecimal(d128 primitive.Decimal128) (Decimal, error) {
	d := decimal.WithContext(decimal.Context128)
	if _, ok := d.SetString(d128.String()); !ok {
		return Decimal{}, errors.Errorf("failed creating Decimal from primitive.Decimal128: %s", d128.String())
	}
	return Decimal{d}, nil
}

var (
	// Decimal0 is the constant 0.
	Decimal0 = Decimal{decimal.WithContext(decimal.Context128)}
	// Decimal1 is the constant 1.
	Decimal1 = Decimal{decimal.WithContext(decimal.Context128).SetUint64(1)}
)

// Decimal represents a native Decimal128 implementation.
type Decimal struct {
	b *decimal.Big
}

// Float64 returns the float64 version of the value.
func (d Decimal) Float64() (float64, bool) {
	return d.b.Float64()
}

// Int64 returns the int64 version of the value.
func (d Decimal) Int64() (int64, bool) {
	if !d.b.IsInt() {
		return 0, false
	}
	return d.b.Int64()
}

// Abs returns the absolute value of d.
func (d Decimal) Abs() Decimal {
	result := decimal.WithContext(decimal.Context128)
	_ = result.Abs(d.b)
	return Decimal{result}
}

// Add returns d + right.
func (d Decimal) Add(right Decimal) Decimal {
	result := decimal.WithContext(decimal.Context128)
	result = result.Add(d.b, right.b)
	return Decimal{result}
}

// Ceil returns the ceiling of d.
func (d Decimal) Ceil() Decimal {
	result := decimal.WithContext(decimal.Context128)
	result = math.Ceil(result, d.b)
	return Decimal{result}
}

// Cmp compares d and right and returns: -1 if x < y 0 if x == y +1 if x > y It does not modify x or y.
func (d Decimal) Cmp(right Decimal) int {
	return d.b.Cmp(right.b)
}

// Div performs d / right.
func (d Decimal) Div(right Decimal) Decimal {
	result := decimal.WithContext(decimal.Context128)
	result = result.Quo(d.b, right.b)
	return Decimal{result}
}

// Exp returns e**a.
func (d Decimal) Exp() Decimal {
	result := decimal.WithContext(decimal.Context128)
	result = math.Exp(result, d.b)
	return Decimal{result}
}

// Floor returns the floor of d.
func (d Decimal) Floor() Decimal {
	result := decimal.WithContext(decimal.Context128)
	result = math.Floor(result, d.b)
	return Decimal{result}
}

// LogBase performs the base base log of d.
func (d Decimal) LogBase(base Decimal) Decimal {
	return d.Log().Div(base.Log())
}

// Log returns the natural log of d.
func (d Decimal) Log() Decimal {
	result := decimal.WithContext(decimal.Context128)
	result = math.Log(result, d.b)
	return Decimal{result}
}

// Log10 returns the base 10 log of d.
func (d Decimal) Log10() Decimal {
	result := decimal.WithContext(decimal.Context128)
	result = math.Log10(result, d.b)
	return Decimal{result}
}

// Mod returns d % right.
func (d Decimal) Mod(right Decimal) Decimal {
	result := decimal.WithContext(decimal.Context128)
	remainder := decimal.WithContext(decimal.Context128)
	_, remainder = result.QuoRem(d.b, right.b, remainder)
	return Decimal{remainder}
}

// Mul returns d * right.
func (d Decimal) Mul(right Decimal) Decimal {
	result := decimal.WithContext(decimal.Context128)
	result = result.Mul(d.b, right.b)
	return Decimal{result}
}

// Pow returns d**exp.
func (d Decimal) Pow(exp Decimal) Decimal {
	result := decimal.WithContext(decimal.Context128)
	result = math.Pow(result, d.b, exp.b)
	return Decimal{result}
}

// RoundToInt returns d rounded down to the nearest int.
func (d Decimal) RoundToInt() Decimal {
	result := decimal.WithContext(decimal.Context128)
	result = result.RoundToInt()
	return Decimal{result}
}

// Sqrt returns the square root of d.
func (d Decimal) Sqrt() Decimal {
	result := decimal.WithContext(decimal.Context128)
	result = math.Sqrt(result, d.b)
	return Decimal{result}
}

// Sub returns d - right.
func (d Decimal) Sub(right Decimal) Decimal {
	result := decimal.WithContext(decimal.Context128)
	result = result.Sub(d.b, right.b)
	return Decimal{result}
}

// Trunc truncates d to the precisision of prec.
func (d Decimal) Trunc(prec int) Decimal {
	result := decimal.WithContext(decimal.Context128)
	ctx := d.b.Context
	ctx.RoundingMode = decimal.ToZero
	result = ctx.Quantize(result.Copy(d.b), prec)
	return Decimal{result}
}
