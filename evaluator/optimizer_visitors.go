package evaluator

// columnGatherer is a visitor that finds all the columns used in a Node.
type columnGatherer struct {
	columns []SQLColumnExpr
}

// visit walks the provided Node, storing any SQLColumnExprs encountered.
func (v *columnGatherer) visit(n Node) (Node, error) {
	n, err := walk(v, n)
	if err != nil {
		return nil, err
	}

	switch typedN := n.(type) {
	case SQLColumnExpr:
		v.columns = append(v.columns, typedN)
	}

	return n, nil
}

// getTableColumnsInExpr finds all columns in the provided SQLExpr that belong
// to the provided MongoSourceStage.
func getTableColumnsInExpr(table *MongoSourceStage, e SQLExpr) ([]SQLColumnExpr, error) {
	v := &columnGatherer{}
	_, err := v.visit(e)
	if err != nil {
		return nil, err
	}

	cols := []SQLColumnExpr{}
	for _, col := range v.columns {
		if containsString(table.aliasNames, col.tableName) {
			cols = append(cols, col)
		}
	}

	return cols, nil
}

type columnFinder struct {
	selectIDsInScope []int
	columns          Columns
	mustBeInScope    bool
}

// referencedColumns takes an expression and returns all unique columns
// referenced in that expression. If mustBeInScope is true, it constrains
// the columns referenced to those that have a select id matching one
// within selectIDsInScope.
func referencedColumns(selectIDsInScope []int, e SQLExpr, mustBeInScope bool) ([]*Column, error) {
	cf := &columnFinder{
		selectIDsInScope: selectIDsInScope,
		mustBeInScope:    mustBeInScope,
	}

	_, err := cf.visit(e)
	if err != nil {
		return nil, err
	}

	return cf.columns.Unique(), nil
}

func (cf *columnFinder) visit(n Node) (Node, error) {
	switch typedN := n.(type) {
	case SQLColumnExpr:
		if !cf.mustBeInScope || containsInt(cf.selectIDsInScope, typedN.selectID) {
			column := NewColumn(typedN.selectID,
				typedN.tableName,
				"",
				typedN.databaseName,
				typedN.columnName,
				"",
				"",
				typedN.columnType.EvalType,
				typedN.columnType.MongoType,
				false,
			)
			cf.columns = append(cf.columns, column)
		}
		return n, nil
	}

	return walk(cf, n)
}

func newColumnTracker() *columnTracker {
	return &columnTracker{
		selectIDs: make(map[int]*sqlColExprCounter),
	}
}

// columnTracker is for scoped handling of column names like a symbol
// table in a compiler. New scopes are introduced by subqueries.
type columnTracker struct {
	selectIDs  map[int]*sqlColExprCounter
	removeMode bool
}

func (t *columnTracker) add(e SQLExpr) {
	t.removeMode = false
	_, err := t.visit(e)
	// This err was previously ignored.
	if err != nil {
		panic(err)
	}
}

func (t *columnTracker) remove(e SQLExpr) {
	t.removeMode = true
	_, err := t.visit(e)
	// This err was previously ignored.
	if err != nil {
		panic(err)
	}
}

// scopedColumnExprsForTables returns the subset of tracked SQLColumnExpr
// values that are within the given select ids and match either the given
// database and table names or match the empty string.
func (t *columnTracker) scopedColumnExprsForTables(selectIDs []int,
	databaseName string, tableNames []string) []SQLColumnExpr {
	var columnExprs []SQLColumnExpr
	for _, selectID := range selectIDs {
		selectIDMap, ok := t.selectIDs[selectID]
		if !ok {
			continue
		}

		for _, expr := range selectIDMap.exprs {
			if expr.databaseName != databaseName &&
				expr.databaseName != "" {
				continue
			}
			if containsString(tableNames, expr.tableName) ||
				expr.tableName == "" {
				columnExprs = append(columnExprs, expr)
			}
		}
	}
	return columnExprs
}

func (t *columnTracker) visit(n Node) (Node, error) {
	switch typedN := n.(type) {
	case SQLColumnExpr:
		selectIDMap, ok := t.selectIDs[typedN.selectID]
		if !ok && !t.removeMode {
			selectIDMap = newSQLColumnExprCounter()
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

func (f *sourceFinder) visit(n Node) (Node, error) {
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

// canOptimizeJoinSubtree returns true if this subtree in n can be optimized.
func canOptimizeJoinSubtree(n Node) bool {
	switch n.(type) {
	case *DynamicSourceStage, *MongoSourceStage, *JoinStage, *SubquerySourceStage, *UnionStage:
		return true
	}
	return false
}

// containsMongoSource returns true if the plan contains a *MongoSourceStage and false otherwise.
func containsMongoSource(n Node) (bool, error) {
	sf := &sourceFinder{}
	_, err := sf.visit(n)
	return sf.source != nil, err
}
