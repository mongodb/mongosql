package evaluator

import "github.com/10gen/sqlproxy/collation"

// CacheStage caches data returned by subqueries so we can avoid repeated pushdown
type CacheStage struct {
	source PlanStage // source is the operator that provides the data to project
	key    int       // the key is used to identify the subquery in the cacheRows slice
}

func NewCacheStage(key int, source PlanStage) *CacheStage {
	return &CacheStage{
		source: source,
		key:    key,
	}
}

func (cs *CacheStage) Open(ctx *ExecutionCtx) (Iter, error) {
	key := "cached_result{" + string(cs.key) + "}"
	if val, ok := ctx.CacheRows[key]; ok {
		if v, ok := val.(*CacheResult); ok {
			return &cachedResultsIter{
				cacheRows: v.cacheRows,
				closeErr:  v.closeErr,
				err:       v.err,
			}, nil
		}
	}

	sourceIter, err := cs.source.Open(ctx)
	if err != nil {
		return nil, err
	}

	result := &CacheResult{}
	ctx.CacheRows[key] = result
	return &CacheIter{
		source:      sourceIter,
		cacheResult: result,
	}, nil
}

func (cs *CacheStage) Columns() (columns []*Column) {
	return cs.source.Columns()
}

func (cs *CacheStage) Collation() *collation.Collation {
	return cs.source.Collation()
}

type CacheResult struct {
	cacheRows []*Row // a slice that holds all cached rows
	closeErr  error  // the error on close of the CacheIter
	err       error
}

type CacheIter struct {
	source      Iter         // the source iterator
	cacheResult *CacheResult // a struct that represents that cache in the CacheRows map
}

func (ci *CacheIter) Next(r *Row) bool {
	if ci.source.Next(r) {
		ci.cacheResult.cacheRows = append(ci.cacheResult.cacheRows, r)
		return true
	}
	return false
}

func (ci *CacheIter) Close() error {
	ci.cacheResult.closeErr = ci.source.Close()
	return ci.cacheResult.closeErr
}

func (ci *CacheIter) Err() error {
	ci.cacheResult.err = ci.source.Err()
	return ci.cacheResult.err
}

type cachedResultsIter struct {
	cacheRows []*Row // a slice that holds all cached rows
	index     int    // determines which data in a map element to return
	closeErr  error
	err       error
}

func (cr *cachedResultsIter) Next(r *Row) bool {
	if cr.index < len(cr.cacheRows) {
		r.Data = cr.cacheRows[cr.index].Data
		cr.index++
		return true
	}
	return false
}

func (cr *cachedResultsIter) Close() error {
	return cr.closeErr
}

func (cr *cachedResultsIter) Err() error {
	return cr.err
}
