package evaluator

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/mgo.v2/bson"
)

// orderedGroup holds all the rows belonging to a given key in the groups
// and an slice of the keys for each group.
type orderedGroup struct {
	groups map[string][]Row
	keys   []string
}

// GroupBy groups records according to one or more fields.
type GroupBy struct {
	// selectExprs holds the SelectExpression that should
	// be present in the result of a grouping. This will
	// include SelectExpressions for aggregates that might
	// not be projected, but are required for further
	// processing, such as when ordering by an aggregate.
	selectExprs SelectExpressions

	// source is the operator that provides the data to group
	source Operator

	// keySelectExprs holds the expression(s) to group by. For example, in
	// select a, count(b) from foo group by a,
	// keyExprs will hold the parsed column name 'a'.
	keyExprs SelectExpressions

	// grouped indicates if the source operator data has been grouped
	grouped bool

	// err holds any error encountered during processing
	err error

	// finalGrouping contains all grouped records and an ordered list of
	// the keys as read from the source operator
	finalGrouping orderedGroup

	// channel on which to send rows derived from the final grouping
	outChan chan AggRowCtx

	ctx *ExecutionCtx
}

func (gb *GroupBy) Open(ctx *ExecutionCtx) error {
	return gb.init(ctx)
}

func (gb *GroupBy) init(ctx *ExecutionCtx) error {
	gb.ctx = ctx
	return gb.source.Open(ctx)
}

func (gb *GroupBy) evaluateGroupByKey(row *Row) (string, error) {

	var gbKey string

	for _, expr := range gb.keyExprs {
		evalCtx := &EvalCtx{Rows: Rows{*row}}
		value, err := expr.Expr.Evaluate(evalCtx)
		if err != nil {
			return "", err
		}

		// TODO: might be better to use a hash for this
		gbKey += fmt.Sprintf("%#v", value)
	}

	return gbKey, nil
}

func (gb *GroupBy) createGroups() error {

	gb.finalGrouping = orderedGroup{
		groups: make(map[string][]Row, 0),
	}

	r := &Row{}

	// iterator source to create groupings
	for gb.source.Next(r) {

		key, err := gb.evaluateGroupByKey(r)
		if err != nil {
			return err
		}

		if gb.finalGrouping.groups[key] == nil {
			gb.finalGrouping.keys = append(gb.finalGrouping.keys, key)
		}

		gb.finalGrouping.groups[key] = append(gb.finalGrouping.groups[key], *r)

		r = &Row{}
	}

	gb.grouped = true

	return gb.source.Err()
}

func (gb *GroupBy) evalAggRow(r []Row) (*Row, error) {

	aggValues := map[string]Values{}

	row := &Row{}

	for _, sExpr := range gb.selectExprs {

		evalCtx := &EvalCtx{Rows: r}

		v, err := sExpr.Expr.Evaluate(evalCtx)
		if err != nil {
			return nil, err
		}

		value := Value{
			Name: sExpr.Name,
			View: sExpr.View,
			Data: v,
		}
		aggValues[sExpr.Table] = append(aggValues[sExpr.Table], value)
	}

	for table, values := range aggValues {
		row.Data = append(row.Data, TableRow{table, values})
	}

	return row, nil
}

func (gb *GroupBy) iterChan() chan AggRowCtx {
	ch := make(chan AggRowCtx)

	go func() {
		for _, key := range gb.finalGrouping.keys {
			v := gb.finalGrouping.groups[key]
			r, err := gb.evalAggRow(v)
			if err != nil {
				gb.err = err
				close(ch)
				return
			}

			// check we have some matching data
			if len(r.Data) != 0 {
				ch <- AggRowCtx{*r, v}
			}
		}
		close(ch)
	}()
	return ch
}

func (gb *GroupBy) Next(row *Row) bool {
	if !gb.grouped {
		if err := gb.createGroups(); err != nil {
			gb.err = err
			return false
		}
		gb.outChan = gb.iterChan()
	}

	rCtx, done := <-gb.outChan
	gb.ctx.GroupRows = rCtx.Ctx
	row.Data = rCtx.Row.Data

	return done
}

func (gb *GroupBy) Close() error {
	return gb.source.Close()
}

func (gb *GroupBy) Err() error {
	if err := gb.source.Err(); err != nil {
		return err
	}
	return gb.err
}

func (gb *GroupBy) OpFields() (columns []*Column) {
	for _, sExpr := range gb.selectExprs {
		column := &Column{
			Name:  sExpr.Name,
			View:  sExpr.View,
			Table: sExpr.Table,
		}
		columns = append(columns, column)
	}
	return columns
}

func (gb *GroupBy) String() string {

	b := bytes.NewBufferString("select exprs ( ")

	for _, expr := range gb.selectExprs {
		b.WriteString(fmt.Sprintf("'%v' ", expr.View))
	}

	b.WriteString(") grouped by ( ")

	for _, expr := range gb.keyExprs {
		b.WriteString(fmt.Sprintf("'%v' ", expr.View))
	}

	b.WriteString(")")

	return b.String()

}

///////////////
//Optimization
///////////////

const (
	groupID             = "_id"
	groupDistinctPrefix = "distinct "
	groupTempTable      = ""
)

// visitGroupBy works by using a visitor to systematically visit and replace certain SQLExpr while generating
// $group and $project stages for the aggregation pipeline.
func (v *optimizer) visitGroupBy(gb *GroupBy) (Operator, error) {

	ms, ok := canPushDown(gb.source)
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

	ms = ms.WithPipeline(pipeline).WithMappingRegistry(mappingRegistry)

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

		switch typedE := keyExpr.Expr.(type) {

		case SQLColumnExpr:

			fieldName, ok := lookupFieldName(typedE.tableName, typedE.columnName)
			if !ok {
				return nil, fmt.Errorf("could not map '%v.%v' to a field", typedE.tableName, typedE.columnName)
			}

			// project to a well-known name of the expr.String(). So, in 'select a from foo group by a',
			// 'a' will get projected as 'foo_DOT_a'
			keys[dottifyFieldName(typedE.String())] = getProjectedFieldName(fieldName)

		default:

			transExpr, ok := TranslateExpr(typedE, lookupFieldName)
			if !ok {
				return nil, fmt.Errorf("could not translate '%v'", typedE.String())
			}

			// project to a well-known name of the expr.String(). So, in 'select a from foo group by a+b',
			// 'a+b' will get projected as 'foo_DOT_a+foo_DOT_b'
			keys[dottifyFieldName(typedE.String())] = transExpr
		}
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
			newExpr = &SQLAggFunctionExpr{
				Name:  typedE.Name,
				Exprs: []SQLExpr{SQLColumnExpr{groupTempTable, fieldName}},
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
			if typedE.Name == "count" && typedE.Exprs[0] == SQLString("*") {
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
			newExpr = SQLColumnExpr{groupTempTable, fieldName}
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
			newExpr := SQLColumnExpr{groupTempTable, fieldName}
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

// getProjectedFieldName returns an interface to project the given field.
func getProjectedFieldName(fieldName string) interface{} {

	names := strings.Split(fieldName, ".")

	if len(names) == 1 {
		return "$" + fieldName
	}

	// TODO: this is to enable our current tests pass - since
	// we allow array mappings like:
	//
	// - name: loc.0
	//   sqlname: latitude
	//
	// In the future, this swath of code should be removed
	// and this function can go away
	value, err := strconv.Atoi(names[len(names)-1])
	if err == nil {
		fieldName = fieldName[0:strings.LastIndex(fieldName, ".")]
		return bson.M{"$arrayElemAt": []interface{}{"$" + fieldName, value}}
	}

	return "$" + fieldName
}
