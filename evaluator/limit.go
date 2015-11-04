package evaluator

// Limit restricts the number of rows returned by a query.
type Limit struct {
	// rowcount imposes a limit on the number of rows returned
	// by this operator
	rowcount float64

	// offset keeps a 0-indexed offset of where to commence
	// returning rows
	offset float64

	// total keeps track of how many rows have gone through
	// this operator.
	total float64

	// source is the source of data for this operator.
	source Operator
}

func (l *Limit) Open(ctx *ExecutionCtx) error {
	return l.source.Open(ctx)
}

func (l *Limit) Next(row *Row) bool {

	if l.offset != 0 {

		r := &Row{}

		for l.source.Next(r) {

			l.total += 1

			if l.total == l.offset {
				break
			}

		}

		if l.total < l.offset {
			return false
		}

		l.offset, l.total = 0, 0
	}

	hasNext := l.source.Next(row)

	l.total += 1

	if l.total > l.rowcount {
		return false
	}

	return hasNext
}

func (l *Limit) OpFields() (columns []*Column) {
	return l.source.OpFields()
}

func (l *Limit) Close() error {
	return l.source.Close()
}

func (l *Limit) Err() error {
	return l.source.Err()
}
