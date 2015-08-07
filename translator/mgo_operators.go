package translator

import (
	"github.com/siddontang/mixer/sqlparser"
)

var (
	// ComparisonExpr operators
	MGO_OPERATORS = map[string]string{
		sqlparser.AST_EQ:       "$eq",
		sqlparser.AST_LT:       "$lt",
		sqlparser.AST_GT:       "$gt",
		sqlparser.AST_LE:       "$lte",
		sqlparser.AST_GE:       "$gte",
		sqlparser.AST_NE:       "$ne",
		sqlparser.AST_NSE:      "???", // TODO
		sqlparser.AST_IN:       "$in",
		sqlparser.AST_NOT_IN:   "$nin",
		sqlparser.AST_LIKE:     "???", // TODO
		sqlparser.AST_NOT_LIKE: "???", // TODO
	}
)
