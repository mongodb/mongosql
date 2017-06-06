package schema

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	yaml "github.com/10gen/candiedyaml"
	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/util/bsonutil"
)

// Schema represents a configuration for a schema.
type Schema struct {
	// Databases are the databases in the schema.
	Databases []*Database `yaml:"schema"`
}

// New creates a new schema.
func New(data []byte) (*Schema, error) {
	s := &Schema{}

	if err := s.Load(data); err != nil {
		return nil, err
	}
	return s, nil
}

// LoadFile loads schema settings from a YML file.
func (s *Schema) LoadFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return s.Load(data)
}

// Load loads schema settings from YML data in a byte slice.
func (s *Schema) Load(data []byte) error {
	var theirs Schema
	if err := yaml.Unmarshal([]byte(data), &theirs); err != nil {
		return err
	}

	for _, theirDb := range theirs.Databases {

		if ourDb, ok := s.Database(theirDb.Name); !ok {
			// The entire DB is missing, copy the whole thing in.
			s.Databases = append(s.Databases, theirDb)
		} else {
			// The schema being loaded refers to a DB that is already loaded
			// in the current schema. Need to merge tables.
			ourDb.Tables = append(ourDb.Tables, theirDb.Tables...)
		}
	}

	return s.Validate()
}

// LoadDir loads schema settings from YML data in all files inside the given directory.
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

// Database gets the database with the given name.
func (s *Schema) Database(name string) (*Database, bool) {
	name = strings.ToLower(name)
	for _, d := range s.Databases {
		if strings.ToLower(d.Name) == name {
			return d, true
		}
	}

	return nil, false
}

// Validate validates a schema.
func (s *Schema) Validate() error {
	for _, d := range s.Databases {
		if err := d.validate(); err != nil {
			return fmt.Errorf("failed to validate database '%s': %v", d.Name, err)
		}
	}
	return nil
}

// Database represents a configuration for a database.
type Database struct {
	// Name is the name of the database.
	Name string `yaml:"db"`
	// Tables are the tables in the database.
	Tables []*Table `yaml:"tables"`
}

// Table gets the table with the given name.
func (d *Database) Table(name string) (*Table, bool) {
	name = strings.ToLower(name)
	for _, t := range d.Tables {
		if strings.ToLower(t.Name) == name {
			return t, true
		}
	}

	return nil, false
}

func (d *Database) validate() error {

	tmap := make(map[string]struct{})

	for _, t := range d.Tables {
		err := t.validate()
		if err != nil {
			return fmt.Errorf("failed to validate table '%s': %v", t.Name, err)
		}

		if _, ok := tmap[strings.ToLower(t.Name)]; ok {
			return fmt.Errorf("duplicate table '%s'", t.Name)
		}

		tmap[t.Name] = struct{}{}
	}

	return nil
}

// Table represents a configuration for a table.
type Table struct {
	// Name is the name of a table.
	Name string `yaml:"table"`
	// CollectionName is the collection name the table maps to in MongoDB.
	CollectionName string `yaml:"collection"`
	// Pipeline are pre-processing directives for how to derive the table from the MongoDB collection.
	Pipeline []bson.D `yaml:"pipeline"`
	// Columns are the columns in the table.
	Columns []*Column `yaml:"columns"`
}

// Column gets the column given the SqlName.
func (t *Table) Column(sqlName string) (*Column, bool) {
	sqlName = strings.ToLower(sqlName)
	for _, c := range t.Columns {
		if strings.ToLower(c.Name) == sqlName {
			return c, true
		}
	}

	return nil, false
}

func (t *Table) validate() error {

	for i := 0; i < len(t.Pipeline); i++ {
		v, err := bsonutil.ConvertJSONValueToBSON(t.Pipeline[i])
		if err != nil {
			return fmt.Errorf("unable to parse extended json: %v", err)
		}
		t.Pipeline[i] = v.(bson.D)
	}

	haveMongoFilter := false

	var geo2DField []*Column
	var resolvedRawColumns []*Column

	cmap := make(map[string]struct{})

	for _, c := range t.Columns {
		err := c.validate()
		if err != nil {
			return fmt.Errorf("failed to validate column '%s': %v", c.Name, err)
		}

		if c.MongoType == MongoFilter {
			if haveMongoFilter {
				return fmt.Errorf("cannot have more than one mongo filter")
			}
			haveMongoFilter = true
		}

		if _, ok := cmap[strings.ToLower(c.SqlName)]; ok {
			return fmt.Errorf("duplicate SQL column '%s'", c.SqlName)
		}

		// we're dealing with a legacy 2d array
		if c.MongoType == MongoGeo2D {
			geo2DField = append(geo2DField, c)
		} else {
			resolvedRawColumns = append(resolvedRawColumns, c)
		}

		cmap[strings.ToLower(c.SqlName)] = struct{}{}
	}

	for _, column := range geo2DField {
		// add longitude and latitude SqlName
		for j, suffix := range []string{"_longitude", "_latitude"} {
			c := &Column{
				Name:      fmt.Sprintf("%v.%v", column.Name, j),
				SqlName:   column.SqlName + suffix,
				SqlType:   SQLArrNumeric,
				MongoType: SQLFloat,
			}
			resolvedRawColumns = append(resolvedRawColumns, c)
		}
	}
	t.Columns = resolvedRawColumns
	return nil
}

// Column represents a configuration for a column.
type Column struct {
	// Name is the name of the field in MongoDB.
	Name string `yaml:"Name"`
	// MongoType is the type of the field in MongoDB.
	MongoType MongoType `yaml:"MongoType"`
	// SqlName is the name of the column to be shown to users.
	SqlName string `yaml:"SqlName"`
	// SqlType is the type to be shown to users.
	SqlType SQLType `yaml:"SqlType"`
}

func (c *Column) validate() error {
	if c.SqlName == "" {
		c.SqlName = c.Name
	}

	if c.SqlName == "" {
		return fmt.Errorf("found column with no name")
	}

	err := fmt.Errorf("cannot map mongo type '%s' to SQL type '%s'", c.MongoType, c.SqlType)
	switch c.MongoType {
	case MongoBool:
		if c.SqlType == SQLBoolean {
			err = nil
		}
	case MongoDate:
		switch c.SqlType {
		case SQLDate, SQLTimestamp:
			err = nil
		}
	case MongoDecimal128:
		switch c.SqlType {
		case SQLDecimal128, SQLNumeric, SQLVarchar:
			err = nil
		}
	case MongoFloat:
		switch c.SqlType {
		case SQLFloat, SQLNumeric, SQLVarchar, SQLArrNumeric:
			err = nil
		}
	case MongoGeo2D:
		if c.SqlType == SQLArrNumeric {
			err = nil
		}
	case MongoInt:
		switch c.SqlType {
		case SQLInt, SQLInt64, SQLNumeric, SQLVarchar:
			err = nil
		}
	case MongoInt64:
		switch c.SqlType {
		case SQLInt64, SQLNumeric, SQLVarchar:
			err = nil
		}
	case MongoNumber:
		switch c.SqlType {
		case SQLInt, SQLInt64, SQLFloat, SQLDecimal128, SQLNumeric:
			err = nil
		}
	case MongoObjectId, MongoString, MongoFilter, MongoUUID, MongoUUIDCSharp, MongoUUIDJava, MongoUUIDOld:
		if c.SqlType == SQLVarchar {
			err = nil
		}
	default:
		err = fmt.Errorf("unsupported mongo type: '%s'", c.MongoType)
	}

	return err
}

// Must is a helper that wraps a call to a function returning (*Schema, error)
// and panics if the error is non-nil. It is intended for use in variable
// initializations such as
//	var t = schema.Must(New(raw))
func Must(c *Schema, err error) *Schema {
	if err != nil {
		panic(err)
	}
	return c
}
