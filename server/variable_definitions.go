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

type variableDefinition interface {
	name() string
}

type settableVariableDefinition interface {
	variableDefinition
	apply(c *conn, scope variableScope, value evaluator.SQLValue) error
}

type readableVariableDefinition interface {
	variableDefinition
	defaultValue() evaluator.SQLValue
}

func getSystemVariableDefinition(name string) (variableDefinition, error) {
	def, ok := systemVariableDefinitions[strings.ToLower(name)]
	if !ok {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ER_UNKNOWN_SYSTEM_VARIABLE, name)
	}

	return def, nil
}

var systemVariableDefinitions = make(map[string]variableDefinition)

const (
	autoCommitVariableName             = "autocommit"
	characterSetClientVariableName     = "character_set_client"
	characterSetConnectionVariableName = "character_set_connection"
	characterSetResultsVariableName    = "character_set_results"
	maxAllowedPacketVariableName       = "max_allowed_packet"
	namesPseudoVariableName            = "names"
)

func init() {
	variables := []variableDefinition{
		&autoCommitVariable{},
		&characterSetClientVariable{},
		&characterSetConnectionVariable{},
		&characterSetResultsVariable{},
		&maxAllowedPacketVariable{},
		&namesPseudoVariable{},
	}

	for _, v := range variables {
		systemVariableDefinitions[v.name()] = v
	}
}

type autoCommitVariable struct{}

func (v *autoCommitVariable) name() string {
	return autoCommitVariableName
}

func (v *autoCommitVariable) apply(c *conn, scope variableScope, value evaluator.SQLValue) error {

	i, ok := value.(evaluator.SQLInt)
	if !ok {
		return mysqlerrors.Newf(mysqlerrors.ER_WRONG_TYPE_FOR_VAR, "arg type for %s: %T", v.name(), value)
	}

	// only true and false
	if i != 0 && i != 1 {
		return mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_VALUE_FOR_VAR, v.name(), i)
	}

	if scope == globalScope {
		c.server.variables.setValue(v.name(), i)
	} else {
		c.variables.setSessionVariable(v.name(), i)
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

type characterSetClientVariable struct{}

func (v *characterSetClientVariable) name() string {
	return characterSetClientVariableName
}

func (v *characterSetClientVariable) apply(c *conn, scope variableScope, value evaluator.SQLValue) error {
	if scope == globalScope {
		return mysqlerrors.Defaultf(mysqlerrors.ER_LOCAL_VARIABLE, v.name())
	}

	name, ok := value.(evaluator.SQLVarchar)
	if !ok {
		return mysqlerrors.Newf(mysqlerrors.ER_WRONG_TYPE_FOR_VAR, "arg type for %s: %T", v.name(), value)
	}

	if name != DEFAULT_CHARSET {
		return mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_VALUE_FOR_VAR, v.name(), name)
	}

	c.variables.setSessionVariable(v.name(), value)
	return nil
}

func (v *characterSetClientVariable) defaultValue() evaluator.SQLValue {
	return evaluator.SQLVarchar(DEFAULT_CHARSET)
}

type characterSetConnectionVariable struct{}

func (v *characterSetConnectionVariable) name() string {
	return characterSetConnectionVariableName
}

func (v *characterSetConnectionVariable) apply(c *conn, scope variableScope, value evaluator.SQLValue) error {
	if scope == globalScope {
		return mysqlerrors.Defaultf(mysqlerrors.ER_LOCAL_VARIABLE, v.name())
	}

	name, ok := value.(evaluator.SQLVarchar)
	if !ok {
		return mysqlerrors.Newf(mysqlerrors.ER_WRONG_TYPE_FOR_VAR, "arg type for %s: %T", v.name(), value)
	}

	if name != DEFAULT_CHARSET {
		return mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_VALUE_FOR_VAR, v.name(), name)
	}

	c.variables.setSessionVariable(v.name(), value)
	c.charset = value.String()
	return nil
}

func (v *characterSetConnectionVariable) defaultValue() evaluator.SQLValue {
	return evaluator.SQLVarchar(DEFAULT_CHARSET)
}

type characterSetResultsVariable struct{}

func (v *characterSetResultsVariable) name() string {
	return characterSetResultsVariableName
}

func (v *characterSetResultsVariable) apply(c *conn, scope variableScope, value evaluator.SQLValue) error {
	if scope == globalScope {
		return mysqlerrors.Defaultf(mysqlerrors.ER_LOCAL_VARIABLE, v.name())
	}

	name, ok := value.(evaluator.SQLVarchar)
	if !ok {
		return mysqlerrors.Newf(mysqlerrors.ER_WRONG_TYPE_FOR_VAR, "arg type for %s: %T", v.name(), value)
	}

	if name != DEFAULT_CHARSET {
		return mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_VALUE_FOR_VAR, v.name(), name)
	}

	c.variables.setSessionVariable(v.name(), value)
	return nil
}

func (v *characterSetResultsVariable) defaultValue() evaluator.SQLValue {
	return evaluator.SQLVarchar(DEFAULT_CHARSET)
}

type maxAllowedPacketVariable struct{}

func (v *maxAllowedPacketVariable) name() string {
	return maxAllowedPacketVariableName
}

func (v *maxAllowedPacketVariable) apply(c *conn, scope variableScope, value evaluator.SQLValue) error {
	if scope == sessionScope {
		return mysqlerrors.Defaultf(mysqlerrors.ER_GLOBAL_VARIABLE, v.name())
	}

	i, ok := value.(evaluator.SQLInt)
	if !ok {
		return mysqlerrors.Newf(mysqlerrors.ER_WRONG_TYPE_FOR_VAR, "arg type for %s: %T", v.name(), value)
	}

	if i < 1024 || i > 1073741824 {
		return mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_VALUE_FOR_VAR, v.name(), i)
	}

	c.server.variables.setValue(v.name(), value)
	return nil
}

func (v *maxAllowedPacketVariable) defaultValue() evaluator.SQLValue {
	return evaluator.SQLInt(4194304)
}

type namesPseudoVariable struct{}

func (v *namesPseudoVariable) name() string {
	return namesPseudoVariableName
}

func (v *namesPseudoVariable) apply(c *conn, scope variableScope, value evaluator.SQLValue) error {
	err := systemVariableDefinitions[characterSetClientVariableName].(settableVariableDefinition).apply(c, scope, value)
	if err != nil {
		return err
	}
	err = systemVariableDefinitions[characterSetConnectionVariableName].(settableVariableDefinition).apply(c, scope, value)
	if err != nil {
		return err
	}
	return systemVariableDefinitions[characterSetResultsVariableName].(settableVariableDefinition).apply(c, scope, value)
}
