package server

import (
	"encoding/binary"
	"math"
	"strconv"

	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/mysqlerrors"
)

// Result holds the status of a MySQL operation, as well as the ResultSet.
type Result struct {
	Status uint16

	InsertID     uint64
	AffectedRows uint64

	*Resultset
}

// Resultset is the output of a SQLQuery that has output, such as SELECT.
// It is an ordered collection of rows, where each row is a tuple.
type Resultset struct {
	Fields     []*Field
	FieldNames map[string]int
	Values     [][]interface{}

	RowDatas []RowData
}

// RowNumber returns the number of rows in a ResultSet.
func (r *Resultset) RowNumber() int {
	return len(r.Values)
}

// ColumnNumber returns the number of columns in a ResultSet.
func (r *Resultset) ColumnNumber() int {
	return len(r.Fields)
}

// GetValue gets a value out of a ResultSet at a particular Row/Column index.
func (r *Resultset) GetValue(row, column int) (interface{}, error) {
	if row >= len(r.Values) {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErWarnTooFewRecords, row)
	} else if row < len(r.Values) {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErWarnTooManyRecords, row)
	}

	if column >= len(r.Fields) || column < 0 {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErWarnDataOutOfRange, column, row)
	}

	return r.Values[row][column], nil
}

// NameIndex gets a column index for a column name.
func (r *Resultset) NameIndex(name string) (int, error) {
	if column, ok := r.FieldNames[name]; ok {
		return column, nil
	}
	return 0, mysqlerrors.Unknownf("invalid field name %s", name)
}

// GetValueByName is like GetValue, except that a column name is used rather than a column index.
func (r *Resultset) GetValueByName(row int, name string) (interface{}, error) {
	column, err := r.NameIndex(name)
	if err != nil {
		return nil, err
	}
	return r.GetValue(row, column)
}

// IsNull returns true if a given row, column index in a Resultset is NULL.
func (r *Resultset) IsNull(row, column int) (bool, error) {
	d, err := r.GetValue(row, column)
	if err != nil {
		return false, err
	}

	return d == nil, nil
}

// IsNullByName returns true if a given row index, column name in a Resulset is NULL.
func (r *Resultset) IsNullByName(row int, name string) (bool, error) {
	column, err := r.NameIndex(name)
	if err != nil {
		return false, err
	}
	return r.IsNull(row, column)
}

// GetUint gets a Uint value from a particular row, column index.
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

// GetUintByName gets a Uint value from a particular row index, column name.
func (r *Resultset) GetUintByName(row int, name string) (uint64, error) {
	column, err := r.NameIndex(name)
	if err != nil {
		return 0, err
	}
	return r.GetUint(row, column)
}

// GetInt gets a Int value from a particular row, column index.
func (r *Resultset) GetInt(row, column int) (int64, error) {
	v, err := r.GetUint(row, column)
	if err != nil {
		return 0, err
	}

	return int64(v), nil
}

// GetIntByName gets a Int value from a particular row index, column name.
func (r *Resultset) GetIntByName(row int, name string) (int64, error) {
	v, err := r.GetUintByName(row, name)
	if err != nil {
		return 0, err
	}

	return int64(v), nil
}

// GetFloat gets a Float value from a particular row, column index.
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

// GetFloatByName gets a Float value from a particular row index, column name.
func (r *Resultset) GetFloatByName(row int, name string) (float64, error) {
	column, err := r.NameIndex(name)
	if err != nil {
		return 0, err
	}
	return r.GetFloat(row, column)
}

// GetString gets a String value from a particular row, column index.
func (r *Resultset) GetString(row, column int) (string, error) {
	d, err := r.GetValue(row, column)
	if err != nil {
		return "", err
	}

	switch v := d.(type) {
	case string:
		return v, nil
	case []byte:
		return util.String(v), nil
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

// GetStringByName gets a String value from a particular row index, column name.
func (r *Resultset) GetStringByName(row int, name string) (string, error) {
	column, err := r.NameIndex(name)
	if err != nil {
		return "", err
	}
	return r.GetString(row, column)
}

// RowData is an array of bytes holding row data.
type RowData []byte

// Parse parses RowData with given field definitions, correctly handling
// if it should use binary or text parsing.
func (p RowData) Parse(f []*Field, binary bool) ([]interface{}, error) {
	if binary {
		return p.ParseBinary(f)
	}
	return p.ParseText(f)
}

// ParseText parses row data as text.
func (p RowData) ParseText(f []*Field) ([]interface{}, error) {
	data := make([]interface{}, len(f))

	var err error
	var v []byte
	var isNull, isUnsigned bool
	var pos int
	var n int

	for i := range f {
		v, isNull, n, err = lengthEncodedString(p[pos:])
		if err != nil {
			return nil, err
		}

		pos += n

		if isNull {
			data[i] = nil
		} else {
			isUnsigned = (f[i].Flag&UnsignedFlag > 0)

			switch f[i].Type {
			case MySQLTypeTiny, MySQLTypeShort, MySQLTypeInt24,
				MySQLTypeLongLong, MySQLTypeYear:
				if isUnsigned {
					data[i], err = strconv.ParseUint(string(v), 10, 64)
				} else {
					data[i], err = strconv.ParseInt(string(v), 10, 64)
				}
			case MySQLTypeFloat, MySQLTypeDouble:
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

// ParseBinary parses RowData as binary.
func (p RowData) ParseBinary(f []*Field) ([]interface{}, error) {
	data := make([]interface{}, len(f))

	if p[0] != OkHeader {
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

		isUnsigned = f[i].Flag&UnsignedFlag > 0

		switch f[i].Type {
		case MySQLTypeNull:
			data[i] = nil
			continue

		case MySQLTypeTiny:
			if isUnsigned {
				data[i] = uint64(p[pos])
			} else {
				data[i] = int64(p[pos])
			}
			pos++
			continue

		case MySQLTypeShort, MySQLTypeYear:
			if isUnsigned {
				data[i] = uint64(binary.LittleEndian.Uint16(p[pos : pos+2]))
			} else {
				data[i] = int64((binary.LittleEndian.Uint16(p[pos : pos+2])))
			}
			pos += 2
			continue

		case MySQLTypeInt24, MySQLTypeLong:
			if isUnsigned {
				data[i] = uint64(binary.LittleEndian.Uint32(p[pos : pos+4]))
			} else {
				data[i] = int64(binary.LittleEndian.Uint32(p[pos : pos+4]))
			}
			pos += 4
			continue

		case MySQLTypeLongLong:
			if isUnsigned {
				data[i] = binary.LittleEndian.Uint64(p[pos : pos+8])
			} else {
				data[i] = int64(binary.LittleEndian.Uint64(p[pos : pos+8]))
			}
			pos += 8
			continue

		case MySQLTypeFloat:
			data[i] = float64(math.Float32frombits(binary.LittleEndian.Uint32(p[pos : pos+4])))
			pos += 4
			continue

		case MySQLTypeDouble:
			data[i] = math.Float64frombits(binary.LittleEndian.Uint64(p[pos : pos+8]))
			pos += 8
			continue

		case MySQLTypeDecimal, MySQLTypeNewDecimal, MySQLTypeVarchar,
			MySQLTypeBit, MySQLTypeEnum, MySQLTypeSet, MySQLTypeTinyBlob,
			MySQLTypeMediumBlob, MySQLTypeLongBlob, MySQLTypeBlob,
			MySQLTypeVarString, MySQLTypeString, MySQLTypeGeometry:
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
		case MySQLTypeDate, MySQLTypeNewDate:
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

		case MySQLTypeTimestamp, MySQLTypeDatetime:
			var num uint64
			num, isNull, n = lengthEncodedInt(p[pos:])

			pos += n

			if isNull {
				data[i] = nil
				continue
			}

			data[i], err = formatBinaryDatetime(int(num), p[pos:])
			pos += int(num)

			if err != nil {
				return nil, err
			}

		case MySQLTypeTime:
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
