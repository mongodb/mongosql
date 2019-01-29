package evaluator

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/10gen/sqlproxy/evaluator/catalog"
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
	dbName                        string
	groupConcatMaxLen             int64
	isMongos                      bool
	lg                            log.Logger
	sqlValueKind                  SQLValueKind
	sqlSelectLimit                uint64
	maxVarcharLength              uint16
	polymorphicTypeConversionMode variable.PolymorphicTypeConversionModeType
	version                       []uint8
}

// NewAlgebrizerConfig returns a new AlgebrizerConfig constructed from the
// provided values. AlgebrizerConfigs should always be constructed via this
// function instead of via a struct literal.
func NewAlgebrizerConfig(lg log.Logger, dbName string, c catalog.Catalog) *AlgebrizerConfig {
	vars := c.Variables()
	return &AlgebrizerConfig{
		lg:                            lg,
		dbName:                        dbName,
		catalog:                       c,
		isMongos:                      vars.GetString(variable.MongoDBTopology) == string(variable.MongosTopology),
		sqlValueKind:                  GetSQLValueKind(vars),
		sqlSelectLimit:                vars.GetUint64(variable.SQLSelectLimit),
		maxVarcharLength:              vars.GetUint16(variable.MongoDBMaxVarcharLength),
		groupConcatMaxLen:             vars.GetInt64(variable.GroupConcatMaxLen),
		polymorphicTypeConversionMode: catalog.GetPolymorphicTypeConversionMode(vars),
		version:                       getMongoDBVersion(vars),
	}
}

// Note: while most errors in the BI-Connector begin with lower case words, any
// algebrizer/mysqlerror begins with a capital letter for consistency with
// MySQL.

// AlgebrizeCommand takes a parsed SQL statement and returns an algebrized form
// of the command.
func AlgebrizeCommand(cfg *AlgebrizerConfig, stmt parser.Statement) (Command, error) {
	g := &selectIDGenerator{}
	algebrizer := &algebrizer{
		cfg:                         cfg,
		selectID:                    g.current,
		selectIDGenerator:           g,
		projectedColumnAggregateMap: make(map[int]SQLExpr),
		columnSet:                   make(map[string]struct{}),
	}

	switch typedStmt := stmt.(type) {
	case *parser.Kill:
		return algebrizer.translateKill(typedStmt)
	case *parser.Flush:
		return algebrizer.translateFlush(typedStmt)
	case *parser.AlterTable:
		return algebrizer.translateAlterTable(typedStmt)
	case *parser.DropTable:
		return algebrizer.translateDropTable(typedStmt)
	case *parser.RenameTable:
		return algebrizer.translateRenameTable(typedStmt)
	case *parser.Set:
		return algebrizer.translateSet(typedStmt)
	case *parser.Use:
		return algebrizer.translateUse(typedStmt)
	default:
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErNotSupportedYet,
			fmt.Sprintf("statement %T", typedStmt))
	}
}

// AlgebrizeQuery translates a parsed SQL statement into a plan stage. If the
// statement cannot be translated, it will return an error.
func AlgebrizeQuery(cfg *AlgebrizerConfig, stmt parser.Statement) (PlanStage, error) {
	g := &selectIDGenerator{}
	algebrizer := &algebrizer{
		cfg:                         cfg,
		selectID:                    g.generate(),
		selectIDGenerator:           g,
		projectedColumnAggregateMap: make(map[int]SQLExpr),
		columnSet:                   make(map[string]struct{}),
		ctes:                        make(ctePlanStages),
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
func (a *algebrizer) newMongoSourceOrDualStage() PlanStage {
	// NewMongoSourceDualStage requires $collStats to work, which was only added in 3.4.
	if !a.versionAtLeast(3, 4, 0) {
		return NewDualStage()
	}

	// NewMongoSourceDualStage requires $collStats to work, but $collStats is unreliable in the
	// sharded case. When running against a mongod, $collStats will return 1 document when run on
	// any database and table, even if those do not exist. However, when running $collStats against
	// a mongos on a table that does not exist, $collStats will return 0 documents. To avoid this
	// situation completely, we do not use MongoSource to push down dual stages if we are running
	// against a mongos.
	if a.cfg.isMongos {
		return NewDualStage()
	}

	dualDb, dualTable, ok := findMongoDatabaseAndTable(a.cfg.catalog)
	if !ok {
		return NewDualStage()
	}

	return NewMongoSourceDualStage(dualDb, dualTable, a.selectID, "")
}

// findMongoDatabaseAndTable searches the catalog for a MongoDBTable and returns it along with its containing database
// if found, otherwise the function returns nil for both.
func findMongoDatabaseAndTable(cl catalog.Catalog) (catalog.Database, catalog.MongoDBTable, bool) {
	for _, db := range cl.Databases() {
		tables := db.Tables()
		for _, table := range tables {
			if mongoTable, ok := table.(catalog.MongoDBTable); ok {
				return db, mongoTable, true
			}
		}
	}
	return nil, nil, false
}

type ctePlanStage struct {
	cte        *parser.CTE
	algebrizer *algebrizer
	planStage  PlanStage
}

type ctePlanStages map[string]*ctePlanStage

type algebrizer struct {
	cfg               *AlgebrizerConfig
	parent            *algebrizer
	selectIDGenerator *selectIDGenerator
	// the selectID to use for projected columns.
	selectID int
	// the selectIDs that are currently used.
	currentSelectIDs []int
	// all the columns in scope.
	columns []*Column
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
}

func (a *algebrizer) valueKind() SQLValueKind {
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
	}
}

func (a *algebrizer) fullName(tableName, columnName string) string {
	fn := columnName
	if tableName != "" {
		fn = tableName + "." + fn
	}

	return fn
}

func (a *algebrizer) lookupColumn(databaseName, tableName, columnName string) (*Column, error) {
	var found *Column
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
	for _, pc := range a.projectedColumns {
		if strings.EqualFold(pc.Name, columnName) {
			if found {
				return nil,
					false,
					mysqlerrors.Defaultf(mysqlerrors.ErNonUniqError,
						columnName,
						a.currentClause)
			}
			result = pc
			found = true
		}
	}

	return &result, found, nil
}

func (a *algebrizer) findSQLColumn(sqlCol SQLColumnExpr) *Column {
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

// getCatalogColumn gets the catalog.Column corresponding to an evaluator.Column.
func (a *algebrizer) getCatalogColumn(column *Column) (catalog.Column, error) {
	if column.OriginalName == "" && column.MongoType == schema.MongoNone {
		// this is from a subquery, and we cannot get a catalog reference.
		return nil, nil
	}
	db, err := a.cfg.catalog.Database(column.Database)
	if err != nil {
		return nil, fmt.Errorf("could not find database: '%s' in catalog", column.Database)
	}
	table, err := db.Table(column.OriginalTable)
	if err != nil {
		return nil, fmt.Errorf("could not find table: '%s' in catalog", column.OriginalTable)
	}
	catColumn, err := table.Column(column.OriginalName)
	if err != nil {
		return nil, fmt.Errorf("could not find column: '%s' in catalog", column.OriginalName)
	}
	return catColumn, nil
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
			column.Name, column.EvalType, column.MongoType)
		catalogColumn, catalogErr := a.getCatalogColumn(column)
		if catalogErr != nil {
			return nil, catalogErr
		}
		mode := string(a.cfg.polymorphicTypeConversionMode)
		if catalogColumn != nil && catalogColumn.ShouldConvert(mode) {
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

	// the column was not available in the current scope, so it must be a correlated
	// column. We will mark the column as correlated and search all parent scopes until
	// we find the select that brings this column into scope.
	if a.parent != nil {
		expr, parentErr := a.parent.resolveColumnExpr(databaseName, tableName, columnName)
		if parentErr == nil {
			a.correlated = true
			col := expr.(SQLColumnExpr)
			col.correlated = true
			return expr, nil
		}
	}

	return nil, err
}

func (a *algebrizer) registerColumns(columns []*Column) error {
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

func (a *algebrizer) translateAlterTable(alter *parser.AlterTable) (*AlterCommand, error) {

	db, err := a.cfg.catalog.Database(a.cfg.dbName)
	if err != nil {
		return nil, err
	}

	tableName := strings.ToLower(alter.Table.Name)
	table, err := db.Table(tableName)
	if err != nil {
		return nil, err
	}

	if _, ok := table.(catalog.MongoDBTable); !ok {
		return nil, fmt.Errorf("cannot alter non-mongodb table %q", parser.String(alter.Table))
	}

	alterations := []*schema.Alteration{}

	for _, spec := range alter.Specs {
		switch spec.Type {
		case parser.AltRenameColumn:
			colName := strings.ToLower(spec.Column.Name)
			_, err := table.Column(colName)
			if err != nil {
				return nil, err
			}
			newColName := strings.ToLower(spec.NewColumn.Name)
			_, err = table.Column(newColName)
			if err == nil {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ErDupFieldname, newColName)
			}
			alteration := &schema.Alteration{
				Timestamp: time.Now(),
				Type:      schema.RenameColumn,
				Db:        a.cfg.dbName,
				Table:     tableName,
				Column:    colName,
				NewColumn: newColName,
			}
			alterations = append(alterations, alteration)

		case parser.AltDropColumn:
			colName := strings.ToLower(spec.Column.Name)
			_, err := table.Column(colName)
			if err != nil {
				return nil, err
			}
			if len(table.Columns()) == 1 {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ErCantRemoveAllFields)
			}
			if strings.Split(colName, ".")[0] == mongoPrimaryKey {
				return nil, fmt.Errorf("cannot drop column %s: not allowed", colName)
			}
			alteration := &schema.Alteration{
				Timestamp: time.Now(),
				Type:      schema.DropColumn,
				Db:        a.cfg.dbName,
				Table:     tableName,
				Column:    colName,
			}
			alterations = append(alterations, alteration)

		case parser.AltModifyColumn:
			colName := strings.ToLower(spec.Column.Name)
			_, err := table.Column(colName)
			if err != nil {
				return nil, err
			}
			alteration := &schema.Alteration{
				Timestamp:     time.Now(),
				Type:          schema.ModifyColumn,
				Db:            a.cfg.dbName,
				Table:         tableName,
				Column:        colName,
				NewColumnType: spec.NewColumnType,
			}
			alterations = append(alterations, alteration)

		case parser.AltRenameTable:
			newTableName := strings.ToLower(spec.NewTable.Name)
			_, err := db.Table(newTableName)
			if err == nil {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ErTableExistsError, newTableName)
			}
			alteration := &schema.Alteration{
				Timestamp: time.Now(),
				Type:      schema.RenameTable,
				Db:        a.cfg.dbName,
				Table:     tableName,
				NewTable:  newTableName,
			}
			alterations = append(alterations, alteration)

		default:
			return nil, fmt.Errorf("invalid Alter Table type %q", spec.Type)
		}
	}

	return &AlterCommand{alterations}, nil
}

// nolint: unparam
func (a *algebrizer) translateDropTable(ddl *parser.DropTable) (*DropCommand, error) {
	return NewDropCommand(ddl.Name.Name), nil
}

func (a *algebrizer) translateRenameTable(rename *parser.RenameTable) (*AlterCommand, error) {

	db, err := a.cfg.catalog.Database(a.cfg.dbName)
	if err != nil {
		return nil, err
	}

	alterations := []*schema.Alteration{}

	for _, spec := range rename.Renames {

		tableName := strings.ToLower(spec.Table.Name)
		table, err := db.Table(tableName)
		if err != nil {
			return nil, err
		}

		if _, ok := table.(catalog.MongoDBTable); !ok {
			return nil, fmt.Errorf("cannot alter non-mongodb table %q", parser.String(spec.Table))
		}

		newTableName := strings.ToLower(spec.NewTable.Name)
		_, err = db.Table(newTableName)
		if err == nil {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErTableExistsError, newTableName)
		}
		alteration := &schema.Alteration{
			Timestamp: time.Now(),
			Type:      schema.RenameTable,
			Db:        a.cfg.dbName,
			Table:     tableName,
			NewTable:  newTableName,
		}
		alterations = append(alterations, alteration)

	}

	return &AlterCommand{alterations}, nil
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

		switch typedE := eval.(type) {
		case SQLUint64:
			offset = Uint64(typedE)
		case SQLInt64:
			if Int64(typedE) < 0 {
				return 0,
					0,
					mysqlerrors.Newf(mysqlerrors.ErSyntaxError,
						"Offset cannot be negative")
			}
			offset = Uint64(typedE)
		default:
			return 0, 0, mysqlerrors.Defaultf(mysqlerrors.ErWrongSpvarTypeInLimit)
		}
	}

	if limit.Rowcount != nil {
		eval, err := a.translateExpr(limit.Rowcount)
		if err != nil {
			return 0, 0, err
		}

		switch typedE := eval.(type) {
		case SQLUint64:
			rowcount = Uint64(typedE)
		case SQLInt64:
			if Int64(typedE) < 0 {
				return 0,
					0,
					mysqlerrors.Newf(mysqlerrors.ErSyntaxError,
						"Rowcount cannot be negative")
			}
			rowcount = Uint64(typedE)
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
	case *parser.SimpleSelect:
		return a.translateSimpleSelect(typedS)
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

func (a *algebrizer) translateSimpleSelect(sel *parser.SimpleSelect) (PlanStage, error) {
	a.currentClause = fieldList
	projectedColumns, err := a.translateSelectExprs(sel.SelectExprs)
	if err != nil {
		return nil, err
	}

	if sel.Limit != nil {
		a.currentClause = limitClause
		offset, limit, err := a.translateLimit(sel.Limit)
		if err != nil {
			return nil, err
		}

		plan := a.newMongoSourceOrDualStage()

		return NewProjectStage(NewLimitStage(plan,
			offset, limit), projectedColumns...,
		), nil
	}

	plan := a.newMongoSourceOrDualStage()

	return NewProjectStage(plan, projectedColumns...), nil
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
				pcs[0].Table, pcs[0].Name, pcs[0].EvalType, schema.MongoNone)
			plan = NewCountStage(mongoSource, pcs[0])
			plan = NewProjectStage(plan, pcs[0])
			return plan, nil
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
			builder.where = NewSQLConvertExpr(builder.where, EvalBoolean)
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
		projectedColumns = Columns(plan.Columns()).ToProjectedColumns()
		plan = NewGroupByStage(plan, projectedColumns.Exprs(), projectedColumns)
	case parser.AST_UNION_ALL:
		plan = NewUnionStage(UnionAll, left, right)
		projectedColumns = Columns(plan.Columns()).ToProjectedColumns()
	default:
		return nil, mysqlerrors.Newf(mysqlerrors.ErNotSupportedYet,
			"Cannot perform set operation '%s'", union.Type)
	}
	return NewProjectStage(plan, projectedColumns...), nil
}

func (a *algebrizer) translateSelectExprs(
	selectExprs parser.SelectExprs) (ProjectedColumns, error) {
	var projectedColumns ProjectedColumns
	hasGlobalStar := false
	mode := string(a.cfg.polymorphicTypeConversionMode)

	for _, selectExpr := range selectExprs {
		switch typedE := selectExpr.(type) {

		case *parser.StarExpr:
			databaseName := typedE.DatabaseName.Else("")
			tableName := typedE.TableName.Else("")

			if tableName == "" && databaseName == "" {
				hasGlobalStar = true
			}

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
					projectedColumn := column.projectAs(column.Name)
					projectedColumn.SelectID = a.selectID
					catalogColumn, catalogErr := a.getCatalogColumn(column)
					if catalogErr != nil {
						return nil, catalogErr
					}
					if catalogColumn != nil && catalogColumn.ShouldConvert(mode) {
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

			if translatedExpr.EvalType() == EvalTuple {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, 1)
			}

			var projectedColumn *ProjectedColumn

			if sqlCol, ok := translatedExpr.(SQLColumnExpr); ok {
				if c := a.findSQLColumn(sqlCol); c != nil {
					projectedColumn = c.projectWithExpr(translatedExpr)
				}
			}

			// This happens when the select expression is more than just a
			// column: it could be a scalar or aggregate function, or
			// any sort of operator like '+'
			if projectedColumn == nil {
				dbName := getDatabaseName(translatedExpr)
				projectedColumn = &ProjectedColumn{
					Expr: translatedExpr,
					Column: NewColumn(a.selectID, "", "", dbName, "", "", "",
						translatedExpr.EvalType(), schema.MongoNone, false),
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

	if hasGlobalStar && len(selectExprs) > 1 {
		return nil, mysqlerrors.Newf(mysqlerrors.ErSyntaxError,
			"Cannot have a '*' in conjunction with any other columns")
	}

	return projectedColumns, nil
}

func (a *algebrizer) translateSet(set *parser.Set) (*SetCommand, error) {
	assignments := []*SQLAssignmentExpr{}
	for _, e := range set.Exprs {
		v, err := a.translateVariableExpr(e.Name)
		if err != nil {
			return nil, err
		}

		expr, err := a.translateExpr(e.Expr)
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
	var colsForProjection []*Column
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
			plan = NewJoinStage(joinKind, plan, temp, NewSQLBool(a.valueKind(), true))
		}
	}

	if isUnqualifiedSelectStar {
		a.columns = colsForProjection
	}

	return plan, nil
}

// getTableName gets the name of the table that contains colName. The table
// name is only specific to the column when tableExpr.(type) = JoinTableExpr.
func (a *algebrizer) getTableName(tableExpr parser.TableExpr, colName string, columns []*Column) (string, error) {
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
	leftCols []*Column,
	rightCols []*Column,
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
	columns     []*Column
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

func (c columnsUsing) Filter() []*Column {
	// keep track of the USING columns to ensure we only project one for each column in the clause
	seenColumns := make(map[string]bool)
	for _, usingColumn := range c.usingCols {
		seenColumns[usingColumn.Name] = false
	}
	var columnsForProjection []*Column
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

func (a *algebrizer) translateTableExpr(tableExpr parser.TableExpr, hasGlobalStraightJoin bool) (PlanStage, []*Column, error) {
	switch typedT := tableExpr.(type) {
	case *parser.AliasedTableExpr:
		return a.translateSimpleTableExpr(typedT.Expr, typedT.As.Else(""))
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
		cols := append(leftCols, rightCols...)

		var predicate SQLExpr = NewSQLBool(a.valueKind(), true)
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
				predicate = NewSQLBool(a.valueKind(), true)
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
	tableExpr parser.SimpleTableExpr, aliasName string) (PlanStage, []*Column, error) {
	switch typedT := tableExpr.(type) {
	case *parser.TableName:
		tableName := strings.ToLower(typedT.Name)
		if aliasName == "" {
			aliasName = tableName
		}

		var plan PlanStage
		var err error

		if strings.EqualFold(tableName, "DUAL") {
			plan = a.newMongoSourceOrDualStage()
		} else {
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
						a.projectedColumns[i] = columns[i].projectAs(expr.Name)
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

			db, dbErr := a.cfg.catalog.Database(dbName)
			if dbErr != nil {
				return nil, nil, dbErr
			}

			table, dbErr := db.Table(tableName)
			if dbErr != nil {
				return nil, nil, dbErr
			}

			switch t := table.(type) {
			case catalog.MongoDBTable:
				plan = NewMongoSourceStage(db, t, a.selectID, aliasName)
			case *catalog.DynamicTable:
				plan = NewDynamicSourceStage(db, t, a.selectID, aliasName)
			default:
				return nil, nil, fmt.Errorf("unknown table type: %T", t)
			}
			err = a.registerTable(dbName, aliasName)

		}

		if err != nil {
			return nil, nil, err
		}

		columns := plan.Columns()

		err = a.registerColumns(columns)
		if err != nil {
			return nil, nil, err
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

	return &SQLSubqueryExpr{
		plan:       plan,
		correlated: correlated,
	}, nil
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

	var err error
	var leftPlan, rightPlan PlanStage
	var leftIsCorrelated, rightIsCorrelated bool
	var leftExpr SQLExpr

	// If the left side is a subquery, translate it into a PlanStage.
	// Otherwise, translate it to a SQLExpr.
	lsq, leftIsPlan := expr.Left.(*parser.Subquery)
	if leftIsPlan {
		leftPlan, leftIsCorrelated, err = a.translateSubqueryPlan(lsq)
	} else {
		leftExpr, err = a.translateExpr(expr.Left)
	}
	if err != nil {
		return nil, err
	}

	// If the right side is a subquery, translate it into a PlanStage.
	rsq, rightIsPlan := expr.Right.(*parser.Subquery)
	if rightIsPlan {
		rightPlan, rightIsCorrelated, err = a.translateSubqueryPlan(rsq)
		if err != nil {
			return nil, err
		}
	}

	switch expr.SubqueryOperator {
	case "":
		// if left and right sides are both subqueries, create a SQLFullSubqueryCmp
		if leftIsPlan && rightIsPlan {
			cmp := NewSQLFullSubqueryCmpExpr(
				leftIsCorrelated, rightIsCorrelated,
				leftPlan, rightPlan,
				expr.Operator,
			)
			return cmp, nil
		}

		// if just the right side is a subquery, create a SQLRightSubqueryCmp
		if rightIsPlan {
			cmp := NewSQLRightSubqueryCmpExpr(
				rightIsCorrelated,
				leftExpr,
				rightPlan,
				expr.Operator,
			)
			return cmp, nil
		}

		// if just the left side is a subquery, our desugarer did not run or
		// has become buggy, panic accordingly.
		if leftIsPlan {
			panic("found left-side only subquery, this should have been rewritten during desugaring")
		}

	case parser.AST_IN:
		// The right Expr must be a subquery at this point.
		if !rightIsPlan {
			panic("right side of an IN expr was not a subquery")
		}

		if leftIsPlan {
			cmp := NewSQLSubqueryInSubqueryExpr(
				leftIsCorrelated,
				rightIsCorrelated,
				leftPlan,
				rightPlan,
			)
			return cmp, nil
		}

		cmp := NewSQLInSubqueryExpr(
			rightIsCorrelated,
			leftExpr,
			rightPlan,
		)
		return cmp, nil

	case parser.AST_NOT_IN:
		// The right Expr must be a subquery at this point.
		if !rightIsPlan {
			panic("right side of a NOT IN expr was not a subquery")
		}

		if leftIsPlan {
			cmp := NewSQLSubqueryNotInSubqueryExpr(
				leftIsCorrelated,
				rightIsCorrelated,
				leftPlan,
				rightPlan,
			)
			return cmp, nil
		}

		cmp := NewSQLNotInSubqueryExpr(
			rightIsCorrelated,
			leftExpr,
			rightPlan,
		)
		return cmp, nil

	case parser.AST_ANY:
		// The right Expr must be a subquery at this point.
		if !rightIsPlan {
			panic("right side of an ANY expr was not a subquery")
		}

		if leftIsPlan {
			cmp := NewSQLSubqueryAnyExpr(
				leftIsCorrelated,
				rightIsCorrelated,
				leftPlan,
				rightPlan,
				expr.Operator,
			)
			return cmp, nil
		}

		cmp := NewSQLAnyExpr(
			rightIsCorrelated,
			leftExpr,
			rightPlan,
			expr.Operator,
		)
		return cmp, nil

	case parser.AST_ALL:
		if !rightIsPlan {
			panic("right side of an ALL expr was not a subquery")
		}

		if leftIsPlan {
			cmp := NewSQLSubqueryAllExpr(
				leftIsCorrelated,
				rightIsCorrelated,
				leftPlan,
				rightPlan,
				expr.Operator,
			)
			return cmp, nil
		}

		cmp := NewSQLAllExpr(
			rightIsCorrelated,
			leftExpr,
			rightPlan,
			expr.Operator,
		)
		return cmp, nil
	}

	// if we get here, we made an error invoking this function
	panic("neither left nor right side of a subquery comparison expr was a subquery")
}

func (a *algebrizer) translateExprHelper(expr parser.Expr) (SQLExpr, error) {
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
			return NewSQLNotExpr(NewSQLIsExpr(left, right)), nil
		}

		if left.EvalType() == EvalTuple || right.EvalType() == EvalTuple {
			panic("tuple comparisons must be eliminated in the desugarer")
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
			date, _, ok := parseDateTime(arg)
			if !ok || date.Hour() > 0 ||
				date.Minute() > 0 ||
				date.Second() > 0 ||
				date.Nanosecond() > 0 {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongValue, "DATE", arg)
			}
			return NewSQLDate(a.valueKind(), date), nil
		case parser.AST_TIME:
			dur, _, ok := strToTime(arg)
			if !ok {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongValue, "TIME", arg)
			}

			date := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC).Add(dur)

			return NewSQLTimestamp(a.valueKind(), date), nil
		case parser.AST_TIMESTAMP, parser.AST_DATETIME:
			date, _, ok := strToDateTime(arg, true)
			if !ok {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongValue, "DATETIME", arg)
			}
			return NewSQLTimestamp(a.valueKind(), date), nil
		default:
			return nil, mysqlerrors.Newf(mysqlerrors.ErNotSupportedYet,
				"No support for constructor '%v'", typedE.Name)
		}
	case *parser.ExistsExpr:
		return a.translateExistsExpr(typedE)
	case *parser.FalseVal:
		return NewSQLBool(a.valueKind(), false), nil
	case *parser.FuncExpr:
		return a.translateFuncExpr(typedE)
	case parser.KeywordVal:
		return NewSQLVarchar(a.valueKind(), string(typedE)), nil
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
		} else {
			escape = NewSQLVarchar(a.valueKind(), "\\")
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
		return NewPolymorphicSQLNull(a.valueKind()), nil
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
			return NewSQLFloat(a.valueKind(), f), nil
		}
		if strings.Contains(exprString, ".") {
			d, err := decimal.NewFromString(exprString)
			if err != nil {
				return nil,
					mysqlerrors.Defaultf(mysqlerrors.ErIllegalValueForType,
						"decimal",
						exprString)

			}
			return NewSQLDecimal128(a.valueKind(), d), nil
		}

		// try to parse as int64 first
		if i, err := strconv.ParseInt(exprString, 10, 64); err == nil {
			return NewSQLInt64(a.valueKind(), i), nil
		}

		// next try to parse as uint64
		if i, err := strconv.ParseUint(exprString, 10, 64); err == nil {
			return NewSQLUint64(a.valueKind(), i), nil
		}

		if useFloats {
			f, err := strconv.ParseFloat(exprString, 64)
			if err != nil {
				return nil,
					mysqlerrors.Defaultf(mysqlerrors.ErIllegalValueForType,
						"integer",
						exprString)
			}
			return NewSQLFloat(a.valueKind(), f), nil
		}

		i, err := decimal.NewFromString(exprString)
		if err != nil {
			return nil,
				mysqlerrors.Defaultf(mysqlerrors.ErIllegalValueForType,
					"integer",
					exprString)
		}
		return NewSQLDecimal128(a.valueKind(), i), nil
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
		return &SQLRegexExpr{operand, pattern}, nil
	case *parser.RLikeExpr:
		operand, err := a.translateExpr(typedE.Operand)
		if err != nil {
			return nil, err
		}

		pattern, err := a.translateExpr(typedE.Pattern)
		if err != nil {
			return nil, err
		}
		return &SQLRegexExpr{operand, pattern}, nil
	case parser.StrVal:
		return NewSQLVarchar(a.valueKind(), string(typedE)), nil
	case *parser.Subquery:
		return a.translateSubqueryExpr(typedE)
	case *parser.TrueVal:
		return NewSQLBool(a.valueKind(), true), nil
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
		return NewPolymorphicSQLNull(a.valueKind()), nil
	default:
		return nil, mysqlerrors.Newf(mysqlerrors.ErNotSupportedYet,
			"No support for '%v'", parser.String(typedE))
	}
}

func (a *algebrizer) translateExpr(expr parser.Expr) (SQLExpr, error) {
	translatedExpr, err := a.translateExprHelper(expr)
	if err != nil {
		return nil, err
	}

	reconciled, err := translatedExpr.reconcile()
	if err != nil {
		return nil, err
	}
	return reconciled, nil
}

func (a *algebrizer) translatePossibleColumnRefExpr(expr parser.Expr) (SQLExpr, error) {
	if numVal, ok := expr.(parser.NumVal); ok {
		n, err := strconv.ParseInt(parser.String(numVal), 10, 64)
		if err != nil {
			return nil, err
		}

		if int(n) > len(a.projectedColumns) {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErBadFieldError,
				strconv.Itoa(int(n)), a.currentClause)
		}

		if n >= 0 {
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

			matcher = NewSQLEqualsExpr(e, cond)
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
		elseValue = NewPolymorphicSQLNull(a.valueKind())
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
		return NewSQLGroupConcatFunctionExpr(distinct, exprs)
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

				exprs = append(exprs, NewSQLVarchar(a.valueKind(), "*"))

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
			groupConcat.Separator = expr.Separator
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

		current.aggregates = append(current.aggregates, aggExpr)
		return NewSQLColumnExpr(current.selectID, getDatabaseName(aggExpr), "", aggExpr.String(),
			aggExpr.EvalType(), schema.MongoNone), nil
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
					exprs = append(exprs, NewSQLVarchar(a.valueKind(), as))
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
		return NewSQLConvertExpr(exprs[0], EvalDate), nil
	case "timestamp":
		if len(exprs) == 1 {
			return NewSQLConvertExpr(exprs[0], EvalDatetime), nil
		} else {
			return NewSQLScalarFunctionExpr(name, exprs)
		}
	case "rand":
		// We need something unique that we can map.
		id := a.getUniqueID()
		return NewSQLScalarFunctionExpr(
			"rand",
			append([]SQLExpr{NewSQLUint64(a.valueKind(), id)}, exprs...),
		)
	case "isnull":
		return NewSQLIsExpr(exprs[0], NewPolymorphicSQLNull(a.valueKind())), nil
	case "date_add", "adddate", "date_sub", "subdate":
		if len(exprs) == 2 {
			return NewSQLScalarFunctionExpr(
				name,
				[]SQLExpr{exprs[0], exprs[1], NewSQLVarchar(a.valueKind(), Day)},
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
				[]SQLExpr{exprs[0], NewSQLInt64(a.valueKind(), 0)},
			)
		}
		return NewSQLScalarFunctionExpr("week", []SQLExpr{exprs[0], NewSQLInt64(a.valueKind(), 3)})
	case "yearweek":
		if len(exprs) == 2 {
			return NewSQLScalarFunctionExpr("yearweek", []SQLExpr{exprs[0], exprs[1]})
		}
		return NewSQLScalarFunctionExpr(
			"yearweek",
			[]SQLExpr{exprs[0], NewSQLInt64(a.valueKind(), 0)},
		)
	default:
		return NewSQLScalarFunctionExpr(name, exprs)
	}

}

func (a *algebrizer) translateVariableExpr(c *parser.ColName) (*SQLVariableExpr, error) {

	v := &SQLVariableExpr{
		Kind:  variable.SystemKind,
		Scope: variable.SessionScope,
	}

	pos := 0
	str := c.Name
	if c.Qualifier.IsSome() {
		str = c.Qualifier.Unwrap() + "." + str
	}

	if str[pos] == '@' {
		pos++
		if len(str) > 1 && str[pos] != '@' {
			v.Kind = variable.UserKind
		} else {
			pos++
		}
	}

	v.Name = str[pos:]

	if v.Kind != variable.UserKind {
		idx := strings.Index(v.Name, ".")
		if idx >= 0 {
			switch strings.ToLower(v.Name[:idx+1]) {
			case "global.":
				v.Name = v.Name[idx+1:]
				v.Scope = variable.GlobalScope
			case "session.", "local.":
				v.Name = v.Name[idx+1:]
			}
		}
	}

	value, err := a.cfg.catalog.Variables().Get(variable.Name(v.Name), v.Scope, v.Kind)
	if err != nil {
		return nil, err
	}

	v.Value = value.Value
	v.SQLType = value.SQLType

	return v, nil
}

func (a *algebrizer) versionAtLeast(major, minor, patch uint8) bool {
	return procutil.VersionAtLeast(a.cfg.version, []uint8{major, minor, patch})
}
