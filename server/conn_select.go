package server

import (
	"bytes"
	"fmt"

	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
)

func makeBindVars(args []interface{}) map[string]interface{} {
	bindVars := make(map[string]interface{}, len(args))

	if args != nil {
		for i, v := range args {
			bindVars[fmt.Sprintf("v%d", i+1)] = v
		}
	}

	return bindVars
}

func (c *conn) handleSelect(stmt parser.SelectStatement, sql string, args []interface{}) error {
	fields, iter, err := c.server.eval.EvaluateQuery(stmt, c)
	if err != nil {
		return err
	}

	return c.streamResultset(fields, iter)
}

func (c *conn) handleSimpleSelect(sql string, stmt *parser.SimpleSelect) error {
	fields, iter, err := c.server.eval.EvaluateQuery(stmt, c)
	if err != nil {
		return err
	}

	return c.streamResultset(fields, iter)
}

func (c *conn) handleFieldList(data []byte) error {

	index := bytes.IndexByte(data, 0x00)
	tableName := String(c.variables.CharacterSetClient.Decode(data[0:index]))
	wildcard := String(c.variables.CharacterSetClient.Decode(data[index+1:]))

	dbName := c.currentDB.Name

	db, err := c.catalog.Database(string(dbName))
	if err != nil {
		return err
	}

	tableSchema, err := db.Table(tableName)
	if err != nil {
		return err
	}

	col, err := collation.Get(c.variables.CharacterSetResults.DefaultCollationName)
	if err != nil {
		return err
	}

	fields := []*Field{}

	for _, column := range tableSchema.Columns() {
		if mongoColumn, ok := column.(*catalog.MongoColumn); ok && mongoColumn.MongoType == schema.MongoFilter {
			continue
		}

		f := &Field{}
		f.Name = []byte(column.Name())
		zeroValue := column.Type().ZeroValue()
		value, err := evaluator.NewSQLValueFromSQLColumnExpr(zeroValue, schema.SQLNone, schema.MongoNone)
		if err != nil {
			return err
		}
		if err = formatField(uint16(col.ID), f, value); err != nil {
			return err
		}
		fields = append(fields, f)
	}

	c.Logger(log.NetworkComponent).Logf(log.DebugHigh, "handleFieldList table: %v, wildcard: %v", tableName, wildcard)

	return c.writeFieldList(c.status(), fields)
}

func (c *conn) writeFieldList(status uint16, fs []*Field) error {
	c.affectedRows = int64(-1)

	data := make([]byte, 4, 1024)

	for _, v := range fs {
		data = data[0:4]
		data = append(data, v.Dump(c.variables.CharacterSetResults)...)
		if err := c.writePacket(data); err != nil {
			return err
		}
	}

	if err := c.writeEOF(status); err != nil {
		return err
	}
	return nil
}
