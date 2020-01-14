package analyzer

import (
	"fmt"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/stringutil"
)

// DefinedFieldsUnique returns the root of all the fields defined by
// this stage. _id will always be the first defined field of
// $group, $bucket, and $bucketAuto.
func DefinedFieldsUnique(stage ast.Stage) []string {
	definedFields := definedFieldsHelper(stage)
	definedFieldRoots := make([]string, 0)
	set := stringutil.NewStringSet()
	for _, s := range definedFields {
		root := GetPathRootString(s)
		if !set.Contains(root) {
			definedFieldRoots = append(definedFieldRoots, root)
			set.Add(root)
		}
	}
	return definedFieldRoots
}

// DefinedFields returns the root of all the fields defined by
// this stage in order, not removing duplicates. _id will always be the first defined field of
// $group, $bucket, and $bucketAuto.
func DefinedFields(stage ast.Stage) []string {
	definedFields := definedFieldsHelper(stage)
	definedFieldRoots := make([]string, 0, len(definedFields))
	for _, s := range definedFields {
		root := GetPathRootString(s)
		definedFieldRoots = append(definedFieldRoots, root)
	}
	return definedFieldRoots
}

// DefinedFieldsFullPath returns the entire path of all the fields
// defined by this stage.
func DefinedFieldsFullPath(stage ast.Stage) []string {
	return definedFieldsHelper(stage)
}

func definedFieldsHelper(stage ast.Stage) []string {
	var ret []string
	switch typedStage := stage.(type) {
	case *ast.AddFieldsStage:
		ret = make([]string, len(typedStage.Items))
		for i, item := range typedStage.Items {
			ret[i] = item.Name
		}
	case *ast.BucketStage:
		if len(typedStage.Output) == 0 {
			ret = []string{"_id", "count"}
		} else {
			ret = make([]string, len(typedStage.Output)+1)
			ret[0] = "_id"
			for i, item := range typedStage.Output {
				ret[i+1] = item.Name
			}
		}
	case *ast.BucketAutoStage:
		if len(typedStage.Output) == 0 {
			ret = make([]string, 2)
			ret[0], ret[1] = "_id", "count"
		} else {
			ret = make([]string, len(typedStage.Output)+1)
			ret[0] = "_id"
			for i, item := range typedStage.Output {
				ret[i+1] = item.Name
			}
		}
	case *ast.CollStatsStage:
		// This is an overly large set of fields, potentially, but defining more is not a problem
		// from optimization standpoint, as CollStats kills any previous fields (and documents!)
		// anyway.
		ret = []string{"ns", "shard", "host", "localTime", "latencyStats", "storageStats", "count"}
	case *ast.CountStage:
		// MongoDB does not allow the $count field to have a '.' in it, so no worries here.
		ret = []string{typedStage.FieldName}
	case *ast.FacetStage:
		ret = make([]string, len(typedStage.Items))
		for i, item := range typedStage.Items {
			ret[i] = item.Name
		}
	case *ast.GroupStage:
		ret = make([]string, len(typedStage.Items)+1)
		ret[0] = "_id"
		for i, item := range typedStage.Items {
			ret[i+1] = item.Name
		}
	case *ast.LookupStage:
		ret = []string{typedStage.As}
	case *ast.ProjectStage:
		for _, item := range typedStage.Items {
			switch typedItem := item.(type) {
			case *ast.IncludeProjectItem:
				field := ast.GetDottedFieldName(typedItem.Ref)
				ret = append(ret, field)
			case *ast.AssignProjectItem:
				ret = append(ret, typedItem.Name)
			}
		}
	case *ast.ReplaceRootStage:
		switch typedRoot := typedStage.NewRoot.(type) {
		case *ast.FieldRef,
			*ast.VariableRef,
			*ast.ArrayIndexRef,
			*ast.FieldOrArrayIndexRef:
			// In this case we don't know what top level fields are defined.
			ret = []string{}
		case *ast.Document:
			ret = make([]string, len(typedRoot.Elements))
			for i, item := range typedRoot.Elements {
				ret[i] = item.Name
			}
		default:
			panic("$replaceRoot must have a document as its argument")
		}
	case *ast.UnwindStage:
		path, ok := typedStage.Path.(ast.Ref)
		if !ok {
			panic(fmt.Sprintf("$unwind stage has path that is not a reference, got %T", typedStage.Path))
		}
		name := ast.GetDottedFieldName(path)
		if typedStage.IncludeArrayIndex != "" {
			ret = []string{name, typedStage.IncludeArrayIndex}
		} else {
			ret = []string{name}
		}
	}
	return ret
}
