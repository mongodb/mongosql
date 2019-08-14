package evaluator

import (
	"github.com/10gen/sqlproxy/internal/versionutil"
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

// RewriteStatement performs any syntactic rewrites for all statements. It will return an error,
// if one of the syntactic rewrites raises an error.
func RewriteStatement(cfg *RewriterConfig, stmt parser.Statement) (parser.Statement, error) {
	oldStmtString := parser.String(stmt)

	// Rewrite the command to remove syntactic sugar and simplify the resulting
	// command plan.
	versionCode, err := versionutil.GetMySQLFixedWidthVersionCode(cfg.mySQLVersion)
	if err != nil {
		return nil, err
	}

	stmt, err = parser.DesugarStatement(stmt, versionCode)
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
