package schema

import (
	"fmt"
	"github.com/siddontang/go-yaml/yaml"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func ParseSchemaData(data []byte) (*Schema, error) {
	var cfg Schema
	if err := yaml.Unmarshal([]byte(data), &cfg); err != nil {
		return nil, err
	}

	cfg.Databases = make(map[string]*Database)

	for _, schemaCfg := range cfg.RawDatabases {

		if _, ok := cfg.Databases[schemaCfg.Name]; ok {
			return nil, fmt.Errorf("duplicate schema [%s].", schemaCfg.Name)
		}

		cfg.Databases[schemaCfg.Name] = schemaCfg

		schemaCfg.Tables = make(map[string]*Table)

		for _, n := range schemaCfg.RawTables {
			err := n.fixTypes()
			if err != nil {
				return nil, err
			}
			schemaCfg.Tables[n.Name] = n
		}

	}

	return &cfg, nil
}

func ParseSchemaFile(fileName string) (*Schema, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	cfg, err := ParseSchemaData(data)
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
