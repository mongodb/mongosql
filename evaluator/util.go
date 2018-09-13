package evaluator

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/memory"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/internal/util/bsonutil"
	"github.com/10gen/sqlproxy/internal/variable"
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
		return &SQLEqualsExpr{left, right}, nil
	case sqlOpLT:
		return &SQLLessThanExpr{left, right}, nil
	case sqlOpGT:
		return &SQLGreaterThanExpr{left, right}, nil
	case sqlOpLTE:
		return &SQLLessThanOrEqualExpr{left, right}, nil
	case sqlOpGTE:
		return &SQLGreaterThanOrEqualExpr{left, right}, nil
	case sqlOpNEQ:
		return &SQLNotEqualsExpr{left, right}, nil
	case sqlOpNSE:
		return &SQLNullSafeEqualsExpr{left, right}, nil
	case sqlOpIn:
		if eval, ok := right.(*SQLSubqueryExpr); ok {
			return &SQLSubqueryCmpExpr{subqueryIn, left, eval, ""}, nil
		}
		return &SQLInExpr{left, right}, nil
	case sqlOpNotIn:
		if eval, ok := right.(*SQLSubqueryExpr); ok {
			return &SQLSubqueryCmpExpr{subqueryNotIn, left, eval, ""}, nil
		}
		return &SQLNotExpr{&SQLInExpr{left, right}}, nil
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
	return sel.QueryGlobals.Distinct || sel.Where != nil ||
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

func generateDbSetFromColumns(columns []*Column) map[string]struct{} {
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

	// Count(*) is not optimizable if aggregations change the cardinality of a collection.
	if bsonutil.ContainsCardinalityAlteringStages(mongoSource.pipeline) {
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

	if string(countFuncExpr.Name) != "count" || len(countFuncExpr.Exprs) != 1 {
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

func parseMongoFilter(left, right SQLExpr) (bson.M, error) {
	if _, ok := right.(SQLVarchar); !ok {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErCantUseOptionHere, left.String())
	}

	filter := bson.M{}
	err := json.Unmarshal([]byte(right.String()), &filter)
	if err != nil {
		return nil, mysqlerrors.Newf(mysqlerrors.ErParseError, "parse mongo filter error: %s", err)
	}

	return filter, nil
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

	end := util.MinInt(len(strArr1), len(strArr2))
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

// ComputeDocNestingDepthWithMaxDepth computes the maximum nesting depth of a document
// with a depth level at which we can abort early to reduce the cost of checking.
func ComputeDocNestingDepthWithMaxDepth(doc interface{}, maxDepth uint32) uint32 {
	var aux func(interface{}, uint32) uint32
	aux = func(currDoc interface{}, depth uint32) uint32 {
		if depth == MaxDepth {
			return MaxDepth + 1
		}
		var maxChildDepth uint32
		switch typedDoc := currDoc.(type) {
		case []bson.D:
			for _, doc := range typedDoc {
				maxChildDepth = util.MaxUint32(maxChildDepth, aux(doc, depth+1))
			}
			return maxChildDepth
		case []interface{}:
			for _, doc := range typedDoc {
				maxChildDepth = util.MaxUint32(maxChildDepth, aux(doc, depth+1))
			}
			return maxChildDepth
		case []bson.M:
			for _, doc := range typedDoc {
				maxChildDepth = util.MaxUint32(maxChildDepth, aux(doc, depth+1))
			}
			return maxChildDepth
		case bson.M:
			for _, doc := range typedDoc {
				maxChildDepth = util.MaxUint32(maxChildDepth, aux(doc, depth+1))
			}
			return maxChildDepth
		case bson.D:
			for _, doc := range typedDoc {
				maxChildDepth = util.MaxUint32(maxChildDepth, aux(doc.Value, depth+1))
			}
			return maxChildDepth
		default:
			return depth

		}
	}
	return aux(doc, 0)
}

// GetSQLValueKind is a utility function that gets the SQLValueKind to use for new
// SQLValues based on the type_conversion_mode variable in the provided container.
func GetSQLValueKind(vars *variable.Container) SQLValueKind {
	mode := vars.GetString(variable.TypeConversionMode)
	switch mode {
	case variable.MongoSQLTypeConversionMode:
		return MongoSQLValueKind
	case variable.MySQLTypeConversionMode:
		return MySQLValueKind
	default:
		panic(fmt.Errorf("cannot get SQLValueKind for type_conversion_mode %q", mode))
	}
}

// newStageMemoryMonitor creates a child of the connection's memory monitor limited
// to the configured max stage stage.
func newStageMemoryMonitor(ctx *ExecutionCtx, stageName string) (*memory.Monitor, error) {
	maxStageSize := ctx.Variables().GetUInt64(variable.MongoDBMaxStageSize)
	return ctx.MemoryMonitor().CreateChild(stageName, maxStageSize)
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

func cloneColumns(columns []*Column) []*Column {
	newColumns := make([]*Column, len(columns))
	for i, col := range columns {
		newColumns[i] = col.clone()
	}

	return newColumns
}
