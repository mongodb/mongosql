package evaluator

import "fmt"

// doArithmetic performs the given arithmetic operation using
// leftVal and rightVal as operands.
func doArithmetic(leftVal, rightVal SQLValue, op ArithmeticOperator) (SQLValue, error) {
	if leftVal.Kind() != rightVal.Kind() {
		err := fmt.Errorf(
			"left SQLValue and right SQLValue are not of same kind (%x and %x, respectively)",
			leftVal.Kind(), rightVal.Kind(),
		)
		panic(err)
	}
	valueKind := leftVal.Kind()

	preferenceType := preferentialType(leftVal, rightVal)
	useDecimal := preferenceType == EvalDecimal128

	leftType := leftVal.EvalType()
	rightType := rightVal.EvalType()

	hasUnsigned := leftType == EvalUint64 || rightType == EvalUint64

	if hasUnsigned {
		useDecimal = true
		preferenceType = EvalDecimal128
	}

	// check if both operands are timestamp or date since
	// arithmetic between time types result in an integer
	if preferenceType == EvalDate || preferenceType == EvalDatetime {
		preferenceType = EvalInt64
	}

	if preferenceType == EvalBoolean {
		preferenceType = EvalDouble
	}

	// use decimal type if Float64() value loses precision
	useDecimal = useDecimal ||
		Int64(leftVal) > maxPrecisionInt ||
		Int64(rightVal) > maxPrecisionInt

	if useDecimal {
		leftDecimal := Decimal(leftVal)
		rightDecimal := Decimal(rightVal)
		switch op {
		case ADD:
			return NewSQLDecimal128(valueKind, leftDecimal.Add(rightDecimal)), nil
		case DIV:
			decimalResult := leftDecimal.Div(rightDecimal)
			// 4 comes from the div_precision_increment variable which
			// we do not allow to be set.
			scale := leftDecimal.Exponent() - 4
			decimalResult = decimalResult.Round(-scale)
			return NewSQLDecimal128(valueKind, decimalResult), nil
		case MULT:
			return NewSQLDecimal128(valueKind, leftDecimal.Mul(rightDecimal)), nil
		case SUB:
			return NewSQLDecimal128(valueKind, leftDecimal.Sub(rightDecimal)), nil
		default:
			return nil, fmt.Errorf("unrecognized arithmetic operator: %v", op)
		}
	}

	valueF := 0.0
	leftFloat := Float64(leftVal)
	rightFloat := Float64(rightVal)
	switch op {
	case ADD:
		valueF = leftFloat + rightFloat
	case DIV:
		floatResult := leftFloat / rightFloat
		return NewSQLFloat(valueKind, floatResult), nil
	case MULT:
		valueF = leftFloat * rightFloat
	case SUB:
		valueF = leftFloat - rightFloat
	default:
		return nil, fmt.Errorf("unrecognized arithmetic operator: %v", op)
	}

	switch preferenceType {
	case EvalInt64, EvalInt32:
		return NewSQLInt64(valueKind, int64(valueF)), nil
	case EvalDouble:
		return NewSQLFloat(valueKind, valueF), nil
	}
	return NewSQLFloat(valueKind, valueF), nil
}
