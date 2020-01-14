package ast

import (
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// Expr is implemented by all expressions in the AST.
type Expr interface {
	Node

	// WalkExpr visit the expression node.
	WalkExpr(v Visitor) Expr

	// MemoryUsage returns a heuristic approximating the amount of memory used
	// by the expression, including any subexpressions.
	MemoryUsage() uint64
}

// Ref is implemented by reference expressions in the AST.
type Ref interface {
	Expr
	WalkRef(v Visitor) Ref
}

// FieldLikeRef is implemented by reference expressions that refer to fields or indexes in documents.
type FieldLikeRef interface {
	Ref
	WalkFieldLikeRef(v Visitor) FieldLikeRef
	ParentExpr() Expr
}

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

// ParentExpr implements the FieldLikeRef interface.
func (n *ArrayIndexRef) ParentExpr() Expr {
	return n.Parent
}

// NewUnary makes a Unary.
func NewUnary(op UnaryOp, expr Expr) *Unary {
	return &Unary{
		Op:   op,
		Expr: expr,
	}
}

// UnaryOp is the type of unary operation.
type UnaryOp string

// constants for the possible unary operations.
const (
	Not   UnaryOp = UnaryOp("$not")
	Abs   UnaryOp = UnaryOp("$abs")
	Ceil  UnaryOp = UnaryOp("$ceil")
	Floor UnaryOp = UnaryOp("$floor")
	Exp   UnaryOp = UnaryOp("$exp")
	Ln    UnaryOp = UnaryOp("$ln")
	Log10 UnaryOp = UnaryOp("$log10")
	Sqrt  UnaryOp = UnaryOp("$sqrt")
)

// Unary is a unary expression.
type Unary struct {
	Op   UnaryOp
	Expr Expr
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
	Compare             BinaryOp = BinaryOp("$cmp")
	Equals              BinaryOp = BinaryOp("$eq")
	GreaterThan         BinaryOp = BinaryOp("$gt")
	GreaterThanOrEquals BinaryOp = BinaryOp("$gte")
	LessThan            BinaryOp = BinaryOp("$lt")
	LessThanOrEquals    BinaryOp = BinaryOp("$lte")
	Nor                 BinaryOp = BinaryOp("$nor")
	NotEquals           BinaryOp = BinaryOp("$ne")
	Or                  BinaryOp = BinaryOp("$or")
	Divide              BinaryOp = BinaryOp("$divide")
	Log                 BinaryOp = BinaryOp("$log")
	Mod                 BinaryOp = BinaryOp("$mod")
	Pow                 BinaryOp = BinaryOp("$pow")
	Subtract            BinaryOp = BinaryOp("$subtract")
	Add                 BinaryOp = BinaryOp("$add")
	Multiply            BinaryOp = BinaryOp("$multiply")
	Concat              BinaryOp = BinaryOp("$concat")
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

// NewTrunc makes a Trunc.
func NewTrunc(number, precision Expr) *Trunc {
	return &Trunc{
		Number:    number,
		Precision: precision,
	}
}

// Trunc is the trunc expression, which truncates a number to a precision.
type Trunc struct {
	Number    Expr
	Precision Expr
}

// NewConstant makes a Constant.
// hasLiteral is optional in order to not break the previous interface.
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

// ParentExpr implements the FieldLikeRef interface.
func (n *FieldOrArrayIndexRef) ParentExpr() Expr {
	return n.Parent
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

// ParentExpr implements the FieldLikeRef interface.
func (n *FieldRef) ParentExpr() Expr {
	return n.Parent
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

// NewMap makes a new Map Expression.
func NewMap(input Expr, as string, in Expr) *Map {
	return &Map{Input: input, As: as, In: in}
}

// Map is the map expression, which maps a function (In) over an array (Input).
type Map struct {
	Input Expr
	As    string
	In    Expr
}

var _ Expr = &Map{}

// NewFilter makes a new Filter Expression.
func NewFilter(input Expr, as string, cond Expr) *Filter {
	return &Filter{Input: input, As: as, Cond: cond}
}

var _ Expr = &Filter{}

// Filter is the filter expression, which filters an array (Input) with a predicate (Cond).
type Filter struct {
	Input Expr
	As    string
	Cond  Expr
}

// NewReduce makes a new Reduce Expression.
func NewReduce(input Expr, initialValue Expr, in Expr) *Reduce {
	return &Reduce{Input: input, InitialValue: initialValue, In: in}
}

// Reduce applies an expression [In] to each element and combines them into a single value.
type Reduce struct {
	Input        Expr
	InitialValue Expr
	In           Expr
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

// NewMatchRegex creates a new MatchRegex.
func NewMatchRegex(expr Expr, pattern, options string) *MatchRegex {
	return &MatchRegex{
		Expr:    expr,
		Pattern: pattern,
		Options: options,
	}
}

// MatchRegex is the match language $regex function.
type MatchRegex struct {
	Expr    Expr
	Pattern string
	Options string
}

// NewVariableRef makes a VariableRef.
func NewVariableRef(name string) *VariableRef {
	return &VariableRef{Name: name}
}

// VariableRef is a reference to a variable.
type VariableRef struct {
	Name string
}

// Exists filters based on whether a document has a field.
type Exists struct {
	FieldRef *FieldRef
	Exists   bool
}

// NewExists makes an Exists.
func NewExists(fieldRef *FieldRef, exists bool) *Exists {
	return &Exists{FieldRef: fieldRef, Exists: exists}
}

// NewMergeObjects makes a MergeObjects.
func NewMergeObjects(exprs ...Expr) *MergeObjects {
	return &MergeObjects{Exprs: exprs}
}

// MergeObjects combines multiple documents into one document.
type MergeObjects struct {
	Exprs []Expr
}
