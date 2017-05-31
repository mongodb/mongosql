package mongodrdl

import (
	"io"
	"os"
	"path/filepath"

	yaml "github.com/10gen/candiedyaml"

	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/mongodrdl/mongo"
	"github.com/10gen/sqlproxy/mongodrdl/relational"
	"github.com/10gen/sqlproxy/options"
)

type SchemaGenerator struct {
	ToolOptions   *options.DrdlOptions
	OutputOptions *options.DrdlOutput
	SampleOptions *options.DrdlSample
	Provider      *mongodb.SessionProvider
	Logger        *log.Logger
}

type Schema struct {
	Databases []*relational.Database `yaml:"schema"`
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
	schemaGen.Provider, err = mongodb.NewDrdlSessionProvider(*schemaGen.ToolOptions)
	if err != nil {
		return err
	}

	return nil
}

func (schemaGen *SchemaGenerator) Generate() (*Schema, error) {
	writer, err := schemaGen.getOutputWriter()
	if err != nil {
		return nil, err
	}

	if writer == nil {
		writer = os.Stdout
	} else {
		defer writer.Close()
	}

	return schemaGen.generate(writer)
}

func (schemaGen *SchemaGenerator) generate(writer io.Writer) (*Schema, error) {
	var err error
	var database *relational.Database

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

func (schemaGen *SchemaGenerator) GenerateWithWriter(writer io.Writer) (*Schema, error) {
	return schemaGen.generate(writer)
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
