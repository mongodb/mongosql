package mongodrdl

import (
	"context"
	"fmt"

	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/sample"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/drdl"

	yaml "github.com/10gen/candiedyaml"
)

const (
	mongoFilterMongoTypeName = "mongo.Filter"
)

// Connect returns a connection to the configured MongoDB cluster.
func (schemaGen *SchemaGenerator) Connect() (*mongodb.Session, error) {
	session, err := schemaGen.Provider.Session(context.Background())
	if err != nil {
		return nil, fmt.Errorf("can't create session: %v", err)
	}
	return session, nil
}

// CollectionSchema returns marshaled bytes of the generated collection schema.
func (schemaGen *SchemaGenerator) CollectionSchema() ([]byte, error) {
	namespaces := []string{fmt.Sprintf("%v.%v",
		schemaGen.ToolOptions.DrdlNamespace.DB,
		schemaGen.ToolOptions.DrdlNamespace.Collection,
	)}
	return schemaGen.schemaForNamespaces(namespaces)
}

// DatabaseSchema returns marshaled bytes of the generated database schema.
func (schemaGen *SchemaGenerator) DatabaseSchema() ([]byte, error) {
	namespaces := []string{
		fmt.Sprintf("%v.*", schemaGen.ToolOptions.DrdlNamespace.DB),
	}
	return schemaGen.schemaForNamespaces(namespaces)
}

// addCustomFilterField adds the custom filter field
// to each table in the given schema.
func (schemaGen *SchemaGenerator) addCustomFilterField(s *schema.Schema) {
	customField := schemaGen.OutputOptions.CustomFilterField
	for _, t := range s.Databases()[0].Tables() {
		c := schema.NewColumn(customField, schema.SQLVarchar,
			customField, mongoFilterMongoTypeName)
		t.AddColumn(schemaGen.Logger, c, false)
	}
}

// schemaForNamespaces returns the YAML marshaled bytes of the sampled
// schema for the namespaces requested.
func (schemaGen *SchemaGenerator) schemaForNamespaces(namespaces []string) ([]byte, error) {
	session, err := schemaGen.Connect()
	if err != nil {
		return nil, err
	}
	defer func() { _ = session.Close() }()

	cfg := &config.SchemaSampleOptions{
		Size:                 schemaGen.SampleOptions.Size,
		Namespaces:           namespaces,
		UUIDSubtype3Encoding: schemaGen.OutputOptions.UUIDSubtype3Encoding,
		PreJoined:            schemaGen.OutputOptions.PreJoined,
	}

	sqldSchema, _, err := sample.Schema(cfg, "mongodrdl", session, schemaGen.Logger)
	if err != nil {
		return nil, err
	}

	numDB := len(sqldSchema.Databases())
	switch numDB {
	case 0:
		return yaml.Marshal(
			&drdl.Schema{
				Databases: []*drdl.Database{
					{
						Name: schemaGen.ToolOptions.DrdlNamespace.DB,
					},
				},
			},
		)
	case 1:
	default:
		panic(fmt.Sprintf("expected 1 database found: %v", numDB))
	}

	// Add a custom filter field if needed.
	if schemaGen.OutputOptions.CustomFilterField != "" {
		schemaGen.addCustomFilterField(sqldSchema)
	}

	return sqldSchema.ToDRDL().ToYAML()
}
