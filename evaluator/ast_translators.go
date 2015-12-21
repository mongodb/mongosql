package evaluator

import (
	"github.com/10gen/sqlproxy/schema"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

// TranslatePredicate attempts to turn the SQLExpr into mongodb query langage.
// It returns 2 things, a translated predicate that can be sent to MongoDB, and
// a SQLExpr that cannot be sent to MongoDB. Either of these may be nil.
func TranslatePredicate(e SQLExpr, db *schema.Database) (bson.M, SQLExpr) {
	switch typedE := e.(type) {
	case *SQLAndExpr:
		left, exLeft := TranslatePredicate(typedE.left, db)
		right, exRight := TranslatePredicate(typedE.right, db)
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
		name, ok := getFieldName(typedE.left, db)
		if !ok {
			return nil, e
		}
		val, ok := getValue(typedE.right)
		if !ok {
			return nil, e
		}
		return bson.M{name: val}, nil
	case *SQLGreaterThanExpr:
		match, ok := translateOperator(db, "$gt", typedE.left, typedE.right)
		if !ok {
			return nil, e
		}
		return match, nil
	case *SQLGreaterThanOrEqualExpr:
		match, ok := translateOperator(db, "$gte", typedE.left, typedE.right)
		if !ok {
			return nil, e
		}
		return match, nil
	case *SQLInExpr:
		name, ok := getFieldName(typedE.left, db)
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
		match, ok := translateOperator(db, "$lt", typedE.left, typedE.right)
		if !ok {
			return nil, e
		}
		return match, nil
	case *SQLLessThanOrEqualExpr:
		match, ok := translateOperator(db, "$lte", typedE.left, typedE.right)
		if !ok {
			return nil, e
		}
		return match, nil
	case *SQLNotEqualsExpr:
		match, ok := translateOperator(db, "$ne", typedE.left, typedE.right)
		if !ok {
			return nil, e
		}
		return match, nil
	case *SQLNotExpr:
		match, ex := TranslatePredicate(typedE.operand, db)
		if match == nil {
			return nil, e
		} else if ex == nil {
			return negate(match), nil
		} else {
			// partial translation of Not
			return negate(match), &SQLNotExpr{ex}
		}

	case *SQLNullCmpExpr:
		name, ok := getFieldName(typedE.operand, db)
		if !ok {
			return nil, e
		}

		return bson.M{name: nil}, nil
	case *SQLOrExpr:
		left, exLeft := TranslatePredicate(typedE.left, db)
		if exLeft != nil {
			// cannot partially translate an OR
			return nil, e
		}
		right, exRight := TranslatePredicate(typedE.right, db)
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

		return db.Tables[field.tableName].Columns[field.fieldName].SqlName, true
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
