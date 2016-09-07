package variable

import (
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
)

const (
	autocommit             Name = "autocommit"
	characterSetClient          = "character_set_client"
	characterSetConnection      = "character_set_connection"
	characterSetDatabase        = "character_set_database"
	characterSetResults         = "character_set_results"
	collationConnection         = "collation_connection"
	collationDatabase           = "collation_database"
	collationServer             = "collation_server"
	maxAllowedPacket            = "max_allowed_packet"
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
		GetValue:         func(c *Container) interface{} { return string(c.CharacterSetClient.Name) },
		SetValue:         setCharacterSetClient,
	}

	definitions[characterSetConnection] = &definition{
		Name:             characterSetConnection,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return string(c.CollationConnection.Charset.Name) },
		SetValue:         setCharacterSetConnection,
	}

	definitions[characterSetDatabase] = &definition{
		Name:             characterSetDatabase,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return string(c.CollationDatabase.Charset.Name) },
		SetValue:         setCharacterSetDatabase,
	}

	definitions[characterSetResults] = &definition{
		Name:             characterSetResults,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue: func(c *Container) interface{} {
			if c.CharacterSetResults.Name == "" {
				return nil
			}
			return string(c.CharacterSetResults.Name)
		},
		SetValue: setCharacterSetResults,
	}

	definitions[collationConnection] = &definition{
		Name:             collationConnection,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return string(c.CollationConnection.Name) },
		SetValue:         setCollationConnection,
	}

	definitions[collationDatabase] = &definition{
		Name:             collationDatabase,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return string(c.CollationDatabase.Name) },
		SetValue:         setCollationDatabase,
	}

	definitions[collationServer] = &definition{
		Name:             collationServer,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return string(c.CollationServer.Name) },
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

	cs, err := collation.GetCharset(collation.CharsetName(s))
	if err != nil {
		return err
	}
	c.CharacterSetClient = cs
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

	cs, err := collation.GetCharset(collation.CharsetName(s))
	if err != nil {
		return err
	}

	col, err := collation.GetByID(cs.DefaultCollationID)
	if err != nil {
		return err
	}

	c.CollationConnection = col
	return nil
}

func setCharacterSetDatabase(c *Container, v interface{}) error {
	if v == nil {
		return invalidValueError(characterSetDatabase, nil)
	}

	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(characterSetDatabase, v)
	}

	cs, err := collation.GetCharset(collation.CharsetName(s))
	if err != nil {
		return err
	}

	col, err := collation.GetByID(cs.DefaultCollationID)
	if err != nil {
		return err
	}

	c.CollationDatabase = col
	return nil
}

func setCharacterSetResults(c *Container, v interface{}) error {
	if v == nil {
		v = ""
	}

	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(characterSetResults, v)
	}

	cs, err := collation.GetCharset(collation.CharsetName(s))
	if err != nil {
		return err
	}
	c.CharacterSetResults = cs
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

	col, err := collation.Get(collation.Name(s))
	if err != nil {
		return err
	}

	c.CollationConnection = col
	return nil
}

func setCollationDatabase(c *Container, v interface{}) error {
	if v == nil {
		return invalidValueError(collationDatabase, nil)
	}

	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(collationDatabase, v)
	}

	col, err := collation.Get(collation.Name(s))
	if err != nil {
		return err
	}

	c.CollationDatabase = col
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

	col, err := collation.Get(collation.Name(s))
	if err != nil {
		return err
	}

	c.CollationServer = col
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

func setSQLAutoIsNull(c *Container, v interface{}) error {
	b, ok := convertBool(v)
	if !ok {
		return wrongTypeError(sqlAutoIsNull, v)
	}

	c.SQLAutoIsNull = b
	return nil
}
