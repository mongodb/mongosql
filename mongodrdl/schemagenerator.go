package mongodrdl

import (
	yaml "github.com/10gen/candiedyaml"
	"github.com/10gen/sqlproxy/client"
	"github.com/10gen/sqlproxy/mongodrdl/mongo"
	"github.com/10gen/sqlproxy/mongodrdl/relational"
	"github.com/10gen/sqlproxy/options"
	"github.com/10gen/sqlproxy/util"
	"io"
	"os"
	"path/filepath"
)

type SchemaGenerator struct {
	ToolOptions   *options.DrdlOptions
	OutputOptions *options.DrdlOutput
	SampleOptions *options.DrdlSample
	provider      *client.SessionProvider
}

type Schema struct {
	Databases []*relational.Database `yaml:"schema"`
}

func NewSchemaGenerator(db, collection, outputFile string, sslOptions *options.DrdlSSL) *SchemaGenerator {
	gen := &SchemaGenerator{
		ToolOptions: &options.DrdlOptions{
			DrdlNamespace: &options.DrdlNamespace{
				DB:         db,
				Collection: collection,
			},
			DrdlSSL: sslOptions,
		},
		OutputOptions: &options.DrdlOutput{
			Out: outputFile,
		},
		SampleOptions: &options.DrdlSample{SampleSize: 1000},
	}

	gen.Init()

	return gen
}

func (schemaGen *SchemaGenerator) Init() error {
	if schemaGen.OutputOptions.Out == "" {
		schemaGen.OutputOptions.Out = "-"
	}
	if schemaGen.ToolOptions.DrdlConnection == nil {
		schemaGen.ToolOptions.DrdlConnection = &options.DrdlConnection{}
	}
	if schemaGen.ToolOptions.DrdlAuth == nil {
		schemaGen.ToolOptions.DrdlAuth = &options.DrdlAuth{}
	}
	if schemaGen.ToolOptions.DrdlSSL == nil {
		schemaGen.ToolOptions.DrdlSSL = &options.DrdlSSL{}
	}

	mongo.UUIDSubtype3Encoding = schemaGen.OutputOptions.UUIDSubtype3Encoding

	var err error

	schemaGen.provider, err = client.NewDrdlSessionProvider(*schemaGen.ToolOptions)
	if err != nil {
		return err
	}

	schemaGen.provider.SetFlags(client.DisableSocketTimeout)

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
