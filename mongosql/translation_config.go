package mongosql

import (
	"context"
	"errors"
	"math"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"

	"go.mongodb.org/mongo-driver/bson"
)

// NewQueryConfigFromTranslationConfig returns a new QueryConfig that uses the
// provided TranslationConfig's values where relevant and default values for
// all other config settings.
func NewQueryConfigFromTranslationConfig(tCfg *TranslationConfig) *evaluator.QueryConfig {
	lgr := log.GlobalLogger()

	rCfg := evaluator.NewRewriterConfig(0, tCfg.defaultDbName, lgr, false, "5.7.12", "localhost", "user")
	aCfg := evaluator.NewAlgebrizerConfig(lgr, tCfg.defaultDbName, tCfg.ctlg,
		variable.NewEmptyContainer(), "", false, values.MongoSQLValueKind, math.MaxUint64,
		math.MaxUint16, 1024, variable.OffPolymorphicTypeConversionMode, tCfg.mdbVersion,
		tCfg.allowCountOptimization, tCfg.useInformationSchemaDual, tCfg.shouldPushDownEmptyResultSet)
	oCfg := evaluator.NewOptimizerConfig(lgr, collation.Default, values.MongoSQLValueKind,
		true, true, true, true, false)
	pCfg := evaluator.NewPushdownConfig(lgr, tCfg.mdbVersion, tCfg.allowShardedLookups,
		tCfg.allowCrossDBLookups, tCfg.allowRowGeneratorOptimization, tCfg.allowUUIDLiteralComparisons,
		true, tCfg.shouldPushDownEmptyResultSet, true, values.MongoSQLValueKind, tCfg.format, tCfg.formatVersion)
	eCfg := evaluator.NewExecutionConfig(lgr, tCfg.defaultDbName, tCfg.mdbVersion, true, 0,
		values.MongoSQLValueKind, errCommandHandler{}, nil)

	return evaluator.NewQueryConfig(lgr, rCfg, aCfg, oCfg, pCfg, eCfg, tCfg.selectStatementsOnly)
}

// TranslationConfig is a container for all the values needed to translate a SQL
// query via the mongosql library.
type TranslationConfig struct {
	// user-configurable fields
	ctlg          catalog.Catalog
	format        string
	formatVersion int
	defaultDbName string

	// mdbVersion is not configurable by clients of the mongosql library.
	// Internally, this is configured for TranslateSQLQuery (the function
	// used by mongotranslate) and is hard-coded for TranslateSQLQueryRaw
	// (the function used by ADL).
	mdbVersion []uint8

	// non-user-configurable fields
	allowShardedLookups           bool
	allowCrossDBLookups           bool
	allowRowGeneratorOptimization bool
	allowUUIDLiteralComparisons   bool
	allowCountOptimization        bool
	useInformationSchemaDual      bool
	selectStatementsOnly          bool
	shouldPushDownEmptyResultSet  bool
}

// NewTranslationConfig returns a new TranslationConfig
func NewTranslationConfig(ctlg catalog.Catalog, format string, formatVersion int,
	defaultDbName string) *TranslationConfig {
	return newTranslationConfig(ctlg, format, formatVersion, defaultDbName,
		[]uint8{100, 0, 0}, true, true, false, false, false, true, true, true)
}

func newTranslationConfig(
	ctlg catalog.Catalog,
	format string,
	formatVersion int,
	defaultDbName string,
	mdbVersion []uint8,
	allowShardedLookups,
	allowCrossDBLookups,
	allowRowGeneratorOptimization,
	allowUUIDLiteralComparisons,
	allowCountOptimization,
	useInformationSchemaDual,
	selectStatementsOnly,
	shouldPushDownEmptyResultSet bool,
) *TranslationConfig {
	return &TranslationConfig{
		ctlg:                          ctlg,
		format:                        format,
		formatVersion:                 formatVersion,
		defaultDbName:                 defaultDbName,
		mdbVersion:                    mdbVersion,
		allowShardedLookups:           allowShardedLookups,
		allowCrossDBLookups:           allowCrossDBLookups,
		allowRowGeneratorOptimization: allowRowGeneratorOptimization,
		allowUUIDLiteralComparisons:   allowUUIDLiteralComparisons,
		allowCountOptimization:        allowCountOptimization,
		useInformationSchemaDual:      useInformationSchemaDual,
		selectStatementsOnly:          selectStatementsOnly,
		shouldPushDownEmptyResultSet:  shouldPushDownEmptyResultSet,
	}
}

type errCommandHandler struct{}

// Aggregate runs the provided aggregation pipeline against the
// specified database and collection.
func (errCommandHandler) Aggregate(ctx context.Context, db, col string, pipeline []bson.D) (mongodb.Cursor, error) {
	return nil, errors.New("command not supported")
}

// Count runs a count command against the specified database and collection.
func (errCommandHandler) Count(ctx context.Context, db, col string) (int, error) {
	return -1, errors.New("command not supported")
}

// DropTable supports dropping tables.
func (errCommandHandler) DropTable(ctx context.Context, db, tbl string) error {
	return errors.New("command not supported")
}

// DropDatabase drops databases.
func (errCommandHandler) DropDatabase(ctx context.Context, db string) error {
	return errors.New("command not supported")
}

// CreateTable supports creating tables.
func (errCommandHandler) CreateTable(ctx context.Context, db string, table *schema.Table) error {
	return errors.New("command not supported")
}

// CreateDatabase creates Databases.
func (errCommandHandler) CreateDatabase(ctx context.Context, db string) error {
	return errors.New("command not supported")
}

// Insert inserts documents into the specified namespace.
func (errCommandHandler) Insert(ctx context.Context, db, table string, docs []interface{}) error {
	return errors.New("command not supported")
}

// Kill kills a Connection or Query (the KillScope). The targetConnID is the
// ID of the connection that is to be killed. The targetConnID may be the
// current connection id.
func (errCommandHandler) Kill(ctx context.Context, targetConnID uint32, ks evaluator.KillScope) error {
	return errors.New("command not supported")
}

// Resample forces a sample refresh. It must occur in the server
// as that is where the schemata are maintained.
func (errCommandHandler) Resample(context.Context) error {
	return errors.New("command not supported")
}

// RotateLogs rotates the log file.
func (errCommandHandler) RotateLogs() error {
	return errors.New("command not supported")
}

// Set sets the value of the specified variable to the provided value.
func (errCommandHandler) Set(variable.Name, variable.Scope, variable.Kind, values.SQLValue) error {
	return errors.New("command not supported")
}

// SetDatabase sets the current database.
func (errCommandHandler) SetDatabase(db string) error {
	return errors.New("command not supported")
}

// SetScopeAuthorized returns an error if the user is not authorized to
// set variables in the provided scope.
func (errCommandHandler) SetScopeAuthorized(variable.Scope) error {
	return errors.New("command not supported")
}

// UnsetDatabase unsets the current database.
func (errCommandHandler) UnsetDatabase() error {
	return errors.New("command not supported")
}
