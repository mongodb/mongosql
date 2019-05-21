package optimizer

import (
	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/parser"
)

// NormalizePipeline puts expressions in stages that may differ syntactically,
// but have the same semantics, into the same syntactic form via parsing and then
// deparsing the pipeline in question.
func NormalizePipeline(pipeline *ast.Pipeline) *ast.Pipeline {
	pipelineV := parser.DeparsePipeline(pipeline)
	out, err := parser.ParsePipeline(pipelineV.Array())
	if err != nil {
		panic(err)
	}
	return out
}
