package planner

import (
	"errors"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"regexp"
)

var ErrUntransformableCondition = errors.New("condition can't be expressed as a field:value pair")

type MatchOperator struct {
	source  Operator
	matcher Matcher
}

func (mo *MatchOperator) Open(ctx *ExecutionCtx) error {
	return mo.source.Open(ctx)
}

func (mo *MatchOperator) Close() error {
	return mo.source.Close()
}

func (mo *MatchOperator) Next(row *Row) bool {
	for mo.source.Next(row) {
		// TODO don't allocate this repeatedly inside the loop to avoid GC?
		ctx := &MatchCtx{[]*Row{row}}
		if mo.matcher.Matches(ctx) {
			return true
		} else {
			// the row from source does not match - keep iterating
			continue
		}
	}
	return false
}

func (mo *MatchOperator) Err() error {
	return nil
}

var SQLNull = SQLNullValue{}

var ErrTypeMismatch = errors.New("type mismatch")

type MatchCtx struct {
	rows []*Row
}

// Tree nodes for evaluating if a row matches
type Matcher interface {
	Matches(*MatchCtx) bool
	Transform() (*bson.D, error)
}

type EmptyMatcher struct {
}

func (em EmptyMatcher) Matches(*MatchCtx) bool {
	return true
}

func (em EmptyMatcher) Transform() (*bson.D, error) {
	return &bson.D{}, nil
}

type rangeNodes struct {
	left, right SQLValue
}

// Matcher
type NotEquals rangeNodes
type Equals rangeNodes
type GreaterThan rangeNodes
type LessThan rangeNodes
type GreaterThanOrEqual rangeNodes
type LessThanOrEqual rangeNodes
type Like rangeNodes

func (neq *NotEquals) Matches(ctx *MatchCtx) bool {
	leftEvald := neq.left.Evaluate(ctx)
	rightEvald := neq.right.Evaluate(ctx)
	if c, err := leftEvald.CompareTo(ctx, rightEvald); err == nil {
		return c != 0
	}
	return false
}

func (eq *Equals) Matches(ctx *MatchCtx) bool {
	leftEvald := eq.left.Evaluate(ctx)
	rightEvald := eq.right.Evaluate(ctx)
	if c, err := leftEvald.CompareTo(ctx, rightEvald); err == nil {
		return c == 0
	}
	return false
}

// makeMQLQueryPair checks that one of the two provided SQLValues is a field, and the other
// is a literal that can be used in a comparison. It returns the field's SQLValue, then the
// literal's, then a bool indicating if the re-ordered pair is the inverse of the order that was
// passed in. An error is returned if both or neither of the values are SQLFields.
func makeMQLQueryPair(left, right SQLValue) (*SQLField, SQLValue, bool, error) {
	//fmt.Printf("left is %#v right is %#v\n", left, right)
	leftField, leftOk := left.(SQLField)
	rightField, rightOk := right.(SQLField)
	//fmt.Println(leftOk, rightOk)
	if leftOk == rightOk {
		return nil, nil, false, ErrUntransformableCondition
	}
	if leftOk {
		//fmt.Println("returning left field first")
		return &leftField, right, false, nil
	}
	//fmt.Println("returning right field")
	return &rightField, left, true, nil
}

func (eq *Equals) Transform() (*bson.D, error) {
	tField, tLiteral, _, err := makeMQLQueryPair(eq.left, eq.right)
	if err != nil {
		return nil, err
	}
	return &bson.D{
		{tField.fieldName, bson.D{{"$eq", tLiteral.MongoValue()}}},
	}, nil
}

func (neq *NotEquals) Transform() (*bson.D, error) {
	tField, tLiteral, _, err := makeMQLQueryPair(neq.left, neq.right)
	if err != nil {
		return nil, err
	}
	return &bson.D{
		{tField.fieldName, bson.D{{"$neq", tLiteral.MongoValue()}}},
	}, nil
}

func (gt *GreaterThan) Matches(ctx *MatchCtx) bool {
	leftEvald := gt.left.Evaluate(ctx)
	rightEvald := gt.right.Evaluate(ctx)
	if c, err := leftEvald.CompareTo(ctx, rightEvald); err == nil {
		return c > 0
	}
	return false
}

func transformComparison(left, right SQLValue, operator, inverse string) (*bson.D, error) {
	tField, tLiteral, inverted, err := makeMQLQueryPair(left, right)
	if err != nil {
		return nil, err
	}

	mongoOperator := operator
	if inverted {
		mongoOperator = inverse
	}
	return &bson.D{
		{tField.fieldName, bson.D{{mongoOperator, tLiteral.MongoValue()}}},
	}, nil
}

func (gt *GreaterThan) Transform() (*bson.D, error) {
	return transformComparison(gt.left, gt.right, "$gt", "$lte")
}

func (gt *GreaterThanOrEqual) Transform() (*bson.D, error) {
	return transformComparison(gt.left, gt.right, "$gte", "$lt")
}

func (gt *LessThanOrEqual) Transform() (*bson.D, error) {
	return transformComparison(gt.left, gt.right, "$lte", "$gt")
}

func (gt *LessThan) Transform() (*bson.D, error) {
	return transformComparison(gt.left, gt.right, "$lt", "$gte")
}

func (lt *LessThan) Matches(ctx *MatchCtx) bool {
	leftEvald := lt.left.Evaluate(ctx)
	rightEvald := lt.right.Evaluate(ctx)
	if c, err := leftEvald.CompareTo(ctx, rightEvald); err == nil {
		return c < 0
	}
	return false
}
func (gte *GreaterThanOrEqual) Matches(ctx *MatchCtx) bool {
	leftEvald := gte.left.Evaluate(ctx)
	rightEvald := gte.right.Evaluate(ctx)
	if c, err := leftEvald.CompareTo(ctx, rightEvald); err == nil {
		return c >= 0
	}
	return false
}
func (lte *LessThanOrEqual) Matches(ctx *MatchCtx) bool {
	leftEvald := lte.left.Evaluate(ctx)
	rightEvald := lte.right.Evaluate(ctx)
	if c, err := leftEvald.CompareTo(ctx, rightEvald); err == nil {
		return c <= 0
	}
	return false
}

func (l *Like) Transform() (*bson.D, error) {
	return transformComparison(l.left, l.right, "$regex", "no inverse like support")
}

func (l *Like) Matches(ctx *MatchCtx) bool {
	reg := l.left.Evaluate(ctx).MongoValue().(string)
	res, err := regexp.Match(reg, []byte(l.right.Evaluate(ctx).MongoValue().(string)))
	if err != nil {
		panic(err)
	}
	return res
}

// ----

type And struct {
	children []Matcher
}

func (and *And) Transform() (*bson.D, error) {
	transformedChildren := make([]*bson.D, 0, len(and.children))
	for _, child := range and.children {
		transformedChild, err := child.Transform()
		if err != nil {
			return nil, err
		}
		transformedChildren = append(transformedChildren, transformedChild)
	}
	return &bson.D{{"$and", transformedChildren}}, nil
}

func (and *And) Matches(ctx *MatchCtx) bool {
	for _, c := range and.children {
		if !c.Matches(ctx) {
			return false
		}
	}
	return true
}

type Or struct {
	children []Matcher
}

func (or *Or) Transform() (*bson.D, error) {
	transformedChildren := make([]*bson.D, 0, len(or.children))
	for _, child := range or.children {
		transformedChild, err := child.Transform()
		if err != nil {
			return nil, err
		}
		transformedChildren = append(transformedChildren, transformedChild)
	}
	return &bson.D{{"$or", transformedChildren}}, nil
}

func (or *Or) Matches(ctx *MatchCtx) bool {
	for _, c := range or.children {
		if c.Matches(ctx) {
			return true
		}
	}
	return false
}

func (not *Not) Transform() (*bson.D, error) {
	return nil, fmt.Errorf("transformation of 'not' expressions not supported")
}

type Not struct {
	child Matcher
}

func (not *Not) Matches(ctx *MatchCtx) bool {
	return !not.child.Matches(ctx)
}

// SQLValue used in computation by matchers
type SQLValue interface {
	Evaluate(*MatchCtx) SQLValue
	MongoValue() interface{}
	Comparable
}

type Comparable interface {
	CompareTo(*MatchCtx, SQLValue) (int, error)
}

type SQLNumeric float64
type SQLString string

func (sn SQLNumeric) Evaluate(_ *MatchCtx) SQLValue {
	return sn
}

func (ss SQLString) Evaluate(_ *MatchCtx) SQLValue {
	return ss
}

func (sn SQLNullValue) MongoValue() interface{} {
	return nil
}

func (sf SQLField) MongoValue() interface{} {
	panic("can't get the mongo value of a field reference.")
}
func (sb SQLBool) MongoValue() interface{} {
	return bool(sb)
}
func (ss SQLString) MongoValue() interface{} {
	return string(ss)
}
func (sn SQLNumeric) MongoValue() interface{} {
	return float64(sn)
}

func (sn SQLNumeric) CompareTo(ctx *MatchCtx, v SQLValue) (int, error) {
	c := v.Evaluate(ctx)
	if n, ok := c.(SQLNumeric); ok {
		return int(float64(sn) - float64(n)), nil
	}
	// can only compare numbers to each other, otherwise treat as error
	return -1, ErrTypeMismatch
}

func (sn SQLString) CompareTo(ctx *MatchCtx, v SQLValue) (int, error) {
	c := v.Evaluate(ctx)
	if n, ok := c.(SQLString); ok {
		s1, s2 := string(sn), string(n)
		if s1 < s2 {
			return -1, nil
		} else if s1 > s2 {
			return 1, nil
		}
		return 0, nil
	}
	// can only compare numbers to each other, otherwise treat as error
	return -1, ErrTypeMismatch
}

type SQLNullValue struct{}

func (nv SQLNullValue) Evaluate(ctx *MatchCtx) SQLValue {
	return nv
}

func (nv SQLNullValue) CompareTo(ctx *MatchCtx, v SQLValue) (int, error) {
	c := v.Evaluate(ctx)
	if _, ok := c.(SQLNullValue); ok {
		return 0, nil
	}
	return -1, nil
}

type SQLField struct {
	tableName string
	fieldName string
}

type SQLBool bool

func (sb SQLBool) Evaluate(ctx *MatchCtx) SQLValue {
	return sb
}

func (sb SQLBool) CompareTo(ctx *MatchCtx, v SQLValue) (int, error) {
	c := v.Evaluate(ctx)
	if n, ok := c.(SQLBool); ok {
		s1, s2 := bool(sb), bool(n)
		if s1 == s2 {
			return 0, nil
		} else if !s1 {
			return -1, nil
		}
		return 1, nil
	}
	// can only compare bool to a bool, otherwise treat as error
	return -1, ErrTypeMismatch
}

func NewSQLField(value interface{}) (SQLValue, error) {
	switch v := value.(type) {
	case nil:
		return SQLNull, nil
	case bson.ObjectId: // ObjectId
		//TODO handle this a special type? just using a string for now.
		return SQLString(v.Hex()), nil
	case bool:
		return SQLBool(v), nil
	case string:
		return SQLString(v), nil

	// TODO - handle overflow/precision of numeric types!
	case int:
		return SQLNumeric(float64(v)), nil
	case int32: // NumberInt
		return SQLNumeric(float64(v)), nil
	case float64:
		return SQLNumeric(float64(v)), nil
	case float32:
		return SQLNumeric(float64(v)), nil
	case int64: // NumberLong
		return SQLNumeric(float64(v)), nil
	default:
		panic("can't convert this type to a SQLValue")
		/*
			case *bson.M: // document
				panic("can't convert this type to a SQLValue")
			case bson.M: // document
				panic("can't convert this type to a SQLValue")
			case map[string]interface{}:
				panic("can't convert this type to a SQLValue")
			case bson.D:
				panic("can't convert this type to a SQLValue")
			case []interface{}: // array
				panic("can't convert this type to a SQLValue")
			case time.Time: // Date
				panic("can't convert this type to a SQLValue")
			case []byte: // BinData (with generic type)
				fallthrough
			case bson.Binary: // BinData
				panic("can't convert this type to a SQLValue")
			case mgo.DBRef: // DBRef
				panic("can't convert this type to a SQLValue")
			case bson.DBPointer: // DBPointer
				fallthrough
			case bson.RegEx: // RegExp
				fallthrough
			case bson.RegEx: // RegExp
				fallthrough
			case bson.MongoTimestamp: // Timestamp
				fallthrough
			case bson.JavaScript: // JavaScript
				fallthrough
		*/
	}
}

func (sqlf SQLField) Evaluate(ctx *MatchCtx) SQLValue {
	// TODO how do we report field not existing? do we just treat is a NULL, or something else?
	for _, row := range ctx.rows {
		for _, data := range row.Data {
			if data.Table == sqlf.tableName {
				if value, hasValue := row.GetField(sqlf.tableName, sqlf.fieldName); hasValue {
					val, err := NewSQLField(value)
					if err != nil {
						panic(err)
					}
					return val
				}
				// field does not exist - return null i guess
				return SQLNull
			}
		}
	}
	return SQLNull
}

func (sqlf SQLField) CompareTo(ctx *MatchCtx, v SQLValue) (int, error) {
	left := sqlf.Evaluate(ctx)
	right := v.Evaluate(ctx)
	return left.CompareTo(ctx, right)
}

// string
// number
// date
// time
