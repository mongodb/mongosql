package ast

// The memory usage returned by the functions in this file is the amount of
// memory stored in BSON values and strings that can be of arbitrary size. It
// does not include pointers and other structural components of the AST. Its
// purposes is to provide a heuristic that increases as the AST grows so that
// we can prevent unlimited blow-up of memory usage during evaluation, not to
// provide an exact count of the memory used by the AST.

// MemoryUsage implements the Expr interface.
func (n *AggExpr) MemoryUsage() uint64 {
	return n.Expr.MemoryUsage()
}

// MemoryUsage implements the Expr interface.
func (n *Array) MemoryUsage() uint64 {
	var mem uint64
	for _, e := range n.Elements {
		mem += e.MemoryUsage()
	}
	return mem
}

// MemoryUsage implements the Expr interface.
func (n *ArrayIndexRef) MemoryUsage() uint64 {
	return n.Index.MemoryUsage() + n.Parent.MemoryUsage()
}

// MemoryUsage implements the Expr interface.
func (n *Unary) MemoryUsage() uint64 {
	return n.Expr.MemoryUsage()
}

// MemoryUsage implements the Expr interface.
func (n *Binary) MemoryUsage() uint64 {
	return n.Left.MemoryUsage() + n.Right.MemoryUsage()
}

// MemoryUsage implements the Expr interface.
func (n *Trunc) MemoryUsage() uint64 {
	return n.Number.MemoryUsage() + n.Precision.MemoryUsage()
}

// MemoryUsage implements the Expr interface.
func (n *Constant) MemoryUsage() uint64 {
	return uint64(len(n.Value.Data))
}

// MemoryUsage implements the Expr interface.
func (n *Document) MemoryUsage() uint64 {
	var mem uint64
	for _, e := range n.Elements {
		mem += uint64(len(e.Name)) + e.Expr.MemoryUsage()
	}
	return mem
}

// MemoryUsage implements the Expr interface.
func (n *FieldOrArrayIndexRef) MemoryUsage() uint64 {
	if n.Parent == nil {
		return 0
	}
	return n.Parent.MemoryUsage()
}

// MemoryUsage implements the Expr interface.
func (n *FieldRef) MemoryUsage() uint64 {
	var pmem uint64
	if n.Parent != nil {
		pmem = n.Parent.MemoryUsage()
	}
	return uint64(len(n.Name)) + pmem
}

// MemoryUsage implements the Expr interface.
func (n *Let) MemoryUsage() uint64 {
	var mem uint64
	for _, v := range n.Variables {
		mem += uint64(len(v.Name)) + v.Expr.MemoryUsage()
	}
	return mem + n.Expr.MemoryUsage()
}

// MemoryUsage implements the Expr interface.
func (n *Conditional) MemoryUsage() uint64 {
	return n.If.MemoryUsage() + n.Then.MemoryUsage() + n.Else.MemoryUsage()
}

// MemoryUsage implements the Expr interface.
func (n *Map) MemoryUsage() uint64 {
	return n.Input.MemoryUsage() + uint64(len(n.As)) + n.In.MemoryUsage()
}

// MemoryUsage implements the Expr interface.
func (n *Filter) MemoryUsage() uint64 {
	return n.Input.MemoryUsage() + uint64(len(n.As)) + n.Cond.MemoryUsage()
}

// MemoryUsage implements the Expr interface.
func (n *Reduce) MemoryUsage() uint64 {
	return n.Input.MemoryUsage() + n.InitialValue.MemoryUsage() + n.In.MemoryUsage()
}

// MemoryUsage implements the Expr interface.
func (n *Function) MemoryUsage() uint64 {
	return uint64(len(n.Name)) + n.Arg.MemoryUsage()
}

// MemoryUsage implements the Expr interface.
func (n *MatchRegex) MemoryUsage() uint64 {
	return n.Expr.MemoryUsage() + uint64(len(n.Pattern)) + uint64(len(n.Options))
}

// MemoryUsage implements the Expr interface.
func (n *VariableRef) MemoryUsage() uint64 {
	return uint64(len(n.Name))
}

// MemoryUsage implements the Expr interface.
func (n *Exists) MemoryUsage() uint64 {
	return n.FieldRef.MemoryUsage()
}

// MemoryUsage implements the Expr interface.
func (n *Unknown) MemoryUsage() uint64 {
	return uint64(len(n.Value.Data))
}
