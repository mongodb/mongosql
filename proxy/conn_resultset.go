package proxy

import (
	"fmt"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/schema"
	"github.com/deafgoat/mixer/hack"
	. "github.com/deafgoat/mixer/mysql"
	"strconv"
)

func formatValue(value interface{}) ([]byte, error) {
	switch v := value.(type) {

	case evaluator.SQLString:
		return hack.Slice(string(v)), nil
	case evaluator.SQLObjectID:
		return hack.Slice(string(v)), nil
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
		return hack.Slice(v.Time.Format(schema.DateFormat)), nil
	case evaluator.SQLTimestamp:
		return hack.Slice(v.Time.Format(schema.TimestampFormat)), nil

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
		return hack.Slice(v), nil
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
	case evaluator.SQLString:
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

func (c *Conn) buildResultset(names []string, values [][]interface{}) (*Resultset, error) {
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
				field.Name = hack.Slice(names[j])

				if err = formatField(field, value); err != nil {
					return nil, err
				}
			}
			b, err = formatValue(value)

			if err != nil {
				return nil, err
			}

			row = append(row, PutLengthEncodedString(b)...)
		}

		r.RowDatas = append(r.RowDatas, row)
	}

	if len(values) == 0 {
		for j, nm := range names {
			field := &Field{}
			r.Fields[j] = field
			field.Name = hack.Slice(nm)

			if err = formatField(field, nil); err != nil {
				return nil, err
			}
		}
	}

	return r, nil
}

func (c *Conn) writeResultset(status uint16, r *Resultset) error {
	c.affectedRows = int64(-1)

	columnLen := PutLengthEncodedInt(uint64(len(r.Fields)))

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
