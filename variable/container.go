package variable

import (
	"fmt"
	"math"
	"strings"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/common"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
)

// Container holds variables based on a scope.
type Container struct {
	scope  Scope
	parent *Container

	// userValues is storage for user variables
	userValues map[Name]interface{}

	//
	// backing storage for non-user variables below
	//

	AutoCommit             bool
	CharacterSetClient     *collation.Charset
	CharacterSetConnection *collation.Charset
	CharacterSetDatabase   *collation.Charset
	CharacterSetResults    *collation.Charset
	CollationConnection    *collation.Collation
	CollationDatabase      *collation.Collation
	CollationServer        *collation.Collation
	MaxAllowedPacket       int64
	MongoDBInfo            *mongodb.Info
	SQLAutoIsNull          bool
	SQLSelectLimit         uint64
	Version                string
	VersionComment         string
	InteractiveTimeoutSecs int64
	WaitTimeoutSecs        int64
}

// NewGlobalContainer creates a container with a GlobalScope.
func NewGlobalContainer() *Container {
	return &Container{
		scope: GlobalScope,

		// default values
		AutoCommit:             true,
		CharacterSetClient:     collation.DefaultCharset,
		CharacterSetConnection: collation.DefaultCharset,
		CharacterSetDatabase:   collation.DefaultCharset,
		CharacterSetResults:    collation.DefaultCharset,
		CollationConnection:    collation.Default,
		CollationDatabase:      collation.Default,
		CollationServer:        collation.Default,
		MaxAllowedPacket:       1073741824,
		MongoDBInfo:            nil,
		SQLAutoIsNull:          false,
		SQLSelectLimit:         math.MaxUint64,
		Version:                "5.7.12",
		VersionComment:         "mongosqld " + common.VersionStr,
		InteractiveTimeoutSecs: 28800,
		WaitTimeoutSecs:        28800,
	}
}

// NewSessionContainer creates a container with a SessionScope.
func NewSessionContainer(global *Container) *Container {
	if global == nil {
		panic("internal error: global cannot be nil")
	}

	c := &Container{
		scope:      SessionScope,
		parent:     global,
		userValues: make(map[Name]interface{}),
	}

	for _, def := range definitions {
		if !def.Dummy && def.GetValue != nil && def.SetValue != nil {
			value := def.GetValue(global)
			def.SetValue(c, value)
		}
	}

	return c
}

// List lists the values for the given scope and kind.
func (c *Container) List(scope Scope, kind Kind) []Value {
	if kind == UserKind {
		if scope != SessionScope {
			panic("internal error: cannot get user variables from a global scope")
		}

		var values []Value
		for k, v := range c.userValues {
			values = append(values, Value{
				Name:    Name(k),
				Kind:    UserKind,
				SQLType: schema.SQLNone,
				Value:   v,
			})
		}

		return values
	}

	if c.scope == scope {
		var values []Value
		for _, def := range definitions {
			if def.GetValue == nil {
				continue
			}

			if def.Kind != kind {
				continue
			}

			values = append(values, Value{
				Name:    def.Name,
				Kind:    def.Kind,
				SQLType: def.SQLType,
				Value:   def.GetValue(c),
			})
		}
		return values
	} else if c.parent != nil {
		return c.parent.List(scope, kind)
	}

	panic(fmt.Sprintf("internal error: illegal scope %v", scope))
}

// Get gets the value of the variable with the specified name, scope, and kind.
func (c *Container) Get(name Name, scope Scope, kind Kind) (Value, error) {

	lowerName := Name(strings.ToLower(string(name)))

	if kind == UserKind {
		if scope != SessionScope {
			panic("internal error: cannot get user variable from a global scope")
		}

		v, _ := c.userValues[lowerName]

		return Value{
			Name:    name,
			Kind:    kind,
			SQLType: schema.MongoNone,
			Value:   v,
		}, nil
	}

	if c.scope == scope {
		if def, ok := definitions[lowerName]; ok && def.Kind == kind {
			return Value{
				Name:    name,
				Kind:    def.Kind,
				SQLType: def.SQLType,
				Value:   def.GetValue(c),
			}, nil
		}
	} else if c.parent != nil {
		return c.parent.Get(name, scope, kind)
	}

	return Value{}, mysqlerrors.Defaultf(mysqlerrors.ER_UNKNOWN_SYSTEM_VARIABLE, name)
}

// Set sets the value of a variable with the specified name, scope, and kind.
func (c *Container) Set(name Name, scope Scope, kind Kind, value interface{}) error {

	lowerName := Name(strings.ToLower(string(name)))

	if kind == UserKind {
		if scope != SessionScope {
			panic("internal error: cannot set user variable on a global scope")
		}

		c.userValues[lowerName] = value
		return nil
	}

	if kind == StatusKind {
		return mysqlerrors.Defaultf(mysqlerrors.ER_UNKNOWN_SYSTEM_VARIABLE, name)
	}

	def, ok := definitions[lowerName]
	if !ok {
		return mysqlerrors.Defaultf(mysqlerrors.ER_UNKNOWN_SYSTEM_VARIABLE, name)
	}

	if (def.AllowedSetScopes & scope) != scope {
		if scope == SessionScope {
			return mysqlerrors.Defaultf(mysqlerrors.ER_GLOBAL_VARIABLE, name)
		}

		return mysqlerrors.Defaultf(mysqlerrors.ER_LOCAL_VARIABLE, name)
	}

	if c.scope == scope {
		return def.SetValue(c, value)
	} else if c.parent != nil {
		return c.parent.Set(name, scope, kind, value)
	}

	panic(fmt.Sprintf("internal error: illegal scope %v", scope))
}
