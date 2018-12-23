package bsonutil

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/10gen/sqlproxy/internal/json"
	"github.com/10gen/sqlproxy/internal/util"

	"github.com/10gen/mongo-go-driver/bson"
)

// ConvertJSONValueToBSON walks through a document or an array and
// replaces any extended JSON value with its corresponding BSON type.
func ConvertJSONValueToBSON(x interface{}) (interface{}, error) {
	switch v := x.(type) {
	case nil:
		return nil, nil
	case bool:
		return v, nil
	case map[string]interface{}: // document
		for key, jsonValue := range v {
			bsonValue, err := ParseJSONValue(jsonValue)
			if err != nil {
				return nil, err
			}
			v[key] = bsonValue
		}
		return v, nil
	case bson.M:
		for key, jsonValue := range v {
			bsonValue, err := ParseJSONValue(jsonValue)
			if err != nil {
				return nil, err
			}
			v[key] = bsonValue
		}
		return v, nil
	case bson.D:
		for i := range v {
			var err error
			v[i].Value, err = ParseJSONValue(v[i].Value)
			if err != nil {
				return nil, err
			}
		}
		return v, nil

	case []interface{}: // array
		for i, jsonValue := range v {
			bsonValue, err := ParseJSONValue(jsonValue)
			if err != nil {
				return nil, err
			}
			v[i] = bsonValue
		}
		return v, nil

	case string, float64, int32, int64:
		return v, nil // require no conversion

	case json.ObjectID: // ObjectId
		s := string(v)
		if !bson.IsObjectIdHex(s) {
			return nil, errors.New("expected ObjectId to contain 24 hexadecimal characters")
		}
		return bson.ObjectIdHex(s), nil

	case json.Decimal128:
		return v.Value, nil

	case json.Date: // Date
		n := int64(v)
		return time.Unix(n/1e3, n%1e3*1e6), nil

	case json.ISODate: // ISODate
		n := string(v)
		return util.FormatDate(n)

	case json.NumberLong: // NumberLong
		return int64(v), nil

	case json.NumberInt: // NumberInt
		return int32(v), nil

	case json.NumberFloat: // NumberFloat
		return float64(v), nil

	case json.NumberUint32: // NumberUint32
		return uint32(v), nil

	case json.NumberUint64: // NumberUint64
		return uint64(v), nil

	case json.BinData: // BinData
		data, err := base64.StdEncoding.DecodeString(v.Base64)
		if err != nil {
			return nil, err
		}
		return bson.Binary{Kind: v.Type, Data: data}, nil

	case json.RegExp: // RegExp
		return bson.RegEx{Pattern: v.Pattern, Options: v.Options}, nil

	case json.Timestamp: // Timestamp
		ts := (int64(v.Seconds) << 32) | int64(v.Increment)
		return bson.MongoTimestamp(ts), nil

	case json.JavaScript: // Javascript
		return bson.JavaScript{Code: v.Code, Scope: v.Scope}, nil

	case json.MinKey: // MinKey
		return bson.MinKey, nil

	case json.MaxKey: // MaxKey
		return bson.MaxKey, nil

	case json.Undefined: // undefined
		return bson.Undefined, nil

	default:
		return nil, fmt.Errorf("conversion of JSON value '%v' of type '%T' not supported", v, v)
	}
}

func convertKeys(v bson.M) (bson.M, error) {
	for key, value := range v {
		jsonValue, err := ConvertBSONValueToJSON(value)
		if err != nil {
			return nil, err
		}
		v[key] = jsonValue
	}
	return v, nil
}

func getConvertedKeys(v bson.M, extJSON bool) (bson.M, error) {
	out := NewM()
	for key, value := range v {
		jsonValue, err := GetBSONValueAsJSON(value, extJSON)
		if err != nil {
			return nil, err
		}
		out[key] = jsonValue
	}
	return out, nil
}

// ConvertBSONValueToJSON walks through a document or an array and
// converts any BSON value to its corresponding extended JSON type.
// It returns the converted JSON document and any error encountered.
func ConvertBSONValueToJSON(x interface{}) (interface{}, error) {
	switch v := x.(type) {
	case nil:
		return nil, nil
	case bool:
		return v, nil

	case *bson.M: // document
		doc, err := convertKeys(*v)
		if err != nil {
			return nil, err
		}
		return doc, err
	case bson.M: // document
		return convertKeys(v)
	case map[string]interface{}:
		return convertKeys(v)
	case bson.D:
		for i, value := range v {
			jsonValue, err := ConvertBSONValueToJSON(value.Value)
			if err != nil {
				return nil, err
			}
			v[i].Value = jsonValue
		}
		return MarshalD(v), nil
	case MarshalD:
		return v, nil
	case []interface{}: // array
		for i, value := range v {
			jsonValue, err := ConvertBSONValueToJSON(value)
			if err != nil {
				return nil, err
			}
			v[i] = jsonValue
		}
		return v, nil

	case string:
		return v, nil // require no conversion

	case int:
		return json.NumberInt(v), nil

	case bson.ObjectId: // ObjectId
		return json.ObjectID(v.Hex()), nil

	case bson.Decimal128:
		y, _ := bson.ParseDecimal128(v.String())
		return json.Decimal128{Value: y}, nil

	case time.Time: // Date
		return json.Date(v.Unix()*1000 + int64(v.Nanosecond()/1e6)), nil

	case int64: // NumberLong
		return json.NumberLong(v), nil

	case int32: // NumberInt
		return json.NumberInt(v), nil

	case float64:
		return json.NumberFloat(v), nil

	case float32:
		return json.NumberFloat(float64(v)), nil

	case []byte: // BinData (with generic type)
		data := base64.StdEncoding.EncodeToString(v)
		return json.BinData{Type: 0x00, Base64: data}, nil

	case bson.Binary: // BinData
		data := base64.StdEncoding.EncodeToString(v.Data)
		return json.BinData{Type: v.Kind, Base64: data}, nil

	case bson.RegEx: // RegExp
		return json.RegExp{Pattern: v.Pattern, Options: v.Options}, nil

	case bson.MongoTimestamp: // Timestamp
		timestamp := int64(v)
		return json.Timestamp{
			Seconds:   uint32(timestamp >> 32),
			Increment: uint32(timestamp),
		}, nil

	case bson.JavaScript: // JavaScript
		var scope interface{}
		var err error
		if v.Scope != nil {
			scope, err = ConvertBSONValueToJSON(v.Scope)
			if err != nil {
				return nil, err
			}
		}
		return json.JavaScript{Code: v.Code, Scope: scope}, nil

	default:
		switch x {
		case bson.MinKey: // MinKey
			return json.MinKey{}, nil

		case bson.MaxKey: // MaxKey
			return json.MaxKey{}, nil

		case bson.Undefined: // undefined
			return json.Undefined{}, nil
		}
	}

	return nil, fmt.Errorf("conversion of BSON value '%v' of type '%T' not supported", x, x)
}

// GetBSONValueAsJSON is equivalent to ConvertBSONValueToJSON, but does not mutate its argument.
func GetBSONValueAsJSON(x interface{}, extJSON bool) (interface{}, error) {
	switch v := x.(type) {
	case nil:
		return nil, nil
	case bool:
		return v, nil

	case *bson.M: // document
		doc, err := getConvertedKeys(*v, extJSON)
		if err != nil {
			return nil, err
		}
		return doc, err
	case bson.M: // document
		return getConvertedKeys(v, extJSON)
	case map[string]interface{}:
		return getConvertedKeys(v, extJSON)
	case bson.D:
		out := NewD()
		for _, value := range v {
			jsonValue, err := GetBSONValueAsJSON(value.Value, extJSON)
			if err != nil {
				return nil, err
			}
			out = append(out, NewDocElem(value.Name,
				jsonValue),
			)
		}
		return MarshalD(out), nil
	case MarshalD:
		out, err := GetBSONValueAsJSON(bson.D(v), extJSON)
		if err != nil {
			return nil, err
		}
		return MarshalD(out.(bson.D)), nil
	case []interface{}: // array
		out := []interface{}{}
		for _, value := range v {
			jsonValue, err := GetBSONValueAsJSON(value, extJSON)
			if err != nil {
				return nil, err
			}
			out = append(out, jsonValue)
		}
		return out, nil
	case []bson.M:
		out := []interface{}{}
		for _, value := range v {
			jsonValue, err := GetBSONValueAsJSON(value, extJSON)
			if err != nil {
				return nil, err
			}
			out = append(out, jsonValue)
		}
		return out, nil
	case []bson.D:
		out := []interface{}{}
		for _, value := range v {
			jsonValue, err := GetBSONValueAsJSON(value, extJSON)
			if err != nil {
				return nil, err
			}
			out = append(out, jsonValue)
		}
		return out, nil
	case string:
		return v, nil // require no conversion

	case int:
		return json.NumberInt(v), nil

	case bson.ObjectId: // ObjectId
		if !extJSON {
			s := fmt.Sprintf("ObjectId('%v')", v.Hex())
			return json.ShellMode(s), nil
		}
		return json.ObjectID(v.Hex()), nil

	case bson.Decimal128:
		y, _ := bson.ParseDecimal128(v.String())
		if !extJSON {
			s := fmt.Sprintf("NumberDecimal('%v')", y)
			return json.ShellMode(s), nil
		}
		return json.Decimal128{Value: y}, nil

	case time.Time: // Date
		date := v.Unix()*1000 + int64(v.Nanosecond()/1e6)
		if !extJSON {
			s := fmt.Sprintf("new Date(%v)", date)
			return json.ShellMode(s), nil
		}
		return json.Date(date), nil

	case int64: // NumberLong
		if !extJSON {
			s := fmt.Sprintf("NumberLong('%v')", v)
			return json.ShellMode(s), nil
		}
		return json.NumberLong(v), nil

	case int32: // NumberInt
		return json.NumberInt(v), nil

	case float64:
		return json.NumberFloat(v), nil

	case uint32:
		return json.NumberUint32(v), nil

	case uint64:
		return json.NumberUint64(v), nil

	case []byte: // BinData (with generic type)
		data := base64.StdEncoding.EncodeToString(v)
		return json.BinData{Type: 0x00, Base64: data}, nil

	case bson.Binary: // BinData
		data := base64.StdEncoding.EncodeToString(v.Data)
		if !extJSON {
			s := fmt.Sprintf("BinData(%v, %v)", v.Kind, data)
			return json.ShellMode(s), nil
		}

		return json.BinData{Type: v.Kind, Base64: data}, nil

	case bson.RegEx: // RegExp
		if !extJSON {
			s := fmt.Sprintf("/%v/%v", v.Pattern, v.Options)
			return json.ShellMode(s), nil
		}
		return json.RegExp{Pattern: v.Pattern, Options: v.Options}, nil

	case bson.MongoTimestamp: // Timestamp
		timestamp := int64(v)
		seconds := uint32(timestamp >> 32)
		increment := uint32(timestamp)

		if !extJSON {
			s := fmt.Sprintf("Timestamp(%v, %v)", seconds, increment)
			return json.ShellMode(s), nil
		}

		return json.Timestamp{
			Seconds:   seconds,
			Increment: increment,
		}, nil

	case bson.JavaScript: // JavaScript
		var scope interface{}
		var err error
		if v.Scope != nil {
			scope, err = GetBSONValueAsJSON(v.Scope, extJSON)
			if err != nil {
				return nil, err
			}
		}
		return json.JavaScript{Code: v.Code, Scope: scope}, nil

	default:
		switch x {
		case bson.MinKey: // MinKey
			if !extJSON {
				return json.ShellMode("MinKey"), nil
			}
			return json.MinKey{}, nil

		case bson.MaxKey: // MaxKey
			if !extJSON {
				return json.ShellMode("MaxKey"), nil
			}
			return json.MaxKey{}, nil

		case bson.Undefined: // undefined
			if !extJSON {
				return json.ShellMode("undefined"), nil
			}
			return json.Undefined{}, nil
		}
	}

	return nil, fmt.Errorf("conversion of BSON value '%v' of type '%T' not supported", x, x)
}
