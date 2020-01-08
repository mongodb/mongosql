package variable

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
)

// Container holds variables based on a scope.
type Container struct {
	lock   sync.RWMutex
	scope  Scope
	parent *Container

	// userValues is storage for user variables
	userValues map[Name]values.SQLValue

	// Backing storage for system variables
	systemVariableContainer

	// Backing storage for non-user MySQL status variables below.
	BytesReceived    *uint64
	BytesSent        *uint64
	Connections      *uint32
	Queries          *uint64
	StartTime        time.Time
	ThreadsConnected *uint32

	AllocatedMemory func() uint64
}

// NewEmptyContainer creates an empty container.
func NewEmptyContainer() *Container {
	return &Container{}
}

// NewGlobalContainer creates a container with a GlobalScope.
func NewGlobalContainer(cfg *config.Config) *Container {

	// Initialize system variables from defaults and, if applicable, config.
	sysVars := systemVariableContainer{}
	sysVars.setDefaults()
	if cfg != nil {
		sysVars.setFromConfig(cfg)
	}

	// Initialize server status variables here
	bytesReceived := uint64(0)
	bytesSent := uint64(0)
	connections := uint32(0)
	queries := uint64(0)
	startTime := time.Now()
	threadsConnected := uint32(0)

	return &Container{
		scope:                   GlobalScope,
		systemVariableContainer: sysVars,

		// Default values for non-user MySQL status variables below.
		BytesReceived:    &bytesReceived,
		BytesSent:        &bytesSent,
		Connections:      &connections,
		Queries:          &queries,
		StartTime:        startTime,
		ThreadsConnected: &threadsConnected,

		AllocatedMemory: func() uint64 { return 0 },
	}
}

// NewSessionContainer creates a container with a SessionScope.
func NewSessionContainer(global *Container) *Container {
	if global == nil {
		panic("global cannot be nil")
	}

	c := &Container{
		scope:           SessionScope,
		parent:          global,
		userValues:      make(map[Name]values.SQLValue),
		AllocatedMemory: func() uint64 { return 0 },
	}

	global.lock.RLock()
	defer global.lock.RUnlock()

	for _, def := range definitions {
		if !def.Dummy && def.GetValue != nil && def.SetValue != nil {
			value := def.GetValue(global)
			if err := def.SetValue(c, value); err != nil {
				// Previously unchecked error.
				panic(err)
			}
		}
	}
	return c
}

// List lists the values for the given scope and kind.
func (c *Container) List(scope Scope, kind Kind) []values.NamedSQLValue {
	if c.scope == GlobalScope && kind == SystemKind {
		c.lock.RLock()
		defer c.lock.RUnlock()
	}

	if kind == UserKind {
		if scope != SessionScope {
			panic("cannot get user variables from a global scope")
		}

		var vs []values.NamedSQLValue
		for name, v := range c.userValues {
			vs = append(vs, values.NewNamedSQLValue(string(name), v))
		}

		return vs
	}

	if c.scope == scope {
		var vs []values.NamedSQLValue
		for name, def := range definitions {
			if def.GetValue == nil {
				continue
			}

			if def.Kind != kind {
				continue
			}

			vs = append(vs, values.NewNamedSQLValue(string(name), def.GetValue(c)))
		}
		return vs
	} else if c.parent != nil {
		return c.parent.List(scope, kind)
	}

	panic(fmt.Sprintf("illegal scope %v", scope))
}

// Get gets the value of the variable with the specified name, scope, and kind.
func (c *Container) Get(name Name, scope Scope, kind Kind) (values.SQLValue, error) {
	if c.scope == GlobalScope && kind == SystemKind {
		c.lock.RLock()
		defer c.lock.RUnlock()
	}
	lowerName := Name(strings.ToLower(string(name)))

	if kind == UserKind {
		if scope != SessionScope {
			panic(fmt.Sprintf("cannot get user variable: %v from a global scope: %v", name, scope))
		}

		v := c.userValues[lowerName]
		if v == nil {
			v = values.NewSQLNull(values.VariableSQLValueKind)
		}

		return v, nil
	}

	if c.scope == scope {
		if def, ok := definitions[lowerName]; ok && def.Kind == kind {
			v := def.GetValue(c)
			if v == nil {
				v = values.NewSQLNull(values.VariableSQLValueKind)
			}

			return v, nil
		}
	} else if c.parent != nil {
		return c.parent.Get(name, scope, kind)
	}

	return values.NewSQLNull(values.VariableSQLValueKind), mysqlerrors.Defaultf(mysqlerrors.ErUnknownSystemVariable, name)
}

// GetBool gets the value of the variable with the specified name for system variable of
// boolean type.
func (c *Container) GetBool(name Name) bool {
	value, err := c.Get(name, c.scope, SystemKind)
	if err != nil {
		panic(fmt.Sprintf("cannot get boolean system variable %v: %v", name, err))
	}

	return value.Value().(bool)
}

// GetCharset gets the value of the variable with the specified name for system variable of
// collation.Charset type.
func (c *Container) GetCharset(name Name) *collation.Charset {
	value, err := c.Get(name, c.scope, SystemKind)
	if err != nil {
		panic(fmt.Sprintf("cannot get string system variable %v: %v", name, err))
	}

	if value == nil {
		return collation.NullCharset
	}

	if value.IsNull() {
		return collation.NullCharset
	}
	cs, err := collation.GetCharset(collation.CharsetName(value.Value().(string)))
	if err != nil {
		panic(err)
	}
	return cs
}

// GetCollation gets the value of the variable with the specified name for system variable of
// collation.Collation type.
func (c *Container) GetCollation(name Name) *collation.Collation {
	cName := c.GetString(name)
	col, err := collation.Get(collation.Name(cName))
	if err != nil {
		panic(err)
	}
	return col
}

// GetInt64 gets the value of the variable with the specified name for system variable of
// int64 type.
func (c *Container) GetInt64(name Name) int64 {
	value, err := c.Get(name, c.scope, SystemKind)
	if err != nil {
		panic(fmt.Sprintf("cannot get int64 system variable %v: %v", name, err))
	}

	return value.SQLInt().Value().(int64)
}

// GetString gets the value of the variable with the specified name for system variable of
// string type.
func (c *Container) GetString(name Name) string {
	value, err := c.Get(name, c.scope, SystemKind)
	if err != nil {
		panic(fmt.Sprintf("cannot get string system variable %v: %v", name, err))
	}

	return value.String()
}

// GetUint64 gets the value of the variable with the specified name for system variable of
// uint64 type.
func (c *Container) GetUint64(name Name) uint64 {
	value, err := c.Get(name, c.scope, SystemKind)
	if err != nil {
		panic(fmt.Sprintf("cannot get uint64 system variable %v: %v", name, err))
	}

	return value.SQLUint().Value().(uint64)
}

// Set sets the value of a variable with the specified name, scope, and kind.
func (c *Container) Set(name Name, scope Scope, kind Kind, value values.SQLValue) error {
	lowerName := Name(strings.ToLower(string(name)))
	switch kind {
	case UserKind:
		if scope != SessionScope {
			panic(fmt.Sprintf("cannot set user variable: %v on a global scope: %v", name, scope))
		}
		c.userValues[lowerName] = value
		return nil
	case StatusKind:
		return mysqlerrors.Defaultf(mysqlerrors.ErUnknownSystemVariable, name)
	case SystemKind:
		if err := validateSetScope(scope, name); err != nil {
			return err
		}
		return c.setSystemVariable(scope, name, value)
	default:
		panic(fmt.Sprintf("unknown variable kind %v", kind))
	}
}

func (c *Container) setSystemVariable(scope Scope, name Name, value values.SQLValue) error {
	def, err := getDefinition(name)
	if err != nil {
		return err
	}

	if c.scope == GlobalScope {
		c.lock.Lock()
		defer c.lock.Unlock()
	}

	if fmt.Sprintf("%v", value) == "default" {
		v, err := NewGlobalContainer(nil).Get(name, GlobalScope, SystemKind)
		if err != nil {
			return err
		}
		return c.setSystemVariable(scope, name, v)
	}

	if c.scope == scope {
		return def.SetValue(c, value)
	} else if c.parent != nil {
		return c.parent.setSystemVariable(scope, name, value)
	}

	panic(fmt.Sprintf("illegal scope %v", scope))
}

// SetSystemVariable sets the value of the variable with the specified name for
// system variable. This function skips set scope validation, so it should only
// be used to initialize variables, not as part of a SET command.
func (c *Container) SetSystemVariable(name Name, value values.SQLValue) {
	err := c.setSystemVariable(c.scope, name, value)
	if err != nil {
		panic(fmt.Sprintf("unexpected error initializing system variable '%s': %v", name, err))
	}
}

func getDefinition(name Name) (*definition, error) {
	lowerName := Name(strings.ToLower(string(name)))
	def, ok := definitions[lowerName]
	if !ok {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErUnknownSystemVariable, name)
	}
	return def, nil
}

func validateSetScope(scope Scope, name Name) error {
	def, err := getDefinition(name)
	if err != nil {
		return err
	}

	if (def.AllowedSetScopes & scope) != scope {
		if def.AllowedSetScopes == Scope(0) {
			return mysqlerrors.Defaultf(mysqlerrors.ErVariableIsReadonly, scope, name)
		}
		if scope == SessionScope {
			return mysqlerrors.Defaultf(mysqlerrors.ErGlobalVariable, name)
		}
		return mysqlerrors.Defaultf(mysqlerrors.ErLocalVariable, name)
	}

	return nil
}
