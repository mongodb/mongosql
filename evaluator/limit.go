package evaluator

import (
	"gopkg.in/mgo.v2/bson"
)

// Limit restricts the number of rows returned by a query.
type Limit struct {
	// rowcount imposes a limit on the number of rows returned
	// by this operator
	rowcount int64

	// offset keeps a 0-indexed offset of where to commence
	// returning rows
	offset int64

	// total keeps track of how many rows have gone through
	// this operator.
	total int64

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

	l.total += 1

	if l.total > l.rowcount {
		return false
	}

	return l.source.Next(row)
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

///////////////
//Optimization
///////////////

func (_ *optimizer) visitLimit(limit *Limit) (Operator, error) {

	sa, ts, ok := canPushDown(limit.source)
	if !ok {
		return limit, nil
	}

	pipeline := ts.pipeline

	if limit.offset > 0 {
		pipeline = append(pipeline, bson.D{{"$skip", limit.offset}})
	}

	if limit.rowcount > 0 {
		pipeline = append(pipeline, bson.D{{"$limit", limit.rowcount}})
	}

	ts = ts.WithPipeline(pipeline)
	sa = sa.WithSource(ts)

	return sa, nil
}
