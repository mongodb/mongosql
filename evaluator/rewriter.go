package evaluator

import (
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/parser"
)

// RewriterConfig is a container for all the values needed to run the syntax
// rewrite phase.
type RewriterConfig struct {
	lg                     log.Logger
	rewriteDistinctAsGroup bool
}

// NewRewriterConfig returns a new RewriterConfig constructed from the
// provided values. RewriterConfigs should always be constructed via this
// function instead of via a struct literal.
func NewRewriterConfig(lg log.Logger, rewriteDistinctAsGroup bool) *RewriterConfig {
	return &RewriterConfig{
		lg:                     lg,
		rewriteDistinctAsGroup: rewriteDistinctAsGroup,
	}
}

// RewriteQuery performs any syntactic rewrites. It will return an error,
// if one of the syntactic rewrites raises an error.
func RewriteQuery(cfg *RewriterConfig, stmt parser.Statement) (parser.Statement, error) {
	// If the user requests that we try to avoid distinct, rewrite the query
	// to avoid distinct.
	if cfg.rewriteDistinctAsGroup {
		return parser.RewriteDistinct(cfg.lg, stmt), nil
	}
	return stmt, nil
}
