package evaluator

import (
	"container/heap"
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
)

const (
	// maxPathEvaluationCost is the maximum number of inner join path subtrees we will traverse
	// before terminating our search.
	maxPathEvaluationCost = 10e5
)

func optimizeInnerJoins(n Node, ctx *EvalCtx, logger log.Logger) (Node, error) {
	v := newInnerJoinOptimizer(ctx, logger)
	newN, err := v.visit(n)
	if err != nil {
		return nil, err
	}
	return newN, nil
}

func newInnerJoinOptimizer(ctx ConnectionCtx, logger log.Logger) *innerJoinOptimizer {
	return &innerJoinOptimizer{
		logger:           logger,
		ctx:              ctx,
		ancestorsAreLeft: true,
		sources:          make(map[string]joinLeafSource),
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
	sources map[string]joinLeafSource

	// predicateParts holds all predicates within the select scope of
	// the join subtree.
	predicateParts expressionParts

	sortablePaths *sortablePaths
	logger        log.Logger
	ctx           ConnectionCtx

	// nPlanStages holds the number of PlanStages contained within the
	// subtree visited by this optimizer
	nPlanStages int

	// hasSubquery is true if a subquery is within the set of data
	// sources referenced in this inner join subtree. We use this to
	// decide whether to compute self-join preference for candidates
	hasSubquery bool

	// optimizedSubtree holds the optimized version for this inner join
	// subtree
	optimizedSubtree Node

	// ancestorsAreLeft is used to track a traversal path that exclusively
	// branches left
	ancestorsAreLeft bool

	// pathEvaluationCost tracks how may subtree paths we've evaluated for this subtree.
	// If we hit the configured maximum cost, we terminate the search.
	pathEvaluationCost int64
}

// innerJoinExpr holds every SQLExpr used in an ON clause and
// isEquality is true if the expression could be used in a lookup
// stage in MongoDB.
type innerJoinExpr struct {
	expr       SQLExpr
	isEquality bool
}

// tableEdge holds the tables referenced within an innerJoinExpr that
// is an equality predicate between two fields e.g. db1.foo.a = db2.bar.b
// would create a table edge with entries "db1.foo, db2.bar".
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
// such that the first table in the edge is the unself-joined table.
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

func (v *innerJoinOptimizer) visit(n Node) (Node, error) {
	var err error

	if _, ok := n.(PlanStage); ok {
		v.nPlanStages++
	}

	switch typedN := n.(type) {
	case *DynamicSourceStage:
		ms := joinLeafSource{
			dataSource: typedN,
		}
		v.sources[fullyQualifiedTableName(typedN.dbName, typedN.aliasName)] = ms
		return n, nil

	case *FilterStage:
		var predicateParts expressionParts
		predicateParts, err = splitExpressionIntoParts(typedN.matcher)
		if err != nil {
			return nil, err
		}

		v.predicateParts = append(v.predicateParts, predicateParts...)

		var source Node
		source, err = v.visit(typedN.source)
		if err != nil {
			return nil, err
		}

		if source != typedN.source {
			if len(predicateParts) > 0 {
				n = NewFilterStage(source.(PlanStage), predicateParts.combine())
			} else {
				n = source
			}
		}
		return n, nil

	case *JoinStage:
		if !canOptimizeJoinSubtree(typedN.left) || !canOptimizeJoinSubtree(typedN.right) {
			return n, nil
		}

		independentlyOptimizeChildren := true
		kind, matcher := typedN.kind, typedN.matcher

		if kind == InnerJoin || kind == StraightJoin {
			exprs := v.getInnerJoinExprs(typedN.matcher)
			if len(exprs) != 0 {
				v.exprs = append(v.exprs, exprs...)
				independentlyOptimizeChildren = false

				oldAncestorsAreLeft := v.ancestorsAreLeft
				v.ancestorsAreLeft = false

				var newR Node
				newR, err = v.visit(typedN.right)
				if err != nil {
					return nil, err
				}

				v.ancestorsAreLeft = oldAncestorsAreLeft

				var newL Node
				newL, err = v.visit(typedN.left)
				if err != nil {
					return nil, err
				}

				if typedN.left != newL.(PlanStage) || typedN.right != newR.(PlanStage) {
					n = NewJoinStage(kind, newL.(PlanStage), newR.(PlanStage), matcher)
				}
			}
		}

		if independentlyOptimizeChildren {
			v.logger.Debugf(log.Dev, "attempting inner join optimization on %v subtree", kind)

			newRightOptimizer := newInnerJoinOptimizer(v.ctx, v.logger)
			newRightOptimizer.predicateParts = v.predicateParts
			var newR Node
			newR, err = newRightOptimizer.visit(typedN.right)
			if err != nil {
				return nil, err
			}

			newLeftOptimizer := newInnerJoinOptimizer(v.ctx, v.logger)
			newLeftOptimizer.predicateParts = v.predicateParts
			var newL Node
			newL, err = newLeftOptimizer.visit(typedN.left)
			if err != nil {
				return nil, err
			}

			if typedN.left != newL.(PlanStage) || typedN.right != newR.(PlanStage) {
				if typedN.kind == InnerJoin || typedN.kind == StraightJoin {
					if newRightOptimizer.nPlanStages > newLeftOptimizer.nPlanStages {
						newL, newR = newR, newL
					}
				}
				n = NewJoinStage(kind, newL.(PlanStage), newR.(PlanStage), matcher)
			}
			return n, nil
		}

		if v.ancestorsAreLeft && v.atReorderingLeaf(typedN.left) {
			v.optimizedSubtree, err = v.reorderInnerJoins()
		}

		if v.optimizedSubtree != nil {
			return v.optimizedSubtree, err
		}
		return n, err

	case *MongoSourceStage:
		ms := joinLeafSource{
			nPipelineStages: len(typedN.pipeline),
			dataSource:      typedN,
		}
		v.sources[fullyQualifiedTableName(typedN.dbName, typedN.aliasNames[0])] = ms
		return n, nil

	case *SQLSubqueryExpr:
		v.logger.Debugf(log.Dev, "attempting to optimize inner "+
			"join in subquery expression:\n '%v'", typedN.String())

		subqueryOptimizer := newInnerJoinOptimizer(v.ctx, v.logger)

		plan, err := subqueryOptimizer.visit(typedN.plan)
		if err != nil {
			return nil, err
		}

		if plan != typedN.plan {
			n = &SQLSubqueryExpr{
				correlated: typedN.correlated,
				plan:       plan.(PlanStage),
				allowRows:  typedN.allowRows,
			}
		}
		return n, nil

	case *SubquerySourceStage:
		v.logger.Debugf(log.Dev, "attempting to optimize inner "+
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
				typedN.dbName,
				typedN.aliasName,
			)
		}

		ms := joinLeafSource{nPipelineStages, n}
		v.nPlanStages += subqueryOptimizer.nPlanStages
		dbNames := generateDbSetFromColumns(typedN.Columns())
		for dbName := range dbNames {
			v.sources[fullyQualifiedTableName(dbName, typedN.aliasName)] = ms
		}
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

// atReorderingLeaf returns true if this (left) Node is
// at a level in the query plan where it can be reordered.
func (v *innerJoinOptimizer) atReorderingLeaf(n Node) bool {
	switch n.(type) {
	case *MongoSourceStage, *SubquerySourceStage:
		return true
	}
	return false
}

// evaluateTreeCandidates builds all possible inner join and
// retains the most optimal subtree it finds .
func (v *innerJoinOptimizer) evaluateTreeCandidates(existingPath path, edges []tableEdge) {
	if v.pathEvaluationCost >= maxPathEvaluationCost {
		return
	}

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
		newPath := make([]tableEdge, len(existingPath)+1)
		copy(newPath, existingPath)
		newPath[len(existingPath)] = edge
		v.pathEvaluationCost++
		v.evaluateTreeCandidates(newPath, remainingEdges)
	}
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
		left := equalityExpr.left.(SQLColumnExpr)
		right := equalityExpr.right.(SQLColumnExpr)
		edge := tableEdge{
			[]string{
				fullyQualifiedTableName(left.databaseName, left.tableName),
				fullyQualifiedTableName(right.databaseName, right.tableName),
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
	exprs := []innerJoinExpr{}

	for _, expr := range conjunctiveExprs {
		hasEquality := false
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
func (v *innerJoinOptimizer) reconstructSubtree(p path) (Node, error) {

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
				leftColumn := eq.left.(SQLColumnExpr)
				rightColumn := eq.right.(SQLColumnExpr)
				left := fullyQualifiedTableName(leftColumn.databaseName, leftColumn.tableName)
				right := fullyQualifiedTableName(rightColumn.databaseName, rightColumn.tableName)
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
		qualifiedTableNames, err := referencedTables(e.expr)
		if err != nil {
			return nil, err
		}
		criterion := freeCriterion{e.expr, qualifiedTableNames}
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

		var selfJoinableCriteria []SQLExpr

		// find any criteria that could be moved to the current join
		// stage level
		newFreeCriteria := []freeCriterion{}
		for _, criterion := range freeCriteria {
			canSelfJoin := true
			for _, table := range criterion.tables {
				if _, ok := subtreeTables[table]; !ok {
					canSelfJoin = false
					break
				}
			}
			if canSelfJoin {
				selfJoinableCriteria = append(selfJoinableCriteria, criterion.expr)
			} else {
				newFreeCriteria = append(newFreeCriteria, criterion)
			}
		}

		freeCriteria = newFreeCriteria

		// move mergable criteria further down the subtree to prune
		// result set sooner
		for _, criterion := range selfJoinableCriteria {
			boundCriterion = &SQLAndExpr{boundCriterion, criterion}
		}

		unselfJoinedSource := v.sources[expr.tables[0]].dataSource.(PlanStage)
		if newN == nil {
			right := v.sources[expr.tables[1]].dataSource.(PlanStage)
			newN, unselfJoinedSource = unselfJoinedSource, right
		}
		newN = NewJoinStage(InnerJoin, newN, unselfJoinedSource, boundCriterion)
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

func (v *innerJoinOptimizer) reorderInnerJoins() (Node, error) {
	equalities := v.getInnerJoinEqualities()
	allCriteria := []SQLExpr{}

	for _, e := range v.exprs {
		allCriteria = append(allCriteria, e.expr)
	}

	cardinalityAlteringPredicates := make(map[string]expressionPart)

	for _, part := range v.predicateParts {
		if tableRefCount := len(part.qualifiedTableNames); tableRefCount != 1 {
			continue
		}
		cardinalityAlteringPredicates[part.qualifiedTableNames[0]] = part
	}

	v.sortablePaths = &sortablePaths{
		cardinalityAlteringPredicates: cardinalityAlteringPredicates,
		cachedPathSelfJoinPotential:   make(map[string]int),
		cachedEdgeSelfJoinPotential:   make(map[string]bool),
		logger:    v.logger,
		matcher:   combineExpressions(allCriteria),
		optimizer: v,
	}

	heap.Init(v.sortablePaths)

	// Prefer the original tree structure, until a better structure is found.
	for i, j := 0, len(equalities)-1; i < j; i, j = i+1, j-1 {
		equalities[i], equalities[j] = equalities[j], equalities[i]
	}

	v.evaluateTreeCandidates(path{}, equalities)
	if v.pathEvaluationCost >= maxPathEvaluationCost {
		v.logger.Debugf(log.Dev, "terminated inner join optimization search (cost at %v)",
			v.pathEvaluationCost)
	}

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
	logger    log.Logger

	// cardinalityAlteringPredicates contains predicates that might alter cardinality.
	cardinalityAlteringPredicates map[string]expressionPart

	// cachedPathSelfJoinPotential holds the self-join potential for candidate paths.
	cachedPathSelfJoinPotential map[string]int

	// cachedEdgeSelfJoinPotential holds the self-join potential for candidate edges.
	cachedEdgeSelfJoinPotential map[string]bool
}

func (s sortablePaths) Len() int {
	return len(s.paths)
}

func (s sortablePaths) Less(i, j int) bool {
	leftEdges, rightEdges := s.paths[i], s.paths[j]

	// if any of the sources is a subquery, we do not need to give
	// preference to candidates with higher self-join potentials since
	// we'll eventually be blocked on a lookup with the subquery
	if !s.optimizer.hasSubquery {
		// TODO: use faster keying method?
		leftName, rightName := leftEdges.String(), rightEdges.String()

		leftMergePotential, ok := s.cachedPathSelfJoinPotential[leftName]
		if !ok {
			leftMergePotential = s.pathSelfJoinPotential(leftEdges)
			s.cachedPathSelfJoinPotential[leftName] = leftMergePotential
		}

		rightMergePotential, ok := s.cachedPathSelfJoinPotential[rightName]
		if !ok {
			rightMergePotential = s.pathSelfJoinPotential(rightEdges)
			s.cachedPathSelfJoinPotential[rightName] = rightMergePotential
		}

		if leftMergePotential != rightMergePotential {
			return leftMergePotential < rightMergePotential
		}
	}

	for idx := range leftEdges {

		left := s.optimizer.sources[leftEdges[idx].tables[0]]
		right := s.optimizer.sources[rightEdges[idx].tables[0]]

		// push dynamic sources to rightmost Node of tree
		if _, ok := left.dataSource.(*DynamicSourceStage); ok {
			return true
		}

		if _, ok := right.dataSource.(*DynamicSourceStage); ok {
			return false
		}

		i := compareInts(left.nPipelineStages, right.nPipelineStages)

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

		// candidate paths with earlier cardinality altering predicates are preferable.
		if _, ok := s.cardinalityAlteringPredicates[leftEdges[idx].tables[0]]; ok {
			return false
		}

		if _, ok := s.cardinalityAlteringPredicates[rightEdges[idx].tables[0]]; ok {
			return true
		}
	}

	return false
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

// pathSelfJoinPotential returns an integer proportional to how many edges in the path could
// be used in a self-joined. The higher the the number, the higher the self-join potential.
func (s sortablePaths) pathSelfJoinPotential(path path) int {

	i := 0

	for _, edge := range path {
		edgeKey := fmt.Sprintf("%v-%v", edge.tables[0], edge.tables[1])
		canSelfJoinEdge, ok := s.cachedEdgeSelfJoinPotential[edgeKey]
		if ok {
			if canSelfJoinEdge {
				i++
				continue
			}
			break
		}

		left := s.optimizer.sources[edge.tables[0]]
		leftSource, ok := left.dataSource.(*MongoSourceStage)
		if !ok {
			break
		}

		right := s.optimizer.sources[edge.tables[1]]
		rightSource, ok := right.dataSource.(*MongoSourceStage)
		if !ok {
			break
		}

		var v *pushDownOptimizer
		canSelfJoinEdge = v.canSelfJoinTables(s.logger, leftSource, rightSource,
			s.matcher, InnerJoin)
		s.cachedEdgeSelfJoinPotential[edgeKey] = canSelfJoinEdge
		reversedEdgeKey := fmt.Sprintf("%v-%v", edge.tables[1], edge.tables[0])
		s.cachedEdgeSelfJoinPotential[reversedEdgeKey] = canSelfJoinEdge
		if !canSelfJoinEdge {
			break
		}
	}

	return i
}
