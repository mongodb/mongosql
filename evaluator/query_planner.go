package evaluator

import (
	"strings"

	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
)

type queryPlanBuilder struct {
	// algebrizer references the algebrizer with which it is communicating during building.
	algebrizer *algebrizer
	// exprCollector collects all the expressions and aggregation expressions used
	// during the query. It allows use to know what columns will be required for a certain
	// stage and also aids in the movement of some expressions from later to earlier in the
	// pipeline. For instance, aggregate expressions need to be moved and replaced in project
	// while building a GroupByStage.
	exprCollector *expressionCollector

	// hasCorrelatedSubquery indicates if a correlated sub query exists in the plan.
	hasCorrelatedSubquery bool

	from     PlanStage
	where    SQLExpr
	groupBy  []SQLExpr
	having   SQLExpr
	distinct bool
	orderBy  []*orderByTerm
	project  ProjectedColumns
	hasLimit bool
	offset   int64
	rowcount int64
}

func (b *queryPlanBuilder) build() PlanStage {

	plan := b.from

	if b.hasCorrelatedSubquery {
		plan = NewSourceAppendStage(plan)
	}

	plan = b.buildWhere(plan)
	plan = b.buildGroupBy(plan)
	plan = b.buildHaving(plan)
	plan = b.buildDistinct(plan)
	plan = b.buildOrderBy(plan)
	plan = b.buildLimit(plan)
	plan = b.buildProject(plan)

	if b.hasCorrelatedSubquery {
		plan = NewSourceRemoveStage(plan)
	}

	if b.hasLimit && b.rowcount == 0 {
		var columns []*Column
		for _, projectedColumn := range b.project {
			column := &Column{
				Name:      projectedColumn.Name,
				Table:     projectedColumn.Table,
				SQLType:   projectedColumn.SQLType,
				MongoType: projectedColumn.MongoType,
			}
			columns = append(columns, column)
		}
		return NewEmptyStage(columns)
	}

	return plan
}

func (b *queryPlanBuilder) buildDistinct(source PlanStage) PlanStage {
	plan := source
	if b.distinct {
		var keys []SQLExpr
		var projectedKeys ProjectedColumns
		for _, c := range b.project {
			projectedKeys = append(projectedKeys, *projectedColumnFromExpr(c.Expr))
			keys = append(keys, c.Expr)

			// don't want these interfering with b.exprCollector.allNonAggReferencedColumns
			b.exprCollector.Remove(c.Expr)
		}

		// projectedColumns will include any column that is not an aggregate function.
		// as well as all the keys.
		projectedColumns := projectedKeys
		for _, e := range b.exprCollector.allNonAggReferencedColumns.copyExprs() {
			pc := projectedColumnFromExpr(e)
			projectedColumns = append(projectedColumns, *pc)
		}

		plan = NewGroupByStage(plan, keys, projectedColumns.Unique())

		// now we must replace all the project values with columns as
		// any that weren't already a column have now been computed.
		projectedColumns = ProjectedColumns{}
		for i, pc := range b.project {
			newExpr := SQLColumnExpr{
				tableName:  projectedKeys[i].Table,
				columnName: projectedKeys[i].Name,
				columnType: schema.ColumnType{
					SQLType:   projectedKeys[i].SQLType,
					MongoType: projectedKeys[i].MongoType,
				},
			}
			projectedColumns = append(projectedColumns, ProjectedColumn{
				Column: pc.Column,
				Expr:   newExpr,
			})
			b.exprCollector.Add(newExpr)
		}

		b.project = projectedColumns
	}

	return plan
}

func (b *queryPlanBuilder) buildGroupBy(source PlanStage) PlanStage {
	plan := source
	if len(b.groupBy) > 0 || len(b.exprCollector.allAggFunctions.exprs) > 0 {
		// do this now so it doesn't throw off the b.exprCollector.allNonAggReferencedColumns.
		b.exprCollector.RemoveAll(b.groupBy)

		// projectedAggregates will include all the aggregates as well
		// as any column that is not an aggregate function.
		var projectedAggregates ProjectedColumns
		for _, e := range b.exprCollector.allNonAggReferencedColumns.copyExprs() {
			pc := projectedColumnFromExpr(e)
			projectedAggregates = append(projectedAggregates, *pc)
		}

		for _, e := range b.exprCollector.allAggFunctions.copyExprs() {
			pc := projectedColumnFromExpr(e)
			projectedAggregates = append(projectedAggregates, *pc)
		}

		plan = NewGroupByStage(plan, b.groupBy, projectedAggregates.Unique())

		// replace aggregation expressions with columns coming out of the GroupByStage
		// because they have already been aggregated and are now just columns.
		b.replaceAggFunctions()
	}

	return plan
}

func (b *queryPlanBuilder) buildHaving(source PlanStage) PlanStage {
	if b.having != nil {
		b.exprCollector.Remove(b.having)
		return NewFilterStage(source, b.having)
	}

	return source
}

func (b *queryPlanBuilder) buildLimit(source PlanStage) PlanStage {
	if b.hasLimit {
		return NewLimitStage(source, b.offset, b.rowcount)
	}
	return source
}

func (b *queryPlanBuilder) buildOrderBy(source PlanStage) PlanStage {
	if len(b.orderBy) > 0 {
		for _, obt := range b.orderBy {
			b.exprCollector.Remove(obt.expr)
		}

		return NewOrderByStage(source, b.orderBy...)
	}

	return source
}

func (b *queryPlanBuilder) buildProject(source PlanStage) PlanStage {
	if len(b.project) > 0 {
		for _, pc := range b.project {
			b.exprCollector.Remove(pc.Expr)
		}
		return NewProjectStage(source, b.project...)
	}

	return source
}

func (b *queryPlanBuilder) buildWhere(source PlanStage) PlanStage {
	if b.where != nil {
		b.exprCollector.Remove(b.where)
		return NewFilterStage(source, b.where)
	}

	return source
}

func (b *queryPlanBuilder) replaceAggFunctions() error {

	// since we are replacing aggregates (which likely include columns) with other columns,
	// we need to update the exprCollection with the new information so that it continues
	// to be correct. Therefore, we'll be removing the old expressions and adding in
	// new ones.

	if len(b.project) > 0 {

		var projectedColumns ProjectedColumns
		for _, pc := range b.project {
			b.exprCollector.Remove(pc.Expr)
			replaced, err := replaceAggFunctionsWithColumns(pc.Expr)
			if err != nil {
				return err
			}
			b.exprCollector.Add(replaced)

			projectedColumns = append(projectedColumns, ProjectedColumn{
				Expr:   replaced,
				Column: pc.Column,
			})
		}
		b.project = projectedColumns
	}

	if b.having != nil {
		b.exprCollector.Remove(b.having)
		having, err := replaceAggFunctionsWithColumns(b.having)
		if err != nil {
			return err
		}
		b.exprCollector.Add(having)

		b.having = having
	}

	if len(b.orderBy) > 0 {
		var orderBy []*orderByTerm
		for _, obt := range b.orderBy {
			b.exprCollector.Remove(obt.expr)
			replaced, err := replaceAggFunctionsWithColumns(obt.expr)
			if err != nil {
				return err
			}
			b.exprCollector.Add(replaced)

			orderBy = append(orderBy, &orderByTerm{
				ascending: obt.ascending,
				expr:      replaced,
			})
		}

		b.orderBy = orderBy
	}

	return nil
}

func (b *queryPlanBuilder) includeGroupBy(groupBy parser.GroupBy) error {
	keys, err := b.algebrizer.translateGroupBy(groupBy)
	if err != nil {
		return err
	}

	b.exprCollector.AddAll(keys)
	b.groupBy = keys
	return nil
}

func (b *queryPlanBuilder) includeHaving(having *parser.Where) error {
	pred, err := b.algebrizer.translateExpr(having.Expr)
	if err != nil {
		return err
	}

	b.exprCollector.Add(pred)
	b.having = pred
	return nil
}

func (b *queryPlanBuilder) includeLimit(limit *parser.Limit) error {
	offset, rowcount, err := b.algebrizer.translateLimit(limit)
	if err != nil {
		return err
	}
	b.hasLimit = true
	b.offset = int64(offset)
	b.rowcount = int64(rowcount)
	return nil
}

func (b *queryPlanBuilder) includeOrderBy(orderBy parser.OrderBy) error {
	terms, err := b.algebrizer.translateOrderBy(orderBy)
	if err != nil {
		return err
	}

	for _, obt := range terms {
		b.exprCollector.Add(obt.expr)
	}
	b.orderBy = terms
	return nil
}

func (b *queryPlanBuilder) includeSelect(selectExprs parser.SelectExprs) error {
	project, err := b.algebrizer.translateSelectExprs(selectExprs)
	if err != nil {
		return err
	}

	for _, pc := range project {
		b.exprCollector.Add(pc.Expr)
	}
	b.project = project
	return nil
}

func (b *queryPlanBuilder) includeWhere(where *parser.Where) error {
	pred, err := b.algebrizer.translateExpr(where.Expr)
	if err != nil {
		return err
	}

	b.exprCollector.Add(pred)
	b.where = pred
	return nil
}

type exprCountMap struct {
	counts map[string]int
	exprs  []SQLExpr
}

func newExprCountMap() *exprCountMap {
	return &exprCountMap{
		counts: make(map[string]int),
	}
}

func (m *exprCountMap) add(e SQLExpr) {
	s := e.String()
	if _, ok := m.counts[s]; ok {
		m.counts[s]++
	} else {
		m.counts[s] = 1
		m.exprs = append(m.exprs, e)
	}
}

func (m *exprCountMap) remove(e SQLExpr) {
	s := e.String()
	for i, expr := range m.exprs {
		if strings.EqualFold(s, expr.String()) {
			m.counts[s]--
			if m.counts[s] == 0 {
				m.exprs = append(m.exprs[:i], m.exprs[i+1:]...)
			}
			return
		}
	}
}

func (m *exprCountMap) copyExprs() []SQLExpr {
	exprs := make([]SQLExpr, len(m.exprs))
	copy(exprs, m.exprs)
	return exprs
}

type expressionCollector struct {
	allReferencedColumns       *exprCountMap
	allNonAggReferencedColumns *exprCountMap
	allAggFunctions            *exprCountMap

	inAggFunc  bool
	removeMode bool
}

func newExpressionCollector() *expressionCollector {
	return &expressionCollector{
		allReferencedColumns:       newExprCountMap(),
		allNonAggReferencedColumns: newExprCountMap(),
		allAggFunctions:            newExprCountMap(),
	}
}

func (c *expressionCollector) Remove(e SQLExpr) {
	c.removeMode = true
	c.Add(e)
	c.removeMode = false
}

func (c *expressionCollector) RemoveAll(e []SQLExpr) {
	c.removeMode = true
	c.AddAll(e)
	c.removeMode = false
}

func (c *expressionCollector) AddAll(exprs []SQLExpr) {
	for _, e := range exprs {
		c.Add(e)
	}
}

func (c *expressionCollector) Add(e SQLExpr) {
	c.visit(e)
}

func (v *expressionCollector) visit(n node) (node, error) {
	switch typedN := n.(type) {
	case *SQLAggFunctionExpr:
		v.inAggFunc = true
		if v.removeMode {
			v.allAggFunctions.remove(typedN)
		} else {
			v.allAggFunctions.add(typedN)
		}
		for _, a := range typedN.Exprs {
			v.visit(a)
		}
		v.inAggFunc = false
		return typedN, nil
	case SQLColumnExpr:
		if v.removeMode {
			v.allReferencedColumns.remove(typedN)
		} else {
			v.allReferencedColumns.add(typedN)
		}
		if !v.inAggFunc {
			if v.removeMode {
				v.allNonAggReferencedColumns.remove(typedN)
			} else {
				v.allNonAggReferencedColumns.add(typedN)
			}
		}
		return typedN, nil
	case *SQLSubqueryExpr:
		// TODO: need to add logic to only collect
		// columns and agg functions that apply
		return n, nil
	default:
		return walk(v, n)
	}
}

type aggFunctionFinder struct {
	aggFuncs []*SQLAggFunctionExpr
}

// getAggFunctions will take an expression and return all
// aggregation functions it finds within the expression.
func getAggFunctions(e SQLExpr) ([]*SQLAggFunctionExpr, error) {
	af := &aggFunctionFinder{}
	_, err := af.visit(e)
	if err != nil {
		return nil, err
	}

	return af.aggFuncs, nil
}

func (af *aggFunctionFinder) visit(n node) (node, error) {
	switch typedN := n.(type) {
	case *SQLExistsExpr, SQLColumnExpr, SQLNullValue, SQLNumeric, SQLVarchar, *SQLVariableExpr:
		return n, nil
	case *SQLAggFunctionExpr:
		af.aggFuncs = append(af.aggFuncs, typedN)
	case *SQLSubqueryExpr:
		// TODO: need to add logic to only collect
		// agg functions that apply
		return n, nil
	default:
		return walk(af, n)
	}

	return n, nil
}

type aggFunctionExprReplacer struct {
}

func replaceAggFunctionsWithColumns(e SQLExpr) (SQLExpr, error) {
	v := &aggFunctionExprReplacer{}
	n, err := v.visit(e)
	if err != nil {
		return nil, err
	}

	return n.(SQLExpr), nil
}

func (v *aggFunctionExprReplacer) visit(n node) (node, error) {
	switch typedN := n.(type) {
	case *SQLAggFunctionExpr:
		columnType := schema.ColumnType{
			SQLType:   typedN.Type(),
			MongoType: schema.MongoNone,
		}
		return SQLColumnExpr{"", typedN.String(), columnType}, nil
	case *SQLSubqueryExpr:
		// TODO: handle parental aggregates in correlated subquery
		return n, nil
	default:
		return walk(v, n)
	}
}

func projectedColumnFromExpr(expr SQLExpr) *ProjectedColumn {
	pc := &ProjectedColumn{
		Column: &Column{
			SQLType: expr.Type(),
		},
		Expr: expr,
	}

	if sqlCol, ok := expr.(SQLColumnExpr); ok {
		pc.Name = sqlCol.columnName
		pc.Table = sqlCol.tableName
		pc.MongoType = sqlCol.columnType.MongoType
	} else {
		pc.Name = expr.String()
	}

	return pc
}
