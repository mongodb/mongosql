package config

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
)

type Column struct {
	Name      string `yaml:"name"`
	Type      string `yaml:"type"`
	Source    string `yaml:"source"`
	MysqlType string `yaml:"mysql_type,omitempty"`
}

type TableConfig struct {
	Table      string    `yaml:"table"`
	Collection string    `yaml:"collection"`
	Pipeline   []bson.M  `yaml:"pipeline,omitempty"`
	Columns    []*Column `yaml:"columns"`
}

type Schema struct {
	DB        string                  `yaml:"db"`
	RawTables []*TableConfig          `yaml:"tables"`
	Tables    map[string]*TableConfig `yaml:"tables_no"`
}

type Config struct {
	Addr     string `yaml:"addr"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	LogLevel string `yaml:"log_level"`

	Url string `yaml:"url"`

	SchemaDir string `yaml:"schema_dir"`

	RawSchemas []*Schema          `yaml:"schema"`
	Schemas    map[string]*Schema `yaml:"schema_no"`
}

// ---

func (c *Column) fixType() error {
	switch c.Type {
	case "string":
		c.MysqlType = "varchar(2048)"
	case "int":
		c.MysqlType = "int(11)"
	default:
		return fmt.Errorf("don't know mysql equivilant for type: %s", c.Type)
	}
	return nil
}

func (t *TableConfig) fixTypes() error {
	for _, c := range t.Columns {
		err := c.fixType()
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
