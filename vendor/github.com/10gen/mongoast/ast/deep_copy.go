package ast

import (
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// DeepCopier is implemented by every Node.
type DeepCopier interface {
	DeepCopy() DeepCopier
}

//---------------------------------
// Pipeline

// DeepCopy implements the DeepCopier interface.
func (n *Pipeline) DeepCopy() DeepCopier {
	var newStages []Stage

	if n.Stages != nil {
		newStages = make([]Stage, len(n.Stages))
		for i, stage := range n.Stages {
			if stage == nil {
				newStages[i] = nil
			} else {
				newStages[i] = stage.DeepCopy().(Stage)
			}
		}
	}

	return NewPipeline(newStages...)
}

//---------------------------------
// Stages

// DeepCopy implements the DeepCopier interface.
func (n *AddFieldsStage) DeepCopy() DeepCopier {
	var newItems []*AddFieldsItem

	if n.Items != nil {
		newItems = make([]*AddFieldsItem, len(n.Items))
		for i, item := range n.Items {
			if item == nil {
				newItems[i] = nil
			} else {
				newItems[i] = item.DeepCopy().(*AddFieldsItem)
			}
		}
	}

	return NewAddFieldsStage(newItems...)
}

// DeepCopy implements the DeepCopier interface.
func (n *BucketStage) DeepCopy() DeepCopier {
	var newGroupBy Expr
	var newBoundaries []bsoncore.Value
	var newDefault *bsoncore.Value
	var newOutput []*GroupItem

	if n.GroupBy != nil {
		newGroupBy = n.GroupBy.DeepCopy().(Expr)
	}

	if n.Boundaries != nil {
		newBoundaries = make([]bsoncore.Value, len(n.Boundaries))
		copy(newBoundaries, n.Boundaries)
	}

	if n.Default != nil {
		defaultValue := *n.Default
		newDefault = &defaultValue
	}

	if n.Output != nil {
		newOutput = make([]*GroupItem, len(n.Output))
		for i, item := range n.Output {
			if item != nil {
				newOutput[i] = item.DeepCopy().(*GroupItem)
			} else {
				newOutput[i] = nil
			}
		}
	}

	return NewBucketStage(newGroupBy, newBoundaries, newDefault, newOutput)
}

// DeepCopy implements the DeepCopier interface.
func (n *BucketAutoStage) DeepCopy() DeepCopier {
	var newGroupBy Expr
	var newOutput []*GroupItem

	if n.GroupBy != nil {
		newGroupBy = n.GroupBy.DeepCopy().(Expr)
	}

	if n.Output != nil {
		newOutput = make([]*GroupItem, len(n.Output))
		for i, item := range n.Output {
			if item != nil {
				newOutput[i] = item.DeepCopy().(*GroupItem)
			} else {
				newOutput[i] = nil
			}
		}
	}

	return NewBucketAutoStage(newGroupBy, n.Buckets, newOutput, n.Granularity)
}

// DeepCopy implements the DeepCopier interface.
func (n *CollStatsStage) DeepCopy() DeepCopier {
	var newLatencyStats *CollStatsLatencyStats
	var newStorageStats *CollStatsStorageStats
	var newCount *CollStatsCount

	if n.LatencyStats != nil {
		newLatencyStats = NewCollStatsLatencyStats(n.LatencyStats.Histograms)
	}

	if n.StorageStats != nil {
		newStorageStats = NewCollStatsStorageStats()
	}

	if n.Count != nil {
		newCount = NewCollStatsCount()
	}

	return NewCollStatsStage(newLatencyStats, newStorageStats, newCount)
}

// DeepCopy implements the DeepCopier interface.
func (n *CountStage) DeepCopy() DeepCopier {
	return NewCountStage(n.FieldName)
}

// DeepCopy implements the DeepCopier interface.
func (n *FacetStage) DeepCopy() DeepCopier {
	var newItems []*FacetItem

	if n.Items != nil {
		newItems = make([]*FacetItem, len(n.Items))
		for i, item := range n.Items {
			if item == nil {
				newItems[i] = nil
			} else {
				newItems[i] = item.DeepCopy().(*FacetItem)
			}
		}
	}

	return NewFacetStage(newItems...)
}

// DeepCopy implements the DeepCopier interface.
func (n *GroupStage) DeepCopy() DeepCopier {
	var newBy Expr
	var newItems []*GroupItem

	if n.By != nil {
		newBy = n.By.DeepCopy().(Expr)
	}

	if n.Items != nil {
		newItems = make([]*GroupItem, len(n.Items))
		for i, item := range n.Items {
			if item == nil {
				newItems[i] = nil
			} else {
				newItems[i] = item.DeepCopy().(*GroupItem)
			}
		}
	}

	return NewGroupStage(newBy, newItems...)
}

// DeepCopy implements the DeepCopier interface.
func (n *LimitStage) DeepCopy() DeepCopier {
	return NewLimitStage(n.Count)
}

// DeepCopy implements the DeepCopier interface.
func (n *LookupStage) DeepCopy() DeepCopier {
	var newPipeline *Pipeline
	var newLet []*LookupLetItem

	if n.Pipeline != nil {
		newPipeline = n.Pipeline.DeepCopy().(*Pipeline)
	}

	if n.Let != nil {
		newLet = make([]*LookupLetItem, len(n.Let))
		for i, item := range n.Let {
			if item == nil {
				newLet[i] = nil
			} else {
				newLet[i] = item.DeepCopy().(*LookupLetItem)
			}
		}
	}

	return NewLookupStage(n.From, n.LocalField, n.ForeignField, n.As, newLet, newPipeline)
}

// DeepCopy implements the DeepCopier interface.
func (n *MatchStage) DeepCopy() DeepCopier {
	var newExpr Expr

	if n.Expr != nil {
		newExpr = n.Expr.DeepCopy().(Expr)
	}

	return NewMatchStage(newExpr)
}

// DeepCopy implements the DeepCopier interface.
func (n *ProjectStage) DeepCopy() DeepCopier {
	var newItems []ProjectItem

	if n.Items != nil {
		newItems = make([]ProjectItem, len(n.Items))
		for i, item := range n.Items {
			if item == nil {
				newItems[i] = nil
			} else {
				newItems[i] = item.DeepCopy().(ProjectItem)
			}
		}
	}

	return NewProjectStage(newItems...)
}

// DeepCopy implements the DeepCopier interface.
func (n *RedactStage) DeepCopy() DeepCopier {
	var newExpr Expr

	if n.Expr != nil {
		newExpr = n.Expr.DeepCopy().(Expr)
	}

	return NewRedactStage(newExpr)
}

// DeepCopy implements the DeepCopier interface.
func (n *ReplaceRootStage) DeepCopy() DeepCopier {
	var newNewRoot Expr

	if n.NewRoot != nil {
		newNewRoot = n.NewRoot.DeepCopy().(Expr)
	}

	return NewReplaceRootStage(newNewRoot)
}

// DeepCopy implements the DeepCopier interface.
func (n *SampleStage) DeepCopy() DeepCopier {
	return NewSampleStage(n.Count)
}

// DeepCopy implements the DeepCopier interface.
func (n *SkipStage) DeepCopy() DeepCopier {
	return NewSkipStage(n.Count)
}

// DeepCopy implements the DeepCopier interface.
func (n *SortStage) DeepCopy() DeepCopier {
	var newItems []*SortItem

	if n.Items != nil {
		newItems = make([]*SortItem, len(n.Items))
		for i, item := range n.Items {
			if item == nil {
				newItems[i] = nil
			} else {
				newItems[i] = item.DeepCopy().(*SortItem)
			}
		}
	}

	return NewSortStage(newItems...)
}

// DeepCopy implements the DeepCopier interface.
func (n *SortByCountStage) DeepCopy() DeepCopier {
	var newExpr Expr

	if n.Expr != nil {
		newExpr = n.Expr.DeepCopy().(Expr)
	}

	return NewSortByCountStage(newExpr)
}

// DeepCopy implements the DeepCopier interface.
func (n *SortedMergeStage) DeepCopy() DeepCopier {
	var newItems []*SortItem

	if n.Items != nil {
		newItems = make([]*SortItem, len(n.Items))
		for i, item := range n.Items {
			if item == nil {
				newItems[i] = nil
			} else {
				newItems[i] = item.DeepCopy().(*SortItem)
			}
		}
	}

	return NewSortedMergeStage(newItems...)
}

// DeepCopy implements the DeepCopier interface.
func (n *UnwindStage) DeepCopy() DeepCopier {
	var newPath *FieldRef

	if n.Path != nil {
		newPath = n.Path.DeepCopy().(*FieldRef)
	}

	return NewUnwindStage(newPath, n.IncludeArrayIndex, n.PreserveNullAndEmptyArrays)
}

//---------------------------------
// Expressions

// DeepCopy implements the DeepCopier interface.
func (n *AggExpr) DeepCopy() DeepCopier {
	var newExpr Expr

	if n.Expr != nil {
		newExpr = n.Expr.DeepCopy().(Expr)
	}

	return NewAggExpr(newExpr)
}

// DeepCopy implements the DeepCopier interface.
func (n *Array) DeepCopy() DeepCopier {
	var newElements []Expr

	if n.Elements != nil {
		newElements = make([]Expr, len(n.Elements))
		for i, e := range n.Elements {
			if e == nil {
				newElements[i] = nil
			} else {
				newElements[i] = e.DeepCopy().(Expr)
			}
		}
	}

	return NewArray(newElements...)
}

// DeepCopy implements the DeepCopier interface.
func (n *ArrayIndexRef) DeepCopy() DeepCopier {
	var newIndex Expr
	var newParent Expr

	if n.Index != nil {
		newIndex = n.Index.DeepCopy().(Expr)
	}

	if n.Parent != nil {
		newParent = n.Parent.DeepCopy().(Expr)
	}

	return NewArrayIndexRef(newIndex, newParent)
}

// DeepCopy implements the DeepCopier interface.
func (n *Binary) DeepCopy() DeepCopier {
	var newLeft Expr
	var newRight Expr

	if n.Left != nil {
		newLeft = n.Left.DeepCopy().(Expr)
	}

	if n.Right != nil {
		newRight = n.Right.DeepCopy().(Expr)
	}

	return NewBinary(n.Op, newLeft, newRight)
}

// DeepCopy implements the DeepCopier interface.
func (n *Document) DeepCopy() DeepCopier {
	var newElements []*DocumentElement

	if n.Elements != nil {
		newElements = make([]*DocumentElement, len(n.Elements))
		for i, e := range n.Elements {
			if e == nil {
				newElements[i] = nil
			} else {
				var newExpr Expr
				if e.Expr != nil {
					newExpr = e.Expr.DeepCopy().(Expr)
				}

				newElements[i] = NewDocumentElement(e.Name, newExpr)
			}
		}
	}

	return NewDocument(newElements...)
}

// DeepCopy implements the DeepCopier interface.
func (n *Constant) DeepCopy() DeepCopier {
	return NewConstant(n.Value)
}

// DeepCopy implements the DeepCopier interface.
func (n *FieldOrArrayIndexRef) DeepCopy() DeepCopier {
	var newParent Expr

	if n.Parent != nil {
		newParent = n.Parent.DeepCopy().(Expr)
	}

	return NewFieldOrArrayIndexRef(n.Number, newParent)
}

// DeepCopy implements the DeepCopier interface.
func (n *FieldRef) DeepCopy() DeepCopier {
	var newParent Expr

	if n.Parent != nil {
		newParent = n.Parent.DeepCopy().(Expr)
	}

	return NewFieldRef(n.Name, newParent)
}

// DeepCopy implements the DeepCopier interface.
func (n *Function) DeepCopy() DeepCopier {
	var newArg Expr

	if n.Arg != nil {
		newArg = n.Arg.DeepCopy().(Expr)
	}

	return NewFunction(n.Name, newArg)
}

// DeepCopy implements the DeepCopier interface.
func (n *Let) DeepCopy() DeepCopier {
	var newExpr Expr
	var newVariables []*LetVariable

	if n.Expr != nil {
		newExpr = n.Expr.DeepCopy().(Expr)
	}

	if n.Variables != nil {
		newVariables = make([]*LetVariable, len(n.Variables))
		for i, v := range n.Variables {
			if v == nil {
				newVariables[i] = nil
			} else {
				newVariables[i] = v.DeepCopy().(*LetVariable)
			}
		}
	}

	return NewLet(newVariables, newExpr)
}

// DeepCopy implements the DeepCopier interface.
func (n *Conditional) DeepCopy() DeepCopier {
	var newIf Expr
	var newThen Expr
	var newElse Expr

	if n.If != nil {
		newIf = n.If.DeepCopy().(Expr)
	}

	if n.Then != nil {
		newThen = n.Then.DeepCopy().(Expr)
	}

	if n.Else != nil {
		newElse = n.Else.DeepCopy().(Expr)
	}

	return NewConditional(newIf, newThen, newElse)
}

// DeepCopy implements the DeepCopier interface.
func (n *Unknown) DeepCopy() DeepCopier {
	return NewUnknown(n.Value)
}

// DeepCopy implements the DeepCopier interface.
func (n *VariableRef) DeepCopy() DeepCopier {
	return NewVariableRef(n.Name)
}

//---------------------------------
// Misc

// DeepCopy implements the DeepCopier interface.
func (n *ExcludeProjectItem) DeepCopy() DeepCopier {
	var newFieldRef *FieldRef

	if n.FieldRef != nil {
		newFieldRef = n.FieldRef.DeepCopy().(*FieldRef)
	}

	return NewExcludeProjectItem(newFieldRef)
}

// DeepCopy implements the DeepCopier interface.
func (n *GroupItem) DeepCopy() DeepCopier {
	var newExpr Expr

	if n.Expr != nil {
		newExpr = n.Expr.DeepCopy().(Expr)
	}

	return NewGroupItem(n.Name, newExpr)
}

// DeepCopy implements the DeepCopier interface.
func (n *IncludeProjectItem) DeepCopy() DeepCopier {
	var newFieldRef *FieldRef

	if n.FieldRef != nil {
		newFieldRef = n.FieldRef.DeepCopy().(*FieldRef)
	}

	return NewIncludeProjectItem(newFieldRef)
}

// DeepCopy implements the DeepCopier interface.
func (n *AssignProjectItem) DeepCopy() DeepCopier {
	var newExpr Expr

	if n.Expr != nil {
		newExpr = n.Expr.DeepCopy().(Expr)
	}

	return NewAssignProjectItem(n.Name, newExpr)
}

// DeepCopy implements the DeepCopier interface.
func (n *LookupLetItem) DeepCopy() DeepCopier {
	var newExpr Expr

	if n.Expr != nil {
		newExpr = n.Expr.DeepCopy().(Expr)
	}

	return NewLookupLetItem(n.Name, newExpr)
}

// DeepCopy implements the DeepCopier interface.
func (n *AddFieldsItem) DeepCopy() DeepCopier {
	var newExpr Expr

	if n.Expr != nil {
		newExpr = n.Expr.DeepCopy().(Expr)
	}

	return NewAddFieldsItem(n.Name, newExpr)
}

// DeepCopy implements the DeepCopier interface.
func (n *SortItem) DeepCopy() DeepCopier {
	var newExpr Expr

	if n.Expr != nil {
		newExpr = n.Expr.DeepCopy().(Expr)
	}

	return NewSortItem(newExpr, n.Descending)
}

// DeepCopy implements the DeepCopier interface.
func (n *LetVariable) DeepCopy() DeepCopier {
	var newExpr Expr

	if n.Expr != nil {
		newExpr = n.Expr.DeepCopy().(Expr)
	}

	return NewLetVariable(n.Name, newExpr)
}

// DeepCopy implements the DeepCopier interface.
func (n *FacetItem) DeepCopy() DeepCopier {
	var newPipeline *Pipeline

	if n.Pipeline != nil {
		newPipeline = n.Pipeline.DeepCopy().(*Pipeline)
	}

	return NewFacetItem(n.Name, newPipeline)
}
