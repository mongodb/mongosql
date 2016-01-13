package schema

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func computeDirectory(originalFileName string, schemaDir string) string {
	if schemaDir[0] == '/' || schemaDir[0] == '\\' {
		return schemaDir
	}

	d, _ := filepath.Split(originalFileName)
	return filepath.Join(d, schemaDir)
}

func ParseSchemaData(data []byte) (*Schema, error) {
	var schema Schema
	if err := yaml.Unmarshal([]byte(data), &schema); err != nil {
		return nil, err
	}

	schema.Databases = make(map[string]*Database)

	for _, db := range schema.RawDatabases {

		if _, ok := schema.Databases[db.Name]; ok {
			return nil, fmt.Errorf("duplicate database [%s].", db.Name)
		}

		if err := PopulateColumnMaps(db); err != nil {
			return nil, err
		}

		schema.Databases[db.Name] = db

	}

	return &schema, nil
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

func PopulateColumnMaps(db *Database) error {

	db.Tables = make(map[string]*Table)

	for _, tbl := range db.RawTables {

		if len(strings.SplitN(tbl.FQNS, ".", 2)) != 2 {
			return fmt.Errorf("invalid collection mapping '%v' (must contain '.') in db '%s' on table '%s'", tbl.FQNS, db.Name, tbl.Name)
		}

		err := tbl.fixTypes()
		if err != nil {
			return err
		}

		// TODO: consider lowercasing table names in config since we do
		// that in the query planner in constructing the TableScan node.
		if _, ok := db.Tables[tbl.Name]; ok {
			return fmt.Errorf("duplicate table [%s].", tbl.Name)
		}

		db.Tables[tbl.Name] = tbl

		tbl.Columns = make(map[string]*Column)
		tbl.SQLColumns = make(map[string]*Column)

		for _, c := range tbl.RawColumns {

			if c.SqlName == "" {
				c.SqlName = c.Name
			}

			if c.SqlName == "" {
				return fmt.Errorf("table [%s] has column with no name.", tbl.Name)
			}

			if _, ok := tbl.Columns[c.Name]; ok {
				return fmt.Errorf("duplicate column [%s].", c.Name)
			}

			tbl.Columns[c.Name] = c
			tbl.SQLColumns[c.SqlName] = c
		}
	}

	return nil
}
