package ast

// Visitor defines methods that are called for nodes during an expression or statement walk.
type Visitor interface {
	Visit(Node) Node
}

// Visit visits the node using the provided visitFunc and returns
// a node and an indication as to whether or not it is different.
func Visit(n Node, visitFn VisitFunc) (Node, bool) {
	return visitNode(visitFn, n)
}

// VisitFunc is a functional implementation of a Visitor.
type VisitFunc func(Visitor, Node) Node

// Visit implements the Visitor interface.
func (fn VisitFunc) Visit(n Node) Node {
	return fn(fn, n)
}

func visitNode(v Visitor, n Node) (Node, bool) {
	newNode := v.Visit(n)
	return newNode, newNode != n
}

func visitPipeline(v Visitor, p *Pipeline) (*Pipeline, bool) {
	newNode, changed := visitNode(v, p)
	return newNode.(*Pipeline), changed
}

func visitExpr(v Visitor, e Expr) (Expr, bool) {
	newNode, changed := visitNode(v, e)
	return newNode.(Expr), changed
}

func visitStage(v Visitor, s Stage) (Stage, bool) {
	newNode, changed := visitNode(v, s)
	return newNode.(Stage), changed
}

//---------------------------------
// Pipeline

// Walk implements the Node interface.
func (n *Pipeline) Walk(v Visitor) Node {
	changed := false
	var newStages []Stage
	for i, stage := range n.Stages {
		newStage, stageChanged := visitStage(v, stage)
		changed = changed || stageChanged

		if changed && newStages == nil {
			newStages = make([]Stage, i, len(n.Stages))
			copy(newStages, n.Stages[:i])
		}

		if changed {
			newStages = append(newStages, newStage)
		}
	}

	if changed {
		cpy := *n
		cpy.Stages = newStages
		return &cpy
	}
	return n
}

//---------------------------------
// Stages

// Walk implements the Node interface.
func (n *AddFieldsStage) Walk(v Visitor) Node {
	return n.WalkStage(v)
}

// WalkStage implements the Stage interface.
func (n *AddFieldsStage) WalkStage(v Visitor) Stage {
	changed := false
	var newItems []*AddFieldsItem
	for i, item := range n.Items {
		newItem, itemChanged := visitNode(v, item)
		changed = changed || itemChanged

		if changed && newItems == nil {
			newItems = make([]*AddFieldsItem, i, len(n.Items))
			copy(newItems, n.Items[:i])
		}

		if changed {
			newItems = append(newItems, newItem.(*AddFieldsItem))
		}
	}

	if changed {
		cpy := *n
		cpy.Items = newItems
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *BucketStage) Walk(v Visitor) Node {
	return n.WalkStage(v)
}

// WalkStage implements the Stage interface.
func (n *BucketStage) WalkStage(v Visitor) Stage {
	newGroupBy, changed := visitExpr(v, n.GroupBy)

	var newOutput []*GroupItem
	if n.Output != nil {
		for i, item := range n.Output {
			newItem, itemChanged := visitNode(v, item)
			changed = changed || itemChanged

			if itemChanged && newOutput == nil {
				newOutput = make([]*GroupItem, i, len(n.Output))
				copy(newOutput, n.Output[:i])
			}

			if newOutput != nil {
				newOutput = append(newOutput, newItem.(*GroupItem))
			}
		}
	}

	if changed {
		cpy := *n
		cpy.GroupBy = newGroupBy
		if newOutput != nil {
			cpy.Output = newOutput
		}
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *BucketAutoStage) Walk(v Visitor) Node {
	return n.WalkStage(v)
}

// WalkStage implements the Stage interface.
func (n *BucketAutoStage) WalkStage(v Visitor) Stage {
	newGroupBy, changed := visitExpr(v, n.GroupBy)

	var newOutput []*GroupItem
	if n.GroupBy != nil {
		for i, item := range n.Output {
			newItem, itemChanged := visitNode(v, item)
			changed = changed || itemChanged

			if itemChanged && newOutput == nil {
				newOutput = make([]*GroupItem, i, len(n.Output))
				copy(newOutput, n.Output[:i])
			}

			if newOutput != nil {
				newOutput = append(newOutput, newItem.(*GroupItem))
			}
		}
	}

	if changed {
		cpy := *n
		cpy.GroupBy = newGroupBy
		if newOutput != nil {
			cpy.Output = newOutput
		}
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *CollStatsStage) Walk(v Visitor) Node {
	return n.WalkStage(v)
}

// WalkStage implements the Stage interface.
func (n *CollStatsStage) WalkStage(v Visitor) Stage {
	return n
}

// Walk implements the Node interface.
func (n *CountStage) Walk(v Visitor) Node {
	return n.WalkStage(v)
}

// WalkStage implements the Stage interface.
func (n *CountStage) WalkStage(v Visitor) Stage {
	return n
}

// Walk implements the Node interface.
func (n *FacetStage) Walk(v Visitor) Node {
	return n.WalkStage(v)
}

// WalkStage implements the Stage interface.
func (n *FacetStage) WalkStage(v Visitor) Stage {
	changed := false
	var newItems []*FacetItem
	for i, item := range n.Items {
		newItem, itemChanged := visitNode(v, item)
		changed = changed || itemChanged

		if changed && newItems == nil {
			newItems = make([]*FacetItem, i, len(n.Items))
			copy(newItems, n.Items[:i])
		}

		if changed {
			newItems = append(newItems, newItem.(*FacetItem))
		}
	}

	if changed {
		cpy := *n
		cpy.Items = newItems
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *GroupStage) Walk(v Visitor) Node {
	return n.WalkStage(v)
}

// WalkStage implements the Stage interface.
func (n *GroupStage) WalkStage(v Visitor) Stage {
	newBy, changed := visitExpr(v, n.By)

	var newItems []*GroupItem
	for i, item := range n.Items {
		newItem, itemChanged := visitNode(v, item)
		changed = changed || itemChanged

		if changed && newItems == nil {
			newItems = make([]*GroupItem, i, len(n.Items))
			copy(newItems, n.Items[:i])
		}

		if changed {
			newItems = append(newItems, newItem.(*GroupItem))
		}
	}

	if changed {
		cpy := *n
		cpy.By = newBy
		cpy.Items = newItems
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *LimitStage) Walk(v Visitor) Node {
	return n.WalkStage(v)
}

// WalkStage implements the Stage interface.
func (n *LimitStage) WalkStage(v Visitor) Stage {
	return n
}

// Walk implements the Node interface.
func (n *LookupStage) Walk(v Visitor) Node {
	return n.WalkStage(v)
}

// WalkStage implements the Stage interface.
func (n *LookupStage) WalkStage(v Visitor) Stage {
	changed := false
	var newLet []*LookupLetItem
	if n.Let != nil {
		for i, item := range n.Let {
			newItem, itemChanged := visitNode(v, item)
			changed = changed || itemChanged

			if changed && newLet == nil {
				newLet = make([]*LookupLetItem, i, len(n.Let))
				copy(newLet, n.Let[:i])
			}

			if newLet != nil {
				newLet = append(newLet, newItem.(*LookupLetItem))
			}
		}
	}
	var newPipeline *Pipeline
	if n.Pipeline != nil {
		var pipelineChanged bool
		newPipeline, pipelineChanged = visitPipeline(v, n.Pipeline)
		changed = changed || pipelineChanged
	}
	if n.LocalField != nil {
		expr, localChanged := visitExpr(v, n.LocalField)
		if localChanged {
			n.LocalField = expr.(*FieldRef)
		}
	}
	if changed {
		cpy := *n
		if newLet != nil {
			cpy.Let = newLet
		}
		if newPipeline != nil {
			cpy.Pipeline = newPipeline
		}
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *MatchStage) Walk(v Visitor) Node {
	return n.WalkStage(v)
}

// WalkStage implements the Stage interface.
func (n *MatchStage) WalkStage(v Visitor) Stage {
	expr, changed := visitExpr(v, n.Expr)
	if changed {
		cpy := *n
		cpy.Expr = expr
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *ProjectStage) Walk(v Visitor) Node {
	return n.WalkStage(v)
}

// WalkStage implements the Stage interface.
func (n *ProjectStage) WalkStage(v Visitor) Stage {
	changed := false
	var newItems []ProjectItem
	for i, item := range n.Items {
		newItem, itemChanged := visitNode(v, item)
		changed = changed || itemChanged

		if changed && newItems == nil {
			newItems = make([]ProjectItem, i, len(n.Items))
			copy(newItems, n.Items[:i])
		}

		if changed {
			newItems = append(newItems, newItem.(ProjectItem))
		}
	}

	if changed {
		cpy := *n
		cpy.Items = newItems
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *RedactStage) Walk(v Visitor) Node {
	return n.WalkStage(v)
}

// WalkStage implements the Stage interface.
func (n *RedactStage) WalkStage(v Visitor) Stage {
	newExpr, changed := visitExpr(v, n.Expr)
	if changed {
		cpy := *n
		cpy.Expr = newExpr
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *ReplaceRootStage) Walk(v Visitor) Node {
	return n.WalkStage(v)
}

// WalkStage implements the Stage interface.
func (n *ReplaceRootStage) WalkStage(v Visitor) Stage {
	newRoot, changed := visitExpr(v, n.NewRoot)
	if changed {
		cpy := *n
		cpy.NewRoot = newRoot
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *SampleStage) Walk(v Visitor) Node {
	return n.WalkStage(v)
}

// WalkStage implements the Stage interface.
func (n *SampleStage) WalkStage(v Visitor) Stage {
	return n
}

// Walk implements the Node interface.
func (n *SkipStage) Walk(v Visitor) Node {
	return n.WalkStage(v)
}

// WalkStage implements the Stage interface.
func (n *SkipStage) WalkStage(v Visitor) Stage {
	return n
}

// Walk implements the Node interface.
func (n *SortStage) Walk(v Visitor) Node {
	return n.WalkStage(v)
}

// WalkStage implements the Stage interface.
func (n *SortStage) WalkStage(v Visitor) Stage {
	changed := false
	var newItems []*SortItem
	for i, item := range n.Items {
		newItemExpr, itemChanged := visitExpr(v, item.Expr)
		changed = changed || itemChanged

		if changed && newItems == nil {
			newItems = make([]*SortItem, i, len(n.Items))
			copy(newItems, n.Items[:i])
		}

		if changed {
			newItems = append(newItems, NewSortItem(
				newItemExpr.(*FieldRef),
				item.Descending,
			))
		}
	}

	if changed {
		cpy := *n
		cpy.Items = newItems
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *SortByCountStage) Walk(v Visitor) Node {
	return n.WalkStage(v)
}

// WalkStage implements the Stage interface.
func (n *SortByCountStage) WalkStage(v Visitor) Stage {
	expr, changed := visitExpr(v, n.Expr)
	if changed {
		cpy := *n
		cpy.Expr = expr
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *SortedMergeStage) Walk(v Visitor) Node {
	return n.WalkStage(v)
}

// WalkStage implements the Stage interface.
func (n *SortedMergeStage) WalkStage(v Visitor) Stage {
	changed := false
	var newItems []*SortItem
	for i, item := range n.Items {
		newItemExpr, itemChanged := visitExpr(v, item.Expr)
		changed = changed || itemChanged

		if changed && newItems == nil {
			newItems = make([]*SortItem, i, len(n.Items))
			copy(newItems, n.Items[:i])
		}

		if changed {
			newItems = append(newItems, NewSortItem(
				newItemExpr.(*FieldRef),
				item.Descending,
			))
		}
	}

	if changed {
		cpy := *n
		cpy.Items = newItems
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *UnwindStage) Walk(v Visitor) Node {
	return n.WalkStage(v)
}

// WalkStage implements the Stage interface.
func (n *UnwindStage) WalkStage(v Visitor) Stage {
	path, changed := visitExpr(v, n.Path)
	if changed {
		cpy := *n
		cpy.Path = path.(*FieldRef)
		return &cpy
	}
	return n
}

//---------------------------------
// Expressions

// Walk implements the Node interface.
func (n *AggExpr) Walk(v Visitor) Node {
	return n.WalkExpr(v)
}

// WalkExpr implements Expr interface.
func (n *AggExpr) WalkExpr(v Visitor) Expr {
	expr, changed := visitExpr(v, n.Expr)
	if changed {
		cpy := *n
		cpy.Expr = expr
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *Array) Walk(v Visitor) Node {
	return n.WalkExpr(v)
}

// WalkExpr implements Expr interface.
func (n *Array) WalkExpr(v Visitor) Expr {
	changed := false
	var newElements []Expr
	for i, e := range n.Elements {
		newElement, elemChanged := visitExpr(v, e)
		changed = changed || elemChanged

		if changed && newElements == nil {
			newElements = make([]Expr, i, len(n.Elements))
			copy(newElements, n.Elements[:i])
		}

		if changed {
			newElements = append(newElements, newElement)
		}
	}

	if changed {
		cpy := *n
		cpy.Elements = newElements
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *ArrayIndexRef) Walk(v Visitor) Node {
	return n.WalkRef(v)
}

// WalkExpr implements Expr interface.
func (n *ArrayIndexRef) WalkExpr(v Visitor) Expr {
	return n.WalkRef(v)
}

// WalkRef implements Ref interface.
func (n *ArrayIndexRef) WalkRef(v Visitor) Ref {
	index, indexChanged := visitExpr(v, n.Index)
	parent, parentChanged := visitExpr(v, n.Parent)
	if indexChanged || parentChanged {
		cpy := *n
		cpy.Index = index
		cpy.Parent = parent
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *Binary) Walk(v Visitor) Node {
	return n.WalkExpr(v)
}

// WalkExpr implements Expr interface.
func (n *Binary) WalkExpr(v Visitor) Expr {
	left, changedL := visitExpr(v, n.Left)
	right, changedR := visitExpr(v, n.Right)
	if changedL || changedR {
		cpy := *n
		cpy.Left = left
		cpy.Right = right
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *MatchRegex) Walk(v Visitor) Node {
	return n
}

// WalkExpr implements Expr interface.
func (n *MatchRegex) WalkExpr(v Visitor) Expr {
	return n
}

// Walk implements the Node interface.
func (n *Conditional) Walk(v Visitor) Node {
	return n.WalkExpr(v)
}

// WalkExpr implements Expr interface.
func (n *Conditional) WalkExpr(v Visitor) Expr {
	newIf, ifChanged := visitExpr(v, n.If)
	newThen, thenChanged := visitExpr(v, n.Then)
	newElse, elseChanged := visitExpr(v, n.Else)
	if ifChanged || thenChanged || elseChanged {
		cpy := *n
		cpy.If = newIf
		cpy.Then = newThen
		cpy.Else = newElse
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *Map) Walk(v Visitor) Node {
	return n.WalkExpr(v)
}

// WalkExpr implements Expr interface.
func (n *Map) WalkExpr(v Visitor) Expr {
	newInput, inputChanged := visitExpr(v, n.Input)
	newIn, inChanged := visitExpr(v, n.In)
	if inputChanged || inChanged {
		cpy := *n
		cpy.Input = newInput
		cpy.As = n.As
		cpy.In = newIn
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *Filter) Walk(v Visitor) Node {
	return n.WalkExpr(v)
}

// WalkExpr implements Expr interface.
func (n *Filter) WalkExpr(v Visitor) Expr {
	newInput, inputChanged := visitExpr(v, n.Input)
	newCond, condChanged := visitExpr(v, n.Cond)
	if inputChanged || condChanged {
		cpy := *n
		cpy.Input = newInput
		cpy.As = n.As
		cpy.Cond = newCond
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *Constant) Walk(v Visitor) Node {
	return n
}

// WalkExpr implements Expr interface.
func (n *Constant) WalkExpr(v Visitor) Expr {
	return n
}

// Walk implements the Node interface.
func (n *Document) Walk(v Visitor) Node {
	return n.WalkExpr(v)
}

// WalkExpr implements Expr interface.
func (n *Document) WalkExpr(v Visitor) Expr {
	changed := false
	var newElements []*DocumentElement
	for i, e := range n.Elements {
		newElementExpr, elemChanged := visitExpr(v, e.Expr)
		changed = changed || elemChanged

		if changed && newElements == nil {
			newElements = make([]*DocumentElement, i, len(n.Elements))
			copy(newElements, n.Elements[:i])
		}

		if changed {
			newElements = append(newElements, &DocumentElement{
				Name: e.Name,
				Expr: newElementExpr,
			})
		}
	}

	if changed {
		cpy := *n
		cpy.Elements = newElements
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *FieldOrArrayIndexRef) Walk(v Visitor) Node {
	return n.WalkRef(v)
}

// WalkExpr implements Expr interface.
func (n *FieldOrArrayIndexRef) WalkExpr(v Visitor) Expr {
	return n.WalkRef(v)
}

// WalkRef implements Ref interface.
func (n *FieldOrArrayIndexRef) WalkRef(v Visitor) Ref {
	parent, changed := visitExpr(v, n.Parent)
	if changed {
		cpy := *n
		cpy.Parent = parent
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *FieldRef) Walk(v Visitor) Node {
	return n.WalkRef(v)
}

// WalkExpr implements Expr interface.
func (n *FieldRef) WalkExpr(v Visitor) Expr {
	return n.WalkRef(v)
}

// WalkRef implements Ref interface.
func (n *FieldRef) WalkRef(v Visitor) Ref {
	if n.Parent != nil {
		parent, changed := visitExpr(v, n.Parent)
		if changed {
			cpy := *n
			cpy.Parent = parent
			return &cpy
		}
	}
	return n
}

// Walk implements the Node interface.
func (n *Function) Walk(v Visitor) Node {
	return n.WalkExpr(v)
}

// WalkExpr implements Expr interface.
func (n *Function) WalkExpr(v Visitor) Expr {
	arg, changed := visitExpr(v, n.Arg)
	if changed {
		cpy := *n
		cpy.Arg = arg
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *Let) Walk(v Visitor) Node {
	return n.WalkExpr(v)
}

// WalkExpr implements Expr interface.
func (n *Let) WalkExpr(v Visitor) Expr {
	changed := false
	var newVariables []*LetVariable
	for i, variable := range n.Variables {
		newVariable, variableChanged := visitNode(v, variable)
		changed = changed || variableChanged

		if changed && newVariables == nil {
			newVariables = make([]*LetVariable, i, len(n.Variables))
			copy(newVariables, n.Variables[:i])
		}

		if newVariables != nil {
			newVariables = append(newVariables, newVariable.(*LetVariable))
		}
	}
	newExpr, exprChanged := visitExpr(v, n.Expr)
	changed = changed || exprChanged
	if changed {
		cpy := *n
		if newVariables != nil {
			cpy.Variables = newVariables
		}
		if exprChanged {
			cpy.Expr = newExpr
		}
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *Unknown) Walk(v Visitor) Node {
	return n
}

// WalkExpr implements Expr interface.
func (n *Unknown) WalkExpr(v Visitor) Expr {
	return n
}

// WalkStage implements Stage interface.
func (n *Unknown) WalkStage(v Visitor) Stage {
	return n
}

// Walk implements the Node interface.
func (n *VariableRef) Walk(v Visitor) Node {
	return n.WalkRef(v)
}

// WalkExpr implements Expr interface.
func (n *VariableRef) WalkExpr(v Visitor) Expr {
	return n.WalkRef(v)
}

// WalkRef implements Ref interface.
func (n *VariableRef) WalkRef(v Visitor) Ref {
	return n
}

//---------------------------------
// Misc

// Walk implements the Node interface.
func (n *AssignProjectItem) Walk(v Visitor) Node {
	return n.WalkProjectItem(v)
}

// WalkProjectItem implements the ProjectItem interface.
func (n *AssignProjectItem) WalkProjectItem(v Visitor) ProjectItem {
	expr, changed := visitExpr(v, n.Expr)
	if changed {
		cpy := *n
		cpy.Expr = expr
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *ExcludeProjectItem) Walk(v Visitor) Node {
	return n.WalkProjectItem(v)
}

// WalkProjectItem implements the ProjectItem interface.
func (n *ExcludeProjectItem) WalkProjectItem(v Visitor) ProjectItem {
	expr, changed := visitExpr(v, n.FieldRef)
	if changed {
		cpy := *n
		cpy.FieldRef = expr.(*FieldRef)
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *GroupItem) Walk(v Visitor) Node {
	expr, changed := visitExpr(v, n.Expr)
	if changed {
		cpy := *n
		cpy.Expr = expr
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *IncludeProjectItem) Walk(v Visitor) Node {
	return n.WalkProjectItem(v)
}

// WalkProjectItem implements the ProjectItem interface.
func (n *IncludeProjectItem) WalkProjectItem(v Visitor) ProjectItem {
	expr, changed := visitExpr(v, n.FieldRef)
	if changed {
		cpy := *n
		cpy.FieldRef = expr.(*FieldRef)
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *LookupLetItem) Walk(v Visitor) Node {
	expr, changed := visitExpr(v, n.Expr)
	if changed {
		cpy := *n
		cpy.Expr = expr
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *AddFieldsItem) Walk(v Visitor) Node {
	expr, changed := visitExpr(v, n.Expr)
	if changed {
		cpy := *n
		cpy.Expr = expr
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *LetVariable) Walk(v Visitor) Node {
	expr, changed := visitExpr(v, n.Expr)
	if changed {
		cpy := *n
		cpy.Expr = expr
		return &cpy
	}
	return n
}

// Walk implements the Node interface.
func (n *FacetItem) Walk(v Visitor) Node {
	pipeline, changed := visitPipeline(v, n.Pipeline)
	if changed {
		cpy := *n
		cpy.Pipeline = pipeline
		return &cpy
	}
	return n
}
