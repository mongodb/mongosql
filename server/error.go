package server

import (
	"errors"
	"fmt"
)

var (
	ErrBadConn       = errors.New("connection was bad")
	ErrMalformPacket = errors.New("Malform packet error")
)

type sqlError struct {
	Code    uint16
	Message string
	State   string
}

func (e *sqlError) Error() string {
	return fmt.Sprintf("ERROR %d (%s): %s", e.Code, e.State, e.Message)
}

//default mysql error, must adapt errname message format
func newDefaultError(errCode uint16, args ...interface{}) *sqlError {
	e := new(sqlError)
	e.Code = errCode

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

func newError(errCode uint16, message string) *sqlError {
	e := new(sqlError)
	e.Code = errCode

	if s, ok := mySQLState[errCode]; ok {
		e.State = s
	} else {
		e.State = DEFAULT_MYSQL_STATE
	}

	e.Message = message

	return e
}
