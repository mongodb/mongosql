package server

import (
	"bytes"

	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
)

func (c *conn) handleSelect(sql string, stmt parser.Statement) error {
	fields, iter, err := evaluator.EvaluateQuery(sql, stmt, c)
	if err != nil {
		if iter != nil {
			iter.Close()
		}
		if ctxErr := c.Context().Err(); ctxErr != nil {
			c.refreshContext()
			err = ctxErr
		}
		return err
	}
	return c.streamResultset(fields, iter)
}

func (c *conn) handleFieldList(data []byte) error {

	index := bytes.IndexByte(data, 0x00)
	charSetClient := c.variables.GetCharset(variable.CharacterSetClient)
	tableName := util.String(charSetClient.Decode(data[0:index]))
	wildcard := util.String(charSetClient.Decode(data[index+1:]))

	db, err := c.catalog.Database(c.DB())
	if err != nil {
		return err
	}

	tableSchema, err := db.Table(tableName)
	if err != nil {
		return err
	}

	col, err := collation.Get(c.variables.GetCharset(variable.CharacterSetResults).DefaultCollationName)
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
		data = append(data, v.Dump(c.variables.GetCharset(variable.CharacterSetResults))...)
		if err := c.writePacket(data); err != nil {
			return err
		}
	}

	return c.writeEOF(status)
}
