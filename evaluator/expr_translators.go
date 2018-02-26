package evaluator

import (
	"encoding/hex"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/10gen/mongo-go-driver/bson"
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

// translatableToAggregation is an interface for any Expr node that can currently
// be translated to MongoDB Aggregation language.
type translatableToAggregation interface {
	ToAggregationLanguage(*PushDownTranslator) (interface{}, bool)
}

// translatableToMatch is an interface for any Expr node that can currently
// be translated to MongoDB Match language.
type translatableToMatch interface {
	ToMatchLanguage(*PushDownTranslator) (bson.M, SQLExpr)
}

// FieldNameLookup is a function that, given a tableName and a columnName, will return
// the field name coming back from mongodb.
type FieldNameLookup func(databaseName, tableName, columnName string) (string, bool)

// PushDownTranslator handles the state necessary to do pushdown translation.
type PushDownTranslator struct {
	LookupFieldName FieldNameLookup
	Ctx             TranslationCtx
}

// NewPushDownTranslator returns a new PushDownTranslator.
func NewPushDownTranslator(lookupFieldName FieldNameLookup, ctx TranslationCtx) *PushDownTranslator {
	return &PushDownTranslator{LookupFieldName: lookupFieldName, Ctx: ctx}
}

// ToAggregationLanguage translates the provided SQLExpr into something that can
// be used in an aggregation pipeline. If the provided SQLExpr cannot be
// translated, the second return value will be false.
func (t *PushDownTranslator) ToAggregationLanguage(e SQLExpr) (interface{}, bool) {
	if expr, ok := e.(translatableToAggregation); ok {
		return expr.ToAggregationLanguage(t)
	}
	return nil, false
}

// ToMatchLanguage translates the provided SQLExpr into something that can
// be used in an match expression. If the SQLExpr can be fully translated, the
// first return value will be the translated expression, and the second will be
// nil. If the provided SQLExpr cannot be fully translated, the first return
// value will be the partially translated expression, and the second will be the
// original SQLExpr.
func (t *PushDownTranslator) ToMatchLanguage(e SQLExpr) (bson.M, SQLExpr) {
	if predicate, ok := e.(translatableToMatch); ok {
		return predicate.ToMatchLanguage(t)
	}
	return nil, e
}

// TranslateExpr is a wrapper around ToAggregationLanguage that will fail to
// tranlate the expr if the resulting aggregation exceeds the maximum allowed
// nesting depth for BSON documents.
func (t *PushDownTranslator) TranslateExpr(e SQLExpr) (interface{}, bool) {
	doc, successful, _ := t.translateExprWithDepth(e)
	return doc, successful
}

func (t *PushDownTranslator) translateExprWithDepth(e SQLExpr) (interface{}, bool, uint32) {
	doc, successful := t.ToAggregationLanguage(e)
	depth := ComputeDocNestingDepthWithMaxDepth(doc, MaxDepth)
	if depth <= MaxDepth {
		return doc, successful, depth
	}
	t.Ctx.Logger().Debugf(log.Dev, "maximum expression depth: %d exceeded, cannot pushdown, expression was: %v", MaxDepth, e)
	return nil, false, 0
}

// TranslatePredicate is a wrapper around ToMatchLanguage that will fail to
// tranlate the expr if the resulting aggregation exceeds the maximum allowed
// nesting depth for BSON documents.
func (t *PushDownTranslator) TranslatePredicate(e SQLExpr) (bson.M, SQLExpr) {
	var doc bson.M
	var expr SQLExpr

	doc, expr, _ = t.translatePredicateWithDepth(e)

	if expr != nil && t.Ctx.VersionAtLeast(3, 6, 0) {
		agg, ok := t.TranslateExpr(e)
		if ok {
			return bson.M{"$expr": agg}, nil
		}
	}

	return doc, expr
}

func (t *PushDownTranslator) translatePredicateWithDepth(e SQLExpr) (bson.M, SQLExpr, uint32) {
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
	t.Ctx.Logger().Debugf(log.Dev, "maximum predicate depth: %d exceeded, cannot pushdown, predicate was: %v", MaxDepth, e)
	return doc, expr, 0
}

func (t *PushDownTranslator) getFieldName(e SQLExpr) (string, bool) {
	switch field := e.(type) {
	case SQLColumnExpr:
		return t.LookupFieldName(field.databaseName, field.tableName, field.columnName)
	default:
		return "", false
	}
}

func (t *PushDownTranslator) getValue(e SQLExpr) (interface{}, bool) {

	cons, ok := e.(SQLValue)
	if !ok {
		return nil, false
	}

	if cons.Type() == schema.SQLDecimal128 {
		return t.translateDecimal(cons)
	}

	return cons.Value(), true
}

func (t *PushDownTranslator) translateDateFormatAsDate(f *SQLScalarFunctionExpr) (interface{}, bool) {
	formatValue, ok := f.Exprs[1].(SQLValue)
	if !ok {
		return nil, false
	}

	date, ok := t.TranslateExpr(f.Exprs[0])
	if !ok {
		return nil, false
	}

	hasYear := false
	hasMonth := false
	hasDay := false
	hasHour := false
	hasMinute := false
	hasSecond := false

	// NOTE: this is a very specific optimization for Tableau's discreet dimension
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
		return nil, false
	}

	var parts []interface{}
	if !hasMonth {
		parts = append(parts, bson.M{mgoOperatorMultiply: []interface{}{
			bson.M{mgoOperatorSubtract: []interface{}{bson.M{mgoOperatorDayOfYear: date}, 1}},
			uint64(24 * time.Hour / time.Millisecond),
		}})

	} else if !hasDay {
		parts = append(parts, bson.M{mgoOperatorMultiply: []interface{}{
			bson.M{mgoOperatorSubtract: []interface{}{bson.M{"$dayOfMonth": date}, 1}},
			uint64(24 * time.Hour / time.Millisecond),
		}})
	}

	if !hasHour {
		parts = append(parts, bson.M{mgoOperatorMultiply: []interface{}{
			bson.M{"$hour": date},
			uint64(time.Hour / time.Millisecond),
		}})
	}
	if !hasMinute {
		parts = append(parts, bson.M{mgoOperatorMultiply: []interface{}{
			bson.M{"$minute": date},
			uint64(time.Minute / time.Millisecond),
		}})
	}
	if !hasSecond {
		parts = append(parts, bson.M{mgoOperatorMultiply: []interface{}{
			bson.M{"$second": date},
			uint64(time.Second / time.Millisecond),
		}})
	}

	parts = append(parts, bson.M{"$millisecond": date})

	var totalMS interface{}
	if len(parts) == 1 {
		totalMS = parts[0]
	} else {
		totalMS = bson.M{mgoOperatorAdd: parts}
	}

	sub := bson.M{mgoOperatorSubtract: []interface{}{
		date,
		totalMS,
	}}

	return wrapInNullCheckedCond(
		nil,
		sub,
		date,
	), true
}

func (t *PushDownTranslator) translateDecimal(cons SQLValue) (interface{}, bool) {
	if !t.Ctx.VersionAtLeast(3, 4, 0) {
		return nil, false
	}

	parsed, err := bson.ParseDecimal128(cons.String())
	if err != nil {
		return nil, false
	}

	return parsed, true
}

func (t *PushDownTranslator) translateOperator(op string, nameExpr, valExpr SQLExpr) (bson.M, bool) {
	name, ok := t.getFieldName(nameExpr)
	if !ok {
		return nil, false
	}

	fieldValue, ok := t.getValue(valExpr)
	if !ok {
		return nil, false
	}

	colExpr, ok := nameExpr.(SQLColumnExpr)
	mType := colExpr.columnType.MongoType
	if ok && IsUUID(mType) {
		binary, ok := GetBinaryFromExpr(mType, valExpr)
		if !ok {
			return nil, false
		}
		fieldValue = binary
	}

	translation := bson.M{name: bson.M{op: fieldValue}}

	if op == mgoOperatorEq {
		translation = bson.M{name: fieldValue}
	}

	return translation, true
}

func negate(op bson.M) bson.M {
	if len(op) == 1 {
		name, value := getSingleMapEntry(op)
		if strings.HasPrefix(name, "$") {
			switch name {
			case mgoOperatorOr:
				return bson.M{mgoOperatorNor: value}
			case mgoOperatorNor:
				return bson.M{mgoOperatorOr: value}
			}
		} else if innerOp, ok := value.(bson.M); ok {
			if len(innerOp) == 1 {
				innerName, innerValue := getSingleMapEntry(innerOp)
				if strings.HasPrefix(innerName, "$") {
					switch innerName {
					case mgoOperatorEq:
						return bson.M{name: bson.M{mgoOperatorNeq: innerValue}}
					case mgoOperatorIn:
						return bson.M{name: bson.M{mgoOperatorNotIn: innerValue}}
					case mgoOperatorNeq:
						return bson.M{name: innerValue}
					case mgoOperatorNotIn:
						return bson.M{name: bson.M{mgoOperatorIn: innerValue}}
					case mgoOperatorRegex:
						return bson.M{name: bson.M{mgoOperatorNotIn: []interface{}{innerValue}}}
					case mgoOperatorNot:
						return bson.M{name: innerValue}
					}

					return bson.M{name: bson.M{mgoOperatorNot: bson.M{innerName: innerValue}}}
				}
			}
		} else {
			// Logical NOT evaluates to 1 if the operand is 0, to 0
			// if the operand is nonzero, and NOT NULL returns NULL.
			// See https://dev.mysql.com/doc/refman/5.7/en/logical-operators.html#operator_not
			// for more.
			translation := bson.M{name: bson.M{mgoOperatorNeq: value}}
			if value != nil {
				translation = bson.M{
					mgoOperatorAnd: []interface{}{
						translation,
						bson.M{name: bson.M{mgoOperatorNeq: nil}},
					},
				}
			}
			return translation
		}
	}

	// $not only works as a meta operator on a single operator
	// so simulate $not using $nor
	return bson.M{mgoOperatorNor: []interface{}{op}}
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
func getProjectedFieldName(fieldName string, fieldType schema.SQLType) interface{} {

	names := strings.Split(fieldName, ".")

	if len(names) == 1 {
		return "$" + fieldName
	}

	value, err := strconv.Atoi(names[len(names)-1])
	// special handling for legacy 2d array
	if err == nil && fieldType == schema.SQLArrNumeric {
		fieldName = fieldName[0:strings.LastIndex(fieldName, ".")]
		return bson.M{"$arrayElemAt": []interface{}{"$" + fieldName, value}}
	}

	return "$" + fieldName
}

//
// Expression Translation Wrappers
//

const (
	mgoOperatorAdd            = "$add"
	mgoOperatorAnd            = "$and"
	mgoOperatorArrElemAt      = "$arrayElemAt"
	mgoOperatorCeil           = "$ceil"
	mgoOperatorConcat         = "$concat"
	mgoOperatorCond           = "$cond"
	mgoOperatorDayOfMonth     = "$dayOfMonth"
	mgoOperatorDayOfWeek      = "$dayOfWeek"
	mgoOperatorDayOfYear      = "$dayOfYear"
	mgoOperatorDateFromParts  = "$dateFromParts"
	mgoOperatorDateFromString = "$dateFromString"
	mgoOperatorDivide         = "$divide"
	mgoOperatorEq             = "$eq"
	mgoOperatorExists         = "$exists"
	mgoOperatorFilter         = "$filter"
	mgoOperatorFloor          = "$floor"
	mgoOperatorGt             = "$gt"
	mgoOperatorGte            = "$gte"
	mgoOperatorHour           = "$hour"
	mgoOperatorIfNull         = "$ifNull"
	mgoOperatorIn             = "$in"
	mgoOperatorIndexOfCP      = "$indexOfCP"
	mgoOperatorLt             = "$lt"
	mgoOperatorLte            = "$lte"
	mgoOperatorLet            = "$let"
	mgoOperatorLiteral        = "$literal"
	mgoOperatorMap            = "$map"
	mgoOperatorMax            = "$max"
	mgoOperatorMin            = "$min"
	mgoOperatorMinute         = "$minute"
	mgoOperatorMillisecond    = "$millisecond"
	mgoOperatorMod            = "$mod"
	mgoOperatorMonth          = "$month"
	mgoOperatorMultiply       = "$multiply"
	mgoOperatorNeq            = "$ne"
	mgoOperatorNotIn          = "$nin"
	mgoOperatorNor            = "$nor"
	mgoOperatorNot            = "$not"
	mgoOperatorOr             = "$or"
	mgoOperatorPow            = "$pow"
	mgoOperatorRange          = "$range"
	mgoOperatorReduce         = "$reduce"
	mgoOperatorRegex          = "$regex"
	mgoOperatorSecond         = "$second"
	mgoOperatorSize           = "$size"
	mgoOperatorSlice          = "$slice"
	mgoOperatorSplit          = "$split"
	mgoOperatorStrlenCP       = "$strLenCP"
	mgoOperatorSubstr         = "$substrCP"
	mgoOperatorSubtract       = "$subtract"
	mgoOperatorSwitch         = "$switch"
	mgoOperatorTrunc          = "$trunc"
	mgoOperatorType           = "$type"
	mgoOperatorYear           = "$year"
)

var (
	mgoNullLiteral         = wrapInLiteral(nil)
	dateComponentSeparator = []interface{}{"!", "\"", "#", wrapInLiteral("$"), "%", "&", "'",
		"(", ")", "*", "+", ",", "-", ".", "/", ":", ";", "<", "=", ">", "?", "@", "[", "\\", "]",
		"^", "_", "`", "{", "|", "}", "~"}
)

// containsBSONType returns an expression that evaluates to true if types contains the BSON type of v.
func containsBSONType(v interface{}, types ...string) bson.M {

	vType := bson.M{mgoOperatorType: v}
	checks := make([]interface{}, len(types))

	for i, t := range types {
		checks[i] = wrapInOp(mgoOperatorEq, vType, t)
	}

	return bson.M{mgoOperatorOr: checks}
}

// getLiteral returns the value of an inner $literal if
// one is present, and nil otherwise.
func getLiteral(v interface{}) (interface{}, bool) {
	if bsonMap, ok := v.(bson.M); ok {
		if bsonVal, ok := bsonMap[mgoOperatorLiteral]; ok {
			return bsonVal, true
		}
	}
	return nil, false
}

// wrapInCase returns an expression to use as one of the branches arguments to wrapInSwitch.
// caseExpr must evaluate to a boolean.
func wrapInCase(caseExpr, thenExpr interface{}) bson.M {
	return bson.M{"case": caseExpr, "then": thenExpr}
}

// wrapInCond returns a document that evalutes to truePart
// if any of conds is true, and falsePart otherwise.
func wrapInCond(truePart, falsePart interface{}, conds ...interface{}) interface{} {
	var condition interface{}

	if len(conds) > 1 {
		condition = bson.M{mgoOperatorOr: conds}
	} else {
		condition = conds[0]
	}

	return bson.M{mgoOperatorCond: []interface{}{condition, truePart, falsePart}}
}

// wrapInDateFormat wraps an Aggregation Expression that evaluates to a date
// in a date_format expression that will use '$dateFromString' to format
// a date to a string.
func wrapInDateFormat(date interface{}, mysqlFormat string) (interface{}, bool) {
	var format string
	for i := 0; i < len(mysqlFormat); i++ {
		if mysqlFormat[i] == '%' {
			if i != len(mysqlFormat)-1 {
				switch mysqlFormat[i+1] {
				case '%':
					format += "%%"
				case 'd':
					format += "%d"
				case 'f':
					format += "%L000"
				case 'H', 'k':
					format += "%H"
				case 'i':
					format += "%M"
				case 'j':
					format += "%j"
				case 'm':
					format += "%m"
				case 's', 'S':
					format += "%S"
				case 'T':
					format += "%H:%M:%S"
				case 'U':
					format += "%U"
				case 'Y':
					format += "%Y"
				default:
					return nil, false
				}
				i++
			} else {
				// MongoDB fails when the last character is a % sign in the format string.
				return nil, false
			}
		} else {
			format += string(mysqlFormat[i])
		}
	}

	return wrapInNullCheckedCond(
		nil,
		bson.M{"$dateToString": bson.M{
			"format": format,
			"date":   date,
		}},
		date,
	), true
}

// wrapInEqCase returns a document that is a case arm that checks equality between expr1 and expr2.
func wrapInEqCase(expr1, expr2, thenExpr interface{}) bson.M {
	caseExpr := wrapInOp(mgoOperatorEq, expr1, expr2)
	return bson.M{"case": caseExpr, "then": thenExpr}
}

// wrapInIfNull returns v if it isn't nil, otherwise, it returns ifNull.
func wrapInIfNull(v, ifNull interface{}) interface{} {
	if value, ok := getLiteral(v); ok {
		if value == nil {
			return ifNull
		}
		return v
	}
	return bson.M{mgoOperatorIfNull: []interface{}{v, ifNull}}
}

// wrapInInRange returns an expression that evaluates to true if val is in range [min, max).
// val must evaluate to a number.
func wrapInInRange(val interface{}, min, max float64) interface{} {
	return wrapInOp(mgoOperatorAnd,
		wrapInOp(mgoOperatorGte, val, min),
		wrapInOp(mgoOperatorLt, val, max))
}

// wrapInIntDiv performs an integer division (truncated division).
func wrapInIntDiv(numerator, denominator interface{}) interface{} {
	return wrapInOp(mgoOperatorTrunc,
		wrapInOp(mgoOperatorDivide, numerator, denominator))
}

// wrapInIsLeap year creates an expression that returns true if the argument is
// a leap year, and false otherwise. This function assume val is an integer
// year.
func wrapInIsLeapYear(val interface{}) bson.M {
	v := "$$val"
	letAssignment := bson.M{
		"val": val,
	}
	// This computes the expression:
	// (v % 4 == 0) && (v % 100 != 0) || (v % 400 == 0).
	return wrapInLet(letAssignment,
		wrapInOp(mgoOperatorOr,
			wrapInOp(mgoOperatorAnd,
				wrapInOp(mgoOperatorEq,
					wrapInOp(mgoOperatorMod, v, 4),
					0),
				wrapInOp(mgoOperatorNeq,
					wrapInOp(mgoOperatorMod, v, 100),
					0),
			),
			wrapInOp(mgoOperatorEq,
				wrapInOp(mgoOperatorMod, v, 400),
				0),
		),
	)
}

// wrapInLet returns a document with v as vars, and i as in.
func wrapInLet(v, i interface{}) bson.M {
	return bson.M{mgoOperatorLet: bson.M{"vars": v, "in": i}}
}

// wrapInLiteral returns a document with v passed to $literal.
func wrapInLiteral(v interface{}) bson.M {
	return bson.M{mgoOperatorLiteral: v}
}

// wrapInMap returns the aggregation expression {$map: {input: input, as: as, in: in }}.
// https://docs.mongodb.com/manual/reference/operator/aggregation/map/
func wrapInMap(input, as, in interface{}) bson.M {
	return bson.M{mgoOperatorMap: bson.M{"input": input, "as": as, "in": in}}
}

func wrapInNullCheck(v interface{}) interface{} {
	if _, ok := getLiteral(v); ok {
		return v
	}
	return wrapInOp(mgoOperatorEq, wrapInIfNull(v, nil), nil)
}

// wrapInNullCheckedCond returns a document that evalutes to truePart
// if any of the null checked conds is true, and falsePart otherwise.
func wrapInNullCheckedCond(truePart, falsePart interface{}, conds ...interface{}) interface{} {
	var condition interface{}
	newConds := []interface{}{}
	for _, cond := range conds {
		if value, ok := getLiteral(cond); !ok {
			newConds = append(newConds, wrapInNullCheck(cond))
		} else if value == nil {
			newConds = append(newConds, true)
		}
	}
	switch len(newConds) {
	case 0:
		return falsePart
	case 1:
		condition = newConds[0]
	default:
		condition = bson.M{mgoOperatorOr: newConds}
	}

	return bson.M{mgoOperatorCond: []interface{}{condition, truePart, falsePart}}
}

// wrapInOp returns a document which passes all arguments to the op.
func wrapInOp(op string, args ...interface{}) interface{} {
	return bson.M{op: args}
}

// wrapInRange returns the aggregation expression {$range: [start, stop, step]}.
// https://docs.mongodb.com/manual/reference/operator/aggregation/range/
func wrapInRange(start, stop, step interface{}) interface{} {
	if step != nil {
		return bson.M{mgoOperatorRange: []interface{}{start, stop, step}}
	}
	return wrapInOp(mgoOperatorRange, start, stop)
}

// wrapInReduce returns the aggregation expression
// {$reduce: {input: input, initialValue: initialValue, in: in }}.
// https://docs.mongodb.com/manual/reference/operator/aggregation/range/
func wrapInReduce(input, initialValue, in interface{}) bson.M {
	return bson.M{mgoOperatorReduce: bson.M{"input": input, "initialValue": initialValue, "in": in}}
}

// wrapInRound returns args rounded.
func wrapInRound(args []interface{}) (bson.M, bool) {
	decimal := 1.0
	if len(args) == 2 {
		bsonMap, ok := args[1].(bson.M)
		if !ok {
			return nil, false
		}

		bsonVal, ok := bsonMap[mgoOperatorLiteral]
		if !ok {
			return nil, false
		}

		placeVal, _ := NewSQLValue(bsonVal, schema.SQLFloat, schema.SQLNone)

		places := placeVal.Float64()

		decimal = math.Pow(float64(10), places)
	}

	if decimal < 1 {
		return wrapInLiteral(0), true
	}

	letAssignment := bson.M{
		"decimal": decimal,
	}

	letEvaluation := bson.M{
		mgoOperatorDivide: []interface{}{
			bson.M{
				mgoOperatorCond: []interface{}{
					bson.M{
						mgoOperatorGte: []interface{}{args[0], 0}},
					bson.M{
						mgoOperatorFloor: bson.M{
							mgoOperatorAdd: []interface{}{
								bson.M{
									mgoOperatorMultiply: []interface{}{
										args[0], "$$decimal",
									},
								},
								0.5,
							},
						},
					},
					bson.M{
						mgoOperatorCeil: bson.M{
							mgoOperatorSubtract: []interface{}{
								bson.M{
									mgoOperatorMultiply: []interface{}{
										args[0], "$$decimal",
									},
								},
								0.5,
							},
						},
					},
				},
			},
			"$$decimal",
		},
	}

	return wrapInLet(letAssignment, letEvaluation), true
}

// wrapInRoundValue generates an expression to round a floating point number
// the way MySQL does. This is the simplest implementation of round I have found:
// https://github.com/golang/go/issues/4594#issuecomment-66073312.
func wrapInRoundValue(val interface{}) interface{} {
	// The MongoDB aggregation language generated by this function implements
	// the following algorithm presented in go code:
	// if x < 0 {
	//      return math.Ceil(x-.5)
	// }
	// return math.Floor(x+.5)
	condExpr := wrapInOp(mgoOperatorLt, val, 0.0)
	lt0 := wrapInOp(mgoOperatorCeil, wrapInOp(mgoOperatorSubtract, val, 0.5))
	gte0 := wrapInOp(mgoOperatorFloor, wrapInOp(mgoOperatorAdd, val, 0.5))
	return wrapInCond(lt0, gte0, condExpr)
}

// wrapInStringToArray converts an expression v (which must evaluate to a string)
// to an array e.g. "hello" -> ["h", "e", "l", "l", "o"] and returns the array.
func wrapInStringToArray(v interface{}) bson.M {
	input := bson.M{mgoOperatorRange: []interface{}{0, bson.M{mgoOperatorStrlenCP: v}}}
	in := bson.M{mgoOperatorSubstr: []interface{}{v, "$$this", 1}}
	return bson.M{mgoOperatorMap: bson.M{"input": input, "in": in}}
}

// wrapInSwitch returns the aggregation expression
// {$switch: branches: branches, default: defaultExpr }
// https://docs.mongodb.com/manual/reference/operator/aggregation/switch/
func wrapInSwitch(defaultExpr interface{}, branches ...bson.M) bson.M {
	return bson.M{mgoOperatorSwitch: bson.M{"branches": branches, "default": defaultExpr}}
}

// wrapLRTrim returns a trimmed version of args.
func wrapLRTrim(isLTrimType bool, args interface{}) interface{} {
	var (
		splitArray   = bson.M{mgoOperatorSplit: []interface{}{args, " "}}
		substrIndex  interface{}
		substrLength interface{}
	)

	if !isLTrimType {
		splitArray = bson.M{"$reverseArray": splitArray}
	}

	mapInput := wrapInLet(bson.M{"splitArray": splitArray},
		bson.M{"$zip": bson.M{
			"inputs": []interface{}{
				"$$splitArray",
				bson.M{mgoOperatorRange: []interface{}{
					0,
					bson.M{mgoOperatorSize: "$$splitArray"}}}}}})

	mapIn := wrapInCond(bson.M{mgoOperatorStrlenCP: args},
		bson.M{mgoOperatorArrElemAt: []interface{}{"$$zipArray", 1}},
		bson.M{mgoOperatorEq: []interface{}{
			bson.M{mgoOperatorArrElemAt: []interface{}{"$$zipArray", 0}}, ""}})

	min := bson.M{mgoOperatorMin: wrapInMap(mapInput, "zipArray", mapIn)}

	if isLTrimType {
		substrIndex = min
		substrLength = bson.M{mgoOperatorStrlenCP: args}
	} else {
		substrIndex = 0
		substrLength = bson.M{mgoOperatorSubtract: []interface{}{
			bson.M{mgoOperatorStrlenCP: args},
			min}}
	}

	return bson.M{
		mgoOperatorSubstr: []interface{}{
			args,
			substrIndex,
			substrLength,
		},
	}
}

// wrapSingleArgFuncWithNullCheck returns a null checked version
// of the arg passed to name.
func wrapSingleArgFuncWithNullCheck(name string, arg interface{}) interface{} {
	return wrapInNullCheckedCond(nil, bson.M{name: arg}, arg)
}

// wrapInWeekCalcluation calculates the week of a given date based on the
// passed argument, expr, which is some MongoDB Aggregation Pipeline
// expression, and the mode, which is an integer.
func wrapInWeekCalculation(expr interface{}, mode int64) interface{} {
	date, year := "$$date", "$$year"
	getJan1 := func() interface{} {
		return bson.M{
			mgoOperatorDateFromParts: bson.M{
				"year":  year,
				"month": 1,
				"day":   1,
			},
		}
	}

	getNextJan1 := func() interface{} {
		return bson.M{
			mgoOperatorDateFromParts: bson.M{
				"year":  wrapInOp(mgoOperatorAdd, year, 1),
				"month": 1,
				"day":   1,
			},
		}
	}

	// generateDaySubtract generates the main week calculation shared
	// by all modes except 0, 2 (since those can use MongoDB's week function).
	// The calculation is:
	// trunc((date - dayOne) / (7 * MillisecondsPerDay) + 1).
	generateDaySubtract := func(dayOne interface{}) interface{} {
		return wrapInOp(mgoOperatorTrunc,
			wrapInOp(mgoOperatorAdd,
				wrapInOp(mgoOperatorDivide,
					wrapInOp(mgoOperatorSubtract, date, dayOne),
					7*MillisecondsPerDay),
				1),
		)
	}

	// generate4DaysBody generates the body for modes where the first
	// week is defined by having 4 days in the year, these are modes
	// 1, 3, 4, and 6.
	generate4DaysBody := func(diffConstant int) interface{} {
		// This description is used for Monday as first day of the
		// week. See below for an explanation of the Sunday first day
		// case. Calculate the first day of the first week of this
		// year based on the dayOfWeek of YYYY-01-01 of this year, note
		// that it may be from the previous year. The Day Diff column
		// is the
		// amount of days to Add or Subtract from YYYY-01-01:
		// Day Of Week for Jan 1   |   Day Diff
		// ---------------------------------------------
		//                     1   |   + 1
		//                     2   |   + 0
		//                     3   |   - 1
		//                     4   |   - 2
		//                     5   |   - 3
		//                     6   |   + 3
		//                     7   |   + 2
		// This can be simplified to:
		// diff = -x + 2
		// if diff > -4 {
		//      return diff
		// }
		// return diff + 7
		// for the Sunday version of this, we need to use -x + 1
		// instead of -x + 2, and that is the only difference, that is
		// the point of "diffConstant", it will be either 1 or 2.
		jan1 := getJan1()
		jan1DayOfWeek := "$$jan1DayOfWeek"
		dayOfWeekLetAssignment := bson.M{
			"jan1DayOfWeek": wrapInOp(mgoOperatorDayOfWeek, jan1),
		}
		dayOne := wrapInOp(mgoOperatorAdd, jan1,
			wrapInOp(mgoOperatorMultiply,
				wrapInLet(
					bson.M{"diff": wrapInOp(mgoOperatorAdd,
						wrapInOp(mgoOperatorMultiply, jan1DayOfWeek, -1),
						diffConstant),
					},
					wrapInCond("$$diff",
						wrapInOp(mgoOperatorAdd, "$$diff", 7),
						wrapInOp(mgoOperatorGt, "$$diff", -4),
					),
				),
				MillisecondsPerDay,
			),
		)
		return wrapInLet(dayOfWeekLetAssignment, generateDaySubtract(dayOne))
	}

	// generateMondayBody generates the body for modes where the first
	// week is defined by having a Monday, these are modes
	// 5 and 7.
	generateMondayBody := func() interface{} {
		// These are more simple than the 4 days mode. The diff from Jan1
		// can be defined using (7 - x + 2) % 7.
		jan1 := getJan1()
		jan1DayOfWeek := "$$jan1DayOfWeek"
		dayOfWeekLetAssignment := bson.M{
			"jan1DayOfWeek": wrapInOp(mgoOperatorDayOfWeek, jan1),
		}
		dayOne := wrapInOp(mgoOperatorAdd, jan1,
			wrapInOp(mgoOperatorMultiply,
				wrapInOp(mgoOperatorMod,
					wrapInOp(mgoOperatorAdd,
						wrapInOp(mgoOperatorSubtract,
							7,
							jan1DayOfWeek,
						),
						2),
					7),
				MillisecondsPerDay,
			),
		)
		return wrapInLet(dayOfWeekLetAssignment, generateDaySubtract(dayOne))
	}

	// wrapInZeroCheck - half of all modes allow weeks numbers
	// between 0-53, and the other half allow 1-53. To compute the week
	// for modes allowing 1-53, we compute the week for the associated 0-53
	// mode, and if it results in week 0, we return week('(year-1)-12-31'),
	// which will be either 52 or 53 as the 1-53 modes consider such a date
	// as being in the previous year. This means that
	// wrapInWeekCalculation must be recursive, which is why it is
	// separated from the FuncToAggregation for weekFunc. Note that the
	// recursive step is, at most, depth 1, because only used in 1-53
	// modes, but recursively calls with a 0-53 mode.
	wrapInZeroCheck := func(body interface{}, m int64) interface{} {
		lastDayLastYear := bson.M{
			mgoOperatorDateFromParts: bson.M{
				"year":  wrapInOp(mgoOperatorSubtract, year, 1),
				"month": 12,
				"day":   31,
			},
		}
		output := "$$output"
		letAssignment := bson.M{
			"output": body,
		}
		return wrapInLet(letAssignment,
			wrapInCond(output,
				wrapInWeekCalculation(lastDayLastYear, m),
				wrapInOp(mgoOperatorNeq, output, 0),
			),
		)
	}

	// wrapInFiftyThreeCheck is used to handle cases where the last week of a
	// year may actually map as the first week of the next year. This is
	// only possible in the cases where the first week is defined by having
	// 4 days in the year, and where 0 weeks are not allowed, so that is
	// modes 3 and 6. In these modes it is possible that 12-31, 12-30, and even
	// 12-29 map to week 1 of the next year. This is similar in design to
	// zeroCheck, except that it is only needed in the modes with 4 days
	// used to decide the first week of the month. We only need to check
	// the day if our computeDaySubtract results in week 53, giving us
	// faster common cases. janOneDaysOfWeek are the days of the week
	// for the next Jan-1 that result in one of the last three days
	// of the year potentially mapping to the next year. Note that
	// MongoDB aggregation pipeline numbers days 1-7, with 1 being Sunday.
	wrapInFiftyThreeCheck := func(body interface{}, janOneDaysOfWeek ...int) interface{} {
		output, day := "$$output", "$$day"
		outputLetAssignment := bson.M{
			"output": body,
		}
		nextJan1 := getNextJan1()
		// Day Of Week for Jan 1  |  First Day In December Mapping to Next Year
		// --------------------------------------------------------------------
		// janOneDaysOfWeek[0]    |  29
		// janOneDaysOfWeek[1]    |  30
		// janOneDaysOfWeek[2]    |  31
		nextJan1DayOfWeek := wrapInOp(mgoOperatorDayOfWeek, nextJan1)
		return wrapInLet(outputLetAssignment,
			wrapInCond(
				wrapInLet(
					bson.M{
						"day": wrapInOp(mgoOperatorDayOfMonth, date),
					},
					wrapInSwitch(
						53,
						wrapInEqCase(nextJan1DayOfWeek, janOneDaysOfWeek[0], wrapInCond(1, 53, wrapInOp(mgoOperatorGte, day, 29))),
						wrapInEqCase(nextJan1DayOfWeek, janOneDaysOfWeek[1], wrapInCond(1, 53, wrapInOp(mgoOperatorGte, day, 30))),
						wrapInEqCase(nextJan1DayOfWeek, janOneDaysOfWeek[2], wrapInCond(1, 53, wrapInOp(mgoOperatorGte, day, 31))),
					),
				),
				output,
				wrapInOp(mgoOperatorEq, output, 53),
			),
		)
	}

	var body interface{}
	switch mode {
	// First day of week: Sunday, with a Sunday in this year.
	// This is what MongoDB's $week function does, so we use it.
	case 0, 2:
		body = wrapSingleArgFuncWithNullCheck("$week", date)
		if mode == 2 {
			body = wrapInZeroCheck(body, 0)
		}
	// First day of week: Monday, with 4 days in this year.
	case 1, 3:
		body = generate4DaysBody(2)
		if mode == 3 {
			body = wrapInZeroCheck(body, 1)
			body = wrapInFiftyThreeCheck(body, 5, 4, 3)
		}
	// First day of week: Sunday, with 4 days in this year.
	case 4, 6:
		body = generate4DaysBody(1)
		if mode == 6 {
			body = wrapInZeroCheck(body, 4)
			body = wrapInFiftyThreeCheck(body, 4, 3, 2)
		}
	// First day of week: Monday, with a Monday in this year.
	case 5, 7:
		body = generateMondayBody()
		if mode == 7 {
			body = wrapInZeroCheck(body, 5)
		}
	}

	// Bind expressions that would be expensive to recompute.
	return wrapInLet(
		bson.M{
			"date": expr,
		}, wrapInLet(
			bson.M{
				"year": wrapInOp(mgoOperatorYear, date),
			}, body),
	)
}
