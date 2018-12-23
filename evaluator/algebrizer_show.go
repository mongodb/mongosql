package evaluator

import (
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/parser"
)

func (a *algebrizer) translateShow(show *parser.Show) (PlanStage, error) {

	switch strings.ToLower(show.Section) {
	case "charset":
		return a.translateShowCharset(show)
	case "collation":
		return a.translateShowCollation(show)
	case "columns":
		return a.translateShowColumns(show)
	case "create database":
		return a.translateShowCreateDatabase(show)
	case "create table":
		return a.translateShowCreateTable(show)
	case "databases", "schemas":
		return a.translateShowDatabases(show)
	case "keys", "index", "indexes":
		return a.translateShowKeys(show)
	case "processlist":
		return a.translateShowProcessList(show)
	case "status":
		return a.translateShowVariables(show, "STATUS")
	case "tables":
		return a.translateShowTables(show)
	case "variables":
		return a.translateShowVariables(show, "VARIABLES")
	default:
		return nil, mysqlerrors.Newf(mysqlerrors.ErNotSupportedYet,
			"no support for show (%s)", show.Section)
	}
}

func (a *algebrizer) translateShowCharset(show *parser.Show) (PlanStage, error) {

	info := showInfo{
		dbName:    catalog.InformationSchemaDatabase,
		tableName: "CHARACTER_SETS",
		columnNames: []string{"CHARACTER_SET_NAME", "DESCRIPTION",
			"DEFAULT_COLLATE_NAME", "MAXLEN"},
		columnAliases: []string{"Charset", "Description", "Default collation", "Maxlen"},
		orderBy:       "Charset",
		predicate:     a.translateShowLikeOrWhere("Charset", show.LikeOrWhere),
	}

	return a.translateShowInfo(&info)
}

func (a *algebrizer) translateShowCollation(show *parser.Show) (PlanStage, error) {

	info := showInfo{
		dbName:    catalog.InformationSchemaDatabase,
		tableName: "COLLATIONS",
		columnNames: []string{"COLLATION_NAME",
			"CHARACTER_SET_NAME",
			"ID", "IS_DEFAULT",
			"IS_COMPILED",
			"SORTLEN"},
		columnAliases: []string{"Collation",
			"Charset",
			"Id",
			"Default",
			"Compiled",
			"Sortlen"},
		orderBy:   "Collation",
		predicate: a.translateShowLikeOrWhere("Collation", show.LikeOrWhere),
	}

	return a.translateShowInfo(&info)
}

func (a *algebrizer) translateShowColumns(show *parser.Show) (PlanStage, error) {
	info := showInfo{
		dbName:    catalog.InformationSchemaDatabase,
		tableName: "COLUMNS",
		columnNames: []string{
			"COLUMN_NAME", "COLUMN_TYPE", "IS_NULLABLE", "COLUMN_KEY",
			"COLUMN_DEFAULT", "EXTRA", "TABLE_NAME", "TABLE_SCHEMA",
			"ORDINAL_POSITION"},
		columnAliases: []string{
			"Field", "Type", "Null", "Key",
			"Default", "Extra"},
		orderBy: "ORDINAL_POSITION",
	}

	if strings.EqualFold(show.Modifier, "full") {
		info.columnNames = append(info.columnNames[:2],
			append([]string{"COLLATION_NAME"},
				info.columnNames[2:]...)...)
		info.columnNames = append(info.columnNames[:7],
			append([]string{"PRIVILEGES",
				"COLUMN_COMMENT"},
				info.columnNames[7:]...)...)
		info.columnAliases = append(info.columnAliases[:2],
			append([]string{"Collation"},
				info.columnAliases[2:]...)...)
		info.columnAliases = append(info.columnAliases,
			"Privileges",
			"Comment")
	}

	dbName := a.cfg.dbName
	table := ""

	switch f := show.From.(type) {
	case parser.StrVal:
		table = string(f)
	case *parser.ColName:
		dbName = f.Qualifier.Else("")
		table = f.Name
	default:
		return nil,
			mysqlerrors.Defaultf(mysqlerrors.ErIllegalValueForType,
				"FROM",
				parser.String(f))
	}

	if dbName == "" {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErNoDbError)
	}

	if db, err := a.cfg.catalog.Database(dbName); err != nil {
		return nil, err
	} else if tbl, err := db.Table(table); err != nil {
		return nil, err
	} else {
		dbName = string(db.Name())
		table = string(tbl.Name())
	}

	info.predicate = &parser.AndExpr{
		Left: &parser.ComparisonExpr{
			Operator: parser.AST_EQ,
			Left: &parser.ColName{
				Name: "TABLE_NAME",
			},
			Right: parser.StrVal([]byte(table)),
		},
		Right: &parser.ComparisonExpr{
			Operator: parser.AST_EQ,
			Left: &parser.ColName{
				Name: "TABLE_SCHEMA",
			},
			Right: parser.StrVal([]byte(dbName)),
		},
	}

	likeWhere := a.translateShowLikeOrWhere("Field", show.LikeOrWhere)
	if likeWhere != nil {
		info.predicate = &parser.AndExpr{
			Left:  info.predicate,
			Right: likeWhere,
		}
	}

	return a.translateShowInfo(&info)
}

func (a *algebrizer) translateShowCreateDatabase(show *parser.Show) (PlanStage, error) {
	dbName := ""

	if show.From != nil {
		switch f := show.From.(type) {
		case parser.StrVal:
			dbName = string(f)
		default:
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErIllegalValueForType,
				"FROM", parser.String(f))
		}
	}

	db, err := a.cfg.catalog.Database(dbName)
	if err != nil {
		return nil, err
	}

	databaseName := string(db.Name())

	return NewProjectStage(
		NewDualStage(),
		ProjectedColumn{
			Column: NewColumn(a.selectID,
				"",
				"",
				"",
				"Database",
				"",
				"",
				EvalString,
				"",
				false,
			),
			Expr: NewSQLVarchar(a.valueKind(), databaseName),
		},
		ProjectedColumn{
			Column: NewColumn(a.selectID,
				"",
				"",
				"",
				"Create Database",
				"",
				"",
				EvalString,
				"",
				false,
			),
			Expr: NewSQLVarchar(a.valueKind(), catalog.GenerateCreateDatabase(databaseName,
				show.Modifier)),
		},
	), nil
}

func (a *algebrizer) translateShowCreateTable(show *parser.Show) (PlanStage, error) {
	dbName := a.cfg.dbName
	tableName := ""

	switch f := show.From.(type) {
	case parser.StrVal:
		tableName = string(f)
	case *parser.ColName:
		dbName = f.Qualifier.Else("")
		tableName = f.Name
	default:
		return nil,
			mysqlerrors.Defaultf(mysqlerrors.ErIllegalValueForType,
				"FROM",
				parser.String(f))
	}

	if dbName == "" {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErNoDbError)
	}

	var table catalog.Table

	if db, err := a.cfg.catalog.Database(dbName); err != nil {
		return nil, err
	} else if table, err = db.Table(tableName); err != nil {
		return nil, err
	}

	createTableSQL := catalog.GenerateCreateTable(table, a.cfg.maxVarcharLength)

	return NewProjectStage(
		NewDualStage(),
		ProjectedColumn{
			Column: NewColumn(a.selectID,
				"",
				"",
				"",
				"Table",
				"",
				"",
				EvalString,
				"",
				false,
			),
			Expr: NewSQLVarchar(a.valueKind(), string(table.Name())),
		},
		ProjectedColumn{
			Column: NewColumn(a.selectID,
				"",
				"",
				"",
				"Create Table",
				"",
				"",
				EvalString,
				"",
				false,
			),
			Expr: NewSQLVarchar(a.valueKind(), createTableSQL),
		},
	), nil
}

func (a *algebrizer) translateShowDatabases(show *parser.Show) (PlanStage, error) {

	info := showInfo{
		dbName:        catalog.InformationSchemaDatabase,
		tableName:     "SCHEMATA",
		columnNames:   []string{"SCHEMA_NAME"},
		columnAliases: []string{"Database"},
		orderBy:       "Database",
		predicate:     a.translateShowLikeOrWhere("Database", show.LikeOrWhere),
	}

	return a.translateShowInfo(&info)
}

func (a *algebrizer) translateShowKeys(show *parser.Show) (PlanStage, error) {
	dbName := a.cfg.dbName
	tableName := ""

	switch f := show.From.(type) {
	case parser.StrVal:
		tableName = string(f)
	case *parser.ColName:
		dbName = f.Qualifier.Else("")
		tableName = f.Name
	default:
		return nil,
			mysqlerrors.Defaultf(mysqlerrors.ErIllegalValueForType, "FROM", parser.String(f))
	}

	if dbName == "" {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErNoDbError)
	}

	if db, err := a.cfg.catalog.Database(dbName); err != nil {
		return nil, err
	} else if tbl, err := db.Table(tableName); err != nil {
		return nil, err
	} else {
		dbName = string(db.Name())
		tableName = string(tbl.Name())
	}

	info := showInfo{
		dbName:    catalog.InformationSchemaDatabase,
		tableName: "STATISTICS",
		columnNames: []string{"TABLE_NAME",
			"NON_UNIQUE",
			"INDEX_NAME",
			"SEQ_IN_INDEX",
			"COLUMN_NAME",
			"COLLATION",
			"CARDINALITY",
			"SUB_PART",
			"PACKED",
			"NULLABLE",
			"INDEX_TYPE",
			"COMMENT",
			"INDEX_COMMENT",
			"TABLE_SCHEMA"},
		columnAliases: []string{"Table",
			"Non_unique",
			"Key_name",
			"Seq_in_index",
			"Column_name",
			"Collation",
			"Cardinality",
			"Sub_part",
			"Packed",
			"Null",
			"Index_type",
			"Comment",
			"Index_comment"},
		orderBy:   "Non_unique",
		predicate: show.LikeOrWhere,
	}

	info.predicate = &parser.AndExpr{
		Left: &parser.ComparisonExpr{
			Operator: parser.AST_EQ,
			Left: &parser.ColName{
				Name: "Table",
			},
			Right: parser.StrVal([]byte(tableName)),
		},
		Right: &parser.ComparisonExpr{
			Operator: parser.AST_EQ,
			Left: &parser.ColName{
				Name: "TABLE_SCHEMA",
			},
			Right: parser.StrVal([]byte(dbName)),
		},
	}

	if show.LikeOrWhere != nil {
		info.predicate = &parser.AndExpr{
			Left:  info.predicate,
			Right: show.LikeOrWhere,
		}
	}

	return a.translateShowInfo(&info)
}

func (a *algebrizer) translateShowTables(show *parser.Show) (PlanStage, error) {
	dbName := a.cfg.dbName

	if show.From != nil {
		switch f := show.From.(type) {
		case parser.StrVal:
			dbName = string(f)
		default:
			return nil,
				mysqlerrors.Defaultf(mysqlerrors.ErIllegalValueForType,
					"FROM",
					parser.String(f))
		}
	}

	var columnName string
	if dbName == "" {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErNoDbError)
	} else if db, err := a.cfg.catalog.Database(dbName); err != nil {
		return nil, err
	} else {
		columnName = "Tables_in_" + dbName
		dbName = string(db.Name())
	}

	info := showInfo{
		dbName:        catalog.InformationSchemaDatabase,
		tableName:     "TABLES",
		columnNames:   []string{"TABLE_NAME", "TABLE_SCHEMA"},
		columnAliases: []string{columnName},
		orderBy:       columnName,
	}

	if strings.EqualFold(show.Modifier, "full") {
		info.columnNames = append(info.columnNames[:1],
			append([]string{"TABLE_TYPE"},
				info.columnNames[1:]...)...)
		info.columnAliases = append(info.columnAliases, "Table_type")
	}

	info.predicate = &parser.ComparisonExpr{
		Operator: parser.AST_EQ,
		Left: &parser.ColName{
			Name: "TABLE_SCHEMA",
		},
		Right: parser.StrVal([]byte(dbName)),
	}

	likeWhere := a.translateShowLikeOrWhere(columnName, show.LikeOrWhere)
	if likeWhere != nil {
		info.predicate = &parser.AndExpr{
			Left:  info.predicate,
			Right: likeWhere,
		}
	}

	return a.translateShowInfo(&info)
}

func (a *algebrizer) translateShowVariables(show *parser.Show, kind string) (PlanStage, error) {
	tableName := strings.ToUpper(show.Modifier)
	kind = strings.ToUpper(kind)

	info := showInfo{
		dbName:        catalog.InformationSchemaDatabase,
		tableName:     fmt.Sprintf("%s_%s", tableName, kind),
		columnNames:   []string{"VARIABLE_NAME", "VARIABLE_VALUE"},
		columnAliases: []string{"Variable_name", "Value"},
		orderBy:       "Variable_name",
		predicate:     a.translateShowLikeOrWhere("Variable_name", show.LikeOrWhere),
	}

	return a.translateShowInfo(&info)
}

func (a *algebrizer) translateShowProcessList(show *parser.Show) (PlanStage, error) {

	var transform map[string]exprTransformer

	// need to truncate to first 100 characters
	if show.Modifier == "" {
		transform = map[string]exprTransformer{
			"Info": func(expr SQLExpr) (SQLExpr, error) {
				return NewSQLScalarFunctionExpr(
					"substring",
					append(
						[]SQLExpr{}, expr,
						NewSQLInt64(a.valueKind(), 1),
						NewSQLInt64(a.valueKind(), 100),
					),
				)
			},
		}
	}

	info := showInfo{
		dbName:        catalog.InformationSchemaDatabase,
		tableName:     "PROCESSLIST",
		columnNames:   []string{"ID", "USER", "HOST", "DB", "COMMAND", "TIME", "STATE", "INFO"},
		columnAliases: []string{"Id", "User", "Host", "db", "Command", "Time", "State", "Info"},
		orderBy:       "Id",

		colExprTransformations: transform,
	}

	return a.translateShowInfo(&info)
}

func (a *algebrizer) translateShowLikeOrWhere(likeColumnName string, expr parser.Expr) parser.Expr {
	if expr == nil {
		return nil
	}

	if strVal, ok := expr.(parser.StrVal); ok {
		expr = &parser.LikeExpr{
			Operator: parser.AST_LIKE,
			Left:     &parser.ColName{Name: likeColumnName},
			Right:    strVal,
			Escape:   parser.StrVal("\\"),
		}
	}

	return expr
}

type exprTransformer func(SQLExpr) (SQLExpr, error)

type showInfo struct {
	dbName                 string
	tableName              string
	columnNames            []string
	columnAliases          []string
	orderBy                string
	predicate              parser.Expr
	colExprTransformations map[string]exprTransformer
}

func (a *algebrizer) translateShowInfo(info *showInfo) (PlanStage, error) {
	subqueryAlgebrizer := a.newSubqueryExprAlgebrizer()
	db, err := subqueryAlgebrizer.cfg.catalog.Database(info.dbName)
	if err != nil {
		panic(err.Error())
	}

	tbl, err := db.Table(info.tableName)
	if err != nil {
		panic(err.Error())
	}

	var plan PlanStage = NewDynamicSourceStage(
		db,
		tbl.(*catalog.DynamicTable),
		subqueryAlgebrizer.selectID,
		string(tbl.Name()),
	)

	columns := Columns(plan.Columns())

	var projectedColumns ProjectedColumns

	for i, columnName := range info.columnNames {
		c, ok := columns.FindByName(columnName)
		if !ok {
			panic(fmt.Sprintf("cannot find column %s", columnName))
		}

		var columnAlias string
		if i < len(info.columnAliases) {
			columnAlias = info.columnAliases[i]
		} else {
			columnAlias = c.Name
		}

		c.OriginalTable = info.tableName
		c.OriginalName = c.Name

		projColumn := ProjectedColumn{c, c.expr()}

		// apply expression colExprTransformations if given
		if info.colExprTransformations != nil {
			transformExpr, ok := info.colExprTransformations[columnAlias]
			if ok {
				var transformation SQLExpr
				transformation, err = transformExpr(projColumn.Expr)
				if err != nil {
					panic(fmt.Sprintf("cannot transform column %s: %v", columnAlias, err))
				}
				projColumn = *c.projectWithExpr(transformation)
			}
		}

		projColumn.Column.Name = columnAlias
		projectedColumns = append(projectedColumns, projColumn)
	}

	subqueryTableName := info.tableName
	plan = NewProjectStage(plan, projectedColumns...)
	plan = NewSubquerySourceStage(plan, subqueryAlgebrizer.selectID,
		info.dbName, subqueryTableName, false)
	err = a.registerTable("", subqueryTableName)
	if err != nil {
		// Previously ignored error should not be possible.
		panic(err)
	}
	err = a.registerColumns(plan.Columns())
	if err != nil {
		// Previously ignored error should not be possible.
		panic(err)
	}
	if info.predicate != nil {
		translated, err := a.translateExpr(info.predicate)
		if err != nil {
			return nil, err
		}
		plan = NewFilterStage(plan, translated)
	}

	if len(info.orderBy) > 0 {
		expr, err := a.resolveColumnExpr(info.dbName, subqueryTableName, info.orderBy)
		if err != nil {
			panic(err.Error())
		}
		plan = NewOrderByStage(plan, &OrderByTerm{
			expr:      expr,
			ascending: true,
		})
	}

	// we might need more columns for the translation than actually
	// need to get returned to the client; for instance, to perform
	// filtering or ordering.
	if len(info.columnNames) > len(info.columnAliases) {
		columns := Columns(plan.Columns())
		var projectedColumns ProjectedColumns
		for _, columnName := range info.columnAliases {
			c, ok := columns.FindByName(columnName)
			if !ok {
				panic(fmt.Sprintf("cannot find column %s", columnName))
			}
			projectedColumn := c.projectAs(c.Name)
			projectedColumn.SelectID = a.selectID
			projectedColumns = append(projectedColumns, projectedColumn)
		}
		plan = NewProjectStage(plan, projectedColumns...)
	}

	return plan, nil
}
