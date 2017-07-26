package evaluator

import (
	"container/heap"
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
)

func optimizeInnerJoins(n node, ctx *EvalCtx, logger *log.Logger) (node, error) {
	v := newInnerJoinOptimizer(ctx, logger)
	newN, err := v.visit(n)
	if err != nil {
		return nil, err
	}
	return newN, nil
}

func newInnerJoinOptimizer(ctx ConnectionCtx, logger *log.Logger) *innerJoinOptimizer {
	return &innerJoinOptimizer{
		logger:  logger,
		ctx:     ctx,
		sources: make(map[string]innerJoinSource),
	}
}

// innerJoinOptimizer holds information that's used to restructure a
// subtree of inner joins to optimize pushdown
type innerJoinOptimizer struct {
	// exprs is a slice of all SQLExprs conjunctive expressions used
	// across the ON clauses in an inner join subtree
	exprs []innerJoinExpr
	// sources is a map of alias name to a PlanStage appearing in an
	// inner join subtree
	sources map[string]innerJoinSource

	sortablePaths *sortablePaths
	logger        *log.Logger
	ctx           ConnectionCtx
	// nPlanStages holds the number of PlanStages contained within the
	// subtree visited by this optimizer
	nPlanStages int
	// hasSubquery is true if a subquery is within the set of data
	// sources referenced in this inner join subtree. We use this to
	// decide whether to compute merge preference for candidates
	hasSubquery bool
	// optimizedSubtree holds the optimized version for this inner join
	// subtree
	optimizedSubtree node
}

// innerJoinTerms holds every SQLExpr used in an ON clause and
// isEquality is true if the expression could be used in a lookup
// stage in MongoDB.
type innerJoinExpr struct {
	expr       SQLExpr
	isEquality bool
}

// innerJoinSource holds all data sources for an inner join subtree.
type innerJoinSource struct {
	// nPipelineStages holds the number of pipeline stages contained
	// within a data source. For subqueries, it adds the number of
	// PlanStages contained within subquery.
	nPipelineStages int
	dataSource      node
}

// tableEdge holds the tables referenced within an innerJoinExpr that
// is an equality predicate between two fields e.g. foo.a = bar.b
// would create a table edge with entries "foo, bar".
type tableEdge struct {
	// tables holds the tables referenced by the equality expression and
	// always has exactly two entries - the left and right data sources.
	tables []string
}

// contains returns true if table is referenced
// on the edge, e.
func (e *tableEdge) contains(table string) bool {
	return util.SliceContains(e.tables, table)
}

// path is a slice of edges.
type path []tableEdge

// contains returns true if edge
// references the same set of tables
// as an already existing edge in p.
func (p *path) contains(edge tableEdge) bool {

	for _, e := range *p {
		if (edge.tables[0] == e.tables[0] &&
			edge.tables[1] == e.tables[1]) ||
			(edge.tables[1] == e.tables[0] &&
				edge.tables[0] == e.tables[1]) {
			return true
		}
	}

	return false
}

// canIncludeEdge returns true if the edge is valid to be
// included to the existing path, and false otherwise.
func (p *path) canIncludeEdge(edge tableEdge) bool {
	for _, e := range *p {
		if util.StringSliceContains(e.tables, edge.tables[0]) ||
			util.StringSliceContains(e.tables, edge.tables[1]) {
			return true
		}
	}

	return len(*p) == 0
}

// satisfiesDependency returns true if the edge can be used to
// lookup a new data source on the existing path, p.
func (p *path) satisfiesDependency(e tableEdge) bool {
	for _, edge := range *p {
		if (edge.tables[0] == e.tables[0] ||
			edge.tables[1] == e.tables[0]) &&
			(edge.tables[0] == e.tables[1] ||
				edge.tables[1] == e.tables[1]) {
			return false
		}
	}

	return true
}

// orderCandidateEdge returns a modified version of the edge
// such that the first table in the edge is the unmerged table.
// If this is already the case, it returns expr unmodified.
func (p *path) orderCandidateEdge(edge tableEdge) tableEdge {
	// if the tables are the same, no reordering needed
	if edge.tables[0] != edge.tables[1] {
		for _, e := range *p {
			// if the first table has already been
			// absorbed, flip the order
			if e.tables[0] == edge.tables[0] ||
				e.tables[1] == edge.tables[0] {
				tables := []string{
					edge.tables[1],
					edge.tables[0],
				}
				return tableEdge{tables}
			}
		}
	}

	return edge
}

func (p *path) String() (s string) {
	for _, e := range *p {
		s += strings.Join(e.tables, "-")
	}
	return s
}

func (v *innerJoinOptimizer) visit(n node) (node, error) {
	var err error

	if _, ok := n.(PlanStage); ok {
		v.nPlanStages += 1
	}

	switch typedN := n.(type) {
	case *DynamicSourceStage:
		ijs := innerJoinSource{
			dataSource: typedN,
		}
		v.sources[typedN.aliasName] = ijs
		return n, nil
	case *JoinStage:
		if !v.canOptimizeInnerJoinSubtree(typedN.left) ||
			!v.canOptimizeInnerJoinSubtree(typedN.right) {
			return n, nil
		}

		independentlyOptimizeChildren := true

		if typedN.kind == InnerJoin || typedN.kind == StraightJoin {
			exprs := v.getInnerJoinExprs(typedN.matcher)
			if len(exprs) != 0 {

				v.exprs = append(v.exprs, exprs...)

				independentlyOptimizeChildren = false

				newRight, err := v.visit(typedN.right)
				if err != nil {
					return nil, err
				}

				newLeft, err := v.visit(typedN.left)
				if err != nil {
					return nil, err
				}

				if typedN.left != newLeft.(PlanStage) || typedN.right != newRight.(PlanStage) {
					n = NewJoinStage(
						typedN.kind,
						newLeft.(PlanStage),
						newRight.(PlanStage),
						typedN.matcher,
					)
				}

			}
		}

		if independentlyOptimizeChildren {
			v.logger.Logf(log.DebugHigh, "attempting inner join "+
				"optimization on %v subtree", typedN.kind)

			newR := newInnerJoinOptimizer(v.ctx, v.logger)
			newRight, err := newR.visit(typedN.right)
			if err != nil {
				return nil, err
			}

			newL := newInnerJoinOptimizer(v.ctx, v.logger)
			newLeft, err := newL.visit(typedN.left)
			if err != nil {
				return nil, err
			}

			if typedN.left != newLeft.(PlanStage) || typedN.right != newRight.(PlanStage) {

				if typedN.kind == InnerJoin || typedN.kind == StraightJoin {
					if newR.nPlanStages > newL.nPlanStages {
						newLeft, newRight = newRight, newLeft
					}
				}

				n = NewJoinStage(
					typedN.kind,
					newLeft.(PlanStage),
					newRight.(PlanStage),
					typedN.matcher,
				)
			}

			return n, nil
		}

		if v.atReorderingLeaf(typedN.left) {
			v.optimizedSubtree, err = v.reorderInnerJoins()
		}

		if v.optimizedSubtree != nil {
			return v.optimizedSubtree, err
		}

		return n, err
	case *MongoSourceStage:
		ijs := innerJoinSource{
			nPipelineStages: len(typedN.pipeline),
			dataSource:      typedN,
		}
		v.sources[typedN.aliasNames[0]] = ijs
		return n, nil
	case *SQLSubqueryExpr:
		v.logger.Logf(log.DebugHigh, "attempting to optimize inner "+
			"join in subquery expression: '%v'", typedN.String())

		subqueryOptimizer := newInnerJoinOptimizer(v.ctx, v.logger)

		plan, err := subqueryOptimizer.visit(typedN.plan)
		if err != nil {
			return nil, err
		}

		if plan != typedN.plan {
			n = &SQLSubqueryExpr{
				correlated: typedN.correlated,
				plan:       plan.(PlanStage),
			}
		}

		return n, nil
	case *SubquerySourceStage:
		v.logger.Logf(log.DebugHigh, "attempting to optimize inner "+
			"join in subquery '%v'", typedN.aliasName)

		subqueryOptimizer := newInnerJoinOptimizer(v.ctx, v.logger)
		plan, err := subqueryOptimizer.visit(typedN.source)
		if err != nil {
			return nil, err
		}

		// We compute the cost of a Subquery as the total
		// of the number of sources contained within it,
		// plus the cost associated with each of its sources
		nPipelineStages := subqueryOptimizer.nPlanStages
		for _, source := range subqueryOptimizer.sources {
			nPipelineStages += source.nPipelineStages
		}

		if typedN.source != plan.(PlanStage) {
			n = NewSubquerySourceStage(
				plan.(PlanStage),
				typedN.selectID,
				typedN.aliasName,
			)
		}

		ijs := innerJoinSource{nPipelineStages, n}
		v.nPlanStages += subqueryOptimizer.nPlanStages
		v.sources[typedN.aliasName] = ijs
		v.hasSubquery = true
		return n, nil
	case *UnionStage:
		newV := newInnerJoinOptimizer(v.ctx, v.logger)
		newRight, err := newV.visit(typedN.right)
		if err != nil {
			return nil, err
		}

		newV = newInnerJoinOptimizer(v.ctx, v.logger)
		newLeft, err := newV.visit(typedN.left)
		if err != nil {
			return nil, err
		}

		if typedN.left != newLeft.(PlanStage) || typedN.right != newRight.(PlanStage) {
			n = NewUnionStage(
				typedN.kind,
				newLeft.(PlanStage),
				newRight.(PlanStage),
			)
		}

		return n, nil
	}

	return walk(v, n)
}

// atReorderingLeaf returns true if this (left) node is
// at a level in the query plan where it can be reordered.
func (v *innerJoinOptimizer) atReorderingLeaf(n node) bool {
	switch n.(type) {
	case *MongoSourceStage, *SubquerySourceStage:
		return true
	}
	return false
}

// evaluateTreeCandidates builds all possible inner join and
// retains the most optimal subtree it finds .
func (v *innerJoinOptimizer) evaluateTreeCandidates(existingPath path, edges []tableEdge) {

	switch len(existingPath) {
	case 0:
	case 1:
		// ensure the left source has the highest cost
		t0 := existingPath[0].tables[0]
		t1 := existingPath[0].tables[1]
		leftCost := v.sources[t0].nPipelineStages
		rightCost := v.sources[t1].nPipelineStages
		if rightCost > leftCost {
			existingPath[0] = tableEdge{[]string{t1, t0}}
		}
		if len(existingPath) == len(v.sources)-1 {
			if len(v.sortablePaths.paths) > 1 {
				heap.Pop(v.sortablePaths)
			}
			heap.Push(v.sortablePaths, existingPath)
			return
		}
	default:
		// order candidate edge as needed
		edgeIdx := len(existingPath) - 1
		newPath := existingPath[:edgeIdx]
		edge := existingPath[edgeIdx]

		if !newPath.satisfiesDependency(edge) {
			return
		}

		existingPath[edgeIdx] = newPath.orderCandidateEdge(edge)

		if len(existingPath) == len(v.sources)-1 {
			if len(v.sortablePaths.paths) > 1 {
				heap.Pop(v.sortablePaths)
			}
			heap.Push(v.sortablePaths, existingPath)
			return
		}

		// This helps us cut down the search space - O(n!) - by ignoring
		// candidate paths that are incompatible with the join criteria
		// à la beam search.
		if !newPath.canIncludeEdge(edge) {
			return
		}
	}

	for i, edge := range edges {
		remainingEdges := make(path, len(edges)-1)
		copy(remainingEdges[:i], edges[:i])
		copy(remainingEdges[i:], edges[i+1:])
		v.evaluateTreeCandidates(append(existingPath, edge), remainingEdges)
	}
}

// canOptimizeInnerJoinSubtree returns true if this subtree
// can be optimized.
func (v *innerJoinOptimizer) canOptimizeInnerJoinSubtree(n node) bool {
	switch n.(type) {
	case *DynamicSourceStage, *MongoSourceStage, *JoinStage, *SubquerySourceStage:
		return true
	}
	return false
}

// getInnerJoinEqualities returns all table edges used in the
// inner join subtree that could potentially be used for
// lookup stage operations in MongoDB.
func (v *innerJoinOptimizer) getInnerJoinEqualities() []tableEdge {
	edges := path{}

	for _, e := range v.exprs {
		// don't add redundant equality edges to slice
		if !e.isEquality {
			continue
		}

		equalityExpr := e.expr.(*SQLEqualsExpr)
		edge := tableEdge{
			[]string{
				equalityExpr.left.(SQLColumnExpr).tableName,
				equalityExpr.right.(SQLColumnExpr).tableName,
			},
		}

		if !edges.contains(edge) {
			edges = append(edges, edge)
		}
	}

	return edges
}

// getInnerJoinExprs takes a SQLExpr and returns all conjunctive terms
// along with a boolean that's true if the term can be used for a lookup
// operation.
func (v *innerJoinOptimizer) getInnerJoinExprs(e SQLExpr) []innerJoinExpr {
	conjunctiveExprs := splitExpression(e)
	hasEquality := false
	exprs := []innerJoinExpr{}

	for _, expr := range conjunctiveExprs {
		hasEquality = false
		if equalExpr, ok := expr.(*SQLEqualsExpr); ok {
			if _, ok := equalExpr.left.(SQLColumnExpr); ok {
				if _, ok = equalExpr.right.(SQLColumnExpr); ok {
					hasEquality = true
				}
			}
		}

		exprs = append(exprs, innerJoinExpr{expr, hasEquality})
	}

	return exprs
}

// reconstructSubtree takes a subtree configuration and returns a
// reconstructed subtree using that configuration.
func (v *innerJoinOptimizer) reconstructSubtree(p path) (node, error) {

	type freeCriterion struct {
		expr   SQLExpr
		tables []string
	}

	boundCriteria := []SQLExpr{}
	freeCriteria := []freeCriterion{}

	for _, edge := range p {
		bound, idx := false, 0
		for j, e := range v.exprs {
			if e.isEquality {
				eq := e.expr.(*SQLEqualsExpr)
				left := eq.left.(SQLColumnExpr).tableName
				right := eq.right.(SQLColumnExpr).tableName
				if edge.contains(left) && edge.contains(right) {
					boundCriteria = append(boundCriteria, eq)
					bound, idx = true, j
					break
				}
			}
		}
		if bound {
			v.exprs = append(v.exprs[:idx], v.exprs[idx+1:]...)
		}
	}

	for _, e := range v.exprs {
		tables, err := referencedTables(e.expr)
		if err != nil {
			return nil, err
		}
		criterion := freeCriterion{e.expr, tables}
		freeCriteria = append(freeCriteria, criterion)
	}

	var newN PlanStage
	subtreeTables := map[string]struct{}{}

	for idx, expr := range p {
		boundCriterion := boundCriteria[idx]
		criterionTables, err := referencedTables(boundCriterion)
		if err != nil {
			return nil, err
		}

		for _, table := range criterionTables {
			subtreeTables[table] = struct{}{}
		}

		var mergeableCriteria []SQLExpr

		// find any criteria that could be moved to the current join
		// stage level
		newFreeCriteria := []freeCriterion{}
		for _, criterion := range freeCriteria {
			canMerge := true
			for _, table := range criterion.tables {
				if _, ok := subtreeTables[table]; !ok {
					canMerge = false
					break
				}
			}
			if canMerge {
				mergeableCriteria = append(mergeableCriteria, criterion.expr)
			} else {
				newFreeCriteria = append(newFreeCriteria, criterion)
			}
		}

		freeCriteria = newFreeCriteria

		// move mergable criteria further down the subtree to prune
		// result set sooner
		for _, criterion := range mergeableCriteria {
			boundCriterion = &SQLAndExpr{boundCriterion, criterion}
		}

		unmergedSource := v.sources[expr.tables[0]].dataSource.(PlanStage)
		if newN == nil {
			right := v.sources[expr.tables[1]].dataSource.(PlanStage)
			newN, unmergedSource = unmergedSource, right
		}
		newN = NewJoinStage(InnerJoin, newN, unmergedSource, boundCriterion)
	}

	if lenFreeCriteria := len(freeCriteria); lenFreeCriteria != 0 {
		msg := fmt.Sprintf("found %v unbound %v after building: %v",
			lenFreeCriteria,
			util.Pluralize(lenFreeCriteria, "criterion", "criteria"),
			freeCriteria,
		)
		panic(msg)
	}

	return newN, nil
}

func (v *innerJoinOptimizer) reorderInnerJoins() (node, error) {
	equalities := v.getInnerJoinEqualities()

	allCriteria := []SQLExpr{}

	for _, e := range v.exprs {
		allCriteria = append(allCriteria, e.expr)
	}

	v.sortablePaths = &sortablePaths{
		optimizer:       v,
		logger:          log.NewLogger(nil),
		matcher:         combineExpressions(allCriteria),
		mergePotentials: make(map[string]int),
	}

	heap.Init(v.sortablePaths)

	v.evaluateTreeCandidates(path{}, equalities)

	optimalPath := v.sortablePaths.Pop()
	if optimalPath == nil {
		return nil, nil
	}

	return v.reconstructSubtree(optimalPath.(path))
}

// sortablePaths sort helpers.
type sortablePaths struct {
	optimizer *innerJoinOptimizer
	paths     []path
	matcher   SQLExpr
	logger    *log.Logger

	// mergePotentials holds the merge potential for candidate paths
	mergePotentials map[string]int
}

func (s sortablePaths) Len() int {
	return len(s.paths)
}

func (s sortablePaths) Less(i, j int) bool {
	leftEdges, rightEdges := s.paths[i], s.paths[j]

	// if any of the sources is a subquery, we do not need to give
	// preference to candidates with higher merge potentials since
	// we'll eventually be blocked on a lookup with the subquery
	if !s.optimizer.hasSubquery {
		// TODO: use faster keying method?
		leftName, rightName := leftEdges.String(), rightEdges.String()

		leftMergePotential, ok := s.mergePotentials[leftName]
		if !ok {
			leftMergePotential = s.mergeTablesPotential(leftEdges)
			s.mergePotentials[leftName] = leftMergePotential
		}

		rightMergePotential, ok := s.mergePotentials[rightName]
		if !ok {
			rightMergePotential = s.mergeTablesPotential(rightEdges)
			s.mergePotentials[rightName] = rightMergePotential
		}

		if leftMergePotential != rightMergePotential {
			return leftMergePotential < rightMergePotential
		}

	}

	for idx, _ := range leftEdges {

		left := s.optimizer.sources[leftEdges[idx].tables[0]]
		right := s.optimizer.sources[rightEdges[idx].tables[0]]

		// push dynamic sources to rightmost node of tree
		if _, ok := left.dataSource.(*DynamicSourceStage); ok {
			return true
		}

		if _, ok := right.dataSource.(*DynamicSourceStage); ok {
			return false
		}

		i, _ := compareInts(left.nPipelineStages, right.nPipelineStages)

		switch idx {
		// determine cost associated with the first path edge
		case 0:
			// we want the leftmost table in the subtree to have the
			// largest number of PlanStages
			if i != 0 {
				return i != 1
			}
		default:
			if i != 0 {
				return i == 1
			}
		}
	}

	return true
}

func (s *sortablePaths) Pop() interface{} {
	if len(s.paths) == 0 {
		return nil
	}
	p := s.paths[len(s.paths)-1]
	s.paths = s.paths[:len(s.paths)-1]
	return p
}

func (s *sortablePaths) Push(p interface{}) {
	s.paths = append(s.paths, p.(path))
}

func (s sortablePaths) Swap(i, j int) {
	s.paths[i], s.paths[j] = s.paths[j], s.paths[i]
}

// mergeTablesPotential returns an int indicative how proportional
// to how many edges in the path can be merged. The higher the
// the number, the higher the merge potential.
func (s sortablePaths) mergeTablesPotential(path path) int {

	i := 0

	for _, edge := range path {

		left := s.optimizer.sources[edge.tables[0]]
		right := s.optimizer.sources[edge.tables[1]]

		leftSource, ok := left.dataSource.(*MongoSourceStage)
		if !ok {
			break
		}

		rightSource, ok := right.dataSource.(*MongoSourceStage)
		if !ok {
			break
		}

		if canMergeTables(s.logger, leftSource, rightSource, s.matcher) {
			i++
		} else {
			break
		}

	}

	return i
}
