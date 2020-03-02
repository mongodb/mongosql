package server

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/log"
)

type stmt struct {
	params int

	args []interface{}

	// s parser.Statement

	sql string
}

func (s *stmt) ResetParams() {
	s.args = make([]interface{}, s.params)
}

func (c *conn) handleStmtPrepare(_ string) error {
	return mysqlerrors.Defaultf(mysqlerrors.ErUnsupportedPs)
}

func (c *conn) handleStmtExecute(ctx context.Context, data []byte) error {
	if len(data) < 9 {
		return errMalformPacket
	}

	/* legacy code; not sure this has ever been tested or used */
	pos := 0
	id := binary.LittleEndian.Uint32(data[0:4])
	pos += 4

	s, ok := c.stmts[id]
	if !ok {
		return mysqlerrors.Defaultf(mysqlerrors.ErUnknownStmtHandler,
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

	lg := c.Logger(log.EvaluatorComponent)

	rCfg := c.getRewriterConfig()
	aCfg := c.getAlgebrizerConfig()
	oCfg := c.getOptimizerConfig()
	pCfg := c.getPushdownConfig()
	eCfg := c.getExecutionConfig()
	qCfg := evaluator.NewQueryConfig(lg, rCfg, aCfg, oCfg, pCfg, eCfg, false)

	_, _ = evaluator.ExecuteSQL(ctx, qCfg, s.sql)

	/* handleSelect no longer exists
	switch stmt := s.s.(type) {
	case *parser.Select:
		_, err = c.handleSelect(ctx, s.sql, stmt)
	default:
		err = mysqlerrors.Defaultf(mysqlerrors.ErUnsupportedPs)
	}
	*/

	s.ResetParams()

	return mysqlerrors.Defaultf(mysqlerrors.ErUnsupportedPs)
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
		case MySQLTypeNull:
			args[i] = nil
			continue

		case MySQLTypeTiny:
			if len(paramValues) < (pos + 1) {
				return errMalformPacket
			}

			if isUnsigned {
				args[i] = paramValues[pos]
			} else {
				args[i] = int8(paramValues[pos])
			}

			pos++
			continue

		case MySQLTypeShort, MySQLTypeYear:
			if len(paramValues) < (pos + 2) {
				return errMalformPacket
			}

			if isUnsigned {
				args[i] = binary.LittleEndian.Uint16(paramValues[pos : pos+2])
			} else {
				args[i] = int16((binary.LittleEndian.Uint16(paramValues[pos : pos+2])))
			}
			pos += 2
			continue

		case MySQLTypeInt24, MySQLTypeLong:
			if len(paramValues) < (pos + 4) {
				return errMalformPacket
			}

			if isUnsigned {
				args[i] = binary.LittleEndian.Uint32(paramValues[pos : pos+4])
			} else {
				args[i] = int32(binary.LittleEndian.Uint32(paramValues[pos : pos+4]))
			}
			pos += 4
			continue

		case MySQLTypeLongLong:
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

		case MySQLTypeFloat:
			if len(paramValues) < (pos + 4) {
				return errMalformPacket
			}

			args[i] = math.Float32frombits(binary.LittleEndian.Uint32(paramValues[pos : pos+4]))
			pos += 4
			continue

		case MySQLTypeDouble:
			if len(paramValues) < (pos + 8) {
				return errMalformPacket
			}

			args[i] = math.Float64frombits(binary.LittleEndian.Uint64(paramValues[pos : pos+8]))
			pos += 8
			continue

		case MySQLTypeDecimal, MySQLTypeNewDecimal, MySQLTypeVarchar,
			MySQLTypeBit, MySQLTypeEnum, MySQLTypeSet, MySQLTypeTinyBlob,
			MySQLTypeMediumBlob, MySQLTypeLongBlob, MySQLTypeBlob,
			MySQLTypeVarString, MySQLTypeString, MySQLTypeGeometry,
			MySQLTypeDate, MySQLTypeNewDate,
			MySQLTypeTimestamp, MySQLTypeDatetime, MySQLTypeTime:
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
		return mysqlerrors.Defaultf(mysqlerrors.ErUnknownStmtHandler,
			strconv.FormatUint(uint64(id), 10), "stmt_send_longdata")
	}

	paramID := binary.LittleEndian.Uint16(data[4:6])
	if paramID >= uint16(s.params) {
		return mysqlerrors.Defaultf(mysqlerrors.ErWrongArguments, "stmt_send_longdata")
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
		return mysqlerrors.Defaultf(mysqlerrors.ErUnknownStmtHandler,
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
