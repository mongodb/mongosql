package desugarer

import (
	"context"

	"github.com/10gen/mongoast/optimizer"
	"github.com/10gen/mongoast/parser"
)

// Desugar desugars any new, unsupported, nonexistent MQL features in the
// pipeline into existing MQL features. This function does not modify the
// input; it returns a new pipeline with the desugared features.
func Desugar(pipelineBytes []byte, db string) ([]byte, error) {
	pipeline, err := parser.ParsePipeline(pipelineBytes)
	if err != nil {
		return nil, err
	}

	optimized := optimizer.RunPasses(context.Background(), pipeline, 0,
		desugarJoins,
		desugarAccumulators,
		desugarSubqueryExprs,
		desugarMatch,
		desugarUnsupportedOperators,
		desugarSQLNullSemantics,
		getSimplifyLookups(db),
	)

	deparsed, err := parser.DeparsePipelineErr(optimized)
	if err != nil {
		return nil, err
	}

	return deparsed.Array(), nil
}
