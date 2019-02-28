package evaluator

import (
	"fmt"

	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
)

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

	preferenceType := preferentialType(leftVal, rightVal)
	useDecimal := preferenceType == types.EvalDecimal128

	leftType := leftVal.EvalType()
	rightType := rightVal.EvalType()

	hasUnsigned := leftType == types.EvalUint64 || rightType == types.EvalUint64

	if hasUnsigned {
		useDecimal = true
		preferenceType = types.EvalDecimal128
	}

	// check if both operands are timestamp or date since
	// arithmetic between time types result in an integer
	if preferenceType == types.EvalDate || preferenceType == types.EvalDatetime {
		preferenceType = types.EvalInt64
	}

	if preferenceType == types.EvalBoolean {
		preferenceType = types.EvalDouble
	}

	// use decimal type if Float64() value loses precision
	useDecimal = useDecimal ||
		values.Int64(leftVal) > maxPrecisionInt ||
		values.Int64(rightVal) > maxPrecisionInt

	if useDecimal {
		leftDecimal := values.Decimal(leftVal)
		rightDecimal := values.Decimal(rightVal)
		switch op {
		case ADD:
			return values.NewSQLDecimal128(valueKind, leftDecimal.Add(rightDecimal)), nil
		case DIV:
			decimalResult := leftDecimal.Div(rightDecimal)
			// 4 comes from the div_precision_increment variable which
			// we do not allow to be set.
			scale := leftDecimal.Exponent() - 4
			decimalResult = decimalResult.Round(-scale)
			return values.NewSQLDecimal128(valueKind, decimalResult), nil
		case MULT:
			return values.NewSQLDecimal128(valueKind, leftDecimal.Mul(rightDecimal)), nil
		case SUB:
			return values.NewSQLDecimal128(valueKind, leftDecimal.Sub(rightDecimal)), nil
		default:
			return nil, fmt.Errorf("unrecognized arithmetic operator: %v", op)
		}
	}

	valueF := 0.0
	leftFloat := values.Float64(leftVal)
	rightFloat := values.Float64(rightVal)
	switch op {
	case ADD:
		valueF = leftFloat + rightFloat
	case DIV:
		floatResult := leftFloat / rightFloat
		return values.NewSQLFloat(valueKind, floatResult), nil
	case MULT:
		valueF = leftFloat * rightFloat
	case SUB:
		valueF = leftFloat - rightFloat
	default:
		return nil, fmt.Errorf("unrecognized arithmetic operator: %v", op)
	}

	switch preferenceType {
	case types.EvalInt64, types.EvalInt32:
		return values.NewSQLInt64(valueKind, int64(valueF)), nil
	case types.EvalDouble:
		return values.NewSQLFloat(valueKind, valueF), nil
	}
	return values.NewSQLFloat(valueKind, valueF), nil
}
