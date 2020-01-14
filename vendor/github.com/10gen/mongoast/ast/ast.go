package ast

import "go.mongodb.org/mongo-driver/x/bsonx/bsoncore"

// Node is implemented by every struct part of the AST.
type Node interface {
	DeepCopier

	Walk(v Visitor) Node
}

// NewPipeline makes a Pipeline.
func NewPipeline(stages ...Stage) *Pipeline {
	return &Pipeline{stages}
}

// Pipeline is composed of a number of stages.
type Pipeline struct {
	Stages []Stage
}

// NewUnknown makes an Unknown.
func NewUnknown(value bsoncore.Value) *Unknown {
	return &Unknown{value}
}

// Unknown handles any unknown stage, expression, or value
// that cannot be introspected on.
type Unknown struct {
	Value bsoncore.Value
}

// StageName implements the Stage interface.
func (n *Unknown) StageName() string {
	doc, _ := n.Value.DocumentOK()
	elems, _ := doc.Elements()
	return elems[0].Key()
}
