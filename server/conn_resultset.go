package server

import (
	"strconv"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
	"github.com/shopspring/decimal"
)

func (c *conn) formatValue(value interface{}) ([]byte, error) {
	switch v := value.(type) {

	case evaluator.SQLVarchar:
		return c.variables.CharacterSetResults.Encode(Slice(string(v))), nil
	case evaluator.SQLObjectID:
		return []byte(string(v)), nil
	case evaluator.SQLInt:
		return strconv.AppendInt(nil, int64(v), 10), nil
	case evaluator.SQLDecimal128:
		return []byte(decimal.Decimal(v).String()), nil
	case evaluator.SQLUint32:
		return strconv.AppendUint(nil, uint64(v), 10), nil
	case evaluator.SQLUint64:
		return strconv.AppendUint(nil, uint64(v), 10), nil
	case evaluator.SQLFloat:
		return strconv.AppendFloat(nil, float64(v), 'f', -1, 64), nil
	case evaluator.SQLValues:
		slice := []byte{}
		for _, value := range v.Values {
			b, err := c.formatValue(value)
			if err != nil {
				return nil, err
			}
			slice = append(slice, b...)
		}
		return slice, nil
	case *evaluator.SQLTupleExpr:
		slice := []byte{}
		for _, expr := range v.Exprs {
			b, err := c.formatValue(expr)
			if err != nil {
				return nil, err
			}
			slice = append(slice, b...)
		}
		return slice, nil
	case evaluator.SQLNullValue, *evaluator.SQLNullValue, evaluator.SQLNoValue:
		return nil, nil
	case evaluator.SQLBool:
		if bool(v) {
			return []byte{'1'}, nil
		}
		return []byte{'0'}, nil
	case *evaluator.SQLValues:
		return c.formatValue(v.Values[0])

	// SQL time related values
	case evaluator.SQLDate:
		return Slice(v.Time.Format(schema.DateFormat)), nil
	case evaluator.SQLTimestamp:
		return Slice(v.Time.Format(schema.TimestampFormat)), nil

	// TODO (INT-1036): get rid of these and only use SQLValues here.
	case int8:
		return strconv.AppendInt(nil, int64(v), 10), nil
	case int16:
		return strconv.AppendInt(nil, int64(v), 10), nil
	case int32:
		return strconv.AppendInt(nil, int64(v), 10), nil
	case int64:
		return strconv.AppendInt(nil, int64(v), 10), nil
	case int:
		return strconv.AppendInt(nil, int64(v), 10), nil
	case uint8:
		return strconv.AppendUint(nil, uint64(v), 10), nil
	case uint16:
		return strconv.AppendUint(nil, uint64(v), 10), nil
	case uint32:
		return strconv.AppendUint(nil, uint64(v), 10), nil
	case uint64:
		return strconv.AppendUint(nil, uint64(v), 10), nil
	case uint:
		return strconv.AppendUint(nil, uint64(v), 10), nil
	case float32:
		return strconv.AppendFloat(nil, float64(v), 'f', -1, 64), nil
	case float64:
		return strconv.AppendFloat(nil, float64(v), 'f', -1, 64), nil
	case []byte:
		return v, nil
	case string:
		return c.variables.CharacterSetResults.Encode(Slice(v)), nil
	case bool:
		if v {
			return []byte{'1'}, nil
		}
		return []byte{'0'}, nil
	case nil:
		return nil, nil
	default:
		return nil, mysqlerrors.Unknownf("invalid type %T", value)
	}
}

func formatField(collationID uint16, field *Field, value interface{}) error {
	switch typedV := value.(type) {

	case evaluator.SQLFloat:
		field.Charset = collationID
		field.Type = MYSQL_TYPE_FLOAT
		field.Flag = BINARY_FLAG | NOT_NULL_FLAG
	case evaluator.SQLDecimal128:
		field.Charset = collationID
		field.Type = MYSQL_TYPE_DECIMAL
		field.Decimal = 0x51
		field.Flag = BINARY_FLAG | NOT_NULL_FLAG
	case evaluator.SQLBool:
		field.Charset = collationID
		field.Type = MYSQL_TYPE_TINY
	case evaluator.SQLUint32:
		field.Charset = collationID
		field.Type = MYSQL_TYPE_LONGLONG
		field.Flag = BINARY_FLAG | NOT_NULL_FLAG | UNSIGNED_FLAG
	case evaluator.SQLUint64:
		field.Charset = collationID
		field.Type = MYSQL_TYPE_LONGLONG
		field.Flag = BINARY_FLAG | NOT_NULL_FLAG | UNSIGNED_FLAG
	case evaluator.SQLInt:
		field.Charset = collationID
		field.Type = MYSQL_TYPE_LONGLONG
		field.Flag = BINARY_FLAG | NOT_NULL_FLAG
	case evaluator.SQLVarchar:
		field.Charset = collationID
		field.Type = MYSQL_TYPE_VAR_STRING
	case evaluator.SQLObjectID:
		field.Charset = collationID
		field.Type = MYSQL_TYPE_VAR_STRING
	case evaluator.SQLValues:
		field.Charset = collationID
		field.Type = MYSQL_TYPE_VAR_STRING
	case nil, *evaluator.SQLNullValue, evaluator.SQLNullValue, evaluator.SQLNoValue:
		field.Charset = collationID
		field.Type = MYSQL_TYPE_NULL
	case evaluator.SQLDate:
		field.Charset = collationID
		field.Type = MYSQL_TYPE_DATE
	case evaluator.SQLTimestamp:
		field.Charset = collationID
		field.Type = MYSQL_TYPE_TIMESTAMP
	case *evaluator.SQLValues:
		if len(typedV.Values) != 1 {
			return mysqlerrors.Defaultf(mysqlerrors.ER_OPERAND_COLUMNS, 1)
		}
	default:
		return mysqlerrors.Unknownf("unsupported type %T for result set", value)
	}

	return nil
}

// streamResultset implements the COM_QUERY response.
// More at https://dev.mysql.com/doc/internals/en/com-query-response.html
func (c *conn) streamResultset(columns []*evaluator.Column, iter evaluator.Iter) error {

	// If the number of columns in the resultset is 0, write an OK packet
	if len(columns) == 0 {
		return c.writeOK(nil)
	}

	c.affectedRows = int64(-1)

	status := c.status()

	columnLen := putLengthEncodedInt(uint64(len(columns)))

	data := make([]byte, 4, 1024)

	data = append(data, columnLen...)

	var err error

	// write column count
	if err = c.writePacket(data); err != nil {
		return err
	}

	col, err := collation.Get(c.variables.CharacterSetResults.DefaultCollationName)
	if err != nil {
		return err
	}

	var wroteHeaders bool

	writeHeaders := func() error {

		numFields := len(columns)

		if numFields == 0 {
			return mysqlerrors.Unknownf("No columns found in result set")
		}

		j := 0

		var value evaluator.SQLValue
		for j < numFields {
			zeroValue := columns[j].SQLType.ZeroValue()
			value, err = evaluator.NewSQLValueFromSQLColumnExpr(zeroValue, columns[j].SQLType, columns[j].MongoType)
			if err != nil {
				return err
			}

			name := c.variables.CharacterSetResults.Encode(Slice(columns[j].Name))
			field := &Field{
				Name: name,
			}

			if err = formatField(uint16(col.ID), field, value); err != nil {
				return err
			}

			data = data[0:4]
			data = append(data, field.Dump(c.variables.CharacterSetResults)...)

			// write a column definition packet for each
			// column in the result set
			if err := c.writePacket(data); err != nil {
				return err
			}

			j++
		}

		// end the column definitions with an EOF packet
		return c.writeEOF(status)
	}

	var b []byte

	rowChan := make(chan []interface{}, 1)

	go func() {
		evaluatorRow := &evaluator.Row{}
		for iter.Next(evaluatorRow) {
			rowChan <- evaluatorRow.GetValues()
			evaluatorRow.Data = evaluator.Values{}
		}
		close(rowChan)
	}()

streamer:
	for {
		select {
		case values, ok := <-rowChan:
			if !ok {
				break streamer
			}

			// write the headers once
			if !wroteHeaders {
				if err = writeHeaders(); err != nil {
					return err
				}
				wroteHeaders = true
			}

			data = data[0:4]

			for _, value := range values {
				b, err = c.formatValue(value)
				if err != nil {
					return err
				}
				if b == nil {
					data = append(data, 0xfb)
				} else {
					data = append(data, putLengthEncodedString(b)...)
				}
			}

			// write each row as a separate packet
			if err := c.writePacket(data); err != nil {
				return err
			}

		case <-c.tomb.Dying():
			iter.Close()
			return c.tomb.Err()
		}
	}

	if err = iter.Close(); err != nil {
		return mysqlerrors.Newf(mysqlerrors.ER_QUERY_ON_FOREIGN_DATA_SOURCE, "iterator close error: %v", err)
	}

	if err = iter.Err(); err != nil {
		return mysqlerrors.Newf(mysqlerrors.ER_QUERY_ON_FOREIGN_DATA_SOURCE, "iterator err: %v", err)
	}

	if !wroteHeaders {
		if err = writeHeaders(); err != nil {
			return err
		}
	}

	return c.writeEOF(status)
}
