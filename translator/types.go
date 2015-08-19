package translator

type FieldComp struct {
	Left       string
	Comparator interface{}
	Right      string
}

type LOJ struct {
}

type ROJ struct{}
type FOJ struct{}
type COJ struct{}
type AsExpr struct{}
type UnionAll struct{}
type Union struct{}
type Relation struct{}
type CrltSubquery struct{}

type NumVal struct {
	Value interface{}
}

type ValTuple struct {
	Children []interface{}
}

type NullVal struct{}
type ColName struct {
	Value string
}

type StrVal struct {
	Value string
}

type BinaryExpr struct {
	Left     interface{}
	Operator interface{}
	Right    interface{}
}

type AndExpr struct {
	Left  interface{}
	Right interface{}
}

type OrExpr struct {
	Left  interface{}
	Right interface{}
}

type ComparisonExpr struct {
	Left       interface{}
	Comparator interface{}
	Right      interface{}
}

type RangeCond struct {
	From  interface{}
	Value string
	To    interface{}
}

type NullCheck struct {
	Value string
}

type UnaryExpr struct {
	Value interface{}
}

type NotExpr struct {
	Value interface{}
}

type ParenBoolExpr struct {
	Value interface{}
}

type Subquery struct {
	Query interface{}
}

type ValArg struct{}
type FuncExpr struct{}
type CaseExpr struct{}
type ExistsExpr struct{}
