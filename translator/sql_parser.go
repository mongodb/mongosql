package translator

import "strings"

import "github.com/erh/mixer/sqlparser"

func ParseSQL(sql string) (sqlparser.Statement, error) {
	stmt, err := sqlparser.Parse(sql)
	if err == nil {
		return stmt, err
	}

	// So, this is ugly.
	// The parser currently treats TABLE and TABLES as tokens
	// So can't parse the query below.
	// We should fix for real in parser
	// TODO HACK AHHH
	sqlUpper := strings.ToUpper(sql)
	idx := strings.Index(sqlUpper, "INFORMATION_SCHEMA.TABLES")
	if idx >= 0 {
		temp := sql[0:idx] + "information_schema.txxxables" + sql[idx+25:]
		stmt, err2 := sqlparser.Parse(temp)
		if err2 == nil {
			return stmt, nil
		}
	}

	return stmt, err
}
