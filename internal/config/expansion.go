package config

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	"github.com/10gen/sqlproxy/internal/httputil"
)

var isExpansionDir = map[string]struct{}{"__exec": {}, "__rest": {}}

// evaluateExpansionDirective checks if `val` has a valid expansion directive - if so, it will be
// evaluated and returned.
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
		if expansions["__rest"] != "" && expansions["__exec"] != "" {
			return "", fmt.Errorf("invalid config: can only use __exec or __rest, not both")
		}
		if expansions["__exec"] != "" {
			if enabledExpansions.Exec {
				return evaluateExec(expansions["__exec"])
			}
			return "", fmt.Errorf("__exec has not been enabled via --configExpand")
		}
		if expansions["__rest"] != "" {
			if enabledExpansions.Rest {
				return evaluateRest(expansions["__rest"])
			}
			return "", fmt.Errorf("__rest has not been enabled via --configExpand")
		}
	}
	return val, nil
}

// evaluateExec evaluates the specified command as if it's a shell or terminal command.
func evaluateExec(command string) (string, error) {
	args := strings.Split(command, " ")
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

// evaluateRest makes a GET request to the specified url and returns the response body.
// If the response has a non-200 status code, an error will be returned.
func evaluateRest(url string) (string, error) {
	response, err := httputil.Get(url)
	if err != nil {
		return "", fmt.Errorf("GET request failed against url %v: %v", url, err)
	}
	if response.StatusCode != 200 {
		return "", fmt.Errorf("GET request resulted in status code %v from url %v", response.StatusCode, url)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read GET request response body from url %v: %v", url, err)
	}
	return string(body), nil
}
