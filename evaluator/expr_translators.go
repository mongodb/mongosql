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
	// maxDepth allowed by BIC's MongoDB driver is 200, keep this up to date
	// if this number should happen to change.  We set it less than 200
	// because expressions generated here can be included in stages
	// and higher expressions created in other places that would also
	// count against the total depth.  It is better to err on the
	// side of caution than hope that we will always have the
	// exact total correct in all contexts.
	maxDepth = 180
)

const (
	mgoOperatorAdd       = "$add"
	mgoOperatorAnd       = "$and"
	mgoOperatorArrElemAt = "$arrayElemAt"
	mgoOperatorCeil      = "$ceil"
	mgoOperatorConcat    = "$concat"
	mgoOperatorCond      = "$cond"
	mgoOperatorDivide    = "$divide"
	mgoOperatorEq        = "$eq"
	mgoOperatorExists    = "$exists"
	mgoOperatorGt        = "$gt"
	mgoOperatorGte       = "$gte"
	mgoOperatorIfNull    = "$ifNull"
	mgoOperatorLt        = "$lt"
	mgoOperatorLte       = "$lte"
	mgoOperatorLet       = "$let"
	mgoOperatorMap       = "$map"
	mgoOperatorMax       = "$max"
	mgoOperatorMin       = "$min"
	mgoOperatorMod       = "$mod"
	mgoOperatorNeq       = "$ne"
	mgoOperatorNot       = "$not"
	mgoOperatorOr        = "$or"
	mgoOperatorRange     = "$range"
	mgoOperatorReduce    = "$reduce"
	mgoOperatorSplit     = "$split"
	mgoOperatorStrlenCP  = "$strLenCP"
	mgoOperatorSubstr    = "$substrCP"
	mgoOperatorSubtract  = "$subtract"
	mgoOperatorSwitch    = "$switch"
	mgoOperatorTrunc     = "$trunc"
	mgoOperatorType      = "$type"
)

var (
	mgoNullLiteral         = bson.M{"$literal": nil}
	mgo0Literal            = bson.M{"$literal": 0}
	mgo1Literal            = bson.M{"$literal": 1}
	dateComponentSeparator = []interface{}{"!", "\"", "#", bson.M{"$literal": "$"}, "%", "&", "'",
		"(", ")", "*", "+", ",", "-", ".", "/", ":", ";", "<", "=", ">", "?", "@", "[", "\\", "]",
		"^", "_", "`", "{", "|", "}", "~"}
)

var (
	getLiteral = func(v interface{}) (interface{}, bool) {
		if bsonMap, ok := v.(bson.M); ok {
			if bsonVal, ok := bsonMap["$literal"]; ok {
				return bsonVal, true
			}
		}
		return nil, false
	}

	wrapInLet = func(vars, in interface{}) bson.M {
		return bson.M{mgoOperatorLet: bson.M{"vars": vars, "in": in}}
	}

	// wrapInMap returns the aggregation expression {$map: {input: input, as: as, in: in }}.
	// https://docs.mongodb.com/manual/reference/operator/aggregation/map/
	wrapInMap = func(input, as, in interface{}) bson.M {
		return bson.M{mgoOperatorMap: bson.M{"input": input, "as": as, "in": in}}
	}

	// wrapInRange returns the aggregation expression {$range: [start, stop, step]}.
	// https://docs.mongodb.com/manual/reference/operator/aggregation/range/
	wrapInRange = func(start, stop, step interface{}) interface{} {
		if step != nil {
			return bson.M{mgoOperatorRange: []interface{}{start, stop, step}}
		}
		return wrapInOp(mgoOperatorRange, start, stop)
	}

	// wrapInReduce returns the aggregation expression
	// {$reduce: {input: input, initialValue: initialValue, in: in }}.
	// https://docs.mongodb.com/manual/reference/operator/aggregation/range/
	wrapInReduce = func(input, initialValue, in interface{}) bson.M {
		return bson.M{mgoOperatorReduce: bson.M{"input": input, "initialValue": initialValue, "in": in}}
	}

	// wrapInSwitch returns the aggregation expression
	// {$switch: branches: branches, default: defaultExpr }
	// https://docs.mongodb.com/manual/reference/operator/aggregation/switch/
	wrapInSwitch = func(defaultExpr interface{}, branches ...bson.M) bson.M {
		return bson.M{mgoOperatorSwitch: bson.M{"branches": branches, "default": defaultExpr}}
	}

	// wrapInCase returns an expression to use as one of the branches arguments to wrapInSwitch.
	// caseExpr must evaluate to a boolean.
	wrapInCase = func(caseExpr, thenExpr interface{}) bson.M {
		return bson.M{"case": caseExpr, "then": thenExpr}
	}

	// wrapInInRange returns an expression that evaluates to true if val is in range [min, max).
	// val must evaluate to a number.
	wrapInInRange = func(val interface{}, min, max float64) interface{} {
		return wrapInOp(mgoOperatorAnd,
			wrapInOp(mgoOperatorGte, val, min),
			wrapInOp(mgoOperatorLt, val, max))
	}

	// containsBSONType returns an expression that evaluates to true if types contains the BSON type of v.
	containsBSONType = func(v interface{}, types ...string) bson.M {

		vType := bson.M{mgoOperatorType: v}
		checks := make([]interface{}, len(types))

		for i, t := range types {
			checks[i] = wrapInOp(mgoOperatorEq, vType, t)
		}

		return bson.M{mgoOperatorOr: checks}
	}

	wrapLRTrim = func(isLTrimType bool, args interface{}) interface{} {
		var (
			splitArray   = bson.M{"$split": []interface{}{args, " "}}
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
					bson.M{"$range": []interface{}{
						0,
						bson.M{"$size": "$$splitArray"}}}}}})

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

	wrapInRound = func(args []interface{}) (bson.M, bool) {
		decimal := 1.0
		if len(args) == 2 {
			bsonMap, ok := args[1].(bson.M)
			if !ok {
				return nil, false
			}

			bsonVal, ok := bsonMap["$literal"]
			if !ok {
				return nil, false
			}

			placeVal, _ := NewSQLValue(bsonVal, schema.SQLFloat, schema.SQLNone)

			places := placeVal.Float64()

			decimal = math.Pow(float64(10), places)
		}

		if decimal < 1 {
			return bson.M{"$literal": 0}, true
		}

		letAssignment := bson.M{
			"decimal": decimal,
		}

		letEvaluation := bson.M{
			"$divide": []interface{}{
				bson.M{
					mgoOperatorCond: []interface{}{
						bson.M{
							mgoOperatorGte: []interface{}{args[0], 0}},
						bson.M{
							"$floor": bson.M{
								"$add": []interface{}{
									bson.M{
										"$multiply": []interface{}{
											args[0], "$$decimal",
										},
									},
									0.5,
								},
							},
						},
						bson.M{
							"$ceil": bson.M{
								"$subtract": []interface{}{
									bson.M{
										"$multiply": []interface{}{
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

	wrapInCond = func(truePart, falsePart interface{}, conds ...interface{}) interface{} {
		var condition interface{}

		if len(conds) > 1 {
			condition = bson.M{mgoOperatorOr: conds}
		} else {
			condition = conds[0]
		}

		return bson.M{mgoOperatorCond: []interface{}{condition, truePart, falsePart}}
	}

	wrapInOp = func(op string, left, right interface{}) interface{} {
		return bson.M{op: []interface{}{left, right}}
	}

	wrapInIfNull = func(v, ifNull interface{}) interface{} {
		if value, ok := getLiteral(v); ok {
			if value == nil {
				return ifNull
			}
			return v
		}
		return bson.M{mgoOperatorIfNull: []interface{}{v, ifNull}}
	}

	wrapInNullCheck = func(v interface{}) interface{} {
		if _, ok := getLiteral(v); ok {
			return v
		}
		return wrapInOp(mgoOperatorEq, wrapInIfNull(v, nil), nil)
	}

	wrapInNullCheckedCond = func(truePart, falsePart interface{}, conds ...interface{}) interface{} {
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

	wrapSingleArgFuncWithNullCheck = func(name string, arg interface{}) interface{} {
		return wrapInNullCheckedCond(nil, bson.M{name: arg}, arg)
	}

	// wrapInStringToArray converts an expression v (which must evaluate to a string) to an array.
	// For example, "hello" -> ["h", "e", "l", "l", "o"].
	wrapInStringToArray = func(v interface{}) bson.M {
		input := bson.M{mgoOperatorRange: []interface{}{0, bson.M{mgoOperatorStrlenCP: v}}}
		in := bson.M{mgoOperatorSubstr: []interface{}{v, "$$this", 1}}
		return bson.M{mgoOperatorMap: bson.M{"input": input, "in": in}}
	}

	// wrapInArrayToString combines an expression (which much evaluate to an array) into a single string.
	wrapInArrayToString = func(v interface{}) bson.M {
		return wrapInReduce(v, "", wrapInOp(mgoOperatorConcat, "$$value", "$$this"))
	}
)

// translatableToAggregation is an interface for any Expr node that can currently
// be translated to MongoDB Aggregation language.
type translatableToAggregation interface {
	ToAggregationLanguage(*pushDownTranslator) (interface{}, bool)
}

// translatableToMatch is an interface for any Expr node that can currently
// be translated to MongoDB Match language.
type translatableToMatch interface {
	ToMatchLanguage(*pushDownTranslator) (bson.M, SQLExpr)
}

// fieldNameLookup is a function that, given a tableName and a columnName, will return
// the field name coming back from mongodb.
type fieldNameLookup func(databaseName, tableName, columnName string) (string, bool)

type pushDownTranslator struct {
	versionAtLeast  func(...uint8) bool
	lookupFieldName fieldNameLookup
	logger          *log.Logger
}

func (t *pushDownTranslator) ToAggregationLanguage(e SQLExpr) (interface{}, bool) {
	if expr, ok := e.(translatableToAggregation); ok {
		return expr.ToAggregationLanguage(t)
	}
	return nil, false
}

func (t *pushDownTranslator) ToMatchLanguage(e SQLExpr) (bson.M, SQLExpr) {
	if predicate, ok := e.(translatableToMatch); ok {
		return predicate.ToMatchLanguage(t)
	}
	return nil, e
}

func (t *pushDownTranslator) TranslateExpr(e SQLExpr) (interface{}, bool) {
	doc, successful, _ := t.TranslateExprWithDepth(e)
	return doc, successful
}

func (t *pushDownTranslator) TranslateExprWithDepth(e SQLExpr) (interface{}, bool, uint32) {
	doc, successful := t.ToAggregationLanguage(e)
	depth := computeDocNestingDepthWithMaxDepth(doc, maxDepth)
	if depth <= maxDepth {
		return doc, successful, depth
	}
	t.logger.Debugf(log.Dev, "maximum expression depth: %d exceeded, cannot pushdown, expression was: %v", maxDepth, e)
	return nil, false, 0
}

func (t *pushDownTranslator) TranslatePredicate(e SQLExpr) (bson.M, SQLExpr) {
	doc, expr, _ := t.TranslatePredicateWithDepth(e)
	return doc, expr
}

func (t *pushDownTranslator) TranslatePredicateWithDepth(e SQLExpr) (bson.M, SQLExpr, uint32) {
	translatable, ok := e.(translatableToMatch)
	if !ok {
		return nil, e, 0
	}
	doc, expr := translatable.ToMatchLanguage(t)
	if doc == nil {
		return nil, expr, 0
	}
	depth := computeDocNestingDepthWithMaxDepth(doc, maxDepth)
	if depth <= maxDepth {
		return doc, expr, depth
	}
	t.logger.Debugf(log.Dev, "maximum predicate depth: %d exceeded, cannot pushdown, predicate was: %v", maxDepth, e)
	return doc, expr, 0
}

func (t *pushDownTranslator) getFieldName(e SQLExpr) (string, bool) {
	switch field := e.(type) {
	case SQLColumnExpr:
		return t.lookupFieldName(field.databaseName, field.tableName, field.columnName)
	default:
		return "", false
	}
}

func (t *pushDownTranslator) getValue(e SQLExpr) (interface{}, bool) {

	cons, ok := e.(SQLValue)
	if !ok {
		return nil, false
	}

	if cons.Type() == schema.SQLDecimal128 {
		return t.translateDecimal(cons)
	}

	return cons.Value(), true
}

func (t *pushDownTranslator) translateDateFormatAsDate(f *SQLScalarFunctionExpr) (interface{}, bool) {
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
		parts = append(parts, bson.M{"$multiply": []interface{}{
			bson.M{"$subtract": []interface{}{bson.M{"$dayOfYear": date}, 1}},
			uint64(24 * time.Hour / time.Millisecond),
		}})

	} else if !hasDay {
		parts = append(parts, bson.M{"$multiply": []interface{}{
			bson.M{"$subtract": []interface{}{bson.M{"$dayOfMonth": date}, 1}},
			uint64(24 * time.Hour / time.Millisecond),
		}})
	}

	if !hasHour {
		parts = append(parts, bson.M{"$multiply": []interface{}{
			bson.M{"$hour": date},
			uint64(time.Hour / time.Millisecond),
		}})
	}
	if !hasMinute {
		parts = append(parts, bson.M{"$multiply": []interface{}{
			bson.M{"$minute": date},
			uint64(time.Minute / time.Millisecond),
		}})
	}
	if !hasSecond {
		parts = append(parts, bson.M{"$multiply": []interface{}{
			bson.M{"$second": date},
			uint64(time.Second / time.Millisecond),
		}})
	}

	parts = append(parts, bson.M{"$millisecond": date})

	var totalMS interface{}
	if len(parts) == 1 {
		totalMS = parts[0]
	} else {
		totalMS = bson.M{"$add": parts}
	}

	sub := bson.M{"$subtract": []interface{}{
		date,
		totalMS,
	}}

	return wrapInNullCheckedCond(
		nil,
		sub,
		date,
	), true
}

func (t *pushDownTranslator) translateDecimal(cons SQLValue) (interface{}, bool) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, false
	}

	parsed, err := bson.ParseDecimal128(cons.String())
	if err != nil {
		return nil, false
	}

	return parsed, true
}

func (t *pushDownTranslator) translateOperator(op string, nameExpr, valExpr SQLExpr) (bson.M, bool) {
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
	if ok && isUUID(mType) {
		binary, ok := getBinaryFromExpr(mType, valExpr)
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
				return bson.M{"$nor": value}
			case "$nor":
				return bson.M{mgoOperatorOr: value}
			}
		} else if innerOp, ok := value.(bson.M); ok {
			if len(innerOp) == 1 {
				innerName, innerValue := getSingleMapEntry(innerOp)
				if strings.HasPrefix(innerName, "$") {
					switch innerName {
					case mgoOperatorEq:
						return bson.M{name: bson.M{mgoOperatorNeq: innerValue}}
					case "$in":
						return bson.M{name: bson.M{"$nin": innerValue}}
					case mgoOperatorNeq:
						return bson.M{name: innerValue}
					case "$nin":
						return bson.M{name: bson.M{"$in": innerValue}}
					case "$regex":
						return bson.M{name: bson.M{"$nin": []interface{}{innerValue}}}
					case "$not":
						return bson.M{name: innerValue}
					}

					return bson.M{name: bson.M{"$not": bson.M{innerName: innerValue}}}
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
	return bson.M{"$nor": []interface{}{op}}
}

// getBinaryFromExpr attempts to convert e to a bson.Binary -
// that represents a MongoDB UUID - using mType.
func getBinaryFromExpr(mType schema.MongoType, e SQLExpr) (bson.Binary, bool) {
	// we accept UUIDs as string arguments
	uuidString := strings.Replace(e.String(), "-", "", -1)
	bytes, err := hex.DecodeString(uuidString)
	if err != nil {
		return bson.Binary{}, false
	}

	err = normalizeUUID(mType, bytes)
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
