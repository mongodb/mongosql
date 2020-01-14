package ast

import (
	"fmt"
	"strings"
)

// GetDottedFieldName gets the dotted field name for any ref type, except for ArrayIndexRef.
func GetDottedFieldName(ref Expr) string {
	return strings.Join(GetDottedFields(ref), ".")
}

// GetDottedFields gets the dotted field parts of ref.
func GetDottedFields(ref Expr) []string {
	paths := []string{}
	cur := ref
	for cur != nil {
		switch typedRef := cur.(type) {
		case *FieldOrArrayIndexRef:
			paths = append(paths, fmt.Sprintf("%d", typedRef.Number))
			cur = typedRef.Parent
		case *FieldRef:
			paths = append(paths, typedRef.Name)
			cur = typedRef.Parent
		case *VariableRef:
			paths = append(paths, "$"+typedRef.Name)
			cur = nil
		default:
			panic(fmt.Sprintf("malformed manually created reference type containing a non reference parent with type %T", cur))
		}
	}

	// Reverse the paths array since we are actually adding parts from right to left.
	for i, j := 0, len(paths)-1; i < j; i, j = i+1, j-1 {
		paths[i], paths[j] = paths[j], paths[i]
	}
	return paths
}

// SplitDottedFieldPath splits a dotted field path into its constituent field parts.
func SplitDottedFieldPath(path string) []string {
	// NOTE: This may not always be true (SERVER-30575), and should be changed if
	// such escaping conventions are ever introduced. This function gives us a
	// unifying and singular place to make changes if the time comes, but for
	// now, it is just a call to strings.Split().
	return strings.Split(path, ".")
}

// IsPureFieldRef returns true if a FieldRef doesn't contains other type of references.
func IsPureFieldRef(fieldRef *FieldRef) bool {
	if fieldRef.Parent == nil {
		return true
	} else if parentFieldRef, ok := fieldRef.Parent.(*FieldRef); ok {
		return IsPureFieldRef(parentFieldRef)
	}
	return false
}
