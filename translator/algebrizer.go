package translator

import (
	"fmt"
	"github.com/mongodb/mongo-tools/common/log"
	"github.com/siddontang/mixer/sqlparser"
	"strings"
)

type ParseCtx struct {
	Column []ColumnInfo
	Table  []TableInfo
	Parent *ParseCtx
}

type TableInfo struct {
	Name map[string]string
}

type ColumnInfo struct {
	// using a mapping in case the name is an alias
	// e.g. SELECT a+b AS x FROM foo WHERE x<10;
	Name  map[string]string
	Table string
}

// algebrizeSimpleTableExpr takes a simple table expression and returns its algebrized nodes.
func algebrizeSimpleTableExpr(stExpr sqlparser.SimpleTableExpr) (interface{}, error) {

	switch expr := stExpr.(type) {

	case *sqlparser.TableName:
		// TODO: ignoring qualifier for now
		return sqlparser.String(expr), nil

	case *sqlparser.Subquery:
		return nil, fmt.Errorf("can't handle subquery expression type %T", expr)

	default:
		return nil, fmt.Errorf("can't handle simple table expression type %T", expr)
	}

}

// getTableInfo takes a select expression and returns the table information.
func getTableInfo(tExprs sqlparser.TableExprs) ([]TableInfo, error) {
	tables := []TableInfo{}

	for _, tExpr := range tExprs {

		switch expr := tExpr.(type) {

		case *sqlparser.AliasedTableExpr:
			stExpr, err := algebrizeSimpleTableExpr(expr.Expr)
			if err != nil {
				return nil, fmt.Errorf("getTableInfo error: %v", err)
			}

			if strVal, ok := stExpr.(string); ok {
				name := string(expr.As)
				if name == "" {
					name = strVal
				}
				table := TableInfo{
					Name: map[string]string{name: strVal},
				}
				tables = append(tables, table)
			} else {
				return nil, fmt.Errorf("unsupported simple table expression alias of %T", expr)
			}

		default:
			return nil, fmt.Errorf("can't handle table expression type %T", expr)
		}
	}
	return tables, nil
}

func parseColumnInfo(alias, name string) (columnInfo ColumnInfo) {
	if alias == "" {
		alias = name
	}
	if i := strings.Index(name, "."); i != -1 {
		columnInfo.Table = name[:i]
		columnInfo.Name = map[string]string{alias: name[i+1:]}
	} else {
		columnInfo.Name = map[string]string{alias: name}
	}
	return
}

// getColumnInfo takes a select expression and returns the column information.
func getColumnInfo(exprs sqlparser.SelectExprs) ([]ColumnInfo, error) {
	columns := []ColumnInfo{}

	for i, sExpr := range exprs {
		log.Logf(log.DebugLow, "handling parsed select expr %v: %#v", i, sExpr)

		switch expr := sExpr.(type) {

		case *sqlparser.StarExpr:
			log.Logf(log.DebugLow, "got star expression, fetching all columns")

			if columns == nil {
				return nil, fmt.Errorf("received multiple star expressions in select statement")
			}
			columns = nil

		case *sqlparser.NonStarExpr:
			c, err := translateExpr(expr.Expr, nil)
			if err != nil {
				return nil, err
			}

			alias := string(expr.As)

			switch name := c.(type) {

			case ColName:
				column := parseColumnInfo(alias, name.Value)
				columns = append(columns, column)

			case StrVal:
				column := parseColumnInfo(alias, name.Value)
				columns = append(columns, column)

			case string:
				column := parseColumnInfo(alias, name)
				columns = append(columns, column)

			default:
				return nil, fmt.Errorf("unsupported column type: %T", c)
			}

		default:
			return nil, fmt.Errorf("unreachable path")
		}
	}

	return columns, nil
}
