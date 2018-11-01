package server

import (
	"strings"

	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/parser"
)

func (c *conn) handleDropTable(ddl *parser.DropTable) error {
	tableName := ddl.Name.Name
	if strings.HasPrefix(tableName, "#Tableau") {
		return c.writeOK(nil)
	}
	return mysqlerrors.Unknownf("cannot drop table (%s)", tableName)
}
