package evaluator

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/internal/util/bsonutil"
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

func translateConvert(expr interface{}, from, to EvalType) interface{} {
	var targetType string
	switch to {
	case EvalBoolean:
		targetType = "bool"
		// If the from type is a string, convert to int before boolean, because
		// mongo type conversion assumes "false" is the only false
		// string, whereas we actually want '0' to be false and any non-zero
		// integer to be true. As it is now, MongoDB will convert the string '0'
		// to true.
		if from == EvalString {
			expr = bsonutil.WrapInConvert(expr, "int", 0, nil)

			// If the from type is a floating point type, we need to round because
			// -0.4 through 0.4 should be treated as false.
		} else if from == EvalDouble || from == EvalDecimal128 {
			expr = bsonutil.WrapInRound(expr)
		}
	case EvalDecimal128:
		targetType = "decimal"
		if from == EvalObjectID {
			expr = translateConvert(expr, from, EvalDatetime)
		}
	case EvalDouble:
		targetType = "double"
		if from == EvalObjectID {
			expr = translateConvert(expr, from, EvalDatetime)
		}
	case EvalInt32, EvalUint32:
		targetType = "int"
		if from == EvalDecimal128 || from == EvalDouble {
			expr = bsonutil.WrapInRound(expr)
		} else if from == EvalObjectID {
			expr = translateConvert(expr, from, EvalDatetime)
		}
	case EvalInt64, EvalUint64:
		targetType = "long"
		if from == EvalDecimal128 || from == EvalDouble {
			expr = bsonutil.WrapInRound(expr)
		} else if from == EvalObjectID {
			expr = translateConvert(expr, from, EvalDatetime)
		}
	case EvalObjectID:
		targetType = "objectId"
	case EvalString:
		targetType = "string"
		// Bools need to be converted to String as "1" or "0", rather than
		// as "true" and "false".
		cond := bsonutil.WrapInOp(bsonutil.OpEq,
			bsonutil.WrapInType(expr),
			"bool",
		)

		expr = bsonutil.WrapInCond(bsonutil.WrapInCond("1", "0", expr), expr, cond)
	case EvalDatetime, EvalDate:
		targetType = "date"
	default:
		panic(fmt.Errorf("target type %s is not a valid target type for $convert",
			string(EvalTypeToSQLType(to))))
	}

	if from == EvalDate {
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

// translatableToAggregation is an interface for any Expr node that can currently
// be translated to MongoDB Aggregation language.
type translatableToAggregation interface {
	ToAggregationLanguage(*PushdownTranslator) (interface{}, PushdownFailure)
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
	LookupFieldName   FieldNameLookup
	Cfg               *PushdownConfig
	piecewiseDeps     []*NonCorrelatedSubqueryFuture
	correlatedColumns []*CorrelatedSubqueryColumnFuture
}

// NewPushdownTranslator returns a new PushdownTranslator.
func NewPushdownTranslator(cfg *PushdownConfig, lookupFieldName FieldNameLookup) *PushdownTranslator {
	return &PushdownTranslator{Cfg: cfg, LookupFieldName: lookupFieldName}
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

func (t *PushdownTranslator) valueKind() SQLValueKind {
	return t.Cfg.sqlValueKind
}

// nolint: unparam
func (t *PushdownTranslator) versionAtLeast(major, minor, patch uint8) bool {
	return util.VersionAtLeast(t.Cfg.mongoDBVersion, []uint8{major, minor, patch})
}

// ToAggregationLanguage translates the provided SQLExpr into something that can
// be used in an aggregation pipeline. If the provided SQLExpr cannot be
// translated, the second return value will be an error.
func (t *PushdownTranslator) ToAggregationLanguage(e SQLExpr) (interface{}, PushdownFailure) {
	if expr, ok := e.(translatableToAggregation); ok {
		return expr.ToAggregationLanguage(t)
	}
	return nil, newPushdownFailure(
		e.ExprName(),
		"expression is not translatable to the aggregation language",
		"expr", e.String(),
	)
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

// TranslateExpr is a wrapper around ToAggregationLanguage that will fail to
// tranlate the expr if the resulting aggregation exceeds the maximum allowed
// nesting depth for BSON documents.
func (t *PushdownTranslator) TranslateExpr(e SQLExpr) (interface{}, PushdownFailure) {
	doc, err, _ := t.translateExprWithDepth(e)
	return doc, err
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

// TranslatePredicate is a wrapper around ToMatchLanguage that will fail to
// tranlate the expr if the resulting aggregation exceeds the maximum allowed
// nesting depth for BSON documents.
func (t *PushdownTranslator) TranslatePredicate(e SQLExpr) (bson.M, SQLExpr) {
	var doc bson.M
	var expr SQLExpr

	doc, expr, _ = t.translatePredicateWithDepth(e)

	if expr != nil && t.versionAtLeast(3, 6, 0) {
		agg, err := t.TranslateExpr(e)
		if err == nil {
			return bson.M{"$expr": agg}, nil
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
		return nil, expr, 0
	}
	depth := ComputeDocNestingDepthWithMaxDepth(doc, MaxDepth)
	if depth <= MaxDepth {
		return doc, expr, depth
	}
	t.Cfg.lg.Debugf(log.Dev,
		"maximum predicate depth: %d exceeded, cannot pushdown, predicate was: %v",
		MaxDepth,
		e)
	return doc, expr, 0
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

	cons, ok := e.(SQLValue)
	if !ok {
		return nil, newPushdownFailure(
			e.ExprName(),
			"SQLExpr is not a SQLValue",
			"expr", fmt.Sprintf("%#v", e),
		)
	}

	if cons.EvalType() == EvalDecimal128 {
		return t.translateDecimal(cons)
	}

	return cons.Value(), nil
}

func (t *PushdownTranslator) translateDateFormatAsDate(f *SQLScalarFunctionExpr) (interface{}, PushdownFailure) {
	_, ok := f.Func.(*dateFormatFunc)
	if !ok {
		panic("should only call with date_format func as argument")
	}

	formatValue, ok := f.Exprs[1].(SQLValue)
	if !ok {
		return nil, newPushdownFailure(f.ExprName(), "format string argument was not literal")
	}

	date, err := t.TranslateExpr(f.Exprs[0])
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
		parts = append(parts, bson.M{bsonutil.OpMultiply: []interface{}{
			bson.M{bsonutil.OpSubtract: []interface{}{bson.M{bsonutil.OpDayOfYear: date}, 1}},
			uint64(24 * time.Hour / time.Millisecond),
		}})

	} else if !hasDay {
		parts = append(parts, bson.M{bsonutil.OpMultiply: []interface{}{
			bson.M{bsonutil.OpSubtract: []interface{}{bson.M{"$dayOfMonth": date}, 1}},
			uint64(24 * time.Hour / time.Millisecond),
		}})
	}

	if !hasHour {
		parts = append(parts, bson.M{bsonutil.OpMultiply: []interface{}{
			bson.M{"$hour": date},
			uint64(time.Hour / time.Millisecond),
		}})
	}
	if !hasMinute {
		parts = append(parts, bson.M{bsonutil.OpMultiply: []interface{}{
			bson.M{"$minute": date},
			uint64(time.Minute / time.Millisecond),
		}})
	}
	if !hasSecond {
		parts = append(parts, bson.M{bsonutil.OpMultiply: []interface{}{
			bson.M{"$second": date},
			uint64(time.Second / time.Millisecond),
		}})
	}

	parts = append(parts, bson.M{"$millisecond": date})

	var totalMS interface{}
	if len(parts) == 1 {
		totalMS = parts[0]
	} else {
		totalMS = bson.M{bsonutil.OpAdd: parts}
	}

	sub := bson.M{bsonutil.OpSubtract: []interface{}{
		date,
		totalMS,
	}}

	return bsonutil.WrapInNullCheckedCond(
		nil,
		sub,
		date,
	), nil
}

func (t *PushdownTranslator) translateDecimal(cons SQLValue) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			cons.ExprName(),
			"cannot translate SQLValue to decimal on MongoDB < 3.4",
		)
	}

	parsed, err := bson.ParseDecimal128(cons.String())
	if err != nil {
		return nil, newPushdownFailure(
			cons.ExprName(),
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
		if IsUUID(mType) {
			fieldValue, ok = GetBinaryFromExpr(mType, valExpr)
			if !ok {
				return nil, false
			}
		} else if mType == schema.MongoObjectID {
			s, ok := valExpr.(SQLVarchar)
			if !ok {
				return nil, false
			}
			fieldValue = bson.ObjectIdHex(String(s))
		}
	}

	translation := bson.M{name: bson.M{op: fieldValue}}

	if op == bsonutil.OpEq {
		translation = bson.M{name: fieldValue}
	}

	return translation, true
}

func negate(op bson.M) bson.M {
	if len(op) == 1 {
		name, value := getSingleMapEntry(op)
		if strings.HasPrefix(name, "$") {
			switch name {
			case bsonutil.OpOr:
				return bson.M{bsonutil.OpNor: value}
			case bsonutil.OpNor:
				return bson.M{bsonutil.OpOr: value}
			}
		} else if innerOp, ok := value.(bson.M); ok {
			if len(innerOp) == 1 {
				innerName, innerValue := getSingleMapEntry(innerOp)
				if strings.HasPrefix(innerName, "$") {
					switch innerName {
					case bsonutil.OpEq:
						return bson.M{name: bson.M{bsonutil.OpNeq: innerValue}}
					case bsonutil.OpIn:
						return bson.M{name: bson.M{bsonutil.OpNotIn: innerValue}}
					case bsonutil.OpNeq:
						return bson.M{name: innerValue}
					case bsonutil.OpNotIn:
						return bson.M{name: bson.M{bsonutil.OpIn: innerValue}}
					case bsonutil.OpRegex:
						return bson.M{name: bson.M{bsonutil.OpNotIn: []interface{}{innerValue}}}
					case bsonutil.OpNot:
						return bson.M{name: innerValue}
					}

					return bson.M{name: bson.M{bsonutil.OpNot: bson.M{innerName: innerValue}}}
				}
			}
		} else {
			// Logical NOT evaluates to 1 if the operand is 0, to 0
			// if the operand is nonzero, and NOT NULL returns NULL.
			// See https://dev.mysql.com/doc/refman/5.7/en/logical-operators.html#operator_not
			// for more.
			translation := bson.M{name: bson.M{bsonutil.OpNeq: value}}
			if value != nil {
				translation = bson.M{
					bsonutil.OpAnd: []interface{}{
						translation,
						bson.M{name: bson.M{bsonutil.OpNeq: nil}},
					},
				}
			}
			return translation
		}
	}

	// $not only works as a meta operator on a single operator
	// so simulate $not using $nor
	return bson.M{bsonutil.OpNor: []interface{}{op}}
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

	err = NormalizeUUID(mType, bytes)
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
func getProjectedFieldName(fieldName string, fieldType EvalType) interface{} {

	names := strings.Split(fieldName, ".")

	if len(names) == 1 {
		return "$" + fieldName
	}

	value, err := strconv.Atoi(names[len(names)-1])
	// special handling for legacy 2d array
	if err == nil && fieldType == EvalArrNumeric {
		fieldName = fieldName[0:strings.LastIndex(fieldName, ".")]
		return bson.M{"$arrayElemAt": []interface{}{"$" + fieldName, value}}
	}

	return "$" + fieldName
}

var (
	mgoNullLiteral         = bsonutil.WrapInLiteral(nil)
	dateComponentSeparator = []interface{}{"!", "\"", "#", bsonutil.WrapInLiteral("$"), "%", "&", "'",
		"(", ")", "*", "+", ",", "-", ".", "/", ":", ";", "<", "=", ">", "?", "@", "[", "\\", "]",
		"^", "_", "`", "{", "|", "}", "~"}
)

// containsBSONType returns an expression that evaluates to true if types
// contains the BSON type of v.
func containsBSONType(v interface{}, types ...string) bson.M {

	vType := bson.M{bsonutil.OpType: v}
	checks := make([]interface{}, len(types))

	for i, t := range types {
		checks[i] = bsonutil.WrapInOp(bsonutil.OpEq, vType, t)
	}

	return bson.M{bsonutil.OpOr: checks}
}
