package evaluator

import (
	"fmt"

	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"

	"github.com/shopspring/decimal"
)

// arithmeticEvalType determines the EvalType of the result of
// an arithmetic operation, given the types of its two operands.
func arithmeticEvalType(left, right types.EvalTyper) types.EvalType {
	preferenceType := preferentialType(left, right)

	if left.EvalType() == types.EvalUint64 || right.EvalType() == types.EvalUint64 {
		preferenceType = types.EvalDecimal128
	}

	switch preferenceType {
	case types.EvalDecimal128: // leave as-is
	case types.EvalDate, types.EvalDatetime, types.EvalInt64, types.EvalInt32:
		// arithmetic between time types results in an integer
		preferenceType = types.EvalInt64
	default:
		preferenceType = types.EvalDouble
	}

	return preferenceType
}

// doArithmetic performs the given arithmetic operation using
// leftVal and rightVal as operands.
func doArithmetic(leftVal, rightVal values.SQLValue, op ArithmeticOperator) (values.SQLValue, error) {
	if leftVal.Kind() != rightVal.Kind() {
		err := fmt.Errorf(
			"left SQLValue and right SQLValue are not of same kind (%x and %x, respectively)",
			leftVal.Kind(), rightVal.Kind(),
		)
		panic(err)
	}
	valueKind := leftVal.Kind()

	preferenceType := arithmeticEvalType(leftVal, rightVal)

	// use decimal type if Float64() value loses precision.
	useDecimal := preferenceType == types.EvalDecimal128 ||
		values.Int64(leftVal) > maxPrecisionInt ||
		values.Int64(rightVal) > maxPrecisionInt ||
		values.Int64(leftVal) < minPrecisionInt ||
		values.Int64(rightVal) < minPrecisionInt

	// match server behavior by performing subtraction of int64s using
	// int64s (this can intentionally lead to overflow).
	if preferenceType == types.EvalInt64 && op == SUB {
		return values.NewSQLInt64(valueKind, values.Int64(leftVal)-values.Int64(rightVal)), nil
	}

	if useDecimal {
		var decimalResult decimal.Decimal
		leftDecimal := values.Decimal(leftVal)
		rightDecimal := values.Decimal(rightVal)
		switch op {
		case ADD:
			decimalResult = leftDecimal.Add(rightDecimal)
		case DIV:
			decimalResult = leftDecimal.Div(rightDecimal)
			// 4 comes from the div_precision_increment variable which
			// we do not allow to be set.
			scale := leftDecimal.Exponent() - 4
			decimalResult = decimalResult.Round(-scale)
			return values.NewSQLDecimal128(valueKind, decimalResult), nil
		case MULT:
			decimalResult = leftDecimal.Mul(rightDecimal)
		case SUB:
			return values.NewSQLDecimal128(valueKind, leftDecimal.Sub(rightDecimal)), nil
		default:
			return nil, fmt.Errorf("unrecognized arithmetic operator: %v", op)
		}

		// perform the operation precisely by using Decimal128s, but
		// then coerce the result into a float to match the server
		// behavior. This is necessary since casting the left and
		// right values to Float64 below would lose precision _before_
		// the operation; the server loses precision _after_ it.
		if preferenceType == types.EvalInt64 && (decimalResult.Cmp(minIntAsDecimal) < 0 || decimalResult.Cmp(maxIntAsDecimal) > 0) {
			f, _ := decimalResult.Float64()
			return values.NewSQLFloat(valueKind, f), nil
		}
		return values.NewSQLDecimal128(valueKind, decimalResult), nil
	}

	valueF := 0.0
	leftFloat := values.Float64(leftVal)
	rightFloat := values.Float64(rightVal)
	switch op {
	case ADD:
		valueF = leftFloat + rightFloat
	case DIV:
		return values.NewSQLFloat(valueKind, leftFloat/rightFloat), nil
	case MULT:
		valueF = leftFloat * rightFloat
	case SUB:
		valueF = leftFloat - rightFloat
	default:
		return nil, fmt.Errorf("unrecognized arithmetic operator: %v", op)
	}

	if preferenceType == types.EvalInt64 {
		return values.NewSQLInt64(valueKind, int64(valueF)), nil
	}

	return values.NewSQLFloat(valueKind, valueF), nil
}
