package server

import (
	"fmt"
	"sort"
	"strings"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/schema"
	"github.com/deafgoat/mixer/sqlparser"
)

func (c *conn) handleShow(sql string, stmt *sqlparser.Show) error {
	var err error
	var r *Resultset
	switch strings.ToLower(stmt.Section) {
	case "databases":
		r, err = c.handleShowDatabases()
	case "tables":
		r, err = c.handleShowTables(sql, stmt)
	case "variables":
		r, err = c.handleShowVariables(sql, stmt)
	case "columns":
		r, err = c.handleShowColumns(sql, stmt)
	default:
		err = fmt.Errorf("no support for show (%s) for now", sql)
	}

	if err != nil {
		return err
	}

	return c.writeResultset(c.status, r)
}

func (c *conn) handleShowDatabases() (*Resultset, error) {
	dbs := make([]interface{}, 0, len(c.server.databases))
	for key := range c.server.databases {
		dbs = append(dbs, key)
	}

	return c.buildSimpleShowResultset(dbs, "Database")
}

func (c *conn) handleShowTables(sql string, stmt *sqlparser.Show) (*Resultset, error) {
	if c.currentDB == nil {
		return nil, fmt.Errorf("no db select for show tables")
	}

	var tables []string
	for table := range c.currentDB.Tables {
		tables = append(tables, table)
	}

	sort.Strings(tables)

	values := make([]interface{}, len(tables))
	for i := range tables {
		values[i] = tables[i]
	}

	return c.buildSimpleShowResultset(values, fmt.Sprintf("Tables_in_%s", c.currentDB.Name))
}

func (c *conn) handleShowVariables(sql string, stmt *sqlparser.Show) (*Resultset, error) {
	variables := []interface{}{}
	return c.buildSimpleShowResultset(variables, "Variable")
}

func (c *conn) handleShowColumns(sql string, stmt *sqlparser.Show) (*Resultset, error) {

	var dbName string
	if c.currentDB != nil {
		dbName = c.currentDB.Name
	}

	table := ""

	switch f := stmt.From.(type) {
	case sqlparser.StrVal:
		table = string(f)
	case *sqlparser.ColName:
		if f.Qualifier != nil {
			dbName = string(f.Qualifier)
		}
		table = string(f.Name)
	default:
		return nil, fmt.Errorf("do not know how to show columns from type: %T", f)
	}

	if stmt.DBFilter != nil {
		switch f := stmt.DBFilter.(type) {
		case sqlparser.StrVal:
			dbName = string(f)
		case *sqlparser.ColName:
			dbName = string(f.Name)
		default:
			return nil, fmt.Errorf("do not know how to in show filter on db type: %T", f)
		}
	}

	db := c.server.databases[dbName]
	if db == nil {
		return nil, newDefaultError(ER_BAD_DB_ERROR, dbName)
	}

	tableSchema := db.Tables[table]
	if tableSchema == nil {
		return nil, fmt.Errorf("table (%s) does not exist in db (%s)", table, dbName)
	}

	full := strings.ToLower(stmt.Modifier) == "full"

	names, values, err := HandleShowColumns(tableSchema, full)
	if err != nil {
		return nil, err
	}

	return c.buildResultset(names, values)
}

func (c *conn) buildSimpleShowResultset(values []interface{}, name string) (*Resultset, error) {

	r := new(Resultset)

	field := &Field{}

	field.Name = Slice(name)
	field.Charset = 33
	field.Type = MYSQL_TYPE_VAR_STRING

	r.Fields = []*Field{field}

	var row []byte
	var err error

	for _, value := range values {
		row, err = formatValue(value)
		if err != nil {
			return nil, err
		}
		r.RowDatas = append(r.RowDatas, putLengthEncodedString(row))
	}

	return r, nil
}

func HandleShowColumns(tableSchema *schema.Table, full bool) ([]string, [][]interface{}, error) {
	if len(tableSchema.RawColumns) == 0 {
		return nil, nil, fmt.Errorf("no configured columns")
	}

	values := make([][]interface{}, len(tableSchema.RawColumns))
	names := []string{"Field", "Type", "Null", "Key", "Default", "Extra"}

	if full {
		names = append(names, []string{"Collation", "Privileges", "Comment"}...)
	}

	for num, col := range tableSchema.RawColumns {
		row := make([]interface{}, len(names))
		row[0] = evaluator.SQLVarchar(col.Name)
		row[1] = evaluator.SQLVarchar(string(col.SqlType))
		row[2] = evaluator.SQLVarchar("YES")
		row[3] = evaluator.SQLVarchar("")
		row[4] = evaluator.SQLNull
		row[5] = evaluator.SQLVarchar("")

		if full {
			row[6] = evaluator.SQLNull
			row[7] = evaluator.SQLVarchar("select")
			row[8] = evaluator.SQLVarchar("")
		}
		values[num] = row
	}
	return names, values, nil
}
