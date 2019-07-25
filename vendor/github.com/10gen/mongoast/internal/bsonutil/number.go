package bsonutil

import (
	"fmt"
	"math"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// AsInt32 gets the bson.Value as an int32 as long as it's numeric and fits inside 32-bits without losing data.
func AsInt32(v bsoncore.Value) int32 {
	i32, err := asInt32(v)
	if err != nil {
		panic(err.Error())
	}

	return i32
}

// AsInt32OK gets the bson.Value as an int32 as long as it's numeric and fits inside 32-bits without losing data.
func AsInt32OK(v bsoncore.Value) (int32, bool) {
	i32, err := asInt32(v)
	if err != nil {
		return i32, false
	}

	return i32, true
}

func asInt32(v bsoncore.Value) (int32, error) {
	switch v.Type {
	case bsontype.Int32:
		return v.Int32(), nil
	case bsontype.Int64:
		i64 := v.Int64()
		if i64 < math.MinInt32 || i64 > math.MaxInt32 {
			return 0, errors.Errorf("%d overflows int32", i64)
		}
		return int32(v.Int64()), nil
	case bsontype.Double:
		f64 := v.Double()
		if math.Floor(f64) != f64 {
			return 0, errors.Errorf("converting %f to an int32 would require truncation", f64)
		}
		if f64 < float64(math.MinInt32) || f64 > float64(math.MaxInt32) {
			return 0, errors.Errorf("%f overflows int32", f64)
		}

		return int32(v.Double()), nil
	}

	return 0, errors.Errorf("cannot convert %v to an int32", v)
}

// AsInt64 gets the bson.Value as an int64 as long as it's numeric and fits inside 64-bits without losing data.
func AsInt64(v bsoncore.Value) int64 {
	i64, err := asInt64(v)
	if err != nil {
		panic(err.Error())
	}

	return i64
}

// AsInt64OK gets the bson.Value as an int64 as long as it's numeric and fits inside 64-bits without losing data.
func AsInt64OK(v bsoncore.Value) (int64, bool) {
	i64, err := asInt64(v)
	if err != nil {
		return i64, false
	}

	return i64, true
}

// AsDecimal128 gets the bson.Value as a Decimal128.
func AsDecimal128(v bsoncore.Value) primitive.Decimal128 {
	d128, err := asDecimal128(v)
	if err != nil {
		panic(err.Error())
	}

	return d128

}

// AsDecimal128OK gets the bson.Value as a Decimal128.
func AsDecimal128OK(v bsoncore.Value) (primitive.Decimal128, bool) {
	d128, err := asDecimal128(v)
	if err != nil {
		return d128, false
	}

	return d128, true
}

func asInt64(v bsoncore.Value) (int64, error) {
	switch v.Type {
	case bsontype.Int32:
		return int64(v.Int32()), nil
	case bsontype.Int64:
		return v.Int64(), nil
	case bsontype.Double:
		f64 := v.Double()
		if math.Floor(f64) != f64 {
			return 0, errors.Errorf("converting %f to an int64 would require truncation", f64)
		}
		if f64 < float64(math.MinInt64) || f64 > float64(math.MaxInt64) {
			return 0, errors.Errorf("%f overflows int64", f64)
		}

		return int64(v.Double()), nil
	}

	return 0, errors.Errorf("cannot convert %v to an int64", v)
}

func asDecimal128(v bsoncore.Value) (primitive.Decimal128, error) {
	switch v.Type {
	case bsontype.Decimal128:
		return v.Decimal128(), nil
	case bsontype.Int32, bsontype.Int64:
		int64V, err := asInt64(v)
		if err != nil { // This shouldn't error, but it doesn't hurt to keep this check.
			return primitive.Decimal128{}, err
		}
		parsedDecimal128, err := primitive.ParseDecimal128(fmt.Sprintf("%d", int64V))
		if err != nil {
			return primitive.Decimal128{}, err
		}
		return parsedDecimal128, nil
	case bsontype.Double:
		doubleV := v.Double()
		// TODO: This will not always be a perfect conversion.
		parsedDecimal128, err := primitive.ParseDecimal128(fmt.Sprintf("%f", doubleV))
		if err != nil {
			return primitive.Decimal128{}, err
		}
		return parsedDecimal128, nil
	}

	return primitive.Decimal128{}, errors.Errorf("cannot convert %v to a Decimal128", v)
}

// MaxNumberType returns the "largest" type between t1 and t2. If one of types
// is not numeric, this function will return that type
func MaxNumberType(t1 bsontype.Type, t2 bsontype.Type) bsontype.Type {
	switch t1 {
	case bsontype.Int32:
		return t2
	case bsontype.Int64:
		switch t2 {
		case bsontype.Int32:
			return bsontype.Int64
		default:
			return t2
		}
	case bsontype.Double:
		switch t2 {
		case bsontype.Int32, bsontype.Int64:
			return bsontype.Double
		default:
			return t2
		}
	case bsontype.Decimal128:
		switch t2 {
		case bsontype.Int32, bsontype.Int64, bsontype.Double:
			return bsontype.Decimal128
		default:
			return t2
		}
	default:
		return t1
	}

}

// AsFloat64 gets the bson.Value as a Float64.
func AsFloat64(v bsoncore.Value) float64 {
	f64, err := asFloat64(v)
	if err != nil {
		panic(err.Error())
	}

	return f64
}

// AsFloat64OK gets the bson.Value as a Float64.
func AsFloat64OK(v bsoncore.Value) (float64, bool) {
	f64, err := asFloat64(v)
	if err != nil {
		return f64, false
	}

	return f64, true
}

func asFloat64(v bsoncore.Value) (float64, error) {
	switch v.Type {
	case bsontype.Int32:
		return float64(v.Int32()), nil
	case bsontype.Int64:
		return float64(v.Int64()), nil
	case bsontype.Double:
		return v.Double(), nil
	}

	return 0, errors.Errorf("cannot convert %v to a Float64", v)
}
