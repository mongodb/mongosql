package evaluator

import (
	"context"
)

// Node is an interface that represents an AST node.
type Node interface {
	// Children returns a slice of all the Node children of the Node.
	Children() []Node
	// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
	ReplaceChild(i int, n Node)
}

// Command is an interface for plan stages that are also SQL commands.
type Command interface {
	Node
	// Execute executes the command.
	Execute(context.Context, *ExecutionConfig, *ExecutionState) error
}

type nodeVisitor interface {
	visit(n Node) (Node, error)
}

// walk handles walking the children of the provided expression, calling
// v.visit on each child. Some visitor implementations may ignore this
// method completely, but most will use it as the default implementation
// for a majority of nodes. It potentially modifies the tree it is walking.
func walk(v nodeVisitor, n Node) (Node, error) {
	for i, child := range n.Children() {
		newChild, err := v.visit(child)
		if err != nil {
			return nil, err
		}
		n.ReplaceChild(i, newChild)
	}
	return n, nil
}
