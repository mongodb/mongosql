package evaluator

type SQLString string

func (ss SQLString) Evaluate(_ *EvalCtx) SQLValue {
	return ss
}

func (ss SQLString) MongoValue() interface{} {
	return string(ss)
}

func (sn SQLString) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	c := v.Evaluate(ctx)
	if n, ok := c.(SQLString); ok {
		s1, s2 := string(sn), string(n)
		if s1 < s2 {
			return -1, nil
		} else if s1 > s2 {
			return 1, nil
		}
		return 0, nil
	}
	// can only compare numbers to each other, otherwise treat as error
	return -1, ErrTypeMismatch
}
