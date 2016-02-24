package schema

import (
	"fmt"
	"time"
)

const (
	SQLString  string = "string"
	SQLInt            = "int"
	SQLFloat          = "float"
	SQLDouble         = "double"
	SQLBoolean        = "boolean"
	SQLVarchar        = "varchar"

	SQLDate      = "date"
	SQLDateTime  = "datetime"
	SQLTime      = "time"
	SQLTimestamp = "timestamp"
	SQLYear      = "year"

	SQLNull = "null"

	/*

		TODO (INT-800): support these

		SQLDecimal  = "decimal"
		SQLDouble   = "double"
		SQLEnum     = "enum"
		SQLGeometry = "geometry"

		SQLBigInt   = "bigint"
		SQLMedInt   = "mediumint"
		SQLSmallInt = "smallint"
		SQLTiny     = "tinyint"

		SQLLongText   = "longtext"
		SQLTinyText   = "tinytext"
		SQLMediumText = "mediumtext"

		SQLSet  = "set"
		SQLChar = "char"
		SQLBit  = "bit"

	*/
)

type (
	Column struct {
		Name      string `yaml:"name"`
		MongoType string `yaml:"type"`
		SqlName   string `yaml:"sqlname"`
		SqlType   string `yaml:"sqltype"`
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

	// MySQL maintains three different time zone settings:
	// - system: host machine time zone
	// - server: server's current time zone
	// - per-connection time zone (using SET)
	// TODO: We might need to update the default locale (time
	// zone) we use in parsing time related types from MongoDB.
	// See https://dev.mysql.com/doc/refman/5.6/en/time-zone-support.html
	DefaultLocale = time.UTC
	DefaultTime   = time.Date(0, 0, 0, 0, 0, 0, 0, DefaultLocale)

	// TimestampCtorFormats holds the various formats for constructing
	// the timestamp
	TimestampCtorFormats = []string{
		"15:4:5",
		"2006-1-2",
		"2006-1-2 15:4:5",
		"2006-1-2 15:4:5.000",
	}
)

func (c *Column) validateType() error {
	switch c.SqlType {

	case "":

	case SQLString:
	case SQLInt:
	case SQLFloat:
	case SQLBoolean:
	case SQLVarchar:

	case SQLYear:
	case SQLDateTime:
	case SQLTimestamp:
	case SQLTime:
	case SQLDate:

	case SQLNull:

	default:
		panic(fmt.Sprintf("don't know MySQL type: %s", c.SqlType))
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
