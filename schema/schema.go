package schema

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	yaml "github.com/10gen/candiedyaml"
	"github.com/mongodb/mongo-tools/common/bsonutil"
	"gopkg.in/mgo.v2/bson"
)

type SQLType string

type MongoType string

type ColumnType struct {
	SQLType   SQLType
	MongoType MongoType
}

const (
	MongoBool       MongoType = "bool"
	MongoDecimal128           = "bson.Decimal128"
	MongoDate                 = "date"
	MongoFilter               = "mongo.Filter"
	MongoFloat                = "float64"
	MongoGeo2D                = "geo.2darray"
	MongoInt                  = "int"
	MongoInt64                = "int64"
	MongoNone                 = ""
	MongoObjectId             = "bson.ObjectId"
	MongoString               = "string"
	MongoUUID                 = "bson.UUID"
	MongoUUIDOld              = "bson.UUID_Old"
	MongoUUIDJava             = "bson.UUID_Java_Legacy"
	MongoUUIDCSharp           = "bson.UUID_CSharp_Legacy"
)

const (
	SQLArrNumeric SQLType = "numeric[]"
	SQLBoolean            = "boolean"
	SQLDate               = "date"
	SQLDecimal128         = "decimal128"
	SQLFloat              = "float64"
	SQLInt                = "int"
	SQLInt64              = "int64"
	SQLNone               = ""
	SQLNull               = "null"
	SQLNumeric            = "numeric"
	SQLObjectID           = "objectid"
	SQLTimestamp          = "timestamp"
	SQLTuple              = "sqltuple"
	SQLUint64             = "sqluint64"
	SQLUUID               = "uuid"
	SQLVarchar            = "varchar"
)

const (
	zeroFloat  = float64(0)
	zeroInt    = int64(0)
	zeroString = ""
)

var (
	zeroDecimal128, _ = bson.ParseDecimal128("0")
	zeroBSON          = bson.ObjectId("")
	zeroTime          = time.Time{}
)

// ZeroValue returns the zero value for sqlType.
func (sqlType SQLType) ZeroValue() interface{} {
	switch sqlType {
	case SQLNumeric, SQLInt, SQLInt64:
		return zeroInt
	case SQLFloat, SQLArrNumeric:
		return zeroFloat
	case SQLVarchar:
		return zeroString
	case SQLTimestamp, SQLDate:
		return zeroTime
	case SQLBoolean:
		return false
	case SQLNone, SQLNull:
		return nil
	case SQLObjectID:
		return zeroBSON
	case SQLDecimal128:
		return zeroDecimal128
	}
	return ""
}

type (
	Column struct {
		Name      string    `yaml:"Name"`
		MongoType MongoType `yaml:"MongoType"`
		SqlName   string    `yaml:"SqlName"`
		SqlType   SQLType   `yaml:"SqlType"`
	}

	Table struct {
		Name           string             `yaml:"table"`
		CollectionName string             `yaml:"collection"`
		Pipeline       []bson.D           `yaml:"pipeline"`
		RawColumns     []*Column          `yaml:"columns"`
		Columns        map[string]*Column `yaml:"-"`
		SQLColumns     map[string]*Column `yaml:"-"`
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
		"2006-1-2 15",
		"2006-1-2 15:4",
		"2006-1-2 15:4:5",
		"2006-1-2 15:4:5.000",
	}
)

func (c *Column) validateType() error {
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
		case SQLFloat, SQLNumeric, SQLVarchar:
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
	case MongoObjectId, MongoString, MongoFilter, MongoUUID, MongoUUIDCSharp, MongoUUIDJava, MongoUUIDOld:
		if c.SqlType == SQLVarchar {
			err = nil
		}
	default:
		err = fmt.Errorf("unsupported mongo type: '%s'", c.MongoType)
	}

	return err
}

func (t *Table) validateColumnTypes() error {

	haveMongoFilter := false

	for _, c := range t.RawColumns {
		err := c.validateType()
		if err != nil {
			return err
		}

		if c.MongoType == MongoFilter {
			if haveMongoFilter {
				return fmt.Errorf("can not have more than one mongo filter in collection '%s'", t.CollectionName)
			}
			haveMongoFilter = true
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

// Load loads schema settings from YML data in a byte slice.
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
		dbName := strings.ToLower(db.Name)

		if _, ok := theirs.Databases[dbName]; ok {
			return fmt.Errorf("duplicate database in schema data: '%s'", db.Name)
		}

		theirs.Databases[dbName] = db

		ours := s.Databases[dbName]

		if ours == nil {
			// The entire DB is missing, copy the whole thing in.
			s.Databases[dbName] = db
			s.RawDatabases = append(s.RawDatabases, db)
		} else {
			// The schema being loaded refers to a DB that is already loaded
			// in the current schema. Need to merge tables.
			for _, table := range db.RawTables {
				tableName := strings.ToLower(table.Name)
				if ours.Tables[tableName] != nil {
					return fmt.Errorf("table config conflict in db: %s table: %s", db.Name, table.Name)
				}

				ours.Tables[tableName] = table
				ours.RawTables = append(ours.RawTables, table)
			}
		}

		if err := PopulateColumnMaps(s.Databases[dbName]); err != nil {
			return err
		}

		if err := HandlePipeline(s.Databases[dbName]); err != nil {
			return err
		}
	}
	return nil
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

func HandlePipeline(db *Database) error {

	for _, tbl := range db.RawTables {
		for i := 0; i < len(tbl.Pipeline); i++ {
			v, err := bsonutil.ConvertJSONValueToBSON(tbl.Pipeline[i])
			if err != nil {
				return fmt.Errorf("unable to parse extended json in table definition %q.%q: %v", db.Name, tbl.Name, err)
			}
			tbl.Pipeline[i] = v.(bson.D)
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

		tblName := strings.ToLower(tbl.Name)

		if _, ok := db.Tables[tblName]; ok {
			return fmt.Errorf("duplicate table [%s].", tbl.Name)
		}

		db.Tables[tblName] = tbl

		tbl.Columns = make(map[string]*Column)
		tbl.SQLColumns = make(map[string]*Column)

		var geo2DField []*Column
		var resolvedRawColumns []*Column

		for _, c := range tbl.RawColumns {

			if c.SqlName == "" {
				c.SqlName = c.Name
			}

			if c.SqlName == "" {
				return fmt.Errorf("table [%s] has column with no name.", tbl.Name)
			}

			if _, ok := tbl.SQLColumns[c.SqlName]; ok {
				return fmt.Errorf("duplicate SQL column [%s] in table [%s].", c.SqlName, tbl.Name)
			}

			// we're dealing with a legacy 2d array
			if c.SqlType == SQLArrNumeric {
				geo2DField = append(geo2DField, c)
			} else {
				tbl.Columns[c.Name] = c
				tbl.SQLColumns[c.SqlName] = c
				resolvedRawColumns = append(resolvedRawColumns, c)
			}
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
				tbl.Columns[c.Name] = c
				tbl.SQLColumns[c.SqlName] = c
				resolvedRawColumns = append(resolvedRawColumns, c)
			}
		}
		tbl.RawColumns = resolvedRawColumns
	}

	return nil
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

// IsSimilar returns true if the logical or comparison
// operations can be carried on leftType against rightType
// with no need for type conversion.
func IsSimilar(leftType, rightType SQLType) bool {
	switch leftType {
	case SQLArrNumeric, SQLFloat, SQLInt, SQLInt64, SQLNumeric, SQLUint64:
		switch rightType {
		case SQLArrNumeric, SQLFloat, SQLInt, SQLInt64, SQLNumeric, SQLUint64:
			return true
		}
	case SQLDate, SQLTimestamp:
		switch rightType {
		case SQLDate, SQLTimestamp:
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
		case SQLDecimal128, SQLInt, SQLInt64, SQLUint64, SQLFloat, SQLNumeric, SQLTimestamp, SQLDate, SQLBoolean:
			return true
		default:
			return false
		}
	case SQLBoolean:
		switch t2 {
		case SQLDecimal128, SQLInt, SQLInt64, SQLUint64, SQLFloat, SQLNumeric, SQLTimestamp, SQLDate:
			return true
		default:
			return false
		}
	case SQLInt, SQLInt64:
		switch t2 {
		case SQLDecimal128, SQLFloat, SQLNumeric, SQLUint64:
			return true
		default:
			return false
		}
	case SQLTimestamp:
		switch t2 {
		case SQLDecimal128, SQLInt, SQLInt64, SQLUint64, SQLFloat, SQLNumeric:
			return true
		default:
			return false
		}
	case SQLUint64:
		switch t2 {
		case SQLDecimal128, SQLFloat:
			return true
		default:
			return false
		}
	case SQLDate:
		switch t2 {
		case SQLDecimal128, SQLInt, SQLInt64, SQLUint64, SQLFloat, SQLNumeric, SQLTimestamp:
			return true
		default:
			return false
		}
	case SQLFloat, SQLNumeric:
		switch t2 {
		case SQLDecimal128:
			return true
		default:
			return false
		}
	}
	return false
}
