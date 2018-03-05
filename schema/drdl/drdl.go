package drdl

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	yaml "github.com/10gen/candiedyaml"
	"github.com/10gen/mongo-go-driver/bson"
)

// Schema represents a DRDL schema definition.
type Schema struct {
	Databases []*Database `yaml:"schema"`
}

// Database represents a DRDL database definition.
type Database struct {
	Name   string   `yaml:"db"`
	Tables []*Table `yaml:"tables"`
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
	MongoName string `yaml:"Name"`
	MongoType string `yaml:"MongoType"`
	SQLName   string `yaml:"SqlName"`
	SQLType   string `yaml:"SqlType"`
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
	var newSchema Schema
	if err := yaml.Unmarshal([]byte(data), &newSchema); err != nil {
		return err
	}

	for _, newDb := range newSchema.Databases {
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
