// Package bsonutil provides utilities for processing BSON data.
package bsonutil

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/mongodb/mongo-tools-common/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// DeepCopyDSlice performs a deep copy of the given bson.D slice.
func DeepCopyDSlice(src []bson.D) []bson.D {
	if src == nil {
		// This makes testing easier: in some places we have nil pipelines
		// whereas in others we have empty pipelines. They are semantically
		// equivalent, so we merge here.
		return NewDArray()
	}
	ret := make([]bson.D, len(src))
	for i := range src {
		ret[i] = deepCopyD(src[i])
	}
	return ret
}

func deepCopyD(src bson.D) bson.D {
	ret := make([]bson.E, len(src))
	for i := range src {
		// DocElem.Value is set to nil as placeholder.
		val := NewDocElem(src[i].Key, nil)
		switch typedElement := src[i].Value.(type) {
		case bson.D:
			val.Value = deepCopyD(typedElement)
		case bson.M:
			val.Value = deepCopyM(typedElement)
		case bson.A:
			val.Value = deepCopyA(typedElement)
		default:
			val.Value = src[i].Value
		}
		ret[i] = val
	}

	return NewD(ret...)
}

func deepCopyM(src bson.M) bson.M {
	ret := make(bson.M, len(src))
	for k, v := range src {
		var val interface{}
		switch typedElement := v.(type) {
		case bson.D:
			val = deepCopyD(typedElement)
		case bson.M:
			val = deepCopyM(typedElement)
		case bson.A:
			val = deepCopyA(typedElement)
		default:
			val = v
		}
		ret[k] = val
	}
	return ret
}

func deepCopyA(src bson.A) interface{} {
	ret := make(bson.A, len(src))
	for i := range src {
		var val interface{}
		switch typedElement := src[i].(type) {
		case bson.D:
			val = deepCopyD(typedElement)
		case bson.M:
			val = deepCopyM(typedElement)
		case []interface{}:
			val = deepCopyA(typedElement)
		default:
			val = src[i]
		}
		ret[i] = val
	}
	return ret
}

// DocSliceToCoreArray converts a slice of bson.Ds into a bsoncore.Array.
func DocSliceToCoreArray(docs []bson.D) (bsoncore.Array, error) {
	aidx, arr := bsoncore.AppendArrayStart(nil)

	for i, doc := range docs {
		docBytes, err := bson.Marshal(&doc)
		if err != nil {
			return nil, err
		}
		arr = bsoncore.AppendDocumentElement(arr, strconv.Itoa(i), docBytes)
	}

	return bsoncore.AppendArrayEnd(arr, aidx)
}

// DocSliceToString turns a slice of bson.Ds into an extended JSON string.
func DocSliceToString(docs []bson.D) (string, error) {
	b := []byte{'['}

	for _, doc := range docs {
		dBytes, err := bson.MarshalExtJSON(doc, false, false)
		if err != nil {
			return "", err
		}

		b = append(b, dBytes...)
		b = append(b, ',')
	}

	// replace the last comma with a closing bracket
	b[len(b)-1] = ']'

	return string(b), nil
}

// NormalizeBSON converts bson.Ds that represent extended JSON types
// into primitive bson types and converts []interface{} into bson.A.
//
// For example, the following bson.D actually represents an extended
// JSON type:
//
//     bson.D{
//       bson.E{
//         Key: "$timestamp",
//         Value: bson.D{
//           bson.E{Key: "t", Value: 42},
//           bson.E{Key: "i", Value: 1},
//         },
//       },
//     }
//
// which is normalized to:
//     primitive.Timestamp{T: 42, I: 1}.
//
// This function is necessary because, as bson.Ds, such values are
// incorrectly typed. The bson.D example above will not be treated
// as a primitive.Timestamp by the Go driver; we need to change it
// to an actual primitive.Timestamp for the Go driver to recognize
// it as one.
//
// The docs argued to this function are read from yaml files which
// are parsed by yaml parsers that only know how to create bson.Ds
// based on the bson.D type:
//     type bson.D struct {
//       Key string
//       Value interface{}
//     }
//
// Therefore, yaml text such as:
//     // (newlines added for readability)
//     docs:
//       - { "_id": {"$oid": "57e193d7a9cc81b4027498b5"},
//           "a": {"$numberInt": "123"},
//           "b": {"$date": "2019-06-28T00:00:00Z"}
//         }
//       - ...
//
// will simply become a slice of bson.Ds with nested bson.Ds:
//     docs = []bson.D{
//       bson.D{
//         bson.E{Key: "_id", Value: bson.D{bson.E{Key: "$oid", Value: "57e193d7a9cc81b4027498b5"}}},
//         bson.E{Key: "a", Value: bson.D{bson.E{Key: "$numberInt", Value: "123"}}},
//         bson.E{Key: "b", Value: bson.D{bson.E{Key: "$date", Value: "2019-06-28T00:00:00Z"}}},
//       },
//       ...
//     }
//
// This function will change that into a slice of bson.Ds with the
// expected primitive bson types:
//     docs = []bson.D{
//       bson.E{Key: "_id", Value: primitive.ObjectID{0x57, 0xe1, 0x93, 0xd7, 0xa9, 0xcc, 0x81, 0xb4, 0x2, 0x74, 0x98, 0xb5}},
//       bson.E{Key: "a", Value: int32(123)},
//       bson.E{Key: "b", Value: primitive.DateTime(1561680000000)},
//     }
//
// Note that NormalizeBSON is intended for use with bson.Ds that
// contain nested extended JSON types. As in, the slice of docs
// must not represent extended JSON at the top-level:
//    valid:   []bson.D{
//               bson.D{bson.E{Key: "a": Value: bson.D{bson.E{Key: "$oid", Value: "57e193d7a9cc81b4027498b5"}}}},
//               bson.D{bson.E{Key: "b": Value: bson.D{bson.E{Key: "$numberDouble", Value: "-Infinity"}}}},
//             }
//    invalid: []bson.D{
//               bson.D{bson.E{Key: "$oid", Value: "57e193d7a9cc81b4027498b5"}},
//               bson.D{bson.E{Key: "$numberDouble", Value: "-Infinity"}},
//             }
//
// Extended JSON parsing is optional via the convertExtJSON argument.
func NormalizeBSON(docs []bson.D, convertExtJSON bool) ([]bson.D, error) {
	newDocs := make([]bson.D, len(docs))

	for i, doc := range docs {
		newDoc, err := normalize(doc, convertExtJSON)
		if err != nil {
			return nil, fmt.Errorf("error normalizing extended json into bson: %v", err)
		}

		newDocs[i] = newDoc.(bson.D)
	}

	return newDocs, nil
}

// normalize handles converting []interface{} into bson.A, and also
// ensures the elements of those slices are normalized. It delegates
// parsing bson.Ds for extended JSON to another function, if needed.
func normalize(in interface{}, convertExtJSON bool) (interface{}, error) {
	var err error

	switch t := in.(type) {
	case bson.D:
		if convertExtJSON {
			return ParseExtendedJSON(t)
		}

		var v interface{}
		newD := make(bson.D, len(t))
		for i, e := range t {
			v, err = normalize(e.Value, convertExtJSON)
			if err != nil {
				return nil, err
			}
			newD[i] = bson.E{Key: e.Key, Value: v}
		}
		return newD, nil

	case []interface{}:
		newA := make(bson.A, len(t))
		for i, e := range t {
			newA[i], err = normalize(e, convertExtJSON)
			if err != nil {
				return nil, err
			}
		}
		return newA, nil

	case bson.A:
		newA := make(bson.A, len(t))
		for i, e := range t {
			newA[i], err = normalize(e, convertExtJSON)
			if err != nil {
				return nil, err
			}
		}
		return newA, nil
	}

	return in, nil
}

// ParseExtendedJSON takes bson.D and inspects it for any extended
// JSON types (e.g. $numberLong). It replaces any such values with
// the corresponding primitive BSON types. For example:
//
//     bson.D{bson.E{Key: "$code", Value: "function() {}"}}
//
// is recognized as extended JSON and parsed into:
//
//     primitive.JavaScript("function() {}")
//
// This function treats the Datetime type as a special case. It is
// more tolerant than the extended JSON spec allows because we may
// have DRDL files in the wild that contain non-canonical extended
// JSON dates.
//
// Strings, booleans, nil, and relaxed int32s, int64s, and doubles
// are left unchanged. Non-extended JSON documents are recursively
// parsed for nested extended JSON.
func ParseExtendedJSON(d bson.D) (interface{}, error) {
	length := len(d)

	// Special case $date: allow for a string representation
	// of a date, as well as relaxed number types.
	if length == 1 && d[0].Key == "$date" {
		switch v := d[0].Value.(type) {
		case string:
			t, err := util.FormatDate(v)
			if err != nil {
				return nil, err
			}
			return primitive.NewDateTimeFromTime(t.(time.Time)), nil
		case int32:
			return primitive.DateTime(v), nil
		case int64:
			return primitive.DateTime(v), nil
		case int:
			return primitive.DateTime(v), nil
		case float64:
			return primitive.DateTime(v), nil

		// {"$date": {"$numberLong": "..."}} is the
		// canonical extended JSON for a Datetime.
		case bson.D:
			dateMap := v.Map()
			if l, ok := dateMap["$numberLong"]; ok && len(dateMap) == 1 {
				switch v := l.(type) {
				case string:
					// all of decimal, hex, and octal are supported here
					n, err := strconv.ParseInt(v, 0, 64)
					if err != nil {
						return nil, err
					}
					return primitive.DateTime(n), nil

				default:
					return 0, errors.New("expected $numberLong field to have string value")
				}
			}

		default:
			return nil, errors.New("invalid type for $date field")
		}
	}

	// If the provided bson.D is actually an extended JSON document,
	// we will use the Go driver's bson.UnmarshalExtJSON to get the
	// appropriate primitive bson value. To do this, we marshal the
	// document here, once, getting it as a slice of bytes.
	docBytes, err := bson.MarshalExtJSON(d, false, false)
	if err != nil {
		panic(err)
	}

	// unmarshal is used below to unmarshal the marshaled
	// bytes into the appropriate primitive bson type.
	unmarshal := func(i interface{}) (interface{}, error) {
		err = bson.UnmarshalExtJSON(docBytes, false, &i)
		if err != nil {
			return nil, err
		}
		return i, nil
	}

	switch length {
	case 1: // document has a single field
		switch d[0].Key {
		case "$oid":
			var v primitive.ObjectID
			return unmarshal(v)

		case "$symbol":
			var v primitive.Symbol
			return unmarshal(v)

		case "$numberInt":
			var v int32
			return unmarshal(v)

		case "$numberLong":
			var v int64
			return unmarshal(v)

		case "$numberDouble":
			var v float64
			return unmarshal(v)

		case "$numberDecimal":
			var v primitive.Decimal128
			return unmarshal(v)

		case "$binary":
			var v primitive.Binary
			return unmarshal(v)

		case "$code":
			var v primitive.JavaScript
			return unmarshal(v)

		case "$timestamp":
			var v primitive.Timestamp
			return unmarshal(v)

		case "$regularExpression":
			var v primitive.Regex
			return unmarshal(v)

		case "$dbPointer":
			var v primitive.DBPointer
			return unmarshal(v)

		case "$minKey":
			return primitive.MinKey{}, nil

		case "$maxKey":
			return primitive.MaxKey{}, nil

		case "$undefined":
			return primitive.Undefined{}, nil
		}

	case 2: // document has two fields
		doc := d.Map()
		if _, ok := doc["$code"]; ok {
			var v primitive.CodeWithScope
			return unmarshal(v)
		}

		// Support legacy binary type
		if binaryVal, ok := doc["$binary"]; ok {
			if typeVal, ok := doc["$type"]; ok {
				bin := primitive.Binary{}
				switch v := binaryVal.(type) {
				case string:
					data, err := base64.StdEncoding.DecodeString(v)
					if err != nil {
						return nil, err
					}
					bin.Data = data
				default:
					return nil, errors.New("expected $binary field to have string value")
				}

				switch v := typeVal.(type) {
				case string:
					kind, err := hex.DecodeString(v)
					if err != nil {
						return nil, err
					} else if len(kind) != 1 {
						err := errors.New("expected single byte (as hex string) for $type field")
						return nil, err
					}
					bin.Subtype = kind[0]
				default:
					return nil, errors.New("expected $type field to have string value")
				}
				return bin, nil
			}
			return nil, errors.New("legacy Binary bson type requires exactly $binary and $type keys")
		}

		// Support legeacy regex type
		if regexVal, ok := doc["$regex"]; ok {
			if optionsVal, ok := doc["$options"]; ok {
				regex := primitive.Regex{}
				switch v := regexVal.(type) {
				case string:
					regex.Pattern = v
				default:
					return nil, errors.New("expected $regex field to have string value")
				}

				switch v := optionsVal.(type) {
				case string:
					regex.Options = v
				default:
					return nil, errors.New("expected $options field to have string value")
				}
				return regex, nil
			}
			return nil, errors.New("legacy Regex bson type requires exactly $regex and $options keys")
		}

		//// Support legacy DBPointer type
		if refVal, ok := doc["$ref"]; ok {
			if idVal, ok := doc["$id"]; ok {
				dbp := primitive.DBPointer{}
				switch v := refVal.(type) {
				case string:
					dbp.DB = v
				default:
					return nil, errors.New("expected $ref field to have string value")
				}

				var oidHex string
				switch v := idVal.(type) {
				case string:
					oidHex = v
				case bson.D:
					if oidVal, ok := v.Map()["$oid"]; ok && len(v) == 1 {
						switch v := oidVal.(type) {
						case string:
							oidHex = v
						default:
							return nil, errors.New("expected $oid field to have string value")
						}
					} else {
						return nil, errors.New("expected $id sub-document to have $oid field")
					}
				default:
					return nil, errors.New("expected $id field to have string or ObjectID value")
				}

				oid, err := primitive.ObjectIDFromHex(oidHex)
				if err != nil {
					return nil, err
				}
				dbp.Pointer = oid

				return dbp, nil
			}
			return nil, errors.New("legacy DBPointer bson type requires exactly $ref and $id keys")
		}
	}

	// If it was not an extended JSON type, check for nested
	// extended JSON types.
	newD := make(bson.D, len(d))
	for i, e := range d {
		v, err := normalize(e.Value, true)
		if err != nil {
			return nil, err
		}
		newD[i] = bson.E{Key: e.Key, Value: v}
	}

	return newD, nil
}
