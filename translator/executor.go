package translator

import ()

type QueryNode byte

const (
	NodeFieldComp QueryNode = iota
	NodeLOJ
	NodeROJ
	NodeFOJ
	NodeCOJ
)

type CompositionPlan byte

const (
	LOJ CompositionPlan = iota
	ROJ
	FOJ
	COJ
	Union
	Subquery
)

// AlgebrizedQuery holds a name resolved form of a select query.
type AlgebrizedQuery struct {
	Collection interface{}
	Filter     interface{}
	Projection string
}

type QueryPlan struct {
	Query       AlgebrizedQuery
	Children    []QueryPlan
	Type        QueryNode
	Composition CompositionPlan
}
