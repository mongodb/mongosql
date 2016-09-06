package server

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"

	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/parser"
)

type stmt struct {
	id uint32

	params  int
	columns int

	args []interface{}

	s parser.Statement

	sql string
}

func (s *stmt) ResetParams() {
	s.args = make([]interface{}, s.params)
}

func (c *conn) handleStmtPrepare(sql string) error {
	/*
		if c.schema == nil {
			return NewDefaultError(ER_NO_DB_ERROR)
		}

		s := new(Stmt)

		sql = strings.TrimRight(sql, ";")

		var err error
		s.s, err = parser.Parse(sql)
		if err != nil {
			return fmt.Errorf(`parse sql "%s" error`, sql)
		}

		s.sql = sql

		var tableName string
		switch s := s.s.(type) {
		case *parser.Select:
			tableName = nstring(s.From)
		case *parser.Insert:
			tableName = nstring(s.Table)
		case *parser.Update:
			tableName = nstring(s.Table)
		case *parser.Delete:
			tableName = nstring(s.Table)
		case *parser.Replace:
			tableName = nstring(s.Table)
		default:
			return fmt.Errorf(`unsupport prepare sql "%s"`, sql)
		}

		r := c.schema.rule.GetRule(tableName)

		n := c.server.getNode(r.Nodes[0])

		if co, err := n.getMasterConn(); err != nil {
			return fmt.Errorf("prepare error %s", err)
		} else {
			defer co.Close()

			if err = co.UseDB(c.schema.db); err != nil {
				return fmt.Errorf("parepre error %s", err)
			}

			if t, err := co.Prepare(sql); err != nil {
				return fmt.Errorf("parepre error %s", err)
			} else {

				s.params = t.ParamNum()
				s.columns = t.ColumnNum()
			}
		}

		s.id = c.stmtId
		c.stmtId++

		if err = c.writePrepare(s); err != nil {
			return err
		}

		s.ResetParams()

		c.stmts[s.id] = s

		return nil
	*/
	return mysqlerrors.Defaultf(mysqlerrors.ER_UNSUPPORTED_PS)
}

func (c *conn) writePrepare(s *stmt) error {
	data := make([]byte, 4, 128)

	//status ok
	data = append(data, 0)
	//stmt id
	data = append(data, uint32ToBytes(s.id)...)
	//number columns
	data = append(data, uint16ToBytes(uint16(s.columns))...)
	//number params
	data = append(data, uint16ToBytes(uint16(s.params))...)
	//filter [00]
	data = append(data, 0)
	//warning count
	data = append(data, 0, 0)

	if err := c.writePacket(data); err != nil {
		return err
	}

	if s.params > 0 {
		paramField := &Field{Name: []byte("?")}
		paramFieldData := paramField.Dump(c.variables.CharacterSetResults)
		for i := 0; i < s.params; i++ {
			data = data[0:4]
			data = append(data, paramFieldData...)

			if err := c.writePacket(data); err != nil {
				return err
			}
		}

		if err := c.writeEOF(c.status()); err != nil {
			return err
		}
	}

	if s.columns > 0 {
		columnField := &Field{}
		columnFieldData := columnField.Dump(c.variables.CharacterSetResults)
		for i := 0; i < s.columns; i++ {
			data = data[0:4]
			data = append(data, columnFieldData...)

			if err := c.writePacket(data); err != nil {
				return err
			}
		}

		if err := c.writeEOF(c.status()); err != nil {
			return err
		}

	}
	return nil
}

func (c *conn) handleStmtExecute(data []byte) error {
	if len(data) < 9 {
		return errMalformPacket
	}

	pos := 0
	id := binary.LittleEndian.Uint32(data[0:4])
	pos += 4

	s, ok := c.stmts[id]
	if !ok {
		return mysqlerrors.Defaultf(mysqlerrors.ER_UNKNOWN_STMT_HANDLER,
			strconv.FormatUint(uint64(id), 10), "stmt_execute")
	}

	flag := data[pos]
	pos++
	//now we only support CURSOR_TYPE_NO_CURSOR flag
	if flag != 0 {
		return fmt.Errorf("unsupported flag %d", flag)
	}

	//skip iteration-count, always 1
	pos += 4

	var nullBitmaps []byte
	var paramTypes []byte
	var paramValues []byte

	paramNum := s.params

	if paramNum > 0 {
		nullBitmapLen := (s.params + 7) >> 3
		if len(data) < (pos + nullBitmapLen + 1) {
			return errMalformPacket
		}
		nullBitmaps = data[pos : pos+nullBitmapLen]
		pos += nullBitmapLen

		//new param bound flag
		if data[pos] == 1 {
			pos++
			if len(data) < (pos + (paramNum << 1)) {
				return errMalformPacket
			}

			paramTypes = data[pos : pos+(paramNum<<1)]
			pos += (paramNum << 1)

			paramValues = data[pos:]
		}

		if err := c.bindStmtArgs(s, nullBitmaps, paramTypes, paramValues); err != nil {
			return err
		}
	}

	var err error

	switch stmt := s.s.(type) {
	case *parser.Select:
		err = c.handleSelect(stmt, s.sql, s.args)
	default:
		err = mysqlerrors.Defaultf(mysqlerrors.ER_UNSUPPORTED_PS)
	}

	s.ResetParams()

	return err
}

func (c *conn) bindStmtArgs(s *stmt, nullBitmap, paramTypes, paramValues []byte) error {
	args := s.args

	pos := 0

	var v []byte
	var n int
	var isNull bool
	var err error

	for i := 0; i < s.params; i++ {
		if nullBitmap[i>>3]&(1<<(uint(i)%8)) > 0 {
			args[i] = nil
			continue
		}

		tp := paramTypes[i<<1]
		isUnsigned := (paramTypes[(i<<1)+1] & 0x80) > 0

		switch tp {
		case MYSQL_TYPE_NULL:
			args[i] = nil
			continue

		case MYSQL_TYPE_TINY:
			if len(paramValues) < (pos + 1) {
				return errMalformPacket
			}

			if isUnsigned {
				args[i] = uint8(paramValues[pos])
			} else {
				args[i] = int8(paramValues[pos])
			}

			pos++
			continue

		case MYSQL_TYPE_SHORT, MYSQL_TYPE_YEAR:
			if len(paramValues) < (pos + 2) {
				return errMalformPacket
			}

			if isUnsigned {
				args[i] = uint16(binary.LittleEndian.Uint16(paramValues[pos : pos+2]))
			} else {
				args[i] = int16((binary.LittleEndian.Uint16(paramValues[pos : pos+2])))
			}
			pos += 2
			continue

		case MYSQL_TYPE_INT24, MYSQL_TYPE_LONG:
			if len(paramValues) < (pos + 4) {
				return errMalformPacket
			}

			if isUnsigned {
				args[i] = uint32(binary.LittleEndian.Uint32(paramValues[pos : pos+4]))
			} else {
				args[i] = int32(binary.LittleEndian.Uint32(paramValues[pos : pos+4]))
			}
			pos += 4
			continue

		case MYSQL_TYPE_LONGLONG:
			if len(paramValues) < (pos + 8) {
				return errMalformPacket
			}

			if isUnsigned {
				args[i] = binary.LittleEndian.Uint64(paramValues[pos : pos+8])
			} else {
				args[i] = int64(binary.LittleEndian.Uint64(paramValues[pos : pos+8]))
			}
			pos += 8
			continue

		case MYSQL_TYPE_FLOAT:
			if len(paramValues) < (pos + 4) {
				return errMalformPacket
			}

			args[i] = float32(math.Float32frombits(binary.LittleEndian.Uint32(paramValues[pos : pos+4])))
			pos += 4
			continue

		case MYSQL_TYPE_DOUBLE:
			if len(paramValues) < (pos + 8) {
				return errMalformPacket
			}

			args[i] = math.Float64frombits(binary.LittleEndian.Uint64(paramValues[pos : pos+8]))
			pos += 8
			continue

		case MYSQL_TYPE_DECIMAL, MYSQL_TYPE_NEWDECIMAL, MYSQL_TYPE_VARCHAR,
			MYSQL_TYPE_BIT, MYSQL_TYPE_ENUM, MYSQL_TYPE_SET, MYSQL_TYPE_TINY_BLOB,
			MYSQL_TYPE_MEDIUM_BLOB, MYSQL_TYPE_LONG_BLOB, MYSQL_TYPE_BLOB,
			MYSQL_TYPE_VAR_STRING, MYSQL_TYPE_STRING, MYSQL_TYPE_GEOMETRY,
			MYSQL_TYPE_DATE, MYSQL_TYPE_NEWDATE,
			MYSQL_TYPE_TIMESTAMP, MYSQL_TYPE_DATETIME, MYSQL_TYPE_TIME:
			if len(paramValues) < (pos + 1) {
				return errMalformPacket
			}

			v, isNull, n, err = lengthEncodedString(paramValues[pos:])
			pos += n
			if err != nil {
				return err
			}

			if !isNull {
				args[i] = v
				continue
			} else {
				args[i] = nil
				continue
			}
		default:
			return mysqlerrors.Unknownf("Stmt Unknown FieldType %d", tp)
		}
	}
	return nil
}

func (c *conn) handleStmtSendLongData(data []byte) error {
	if len(data) < 6 {
		return errMalformPacket
	}

	id := binary.LittleEndian.Uint32(data[0:4])

	s, ok := c.stmts[id]
	if !ok {
		return mysqlerrors.Defaultf(mysqlerrors.ER_UNKNOWN_STMT_HANDLER,
			strconv.FormatUint(uint64(id), 10), "stmt_send_longdata")
	}

	paramID := binary.LittleEndian.Uint16(data[4:6])
	if paramID >= uint16(s.params) {
		return mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_ARGUMENTS, "stmt_send_longdata")
	}

	if s.args[paramID] == nil {
		s.args[paramID] = data[6:]
	} else {
		if b, ok := s.args[paramID].([]byte); ok {
			b = append(b, data[6:]...)
			s.args[paramID] = b
		} else {
			return mysqlerrors.Unknownf("invalid param long data type %T", s.args[paramID])
		}
	}

	return nil
}

func (c *conn) handleStmtReset(data []byte) error {
	if len(data) < 4 {
		return errMalformPacket
	}

	id := binary.LittleEndian.Uint32(data[0:4])

	s, ok := c.stmts[id]
	if !ok {
		return mysqlerrors.Defaultf(mysqlerrors.ER_UNKNOWN_STMT_HANDLER,
			strconv.FormatUint(uint64(id), 10), "stmt_reset")
	}

	s.ResetParams()

	return c.writeOK(nil)
}

func (c *conn) handleStmtClose(data []byte) error {
	if len(data) < 4 {
		return nil
	}

	id := binary.LittleEndian.Uint32(data[0:4])

	delete(c.stmts, id)

	return nil
}
