package drdl

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"golang.org/x/text/encoding/unicode"

	yaml "github.com/10gen/candiedyaml"
	"github.com/10gen/sqlproxy/internal/bsonutil"

	oldbson "github.com/10gen/mongo-go-driver/bson"

	"go.mongodb.org/mongo-driver/bson"
)

// These are components of a Byte Order Mark which appears at the
// beginning of a UTF-16 file.
const (
	utfBOMLittle = 0xFF
	utfBOMBig    = 0xFE
)

// Schema represents a DRDL schema definition.
type Schema struct {
	Databases []*Database `yaml:"schema" bson:"databases"`
}

// Database represents a DRDL database definition.
type Database struct {
	Name   string   `yaml:"db" bson:"name"`
	Tables []*Table `yaml:"tables" bson:"tables"`
}

// Table represents a DRDL table definition.
type Table struct {
	SQLName   string    `yaml:"table"`
	MongoName string    `yaml:"collection"`
	Pipeline  []bson.D  `yaml:"pipeline"`
	Columns   []*Column `yaml:"columns"`
}

// Column represents a DRDL column definition.
type Column struct {
	MongoName string `yaml:"Name" bson:"mongo_name"`
	MongoType string `yaml:"MongoType" bson:"mongo_type"`
	SQLName   string `yaml:"SqlName" bson:"sql_name"`
	SQLType   string `yaml:"SqlType" bson:"sql_type"`
}

// NewFromDir returns a new Schema loaded from DRDL files in the specified
// directory.
func NewFromDir(path string) (*Schema, error) {
	schema := &Schema{}
	err := schema.LoadDir(path)
	if err != nil {
		return nil, err
	}
	return schema, nil
}

// NewFromFile returns a new Schema loaded from the specified file.
func NewFromFile(path string) (*Schema, error) {
	schema := &Schema{}
	err := schema.LoadFile(path)
	if err != nil {
		return nil, err
	}
	return schema, nil
}

// NewFromBytes returns a new Schema loaded from the provided byte slice.
func NewFromBytes(data []byte) (*Schema, error) {
	schema := &Schema{}
	err := schema.Load(data)
	if err != nil {
		return nil, err
	}
	return schema, nil
}

// LoadDir loads schema definitions from all files inside the given directory
// into this Schema.
func (s *Schema) LoadDir(root string) error {
	files, err := ioutil.ReadDir(root)
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if strings.ContainsAny(f.Name(), "#~") {
			continue
		}

		fullPath := filepath.Join(root, f.Name())
		err = s.LoadFile(fullPath)
		if err != nil {
			return fmt.Errorf("in schema file %v: %v", fullPath, err)
		}
	}
	return nil
}

// LoadFile loads the schema definition from the specified file into this Schema.
func (s *Schema) LoadFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return s.Load(data)
}

// Load loads the schema definition from the provided byte slice into this Schema.
func (s *Schema) Load(data []byte) error {

	var err error
	// Convert from little endian if byte order mark is in little->big order.
	if len(data) > 2 && data[0] == utfBOMLittle && data[1] == utfBOMBig {
		data, err = unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder().Bytes(data)
		if err != nil {
			return err
		}

	}
	// Convert from big endian if byte order mark is in big->little order.
	if len(data) > 2 && data[0] == utfBOMBig && data[1] == utfBOMLittle {
		data, err = unicode.UTF16(unicode.BigEndian, unicode.UseBOM).NewDecoder().Bytes(data)
		if err != nil {
			return err
		}
	}

	var newSchema Schema
	decoder := yaml.NewDecoder(bytes.NewBuffer(data))
	decoder.StrictMode(true)

	if err = decoder.Decode(&newSchema); err != nil {
		return err
	}

	for _, newDb := range newSchema.Databases {
		for _, tbl := range newDb.Tables {
			// Normalize any []interface{} into bson.A.
			tbl.Pipeline, err = bsonutil.NormalizeBSON(tbl.Pipeline, false)
			if err != nil {
				return err
			}
		}
		var existingDb *Database
		for _, db := range s.Databases {
			if db.Name == newDb.Name {
				existingDb = db
				break
			}
		}
		if existingDb == nil {
			// The entire DB is missing, copy the whole thing in.
			s.Databases = append(s.Databases, newDb)
		} else {
			// The schema being loaded refers to a DB that is already loaded
			// in the current schema. Need to merge tables.
			existingDb.Tables = append(existingDb.Tables, newDb.Tables...)
		}
	}

	return nil
}

// ToYAML marshals the schema to YAML.
func (s *Schema) ToYAML() ([]byte, error) {
	return yaml.Marshal(s)
}

// MarshalYAML marshals the table to YAML.
func (t *Table) MarshalYAML() (string, interface{}, error) {
	return "", struct {
		SQLName   string                   `yaml:"table"`
		MongoName string                   `yaml:"collection"`
		Pipeline  []map[string]interface{} `yaml:"pipeline"`
		Columns   []*Column                `yaml:"columns"`
	}{
		SQLName:   t.SQLName,
		MongoName: t.MongoName,
		Pipeline:  bsonutil.PipelineToMapSlice(t.Pipeline),
		Columns:   t.Columns,
	}, nil
}

type mongoStorableTable struct {
	SQLName   string    `bson:"sql_name"`
	MongoName string    `bson:"mongo_name"`
	Pipeline  string    `bson:"pipeline"`
	Columns   []*Column `bson:"columns"`
}

// GetBSON provides a custom struct that should be marshalled into a BSON
// document in place of this Table. Table deviates from the default BSON
// marshalling implementation by marshalling the `Pipeline` field as a JSON
// string instead of BSON arrays and documents. This is necessary in order to
// store the table in MongoDB, since $-prefixed keys are not allowed.
func (t *Table) GetBSON() (interface{}, error) {
	pipelineJSON, err := bsonutil.DocSliceToString(t.Pipeline)
	if err != nil {
		return nil, err
	}

	return &mongoStorableTable{
		SQLName:   t.SQLName,
		MongoName: t.MongoName,
		Pipeline:  pipelineJSON,
		Columns:   t.Columns,
	}, nil
}

// SetBSON unmarshals the provided oldbson.Raw into the Table.
func (t *Table) SetBSON(raw oldbson.Raw) error {
	tbl := &mongoStorableTable{}

	err := raw.Unmarshal(tbl)
	if err != nil {
		return err
	}

	pipeline := []bson.D{}
	err = bson.UnmarshalExtJSON([]byte(tbl.Pipeline), false, &pipeline)
	if err != nil {
		return err
	}

	t.SQLName = tbl.SQLName
	t.MongoName = tbl.MongoName
	t.Pipeline = pipeline
	t.Columns = tbl.Columns

	return nil
}

// MarshalBSON is a custom implementation for marshalling a Table into raw
// BSON bytes. Table deviates from the default BSON marshalling implementation
// by marshalling the `Pipeline` field as a JSON string instead of BSON arrays
// and documents. This is necessary in order to store the table in MongoDB, since
// $-prefixed keys are not allowed.
func (t *Table) MarshalBSON() ([]byte, error) {
	pipelineJSON, err := bsonutil.DocSliceToString(t.Pipeline)
	if err != nil {
		return nil, err
	}

	tbl := &mongoStorableTable{
		SQLName:   t.SQLName,
		MongoName: t.MongoName,
		Pipeline:  pipelineJSON,
		Columns:   t.Columns,
	}

	return bson.Marshal(tbl)
}

// UnmarshalBSON unmarshals the provided raw bytes into the Table.
func (t *Table) UnmarshalBSON(raw []byte) error {
	tbl := &mongoStorableTable{}

	err := bson.Unmarshal(raw, tbl)
	if err != nil {
		return err
	}

	pipeline := make([]bson.D, 0)
	err = bson.UnmarshalExtJSON([]byte(tbl.Pipeline), false, pipeline)
	if err != nil {
		return err
	}

	t.SQLName = tbl.SQLName
	t.MongoName = tbl.MongoName
	t.Pipeline = pipeline
	t.Columns = tbl.Columns

	return nil
}
