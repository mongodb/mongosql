package proxy

import (
	"bytes"
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"github.com/mongodb/mongo-tools/common/log"
	. "github.com/siddontang/mixer/mysql"
	"strings"
)

func (c *Conn) handleSimpleSelect(sql string, stmt *sqlparser.SimpleSelect) error {
	log.Logf(log.DebugLow, "handleSimpleSelect")

	if len(stmt.SelectExprs) != 1 {
		return fmt.Errorf("support SimpleSelect SelectExprs, %s", sql)
	}

	expr, ok := stmt.SelectExprs[0].(*sqlparser.NonStarExpr)
	if !ok {
		return fmt.Errorf("support SimpleSelect NonStarExpr, %s", sql)
	}

	var f *sqlparser.FuncExpr
	f, ok = expr.Expr.(*sqlparser.FuncExpr)
	if !ok {
		return fmt.Errorf("support SimpleSelect FuncExpr, %s %T", sql, expr)
	}

	var r *Resultset
	var err error

	switch strings.ToLower(string(f.Name)) {
	case "last_insert_id":
		r, err = c.buildSimpleSelectResult(c.lastInsertId, f.Name, expr.As)
	case "row_count":
		r, err = c.buildSimpleSelectResult(c.affectedRows, f.Name, expr.As)
	case "version":
		r, err = c.buildSimpleSelectResult(ServerVersion, f.Name, expr.As)
	case "connection_id":
		r, err = c.buildSimpleSelectResult(c.connectionId, f.Name, expr.As)
	case "database":
		if c.currentSchema != nil {
			r, err = c.buildSimpleSelectResult(c.currentSchema.DB, f.Name, expr.As)
		} else {
			r, err = c.buildSimpleSelectResult("NULL", f.Name, expr.As)
		}
	default:
		return fmt.Errorf("function %s not support", f.Name)
	}

	if err != nil {
		return err
	}

	return c.writeResultset(c.status, r)
}

func (c *Conn) buildSimpleSelectResult(value interface{}, name []byte, asName []byte) (*Resultset, error) {
	field := &Field{}

	field.Name = name

	if asName != nil {
		field.Name = asName
	}

	field.OrgName = name

	formatField(field, value)

	r := &Resultset{Fields: []*Field{field}}
	row, err := formatValue(value)
	if err != nil {
		return nil, err
	}
	r.RowDatas = append(r.RowDatas, PutLengthEncodedString(row))

	return r, nil
}

func (c *Conn) handleFieldList(data []byte) error {

	index := bytes.IndexByte(data, 0x00)
	table := string(data[0:index])
	wildcard := string(data[index+1:])

	// TODO: hack but valid since _id always exists in MongoDB
	log.Logf(log.DebugLow, "handleFieldList table: %v", table)
	log.Logf(log.DebugLow, "handleFieldList wildcard: %v", wildcard)

	f := &Field{}
	f.Name = []byte("_id")
	f.OrgName = f.Name

	return c.writeFieldList(c.status, []*Field{f})
}

func (c *Conn) writeFieldList(status uint16, fs []*Field) error {
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
