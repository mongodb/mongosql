package config

import (
	"fmt"
	"os/exec"
	"strings"
)

var isExpansionDir = map[string]struct{}{"__exec": {}, "__rest": {}}

// evaluateExpansionDirective checks if the user has enabled the usage of expansion directives. If the current
// field being processed has one, it will be evaluated and returned.
func evaluateExpansionDirective(enabledExpansions Expansion, val interface{}) (interface{}, error) {
	expansions := make(map[string]string)
	tval, ok := val.(map[interface{}]interface{})
	if ok && (enabledExpansions.Exec || enabledExpansions.Rest) {
		// Verify that each field is a valid expansion directive
		for field, value := range tval {
			if _, ok := isExpansionDir[field.(string)]; ok {
				expansions[field.(string)] = value.(string)
			} else {
				return val, nil
			}
		}
		return evaluate(expansions, enabledExpansions)
	}
	return val, nil
}

// evaluate executes the value of an '__exec' field as if it's a shell or terminal command
// and returns the output.
func evaluate(expansions map[string]string, enabledExpansions Expansion) (interface{}, error) {
	if expansions["__rest"] != "" && expansions["__exec"] != "" {
		return "", fmt.Errorf("invalid config: can only use __exec or __rest, not both")
	} else if expansions["__exec"] != "" {
		if enabledExpansions.Exec {
			args := strings.Split(expansions["__exec"], " ")
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

			// exec.Command() terminates its result with a newline.
			return strings.TrimSuffix(string(result), "\n"), nil
		}
		return "", fmt.Errorf("__exec has not been enabled via --configExpand")
	} else {
		// TODO: add __rest, type, trim, digest, and digest_key support
		return "", nil
	}
}
