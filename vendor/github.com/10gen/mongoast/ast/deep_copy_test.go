package ast_test

import (
	"reflect"
	"testing"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/bsonutil"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func TestDeepCopy(t *testing.T) {
	testCases := []struct {
		name string
		src  ast.DeepCopier
	}{
		{
			"Pipeline",
			ast.NewPipeline(
				ast.NewMatchStage(ast.NewBinary(
					"$or",
					ast.NewDocument(
						ast.NewDocumentElement("x", ast.NewConstant(bsonutil.Int64(1))),
					),
					ast.NewDocument(
						ast.NewDocumentElement("x", ast.NewConstant(bsonutil.Int64(2))),
					),
				)),
				ast.NewSkipStage(3),
				ast.NewProjectStage(
					ast.NewExcludeProjectItem(ast.NewFieldRef("_id", nil)),
					ast.NewAssignProjectItem("a", ast.NewBinary(
						"$add",
						ast.NewFieldRef("x", nil),
						ast.NewFieldRef("x", nil),
					)),
				),
				ast.NewLimitStage(5),
			),
		},
		{"Empty Pipeline", ast.NewPipeline()},
		{
			"Pipeline with nils",
			ast.NewPipeline(
				ast.NewProjectStage(
					ast.NewExcludeProjectItem(ast.NewFieldRef("_id", nil)),
					ast.NewAssignProjectItem("a", ast.NewBinary(
						"$add",
						ast.NewFieldRef("x", nil),
						ast.NewFieldRef("x", nil),
					)),
				),
				nil,
				ast.NewSkipStage(3),
				nil,
			),
		},
		{
			"AddFieldsStage",
			ast.NewAddFieldsStage(
				ast.NewAddFieldsItem("x", ast.NewFieldRef("a", nil)),
				ast.NewAddFieldsItem("y", ast.NewFieldRef("b", nil)),
			),
		},
		{"Empty AddFieldsStage", ast.NewAddFieldsStage()},
		{
			"AddFieldsStage with nils",
			ast.NewAddFieldsStage(
				nil,
				ast.NewAddFieldsItem("x", ast.NewFieldRef("a", nil)),
				nil,
				ast.NewAddFieldsItem("y", ast.NewFieldRef("b", nil)),
			),
		},
		{
			"BucketStage",
			ast.NewBucketStage(
				ast.NewFieldRef("x", nil),
				[]bsoncore.Value{
					bsonutil.Int32(0),
					bsonutil.Int32(10),
					bsonutil.Int32(20),
				},
				bsonutil.ValuePtr(bsonutil.Int32(-1)),
				[]*ast.GroupItem{
					ast.NewGroupItem(
						"a", ast.NewFunction(
							"$sum", ast.NewFieldRef("a", nil),
						),
					),
					ast.NewGroupItem(
						"b", ast.NewFunction(
							"$sum", ast.NewFieldRef("b", nil),
						),
					),
				},
			),
		},
		{
			"BucketStage with nils",
			ast.NewBucketStage(
				ast.NewFieldRef("x", nil),
				[]bsoncore.Value{
					bsonutil.Int32(0),
					bsonutil.Int32(10),
					bsonutil.Int32(20),
				},
				nil,
				nil,
			),
		},
		{
			"BucketStage with nils",
			ast.NewBucketStage(
				nil,
				nil,
				nil,
				[]*ast.GroupItem{
					nil,
					nil,
				},
			),
		},
		{
			"BucketAutoStage",
			ast.NewBucketAutoStage(
				ast.NewFieldRef("x", nil),
				2,
				[]*ast.GroupItem{
					ast.NewGroupItem(
						"a", ast.NewFunction(
							"$sum", ast.NewFieldRef("a", nil),
						),
					),
					ast.NewGroupItem(
						"b", ast.NewFunction(
							"$sum", ast.NewFieldRef("b", nil),
						),
					),
				},
				"R5",
			),
		},
		{
			"BucketAutoStage with nils",
			ast.NewBucketAutoStage(
				ast.NewFieldRef("x", nil),
				2,
				nil,
				"",
			),
		},
		{
			"BucketAutoStage with nils",
			ast.NewBucketAutoStage(
				ast.NewFieldRef("x", nil),
				0,
				[]*ast.GroupItem{
					nil,
					nil,
				},
				"",
			),
		},
		{
			"CollStatsStage",
			ast.NewCollStatsStage(
				ast.NewCollStatsLatencyStats(true),
				ast.NewCollStatsStorageStats(),
				ast.NewCollStatsCount(),
			),
		},
		{"CollStatsStage with nils", ast.NewCollStatsStage(nil, nil, nil)},
		{"CountStage", ast.NewCountStage("x")},
		{
			"FacetStage",
			ast.NewFacetStage(
				ast.NewFacetItem(
					"a1", ast.NewPipeline(
						ast.NewMatchStage(
							ast.NewBinary(
								ast.Equals,
								ast.NewFieldRef("a", nil),
								ast.NewConstant(bsonutil.Int32(1)),
							),
						),
					),
				),
				ast.NewFacetItem(
					"a2", ast.NewPipeline(
						ast.NewMatchStage(
							ast.NewBinary(
								ast.Equals,
								ast.NewFieldRef("a", nil),
								ast.NewConstant(bsonutil.Int32(2)),
							),
						),
					),
				),
			),
		},
		{
			"FacetStage with nils",
			ast.NewFacetStage(
				ast.NewFacetItem("x", nil),
			),
		},
		{
			"GroupStage",
			ast.NewGroupStage(
				ast.NewFieldRef("x", nil),
				ast.NewGroupItem("a", ast.NewVariableRef("b")),
				ast.NewGroupItem("c", ast.NewVariableRef("d")),
			),
		},
		{"Empty GroupStage", ast.NewGroupStage(nil)},
		{
			"GroupStage with nils",
			ast.NewGroupStage(
				ast.NewFieldRef("x", nil),
				nil,
				ast.NewGroupItem("c", ast.NewVariableRef("d")),
			),
		},
		{"LimitStage", ast.NewLimitStage(1)},
		{
			"LookupStage",
			ast.NewLookupStage(
				"foo", nil, "", "x",
				[]*ast.LookupLetItem{
					ast.NewLookupLetItem("y", ast.NewFieldRef("a", nil)),
				},
				ast.NewPipeline(
					ast.NewSampleStage(5),
				),
			),
		},
		{"LookupStage with nils", ast.NewLookupStage("foo", ast.NewFieldRef("a", nil), "b", "x", nil, nil)},
		{
			"MatchStage",
			ast.NewMatchStage(ast.NewFieldRef("x", nil)),
		},
		{"MatchStage with nil", ast.NewMatchStage(nil)},
		{
			"ProjectStage",
			ast.NewProjectStage(
				ast.NewAssignProjectItem(
					"x",
					ast.NewBinary(
						"$and",
						ast.NewFieldRef("a", nil),
						ast.NewFieldRef("b", nil),
					),
				),
				ast.NewIncludeProjectItem(ast.NewFieldRef("y", nil)),
				ast.NewExcludeProjectItem(ast.NewFieldRef("z", nil)),
			),
		},
		{"Empty ProjectStage", ast.NewProjectStage()},
		{
			"ProjectStage with nils",
			ast.NewProjectStage(
				ast.NewAssignProjectItem(
					"x",
					ast.NewBinary(
						"$and",
						ast.NewFieldRef("a", nil),
						ast.NewFieldRef("b", nil),
					),
				),
				nil,
				ast.NewExcludeProjectItem(ast.NewFieldRef("z", nil)),
			),
		},
		{
			"RedactStage",
			ast.NewRedactStage(
				ast.NewConditional(
					ast.NewBinary(
						ast.Equals,
						ast.NewFieldRef("a", nil),
						ast.NewConstant(bsonutil.Int32(5)),
					),
					ast.NewVariableRef("DESCEND"),
					ast.NewVariableRef("PRUNE"),
				),
			),
		},
		{
			"RedactStage with nil",
			ast.NewRedactStage(nil),
		},
		{
			"ReplaceRootStage",
			ast.NewReplaceRootStage(
				ast.NewDocument(
					ast.NewDocumentElement(
						"a", ast.NewFieldRef("foo", nil),
					),
				),
			),
		},
		{"ReplaceRootStage with nil", ast.NewReplaceRootStage(nil)},
		{"SampleStage", ast.NewSampleStage(1)},
		{"SkipStage", ast.NewSkipStage(1)},
		{
			"SortStage",
			ast.NewSortStage(
				ast.NewSortItem(ast.NewVariableRef("x"), true),
				ast.NewSortItem(ast.NewFieldRef("y", nil), true),
			),
		},
		{"Empty SortStage", ast.NewSortStage()},
		{
			"SortStage with nils",
			ast.NewSortStage(
				nil,
				ast.NewSortItem(ast.NewFieldRef("y", nil), true),
			),
		},
		{
			"SortByCountStage",
			ast.NewSortByCountStage(
				ast.NewFieldRef("x", nil),
			),
		},
		{"SortByCountStage with nil", ast.NewSortByCountStage(nil)},
		{
			"SortedMergeStage",
			ast.NewSortedMergeStage(
				ast.NewSortItem(ast.NewVariableRef("x"), true),
				ast.NewSortItem(ast.NewFieldRef("y", nil), true),
			),
		},
		{"Empty SortedMergeStage", ast.NewSortedMergeStage()},
		{
			"SortedMergeStage with nils",
			ast.NewSortedMergeStage(
				ast.NewSortItem(ast.NewVariableRef("x"), true),
				nil,
			),
		},
		{
			"UnwindStage",
			ast.NewUnwindStage(
				ast.NewFieldRef("x", nil),
				"y", true,
			),
		},
		{"UnwindStage with nil", ast.NewUnwindStage(nil, "x", false)},
		{"AggExpr", ast.NewAggExpr(ast.NewFunction("$sum", ast.NewFieldRef("x", nil)))},
		{"AggExpr with nil", ast.NewAggExpr(nil)},
		{
			"Array",
			ast.NewArray(
				ast.NewVariableRef("x"),
				ast.NewArray(
					ast.NewVariableRef("y"),
					ast.NewVariableRef("z"),
				),
			),
		},
		{"Empty Array", ast.NewArray()},
		{
			"Array with nils",
			ast.NewArray(
				ast.NewVariableRef("x"),
				nil,
				ast.NewArray(
					nil,
					ast.NewVariableRef("y"),
					ast.NewVariableRef("z"),
				),
				nil,
			),
		},
		{
			"ArrayIndexRef",
			ast.NewArrayIndexRef(
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewVariableRef("x"),
			),
		},
		{"ArrayIndexRef with nil", ast.NewArrayIndexRef(nil, nil)},
		{"Binary", ast.NewBinary("$and", ast.NewVariableRef("x"), ast.NewVariableRef("y"))},
		{"Binary with nils", ast.NewBinary("$and", nil, nil)},
		{
			"Document",
			ast.NewDocument(
				ast.NewDocumentElement("x", ast.NewVariableRef("y")),
				ast.NewDocumentElement("a", ast.NewVariableRef("b")),
			),
		},
		{"Empty Document", ast.NewDocument()},
		{
			"Document with nils",
			ast.NewDocument(
				ast.NewDocumentElement("x", ast.NewVariableRef("y")),
				nil,
				ast.NewDocumentElement("a", ast.NewVariableRef("b")),
				nil,
			),
		},
		{"Constant", ast.NewConstant(bsonutil.String("x"))},
		{
			"FieldOrArrayIndexRef",
			ast.NewFieldOrArrayIndexRef(1,
				ast.NewFieldOrArrayIndexRef(2, nil),
			),
		},
		{"FieldRef", ast.NewFieldRef("x", ast.NewFieldRef("y", nil))},
		{"Function", ast.NewFunction("$sum", ast.NewFieldRef("x", nil))},
		{"Function with nil", ast.NewFunction("$sum", nil)},
		{
			"Let",
			ast.NewLet([]*ast.LetVariable{
				ast.NewLetVariable("a", ast.NewConstant(bsonutil.Int64(0))),
				ast.NewLetVariable("b", ast.NewConstant(bsonutil.Int64(1))),
			},
				ast.NewBinary("$add", ast.NewVariableRef("a"), ast.NewVariableRef("b")),
			),
		},
		{"Empty Let", ast.NewLet(nil, nil)},
		{
			"Let with nils",
			ast.NewLet([]*ast.LetVariable{
				ast.NewLetVariable("a", ast.NewConstant(bsonutil.Int64(0))),
				nil,
			},
				nil,
			),
		},
		{
			"Conditional",
			ast.NewConditional(
				ast.NewBinary(
					ast.Equals,
					ast.NewFieldRef("a", nil),
					ast.NewConstant(bsonutil.Int32(5)),
				),
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(0)),
			),
		},
		{"Empty Conditional", ast.NewConditional(nil, nil, nil)},
		{"Unknown", ast.NewUnknown(bsonutil.String("x"))},
		{"VariableRef", ast.NewVariableRef("x")},
		{"ExcludeProjectItem", ast.NewExcludeProjectItem(ast.NewFieldRef("x", nil))},
		{"ExcludeProjectItem with nil", ast.NewExcludeProjectItem(nil)},
		{"GroupItem", ast.NewGroupItem("x", ast.NewVariableRef("y"))},
		{"GroupItem with nil", ast.NewGroupItem("x", nil)},
		{"IncludeProjectItem", ast.NewIncludeProjectItem(ast.NewFieldRef("x", nil))},
		{"IncludeProjectItem with nil", ast.NewIncludeProjectItem(nil)},
		{"AssignProjectItem", ast.NewAssignProjectItem("x", ast.NewVariableRef("y"))},
		{"AssignProjectItem with nil", ast.NewAssignProjectItem("x", nil)},
		{"LookupLetItem", ast.NewLookupLetItem("x", ast.NewVariableRef("y"))},
		{"LookupLetItem with nil", ast.NewLookupLetItem("x", nil)},
		{"AddFieldsItem", ast.NewAddFieldsItem("x", ast.NewVariableRef("y"))},
		{"AddFieldsItem with nil", ast.NewAddFieldsItem("x", nil)},
		{"SortItem", ast.NewSortItem(ast.NewFieldRef("x", nil), true)},
		{"SortItem with nil", ast.NewSortItem(nil, false)},
		{"LetVariable", ast.NewLetVariable("x", ast.NewConstant(bsonutil.String("y")))},
		{"LetVariable with nil", ast.NewLetVariable("x", nil)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			copy := tc.src.DeepCopy()

			if !isDeepCopy(tc.src, copy) {
				t.Fatal("not a deep copy")
			}
		})
	}
}

func isDeepCopy(n1, n2 ast.DeepCopier) bool {
	v1 := reflect.ValueOf(n1)
	v2 := reflect.ValueOf(n2)

	// False if they are not the same Kind or Type.
	if v1.Kind() != v2.Kind() || v1.Type() != v2.Type() {
		return false
	}

	// True if both are nil, false is just one is nil.
	if n1 == nil || v1.IsNil() {
		return n2 == nil || v2.IsNil()
	}

	if n2 == nil || v2.IsNil() {
		return false
	}

	// False if they point to the same value.
	if n1 == n2 {
		return false
	}

	if v1.Kind() == reflect.Ptr {
		v1 = v1.Elem()
		v2 = v2.Elem()
	}

	for i := 0; i < v1.NumField(); i++ {
		f1 := v1.Field(i)
		f2 := v2.Field(i)

		// False if their corresponding fields are not the same Kind or Type.
		if f1.Kind() != f2.Kind() || f1.Type() != f2.Type() {
			return false
		}

		switch f1.Kind() {
		case reflect.Interface, reflect.Ptr:
			// False if they are equal pointers.
			if f1 == f2 {
				return false
			}

			n1, f1IsNode := f1.Interface().(ast.Node)
			n2, f2IsNode := f2.Interface().(ast.Node)

			if f1IsNode && f2IsNode {
				// False if they are both ast.Nodes and not deep copies of each other.
				if !isDeepCopy(n1, n2) {
					return false
				}

				continue
			}

			// False if the values are not equal.
			if !reflect.DeepEqual(f1.Interface(), f2.Interface()) {
				return false
			}

		case reflect.Array, reflect.Slice:
			l1 := f1.Len()
			l2 := f2.Len()

			// False if the slices have different lengths
			if l1 != l2 {
				return false
			}

			for i := 0; i < l1; i++ {
				e1 := f1.Index(i)
				e2 := f2.Index(i)

				// False if the corresponding slice elements are not the same Kind or Type.
				if e1.Kind() != e2.Kind() || e1.Type() != e2.Type() {
					return false
				}

				n1, f1IsNode := e1.Interface().(ast.Node)
				n2, f2IsNode := e2.Interface().(ast.Node)

				if f1IsNode && f2IsNode {
					// False if they are both ast.Nodes and not deep copies of each other.
					if !isDeepCopy(n1, n2) {
						return false
					}

					continue
				}

				// False if the values are not equal.
				if !reflect.DeepEqual(f1.Interface(), f2.Interface()) {
					return false
				}
			}

		default:
			// False if the values are not equal.
			if !reflect.DeepEqual(f1.Interface(), f2.Interface()) {
				return false
			}
		}
	}

	return true
}
