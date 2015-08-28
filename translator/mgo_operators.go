package translator

import (
	"github.com/erh/mixer/sqlparser"
)

var (
	MgoEq      = "$eq"
	MgoLt      = "$lt"
	MgoGt      = "$gt"
	MgoLe      = "$lte"
	MgoGe      = "$gte"
	MgoNe      = "$ne"
	MgoNse     = "???" // TODO: NULL_SAFE_EQUAL
	MgoIn      = "$in"
	MgoNotIn   = "$nin"
	MgoLike    = "$regex"      // TODO: LIKE EXPRESSIONS
	MgoNotLike = "$nyi-!regex" // TODO: NOT LIKE EXPRESSIONS
	MgoAnd     = "$and"
	MgoOr      = "$or"

	oprtMap = map[string]string{
		sqlparser.AST_EQ:       MgoEq,
		sqlparser.AST_LT:       MgoLt,
		sqlparser.AST_GT:       MgoGt,
		sqlparser.AST_LE:       MgoLe,
		sqlparser.AST_GE:       MgoGe,
		sqlparser.AST_NE:       MgoNe,
		sqlparser.AST_NSE:      MgoNse,
		sqlparser.AST_IN:       MgoIn,
		sqlparser.AST_NOT_IN:   MgoNotIn,
		sqlparser.AST_LIKE:     MgoLike,
		sqlparser.AST_NOT_LIKE: MgoNotLike,
	}
)
