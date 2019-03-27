package astutil

import (
	"github.com/10gen/mongoast/ast"
	"github.com/10gen/sqlproxy/internal/mathutil"
)

// UnwindInfo contains the relevant info from an $unwind operation.
type UnwindInfo struct {
	// Position in the original pipeline.
	StageNumber int
	// Path of the $unwind.
	Path string
	// Index name.
	Index string
}

func (in *UnwindInfo) getPath() string {
	return in.Path
}

func (in *UnwindInfo) getIndex() string {
	return in.Index
}

// GetPipelineUnwindFields get all the unwind fields for a pipeline, in order
func GetPipelineUnwindFields(stages []ast.Stage) []UnwindInfo {
	unwinds := make([]UnwindInfo, 0, len(stages))
	for i, stage := range stages {
		if unwind, isUnwind := stage.(*ast.UnwindStage); isUnwind {
			path := FieldRefString(unwind.Path)
			index := unwind.IncludeArrayIndex
			unwinds = append(unwinds, UnwindInfo{StageNumber: i, Path: path, Index: index})
		}
	}

	return unwinds
}

// FindUnwindForPath finds an unwind in an []UnwindInfo that has the proper
// unwind path
func FindUnwindForPath(unwinds []UnwindInfo, path string) (UnwindInfo, bool) {
	for _, unwind := range unwinds {
		if unwind.Path == path {
			return unwind, true
		}
	}
	return UnwindInfo{StageNumber: -1, Path: "", Index: ""}, false
}

func getFields(in []UnwindInfo, m func(v *UnwindInfo) string) []string {
	ret := make([]string, len(in))
	for i, v := range in {
		ret[i] = m(&v)
	}
	return ret
}

// GetPaths gets the paths from a slice of UnwindInfo
// as a slice of strings.
func GetPaths(in []UnwindInfo) []string {
	return getFields(in, (*UnwindInfo).getPath)
}

// GetIndexes gets the index name from a slice of UnwindInfo
// as a slice of strings.
func GetIndexes(in []UnwindInfo) []string {
	return getFields(in, (*UnwindInfo).getIndex)
}

// GetUnwindSuffix will give the remaining unwinds for two slices of unwinds
// after matching on unwind path.
func GetUnwindSuffix(unwinds1, unwinds2 []UnwindInfo) ([]UnwindInfo, bool) {
	ret := make([]UnwindInfo, 0)
	end := mathutil.MinInt(len(unwinds1), len(unwinds2))
	i := 0
	for ; i < end; i++ {
		// Prefixes are incompatible, so there is no suffix
		// don't check index, assume that is correct
		if unwinds1[i].Path != unwinds2[i].Path {
			return nil, false
		}
	}
	var tail []UnwindInfo
	if len(unwinds1) <= len(unwinds2) {
		tail = unwinds2
	} else {
		tail = unwinds1

	}
	for ; i < len(tail); i++ {
		ret = append(ret, tail[i])
	}
	return ret, true
}
