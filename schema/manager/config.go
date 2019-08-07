package manager

import (
	"time"

	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/sample"
)

// SchemaMode is an enum type that represents the different operating modes for
// the schema manager.
type SchemaMode string

// These constants represent the valid SchemaMode values.
const (
	WriteSchemaMode      SchemaMode = "write"
	FileBasedSchemaMode  SchemaMode = "file"
	StandaloneSchemaMode SchemaMode = "standalone"
	AutoSchemaMode       SchemaMode = "auto"
	CustomSchemaMode     SchemaMode = "custom"
)

// Config is an interface with functions that return various values that are
// used internally by the Manager to control its behavior. It is an interface
// instead of a struct because the values it provides are allowed to change over
// time, and the Manager's behavior will change accordingly.
type Config interface {
	// Mode returns the SchemaMode (Write, Standalone, Auto, or Custom) that the
	// Manager should operate in.
	Mode() SchemaMode
	// RefreshInterval returns the interval at which the Manager should refresh
	// its schema.
	RefreshInterval() time.Duration
	// SampleConfig returns the configuration to use when creating a schema by
	// sampling.
	SampleConfig() sample.Config
	// SchemaName returns the name of the stored schema that should be used when
	// fetching or persisting stored schemas.
	SchemaName() string
	// FileBasedSchema returns the schema that was loaded from a DRDL file, for
	// use in FileBasedSchemaMode.
	FileBasedSchema() *schema.Schema
}

// mongosqldConfig is an implementation of the Config interface that uses
// mongosqld's config file and variable container as the source for its values.
type mongosqldConfig struct {
	cfg             *config.Schema
	vars            *variable.Container
	fileBasedSchema *schema.Schema
}

// NewMongosqldConfig returns a Config interface that uses the provided schema
// config, variable container, and file-based schema as the source for its values.
func NewMongosqldConfig(cfg *config.Schema, vars *variable.Container, fileBasedSchema *schema.Schema) Config {
	return mongosqldConfig{
		cfg:             cfg,
		vars:            vars,
		fileBasedSchema: fileBasedSchema,
	}
}

func (mc mongosqldConfig) Mode() SchemaMode {
	if mc.cfg.WriteMode {
		return WriteSchemaMode
	}
	switch mc.cfg.Stored.Mode {
	case config.CustomStoredSchemaMode:
		return CustomSchemaMode
	case config.AutoStoredSchemaMode:
		return AutoSchemaMode
	case config.NoStoredSchemaMode:
		if mc.cfg.Path == "" {
			return StandaloneSchemaMode
		}
		return FileBasedSchemaMode
	}
	panic("cannot determine schema mode")
}

func (mc mongosqldConfig) RefreshInterval() time.Duration {
	return time.Duration(mc.vars.GetInt64(variable.SampleRefreshIntervalSecs)) * time.Second
}

func (mc mongosqldConfig) SampleConfig() sample.Config {
	return sample.NewMongosqldConfig(mc.cfg, mc.vars)
}

func (mc mongosqldConfig) SchemaName() string {
	return mc.cfg.Stored.Name
}

func (mc mongosqldConfig) FileBasedSchema() *schema.Schema {
	return mc.fileBasedSchema.DeepCopy()
}
