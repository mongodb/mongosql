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

const (
	mongoPrimaryKey = "_id"
)

// Schema represents a configuration for a schema.
type Schema struct {
	// Alterations is a slice of alterations to be applied to
	// this relational schema
	Alterations []*Alteration
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
	// parent is a pointer to an array table's parent table.
	// This field is only used during mongo-to-relational schema translation.
	parent *Table
	// primaryKey is a slice of all the columns that comprise the primary key.
	// This field is only used during mongo-to-relational schema translation.
	primaryKey []*Column
	// isPostProcessed indicates if this table has been processed after schema
	// mapping.
	isPostProcessed bool
}

// New creates a new schema.
func New(data []byte, lg *log.Logger) (*Schema, error) {
	s := &Schema{}

	if err := s.Load(data, lg); err != nil {
		return nil, err
	}
	return s, nil
}

// Altered returns a new Schema that is equivalent to the current schema with
// its alterations applied. The returned schema will have an empty Alterations
// slice.
func (s *Schema) Altered() (*Schema, error) {
	if len(s.Alterations) == 0 {
		return s, nil
	}

	newSchema := s.DeepCopy()
	for _, a := range s.Alterations {
		err := a.alter(newSchema)
		if err != nil {
			return nil, fmt.Errorf("could not alter schema: %v", err)
		}
	}
	newSchema.Alterations = nil
	return newSchema, nil
}

// Database gets the database with the given name.
func (s *Schema) Database(name string) (*Database, bool) {
	for _, d := range s.Databases {
		if strings.EqualFold(d.Name, name) {
			return d, true
		}
	}

	return nil, false
}

// DeepCopy returns a deep copy of a Schema
func (s *Schema) DeepCopy() *Schema {
	if s == nil {
		return nil
	}

	dbs := []*Database{}
	for _, db := range s.Databases {
		dbs = append(dbs, db.deepCopy())
	}

	alts := []*Alteration{}
	for _, alt := range s.Alterations {
		alts = append(alts, alt)
	}

	return &Schema{
		alts,
		dbs,
	}
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

// Load loads schema settings from YML data in a byte slice.
func (s *Schema) Load(data []byte, lgr *log.Logger) error {
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

	newDatabases := []*Database{}
	for _, db := range s.Databases {
		newTables := []*Table{}
		for _, tb := range db.Tables {
			err := tb.PostProcess(false, lgr)
			if err != nil {
				return err
			}
			newTables = append(newTables, tb.deepCopy())
		}
		db.Tables = newTables

		newDatabases = append(newDatabases, db.deepCopy())
	}

	s.Databases = newDatabases

	return s.Validate()

}

// LoadDir loads schema settings from YML data in all files inside the given directory.
func (s *Schema) LoadDir(root string, lgr *log.Logger) error {
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
		err = s.LoadFile(fullPath, lgr)
		if err != nil {
			return fmt.Errorf("in schema file %v: %v", fullPath, err)
		}
	}
	return nil
}

// LoadFile loads schema settings from a YML file.
func (s *Schema) LoadFile(filename string, lgr *log.Logger) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return s.Load(data, lgr)
}

// Validate validates a schema.
func (s *Schema) Validate() error {
	for _, d := range s.Databases {
		err := d.validate()
		if err != nil {
			return fmt.Errorf("failed to validate database '%s': %v", d.Name, err)
		}
	}
	return nil
}

// AddTable adds the given table to a database, returning an error if the
// database already has a table with the same name.
func (d *Database) AddTable(t *Table, lgr *log.Logger) error {
	if _, ok := d.Table(t.Name); ok {
		initName := t.Name
		t.Name = d.uniqueTableName(t.Name)
		if t.Name != initName {
			lgr.Warnf(log.Dev, "found 2 namespaces with the same case-insensitive "+
				"name: renamed %q to %q", initName, t.Name)
		}
	}

	d.Tables = append(d.Tables, t)
	return nil
}

func (d *Database) contains(tableName string) bool {
	for _, t := range d.Tables {
		if strings.EqualFold(t.Name, tableName) {
			return true
		}
	}
	return false
}

func (d *Database) deepCopy() *Database {
	tables := []*Table{}
	for _, t := range d.Tables {
		tables = append(tables, t.deepCopy())
	}
	return newDatabase(d.Name, tables)
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
func (d *Database) Map(js *mongo.Schema, name string, shouldPreJoin bool,
	uuidSubtype3Encoding string, lg *log.Logger) error {

	// create the table into which we will map this collection's fields.
	// this table has the same name as the collection it is mapped from.
	// unless we have array fields, this is the only table we will create.
	t := NewTable(name, name, nil, nil, nil, nil, false)

	// initialize the top-level mapping context
	ctx := &mappingContext{
		logger:               lg,
		db:                   d,
		table:                t,
		uuidSubtype3Encoding: uuidSubtype3Encoding,
	}

	// map the collection schema to a relational schema
	err := ctx.mapObjectSchema(js)
	if err != nil {
		return err
	}

	// add this table to the database
	err = d.AddTable(t, lg)
	if err != nil {
		return err
	}

	tables := []*Table{}
	for _, table := range d.Tables {
		err = table.PostProcess(shouldPreJoin, lg)
		if err != nil {
			return err
		}

		// don't add tables without any columns
		if len(table.Columns) == 0 {
			lg.Debugf(log.Dev, "omitting table %q: has no columns", t.Name)
			continue
		}

		tables = append(tables, table)
	}

	d.Tables = tables

	// validate the db schema
	err = d.validate()
	if err != nil {
		return err
	}

	d.Sort()

	lg.Debugf(log.Dev, "mapped new table %q", name)
	return nil
}

// Sort sorts a database's tables lexicographically.
func (d *Database) Sort() {
	sort.Slice(d.Tables, func(i, j int) bool {
		return d.Tables[i].Name < d.Tables[j].Name
	})
}

// Table gets the table with the given name.
func (d *Database) Table(name string) (*Table, bool) {
	for _, t := range d.Tables {
		if strings.EqualFold(t.Name, name) {
			return t, true
		}
	}

	return nil, false
}

func (d *Database) uniqueTableName(tableName string) string {
	retTableName := tableName
	i := 0
	for {
		if d.contains(retTableName) {
			retTableName = fmt.Sprintf("%v_%v", tableName, i)
			i++
			continue
		}
		return retTableName
	}
}

func (d *Database) validate() error {
	tmap := make(map[string]struct{})
	for _, t := range d.Tables {
		err := t.validate()
		if err != nil {
			return fmt.Errorf("failed to validate table '%s': %v", t.Name, err)
		}

		key := strings.ToLower(t.Name)
		if _, ok := tmap[key]; ok {
			return fmt.Errorf("duplicated name for table '%s'", t.Name)
		}
		tmap[key] = struct{}{}
	}

	return nil
}

// AddColumn adds the given column to a table, returning an error if the table
// already has a column of the same name.
func (t *Table) AddColumn(c *Column, lgr *log.Logger) error {
	if _, ok := t.Column(c.SQLName); ok {
		initName := c.SQLName
		c.SQLName = t.uniqueColumnName(c.SQLName)
		if c.SQLName != initName {
			lgr.Warnf(log.Admin, "found 2 columns with the same case-insensitive "+
				"name: renamed %q to %q", initName, c.SQLName)
		}
	}
	t.Columns = append(t.Columns, c)
	return nil
}

// Column gets the column with the given mongoName.
func (t *Table) Column(mongoName string) (*Column, bool) {
	for _, c := range t.Columns {
		if strings.EqualFold(c.Name, mongoName) {
			return c, true
		}
	}

	return nil, false
}

func (t *Table) contains(columnName string) bool {
	for _, t := range t.Columns {
		if strings.EqualFold(t.SQLName, columnName) {
			return true
		}
	}
	return false
}

// copyParent modifies a table to include columns (and pipeline stages) from its
// parent table. Note that the only tables with parents are array tables;
// passing a non-array table to this function will have no effect.
//
// If shouldPreJoin is false, the table will be modified to include the primary-key
// columns from the parent (which will allow the user to join the array table with
// its parent). If shouldPreJoin is true, the table will include _all_ columns
// from the parent, effectively creating a "pre-joined" table.
//
// This function assumes that copyParent has already been called on all of the
// table's ancestors.
func (t *Table) copyParent(shouldPreJoin bool, lg *log.Logger) error {
	if t.parent == nil {
		return nil
	}

	// prepend the parent's primary key columns to this table's primary key columns
	// a column is a primary key column if it was mapped from the top-level _id
	// field (or, if _id is a document, one of _id's fields)
	for _, parentPK := range t.parent.primaryKey {
		t.primaryKey = append(t.primaryKey, parentPK)
	}

	// determine which columns should be pre-joined from the parent
	source := t.parent.primaryKey
	if shouldPreJoin {
		source = t.parent.Columns
	}

	// include the chosen columns from the parent
	for _, c := range source {
		err := t.AddColumn(c, lg)
		if err != nil {
			return err
		}
	}

	// prepend the parent's pipeline to this table's pipeline
	parentPipeline := make([]bson.D, len(t.parent.Pipeline), len(t.parent.Pipeline)+len(t.Pipeline))
	copy(parentPipeline, t.parent.Pipeline)
	t.Pipeline = append(parentPipeline, t.Pipeline...)
	return nil
}

func (t *Table) deepCopy() *Table {
	cols := []*Column{}
	for _, c := range t.Columns {
		cols = append(cols, c.deepCopy())
	}

	pkCols := []*Column{}
	for _, c := range t.primaryKey {
		pkCols = append(pkCols, c.deepCopy())
	}

	var parent *Table
	if t.parent != nil {
		parent = t.parent.deepCopy()
	}

	pipeline := make([]bson.D, len(t.Pipeline))
	copy(pipeline, t.Pipeline)

	return NewTable(t.Name, t.CollectionName, pipeline, cols, parent, pkCols, t.isPostProcessed)
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

// IsPrimaryKey returns true if mongoName is part of a primary
// key and false otherwise.
func (t *Table) IsPrimaryKey(mongoName string) bool {
	if mongoName == "_id" {
		return true
	}

	for _, d := range t.Pipeline {
		unwindVal, ok := d.Map()["$unwind"]
		if !ok {
			return false
		}

		unwind, ok := unwindVal.(bson.D)
		if !ok {
			return false
		}

		arrayIndexNameVal, ok := unwind.Map()["includeArrayIndex"]
		if !ok {
			continue
		}

		arrayIndexName, ok := arrayIndexNameVal.(string)
		if !ok {
			continue
		}

		if mongoName == arrayIndexName {
			return true
		}
	}

	return false
}

// PostProcess goes through the fully mapped relational schema
// for this table and does the following:
// - Fully resolves any column names
// - Pre-joins array tables, if necessary, from the parent table
// - Sorts the columns alphabetically.
func (t *Table) PostProcess(shouldPreJoin bool, lg *log.Logger) error {
	if t.isPostProcessed {
		return nil
	}

	err := t.resolveColumns(lg)
	if err != nil {
		return err
	}

	err = t.copyParent(shouldPreJoin, lg)
	if err != nil {
		return err
	}

	t.Sort()

	t.isPostProcessed = true
	return nil
}

// resolveColumns examines all the tables that have been mapped from the
// JSON schema and does the following:
// - Removes any columns that only contain whitespace characters
// - Adds a coordinate suffix to geo-2D index columns
func (t *Table) resolveColumns(lg *log.Logger) error {
	var geo2DField []*Column
	var resolvedRawColumns []*Column

	for _, c := range t.Columns {
		// don't map columns with whitespace keys
		if strings.Trim(c.Name, " ") == "" {
			lg.Debugf(log.Admin, "omitting column with whitespace key %q", c.SQLName)
			continue
		}

		if c.SQLName == "" {
			c.SQLName = c.Name
		}

		// we're dealing with a legacy 2d array
		if c.MongoType == MongoGeo2D {
			lg.Debugf(log.Admin, "adding coordinate suffix to geo2d column %q", c.SQLName)
			geo2DField = append(geo2DField, c)
		} else {
			resolvedRawColumns = append(resolvedRawColumns, c)
		}
	}

	if len(resolvedRawColumns) == len(t.Columns) {
		return nil
	}

	t.Columns = []*Column{}

	for _, c := range resolvedRawColumns {
		if err := t.AddColumn(c, lg); err != nil {
			return err
		}
	}

	for _, c := range geo2DField {
		// add longitude and latitude SQLName
		for j, suffix := range []string{"_longitude", "_latitude"} {
			initSQLName := c.SQLName + suffix
			newC := &Column{
				Name:      fmt.Sprintf("%v.%v", c.Name, j),
				SQLName:   initSQLName,
				SQLType:   SQLArrNumeric,
				MongoType: MongoFloat,
			}

			if err := t.AddColumn(newC, lg); err != nil {
				return err
			}
		}
	}

	return nil
}

// Sort sorts a table's columns lexicographically.
func (t *Table) Sort() {
	sort.Slice(t.Columns, func(i, j int) bool {
		// sort _id before other columns
		if t.Columns[i].Name == mongoPrimaryKey {
			return true
		}
		if t.Columns[j].Name == mongoPrimaryKey {
			return false
		}
		return t.Columns[i].Name < t.Columns[j].Name
	})
}

func (t *Table) uniqueColumnName(columnName string) string {
	retColumnName := columnName
	i := 0
	for {
		if t.contains(retColumnName) {
			retColumnName = fmt.Sprintf("%v_%v", columnName, i)
			i++
			continue
		}
		return retColumnName
	}
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

		key := strings.ToLower(c.SQLName)
		if _, ok := cmap[key]; ok {
			return fmt.Errorf("duplicate SQL column: '%s'", c.SQLName)
		}
		cmap[key] = struct{}{}
	}

	return nil
}

// Column represents a configuration for a column.
type Column struct {
	// Name is the name of the field in MongoDB.
	Name string `yaml:"Name"`
	// MongoType is the type of the field in MongoDB.
	MongoType MongoType `yaml:"MongoType"`
	// SQLName is the name of the column to be shown to users.
	SQLName string `yaml:"SqlName"`
	// SQLType is the type to be shown to users.
	SQLType SQLType `yaml:"SqlType"`
}

func (c *Column) deepCopy() *Column {
	return &Column{c.Name, c.MongoType, c.SQLName, c.SQLType}
}

// Equals checks whether a Column is equal to the provided Column.
func (c *Column) Equals(other *Column) error {
	if c.Name != other.Name {
		return fmt.Errorf("column names not equal.\nthis: '%s'\nother: '%s'", c.Name, other.Name)
	}
	if c.MongoType != other.MongoType {
		return fmt.Errorf("mongotypes not equal.\nthis: '%s'\nother: '%s'", c.MongoType, other.MongoType)
	}
	if c.SQLName != other.SQLName {
		return fmt.Errorf("sqlnames not equal.\nthis: '%s'\n other: '%s'", c.SQLName, other.SQLName)
	}
	if c.SQLType != other.SQLType {
		return fmt.Errorf("sqltypes not equal.\nthis: '%s'\nother: '%s'", c.SQLType, other.SQLType)
	}
	return nil
}

func (c *Column) validate() error {
	if c.SQLName == "" {
		return fmt.Errorf("found empty name for SQL column: %#v", c)
	}

	err := fmt.Errorf("cannot map mongo type '%s' to SQL type '%s'", c.MongoType, c.SQLType)
	switch c.MongoType {
	case MongoBool:
		if c.SQLType == SQLBoolean {
			err = nil
		}
	case MongoDate:
		switch c.SQLType {
		case SQLDate, SQLTimestamp:
			err = nil
		}
	case MongoDecimal128:
		switch c.SQLType {
		case SQLDecimal128, SQLNumeric, SQLVarchar:
			err = nil
		}
	case MongoFloat:
		switch c.SQLType {
		case SQLFloat, SQLNumeric, SQLVarchar, SQLArrNumeric:
			err = nil
		}
	case MongoGeo2D:
		if c.SQLType == SQLArrNumeric {
			err = nil
		}
	case MongoInt:
		switch c.SQLType {
		case SQLInt, SQLInt64, SQLNumeric, SQLVarchar:
			err = nil
		}
	case MongoInt64:
		switch c.SQLType {
		case SQLInt64, SQLNumeric, SQLVarchar:
			err = nil
		}
	case MongoNumber:
		switch c.SQLType {
		case SQLInt, SQLInt64, SQLFloat, SQLDecimal128, SQLNumeric:
			err = nil
		}
	case MongoObjectID, MongoString, MongoFilter, MongoUUID, MongoUUIDCSharp, MongoUUIDJava, MongoUUIDOld:
		if c.SQLType == SQLVarchar {
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
