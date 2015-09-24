package evaluator

type SQLNumeric float64

func (sn SQLNumeric) Evaluate(_ *EvalCtx) SQLValue {
	return sn
}

func (sn SQLNumeric) MongoValue() interface{} {
	return float64(sn)
}

func (sn SQLNumeric) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	c := v.Evaluate(ctx)
	if n, ok := c.(SQLNumeric); ok {
		return int(float64(sn) - float64(n)), nil
	}
	// can only compare numbers to each other, otherwise treat as error
	return -1, ErrTypeMismatch
}
