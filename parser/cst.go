package parser

// A node in the CST.
// Currently this is a uniform interface on top of differently-shaped
// parse nodes.
type CST interface {
	// Children iterates through all direct children of this node.
	Children() []CST
	// ReplaceChild changes the value of a particular child node.
	ReplaceChild(i int, child CST)
	// Copy produces a deep copy of this node.
	Copy() CST
}

// Children iterates through all direct children of this node.
func (node *Use) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node *Use) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node *Use) Copy() CST {
	var newDBName []byte
	if node.DBName == nil {
		newDBName = nil
	} else {
		newDBName = make([]byte, len(node.DBName))
		copy(newDBName, node.DBName)
	}

	return &Use{
		newDBName,
	}
}

var _ CST = (*Use)(nil)

// Children iterates through all direct children of this node.
func (node *CTE) Children() []CST {
	return []CST{
		node.TableName,
		node.ColumnExprs,
		node.Query,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *CTE) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.TableName = nil
		} else {
			node.TableName = child.(*TableName)
		}
	case 1:
		if child == nil {
			node.ColumnExprs = nil
		} else {
			node.ColumnExprs = child.(ColumnExprs)
		}
	case 2:
		if child == nil {
			node.Query = nil
		} else {
			node.Query = child.(SelectStatement)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *CTE) Copy() CST {
	var tableName *TableName
	if node.TableName == nil {
		tableName = nil
	} else {
		tableName = node.TableName.Copy().(*TableName)
	}
	var columnExprs ColumnExprs
	if node.ColumnExprs == nil {
		columnExprs = nil
	} else {
		columnExprs = node.ColumnExprs.Copy().(ColumnExprs)
	}
	var query SelectStatement
	if node.Query == nil {
		query = nil
	} else {
		query = node.Query.Copy().(SelectStatement)
	}

	return &CTE{
		tableName,
		columnExprs,
		query,
	}
}

var _ CST = (*CTE)(nil)

// Children iterates through all direct children of this node.
func (node CTEs) Children() []CST {
	var result []CST
	for _, cte := range node {
		result = append(result, cte)
	}
	return result
}

// ReplaceChild changes the value of a particular child node.
func (node CTEs) ReplaceChild(i int, child CST) {
	if child == nil {
		node[i] = nil
	} else {
		node[i] = child.(*CTE)
	}
}

// Copy produces a deep copy of this node.
func (node CTEs) Copy() CST {
	newNode := make(CTEs, len(node))
	for i, cte := range node {
		if cte == nil {
			newNode[i] = nil
		} else {
			newNode[i] = cte.Copy().(*CTE)
		}
	}
	return newNode
}

var _ CST = (CTEs)(nil)

// Children iterates through all direct children of this node.
func (node *With) Children() []CST {
	return []CST{
		node.CTEs,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *With) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.CTEs = nil
		} else {
			node.CTEs = child.(CTEs)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *With) Copy() CST {
	var ctes CTEs
	if node.CTEs == nil {
		ctes = nil
	} else {
		ctes = node.CTEs.Copy().(CTEs)
	}

	return &With{
		ctes,
		node.Recursive,
	}
}

var _ CST = (*With)(nil)

// Children iterates through all direct children of this node.
func (node *Select) Children() []CST {
	result := []CST{
		node.With,
		node.Comments,
		node.QueryGlobals,
		node.SelectExprs,
		node.From,
		node.Where,
		node.GroupBy,
		node.Having,
		node.OrderBy,
		node.Limit,
	}
	return result
}

// ReplaceChild changes the value of a particular child node.
func (node *Select) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.With = nil
		} else {
			node.With = child.(*With)
		}
	case 1:
		if child == nil {
			node.Comments = nil
		} else {
			node.Comments = child.(Comments)
		}
	case 2:
		if child == nil {
			node.QueryGlobals = nil
		} else {
			node.QueryGlobals = child.(*QueryGlobals)
		}
	case 3:
		if child == nil {
			node.SelectExprs = nil
		} else {
			node.SelectExprs = child.(SelectExprs)
		}
	case 4:
		if child == nil {
			node.From = nil
		} else {
			node.From = child.(TableExprs)
		}
	case 5:
		if child == nil {
			node.Where = nil
		} else {
			node.Where = child.(*Where)
		}
	case 6:
		if child == nil {
			node.GroupBy = nil
		} else {
			node.GroupBy = child.(GroupBy)
		}
	case 7:
		if child == nil {
			node.Having = nil
		} else {
			node.Having = child.(*Where)
		}
	case 8:
		if child == nil {
			node.OrderBy = nil
		} else {
			node.OrderBy = child.(OrderBy)
		}
	case 9:
		if child == nil {
			node.Limit = nil
		} else {
			node.Limit = child.(*Limit)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *Select) Copy() CST {
	var with *With
	if node.With == nil {
		with = nil
	} else {
		with = node.With.Copy().(*With)
	}
	var comments Comments
	if node.Comments == nil {
		comments = nil
	} else {
		comments = node.Comments.Copy().(Comments)
	}
	var queryGlobals *QueryGlobals
	if node.QueryGlobals == nil {
		queryGlobals = nil
	} else {
		queryGlobals = node.QueryGlobals.Copy().(*QueryGlobals)
	}
	var selectExprs SelectExprs
	if node.SelectExprs == nil {
		selectExprs = nil
	} else {
		selectExprs = node.SelectExprs.Copy().(SelectExprs)
	}
	var from TableExprs
	if node.From == nil {
		from = nil
	} else {
		from = node.From.Copy().(TableExprs)
	}
	var where *Where
	if node.Where == nil {
		where = nil
	} else {
		where = node.Where.Copy().(*Where)
	}
	var groupBy GroupBy
	if node.GroupBy == nil {
		groupBy = nil
	} else {
		groupBy = node.GroupBy.Copy().(GroupBy)
	}
	var having *Where
	if node.Having == nil {
		having = nil
	} else {
		having = node.Having.Copy().(*Where)
	}
	var orderBy OrderBy
	if node.OrderBy == nil {
		orderBy = nil
	} else {
		orderBy = node.OrderBy.Copy().(OrderBy)
	}
	var limit *Limit
	if node.Limit == nil {
		limit = nil
	} else {
		limit = node.Limit.Copy().(*Limit)
	}

	return &Select{
		with,
		comments,
		queryGlobals,
		selectExprs,
		from,
		where,
		groupBy,
		having,
		orderBy,
		limit,
		node.Lock,
	}
}

var _ CST = (*Select)(nil)

// Children iterates through all direct children of this node.
func (node *QueryGlobals) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node *QueryGlobals) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node *QueryGlobals) Copy() CST {
	return &QueryGlobals{
		node.Distinct,
		node.StraightJoin,
	}
}

var _ CST = (*QueryGlobals)(nil)

// Children iterates through all direct children of this node.
func (node *Union) Children() []CST {
	return []CST{
		node.With,
		node.Left,
		node.Right,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *Union) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.With = nil
		} else {
			node.With = child.(*With)
		}
	case 1:
		if child == nil {
			node.Left = nil
		} else {
			node.Left = child.(SelectStatement)
		}
	case 2:
		if child == nil {
			node.Right = nil
		} else {
			node.Right = child.(SelectStatement)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *Union) Copy() CST {
	var with *With
	if node.With == nil {
		with = nil
	} else {
		with = node.With.Copy().(*With)
	}
	var left SelectStatement
	if node.Left == nil {
		left = nil
	} else {
		left = node.Left.Copy().(SelectStatement)
	}
	var right SelectStatement
	if node.Right == nil {
		right = nil
	} else {
		right = node.Right.Copy().(SelectStatement)
	}

	return &Union{
		with,
		node.Type,
		left,
		right,
	}
}

var _ CST = (*Union)(nil)

// Children iterates through all direct children of this node.
func (node *Set) Children() []CST {
	return []CST{
		node.Comments,
		node.Exprs,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *Set) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Comments = nil
		} else {
			node.Comments = child.(Comments)
		}
	case 1:
		if child == nil {
			node.Exprs = nil
		} else {
			node.Exprs = child.(UpdateExprs)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *Set) Copy() CST {
	var comments Comments
	if node.Comments == nil {
		comments = nil
	} else {
		comments = node.Comments.Copy().(Comments)
	}
	var exprs UpdateExprs
	if node.Exprs == nil {
		exprs = nil
	} else {
		exprs = node.Exprs.Copy().(UpdateExprs)
	}

	return &Set{
		node.Scope,
		comments,
		exprs,
	}
}

var _ CST = (*Set)(nil)

// Children iterates through all direct children of this node.
func (node *DropTable) Children() []CST {
	return []CST{
		node.Name,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *DropTable) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Name = nil
		} else {
			node.Name = child.(*TableName)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *DropTable) Copy() CST {
	var name *TableName
	if node.Name == nil {
		name = nil
	} else {
		name = node.Name.Copy().(*TableName)
	}
	var newOpt []byte
	if node.Opt == nil {
		newOpt = nil
	} else {
		newOpt = make([]byte, len(node.Opt))
		copy(newOpt, node.Opt)
	}

	return &DropTable{
		name,
		node.Exists,
		node.Temporary,
		newOpt,
	}
}

var _ CST = (*DropTable)(nil)

// Children iterates through all direct children of this node.
func (node Comments) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node Comments) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node Comments) Copy() CST {
	newComments := make(Comments, len(node))
	for i, comm := range node {
		var newComment []byte
		if comm == nil {
			newComment = nil
		} else {
			newComment = make([]byte, len(comm))
			copy(newComment, comm)
		}

		newComments[i] = newComment
	}
	return newComments
}

var _ CST = (Comments)(nil)

// Children iterates through all direct children of this node.
func (node SelectExprs) Children() []CST {
	var result []CST
	for _, child := range node {
		result = append(result, child)
	}
	return result
}

// ReplaceChild changes the value of a particular child node.
func (node SelectExprs) ReplaceChild(i int, child CST) {
	if child == nil {
		node[i] = nil
	} else {
		node[i] = child.(SelectExpr)
	}
}

// Copy produces a deep copy of this node.
func (node SelectExprs) Copy() CST {
	newNode := make(SelectExprs, len(node))
	for i, se := range node {
		if se == nil {
			newNode[i] = nil
		} else {
			newNode[i] = se.Copy().(SelectExpr)
		}
	}
	return newNode
}

var _ CST = (SelectExprs)(nil)

// Children iterates through all direct children of this node.
func (node *StarExpr) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node *StarExpr) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node *StarExpr) Copy() CST {
	var newDatabaseName []byte
	if node.DatabaseName == nil {
		newDatabaseName = nil
	} else {
		newDatabaseName = make([]byte, len(node.DatabaseName))
		copy(newDatabaseName, node.DatabaseName)
	}
	var newTableName []byte
	if node.TableName == nil {
		newTableName = nil
	} else {
		newTableName = make([]byte, len(node.TableName))
		copy(newTableName, node.TableName)
	}

	return &StarExpr{
		newDatabaseName,
		newTableName,
	}
}

var _ CST = (*StarExpr)(nil)

// Children iterates through all direct children of this node.
func (node *NonStarExpr) Children() []CST {
	return []CST{
		node.Expr,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *NonStarExpr) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Expr = nil
		} else {
			node.Expr = child.(Expr)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *NonStarExpr) Copy() CST {
	var expr Expr
	if node.Expr == nil {
		expr = nil
	} else {
		expr = node.Expr.Copy().(Expr)
	}
	var newAs []byte
	if node.As == nil {
		newAs = nil
	} else {
		newAs = make([]byte, len(node.As))
		copy(newAs, node.As)
	}

	return &NonStarExpr{
		expr,
		newAs,
	}
}

var _ CST = (*NonStarExpr)(nil)

// Children iterates through all direct children of this node.
func (node Columns) Children() []CST {
	var result []CST
	for _, child := range node {
		result = append(result, child)
	}
	return result
}

// ReplaceChild changes the value of a particular child node.
func (node Columns) ReplaceChild(i int, child CST) {
	if child == nil {
		node[i] = nil
	} else {
		node[i] = child.(SelectExpr)
	}
}

// Copy produces a deep copy of this node.
func (node Columns) Copy() CST {
	newNode := make(Columns, len(node))
	for i, se := range node {
		if se == nil {
			newNode[i] = nil
		} else {
			newNode[i] = se.Copy().(SelectExpr)
		}
	}
	return newNode
}

var _ CST = (Columns)(nil)

// Children iterates through all direct children of this node.
func (node ColumnExprs) Children() []CST {
	var result []CST
	for _, child := range node {
		result = append(result, child)
	}
	return result
}

// ReplaceChild changes the value of a particular child node.
func (node ColumnExprs) ReplaceChild(i int, child CST) {
	if child == nil {
		node[i] = nil
	} else {
		node[i] = child.(*ColName)
	}
}

// Copy produces a deep copy of this node.
func (node ColumnExprs) Copy() CST {
	newNode := make(ColumnExprs, len(node))
	for i, cn := range node {
		if cn == nil {
			newNode[i] = nil
		} else {
			newNode[i] = cn.Copy().(*ColName)
		}
	}
	return newNode
}

var _ CST = (ColumnExprs)(nil)

// Children iterates through all direct children of this node.
func (node TableExprs) Children() []CST {
	var result []CST
	for _, child := range node {
		result = append(result, child)
	}
	return result
}

// ReplaceChild changes the value of a particular child node.
func (node TableExprs) ReplaceChild(i int, child CST) {
	if child == nil {
		node[i] = nil
	} else {
		node[i] = child.(TableExpr)
	}
}

// Copy produces a deep copy of this node.
func (node TableExprs) Copy() CST {
	newNode := make(TableExprs, len(node))
	for i, te := range node {
		if te == nil {
			newNode[i] = nil
		} else {
			newNode[i] = te.Copy().(TableExpr)
		}
	}
	return newNode
}

var _ CST = (TableExprs)(nil)

// Children iterates through all direct children of this node.
func (node *AliasedTableExpr) Children() []CST {
	return []CST{
		node.Expr,
		node.Hints,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *AliasedTableExpr) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Expr = nil
		} else {
			node.Expr = child.(SimpleTableExpr)
		}
	case 1:
		if child == nil {
			node.Hints = nil
		} else {
			node.Hints = child.(*IndexHints)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *AliasedTableExpr) Copy() CST {
	var expr SimpleTableExpr
	if node.Expr == nil {
		expr = nil
	} else {
		expr = node.Expr.Copy().(SimpleTableExpr)
	}
	var newAs []byte
	if node.As == nil {
		newAs = nil
	} else {
		newAs = make([]byte, len(node.As))
		copy(newAs, node.As)
	}
	var hints *IndexHints
	if node.Hints == nil {
		hints = nil
	} else {
		hints = node.Hints.Copy().(*IndexHints)
	}

	return &AliasedTableExpr{
		expr,
		newAs,
		hints,
	}
}

var _ CST = (*AliasedTableExpr)(nil)

// Children iterates through all direct children of this node.
func (node *TableName) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node *TableName) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node *TableName) Copy() CST {
	var newName []byte
	if node.Name == nil {
		newName = nil
	} else {
		newName = make([]byte, len(node.Name))
		copy(newName, node.Name)
	}
	var newQualifier []byte
	if node.Qualifier == nil {
		newQualifier = nil
	} else {
		newQualifier = make([]byte, len(node.Qualifier))
		copy(newQualifier, node.Qualifier)
	}

	return &TableName{
		newName,
		newQualifier,
	}
}

var _ CST = (*TableName)(nil)

// Children iterates through all direct children of this node.
func (node *ParenTableExpr) Children() []CST {
	return []CST{
		node.Expr,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *ParenTableExpr) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Expr = nil
		} else {
			node.Expr = child.(TableExpr)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *ParenTableExpr) Copy() CST {
	var expr TableExpr
	if node.Expr == nil {
		expr = nil
	} else {
		expr = node.Expr.Copy().(TableExpr)
	}

	return &ParenTableExpr{
		expr,
	}
}

var _ CST = (*ParenTableExpr)(nil)

// Children iterates through all direct children of this node.
func (node *JoinTableExpr) Children() []CST {
	return []CST{
		node.LeftExpr,
		node.RightExpr,
		node.On,
		node.Using,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *JoinTableExpr) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.LeftExpr = nil
		} else {
			node.LeftExpr = child.(TableExpr)
		}
	case 1:
		if child == nil {
			node.RightExpr = nil
		} else {
			node.RightExpr = child.(TableExpr)
		}
	case 2:
		if child == nil {
			node.On = nil
		} else {
			node.On = child.(Expr)
		}
	case 3:
		if child == nil {
			node.Using = nil
		} else {
			node.Using = child.(ColumnExprs)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *JoinTableExpr) Copy() CST {
	var leftExpr TableExpr
	if node.LeftExpr == nil {
		leftExpr = nil
	} else {
		leftExpr = node.LeftExpr.Copy().(TableExpr)
	}
	var rightExpr TableExpr
	if node.RightExpr == nil {
		rightExpr = nil
	} else {
		rightExpr = node.RightExpr.Copy().(TableExpr)
	}
	var on Expr
	if node.On == nil {
		on = nil
	} else {
		on = node.On.Copy().(Expr)
	}
	var using ColumnExprs
	if node.Using == nil {
		using = nil
	} else {
		using = node.Using.Copy().(ColumnExprs)
	}

	return &JoinTableExpr{
		leftExpr,
		node.Join,
		rightExpr,
		on,
		using,
	}
}

var _ CST = (*JoinTableExpr)(nil)

// Children iterates through all direct children of this node.
func (node *IndexHints) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node *IndexHints) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node *IndexHints) Copy() CST {
	newIndexes := make([][]byte, len(node.Indexes))
	for i, index := range node.Indexes {
		var newIndex []byte
		if index == nil {
			newIndex = nil
		} else {
			newIndex := make([]byte, len(index))
			copy(newIndex, index)
		}

		newIndexes[i] = newIndex
	}
	return &IndexHints{
		node.Type,
		newIndexes,
	}
}

var _ CST = (*IndexHints)(nil)

// Children iterates through all direct children of this node.
func (node *Where) Children() []CST {
	return []CST{
		node.Expr,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *Where) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Expr = nil
		} else {
			node.Expr = child.(Expr)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *Where) Copy() CST {
	var expr Expr
	if node.Expr == nil {
		expr = nil
	} else {
		expr = node.Expr.Copy().(Expr)
	}

	return &Where{
		node.Type,
		expr,
	}
}

var _ CST = (*Where)(nil)

// Children iterates through all direct children of this node.
func (node *AndExpr) Children() []CST {
	return []CST{
		node.Left,
		node.Right,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *AndExpr) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Left = nil
		} else {
			node.Left = child.(Expr)
		}
	case 1:
		if child == nil {
			node.Right = nil
		} else {
			node.Right = child.(Expr)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *AndExpr) Copy() CST {
	var left Expr
	if node.Left == nil {
		left = nil
	} else {
		left = node.Left.Copy().(Expr)
	}
	var right Expr
	if node.Right == nil {
		right = nil
	} else {
		right = node.Right.Copy().(Expr)
	}

	return &AndExpr{
		left,
		right,
	}
}

var _ CST = (*AndExpr)(nil)

// Children iterates through all direct children of this node.
func (node *OrExpr) Children() []CST {
	return []CST{
		node.Left,
		node.Right,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *OrExpr) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Left = nil
		} else {
			node.Left = child.(Expr)
		}
	case 1:
		if child == nil {
			node.Right = nil
		} else {
			node.Right = child.(Expr)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *OrExpr) Copy() CST {
	var left Expr
	if node.Left == nil {
		left = nil
	} else {
		left = node.Left.Copy().(Expr)
	}
	var right Expr
	if node.Right == nil {
		right = nil
	} else {
		right = node.Right.Copy().(Expr)
	}

	return &OrExpr{
		left,
		right,
	}
}

var _ CST = (*OrExpr)(nil)

// Children iterates through all direct children of this node.
func (node *XorExpr) Children() []CST {
	return []CST{
		node.Left,
		node.Right,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *XorExpr) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Left = nil
		} else {
			node.Left = child.(Expr)
		}
	case 1:
		if child == nil {
			node.Right = nil
		} else {
			node.Right = child.(Expr)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *XorExpr) Copy() CST {
	var left Expr
	if node.Left == nil {
		left = nil
	} else {
		left = node.Left.Copy().(Expr)
	}
	var right Expr
	if node.Right == nil {
		right = nil
	} else {
		right = node.Right.Copy().(Expr)
	}

	return &XorExpr{
		left,
		right,
	}
}

var _ CST = (*XorExpr)(nil)

// Children iterates through all direct children of this node.
func (node *NotExpr) Children() []CST {
	return []CST{
		node.Expr,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *NotExpr) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Expr = nil
		} else {
			node.Expr = child.(Expr)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *NotExpr) Copy() CST {
	var expr Expr
	if node.Expr == nil {
		expr = nil
	} else {
		expr = node.Expr.Copy().(Expr)
	}

	return &NotExpr{
		expr,
	}
}

var _ CST = (*NotExpr)(nil)

// Children iterates through all direct children of this node.
func (node *ComparisonExpr) Children() []CST {
	return []CST{
		node.Left,
		node.Right,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *ComparisonExpr) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Left = nil
		} else {
			node.Left = child.(Expr)
		}
	case 1:
		if child == nil {
			node.Right = nil
		} else {
			node.Right = child.(Expr)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *ComparisonExpr) Copy() CST {
	var left Expr
	if node.Left == nil {
		left = nil
	} else {
		left = node.Left.Copy().(Expr)
	}
	var right Expr
	if node.Right == nil {
		right = nil
	} else {
		right = node.Right.Copy().(Expr)
	}

	return &ComparisonExpr{
		node.Operator,
		left,
		right,
		node.SubqueryOperator,
	}
}

var _ CST = (*ComparisonExpr)(nil)

// Children iterates through all direct children of this node.
func (node *LikeExpr) Children() []CST {
	return []CST{
		node.Left,
		node.Right,
		node.Escape,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *LikeExpr) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Left = nil
		} else {
			node.Left = child.(Expr)
		}
	case 1:
		if child == nil {
			node.Right = nil
		} else {
			node.Right = child.(Expr)
		}
	case 2:
		if child == nil {
			node.Escape = nil
		} else {
			node.Escape = child.(Expr)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *LikeExpr) Copy() CST {
	var left Expr
	if node.Left == nil {
		left = nil
	} else {
		left = node.Left.Copy().(Expr)
	}
	var right Expr
	if node.Right == nil {
		right = nil
	} else {
		right = node.Right.Copy().(Expr)
	}
	var escape Expr
	if node.Escape == nil {
		escape = nil
	} else {
		escape = node.Escape.Copy().(Expr)
	}

	return &LikeExpr{
		node.Operator,
		left,
		right,
		escape,
	}
}

var _ CST = (*LikeExpr)(nil)

// Children iterates through all direct children of this node.
func (node *RangeCond) Children() []CST {
	return []CST{
		node.Left,
		node.From,
		node.To,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *RangeCond) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Left = nil
		} else {
			node.Left = child.(Expr)
		}
	case 1:
		if child == nil {
			node.From = nil
		} else {
			node.From = child.(Expr)
		}
	case 2:
		if child == nil {
			node.To = nil
		} else {
			node.To = child.(Expr)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *RangeCond) Copy() CST {
	var left Expr
	if node.Left == nil {
		left = nil
	} else {
		left = node.Left.Copy().(Expr)
	}
	var from Expr
	if node.From == nil {
		from = nil
	} else {
		from = node.From.Copy().(Expr)
	}
	var to Expr
	if node.To == nil {
		to = nil
	} else {
		to = node.To.Copy().(Expr)
	}

	return &RangeCond{
		node.Operator,
		left,
		from,
		to,
	}
}

var _ CST = (*RangeCond)(nil)

// Children iterates through all direct children of this node.
func (node *RegexExpr) Children() []CST {
	return []CST{
		node.Operand,
		node.Pattern,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *RegexExpr) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Operand = nil
		} else {
			node.Operand = child.(Expr)
		}
	case 1:
		if child == nil {
			node.Pattern = nil
		} else {
			node.Pattern = child.(Expr)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *RegexExpr) Copy() CST {
	var operand Expr
	if node.Operand == nil {
		operand = nil
	} else {
		operand = node.Operand.Copy().(Expr)
	}
	var pattern Expr
	if node.Pattern == nil {
		pattern = nil
	} else {
		pattern = node.Pattern.Copy().(Expr)
	}

	return &RegexExpr{
		operand,
		pattern,
	}
}

var _ CST = (*RegexExpr)(nil)

// Children iterates through all direct children of this node.
func (node *RLikeExpr) Children() []CST {
	return []CST{
		node.Operand,
		node.Pattern,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *RLikeExpr) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Operand = nil
		} else {
			node.Operand = child.(Expr)
		}
	case 1:
		if child == nil {
			node.Pattern = nil
		} else {
			node.Pattern = child.(Expr)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *RLikeExpr) Copy() CST {
	var operand Expr
	if node.Operand == nil {
		operand = nil
	} else {
		operand = node.Operand.Copy().(Expr)
	}
	var pattern Expr
	if node.Pattern == nil {
		pattern = nil
	} else {
		pattern = node.Pattern.Copy().(Expr)
	}

	return &RLikeExpr{
		operand,
		pattern,
	}
}

var _ CST = (*RLikeExpr)(nil)

// Children iterates through all direct children of this node.
func (node *ExistsExpr) Children() []CST {
	return []CST{
		node.Subquery,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *ExistsExpr) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Subquery = nil
		} else {
			node.Subquery = child.(*Subquery)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *ExistsExpr) Copy() CST {
	var subquery *Subquery
	if node.Subquery == nil {
		subquery = nil
	} else {
		subquery = node.Subquery.Copy().(*Subquery)
	}

	return &ExistsExpr{
		subquery,
	}
}

var _ CST = (*ExistsExpr)(nil)

// Children iterates through all direct children of this node.
func (node *DateVal) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node *DateVal) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node *DateVal) Copy() CST {
	var newVal []byte
	if node.Val == nil {
		newVal = nil
	} else {
		newVal = make([]byte, len(node.Val))
		copy(newVal, node.Val)
	}

	return &DateVal{
		node.Name,
		newVal,
	}
}

var _ CST = (*DateVal)(nil)

// Children iterates through all direct children of this node.
func (node StrVal) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node StrVal) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node StrVal) Copy() CST {
	newNode := make(StrVal, len(node))
	copy(newNode, node)
	return newNode
}

var _ CST = (StrVal)(nil)

// Children iterates through all direct children of this node.
func (node NumVal) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node NumVal) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node NumVal) Copy() CST {
	newNode := make(NumVal, len(node))
	copy(newNode, node)
	return newNode
}

var _ CST = (NumVal)(nil)

// Children iterates through all direct children of this node.
func (node ValArg) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node ValArg) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node ValArg) Copy() CST {
	newNode := make(ValArg, len(node))
	copy(newNode, node)
	return newNode
}

var _ CST = (ValArg)(nil)

// Children iterates through all direct children of this node.
func (node KeywordVal) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node KeywordVal) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node KeywordVal) Copy() CST {
	newNode := make(KeywordVal, len(node))
	copy(newNode, node)
	return newNode
}

var _ CST = (KeywordVal)(nil)

// Children iterates through all direct children of this node.
func (node *NullVal) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node *NullVal) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node *NullVal) Copy() CST {
	return &NullVal{}
}

var _ CST = (*NullVal)(nil)

// Children iterates through all direct children of this node.
func (node *TrueVal) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node *TrueVal) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node *TrueVal) Copy() CST {
	return &TrueVal{}
}

var _ CST = (*TrueVal)(nil)

// Children iterates through all direct children of this node.
func (node *FalseVal) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node *FalseVal) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node *FalseVal) Copy() CST {
	return &FalseVal{}
}

var _ CST = (*FalseVal)(nil)

// Children iterates through all direct children of this node.
func (node *UnknownVal) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node *UnknownVal) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node *UnknownVal) Copy() CST {
	return &UnknownVal{}
}

var _ CST = (*UnknownVal)(nil)

// Children iterates through all direct children of this node.
func (node *ColName) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node *ColName) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node *ColName) Copy() CST {
	var newDatabase []byte
	if node.Database == nil {
		newDatabase = nil
	} else {
		newDatabase = make([]byte, len(node.Database))
		copy(newDatabase, node.Database)
	}
	var newName []byte
	if node.Name == nil {
		newName = nil
	} else {
		newName = make([]byte, len(node.Name))
		copy(newName, node.Name)
	}
	var newQualifier []byte
	if node.Qualifier == nil {
		newQualifier = nil
	} else {
		newQualifier = make([]byte, len(node.Qualifier))
		copy(newQualifier, node.Qualifier)
	}

	return &ColName{
		newDatabase,
		newName,
		newQualifier,
	}
}

var _ CST = (*ColName)(nil)

// Children iterates through all direct children of this node.
func (node ValTuple) Children() []CST {
	var result []CST
	for _, child := range node {
		result = append(result, child)
	}
	return result
}

// ReplaceChild changes the value of a particular child node.
func (node ValTuple) ReplaceChild(i int, child CST) {
	if child == nil {
		node[i] = nil
	} else {
		node[i] = child.(Expr)
	}
}

// Copy produces a deep copy of this node.
func (node ValTuple) Copy() CST {
	newNode := make(ValTuple, len(node))
	for i, e := range node {
		if e == nil {
			newNode[i] = nil
		} else {
			newNode[i] = e.Copy().(Expr)
		}
	}
	return newNode
}

var _ CST = (ValTuple)(nil)

// Children iterates through all direct children of this node.
func (node Exprs) Children() []CST {
	var result []CST
	for _, child := range node {
		result = append(result, child)
	}
	return result
}

// ReplaceChild changes the value of a particular child node.
func (node Exprs) ReplaceChild(i int, child CST) {
	if child == nil {
		node[i] = nil
	} else {
		node[i] = child.(Expr)
	}
}

// Copy produces a deep copy of this node.
func (node Exprs) Copy() CST {
	newNode := make(ValTuple, len(node))
	for i, e := range node {
		if e == nil {
			newNode[i] = nil
		} else {
			newNode[i] = e.Copy().(Expr)
		}
	}
	return newNode
}

var _ CST = (Exprs)(nil)

// Children iterates through all direct children of this node.
func (node *Subquery) Children() []CST {
	return []CST{
		node.Select,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *Subquery) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Select = nil
		} else {
			node.Select = child.(SelectStatement)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *Subquery) Copy() CST {
	var selectC SelectStatement
	if node.Select == nil {
		selectC = nil
	} else {
		selectC = node.Select.Copy().(SelectStatement)
	}

	return &Subquery{
		selectC,
		node.IsDerived,
	}
}

var _ CST = (*Subquery)(nil)

// Children iterates through all direct children of this node.
func (node *BinaryExpr) Children() []CST {
	return []CST{
		node.Left,
		node.Right,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *BinaryExpr) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Left = nil
		} else {
			node.Left = child.(Expr)
		}
	case 1:
		if child == nil {
			node.Right = nil
		} else {
			node.Right = child.(Expr)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *BinaryExpr) Copy() CST {
	var left Expr
	if node.Left == nil {
		left = nil
	} else {
		left = node.Left.Copy().(Expr)
	}
	var right Expr
	if node.Right == nil {
		right = nil
	} else {
		right = node.Right.Copy().(Expr)
	}

	return &BinaryExpr{
		node.Operator,
		left,
		right,
	}
}

var _ CST = (*BinaryExpr)(nil)

// Children iterates through all direct children of this node.
func (node *UnaryExpr) Children() []CST {
	return []CST{
		node.Expr,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *UnaryExpr) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Expr = nil
		} else {
			node.Expr = child.(Expr)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *UnaryExpr) Copy() CST {
	var expr Expr
	if node.Expr == nil {
		expr = nil
	} else {
		expr = node.Expr.Copy().(Expr)
	}

	return &UnaryExpr{
		node.Operator,
		expr,
	}
}

var _ CST = (*UnaryExpr)(nil)

// Children iterates through all direct children of this node.
func (node *FuncExpr) Children() []CST {
	return []CST{
		node.Exprs,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *FuncExpr) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Exprs = nil
		} else {
			node.Exprs = child.(SelectExprs)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *FuncExpr) Copy() CST {
	var newName []byte
	if node.Name == nil {
		newName = nil
	} else {
		newName = make([]byte, len(node.Name))
		copy(newName, node.Name)
	}
	var exprs SelectExprs
	if node.Exprs == nil {
		exprs = nil
	} else {
		exprs = node.Exprs.Copy().(SelectExprs)
	}
	var ord OrderBy
	if node.OrderBy == nil {
		ord = nil
	} else {
		ord = node.OrderBy.Copy().(OrderBy)
	}
	var sep []byte
	if node.Separator == nil {
		sep = nil
	} else {
		sep = make([]byte, len(node.Separator))
		copy(sep, node.Separator)
	}

	return &FuncExpr{
		newName,
		node.Distinct,
		exprs,
		ord,
		sep,
	}
}

var _ CST = (*FuncExpr)(nil)

// Children iterates through all direct children of this node.
func (node *CaseExpr) Children() []CST {
	return []CST{
		node.Expr,
		node.Whens,
		node.Else,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *CaseExpr) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Expr = nil
		} else {
			node.Expr = child.(Expr)
		}
	case 1:
		if child == nil {
			node.Whens = nil
		} else {
			node.Whens = child.(Whens)
		}
	case 2:
		if child == nil {
			node.Else = nil
		} else {
			node.Else = child.(Expr)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *CaseExpr) Copy() CST {
	var expr Expr
	if node.Expr == nil {
		expr = nil
	} else {
		expr = node.Expr.Copy().(Expr)
	}
	var whens Whens
	if node.Whens == nil {
		whens = nil
	} else {
		whens = node.Whens.Copy().(Whens)
	}
	var elseC Expr
	if node.Else == nil {
		elseC = nil
	} else {
		elseC = node.Else.Copy().(Expr)
	}

	return &CaseExpr{
		expr,
		whens,
		elseC,
	}
}

var _ CST = (*CaseExpr)(nil)

// Children iterates through all direct children of this node.
func (node Whens) Children() []CST {
	var result []CST
	for _, child := range node {
		result = append(result, child)
	}
	return result
}

// ReplaceChild changes the value of a particular child node.
func (node Whens) ReplaceChild(i int, child CST) {
	if child == nil {
		node[i] = nil
	} else {
		node[i] = child.(*When)
	}
}

// Copy produces a deep copy of this node.
func (node Whens) Copy() CST {
	newNode := make(Whens, len(node))
	for i, w := range node {
		if w == nil {
			newNode[i] = nil
		} else {
			newNode[i] = w.Copy().(*When)
		}
	}
	return newNode
}

var _ CST = (Whens)(nil)

// Children iterates through all direct children of this node.
func (node *When) Children() []CST {
	return []CST{
		node.Cond,
		node.Val,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *When) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Cond = nil
		} else {
			node.Cond = child.(Expr)
		}
	case 1:
		node.Val = child.(Expr)
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *When) Copy() CST {
	var cond Expr
	if node.Cond == nil {
		cond = nil
	} else {
		cond = node.Cond.Copy().(Expr)
	}
	var val Expr
	if node.Val == nil {
		val = nil
	} else {
		val = node.Val.Copy().(Expr)
	}

	return &When{
		cond,
		val,
	}
}

var _ CST = (*When)(nil)

// Children iterates through all direct children of this node.
func (node GroupBy) Children() []CST {
	var result []CST
	for _, child := range node {
		result = append(result, child)
	}
	return result
}

// ReplaceChild changes the value of a particular child node.
func (node GroupBy) ReplaceChild(i int, child CST) {
	if child == nil {
		node[i] = nil
	} else {
		node[i] = child.(Expr)
	}
}

// Copy produces a deep copy of this node.
func (node GroupBy) Copy() CST {
	newNode := make(GroupBy, len(node))
	for i, e := range node {
		if e == nil {
			newNode[i] = nil
		} else {
			newNode[i] = e.Copy().(Expr)
		}
	}
	return newNode
}

var _ CST = (GroupBy)(nil)

// Children iterates through all direct children of this node.
func (node OrderBy) Children() []CST {
	var result []CST
	for _, child := range node {
		result = append(result, child)
	}
	return result
}

// ReplaceChild changes the value of a particular child node.
func (node OrderBy) ReplaceChild(i int, child CST) {
	if child == nil {
		node[i] = nil
	} else {
		node[i] = child.(*Order)
	}
}

// Copy produces a deep copy of this node.
func (node OrderBy) Copy() CST {
	newNode := make(OrderBy, len(node))
	for i, e := range node {
		if e == nil {
			newNode[i] = nil
		} else {
			newNode[i] = e.Copy().(*Order)
		}
	}
	return newNode
}

var _ CST = (OrderBy)(nil)

// Children iterates through all direct children of this node.
func (node *Order) Children() []CST {
	return []CST{
		node.Expr,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *Order) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Expr = nil
		} else {
			node.Expr = child.(Expr)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *Order) Copy() CST {
	var expr Expr
	if node.Expr == nil {
		expr = nil
	} else {
		expr = node.Expr.Copy().(Expr)
	}

	return &Order{
		expr,
		node.Direction,
	}
}

var _ CST = (*Order)(nil)

// Children iterates through all direct children of this node.
func (node *Limit) Children() []CST {
	return []CST{
		node.Offset,
		node.Rowcount,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *Limit) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Offset = nil
		} else {
			node.Offset = child.(Expr)
		}
	case 1:
		if child == nil {
			node.Rowcount = nil
		} else {
			node.Rowcount = child.(Expr)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *Limit) Copy() CST {
	var offset Expr
	if node.Offset == nil {
		offset = nil
	} else {
		offset = node.Offset.Copy().(Expr)
	}
	var rowcount Expr
	if node.Rowcount == nil {
		rowcount = nil
	} else {
		rowcount = node.Rowcount.Copy().(Expr)
	}

	return &Limit{
		offset,
		rowcount,
	}
}

var _ CST = (*Limit)(nil)

// Children iterates through all direct children of this node.
func (node UpdateExprs) Children() []CST {
	var result []CST
	for _, child := range node {
		result = append(result, child)
	}
	return result
}

// ReplaceChild changes the value of a particular child node.
func (node UpdateExprs) ReplaceChild(i int, child CST) {
	if child == nil {
		node[i] = nil
	} else {
		node[i] = child.(*UpdateExpr)
	}
}

// Copy produces a deep copy of this node.
func (node UpdateExprs) Copy() CST {
	newNode := make(UpdateExprs, len(node))
	for i, ue := range node {
		if ue == nil {
			newNode[i] = nil
		} else {
			newNode[i] = ue.Copy().(*UpdateExpr)
		}
	}
	return newNode
}

var _ CST = (UpdateExprs)(nil)

// Children iterates through all direct children of this node.
func (node *UpdateExpr) Children() []CST {
	return []CST{
		node.Name,
		node.Expr,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *UpdateExpr) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Name = nil
		} else {
			node.Name = child.(*ColName)
		}
	case 1:
		if child == nil {
			node.Expr = nil
		} else {
			node.Expr = child.(Expr)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *UpdateExpr) Copy() CST {
	var name *ColName
	if node.Name == nil {
		name = nil
	} else {
		name = node.Name.Copy().(*ColName)
	}
	var expr Expr
	if node.Expr == nil {
		expr = nil
	} else {
		expr = node.Expr.Copy().(Expr)
	}

	return &UpdateExpr{
		name,
		expr,
	}
}

var _ CST = (*UpdateExpr)(nil)

// Children iterates through all direct children of this node.
func (node *SimpleSelect) Children() []CST {
	return []CST{
		node.Comments,
		node.QueryGlobals,
		node.SelectExprs,
		node.Limit,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *SimpleSelect) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Comments = nil
		} else {
			node.Comments = child.(Comments)
		}
	case 1:
		if child == nil {
			node.QueryGlobals = nil
		} else {
			node.QueryGlobals = child.(*QueryGlobals)
		}
	case 2:
		if child == nil {
			node.SelectExprs = nil
		} else {
			node.SelectExprs = child.(SelectExprs)
		}
	case 3:
		if child == nil {
			node.Limit = nil
		} else {
			node.Limit = child.(*Limit)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *SimpleSelect) Copy() CST {
	var comments Comments
	if node.Comments == nil {
		comments = nil
	} else {
		comments = node.Comments.Copy().(Comments)
	}
	var queryGlobals *QueryGlobals
	if node.QueryGlobals == nil {
		queryGlobals = nil
	} else {
		queryGlobals = node.QueryGlobals.Copy().(*QueryGlobals)
	}
	var selectExprs SelectExprs
	if node.SelectExprs == nil {
		selectExprs = nil
	} else {
		selectExprs = node.SelectExprs.Copy().(SelectExprs)
	}
	var limit *Limit
	if node.Limit == nil {
		limit = nil
	} else {
		limit = node.Limit.Copy().(*Limit)
	}

	return &SimpleSelect{
		comments,
		queryGlobals,
		selectExprs,
		limit,
	}
}

var _ CST = (*SimpleSelect)(nil)

// Children iterates through all direct children of this node.
func (node *Show) Children() []CST {
	return []CST{
		node.From,
		node.LikeOrWhere,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *Show) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.From = nil
		} else {
			node.From = child.(Expr)
		}
	case 1:
		if child == nil {
			node.LikeOrWhere = nil
		} else {
			node.LikeOrWhere = child.(Expr)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *Show) Copy() CST {
	var from Expr
	if node.From == nil {
		from = nil
	} else {
		from = node.From.Copy().(Expr)
	}
	var likeOrWhere Expr
	if node.LikeOrWhere == nil {
		likeOrWhere = nil
	} else {
		likeOrWhere = node.LikeOrWhere.Copy().(Expr)
	}

	return &Show{
		node.Section,
		node.Key,
		from,
		likeOrWhere,
		node.Modifier,
	}
}

var _ CST = (*Show)(nil)

// Children iterates through all direct children of this node.
func (node *Explain) Children() []CST {
	return []CST{
		node.Table,
		node.Column,
		node.Statement,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *Explain) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Table = nil
		} else {
			node.Table = child.(*TableName)
		}
	case 1:
		if child == nil {
			node.Column = nil
		} else {
			node.Column = child.(*ColName)
		}
	case 2:
		if child == nil {
			node.Statement = nil
		} else {
			node.Statement = child.(Statement)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *Explain) Copy() CST {
	var table *TableName
	if node.Table == nil {
		table = nil
	} else {
		table = node.Table.Copy().(*TableName)
	}
	var column *ColName
	if node.Column == nil {
		column = nil
	} else {
		column = node.Column.Copy().(*ColName)
	}
	var newConnection []byte
	if node.Connection == nil {
		newConnection = nil
	} else {
		newConnection = make([]byte, len(node.Connection))
		copy(newConnection, node.Connection)
	}
	var statement Statement
	if node.Statement == nil {
		statement = nil
	} else {
		statement = node.Statement.Copy().(Statement)
	}

	return &Explain{
		node.Section,
		table,
		column,
		node.ExplainType,
		newConnection,
		statement,
	}
}

var _ CST = (*Explain)(nil)

// Children iterates through all direct children of this node.
func (node *Kill) Children() []CST {
	return []CST{
		node.ID,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *Kill) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.ID = nil
		} else {
			node.ID = child.(Expr)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *Kill) Copy() CST {
	var id Expr
	if node.ID == nil {
		id = nil
	} else {
		id = node.ID.Copy().(Expr)
	}

	return &Kill{
		node.Scope,
		id,
	}
}

var _ CST = (*Kill)(nil)

// Children iterates through all direct children of this node.
func (node *Flush) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node *Flush) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node *Flush) Copy() CST {
	return &Flush{
		node.Kind,
	}
}

var _ CST = (*Flush)(nil)

// Children iterates through all direct children of this node.
func (node *AlterTable) Children() []CST {
	return []CST{
		node.Table,
		node.Specs,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *AlterTable) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Table = nil
		} else {
			node.Table = child.(*TableName)
		}
	case 1:
		if child == nil {
			node.Specs = nil
		} else {
			node.Specs = child.(AlterSpecs)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *AlterTable) Copy() CST {
	var table *TableName
	if node.Table == nil {
		table = nil
	} else {
		table = node.Table.Copy().(*TableName)
	}
	var specs AlterSpecs
	if node.Specs == nil {
		specs = nil
	} else {
		specs = node.Specs.Copy().(AlterSpecs)
	}

	return &AlterTable{
		table,
		specs,
	}
}

var _ CST = (*AlterTable)(nil)

// Children iterates through all direct children of this node.
func (node *AlterSpec) Children() []CST {
	return []CST{
		node.Column,
		node.NewColumn,
		node.NewTable,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *AlterSpec) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Column = nil
		} else {
			node.Column = child.(*ColName)
		}
	case 1:
		if child == nil {
			node.NewColumn = nil
		} else {
			node.NewColumn = child.(*ColName)
		}
	case 2:
		if child == nil {
			node.NewTable = nil
		} else {
			node.NewTable = child.(*TableName)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *AlterSpec) Copy() CST {
	var column *ColName
	if node.Column == nil {
		column = nil
	} else {
		column = node.Column.Copy().(*ColName)
	}
	var newColumn *ColName
	if node.NewColumn == nil {
		newColumn = nil
	} else {
		newColumn = node.NewColumn.Copy().(*ColName)
	}
	var newTable *TableName
	if node.NewTable == nil {
		newTable = nil
	} else {
		newTable = node.NewTable.Copy().(*TableName)
	}

	return &AlterSpec{
		node.Type,
		column,
		newColumn,
		newTable,
		node.NewColumnType,
	}
}

var _ CST = (*AlterSpec)(nil)

// Children iterates through all direct children of this node.
func (node AlterSpecs) Children() []CST {
	var result []CST
	for _, child := range node {
		result = append(result, child)
	}
	return result
}

// ReplaceChild changes the value of a particular child node.
func (node AlterSpecs) ReplaceChild(i int, child CST) {
	if child == nil {
		node[i] = nil
	} else {
		node[i] = child.(*AlterSpec)
	}
}

// Copy produces a deep copy of this node.
func (node AlterSpecs) Copy() CST {
	newNode := make(AlterSpecs, len(node))
	for i, as := range node {
		if as == nil {
			newNode[i] = nil
		} else {
			newNode[i] = as.Copy().(*AlterSpec)
		}
	}
	return newNode
}

var _ CST = (AlterSpecs)(nil)

// Children iterates through all direct children of this node.
func (node *RenameTable) Children() []CST {
	return []CST{
		node.Renames,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *RenameTable) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Renames = nil
		} else {
			node.Renames = child.(RenameSpecs)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *RenameTable) Copy() CST {
	var renames RenameSpecs
	if node.Renames == nil {
		renames = nil
	} else {
		renames = node.Renames.Copy().(RenameSpecs)
	}

	return &RenameTable{
		renames,
	}
}

var _ CST = (*RenameTable)(nil)

// Children iterates through all direct children of this node.
func (node *RenameSpec) Children() []CST {
	return []CST{
		node.Table,
		node.NewTable,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *RenameSpec) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Table = nil
		} else {
			node.Table = child.(*TableName)
		}
	case 1:
		if child == nil {
			node.NewTable = nil
		} else {
			node.NewTable = child.(*TableName)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *RenameSpec) Copy() CST {
	var table *TableName
	if node.Table == nil {
		table = nil
	} else {
		table = node.Table.Copy().(*TableName)
	}
	var newTable *TableName
	if node.NewTable == nil {
		newTable = nil
	} else {
		newTable = node.NewTable.Copy().(*TableName)
	}

	return &RenameSpec{
		table,
		newTable,
	}
}

var _ CST = (*RenameSpec)(nil)

// Children iterates through all direct children of this node.
func (node RenameSpecs) Children() []CST {
	var result []CST
	for _, child := range node {
		result = append(result, child)
	}
	return result
}

// ReplaceChild changes the value of a particular child node.
func (node RenameSpecs) ReplaceChild(i int, child CST) {
	if child == nil {
		node[i] = nil
	} else {
		node[i] = child.(*RenameSpec)
	}
}

// Copy produces a deep copy of this node.
func (node RenameSpecs) Copy() CST {
	newNode := make(RenameSpecs, len(node))
	for i, rs := range node {
		if rs == nil {
			newNode[i] = nil
		} else {
			newNode[i] = rs.Copy().(*RenameSpec)
		}
	}
	return newNode
}

var _ CST = (RenameSpecs)(nil)
