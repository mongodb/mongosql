package config

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func fromMap(key string, v reflect.Value, values map[interface{}]interface{}, enabledExpansions Expansion) error {

	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("expected a pointer")
	}

	e := v.Elem()
	for i := 0; i < e.NumField(); i++ {
		name, _ := getNameAndProtected(e.Type().Field(i))

		val, found := values[name]
		if !found {
			continue
		}

		delete(values, name)

		newKey := joinKeys(key, name)

		eval, err := evaluateExpansionDirective(enabledExpansions, val)
		if err != nil {
			return err
		}
		val = eval

		switch e.Field(i).Kind() {
		case reflect.Struct:
			tval, ok := val.(map[interface{}]interface{})
			if !ok {
				return fmt.Errorf("invalid value for %s, expected a map: %v(%T)", newKey, val, val)
			}

			err := fromMap(newKey, e.Field(i).Addr(), tval, enabledExpansions)
			if err != nil {
				return err
			}
		case reflect.Slice:
			switch e.Field(i).Type().Elem().Kind() {
			case reflect.String:
				var tval []string
				ival, ok := val.([]interface{})
				if ok {
					for j := 0; j < len(ival); j++ {
						sval, ok := ival[j].(string)
						if !ok {
							return fmt.Errorf(
								"invalid value for %s, expected a string: %v(%T)",
								newKey, val, val,
							)
						}
						tval = append(tval, sval)
					}
				} else {
					// see if they are comma-separated
					sval, ok := val.(string)
					if !ok {
						return fmt.Errorf(
							"invalid value for %s, expected a string: %v(%T)",
							newKey, val, val,
						)
					}

					tval = strings.Split(sval, ",")
				}

				e.Field(i).Set(reflect.ValueOf(tval))
			}
		case reflect.Bool:
			tval, ok := val.(bool)
			if !ok {
				return fmt.Errorf("invalid value for %s, expected a bool: %v(%T)", newKey, val, val)
			}

			e.Field(i).SetBool(tval)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			tval, err := convertToInt64(val)
			if err != nil {
				return fmt.Errorf("invalid value for %s: %v", newKey, err)
			}
			e.Field(i).SetInt(tval)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tval, err := convertToUint64(val)
			if err != nil {
				return fmt.Errorf("invalid value for %s: %v", newKey, err)
			}
			e.Field(i).SetUint(tval)
		case reflect.String:
			tval, ok := val.(string)
			if !ok {
				return fmt.Errorf(
					"invalid value for %s, expected a string: %v(%T)",
					newKey, val, val,
				)
			}

			e.Field(i).SetString(tval)
		default:
			panic(fmt.Sprintf("unsupported config field of kind %v", e.Field(i).Kind()))
		}
	}

	for k := range values {
		if key == "" {
			return fmt.Errorf("unrecognized key '%v'", k)
		}
		return fmt.Errorf("unrecognized key '%s.%v'", key, k)
	}

	return nil
}

func convertToInt64(v interface{}) (int64, error) {
	switch tv := v.(type) {
	case int:
		return int64(tv), nil
	case int8:
		return int64(tv), nil
	case int16:
		return int64(tv), nil
	case int32:
		return int64(tv), nil
	case int64:
		return tv, nil
	case uint:
		return int64(tv), nil
	case uint8:
		return int64(tv), nil
	case uint16:
		return int64(tv), nil
	case uint32:
		return int64(tv), nil
	case uint64:
		return int64(tv), nil
	case string:
		return strconv.ParseInt(tv, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", v)
	}
}

func convertToUint64(v interface{}) (uint64, error) {
	switch tv := v.(type) {
	case int:
		return uint64(tv), nil
	case int8:
		return uint64(tv), nil
	case int16:
		return uint64(tv), nil
	case int32:
		return uint64(tv), nil
	case int64:
		return uint64(tv), nil
	case uint:
		return uint64(tv), nil
	case uint8:
		return uint64(tv), nil
	case uint16:
		return uint64(tv), nil
	case uint32:
		return uint64(tv), nil
	case uint64:
		return tv, nil
	case string:
		return strconv.ParseUint(tv, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to uint64", v)
	}
}

func joinKeys(k1, k2 string) string {
	if k1 == "" {
		return k2
	}

	return fmt.Sprintf("%s.%s", k1, k2)
}

func toJSON(w *bytes.Buffer, v1, v2 reflect.Value) {
	printed := false
	for i := 0; i < v1.NumField(); i++ {

		f1 := v1.Field(i)
		f2 := v2.Field(i)
		f2iface := f2.Interface()
		if reflect.DeepEqual(f1.Interface(), f2iface) {
			// don't print out structs/values that are equal
			continue
		}

		if printed {
			w.WriteString(", ")
		}

		name, protected := getNameAndProtected(v1.Type().Field(i))
		w.WriteString(name + ": ")

		printed = true
		switch f1.Kind() {
		case reflect.Struct:
			w.WriteString("{")
			toJSON(w, f1, f2)
			w.WriteString("}")
		case reflect.String:
			var f2String string
			if protected {
				f2String = "<protected>"
			} else {
				f2String = f2.String()
			}
			w.WriteString(strconv.Quote(f2String))
		default:
			var f2String string
			if protected {
				f2String = "<protected>"
			} else {
				f2String = fmt.Sprintf("%v", f2)
			}
			w.WriteString(f2String)
		}
	}
}

func getNameAndProtected(f reflect.StructField) (string, bool) {
	var name string
	protected := false
	if flagString, ok := f.Tag.Lookup("config"); ok {
		name, protected = parseFlagString(flagString)
	}

	if name == "" {
		name = f.Name
		r, n := utf8.DecodeRuneInString(name)
		name = string(unicode.ToLower(r)) + name[n:]
	}

	return name, protected
}

func parseFlagString(flagString string) (string, bool) {
	parts := strings.Split(flagString, ",")
	if len(parts) > 2 {
		panic("invalid config flag")
	}

	protected := false
	if len(parts) == 2 {
		protected = parts[1] == "protected"
	}

	return parts[0], protected
}
