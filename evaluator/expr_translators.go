package evaluator

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/procutil"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
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

func translateConvert(expr interface{}, from, to types.EvalType) interface{} {
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
			expr = bsonutil.WrapInConvert(expr, "int", 0, nil)

			// If the from type is a floating point type, we need to round because
			// -0.4 through 0.4 should be treated as false.
		} else if from == types.EvalDouble || from == types.EvalDecimal128 {
			expr = bsonutil.WrapInRound(expr)
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
			expr = bsonutil.WrapInRound(expr)
		} else if from == types.EvalObjectID {
			expr = translateConvert(expr, from, types.EvalDatetime)
		}
	case types.EvalInt64, types.EvalUint64:
		targetType = "long"
		if from == types.EvalDecimal128 || from == types.EvalDouble {
			expr = bsonutil.WrapInRound(expr)
		} else if from == types.EvalObjectID {
			expr = translateConvert(expr, from, types.EvalDatetime)
		}
	case types.EvalObjectID:
		targetType = "objectId"
	case types.EvalString:
		targetType = "string"
		// Bools need to be converted to String as "1" or "0", rather than
		// as "true" and "false".
		cond := bsonutil.WrapInOp(bsonutil.OpEq,
			bsonutil.WrapInType(expr),
			"bool",
		)

		expr = bsonutil.WrapInCond(bsonutil.WrapInCond("1", "0", expr), expr, cond)
	case types.EvalDatetime, types.EvalDate:
		targetType = "date"
	default:
		panic(fmt.Errorf("target type %s is not a valid target type for $convert",
			string(types.EvalTypeToSQLType(to))))
	}

	if from == types.EvalDate {
		// Need to special-case date-to-string.
		if targetType == "string" {
			converted := bsonutil.WrapInDateToString(expr, "%Y-%m-%d")
			return converted
		}

		// If the expression is a date, mask its time fields.
		expr = bsonutil.WrapInDateFromParts(expr, expr, expr)
	}

	var defaultVal interface{}
	switch targetType {
	case "bool":
		defaultVal = false
	case "decimal":
		defaultVal = 0
	case "double":
		defaultVal = 0
	case "int":
		defaultVal = 0
	case "long":
		defaultVal = 0
	}

	return bsonutil.WrapInConvert(expr, targetType, defaultVal, nil)
}

// translatableToMatch is an interface for any Expr node that can currently
// be translated to MongoDB Match language.
type translatableToMatch interface {
	ToMatchLanguage(*PushdownTranslator) (bson.M, SQLExpr)
}

// FieldNameLookup is a function that, given a tableName and a columnName, will return
// the field name coming back from mongodb.
type FieldNameLookup func(databaseName, tableName, columnName string) (string, bool)

// PushdownTranslator handles the state necessary to do pushdown translation.
type PushdownTranslator struct {
	LookupFieldName    FieldNameLookup
	Cfg                *PushdownConfig
	piecewiseDeps      []*NonCorrelatedSubqueryFuture
	correlatedColumns  []*CorrelatedSubqueryColumnFuture
	columnsToNullCheck map[string]struct{}
}

// NewPushdownTranslator returns a new PushdownTranslator.
func NewPushdownTranslator(cfg *PushdownConfig, lookupFieldName FieldNameLookup) *PushdownTranslator {
	return &PushdownTranslator{Cfg: cfg, LookupFieldName: lookupFieldName, columnsToNullCheck: map[string]struct{}{}}
}

func (t *PushdownTranslator) addNonCorrelatedSubqueryFuture(p PlanStage) *NonCorrelatedSubqueryFuture {
	piece := NewNonCorrelatedSubqueryFuture(p)
	t.piecewiseDeps = append(t.piecewiseDeps, piece)
	return piece
}

func (t *PushdownTranslator) addCorrelatedSubqueryColumnFuture(c *SQLColumnExpr) *CorrelatedSubqueryColumnFuture {
	cc := NewCorrelatedSubqueryColumnFuture(c)
	t.correlatedColumns = append(t.correlatedColumns, cc)
	return cc
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
func (t *PushdownTranslator) ToAggregationLanguage(e SQLExpr) (interface{}, PushdownFailure) {
	return e.ToAggregationLanguage(t)
}

// ToAggregationPredicate translates the provided SQLExpr to the aggregation
// language to be evaluated as a predicate directly in a $match stage via $expr.
func (t *PushdownTranslator) ToAggregationPredicate(e SQLExpr) (interface{}, PushdownFailure) {
	return e.ToAggregationPredicate(t)
}

// ToMatchLanguage translates the provided SQLExpr into something that can
// be used in an match expression. If the SQLExpr can be fully translated, the
// first return value will be the translated expression, and the second will be
// nil. If the provided SQLExpr cannot be fully translated, the first return
// value will be the partially translated expression, and the second will be the
// original SQLExpr.
func (t *PushdownTranslator) ToMatchLanguage(e SQLExpr) (bson.M, SQLExpr) {
	if predicate, ok := e.(translatableToMatch); ok {
		return predicate.ToMatchLanguage(t)
	}
	return nil, e
}

// withNullCheckedColumnsScope wraps the argument in a $let with the
// variable bindings for this translator's columns to null-check (if any exist).
func (t *PushdownTranslator) withNullCheckedColumnsScope(evaluation interface{}) interface{} {
	assignments := make([]bson.DocElem, len(t.columnsToNullCheck))
	i := 0
	for columnName := range t.columnsToNullCheck {
		assignments[i] = bsonutil.NewDocElem(
			toNullCheckedLetVarName(columnName), bsonutil.WrapInNullCheck(columnName))
		i++
	}

	return wrapInLet(assignments, evaluation)
}

// TranslateExpr is a wrapper around ToAggregationLanguage that will fail to
// translate the expr if the resulting aggregation exceeds the maximum allowed
// nesting depth for BSON documents.
func (t *PushdownTranslator) TranslateExpr(e SQLExpr) (interface{}, PushdownFailure) {
	doc, err, _ := t.translateExprWithDepth(e)
	if err != nil {
		return doc, err
	}

	return t.withNullCheckedColumnsScope(doc), nil
}

// nolint: unparam
func (t *PushdownTranslator) translateExprWithDepth(e SQLExpr) (interface{}, PushdownFailure, uint32) {
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
func (t *PushdownTranslator) TranslateAggPredicate(e SQLExpr) (interface{}, PushdownFailure) {
	doc, err, _ := t.translateAggPredicateWithDepth(e)
	if err != nil {
		return doc, err
	}

	return t.withNullCheckedColumnsScope(doc), nil
}

func (t *PushdownTranslator) translateAggPredicateWithDepth(e SQLExpr) (interface{}, PushdownFailure, uint32) {
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
func (t *PushdownTranslator) TranslatePredicate(e SQLExpr) (bson.M, SQLExpr) {
	var doc bson.M
	var expr SQLExpr

	doc, expr, _ = t.translatePredicateWithDepth(e)

	if expr != nil && t.versionAtLeast(3, 6, 0) {
		agg, err := t.TranslateAggPredicate(e)
		if err == nil {
			return bsonutil.NewM(bsonutil.NewDocElem("$expr", agg)), nil
		}
	}

	return doc, expr
}

// nolint: unparam
func (t *PushdownTranslator) translatePredicateWithDepth(e SQLExpr) (bson.M, SQLExpr, uint32) {
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

func (t *PushdownTranslator) getFieldName(e SQLExpr) (string, bool) {
	switch field := e.(type) {
	case SQLColumnExpr:
		return t.LookupFieldName(field.databaseName, field.tableName, field.columnName)
	default:
		return "", false
	}
}

func (t *PushdownTranslator) getValue(e SQLExpr) (interface{}, PushdownFailure) {

	cons, ok := e.(SQLValueExpr)
	if !ok {
		return nil, newPushdownFailure(
			e.ExprName(),
			"SQLExpr is not a SQLValueExpr",
			"expr", fmt.Sprintf("%#v", e),
		)
	}

	if cons.EvalType() == types.EvalDecimal128 {
		return t.translateDecimal(cons.Value, cons.ExprName())
	}

	return cons.Value.Value(), nil
}

func (t *PushdownTranslator) translateDateFormatAsDate(f *dateFormatFunc) (interface{}, PushdownFailure) {
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

	var parts []interface{}
	if !hasMonth {
		parts = append(parts, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpMultiply, bsonutil.NewArray(
			bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpSubtract, bsonutil.NewArray(
				bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpDayOfYear, date)),
				1,
			))),
			uint64(24*time.Hour/time.Millisecond),
		)),
		))

	} else if !hasDay {
		parts = append(parts, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpMultiply, bsonutil.NewArray(
			bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpSubtract, bsonutil.NewArray(
				bsonutil.NewM(bsonutil.NewDocElem("$dayOfMonth", date)),
				1,
			))),
			uint64(24*time.Hour/time.Millisecond),
		)),
		))
	}

	if !hasHour {
		parts = append(parts, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpMultiply, bsonutil.NewArray(
			bsonutil.NewM(bsonutil.NewDocElem("$hour", date)),
			uint64(time.Hour/time.Millisecond),
		)),
		))
	}
	if !hasMinute {
		parts = append(parts, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpMultiply, bsonutil.NewArray(
			bsonutil.NewM(bsonutil.NewDocElem("$minute", date)),
			uint64(time.Minute/time.Millisecond),
		)),
		))
	}
	if !hasSecond {
		parts = append(parts, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpMultiply, bsonutil.NewArray(
			bsonutil.NewM(bsonutil.NewDocElem("$second", date)),
			uint64(time.Second/time.Millisecond),
		)),
		))
	}

	parts = append(parts, bsonutil.NewM(bsonutil.NewDocElem("$millisecond", date)))

	var totalMS interface{}
	if len(parts) == 1 {
		totalMS = parts[0]
	} else {
		totalMS = bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpAdd, parts))
	}

	sub := bsonutil.NewM(
		bsonutil.NewDocElem(bsonutil.OpSubtract, bsonutil.NewArray(
			date,
			totalMS,
		)),
	)

	return bsonutil.WrapInNullCheckedCond(
		nil,
		sub,
		date,
	), nil
}

func (t *PushdownTranslator) translateDecimal(cons values.SQLValue, exprName string) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			exprName,
			"cannot translate SQLValue to decimal on MongoDB < 3.4",
		)
	}

	parsed, err := bson.ParseDecimal128(cons.String())
	if err != nil {
		return nil, newPushdownFailure(
			exprName,
			"failed to parse decimal from SQLValue string",
			"string", cons.String(),
			"error", err.Error(),
		)
	}

	return parsed, nil
}

func (t *PushdownTranslator) translateOperator(op string, nameExpr, valExpr SQLExpr) (bson.M, bool) {
	name, ok := t.getFieldName(nameExpr)
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
			fieldValue, ok = GetBinaryFromExpr(mType, valExpr)
			if !ok {
				return nil, false
			}
		} else if mType == schema.MongoObjectID {
			// We know this type assert is safe because of the call to
			// t.getValue(valExpr) above.
			switch typed := valExpr.(SQLValueExpr).Value.(type) {
			case values.SQLVarchar:
				fieldValue = bson.ObjectIdHex(values.String(typed))
			case values.SQLObjectID:
				fieldValue = typed.Value()
			default:
				return nil, false
			}
		}
	}

	translation := bsonutil.NewM(bsonutil.NewDocElem(name, bsonutil.NewM(bsonutil.NewDocElem(op, fieldValue))))

	if op == bsonutil.OpEq {
		translation = bsonutil.NewM(bsonutil.NewDocElem(name, fieldValue))
	}

	return translation, true
}

func negate(op bson.M) bson.M {
	if len(op) == 1 {
		name, value := getSingleMapEntry(op)
		if strings.HasPrefix(name, "$") {
			switch name {
			case bsonutil.OpOr:
				return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpNor, value))
			case bsonutil.OpNor:
				return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpOr, value))
			}
		} else if innerOp, ok := value.(bson.M); ok {
			if len(innerOp) == 1 {
				innerName, innerValue := getSingleMapEntry(innerOp)
				if strings.HasPrefix(innerName, "$") {
					switch innerName {
					case bsonutil.OpEq:
						return bsonutil.NewM(bsonutil.NewDocElem(name, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpNeq, innerValue))))
					case bsonutil.OpIn:
						return bsonutil.NewM(bsonutil.NewDocElem(name, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpNotIn, innerValue))))
					case bsonutil.OpNeq:
						return bsonutil.NewM(bsonutil.NewDocElem(name, innerValue))
					case bsonutil.OpNotIn:
						return bsonutil.NewM(bsonutil.NewDocElem(name, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpIn, innerValue))))
					case bsonutil.OpRegex:
						return bsonutil.NewM(bsonutil.NewDocElem(name, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpNotIn, bsonutil.NewArray(
							innerValue,
						)))))
					case bsonutil.OpNot:
						return bsonutil.NewM(bsonutil.NewDocElem(name, innerValue))
					}

					return bsonutil.NewM(bsonutil.NewDocElem(name, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpNot, bsonutil.NewM(bsonutil.NewDocElem(innerName, innerValue))))))
				}
			}
		} else {
			// Logical NOT evaluates to 1 if the operand is 0, to 0
			// if the operand is nonzero, and NOT NULL returns NULL.
			// See https://dev.mysql.com/doc/refman/5.7/en/logical-operators.html#operator_not
			// for more.
			translation := bsonutil.NewM(bsonutil.NewDocElem(name, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpNeq, value))))
			if value != nil {
				translation = bsonutil.NewM(
					bsonutil.NewDocElem(bsonutil.OpAnd, bsonutil.NewArray(
						translation,
						bsonutil.NewM(bsonutil.NewDocElem(name, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpNeq, nil)))),
					)),
				)

			}
			return translation
		}
	}

	// $not only works as a meta operator on a single operator
	// so simulate $not using $nor
	return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpNor, bsonutil.NewArray(
		op,
	)))
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

func getSingleMapEntry(m bson.M) (string, interface{}) {
	if len(m) > 1 {
		panic("map has too many entries.")
	}

	for k, v := range m {
		return k, v
	}

	panic("map has no entries!")
}

// getProjectedFieldName returns an interface to project the given field.
func getProjectedFieldName(fieldName string, fieldType types.EvalType) interface{} {

	names := strings.Split(fieldName, ".")

	if len(names) == 1 {
		return "$" + fieldName
	}

	value, err := strconv.Atoi(names[len(names)-1])
	// special handling for legacy 2d array
	if err == nil && fieldType == types.EvalArrNumeric {
		fieldName = fieldName[0:strings.LastIndex(fieldName, ".")]
		return bsonutil.NewM(bsonutil.NewDocElem("$arrayElemAt", bsonutil.NewArray(
			"$"+fieldName,
			value,
		)))
	}

	return "$" + fieldName
}

// containsBSONType returns an expression that evaluates to true if types
// contains the BSON type of v.
func containsBSONType(v interface{}, types ...string) bson.M {

	vType := bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpType, v))
	checks := make([]interface{}, len(types))

	for i, t := range types {
		checks[i] = bsonutil.WrapInOp(bsonutil.OpEq, vType, t)
	}

	return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpOr, checks))
}
