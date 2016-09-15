package evaluator

import "github.com/10gen/sqlproxy/collation"

// Limit restricts the number of rows returned by a query.
type LimitStage struct {
	// limit is the maximum number of rows to return
	limit int64

	// offset keeps a 0-indexed offset of where to start returning rows
	offset int64

	source PlanStage
}

func NewLimitStage(source PlanStage, offset int64, limit int64) *LimitStage {
	return &LimitStage{
		source: source,
		offset: offset,
		limit:  limit,
	}
}

type LimitIter struct {
	limit, offset, total int64

	source Iter
}

func (l *LimitStage) Open(ctx *ExecutionCtx) (Iter, error) {
	sourceIter, err := l.source.Open(ctx)
	if err != nil {
		return nil, err
	}
	return &LimitIter{
		limit:  l.limit,
		offset: l.offset,
		source: sourceIter,
	}, nil
}

func (l *LimitIter) Next(row *Row) bool {

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

	if l.total > l.limit {
		return false
	}

	return l.source.Next(row)
}

func (l *LimitStage) Columns() (columns []*Column) {
	return l.source.Columns()
}

func (l *LimitStage) Collation() *collation.Collation {
	return l.source.Collation()
}

func (l *LimitIter) Close() error {
	return l.source.Close()
}

func (l *LimitIter) Err() error {
	return l.source.Err()
}
