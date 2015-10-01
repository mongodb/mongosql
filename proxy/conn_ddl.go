package proxy

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"strings"
)

func (c *Conn) handleDDL(ddl *sqlparser.DDL) error {
	switch ddl.Action {
	case "drop":
		tableName := string(ddl.Table)
		if strings.Index(tableName, "#Tableau") == 0 {
			return c.writeOK(nil)
		}
		return fmt.Errorf("cannot drop table (%s)", tableName)
	default:
		return fmt.Errorf("unknown ddl operator (%s)", ddl.Action)
	}
}
