package translator

import "github.com/erh/mixer/sqlparser"

func ParseSQL(sql string) (sqlparser.Statement, error) {
	stmt, err := sqlparser.Parse(sql)
	if err == nil {
		return stmt, err
	}

	return stmt, err
}
