package mongodrdl

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	yaml "github.com/10gen/candiedyaml"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/options"
	"github.com/10gen/sqlproxy/internal/sample"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/drdl"
)

const (
	mongoFilterMongoTypeName = "mongo.Filter"
)

// GenerateSchema outputs a DRDL schema according to the provided DrdlOptions.
func GenerateSchema(lg log.Logger, opts options.DrdlOptions) error {

	// get the writer where we will send the generated schema
	writer, err := getOutputWriter(opts.DrdlOutput.Out)
	if err != nil {
		return err
	}
	defer func() { _ = writer.Close() }()

	// get the schema bytes
	var schemaBytes []byte
	if opts.Collection == "" {
		schemaBytes, err = databaseSchema(lg, opts)
		if err != nil {
			return err
		}
	} else {
		schemaBytes, err = collectionSchema(lg, opts)
		if err != nil {
			return err
		}
	}

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

	return os.Create(util.ToUniversalPath(out))
}

// collectionSchema returns marshaled bytes of the generated collection schema.
func collectionSchema(lg log.Logger, opts options.DrdlOptions) ([]byte, error) {
	namespaces := []string{fmt.Sprintf("%v.%v",
		opts.DrdlNamespace.DB,
		opts.DrdlNamespace.Collection,
	)}
	return schemaForNamespaces(lg, opts, namespaces)
}

// databaseSchema returns marshaled bytes of the generated database schema.
func databaseSchema(lg log.Logger, opts options.DrdlOptions) ([]byte, error) {
	namespaces := []string{
		fmt.Sprintf("%v.*", opts.DrdlNamespace.DB),
	}
	return schemaForNamespaces(lg, opts, namespaces)
}

// schemaForNamespaces returns the YAML marshaled bytes of the sampled
// schema for the namespaces requested.
func schemaForNamespaces(lg log.Logger, opts options.DrdlOptions, ns []string) ([]byte, error) {
	session, err := getSession(opts)
	if err != nil {
		return nil, err
	}
	defer func() { _ = session.Close() }()

	cfg := &config.SchemaSampleOptions{
		Size:                   opts.DrdlSample.Size,
		Namespaces:             ns,
		UUIDSubtype3Encoding:   opts.DrdlOutput.UUIDSubtype3Encoding,
		PreJoin:                opts.DrdlOutput.PreJoined,
		SchemaMappingHeuristic: config.LatticeMappingMode,
	}

	sqldSchema, _, err := sample.Schema(sample.NewSchemaSampleOptions(cfg),
		"mongodrdl", session, lg)
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
						Name: opts.DrdlNamespace.DB,
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
				customField, mongoFilterMongoTypeName)
			t.AddColumn(lg, c, false)
		}
	}

	return sqldSchema.ToDRDL().ToYAML()
}

// getSession returns a mongodb.Session with the connection options specified
// by the provided DrdlOptions.
func getSession(opts options.DrdlOptions) (*mongodb.Session, error) {
	sp, err := mongodb.NewDrdlSessionProvider(opts)
	if err != nil {
		return nil, err
	}

	session, err := sp.Session(context.Background())
	if err != nil {
		return nil, fmt.Errorf("can't create session: %v", err)
	}

	return session, nil
}
