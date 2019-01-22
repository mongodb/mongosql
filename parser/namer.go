package parser

import (
	"fmt"
	"strings"
)

// namer walks the CST and gives the SELECT statements column names to match
// MySQL's auto-aliasing behavior.
//
// Note: We should get the correct column names from the parser, instead of
// generating them by decompiling the parse tree. This is the best option
// available to us in the meantime, as the parser changes needed to do this are
// very invasive (see BI-1897).
type namer struct{}

// PreVisit is called for every node before its children are walked.
func (*namer) PreVisit(current CST) (CST, error) {
	switch node := current.(type) {
	case *Select:
		for _, expr := range node.SelectExprs {
			if column, yes := expr.(*NonStarExpr); yes && column.As.IsNone() {
				if colRef, isColRef := column.Expr.(*ColName); isColRef {
					// Lone column references are named without their
					// qualifiers in MySQL.
					column.As.Set(colRef.Name)
				} else {
					column.As.Set(String(column))
				}
			}
		}
	case *SimpleSelect:
		for _, expr := range node.SelectExprs {
			if column, yes := expr.(*NonStarExpr); yes && column.As.IsNone() {
				if colRef, isColRef := column.Expr.(*ColName); isColRef {
					// Lone column references are named without their qualifiers
					// in MySQL. However, for global variables, our parser puts
					// the "@@global" into the qualifier, so we need to check
					// there in order to properly name variable columns.

					tableName := colRef.Qualifier.Else("")
					colName := colRef.Name
					isVariable := strings.HasPrefix(tableName, "@") || (tableName == "" && strings.HasPrefix(colName, "@"))

					if isVariable && tableName != "" {
						column.As.Set(fmt.Sprintf("%s.%s", tableName, colName))
					} else {
						column.As.Set(colRef.Name)
					}
				} else {
					column.As.Set(String(column))
				}
			}
		}
	}
	return current, nil
}

var _ Walker = (*namer)(nil)

// PostVisit is called for every node after its children are walked.
func (*namer) PostVisit(current CST) (CST, error) {
	return current, nil
}

// NameColumns is a stopgap measure that prints column names by reconstructing
// what should have been saved by the parser.
// These column names fill out empty alias slots in the CST.
func NameColumns(statement Statement) Statement {
	// This call ignores errors because it cannot fail.
	result, _ := Walk(&namer{}, statement)
	return result.(Statement)
}
