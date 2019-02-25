package variable

import (
	"fmt"

	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
)

// MySQL has strange semantics with respect to type conversion for variables.
//
// First, in most cases, the values must have the exact same type as the variables rather than
// coercing types as is normal for MySQL in other evaluation contexts.
//
// Second, booleans are interesting because MySQL does not have a true boolean type, but rather uses
// integers as booleans. However, rather than treating all non-0 results as true, like in other
// evaluation contexts, they treat any value that is not literal true, false, 0, or 1 as incorrect,
// including 1.0, and 0.0, for which it throws ErWrongTypeForVar.  For integers, mysql throws
// ErWrongValueForVar, if they are not 1 or 0.  The keywords true and false are always allowed in
// Int variable contexts, likely because, again, MySQL really doesn't make a distinction between
// integers and booleans.
func convertSQLBool(name Name, v values.SQLValue) (values.SQLBool, error) {
	switch typedV := v.(type) {
	case values.SQLBool:
		return typedV, nil
	case values.SQLInt64:
		i := typedV.Value().(int64)
		if i == 0 || i == 1 {
			return typedV.SQLBool(), nil
		}
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongValueForVar, name, v.String())
	}
	return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongTypeForVar, name)
}

func convertSQLInt64(name Name, v values.SQLValue) (values.SQLInt64, error) {
	switch typedV := v.(type) {
	case values.SQLBool, values.SQLInt64, values.SQLUint64:
		return typedV.SQLInt(), nil
	}
	return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongTypeForVar, name)
}

func convertSQLUint64(name Name, v values.SQLValue) (values.SQLUint64, error) {
	switch typedV := v.(type) {
	case values.SQLInt64:
		if typedV.Value().(int64) < 0 {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongValueForVar, name, v.String())
		}
		return typedV.SQLUint(), nil
	case values.SQLBool, values.SQLUint64:
		return typedV.SQLUint(), nil
	}
	return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongTypeForVar, name)
}

func convertSQLVarchar(name Name, v values.SQLValue) (values.SQLVarchar, error) {
	if out, ok := v.(values.SQLVarchar); ok {
		return out, nil
	}
	return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongTypeForVar, name)
}

func lessThan(val values.SQLValue, i int64) bool {
	switch val.(type) {
	case values.SQLUint64:
		return val.Value().(uint64) < uint64(i)
	case values.SQLInt64:
		return val.Value().(int64) < i
	}
	panic(fmt.Sprintf("bad type %T for lessThan", val))
}

func greaterThan(val values.SQLValue, i int64) bool {
	switch val.(type) {
	case values.SQLUint64:
		return val.Value().(uint64) > uint64(i)
	case values.SQLInt64:
		return val.Value().(int64) > i
	}
	panic(fmt.Sprintf("bad type %T for greaterThan", val))
}

// nolint: unparam
func invalidValueError(n Name, v values.SQLValue) error {
	return mysqlerrors.Defaultf(mysqlerrors.ErWrongValueForVar, n, fmt.Sprintf("%v", v))
}
