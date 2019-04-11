package ast

import (
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// Expr is implemented by all expressions in the AST.
type Expr interface {
	Node
	exprNode()
}

func (n *AggExpr) exprNode()              {}
func (n *Array) exprNode()                {}
func (n *ArrayIndexRef) exprNode()        {}
func (n *Binary) exprNode()               {}
func (n *Constant) exprNode()             {}
func (n *Document) exprNode()             {}
func (n *FieldOrArrayIndexRef) exprNode() {}
func (n *FieldRef) exprNode()             {}
func (n *Function) exprNode()             {}
func (n *Let) exprNode()                  {}
func (n *Conditional) exprNode()          {}
func (n *Unknown) exprNode()              {}
func (n *VariableRef) exprNode()          {}

// Ref is implemented by reference expressions in the AST.
type Ref interface {
	Expr
	ref() //nolint:unused
}

func (n *ArrayIndexRef) ref()        {}
func (n *FieldOrArrayIndexRef) ref() {}
func (n *FieldRef) ref()             {}
func (n *VariableRef) ref()          {}

// NewAggExpr makes an AggExpr.
func NewAggExpr(expr Expr) *AggExpr {
	return &AggExpr{
		Expr: expr,
	}
}

// AggExpr is an aggregation expression embedded into a query with $expr.
type AggExpr struct {
	Expr Expr
}

// NewArray makes an Array.
func NewArray(elements ...Expr) *Array {
	return &Array{elements}
}

// Array is an array creation expression.
type Array struct {
	Elements []Expr
}

// NewArrayIndexRef makes an ArrayIndexRef.
func NewArrayIndexRef(index Expr, parent Expr) *ArrayIndexRef {
	return &ArrayIndexRef{
		Index:  index,
		Parent: parent,
	}
}

// ArrayIndexRef is a reference to an array index in Expr.
type ArrayIndexRef struct {
	Index  Expr
	Parent Expr
}

// NewBinary makes a Binary.
func NewBinary(op BinaryOp, left Expr, right Expr) *Binary {
	return &Binary{
		Op:    op,
		Left:  left,
		Right: right,
	}
}

// BinaryOp is the type of binary operation.
type BinaryOp string

// constants for the possible binary operations.
const (
	And                 BinaryOp = BinaryOp("$and")
	Equals              BinaryOp = BinaryOp("$eq")
	GreaterThan         BinaryOp = BinaryOp("$gt")
	GreaterThanOrEquals BinaryOp = BinaryOp("$gte")
	LessThan            BinaryOp = BinaryOp("$lt")
	LessThanOrEquals    BinaryOp = BinaryOp("$lte")
	Nor                 BinaryOp = BinaryOp("$nor")
	NotEquals           BinaryOp = BinaryOp("$ne")
	Or                  BinaryOp = BinaryOp("$or")
)

// Flip flips the direction of a less than or greater than operator.
func (op BinaryOp) Flip() BinaryOp {
	switch op {
	case LessThan:
		return GreaterThan
	case LessThanOrEquals:
		return GreaterThanOrEquals
	case GreaterThan:
		return LessThan
	case GreaterThanOrEquals:
		return LessThanOrEquals
	default:
		return op
	}
}

// Binary is a binary expression.
type Binary struct {
	Op    BinaryOp
	Left  Expr
	Right Expr
}

// NewConstant makes a Constant.
func NewConstant(value bsoncore.Value) *Constant {
	return &Constant{Value: value}
}

// Constant is a literal value.
type Constant struct {
	Value bsoncore.Value
}

// NewDocument makes a document.
func NewDocument(elems ...*DocumentElement) *Document {
	return &Document{elems}
}

// NewDocumentElement makes a document element.
func NewDocumentElement(name string, expr Expr) *DocumentElement {
	return &DocumentElement{name, expr}
}

// DocumentElement is an element of a Document.
type DocumentElement struct {
	Name string
	Expr Expr
}

// Document is a document creation expression.
type Document struct {
	Elements []*DocumentElement
}

// FieldsMap returns the Elements of a Document as a
// map from string to Expr.
func (n *Document) FieldsMap() map[string]Expr {
	elements := make(map[string]Expr)
	for _, e := range n.Elements {
		elements[e.Name] = e.Expr
	}
	return elements
}

// NewFieldOrArrayIndexRef makes a FieldOrArrayIndexRef.
func NewFieldOrArrayIndexRef(number int32, parent Expr) *FieldOrArrayIndexRef {
	return &FieldOrArrayIndexRef{
		Number: number,
		Parent: parent,
	}
}

// FieldOrArrayIndexRef is a reference to an array index in Expr.
type FieldOrArrayIndexRef struct {
	Number int32
	Parent Expr
}

// NewFieldRef makes a FieldRef.
func NewFieldRef(name string, parent Expr) *FieldRef {
	return &FieldRef{
		Name:   name,
		Parent: parent,
	}
}

// FieldRef is a reference to a field.
type FieldRef struct {
	Name   string
	Parent Expr
}

// NewLet makes a Let.
func NewLet(variables []*LetVariable, expr Expr) *Let {
	return &Let{
		Variables: variables,
		Expr:      expr,
	}
}

// NewLetVariable makes a LetVariable.
func NewLetVariable(name string, expr Expr) *LetVariable {
	return &LetVariable{
		Name: name,
		Expr: expr,
	}
}

// LetVariable specifies a variable to be bound by a Let.
type LetVariable struct {
	Name string
	Expr Expr
}

// Let binds one or more variables to the values of the specified expressions.
type Let struct {
	Variables []*LetVariable
	Expr      Expr
}

// NewConditional make a Conditional.
func NewConditional(ifClause Expr, thenClause Expr, elseClause Expr) *Conditional {
	return &Conditional{
		If:   ifClause,
		Then: thenClause,
		Else: elseClause,
	}
}

// Conditional selects an expression based on a condition.
type Conditional struct {
	If   Expr
	Then Expr
	Else Expr
}

// NewFunction makes a Function.
func NewFunction(name string, arg Expr) *Function {
	return &Function{name, arg}
}

// Function is a function expression.
type Function struct {
	Name string
	Arg  Expr
}

// NewVariableRef makes a VariableRef.
func NewVariableRef(name string) *VariableRef {
	return &VariableRef{Name: name}
}

// VariableRef is a reference to a variable.
type VariableRef struct {
	Name string
}
