package evaluator

import (
	"errors"
	"fmt"
	"github.com/10gen/sqlproxy/schema"
	"github.com/mongodb/mongo-tools/common/log"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"strings"
)

const (
	Dot           = "_DOT_"
	HavingMatcher = "_HAVING_"
	Oid           = "_id"
)

var (
	ErrPushDown = errors.New("can not push down query")
)

type fieldNameLookup func(tableName, columName string) (string, bool)

// getAggBSON creates a BSON document for the
// given aggregation expression
func getAggBSON(name string, expr interface{}) bson.M {

	if name == "count" {
		return bson.M{
			"$sum": bson.M{
				"$cond": []interface{}{
					bson.M{
						"$eq": []interface{}{
							bson.M{
								"$ifNull": []interface{}{
									expr,
									nil,
								},
							},
							nil,
						},
					},
					0,
					1,
				},
			},
		}
	}

	return bson.M{"$" + name: expr}
}

// getDottedBSON returns a BSON document to project the given
// field and a boolean indicating if it references an array
// object.
func getDottedBSON(field string) (bson.M, bool) {

	names := strings.Split(field, ".")

	if len(names) == 1 {
		return bson.M{"$first": "$" + dottifyFieldName(field)}, false
	}

	// TODO: this is to enable our current tests pass - since
	// we allow array mappings like:
	//
	// - name: loc.0
	//   sqlname: latitude
	//
	value, err := strconv.Atoi(names[len(names)-1])
	if err == nil {
		field = field[0:strings.LastIndex(field, ".")]
		return bson.M{"$arrayElemAt": []interface{}{"$" + field, value}}, true
	}

	return bson.M{"$first": "$" + dottifyFieldName(field)}, false
}

// projectGroupBy takes a prior group document together with a map of distinct aggregate
// functions and returns a projection document that can be used in the PROJECT stage of
// an aggregation pipeline.
func projectGroupBy(groupBy bson.M, distinctAggFuncs map[string]*SQLAggFunctionExpr) bson.M {

	projectClause := bson.M{}

	for key, value := range groupBy {
		if key == Oid {
			for name, _ := range value.(bson.M) {
				name = dottifyFieldName(name)
				projectClause[name] = fmt.Sprintf("$%v.%v", Oid, name)
			}
		}

		if distinctAggFunc := distinctAggFuncs[key]; distinctAggFunc != nil {
			projectClause[key] = getAggBSON(distinctAggFunc.Name, "$"+key)
		} else {
			projectClause[key] = "$" + key
		}
	}

	projectClause[Oid] = 0

	return projectClause
}

// translateGroupByKeys returns a BSON document that can be used as the _id field
// in the GROUP stage of an aggregation pipeline.
func translateGroupByKeys(exprs SelectExpressions, db *schema.Database, evaluated bool) (bson.M, error) {

	oid := bson.M{}

	// translate the group by keys
	for _, e := range exprs {

		expr := e.Expr

		switch typedE := expr.(type) {

		case SQLFieldExpr:

			name, ok := getFieldName2(typedE, db)
			if !ok {
				return nil, fmt.Errorf("could not find field name for group function")
			}

			if strings.Contains(name, ".") {
				oid[dottifyFieldName(name)], _ = getDottedBSON(name)
			} else {
				oid[name] = "$" + name
			}

		default:
			//
			// handles expressions like:
			//
			// select a + b as c from bar group by c order by c
			//
			transExpr, err := TranslateExpr(expr, db, evaluated)
			if err != nil {
				return nil, err
			}

			oid[dottifyFieldName(typedE.String())] = transExpr

		}

	}

	return oid, nil

}

// TranslateGroupBy attempts to turn the SQLExpr into mongodb query language.
// It returns a translated group by aggregation stage that can be sent to MongoDB.
func TranslateGroupBy(gb *GroupBy, db *schema.Database, table *schema.Table) (bson.M, map[string]*SQLAggFunctionExpr, error) {

	groupBy := bson.M{}
	distinctAggFuncs := make(map[string]*SQLAggFunctionExpr)

	// translate all the expressions referenced in the statement
	for _, sExpr := range gb.sExprs {

		refAggExprs, err := getAggFunctions(sExpr.Expr)
		if err != nil {
			return nil, nil, err
		}

		if len(refAggExprs) != 0 {

			for _, expr := range refAggExprs {

				if expr.Distinct {
					distinctAggFuncs[dottifyFieldName(expr.String())] = expr
					continue
				}

				transExpr, err := TranslateExpr(expr, db, gb.evaluated)
				if err != nil {
					return nil, nil, err
				}

				groupBy[dottifyFieldName(expr.String())] = transExpr

			}
		}

		// project any referenced columns in the select expression
		columns, err := referencedColumns(sExpr.Expr)
		if err != nil {
			return nil, nil, err
		}

		for _, column := range columns {

			view := dottifyFieldName(column.Name)

			name := column.Name

			if groupBy[name] != nil {
				continue
			}

			doc, isArray := getDottedBSON(table.SQLColumns[name].Name)
			if isArray {
				doc = bson.M{"$first": doc}
			}

			if len(refAggExprs) == 0 {
				groupBy[view] = doc
			}

		}

	}

	for view, distinctAggFunc := range distinctAggFuncs {
		transExpr, err := TranslateExpr(distinctAggFunc.Exprs[0], db, gb.evaluated)
		if err != nil {
			return nil, nil, err
		}

		groupBy[view] = bson.M{"$addToSet": transExpr}
	}

	oid, err := translateGroupByKeys(gb.exprs, db, gb.evaluated)
	if err != nil {
		return nil, nil, err
	}

	// remove group fields that are part of group identifier
	for k, _ := range oid {
		if groupBy[k] != nil {
			delete(groupBy, k)
		}
	}

	groupBy[Oid] = oid

	return groupBy, distinctAggFuncs, nil
}

// TranslateExpr attempts to turn the SQLExpr into MongoD query language.
func TranslateExpr(e SQLExpr, db *schema.Database, evaluated bool) (interface{}, error) {

	switch typedE := e.(type) {

	case *SQLAddExpr:

		left, err := TranslateExpr(typedE.left, db, evaluated)
		if err != nil {
			return nil, err
		}

		right, err := TranslateExpr(typedE.right, db, evaluated)
		if err != nil {
			return nil, err
		}

		return bson.M{"$add": []interface{}{left, right}}, nil

	case *SQLAggFunctionExpr:

		if evaluated {
			return "$" + dottifyFieldName(e.String()), nil
		}

		transExpr, err := TranslateExpr(typedE.Exprs[0], db, evaluated)
		if err != nil {
			return nil, err
		}

		return getAggBSON(typedE.Name, transExpr), nil

	case *SQLAndExpr:

		left, err := TranslateExpr(typedE.left, db, evaluated)
		if err != nil {
			return nil, err
		}

		right, err := TranslateExpr(typedE.right, db, evaluated)
		if err != nil {
			return nil, err
		}

		return bson.M{"$and": []interface{}{left, right}}, nil

	case *SQLCtorExpr:

		expr, err := typedE.Evaluate(nil)
		if err != nil {
			return nil, err
		}

		return TranslateExpr(expr, db, evaluated)

	case SQLDate:

		return typedE.Time, nil

	case SQLDateTime:

		return typedE.Time, nil

	case *SQLDivideExpr:

		left, err := TranslateExpr(typedE.left, db, evaluated)
		if err != nil {
			return nil, err
		}

		right, err := TranslateExpr(typedE.right, db, evaluated)
		if err != nil {
			return nil, err
		}

		return bson.M{"$divide": []interface{}{left, right}}, nil

	case *SQLEqualsExpr:

		left, err := TranslateExpr(typedE.left, db, evaluated)
		if err != nil {
			return nil, err
		}

		right, err := TranslateExpr(typedE.right, db, evaluated)
		if err != nil {
			return nil, err
		}

		return bson.M{"$eq": []interface{}{left, right}}, nil

	case SQLFieldExpr:

		name, ok := getFieldName2(typedE, db)
		if !ok {
			return nil, fmt.Errorf("invalid field name: %v", typedE)
		}

		return "$" + dottifyFieldName(name), nil

	case *SQLGreaterThanExpr:

		left, err := TranslateExpr(typedE.left, db, evaluated)
		if err != nil {
			return nil, err
		}

		right, err := TranslateExpr(typedE.right, db, evaluated)
		if err != nil {
			return nil, err
		}

		return bson.M{"$gt": []interface{}{left, right}}, nil

	case *SQLGreaterThanOrEqualExpr:

		left, err := TranslateExpr(typedE.left, db, evaluated)
		if err != nil {
			return nil, err
		}

		right, err := TranslateExpr(typedE.right, db, evaluated)
		if err != nil {
			return nil, err
		}

		return bson.M{"$gte": []interface{}{left, right}}, nil

	case *SQLLessThanExpr:

		left, err := TranslateExpr(typedE.left, db, evaluated)
		if err != nil {
			return nil, err
		}

		right, err := TranslateExpr(typedE.right, db, evaluated)
		if err != nil {
			return nil, err
		}

		return bson.M{"$lt": []interface{}{left, right}}, nil

	case *SQLLessThanOrEqualExpr:

		left, err := TranslateExpr(typedE.left, db, evaluated)
		if err != nil {
			return nil, err
		}

		right, err := TranslateExpr(typedE.right, db, evaluated)
		if err != nil {
			return nil, err
		}

		return bson.M{"$lte": []interface{}{left, right}}, nil

	case *SQLMultiplyExpr:

		left, err := TranslateExpr(typedE.left, db, evaluated)
		if err != nil {
			return nil, err
		}

		right, err := TranslateExpr(typedE.right, db, evaluated)
		if err != nil {
			return nil, err
		}

		return bson.M{"$multiply": []interface{}{left, right}}, nil

	case *SQLNotExpr:

		op, err := TranslateExpr(typedE.operand, db, evaluated)
		if err != nil {
			return nil, err
		}

		return bson.M{"$not": []interface{}{op}}, nil

	case *SQLNotEqualsExpr:

		left, err := TranslateExpr(typedE.left, db, evaluated)
		if err != nil {
			return nil, err
		}

		right, err := TranslateExpr(typedE.right, db, evaluated)
		if err != nil {
			return nil, err
		}

		return bson.M{"$ne": []interface{}{left, right}}, nil

	case *SQLNullCmpExpr:

		op, err := TranslateExpr(typedE.operand, db, evaluated)
		if err != nil {
			return nil, err
		}

		return bson.M{"$eq": []interface{}{op, nil}}, nil

	case SQLNullValue:

		return nil, nil

	case *SQLOrExpr:

		left, err := TranslateExpr(typedE.left, db, evaluated)
		if err != nil {
			return nil, err
		}

		right, err := TranslateExpr(typedE.right, db, evaluated)
		if err != nil {
			return nil, err
		}

		return bson.M{"$or": []interface{}{left, right}}, nil

	case *SQLScalarFunctionExpr:

		// TODO: can use some data and string aggregation functions here

	case *SQLSubqueryCmpExpr:

		// unsupported

	case *SQLSubtractExpr:

		left, err := TranslateExpr(typedE.left, db, evaluated)
		if err != nil {
			return nil, err
		}

		right, err := TranslateExpr(typedE.right, db, evaluated)
		if err != nil {
			return nil, err
		}

		return bson.M{"$subtract": []interface{}{left, right}}, nil

	case SQLInt, SQLUint32, SQLFloat, SQLBool, SQLString:

		return bson.M{"$literal": typedE}, nil

	case SQLTime:

		return typedE.Time, nil

	case SQLTimestamp:

		return typedE.Time, nil

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

	return nil, ErrPushDown

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

func getFieldName2(e SQLExpr, db *schema.Database) (string, bool) {

	switch field := e.(type) {

	case SQLFieldExpr:

		table := db.Tables[field.tableName]
		if table == nil {
			return "", false
		}

		column := table.SQLColumns[field.fieldName]
		if column == nil {
			return "", false
		}

		return column.Name, true

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
