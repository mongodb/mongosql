package evaluator

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
)

func OptimizeOperator(ctx *ExecutionCtx, o Operator) (Operator, error) {
	v := &optimizer{ctx: ctx}
	return v.Visit(o)
}

type optimizer struct {
	ctx    *ExecutionCtx
	fields map[string]map[string]string
}

func (v *optimizer) registerFieldName(tbl, column, field string) {

	if v.fields == nil {
		v.fields = make(map[string]map[string]string)
	}

	if _, ok := v.fields[tbl]; !ok {
		v.fields[tbl] = make(map[string]string)
	}

	v.fields[tbl][column] = field
}

func (v *optimizer) getFieldName(tbl, column string) string {
	if v.fields == nil {
		return column
	}

	columnToField, ok := v.fields[tbl]
	if !ok {
		// no mapping exists, so we use the column name
		return column
	}

	field, ok := columnToField[column]
	if !ok {
		// no mapping exists, so we use the column name
		return column
	}

	return field
}

func (v *optimizer) Visit(opr Operator) (Operator, error) {

	o, err := walkOperatorTree(v, opr)
	if err != nil {
		return nil, err
	}

	switch typedO := o.(type) {

	case *Filter:
		return v.visitFilter(typedO)
	case *GroupBy:
		return v.visitGroupBy(typedO)
	case *Limit:
		return v.visitLimit(typedO)
	case *TableScan:
		return v.visitTableScan(typedO)
	}

	return o, err
}

func (v *optimizer) visitFilter(filter *Filter) (Operator, error) {

	sa, ts, ok := canPushDown(filter.source)
	if !ok {
		return filter, nil
	}

	optimizedExpr, err := OptimizeSQLExpr(filter.matcher)
	if err != nil {
		return nil, err
	}

	pipeline := ts.pipeline
	var localMatcher SQLExpr

	if value, ok := optimizedExpr.(SQLValue); ok {
		// our optimized expression has left us with just a value,
		// we can see if it matches right now. If so, we eliminate
		// the filter from the tree. Otherwise, we return an
		// operator that yields no rows.

		matches, err := Matches(value, nil)
		if err != nil {
			return nil, err
		}
		if !matches {
			return &Empty{}, nil
		}

		// otherwise, the filter simply gets removed from the tree

	} else {
		var matchBody bson.M
		matchBody, localMatcher = TranslatePredicate(optimizedExpr, v.ctx.Schema.Databases[v.ctx.Db])

		if matchBody == nil {
			// no pieces of the matcher are able to be pushed down,
			// so there is no change in the operator tree.
			return filter, nil
		}

		pipeline = append(ts.pipeline, bson.D{{"$match", matchBody}})
	}

	// if we end up here, it's because we have messed with the pipeline
	// in the current table scan operator, so we need to reconstruct the
	// operator nodes.
	ts = &TableScan{
		pipeline:   pipeline,
		dbName:     ts.dbName,
		tableName:  ts.tableName,
		matcher:    ts.matcher,
		aggregated: ts.aggregated,
	}

	sa = &SourceAppend{
		source:      ts,
		hasSubquery: sa.hasSubquery,
	}

	if localMatcher != nil {
		// we ended up here because we have a predicate
		// that can be partially pushed down, so we construct
		// a new filter with only the part remaining that
		// cannot be pushed down.
		filter = &Filter{
			source:      sa,
			matcher:     localMatcher,
			hasSubquery: filter.hasSubquery,
		}

		return filter, nil
	}

	// everything was able to be pushed down, so the filter
	// is removed from the plan.
	return sa, nil
}

func (v *optimizer) visitGroupBy(gb *GroupBy) (Operator, error) {

	sa, ts, ok := canPushDown(gb.source)
	if !ok {
		return gb, nil
	}

	pipeline := ts.pipeline

	db := v.ctx.Schema.Databases[v.ctx.Db]

	table := db.Tables[ts.tableName]

	groupClause, distinctAggFuncs, err := TranslateGroupBy(gb, db, table)
	if err != nil {
		// we couldn't push down the GROUP BY clause
		if err == ErrPushDown {
			return gb, nil
		}
		return nil, err
	}

	// rewrite the grouped columns
	projectClause := projectGroupBy(groupClause, distinctAggFuncs)

	pipeline = append(pipeline, bson.D{{"$group", groupClause}})

	if gb.matcher != nil && gb.matcher != SQLTrue {

		havingClause, err := TranslateExpr(gb.matcher, db, true)
		if err != nil {
			if err == ErrPushDown {
				return gb, nil
			}
			return nil, err
		}

		if havingClause != nil {
			projectClause[HavingMatcher] = havingClause
		}
	}

	// update the table source for all columns
	for _, column := range gb.sExprs {
		column.Table = ts.tableName
	}

	pipeline = append(pipeline, bson.D{{"$project", projectClause}})

	if projectClause[HavingMatcher] != nil {
		havingClause := bson.M{
			HavingMatcher: bson.M{
				"$ne": false,
			},
		}

		pipeline = append(pipeline, bson.D{{"$match", havingClause}})
	}

	ts = &TableScan{
		pipeline:   pipeline,
		dbName:     ts.dbName,
		tableName:  ts.tableName,
		matcher:    ts.matcher,
		aggregated: true,
	}

	sa = &SourceAppend{
		source:      ts,
		hasSubquery: sa.hasSubquery,
	}

	return sa, nil
}

func (_ *optimizer) visitLimit(limit *Limit) (Operator, error) {

	sa, ts, ok := canPushDown(limit.source)
	if !ok {
		return limit, nil
	}

	pipeline := ts.pipeline

	if limit.offset > 0 {
		pipeline = append(pipeline, bson.D{{"$skip", limit.offset}})
	}

	if limit.rowcount > 0 {
		pipeline = append(pipeline, bson.D{{"$limit", limit.rowcount}})
	}

	ts = &TableScan{
		pipeline:   pipeline,
		dbName:     ts.dbName,
		tableName:  ts.tableName,
		matcher:    ts.matcher,
		aggregated: true,
	}

	sa = &SourceAppend{
		source:      ts,
		hasSubquery: sa.hasSubquery,
	}

	return sa, nil
}

func (v *optimizer) visitTableScan(ts *TableScan) (Operator, error) {

	db, ok := v.ctx.Schema.Databases[v.ctx.Db]
	if !ok {
		return nil, fmt.Errorf("The database %v is invalid.", v.ctx.Db)
	}

	table, ok := db.Tables[ts.tableName]
	if !ok {
		return nil, fmt.Errorf("The table %v.%v is invalid.", v.ctx.Db, ts.tableName)
	}

	// go through the tableSchema and register all the "columns" and their field names.
	// any fields that aren't in the schema are unknown to us and cannot be used.
	for _, column := range table.Columns {
		v.registerFieldName(table.Name, column.SqlName, column.Name)
	}

	return ts, nil
}

func canPushDown(op Operator) (*SourceAppend, *TableScan, bool) {

	// we can only optimize an operator whose source is a SourceAppend
	// with a source of a TableScan
	sa, ok := op.(*SourceAppend)
	if !ok {
		return nil, nil, false
	}
	ts, ok := sa.source.(*TableScan)
	if !ok {
		return nil, nil, false
	}

	return sa, ts, true
}
