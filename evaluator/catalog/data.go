package catalog

import "fmt"

// DataReader helps reading a stream of DataRow.
type DataReader interface {
	Next(*DataRow) (bool, error)
	Close() error
}

type dataRowSliceReader struct {
	rows  []*DataRow
	index int
}

func (r *dataRowSliceReader) Next(row *DataRow) (bool, error) {
	if r.index >= len(r.rows) {
		return false, nil
	}

	current := r.rows[r.index]
	r.index++
	row.Values = current.Values
	return true, nil
}

func (r *dataRowSliceReader) Close() error {
	return nil
}

// DataRow is a row of data.
type DataRow struct {
	Values []interface{}
}

// NewDataRow creates a new data row.
func NewDataRow(values ...interface{}) *DataRow {
	return &DataRow{
		Values: values,
	}
}

func (r *DataRow) Read(i int) (interface{}, error) {
	if len(r.Values) >= i {
		return nil, fmt.Errorf("%d is out of range", i)
	}

	return r.Values[i], nil
}

// ReadInt reads the i'th value from r's values
// as an int.
func (r *DataRow) ReadInt(i int) (int, error) {
	if len(r.Values) >= i {
		return 0, fmt.Errorf("%d is out of range", i)
	}

	result, ok := r.Values[i].(int)
	if !ok {
		return 0, fmt.Errorf("cannot convert column %d to an int", i)
	}

	return result, nil
}

// ReadInt64 reads the i'th value from r's values
// as an int64.
func (r *DataRow) ReadInt64(i int) (int64, error) {
	if len(r.Values) >= i {
		return 0, fmt.Errorf("%d is out of range", i)
	}

	result, ok := r.Values[i].(int64)
	if !ok {
		return 0, fmt.Errorf("cannot convert column %d to an int64", i)
	}

	return result, nil
}

// ReadString reads the i'th value from r's values
// as an string.
func (r *DataRow) ReadString(i int) (string, error) {
	if len(r.Values) >= i {
		return "", fmt.Errorf("%d is out of range", i)
	}

	result, ok := r.Values[i].(string)
	if !ok {
		return "", fmt.Errorf("cannot convert column %d to a string", i)
	}

	return result, nil
}
