package variable

import (
	"runtime"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
)

const (
	Autocommit             Name = "autocommit"
	CharacterSetClient          = "character_set_client"
	CharacterSetConnection      = "character_set_connection"
	CharacterSetDatabase        = "character_set_database"
	CharacterSetResults         = "character_set_results"
	CollationConnection         = "collation_connection"
	CollationDatabase           = "collation_database"
	CollationServer             = "collation_server"
	InteractiveTimeoutSecs      = "interactive_timeout"
	MaxAllowedPacket            = "max_allowed_packet"
	MongoDBGitVersion           = "mongodb_git_version"
	MongoDBVersion              = "mongodb_version"
	SqlAutoIsNull               = "sql_auto_is_null"
	Version                     = "version"
	VersionComment              = "version_comment"
	WaitTimeoutSecs             = "wait_timeout"
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

const (
	isWindowsOS = runtime.GOOS == "windows"
)

func init() {
	definitions[Autocommit] = &definition{
		Name:             Autocommit,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLBoolean,
		GetValue:         func(c *Container) interface{} { return c.AutoCommit },
		SetValue:         setAutoCommit,
	}

	definitions[CharacterSetClient] = &definition{
		Name:             CharacterSetClient,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return string(c.CharacterSetClient.Name) },
		SetValue:         setCharacterSetClient,
	}

	definitions[CharacterSetConnection] = &definition{
		Name:             CharacterSetConnection,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return string(c.CollationConnection.Charset.Name) },
		SetValue:         setCharacterSetConnection,
	}

	definitions[CharacterSetDatabase] = &definition{
		Name:             CharacterSetDatabase,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return string(c.CollationDatabase.Charset.Name) },
		SetValue:         setCharacterSetDatabase,
	}

	definitions[CharacterSetResults] = &definition{
		Name:             CharacterSetResults,
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

	definitions[CollationConnection] = &definition{
		Name:             CollationConnection,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return string(c.CollationConnection.Name) },
		SetValue:         setCollationConnection,
	}

	definitions[CollationDatabase] = &definition{
		Name:             CollationDatabase,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return string(c.CollationDatabase.Name) },
		SetValue:         setCollationDatabase,
	}

	definitions[CollationServer] = &definition{
		Name:             CollationServer,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return string(c.CollationServer.Name) },
		SetValue:         setCollationServer,
	}

	definitions[InteractiveTimeoutSecs] = &definition{
		Name:             InteractiveTimeoutSecs,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLInt64,
		GetValue:         func(c *Container) interface{} { return c.InteractiveTimeoutSecs },
		SetValue:         setInteractiveTimeoutSecs,
	}

	definitions[MaxAllowedPacket] = &definition{
		Name:             MaxAllowedPacket,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLInt,
		GetValue:         func(c *Container) interface{} { return c.MaxAllowedPacket },
		SetValue:         setMaxAllowedPacket,
	}

	definitions[MongoDBGitVersion] = &definition{
		Name:             MongoDBGitVersion,
		Kind:             SystemKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		SQLType:          schema.SQLBoolean,
		GetValue: func(c *Container) interface{} {
			if c.MongoDBInfo == nil {
				return nil
			}
			return c.MongoDBInfo.GitVersion
		},
	}

	definitions[MongoDBVersion] = &definition{
		Name:             MongoDBVersion,
		Kind:             SystemKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		SQLType:          schema.SQLBoolean,
		GetValue: func(c *Container) interface{} {
			if c.MongoDBInfo == nil {
				return nil
			}
			return c.MongoDBInfo.Version
		},
	}

	definitions[SqlAutoIsNull] = &definition{
		Name:             SqlAutoIsNull,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLBoolean,
		GetValue:         func(c *Container) interface{} { return c.SQLAutoIsNull },
		SetValue:         setSQLAutoIsNull,
	}

	definitions[Version] = &definition{
		Name:             Version,
		Kind:             SystemKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return c.Version },
		SetValue:         func(c *Container, v interface{}) error { c.Version = v.(string); return nil },
	}

	definitions[VersionComment] = &definition{
		Name:             VersionComment,
		Kind:             SystemKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return c.VersionComment },
		SetValue:         func(c *Container, v interface{}) error { c.VersionComment = v.(string); return nil },
	}

	definitions[WaitTimeoutSecs] = &definition{
		Name:             WaitTimeoutSecs,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLInt64,
		GetValue:         func(c *Container) interface{} { return c.WaitTimeoutSecs },
		SetValue:         setWaitTimeoutSecs,
	}
}

func setAutoCommit(c *Container, v interface{}) error {
	b, ok := convertBool(v)
	if !ok {
		return wrongTypeError(Autocommit, v)
	}

	c.AutoCommit = b
	return nil
}

func setCharacterSetClient(c *Container, v interface{}) error {
	if v == nil {
		return invalidValueError(CharacterSetClient, nil)
	}

	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(CharacterSetClient, v)
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
		return invalidValueError(CharacterSetConnection, nil)
	}

	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(CharacterSetConnection, v)
	}

	cs, err := collation.GetCharset(collation.CharsetName(s))
	if err != nil {
		return err
	}

	col, err := collation.Get(cs.DefaultCollationName)
	if err != nil {
		return err
	}

	c.CollationConnection = col
	return nil
}

func setCharacterSetDatabase(c *Container, v interface{}) error {
	if v == nil {
		return invalidValueError(CharacterSetDatabase, nil)
	}

	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(CharacterSetDatabase, v)
	}

	cs, err := collation.GetCharset(collation.CharsetName(s))
	if err != nil {
		return err
	}

	col, err := collation.Get(cs.DefaultCollationName)
	if err != nil {
		return err
	}

	c.CollationDatabase = col
	return nil
}

func setCharacterSetResults(c *Container, v interface{}) error {
	if v == nil {
		c.CharacterSetResults = collation.NullCharset
		return nil
	}

	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(CharacterSetResults, v)
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
		return invalidValueError(CollationConnection, nil)
	}

	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(CollationConnection, v)
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
		return invalidValueError(CollationDatabase, nil)
	}

	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(CollationDatabase, v)
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
		return invalidValueError(CollationServer, nil)
	}

	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(CollationServer, v)
	}

	col, err := collation.Get(collation.Name(s))
	if err != nil {
		return err
	}

	c.CollationServer = col
	return nil
}

func setInteractiveTimeoutSecs(c *Container, v interface{}) error {
	i, ok := convertInt64(v)
	if !ok {
		return wrongTypeError(InteractiveTimeoutSecs, v)
	}

	if i < 1 {
		return mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_VALUE_FOR_VAR, InteractiveTimeoutSecs, i)
	}

	c.InteractiveTimeoutSecs = i
	return nil
}

func setMaxAllowedPacket(c *Container, v interface{}) error {
	i, ok := convertInt64(v)
	if !ok {
		return wrongTypeError(MaxAllowedPacket, v)
	}

	if i < 1024 || i > 1073741824 {
		return mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_VALUE_FOR_VAR, MaxAllowedPacket, i)
	}

	c.MaxAllowedPacket = i
	return nil
}

func setSQLAutoIsNull(c *Container, v interface{}) error {
	b, ok := convertBool(v)
	if !ok {
		return wrongTypeError(SqlAutoIsNull, v)
	}

	c.SQLAutoIsNull = b
	return nil
}

func setWaitTimeoutSecs(c *Container, v interface{}) error {
	i, ok := convertInt64(v)
	if !ok {
		return wrongTypeError(WaitTimeoutSecs, v)
	}

	upperLimit := int64(31536000)

	if isWindowsOS {
		upperLimit = int64(2147483)
	}

	if i < 1 || i > upperLimit {
		return mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_VALUE_FOR_VAR, WaitTimeoutSecs, i)
	}

	c.WaitTimeoutSecs = i
	return nil
}
