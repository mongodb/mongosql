package evaluator

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/10gen/mongo-go-driver/bson"

	"github.com/10gen/mongoast/ast"

	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/astutil"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/internal/procutil"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"

	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	// MaxDepth allowed by BIC's MongoDB driver is 200, keep this up to date
	// if this number should happen to change. We set it less than 200
	// because expressions generated here can be included in stages
	// and higher expressions created in other places that would also
	// count against the total depth. It is better to err on the
	// side of caution than hope that we will always have the
	// exact total correct in all contexts.
	MaxDepth = 180
)

func translateConvert(expr ast.Expr, from, to types.EvalType) ast.Expr {
	var targetType string
	switch to {
	case types.EvalBoolean:
		targetType = "bool"
		// If the from type is a string, convert to int before boolean, because
		// mongo type conversion assumes "false" is the only false
		// string, whereas we actually want '0' to be false and any non-zero
		// integer to be true. As it is now, MongoDB will convert the string '0'
		// to true.
		if from == types.EvalString {
			expr = astutil.WrapInConvert(expr, "int", astutil.ZeroInt32Literal, astutil.NullLiteral)

			// If the from type is a floating point type, we need to round because
			// -0.4 through 0.4 should be treated as false.
		} else if from == types.EvalDouble || from == types.EvalDecimal128 {
			expr = astutil.WrapInRound(expr)
		}
	case types.EvalDecimal128:
		targetType = "decimal"
		if from == types.EvalObjectID {
			expr = translateConvert(expr, from, types.EvalDatetime)
		}
	case types.EvalDouble:
		targetType = "double"
		if from == types.EvalObjectID {
			expr = translateConvert(expr, from, types.EvalDatetime)
		}
	case types.EvalInt32, types.EvalUint32:
		targetType = "int"
		if from == types.EvalDecimal128 || from == types.EvalDouble {
			expr = astutil.WrapInRound(expr)
		} else if from == types.EvalObjectID {
			expr = translateConvert(expr, from, types.EvalDatetime)
		}
	case types.EvalInt64, types.EvalUint64:
		targetType = "long"
		if from == types.EvalDecimal128 || from == types.EvalDouble {
			expr = astutil.WrapInRound(expr)
		} else if from == types.EvalObjectID {
			expr = translateConvert(expr, from, types.EvalDatetime)
		}
	case types.EvalObjectID:
		targetType = "objectId"
	case types.EvalString:
		targetType = "string"
		// Bools need to be converted to String as "1" or "0", rather than
		// as "true" and "false".
		cond := ast.NewBinary(bsonutil.OpEq,
			astutil.WrapInType(expr),
			astutil.StringValue("bool"),
		)

		expr = astutil.WrapInCond(
			astutil.WrapInCond(astutil.StringValue("1"), astutil.StringValue("0"), expr),
			expr,
			cond,
		)
	case types.EvalDatetime, types.EvalDate:
		targetType = "date"
	default:
		panic(fmt.Errorf("target type %s is not a valid target type for $convert",
			string(types.EvalTypeToSQLType(to))))
	}

	if from == types.EvalDate {
		// Need to special-case date-to-string.
		if targetType == "string" {
			converted := astutil.WrapInDateToString(expr, "%Y-%m-%d")
			return converted
		}

		// If the expression is a date, mask its time fields.
		expr = astutil.WrapInDateFromParts(expr, expr, expr)
	}

	defaultVal := astutil.NullLiteral
	switch targetType {
	case "bool":
		defaultVal = astutil.FalseLiteral
	case "decimal":
		defaultVal = astutil.ZeroInt32Literal
	case "double":
		defaultVal = astutil.ZeroInt32Literal
	case "int":
		defaultVal = astutil.ZeroInt32Literal
	case "long":
		defaultVal = astutil.ZeroInt32Literal
	}

	return astutil.WrapInConvert(expr, targetType, defaultVal, astutil.NullLiteral)
}

// translatableToMatch is an interface for any Expr node that can currently
// be translated to MongoDB Match language.
type translatableToMatch interface {
	ToMatchLanguage(*PushdownTranslator) (ast.Expr, SQLExpr)
}

// FieldRefLookup is a function that, given a tableName and a columnName, will return
// the field name coming back from mongodb. It may return a variable reference in the
// case of left joins where columns with the "as" field as their prefix may be renamed
// using "this" (a variable reference) in $filter and $map functions.
type FieldRefLookup func(databaseName, tableName, columnName string) (ast.Ref, bool)

// PushdownTranslator handles the state necessary to do pushdown translation.
type PushdownTranslator struct {
	LookupFieldRef     FieldRefLookup
	Cfg                *PushdownConfig
	columnsToNullCheck map[string]struct{}
	subqueryCmpStages  []ast.Stage
}

// NewPushdownTranslator returns a new PushdownTranslator.
func NewPushdownTranslator(cfg *PushdownConfig, lookupFieldRef FieldRefLookup) *PushdownTranslator {
	return &PushdownTranslator{
		Cfg:                cfg,
		LookupFieldRef:     lookupFieldRef,
		columnsToNullCheck: map[string]struct{}{},
		subqueryCmpStages:  []ast.Stage{},
	}
}

func (t *PushdownTranslator) addSubqueryCmpLookupStage(subPlanMs *MongoSourceStage) error {
	// cannot use expressive lookup before 3.6
	if !t.versionAtLeast(3, 6, 0) {
		return fmt.Errorf("cannot push down subquery comparison stage to " +
			"expressive lookup: expressive lookup not available")
	}

	if subPlanMs.isShardedCollection[subPlanMs.Collection()] {
		return fmt.Errorf("cannot use expressive $lookup on a sharded collection")
	}

	collName := subPlanMs.Collection()
	lookup := ast.NewLookupStage(
		collName, "", "",
		getSubqueryLookupField(collName, subPlanMs.selectIDs),
		[]*ast.LookupLetItem{},
		subPlanMs.pipeline,
	)

	t.subqueryCmpStages = append(t.subqueryCmpStages, lookup)

	return nil
}

func (t *PushdownTranslator) valueKind() values.SQLValueKind {
	return t.Cfg.sqlValueKind
}

// nolint: unparam
func (t *PushdownTranslator) versionAtLeast(major, minor, patch uint8) bool {
	return procutil.VersionAtLeast(t.Cfg.mongoDBVersion, []uint8{major, minor, patch})
}

// ClearColumnsToNullCheck clears this translator's collection of columns
// to null-check.
func (t *PushdownTranslator) ClearColumnsToNullCheck() {
	t.columnsToNullCheck = map[string]struct{}{}
}

// ColumnsToNullCheck returns the columnsToNullCheck map.
func (t *PushdownTranslator) ColumnsToNullCheck() map[string]struct{} {
	return t.columnsToNullCheck
}

// ToAggregationLanguage translates the provided SQLExpr into something that can
// be used in an aggregation pipeline. If the provided SQLExpr cannot be
// translated, the second return value will be an error.
func (t *PushdownTranslator) ToAggregationLanguage(e SQLExpr) (ast.Expr, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationPredicate translates the provided SQLExpr to the aggregation
// language to be evaluated as a predicate directly in a $match stage via $expr.
func (t *PushdownTranslator) ToAggregationPredicate(e SQLExpr) (ast.Expr, PushdownFailure) {
	return e.ToAggregationPredicate(t)
}

// ToMatchLanguage translates the provided SQLExpr into something that can
// be used in an match expression. If the SQLExpr can be fully translated, the
// first return value will be the translated expression, and the second will be
// nil. If the provided SQLExpr cannot be fully translated, the first return
// value will be the partially translated expression, and the second will be the
// original SQLExpr.
func (t *PushdownTranslator) ToMatchLanguage(e SQLExpr) (ast.Expr, SQLExpr) {
	if predicate, ok := e.(translatableToMatch); ok {
		return predicate.ToMatchLanguage(t)
	}
	return nil, e
}

// withNullCheckedColumnsScope wraps the argument in a $let with the
// variable bindings for this translator's columns to null-check (if any exist).
func (t *PushdownTranslator) withNullCheckedColumnsScope(evaluation ast.Expr) ast.Expr {
	assignments := make([]*ast.LetVariable, len(t.columnsToNullCheck))
	i := 0
	for columnName := range t.columnsToNullCheck {
		assignments[i] = ast.NewLetVariable(
			toNullCheckedLetVarName(columnName), astutil.WrapInNullCheck(astutil.FieldRefFromFieldName(columnName)))
		i++
	}

	// Sort the assignments so the order of let variables is deterministic.
	// This is useful for testing.
	sort.Slice(assignments, func(i, j int) bool {
		return assignments[i].Name < assignments[j].Name
	})

	return wrapInLet(assignments, evaluation)
}

// TranslateExpr is a wrapper around ToAggregationLanguage that will fail to
// translate the expr if the resulting aggregation exceeds the maximum allowed
// nesting depth for BSON documents.
func (t *PushdownTranslator) TranslateExpr(e SQLExpr) (ast.Expr, PushdownFailure) {
	doc, err, _ := t.translateExprWithDepth(e)
	if err != nil {
		return doc, err
	}

	return t.withNullCheckedColumnsScope(doc), nil
}

// nolint: unparam
func (t *PushdownTranslator) translateExprWithDepth(e SQLExpr) (ast.Expr, PushdownFailure, uint32) {
	doc, err := t.ToAggregationLanguage(e)
	depth := ComputeDocNestingDepthWithMaxDepth(doc, MaxDepth)
	if depth <= MaxDepth {
		return doc, err, depth
	}

	t.Cfg.lg.Debugf(log.Dev,
		"maximum expression depth: %d exceeded, cannot pushdown, expression was: %v",
		MaxDepth, e)

	err = newPushdownFailure(
		e.ExprName(),
		"maximum pipeline nesting depth exceeded",
		"depth", strconv.Itoa(MaxDepth),
		"expression", fmt.Sprintf("%v", e),
	)

	return nil, err, 0
}

// TranslateAggPredicate is a wrapper around ToAggregationPredicate that will
// fail to translate the expr if the resulting aggregation exceeds the maximum
// allowed nesting depth for BSON documents.
func (t *PushdownTranslator) TranslateAggPredicate(e SQLExpr) (ast.Expr, PushdownFailure) {
	doc, err, _ := t.translateAggPredicateWithDepth(e)
	if err != nil {
		return doc, err
	}

	return t.withNullCheckedColumnsScope(doc), nil
}

func (t *PushdownTranslator) translateAggPredicateWithDepth(e SQLExpr) (ast.Expr, PushdownFailure, uint32) {
	doc, err := t.ToAggregationPredicate(e)
	depth := ComputeDocNestingDepthWithMaxDepth(doc, MaxDepth)
	if depth <= MaxDepth {
		return doc, err, depth
	}

	t.Cfg.lg.Debugf(log.Dev,
		"maximum expression depth: %d exceeded, cannot pushdown, expression was: %v",
		MaxDepth, e)

	err = newPushdownFailure(
		e.ExprName(),
		"maximum pipeline nesting depth exceeded",
		"depth", strconv.Itoa(MaxDepth),
		"expression", fmt.Sprintf("%v", e),
	)

	return nil, err, 0
}

// TranslatePredicate is a wrapper around ToMatchLanguage that will fail to
// translate the expr if the resulting aggregation exceeds the maximum allowed
// nesting depth for BSON documents.
func (t *PushdownTranslator) TranslatePredicate(e SQLExpr) (ast.Expr, SQLExpr) {
	var doc ast.Expr
	var expr SQLExpr

	doc, expr, _ = t.translatePredicateWithDepth(e)

	if expr != nil && t.versionAtLeast(3, 6, 0) {
		agg, err := t.TranslateAggPredicate(e)
		if err == nil {
			return ast.NewAggExpr(agg), nil
		}
	}

	return doc, expr
}

// nolint: unparam
func (t *PushdownTranslator) translatePredicateWithDepth(e SQLExpr) (ast.Expr, SQLExpr, uint32) {
	translatable, ok := e.(translatableToMatch)
	if !ok {
		return nil, e, 0
	}
	doc, expr := translatable.ToMatchLanguage(t)
	if doc == nil {
		return nil, e, 0
	}
	depth := ComputeDocNestingDepthWithMaxDepth(doc, MaxDepth)
	if depth <= MaxDepth {
		return doc, expr, depth
	}
	t.Cfg.lg.Debugf(log.Dev,
		"maximum predicate depth: %d exceeded, cannot pushdown, predicate was: %v",
		MaxDepth,
		e)
	return nil, e, 0
}

func (t *PushdownTranslator) getFieldOrVariableRef(e SQLExpr) (ast.Ref, bool) {
	switch field := e.(type) {
	case SQLColumnExpr:
		return t.LookupFieldRef(field.databaseName, field.tableName, field.columnName)
	default:
		return nil, false
	}
}

func (t *PushdownTranslator) getValue(e SQLExpr) (*ast.Constant, PushdownFailure) {
	cons, ok := e.(SQLValueExpr)
	if !ok {
		return nil, newPushdownFailure(
			e.ExprName(),
			"SQLExpr is not a SQLValueExpr",
			"expr", fmt.Sprintf("%#v", e),
		)
	}

	if cons.Value.IsNull() {
		return astutil.NullConstant(), nil
	}

	if cons.EvalType() == types.EvalDecimal128 {
		return t.translateDecimal(cons.Value, cons.ExprName())
	}

	v, err := cons.Value.BSONValue()
	if err != nil {
		return nil, newPushdownFailure(
			cons.ExprName(),
			"failed to get value from SQLValue",
			"error", fmt.Sprintf("%v", err),
		)
	}

	return ast.NewConstant(v), nil
}

func (t *PushdownTranslator) translateDateFormatAsDate(f *dateFormatFunc) (ast.Expr, PushdownFailure) {
	formatValue, ok := f.args[1].(SQLValueExpr)
	if !ok {
		return nil, newPushdownFailure(f.ExprName(), "format string argument was not literal")
	}

	date, err := t.TranslateExpr(f.args[0])
	if err != nil {
		return nil, err
	}

	hasYear := false
	hasMonth := false
	hasDay := false
	hasHour := false
	hasMinute := false
	hasSecond := false

	// NOTE: this is a very specific optimization for Tableau's discrete dimension
	// functionality, which only generates the below formats. MongoDB 3.6 will support
	// converting a string back into a date and the below optimizations won't be needed.
	switch formatValue.String() {
	case "%Y-01-01", "%Y-01-01 00:00:00":
		hasYear = true
	case "%Y-%m-01", "%Y-%m-01 00:00:00":
		hasYear = true
		hasMonth = true
	case "%Y-%m-%d", "%Y-%m-%d 00:00:00":
		hasYear = true
		hasMonth = true
		hasDay = true
	case "%Y-%m-%d %H:00:00":
		hasYear = true
		hasMonth = true
		hasDay = true
		hasHour = true
	case "%Y-%m-%d %H:%i:00":
		hasYear = true
		hasMonth = true
		hasDay = true
		hasHour = true
		hasMinute = true
	case "%Y-%m-%d %H:%i:%s":
		hasYear = true
		hasMonth = true
		hasDay = true
		hasHour = true
		hasMinute = true
		hasSecond = true
	}

	if !hasYear {
		return nil, newPushdownFailure(f.ExprName(), "no year in date format string")
	}

	parts := make([]ast.Expr, 0, 5)
	if !hasMonth {
		parts = append(parts, ast.NewBinary(bsonutil.OpMultiply,
			astutil.Int32Value(int32(24*time.Hour/time.Millisecond)),
			ast.NewBinary(bsonutil.OpSubtract,
				ast.NewFunction(bsonutil.OpDayOfYear, date),
				astutil.OneInt32Literal,
			),
		))

	} else if !hasDay {
		parts = append(parts, ast.NewBinary(bsonutil.OpMultiply,
			astutil.Int32Value(int32(24*time.Hour/time.Millisecond)),
			ast.NewBinary(bsonutil.OpSubtract,
				ast.NewFunction(bsonutil.OpDayOfMonth, date),
				astutil.OneInt32Literal,
			),
		))
	}

	if !hasHour {
		parts = append(parts, ast.NewBinary(bsonutil.OpMultiply,
			astutil.Int32Value(int32(time.Hour/time.Millisecond)),
			ast.NewFunction(bsonutil.OpHour, date),
		))
	}
	if !hasMinute {
		parts = append(parts, ast.NewBinary(bsonutil.OpMultiply,
			astutil.Int32Value(int32(time.Minute/time.Millisecond)),
			ast.NewFunction(bsonutil.OpMinute, date),
		))
	}
	if !hasSecond {
		parts = append(parts, ast.NewBinary(bsonutil.OpMultiply,
			astutil.Int32Value(int32(time.Second/time.Millisecond)),
			ast.NewFunction(bsonutil.OpSecond, date),
		))
	}

	parts = append(parts, ast.NewFunction(bsonutil.OpMillisecond, date))

	var totalMS ast.Expr
	if len(parts) == 1 {
		totalMS = parts[0]
	} else {
		totalMS = astutil.WrapInOp(bsonutil.OpAdd, parts...)
	}

	sub := ast.NewBinary(bsonutil.OpSubtract, date, totalMS)

	return astutil.WrapInNullCheckedCond(astutil.NullLiteral, sub, date), nil
}

func (t *PushdownTranslator) translateDecimal(cons values.SQLValue, exprName string) (*ast.Constant, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			exprName,
			"cannot translate SQLValue to decimal on MongoDB < 3.4",
		)
	}

	d, err := cons.BSONValue()
	if err != nil {
		return nil, newPushdownFailure(exprName, "cannot translate SQLValue to decimal", fmt.Sprintf("%v", err))
	}

	return ast.NewConstant(d), nil
}

func (t *PushdownTranslator) translateOperator(op string, nameExpr, valExpr SQLExpr) (*ast.Binary, bool) {
	ref, ok := t.getFieldOrVariableRef(nameExpr)
	if !ok {
		return nil, false
	}

	fieldValue, err := t.getValue(valExpr)
	if err != nil {
		return nil, false
	}

	colExpr, ok := nameExpr.(SQLColumnExpr)
	mType := colExpr.columnType.MongoType
	if ok {
		if values.IsUUID(mType) {
			bin, ok := GetBinaryFromExpr(mType, valExpr)
			if !ok {
				return nil, false
			}
			fieldValue = astutil.BinaryConstant(bin)
		} else if mType == schema.MongoObjectID {
			// We know this type assert is safe because of the call to
			// t.getValue(valExpr) above.
			var hex string
			switch typed := valExpr.(SQLValueExpr).Value.(type) {
			case values.SQLVarchar:
				hex = string(bson.ObjectIdHex(values.String(typed)))
			case values.SQLObjectID:
				hex = typed.Value().(bson.ObjectId).Hex()
			default:
				return nil, false
			}
			oid, err := primitive.ObjectIDFromHex(hex)
			if err != nil {
				panic(fmt.Sprintf("failed to convert ObjectID SQLValue into ast.Constant: %v", err))
			}
			fieldValue = astutil.ObjectIDConstant(oid)
		}
	}

	return ast.NewBinary(ast.BinaryOp(op), ref, fieldValue), true
}

// getSubqueryLookupField returns a reference for the array that is embedded
// into the pipeline documents from a $lookup that is used for pushing down a
// subquery.
func getSubqueryLookupField(table string, selectIDs []int) string {
	return fmt.Sprintf("__subquery_%v_%v", table, selectIDs)
}

// lookupArrayRef returns a reference to an array of field values that
// correspond to the given Column and MongoSourceStage that is returned from a
// $lookup.
// NOTE: This utilizes the behavior of addressing a fieldname on an array of
// documents, wherein the referred field of the constituent member documents are
// extracted out into an array.
func lookupArrayRef(subPlanMs *MongoSourceStage, col *results.Column) (*ast.Function, error) {
	matchFieldName, found := subPlanMs.mappingRegistry.lookupFieldName(
		col.Database,
		col.Table,
		col.Name,
	)
	if !found {
		return nil, fmt.Errorf("could not find field name using: %v, %v, %v",
			col.Database, col.OriginalTable, col.MappingRegistryName)
	}

	collName := subPlanMs.Collection()
	lookupArray := ast.NewFieldRef(getSubqueryLookupField(collName, subPlanMs.selectIDs), nil)

	ref := astutil.FieldRefFromFieldName(matchFieldName)
	ref.Parent = ast.NewVariableRef("this")
	in := astutil.WrapInIfNull(ref, astutil.NullLiteral)

	return astutil.WrapInMap(lookupArray, "this", in), nil
}

// constructNullChecks creates a slice of equality checks for each given arg to
// check if it is NULL.
func constructNullChecks(args []ast.Expr) []ast.Expr {
	nullChecks := make([]ast.Expr, len(args))
	for i, arg := range args {
		nullChecks[i] = ast.NewBinary(bsonutil.OpEq, arg, astutil.NullLiteral)
	}

	return nullChecks
}

// threeValueLogicCheck returns an aggregation language expression that ensures
// that we handle comparisons correctly in the presence of MySQL NULLs. This
// involves essentially ensuring that NULLs always evaluate to a falsy result.
// This is necessary due to the three-valued boolean logic of SQL.
func threeValueLogicCheck(predicate ast.Expr, firstOperand ast.Expr) ast.Expr {
	// TODO BI-1885: This code will need to be changed to consider the case where
	// the operand is a subquery (or a tuple, which is desugared into a subquery
	// anyways) and break apart the columns so that each column value will be
	// checked against null. This currently is not happening because the
	// implementation for this can't be tested until we support correlated
	// subqueries.
	args := []ast.Expr{firstOperand}

	nullChecks := constructNullChecks(args)

	notNullCheck := ast.NewFunction(bsonutil.OpNot, astutil.WrapInOp(bsonutil.OpOr, nullChecks...))

	return ast.NewBinary(ast.And, predicate, notNullCheck)
}

// opFromSQLOpForSubqueryCmp takes a given parser op string and returns its equivalent
// bsonutil/aggregation operator string.
// e.g.
// parser.AST_EQ -> bsonutil.OpEq
func opFromSQLOpForSubqueryCmp(op string) (ast.BinaryOp, error) {
	switch op {
	case parser.AST_EQ:
		return ast.Equals, nil
	case parser.AST_NE:
		return ast.NotEquals, nil
	case parser.AST_LT:
		return ast.LessThan, nil
	case parser.AST_LE:
		return ast.LessThanOrEquals, nil
	case parser.AST_GT:
		return ast.GreaterThan, nil
	case parser.AST_GE:
		return ast.GreaterThanOrEquals, nil
	}

	return "", fmt.Errorf("unrecognized comparison operator %v", op)
}

// constructMapArray creates the array argument that can be later passed into a
// $map.
func constructMapArray(subPlanMs *MongoSourceStage) (*ast.Function, error) {
	var argArray *ast.Function
	if len(subPlanMs.Columns()) > 1 {
		// If the number of columns coming from the subquery is more than 1, we are
		// dealing with tuple-based equality checking. In order to have this
		// actually be correct, a simple $in is not enough, we have to create
		// tuples in the agg pipeline ourselves and match against those.
		inputs := make([]ast.Expr, len(subPlanMs.Columns()))
		lastDBName := subPlanMs.Columns()[0].Database
		for i := range subPlanMs.Columns() {
			col := subPlanMs.Columns()[i]
			curDBName := col.Database
			if lastDBName != curDBName {
				return nil, fmt.Errorf("cannot use expressive $lookup across databases")
			}
			ref, err := lookupArrayRef(subPlanMs, col)
			if err != nil {
				return nil, err
			}
			inputs[i] = ref
		}

		// No reason to check the mongod version here, since if they are too old to
		// support $zip, they definitely can't support expressive $lookup!
		argArray = ast.NewFunction(bsonutil.OpZip, ast.NewDocument(
			ast.NewDocumentElement("inputs", ast.NewArray(inputs...)),
		))
	} else {
		var err error
		col := subPlanMs.Columns()[0]
		argArray, err = lookupArrayRef(subPlanMs, col)
		if err != nil {
			return nil, err
		}
	}

	return argArray, nil
}

// constructMapCmp creates a $map with which to map the given array of
// arguments into an array of booleans based on the given left argument and
// comparison operator.
func constructMapCmp(left ast.Expr, array ast.Expr, cmpOp string) (*ast.Function, error) {
	mgoOp, err := opFromSQLOpForSubqueryCmp(cmpOp)
	if err != nil {
		return nil, err
	}

	operatorCmp := ast.NewBinary(mgoOp, left, astutil.ThisVarRef)

	threeValuedCmp := threeValueLogicCheck(operatorCmp, left)

	// This is nice, but we actually need to add a null check for $$this!
	// Unfortunately, "$$this" is not a SQLExpr so we can't just wrap it in three
	// valued logic like we do above. This is NOT due to three valued boolean
	// logic, but because of `mongod`'s type comparison taking precedence over
	// value!
	// e.g. 4 > null = TRUE
	thisNotNull := ast.NewFunction(
		bsonutil.OpNot,
		ast.NewBinary(ast.Equals, astutil.ThisVarRef, astutil.NullLiteral),
	)

	threeValuedCmp = ast.NewBinary(ast.And, threeValuedCmp, thisNotNull)

	return astutil.WrapInMap(array, "this", threeValuedCmp), nil
}

// boolConvertMapForCmpOp takes an operator, and creates a mapping that allows
// for the reduction or folding of an array of comparison operator results into
// a single boolean value, through the (eventual) use of things like
// $allElementsTrue. This function is intended as a means of avoiding
// duplication in SQLSubqueryCmpExpr toAggregationLanguage() implementations,
// and performs the work of updating the pushdown translator's
// subqueryCmpStages.
func (t *PushdownTranslator) boolConvertMapForCmpOp(left ast.Expr,
	subPlanMs *MongoSourceStage, op string) (*ast.Function, error) {
	err := t.addSubqueryCmpLookupStage(subPlanMs)
	if err != nil {
		return nil, err
	}

	mapArray, err := constructMapArray(subPlanMs)
	if err != nil {
		return nil, err
	}

	mapCmp, err := constructMapCmp(left, mapArray, op)
	if err != nil {
		return nil, err
	}

	return mapCmp, nil
}

// mapCmpForDoubleSubquery takes a double-subquery comparison expression based
// on the given op, and produces a mapping from the row-wise comparisons
// between the two subqueries to an array of booleans representing the
// comparison results per row.
func (t *PushdownTranslator) mapCmpForDoubleSubquery(e SQLExpr, leftPlan PlanStage, rightPlan PlanStage, op string) (*ast.Function, PushdownFailure) {
	leftPlanMs, ok := leftPlan.(*MongoSourceStage)
	if !ok {
		return nil, innerSubqueryPushdownFailure(e)
	}

	if leftPlanMs.LimitRowCount != 1 && !leftPlanMs.IsDual() {
		return nil, multiRowSubqueryPushdownFailure(e)
	}

	lookupStageAddErr := t.addSubqueryCmpLookupStage(leftPlanMs)
	if lookupStageAddErr != nil {
		return nil, wrapExprErrWithPushdownFailure(e, lookupStageAddErr)
	}

	rightPlanMs, ok := rightPlan.(*MongoSourceStage)
	if !ok {
		return nil, innerSubqueryPushdownFailure(e)
	}

	// MySQL hardcode-desugars IN -> = ANY and NOT IN -> <> ALL, so we need to
	// mimic this here.
	numLeftCols := len(leftPlanMs.Columns())
	_, isSQLAny := e.(*SQLSubqueryAnyExpr)
	_, isSQLAll := e.(*SQLSubqueryAllExpr)
	if numLeftCols > 1 && !((isSQLAny && op == "=") || (isSQLAll && op == "!=")) {
		return nil, wrapExprErrWithPushdownFailure(e, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, 1))
	}

	leftCol := leftPlanMs.Columns()[0]
	var leftRefValue ast.Expr

	if numLeftCols > 1 {
		leftRefs := make([]ast.Expr, numLeftCols)
		for i, col := range leftPlanMs.Columns() {
			leftRef, err := lookupArrayRef(leftPlanMs, col)
			if err != nil {
				return nil, wrapExprErrWithPushdownFailure(e, err)
			}
			leftRefs[i] = astutil.WrapInOp(bsonutil.OpArrElemAt, leftRef, astutil.ZeroInt32Literal)
		}
		leftRefValue = ast.NewArray(leftRefs...)
	} else {
		leftRef, err := lookupArrayRef(leftPlanMs, leftCol)
		if err != nil {
			return nil, wrapExprErrWithPushdownFailure(e, err)
		}
		leftRefValue = astutil.WrapInOp(bsonutil.OpArrElemAt, leftRef, astutil.ZeroInt32Literal)
	}

	mapCmp, err := t.boolConvertMapForCmpOp(leftRefValue, rightPlanMs, op)
	if err != nil {
		return nil, wrapExprErrWithPushdownFailure(e, err)
	}

	return mapCmp, nil
}

func negate(op ast.Expr) ast.Expr {
	switch t := op.(type) {
	case *ast.Binary:
		return negateBinary(t)
	case *ast.Function:
		return negateFunction(t)
	}

	return astutil.WrapInOp(bsonutil.OpNor, op)
}

func negateBinary(op *ast.Binary) ast.Expr {
	switch op.Op {
	case bsonutil.OpAnd:
		return astutil.WrapInOp(bsonutil.OpNor, op)
	case bsonutil.OpNor:
		return ast.NewBinary(bsonutil.OpOr, op.Left, op.Right)
	case bsonutil.OpOr:
		return ast.NewBinary(bsonutil.OpNor, op.Left, op.Right)
	case bsonutil.OpEq:
		negation := ast.NewBinary(bsonutil.OpNeq, op.Left, op.Right)

		// Negating equals may require adding a null-check:
		// not (a = 1), for example, should become
		// { $and: [{ a: { $ne: 1} }, { a: { $ne: null } }] }
		if v, ok := op.Right.(*ast.Constant); ok && v.Value.Type != bsontype.Null {
			negation = ast.NewBinary(bsonutil.OpAnd,
				negation,
				ast.NewBinary(bsonutil.OpNeq, op.Left, astutil.NullLiteral),
			)
		}
		return negation
	case bsonutil.OpGt:
		return ast.NewBinary(bsonutil.OpLte, op.Left, op.Right)
	case bsonutil.OpGte:
		return ast.NewBinary(bsonutil.OpLt, op.Left, op.Right)
	case bsonutil.OpLt:
		return ast.NewBinary(bsonutil.OpGte, op.Left, op.Right)
	case bsonutil.OpLte:
		return ast.NewBinary(bsonutil.OpGt, op.Left, op.Right)
	case bsonutil.OpNeq:
		return ast.NewBinary(bsonutil.OpEq, op.Left, op.Right)
	}

	return op
}

func negateFunction(op *ast.Function) ast.Expr {
	switch op.Name {
	case bsonutil.OpOr:
		return ast.NewFunction(bsonutil.OpNor, op.Arg)
	case bsonutil.OpNor:
		return ast.NewFunction(bsonutil.OpOr, op.Arg)
	case bsonutil.OpAnd:
		return astutil.WrapInOp(bsonutil.OpNor, op)
	}

	return op
}

// GetBinaryFromExpr attempts to convert e to a bson.Binary -
// that represents a MongoDB UUID - using mType.
func GetBinaryFromExpr(mType schema.MongoType, e SQLExpr) (bson.Binary, bool) {
	// we accept UUIDs as string arguments
	uuidString := strings.Replace(e.String(), "-", "", -1)
	bytes, err := hex.DecodeString(uuidString)
	if err != nil {
		return bson.Binary{}, false
	}

	err = values.NormalizeUUID(mType, bytes)
	if err != nil {
		return bson.Binary{}, false
	}

	if mType == schema.MongoUUID {
		return bson.Binary{Kind: 0x04, Data: bytes}, true
	}

	return bson.Binary{Kind: 0x03, Data: bytes}, true
}

// getProjectedFieldName returns an ast.Expr to project the given field.
func getProjectedFieldName(fieldName string, fieldType types.EvalType) ast.Expr {
	names := strings.Split(fieldName, ".")

	if len(names) == 1 {
		return ast.NewFieldRef(fieldName, nil)
	}

	value, err := strconv.Atoi(names[len(names)-1])
	// special handling for legacy 2d array
	if err == nil && fieldType == types.EvalArrNumeric {
		fieldName = fieldName[0:strings.LastIndex(fieldName, ".")]
		return astutil.WrapInOp(bsonutil.OpArrElemAt,
			astutil.FieldRefFromFieldName(fieldName),
			astutil.Int64Value(int64(value)),
		)
	}

	return astutil.FieldRefFromFieldName(fieldName)
}

// containsBSONType returns an expression that evaluates to true if types
// contains the BSON type of v.
func containsBSONType(v ast.Expr, types ...string) *ast.Function {
	vType := astutil.WrapInType(v)
	checks := make([]ast.Expr, len(types))

	for i, t := range types {
		checks[i] = ast.NewBinary(bsonutil.OpEq, vType, astutil.StringValue(t))
	}

	return astutil.WrapInOp(bsonutil.OpOr, checks...)
}
