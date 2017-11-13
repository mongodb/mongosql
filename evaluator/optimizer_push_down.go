package evaluator

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
)

func optimizePushDown(n node, ctx *EvalCtx, logger *log.Logger) (node, error) {
	if os.Getenv(NoPushDown) != "" {
		logger.Warnf(log.Admin, "pushdown is disabled: skipping pushdown optimizer")
		return n, nil
	}

	v := newPushDownOptimizer(ctx, logger)

	n, err := v.visit(n)
	if err != nil {
		return nil, err
	}

	return n, nil
}

type pushDownOptimizer struct {
	logger                *log.Logger
	ctx                   ConnectionCtx
	selectIDsInScope      []int
	tableNamesInScope     map[string][]string
	columnTracker         *columnTracker
	leftJoinOriginalNames map[string]map[string]map[string]string
}

func newPushDownOptimizer(ctx ConnectionCtx, logger *log.Logger) *pushDownOptimizer {
	return &pushDownOptimizer{
		logger:                logger,
		ctx:                   ctx,
		columnTracker:         newColumnTracker(),
		leftJoinOriginalNames: make(map[string]map[string]map[string]string),
	}
}

func (v *pushDownOptimizer) addSelectIDsInScope(selectIDs ...int) {
	for _, selectID := range selectIDs {
		if !containsInt(v.selectIDsInScope, selectID) {
			v.selectIDsInScope = append(v.selectIDsInScope, selectID)
		}
	}
}

func (v *pushDownOptimizer) addTableNamesInScope(databaseName string, tableNames ...string) {
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

// buildAddFieldsOrProject will build an addField if the server version is > 3.4.0,
// if it is less, it will build a project with everything not in the passed in body projected as 1,
// it will also skip any paths prefixed by a string in prefixesToSkip (mainly for avoiding conflicts).
func (v *pushDownOptimizer) buildAddFieldsOrProject(body bson.M, prefixesToSkip []string, mrs ...*mappingRegistry) bson.D {
	if v.ctx.Variables().MongoDBInfo.VersionAtLeast(3, 4, 0) {
		return bson.D{{"$addFields", body}}
	}
	// Make sure any prefix ends with '.'
	for i, prefix := range prefixesToSkip {
		if prefixesToSkip[len(prefixesToSkip)-1] != "." {
			prefixesToSkip[i] = prefix + "."
		}
	}
	ret := bson.D{{"$project", body}}

	// We now need to make sure we project all the existing columns from all mapping registries.
	for _, mr := range mrs {
	TOP:
		for _, c := range mr.columns {
			field, ok := mr.lookupFieldName(c.Database, c.Table, c.Name)
			if !ok {
				panic(fmt.Sprintf("cannot find referenced join column %#v in local lookup in buildAddFieldsOrProject", c))
			}
			// Do not overwrite things already in the projectBody, and do not add paths
			// prefixed by our asField, because we will get conflicts.

			if _, ok := body[field]; !ok {
				// Again, only keep if there isn't a prefix conflict.
				for _, prefix := range prefixesToSkip {
					if strings.HasPrefix(field, prefix) {
						continue TOP
					}
				}
				body[field] = 1
			}
		}
	}
	return ret
}

func (v *pushDownOptimizer) visit(n node) (node, error) {

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

	n, err := walk(v, n)
	if err != nil {
		return nil, err
	}

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
	case *SQLSubqueryExpr:
		// SQLSubqueryExpr only applies to non-from clauses. This means that
		// any new selectIDs found inside a SQLSubqueryExpr are invalid outside
		// of it. However, the selectIDs outside of it are valid inside. This is
		// the definition of a correlated subquery. So, we'll save off the current
		// selectIDs and restore them afterwards.

		oldSelectIDsInScope := v.selectIDsInScope
		oldTableNamesInScope := v.tableNamesInScope
		oldLeftJoinOriginalNames := v.leftJoinOriginalNames

		n, err = walk(v, n)
		if err != nil {
			return nil, err
		}

		v.selectIDsInScope = oldSelectIDsInScope
		v.tableNamesInScope = oldTableNamesInScope
		v.leftJoinOriginalNames = oldLeftJoinOriginalNames
	// Push Down
	case *FilterStage:
		n, err = v.visitFilter(typedN)
		if err != nil {
			return nil, fmt.Errorf("unable to optimize filter: %v", err)
		}

		v.columnTracker.remove(typedN.matcher)

		if fs, ok := n.(*FilterStage); ok {
			v.columnTracker.add(fs.matcher)
			if _, ok := fs.source.(*MongoSourceStage); ok {
				projSource, err := v.pushdownProject(
					v.columnTracker.getColumnsForSelectIDs(v.selectIDsInScope),
					fs.source)
				if err != nil {
					return nil, fmt.Errorf("unable to optimize filter project: %v", err)
				}
				n = NewFilterStage(projSource, fs.matcher)
			}
			v.columnTracker.remove(fs.matcher)
		}

	case *GroupByStage:
		n, err = v.visitGroupBy(typedN)
		if err != nil {
			return nil, fmt.Errorf("unable to optimize group by: %v", err)
		}

		if _, ok := typedN.source.(*MongoSourceStage); ok && n == typedN {
			projSource, err := v.pushdownProject(
				v.columnTracker.getColumnsForSelectIDs(v.selectIDsInScope),
				typedN.source)
			if err != nil {
				return nil, fmt.Errorf("unable to optimize group by project: %v", err)
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
			return nil, fmt.Errorf("unable to optimize join: %v", err)
		}

		if typedN, ok := n.(*JoinStage); ok {
			left := typedN.left
			right := typedN.right

			// For inner joins, attempt to optimize by flipping children.
			if typedN.kind == innerJoin {
				v.logger.Debugf(log.Dev, "attempting to optimize inner join via flip")
				j := NewJoinStage(typedN.kind, typedN.right, typedN.left, typedN.matcher)
				newJ, err := v.visitJoin(j)
				if err == nil {
					n = newJ
				}
			} else if typedN.kind == rightJoin {
				// For right joins, attempt to optimize using a left join.
				v.logger.Debugf(log.Dev, "attempting to optimize right join via flip")
				j := NewJoinStage(leftJoin, typedN.right, typedN.left, typedN.matcher)
				newJ, err := v.visitJoin(j)
				if err == nil {
					n = newJ
				}
			}

			if _, ok := n.(*JoinStage); ok {

				if ms, ok := left.(*MongoSourceStage); ok {
					requiredColumns := v.getRequiredColumnsForJoinSide(
						ms.dbName,
						ms.aliasNames,
						v.columnTracker.getColumnsForSelectIDs(v.selectIDsInScope))
					left, err = v.pushdownProject(requiredColumns, ms.clone())
					if err != nil {
						return nil, fmt.Errorf("unable to optimize JoinStage left project: %v", err)
					}
				}

				if ms, ok := right.(*MongoSourceStage); ok {
					requiredColumns := v.getRequiredColumnsForJoinSide(
						ms.dbName,
						ms.aliasNames,
						v.columnTracker.getColumnsForSelectIDs(v.selectIDsInScope))
					right, err = v.pushdownProject(requiredColumns, ms.clone())
					if err != nil {
						return nil, fmt.Errorf("unable to optimize JoinStage right project: %v", err)
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
			return nil, fmt.Errorf("unable to optimize limit: %v", err)
		}
	case *OrderByStage:
		n, err = v.visitOrderBy(typedN)
		if err != nil {
			return nil, fmt.Errorf("unable to optimize order by: %v", err)
		}

		if _, ok := typedN.source.(*MongoSourceStage); ok && n == typedN {
			projSource, err := v.pushdownProject(
				v.columnTracker.getColumnsForSelectIDs(v.selectIDsInScope),
				typedN.source)
			if err != nil {
				return nil, fmt.Errorf("unable to optimize order by project: %v", err)
			}
			n = NewOrderByStage(projSource, typedN.terms...)
		}

		for _, term := range typedN.terms {
			v.columnTracker.remove(term.expr)
		}
	case *ProjectStage:
		n, err = v.visitProject(typedN)
		if err != nil {
			return nil, fmt.Errorf("unable to optimize project: %v", err)
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
			return nil, fmt.Errorf("unable to optimize subquery source: %v", err)
		}

		v.selectIDsInScope = oldSelectIDsInScope
		v.tableNamesInScope = oldTableNamesInScope
		v.addSelectIDsInScope(typedN.selectID)
		v.addTableNamesInScope(typedN.aliasName)
	}

	return n, nil
}

func (v *pushDownOptimizer) canPushDown(ps PlanStage) (*MongoSourceStage, bool) {

	ms, ok := ps.(*MongoSourceStage)
	if !ok {
		return nil, false
	}

	return ms, true
}

const (
	projectPredicateFieldName = "__predicate"
)

func (v *pushDownOptimizer) extractPreUnwindMatch(mr *mappingRegistry, expr SQLExpr, unwoundPath, unwoundIndexPath string) (bson.D, bool) {
	parts := splitExpression(expr)

	var partsToMove []SQLExpr
	useElemMatch := true
	// Find any part that is composed solely of fields prefixed by the unwoundPath.
	for _, part := range parts {
		columns, err := referencedColumns(v.selectIDsInScope, part)
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

	lookupFieldName := mr.lookupFieldName
	if useElemMatch {
		lookupFieldName = func(databaseName, tableName, columnName string) (string, bool) {
			// we are going to strip the prefix off of the fieldNames because $elemMatch syntax
			// is interesting. We know this won't fail because we've already done it for all
			// combinations.
			fieldName, _ := mr.lookupFieldName(databaseName, tableName, columnName)
			return strings.TrimPrefix(fieldName, unwoundPath+"."), true
		}
	}

	t := &pushDownTranslator{
		versionAtLeast:  v.ctx.Variables().MongoDBInfo.VersionAtLeast,
		lookupFieldName: lookupFieldName,
		logger:          v.logger,
	}

	combined := combineExpressions(partsToMove)

	// We don't care about the remaining. We will still be placing a match after the unwind,
	// so anything we can't do here gets handled there anyways.
	matchBody, _ := t.TranslatePredicate(combined)
	if matchBody == nil {
		// Nothing to do.
		return nil, false
	}

	if useElemMatch {
		matchBody = bson.M{
			unwoundPath: bson.M{
				"$elemMatch": matchBody,
			},
		}
	}

	return bson.D{{"$match", matchBody}}, true
}

func (v *pushDownOptimizer) visitFilter(filter *FilterStage) (PlanStage, error) {

	ms, ok := v.canPushDown(filter.source)
	if !ok {
		return filter, nil
	}

	pipeline := append([]bson.D{}, ms.pipeline...)
	var localMatcher SQLExpr

	if value, ok := filter.matcher.(SQLValue); ok {
		// Our optimized expression has left us with just a value,
		// we can see if it matches right now. If so, we eliminate
		// the filter from the tree. Otherwise, we return an
		// operator that yields no rows.
		matches, err := Matches(value, nil)
		if err != nil {
			return nil, err
		}

		if !matches {
			return &EmptyStage{filter.Columns(), filter.Collation()}, nil
		}

		// Otherwise, the filter simply gets removed from the tree.

	} else {
		if len(pipeline) == 1 && pipeline[0][0].Name == "$unwind" {
			// Before pushing down the match, if the current pipeline contains
			// an $unwind as the first stage in the pipeline, try to place any criteria
			// for the unwound array before the $unwind using an $elemMatch. These will
			// need to still stay after the $unwind as well, but this should cut down on
			// the number of documents passing through the $unwind clause while also allowing
			// the use of an index.
			// NOTE: putting a match between a lookup and an unwind causes a server optimization
			// to get skipped.
			v.logger.Debugf(log.Dev, "attempting to add a redundant match before unwind")

			var path string
			var indexPath string
			if path, ok = pipeline[0][0].Value.(string); !ok {
				var unwindBody bson.M
				if unwindBody, ok = pipeline[0][0].Value.(bson.M); !ok {
					unwindBody = pipeline[0][0].Value.(bson.D).Map()
				}

				path = unwindBody["path"].(string)
				if ip, ok := unwindBody["includeArrayIndex"]; ok {
					indexPath = ip.(string)
				}
			}

			if preUnwindMatch, ok := v.extractPreUnwindMatch(ms.mappingRegistry, filter.matcher, path[1:], indexPath); ok {
				pipeline = append([]bson.D{preUnwindMatch}, pipeline...)
			}
		}

		var matchBody bson.M
		t := &pushDownTranslator{
			versionAtLeast:  v.ctx.Variables().MongoDBInfo.VersionAtLeast,
			lookupFieldName: ms.mappingRegistry.lookupFieldName,
			logger:          v.logger,
		}
		matchBody, localMatcher = t.TranslatePredicate(filter.matcher)
		if matchBody != nil {
			pipeline = append(pipeline, bson.D{{"$match", matchBody}})
		}

		if localMatcher != nil {
			// We have a predicate that completely or partially couldn't be handled by $match.
			// Attempt to push it down as part of a $project/$match combination.
			if predicate, ok := t.TranslateExpr(localMatcher); ok {

				// MySQL's version of truthiness is different than MongoDB's. We need to modify
				// the predicate to account for this difference. It looks, effectively, like this:
				predicate = bson.D{
					{"$let", bson.D{
						{"vars", bson.M{"predicate": predicate}},
						{"in", bson.D{
							{"$cond", []interface{}{
								bson.D{{"$or", []interface{}{
									bson.D{{"$eq", []interface{}{"$$predicate", false}}},
									bson.D{{"$eq", []interface{}{"$$predicate", 0}}},
									bson.D{{"$eq", []interface{}{"$$predicate", "0"}}},
									bson.D{{"$eq", []interface{}{"$$predicate", "-0"}}},
									bson.D{{"$eq", []interface{}{"$$predicate", "0.0"}}},
									bson.D{{"$eq", []interface{}{"$$predicate", "-0.0"}}},
									bson.D{{"$eq", []interface{}{"$$predicate", nil}}},
								}}},
								false,
								true,
							}},
						}},
					}},
				}
				fieldName := v.uniqueFieldName(projectPredicateFieldName, ms.mappingRegistry)
				stageBody := bson.M{
					fieldName: predicate,
				}
				predicateEvaluationStage := v.buildAddFieldsOrProject(stageBody, []string{}, ms.mappingRegistry)
				pipeline = append(pipeline, predicateEvaluationStage,
					bson.D{{"$match", bson.M{fieldName: true}}})

				localMatcher = nil
			}

			if matchBody == nil && localMatcher != nil {
				// No pieces of the matcher are able to be pushed down,
				// so there is no change in the operator tree.
				v.logger.Debugf(log.Dev, "cannot push down filter expression: %v", filter.matcher.String())
				return filter, nil
			}
		}
	}

	// If we end up here, it's because we have messed with the pipeline
	// in the current table scan operator, so we need to reconstruct the
	// operator nodes.
	ms = ms.clone()
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
	groupID             = "_id"
	groupDistinctPrefix = "distinct "
	groupTempTable      = ""
	groupTempDB         = ""
)

// visitGroupBy works by using a visitor to systematically visit and replace certain SQLExpr while generating
// $group and $project stages for the aggregation pipeline.
func (v *pushDownOptimizer) visitGroupBy(gb *GroupByStage) (PlanStage, error) {

	ms, ok := v.canPushDown(gb.source)
	if !ok {
		return gb, nil
	}

	pipeline := ms.pipeline

	// 1. Translate keys.
	keys, err := v.translateGroupByKeys(gb.keys, ms.mappingRegistry.lookupFieldName)
	if err != nil {
		v.logger.Warnf(log.Dev, "cannot translate group by keys: %v", err)
		return gb, nil
	}

	// 2. Translate aggregations.
	result, err := v.translateGroupByAggregates(gb.keys, gb.projectedColumns, ms.mappingRegistry.lookupFieldName)
	if err != nil {
		v.logger.Warnf(log.Dev, "cannot translate group by aggregates: %v", err)
		return gb, nil
	}

	result.group[groupID] = keys
	pipeline = append(pipeline, bson.D{{"$group", result.group}})

	var mr *mappingRegistry
	// 3. Translate the final project if necessary.
	if result.requiresTwoSteps {
		project, err := v.translateGroupByProject(result.mappedProjectedColumns, result.mappingRegistry.lookupFieldName)
		if err != nil {
			v.logger.Warnf(log.Dev, "cannot translate group by project: %v", err)
			return gb, nil
		}
		pipeline = append(pipeline, bson.D{{"$project", project}})

		// 4. Fix up the MongoSourceStage - None of the current registrations in mappingRegistry are valid any longer.
		// We need to clear them out and re-register the new columns.
		mr = &mappingRegistry{}
		for _, mappedProjectedColumn := range result.mappedProjectedColumns {
			// at this point, our project has all the stringified names of the select expressions, so we need to re-map them
			// each column to its new MongoDB name. This process is what makes the push-down transparent to subsequent operators
			// in the tree that either haven't yet been pushed down, or cannot be. Either way, the output of a push-down must be
			// exactly the same as the output of a non-pushed-down group.
			mr.addColumn(mappedProjectedColumn.projectedColumn.Column)
			mr.registerMapping(
				mappedProjectedColumn.projectedColumn.Database,
				mappedProjectedColumn.projectedColumn.Table,
				mappedProjectedColumn.projectedColumn.Name,
				sanitizeFieldName(mappedProjectedColumn.projectedColumn.Expr.String()),
			)
		}
	} else {
		mr = &mappingRegistry{}
		for _, mpc := range result.mappedProjectedColumns {
			// At this point, we pushed down a group, but we still need to map the projected column name
			// to the expressions that were pushed down. We know that all the pushed down exprs are now
			// columns, so we simply look up the original field name and use that.
			columnExpr, ok := mpc.expr.(SQLColumnExpr)
			if !ok {
				v.logger.Warnf(log.Dev, "expr was not a column")
				return gb, nil
			}
			mr.addColumn(mpc.projectedColumn.Column)
			fieldName, ok := result.mappingRegistry.lookupFieldName(columnExpr.databaseName, columnExpr.tableName, columnExpr.columnName)
			if !ok {
				v.logger.Warnf(log.Dev, "could not find translated aggregate's field name")
				return gb, nil
			}
			mr.registerMapping(
				mpc.projectedColumn.Database,
				mpc.projectedColumn.Table,
				mpc.projectedColumn.Name,
				fieldName,
			)
		}
	}

	ms = ms.clone()
	ms.pipeline = pipeline
	ms.mappingRegistry = mr

	return ms, nil
}

// translateGroupByKeys takes the key expressions and builds an _id document. All keys, even single keys,
// will be nested underneath the '_id' field. In addition, each field's name will be the stringified
// version of its SQLExpr.
// For example, 'select a, b from foo group by c' will build an id document that looks like this:
//      _id: { foo_DOT_c: "$c" }
//
// Likewise, multiple columns will build something similar.
// For example, 'select a, b from foo group by c,d' will build an id document that looks like this:
//      _id: { foo_DOT_c: "$c", foo_DOT_d: "$c" }
//
// Finally, anything other than a column will still build similarly.
// For example, 'select a, b from foo group by c+d' will build an id document that looks like this:
//      _id: { "foo_DOT_c+foo_DOT_d": { $add: ["$c", "$d"] } }
//
// All projected names are the fully qualified name from SQL, ignoring the mongodb name except for when
// referencing the underlying field.
func (v *pushDownOptimizer) translateGroupByKeys(keys []SQLExpr, lookupFieldName fieldNameLookup) (bson.D, error) {

	keyDocumentElements := bson.D{}

	t := &pushDownTranslator{
		versionAtLeast:  v.ctx.Variables().MongoDBInfo.VersionAtLeast,
		lookupFieldName: lookupFieldName,
		logger:          v.logger,
	}

	for _, key := range keys {
		translatedKey, ok := t.TranslateExpr(key)
		if !ok {
			return nil, fmt.Errorf("could not translate '%v'", key.String())
		}

		keyDocumentElements = append(keyDocumentElements, bson.DocElem{sanitizeFieldName(key.String()), translatedKey})
	}

	return keyDocumentElements, nil
}

// translateGroupByAggregatesResult is just a holder for the results from translateGroupByAggregates.
type translateGroupByAggregatesResult struct {
	group                  bson.M
	mappedProjectedColumns []*mappedProjectedColumn
	mappingRegistry        *mappingRegistry
	requiresTwoSteps       bool
}

type mappedProjectedColumn struct {
	projectedColumn ProjectedColumn
	expr            SQLExpr
}

// translateGroupByAggregates takes the key expressions and the select expressions and builds a
// $group stage. It does this by employing a visitor that walks each of the select expressions individually
// and, depending on the type of expression, builds a full solution or a partial solution to accomplishing
// the goal. For example, the query 'select sum(a) from foo' can be fully realized with a single $group, where
// as 'select sum(distinct a) from foo' requires a $group which adds 'a' to a set ($addToSet) and a subsequent
// $project which then does a $sum on the set created in the $group. Currently, we always build the $project
// whether it's necessary or not.
//
// In addition to generating the group, new expressions and a mapping registry are also returned that
// account for the namings and resulting expressions from the $group. For example, 'select sum(a) from foo'
// will take in a SQLAggFunctionExpr for the 'sum(a)' and return a SQLFieldExpr that represents the already
// summed 'a' field such that the subsequent $project doesn't need to care about the aggregation. In the same way,
// sum(distinct a) will take in a SQLAggFunctionExpr which refers to the column 'a' and return a new SQLAggFunctionExpr
// which refers to the newly created $addToSet field called 'distinct foo_DOT_a'. This way, the subsequent $project
// now has the correct reference to the field name in the $group.
func (v *pushDownOptimizer) translateGroupByAggregates(keys []SQLExpr, projectedColumns ProjectedColumns, lookupFieldName fieldNameLookup) (*translateGroupByAggregatesResult, error) {

	// For example, in "select a + sum(b) from bar group by a", we should not create
	// an aggregate for a because it's part of the key.
	isGroupKey := func(expr SQLExpr) bool {
		exprString := expr.String()
		for _, key := range keys {
			if exprString == key.String() {
				return true
			}
		}

		return false
	}

	// This represents all the expressions that should be passed on to $project such that translateGroupByProject
	// is able to do its work without redoing a bunch of the conditionals and special casing here.
	mappedProjectedColumns := []*mappedProjectedColumn{}

	// translator will "accumulate" all the group fields. Below, we iterate over each select expressions, which
	// account for all the fields that need to be present in the $group.
	translator := &groupByAggregateTranslator{bson.M{}, v.ctx, isGroupKey, lookupFieldName, &mappingRegistry{}, false, v.logger}

	for _, projectedColumn := range projectedColumns {

		newExpr, err := translator.visit(projectedColumn.Expr)
		if err != nil {
			return nil, err
		}

		mappedProjectedColumn := &mappedProjectedColumn{
			expr:            newExpr.(SQLExpr),
			projectedColumn: projectedColumn,
		}

		mappedProjectedColumns = append(mappedProjectedColumns, mappedProjectedColumn)
	}

	return &translateGroupByAggregatesResult{translator.group, mappedProjectedColumns, translator.mappingRegistry, translator.requiresTwoSteps}, nil
}

type groupByAggregateTranslator struct {
	group            bson.M
	ctx              ConnectionCtx
	isGroupKey       func(SQLExpr) bool
	lookupFieldName  fieldNameLookup
	mappingRegistry  *mappingRegistry
	requiresTwoSteps bool
	logger           *log.Logger
}

const (
	sumAggregateCountSuffix = "_count"
)

// Visit recursively visits each expression in the tree, adds the relavent $group entries, and returns
// an expression that can be used to build a subsequent $project.
func (v *groupByAggregateTranslator) visit(n node) (node, error) {
	t := &pushDownTranslator{
		versionAtLeast:  v.ctx.Variables().MongoDBInfo.VersionAtLeast,
		lookupFieldName: v.lookupFieldName,
		logger:          v.logger,
	}
	switch typedN := n.(type) {
	case SQLColumnExpr:
		fieldName, ok := v.lookupFieldName(typedN.databaseName, typedN.tableName, typedN.columnName)
		if !ok {
			return nil, fmt.Errorf("could not map %v.%v to a field", typedN.tableName, typedN.columnName)
		}
		if !v.isGroupKey(typedN) {
			// Since it's not an aggregation function, this implies that it takes the first value of the column.
			// So project the field, and register the mapping.
			v.group[sanitizeFieldName(typedN.String())] = bson.M{"$first": getProjectedFieldName(fieldName, typedN.Type())}
			v.mappingRegistry.registerMapping(typedN.databaseName, typedN.tableName, typedN.columnName, sanitizeFieldName(typedN.String()))
		} else {
			// The _id is added to the $group in translateGroupByKeys. We will only be here if the user has also projected
			// the group key, in which we'll need this to look it up in translateGroupByProject under its name. Hence, all
			// we need to do is register the mapping.
			v.mappingRegistry.registerMapping(typedN.databaseName, typedN.tableName, typedN.columnName, groupID+"."+sanitizeFieldName(typedN.String()))
		}
		return typedN, nil
	case *SQLAggFunctionExpr:
		var newExpr SQLExpr
		if typedN.Distinct {
			// Distinct aggregation expressions are two-step aggregations. In the $group stage, we use $addToSet
			// to handle whatever the distinct expression is, which could be a simply field name, or something
			// more complex like a mathematical computation. We don't care either way, and TranslateExpr handles
			// generating the correct thing. Once this is done, we create a new SQLAggFunctionExpr whose argument
			// maps to the newly named field containing the set of values to perform the aggregation on.
			v.requiresTwoSteps = true
			trans, ok := t.TranslateExpr(typedN.Exprs[0])
			if !ok {
				return nil, fmt.Errorf("could not translate '%v'", typedN.String())
			}
			fieldName := groupDistinctPrefix + sanitizeFieldName(typedN.Exprs[0].String())
			newExpr = &SQLAggFunctionExpr{
				Name: typedN.Name,
				Exprs: []SQLExpr{
					NewSQLColumnExpr(
						0,
						groupTempDB,
						groupTempTable,
						fieldName,
						typedN.Type(),
						schema.MongoNone,
					),
				},
			}
			v.group[fieldName] = bson.M{"$addToSet": trans}
			v.mappingRegistry.registerMapping(groupTempDB, groupTempTable, fieldName, fieldName)
		} else {
			// Non-distinct aggregation functions are one-step aggregations that can be performed completely by the
			// $group. Here, we simply build the correct aggregation function for $group and create a new expression
			// that references the resulting field. There isn't a need to keep the aggregation function around because
			// the aggregation has already been done and all that's left is a field for $project to reference.

			// Count is special since MongoDB doesn't have a native count function. Instead, we use $sum. However,
			// depending on the form that count takes, we need to different things. 'count(*)' is just {$sum: 1},
			// but 'count(a)' requires that we not count nulls, undefineds, and missing fields. Hence, it becomes
			// a $sum with $cond and $ifNull.
			var trans interface{}
			var ok bool
			if typedN.Name == countAggregateName && typedN.Exprs[0] == SQLVarchar("*") {
				trans = bson.M{"$sum": 1}
			} else if typedN.Name == countAggregateName {
				trans, ok = t.TranslateExpr(typedN.Exprs[0])
				if !ok {
					return nil, fmt.Errorf("could not translate '%v'", typedN.Exprs[0].String())
				}

				trans = getCountAggregation(trans)
			} else {
				trans, ok = t.TranslateExpr(typedN)
				if !ok {
					return nil, fmt.Errorf("could not translate '%v'", typedN.String())
				}
			}

			fieldName := sanitizeFieldName(typedN.String())
			v.group[fieldName] = trans
			v.mappingRegistry.registerMapping(groupTempDB, groupTempTable, fieldName, fieldName)

			if typedN.Name == sumAggregateName {
				// Summing a column with all nulls should result in a null sum. However, MongoDB
				// returns 0. So, we'll add in an arbitrary count operator to count the number
				// of non-nulls and, in the following $project, we'll check this to know whether
				// or not to use the sum or to use null.
				v.requiresTwoSteps = true
				countTrans, ok := t.TranslateExpr(typedN.Exprs[0])
				if !ok {
					return nil, fmt.Errorf("could not translate '%v'", typedN.Exprs[0].String())
				}
				countFieldName := sanitizeFieldName(typedN.String() + sumAggregateCountSuffix)
				v.group[countFieldName] = getCountAggregation(countTrans)
				v.mappingRegistry.registerMapping(groupTempDB, groupTempTable, countFieldName, countFieldName)

				newExpr = NewIfScalarFunctionExpr(
					NewSQLColumnExpr(0, groupTempDB, groupTempTable, countFieldName, schema.SQLInt64, schema.MongoNone),
					NewSQLColumnExpr(0, groupTempDB, groupTempTable, fieldName, typedN.Type(), schema.MongoNone),
					SQLNull,
				)
			} else {
				newExpr = NewSQLColumnExpr(0, groupTempDB, groupTempTable, fieldName, typedN.Type(), schema.MongoNone)
			}

		}

		return newExpr, nil

	case SQLExpr:
		if v.isGroupKey(typedN) {
			// The _id is added to the $group in translateGroupByKeys. We will only be here if the user has also projected
			// the group key, in which we'll need this to look it up in translateGroupByProject under its name. In this,
			// we need to create a new expr that is simply a field pointing at the nested identifier and register that
			// mapping.
			fieldName := sanitizeFieldName(typedN.String())
			newExpr := NewSQLColumnExpr(0, groupTempDB, groupTempTable, fieldName, typedN.Type(), schema.MongoNone)
			v.mappingRegistry.registerMapping(groupTempDB, groupTempTable, fieldName, groupID+"."+fieldName)
			return newExpr, nil
		}

		// We'll only get here for two-step expressions. An example is
		// 'select a + b from foo group by a' or 'select b + sum(c) from foo group by a'. In this case,
		// we'll descend into the tree recursively which will build up the $group for the necessary pieces.
		// Finally, return the now changed expression such that $project can act on them appropriately.
		v.requiresTwoSteps = true
		return walk(v, n)
	default:
		// PlanStages will end up here and we don't need to do anything in them.
		return n, nil
	}
}

// getCountAggregation is used when a non-star count is used. {sum:1} isn't valid because nulls, undefineds, and
// missing values should not be part of result. Because MongoDB doesn't have a proper count operator, we have to
// do some extra checks along the way.
func getCountAggregation(expr interface{}) bson.M {
	return bson.M{
		"$sum": bson.M{
			"$cond": []interface{}{
				bson.M{
					"$eq": []interface{}{
						bson.M{
							"$ifNull": []interface{}{
								expr,
								nil,
							},
						},
						nil,
					},
				},
				0,
				1,
			},
		},
	}
}

// translateGroupByProject takes the expressions and builds a $project. All the hard work was done in
// translateGroupByAggregates, so this is simply a process of either adding a field to the $project, or
// completing two-step aggregations. Two-step aggregations that needs completing are expressions like
// 'sum(distinct a)' or 'a + b' where b was part of the group key.
func (v *pushDownOptimizer) translateGroupByProject(mappedProjectedColumns []*mappedProjectedColumn, lookupFieldName fieldNameLookup) (bson.M, error) {
	project := bson.M{groupID: 0}

	t := &pushDownTranslator{
		versionAtLeast:  v.ctx.Variables().MongoDBInfo.VersionAtLeast,
		lookupFieldName: lookupFieldName,
		logger:          v.logger,
	}

	for _, mappedProjectedColumn := range mappedProjectedColumns {

		mappedName := sanitizeFieldName(mappedProjectedColumn.projectedColumn.Expr.String())
		switch typedE := mappedProjectedColumn.expr.(type) {
		case SQLColumnExpr:
			// Any one-step aggregations will end up here as they were fully performed in the $group. So, simple
			// column references ('select a') and simple aggregations ('select sum(a)').
			fieldName, ok := lookupFieldName(typedE.databaseName, typedE.tableName, typedE.columnName)
			if !ok {
				return nil, fmt.Errorf("unable to get a field name for %v.%v", typedE.tableName, typedE.columnName)
			}

			project[mappedName] = "$" + fieldName
		default:
			// Any two-step aggregations will end up here to complete the second step.
			trans, ok := t.TranslateExpr(mappedProjectedColumn.expr)
			if !ok {
				return nil, fmt.Errorf("unable to translate '%v'", mappedProjectedColumn.expr.String())
			}
			project[mappedName] = trans
		}
	}

	return project, nil
}

const (
	joinedFieldNamePrefix    = "__joined_"
	leftJoinExcludeFieldName = "__leftjoin_exclude"
)

// buildRemainingPredicateForLeftJoin will return 2 items; first a $project to
// put before the unwind, and a $match to put after the unwind. The remaining
// predicate SQLExpr is used to build the $project (or $addFields) and the
// $match. asField is the name of the array field to check. foreignIndex is
// the name of the foreign pipeline unwind, if any (passing an empty string is
// safe if there is no foreign unwind), because we cannot build a predicate
// using the foreignIndex because it creates a circular dependency in the
// pipeline (the foreign unwind must go afther the $project/$addFields which
// must use the field created by the foreign unwind)
func (v *pushDownOptimizer) buildRemainingPredicateForLeftJoin(leftMappingRegistry, combinedMappingRegistry *mappingRegistry, remainingPredicate SQLExpr, asField, foreignIndex string, preserveIndex bool) (bson.D, bson.D, bool) {
	registries := []*mappingRegistry{combinedMappingRegistry}
	fixedLookupFieldName := func(db, tbl, col string) (string, bool) {
		// Join predicates should always be based on the original field, rather than the added
		// fields that have been added for left joins. The only way the value being NULL could
		// matter is if <=> or is NULL is in the predicate, and even if that is the case,
		// it would already be NULL from being a left join, anyway. If it is instead <> NULL
		// the predicate is essentially a no-op
		fieldName, _, _, ok := v.lookupSQLColumnForJoin(db, tbl, col, registries)
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
			v.logger.Debugf(log.Dev, logPrefix+": cannot use foreign unwind index: %q in left join criteria because use occurs before foreign unwind moving on...", foreignIndex)

			return "", false
		}

		// Inside a $filter and $map (which use the result of this function), columns with the asField
		// prefix will have their prefix renamed. As such, we need to intercept this call and handle
		// that translation early. For instance, if the asField as $b.child and the field ends up as
		// $b.child.myField, then the result will be $$this.myField.
		// NOTE: it is important to use asField + "." as the prefix, because otherwise we will
		// end up renaming something we generate in unwinds like c_idx to this_idx... which is wrong
		// We then also need the condition where fieldName == asField, since prefix will no longer
		// catch it, since we have added the "."
		if fieldName == asField || strings.HasPrefix(fieldName, asField+".") {
			fieldName = strings.Replace(fieldName, asField, "$this", 1)
		}

		return fieldName, true
	}

	t := &pushDownTranslator{
		versionAtLeast:  v.ctx.Variables().MongoDBInfo.VersionAtLeast,
		lookupFieldName: fixedLookupFieldName,
		logger:          v.logger,
	}

	ifPart, ok := t.TranslateExpr(remainingPredicate)
	if !ok {
		v.logger.Warnf(log.Dev, "cannot translate remaining left join predicate %#v", remainingPredicate)
		return nil, nil, false
	}

	var projectBody bson.M
	var match bson.D
	if preserveIndex {
		dolAsField := "$" + asField
		// This is interesting. First, we are going to create variable that marks every item in the array that should
		// be excluded. Using that variable, we'll then create a condition. If we filter all the items out that
		// should be excluded and end up with 0 items, we set the field to an empty array. Otherwise, we keep the array
		// with the marked items and use a match after the unwind to get rid of the rows that don't belong.
		// The reason we have to do this is because, even when no items from the "right" side of a join match, we still need to
		// include the left side one time. However, we can't just eliminate the non-matching array items now (using $filter)
		// because we need to maintain the array index of the items that do match.
		projectBody = bson.M{
			asField: bson.M{
				"$let": bson.M{
					"vars": bson.M{
						"mapped": bson.M{
							"$map": bson.M{
								"input": bson.M{
									"$cond": bson.M{
										"if":   bson.M{"$isArray": dolAsField},
										"then": dolAsField,
										"else": bson.M{"$cond": bson.M{
											// It is very important that we map null and missing to []
											// rather than [null] because [null] is semantically different:
											// When we form a child table with {..., x : [null], ...}
											// we have one row with one primary key x_idx = 0 with null
											// as a value. When we form a child table with [], null, or missing,
											// we produce 0 rows. Mapping null (or missing) to [null] breaks
											// this semantics, and ruins the fields added for self-join
											// optimized left-joins
											"if":   bson.M{"$lte": []interface{}{dolAsField, nil}},
											"then": []interface{}{},
											"else": []interface{}{dolAsField},
										},
										},
									},
								},
								"as": "this",
								"in": bson.M{
									"$cond": bson.M{
										"if":   ifPart,
										"then": "$$this",
										"else": bson.M{
											leftJoinExcludeFieldName: bson.M{"$literal": true},
										},
									},
								},
							},
						},
					},
					"in": bson.M{
						"$cond": bson.M{
							"if": bson.M{
								"$gt": []interface{}{
									bson.M{
										"$size": bson.M{
											"$filter": bson.M{
												"input": "$$mapped",
												"as":    "this",
												"cond": bson.M{
													"$ne": []interface{}{
														"$$this." + leftJoinExcludeFieldName,
														true,
													},
												},
											},
										},
									},
									0,
								},
							},
							"then": "$$mapped",
							"else": []interface{}{},
						},
					},
				},
			},
		}
		predicateExclusionField := asField + "." + leftJoinExcludeFieldName
		match = bson.D{{"$match", bson.M{predicateExclusionField: bson.M{"$ne": true}}}}
	} else {
		// In this case, we can simply filter the array because we don't care about preserving the index.
		// If the predicate doesn't pass, then we set the 'as' field to nil.
		projectBody = bson.M{
			asField: bson.M{
				"$filter": bson.M{
					"input": "$" + asField,
					"as":    "this",
					"cond":  ifPart,
				},
			},
		}
	}
	return v.buildAddFieldsOrProject(projectBody, []string{asField}, combinedMappingRegistry), match, true
}

func (v *pushDownOptimizer) selfJoinOptimizeTables(msLocal, msForeign *MongoSourceStage, join *JoinStage) (PlanStage, error) {
	var foreignRegistryBackup *mappingRegistry
	// If we fail to translate a left join predicate later, we will need to restore this
	// if, instead this is an inner join, there is nothing to worry about
	if join.kind == leftJoin {
		foreignRegistryBackup = msForeign.mappingRegistry.copy()
	}

	newPipeline, err := v.selfJoinOptimizePipeline(msLocal, msForeign, join.kind)
	if err != nil {
		v.logger.Warnf(log.Dev, "cannot self-join optimize pipelines: %v", err)
		return nil, nil
	}

	ms := msLocal.clone()
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
					newMappingRegistry.registerMapping(database, tableName, columnName,
						fieldName)
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
		if join.kind == innerJoin || join.kind == straightJoin {
			// This isn't a left join, so we do not have to worry about
			// failing to build the left-join predicate and can copy
			// back the newMappingRegistry and newPipeline.stages
			ms.mappingRegistry, ms.pipeline = newMappingRegistry, newPipeline.stages
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
		localUnwinds, totalUnwinds := getPipelineUnwindFields(msLocal.pipeline), getPipelineUnwindFields(newPipeline.stages)
		unwindSuffix, _ := getUnwindSuffix(localUnwinds, totalUnwinds)
		insertionPoint := 0
		if len(localUnwinds) != 0 {
			insertionPointPath := localUnwinds[len(localUnwinds)-1].path
			insertionPointUnwind, ok := findUnwindForPath(totalUnwinds, insertionPointPath)
			if !ok {
				panic(fmt.Sprintf("could not find unwind for path %v in pipeline %v, "+
					"this should never happen)",
					insertionPointPath, newPipeline.stages))
			}
			insertionPoint = insertionPointUnwind.stageNumber + 1
		}

		project, match, ok := v.buildRemainingPredicateForLeftJoin(
			msLocal.mappingRegistry,
			newMappingRegistry,
			remainingPredicate,
			strings.Replace(unwindSuffix[0].path, "$", "", 1),
			unwindSuffix[0].index,
			true,
		)

		if !ok {
			// We failed to translate, make sure to restore the foriegn
			// mapping registry
			msForeign.mappingRegistry = foreignRegistryBackup
			return join, nil
		}

		if match != nil {
			newPipeline.stages = append(newPipeline.stages, match)
		}

		// Insert project after the first.
		newPipeline.stages = insertPipelineStageAt(newPipeline.stages, project, insertionPoint)
	}

	ms.mappingRegistry, ms.pipeline = newMappingRegistry, newPipeline.stages
	return ms, nil
}

func (v *pushDownOptimizer) visitJoin(join *JoinStage) (PlanStage, error) {
	if join.matcher == nil {
		v.logger.Warnf(log.Dev, "cannot push down join stage, matcher is nil")
		return join, nil
	}

	// 1. the join type must be usable. MongoDB can only do an inner join and a left outer join.
	// While we can flip a right outer join to a left outer join, because we don't have information
	// about the target collections, we need to indicate to users that the right-side may NOT be
	// sharded. Flipping these around would make this difficult for a user to fix. Instead, we'll
	// just tell them to make their right outer joins -> left outer joins.
	var localSource, foreignSource PlanStage
	var kind joinKind

	switch join.kind {
	case innerJoin:
		localSource = join.left
		foreignSource = join.right
		kind = innerJoin
	case leftJoin:
		localSource = join.left
		foreignSource = join.right
		kind = leftJoin
	case straightJoin:
		localSource = join.left
		foreignSource = join.right
		kind = straightJoin
	default:
		v.logger.Warnf(log.Dev, "cannot push down %v", join.kind)
		return join, nil
	}

	// 2. we have to be able to push both down and the foreign MongoSource
	// operator must have nothing in its pipeline.
	msLocal, ok := localSource.(*MongoSourceStage)
	if !ok {
		return join, nil
	}

	msForeign, ok := foreignSource.(*MongoSourceStage)
	if !ok {
		return join, nil
	}

	if msLocal.dbName != msForeign.dbName {
		v.logger.Warnf(log.Dev, "cannot optimize join stage, local database is different from foreign database")
		return join, nil
	}

	for i, collection := range msForeign.collectionNames {
		isSharded, _ := msForeign.isShardedCollection[collection]
		if isSharded {
			v.logger.Warnf(log.Dev, "unable to translate join "+
				"stage to lookup: foreign table %q is sharded", msForeign.tableNames[i])
			return join, nil
		}
	}
	// Before attempting the self-join optimization, check that the
	// underlying collection is the same for both tables and that the join
	// criteria holds the primary key for both.
	if v.canSelfJoinTables(v.logger, msLocal, msForeign, join.matcher, join.kind) {
		ms, err := v.selfJoinOptimizeTables(msLocal, msForeign, join)
		//for tables in different databases, it is not possible to push down since this isn't
		//supported in MongoDB, just do the join in memory

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

	v.logger.Debugf(log.Dev, "attempting to translate join stage "+
		"to lookup")

	lenForeignPipeline := len(msForeign.pipeline)
	foreignHasUnwind := false

	if lenForeignPipeline > 0 {
		var unwindInterface interface{}
		unwindInterface, foreignHasUnwind = msForeign.pipeline[0].Map()["$unwind"]
		if !foreignHasUnwind || lenForeignPipeline > 1 {
			v.logger.Warnf(log.Dev, "unable to translate join "+
				"stage to lookup: foreign table has pipeline: %#v",
				msForeign.pipeline)
			return join, nil
		}
		unwind := unwindInterface.(bson.D)
		// These registries will be needed in the loop over
		// join exprs below.
		registries := []*mappingRegistry{
			msLocal.mappingRegistry,
			msForeign.mappingRegistry,
		}
		// Check to make sure the single unwind in the foreign pipeline
		// doesn't have an array index created by the unwind in its
		// join condition, otherwise we build an impossible $lookup
		// and an empty return set.
		if unwindIndexName, foreignUnwindHasIndex := unwind.Map()["includeArrayIndex"]; foreignUnwindHasIndex {
			exprs := splitExpression(join.matcher)
			for _, expr := range exprs {
				// Ignore non-equalExpr join conditions, since
				// they will be handled after any foreign
				// $unwinds as a $match or remaining left join predicate
				// (see buildRemainingLeftJoinPredicate) and thus not
				// cause any issues.
				if equalExpr, ok := expr.(*SQLEqualsExpr); ok {
					column1, _ := equalExpr.left.(SQLColumnExpr)
					column2, _ := equalExpr.right.(SQLColumnExpr)
					// It's possible that someone could use
					// the foreign table on either or both
					// sides of the join equivalence, so we
					// can't use else here.
					if containsString(msForeign.aliasNames, column1.tableName) {
						_, columnName, _, _ := v.lookupSQLColumnForJoin(column1.databaseName, column1.tableName, column1.columnName, registries)

						if columnName == unwindIndexName {
							v.logger.Debugf(log.Dev, "$lookup translation: cannot use foreign unwind index: %q in equality criteria because use in $lookup occurs before foreign unwind, moving on...", unwindIndexName)
							return join, nil
						}
					}
					if containsString(msForeign.aliasNames, column2.tableName) {
						_, columnName, _, _ := v.lookupSQLColumnForJoin(column2.databaseName, column2.tableName, column2.columnName, registries)

						if columnName == unwindIndexName {
							v.logger.Debugf(log.Dev, "$lookup translation: cannot use foreign unwind index: %q in equality criteria, because use in $lookup occurs before foreign unwind, moving on...", unwindIndexName)
							return join, nil
						}
					}
				}
			}
		}
	}

	// 3. Find the local column and the foreign column.
	lookupInfo, err := getLocalAndForeignColumns(msLocal, msForeign,
		join.matcher)
	if err != nil {
		v.logger.Warnf(log.Dev, "unable to translate join "+
			"stage to lookup: %v", err)
		return join, nil
	}

	// Prevent join pushdown when UUID subtype 3 encoding is different.
	localMongoType := lookupInfo.localColumn.columnType.MongoType
	foreignMongoType := lookupInfo.foreignColumn.columnType.MongoType

	if isUUID(localMongoType) && isUUID(foreignMongoType) {
		if localMongoType != foreignMongoType {
			v.logger.Warnf(log.Dev, "unable to translate join "+
				"stage to lookup: found different criteria UUID - %v "+
				"and %v", localMongoType, foreignMongoType)
			return join, nil
		}
	}

	// 4. Get the lookup fields.
	localFieldName, ok := msLocal.mappingRegistry.lookupFieldName(
		lookupInfo.localColumn.databaseName,
		lookupInfo.localColumn.tableName,
		lookupInfo.localColumn.columnName)
	if !ok {
		v.logger.Warnf(log.Dev, "cannot find referenced local join "+
			"column %#v in lookup", lookupInfo.localColumn)
		return join, nil
	}

	foreignFieldName, ok := msForeign.mappingRegistry.lookupFieldName(
		lookupInfo.foreignColumn.databaseName,
		lookupInfo.foreignColumn.tableName,
		lookupInfo.foreignColumn.columnName)
	if !ok {
		v.logger.Warnf(log.Dev, "cannot find referenced foreign join "+
			"column %#v in lookup", lookupInfo.foreignColumn)
		return join, nil
	}

	asField := sanitizeFieldName(joinedFieldNamePrefix + msForeign.aliasNames[0])

	// 5. Compute all the mappings from the msForeign mapping registry
	// to be nested under the 'asField' we used above.
	newMappingRegistry := msLocal.mappingRegistry.copy()

	newMappingRegistry.columns = append(newMappingRegistry.columns,
		msForeign.mappingRegistry.columns...)
	if msForeign.mappingRegistry.fields != nil {
		for database, tables := range msForeign.mappingRegistry.fields {
			for tableName, columns := range tables {
				for columnName, fieldName := range columns {
					newMappingRegistry.registerMapping(database, tableName, columnName,
						asField+"."+fieldName)
				}
			}
		}
	}

	// 6. Build the pipeline.
	pipeline := msLocal.pipeline

	if kind == innerJoin || kind == straightJoin {
		// Because MongoDB does not compare nulls in the same way as MySQL, we need an extra
		// $match to ensure account for this incompatibility. Effectively, we eliminate the
		// left hand document when the left hand field is null.
		match := bson.M{localFieldName: bson.M{"$ne": nil}}
		pipeline = append(pipeline, bson.D{{"$match", match}})
	}

	lookup := bson.M{
		"from":         msForeign.collectionNames[0],
		"localField":   localFieldName,
		"foreignField": foreignFieldName,
		"as":           asField,
	}

	pipeline = append(pipeline, bson.D{{"$lookup", lookup}})

	if kind == leftJoin {
		// Because MongoDB does not compare nulls in the same way as MySQL, we need an extra
		// $project to account for this incompatibility. Effectively, when our left
		// hand field is null, we'll empty the joined results prior to unwinding.
		project := bson.M{}

		// Enumerate all the local fields.
		for _, c := range msLocal.mappingRegistry.columns {
			fieldName, ok := msLocal.mappingRegistry.lookupFieldName(
				c.Database, c.Table, c.Name)
			if !ok {
				panic(fmt.Sprintf("Unable to find field mapping for column %s.%s.%s. This "+
					"should never happen.", c.Database, c.Table, c.Name))
			}
			project[fieldName] = 1
		}

		project[asField] = bson.M{"$cond": []interface{}{
			bson.M{"$eq": []interface{}{
				bson.M{"$ifNull": []interface{}{"$" + localFieldName, nil}},
				nil,
			}},
			bson.M{"$literal": []interface{}{}},
			"$" + asField,
		}}

		pipeline = append(pipeline, bson.D{{"$project", project}})
	}

	unwind := bson.D{{
		"$unwind", bson.M{
			"path": "$" + asField,
			"preserveNullAndEmptyArrays": kind == leftJoin,
		},
	}}

	lookupOnArrayField := false
	oldForeignIndex := ""

	if foreignHasUnwind {
		foreignMapped := msForeign.pipeline[0].Map()["$unwind"].(bson.D).Map()
		oldForeignPath := fmt.Sprintf("%v", foreignMapped["path"])
		oldForeignIndex = asField + "." + fmt.Sprintf("%v", foreignMapped["includeArrayIndex"])
		lookupOnArrayField = strings.Split(foreignFieldName, ".")[0] == oldForeignPath[1:]
	}

	if lookupInfo.remainingPredicate != nil && kind == leftJoin {
		if lookupOnArrayField && len(strings.Split(foreignFieldName, ".")) > 1 {
			v.logger.Warnf(log.Dev, "unable to translate left join "+
				"stage to lookup: lookup on nested array field")
			return join, nil
		}

		project, match, ok := v.buildRemainingPredicateForLeftJoin(
			msLocal.mappingRegistry,
			newMappingRegistry,
			lookupInfo.remainingPredicate,
			asField,
			oldForeignIndex,
			false,
		)
		if !ok {
			return join, nil
		}

		pipeline = append(pipeline, project, unwind)

		if match != nil {
			pipeline = append(pipeline, match)
		}
	} else {
		pipeline = append(pipeline, unwind)
	}

	// This handles merging foreign tables referenced in joins
	// that contain a single $unwind pipeline stage.
	if foreignHasUnwind {
		foreignMapped := msForeign.pipeline[0].Map()["$unwind"].(bson.D).Map()
		path := fmt.Sprintf("%v", foreignMapped["path"])

		// Strip dollar sign prefix, and prepend with asField.
		if path != "" {
			path = fmt.Sprintf("$%v.%v", asField, path[1:])
		} else {
			v.logger.Warnf(log.Dev, "empty $unwind path specification")
			return join, nil
		}

		// For left joins, preserve null and empty arrays to ensure
		// that every local pipeline record gets returned.
		idx := fmt.Sprintf("%v.%v", asField, foreignMapped["includeArrayIndex"])
		unwind := bson.D{
			{"path", path},
			{"includeArrayIndex", idx},
		}

		unwind = append(unwind, bson.DocElem{"preserveNullAndEmptyArrays",
			join.kind == leftJoin})

		v.logger.Debugf(log.Dev, "consolidating foreign unwind "+
			"into local pipeline")

		pipeline = append(pipeline, bson.D{{"$unwind", unwind}})

		// Handle edge case where the lookup field is the same as the
		// $unwind's array path. In this case, we must apply an
		// additional filter to remove records in the now unwound array
		// that don't meet the lookup criteria.
		if lookupOnArrayField {
			foreignField := fmt.Sprintf("$%v.%v", asField, foreignFieldName)
			filter := bson.M{mgoOperatorEq: []interface{}{
				foreignField, "$" + localFieldName,
			}}

			fieldName := v.uniqueFieldName(projectPredicateFieldName, newMappingRegistry)
			stageBody := bson.M{
				fieldName: filter,
			}
			predicateEvaluationStage := v.buildAddFieldsOrProject(stageBody, []string{}, newMappingRegistry)

			match := bson.M{fieldName: true}
			if join.kind == leftJoin {
				// For left joins, we need to ensure we retain records
				// from the left child - in case the unwound array was
				// empty or null.
				match = bson.M{mgoOperatorOr: []interface{}{match,
					bson.M{path[1:]: bson.M{mgoOperatorExists: false}}}}
			}
			pipeline = append(pipeline, predicateEvaluationStage,
				bson.D{{"$match", match}},
			)
		}
	}

	// 7. Build the new operators.
	msLocal.aliasNames = append(msLocal.aliasNames, msForeign.aliasNames...)
	msLocal.tableNames = append(msLocal.tableNames, msForeign.tableNames...)
	msLocal.collectionNames = append(msLocal.collectionNames, msForeign.collectionNames...)
	for key, val := range msForeign.isShardedCollection {
		msLocal.isShardedCollection[key] = val
	}
	ms := msLocal.clone()
	ms.selectIDs = append(ms.selectIDs, msForeign.selectIDs...)
	ms.pipeline = pipeline
	ms.mappingRegistry = newMappingRegistry

	if lookupInfo.remainingPredicate != nil && (kind == innerJoin || kind == straightJoin) {
		f, err := v.visit(NewFilterStage(ms, lookupInfo.remainingPredicate))
		if err != nil {
			return nil, err
		}

		return f.(PlanStage), nil
	}

	v.logger.Debugf(log.Dev, "successfully translated join stage to lookup")

	return ms, nil
}

type lookupInfo struct {
	localColumn        *SQLColumnExpr
	foreignColumn      *SQLColumnExpr
	remainingPredicate SQLExpr
}

func getLocalAndForeignColumns(localTable, foreignTable *MongoSourceStage, e SQLExpr) (*lookupInfo, error) {
	exprs := splitExpression(e)
	// Find a SQLEqualsExpr where the left and right are columns from the local and foreign tables.
	for i, expr := range exprs {
		if equalExpr, ok := expr.(*SQLEqualsExpr); ok {
			// We must have a field from the left table and a field from the right table.
			if column1, ok := equalExpr.left.(SQLColumnExpr); ok {
				if column2, ok := equalExpr.right.(SQLColumnExpr); ok {
					var localColumn, foreignColumn *SQLColumnExpr
					if containsString(localTable.aliasNames, column1.tableName) {
						localColumn = &column1
					} else if containsString(foreignTable.aliasNames, column1.tableName) {
						foreignColumn = &column1
					}

					if containsString(localTable.aliasNames, column2.tableName) {
						localColumn = &column2
					} else if containsString(foreignTable.aliasNames, column2.tableName) {
						foreignColumn = &column2
					}

					if localColumn != nil && foreignColumn != nil {
						combined := combineExpressions(append(exprs[:i], exprs[i+1:]...))
						return &lookupInfo{localColumn, foreignColumn, combined}, nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("join condition cannot be pushed down '%v'", e)
}

// lookupSQLColumnForJoin looks up the _original_ field name for a given
// table.column in a join constraint. This needs to take into account any
// renames that have occurred due to self join optimized left joins The reason
// we do this is that we are using the original field name to denote the
// semantic identity of columns for the purposes of PK equality matching
// constraints as we need to identify two columns as being semantically
// isomorphic if they have been aliased at the SQL level.
func (v *pushDownOptimizer) lookupSQLColumnForJoin(databaseName, tableName, columnName string, mappingRegistries []*mappingRegistry) (string, string, int, bool) {
	var renamedField string
	var ok bool
	if v == nil {
		renamedField = ""
	} else if renamedField, ok = v.leftJoinOriginalNames[databaseName][tableName][columnName]; !ok {
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
	stages           []bson.D
	arrayPaths       []string
	arrayPathIndexes []string
}

func (v *pushDownOptimizer) selfJoinOptimizePipeline(local, foreign *MongoSourceStage, kind joinKind) (*consolidatedPipeline, error) {
	pipeline := &consolidatedPipeline{}

	augmentProjection := func(stage interface{}) (bson.D, error) {
		project, ok := stage.(bson.M)
		if !ok {
			return bson.D{}, fmt.Errorf("found unexpected type %T for project stage in local table (%v) pipeline", stage, local.aliasNames)
		}

		prefixPathPresent := func(project bson.M, fieldName string) bool {
			names := strings.Split(fieldName, ".")
			for i := 0; i < len(names); i++ {
				if _, ok := project[sanitizeFieldName(strings.Join(names[:i], "."))]; ok {
					return true
				}
			}
			return false
		}

		for _, c := range foreign.mappingRegistry.columns {
			fieldName, ok := foreign.mappingRegistry.lookupFieldName(c.Database, c.Table, c.Name)
			if !ok {
				return nil, fmt.Errorf("cannot find referenced foreign column %v.%v.%v in projection lookup", c.Database, c.Table, c.Name)
			}

			if _, ok := project[fieldName]; !ok && !prefixPathPresent(project, fieldName) {
				v.logger.Debugf(log.Dev, "augmenting local table with column '%v.%v'.'%v'", c.Database, c.Table, c.Name)
				project[fieldName] = 1
				foreign.mappingRegistry.registerMapping(c.Database, c.Table, c.Name, fieldName)
			}
		}

		return bson.D{{"$project", project}}, nil
	}

	getPathsAndPipeline := func(stages []bson.D, isLocal bool) error {
		for _, stage := range stages {
			bsonStage, ok := stage.Map()["$unwind"]
			if !ok {
				if isLocal {
					// For projections, ensure all foreign columns are included.
					if bsonStage, ok = stage.Map()["$project"]; ok {
						project, err := augmentProjection(bsonStage)
						if err != nil {
							return err
						}
						pipeline.stages = append(pipeline.stages, project)
					} else {
						pipeline.stages = append(pipeline.stages, stage)
					}
					continue
				} else {
					return fmt.Errorf("found stage in foreign table (%v) pipeline: %#v that cannot be self-join optimized", foreign.aliasNames, stage)
				}
			}

			unwind, ok := bsonStage.(bson.D)
			// It is possible for a bson.D to contain a bson.M
			// because of our $unwind on the $lookup created field
			// following a $lookup. We never hit this case
			// before BI-1154 because we were self-join optimizing joins
			// too often, and this requires a multi-way join where
			// one source for one of the joins results in a $lookup.
			var fields bson.M
			if ok {
				fields = unwind.Map()
			} else {
				fields = bsonStage.(bson.M)
			}

			iPath, ok := fields["path"]
			if !ok {
				return fmt.Errorf("could not find unwind path in foreign table %v: %#v", foreign.aliasNames, stage)
			}

			iIndex, ok := fields["includeArrayIndex"]
			if !ok {
				return fmt.Errorf("could not find unwind array index in foreign table %v: %#v", foreign.aliasNames, stage)
			}

			// Strip dollar sign prefix.
			path := fmt.Sprintf("%v", iPath)
			if path != "" {
				path = path[1:]
			}

			arrayIdx := fmt.Sprintf("%v", iIndex)

			if !util.StringSliceContains(pipeline.arrayPathIndexes, arrayIdx) {
				pipeline.arrayPaths = append(pipeline.arrayPaths, path)
				pipeline.arrayPathIndexes = append(pipeline.arrayPathIndexes, arrayIdx)
				_, ok = fields["preserveNullAndEmptyArrays"]
				if !ok && kind == leftJoin && !isLocal {
					unwind = append(unwind, bson.DocElem{"preserveNullAndEmptyArrays", true})
				}
				pipeline.stages = append(pipeline.stages, bson.D{{"$unwind", unwind}})
			}

		}
		return nil
	}

	if err := getPathsAndPipeline(local.pipeline, true); err != nil {
		return nil, err
	}

	if err := getPathsAndPipeline(foreign.pipeline, false); err != nil {
		return nil, err
	}

	if kind == leftJoin {
		localUnwindFields := getPipelineUnwindFields(local.pipeline)
		foreignUnwindFields := getPipelineUnwindFields(foreign.pipeline)
		totalUnwindFields := getPipelineUnwindFields(pipeline.stages)
		// If the local has more unwinds than the foreign, this is equivalent to an
		// inner join, just return the optimized pipeline.
		if len(localUnwindFields) > len(foreignUnwindFields) {
			return pipeline, nil
		}
		unwindSuffix, ok := getUnwindSuffix(localUnwindFields, foreignUnwindFields)

		// It is safe to to allow left joins with non-progeny as long as
		// the foreign pipeline only has 1 unwind.
		if !ok && len(foreignUnwindFields) > 1 {
			panic("Unwind prefixes do not match, this should have disallowed self-join optimization. " +
				"This should never happen.")
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
		unwindSuffixIndexes := getIndexes(unwindSuffix)
		unwindSuffixPaths := getPaths(unwindSuffix)

		// Insertion point should be *after* the first unwind in the
		// unwindSuffix If it is inserted before, it will not always
		// work, and when it does work it is due to luck, not correct
		// semantics, but the stageNumber in the unwindSuffix may not
		// be the same as the stageNumber for that $unwind in the new
		// self-join optimized pipeline, which is what we actually
		// need.
		insertionPointPath := unwindSuffix[0].path
		insertionPointUnwind, ok := findUnwindForPath(totalUnwindFields, insertionPointPath)
		if !ok {
			panic(fmt.Sprintf("Could not find unwind for path %v in pipeline %v, "+
				"this should never happen)",
				insertionPointPath, pipeline.stages))
		}
		insertionPoint := insertionPointUnwind.stageNumber

		addFieldsBody := bson.M{}
		for databaseName, tables := range foreign.mappingRegistry.fields {
			leftJoinOriginalNames, ok := v.leftJoinOriginalNames[databaseName]
			if !ok {
				leftJoinOriginalNames = make(map[string]map[string]string)
				v.leftJoinOriginalNames[databaseName] = leftJoinOriginalNames
			}
			for tableName, fields := range tables {
				leftJoinOriginalNames, ok := v.leftJoinOriginalNames[databaseName][tableName]
				if !ok {
					leftJoinOriginalNames = make(map[string]string)
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
					if pathStartsWithAny(unwindSuffixPaths, "$"+docCol) ||
						util.StringSliceContains(unwindSuffixIndexes, docCol) {
						continue
					}
					uniqueDocCol := v.uniqueFieldName(docCol, local.mappingRegistry, foreign.mappingRegistry)

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
					addFieldsBody[uniqueDocCol] = buildLeftSelfJoinPKAddFieldBody(
						unwindSuffix[0].path, "$"+docCol)
				}
			}
		}
		addFields := v.buildAddFieldsOrProject(addFieldsBody, []string{}, local.mappingRegistry, foreign.mappingRegistry)
		pipeline.stages = insertPipelineStageAt(pipeline.stages, addFields, insertionPoint)
	}

	return pipeline, nil
}

// buildLeftSelfJoinPKAddFieldBody builds the conditional for an AddField,
// checking column columnCheck for missing, NULL, or empty.
func buildLeftSelfJoinPKAddFieldBody(columnCheck, columnCopy string) bson.D {
	lteNil := bson.D{{"$lte", []interface{}{columnCheck, nil}}}
	eqEmpty := bson.D{{"$eq", []interface{}{columnCheck, []interface{}{}}}}
	totalOr := bson.D{{"$or", []interface{}{lteNil, eqEmpty}}}

	addFieldBody := bson.D{
		{"$cond", []interface{}{totalOr, nil, columnCopy}},
	}

	return addFieldBody
}

func (v *pushDownOptimizer) visitLimit(limit *LimitStage) (PlanStage, error) {

	ms, ok := v.canPushDown(limit.source)
	if !ok {
		return limit, nil
	}

	pipeline := ms.pipeline

	if limit.offset > 0 {
		if limit.offset > math.MaxInt64 {
			return nil, fmt.Errorf("limit with offset '%d' cannot be pushed down", limit.offset)
		}
		pipeline = append(pipeline, bson.D{{"$skip", int64(limit.offset)}})
	}

	if limit.limit > 0 {
		if limit.limit > math.MaxInt64 {
			return nil, fmt.Errorf("limit with rowcount '%d' cannot be pushed down", limit.limit)
		}
		pipeline = append(pipeline, bson.D{{"$limit", int64(limit.limit)}})
	}

	ms = ms.clone()
	ms.pipeline = pipeline
	return ms, nil
}

func (v *pushDownOptimizer) visitOrderBy(orderBy *OrderByStage) (PlanStage, error) {

	ms, ok := v.canPushDown(orderBy.source)
	if !ok {
		return orderBy, nil
	}

	sort := bson.D{}
	var newFields bson.M
	var t *pushDownTranslator

	for _, term := range orderBy.terms {

		var databaseName, tableName, columnName string

		switch typedE := term.expr.(type) {
		case SQLColumnExpr:
			databaseName, tableName, columnName = typedE.databaseName, typedE.tableName, typedE.columnName
		default:
			// MongoDB only allows sorting by a field, so pushing down a
			// non-field requires it to be pre-calculated by a previous
			// push down. If it has been pre-calculated, then it will
			// exist in the mapping registry. Otherwise, it won't, and
			// we'll need to push this down with a $project or $addFields.
			columnName = typedE.String()
		}

		fieldName, ok := ms.mappingRegistry.lookupFieldName(databaseName, tableName, columnName)
		if !ok {
			// Since we can't push this down, we'll attempt to build up a $project/$addFields
			// that will allow us to push this down using aggregation language, then sort by the
			// added columns.
			if t == nil {
				t = &pushDownTranslator{
					versionAtLeast:  v.ctx.Variables().MongoDBInfo.VersionAtLeast,
					lookupFieldName: ms.mappingRegistry.lookupFieldName,
					logger:          v.logger,
				}
			}

			var translated interface{}
			if translated, ok = t.TranslateExpr(term.expr); !ok {
				v.logger.Warnf(log.Dev, "unable to push down order by due to term '%v'", columnName)
				return orderBy, nil
			}

			if newFields == nil {
				newFields = bson.M{}
			}

			fieldName = v.uniqueFieldName(sanitizeFieldName(columnName), ms.mappingRegistry)
			newFields[fieldName] = translated
		}
		direction := 1
		if !term.ascending {
			direction = -1
		}
		sort = append(sort, bson.DocElem{fieldName, direction})
	}

	pipeline := ms.pipeline

	if len(newFields) > 0 {
		// NOTE: there is no reason to mess with the mapping registry
		// because the added fields are only used in the immediate
		// $sort stage and will never be referenced again.
		stageName, stageBody := "$addFields", bson.M{}
		if !t.versionAtLeast(3, 4, 0) {
			stageBody = v.projectAllColumns(ms.mappingRegistry)
			stageName = "$project"
		}

		for k, v := range newFields {
			stageBody[k] = v
		}

		pipeline = append(pipeline, bson.D{{stageName, stageBody}})
	}

	pipeline = append(pipeline, bson.D{{"$sort", sort}})

	ms = ms.clone()
	ms.pipeline = pipeline
	return ms, nil
}

const (
	emptyFieldNamePrefix = "__empty"
)

func (v *pushDownOptimizer) visitProject(project *ProjectStage) (PlanStage, error) {

	// Check if we can optimize further, if the child operator has a MongoSource.
	ms, ok := v.canPushDown(project.source)
	if !ok {
		return project, nil
	}

	fieldsToProject := bson.M{}

	// This will contain the rebuilt mapping registry reflecting fields re-mapped by projection.
	fixedMappingRegistry := mappingRegistry{}

	fixedProjectedColumns := ProjectedColumns{}

	// Track whether or not we've successfully mapped every field into the $project of the source.
	// If so, this Project node can be removed from the query plan tree.
	canReplaceProject := true

	t := &pushDownTranslator{
		versionAtLeast:  v.ctx.Variables().MongoDBInfo.VersionAtLeast,
		lookupFieldName: ms.mappingRegistry.lookupFieldName,
		logger:          v.logger,
	}

	for _, projectedColumn := range project.projectedColumns {
		// Convert the column's SQL expression into an expression in mongo query language.
		projectedField, ok := t.TranslateExpr(projectedColumn.Expr)
		if !ok {
			v.logger.Debugf(log.Dev, "could not translate '%v'", projectedColumn.Expr.String())
			// Expression can't be translated, so it can't be projected.
			// We skip it and leave this Project node in the query plan so that it still gets
			// evaluated during execution.
			canReplaceProject = false

			// There might still be fields referenced in this expression
			// that we still need to project, so collect them and add them to the projection.
			refdCols, err := referencedColumns(v.selectIDsInScope, projectedColumn.Expr)
			if err != nil {
				v.logger.Warnf(log.Dev, "cannot find referenced project expression: %v", err)
				return nil, err
			}

			for _, refdCol := range refdCols {
				refdCol.PrimaryKey = projectedColumn.PrimaryKey
				fieldName, ok := ms.mappingRegistry.lookupFieldName(refdCol.Database, refdCol.Table, refdCol.Name)
				if !ok {
					v.logger.Warnf(log.Dev, "cannot find referenced column %#v in lookup", refdCol)
					return project, nil
				}

				safeFieldName := sanitizeFieldName(fieldName)
				fieldsToProject[safeFieldName] = getProjectedFieldName(fieldName, refdCol.SQLType)
				fixedMappingRegistry.addColumn(refdCol)
				fixedMappingRegistry.registerMapping(refdCol.Database, refdCol.Table, refdCol.Name, safeFieldName)
			}

			fixedProjectedColumns = append(fixedProjectedColumns, projectedColumn)
		} else {
			safeFieldName := sanitizeFieldName(projectedColumn.Expr.String())

			// If the name turns out to be empty, we need to assign our own.
			if safeFieldName == "" {
				safeFieldName = emptyFieldNamePrefix
			}
			safeFieldName = v.uniqueFieldName(safeFieldName, &fixedMappingRegistry)

			fieldsToProject[safeFieldName] = projectedField
			fixedMappingRegistry.addColumn(projectedColumn.Column)

			registryName := v.uniqueRegistryName(&fixedMappingRegistry, projectedColumn.Database, projectedColumn.Table, projectedColumn.Name)
			if projectedColumn.Name != registryName {
				projectedColumn.MappingRegistryName = registryName
			}

			fixedMappingRegistry.registerMapping(projectedColumn.Database, projectedColumn.Table, registryName, safeFieldName)

			columnExpr := NewSQLColumnExpr(
				projectedColumn.SelectID,
				projectedColumn.Column.Database,
				projectedColumn.Column.Table,
				projectedColumn.Column.Name,
				projectedColumn.SQLType,
				projectedColumn.MongoType)

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

	ms = ms.clone()
	ms.pipeline = append(ms.pipeline, bson.D{{"$project", fieldsToProject}})
	ms.mappingRegistry = &fixedMappingRegistry

	if canReplaceProject {
		return ms, nil
	}

	project = project.clone()
	project.source = ms
	project.projectedColumns = fixedProjectedColumns

	return project, nil
}

func (v *pushDownOptimizer) visitSubquerySource(subquery *SubquerySourceStage) (PlanStage, error) {
	// Check if we can optimize further, if the child operator has a MongoSource.
	ms, ok := v.canPushDown(subquery.source)
	if !ok {
		return subquery, nil
	}

	mappingRegistry := mappingRegistry{}
	for _, column := range ms.mappingRegistry.columns {
		fieldName, ok := ms.mappingRegistry.lookupFieldName(column.Database, column.Table, column.Name)
		if !ok {
			v.logger.Warnf(log.Dev, "cannot find referenced subquery column %#v in lookup", column)
			return subquery, nil
		}

		mappingRegistry.addColumn(NewColumn(column.SelectID, subquery.aliasName, column.Table, column.Database, column.Name, column.OriginalName, column.MappingRegistryName,
			column.SQLType, column.MongoType, false))
		mappingRegistry.registerMapping(column.Database, subquery.aliasName, column.Name, fieldName)
	}

	ms = ms.clone()
	ms.aliasNames = []string{subquery.aliasName}
	ms.mappingRegistry = &mappingRegistry
	return ms, nil
}

type columnFinder struct {
	selectIDsInScope []int
	columns          Columns
}

// referencedColumns will take an expression and return all the columns
// referenced in the expression.
func referencedColumns(selectIDsInScope []int, e SQLExpr) ([]*Column, error) {

	cf := &columnFinder{selectIDsInScope: selectIDsInScope}

	_, err := cf.visit(e)
	if err != nil {
		return nil, err
	}

	return cf.columns.Unique(), nil
}

func (cf *columnFinder) visit(n node) (node, error) {
	switch typedN := n.(type) {
	case SQLColumnExpr:
		if containsInt(cf.selectIDsInScope, typedN.selectID) {
			column := NewColumn(typedN.selectID, typedN.tableName, "", typedN.databaseName, typedN.columnName, "", "",
				typedN.columnType.SQLType, typedN.columnType.MongoType, false)
			cf.columns = append(cf.columns, column)
		}
		return n, nil
	}

	return walk(cf, n)
}

// pushdownProject is called when a stage could not be pushed down. It uses
// requiredColumns to create and visit a new projectStage in order to project
// out only the columns needed for the rest of the query so that we do not have
// to pull all data from a table into memory.
func (v *pushDownOptimizer) pushdownProject(requiredColumns []SQLExpr, source PlanStage) (PlanStage, error) {
	sf := &sourceFinder{}
	_, err := sf.visit(source)
	if err != nil {
		return nil, err
	}

	var projectedCols []ProjectedColumn
	for _, expr := range requiredColumns {
		if sqlColExpr, ok := expr.(SQLColumnExpr); ok {
			column := NewColumn(sqlColExpr.selectID, sqlColExpr.tableName, "", sqlColExpr.databaseName, sqlColExpr.columnName, "", "", sqlColExpr.Type(), sqlColExpr.columnType.MongoType, sf.source.mappingRegistry.isPrimaryKey(sqlColExpr.columnName))
			projectedCols = append(projectedCols, ProjectedColumn{Column: column, Expr: expr})
		}
	}
	project := NewProjectStage(source, projectedCols...)
	projSource, err := v.visitProject(project)
	if err != nil {
		return nil, fmt.Errorf("unable to optimize project: %v", err)
	}
	return projSource, nil
}

func (v *pushDownOptimizer) getRequiredColumnsForJoinSide(databaseName string, tableNames []string, requiredColumns []SQLExpr) []SQLExpr {
	var result []SQLExpr
	for _, expr := range requiredColumns {
		columnExpr, ok := expr.(SQLColumnExpr)
		if !ok {
			continue
		}

		if columnExpr.databaseName != databaseName {
			continue
		}

		if containsString(tableNames, columnExpr.tableName) {
			result = append(result, expr)
		}
	}

	return result
}

func (v *pushDownOptimizer) projectAllColumns(mr *mappingRegistry) bson.M {
	projectBody := bson.M{}
	for _, c := range mr.columns {
		field, ok := mr.lookupFieldName(c.Database, c.Table, c.Name)
		if !ok {
			panic("Unable to find field mapping for column. This should never happen.")
		}
		projectBody[field] = 1
	}
	return projectBody
}

// uniqueFieldName creates a field name that is unique across all tables in a
// set of registries.
func (v *pushDownOptimizer) uniqueFieldName(fieldName string, mrs ...*mappingRegistry) string {
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

// uniqueRegistryName creates a name that is unique to a table: they can be
// repeated in separate tables, use uniqueFieldName for a field name that is
// unique in the entire registry (or set of registries).
func (v *pushDownOptimizer) uniqueRegistryName(mr *mappingRegistry, databaseName, tableName, columnName string) string {
	if _, hasLookup := mr.lookupFieldName(databaseName, tableName, columnName); !hasLookup {
		return columnName
	}

	retColumnName := columnName

	i := 1
	for {
		retColumnName = fmt.Sprintf("%v_%v", columnName, i)
		if _, hasLookup := mr.lookupFieldName(databaseName, tableName, retColumnName); !hasLookup {
			return retColumnName
		}
		i++
	}
}

func newColumnTracker() *columnTracker {
	return &columnTracker{
		selectIDs: make(map[int]*exprCountMap),
	}
}

// columnTracker is for scoped handling of column names like a symbol table in
// a compiler. New scopes are introduced by sub-queries.
type columnTracker struct {
	selectIDs  map[int]*exprCountMap
	removeMode bool
}

func (t *columnTracker) getColumnsForSelectIDs(selectIDs []int) []SQLExpr {
	var columns []SQLExpr
	for _, selectID := range selectIDs {
		if selectIDMap, ok := t.selectIDs[selectID]; ok {
			columns = append(columns, selectIDMap.exprs...)
		}
	}

	return columns
}

func (t *columnTracker) add(e SQLExpr) {
	t.removeMode = false
	t.visit(e)
}

func (t *columnTracker) remove(e SQLExpr) {
	t.removeMode = true
	t.visit(e)
}

func (t *columnTracker) visit(n node) (node, error) {
	switch typedN := n.(type) {
	case SQLColumnExpr:
		selectIDMap, ok := t.selectIDs[typedN.selectID]
		if !ok && !t.removeMode {
			selectIDMap = newExprCountMap()
			t.selectIDs[typedN.selectID] = selectIDMap
		}

		if t.removeMode {
			if selectIDMap != nil {
				selectIDMap.remove(typedN)
				if len(selectIDMap.exprs) == 0 {
					delete(t.selectIDs, typedN.selectID)
				}
			}
		} else {
			selectIDMap.add(typedN)
		}

		return n, nil
	}

	return walk(t, n)
}

// sourceFinder is used within projection pushdown to locate the MongoSource
// stage to project.
type sourceFinder struct {
	source *MongoSourceStage
}

func (f *sourceFinder) visit(n node) (node, error) {
	n, err := walk(f, n)
	if err != nil {
		return nil, err
	}

	switch typedN := n.(type) {
	case *MongoSourceStage:
		f.source = typedN
	}

	return n, nil
}

func (v *pushDownOptimizer) canSelfJoinTables(logger *log.Logger, local, foreign *MongoSourceStage, matcher SQLExpr, kind joinKind) bool {
	return sharesRootTable(logger, local, foreign) &&
		v.meetsSelfJoinPKCriteria(logger, local, foreign, matcher) &&
		(kind != leftJoin || v.meetsLeftSelfJoinPipelineCriteria(logger, local, foreign, matcher))
}

func (v *pushDownOptimizer) remainingJoinPredicate(msLocal, msForeign *MongoSourceStage, e SQLExpr) []SQLExpr {
	exprs, newExprs := splitExpression(e), []SQLExpr{}
	registries := []*mappingRegistry{msLocal.mappingRegistry,
		msForeign.mappingRegistry}
	for _, expr := range exprs {
		if equalExpr, ok := expr.(*SQLEqualsExpr); ok {
			c1, _ := equalExpr.left.(SQLColumnExpr)
			c2, _ := equalExpr.right.(SQLColumnExpr)
			if c1.selectID == c2.selectID {

				originalC1Name, _, c1RegistryIdx, ok := v.lookupSQLColumnForJoin(c1.databaseName,
					c1.tableName, c1.columnName, registries)
				if !ok {
					panic("Unable to find field mapping for self-join optimization " +
						"c1. This should never happen.")
				}

				originalC2Name, _, c2RegistryIdx, ok := v.lookupSQLColumnForJoin(c2.databaseName,
					c2.tableName, c2.columnName, registries)
				if !ok {
					panic("Unable to find field mapping for self-join optimization " +
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

func (v *pushDownOptimizer) meetsLeftSelfJoinPipelineCriteria(logger *log.Logger, local, foreign *MongoSourceStage, matcher SQLExpr) bool {
	hasRemainingPredicate := len(v.remainingJoinPredicate(local, foreign, matcher)) != 0
	// Get the paths of each unwind in each pipeline as an array of strings.
	localUnwindPipelineFields, foreignUnwindPipelineFields := getPipelineUnwindFields(local.pipeline), getPipelineUnwindFields(foreign.pipeline)

	lenLocal, lenForeign := len(localUnwindPipelineFields), len(foreignUnwindPipelineFields)

	// We can't have any issues with embedded NULLs and empties if the
	// foreign pipeline only has one $unwind and there is no remaining
	// predicate.
	if lenForeign == 1 {
		return true
	}

	localUnwindPipelinePaths, foreignUnwindPipelinePaths := getPaths(localUnwindPipelineFields), getPaths(foreignUnwindPipelineFields)

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
	// because of preserveNullAndEmptyArrays we would only want to keep
	// NULLs and empties from the top level of unwinding in the foreign
	// side. Theoretical attempts to do such filtering have all fallen
	// down on edge cases. This is also why we disallow progeny with more
	// than 1 generation difference.
	logger.Debugf(log.Dev, "self-join optimization: could not optimize because of left join "+
		"criteria failure - local unwind paths: %v foreign unwind paths %v",
		localUnwindPipelinePaths, foreignUnwindPipelinePaths)
	return false
}

func (v *pushDownOptimizer) meetsSelfJoinPKCriteria(logger *log.Logger, local, foreign *MongoSourceStage, matcher SQLExpr) bool {
	// Don't perform optimization on MongoDB views as
	// renames might have occured on fields.
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

	getPKs := func(columns []*Column, table string) map[string]struct{} {
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
	localPKs := getPKs(local.mappingRegistry.columns, local.aliasNames[0])
	foreignPKs := getPKs(foreign.mappingRegistry.columns, foreign.aliasNames[0])
	intersectionPKs := intersectionStringSet(localPKs, foreignPKs)

	numRequiredPKConjunctions := len(intersectionPKs)

	if numRequiredPKConjunctions == 0 {
		logger.Debugf(log.Dev, "cannot use self-join optimization, table "+
			"has no primary key")
		return false
	}

	numPKConjunctions := 0

	logger.Debugf(log.Dev, "self-join optimization: examining match criteria...")

	registries := []*mappingRegistry{
		local.mappingRegistry,
		foreign.mappingRegistry,
	}

	seenPrimaryKeys := make(map[string]struct{})

	for _, expr := range exprs {
		if equalExpr, ok := expr.(*SQLEqualsExpr); ok {
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

			originalC1Name, _, c1RegistryIdx, ok := v.lookupSQLColumnForJoin(column1.databaseName,
				column1.tableName, column1.columnName, registries)
			if !ok {
				panic(fmt.Sprintf("Unable to find field mapping for merge column1:  %s.%s.%s."+
					"This should never happen.", column1.databaseName, column1.tableName, column1.columnName))
			}
			originalC2Name, _, c2RegistryIdx, ok := v.lookupSQLColumnForJoin(column2.databaseName,
				column2.tableName, column2.columnName, registries)
			if !ok {
				panic(fmt.Sprintf("Unable to find field mapping for merge column2: %s.%s.%s."+
					"This should never happen.", column2.databaseName, column2.tableName, column2.columnName))
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
			numPKConjunctions, util.Pluralize(numPKConjunctions, "pair",
				"pairs"), numRequiredPKConjunctions, loggingPKSetStr)
		return false
	}

	return true
}
