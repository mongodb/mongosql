package parser

// QueryStats collects some query metadata from a parse tree.
type QueryStats struct {
	Functions  map[string]int
	Joins      map[string]int
	Unions     map[string]int
	Subqueries map[string]int
}

// GetQueryStats returns QueryStats for the provided statement.
func GetQueryStats(stmt Statement) *QueryStats {
	stats := newQueryStats()
	_, err := walk(stats, stmt)
	if err != nil {
		panic(err)
	}
	return stats
}

func newQueryStats() *QueryStats {
	return &QueryStats{
		Functions:  make(map[string]int),
		Joins:      make(map[string]int),
		Unions:     make(map[string]int),
		Subqueries: make(map[string]int),
	}
}

func (s *QueryStats) PreVisit(current CST) (CST, error) {
	switch typed := current.(type) {
	case *FuncExpr:
		name := typed.Name
		s.Functions[name] += 1
	case *JoinTableExpr:
		kind := typed.Join
		s.Joins[kind] += 1
	case TableExprs:
		if len(typed) > 1 {
			kind := "comma"
			s.Joins[kind] += 1
		}
	case *Union:
		kind := typed.Type
		s.Unions[kind] += 1
	case *Subquery:
		kind := "expr"
		if typed.IsDerived {
			kind = "derived_table"
		}
		s.Subqueries[kind] += 1
	}
	return current, nil
}

func (*QueryStats) PostVisit(current CST) (CST, error) {
	return current, nil
}
