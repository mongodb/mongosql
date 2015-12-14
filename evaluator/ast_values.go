package evaluator

import (
	"fmt"
	"math"
	"time"
)

//
// SQLBool represents a boolean.
//
type SQLBool bool

// SQLTrue is a constant SQLBool(true).
const SQLTrue = SQLBool(true)

// SQLFalse is a constant SQLBool(false).
const SQLFalse = SQLBool(false)

func (sb SQLBool) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return sb, nil
}

func (sb SQLBool) CompareTo(v SQLValue) (int, error) {
	if n, ok := v.(SQLBool); ok {
		s1, s2 := bool(sb), bool(n)
		if s1 == s2 {
			return 0, nil
		} else if !s1 {
			return -1, nil
		}
		return 1, nil
	}
	// TODO: support comparing with SQLInt, SQLFloat, etc
	return 1, fmt.Errorf("type mismatch")
}

//
// SQLDate represents a date.
//
type SQLDate struct {
	Time time.Time
}

func (sd SQLDate) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return sd, nil
}

func (sd SQLDate) CompareTo(v SQLValue) (int, error) {
	if cmp, ok := v.(SQLDate); ok {
		if sd.Time.After(cmp.Time) {
			return 1, nil
		} else if sd.Time.Before(cmp.Time) {
			return -1, nil
		}
	}
	// TODO: type sort order implementation
	return 0, nil
}

//
// SQLDateTime represents a datetime.
//
type SQLDateTime struct {
	Time time.Time
}

func (sd SQLDateTime) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return sd, nil
}

func (sd SQLDateTime) CompareTo(v SQLValue) (int, error) {
	if cmp, ok := v.(SQLDate); ok {
		if sd.Time.After(cmp.Time) {
			return 1, nil
		} else if sd.Time.Before(cmp.Time) {
			return -1, nil
		}
	}

	// TODO: type sort order implementation
	return 0, nil
}

//
// SQLFloat represents a float.
//
type SQLFloat float64

func (sf SQLFloat) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return sf, nil
}

func (sf SQLFloat) CompareTo(v SQLValue) (int, error) {
	if n, ok := v.(SQLNumeric); ok {
		cmp := sf.Float64() - n.Float64()
		if cmp > 0 {
			return 1, nil
		} else if cmp < 0 {
			return -1, nil
		}
		return 0, nil
	}
	return -1, nil
}

func (sf SQLFloat) Add(o SQLNumeric) SQLNumeric {
	return SQLFloat(float64(sf) + o.Float64())
}

func (sf SQLFloat) Float64() float64 {
	return float64(sf)
}

func (sf SQLFloat) Product(o SQLNumeric) SQLNumeric {
	return SQLFloat(float64(sf) * o.Float64())
}

func (sf SQLFloat) Sub(o SQLNumeric) SQLNumeric {
	return SQLFloat(float64(sf) - o.Float64())
}

//
// SQLInt represents a 64-bit integer value.
//
type SQLInt int64

func (si SQLInt) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return si, nil
}

func (si SQLInt) CompareTo(v SQLValue) (int, error) {
	if n, ok := v.(SQLInt); ok {
		cmp := int64(si) - int64(n)
		if cmp > 0 {
			return 1, nil
		} else if cmp < 0 {
			return -1, nil
		}
		return 0, nil
	} else if n, ok := v.(SQLFloat); ok {
		cmp := si.Float64() - n.Float64()
		if cmp > 0 {
			return 1, nil
		} else if cmp < 0 {
			return -1, nil
		}
		return 0, nil
	}
	return -1, nil
}

func (si SQLInt) Add(o SQLNumeric) SQLNumeric {
	if oi, ok := o.(SQLInt); ok {
		return SQLInt(int64(si) + int64(oi))
	}
	return SQLFloat(si.Float64() + o.Float64())
}

func (si SQLInt) Float64() float64 {
	return float64(si)
}

func (si SQLInt) Product(o SQLNumeric) SQLNumeric {
	if oi, ok := o.(SQLInt); ok {
		return SQLInt(int64(si) * int64(oi))
	}
	return SQLFloat(si.Float64() * o.Float64())
}

func (si SQLInt) Sub(o SQLNumeric) SQLNumeric {
	if oi, ok := o.(SQLInt); ok {
		return SQLInt(int64(si) - int64(oi))
	}
	return SQLFloat(si.Float64() - o.Float64())
}

//
// SQLNullValue represents a null.
//
type SQLNullValue struct{}

// SQLNull is a constant SQLNullValue.
var SQLNull = SQLNullValue{}

func (nv SQLNullValue) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return nv, nil
}

func (nv SQLNullValue) CompareTo(v SQLValue) (int, error) {
	if _, ok := v.(SQLNullValue); ok {
		return 0, nil
	}
	return 1, nil
}

//
// SQLString represents a string value.
//
type SQLString string

func (ss SQLString) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return ss, nil
}

func (sn SQLString) CompareTo(v SQLValue) (int, error) {
	if n, ok := v.(SQLString); ok {
		s1, s2 := string(sn), string(n)
		if s1 < s2 {
			return -1, nil
		} else if s1 > s2 {
			return 1, nil
		}
		return 0, nil
	}
	// can only compare numbers to each other, otherwise treat as error
	return 1, fmt.Errorf("type mismatch")
}

//
// SQLTime represents a time value.
//
type SQLTime struct {
	Time time.Time
}

func (st SQLTime) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return st, nil
}

func (st SQLTime) CompareTo(v SQLValue) (int, error) {
	if cmp, ok := v.(SQLDate); ok {
		if st.Time.After(cmp.Time) {
			return 1, nil
		} else if st.Time.Before(cmp.Time) {
			return -1, nil
		}
	}
	// TODO: type sort order implementation
	return 0, nil
}

//
// SQLTimestamp represents a timestamp value.
//
type SQLTimestamp struct {
	Time time.Time
}

func (st SQLTimestamp) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return st, nil
}

func (st SQLTimestamp) CompareTo(v SQLValue) (int, error) {
	if cmp, ok := v.(SQLDate); ok {
		if st.Time.After(cmp.Time) {
			return 1, nil
		} else if st.Time.Before(cmp.Time) {
			return -1, nil
		}
	}
	// TODO: type sort order implementation
	return 0, nil
}

//
// SQLValues represents multiple sql values.
//
type SQLValues []SQLValue

func (sv SQLValues) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return sv, nil
}

func (sv SQLValues) CompareTo(v SQLValue) (int, error) {

	r, ok := v.(SQLValues)
	if !ok {
		//
		// allows for implicit row value comparisons such as:
		//
		// select a, b from foo where (a) < 3;
		//
		if len(sv) != 1 {
			return 1, fmt.Errorf("Operand should contain %v columns", len(sv))
		}
		r = append(r, v)
	} else if len(sv) != len(r) {
		return 1, fmt.Errorf("Operand should contain %v columns", len(sv))
	}

	for i := 0; i < len(sv); i++ {

		c, err := sv[i].CompareTo(r[i])
		if err != nil {
			return 1, err
		}

		if c != 0 {
			return c, nil
		}

	}

	return 0, nil
}

//
// SQLUint32 represents an unsigned 32-bit integer.
//
type SQLUint32 uint32

func (su SQLUint32) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return su, nil
}

func (su SQLUint32) CompareTo(v SQLValue) (int, error) {
	if n, ok := v.(SQLUint32); ok {
		cmp := uint32(su) - uint32(n)
		if cmp > 0 {
			return 1, nil
		} else if cmp < 0 {
			return -1, nil
		}
		return 0, nil
	} else if n, ok := v.(SQLFloat); ok {
		cmp := su.Float64() - n.Float64()
		if cmp > 0 {
			return 1, nil
		} else if cmp < 0 {
			return -1, nil
		}
		return 0, nil
	}
	return -1, nil
}

func (su SQLUint32) Add(o SQLNumeric) SQLNumeric {
	if oi, ok := o.(SQLUint32); ok {
		return SQLUint32(uint32(su) + uint32(oi))
	}
	return SQLFloat(su.Float64() + o.Float64())
}

func (su SQLUint32) Float64() float64 {
	return float64(su)
}

func (su SQLUint32) Product(o SQLNumeric) SQLNumeric {
	if oi, ok := o.(SQLUint32); ok {
		return SQLUint32(uint32(su) * uint32(oi))
	}
	return SQLFloat(su.Float64() * o.Float64())
}

func (su SQLUint32) Sub(o SQLNumeric) SQLNumeric {
	if oi, ok := o.(SQLUint32); ok {
		return SQLUint32(uint32(su) - uint32(oi))
	}
	return SQLFloat(su.Float64() - o.Float64())
}

//
// round returns the closest integer value to the float - round half down
// for negative values and round half up otherwise.
func round(f float64) int64 {
	v := f

	if v < 0.0 {
		v += 0.5
	}

	if f < 0 && v == math.Floor(v) {
		return int64(v - 1)
	}

	return int64(math.Floor(v))
}
