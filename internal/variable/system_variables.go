package variable

import (
	"fmt"
	"math"

	"github.com/10gen/sqlproxy/internal/collation"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
)

// System Variable Names
const (
	// Non-user MySQL system variables below.
	Autocommit             Name = "autocommit"
	CharacterSetClient     Name = "character_set_client"
	CharacterSetConnection Name = "character_set_connection"
	CharacterSetDatabase   Name = "character_set_database"
	CharacterSetResults    Name = "character_set_results"
	CollationConnection    Name = "collation_connection"
	CollationDatabase      Name = "collation_database"
	CollationServer        Name = "collation_server"
	GroupConcatMaxLen      Name = "group_concat_max_len"
	InteractiveTimeoutSecs Name = "interactive_timeout"
	MaxAllowedPacket       Name = "max_allowed_packet"
	MaxConnections         Name = "max_connections"
	MaxTimeMS              Name = "max_execution_time"
	Socket                 Name = "socket"
	SQLAutoIsNull          Name = "sql_auto_is_null"
	SQLSelectLimit         Name = "sql_select_limit"
	Version                Name = "version"
	VersionComment         Name = "version_comment"
	WaitTimeoutSecs        Name = "wait_timeout"

	// mongosqld-defined system variables below.
	EnableTableAlterations        Name = "enable_table_alterations"
	FullPushdownExecMode          Name = "full_pushdown_exec_mode"
	LogLevel                      Name = "log_level"
	MaxNestedTableDepth           Name = "max_nested_table_depth"
	MaxNumColumnsPerTable         Name = "max_num_columns_per_table"
	MetricsBackend                Name = "metrics_backend"
	MongoDBMaxServerSize          Name = "mongodb_max_server_size"
	MongoDBMaxConnectionSize      Name = "mongodb_max_connection_size"
	MongoDBMaxStageSize           Name = "mongodb_max_stage_size"
	MongoDBMaxVarcharLength       Name = "mongodb_max_varchar_length"
	MongoDBVersionCompatibility   Name = "mongodb_version_compatibility"
	MongoDBGitVersion             Name = "mongodb_git_version"
	MongoDBVersion                Name = "mongodb_version"
	MongosqldVersion              Name = "mongosqld_version"
	OptimizeCrossJoins            Name = "optimize_cross_joins"
	OptimizeEvaluations           Name = "optimize_evaluations"
	OptimizeFiltering             Name = "optimize_filtering"
	OptimizeInnerJoins            Name = "optimize_inner_joins"
	OptimizeSelfJoins             Name = "optimize_self_joins"
	OptimizeViewSampling          Name = "optimize_view_sampling"
	Pushdown                      Name = "pushdown"
	PolymorphicTypeConversionMode Name = "polymorphic_type_conversion_mode"
	RewriteDistinctAsGroup        Name = "rewrite_distinct_as_group"
	SampleRefreshIntervalSecs     Name = "sample_refresh_interval_secs"
	SampleSize                    Name = "sample_size"
	SchemaMappingMode             Name = "schema_mapping_mode"
	TypeConversionMode            Name = "type_conversion_mode"
)

// GetPolymorphicTypeConversionMode converts a string to a PolymorphicConversionMode if it
// is viable, or else panics.
func GetPolymorphicTypeConversionMode(vars *Container) PolymorphicTypeConversionModeType {
	str := vars.GetString(PolymorphicTypeConversionMode)
	out := PolymorphicTypeConversionModeType(str)
	switch out {
	case PolymorphicTypeConversionTypeModeFast, PolymorphicTypeConversionModeSafe,
		PolymorphicTypeConversionModeOff:
		return out
	}
	panic(fmt.Sprintf("'%s' is not a valid value for PolymorphicTypeConversionMode", str))
}

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
		GetValue: func(c *Container) interface{} {
			return string(c.characterSetClient.Name)
		},
		GetRawValue: func(c *Container) interface{} { return c.characterSetClient },
		SetValue:    setCharacterSetClient,
	}

	definitions[CharacterSetConnection] = &definition{
		Name:             CharacterSetConnection,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue: func(c *Container) interface{} {
			return string(c.characterSetConnection.Name)
		},
		GetRawValue: func(c *Container) interface{} { return c.characterSetConnection },
		SetValue:    setCharacterSetConnection,
	}

	definitions[CharacterSetDatabase] = &definition{
		Name:             CharacterSetDatabase,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue: func(c *Container) interface{} {
			return string(c.characterSetDatabase.Name)
		},
		GetRawValue: func(c *Container) interface{} { return c.characterSetDatabase },
		SetValue:    setCharacterSetDatabase,
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
		GetValue: func(c *Container) interface{} {
			return string(c.collationConnection.Name)
		},
		GetRawValue: func(c *Container) interface{} { return c.collationConnection },
		SetValue:    setCollationConnection,
	}

	definitions[CollationDatabase] = &definition{
		Name:             CollationDatabase,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue: func(c *Container) interface{} {
			return string(c.collationDatabase.Name)
		},
		GetRawValue: func(c *Container) interface{} { return c.collationDatabase },
		SetValue:    setCollationDatabase,
	}

	definitions[CollationServer] = &definition{
		Name:             CollationServer,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue: func(c *Container) interface{} {
			return string(c.collationServer.Name)
		},
		GetRawValue: func(c *Container) interface{} { return c.collationServer },
		SetValue:    setCollationServer,
	}

	definitions[EnableTableAlterations] = &definition{
		Name:             EnableTableAlterations,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope,
		SQLType:          schema.SQLBoolean,
		GetValue:         func(c *Container) interface{} { return c.enableTableAlterations },
		SetValue:         setEnableTableAlterations,
	}

	definitions[FullPushdownExecMode] = &definition{
		Name:             FullPushdownExecMode,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLBoolean,
		GetValue:         func(c *Container) interface{} { return c.FullPushdownExecMode },
		SetValue:         setFullPushdownExecMode,
	}

	definitions[GroupConcatMaxLen] = &definition{
		Name:             GroupConcatMaxLen,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLInt,
		GetValue: func(c *Container) interface{} {
			return c.GroupConcatMaxLen
		},
		SetValue: setGroupConcatMaxLen,
	}

	definitions[InteractiveTimeoutSecs] = &definition{
		Name:             InteractiveTimeoutSecs,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLInt,
		GetValue:         func(c *Container) interface{} { return c.interactiveTimeoutSecs },
		SetValue:         setInteractiveTimeoutSecs,
	}

	definitions[LogLevel] = &definition{
		Name:             LogLevel,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope,
		SQLType:          schema.SQLInt,
		GetValue:         func(c *Container) interface{} { return c.logLevel },
		SetValue:         setLogLevel,
	}

	definitions[MaxNestedTableDepth] = &definition{
		Name:             MaxNestedTableDepth,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope,
		SQLType:          schema.SQLInt,
		GetValue:         func(c *Container) interface{} { return c.maxNestedTableDepth },
		SetValue:         setMaxNestedTableDepth,
	}

	definitions[MaxNumColumnsPerTable] = &definition{
		Name:             MaxNumColumnsPerTable,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope,
		SQLType:          schema.SQLInt,
		GetValue:         func(c *Container) interface{} { return c.maxNumColumnsPerTable },
		SetValue:         setMaxNumColumnsPerTable,
	}

	definitions[MetricsBackend] = &definition{
		Name:             MetricsBackend,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return c.metricsBackend },
		SetValue:         setMetricsBackend,
	}

	definitions[MaxAllowedPacket] = &definition{
		Name:             MaxAllowedPacket,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLInt,
		GetValue:         func(c *Container) interface{} { return c.maxAllowedPacket },
		SetValue:         setMaxAllowedPacket,
	}

	definitions[MaxConnections] = &definition{
		Name:             MaxConnections,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope,
		SQLType:          schema.SQLInt,
		GetValue:         func(c *Container) interface{} { return c.MaxConnections },
		SetValue:         setMaxConnections,
	}

	definitions[MaxTimeMS] = &definition{
		Name:             MaxTimeMS,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLInt,
		GetValue:         func(c *Container) interface{} { return c.MaxTimeMS },
		SetValue:         setMaxTimeMS,
	}

	definitions[MongoDBMaxServerSize] = &definition{
		Name:             MongoDBMaxServerSize,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLUint,
		GetValue:         func(c *Container) interface{} { return c.mongoDBMaxServerSize },
		SetValue:         setMongoDBMaxServerSize,
	}

	definitions[MongoDBMaxConnectionSize] = &definition{
		Name:             MongoDBMaxConnectionSize,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLUint,
		GetValue:         func(c *Container) interface{} { return c.mongoDBMaxConnectionSize },
		SetValue:         setMongoDBMaxConnectionSize,
	}

	definitions[MongoDBMaxStageSize] = &definition{
		Name:             MongoDBMaxStageSize,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLUint,
		GetValue:         func(c *Container) interface{} { return c.mongoDBMaxStageSize },
		SetValue:         setMongoDBMaxStageSize,
	}

	definitions[MongoDBMaxVarcharLength] = &definition{
		Name:             MongoDBMaxVarcharLength,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLUint,
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

	definitions[MongosqldVersion] = &definition{
		Name:             MongosqldVersion,
		Kind:             SystemKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return c.mongosqldVersion },
		SetValue: func(c *Container, v interface{}) error {
			c.mongosqldVersion = v.(string)
			return nil
		},
	}

	definitions[OptimizeCrossJoins] = &definition{
		Name:             OptimizeCrossJoins,
		Kind:             SystemKind,
		AllowedSetScopes: SessionScope,
		SQLType:          schema.SQLBoolean,
		GetValue:         func(c *Container) interface{} { return c.OptimizeCrossJoins },
		SetValue:         setOptimizeCrossJoins,
	}

	definitions[OptimizeEvaluations] = &definition{
		Name:             OptimizeEvaluations,
		Kind:             SystemKind,
		AllowedSetScopes: SessionScope,
		SQLType:          schema.SQLBoolean,
		GetValue:         func(c *Container) interface{} { return c.OptimizeEvaluations },
		SetValue:         setOptimizeEvaluations,
	}

	definitions[OptimizeFiltering] = &definition{
		Name:             OptimizeFiltering,
		Kind:             SystemKind,
		AllowedSetScopes: SessionScope,
		SQLType:          schema.SQLBoolean,
		GetValue:         func(c *Container) interface{} { return c.OptimizeFiltering },
		SetValue:         setOptimizeFiltering,
	}

	definitions[OptimizeInnerJoins] = &definition{
		Name:             OptimizeInnerJoins,
		Kind:             SystemKind,
		AllowedSetScopes: SessionScope,
		SQLType:          schema.SQLBoolean,
		GetValue:         func(c *Container) interface{} { return c.OptimizeInnerJoins },
		SetValue:         setOptimizeInnerJoins,
	}

	definitions[OptimizeSelfJoins] = &definition{
		Name:             OptimizeSelfJoins,
		Kind:             SystemKind,
		AllowedSetScopes: SessionScope,
		SQLType:          schema.SQLBoolean,
		GetValue:         func(c *Container) interface{} { return c.OptimizeSelfJoins },
		SetValue:         setOptimizeSelfJoins,
	}

	definitions[Pushdown] = &definition{
		Name:             Pushdown,
		Kind:             SystemKind,
		AllowedSetScopes: SessionScope,
		SQLType:          schema.SQLBoolean,
		GetValue:         func(c *Container) interface{} { return c.Pushdown },
		SetValue:         setPushdown,
	}

	definitions[OptimizeViewSampling] = &definition{
		Name:             OptimizeViewSampling,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope,
		SQLType:          schema.SQLBoolean,
		GetValue:         func(c *Container) interface{} { return c.OptimizeViewSampling },
		SetValue:         setOptimizeViewSampling,
	}

	definitions[PolymorphicTypeConversionMode] = &definition{
		Name:             PolymorphicTypeConversionMode,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue: func(c *Container) interface{} {
			return c.PolymorphicTypeConversionMode
		},
		GetRawValue: func(c *Container) interface{} { return c.collationServer },
		SetValue:    setPolymorphicTypeConversionMode,
	}

	definitions[RewriteDistinctAsGroup] = &definition{
		Name:             RewriteDistinctAsGroup,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLBoolean,
		GetValue:         func(c *Container) interface{} { return c.rewriteDistinctAsGroup },
		SetValue:         setRewriteDistinctAsGroup,
	}

	definitions[SampleRefreshIntervalSecs] = &definition{
		Name:             SampleRefreshIntervalSecs,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope,
		SQLType:          schema.SQLInt,
		GetValue:         func(c *Container) interface{} { return c.sampleRefreshIntervalSecs },
		SetValue:         setSampleRefreshIntervalSecs,
	}

	definitions[SampleSize] = &definition{
		Name:             SampleSize,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope,
		SQLType:          schema.SQLInt,
		GetValue:         func(c *Container) interface{} { return c.sampleSize },
		SetValue:         setSampleSize,
	}

	definitions[SchemaMappingMode] = &definition{
		Name:             SchemaMappingMode,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return c.SchemaMappingMode },
		SetValue:         setSchemaMappingMode,
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
		SQLType:          schema.SQLUint,
		GetValue:         func(c *Container) interface{} { return c.sqlSelectLimit },
		SetValue:         setSQLSelectLimit,
	}

	definitions[TypeConversionMode] = &definition{
		Name:             TypeConversionMode,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return c.typeConversionMode },
		SetValue:         setTypeConversionMode,
	}

	definitions[Version] = &definition{
		Name:             Version,
		Kind:             SystemKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return c.version },
		SetValue: func(c *Container, v interface{}) error {
			c.version = v.(string)
			return nil
		},
	}

	definitions[VersionComment] = &definition{
		Name:             VersionComment,
		Kind:             SystemKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		SQLType:          schema.SQLVarchar,
		GetValue:         func(c *Container) interface{} { return c.versionComment },
		SetValue: func(c *Container, v interface{}) error {
			c.versionComment = v.(string)
			return nil
		},
	}

	definitions[WaitTimeoutSecs] = &definition{
		Name:             WaitTimeoutSecs,
		Kind:             SystemKind,
		AllowedSetScopes: GlobalScope | SessionScope,
		SQLType:          schema.SQLInt,
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

func setEnableTableAlterations(c *Container, v interface{}) error {
	b, ok := convertBool(v)
	if !ok {
		return wrongTypeError(EnableTableAlterations, v)
	}

	c.enableTableAlterations = b
	return nil
}

func setGroupConcatMaxLen(c *Container, v interface{}) error {
	i, ok := convertInt64(v)
	if !ok {
		return wrongTypeError(GroupConcatMaxLen, v)
	}

	// MySQL's minimum group_concat_max_len value is 4. When a user tries to set the
	// group_concat_max_len system variable to a value < 4, rather than throwing an error,
	// we set the value to the minimum.
	if i < 4 {
		i = 4
	}

	c.GroupConcatMaxLen = i
	return nil
}

func setInteractiveTimeoutSecs(c *Container, v interface{}) error {
	i, ok := convertInt64(v)
	if !ok {
		return wrongTypeError(InteractiveTimeoutSecs, v)
	}

	if i < 1 {
		return mysqlerrors.Defaultf(mysqlerrors.ErWrongValueForVar,
			InteractiveTimeoutSecs, fmt.Sprintf("%v", i))
	}

	c.interactiveTimeoutSecs = i
	return nil
}

func setLogLevel(c *Container, v interface{}) error {
	i, ok := convertInt64(v)
	if !ok {
		return wrongTypeError(LogLevel, v)
	}

	// Changes the global logger's verbosity to whatever the user inputted.
	// Too high and too low values are handled in  log.SetVerbosity
	// The global logger is the parent of every component logger so this
	// changes all of their verbosity's as well
	log.SetVerbosity(log.Verbosity(i))
	normalizedVerbosity := log.NormalizeVerbosityLevel(i)
	c.logLevel = normalizedVerbosity

	return nil
}

func setMaxAllowedPacket(c *Container, v interface{}) error {
	i, ok := convertInt64(v)
	if !ok {
		return wrongTypeError(MaxAllowedPacket, v)
	}

	if i < 1024 || i > 1073741824 {
		return mysqlerrors.Defaultf(mysqlerrors.ErWrongValueForVar,
			MaxAllowedPacket, fmt.Sprintf("%v", i))
	}

	c.maxAllowedPacket = i
	return nil
}

func setMaxConnections(c *Container, v interface{}) error {
	i, ok := convertInt64(v)
	if !ok {
		return wrongTypeError(MaxConnections, v)
	}

	if i < 0 {
		return mysqlerrors.Defaultf(mysqlerrors.ErWrongValueForVar, MaxConnections, i)
	}

	c.MaxConnections = i

	return nil
}

func setMaxNestedTableDepth(c *Container, v interface{}) error {
	i, ok := convertInt64(v)
	if !ok {
		return wrongTypeError(MaxNestedTableDepth, v)
	}

	if i < 0 {
		return mysqlerrors.Defaultf(mysqlerrors.ErWrongValueForVar, MaxNestedTableDepth, i)
	}

	c.maxNestedTableDepth = i
	return nil
}

func setMaxNumColumnsPerTable(c *Container, v interface{}) error {
	i, ok := convertInt64(v)
	if !ok {
		return wrongTypeError(MaxNumColumnsPerTable, v)
	}

	if i <= 0 {
		return mysqlerrors.Defaultf(mysqlerrors.ErWrongValueForVar, MaxNumColumnsPerTable, i)
	}

	c.maxNumColumnsPerTable = i
	return nil
}

func setMaxTimeMS(c *Container, v interface{}) error {
	i, ok := convertInt64(v)
	if !ok {
		return wrongTypeError(MaxTimeMS, v)
	}

	if i < 0 {
		return mysqlerrors.Defaultf(mysqlerrors.ErWrongValueForVar, MaxTimeMS, i)
	}

	c.MaxTimeMS = i

	return nil
}

func setMetricsBackend(c *Container, v interface{}) error {
	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(MetricsBackend, v)
	}

	switch s {
	case NoMetricsBackend, LogMetricsBackend, StitchMetricsBackend:
		c.metricsBackend = s
	default:
		return mysqlerrors.Defaultf(
			mysqlerrors.ErWrongValueForVar,
			MetricsBackend, s,
		)
	}

	return nil
}

func setMongoDBMaxServerSize(c *Container, v interface{}) error {
	i, ok := convertUint64(v)
	if !ok {
		if j, ok := convertInt64(v); ok {
			if j < 0 {
				i = 0
			} else {
				i = uint64(j)
			}
		} else {
			return wrongTypeError(MongoDBMaxServerSize, v)
		}
	}

	c.mongoDBMaxServerSize = i
	return nil
}

func setMongoDBMaxConnectionSize(c *Container, v interface{}) error {
	i, ok := convertUint64(v)
	if !ok {
		if j, ok := convertInt64(v); ok {
			if j < 0 {
				i = 0
			} else {
				i = uint64(j)
			}
		} else {
			return wrongTypeError(MongoDBMaxConnectionSize, v)
		}
	}

	c.mongoDBMaxConnectionSize = i
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

func setFullPushdownExecMode(c *Container, v interface{}) error {
	b, ok := convertBool(v)
	if !ok {
		return wrongTypeError(FullPushdownExecMode, v)
	}

	c.FullPushdownExecMode = b
	return nil
}

func setOptimizeCrossJoins(c *Container, v interface{}) error {
	b, ok := convertBool(v)
	if !ok {
		return wrongTypeError(OptimizeCrossJoins, v)
	}

	c.OptimizeCrossJoins = b
	return nil
}

func setOptimizeEvaluations(c *Container, v interface{}) error {
	b, ok := convertBool(v)
	if !ok {
		return wrongTypeError(OptimizeEvaluations, v)
	}

	c.OptimizeEvaluations = b
	return nil
}

func setOptimizeFiltering(c *Container, v interface{}) error {
	b, ok := convertBool(v)
	if !ok {
		return wrongTypeError(OptimizeFiltering, v)
	}

	c.OptimizeFiltering = b
	return nil
}

func setOptimizeInnerJoins(c *Container, v interface{}) error {
	b, ok := convertBool(v)
	if !ok {
		return wrongTypeError(OptimizeInnerJoins, v)
	}

	c.OptimizeInnerJoins = b
	return nil
}

func setPushdown(c *Container, v interface{}) error {
	b, ok := convertBool(v)
	if !ok {
		return wrongTypeError(Pushdown, v)
	}

	c.Pushdown = b
	return nil
}

func setOptimizeSelfJoins(c *Container, v interface{}) error {
	b, ok := convertBool(v)
	if !ok {
		return wrongTypeError(OptimizeSelfJoins, v)
	}

	c.OptimizeSelfJoins = b
	return nil
}

func setOptimizeViewSampling(c *Container, v interface{}) error {
	b, ok := convertBool(v)
	if !ok {
		return wrongTypeError(OptimizeViewSampling, v)
	}

	c.OptimizeViewSampling = b
	return nil
}

func setPolymorphicTypeConversionMode(c *Container, v interface{}) error {
	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(PolymorphicTypeConversionMode, v)
	}
	switch s {
	case string(PolymorphicTypeConversionTypeModeFast), string(PolymorphicTypeConversionModeSafe),
		string(PolymorphicTypeConversionModeOff):
		c.PolymorphicTypeConversionMode = s
	default:
		return wrongStringValueError(PolymorphicTypeConversionMode, s,
			fmt.Sprintf("'%s'|'%s'|'%s'", PolymorphicTypeConversionTypeModeFast,
				PolymorphicTypeConversionModeSafe,
				PolymorphicTypeConversionModeOff))
	}

	c.PolymorphicTypeConversionMode = s
	return nil
}

func setSampleRefreshIntervalSecs(c *Container, v interface{}) error {
	i, ok := convertInt64(v)
	if !ok {
		return wrongTypeError(SampleRefreshIntervalSecs, v)
	}

	if i < 0 {
		return mysqlerrors.Defaultf(mysqlerrors.ErWrongValueForVar, SampleRefreshIntervalSecs, i)
	}

	c.sampleRefreshIntervalSecs = i
	return nil
}

func setRewriteDistinctAsGroup(c *Container, v interface{}) error {
	b, ok := convertBool(v)
	if !ok {
		return wrongTypeError(RewriteDistinctAsGroup, v)
	}

	c.rewriteDistinctAsGroup = b
	return nil
}

func setSampleSize(c *Container, v interface{}) error {
	i, ok := convertInt64(v)
	if !ok {
		return wrongTypeError(SampleSize, v)
	}

	if i < 0 {
		return mysqlerrors.Defaultf(mysqlerrors.ErWrongValueForVar, SampleSize, i)
	}

	c.sampleSize = i
	return nil
}

func setSchemaMappingMode(c *Container, v interface{}) error {
	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(SchemaMappingMode, v)
	}
	switch s {
	case config.MajorityMappingMode, config.LatticeMappingMode:
		c.SchemaMappingMode = s
	default:
		return wrongStringValueError(SchemaMappingMode, s,
			fmt.Sprintf("'%s'|'%s'", config.MajorityMappingMode, config.LatticeMappingMode))
	}

	c.SchemaMappingMode = s
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

func setTypeConversionMode(c *Container, v interface{}) error {
	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(TypeConversionMode, v)
	}

	switch s {
	case MongoSQLTypeConversionMode, MySQLTypeConversionMode:
		c.typeConversionMode = s
	default:
		return mysqlerrors.Defaultf(mysqlerrors.ErWrongValueForVar, TypeConversionMode, s)
	}

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
		return mysqlerrors.Defaultf(mysqlerrors.ErWrongValueForVar,
			WaitTimeoutSecs, fmt.Sprintf("%v", i))
	}

	c.waitTimeoutSecs = i
	return nil
}
