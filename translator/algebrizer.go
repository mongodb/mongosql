package translator

import (
	"fmt"
	"github.com/mongodb/mongo-tools/common/log"
	"github.com/siddontang/mixer/sqlparser"
)

/*
column_name
    The original name of the column that you wish to alias.
table_name
    The original name of the table that you wish to alias.
alias_name
    The temporary name to assign.


*/

type NodeType int

const (
	Column NodeType = 0
	Table  NodeType = 1
)

type AlgebrizerNodes struct {
	Nodes []AlgebrizerNode
}

type AlgebrizerNode struct {
	nType  NodeType // can be column, table or temporary name
	nName  string
	nAlias string
	depth  int
}

func algebrizeSelectStmt(ss sqlparser.SelectStatement) (*AlgebrizerNodes, error) {
	log.Logf(log.DebugLow, "allez vous")
	algebrizerNodes := &AlgebrizerNodes{}

	switch expr := ss.(type) {

	case *sqlparser.Select:
		for _, e := range expr.From {
			tNodes, err := algebrizeTableExpr(e)
			if err != nil {
				return nil, fmt.Errorf("algebrizing error: %v", err)
			}
			algebrizerNodes.Nodes = append(algebrizerNodes.Nodes, tNodes.Nodes...)
		}
		return algebrizerNodes, nil

	default:
		return nil, nil

	}
}

// TODO: temporarily ignoring scope
func algebrizeTableExpr(tExpr sqlparser.TableExpr) (*AlgebrizerNodes, error) {
	algebrizerNodes := &AlgebrizerNodes{}

	switch expr := tExpr.(type) {

	case *sqlparser.AliasedTableExpr:
		stExpr, err := algebrizeSimpleTableExpr(expr.Expr)
		if err != nil {
			return nil, fmt.Errorf("algebrizing error: %v", err)
		}

		if strVal, ok := stExpr.(string); ok {
			node := AlgebrizerNode{Table, strVal, string(expr.As), 0}
			algebrizerNodes.Nodes = append(algebrizerNodes.Nodes, node)
		} else {
			return nil, fmt.Errorf("unsupported simple table expression alias of %T", expr)
		}
		return algebrizerNodes, nil

	default:
		return nil, fmt.Errorf("can't handle table expression type %T", expr)
	}

}

// algebrizeSimpleTableExpr takes a simple table expression and returns its algebrized nodes.
func algebrizeSimpleTableExpr(stExpr sqlparser.SimpleTableExpr) (interface{}, error) {

	switch expr := stExpr.(type) {

	case *sqlparser.TableName:
		// TODO: ignoring qualifier for now
		return sqlparser.String(expr), nil

	case *sqlparser.Subquery:
		return algebrizeSelectStmt(expr.Select)

	default:
		return nil, fmt.Errorf("can't handle simple table expression type %T", expr)
	}

}
