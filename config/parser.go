package config

import (
	"fmt"
	"github.com/siddontang/go-yaml/yaml"
	"io/ioutil"
	"path/filepath"
	"strings"
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
			err := n.fixTypes()
			if err != nil {
				return nil, err
			}
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

	cfg, err := ParseConfigData(data)
	if err != nil {
		return nil, err
	}

	if len(cfg.SchemaDir) > 0 {
		newDir := computeDirectory(fileName, cfg.SchemaDir)

		files, err := ioutil.ReadDir(newDir)
		if err != nil {
			return nil, err
		}

		for _, f := range files {
			if f.IsDir() {
				continue
			}
			if strings.ContainsAny(f.Name(), "#~") {
				continue
			}

			subData, err := ioutil.ReadFile(filepath.Join(newDir, f.Name()))
			if err != nil {
				return nil, err
			}

			err = cfg.ingestSubFile(subData)
			if err != nil {
				return nil, err
			}

		}
	}

	return cfg, nil
}

// --

func computeDirectory(originalFileName string, schemaDir string) string {
	if schemaDir[0] == '/' || schemaDir[0] == '\\' {
		return schemaDir
	}

	d, _ := filepath.Split(originalFileName)
	return filepath.Join(d, schemaDir)
}
