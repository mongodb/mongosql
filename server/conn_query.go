package server

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/deafgoat/mixer/sqlparser"
	"github.com/mongodb/mongo-tools/common/log"
)

func (c *conn) handleQuery(sql string) (err error) {
	log.Logf(log.DebugLow, "[conn%v] %s\n", c.connectionID, sql)

	defer func() {
		if e := recover(); e != nil {
			log.Logf(log.Always, "[conn%v] %s\n", c.connectionID, debug.Stack())
			err = fmt.Errorf("execute %s error %v", sql, e)
			return
		}
	}()

	sql = strings.TrimRight(sql, ";")

	var stmt sqlparser.Statement
	stmt, err = sqlparser.Parse(sql)
	if err != nil {

		// This is an ugly hack such that if someone tries to set some parameter to the default
		// ignore.  This is because the sql parser barfs.  We should probably fix there for reals.
		sqlUpper := strings.ToUpper(sql)
		if len(sqlUpper) > 3 && sqlUpper[0:4] == "SET " {
			if len(sqlUpper) > 7 && sqlUpper[len(sqlUpper)-8:] == "=DEFAULT" {
				// wow, this is ugly
				return c.writeOK(nil)
			}
		}

		return fmt.Errorf(`parse sql "%s" error: %s`, sql, err)
	}

	switch v := stmt.(type) {
	case *sqlparser.Select:
		return c.handleSelect(v, sql, nil)
	case *sqlparser.Set:
		return c.handleSet(v)
	case *sqlparser.SimpleSelect:
		return c.handleSimpleSelect(sql, v)
	case *sqlparser.Show:
		return c.handleShow(sql, v)
	case *sqlparser.DDL:
		return c.handleDDL(v)
	default:
		return fmt.Errorf("statement %T not supported", stmt)
	}
}
