package proxy

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"github.com/mongodb/mongo-tools/common/log"
	. "github.com/siddontang/mixer/mysql"
	"strings"
)

var nstring = sqlparser.String

func (c *Conn) handleSet(stmt *sqlparser.Set) error {
	if len(stmt.Exprs) != 1 {
		return fmt.Errorf("must set one item once, not %s", nstring(stmt))
	}

	k := string(stmt.Exprs[0].Name.Name)

	switch strings.ToLower(k) {
	case `autocommit`:
		return c.handleSetAutoCommit(stmt.Exprs[0].Expr)
	case `names`:
		return c.handleSetNames(stmt.Exprs[0].Expr)
	case `character_set_results`:
		return c.handleSetCharacterResults(stmt.Exprs[0].Expr)
	case `sql_auto_is_null`:
		return c.writeOK(nil)
	case `@@sql_select_limit`:
		// TODO
		return c.writeOK(nil)
	default:
		log.Logf(log.Always, "%s", sqlparser.String(stmt))
		return fmt.Errorf("set (%s) is not supported now", k)
	}
}

func (c *Conn) handleSetAutoCommit(val sqlparser.ValExpr) error {
	value, ok := val.(sqlparser.NumVal)
	if !ok {
		return fmt.Errorf("set autocommit error")
	}
	switch value[0] {
	case '1':
		c.status |= SERVER_STATUS_AUTOCOMMIT
	case '0':
		c.status &= ^SERVER_STATUS_AUTOCOMMIT
	default:
		return fmt.Errorf("invalid autocommit flag %s", value)
	}

	return c.writeOK(nil)
}

func (c *Conn) handleSetNames(val sqlparser.ValExpr) error {
	value, ok := val.(sqlparser.StrVal)
	if !ok {
		return fmt.Errorf("set names charset error")
	}

	charset := strings.ToLower(string(value))
	cid, ok := CharsetIds[charset]
	if !ok {
		return fmt.Errorf("invalid charset %s", charset)
	}

	c.charset = charset
	c.collation = cid

	return c.writeOK(nil)
}

func (c *Conn) handleSetCharacterResults(val sqlparser.ValExpr) error {
	switch expr := val.(type) {
	case *sqlparser.NullVal:
		return c.writeOK(nil)
	default:
		return fmt.Errorf("do not know how to set CHARACTER_SET_RESULTS to: %T %s", expr, sqlparser.String(expr))
	}
}
