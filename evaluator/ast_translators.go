package evaluator

import (
	"github.com/10gen/sqlproxy/schema"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

func TranslatePredicate(e SQLExpr, db *schema.Database) (bson.M, bool) {
	switch typedE := e.(type) {
	case *SQLAndExpr:
		left, ok := TranslatePredicate(typedE.left, db)
		if !ok {
			return nil, false
		}
		right, ok := TranslatePredicate(typedE.right, db)
		if !ok {
			return nil, false
		}

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

		return bson.M{"$and": cond}, true
	case *SQLEqualsExpr:
		name, ok := getFieldName(typedE.left, db)
		if !ok {
			return nil, false
		}
		val, ok := getValue(typedE.right)
		if !ok {
			return nil, false
		}
		return bson.M{name: val}, true
	case *SQLGreaterThanExpr:
		return translateOperator(db, "$gt", typedE.left, typedE.right)
	case *SQLGreaterThanOrEqualExpr:
		return translateOperator(db, "$gte", typedE.left, typedE.right)
	case *SQLInExpr:
		name, ok := getFieldName(typedE.left, db)
		if !ok {
			return nil, false
		}

		tuple, ok := typedE.right.(*SQLTupleExpr)
		if !ok {
			return nil, false
		}

		values := []interface{}{}
		for _, valExpr := range tuple.Exprs {
			value, ok := getValue(valExpr)
			if !ok {
				return nil, false
			}

			values = append(values, value)
		}

		return bson.M{name: bson.M{"$in": values}}, true
	case *SQLLessThanExpr:
		return translateOperator(db, "$lt", typedE.left, typedE.right)
	case *SQLLessThanOrEqualExpr:
		return translateOperator(db, "$lte", typedE.left, typedE.right)
	case *SQLNotEqualsExpr:
		return translateOperator(db, "$ne", typedE.left, typedE.right)
	case *SQLNotExpr:
		op, ok := TranslatePredicate(typedE.operand, db)
		if !ok {
			return nil, false
		}

		return negate(op), true
	case *SQLNullCmpExpr:
		name, ok := getFieldName(typedE.operand, db)
		if !ok {
			return nil, false
		}

		return bson.M{name: nil}, true
	case *SQLOrExpr:
		left, ok := TranslatePredicate(typedE.left, db)
		if !ok {
			return nil, false
		}
		right, ok := TranslatePredicate(typedE.right, db)
		if !ok {
			return nil, false
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

		return bson.M{"$or": cond}, true
	}

	return nil, false
}

func translateOperator(db *schema.Database, op string, nameExpr SQLExpr, valExpr SQLExpr) (bson.M, bool) {
	name, ok := getFieldName(nameExpr, db)
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

func getFieldName(e SQLExpr, db *schema.Database) (string, bool) {
	if field, ok := e.(SQLFieldExpr); ok {

		tbl := db.Tables[field.tableName]

		for _, c := range tbl.Columns {
			if c.SqlName == field.fieldName {
				if c.Name == "" {
					return c.SqlName, true
				}

				return c.Name, true
			}
		}
	}

	return "", false
}

func getValue(e SQLExpr) (interface{}, bool) {

	cons, ok := e.(SQLValue)
	if !ok {
		return nil, false
	}

	return cons.Value(), true
}
