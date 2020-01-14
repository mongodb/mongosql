package evaluator

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/internal/option"
	"github.com/10gen/sqlproxy/internal/procutil"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"

	"github.com/shopspring/decimal"
)

// AlgebrizerConfig is a container for all the values needed to run the algebrizer.
type AlgebrizerConfig struct {
	catalog                       catalog.Catalog
	variables                     *variable.Container
	dbName                        string
	groupConcatMaxLen             int64
	isMongos                      bool
	isWriteMode                   bool
	lg                            log.Logger
	sqlValueKind                  values.SQLValueKind
	sqlSelectLimit                uint64
	maxVarcharLength              uint64
	polymorphicTypeConversionMode string
	version                       []uint8
	allowCountOptimization        bool
	useInformationSchemaDual      bool
}

// NewAlgebrizerConfig returns a new AlgebrizerConfig constructed from the
// provided values. AlgebrizerConfigs should always be constructed via this
// function instead of via a struct literal.
func NewAlgebrizerConfig(
	lg log.Logger,
	dbName string,
	c catalog.Catalog,
	vars *variable.Container,
	mongoDBTopology string,
	isWriteMode bool,
	sqlValueKind values.SQLValueKind,
	sqlSelectLimit,
	maxVarcharLength uint64,
	groupConcatMaxLen int64,
	polymorphicTypeConversionMode string,
	mdbVersion []uint8,
	allowCountOptimization,
	useInformationSchemaDual bool,
) *AlgebrizerConfig {
	return &AlgebrizerConfig{
		lg:                            lg,
		dbName:                        dbName,
		catalog:                       c,
		variables:                     vars,
		isMongos:                      mongoDBTopology == "mongos",
		isWriteMode:                   isWriteMode,
		sqlValueKind:                  sqlValueKind,
		sqlSelectLimit:                sqlSelectLimit,
		maxVarcharLength:              maxVarcharLength,
		groupConcatMaxLen:             groupConcatMaxLen,
		polymorphicTypeConversionMode: polymorphicTypeConversionMode,
		version:                       mdbVersion,
		allowCountOptimization:        allowCountOptimization,
		useInformationSchemaDual:      useInformationSchemaDual,
	}
}

// Note: while most errors in the BI-Connector begin with lower case words, any
// algebrizer/mysqlerror begins with a capital letter for consistency with
// MySQL.

// AlgebrizeCommand takes a parsed SQL statement and returns an algebrized form
// of the command.
func AlgebrizeCommand(ctx context.Context, cfg *AlgebrizerConfig, stmt parser.Statement) (Command, error) {
	g := &selectIDGenerator{}
	algebrizer := &algebrizer{
		cfg:                         cfg,
		ctx:                         ctx,
		selectID:                    g.current,
		selectIDGenerator:           g,
		projectedColumnAggregateMap: make(map[int]SQLExpr),
		columnSet:                   make(map[string]struct{}),
		isTopLevel:                  true,
	}

	switch typedStmt := stmt.(type) {
	case *parser.Kill:
		return algebrizer.translateKill(typedStmt)
	case *parser.Flush:
		return algebrizer.translateFlush(typedStmt)
	case *parser.DropTable:
		return algebrizer.translateDropTable(typedStmt), nil
	case *parser.Set:
		return algebrizer.translateSet(typedStmt)
	case *parser.Use:
		return algebrizer.translateUse(typedStmt)
	case *parser.CreateTable:
		return algebrizer.translateCreateTable(typedStmt)
	case *parser.CreateDatabase:
		return algebrizer.translateCreateDatabase(typedStmt)
	case *parser.DropDatabase:
		return algebrizer.translateDropDatabase(typedStmt)
	case *parser.Insert:
		return algebrizer.translateInsert(typedStmt)
	default:
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErNotSupportedYet,
			fmt.Sprintf("statement %T", typedStmt))
	}
}

// AlgebrizeQuery translates a parsed SQL statement into a plan stage. If the
// statement cannot be translated, it will return an error.
func AlgebrizeQuery(ctx context.Context, cfg *AlgebrizerConfig, stmt parser.Statement) (PlanStage, error) {
	g := &selectIDGenerator{}
	algebrizer := &algebrizer{
		cfg:                         cfg,
		ctx:                         ctx,
		selectID:                    g.generate(),
		selectIDGenerator:           g,
		projectedColumnAggregateMap: make(map[int]SQLExpr),
		columnSet:                   make(map[string]struct{}),
		ctes:                        make(ctePlanStages),
		isTopLevel:                  true,
	}

	switch typedStmt := stmt.(type) {
	case parser.SelectStatement:
		return algebrizer.translateRootSelectStatement(typedStmt)
	case *parser.Show:
		return algebrizer.translateShow(typedStmt)
	default:
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErNotSupportedYet,
			fmt.Sprintf("statement %T", typedStmt))
	}
}

type selectIDGenerator struct {
	current int
}

func (g *selectIDGenerator) generate() int {
	g.current++
	return g.current
}

const (
	fromClause   = "from clause"
	whereClause  = "where clause"
	fieldList    = "field list"
	groupClause  = "group clause"
	havingClause = "having clause"
	orderClause  = "order clause"
	limitClause  = "limit clause"
)

// JoinKind specifies the the type of join for
// a given joiner.
type JoinKind string

// These are the possible values for the JoinKind enum.
const (
	InnerJoin        JoinKind = parser.AST_JOIN
	StraightJoin     JoinKind = parser.AST_STRAIGHT_JOIN
	LeftJoin         JoinKind = parser.AST_LEFT_JOIN
	RightJoin        JoinKind = parser.AST_RIGHT_JOIN
	CrossJoin        JoinKind = parser.AST_CROSS_JOIN
	NaturalJoin      JoinKind = parser.AST_NATURAL_JOIN
	NaturalRightJoin JoinKind = parser.AST_NATURAL_RIGHT_JOIN
	NaturalLeftJoin  JoinKind = parser.AST_NATURAL_LEFT_JOIN
)

func cloneCTEs(ctes ctePlanStages) ctePlanStages {
	c := make(ctePlanStages, len(ctes))
	for k, v := range ctes {
		c[k] = &ctePlanStage{v.cte, v.algebrizer, v.planStage}
	}
	return c
}

// newMongoSourceOrDualStage returns a new MongoSource if at least one MongoTable is found in
// the catalog, otherwise it returns a new Dual stage.
func (a *algebrizer) newMongoSourceOrDualStage() (PlanStage, error) {
	if a.cfg.useInformationSchemaDual {
		return NewMongoSourceStage(catalog.InformationSchemaDatabase, &catalog.InformationSchemaDual, a.selectID, "dual"), nil
	}

	// Don't try to push down the dual stage if it is a top-level dual stage
	if a.isTopLevel {
		return NewDualStage(), nil
	}

	// NewMongoSourceDualStage requires $collStats to work, which was only added in 3.4.
	if !a.versionAtLeast(3, 4, 0) {
		return NewDualStage(), nil
	}

	// NewMongoSourceDualStage requires $collStats to work, but $collStats is unreliable in the
	// sharded case. When running against a mongod, $collStats will return 1 document when run on
	// any database and table, even if those do not exist. However, when running $collStats against
	// a mongos on a table that does not exist, $collStats will return 0 documents. To avoid this
	// situation completely, we do not use MongoSource to push down dual stages if we are running
	// against a mongos.
	if a.cfg.isMongos {
		return NewDualStage(), nil
	}

	dualDb, dualTable, err := findMongoDatabaseAndTable(a.ctx, a.cfg.catalog)
	if err != nil && err.code != erNoMongoDBTableFound {
		return nil, err
	}

	return NewMongoSourceDualStage(dualDb.Name(), dualTable, a.selectID, ""), nil
}

type dualTableLookupError struct {
	code    uint16
	message string
}

const (
	erNoMongoDBTableFound = iota
	erDatabaseLookupFailed
	erTableLookupFailed
)

func newDualTableLookupError(code uint16, message string) *dualTableLookupError {
	return &dualTableLookupError{
		code:    code,
		message: message,
	}
}

func (d *dualTableLookupError) Error() string {
	return d.message
}

// findMongoDatabaseAndTable searches the catalog for a MongoDBTable and returns it along with its containing database
// if found, otherwise the function returns nil for both.
func findMongoDatabaseAndTable(ctx context.Context, cl catalog.Catalog) (catalog.Database, catalog.MongoDBTable, *dualTableLookupError) {
	dbs, err := cl.Databases(ctx)
	if err != nil {
		return nil, nil, newDualTableLookupError(erDatabaseLookupFailed, err.Error())
	}

	for _, db := range dbs {
		tables, err := db.Tables(ctx)
		if err != nil {
			return nil, nil, newDualTableLookupError(erTableLookupFailed, err.Error())
		}

		for _, table := range tables {
			if mongoTable, ok := table.(catalog.MongoDBTable); ok {
				return db, mongoTable, nil
			}
		}
	}
	return nil, nil, newDualTableLookupError(erNoMongoDBTableFound, "")
}

type ctePlanStage struct {
	cte        *parser.CTE
	algebrizer *algebrizer
	planStage  PlanStage
}

type ctePlanStages map[string]*ctePlanStage

type algebrizer struct {
	cfg               *AlgebrizerConfig
	ctx               context.Context
	parent            *algebrizer
	selectIDGenerator *selectIDGenerator
	// the selectID to use for projected columns.
	selectID int
	// the selectIDs that are currently used.
	currentSelectIDs []int
	// all the columns in scope.
	columns []*results.Column
	// fully-qualified names of all the columns in scope.
	columnSet map[string]struct{}
	// all the table names in scope.
	tableNames []string
	// indicates whether this context is using columns in its parent.
	correlated bool
	// aggregates found in the current scope.
	aggregates []SQLAggFunctionExpr
	// columns to be projected from this scope.
	projectedColumns ProjectedColumns
	// indicates whether the projected column contains an aggregate.
	projectedColumnAggregateMap map[int]SQLExpr
	// indicates whether to resolve a column using the projected columns first or second.
	resolveProjectedColumnsFirst bool
	// tracks the current clause being processed for the purposes of error messages.
	currentClause string
	// We need to keep track of the ctes as we descend down the scope.
	ctes ctePlanStages
	// Track whether or not this algebrizer is working on the top level of the
	// query. This is used specifically for avoiding pushdown of top-level dual
	// stages.
	isTopLevel bool
}

func (a *algebrizer) valueKind() values.SQLValueKind {
	return a.cfg.sqlValueKind
}

func (a *algebrizer) clone() *algebrizer {
	return &algebrizer{
		cfg:                         a.cfg,
		parent:                      a.parent,
		selectID:                    a.selectID,
		selectIDGenerator:           a.selectIDGenerator,
		projectedColumnAggregateMap: make(map[int]SQLExpr),
		columnSet:                   make(map[string]struct{}),
		ctes:                        cloneCTEs(a.ctes),
		isTopLevel:                  a.isTopLevel,
	}
}

func (a *algebrizer) newSubqueryExprAlgebrizer() *algebrizer {
	return &algebrizer{
		cfg:                         a.cfg,
		parent:                      a,
		selectID:                    a.selectIDGenerator.generate(),
		selectIDGenerator:           a.selectIDGenerator,
		projectedColumnAggregateMap: make(map[int]SQLExpr),
		columnSet:                   make(map[string]struct{}),
		ctes:                        cloneCTEs(a.ctes),
		isTopLevel:                  false,
	}
}

func (a *algebrizer) newDerivedTableAlgebrizer() *algebrizer {
	return &algebrizer{
		cfg:                         a.cfg,
		selectID:                    a.selectIDGenerator.generate(),
		selectIDGenerator:           a.selectIDGenerator,
		projectedColumnAggregateMap: make(map[int]SQLExpr),
		columnSet:                   make(map[string]struct{}),
		ctes:                        cloneCTEs(a.ctes),
		isTopLevel:                  false,
	}
}

func (a *algebrizer) fullName(tableName, columnName string) string {
	fn := columnName
	if tableName != "" {
		fn = tableName + "." + fn
	}

	return fn
}

func (a *algebrizer) lookupColumn(databaseName, tableName, columnName string) (*results.Column, error) {
	var found *results.Column
	for _, column := range a.columns {
		if strings.EqualFold(column.Name, columnName) &&
			(tableName == "" || strings.EqualFold(column.Table, tableName)) &&
			(databaseName == "" || strings.EqualFold(column.Database, databaseName)) {
			if found != nil {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ErNonUniqError,
					a.fullName(tableName, columnName), a.currentClause)
			}
			found = column
		}
	}

	if found == nil {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErBadFieldError,
			a.fullName(tableName, columnName), a.currentClause)
	}

	return found, nil
}

func (a *algebrizer) lookupProjectedColumn(columnName string) (*ProjectedColumn, bool, error) {
	var result ProjectedColumn
	found := false
	isResultComputed := false
	for index, pc := range a.projectedColumns {
		if strings.EqualFold(pc.Name, columnName) {
			// Two columns are ambiguous if the column aliases are the same but the original column
			// names are not (and neither is a computed column like an aggregate function, value
			// literal, + operator, etc.).
			if found &&
				!strings.EqualFold(pc.OriginalName, result.OriginalName) &&
				!strings.EqualFold(pc.OriginalName, "") &&
				!strings.EqualFold(result.OriginalName, "") {
				return nil,
					false,
					mysqlerrors.Defaultf(mysqlerrors.ErNonUniqError,
						columnName,
						a.currentClause)
			}
			// If lookupProjectedColumn is called during the translation of a group clause, we must
			// check that the column requested to group by is not an aggregation function already
			// mapped into projectedColumnAggregateMap. If that's the case, we return an error and
			// do not proceed to pushdown.
			if strings.EqualFold(a.currentClause, groupClause) {
				if _, ok := a.projectedColumnAggregateMap[index+1]; ok {
					return nil,
						false,
						mysqlerrors.Defaultf(mysqlerrors.ErWrongGroupField,
							columnName)
				}
			}
			// Store current column if it's the first match, or if it's computed. Note that result
			// can only be overwritten if we find a match for a non-computed column followed by a
			// match for a computed column (aggregate function, value literal, + operator, etc.).
			if !found || !isResultComputed {
				result = pc
				found = true
				isResultComputed = strings.EqualFold(pc.OriginalName, "")
			}
		}
	}

	return &result, found, nil
}

func (a *algebrizer) findSQLColumn(sqlCol SQLColumnExpr) *results.Column {
	for _, c := range a.columns {
		if strings.EqualFold(c.Database, sqlCol.databaseName) &&
			strings.EqualFold(c.Table, sqlCol.tableName) &&
			strings.EqualFold(c.Name, sqlCol.columnName) {
			return c
		}
	}

	if a.correlated {
		return a.parent.findSQLColumn(sqlCol)
	}
	return nil
}

var (
	// a unique counter, anywhere a unique id is needed.
	uniqueCount      uint64
	uniqueCountMutex = &sync.Mutex{}
)

// getUniqueID returns a unique uint64 using the uniqueCount counter.
func (a *algebrizer) getUniqueID() uint64 {
	uniqueCountMutex.Lock()
	defer uniqueCountMutex.Unlock()
	i := uniqueCount
	// unint64 wraps around to 0 on overflow, which should be more than sufficient.
	// This will only fail if we have more than 2^64 expressions that need a uniqueId
	// in one query. Given memory constraints, such is infeasible, anyway.
	uniqueCount++
	return i
}

// isAggFunction returns true if the byte slice e contains the name of an
// aggregate function and false otherwise.
func (a *algebrizer) isAggFunction(name string) bool {
	name = strings.ToLower(name)
	_, ok := parser.AggregationFunctions[name]
	return ok
}

// ShouldConvert returns true if `c` is a column must be wrapped in a convert expression.
// That means that either `c` contained polymorphic data types during sampling and the
// PolymorphicConversionMode is "PolymorphicConversionTypeModeFast", or simply if the
// PolymorphicConversionMode is "PolymorphicConversionModeSafe". "PolymorphicTypeConversionModeOff"
// always returns false.
func shouldConvert(c *results.Column, mode string) bool {
	if mode == variable.OffPolymorphicTypeConversionMode {
		return false
	}
	if mode == variable.SafePolymorphicTypeConversionMode {
		return true
	}
	// In fast mode, we only want to introduce converts when we think they are
	// necessary to avoid query aggregation failures. Two places we know that
	// can definitely introduce aggregation failures are fields that were
	// sampled as polymorphic, and fields that have had their type altered with
	// the ALTER statement.
	return c.IsPolymorphic || c.HasAlteredType
}

func (a *algebrizer) resolveColumnExpr(databaseName, tableName,
	columnName string) (SQLExpr, error) {

	if a.resolveProjectedColumnsFirst && tableName == "" {
		expr, ok, err := a.lookupProjectedColumn(columnName)
		if err != nil {
			return nil, err
		}
		if ok {
			return expr.Expr, nil
		}
	}

	column, err := a.lookupColumn(databaseName, tableName, columnName)
	if err == nil {
		if a.currentClause != whereClause && column.MongoType == schema.MongoFilter {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErBadFieldError, column.Name, column.Table)
		}
		colExpr := NewSQLColumnExpr(column.SelectID, column.Database, column.Table,
			column.Name, column.EvalType, column.MongoType, false, column.Nullable)
		mode := a.cfg.polymorphicTypeConversionMode
		if shouldConvert(column, mode) {
			return NewSQLConvertExpr(colExpr, column.EvalType), nil
		}
		return colExpr, nil
	}

	if !a.resolveProjectedColumnsFirst && tableName == "" {
		expr, ok, lookupError := a.lookupProjectedColumn(columnName)
		if lookupError != nil {
			return nil, lookupError
		}
		if ok {
			return expr.Expr, nil
		}
	}

	// lookupColumn returns an error if the column referenced is ambiguous or if it doesn't exist
	// in the current scope. If the column is ambiguous, we want to return that error to the user.
	// Otherwise the column was not found in the current scope, so it could be a correlated column.
	// We search parent scopes until we find the select that brings this column into scope, if any.
	if errSQL, ok := err.(*mysqlerrors.MySQLError); ok && errSQL.Code == mysqlerrors.ErNonUniqError {
		return nil, err
	}
	if a.parent != nil {
		expr, parentErr := a.parent.resolveColumnExpr(databaseName, tableName, columnName)
		if parentErr == nil {
			a.correlated = true
			col := expr.(SQLColumnExpr)
			col.correlated = true
			return col, nil
		}
	}

	return nil, err
}

func (a *algebrizer) registerColumns(columns []*results.Column) error {
	var sb strings.Builder
	var empty struct{}

	// this ensures that we have no duplicate columns. We have to check duplicates
	// against the existing columns as well as against itself.
	for _, c := range columns {
		sb.WriteString(strings.ToLower(c.Database))
		sb.WriteString(strings.ToLower(c.Table))
		sb.WriteString(strings.ToLower(c.Name))
		if _, ok := a.columnSet[sb.String()]; ok {
			return mysqlerrors.Defaultf(mysqlerrors.ErDupFieldname, a.fullName(c.Table, c.Name))
		}
		a.columnSet[sb.String()] = empty
		a.columns = append(a.columns, c)
		if !containsInt(a.currentSelectIDs, c.SelectID) {
			a.currentSelectIDs = append(a.currentSelectIDs, c.SelectID)
		}
		sb.Reset()
	}

	return nil
}

// registerTable ensures that we have no duplicate table names or aliases.
func (a *algebrizer) registerTable(dbName, tableName string) error {
	qualifiedTableName := fullyQualifiedTableName(dbName, tableName)
	for _, registeredName := range a.tableNames {
		if strings.EqualFold(qualifiedTableName, registeredName) {
			return mysqlerrors.Defaultf(mysqlerrors.ErNonuniqTable, tableName)
		}
	}

	a.tableNames = append(a.tableNames, qualifiedTableName)

	return nil
}

func (a *algebrizer) translateFlush(flush *parser.Flush) (*FlushCommand, error) {
	switch flush.Kind {
	case parser.FlushLogs:
		return NewFlushCommand(FlushLogs), nil
	case parser.FlushSample:
		return NewFlushCommand(FlushSample), nil
	}

	return nil, fmt.Errorf("unsupported flush kind: %v", flush.Kind)
}

func (a *algebrizer) translateDropTable(ddl *parser.DropTable) *DropTableCommand {
	// DropTable is allowed outside of --writeMode, so this is infallible.
	return NewDropTableCommand(a.cfg.catalog,
		ddl.Name.Qualifier.Else(a.cfg.dbName),
		ddl.Name.Name,
		ddl.IfExists)
}

func (a *algebrizer) translateDropDatabase(ddl *parser.DropDatabase) (*DropDatabaseCommand, error) {
	if !a.cfg.isWriteMode {
		return nil, fmt.Errorf("drop database requires --writeMode")
	}
	return NewDropDatabaseCommand(a.cfg.catalog,
		ddl.Name,
		ddl.IfExists), nil
}

func (a *algebrizer) translateInsert(ins *parser.Insert) (*InsertCommand, error) {
	if !a.cfg.isWriteMode {
		return nil, fmt.Errorf("insert requires --writeMode")
	}
	// We need the dbName for the mongodb insert command.
	dbName := ins.Table.Qualifier.Else(a.cfg.dbName)
	db, err := a.cfg.catalog.Database(a.ctx, dbName)
	if err != nil {
		return nil, err
	}
	// We need the table name as well as other information from the table
	// for the insert command and proper value type conversion.
	table, err := db.Table(a.ctx, ins.Table.Name)
	if err != nil {
		return nil, err
	}
	// tableCols will be used for finding the appropriate type for
	// a column as well the exact MongoName for generating the proper
	// document to insert.
	tableCols := table.Columns()
	insertColNames := make([]string, len(ins.Columns))
	// The ins.Columns are SQLColumn names, we need to map
	// them to the underlying MongoNames through the catalog.
	for i := range ins.Columns {
		tableColRef, colErr := table.Column(ins.Columns[i].Name)
		if colErr != nil {
			return nil, colErr
		}
		insertColNames[i] = tableColRef.MongoName
	}
	// numInsertCols is the number of columns specified in the insert statement:
	// for
	//   insert into foo(x,y) values ...
	// numInsertCols would be 2
	numInsertCols := len(insertColNames)
	// numTableCols is the number of columns in the table, e.g.:
	// for
	//   create table foo(x int, y int, z varchar(15))
	// numTableCols would be 3
	// numTableCols is used to decide how many []SQLExpr to put in a given row,
	// and to check that the number of specified values in a row is correct
	// when numInsertCols is 0.
	numTableCols := len(tableCols)
	// get the positionMap (see generatePositionMap comment for more details).
	positionMap := a.generatePositionMap(insertColNames, tableCols)
	// now convert the rows of values from parser.ValueListList to [][]SQLExpr.
	exprs, err := a.valueListListsToSQLExprListLists(numInsertCols, numTableCols, ins.Values)
	if err != nil {
		return nil, err
	}
	return NewInsertCommand(dbName, ins.Table.Name, tableCols, positionMap, exprs), nil
}

// generatePositionMap generates a positionMap for sending to the InsertCommand.
// A positionMap is a mapping from a given insert column's mongo name to a position index into each row
// of the valuesListList. Any column that does not appear in the insert columns list will not appear
// in this map unless _no_ columns are specified (in which case they will all appear with the
// position index corresponding to the table position).
//
// This map is necessary because the insert columns list can be in a different order than the table
// columns, e.g.:
//   create table foo(x int, y int);
//   insert into foo(y, x) values(3,4);
func (a *algebrizer) generatePositionMap(insertColNames []string, tableCols results.Columns) map[string]int {
	var positionMap map[string]int
	if len(insertColNames) != 0 {
		positionMap = make(map[string]int, len(insertColNames))
		for i, colName := range insertColNames {
			positionMap[colName] = i
		}
	} else {
		// If the insert statement did not specify columns, the positionMap
		// is just the same order as the columns in the table. e.g. tablePositions.
		positionMap = make(map[string]int, len(tableCols))
		for i, col := range tableCols {
			positionMap[col.MongoName] = i
		}
	}
	return positionMap
}

func (a *algebrizer) valueListListsToSQLExprListLists(numInsertCols, numTableCols int,
	valLists parser.ValueListList) ([][]SQLExpr, error) {
	exprListList := make([][]SQLExpr, len(valLists))
	firstSize := len(valLists[0])
	for i := range valLists {
		// MySQL enforces that every row has the same size. This means you
		// can't mix empty rows with non-empty rows even for an insert with
		// no columns specified.
		if len(valLists[i]) != firstSize {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongValueCountOnRow, i+1)
		}
		var err error
		exprListList[i], err = a.valueListToSQLExprList(i, numInsertCols, numTableCols, valLists[i])
		if err != nil {
			return nil, err
		}
	}
	return exprListList, nil
}

func (a *algebrizer) valueListToSQLExprList(i, numInsertCols, numTableCols int, vals parser.ValueList) ([]SQLExpr, error) {
	// If no columns are specified, and the row is empty, we insert a row
	// of all default values.
	if numInsertCols == 0 && len(vals) == 0 {
		exprList := make([]SQLExpr, numTableCols)
		for i := range exprList {
			exprList[i] = NewSQLValueExpr(values.NewSQLNull(a.valueKind()))
		}
		return exprList, nil
	}
	// If columns are specified, there must be as many values as there are specified columns.
	if numInsertCols > 0 && numInsertCols != len(vals) ||
		// If no columns are specified, there must be as many values as there are columns in the table.
		numInsertCols == 0 && numTableCols != len(vals) {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongValueCountOnRow, i+1)
	}

	exprList := make([]SQLExpr, len(vals))
	for i, val := range vals {
		// We just translate DEFAULT to NULL here. If we ever support non-NULL
		// DEFAULTs, they will likewise be translated here.
		if _, ok := val.(parser.Default); ok {
			exprList[i] = NewSQLValueExpr(values.NewSQLNull(a.valueKind()))
			continue
		}
		converted, err := a.translateExpr(val)
		if err != nil {
			panic(fmt.Sprintf("got error: %v in convertParserValue", err))
		}
		switch converted.(type) {
		case SQLValueExpr:
			exprList[i] = converted
		default:
			// This would suggest an error in the parser.
			panic(fmt.Sprintf("insert only accepts values not %T", converted))
		}
	}
	return exprList, nil
}

func (a *algebrizer) translateCreateDatabase(ddl *parser.CreateDatabase) (*CreateDatabaseCommand, error) {
	if !a.cfg.isWriteMode {
		return nil, fmt.Errorf("create database requires --writeMode")
	}
	return NewCreateDatabaseCommand(
		a.cfg.catalog,
		ddl.Name,
		ddl.IfNotExists), nil
}

func getCreateTableColumns(colDefs []*parser.ColumnDefinition) []*schema.Column {
	cols := make([]*schema.Column, len(colDefs))
	for i, def := range colDefs {
		sqlType, err := schema.GetSQLType(def.Type.BaseType)
		if err != nil {
			// This suggests an error in the CST command_desugarer.
			panic(err)
		}
		cols[i] = schema.NewColumn(
			def.Name.Name,
			sqlType,
			def.Name.Name,
			schema.GetMongoTypeFromSQLType(sqlType),
			def.Null,
			def.Comment,
		)
	}
	return cols
}

func getCreateTableIndexes(colNames map[string]struct{}, indexDefs []*parser.IndexDefinition,
	colDefs []*parser.ColumnDefinition) (schema.Indexes, error) {
	indexes := make(schema.Indexes, len(indexDefs))
	for i, def := range indexDefs {
		name := ""
		if def.Name.IsSome() {
			name = def.Name.Unwrap()
		}
		parts := make([]schema.IndexPart, len(def.KeyParts))
		for j, partDef := range def.KeyParts {
			if _, ok := colNames[partDef.Column.Name]; !ok {
				return nil, fmt.Errorf("index defined on non-existent column '%s'", partDef.Column.Name)
			}
			parts[j] = schema.NewIndexPart(partDef.Column.Name, partDef.Direction)
		}
		indexes[i] = schema.NewIndex(name, def.Unique, def.FullText, parts)
	}
	for _, def := range colDefs {
		if def.Unique {
			index := schema.NewIndex(def.Name.Name+"_unique", true, false, []schema.IndexPart{
				schema.NewIndexPart(def.Name.Name, 1),
			})
			indexes = append(indexes, index)
		}
	}
	return indexes, nil
}

func (a *algebrizer) translateCreateTable(ddl *parser.CreateTable) (*CreateTableCommand, error) {
	if !a.cfg.isWriteMode {
		return nil, fmt.Errorf("create table requires --writeMode")
	}
	colDefs, indexDefs := ddl.GetColumnAndIndexDefintions()
	columns := getCreateTableColumns(colDefs)
	colNames := make(map[string]struct{})
	for _, col := range columns {
		colNames[col.SQLName()] = struct{}{}
	}
	indexes, err := getCreateTableIndexes(colNames, indexDefs, colDefs)
	if err != nil {
		return nil, err
	}
	logger := log.NewComponentLogger(log.SchemaComponent, log.GlobalLogger())
	comment := option.NoneString()
	for _, opt := range ddl.TableOptions {
		switch typedOpt := opt.(type) {
		case parser.TableComment:
			comment = option.SomeString(string(typedOpt))
		}
	}
	table, err := schema.NewTable(logger,
		ddl.Name.Name,
		ddl.Name.Name,
		nil,
		columns,
		indexes,
		comment,
	)
	if err != nil {
		return nil, err
	}
	return NewCreateTableCommand(a.cfg.catalog,
		ddl.Name.Qualifier.Else(a.cfg.dbName),
		table,
		ddl.IfNotExists), nil
}

func (a *algebrizer) translateGroupBy(groupby parser.GroupBy) ([]SQLExpr, error) {
	// Make sure to remove duplicate keys. They are entirely unnecessary,
	// and also cause failures if we push down (all keys in a group by must
	// be unique in MongoDB). Keys are duplicate if they have the same
	// string representation. Unfortunately, since we do not de-duplicate
	// repeated sub expressions in the parser, this is the best we can
	// do, for now. If we ever add de-deuplication in the parser, this
	// can be changed.
	uniqueKeys := make(map[string]SQLExpr)
	var keys []SQLExpr
	for _, g := range groupby {

		key, err := a.translatePossibleColumnRefExpr(g)
		if err != nil {
			return nil, err
		}

		keyStr := key.String()
		if _, ok := uniqueKeys[keyStr]; !ok {
			uniqueKeys[key.String()] = key
			keys = append(keys, key)
		}
	}

	return keys, nil
}

func (a *algebrizer) translateKill(kill *parser.Kill) (*KillCommand, error) {
	killID, err := a.translateExpr(kill.ID)
	if err != nil {
		return nil, err
	}

	switch kill.Scope {
	case parser.AST_KILL_QUERY:
		return NewKillCommand(killID, KillQuery), nil
	default:
		return NewKillCommand(killID, KillConnection), nil
	}
}

func (a *algebrizer) translateLimit(limit *parser.Limit) (uint64, uint64, error) {
	var rowcount, offset uint64

	if limit.Offset != nil {
		eval, err := a.translateExpr(limit.Offset)
		if err != nil {
			return 0, 0, err
		}

		val, ok := eval.(SQLValueExpr)
		if !ok {
			return 0, 0, mysqlerrors.Defaultf(mysqlerrors.ErWrongSpvarTypeInLimit)
		}
		switch typedE := val.Value.(type) {
		case values.SQLUint64:
			offset = values.Uint64(typedE)
		case values.SQLInt64:
			if values.Int64(typedE) < 0 {
				return 0,
					0,
					mysqlerrors.Newf(mysqlerrors.ErSyntaxError,
						"Offset cannot be negative")
			}
			offset = values.Uint64(typedE)
		default:
			return 0, 0, mysqlerrors.Defaultf(mysqlerrors.ErWrongSpvarTypeInLimit)
		}
	}

	if limit.Rowcount != nil {
		eval, err := a.translateExpr(limit.Rowcount)
		if err != nil {
			return 0, 0, err
		}

		val, ok := eval.(SQLValueExpr)
		if !ok {
			return 0, 0, mysqlerrors.Defaultf(mysqlerrors.ErWrongSpvarTypeInLimit)
		}
		switch typedE := val.Value.(type) {
		case values.SQLUint64:
			rowcount = values.Uint64(typedE)
		case values.SQLInt64:
			if values.Int64(typedE) < 0 {
				return 0,
					0,
					mysqlerrors.Newf(mysqlerrors.ErSyntaxError,
						"Rowcount cannot be negative")
			}
			rowcount = values.Uint64(typedE)
		default:
			return 0, 0, mysqlerrors.Defaultf(mysqlerrors.ErWrongSpvarTypeInLimit)
		}
	}

	return offset, rowcount, nil
}

func (a *algebrizer) translateOrderBy(orderby parser.OrderBy) ([]*OrderByTerm, error) {
	var terms []*OrderByTerm
	for _, o := range orderby {
		term, err := a.translateOrder(o)
		if err != nil {
			return nil, err
		}

		terms = append(terms, term)
	}

	return terms, nil
}

func (a *algebrizer) translateOrder(order *parser.Order) (*OrderByTerm, error) {
	ascending := !strings.EqualFold(order.Direction, parser.AST_DESC)
	e, err := a.translatePossibleColumnRefExpr(order.Expr)
	if err != nil {
		return nil, err
	}

	return &OrderByTerm{
		expr:      e,
		ascending: ascending,
	}, nil
}

func (a *algebrizer) translateRootSelectStatement(selectStatement parser.SelectStatement) (
	plan PlanStage, err error) {
	defer func() {
		if err != nil {
			if ps, ok := plan.(*ProjectStage); ok {
				panic(fmt.Sprintf("non-project top-level stage: %T", ps))
			}
		}
	}()

	plan, err = a.translateSelectStatement(selectStatement)
	if err != nil {
		return nil, err
	}

	// explicit limit takes precedence
	s, ok := selectStatement.(*parser.Select)
	if ok && s.Limit != nil {
		return plan, nil
	}

	// only add the system-wide limit if it has been changed from the default
	// otherwise, we can't push down queries by default
	sqlSelectLimit := a.cfg.sqlSelectLimit
	if sqlSelectLimit != math.MaxUint64 {
		if pr, ok := plan.(*ProjectStage); ok {
			plan = NewLimitStage(pr.source, 0, sqlSelectLimit)
			plan = NewProjectStage(plan, pr.projectedColumns...)
		}
	}

	return plan, nil
}

func (a *algebrizer) translateCTEs(ctes parser.CTEs) error {
	var empty struct{}
	// You can't have multiple ctes with the same name at the current level.
	seenAliasesSet := make(map[string]struct{}, len(ctes))
	for _, cte := range ctes {
		strName := strings.ToLower(cte.TableName.Name)
		if _, ok := seenAliasesSet[strName]; ok {
			return mysqlerrors.Defaultf(mysqlerrors.ErNonuniqTable, strName)
		}
		seenAliasesSet[strName] = empty
		cteAlgebrizer := a.newDerivedTableAlgebrizer()
		plan, err := cteAlgebrizer.translateSelectStatement(cte.Query)
		if err != nil {
			return err
		}
		evaluator := &ctePlanStage{cte, cteAlgebrizer, plan}
		// You can override at different levels however.
		a.ctes[strName] = evaluator
	}
	return nil
}

func (a *algebrizer) translateSelectStatement(
	selectStatement parser.SelectStatement) (PlanStage, error) {
	switch typedS := selectStatement.(type) {
	case *parser.Select:
		if typedS.With != nil {
			if typedS.With.Recursive {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ErRecursiveCTEsNotSupported)
			}
			err := a.translateCTEs(typedS.With.CTEs)
			if err != nil {
				return nil, err
			}
		}
		return a.translateSelect(typedS)
	case *parser.Union:
		if typedS.With != nil {
			if typedS.With.Recursive {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ErRecursiveCTEsNotSupported)
			}
			err := a.translateCTEs(typedS.With.CTEs)
			if err != nil {
				return nil, err
			}
		}
		return a.translateUnion(typedS)
	default:
		return nil,
			mysqlerrors.Defaultf(mysqlerrors.ErNotSupportedYet,
				parser.String(selectStatement))
	}
}

func (a *algebrizer) translateSelect(sel *parser.Select) (PlanStage, error) {
	builder := &queryPlanBuilder{
		algebrizer: a,
		selectID:   a.selectID,
	}

	// 1. Translate all the tables, subqueries, and joins in the FROM clause.
	// This establishes all the columns which are in scope.
	if sel.From != nil {
		a.currentClause = fromClause
		var isUnqualifiedSelectStar bool
		if len(sel.SelectExprs) == 1 {
			if expr, ok := sel.SelectExprs[0].(*parser.StarExpr); ok {
				isUnqualifiedSelectStar = expr.TableName.IsNone()
			}
		}

		// hasGlobalStraightJoin indicates whether the query uses the "select straight_join" syntax
		var hasGlobalStraightJoin bool
		if sel.QueryGlobals != nil {
			hasGlobalStraightJoin = sel.QueryGlobals.StraightJoin
		} else {
			hasGlobalStraightJoin = false
		}

		plan, err := a.translateTableExprs(sel.From, isUnqualifiedSelectStar, hasGlobalStraightJoin)
		if err != nil {
			return nil, err
		}

		if a.cfg.allowCountOptimization {
			if mongoSource, ok := isCountOptimizable(sel, plan); ok {
				a.currentClause = fieldList
				pcs, translateErr := a.translateSelectExprs(sel.SelectExprs)
				if translateErr != nil {
					return nil, translateErr
				}
				a.projectedColumns = pcs

				if sel.OrderBy != nil {
					a.currentClause = orderClause
					_, translateErr := a.translateOrderBy(sel.OrderBy)
					if translateErr != nil {
						return nil, translateErr
					}
				}

				pcs[0].Expr = NewSQLColumnExpr(pcs[0].SelectID, pcs[0].Database,
					pcs[0].Table, pcs[0].Name, pcs[0].EvalType, schema.MongoNone, false, pcs[0].Column.Nullable)
				plan = NewCountStage(mongoSource, pcs[0])
				plan = NewProjectStage(plan, pcs[0])
				return plan, nil
			}
		}

		builder.from = plan

		selectIDsInScope := a.currentSelectIDs
		parent := a.parent
		for parent != nil {
			selectIDsInScope = append(selectIDsInScope, parent.currentSelectIDs...)
			parent = parent.parent
		}
		builder.exprCollector = newSQLColExprCollector(selectIDsInScope)

		if plan != nil {
			err = builder.includeFrom(plan)
		}
		if err != nil {
			return nil, err
		}
	}

	// 2. Translate all the other clauses from this scope. We aren't going to create the plan stages
	// yet because the expressions may need to be substituted if a group by exists.
	if sel.Where != nil {
		a.currentClause = whereClause
		err := builder.includeWhere(sel.Where)

		if builder.where != nil && !isBooleanComparable(builder.where.EvalType()) {
			builder.where = NewSQLConvertExpr(builder.where, types.EvalBoolean)
		}
		if err != nil {
			return nil, err
		}
	}

	if sel.SelectExprs != nil {
		a.currentClause = fieldList
		err := builder.includeSelect(sel.SelectExprs)
		if err != nil {
			return nil, err
		}

		// set projected columns globally because column resolution depends on
		// this list from which GROUP BY and HAVING resolve from it second, and
		// ORDER BY resolves from it first.
		a.projectedColumns = builder.project
	}

	if sel.GroupBy != nil {
		a.currentClause = groupClause
		err := builder.includeGroupBy(sel.GroupBy)
		if err != nil {
			return nil, err
		}
	}

	if sel.Having != nil {
		a.currentClause = havingClause
		err := builder.includeHaving(sel.Having)
		if err != nil {
			return nil, err
		}
	}

	if sel.QueryGlobals != nil {
		builder.distinct = sel.QueryGlobals.Distinct
	}

	// order by resolves from the projected columns first
	a.resolveProjectedColumnsFirst = true

	if sel.OrderBy != nil {
		a.currentClause = orderClause
		err := builder.includeOrderBy(sel.OrderBy)
		if err != nil {
			return nil, err
		}
	}

	if sel.Limit != nil {
		a.currentClause = limitClause
		err := builder.includeLimit(sel.Limit)
		if err != nil {
			return nil, err
		}
	}

	builder.includeAggregates(a.aggregates)

	// 3. Build the stages.
	return builder.build(), nil
}

func (a *algebrizer) translateUnion(union *parser.Union) (PlanStage, error) {
	var err error

	left, err := a.translateSelectStatement(union.Left)
	if err != nil {
		return nil, err
	}

	right, err := a.clone().translateSelectStatement(union.Right)
	if err != nil {
		return nil, err
	}

	leftCols := left.Columns()
	rightCols := right.Columns()

	if len(leftCols) != len(rightCols) {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongNumberOfColumnsInSelect)
	}

	var projectedColumns ProjectedColumns
	var plan PlanStage

	switch union.Type {
	case parser.AST_UNION:
		plan = NewUnionStage(UnionDistinct, left, right)
		projectedColumns = columnsToProjectedColumns(plan.Columns())
		plan = NewGroupByStage(plan, projectedColumns.Exprs(), projectedColumns)
	case parser.AST_UNION_ALL:
		plan = NewUnionStage(UnionAll, left, right)
		projectedColumns = columnsToProjectedColumns(plan.Columns())
	default:
		return nil, mysqlerrors.Newf(mysqlerrors.ErNotSupportedYet,
			"Cannot perform set operation '%s'", union.Type)
	}
	return NewProjectStage(plan, projectedColumns...), nil
}

func (a *algebrizer) translateSelectExprs(
	selectExprs parser.SelectExprs) (ProjectedColumns, error) {
	var projectedColumns ProjectedColumns
	mode := a.cfg.polymorphicTypeConversionMode

	for _, selectExpr := range selectExprs {
		switch typedE := selectExpr.(type) {

		case *parser.StarExpr:
			databaseName := typedE.DatabaseName.Else("")
			tableName := typedE.TableName.Else("")

			for _, column := range a.columns {
				if column.MongoType == schema.MongoFilter {
					continue
				}
				// If the form is of string dot wildcard, sql
				// automatically assumes the first string
				// refers to a table.
				if (tableName == "" && databaseName == "") ||
					(databaseName == "" && strings.EqualFold(tableName, column.Table)) ||
					(strings.EqualFold(tableName, column.Table) &&
						strings.EqualFold(databaseName, column.Database)) {
					projectedColumn := newProjectedColumnFromColumn(column)
					projectedColumn.SelectID = a.selectID
					if shouldConvert(column, mode) {
						projectedColumn.Expr = NewSQLConvertExpr(projectedColumn.Expr, column.EvalType)
					}
					projectedColumns = append(projectedColumns, projectedColumn)
				}
			}

		case *parser.NonStarExpr:

			currentAggregateLength := len(a.aggregates)
			translatedExpr, err := a.translateExpr(typedE.Expr)
			if err != nil {
				return nil, err
			}

			if currentAggregateLength < len(a.aggregates) {
				a.projectedColumnAggregateMap[len(projectedColumns)+1] =
					a.aggregates[currentAggregateLength]
			}

			var projectedColumn *ProjectedColumn

			if sqlCol, ok := translatedExpr.(SQLColumnExpr); ok {
				if c := a.findSQLColumn(sqlCol); c != nil {
					projectedColumn = newProjectedColumnFromColumnWithExpr(c, translatedExpr)
				}
			}

			// This happens when the select expression is more than just a
			// column: it could be a scalar or aggregate function, or
			// any sort of operator like '+'
			if projectedColumn == nil {
				dbName := getDatabaseName(translatedExpr)
				cb := results.NewColumnBuilder()
				cb.SetColumnType(results.NewColumnType(translatedExpr.EvalType(), schema.MongoNone))
				cb.SetSelectID(a.selectID)
				cb.SetTable("")
				cb.SetOriginalTable("")
				cb.SetDatabase(dbName)
				cb.SetName("")
				cb.SetOriginalName("")
				cb.SetMappingRegistryName("")
				cb.SetMongoName("")
				cb.SetPrimaryKey(false)
				cb.SetComments("")
				cb.SetIsPolymorphic(false)
				cb.SetHasAlteredType(false)
				cb.SetNullable(true)
				c := cb.Build()
				projectedColumn = &ProjectedColumn{
					Expr:   translatedExpr,
					Column: c,
				}
			}
			if _, ok := translatedExpr.(*SQLVariableExpr); !ok {
				if sqlCol, ok := typedE.Expr.(*parser.ColName); ok {
					projectedColumn.Name = sqlCol.Name
				}
			}
			if typedE.As.IsSome() {
				projectedColumn.Name = typedE.As.Unwrap()
			} else if projectedColumn.Name == "" {
				projectedColumn.Name = parser.String(typedE)
			}

			projectedColumns = append(projectedColumns, *projectedColumn)
		}
	}

	return projectedColumns, nil
}

var charsetOrCollationVars = map[string]struct{}{
	"character_set_client": {}, "@character_set_client": {}, "@@character_set_client": {},
	"character_set_connection": {}, "@character_set_connection": {}, "@@character_set_connection": {},
	"character_set_database": {}, "@character_set_database": {}, "@@character_set_database": {},
	"character_set_file_system": {}, "@character_set_file_system": {}, "@@character_set_file_system": {},
	"character_set_results": {}, "@character_set_results": {}, "@@character_set_results": {},
	"character_set_server": {}, "@character_set_server": {}, "@@character_set_server": {},
	"character_set_system": {}, "@character_set_system": {}, "@@character_set_system": {},
	"collation_connection": {}, "@collation_connection": {}, "@@collation_connection": {},
	"collation_database": {}, "@collation_database": {}, "@@collation_database": {},
	"collation_server": {}, "@collation_server": {}, "@@collation_server": {},
}

// translateSetCollationOrCharsetExpr handles translating the SetExpr for a variable set
// where the variable is one of our charset or collation variables denoted by the set above.
func (a *algebrizer) translateSetCollationOrCharsetExpr(e parser.Expr) (SQLExpr, error) {
	if col, ok := e.(*parser.ColName); ok && !strings.HasPrefix(col.Name, "@") {
		// If the expression is a name that does not begin with @, we just create
		// a new string representation of the collation name as the looked up value.
		// This is necessary because MySQL allows identifiers as collations, and
		// our variable setting code assumes string values for collations. If it begins with
		// @ we will continue to treat it as a variable.
		return NewSQLValueExpr(values.NewSQLVarchar(a.valueKind(), strings.ToLower(col.Name))), nil
	}
	expr, err := a.translateExpr(e)
	if err != nil {
		return nil, err
	}
	return expr, nil
}

func (a *algebrizer) translateSet(set *parser.Set) (*SetCommand, error) {
	assignments := []*SQLAssignmentExpr{}
	for _, e := range set.Exprs {
		v, err := a.translateVariableExpr(e.Name)
		if err != nil {
			return nil, err
		}

		var expr SQLExpr
		if _, ok := charsetOrCollationVars[e.Name.Name]; ok {
			expr, err = a.translateSetCollationOrCharsetExpr(e.Expr)
		} else {
			expr, err = a.translateExpr(e.Expr)
		}
		if err != nil {
			return nil, err
		}

		if set.Scope != "" {
			v.Scope = variable.ScopeFromString(set.Scope)
		}

		assignments = append(assignments, &SQLAssignmentExpr{
			variable: v,
			expr:     expr,
		})
	}

	return NewSetCommand(assignments), nil
}

// nolint: unparam
func (a *algebrizer) translateUse(use *parser.Use) (*UseCommand, error) {
	return NewUseCommand(use.DBName), nil
}

func (a *algebrizer) translateTableExprs(tableExprs parser.TableExprs,
	isUnqualifiedSelectStar bool, hasGlobalStraightJoin bool) (PlanStage, error) {
	var plan PlanStage
	var colsForProjection []*results.Column
	for i, tableExpr := range tableExprs {
		temp, cols, err := a.translateTableExpr(tableExpr, hasGlobalStraightJoin)
		if err != nil {
			return nil, err
		}
		colsForProjection = append(colsForProjection, cols...)
		if i == 0 {
			plan = temp
		} else {
			// Commas in from clause are translated into cross joins, unless the query
			// uses the "select straight_join" syntax, in which case they are interpreted
			// as straight_joins.
			var joinKind JoinKind
			if hasGlobalStraightJoin {
				joinKind = StraightJoin
			} else {
				joinKind = CrossJoin
			}
			plan = NewJoinStage(joinKind, plan, temp, NewSQLValueExpr(values.NewSQLBool(a.valueKind(), true)))
		}
	}

	if isUnqualifiedSelectStar {
		if len(tableExprs) == 1 {
			if _, isDual := tableExprs[0].(*parser.DualTableExpr); isDual {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ErNoTablesUsed)
			}
		}

		a.columns = colsForProjection
	}

	return plan, nil
}

// getTableName gets the name of the table that contains colName. The table
// name is only specific to the column when tableExpr.(type) = JoinTableExpr.
func (a *algebrizer) getTableName(tableExpr parser.TableExpr, colName string, columns []*results.Column) (string, error) {
	switch typedE := tableExpr.(type) {
	case *parser.AliasedTableExpr:
		if typedE.As.IsSome() {
			return typedE.As.Unwrap(), nil
		}
		// only legal type is TableName, as other AliasedTableExpr's must have AS clause specified
		name, ok := typedE.Expr.(*parser.TableName)
		if !ok {
			return "", mysqlerrors.Newf(mysqlerrors.ErParseError, "A %s must have an alias", typedE.Expr)
		}
		return name.Name, nil
	case *parser.ParenTableExpr:
		return a.getTableName(typedE.Expr, colName, columns)
	case *parser.JoinTableExpr:
		var tableName string
		var tableFound bool
		// find the name of the table the column was in before the join
		for _, column := range columns {
			if column.Name == colName {
				if !tableFound {
					tableName = column.Table
					tableFound = true
				} else {
					return "", mysqlerrors.Defaultf(mysqlerrors.ErNonUniqError, colName, a.currentClause)
				}
			}
		}
		// if column named colName was not found in any table in the JoinTableExpr
		if !tableFound {
			return "", mysqlerrors.Defaultf(mysqlerrors.ErBadFieldError, colName, a.currentClause)
		}
		return tableName, nil
	default:
		return "", mysqlerrors.Defaultf(mysqlerrors.ErNotSupportedYet, typedE)
	}
}

// convertToAnd is a function takes in a list of columns returns an expression
// which ANDs together a series of expressions comparing these columns, eg. (a,
// b, d) -> foo.a=bar.a && foo.b=bar.b && foo.d=bar.d.
func (a *algebrizer) convertToAnd(
	columns parser.ColumnExprs,
	leftCols []*results.Column,
	rightCols []*results.Column,
	tableExpr *parser.JoinTableExpr) (parser.Expr, error) {
	var expression parser.Expr
	seenColumns := make(map[string]struct{})
	leftDatabaseName := leftCols[0].Database
	rightDatabaseName := rightCols[0].Database
	for _, column := range columns {
		colName := column.Name
		// if column is not already in the comparison
		if _, ok := seenColumns[colName]; !ok {
			var emptyStruct struct{}
			seenColumns[colName] = emptyStruct
			// extract the left and right table names
			leftTableName, err := a.getTableName((*tableExpr).LeftExpr, colName, leftCols)
			if err != nil {
				return nil, err
			}
			rightTableName, err := a.getTableName((*tableExpr).RightExpr, colName, rightCols)
			if err != nil {
				return nil, err
			}

			// Need to assign databases for cross db joins.
			leftExpr := &parser.ColName{
				Database:  option.SomeString(leftDatabaseName),
				Qualifier: option.SomeString(leftTableName),
				Name:      colName,
			}
			rightExpr := &parser.ColName{
				Database:  option.SomeString(rightDatabaseName),
				Qualifier: option.SomeString(rightTableName),
				Name:      colName,
			}
			comparison := &parser.ComparisonExpr{
				Operator: parser.AST_EQ,
				Left:     leftExpr,
				Right:    rightExpr,
			}
			// add to final expression
			if expression == nil {
				expression = comparison
			} else {
				expression = &parser.AndExpr{
					Left:  expression,
					Right: comparison,
				}
			}
		}
	}
	return expression, nil
}

type columnsUsing struct {
	columns     []*results.Column
	usingCols   parser.ColumnExprs
	kind        JoinKind
	rightTables map[string]struct{}
}

func (c columnsUsing) Len() int {
	return len(c.columns)
}

// Less reorders the columns in a JoinTableExpr when the table expr has a USING
// clause. It places columns in the USING clause before those not in USING,
// and when the join type is a right join, it puts columns from the right
// expression before the columns from the left expression.
func (c columnsUsing) Less(i, j int) bool {
	var iInUsing, jInUsing bool
	for _, column := range c.usingCols {
		if column.Name == c.columns[i].Name {
			iInUsing = true
		}
		if column.Name == c.columns[j].Name {
			jInUsing = true
		}
	}
	if iInUsing && !jInUsing {
		return true
	} else if !iInUsing && jInUsing {
		return false
	} else {
		_, iInRight := c.rightTables[fullyQualifiedTableName(c.columns[i].Database,
			c.columns[i].Table)]
		_, jInRight := c.rightTables[fullyQualifiedTableName(c.columns[j].Database,
			c.columns[j].Table)]
		if iInRight && !jInRight {
			return (c.kind == RightJoin || c.kind == NaturalRightJoin)
		} else if jInRight && !iInRight {
			return !(c.kind == RightJoin || c.kind == NaturalRightJoin)
		}
	}
	return false
}

func (c columnsUsing) Swap(i, j int) {
	c.columns[i], c.columns[j] = c.columns[j], c.columns[i]
}

func (c columnsUsing) Sort() {
	sort.Stable(c)
}

func (c columnsUsing) Filter() []*results.Column {
	// keep track of the USING columns to ensure we only project one for each column in the clause
	seenColumns := make(map[string]bool)
	for _, usingColumn := range c.usingCols {
		seenColumns[usingColumn.Name] = false
	}
	var columnsForProjection []*results.Column
	for _, column := range c.columns {
		// if column is a USING column, add it only if it's the first instance
		if seenBefore, ok := seenColumns[column.Name]; ok {
			if seenBefore {
				continue
			}
			seenColumns[column.Name] = true
		}
		columnsForProjection = append(columnsForProjection, column)
	}
	return columnsForProjection
}

func resolveJoinKind(kind JoinKind, hasGlobalStraightJoin bool) JoinKind {
	if (kind == InnerJoin || kind == NaturalJoin || kind == CrossJoin) && hasGlobalStraightJoin {
		return StraightJoin
	} else if kind == NaturalJoin {
		return InnerJoin
	} else if kind == NaturalLeftJoin {
		return LeftJoin
	} else if kind == NaturalRightJoin {
		return RightJoin
	}
	return kind
}

func optimizeJoinKind(kind JoinKind, onClause parser.Expr, filterCols parser.ColumnExprs) JoinKind {
	hasCriteria := (filterCols != nil || onClause != nil)
	if kind == CrossJoin && hasCriteria {
		return InnerJoin
	} else if (kind == InnerJoin || kind == RightJoin || kind == LeftJoin) && !hasCriteria {
		return CrossJoin
	}
	return kind
}

func (a *algebrizer) translateTableExpr(tableExpr parser.TableExpr, hasGlobalStraightJoin bool) (PlanStage, []*results.Column, error) {
	switch typedT := tableExpr.(type) {
	case *parser.AliasedTableExpr:
		return a.translateSimpleTableExpr(typedT.Expr, typedT.As.Else(""))
	case *parser.DualTableExpr:
		return a.translateSimpleTableExpr(typedT, "")
	case *parser.ParenTableExpr:
		return a.translateTableExpr(typedT.Expr, hasGlobalStraightJoin)
	case parser.SimpleTableExpr:
		return a.translateSimpleTableExpr(typedT, "")
	case *parser.JoinTableExpr:
		kind := JoinKind(typedT.Join)
		switch kind {
		case NaturalJoin, NaturalRightJoin, NaturalLeftJoin:
			if !(typedT.On == nil && typedT.Using == nil) {
				return nil, nil, mysqlerrors.Newf(mysqlerrors.ErParseError,
					"A %s cannot have join criteria",
					typedT.Join)
			}

		}

		left, leftCols, err := a.translateTableExpr(typedT.LeftExpr, hasGlobalStraightJoin)
		if err != nil {
			return nil, nil, err
		}
		right, rightCols, err := a.translateTableExpr(typedT.RightExpr, hasGlobalStraightJoin)
		if err != nil {
			return nil, nil, err
		}

		// Ensure that any operations on cols don't affect leftCols or rightCols.
		cols := make([]*results.Column, 0, len(leftCols)+len(rightCols))
		cols = append(cols, leftCols...)
		cols = append(cols, rightCols...)

		var predicate SQLExpr = NewSQLValueExpr(values.NewSQLBool(a.valueKind(), true))
		var filterCols parser.ColumnExprs
		if typedT.On != nil {
			predicate, err = a.translateExpr(typedT.On)
			if err != nil {
				return nil, nil, err
			}
		} else if typedT.Using != nil || kind == NaturalJoin ||
			kind == NaturalLeftJoin || kind == NaturalRightJoin {
			if typedT.Using != nil {
				filterCols = typedT.Using
			} else {
				// construct a list of all the common columns
				// in the natural join case

				// this ensures the columns values are pulled
				// from the correct database when appending
				// rows that did not have an exact match from
				// both dbs on common columns
				filterDb := leftCols[0].Database
				if kind == NaturalRightJoin {
					filterDb = rightCols[0].Database
				}

				rightColMap := make(map[string]struct{})
				var emptyStruct struct{}
				for _, rightCol := range rightCols {
					rightColMap[rightCol.Name] = emptyStruct
				}
				for _, leftCol := range leftCols {
					if _, ok := rightColMap[leftCol.Name]; ok {
						colName := &parser.ColName{
							Database:  option.SomeString(filterDb),
							Qualifier: option.NoneString(),
							Name:      leftCol.Name,
						}
						filterCols = append(filterCols, colName)
					}
				}
			}

			comparison, err := a.convertToAnd(filterCols, leftCols, rightCols, typedT)
			if err != nil {
				return nil, nil, err
			}

			if comparison == nil {
				predicate = NewSQLValueExpr(values.NewSQLBool(a.valueKind(), true))
			} else {
				predicate, err = a.translateExpr(comparison)
				if err != nil {
					return nil, nil, err
				}
			}

			if filterCols != nil {
				rightTables := make(map[string]struct{})
				var emptyStruct struct{}
				for _, column := range rightCols {
					rightTables[fullyQualifiedTableName(column.Database,
						column.Table)] = emptyStruct
				}

				sortableFilterableColumns := &columnsUsing{cols, filterCols, kind, rightTables}
				sortableFilterableColumns.Sort()
				cols = sortableFilterableColumns.Filter()
			}
		} else if kind == LeftJoin || kind == RightJoin {
			return nil, nil, mysqlerrors.Newf(mysqlerrors.ErParseError,
				"A %s requires criteria", typedT.Join)
		}

		kind = resolveJoinKind(kind, hasGlobalStraightJoin)
		kind = optimizeJoinKind(kind, typedT.On, filterCols)
		return NewJoinStage(kind, left, right, predicate), cols, nil
	default:
		return nil, nil, mysqlerrors.Defaultf(mysqlerrors.ErNotSupportedYet,
			parser.String(tableExpr))
	}
}

func (a *algebrizer) translateSimpleTableExpr(
	tableExpr parser.SimpleTableExpr, aliasName string) (PlanStage, []*results.Column, error) {
	switch typedT := tableExpr.(type) {
	case *parser.TableName:
		tableName := strings.ToLower(typedT.Name)
		if aliasName == "" {
			aliasName = tableName
		}

		var plan PlanStage
		var err error

		cteEvaluator := a.ctes[tableName]
		// CTEs are not part of the database's namespace, so if database Qualifier
		// is present this TableExpr is definitely not referring to a CTE.
		// For example, with cte as (select * from tbl) select * from db.cte;
		// will not reference the CTE created at the start of the query.
		if typedT.Qualifier.IsNone() && cteEvaluator != nil {
			cte := cteEvaluator.cte
			subqueryAlgebrizer := cteEvaluator.algebrizer
			plan = cteEvaluator.planStage.clone()

			dbName := databaseFromPlanStage(plan)
			plan = NewSubquerySourceStage(plan, subqueryAlgebrizer.selectID,
				dbName, aliasName, true)
			err = a.registerTable(dbName, aliasName)
			if err != nil {
				return nil, nil, err
			}

			columns := plan.Columns()
			if cte.ColumnExprs != nil {
				if len(cte.ColumnExprs) != len(columns) {
					// It's a confusing error message but that is MySQL's behavior.
					return nil, nil, mysqlerrors.Defaultf(mysqlerrors.ErViewWrongList)
				}
				a.projectedColumns = make(ProjectedColumns, len(cte.ColumnExprs))
				for i, expr := range cte.ColumnExprs {
					a.projectedColumns[i] = newProjectedColumnFromColumnWithName(columns[i], expr.Name)
				}
			}
			err = a.registerColumns(columns)
			if err != nil {
				return nil, nil, err
			}

			return plan, columns, nil
		}

		// Reach here if the table name in the from is not declared in a CTE.
		dbName := typedT.Qualifier.Else("")
		if dbName == "" {
			dbName = a.cfg.dbName
		}
		dbName = strings.ToLower(dbName)

		db, dbErr := a.cfg.catalog.Database(a.ctx, dbName)
		if dbErr != nil {
			return nil, nil, dbErr
		}

		table, dbErr := db.Table(a.ctx, tableName)
		if dbErr != nil {
			return nil, nil, dbErr
		}

		switch t := table.(type) {
		case catalog.MongoDBTable:
			plan = NewMongoSourceStage(db.Name(), t, a.selectID, aliasName)
		case *catalog.DynamicTable:
			plan = NewDynamicSourceStage(db, t, a.selectID, aliasName)
		default:
			return nil, nil, fmt.Errorf("unknown table type: %T", t)
		}
		err = a.registerTable(dbName, aliasName)

		if err != nil {
			return nil, nil, err
		}

		columns := plan.Columns()
		if !strings.EqualFold(tableName, "DUAL") {
			err = a.registerColumns(columns)
			if err != nil {
				return nil, nil, err
			}
		}

		return plan, columns, nil
	case *parser.Subquery:

		if aliasName == "" && typedT.IsDerived {
			return nil, nil, mysqlerrors.Defaultf(mysqlerrors.ErDerivedMustHaveAlias)
		}

		subqueryAlgebrizer := a.newDerivedTableAlgebrizer()

		plan, err := subqueryAlgebrizer.translateSelectStatement(typedT.Select)
		if err != nil {
			return nil, nil, err
		}

		// We need fully qualified table names for our optimizers so we propagate the database name
		// here if the subquery doesn't contain columns from multiple databases (e.g. with cross
		// database joins) and the empty string otherwise.
		dbName := databaseFromPlanStage(plan)

		plan = NewSubquerySourceStage(plan, subqueryAlgebrizer.selectID, dbName, aliasName, false)

		// database is not set here because duplicate tables are not allowed in a select query
		// ignoring it avoids false positives.
		err = a.registerTable(dbName, aliasName)
		if err != nil {
			return nil, nil, err
		}

		columns := plan.Columns()
		err = a.registerColumns(columns)
		if err != nil {
			return nil, nil, err
		}

		return plan, columns, nil
	case *parser.DualTableExpr:
		plan, err := a.newMongoSourceOrDualStage()
		if err != nil {
			return nil, nil, err
		}

		columns := plan.Columns()
		return plan, columns, nil
	default:
		return nil,
			nil,
			mysqlerrors.Defaultf(mysqlerrors.ErNotSupportedYet,
				parser.String(tableExpr))
	}
}

func (a *algebrizer) translateSubqueryPlan(expr *parser.Subquery) (PlanStage, bool, error) {
	subqueryAlgebrizer := a.newSubqueryExprAlgebrizer()

	plan, err := subqueryAlgebrizer.translateSelectStatement(expr.Select)
	if err != nil {
		return nil, false, err
	}

	return plan, subqueryAlgebrizer.correlated, nil
}

func (a *algebrizer) translateSubqueryExpr(expr *parser.Subquery) (*SQLSubqueryExpr, error) {
	plan, correlated, err := a.translateSubqueryPlan(expr)
	if err != nil {
		return nil, err
	}

	// SQLSubqueryExprs are only valid if they return a single value. As such,
	// we know immediately that there is an error if we have more than one
	// column in the plan.
	numCols := len(plan.Columns())
	if numCols != 1 {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, 1)
	}

	return NewSQLSubqueryExpr(correlated, false, plan), nil
}

func (a *algebrizer) translateExistsExpr(expr *parser.ExistsExpr) (*SQLExistsExpr, error) {
	subqueryAlgebrizer := a.newSubqueryExprAlgebrizer()

	plan, err := subqueryAlgebrizer.translateSelectStatement(expr.Subquery.Select)
	if err != nil {
		return nil, err
	}

	return &SQLExistsExpr{
		plan:       plan,
		correlated: subqueryAlgebrizer.correlated,
	}, nil
}

// This includes both comparisons with a quantifying keyword such as ANY or ALL
// and those without.
// Note: this does not handle EXISTS or subqueries resulting in a
// single-column scalar.
func (a *algebrizer) translateSubqueryCmpExpr(expr *parser.ComparisonExpr) (SQLExpr, error) {
	// Both the left and right Exprs must be subqueries at this point.
	lsq, leftIsPlan := expr.Left.(*parser.Subquery)
	rsq, rightIsPlan := expr.Right.(*parser.Subquery)

	// If one is not, our desugarer did not run or has become buggy, panic accordingly.
	if !leftIsPlan {
		panic(fmt.Sprintf("expected parser.Subquery for left side of subquery cmp expr, but got %T", expr.Left))
	}
	if !rightIsPlan {
		panic(fmt.Sprintf("expected parser.Subquery for right side of subquery cmp expr, but got %T", expr.Right))
	}

	leftPlan, leftIsCorrelated, err := a.translateSubqueryPlan(lsq)
	if err != nil {
		return nil, err
	}

	rightPlan, rightIsCorrelated, err := a.translateSubqueryPlan(rsq)
	if err != nil {
		return nil, err
	}

	// Make sure the right and left sides of the comparison have the same number of columns.
	leftProject, ok := leftPlan.(*ProjectStage)
	if !ok {
		panic(fmt.Sprintf("expected ProjectStage for left PlanStage, got %T", leftPlan))
	}
	rightProject, ok := rightPlan.(*ProjectStage)
	if !ok {
		panic(fmt.Sprintf("expected ProjectStage for right PlanStage, got %T", rightPlan))
	}

	if len(leftProject.projectedColumns) != len(rightProject.projectedColumns) {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns,
			len(rightProject.projectedColumns))
	}

	switch expr.SubqueryOperator {
	case "":
		return NewSQLSubqueryCmpExpr(
			leftIsCorrelated, rightIsCorrelated,
			leftPlan, rightPlan,
			expr.Operator,
		), nil

	case parser.AST_ANY:
		return NewSQLSubqueryAnyExpr(
			leftIsCorrelated,
			rightIsCorrelated,
			leftPlan,
			rightPlan,
			expr.Operator,
		), nil

	case parser.AST_ALL:
		return NewSQLSubqueryAllExpr(
			leftIsCorrelated,
			rightIsCorrelated,
			leftPlan,
			rightPlan,
			expr.Operator,
		), nil
	}

	// if we get here, we made an error invoking this function
	panic("neither left nor right side of a subquery comparison expr was a subquery")
}

func (a *algebrizer) translateExpr(expr parser.Expr) (SQLExpr, error) {
	switch typedE := expr.(type) {
	case *parser.AndExpr:

		left, err := a.translateExpr(typedE.Left)
		if err != nil {
			return nil, err
		}

		right, err := a.translateExpr(typedE.Right)
		if err != nil {
			return nil, err
		}

		return NewSQLAndExpr(left, right), nil
	case *parser.BinaryExpr:
		left, err := a.translateExpr(typedE.Left)
		if err != nil {
			return nil, err
		}

		right, err := a.translateExpr(typedE.Right)
		if err != nil {
			return nil, err
		}

		switch typedE.Operator {
		case parser.AST_PLUS:
			return NewSQLAddExpr(left, right), nil
		case parser.AST_MINUS:
			return NewSQLSubtractExpr(left, right), nil
		case parser.AST_MULT:
			return NewSQLMultiplyExpr(left, right), nil
		case parser.AST_DIV:
			return NewSQLDivideExpr(left, right), nil
		case parser.AST_IDIV:
			return NewSQLIDivideExpr(left, right), nil
		case parser.AST_MOD:
			return NewSQLModExpr(left, right), nil
		default:
			return nil, mysqlerrors.Newf(mysqlerrors.ErNotSupportedYet,
				"No support for binary operator '%v'", typedE.Operator)
		}
	case *parser.CaseExpr:
		return a.translateCaseExpr(typedE)
	case *parser.ColName:
		databaseName := typedE.Database.Else("")
		tableName := typedE.Qualifier.Else("")
		columnName := typedE.Name

		if strings.HasPrefix(tableName, "@") || (tableName == "" && strings.HasPrefix(columnName, "@")) {
			return a.translateVariableExpr(typedE)
		}

		return a.resolveColumnExpr(databaseName, tableName, columnName)
	case *parser.ComparisonExpr:
		_, leftIs := typedE.Left.(*parser.Subquery)
		_, rightIs := typedE.Right.(*parser.Subquery)
		if leftIs || rightIs {
			return a.translateSubqueryCmpExpr(typedE)
		}

		left, err := a.translateExpr(typedE.Left)
		if err != nil {
			return nil, err
		}

		right, err := a.translateExpr(typedE.Right)
		if err != nil {
			return nil, err
		}

		if typedE.Operator == parser.AST_IS {
			return NewSQLIsExpr(left, right), nil
		} else if typedE.Operator == parser.AST_IS_NOT {
			panic("IS NOT must be eliminated in the desugarer")
		}

		comp, err := comparisonExpr(left, right, typedE.Operator)
		if err != nil {
			return nil, err
		}

		return comp, err

	case *parser.DateVal:

		arg := typedE.Val

		switch typedE.Name {
		case parser.AST_DATE:
			date, _, ok := values.ParseDateTime(arg)
			if !ok || date.Hour() > 0 ||
				date.Minute() > 0 ||
				date.Second() > 0 ||
				date.Nanosecond() > 0 {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongValue, "DATE", arg)
			}
			return NewSQLValueExpr(values.NewSQLDate(a.valueKind(), date)), nil
		case parser.AST_TIME:
			dur, _, ok := strToTime(arg)
			if !ok {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongValue, "TIME", arg)
			}

			date := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC).Add(dur)

			return NewSQLValueExpr(values.NewSQLTimestamp(a.valueKind(), date)), nil
		case parser.AST_TIMESTAMP, parser.AST_DATETIME:
			date, _, ok := values.StrToDateTime(arg, true)
			if !ok {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongValue, "DATETIME", arg)
			}
			return NewSQLValueExpr(values.NewSQLTimestamp(a.valueKind(), date)), nil
		default:
			return nil, mysqlerrors.Newf(mysqlerrors.ErNotSupportedYet,
				"No support for constructor '%v'", typedE.Name)
		}
	case *parser.ExistsExpr:
		return a.translateExistsExpr(typedE)
	case *parser.FalseVal:
		return NewSQLValueExpr(values.NewSQLBool(a.valueKind(), false)), nil
	case *parser.FuncExpr:
		return a.translateFuncExpr(typedE)
	case parser.KeywordVal:
		return NewSQLValueExpr(values.NewSQLVarchar(a.valueKind(), string(typedE))), nil
	case *parser.LikeExpr:
		left, err := a.translateExpr(typedE.Left)
		if err != nil {
			return nil, err
		}

		right, err := a.translateExpr(typedE.Right)
		if err != nil {
			return nil, err
		}

		var escape SQLExpr
		if typedE.Escape != nil {
			escape, err = a.translateExpr(typedE.Escape)
			if err != nil {
				return nil, err
			}
			// At this point, we can check that if the escape parameter is a scalar string,
			// it must have a length of 1 to be valid. Nonscalar types will result in
			// pushdown failures and then will be checked for validity in SQLLikeExpr's implementation
			// of Evaluate.
			if escValue, ok := escape.(SQLValueExpr); ok {
				if len(escValue.String()) > 1 {
					return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongArguments, "ESCAPE")
				}
			}
		} else {
			escape = NewSQLValueExpr(values.NewSQLVarchar(a.valueKind(), "\\"))
		}

		caseSensitive := typedE.Operator == parser.AST_NOT_LIKE_BINARY ||
			typedE.Operator == parser.AST_LIKE_BINARY

		expr := NewSQLLikeExpr(left, right, escape, caseSensitive)

		if typedE.Operator == parser.AST_NOT_LIKE ||
			typedE.Operator == parser.AST_NOT_LIKE_BINARY {
			return NewSQLNotExpr(expr), nil
		}

		return expr, nil
	case *parser.NotExpr:
		child, err := a.translateExpr(typedE.Expr)
		if err != nil {
			return nil, err
		}

		return NewSQLNotExpr(child), nil
	case *parser.NullVal:
		return NewSQLValueExpr(values.NewSQLNull(a.valueKind())), nil
	case parser.NumVal:
		exprString := parser.String(expr)

		useFloats := !a.versionAtLeast(3, 3, 15)

		// http://dev.mysql.com/doc/refman/5.7/en/precision-math-numbers.html
		// Because MongoDB 3.2 does not support decimals, we are going
		// to override any decimal literal and parse it as a float when
		// connected to < MongoDB 3.4 AND the user hasn't informed us
		// that we should force the use of decimals against 3.2.
		if strings.ContainsAny(exprString, "Ee") ||
			(useFloats && strings.Contains(exprString, ".")) {
			f, err := strconv.ParseFloat(exprString, 64)
			if err != nil {
				return nil,
					mysqlerrors.Defaultf(mysqlerrors.ErIllegalValueForType,
						"double",
						exprString)
			}
			return NewSQLValueExpr(values.NewSQLFloat(a.valueKind(), f)), nil
		}
		if strings.Contains(exprString, ".") {
			d, err := decimal.NewFromString(exprString)
			if err != nil {
				return nil,
					mysqlerrors.Defaultf(mysqlerrors.ErIllegalValueForType,
						"decimal",
						exprString)

			}
			return NewSQLValueExpr(values.NewSQLDecimal128(a.valueKind(), d)), nil
		}

		// try to parse as int64 first
		if i, err := strconv.ParseInt(exprString, 10, 64); err == nil {
			return NewSQLValueExpr(values.NewSQLInt64(a.valueKind(), i)), nil
		}

		// next try to parse as uint64
		if i, err := strconv.ParseUint(exprString, 10, 64); err == nil {
			return NewSQLValueExpr(values.NewSQLUint64(a.valueKind(), i)), nil
		}

		if useFloats {
			f, err := strconv.ParseFloat(exprString, 64)
			if err != nil {
				return nil,
					mysqlerrors.Defaultf(mysqlerrors.ErIllegalValueForType,
						"integer",
						exprString)
			}
			return NewSQLValueExpr(values.NewSQLFloat(a.valueKind(), f)), nil
		}

		i, err := decimal.NewFromString(exprString)
		if err != nil {
			return nil,
				mysqlerrors.Defaultf(mysqlerrors.ErIllegalValueForType,
					"integer",
					exprString)
		}
		return NewSQLValueExpr(values.NewSQLDecimal128(a.valueKind(), i)), nil
	case *parser.OrExpr:

		left, err := a.translateExpr(typedE.Left)
		if err != nil {
			return nil, err
		}

		right, err := a.translateExpr(typedE.Right)
		if err != nil {
			return nil, err
		}

		return NewSQLOrExpr(left, right), nil
	case *parser.XorExpr:

		left, err := a.translateExpr(typedE.Left)
		if err != nil {
			return nil, err
		}

		right, err := a.translateExpr(typedE.Right)
		if err != nil {
			return nil, err
		}

		return NewSQLXorExpr(left, right), nil
	case *parser.RegexExpr:
		operand, err := a.translateExpr(typedE.Operand)
		if err != nil {
			return nil, err
		}

		pattern, err := a.translateExpr(typedE.Pattern)
		if err != nil {
			return nil, err
		}

		// BINARY Sting support will be added in BI-2327
		return NewSQLRegexExpr(operand, pattern), nil
	case *parser.RLikeExpr:
		operand, err := a.translateExpr(typedE.Operand)
		if err != nil {
			return nil, err
		}

		pattern, err := a.translateExpr(typedE.Pattern)
		if err != nil {
			return nil, err
		}
		return NewSQLRegexExpr(operand, pattern), nil
	case parser.StrVal:
		return NewSQLValueExpr(values.NewSQLVarchar(a.valueKind(), string(typedE))), nil
	case *parser.Subquery:
		return a.translateSubqueryExpr(typedE)
	case *parser.TrueVal:
		return NewSQLValueExpr(values.NewSQLBool(a.valueKind(), true)), nil
	case *parser.UnaryExpr:

		child, err := a.translateExpr(typedE.Expr)
		if err != nil {
			return nil, err
		}

		switch typedE.Operator {
		case parser.AST_UMINUS:
			return NewSQLUnaryMinusExpr(child), nil
		case parser.AST_TILDA:
			return NewSQLTildeExpr(child), nil
		case parser.AST_UPLUS:
			return child, nil
		}

		return nil, mysqlerrors.Newf(mysqlerrors.ErNotSupportedYet,
			"No support for operator '%v'", typedE.Operator)
	case *parser.UnknownVal:
		return NewSQLValueExpr(values.NewSQLNull(a.valueKind())), nil
	default:
		return nil, mysqlerrors.Newf(mysqlerrors.ErNotSupportedYet,
			"No support for '%v':%T", parser.String(typedE), typedE)
	}
}

func (a *algebrizer) translatePossibleColumnRefExpr(expr parser.Expr) (SQLExpr, error) {
	if numVal, ok := expr.(parser.NumVal); ok {
		n, err := strconv.ParseInt(parser.String(numVal), 10, 64)
		if err != nil {
			return nil, err
		}

		if n == 0 || int(n) > len(a.projectedColumns) {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErBadFieldError,
				strconv.Itoa(int(n)), a.currentClause)
		}

		if n > 0 {
			if a.currentClause == groupClause {
				if agg, ok := a.projectedColumnAggregateMap[int(n)]; ok {
					return nil,
						mysqlerrors.Defaultf(mysqlerrors.ErWrongGroupField,
							agg.String())
				}
			}
			return a.projectedColumns[n-1].Expr, nil
		}
	}

	return a.translateExpr(expr)
}

func (a *algebrizer) translateCaseExpr(expr *parser.CaseExpr) (SQLExpr, error) {
	// There are two kinds of case expression.
	//
	// 1. For simple case expressions, we create an equality matcher that compares
	// the expression against each value in the list of cases.
	//
	// 2. For searched case expressions, we create a matcher based on the boolean
	// expression in each when condition.

	var e SQLExpr
	var err error

	if expr.Expr != nil {
		e, err = a.translateExpr(expr.Expr)
		if err != nil {
			return nil, err
		}
	}

	var conditions []caseCondition
	var matcher SQLExpr

	for _, when := range expr.Whens {

		// searched case
		if expr.Expr == nil {
			matcher, err = a.translateExpr(when.Cond)
			if err != nil {
				return nil, err
			}
		} else {
			// TODO: support simple case in parser
			var cond SQLExpr
			cond, err = a.translateExpr(when.Cond)
			if err != nil {
				return nil, err
			}

			matcher = NewSQLComparisonExpr(EQ, e, cond)
		}

		var then SQLExpr
		then, err = a.translateExpr(when.Val)
		if err != nil {
			return nil, err
		}

		conditions = append(conditions, caseCondition{matcher, then})
	}

	var elseValue SQLExpr
	if expr.Else == nil {
		elseValue = NewSQLValueExpr(values.NewSQLNull(a.valueKind()))
	} else if elseValue, err = a.translateExpr(expr.Else); err != nil {
		return nil, err
	}

	value := &SQLCaseExpr{
		elseValue:      elseValue,
		caseConditions: conditions,
	}

	// TODO: You cannot specify the literal NULL for every return expr
	// and the else expr.
	return value, nil
}

// NewSQLAggregationFunctionExpr builds a new SQLAggregationFunctionExpr.
func NewSQLAggregationFunctionExpr(name string, distinct bool, exprs []SQLExpr) SQLAggFunctionExpr {
	switch name {
	case parser.AvgAggregateName:
		return NewSQLAvgFunctionExpr(distinct, exprs)
	case parser.SumAggregateName:
		return NewSQLSumFunctionExpr(distinct, exprs)
	case parser.CountAggregateName:
		return NewSQLCountFunctionExpr(distinct, exprs)
	case parser.GroupConcatAggregateName:
		return NewSQLGroupConcatFunctionExpr(distinct, exprs, option.NoneString(), 0)
	case parser.MaxAggregateName:
		return NewSQLMaxFunctionExpr(distinct, exprs)
	case parser.MinAggregateName:
		return NewSQLMinFunctionExpr(distinct, exprs)
	case parser.StdAggregateName, parser.StdDevAggregateName, parser.StdDevPopAggregateName:
		return NewSQLStdDevFunctionExpr(name, distinct, exprs)
	case parser.StdDevSampleAggregateName:
		return NewSQLStdDevSampleFunctionExpr(distinct, exprs)
	default:
		panic(fmt.Errorf("aggregate function '%v' is not supported", name))
	}
}

func (a *algebrizer) translateFuncExpr(expr *parser.FuncExpr) (SQLExpr, error) {

	exprs := []SQLExpr{}
	name := expr.Name

	if a.isAggFunction(name) {
		if len(expr.Exprs) != 1 && name != "group_concat" {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongArguments, name)
		}

		for _, e := range expr.Exprs {
			switch typedE := e.(type) {
			case *parser.StarExpr:

				if name != "count" {
					return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongArguments, name)
				}

				if expr.Distinct {
					return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongArguments, name)
				}

				exprs = append(exprs, NewSQLValueExpr(values.NewSQLVarchar(a.valueKind(), "*")))

			case *parser.NonStarExpr:

				sqlExpr, err := a.translateExpr(typedE.Expr)
				if err != nil {
					return nil, err
				}
				exprs = append(exprs, sqlExpr)
			default:
				return nil, mysqlerrors.Defaultf(mysqlerrors.ErNotSupportedYet,
					parser.String(e))
			}
		}

		aggExpr := NewSQLAggregationFunctionExpr(name, expr.Distinct, exprs)
		if groupConcat, ok := aggExpr.(*SQLGroupConcatFunctionExpr); ok {
			if expr.Separator.IsSome() {
				groupConcat.Separator = expr.Separator
			} else {
				groupConcat.Separator = option.SomeString(",")
			}
			groupConcat.GroupConcatMaxLen = int(a.cfg.groupConcatMaxLen)
		}

		// We are going to replace the aggregate with a column in the
		// tree and put the aggregate into the algebrizer (which could
		// be us or any of our parents) that is supposed to do the
		// aggregation. We determine this by seeing if any of the
		// columns referenced inside the aggregate pertain to the
		// algebrizer.

		// figure out which "select" should be responsible for
		// aggregating this guy.
		usedSelectIDs := gatherSelectIDs(aggExpr)

		current := a
		if len(usedSelectIDs) > 0 {
			// we must exist somewhere, because we were able to
			// resolve any columns references. As such, we are
			// going to walk up the tree, starting with us, to
			// figure out which algebrizer we belong to. Then drop
			// in a SQLColumnExpr here that references the
			// aggregate.
			for current != nil {
				if containsAnyInt(current.currentSelectIDs, usedSelectIDs) {
					break
				}
				current = current.parent
			}
		}

		if current == nil {
			// If we ever get here, this is a bug somewhere. It
			// means we had created a SQLColumnExpr but associated
			// an invalid selectID with it.
			return nil, mysqlerrors.Unknownf("Aggregate doesn't include" +
				" any relevant columns")
		}

		if current.currentClause == whereClause {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErInvalidGroupFuncUse)
		} else if current.currentClause == groupClause {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongGroupField, aggExpr.String())
		}

		// If the algebrizer responsible for aggregation (current) is not the
		// same as the algebrizer the aggregation originally comes from (a),
		// then the replacement column must be correlated. That is because the
		// aggregation function will be added to current.aggregates and current
		// is a parent of this algebrizer, a.
		correlated := a != current
		col := NewSQLColumnExpr(current.selectID, getDatabaseName(aggExpr), "", aggExpr.String(),
			aggExpr.EvalType(), schema.MongoNone, correlated, true)

		var err error
		if correlated {
			// If the new column is correlated, mark this algebrizer as correlated.
			a.correlated = true

			// Since the aggExpr is being moved to a parent, it must be "decorrelated".
			// As in, any SQLColumnExprs in the aggExpr that were previously marked as
			// correlated should be marked as _not_ correlated now that the aggExpr is
			// being moved to the parent where the correlation source comes from.
			aggExpr, err = decorrelateAggFunctionExpr(aggExpr)
			if err != nil {
				return nil, err
			}
		}

		current.aggregates = append(current.aggregates, aggExpr)

		return col, nil
	}

	for _, e := range expr.Exprs {

		switch typedE := e.(type) {
		case *parser.StarExpr:
			if !strings.EqualFold(name, "count") {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongArguments, name)
			}
		case *parser.NonStarExpr:
			sqlExpr, err := a.translateExpr(typedE.Expr)
			if err != nil {
				return nil, err
			}

			exprs = append(exprs, sqlExpr)

			if typedE.As.IsSome() {
				as := typedE.As.Unwrap()
				switch strings.ToLower(as) {
				case "cast":
					exprs = append(exprs, NewSQLValueExpr(values.NewSQLVarchar(a.valueKind(), as)))
				default:
					return nil,
						mysqlerrors.Defaultf(mysqlerrors.ErNotSupportedYet,
							parser.String(e))
				}
			}
		default:
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErNotSupportedYet, parser.String(expr))
		}

	}

	// Handle any special translations for specific scalar functions here:
	switch name {
	case "benchmark":
		if len(exprs) != 2 {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongArguments, name)
		}
		return &SQLBenchmarkExpr{exprs[0], exprs[1]}, nil
	case "date":
		return NewSQLConvertExpr(exprs[0], types.EvalDate), nil
	case "timestamp":
		if len(exprs) == 1 {
			return NewSQLConvertExpr(exprs[0], types.EvalDatetime), nil
		}
		return NewSQLScalarFunctionExpr(name, exprs)
	case "rand":
		// We need something unique that we can map.
		id := a.getUniqueID()
		return NewSQLScalarFunctionExpr(
			"rand",
			append([]SQLExpr{NewSQLValueExpr(values.NewSQLUint64(a.valueKind(), id))}, exprs...),
		)
	case "isnull":
		return NewSQLIsExpr(exprs[0], NewSQLValueExpr(values.NewSQLNull(a.valueKind()))), nil
	case "date_add", "adddate", "date_sub", "subdate":
		if len(exprs) == 2 {
			return NewSQLScalarFunctionExpr(
				name,
				[]SQLExpr{exprs[0], exprs[1], NewSQLValueExpr(values.NewSQLVarchar(a.valueKind(), Day))},
			)
		}
		return NewSQLScalarFunctionExpr(name, []SQLExpr{exprs[0], exprs[1], exprs[2]})
	case "week", "weekofyear":
		if len(exprs) == 2 {
			return NewSQLScalarFunctionExpr("week", []SQLExpr{exprs[0], exprs[1]})
		}
		if name == "week" {
			return NewSQLScalarFunctionExpr(
				"week",
				[]SQLExpr{exprs[0], NewSQLValueExpr(values.NewSQLInt64(a.valueKind(), 0))},
			)
		}
		return NewSQLScalarFunctionExpr("week", []SQLExpr{exprs[0], NewSQLValueExpr(values.NewSQLInt64(a.valueKind(), 3))})
	case "yearweek":
		if len(exprs) == 2 {
			return NewSQLScalarFunctionExpr("yearweek", []SQLExpr{exprs[0], exprs[1]})
		}
		return NewSQLScalarFunctionExpr(
			"yearweek",
			[]SQLExpr{exprs[0], NewSQLValueExpr(values.NewSQLInt64(a.valueKind(), 0))},
		)
	default:
		return NewSQLScalarFunctionExpr(name, exprs)
	}

}

func (a *algebrizer) translateVariableExpr(c *parser.ColName) (*SQLVariableExpr, error) {

	kind := variable.SystemKind
	scope := variable.SessionScope

	pos := 0
	str := c.Name
	if c.Qualifier.IsSome() {
		str = c.Qualifier.Unwrap() + "." + str
	}

	if str[pos] == '@' {
		pos++
		if len(str) > 1 && str[pos] != '@' {
			kind = variable.UserKind
		} else {
			pos++
		}
	}

	name := str[pos:]

	if kind != variable.UserKind {
		idx := strings.Index(name, ".")
		if idx >= 0 {
			switch strings.ToLower(name[:idx+1]) {
			case "global.":
				name = name[idx+1:]
				scope = variable.GlobalScope
			case "session.", "local.":
				name = name[idx+1:]
			}
		}
	}

	value, err := a.cfg.variables.Get(variable.Name(name), scope, kind)
	if err != nil {
		return nil, err
	}

	return NewSQLVariableExpr(name, kind, scope, value), nil
}

func (a *algebrizer) versionAtLeast(major, minor, patch uint8) bool {
	return procutil.VersionAtLeast(a.cfg.version, []uint8{major, minor, patch})
}

// aggFuncColumnDecorrelator is used to traverse a SQLAggFunctionExpr and
// mark any SQLColumnExprs as non-correlated.
type aggFuncColumnDecorrelator struct{}

func (v *aggFuncColumnDecorrelator) visit(n Node) (Node, error) {
	switch typedN := n.(type) {
	case SQLColumnExpr:
		if typedN.correlated {
			typedN.correlated = false
			return typedN, nil
		}
	case *SQLSubqueryExpr, *SQLSubqueryCmpExpr, *SQLSubqueryAnyExpr, *SQLSubqueryAllExpr:
		return n, nil
	}
	return walk(v, n)
}

// decorrelateAggFunctionExpr walks a SQLAggFunctionExpr and marks any
// SQLColumnExprs as non-correlated. SQLColumnExprs that are nested in
// subquery exprs will not be modified.
func decorrelateAggFunctionExpr(expr SQLAggFunctionExpr) (SQLAggFunctionExpr, error) {
	decorrelator := &aggFuncColumnDecorrelator{}
	newExpr, err := walk(decorrelator, expr)
	if err != nil {
		return nil, err
	}

	return newExpr.(SQLAggFunctionExpr), nil
}
