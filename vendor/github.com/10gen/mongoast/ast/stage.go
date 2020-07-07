package ast

import (
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// Stage is implemented by all expressions in the AST.
//
// WalkStage applies the Visitor to all children of the stage.  It must not
// modify the receiver.  If any children are modified, it must return a copy of
// the receiver with modifications.  Generally, the Walk method of a Stage will
// just delegate to WalkStage.
type Stage interface {
	Node
	StageName() string
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

// StageName implements the Stage interface.
func (*AddFieldsStage) StageName() string {
	return "$addFields"
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

// StageName implements the Stage interface.
func (*BucketStage) StageName() string {
	return "$bucket"
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

// StageName implements the Stage interface.
func (*BucketAutoStage) StageName() string {
	return "$bucketAuto"
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

// StageName implements the StageName interface.
func (*CollStatsStage) StageName() string {
	return "$collStats"
}

// NewCountStage makes a CountStage.
func NewCountStage(fieldName string) *CountStage {
	return &CountStage{fieldName}
}

// CountStage counts the documents and sets the fieldName to that value.
type CountStage struct {
	FieldName string
}

// StageName implements the Stage interface.
func (*CountStage) StageName() string {
	return "$count"
}

// NewCurrentOpStage makes a CurrentOpStage.
func NewCurrentOpStage(allUsers, idleConnections, idleCursors, idleSessions, localOps, debug bool) *CurrentOpStage {
	return &CurrentOpStage{
		AllUsers:        allUsers,
		IdleConnections: idleConnections,
		IdleCursors:     idleCursors,
		IdleSessions:    idleSessions,
		LocalOps:        localOps,
		Debug:           debug,
	}
}

// CurrentOpStage returns a list of currently running queries.
type CurrentOpStage struct {
	AllUsers        bool
	IdleConnections bool
	IdleCursors     bool
	IdleSessions    bool
	LocalOps        bool
	Debug           bool
}

// StageName implements the Stage interface.
func (*CurrentOpStage) StageName() string {
	return "$currentOp"
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

// StageName implements the Stage interface.
func (*FacetStage) StageName() string {
	return "$facet"
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

// StageName implements the Stage interface.
func (*GroupStage) StageName() string {
	return "$group"
}

// NewIndexStatsStage makes an IndexStatsStage.
func NewIndexStatsStage() *IndexStatsStage {
	return &IndexStatsStage{}
}

// IndexStatsStage gets index statistics.
type IndexStatsStage struct{}

// StageName implements the Stage interface.
func (*IndexStatsStage) StageName() string {
	return "$indexStats"
}

// NewLimitStage makes a LimitStage.
func NewLimitStage(count int64) *LimitStage {
	return &LimitStage{count}
}

// LimitStage limits the overall output to the specified count.
type LimitStage struct {
	Count int64
}

// StageName implements the Stage interface.
func (*LimitStage) StageName() string {
	return "$limit"
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

func NewLookupStageWithDB(db, coll string, localField *FieldRef, foreignField, as string, let []*LookupLetItem, pipeline *Pipeline) *LookupStage {
	return &LookupStage{
		From:         coll,
		LocalField:   localField,
		ForeignField: foreignField,
		As:           as,
		Let:          let,
		Pipeline:     pipeline,
		FromDB:       db,
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
	FromDB       string
}

// StageName implements the Stage interface.
func (*LookupStage) StageName() string {
	return "$lookup"
}

// NewMatchStage makes a MatchStage.
func NewMatchStage(expr Expr) *MatchStage {
	return &MatchStage{Expr: expr}
}

// MatchStage is a filtering stage.
type MatchStage struct {
	Expr Expr
}

// StageName implements the Stage interface.
func (*MatchStage) StageName() string {
	return "$match"
}

// NewOutStage makes an OutStage.
func NewOutStage(collectionName string) *OutStage {
	return &OutStage{CollectionName: collectionName}
}

// OutToAtlasFields defines the fields for output to Atlas.
type OutToAtlasFields struct {
	ProjectID      string
	ClusterName    string
	DatabaseName   string
	CollectionName string
}

// OutToS3Fields defines the fields for output to S3.
type OutToS3Fields struct {
	Bucket           Expr
	Filename         Expr
	Region           string
	Format           string
	MaxFileSizeBytes int64
}

// OutStage is an output stage. This stage can take several different forms,
// each of which is represented by a different field on this structure. Exactly
// one field should be non-empty.
type OutStage struct {
	CollectionName string
	Atlas          *OutToAtlasFields
	S3             *OutToS3Fields
	S3URL          string
}

// StageName implements the Stage interface.
func (*OutStage) StageName() string {
	return "$out"
}

// NewOutToAtlasStage makes an OutStage for output to Atlas.
// { $out: { atlas: { projectID: "123abc", clusterName: "test", db: "foo", coll: "bar" } } }
func NewOutToAtlasStage(projectID, clusterName, databaseName, collectionName string) *OutStage {
	return &OutStage{
		Atlas: &OutToAtlasFields{
			ProjectID:      projectID,
			ClusterName:    clusterName,
			DatabaseName:   databaseName,
			CollectionName: collectionName,
		},
	}
}

// NewOutToS3Stage makes an OutStage for output to S3.
// { $out: { s3: { bucket: "foo", filename: "bar" } } }
func NewOutToS3Stage(bucket, filename Expr, region, format string, maxFileSizeBytes int64) *OutStage {
	return &OutStage{
		S3: &OutToS3Fields{
			Bucket:           bucket,
			Filename:         filename,
			Region:           region,
			Format:           format,
			MaxFileSizeBytes: maxFileSizeBytes,
		},
	}
}

// NewOutToS3URLStage makes an OutStage for output to an S3 URL.
// { $out: { s3: "s3://foo/bar" } }
func NewOutToS3URLStage(url string) *OutStage {
	return &OutStage{
		S3URL: url,
	}
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
	return GetDottedFieldName(pi.Ref)
}

// NewExcludeProjectItem makes an ExcludeProjectItem.
func NewExcludeProjectItem(ref FieldLikeRef) *ExcludeProjectItem {
	return &ExcludeProjectItem{
		Ref: ref,
	}
}

// ExcludeProjectItem is excluded from output.
type ExcludeProjectItem struct {
	Ref FieldLikeRef
}

// GetName returns the name of the item.
func (pi *ExcludeProjectItem) GetName() string {
	return GetDottedFieldName(pi.Ref)
}

// NewIncludeProjectItem makes an IncludeProjectItem.
func NewIncludeProjectItem(ref FieldLikeRef) *IncludeProjectItem {
	return &IncludeProjectItem{
		Ref: ref,
	}
}

// IncludeProjectItem is included in output.
type IncludeProjectItem struct {
	Ref FieldLikeRef
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

// StageName implements the Stage interface.
func (*ProjectStage) StageName() string {
	return "$project"
}

// AssignItems returns all the AssignProjectItems.
func (n *ProjectStage) AssignItems() []*AssignProjectItem {
	var result []*AssignProjectItem
	for _, i := range n.Items {
		if api, ok := i.(*AssignProjectItem); ok {
			result = append(result, api)
		}
	}

	return result
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
			m[GetDottedFieldName(epi.Ref)] = struct{}{}
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

// StageName implements the Stage interface.
func (*RedactStage) StageName() string {
	return "$redact"
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

// StageName implements the Stage interface.
func (*ReplaceRootStage) StageName() string {
	return "$replaceRoot"
}

// NewSampleStage makes a SampleStage.
func NewSampleStage(count int64) *SampleStage {
	return &SampleStage{count}
}

// SampleStage returns a specified number of random documents from a collection.
type SampleStage struct {
	Count int64
}

// StageName implements the Stage interface.
func (*SampleStage) StageName() string {
	return "$sample"
}

// NewSkipStage makes a SkipStage.
func NewSkipStage(count int64) *SkipStage {
	return &SkipStage{count}
}

// SkipStage skips the number of documents before returning results.
type SkipStage struct {
	Count int64
}

// StageName implements the Stage interface.
func (*SkipStage) StageName() string {
	return "$skip"
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

// SortingStage is an interface for stages that have a slice of SortItems.
type SortingStage interface {
	Stage
	SortItems() []*SortItem
}

// NewSortStage makes a sort stage.
func NewSortStage(items ...*SortItem) *SortStage {
	return &SortStage{items}
}

// SortStage is a sorting stage.
type SortStage struct {
	Items []*SortItem
}

// StageName implements the Stage interface.
func (*SortStage) StageName() string {
	return "$sort"
}

// SortItems implements the SortingStage interface.
func (s *SortStage) SortItems() []*SortItem {
	return s.Items
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

// StageName implements the Stage interface.
func (*SortByCountStage) StageName() string {
	return "$sortByCount"
}

// NewSortByExprStage makes a sort by expression stage.
func NewSortByExprStage(items ...*SortItem) *SortByExprStage {
	return &SortByExprStage{items}
}

// SortByExprStage is a sorting stage that allows sorting on any
// expression, not just field refs.
type SortByExprStage struct {
	Items []*SortItem
}

// StageName implements the Stage interface.
func (*SortByExprStage) StageName() string {
	return "$sortByExpr"
}

// SortItems implements the SortingStage interface.
func (s *SortByExprStage) SortItems() []*SortItem {
	return s.Items
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

// StageName implements the Stage interface.
func (*SortedMergeStage) StageName() string {
	return "$sortedMerge"
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

// StageName implements the Stage interface.
func (*UnwindStage) StageName() string {
	return "$unwind"
}

// NewUnionWithStage makes a UnionWithStage.
func NewUnionWithStage(coll string, pipeline *Pipeline) *UnionWithStage {
	return &UnionWithStage{
		coll,
		pipeline,
	}
}

// UnionWithStage unions two collections together.
type UnionWithStage struct {
	Coll     string
	Pipeline *Pipeline
}

// StageName implements the Stage interface.
func (*UnionWithStage) StageName() string {
	return "$unionWith"
}
