package schema

import (
	"fmt"
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
		Name       string                   `yaml:"table"`
		FQNS       string                   `yaml:"collection"`
		Pipeline   []map[string]interface{} `yaml:"pipeline"`
		RawColumns []*Column                `yaml:"columns"`
		Columns    map[string]*Column       `yaml:"-"`
		SQLColumns map[string]*Column       `yaml:"-"`
	}

	Database struct {
		Name      string            `yaml:"db"`
		RawTables []*Table          `yaml:"tables"`
		Tables    map[string]*Table `yaml:"-"`
	}

	Schema struct {
		Addr     string `yaml:"addr"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		LogLevel string `yaml:"log_level"`
		SSL      *SSL   `yaml:"ssl"`

		Url       string `yaml:"url"`
		SchemaDir string `yaml:"schema_dir"`

		RawDatabases []*Database          `yaml:"schema"`
		Databases    map[string]*Database `yaml:"-"`
	}

	SSL struct {
		// Don't require the certificate presented by the server to be valid
		AllowInvalidCerts bool `yaml:"allow_invalid_certs"`

		// Path to file containing cert and private key to present to the server
		PEMKeyFile string `yaml:"pem_key_file"`

		// File containing CA Certs - may be left blank if AllowInvalidCerts is true.
		CAFile string `yaml:"ca_file"`
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

func (t *Table) fixTypes() error {
	for _, c := range t.RawColumns {
		err := c.validateType()
		if err != nil {
			return err
		}
	}
	return nil
}

// Must is a helper that wraps a call to a function returning (*Schema, error)
// and panics if the error is non-nil. It is intended for use in variable
// initializations such as
//	var t = schema.Must(ParseSchemaData(raw))
func Must(c *Schema, err error) *Schema {
	if err != nil {
		panic(err)
	}
	return c
}

func (c *Schema) ingestSubFile(data []byte) error {
	temp, err := ParseSchemaData(data)
	if err != nil {
		return err
	}

	if len(temp.Addr) > 0 {
		return fmt.Errorf("cannot set addr in sub file")
	}

	if len(temp.User) > 0 {
		return fmt.Errorf("cannot set user in sub file")
	}

	if len(temp.Password) > 0 {
		return fmt.Errorf("cannot set password in sub file")
	}

	if len(temp.LogLevel) > 0 {
		return fmt.Errorf("cannot set log_level in sub file")
	}

	if len(temp.Url) > 0 {
		return fmt.Errorf("cannot set url in sub file")
	}

	if len(temp.SchemaDir) > 0 {
		return fmt.Errorf("cannot set schema_dir in sub file")
	}

	for name, schema := range temp.Databases {

		ours := c.Databases[name]

		if ours == nil {
			// entire db missing
			c.Databases[name] = schema
			c.RawDatabases = append(c.RawDatabases, schema)
			continue
		}

		// have to merge tables
		for table, tableSchema := range schema.Tables {
			if ours.Tables[table] != nil {
				return fmt.Errorf("table config conflict in db: %s table: %s", name, table)
			}

			ours.Tables[table] = tableSchema
			ours.RawTables = append(ours.RawTables, tableSchema)
		}
	}

	return nil
}

// CanCompare returns true if sqlValue can be converted to a
// value comparable to mType.
func CanCompare(leftType, rightType SQLType) bool {

	if leftType == SQLNull || rightType == SQLNull {
		return true
	}

	if leftType == SQLNone || rightType == SQLNone {
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
