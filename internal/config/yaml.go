package config

import (
	"io"
	"reflect"

	"github.com/10gen/candiedyaml"
)

// ParseYaml parses the yaml in the reader into the cfg.
func ParseYaml(cfg *Config, r io.Reader) error {
	decoder := candiedyaml.NewDecoder(r)
	decoder.StrictMode(true)

	root := make(map[interface{}]interface{})
	err := decoder.Decode(&root)
	if err != nil {
		return err
	}

	return fromMap("", reflect.ValueOf(cfg), root)
}
