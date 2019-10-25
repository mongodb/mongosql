package ast

import (
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// Stage is implemented by all expressions in the AST.
type Stage interface {
	Node
	WalkStage(v Visitor) Stage
}

// NewAddFieldsStage makes an AddFieldsStage.
func NewAddFieldsStage(items ...*AddFieldsItem) *AddFieldsStage {
	return &AddFieldsStage{
		Items: items,
	}
}

// NewAddFieldsItem makes an AddFieldsItem.
func NewAddFieldsItem(name string, expr Expr) *AddFieldsItem {
	return &AddFieldsItem{
		Name: name,
		Expr: expr,
	}
}

// AddFieldsItem specifies a field to be added.
type AddFieldsItem struct {
	Name string
	Expr Expr
}

// AddFieldsStage adds fields to the documents in a collection.
type AddFieldsStage struct {
	Items []*AddFieldsItem
}

// NewBucketStage makes a BucketStage.
func NewBucketStage(groupBy Expr, boundaries []bsoncore.Value, defaultID *bsoncore.Value, output []*GroupItem) *BucketStage {
	return &BucketStage{
		GroupBy:    groupBy,
		Boundaries: boundaries,
		Default:    defaultID,
		Output:     output,
	}
}

// BucketStage divides the documents in a collection into buckets based on
// the specified boundaries.
type BucketStage struct {
	GroupBy    Expr
	Boundaries []bsoncore.Value
	Default    *bsoncore.Value
	Output     []*GroupItem
}

// NewBucketAutoStage makes a BucketAutoStage.
func NewBucketAutoStage(groupBy Expr, buckets int64, output []*GroupItem, granularity string) *BucketAutoStage {
	return &BucketAutoStage{
		GroupBy:     groupBy,
		Buckets:     buckets,
		Output:      output,
		Granularity: granularity,
	}
}

// BucketAutoStage divides the documents in a collections into buckets based on
// automatically calculated boundaries.
type BucketAutoStage struct {
	GroupBy     Expr
	Buckets     int64
	Output      []*GroupItem
	Granularity string
}

// NewCollStatsStage makes a CollStatsStage.
func NewCollStatsStage(latencyStats *CollStatsLatencyStats, storageStats *CollStatsStorageStats, count *CollStatsCount) *CollStatsStage {
	return &CollStatsStage{
		LatencyStats: latencyStats,
		StorageStats: storageStats,
		Count:        count,
	}
}

// NewCollStatsLatencyStats makes a CollStatsLatencyStats.
func NewCollStatsLatencyStats(histograms bool) *CollStatsLatencyStats {
	return &CollStatsLatencyStats{
		Histograms: histograms,
	}
}

// NewCollStatsStorageStats makes a CollStatsStorageStats.
func NewCollStatsStorageStats() *CollStatsStorageStats {
	return &CollStatsStorageStats{}
}

// NewCollStatsCount makes a CollStatsCount.
func NewCollStatsCount() *CollStatsCount {
	return &CollStatsCount{}
}

// CollStatsLatencyStats requests latency statistics.
type CollStatsLatencyStats struct {
	Histograms bool
}

// CollStatsStorageStats requests storage statistics.
type CollStatsStorageStats struct{}

// CollStatsCount requests the document count.
type CollStatsCount struct{}

// CollStatsStage requests collection statistics.
type CollStatsStage struct {
	LatencyStats *CollStatsLatencyStats
	StorageStats *CollStatsStorageStats
	Count        *CollStatsCount
}

// NewCountStage makes a CountStage.
func NewCountStage(fieldName string) *CountStage {
	return &CountStage{fieldName}
}

// CountStage counts the documents and sets the fieldName to that value.
type CountStage struct {
	FieldName string
}

// NewFacetStage makes a FacetStage.
func NewFacetStage(items ...*FacetItem) *FacetStage {
	return &FacetStage{
		Items: items,
	}
}

// NewFacetItem makes a FacetItem.
func NewFacetItem(name string, pipeline *Pipeline) *FacetItem {
	return &FacetItem{
		Name:     name,
		Pipeline: pipeline,
	}
}

// FacetItem is an item for a FacetStage.
type FacetItem struct {
	Name     string
	Pipeline *Pipeline
}

// FacetStage runs multiple aggregation pipelines.
type FacetStage struct {
	Items []*FacetItem
}

// NewGroupStage makes a GroupStage.
func NewGroupStage(by Expr, items ...*GroupItem) *GroupStage {
	return &GroupStage{by, items}
}

// NewGroupItem makes a GroupItem.
func NewGroupItem(name string, expr Expr) *GroupItem {
	return &GroupItem{name, expr}
}

// GroupItem is an item for a GroupStage.
type GroupItem struct {
	Name string
	Expr Expr
}

// GroupStage groups documents together.
type GroupStage struct {
	By    Expr
	Items []*GroupItem
}

// NewIndexStatsStage makes an IndexStatsStage.
func NewIndexStatsStage() *IndexStatsStage {
	return &IndexStatsStage{}
}

// IndexStatsStage gets index statistics.
type IndexStatsStage struct{}

// NewLimitStage makes a LimitStage.
func NewLimitStage(count int64) *LimitStage {
	return &LimitStage{count}
}

// LimitStage limits the overall output to the specified count.
type LimitStage struct {
	Count int64
}

// NewLookupStage makes a LookupStage.
func NewLookupStage(from string, localField *FieldRef, foreignField, as string, let []*LookupLetItem, pipeline *Pipeline) *LookupStage {
	return &LookupStage{
		From:         from,
		LocalField:   localField,
		ForeignField: foreignField,
		As:           as,
		Let:          let,
		Pipeline:     pipeline,
	}
}

// NewLookupLetItem makes a LookupLetItem
func NewLookupLetItem(name string, expr Expr) *LookupLetItem {
	return &LookupLetItem{
		Name: name,
		Expr: expr,
	}
}

// LookupLetItem is an item for the let field of a LookupStage.
type LookupLetItem struct {
	Name string
	Expr Expr
}

// LookupStage pulls in documents from a different collection in the same database.
type LookupStage struct {
	From         string
	LocalField   *FieldRef
	ForeignField string
	As           string
	Let          []*LookupLetItem
	Pipeline     *Pipeline
}

// NewMatchStage makes a MatchStage.
func NewMatchStage(expr Expr) *MatchStage {
	return &MatchStage{Expr: expr}
}

// MatchStage is a filtering stage.
type MatchStage struct {
	Expr Expr
}

// NewOutStage makes an OutStage.
func NewOutStage(collectionName string) *OutStage {
	return &OutStage{CollectionName: collectionName}
}

// OutStage is an output stage.
type OutStage struct {
	CollectionName string
}

// ProjectItem is an item in a ProjectStage.
type ProjectItem interface {
	Node
	GetName() string

	WalkProjectItem(v Visitor) ProjectItem
}

// NewAssignProjectItem makes an AssignProjectItem.
func NewAssignProjectItem(name string, expr Expr) *AssignProjectItem {
	return &AssignProjectItem{
		Name: name,
		Expr: expr,
	}
}

// AssignProjectItem assigns a field to the value of an expression.
type AssignProjectItem struct {
	Name string
	Expr Expr
}

// GetName returns the name of the item.
func (pi *AssignProjectItem) GetName() string {
	return pi.Name
}

// GetName returns the name of the item.
func (pi *IncludeProjectItem) GetName() string {
	return GetDottedFieldName(pi.FieldRef)
}

// NewExcludeProjectItem makes an ExcludeProjectItem.
func NewExcludeProjectItem(fieldRef *FieldRef) *ExcludeProjectItem {
	return &ExcludeProjectItem{
		FieldRef: fieldRef,
	}
}

// ExcludeProjectItem is excluded from output.
type ExcludeProjectItem struct {
	FieldRef *FieldRef
}

// GetName returns the name of the item.
func (pi *ExcludeProjectItem) GetName() string {
	return GetDottedFieldName(pi.FieldRef)
}

// NewIncludeProjectItem makes an IncludeProjectItem.
func NewIncludeProjectItem(fieldRef *FieldRef) *IncludeProjectItem {
	return &IncludeProjectItem{
		FieldRef: fieldRef,
	}
}

// IncludeProjectItem is included in output.
type IncludeProjectItem struct {
	FieldRef *FieldRef
}

// NewProjectStage makes a ProjectStage.
func NewProjectStage(items ...ProjectItem) *ProjectStage {
	return &ProjectStage{
		Items: items,
	}
}

// ProjectStage is a projection Stage.
type ProjectStage struct {
	Items []ProjectItem
}

// IsExclusion returns true if this project is an exclusion.
func (n *ProjectStage) IsExclusion() bool {
	for _, i := range n.Items {
		if _, ok := i.(*ExcludeProjectItem); !ok && (len(n.Items) == 1 || i.GetName() != "_id") {
			return false
		}
	}
	return true
}

// IsInclusion returns true if this project is an inclusion.
func (n *ProjectStage) IsInclusion() bool {
	return !n.IsExclusion()
}

// ExcludeItems returns all the items that should be excluded.
func (n *ProjectStage) ExcludeItems() map[string]struct{} {
	m := make(map[string]struct{})
	for _, i := range n.Items {
		if epi, ok := i.(*ExcludeProjectItem); ok {
			m[GetDottedFieldName(epi.FieldRef)] = struct{}{}
		}
	}

	return m
}

// NonExcludeItems returns all the items that are not exclude items.
func (n *ProjectStage) NonExcludeItems() []ProjectItem {
	var result []ProjectItem
	for _, i := range n.Items {
		switch i.(type) {
		case *ExcludeProjectItem:
		default:
			result = append(result, i)
		}
	}
	return result
}

// IncludeItems returns all the items that should be included.
func (n *ProjectStage) IncludeItems() []*IncludeProjectItem {
	var result []*IncludeProjectItem
	for _, i := range n.Items {
		if ipi, ok := i.(*IncludeProjectItem); ok {
			result = append(result, ipi)
		}
	}

	return result
}

// NewRedactStage makes a RedactStage.
func NewRedactStage(expr Expr) *RedactStage {
	return &RedactStage{
		Expr: expr,
	}
}

// RedactStage redacts data from documents.
type RedactStage struct {
	Expr Expr
}

// NewReplaceRootStage makes a ReplaceRootStage.
func NewReplaceRootStage(newRoot Expr) *ReplaceRootStage {
	return &ReplaceRootStage{
		NewRoot: newRoot,
	}
}

// ReplaceRootStage replaces the entire document with the value of the
// specified expression.
type ReplaceRootStage struct {
	NewRoot Expr
}

// NewSampleStage makes a SampleStage.
func NewSampleStage(count int64) *SampleStage {
	return &SampleStage{count}
}

// SampleStage returns a specified number of random documents from a collection.
type SampleStage struct {
	Count int64
}

// NewSkipStage makes a SkipStage.
func NewSkipStage(count int64) *SkipStage {
	return &SkipStage{count}
}

// SkipStage skips the number of documents before returning results.
type SkipStage struct {
	Count int64
}

// NewSortItem makes a sort item.
func NewSortItem(expr Expr, descending bool) *SortItem {
	return &SortItem{expr, descending}
}

// SortItem is an item to sort by.
type SortItem struct {
	Expr       Expr
	Descending bool
}

// NewSortStage makes a sort stage.
func NewSortStage(items ...*SortItem) *SortStage {
	return &SortStage{items}
}

// SortStage is a sorting stage.
type SortStage struct {
	Items []*SortItem
}

// NewSortByCountStage makes a sort by count stage.
func NewSortByCountStage(expr Expr) *SortByCountStage {
	return &SortByCountStage{
		Expr: expr,
	}
}

// SortByCountStage groups by the specified expression and then sorts by count.
type SortByCountStage struct {
	Expr Expr
}

// NewSortedMergeStage makes a sorted merge stage.
func NewSortedMergeStage(items ...*SortItem) *SortedMergeStage {
	return &SortedMergeStage{items}
}

// SortedMergeStage is a stage which merges incoming sorted streams
// into a single sorted stream.
type SortedMergeStage struct {
	Items []*SortItem
}

// NewUnwindStage makes an unwind stage.
func NewUnwindStage(field Ref, arrayIndexField string, preserveNullAndEmptyArrays bool) *UnwindStage {
	return &UnwindStage{field, arrayIndexField, preserveNullAndEmptyArrays}
}

// UnwindStage is a stage that unwinds a particular field.
type UnwindStage struct {
	Path                       Ref
	IncludeArrayIndex          string
	PreserveNullAndEmptyArrays bool
}
