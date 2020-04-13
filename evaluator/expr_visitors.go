package evaluator

import (
	"fmt"
	"sort"

	"github.com/10gen/sqlproxy/internal/strutil"
)

// joinOnExpression holds the SQLColumnExpr value
// used within a join stage matcher.
type joinOnExpression struct {
	exprCollector *sqlColExprCollector
}

func (v *joinOnExpression) visit(n Node) (Node, error) {
	switch typedN := n.(type) {
	case *JoinStage:
		_, err := v.exprCollector.visit(typedN.matcher)
		if err != nil {
			return nil, err
		}
		return typedN, nil
	default:
		return walk(v, n)
	}
}

// sqlColExprCounter is a used to hold and count all
// SQLColumnExpr values found during a Node visit. It
// also stores the selectID for the source query that
// this counter is used for.
type sqlColExprCounter struct {
	counts          map[string]int
	exprs           []SQLColumnExpr
	isCaseSensitive bool
	selectID        int
}

// newSQLColumnExprCounter returns a new sqlColExprCounter.
func newSQLColumnExprCounter(selectID int, isCaseSensitive bool) *sqlColExprCounter {
	return &sqlColExprCounter{
		counts:          make(map[string]int),
		isCaseSensitive: isCaseSensitive,
		selectID:        selectID,
	}
}

func (c *sqlColExprCounter) add(e SQLColumnExpr) {
	// If the column has this counter's selectID, it should not be marked
	// as correlated because its source is in the query that this counter
	// is used for. If it does not have that selectID, it must be marked
	// as correlated because its source must come from some outer query.
	e.correlated = e.correlated && e.selectID != c.selectID

	s := e.String()
	if _, ok := c.counts[s]; ok {
		c.counts[s]++
	} else {
		c.counts[s] = 1
		c.exprs = append(c.exprs, e)
	}
}

func (c *sqlColExprCounter) copyExprs() []SQLColumnExpr {
	exprs := make([]SQLColumnExpr, len(c.exprs))
	copy(exprs, c.exprs)
	return exprs
}

func (c *sqlColExprCounter) remove(e SQLColumnExpr) {
	s := e.String()
	for i, expr := range c.exprs {
		if strutil.CompareStrings(s, expr.String(), c.isCaseSensitive) {
			c.counts[s]--
			if c.counts[s] == 0 {
				delete(c.counts, s)
				c.exprs = append(c.exprs[:i], c.exprs[i+1:]...)
			}
			return
		}
	}
}

// sqlColExprCollector is used to track which SQLColumnExpr values
// have been visited and the select context in which they were found.
type sqlColExprCollector struct {
	selectIDs         []int
	referencedColumns *sqlColExprCounter
	removeMode        bool
}

// newSQLColExprCollector returns a new sqlColExprCollector.
func newSQLColExprCollector(selectIDs []int, isCaseSensitive bool) *sqlColExprCollector {
	// Get the selectID for the most deeply nested subquery. It is the
	// last in sorted order (aka, the largest). This is needed for the
	// sqlColExprCounter to keep track of correlation.
	sort.Ints(selectIDs)
	last := -1
	if len(selectIDs) > 0 {
		last = selectIDs[len(selectIDs)-1]
	}

	return &sqlColExprCollector{
		selectIDs:         selectIDs,
		referencedColumns: newSQLColumnExprCounter(last, isCaseSensitive),
	}
}

// Remove visits and removes the SQLColumnExpr
// values within e from the expression collector.
func (c *sqlColExprCollector) Remove(e SQLExpr) {
	c.removeMode = true
	c.Add(e)
	c.removeMode = false
}

// RemoveAll visits and removes the SQLColumnExpr values
// within each element of the slice, e, from the
// expression collector.
func (c *sqlColExprCollector) RemoveAll(e []SQLExpr) {
	c.removeMode = true
	c.AddAll(e)
	c.removeMode = false
}

// AddAll visits and adds the SQLColumnExpr values
// within each element of the slice, e, to the
// expression collector.
func (c *sqlColExprCollector) AddAll(exprs []SQLExpr) {
	for _, e := range exprs {
		c.Add(e)
	}
}

// Add visits and adds the SQLColumnExpr values
// within e to the expression collector.
func (c *sqlColExprCollector) Add(e SQLExpr) {
	_, err := c.visit(e)
	// This err was previously ignored.
	if err != nil {
		panic(err)
	}
}

func (c *sqlColExprCollector) visit(n Node) (Node, error) {
	switch typedN := n.(type) {
	case SQLColumnExpr:
		if containsInt(c.selectIDs, typedN.selectID) {
			if c.removeMode {
				c.referencedColumns.remove(typedN)
			} else {
				c.referencedColumns.add(typedN)
			}
		}
		return typedN, nil
	case *SQLSubqueryExpr:
		if typedN.correlated {
			return walk(c, n)
		}
		return n, nil
	case SQLDoubleSubqueryExpr:
		if typedN.LeftCorrelated() {
			leftPlan, err := walk(c, typedN.LeftPlan())
			if err != nil {
				return nil, err
			}
			typedN.SetLeftPlan(leftPlan.(PlanStage))
		}
		if typedN.RightCorrelated() {
			rightPlan, err := walk(c, typedN.RightPlan())
			if err != nil {
				return nil, err
			}
			typedN.SetRightPlan(rightPlan.(PlanStage))
		}
		return typedN.(Node), nil
	default:
		return walk(c, n)
	}
}

// getJoinOnExpressions gets all column expressions referenced in
// a join 'on' clause in the given plan stage.
func (c *sqlColExprCollector) getJoinOnExpressions(ps PlanStage) {
	v := &joinOnExpression{c}
	_, err := v.visit(ps)
	if err != nil {
		panic(err) // This error should always be nil.
	}
}

type selectIDGatherer struct {
	selectIDs []int
}

func gatherSelectIDs(n Node) []int {
	v := &selectIDGatherer{}
	_, err := v.visit(n)
	if err != nil {
		panic(fmt.Errorf("selectIDGatherer returned unexpected error: %v", err))
	}
	return v.selectIDs
}

func (v *selectIDGatherer) visit(n Node) (Node, error) {
	n, err := walk(v, n)
	if err != nil {
		return nil, err
	}

	switch typedN := n.(type) {
	case SQLColumnExpr:
		v.selectIDs = append(v.selectIDs, typedN.selectID)
	}

	return n, nil
}

type databaseNameFinder struct {
	databaseName         string
	hasMultipleDatabases bool
}

// getDatabaseName returns the name of the database from where the SQLColumnExpr values in n are
// derived. It returns the empty string if the values come from more than one database or the
// dual database.
func getDatabaseName(n Node) string {
	v := &databaseNameFinder{}
	_, err := v.visit(n)
	if err != nil {
		panic(fmt.Errorf("databaseNameFinder returned unexpected error: %v", err))
	}
	if v.hasMultipleDatabases {
		return ""
	}
	return v.databaseName
}

func (v *databaseNameFinder) visit(n Node) (Node, error) {
	if v.hasMultipleDatabases {
		return n, nil
	}

	n, err := walk(v, n)
	if err != nil {
		return nil, err
	}

	switch typedN := n.(type) {
	case SQLColumnExpr:
		if v.databaseName == "" {
			v.databaseName = typedN.databaseName
		} else if typedN.databaseName != v.databaseName {
			v.databaseName = ""
			v.hasMultipleDatabases = true
		}
	}
	return n, nil
}
