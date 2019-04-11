package analyzer

import (
	"strings"

	"github.com/10gen/mongoast/ast"
)

// IsDotPrefixOfString returns true if the first path is a prefix of the second,
// e.g., a.b is a prefix of a.b.c
// but, aa is not a prefix of aaa.b
func IsDotPrefixOfString(possiblePrefix, totalPath string) bool {
	possiblePrefixParts := strings.Split(possiblePrefix, ".")
	totalPathParts := strings.Split(totalPath, ".")
	if len(possiblePrefixParts) > len(totalPathParts) {
		return false
	}
	for i, part := range possiblePrefixParts {
		if part != totalPathParts[i] {
			return false
		}
	}
	return true
}

func getDotPrefixesOfArray(pathParts []string) []string {
	ret := make([]string, len(pathParts))
	var cur string
	for i, part := range pathParts {
		if i == 0 {
			cur = part
		} else {
			cur += "." + part
		}
		ret[i] = cur
	}
	return ret
}

// GetDotPrefixesOfString returns all the dot prefixes of a path:
// for a.b.c we get a, a.b, and a.b.c.
func GetDotPrefixesOfString(path string) []string {
	pathParts := strings.Split(path, ".")
	return getDotPrefixesOfArray(pathParts)
}

// GetDotPrefixesOfFieldRef returns all the dot prefiexs of a *ast.FieldRef.
func GetDotPrefixesOfFieldRef(ref *ast.FieldRef) []string {
	// This is not the most efficient way to do this, but it reduces code.
	// If performance ever become an issue here (highly unlikely), rewrite this.
	pathPaths := strings.Split(ast.GetDottedFieldName(ref), ".")
	return getDotPrefixesOfArray(pathPaths)
}

// GetPathRootString gets the root name for a path, e.g., a.b.c, the root is a. This
// is an important part of our field definition interface.
func GetPathRootString(path string) string {
	return strings.Split(path, ".")[0]
}

// GetPathRootFromRef gets the root name for a path, e.g., a.b.c, the root is a. This
// is an important part of our field definition interface. Returns true if the root is a
// field, and false if it is not.
func GetPathRootFromRef(field ast.Expr) (string, bool) {
	var cur ast.Expr = field
	for {
		switch typedExpr := cur.(type) {
		case (*ast.FieldRef):
			if typedExpr.Parent == nil {
				return typedExpr.Name, true
			}
			cur = typedExpr.Parent
		case (*ast.FieldOrArrayIndexRef):
			if typedExpr.Parent == nil {
				return "", false
			}
			cur = typedExpr.Parent
		case (*ast.ArrayIndexRef):
			if typedExpr.Parent == nil {
				return "", false
			}
			cur = typedExpr.Parent
		default:
			return "", false
		}
	}
}

// GetVariableRefRootNameFromRef gets the variable root for a path, e.g., $$a.b.c,
// the root is a. Returns true if the root is a variable, and false if it is not.
func GetVariableRefRootNameFromRef(field ast.Expr) (string, bool) {
	var cur ast.Expr = field
	for {
		switch typedExpr := cur.(type) {
		case (*ast.FieldRef):
			if typedExpr.Parent == nil {
				return "", false
			}
			cur = typedExpr.Parent
		case (*ast.FieldOrArrayIndexRef):
			if typedExpr.Parent == nil {
				return "", false
			}
			cur = typedExpr.Parent
		case (*ast.ArrayIndexRef):
			if typedExpr.Parent == nil {
				return "", false
			}
			cur = typedExpr.Parent
		case (*ast.VariableRef):
			return typedExpr.Name, true
		default:
			return "", false
		}
	}
}
