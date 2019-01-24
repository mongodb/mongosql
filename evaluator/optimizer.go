package evaluator

import (
	"context"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/log"
)

var (
	// commutativeJoinKinds holds JoinKinds that can be reordered without
	// any loss of semantic meaning - i.e. commutative JoinKinds.
	commutativeJoinKinds = []string{
		string(CrossJoin),
		string(InnerJoin),
		string(StraightJoin),
	}
)

// OptimizerConfig is a container for all the values needed to run the optimizers.
type OptimizerConfig struct {
	lg           log.Logger
	collation    *collation.Collation
	sqlValueKind SQLValueKind

	// flags for enabling/disabling individual optimizers
	optimizeCrossJoins  bool
	optimizeEvaluations bool
	optimizeFiltering   bool
	optimizeInnerJoins  bool
}

// NewOptimizerConfig returns a new OptimizerConfig constructed from the
// provided values. OptimizerConfigs should always be constructed via this
// function instead of via a struct literal.
func NewOptimizerConfig(lg log.Logger, vars catalog.VariableContainer) *OptimizerConfig {
	return &OptimizerConfig{
		lg:                  lg,
		collation:           vars.GetCollation(variable.CollationConnection),
		sqlValueKind:        GetSQLValueKind(vars),
		optimizeCrossJoins:  vars.GetBool(variable.OptimizeCrossJoins),
		optimizeEvaluations: vars.GetBool(variable.OptimizeEvaluations),
		optimizeFiltering:   vars.GetBool(variable.OptimizeFiltering),
		optimizeInnerJoins:  vars.GetBool(variable.OptimizeInnerJoins),
	}
}

// OptimizePlan applies optimizations to the plan tree to
// aid in performance.
func OptimizePlan(ctx context.Context, cfg *OptimizerConfig, p PlanStage) (PlanStage, error) {
	cfg.lg.Debugf(log.Dev, "optimizing query plan: \n%v", PrettyPrintPlan(p))
	n, err := optimize(ctx, cfg, p)
	if err != nil {
		return nil, err
	}
	return n.(PlanStage), nil
}

type optimizerStage struct {
	name string
	f    func(*OptimizerConfig, Node) (Node, error)
}

var optimizerStages = []optimizerStage{
	{"evaluations", OptimizeEvaluations},
	{"cross joins", optimizeCrossJoins},
	{"inner join", optimizeInnerJoins},
	{"filtering", optimizeFiltering},
}

func optimize(ctx context.Context, cfg *OptimizerConfig, n Node) (Node, error) {
	for _, stage := range optimizerStages {
		cfg.lg.Infof(log.Dev, "running optimization stage '%s'", stage.name)
		var err error
		n, err = stage.f(cfg, n)

		_, pde := err.(*pushdownError)
		if err != nil && !pde {
			cfg.lg.Warnf(log.Admin, "error running optimization stage '%s': %v", stage.name, err)
			// don't exit here. Just because we couldn't apply one optimization doesn't mean
			// others aren't valid
		} else {
			cfg.lg.Debugf(log.Dev, "optimized plan after"+
				" '%s': \n%v", stage.name, prettyPrintNode(n))
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}

	return n, nil
}

func combineExpressions(exprs []SQLExpr) SQLExpr {
	var combined SQLExpr
	if len(exprs) > 0 {
		combined = exprs[0]
		for _, expr := range exprs[1:] {
			combined = NewSQLAndExpr(combined, expr)
		}
	}
	return combined
}

func sharesRootTable(logger log.Logger, local, foreign *MongoSourceStage) bool {
	baseCollectionName := local.collectionNames[0]
	if local.dbName != foreign.dbName {
		return false
	}
	logger.Debugf(log.Dev, "attempting to use self-join optimization for tables %v and %v",
		local.aliasNames, foreign.aliasNames)

	for _, collectionName := range append(local.collectionNames[1:],
		foreign.collectionNames...) {
		if collectionName != baseCollectionName {
			logger.Debugf(log.Dev, "cannot use self-join optimization, "+
				"pipeline has different root tables: %v and %v",
				baseCollectionName, collectionName)
			return false
		}
	}

	return true
}

func splitExpression(e SQLExpr) []SQLExpr {
	andE, ok := e.(*SQLAndExpr)
	if !ok {
		return []SQLExpr{e}
	}

	left := splitExpression(andE.left)
	right := splitExpression(andE.right)
	return append(left, right...)
}

// getConjunctiveTerms splits hierarchical SQLAndExprs within e
// into a flattened list.
func getConjunctiveTerms(e SQLExpr) (expressionParts, error) {
	exprs := splitExpression(e)
	result := []expressionPart{}
	for _, expr := range exprs {
		qualifiedTableNames, err := referencedTables(expr)
		if err != nil {
			return nil, err
		}
		result = append(result, expressionPart{expr, qualifiedTableNames})
	}
	return result, nil
}

type expressionParts []expressionPart

func (parts expressionParts) combine() SQLExpr {
	var combined SQLExpr
	if len(parts) > 0 {
		combined = parts[0].expr
		for _, part := range parts[1:] {
			combined = NewSQLAndExpr(combined, part.expr)
		}
	}
	return combined
}

type expressionPart struct {
	expr                SQLExpr
	qualifiedTableNames []string
}

func referencedTables(e SQLExpr) ([]string, error) {
	finder := &sqlExprReferencedTableCollector{
		qualifiedTableNames: make(map[string]struct{}),
	}
	_, err := finder.visit(e)
	if err != nil {
		return nil, err
	}
	qualifiedTableNames := []string{}
	for fqtn := range finder.qualifiedTableNames {
		qualifiedTableNames = append(qualifiedTableNames, fqtn)
	}
	return qualifiedTableNames, nil
}

type sqlExprReferencedTableCollector struct {
	qualifiedTableNames map[string]struct{}
}

func (v *sqlExprReferencedTableCollector) visit(n Node) (Node, error) {
	switch typedN := n.(type) {
	case SQLColumnExpr:
		if _, ok := v.qualifiedTableNames[typedN.tableName]; !ok {
			fqtn := fullyQualifiedTableName(typedN.databaseName, typedN.tableName)
			v.qualifiedTableNames[fqtn] = struct{}{}
		}
	}
	return walk(v, n)
}

// joinLeafSource holds all data sources for a join subtree.
type joinLeafSource struct {
	// nPipelineStages holds the number of pipeline stages contained
	// within a data source. For subqueries, it adds the number of
	// PlanStages contained within subquery.
	nPipelineStages int
	dataSource      Node
}
