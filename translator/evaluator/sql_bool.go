package evaluator

type SQLBool bool

func (sb SQLBool) Evaluate(ctx *EvalCtx) SQLValue {
	return sb
}

func (sb SQLBool) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	c := v.Evaluate(ctx)
	if n, ok := c.(SQLBool); ok {
		s1, s2 := bool(sb), bool(n)
		if s1 == s2 {
			return 0, nil
		} else if !s1 {
			return -1, nil
		}
		return 1, nil
	}
	// can only compare bool to a bool, otherwise treat as error
	return -1, ErrTypeMismatch
}

func (sb SQLBool) MongoValue() interface{} {
	return bool(sb)
}
