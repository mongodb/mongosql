package json

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// JSONDateFormat is the constant for formatting dates in a JSON document.
const JSONDateFormat = "2006-01-02T15:04:05.000Z"

// MarshalJSON marshals BinData b to bytes.
func (b BinData) MarshalJSON() ([]byte, error) {
	data := fmt.Sprintf(`{ "$binary": "%v", "$type": "%0x" }`,
		b.Base64, []byte{b.Type})
	return []byte(data), nil
}

// MarshalJSON marshals Decimal128 d128 to bytes.
func (d128 Decimal128) MarshalJSON() ([]byte, error) {
	s := d128.Value.String()
	return []byte(fmt.Sprintf(`{ "$numberDecimal" : "%s" }`, s)), nil
}

// MarshalJSON marshals JavaScript js to bytes.
func (js JavaScript) MarshalJSON() ([]byte, error) {
	data := []byte(fmt.Sprintf(`{ "$code": %q`, js.Code))

	scopeChunk := []byte{}
	var err error
	if js.Scope != nil {
		scopeChunk, err = Marshal(js.Scope)
		if err != nil {
			return nil, err
		}
		scopeChunk = []byte(fmt.Sprintf(`, "$scope": %v `, string(scopeChunk)))
	}
	scopeChunk = append(scopeChunk, '}')

	data = append(data, scopeChunk...)
	return data, nil
}

// MarshalJSON marshals Date d to bytes.
func (d Date) MarshalJSON() ([]byte, error) {
	var data string
	n := int64(d)
	if d.isFormatable() {
		t := time.Unix(n/1e3, n%1e3*1e6)
		data = fmt.Sprintf(`{ "$date": "%v" }`, t.UTC().Format(JSONDateFormat))
	} else {
		data = fmt.Sprintf(`{ "$date": { "$numberLong" : "%v" }}`, n)
	}

	return []byte(data), nil
}

// MarshalJSON marshals DBRef d to bytes.
func (d DBRef) MarshalJSON() ([]byte, error) {
	// Convert the $id field to JSON
	idChunk, err := Marshal(d.ID)
	if err != nil {
		return nil, err
	}

	// Need to form JSON like { "$ref": "REF", "$id": ID, "$db": "DB" }
	// so piece chunks together since can only get $id field as bytes.
	refChunk := []byte(fmt.Sprintf(`{ "$ref": "%v", "$id": `, d.Collection))

	dbChunk := []byte{}
	if d.Database != "" {
		dbChunk = []byte(fmt.Sprintf(`, "$db": "%v" `, d.Database))
	}
	dbChunk = append(dbChunk, '}')

	data := make([]byte, len(refChunk)+len(idChunk)+len(dbChunk))
	copy(data, refChunk)
	copy(data[len(refChunk):], idChunk)
	copy(data[len(refChunk)+len(idChunk):], dbChunk)

	return data, nil
}

// MarshalJSON marshals DBPointer d to bytes.
func (d DBPointer) MarshalJSON() ([]byte, error) {
	buffer := bytes.Buffer{}
	// Convert the $id field to JSON
	idChunk, err := Marshal(d.ID)
	if err != nil {
		return nil, err
	}
	buffer.Write([]byte(fmt.Sprintf(`{ "$ref": "%v", "$id": { "$oid" : `, d.Namespace)))
	buffer.Write(idChunk)
	buffer.Write([]byte("}}"))
	return buffer.Bytes(), nil
}

// MarshalJSON marshals a MinKey to bytes.
func (MinKey) MarshalJSON() ([]byte, error) {
	data := `{ "$minKey": 1 }`
	return []byte(data), nil
}

// MarshalJSON marshals a MaxKey to bytes.
func (MaxKey) MarshalJSON() ([]byte, error) {
	data := `{ "$maxKey": 1 }`
	return []byte(data), nil
}

// MarshalJSON marshals NumberInt n to bytes.
func (n NumberInt) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(n), 10)), nil
}

// MarshalJSON marshals NumberLong n to bytes.
func (n NumberLong) MarshalJSON() ([]byte, error) {
	data := fmt.Sprintf(`{ "$numberLong": "%v" }`, int64(n))
	return []byte(data), nil
}

// MarshalJSON marshals NumberFloat n to bytes.
func (n NumberFloat) MarshalJSON() ([]byte, error) {
	// check floats for infinity and return +Infinity or -Infinity if so
	if math.IsInf(float64(n), 1) {
		return []byte("+Infinity"), nil
	}
	if math.IsInf(float64(n), -1) {
		return []byte("-Infinity"), nil
	}

	floatString := strconv.FormatFloat(float64(n), 'g', -1, 64)

	// determine if the float has a decimal point and if not
	// add one to maintain consistency when importing.
	if _, d := math.Modf(float64(n)); d == 0 {
		// check for 'e' to determine if the float's formatted in scientific notation
		if strings.IndexByte(floatString, 'e') == -1 {
			return []byte(floatString + ".0"), nil
		}

	}
	return []byte(floatString), nil
}

// MarshalJSON assumes that o represents a valid ObjectId
// (composed of 24 hexadecimal characters).
func (o ObjectID) MarshalJSON() ([]byte, error) {
	data := fmt.Sprintf(`{ "$oid": "%v" }`, string(o))
	return []byte(data), nil
}

// MarshalJSON marshals RegExp r to bytes.
func (r RegExp) MarshalJSON() ([]byte, error) {
	pattern, err := Marshal(r.Pattern)
	if err != nil {
		return nil, err
	}
	data := fmt.Sprintf(`{ "$regex": %v, "$options": "%v" }`,
		string(pattern), r.Options)
	return []byte(data), nil
}

// MarshalJSON marshals Timestamp t to bytes.
func (t Timestamp) MarshalJSON() ([]byte, error) {
	data := fmt.Sprintf(`{ "$timestamp": { "t": %v, "i": %v } }`,
		t.Seconds, t.Increment)
	return []byte(data), nil
}

// MarshalJSON marshals and Undefined to bytes.
func (Undefined) MarshalJSON() ([]byte, error) {
	data := `{ "$undefined": true }`
	return []byte(data), nil
}
