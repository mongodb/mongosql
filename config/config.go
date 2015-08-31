package config

import (
	"gopkg.in/mgo.v2/bson"
)

type TableConfig struct {
	Table      string   `yaml:"table"`
	Collection string   `yaml:"collection"`
	Pipeline   []bson.M `yaml:"pipeline"`
	Columns map[string]string `yaml:"columns"`
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
