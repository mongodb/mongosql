package server

import (
	"context"
	"strings"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/parser"
)

func (c *conn) handleExplain(ctx context.Context, sql string, stmt *parser.Explain) error {
	switch strings.ToLower(stmt.Section) {
	case "table":
		return c.handleExplainTable(ctx, sql, stmt)
	case "plan":
		return c.handleExplainPlan(ctx, sql, stmt)
	default:
		return mysqlerrors.Newf(mysqlerrors.ErNotSupportedYet, "no support for explain (%s) "+
			"for now", sql) // unreachable
	}
}

func (c *conn) handleExplainTable(ctx context.Context, sql string, stmt *parser.Explain) error {
	show := &parser.Show{
		Section: "columns",
		From:    parser.StrVal(stmt.Table.Name),
	}

	if stmt.Column != nil {
		show.LikeOrWhere = parser.StrVal(stmt.Column.Name)
	}

	return c.handleShow(ctx, sql, show)
}

func (c *conn) handleExplainPlan(ctx context.Context, sql string, stmt *parser.Explain) error {
	rCfg := c.getRewriterConfig()
	aCfg := c.getAlgebrizerConfig(sql, stmt.Statement)
	oCfg := c.getOptimizerConfig()
	pCfg := c.getPushdownConfig()
	eCfg := c.getExecutionConfig()

	res, err := evaluator.EvaluateExplain(ctx, rCfg, aCfg, oCfg, pCfg, eCfg)
	if err != nil {
		return err
	}

	return c.streamResultset(ctx, res.Columns, res.Iter)
}
