package evaluator

type SQLNullValue struct{}

var SQLNull = SQLNullValue{}

func (nv SQLNullValue) Evaluate(ctx *EvalCtx) SQLValue {
	return nv
}

func (nv SQLNullValue) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	c := v.Evaluate(ctx)
	if _, ok := c.(SQLNullValue); ok {
		return 0, nil
	}
	return -1, nil
}

func (sn SQLNullValue) MongoValue() interface{} {
	return nil
}
