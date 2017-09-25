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

func (c *conn) handleSelect(sql string, stmt parser.SelectStatement) error {
	fields, iter, err := evaluator.EvaluateQuery(sql, stmt, c)
	if err != nil {
		return err
	}

	return c.streamResultset(fields, iter)
}

func (c *conn) handleFieldList(data []byte) error {

	index := bytes.IndexByte(data, 0x00)
	tableName := String(c.variables.CharacterSetClient.Decode(data[0:index]))
	wildcard := String(c.variables.CharacterSetClient.Decode(data[index+1:]))

	db, err := c.catalog.Database(c.DB())
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
		mongoColumn, ok := column.(*catalog.MongoColumn)
		if ok && mongoColumn.MongoType == schema.MongoFilter {
			continue
		}

		name, table, database := []byte(column.Name()),
			[]byte(tableName), []byte(catalog.InformationSchemaDatabase)

		field := &Field{
			Name:          name,
			OriginalName:  name,
			Schema:        database,
			Table:         table,
			OriginalTable: table,
		}

		zeroValue := column.Type().ZeroValue()
		value, err := evaluator.NewSQLValueFromSQLColumnExpr(
			zeroValue, schema.SQLNone, schema.MongoNone)
		if err != nil {
			return err
		}

		err = formatField(c.variables, uint16(col.ID), field, value)
		if err != nil {
			return err
		}
		fields = append(fields, field)
	}

	c.Logger(log.NetworkComponent).Debugf(log.Dev, "handleFieldList table: %v, wildcard: %v", tableName, wildcard)

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
