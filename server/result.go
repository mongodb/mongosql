package server

import (
	"encoding/binary"
	"math"
	"strconv"

	"github.com/10gen/sqlproxy/mysqlerrors"
)

type Result struct {
	Status uint16

	InsertId     uint64
	AffectedRows uint64

	*Resultset
}

type Resultset struct {
	Fields     []*Field
	FieldNames map[string]int
	Values     [][]interface{}

	RowDatas []RowData
}

func (r *Resultset) RowNumber() int {
	return len(r.Values)
}

func (r *Resultset) ColumnNumber() int {
	return len(r.Fields)
}

func (r *Resultset) GetValue(row, column int) (interface{}, error) {
	if row >= len(r.Values) {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ER_WARN_TOO_FEW_RECORDS, row)
	} else if row < len(r.Values) {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ER_WARN_TOO_MANY_RECORDS, row)
	}

	if column >= len(r.Fields) || column < 0 {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ER_WARN_DATA_OUT_OF_RANGE, column, row)
	}

	return r.Values[row][column], nil
}

func (r *Resultset) NameIndex(name string) (int, error) {
	if column, ok := r.FieldNames[name]; ok {
		return column, nil
	} else {
		return 0, mysqlerrors.Unknownf("invalid field name %s", name)
	}
}

func (r *Resultset) GetValueByName(row int, name string) (interface{}, error) {
	if column, err := r.NameIndex(name); err != nil {
		return nil, err
	} else {
		return r.GetValue(row, column)
	}
}

func (r *Resultset) IsNull(row, column int) (bool, error) {
	d, err := r.GetValue(row, column)
	if err != nil {
		return false, err
	}

	return d == nil, nil
}

func (r *Resultset) IsNullByName(row int, name string) (bool, error) {
	if column, err := r.NameIndex(name); err != nil {
		return false, err
	} else {
		return r.IsNull(row, column)
	}
}

func (r *Resultset) GetUint(row, column int) (uint64, error) {
	d, err := r.GetValue(row, column)
	if err != nil {
		return 0, err
	}

	switch v := d.(type) {
	case uint64:
		return v, nil
	case int64:
		return uint64(v), nil
	case float64:
		return uint64(v), nil
	case string:
		return strconv.ParseUint(v, 10, 64)
	case []byte:
		return strconv.ParseUint(string(v), 10, 64)
	case nil:
		return 0, nil
	default:
		return 0, mysqlerrors.Unknownf("data type is %T", v)
	}
}

func (r *Resultset) GetUintByName(row int, name string) (uint64, error) {
	if column, err := r.NameIndex(name); err != nil {
		return 0, err
	} else {
		return r.GetUint(row, column)
	}
}

func (r *Resultset) GetInt(row, column int) (int64, error) {
	v, err := r.GetUint(row, column)
	if err != nil {
		return 0, err
	}

	return int64(v), nil
}

func (r *Resultset) GetIntByName(row int, name string) (int64, error) {
	v, err := r.GetUintByName(row, name)
	if err != nil {
		return 0, err
	}

	return int64(v), nil
}

func (r *Resultset) GetFloat(row, column int) (float64, error) {
	d, err := r.GetValue(row, column)
	if err != nil {
		return 0, err
	}

	switch v := d.(type) {
	case float64:
		return v, nil
	case uint64:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	case []byte:
		return strconv.ParseFloat(string(v), 64)
	case nil:
		return 0, nil
	default:
		return 0, mysqlerrors.Unknownf("data type is %T", v)
	}
}

func (r *Resultset) GetFloatByName(row int, name string) (float64, error) {
	if column, err := r.NameIndex(name); err != nil {
		return 0, err
	} else {
		return r.GetFloat(row, column)
	}
}

func (r *Resultset) GetString(row, column int) (string, error) {
	d, err := r.GetValue(row, column)
	if err != nil {
		return "", err
	}

	switch v := d.(type) {
	case string:
		return v, nil
	case []byte:
		return String(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case nil:
		return "", nil
	default:
		return "", mysqlerrors.Unknownf("data type is %T", v)
	}
}

func (r *Resultset) GetStringByName(row int, name string) (string, error) {
	if column, err := r.NameIndex(name); err != nil {
		return "", err
	} else {
		return r.GetString(row, column)
	}
}

type RowData []byte

func (p RowData) Parse(f []*Field, binary bool) ([]interface{}, error) {
	if binary {
		return p.ParseBinary(f)
	} else {
		return p.ParseText(f)
	}
}

func (p RowData) ParseText(f []*Field) ([]interface{}, error) {
	data := make([]interface{}, len(f))

	var err error
	var v []byte
	var isNull, isUnsigned bool
	var pos int = 0
	var n int = 0

	for i := range f {
		v, isNull, n, err = lengthEncodedString(p[pos:])
		if err != nil {
			return nil, err
		}

		pos += n

		if isNull {
			data[i] = nil
		} else {
			isUnsigned = (f[i].Flag&UNSIGNED_FLAG > 0)

			switch f[i].Type {
			case MYSQL_TYPE_TINY, MYSQL_TYPE_SHORT, MYSQL_TYPE_INT24,
				MYSQL_TYPE_LONGLONG, MYSQL_TYPE_YEAR:
				if isUnsigned {
					data[i], err = strconv.ParseUint(string(v), 10, 64)
				} else {
					data[i], err = strconv.ParseInt(string(v), 10, 64)
				}
			case MYSQL_TYPE_FLOAT, MYSQL_TYPE_DOUBLE:
				data[i], err = strconv.ParseFloat(string(v), 64)
			default:
				data[i] = v
			}

			if err != nil {
				return nil, err
			}
		}
	}

	return data, nil
}

func (p RowData) ParseBinary(f []*Field) ([]interface{}, error) {
	data := make([]interface{}, len(f))

	if p[0] != OK_HEADER {
		return nil, errMalformPacket
	}

	pos := 1 + ((len(f) + 7 + 2) >> 3)

	nullBitmap := p[1:pos]

	var isUnsigned bool
	var isNull bool
	var n int
	var err error
	var v []byte
	for i := range data {
		if nullBitmap[(i+2)/8]&(1<<(uint(i+2)%8)) > 0 {
			data[i] = nil
			continue
		}

		isUnsigned = f[i].Flag&UNSIGNED_FLAG > 0

		switch f[i].Type {
		case MYSQL_TYPE_NULL:
			data[i] = nil
			continue

		case MYSQL_TYPE_TINY:
			if isUnsigned {
				data[i] = uint64(p[pos])
			} else {
				data[i] = int64(p[pos])
			}
			pos++
			continue

		case MYSQL_TYPE_SHORT, MYSQL_TYPE_YEAR:
			if isUnsigned {
				data[i] = uint64(binary.LittleEndian.Uint16(p[pos : pos+2]))
			} else {
				data[i] = int64((binary.LittleEndian.Uint16(p[pos : pos+2])))
			}
			pos += 2
			continue

		case MYSQL_TYPE_INT24, MYSQL_TYPE_LONG:
			if isUnsigned {
				data[i] = uint64(binary.LittleEndian.Uint32(p[pos : pos+4]))
			} else {
				data[i] = int64(binary.LittleEndian.Uint32(p[pos : pos+4]))
			}
			pos += 4
			continue

		case MYSQL_TYPE_LONGLONG:
			if isUnsigned {
				data[i] = binary.LittleEndian.Uint64(p[pos : pos+8])
			} else {
				data[i] = int64(binary.LittleEndian.Uint64(p[pos : pos+8]))
			}
			pos += 8
			continue

		case MYSQL_TYPE_FLOAT:
			data[i] = float64(math.Float32frombits(binary.LittleEndian.Uint32(p[pos : pos+4])))
			pos += 4
			continue

		case MYSQL_TYPE_DOUBLE:
			data[i] = math.Float64frombits(binary.LittleEndian.Uint64(p[pos : pos+8]))
			pos += 8
			continue

		case MYSQL_TYPE_DECIMAL, MYSQL_TYPE_NEWDECIMAL, MYSQL_TYPE_VARCHAR,
			MYSQL_TYPE_BIT, MYSQL_TYPE_ENUM, MYSQL_TYPE_SET, MYSQL_TYPE_TINY_BLOB,
			MYSQL_TYPE_MEDIUM_BLOB, MYSQL_TYPE_LONG_BLOB, MYSQL_TYPE_BLOB,
			MYSQL_TYPE_VAR_STRING, MYSQL_TYPE_STRING, MYSQL_TYPE_GEOMETRY:
			v, isNull, n, err = lengthEncodedString(p[pos:])
			pos += n
			if err != nil {
				return nil, err
			}

			if !isNull {
				data[i] = v
				continue
			} else {
				data[i] = nil
				continue
			}
		case MYSQL_TYPE_DATE, MYSQL_TYPE_NEWDATE:
			var num uint64
			num, isNull, n = lengthEncodedInt(p[pos:])

			pos += n

			if isNull {
				data[i] = nil
				continue
			}

			data[i], err = formatBinaryDate(int(num), p[pos:])
			pos += int(num)

			if err != nil {
				return nil, err
			}

		case MYSQL_TYPE_TIMESTAMP, MYSQL_TYPE_DATETIME:
			var num uint64
			num, isNull, n = lengthEncodedInt(p[pos:])

			pos += n

			if isNull {
				data[i] = nil
				continue
			}

			data[i], err = formatBinaryDateTime(int(num), p[pos:])
			pos += int(num)

			if err != nil {
				return nil, err
			}

		case MYSQL_TYPE_TIME:
			var num uint64
			num, isNull, n = lengthEncodedInt(p[pos:])

			pos += n

			if isNull {
				data[i] = nil
				continue
			}

			data[i], err = formatBinaryTime(int(num), p[pos:])
			pos += int(num)

			if err != nil {
				return nil, err
			}

		default:
			return nil, mysqlerrors.Unknownf("unknown FieldType %d %s", f[i].Type, f[i].Name)
		}
	}

	return data, nil
}
