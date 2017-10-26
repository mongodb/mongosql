package variable

import (
	"math"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
)

// System Variable Names
const (
	Autocommit                  Name = "autocommit"
	CharacterSetClient               = "character_set_client"
	CharacterSetConnection           = "character_set_connection"
	CharacterSetDatabase             = "character_set_database"
	CharacterSetResults              = "character_set_results"
	CollationConnection              = "collation_connection"
	CollationDatabase                = "collation_database"
	CollationServer                  = "collation_server"
	InteractiveTimeoutSecs           = "interactive_timeout"
	MaxAllowedPacket                 = "max_allowed_packet"
	MongoDBMaxStageSize              = "mongodb_max_stage_size"
	MongoDBMaxVarcharLength          = "mongodb_max_varchar_length"
	MongoDBVersionCompatibility      = "mongodb_version_compatibility"
	MongoDBGitVersion                = "mongodb_git_version"
	MongoDBVersion                   = "mongodb_version"
	Socket                           = "socket"
	SQLAutoIsNull                    = "sql_auto_is_null"
	SQLSelectLimit                   = "sql_select_limit"
	Version                          = "version"
	VersionComment                   = "version_comment"
	WaitTimeoutSecs                  = "wait_timeout"
)

func init() {
	//  System Variable Definitions
	definitions[Autocommit] = &definition{
		Name:             Autocommit,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLBoolean,
		GetValue:         func(c *Container) interface{} { return c.autoCommit },
		SetValue:         setAutoCommit,
	}

	definitions[CharacterSetClient] = &definition{
		Name:             CharacterSetClient,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return string(c.characterSetClient.Name) },
		GetRawValue:      func(c *Container) interface{} { return c.characterSetClient },
		SetValue:         setCharacterSetClient,
	}

	definitions[CharacterSetConnection] = &definition{
		Name:             CharacterSetConnection,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return string(c.characterSetConnection.Name) },
		GetRawValue:      func(c *Container) interface{} { return c.characterSetConnection },
		SetValue:         setCharacterSetConnection,
	}

	definitions[CharacterSetDatabase] = &definition{
		Name:             CharacterSetDatabase,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return string(c.characterSetDatabase.Name) },
		GetRawValue:      func(c *Container) interface{} { return c.characterSetDatabase },
		SetValue:         setCharacterSetDatabase,
	}

	definitions[CharacterSetResults] = &definition{
		Name:             CharacterSetResults,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue: func(c *Container) interface{} {
			if c.characterSetResults.Name == "" {
				return nil
			}
			return string(c.characterSetResults.Name)
		},
		GetRawValue: func(c *Container) interface{} { return c.characterSetResults },
		SetValue:    setCharacterSetResults,
	}

	definitions[CollationConnection] = &definition{
		Name:             CollationConnection,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return string(c.collationConnection.Name) },
		GetRawValue:      func(c *Container) interface{} { return c.collationConnection },
		SetValue:         setCollationConnection,
	}

	definitions[CollationDatabase] = &definition{
		Name:             CollationDatabase,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return string(c.collationDatabase.Name) },
		GetRawValue:      func(c *Container) interface{} { return c.collationDatabase },
		SetValue:         setCollationDatabase,
	}

	definitions[CollationServer] = &definition{
		Name:             CollationServer,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return string(c.collationServer.Name) },
		GetRawValue:      func(c *Container) interface{} { return c.collationServer },
		SetValue:         setCollationServer,
	}

	definitions[InteractiveTimeoutSecs] = &definition{
		Name:             InteractiveTimeoutSecs,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLInt64,
		GetValue:         func(c *Container) interface{} { return c.interactiveTimeoutSecs },
		SetValue:         setInteractiveTimeoutSecs,
	}

	definitions[MaxAllowedPacket] = &definition{
		Name:             MaxAllowedPacket,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLInt,
		GetValue:         func(c *Container) interface{} { return c.maxAllowedPacket },
		SetValue:         setMaxAllowedPacket,
	}

	definitions[MongoDBMaxStageSize] = &definition{
		Name:             MongoDBMaxStageSize,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLUint64,
		GetValue:         func(c *Container) interface{} { return c.mongoDBMaxStageSize },
		SetValue:         setMongoDBMaxStageSize,
	}

	definitions[MongoDBMaxVarcharLength] = &definition{
		Name:             MongoDBMaxVarcharLength,
		Kind:             SystemKind,
		AllowedSetScopes: SessionScope,
		SQLType:          schema.SQLUint64,
		GetValue:         func(c *Container) interface{} { return c.mongoDBMaxVarcharLength },
		SetValue:         setMongoDBMaxVarcharLength,
	}

	definitions[MongoDBVersionCompatibility] = &definition{
		Name:             MongoDBVersionCompatibility,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return c.mongoDBVersionCompatibility },
		SetValue:         setMongoDBVersionCompatibility,
	}

	definitions[MongoDBGitVersion] = &definition{
		Name:             MongoDBGitVersion,
		Kind:             SystemKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		SQLType:          schema.SQLVarchar,
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
		SQLType:          schema.SQLVarchar,
		GetValue: func(c *Container) interface{} {
			if c.MongoDBInfo == nil {
				return nil
			}
			return c.MongoDBInfo.Version
		},
	}

	definitions[Socket] = &definition{
		Name:             Socket,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return c.socket },
		SetValue:         setSocket,
	}

	definitions[SQLAutoIsNull] = &definition{
		Name:             SQLAutoIsNull,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLBoolean,
		GetValue:         func(c *Container) interface{} { return c.sqlAutoIsNull },
		SetValue:         setSQLAutoIsNull,
	}

	definitions[SQLSelectLimit] = &definition{
		Name:             SQLSelectLimit,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLUint64,
		GetValue:         func(c *Container) interface{} { return c.sqlSelectLimit },
		SetValue:         setSQLSelectLimit,
	}

	definitions[Version] = &definition{
		Name:             Version,
		Kind:             SystemKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return c.version },
		SetValue:         func(c *Container, v interface{}) error { c.version = v.(string); return nil },
	}

	definitions[VersionComment] = &definition{
		Name:             VersionComment,
		Kind:             SystemKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return c.versionComment },
		SetValue:         func(c *Container, v interface{}) error { c.versionComment = v.(string); return nil },
	}

	definitions[WaitTimeoutSecs] = &definition{
		Name:             WaitTimeoutSecs,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLInt64,
		GetValue:         func(c *Container) interface{} { return c.waitTimeoutSecs },
		SetValue:         setWaitTimeoutSecs,
	}
}

func setAutoCommit(c *Container, v interface{}) error {
	b, ok := convertBool(v)
	if !ok {
		return wrongTypeError(Autocommit, v)
	}

	c.autoCommit = b
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
	c.characterSetClient = cs
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

	c.characterSetConnection = cs
	c.collationConnection = col
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

	c.characterSetDatabase = cs
	c.collationDatabase = col
	return nil
}

func setCharacterSetResults(c *Container, v interface{}) error {
	if v == nil {
		c.characterSetResults = collation.NullCharset
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
	c.characterSetResults = cs
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

	cs, err := collation.GetCharset(col.CharsetName)
	if err != nil {
		return err
	}

	c.characterSetConnection = cs
	c.collationConnection = col
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

	cs, err := collation.GetCharset(col.CharsetName)
	if err != nil {
		return err
	}

	c.characterSetDatabase = cs
	c.collationDatabase = col
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

	c.collationServer = col
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

	c.interactiveTimeoutSecs = i
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

	c.maxAllowedPacket = i
	return nil
}

func setMongoDBMaxStageSize(c *Container, v interface{}) error {
	i, ok := convertUint64(v)
	if !ok {
		if j, ok := convertInt64(v); ok {
			if j < 0 {
				i = 0
			} else {
				i = uint64(j)
			}
		} else {
			return wrongTypeError(MongoDBMaxStageSize, v)
		}
	}

	c.mongoDBMaxStageSize = i
	return nil
}

func setMongoDBMaxVarcharLength(c *Container, v interface{}) error {
	i, ok := convertUint64(v)
	if !ok {
		if j, ok := convertInt64(v); ok {
			if j < 0 {
				i = 0
			} else {
				i = uint64(j)
			}
		} else {
			return wrongTypeError(MongoDBMaxVarcharLength, v)
		}
	}

	if i > math.MaxUint16 {
		i = math.MaxUint16
	}

	c.mongoDBMaxVarcharLength = uint16(i)
	return nil
}

func setMongoDBVersionCompatibility(c *Container, v interface{}) error {
	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(MongoDBVersionCompatibility, v)
	}

	if c.MongoDBInfo != nil {
		err := c.MongoDBInfo.SetCompatibleVersion(s)
		if err != nil {
			return err
		}
	}
	c.mongoDBVersionCompatibility = s
	return nil
}

func setSocket(c *Container, v interface{}) error {
	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(Socket, v)
	}

	c.socket = s
	return nil
}

func setSQLAutoIsNull(c *Container, v interface{}) error {
	b, ok := convertBool(v)
	if !ok {
		return wrongTypeError(SQLAutoIsNull, v)
	}

	c.sqlAutoIsNull = b
	return nil
}

func setSQLSelectLimit(c *Container, v interface{}) error {
	i, ok := convertUint64(v)
	if !ok {
		if j, ok := convertInt64(v); ok {
			if j < 0 {
				i = 0
			} else {
				i = uint64(j)
			}
		} else {
			return wrongTypeError(SQLSelectLimit, v)
		}
	}

	c.sqlSelectLimit = i
	return nil
}

func setWaitTimeoutSecs(c *Container, v interface{}) error {
	i, ok := convertInt64(v)
	if !ok {
		return wrongTypeError(WaitTimeoutSecs, v)
	}

	upperLimit := int64(31536000)

	if util.IsWindowsOS {
		upperLimit = int64(2147483)
	}

	if i < 1 || i > upperLimit {
		return mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_VALUE_FOR_VAR, WaitTimeoutSecs, i)
	}

	c.waitTimeoutSecs = i
	return nil
}
