package analyzer

import (
	"fmt"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/stringutil"
)

// DefinedFieldsUnique returns the root of all the fields defined by
// this stage. _id will always be the first defined field of
// $group, $bucket, and $bucketAuto.
func DefinedFieldsUnique(stage ast.Stage) ([]string, bool) {
	definedFields, complete := definedFieldsHelper(stage)
	definedFieldRoots := make([]string, 0)
	set := stringutil.NewStringSet()
	for _, s := range definedFields {
		root := GetPathRootString(s)
		if !set.Contains(root) {
			definedFieldRoots = append(definedFieldRoots, root)
			set.Add(root)
		}
	}
	return definedFieldRoots, complete
}

// DefinedFields returns the root of all the fields defined by
// this stage in order, not removing duplicates. _id will always be the first defined field of
// $group, $bucket, and $bucketAuto. In addition, a boolean is returned indicating whether
// or not the list of defined fields is complete.
func DefinedFields(stage ast.Stage) ([]string, bool) {
	definedFields, complete := definedFieldsHelper(stage)
	definedFieldRoots := make([]string, 0, len(definedFields))
	for _, s := range definedFields {
		root := GetPathRootString(s)
		definedFieldRoots = append(definedFieldRoots, root)
	}
	return definedFieldRoots, complete
}

// DefinedFieldsFullPath returns the entire path of all the fields
// defined by this stage. In addition, a boolean is returned indicating whether
// or not the list of defined fields is complete.
func DefinedFieldsFullPath(stage ast.Stage) ([]string, bool) {
	return definedFieldsHelper(stage)
}

func definedFieldsHelper(stage ast.Stage) ([]string, bool) {
	var ret []string
	var complete bool
	switch typedStage := stage.(type) {
	case *ast.AddFieldsStage:
		ret = make([]string, len(typedStage.Items))
		for i, item := range typedStage.Items {
			ret[i] = item.Name
		}
		complete = true
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
		complete = true
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
		complete = true
	case *ast.CollStatsStage:
		// This is an overly large set of fields, potentially, but defining more is not a problem
		// from optimization standpoint, as CollStats kills any previous fields (and documents!)
		// anyway.
		ret = []string{"ns", "shard", "host", "localTime", "latencyStats", "storageStats", "count"}
		complete = true
	case *ast.CountStage:
		// MongoDB does not allow the $count field to have a '.' in it, so no worries here.
		ret = []string{typedStage.FieldName}
		complete = true
	case *ast.FacetStage:
		ret = make([]string, len(typedStage.Items))
		for i, item := range typedStage.Items {
			ret[i] = item.Name
		}
		complete = true
	case *ast.GroupStage:
		ret = make([]string, len(typedStage.Items)+1)
		ret[0] = "_id"
		for i, item := range typedStage.Items {
			ret[i+1] = item.Name
		}
		complete = true
	case *ast.LookupStage:
		ret = []string{typedStage.As}
		complete = true
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
		complete = true
	case *ast.ReplaceRootStage:
		switch typedRoot := typedStage.NewRoot.(type) {
		case *ast.FieldRef,
			*ast.VariableRef,
			*ast.ArrayIndexRef,
			*ast.FieldOrArrayIndexRef:
			// In this case we don't know what top level fields are defined.
			ret = []string{}
			complete = false
		case *ast.Document:
			ret = make([]string, len(typedRoot.Elements))
			for i, item := range typedRoot.Elements {
				ret[i] = item.Name
			}
			complete = true
		default:
			panic("$replaceRoot must have a document as its argument")
		}
	case *ast.UnionWithStage:
		ret = []string{}
		complete = false
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
		complete = true
	}
	return ret, complete
}
