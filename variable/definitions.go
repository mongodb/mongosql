package variable

import (
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
)

const (
	autocommit             Name = "autocommit"
	characterSetClient          = "character_set_client"
	characterSetConnection      = "character_set_connection"
	characterSetResults         = "character_set_results"
	collationConnection         = "collation_connection"
	collationServer             = "collation_server"
	maxAllowedPacket            = "max_allowed_packet"
	names                       = "names"
	sqlAutoIsNull               = "sql_auto_is_null"
	version                     = "version"
	versionComment              = "version_comment"
)

type definition struct {
	Name             Name
	Kind             Kind
	AllowedSetScopes Scope

	SQLType schema.SQLType

	GetValue func(container *Container) interface{}
	SetValue func(container *Container, value interface{}) error
}

var definitions = make(map[Name]*definition)

func init() {
	definitions[autocommit] = &definition{
		Name:             autocommit,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLBoolean,
		GetValue:         func(c *Container) interface{} { return c.AutoCommit },
		SetValue:         setAutoCommit,
	}

	definitions[characterSetClient] = &definition{
		Name:             characterSetClient,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return c.CharacterSetClient },
		SetValue:         setCharacterSetClient,
	}

	definitions[characterSetConnection] = &definition{
		Name:             characterSetConnection,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return c.CharacterSetConnection },
		SetValue:         setCharacterSetConnection,
	}

	definitions[characterSetResults] = &definition{
		Name:             characterSetResults,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return c.CharacterSetResults },
		SetValue:         setCharacterSetResults,
	}

	definitions[collationConnection] = &definition{
		Name:             collationConnection,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return c.CollationConnection },
		SetValue:         setCollationConnection,
	}

	definitions[collationServer] = &definition{
		Name:             collationServer,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return c.CollationServer },
		SetValue:         setCollationServer,
	}

	definitions[maxAllowedPacket] = &definition{
		Name:             maxAllowedPacket,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLInt,
		GetValue:         func(c *Container) interface{} { return c.MaxAllowedPacket },
		SetValue:         setMaxAllowedPacket,
	}

	definitions[names] = &definition{
		Name:             names,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		// cannot get this value...
		SetValue: setNames,
	}

	definitions[sqlAutoIsNull] = &definition{
		Name:             sqlAutoIsNull,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLBoolean,
		GetValue:         func(c *Container) interface{} { return c.SQLAutoIsNull },
		SetValue:         setSQLAutoIsNull,
	}

	definitions[version] = &definition{
		Name:             version,
		Kind:             SystemKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return c.Version },
		SetValue:         func(c *Container, v interface{}) error { c.Version = v.(string); return nil },
	}

	definitions[versionComment] = &definition{
		Name:             versionComment,
		Kind:             SystemKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return c.VersionComment },
		SetValue:         func(c *Container, v interface{}) error { c.VersionComment = v.(string); return nil },
	}
}

func setAutoCommit(c *Container, v interface{}) error {
	b, ok := convertBool(v)
	if !ok {
		return wrongTypeError(autocommit, v)
	}

	c.AutoCommit = b
	return nil
}

func setCharacterSetClient(c *Container, v interface{}) error {
	if v == nil {
		return invalidValueError(characterSetClient, nil)
	}

	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(characterSetClient, v)
	}

	// TODO: validate valid character set

	c.CharacterSetClient = s
	return nil
}

func setCharacterSetConnection(c *Container, v interface{}) error {
	if v == nil {
		return invalidValueError(characterSetConnection, nil)
	}

	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(characterSetConnection, v)
	}

	// TODO: set collation_connection to the default collation for charset.

	// TODO: validate valid character set

	c.CharacterSetConnection = s
	return nil
}

func setCharacterSetResults(c *Container, v interface{}) error {
	if v == nil {
		return invalidValueError(characterSetResults, nil)
	}

	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(characterSetResults, v)
	}

	// TODO: validate valid character set

	c.CharacterSetResults = s
	return nil
}

func setCollationConnection(c *Container, v interface{}) error {
	if v == nil {
		return invalidValueError(collationConnection, nil)
	}

	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(collationConnection, v)
	}

	// TODO: validate valid collation

	c.CollationConnection = s
	return nil
}

func setCollationServer(c *Container, v interface{}) error {
	if v == nil {
		return invalidValueError(collationServer, nil)
	}

	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(collationServer, v)
	}

	// TODO: validate valid collation

	c.CollationServer = s
	return nil
}

func setMaxAllowedPacket(c *Container, v interface{}) error {
	i, ok := convertInt64(v)
	if !ok {
		return wrongTypeError(maxAllowedPacket, v)
	}

	if i < 1024 || i > 1073741824 {
		return mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_VALUE_FOR_VAR, maxAllowedPacket, i)
	}

	c.MaxAllowedPacket = i
	return nil
}

func setNames(c *Container, v interface{}) error {
	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(names, v)
	}

	c.CharacterSetClient = s
	c.CharacterSetConnection = s
	c.CharacterSetResults = s

	return nil
}

func setSQLAutoIsNull(c *Container, v interface{}) error {
	b, ok := convertBool(v)
	if !ok {
		return wrongTypeError(maxAllowedPacket, v)
	}

	c.SQLAutoIsNull = b
	return nil
}
