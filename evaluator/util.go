package evaluator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/10gen/mongoast/ast"
	astparser "github.com/10gen/mongoast/parser"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/internal/astutil"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/mathutil"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/internal/procutil"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
)

// These constants are used as placeholders for special characters in the field
// names we project in our aggregration pipelines.
const (
	Dot    = "_DOT_"
	Dollar = "_DOLLAR_"
)

const (
	sqlOpEQ    = "="
	sqlOpNEQ   = "!="
	sqlOpNSE   = "<=>"
	sqlOpGT    = ">"
	sqlOpGTE   = ">="
	sqlOpLT    = "<"
	sqlOpLTE   = "<="
	sqlOpIn    = "in"
	sqlOpNotIn = "not in"
)

// These regexes are used to ensure we use valid field names when we're creating
// user variables in $let var blocks. See documentation at:
// https://docs.mongodb.com/master/reference/aggregation-variables/#agg-user-variables
var (
	validStartFieldNameRegex        = regexp.MustCompile(`^[[:lower:]]+$`)
	validFieldNameRegex             = regexp.MustCompile(`^[[:alnum:][:^ascii:]_]+$`)
	replaceInvalidFieldNameRegex    = regexp.MustCompile("[^a-zA-Z0-9]+")
	dollarLetStartReplacementChar   = "z"
	dollarLetGenericReplacementChar = "_"
)

// comparisonExpr returns a SQLExpr formed using op comparison operator.
func comparisonExpr(left, right SQLExpr, op string) (SQLExpr, error) {
	switch op {
	case sqlOpEQ:
		// BI-235: Allow users to pass custom $match stage
		if isMongoFilterExpr(left) {
			filter, err := parseMongoFilter(left, right)
			if err != nil {
				return nil, err
			}
			return &MongoFilterExpr{left.(SQLColumnExpr), right, filter}, nil
		}
		return NewSQLComparisonExpr(EQ, left, right), nil
	case sqlOpLT:
		return NewSQLComparisonExpr(LT, left, right), nil
	case sqlOpLTE:
		return NewSQLComparisonExpr(LTE, left, right), nil
	case sqlOpNEQ:
		return NewSQLComparisonExpr(NEQ, left, right), nil
	case sqlOpNSE:
		return NewSQLNullSafeEqualsExpr(left, right), nil
	case sqlOpIn:
		panic("IN must be eliminated in the desugarer")
	case sqlOpNotIn:
		panic("NOT IN must be eliminated in the desugarer")
	case sqlOpGT:
		panic("GT must be eliminated in the desugarer")
	case sqlOpGTE:
		panic("GTE must be eliminated in the desugarer")
	default:
		return nil,
			mysqlerrors.Newf(mysqlerrors.ErNotSupportedYet,
				"No support for comparison operator '%v'",
				op)
	}
}

// evaluateComparison performs a pairwise comparison of left and right using
// the provided comparison op.
func evaluateComparison(left, right []values.SQLValue, op string, knd values.SQLValueKind,
	collation *collation.Collation) (values.SQLValue, error) {
	// any comparison operator other than null-safe equal will return null
	// if the left or right side has a null value.
	if op != sqlOpNSE {
		if values.HasNullValue(left...) || values.HasNullValue(right...) {
			return values.NewSQLNull(knd), nil
		}
	}

	c, err := values.CompareToPairwise(left, right, collation)
	if err != nil {
		return nil, err
	}

	switch op {
	case sqlOpEQ:
		return values.NewSQLBool(knd, c == 0), nil
	case sqlOpLT:
		return values.NewSQLBool(knd, c == -1), nil
	case sqlOpGT:
		return values.NewSQLBool(knd, c == 1), nil
	case sqlOpLTE:
		return values.NewSQLBool(knd, c <= 0), nil
	case sqlOpGTE:
		return values.NewSQLBool(knd, c >= 0), nil
	case sqlOpNEQ:
		return values.NewSQLBool(knd, c != 0), nil
	case sqlOpNSE:
		return values.NewSQLBool(knd, c == 0), nil
	case sqlOpIn:
		panic("IN must be eliminated in the desugarer")
	case sqlOpNotIn:
		panic("NOT IN must be eliminated in the desugarer")
	default:
		return nil,
			mysqlerrors.Newf(mysqlerrors.ErNotSupportedYet,
				"No support for comparison operator '%v'",
				op)
	}
}

func containsAnyInt(ints []int, test []int) bool {
	for _, value := range test {
		if containsInt(ints, value) {
			return true
		}
	}

	return false
}

func containsCardinalityAlteringClause(sel *parser.Select) bool {
	return sel.QueryGlobals != nil && sel.QueryGlobals.Distinct || sel.Where != nil ||
		sel.GroupBy != nil || sel.Having != nil || sel.Limit != nil
}

func containsInt(ints []int, i int) bool {
	for _, value := range ints {
		if value == i {
			return true
		}
	}
	return false
}

func containsStringFunc(strs []string, str string, f func(string, string) bool) bool {
	for _, n := range strs {
		if f(n, str) {
			return true
		}
	}

	return false
}

func containsString(strs []string, str string) bool {
	return containsStringFunc(strs, str, func(s1, s2 string) bool {
		return s1 == s2
	})
}

func fullyQualifiedTableName(databaseName, tableName string) string {
	qualifiedName := tableName
	if databaseName != "" {
		qualifiedName = databaseName + "." + tableName
	}
	return qualifiedName
}

func generateDbSetFromColumns(columns []*results.Column) map[string]struct{} {
	dbNames := make(map[string]struct{})
	for _, c := range columns {
		dbNames[c.Database] = struct{}{}
	}
	return dbNames
}

// handleError returns a function that wraps receives on errChan.
func handleError(errChan chan error) func(err interface{}) {
	return func(err interface{}) {
		errChan <- fmt.Errorf("%v", err)
	}
}

// insersectionStringSet gives the set intersection of two string sets
func intersectionStringSet(left, right map[string]struct{}) map[string]struct{} {
	ret := make(map[string]struct{})
	for k := range left {
		if _, ok := right[k]; ok {
			ret[k] = struct{}{}
		}
	}
	return ret
}

func isAliasedTableExpr(table parser.TableExprs) bool {
	if len(table) != 1 {
		return false
	}

	aliasedTableExpr, ok := table[0].(*parser.AliasedTableExpr)
	if !ok {
		return false
	}

	if _, ok := aliasedTableExpr.Expr.(*parser.TableName); !ok {
		return false
	}

	return true
}

func isCountOptimizable(sel *parser.Select, plan PlanStage) (*MongoSourceStage, bool) {
	if containsCardinalityAlteringClause(sel) {
		return nil, false
	}

	if !isAliasedTableExpr(sel.From) || !isCountStarExpr(sel.SelectExprs) {
		return nil, false
	}

	mongoSource, ok := plan.(*MongoSourceStage)
	if !ok {
		return nil, false
	}

	// Count(*) is not optimizable on sharded collections.
	if mongoSource.isShardedCollection[mongoSource.collectionNames[0]] {
		return nil, false
	}

	docSlice, err := astutil.DeparsePipeline(mongoSource.pipeline)
	if err != nil {
		return nil, false
	}

	// Count(*) is not optimizable if aggregations change the cardinality of a collection.
	if bsonutil.ContainsCardinalityAlteringStages(docSlice) {
		return nil, false
	}

	return mongoSource, true
}

func isCountStarExpr(sel parser.SelectExprs) bool {
	if len(sel) != 1 {
		return false
	}

	nonStarExpr, ok := sel[0].(*parser.NonStarExpr)
	if !ok {
		return false
	}

	countFuncExpr, ok := nonStarExpr.Expr.(*parser.FuncExpr)
	if !ok {
		return false
	}

	if countFuncExpr.Name != "count" || len(countFuncExpr.Exprs) != 1 {
		return false
	}

	if _, ok := countFuncExpr.Exprs[0].(*parser.StarExpr); !ok {
		return false
	}

	return true
}

func isMongoFilterExpr(expr SQLExpr) bool {
	if colExpr, ok := expr.(SQLColumnExpr); ok {
		if colExpr.columnType.MongoType == schema.MongoFilter {
			return true
		}
	}
	return false
}

// keysStringSet returns a slice of the keys of a string-set (struct{} does not
// implement interface{})
func keysStringSet(set map[string]struct{}) []string {
	keys := make([]string, len(set))

	i := 0
	for k := range set {
		keys[i] = k
		i++
	}
	return keys
}

func parseMongoFilter(left, right SQLExpr) (ast.Expr, error) {
	if valExpr, ok := right.(SQLValueExpr); ok {
		if _, isVarchar := valExpr.Value.(values.SQLVarchar); !isVarchar {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErCantUseOptionHere, left.String())
		}
	} else {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErCantUseOptionHere, left.String())
	}

	return astparser.ParseMatchExprJSON(right.String())
}

// pathStartsWithAny returns true if any of the strings in prefixes is a
// prefix of path
func pathStartsWithAny(prefixes []string, path string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// sharesPrefix returns true if one []string is a prefix of the other
func sharesPrefix(strArr1, strArr2 []string) bool {
	// check that all the strings in strArr1 are equal to the same
	// position string in strArr2 for the length of of the shorter

	end := mathutil.MinInt(len(strArr1), len(strArr2))
	for i := 0; i < end; i++ {
		if strArr1[i] != strArr2[i] {
			return false
		}
	}
	return true
}

// sanitizeFieldName translates any disallowed characters in a field name into
// an appropriate replacement.
func sanitizeFieldName(fieldName string) string {
	r := strings.Replace(fieldName, ".", Dot, -1)
	return strings.Replace(r, "$", Dollar, -1)
}

// sanitizeLetVarName replaces invalid characters in let variables with '_'.
// If the first character is not in [a-z], it will be replaced with 'z'.
func sanitizeLetVarName(varName string) string {
	if !validStartFieldNameRegex.MatchString(string(varName[0])) {
		varName = dollarLetStartReplacementChar + varName[1:]
	}

	if !validFieldNameRegex.MatchString(varName) {
		varName = replaceInvalidFieldNameRegex.ReplaceAllString(varName,
			dollarLetGenericReplacementChar)
	}
	return varName
}

// ComputeDocNestingDepthWithMaxDepth computes the maximum nesting depth of a document
// with a depth level at which we can abort early to reduce the cost of checking.
func ComputeDocNestingDepthWithMaxDepth(doc ast.Expr, maxDepth uint32) uint32 {
	var aux func(ast.Expr, uint32) uint32
	aux = func(current ast.Expr, depth uint32) uint32 {
		if depth == MaxDepth {
			return MaxDepth + 1
		}
		var maxChildDepth uint32
		switch t := current.(type) {
		case *ast.AggExpr:
			// +1 for the expr because of the following nesting structure:
			// {
			//   $expr: <expr> (+1)
			// }
			return aux(t.Expr, depth+1)
		case *ast.Array:
			// +1 for each element because of the following nesting structure:
			// [
			//   <element>, (+1)
			//   ...,
			// ]
			for _, e := range t.Elements {
				maxChildDepth = mathutil.MaxUint32(maxChildDepth, aux(e, depth+1))
			}
			return maxChildDepth
		case *ast.Binary:
			// +2 for each argument because of the following nesting structure:
			//   {
			//     <binop>: [ (+1)
			//       left, (+2)
			//       right,
			//     ]
			//   }
			// Binary has an implicit array for its arguments.
			maxChildDepth = mathutil.MaxUint32(maxChildDepth, aux(t.Left, depth+2))
			maxChildDepth = mathutil.MaxUint32(maxChildDepth, aux(t.Right, depth+2))
			return maxChildDepth
		case *ast.Document:
			// +1 for each element because of the following nesting structure:
			// {
			//   <name>: <expr>, (+1)
			//   ...,
			// }
			for _, e := range t.Elements {
				maxChildDepth = mathutil.MaxUint32(maxChildDepth, aux(e.Expr, depth+1))
			}
			return maxChildDepth
		case *ast.Function:
			// +1 for the argument because of the following nesting structure:
			// {
			//   <func>: <arg> (+1)
			// }
			return aux(t.Arg, depth+1)
		case *ast.Let:
			// +2 for "in" and +3 for "vars" because of the following nesting structure:
			//   {
			//     $let: { (+1)
			//       vars: { (+2)
			//         <var>: <expr>, (+3)
			//         ...
			//       },
			//       in: { (+2)
			//         ...
			//       }
			//     }
			maxChildDepth = mathutil.MaxUint32(maxChildDepth, aux(t.Expr, depth+2))
			for _, v := range t.Variables {
				maxChildDepth = mathutil.MaxUint32(maxChildDepth, aux(v.Expr, depth+3))
			}
			return maxChildDepth
		default:
			return depth
		}
	}
	return aux(doc, 0)
}

// GoValueToSQLValue is only needed for dynamic sources and reading variables
// and a few places in testing. As the name suggests, it converts a go value
// to a SQLValue.
func GoValueToSQLValue(kind values.SQLValueKind, v interface{}) values.SQLValue {
	switch vTyped := v.(type) {
	case nil:
		return values.NewSQLNull(kind)
	case bool:
		return values.NewSQLBool(kind, vTyped)
	case int:
		return values.NewSQLInt64(kind, int64(vTyped))
	case int64:
		return values.NewSQLInt64(kind, vTyped)
	case float64:
		return values.NewSQLFloat(kind, vTyped)
	case uint16:
		return values.NewSQLUint64(kind, uint64(vTyped))
	case uint32:
		return values.NewSQLUint64(kind, uint64(vTyped))
	case uint64:
		return values.NewSQLUint64(kind, vTyped)
	case string:
		return values.NewSQLVarchar(kind, vTyped)
	case variable.Name:
		return values.NewSQLVarchar(kind, string(vTyped))
	default:
		panic(fmt.Sprintf(
			"unexpected go type %T from dynamic source or system variable in GoValueToSQLValue",
			vTyped))
	}
}

// GetMongoDBVersion is a utility function that gets the MongoDB version to use for new
// configurations based on the mongodb_version_compatibility or mongodb_version variable
// in the provided container.
func GetMongoDBVersion(vars *variable.Container) []uint8 {
	compatibilityVersion := vars.GetString(variable.MongoDBVersionCompatibility)
	if len(compatibilityVersion) == 0 {
		compatibilityVersion = vars.GetString(variable.MongoDBVersion)
	}
	version, err := procutil.VersionToSlice(compatibilityVersion)
	if err != nil {
		panic(fmt.Sprintf("invalid version provided: %v", compatibilityVersion))
	}
	return version
}

// GetSQLValueKind is a utility function that gets the values.SQLValueKind to use for new
// SQLValues based on the type_conversion_mode variable in the provided container.
func GetSQLValueKind(vars *variable.Container) values.SQLValueKind {
	mode := vars.GetString(variable.TypeConversionMode)
	switch mode {
	case variable.MongoSQLTypeConversionMode:
		return values.MongoSQLValueKind
	case variable.MySQLTypeConversionMode:
		return values.MySQLValueKind
	default:
		panic(fmt.Errorf("cannot get values.SQLValueKind for type_conversion_mode %q", mode))
	}
}

func absInt64(i int64) int64 {
	// make a mask of the sign bit
	mask := i >> 63
	// toggle the bits if value is negative
	i ^= mask
	// subtracting the mask is the same as adding 1 if the number was negative
	i -= mask
	return i
}

func cloneColumns(columns []*results.Column) []*results.Column {
	newColumns := make([]*results.Column, len(columns))
	for i, col := range columns {
		newColumns[i] = col.Clone()
	}

	return newColumns
}
