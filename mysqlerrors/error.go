package mysqlerrors

import "fmt"

type MySqlError struct {
	Code    uint16
	Message string
	State   string
}

func (e *MySqlError) Error() string {
	return fmt.Sprintf("ERROR %d (%s): %s", e.Code, e.State, e.Message)
}

// Unknownf creates an unknown error from the specified message.
func Unknownf(format string, args ...interface{}) *MySqlError {
	return Newf(ER_UNKNOWN_ERROR, format, args...)
}

// Defaultf creates the default error message for the given errCode.
func Defaultf(errCode uint16, args ...interface{}) *MySqlError {
	e := &MySqlError{
		Code: errCode,
	}

	if s, ok := mySQLState[errCode]; ok {
		e.State = s
	} else {
		e.State = DEFAULT_MYSQL_STATE
	}

	if format, ok := mySQLErrName[errCode]; ok {
		e.Message = fmt.Sprintf(format, args...)
	} else {
		e.Message = fmt.Sprint(args...)
	}

	return e
}

// Newf creates a MySqlError for the specified errCode using the custom message.
func Newf(errCode uint16, format string, args ...interface{}) *MySqlError {
	e := &MySqlError{
		Code:    errCode,
		Message: fmt.Sprintf(format, args...),
	}

	if s, ok := mySQLState[errCode]; ok {
		e.State = s
	} else {
		e.State = DEFAULT_MYSQL_STATE
	}

	return e
}
