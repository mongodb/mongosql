package server

import (
	"strings"

	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/parser"
	"github.com/mongodb/mongo-tools/common/log"
)

var nstring = parser.String

func (c *conn) handleSet(stmt *parser.Set) error {
	if len(stmt.Exprs) != 1 {
		return mysqlerrors.Unknownf("must set one item once, not %s", nstring(stmt))
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
		log.Logf(log.Always, "%s", parser.String(stmt))
		return mysqlerrors.Defaultf(mysqlerrors.ER_UNKNOWN_SYSTEM_VARIABLE, k)
	}
}

func (c *conn) handleSetAutoCommit(val parser.ValExpr) error {
	value, ok := val.(parser.NumVal)
	if !ok {
		return mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_TYPE_FOR_VAR, "autocommit")
	}

	switch value[0] {
	case '1':
		c.status |= SERVER_STATUS_AUTOCOMMIT
	case '0':
		c.status &= ^SERVER_STATUS_AUTOCOMMIT
	default:
		return mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_VALUE_FOR_VAR, "autocommit", value)
	}

	return c.writeOK(nil)
}

func (c *conn) handleSetNames(val parser.ValExpr) error {
	value, ok := val.(parser.StrVal)
	if !ok {
		return mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_TYPE_FOR_VAR, "names")
	}

	charset := strings.ToLower(string(value))
	cid, ok := charsetIds[charset]
	if !ok {
		return mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_VALUE_FOR_VAR, "names", charset)
	}

	c.charset = charset
	c.collation = cid

	return c.writeOK(nil)
}

func (c *conn) handleSetCharacterResults(val parser.ValExpr) error {
	switch expr := val.(type) {
	case *parser.NullVal:
		return c.writeOK(nil)
	default:
		return mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_VALUE_FOR_VAR, "character_set_results", parser.String(expr))
	}
}
