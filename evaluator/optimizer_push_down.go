package evaluator

import (
	"fmt"

	"github.com/10gen/sqlproxy/schema"

	"gopkg.in/mgo.v2/bson"
)

func optimizePushDown(planCtx *PlanCtx, o PlanStage) (PlanStage, error) {
	v := &pushDownOptimizer{planCtx}
	return v.Visit(o)
}

type pushDownOptimizer struct {
	planCtx *PlanCtx
}

func (v *pushDownOptimizer) Visit(p PlanStage) (PlanStage, error) {
	p, err := walkPlanTree(v, p)
	if err != nil {
		return nil, err
	}

	switch typedP := p.(type) {
	case *FilterStage:
		p, err = v.visitFilter(typedP)
		if err != nil {
			return nil, fmt.Errorf("unable to optimize filter: %v", err)
		}
	case *GroupByStage:
		p, err = v.visitGroupBy(typedP)
		if err != nil {
			return nil, fmt.Errorf("unable to optimize group by: %v", err)
		}
	case *JoinStage:
		p, err = v.visitJoin(typedP)
		if err != nil {
			return nil, fmt.Errorf("unable to optimize join: %v", err)
		}
	case *LimitStage:
		p, err = v.visitLimit(typedP)
		if err != nil {
			return nil, fmt.Errorf("unable to optimize limit: %v", err)
		}
	case *OrderByStage:
		p, err = v.visitOrderBy(typedP)
		if err != nil {
			return nil, fmt.Errorf("unable to optimize order by: %v", err)
		}
	case *ProjectStage:
		p, err = v.visitProject(typedP)
		if err != nil {
			return nil, fmt.Errorf("unable to optimize project: %v", err)
		}
	}

	return p, nil
}

func (v *pushDownOptimizer) canPushDown(ps PlanStage) (*MongoSourceStage, bool) {

	ms, ok := ps.(*MongoSourceStage)
	if !ok {
		return nil, false
	}

	return ms, true
}

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
			return &EmptyStage{}, nil
		}

		// otherwise, the filter simply gets removed from the tree

	} else {
		var matchBody bson.M
		matchBody, localMatcher = TranslatePredicate(filter.matcher, ms.mappingRegistry.lookupFieldName)
		if matchBody == nil {
			// no pieces of the matcher are able to be pushed down,
			// so there is no change in the operator tree.
			return filter, nil
		}

		pipeline = append(ms.pipeline, bson.D{{"$match", matchBody}})
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
		return &FilterStage{
			source:      ms,
			hasSubquery: filter.hasSubquery,
			matcher:     localMatcher,
		}, nil
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
	keys, err := translateGroupByKeys(gb.keyExprs, ms.mappingRegistry.lookupFieldName)
	if err != nil {
		return gb, nil
	}

	// 2. Translate aggregations
	result, err := translateGroupByAggregates(gb.keyExprs, gb.selectExprs, ms.mappingRegistry.lookupFieldName)
	if err != nil {
		return gb, nil
	}

	result.group[groupID] = keys

	// 3. Translate the final project
	project, err := translateGroupByProject(result.exprs, result.mappingRegistry.lookupFieldName)
	if err != nil {
		return gb, nil
	}

	// 4. append to the pipeline
	pipeline = append(pipeline, bson.D{{"$group", result.group}})
	pipeline = append(pipeline, bson.D{{"$project", project}})

	// 5. Fix up the TableScan operator - None of the current registrations in mappingRegistry are valid any longer.
	// We need to clear them out and re-register the new columns.
	mappingRegistry := &mappingRegistry{}
	for _, sExpr := range gb.selectExprs {
		// at this point, our project has all the stringified names of the select expressions, so we need to re-map them
		// each column to its new MongoDB name. This process is what makes the push-down transparent to subsequent operators
		// in the tree that either haven't yet been pushed down, or cannot be. Either way, we output of a push-down must be
		// exactly the same as the output of a non-pushed-down group.
		mappingRegistry.addColumn(sExpr.Column)
		mappingRegistry.registerMapping(sExpr.Column.Table, sExpr.Column.Name, dottifyFieldName(sExpr.Expr.String()))
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
func translateGroupByKeys(keyExprs SelectExpressions, lookupFieldName fieldNameLookup) (bson.M, error) {

	keys := bson.M{}

	for _, keyExpr := range keyExprs {
		translatedKey, ok := TranslateExpr(keyExpr.Expr, lookupFieldName)
		if !ok {
			return nil, fmt.Errorf("could not translate '%v'", keyExpr.Expr.String())
		}

		keys[dottifyFieldName(keyExpr.Expr.String())] = translatedKey
	}

	return keys, nil
}

// translateGroupByAggregatesResult is just a holder for the results from translateGroupByAggregates.
type translateGroupByAggregatesResult struct {
	group           bson.M
	exprs           []SQLExpr
	mappingRegistry *mappingRegistry
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
func translateGroupByAggregates(keyExprs, selectExprs SelectExpressions, lookupFieldName fieldNameLookup) (*translateGroupByAggregatesResult, error) {

	// For example, in "select a + sum(b) from bar group by a", we should not create
	// an aggregate for a because it's part of the key.
	isGroupKey := func(expr SQLExpr) bool {
		exprString := expr.String()
		for _, sExpr := range keyExprs {
			if exprString == sExpr.Expr.String() {
				return true
			}
		}

		return false
	}

	// represents all the expressions that should be passed on to $project such that translateGroupByProject
	// is able to do its work without redoing a bunch of the conditionals and special casing here.
	newExprs := []SQLExpr{}

	// translator will "accumulate" all the group fields. Below, we iterate over each select expressions, which
	// account for all the fields that need to be present in the $group.
	translator := &groupByAggregateTranslator{bson.M{}, isGroupKey, lookupFieldName, &mappingRegistry{}}

	for _, selectExpr := range selectExprs {

		newExpr, err := translator.Visit(selectExpr.Expr)
		if err != nil {
			return nil, err
		}

		newExprs = append(newExprs, newExpr)
	}

	return &translateGroupByAggregatesResult{translator.group, newExprs, translator.mappingRegistry}, nil
}

type groupByAggregateTranslator struct {
	group           bson.M
	isGroupKey      func(SQLExpr) bool
	lookupFieldName fieldNameLookup
	mappingRegistry *mappingRegistry
}

// Visit recursively visits each expression in the tree, adds the relavent $group entries, and returns
// an expression that can be used to generate a subsequent $project.
func (v *groupByAggregateTranslator) Visit(e SQLExpr) (SQLExpr, error) {

	switch typedE := e.(type) {
	case SQLColumnExpr:
		fieldName, ok := v.lookupFieldName(typedE.tableName, typedE.columnName)
		if !ok {
			return nil, fmt.Errorf("could not map %v.%v to a field", typedE.tableName, typedE.columnName)
		}
		if !v.isGroupKey(typedE) {
			// since it's not an aggregation function, this implies that it takes the first value of the column.
			// So project the field, and register the mapping.
			v.group[dottifyFieldName(typedE.String())] = bson.M{"$first": getProjectedFieldName(fieldName)}
			v.mappingRegistry.registerMapping(typedE.tableName, typedE.columnName, dottifyFieldName(typedE.String()))
		} else {
			// the _id is added to the $group in translateGroupByKeys. We will only be here if the user has also projected
			// the group key, in which we'll need this to look it up in translateGroupByProject under its name. Hence, all
			// we need to do is register the mapping.
			v.mappingRegistry.registerMapping(typedE.tableName, typedE.columnName, groupID+"."+dottifyFieldName(typedE.String()))
		}
		return typedE, nil
	case *SQLAggFunctionExpr:
		var newExpr SQLExpr
		if typedE.Distinct {
			// Distinct aggregation expressions are two-step aggregations. In the $group stage, we use $addToSet
			// to handle whatever the distinct expression is, which could be a simply field name, or something
			// more complex like a mathematical computation. We don't care either way, and TranslateExpr handles
			// generating the correct thing. Once this is done, we create a new SQLAggFunctionExpr whose argument
			// maps to the newly named field containing the set of values to perform the aggregation on.
			trans, ok := TranslateExpr(typedE.Exprs[0], v.lookupFieldName)
			if !ok {
				return nil, fmt.Errorf("could not translate '%v'", typedE.String())
			}
			fieldName := groupDistinctPrefix + dottifyFieldName(typedE.Exprs[0].String())
			columnType := schema.ColumnType{typedE.Type(), schema.MongoNone}
			newExpr = &SQLAggFunctionExpr{
				Name:  typedE.Name,
				Exprs: []SQLExpr{SQLColumnExpr{groupTempTable, fieldName, columnType}},
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
			if typedE.Name == "count" && typedE.Exprs[0] == SQLVarchar("*") {
				trans = bson.M{"$sum": 1}
			} else if typedE.Name == "count" {
				trans, ok = TranslateExpr(typedE.Exprs[0], v.lookupFieldName)
				if !ok {
					return nil, fmt.Errorf("could not translate '%v'", typedE.Exprs[0].String())
				}

				trans = getCountAggregation(trans)
			} else {
				trans, ok = TranslateExpr(typedE, v.lookupFieldName)
				if !ok {
					return nil, fmt.Errorf("could not translate '%v'", typedE.String())
				}
			}
			fieldName := dottifyFieldName(typedE.String())
			columnType := schema.ColumnType{typedE.Type(), schema.MongoNone}
			newExpr = SQLColumnExpr{groupTempTable, fieldName, columnType}
			v.group[fieldName] = trans
			v.mappingRegistry.registerMapping(groupTempTable, fieldName, fieldName)
		}

		return newExpr, nil

	default:
		if v.isGroupKey(e) {
			// the _id is added to the $group in translateGroupByKeys. We will only be here if the user has also projected
			// the group key, in which we'll need this to look it up in translateGroupByProject under its name. In this,
			// we need to create a new expr that is simply a field pointing at the nested identifier and register that
			// mapping.
			fieldName := dottifyFieldName(e.String())
			columnType := schema.ColumnType{typedE.Type(), schema.MongoNone}
			newExpr := SQLColumnExpr{groupTempTable, fieldName, columnType}
			v.mappingRegistry.registerMapping(groupTempTable, fieldName, groupID+"."+fieldName)
			return newExpr, nil
		}

		// We'll only get here for two-step expressions. An example is
		// 'select a + b from foo group by a' or 'select b + sum(c) from foo group by a'. In this case,
		// we'll descend into the tree recursively which will build up the $group for the necessary pieces.
		// Finally, return the now changed expression such that $project can act on them appropriately.
		return walk(v, e)
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
func translateGroupByProject(exprs []SQLExpr, lookupFieldName fieldNameLookup) (bson.M, error) {
	project := bson.M{groupID: 0}
	for _, expr := range exprs {

		switch typedE := expr.(type) {
		case SQLColumnExpr:
			// Any one-step aggregations will end up here as they were fully performed in the $group. So, simple
			// column references ('select a') and simple aggregations ('select sum(a)').
			fieldName, ok := lookupFieldName(typedE.tableName, typedE.columnName)
			if !ok {
				return nil, fmt.Errorf("unable to get a field name for %v.%v", typedE.tableName, typedE.columnName)
			}

			project[dottifyFieldName(typedE.String())] = "$" + fieldName
		default:
			// Any two-step aggregations will end up here to complete the second step.
			trans, ok := TranslateExpr(expr, lookupFieldName)
			if !ok {
				return nil, fmt.Errorf("unable to translate '%v'", expr.String())
			}
			project[dottifyFieldName(typedE.String())] = trans
		}
	}

	return project, nil
}

const (
	joinedFieldNamePrefix = "__joined_"
)

func (v *pushDownOptimizer) visitJoin(join *JoinStage) (PlanStage, error) {

	if join.matcher == nil {
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
	default:
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

	if len(msForeign.pipeline) > 0 {
		return join, nil
	}

	// 3. find the local column and the foreign column
	lookupInfo, err := getLocalAndForeignColumns(msLocal.aliasName, msForeign.aliasName, join.matcher)
	if err != nil {
		return join, nil
	}

	// 4. get lookup fields
	localFieldName, ok := msLocal.mappingRegistry.lookupFieldName(lookupInfo.localColumn.tableName, lookupInfo.localColumn.columnName)
	if !ok {
		return join, nil
	}
	foreignFieldName, ok := msForeign.mappingRegistry.lookupFieldName(lookupInfo.foreignColumn.tableName, lookupInfo.foreignColumn.columnName)
	if !ok {
		return join, nil
	}

	asField := dottifyFieldName(joinedFieldNamePrefix + msForeign.collectionName)

	// 5. compute all the mappings from the msForeign mapping registry to be nested under
	// the 'asField' we used above.
	newMappingRegistry := msLocal.mappingRegistry.copy()

	newMappingRegistry.columns = append(newMappingRegistry.columns, msForeign.mappingRegistry.columns...)
	if msForeign.mappingRegistry.fields != nil {
		for tableName, columns := range msForeign.mappingRegistry.fields {
			for columnName, fieldName := range columns {
				newMappingRegistry.registerMapping(tableName, columnName, asField+"."+fieldName)
			}
		}
	}

	// 6. build the pipeline
	pipeline := msLocal.pipeline

	lookup := bson.M{
		"from":         msForeign.collectionName,
		"localField":   localFieldName,
		"foreignField": foreignFieldName,
		"as":           asField,
	}

	pipeline = append(pipeline, bson.D{{"$lookup", lookup}})

	unwind := bson.M{
		"path": "$" + asField,
		"preserveNullAndEmptyArrays": joinKind == LeftJoin,
	}

	pipeline = append(pipeline, bson.D{{"$unwind", unwind}})

	if lookupInfo.remainingPredicate != nil && joinKind == LeftJoin {

		ifPart, ok := TranslateExpr(lookupInfo.remainingPredicate, newMappingRegistry.lookupFieldName)
		if !ok {
			return join, nil
		}

		// if the predicate doesn't pass, then we set the 'as' field to nil
		project := bson.M{
			asField: bson.M{
				"$cond": bson.M{
					"if":   ifPart,
					"then": "$" + asField,
					"else": nil,
				},
			},
		}

		// we now need to make sure we project all the existing columns from the local mongo source
		for _, c := range msLocal.mappingRegistry.columns {
			field, ok := msLocal.mappingRegistry.lookupFieldName(c.Table, c.Name)
			if !ok {
				return join, nil
			}
			project[field] = 1
		}

		pipeline = append(pipeline, bson.D{{"$project", project}})
	}

	// 7. build the new operators
	ms := msLocal.clone()
	ms.pipeline = pipeline
	ms.mappingRegistry = newMappingRegistry

	if lookupInfo.remainingPredicate != nil && joinKind == InnerJoin {
		return v.Visit(&FilterStage{
			source:  ms,
			matcher: lookupInfo.remainingPredicate,
		})
	}

	return ms, nil
}

type lookupInfo struct {
	localColumn        *SQLColumnExpr
	foreignColumn      *SQLColumnExpr
	remainingPredicate SQLExpr
}

func getLocalAndForeignColumns(localTableName, foreignTableName string, e SQLExpr) (*lookupInfo, error) {
	exprs := splitExpression(e)

	// find a SQLEqualsExpr where the left and right are columns from the local and foreign tables
	for i, expr := range exprs {
		if equalExpr, ok := expr.(*SQLEqualsExpr); ok {
			// we must have a field from the left table and a field from the right table
			if column1, ok := equalExpr.left.(SQLColumnExpr); ok {
				if column2, ok := equalExpr.right.(SQLColumnExpr); ok {
					var localColumn, foreignColumn *SQLColumnExpr
					if column1.tableName == localTableName {
						localColumn = &column1
					} else if column1.tableName == foreignTableName {
						foreignColumn = &column1
					}

					if column2.tableName == localTableName {
						localColumn = &column2
					} else if column2.tableName == foreignTableName {
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

func (v *pushDownOptimizer) visitLimit(limit *LimitStage) (PlanStage, error) {

	ms, ok := v.canPushDown(limit.source)
	if !ok {
		return limit, nil
	}

	pipeline := ms.pipeline

	if limit.offset > 0 {
		pipeline = append(pipeline, bson.D{{"$skip", limit.offset}})
	}

	if limit.limit > 0 {
		pipeline = append(pipeline, bson.D{{"$limit", limit.limit}})
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

	for _, key := range orderBy.keys {

		var tableName, columnName string

		switch typedE := key.expr.Expr.(type) {
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
			return orderBy, nil
		}
		direction := 1
		if !key.ascending {
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

	fixedExpressions := SelectExpressions{}

	tables := v.planCtx.Schema.Databases[v.planCtx.Db].Tables

	// Track whether or not we've successfully mapped every field into the $project of the source.
	// If so, this Project node can be removed from the query plan tree.
	canReplaceProject := true

	for _, exp := range project.sExprs {

		// Convert the column's SQL expression into an expression in mongo query language.
		projectedField, ok := TranslateExpr(exp.Expr, ms.mappingRegistry.lookupFieldName)
		if !ok {
			// Expression can't be translated, so it can't be projected.
			// We skip it and leave this Project node in the query plan so that it still gets
			// evaluated during execution.
			canReplaceProject = false
			fixedExpressions = append(fixedExpressions, exp)

			// There might still be fields referenced in this expression
			// that we still need to project, so collect them and add them to the projection.
			refdCols, err := referencedColumns(exp.Expr, tables)
			if err != nil {
				return nil, err
			}
			for _, refdCol := range refdCols {
				fieldName, ok := ms.mappingRegistry.lookupFieldName(refdCol.Table, refdCol.Name)
				if !ok {
					// TODO log that optimization gave up here.
					return project, nil
				}

				fieldsToProject[dottifyFieldName(fieldName)] = getProjectedFieldName(fieldName)
			}
			continue
		}

		safeFieldName := dottifyFieldName(exp.Expr.String())

		fieldsToProject[safeFieldName] = projectedField

		fixedMappingRegistry.addColumn(exp.Column)
		fixedMappingRegistry.registerMapping(exp.Column.Table, exp.Column.Name, safeFieldName)
		columnType := schema.ColumnType{exp.Column.SQLType, exp.Column.MongoType}
		columnExpr := SQLColumnExpr{exp.Column.Table, exp.Column.Name, columnType}
		fixedExpressions = append(fixedExpressions,
			SelectExpression{
				Column: exp.Column,
				Expr:   columnExpr,
			},
		)

	}

	pipeline := ms.pipeline
	pipeline = append(pipeline, bson.D{{"$project", fieldsToProject}})
	ms = ms.clone()
	ms.pipeline = pipeline
	ms.mappingRegistry = &fixedMappingRegistry

	if canReplaceProject {
		return ms, nil
	}

	project = project.clone()
	project.source = ms
	return project, nil
}
