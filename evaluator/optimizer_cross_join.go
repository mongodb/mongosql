package evaluator

import (
	"fmt"
	"sort"

	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/strutil"
	"github.com/10gen/sqlproxy/log"
)

func optimizeCrossJoins(cfg *OptimizerConfig, n Node) (Node, error) {
	if !cfg.optimizeCrossJoins {
		cfg.lg.Warnf(log.Admin, "optimize_cross_joins is false: skipping cross join optimizer")
		return n, nil
	}

	n, err := newCrossJoinOptimizer(cfg).visit(n)
	if err != nil {
		return nil, err
	}

	return n, nil
}

func newCrossJoinOptimizer(cfg *OptimizerConfig) *crossJoinOptimizer {
	return &crossJoinOptimizer{
		cfg:                 cfg,
		planStages:          make(map[string]joinLeafSource),
		qualifiedTableNames: make(map[string]struct{}),
	}
}

type crossJoinOptimizer struct {
	cfg *OptimizerConfig
	// filter holds conjunctive terms from a filter expression on the cross join subtree.
	// It is updated as the subtree is traversed.
	filter expressionParts
	// filterAndJoinCriteria is the same as filter above, except it includes predicates
	// gotten from the subtree's join criteria.
	filterAndJoinCriteria expressionParts
	// isChildJoinNode is true if the optimizer visitor is at a node rooting a cross join subtree
	// and false otherwise.
	isChildJoinNode bool
	// planStages is a map of fully qualified table names to leaf PlanStages.
	planStages map[string]joinLeafSource
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

	valueKind := v.cfg.sqlValueKind
	predicatesToSkip := expressionParts{}
	predicatesToUse := make(map[string]SQLExpr)
	cardinalityAlteringPredicates := make(map[string]expressionPart)

	g := newCrossJoinGraph()
	createJoinTableKey := func(leftTable, rightTable string) string {
		return fmt.Sprintf("%v-%v", leftTable, rightTable)
	}

	// Create the cross join graph using conjunctive terms - if any exists - within the predicate
	// applied on the entire subtree.
	for _, predicate := range v.filterAndJoinCriteria {
		// All the table names in the predicate must be within the optimizer's scope.
		canUsePredicate := true
		for _, n := range predicate.qualifiedTableNames {
			if _, ok := v.qualifiedTableNames[n]; !ok {
				predicatesToSkip = append(predicatesToSkip, predicate)
				canUsePredicate = false
				break
			}
		}

		if !canUsePredicate {
			continue
		}

		if tableRefCount := len(predicate.qualifiedTableNames); tableRefCount != 2 {
			if tableRefCount == 1 {
				cardinalityAlteringPredicates[predicate.qualifiedTableNames[0]] = predicate
			}
			predicatesToSkip = append(predicatesToSkip, predicate)
			continue
		}

		leftTable, rightTable := predicate.qualifiedTableNames[0], predicate.qualifiedTableNames[1]

		joinTableKey := createJoinTableKey(leftTable, rightTable)
		reversedJoinTableKey := createJoinTableKey(rightTable, leftTable)
		if _, ok := predicatesToUse[joinTableKey]; !ok {
			predicatesToUse[joinTableKey] = predicate.expr
			predicatesToUse[reversedJoinTableKey] = predicate.expr
			g.addEdge(leftTable, rightTable)
		} else {
			predicatesToSkip = append(predicatesToSkip, predicate)
		}
	}

	// Get all connected components present in the graph - ordering each component with a
	// cardinality altering predicate first.
	sortedComponents := g.connectedComponents(cardinalityAlteringPredicates)

	// Now sort all components so those with cardinality altering predicates come first.
	sort.Sort(sortedComponents)
	joinedTables := make(map[string]struct{}, len(v.planStages))

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

			unJoinedPlanStage, ok := v.planStages[unJoinedTable].dataSource.(PlanStage)
			if !ok {
				panic(fmt.Sprintf("cross join optimizer: expected PlanStage for table %v, got %T",
					unJoinedTable, v.planStages[unJoinedTable].dataSource))
			}
			joinedTables[unJoinedTable] = struct{}{}

			var predicateExpr SQLExpr
			if componentN == nil {
				var right PlanStage
				right, ok = v.planStages[rightTable].dataSource.(PlanStage)
				if !ok {
					componentN = unJoinedPlanStage
					break
				}
				componentN, unJoinedPlanStage = unJoinedPlanStage, right
				joinedTables[rightTable] = struct{}{}
				predicateExpr = predicatesToUse[createJoinTableKey(unJoinedTable, rightTable)]
			}

			if predicateExpr == nil {
				for k := 0; k <= j; k++ {
					key := createJoinTableKey(sortedComponent[k], unJoinedTable)
					predicateExpr, ok = predicatesToUse[key]
					if ok {
						break
					}
				}
			}

			if predicateExpr == nil {
				v.cfg.lg.Warnf(log.Dev, "cross join optimizer: couldn't find link to table %v",
					unJoinedTable)
				return nil
			}
			componentN = NewJoinStage(InnerJoin, componentN, unJoinedPlanStage, predicateExpr)
		}

		if newN == nil {
			newN = componentN
		} else {
			newN = NewJoinStage(CrossJoin, newN, componentN, NewSQLValueExpr(values.NewSQLBool(valueKind, true)))
		}
	}

	// Sort all source PlanStages within the subtree.
	sortedPlanStages := []string{}
	for planStage := range v.planStages {
		sortedPlanStages = append(sortedPlanStages, planStage)
	}

	sort.Strings(sortedPlanStages)

	// Add back source plan stages that weren't found in the conjunctive predicate terms.
	for _, name := range sortedPlanStages {
		// If the PlanStage has already been added, move on.
		if _, ok := joinedTables[name]; ok {
			continue
		}

		newPlanStage, ok := v.planStages[name].dataSource.(PlanStage)
		if !ok {
			panic(fmt.Sprintf("cross join optimizer: expected PlanStage for source %v, got %T",
				name, v.planStages[name].dataSource))
		}

		if newN == nil {
			newN = newPlanStage
		} else {
			newN = NewJoinStage(CrossJoin, newN, newPlanStage, NewSQLValueExpr(values.NewSQLBool(valueKind, true)))
		}
	}

	// Add back any conjunctive predicate terms that weren't used.
	if len(predicatesToSkip) > 0 {
		newPlanStage, ok := newN.(PlanStage)
		if !ok {
			panic(fmt.Sprintf("cross join optimizer: expected PlanStage for skipped predicates"+
				" got %T", newN))
		}
		// Reassign the covering Filter to nil and add the predicates that were skipped back to
		// the subtree.
		newN, v.filter = NewFilterStage(newPlanStage, predicatesToSkip.combine()), nil
	}

	return newN
}

func (v *crossJoinOptimizer) visit(n Node) (Node, error) {
	valueKind := v.cfg.sqlValueKind
	var err error
	switch typedN := n.(type) {
	case *DynamicSourceStage:
		fqtn := fullyQualifiedTableName(typedN.dbName, typedN.aliasName)
		v.planStages[fqtn] = joinLeafSource{dataSource: typedN}
		v.rightMostTableName = fqtn
		return n, nil

	case *FilterStage:
		// Save the old predicates before assigning a new one.
		oldFilter, oldfilterAndJoinCriteria := v.filter, v.filterAndJoinCriteria
		v.filter, err = getConjunctiveTerms(typedN.matcher)
		if err != nil {
			return nil, err
		}
		v.filterAndJoinCriteria = v.filter

		// Walk the children and let the joins optimize with relevant predicates.
		source, err := v.visit(typedN.source)
		if err != nil {
			return nil, err
		}

		if source != typedN.source {
			if len(v.filter) > 0 {
				// If the predicates haven't been fully utilized, add a Filter back into the tree
				// with the remaining predicates.
				n = NewFilterStage(source.(PlanStage), v.filter.combine())
			} else {
				n = source
			}
		}

		// Reset the predicates back to the way it was.
		v.filter, v.filterAndJoinCriteria = oldFilter, oldfilterAndJoinCriteria
		return n, nil

	case *JoinStage:
		if !canOptimizeJoinSubtree(typedN.left) || !canOptimizeJoinSubtree(typedN.right) {
			return n, nil
		}

		hasCardinalityReducingJoinCriteria := typedN.matcher != nil
		if hasCardinalityReducingJoinCriteria {
			switch typedM := typedN.matcher.(type) {
			case SQLValueExpr:
				// If the criteria is the boolean true - (1) - then it's not
				// helpful in reducing cardinality.
				hasCardinalityReducingJoinCriteria = values.Float64(typedM.Value) != 1
			default:
				conjunctiveCriteria, err := getConjunctiveTerms(typedN.matcher)
				if err != nil {
					return nil, err
				}
				v.filterAndJoinCriteria = append(v.filterAndJoinCriteria, conjunctiveCriteria...)
			}
		}

		isCommutativeJoinKind := strutil.StringSliceContains(commutativeJoinKinds, string(typedN.kind))

		// We have a filter and a join without any criteria and can thus apply the cross join
		// optimization.
		if isCommutativeJoinKind {
			if !hasCardinalityReducingJoinCriteria && len(v.filter) > 0 {
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

				// Go through each predicate of the predicate and figure out which
				// ones are associated to the tables in the current join.
				predicatesToUse := expressionParts{}
				savedFilterPredicates := v.filter
				v.filter = nil
				for _, predicate := range savedFilterPredicates {
					if v.canUsePredicateInJoinClause(predicate) {
						predicatesToUse = append(predicatesToUse, predicate)
					} else {
						v.filter = append(v.filter, predicate)
					}
				}

				// If predicates of the left or right have been changed, we need a new join
				// operator.
				if len(predicatesToUse) > 0 || left != typedN.left || right != typedN.right {
					var predicate SQLExpr
					kind := CrossJoin
					if len(predicatesToUse) > 0 {
						kind = InnerJoin
						predicate = predicatesToUse.combine()
					}

					if predicate == nil {
						predicate = NewSQLValueExpr(values.NewSQLBool(valueKind, true))
					}

					n = NewJoinStage(kind, left.(PlanStage), right.(PlanStage), predicate)
				}

				// We are now at the root node for the entire cross join subtree, try to reconstruct
				// the entire subtree in a more optimal fashion if there are sufficient sources to
				// reorder. We check for the number of PlanStages in the optimizer is greater than
				// since we might have a query plan tree thatonly has one side visited but not the
				// other, e.g.
				//
				//          cross join
				//          /       \
				// right join      cross join
				//
				if !v.isChildJoinNode && len(v.planStages) > 1 {
					if r := v.optimizeSubtree(); r != nil {
						n = r
					}
				}
			}
		} else {
			// If we hit a node level where we're unable to optimize - e.g. if it's a left join or a
			// right join - we can possibly further optimize the subtree rooted at this node.
			// For example, in the plan tree below, we can optimize the subtree rooted in B.
			//
			//				A(CrossJoin)
			//				/	\
			//			B(RightJoin)	 C
			//			/	\
			//		D(CrossJoin)	 E
			//		/	\
			//		F	 G

			newL, err := newCrossJoinOptimizer(v.cfg).visit(typedN.left)
			if err != nil {
				return nil, err
			}
			newR, err := newCrossJoinOptimizer(v.cfg).visit(typedN.right)
			if err != nil {
				return nil, err
			}
			if typedN.left != newL.(PlanStage) || typedN.right != newR.(PlanStage) {
				n = NewJoinStage(typedN.kind, newL.(PlanStage), newR.(PlanStage), typedN.matcher)
			}
		}

		return n, nil

	case *MongoSourceStage:
		for _, alias := range typedN.aliasNames {
			fqtn := fullyQualifiedTableName(typedN.dbName, alias)
			v.qualifiedTableNames[fqtn] = struct{}{}
		}

		fqtn := fullyQualifiedTableName(typedN.dbName, typedN.aliasNames[0])
		v.planStages[fqtn] = joinLeafSource{dataSource: typedN}
		v.rightMostTableName = fqtn
		return n, nil

	case *SubquerySourceStage:
		subqueryOptimizer := newCrossJoinOptimizer(v.cfg)
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
		v.planStages[fqtn] = joinLeafSource{dataSource: n}
		v.rightMostTableName = fqtn
		return n, nil

	case *UnionStage:
		newV := newCrossJoinOptimizer(v.cfg)
		newRight, err := newV.visit(typedN.right)
		if err != nil {
			return nil, err
		}

		newV = newCrossJoinOptimizer(v.cfg)
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

func (v *crossJoinOptimizer) canUsePredicateInJoinClause(predicate expressionPart) bool {
	if len(predicate.qualifiedTableNames) > 0 {
		// The right-most table must be present in the predicate.
		if !containsString(predicate.qualifiedTableNames, v.rightMostTableName) {
			return false
		}

		// All the names in the predicate must be in scope.
		for _, n := range predicate.qualifiedTableNames {
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
