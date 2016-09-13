package mongodrdl

import (
	"fmt"
	yaml "github.com/10gen/candiedyaml"
	"github.com/10gen/sqlproxy/mongodrdl/mongo"
	"github.com/10gen/sqlproxy/mongodrdl/relational"
	"github.com/mongodb/mongo-tools/common/db"
	"github.com/mongodb/mongo-tools/common/options"
	"github.com/mongodb/mongo-tools/common/util"
	"io"
	"os"
	"path/filepath"
)

type SchemaGenerator struct {
	ToolOptions   *options.ToolOptions
	OutputOptions *OutputOptions
	SampleOptions *SampleOptions
	provider      *db.SessionProvider
}

type Schema struct {
	Databases []*relational.Database `yaml:"schema"`
}

func NewSchemaGenerator(db, collection, outputFile string, sslOptions *options.SSL) *SchemaGenerator {
	gen := &SchemaGenerator{
		ToolOptions: &options.ToolOptions{
			Namespace: &options.Namespace{
				DB:         db,
				Collection: collection,
			},
			SSL: sslOptions,
		},
		OutputOptions: &OutputOptions{
			Out: outputFile,
		},
		SampleOptions: &SampleOptions{SampleSize: 1000},
	}

	gen.Init()

	return gen
}

func (schemaGen *SchemaGenerator) Init() error {
	err := schemaGen.validateOptions()
	if err != nil {
		return err
	}

	if schemaGen.OutputOptions.Out == "" {
		schemaGen.OutputOptions.Out = "-"
	}
	if schemaGen.ToolOptions.Connection == nil {
		schemaGen.ToolOptions.Connection = &options.Connection{}
	}
	if schemaGen.ToolOptions.Auth == nil {
		schemaGen.ToolOptions.Auth = &options.Auth{}
	}
	if schemaGen.ToolOptions.SSL == nil {
		schemaGen.ToolOptions.SSL = &options.SSL{}
	}

	mongo.UUIDSubtype3Encoding = schemaGen.OutputOptions.UUIDSubtype3Encoding

	schemaGen.provider, err = db.NewSessionProvider(*schemaGen.ToolOptions)
	if err != nil {
		return err
	}

	schemaGen.provider.SetFlags(db.DisableSocketTimeout)

	return nil
}

func (schemaGen *SchemaGenerator) Generate() (*Schema, error) {
	var err error
	var database *relational.Database

	writer, err := schemaGen.getOutputWriter()
	if err != nil {
		return nil, err
	}
	if writer == nil {
		writer = os.Stdout
	} else {
		defer writer.Close()
	}

	switch {
	case schemaGen.ToolOptions.Collection == "":
		database, err = schemaGen.ExportSchemaForDatabase()
	default:
		database, err = schemaGen.ExportSchemaForCollection()
	}
	if err != nil {
		return nil, err
	}

	if schemaGen.OutputOptions.CustomFilterField != "" {
		for _, t := range database.Tables {
			t.AddColumn(schemaGen.OutputOptions.CustomFilterField, relational.MongoFilterMongoTypeName)
		}
	}

	schema := &Schema{
		Databases: []*relational.Database{database},
	}

	bytes, err := yaml.Marshal(schema)
	if err != nil {
		return nil, err
	}

	_, err = writer.Write(bytes)
	if err != nil {
		return nil, err
	}
	return schema, nil
}

func (schemaGen *SchemaGenerator) validateOptions() error {
	switch {
	case schemaGen.ToolOptions.Namespace.DB == "":
		return fmt.Errorf("cannot export a schema without a specified database")
	case schemaGen.ToolOptions.Namespace.DB == "" && schemaGen.ToolOptions.Namespace.Collection != "":
		return fmt.Errorf("cannot export a schema for a collection without a specified database")
	}
	return nil
}

func (schemaGen *SchemaGenerator) getOutputWriter() (io.WriteCloser, error) {
	var writer io.WriteCloser

	if schemaGen.OutputOptions.Out != "-" {
		fileDir := filepath.Dir(schemaGen.OutputOptions.Out)
		err := os.MkdirAll(fileDir, 0750)
		if err != nil {
			return nil, err
		}

		writer, err = os.Create(util.ToUniversalPath(schemaGen.OutputOptions.Out))
		if err != nil {
			return nil, err
		}
		return writer, err
	}

	return writer, nil
}
