package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/internal/option"
	"github.com/10gen/sqlproxy/internal/versionutil"
)

// DesugarStatement is a compiler phase that occurs after parsing and before
// algebrization. This phase converts a CST from its input form to an equivalent
// simpler form. Constructs that exist in the input can be wholly removed in the
// output. Operations in this phase should be simple. CSTs leave the deeper
// structure of the query obfuscated and no attempt to uncover it should be made.
func DesugarStatement(statement Statement, versionCode versionutil.MySQLFixedWidthVersionCode,
	currentDbName string) (Statement, error) {
	type desugarPass struct {
		pass Walker
		// prePassDebuggingMessage will be printed before the pass, if it is not NoneString.
		prePassDebuggingMessage option.String
		// prePassDebuggingMessage will be printed after the pass, if it is not NoneString.
		postPassDebuggingMessage option.String
	}

	desugarers := []desugarPass{
		{&evaluateConditionalComment{versionCode}, option.NoneString(), option.NoneString()},
		{&createTableTypeDesugarer{}, option.NoneString(), option.NoneString()},
		{&namer{}, option.NoneString(), option.NoneString()},
		{&isNotDesugarer{}, option.NoneString(), option.NoneString()},
		{&unwrapSingleTuples{}, option.NoneString(), option.NoneString()},
		{&someToAnyDesugarer{}, option.NoneString(), option.NoneString()},
		{&betweenDesugarer{}, option.NoneString(), option.NoneString()},
		{&ifToCaseDesugarer{}, option.NoneString(), option.NoneString()},
		{&inSubqueryDesugarer{}, option.NoneString(), option.NoneString()},
		{&inListConverter{}, option.NoneString(), option.NoneString()},
		{&subqueryComparisonConverter{}, option.NoneString(), option.NoneString()},
		{&tupleComparisonDesugarer{}, option.NoneString(), option.NoneString()},
		{&makeDualExplicit{}, option.NoneString(), option.NoneString()},
		{&gtDesugarer{}, option.NoneString(), option.NoneString()},
		{&showOrExplainDesugarer{currentDbName: currentDbName}, option.NoneString(), option.NoneString()},
	}

	result := statement.(CST)
	var err error
	for _, pass := range desugarers {
		if pass.prePassDebuggingMessage != option.NoneString() {
			fmt.Printf(pass.prePassDebuggingMessage.Unwrap(), result)
		}
		result, err = Walk(pass.pass, result)
		if err != nil {
			return nil, err
		}
		if pass.postPassDebuggingMessage != option.NoneString() {
			fmt.Printf(pass.postPassDebuggingMessage.Unwrap(), result)
		}
	}

	return result.(Statement), nil
}

// createTableTypeDesugarer desugarers allowed type names into the type names we actually support.
type createTableTypeDesugarer struct{}

// PreVisit is called for every node before its children are walked.
func (*createTableTypeDesugarer) PreVisit(current CST) (CST, error) {
	return current, nil
}

// PostVisit is called for every node after its children are walked.
func (*createTableTypeDesugarer) PostVisit(current CST) (CST, error) {
	colDef, ok := current.(*ColumnDefinition)
	if !ok {
		return current, nil
	}
	switch colDef.Type.BaseType {
	case "bool", "bit":
		if !colDef.Type.Width.IsSome() || colDef.Type.Width.Unwrap() == 1 {
			colDef.Type.BaseType = "boolean"
		} else /* This case is impossible unless the base type was bit */ {
			return nil, fmt.Errorf("bit(n) for n > 1 is not allowed at this time, found n = %d",
				colDef.Type.Width.Unwrap())
		}
	case "datetime":
		colDef.Type.BaseType = "timestamp"
	case "tinyint", "smallint", "integer", "bigint":
		colDef.Type.BaseType = "int"
	case "char", "text", "tinytext", "mediumtext", "longtext":
		colDef.Type.BaseType = "varchar"
	case "double":
		colDef.Type.BaseType = "float"
	}
	return current, nil
}

var _ Walker = (*evaluateConditionalComment)(nil)

// evaluateConditionalComment replaces a ConditionallyExecutableComment with the underlying
// Statement if the version code used is less than or equal to the server version code,
// otherwise, it returns an IgnoredStatement.
type evaluateConditionalComment struct {
	versionCode versionutil.MySQLFixedWidthVersionCode
}

// PreVisit is called for every node before its children are walked.
func (e *evaluateConditionalComment) PreVisit(current CST) (CST, error) {
	if v, ok := current.(*ConditionallyExecutableComment); ok {
		if v.VersionCode <= e.versionCode {
			stmt, err := Parse(v.SQL)
			if err != nil {
				return nil, err
			}
			return stmt, nil
		}
		buf := NewTrackedBuffer(nil)
		v.Format(buf)
		return &IgnoredStatement{
			Statement: UnexecutableComment(buf.String()),
		}, nil
	}
	return current, nil
}

// PostVisit is called for every node after its children are walked.
func (*evaluateConditionalComment) PostVisit(current CST) (CST, error) {
	return current, nil
}

var _ Walker = (*isNotDesugarer)(nil)

// isNotDesugarer replaces `x IS NOT y` with `NOT(x IS y)`
type isNotDesugarer struct{}

// PreVisit is called for every node before its children are walked.
func (*isNotDesugarer) PreVisit(current CST) (CST, error) {
	return current, nil
}

// PostVisit is called for every node after its children are walked.
func (*isNotDesugarer) PostVisit(current CST) (CST, error) {
	cmp, ok := current.(*ComparisonExpr)
	if !ok {
		return current, nil
	}

	switch cmp.Operator {
	case AST_IS_NOT:
		return &NotExpr{
			Expr: &ComparisonExpr{
				Left:     cmp.Left,
				Right:    cmp.Right,
				Operator: AST_IS,
			},
		}, nil
	default:
		return current, nil
	}
}

// makeDualExplicit sets the name of the From field in a Select statement
// to "DUAL" when the user does not explicitly name the DUAL table.
type makeDualExplicit struct{}

// PreVisit is called for every node before its children are walked.
func (*makeDualExplicit) PreVisit(current CST) (CST, error) {
	return current, nil
}

// PostVisit is called for every node after its children are walked.
func (*makeDualExplicit) PostVisit(current CST) (CST, error) {
	if node, isSelect := current.(*Select); isSelect {
		if node.From == nil {
			node.From = TableExprs{&AliasedTableExpr{
				Expr: &TableName{Qualifier: option.NoneString(),
					Name: "DUAL"},
				As:    option.NoneString(),
				Hints: nil},
			}
		}
	}
	return current, nil
}

var _ Walker = (*makeDualExplicit)(nil)

// betweenDesugarer replaces BETWEEN (and NOT BETWEEN) with comparisons.
type betweenDesugarer struct{}

// PreVisit is called for every node before its children are walked.
func (*betweenDesugarer) PreVisit(current CST) (CST, error) {
	return current, nil
}

// PostVisit is called for every node after its children are walked.
func (*betweenDesugarer) PostVisit(current CST) (CST, error) {
	if node, isRangeCond := current.(*RangeCond); isRangeCond {
		switch node.Operator {
		case AST_BETWEEN:
			return &AndExpr{
				Left: &ComparisonExpr{
					Left:     node.Left,
					Operator: AST_GE,
					Right:    node.From,
				},
				Right: &ComparisonExpr{
					Left:     node.Left,
					Operator: AST_LE,
					Right:    node.To,
				},
			}, nil
		case AST_NOT_BETWEEN:
			return &OrExpr{
				Left: &ComparisonExpr{
					Left:     node.Left,
					Operator: AST_LT,
					Right:    node.From,
				},
				Right: &ComparisonExpr{
					Left:     node.Left,
					Operator: AST_GT,
					Right:    node.To,
				},
			}, nil
		}
	}
	return current, nil
}

var _ Walker = (*betweenDesugarer)(nil)

// gtDesugarer replaces GT with LT and GE with LE
type gtDesugarer struct{}

// PreVisit is called for every node before its children are walked.
func (*gtDesugarer) PreVisit(current CST) (CST, error) {
	return current, nil
}

// flip returns the appropriate operator after flipping its direction.
// GT becomes LT and GE becomes LE.
func flip(op string) string {
	switch op {
	case AST_GT:
		return AST_LT
	case AST_GE:
		return AST_LE
	}

	panic(fmt.Sprintf("not supposed to flip %v in the desugarer", op))
}

// PostVisit is called for every node after its children are walked.
// Desugaring is skipped if the current node is a subquery.
func (*gtDesugarer) PostVisit(current CST) (CST, error) {
	cmp, ok := current.(*ComparisonExpr)
	if !ok {
		return current, nil
	}
	switch cmp.Operator {
	case AST_GT, AST_GE:
		_, leftIsSubquery := cmp.Left.(*Subquery)
		_, rightIsSubquery := cmp.Right.(*Subquery)

		if leftIsSubquery || rightIsSubquery {
			return current, nil
		}

		return &ComparisonExpr{
			Left:     cmp.Right,
			Operator: flip(cmp.Operator),
			Right:    cmp.Left,
		}, nil
	default:
		return current, nil
	}
}

var _ Walker = (*gtDesugarer)(nil)

// showOrExplainDesugarer replaces a show or explain statement with an equivalent select statement
// that selects the same information from the information_schema database.
// An "explain tbl_name" statement is synonymous with a "show columns from tbl_name" statement.
type showOrExplainDesugarer struct {
	currentDbName string
}

// showInfo stores desugaring information for each show statement:
// - currentDbName:         Current name of the database that the show command is being issued from.
// - innerSelectColNames:   Slice of columns to select/project from fromTableName.
// - innerSelectColAliases: Slice of corresponding aliases for each column in innerSelectColNames.
// - outerSelectColAliases: Slice of column aliases in innerSelectColAliases to retain (some show
//                          statements have 'full' modifiers which, when specified, increases the
//                          final number of projected columns).
// - fromTableName:         The table to query from in the information_schema database.
// - orderByColName:        The final column to order by in outerSelectColAliases.
// - selectWhereExpr:       A where expression that combines any user-specified like/where
//                          expression and any where expression required for desugaring.
type showInfo struct {
	currentDbName         string
	innerSelectColNames   []string
	innerSelectColAliases []string
	outerSelectColAliases []string
	fromTableName         string
	orderByColName        string
	selectWhereExpr       Expr
}

// PreVisit is called for every node before its children are walked.
func (*showOrExplainDesugarer) PreVisit(current CST) (CST, error) {
	return current, nil
}

// PostVisit is called for every node after its children are walked.
// Desugaring is skipped if the current node is a subquery.
func (showCtx *showOrExplainDesugarer) PostVisit(current CST) (CST, error) {
	// Desugar only if the node is a show statement or an "explain tbl_name [col_name]" statement,
	// where col_name is optional.
	// If we have an "explain tbl_name [col_name]" statement, translate that first into an
	// equivalent "show columns from tbl_name [like col_name]" statement.
	var show *Show
	var explain *Explain
	switch current.(type) {
	case *Show:
		show = current.(*Show)
	case *Explain:
		explain = current.(*Explain)
		show = desugarExplainIntoShowColumns(explain)
		if show == nil {
			return current, nil
		}
	default:
		return current, nil
	}

	// Pass the current database name to all desugaring subcalls
	sInfo := &showInfo{}
	sInfo.currentDbName = showCtx.currentDbName

	// Desugar into a select expression accordingly based on the show statement's section.
	switch strings.ToLower(show.Section) {
	case "charset":
		return desugarShowCharset(sInfo, show), nil
	case "collation":
		return desugarShowCollation(sInfo, show), nil
	case "columns":
		return desugarShowColumns(sInfo, show)
	case "databases", "schemas":
		return desugarShowDatabases(sInfo, show), nil
	case "keys", "index", "indexes":
		return desugarShowKeys(sInfo, show)
	case "processlist":
		return desugarShowProcessList(sInfo, show), nil
	case "status":
		return desugarShowStatus(sInfo, show), nil
	case "tables":
		return desugarShowTables(sInfo, show)
	case "variables":
		return desugarShowVariables(sInfo, show), nil
	default:
		// We may have a "show create database/table" command, so defer to the algebrizer for those.
		return current, nil
	}
}

// desugarShowCharset desugars:
//     show charset [where/like "blah"]
// into:
//     select Charset, Description, ...
//     from (select CHARACTER_SET_NAME as Charset, DESCRIPTION as Description, ...
//           from information_schema.CHARACTER_SETS)
//     [where/like "blah"]
//     order by Charset
//
// (Note: Brackets indicate an optional clause.)
func desugarShowCharset(sInfo *showInfo, show *Show) *Select {
	sInfo.innerSelectColNames = []string{"CHARACTER_SET_NAME", "DESCRIPTION", "DEFAULT_COLLATE_NAME", "MAXLEN"}
	sInfo.innerSelectColAliases = []string{"Charset", "Description", "Default collation", "Maxlen"}
	sInfo.outerSelectColAliases = sInfo.innerSelectColAliases
	sInfo.fromTableName = "CHARACTER_SETS"
	sInfo.orderByColName = "Charset"
	return buildSelectExpr(sInfo, show)
}

// desugarShowCollation desugars:
//     show collation [where/like "blah"]
// into:
//     select Collation, Charset, ...
//     from (select COLLATION_NAME as Collation, CHARACTER_SET_NAME as Charset, ...
//           from information_schema.COLLATIONS)
//     [where/like "blah"]
//     order by Collation
//
// (Note: Brackets indicate an optional clause.)
func desugarShowCollation(sInfo *showInfo, show *Show) *Select {
	sInfo.innerSelectColNames = []string{"COLLATION_NAME", "CHARACTER_SET_NAME", "ID", "IS_DEFAULT",
		"IS_COMPILED", "SORTLEN"}
	sInfo.innerSelectColAliases = []string{"Collation", "Charset", "Id", "Default", "Compiled", "Sortlen"}
	sInfo.outerSelectColAliases = sInfo.innerSelectColAliases
	sInfo.fromTableName = "COLLATIONS"
	sInfo.orderByColName = "Collation"
	return buildSelectExpr(sInfo, show)
}

// desugarExplainIntoShowColumns desugars:
//     explain tbl_name [col_name]
// into:
//     show columns from tbl_name [like col_name]
//
// (Note: Brackets indicate an optional argument.)
func desugarExplainIntoShowColumns(explain *Explain) *Show {
	// We don't desugar unless we have an "explain tbl_name [col_name]" statement.
	if !strings.EqualFold(explain.Section, "table") {
		return nil
	}

	// Construct and return the equivalent show statement.
	show := &Show{
		Section: "columns",
		From:    StrVal(explain.Table.Name),
	}
	if explain.Column != nil {
		show.Predicate = &ShowPredicate{}
		show.Predicate.Like = option.SomeString(explain.Column.Name)
	}
	return show
}

// desugarShowColumns desugars:
//     show [full] columns from tbl_name [from db_name] [where/like "blah"]
// into:
//     select Field, Type, ...
//     from (select COLUMN_NAME as Field, COLUMN_TYPE as Type, ...
//           from information_schema.COLUMNS)
//     where TABLE_NAME like tbl_name and TABLE_SCHEMA like db_name [and where/like "blah"]
//     order by ORDINAL_POSITION
//
// (Note: Brackets indicate optional clauses/arguments. If db_name is omitted, the current
//        database is assumed.)
func desugarShowColumns(sInfo *showInfo, show *Show) (*Select, error) {
	sInfo.innerSelectColNames = []string{"COLUMN_NAME", "COLUMN_TYPE", "COLLATION_NAME",
		"IS_NULLABLE", "COLUMN_KEY", "COLUMN_DEFAULT", "EXTRA", "PRIVILEGES", "COLUMN_COMMENT",
		"TABLE_NAME", "TABLE_SCHEMA", "ORDINAL_POSITION"}
	sInfo.innerSelectColAliases = []string{"Field", "Type", "Collation", "Null", "Key", "Default",
		"Extra", "Privileges", "Comment", "TABLE_NAME", "TABLE_SCHEMA", "ORDINAL_POSITION"}
	sInfo.fromTableName = "COLUMNS"
	sInfo.orderByColName = "ORDINAL_POSITION"

	// Include the Collation, Privileges, and Comment columns if the user requests them by
	// specifying the 'full' modifier. Otherwise omit them.
	// Always omit the TABLE_NAME, TABLE_SCHEMA, AND ORDINAL_POSITION columns as they are only
	// relevant (but necessary) to our where/order by clauses for the inner select.
	if strings.EqualFold(show.Modifier, "full") {
		sInfo.outerSelectColAliases = []string{"Field", "Type", "Collation", "Null", "Key",
			"Default", "Extra", "Privileges", "Comment"}
	} else {
		sInfo.outerSelectColAliases = []string{"Field", "Type", "Null", "Key", "Default", "Extra"}
	}

	// Parse any provided table name and database name.
	tableName, dbName, err := parseTableNameAndDbName(sInfo, show)
	if err != nil {
		return nil, err
	}

	// Make like expressions for case-insensitive matches of the table name and database name.
	sInfo.selectWhereExpr = &AndExpr{
		Left: &LikeExpr{
			Operator: AST_LIKE,
			Left:     &ColName{Name: "TABLE_NAME"},
			Right:    StrVal(tableName),
			Escape:   StrVal("\\"),
		},
		Right: &LikeExpr{
			Operator: AST_LIKE,
			Left:     &ColName{Name: "TABLE_SCHEMA"},
			Right:    StrVal(dbName),
			Escape:   StrVal("\\"),
		},
	}

	return buildSelectExpr(sInfo, show), nil
}

// desugarShowDatabases desugars:
//     show databases [where/like "blah"]
// into:
//     select Database
//     from (select SCHEMA_NAME as Database
//           from information_schema.SCHEMATA)
//     [where/like "blah"]
//     order by Database
//
// (Note: Brackets indicate an optional clause.)
func desugarShowDatabases(sInfo *showInfo, show *Show) *Select {
	sInfo.innerSelectColNames = []string{"SCHEMA_NAME"}
	sInfo.innerSelectColAliases = []string{"Database"}
	sInfo.outerSelectColAliases = sInfo.innerSelectColAliases
	sInfo.fromTableName = "SCHEMATA"
	sInfo.orderByColName = "Database"
	return buildSelectExpr(sInfo, show)
}

// desugarShowKeys desugars:
//     show keys from tbl_name [from db_name] [where/like "blah"]
// into:
//     select Field, Type, ...
//     from (select COLUMN_NAME as Field, COLUMN_TYPE as Type, ...
//           from information_schema.COLUMNS)
//     where TABLE_NAME like tbl_name and TABLE_SCHEMA like db_name [and where/like "blah"]
//     order by ORDINAL_POSITION
//
// (Note: Brackets indicate optional clauses. If db_name is omitted, the current
//        database is assumed.)
func desugarShowKeys(sInfo *showInfo, show *Show) (*Select, error) {
	sInfo.innerSelectColNames = []string{"TABLE_NAME", "NON_UNIQUE", "INDEX_NAME", "SEQ_IN_INDEX",
		"COLUMN_NAME", "COLLATION", "CARDINALITY", "SUB_PART", "PACKED", "NULLABLE", "INDEX_TYPE",
		"COMMENT", "INDEX_COMMENT", "TABLE_SCHEMA"}
	sInfo.innerSelectColAliases = []string{"Table", "Non_unique", "Key_name", "Seq_in_index",
		"Column_name", "Collation", "Cardinality", "Sub_part", "Packed", "Null", "Index_type",
		"Comment", "Index_comment", "TABLE_SCHEMA"}
	sInfo.fromTableName = "STATISTICS"
	sInfo.orderByColName = "Non_unique"

	// Omit the TABLE_SCHEMA column as it's only relevant (but necessary) to our where clause
	// for the inner select.
	sInfo.outerSelectColAliases = sInfo.innerSelectColAliases[:len(sInfo.innerSelectColAliases)-1]

	// Parse any provided table name and database name.
	tableName, dbName, err := parseTableNameAndDbName(sInfo, show)
	if err != nil {
		return nil, err
	}

	// Make an expression for a case-insensitive match of the table name and database name.
	sInfo.selectWhereExpr = &AndExpr{
		Left: &LikeExpr{
			Operator: AST_LIKE,
			Left:     &ColName{Name: "Table"},
			Right:    StrVal(tableName),
			Escape:   StrVal("\\"),
		},
		Right: &LikeExpr{
			Operator: AST_LIKE,
			Left:     &ColName{Name: "TABLE_SCHEMA"},
			Right:    StrVal(dbName),
			Escape:   StrVal("\\"),
		},
	}

	return buildSelectExpr(sInfo, show), nil
}

// desugarShowProcessList desugars:
//     show processlist [where/like "blah"]
// into:
//     select Id, User, ...
//     from (select ID as Id, USER as User, ...
//           from information_schema.PROCESSLIST)
//     [where/like "blah"]
//     order by Id
//
// (Note: Brackets indicate an optional clause.)
func desugarShowProcessList(sInfo *showInfo, show *Show) *Select {
	sInfo.innerSelectColNames = []string{"ID", "USER", "HOST", "DB", "COMMAND", "TIME", "STATE", "INFO"}
	sInfo.innerSelectColAliases = []string{"Id", "User", "Host", "db", "Command", "Time", "State", "Info"}
	sInfo.outerSelectColAliases = sInfo.innerSelectColAliases
	sInfo.fromTableName = "PROCESSLIST"
	sInfo.orderByColName = "Id"
	return buildSelectExpr(sInfo, show)
}

// desugarShowStatus desugars:
//     show status [where/like "blah"]
// into:
//     select Variable_name, Value
//     from (select VARIABLE_NAME as Variable_name, VARIABLE_VALUE as Value
//           from information_schema.SESSION_STATUS)
//     [where/like "blah"]
//     order by Variable_name
//
// (Note: Brackets indicate an optional clause.)
func desugarShowStatus(sInfo *showInfo, show *Show) *Select {
	sInfo.innerSelectColNames = []string{"VARIABLE_NAME", "VARIABLE_VALUE"}
	sInfo.innerSelectColAliases = []string{"Variable_name", "Value"}
	sInfo.outerSelectColAliases = sInfo.innerSelectColAliases
	sInfo.fromTableName = fmt.Sprintf("%s_%s", strings.ToUpper(show.Modifier), "STATUS")
	sInfo.orderByColName = "Variable_name"
	return buildSelectExpr(sInfo, show)
}

// desugarShowTables desugars:
//     show [full] tables [from db_name] [where/like "blah"]
// into:
//     select Tables_in_db_name, Table_type, ...
//     from (select TABLE_NAME as Tables_in_db_name, TABLE_TYPE as Table_type, ...
//           from information_schema.TABLES)
//     [where TABLE_SCHEMA like db_name] [and where/like "blah"]
//     order by Tables_in_db_name
//
// (Note: Brackets indicate optional clauses/arguments. If db_name is omitted, the current
//        database is assumed.)
func desugarShowTables(sInfo *showInfo, show *Show) (*Select, error) {
	// Parse any provided database name.
	_, dbName, err := parseTableNameAndDbName(sInfo, show)
	if err != nil {
		return nil, err
	}

	// Construct the output column name by appending the database name.
	outputColName := "Tables_in_" + dbName

	sInfo.innerSelectColNames = []string{"TABLE_NAME", "TABLE_TYPE", "TABLE_SCHEMA"}
	sInfo.innerSelectColAliases = []string{outputColName, "Table_type", "TABLE_SCHEMA"}
	sInfo.fromTableName = "TABLES"
	sInfo.orderByColName = outputColName

	// If the 'full' modifier is specified, output the Tables_in_dbName and Table_type columns
	// (this modifier simply projects an additional column here).
	// Otherwise just output the Tables_in_dbName column.

	// Include the Table_type column if the user requests it by specifying the 'full' modifier.
	// Otherwise omit it.
	// Always omit the TABLE_SCHEMA column as it's only relevant (but necessary) to our where
	// clause for the inner select.
	if strings.EqualFold(show.Modifier, "full") {
		sInfo.outerSelectColAliases = []string{outputColName, "Table_type"}
	} else {
		sInfo.outerSelectColAliases = []string{outputColName}
	}

	// Make an expression for a case-insensitive match of the database name.
	sInfo.selectWhereExpr = &LikeExpr{
		Operator: AST_LIKE,
		Left:     &ColName{Name: "TABLE_SCHEMA"},
		Right:    StrVal(dbName),
		Escape:   StrVal("\\"),
	}

	return buildSelectExpr(sInfo, show), nil
}

// desugarShowVariables desugars:
//     show [global | session] variables [where/like "blah"]
// into:
//     select Variable_name, Value
//     from (select VARIABLE_NAME as Variable_name, VARIABLE_VALUE as Value
//           from information_schema.<MODIFIER>_VARIABLES)
//     [where/like "blah"]
//     order by Variable_name
//
// (Note: Brackets indicate an optional clause.)
func desugarShowVariables(sInfo *showInfo, show *Show) *Select {
	sInfo.innerSelectColNames = []string{"VARIABLE_NAME", "VARIABLE_VALUE"}
	sInfo.innerSelectColAliases = []string{"Variable_name", "Value"}
	sInfo.outerSelectColAliases = sInfo.innerSelectColAliases
	sInfo.fromTableName = fmt.Sprintf("%s_%s", strings.ToUpper(show.Modifier), "VARIABLES")
	sInfo.orderByColName = "Variable_name"
	return buildSelectExpr(sInfo, show)
}

// buildSelectExpr returns an equivalent select expression for a show expression with the
// following components:
// - A subquery returning all relevant information_schema columns with their show command aliases.
//   The show command is an outlier by only allowing users to reference column aliases, rather
//   than the actual column names in the information_schema database. We use a subquery here to
//   allow column aliases to resolve first, so that any user-specified where/like clause can
//   reference a column by its alias in the outer select statement.
// - An outer query that selects the relevant columns by their aliases.
// - A where clause that combines any user-specified like/where clause, and any where clause
//   needed from desugaring (e.g. "show tables" requires us to match the database name to the
//   TABLE_SCHEMA column in information_schema).
// - An order by clause to sort the data accordingly.
func buildSelectExpr(sInfo *showInfo, show *Show) *Select {
	return &Select{
		SelectExprs: buildSelectColExpr(sInfo.outerSelectColAliases),
		From: []TableExpr{
			&AliasedTableExpr{
				Expr: &Subquery{
					Select: &Select{
						SelectExprs: buildSelectColAliasExpr(sInfo.innerSelectColNames, sInfo.innerSelectColAliases),
						From: []TableExpr{
							&AliasedTableExpr{
								Expr: &TableName{
									Qualifier: option.SomeString("information_schema"),
									Name:      sInfo.fromTableName,
								},
							},
						},
					},
				},
			},
		},
		Where: buildSelectWhereExpr(sInfo.selectWhereExpr, show.Predicate, sInfo.outerSelectColAliases[0]),
		OrderBy: []*Order{
			{
				Expr:      &ColName{Name: sInfo.orderByColName},
				Direction: AST_ASC,
			},
		},
	}
}

// buildSelectColExpr builds an array of columns names in the form of a []SelectExpr.
func buildSelectColExpr(colNames []string) []SelectExpr {
	if len(colNames) == 0 {
		panic("column names is empty")
	}

	var colExpr []SelectExpr
	for _, colName := range colNames {
		colExpr = append(colExpr, &NonStarExpr{
			Expr: &ColName{Name: colName},
		})
	}

	return colExpr
}

// buildSelectColAliasExpr builds an array of columns names and their aliases in the form of a []SelectExpr.
func buildSelectColAliasExpr(colNames []string, colAliases []string) []SelectExpr {
	if len(colNames) == 0 || len(colAliases) == 0 {
		panic(fmt.Sprintf("either column names (len=%d) or column aliases (len=%d) is empty",
			len(colNames), len(colAliases)))
	}

	if len(colNames) != len(colAliases) {
		panic(fmt.Sprintf("number of column names (len=%d) differs from column aliases (len=%d)",
			len(colNames), len(colAliases)))
	}

	var colExpr []SelectExpr
	for i, colName := range colNames {
		colExpr = append(colExpr, &NonStarExpr{
			Expr: &ColName{Name: colName},
			As:   option.SomeString(colAliases[i]),
		})
	}

	return colExpr
}

// buildSelectWhereExpr builds a where expression for desugaring a show statement by going through
// all cases for combining any user-specified like/where clause and any where clause from desugaring.
func buildSelectWhereExpr(selectWhereExpr Expr, showPred *ShowPredicate, likeColName string) *Where {
	// If user inputted no like/where clause, just return the where clause generated from desugaring (if any).
	if showPred == nil {
		if selectWhereExpr == nil {
			return nil
		}
		return NewWhere(AST_WHERE, selectWhereExpr)
	}

	// Stores the user's like or where clause.
	var showLikeOrWhereExpr Expr

	// Make a LikeExpr if the user provides a like clause. A like clause consisting of just a string
	// (e.g. show tables like "%blah%") references likeColName (typically the left-most column).
	if showPred.Like.IsSome() {
		showLikeOrWhereExpr = &LikeExpr{
			Operator: AST_LIKE,
			Left:     &ColName{Name: likeColName},
			Right:    StrVal(showPred.Like.Unwrap()),
			Escape:   StrVal("\\"),
		}
	} else {
		// If the user specified no like clause, store a possible where clause. Note that it's not
		// possible for like and where clauses to be specified simultaneously. If they are used in
		// conjunction (e.g. show tables where Tables_in_tbl like "%blah%"), the like acts as a
		// two-value comparison operator for the where clause rather than its own standalone clause.
		showLikeOrWhereExpr = showPred.Where
	}

	// If no where clause was generated from desugaring, return the user's like or where clause.
	if selectWhereExpr == nil {
		return NewWhere(AST_WHERE, showLikeOrWhereExpr)
	}

	// Otherwise we combine the clauses using an AndExpr.
	return NewWhere(AST_WHERE, &AndExpr{
		Left:  selectWhereExpr,
		Right: showLikeOrWhereExpr,
	})
}

// parseTableNameAndDbName parses the database name and table name for "show columns" and
// "show keys" (or just the database name for "show tables") with the following steps:
// - Assume the desired database is the current database and that a table name will be specified,
//   if required for the provided show command.
// - If the 'from' clause is a string (e.g. "show columns/keys from tbl" or "show tables from db"),
//   store that accordingly as either the table name or the database name.
// - Otherwise if the 'from' clause simultaneously specifies a table name and a database name
//   (e.g. "show columns/keys from tbl from db"), store that table name and overwrite the database name.
// - Return an error if the 'from' clause is an invalid type, or if there's no current database
//   and the user never specified one.
// (Note: if f is a *ColName, the parser ensures that f.Qualifier has a value)
func parseTableNameAndDbName(sInfo *showInfo, show *Show) (string, string, error) {
	if sInfo == nil || show == nil {
		panic(fmt.Sprintf("either *showInfo (nil=%t) or *Show (nil=%t) is nil",
			sInfo == nil, show == nil))
	}

	// Parsing the table name and database name only applies to certain show statements.
	showSection := strings.ToLower(show.Section)
	switch showSection {
	case "columns", "keys", "index", "indexes", "tables":
		break
	default:
		panic(fmt.Sprintf("'show %s' cannot be parsed for table name or database name", showSection))
	}

	// Parse any user-specified table name or database name in the 'from' clause.
	tableName := ""
	dbName := sInfo.currentDbName
	if show.From != nil {
		switch f := show.From.(type) {
		case StrVal:
			if strings.EqualFold(showSection, "tables") {
				dbName = string(f)
			} else {
				tableName = string(f)
			}
		case *ColName:
			tableName = f.Name
			dbName = f.Qualifier.Expect("parser allowed column qualifier with no value in 'from' clause")
		default:
			return "", "", mysqlerrors.Defaultf(mysqlerrors.ErIllegalValueForType, "FROM", String(f))
		}
	}

	if dbName == "" {
		return "", "", mysqlerrors.Defaultf(mysqlerrors.ErNoDbError)
	}

	return tableName, dbName, nil
}

var _ Walker = (*showOrExplainDesugarer)(nil)

// ifToCaseDesugarer replaces IF scalar functions with CaseExprs.
type ifToCaseDesugarer struct{}

// PreVisit is called for every node before its children are walked.
func (*ifToCaseDesugarer) PreVisit(current CST) (CST, error) {
	return current, nil
}

func desugarCoalesce(node *FuncExpr) (CST, error) {
	if len(node.Exprs) < 1 {
		return node, nil
	}
	caseConditions := make([]*When, len(node.Exprs))
	for i, expr := range node.Exprs {
		caseConditions[i] = &When{
			Cond: &NotExpr{
				Expr: &ComparisonExpr{
					Operator: AST_IS,
					Left:     expr.(*NonStarExpr).Expr,
					Right:    &NullVal{},
				},
			},
			Val: expr.(*NonStarExpr).Expr,
		}
	}
	return &CaseExpr{
		Expr:  nil,
		Whens: caseConditions,
		Else:  nil,
	}, nil
}

func desugarElt(node *FuncExpr) (CST, error) {
	if len(node.Exprs) < 2 {
		return node, nil
	}
	caseConditions := make([]*When, len(node.Exprs)-1)
	for i, expr := range node.Exprs[1:] {
		caseConditions[i] = &When{
			Cond: &ComparisonExpr{
				Operator: AST_EQ,
				Left:     node.Exprs[0].(*NonStarExpr).Expr,
				Right:    NumVal(strconv.Itoa(i + 1)),
			},
			Val: expr.(*NonStarExpr).Expr,
		}
	}
	return &CaseExpr{
		Expr:  nil,
		Whens: caseConditions,
		Else:  nil,
	}, nil
}

func desugarField(node *FuncExpr) (CST, error) {
	if len(node.Exprs) < 2 {
		return node, nil
	}
	caseConditions := make([]*When, len(node.Exprs)-1)
	for i, expr := range node.Exprs[1:] {
		caseConditions[i] = &When{
			Cond: &ComparisonExpr{
				Operator: AST_EQ,
				Left:     node.Exprs[0].(*NonStarExpr).Expr,
				Right:    expr.(*NonStarExpr).Expr,
			},
			Val: NumVal(strconv.Itoa(i + 1)),
		}
	}
	return &CaseExpr{
		Expr:  nil,
		Whens: caseConditions,
		Else:  NumVal("0"),
	}, nil
}

func desugarIf(node *FuncExpr) (CST, error) {
	if len(node.Exprs) != 3 {
		return node, nil
	}
	return &CaseExpr{
		Expr: nil,
		Whens: []*When{
			{Cond: node.Exprs[0].(*NonStarExpr).Expr,
				Val: node.Exprs[1].(*NonStarExpr).Expr,
			},
		},
		Else: node.Exprs[2].(*NonStarExpr).Expr,
	}, nil
}

func desugarIfNull(node *FuncExpr) (CST, error) {
	if len(node.Exprs) != 2 {
		return node, nil
	}
	return &CaseExpr{
		Expr: nil,
		Whens: []*When{
			{Cond: &ComparisonExpr{
				Operator: AST_IS,
				Left:     node.Exprs[0].(*NonStarExpr).Expr,
				Right:    &NullVal{},
			},
				Val: node.Exprs[1].(*NonStarExpr).Expr,
			},
		},
		Else: node.Exprs[0].(*NonStarExpr).Expr,
	}, nil
}

func desugarInterval(node *FuncExpr) (CST, error) {
	if len(node.Exprs) < 2 {
		return node, nil
	}
	caseConditions := make([]*When, len(node.Exprs))
	caseConditions[0] = &When{
		Cond: &ComparisonExpr{
			Operator: AST_IS,
			Left:     node.Exprs[0].(*NonStarExpr).Expr,
			Right:    &NullVal{},
		},
		Val: NumVal("-1"),
	}
	for i, expr := range node.Exprs[1:] {
		caseConditions[i+1] = &When{
			Cond: &ComparisonExpr{
				Operator: AST_LT,
				Left:     node.Exprs[0].(*NonStarExpr).Expr,
				Right:    expr.(*NonStarExpr).Expr,
			},
			Val: NumVal(strconv.Itoa(i)),
		}
	}
	return &CaseExpr{
		Expr:  nil,
		Whens: caseConditions,
		Else:  NumVal(strconv.Itoa(len(caseConditions) - 1)),
	}, nil
}

func desugarNullIf(node *FuncExpr) (CST, error) {
	if len(node.Exprs) != 2 {
		return node, nil
	}
	return &CaseExpr{
		Expr: nil,
		Whens: []*When{
			{Cond: &ComparisonExpr{
				Operator: AST_EQ,
				Left:     node.Exprs[0].(*NonStarExpr).Expr,
				Right:    node.Exprs[1].(*NonStarExpr).Expr,
			},
				Val: &NullVal{},
			},
		},
		Else: node.Exprs[0].(*NonStarExpr).Expr,
	}, nil
}

// PostVisit is called for every node after its children are walked.
func (*ifToCaseDesugarer) PostVisit(current CST) (CST, error) {
	if node, isFunc := current.(*FuncExpr); isFunc {
		switch strings.ToLower(node.Name) {
		case "coalesce":
			return desugarCoalesce(node)
		case "elt":
			return desugarElt(node)
		case "field":
			return desugarField(node)
		case "if":
			return desugarIf(node)
		case "interval":
			return desugarInterval(node)
		case "ifnull":
			return desugarIfNull(node)
		case "nullif":
			return desugarNullIf(node)
		}
	}
	return current, nil
}

var _ Walker = (*ifToCaseDesugarer)(nil)

// unwrapSingleTuples is a desugarer that removes single-element tuples
// generated by the parser. Desugarers orchestrates a desugaring phase on the
// CST by implementing the Walker interface.
type unwrapSingleTuples struct{}

// PreVisit is called for every node before its children are walked.
func (*unwrapSingleTuples) PreVisit(current CST) (CST, error) {
	return current, nil
}

// PostVisit is called for every node after its children are walked.
func (*unwrapSingleTuples) PostVisit(current CST) (CST, error) {
	return detupleWrappedExpr(current), nil
}

var _ Walker = (*unwrapSingleTuples)(nil)

// detupleWrappedExpr removes tuples that were placed around expressions in the
// parser where parentheses existed.
//
// Note: This should not be necessary. However, our parser interprets every set
// of parentheses (even those in arithmetic exprs, for example) as a tuple, so
// we need to get rid of them.
func detupleWrappedExpr(node CST) CST {
	if tuple, isTuple := node.(ValTuple); isTuple && len(tuple) == 1 {
		return tuple[0]
	}
	return node
}

// someToAnyDesugarer replaces SOME with ANY, as they are aliases.
type someToAnyDesugarer struct{}

// PreVisit is called for every node before its children are walked.
func (*someToAnyDesugarer) PreVisit(current CST) (CST, error) {
	return current, nil
}

// PostVisit is called for every node after its children are walked.
func (*someToAnyDesugarer) PostVisit(current CST) (CST, error) {
	if node, isComp := current.(*ComparisonExpr); isComp {
		if node.SubqueryOperator == AST_SOME {
			return &ComparisonExpr{
				Operator:         node.Operator,
				Left:             node.Left,
				Right:            node.Right,
				SubqueryOperator: AST_ANY,
			}, nil
		}
	}
	return current, nil
}

var _ Walker = (*someToAnyDesugarer)(nil)

// inSubqueryDesugarer replaces IN (subquery) with = ANY (subquery)
// and NOT IN (subquery) with <> ALL (subquery).
type inSubqueryDesugarer struct{}

// PreVisit is called for every node before its children are walked.
func (*inSubqueryDesugarer) PreVisit(current CST) (CST, error) {
	return current, nil
}

// PostVisit is called for every node after its children are walked.
func (*inSubqueryDesugarer) PostVisit(current CST) (CST, error) {
	if node, isComp := current.(*ComparisonExpr); isComp {
		switch node.SubqueryOperator {
		case AST_NOT_IN:
			return &ComparisonExpr{
				AST_NE,
				node.Left,
				node.Right,
				AST_ALL,
			}, nil
		case AST_IN:
			return &ComparisonExpr{
				AST_EQ,
				node.Left,
				node.Right,
				AST_ANY,
			}, nil
		}
	}
	return current, nil
}

var _ Walker = (*inSubqueryDesugarer)(nil)

// inListConverter is a desugarer that breakes IN lists into boolean
// comparisons.
// Desugarers orchestrates a desugaring phase on the CST by implementing
// the Walker interface.
type inListConverter struct{}

// PreVisit is called for every node before its children are walked. PreVisit
// desugars NOT IN nodes to IN nodes, which will themselves be desugared further
// in the PostVisit function.
func (*inListConverter) PreVisit(current CST) (CST, error) {
	if node, isComp := current.(*ComparisonExpr); isComp {
		if node.Operator == AST_NOT_IN {
			// ignore NOT IN subquery, that is a different expression
			if _, isSub := node.Right.(*Subquery); !isSub {
				current = breakUpNotIn(node)
			}
		}
	}
	return current, nil
}

// PostVisit is called for every node after its children are walked.
func (*inListConverter) PostVisit(current CST) (CST, error) {
	if node, isComp := current.(*ComparisonExpr); isComp {
		if node.Operator == AST_IN {
			// ignore IN subquery, that is a different expression
			if _, isSub := node.Right.(*Subquery); !isSub {
				if tuple, isTuple := node.Right.(ValTuple); isTuple {
					current = inListToDisjunction(node.Left, tuple)
				} else {
					current = inListToDisjunction(node.Left, []Expr{node.Right})
				}
			}
		}
	}
	return current, nil
}

var _ Walker = (*inListConverter)(nil)

// breakUpNotIn rewrites a NOT IN list expression as a boolean NOT of an IN list
// expression. NOT IN list is of the form a NOT IN (b, c...) and is expressable
// as NOT (a IN (b, c...))
func breakUpNotIn(node *ComparisonExpr) Expr {
	node.Operator = AST_IN
	return &NotExpr{node}
}

// detupleEquality rewrites an IN list expression as a disjunction by enumerating every
// equality comparison.
// IN list is of the form a in (b, c...) and is expressable as
// a = b OR a = c OR ...
func inListToDisjunction(leftExpr Expr, rightExprs ValTuple) Expr {
	var makeDisjunction func(leftExpr Expr, rightExprs ValTuple) Expr
	makeDisjunction = func(leftExpr Expr, rightExprs ValTuple) Expr {
		if len(rightExprs) == 1 {
			return &ComparisonExpr{
				AST_EQ,
				leftExpr,
				rightExprs[0],
				"",
			}
		}
		return &OrExpr{
			&ComparisonExpr{
				AST_EQ,
				leftExpr,
				rightExprs[0],
				"",
			},
			makeDisjunction(leftExpr.Copy().(Expr), rightExprs[1:]),
		}
	}
	return makeDisjunction(leftExpr, rightExprs)
}

// tupleComparisonDesugarer is a desugarer that rewrites multi-element tuples into other
// expressions. For example, the tupleComparisonDesugarer will rewrite `select (a, b) <
// (c, d) from foo` as `select a < c or a = c and b < d from foo`.
type tupleComparisonDesugarer struct{}

// PreVisit is called for every node before its children are walked.
func (*tupleComparisonDesugarer) PreVisit(current CST) (CST, error) {
	return current, nil
}

// PostVisit is called for every node after its children are walked.
func (*tupleComparisonDesugarer) PostVisit(current CST) (CST, error) {
	if node, isComp := current.(*ComparisonExpr); isComp {
		_, leftIsTuple := node.Left.(ValTuple)
		_, rightIsTuple := node.Right.(ValTuple)

		if leftIsTuple || rightIsTuple {
			var err error
			current, err = detupleToBooleanCompare(node)
			if err != nil {
				return nil, err
			}
		}
	}
	return current, nil
}

var _ Walker = (*tupleComparisonDesugarer)(nil)

// Tuples are only legal in comparisons.
// Tuple comparisons are either equivalent to conjunctions or disjunctions
// of comparisons or subquery comparisons.
// This function handles the first case.
func detupleToBooleanCompare(node *ComparisonExpr) (Expr, error) {
	// Check outermost left side to ensure it is a tuple.
	left, leftIsTuple := node.Left.(ValTuple)
	if !leftIsTuple {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, 1)
	}

	// Ensure both sides are uniform. And extract values from nests.
	var checkLengthAndExtract func(left ValTuple, rightUnchecked Expr) (Exprs, Exprs, error)
	checkLengthAndExtract = func(left ValTuple, rightUnchecked Expr) (Exprs, Exprs, error) {
		var leftExprs Exprs
		var rightExprs Exprs

		// Make sure the right hand side is a tuple first.
		right, rightIsTuple := rightUnchecked.(ValTuple)
		if !rightIsTuple {
			return nil, nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, len(left))
		}

		// Make sure the lengths are the same.
		if len(left) != len(right) {
			return nil, nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, len(left))
		}

		// Collect values in turn from each tuple.
		for i, leftExpr := range left {
			rightExpr := right[i]

			// Recursively descend into left side if a tuple is located.
			leftElem, innerLeftIsTuple := leftExpr.(ValTuple)
			if innerLeftIsTuple {
				leftAdditions, rightAdditions, err := checkLengthAndExtract(leftElem, rightExpr)
				if err != nil {
					return nil, nil, err
				}
				leftExprs = append(leftExprs, leftAdditions...)
				rightExprs = append(rightExprs, rightAdditions...)
			} else {
				// If the right hand side is a tuple here, it doesn't match the expression
				// on the left.
				if _, innerRightIsTuple := rightExpr.(ValTuple); innerRightIsTuple {
					return nil, nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, 1)
				}
				leftExprs = append(leftExprs, leftExpr)
				rightExprs = append(rightExprs, rightExpr)
			}
		}

		return leftExprs, rightExprs, nil
	}

	leftExprs, rightExprs, err := checkLengthAndExtract(left, node.Right)
	if err != nil {
		return nil, err
	}
	switch node.Operator {
	case string(AST_EQ), string(AST_NSE):
		return detupleEquality(node.Operator, leftExprs, rightExprs), nil
	case string(AST_NE):
		eq := detupleEquality(string(AST_EQ), leftExprs, rightExprs)
		return &NotExpr{eq}, nil
	default:
		return detupleInequality(node.Operator, leftExprs, rightExprs), nil
	}
}

// detupleEquality rewrites a tuple as a conjunction by moving from the former to the later:
// Tuple equality expressions of the form (a1, b1, ...) op (a2, b2, ...) are expressable as
// a1 op a2 AND b1 op b2 AND ...
func detupleEquality(operator string, leftExprs, rightExprs Exprs) Expr {
	var makeConjunction func(operator string, leftExprs, rightExprs Exprs) Expr
	makeConjunction = func(operator string, leftExprs, rightExprs Exprs) Expr {
		if len(leftExprs) == 1 {
			return &ComparisonExpr{
				operator,
				leftExprs[0],
				rightExprs[0],
				"",
			}
		}
		return &AndExpr{
			&ComparisonExpr{
				operator,
				leftExprs[0],
				rightExprs[0],
				"",
			},
			makeConjunction(operator, leftExprs[1:], rightExprs[1:]),
		}
	}
	return makeConjunction(operator, leftExprs, rightExprs)
}

// detupleInequality rewrites a tuple as a disjunction by moving from the former to the later:
// Tuple inequality expressions of the form (a1, b1, ...) op (a2, b2, ...) are expressable as
// a1 op a2 OR (a1 = a2 AND b1 op b2 OR (...))
func detupleInequality(operator string, leftExprs, rightExprs Exprs) Expr {
	var makeDisjunction func(operator string, leftExprs, rightExprs Exprs) Expr
	makeDisjunction = func(operator string, leftExprs, rightExprs Exprs) Expr {
		if len(leftExprs) == 1 {
			return &ComparisonExpr{
				operator,
				leftExprs[0],
				rightExprs[0],
				"",
			}
		}
		return &OrExpr{
			&ComparisonExpr{
				operator,
				leftExprs[0],
				rightExprs[0],
				"",
			},
			&AndExpr{
				&ComparisonExpr{
					AST_EQ,
					leftExprs[0],
					rightExprs[0],
					"",
				},
				makeDisjunction(operator, leftExprs[1:], rightExprs[1:]),
			},
		}
	}

	switch operator {
	case string(AST_LE):
		operator = string(AST_LT)
	case string(AST_GE):
		operator = string(AST_GT)
	}

	return makeDisjunction(operator, leftExprs, rightExprs)
}

// detupleToSubquery rewrites a tuple to a subquery. Tuples are only legal in
// comparisons. Tuple comparisons are either equivalent to conjunctions or
// disjunctions of comparisons or subquery comparisons. This function handles
// the second case.
func detupleToSubquery(node ValTuple) *Subquery {
	selExprs := make(SelectExprs, len(node))
	for i, expr := range node {
		selExprs[i] = &NonStarExpr{Expr: expr}
	}
	return &Subquery{
		&Select{SelectExprs: selExprs, QueryGlobals: &QueryGlobals{}},
		false,
	}
}

// toTuple returns the Expr as a tuple. If it already is one, it is returned
// unchanged; if it is not, it is returned as a single-value tuple.
func toTuple(e Expr) ValTuple {
	if tuple, isTuple := e.(ValTuple); isTuple {
		return tuple
	}
	return ValTuple{e}
}

// subqueryComparisonConverter is a desugarer that rewrites comparison
// expressions that have a subquery on one side and not the other to
// have subqueries on both sides.
type subqueryComparisonConverter struct{}

var _ Walker = (*subqueryComparisonConverter)(nil)

// PreVisit is called for every node before its children are walked.
func (*subqueryComparisonConverter) PreVisit(current CST) (CST, error) {
	return current, nil
}

// PostVisit is called for every node after its children are walked.
func (*subqueryComparisonConverter) PostVisit(current CST) (CST, error) {
	if node, isComp := current.(*ComparisonExpr); isComp {
		_, leftIsSubquery := node.Left.(*Subquery)
		_, rightIsSubquery := node.Right.(*Subquery)

		// Even non-tuple comparison operands are converted into subqueries,
		// as this allows us to always produce full subquery comparisons
		// where the left and right operands are both subqueries.
		if !leftIsSubquery && rightIsSubquery {
			node.Left = detupleToSubquery(toTuple(node.Left))
		} else if !rightIsSubquery && leftIsSubquery {
			node.Right = detupleToSubquery(toTuple(node.Right))
		}
	}
	return current, nil
}
