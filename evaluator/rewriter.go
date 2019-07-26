package evaluator

import (
	"github.com/10gen/sqlproxy/internal/option"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/parser"
)

// RewriterConfig is a container for all the values needed to run the syntax
// rewrite phase.
type RewriterConfig struct {
	connID                 uint64
	dbName                 string
	lg                     log.Logger
	rewriteDistinctAsGroup bool
	mySQLVersion           string
	remoteHost             string
	user                   string
}

// NewRewriterConfig returns a new RewriterConfig constructed from the
// provided values. RewriterConfigs should always be constructed via this
// function instead of via a struct literal.
func NewRewriterConfig(connID uint64, dbName string, lg log.Logger, rewriteDistinctAsGroup bool, mySQLVersion, remoteHost, user string) *RewriterConfig {
	return &RewriterConfig{
		connID:                 connID,
		dbName:                 dbName,
		lg:                     lg,
		rewriteDistinctAsGroup: rewriteDistinctAsGroup,
		mySQLVersion:           mySQLVersion,
		remoteHost:             remoteHost,
		user:                   user,
	}
}

// RewriteQuery performs any syntactic rewrites for queries. It will return an error,
// if one of the syntactic rewrites raises an error.
func RewriteQuery(cfg *RewriterConfig, stmt parser.Statement) (parser.Statement, error) {
	return rewriteCommon(cfg, stmt, option.NoneString())
}

// RewriteCommand performs any syntactic rewrites for commands. It will return an error,
// if one of the syntactic rewrites raises an error.
func RewriteCommand(cfg *RewriterConfig, stmt parser.Statement) (parser.Statement, error) {
	oldStmtString := parser.String(stmt)

	// Rewrite the command to remove syntactic sugar and simplify the resulting
	// command plan.
	stmt, err := parser.DesugarCommand(stmt)
	if err != nil {
		return nil, err
	}

	// Now apply rewrites that are specific to queries, but not commands (since commands may
	// contain queries).
	return rewriteCommon(cfg, stmt, option.SomeString(oldStmtString))
}

// rewriteCommon performs any syntactic rewrites shared by queries and commands. It will return an error,
// if one of the syntactic rewrites raises an error.
func rewriteCommon(cfg *RewriterConfig, stmt parser.Statement, originalStmt option.String) (parser.Statement, error) {
	oldStmtString := originalStmt.Else(parser.String(stmt))

	// Add explicit aliases for columns without them, so that the column headers
	// don't get changed by any of the following rewrites.
	stmt = parser.NameColumns(stmt)

	// Rewrite the query to remove syntactic sugar and simplify the resulting
	// query plan.
	stmt, err := parser.DesugarQuery(stmt)
	if err != nil {
		return nil, err
	}

	// If the user requests that we try to avoid distinct, rewrite the query
	// to avoid distinct.
	if cfg.rewriteDistinctAsGroup {
		stmt = parser.RewriteDistinct(stmt)
	}

	stmt, err = parser.RewriteConstantScalarFunctions(stmt, cfg.connID, cfg.dbName, cfg.mySQLVersion, cfg.remoteHost, cfg.user)
	if err != nil {
		return nil, err
	}

	// We check the queries for string equality to tell if a rewrite occurs. The
	// reason for this is the rewrites might happen under the root, as is the
	// case of unions.
	newStmtString := parser.String(stmt)
	if oldStmtString != newStmtString {
		cfg.lg.Debugf(log.Admin, "rewritten query: `%s`", newStmtString)
	}

	return stmt, nil
}
