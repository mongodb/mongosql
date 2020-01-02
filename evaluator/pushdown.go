package evaluator

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/optimizer"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/internal/astutil"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/procutil"
	"github.com/10gen/sqlproxy/internal/strutil"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
)

const (
	filterStageName         = "FilterStage"
	groupByStageName        = "GroupByStage"
	joinStageName           = "JoinStage"
	limitStageName          = "LimitStage"
	orderByStageName        = "OrderByStage"
	projectStageName        = "ProjectStage"
	subquerySourceStageName = "SubquerySourceStage"
	unionStageName          = "UnionStage"
)

// PushdownConfig is a container for all the values needed
// to run the pushdown translator.
type PushdownConfig struct {
	lg                log.Logger
	shouldPushDown    bool
	mongoDBVersion    []uint8
	pushDownSelfJoins bool
	sqlValueKind      values.SQLValueKind
	format            string
	formatVersion     int
}

// NewPushdownConfig returns a new PushdownConfig constructed from the
// provided values. PushdownConfigs should always be constructed via this
// function instead of via a struct literal.
func NewPushdownConfig(lg log.Logger, vars catalog.VariableContainer, format string, formatVersion int) *PushdownConfig {
	return &PushdownConfig{
		lg:                lg,
		mongoDBVersion:    getMongoDBVersion(vars),
		shouldPushDown:    vars.GetBool(variable.Pushdown),
		pushDownSelfJoins: vars.GetBool(variable.OptimizeSelfJoins),
		sqlValueKind:      GetSQLValueKind(vars),
		format:            format,
		formatVersion:     formatVersion,
	}
}

// PushdownPlan translates as much of the provided plan as possible into
// an aggregation pipeline, returning an updated plan. If the resulting
// plan is not fully pushed down, a pushdownFailor will be returned, but
// the returned plan is still valid. If any other kind of error occurs,
// it will be returned along wth a nil plan.
func PushdownPlan(ctx context.Context, cfg *PushdownConfig, p PlanStage) (PlanStage, PushdownError) {

	if !cfg.shouldPushDown {
		cfg.lg.Warnf(log.Admin, "pushdown is disabled, skipping translation")
		return p, nil
	}

	v := newPushdownVisitor(cfg)

	n, err := v.visit(p)
	if err != nil {
		return nil, fatalPushdownError(err)
	}
	p = n.(PlanStage)

	if cfg.format != NoOutputFormat {
		cfg.lg.Debugf(log.Admin, "formatting results")
		if ms, ok := p.(*MongoSourceStage); ok {
			var project *ast.ProjectStage
			project, err = formatPushdownOutput(ms, cfg.format, cfg.formatVersion)
			if err != nil {
				cfg.lg.Debugf(log.Admin, "failed to format: %v", err)
			} else {
				ms.pipeline.Stages = append(ms.pipeline.Stages, project)
			}
		} else {
			cfg.lg.Debugf(log.Admin, "cannot format results; plan is not a MongoSourceStage")
		}
	}

	cfg.lg.Debugf(log.Admin,
		"plan before pipeline optimization: \n%v",
		PrettyPrintPlan(p),
	)
	pipelineOptimizationVisitor := newPipelineOptimizationVisitor(ctx)
	n, err = pipelineOptimizationVisitor.visit(p)
	if err != nil {
		return nil, fatalPushdownError(err)
	}

	p = n.(PlanStage)

	cfg.lg.Debugf(log.Admin,
		"plan after pipeline optimization: \n%v",
		PrettyPrintPlan(p),
	)

	if len(v.pushdownFailures) != 0 {
		return p, nonFatalPushdownError(v.pushdownFailures)
	}

	return p, nil
}

// dbData is a map from database name to tableData.
type dbData map[string]tableData

// tableData is a map from table name to columnData.
type tableData map[string]columnData

// columnData is map from original column name to its rename.
type columnData map[string]string

var emptyDbData = make(dbData)

// correlatedColumnDb is a map from database name to tableData.
type correlatedColumnDb map[string]correlatedColumnTable

// correlatedColumnTable is a map from table name to columnData.
type correlatedColumnTable map[string]correlatedColumnName

// correlatedColumnName is map from original column name to its new name
type correlatedColumnName map[string]string

type pushdownVisitor struct {
	cfg    *PushdownConfig
	logger log.Logger
	// selectIDsInScope keeps track of which selectIDs are in scope at a given
	// point in the query plan, this allows us to keep track of which referenced
	// columns are actually defined at this current point in the query plan, which
	// is needed for various things, such as figuring out which fields must be projected
	// in the case of a partially pushed down ProjectStage.
	selectIDsInScope []int
	// tableNamesInScope keeps track of the table names visible at a current
	// point in the query plan for each database.
	tableNamesInScope map[string][]string
	columnTracker     *columnTracker
	// leftJoinOriginalNames keeps track of the original names of nullable columns in left joins that
	// have been self-join optimized.  The reason for this is we need to know the original name when
	// the column in used in a predicate for matching purposes. The renamed field only exists to add
	// nulls where they should be added due to failure in an `on` clause.
	leftJoinOriginalNames dbData
	depth                 int
	// pushdownFailures keeps track of the pushdownFailures associated with each PlanStage in a
	// Plan.
	pushdownFailures map[PlanStage][]PushdownFailure
	// canPushdownCorrelated tells the ToAggregationLanguage method for SQLColumnExpr whether it is
	// safe to push down when the column is correlated. This is necessary because we need to save
	// the original subquery plan(s) in case push down fails at a higher level. The essential
	// problem is that for every type of expression other than *correlated* subqueries, it is safe
	// to pushdown in a completely bottom-up fashion, but for correlated subqueries it is not
	// because the child plans must be returned to non-pushed down form (at least with respected to
	// correlated columns) when the subquery itself fails to pushdown for some other reason (e.g., a
	// lack of a `limit 1`).
	canPushdownCorrelated bool
	// correlatedColumns keeps track of the correlated columns that exist such that when we translate
	// the subquery that uses the correlatedColumns we know if it can still be pushed down. That is,
	// if a column in the subquery does not appear in the correlatedColumns map, the subquery cannot
	// be pushed down.
	correlatedColumns correlatedColumnDb
	// freshCorrelatedVarCounter is a simple counter such that we may generate fresh variable names in each
	// subquery expression.
	freshCorrelatedVarCounter int
}

func (v *pushdownVisitor) savePushdownStateForSubquery() (bool, []int, map[string][]string) {
	oldSelectIDsInScope := make([]int, len(v.selectIDsInScope))
	copy(oldSelectIDsInScope, v.selectIDsInScope)
	oldTableNamesInScope := make(map[string][]string, len(v.tableNamesInScope))
	for key, value := range v.tableNamesInScope {
		oldValue := make([]string, len(value))
		copy(oldValue, value)
		oldTableNamesInScope[key] = oldValue
	}
	return v.canPushdownCorrelated, oldSelectIDsInScope, oldTableNamesInScope
}

func (v *pushdownVisitor) restorePushdownStateForSubquery(oldCanPushdownCorrelated bool,
	oldSelectIDsInScope []int, oldTableNamesInScope map[string][]string) {
	v.canPushdownCorrelated = oldCanPushdownCorrelated
	v.selectIDsInScope = oldSelectIDsInScope
	v.tableNamesInScope = oldTableNamesInScope
}

func newPushdownVisitor(cfg *PushdownConfig) *pushdownVisitor {
	return &pushdownVisitor{
		cfg:                       cfg,
		logger:                    cfg.lg,
		depth:                     0,
		columnTracker:             newColumnTracker(),
		leftJoinOriginalNames:     make(dbData),
		pushdownFailures:          make(map[PlanStage][]PushdownFailure),
		canPushdownCorrelated:     false,
		correlatedColumns:         make(correlatedColumnDb),
		freshCorrelatedVarCounter: 0,
	}
}

// addCorrelatedColumnName adds correlated columns as we see them.
func (v *pushdownVisitor) addCorrelatedColumnName(dbName, tableName, columnName string) ast.Ref {
	// ensure that the correlatedColumns datastructure is initialized for this dbName and tableName.
	v.initCorrelatedColumns(dbName, tableName)

	if correlatedColumnName, ok := v.correlatedColumns[dbName][tableName][columnName]; ok {
		return ast.NewVariableRef(correlatedColumnName)
	}

	name := fmt.Sprintf("bic_correlated_var_%d", v.freshCorrelatedVarCounter)
	v.freshCorrelatedVarCounter++
	v.correlatedColumns[dbName][tableName][columnName] = name
	return ast.NewVariableRef(name)
}

func (v *pushdownVisitor) initCorrelatedColumns(dbName, tableName string) {
	if _, ok := v.correlatedColumns[dbName]; !ok {
		v.correlatedColumns[dbName] = make(correlatedColumnTable)
	}

	if _, ok := v.correlatedColumns[dbName][tableName]; !ok {
		v.correlatedColumns[dbName][tableName] = make(correlatedColumnName)
	}
}

func (v *pushdownVisitor) addNewPushdownFailure(ps PlanStage, name, msg string, meta ...string) {
	f := newPushdownFailure(name, msg, meta...)
	v.addPushdownFailure(ps, f)
}

func (v *pushdownVisitor) addPushdownFailure(ps PlanStage, f PushdownFailure) {
	failures := v.pushdownFailures[ps]
	v.pushdownFailures[ps] = append(failures, f)
}

func newTransitivePushdownFailure(name string) PushdownFailure {
	return newPushdownFailure(name, "unable to push down source stage")
}

func (v *pushdownVisitor) addTransitivePushdownFailure(ps PlanStage, name string) {
	v.addPushdownFailure(ps, newTransitivePushdownFailure(name))
}

func (v *pushdownVisitor) clearPushdownFailures(ps PlanStage) {
	v.pushdownFailures[ps] = nil
}

func (v *pushdownVisitor) addSelectIDsInScope(selectIDs ...int) {
	for _, selectID := range selectIDs {
		if !containsInt(v.selectIDsInScope, selectID) {
			v.selectIDsInScope = append(v.selectIDsInScope, selectID)
		}
	}
}

func (v *pushdownVisitor) addTableNamesInScope(databaseName string, tableNames ...string) {
	if v.tableNamesInScope == nil {
		v.tableNamesInScope = make(map[string][]string)
	}
	for _, tableName := range tableNames {
		if _, ok := v.tableNamesInScope[databaseName]; !ok {
			v.tableNamesInScope[databaseName] = []string{}
		}

		if !containsString(v.tableNamesInScope[databaseName], tableName) {
			v.tableNamesInScope[databaseName] = append(v.tableNamesInScope[databaseName], tableName)
		}
	}
}

// buildAddFieldsOrProject will build an addField if the server version is > 3.4.0, if it is less,
// it will build a project with everything not in the passed in body projected as 1, it will also
// skip any paths prefixed by a string in prefixesToSkip (mainly for avoiding conflicts).
func (v *pushdownVisitor) buildAddFieldsOrProject(body []*ast.AddFieldsItem, prefixesToSkip []string, mrs ...*mappingRegistry) ast.Stage {
	if procutil.VersionAtLeast(v.cfg.mongoDBVersion, []uint8{3, 4, 0}) {
		return ast.NewAddFieldsStage(body...)
	}
	// Make sure any prefix ends with '.'
	for i, prefix := range prefixesToSkip {
		if prefixesToSkip[len(prefixesToSkip)-1] != "." {
			prefixesToSkip[i] = prefix + "."
		}
	}

	// To keep track of the assigned fields in the body.
	projectedFields := map[string]struct{}{}

	// Copy the ast.AddFieldsItems into ast.ProjectItems.
	projectItems := make([]ast.ProjectItem, len(body))
	for i, afi := range body {
		projectItems[i] = ast.NewAssignProjectItem(afi.Name, afi.Expr)
		projectedFields[afi.Name] = struct{}{}
	}

	// We now need to make sure we project all the existing columns from all mapping registries.
	for _, mr := range mrs {
	TOP:
		for _, c := range mr.columns {
			fieldName, ok := mr.lookupFieldName(c.Database, c.Table, c.Name)
			if !ok {
				panic(fmt.Sprintf("cannot find referenced join column %#v in local lookup in"+
					" buildAddFieldsOrProject", c))
			}
			// Do not overwrite things already in the projectBody, and do not add paths
			// prefixed by our asField, because we will get conflicts.
			if _, ok := projectedFields[fieldName]; !ok {
				// Again, only keep if there isn't a prefix conflict.
				for _, prefix := range prefixesToSkip {
					if strings.HasPrefix(fieldName, prefix) {
						continue TOP
					}
				}

				projectItems = append(projectItems, ast.NewIncludeProjectItem(astutil.FieldRefFromFieldName(fieldName)))
				projectedFields[fieldName] = struct{}{}
			}
		}
	}

	return ast.NewProjectStage(projectItems...)
}

func (v *pushdownVisitor) visit(n Node) (Node, error) {
	originalN := n
	// first do some analysis.
	switch typedN := n.(type) {
	case *FilterStage:
		v.columnTracker.add(typedN.matcher)
	case *GroupByStage:
		for _, pc := range typedN.projectedColumns {
			v.columnTracker.add(pc.Expr)
		}
		for _, key := range typedN.keys {
			v.columnTracker.add(key)
		}
	case *JoinStage:
		if typedN.matcher != nil {
			v.columnTracker.add(typedN.matcher)
		}
	case *OrderByStage:
		for _, term := range typedN.terms {
			v.columnTracker.add(term.expr)
		}
	case *ProjectStage:
		for _, pc := range typedN.projectedColumns {
			v.columnTracker.add(pc.Expr)
		}
	}

	// second walk all child stages.
	var err error
	v.depth++
	switch n.(type) {
	case *SQLSubqueryExpr, SQLDoubleSubqueryExpr:
		// The various SQLSubqueryExpr's only apply to non-from clauses. This means that any new
		// selectIDs found inside a SQLSubqueryExpr are invalid outside of it. However, the
		// selectIDs outside of it are valid inside. This is the definition of a correlated
		// subquery. So, we'll save off the current selectIDs and restore them afterwards.

		oldSelectIDsInScope := v.selectIDsInScope
		oldTableNamesInScope := v.tableNamesInScope
		oldLeftJoinOriginalNames := v.leftJoinOriginalNames
		oldCanPushdownCorrelated := v.canPushdownCorrelated
		oldCorrelatedColumns := v.correlatedColumns

		v.canPushdownCorrelated = false
		n, err = walk(v, n)
		if err != nil {
			return nil, err
		}

		v.correlatedColumns = oldCorrelatedColumns
		v.canPushdownCorrelated = oldCanPushdownCorrelated
		v.leftJoinOriginalNames = oldLeftJoinOriginalNames
		v.tableNamesInScope = oldTableNamesInScope
		v.selectIDsInScope = oldSelectIDsInScope
	default:
		n, err = walk(v, n)
		if err != nil {
			return nil, err
		}
	}
	v.depth--

	var projSource PlanStage
	switch typedN := n.(type) {
	// Since we are walking to the bottom of the tree, we'll collect all
	// the selectIDs that are currently in scope. In the case of Joins,
	// this could be a combination of the below select ID sources.
	case *MongoSourceStage:
		v.addSelectIDsInScope(typedN.selectIDs...)
		v.addTableNamesInScope(typedN.dbName, typedN.aliasNames...)
	case *BSONSourceStage:
		v.addSelectIDsInScope(typedN.selectID)
	case *DynamicSourceStage:
		v.addSelectIDsInScope(typedN.selectID)
	// Pushdown
	case *FilterStage:
		n, err = v.visitFilter(typedN)
		if err != nil {
			return nil, fmt.Errorf("unable to pushdown filter: %v", err)
		}

		v.columnTracker.remove(typedN.matcher)

		if fs, ok := n.(*FilterStage); ok {
			v.columnTracker.add(fs.matcher)
			if ms, ok := fs.source.(*MongoSourceStage); ok {
				columnExprs := v.columnTracker.scopedColumnExprsForTables(v.selectIDsInScope, ms.dbName, ms.aliasNames)
				projSource, err = v.pushdownProject(columnExprs, fs.source)
				if err != nil {
					return nil, fmt.Errorf("unable to pushdown filter project: %v", err)
				}
				n = NewFilterStage(projSource, fs.matcher)
			}
			v.columnTracker.remove(fs.matcher)
		}

	case *GroupByStage:
		n, err = v.visitGroupBy(typedN)
		if err != nil {
			return nil, fmt.Errorf("unable to pushdown group by: %v", err)
		}

		if ms, ok := typedN.source.(*MongoSourceStage); ok && n == typedN {
			columnExprs := v.columnTracker.scopedColumnExprsForTables(
				v.selectIDsInScope, ms.dbName, ms.aliasNames)
			projSource, err = v.pushdownProject(columnExprs, typedN.source)
			if err != nil {
				return nil, fmt.Errorf("unable to pushdown group by project: %v", err)
			}
			n = NewGroupByStage(projSource, typedN.keys, typedN.projectedColumns)
		}

		for _, pc := range typedN.projectedColumns {
			v.columnTracker.remove(pc.Expr)
		}

		for _, key := range typedN.keys {
			v.columnTracker.remove(key)
		}

	case *JoinStage:
		n, err = v.visitJoin(typedN)
		if err != nil {
			return nil, fmt.Errorf("unable to pushdown join: %v", err)
		}

		if joinNode, joinOk := n.(*JoinStage); joinOk {
			left := joinNode.left
			right := joinNode.right

			// For inner joins, attempt to pushdown by flipping children.
			if joinNode.kind == InnerJoin {
				v.logger.Debugf(log.Dev, "attempting to pushdown inner join via flip")
				j := NewJoinStage(joinNode.kind, typedN.right, typedN.left, typedN.matcher)
				newJ, newErr := v.visitJoin(j)
				if newErr == nil {
					n = newJ
				}
			} else if joinNode.kind == RightJoin {
				// For right joins, attempt to pushdown using a left join.
				v.logger.Debugf(log.Dev, "attempting to pushdown right join via flip")
				j := NewJoinStage(LeftJoin, joinNode.right, typedN.left, typedN.matcher)
				newJ, newErr := v.visitJoin(j)
				if newErr == nil {
					n = newJ
				}
			}

			// attempt to pushdown by translating to expressive $lookup
			if stillAJoinNode, ok := n.(*JoinStage); ok {
				newJ, newErr := v.visitExpressiveJoin(stillAJoinNode)
				if newErr == nil {
					n = newJ
				}
			}

			if _, ok := n.(*JoinStage); ok {
				if ms, ok := left.(*MongoSourceStage); ok {
					columnExprs := v.columnTracker.scopedColumnExprsForTables(
						v.selectIDsInScope, ms.dbName, ms.aliasNames)
					left, err = v.pushdownProject(columnExprs, ms.clone())
					if err != nil {
						return nil, fmt.Errorf("unable to pushdown join.left project: %v", err)
					}
				}

				if ms, ok := right.(*MongoSourceStage); ok {
					columnExprs := v.columnTracker.scopedColumnExprsForTables(
						v.selectIDsInScope, ms.dbName, ms.aliasNames)

					right, err = v.pushdownProject(columnExprs, ms.clone())
					if err != nil {
						return nil, fmt.Errorf("unable to pushdown join.right project: %v", err)
					}
				}

				if left != typedN.left || right != typedN.right {
					n = NewJoinStage(typedN.kind, left, right, typedN.matcher)
				}
			}
		}

		if typedN.matcher != nil {
			v.columnTracker.remove(typedN.matcher)
		}

	case *LimitStage:
		n, err = v.visitLimit(typedN)
		if err != nil {
			return nil, fmt.Errorf("unable to pushdown limit: %v", err)
		}
	case *OrderByStage:
		n, err = v.visitOrderBy(typedN)
		if err != nil {
			return nil, fmt.Errorf("unable to pushdown order by: %v", err)
		}

		if ms, ok := typedN.source.(*MongoSourceStage); ok && n == typedN {
			columnExprs := v.columnTracker.scopedColumnExprsForTables(
				v.selectIDsInScope, ms.dbName, ms.aliasNames)
			projSource, pushdownProjectErr := v.pushdownProject(columnExprs, typedN.source)
			if pushdownProjectErr != nil {
				return nil, fmt.Errorf("unable to pushdown order by project: %v",
					pushdownProjectErr)
			}
			n = NewOrderByStage(projSource, typedN.terms...)
		}

		for _, term := range typedN.terms {
			v.columnTracker.remove(term.expr)
		}
	case *ProjectStage:
		n, err = v.visitProject(typedN)
		if err != nil {
			return nil, fmt.Errorf("unable to pushdown project: %v", err)
		}

		for _, pc := range typedN.projectedColumns {
			v.columnTracker.remove(pc.Expr)
		}

	case *SubquerySourceStage:
		oldSelectIDsInScope := v.selectIDsInScope
		oldTableNamesInScope := v.tableNamesInScope

		// Inside a SubquerySourceStage, there are no selectIDs or tableNames
		// in scope. However, after we are finished, the existing selectIDs
		// and tableNames are in scope as well as the additional selectID and
		// aliasName of the subquery.
		v.selectIDsInScope = []int{}
		v.tableNamesInScope = make(map[string][]string)

		n, err = v.visitSubquerySource(typedN)
		if err != nil {
			return nil, fmt.Errorf("unable to pushdown subquery source: %v", err)
		}

		v.selectIDsInScope = oldSelectIDsInScope
		v.tableNamesInScope = oldTableNamesInScope
		v.addSelectIDsInScope(typedN.selectID)
		v.addTableNamesInScope(typedN.dbName, typedN.aliasName)
	case *UnionStage:
		v.addNewPushdownFailure(typedN, unionStageName, "unions cannot be pushed down")
	}

	// Hack: if the node changed we need to remap the pushdownFailures,
	// since that is keyed on a pointer value. At some point we should
	// clean the pushdown visitor up to not produce new nodes, or
	// key this map off something other than pointer values.
	if p, ok := n.(PlanStage); ok && originalN != n {
		if _, ok := v.pushdownFailures[p]; !ok {
			v.pushdownFailures[p] = v.pushdownFailures[originalN.(PlanStage)]
		}
	}
	return n, nil
}

func (v *pushdownVisitor) canPushdown(ps PlanStage) (*MongoSourceStage, bool) {

	ms, ok := ps.(*MongoSourceStage)
	if !ok {
		return nil, false
	}

	return ms, true
}

const (
	projectPredicateFieldName = "__predicate"
)

func (v *pushdownVisitor) extractPreUnwindMatch(mr *mappingRegistry, expr SQLExpr, unwoundPath,
	unwoundIndexPath string) (*ast.MatchStage, bool) {
	parts := splitExpression(expr)

	var partsToMove []SQLExpr
	useElemMatch := true
	// Find any part that is composed solely of fields prefixed by the unwoundPath.
	for _, part := range parts {
		columns, err := referencedColumns(v.selectIDsInScope, part, true)
		if err != nil {
			return nil, false
		}
		valid := true
		for _, column := range columns {
			fieldName, ok := mr.lookupFieldName(column.Database, column.Table, column.Name)
			if !ok {
				return nil, false
			}

			if fieldName == unwoundPath {
				// This means that we are unwinding on an array of scalars. If this is the
				// case, we are not going to use $elemMatch because the $elemMatch language
				// for scalars is different and doesn't support everything that is possible
				// in SQL.
				useElemMatch = false
			} else if fieldName == unwoundIndexPath || !strings.HasPrefix(fieldName, unwoundPath+".") {
				valid = false
				break
			}
		}

		if valid {
			partsToMove = append(partsToMove, part)
		}
	}

	lookupFieldRef := mr.lookupFieldRef
	if useElemMatch {
		lookupFieldRef = func(databaseName, tableName, columnName string) (ast.Ref, bool) {
			// we are going to strip the prefix off of the fieldNames because $elemMatch syntax
			// is interesting. We know this won't fail because we've already done it for all
			// combinations.
			fieldName, _ := mr.lookupFieldName(databaseName, tableName, columnName)
			return astutil.FieldRefFromFieldName(strings.TrimPrefix(fieldName, unwoundPath+".")), true
		}
	}

	t := newInternalPushdownTranslator(v.cfg, lookupFieldRef, v)

	combined := combineExpressions(partsToMove)

	// We don't care about the remaining. We will still be placing a match after the unwind,
	// so anything we can't do here gets handled there anyways.
	matchBody, _, _ := t.TranslatePredicate(combined)
	if matchBody == nil {
		// Nothing to do.
		return nil, false
	}

	// We cannot put $expr inside $elemMatch
	if _, ok := matchBody.(*ast.AggExpr); ok {
		return nil, false
	}

	if useElemMatch {
		matchBody = ast.NewDocument(
			ast.NewDocumentElement(unwoundPath, ast.NewDocument(
				ast.NewDocumentElement("$elemMatch", matchBody),
			)),
		)
	}

	return ast.NewMatchStage(matchBody), true
}

// nolint: unparam
func (v *pushdownVisitor) visitFilter(filter *FilterStage) (PlanStage, error) {

	ms, ok := v.canPushdown(filter.source)
	if !ok {
		v.addTransitivePushdownFailure(filter, filterStageName)
		return filter, nil
	}

	pipeline := ast.NewPipeline(ms.pipeline.Stages...)
	var t *PushdownTranslator
	var localMatcher SQLExpr
	t = newInternalPushdownTranslator(v.cfg, ms.mappingRegistry.lookupFieldRef, v)

	if valueExpr, ok := filter.matcher.(SQLValueExpr); ok {
		value := valueExpr.Value
		// Our pushed down expression has left us with just a value,
		// we can see if it matches right now. If so, we eliminate
		// the filter from the tree. Otherwise, we return an
		// operator that yields no rows.
		if !values.Bool(value) {
			return &EmptyStage{filter.Columns(), filter.Collation()}, nil
		}

		// Otherwise, the filter simply gets removed from the tree.

	} else {
		if len(pipeline.Stages) == 1 {
			if unwind, isUnwind := pipeline.Stages[0].(*ast.UnwindStage); isUnwind {
				// Before pushing down the match, if the current pipeline contains
				// an $unwind as the first stage in the pipeline, try to place any criteria
				// for the unwound array before the $unwind using an $elemMatch. These will
				// need to still stay after the $unwind as well, but this should cut down on
				// the number of documents passing through the $unwind clause while also allowing
				// the use of an index.
				// NOTE: putting a match between a lookup and an unwind causes a server optimization
				// to get skipped.
				v.logger.Debugf(log.Dev, "attempting to add a redundant match before unwind")

				path := astutil.RefString(unwind.Path)
				indexPath := unwind.IncludeArrayIndex

				if preUnwindMatch, ok := v.extractPreUnwindMatch(ms.mappingRegistry, filter.matcher,
					path[1:], indexPath); ok {
					pipeline.Stages = append([]ast.Stage{preUnwindMatch}, pipeline.Stages...)
				}
			}
		}

		var matchBody ast.Expr
		matchBody, localMatcher, _ = t.TranslatePredicate(filter.matcher)
		if matchBody != nil {
			pipeline.Stages = append(pipeline.Stages, t.subqueryLookupStages...)
			pipeline.Stages = append(pipeline.Stages, ast.NewMatchStage(matchBody))
		}

		if localMatcher != nil {
			// We have a predicate that completely or partially couldn't be handled by $match.
			// Attempt to push it down as part of a $project/$match combination.
			if predicate, err := t.TranslateExpr(localMatcher); err == nil {

				// MySQL's version of truthiness is different than MongoDB's. We need to modify
				// the predicate to account for this difference. It looks, effectively, like this:
				predicateRef := ast.NewVariableRef("predicate")
				predicate = ast.NewLet(
					[]*ast.LetVariable{ast.NewLetVariable("predicate", predicate)},
					astutil.WrapInOp(bsonutil.OpAnd,
						ast.NewBinary(bsonutil.OpNeq, predicateRef, astutil.FalseLiteral),
						ast.NewBinary(bsonutil.OpNeq, predicateRef, astutil.ZeroInt32Literal),
						ast.NewBinary(bsonutil.OpNeq, predicateRef, astutil.StringValue("0")),
						ast.NewBinary(bsonutil.OpNeq, predicateRef, astutil.StringValue("-0")),
						ast.NewBinary(bsonutil.OpNeq, predicateRef, astutil.StringValue("0.0")),
						ast.NewBinary(bsonutil.OpNeq, predicateRef, astutil.StringValue("-0.0")),
						ast.NewBinary(bsonutil.OpNeq, predicateRef, astutil.NullLiteral),
					),
				)

				fieldName := v.uniqueFieldName(projectPredicateFieldName, ms.mappingRegistry)
				stageBody := []*ast.AddFieldsItem{
					ast.NewAddFieldsItem(fieldName, predicate),
				}

				predicateEvaluationStage := v.buildAddFieldsOrProject(stageBody, []string{}, ms.mappingRegistry)
				pipeline.Stages = append(pipeline.Stages, t.subqueryLookupStages...)
				pipeline.Stages = append(pipeline.Stages,
					predicateEvaluationStage,
					ast.NewMatchStage(
						ast.NewBinary(bsonutil.OpEq,
							astutil.FieldRefFromFieldName(fieldName),
							astutil.TrueLiteral,
						),
					),
				)

				localMatcher = nil
			}

			if matchBody == nil && localMatcher != nil {
				// No pieces of the matcher are able to be pushed down,
				// so there is no change in the operator tree.
				v.logger.Debugf(log.Dev, "cannot push down filter expression: \n%v", filter.matcher.String())
				return filter, nil
			}
		}
	}

	// If we end up here, it's because we have messed with the pipeline
	// in the current table scan operator, so we need to reconstruct the
	// operator nodes.
	ms = ms.clone().(*MongoSourceStage)
	ms.pipeline = pipeline

	if localMatcher != nil {
		// We ended up here because we have a predicate
		// that can be partially pushed down, so we construct
		// a new filter with only the part remaining that
		// cannot be pushed down.
		return NewFilterStage(ms, localMatcher), nil
	}

	// Everything was able to be pushed down, so the filter
	// is removed from the plan.
	return ms, nil
}

const (
	groupID             = mongoPrimaryKey
	groupDistinctPrefix = "distinct "
	groupTempTable      = ""
)

// visitGroupBy works by using a visitor to systematically visit and replace certain SQLExpr
// while generating $group and $project stages for the aggregation pipeline.
// nolint: unparam
func (v *pushdownVisitor) visitGroupBy(gb *GroupByStage) (PlanStage, error) {

	ms, ok := v.canPushdown(gb.source)
	if !ok {
		v.addTransitivePushdownFailure(gb, groupByStageName)
		return gb, nil
	}

	pipeline := ast.NewPipeline(ms.pipeline.Stages...)

	// 1. Translate keys.
	keys, keyNameMapping, subqueryCmpStages, err := v.translateGroupByKeys(gb.keys, ms.mappingRegistry.lookupFieldRef)
	if err != nil {
		v.logger.Warnf(log.Dev, "cannot translate group by keys: %v", err)
		v.addPushdownFailure(gb, err)
		return gb, nil
	}

	pipeline.Stages = append(pipeline.Stages, subqueryCmpStages...)

	// 2. Translate aggregations.
	result, subqueryCmpStages, err := v.translateGroupByAggregates(keyNameMapping, gb.projectedColumns, ms.mappingRegistry.lookupFieldRef)
	if err != nil {
		v.logger.Warnf(log.Dev, "cannot translate group by aggregates: %v", err)
		v.addPushdownFailure(gb, err)
		return gb, nil
	}

	pipeline.Stages = append(pipeline.Stages, subqueryCmpStages...)
	pipeline.Stages = append(pipeline.Stages, ast.NewGroupStage(keys, result.groupItems...))

	// 3. Translate the final project if necessary.
	var mr *mappingRegistry
	if result.requiresTwoSteps {
		project, err := v.translateGroupByProject(result.mappedProjectedColumns, result.mappingRegistry.lookupFieldRef)
		if err != nil {
			v.logger.Warnf(log.Dev, "cannot translate group by project: %v", err)
			v.addPushdownFailure(gb, err)
			return gb, nil
		}
		pipeline.Stages = append(pipeline.Stages, project)

		// 4. Fix up the MongoSourceStage - None of the current registrations in mappingRegistry
		// are valid any longer. We need to clear them out and re-register the new columns.
		mr = newMappingRegistry()
		for _, mappedProjectedColumn := range result.mappedProjectedColumns {
			// at this point, our project has all the stringified names of the select expressions,
			// so we need to re-map them each column to its new MongoDB name. This process is what
			// makes the push-down transparent to subsequent operators in the tree that either
			// haven't yet been pushed down, or cannot be. Either way, the output of a push-down
			// must be exactly the same as the output of a non-pushed-down group.
			if mr.registerMapping(
				mappedProjectedColumn.projectedColumn.Database,
				mappedProjectedColumn.projectedColumn.Table,
				mappedProjectedColumn.projectedColumn.Name,
				sanitizeFieldName(mappedProjectedColumn.projectedColumn.String()),
				false,
			) {
				mr.addColumn(mappedProjectedColumn.projectedColumn.Column)
			}
		}
	} else {
		mr = newMappingRegistry()
		for _, mpc := range result.mappedProjectedColumns {
			// At this point, we pushed down a group, but we still need to map the projected column
			// name to the expressions that were pushed down. We know that all the pushed down exprs
			// are now columns, so we simply look up the original field name and use that.
			columnExpr, ok := mpc.expr.(SQLColumnExpr)
			if !ok {
				v.logger.Warnf(log.Dev, "expr was not a column, expr was %v", mpc.expr)
				v.addNewPushdownFailure(gb, groupByStageName, "pushed-down expr not a column")
				return gb, nil
			}
			fieldName, ok := result.mappingRegistry.lookupFieldName(
				columnExpr.databaseName,
				columnExpr.tableName,
				columnExpr.columnName,
			)
			if !ok {
				v.logger.Warnf(log.Dev, "could not find translated aggregate's field name")
				v.addNewPushdownFailure(gb, groupByStageName, "could not find translated aggregate's field name")
				return gb, nil
			}
			if mr.registerMapping(
				mpc.projectedColumn.Database,
				mpc.projectedColumn.Table,
				mpc.projectedColumn.Name,
				fieldName,
				false,
			) {
				mr.addColumn(mpc.projectedColumn.Column)
			}
		}
	}

	ms = ms.clone().(*MongoSourceStage)
	ms.pipeline = pipeline
	ms.mappingRegistry = mr

	return ms, nil
}

// translateGroupByKeys takes the key expressions and builds an _id document. All keys, even single
// keys, will be nested underneath the '_id' field. Each field's name will be "group_key_i" where
// "i" is the index in the keys slice. The mapping from stringified SQLExpr to key name will be
// returned in addition to the _id document so the names can be looked up as needed.
// For example, 'select a, b from foo group by c' will build an id document that looks like this:
//      _id: { group_key_0: "$c" }
// and a key name mapping that looks like this:
//     map[string]string{"foo_DOT_c": "group_key_0"}
//
// Likewise, multiple columns will build something similar.
// For example, 'select a, b from foo group by c,d' will build an id document that looks like this:
//      _id: { group_key_0: "$c", group_key_1: "$d" }
// and the appropriate mapping.
//
// Finally, anything other than a column will still build similarly.
// For example, 'select a, b from foo group by c+d' will build an id document that looks like this:
//      _id: { group_key_0: { $add: ["$c", "$d"] } }
// with the expected mapping:
//      map[string]string{"foo_DOT_c+foo_DOT_d": "group_key_0"}
//
// All projected names are the fully qualified name from SQL, ignoring the mongodb name except for
// when referencing the underlying field.
func (v *pushdownVisitor) translateGroupByKeys(keys []SQLExpr, lookupFieldRef FieldRefLookup) (*ast.Document, map[SQLExpr]string, []ast.Stage, PushdownFailure) {
	t := newInternalPushdownTranslator(v.cfg, lookupFieldRef, v)

	keyDocumentElements := make([]*ast.DocumentElement, len(keys))
	keyNameMapping := make(map[SQLExpr]string)

	for i, key := range keys {
		translatedKey, err := t.TranslateExpr(key)
		if err != nil {
			return nil, nil, nil, err
		}

		uniqueKeyName := fmt.Sprintf("group_key_%d", i)
		keyDocumentElements[i] = ast.NewDocumentElement(uniqueKeyName, translatedKey)
		keyNameMapping[key] = uniqueKeyName
	}

	return ast.NewDocument(keyDocumentElements...), keyNameMapping, t.subqueryLookupStages, nil
}

// translateGroupByAggregatesResult is just a holder for the results from the
// translateGroupByAggregates function.
type translateGroupByAggregatesResult struct {
	groupItems             []*ast.GroupItem
	mappedProjectedColumns []*mappedProjectedColumn
	mappingRegistry        *mappingRegistry
	requiresTwoSteps       bool
}

type mappedProjectedColumn struct {
	projectedColumn ProjectedColumn
	expr            SQLExpr
}

// translateGroupByAggregates takes the key expressions and the select expressions and builds a
// $group stage. It does this by employing a visitor that walks each of the select expressions
// individually and, depending on the type of expression, builds a full solution or a partial
// solution to accomplishing the goal. For example, the query 'select sum(a) from foo' can be fully
// realized with a single $group, whereas 'select sum(distinct a) from foo' requires a $group which
// adds 'a' to a set ($addToSet) and a subsequent $project which then does a $sum on the set created
// in the $group. Currently, we always build the $project whether it's necessary or not.
//
// In addition to generating the group, new expressions and a mapping registry are also returned
// that account for the namings and resulting expressions from the $group. For example,
//
// 'select sum(a) from foo'
//
// will take in a SQLAggFunctionExpr for the 'sum(a)' and return a SQLFieldExpr that represents the
// already
// summed 'a' field such that the subsequent $project doesn't need to care about the aggregation.
// In the same way, sum(distinct a) will take in a SQLAggFunctionExpr which refers to the column 'a'
// and return a new SQLAggFunctionExpr which refers to the newly created $addToSet field called
// 'distinct foo_DOT_a'. This way, the subsequent $project now has the correct reference to the
// field name in the $group.
func (v *pushdownVisitor) translateGroupByAggregates(keys map[SQLExpr]string, projectedColumns ProjectedColumns,
	lookupFieldRef FieldRefLookup) (*translateGroupByAggregatesResult, []ast.Stage, PushdownFailure) {

	// This represents all the expressions that should be passed on to $project such that
	// translateGroupByProject is able to do its work without redoing a bunch of the conditionals
	// and special casing here.
	mappedProjectedColumns := []*mappedProjectedColumn{}

	// translator will "accumulate" all the group fields. Below, we iterate over each select
	// expressions, which account for all the fields that need to be present in the $group.
	translator := &groupByAggregateTranslator{
		cfg:               v.cfg,
		groupItems:        []*ast.GroupItem{},
		groupKeyNames:     keys,
		lookupFieldRef:    lookupFieldRef,
		mappingRegistry:   newMappingRegistry(),
		requiresTwoSteps:  false,
		logger:            v.logger,
		correlatedColumns: v.correlatedColumns,
		pushdownVisitor:   v,
		pushdownTranslator: newInternalPushdownTranslator(v.cfg, lookupFieldRef,
			v),
	}

	for _, projectedColumn := range projectedColumns {

		newExpr, err := translator.visit(projectedColumn.Expr)
		if err != nil {
			if pdf, ok := err.(PushdownFailure); ok {
				return nil, nil, pdf
			}
			panic(fmt.Errorf("encountered fatal error while translating aggregates in %v: %v",
				groupByStageName, err.Error()))
		}

		mappedProjectedColumn := &mappedProjectedColumn{
			expr:            newExpr.(SQLExpr),
			projectedColumn: projectedColumn,
		}

		mappedProjectedColumns = append(mappedProjectedColumns, mappedProjectedColumn)
	}

	return &translateGroupByAggregatesResult{translator.groupItems, mappedProjectedColumns,
		translator.mappingRegistry, translator.requiresTwoSteps}, translator.pushdownTranslator.subqueryLookupStages, nil
}

type groupByAggregateTranslator struct {
	cfg                *PushdownConfig
	groupItems         []*ast.GroupItem
	groupKeyNames      map[SQLExpr]string // a map from SQLExpr to group key names used in the _id document.
	lookupFieldRef     FieldRefLookup
	mappingRegistry    *mappingRegistry
	requiresTwoSteps   bool
	logger             log.Logger
	correlatedColumns  correlatedColumnDb
	pushdownVisitor    *pushdownVisitor
	pushdownTranslator *PushdownTranslator
}

const (
	sumAggregateCountSuffix = "_count"
)

// Visit recursively visits each expression in the tree, adds the relevant $group entries, and
// returns an expression that can be used to build a subsequent $project.
func (v *groupByAggregateTranslator) visit(n Node) (Node, error) {
	switch typedN := n.(type) {
	case SQLColumnExpr:
		ref, ok := v.lookupFieldRef(typedN.databaseName, typedN.tableName, typedN.columnName)
		fieldRef, isFieldRef := ref.(*ast.FieldRef)
		if !ok || !isFieldRef {
			return nil, fmt.Errorf("could not map %v.%v to a field", typedN.tableName,
				typedN.columnName)
		}
		if keyName, ok := v.groupKeyNames[typedN]; !ok {
			// Since it's not an aggregation function, this implies that it takes the first value of
			// the column. So project the field, and register the mapping.
			v.groupItems = append(v.groupItems, ast.NewGroupItem(
				sanitizeFieldName(typedN.String()),
				ast.NewFunction(bsonutil.OpFirst, getProjectedFieldName(astutil.RefString(fieldRef), typedN.EvalType())),
			))
			v.mappingRegistry.registerMapping(typedN.databaseName, typedN.tableName, typedN.columnName, sanitizeFieldName(typedN.String()), false)
		} else {
			// The _id is added to the $group in translateGroupByKeys. We will only be here if the
			// user has also projected the group key, in which we'll need this to look it up in
			// translateGroupByProject under its name. Hence, all we need to do is register the
			// mapping.
			v.mappingRegistry.registerMapping(typedN.databaseName, typedN.tableName, typedN.columnName, groupID+"."+keyName, false)
		}
		return typedN, nil
	case SQLAggFunctionExpr:

		dbName := getDatabaseName(typedN)

		var newExpr SQLExpr
		groupConcat, isGroupConcat := typedN.(*SQLGroupConcatFunctionExpr)
		if typedN.Distinct() || isGroupConcat {
			// Distinct aggregation expressions are two-step aggregations. In the $group stage, we
			// use $addToSet to handle whatever the distinct expression is, which could simply be a
			// field name, or something more complex like a mathematical computation. We don't care
			// either way, and TranslateExpr handles generating the correct thing. Once this is
			// done, we create a new SQLAggFunctionExpr whose argument maps to the newly named field
			// containing the set of values to perform the aggregation on.

			// Group_concat aggregation expressions are always two-step aggregations, regardless of
			// whether they are distinct. In the $group stage, we construct the list of entries to
			// the result string. In the $project stage, we concatenate these entries together.

			// $reduce was introduced in Mongo 3.4, so we cannot push down the query if
			// the user is using an earlier Mongo version.
			if isGroupConcat && !procutil.VersionAtLeast(v.cfg.mongoDBVersion, []uint8{3, 4, 0}) {
				return nil, newPushdownFailure(typedN.Name(), "cannot push down group_concat for versions < 3.4")
			}

			v.requiresTwoSteps = true

			fieldName := ""
			operator := bsonutil.OpPush
			if typedN.Distinct() {
				fieldName = groupDistinctPrefix
				operator = bsonutil.OpAddToSet
			}
			fieldName = fieldName + sanitizeFieldName(SQLExprs(typedN.Exprs()).String())

			var trans ast.Expr
			var pushdownFail PushdownFailure
			if isGroupConcat {
				var translatedExprs []ast.Expr
				for _, expr := range typedN.Exprs() {
					trans, pushdownFail = v.pushdownTranslator.TranslateExpr(expr)
					if pushdownFail != nil {
						return nil, pushdownFail
					}

					translatedExprs = append(translatedExprs, trans)
				}

				// concatenatedArguments holds a concatenated string of any number of expressions
				// that are referenced within this group. At the end of the loop, it holds each
				// combined expression that will subsequently be concatenated with a separator.
				// For example, in the query `select group_concat(a, b separator ",")`,
				// concatenatedArguments will be `$concat: ["$a", "$b"]`
				var concatenatedArguments ast.Expr
				if len(translatedExprs) == 1 {
					concatenatedArguments = translatedExprs[0]
				} else {
					concatenatedArguments = astutil.WrapInConcat(translatedExprs)
				}

				v.groupItems = append(v.groupItems, ast.NewGroupItem(fieldName, ast.NewFunction(operator, concatenatedArguments)))
			} else {
				trans, pushdownFail = v.pushdownTranslator.TranslateExpr(typedN.Exprs()[0])
				if pushdownFail != nil {
					return nil, pushdownFail
				}

				v.groupItems = append(v.groupItems, ast.NewGroupItem(fieldName, ast.NewFunction(operator, trans)))
			}

			exprs := []SQLExpr{
				NewSQLColumnExpr(
					0,
					dbName,
					groupTempTable,
					fieldName,
					typedN.EvalType(),
					schema.MongoNone,
					false,
					true,
				),
			}
			newExpr = NewSQLAggregationFunctionExpr(typedN.Name(), false, exprs)
			if isGroupConcat {
				newGroupConcat := newExpr.(*SQLGroupConcatFunctionExpr)
				newGroupConcat.Separator = groupConcat.Separator
				newGroupConcat.GroupConcatMaxLen = groupConcat.GroupConcatMaxLen
			}

			v.mappingRegistry.registerMapping(dbName, groupTempTable, fieldName, fieldName, false)
		} else {
			// Non-distinct aggregation functions are one-step aggregations that can be performed
			// completely by the $group. Here, we simply build the correct aggregation function for
			// $group and create a new expression that references the resulting field. There isn't a
			// need to keep the aggregation function around because the aggregation has already been
			// done and all that's left is a field for $project to reference.

			// Count is special since MongoDB doesn't have a native count function. Instead, we use
			// $sum. However, depending on the form that count takes, we need to different things.
			// 'count(*)' is just {$sum: 1}, but 'count(a)' requires that we not count null,
			// undefined, and missing fields. Hence, it becomes a $sum with $cond and $ifNull.

			var trans ast.Expr
			var pushdownFail PushdownFailure
			_, isCount := typedN.(*SQLCountFunctionExpr)
			if isCount {
				if typedN.Exprs()[0].String() == "*" {
					trans = ast.NewFunction(bsonutil.OpSum, astutil.OneInt32Literal)
				} else {
					trans, pushdownFail = v.pushdownTranslator.TranslateExpr(typedN.Exprs()[0])
					if pushdownFail != nil {
						return nil, pushdownFail
					}

					trans = getCountAggregation(trans)
				}
			} else {
				trans, pushdownFail = v.pushdownTranslator.TranslateExpr(typedN)
				if pushdownFail != nil {
					return nil, pushdownFail
				}
			}

			fieldName := sanitizeFieldName(typedN.String())
			v.groupItems = append(v.groupItems, ast.NewGroupItem(fieldName, trans))
			v.mappingRegistry.registerMapping(dbName, groupTempTable, fieldName, fieldName, false)

			_, isSum := typedN.(*SQLSumFunctionExpr)
			if isSum {
				// Summing a column with all nulls should result in a null sum. However, MongoDB
				// returns 0. So, we'll add in an arbitrary count operator to count the number
				// of non-nulls and, in the following $project, we'll check this to know whether
				// or not to use the sum or to use null.
				v.requiresTwoSteps = true
				countTrans, pushdownFail := v.pushdownTranslator.TranslateExpr(typedN.Exprs()[0])
				if pushdownFail != nil {
					return nil, pushdownFail
				}
				countFieldName := sanitizeFieldName(typedN.String() + sumAggregateCountSuffix)
				v.groupItems = append(v.groupItems, ast.NewGroupItem(countFieldName, getCountAggregation(countTrans)))
				v.mappingRegistry.registerMapping(dbName, groupTempTable, countFieldName, countFieldName, false)

				newExpr = NewSQLCaseExpr(
					NewSQLValueExpr(values.NewSQLNull(v.pushdownTranslator.valueKind())),
					newCaseCondition(
						NewSQLColumnExpr(0, dbName, groupTempTable, countFieldName,
							types.EvalInt64, schema.MongoNone, false, true),
						NewSQLColumnExpr(0, dbName, groupTempTable, fieldName,
							typedN.EvalType(), schema.MongoNone, false, true),
					),
				)
			} else {
				newExpr = NewSQLColumnExpr(0, dbName, groupTempTable, fieldName,
					typedN.EvalType(), schema.MongoNone, false, true)
			}
		}

		return newExpr, nil

	case SQLExpr:
		if keyName, ok := v.groupKeyNames[typedN]; ok {
			// The _id is added to the $group in translateGroupByKeys. We will only be here if the
			// user has also projected the group key, in which we'll need this to look it up in
			// translateGroupByProject under its name. In this, we need to create a new expr that is
			// simply a field pointing at the nested identifier and register that mapping.
			fieldName := sanitizeFieldName(typedN.String())
			dbName := getDatabaseName(typedN)
			newExpr := NewSQLColumnExpr(0, dbName, groupTempTable, fieldName,
				typedN.EvalType(), schema.MongoNone, false, true)
			v.mappingRegistry.registerMapping(dbName, groupTempTable, fieldName, groupID+"."+keyName, false)
			return newExpr, nil
		}

		// We'll only get here for two-step expressions. An example is
		// 'select a + b from foo group by a' or 'select b + sum(c) from foo group by a'. In this
		// case, we'll descend into the tree recursively which will build up the $group for the
		// necessary pieces. Finally, return the now changed expression such that $project can act
		// on them appropriately.
		v.requiresTwoSteps = true
		newN, err := walk(v, n)
		if err != nil {
			panic(err)
		}
		return newN, nil
	default:
		// PlanStages will end up here and we don't need to do anything in them.
		return n, nil
	}
}

// getCountAggregation is used when a non-star count is used. {sum:1} isn't valid because null,
// undefined, and missing values should not be part of result. Because MongoDB doesn't have a
// proper count operator, we have to
// do some extra checks along the way.
func getCountAggregation(expr ast.Expr) *ast.Function {
	return ast.NewFunction(bsonutil.OpSum,
		astutil.WrapInNullCheckedCond(astutil.ZeroInt32Literal, astutil.OneInt32Literal, expr))
}

// translateGroupByProject takes the expressions and builds a $project. All the hard work was done
// in translateGroupByAggregates, so this is simply a process of either adding a field to the
// $project, or completing two-step aggregations. Two-step aggregations that need completing are
// expressions like 'sum(distinct a)' or 'a + b' where b was part of the group key.
func (v *pushdownVisitor) translateGroupByProject(mappedProjectedColumns []*mappedProjectedColumn, lookupFieldRef FieldRefLookup) (*ast.ProjectStage, PushdownFailure) {
	t := newInternalPushdownTranslator(v.cfg, lookupFieldRef, v)

	projectItems := make([]ast.ProjectItem, len(mappedProjectedColumns)+1)
	projectItems[0] = ast.NewExcludeProjectItem(ast.NewFieldRef(groupID, nil))
	for i, mappedProjectedColumn := range mappedProjectedColumns {

		mappedName := sanitizeFieldName(mappedProjectedColumn.projectedColumn.String())
		switch typedE := mappedProjectedColumn.expr.(type) {
		case SQLColumnExpr:
			// Any one-step aggregations will end up here as they were fully performed in the
			// $group. So, simple column references ('select a') and simple aggregations:
			// ('select sum(a)').
			ref, ok := lookupFieldRef(typedE.databaseName, typedE.tableName, typedE.columnName)
			if !ok {
				return nil, newPushdownFailure(
					groupByStageName,
					"unable to get field name for column",
					"tableName", typedE.tableName,
					"columnName", typedE.columnName,
				)
			}

			projectItems[i+1] = ast.NewAssignProjectItem(mappedName, ref)
		default:
			// Any two-step aggregations will end up here to complete the second step.
			trans, err := t.TranslateExpr(mappedProjectedColumn.expr)
			if err != nil {
				return nil, err
			}

			projectItems[i+1] = ast.NewAssignProjectItem(mappedName, trans)
		}
	}

	return ast.NewProjectStage(projectItems...), nil
}

const (
	joinedFieldNamePrefix    = "__joined_"
	leftJoinExcludeFieldName = "__leftjoin_exclude"
)

func (v *pushdownVisitor) getFixedLookupFieldRef(
	combinedMappingRegistry *mappingRegistry,
	db, tbl, col, asField, foreignIndex string,
	preserveIndex bool,
) (ast.Ref, bool) {
	registries := []*mappingRegistry{combinedMappingRegistry}

	// Join predicates should always be based on the original field, rather than the added
	// fields that have been added for left joins. The only way the value being NULL could
	// matter is if <=> or is NULL is in the predicate, and even if that is the case,
	// it would already be NULL from being a left join, anyway. If it is instead <> NULL
	// the predicate is essentially a no-op
	fieldName, _, _, ok := lookupSQLColumnForJoin(db, tbl, col, registries, v.leftJoinOriginalNames)
	if !ok {
		panic(fmt.Sprintf("could not find column: %q.%q, "+
			"this should never happen, registries were: %v", tbl, col, registries))
	}
	if fieldName == foreignIndex {
		logPrefix := "$lookup translation"
		// preserveIndex is used when we are doing self-join optimization,
		// it is false for $lookup, so we can use that to set out log message
		if preserveIndex {
			logPrefix = "self-join optimization"
		}
		v.logger.Debugf(log.Dev, logPrefix+": cannot use foreign unwind index: %q in left "+
			"join criteria because use occurs before foreign unwind moving on...", foreignIndex)

		return nil, false
	}

	// Inside a $filter and $map (which use the result of this function), columns with the
	// asField prefix will have their prefix renamed. As such, we need to intercept this call
	// and handle that translation early. For instance, if the asField is $b.child and the
	// field ends up as $b.child.myField, then the result will be $$this.myField.
	// NOTE: it is important to use asField + "." as the prefix, because otherwise we will
	// end up renaming something we generate in unwinds like c_idx to this_idx... which is wrong
	// We then also need the condition where fieldName == asField, since prefix will no longer
	// catch it, since we have added the "."
	if strings.HasPrefix(fieldName, asField+".") {
		ref := astutil.FieldRefFromFieldNameWithParent(strings.TrimPrefix(fieldName, asField+"."), ast.NewVariableRef("this"))
		return ref, true
	}

	if fieldName == asField {
		return ast.NewVariableRef("this"), true
	}

	return astutil.FieldRefFromFieldName(fieldName), true
}

// buildRemainingPredicateForLeftJoin will return 2 items; first a $project to
// put before the unwind, and a $match to put after the unwind. The remaining
// predicate SQLExpr is used to build the $project (or $addFields) and the
// $match. asField is the name of the array field to check. foreignIndex is
// the name of the foreign pipeline unwind, if any (passing an empty string is
// safe if there is no foreign unwind), because we cannot build a predicate
// using the foreignIndex because it creates a circular dependency in the
// pipeline (the foreign unwind must go afther the $project/$addFields which
// must use the field created by the foreign unwind)
func (v *pushdownVisitor) buildRemainingPredicateForLeftJoin(
	combinedMappingRegistry *mappingRegistry,
	remainingPredicate SQLExpr,
	asField, foreignIndex string,
	preserveIndex bool,
) (ast.Stage, *ast.MatchStage, PushdownFailure) {
	fixedLookupFieldName := func(db, tbl, col string) (ast.Ref, bool) {
		return v.getFixedLookupFieldRef(combinedMappingRegistry, db, tbl, col, asField, foreignIndex, preserveIndex)
	}
	t := newInternalPushdownTranslator(v.cfg, fixedLookupFieldName, v)

	cond, err := t.TranslateExpr(remainingPredicate)
	if err != nil {
		v.logger.Warnf(log.Dev, "cannot translate remaining left join predicate %#v", remainingPredicate)
		return nil, nil, err
	}

	projectBody := make([]*ast.AddFieldsItem, 1)
	var match *ast.MatchStage
	dolAsField := astutil.FieldRefFromFieldName(asField)
	if preserveIndex {
		// This is interesting. First, we are going to create variable that marks every item in the
		// array that should be excluded. Using that variable, we'll then create a condition. If we
		// filter all the items out that should be excluded and end up with 0 items, we set the
		// field to an empty array. Otherwise, we keep the array with the marked items and use a
		// match after the unwind to get rid of the rows that don't belong. The reason we have to do
		// this is because, even when no items from the "right" side of a join match, we still need
		// to include the left side one time. However, we can't just eliminate the non-matching
		// array items now (using $filter) because we need to maintain the array index of the items
		// that do match.
		mappedRef := ast.NewVariableRef("mapped")
		projectBody[0] = ast.NewAddFieldsItem(asField, ast.NewLet(
			[]*ast.LetVariable{
				ast.NewLetVariable("mapped", astutil.WrapInMap(
					astutil.WrapInCond(
						dolAsField,

						// It is very important that we map null and missing to
						// [] rather than [null] because [null] is semantically
						// different:
						// When we form a child table with
						// {..., x : [null], ...}
						// we have one row with one primary key x_idx = 0 with
						// null as a value. When we form a child table with
						// [], null, or missing,
						// we produce 0 rows. Mapping null (or missing) to
						// [null] breaks
						// this semantics, and ruins the fields added for
						// self-join optimized left-joins
						astutil.WrapInNullCheckedCond(
							ast.NewArray(),
							ast.NewArray(dolAsField),
							dolAsField,
						),
						ast.NewFunction(bsonutil.OpIsArray, dolAsField),
					),
					"this",
					astutil.WrapInCond(
						astutil.ThisVarRef,
						ast.NewDocument(ast.NewDocumentElement(leftJoinExcludeFieldName, astutil.BooleanConstant(true))),
						cond,
					),
				)),
			},
			astutil.WrapInCond(
				mappedRef,
				ast.NewArray(),
				ast.NewBinary(bsonutil.OpGt,
					ast.NewFunction(bsonutil.OpSize,
						astutil.WrapInFilter(
							mappedRef, "this",
							ast.NewBinary(bsonutil.OpNeq,
								ast.NewVariableRef("this."+leftJoinExcludeFieldName),
								astutil.TrueLiteral,
							),
						),
					),
					astutil.ZeroInt32Literal,
				),
			),
		))

		match = ast.NewMatchStage(
			ast.NewBinary(bsonutil.OpNeq,
				ast.NewFieldRef(leftJoinExcludeFieldName, astutil.FieldRefFromFieldName(asField)),
				astutil.TrueLiteral,
			),
		)
	} else {
		// In this case, we can simply filter the array because we don't care about preserving the
		// index. If the predicate doesn't pass, then we set the 'as' field to nil.
		projectBody[0] = ast.NewAddFieldsItem(asField,
			astutil.WrapInFilter(dolAsField, "this", cond),
		)

	}
	projection := v.buildAddFieldsOrProject(projectBody, []string{asField}, combinedMappingRegistry)
	return projection, match, nil
}

func (v *pushdownVisitor) optimizeSelfJoinTables(msLocal, msForeign *MongoSourceStage, join *JoinStage) (PlanStage, error) {
	var foreignRegistryBackup *mappingRegistry
	// If we fail to translate a left join predicate later, we will need to restore this
	// if, instead this is an inner join, there is nothing to worry about
	if join.kind == LeftJoin {
		foreignRegistryBackup = msForeign.mappingRegistry.copy()
	}

	newPipeline, err := v.optimizeSelfJoinPipeline(msLocal, msForeign, join.kind)
	if err != nil {
		v.logger.Warnf(log.Dev, "cannot self-join optimize pipelines: %v", err)
		return nil, nil
	}

	ms := msLocal.clone().(*MongoSourceStage)
	ms.selectIDs = append(ms.selectIDs, msForeign.selectIDs...)
	ms.aliasNames = append(ms.aliasNames, msForeign.aliasNames...)
	ms.tableNames = append(ms.tableNames, msForeign.tableNames...)
	ms.collectionNames = append(ms.collectionNames,
		msForeign.collectionNames...)
	for key, val := range msForeign.isShardedCollection {
		msLocal.isShardedCollection[key] = val
	}

	newMappingRegistry := ms.mappingRegistry.copy()
	newMappingRegistry.columns = append(newMappingRegistry.columns,
		msForeign.mappingRegistry.columns...)
	if msForeign.mappingRegistry.fields != nil {
		for database, tables := range msForeign.mappingRegistry.fields {
			for tableName, columns := range tables {
				for columnName, fieldName := range columns {
					newMappingRegistry.registerMapping(database, tableName, columnName, fieldName, false)
				}
			}
		}
	}

	// Do not copy back the newMappingRegistry and newPipeline.stages until
	// we are sure that we can correctly translate the remaining join
	// predicate, because it is still possible that there will need to be a
	// deoptimization back to $lookup or in memory join. Unfortunately,
	// checking the conditions that can cause failure here earlier would
	// be more expensive.

	remainingPredicate := combineExpressions(
		v.remainingJoinPredicate(msLocal, msForeign, join.matcher))

	if remainingPredicate != nil {
		if join.kind == InnerJoin || join.kind == StraightJoin {
			// This isn't a left join, so we do not have to worry about
			// failing to build the left-join predicate and can copy
			// back the newMappingRegistry and newPipeline.stages
			ms.mappingRegistry, ms.pipeline = newMappingRegistry, ast.NewPipeline(newPipeline.stages...)
			v.logger.Debugf(log.Dev, "self-join optimization: creating filter "+
				"stage for remaining predicate: %v",
				remainingPredicate.String())
			f, err := v.visit(NewFilterStage(ms, remainingPredicate))
			if err != nil {
				return nil, err
			}
			return f.(PlanStage), nil
		}

		// This "predicate" must get inserted before the addFields from
		// the right side. The predicate should be based on the first unwind
		// from the right side, the insertion point should be immediately after the
		// last local unwind, this ensures it is put before the addFields introduced
		// by left join self-optimization.
		localUnwinds, totalUnwinds := astutil.GetPipelineUnwindFields(msLocal.pipeline.Stages),
			astutil.GetPipelineUnwindFields(newPipeline.stages)
		unwindSuffix, _ := astutil.GetUnwindSuffix(localUnwinds, totalUnwinds)
		insertionPoint := 0
		if len(localUnwinds) != 0 {
			insertionPointPath := localUnwinds[len(localUnwinds)-1].Path
			insertionPointUnwind, ok := astutil.FindUnwindForPath(totalUnwinds, insertionPointPath)
			if !ok {
				panic(fmt.Sprintf("could not find unwind for path %v in pipeline %v, "+
					"this should never happen)",
					insertionPointPath, newPipeline.stages))
			}
			insertionPoint = insertionPointUnwind.StageNumber + 1
		}

		project, match, err := v.buildRemainingPredicateForLeftJoin(
			newMappingRegistry,
			remainingPredicate,
			strings.Replace(unwindSuffix[0].Path, "$", "", 1),
			unwindSuffix[0].Index,
			true,
		)
		if err != nil {
			// We failed to translate, make sure to restore the foreign
			// mapping registry
			msForeign.mappingRegistry = foreignRegistryBackup
			return join, nil
		}

		if match != nil {
			newPipeline.stages = append(newPipeline.stages, match)
		}

		// Insert project after the first.
		newPipeline.stages = astutil.InsertPipelineStageAt(newPipeline.stages, project, insertionPoint)
	}

	ms.mappingRegistry, ms.pipeline = newMappingRegistry, ast.NewPipeline(newPipeline.stages...)
	return ms, nil
}

// SimplifyFalseJoinCriterion will check join for a null or false criterion and return
// a replacement plan stage which avoids contacting MongoDB when no rows are required
// for one or both sides of the join.
func (v *pushdownVisitor) simplifyFalseJoinCriterion(join *JoinStage) PlanStage {

	// It is sufficient to check if there is a single false or null criterion since
	// partial evaluation is complete.
	crit, ok := join.matcher.(SQLValueExpr)
	if !(ok && (values.IsFalsy(crit.Value) || values.HasNullValue(crit.Value))) {
		return nil
	}

	if join.kind == LeftJoin {
		// We have to be able to push down the sources if this is a left join.
		msLocal, ok := join.left.(*MongoSourceStage)
		if !ok {
			return nil
		}

		msForeign, ok := join.right.(*MongoSourceStage)
		if !ok {
			return nil
		}

		// Field names are needed for projection of the fields we're leaving behind in the
		// right side of this join. Here is a disambiguated prefix.
		prefix := v.uniqueFieldName(
			sanitizeFieldName(joinedFieldNamePrefix+msForeign.aliasNames[0]),
			msLocal.mappingRegistry,
		)

		// We need to update the mapping resgistry to contain the additional fields.
		newMappingRegistry := msLocal.mappingRegistry.merge(msForeign.mappingRegistry, prefix)

		// Finally, if this is a left outer join we can eliminate all rows from the right.
		v.logger.Debugf(log.Dev, "successfully translated left join stage on false/null "+
			"criterion to left table access")
		msLocal.mappingRegistry = newMappingRegistry
		return join.left
	}

	// If this is an inner join we can eliminate all rows.
	v.logger.Debugf(log.Dev, "successfully translated join stage on false/null criterion "+
		"to empty stage")
	return NewEmptyStage(join.Columns(), join.Collation())
}

func (v *pushdownVisitor) visitJoin(join *JoinStage) (PlanStage, error) {
	v.logger.Debugf(log.Dev, "attempting to translate join stage")

	if join.matcher == nil {
		v.logger.Warnf(log.Dev, "cannot push down join stage, matcher is nil")
		v.addNewPushdownFailure(join, joinStageName, "matcher is nil")
		return join, nil
	}

	// Ensure we are able to pushdown this kind of join.
	if failure := v.canPushdownJoinKind(join.kind); failure != nil {
		v.addPushdownFailure(join, failure)
		return join, nil
	}

	// Attempt to simplify joins that have falsy constant join criteria.
	if replace := v.simplifyFalseJoinCriterion(join); replace != nil {
		return replace, nil
	}

	// Make sure the local table is a MongoSourceStage.
	msLocal, ok := join.left.(*MongoSourceStage)
	if !ok {
		v.addTransitivePushdownFailure(join, joinStageName)
		return join, nil
	}

	// Make sure the foreign table is a MongoSourceStage.
	msForeign, ok := join.right.(*MongoSourceStage)
	if !ok {
		v.addTransitivePushdownFailure(join, joinStageName)
		return join, nil
	}

	// We can't push down cross-database joins.
	if msLocal.dbName != msForeign.dbName {
		v.logger.Warnf(log.Dev,
			"cannot pushdown join stage, local database is different from foreign database")
		v.addNewPushdownFailure(join, joinStageName, "local and foreign databases are different")
		return join, nil
	}

	// See if we can push down self-joins without using $lookup.
	optimizedJoinPlanStage, err := v.attemptToOptimizeSelfJoins(join, msLocal, msForeign)
	if err != nil {
		return nil, err
	}
	if optimizedJoinPlanStage != nil {
		return optimizedJoinPlanStage, nil
	}

	// If a foreign table is sharded, we can't push down the join to a $lookup.
	for i, collection := range msForeign.collectionNames {
		var isSharded bool
		isSharded, ok = msForeign.isShardedCollection[collection]
		if !ok {
			// If this happens, there is a serious programming error.
			panic(fmt.Errorf("could not determine whether collection %q is sharded", collection))
		}
		if isSharded {
			v.logger.Warnf(log.Dev,
				"unable to translate join stage to lookup: foreign table %q is sharded",
				msForeign.tableNames[i])
			v.addNewPushdownFailure(join, joinStageName, "foreign table's collection is sharded")
			return join, nil
		}
	}

	// When the foreign source has a pipeline, ensure we can push it down.
	if len(msForeign.pipeline.Stages) > 0 {
		if ok := v.canPushdownForeignJoinSourcePipeline(join, msLocal, msForeign); !ok {
			return join, nil
		}
	}

	// Find the local column and the foreign column.
	lookupInfo, err := getLocalAndForeignColumns(msLocal, msForeign, join.matcher)
	if err != nil {
		v.logger.Warnf(log.Dev, "unable to translate join stage to lookup: %v", err)
		v.addNewPushdownFailure(
			join, joinStageName,
			"unable to get local and foreign columns",
			"error", err.Error(),
		)
		return join, nil
	}

	// Prevent join pushdown when UUID subtype 3 encoding is different.
	if failure := v.doesJoinHaveIncompatibleUUIDs(lookupInfo); failure != nil {
		v.addPushdownFailure(join, failure)
		return join, nil
	}

	pipeline, failure := v.buildJoinPushdownPipeline(join, lookupInfo)
	if failure != nil {
		v.addPushdownFailure(join, failure)
		return join, nil
	}

	if lookupInfo.remainingPredicate != nil && (join.kind == InnerJoin || join.kind == StraightJoin) {
		f, err := v.visit(NewFilterStage(pipeline, lookupInfo.remainingPredicate))
		if err != nil {
			return nil, err
		}

		return f.(PlanStage), nil
	}

	v.logger.Debugf(log.Dev, "successfully translated join stage to lookup")

	return pipeline, nil
}

// nolint: unparam
func (v *pushdownVisitor) visitExpressiveJoin(join *JoinStage) (PlanStage, error) {

	// cannot use expressive lookup before 3.6
	if !procutil.VersionAtLeast(v.cfg.mongoDBVersion, []uint8{3, 6, 0}) {
		v.logger.Warnf(log.Dev, "cannot push down join stage to expressive lookup: expressive lookup not available")
		v.addNewPushdownFailure(join, joinStageName, "cannot push down expressive $lookup to MongoDB < 3.6")
		return join, nil
	}

	v.logger.Debugf(log.Dev, "attempting to translate join stage to expressive lookup")
	v.clearPushdownFailures(join)

	// the join type must be usable. MongoDB can only do an inner join and a left outer join (as
	// well as a cross join, which we represent as an inner join with no matcher, and a right
	// outer join, which we flip to represent as a left join).
	var localSource, foreignSource PlanStage
	kind := join.kind

	switch kind {
	case InnerJoin, CrossJoin, LeftJoin, StraightJoin:
		localSource = join.left
		foreignSource = join.right
	case RightJoin:
		v.logger.Infof(log.Admin, "flipping right join and optimizing as left join")
		localSource = join.right
		foreignSource = join.left
		kind = LeftJoin
	default:
		v.logger.Warnf(log.Dev, "cannot push down %v", kind)
		v.addNewPushdownFailure(join, joinStageName, "join kind is not inner, cross, left, right, or straight")
		return join, nil
	}

	// we have to be able to push both tables down
	msLocal, ok := localSource.(*MongoSourceStage)
	if !ok {
		v.addTransitivePushdownFailure(join, joinStageName)
		return join, nil
	}
	msForeign, ok := foreignSource.(*MongoSourceStage)
	if !ok {
		v.addTransitivePushdownFailure(join, joinStageName)
		return join, nil
	}

	// the tables must both belong to the same MongoDB database
	if msLocal.dbName != msForeign.dbName {
		if !msForeign.IsDual() {
			v.addNewPushdownFailure(join, joinStageName, "local database is different from foreign database")
			return join, nil
		}
		// If msForeign is in fact a DUAL, then we should be able to just "fix" the
		// database/collection choice we arbitrarily made when creating it to match
		// the local database, and still JOIN successfully.
		msForeign.dbName = msLocal.dbName
		msForeign.collectionNames = msLocal.collectionNames
		msForeign.isShardedCollection = msLocal.isShardedCollection
	}

	// The foreign table must not be sharded.
	for i, collection := range msForeign.collectionNames {
		var isSharded bool
		isSharded, ok = msForeign.isShardedCollection[collection]
		if !ok {
			// if this happens, there is a serious programming error
			panic(fmt.Errorf("could not determine whether collection %q is sharded", collection))
		}
		if isSharded {
			v.logger.Warnf(log.Dev, "unable to translate join stage to expressive lookup: right table %q is sharded", msForeign.tableNames[i])
			v.addNewPushdownFailure(join, joinStageName, "right table's collection is sharded")
			return join, nil
		}
	}

	// defaults for expressive lookup mappings/pipeline
	localMappings := make([]*ast.LookupLetItem, 0)
	matchPipeline := make([]ast.Stage, 0)

	if join.matcher != nil {
		// find the local columns used in the join matcher
		localCols, err := getTableColumnsInExpr(msLocal, join.matcher)
		if err != nil {
			v.logger.Warnf(log.Dev, "unable to translate join stage to expressive lookup: %v", err)
			v.addNewPushdownFailure(
				join, joinStageName,
				"unable to get local columns",
				"error", err.Error(),
				"local_mongo_source", fmt.Sprintf("%+v", msLocal),
			)
			return join, nil
		}

		// find the foreign columns used in the join matcher
		foreignCols, err := getTableColumnsInExpr(msForeign, join.matcher)
		if err != nil {
			v.logger.Warnf(log.Dev, "unable to translate join stage to expressive lookup: %v", err)
			v.addNewPushdownFailure(
				join, joinStageName,
				"unable to get foreign columns",
				"error", err.Error(),
				"foreign_mongo_source", fmt.Sprintf("%+v", msForeign),
			)
			return join, nil
		}

		// do not push down if different uuid types are used in the join predicate
		// this is to match visitJoin's behavior (not pushing down comparisons of
		// differently typed UUIDs). Once BI-1447 is addressed, this block should
		// be removed, since TranslateExpr will properly handle UUID pushdown (or
		// lack thereof).
		var uuidType schema.MongoType
		for _, col := range append(localCols, foreignCols...) {
			mongoType := col.columnType.MongoType
			if values.IsUUID(col.columnType.MongoType) {
				if uuidType != "" && uuidType != mongoType {
					v.logger.Warnf(log.Dev, "unable to translate join "+
						"stage to expressive lookup: join criteria uses"+
						"more than one UUID encoding")
					v.addNewPushdownFailure(
						join, joinStageName,
						"join tables use different uuid subtype 3 encodings",
						"first_uuid_type", string(uuidType),
						"second_uuid_type", string(mongoType),
					)
					return join, nil
				}
				uuidType = mongoType
			}
		}

		// build the mapping of local variables to pipeline variables
		foreignPipelineRegistry := msForeign.mappingRegistry.copy()
		for _, col := range localCols {
			var field string
			field, ok = msLocal.mappingRegistry.lookupFieldName(
				col.databaseName,
				col.tableName,
				col.columnName,
			)
			if !ok {
				v.logger.Warnf(log.Dev, "cannot find referenced foreign join column %#v in expressive lookup", col)
				v.addNewPushdownFailure(
					join, joinStageName,
					"cannot find foreign column",
					"column", col.String(),
				)
				return join, nil
			}

			sanitized := sanitizeFieldName(field)
			newField := v.uniqueFieldName("local_table__"+sanitized, foreignPipelineRegistry)
			newField = v.uniqueLetVarName(newField)
			localMappings = append(localMappings, ast.NewLookupLetItem(newField, astutil.FieldRefFromFieldName(field)))

			foreignPipelineRegistry.registerMapping(col.databaseName, col.tableName, col.columnName, newField, true)
		}

		// create the pushdown translator
		t := newInternalPushdownTranslator(v.cfg, foreignPipelineRegistry.lookupFieldRef, v)
		matchPipeline = append([]ast.Stage{}, msForeign.pipeline.Stages...)

		// When the join matcher is the bool `true`, like in a cross join,
		// we do not need to add an additional match pipeline.
		matcherValExpr, ok := join.matcher.(SQLValueExpr)
		if !ok || matcherValExpr.Value.Value() != true {
			// Build the foreign pipeline.
			translated, _, pf := t.TranslatePredicate(join.matcher)
			if pf != nil {
				v.logger.Warnf(log.Dev, "unable to translate join criteria: %v", join.matcher)
				v.addPushdownFailure(join, pf)
				return join, nil
			}

			matchPipeline = append(matchPipeline, t.subqueryLookupStages...)
			matchPipeline = append(matchPipeline, ast.NewMatchStage(translated))
		}
	}

	pipeline := ast.NewPipeline(msLocal.pipeline.Stages...)

	// create and append the lookup and unwind stages to the pipeline
	asField := v.uniqueFieldName(
		sanitizeFieldName(joinedFieldNamePrefix+msForeign.aliasNames[0]),
		msLocal.mappingRegistry,
	)
	lookup := ast.NewLookupStage(
		msForeign.collectionNames[0], nil, "", asField,
		localMappings,
		ast.NewPipeline(matchPipeline...),
	)

	unwind := ast.NewUnwindStage(astutil.FieldRefFromFieldName(asField), "", kind == LeftJoin)

	pipeline.Stages = append(pipeline.Stages, lookup, unwind)

	// create the new MongoSourceStage that makes up the newly joined table
	ms := msLocal.clone().(*MongoSourceStage)
	ms.selectIDs = append(ms.selectIDs, msForeign.selectIDs...)
	ms.aliasNames = append(ms.aliasNames, msForeign.aliasNames...)
	ms.tableNames = append(ms.tableNames, msForeign.tableNames...)
	ms.collectionNames = append(ms.collectionNames,
		msForeign.collectionNames...)
	for key, val := range msForeign.isShardedCollection {
		msLocal.isShardedCollection[key] = val
	}

	// create the new mappingRegistry that makes up the newly joined table
	newMappingRegistry := ms.mappingRegistry.copy()
	newMappingRegistry.columns = append(newMappingRegistry.columns,
		msForeign.mappingRegistry.columns...)
	if msForeign.mappingRegistry.fields != nil {
		for database, tables := range msForeign.mappingRegistry.fields {
			for tableName, columns := range tables {
				for columnName, fieldName := range columns {
					newMappingRegistry.registerMapping(database, tableName, columnName, asField+"."+fieldName, false)
				}
			}
		}
	}

	ms.pipeline = pipeline
	ms.mappingRegistry = newMappingRegistry
	ms.LimitRowCount = int(math.Max(-1.0, float64(msLocal.LimitRowCount*msForeign.LimitRowCount)))

	v.logger.Debugf(log.Dev, "successfully translated join stage to expressive lookup")

	return ms, nil
}

type lookupInfo struct {
	localColumn        *SQLColumnExpr
	foreignColumn      *SQLColumnExpr
	remainingPredicate SQLExpr
}

// getLocalAndForeignColumns takes the local and foreign tables and predicate
// of a join, and returns a column from each table on whose equality the tables
// can be joined, plus the remainder of the join predicate left over after
// removing the equality condition on the two returned columns.
func getLocalAndForeignColumns(localTable, foreignTable *MongoSourceStage, e SQLExpr) (
	*lookupInfo, error) {

	// flatten and split the expression tree on AND exprs,
	// returning a list of the conjunctive expressions
	exprs := splitExpression(e)

	var errMsg string

	// find a SQLEqualsExpr in the list of split exprs
	for i, expr := range exprs {
		errMsg = ""
		if equalExpr, ok := expr.(*SQLComparisonExpr); ok && equalExpr.op == EQ {
			// the left and right sides of this SQLEqualsExpr must be columns
			if column1, ok := equalExpr.left.(SQLColumnExpr); ok {
				if column2, ok := equalExpr.right.(SQLColumnExpr); ok {

					// we must have one column each from the local and foreign tables
					var localColumn, foreignColumn *SQLColumnExpr

					if containsString(localTable.aliasNames, column1.tableName) {
						localColumn = &column1
						errMsg = fmt.Sprintf("%s not within local tables %q", equalExpr.left,
							localTable.aliasNames)
					} else if containsString(foreignTable.aliasNames, column1.tableName) {
						foreignColumn = &column1
						errMsg = fmt.Sprintf("%s not within foreign tables %q", equalExpr.left,
							foreignTable.aliasNames)
					}

					if containsString(localTable.aliasNames, column2.tableName) {
						localColumn = &column2
						errMsg = fmt.Sprintf("%s not within local tables %q", equalExpr.right,
							localTable.aliasNames)
					} else if containsString(foreignTable.aliasNames, column2.tableName) {
						foreignColumn = &column2
						errMsg = fmt.Sprintf("%s not within foreign tables %q", equalExpr.right,
							foreignTable.aliasNames)
					}

					// if we have one column from each table being joined, return
					// these two columns along with the AND of the remaining exprs
					if localColumn != nil && foreignColumn != nil {
						combined := combineExpressions(append(exprs[:i], exprs[i+1:]...))
						return &lookupInfo{localColumn, foreignColumn, combined}, nil
					}
				}
				if errMsg == "" {
					errMsg = fmt.Sprintf("%s is not a sql column (%T)", equalExpr.right,
						equalExpr.right)
				}
			}
			if errMsg == "" {
				errMsg = fmt.Sprintf("%s is not a sql column (%T)", equalExpr.left, equalExpr.left)
			}
		}
	}

	if errMsg == "" {
		errMsg = "no column equality comparison found"
	}

	return nil, fmt.Errorf("join criteria cannot be pushed down '%v': %s", e, errMsg)
}

// lookupSQLColumnForJoin looks up the _original_ field name for a given
// table.column in a join constraint. This needs to take into account any
// renames that have occurred due to self join optimized left joins. The reason
// we do this is that we are using the original field name to denote the
// semantic identity of columns for the purposes of PK equality matching
// constraints as we need to identify two columns as being semantically
// isomorphic if they have been aliased at the SQL level.
func lookupSQLColumnForJoin(databaseName, tableName, columnName string,
	mappingRegistries []*mappingRegistry, leftJoinOriginalNames dbData) (string, string, int, bool) {
	var renamedField string
	var ok bool
	if renamedField, ok = leftJoinOriginalNames[databaseName][tableName][columnName]; !ok {
		renamedField = ""
	}
	for i, registry := range mappingRegistries {
		fieldName, ok := registry.lookupFieldName(databaseName, tableName, columnName)
		if ok {
			if renamedField == "" {
				renamedField = fieldName
			}
			return renamedField, fieldName, i, true
		}
	}
	return "", "", 0, false
}

type consolidatedPipeline struct {
	stages           []ast.Stage
	arrayPaths       []string
	arrayPathIndexes []string
}

func getProjectedFieldNames(project *ast.ProjectStage) map[string]struct{} {
	names := map[string]struct{}{}

	for _, item := range project.Items {
		switch t := item.(type) {
		case *ast.AssignProjectItem:
			names[t.Name] = struct{}{}
		case *ast.IncludeProjectItem:
			names[astutil.RefString(t.FieldRef)] = struct{}{}
		case *ast.ExcludeProjectItem:
			names[astutil.RefString(t.FieldRef)] = struct{}{}
		}
	}

	return names
}

func (v *pushdownVisitor) optimizeSelfJoinPipeline(local, foreign *MongoSourceStage,
	kind JoinKind) (*consolidatedPipeline, error) {
	pipeline := &consolidatedPipeline{}

	augmentProjection := func(project *ast.ProjectStage) (*ast.ProjectStage, error) {
		projectedFieldNames := getProjectedFieldNames(project)

		prefixPathPresent := func(fieldName string) bool {
			names := strings.Split(fieldName, ".")
			for i := 0; i < len(names); i++ {
				if _, ok := projectedFieldNames[sanitizeFieldName(strings.Join(names[:i], "."))]; ok {
					return true
				}
			}
			return false
		}

		for _, c := range foreign.mappingRegistry.columns {
			fieldName, ok := foreign.mappingRegistry.lookupFieldName(c.Database, c.Table, c.Name)
			if !ok {
				return nil, fmt.Errorf("cannot find referenced foreign column %v.%v.%v in "+
					"projection lookup", c.Database, c.Table, c.Name)
			}

			if _, ok := projectedFieldNames[fieldName]; !ok && !prefixPathPresent(fieldName) {
				v.logger.Debugf(log.Dev, "augmenting local table with column '%v.%v'.'%v'",
					c.Database, c.Table, c.Name)
				project.Items = append(project.Items, ast.NewIncludeProjectItem(astutil.FieldRefFromFieldName(fieldName)))
				projectedFieldNames[fieldName] = struct{}{}
				foreign.mappingRegistry.registerMapping(c.Database, c.Table, c.Name, fieldName, false)
			}
		}

		return ast.NewProjectStage(project.Items...), nil
	}

	getPathsAndPipeline := func(stages []ast.Stage, isLocal bool) error {
		for _, stage := range stages {
			unwind, isUnwind := stage.(*ast.UnwindStage)
			if !isUnwind {
				if isLocal {
					// For projections, ensure all foreign columns are included.
					if projectStage, isProject := stage.(*ast.ProjectStage); isProject {
						project, err := augmentProjection(projectStage)
						if err != nil {
							return err
						}
						pipeline.stages = append(pipeline.stages, project)
					} else {
						pipeline.stages = append(pipeline.stages, stage)
					}
					continue
				} else {
					return fmt.Errorf("found stage in foreign table (%v) pipeline: %#v that "+
						"cannot be self-join optimized", foreign.aliasNames, stage)
				}
			}

			path := astutil.RefString(unwind.Path)
			arrayIdx := unwind.IncludeArrayIndex

			if !strutil.StringSliceContains(pipeline.arrayPathIndexes, arrayIdx) {
				pipeline.arrayPaths = append(pipeline.arrayPaths, path)
				pipeline.arrayPathIndexes = append(pipeline.arrayPathIndexes, arrayIdx)
				if kind == LeftJoin && !isLocal {
					unwind.PreserveNullAndEmptyArrays = true
				}
				pipeline.stages = append(pipeline.stages, ast.NewUnwindStage(unwind.Path, arrayIdx, unwind.PreserveNullAndEmptyArrays))
			}

		}
		return nil
	}

	if err := getPathsAndPipeline(local.pipeline.Stages, true); err != nil {
		return nil, err
	}

	if err := getPathsAndPipeline(foreign.pipeline.Stages, false); err != nil {
		return nil, err
	}

	if kind == LeftJoin {
		localUnwindFields := astutil.GetPipelineUnwindFields(local.pipeline.Stages)
		foreignUnwindFields := astutil.GetPipelineUnwindFields(foreign.pipeline.Stages)
		totalUnwindFields := astutil.GetPipelineUnwindFields(pipeline.stages)
		// If the local has more unwinds than the foreign, this is equivalent to an
		// inner join, just return the optimized pipeline.
		if len(localUnwindFields) > len(foreignUnwindFields) {
			return pipeline, nil
		}
		unwindSuffix, ok := astutil.GetUnwindSuffix(localUnwindFields, foreignUnwindFields)

		// It is safe to to allow left joins with non-progeny as long as
		// the foreign pipeline only has 1 unwind.
		if !ok && len(foreignUnwindFields) > 1 {
			panic("unwind prefixes do not match, this should have disallowed self-join " +
				"optimization. This should never happen.")
		}

		// So this is interesting. Get the suffix of unwinds that
		// don't match in the local and foreign sides If we get to this
		// point, it must mean that the suffix has size one, and that
		// it comes from the foreign pipeline, but we do this because
		// we need the actual stage position in the pipeline, as there
		// could be any number of non-unwind stages in the pipeline.
		// The indexes of this suffix will be those indexes we do not
		// wish to remap in our $addFields stage. As a simple case, if
		// we are joining the parent document to an array field, the
		// local table will have 0 unwinds and the foreign table will
		// have 1, and we will insert before that 1 unwind and we will
		// not want to remap the PK resulting from that unwind, while
		// we do want to remap the PK from the document (the _id).

		// This is a degenerate case where they do not have a shared
		// prefix but we allow it because foreign unwind depth is 1.
		// If this happens, we just use foreignUnwindFields as
		// unwindSuffix.
		if len(unwindSuffix) == 0 {
			unwindSuffix = foreignUnwindFields
		}

		// If there's no unwindSuffix from the foreign pipeline then we can't optimize here.
		if len(unwindSuffix) == 0 {
			return pipeline, nil
		}

		unwindSuffixIndexes := astutil.GetIndexes(unwindSuffix)
		unwindSuffixPaths := astutil.GetPaths(unwindSuffix)

		// Insertion point should be *after* the first unwind in the
		// unwindSuffix If it is inserted before, it will not always
		// work, and when it does work it is due to luck, not correct
		// semantics, but the StageNumber in the unwindSuffix may not
		// be the same as the StageNumber for that $unwind in the new
		// self-join optimized pipeline, which is what we actually
		// need.
		insertionPointPath := unwindSuffix[0].Path
		insertionPointUnwind, ok := astutil.FindUnwindForPath(totalUnwindFields,
			insertionPointPath)
		if !ok {
			panic(fmt.Sprintf("could not find unwind for path %v in pipeline %v, "+
				"this should never happen)",
				insertionPointPath, pipeline.stages))
		}
		insertionPoint := insertionPointUnwind.StageNumber

		addFieldsBody := make([]*ast.AddFieldsItem, 0)
		for databaseName, tables := range foreign.mappingRegistry.fields {
			_, ok := v.leftJoinOriginalNames[databaseName]
			if !ok {
				v.leftJoinOriginalNames[databaseName] = make(tableData)
			}

			for tableName, fields := range tables {
				leftJoinOriginalNames, ok := v.leftJoinOriginalNames[databaseName][tableName]
				if !ok {
					leftJoinOriginalNames = make(columnData)
					v.leftJoinOriginalNames[databaseName][tableName] = leftJoinOriginalNames
				}
				for tableCol, docCol := range fields {

					// We only want to add fields for primary keys
					// that are above the $addFields stage in the
					// pipeline as these are the only PKs that can
					// be NULL in a left join that won't be made
					// NULL by the unwind itself (and addingFields
					// for them can actually result in getting
					// NON-NULL values that *should be* NULL.
					if pathStartsWithAny(unwindSuffixPaths, docCol) ||
						strutil.StringSliceContains(unwindSuffixIndexes, docCol) {
						continue
					}
					uniqueDocCol := v.uniqueFieldName(docCol, local.mappingRegistry,
						foreign.mappingRegistry)

					// Keep track of renamed (remapped) doc fields
					// for purposes of constructing the remaining
					// Join constraints.
					leftJoinOriginalNames[tableCol] = docCol

					// Now, actually rename the doc fields in the
					// mappingRegistry to ensure that the final
					// projection will be correct
					fields[tableCol] = uniqueDocCol

					// Add to the addFieldsBody, one bson.DocElem
					// will be added for each PK that we need to
					// remap. Only PKs need be remapped as any
					// other values will end up NULL by
					// construction, where they need be NULL.
					addFieldsBody = append(addFieldsBody, ast.NewAddFieldsItem(
						uniqueDocCol,
						buildLeftSelfJoinPKAddFieldBody(
							astutil.FieldRefFromFieldName(unwindSuffix[0].Path),
							astutil.FieldRefFromFieldName(docCol),
						),
					))
				}
			}
		}
		addFields := v.buildAddFieldsOrProject(addFieldsBody, []string{}, local.mappingRegistry,
			foreign.mappingRegistry)
		pipeline.stages = astutil.InsertPipelineStageAt(pipeline.stages, addFields, insertionPoint)
	}

	return pipeline, nil
}

// buildLeftSelfJoinPKAddFieldBody builds the conditional for an AddField,
// checking column columnCheck for missing, NULL, or empty.
func buildLeftSelfJoinPKAddFieldBody(columnCheck, columnCopy *ast.FieldRef) ast.Expr {
	return astutil.WrapInCond(
		astutil.NullLiteral,
		columnCopy,
		astutil.WrapInNullCheck(columnCheck),
		ast.NewBinary(bsonutil.OpEq, columnCheck, ast.NewArray()),
	)
}

func createEmptyResultsPipeline(mongoDBVersion []uint8) []ast.Stage {

	// $collStats is used when possible because it is more efficient than doing limit:1
	// in the case of views. This is because it avoids going through the view's pipeline.
	emptyPipeline := []ast.Stage{
		ast.NewCollStatsStage(nil, nil, nil),

		// $match will return false, causing this pipeline to return no documents.
		ast.NewMatchStage(
			ast.NewBinary(bsonutil.OpEq,
				ast.NewFieldRef(falsyPredicateField, nil),
				astutil.Int32Value(2),
			),
		),
	}

	// $collStats is not available in 3.2, so use limit:1 to get at most one document.
	if !procutil.VersionAtLeast(mongoDBVersion, []uint8{3, 4, 0}) {
		emptyPipeline[0] = ast.NewLimitStage(1)
	}

	return emptyPipeline
}

func (v *pushdownVisitor) visitLimit(limit *LimitStage) (PlanStage, error) {

	ms, ok := v.canPushdown(limit.source)
	if !ok {
		v.addTransitivePushdownFailure(limit, limitStageName)
		return limit, nil
	}

	pipeline := ast.NewPipeline(ms.pipeline.Stages...)

	if limit.offset > 0 {
		if limit.offset > math.MaxInt64 {
			return nil, fmt.Errorf("limit with offset '%d' cannot be pushed down", limit.offset)
		}
		pipeline.Stages = append(pipeline.Stages, ast.NewSkipStage(int64(limit.offset)))
	}

	if limit.limit > 0 {
		if limit.limit > math.MaxInt64 {
			return nil, fmt.Errorf("limit with rowcount '%d' cannot be pushed down", limit.limit)
		}
		pipeline.Stages = append(pipeline.Stages, ast.NewLimitStage(int64(limit.limit)))
	}

	// If limit is zero, swap out for empty pipeline.
	if limit.limit == 0 {

		ms = ms.clone().(*MongoSourceStage)
		emptyPipeline := createEmptyResultsPipeline(v.cfg.mongoDBVersion)

		ms.pipeline = ast.NewPipeline(emptyPipeline...)
		ms.LimitRowCount = 0
		return ms, nil
	}

	ms = ms.clone().(*MongoSourceStage)
	ms.pipeline = pipeline
	ms.LimitRowCount = int(limit.limit)
	return ms, nil
}

// nolint: unparam
func (v *pushdownVisitor) visitOrderBy(orderBy *OrderByStage) (PlanStage, error) {

	ms, ok := v.canPushdown(orderBy.source)
	if !ok {
		v.addTransitivePushdownFailure(orderBy, orderByStageName)
		return orderBy, nil
	}

	sortItems := make([]*ast.SortItem, len(orderBy.terms))
	newFields := map[string]ast.Expr{}
	var t *PushdownTranslator

	for i, term := range orderBy.terms {

		var databaseName, tableName, columnName string

		// MongoDB only allows sorting by a field, so pushing down a
		// non-field requires it to be pre-calculated by a previous
		// push down. If it has been pre-calculated, then it will
		// exist in the mapping registry. Otherwise, it won't, and
		// we'll need to push this down with a $project or $addFields.
		switch typedE := term.expr.(type) {
		case SQLColumnExpr:
			databaseName, tableName, columnName = typedE.databaseName, typedE.tableName,
				typedE.columnName
		case *SQLSubqueryExpr:
			// this saves us from putting a pretty-printed plan as a project key.
			columnName = typedE.plan.Columns()[0].Name
		default:
			columnName = typedE.String()
		}

		fieldName, ok := ms.mappingRegistry.lookupFieldName(databaseName, tableName, columnName)
		if !ok {
			// Since we can't push this down, we'll attempt to build up a $project/$addFields
			// that will allow us to push this down using aggregation language, then sort by the
			// added columns.
			if t == nil {
				t = newInternalPushdownTranslator(v.cfg, ms.mappingRegistry.lookupFieldRef, v)
			}

			var translated ast.Expr
			var err PushdownFailure

			if translated, err = t.TranslateExpr(term.expr); err != nil {
				v.logger.Warnf(log.Dev, "unable to push down order by due to term \n'%v'", columnName)
				v.addPushdownFailure(orderBy, err)
				return orderBy, nil
			}

			fieldName = v.uniqueFieldName(sanitizeFieldName(columnName), ms.mappingRegistry)
			newFields[fieldName] = translated
		}
		sortItems[i] = ast.NewSortItem(astutil.FieldRefFromFieldName(fieldName), !term.ascending)
	}

	pipeline := ast.NewPipeline(ms.pipeline.Stages...)
	if t != nil {
		pipeline.Stages = append(pipeline.Stages, t.subqueryLookupStages...)
	}

	if len(newFields) > 0 {
		// NOTE: there is no reason to mess with the mapping registry
		// because the added fields are only used in the immediate
		// $sort stage and will never be referenced again.
		if !t.versionAtLeast(3, 4, 0) {
			project := v.projectAllColumns(ms.mappingRegistry, newFields)
			pipeline.Stages = append(pipeline.Stages, project)
		} else {
			addFieldsItems := make([]*ast.AddFieldsItem, len(newFields))
			i := 0
			for k, v := range newFields {
				addFieldsItems[i] = ast.NewAddFieldsItem(k, v)
				i++
			}

			pipeline.Stages = append(pipeline.Stages, ast.NewAddFieldsStage(addFieldsItems...))
		}

	}

	pipeline.Stages = append(pipeline.Stages, ast.NewSortStage(sortItems...))

	ms = ms.clone().(*MongoSourceStage)
	ms.pipeline = pipeline
	return ms, nil
}

const (
	emptyFieldNamePrefix = "__empty"
)

// hasColumnReference checks if any SQLColumnExpr is referenced
// within any of the expressions in projectedColumns.
func (v *pushdownVisitor) hasColumnReference(projectedColumns ProjectedColumns) (bool, error) {
	for _, projectedColumn := range projectedColumns {
		refdCols, err := referencedColumns(v.selectIDsInScope, projectedColumn.Expr, false)
		if err != nil {
			return false, err
		}
		if refdCols != nil {
			return true, nil
		}
	}
	return false, nil
}

func (v *pushdownVisitor) visitProject(project *ProjectStage) (PlanStage, error) {
	// Check if we can pushdown further, if the child operator has a MongoSource.
	ms, ok := v.canPushdown(project.source)
	if !ok {
		v.addTransitivePushdownFailure(project, projectStageName)
		return project, nil
	}

	fieldsToProject := make([]ast.ProjectItem, 0)
	uniqueFields := make(map[string]struct{})

	// Check if this project stage is the topmost stage.
	// If ms is also a Dual Stage, continue with normal pushdown instead of using RowGenerator optimization.
	// This is so that in Dual cases full pushdown can be achieved and the fastIter can be used.
	if v.depth == 0 && !ms.IsDual() {
		hasColumnRef, err := v.hasColumnReference(project.projectedColumns)
		if err != nil {
			v.logger.Warnf(log.Dev, "cannot find referenced project expression: %v", err)
			return nil, err
		}

		// If no columns are referenced, we can apply the row generator optimization.
		if !hasColumnRef {
			var stage ast.Stage
			if procutil.VersionAtLeast(v.cfg.mongoDBVersion, []uint8{3, 4, 0}) {
				stage = ast.NewCountStage("rowCount")
			} else {
				stage = ast.NewGroupStage(
					ast.NewDocument(),
					ast.NewGroupItem(
						"rowCount",
						ast.NewFunction(bsonutil.OpSum, astutil.OneInt32Literal),
					),
				)
			}

			newMappingRegistry := newMappingRegistry()
			newColumn := results.NewColumn(ms.selectIDs[0], "", "", "", "rowCount", "", "rowCount",
				types.EvalUint64, schema.MongoInt64, false, true)
			if newMappingRegistry.registerMapping(newColumn.Database, newColumn.Table, newColumn.Name, newColumn.MappingRegistryName, false) {
				newMappingRegistry.addColumn(newColumn)
			}

			ms = ms.clone().(*MongoSourceStage)
			ms.pipeline.Stages = append(ms.pipeline.Stages, stage)
			ms.mappingRegistry = newMappingRegistry
			rg := NewRowGeneratorStage(ms, newColumn)
			newProject := project.clone().(*ProjectStage)
			newProject.source = rg
			return newProject, nil
		}
	}

	// This will contain the rebuilt mapping registry reflecting fields re-mapped by projection.
	fixedMappingRegistry := newMappingRegistry()

	fixedProjectedColumns := ProjectedColumns{}

	// Track whether or not we've successfully mapped every field into the $project of the source.
	// If so, this Project node can be removed from the query plan tree.
	canReplaceProject := true

	t := newInternalPushdownTranslator(v.cfg, ms.mappingRegistry.lookupFieldRef, v)

	for _, projectedColumn := range project.projectedColumns {
		// Convert the column's SQL expression into an expression in mongo query language.
		projectedField, err := t.TranslateExpr(projectedColumn.Expr)
		if err != nil {
			v.addPushdownFailure(project, err)
			v.logger.Debugf(log.Dev, "could not translate projected column '%v'", projectedColumn.String())

			// Expression can't be translated, so it can't be projected.
			// We skip it and leave this Project node in the query plan so that it still gets
			// evaluated during execution.
			canReplaceProject = false

			// There might still be fields referenced in this expression
			// that we still need to project, so collect them and add them to the projection.
			refdCols, err := referencedColumns(v.selectIDsInScope, projectedColumn.Expr, true)
			if err != nil {
				v.logger.Warnf(log.Dev, "cannot find referenced project expression: %v", err)
				v.addNewPushdownFailure(
					project, projectStageName,
					"cannot find referenced project expression",
					"error", err.Error(),
					"expr", projectedColumn.String(),
				)
				return nil, err
			}

			for _, refdCol := range refdCols {
				refdCol.PrimaryKey = projectedColumn.PrimaryKey
				fieldName, ok := ms.mappingRegistry.lookupFieldName(refdCol.Database, refdCol.Table, refdCol.Name)
				if !ok {
					// TODO: BI-2339, change back to panic once we can.
					//panic(fmt.Sprintf("cannot find referenced column %#v in registry", refdCol))
					v.logger.Warnf(log.Dev, "cannot find referenced column %#v in registry",
						refdCol)
					return project, nil
				}

				safeFieldName := sanitizeFieldName(fieldName)
				if _, ok := uniqueFields[safeFieldName]; !ok {
					fieldsToProject = append(fieldsToProject,
						ast.NewAssignProjectItem(safeFieldName,
							getProjectedFieldName(fieldName, refdCol.EvalType)))
					uniqueFields[safeFieldName] = struct{}{}
				}
				if fixedMappingRegistry.registerMapping(refdCol.Database, refdCol.Table, refdCol.Name, safeFieldName, false) {
					fixedMappingRegistry.addColumn(refdCol)
				}
			}

			fixedProjectedColumns = append(fixedProjectedColumns, projectedColumn)
		} else {
			safeFieldName := sanitizeFieldName(projectedColumn.String())
			// If the name turns out to be empty, we need to assign our own.
			if safeFieldName == "" {
				safeFieldName = emptyFieldNamePrefix
			}
			safeFieldName = v.uniqueFieldName(safeFieldName, fixedMappingRegistry)

			if _, ok := uniqueFields[safeFieldName]; !ok {
				fieldsToProject = append(fieldsToProject, ast.NewAssignProjectItem(safeFieldName, projectedField))
				uniqueFields[safeFieldName] = struct{}{}
			}
			registryName := v.uniqueRegistryName(fixedMappingRegistry, projectedColumn.Database,
				projectedColumn.Table, projectedColumn.Name)

			if projectedColumn.Name != registryName {
				projectedColumn.MappingRegistryName = registryName
			}

			if fixedMappingRegistry.registerMapping(projectedColumn.Database, projectedColumn.Table, registryName, safeFieldName, false) {
				fixedMappingRegistry.addColumn(projectedColumn.Column)
			}

			columnExpr := NewSQLColumnExpr(
				projectedColumn.SelectID,
				projectedColumn.Database,
				projectedColumn.Table,
				projectedColumn.Name,
				projectedColumn.EvalType,
				projectedColumn.MongoType,
				false,
				projectedColumn.Column.Nullable,
			)

			fixedProjectedColumns = append(fixedProjectedColumns,
				ProjectedColumn{
					Column: projectedColumn.Column,
					Expr:   columnExpr,
				},
			)
		}
	}

	if len(fieldsToProject) == 0 {
		v.logger.Warnf(log.Dev, "no fields for project push down")
		return project, nil
	}

	if v.depth == 0 {
		if _, ok := uniqueFields[mongoPrimaryKey]; !ok {
			// If we make it to here, we are at the top level, but we have column references.
			// Get rid of _id, it will be projected to a fully qualified name, if actually needed,
			// or it would already exist in uniqueFields. This saves us some data across the wire.
			fieldsToProject = append(fieldsToProject, ast.NewExcludeProjectItem(ast.NewFieldRef(mongoPrimaryKey, nil)))
		}
	}

	ms = ms.clone().(*MongoSourceStage)
	ms.pipeline.Stages = append(ms.pipeline.Stages, t.subqueryLookupStages...)
	ms.pipeline.Stages = append(ms.pipeline.Stages, ast.NewProjectStage(fieldsToProject...))
	ms.mappingRegistry = fixedMappingRegistry

	if canReplaceProject {
		return ms, nil
	}

	project = project.clone().(*ProjectStage)
	project.source = ms
	project.projectedColumns = fixedProjectedColumns

	return project, nil
}

// nolint: unparam
func (v *pushdownVisitor) visitSubquerySource(subquery *SubquerySourceStage) (PlanStage, error) {
	// Check if we can pushdown further, if the child operator has a MongoSource.
	ms, ok := v.canPushdown(subquery.source)
	if !ok {
		v.addTransitivePushdownFailure(subquery, subquerySourceStageName)
		return subquery, nil
	}

	mr := newMappingRegistry()
	for _, column := range ms.mappingRegistry.columns {
		fieldName, ok := ms.mappingRegistry.lookupFieldName(column.Database, column.Table, column.Name)
		if !ok {
			v.logger.Warnf(log.Dev, "cannot find referenced subquery column %#v in lookup", column)
			v.addNewPushdownFailure(
				subquery, subquerySourceStageName,
				"cannot find referenced subquery column in lookup",
				"column", fmt.Sprintf("%#v", column),
			)
			return subquery, nil
		}

		if mr.registerMapping(column.Database, subquery.aliasName, column.Name, fieldName, false) {
			newColumn := column.Clone()
			newColumn.Table = subquery.aliasName
			mr.addColumn(newColumn)
		}
	}

	ms = ms.clone().(*MongoSourceStage)
	ms.aliasNames = []string{subquery.aliasName}
	ms.mappingRegistry = mr
	return ms, nil
}

// pushdownProject is called when a stage could not be pushed down. It uses
// columnExprs to create and visit a new projectStage in order to project
// out only the columns needed for the rest of the query so that we do not have
// to pull all data from a table into memory.
func (v *pushdownVisitor) pushdownProject(columnExprs []SQLColumnExpr, source PlanStage) (PlanStage, error) {
	sf := &sourceFinder{}
	_, err := sf.visit(source)
	if err != nil {
		return nil, err
	}

	var columns ProjectedColumns
	for _, columnExpr := range columnExprs {
		isPK := sf.source.mappingRegistry.isPrimaryKey(columnExpr.columnName)
		_, ok := sf.source.mappingRegistry.lookupFieldName(
			columnExpr.databaseName,
			columnExpr.tableName,
			columnExpr.columnName)
		if !ok {
			// skip any column that we need, but that does not exist in the source
			continue
		}

		// Mark the columns as not correlated. In the context of the newly
		// created projectStage, these columns are no longer correlated.
		// Consider the following example:
		//
		//   select * from foo where a in (select a from foo);
		//
		//   this is rewritten as:
		//   select * from foo where (select a) in (select a from foo);
		//
		// The FilterStage will fail to pushdown because of the correlated
		// SQLColumnExpr in the left subquery. Since it fails to pushdown,
		// we'll replace its source with the projectStage created by this
		// function. In this projectStage, none of the columns should be
		// marked as correlated even if their corresponding columnExpr is
		// since this projectStage will act as the (non-correlated) source
		// for the columnExpr.
		columnExpr.correlated = false

		column := NewColumnFromSQLColumnExpr(columnExpr, isPK)
		columns = append(columns, ProjectedColumn{Column: column, Expr: columnExpr})
	}

	plan, err := v.visitProject(NewProjectStage(source, columns...))
	if err != nil {
		return nil, fmt.Errorf("unable to push down project: %v", err)
	}
	return plan, nil
}

func (v *pushdownVisitor) projectAllColumns(mr *mappingRegistry, newFields map[string]ast.Expr) *ast.ProjectStage {
	projectItems := make([]ast.ProjectItem, len(mr.columns)+len(newFields))
	for i, c := range mr.columns {
		fieldName, ok := mr.lookupFieldName(c.Database, c.Table, c.Name)
		if !ok {
			panic("unable to find field mapping for column. This should never happen.")
		}
		projectItems[i] = ast.NewIncludeProjectItem(astutil.FieldRefFromFieldName(fieldName))
	}
	i := len(mr.columns)
	for fieldName, expr := range newFields {
		projectItems[i] = ast.NewAssignProjectItem(fieldName, expr)
		i++
	}
	return ast.NewProjectStage(projectItems...)
}

// uniqueFieldName creates a field name that is unique across all tables in a
// set of registries.
func (v *pushdownVisitor) uniqueFieldName(fieldName string, mrs ...*mappingRegistry) string {
	retFieldName := fieldName
	i := 0

TOP:
	for {
		for _, mr := range mrs {
			if mr.containsFieldName(retFieldName) {
				retFieldName = fmt.Sprintf("%v_%v", fieldName, i)
				i++
				continue TOP
			}
		}
		return retFieldName
	}
}

// uniqueLetVarName creates a field name that is unique across all tables in a
// set of registries for use within a $let var block.
func (v *pushdownVisitor) uniqueLetVarName(fieldName string, mrs ...*mappingRegistry) string {
	return v.uniqueFieldName(sanitizeLetVarName(fieldName), mrs...)
}

// uniqueRegistryName creates a name that is unique to a table: they can be
// repeated in separate tables, use uniqueFieldName for a field name that is
// unique in the entire registry (or set of registries).
func (v *pushdownVisitor) uniqueRegistryName(mr *mappingRegistry, databaseName, tableName,
	columnName string) string {
	if _, hasField := mr.lookupFieldName(databaseName, tableName, columnName); !hasField {
		return columnName
	}

	i := 1
	for {
		retColumnName := fmt.Sprintf("%v_%v", columnName, i)
		if _, hasField := mr.lookupFieldName(databaseName, tableName, retColumnName); !hasField {
			return retColumnName
		}
		i++
	}
}

func (v *pushdownVisitor) canSelfJoinTables(logger log.Logger, local, foreign *MongoSourceStage,
	matcher SQLExpr, kind JoinKind) bool {
	return canSelfInnerJoinTables(logger, local, foreign, matcher, v.leftJoinOriginalNames) &&
		(kind != LeftJoin || v.meetsLeftSelfJoinPipelineCriteria(logger, local, foreign, matcher))
}

func canSelfInnerJoinTables(logger log.Logger, local, foreign *MongoSourceStage,
	matcher SQLExpr, leftJoinOriginalNames dbData) bool {
	return sharesRootTable(logger, local, foreign) &&
		meetsSelfJoinPKCriteria(logger, local, foreign, matcher, leftJoinOriginalNames)
}

func (v *pushdownVisitor) remainingJoinPredicate(msLocal, msForeign *MongoSourceStage,
	e SQLExpr) []SQLExpr {
	exprs, newExprs := splitExpression(e), []SQLExpr{}
	registries := []*mappingRegistry{msLocal.mappingRegistry,
		msForeign.mappingRegistry}
	for _, expr := range exprs {
		if equalExpr, ok := expr.(*SQLComparisonExpr); ok && equalExpr.op == EQ {
			c1, _ := equalExpr.left.(SQLColumnExpr)
			c2, _ := equalExpr.right.(SQLColumnExpr)
			if c1.selectID == c2.selectID {

				originalC1Name, _, c1RegistryIdx, ok := lookupSQLColumnForJoin(c1.databaseName,
					c1.tableName, c1.columnName, registries, v.leftJoinOriginalNames)
				if !ok {
					panic("unable to find field mapping for self-join optimization " +
						"c1. This should never happen.")
				}

				originalC2Name, _, c2RegistryIdx, ok := lookupSQLColumnForJoin(c2.databaseName,
					c2.tableName, c2.columnName, registries, v.leftJoinOriginalNames)
				if !ok {
					panic("unable to find field mapping for self-join optimization " +
						"c2. This should never happen.")
				}

				c1IsPK := registries[c1RegistryIdx].
					isPrimaryKey(c1.columnName)
				c2IsPK := registries[c2RegistryIdx].
					isPrimaryKey(c2.columnName)

				if c1IsPK && c2IsPK && originalC1Name == originalC2Name {
					continue
				}
			}
		}
		newExprs = append(newExprs, expr)
	}
	return newExprs
}

func (v *pushdownVisitor) meetsLeftSelfJoinPipelineCriteria(logger log.Logger, local,
	foreign *MongoSourceStage, matcher SQLExpr) bool {
	hasRemainingPredicate := len(v.remainingJoinPredicate(local, foreign, matcher)) != 0
	// Get the paths of each unwind in each pipeline as an array of strings.
	localUnwindPipelineFields := astutil.GetPipelineUnwindFields(local.pipeline.Stages)
	foreignUnwindPipelineFields := astutil.GetPipelineUnwindFields(foreign.pipeline.Stages)

	lenLocal, lenForeign := len(localUnwindPipelineFields), len(foreignUnwindPipelineFields)

	// We can't have any issues with embedded NULLs and empties if the
	// foreign pipeline only has one $unwind and there is no remaining
	// predicate.
	if lenForeign == 1 {
		return true
	}

	// We meet the left self join criteria if both sides of the joins have no pipelines and there
	// are no remaining predicates after the matcher has been extracted.
	if lenLocal == lenForeign && lenLocal == 0 && !hasRemainingPredicate {
		return true
	}

	localUnwindPipelinePaths, foreignUnwindPipelinePaths := astutil.GetPaths(
		localUnwindPipelineFields),
		astutil.GetPaths(foreignUnwindPipelineFields)

	// sharesPrefix ensures that one of these tables is a progeny of the
	// other. If the progeny is on the foreign side, it can be no younger
	// than the child. If it is on the local side, we don't care how far
	// removed they are because it is impossible for a key to exist in the
	// progeny that does not exist in the parent, but not vice versa. It's
	// only true, however, that the local side can be younger than the
	// foreign side if there is no remaining predicate.
	if sharesPrefix(localUnwindPipelinePaths, foreignUnwindPipelinePaths) &&
		(lenForeign == lenLocal+1 ||
			lenForeign <= lenLocal && !hasRemainingPredicate) {
		// Building the remaining predicate is completely wrong when
		// the local side is older than the foreign side, or,
		// unfortunately, the same table. On the bright side, we can
		// generally anticipate that the left side of left joins will
		// generally be the younger in our users queries.
		return true
	}
	// Non-shared prefix paths can pose problems, regardless of length,
	// because of mgoPreserveNullAndEmptyArrays we would only want to keep
	// NULLs and empties from the top level of unwinding in the foreign
	// side. Theoretical attempts to do such filtering have all fallen
	// down on edge cases. This is also why we disallow progeny with more
	// than 1 generation difference.
	logger.Debugf(log.Dev, "self-join optimization: could not optimize because of left join "+
		"criteria - local unwind paths are: %v foreign unwind paths %v",
		localUnwindPipelinePaths, foreignUnwindPipelinePaths)
	return false
}

func meetsSelfJoinPKCriteria(logger log.Logger, local,
	foreign *MongoSourceStage, matcher SQLExpr, leftJoinOriginalNames dbData) bool {
	// Don't perform optimization on MongoDB views as
	// renames might have occurred on fields.
	if local.isView() {
		logger.Debugf(log.Dev, "cannot use self-join optimization, local "+
			"table is MongoDB view")
		return false
	}

	if foreign.isView() {
		logger.Debugf(log.Dev, "cannot use self-join optimization, foreign "+
			"table is MongoDB view")
		return false
	}

	exprs := splitExpression(matcher)

	getPKs := func(columns []*results.Column) map[string]struct{} {
		keys := make(map[string]struct{})
		for _, c := range columns {
			if _, counted := keys[c.Name]; !counted && c.PrimaryKey {
				keys[c.Name] = struct{}{}
			}
		}
		return keys
	}

	// Whether or not we are joining the same tables or different tables
	// derived from a single collection, we need to join on the entire PK
	// intersection in order for self-join optimization to be semantically
	// correct. We set the number of PK matches needed based on the
	// cardinality of the intersection and then assure below that that
	// number is met.
	localPKs := getPKs(local.mappingRegistry.columns)
	foreignPKs := getPKs(foreign.mappingRegistry.columns)
	intersectionPKs := intersectionStringSet(localPKs, foreignPKs)

	numRequiredPKConjunctions := len(intersectionPKs)

	if numRequiredPKConjunctions == 0 {
		logger.Debugf(log.Dev, "cannot use self-join optimization, table has no primary key")
		return false
	}

	numPKConjunctions := 0

	logger.Debugf(log.Dev, "self-join optimization: examining match criteria...")

	registries := []*mappingRegistry{local.mappingRegistry, foreign.mappingRegistry}

	seenPrimaryKeys := make(map[string]struct{})

	for _, expr := range exprs {
		if equalExpr, ok := expr.(*SQLComparisonExpr); ok && equalExpr.op == EQ {
			column1, _ := equalExpr.left.(SQLColumnExpr)
			column2, _ := equalExpr.right.(SQLColumnExpr)

			invalidLeftColumn := (!containsString(local.aliasNames, column1.tableName) &&
				!containsString(foreign.aliasNames, column1.tableName)) ||
				(local.dbName != column1.databaseName && foreign.dbName != column1.databaseName)
			invalidRightColumn := (!containsString(local.aliasNames, column2.tableName) &&
				!containsString(foreign.aliasNames, column2.tableName)) ||
				(local.dbName != column2.databaseName && foreign.dbName != column2.databaseName)

			if invalidLeftColumn || invalidRightColumn {
				logger.Debugf(log.Dev, "self-join optimization: found unexpected "+
					"table references, moving on...")
				continue
			}

			if column1.selectID != column2.selectID {
				logger.Debugf(log.Dev, "self-join optimization: found unmatched "+
					"select identifiers (%v and %v), moving on...",
					column1.selectID, column2.selectID)
				continue
			}

			originalC1Name, _, c1RegistryIdx, ok := lookupSQLColumnForJoin(column1.databaseName,
				column1.tableName, column1.columnName, registries, leftJoinOriginalNames)
			if !ok {
				panic(fmt.Sprintf("unable to find field mapping for merge column1:  %s.%s.%s."+
					" This should never happen.", column1.databaseName, column1.tableName,
					column1.columnName))
			}
			originalC2Name, _, c2RegistryIdx, ok := lookupSQLColumnForJoin(column2.databaseName,
				column2.tableName, column2.columnName, registries, leftJoinOriginalNames)
			if !ok {
				panic(fmt.Sprintf("unable to find field mapping for merge column2: %s.%s.%s."+
					" This should never happen.", column2.databaseName, column2.tableName,
					column2.columnName))
			}

			c1IsPK := registries[c1RegistryIdx].isPrimaryKey(column1.columnName)
			c2IsPK := registries[c2RegistryIdx].isPrimaryKey(column2.columnName)

			if !c1IsPK || !c2IsPK {
				logger.Debugf(log.Dev, "self-join optimization: criteria contains "+
					"non-primary key (%v and %v), moving on...",
					column1.String(), column2.String())
				continue
			}

			if originalC1Name != originalC2Name {
				logger.Debugf(log.Dev, "self-join optimization: criteria contains "+
					"unmatched primary keys (%v and %v), moving on...",
					originalC1Name, originalC2Name)
				continue
			}

			if _, ok := seenPrimaryKeys[originalC1Name]; ok {
				logger.Debugf(log.Dev, "self-join optimization: ignoring duplicate "+
					"primary key criteria '%v' and moving on...",
					column1.String())
				continue
			}

			seenPrimaryKeys[originalC1Name] = struct{}{}

			numPKConjunctions++
		}
	}

	if numPKConjunctions < numRequiredPKConjunctions {
		loggingPKSetStr := strings.Join(keysStringSet(intersectionPKs), ", ")
		logger.Debugf(log.Dev, "self-join optimization: criteria conjunction "+
			"contains %v unique primary key equality %v but need %v - %q",
			numPKConjunctions, strutil.Pluralize(numPKConjunctions, "pair",
				"pairs"), numRequiredPKConjunctions, loggingPKSetStr)
		return false
	}

	return true
}

// canPushdownJoinKind returns nil if the join's kind can be pushed down to MongoDB.
// MongoDB can only do an inner join and a left outer join.
// If the join kind cannot be pushed down this method returns a pushdown failure.
func (v *pushdownVisitor) canPushdownJoinKind(kind JoinKind) PushdownFailure {
	switch kind {
	case InnerJoin, LeftJoin, StraightJoin:
		return nil
	default:
		return newPushdownFailure(joinStageName, fmt.Sprintf("regular $lookup cannot push down join kind '%v' - join kind is not inner, left, or straight", kind))
	}
}

func (v *pushdownVisitor) attemptToOptimizeSelfJoins(join *JoinStage, msLocal, msForeign *MongoSourceStage) (PlanStage, error) {
	if v.cfg.pushDownSelfJoins {

		// Before attempting the self-join optimization, check that the
		// underlying collection is the same for both tables and that the join
		// criteria holds the primary key for both.
		if v.canSelfJoinTables(v.logger, msLocal, msForeign, join.matcher, join.kind) {
			ms, err := v.optimizeSelfJoinTables(msLocal, msForeign, join)
			// For tables in different databases, it is not possible to push down since this isn't
			// supported in MongoDB, just do the join in memory.
			if err != nil {
				return nil, err
			}

			if ms != nil {
				v.logger.Debugf(log.Dev, "successfully self-join optimized tables %v "+
					"and %v", msLocal.aliasNames, msForeign.aliasNames)
				return ms, nil
			}

			v.logger.Debugf(log.Dev, "unable to self-join optimize tables %v and %v",
				msLocal.aliasNames, msForeign.aliasNames)
		}
	} else {
		v.logger.Warnf(log.Admin, "optimize_self_joins is false: skipping self join optimization")
	}

	return nil, nil
}

func (v *pushdownVisitor) canPushdownForeignJoinSourcePipeline(join *JoinStage, msLocal, msForeign *MongoSourceStage) bool {
	lenForeignPipeline := len(msForeign.pipeline.Stages)

	if lenForeignPipeline > 1 {
		v.logger.Warnf(log.Dev,
			"unable to translate join stage to lookup: foreign table pipeline has more than one stage")
		v.addNewPushdownFailure(join, joinStageName, "foreign table pipeline has more than one stage")
		return false
	} else if lenForeignPipeline > 0 {
		unwind, foreignHasUnwind := msForeign.pipeline.Stages[0].(*ast.UnwindStage)
		if !foreignHasUnwind {
			v.logger.Warnf(log.Dev,
				"unable to translate join stage to lookup: foreign table pipeline stage is not $unwind")
			v.addNewPushdownFailure(join, joinStageName, "foreign table pipeline stage is not $unwind")
			return false
		}
		// These registries will be needed in the loop over join exprs below.
		registries := []*mappingRegistry{
			msLocal.mappingRegistry,
			msForeign.mappingRegistry,
		}
		// Check to make sure the single unwind in the foreign pipeline
		// doesn't have an array index created by the unwind in its
		// join condition, otherwise we build an impossible $lookup
		// and an empty return set.
		if unwind.IncludeArrayIndex != "" {
			exprs := splitExpression(join.matcher)
			for _, expr := range exprs {
				// Ignore non-equalExpr join conditions, since
				// they will be handled after any foreign
				// $unwinds as a $match or remaining left join predicate
				// (see buildRemainingLeftJoinPredicate) and thus not
				// cause any issues.
				if equalExpr, ok := expr.(*SQLComparisonExpr); ok && equalExpr.op == EQ {
					column1, _ := equalExpr.left.(SQLColumnExpr)
					column2, _ := equalExpr.right.(SQLColumnExpr)
					// It's possible that someone could use
					// the foreign table on either or both
					// sides of the join equivalence, so we
					// can't use else here.
					if containsString(msForeign.aliasNames, column1.tableName) {
						_, columnName, _, _ := lookupSQLColumnForJoin(column1.databaseName,
							column1.tableName, column1.columnName, registries, v.leftJoinOriginalNames)

						if columnName == unwind.IncludeArrayIndex {
							v.logger.Debugf(log.Dev, "$lookup translation: cannot use foreign "+
								"unwind index: %q in equality criteria because use in $lookup "+
								"occurs before foreign unwind, moving on...", unwind.IncludeArrayIndex)
							v.addNewPushdownFailure(join, joinStageName, "foreign unwind index use in $lookup occurs before $unwind")
							return false
						}
					}
					if containsString(msForeign.aliasNames, column2.tableName) {
						_, columnName, _, _ := lookupSQLColumnForJoin(column2.databaseName,
							column2.tableName, column2.columnName, registries, v.leftJoinOriginalNames)

						if columnName == unwind.IncludeArrayIndex {
							v.logger.Debugf(log.Dev, "$lookup translation: cannot use foreign "+
								"unwind index: %q in equality criteria, because use in $lookup "+
								"occurs before foreign unwind, moving on...", unwind.IncludeArrayIndex)
							v.addNewPushdownFailure(join, joinStageName, "foreign unwind index use in $lookup occurs before $unwind")
							return false
						}
					}
				}
			}
		}
	}
	return true
}

func (v *pushdownVisitor) doesJoinHaveIncompatibleUUIDs(lookupInfo *lookupInfo) PushdownFailure {
	// Prevent join pushdown when UUID subtype 3 encoding is different.
	localMongoType := lookupInfo.localColumn.columnType.MongoType
	foreignMongoType := lookupInfo.foreignColumn.columnType.MongoType

	if values.IsUUID(localMongoType) && values.IsUUID(foreignMongoType) {
		if localMongoType != foreignMongoType {
			v.logger.Warnf(log.Dev,
				"unable to translate join stage to lookup: found different criteria UUID - %v and %v",
				localMongoType, foreignMongoType,
			)
			return newPushdownFailure(
				joinStageName,
				"different UUID subtype 3 encodings in left and right tables",
				"localMongoType", string(localMongoType),
				"foreignMongoType", string(foreignMongoType),
			)
		}
	}

	return nil
}

func (v *pushdownVisitor) getLookupFieldsFromJoinSources(msLocal, msForeign *MongoSourceStage, lookupInfo *lookupInfo) (string, string, PushdownFailure) {
	localFieldName, ok := msLocal.mappingRegistry.lookupFieldName(
		lookupInfo.localColumn.databaseName,
		lookupInfo.localColumn.tableName,
		lookupInfo.localColumn.columnName)
	if !ok {
		v.logger.Warnf(log.Dev, "cannot find referenced local join column %#v in lookup", lookupInfo.localColumn)
		return "", "", newPushdownFailure(joinStageName, "cannot find referenced local column",
			"column", lookupInfo.localColumn.String(),
		)
	}

	foreignFieldName, ok := msForeign.mappingRegistry.lookupFieldName(
		lookupInfo.foreignColumn.databaseName,
		lookupInfo.foreignColumn.tableName,
		lookupInfo.foreignColumn.columnName)
	if !ok {
		v.logger.Warnf(log.Dev, "cannot find referenced foreign join column %#v in lookup", lookupInfo.foreignColumn)
		return "", "", newPushdownFailure(joinStageName, "cannot find referenced foreign column",
			"column", lookupInfo.foreignColumn.String(),
		)
	}

	return localFieldName, foreignFieldName, nil
}

// createNullSafeLocalPipeline returns a pipeline that helps us treat nulls in mongodb like mysql.
// Because MongoDB does not compare nulls in the same way as MySQL, we need an extra
// $project to account for this incompatibility. Effectively, when our left
// hand field is null, we'll empty the joined results prior to unwinding.
func createNullSafeLocalPipeline(msLocal *MongoSourceStage, localFieldName, asField string) *ast.ProjectStage {
	projectItems := make([]ast.ProjectItem, 1, len(msLocal.mappingRegistry.columns)+1)

	projectItems[0] = ast.NewAssignProjectItem(asField, astutil.WrapInNullCheckedCond(
		ast.NewArray(),
		astutil.FieldRefFromFieldName(asField),
		astutil.FieldRefFromFieldName(localFieldName),
	))

	// Enumerate all the local fields.
	for _, c := range msLocal.mappingRegistry.columns {
		fieldName, ok := msLocal.mappingRegistry.lookupFieldName(
			c.Database, c.Table, c.Name)
		if !ok {
			panic(fmt.Sprintf("unable to find field mapping for column %s.%s.%s. This "+
				"should never happen.", c.Database, c.Table, c.Name))
		}

		if fieldName == asField {
			continue
		}

		projectItems = append(projectItems, ast.NewIncludeProjectItem(astutil.FieldRefFromFieldName(fieldName)))
	}

	return ast.NewProjectStage(projectItems...)
}

func (v *pushdownVisitor) generateForeignUnwindPipeline(join *JoinStage, newMappingRegistry *mappingRegistry,
	localFieldName, foreignFieldName, asField string, lookupOnUnwindPath bool) ([]ast.Stage, PushdownFailure) {

	msForeign, _ := join.right.(*MongoSourceStage)

	foreignMapped := msForeign.pipeline.Stages[0].(*ast.UnwindStage)
	foreignUnwindPath := astutil.RefString(foreignMapped.Path)

	// Strip dollar sign prefix, and prepend with asField.
	if foreignUnwindPath != "" {
		foreignUnwindPath = fmt.Sprintf("%v.%v", asField, foreignUnwindPath)
	} else {
		v.logger.Warnf(log.Dev, "empty $unwind path specification")
		return nil, newPushdownFailure(joinStageName, "empty $unwind path specification")
	}

	// For left joins, preserve null and empty arrays to ensure
	// that every local pipeline record gets returned.
	idx := fmt.Sprintf("%v.%v", asField, foreignMapped.IncludeArrayIndex)
	foreignUnwind := ast.NewUnwindStage(astutil.FieldRefFromFieldName(foreignUnwindPath), idx, join.kind == LeftJoin)

	v.logger.Debugf(log.Dev, "consolidating foreign unwind into local pipeline")

	stages := []ast.Stage{foreignUnwind}

	// Handle edge case where the lookup field is the same as the
	// $unwind's array path. In this case, we must apply an
	// additional filter to remove records in the now unwound array
	// that don't meet the lookup criteria.
	if lookupOnUnwindPath {
		filter := ast.NewBinary(bsonutil.OpEq,
			astutil.FieldRefFromFieldName(fmt.Sprintf("%s.%s", asField, foreignFieldName)),
			astutil.FieldRefFromFieldName(localFieldName),
		)

		fieldName := v.uniqueFieldName(projectPredicateFieldName, newMappingRegistry)
		stageBody := []*ast.AddFieldsItem{ast.NewAddFieldsItem(fieldName, filter)}
		predicateEvaluationStage := v.buildAddFieldsOrProject(stageBody, []string{}, newMappingRegistry)
		stages = append(stages, predicateEvaluationStage)

		var match ast.Expr = ast.NewBinary(bsonutil.OpEq, astutil.FieldRefFromFieldName(fieldName), astutil.TrueLiteral)
		if join.kind == LeftJoin {
			// For left joins, we need to ensure we retain records from the
			// left child - in case the unwound array was empty or null.
			match = ast.NewBinary(bsonutil.OpOr,
				match,
				ast.NewDocument(
					ast.NewDocumentElement(
						foreignUnwindPath,
						ast.NewFunction(bsonutil.OpExists, astutil.FalseLiteral),
					),
				),
			)
		}

		stages = append(stages, ast.NewMatchStage(match))
	}

	return stages, nil
}

func (v *pushdownVisitor) buildJoinPushdownPipeline(join *JoinStage, lookupInfo *lookupInfo) (PlanStage, PushdownFailure) {
	msLocal, _ := join.left.(*MongoSourceStage)
	msForeign, _ := join.right.(*MongoSourceStage)

	// Get the lookup fields.
	localFieldName, foreignFieldName, failure := v.getLookupFieldsFromJoinSources(msLocal,
		msForeign, lookupInfo)
	if failure != nil {
		return nil, failure
	}

	// Create a field name that we will add the looked-up documents to.
	asField := v.uniqueFieldName(
		sanitizeFieldName(joinedFieldNamePrefix+msForeign.aliasNames[0]),
		msLocal.mappingRegistry,
	)
	// Compute all the mappings from the msForeign mapping registry
	// to be nested under the 'asField' we used above.
	newMappingRegistry := msLocal.mappingRegistry.merge(msForeign.mappingRegistry, asField)

	pipeline := ast.NewPipeline(msLocal.pipeline.Stages...)

	localField := astutil.FieldRefFromFieldName(localFieldName)
	if join.kind == InnerJoin || join.kind == StraightJoin {
		// Because MongoDB does not compare nulls in the same way as MySQL, we need an extra
		// $match to ensure account for this incompatibility. Effectively, we eliminate the
		// left hand document when the left hand field is null.
		pipeline.Stages = append(pipeline.Stages, ast.NewMatchStage(
			ast.NewBinary(bsonutil.OpNeq,
				localField,
				astutil.NullLiteral,
			),
		))
	}

	pipeline.Stages = append(pipeline.Stages, ast.NewLookupStage(
		msForeign.collectionNames[0], localField, foreignFieldName, asField, nil, nil))

	if join.kind == LeftJoin {
		pipeline.Stages = append(pipeline.Stages, createNullSafeLocalPipeline(msLocal, localFieldName, asField))
	}

	unwind := ast.NewUnwindStage(astutil.FieldRefFromFieldName(asField), "", join.kind == LeftJoin)

	lookupOnUnwindPath := false
	oldForeignIndex := ""

	foreignHasUnwind := false
	if len(msForeign.pipeline.Stages) > 0 {
		var foreignMapped *ast.UnwindStage
		if foreignMapped, foreignHasUnwind = msForeign.pipeline.Stages[0].(*ast.UnwindStage); foreignHasUnwind {
			oldForeignPath := astutil.RefString(foreignMapped.Path)
			oldForeignIndex = asField + "." + foreignMapped.IncludeArrayIndex
			lookupOnUnwindPath = strings.Split(foreignFieldName, ".")[0] == oldForeignPath
		}
	}

	// Create pipeline stages for the remaining predicate for left joins if we can.
	if lookupInfo.remainingPredicate != nil && join.kind == LeftJoin {
		if lookupOnUnwindPath && len(strings.Split(foreignFieldName, ".")) > 1 {
			v.logger.Warnf(log.Dev, "unable to translate left join stage to lookup: lookup on nested array field")
			return nil, newPushdownFailure(joinStageName, "lookup on nested array field")
		}

		// Enumerate the columns in the remaining predicate that come from the foreign table.
		foreignCols, err := getTableColumnsInExpr(msForeign, lookupInfo.remainingPredicate)
		if err != nil {
			v.logger.Warnf(log.Dev, "error while visiting left join's remaining predicate: %v", err)
			return nil, newPushdownFailure(joinStageName, "error visiting left join's remaining predicate",
				"error", err.Error(),
			)
		}

		// If the foreign table is an array table and the remaining predicate
		// references a foreign column, we won't translate this.
		if foreignHasUnwind && len(foreignCols) > 0 {
			v.logger.Warnf(log.Dev, "unable to translate left join stage to lookup: remaining predicate references foreign table")
			return nil, newPushdownFailure(joinStageName,
				"remaining left join predicate references foreign table")
		}

		project, match, failure := v.buildRemainingPredicateForLeftJoin(
			newMappingRegistry,
			lookupInfo.remainingPredicate,
			asField,
			oldForeignIndex,
			false,
		)
		if failure != nil {
			return nil, failure
		}

		pipeline.Stages = append(pipeline.Stages, project, unwind)

		if match != nil {
			pipeline.Stages = append(pipeline.Stages, match)
		}
	} else {
		pipeline.Stages = append(pipeline.Stages, unwind)
	}

	// This handles merging foreign tables referenced in joins
	// that contain a single $unwind pipeline stage.
	if foreignHasUnwind {
		foreignUnwindPipeline, failure := v.generateForeignUnwindPipeline(join, newMappingRegistry,
			localFieldName, foreignFieldName, asField, lookupOnUnwindPath)
		if failure != nil {
			return nil, failure
		}

		pipeline.Stages = append(pipeline.Stages, foreignUnwindPipeline...)
	}

	// Build the new operators.
	msLocal.aliasNames = append(msLocal.aliasNames, msForeign.aliasNames...)
	msLocal.tableNames = append(msLocal.tableNames, msForeign.tableNames...)
	msLocal.collectionNames = append(msLocal.collectionNames, msForeign.collectionNames...)
	for key, val := range msForeign.isShardedCollection {
		msLocal.isShardedCollection[key] = val
	}
	ms := msLocal.clone().(*MongoSourceStage)
	ms.selectIDs = append(ms.selectIDs, msForeign.selectIDs...)
	ms.pipeline = pipeline
	ms.mappingRegistry = newMappingRegistry
	ms.LimitRowCount = int(math.Max(-1.0, float64(msLocal.LimitRowCount*msForeign.LimitRowCount)))

	return ms, nil
}

type pipelineOptimizationVisitor struct {
	ctx context.Context
}

func newPipelineOptimizationVisitor(ctx context.Context) pipelineOptimizationVisitor {
	return pipelineOptimizationVisitor{ctx: ctx}
}

func (v *pipelineOptimizationVisitor) visit(n Node) (Node, error) {

	switch typedN := n.(type) {
	case *MongoSourceStage:
		optimizedPipeline := optimizer.Optimize(v.ctx, typedN.pipeline)
		typedN.pipeline = flattenBinaryExprs(optimizedPipeline)
		return n, nil
	}
	return walk(v, n)
}

// flattenBinaryExprs walks a pipeline and flattens nested $add, $and, $multiply and $or
// ast.Binary expressions into ast.Functions.
// For example,
//   Binary{Op: "$add", Left: "$a", Right: Binary{Op: "$add", Left: "$b", Right: "$c"}}
// flattens into
//   Function{Op: "$add", Arg: ["$a", "$b", "$c"]}
// This function
func flattenBinaryExprs(pipeline *ast.Pipeline) *ast.Pipeline {
	newPipeline, _ := ast.Visit(pipeline, func(v ast.Visitor, n ast.Node) ast.Node {
		switch tn := n.(type) {
		case *ast.Binary:
			switch tn.Op {
			case "$add", "$and", "$multiply", "$or":
				args := flattenBinaryExprArgs(tn.Op, tn.Left, tn.Right)
				n = ast.NewFunction(string(tn.Op), ast.NewArray(args...))
			}
		}
		return n.Walk(v)
	})

	return newPipeline.(*ast.Pipeline)
}

func flattenBinaryExprArgs(op ast.BinaryOp, left, right ast.Expr) []ast.Expr {
	args := make([]ast.Expr, 0)

	for _, expr := range []ast.Expr{left, right} {
		if bin, isBinary := expr.(*ast.Binary); isBinary && bin.Op == op {
			args = append(args, flattenBinaryExprArgs(op, bin.Left, bin.Right)...)
		} else {
			args = append(args, expr)
		}
	}

	return args
}
