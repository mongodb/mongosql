package mongodrdl

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/option"
	"github.com/10gen/sqlproxy/internal/strutil"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/drdl"
	"github.com/10gen/sqlproxy/schema/sample"

	yaml "github.com/10gen/candiedyaml"
)

const (
	mongoFilterMongoTypeName = "mongo.Filter"
)

// GenerateSchema outputs a DRDL schema according to the provided DrdlOptions.
func GenerateSchema(ctx context.Context, lg log.Logger, opts DrdlOptions) error {
	var err error
	// get the schema bytes
	var schemaBytes []byte
	if opts.Collection() == "" {
		schemaBytes, err = databaseSchema(ctx, lg, opts)
		if err != nil {
			return err
		}
	} else {
		schemaBytes, err = collectionSchema(ctx, lg, opts)
		if err != nil {
			return err
		}
	}

	// Create a writer where we will send the generated schema to.
	writer, err := getOutputWriter(opts.DrdlOutput.Out)
	if err != nil {
		return err
	}
	defer func() { _ = writer.Close() }()

	// write the generated schema
	_, err = writer.Write(schemaBytes)

	// flush any buffered log messages and return
	log.Flush()
	return err
}

// getOutputWriter returns an io.WriteCloser for the file with the provided
// name. The target file, along with any missing directories, will be created.
// If the target filename is empty, os.Stdout will be returned.
func getOutputWriter(out string) (io.WriteCloser, error) {
	if out == "" {
		return os.Stdout, nil
	}

	fileDir := filepath.Dir(out)
	err := os.MkdirAll(fileDir, 0750)
	if err != nil {
		return nil, err
	}

	return os.Create(strutil.ToUniversalPath(out))
}

// collectionSchema returns marshaled bytes of the generated collection schema.
func collectionSchema(ctx context.Context, lg log.Logger, opts DrdlOptions) ([]byte, error) {
	namespaces := []string{fmt.Sprintf("%v.%v",
		opts.DB(),
		opts.Collection(),
	)}
	return schemaForNamespaces(ctx, lg, opts, namespaces)
}

// databaseSchema returns marshaled bytes of the generated database schema.
func databaseSchema(ctx context.Context, lg log.Logger, opts DrdlOptions) ([]byte, error) {
	namespaces := []string{
		fmt.Sprintf("%v.*", opts.DB()),
	}
	return schemaForNamespaces(ctx, lg, opts, namespaces)
}

// schemaForNamespaces returns the YAML marshaled bytes of the sampled
// schema for the namespaces requested.
func schemaForNamespaces(ctx context.Context, lg log.Logger, opts DrdlOptions, ns []string) ([]byte, error) {
	sp, err := newDrdlSessionProvider(opts)
	if err != nil {
		return nil, err
	}
	defer sp.Close()

	sampleCfg := newSampleConfig(opts, ns)
	sampler := sample.NewSampler(sampleCfg, lg, sp)
	sqldSchema, err := sampler.Sample(ctx)
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
						Name: opts.DB(),
					},
				},
			},
		)
	case 1:
	default:
		panic(fmt.Sprintf("expected 1 database found: %v", numDB))
	}

	// Add a custom filter field if needed.
	if opts.DrdlOutput.CustomFilterField != "" {
		customField := opts.DrdlOutput.CustomFilterField
		for _, t := range sqldSchema.Databases()[0].Tables() {
			c := schema.NewColumn(customField, schema.SQLVarchar,
				customField, mongoFilterMongoTypeName, true, option.NoneString())
			t.AddColumn(lg, c, false)
		}
	}

	return sqldSchema.ToDRDL().ToYAML()
}

func newSampleConfig(opts DrdlOptions, ns []string) sample.Config {
	cfg := &config.Schema{
		Stored: config.SchemaStorageOptions{
			Source: "",
		},
		Sample: config.NewSchemaSampleOptions(
			50,                                   // maxNestedTableDepth
			2000,                                 // maxNumColumnsPerTable
			2000,                                 // maxNumFieldsPerCollection
			ns,                                   // namespaces
			true,                                 // optimizeViewSampling
			opts.DrdlOutput.PreJoined,            // preJoin
			config.LatticeMappingMode,            // schemaMappingMode
			opts.DrdlSample.Size,                 // size
			opts.DrdlOutput.UUIDSubtype3Encoding, // uuidSubtype3Encoding
		),
	}
	return sample.NewMongosqldConfig(cfg, nil)
}
