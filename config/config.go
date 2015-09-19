package config

import (
	"gopkg.in/mgo.v2/bson"
	"fmt"
)

type Column struct {
	Name      string `yaml:"name"`
	Type      string `yaml:"type"`
	MysqlType string `yaml:"mysql_type"`
}

type TableConfig struct {
	Table      string   `yaml:"table"`
	Collection string   `yaml:"collection"`
	Pipeline   []bson.M `yaml:"pipeline"`
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

	RawSchemas []*Schema          `yaml:"schema"`
	Schemas    map[string]*Schema `yaml:"schema_no"`
}

// ---

func (c *Column)fixType() error {
	switch (c.Type) {
	case "string":
		c.MysqlType = "varchar(2048)"
	case "int":
		c.MysqlType = "int(11)"
	default:
		return fmt.Errorf("don't know mysql equivilant for type: %s", c.Type)
	}
	return nil
}

func (t *TableConfig)fixTypes() error {
	for _, c := range(t.Columns) {
		err := c.fixType()
		if err != nil {
			return err
		}
	}
	return nil
}
