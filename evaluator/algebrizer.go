package evaluator

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
	"github.com/shopspring/decimal"
)

// Note: while most errors in the BI-Connector begin with lower case words, any
// algebrizer/mysqlerror begins with a capital letter for consistency with
// MySQL.

// AlgebrizeCommand takes a parsed SQL statement and returns an algebrized form
// of the command.
func AlgebrizeCommand(stmt parser.Statement, dbName string,
	vars *variable.Container, catalog *catalog.Catalog) (Command, error) {
	g := &selectIDGenerator{}
	l := log.NewComponentLogger(log.AlgebrizerComponent, log.GlobalLogger())
	algebrizer := &algebrizer{
		dbName:                      dbName,
		catalog:                     catalog,
		logger:                      l,
		selectID:                    g.current,
		selectIDGenerator:           g,
		projectedColumnAggregateMap: make(map[int]SQLExpr),
		variables:                   vars,
	}

	switch typedStmt := stmt.(type) {
	case *parser.Kill:
		return algebrizer.translateKill(typedStmt)
	case *parser.Set:
		return algebrizer.translateSet(typedStmt)
	case *parser.Flush:
		return algebrizer.translateFlush(typedStmt)
	case *parser.AlterTable:
		return algebrizer.translateAlterTable(typedStmt)
	case *parser.RenameTable:
		return algebrizer.translateRenameTable(typedStmt)
	default:
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErNotSupportedYet,
			fmt.Sprintf("statement %T", typedStmt))
	}
}

// AlgebrizeQuery translates a parsed SQL statement into a plan stage. If the
// statement cannot be translated, it will return an error.
func AlgebrizeQuery(statement parser.Statement, dbName string,
	vars *variable.Container, catalog *catalog.Catalog) (PlanStage, error) {
	g := &selectIDGenerator{}
	l := log.NewComponentLogger(log.AlgebrizerComponent, log.GlobalLogger())
	algebrizer := &algebrizer{
		dbName:                      dbName,
		catalog:                     catalog,
		logger:                      l,
		selectID:                    g.generate(),
		selectIDGenerator:           g,
		projectedColumnAggregateMap: make(map[int]SQLExpr),
		variables:                   vars,
	}

	switch typedStmt := statement.(type) {
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

type algebrizer struct {
	parent            *algebrizer
	catalog           *catalog.Catalog
	variables         *variable.Container
	selectIDGenerator *selectIDGenerator
	logger            log.Logger
	// the selectID to use for projected columns.
	selectID int
	// the selectIDs that are currently used.
	currentSelectIDs []int
	// the default database name.
	dbName string
	// all the columns in scope.
	columns []*Column
	// all the table names in scope.
	tableNames []string
	// indicates whether this context is using columns in its parent.
	correlated bool
	// aggregates found in the current scope.
	aggregates []*SQLAggFunctionExpr
	// columns to be projected from this scope.
	projectedColumns ProjectedColumns
	// indicates whether the projected column contains an aggregate.
	projectedColumnAggregateMap map[int]SQLExpr
	// indicates whether to resolve a column using the projected columns first or second.
	resolveProjectedColumnsFirst bool
	// tracks the current clause being processed for the purposes of error messages.
	currentClause string
}

func (a *algebrizer) clone() *algebrizer {
	return &algebrizer{
		parent:                      a,
		catalog:                     a.catalog,
		dbName:                      a.dbName,
		selectID:                    a.selectID,
		selectIDGenerator:           a.selectIDGenerator,
		projectedColumnAggregateMap: make(map[int]SQLExpr),
		variables:                   a.variables,
	}
}

func (a *algebrizer) newSubqueryAlgebrizer() *algebrizer {
	return &algebrizer{
		dbName:                      a.dbName,
		catalog:                     a.catalog,
		selectID:                    a.selectIDGenerator.generate(),
		selectIDGenerator:           a.selectIDGenerator,
		projectedColumnAggregateMap: make(map[int]SQLExpr),
		variables:                   a.variables,
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

func (a *algebrizer) resolveColumnExpr(dataBaseName, tableName,
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

	column, err := a.lookupColumn(dataBaseName, tableName, columnName)
	if err == nil {
		if a.currentClause != whereClause && column.MongoType == schema.MongoFilter {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErBadFieldError, column.Name, column.Table)
		}
		return NewSQLColumnExpr(column.SelectID, column.Database, column.Table,
			column.Name, column.SQLType, column.MongoType), nil
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

	// we didn't find it in the current scope, so we need to search our parent,
	// and let it search its parent, etc.
	if a.parent != nil {
		expr, parentErr := a.parent.resolveColumnExpr(dataBaseName, tableName, columnName)
		if parentErr == nil {
			a.correlated = true
			return expr, nil
		}
	}

	return nil, err
}

func (a *algebrizer) registerColumns(columns []*Column) error {
	contains := func(c *Column) bool {
		for _, c2 := range a.columns {
			// we don't use SelectID here because it's irrelevant to whether a query
			// is semantically valid.
			if strings.EqualFold(c.Name, c2.Name) && strings.EqualFold(c.Table, c2.Table) &&
				strings.EqualFold(c.Database, c2.Database) {
				return true
			}
		}
		return false
	}

	// this ensures that we have no duplicate columns. We have to check duplicates
	// against the existing columns as well as against itself.
	for _, c := range columns {
		if contains(c) {
			return mysqlerrors.Defaultf(mysqlerrors.ErDupFieldname, a.fullName(c.Table, c.Name))
		}
		a.columns = append(a.columns, c)
		if !containsInt(a.currentSelectIDs, c.SelectID) {
			a.currentSelectIDs = append(a.currentSelectIDs, c.SelectID)
		}
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

// isAggFunction returns true if the byte slice e contains the name of an
// aggregate function and false otherwise.
func (a *algebrizer) isAggFunction(name string) bool {
	switch strings.ToLower(name) {
	case "avg", "sum", "count", "max", "min", "std", "stddev", "stddev_pop", "stddev_samp":
		return true
	default:
		return false
	}
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

	db, err := a.catalog.Database(a.dbName)
	if err != nil {
		return nil, err
	}

	tableName := strings.ToLower(string(alter.Table.Name))
	table, err := db.Table(tableName)
	if err != nil {
		return nil, err
	}

	_, ok := table.(*catalog.MongoTable)
	if !ok {
		return nil, fmt.Errorf("cannot alter non-mongodb table %q", alter.Table)
	}

	alterations := []*schema.Alteration{}

	for _, spec := range alter.Specs {
		switch spec.Type {
		case schema.RenameColumn:
			colName := strings.ToLower(string(spec.Column.Name))
			_, err := table.Column(colName)
			if err != nil {
				return nil, err
			}
			newColName := strings.ToLower(string(spec.NewColumn.Name))
			_, err = table.Column(newColName)
			if err == nil {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ErDupFieldname, newColName)
			}
			alteration := &schema.Alteration{
				Timestamp: time.Now(),
				Type:      schema.RenameColumn,
				Db:        a.dbName,
				Table:     tableName,
				Column:    colName,
				NewColumn: newColName,
			}
			alterations = append(alterations, alteration)

		case schema.DropColumn:
			colName := strings.ToLower(string(spec.Column.Name))
			_, err := table.Column(colName)
			if err != nil {
				return nil, err
			}
			if len(table.Columns()) == 1 {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ErCantRemoveAllFields)
			}
			if strings.Split(colName, ".")[0] == "_id" {
				return nil, fmt.Errorf("cannot drop column %s: not allowed", colName)
			}
			alteration := &schema.Alteration{
				Timestamp: time.Now(),
				Type:      schema.DropColumn,
				Db:        a.dbName,
				Table:     tableName,
				Column:    colName,
			}
			alterations = append(alterations, alteration)

		case schema.RenameTable:
			newTableName := strings.ToLower(string(spec.NewTable.Name))
			_, err := db.Table(newTableName)
			if err == nil {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ErTableExistsError, newTableName)
			}
			alteration := &schema.Alteration{
				Timestamp: time.Now(),
				Type:      schema.RenameTable,
				Db:        a.dbName,
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

func (a *algebrizer) translateRenameTable(rename *parser.RenameTable) (*AlterCommand, error) {

	db, err := a.catalog.Database(a.dbName)
	if err != nil {
		return nil, err
	}

	alterations := []*schema.Alteration{}

	for _, spec := range rename.Renames {

		tableName := strings.ToLower(string(spec.Table.Name))
		table, err := db.Table(tableName)
		if err != nil {
			return nil, err
		}

		_, ok := table.(*catalog.MongoTable)
		if !ok {
			return nil, fmt.Errorf("cannot alter non-mongodb table %q", spec.Table)
		}

		newTableName := strings.ToLower(string(spec.NewTable.Name))
		_, err = db.Table(newTableName)
		if err == nil {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErTableExistsError, newTableName)
		}
		alteration := &schema.Alteration{
			Timestamp: time.Now(),
			Type:      schema.RenameTable,
			Db:        a.dbName,
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

func (a *algebrizer) translateLimit(limit *parser.Limit) (SQLUint64, SQLUint64, error) {
	var rowcount SQLUint64
	var offset SQLUint64

	if limit.Offset != nil {
		eval, err := a.translateExpr(limit.Offset)
		if err != nil {
			return 0, 0, err
		}

		switch typedE := eval.(type) {
		case SQLUint64:
			offset = typedE
		case SQLInt:
			if typedE < 0 {
				return 0,
					0,
					mysqlerrors.Newf(mysqlerrors.ErSyntaxError,
						"Offset cannot be negative")
			}
			offset = SQLUint64(typedE.Uint64())
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
			rowcount = typedE
		case SQLInt:
			if typedE < 0 {
				return 0,
					0,
					mysqlerrors.Newf(mysqlerrors.ErSyntaxError,
						"Rowcount cannot be negative")
			}
			rowcount = SQLUint64(typedE.Uint64())
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
	sqlSelectLimit := a.variables.GetUInt64(variable.SQLSelectLimit)
	if sqlSelectLimit != math.MaxUint64 {
		if pr, ok := plan.(*ProjectStage); ok {
			plan = NewLimitStage(pr.source, 0, sqlSelectLimit)
			plan = NewProjectStage(plan, pr.projectedColumns...)
		}
	}

	return plan, nil
}

func (a *algebrizer) translateSelectStatement(
	selectStatement parser.SelectStatement) (PlanStage, error) {
	switch typedS := selectStatement.(type) {
	case *parser.Select:
		return a.translateSelect(typedS)
	case *parser.SimpleSelect:
		return a.translateSimpleSelect(typedS)
	case *parser.Union:
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
		return NewProjectStage(NewLimitStage(NewDualStage(),
			uint64(offset), uint64(limit)), projectedColumns...,
		), nil
	}

	return NewProjectStage(NewDualStage(), projectedColumns...), nil
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
				if expr.TableName == nil {
					isUnqualifiedSelectStar = true
				}
			}
		}

		plan, err := a.translateTableExprs(sel.From, isUnqualifiedSelectStar)
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
				pcs[0].Table, pcs[0].Name, pcs[0].SQLType, schema.MongoNone)
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

		err = builder.includeFrom(plan)
		if err != nil {
			return nil, err
		}
	}

	// 2. Translate all the other clauses from this scope. We aren't going to create the plan stages
	// yet because the expressions may need to be substituted if a group by exists.
	if sel.Where != nil {
		a.currentClause = whereClause
		err := builder.includeWhere(sel.Where)
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

	builder.distinct = sel.Distinct == parser.AST_DISTINCT

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
	for _, selectExpr := range selectExprs {
		switch typedE := selectExpr.(type) {

		case *parser.StarExpr:
			databaseName, tableName := "", ""
			if typedE.DatabaseName != nil {
				databaseName = string(typedE.DatabaseName)
			}

			if typedE.TableName != nil {
				tableName = string(typedE.TableName)
			}

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

			if translatedExpr.Type() == schema.SQLTuple {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, 1)
			}

			var projectedColumn *ProjectedColumn

			if sqlCol, ok := translatedExpr.(SQLColumnExpr); ok {
				if c := a.findSQLColumn(sqlCol); c != nil {
					projectedColumn = c.projectWithExpr(translatedExpr)
				}
			}

			// This happens when there is an aggregate
			// function in the select expression
			if projectedColumn == nil {
				projectedColumn = &ProjectedColumn{
					Expr: translatedExpr,
					Column: NewColumn(a.selectID, "", "", "", "", "", "",
						translatedExpr.Type(), schema.MongoNone, false),
				}
			}

			if _, ok := translatedExpr.(*SQLVariableExpr); !ok {
				if sqlCol, ok := typedE.Expr.(*parser.ColName); ok {
					projectedColumn.Name = string(sqlCol.Name)
				}
			}

			if typedE.As != nil {
				projectedColumn.Name = string(typedE.As)
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

func (a *algebrizer) translateTableExprs(tableExprs parser.TableExprs,
	isUnqualifiedSelectStar bool) (PlanStage, error) {
	var plan PlanStage
	var colsForProjection []*Column
	for i, tableExpr := range tableExprs {
		temp, cols, err := a.translateTableExpr(tableExpr)
		if err != nil {
			return nil, err
		}
		colsForProjection = append(colsForProjection, cols...)
		if i == 0 {
			plan = temp
		} else {
			plan = NewJoinStage(CrossJoin, plan, temp, SQLTrue)
		}
	}

	if isUnqualifiedSelectStar {
		a.columns = colsForProjection
	}

	return plan, nil
}

// getTableName gets the name of the table that contains colName. The table
// name is only specific to the column when tableExpr.(type) = JoinTableExpr.
func (a *algebrizer) getTableName(
	tableExpr parser.TableExpr, colName string, columns []*Column) ([]byte, error) {
	switch typedE := tableExpr.(type) {
	case *parser.AliasedTableExpr:
		if typedE.As != nil {
			return typedE.As, nil
		}
		// only legal type is TableName, as other AliasedTableExpr's must have AS clause specified
		name, ok := typedE.Expr.(*parser.TableName)
		if !ok {
			return nil,
				mysqlerrors.Newf(mysqlerrors.ErParseError,
					"A %s must have an alias",
					typedE.Expr)
		}
		return name.Name, nil
	case *parser.ParenTableExpr:
		return a.getTableName(typedE.Expr, colName, columns)
	case *parser.JoinTableExpr:
		var tableName []byte
		// find the name of the table the column was in before the join
		for _, column := range columns {
			if column.Name == colName {
				if tableName == nil {
					tableName = []byte(column.Table)
				} else {
					return nil,
						mysqlerrors.Defaultf(mysqlerrors.ErNonUniqError,
							colName,
							a.currentClause)
				}
			}
		}
		// if column named colName was not found in any table in the JoinTableExpr
		if tableName == nil {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErBadFieldError, colName, a.currentClause)
		}
		return tableName, nil
	default:
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErNotSupportedYet, typedE)
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
		colName := string(column.Name)
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
				Database:  []byte(leftDatabaseName),
				Name:      []byte(colName),
				Qualifier: leftTableName,
			}
			rightExpr := &parser.ColName{
				Database:  []byte(rightDatabaseName),
				Name:      []byte(colName),
				Qualifier: rightTableName,
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
		if string(column.Name) == c.columns[i].Name {
			iInUsing = true
		}
		if string(column.Name) == c.columns[j].Name {
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
		seenColumns[string(usingColumn.Name)] = false
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

func resolveJoinKind(kind JoinKind) JoinKind {
	if kind == NaturalJoin {
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

func (a *algebrizer) translateTableExpr(tableExpr parser.TableExpr) (PlanStage, []*Column, error) {
	switch typedT := tableExpr.(type) {
	case *parser.AliasedTableExpr:
		return a.translateSimpleTableExpr(typedT.Expr, string(typedT.As))
	case *parser.ParenTableExpr:
		return a.translateTableExpr(typedT.Expr)
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

		left, leftCols, err := a.translateTableExpr(typedT.LeftExpr)
		if err != nil {
			return nil, nil, err
		}
		right, rightCols, err := a.translateTableExpr(typedT.RightExpr)
		if err != nil {
			return nil, nil, err
		}
		cols := append(leftCols, rightCols...)

		var predicate SQLExpr = SQLTrue
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
							Database:  []byte(filterDb),
							Name:      []byte(leftCol.Name),
							Qualifier: make([]byte, 0),
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
				predicate = SQLTrue
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

		kind = resolveJoinKind(kind)
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

		tableName := strings.ToLower(string(typedT.Name))
		if aliasName == "" {
			aliasName = tableName
		}

		dbName := string(typedT.Qualifier)
		if dbName == "" {
			dbName = a.dbName
		}

		dbName = strings.ToLower(dbName)

		var plan PlanStage
		var err error

		if strings.EqualFold(tableName, "DUAL") {
			plan = NewDualStage()
		} else {
			db, dbErr := a.catalog.Database(dbName)
			if dbErr != nil {
				return nil, nil, dbErr
			}

			table, dbErr := db.Table(tableName)
			if dbErr != nil {
				return nil, nil, dbErr
			}

			switch t := table.(type) {
			case *catalog.MongoTable:
				plan = NewMongoSourceStage(db, t, a.selectID, aliasName)
			case *catalog.DynamicTable:
				plan = NewDynamicSourceStage(db, t, a.selectID, aliasName)
			default:
				return nil, nil, fmt.Errorf("unknown table type: %T", t)
			}
		}

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
	case *parser.Subquery:

		if aliasName == "" && typedT.IsDerived {
			return nil, nil, mysqlerrors.Defaultf(mysqlerrors.ErDerivedMustHaveAlias)
		}

		subqueryAlgebrizer := a.newSubqueryAlgebrizer()

		plan, err := subqueryAlgebrizer.translateSelectStatement(typedT.Select)
		if err != nil {
			return nil, nil, err
		}

		plan = NewSubquerySourceStage(plan, subqueryAlgebrizer.selectID, aliasName)

		// database is not set here because duplicate tables are not allowed in a select query
		// ignoring it avoids false positives
		err = a.registerTable("", aliasName)
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

func (a *algebrizer) translateSubqueryExpr(expr *parser.Subquery) (*SQLSubqueryExpr, error) {
	subqueryAlgebrizer := &algebrizer{
		parent:                      a,
		dbName:                      a.dbName,
		catalog:                     a.catalog,
		selectID:                    a.selectIDGenerator.generate(),
		selectIDGenerator:           a.selectIDGenerator,
		projectedColumnAggregateMap: make(map[int]SQLExpr),
		variables:                   a.variables,
	}

	plan, err := subqueryAlgebrizer.translateSelectStatement(expr.Select)
	if err != nil {
		return nil, err
	}

	return &SQLSubqueryExpr{
		plan:       plan,
		correlated: subqueryAlgebrizer.correlated,
	}, nil
}

func (a *algebrizer) translateExpr(expr parser.Expr) (SQLExpr, error) {
	switch typedE := expr.(type) {
	case *parser.AndExpr:

		left, right, err := a.translateLeftRightExprs(typedE.Left, typedE.Right, false)
		if err != nil {
			return nil, err
		}

		return &SQLAndExpr{left, right}, nil
	case *parser.BinaryExpr:
		left, err := a.translateExpr(typedE.Left)
		if err != nil {
			return nil, err
		}

		right, err := a.translateExpr(typedE.Right)
		if err != nil {
			return nil, err
		}

		leftTy, rightTy := left.Type(), right.Type()
		// Arithmetic with Timestamps should be floating points due to fractional seconds.
		// Arithmetic with Date should be integer.
		if leftTy == schema.SQLTimestamp || rightTy == schema.SQLTimestamp {
			left, _, err = ReconcileSQLExprs(left, SQLDecimal128(decimal.NewFromFloat(0.0)))
			if err != nil {
				return nil, err
			}

			_, right, err = ReconcileSQLExprs(SQLDecimal128(decimal.NewFromFloat(0.0)), right)
			if err != nil {
				return nil, err
			}
		} else if leftTy == schema.SQLDate || rightTy == schema.SQLDate {
			left, _, err = ReconcileSQLExprs(left, SQLInt(0))
			if err != nil {
				return nil, err
			}

			_, right, err = ReconcileSQLExprs(SQLInt(0), right)
			if err != nil {
				return nil, err
			}
		} else {
			left, right, err = ReconcileSQLExprs(left, right)
			if err != nil {
				return nil, err
			}
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
				"No support for binary operator '%v'", string(typedE.Operator))
		}
	case *parser.CaseExpr:
		return a.translateCaseExpr(typedE)
	case *parser.ColName:
		dataBaseName := ""
		if typedE.Database != nil {
			dataBaseName = string(typedE.Database)
		}

		tableName := ""
		if typedE.Qualifier != nil {
			tableName = string(typedE.Qualifier)
		}

		columnName := string(typedE.Name)

		if strings.HasPrefix(tableName,
			"@") || (tableName == "" && strings.HasPrefix(columnName,
			"@")) {
			return a.translateVariableExpr(typedE)
		}

		return a.resolveColumnExpr(dataBaseName, tableName, columnName)
	case *parser.ComparisonExpr:
		reconcile := true
		if (typedE.Operator == parser.AST_EQ && typedE.SubqueryOperator == "") ||
			typedE.Operator == parser.AST_IS ||
			typedE.Operator == parser.AST_IS_NOT ||
			typedE.SubqueryOperator != "" {

			reconcile = false
		}

		left, right, err := a.translateLeftRightExprs(typedE.Left, typedE.Right, reconcile)
		if err != nil {
			return nil, err
		}

		if typedE.Operator == parser.AST_IS {
			return NewSQLIsExpr(left, right), nil
		} else if typedE.Operator == parser.AST_IS_NOT {
			return &SQLNotExpr{NewSQLIsExpr(left, right)}, nil
		}

		if typedE.SubqueryOperator != "" {
			if eval, ok := right.(*SQLSubqueryExpr); ok {
				// Subqueries in predicate can return more than one row during
				// ALL, ANY, IN, NOT_IN, and SOME comparisons
				eval.allowRows = true
				switch typedE.SubqueryOperator {
				case parser.AST_ALL:
					return &SQLSubqueryCmpExpr{subqueryAll, left, eval, typedE.Operator}, nil
				case parser.AST_ANY:
					return &SQLSubqueryCmpExpr{subqueryAny, left, eval, typedE.Operator}, nil
				case parser.AST_IN:
					return &SQLSubqueryCmpExpr{subqueryIn, left, eval, typedE.Operator}, nil
				case parser.AST_NOT_IN:
					return &SQLSubqueryCmpExpr{subqueryNotIn, left, eval, typedE.Operator}, nil
				case parser.AST_SOME:
					return &SQLSubqueryCmpExpr{subquerySome, left, eval, typedE.Operator}, nil
				}
			}
			// this should be unreachable because of the parser
			return nil, mysqlerrors.Newf(mysqlerrors.ErNotSupportedYet,
				"No support for '%v' with '%T'", typedE.SubqueryOperator, right)
		}

		_, leftIsSubquery := left.(*SQLSubqueryExpr)
		_, rightIsSubquery := right.(*SQLSubqueryExpr)
		canTranslate := (!leftIsSubquery && !rightIsSubquery) &&
			(left.Type() == schema.SQLTuple || right.Type() == schema.SQLTuple)

		shouldTranslate := containsString(
			[]string{
				sqlOpNEQ, sqlOpEQ,
				sqlOpGT, sqlOpGTE,
				sqlOpLT, sqlOpLTE,
				sqlOpNSE,
			}, typedE.Operator)

		if canTranslate && shouldTranslate {
			return translateTupleExpr(left, right, typedE.Operator)
		}

		comp, err := comparisonExpr(left, right, typedE.Operator)
		if err != nil {
			return nil, err
		}

		if rec, ok := comp.(reconcilingSQLExpr); ok {
			comp, err = rec.reconcile()
		}
		return comp, err

	case parser.DateVal:

		arg := string(typedE.Val)

		switch typedE.Name {
		case parser.AST_DATE:
			date, _, ok := parseDateTime(arg)
			if !ok || date.Hour() > 0 ||
				date.Minute() > 0 ||
				date.Second() > 0 ||
				date.Nanosecond() > 0 {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongValue, "DATE", arg)
			}
			return SQLDate{date}, nil
		case parser.AST_TIME:
			dur, _, ok := strToTime(arg)
			if !ok {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongValue, "TIME", arg)
			}

			date := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC).Add(dur)

			return SQLTimestamp{date}, nil
		case parser.AST_TIMESTAMP, parser.AST_DATETIME:
			date, _, ok := strToDateTime(arg, true)
			if !ok {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongValue, "DATETIME", arg)
			}
			return SQLTimestamp{date}, nil
		default:
			return nil, mysqlerrors.Newf(mysqlerrors.ErNotSupportedYet,
				"No support for constructor '%v'", typedE.Name)
		}
	case *parser.ExistsExpr:
		subquery, err := a.translateSubqueryExpr(typedE.Subquery)
		if err != nil {
			return nil, err
		}
		return &SQLExistsExpr{subquery}, nil
	case *parser.FalseVal:
		return SQLFalse, nil
	case *parser.FuncExpr:
		return a.translateFuncExpr(typedE)
	case parser.KeywordVal:
		return SQLVarchar(string(typedE)), nil
	case *parser.LikeExpr:
		left, right, err := a.translateLeftRightExprs(typedE.Left, typedE.Right, false)
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
			escape = SQLVarchar("\\")
		}

		expr := &SQLLikeExpr{left, right, escape}

		if typedE.Operator == parser.AST_NOT_LIKE {
			return &SQLNotExpr{expr}, nil
		}

		return expr, nil
	case *parser.NotExpr:
		child, err := a.translateExpr(typedE.Expr)
		if err != nil {
			return nil, err
		}

		return &SQLNotExpr{child}, nil
	case *parser.NullVal:
		return SQLNull, nil
	case parser.NumVal:
		exprString := parser.String(expr)

		useFloats := !a.variables.MongoDBInfo.VersionAtLeast(3, 3, 15)

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
			return SQLFloat(f), nil
		}
		if strings.Contains(exprString, ".") {
			d, err := decimal.NewFromString(exprString)
			if err != nil {
				return nil,
					mysqlerrors.Defaultf(mysqlerrors.ErIllegalValueForType,
						"decimal",
						exprString)
			}
			return SQLDecimal128(d), nil
		}

		// try to parse as int64 first
		if i, err := strconv.ParseInt(exprString, 10, 64); err == nil {
			return SQLInt(i), nil
		}

		// next try to parse as uint64
		if i, err := strconv.ParseUint(exprString, 10, 64); err == nil {
			return SQLUint64(i), nil
		}

		if useFloats {
			f, err := strconv.ParseFloat(exprString, 64)
			if err != nil {
				return nil,
					mysqlerrors.Defaultf(mysqlerrors.ErIllegalValueForType,
						"integer",
						exprString)
			}
			return SQLFloat(f), nil
		}

		i, err := decimal.NewFromString(exprString)
		if err != nil {
			return nil,
				mysqlerrors.Defaultf(mysqlerrors.ErIllegalValueForType,
					"integer",
					exprString)
		}
		return SQLDecimal128(i), nil
	case *parser.OrExpr:

		left, right, err := a.translateLeftRightExprs(typedE.Left, typedE.Right, false)
		if err != nil {
			return nil, err
		}

		return &SQLOrExpr{left, right}, nil
	case *parser.XorExpr:

		left, right, err := a.translateLeftRightExprs(typedE.Left, typedE.Right, false)
		if err != nil {
			return nil, err
		}

		return &SQLXorExpr{left, right}, nil
	case *parser.RangeCond:
		left, err := a.translateExpr(typedE.Left)
		if err != nil {
			return nil, err
		}

		to, err := a.translateExpr(typedE.To)
		if err != nil {
			return nil, err
		}

		from, err := a.translateExpr(typedE.From)
		if err != nil {
			return nil, err
		}

		left, from, err = ReconcileSQLExprs(left, from)
		if err != nil {
			return nil, err
		}

		lower := &SQLGreaterThanOrEqualExpr{left, from}

		left, to, err = ReconcileSQLExprs(left, to)
		if err != nil {
			return nil, err
		}

		upper := &SQLLessThanOrEqualExpr{left, to}

		var m SQLExpr = &SQLAndExpr{lower, upper}

		if typedE.Operator == parser.AST_NOT_BETWEEN {
			return &SQLNotExpr{m}, nil
		}

		return m, nil
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
		return SQLVarchar(string(typedE)), nil
	case *parser.Subquery:
		return a.translateSubqueryExpr(typedE)
	case *parser.TrueVal:
		return SQLTrue, nil
	case *parser.UnaryExpr:

		child, err := a.translateExpr(typedE.Expr)
		if err != nil {
			return nil, err
		}

		switch typedE.Operator {
		case parser.AST_UMINUS:
			switch child.Type() {
			case schema.SQLNull, schema.SQLDecimal128, schema.SQLFloat,
				schema.SQLNumeric, schema.SQLArrNumeric, schema.SQLInt, schema.SQLInt64:
				return &SQLUnaryMinusExpr{child}, nil
			case schema.SQLVarchar:
				return &SQLUnaryMinusExpr{&SQLConvertExpr{child, schema.SQLFloat, SQLNone}}, nil
			}
			return &SQLUnaryMinusExpr{&SQLConvertExpr{child, schema.SQLInt, SQLNone}}, nil
		case parser.AST_TILDA:
			return &SQLUnaryTildeExpr{child}, nil
		case parser.AST_UPLUS:
			return child, nil
		}

		return nil, mysqlerrors.Newf(mysqlerrors.ErNotSupportedYet,
			"No support for operator '%v'", typedE.Operator)
	case *parser.UnknownVal:
		return SQLNull, nil
	case parser.ValTuple:

		var exprs []SQLExpr

		for _, e := range typedE {
			newExpr, err := a.translateExpr(e)
			if err != nil {
				return nil, err
			}

			exprs = append(exprs, newExpr)
		}

		if len(exprs) == 1 {
			// TODO: remove this check from ast_factories.go and add test.
			return exprs[0], nil
		}

		return &SQLTupleExpr{exprs}, nil
	default:
		return nil, mysqlerrors.Newf(mysqlerrors.ErNotSupportedYet,
			"No support for '%v'", parser.String(typedE))
	}
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

func (a *algebrizer) translateLeftRightExprs(
	left, right parser.Expr,
	reconcile bool) (SQLExpr, SQLExpr, error) {
	leftEval, err := a.translateExpr(left)
	if err != nil {
		return nil, nil, err
	}

	rightEval, err := a.translateExpr(right)
	if err != nil {
		return nil, nil, err
	}

	if reconcile {
		leftEval, rightEval, err = ReconcileSQLExprs(leftEval, rightEval)
	}

	return leftEval, rightEval, err
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

			matcher = &SQLEqualsExpr{e, cond}
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
		elseValue = SQLNull
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

func (a *algebrizer) translateFuncExpr(expr *parser.FuncExpr) (SQLExpr, error) {

	exprs := []SQLExpr{}
	name := string(expr.Name)

	if a.isAggFunction(name) {

		if len(expr.Exprs) != 1 {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongArguments, name)
		}

		e := expr.Exprs[0]

		switch typedE := e.(type) {
		case *parser.StarExpr:

			if name != "count" {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongArguments, name)
			}

			if expr.Distinct {
				return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongArguments, name)
			}

			exprs = append(exprs, SQLVarchar("*"))

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

		aggExpr := &SQLAggFunctionExpr{name, expr.Distinct, exprs}

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
		return NewSQLColumnExpr(current.selectID, "", "", aggExpr.String(),
			aggExpr.Type(), schema.MongoNone), nil
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

			if typedE.As != nil {
				as := string(typedE.As)
				switch strings.ToLower(as) {
				case "cast":
					exprs = append(exprs, SQLVarchar(as))
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
	case "rand":
		// We need something unique that we can map.
		id := util.GetUniqueID()
		return NewSQLScalarFunctionExpr("rand", append([]SQLExpr{SQLUint64(id)}, exprs...))
	case "isnull":
		return NewSQLIsExpr(exprs[0], SQLNull), nil
	case "week_day", "last_day", "to_days":
		dateArg, err := NewSQLScalarFunctionExpr("date", exprs)
		if err != nil {
			return nil, err
		}
		return NewSQLScalarFunctionExpr(name, []SQLExpr{dateArg})
	case "to_seconds":
		dateArg, err := NewSQLScalarFunctionExpr("timestamp", exprs)
		if err != nil {
			return nil, err
		}
		return NewSQLScalarFunctionExpr(name, []SQLExpr{dateArg})
	case "date_add", "adddate", "date_sub", "subdate":
		tsArg, err := NewSQLScalarFunctionExpr("timestamp", exprs[0:1])
		if err != nil {
			return nil, err
		}
		if len(exprs) == 2 {
			return NewSQLScalarFunctionExpr(name, []SQLExpr{tsArg, exprs[1], SQLVarchar(Day)})
		}
		return NewSQLScalarFunctionExpr(name, []SQLExpr{tsArg, exprs[1], exprs[2]})
	case "timestampadd":
		tsArg, err := NewSQLScalarFunctionExpr("timestamp", exprs[2:])
		if err != nil {
			return nil, err
		}
		return NewSQLScalarFunctionExpr(name, []SQLExpr{exprs[0], exprs[1], tsArg})
	case "timestampdiff":
		tsArg1, err := NewSQLScalarFunctionExpr("timestamp", exprs[1:2])
		if err != nil {
			return nil, err
		}
		tsArg2, err := NewSQLScalarFunctionExpr("timestamp", exprs[2:3])
		if err != nil {
			return nil, err
		}
		return NewSQLScalarFunctionExpr(name, []SQLExpr{exprs[0], tsArg1, tsArg2})
	case "week", "weekofyear":
		dateArg, err := NewSQLScalarFunctionExpr("date", exprs[0:1])
		if err != nil {
			return nil, err
		}
		if len(exprs) == 2 {
			return NewSQLScalarFunctionExpr("week", []SQLExpr{dateArg, exprs[1]})
		}
		if name == "week" {
			return NewSQLScalarFunctionExpr("week", []SQLExpr{dateArg, SQLInt(0)})
		}
		return NewSQLScalarFunctionExpr("week", []SQLExpr{dateArg, SQLInt(3)})
	case "yearweek":
		dateArg, err := NewSQLScalarFunctionExpr("date", exprs[0:1])
		if err != nil {
			return nil, err
		}
		if len(exprs) == 2 {
			return NewSQLScalarFunctionExpr("yearweek", []SQLExpr{dateArg, exprs[1]})
		}
		return NewSQLScalarFunctionExpr("yearweek", []SQLExpr{dateArg, SQLInt(0)})
	default:
		return NewSQLScalarFunctionExpr(name, exprs)
	}

}

func (a *algebrizer) translateVariableExpr(c *parser.ColName) (*SQLVariableExpr, error) {

	v := &SQLVariableExpr{
		Kind:    variable.SystemKind,
		Scope:   variable.SessionScope,
		sqlType: schema.SQLNone,
	}

	pos := 0
	str := string(c.Name)
	if c.Qualifier != nil {
		str = string(c.Qualifier) + "." + str
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

	value, err := a.variables.Get(variable.Name(v.Name), v.Scope, v.Kind)
	if err != nil {
		return nil, err
	}

	v.sqlType = value.SQLType

	return v, nil
}

type selectIDGatherer struct {
	selectIDs []int
}

func gatherSelectIDs(n Node) []int {
	v := &selectIDGatherer{}
	_, err := v.visit(n)
	if err != nil {
		panic(fmt.Errorf("selectIDGatherer returned unexpected error: %v", err))
	}
	return v.selectIDs
}

func (v *selectIDGatherer) visit(n Node) (Node, error) {
	n, err := walk(v, n)
	if err != nil {
		return nil, err
	}

	switch typedN := n.(type) {
	case SQLColumnExpr:
		v.selectIDs = append(v.selectIDs, typedN.selectID)
	}

	return n, nil
}
