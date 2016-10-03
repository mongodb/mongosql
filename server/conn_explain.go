package server

import (
	"strings"

	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/parser"
)

func (c *conn) handleExplain(sql string, stmt *parser.Explain) error {
	switch strings.ToLower(stmt.Section) {
	case "table":
		return c.handleExplainTable(sql, stmt)
	case "plan":
		return c.handleExplainPlan(sql, stmt)
	default:
		return mysqlerrors.Newf(mysqlerrors.ER_NOT_SUPPORTED_YET, "no support for explain (%s) for now", sql) // unreachable
	}
}

func (c *conn) handleExplainTable(sql string, stmt *parser.Explain) error {
	show := &parser.Show{
		Section: "columns",
		From:    parser.StrVal(stmt.Table.Name),
	}

	if stmt.Column != nil {
		show.LikeOrWhere = parser.StrVal(stmt.Column.Name)
	}

	return c.handleShowColumns(sql, show)
}

func (c *conn) handleExplainPlan(sql string, stmt *parser.Explain) error {
	return mysqlerrors.Newf(mysqlerrors.ER_NOT_SUPPORTED_YET, "no support for explain (%s) for now", sql)
}
