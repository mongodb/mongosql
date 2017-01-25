package evaluator

import (
	"fmt"
	"math"
	"strings"

	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/util"
	"gopkg.in/mgo.v2/bson"
)

func optimizePushDown(ctx ConnectionCtx, logger *log.Logger, n node) (node, error) {
	v := &pushDownOptimizer{logger: logger, ctx: ctx}
	n, err := v.visit(n)
	if err != nil {
		return nil, err
	}

	return n, nil
}

type pushDownOptimizer struct {
	logger            *log.Logger
	ctx               ConnectionCtx
	selectIDsInScope  []int
	tableNamesInScope []string
}

func (v *pushDownOptimizer) visit(n node) (node, error) {
	n, err := walk(v, n)
	if err != nil {
		return nil, err
	}

	switch typedN := n.(type) {
	// Since we are walking to the bottom of the tree, we'll collect all
	// the selectIDs that are currently in scope. In the case of Joins,
	// this could be a combination of the below select ID sources.
	case *MongoSourceStage:
		v.selectIDsInScope = append(v.selectIDsInScope, typedN.selectIDs...)
		v.tableNamesInScope = append(v.tableNamesInScope, typedN.aliasNames...)
	case *BSONSourceStage:
		v.selectIDsInScope = append(v.selectIDsInScope, typedN.selectID)
	case *DynamicSourceStage:
		v.selectIDsInScope = append(v.selectIDsInScope, typedN.selectID)
	case *SQLSubqueryExpr:
		// SQLSubqueryExpr only applies to non-from clauses. This means that
		// any new selectIDs found inside a SQLSubqueryExpr are invalid outside
		// of it. However, the selectIDs outside of it are valid inside. This is
		// the definition of a correlated subquery. So, we'll save off the current
		// selectIDs and restore them afterwards.

		oldSelectIDsInScope := v.selectIDsInScope
		oldTableNamesInScope := v.tableNamesInScope

		n, err = walk(v, n)
		if err != nil {
			return nil, err
		}

		v.selectIDsInScope = oldSelectIDsInScope
		v.tableNamesInScope = oldTableNamesInScope

	// Push Down
	case *FilterStage:
		n, err = v.visitFilter(typedN)
		if err != nil {
			return nil, fmt.Errorf("unable to optimize filter: %v", err)
		}

		if fs, ok := n.(*FilterStage); ok {
			if _, ok := fs.source.(*MongoSourceStage); ok {
				projSource, err := v.pushdownProject(fs.requiredColumns, fs.source)
				if err != nil {
					return nil, fmt.Errorf("unable to optimize filter project: %v", err)
				}
				n = NewFilterStage(projSource, fs.matcher, fs.requiredColumns)
			}
		}
	case *GroupByStage:
		n, err = v.visitGroupBy(typedN)
		if err != nil {
			return nil, fmt.Errorf("unable to optimize group by: %v", err)
		}

		if _, ok := typedN.source.(*MongoSourceStage); ok && n == typedN {
			projSource, err := v.pushdownProject(typedN.requiredColumns, typedN.source)
			if err != nil {
				return nil, fmt.Errorf("unable to optimize group by project: %v", err)
			}
			n = NewGroupByStage(projSource, typedN.keys, typedN.projectedColumns, typedN.requiredColumns)
		}

	case *JoinStage:
		n, err = v.visitJoin(typedN)
		if err != nil {
			return nil, fmt.Errorf("unable to optimize join: %v", err)
		}

		if typedN, ok := n.(*JoinStage); ok {
			left := typedN.left
			right := typedN.right

			// for inner joins, attempt to optimize by flipping children
			if typedN.kind == InnerJoin {
				v.logger.Warnf(log.DebugHigh, "attempting to optimize inner join via flip")
				j := NewJoinStage(typedN.kind, typedN.right, typedN.left, typedN.matcher,
					typedN.requiredColumns)
				newJ, err := v.visitJoin(j)
				if err == nil {
					n = newJ
				}
			} else if typedN.kind == RightJoin {
				// for right joins, attempt to optimize using a left join
				v.logger.Warnf(log.DebugHigh, "attempting to optimize right join via flip")
				j := NewJoinStage(LeftJoin, typedN.right, typedN.left, typedN.matcher,
					typedN.requiredColumns)
				newJ, err := v.visitJoin(j)
				if err == nil {
					n = newJ
				}
			}

			if _, ok := n.(*JoinStage); ok {

				if ms, ok := left.(*MongoSourceStage); ok {
					requiredColumns := v.getRequiredColumnsForJoinSide(
						ms.aliasNames, typedN.requiredColumns)
					left, err = v.pushdownProject(requiredColumns, ms.clone())
					if err != nil {
						return nil, fmt.Errorf("unable to optimize JoinStage left project: %v", err)
					}
				}

				if ms, ok := right.(*MongoSourceStage); ok {
					requiredColumns := v.getRequiredColumnsForJoinSide(
						ms.aliasNames, typedN.requiredColumns)
					right, err = v.pushdownProject(requiredColumns, ms.clone())
					if err != nil {
						return nil, fmt.Errorf("unable to optimize JoinStage right project: %v", err)
					}
				}

				if left != typedN.left || right != typedN.right {
					n = NewJoinStage(typedN.kind, left, right, typedN.matcher, typedN.requiredColumns)
				}

			}

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
			projSource, err := v.pushdownProject(typedN.requiredColumns, typedN.source)
			if err != nil {
				return nil, fmt.Errorf("unable to optimize order by project: %v", err)
			}
			n = NewOrderByStage(projSource, typedN.requiredColumns, typedN.terms...)
		}
	case *ProjectStage:
		n, err = v.visitProject(typedN)
		if err != nil {
			return nil, fmt.Errorf("unable to optimize project: %v", err)
		}
	case *SubquerySourceStage:
		oldSelectIDsInScope := v.selectIDsInScope
		oldTableNamesInScope := v.tableNamesInScope

		// Inside a SubquerySourceStage, there are no selectIDs or tableNames
		// in scope. However, after we are finished, the existing selectIDs
		// and tableNames are in scope as well as the additional selectID and
		// aliasName of the subquery.
		v.selectIDsInScope = []int{}
		v.tableNamesInScope = []string{}

		n, err = v.visitSubquerySource(typedN)
		if err != nil {
			return nil, fmt.Errorf("unable to optimize subquery source: %v", err)
		}

		v.selectIDsInScope = append(oldSelectIDsInScope, typedN.selectID)
		v.tableNamesInScope = append(oldTableNamesInScope, typedN.aliasName)
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

func (v *pushDownOptimizer) distinctProjectFieldName(project bson.M) string {
	fieldName := projectPredicateFieldName
	i := 0
	for {
		if _, ok := project[fieldName]; !ok {
			break
		}
		i++
		fieldName = fmt.Sprintf("%s%d", projectPredicateFieldName, i)
	}

	return fieldName
}

const (
	projectPredicateFieldName = "__predicate"
)

func (v *pushDownOptimizer) visitFilter(filter *FilterStage) (PlanStage, error) {

	ms, ok := v.canPushDown(filter.source)
	if !ok {
		return filter, nil
	}

	pipeline := ms.pipeline
	var localMatcher SQLExpr

	if value, ok := filter.matcher.(SQLValue); ok {
		// our optimized expression has left us with just a value,
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

		// otherwise, the filter simply gets removed from the tree

	} else {
		var matchBody bson.M
		t := &pushDownTranslator{
			versionAtLeast:  v.ctx.Variables().MongoDBInfo.VersionAtLeast,
			lookupFieldName: ms.mappingRegistry.lookupFieldName,
		}
		matchBody, localMatcher = t.TranslatePredicate(filter.matcher)
		if matchBody != nil {
			pipeline = append(pipeline, bson.D{{"$match", matchBody}})
		}

		if localMatcher != nil {
			// we have a predicate that completely or partially couldn't be handled by $match.
			// Attempt to push it down as part of a $project/$match combination.
			if predicate, ok := t.TranslateExpr(localMatcher); ok {
				projectBody := v.projectAllColumns(ms.mappingRegistry)
				fieldName := v.distinctProjectFieldName(projectBody)
				projectBody[fieldName] = predicate
				pipeline = append(pipeline,
					bson.D{{"$project", projectBody}},
					bson.D{{"$match", bson.M{fieldName: true}}})

				localMatcher = nil
			}

			if matchBody == nil && localMatcher != nil {
				// no pieces of the matcher are able to be pushed down,
				// so there is no change in the operator tree.
				v.logger.Warnf(log.DebugHigh, "cannot push down filter expression: %v", filter.matcher.String())
				return filter, nil
			}
		}
	}

	// if we end up here, it's because we have messed with the pipeline
	// in the current table scan operator, so we need to reconstruct the
	// operator nodes.
	ms = ms.clone()
	ms.pipeline = pipeline

	if localMatcher != nil {
		// we ended up here because we have a predicate
		// that can be partially pushed down, so we construct
		// a new filter with only the part remaining that
		// cannot be pushed down.
		return NewFilterStage(ms, localMatcher, filter.requiredColumns), nil
	}

	// everything was able to be pushed down, so the filter
	// is removed from the plan.
	return ms, nil
}

const (
	groupID             = "_id"
	groupDistinctPrefix = "distinct "
	groupTempTable      = ""
)

// visitGroupBy works by using a visitor to systematically visit and replace certain SQLExpr while generating
// $group and $project stages for the aggregation pipeline.
func (v *pushDownOptimizer) visitGroupBy(gb *GroupByStage) (PlanStage, error) {

	ms, ok := v.canPushDown(gb.source)
	if !ok {
		return gb, nil
	}

	pipeline := ms.pipeline

	// 1. Translate keys
	keys, err := v.translateGroupByKeys(gb.keys, ms.mappingRegistry.lookupFieldName)
	if err != nil {
		v.logger.Warnf(log.DebugHigh, "cannot translate group by keys: %v", err)
		return gb, nil
	}

	// 2. Translate aggregations
	result, err := v.translateGroupByAggregates(gb.keys, gb.projectedColumns, ms.mappingRegistry.lookupFieldName)
	if err != nil {
		v.logger.Warnf(log.DebugHigh, "cannot translate group by aggregates: %v", err)
		return gb, nil
	}

	result.group[groupID] = keys

	// 3. Translate the final project
	project, err := v.translateGroupByProject(result.exprs, result.mappingRegistry.lookupFieldName)
	if err != nil {
		v.logger.Warnf(log.DebugHigh, "cannot translate group by project: %v", err)
		return gb, nil
	}

	// 4. append to the pipeline
	pipeline = append(pipeline, bson.D{{"$group", result.group}})
	pipeline = append(pipeline, bson.D{{"$project", project}})

	// 5. Fix up the TableScan operator - None of the current registrations in mappingRegistry are valid any longer.
	// We need to clear them out and re-register the new columns.
	mappingRegistry := &mappingRegistry{}
	for _, projectedColumn := range gb.projectedColumns {
		// at this point, our project has all the stringified names of the select expressions, so we need to re-map them
		// each column to its new MongoDB name. This process is what makes the push-down transparent to subsequent operators
		// in the tree that either haven't yet been pushed down, or cannot be. Either way, we output of a push-down must be
		// exactly the same as the output of a non-pushed-down group.
		mappingRegistry.addColumn(projectedColumn.Column)
		mappingRegistry.registerMapping(
			projectedColumn.Table,
			projectedColumn.Name,
			sanitizeFieldName(projectedColumn.Expr.String()),
		)
	}

	ms = ms.clone()
	ms.pipeline = pipeline
	ms.mappingRegistry = mappingRegistry

	return ms, nil
}

// translateGroupByKeys takes the key expressions and generates an _id document. All keys, even single keys,
// will be nested underneath the '_id' field. In addition, each field's name will be the stringified
// version of its SQLExpr.
// For example, 'select a, b from foo group by c' will generate an id document that looks like this:
//      _id: { foo_DOT_c: "$c" }
//
// Likewise, multiple columns will generate something similar.
// For example, 'select a, b from foo group by c,d' will generate an id document that looks like this:
//      _id: { foo_DOT_c: "$c", foo_DOT_d: "$c" }
//
// Finally, anything other than a column will still generate similarly.
// For example, 'select a, b from foo group by c+d' will generate an id document that looks like this:
//      _id: { "foo_DOT_c+foo_DOT_d": { $add: ["$c", "$d"] } }
//
// All projected names are the fully qualified name from SQL, ignoring the mongodb name except for when
// referencing the underlying field.
func (v *pushDownOptimizer) translateGroupByKeys(keys []SQLExpr, lookupFieldName fieldNameLookup) (bson.D, error) {

	keyDocumentElements := bson.D{}

	t := &pushDownTranslator{
		versionAtLeast:  v.ctx.Variables().MongoDBInfo.VersionAtLeast,
		lookupFieldName: lookupFieldName,
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
	group           bson.M
	exprs           []*namedExpr
	mappingRegistry *mappingRegistry
}

type namedExpr struct {
	name string
	expr SQLExpr
}

// translateGroupByAggregates takes the key expressions and the select expressions and generates a
// $group stage. It does this by employing a visitor that walks each of the select expressions individually
// and, depending on the type of expression, generates a full solution or a partial solution to accomplishing
// the goal. For example, the query 'select sum(a) from foo' can be fully realized with a single $group, where
// as 'select sum(distinct a) from foo' requires a $group which adds 'a' to a set ($addToSet) and a subsequent
// $project which then does a $sum on the set created in the $group. Currently, we always generate the $project
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

	// represents all the expressions that should be passed on to $project such that translateGroupByProject
	// is able to do its work without redoing a bunch of the conditionals and special casing here.
	namedExprs := []*namedExpr{}

	// translator will "accumulate" all the group fields. Below, we iterate over each select expressions, which
	// account for all the fields that need to be present in the $group.
	translator := &groupByAggregateTranslator{bson.M{}, v.ctx, isGroupKey, lookupFieldName, &mappingRegistry{}}

	for _, projectedColumn := range projectedColumns {

		newExpr, err := translator.visit(projectedColumn.Expr)
		if err != nil {
			return nil, err
		}

		namedExpr := &namedExpr{
			name: sanitizeFieldName(projectedColumn.Expr.String()),
			expr: newExpr.(SQLExpr),
		}

		namedExprs = append(namedExprs, namedExpr)
	}

	return &translateGroupByAggregatesResult{translator.group, namedExprs, translator.mappingRegistry}, nil
}

type groupByAggregateTranslator struct {
	group           bson.M
	ctx             ConnectionCtx
	isGroupKey      func(SQLExpr) bool
	lookupFieldName fieldNameLookup
	mappingRegistry *mappingRegistry
}

const (
	sumAggregateCountSuffix = "_count"
)

// Visit recursively visits each expression in the tree, adds the relavent $group entries, and returns
// an expression that can be used to generate a subsequent $project.
func (v *groupByAggregateTranslator) visit(n node) (node, error) {
	t := &pushDownTranslator{
		versionAtLeast:  v.ctx.Variables().MongoDBInfo.VersionAtLeast,
		lookupFieldName: v.lookupFieldName,
	}
	switch typedN := n.(type) {
	case SQLColumnExpr:
		fieldName, ok := v.lookupFieldName(typedN.tableName, typedN.columnName)
		if !ok {
			return nil, fmt.Errorf("could not map %v.%v to a field", typedN.tableName, typedN.columnName)
		}
		if !v.isGroupKey(typedN) {
			// since it's not an aggregation function, this implies that it takes the first value of the column.
			// So project the field, and register the mapping.
			v.group[sanitizeFieldName(typedN.String())] = bson.M{"$first": getProjectedFieldName(fieldName, typedN.Type())}
			v.mappingRegistry.registerMapping(typedN.tableName, typedN.columnName, sanitizeFieldName(typedN.String()))
		} else {
			// the _id is added to the $group in translateGroupByKeys. We will only be here if the user has also projected
			// the group key, in which we'll need this to look it up in translateGroupByProject under its name. Hence, all
			// we need to do is register the mapping.
			v.mappingRegistry.registerMapping(typedN.tableName, typedN.columnName, groupID+"."+sanitizeFieldName(typedN.String()))
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
						groupTempTable,
						fieldName,
						typedN.Type(),
						schema.MongoNone,
					),
				},
			}
			v.group[fieldName] = bson.M{"$addToSet": trans}
			v.mappingRegistry.registerMapping(groupTempTable, fieldName, fieldName)
		} else {
			// Non-distinct aggregation functions are one-step aggregations that can be performed completely by the
			// $group. Here, we simply generate the correct aggregation function for $group and create a new expression
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
			v.mappingRegistry.registerMapping(groupTempTable, fieldName, fieldName)

			if typedN.Name == sumAggregateName {
				// summing a column with all nulls should result in a null sum. However, MongoDB
				// returns 0. So, we'll add in an arbitrary count operator to count the number
				// of non-nulls and, in the following $project, we'll check this to know whether
				// or not to use the sum or to use null.
				countTrans, ok := t.TranslateExpr(typedN.Exprs[0])
				if !ok {
					return nil, fmt.Errorf("could not translate '%v'", typedN.Exprs[0].String())
				}
				countFieldName := sanitizeFieldName(typedN.String() + sumAggregateCountSuffix)
				v.group[countFieldName] = getCountAggregation(countTrans)
				v.mappingRegistry.registerMapping(groupTempTable, countFieldName, countFieldName)

				newExpr = NewIfScalarFunctionExpr(
					NewSQLColumnExpr(0, groupTempTable, countFieldName, schema.SQLInt64, schema.MongoNone),
					NewSQLColumnExpr(0, groupTempTable, fieldName, typedN.Type(), schema.MongoNone),
					SQLNull,
				)
			} else {
				newExpr = NewSQLColumnExpr(0, groupTempTable, fieldName, typedN.Type(), schema.MongoNone)
			}

		}

		return newExpr, nil

	case SQLExpr:
		if v.isGroupKey(typedN) {
			// the _id is added to the $group in translateGroupByKeys. We will only be here if the user has also projected
			// the group key, in which we'll need this to look it up in translateGroupByProject under its name. In this,
			// we need to create a new expr that is simply a field pointing at the nested identifier and register that
			// mapping.
			fieldName := sanitizeFieldName(typedN.String())
			newExpr := NewSQLColumnExpr(0, groupTempTable, fieldName, typedN.Type(), schema.MongoNone)
			v.mappingRegistry.registerMapping(groupTempTable, fieldName, groupID+"."+fieldName)
			return newExpr, nil
		}

		// We'll only get here for two-step expressions. An example is
		// 'select a + b from foo group by a' or 'select b + sum(c) from foo group by a'. In this case,
		// we'll descend into the tree recursively which will build up the $group for the necessary pieces.
		// Finally, return the now changed expression such that $project can act on them appropriately.
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

// translateGroupByProject takes the expressions and generates a $project. All the hard work was done in
// translateGroupByAggregates, so this is simply a process of either adding a field to the $project, or
// completing two-step aggregations. Two-step aggregations that needs completing are expressions like
// 'sum(distinct a)' or 'a + b' where b was part of the group key.
func (v *pushDownOptimizer) translateGroupByProject(exprs []*namedExpr, lookupFieldName fieldNameLookup) (bson.M, error) {
	project := bson.M{groupID: 0}

	t := &pushDownTranslator{
		versionAtLeast:  v.ctx.Variables().MongoDBInfo.VersionAtLeast,
		lookupFieldName: lookupFieldName,
	}

	for _, expr := range exprs {

		switch typedE := expr.expr.(type) {
		case SQLColumnExpr:
			// Any one-step aggregations will end up here as they were fully performed in the $group. So, simple
			// column references ('select a') and simple aggregations ('select sum(a)').
			fieldName, ok := lookupFieldName(typedE.tableName, typedE.columnName)
			if !ok {
				return nil, fmt.Errorf("unable to get a field name for %v.%v", typedE.tableName, typedE.columnName)
			}

			project[expr.name] = "$" + fieldName
		default:
			// Any two-step aggregations will end up here to complete the second step.
			trans, ok := t.TranslateExpr(expr.expr)
			if !ok {
				return nil, fmt.Errorf("unable to translate '%v'", expr.expr.String())
			}
			project[expr.name] = trans
		}
	}

	return project, nil
}

const (
	joinedFieldNamePrefix    = "__joined_"
	leftJoinExcludeFieldName = "__leftjoin_exclude"
)

// buildRemainingPredicateForLeftJoin will return 2 items; first a $project to put before the unwind, and a $match to put after the unwind.
func (v *pushDownOptimizer) buildRemainingPredicateForLeftJoin(leftMappingRegistry, combinedMappingRegistry *mappingRegistry, remainingPredicate SQLExpr, asField string, preserveIndex bool) (bson.D, bson.D, bool) {

	fixedLookupFieldName := func(tbl, col string) (string, bool) {
		fieldName, ok := combinedMappingRegistry.lookupFieldName(tbl, col)
		if !ok {
			return "", false
		}

		// inside a $filter and $map (which use the result of this function), columns with the asField
		// prefix will have their prefix renamed. As such, we need to intercept this call and handle
		// that translation early. For instance, if the asField as $b.child and the field ends up as
		// $b.child.myField, then the result will be $$this.myField.
		if strings.HasPrefix(fieldName, asField) {
			fieldName = strings.Replace(fieldName, asField, "$this", 1)
		}

		return fieldName, true
	}

	t := &pushDownTranslator{
		versionAtLeast:  v.ctx.Variables().MongoDBInfo.VersionAtLeast,
		lookupFieldName: fixedLookupFieldName,
	}
	ifPart, ok := t.TranslateExpr(remainingPredicate)
	if !ok {
		v.logger.Warnf(log.DebugHigh, "cannot translate remaining left join predicate %#v", remainingPredicate)
		return nil, nil, false
	}

	var projectBody bson.M
	var match bson.D
	if preserveIndex {
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
										"if":   bson.M{"$isArray": "$" + asField},
										"then": "$" + asField,
										"else": []interface{}{"$" + asField},
									},
								},
								"as": "this",
								"in": bson.M{
									"$cond": bson.M{
										"if":   ifPart,
										"then": "$$this",
										"else": bson.M{
											leftJoinExcludeFieldName: bson.M{
												"$literal": true,
											},
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
		// If the predicate doesn't pass, then we set the 'as' field to nil
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

	// we now need to make sure we project all the existing columns from the local mongo source
	for _, c := range leftMappingRegistry.columns {
		field, ok := leftMappingRegistry.lookupFieldName(c.Table, c.Name)
		if !ok {
			v.logger.Warnf(log.DebugHigh, "cannot find referenced join column %#v in local lookup", c)
			return nil, nil, false
		}
		projectBody[field] = 1
	}

	return bson.D{{"$project", projectBody}}, match, true
}

func (v *pushDownOptimizer) mergeTables(msLocal, msForeign *MongoSourceStage, join *JoinStage) (PlanStage, error) {
	filterJoinCriteria := func(e SQLExpr, keepJoinCriteria bool) []SQLExpr {
		exprs, newExprs := splitExpression(e), []SQLExpr{}
		registries := []*mappingRegistry{msLocal.mappingRegistry,
			msForeign.mappingRegistry}
		for _, expr := range exprs {
			if equalExpr, ok := expr.(*SQLEqualsExpr); ok {
				c1, _ := equalExpr.left.(SQLColumnExpr)
				c2, _ := equalExpr.right.(SQLColumnExpr)
				if c1.selectID == c2.selectID {

					c1Name, c1RegistryIdx, ok := lookupSQLColumn(
						c1.tableName, c1.columnName, registries)
					if !ok {
						panic("Unable to find field mapping for merge " +
							"c1. This should never happen.")
					}

					c2Name, c2RegistryIdx, ok := lookupSQLColumn(
						c2.tableName, c2.columnName, registries)
					if !ok {
						panic("Unable to find field mapping for merge " +
							"c2. This should never happen.")
					}

					c1IsPK := registries[c1RegistryIdx].
						isPrimaryKey(c1.columnName)
					c2IsPK := registries[c2RegistryIdx].
						isPrimaryKey(c2.columnName)

					if c1IsPK && c2IsPK && c1Name == c2Name {
						// use columns from local table
						if keepJoinCriteria {
							if containsString(msLocal.aliasNames,
								c1.tableName) {
								newExprs = append(newExprs, c1)
							} else {
								newExprs = append(newExprs, c2)
							}
						}
					} else if !keepJoinCriteria {
						newExprs = append(newExprs, expr)
					}
				} else if !keepJoinCriteria {
					newExprs = append(newExprs, expr)
				}
			} else if !keepJoinCriteria {
				newExprs = append(newExprs, expr)
			}
		}
		return newExprs
	}

	newPipeline, err := v.mergePipeline(msLocal, msForeign, join.kind)
	if err != nil {
		v.logger.Logf(log.DebugHigh, "cannot merge pipelines: %v", err)
		return nil, nil
	}

	ms := msLocal.clone()
	ms.selectIDs = append(ms.selectIDs, msForeign.selectIDs...)
	ms.aliasNames = append(ms.aliasNames, msForeign.aliasNames...)
	ms.tableNames = append(ms.tableNames, msForeign.tableNames...)
	ms.collectionNames = append(ms.collectionNames,
		msForeign.collectionNames...)

	newMappingRegistry := ms.mappingRegistry.copy()
	newMappingRegistry.columns = append(newMappingRegistry.columns,
		msForeign.mappingRegistry.columns...)
	if msForeign.mappingRegistry.fields != nil {
		for tableName, columns := range msForeign.mappingRegistry.fields {
			for columnName, fieldName := range columns {
				newMappingRegistry.registerMapping(tableName, columnName,
					fieldName)
			}
		}
	}

	ms.mappingRegistry, ms.pipeline = newMappingRegistry, newPipeline.stages

	remainingPredicate := combineExpressions(filterJoinCriteria(join.matcher,
		false))
	if remainingPredicate != nil {
		if join.kind == InnerJoin || join.kind == StraightJoin {
			v.logger.Logf(log.DebugHigh, "join merge: creating filter "+
				"stage for remaining predicate: %v",
				remainingPredicate.String())
			f, err := v.visit(NewFilterStage(ms, remainingPredicate,
				join.requiredColumns))
			if err != nil {
				return nil, err
			}
			return f.(PlanStage), nil
		}

		// this "predicate" must get inserted before the unwinds from
		// the right side.
		insertionPoint := len(msLocal.pipeline)
		project, match, ok := v.buildRemainingPredicateForLeftJoin(
			msLocal.mappingRegistry,
			newMappingRegistry,
			remainingPredicate,
			newPipeline.arrayPaths[insertionPoint],
			true,
		)
		if !ok {
			return join, nil
		}

		if match != nil {
			ms.pipeline = append(ms.pipeline, match)
		}

		// insert project after the first
		ms.pipeline = append(ms.pipeline, nil)
		copy(ms.pipeline[insertionPoint+1:], ms.pipeline[insertionPoint:])
		ms.pipeline[insertionPoint] = project
	}

	return ms, nil
}

func (v *pushDownOptimizer) meetsMergePKCriteria(local, foreign *MongoSourceStage, matcher SQLExpr) bool {
	// don't perform optimization on MongoDB views as
	// renames might have occured on fields.
	if local.isView() {
		v.logger.Logf(log.DebugHigh, "cannot merge join stage, local "+
			"table is MongoDB view")
		return false
	}

	if foreign.isView() {
		v.logger.Logf(log.DebugHigh, "cannot merge join stage, foreign "+
			"table is MongoDB view")
		return false
	}

	exprs := splitExpression(matcher)

	numPK := func(columns []*Column, table string) int {
		n, keys := 0, make(map[string]struct{})
		for _, c := range columns {
			if _, counted := keys[c.Name]; !counted && c.PrimaryKey &&
				c.Table == table {
				n, keys[c.Name] = n+1, struct{}{}
			}
		}
		return n
	}

	// When we attempt to merge two different arrays (tables) of
	// the same underlying MongoDB collection, we only require
	// one primary key equality match.
	//
	// For example, the join criteria "test_array1 JOIN test_array1"
	// requires as many primary keys as are specified in the DRDL
	// file for "test_array1". However, if we attempted to merge
	// "test_array1 JOIN test_array2", instead we'd only need an
	// equality on the base table's primary key field.
	//
	isSameArrayJoin := local.tableNames[0] == foreign.tableNames[0]
	numRequiredPKConjunctions := 1

	if isSameArrayJoin && len(local.tableNames) == 1 {
		numLocalPK := numPK(local.mappingRegistry.columns, local.aliasNames[0])
		numForeignPK := numPK(foreign.mappingRegistry.columns, foreign.aliasNames[0])
		numRequiredPKConjunctions = util.MinInt(numLocalPK, numForeignPK)
	}

	if numRequiredPKConjunctions == 0 {
		v.logger.Logf(log.DebugHigh, "cannot merge join stage, table "+
			"has no primary key")
		return false
	}

	numPKConjunctions := 0

	v.logger.Logf(log.DebugHigh, "join merge: examining match criteria...")

	registries := []*mappingRegistry{
		local.mappingRegistry,
		foreign.mappingRegistry,
	}

	seenPrimaryKeys := make(map[string]struct{})

	for _, expr := range exprs {
		if equalExpr, ok := expr.(*SQLEqualsExpr); ok {
			column1, _ := equalExpr.left.(SQLColumnExpr)
			column2, _ := equalExpr.right.(SQLColumnExpr)

			invalidLeftColumn := !containsString(local.aliasNames,
				column1.tableName) &&
				!containsString(foreign.aliasNames, column1.tableName)
			invalidRightColumn := !containsString(local.aliasNames,
				column2.tableName) &&
				!containsString(foreign.aliasNames, column2.tableName)

			if invalidLeftColumn || invalidRightColumn {
				v.logger.Logf(log.DebugHigh, "join merge: found unexpected "+
					"table references, moving on...")
				continue
			}

			if column1.selectID != column2.selectID {
				v.logger.Logf(log.DebugHigh, "join merge: found unmatched "+
					"select identifiers (%v and %v), moving on...",
					column1.selectID, column2.selectID)
				continue
			}

			columnOneName, c1RegistryIdx, ok := lookupSQLColumn(
				column1.tableName, column1.columnName, registries)
			if !ok {
				panic("Unable to find field mapping for merge column1. " +
					"This should never happen.")
			}

			columnTwoName, c2RegistryIdx, ok := lookupSQLColumn(
				column2.tableName, column2.columnName, registries)
			if !ok {
				panic("Unable to find field mapping for merge column2. " +
					"This should never happen.")
			}

			c1IsPK := registries[c1RegistryIdx].isPrimaryKey(column1.columnName)
			c2IsPK := registries[c2RegistryIdx].isPrimaryKey(column2.columnName)

			if !c1IsPK || !c2IsPK {
				v.logger.Logf(log.DebugHigh, "join merge: criteria contains "+
					"non-primary key (%v and %v), moving on...",
					column1.String(), column2.String())
				continue
			}

			if columnOneName != columnTwoName {
				v.logger.Logf(log.DebugHigh, "join merge: criteria contains "+
					"unmatched primary keys (%v and %v), moving on...",
					columnOneName, columnTwoName)
				continue
			}

			if _, ok := seenPrimaryKeys[columnOneName]; ok {
				v.logger.Logf(log.DebugHigh, "join merge: ignoring duplicate "+
					"primary key criteria '%v' and moving on...",
					column1.String())
				continue
			}

			seenPrimaryKeys[columnOneName] = struct{}{}

			numPKConjunctions++
		}
	}

	if numPKConjunctions < numRequiredPKConjunctions {
		v.logger.Logf(log.DebugHigh, "join merge: criteria conjunction "+
			"contains %v unique primary key equality %v (need %v)",
			numPKConjunctions, util.Pluralize(numPKConjunctions, "pair",
				"pairs"), numRequiredPKConjunctions)
		return false
	}

	return true
}

func (v *pushDownOptimizer) sharesRootTable(local, foreign *MongoSourceStage) bool {
	baseCollectionName := local.collectionNames[0]

	v.logger.Logf(log.DebugHigh, "attempting to merge tables %v and %v",
		local.aliasNames, foreign.aliasNames)

	for _, collectionName := range append(local.collectionNames[1:],
		foreign.collectionNames...) {
		if collectionName != baseCollectionName {
			v.logger.Logf(log.DebugHigh, "cannot merge join stage, foreign "+
				"table has pipeline with different root tables: %v and %v",
				collectionName, baseCollectionName)
			return false
		}
	}

	return true
}

func (v *pushDownOptimizer) visitJoin(join *JoinStage) (PlanStage, error) {

	if join.matcher == nil {
		v.logger.Warnf(log.DebugHigh, "cannot push down join stage, matcher is nil")
		return join, nil
	}

	// 1. the join type must be usable. MongoDB can only do an inner join and a left outer join.
	// While we can flip a right outer join to a left outer join, because we don't have information
	// about the target collections, we need to indicate to users that the right-side may NOT be
	// sharded. Flipping these around would make this difficult for a user to fix. Instead, we'll
	// just tell them to make their right outer joins -> left outer joins.
	var localSource, foreignSource PlanStage
	var joinKind JoinKind

	switch join.kind {
	case InnerJoin:
		localSource = join.left
		foreignSource = join.right
		joinKind = InnerJoin
	case LeftJoin:
		localSource = join.left
		foreignSource = join.right
		joinKind = LeftJoin
	case StraightJoin:
		localSource = join.left
		foreignSource = join.right
		joinKind = StraightJoin
	default:
		v.logger.Warnf(log.DebugHigh, "cannot push down %v", join.kind)
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

	// Before attempting to merge, check that the underlying collection
	// is the same for both tables and that the join criteria holds the
	// primary key for both
	if v.sharesRootTable(msLocal, msForeign) &&
		v.meetsMergePKCriteria(msLocal, msForeign, join.matcher) {
		ms, err := v.mergeTables(msLocal, msForeign, join)
		if err != nil {
			return nil, err
		}

		if ms != nil {
			v.logger.Logf(log.DebugHigh, "successfully merged tables %v "+
				"and %v", msLocal.aliasNames, msForeign.aliasNames)
			return ms, nil
		}

		v.logger.Logf(log.DebugHigh, "unable to merge tables %v and %v",
			msLocal.aliasNames, msForeign.aliasNames)
	}

	v.logger.Warnf(log.DebugHigh, "attempting to translate join stage "+
		"to lookup")

	lenForeignPipeline := len(msForeign.pipeline)
	foreignHasUnwind := false

	if lenForeignPipeline > 0 {
		_, foreignHasUnwind = msForeign.pipeline[0].Map()["$unwind"]
		if !foreignHasUnwind || lenForeignPipeline > 1 {
			v.logger.Warnf(log.DebugHigh, "unable to translate join "+
				"stage to lookup: foreign table has pipeline: %#v",
				msForeign.pipeline)
			return join, nil
		}
	}

	// 3. find the local column and the foreign column
	lookupInfo, err := getLocalAndForeignColumns(msLocal, msForeign,
		join.matcher)
	if err != nil {
		v.logger.Warnf(log.DebugHigh, "unable to translate join "+
			"stage to lookup: %v", err)
		return join, nil
	}

	// prevent join pushdown when UUID subtype 3 encoding is different
	localMongoType := lookupInfo.localColumn.columnType.MongoType
	foreignMongoType := lookupInfo.foreignColumn.columnType.MongoType

	if isUUID(localMongoType) && isUUID(foreignMongoType) {
		if localMongoType != foreignMongoType {
			v.logger.Warnf(log.DebugHigh, "unable to translate join "+
				"stage to lookup: found different criteria UUID - %v "+
				"and %v", localMongoType, foreignMongoType)
			return join, nil
		}
	}

	// 4. get lookup fields
	localFieldName, ok := msLocal.mappingRegistry.lookupFieldName(
		lookupInfo.localColumn.tableName, lookupInfo.localColumn.columnName)
	if !ok {
		v.logger.Errf(log.DebugHigh, "cannot find referenced local join "+
			"column %#v in lookup", lookupInfo.localColumn)
		return join, nil
	}

	foreignFieldName, ok := msForeign.mappingRegistry.lookupFieldName(
		lookupInfo.foreignColumn.tableName, lookupInfo.foreignColumn.columnName)
	if !ok {
		v.logger.Errf(log.DebugHigh, "cannot find referenced foreign join "+
			"column %#v in lookup", lookupInfo.foreignColumn)
		return join, nil
	}

	asField := sanitizeFieldName(joinedFieldNamePrefix + msForeign.aliasNames[0])

	// 5. compute all the mappings from the msForeign mapping registry
	// to be nested under the 'asField' we used above.
	newMappingRegistry := msLocal.mappingRegistry.copy()

	newMappingRegistry.columns = append(newMappingRegistry.columns,
		msForeign.mappingRegistry.columns...)
	if msForeign.mappingRegistry.fields != nil {
		for tableName, columns := range msForeign.mappingRegistry.fields {
			for columnName, fieldName := range columns {
				newMappingRegistry.registerMapping(tableName, columnName,
					asField+"."+fieldName)
			}
		}
	}

	// 6. build the pipeline
	pipeline := msLocal.pipeline

	if joinKind == InnerJoin || joinKind == StraightJoin {
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

	if joinKind == LeftJoin {
		// Because MongoDB does not compare nulls in the same way as MySQL, we need an extra
		// $project to ensure account for this incompatibility. Effectively, when our left
		// hand field is null, we'll empty the joined results prior to unwinding.
		project := bson.M{}

		// enumerate all local fields
		for _, c := range msLocal.mappingRegistry.columns {
			fieldName, ok := msLocal.mappingRegistry.lookupFieldName(
				c.Table, c.Name)
			if !ok {
				panic("Unable to find field mapping for column. This " +
					"should never happen.")
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
			"preserveNullAndEmptyArrays": joinKind == LeftJoin,
		},
	}}

	lookupOnArrayField := false

	if foreignHasUnwind {
		foreignMapped := msForeign.pipeline[0].Map()["$unwind"].(bson.D).Map()
		oldForeignPath := fmt.Sprintf("%v", foreignMapped["path"])
		lookupOnArrayField = strings.Split(foreignFieldName, ".")[0] == oldForeignPath[1:]
	}

	if lookupInfo.remainingPredicate != nil && joinKind == LeftJoin {
		if lookupOnArrayField && len(strings.Split(foreignFieldName, ".")) > 1 {
			v.logger.Warnf(log.DebugHigh, "unable to translate left join "+
				"stage to lookup: lookup on nested array field")
			return join, nil
		}

		project, match, ok := v.buildRemainingPredicateForLeftJoin(
			msLocal.mappingRegistry,
			newMappingRegistry,
			lookupInfo.remainingPredicate,
			asField,
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

		// strip dollar sign prefix
		if path != "" {
			path = fmt.Sprintf("$%v.%v", asField, path[1:])
		} else {
			v.logger.Errf(log.DebugHigh, "empty $unwind path specification")
			return join, nil
		}

		// For left joins, preserve null and empty arrays to ensure
		// that every local pipeline record gets returned
		idx := fmt.Sprintf("%v.%v", asField, foreignMapped["includeArrayIndex"])
		unwind := bson.D{
			{"path", path},
			{"includeArrayIndex", idx},
		}

		unwind = append(unwind, bson.DocElem{"preserveNullAndEmptyArrays",
			join.kind == LeftJoin})

		v.logger.Warnf(log.DebugHigh, "consolidating foreign unwind "+
			"into local pipeline")

		pipeline = append(pipeline, bson.D{{"$unwind", unwind}})

		// Handle edge case where the lookup field is the same as the
		// $unwind's array path. In this case, we must apply an
		// additional filter to remove records in the now unwound array
		// that don't meet the lookup criteria.
		if lookupOnArrayField {
			foreignField := fmt.Sprintf("$%v.%v", asField, foreignFieldName)
			filter := bson.M{mgoOperatorEQ: []interface{}{
				foreignField, "$" + localFieldName,
			}}

			projectBody := bson.M{}

			for _, columns := range newMappingRegistry.fields {
				for _, fieldName := range columns {
					projectBody[fieldName] = 1
				}
			}

			fieldName := v.distinctProjectFieldName(projectBody)
			projectBody[fieldName] = filter

			match := bson.M{fieldName: true}
			if join.kind == LeftJoin {
				// for left joins, we need to ensure we retain records
				// from the left child - in case the unwound array was
				// empty or null
				match = bson.M{mgoOperatorOR: []interface{}{match,
					bson.M{path[1:]: bson.M{mgoOperatorExists: false}}}}
			}
			pipeline = append(pipeline,
				bson.D{{"$project", projectBody}},
				bson.D{{"$match", match}},
			)
		}
	}

	// 7. build the new operators
	msLocal.aliasNames = append(msLocal.aliasNames, msForeign.aliasNames...)
	msLocal.tableNames = append(msLocal.tableNames, msForeign.tableNames...)
	msLocal.collectionNames = append(msLocal.collectionNames, msForeign.collectionNames...)
	ms := msLocal.clone()
	ms.selectIDs = append(ms.selectIDs, msForeign.selectIDs...)
	ms.pipeline = pipeline
	ms.mappingRegistry = newMappingRegistry

	if lookupInfo.remainingPredicate != nil && (joinKind == InnerJoin || joinKind == StraightJoin) {
		f, err := v.visit(NewFilterStage(ms, lookupInfo.remainingPredicate, join.requiredColumns))
		if err != nil {
			return nil, err
		}

		return f.(PlanStage), nil
	}

	v.logger.Warnf(log.DebugHigh, "successfully translated join stage to lookup")

	return ms, nil
}

type lookupInfo struct {
	localColumn        *SQLColumnExpr
	foreignColumn      *SQLColumnExpr
	remainingPredicate SQLExpr
}

func getLocalAndForeignColumns(localTable, foreignTable *MongoSourceStage, e SQLExpr) (*lookupInfo, error) {
	exprs := splitExpression(e)
	// find a SQLEqualsExpr where the left and right are columns from the local and foreign tables
	for i, expr := range exprs {
		if equalExpr, ok := expr.(*SQLEqualsExpr); ok {
			// we must have a field from the left table and a field from the right table
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

func lookupSQLColumn(tableName, columnName string, mappingRegistries []*mappingRegistry) (string, int, bool) {
	for i, registry := range mappingRegistries {
		fieldName, ok := registry.lookupFieldName(tableName, columnName)
		if ok {
			return fieldName, i, true
		}
	}
	return "", 0, false
}

type consolidatedPipeline struct {
	stages           []bson.D
	arrayPaths       []string
	arrayPathIndexes []string
}

func (v *pushDownOptimizer) mergePipeline(local, foreign *MongoSourceStage, joinKind JoinKind) (*consolidatedPipeline, error) {
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
			fieldName, ok := foreign.mappingRegistry.lookupFieldName(c.Table, c.Name)
			if !ok {
				return nil, fmt.Errorf("cannot find referenced foreign column %v.%v in projection lookup", c.Table, c.Name)
			}

			if _, ok := project[fieldName]; !ok && !prefixPathPresent(project, fieldName) {
				v.logger.Logf(log.DebugHigh, "augmenting local table with column '%v'.'%v'", c.Table, c.Name)
				project[fieldName] = 1
				foreign.mappingRegistry.registerMapping(c.Table, c.Name, fieldName)
			}
		}

		return bson.D{{"$project", project}}, nil
	}

	getPathsAndPipeline := func(stages []bson.D, isLocal bool) error {
		for _, stage := range stages {
			bsonStage, ok := stage.Map()["$unwind"]
			if !ok {
				if isLocal {
					// for projections, ensure all foreign columns are included
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
					return fmt.Errorf("found unmergeable stage in foreign table (%v) pipeline: %#v", foreign.aliasNames, stage)
				}
			}

			unwind := bsonStage.(bson.D)
			fields := unwind.Map()

			iPath, ok := fields["path"]
			if !ok {
				return fmt.Errorf("could not find unwind path in foreign table %v: %#v", foreign.aliasNames, stage)
			}

			iIndex, ok := fields["includeArrayIndex"]
			if !ok {
				return fmt.Errorf("could not find unwind array index in foreign table %v: %#v", foreign.aliasNames, stage)
			}

			// strip dollar sign prefix
			path := fmt.Sprintf("%v", iPath)
			if path != "" {
				path = path[1:]
			}

			arrayIdx := fmt.Sprintf("%v", iIndex)

			if !util.StringSliceContains(pipeline.arrayPathIndexes, arrayIdx) {
				pipeline.arrayPaths = append(pipeline.arrayPaths, path)
				pipeline.arrayPathIndexes = append(pipeline.arrayPathIndexes, arrayIdx)
				_, ok = fields["preserveNullAndEmptyArrays"]
				if !ok && joinKind == LeftJoin && !isLocal {
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

	return pipeline, nil
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

	for _, term := range orderBy.terms {

		var tableName, columnName string

		switch typedE := term.expr.(type) {
		case SQLColumnExpr:
			tableName, columnName = typedE.tableName, typedE.columnName
		default:
			// MongoDB only allows sorting by a field, so pushing down a
			// non-field requires it to be pre-calculated by a previous
			// push down. If it has been pre-calculated, then it will
			// exist in the mapping registry. Otherwise, it won't, and
			// the order by cannot be pushed down.
			columnName = typedE.String()
		}

		fieldName, ok := ms.mappingRegistry.lookupFieldName(tableName, columnName)
		if !ok {
			v.logger.Warnf(log.DebugHigh, "cannot find referenced order by column %v.%v in lookup", tableName, columnName)
			return orderBy, nil
		}
		direction := 1
		if !term.ascending {
			direction = -1
		}
		sort = append(sort, bson.DocElem{fieldName, direction})
	}

	pipeline := ms.pipeline
	pipeline = append(pipeline, bson.D{{"$sort", sort}})

	ms = ms.clone()
	ms.pipeline = pipeline
	return ms, nil
}

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

	// if we have a projection that results in multiple columns with the same qualifier and name,
	// we cannot push this down.
	uniqueProjectedColumns := project.projectedColumns.Unique()
	if len(uniqueProjectedColumns) != len(project.projectedColumns) {
		return project, nil
	}

	t := &pushDownTranslator{
		versionAtLeast:  v.ctx.Variables().MongoDBInfo.VersionAtLeast,
		lookupFieldName: ms.mappingRegistry.lookupFieldName,
	}

	for _, projectedColumn := range project.projectedColumns {

		// Convert the column's SQL expression into an expression in mongo query language.
		projectedField, ok := t.TranslateExpr(projectedColumn.Expr)
		if !ok {
			// Expression can't be translated, so it can't be projected.
			// We skip it and leave this Project node in the query plan so that it still gets
			// evaluated during execution.
			canReplaceProject = false

			// There might still be fields referenced in this expression
			// that we still need to project, so collect them and add them to the projection.
			refdCols, err := referencedColumns(v.selectIDsInScope, projectedColumn.Expr)
			if err != nil {
				v.logger.Warnf(log.DebugHigh, "cannot find referenced project expression: %v", err)
				return nil, err
			}

			for _, refdCol := range refdCols {
				refdCol.PrimaryKey = projectedColumn.PrimaryKey
				fieldName, ok := ms.mappingRegistry.lookupFieldName(refdCol.Table, refdCol.Name)
				if !ok {
					v.logger.Warnf(log.DebugHigh, "cannot find referenced column %#v in lookup", refdCol)
					return project, nil
				}

				safeFieldName := sanitizeFieldName(fieldName)
				fieldsToProject[safeFieldName] = getProjectedFieldName(fieldName, refdCol.SQLType)
				fixedMappingRegistry.addColumn(refdCol)
				fixedMappingRegistry.registerMapping(refdCol.Table, refdCol.Name, safeFieldName)
			}

			fixedProjectedColumns = append(fixedProjectedColumns, projectedColumn)
		} else {
			safeFieldName := sanitizeFieldName(projectedColumn.Expr.String())
			fieldsToProject[safeFieldName] = projectedField
			fixedMappingRegistry.addColumn(projectedColumn.Column)
			fixedMappingRegistry.registerMapping(projectedColumn.Table, projectedColumn.Name, safeFieldName)

			columnExpr := NewSQLColumnExpr(projectedColumn.SelectID, projectedColumn.Column.Table, projectedColumn.Column.Name, projectedColumn.SQLType, projectedColumn.MongoType)

			fixedProjectedColumns = append(fixedProjectedColumns,
				ProjectedColumn{
					Column: projectedColumn.Column,
					Expr:   columnExpr,
				},
			)
		}

	}

	if len(fieldsToProject) == 0 {
		v.logger.Warnf(log.DebugHigh, "no fields for project push down")
		return project, nil
	}

	pipeline := append(ms.pipeline, bson.D{{"$project", fieldsToProject}})
	ms = ms.clone()
	ms.pipeline = pipeline
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
		fieldName, ok := ms.mappingRegistry.lookupFieldName(column.Table, column.Name)
		if !ok {
			v.logger.Warnf(log.DebugHigh, "cannot find referenced subquery column %#v in lookup", column)
			return subquery, nil
		}

		mappingRegistry.addColumn(&Column{
			SelectID:  column.SelectID,
			Name:      column.Name,
			Table:     subquery.aliasName,
			SQLType:   column.SQLType,
			MongoType: column.MongoType,
		})

		mappingRegistry.registerMapping(subquery.aliasName, column.Name, fieldName)
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

// referencedColumns will take an expression and return all the columns referenced in the expression
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
			column := &Column{
				SelectID:  typedN.selectID,
				Table:     typedN.tableName,
				Name:      typedN.columnName,
				MongoType: typedN.columnType.MongoType,
				SQLType:   typedN.columnType.SQLType,
			}

			cf.columns = append(cf.columns, column)
		}
		return n, nil
	}

	return walk(cf, n)
}

// pushdownProject is called when a stage could not be pushed down. It uses requiredColumns to create and
// visit a new projectStage in order to project out only the columns needed for the rest of the query so that
// we do not have to pull all data from a table into memory.
func (v *pushDownOptimizer) pushdownProject(requiredColumns []SQLExpr, source PlanStage) (PlanStage, error) {
	sf := &sourceFinder{}
	_, err := sf.visit(source)
	if err != nil {
		return nil, err
	}

	var projectedCols []ProjectedColumn
	for _, expr := range requiredColumns {

		if sqlColExpr, ok := expr.(SQLColumnExpr); ok {
			column := &Column{
				SelectID:   sqlColExpr.selectID,
				Table:      sqlColExpr.tableName,
				Name:       sqlColExpr.columnName,
				SQLType:    sqlColExpr.Type(),
				MongoType:  sqlColExpr.columnType.MongoType,
				PrimaryKey: sf.source.mappingRegistry.isPrimaryKey(sqlColExpr.columnName),
			}
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

func (v *pushDownOptimizer) getRequiredColumnsForJoinSide(tableNames []string, requiredColumns []SQLExpr) []SQLExpr {
	var result []SQLExpr
	for _, expr := range requiredColumns {
		tableName := strings.Split(expr.String(), ".")[0]
		if containsString(tableNames, tableName) {
			result = append(result, expr)
		}
	}

	return result
}

func (v *pushDownOptimizer) projectAllColumns(mr *mappingRegistry) bson.M {
	projectBody := bson.M{}
	for _, c := range mr.columns {
		field, ok := mr.lookupFieldName(c.Table, c.Name)
		if !ok {
			panic("Unable to find field mapping for column. This should never happen.")
		}
		projectBody[field] = 1
	}
	return projectBody
}

// sourceFinder is used within projection pushdown
// to locate the MongoSource stage to project.
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
