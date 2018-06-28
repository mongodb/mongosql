package evaluator

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/memory"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
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

var bsonDType = reflect.TypeOf(bson.D{})

// numberedDoc gives an enumeration for bson.D's. This allows us to retain the
// original pipeline stage number for a bson.D that we have projected out (such
// as if we want all $unwinds, or all $addFields)
type numberedDoc struct {
	number int
	doc    bson.D
}

type unwindInfo struct {
	stageNumber int
	path        string
	index       string
}

func (in *unwindInfo) getPath() string {
	return in.path
}

func (in *unwindInfo) getIndex() string {
	return in.index
}

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

func containsCardinalityAlteringStages(pipeline []bson.D) bool {
	for _, doc := range pipeline {
		for k := range doc.Map() {
			switch k {
			case "$addFields":
				continue
			case "$graphLookup":
				continue
			case "$lookup":
				continue
			case "$out":
				continue
			case "$project":
				continue
			case "$replaceRoot":
				continue
			case "$sort":
				continue
			default:
				return true
			}
		}
	}

	return false
}

func containsInt(ints []int, i int) bool {
	for _, value := range ints {
		if value == i {
			return true
		}
	}
	return false
}

func containsCardinalityAlteringClause(sel *parser.Select) bool {
	return sel.Distinct == parser.AST_DISTINCT || sel.Where != nil ||
		sel.GroupBy != nil || sel.Having != nil || sel.Limit != nil
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

// ExtractFieldByName takes a field name and document, and returns a value representing
// the value of that field in the document in a format that can be printed as a string.
// It will also handle dot-delimited field names for nested arrays or documents.
func ExtractFieldByName(fieldName string, document interface{}) (interface{}, bool) {
	dotParts := strings.Split(fieldName, ".")
	var subdoc interface{} = document

	for _, path := range dotParts {
		docValue := reflect.ValueOf(subdoc)
		if !docValue.IsValid() {
			return nil, false
		}
		docType := docValue.Type()
		docKind := docType.Kind()
		if docKind == reflect.Map {
			subdocVal := docValue.MapIndex(reflect.ValueOf(path))
			if subdocVal.Kind() == reflect.Invalid {
				return nil, false
			}
			subdoc = subdocVal.Interface()
		} else if docKind == reflect.Slice {
			if docType == bsonDType {
				// dive into a D as a document
				asD := subdoc.(bson.D)
				var err error
				subdoc, err = findValueByKey(path, &asD)
				if err != nil {
					return nil, false
				}
			} else {
				//  check that the path can be converted to int
				arrayIndex, err := strconv.Atoi(path)
				if err != nil {
					return nil, false
				}
				// bounds check for slice
				if arrayIndex < 0 || arrayIndex >= docValue.Len() {
					return nil, false
				}
				subdocVal := docValue.Index(arrayIndex)
				if subdocVal.Kind() == reflect.Invalid {
					return nil, false
				}
				subdoc = subdocVal.Interface()
			}
		} else {
			// trying to index into a non-compound type - just return blank.
			return nil, false
		}
	}
	return subdoc, true
}

// findValueByKey returns the value of keyName in document. If keyName is not found
// in the top-level of the document, ErrNoSuchField is returned as the error.
func findValueByKey(keyName string, document *bson.D) (interface{}, error) {
	for _, key := range *document {
		if key.Name == keyName {
			return key.Value, nil
		}
	}
	return nil, fmt.Errorf("no such field: %v", keyName)
}

// findUnwindForPaths finds an unwind in an []unwindInfo that has the proper
// unwind path
func findUnwindForPath(unwinds []unwindInfo, path string) (unwindInfo, bool) {
	for _, unwind := range unwinds {
		if unwind.path == path {
			return unwind, true
		}
	}
	return unwindInfo{stageNumber: -1, path: "", index: ""}, false
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

// nolint: unparam
func getPipelineStages(stage string, pipeline []bson.D) []*numberedDoc {
	ret := make([]*numberedDoc, 0)
	for i, doc := range pipeline {
		if _, ok := doc.Map()[stage]; ok {
			ret = append(ret, &numberedDoc{number: i, doc: doc})
		}
	}
	return ret
}

func getPipelineUnwinds(pipeline []bson.D) []*numberedDoc {
	return getPipelineStages("$unwind", pipeline)
}

func getPaths(in []unwindInfo) []string {
	return getFields(in, (*unwindInfo).getPath)
}

func getIndexes(in []unwindInfo) []string {
	return getFields(in, (*unwindInfo).getIndex)
}

func getFields(in []unwindInfo, m func(v *unwindInfo) string) []string {
	ret := make([]string, len(in))
	for i, v := range in {
		ret[i] = m(&v)
	}
	return ret
}

// getPipelineUnwindFields get all the unwind fields for a pipeline, in order
func getPipelineUnwindFields(pipeline []bson.D) []unwindInfo {
	unwinds := getPipelineUnwinds(pipeline)
	ret := make([]unwindInfo, len(unwinds))
	var path string
	var index string
	for i, numberedDoc := range unwinds {
		doc := numberedDoc.doc
		unwind := doc.Map()["$unwind"]
		unwindDoc, ok := unwind.(bson.D)
		var fields bson.M
		if ok {
			fields = unwindDoc.Map()
		} else {
			fields = unwind.(bson.M)
		}
		path = fields["path"].(string)
		if index, ok = fields["includeArrayIndex"].(string); !ok {
			index = ""
		}
		ret[i] = unwindInfo{stageNumber: numberedDoc.number, path: path, index: index}
	}
	return ret
}

// getUnwindSuffix will give the remaining unwinds for two slices of unwinds
// after matching on unwind path.
func getUnwindSuffix(unwinds1, unwinds2 []unwindInfo) ([]unwindInfo, bool) {
	ret := make([]unwindInfo, 0)
	end := util.MinInt(len(unwinds1), len(unwinds2))
	i := 0
	for ; i < end; i++ {
		// Prefixes are incompatible, so there is no suffix
		// don't check index, assume that is correct
		if unwinds1[i].path != unwinds2[i].path {
			return nil, false
		}
	}
	var tail []unwindInfo
	if len(unwinds1) <= len(unwinds2) {
		tail = unwinds2
	} else {
		tail = unwinds1

	}
	for ; i < len(tail); i++ {
		ret = append(ret, tail[i])
	}
	return ret, true
}

// handleError returns a function that wraps receives on errChan.
func handleError(errChan chan error) func(err interface{}) {
	return func(err interface{}) {
		errChan <- fmt.Errorf("%v", err)
	}
}

// insertPipelineStageAt will insert a pipeline stage (bson.D) at a given place
// in a []bson.D, copying the tail out so that no stages are lost
func insertPipelineStageAt(pipeline []bson.D, val bson.D, i int) []bson.D {
	return append(pipeline[:i], append([]bson.D{val}, pipeline[i:]...)...)
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
	if containsCardinalityAlteringStages(mongoSource.pipeline) {
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
