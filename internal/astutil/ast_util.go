package astutil

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/astprint"
)

// re1 and re2 hold regular expressions that will match special characters
// surrounding strings in the pipeline compatible with the mongo shell.
var (
	re1 = regexp.MustCompile(`"!!!`)
	re2 = regexp.MustCompile(`!!!"`)
)

// AllParentsAreFieldRefs returns true if every parent along the
// FieldRef path is also a FieldRef.
func AllParentsAreFieldRefs(fr *ast.FieldRef) bool {
	if fr == nil {
		return false
	}

	for fr.Parent != nil {
		// if the parent is an ast.FieldRef, continue by checking
		// it's parent.
		if parent, ok := fr.Parent.(*ast.FieldRef); ok {
			fr = parent
			continue
		}

		// if it is not an ast.FieldRef, return false
		return false
	}

	return true
}

// DeepCopyPipeline performs a deep copy of the given ast.Pipeline.
func DeepCopyPipeline(src *ast.Pipeline) *ast.Pipeline {
	if src == nil {
		// This makes testing easier: in some places we have nil pipelines
		// whereas in others we have empty pipelines. They are semantically
		// equivalent, so we merge here.
		return ast.NewPipeline()
	}

	return src.DeepCopy().(*ast.Pipeline)
}

// FieldRefString returns the string representation of the full path
// for the provided ast.FieldRef. It returns an unquoted string with
// no preceding "$" (or "$$" in the case that the parent was an
// ast.VariableRef).
func FieldRefString(fr *ast.FieldRef) string {
	if fr == nil {
		return ""
	}

	fullPath := ast.GetDottedFieldName(fr)

	if fullPath != "" && fullPath[0] == '$' {
		return fullPath[1:]
	}

	return fullPath
}

// GetRefName returns the name of the ast.Ref if it is an ast.FieldRef
// or ast.VariableRef (since those are the only ast.Refs that have names).
func GetRefName(ref ast.Ref) (string, bool) {
	switch t := ref.(type) {
	case *ast.FieldRef:
		return FieldRefString(t), true
	case *ast.VariableRef:
		return t.Name, true
	default:
		return "", false
	}
}

// InsertPipelineStageAt will insert a pipeline stage (ast.Stage) at a given place
// in a []ast.Stage, copying the tail out so that no stages are lost.
func InsertPipelineStageAt(pipeline []ast.Stage, stage ast.Stage, i int) []ast.Stage {
	return append(pipeline[:i], append([]ast.Stage{stage}, pipeline[i:]...)...)
}

// PipelineJSON takes an ast.Pipeline and marshals it into a json byte array.
func PipelineJSON(pipeline *ast.Pipeline, depth int, newline bool) ([]byte, error) {
	buf := bytes.Buffer{}
	n := len(pipeline.Stages)

	for i, s := range pipeline.Stages {
		PrintTabs(&buf, depth)
		buf.WriteString(astprint.String(s))
		if i != n-1 {
			if newline {
				buf.WriteString(",\n")
			} else {
				buf.WriteString(",")
			}
		}
	}

	bts := buf.Bytes()

	// remove special characters and quotation marks from pipeline string
	bts = re1.ReplaceAll(bts, []byte{})
	bts = re2.ReplaceAll(bts, []byte{})

	return bts, nil
}

// PipelineString returns a byte array with the supplied ast.Pipeline in string form.
func PipelineString(pipeline *ast.Pipeline, depth int) []byte {
	buf := bytes.Buffer{}
	for i, stage := range pipeline.Stages {
		PrintTabs(&buf, depth)
		buf.WriteString(fmt.Sprintf("  stage %v: '%v'\n", i+1, astprint.String(stage)))
	}
	return buf.Bytes()
}

// PrintTabs writes the specified number of tabs into the supplied byte buffer.
func PrintTabs(b *bytes.Buffer, d int) {
	for i := 0; i < d; i++ {
		b.WriteString("\t")
	}
}
