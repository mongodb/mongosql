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

type QueryPlan struct {
	Query       AlgebrizedQuery
	Children    []QueryPlan
	Type        QueryNode
	Composition CompositionPlan
}
