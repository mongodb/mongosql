package evaluator

import (
	"strings"

	"github.com/mongodb/mongo-tools/common/log"
	"gopkg.in/mgo.v2/bson"
)

// a function that, given a tableName and a columnName, will return
// the field name coming back from mongodb.
type fieldNameLookup func(tableName, columName string) (string, bool)

// TranslateExpr attempts to turn the SQLExpr into MongoD query language.
func TranslateExpr(e SQLExpr, lookupFieldName fieldNameLookup) (interface{}, bool) {

	switch typedE := e.(type) {

	case *SQLAddExpr:

		left, ok := TranslateExpr(typedE.left, lookupFieldName)
		if !ok {
			return nil, false
		}

		right, ok := TranslateExpr(typedE.right, lookupFieldName)
		if !ok {
			return nil, false
		}

		return bson.M{"$add": []interface{}{left, right}}, true

	case *SQLAggFunctionExpr:

		transExpr, ok := TranslateExpr(typedE.Exprs[0], lookupFieldName)
		if !ok {
			return nil, false
		}

		name := typedE.Name

		if name == "count" && typedE.Exprs[0] == SQLString("*") {
			return bson.M{"$size": transExpr}, true
		} else if name == "count" {
			// The below ensure that nulls, undefined, and missing fields
			// are not part of the count.
			return bson.M{
				"$sum": bson.M{
					"$map": bson.M{
						"input": transExpr,
						"as":    "i",
						"in": bson.M{
							"$cond": []interface{}{
								bson.M{"$eq": []interface{}{
									bson.M{"$ifNull": []interface{}{
										"$$i",
										nil}},
									nil}},
								0,
								1,
							},
						},
					},
				},
			}, true
		}

		return bson.M{"$" + name: transExpr}, true

	case *SQLAndExpr:

		left, ok := TranslateExpr(typedE.left, lookupFieldName)
		if !ok {
			return nil, false
		}

		right, ok := TranslateExpr(typedE.right, lookupFieldName)
		if !ok {
			return nil, false
		}

		return bson.M{"$and": []interface{}{left, right}}, true

	case *SQLCtorExpr:

		expr, err := typedE.Evaluate(nil)
		if err != nil {
			return nil, false
		}

		return TranslateExpr(expr, lookupFieldName)

	case SQLDate:

		return typedE.Time, true

	case SQLDateTime:

		return typedE.Time, true

	case *SQLDivideExpr:

		left, ok := TranslateExpr(typedE.left, lookupFieldName)
		if !ok {
			return nil, false
		}

		right, ok := TranslateExpr(typedE.right, lookupFieldName)
		if !ok {
			return nil, false
		}

		return bson.M{"$divide": []interface{}{left, right}}, true

	case *SQLEqualsExpr:

		left, ok := TranslateExpr(typedE.left, lookupFieldName)
		if !ok {
			return nil, false
		}

		right, ok := TranslateExpr(typedE.right, lookupFieldName)
		if !ok {
			return nil, false
		}

		return bson.M{"$eq": []interface{}{left, right}}, true

	case SQLFieldExpr:

		name, ok := lookupFieldName(typedE.tableName, typedE.fieldName)
		if !ok {
			return nil, false
		}

		return "$" + name, true

	case *SQLGreaterThanExpr:

		left, ok := TranslateExpr(typedE.left, lookupFieldName)
		if !ok {
			return nil, false
		}

		right, ok := TranslateExpr(typedE.right, lookupFieldName)
		if !ok {
			return nil, false
		}

		return bson.M{"$gt": []interface{}{left, right}}, true

	case *SQLGreaterThanOrEqualExpr:

		left, ok := TranslateExpr(typedE.left, lookupFieldName)
		if !ok {
			return nil, false
		}

		right, ok := TranslateExpr(typedE.right, lookupFieldName)
		if !ok {
			return nil, false
		}

		return bson.M{"$gte": []interface{}{left, right}}, true

	case *SQLLessThanExpr:

		left, ok := TranslateExpr(typedE.left, lookupFieldName)
		if !ok {
			return nil, false
		}

		right, ok := TranslateExpr(typedE.right, lookupFieldName)
		if !ok {
			return nil, false
		}

		return bson.M{"$lt": []interface{}{left, right}}, true

	case *SQLLessThanOrEqualExpr:

		left, ok := TranslateExpr(typedE.left, lookupFieldName)
		if !ok {
			return nil, false
		}

		right, ok := TranslateExpr(typedE.right, lookupFieldName)
		if !ok {
			return nil, false
		}

		return bson.M{"$lte": []interface{}{left, right}}, true

	case *SQLMultiplyExpr:

		left, ok := TranslateExpr(typedE.left, lookupFieldName)
		if !ok {
			return nil, false
		}

		right, ok := TranslateExpr(typedE.right, lookupFieldName)
		if !ok {
			return nil, false
		}

		return bson.M{"$multiply": []interface{}{left, right}}, true

	case *SQLNotExpr:

		op, ok := TranslateExpr(typedE.operand, lookupFieldName)
		if !ok {
			return nil, false
		}

		return bson.M{"$not": []interface{}{op}}, true

	case *SQLNotEqualsExpr:

		left, ok := TranslateExpr(typedE.left, lookupFieldName)
		if !ok {
			return nil, false
		}

		right, ok := TranslateExpr(typedE.right, lookupFieldName)
		if !ok {
			return nil, false
		}

		return bson.M{"$ne": []interface{}{left, right}}, true

	case *SQLNullCmpExpr:

		op, ok := TranslateExpr(typedE.operand, lookupFieldName)
		if !ok {
			return nil, false
		}

		return bson.M{"$eq": []interface{}{op, nil}}, true

	case SQLNullValue:

		return nil, true

	case *SQLOrExpr:

		left, ok := TranslateExpr(typedE.left, lookupFieldName)
		if !ok {
			return nil, false
		}

		right, ok := TranslateExpr(typedE.right, lookupFieldName)
		if !ok {
			return nil, false
		}

		return bson.M{"$or": []interface{}{left, right}}, true

	case *SQLScalarFunctionExpr:

		// TODO: can use some data and string aggregation functions here

	case *SQLSubqueryCmpExpr:

		// unsupported

	case *SQLSubtractExpr:

		left, ok := TranslateExpr(typedE.left, lookupFieldName)
		if !ok {
			return nil, false
		}

		right, ok := TranslateExpr(typedE.right, lookupFieldName)
		if !ok {
			return nil, false
		}

		return bson.M{"$subtract": []interface{}{left, right}}, true

	case SQLInt, SQLUint32, SQLFloat, SQLBool, SQLString:

		return bson.M{"$literal": typedE}, true

	case SQLTime:

		return typedE.Time, true

	case SQLTimestamp:

		return typedE.Time, true

		/*
			TODO: implement these
			case *SQLCaseExpr:
			case *SQLUnaryMinusExpr:
			case *SQLUnaryTildeExpr:
			case *SQLTupleExpr:
			case *SQLInExpr:
		*/

	}

	log.Logf(log.DebugHigh, "Unable to push down group down expression: %#v (%T)\n", e, e)

	return nil, false

}

// TranslatePredicate attempts to turn the SQLExpr into mongodb query language.
// It returns 2 things, a translated predicate that can be sent to MongoDB and
// a SQLExpr that cannot be sent to MongoDB. Either of these may be nil.
func TranslatePredicate(e SQLExpr, lookupFieldName fieldNameLookup) (bson.M, SQLExpr) {
	switch typedE := e.(type) {
	case *SQLAndExpr:
		left, exLeft := TranslatePredicate(typedE.left, lookupFieldName)
		right, exRight := TranslatePredicate(typedE.right, lookupFieldName)
		var match bson.M
		if left == nil && right == nil {
			return nil, e
		} else if left != nil && right == nil {
			match = left
		} else if left == nil && right != nil {
			match = right
		} else {
			cond := []interface{}{}
			if v, ok := left["$and"]; ok {
				array := v.([]interface{})
				cond = append(cond, array...)
			} else {
				cond = append(cond, left)
			}

			if v, ok := right["$and"]; ok {
				array := v.([]interface{})
				cond = append(cond, array...)
			} else {
				cond = append(cond, right)
			}

			match = bson.M{"$and": cond}
		}

		if exLeft == nil && exRight == nil {
			return match, nil
		} else if exLeft != nil && exRight == nil {
			return match, exLeft
		} else if exLeft == nil && exRight != nil {
			return match, exRight
		} else {
			return match, &SQLAndExpr{exLeft, exRight}
		}

	case *SQLEqualsExpr:
		name, ok := getFieldName(typedE.left, lookupFieldName)
		if !ok {
			return nil, e
		}
		val, ok := getValue(typedE.right)
		if !ok {
			return nil, e
		}
		return bson.M{name: val}, nil
	case *SQLGreaterThanExpr:
		match, ok := translateOperator("$gt", typedE.left, typedE.right, lookupFieldName)
		if !ok {
			return nil, e
		}
		return match, nil
	case *SQLGreaterThanOrEqualExpr:
		match, ok := translateOperator("$gte", typedE.left, typedE.right, lookupFieldName)
		if !ok {
			return nil, e
		}
		return match, nil
	case *SQLInExpr:
		name, ok := getFieldName(typedE.left, lookupFieldName)
		if !ok {
			return nil, e
		}

		tuple, ok := typedE.right.(*SQLTupleExpr)
		if !ok {
			return nil, e
		}

		values := []interface{}{}
		for _, valExpr := range tuple.Exprs {
			value, ok := getValue(valExpr)
			if !ok {
				return nil, e
			}

			values = append(values, value)
		}

		return bson.M{name: bson.M{"$in": values}}, nil
	case *SQLLessThanExpr:
		match, ok := translateOperator("$lt", typedE.left, typedE.right, lookupFieldName)
		if !ok {
			return nil, e
		}
		return match, nil
	case *SQLLessThanOrEqualExpr:
		match, ok := translateOperator("$lte", typedE.left, typedE.right, lookupFieldName)
		if !ok {
			return nil, e
		}
		return match, nil
	case *SQLNotEqualsExpr:
		match, ok := translateOperator("$ne", typedE.left, typedE.right, lookupFieldName)
		if !ok {
			return nil, e
		}
		return match, nil
	case *SQLNotExpr:
		match, ex := TranslatePredicate(typedE.operand, lookupFieldName)
		if match == nil {
			return nil, e
		} else if ex == nil {
			return negate(match), nil
		} else {
			// partial translation of Not
			return negate(match), &SQLNotExpr{ex}
		}

	case *SQLNullCmpExpr:
		name, ok := getFieldName(typedE.operand, lookupFieldName)
		if !ok {
			return nil, e
		}

		return bson.M{name: nil}, nil
	case *SQLOrExpr:
		left, exLeft := TranslatePredicate(typedE.left, lookupFieldName)
		if exLeft != nil {
			// cannot partially translate an OR
			return nil, e
		}
		right, exRight := TranslatePredicate(typedE.right, lookupFieldName)
		if exRight != nil {
			// cannot partially translate an OR
			return nil, e
		}

		cond := []interface{}{}

		if v, ok := left["$or"]; ok {
			array := v.([]interface{})
			cond = append(cond, array...)
		} else {
			cond = append(cond, left)
		}

		if v, ok := right["$or"]; ok {
			array := v.([]interface{})
			cond = append(cond, array...)
		} else {
			cond = append(cond, right)
		}

		return bson.M{"$or": cond}, nil
	}

	return nil, e
}

func translateOperator(op string, nameExpr SQLExpr, valExpr SQLExpr, lookupFieldName fieldNameLookup) (bson.M, bool) {
	name, ok := getFieldName(nameExpr, lookupFieldName)
	if !ok {
		return nil, false
	}
	val, ok := getValue(valExpr)
	if !ok {
		return nil, false
	}
	return bson.M{name: bson.M{op: val}}, true
}

func negate(op bson.M) bson.M {
	if len(op) == 1 {
		name, value := getSingleMapEntry(op)
		if strings.HasPrefix(name, "$") {
			switch name {
			case "$or":
				return bson.M{"$nor": value}
			case "$nor":
				return bson.M{"$or": value}
			}
		} else if innerOp, ok := value.(bson.M); ok {
			if len(innerOp) == 1 {
				innerName, innerValue := getSingleMapEntry(innerOp)
				if strings.HasPrefix(innerName, "$") {
					switch innerName {
					case "$eq":
						return bson.M{name: bson.M{"$ne": innerValue}}
					case "$in":
						return bson.M{name: bson.M{"$nin": innerValue}}
					case "$ne":
						return bson.M{name: innerValue}
					case "$nin":
						return bson.M{name: bson.M{"$in": innerValue}}
					case "$not":
						return bson.M{name: innerValue}
					}

					return bson.M{name: bson.M{"$not": bson.M{innerName: innerValue}}}
				}
			}
		} else {
			return bson.M{name: bson.M{"$ne": value}}
		}
	}

	// $not only works as a meta operator on a single operator
	// so simulate $not using $nor
	return bson.M{"$nor": []interface{}{op}}
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

func getFieldName(e SQLExpr, lookupFieldName fieldNameLookup) (string, bool) {
	switch field := e.(type) {
	case SQLFieldExpr:
		return lookupFieldName(field.tableName, field.fieldName)
	default:
		return "", false
	}
}

func getValue(e SQLExpr) (interface{}, bool) {

	cons, ok := e.(SQLValue)
	if !ok {
		return nil, false
	}

	return cons.Value(), true
}
