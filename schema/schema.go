package schema

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"
)

type SQLType string

type MongoType string

type ColumnType struct {
	SQLType   SQLType
	MongoType MongoType
}

const (
	SQLNumeric    SQLType = "numeric"
	SQLInt                = "int"
	SQLInt64              = "int64"
	SQLFloat              = "float64"
	SQLVarchar            = "varchar"
	SQLTimestamp          = "timestamp"
	SQLDate               = "date"
	SQLBoolean            = "boolean"
	SQLArrNumeric         = "numeric[]"
	SQLNull               = "null"
	SQLNone               = ""
	SQLObjectID           = "objectid"
	SQLTuple              = "sqltuple"
)

const (
	MongoInt      MongoType = "int"
	MongoInt64              = "int64"
	MongoFloat              = "float64"
	MongoDecimal            = "decimal"
	MongoString             = "string"
	MongoGeo2D              = "geo.2darray"
	MongoObjectId           = "bson.ObjectId"
	MongoBool               = "bool"
	MongoDate               = "date"
	MongoNone               = ""
)

type (
	Column struct {
		Name      string    `yaml:"Name"`
		MongoType MongoType `yaml:"MongoType"`
		SqlName   string    `yaml:"SqlName"`
		SqlType   SQLType   `yaml:"SqlType"`
	}

	Table struct {
		Name           string                   `yaml:"table"`
		CollectionName string                   `yaml:"collection"`
		Pipeline       []map[string]interface{} `yaml:"pipeline"`
		RawColumns     []*Column                `yaml:"columns"`
		Columns        map[string]*Column       `yaml:"-"`
		SQLColumns     map[string]*Column       `yaml:"-"`
	}

	Database struct {
		Name      string            `yaml:"db"`
		RawTables []*Table          `yaml:"tables"`
		Tables    map[string]*Table `yaml:"-"`
	}

	Schema struct {
		RawDatabases []*Database          `yaml:"schema"`
		Databases    map[string]*Database `yaml:"-"`
	}
)

const (
	DateFormat      = "2006-01-02"
	TimeFormat      = "15:04:05"
	TimestampFormat = "2006-01-02 15:04:05"
)

var (
	DefaultLocale = time.UTC

	// TimestampCtorFormats holds the various formats
	// for constructing a SQL timestamp.
	TimestampCtorFormats = []string{
		"15:4:5",
		"2006-1-2",
		"2006-1-2 15:4:5",
		"2006-1-2 15:4:5.000",
	}
)

func (c *Column) validateType() error {
	switch MongoType(c.MongoType) {
	case MongoInt:
		switch SQLType(c.SqlType) {
		case SQLInt, SQLInt64, SQLNumeric, SQLVarchar:
		default:
			return fmt.Errorf("cannot map mongo type '%s' to SQL type '%s'", c.MongoType, c.SqlType)
		}
	case MongoInt64:
		switch SQLType(c.SqlType) {
		case SQLInt64, SQLNumeric, SQLVarchar:
		default:
			return fmt.Errorf("cannot map mongo type '%s' to SQL type '%s'", c.MongoType, c.SqlType)
		}
	case MongoFloat:
		switch SQLType(c.SqlType) {
		case SQLFloat, SQLNumeric, SQLVarchar:
		default:
			return fmt.Errorf("cannot map mongo type '%s' to SQL type '%s'", c.MongoType, c.SqlType)
		}
	case MongoDecimal:
		switch SQLType(c.SqlType) {
		case SQLNumeric, SQLVarchar:
		default:
			return fmt.Errorf("cannot map mongo type '%s' to SQL type '%s'", c.MongoType, c.SqlType)
		}
	case MongoString:
		switch SQLType(c.SqlType) {
		case SQLVarchar:
		default:
			return fmt.Errorf("cannot map mongo type '%s' to SQL type '%s'", c.MongoType, c.SqlType)
		}
	case MongoGeo2D:
		switch SQLType(c.SqlType) {
		case SQLArrNumeric:
		default:
			// TODO (INT-1015): remove once pipeline is supported
			// return fmt.Errorf("cannot map mongo type '%s' to SQL type '%s'", c.MongoType, c.SqlType)
		}
	case MongoObjectId:
		switch SQLType(c.SqlType) {
		case SQLVarchar:
		default:
			return fmt.Errorf("cannot map mongo type '%s' to SQL type '%s'", c.MongoType, c.SqlType)
		}
	case MongoBool:
		switch SQLType(c.SqlType) {
		case SQLBoolean:
		default:
			return fmt.Errorf("cannot map mongo type '%s' to SQL type '%s'", c.MongoType, c.SqlType)
		}
	case MongoDate:
		switch SQLType(c.SqlType) {
		case SQLDate, SQLTimestamp:
		default:
			return fmt.Errorf("cannot map mongo type '%s' to SQL type '%s'", c.MongoType, c.SqlType)
		}
	default:
		return fmt.Errorf("unsupported mongo type: '%s'", c.MongoType)
	}
	return nil
}

func (t *Table) validateColumnTypes() error {
	for _, c := range t.RawColumns {
		err := c.validateType()
		if err != nil {
			return err
		}
	}
	return nil
}

func New(data []byte) (*Schema, error) {
	s := &Schema{}
	err := s.Load(data)
	if err != nil {
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

// LoadFile loads schema settings from YML data in a byte slice.
func (s *Schema) Load(data []byte) error {
	if s.Databases == nil {
		s.Databases = make(map[string]*Database)
	}

	var theirs Schema
	if err := yaml.Unmarshal([]byte(data), &theirs); err != nil {
		return err
	}

	theirs.Databases = make(map[string]*Database)

	for _, db := range theirs.RawDatabases {
		if _, ok := theirs.Databases[db.Name]; ok {
			return fmt.Errorf("duplicate database in schema data: '%s'", db.Name)
		}

		if err := PopulateColumnMaps(db); err != nil {
			return err
		}

		theirs.Databases[db.Name] = db
	}

	for name, schema := range theirs.Databases {
		ours := s.Databases[name]

		if ours == nil {
			// The entire DB is missing, copy the whole thing in.
			s.Databases[name] = schema
			s.RawDatabases = append(s.RawDatabases, schema)
		} else {
			// The schema being loaded refers to a DB that is already loaded
			// in the current schema. Need to merge tables.
			for table, tableSchema := range schema.Tables {
				if ours.Tables[table] != nil {
					return fmt.Errorf("table config conflict in db: %s table: %s", name, table)
				}

				ours.Tables[table] = tableSchema
				ours.RawTables = append(ours.RawTables, tableSchema)
			}
		}
		if err := PopulateColumnMaps(s.Databases[name]); err != nil {
			return err
		}
	}
	return nil
}

// LoadFile loads schema settings from YML data in all files inside the given directory.
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

func PopulateColumnMaps(db *Database) error {
	db.Tables = make(map[string]*Table)

	for _, tbl := range db.RawTables {
		err := tbl.validateColumnTypes()
		if err != nil {
			return err
		}

		// TODO: consider lowercasing table names in config since we do
		// that in the query planner in constructing the TableScan node.
		if _, ok := db.Tables[tbl.Name]; ok {
			return fmt.Errorf("duplicate table [%s].", tbl.Name)
		}

		db.Tables[tbl.Name] = tbl

		tbl.Columns = make(map[string]*Column)
		tbl.SQLColumns = make(map[string]*Column)

		for _, c := range tbl.RawColumns {

			if c.SqlName == "" {
				c.SqlName = c.Name
			}

			if c.SqlName == "" {
				return fmt.Errorf("table [%s] has column with no name.", tbl.Name)
			}

			if _, ok := tbl.Columns[c.Name]; ok {
				return fmt.Errorf("duplicate column [%s].", c.Name)
			}

			tbl.Columns[c.Name] = c
			tbl.SQLColumns[c.SqlName] = c
		}
	}

	return nil
}

// isComparableType returns true if any of the types
// in sqlTypes can be compared with any other type.
func isComparableType(sqlTypes ...SQLType) bool {
	for _, sqlType := range sqlTypes {
		if sqlType == SQLNull || sqlType == SQLNone || sqlType == SQLTuple {
			return true
		}
	}
	return false
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

// CanCompare returns true if sqlValue can be converted to a
// value comparable to mType.
func CanCompare(leftType, rightType SQLType) bool {

	if isComparableType(leftType, rightType) {
		return true
	}

	switch leftType {

	case SQLArrNumeric, SQLFloat, SQLInt, SQLInt64, SQLNumeric:
		switch rightType {
		case SQLArrNumeric, SQLBoolean, SQLFloat, SQLInt, SQLInt64, SQLNumeric:
			return true
		}
	case SQLBoolean:
		switch rightType {
		case SQLArrNumeric, SQLBoolean, SQLFloat, SQLInt, SQLInt64, SQLNumeric:
			return true
		}
	case SQLDate, SQLTimestamp:
		switch rightType {
		case SQLDate, SQLTimestamp, SQLVarchar:
			return true
		}
	case SQLObjectID:
		switch rightType {
		case SQLObjectID, SQLVarchar:
			return true
		}
	case SQLVarchar:
		switch rightType {
		case SQLDate, SQLTimestamp, SQLVarchar:
			return true
		}
	}
	return false
}

// IsSimilar returns true if the logical or comparison
// operations can be carried on leftType against rightType
// with no need for type conversion in MongoDB.
func IsSimilar(leftType, rightType SQLType) bool {
	switch leftType {
	case SQLArrNumeric, SQLFloat, SQLInt, SQLInt64, SQLNumeric:
		switch rightType {
		case SQLArrNumeric, SQLFloat, SQLInt, SQLInt64, SQLNumeric:
			return true
		}
	}
	return false
}

type SQLTypes []SQLType

func (types SQLTypes) Len() int {
	return len(types)
}

func (types SQLTypes) Swap(i, j int) {
	types[i], types[j] = types[j], types[i]
}

func (types SQLTypes) Less(i, j int) bool {

	t1 := types[i]
	t2 := types[j]

	switch t1 {
	case SQLNone, SQLNull:
		return true
	case SQLVarchar:
		switch t2 {
		case SQLVarchar, SQLInt, SQLInt64, SQLFloat, SQLNumeric, SQLNone, SQLNull:
			return false
		case SQLTimestamp, SQLDate:
			return true
		default:
			return false
		}
	case SQLInt, SQLInt64, SQLFloat, SQLNumeric:
		switch t2 {
		case SQLInt, SQLInt64, SQLFloat, SQLNumeric, SQLNone, SQLNull:
			return false
		case SQLTimestamp, SQLDate, SQLVarchar:
			return true
		default:
			return false
		}
	case SQLTimestamp, SQLDate:
		return false
	}
	return false
}
