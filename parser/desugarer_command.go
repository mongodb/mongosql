package parser

import (
	"fmt"

	"github.com/10gen/sqlproxy/internal/option"
)

// DesugarCommand is a compiler phase that occurs after parsing and before
// algebrization. This phase converts a CST from its input form to an equivalent
// simpler form. Constructs that exist in the input can be wholly removed in the
// output. Operations in this phase should be simple. CSTs leave the deeper
// structure of the query obfuscated and no attempt to uncover it should be made.
func DesugarCommand(statement Statement) (Statement, error) {
	type desugarPass struct {
		pass Walker
		// prePassDebuggingMessage will be printed before the pass, if it is not NoneString.
		prePassDebuggingMessage option.String
		// prePassDebuggingMessage will be printed after the pass, if it is not NoneString.
		postPassDebuggingMessage option.String
	}

	desugarers := []desugarPass{
		{&CreateTableTypeDesugarer{}, option.NoneString(), option.NoneString()},
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

var _ Walker = (*CreateTableTypeDesugarer)(nil)

// CreateTableTypeDesugarer replaces `x IS NOT y` with `NOT(x IS y)`
type CreateTableTypeDesugarer struct{}

// PreVisit is called for every node before its children are walked.
func (*CreateTableTypeDesugarer) PreVisit(current CST) (CST, error) {
	return current, nil
}

// PostVisit is called for every node after its children are walked.
func (*CreateTableTypeDesugarer) PostVisit(current CST) (CST, error) {
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
