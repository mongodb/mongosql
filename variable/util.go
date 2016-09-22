package variable

import "github.com/10gen/sqlproxy/mysqlerrors"

func convertBool(v interface{}) (bool, bool) {
	switch tv := v.(type) {
	case bool:
		return tv, true
	case byte:
		return tv == 1, true
	case int:
		return tv == 1, true
	case int16:
		return tv == 1, true
	case int32:
		return tv == 1, true
	case int64:
		return tv == 1, true
	default:
		return false, false
	}
}

func convertInt64(v interface{}) (int64, bool) {
	switch tv := v.(type) {
	case byte:
		return int64(tv), true
	case int:
		return int64(tv), true
	case int16:
		return int64(tv), true
	case int32:
		return int64(tv), true
	case int64:
		return tv, true
	default:
		return 0, false
	}
}

func convertString(v interface{}) (string, bool) {
	s, ok := v.(string)
	return s, ok
}

func kindToString(k Kind) string {
	switch k {
	case SystemKind:
		return "system"
	case StatusKind:
		return "status"
	case UserKind:
		return "user"
	default:
		return "unknown kind"
	}
}

func invalidValueError(n Name, v interface{}) error {
	return mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_VALUE_FOR_VAR, n, v)

}

func wrongTypeError(n Name, v interface{}) error {
	return mysqlerrors.Newf(mysqlerrors.ER_WRONG_TYPE_FOR_VAR, "Incorrect arg type for variable %s: %T", n, v)
}
