package schema

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	yaml "github.com/10gen/candiedyaml"
	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/util/bsonutil"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema/mongo"
)

// Schema represents a configuration for a schema.
type Schema struct {
	// Databases are the databases in the schema.
	Databases []*Database `yaml:"schema"`
}

// Database represents a configuration for a database.
type Database struct {
	// Name is the name of the database.
	Name string `yaml:"db"`
	// Tables are the tables in the database.
	Tables []*Table `yaml:"tables"`
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
	// parent is a pointer to an array table's parent table
	// this field is only used during mongo-to-relational schema translation
	parent *Table
	// primaryKey is a slice of all the columns that comprise the primary key
	// this field is only used during mongo-to-relational schema translation
	primaryKey []*Column
}

// New creates a new schema.
func New(data []byte) (*Schema, error) {
	s := &Schema{}

	if err := s.Load(data); err != nil {
		return nil, err
	}
	return s, nil
}

// Equals checks whether a Schema is equal to the provided Schema.
func (s *Schema) Equals(other *Schema) error {
	for i, db := range s.Databases {
		otherDb := other.Databases[i]
		err := db.Equals(otherDb)
		if err != nil {
			return fmt.Errorf("Databases not equal at index %d: %v", i, err)
		}
	}
	return nil
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

// AddTable adds the given table to a database, returning an error if the
// database already has a table with the same name.
func (d *Database) AddTable(t *Table) error {
	if _, ok := d.Table(t.Name); ok {
		return fmt.Errorf("Database '%s' already has a table '%s'", d.Name, t.Name)
	}
	d.Tables = append(d.Tables, t)
	return nil
}

// Equals checks whether a Database is equal to the provided Database.
func (d *Database) Equals(other *Database) error {
	if d.Name != other.Name {
		return fmt.Errorf("db names not equal.\nthis: '%s'\nother: '%s'", d.Name, other.Name)
	}

	for i, table := range d.Tables {
		otherTable := other.Tables[i]
		err := table.Equals(otherTable)
		if err != nil {
			return fmt.Errorf("Tables not equal at index %d: %v", i, err)
		}
	}

	return nil
}

// Map takes a mongo schema that describes a collection with the provided name
// and creates a set of tables in the Database that comprise a relational
// equivalent of that schema. If preJoined is true, the tables generated for
// array fields will include parent fields, effectively resulting in pre-joined
// tables.
func (d *Database) Map(js *mongo.Schema, name string, preJoined bool, lg log.Logger) error {

	// create the table into which we will map this collection's fields.
	// this table has the same name as the collection it is mapped from.
	// unless we have array fields, this is the only table we will create.
	t := newTable(name, name)
	err := d.AddTable(t)
	if err != nil {
		return err
	}
	lg.Logf(log.Info, "Mapped new table %q", name)

	// initialize the top-level mapping context with the logger, db, and table
	ctx := &mappingContext{
		logger: lg,
		db:     d,
		table:  t,
	}

	// map the collection schema to a relational schema
	err = ctx.mapObjectSchema(js)
	if err != nil {
		return err
	}

	// pre-join array tables, remove columnless tables, and sort columns
	var tables []*Table
	for _, t := range d.Tables {
		t.copyParent(!preJoined)
		t.Sort()
		if len(t.Columns) > 0 {
			tables = append(tables, t)
		} else {
			lg.Logf(
				log.Info,
				"Omitting table %q: has no columns",
				t.Name,
			)
		}
	}
	d.Tables = tables

	// validate the db schema (also performs some transformations)
	err = d.validate()
	if err != nil {
		return err
	}

	return nil
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

// AddColumn adds the given column to a table, returning an error if the table
// already has a column of the same name.
func (t *Table) AddColumn(c *Column) error {
	if _, ok := t.Column(c.Name); ok {
		return fmt.Errorf("Table '%s' already has a column '%s'", t.Name, c.Name)
	}
	t.Columns = append(t.Columns, c)
	return nil
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

// copyParent modifies a table to include columns (and pipeline stages) from its
// parent table. Note that the only tables with parents are array tables;
// passing a non-array table to this function will have no effect.
//
// If primaryKeyOnly is true, the table will be modified to include the primary-key
// columns from the parent (which will allow the user to join the array table with
// its parent). If primaryKeyOnly is false, the table will include _all_ columns
// from the parent, effectively creating a "pre-joined" table.
//
// This function assumes that copyParent has already been called on all of the
// table's ancestors.
func (t *Table) copyParent(primaryKeyOnly bool) error {
	if t.parent == nil {
		return nil
	}

	// prepend the parent's primary key columns to this table's primary key columns
	// a column is a primary key column if it was mapped from the top-level _id
	// field (or, if _id is a document, one of _id's fields)
	parentPrimaryKey := make([]*Column, len(t.parent.primaryKey))
	copy(parentPrimaryKey, t.parent.primaryKey)
	t.primaryKey = append(parentPrimaryKey, t.primaryKey...)
	parentPrimaryKey = nil

	// determine which columns should be pre-joined from the parent
	source := t.parent.Columns
	if primaryKeyOnly {
		source = t.parent.primaryKey
	}

	// include the chosen columns from the parent
	for _, copy := range source {
		err := t.AddColumn(copy)
		if err != nil {
			return err
		}
	}

	// prepend the parent's pipeline to this table's pipeline
	parentPipeline := make([]bson.D, len(t.parent.Pipeline), len(t.parent.Pipeline)+len(t.Pipeline))
	copy(parentPipeline, t.parent.Pipeline)
	t.Pipeline = append(parentPipeline, t.Pipeline...)
	parentPipeline = nil
	return nil
}

// Equals checks whether a Table is equal to the provided Table.
func (t *Table) Equals(other *Table) error {
	if t.Name != other.Name {
		return fmt.Errorf("table names not equal.\nthis: '%s'\nother: '%s'", t.Name, other.Name)
	}

	if t.CollectionName != other.CollectionName {
		return fmt.Errorf("table collection names not equal.\nthis: '%s'\nother: '%s'", t.CollectionName, other.CollectionName)
	}

	if len(t.Pipeline) != len(other.Pipeline) {
		return fmt.Errorf("pipeline lengths not equal.\nthis: %d\nother: %d", len(t.Pipeline), len(other.Pipeline))
	}

	if len(t.Pipeline) > 0 && !reflect.DeepEqual(t.Pipeline, other.Pipeline) {
		return fmt.Errorf("table's pipelines not equal.\nthis: %+v\nother: %+v", t.Pipeline, other.Pipeline)
	}

	if len(t.Columns) != len(other.Columns) {
		return fmt.Errorf("column counts not equal.\nthis: %d\nother: %d", len(t.Columns), len(other.Columns))
	}

	for i, column := range t.Columns {
		otherColumn := other.Columns[i]
		err := column.Equals(otherColumn)
		if err != nil {
			return fmt.Errorf("columns not equal at index %d: %v", i, err)
		}
	}

	return nil
}

// Sort sorts a table's columns lexicographically.
func (t *Table) Sort() {
	sort.Slice(t.Columns, func(i, j int) bool {
		return t.Columns[i].Name < t.Columns[j].Name
	})
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

// Equals checks whether a Column is equal to the provided Column.
func (c *Column) Equals(other *Column) error {
	if c.Name != other.Name {
		return fmt.Errorf("column names not equal.\nthis: '%s'\nother: '%s'", c.Name, other.Name)
	}
	if c.MongoType != other.MongoType {
		return fmt.Errorf("mongotypes not equal.\nthis: '%s'\nother: '%s'", c.MongoType, other.MongoType)
	}
	if c.SqlName != other.SqlName {
		return fmt.Errorf("sqlnames not equal.\nthis: '%s'\n other: '%s'", c.SqlName, other.SqlName)
	}
	if c.SqlType != other.SqlType {
		return fmt.Errorf("sqltypes not equal.\nthis: '%s'\nother: '%s'", c.SqlType, other.SqlType)
	}
	return nil
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
