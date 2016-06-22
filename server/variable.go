package server

import (
	"sync"

	"github.com/10gen/sqlproxy/evaluator"
)

type sessionVariableContainer struct {
	systemVariables map[string]evaluator.SQLValue
	userVariables   map[string]evaluator.SQLValue
}

func newSessionVariableContainer(global *globalVariableContainer) *sessionVariableContainer {

	session := &sessionVariableContainer{}

	global.RLock()
	defer global.RUnlock()

	if global.systemVariables != nil {
		session.systemVariables = make(map[string]evaluator.SQLValue)
		for k, v := range global.systemVariables {
			session.systemVariables[k] = v
		}
	}

	return session
}

func (c *sessionVariableContainer) getSessionVariable(name string) (evaluator.SQLValue, bool) {
	if c.systemVariables != nil {
		if value, ok := c.systemVariables[name]; ok {
			return value, true
		}
	}

	return nil, false
}

func (c *sessionVariableContainer) getUserVariable(name string) (evaluator.SQLValue, bool) {
	if c.userVariables != nil {
		if v, ok := c.userVariables[name]; ok {
			return v, true
		}
	}

	return nil, false
}

func (c *sessionVariableContainer) setSessionVariable(name string, value evaluator.SQLValue) {
	if c.systemVariables == nil {
		c.systemVariables = make(map[string]evaluator.SQLValue)
	}

	c.systemVariables[name] = value
}

func (c *sessionVariableContainer) setUserVariable(name string, value evaluator.SQLValue) {
	if c.userVariables == nil {
		c.userVariables = make(map[string]evaluator.SQLValue)
	}

	c.userVariables[name] = value
}

type globalVariableContainer struct {
	sync.RWMutex
	systemVariables map[string]evaluator.SQLValue
}

func newGlobalVariableContainer() *globalVariableContainer {
	return &globalVariableContainer{}
}

func (c *globalVariableContainer) getValue(name string) (evaluator.SQLValue, bool) {
	c.RLock()
	defer c.RUnlock()
	if c.systemVariables != nil {
		if value, ok := c.systemVariables[name]; ok {
			return value, true
		}
	}

	return nil, false
}

func (c *globalVariableContainer) setValue(name string, value evaluator.SQLValue) {
	c.Lock()
	defer c.Unlock()
	if c.systemVariables == nil {
		c.systemVariables = make(map[string]evaluator.SQLValue)
	}
	c.systemVariables[name] = value
}
