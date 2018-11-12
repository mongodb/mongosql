package variable

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/10gen/sqlproxy/internal/collation"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
)

// These are the permitted values for the type_conversion_mode variable.
const (
	MongoSQLTypeConversionMode = "mongosql"
	MySQLTypeConversionMode    = "mysql"
)

// These are the permitted values for the metrics_backend variable.
const (
	NoMetricsBackend     = "off"
	LogMetricsBackend    = "log"
	StitchMetricsBackend = "stitch"
)

// PolymorphicTypeConversionModeType is an enum of PolymorphicTypeConversionMode values.
type PolymorphicTypeConversionModeType string

const (
	// PolymorphicTypeConversionModeSafe sets the polymorphic_type_conversion_mode to safe,
	// which inserts SQLConvertExprs to the column type on any column from a MongoSourceStage.
	// It is "safe", because it should protect against any unsampled data types on versions
	// of MongoDB server that support $convert.
	PolymorphicTypeConversionModeSafe PolymorphicTypeConversionModeType = "safe"
	// PolymorphicTypeConversionTypeModeFast sets the polymorphic_type_conversion_mode to fast.
	// This mode will insert SQLConvertExprs on any column from a MongoSourceStage that was
	// sampled with more than one type, but not on others.
	PolymorphicTypeConversionTypeModeFast PolymorphicTypeConversionModeType = "fast"
	// PolymorphicTypeConversionModeOff sets the polymorphic_type_conversion_mode to off.
	// No extra SQLConvertExprs are inserted in this mode.
	PolymorphicTypeConversionModeOff PolymorphicTypeConversionModeType = "off"
)

const (
	defaultMetricsBackend     = NoMetricsBackend
	defaultTypeConversionMode = MongoSQLTypeConversionMode
)

// Container holds variables based on a scope.
type Container struct {
	lock   sync.RWMutex
	scope  Scope
	parent *Container

	// userValues is storage for user variables
	userValues map[Name]interface{}

	// Backing storage for non-user MySQL system variables below.
	autoCommit             bool
	characterSetClient     *collation.Charset
	characterSetConnection *collation.Charset
	characterSetDatabase   *collation.Charset
	characterSetResults    *collation.Charset
	collationConnection    *collation.Collation
	collationDatabase      *collation.Collation
	collationServer        *collation.Collation
	interactiveTimeoutSecs int64
	maxAllowedPacket       int64
	MaxConnections         int64
	MaxTimeMS              int64
	socket                 string
	sqlAutoIsNull          bool
	sqlSelectLimit         uint64
	version                string
	versionComment         string
	waitTimeoutSecs        int64

	// Backing storage for non-user MySQL status variables below.
	BytesReceived    *uint64
	BytesSent        *uint64
	Connections      *uint32
	Queries          *uint64
	StartTime        time.Time
	ThreadsConnected *uint32

	// Backing storage for mongosqld-defined variables below.
	enableTableAlterations        bool
	FullPushdownExecMode          bool
	GroupConcatMaxLen             int64
	logLevel                      int64
	maxNestedTableDepth           int64
	maxNumColumnsPerTable         int64
	metricsBackend                string
	mongoDBMaxServerSize          uint64
	mongoDBMaxConnectionSize      uint64
	mongoDBMaxStageSize           uint64
	mongoDBMaxVarcharLength       uint16
	MongoDBInfo                   *mongodb.Info
	mongoDBVersionCompatibility   string
	OptimizeCrossJoins            bool
	OptimizeEvaluations           bool
	OptimizeFiltering             bool
	OptimizeInnerJoins            bool
	OptimizeSelfJoins             bool
	OptimizeViewSampling          bool
	PolymorphicTypeConversionMode string
	Pushdown                      bool
	sampleRefreshIntervalSecs     int64
	sampleSize                    int64
	SchemaMappingHeuristic        string
	rewriteDistinctAsGroup        bool
	typeConversionMode            string

	AllocatedMemory func() uint64
}

// NewGlobalContainer creates a container with a GlobalScope.
func NewGlobalContainer(cfg *config.Config) *Container {

	// Initialize server status variables here
	bytesReceived := uint64(0)
	bytesSent := uint64(0)
	connections := uint32(0)
	queries := uint64(0)
	startTime := time.Now()
	threadsConnected := uint32(0)

	// These variables' default values can be set in the
	// setParameter section of the config.
	enableTableAlterations := false
	metricsBackend := defaultMetricsBackend
	optimizeEvaluations := true
	optimizeCrossJoins := true
	optimizeInnerJoins := true
	optimizeFiltering := true
	optimizeSelfJoins := true
	optimizeViewSampling := true
	polymorphicTypeConversionMode := string(PolymorphicTypeConversionModeOff)
	pushdown := true
	rewriteDistinctAsGroup := false
	typeConversionMode := defaultTypeConversionMode

	// These variables' default values can be set via other
	// config flags/config file options.
	mappingHeuristic := string(config.LatticeMappingMode)
	sampleSize := int64(config.DefaultSampleSize)
	sampleRefreshIntervalSecs := int64(config.DefaultSampleRefreshIntervalSecs)
	maxNumColumnsPerTable := int64(config.DefaultMaxNumColumnsPerTable)
	maxNestedTableDepth := int64(config.DefaultMaxNestedTableDepth)

	logLevel := int64(0)
	if cfg != nil {
		logLevel = int64(cfg.SystemLog.Verbosity)
		if cfg.SystemLog.Quiet {
			logLevel = -1
		}
	}

	if cfg != nil {
		// defaults from SetParameter config section
		enableTableAlterations = cfg.SetParameter.EnableTableAlterations
		metricsBackend = cfg.SetParameter.MetricsBackend
		optimizeEvaluations = cfg.SetParameter.OptimizeEvaluations
		optimizeCrossJoins = cfg.SetParameter.OptimizeCrossJoins
		optimizeInnerJoins = cfg.SetParameter.OptimizeInnerJoins
		optimizeFiltering = cfg.SetParameter.OptimizeFiltering
		optimizeSelfJoins = cfg.SetParameter.OptimizeSelfJoins
		optimizeViewSampling = cfg.SetParameter.OptimizeViewSampling
		polymorphicTypeConversionMode = cfg.SetParameter.PolymorphicTypeConversionMode
		pushdown = cfg.SetParameter.Pushdown
		rewriteDistinctAsGroup = cfg.SetParameter.RewriteDistinctAsGroup
		typeConversionMode = cfg.SetParameter.TypeConversionMode

		// defaults from other config sections
		mappingHeuristic = string(cfg.Schema.Sample.SchemaMappingHeuristic)
		sampleSize = cfg.Schema.Sample.Size
		sampleRefreshIntervalSecs = cfg.Schema.Sample.RefreshIntervalSecs
		maxNumColumnsPerTable = cfg.Schema.Sample.MaxNumColumnsPerTable
		maxNestedTableDepth = cfg.Schema.Sample.MaxNestedTableDepth
	}
	logLevel = log.NormalizeVerbosityLevel(logLevel)

	container := &Container{
		scope: GlobalScope,

		// Default values for non-user MySQL system variables below.
		autoCommit:             true,
		characterSetClient:     collation.DefaultCharset,
		characterSetConnection: collation.DefaultCharset,
		characterSetDatabase:   collation.DefaultCharset,
		characterSetResults:    collation.DefaultCharset,
		collationConnection:    collation.Default,
		collationDatabase:      collation.Default,
		collationServer:        collation.Default,
		GroupConcatMaxLen:      1024,
		interactiveTimeoutSecs: 28800,
		maxAllowedPacket:       config.DefaultMaxAllowedPacket,
		MaxConnections:         0, // represents unlimited connections
		MaxTimeMS:              0, // A value of 0 represents no timeout is enabled.
		socket:                 "",
		sqlAutoIsNull:          false,
		sqlSelectLimit:         math.MaxUint64,
		version:                "5.7.12",
		versionComment:         "mongosqld " + config.VersionStr,
		waitTimeoutSecs:        28800,

		// Default values for non-user MySQL status variables below.
		BytesReceived:    &bytesReceived,
		BytesSent:        &bytesSent,
		Connections:      &connections,
		Queries:          &queries,
		StartTime:        startTime,
		ThreadsConnected: &threadsConnected,

		// Default values for mongosqld-defined variables.
		enableTableAlterations:        enableTableAlterations,
		FullPushdownExecMode:          false,
		logLevel:                      logLevel,
		maxNumColumnsPerTable:         maxNumColumnsPerTable,
		maxNestedTableDepth:           maxNestedTableDepth,
		metricsBackend:                metricsBackend,
		mongoDBMaxServerSize:          0,
		mongoDBMaxConnectionSize:      0,
		mongoDBMaxStageSize:           0,
		mongoDBMaxVarcharLength:       math.MaxUint16,
		MongoDBInfo:                   nil,
		mongoDBVersionCompatibility:   "",
		OptimizeEvaluations:           optimizeEvaluations,
		OptimizeCrossJoins:            optimizeCrossJoins,
		OptimizeInnerJoins:            optimizeInnerJoins,
		OptimizeFiltering:             optimizeFiltering,
		OptimizeSelfJoins:             optimizeSelfJoins,
		OptimizeViewSampling:          optimizeViewSampling,
		PolymorphicTypeConversionMode: polymorphicTypeConversionMode,
		Pushdown:                      pushdown,
		sampleRefreshIntervalSecs:     sampleRefreshIntervalSecs,
		sampleSize:                    sampleSize,
		SchemaMappingHeuristic:        mappingHeuristic,
		rewriteDistinctAsGroup:        rewriteDistinctAsGroup,
		typeConversionMode:            typeConversionMode,

		AllocatedMemory: func() uint64 { return 0 },
	}

	// Initializing Global Container
	if cfg != nil {
		container.mongoDBMaxServerSize = cfg.Runtime.Memory.MaxPerServer
		container.mongoDBMaxConnectionSize = cfg.Runtime.Memory.MaxPerConnection
		container.mongoDBMaxStageSize = cfg.Runtime.Memory.MaxPerStage
		container.mongoDBMaxVarcharLength = cfg.Schema.MaxVarcharLength
		container.mongoDBVersionCompatibility = cfg.MongoDB.VersionCompatibility
	}

	return container
}

// NewSessionContainer creates a container with a SessionScope.
func NewSessionContainer(global *Container) *Container {
	if global == nil {
		panic("global cannot be nil")
	}

	c := &Container{
		scope:           SessionScope,
		parent:          global,
		userValues:      make(map[Name]interface{}),
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
func (c *Container) List(scope Scope, kind Kind) []Value {
	if c.scope == GlobalScope && kind == SystemKind {
		c.lock.RLock()
		defer c.lock.RUnlock()
	}

	if kind == UserKind {
		if scope != SessionScope {
			panic("cannot get user variables from a global scope")
		}

		var values []Value
		for k, v := range c.userValues {
			values = append(values, Value{
				Name:    k,
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

	panic(fmt.Sprintf("illegal scope %v", scope))
}

// Get gets the value of the variable with the specified name, scope, and kind.
func (c *Container) Get(name Name, scope Scope, kind Kind) (Value, error) {
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

		return Value{
			Name:    name,
			Kind:    kind,
			SQLType: schema.SQLNone,
			Value:   v,
		}, nil
	}

	if c.scope == scope {
		if def, ok := definitions[lowerName]; ok && def.Kind == kind {
			if def.GetRawValue != nil {
				return Value{
					Name:     name,
					Kind:     def.Kind,
					SQLType:  def.SQLType,
					Value:    def.GetValue(c),
					RawValue: def.GetRawValue(c),
				}, nil
			}
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

	return Value{}, mysqlerrors.Defaultf(mysqlerrors.ErUnknownSystemVariable, name)
}

// GetBool gets the value of the variable with the specified name for system variable of
// boolean type.
func (c *Container) GetBool(name Name) bool {
	value, err := c.Get(name, c.scope, SystemKind)
	if err != nil {
		panic(fmt.Sprintf("cannot get boolean system variable %v: %v", name, err))
	}

	return value.Value.(bool)
}

// GetCharset gets the value of the variable with the specified name for system variable of
// collation.Charset type.
func (c *Container) GetCharset(name Name) *collation.Charset {
	value, err := c.Get(name, c.scope, SystemKind)
	if err != nil {
		panic(fmt.Sprintf("cannot get collation.Charset system variable %v: %v", name, err))
	}

	return value.RawValue.(*collation.Charset)
}

// GetCollation gets the value of the variable with the specified name for system variable of
// collation.Collation type.
func (c *Container) GetCollation(name Name) *collation.Collation {
	value, err := c.Get(name, c.scope, SystemKind)
	if err != nil {
		panic(fmt.Sprintf("cannot get collation.Collation system variable %v: %v", name, err))
	}

	return value.RawValue.(*collation.Collation)
}

// GetInt64 gets the value of the variable with the specified name for system variable of
// int64 type.
func (c *Container) GetInt64(name Name) int64 {
	value, err := c.Get(name, c.scope, SystemKind)
	if err != nil {
		panic(fmt.Sprintf("cannot get int64 system variable %v: %v", name, err))
	}

	return value.Value.(int64)
}

// GetString gets the value of the variable with the specified name for system variable of
// string type.
func (c *Container) GetString(name Name) string {
	value, err := c.Get(name, c.scope, SystemKind)
	if err != nil {
		panic(fmt.Sprintf("cannot get string system variable %v: %v", name, err))
	}

	return value.Value.(string)
}

// GetUInt16 gets the value of the variable with the specified name for system variable of
// uint16 type.
func (c *Container) GetUInt16(name Name) uint16 {
	value, err := c.Get(name, c.scope, SystemKind)
	if err != nil {
		panic(fmt.Sprintf("cannot get uint16 system variable %v: %v", name, err))
	}

	return value.Value.(uint16)
}

// GetUInt64 gets the value of the variable with the specified name for system variable of
// uint64 type.
func (c *Container) GetUInt64(name Name) uint64 {
	value, err := c.Get(name, c.scope, SystemKind)
	if err != nil {
		panic(fmt.Sprintf("cannot get uint64 system variable %v: %v", name, err))
	}

	return value.Value.(uint64)
}

// Set sets the value of a variable with the specified name, scope, and kind.
func (c *Container) Set(name Name, scope Scope, kind Kind, value interface{}) error {
	if c.scope == GlobalScope && kind == SystemKind {
		c.lock.Lock()
		defer c.lock.Unlock()
	}

	lowerName := Name(strings.ToLower(string(name)))
	if kind == UserKind {
		if scope != SessionScope {
			panic(fmt.Sprintf("cannot set user variable: %v on a global scope: %v", name, scope))
		}

		c.userValues[lowerName] = value
		return nil
	}

	if kind == StatusKind {
		return mysqlerrors.Defaultf(mysqlerrors.ErUnknownSystemVariable, name)
	}

	def, ok := definitions[lowerName]
	if !ok {
		return mysqlerrors.Defaultf(mysqlerrors.ErUnknownSystemVariable, name)
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

	if def.SetValue == nil {
		return mysqlerrors.Defaultf(mysqlerrors.ErVariableIsReadonly, kindToString(kind), name)
	}

	if fmt.Sprintf("%v", value) == "default" {
		value, err := NewGlobalContainer(nil).Get(name, GlobalScope, kind)
		if err != nil {
			return err
		}
		return c.Set(name, scope, kind, value.Value)
	}

	if c.scope == scope {
		return def.SetValue(c, value)
	} else if c.parent != nil {
		return c.parent.Set(name, scope, kind, value)
	}

	panic(fmt.Sprintf("illegal scope %v", scope))
}

// SetSystemVariable sets the value of the variable with the specified name for system variable.
func (c *Container) SetSystemVariable(name Name, value interface{}) {
	err := c.Set(name, c.scope, SystemKind, value)
	if err != nil {
		panic(fmt.Sprintf("cannot set system variable %v: %v", name, err))
	}
}
