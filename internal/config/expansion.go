package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	"github.com/10gen/candiedyaml"

	"github.com/10gen/sqlproxy/internal/httputil"
)

var isExpansionField = map[string]struct{}{
	"__exec": {},
	"__rest": {},
	"type":   {},
}

// configExpander holds information about expansion directive evaluation.
type configExpander struct {
	enabledExpansions EnabledExpansions
	isRoot            bool
	isExpandedYaml    bool
}

func newConfigExpander(e EnabledExpansions) configExpander {
	return configExpander{
		enabledExpansions: e,
		isRoot:            true,
		isExpandedYaml:    false,
	}
}

// parse checks if the specified yaml document has an expansion directive, and if so, it parses
// it into an expansionBlock.
func (e *configExpander) parse(yaml map[interface{}]interface{}) (*expansionBlock, error) {
	if e.isRoot && len(yaml) > 2 {
		return nil, fmt.Errorf("a top-level expansion directive and its options must be the only values specified in the config file")
	}
	expansionBlock := newExpansionBlock()
	for field, value := range yaml {
		switch field.(string) {
		case "__exec":
			if !e.enabledExpansions.Exec {
				return nil, fmt.Errorf("__exec has not been enabled via --configExpand")
			}
			expansionBlock.exec = value.(string)
		case "__rest":
			if !e.enabledExpansions.Rest {
				return nil, fmt.Errorf("__rest has not been enabled via --configExpand")
			}
			expansionBlock.rest = value.(string)
		case "type":
			if err := validateType(value.(string), e.isRoot); err != nil {
				return nil, err
			}
			expansionBlock.typeOpt = value.(string)
		default:
			return nil, nil
		}
		if e.isExpandedYaml {
			return nil, fmt.Errorf("expansion directive recursion is not supported")
		}
	}
	if expansionBlock.rest != "" && expansionBlock.exec != "" {
		return nil, fmt.Errorf("can only use __exec or __rest, not both")
	}
	if e.isRoot && expansionBlock.typeOpt != "yaml" {
		return nil, fmt.Errorf("set {type: 'yaml'} if the config has a top-level expansion directive")
	}
	return expansionBlock, nil
}

// evaluate checks if `val` has a valid expansion directive - if so, it will be evaluated and returned.
func (e *configExpander) evaluate(val interface{}) (interface{}, error) {
	tval, ok := val.(map[interface{}]interface{})
	if ok && (e.enabledExpansions.Exec || e.enabledExpansions.Rest) {
		expansionBlock, err := e.parse(tval)
		if err != nil {
			return nil, err
		}
		if expansionBlock == nil {
			return val, nil
		}
		var eval string
		var evalErr error
		if expansionBlock.exec != "" {
			eval, evalErr = expansionBlock.evaluateExec()
		}
		if expansionBlock.rest != "" {
			eval, evalErr = expansionBlock.evaluateRest()
		}
		if evalErr != nil {
			return nil, evalErr
		}
		if expansionBlock.typeOpt == "yaml" {
			return parseYaml(eval)
		}
		return eval, nil
	}
	return val, nil
}

type expansionBlock struct {
	exec    string
	rest    string
	typeOpt string
}

func newExpansionBlock() *expansionBlock {
	return &expansionBlock{
		exec:    "",
		rest:    "",
		typeOpt: "string",
	}
}

// evaluateExec evaluates an exec command as if it's a shell or terminal command.
func (e *expansionBlock) evaluateExec() (string, error) {
	args := strings.Split(e.exec, " ")
	var result []byte
	var err error
	if len(args) == 1 {
		result, err = exec.Command(args[0]).Output()
	} else {
		result, err = exec.Command(args[0], args[1:]...).Output()
	}
	if err != nil {
		return "", fmt.Errorf("error executing '__exec' command: %v", err)
	}

	return strings.TrimSuffix(string(result), "\n"), nil
}

// evaluateRest makes a GET request to the specified url and returns the response body. If the
// response has a non-200 status code, an error will be returned.
func (e *expansionBlock) evaluateRest() (string, error) {
	response, err := httputil.Get(e.rest)
	if err != nil {
		return "", fmt.Errorf("GET request failed against url %v: %v", e.rest, err)
	}
	if response.StatusCode != 200 {
		return "", fmt.Errorf("GET request resulted in status code %v from url %v", response.StatusCode, e.rest)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read GET request response body from url %v: %v", e.rest, err)
	}
	return string(body), nil
}

// parseYaml decodes the specified string into a new yaml document.
func parseYaml(value string) (map[interface{}]interface{}, error) {
	yaml := make(map[interface{}]interface{})
	decoder := candiedyaml.NewDecoder(bytes.NewReader([]byte(value)))
	decoder.StrictMode(true)
	err := decoder.Decode(&yaml)
	if err != nil {
		return nil, err
	}
	return yaml, nil
}

// validateType returns an error if the value of "type" is invalid.
func validateType(typeOpt string, isRoot bool) error {
	switch typeOpt {
	case "string":
		return nil
	case "yaml":
		if !isRoot {
			return fmt.Errorf("{type: 'yaml'} is only supported for top-level expansion directives")
		}
		return nil
	default:
		return fmt.Errorf("invalid config: {type: \"%v\"}", typeOpt)
	}
}
