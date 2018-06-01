package evaluator

import "fmt"

// Node is an interface that represents an AST node.
type Node interface {
	astnode()
}

// LeafNode is an interface that represents an AST leaf node.
type LeafNode interface {
	astLeafNode()
}

// Command is an interface for plan stages that are also SQL commands.
type Command interface {
	Node
	// Authorize is for authorization of the command.  All commands must report
	// whether or not they are authorized. Having this function ensures that we
	// will always require authorization to be thought about as new Commands
	// are added. The method should return nil if authorized, otherwise an
	// error explaining why the user is not authorized.
	Authorize(ctx *ExecutionCtx) error
	// Execute executes the command.
	Execute(ctx *ExecutionCtx) Executor
}

type nodeVisitor interface {
	visit(n Node) (Node, error)
}

type normalizingNode interface {
	// Normalize will attempt to change the Node into
	// a more recognizable form (that may be more amenable
	// to MongoDB's query language) and/or applies short circuiting
	// rules that makes evaluation unnecessary based on
	// recognizable patterns. Each Node is responsible
	// for deciding those patterns itself.
	Normalize() Node
}

// In-Memory leaf PlanStages
func (ps *BSONSourceStage) astLeafNode()    {}
func (ps *CacheStage) astLeafNode()         {}
func (ps *CountStage) astLeafNode()         {}
func (ps *DualStage) astLeafNode()          {}
func (ps *DynamicSourceStage) astLeafNode() {}
func (ps *EmptyStage) astLeafNode()         {}
func (ps *MongoSourceStage) astLeafNode()   {}

// PlanStages
func (ps *BSONSourceStage) astnode()     {}
func (ps *CacheStage) astnode()          {}
func (ps *CountStage) astnode()          {}
func (ps *DynamicSourceStage) astnode()  {}
func (ps *DualStage) astnode()           {}
func (ps *EmptyStage) astnode()          {}
func (ps *FilterStage) astnode()         {}
func (ps *GroupByStage) astnode()        {}
func (ps *JoinStage) astnode()           {}
func (ps *LimitStage) astnode()          {}
func (ps *MongoSourceStage) astnode()    {}
func (ps *OrderByStage) astnode()        {}
func (ps *ProjectStage) astnode()        {}
func (ps *RowGeneratorStage) astnode()   {}
func (ps *SubquerySourceStage) astnode() {}
func (ps *UnionStage) astnode()          {}

// CommandStages
func (c *AlterCommand) astnode() {}
func (c *FlushCommand) astnode() {}
func (c *KillCommand) astnode()  {}
func (c *SetCommand) astnode()   {}

// Expressions
func (m *MongoFilterExpr) astnode()           {}
func (e *SQLAggFunctionExpr) astnode()        {}
func (e *SQLAddExpr) astnode()                {}
func (e *SQLAndExpr) astnode()                {}
func (e *SQLAssignmentExpr) astnode()         {}
func (e *SQLBenchmarkExpr) astnode()          {}
func (e *SQLCaseExpr) astnode()               {}
func (e SQLColumnExpr) astnode()              {}
func (e *SQLConvertExpr) astnode()            {}
func (e *SQLDivideExpr) astnode()             {}
func (e *SQLEqualsExpr) astnode()             {}
func (e *SQLExistsExpr) astnode()             {}
func (e *SQLGreaterThanExpr) astnode()        {}
func (e *SQLGreaterThanOrEqualExpr) astnode() {}
func (e *SQLIDivideExpr) astnode()            {}
func (e *SQLInExpr) astnode()                 {}
func (e *SQLIsExpr) astnode()                 {}
func (e *SQLLessThanExpr) astnode()           {}
func (e *SQLLessThanOrEqualExpr) astnode()    {}
func (e *SQLLikeExpr) astnode()               {}
func (e *SQLModExpr) astnode()                {}
func (e *SQLMultiplyExpr) astnode()           {}
func (e *SQLNotExpr) astnode()                {}
func (e *SQLNotEqualsExpr) astnode()          {}
func (e *SQLNullSafeEqualsExpr) astnode()     {}
func (e *SQLOrExpr) astnode()                 {}
func (e *SQLXorExpr) astnode()                {}
func (e *SQLRegexExpr) astnode()              {}
func (e *SQLScalarFunctionExpr) astnode()     {}
func (e *SQLSubqueryCmpExpr) astnode()        {}
func (e *SQLSubqueryExpr) astnode()           {}
func (e *SQLSubtractExpr) astnode()           {}
func (e *SQLUnaryMinusExpr) astnode()         {}
func (e *SQLUnaryTildeExpr) astnode()         {}
func (e *SQLTupleExpr) astnode()              {}
func (e *SQLVariableExpr) astnode()           {}

// Values
func (v SQLBool) astnode()       {}
func (v SQLDate) astnode()       {}
func (v SQLDecimal128) astnode() {}
func (v SQLFloat) astnode()      {}
func (v SQLInt) astnode()        {}
func (v SQLNoValue) astnode()    {}
func (v SQLNullValue) astnode()  {}
func (v SQLObjectID) astnode()   {}
func (v SQLVarchar) astnode()    {}
func (v SQLTimestamp) astnode()  {}
func (v *SQLValues) astnode()    {}
func (v SQLUint32) astnode()     {}
func (v SQLUint64) astnode()     {}
func (v SQLUUID) astnode()       {}

// walk handles walking the children of the provided expression, calling
// v.visit on each child. Some visitor implementations may ignore this
// method completely, but most will use it as the default implementation
// for a majority of nodes.
func walk(v nodeVisitor, n Node) (Node, error) {
	visitExpr := func(e SQLExpr) (SQLExpr, error) {
		node, err := v.visit(e)
		if err != nil {
			return nil, err
		}

		newE, ok := node.(SQLExpr)
		if !ok {
			return nil, fmt.Errorf("expected SQLExpr, but got %T", node)
		}

		return newE, nil
	}

	visitExprSlice := func(exprs *[]SQLExpr) (*[]SQLExpr, error) {
		hasNew := false
		var newExprs []SQLExpr
		for i, e := range *exprs {
			newE, err := visitExpr(e)
			if err != nil {
				return nil, err
			}

			if !hasNew && e != newE {
				hasNew = true
				newExprs = make([]SQLExpr, i, len(*exprs))
				copy(newExprs, (*exprs)[:i])
			}

			if hasNew {
				newExprs = append(newExprs, newE)
			}
		}

		if hasNew {
			return &newExprs, nil
		}

		return exprs, nil
	}

	visitAssignmentSlice := func(exprs *[]*SQLAssignmentExpr) (*[]*SQLAssignmentExpr, error) {
		hasNew := false
		var newExprs []*SQLAssignmentExpr
		for i, e := range *exprs {
			temp, err := visitExpr(e)
			if err != nil {
				return nil, err
			}

			newE, ok := temp.(*SQLAssignmentExpr)
			if !ok {
				return nil,
					fmt.Errorf("expected an evaluator.*SQLAssignmentExpr, but got a %T",
						temp)
			}

			if !hasNew && e != newE {
				hasNew = true
				newExprs = make([]*SQLAssignmentExpr, i, len(*exprs))
				copy(newExprs, (*exprs)[:i])
			}

			if hasNew {
				newExprs = append(newExprs, newE)
			}
		}

		if hasNew {
			return &newExprs, nil
		}

		return exprs, nil
	}

	visitOrderByTerms := func(terms *[]*OrderByTerm) (*[]*OrderByTerm, error) {
		hasNew := false
		var newTerms []*OrderByTerm
		for i, t := range *terms {
			newE, err := visitExpr(t.expr)
			if err != nil {
				return nil, err
			}

			if !hasNew && t.expr != newE {
				hasNew = true
				newTerms = make([]*OrderByTerm, i, len(*terms))
				copy(newTerms, (*terms)[:i])
			}

			if hasNew {
				newTerms = append(newTerms, &OrderByTerm{
					ascending: t.ascending,
					expr:      newE,
				})
			}
		}

		if hasNew {
			return &newTerms, nil
		}

		return terms, nil
	}

	visitProjectedColumns := func(pcs *ProjectedColumns) (*ProjectedColumns, error) {
		hasNew := false
		var newPcs ProjectedColumns
		for i, pc := range *pcs {
			newE, err := visitExpr(pc.Expr)
			if err != nil {
				return nil, err
			}

			if !hasNew && pc.Expr != newE {
				hasNew = true
				newPcs = make([]ProjectedColumn, i, len(*pcs))
				copy(newPcs, (*pcs)[:i])
			}

			if hasNew {
				newPcs = append(newPcs, ProjectedColumn{
					Column: NewColumn(pc.SelectID,
						pc.Table,
						pc.OriginalTable,
						pc.Database,
						pc.Name,
						pc.OriginalName,
						pc.MappingRegistryName,
						pc.SQLType,
						pc.MongoType,
						pc.PrimaryKey),
					Expr: newE,
				})
			}
		}

		if hasNew {
			return &newPcs, nil
		}

		return pcs, nil
	}

	visitPlanStage := func(s PlanStage) (PlanStage, error) {
		node, err := v.visit(s)
		if err != nil {
			return nil, err
		}
		newS, ok := node.(PlanStage)
		if !ok {
			return nil, fmt.Errorf("expected PlanStage, but got %T", node)
		}

		return newS, nil
	}

	switch typedN := n.(type) {

	case LeafNode:
	// nothing to do for leaf nodes.

	// PlanStages
	case *FilterStage:
		source, err := visitPlanStage(typedN.source)
		if err != nil {
			return nil, err
		}

		matcher, err := visitExpr(typedN.matcher)
		if err != nil {
			return nil, err
		}

		if typedN.source != source || typedN.matcher != matcher {
			n = NewFilterStage(source, matcher)
		}
	case *GroupByStage:
		source, err := visitPlanStage(typedN.source)
		if err != nil {
			return nil, err
		}

		keys, err := visitExprSlice(&typedN.keys)
		if err != nil {
			return nil, err
		}

		pcs, err := visitProjectedColumns(&typedN.projectedColumns)
		if err != nil {
			return nil, err
		}

		if typedN.source != source || &typedN.keys != keys || &typedN.projectedColumns != pcs {
			n = NewGroupByStage(source, *keys, *pcs)
		}
	case *JoinStage:
		left, err := visitPlanStage(typedN.left)
		if err != nil {
			return nil, err
		}

		right, err := visitPlanStage(typedN.right)
		if err != nil {
			return nil, err
		}

		matcher, err := visitExpr(typedN.matcher)
		if err != nil {
			return nil, err
		}

		if typedN.left != left || typedN.right != right || typedN.matcher != matcher {
			n = NewJoinStage(typedN.kind, left, right, matcher)
		}
	case *LimitStage:
		source, err := visitPlanStage(typedN.source)
		if err != nil {
			return nil, err
		}

		if typedN.source != source {
			n = NewLimitStage(source, typedN.offset, typedN.limit)
		}
	case *OrderByStage:
		source, err := visitPlanStage(typedN.source)
		if err != nil {
			return nil, err
		}

		terms, err := visitOrderByTerms(&typedN.terms)
		if err != nil {
			return nil, err
		}

		if typedN.source != source || &typedN.terms != terms {
			n = NewOrderByStage(source, *terms...)
		}
	case *ProjectStage:
		source, err := visitPlanStage(typedN.source)
		if err != nil {
			return nil, err
		}

		pcs, err := visitProjectedColumns(&typedN.projectedColumns)
		if err != nil {
			return nil, err
		}

		if typedN.source != source || &typedN.projectedColumns != pcs {
			n = NewProjectStage(source, *pcs...)
		}
	case *RowGeneratorStage:
		source, err := visitPlanStage(typedN.source)
		if err != nil {
			return nil, err
		}

		if typedN.source != source {
			n = NewRowGeneratorStage(source, typedN.rowCountColumn)
		}
	case *SubquerySourceStage:
		source, err := visitPlanStage(typedN.source)
		if err != nil {
			return nil, err
		}

		if typedN.source != source {
			n = NewSubquerySourceStage(source, typedN.selectID, typedN.dbName, typedN.aliasName)
		}
	case *UnionStage:
		left, err := visitPlanStage(typedN.left)
		if err != nil {
			return nil, err
		}

		right, err := visitPlanStage(typedN.right)
		if err != nil {
			return nil, err
		}

		if typedN.left != left || typedN.right != right {
			n = NewUnionStage(typedN.kind, left, right)
		}

	// Other Stages
	case *KillCommand:
		visitID, err := visitExpr(typedN.ID)
		if err != nil {
			return nil, err
		}

		if typedN.ID != visitID {
			return NewKillCommand(visitID, typedN.Scope), nil
		}
	case *SetCommand:
		exprs, err := visitAssignmentSlice(&typedN.assignments)
		if err != nil {
			return nil, err
		}

		if &typedN.assignments != exprs {
			return NewSetCommand(*exprs), nil
		}
	case *AlterCommand, *FlushCommand:
		// nothing to do

	// Expressions
	case *MongoFilterExpr:
		// nothing to do
	case *SQLAggFunctionExpr:
		exprs, err := visitExprSlice(&typedN.Exprs)
		if err != nil {
			return nil, err
		}

		if &typedN.Exprs != exprs {
			n = &SQLAggFunctionExpr{typedN.Name, typedN.Distinct, *exprs}
		}
	case *SQLAddExpr:
		left, err := visitExpr(typedN.left)
		if err != nil {
			return nil, err
		}
		right, err := visitExpr(typedN.right)
		if err != nil {
			return nil, err
		}

		if typedN.left != left || typedN.right != right {
			n = &SQLAddExpr{left, right}
		}

	case *SQLAndExpr:
		left, err := visitExpr(typedN.left)
		if err != nil {
			return nil, err
		}
		right, err := visitExpr(typedN.right)
		if err != nil {
			return nil, err
		}
		if typedN.left != left || typedN.right != right {
			n = &SQLAndExpr{left, right}
		}

	case *SQLAssignmentExpr:
		temp, err := visitExpr(typedN.variable)
		if err != nil {
			return nil, err
		}

		variable, ok := temp.(*SQLVariableExpr)
		if !ok {
			return nil,
				fmt.Errorf("SQLAssignmentExpr requires an evaluator.*SQLVariableExpr, but got a %T",
					temp)
		}

		expr, err := visitExpr(typedN.expr)
		if err != nil {
			return nil, err
		}

		if typedN.variable != variable || typedN.expr != expr {
			n = &SQLAssignmentExpr{
				variable: variable,
				expr:     expr,
			}
		}
	case *SQLBenchmarkExpr:
		// Don't optimize.

	case *SQLCaseExpr:
		hasNewCond := false
		newConds := []caseCondition{}
		for _, cond := range typedN.caseConditions {
			m, err := visitExpr(cond.matcher)
			if err != nil {
				return nil, err
			}
			t, err := visitExpr(cond.then)
			if err != nil {
				return nil, err
			}

			newCond := cond
			if cond.matcher != m || cond.then != t {
				newCond = caseCondition{m, t}
				hasNewCond = true
			}

			newConds = append(newConds, newCond)
		}

		newElse, err := visitExpr(typedN.elseValue)
		if err != nil {
			return nil, err
		}

		if hasNewCond || typedN.elseValue != newElse {
			n = &SQLCaseExpr{newElse, newConds}
		}
	case SQLColumnExpr:
		// no children
	case *SQLConvertExpr:
		expr, err := visitExpr(typedN.expr)
		if err != nil {
			return nil, err
		}
		if typedN.expr != expr {
			n = &SQLConvertExpr{expr, typedN.convType, SQLNone}
		}
	case *SQLDivideExpr:
		left, err := visitExpr(typedN.left)
		if err != nil {
			return nil, err
		}
		right, err := visitExpr(typedN.right)
		if err != nil {
			return nil, err
		}
		if typedN.left != left || typedN.right != right {
			n = &SQLDivideExpr{left, right}
		}
	case *SQLEqualsExpr:
		left, err := visitExpr(typedN.left)
		if err != nil {
			return nil, err
		}
		right, err := visitExpr(typedN.right)
		if err != nil {
			return nil, err
		}
		if typedN.left != left || typedN.right != right {
			n = &SQLEqualsExpr{left, right}
		}

	case *SQLExistsExpr:
		expr, err := visitExpr(typedN.expr)
		if err != nil {
			return nil, err
		}

		sub, ok := expr.(*SQLSubqueryExpr)
		if !ok {
			return nil, fmt.Errorf("SQLExistsExpr"+
				" requires an evaluator.*SQLSubqueryExpr, but got a %T", sub)
		}

		if typedN.expr != expr {
			n = &SQLExistsExpr{sub}
		}
	case *SQLGreaterThanExpr:
		left, err := visitExpr(typedN.left)
		if err != nil {
			return nil, err
		}
		right, err := visitExpr(typedN.right)
		if err != nil {
			return nil, err
		}
		if typedN.left != left || typedN.right != right {
			n = &SQLGreaterThanExpr{left, right}
		}

	case *SQLGreaterThanOrEqualExpr:
		left, err := visitExpr(typedN.left)
		if err != nil {
			return nil, err
		}
		right, err := visitExpr(typedN.right)
		if err != nil {
			return nil, err
		}
		if typedN.left != left || typedN.right != right {
			n = &SQLGreaterThanOrEqualExpr{left, right}
		}

	case *SQLIDivideExpr:
		left, err := visitExpr(typedN.left)
		if err != nil {
			return nil, err
		}
		right, err := visitExpr(typedN.right)
		if err != nil {
			return nil, err
		}
		if typedN.left != left || typedN.right != right {
			n = &SQLIDivideExpr{left, right}
		}

	case *SQLInExpr:
		left, err := visitExpr(typedN.left)
		if err != nil {
			return nil, err
		}
		right, err := visitExpr(typedN.right)
		if err != nil {
			return nil, err
		}
		if typedN.left != left || typedN.right != right {
			n = &SQLInExpr{left, right}
		}
	case *SQLIsExpr:
		// The right child does not need to be evaluated because it
		// will only ever be True, False, Null or Unknown.
		left, err := visitExpr(typedN.left)
		if err != nil {
			return nil, err
		}

		if typedN.left != left {
			n = NewSQLIsExpr(left, typedN.right)
		}

	case *SQLLessThanExpr:
		left, err := visitExpr(typedN.left)
		if err != nil {
			return nil, err
		}
		right, err := visitExpr(typedN.right)
		if err != nil {
			return nil, err
		}
		if typedN.left != left || typedN.right != right {
			n = &SQLLessThanExpr{left, right}
		}

	case *SQLLessThanOrEqualExpr:
		left, err := visitExpr(typedN.left)
		if err != nil {
			return nil, err
		}
		right, err := visitExpr(typedN.right)
		if err != nil {
			return nil, err
		}
		if typedN.left != left || typedN.right != right {
			n = &SQLLessThanOrEqualExpr{left, right}
		}

	case *SQLLikeExpr:
		left, err := visitExpr(typedN.left)
		if err != nil {
			return nil, err
		}
		right, err := visitExpr(typedN.right)
		if err != nil {
			return nil, err
		}
		escape, err := visitExpr(typedN.escape)
		if err != nil {
			return nil, err
		}
		if typedN.left != left || typedN.right != right || typedN.escape != escape {
			n = &SQLLikeExpr{left, right, escape}
		}

	case *SQLModExpr:
		left, err := visitExpr(typedN.left)
		if err != nil {
			return nil, err
		}
		right, err := visitExpr(typedN.right)
		if err != nil {
			return nil, err
		}
		if typedN.left != left || typedN.right != right {
			n = &SQLModExpr{left, right}
		}

	case *SQLMultiplyExpr:
		left, err := visitExpr(typedN.left)
		if err != nil {
			return nil, err
		}
		right, err := visitExpr(typedN.right)
		if err != nil {
			return nil, err
		}
		if typedN.left != left || typedN.right != right {
			n = &SQLMultiplyExpr{left, right}
		}

	case *SQLNotExpr:
		operand, err := visitExpr(typedN.SQLExpr)
		if err != nil {
			return nil, err
		}
		if typedN.SQLExpr != operand {
			n = &SQLNotExpr{operand}
		}

	case *SQLNotEqualsExpr:
		left, err := visitExpr(typedN.left)
		if err != nil {
			return nil, err
		}
		right, err := visitExpr(typedN.right)
		if err != nil {
			return nil, err
		}
		if typedN.left != left || typedN.right != right {
			n = &SQLNotEqualsExpr{left, right}
		}

	case *SQLNullSafeEqualsExpr:
		left, err := visitExpr(typedN.left)
		if err != nil {
			return nil, err
		}
		right, err := visitExpr(typedN.right)
		if err != nil {
			return nil, err
		}
		if typedN.left != left || typedN.right != right {
			n = &SQLNullSafeEqualsExpr{left, right}
		}

	case *SQLOrExpr:
		left, err := visitExpr(typedN.left)
		if err != nil {
			return nil, err
		}
		right, err := visitExpr(typedN.right)
		if err != nil {
			return nil, err
		}
		if typedN.left != left || typedN.right != right {
			n = &SQLOrExpr{left, right}
		}

	case *SQLXorExpr:
		left, err := visitExpr(typedN.left)
		if err != nil {
			return nil, err
		}
		right, err := visitExpr(typedN.right)
		if err != nil {
			return nil, err
		}
		if typedN.left != left || typedN.right != right {
			n = &SQLXorExpr{left, right}
		}

	case *SQLRegexExpr:
		operand, err := visitExpr(typedN.operand)
		if err != nil {
			return nil, err
		}
		pattern, err := visitExpr(typedN.pattern)
		if err != nil {
			return nil, err
		}
		if typedN.operand != operand || typedN.pattern != pattern {
			n = &SQLRegexExpr{operand: operand, pattern: pattern}
		}

	case *SQLScalarFunctionExpr:

		exprs, err := visitExprSlice(&typedN.Exprs)
		if err != nil {
			return nil, err
		}

		if &typedN.Exprs != exprs {
			n, err = NewSQLScalarFunctionExpr(typedN.Name, *exprs)
			if err != nil {
				return nil, err
			}
		}
	case *SQLSubqueryCmpExpr:
		left, err := visitExpr(typedN.left)
		if err != nil {
			return nil, err
		}
		sub, err := visitExpr(typedN.subqueryExpr)
		if err != nil {
			return nil, err
		}

		subqueryExpr, ok := sub.(*SQLSubqueryExpr)
		if !ok {
			return nil,
				fmt.Errorf("SQLSubqueryCmpExpr"+
					" requires an evaluator.*SQLSubqueryExpr, but got a %T",
					sub)
		}

		if typedN.left != left || typedN.subqueryExpr != subqueryExpr {
			n = &SQLSubqueryCmpExpr{typedN.subqueryOp, left, subqueryExpr, typedN.operator}
		}

	case *SQLSubqueryExpr:
		plan, err := visitPlanStage(typedN.plan)
		if err != nil {
			return nil, err
		}

		if typedN.plan != plan {
			n = &SQLSubqueryExpr{
				correlated: typedN.correlated,
				plan:       plan,
				allowRows:  typedN.allowRows,
			}
		}
	case *SQLSubtractExpr:
		left, err := visitExpr(typedN.left)
		if err != nil {
			return nil, err
		}
		right, err := visitExpr(typedN.right)
		if err != nil {
			return nil, err
		}
		if typedN.left != left || typedN.right != right {
			n = &SQLSubtractExpr{left, right}
		}

	case *SQLUnaryMinusExpr:
		operand, err := visitExpr(typedN.SQLExpr)
		if err != nil {
			return nil, err
		}
		if typedN.SQLExpr != operand {
			n = &SQLUnaryMinusExpr{operand}
		}

	case *SQLUnaryTildeExpr:
		operand, err := visitExpr(typedN.SQLExpr)
		if err != nil {
			return nil, err
		}
		if typedN.SQLExpr != operand {
			n = &SQLUnaryTildeExpr{operand}
		}

	case *SQLTupleExpr:
		exprs, err := visitExprSlice(&typedN.Exprs)
		if err != nil {
			return nil, err
		}
		if &typedN.Exprs != exprs {
			n = &SQLTupleExpr{*exprs}
		}
	case *SQLVariableExpr:
		// nothing to do

	// values
	case SQLBool,
		SQLDate,
		SQLDecimal128,
		SQLFloat,
		SQLInt,
		SQLNoValue,
		SQLNullValue,
		SQLObjectID,
		SQLVarchar,
		SQLTimestamp,
		SQLUint32,
		SQLUint64,
		SQLUUID:
		// nothing to do
	case *SQLValues:
		hasNewValue := false
		newValues := []SQLValue{}
		for _, value := range typedN.Values {
			newValueExpr, err := visitExpr(value)
			if err != nil {
				return nil, err
			}
			newValue, ok := newValueExpr.(SQLValue)
			if !ok {
				return nil,
					fmt.Errorf("evaluator.SQLValues requires an evaluator.SQLValue, but got a %T",
						newValueExpr)
			}

			if value != newValue {
				hasNewValue = true
			}

			newValues = append(newValues, newValue)
		}

		if hasNewValue {
			n = &SQLValues{newValues}
		}
	default:
		return nil, fmt.Errorf("unsupported node: %T", typedN)
	}

	return n, nil
}
