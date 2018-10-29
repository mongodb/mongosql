package server

import (
	"bytes"
	"context"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/catalog"
	"github.com/10gen/sqlproxy/internal/collation"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/internal/variable"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
)

func (c *conn) handleSelect(ctx context.Context, sql string, stmt parser.Statement) error {
	aCfg := c.getAlgebrizerConfig(sql, stmt)
	oCfg := c.getOptimizerConfig()
	pCfg := c.getPushdownConfig()
	eCfg := c.getExecutionConfig()

	res, err := evaluator.EvaluateQuery(ctx, aCfg, oCfg, pCfg, eCfg)

	if err != nil {
		return err
	}
	return c.streamResultset(ctx, res.Columns, res.Iter)
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

	col, err := collation.Get(
		c.variables.GetCharset(variable.CharacterSetResults).DefaultCollationName)
	if err != nil {
		return err
	}

	fields := []*Field{}

	valueKind := evaluator.GetSQLValueKind(c.variables)
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
			Charset:       uint16(col.ID),
		}

		zeroValue := evaluator.SQLTypeToEvalType(column.Type()).ZeroValue(valueKind)
		if err = formatHeaderField(c.variables, field, zeroValue); err != nil {
			return err
		}
		fields = append(fields, field)
	}

	c.Logger(log.NetworkComponent).Debugf(log.Dev, "handleFieldList table: %v, wildcard: %v",
		tableName, wildcard)

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
