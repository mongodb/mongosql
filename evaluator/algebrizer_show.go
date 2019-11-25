package evaluator

import (
	"strings"

	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
)

func (a *algebrizer) translateShow(show *parser.Show) (PlanStage, error) {

	switch strings.ToLower(show.Section) {
	case "create database":
		return a.translateShowCreateDatabase(show)
	case "create table":
		return a.translateShowCreateTable(show)
	default:
		return nil, mysqlerrors.Newf(mysqlerrors.ErNotSupportedYet,
			"no support for show (%s)", show.Section)
	}
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
			Column: results.NewColumn(a.selectID,
				"",
				"",
				"",
				"Database",
				"",
				"",
				types.EvalString,
				schema.MongoNone,
				false,
				true,
			),
			Expr: NewSQLValueExpr(values.NewSQLVarchar(a.valueKind(), databaseName)),
		},
		ProjectedColumn{
			Column: results.NewColumn(a.selectID,
				"",
				"",
				"",
				"Create Database",
				"",
				"",
				types.EvalString,
				schema.MongoNone,
				false,
				true,
			),
			Expr: NewSQLValueExpr(values.NewSQLVarchar(a.valueKind(), catalog.GenerateCreateDatabase(databaseName,
				show.Modifier))),
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
			Column: results.NewColumn(a.selectID,
				"",
				"",
				"",
				"Table",
				"",
				"",
				types.EvalString,
				schema.MongoNone,
				false,
				true,
			),
			Expr: NewSQLValueExpr(values.NewSQLVarchar(a.valueKind(), table.Name())),
		},
		ProjectedColumn{
			Column: results.NewColumn(a.selectID,
				"",
				"",
				"",
				"Create Table",
				"",
				"",
				types.EvalString,
				schema.MongoNone,
				false,
				true,
			),
			Expr: NewSQLValueExpr(values.NewSQLVarchar(a.valueKind(), createTableSQL)),
		},
	), nil
}
