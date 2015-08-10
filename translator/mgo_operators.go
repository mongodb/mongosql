package translator

import (
	"github.com/siddontang/mixer/sqlparser"
)

var (
	MgoEq      = "$eq"
	MgoLt      = "$lt"
	MgoGt      = "$gt"
	MgoLe      = "$lte"
	MgoGe      = "$gte"
	MgoNe      = "$ne"
	MgoNse     = "???" // TODO
	MgoIn      = "$in"
	MgoNotIn   = "$nin"
	MgoLike    = "???" // TODO
	MgoNotLike = "???" // TODO
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
