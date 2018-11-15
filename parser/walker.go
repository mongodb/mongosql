package parser

import (
	"reflect"
)

// A parser.walker visits each node in a CST allowing in-place mutation.
// Implement this interface for each phase of operation on the CST.
type walker interface {
	// PreVisit is called for every node before its children are walked.
	// Mutation of the current node is enabled through the current
	// interface.
	// Mutation of current node is enabled by the return value.
	// No change is implemented by returning the unmodified argument.
	// nils are set in place of the current node if returned.
	PreVisit(current CST) (CST, error)
	// PostVisit is called for every node after its children are walked.
	// Mutation of the current node is enabled through the current
	// interface.
	// Mutation of current node is enabled by the return value.
	// No change is implemented by returning the unmodified argument.
	// nils are set in place of the current node if returned.
	PostVisit(current CST) (CST, error)
}

// walk implements the routing logic to visit each node in a CST tree.
// The walk is ended eagerly if an error occurs.
func walk(w walker, node CST) (CST, error) {
	node, err := w.PreVisit(node)
	if err != nil {
		return nil, err
	}

	// We can walk to nils but we cannot take their children.
	// Reflection is necessary due to the nil error:
	// https://golang.org/doc/faq#nil_error
	if node != nil {
		val := reflect.ValueOf(node)
		if val.Kind() == reflect.String || !val.IsNil() {
			for i, child := range node.Children() {
				replacement, werr := walk(w, child)
				if werr != nil {
					return nil, werr
				}
				node.ReplaceChild(i, replacement)
			}
		}
	}

	node, err = w.PostVisit(node)
	if err != nil {
		return nil, err
	}
	return node, nil
}
