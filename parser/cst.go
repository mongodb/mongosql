package parser

// CST represents a node in the CST.
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
	return &Use{
		node.DBName,
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
	if node.TableName != nil {
		tableName = node.TableName.Copy().(*TableName)
	}
	var columnExprs ColumnExprs
	if node.ColumnExprs != nil {
		columnExprs = node.ColumnExprs.Copy().(ColumnExprs)
	}
	var query SelectStatement
	if node.Query != nil {
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
		if cte != nil {
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
	if node.CTEs != nil {
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
	if node.With != nil {
		with = node.With.Copy().(*With)
	}
	var comments Comments
	if node.Comments != nil {
		comments = node.Comments.Copy().(Comments)
	}
	var queryGlobals *QueryGlobals
	if node.QueryGlobals != nil {
		queryGlobals = node.QueryGlobals.Copy().(*QueryGlobals)
	}
	var selectExprs SelectExprs
	if node.SelectExprs != nil {
		selectExprs = node.SelectExprs.Copy().(SelectExprs)
	}
	var from TableExprs
	if node.From != nil {
		from = node.From.Copy().(TableExprs)
	}
	var where *Where
	if node.Where != nil {
		where = node.Where.Copy().(*Where)
	}
	var groupBy GroupBy
	if node.GroupBy != nil {
		groupBy = node.GroupBy.Copy().(GroupBy)
	}
	var having *Where
	if node.Having != nil {
		having = node.Having.Copy().(*Where)
	}
	var orderBy OrderBy
	if node.OrderBy != nil {
		orderBy = node.OrderBy.Copy().(OrderBy)
	}
	var limit *Limit
	if node.Limit != nil {
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
	if node.With != nil {
		with = node.With.Copy().(*With)
	}
	var left SelectStatement
	if node.Left != nil {
		left = node.Left.Copy().(SelectStatement)
	}
	var right SelectStatement
	if node.Right != nil {
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
	if node.Comments != nil {
		comments = node.Comments.Copy().(Comments)
	}
	var exprs UpdateExprs
	if node.Exprs != nil {
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
	if node.Name != nil {
		name = node.Name.Copy().(*TableName)
	}

	return &DropTable{
		name,
		node.IfExists,
		node.Opt.Copy(),
	}
}

var _ CST = (*DropTable)(nil)

// Children iterates through all direct children of this node.
func (node *DropDatabase) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node *DropDatabase) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node *DropDatabase) Copy() CST {
	return &DropDatabase{
		node.Name,
		node.IfExists,
	}
}

var _ CST = (*DropDatabase)(nil)

// Children iterates through all direct children of this node.
func (node *CreateDatabase) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node *CreateDatabase) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node *CreateDatabase) Copy() CST {

	return &CreateDatabase{
		node.Name,
		node.IfNotExists,
	}
}

var _ CST = (*CreateDatabase)(nil)

// Children iterates through all direct children of this node.
func (node *CreateTable) Children() []CST {
	ret := make([]CST, 0, 1+len(node.Definitions)+len(node.TableOptions))
	ret = append(ret, node.Name)
	for _, def := range node.Definitions {
		ret = append(ret, def)
	}
	for _, opt := range node.TableOptions {
		ret = append(ret, opt)
	}
	return ret
}

// ReplaceChild changes the value of a particular child node.
func (node *CreateTable) ReplaceChild(i int, child CST) {
	if i == 0 {
		node.Name = child.(*TableName)
		return
	}
	i--
	if i < len(node.Definitions) {
		node.Definitions[i] = child.(ColumnOrIndexDefinition)
		return
	}
	i -= len(node.Definitions)
	if i < len(node.TableOptions) {
		node.TableOptions[i] = child.(TableOption)
		return
	}
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node *CreateTable) Copy() CST {
	newDefs := make([]ColumnOrIndexDefinition, len(node.Definitions))
	for i, def := range node.Definitions {
		newDefs[i] = def.Copy().(ColumnOrIndexDefinition)
	}
	newOptions := make([]TableOption, len(node.TableOptions))
	for i, opt := range node.TableOptions {
		newOptions[i] = opt.Copy().(TableOption)
	}
	return &CreateTable{
		node.Name.Copy().(*TableName),
		node.IfNotExists,
		newDefs,
		newOptions,
	}
}

var _ CST = (*CreateTable)(nil)

// Children iterates through all direct children of this node.
func (node *ColumnDefinition) Children() []CST {
	return []CST{node.Name}
}

// ReplaceChild changes the value of a particular child node.
func (node *ColumnDefinition) ReplaceChild(i int, child CST) {
	if i == 0 {
		node.Name = child.(*ColName)
		return
	}
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node *ColumnDefinition) Copy() CST {
	return &ColumnDefinition{
		Name:    node.Name.Copy().(*ColName),
		Type:    node.Type,
		Null:    node.Null,
		Unique:  node.Unique,
		Comment: node.Comment,
	}
}

var _ CST = (*ColumnDefinition)(nil)

// Children iterates through all direct children of this node.
func (node *IndexDefinition) Children() []CST {
	ret := make([]CST, len(node.KeyParts))
	for i, part := range node.KeyParts {
		ret[i] = part
	}
	return ret
}

// ReplaceChild changes the value of a particular child node.
func (node *IndexDefinition) ReplaceChild(i int, child CST) {
	if i < len(node.KeyParts) {
		node.KeyParts[i] = child.(KeyPart)
		return
	}
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node *IndexDefinition) Copy() CST {
	clonedKeyParts := make([]KeyPart, len(node.KeyParts))
	for i, part := range node.KeyParts {
		clonedKeyParts[i] = part.Copy().(KeyPart)
	}
	return &IndexDefinition{
		Name:     node.Name,
		Unique:   node.Unique,
		FullText: node.FullText,
		KeyParts: clonedKeyParts,
	}
}

var _ CST = (*IndexDefinition)(nil)

// Children iterates through all direct children of this node.
func (node KeyPart) Children() []CST {
	return []CST{node.Column}
}

// ReplaceChild changes the value of a particular child node.
func (node KeyPart) ReplaceChild(i int, child CST) {
	if i == 0 {
		node.Column = child.(*ColName)
		return
	}
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node KeyPart) Copy() CST {
	return KeyPart{
		Column:    node.Column.Copy().(*ColName),
		Direction: node.Direction,
	}
}

var _ CST = KeyPart{}

// Children iterates through all direct children of this node.
func (node TableComment) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node TableComment) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node TableComment) Copy() CST {
	return node
}

var _ CST = TableComment("")

// Children iterates through all direct children of this node.
func (node IgnoredTableOption) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node IgnoredTableOption) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node IgnoredTableOption) Copy() CST {
	return node
}

var _ CST = IgnoredTableOption{}

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
	copy(newComments, node)
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
		if se != nil {
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
	return &StarExpr{
		node.DatabaseName.Copy(),
		node.TableName.Copy(),
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
	if node.Expr != nil {
		expr = node.Expr.Copy().(Expr)
	}

	return &NonStarExpr{
		expr,
		node.As.Copy(),
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
		if se != nil {
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
		if cn != nil {
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
		if te != nil {
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
	if node.Expr != nil {
		expr = node.Expr.Copy().(SimpleTableExpr)
	}
	var hints *IndexHints
	if node.Hints != nil {
		hints = node.Hints.Copy().(*IndexHints)
	}

	return &AliasedTableExpr{
		expr,
		node.As.Copy(),
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
	return &TableName{
		node.Qualifier.Copy(),
		node.Name,
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
	if node.Expr != nil {
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
	if node.LeftExpr != nil {
		leftExpr = node.LeftExpr.Copy().(TableExpr)
	}
	var rightExpr TableExpr
	if node.RightExpr != nil {
		rightExpr = node.RightExpr.Copy().(TableExpr)
	}
	var on Expr
	if node.On != nil {
		on = node.On.Copy().(Expr)
	}
	var using ColumnExprs
	if node.Using != nil {
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
	newIndexes := make([]string, len(node.Indexes))
	copy(newIndexes, node.Indexes)
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
	if node.Expr != nil {
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
	if node.Left != nil {
		left = node.Left.Copy().(Expr)
	}
	var right Expr
	if node.Right != nil {
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
	if node.Left != nil {
		left = node.Left.Copy().(Expr)
	}
	var right Expr
	if node.Right != nil {
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
	if node.Left != nil {
		left = node.Left.Copy().(Expr)
	}
	var right Expr
	if node.Right != nil {
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
	if node.Expr != nil {
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
	if node.Left != nil {
		left = node.Left.Copy().(Expr)
	}
	var right Expr
	if node.Right != nil {
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
	if node.Left != nil {
		left = node.Left.Copy().(Expr)
	}
	var right Expr
	if node.Right != nil {
		right = node.Right.Copy().(Expr)
	}
	var escape Expr
	if node.Escape != nil {
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
	if node.Left != nil {
		left = node.Left.Copy().(Expr)
	}
	var from Expr
	if node.From != nil {
		from = node.From.Copy().(Expr)
	}
	var to Expr
	if node.To != nil {
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
	if node.Operand != nil {
		operand = node.Operand.Copy().(Expr)
	}
	var pattern Expr
	if node.Pattern != nil {
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
	if node.Operand != nil {
		operand = node.Operand.Copy().(Expr)
	}
	var pattern Expr
	if node.Pattern != nil {
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
	if node.Subquery != nil {
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
	return &DateVal{
		node.Name,
		node.Val,
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
	return node
}

var _ CST = (StrVal)("")

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
	return node
}

var _ CST = (NumVal)("")

// Children iterates through all direct children of this node.
func (node Default) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node Default) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node Default) Copy() CST {
	return node
}

var _ CST = Default{}

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
	return node
}

var _ CST = (ValArg)("")

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
	return node
}

var _ CST = (KeywordVal)("")

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
	return &ColName{
		node.Database.Copy(),
		node.Qualifier.Copy(),
		node.Name,
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
		if e != nil {
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
		if e != nil {
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
	if node.Select != nil {
		selectC = node.Select.Copy().(SelectStatement)
	}

	return &Subquery{
		selectC,
		node.IsDerived,
	}
}

var _ CST = (*Subquery)(nil)

// Children iterates through all direct children of this node.
func (node *DualTableExpr) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node *DualTableExpr) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds; DualTableExpr has no children")
}

// Copy produces a deep copy of this node.
func (node *DualTableExpr) Copy() CST {
	return &DualTableExpr{}
}

var _ CST = (*DualTableExpr)(nil)

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
	if node.Left != nil {
		left = node.Left.Copy().(Expr)
	}
	var right Expr
	if node.Right != nil {
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
	if node.Expr != nil {
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
	var exprs SelectExprs
	if node.Exprs != nil {
		exprs = node.Exprs.Copy().(SelectExprs)
	}
	var ord OrderBy
	if node.OrderBy != nil {
		ord = node.OrderBy.Copy().(OrderBy)
	}

	return &FuncExpr{
		node.Name,
		node.Distinct,
		exprs,
		ord,
		node.Separator.Copy(),
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
	if node.Expr != nil {
		expr = node.Expr.Copy().(Expr)
	}
	var whens Whens
	if node.Whens != nil {
		whens = node.Whens.Copy().(Whens)
	}
	var elseC Expr
	if node.Else != nil {
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
		if w != nil {
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
	if node.Cond != nil {
		cond = node.Cond.Copy().(Expr)
	}
	var val Expr
	if node.Val != nil {
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
		if e != nil {
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
		if e != nil {
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
	if node.Expr != nil {
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
	if node.Offset != nil {
		offset = node.Offset.Copy().(Expr)
	}
	var rowcount Expr
	if node.Rowcount != nil {
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
		if ue != nil {
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
	if node.Name != nil {
		name = node.Name.Copy().(*ColName)
	}
	var expr Expr
	if node.Expr != nil {
		expr = node.Expr.Copy().(Expr)
	}

	return &UpdateExpr{
		name,
		expr,
	}
}

var _ CST = (*UpdateExpr)(nil)

// Children iterates through all direct children of this node.
func (node *Show) Children() []CST {
	return []CST{
		node.From,
		node.Predicate,
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
			node.Predicate = nil
		} else {
			node.Predicate = child.(*ShowPredicate)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *Show) Copy() CST {
	var from Expr
	if node.From != nil {
		from = node.From.Copy().(Expr)
	}
	var predicate *ShowPredicate
	if node.Predicate != nil {
		predicate = node.Predicate.Copy().(*ShowPredicate)
	}

	return &Show{
		node.Section,
		node.Key,
		from,
		predicate,
		node.Modifier,
	}
}

var _ CST = (*Show)(nil)

// Children iterates through all direct children of this node.
func (node *ShowPredicate) Children() []CST {
	return []CST{
		node.Where,
	}
}

// ReplaceChild changes the value of a particular child node.
func (node *ShowPredicate) ReplaceChild(i int, child CST) {
	switch i {
	case 0:
		if child == nil {
			node.Where = nil
		} else {
			node.Where = child.(Expr)
		}
	default:
		panic("ReplaceChild out of bounds")
	}
}

// Copy produces a deep copy of this node.
func (node *ShowPredicate) Copy() CST {
	like := node.Like
	var where Expr
	if node.Where != nil {
		where = node.Where.Copy().(Expr)
	}

	return &ShowPredicate{
		like,
		where,
	}
}

var _ CST = (*ShowPredicate)(nil)

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
	if node.Table != nil {
		table = node.Table.Copy().(*TableName)
	}
	var column *ColName
	if node.Column != nil {
		column = node.Column.Copy().(*ColName)
	}
	var statement Statement
	if node.Statement != nil {
		statement = node.Statement.Copy().(Statement)
	}

	return &Explain{
		node.Section,
		table,
		column,
		node.ExplainType,
		node.Connection.Copy(),
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
	if node.ID != nil {
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
	if node.Table != nil {
		table = node.Table.Copy().(*TableName)
	}
	var specs AlterSpecs
	if node.Specs != nil {
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
	if node.Column != nil {
		column = node.Column.Copy().(*ColName)
	}
	var newColumn *ColName
	if node.NewColumn != nil {
		newColumn = node.NewColumn.Copy().(*ColName)
	}
	var newTable *TableName
	if node.NewTable != nil {
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
		if as != nil {
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
	if node.Renames != nil {
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
	if node.Table != nil {
		table = node.Table.Copy().(*TableName)
	}
	var newTable *TableName
	if node.NewTable != nil {
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
		if rs != nil {
			newNode[i] = rs.Copy().(*RenameSpec)
		}
	}
	return newNode
}

var _ CST = (RenameSpecs)(nil)

// Children iterates through all direct children of this node.
func (node *IgnoredStatement) Children() []CST {
	return []CST{node.Statement}
}

// ReplaceChild changes the value of a particular child node.
func (node *IgnoredStatement) ReplaceChild(i int, child CST) {
	if i != 0 {
		panic("ReplaceChild out of bounds")
	}
	node.Statement = child.(IgnorableStatement)
}

// Copy produces a deep copy of this node.
func (node *IgnoredStatement) Copy() CST {
	return &IgnoredStatement{
		Statement: node.Statement.Copy().(IgnorableStatement),
	}
}

var _ CST = (*IgnoredStatement)(nil)

// Children iterates through all direct children of this node.
func (node LockTables) Children() []CST {
	retList := make([]CST, len(node.LockList))
	for i := range node.LockList {
		retList[i] = node.LockList[i]
	}
	return retList
}

// ReplaceChild changes the value of a particular child node.
func (node LockTables) ReplaceChild(i int, child CST) {
	if i > len(node.LockList) {
		panic("ReplaceChild out of bounds")
	}
	node.LockList[i] = child.(TableLock)
}

// Copy produces a deep copy of this node.
func (node LockTables) Copy() CST {
	retList := make([]TableLock, len(node.LockList))
	for i := range node.LockList {
		retList[i] = node.LockList[i].Copy().(TableLock)
	}
	return LockTables{
		LockList: retList,
	}
}

var _ CST = LockTables{}

// Children iterates through all direct children of this node.
func (node TableLock) Children() []CST {
	return []CST{node.TableName}
}

// ReplaceChild changes the value of a particular child node.
func (node TableLock) ReplaceChild(i int, child CST) {
	if i != 0 {
		panic("ReplaceChild out of bounds")
	}
	node.TableName = child.(*TableName)
}

// Copy produces a deep copy of this node.
func (node TableLock) Copy() CST {
	return TableLock{
		TableName: node.TableName.Copy().(*TableName),
		Alias:     node.Alias,
		LockType:  node.LockType,
	}
}

var _ CST = TableLock{}

// Children iterates through all direct children of this node.
func (node UnlockTables) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node UnlockTables) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node UnlockTables) Copy() CST {
	return node
}

var _ CST = UnlockTables{}

// Children iterates through all direct children of this node.
func (node EnableKeys) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node EnableKeys) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node EnableKeys) Copy() CST {
	return node
}

var _ CST = EnableKeys{}

// Children iterates through all direct children of this node.
func (node DisableKeys) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node DisableKeys) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node DisableKeys) Copy() CST {
	return node
}

var _ CST = DisableKeys{}

// Children iterates through all direct children of this node.
func (node *Insert) Children() []CST {
	ret := make([]CST, len(node.Columns)+2)
	ret[0] = node.Table
	i := 1
	for _, col := range node.Columns {
		ret[i] = col
		i++
	}
	ret[i] = node.Values
	return ret
}

// ReplaceChild changes the value of a particular child node.
func (node *Insert) ReplaceChild(i int, child CST) {
	if i > len(node.Columns)+1 {
		panic("ReplaceChild out of bounds")
	}
	switch i {
	case 0:
		node.Table = child.(*TableName)
	case len(node.Columns) + 1:
		node.Values = child.(ValueListList)
	default:
		node.Columns[i-1] = child.(*ColName)
	}
}

// Copy produces a deep copy of this node.
func (node *Insert) Copy() CST {
	cols := make([]*ColName, len(node.Columns))
	for i := range node.Columns {
		cols[i] = node.Columns[i].Copy().(*ColName)
	}

	return &Insert{
		Table:   node.Table.Copy().(*TableName),
		Columns: cols,
		Values:  node.Values.Copy().(ValueListList),
	}
}

var _ CST = (*Insert)(nil)

// Children iterates through all direct children of this node.
func (node ValueListList) Children() []CST {
	ret := make([]CST, len(node))
	for i := range node {
		ret[i] = node[i]
	}
	return ret
}

// ReplaceChild changes the value of a particular child node.
func (node ValueListList) ReplaceChild(i int, child CST) {
	node[i] = child.(ValueList)
}

// Copy produces a deep copy of this node.
func (node ValueListList) Copy() CST {
	ret := make(ValueListList, len(node))
	for i := range node {
		ret[i] = node[i].Copy().(ValueList)
	}
	return ret
}

var _ CST = ValueListList{}

// Children iterates through all direct children of this node.
func (node ValueList) Children() []CST {
	ret := make([]CST, len(node))
	for i := range node {
		ret[i] = node[i]
	}
	return ret
}

// ReplaceChild changes the value of a particular child node.
func (node ValueList) ReplaceChild(i int, child CST) {
	node[i] = child.(Value)
}

// Copy produces a deep copy of this node.
func (node ValueList) Copy() CST {
	ret := make(ValueList, len(node))
	for i := range node {
		ret[i] = node[i].Copy().(Value)
	}
	return ret
}

var _ CST = ValueList{}

// Children iterates through all direct children of this node.
func (node UnexecutableComment) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node UnexecutableComment) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node UnexecutableComment) Copy() CST {
	return node
}

var _ CST = UnexecutableComment("")

// Children iterates through all direct children of this node.
func (node *ConditionallyExecutableComment) Children() []CST {
	return []CST{}
}

// ReplaceChild changes the value of a particular child node.
func (node *ConditionallyExecutableComment) ReplaceChild(i int, child CST) {
	panic("ReplaceChild out of bounds")
}

// Copy produces a deep copy of this node.
func (node *ConditionallyExecutableComment) Copy() CST {

	return &ConditionallyExecutableComment{
		VersionCode: node.VersionCode,
		SQL:         node.SQL,
	}
}

var _ CST = (*ConditionallyExecutableComment)(nil)
