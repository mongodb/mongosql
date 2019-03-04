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
	_, err := Walk(stats, stmt)
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

// PreVisit is called for every node before its children are walked.
func (s *QueryStats) PreVisit(current CST) (CST, error) {
	switch typed := current.(type) {
	case *FuncExpr:
		name := typed.Name
		s.Functions[name]++
	case *JoinTableExpr:
		kind := typed.Join
		s.Joins[kind]++
	case TableExprs:
		if len(typed) > 1 {
			kind := "comma"
			s.Joins[kind]++
		}
	case *Union:
		kind := typed.Type
		s.Unions[kind]++
	case *Subquery:
		kind := "expr"
		if typed.IsDerived {
			kind = "derived_table"
		}
		s.Subqueries[kind]++
	}
	return current, nil
}

// PostVisit is called for every node after its children are walked.
func (*QueryStats) PostVisit(current CST) (CST, error) {
	return current, nil
}
