package mongodrdl

import (
	"io"
	"os"
	"path/filepath"

	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/options"
)

// SchemaGenerator is used to configure schema generation.
type SchemaGenerator struct {
	ToolOptions   *options.DrdlOptions
	OutputOptions *options.DrdlOutput
	SampleOptions *options.DrdlSample
	Provider      *mongodb.SessionProvider
	Logger        *log.Logger
}

// Init initializes the schema generator.
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

	var err error
	schemaGen.Provider, err = mongodb.NewDrdlSessionProvider(*schemaGen.ToolOptions)
	if err != nil {
		return err
	}

	return nil
}

// Generate generates the schema and writes it to the configured output.
func (schemaGen *SchemaGenerator) Generate() error {
	writer, err := schemaGen.getOutputWriter()
	if err != nil {
		return err
	}

	if writer == nil {
		writer = os.Stdout
	} else {
		defer writer.Close()
	}

	err = schemaGen.GenerateWithWriter(writer)
	schemaGen.Logger.Flush()
	return err
}

// GenerateWithWriter generates the schema and writes it using the given writer.
func (schemaGen *SchemaGenerator) GenerateWithWriter(writer io.Writer) error {
	var err error
	var b []byte

	switch {
	case schemaGen.ToolOptions.Collection == "":
		b, err = schemaGen.DatabaseSchema()
	default:
		b, err = schemaGen.CollectionSchema()
	}
	if err != nil {
		return err
	}

	_, err = writer.Write(b)
	return err
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
