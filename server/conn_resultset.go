package server

import (
	"fmt"
	"strconv"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/schema"
	"github.com/mongodb/mongo-tools/common/log"
)

func formatValue(value interface{}) ([]byte, error) {
	switch v := value.(type) {

	case evaluator.SQLVarchar:
		return Slice(string(v)), nil
	case evaluator.SQLObjectID:
		return Slice(string(v)), nil
	case evaluator.SQLInt:
		return strconv.AppendInt(nil, int64(v), 10), nil
	case evaluator.SQLUint32:
		return strconv.AppendUint(nil, uint64(v), 10), nil
	case evaluator.SQLFloat:
		return strconv.AppendFloat(nil, float64(v), 'f', -1, 64), nil
	case evaluator.SQLValues:
		slice := []byte{}
		for _, value := range v.Values {
			b, err := formatValue(value)
			if err != nil {
				return nil, err
			}
			slice = append(slice, b...)
		}
		return slice, nil
	case *evaluator.SQLTupleExpr:
		slice := []byte{}
		for _, expr := range v.Exprs {
			b, err := formatValue(expr)
			if err != nil {
				return nil, err
			}
			slice = append(slice, b...)
		}
		return slice, nil
	case evaluator.SQLNullValue, *evaluator.SQLNullValue, evaluator.SQLNoValue:
		return nil, nil
	case evaluator.SQLBool:
		return strconv.AppendBool(nil, bool(v)), nil
	case *evaluator.SQLValues:
		return formatValue(v.Values[0])

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
		return Slice(v), nil
	case bool:
		return strconv.AppendBool(nil, v), nil
	case nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("invalid type %T", value)
	}
}

func formatField(field *Field, value interface{}) error {
	switch typedV := value.(type) {

	case evaluator.SQLFloat:
		field.Charset = 63
		field.Type = MYSQL_TYPE_FLOAT
		field.Flag = BINARY_FLAG | NOT_NULL_FLAG
	case evaluator.SQLBool:
		field.Charset = 33
		field.Type = MYSQL_TYPE_BIT
	case evaluator.SQLUint32:
		field.Charset = 63
		field.Type = MYSQL_TYPE_LONGLONG
		field.Flag = BINARY_FLAG | NOT_NULL_FLAG | UNSIGNED_FLAG
	case evaluator.SQLInt:
		field.Charset = 63
		field.Type = MYSQL_TYPE_LONGLONG
		field.Flag = BINARY_FLAG | NOT_NULL_FLAG
	case evaluator.SQLVarchar:
		field.Charset = 33
		field.Type = MYSQL_TYPE_VAR_STRING
	case evaluator.SQLObjectID:
		field.Charset = 33
		field.Type = MYSQL_TYPE_VAR_STRING
	case evaluator.SQLValues:
		field.Charset = 33
		field.Type = MYSQL_TYPE_VAR_STRING
	case nil, *evaluator.SQLNullValue, evaluator.SQLNullValue, evaluator.SQLNoValue:
		field.Charset = 33
		field.Type = MYSQL_TYPE_NULL
	case evaluator.SQLDate:
		field.Charset = 33
		field.Type = MYSQL_TYPE_DATE
	case evaluator.SQLTimestamp:
		field.Charset = 33
		field.Type = MYSQL_TYPE_TIMESTAMP
	case *evaluator.SQLValues:
		if len(typedV.Values) != 1 {
			return fmt.Errorf("Operand should contain 1 column")
		}
	default:
		return fmt.Errorf("unsupported type %T for result set", value)
	}

	return nil
}

func (c *conn) buildResultset(names []string, values [][]interface{}) (*Resultset, error) {
	r := new(Resultset)

	r.Fields = make([]*Field, len(names))

	var b []byte
	var err error

	for i, vs := range values {
		if len(vs) != len(r.Fields) {
			return nil, fmt.Errorf("row %d has %d column not equal %d", i, len(vs), len(r.Fields))
		}

		var row []byte
		for j, value := range vs {
			if i == 0 {
				field := &Field{}
				r.Fields[j] = field
				field.Name = Slice(names[j])

				if err = formatField(field, value); err != nil {
					return nil, err
				}
			}

			b, err = formatValue(value)
			if err != nil {
				return nil, err
			}

			if b == nil {
				row = append(row, 0xfb)
			} else {
				row = append(row, putLengthEncodedString(b)...)
			}
		}

		r.RowDatas = append(r.RowDatas, row)
	}

	if len(values) == 0 {
		for j, nm := range names {
			field := &Field{}
			r.Fields[j] = field
			field.Name = Slice(nm)

			if err = formatField(field, evaluator.SQLVarchar("")); err != nil {
				return nil, err
			}
		}
	}

	return r, nil
}

// streamResultset implements the COM_QUERY response.
// More at https://dev.mysql.com/doc/internals/en/com-query-response.html
func (c *conn) streamResultset(columns []*evaluator.Column, iter evaluator.Iter) error {

	// If the number of columns in the resultset is 0, write an OK packet
	if len(columns) == 0 {
		return c.writeOK(nil)
	}

	c.affectedRows = int64(-1)

	status := c.status

	columnLen := putLengthEncodedInt(uint64(len(columns)))

	data := make([]byte, 4, 1024)

	data = append(data, columnLen...)

	var err error

	// write column count
	if err = c.writePacket(data); err != nil {
		return err
	}

	var wroteHeaders bool

	writeHeaders := func(values []interface{}) error {

		numFields := len(values)

		if values == nil {
			numFields = len(columns)
		}

		if numFields == 0 {
			return fmt.Errorf("no headers returned")
		}

		var value interface{}

		j := 0

		for j < numFields {
			if values == nil {
				zeroValue := columns[j].SQLType.ZeroValue()
				value, err = evaluator.NewSQLValue(zeroValue, schema.SQLNone, schema.MongoNone)
				if err != nil {
					return err
				}
			} else {
				value = values[j]
			}

			field := &Field{
				Name: Slice(columns[j].View),
			}

			if err = formatField(field, value); err != nil {
				return err
			}

			data = data[0:4]
			data = append(data, field.Dump()...)

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

	evaluatorRow := &evaluator.Row{}

	for iter.Next(evaluatorRow) {

		values := evaluatorRow.GetValues(columns)

		// write the headers once
		if !wroteHeaders {
			if err = writeHeaders(values); err != nil {
				return err
			}
			wroteHeaders = true
		}

		data = data[0:4]

		for _, value := range values {
			b, err = formatValue(value)
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

		evaluatorRow.Data = []evaluator.TableRow{}
	}

	if err = iter.Close(); err != nil {
		return fmt.Errorf("iterator close error: %v", err)
	}

	if err = iter.Err(); err != nil {
		return fmt.Errorf("iterator err: %v", err)
	}

	log.Logf(log.DebugLow, "[conn%v] done executing plan", c.ConnectionId())

	if !wroteHeaders {
		if err = writeHeaders(nil); err != nil {
			return err
		}
	}

	return c.writeEOF(status)
}

func (c *conn) writeResultset(status uint16, r *Resultset) error {
	c.affectedRows = int64(-1)

	columnLen := putLengthEncodedInt(uint64(len(r.Fields)))

	data := make([]byte, 4, 1024)

	data = append(data, columnLen...)
	if err := c.writePacket(data); err != nil {
		return err
	}

	for _, v := range r.Fields {
		data = data[0:4]
		data = append(data, v.Dump()...)
		if err := c.writePacket(data); err != nil {
			return err
		}
	}

	if err := c.writeEOF(status); err != nil {
		return err
	}

	for _, v := range r.RowDatas {
		data = data[0:4]
		data = append(data, v...)
		if err := c.writePacket(data); err != nil {
			return err
		}
	}

	if err := c.writeEOF(status); err != nil {
		return err
	}

	return nil
}
