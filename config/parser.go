package config

import (
	"fmt"
	"github.com/siddontang/go-yaml/yaml"
	"io/ioutil"
)

func ParseConfigData(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal([]byte(data), &cfg); err != nil {
		return nil, err
	}

	cfg.Schemas = make(map[string]*Schema)
	for _, schemaCfg := range cfg.RawSchemas {
		if _, ok := cfg.Schemas[schemaCfg.DB]; ok {
			return nil, fmt.Errorf("duplicate schema [%s].", schemaCfg.DB)
		}
		cfg.Schemas[schemaCfg.DB] = schemaCfg

		schemaCfg.Tables = make(map[string]*TableConfig)
		for _, n := range schemaCfg.RawTables {
			schemaCfg.Tables[n.Table] = n
		}

	}

	//fmt.Printf("%+v\n", cfg)

	return &cfg, nil
}

func ParseConfigFile(fileName string) (*Config, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return ParseConfigData(data)
}
