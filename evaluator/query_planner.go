package evaluator

import (
	"strings"

	"github.com/10gen/sqlproxy/collation"
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

	selectID   int
	aggregates []*SQLAggFunctionExpr

	from     PlanStage
	join     []SQLExpr
	where    SQLExpr
	groupBy  []SQLExpr
	having   SQLExpr
	distinct bool
	orderBy  []*orderByTerm
	project  ProjectedColumns
	hasLimit bool
	offset   uint64
	rowcount uint64
}

func (b *queryPlanBuilder) build() PlanStage {

	if b.hasLimit && b.rowcount == 0 {
		var columns []*Column
		for _, projectedColumn := range b.project {
			columns = append(columns, projectedColumn.Column)
		}
		return NewEmptyStage(columns, collation.Default)
	}

	plan := b.buildFrom(b.from)
	plan = b.buildWhere(plan)
	plan = b.buildGroupBy(plan)
	plan = b.buildHaving(plan)
	plan = b.buildDistinct(plan)
	plan = b.buildOrderBy(plan)
	plan = b.buildLimit(plan)
	plan = b.buildProject(plan)
	return plan
}

func (b *queryPlanBuilder) buildDistinct(source PlanStage) PlanStage {
	plan := source
	if b.distinct {
		var keys []SQLExpr
		var projectedKeys ProjectedColumns
		reqCols := b.exprCollector.referencedColumns.copyExprs()

		for _, c := range b.project {
			projectedKeys = append(projectedKeys, *b.projectedColumnFromExpr(c.Expr))
			keys = append(keys, c.Expr)

			// don't want these interfering with b.exprCollector.referencedColumns
			b.exprCollector.Remove(c.Expr)
		}

		// projectedColumns will include any column that is not an aggregate function.
		// as well as all the keys.
		projectedColumns := projectedKeys
		for _, e := range b.exprCollector.referencedColumns.copyExprs() {
			pc := b.projectedColumnFromExpr(e)
			projectedColumns = append(projectedColumns, *pc)
		}

		plan = NewGroupByStage(plan, keys, projectedColumns.Unique(), reqCols)

		// now we must replace all the project values with columns as
		// any that weren't already a column have now been computed.
		projectedColumns = ProjectedColumns{}
		for i, pc := range b.project {
			newExpr := NewSQLColumnExpr(
				b.selectID,
				projectedKeys[i].Table,
				projectedKeys[i].Name,
				projectedKeys[i].SQLType,
				projectedKeys[i].MongoType)

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
	if len(b.groupBy) > 0 || len(b.aggregates) > 0 {
		reqCols := b.exprCollector.referencedColumns.copyExprs()

		b.exprCollector.RemoveAll(b.groupBy)
		for _, a := range b.aggregates {
			b.exprCollector.Remove(a)
		}

		// projectedAggregates will include all the aggregates as well
		// as any column that was referenced.
		var projectedAggregates ProjectedColumns
		for _, e := range b.exprCollector.referencedColumns.copyExprs() {
			if col, ok := e.(SQLColumnExpr); ok {
				if !col.isAggregateReplacementColumn() {
					pc := b.projectedColumnFromExpr(e)
					projectedAggregates = append(projectedAggregates, *pc)
				}
			} else {
				pc := b.projectedColumnFromExpr(e)
				projectedAggregates = append(projectedAggregates, *pc)
			}
		}
		for _, e := range b.aggregates {
			pc := b.projectedColumnFromExpr(e)
			projectedAggregates = append(projectedAggregates, *pc)
		}

		plan = NewGroupByStage(plan, b.groupBy, projectedAggregates.Unique(), reqCols)
	}

	return plan
}

func (b *queryPlanBuilder) buildHaving(source PlanStage) PlanStage {
	if b.having != nil {
		reqCols := b.exprCollector.referencedColumns.copyExprs()

		b.exprCollector.Remove(b.having)
		return NewFilterStage(source, b.having, reqCols)
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
		reqCols := b.exprCollector.referencedColumns.copyExprs()

		for _, obt := range b.orderBy {
			b.exprCollector.Remove(obt.expr)
		}

		return NewOrderByStage(source, reqCols, b.orderBy...)
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
		var reqCols []SQLExpr
		for _, e := range b.exprCollector.referencedColumns.copyExprs() {
			if col, ok := e.(SQLColumnExpr); ok {
				if !col.isAggregateReplacementColumn() {
					reqCols = append(reqCols, col)
				}
			}
		}

		b.exprCollector.Remove(b.where)

		return NewFilterStage(source, b.where, reqCols)
	}
	return source
}

func (b *queryPlanBuilder) buildFrom(source PlanStage) PlanStage {
	switch typedS := source.(type) {
	case *JoinStage:
		if typedL, ok := typedS.left.(*JoinStage); ok {
			typedS.left = b.buildFrom(typedL)
		}
		if b.from != nil {
			reqCols := b.exprCollector.referencedColumns.copyExprs()

			if b.join != nil {
				for _, c := range b.join {
					if strings.Contains(typedS.matcher.String(), c.String()) {
						b.exprCollector.Remove(c)
					}
				}
			}
			return NewJoinStage(typedS.kind, typedS.left, typedS.right, typedS.matcher, reqCols)
		}
	}
	return source
}

func (b *queryPlanBuilder) includeFrom(p PlanStage) error {
	switch typedP := p.(type) {
	case *JoinStage:
		if typedL, ok := typedP.left.(*JoinStage); ok {
			err := b.includeFrom(typedL)
			if err != nil {
				return err
			}
		}
	}
	b.exprCollector.getJoinOnVals(p)
	b.join = b.exprCollector.referencedColumns.exprs
	return nil
}

func (b *queryPlanBuilder) includeAggregates(aggs []*SQLAggFunctionExpr) {
	b.aggregates = aggs
	for _, a := range b.aggregates {
		b.exprCollector.Add(a)
	}
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
	b.offset = uint64(offset)
	b.rowcount = uint64(rowcount)
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

func (b *queryPlanBuilder) projectedColumnFromExpr(expr SQLExpr) *ProjectedColumn {
	pc := &ProjectedColumn{
		Column: &Column{
			SelectID: b.selectID,
		},
		Expr: expr,
	}

	if sqlCol, ok := expr.(SQLColumnExpr); ok {
		pc.Name = sqlCol.columnName
		pc.Table = sqlCol.tableName
		pc.SQLType = sqlCol.columnType.SQLType
		pc.MongoType = sqlCol.columnType.MongoType
	} else {
		pc.Name = expr.String()
		pc.SQLType = expr.Type()
		pc.MongoType = schema.MongoNone
	}

	return pc
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
	selectIDs         []int
	referencedColumns *exprCountMap

	removeMode bool
}

func newExpressionCollector(selectIDs []int) *expressionCollector {
	return &expressionCollector{
		selectIDs:         selectIDs,
		referencedColumns: newExprCountMap(),
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
	case SQLColumnExpr:
		if containsInt(v.selectIDs, typedN.selectID) {
			if v.removeMode {
				v.referencedColumns.remove(typedN)
			} else {
				v.referencedColumns.add(typedN)
			}
		}
		return typedN, nil
	case *SQLSubqueryExpr:
		if typedN.correlated {
			return walk(v, n)
		}
		return n, nil
	default:
		return walk(v, n)
	}
}

func (c *expressionCollector) getJoinOnVals(ps PlanStage) {
	v := &joinOnVals{c}
	v.visit(ps)
}

type joinOnVals struct {
	exprCollector *expressionCollector
}

func (v *joinOnVals) visit(n node) (node, error) {
	switch typedN := n.(type) {
	case *JoinStage:
		v.exprCollector.Add(typedN.matcher)
		return typedN, nil
	default:
		return walk(v, n)
	}
}
