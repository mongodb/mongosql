package evaluator

import (
	"encoding/hex"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/10gen/sqlproxy/schema"
	"github.com/mongodb/mongo-tools/common/log"
	"gopkg.in/mgo.v2/bson"
)

const (
	mgoOperatorEQ  = "$eq"
	mgoOperatorNEQ = "$ne"
	mgoOperatorGT  = "$gt"
	mgoOperatorGTE = "$gte"
	mgoOperatorLT  = "$lt"
	mgoOperatorLTE = "$lte"

	mgoOperatorIfNull = "$ifNull"
	mgoOperatorOR     = "$or"
	mgoOperatorCond   = "$cond"
	mgoOperatorAND    = "$and"
)

var (
	mgoNullLiteral = bson.M{"$literal": nil}
)

// a function that, given a tableName and a columnName, will return
// the field name coming back from mongodb.
type fieldNameLookup func(tableName, columnName string) (string, bool)

// TranslateExpr attempts to turn the SQLExpr into MongoDB query language.
func TranslateExpr(e SQLExpr, lookupFieldName fieldNameLookup) (interface{}, bool) {

	wrapInOp := func(op string, left, right interface{}) interface{} {
		return bson.M{op: []interface{}{left, right}}
	}

	wrapInIfNull := func(v, ifNull interface{}) interface{} {
		return bson.M{mgoOperatorIfNull: []interface{}{v, ifNull}}
	}

	wrapInNullCheck := func(v interface{}) interface{} {
		return wrapInOp(mgoOperatorEQ, wrapInIfNull(v, nil), nil)
	}

	wrapInCond := func(truePart, falsePart interface{}, conds ...interface{}) interface{} {
		var condition interface{}

		if len(conds) > 1 {
			condition = bson.M{mgoOperatorOR: conds}
		} else {
			condition = conds[0]
		}

		return bson.M{mgoOperatorCond: []interface{}{condition, truePart, falsePart}}
	}

	wrapSingleArgFuncWithNullCheck := func(name string, arg interface{}) interface{} {
		return wrapInCond(
			nil,
			bson.M{name: arg},
			wrapInNullCheck(arg),
		)
	}

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
		if !ok || transExpr == nil {
			return nil, false
		}

		dataType := typedE.Exprs[0].Type()
		if dataType == schema.SQLTimestamp || dataType == schema.SQLDate {
			return nil, false
		}

		name := typedE.Name

		switch name {
		case countAggregateName:
			if typedE.Exprs[0] == SQLVarchar("*") {
				return bson.M{"$size": transExpr}, true
			}
			// The below ensure that nulls, undefined, and missing fields
			// are not part of the count.
			return bson.M{
				"$sum": bson.M{
					"$map": bson.M{
						"input": transExpr,
						"as":    "i",
						"in": bson.M{
							mgoOperatorCond: []interface{}{
								bson.M{mgoOperatorEQ: []interface{}{
									bson.M{mgoOperatorIfNull: []interface{}{
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
		case stdAggregateName, stddevAggregateName, stddevPopAggregateName:
			return bson.M{"$stdDevPop": transExpr}, true
		case stddevSampleAggregateName:
			return bson.M{"$stdDevSamp": transExpr}, true
		default:
			return bson.M{"$" + name: transExpr}, true
		}

	case *SQLAndExpr:

		left, ok := TranslateExpr(typedE.left, lookupFieldName)
		if !ok {
			return nil, false
		}

		right, ok := TranslateExpr(typedE.right, lookupFieldName)
		if !ok {
			return nil, false
		}

		return bson.M{mgoOperatorCond: []interface{}{
			bson.M{mgoOperatorOR: []interface{}{
				bson.M{mgoOperatorEQ: []interface{}{
					bson.M{
						mgoOperatorIfNull: []interface{}{left, nil}},
					nil,
				}},
				bson.M{mgoOperatorEQ: []interface{}{
					bson.M{
						mgoOperatorIfNull: []interface{}{right, nil}},
					nil,
				}}}},
			bson.M{mgoOperatorCond: []interface{}{
				bson.M{mgoOperatorOR: []interface{}{
					bson.M{mgoOperatorEQ: []interface{}{
						left, false}},
					bson.M{mgoOperatorEQ: []interface{}{
						right, false}},
					bson.M{mgoOperatorEQ: []interface{}{
						left, 0}},
					bson.M{mgoOperatorEQ: []interface{}{
						right, 0}}}},
				bson.M{mgoOperatorAND: []interface{}{left, right}},
				mgoNullLiteral}},
			bson.M{mgoOperatorAND: []interface{}{left, right}}}}, true

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

		return wrapInCond(
			nil,
			bson.M{mgoOperatorEQ: []interface{}{left, right}},
			wrapInNullCheck(left),
			wrapInNullCheck(right),
		), true

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

		return wrapInCond(
			nil,
			bson.M{mgoOperatorGT: []interface{}{left, right}},
			wrapInNullCheck(left),
			wrapInNullCheck(right),
		), true

	case *SQLGreaterThanOrEqualExpr:
		left, ok := TranslateExpr(typedE.left, lookupFieldName)
		if !ok {
			return nil, false
		}

		right, ok := TranslateExpr(typedE.right, lookupFieldName)
		if !ok {
			return nil, false
		}

		return wrapInCond(
			nil,
			bson.M{mgoOperatorGTE: []interface{}{left, right}},
			wrapInNullCheck(left),
			wrapInNullCheck(right),
		), true

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

		return wrapInCond(
			nil,
			bson.M{mgoOperatorLT: []interface{}{left, right}},
			wrapInNullCheck(left),
			wrapInNullCheck(right),
		), true

	case *SQLLessThanOrEqualExpr:
		left, ok := TranslateExpr(typedE.left, lookupFieldName)
		if !ok {
			return nil, false
		}

		right, ok := TranslateExpr(typedE.right, lookupFieldName)
		if !ok {
			return nil, false
		}

		return wrapInCond(
			nil,
			bson.M{mgoOperatorLTE: []interface{}{left, right}},
			wrapInNullCheck(left),
			wrapInNullCheck(right),
		), true

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

		return wrapInCond(
			nil,
			bson.M{"$not": op},
			wrapInNullCheck(op),
		), true

	case *SQLNotEqualsExpr:
		left, ok := TranslateExpr(typedE.left, lookupFieldName)
		if !ok {
			return nil, false
		}

		right, ok := TranslateExpr(typedE.right, lookupFieldName)
		if !ok {
			return nil, false
		}

		return wrapInCond(
			nil,
			bson.M{mgoOperatorNEQ: []interface{}{left, right}},
			wrapInNullCheck(left),
			wrapInNullCheck(right),
		), true

	case *SQLOrExpr:
		left, ok := TranslateExpr(typedE.left, lookupFieldName)
		if !ok {
			return nil, false
		}

		right, ok := TranslateExpr(typedE.right, lookupFieldName)
		if !ok {
			return nil, false
		}

		leftIsFalse := bson.M{mgoOperatorOR: []interface{}{
			bson.M{mgoOperatorEQ: []interface{}{left, false}},
			bson.M{mgoOperatorEQ: []interface{}{left, 0}},
		}}

		leftIsTrue := bson.M{mgoOperatorOR: []interface{}{
			bson.M{mgoOperatorNEQ: []interface{}{left, false}},
			bson.M{mgoOperatorNEQ: []interface{}{left, 0}},
		}}

		rightIsFalse := bson.M{mgoOperatorOR: []interface{}{
			bson.M{mgoOperatorEQ: []interface{}{right, false}},
			bson.M{mgoOperatorEQ: []interface{}{right, 0}},
		}}

		rightIsTrue := bson.M{mgoOperatorOR: []interface{}{
			bson.M{mgoOperatorNEQ: []interface{}{right, false}},
			bson.M{mgoOperatorNEQ: []interface{}{right, 0}},
		}}

		leftIsNull := bson.M{mgoOperatorEQ: []interface{}{
			bson.M{
				mgoOperatorIfNull: []interface{}{left, nil}},
			nil,
		}}

		rightIsNull := bson.M{mgoOperatorEQ: []interface{}{
			bson.M{
				mgoOperatorIfNull: []interface{}{right, nil}},
			nil,
		}}

		nullOrFalse := bson.M{mgoOperatorOR: []interface{}{
			bson.M{mgoOperatorAND: []interface{}{
				rightIsNull, leftIsFalse,
			}},
			bson.M{mgoOperatorAND: []interface{}{
				leftIsNull, rightIsFalse,
			}},
		}}

		nullOrTrue := bson.M{mgoOperatorOR: []interface{}{
			bson.M{mgoOperatorAND: []interface{}{
				rightIsNull, leftIsTrue,
			}},
			bson.M{mgoOperatorAND: []interface{}{
				leftIsNull, rightIsTrue,
			}},
		}}

		nullOrNull := bson.M{mgoOperatorAND: []interface{}{
			leftIsNull, rightIsNull,
		}}

		return bson.M{mgoOperatorCond: []interface{}{
			nullOrNull,
			mgoNullLiteral,
			wrapInCond(
				mgoNullLiteral,
				wrapInCond(
					true,
					wrapInCond(
						mgoNullLiteral,
						bson.M{mgoOperatorOR: []interface{}{left, right}},
						wrapInNullCheck(left),
						wrapInNullCheck(right),
					),
					nullOrTrue,
				),
				nullOrFalse,
			)}}, true

	case *SQLXorExpr:
		left, ok := TranslateExpr(typedE.left, lookupFieldName)
		if !ok {
			return nil, false
		}

		right, ok := TranslateExpr(typedE.right, lookupFieldName)
		if !ok {
			return nil, false
		}

		return bson.M{mgoOperatorCond: []interface{}{
			bson.M{mgoOperatorOR: []interface{}{
				bson.M{mgoOperatorEQ: []interface{}{
					bson.M{
						mgoOperatorIfNull: []interface{}{left, nil}},
					nil,
				}},
				bson.M{mgoOperatorEQ: []interface{}{
					bson.M{
						mgoOperatorIfNull: []interface{}{right, nil}},
					nil,
				}}}},
			mgoNullLiteral,
			bson.M{mgoOperatorAND: []interface{}{
				bson.M{mgoOperatorOR: []interface{}{left, right}},
				bson.M{"$not": bson.M{mgoOperatorAND: []interface{}{left, right}}}}}}}, true

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
			var coalesce func([]interface{}) interface{}
			coalesce = func(args []interface{}) interface{} {
				if len(args) == 0 {
					return nil
				}
				replacement := coalesce(args[1:])
				return bson.M{mgoOperatorIfNull: []interface{}{args[0], replacement}}
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
					bson.M{mgoOperatorCond: []interface{}{
						bson.M{mgoOperatorEQ: []interface{}{
							bson.M{mgoOperatorIfNull: []interface{}{value, nil}},
							nil}},
						bson.M{"$literal": ""}, value}},
					bson.M{mgoOperatorCond: []interface{}{
						bson.M{mgoOperatorEQ: []interface{}{
							bson.M{mgoOperatorIfNull: []interface{}{value, nil}},
							nil}},
						bson.M{"$literal": ""}, args[0]}})
			}

			return bson.M{"$concat": pushArgs[:len(pushArgs)-1]}, true
		case "dayname":
			if len(args) != 1 {
				return nil, false
			}

			return wrapInCond(
				nil,
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
						1}}}},
				wrapInNullCheck(args[0]),
			), true
		case "day", "dayofmonth":
			if len(args) != 1 {
				return nil, false
			}

			return wrapSingleArgFuncWithNullCheck("$dayOfMonth", args[0]), true
		case "dayofweek":
			if len(args) != 1 {
				return nil, false
			}

			return wrapSingleArgFuncWithNullCheck("$dayOfWeek", args[0]), true
		case "dayofyear":
			if len(args) != 1 {
				return nil, false
			}

			return wrapSingleArgFuncWithNullCheck("$dayOfYear", args[0]), true
		case "exp":
			if len(args) != 1 {
				return nil, false
			}

			return bson.M{"$exp": args[0]}, true
		case "extract":
			if len(args) != 2 {
				return nil, false
			}

			bsonMap, ok := args[0].(bson.M)
			if !ok {
				return nil, false
			}

			bsonVal, ok := bsonMap["$literal"]
			if !ok {
				return nil, false
			}

			unitVal, ok := bsonVal.(SQLValue)
			if !ok {
				return nil, false
			}

			unit := unitVal.String()

			switch unit {
			case "year", "month", "hour", "minute", "second":
				return wrapSingleArgFuncWithNullCheck("$"+unit, args[1]), true
			case "day":
				return wrapSingleArgFuncWithNullCheck("$dayOfMonth", args[1]), true
			}

		case "floor":
			if len(args) != 1 {
				return nil, false
			}

			return bson.M{"$floor": args[0]}, true
		case "hour":
			if len(args) != 1 {
				return nil, false
			}

			return wrapSingleArgFuncWithNullCheck("$hour", args[0]), true
		case "if":
			if len(args) != 3 {
				return nil, false
			}

			return wrapInCond(
				args[2],
				args[1],
				wrapInNullCheck(args[0]),
				wrapInOp(mgoOperatorEQ, args[0], 0),
				wrapInOp(mgoOperatorEQ, args[0], false),
			), true
		case "ifnull":
			if len(args) != 2 {
				return nil, false
			}

			return wrapInIfNull(args[0], args[1]), true
		case "isnull":
			if len(args) != 1 {
				return nil, false
			}

			return wrapInCond(
				1,
				0,
				wrapInNullCheck(args[0]),
			), true
		case "left":
			if len(args) != 2 {
				return nil, false
			}

			return wrapInCond(
				nil,
				bson.M{"$substr": []interface{}{args[0], 0, args[1]}},
				wrapInNullCheck(args[0]),
				wrapInNullCheck(args[1]),
			), true
		case "lcase", "lower":
			if len(args) != 1 {
				return nil, false
			}

			return wrapSingleArgFuncWithNullCheck("$toLower", args[0]), true
		case "log", "ln":
			if len(args) != 1 {
				return nil, false
			}
			return bson.M{mgoOperatorCond: []interface{}{
				bson.M{mgoOperatorGT: []interface{}{args[0], 0}},
				bson.M{"$ln": args[0]},
				mgoNullLiteral}}, true
		case "log2":
			if len(args) != 1 {
				return nil, false
			}

			return bson.M{mgoOperatorCond: []interface{}{
				bson.M{mgoOperatorGT: []interface{}{args[0], 0}},
				bson.M{"$log": []interface{}{args[0], 2}},
				mgoNullLiteral}}, true
		case "log10":
			if len(args) != 1 {
				return nil, false
			}

			return bson.M{mgoOperatorCond: []interface{}{
				bson.M{mgoOperatorGT: []interface{}{args[0], 0}},
				bson.M{"$log10": args[0]},
				mgoNullLiteral}}, true
		case "mod":
			if len(args) != 2 {
				return nil, false
			}

			return bson.M{"$mod": []interface{}{args[0], args[1]}}, true
		case "minute":
			if len(args) != 1 {
				return nil, false
			}

			return wrapSingleArgFuncWithNullCheck("$minute", args[0]), true
		case "month":
			if len(args) != 1 {
				return nil, false
			}

			return wrapSingleArgFuncWithNullCheck("$month", args[0]), true
		case "monthname":
			if len(args) != 1 {
				return nil, false
			}

			return wrapInCond(
				nil,
				bson.M{"$arrayElemAt": []interface{}{
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
						1}}}},
				wrapInNullCheck(args[0]),
			), true
		case "nullif":
			if len(args) != 2 {
				return nil, false
			}

			return wrapInCond(
				nil,
				wrapInCond(
					nil,
					args[0],
					wrapInOp(mgoOperatorEQ, args[0], args[1]),
				),
				wrapInNullCheck(args[0]),
			), true
		case "pow", "power":
			if len(args) != 2 {
				return nil, false
			}

			return bson.M{"$pow": []interface{}{args[0], args[1]}}, true
		case "quarter":
			if len(args) != 1 {
				return nil, false
			}

			return wrapInCond(
				nil,
				bson.M{"$arrayElemAt": []interface{}{
					[]interface{}{1, 1, 1, 2, 2, 2, 3, 3, 3, 4, 4, 4},
					bson.M{"$subtract": []interface{}{
						bson.M{"$month": args[0]},
						1}}}},
				wrapInNullCheck(args[0]),
			), true
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

				placeVal, ok := bsonVal.(SQLValue)
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
				bson.M{mgoOperatorCond: []interface{}{
					bson.M{mgoOperatorGTE: []interface{}{args[0], 0}},
					bson.M{"$floor": bson.M{"$add": []interface{}{
						bson.M{"$multiply": []interface{}{args[0], decimal}}, 0.5}}},
					bson.M{"$ceil": bson.M{"$subtract": []interface{}{
						bson.M{"$multiply": []interface{}{args[0], decimal}}, 0.5}}}}}, decimal}}, true
		case "second":
			if len(args) != 1 {
				return nil, false
			}

			return wrapSingleArgFuncWithNullCheck("$second", args[0]), true
		case "sqrt":
			if len(args) != 1 {
				return nil, false
			}

			return wrapInCond(
				bson.M{"$sqrt": args[0]},
				nil,
				bson.M{mgoOperatorGTE: []interface{}{args[0], 0}},
			), true
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

			arg1Val, ok := bsonVal.(SQLValue)
			if !ok {
				return nil, false
			}

			arg1 := int(arg1Val.Float64()) - 1

			var arg2 interface{} = -1
			if len(args) == 3 {
				arg2 = args[2]
			}

			return wrapInCond(
				nil,
				bson.M{"$substr": []interface{}{args[0], arg1, arg2}},
				wrapInNullCheck(args[0]),
			), true
		case "truncate":
			if len(args) != 2 {
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

			dVal, ok := bsonVal.(SQLValue)
			if !ok {
				return nil, false
			}

			d := dVal.Float64()
			if d >= 0 {
				pow := math.Pow(10, d)
				return bson.M{"$divide": []interface{}{
					bson.M{mgoOperatorCond: []interface{}{
						bson.M{mgoOperatorGTE: []interface{}{args[0], 0}},
						bson.M{"$floor": bson.M{"$multiply": []interface{}{
							args[0], pow}}},
						bson.M{"$ceil": bson.M{"$multiply": []interface{}{
							args[0], pow}}}}},
					pow}}, true
			} else {
				pow := math.Pow(10, math.Abs(d))
				return bson.M{"$multiply": []interface{}{
					bson.M{mgoOperatorCond: []interface{}{
						bson.M{mgoOperatorGTE: []interface{}{args[0], 0}},
						bson.M{"$floor": bson.M{"$divide": []interface{}{
							args[0], pow}}},
						bson.M{"$ceil": bson.M{"$divide": []interface{}{
							args[0], pow}}}}},
					pow}}, true
			}
		case "week":
			if len(args) < 1 || len(args) > 2 {
				return nil, false
			}

			mode := 0
			if len(args) == 2 {
				bsonMap, ok := args[1].(bson.M)
				if !ok {
					return nil, false
				}

				bsonVal, ok := bsonMap["$literal"]
				if !ok {
					return nil, false
				}

				arg1Val, ok := bsonVal.(SQLValue)
				if !ok {
					return nil, false
				}

				mode = int(arg1Val.Float64())
			}

			if mode == 0 {
				return wrapSingleArgFuncWithNullCheck("$week", args[0]), true
			}
		case "weekday":
			if len(args) != 1 {
				return nil, false
			}

			return wrapInCond(
				nil,
				bson.M{"$mod": []interface{}{
					bson.M{"$add": []interface{}{
						bson.M{"$mod": []interface{}{
							bson.M{"$subtract": []interface{}{
								bson.M{"$dayOfWeek": args[0]}, 2,
							}}, 7,
						}}, 7,
					}}, 7,
				}},
				wrapInNullCheck(args[0]),
			), true
		case "ucase", "upper":
			if len(args) != 1 {
				return nil, false
			}

			return wrapSingleArgFuncWithNullCheck("$toUpper", args[0]), true
		case "year":
			if len(args) != 1 {
				return nil, false
			}

			return wrapSingleArgFuncWithNullCheck("$year", args[0]), true
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

	case SQLBool, SQLFloat, SQLInt, SQLUint32, SQLUint64, SQLVarchar:
		return bson.M{"$literal": typedE}, true

	case SQLUUID:
		value := bson.Binary{Kind: 0x03, Data: typedE.bytes}
		if typedE.kind == schema.MongoUUID {
			value.Kind = 0x04
		}
		return value, true

	case SQLNullValue:
		return mgoNullLiteral, true

	case SQLTimestamp:
		return bson.M{"$literal": typedE.Time.Format(schema.TimestampFormat)}, true

		/*
			TODO: implement these
			case *SQLUnaryTildeExpr:*/

	case *SQLCaseExpr:
		elseValue, ok := TranslateExpr(typedE.elseValue, lookupFieldName)
		if !ok {
			return nil, false
		}

		var conditions []interface{}
		var thens []interface{}
		for _, condition := range typedE.caseConditions {
			var c interface{}
			if matcher, ok := condition.matcher.(*SQLEqualsExpr); ok {
				newMatcher := &SQLOrExpr{matcher, &SQLEqualsExpr{matcher.left, SQLTrue}}
				c, ok = TranslateExpr(newMatcher, lookupFieldName)
				if !ok {
					return nil, false
				}
			} else {
				c, ok = TranslateExpr(condition.matcher, lookupFieldName)
				if !ok {
					return nil, false
				}
			}

			then, ok := TranslateExpr(condition.then, lookupFieldName)
			if !ok {
				return nil, false
			}

			conditions = append(conditions, c)
			thens = append(thens, then)
		}

		if len(conditions) != len(thens) {
			return nil, false
		}

		cases := elseValue

		for i := len(conditions) - 1; i >= 0; i-- {
			cases = wrapInCond(thens[i], cases, conditions[i])
		}

		return cases, true

	case *SQLTupleExpr:
		var transExprs []interface{}

		for _, expr := range typedE.Exprs {
			transExpr, ok := TranslateExpr(expr, lookupFieldName)
			if !ok {
				return nil, false
			}
			transExprs = append(transExprs, transExpr)
		}

		return transExprs, true

	case *SQLUnaryMinusExpr:
		operand, ok := TranslateExpr(typedE.operand, lookupFieldName)
		if !ok {
			return nil, false
		}

		return wrapInCond(
			nil,
			bson.M{"$multiply": []interface{}{-1, operand}},
			wrapInNullCheck(operand),
		), true

	case *SQLInExpr:
		left, ok := TranslateExpr(typedE.left, lookupFieldName)
		if !ok {
			return nil, false
		}

		exprs := getSQLInExprs(typedE.right)
		if exprs == nil {
			return nil, false
		}

		nullInValues := false
		var right []interface{}
		for _, expr := range exprs {
			if expr == SQLNull {
				nullInValues = true
				continue
			}
			val, ok := TranslateExpr(expr, lookupFieldName)
			if !ok {
				return nil, false
			}
			right = append(right, val)
		}

		return wrapInCond(
			nil,
			wrapInCond(
				true,
				wrapInCond(
					nil,
					false,
					bson.M{mgoOperatorEQ: []interface{}{nullInValues, true}},
				),
				bson.M{mgoOperatorGT: []interface{}{
					bson.M{"$size": bson.M{"$filter": bson.M{"input": right,
						"as":   "item",
						"cond": bson.M{mgoOperatorEQ: []interface{}{"$$item", left}},
					}}},
					bson.M{"$literal": 0},
				}}),
			wrapInNullCheck(left),
		), true

	case *SQLValues:
		var transExprs []interface{}

		for _, expr := range typedE.Values {
			transExpr, ok := TranslateExpr(expr, lookupFieldName)
			if !ok {
				return nil, false
			}
			transExprs = append(transExprs, transExpr)
		}

		return transExprs, true
	}
	log.Logf(log.DebugHigh, "Unable to push down expression: %#v (%T)\n", e, e)
	return nil, false

}

// TranslatePredicate attempts to turn the SQLExpr into mongodb query language.
// It returns 2 things, a translated predicate that can be sent to MongoDB and
// a SQLExpr that cannot be sent to MongoDB. Either of these may be nil.
func TranslatePredicate(e SQLExpr, lookupFieldName fieldNameLookup) (bson.M, SQLExpr) {

	switch typedE := e.(type) {
	case *MongoFilterExpr:
		return typedE.query, nil
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
			if v, ok := left[mgoOperatorAND]; ok {
				array := v.([]interface{})
				cond = append(cond, array...)
			} else {
				cond = append(cond, left)
			}

			if v, ok := right[mgoOperatorAND]; ok {
				array := v.([]interface{})
				cond = append(cond, array...)
			} else {
				cond = append(cond, right)
			}

			match = bson.M{mgoOperatorAND: cond}
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
		match, ok := translateOperator(mgoOperatorEQ, typedE.left, typedE.right, lookupFieldName)
		if !ok {
			return nil, e
		}
		return match, nil
	case *SQLGreaterThanExpr:
		match, ok := translateOperator(mgoOperatorGT, typedE.left, typedE.right, lookupFieldName)
		if !ok {
			return nil, e
		}
		return match, nil
	case *SQLGreaterThanOrEqualExpr:
		match, ok := translateOperator(mgoOperatorGTE, typedE.left, typedE.right, lookupFieldName)
		if !ok {
			return nil, e
		}
		return match, nil
	case *SQLInExpr:
		name, ok := getFieldName(typedE.left, lookupFieldName)
		if !ok {
			return nil, e
		}

		exprs := getSQLInExprs(typedE.right)
		if exprs == nil {
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
		match, ok := translateOperator(mgoOperatorLT, typedE.left, typedE.right, lookupFieldName)
		if !ok {
			return nil, e
		}
		return match, nil
	case *SQLLessThanOrEqualExpr:
		match, ok := translateOperator(mgoOperatorLTE, typedE.left, typedE.right, lookupFieldName)
		if !ok {
			return nil, e
		}
		return match, nil
	case *SQLLikeExpr:
		// we cannot do a like comparison on an ObjectID in mongodb.
		if typedE.left.Type() == schema.SQLObjectID {
			return nil, e
		}

		name, ok := getFieldName(typedE.left, lookupFieldName)
		if !ok {
			return nil, e
		}

		value, ok := typedE.right.(SQLValue)
		if !ok {
			return nil, e
		}

		if hasNullValue(value) {
			return nil, e
		}

		pattern := convertSQLValueToPattern(value)

		return bson.M{name: bson.D{{"$regex", pattern}, {"$options", "i"}}}, nil
	case *SQLNotEqualsExpr:
		match, ok := translateOperator(mgoOperatorNEQ, typedE.left, typedE.right, lookupFieldName)
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

		if v, ok := left[mgoOperatorOR]; ok {
			array := v.([]interface{})
			cond = append(cond, array...)
		} else {
			cond = append(cond, left)
		}

		if v, ok := right[mgoOperatorOR]; ok {
			array := v.([]interface{})
			cond = append(cond, array...)
		} else {
			cond = append(cond, right)
		}

		return bson.M{mgoOperatorOR: cond}, nil
	case *SQLRegexExpr:
		name, ok := getFieldName(typedE.operand, lookupFieldName)
		if !ok {
			return nil, e
		}

		pattern, ok := typedE.pattern.(SQLVarchar)
		if !ok {
			return nil, e
		}
		// We need to check if the pattern is valid Extended POSIX regex
		// because MongoDB supports a superset of this specification called
		// PCRE.
		_, err := regexp.CompilePOSIX(pattern.String())
		if err != nil {
			return nil, e
		}
		return bson.M{name: bson.M{"$regex": bson.RegEx{pattern.String(), ""}}}, nil
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

	if op == mgoOperatorEQ {
		translation = bson.M{name: fieldValue}
	}

	return translation, true
}

func negate(op bson.M) bson.M {
	if len(op) == 1 {
		name, value := getSingleMapEntry(op)
		if strings.HasPrefix(name, "$") {
			switch name {
			case mgoOperatorOR:
				return bson.M{"$nor": value}
			case "$nor":
				return bson.M{mgoOperatorOR: value}
			}
		} else if innerOp, ok := value.(bson.M); ok {
			if len(innerOp) == 1 {
				innerName, innerValue := getSingleMapEntry(innerOp)
				if strings.HasPrefix(innerName, "$") {
					switch innerName {
					case mgoOperatorEQ:
						return bson.M{name: bson.M{mgoOperatorNEQ: innerValue}}
					case "$in":
						return bson.M{name: bson.M{"$nin": innerValue}}
					case mgoOperatorNEQ:
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
			return bson.M{name: bson.M{mgoOperatorNEQ: value}}
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

	if cons.Type() == schema.SQLDecimal128 {
		return nil, false
	}

	return cons.Value(), true
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
