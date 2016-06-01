package server

import (
	"strings"

	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/deafgoat/mixer/sqlparser"
)

func (c *conn) handleDDL(ddl *sqlparser.DDL) error {
	switch ddl.Action {
	case "drop":
		tableName := string(ddl.Table)
		if strings.Index(tableName, "#Tableau") == 0 {
			return c.writeOK(nil)
		}
		return mysqlerrors.Unknownf("cannot drop table (%s)", tableName)
	default:
		return mysqlerrors.Unknownf("unknown ddl operator (%s)", ddl.Action)
	}
}
