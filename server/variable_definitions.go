package server

import (
	"strings"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/mysqlerrors"
)

type variableScope int

const (
	globalScope variableScope = iota
	sessionScope
)

type systemVariableDefinition interface {
	defaultValue() evaluator.SQLValue
	apply(c *conn, scope variableScope, value evaluator.SQLValue) error
}

func getSystemVariableDefinition(name string) (systemVariableDefinition, error) {
	def, ok := systemVariableDefinitions[strings.ToLower(name)]
	if !ok {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ER_UNKNOWN_SYSTEM_VARIABLE, name)
	}

	return def, nil
}

var systemVariableDefinitions = map[string]systemVariableDefinition{
	"autocommit":         &autoCommitVariable{"autocommit"},
	"max_allowed_packet": &maxAllowedPacketVariable{"max_allowed_packet"},
}

type autoCommitVariable struct {
	name string
}

func (v *autoCommitVariable) apply(c *conn, scope variableScope, value evaluator.SQLValue) error {

	i, ok := value.(evaluator.SQLInt)
	if !ok {
		return mysqlerrors.Newf(mysqlerrors.ER_WRONG_TYPE_FOR_VAR, "arg type for %s: %T", v.name, value)
	}

	// only true and false
	if i != 0 && i != 1 {
		return mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_VALUE_FOR_VAR, v.name, i)
	}

	if scope == globalScope {
		c.server.variables.setValue(v.name, i)
	} else {
		c.variables.setSessionVariable(v.name, i)
		if i == 1 {
			c.status |= SERVER_STATUS_AUTOCOMMIT
		} else {
			c.status &= ^SERVER_STATUS_AUTOCOMMIT
		}
	}

	return nil
}

func (v *autoCommitVariable) defaultValue() evaluator.SQLValue {
	return evaluator.SQLInt(1)
}

type maxAllowedPacketVariable struct {
	name string
}

func (v *maxAllowedPacketVariable) apply(c *conn, scope variableScope, value evaluator.SQLValue) error {
	if scope == sessionScope {
		return mysqlerrors.Defaultf(mysqlerrors.ER_VARIABLE_IS_READONLY, "SESSION", v.name, "GLOBAL")
	}

	i, ok := value.(evaluator.SQLInt)
	if !ok {
		return mysqlerrors.Newf(mysqlerrors.ER_WRONG_TYPE_FOR_VAR, "arg type for %s: %T", v.name, value)
	}

	if i < 1024 || i > 1073741824 {
		return mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_VALUE_FOR_VAR, v.name, i)
	}

	c.server.variables.setValue(v.name, value)
	return nil
}

func (v *maxAllowedPacketVariable) defaultValue() evaluator.SQLValue {
	return evaluator.SQLInt(4194304)
}
