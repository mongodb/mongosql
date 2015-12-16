package config

import (
	"fmt"
	"time"
)

const (
	SQLString  string = "string"
	SQLInt     string = "int"
	SQLFloat   string = "float"
	SQLBlob    string = "text"
	SQLVarchar string = "varchar"

	SQLDate      string = "date"
	SQLDatetime  string = "datetime"
	SQLTime      string = "time"
	SQLTimestamp string = "timestamp"
	SQLYear      string = "year"

	SQLDecimal  string = "decimal"
	SQLDouble   string = "double"
	SQLEnum     string = "enum"
	SQLGeometry string = "geometry"

	SQLBigInt   string = "bigint"
	SQLMedInt   string = "mediumint"
	SQLSmallInt string = "smallint"
	SQLTiny     string = "tinyint"

	SQLLongText   string = "longtext"
	SQLTinyText   string = "tinytext"
	SQLMediumText string = "mediumtext"

	SQLNull string = "null"
	SQLSet  string = "set"
	SQLChar string = "char"
	SQLBit  string = "bit"
)

type (
	Column struct {
		Name      string `yaml:"name"`
		MongoType string `yaml:"type"`
		SqlName   string `yaml:"sql_name"`
		SqlType   string `yaml:"sql_type"`
	}

	Table struct {
		Name           string                   `yaml:"table"`
		CollectionName string                   `yaml:"collection"`
		Pipeline       []map[string]interface{} `yaml:"pipeline"`
		Columns        []*Column                `yaml:"columns"`
		Parent         *Table                   `yaml:"-"`
	}

	Schema struct {
		DB        string            `yaml:"db"`
		RawTables []*Table          `yaml:"tables"`
		Tables    map[string]*Table `yaml:"tables_no"`
	}

	Config struct {
		Addr     string `yaml:"addr"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		LogLevel string `yaml:"log_level"`

		Url       string `yaml:"url"`
		SchemaDir string `yaml:"schema_dir"`

		RawSchemas []*Schema          `yaml:"schema"`
		Schemas    map[string]*Schema `yaml:"schema_no"`
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
)

func (c *Column) validateType() error {
	switch c.SqlType {

	case "":

	case SQLString:
	case SQLInt:
	case SQLFloat:
	case SQLBlob:
	case SQLVarchar:

	case SQLYear:
	case SQLDatetime:
	case SQLTimestamp:
	case SQLTime:
	case SQLDate:

	case SQLDecimal:
	case SQLDouble:
	case SQLEnum:
	case SQLGeometry:

	case SQLBigInt:
	case SQLMedInt:
	case SQLSmallInt:
	case SQLTiny:

	case SQLLongText:
	case SQLTinyText:
	case SQLMediumText:

	case SQLNull:
	case SQLSet:
	case SQLChar:
	case SQLBit:

	default:
		panic(fmt.Sprintf("don't know MySQL type: %s", c.SqlType))
	}
	return nil
}

func (t *Table) fixTypes() error {
	for _, c := range t.Columns {
		err := c.validateType()
		if err != nil {
			return err
		}
	}
	return nil
}

// Must is a helper that wraps a call to a function returning (*Config, error)
// and panics if the error is non-nil. It is intended for use in variable
// initializations such as
//	var t = config.Must(ParseConfigData(raw))

func Must(c *Config, err error) *Config {
	if err != nil {
		panic(err)
	}
	return c
}

func (c *Config) ingestSubFile(data []byte) error {
	temp, err := ParseConfigData(data)
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

	for name, schema := range temp.Schemas {

		ours := c.Schemas[name]

		if ours == nil {
			// entire db missing
			c.Schemas[name] = schema
			c.RawSchemas = append(c.RawSchemas, schema)
			continue
		}

		// have to merge tables
		for table, tableConfig := range schema.Tables {
			if ours.Tables[table] != nil {
				return fmt.Errorf("table config conflict db: %s table: %s", name, table)
			}
			ours.Tables[table] = tableConfig
			ours.RawTables = append(ours.RawTables, tableConfig)
		}

	}

	return nil
}
