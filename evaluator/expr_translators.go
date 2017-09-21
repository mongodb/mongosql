package evaluator

import (
	"encoding/hex"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/schema"
)

const (
	mgoOperatorEQ     = "$eq"
	mgoOperatorNEQ    = "$ne"
	mgoOperatorGT     = "$gt"
	mgoOperatorGTE    = "$gte"
	mgoOperatorLT     = "$lt"
	mgoOperatorLTE    = "$lte"
	mgoOperatorExists = "$exists"

	mgoOperatorIfNull = "$ifNull"
	mgoOperatorOR     = "$or"
	mgoOperatorCond   = "$cond"
	mgoOperatorAND    = "$and"
)

var (
	mgoNullLiteral = bson.M{"$literal": nil}
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

		return bson.M{"$divide": []interface{}{
			bson.M{mgoOperatorCond: []interface{}{
				bson.M{mgoOperatorGTE: []interface{}{args[0], 0}},
				bson.M{"$floor": bson.M{"$add": []interface{}{
					bson.M{"$multiply": []interface{}{args[0], decimal}}, 0.5}}},
				bson.M{"$ceil": bson.M{"$subtract": []interface{}{
					bson.M{"$multiply": []interface{}{args[0], decimal}}, 0.5}}}}}, decimal}}, true
	}

	wrapInCond = func(truePart, falsePart interface{}, conds ...interface{}) interface{} {
		var condition interface{}

		if len(conds) > 1 {
			condition = bson.M{mgoOperatorOR: conds}
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
		return wrapInOp(mgoOperatorEQ, wrapInIfNull(v, nil), nil)
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
			condition = bson.M{mgoOperatorOR: newConds}
		}

		return bson.M{mgoOperatorCond: []interface{}{condition, truePart, falsePart}}
	}

	wrapSingleArgFuncWithNullCheck = func(name string, arg interface{}) interface{} {
		return wrapInNullCheckedCond(nil, bson.M{name: arg}, arg)
	}
)

type pushDownTranslator struct {
	versionAtLeast  func(...uint8) bool
	lookupFieldName fieldNameLookup
}

// a function that, given a tableName and a columnName, will return
// the field name coming back from mongodb.
type fieldNameLookup func(tableName, columnName string) (string, bool)

// TranslateExpr attempts to turn the SQLExpr into MongoDB query language.
func (t *pushDownTranslator) TranslateExpr(e SQLExpr) (interface{}, bool) {
	switch typedE := e.(type) {

	case *SQLAddExpr:
		left, ok := t.TranslateExpr(typedE.left)
		if !ok {
			return nil, false
		}

		right, ok := t.TranslateExpr(typedE.right)
		if !ok {
			return nil, false
		}

		return bson.M{"$add": []interface{}{left, right}}, true

	case *SQLAggFunctionExpr:
		transExpr, ok := t.TranslateExpr(typedE.Exprs[0])
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

		left, ok := t.TranslateExpr(typedE.left)
		if !ok {
			return nil, false
		}

		right, ok := t.TranslateExpr(typedE.right)
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
		left, ok := t.TranslateExpr(typedE.left)
		if !ok {
			return nil, false
		}

		right, ok := t.TranslateExpr(typedE.right)
		if !ok {
			return nil, false
		}

		return wrapInCond(
			nil,
			bson.M{"$divide": []interface{}{left, right}},
			bson.M{"$eq": []interface{}{right, 0}},
		), true

	case *SQLEqualsExpr:
		left, ok := t.TranslateExpr(typedE.left)
		if !ok {
			return nil, false
		}

		right, ok := t.TranslateExpr(typedE.right)
		if !ok {
			return nil, false
		}

		return wrapInNullCheckedCond(
			nil,
			bson.M{mgoOperatorEQ: []interface{}{left, right}},
			left, right,
		), true

	case SQLColumnExpr:

		name, ok := t.lookupFieldName(typedE.tableName, typedE.columnName)
		if !ok {
			return nil, false
		}

		return getProjectedFieldName(name, typedE.columnType.SQLType), true

	case *SQLGreaterThanExpr:
		left, ok := t.TranslateExpr(typedE.left)
		if !ok {
			return nil, false
		}

		right, ok := t.TranslateExpr(typedE.right)
		if !ok {
			return nil, false
		}

		return wrapInNullCheckedCond(
			nil,
			bson.M{mgoOperatorGT: []interface{}{left, right}},
			left, right,
		), true

	case *SQLGreaterThanOrEqualExpr:
		left, ok := t.TranslateExpr(typedE.left)
		if !ok {
			return nil, false
		}

		right, ok := t.TranslateExpr(typedE.right)
		if !ok {
			return nil, false
		}

		return wrapInNullCheckedCond(
			nil,
			bson.M{mgoOperatorGTE: []interface{}{left, right}},
			left, right,
		), true

	case *SQLIDivideExpr:
		left, ok := t.TranslateExpr(typedE.left)
		if !ok {
			return nil, false
		}

		right, ok := t.TranslateExpr(typedE.right)
		if !ok {
			return nil, false
		}

		return wrapInCond(
			nil,
			bson.M{"$trunc": []interface{}{bson.M{"$divide": []interface{}{left, right}}}},
			bson.M{"$eq": []interface{}{right, 0}},
		), true

	case *SQLLessThanExpr:
		left, ok := t.TranslateExpr(typedE.left)
		if !ok {
			return nil, false
		}

		right, ok := t.TranslateExpr(typedE.right)
		if !ok {
			return nil, false
		}

		return wrapInNullCheckedCond(
			nil,
			bson.M{mgoOperatorLT: []interface{}{left, right}},
			left, right,
		), true

	case *SQLLessThanOrEqualExpr:
		left, ok := t.TranslateExpr(typedE.left)
		if !ok {
			return nil, false
		}

		right, ok := t.TranslateExpr(typedE.right)
		if !ok {
			return nil, false
		}

		return wrapInNullCheckedCond(
			nil,
			bson.M{mgoOperatorLTE: []interface{}{left, right}},
			left, right,
		), true

	case *SQLModExpr:
		left, ok := t.TranslateExpr(typedE.left)
		if !ok {
			return nil, false
		}

		right, ok := t.TranslateExpr(typedE.right)
		if !ok {
			return nil, false
		}

		return bson.M{"$mod": []interface{}{left, right}}, true

	case *SQLMultiplyExpr:
		left, ok := t.TranslateExpr(typedE.left)
		if !ok {
			return nil, false
		}

		right, ok := t.TranslateExpr(typedE.right)
		if !ok {
			return nil, false
		}

		return bson.M{"$multiply": []interface{}{left, right}}, true

	case *SQLNotExpr:
		op, ok := t.TranslateExpr(typedE.operand)
		if !ok {
			return nil, false
		}

		return wrapInNullCheckedCond(nil, bson.M{"$not": op}, op), true

	case *SQLNotEqualsExpr:
		left, ok := t.TranslateExpr(typedE.left)
		if !ok {
			return nil, false
		}

		right, ok := t.TranslateExpr(typedE.right)
		if !ok {
			return nil, false
		}

		return wrapInNullCheckedCond(
			nil,
			bson.M{mgoOperatorNEQ: []interface{}{left, right}},
			left, right,
		), true

	case *SQLNullSafeEqualsExpr:
		left, ok := t.TranslateExpr(typedE.left)
		if !ok {
			return nil, false
		}

		right, ok := t.TranslateExpr(typedE.right)
		if !ok {
			return nil, false
		}

		return bson.M{mgoOperatorEQ: []interface{}{left, right}}, true

	case *SQLOrExpr:
		left, ok := t.TranslateExpr(typedE.left)
		if !ok {
			return nil, false
		}

		right, ok := t.TranslateExpr(typedE.right)
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
					wrapInNullCheckedCond(
						mgoNullLiteral,
						bson.M{mgoOperatorOR: []interface{}{left, right}},
						left, right,
					),
					nullOrTrue,
				),
				nullOrFalse,
			)}}, true

	case *SQLXorExpr:
		left, ok := t.TranslateExpr(typedE.left)
		if !ok {
			return nil, false
		}

		right, ok := t.TranslateExpr(typedE.right)
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
		translateArgs := func() ([]interface{}, bool) {
			args := []interface{}{}
			for _, e := range typedE.Exprs {
				r, ok := t.TranslateExpr(e)
				if !ok {
					return nil, false
				}
				args = append(args, r)
			}
			return args, true
		}

		switch typedE.Name {
		case "abs":
			if len(typedE.Exprs) != 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return bson.M{"$abs": args[0]}, true
		case "adddate", "date_add", "subdate", "date_sub":
			if len(typedE.Exprs) != 3 {
				return nil, false
			}

			var date interface{}
			var ok bool
			if typedE.Name == "adddate" {
				// implementation for ADDDATE(DATE_FORMAT("..."), INTERVAL 0 SECOND)
				if f, ok := typedE.Exprs[0].(*SQLScalarFunctionExpr); ok && f.Name == "date_format" {
					if date, ok = t.translateDateFormatAsDate(f); !ok {
						date = nil
					}
				}
			}

			if date == nil {
				switch typedE.Exprs[0].Type() {
				case schema.SQLDate, schema.SQLTimestamp:
				default:
					return nil, false
				}

				if date, ok = t.TranslateExpr(typedE.Exprs[0]); !ok {
					return nil, false
				}
			}

			intervalValue, ok := typedE.Exprs[1].(SQLValue)
			if !ok {
				return nil, false
			}

			if intervalValue.Float64() == 0 {
				return date, true
			}

			unitValue, ok := typedE.Exprs[2].(SQLValue)
			if !ok {
				return nil, false
			}

			unitInterval, neg := dateArithmeticArgs(unitValue.String(), intervalValue)
			unit, interval, err := calculateInterval(unitValue.String(), unitInterval, neg)
			if err != nil {
				return nil, false
			}

			ms, err := unitIntervalToMilliseconds(unit, int64(interval))
			if err != nil {
				return nil, false
			}

			if typedE.Name == "subdate" || typedE.Name == "date_sub" {
				ms *= -1
			}

			return wrapInNullCheckedCond(
				nil,
				wrapInOp("$add", date, ms),
				date,
			), true
		case "datediff":
			if len(typedE.Exprs) != 2 {
				return nil, false
			}

			var date1, date2 interface{}
			var ok bool

			parseArgs := func(expr SQLExpr) (interface{}, bool) {
				if value, ok := expr.(SQLValue); ok {

					date, ok := strToDateTime(value.String(), false)
					if !ok {
						return nil, false
					}

					date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, schema.DefaultLocale)
					return date, true
				} else {
					exprType := expr.Type()
					if exprType == schema.SQLTimestamp || exprType == schema.SQLDate {
						date, ok := t.TranslateExpr(expr)
						if !ok {
							return nil, false
						}
						return date, true
					} else {
						return nil, false
					}
				}
			}

			if date1, ok = parseArgs(typedE.Exprs[0]); !ok {
				return nil, false
			}
			if date2, ok = parseArgs(typedE.Exprs[1]); !ok {
				return nil, false
			}

			days := wrapInOp("$divide", wrapInOp("$subtract", date1, date2), 86400000)
			bound := wrapInCond(106751, -106751, wrapInOp("$gt", days, 106751))

			return wrapInNullCheckedCond(
				nil,
				wrapInCond(
					bound,
					days,
					wrapInOp("$gt", days, 106751),
					wrapInOp("$lt", days, -106751),
				),
				date1,
				date2,
			), true

		case "ceil":
			if len(typedE.Exprs) != 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return bson.M{"$ceil": args[0]}, true
		case "char_length", "character_length":
			if !t.versionAtLeast(3, 4, 0) {
				return nil, false
			}
			if len(typedE.Exprs) != 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return wrapSingleArgFuncWithNullCheck("$strLenCP", args[0]), true
		case "coalesce":
			var coalesce func([]interface{}) interface{}
			coalesce = func(args []interface{}) interface{} {
				if len(args) == 0 {
					return nil
				}
				replacement := coalesce(args[1:])
				return bson.M{mgoOperatorIfNull: []interface{}{args[0], replacement}}
			}

			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return coalesce(args), true
		case "concat":
			if len(typedE.Exprs) < 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return bson.M{"$concat": args}, true
		case "concat_ws":
			if len(typedE.Exprs) < 2 {
				return nil, false
			}

			args, ok := translateArgs()
			if !ok {
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
		case "date_format":
			if len(typedE.Exprs) != 2 {
				return nil, false
			}

			date, ok := t.TranslateExpr(typedE.Exprs[0])
			if !ok {
				return nil, false
			}

			formatValue, ok := typedE.Exprs[1].(SQLValue)
			if !ok {
				return nil, false
			}

			mysqlFormat := formatValue.String()
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
		case "dayname":
			if len(typedE.Exprs) != 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return wrapInNullCheckedCond(
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
				args[0],
			), true
		case "day", "dayofmonth":
			if len(typedE.Exprs) != 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return wrapSingleArgFuncWithNullCheck("$dayOfMonth", args[0]), true
		case "dayofweek":
			if len(typedE.Exprs) != 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return wrapSingleArgFuncWithNullCheck("$dayOfWeek", args[0]), true
		case "dayofyear":
			if len(typedE.Exprs) != 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return wrapSingleArgFuncWithNullCheck("$dayOfYear", args[0]), true
		case "exp":
			if len(typedE.Exprs) != 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return bson.M{"$exp": args[0]}, true
		case "extract":
			if len(typedE.Exprs) != 2 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
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

			unitVal, _ := NewSQLValue(bsonVal, schema.SQLVarchar, schema.SQLNone)

			unit := unitVal.String()

			switch unit {
			case "year", "month", "hour", "minute", "second":
				return wrapSingleArgFuncWithNullCheck("$"+unit, args[1]), true
			case "day":
				return wrapSingleArgFuncWithNullCheck("$dayOfMonth", args[1]), true
			}

		case "floor":
			if len(typedE.Exprs) != 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return bson.M{"$floor": args[0]}, true
		case "greatest":
			// we can only push down if the types are similar
			for i := 1; i < len(typedE.Exprs); i++ {
				if !isSimilar(typedE.Exprs[0].Type(), typedE.Exprs[i].Type()) {
					return nil, false
				}
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return wrapInNullCheckedCond(
				nil,
				bson.M{"$max": args},
				args...,
			), true
		case "hour":
			if len(typedE.Exprs) != 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return wrapSingleArgFuncWithNullCheck("$hour", args[0]), true
		case "if":
			if len(typedE.Exprs) != 3 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
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
			if len(typedE.Exprs) != 2 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return wrapInIfNull(args[0], args[1]), true
		case "interval":
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}
			return wrapInCond(
				bson.M{"$literal": -1},
				bson.M{
					"$reduce": bson.M{
						"input":        args[1:],
						"initialValue": bson.M{"$literal": 0},
						"in": wrapInCond(
							bson.M{"$add": []interface{}{"$$value", bson.M{"$literal": 1}}},
							"$$value",
							bson.M{mgoOperatorGTE: []interface{}{args[0], "$$this"}},
						),
					},
				},
				wrapInNullCheck(args[0]),
			), true
		case "isnull":
			if len(typedE.Exprs) != 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return wrapInNullCheckedCond(1, 0, args[0]), true
		case "least":
			// we can only push down if the types are similar
			for i := 1; i < len(typedE.Exprs); i++ {
				if !isSimilar(typedE.Exprs[0].Type(), typedE.Exprs[i].Type()) {
					return nil, false
				}
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return wrapInNullCheckedCond(
				nil,
				bson.M{"$min": args},
				args...,
			), true
		case "left":
			if !t.versionAtLeast(3, 4, 0) {
				return nil, false
			}
			if len(typedE.Exprs) != 2 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return wrapInNullCheckedCond(
				nil,
				bson.M{"$substrCP": []interface{}{args[0], 0, args[1]}},
				args[0], args[1],
			), true
		case "length":
			if !t.versionAtLeast(3, 4, 0) {
				return nil, false
			}
			if len(typedE.Exprs) != 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return wrapSingleArgFuncWithNullCheck("$strLenBytes", args[0]), true
		case "lcase", "lower":
			if len(typedE.Exprs) != 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return wrapSingleArgFuncWithNullCheck("$toLower", args[0]), true
		case "locate":
			if !t.versionAtLeast(3, 4, 0) {
				return nil, false
			}
			if !(len(typedE.Exprs) == 2 || len(typedE.Exprs) == 3) {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			var indexOfCPArgs []interface{}
			if len(args) == 2 {
				indexOfCPArgs = []interface{}{args[1], args[0]}
			} else {
				indexOfCPArgs = []interface{}{args[1], args[0], wrapInOp("$subtract", args[2], 1)}
			}

			return wrapInNullCheckedCond(
				nil,
				wrapInOp("$add", bson.M{"$indexOfCP": indexOfCPArgs}, 1),
				args[1], args[0],
			), true
		case "log", "ln":
			if len(typedE.Exprs) != 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}
			return bson.M{mgoOperatorCond: []interface{}{
				bson.M{mgoOperatorGT: []interface{}{args[0], 0}},
				bson.M{"$ln": args[0]},
				mgoNullLiteral}}, true
		case "log2":
			if len(typedE.Exprs) != 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return bson.M{mgoOperatorCond: []interface{}{
				bson.M{mgoOperatorGT: []interface{}{args[0], 0}},
				bson.M{"$log": []interface{}{args[0], 2}},
				mgoNullLiteral}}, true
		case "log10":
			if len(typedE.Exprs) != 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return bson.M{mgoOperatorCond: []interface{}{
				bson.M{mgoOperatorGT: []interface{}{args[0], 0}},
				bson.M{"$log10": args[0]},
				mgoNullLiteral}}, true
		case "mod":
			if len(typedE.Exprs) != 2 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return bson.M{"$mod": []interface{}{args[0], args[1]}}, true
		case "microsecond":
			if len(typedE.Exprs) != 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return wrapInNullCheckedCond(
				nil,
				bson.M{"$multiply": []interface{}{
					bson.M{"$millisecond": args[0]}, 1000,
				}},
				args[0],
			), true

		case "minute":
			if len(typedE.Exprs) != 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return wrapSingleArgFuncWithNullCheck("$minute", args[0]), true
		case "month":
			if len(typedE.Exprs) != 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return wrapSingleArgFuncWithNullCheck("$month", args[0]), true
		case "monthname":
			if len(typedE.Exprs) != 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return wrapInNullCheckedCond(
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
				args[0],
			), true
		case "nullif":
			if len(typedE.Exprs) != 2 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return wrapInNullCheckedCond(
				nil,
				wrapInCond(
					nil,
					args[0],
					wrapInOp(mgoOperatorEQ, args[0], args[1]),
				),
				args[0],
			), true
		case "quarter":
			if len(typedE.Exprs) != 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return wrapInNullCheckedCond(
				nil,
				bson.M{"$arrayElemAt": []interface{}{
					[]interface{}{1, 1, 1, 2, 2, 2, 3, 3, 3, 4, 4, 4},
					bson.M{"$subtract": []interface{}{
						bson.M{"$month": args[0]},
						1}}}},
				args[0],
			), true
		case "repeat":
			if !t.versionAtLeast(3, 4, 0) {
				return nil, false
			}

			if len(typedE.Exprs) != 2 {
				return nil, false
			}

			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			str := args[0]

			// num must be rounded to match mysql
			num, ok := wrapInRound(args[1:])
			if !ok {
				return nil, false
			}

			// create array w/ args[1] values e.g. [0,1,2]
			rangeArr := bson.M{"$range": []interface{}{0, num, 1}}

			// create array of len arg[1], with each item being arg[0]
			mapArgs := bson.M{"input": rangeArr, "in": str}
			mapWithArgs := bson.M{"$map": mapArgs}

			// append all values of this array together
			inArg := bson.M{"$concat": []interface{}{"$$this",
				"$$value"}}
			reduceArgs := bson.M{"input": mapWithArgs, "initialValue": "", "in": inArg}

			repeat := bson.M{"$reduce": reduceArgs}

			return wrapInNullCheckedCond(nil, repeat, str, num), true

		case "right":
			if !t.versionAtLeast(3, 4, 0) {
				return nil, false
			}
			if len(typedE.Exprs) != 2 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return wrapInNullCheckedCond(
				nil,
				bson.M{"$substrCP": []interface{}{
					args[0],
					bson.M{"$max": []interface{}{
						bson.M{"$subtract": []interface{}{
							bson.M{"$strLenCP": args[0]},
							args[1],
						}},
						0,
					}},
					args[1]}},
				args[0], args[1],
			), true
		case "round":
			if !(len(typedE.Exprs) == 2 || len(typedE.Exprs) == 1) {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return wrapInRound(args)

		case "second":
			if len(typedE.Exprs) != 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return wrapSingleArgFuncWithNullCheck("$second", args[0]), true
		case "sqrt":
			if len(typedE.Exprs) != 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return wrapInCond(
				bson.M{"$sqrt": args[0]},
				nil,
				bson.M{mgoOperatorGTE: []interface{}{args[0], 0}},
			), true
		case "substring", "substr", "mid":
			if !t.versionAtLeast(3, 4, 0) {
				return nil, false
			}
			if (len(typedE.Exprs) != 2 && len(typedE.Exprs) != 3) || (len(typedE.Exprs) == 2 && typedE.Name == "mid") {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
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

			arg1Val, _ := NewSQLValue(bsonVal, schema.SQLInt, schema.SQLNone)

			arg1 := int(arg1Val.Float64()) - 1

			var arg2 interface{} = bson.M{"$strLenCP": args[0]}
			if len(args) == 3 {
				arg2 = args[2]
			}

			return wrapInNullCheckedCond(
				nil,
				bson.M{"$substrCP": []interface{}{args[0], arg1, arg2}},
				args[0],
			), true
		case "truncate":
			if len(typedE.Exprs) != 2 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
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

			dVal, _ := NewSQLValue(bsonVal, schema.SQLFloat, schema.SQLNone)

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
			}

			pow := math.Pow(10, math.Abs(d))
			return bson.M{"$multiply": []interface{}{
				bson.M{mgoOperatorCond: []interface{}{
					bson.M{mgoOperatorGTE: []interface{}{args[0], 0}},
					bson.M{"$floor": bson.M{"$divide": []interface{}{
						args[0], pow}}},
					bson.M{"$ceil": bson.M{"$divide": []interface{}{
						args[0], pow}}}}},
				pow}}, true
		case "week":
			if len(typedE.Exprs) < 1 || len(typedE.Exprs) > 2 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			mode := int64(0)
			if len(args) == 2 {
				bsonMap, ok := args[1].(bson.M)
				if !ok {
					return nil, false
				}

				bsonVal, ok := bsonMap["$literal"]
				if !ok {
					return nil, false
				}

				arg1Val, _ := NewSQLValue(bsonVal, schema.SQLInt, schema.SQLNone)
				mode = arg1Val.Int64()
			}

			if mode == 0 {
				return wrapSingleArgFuncWithNullCheck("$week", args[0]), true
			}
		case "weekday":
			if len(typedE.Exprs) != 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return wrapInNullCheckedCond(
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
				args[0],
			), true
		case "ucase", "upper":
			if len(typedE.Exprs) != 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return wrapSingleArgFuncWithNullCheck("$toUpper", args[0]), true
		case "year":
			if len(typedE.Exprs) != 1 {
				return nil, false
			}
			args, ok := translateArgs()
			if !ok {
				return nil, false
			}

			return wrapSingleArgFuncWithNullCheck("$year", args[0]), true
		}
	case *SQLSubqueryCmpExpr:
		// unsupported

	case *SQLSubtractExpr:
		left, ok := t.TranslateExpr(typedE.left)
		if !ok {
			return nil, false
		}

		right, ok := t.TranslateExpr(typedE.right)
		if !ok {
			return nil, false
		}

		return bson.M{"$subtract": []interface{}{left, right}}, true

	case *SQLCaseExpr:
		elseValue, ok := t.TranslateExpr(typedE.elseValue)
		if !ok {
			return nil, false
		}

		var conditions []interface{}
		var thens []interface{}
		for _, condition := range typedE.caseConditions {
			var c interface{}
			if matcher, ok := condition.matcher.(*SQLEqualsExpr); ok {
				newMatcher := &SQLOrExpr{matcher, &SQLEqualsExpr{matcher.left, SQLTrue}}
				c, ok = t.TranslateExpr(newMatcher)
				if !ok {
					return nil, false
				}
			} else {
				c, ok = t.TranslateExpr(condition.matcher)
				if !ok {
					return nil, false
				}
			}

			then, ok := t.TranslateExpr(condition.then)
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
			transExpr, ok := t.TranslateExpr(expr)
			if !ok {
				return nil, false
			}
			transExprs = append(transExprs, transExpr)
		}

		return transExprs, true

	case *SQLUnaryMinusExpr:
		operand, ok := t.TranslateExpr(typedE.operand)
		if !ok {
			return nil, false
		}

		return wrapInNullCheckedCond(
			nil,
			bson.M{"$multiply": []interface{}{-1, operand}},
			operand,
		), true

	case *SQLInExpr:
		left, ok := t.TranslateExpr(typedE.left)
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
			val, ok := t.TranslateExpr(expr)
			if !ok {
				return nil, false
			}
			right = append(right, val)
		}

		return wrapInNullCheckedCond(
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
			left,
		), true

	case *SQLIsExpr:
		left, ok := t.TranslateExpr(typedE.left)
		if !ok {
			return nil, false
		}

		right, ok := t.TranslateExpr(typedE.right)
		if !ok {
			return nil, false
		}

		// if right side is {null,unknown}, it's a simple case
		if typedE.right == SQLNull {
			return wrapInOp(mgoOperatorEQ,
				wrapInIfNull(left, mgoNullLiteral),
				right,
			), true
		}
		// otherwise, the right side is a boolean

		// if left side is a boolean, this is still simple
		if typedE.left.Type() == schema.SQLBoolean {
			return wrapInOp(mgoOperatorEQ,
				left,
				right,
			), true
		}

		// otherwise, left side is a number type
		if typedE.right == SQLTrue {
			return wrapInCond(
				false,
				wrapInOp(mgoOperatorNEQ,
					left,
					0,
				),
				wrapInNullCheck(left),
			), true
		} else if typedE.right == SQLFalse {
			return wrapInOp(mgoOperatorEQ,
				left,
				0,
			), true
		}

	// SQL Values
	case SQLDecimal128:
		d, ok := t.translateDecimal(typedE)
		if !ok {
			return nil, false
		}
		return bson.M{"$literal": d}, true

	case SQLDate:
		return bson.M{"$literal": typedE.Time}, true

	case SQLUint64:
		val, ok := t.getValue(typedE)
		if !ok {
			return nil, false
		}

		ui := val.(uint64)
		if ui > math.MaxInt64 {
			return nil, false
		}
		return bson.M{"$literal": val}, true

	case SQLBool:
		return bson.M{"$literal": typedE.Bool()}, true

	case SQLFloat:
		return bson.M{"$literal": typedE.Value()}, true

	case SQLInt:
		return bson.M{"$literal": typedE.Value()}, true

	case SQLUint32:
		return bson.M{"$literal": typedE.Value()}, true

	case SQLVarchar:
		return bson.M{"$literal": typedE.Value()}, true

	case SQLUUID:
		value := bson.Binary{Kind: 0x03, Data: typedE.bytes}
		if typedE.kind == schema.MongoUUID {
			value.Kind = 0x04
		}
		return value, true

	case SQLNullValue:
		return mgoNullLiteral, true

	case SQLTimestamp:
		return bson.M{"$literal": typedE.Time}, true

	case *SQLValues:
		var transExprs []interface{}

		for _, expr := range typedE.Values {
			transExpr, ok := t.TranslateExpr(expr)
			if !ok {
				return nil, false
			}
			transExprs = append(transExprs, transExpr)
		}

		return transExprs, true
	}
	return nil, false
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

// TranslatePredicate attempts to turn the SQLExpr into mongodb query language.
// It returns 2 things, a translated predicate that can be sent to MongoDB and
// a SQLExpr that cannot be sent to MongoDB. Either of these may be nil.
func (t *pushDownTranslator) TranslatePredicate(e SQLExpr) (bson.M, SQLExpr) {

	switch typedE := e.(type) {
	case *MongoFilterExpr:
		return typedE.query, nil
	case *SQLAndExpr:
		left, exLeft := t.TranslatePredicate(typedE.left)
		right, exRight := t.TranslatePredicate(typedE.right)

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
	case SQLColumnExpr:
		name, ok := t.lookupFieldName(typedE.tableName, typedE.columnName)
		if !ok {
			return nil, e
		}

		if typedE.Type() != schema.SQLBoolean {
			return bson.M{
				name: bson.M{
					mgoOperatorNEQ: nil,
				},
			}, e
		}

		return bson.M{
			mgoOperatorAND: []interface{}{
				bson.M{
					name: bson.M{
						mgoOperatorNEQ: false,
					},
				},
				bson.M{
					name: bson.M{
						mgoOperatorNEQ: nil,
					},
				},
				bson.M{
					name: bson.M{
						mgoOperatorNEQ: 0,
					},
				},
			},
		}, nil
	case *SQLEqualsExpr:
		match, ok := t.translateOperator(mgoOperatorEQ, typedE.left, typedE.right)
		if !ok {
			return nil, e
		}
		return match, nil
	case *SQLGreaterThanExpr:
		match, ok := t.translateOperator(mgoOperatorGT, typedE.left, typedE.right)
		if !ok {
			return nil, e
		}
		return match, nil
	case *SQLGreaterThanOrEqualExpr:
		match, ok := t.translateOperator(mgoOperatorGTE, typedE.left, typedE.right)
		if !ok {
			return nil, e
		}
		return match, nil
	case *SQLInExpr:
		name, ok := t.getFieldName(typedE.left)
		if !ok {
			return nil, e
		}

		exprs := getSQLInExprs(typedE.right)
		if exprs == nil {
			return nil, e
		}

		values := []interface{}{}

		for _, expr := range exprs {
			value, ok := t.getValue(expr)
			if !ok {
				return nil, e
			}
			values = append(values, value)
		}

		return bson.M{name: bson.M{"$in": values}}, nil
	case *SQLIsExpr:
		name, ok := t.getFieldName(typedE.left)
		if !ok {
			return nil, e
		}
		switch typedE.right {
		case SQLNull:
			return bson.M{name: nil}, nil
		case SQLFalse:
			if typedE.left.Type() == schema.SQLBoolean {
				return bson.M{name: false}, nil
			}
			return bson.M{name: 0}, nil
		case SQLTrue:
			if typedE.left.Type() == schema.SQLBoolean {
				return bson.M{name: true}, nil
			}
			return bson.M{
				mgoOperatorAND: []interface{}{
					bson.M{name: bson.M{mgoOperatorNEQ: 0}},
					bson.M{name: bson.M{mgoOperatorNEQ: nil}},
				},
			}, nil
		}
		return nil, e
	case *SQLLessThanExpr:
		match, ok := t.translateOperator(mgoOperatorLT, typedE.left, typedE.right)
		if !ok {
			return nil, e
		}
		return match, nil
	case *SQLLessThanOrEqualExpr:
		match, ok := t.translateOperator(mgoOperatorLTE, typedE.left, typedE.right)
		if !ok {
			return nil, e
		}
		return match, nil
	case *SQLLikeExpr:
		// we cannot do a like comparison on an ObjectID in mongodb.
		if typedE.left.Type() == schema.SQLObjectID {
			return nil, e
		}

		name, ok := t.getFieldName(typedE.left)
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

		escape, ok := typedE.escape.(SQLValue)
		if !ok {
			return nil, e
		}

		escapeSeq := []rune(escape.String())
		if len(escapeSeq) > 1 {
			return nil, e
		}

		var escapeChar rune
		if len(escapeSeq) == 1 {
			escapeChar = escapeSeq[0]
		}

		pattern := convertSQLValueToPattern(value, escapeChar)

		return bson.M{name: bson.M{"$regex": bson.RegEx{pattern, "i"}}}, nil
	case *SQLNotEqualsExpr:
		match, ok := t.translateOperator(mgoOperatorNEQ, typedE.left, typedE.right)
		if !ok {
			return nil, e
		}

		value, ok := t.getValue(typedE.right)
		if !ok {
			return nil, e
		}

		if value != nil {
			name, ok := t.getFieldName(typedE.left)
			if !ok {
				return nil, e
			}
			match = bson.M{
				mgoOperatorAND: []interface{}{match,
					bson.M{name: bson.M{mgoOperatorNEQ: nil}},
				},
			}
		}

		return match, nil
	case *SQLNotExpr:
		match, ex := t.TranslatePredicate(typedE.operand)
		if match == nil {
			return nil, e
		} else if ex == nil {
			return negate(match), nil
		} else {
			// partial translation of Not
			return negate(match), &SQLNotExpr{ex}
		}

	case *SQLOrExpr:
		left, exLeft := t.TranslatePredicate(typedE.left)
		if exLeft != nil {
			// cannot partially translate an OR
			return nil, e
		}
		right, exRight := t.TranslatePredicate(typedE.right)
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
		name, ok := t.getFieldName(typedE.operand)
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
			// Logical NOT evaluates to 1 if the operand is 0, to 0
			// if the operand is nonzero, and NOT NULL returns NULL.
			// See https://dev.mysql.com/doc/refman/5.7/en/logical-operators.html#operator_not
			// for more.
			translation := bson.M{name: bson.M{mgoOperatorNEQ: value}}
			if value != nil {
				translation = bson.M{
					mgoOperatorAND: []interface{}{
						translation,
						bson.M{name: bson.M{mgoOperatorNEQ: nil}},
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

func (t *pushDownTranslator) getFieldName(e SQLExpr) (string, bool) {
	switch field := e.(type) {
	case SQLColumnExpr:
		return t.lookupFieldName(field.tableName, field.columnName)
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
