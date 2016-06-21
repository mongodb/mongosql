package evaluator

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/10gen/sqlproxy/schema"
	"github.com/mongodb/mongo-tools/common/log"
	"gopkg.in/mgo.v2/bson"
)

// a function that, given a tableName and a columnName, will return
// the field name coming back from mongodb.
type fieldNameLookup func(tableName, columnName string) (string, bool)

// TranslateExpr attempts to turn the SQLExpr into MongoDB query language.
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

		if name == "count" && typedE.Exprs[0] == SQLVarchar("*") {
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

	case SQLColumnExpr:

		name, ok := lookupFieldName(typedE.tableName, typedE.columnName)
		if !ok {
			return nil, false
		}
		return getProjectedFieldName(name, typedE.columnType.SQLType), true

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

	case *SQLIDivideExpr:

		left, ok := TranslateExpr(typedE.left, lookupFieldName)
		if !ok {
			return nil, false
		}

		right, ok := TranslateExpr(typedE.right, lookupFieldName)
		if !ok {
			return nil, false
		}

		return bson.M{"$trunc": []interface{}{
			bson.M{"$div": []interface{}{left, right}}}}, true

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

	case *SQLModExpr:

		left, ok := TranslateExpr(typedE.left, lookupFieldName)
		if !ok {
			return nil, false
		}

		right, ok := TranslateExpr(typedE.right, lookupFieldName)
		if !ok {
			return nil, false
		}

		return bson.M{"$mod": []interface{}{left, right}}, true

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

		args, ok := translateExprs(lookupFieldName, typedE.Exprs...)
		if !ok {
			return nil, false
		}

		switch typedE.Name {
		case "abs":
			if len(args) != 1 {
				return nil, false
			}
			return bson.M{"$abs": args[0]}, true
		case "coalesce":
			var coalesce func(args []interface{}) interface{}

			coalesce = func(args []interface{}) interface{} {
				if len(args) == 0 {
					return nil
				}
				replacement := coalesce(args[1:])
				return bson.M{"$ifNull": []interface{}{args[0], replacement}}
			}

			return coalesce(args), true
		case "concat":
			if len(args) < 1 {
				return nil, false
			}
			return bson.M{"$concat": args}, true
		case "concat_ws":
			if len(args) < 2 {
				return nil, false
			}

			var pushArgs []interface{}

			for _, value := range args[1:] {
				pushArgs = append(pushArgs,
					bson.M{"$cond": []interface{}{
						bson.M{"$eq": []interface{}{
							bson.M{"$ifNull": []interface{}{value, nil}},
							nil}},
						bson.M{"$literal": ""}, value}},
					bson.M{"$cond": []interface{}{
						bson.M{"$eq": []interface{}{
							bson.M{"$ifNull": []interface{}{value, nil}},
							nil}},
						bson.M{"$literal": ""}, args[0]}})
			}

			return bson.M{"$concat": pushArgs[:len(pushArgs)-1]}, true
		case "dayname":
			if len(args) != 1 {
				return nil, false
			}

			return bson.M{"$cond": []interface{}{
				bson.M{"$eq": []interface{}{
					bson.M{"$ifNull": []interface{}{args[0], nil}},
					nil}},
				bson.M{"$literal": nil},
				bson.M{"$arrayElemAt": []interface{}{
					[]interface{}{
						time.Sunday.String(),
						time.Monday.String(),
						time.Tuesday.String(),
						time.Wednesday.String(),
						time.Thursday.String(),
						time.Friday.String(),
						time.Saturday.String(),
					},
					bson.M{"$subtract": []interface{}{
						bson.M{"$dayOfWeek": args[0]},
						1}}}}}}, true
		case "day", "dayofmonth":
			if len(args) != 1 {
				return nil, false
			}

			return bson.M{"$cond": []interface{}{
				bson.M{"$eq": []interface{}{
					bson.M{"$ifNull": []interface{}{args[0], nil}},
					nil}},
				bson.M{"$literal": nil},
				bson.M{"$dayOfMonth": args[0]}}}, true
		case "dayofweek":
			if len(args) != 1 {
				return nil, false
			}

			return bson.M{"$cond": []interface{}{
				bson.M{"$eq": []interface{}{
					bson.M{"$ifNull": []interface{}{args[0], nil}},
					nil}},
				bson.M{"$literal": nil},
				bson.M{"$dayOfWeek": args[0]}}}, true
		case "dayofyear":
			if len(args) != 1 {
				return nil, false
			}

			return bson.M{"$cond": []interface{}{
				bson.M{"$eq": []interface{}{
					bson.M{"$ifNull": []interface{}{args[0], nil}},
					nil}},
				bson.M{"$literal": nil},
				bson.M{"$dayOfYear": args[0]}}}, true
		case "exp":
			if len(args) != 1 {
				return nil, false
			}

			return bson.M{"$exp": args[0]}, true
		case "floor":
			if len(args) != 1 {
				return nil, false
			}

			return bson.M{"$floor": args[0]}, true
		case "hour":
			if len(args) != 1 {
				return nil, false
			}

			return bson.M{"$cond": []interface{}{
				bson.M{"$eq": []interface{}{
					bson.M{"$ifNull": []interface{}{args[0], nil}},
					nil}},
				bson.M{"$literal": nil},
				bson.M{"$hour": args[0]}}}, true
		case "isnull":
			if len(args) != 1 {
				return nil, false
			}

			return bson.M{"$cond": []interface{}{
				bson.M{"$eq": []interface{}{
					bson.M{"$ifNull": []interface{}{args[0], nil}},
					nil}}, 1, 0}}, true
		case "left":
			if len(args) != 2 {
				return nil, false
			}

			return bson.M{"$substr": []interface{}{args[0], 0, args[1]}}, true
		case "lcase", "lower":
			if len(args) != 1 {
				return nil, false
			}
			return bson.M{"$toLower": args[0]}, true
		case "log", "ln":
			if len(args) != 1 {
				return nil, false
			}
			return bson.M{"$cond": []interface{}{
				bson.M{"$gt": []interface{}{args[0], 0}},
				bson.M{"$ln": args[0]},
				bson.M{"$literal": nil}}}, true
		case "log2":
			if len(args) != 1 {
				return nil, false
			}

			return bson.M{"$cond": []interface{}{
				bson.M{"$gt": []interface{}{args[0], 0}},
				bson.M{"$log": []interface{}{args[0], 2}},
				bson.M{"$literal": nil}}}, true
		case "log10":
			if len(args) != 1 {
				return nil, false
			}

			return bson.M{"$cond": []interface{}{
				bson.M{"$gt": []interface{}{args[0], 0}},
				bson.M{"$log10": args[0]},
				bson.M{"$literal": nil}}}, true
		case "mod":
			if len(args) != 2 {
				return nil, false
			}

			return bson.M{"$mod": []interface{}{args[0], args[1]}}, true
		case "minute":
			if len(args) != 1 {
				return nil, false
			}

			return bson.M{"$cond": []interface{}{
				bson.M{"$eq": []interface{}{
					bson.M{"$ifNull": []interface{}{args[0], nil}},
					nil}},
				bson.M{"$literal": nil},
				bson.M{"$minute": args[0]}}}, true
		case "month":
			if len(args) != 1 {
				return nil, false
			}

			return bson.M{"$month": args[0]}, true
		case "monthname":
			if len(args) != 1 {
				return nil, false
			}

			return bson.M{"$arrayElemAt": []interface{}{
				[]interface{}{
					time.January.String(),
					time.February.String(),
					time.March.String(),
					time.April.String(),
					time.May.String(),
					time.June.String(),
					time.July.String(),
					time.August.String(),
					time.September.String(),
					time.October.String(),
					time.November.String(),
					time.December.String(),
				},
				bson.M{"$subtract": []interface{}{
					bson.M{"$month": args[0]},
					1}}}}, true
		case "pow", "power":
			if len(args) != 2 {
				return nil, false
			}

			return bson.M{"$pow": []interface{}{args[0], args[1]}}, true
		case "quarter":
			if len(args) != 1 {
				return nil, false
			}

			return bson.M{"$arrayElemAt": []interface{}{
				[]interface{}{1, 1, 1, 2, 2, 2, 3, 3, 3, 4, 4, 4},
				bson.M{"$subtract": []interface{}{
					bson.M{"$month": args[0]},
					1}}}}, true
		case "round":
			if !(len(args) == 2 || len(args) == 1) {
				return nil, false
			}
			var decimal float64
			if len(args) == 2 {
				bsonMap, ok := args[1].(bson.M)
				if !ok {
					return nil, false
				}

				bsonVal, ok := bsonMap["$literal"]
				if !ok {
					return nil, false
				}

				placeVal, ok := bsonVal.(SQLNumeric)
				if !ok {
					return nil, false
				}

				places := placeVal.Float64()
				decimal = math.Pow(float64(10), places)
			} else {
				decimal = 1
			}

			if decimal < 1 {
				return bson.M{"$literal": 0}, true
			}

			return bson.M{"$divide": []interface{}{
				bson.M{"$cond": []interface{}{
					bson.M{"$gte": []interface{}{args[0], 0}},
					bson.M{"$floor": bson.M{"$add": []interface{}{
						bson.M{"$multiply": []interface{}{args[0], decimal}}, 0.5}}},
					bson.M{"$ceil": bson.M{"$subtract": []interface{}{
						bson.M{"$multiply": []interface{}{args[0], decimal}}, 0.5}}}}}, decimal}}, true
		case "second":
			if len(args) != 1 {
				return nil, false
			}

			return bson.M{"$second": args[0]}, true
		case "sqrt":
			if len(args) != 1 {
				return nil, false
			}
			return bson.M{"$cond": []interface{}{
				bson.M{"$gte": []interface{}{args[0], 0}},
				bson.M{"$sqrt": args[0]},
				bson.M{"$literal": nil}}}, true
		case "substring", "substr":
			if len(args) != 2 && len(args) != 3 {
				return nil, false
			}

			bsonMap, ok := args[1].(bson.M)
			if !ok {
				return nil, false
			}

			bsonVal, ok := bsonMap["$literal"]
			if !ok {
				return nil, false
			}

			arg1Val, ok := bsonVal.(SQLNumeric)
			if !ok {
				return nil, false
			}

			arg1 := int(arg1Val.Float64()) - 1

			var arg2 interface{} = -1
			if len(args) == 3 {
				arg2 = args[2]
			}

			return bson.M{"$substr": []interface{}{args[0], arg1, arg2}}, true
		case "week":
			// TODO: this needs to take into account the second argument
			if len(args) != 1 {
				return nil, false
			}

			return bson.M{"$cond": []interface{}{
				bson.M{"$eq": []interface{}{
					bson.M{"$ifNull": []interface{}{args[0], nil}},
					nil}},
				bson.M{"$literal": nil},
				bson.M{"$week": args[0]}}}, true
		case "ucase", "upper":
			if len(args) != 1 {
				return nil, false
			}

			return bson.M{"$toUpper": args[0]}, true
		case "year":
			if len(args) != 1 {
				return nil, false
			}

			return bson.M{"$cond": []interface{}{
				bson.M{"$eq": []interface{}{
					bson.M{"$ifNull": []interface{}{args[0], nil}},
					nil}},
				bson.M{"$literal": nil},
				bson.M{"$year": args[0]}}}, true
		}
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

	// SQL builtin types

	case SQLDate:

		return bson.M{"$literal": typedE.Time.Format(schema.DateFormat)}, true

	case SQLBool, SQLFloat, SQLInt, SQLUint32, SQLVarchar:

		return bson.M{"$literal": typedE}, true

	case SQLNullValue:

		return bson.M{"$literal": nil}, true

	case SQLTimestamp:

		return bson.M{"$literal": typedE.Time.Format(schema.TimestampFormat)}, true

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

		fieldValue, ok := getValue(typedE.right)
		if !ok {
			return nil, e
		}

		return bson.M{name: fieldValue}, nil
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

		var exprs []SQLExpr

		// The right child could be a non-SQLValues SQLValue
		// if OptimizeExpr optimizes away the tuple. For
		// example in these sorts of cases: (1) or (8-7).
		// It could be of type *SQLValues when each of the
		// expressions in the tuple are evaluated to a SQLValue.
		// Finally, it could be of type *SQLTupleExpr when
		// OptimizeExpr yielded no change.
		sqlValue, isSQLValue := typedE.right.(SQLValue)
		sqlValues, isSQLValues := typedE.right.(*SQLValues)
		sqlTupleExpr, isSQLTupleExpr := typedE.right.(*SQLTupleExpr)

		if isSQLValues {
			for _, value := range sqlValues.Values {
				exprs = append(exprs, value.(SQLExpr))
			}
		} else if isSQLValue {
			exprs = []SQLExpr{sqlValue.(SQLExpr)}
		} else if isSQLTupleExpr {
			exprs = sqlTupleExpr.Exprs
		} else {
			return nil, e
		}

		values := []interface{}{}

		for _, expr := range exprs {
			value, ok := getValue(expr)
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

func translateOperator(op string, nameExpr, valExpr SQLExpr, lookupFieldName fieldNameLookup) (bson.M, bool) {
	name, ok := getFieldName(nameExpr, lookupFieldName)
	if !ok {
		return nil, false
	}

	fieldValue, ok := getValue(valExpr)
	if !ok {
		return nil, false
	}

	return bson.M{name: bson.M{op: fieldValue}}, true
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
	case SQLColumnExpr:
		return lookupFieldName(field.tableName, field.columnName)
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

func translateExprs(lookupFieldName fieldNameLookup, exprs ...SQLExpr) ([]interface{}, bool) {
	results := []interface{}{}
	for _, e := range exprs {
		r, ok := TranslateExpr(e, lookupFieldName)
		if !ok {
			return nil, false
		}

		results = append(results, r)
	}

	return results, true
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
