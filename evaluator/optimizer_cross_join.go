package evaluator

import (
	"fmt"
	"sort"

	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/variable"
)

func optimizeCrossJoins(n Node, ctx *EvalCtx, logger log.Logger) (Node, error) {
	optimizeCrossJoins := ctx.Variables().GetBool(variable.OptimizeCrossJoins)

	if !optimizeCrossJoins {
		logger.Warnf(log.Admin, "optimize_cross_joins is false: skipping cross join optimizer")
		return n, nil
	}

	n, err := newCrossJoinOptimizer(logger).visit(n)
	if err != nil {
		return nil, err
	}

	return n, nil
}

func newCrossJoinOptimizer(logger log.Logger) *crossJoinOptimizer {
	return &crossJoinOptimizer{
		sources:             make(map[string]joinLeafSource),
		qualifiedTableNames: make(map[string]struct{}),
		logger:              logger,
	}
}

type crossJoinOptimizer struct {
	// predicateParts holds conjunctive terms from a filter expression on the cross join subtree.
	// It is updated as the subtree is traversed.
	predicateParts expressionParts
	// fullPredicateParts is the same as predicateParts above, except it includes predicates
	// gotten from the subtree's join criteria.
	fullPredicateParts expressionParts
	// isChildJoinNode is true if the optimizer visitor is at a node rooting a cross join subtree
	// and false otherwise.
	isChildJoinNode bool
	// logger is a logger.
	logger log.Logger
	// sources is a map of fully qualified table names to leaf sources.
	sources map[string]joinLeafSource
	// qualifiedTableNames is used to track which table names have been seen during this subtree
	// traversal.
	qualifiedTableNames map[string]struct{}
	// rightMostTableName holds the table name of the rightmost table seen during traversal.
	rightMostTableName string
}

// optimizeSubtree takes a cross-join subtree configuration and returns a reconstructed
// inner-join subtree that groups leaf nodes with equality predicate terms close together.
// It returns nil if it's unable to perform the reconstruction.
func (v *crossJoinOptimizer) optimizeSubtree() Node {

	/* Here's what happens in this function:
	1. We use all conjunctive predicate terms that were found above the cross join subtree to
		figure out what pair of tables can be subsequently linked using a $lookup.
	2. These predicates are then used to build a graph where the nodes are tables and the links
		are undirected edges between pairs of tables.
	3. Using the graph, we find all connected components - these represent subsets of
		conjunctive terms that can be translated into sequential lookup operations.
		Within each component, we try to sort the nodes in order of which have a cardinality
		altering predicate - the goal being to cut down cardinality sooner.
	4. With these now sorted connected components, we reconstruct the cross join subtree using
		the initial cross join operation.
	*/

	skipParts := expressionParts{}
	partsToUse := make(map[string]SQLExpr)
	cardinalityAlteringPredicates := make(map[string]expressionPart)

	g := newCrossJoinGraph()
	createKey := func(leftTable, rightTable string) string {
		return fmt.Sprintf("%v-%v", leftTable, rightTable)
	}

	// Create the cross join graph using conjunctive terms - if any exists - within the predicate
	// applied on the entire subtree.
	for _, part := range v.fullPredicateParts {
		if tableRefCount := len(part.qualifiedTableNames); tableRefCount != 2 {
			if tableRefCount == 1 {
				cardinalityAlteringPredicates[part.qualifiedTableNames[0]] = part
			}
			skipParts = append(skipParts, part)
			continue
		}

		leftTable, rightTable := part.qualifiedTableNames[0], part.qualifiedTableNames[1]

		key := createKey(leftTable, rightTable)
		reversedKey := createKey(rightTable, leftTable)
		if _, ok := partsToUse[key]; !ok {
			partsToUse[key], partsToUse[reversedKey] = part.expr, part.expr
			g.addEdge(leftTable, rightTable)
		} else {
			skipParts = append(skipParts, part)
		}
	}

	// Get all connected components present in the graph - ordering each component with a
	// cardinality altering predicate first.
	sortedComponents := g.connectedComponents(cardinalityAlteringPredicates)

	// Now sort all components so those with cardinality altering predicates come first.
	sort.Sort(sortedComponents)
	joinedTables := make(map[string]struct{}, len(v.sources))

	var newN PlanStage

	// Join each reconstructed inner join subtree.
	for i := 0; i < len(sortedComponents); i++ {
		sortedComponent := sortedComponents[i].component
		var componentN PlanStage
		// Within each component, translate the cross join subtree to an inner join using the
		// predicates linking each pair of nodes.
		for j := 0; j < len(sortedComponent)-1; j++ {
			leftTable, rightTable := sortedComponent[j], sortedComponent[j+1]
			unJoinedTable := leftTable
			if _, ok := joinedTables[leftTable]; ok {
				unJoinedTable = rightTable
			}
			unselfJoinedSource, ok := v.sources[unJoinedTable].dataSource.(PlanStage)
			if !ok {
				panic(fmt.Sprintf("cross join optimizer: expected PlanStage for table %v, got %T",
					unJoinedTable, v.sources[unJoinedTable].dataSource))
			}
			joinedTables[unJoinedTable] = struct{}{}

			var partExpr SQLExpr
			if componentN == nil {
				var right PlanStage
				right, ok = v.sources[rightTable].dataSource.(PlanStage)
				if !ok {
					componentN = unselfJoinedSource
					break
				}
				componentN, unselfJoinedSource = unselfJoinedSource, right
				joinedTables[rightTable] = struct{}{}
				partExpr = partsToUse[createKey(unJoinedTable, rightTable)]
			}

			if partExpr == nil {
				for k := 0; k <= j; k++ {
					partExpr, ok = partsToUse[createKey(sortedComponent[k], unJoinedTable)]
					if ok {
						break
					}
				}
			}

			if partExpr == nil {
				v.logger.Warnf(log.Dev, "cross join optimizer: couldn't find link to table %v",
					unJoinedTable)
				return nil
			}
			componentN = NewJoinStage(InnerJoin, componentN, unselfJoinedSource, partExpr)
		}

		if newN == nil {
			newN = componentN
		} else {
			newN = NewJoinStage(CrossJoin, newN, componentN, SQLTrue)
		}
	}

	// Sort all sources within the subtree.
	sortedSources := []string{}
	for source := range v.sources {
		sortedSources = append(sortedSources, source)
	}

	sort.Strings(sortedSources)

	// Add back source plan stages that weren't found in the conjunctive predicate terms.
	for _, name := range sortedSources {
		if _, ok := joinedTables[name]; ok {
			continue
		}
		newSource, ok := v.sources[name].dataSource.(PlanStage)
		if !ok {
			panic(fmt.Sprintf("cross join optimizer: expected PlanStage for source %v, got %T",
				name, v.sources[name].dataSource))
		}
		if newN == nil {
			newN = newSource
		} else {
			newN = NewJoinStage(CrossJoin, newN, newSource, SQLTrue)
		}
	}

	// Add back any conjunctive predicate terms that weren't used.
	if len(skipParts) > 0 {
		newNPlanStage, ok := newN.(PlanStage)
		if !ok {
			panic(fmt.Sprintf("cross join optimizer: expected PlanStage for skipped parts got %T",
				newN))
		}
		newN = NewFilterStage(newNPlanStage, skipParts.combine())
	}

	return newN
}

func (v *crossJoinOptimizer) visit(n Node) (Node, error) {
	var err error
	switch typedN := n.(type) {
	case *DynamicSourceStage:
		fqtn := fullyQualifiedTableName(typedN.dbName, typedN.aliasName)
		v.sources[fqtn] = joinLeafSource{dataSource: typedN}
		v.rightMostTableName = fqtn
		return n, nil

	case *FilterStage:
		// save the old parts before assigning a new one
		old := v.predicateParts
		v.predicateParts, err = splitExpressionIntoParts(typedN.matcher)
		if err != nil {
			return nil, err
		}
		v.fullPredicateParts = v.predicateParts

		// Walk the children and let the joins optimize with relevant predicateParts
		source, err := v.visit(typedN.source)
		if err != nil {
			return nil, err
		}

		if source != typedN.source {
			if len(v.predicateParts) > 0 {
				// if the parts haven't been fully utilized,
				// add a Filter back into the tree with the remaining
				// parts.
				n = NewFilterStage(source.(PlanStage), v.predicateParts.combine())
			} else {
				n = source
			}
		}

		// reset the parts back to the way it was
		v.predicateParts = old
		return n, nil

	case *JoinStage:
		if !canOptimizeJoinSubtree(typedN.left) || !canOptimizeJoinSubtree(typedN.right) {
			return n, nil
		}

		matcherOk := typedN.matcher == nil
		if !matcherOk {
			switch typedM := typedN.matcher.(type) {
			case SQLBool:
				matcherOk = Float64(typedM) > 0
			default:
				predicateParts, err := splitExpressionIntoParts(typedN.matcher)
				if err != nil {
					return nil, err
				}
				v.fullPredicateParts = append(v.fullPredicateParts, predicateParts...)
			}
		}

		// We have a filter and a join without any criteria and can thus apply the cross join
		// optimization.
		if matcherOk && len(v.predicateParts) > 0 &&
			(typedN.kind == InnerJoin || typedN.kind == CrossJoin || typedN.kind == StraightJoin) {
			oldIsChildJoinNode := v.isChildJoinNode
			v.isChildJoinNode = true
			left, err := v.visit(typedN.left)
			if err != nil {
				return nil, err
			}

			// Save table names from left subtree.
			tableNames := make(map[string]struct{}, len(v.qualifiedTableNames))
			for tableName := range v.qualifiedTableNames {
				tableNames[tableName] = struct{}{}
			}

			v.qualifiedTableNames = make(map[string]struct{})
			right, err := v.visit(typedN.right)
			if err != nil {
				return nil, err
			}
			v.isChildJoinNode = oldIsChildJoinNode

			// Merge the table names from both the left and right subtrees.
			for tableName := range tableNames {
				v.qualifiedTableNames[tableName] = struct{}{}
			}
			// go through each part of the predicate and figure out which
			// ones are associated to the tables in the current join.
			partsToUse := expressionParts{}
			savedParts := v.predicateParts
			v.predicateParts = nil
			for _, part := range savedParts {
				if v.canUseExpressionPartInJoinClause(part) {
					partsToUse = append(partsToUse, part)
				} else {
					v.predicateParts = append(v.predicateParts, part)
				}
			}

			// If parts of the left or right have been changed, we need a new join operator.
			if len(partsToUse) > 0 || left != typedN.left || right != typedN.right {
				var predicate SQLExpr
				kind := CrossJoin
				if len(partsToUse) > 0 {
					kind = InnerJoin
					predicate = partsToUse.combine()
				}

				if predicate == nil {
					predicate = SQLTrue
				}

				n = NewJoinStage(kind, left.(PlanStage), right.(PlanStage), predicate)
			}

			// We are now at the root node for the entire cross join subtree, try to reconstruct
			// the entire subtree in a more optimal fashion
			if !v.isChildJoinNode {
				if r := v.optimizeSubtree(); r != nil {
					n, v.predicateParts = r, nil
				}
			}

			return n, nil
		}

	case *MongoSourceStage:
		for _, alias := range typedN.aliasNames {
			fqtn := fullyQualifiedTableName(typedN.dbName, alias)
			v.qualifiedTableNames[fqtn] = struct{}{}
		}

		fqtn := fullyQualifiedTableName(typedN.dbName, typedN.aliasNames[0])
		v.sources[fqtn] = joinLeafSource{dataSource: typedN}
		v.rightMostTableName = fqtn
		return n, nil

	case *SubquerySourceStage:
		subqueryOptimizer := newCrossJoinOptimizer(v.logger)
		plan, err := subqueryOptimizer.visit(typedN.source)
		if err != nil {
			return nil, err
		}

		if typedN.source != plan.(PlanStage) {
			n = NewSubquerySourceStage(
				plan.(PlanStage),
				typedN.selectID,
				typedN.dbName,
				typedN.aliasName,
				typedN.fromCTE,
			)
		}

		fqtn := fullyQualifiedTableName(typedN.dbName, typedN.aliasName)
		v.qualifiedTableNames[fqtn] = struct{}{}
		v.sources[fqtn] = joinLeafSource{dataSource: n}
		v.rightMostTableName = fqtn
		return n, nil

	case *UnionStage:
		newV := newCrossJoinOptimizer(v.logger)
		newRight, err := newV.visit(typedN.right)
		if err != nil {
			return nil, err
		}

		newV = newCrossJoinOptimizer(v.logger)
		newLeft, err := newV.visit(typedN.left)
		if err != nil {
			return nil, err
		}

		if typedN.left != newLeft.(PlanStage) || typedN.right != newRight.(PlanStage) {
			n = NewUnionStage(typedN.kind, newLeft.(PlanStage), newRight.(PlanStage))
		}
		return n, nil
	}

	return walk(v, n)
}

func (v *crossJoinOptimizer) canUseExpressionPartInJoinClause(part expressionPart) bool {
	if len(part.qualifiedTableNames) > 0 {
		// the right-most table must be present in the part
		if !containsString(part.qualifiedTableNames, v.rightMostTableName) {
			return false
		}

		// all the names in the part must be in scope
		for _, n := range part.qualifiedTableNames {
			if _, ok := v.qualifiedTableNames[n]; !ok {
				return false
			}
		}

		return true
	}

	return false
}

// crossJoinGraph holds a graph that a cross join subtree. It is used during cross join optimization
// by modelling predicates containing two tables as undirected edges (neighbors) in the graph.
// For example in the query `select * from a, b where a.id = b.id` the graph created will contain
// an undirected edge between `a` and `b`.
type crossJoinGraph struct {
	neighbors map[string][]string
}

// newCrossJoinGraph creates a new crossJoinGraph.
func newCrossJoinGraph() *crossJoinGraph {
	return &crossJoinGraph{
		neighbors: make(map[string][]string),
	}
}

// addEdge adds an undirected edge between from and to.
func (g *crossJoinGraph) addEdge(from, to string) {
	g.neighbors[from] = append(g.neighbors[from], to)
	g.neighbors[to] = append(g.neighbors[to], from)
}

// getComponents recursively goes through the graph and returns a traversal of all unvisited nodes
// reachable from node. It reorders the nodes starting with whichever node contains a
// cardinality altering predicate, if one exists.
func (g *crossJoinGraph) getComponents(node string, visited map[string]struct{},
	cardinalityAlteringPredicates map[string]expressionPart) sortableComponent {

	nodes := []string{}
	var visit func(string, map[string]struct{})
	visit = func(node string, visited map[string]struct{}) {
		visited[node] = struct{}{}
		nodes = append(nodes, node)
		for _, neighbors := range g.neighbors[node] {
			if _, ok := visited[neighbors]; !ok {
				visit(neighbors, visited)
			}
		}
	}

	visit(node, visited)

	for i, node := range nodes {
		if _, ok := cardinalityAlteringPredicates[node]; ok {
			newNodes := []string{nodes[i]}
			for j := i - 1; j >= 0; j-- {
				newNodes = append(newNodes, nodes[j])
			}

			for k := i + 1; k < len(nodes); k++ {
				newNodes = append(newNodes, nodes[k])
			}
			return sortableComponent{newNodes, true}
		}
	}
	return sortableComponent{nodes, false}
}

// connectedComponents performs a depth-first traversal of the cross join graph and returns a slice
// of all the connected components it finds.
func (g *crossJoinGraph) connectedComponents(
	cardinalityAlteringPredicates map[string]expressionPart) sortableComponents {

	visited := map[string]struct{}{}
	sortedNeighbors, connectedComponents := []string{}, sortableComponents{}
	for node := range g.neighbors {
		sortedNeighbors = append(sortedNeighbors, node)
	}
	sort.Strings(sortedNeighbors)
	for _, node := range sortedNeighbors {
		if _, ok := visited[node]; !ok {
			connectedComponent := g.getComponents(node, visited, cardinalityAlteringPredicates)
			connectedComponents = append(connectedComponents, connectedComponent)
		}
	}
	return connectedComponents
}

// sortableComponents sort helpers.
type sortableComponents []sortableComponent

type sortableComponent struct {
	component                       []string
	hasCardinalityAlteringPredicate bool
}

func (s sortableComponents) Len() int {
	return len(s)
}

func (s sortableComponents) Less(i, j int) bool {
	// Inner join subtrees containing cardinality-altering predicates should be executed sooner.
	if s[i].hasCardinalityAlteringPredicate {
		return true
	}
	if s[j].hasCardinalityAlteringPredicate {
		return false
	}
	return len(s[i].component) > len(s[j].component)
}

func (s sortableComponents) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
