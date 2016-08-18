package server

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
	"github.com/mongodb/mongo-tools/common/log"
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

func (c *conn) handleSelect(stmt *parser.Select, sql string, args []interface{}) error {
	log.Logf(log.DebugHigh, "[conn%v] parsed statement: %#v", c.connectionID, stmt)

	fields, iter, err := c.server.eval.Evaluate(stmt, c)
	if err != nil {
		return err
	}

	return c.streamResultset(fields, iter)
}

func (c *conn) handleSimpleSelect(sql string, stmt *parser.SimpleSelect) error {
	log.Logf(log.DebugHigh, "[conn%v] parsed statement: %#v", c.connectionID, stmt)

	fields, iter, err := c.server.eval.Evaluate(stmt, c)
	if err != nil {
		return err
	}

	return c.streamResultset(fields, iter)
}

func (c *conn) buildSimpleSelectResult(value interface{}, name []byte, asName []byte) (*Resultset, error) {
	field := &Field{}

	field.Name = name

	if asName != nil {
		field.Name = asName
	}

	field.OrgName = name

	collationID := uint16(c.getCollationID())

	formatField(collationID, field, value)

	r := &Resultset{Fields: []*Field{field}}
	row, err := formatValue(value)
	if err != nil {
		return nil, err
	}
	r.RowDatas = append(r.RowDatas, putLengthEncodedString(row))

	return r, nil
}

func (c *conn) handleFieldList(data []byte) error {

	index := bytes.IndexByte(data, 0x00)
	table := string(data[0:index])
	wildcard := string(data[index+1:])

	dbName := c.currentDB.Name

	db := c.server.databases[strings.ToLower(dbName)]
	if db == nil {
		return mysqlerrors.Defaultf(mysqlerrors.ER_BAD_DB_ERROR, dbName)
	}

	tableSchema := db.Tables[strings.ToLower(table)]
	if tableSchema == nil {
		return mysqlerrors.Defaultf(mysqlerrors.ER_UNKNOWN_TABLE, table, dbName)
	}

	collationID := uint16(c.getCollationID())

	fields := []*Field{}

	for _, field := range tableSchema.RawColumns {
		f := &Field{}
		f.Name = []byte(field.SqlName)
		zeroValue := field.SqlType.ZeroValue()
		value, err := evaluator.NewSQLValueFromSQLColumnExpr(zeroValue, schema.SQLNone, schema.MongoNone)
		if err != nil {
			return err
		}
		if err = formatField(collationID, f, value); err != nil {
			return err
		}
		fields = append(fields, f)
	}

	log.Logf(log.DebugLow, "handleFieldList table: %v, wildcard: %v", table, wildcard)

	return c.writeFieldList(c.status, fields)
}

func (c *conn) writeFieldList(status uint16, fs []*Field) error {
	c.affectedRows = int64(-1)

	data := make([]byte, 4, 1024)

	for _, v := range fs {
		data = data[0:4]
		data = append(data, v.Dump()...)
		if err := c.writePacket(data); err != nil {
			return err
		}
	}

	if err := c.writeEOF(status); err != nil {
		return err
	}
	return nil
}
