package config

import (
	"io"
	"reflect"

	"github.com/10gen/candiedyaml"
)

// ParseYaml parses the yaml in the reader into the cfg.
func ParseYaml(cfg *Config, r io.Reader, enabledExpansions EnabledExpansions) error {
	decoder := candiedyaml.NewDecoder(r)
	decoder.StrictMode(true)

	root := make(map[interface{}]interface{})
	err := decoder.Decode(&root)
	if err != nil {
		return err
	}

	err = fromMap("", reflect.ValueOf(cfg), root, newConfigExpander(enabledExpansions))
	if err != nil {
		return err
	}

	return postProcess(cfg)
}

// postProcess performs some transformations on a parsed config. Notably, it
// moves values provided at deprecated config field paths to their updated
// fields.
func postProcess(cfg *Config) error {
	if cfg.Schema.RefreshIntervalSecs == DefaultRefreshIntervalSecs {
		cfg.Schema.RefreshIntervalSecs = cfg.Schema.Sample.RefreshIntervalSecsDeprecated
	}
	cfg.Schema.Sample.RefreshIntervalSecsDeprecated = DefaultRefreshIntervalSecs
	return nil
}
