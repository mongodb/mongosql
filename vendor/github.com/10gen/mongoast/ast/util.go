package ast

import (
	"fmt"
)

// GetDottedFieldName gets the dotted field name for any ref type, except for ArrayIndexRef.
func GetDottedFieldName(ref Expr) string {
	path := ""
	cur := ref
	for cur != nil {
		switch typedRef := cur.(type) {
		case *FieldOrArrayIndexRef:
			path = fmt.Sprintf("%d.%s", typedRef.Number, path)
			cur = typedRef.Parent
		case *FieldRef:
			path = fmt.Sprintf("%s.%s", typedRef.Name, path)
			cur = typedRef.Parent
		case *VariableRef:
			path = fmt.Sprintf("$%s.%s", typedRef.Name, path)
			cur = nil
		default:
			panic(fmt.Sprintf("malformed manually created reference type containing a non reference parent with type %T", cur))
		}
	}
	return path[:len(path)-1]
}
