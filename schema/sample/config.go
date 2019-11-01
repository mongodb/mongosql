package sample

import (
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/internal/config"
)

// Config is an interface with functions that return various values that are
// used internally by a Sampler to control its behavior. It is an interface
// instead of a struct because the values it provides are allowed to change over
// time, and we want the Sampler to use the most up to date options each time it
// accesses them.
type Config interface {
	Source() string
	Size() int64
	MaxNumFieldsPerCollection() int64
	MaxNestedTableDepth() int64
	OptimizeViewSampling() bool
	PreJoin() bool
	Namespaces() []string
	UUIDSubtype3Encoding() string
	SchemaMappingMode() config.MappingMode
	SchemaMode() config.StoredSchemaMode
	WriteMode() bool
}

// NewMongosqldConfig returns a new config that sources its options from a mongosqld
// Schema configuration.
func NewMongosqldConfig(cfg *config.Schema, vars *variable.Container) Config {
	return mongosqldConfig{
		cfg:  cfg,
		vars: vars,
	}
}

type mongosqldConfig struct {
	cfg  *config.Schema
	vars *variable.Container
}

func (m mongosqldConfig) Source() string {
	return m.cfg.Stored.Source
}

func (m mongosqldConfig) Size() int64 {
	if m.vars != nil {
		return m.vars.GetInt64(variable.SampleSize)
	}
	return m.cfg.Sample.Size
}

func (m mongosqldConfig) MaxNumFieldsPerCollection() int64 {
	if m.vars != nil {
		return m.vars.GetInt64(variable.MaxNumFieldsPerCollection)
	}
	return m.cfg.Sample.MaxNumFieldsPerCollection
}

func (m mongosqldConfig) MaxNestedTableDepth() int64 {
	if m.vars != nil {
		return m.vars.GetInt64(variable.MaxNestedTableDepth)
	}
	return m.cfg.Sample.MaxNestedTableDepth
}

func (m mongosqldConfig) OptimizeViewSampling() bool {
	if m.vars != nil {
		return m.vars.GetBool(variable.OptimizeViewSampling)
	}
	return m.cfg.Sample.OptimizeViewSampling
}

func (m mongosqldConfig) PreJoin() bool {
	return m.cfg.Sample.PreJoin
}

func (m mongosqldConfig) Namespaces() []string {
	return m.cfg.Sample.Namespaces
}

func (m mongosqldConfig) UUIDSubtype3Encoding() string {
	return m.cfg.Sample.UUIDSubtype3Encoding
}

func (m mongosqldConfig) SchemaMappingMode() config.MappingMode {
	if m.vars != nil {
		return config.GetMappingMode(m.vars.GetString(variable.SchemaMappingMode))
	}
	return m.cfg.Sample.SchemaMappingMode
}

func (m mongosqldConfig) SchemaMode() config.StoredSchemaMode {
	return m.cfg.Stored.Mode
}

func (m mongosqldConfig) WriteMode() bool {
	return m.cfg.WriteMode
}
