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
}

// referencedColumns will take an expression and return all the columns
// referenced in the expression.
func referencedColumns(selectIDsInScope []int, e SQLExpr) ([]*Column, error) {

	cf := &columnFinder{selectIDsInScope: selectIDsInScope}

	_, err := cf.visit(e)
	if err != nil {
		return nil, err
	}

	return cf.columns.Unique(), nil
}

func (cf *columnFinder) visit(n Node) (Node, error) {
	switch typedN := n.(type) {
	case SQLColumnExpr:
		if containsInt(cf.selectIDsInScope, typedN.selectID) {
			column := NewColumn(typedN.selectID, typedN.tableName, "", typedN.databaseName, typedN.columnName, "", "",
				typedN.columnType.SQLType, typedN.columnType.MongoType, false)
			cf.columns = append(cf.columns, column)
		}
		return n, nil
	}

	return walk(cf, n)
}

func newColumnTracker() *columnTracker {
	return &columnTracker{
		selectIDs: make(map[int]*exprCountMap),
	}
}

// columnTracker is for scoped handling of column names like a symbol table in
// a compiler. New scopes are introduced by sub-queries.
type columnTracker struct {
	selectIDs  map[int]*exprCountMap
	removeMode bool
}

func (t *columnTracker) getColumnsForSelectIDs(selectIDs []int) []SQLExpr {
	var columns []SQLExpr
	for _, selectID := range selectIDs {
		if selectIDMap, ok := t.selectIDs[selectID]; ok {
			columns = append(columns, selectIDMap.exprs...)
		}
	}

	return columns
}

func (t *columnTracker) add(e SQLExpr) {
	t.removeMode = false
	t.visit(e)
}

func (t *columnTracker) remove(e SQLExpr) {
	t.removeMode = true
	t.visit(e)
}

func (t *columnTracker) visit(n Node) (Node, error) {
	switch typedN := n.(type) {
	case SQLColumnExpr:
		selectIDMap, ok := t.selectIDs[typedN.selectID]
		if !ok && !t.removeMode {
			selectIDMap = newExprCountMap()
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
