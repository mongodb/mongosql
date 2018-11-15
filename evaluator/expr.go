package evaluator

import (
	"context"
	"fmt"
	"math"
	"regexp"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/internal/util/bsonutil"
	"github.com/10gen/sqlproxy/internal/variable"
	"github.com/10gen/sqlproxy/schema"
)

// These are the possible values for the ArithmeticOperator enum.
const (
	ADD ArithmeticOperator = iota
	DIV
	MULT
	SUB
)

const (
	// GlobalStatus is a global server status variable.
	GlobalStatus VariableKind = "global_status"
	// GlobalVariable is a global system variable.
	GlobalVariable VariableKind = "global"
	// SessionStatus is a session(local) server status variable.
	SessionStatus VariableKind = "session_status"
	// SessionVariable is a session(local) variable.
	SessionVariable VariableKind = "session"
	// UserVariable is a custom variable associated with a session(local).
	UserVariable VariableKind = "user"
)

// SQLExpr is the base type for a SQL expression.
type SQLExpr interface {
	Node
	// Evaluate evaluates the receiver expression in memory.
	Evaluate(context.Context, *ExecutionConfig, *ExecutionState) (SQLValue, error)
	// String renders a string representation of the receiver expression.
	String() string
	// EvalType returns the EvalType resulting from evaluating the expression
	// (for instance, SQLEqualsExpr.EvalType() returns EvalBoolean).
	EvalType() EvalType
}

type reconcilingSQLExpr interface {
	SQLExpr
	reconcile() (SQLExpr, error)
}

// ArithmeticOperator is a type that defines the arithmetic operators: add,
// subtract, multiply, divide.
type ArithmeticOperator byte

// MongoFilterExpr holds a MongoDB filter expression.
type MongoFilterExpr struct {
	column SQLColumnExpr
	expr   SQLExpr
	query  bson.M
}

var _ translatableToMatch = (*MongoFilterExpr)(nil)

// Evaluate evaluates a MongoFilterExpr into a SQLValue.
func (fe *MongoFilterExpr) Evaluate(context.Context, *ExecutionConfig, *ExecutionState) (SQLValue, error) {
	return nil, fmt.Errorf("could not evaluate predicate with mongo filter expression")
}

func (fe *MongoFilterExpr) String() string {
	return fmt.Sprintf("%v=%v", fe.column.String(), fe.expr.String())
}

// ToMatchLanguage translates MongoFilterExpr into something that can
// be used in an match expression. If MongoFilterExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original MongoFilterExpr.
func (fe *MongoFilterExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	return fe.query, nil
}

// EvalType returns the EvalType associated with MongoFilterExpr.
func (*MongoFilterExpr) EvalType() EvalType {
	return EvalBoolean
}

// eatChildren attempts to consume the immediate left and right children of the
// current node. Consumption consists of removal of the node and adoption of its
// children. Consumption of a node will succeed only if the consumer and consumed
// are of the same type. The result of the operation is children for the current
// node to adopt.
func eatChildren(operatorName string, left, right interface{}) []interface{} {
	eat := func(child interface{}) []interface{} {
		if bs, ok := child.(bson.M); ok {
			if descendants, ok := bs[operatorName]; ok {
				return descendants.([]interface{})
			}
		}
		return nil
	}

	newChildren := make([]interface{}, 0)
	if desc := eat(left); desc != nil {
		newChildren = append(newChildren, desc...)
	} else {
		newChildren = append(newChildren, left)
	}
	if desc := eat(right); desc != nil {
		newChildren = append(newChildren, desc...)
	} else {
		newChildren = append(newChildren, right)
	}
	return newChildren
}

// SQLAddExpr evaluates to the sum of two expressions.
type SQLAddExpr sqlBinaryNode

// Evaluate evaluates a SQLAddExpr into a SQLValue.
func (add *SQLAddExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	leftVal, err := add.left.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	rightVal, err := add.right.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(cfg.sqlValueKind, add.EvalType()), nil
	}

	return doArithmetic(leftVal, rightVal, ADD)
}

func (add *SQLAddExpr) String() string {
	return fmt.Sprintf("%v+%v", add.left, add.right)
}

// ToAggregationLanguage translates SQLAddExpr into something that can
// be used in an aggregation pipeline. If SQLAddExpr cannot be translated,
// it will return nil and error.
func (add *SQLAddExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	left, err := t.ToAggregationLanguage(add.left)
	if err != nil {
		return nil, err
	}

	right, err := t.ToAggregationLanguage(add.right)
	if err != nil {
		return nil, err
	}

	return bson.M{bsonutil.OpAdd: eatChildren(bsonutil.OpAdd, left, right)}, nil
}

// EvalType returns the EvalType associated with SQLAddExpr.
func (add *SQLAddExpr) EvalType() EvalType {
	return EvalDouble
}

// SQLAndExpr evaluates to true if and only if all its children evaluate to true.
type SQLAndExpr sqlBinaryNode

var _ reconcilingSQLExpr = (*SQLAndExpr)(nil)
var _ translatableToAggregation = (*SQLAndExpr)(nil)
var _ translatableToMatch = (*SQLAndExpr)(nil)

// NewSQLAndExpr is a constructor for SQLAndExpr.
func NewSQLAndExpr(left, right SQLExpr) *SQLAndExpr {
	return &SQLAndExpr{
		left:  left,
		right: right,
	}
}

// Evaluate evaluates a SQLAndExpr into a SQLValue.
func (and *SQLAndExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	left, err := and.left.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	right, err := and.right.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if IsFalsy(left) || IsFalsy(right) {
		return NewSQLBool(cfg.sqlValueKind, false), nil
	}

	if hasNullValue(left, right) {
		return NewSQLNull(cfg.sqlValueKind, and.EvalType()), nil
	}

	return NewSQLBool(cfg.sqlValueKind, true), nil
}

// Normalize will attempt to change SQLAndExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (and *SQLAndExpr) Normalize(kind SQLValueKind) Node {
	left, leftOk := and.left.(SQLValue)
	if leftOk && IsFalsy(left) {
		return NewSQLBool(kind, false)
	} else if leftOk && Bool(left) {
		if and.right.EvalType() == EvalBoolean {
			return and.right
		}
		return NewSQLConvertExpr(and.right, EvalBoolean)
	}

	right, rightOk := and.right.(SQLValue)
	if rightOk && IsFalsy(right) {
		return NewSQLBool(kind, false)
	} else if rightOk && Bool(right) {
		if and.left.EvalType() == EvalBoolean {
			return and.left
		}
		return NewSQLConvertExpr(and.left, EvalBoolean)
	}

	return and
}

func (and *SQLAndExpr) String() string {
	return fmt.Sprintf("%v and %v", and.left, and.right)
}

func (and *SQLAndExpr) reconcile() (SQLExpr, error) {
	left := and.left
	right := and.right

	if !isBooleanComparable(left.EvalType()) {
		left = NewSQLConvertExpr(left, EvalBoolean)
	}
	if !isBooleanComparable(right.EvalType()) {
		right = NewSQLConvertExpr(right, EvalBoolean)
	}
	return &SQLAndExpr{left, right}, nil
}

// ToAggregationLanguage translates SQLAndExpr into something that can
// be used in an aggregation pipeline. If SQLAndExpr cannot be translated,
// it will return nil and error.
func (and *SQLAndExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {

	left, err := t.ToAggregationLanguage(and.left)
	if err != nil {
		return nil, err
	}

	right, err := t.ToAggregationLanguage(and.right)
	if err != nil {
		return nil, err
	}

	letAssignment := bson.M{
		"left": left, "right": right,
	}

	letEvaluation := bson.M{
		bsonutil.OpCond: []interface{}{
			bson.M{
				bsonutil.OpOr: []interface{}{
					bson.M{
						bsonutil.OpEq: []interface{}{
							bson.M{
								bsonutil.OpIfNull: []interface{}{"$$left", nil}},
							nil,
						},
					},
					bson.M{
						bsonutil.OpEq: []interface{}{
							bson.M{
								bsonutil.OpIfNull: []interface{}{"$$right", nil}},
							nil,
						},
					},
				},
			},
			bson.M{
				bsonutil.OpCond: []interface{}{
					bson.M{
						bsonutil.OpOr: []interface{}{
							bson.M{
								bsonutil.OpEq: []interface{}{"$$left", false}},
							bson.M{
								bsonutil.OpEq: []interface{}{"$$right", false}},
							bson.M{
								bsonutil.OpEq: []interface{}{"$$left", 0}},
							bson.M{
								bsonutil.OpEq: []interface{}{"$$right", 0}},
						},
					},
					bson.M{
						bsonutil.OpAnd: []interface{}{"$$left", "$$right"},
					},
					mgoNullLiteral,
				},
			},
			bson.M{
				bsonutil.OpAnd: []interface{}{"$$left", "$$right"},
			},
		},
	}

	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil

}

// ToMatchLanguage translates SQLAndExpr into something that can
// be used in an match expression. If SQLAndExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLAndExpr.
func (and *SQLAndExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	left, exLeft := t.ToMatchLanguage(and.left)
	right, exRight := t.ToMatchLanguage(and.right)

	var match bson.M
	if left == nil && right == nil {
		return nil, and
	} else if left != nil && right == nil {
		match = left
	} else if left == nil && right != nil {
		match = right
	} else {
		cond := []interface{}{}
		if v, ok := left[bsonutil.OpAnd]; ok {
			array := v.([]interface{})
			cond = append(cond, array...)
		} else {
			cond = append(cond, left)
		}

		if v, ok := right[bsonutil.OpAnd]; ok {
			array := v.([]interface{})
			cond = append(cond, array...)
		} else {
			cond = append(cond, right)
		}

		match = bson.M{bsonutil.OpAnd: cond}
	}

	if exLeft == nil && exRight == nil {
		return match, nil
	} else if exLeft != nil && exRight == nil {
		return match, exLeft
	} else if exLeft == nil && exRight != nil {
		return match, exRight
	}
	return match, &SQLAndExpr{exLeft, exRight}
}

// EvalType returns the EvalType associated with SQLAndExpr.
func (*SQLAndExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLAssignmentExpr handles assigning a value to a variable.
type SQLAssignmentExpr struct {
	variable *SQLVariableExpr
	expr     SQLExpr
}

// NewSQLAssignmentExpr is a constructor for SQLAssignmentExpr.
func NewSQLAssignmentExpr(variable *SQLVariableExpr, expr SQLExpr) *SQLAssignmentExpr {
	return &SQLAssignmentExpr{
		variable: variable,
		expr:     expr,
	}
}

// Evaluate evaluates a SQLAssignmentExpr into a SQLValue.
func (e *SQLAssignmentExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	value, err := e.expr.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	return value, err
}

func (e *SQLAssignmentExpr) String() string {
	return fmt.Sprintf("%s := %s", e.variable.String(), e.expr.String())
}

// EvalType returns the EvalType associated with SQLAssignmentExpr.
func (e *SQLAssignmentExpr) EvalType() EvalType {
	return e.expr.EvalType()
}

// SQLBenchmarkExpr evaluates expr the number of times given by count.
// https://dev.mysql.com/doc/refman/5.5/en/information-functions.html#function_benchmark
type SQLBenchmarkExpr struct {
	count SQLExpr
	expr  SQLExpr
}

// NewSQLBenchmarkExpr is a constructor for SQLBenchmarkExpr.
func NewSQLBenchmarkExpr(count, expr SQLExpr) *SQLBenchmarkExpr {
	return &SQLBenchmarkExpr{
		count: count,
		expr:  expr,
	}
}

// SkipConstantFolding indicates that we should not attempt to
// constant-fold this expression.
func (e SQLBenchmarkExpr) SkipConstantFolding() bool {
	return true
}

// Evaluate evaluates a SQLBenchmarkExpr into a SQLValue.
func (e SQLBenchmarkExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {

	count, err := e.count.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	for i := int64(0); i < Int64(count); i++ {
		_, err := e.expr.Evaluate(ctx, cfg, st)
		if err != nil {
			return nil, err
		}
	}

	return NewSQLInt64(cfg.sqlValueKind, 0), nil
}

func (e SQLBenchmarkExpr) String() string {
	return fmt.Sprintf("benchmark(%s, %s)", e.count.String(), e.expr.String())
}

// EvalType returns the EvalType associated with SQLBenchmarkExpr.
func (e SQLBenchmarkExpr) EvalType() EvalType {
	return EvalInt64
}

// SQLCaseExpr holds a number of cases to evaluate as well as the value
// to return if any of the cases is matched. If none is matched,
// 'elseValue' is evaluated and returned.
type SQLCaseExpr struct {
	elseValue      SQLExpr
	caseConditions []caseCondition
}

var _ translatableToAggregation = (*SQLCaseExpr)(nil)

// Evaluate evaluates a SQLCaseExpr into a SQLValue.
func (e SQLCaseExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	for _, condition := range e.caseConditions {
		result, err := condition.matcher.Evaluate(ctx, cfg, st)
		if err != nil {
			return nil, err
		}

		if Bool(result) {
			return condition.then.Evaluate(ctx, cfg, st)
		}
	}

	return e.elseValue.Evaluate(ctx, cfg, st)
}

func (e SQLCaseExpr) String() string {
	str := fmt.Sprintf("case ")
	for _, cond := range e.caseConditions {
		str += fmt.Sprintf("%v ", cond.String())
	}
	if e.elseValue != nil {
		str += fmt.Sprintf("else %v ", e.elseValue.String())
	}
	str += fmt.Sprintf("end")
	return str
}

// ToAggregationLanguage translates SQLCaseExpr into something that can
// be used in an aggregation pipeline. If SQLCaseExpr cannot be translated,
// it will return nil and error.
func (e *SQLCaseExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	elseValue, err := t.ToAggregationLanguage(e.elseValue)
	if err != nil {
		return nil, err
	}

	var conditions []interface{}
	var thens []interface{}
	for _, condition := range e.caseConditions {
		var c interface{}
		if matcher, ok := condition.matcher.(*SQLEqualsExpr); ok {
			newMatcher := &SQLOrExpr{
				matcher,
				&SQLEqualsExpr{matcher.left, NewSQLBool(t.valueKind(), true)}}
			c, err = t.ToAggregationLanguage(newMatcher)
			if err != nil {
				return nil, err
			}
		} else {
			c, err = t.ToAggregationLanguage(condition.matcher)
			if err != nil {
				return nil, err
			}
		}

		then, err := t.ToAggregationLanguage(condition.then)
		if err != nil {
			return nil, err
		}

		conditions = append(conditions, c)
		thens = append(thens, then)
	}

	if len(conditions) != len(thens) {
		return nil, fmt.Errorf("could not translate %v to aggregation language", e)
	}

	cases := elseValue

	for i := len(conditions) - 1; i >= 0; i-- {
		cases = bsonutil.WrapInCond(thens[i], cases, conditions[i])
	}

	return cases, nil

}

// EvalType returns the EvalType associated with SQLCaseExpr.
func (e SQLCaseExpr) EvalType() EvalType {
	conds := []SQLExpr{e.elseValue}
	for _, cond := range e.caseConditions {
		conds = append(conds, cond.then)
	}
	return preferentialType(conds...)
}

// SQLColumnExpr represents a column reference.
type SQLColumnExpr struct {
	selectID     int
	databaseName string
	tableName    string
	columnName   string
	columnType   ColumnType
	correlated   bool
}

var _ translatableToAggregation = (*SQLColumnExpr)(nil)
var _ translatableToMatch = (*SQLColumnExpr)(nil)

// Evaluate evaluates a SQLColumnExpr into a SQLValue.
func (c SQLColumnExpr) Evaluate(_ context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {

	for _, row := range st.rows {
		if value, ok := row.GetField(c.selectID, c.databaseName, c.tableName, c.columnName); ok {
			return ConvertTo(value, c.EvalType()), nil
		}
	}

	for _, row := range st.correlatedRows {
		if value, ok := row.GetField(c.selectID, c.databaseName, c.tableName, c.columnName); ok {
			return ConvertTo(value, c.EvalType()), nil
		}
	}

	// TODO BI-1883
	return NewSQLNull(cfg.sqlValueKind, c.EvalType()), nil
}

func (c SQLColumnExpr) String() string {
	var str string
	if c.databaseName != "" {
		str += c.databaseName + "."
	}

	if c.tableName != "" {
		str += c.tableName + "."
	}
	str += c.columnName
	return str
}

// ToAggregationLanguage translates SQLColumnExpr into something that can
// be used in an aggregation pipeline. If SQLColumnExpr cannot be translated,
// it will return nil and error.
func (c SQLColumnExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	if c.correlated {
		cc := t.addCorrelatedSubqueryColumnFuture(&c)
		return bsonutil.WrapInLiteral(cc), nil
	}

	name, ok := t.LookupFieldName(c.databaseName, c.tableName, c.columnName)
	if !ok {
		return nil, fmt.Errorf("cannot translate %v to aggregation language", c)
	}

	return getProjectedFieldName(name, c.columnType.EvalType), nil
}

// ToMatchLanguage translates SQLColumnExpr into something that can
// be used in an match expression. If SQLColumnExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLColumnExpr.
func (c SQLColumnExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	if c.correlated {
		cc := t.addCorrelatedSubqueryColumnFuture(&c)
		return bsonutil.WrapInLiteral(cc), nil
	}
	name, ok := t.LookupFieldName(c.databaseName, c.tableName, c.columnName)
	if !ok {
		return nil, c
	}

	if c.EvalType() != EvalBoolean {
		return bson.M{
			name: bson.M{
				bsonutil.OpNeq: nil,
			},
		}, c
	}

	return bson.M{
		bsonutil.OpAnd: []interface{}{
			bson.M{
				name: bson.M{
					bsonutil.OpNeq: false,
				},
			},
			bson.M{
				name: bson.M{
					bsonutil.OpNeq: nil,
				},
			},
			bson.M{
				name: bson.M{
					bsonutil.OpNeq: 0,
				},
			},
		},
	}, nil
}

// EvalType returns the EvalType associated with SQLColumnExpr.
func (c SQLColumnExpr) EvalType() EvalType {
	return c.columnType.EvalType
}

func (c SQLColumnExpr) isAggregateReplacementColumn() bool {
	return c.tableName == ""
}

// SQLConvertExpr represents a conversion
// of the expression expr to the target
// EvalType.
type SQLConvertExpr struct {
	expr       SQLExpr
	targetType EvalType
}

var _ translatableToAggregation = (*SQLConvertExpr)(nil)

// NewSQLConvertExpr is a constructor for SQLConvertExpr.
func NewSQLConvertExpr(expr SQLExpr, targetType EvalType) *SQLConvertExpr {
	return &SQLConvertExpr{
		expr:       expr,
		targetType: targetType,
	}
}

// Evaluate evaluates a SQLConvertExpr into a SQLValue.
func (ce *SQLConvertExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	// collapse nested SQLConvertExprs
	if sce, ok := ce.expr.(*SQLConvertExpr); ok {
		ce.expr = sce.expr
	}

	v, err := ce.expr.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	return ConvertTo(v, ce.targetType), nil
}

func (ce *SQLConvertExpr) String() string {
	prettyTypeName := string(EvalTypeToSQLType(ce.targetType))
	return "Convert(" + ce.expr.String() + ", " + prettyTypeName + ")"
}

// EvalType returns the EvalType associated with SQLConvertExpr.
func (ce *SQLConvertExpr) EvalType() EvalType {
	return ce.targetType
}

func (ce *SQLConvertExpr) translateMongoSQL(t *PushdownTranslator) (interface{}, error) {
	if !t.versionAtLeast(4, 0, 0) {
		return nil, fmt.Errorf("mongosql mode convert cannot be pushed" +
			" down on MongoDB versions < 4.0")
	}

	expr, err := t.ToAggregationLanguage(ce.expr)
	if err != nil {
		return nil, err
	}

	converted := translateConvert(expr, ce.expr.EvalType(), ce.targetType)
	return converted, nil
}

func (ce *SQLConvertExpr) translateMySQL(t *PushdownTranslator) (interface{}, error) {
	//
	// The following type conversions are pushed down:
	//
	//     any numeric type -> any numeric type
	//     any numeric type -> string
	//     any numeric type -> bool
	//     datetime         -> date
	//     datetime         -> string
	//     datetime         -> any numeric type
	//     date             -> datetime
	//     date             -> string
	//     date             -> any numeric type
	//     objectid         -> string
	//

	if !t.versionAtLeast(3, 6, 0) {
		return nil, fmt.Errorf("mysql mode convert cannot be pushed" +
			" down on MongoDB versions < 3.6")
	}

	fromType := ce.expr.EvalType()
	toType := ce.targetType

	expr, err := t.ToAggregationLanguage(ce.expr)
	if err != nil {
		return nil, err
	}

	switch fromType {
	case EvalInt32, EvalInt64,
		EvalUint32, EvalUint64,
		EvalDecimal128, EvalBoolean:

		switch toType {
		case EvalInt32, EvalInt64,
			EvalUint32, EvalUint64,
			EvalDecimal128, EvalDouble,
			EvalString, EvalBoolean:
			return ce.translateMongoSQL(t)
		}

	case EvalDouble:
		switch toType {
		case EvalInt32, EvalInt64,
			EvalUint32, EvalUint64,
			EvalDecimal128, EvalBoolean:
			return ce.translateMongoSQL(t)
		}

	case EvalDatetime:
		year := bson.M{"$year": expr}
		month := bson.M{"$month": expr}
		day := bson.M{"$dayOfMonth": expr}
		hour := bson.M{"$hour": expr}
		minute := bson.M{"$minute": expr}
		second := bson.M{"$second": expr}
		millisecond := bson.M{"$millisecond": expr}

		switch toType {
		case EvalDate:
			asDate := bson.M{"$dateFromParts": bson.M{
				"year":  year,
				"month": month,
				"day":   day,
			}}
			return asDate, nil

		case EvalInt32, EvalInt64, EvalUint32, EvalUint64:
			asNum := bson.M{"$add": []interface{}{
				second,
				bson.M{"$multiply": []interface{}{minute, 100}},
				bson.M{"$multiply": []interface{}{hour, 10000}},
				bson.M{"$multiply": []interface{}{day, 1000000}},
				bson.M{"$multiply": []interface{}{month, 100000000}},
				bson.M{"$multiply": []interface{}{year, 10000000000}},
			}}
			return asNum, nil

		case EvalDecimal128, EvalDouble:
			asNum := bson.M{"$add": []interface{}{
				bson.M{"$divide": []interface{}{millisecond, 1000}},
				second,
				bson.M{"$multiply": []interface{}{minute, 100}},
				bson.M{"$multiply": []interface{}{hour, 10000}},
				bson.M{"$multiply": []interface{}{day, 1000000}},
				bson.M{"$multiply": []interface{}{month, 100000000}},
				bson.M{"$multiply": []interface{}{year, 10000000000}},
			}}
			return asNum, nil

		case EvalString:
			asString := bson.M{"$dateToString": bson.M{
				"date":   expr,
				"format": "%Y-%m-%d %H:%M:%S.%L000",
			}}
			return asString, nil

		}

	case EvalDate:
		year := bson.M{"$year": expr}
		month := bson.M{"$month": expr}
		day := bson.M{"$dayOfMonth": expr}

		switch toType {
		case EvalDatetime:
			asDate := bson.M{"$dateFromParts": bson.M{
				"year":  year,
				"month": month,
				"day":   day,
			}}
			return asDate, nil

		case EvalInt32, EvalInt64,
			EvalUint32, EvalUint64,
			EvalDecimal128, EvalDouble:

			asNum := bson.M{"$add": []interface{}{
				day,
				bson.M{"$multiply": []interface{}{month, 100}},
				bson.M{"$multiply": []interface{}{year, 10000}},
			}}
			return asNum, nil

		case EvalString:
			asString := bson.M{"$dateToString": bson.M{
				"date":   expr,
				"format": "%Y-%m-%d",
			}}
			return asString, nil

		}

	case EvalObjectID:
		switch toType {
		case EvalString:
			return ce.translateMongoSQL(t)
		}

	default:
		// mysql-mode pushdown not yet implemented for conversions from other types
	}

	return nil, fmt.Errorf("mysql conversion cannot be pushdown with from type '%s'",
		EvalTypeToMongoType(fromType))
}

// ToAggregationLanguage translates SQLConvertExpr into something that can
// be used in an aggregation pipeline. At the moment, SQLConvertExpr cannot be
// translated, so this function will always return nil and error.
func (ce *SQLConvertExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	if ce.targetType == EvalObjectID {
		sv, ok := ce.expr.(SQLVarchar)
		if !ok {
			return nil, fmt.Errorf("can only push down SQLVarchar as ObjectId")
		}
		return sv.SQLObjectID().(translatableToAggregation).ToAggregationLanguage(t)
	}

	mode := t.Cfg.sqlValueKind
	switch mode {
	case MySQLValueKind:
		return ce.translateMySQL(t)
	case MongoSQLValueKind:
		return ce.translateMongoSQL(t)
	default:
		panic(fmt.Errorf("impossible value %v for cfg.sqlValueKind", mode))
	}
}

// SQLDivideExpr evaluates to the quotient of the left expression divided by the right.
type SQLDivideExpr sqlBinaryNode

var _ translatableToAggregation = (*SQLDivideExpr)(nil)

// Evaluate evaluates a SQLDivideExpr into a SQLValue.
func (div *SQLDivideExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	leftVal, err := div.left.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	rightVal, err := div.right.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	if Float64(rightVal) == 0 || hasNullValue(leftVal, rightVal) {
		return NewSQLNull(cfg.sqlValueKind, div.EvalType()), nil
	}

	return doArithmetic(leftVal, rightVal, DIV)
}

// Normalize will attempt to change SQLDivideExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (div *SQLDivideExpr) Normalize(kind SQLValueKind) Node {
	if hasNullExpr(div.left, div.right) {
		return NewSQLNull(kind, div.EvalType())
	}
	return div
}

func (div *SQLDivideExpr) String() string {
	return fmt.Sprintf("%v/%v", div.left, div.right)
}

// ToAggregationLanguage translates SQLDivideExpr into something that can
// be used in an aggregation pipeline. If SQLDivideExpr cannot be translated,
// it will return nil and error.
func (div *SQLDivideExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	left, err := t.ToAggregationLanguage(div.left)
	if err != nil {
		return nil, err
	}

	right, err := t.ToAggregationLanguage(div.right)
	if err != nil {
		return nil, err
	}

	letAssignment := bson.M{
		"left": left, "right": right,
	}

	letEvaluation := bsonutil.WrapInCond(
		nil,
		bson.M{bsonutil.OpDivide: []interface{}{"$$left", "$$right"}},
		bson.M{bsonutil.OpEq: []interface{}{"$$right", 0}},
	)

	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil

}

// EvalType returns the EvalType associated with SQLDivideExpr.
func (div *SQLDivideExpr) EvalType() EvalType {
	return EvalDouble
}

// SQLEqualsExpr evaluates to true if the left equals the right.
type SQLEqualsExpr sqlBinaryNode

var _ reconcilingSQLExpr = (*SQLEqualsExpr)(nil)
var _ translatableToAggregation = (*SQLEqualsExpr)(nil)
var _ translatableToMatch = (*SQLEqualsExpr)(nil)

// NewSQLEqualsExpr is a constructor for SQLEqualsExpr.
func NewSQLEqualsExpr(left, right SQLExpr) *SQLEqualsExpr {
	return &SQLEqualsExpr{
		left:  left,
		right: right,
	}
}

// Evaluate evaluates a SQLEqualsExpr into a SQLValue.
func (eq *SQLEqualsExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	leftVal, err := eq.left.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	rightVal, err := eq.right.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(cfg.sqlValueKind, eq.EvalType()), nil
	}

	c, err := CompareTo(leftVal, rightVal, st.collation)
	if err == nil {
		return NewSQLBool(cfg.sqlValueKind, c == 0), nil
	}

	return NewSQLBool(cfg.sqlValueKind, false), err
}

// Normalize will attempt to change SQLEqualsExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (eq *SQLEqualsExpr) Normalize(kind SQLValueKind) Node {
	if hasNullExpr(eq.left, eq.right) {
		return NewSQLNull(kind, eq.EvalType())
	}

	if shouldFlip(sqlBinaryNode(*eq)) {
		return &SQLEqualsExpr{eq.right, eq.left}
	}

	return eq
}

func (eq *SQLEqualsExpr) String() string {
	return fmt.Sprintf("%v = %v", eq.left, eq.right)
}

// ToAggregationLanguage translates SQLEqualsExpr into something that can
// be used in an aggregation pipeline. If SQLEqualsExpr cannot be translated,
// it will return nil and error.
func (eq *SQLEqualsExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	left, err := t.ToAggregationLanguage(eq.left)
	if err != nil {
		return nil, err
	}

	right, err := t.ToAggregationLanguage(eq.right)
	if err != nil {
		return nil, err
	}

	letAssignment := bson.M{
		"left": left, "right": right,
	}

	letEvaluation := bsonutil.WrapInNullCheckedCond(
		nil,
		bson.M{
			bsonutil.OpEq: []interface{}{"$$left", "$$right"},
		},
		"$$left",
		"$$right",
	)

	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil
}

// ToMatchLanguage translates SQLEqualsExpr into something that can
// be used in an match expression. If SQLEqualsExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLEqualsExpr.
func (eq *SQLEqualsExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	match, ok := t.translateOperator(bsonutil.OpEq, eq.left, eq.right)
	if !ok {
		return nil, eq
	}
	return match, nil
}

// EvalType returns the EvalType associated with SQLEqualsExpr.
func (*SQLEqualsExpr) EvalType() EvalType {
	return EvalBoolean
}

func (eq *SQLEqualsExpr) reconcile() (SQLExpr, error) {
	var reconciled bool
	var err error

	left := eq.left
	right := eq.right

	if isBooleanColumnAndNumber(left, right) || isBooleanColumnAndNumber(right, left) {
		var col SQLColumnExpr
		var lit SQLNumber

		switch left.EvalType() {
		case EvalBoolean:
			col = left.(SQLColumnExpr)
			lit = right.(SQLNumber)
		default:
			col = right.(SQLColumnExpr)
			lit = left.(SQLNumber)
		}

		if ilit := Int64(lit); ilit == 1 || ilit == 0 {
			left = col
			right = NewSQLConvertExpr(lit.(SQLExpr), EvalBoolean)
			reconciled = true
		}
	}

	if !reconciled {
		left, right, err = ReconcileSQLExprs(eq.left, eq.right, nil)
	}

	return &SQLEqualsExpr{left, right}, err
}

// SQLExistsExpr is a wrapper around a PlanStage representing a check for
// results from a subquery. It evaluates to true if any result is returned
// from the subquery. A SQLExistsExpr always evaluates to a boolean.
type SQLExistsExpr struct {
	correlated bool
	plan       PlanStage
	// We always cache non-correlated subquery results in their entirety.
	// This is a fine place to be more clever in the future.
	// SQLExistsExpr is cached as a boolean.
	cache SQLBool
}

// NewSQLExistsExpr is a constructor for SQLExistsExpr.
func NewSQLExistsExpr(correlated bool, plan PlanStage) *SQLExistsExpr {
	return &SQLExistsExpr{
		correlated: correlated,
		plan:       plan,
	}
}

// SkipConstantFolding indicates that we should not attempt to
// constant-fold this expression.
func (se *SQLExistsExpr) SkipConstantFolding() bool {
	return true
}

func (*SQLExistsExpr) evaluateFromPlan(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, plan PlanStage) (SQLBool, error) {
	iter, err := plan.Open(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	row := &Row{}

	hasNext := iter.Next(ctx, row)
	// release this memory here... it will be re-allocated by a consuming
	// stage
	if err = cfg.memoryMonitor.Release(row.Data.Size()); err != nil {
		_ = iter.Close()
		return nil, err
	}
	if hasNext {
		return NewSQLBool(cfg.sqlValueKind, true), iter.Close()
	}
	return NewSQLBool(cfg.sqlValueKind, false), iter.Close()
}

// Evaluate evaluates a SQLExistsExpr into a SQLValue.
// EXISTS returns true if its subquery returns at least one row.
// False is returned if there are no rows. EXISTS never returns NULL.
func (se *SQLExistsExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	if se.correlated {
		return se.evaluateFromPlan(ctx, cfg, st.SubqueryState(), se.plan)
	}
	if se.cache == nil {
		var err error
		// Populate cache.
		se.cache, err = se.evaluateFromPlan(ctx, cfg, st, se.plan)
		if err != nil {
			return nil, err
		}
	}
	// Read from cache.
	return se.cache, nil
}

func (se *SQLExistsExpr) String() string {
	return fmt.Sprintf("exists(%s)", PrettyPrintPlan(se.plan))
}

// EvalType returns the EvalType associated with SQLExistsExpr.
func (*SQLExistsExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLGreaterThanExpr evaluates to true when the left is greater than the right.
type SQLGreaterThanExpr sqlBinaryNode

var _ translatableToAggregation = (*SQLGreaterThanExpr)(nil)
var _ translatableToMatch = (*SQLGreaterThanExpr)(nil)

// NewSQLGreaterThanExpr is a constructor for SQLGreaterThanExpr.
func NewSQLGreaterThanExpr(left, right SQLExpr) *SQLGreaterThanExpr {
	return &SQLGreaterThanExpr{
		left:  left,
		right: right,
	}
}

// Evaluate evaluates a SQLGreaterThanExpr into a SQLValue.
func (gt *SQLGreaterThanExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {

	leftVal, err := gt.left.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	rightVal, err := gt.right.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(cfg.sqlValueKind, gt.EvalType()), nil
	}

	c, err := CompareTo(leftVal, rightVal, st.collation)
	if err == nil {
		return NewSQLBool(cfg.sqlValueKind, c > 0), nil
	}
	return NewSQLBool(cfg.sqlValueKind, false), err
}

// Normalize will attempt to change SQLGreaterThanExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (gt *SQLGreaterThanExpr) Normalize(kind SQLValueKind) Node {
	if hasNullExpr(gt.left, gt.right) {
		return NewSQLNull(kind, gt.EvalType())
	}

	if shouldFlip(sqlBinaryNode(*gt)) {
		return &SQLLessThanExpr{gt.right, gt.left}
	}

	return gt
}

func (gt *SQLGreaterThanExpr) String() string {
	return fmt.Sprintf("%v>%v", gt.left, gt.right)
}

// ToAggregationLanguage translates SQLGreaterThanExpr into something that can
// be used in an aggregation pipeline. If SQLGreaterThanExpr cannot be translated,
// it will return nil and error.
func (gt *SQLGreaterThanExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	left, err := t.ToAggregationLanguage(gt.left)
	if err != nil {
		return nil, err
	}

	right, err := t.ToAggregationLanguage(gt.right)
	if err != nil {
		return nil, err
	}

	letAssignment := bson.M{
		"left": left, "right": right,
	}

	letEvaluation := bsonutil.WrapInNullCheckedCond(
		nil,
		bson.M{
			bsonutil.OpGt: []interface{}{"$$left", "$$right"},
		},
		"$$left",
		"$$right",
	)

	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil

}

// ToMatchLanguage translates SQLGreaterThanExpr into something that can
// be used in an match expression. If SQLGreaterThanExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLGreaterThanExpr.
func (gt *SQLGreaterThanExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	match, ok := t.translateOperator(bsonutil.OpGt, gt.left, gt.right)
	if !ok {
		return nil, gt
	}
	return match, nil
}

// EvalType returns the EvalType associated with SQLGreaterThanExpr.
func (*SQLGreaterThanExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLGreaterThanOrEqualExpr evaluates to true when the left is greater than or equal to the right.
type SQLGreaterThanOrEqualExpr sqlBinaryNode

var _ translatableToAggregation = (*SQLGreaterThanOrEqualExpr)(nil)

// NewSQLGreaterThanOrEqualExpr is a constructor for SQLGreaterThanOrEqualExpr.
func NewSQLGreaterThanOrEqualExpr(left, right SQLExpr) *SQLGreaterThanOrEqualExpr {
	return &SQLGreaterThanOrEqualExpr{
		left:  left,
		right: right,
	}
}

// Evaluate evaluates a SQLGreaterThanOrEqualExpr into a SQLValue.
func (gte *SQLGreaterThanOrEqualExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {

	leftVal, err := gte.left.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	rightVal, err := gte.right.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(cfg.sqlValueKind, gte.EvalType()), nil
	}

	c, err := CompareTo(leftVal, rightVal, st.collation)
	if err == nil {
		return NewSQLBool(cfg.sqlValueKind, c >= 0), nil
	}

	return NewSQLBool(cfg.sqlValueKind, false), err
}

// Normalize will attempt to change SQLGreaterThanOrEqualExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (gte *SQLGreaterThanOrEqualExpr) Normalize(kind SQLValueKind) Node {
	if hasNullExpr(gte.left, gte.right) {
		return NewSQLNull(kind, gte.EvalType())
	}

	if shouldFlip(sqlBinaryNode(*gte)) {
		return &SQLLessThanOrEqualExpr{gte.right, gte.left}
	}

	return gte
}

func (gte *SQLGreaterThanOrEqualExpr) String() string {
	return fmt.Sprintf("%v>=%v", gte.left, gte.right)
}

// ToAggregationLanguage translates SQLGreaterThanOrEqualExpr into something
// that can be used in an aggregation pipeline. If SQLGreaterThanOrEqualExpr
// cannot be translated, it will return nil and error.
func (gte *SQLGreaterThanOrEqualExpr) ToAggregationLanguage(
	t *PushdownTranslator) (interface{}, error) {
	left, err := t.ToAggregationLanguage(gte.left)
	if err != nil {
		return nil, err
	}

	right, err := t.ToAggregationLanguage(gte.right)
	if err != nil {
		return nil, err
	}

	letAssignment := bson.M{
		"left": left, "right": right,
	}

	letEvaluation := bsonutil.WrapInNullCheckedCond(
		nil,
		bson.M{
			bsonutil.OpGte: []interface{}{"$$left", "$$right"},
		},
		"$$left",
		"$$right",
	)

	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil

}

// ToMatchLanguage translates SQLGreaterThanOrEqualExpr into something that can
// be used in an match expression. If SQLGreaterThanOrEqualExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLGreaterThanOrEqualExpr.
func (gte *SQLGreaterThanOrEqualExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	match, ok := t.translateOperator(bsonutil.OpGte, gte.left, gte.right)
	if !ok {
		return nil, gte
	}
	return match, nil
}

// EvalType returns the EvalType associated with SQLGreaterThanOrEqualExpr.
func (*SQLGreaterThanOrEqualExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLIDivideExpr evaluates the integer quotient of the left expression divided by the right.
type SQLIDivideExpr sqlBinaryNode

var _ translatableToAggregation = (*SQLIDivideExpr)(nil)

// Evaluate evaluates a SQLIDivideExpr into a SQLValue.
func (div *SQLIDivideExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	leftVal, err := div.left.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	rightVal, err := div.right.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	frightVal := Float64(rightVal)
	if frightVal == 0.0 || hasNullValue(leftVal, rightVal) {
		// NOTE: this is per the mysql manual.
		return NewSQLNull(cfg.sqlValueKind, div.EvalType()), nil
	}

	return NewSQLInt64(cfg.sqlValueKind, int64(Float64(leftVal)/frightVal)), nil
}

// Normalize will attempt to change SQLIDivideExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (div *SQLIDivideExpr) Normalize(kind SQLValueKind) Node {
	if hasNullExpr(div.left, div.right) {
		return NewSQLNull(kind, div.EvalType())
	}
	return div
}

func (div *SQLIDivideExpr) String() string {
	return fmt.Sprintf("%v/%v", div.left, div.right)
}

// ToAggregationLanguage translates SQLIDivideExpr into something that can
// be used in an aggregation pipeline. If SQLIDivideExpr cannot be translated,
// it will return nil and error.
func (div *SQLIDivideExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	left, err := t.ToAggregationLanguage(div.left)
	if err != nil {
		return nil, err
	}

	right, err := t.ToAggregationLanguage(div.right)
	if err != nil {
		return nil, err
	}

	letAssignment := bson.M{
		"left": left, "right": right,
	}

	letEvaluation := bsonutil.WrapInCond(
		nil,
		bson.M{
			"$trunc": []interface{}{
				bson.M{
					bsonutil.OpDivide: []interface{}{"$$left", "$$right"},
				},
			},
		},
		bson.M{bsonutil.OpEq: []interface{}{"$$right", 0}},
	)

	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil

}

// EvalType returns the EvalType associated with SQLIDivideExpr.
func (div *SQLIDivideExpr) EvalType() EvalType {
	return preferentialType(div.left, div.right)
}

// SQLInExpr evaluates to true if the left is in any of the values on the right.
type SQLInExpr sqlBinaryNode

var _ translatableToAggregation = (*SQLInExpr)(nil)

// Evaluate evaluates a SQLInExpr into a SQLValue.
func (in *SQLInExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	left, err := in.left.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	right, err := in.right.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	// right child must be of type SQLValues
	rightChild, ok := right.(*SQLValues)
	if !ok {
		child, typeOk := right.(SQLValue)
		if !typeOk {
			return NewSQLBool(cfg.sqlValueKind, false),
				fmt.Errorf("right 'in' expression is type %T - expected tuple",
					right)
		}
		rightChild = &SQLValues{[]SQLValue{child}}
	}

	leftChild, ok := left.(*SQLValues)
	if ok && in.left.EvalType() != EvalTuple {
		if len(leftChild.Values) != 1 {
			return NewSQLBool(cfg.sqlValueKind, false),
				fmt.Errorf("left operand should contain 1 column - got %v",
					len(leftChild.Values))
		}
		left = leftChild.Values[0]
	} else if in.left.EvalType() == EvalTuple {
		left = &SQLValues{leftChild.Values}
	} else if left.IsNull() {
		return NewSQLNull(cfg.sqlValueKind, in.EvalType()), nil
	}

	nullInValues := false
	for _, right := range rightChild.Values {
		if right.IsNull() {
			nullInValues = true
		}
		eq := &SQLEqualsExpr{left, right}
		result, err := eq.Evaluate(ctx, cfg, st)
		if err != nil {
			return NewSQLBool(cfg.sqlValueKind, false), err
		}
		if Bool(result) {
			return NewSQLBool(cfg.sqlValueKind, true), nil
		}
	}

	if nullInValues {
		return NewSQLNull(cfg.sqlValueKind, in.EvalType()), nil
	}

	return NewSQLBool(cfg.sqlValueKind, false), nil
}

// Normalize will attempt to change SQLInExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (in *SQLInExpr) Normalize(kind SQLValueKind) Node {
	if hasNullExpr(in.left) {
		return NewSQLNull(kind, in.EvalType())
	}

	return in
}

func (in *SQLInExpr) String() string {
	return fmt.Sprintf("%v in %v", in.left, in.right)
}

// ToAggregationLanguage translates SQLInExpr into something that can
// be used in an aggregation pipeline. If SQLInExpr cannot be translated,
// it will return nil and error.
func (in *SQLInExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	left, err := t.ToAggregationLanguage(in.left)
	if err != nil {
		return nil, err
	}

	exprs := getSQLInExprs(in.right)
	if exprs == nil {
		return nil, fmt.Errorf("cannot translate %v to aggregation language", in)
	}

	nullInValues := false
	var right []interface{}
	for _, expr := range exprs {
		sqlVal, ok := expr.(SQLValue)
		if ok && sqlVal.IsNull() {
			nullInValues = true
			continue
		}

		val, err := t.ToAggregationLanguage(expr)
		if err != nil {
			return nil, err
		}
		right = append(right, val)
	}

	return bsonutil.WrapInNullCheckedCond(
		nil,
		bsonutil.WrapInCond(
			true,
			bsonutil.WrapInCond(
				nil,
				false,
				bson.M{bsonutil.OpEq: []interface{}{nullInValues, true}},
			),
			bson.M{bsonutil.OpGt: []interface{}{
				bson.M{bsonutil.OpSize: bson.M{bsonutil.OpFilter: bson.M{"input": right,
					"as":   "item",
					"cond": bson.M{bsonutil.OpEq: []interface{}{"$$item", left}},
				}}},
				bsonutil.WrapInLiteral(0),
			}}),
		left,
	), nil

}

// ToMatchLanguage translates SQLInExpr into something that can
// be used in an match expression. If SQLInExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLInExpr.
func (in *SQLInExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	name, ok := t.getFieldName(in.left)
	if !ok {
		return nil, in
	}

	exprs := getSQLInExprs(in.right)
	if exprs == nil {
		return nil, in
	}

	values := []interface{}{}

	for _, expr := range exprs {
		value, ok := t.getValue(expr)
		if !ok {
			return nil, in
		}
		values = append(values, value)
	}

	return bson.M{name: bson.M{bsonutil.OpIn: values}}, nil
}

// EvalType returns the EvalType associated with SQLInExpr.
func (*SQLInExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLIsExpr evaluates to true if the left is equal to the boolean value on the right.
type SQLIsExpr sqlBinaryNode

var _ translatableToAggregation = (*SQLIsExpr)(nil)
var _ translatableToMatch = (*SQLIsExpr)(nil)

// Evaluate evaluates a SQLIsExpr into a SQLValue.
func (is *SQLIsExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	leftVal, err := is.left.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	rightVal, err := is.right.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	if leftVal.IsNull() {
		if _, ok := rightVal.(SQLBool); ok {
			return NewSQLBool(cfg.sqlValueKind, false), nil
		}
		return NewSQLBool(cfg.sqlValueKind, true), nil
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLBool(cfg.sqlValueKind, false), nil
	}

	if Bool(leftVal) && Bool(rightVal) || IsFalsy(leftVal) && IsFalsy(rightVal) {
		return NewSQLBool(cfg.sqlValueKind, true), nil
	}

	return NewSQLBool(cfg.sqlValueKind, false), nil

}

func (is *SQLIsExpr) String() string {
	return fmt.Sprintf("%v is %v", is.left, is.right)
}

// ToAggregationLanguage translates SQLIsExpr into something that can
// be used in an aggregation pipeline. If SQLIsExpr cannot be translated,
// it will return nil and error.
func (is *SQLIsExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	left, err := t.ToAggregationLanguage(is.left)
	if err != nil {
		return nil, err
	}

	// if right side is {null,unknown}, it's a simple case
	sqlVal, ok := is.right.(SQLValue)
	if ok && sqlVal.IsNull() {
		return bsonutil.WrapInOp(bsonutil.OpLte,
			left,
			bsonutil.WrapInLiteral(nil),
		), nil
	}

	right, err := t.ToAggregationLanguage(is.right)
	if err != nil {
		return nil, err
	}

	// if left side is a boolean, this is still simple
	if is.left.EvalType() == EvalBoolean {
		return bsonutil.WrapInOp(bsonutil.OpEq,
			left,
			right,
		), nil
	}

	// otherwise, left side is a number type
	if is.right == NewSQLBool(t.valueKind(), true) {
		return bsonutil.WrapInCond(
			false,
			bsonutil.WrapInOp(bsonutil.OpNeq,
				left,
				0,
			),
			bsonutil.WrapInNullCheck(left),
		), nil
	} else if is.right == NewSQLBool(t.valueKind(), false) {
		return bsonutil.WrapInOp(bsonutil.OpEq,
			left,
			0,
		), nil
	}

	// SQL Values
	return nil, fmt.Errorf("cannot translate %v to aggregation language", is)
}

// ToMatchLanguage translates SQLIsExpr into something that can
// be used in an match expression. If SQLIsExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLIsExpr.
func (is *SQLIsExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	name, ok := t.getFieldName(is.left)
	if !ok {
		return nil, is
	}

	rightVal, ok := is.right.(SQLValue)
	if !ok {
		return nil, is
	}

	if rightVal.IsNull() {
		return bson.M{name: nil}, nil
	}

	rightBool, ok := rightVal.(SQLBool)
	if !ok {
		return nil, is
	}

	if rightBool.Value().(bool) {
		if is.left.EvalType() == EvalBoolean {
			return bson.M{name: true}, nil
		}
		return bson.M{
			bsonutil.OpAnd: []interface{}{
				bson.M{name: bson.M{bsonutil.OpNeq: 0}},
				bson.M{name: bson.M{bsonutil.OpNeq: nil}},
			},
		}, nil
	}

	if is.left.EvalType() == EvalBoolean {
		return bson.M{name: false}, nil
	}
	return bson.M{name: 0}, nil
}

// EvalType returns the EvalType associated with SQLIsExpr.
func (*SQLIsExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLLessThanExpr evaluates to true when the left is less than the right.
type SQLLessThanExpr sqlBinaryNode

var _ translatableToAggregation = (*SQLLessThanExpr)(nil)
var _ translatableToMatch = (*SQLLessThanExpr)(nil)

// NewSQLLessThanExpr is a constructor for SQLLessThanExpr.
func NewSQLLessThanExpr(left, right SQLExpr) *SQLLessThanExpr {
	return &SQLLessThanExpr{
		left:  left,
		right: right,
	}
}

// Evaluate evaluates a SQLLessThanExpr into a SQLValue.
func (lt *SQLLessThanExpr) Evaluate(ctx context.Context,
	cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {

	leftVal, err := lt.left.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	rightVal, err := lt.right.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(cfg.sqlValueKind, lt.EvalType()), nil
	}

	c, err := CompareTo(leftVal, rightVal, st.collation)
	if err == nil {
		return NewSQLBool(cfg.sqlValueKind, c < 0), nil
	}
	return NewSQLBool(cfg.sqlValueKind, false), err
}

// Normalize will attempt to change SQLLessThanExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (lt *SQLLessThanExpr) Normalize(kind SQLValueKind) Node {
	if hasNullExpr(lt.left, lt.right) {
		return NewSQLNull(kind, lt.EvalType())
	}

	if shouldFlip(sqlBinaryNode(*lt)) {
		return &SQLGreaterThanExpr{lt.right, lt.left}
	}

	return lt
}

func (lt *SQLLessThanExpr) String() string {
	return fmt.Sprintf("%v<%v", lt.left, lt.right)
}

// ToAggregationLanguage translates SQLLessThanExpr into something that can
// be used in an aggregation pipeline. If SQLLessThanExpr cannot be translated,
// it will return nil and error.
func (lt *SQLLessThanExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	left, err := t.ToAggregationLanguage(lt.left)
	if err != nil {
		return nil, err
	}

	right, err := t.ToAggregationLanguage(lt.right)
	if err != nil {
		return nil, err
	}

	letAssignment := bson.M{
		"left": left, "right": right,
	}

	letEvaluation := bsonutil.WrapInNullCheckedCond(
		nil,
		bson.M{
			bsonutil.OpLt: []interface{}{"$$left", "$$right"},
		},
		"$$left", "$$right",
	)

	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil

}

// ToMatchLanguage translates SQLLessThanExpr into something that can
// be used in an match expression. If SQLLessThanExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLLessThanExpr.
func (lt *SQLLessThanExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	match, ok := t.translateOperator(bsonutil.OpLt, lt.left, lt.right)
	if !ok {
		return nil, lt
	}
	return match, nil
}

// EvalType returns the EvalType associated with SQLLessThanExpr.
func (*SQLLessThanExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLLessThanOrEqualExpr evaluates to true when the left is less than or equal to the right.
type SQLLessThanOrEqualExpr sqlBinaryNode

var _ translatableToAggregation = (*SQLLessThanOrEqualExpr)(nil)
var _ translatableToMatch = (*SQLLessThanOrEqualExpr)(nil)

// NewSQLLessThanOrEqualExpr is a constructor for SQLLessThanOrEqualExpr.
func NewSQLLessThanOrEqualExpr(left, right SQLExpr) *SQLLessThanOrEqualExpr {
	return &SQLLessThanOrEqualExpr{
		left:  left,
		right: right,
	}
}

// Evaluate evaluates a SQLLessThanOrEqualExpr into a SQLValue.
func (lte *SQLLessThanOrEqualExpr) Evaluate(ctx context.Context,
	cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {

	leftVal, err := lte.left.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	rightVal, err := lte.right.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(cfg.sqlValueKind, lte.EvalType()), nil
	}

	c, err := CompareTo(leftVal, rightVal, st.collation)
	if err == nil {
		return NewSQLBool(cfg.sqlValueKind, c <= 0), nil
	}
	return NewSQLBool(cfg.sqlValueKind, false), err
}

// Normalize will attempt to change SQLLessThanOrEqualExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (lte *SQLLessThanOrEqualExpr) Normalize(kind SQLValueKind) Node {
	if hasNullExpr(lte.left, lte.right) {
		return NewSQLNull(kind, lte.EvalType())
	}

	if shouldFlip(sqlBinaryNode(*lte)) {
		return &SQLGreaterThanOrEqualExpr{lte.right, lte.left}
	}

	return lte
}

func (lte *SQLLessThanOrEqualExpr) String() string {
	return fmt.Sprintf("%v<=%v", lte.left, lte.right)
}

// ToAggregationLanguage translates SQLLessThanOrEqualExpr into something that can
// be used in an aggregation pipeline. If SQLLessThanOrEqualExpr cannot be translated,
// it will return nil and error.
func (lte *SQLLessThanOrEqualExpr) ToAggregationLanguage(
	t *PushdownTranslator) (interface{}, error) {
	left, err := t.ToAggregationLanguage(lte.left)
	if err != nil {
		return nil, err
	}

	right, err := t.ToAggregationLanguage(lte.right)
	if err != nil {
		return nil, err
	}

	letAssignment := bson.M{
		"left": left, "right": right,
	}

	letEvaluation := bsonutil.WrapInNullCheckedCond(
		nil,
		bson.M{
			bsonutil.OpLte: []interface{}{"$$left", "$$right"},
		},
		"$$left", "$$right",
	)

	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil

}

// ToMatchLanguage translates SQLLessThanOrEqualExpr into something that can
// be used in an match expression. If SQLLessThanOrEqualExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLLessThanOrEqualExpr.
func (lte *SQLLessThanOrEqualExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	match, ok := t.translateOperator(bsonutil.OpLte, lte.left, lte.right)
	if !ok {
		return nil, lte
	}
	return match, nil
}

// EvalType returns the EvalType associated with SQLLessThanOrEqualExpr.
func (*SQLLessThanOrEqualExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLLikeExpr evaluates to true if the left is 'like' the right.
type SQLLikeExpr struct {
	left          SQLExpr
	right         SQLExpr
	escape        SQLExpr
	caseSensitive bool
}

var _ translatableToMatch = (*SQLLikeExpr)(nil)

// NewSQLLikeExpr is a constructor for SQLLikeExpr.
func NewSQLLikeExpr(left, right, escape SQLExpr, caseSensitive bool) *SQLLikeExpr {
	return &SQLLikeExpr{
		left:          left,
		right:         right,
		escape:        escape,
		caseSensitive: caseSensitive,
	}
}

// Evaluate evaluates a SQLLikeExpr into a SQLValue.
func (l *SQLLikeExpr) Evaluate(ctx context.Context,
	cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {

	value, err := l.left.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if hasNullValue(value) {
		return NewSQLNull(cfg.sqlValueKind, l.EvalType()), nil
	}

	data := value.String()

	value, err = l.right.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if hasNullValue(value) {
		return NewSQLNull(cfg.sqlValueKind, l.EvalType()), nil
	}

	escape, err := l.escape.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	escapeSeq := []rune(escape.String())
	if len(escapeSeq) > 1 {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErWrongArguments, "ESCAPE")
	}

	var escapeChar rune
	if len(escapeSeq) == 1 {
		escapeChar = escapeSeq[0]
	}

	pattern := "(?i)"
	if l.caseSensitive {
		pattern = ""
	}
	pattern += ConvertSQLValueToPattern(value, escapeChar)

	matches, err := regexp.Match(pattern, []byte(data))
	if err != nil {
		return nil, err
	}

	return NewSQLBool(cfg.sqlValueKind, matches), nil
}

// Normalize will attempt to change SQLLikeExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (l *SQLLikeExpr) Normalize(kind SQLValueKind) Node {
	if right, ok := l.right.(SQLValue); ok {
		if hasNullValue(right) {
			return NewSQLNull(kind, l.EvalType())
		}
	}

	return l
}

func (l *SQLLikeExpr) String() string {
	likeType := "like"
	if l.caseSensitive {
		likeType = "like binary"
	}
	return fmt.Sprintf("%v %v %v", l.left, likeType, l.right)
}

// ToMatchLanguage translates SQLLikeExpr into something that can
// be used in an match expression. If SQLLikeExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLLikeExpr.
func (l *SQLLikeExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	// we cannot do a like comparison on an ObjectID in mongodb.
	if c, ok := l.left.(SQLColumnExpr); ok &&
		c.columnType.MongoType == schema.MongoObjectID {
		return nil, l
	}

	name, ok := t.getFieldName(l.left)
	if !ok {
		return nil, l
	}

	value, ok := l.right.(SQLValue)
	if !ok {
		return nil, l
	}

	if hasNullValue(value) {
		return nil, l
	}

	escape, ok := l.escape.(SQLValue)
	if !ok {
		return nil, l
	}

	escapeSeq := []rune(escape.String())
	if len(escapeSeq) > 1 {
		return nil, l
	}

	var escapeChar rune
	if len(escapeSeq) == 1 {
		escapeChar = escapeSeq[0]
	}

	pattern := ConvertSQLValueToPattern(value, escapeChar)
	opts := "i"
	if l.caseSensitive {
		opts = ""
	}

	return bson.M{name: bson.M{bsonutil.OpRegex: bson.RegEx{Pattern: pattern, Options: opts}}}, nil
}

// EvalType returns the EvalType associated with SQLLikeExpr.
func (*SQLLikeExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLModExpr evaluates the modulus of two expressions
type SQLModExpr sqlBinaryNode

var _ translatableToAggregation = (*SQLModExpr)(nil)

// Evaluate evaluates a SQLModExpr into a SQLValue.
func (mod *SQLModExpr) Evaluate(ctx context.Context,
	cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	leftVal, err := mod.left.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	rightVal, err := mod.right.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	frightVal := Float64(rightVal)
	if math.Abs(frightVal) == 0.0 || hasNullValue(leftVal, rightVal) {
		return NewSQLNull(cfg.sqlValueKind, mod.EvalType()), nil
	}

	modVal := math.Mod(Float64(leftVal), frightVal)
	if modVal == -0 {
		modVal *= -1
	}

	return NewSQLFloat(cfg.sqlValueKind, modVal), nil
}

func (mod *SQLModExpr) String() string {
	return fmt.Sprintf("%v/%v", mod.left, mod.right)
}

// ToAggregationLanguage translates SQLModExpr into something that can
// be used in an aggregation pipeline. If SQLModExpr cannot be translated,
// it will return nil and error.
func (mod *SQLModExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	left, err := t.ToAggregationLanguage(mod.left)
	if err != nil {
		return nil, err
	}

	right, err := t.ToAggregationLanguage(mod.right)
	if err != nil {
		return nil, err
	}

	return bson.M{bsonutil.OpMod: []interface{}{left, right}}, nil

}

// EvalType returns the EvalType associated with SQLModExpr.
func (mod *SQLModExpr) EvalType() EvalType {
	return preferentialType(mod.left, mod.right)
}

// SQLMultiplyExpr evaluates to the product of two expressions
type SQLMultiplyExpr sqlBinaryNode

var _ translatableToAggregation = (*SQLMultiplyExpr)(nil)

// Evaluate evaluates a SQLMultiplyExpr into a SQLValue.
func (mult *SQLMultiplyExpr) Evaluate(ctx context.Context,
	cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	leftVal, err := mult.left.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	rightVal, err := mult.right.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(cfg.sqlValueKind, mult.EvalType()), nil
	}

	return doArithmetic(leftVal, rightVal, MULT)
}

func (mult *SQLMultiplyExpr) String() string {
	return fmt.Sprintf("%v*%v", mult.left, mult.right)
}

// ToAggregationLanguage translates SQLMultiplyExpr into something that can
// be used in an aggregation pipeline. If SQLMultiplyExpr cannot be translated,
// it will return nil and error.
func (mult *SQLMultiplyExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	left, err := t.ToAggregationLanguage(mult.left)
	if err != nil {
		return nil, err
	}

	right, err := t.ToAggregationLanguage(mult.right)
	if err != nil {
		return nil, err
	}

	return bson.M{bsonutil.OpMultiply: eatChildren(bsonutil.OpMultiply, left, right)}, nil
}

// EvalType returns the EvalType associated with SQLMultiplyExpr.
func (mult *SQLMultiplyExpr) EvalType() EvalType {
	return EvalDouble
}

// SQLNotEqualsExpr evaluates to true if the left does not equal the right.
type SQLNotEqualsExpr sqlBinaryNode

var _ translatableToAggregation = (*SQLNotEqualsExpr)(nil)
var _ translatableToMatch = (*SQLNotEqualsExpr)(nil)

// NewSQLNotEqualsExpr is a constructor for SQLNotEqualsExpr.
func NewSQLNotEqualsExpr(left, right SQLExpr) *SQLNotEqualsExpr {
	return &SQLNotEqualsExpr{
		left:  left,
		right: right,
	}
}

// Evaluate evaluates a SQLNotEqualsExpr into a SQLValue.
func (neq *SQLNotEqualsExpr) Evaluate(ctx context.Context,
	cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {

	leftVal, err := neq.left.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	rightVal, err := neq.right.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(cfg.sqlValueKind, neq.EvalType()), nil
	}

	c, err := CompareTo(leftVal, rightVal, st.collation)
	if err == nil {
		return NewSQLBool(cfg.sqlValueKind, c != 0), nil
	}

	return NewSQLBool(cfg.sqlValueKind, false), err
}

// Normalize will attempt to change SQLNotEqualsExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (neq *SQLNotEqualsExpr) Normalize(kind SQLValueKind) Node {
	if hasNullExpr(neq.left, neq.right) {
		return NewSQLNull(kind, neq.EvalType())
	}

	if shouldFlip(sqlBinaryNode(*neq)) {
		return &SQLNotEqualsExpr{neq.right, neq.left}
	}

	return neq
}

func (neq *SQLNotEqualsExpr) String() string {
	return fmt.Sprintf("%v != %v", neq.left, neq.right)
}

// ToAggregationLanguage translates SQLNotEqualsExpr into something that can
// be used in an aggregation pipeline. If SQLNotEqualsExpr cannot be translated,
// it will return nil and error.
func (neq *SQLNotEqualsExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	left, err := t.ToAggregationLanguage(neq.left)
	if err != nil {
		return nil, err
	}

	right, err := t.ToAggregationLanguage(neq.right)
	if err != nil {
		return nil, err
	}

	letAssignment := bson.M{
		"left": left, "right": right,
	}

	letEvaluation := bsonutil.WrapInNullCheckedCond(
		nil,
		bson.M{
			bsonutil.OpNeq: []interface{}{"$$left", "$$right"},
		},
		"$$left", "$$right",
	)

	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil

}

// ToMatchLanguage translates SQLNotEqualsExpr into something that can
// be used in an match expression. If SQLNotEqualsExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLNotEqualsExpr.
func (neq *SQLNotEqualsExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	match, ok := t.translateOperator(bsonutil.OpNeq, neq.left, neq.right)
	if !ok {
		return nil, neq
	}

	value, ok := t.getValue(neq.right)
	if !ok {
		return nil, neq
	}

	if value != nil {
		name, ok := t.getFieldName(neq.left)
		if !ok {
			return nil, neq
		}
		match = bson.M{
			bsonutil.OpAnd: []interface{}{match,
				bson.M{name: bson.M{bsonutil.OpNeq: nil}},
			},
		}
	}

	return match, nil
}

// EvalType returns the EvalType associated with SQLNotEqualsExpr.
func (*SQLNotEqualsExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLNotExpr evaluates to the inverse of its child.
type SQLNotExpr sqlUnaryNode

var _ reconcilingSQLExpr = (*SQLNotExpr)(nil)
var _ translatableToAggregation = (*SQLNotExpr)(nil)
var _ translatableToMatch = (*SQLNotExpr)(nil)

// NewSQLNotExpr is a constructor for SQLNotExpr.
func NewSQLNotExpr(operand SQLExpr) *SQLNotExpr {
	return &SQLNotExpr{operand}
}

// Evaluate evaluates a SQLNotExpr into a SQLValue.
func (not *SQLNotExpr) Evaluate(ctx context.Context,
	cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	operand, err := not.SQLExpr.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if hasNullValue(operand) {
		return NewSQLNull(cfg.sqlValueKind, not.EvalType()), nil
	}

	if !Bool(operand) {
		return NewSQLBool(cfg.sqlValueKind, true), nil
	}

	return NewSQLBool(cfg.sqlValueKind, false), nil
}

// Normalize will attempt to change SQLNotExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (not *SQLNotExpr) Normalize(kind SQLValueKind) Node {
	if operand, ok := not.SQLExpr.(SQLValue); ok {
		if hasNullValue(operand) {
			return NewSQLNull(kind, not.EvalType())
		}

		if Bool(operand) {
			return NewSQLBool(kind, false)
		} else if IsFalsy(operand) {
			return NewSQLBool(kind, true)
		}
	}

	return not
}

func (not *SQLNotExpr) String() string {
	return fmt.Sprintf("not %v", not.SQLExpr)
}

func (not *SQLNotExpr) reconcile() (SQLExpr, error) {
	expr := not.SQLExpr
	if !isBooleanComparable(expr.EvalType()) {
		expr = NewSQLConvertExpr(expr, EvalBoolean)
	}
	return &SQLNotExpr{expr}, nil
}

// ToAggregationLanguage translates SQLNotExpr into something that can
// be used in an aggregation pipeline. If SQLNotExpr cannot be translated,
// it will return nil and error.
func (not *SQLNotExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	op, err := t.ToAggregationLanguage(not.SQLExpr)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapInNullCheckedCond(nil, bson.M{bsonutil.OpNot: op}, op), nil

}

// ToMatchLanguage translates SQLNotExpr into something that can
// be used in an match expression. If SQLNotExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLNotExpr.
func (not *SQLNotExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	match, ex := t.ToMatchLanguage(not.SQLExpr)
	if match == nil {
		return nil, not
	} else if ex == nil {
		return negate(match), nil
	} else {
		// partial translation of Not
		return negate(match), &SQLNotExpr{ex}
	}

}

// EvalType returns the EvalType associated with SQLNotExpr.
func (*SQLNotExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLNullSafeEqualsExpr behaves like the = operator,
// but returns 1 rather than NULL if both operands are
// NULL, and 0 rather than NULL if one operand is NULL.
type SQLNullSafeEqualsExpr sqlBinaryNode

var _ translatableToAggregation = (*SQLNullSafeEqualsExpr)(nil)

// NewSQLNullSafeEqualsExpr is a constructor for SQLNullSafeEqualsExpr.
func NewSQLNullSafeEqualsExpr(left, right SQLExpr) *SQLNullSafeEqualsExpr {
	return &SQLNullSafeEqualsExpr{
		left:  left,
		right: right,
	}
}

// Evaluate evaluates a SQLNullSafeEqualsExpr into a SQLValue.
func (nse *SQLNullSafeEqualsExpr) Evaluate(ctx context.Context,
	cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {

	leftVal, err := nse.left.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	rightVal, err := nse.right.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	if leftVal.IsNull() {
		if rightVal.IsNull() {
			return NewSQLBool(cfg.sqlValueKind, true), nil
		}
		return NewSQLBool(cfg.sqlValueKind, false), nil
	}

	if rightVal.IsNull() {
		if leftVal.IsNull() {
			return NewSQLBool(cfg.sqlValueKind, true), nil
		}
		return NewSQLBool(cfg.sqlValueKind, false), nil
	}

	c, err := CompareTo(leftVal, rightVal, st.collation)
	if err == nil {
		return NewSQLBool(cfg.sqlValueKind, c == 0), nil
	}

	return NewSQLBool(cfg.sqlValueKind, false), err
}

// Normalize will attempt to change SQLNullSafeEqualsExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (nse *SQLNullSafeEqualsExpr) Normalize(kind SQLValueKind) Node {
	leftVal, leftIsVal := nse.left.(SQLValue)
	rightVal, rightIsVal := nse.right.(SQLValue)

	if !leftIsVal || !rightIsVal {
		return nse
	}

	if leftVal.IsNull() {
		if rightVal.IsNull() {
			return NewSQLBool(kind, true)
		}
		return NewSQLBool(kind, false)
	}

	if rightVal.IsNull() {
		if leftVal.IsNull() {
			return NewSQLBool(kind, true)
		}
		return NewSQLBool(kind, false)
	}

	return nse
}

func (nse *SQLNullSafeEqualsExpr) String() string {
	return fmt.Sprintf("%v <=> %v", nse.left, nse.right)
}

// ToAggregationLanguage translates SQLNullSafeEqualsExpr into something that can
// be used in an aggregation pipeline. If SQLNullSafeEqualsExpr cannot be translated,
// it will return nil and error.
func (nse *SQLNullSafeEqualsExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{},
	error) {
	left, err := t.ToAggregationLanguage(nse.left)
	if err != nil {
		return nil, err
	}

	right, err := t.ToAggregationLanguage(nse.right)
	if err != nil {
		return nil, err
	}

	return bson.M{bsonutil.OpEq: []interface{}{
		bson.M{bsonutil.OpIfNull: []interface{}{left, nil}},
		bson.M{bsonutil.OpIfNull: []interface{}{right, nil}},
	}}, nil
}

// EvalType returns the EvalType associated with SQLNullSafeEqualsExpr.
func (*SQLNullSafeEqualsExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLOrExpr evaluates to true if any of its children evaluate to true.
type SQLOrExpr sqlBinaryNode

var _ reconcilingSQLExpr = (*SQLOrExpr)(nil)
var _ translatableToAggregation = (*SQLOrExpr)(nil)
var _ translatableToMatch = (*SQLOrExpr)(nil)

// NewSQLOrExpr is a constructor for SQLOrExpr.
func NewSQLOrExpr(left, right SQLExpr) *SQLOrExpr {
	return &SQLOrExpr{
		left:  left,
		right: right,
	}
}

// Evaluate evaluates a SQLOrExpr into a SQLValue.
func (or *SQLOrExpr) Evaluate(ctx context.Context,
	cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	left, err := or.left.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	right, err := or.right.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if Bool(left) || Bool(right) {
		return NewSQLBool(cfg.sqlValueKind, true), nil
	}

	if hasNullValue(left, right) {
		return NewSQLNull(cfg.sqlValueKind, or.EvalType()), nil
	}

	return NewSQLBool(cfg.sqlValueKind, false), nil
}

// Normalize will attempt to change SQLOrExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (or *SQLOrExpr) Normalize(kind SQLValueKind) Node {
	left, leftOk := or.left.(SQLValue)

	if leftOk && Bool(left) {
		return NewSQLBool(kind, true)
	} else if leftOk && IsFalsy(left) {
		if or.right.EvalType() == EvalBoolean {
			return or.right
		}
		return NewSQLConvertExpr(or.right, EvalBoolean)
	}

	right, rightOk := or.right.(SQLValue)
	if rightOk && Bool(right) {
		return NewSQLBool(kind, true)
	} else if rightOk && IsFalsy(right) {
		if or.left.EvalType() == EvalBoolean {
			return or.left
		}
		return NewSQLConvertExpr(or.left, EvalBoolean)
	}

	return or
}

func (or *SQLOrExpr) String() string {
	return fmt.Sprintf("%v or %v", or.left, or.right)
}

func (or *SQLOrExpr) reconcile() (SQLExpr, error) {
	left := or.left
	right := or.right

	if !isBooleanComparable(left.EvalType()) {
		left = NewSQLConvertExpr(left, EvalBoolean)
	}
	if !isBooleanComparable(right.EvalType()) {
		right = NewSQLConvertExpr(right, EvalBoolean)
	}
	return &SQLOrExpr{left, right}, nil
}

// ToAggregationLanguage translates SQLOrExpr into something that can
// be used in an aggregation pipeline. If SQLOrExpr cannot be translated,
// it will return nil and error.
func (or *SQLOrExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	left, err := t.ToAggregationLanguage(or.left)
	if err != nil {
		return nil, err
	}

	right, err := t.ToAggregationLanguage(or.right)
	if err != nil {
		return nil, err
	}

	letAssignment := bson.M{
		"left": left, "right": right,
	}

	leftIsFalse := bson.M{bsonutil.OpOr: []interface{}{
		bson.M{bsonutil.OpEq: []interface{}{"$$left", false}},
		bson.M{bsonutil.OpEq: []interface{}{"$$left", 0}},
	}}

	leftIsTrue := bson.M{bsonutil.OpOr: []interface{}{
		bson.M{bsonutil.OpNeq: []interface{}{"$$left", false}},
		bson.M{bsonutil.OpNeq: []interface{}{"$$left", 0}},
	}}

	rightIsFalse := bson.M{bsonutil.OpOr: []interface{}{
		bson.M{bsonutil.OpEq: []interface{}{"$$right", false}},
		bson.M{bsonutil.OpEq: []interface{}{"$$right", 0}},
	}}

	rightIsTrue := bson.M{bsonutil.OpOr: []interface{}{
		bson.M{bsonutil.OpNeq: []interface{}{"$$right", false}},
		bson.M{bsonutil.OpNeq: []interface{}{"$$right", 0}},
	}}

	leftIsNull := bson.M{bsonutil.OpEq: []interface{}{
		bson.M{
			bsonutil.OpIfNull: []interface{}{"$$left", nil}},
		nil,
	}}

	rightIsNull := bson.M{bsonutil.OpEq: []interface{}{
		bson.M{
			bsonutil.OpIfNull: []interface{}{"$$right", nil}},
		nil,
	}}

	nullOrFalse := bson.M{bsonutil.OpOr: []interface{}{
		bson.M{bsonutil.OpAnd: []interface{}{
			rightIsNull, leftIsFalse,
		}},
		bson.M{bsonutil.OpAnd: []interface{}{
			leftIsNull, rightIsFalse,
		}},
	}}

	nullOrTrue := bson.M{bsonutil.OpOr: []interface{}{
		bson.M{bsonutil.OpAnd: []interface{}{
			rightIsNull, leftIsTrue,
		}},
		bson.M{bsonutil.OpAnd: []interface{}{
			leftIsNull, rightIsTrue,
		}},
	}}

	nullOrNull := bson.M{bsonutil.OpAnd: []interface{}{
		leftIsNull, rightIsNull,
	}}

	letEvaluation := bson.M{
		bsonutil.OpCond: []interface{}{
			nullOrNull,
			mgoNullLiteral,
			bsonutil.WrapInCond(
				mgoNullLiteral,
				bsonutil.WrapInCond(
					true,
					bsonutil.WrapInNullCheckedCond(
						mgoNullLiteral,
						bson.M{
							bsonutil.OpOr: []interface{}{"$$left", "$$right"},
						},
						"$$left", "$$right",
					),
					nullOrTrue,
				),
				nullOrFalse,
			),
		},
	}

	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil

}

// ToMatchLanguage translates SQLOrExpr into something that can
// be used in an match expression. If SQLOrExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLOrExpr.
func (or *SQLOrExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	left, exLeft := t.ToMatchLanguage(or.left)
	if exLeft != nil {
		// cannot partially translate an OR
		return nil, or
	}
	right, exRight := t.ToMatchLanguage(or.right)
	if exRight != nil {
		// cannot partially translate an OR
		return nil, or
	}

	cond := []interface{}{}

	if v, ok := left[bsonutil.OpOr]; ok {
		array := v.([]interface{})
		cond = append(cond, array...)
	} else {
		cond = append(cond, left)
	}

	if v, ok := right[bsonutil.OpOr]; ok {
		array := v.([]interface{})
		cond = append(cond, array...)
	} else {
		cond = append(cond, right)
	}

	return bson.M{bsonutil.OpOr: cond}, nil
}

// EvalType returns the EvalType associated with SQLOrExpr.
func (*SQLOrExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLRegexExpr evaluates to true if the operand matches the regex patttern.
type SQLRegexExpr struct {
	operand, pattern SQLExpr
}

var _ translatableToMatch = (*SQLRegexExpr)(nil)

// Evaluate evaluates a SQLRegexExpr into a SQLValue.
func (reg *SQLRegexExpr) Evaluate(ctx context.Context,
	cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	operandVal, err := reg.operand.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	patternVal, err := reg.pattern.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	if hasNullValue(operandVal, patternVal) {
		return NewSQLNull(cfg.sqlValueKind, reg.EvalType()), nil
	}

	pattern, patternOK := patternVal.(SQLVarchar)
	if patternOK {
		matcher, err := regexp.CompilePOSIX(pattern.String())
		if err != nil {
			return NewSQLBool(cfg.sqlValueKind, false), err
		}
		match := matcher.Find([]byte(operandVal.String()))
		if match != nil {
			return NewSQLBool(cfg.sqlValueKind, true), nil
		}
	}
	return NewSQLBool(cfg.sqlValueKind, false), nil
}

func (reg *SQLRegexExpr) String() string {
	return fmt.Sprintf("%s matches %s", reg.operand.String(), reg.pattern.String())
}

// ToMatchLanguage translates SQLRegexExpr into something that can
// be used in an match expression. If SQLRegexExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLRegexExpr.
func (reg *SQLRegexExpr) ToMatchLanguage(t *PushdownTranslator) (bson.M, SQLExpr) {
	name, ok := t.getFieldName(reg.operand)
	if !ok {
		return nil, reg
	}

	pattern, ok := reg.pattern.(SQLVarchar)
	if !ok {
		return nil, reg
	}
	// We need to check if the pattern is valid Extended POSIX regex
	// because MongoDB supports a superset of this specification called
	// PCRE.
	_, err := regexp.CompilePOSIX(pattern.String())
	if err != nil {
		return nil, reg
	}
	return bson.M{
		name: bson.M{
			bsonutil.OpRegex: bson.RegEx{
				Pattern: pattern.String(),
				Options: "",
			},
		},
	}, nil
}

// EvalType returns the EvalType associated with SQLRegexExpr.
func (*SQLRegexExpr) EvalType() EvalType {
	return EvalBoolean
}

// evaluatePlan converts a PlanStage into a table in memory, represented
// as a SQLValues of SQLValues. This table is used as the runtime value of a
// subquery expression and can be cached. Optimization opportunity:
// this function copies all of its input data, value-by-value.
func evaluatePlan(ctx context.Context, cfg *ExecutionConfig,
	st *ExecutionState, plan PlanStage) (*SQLValues, error) {

	iter, err := plan.Open(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	row := &Row{}
	valueTable := &SQLValues{}

	for iter.Next(ctx, row) {
		valueRow := &SQLValues{}
		// release this memory here... it will be re-allocated by a consuming
		// stage
		if err = cfg.memoryMonitor.Release(row.Data.Size()); err != nil {
			_ = iter.Close()
			return nil, err
		}

		// The full data copy here is unwanted.
		// This is a good place to attempt to improve performance.
		for _, value := range row.Data {
			valueRow.Values = append(valueRow.Values, value.Data)
		}
		valueTable.Values = append(valueTable.Values, valueRow)
	}

	return valueTable, iter.Close()
}

// evaluatePlanToScalar converts a PlanStage into a row in memory, represented
// as a SQLValues. This row is used as the runtime value of a
// subquery expression and can be cached. Optimization opportunity:
// this function copies all of its input data, value-by-value.
// This function implements the MySQL behavior of evaluating an empty input
// into a row of NULLs with the same dimension as the input.
func evaluatePlanToScalar(ctx context.Context, cfg *ExecutionConfig,
	st *ExecutionState, plan PlanStage) (*SQLValues, error) {

	iter, err := plan.Open(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	row := &Row{}

	valueRow := &SQLValues{}
	if iter.Next(ctx, row) {
		// release this memory here... it will be re-allocated by a consuming
		// stage
		if err = cfg.memoryMonitor.Release(row.Data.Size()); err != nil {
			_ = iter.Close()
			return nil, err
		}

		// The full data copy here is unwanted.
		// This is a good place to attempt to improve performance.
		for _, value := range row.Data {
			valueRow.Values = append(valueRow.Values, value.Data)
		}
	} else {
		// MySQL specific behavior here.
		for lcv := 0; lcv < len(plan.Columns()); lcv++ {
			valueRow.Values = append(valueRow.Values,
				NewSQLNullUntyped(cfg.sqlValueKind))
		}
	}

	// input must not have cardinality > 1
	if iter.Next(ctx, &Row{}) {
		_ = iter.Close()
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErSubqueryNoOneRow)
	}

	return valueRow, iter.Close()
}

// SQLInSubqueryExpr evaluates to true if the left expression is equal to any
// of the rows returned by the right subquery.
// Multi-column right subqueries are valid if the left is a tuple or
// subquery with the same number of columns.
// Multirow left subqueries are never valid.
// Note: This should not be confused with SQL's IN list construct which uses
// the same keyword.
// Note: This should not be. {A NOT IN (...)} is trivial to rewrite to
// {A = ANY (...)}.
type SQLInSubqueryExpr struct {
	correlated bool
	left       SQLExpr
	plan       PlanStage
	// We always cache non-correlated subquery results in their entirety.
	// This is a fine place to be more clever in the future.
	// SQLInSubqueryExpr can cache a whole table, with each row being compared
	// to the value result of the left expression.
	cache *SQLValues
}

// NewSQLInSubqueryExpr is a constructor for SQLInSubqueryExpr.
func NewSQLInSubqueryExpr(
	correlated bool,
	left SQLExpr,
	plan PlanStage) *SQLInSubqueryExpr {
	return &SQLInSubqueryExpr{
		correlated: correlated,
		left:       left,
		plan:       plan,
	}
}

// SkipConstantFolding indicates that we should not attempt to
// constant-fold this expression.
func (*SQLInSubqueryExpr) SkipConstantFolding() bool {
	return true
}

// Evaluate evaluates a SQLInSubqueryExpr into a SQLValue.
// IN performs a series of comparisons. IN always performs equality comparisons.
// The resulting comparisons within columns of a row are ANDed together.
// Comparisons from separate rows are ORed together.
// Using SQL three-value boolean logic, the results are as follows:
// If a series of comparisons within any row is all true, the result is true.
// If not, if any of the series returns NULL (the series contains NULL and no falses),
// the result is NULL.
// Else, the result is false.
func (si *SQLInSubqueryExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var table *SQLValues
	var err error
	if si.correlated {
		table, err = evaluatePlan(ctx, cfg, st.SubqueryState(), si.plan)
		if err != nil {
			return nil, err
		}
	} else {
		if si.cache == nil {
			// Populate cache.
			si.cache, err = evaluatePlan(ctx, cfg, st, si.plan)
			if err != nil {
				return nil, err
			}
			err = cfg.memoryMonitor.AcquireGlobal(si.cache.Size())
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		table = si.cache
	}

	// Determine length of the left expression.
	var leftLen int
	leftValues, isVals := si.left.(*SQLValues)
	leftTuple, isTups := si.left.(*SQLTupleExpr)
	if isVals {
		leftLen = len(leftValues.Values)
	} else if isTups {
		leftLen = len(leftTuple.Exprs)
	} else {
		leftLen = 1
	}

	sawNull := false
	for _, row := range table.Values {
		right := row.(*SQLValues)

		// Make sure the subquery returns the same number of columns as what
		// it's being compared to.
		// Note: This is redundant to do for each row.
		if leftLen != len(right.Values) {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, leftLen)
		}

		eq := &SQLEqualsExpr{si.left, right}
		var result SQLValue
		result, err = eq.Evaluate(ctx, cfg, st)
		if err != nil {
			return nil, err
		}
		if Bool(result) {
			return NewSQLBool(cfg.sqlValueKind, true), nil
		}
		if result.IsNull() {
			sawNull = true
		}
	}

	// left expression not found in right table.
	if sawNull {
		return NewSQLNullUntyped(cfg.sqlValueKind), nil
	}
	return NewSQLBool(cfg.sqlValueKind, false), nil
}

func (si *SQLInSubqueryExpr) String() string {
	return fmt.Sprintf("%v in (%s)", si.left, PrettyPrintPlan(si.plan))
}

// EvalType returns the EvalType associated with SQLInSubqueryExpr.
func (*SQLInSubqueryExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLNotInSubqueryExpr evaluates to true if the left expression is not equal
// to all of the rows returned by the right subquery.
// Multi-column right subqueries are valid if the left is a tuple or
// subquery with the same number of columns.
// Multirow left subqueries are never valid.
// Note: This should not be confused with SQL's NOT IN list construct which uses
// the same keyword.
// Note: This should not be. {A NOT IN (...)} is trivial to rewrite to
// {A <> ALL (...)}.
type SQLNotInSubqueryExpr struct {
	correlated bool
	left       SQLExpr
	plan       PlanStage
	// We always cache non-correlated subquery results in their entirety.
	// This is a fine place to be more clever in the future.
	// SQLNotInSubqueryExpr can cache a whole table, with each row being compared
	// to the value result of the left expression.
	cache *SQLValues
}

// NewSQLNotInSubqueryExpr is a constructor for SQLNotInSubqueryExpr.
func NewSQLNotInSubqueryExpr(
	correlated bool,
	left SQLExpr,
	plan PlanStage) *SQLNotInSubqueryExpr {
	return &SQLNotInSubqueryExpr{
		correlated: correlated,
		left:       left,
		plan:       plan,
	}
}

// SkipConstantFolding indicates that we should not attempt to
// constant-fold this expression.
func (*SQLNotInSubqueryExpr) SkipConstantFolding() bool {
	return true
}

// Evaluate evaluates a SQLNotInSubqueryExpr into a SQLValue.
// NOT IN performs a series of comparisons. NOT IN always performs not-equals comparisons.
// The resulting comparisons within columns of a row are ANDed together.
// Comparisons from separate rows are ORed together.
// Using SQL three-value boolean logic, the results are as follows:
// If a series of comparisons within any row is all true, the result is false.
// If not, if any of the series returns NULL (the series contains NULL and no falses),
// the result is NULL.
// Else, the result is true.
func (ni *SQLNotInSubqueryExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var table *SQLValues
	var err error
	if ni.correlated {
		table, err = evaluatePlan(ctx, cfg, st.SubqueryState(), ni.plan)
		if err != nil {
			return nil, err
		}
	} else {
		if ni.cache == nil {
			// Populate cache.
			ni.cache, err = evaluatePlan(ctx, cfg, st, ni.plan)
			if err != nil {
				return nil, err
			}
			err = cfg.memoryMonitor.AcquireGlobal(ni.cache.Size())
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		table = ni.cache
	}

	// Determine length of the left expression.
	var leftLen int
	leftValues, isVals := ni.left.(*SQLValues)
	leftTuple, isTups := ni.left.(*SQLTupleExpr)
	if isVals {
		leftLen = len(leftValues.Values)
	} else if isTups {
		leftLen = len(leftTuple.Exprs)
	} else {
		leftLen = 1
	}

	sawNull := false
	for _, row := range table.Values {
		right := row.(*SQLValues)

		// Make sure the subquery returns the same number of columns as what
		// it's being compared to.
		// Note: This is redundant to do for each row.
		if leftLen != len(right.Values) {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, leftLen)
		}

		eq := &SQLNotEqualsExpr{ni.left, right}
		var result SQLValue
		result, err = eq.Evaluate(ctx, cfg, st)
		if err != nil {
			return nil, err
		}
		if !Bool(result) {
			return NewSQLBool(cfg.sqlValueKind, false), nil
		}
		if result.IsNull() {
			sawNull = true
		}
	}

	// left expression not found in right table.
	if sawNull {
		return NewSQLNullUntyped(cfg.sqlValueKind), nil
	}
	return NewSQLBool(cfg.sqlValueKind, true), nil
}

func (ni *SQLNotInSubqueryExpr) String() string {
	return fmt.Sprintf("%v not in (%s)", ni.left, PrettyPrintPlan(ni.plan))
}

// EvalType returns the EvalType associated with SQLNotInSubqueryExpr.
func (*SQLNotInSubqueryExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLAnyExpr evaluates to true if the left expression compares true to
// any of the rows returned by the right subquery by a provided comparison
// operator.
// Multi-column right subqueries are valid if the left is a tuple or
// subquery with the same number of columns.
// Multirow left subqueries are never valid.
type SQLAnyExpr struct {
	correlated bool
	left       SQLExpr
	plan       PlanStage
	operator   string
	// We always cache non-correlated subquery results in their entirety.
	// This is a fine place to be more clever in the future.
	// SQLAnyExpr can cache a whole table, with each row being compared
	// to the value result of the left expression.
	cache *SQLValues
}

// NewSQLAnyExpr is a constructor for SQLAnyExpr.
func NewSQLAnyExpr(
	correlated bool,
	left SQLExpr,
	plan PlanStage,
	operator string) *SQLAnyExpr {
	return &SQLAnyExpr{
		correlated: correlated,
		left:       left,
		plan:       plan,
		operator:   operator,
	}
}

// SkipConstantFolding indicates that we should not attempt to
// constant-fold this expression.
func (*SQLAnyExpr) SkipConstantFolding() bool {
	return true
}

// Evaluate evaluates a SQLAnyExpr into a SQLValue.
// ANY performs a series of comparisons. ANY uses the provided comparison operator.
// The resulting comparisons within columns of a row are ANDed together.
// Comparisons from separate rows are ORed together.
// Using SQL three-value boolean logic, the results are as follows:
// If a series of comparisons within any row is all true, the result is false.
// If not, if any of the series returns NULL (the series contains NULL and no falses),
// the result is NULL.
// Else, the result is true.
func (sa *SQLAnyExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var table *SQLValues
	var err error
	if sa.correlated {
		table, err = evaluatePlan(ctx, cfg, st.SubqueryState(), sa.plan)
		if err != nil {
			return nil, err
		}
	} else {
		if sa.cache == nil {
			// Populate cache.
			sa.cache, err = evaluatePlan(ctx, cfg, st, sa.plan)
			if err != nil {
				return nil, err
			}
			err = cfg.memoryMonitor.AcquireGlobal(sa.cache.Size())
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		table = sa.cache
	}

	// Determine length of the left expression.
	var leftLen int
	leftValues, isVals := sa.left.(*SQLValues)
	leftTuple, isTups := sa.left.(*SQLTupleExpr)
	if isVals {
		leftLen = len(leftValues.Values)
	} else if isTups {
		leftLen = len(leftTuple.Exprs)
	} else {
		leftLen = 1
	}

	sawNull := false
	for _, row := range table.Values {
		right := row.(*SQLValues)

		// Make sure the subquery returns the same number of columns as what
		// it's being compared to.
		// Note: This is redundant to do for each row.
		if leftLen != len(right.Values) {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, leftLen)
		}

		var comp SQLExpr
		comp, err = comparisonExpr(sa.left, right, sa.operator)
		if err != nil {
			return nil, err
		}
		var result SQLValue
		result, err = comp.Evaluate(ctx, cfg, st)
		if err != nil {
			return nil, err
		}
		if Bool(result) {
			return NewSQLBool(cfg.sqlValueKind, true), nil
		}
		if result.IsNull() {
			sawNull = true
		}
	}

	// left expression not comparing successfully to any row in the right table
	if sawNull {
		return NewSQLNullUntyped(cfg.sqlValueKind), nil
	}
	return NewSQLBool(cfg.sqlValueKind, false), nil
}

func (sa *SQLAnyExpr) String() string {
	return fmt.Sprintf("%v %s any (%s)", sa.left, sa.operator, PrettyPrintPlan(sa.plan))
}

// EvalType returns the EvalType associated with SQLAnyExpr.
func (*SQLAnyExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLSomeExpr evaluates to true if the left expression compares true to
// any of the rows returned by the right subquery by a provided comparison
// operator.
// Multi-column right subqueries are valid if the left is a tuple or
// subquery with the same number of columns.
// Multirow left subqueries are never valid.
// Note: This should not be. SOME and ANY are aliases of each other.
type SQLSomeExpr struct {
	correlated bool
	left       SQLExpr
	plan       PlanStage
	operator   string
	// We always cache non-correlated subquery results in their entirety.
	// This is a fine place to be more clever in the future.
	// SQLSomeExpr can cache a whole table, with each row being compared
	// to the value result of the left expression.
	cache *SQLValues
}

// NewSQLSomeExpr is a constructor for SQLSomeExpr.
func NewSQLSomeExpr(
	correlated bool,
	left SQLExpr,
	plan PlanStage,
	operator string) *SQLSomeExpr {
	return &SQLSomeExpr{
		correlated: correlated,
		left:       left,
		plan:       plan,
		operator:   operator,
	}
}

// SkipConstantFolding indicates that we should not attempt to
// constant-fold this expression.
func (*SQLSomeExpr) SkipConstantFolding() bool {
	return true
}

// Evaluate evaluates a SQLSomeExpr into a SQLValue.
func (ss *SQLSomeExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var table *SQLValues
	var err error
	if ss.correlated {
		table, err = evaluatePlan(ctx, cfg, st.SubqueryState(), ss.plan)
		if err != nil {
			return nil, err
		}
	} else {
		if ss.cache == nil {
			// Populate cache.
			ss.cache, err = evaluatePlan(ctx, cfg, st, ss.plan)
			if err != nil {
				return nil, err
			}
			err = cfg.memoryMonitor.AcquireGlobal(ss.cache.Size())
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		table = ss.cache
	}

	// Determine length of the left expression.
	var leftLen int
	leftValues, isVals := ss.left.(*SQLValues)
	leftTuple, isTups := ss.left.(*SQLTupleExpr)
	if isVals {
		leftLen = len(leftValues.Values)
	} else if isTups {
		leftLen = len(leftTuple.Exprs)
	} else {
		leftLen = 1
	}

	sawNull := false
	for _, row := range table.Values {
		right := row.(*SQLValues)

		// Make sure the subquery returns the same number of columns as what
		// it's being compared to.
		// Note: This is redundant to do for each row.
		if leftLen != len(right.Values) {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, leftLen)
		}

		var comp SQLExpr
		comp, err = comparisonExpr(ss.left, right, ss.operator)
		if err != nil {
			return nil, err
		}
		var result SQLValue
		result, err = comp.Evaluate(ctx, cfg, st)
		if err != nil {
			return nil, err
		}
		if Bool(result) {
			return NewSQLBool(cfg.sqlValueKind, true), nil
		}
		if result.IsNull() {
			sawNull = true
		}
	}

	// left expression not comparing successfully to any row in the right table
	if sawNull {
		return NewSQLNullUntyped(cfg.sqlValueKind), nil
	}
	return NewSQLBool(cfg.sqlValueKind, false), nil
}

func (ss *SQLSomeExpr) String() string {
	return fmt.Sprintf("%v %s some (%s)", ss.left, ss.operator, PrettyPrintPlan(ss.plan))
}

// EvalType returns the EvalType associated with SQLSomeExpr.
func (*SQLSomeExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLAllExpr evaluates to true if the left expression compares true to
// all of the rows returned by the right subquery by a provided comparison
// operator.
// Multi-column right subqueries are valid if the left is a tuple or
// subquery with the same number of columns.
// Multirow left subqueries are never valid.
type SQLAllExpr struct {
	correlated bool
	left       SQLExpr
	plan       PlanStage
	operator   string
	// We always cache non-correlated subquery results in their entirety.
	// This is a fine place to be more clever in the future.
	// SQLAllExpr can cache a whole table, with each row being compared
	// to the value result of the left expression.
	cache *SQLValues
}

// NewSQLAllExpr is a constructor for SQLAllExpr.
func NewSQLAllExpr(
	correlated bool,
	left SQLExpr,
	plan PlanStage,
	operator string) *SQLAllExpr {
	return &SQLAllExpr{
		correlated: correlated,
		left:       left,
		plan:       plan,
		operator:   operator,
	}
}

// SkipConstantFolding indicates that we should not attempt to
// constant-fold this expression.
func (*SQLAllExpr) SkipConstantFolding() bool {
	return true
}

// Evaluate evaluates a SQLAllExpr into a SQLValue.
func (sa *SQLAllExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var table *SQLValues
	var err error
	if sa.correlated {
		table, err = evaluatePlan(ctx, cfg, st.SubqueryState(), sa.plan)
		if err != nil {
			return nil, err
		}
	} else {
		if sa.cache == nil {
			// Populate cache.
			sa.cache, err = evaluatePlan(ctx, cfg, st, sa.plan)
			if err != nil {
				return nil, err
			}
			err = cfg.memoryMonitor.AcquireGlobal(sa.cache.Size())
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		table = sa.cache
	}

	// Determine length of the left expression.
	var leftLen int
	leftValues, isVals := sa.left.(*SQLValues)
	leftTuple, isTups := sa.left.(*SQLTupleExpr)
	if isVals {
		leftLen = len(leftValues.Values)
	} else if isTups {
		leftLen = len(leftTuple.Exprs)
	} else {
		leftLen = 1
	}

	sawNull := false
	for _, row := range table.Values {
		right := row.(*SQLValues)

		// Make sure the subquery returns the same number of columns as what
		// it's being compared to.
		// Note: This is redundant to do for each row.
		if leftLen != len(right.Values) {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, leftLen)
		}

		var comp SQLExpr
		comp, err = comparisonExpr(sa.left, right, sa.operator)
		if err != nil {
			return nil, err
		}
		var result SQLValue
		result, err = comp.Evaluate(ctx, cfg, st)
		if err != nil {
			return nil, err
		}
		if !Bool(result) {
			return NewSQLBool(cfg.sqlValueKind, false), nil
		}
		if result.IsNull() {
			sawNull = true
		}
	}

	// left expression compared successfully to all rows in the right table
	if sawNull {
		return NewSQLNullUntyped(cfg.sqlValueKind), nil
	}
	return NewSQLBool(cfg.sqlValueKind, true), nil
}

func (sa *SQLAllExpr) String() string {
	return fmt.Sprintf("%v %s all (%s)", sa.left, sa.operator, PrettyPrintPlan(sa.plan))
}

// EvalType returns the EvalType associated with SQLAllExpr.
func (*SQLAllExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLRightSubqueryCmpExpr evaluates to true if the left expression compares true to
// the single row returned by the right subquery by a provided comparison
// operator. The right subquery must be scalar. The left expression is not a subquery.
// See SQLLeftSubqueryCmpExpr and SQLFullSubqueryCmpExpr for representation of other
// cases.
type SQLRightSubqueryCmpExpr struct {
	correlated bool
	left       SQLExpr
	plan       PlanStage
	operator   string
	// We always cache non-correlated subquery results in their entirety.
	// This is a fine place to be more clever in the future.
	// SQLRightSubqueryCmpExpr caches a scalar but it can be multicolumn.
	cache *SQLValues
}

// NewSQLRightSubqueryCmpExpr is a constructor for SQLRightSubqueryCmpExpr.
func NewSQLRightSubqueryCmpExpr(
	correlated bool,
	left SQLExpr,
	plan PlanStage,
	operator string) *SQLRightSubqueryCmpExpr {
	return &SQLRightSubqueryCmpExpr{
		correlated: correlated,
		left:       left,
		plan:       plan,
		operator:   operator,
	}
}

// SkipConstantFolding indicates that we should not attempt to
// constant-fold this expression.
func (*SQLRightSubqueryCmpExpr) SkipConstantFolding() bool {
	return true
}

// Evaluate evaluates a SQLRightSubqueryCmpExpr into a SQLValue.
func (sr *SQLRightSubqueryCmpExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var row *SQLValues
	var err error
	if sr.correlated {
		row, err = evaluatePlanToScalar(ctx, cfg, st.SubqueryState(), sr.plan)
		if err != nil {
			return nil, err
		}
	} else {
		if sr.cache == nil {
			// Populate cache.
			sr.cache, err = evaluatePlanToScalar(ctx, cfg, st, sr.plan)
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		row = sr.cache
	}

	// Determine length of the left expression.
	var leftLen int
	leftValues, isVals := sr.left.(*SQLValues)
	leftTuple, isTups := sr.left.(*SQLTupleExpr)
	if isVals {
		leftLen = len(leftValues.Values)
	} else if isTups {
		leftLen = len(leftTuple.Exprs)
	} else {
		leftLen = 1
	}

	// Make sure the subquery returns the same number of columns as what
	// it's being compared to.
	if leftLen != len(row.Values) {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, leftLen)
	}

	var comp SQLExpr
	comp, err = comparisonExpr(sr.left, row, sr.operator)
	if err != nil {
		return nil, err
	}
	return comp.Evaluate(ctx, cfg, st)
}

func (sr *SQLRightSubqueryCmpExpr) String() string {
	return fmt.Sprintf("%v %s (%s)", sr.left, sr.operator, PrettyPrintPlan(sr.plan))
}

// EvalType returns the EvalType associated with SQLRightSubqueryCmpExpr.
func (*SQLRightSubqueryCmpExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLLeftSubqueryCmpExpr evaluates to true if the right expression compares true to
// the single row returned by the left subquery by a provided comparison
// operator. The left subquery must be scalar. The right expression is not a subquery.
// See SQLRightSubqueryCmpExpr and SQLFullSubqueryCmpExpr for representation of other
// cases.
// Note: This should not be. This can be flipped into a SQLRightSubqueryCmpExpr.
type SQLLeftSubqueryCmpExpr struct {
	correlated bool
	right      SQLExpr
	plan       PlanStage
	operator   string
	// We always cache non-correlated subquery results in their entirety.
	// This is a fine place to be more clever in the future.
	// SQLLeftSubqueryCmpExpr caches a scalar but it can be multicolumn.
	cache *SQLValues
}

// NewSQLLeftSubqueryCmpExpr is a constructor for SQLLeftSubqueryCmpExpr.
func NewSQLLeftSubqueryCmpExpr(
	correlated bool,
	right SQLExpr,
	plan PlanStage,
	operator string) *SQLLeftSubqueryCmpExpr {
	return &SQLLeftSubqueryCmpExpr{
		correlated: correlated,
		right:      right,
		plan:       plan,
		operator:   operator,
	}
}

// SkipConstantFolding indicates that we should not attempt to
// constant-fold this expression.
func (*SQLLeftSubqueryCmpExpr) SkipConstantFolding() bool {
	return true
}

// Evaluate evaluates a SQLLeftSubqueryCmpExpr into a SQLValue.
func (sl *SQLLeftSubqueryCmpExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var row *SQLValues
	var err error
	if sl.correlated {
		row, err = evaluatePlanToScalar(ctx, cfg, st.SubqueryState(), sl.plan)
		if err != nil {
			return nil, err
		}
	} else {
		if sl.cache == nil {
			// Populate cache.
			sl.cache, err = evaluatePlanToScalar(ctx, cfg, st, sl.plan)
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		row = sl.cache
	}

	// Determine length of the right expression.
	var rightLen int
	rightValues, isVals := sl.right.(*SQLValues)
	rightTuple, isTups := sl.right.(*SQLTupleExpr)
	if isVals {
		rightLen = len(rightValues.Values)
	} else if isTups {
		rightLen = len(rightTuple.Exprs)
	} else {
		rightLen = 1
	}

	// Make sure the subquery returns the same number of columns as what
	// it's being compared to.
	if rightLen != len(row.Values) {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, len(row.Values))
	}

	var comp SQLExpr
	comp, err = comparisonExpr(sl.right, row, sl.operator)
	if err != nil {
		return nil, err
	}
	return comp.Evaluate(ctx, cfg, st)
}

func (sl *SQLLeftSubqueryCmpExpr) String() string {
	return fmt.Sprintf("(%s) %s %v", PrettyPrintPlan(sl.plan), sl.operator, sl.right)
}

// EvalType returns the EvalType associated with SQLLeftSubqueryCmpExpr.
func (*SQLLeftSubqueryCmpExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLFullSubqueryCmpExpr evaluates to true if the right subquery compares true to
// the left subquery by a provided comparison operator.
// The left and right subqueries need not be scalar but must produce the same number of rows.
// See SQLRightSubqueryCmpExpr and SQLFullSubqueryCmpExpr for representation of other
// cases.
type SQLFullSubqueryCmpExpr struct {
	leftCorrelated  bool
	rightCorrelated bool
	leftPlan        PlanStage
	rightPlan       PlanStage
	operator        string
	// We always cache non-correlated subquery results in their entirety.
	// This cache is for the left-hand side.
	// SQLFullSubqueryCmpExpr's left cache is scalar but it can be multicolumn.
	leftCache *SQLValues
	// This cache is for the right-hand side.
	// SQLFullSubqueryCmpExpr's right cache is scalar but it can be multicolumn.
	rightCache *SQLValues
	// This cache is for the result. It is used if both sides are non-correlated.
	// This cache consists of a boolean.
	fullCache SQLBool
}

// NewSQLFullSubqueryCmpExpr is a constructor for SQLFullSubqueryCmpExpr.
func NewSQLFullSubqueryCmpExpr(
	leftCorrelated bool,
	rightCorrelated bool,
	leftPlan PlanStage,
	rightPlan PlanStage,
	operator string) *SQLFullSubqueryCmpExpr {
	return &SQLFullSubqueryCmpExpr{
		leftCorrelated:  leftCorrelated,
		rightCorrelated: rightCorrelated,
		leftPlan:        leftPlan,
		rightPlan:       rightPlan,
		operator:        operator,
	}
}

// SkipConstantFolding indicates that we should not attempt to
// constant-fold this expression.
func (*SQLFullSubqueryCmpExpr) SkipConstantFolding() bool {
	return true
}

// Evaluate evaluates a SQLFullSubqueryCmpExpr into a SQLValue.
func (sf *SQLFullSubqueryCmpExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	if !sf.leftCorrelated && !sf.rightCorrelated && sf.fullCache != nil {
		return sf.fullCache, nil
	}

	var leftRow *SQLValues
	var rightRow *SQLValues
	var err error
	if sf.leftCorrelated {
		leftRow, err = evaluatePlanToScalar(ctx, cfg, st.SubqueryState(), sf.leftPlan)
		if err != nil {
			return nil, err
		}
		if sf.rightCache == nil {
			// Populate cache.
			sf.rightCache, err = evaluatePlanToScalar(ctx, cfg, st, sf.rightPlan)
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		rightRow = sf.rightCache
	} else if sf.rightCorrelated {
		rightRow, err = evaluatePlanToScalar(ctx, cfg, st.SubqueryState(), sf.rightPlan)
		if err != nil {
			return nil, err
		}
		if sf.leftCache == nil {
			// Populate cache.
			sf.leftCache, err = evaluatePlanToScalar(ctx, cfg, st, sf.leftPlan)
			if err != nil {
				return nil, err
			}
		}
		// Read from cache.
		leftRow = sf.leftCache
		// Either both sides are correlated or neither are.
	} else {
		leftRow, err = evaluatePlanToScalar(ctx, cfg, st.SubqueryState(), sf.leftPlan)
		if err != nil {
			return nil, err
		}
		rightRow, err = evaluatePlanToScalar(ctx, cfg, st.SubqueryState(), sf.rightPlan)
		if err != nil {
			return nil, err
		}
	}

	// Make sure both subqueres return the same number of columns.
	if len(leftRow.Values) != len(rightRow.Values) {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, len(rightRow.Values))
	}

	var comp SQLExpr
	comp, err = comparisonExpr(leftRow, rightRow, sf.operator)
	if err != nil {
		return nil, err
	}
	var result SQLValue
	result, err = comp.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	// Populate full cache.
	if !sf.leftCorrelated && !sf.rightCorrelated {
		sf.fullCache = result.(SQLBool)
	}

	return result, nil
}

func (sf *SQLFullSubqueryCmpExpr) String() string {
	return fmt.Sprintf("(%s) %s (%s)", PrettyPrintPlan(sf.leftPlan),
		sf.operator, PrettyPrintPlan(sf.rightPlan))
}

// EvalType returns the EvalType associated with SQLFullSubqueryCmpExpr.
func (*SQLFullSubqueryCmpExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLSubqueryExpr is a wrapper around a parser.SelectStatement representing a subquery
// outside of an EXISTS expression. A SQLSubqueryExpr always evaluates to a single-column
// scalar.
type SQLSubqueryExpr struct {
	correlated bool
	allowRows  bool
	plan       PlanStage
	// We always cache non-correlated subquery results in their entirety.
	// This is a fine place to be more clever in the future.
	// SQLSubqueryExpr caches a single-column scalar.
	cache SQLValue
}

// NewSQLSubqueryExpr is a constructor for SQLSubqueryExpr.
func NewSQLSubqueryExpr(correlated, allowRows bool, plan PlanStage) *SQLSubqueryExpr {
	return &SQLSubqueryExpr{
		correlated: correlated,
		allowRows:  allowRows,
		plan:       plan,
	}
}

// ToAggregationLanguage translates SQLSubqueryExpr into something that can
// be used in an aggregation pipeline. If SQLSubqueryExpr cannot be translated,
// it will return nil and error.
func (se *SQLSubqueryExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	if se.correlated {
		return nil, fmt.Errorf("could not pushdown correlated subquery")
	}

	piece := t.addNonCorrelatedSubqueryFuture(se.plan)
	return bsonutil.WrapInLiteral(piece), nil
}

func (se *SQLSubqueryExpr) evaluateFromPlan(ctx context.Context,
	cfg *ExecutionConfig, st *ExecutionState, plan PlanStage) (SQLValue, error) {
	var err error
	var iter Iter
	defer func() {
		if iter != nil {
			if err == nil {
				err = iter.Close()
			} else {
				// If we already have an err, keep the previous err rather
				// than getting a new one.
				_ = iter.Close()
			}

			if err == nil {
				err = iter.Err()
			}
		}
	}()

	iter, err = plan.Open(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	row := &Row{}

	hasNext := iter.Next(ctx, row)
	if hasNext {

		// release this memory here... it will be re-allocated by a consuming stage
		if err = cfg.memoryMonitor.Release(row.Data.Size()); err != nil {
			_ = iter.Close()
			return nil, err
		}

		// Filter has to check the entire source to return an accurate 'hasNext'
		if iter.Next(ctx, &Row{}) {
			_ = iter.Close()
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErSubqueryNoOneRow)
		}
	}

	switch len(row.Data) {
	case 0:
		return NewSQLNull(cfg.sqlValueKind, se.EvalType()), iter.Close()
	case 1:
		return row.Data[0].Data, iter.Close()
	default:
		eval := &SQLValues{}
		for _, value := range row.Data {
			eval.Values = append(eval.Values, value.Data)
		}
		return eval, iter.Close()
	}
}

// Evaluate evaluates a SQLSubqueryExpr into a SQLValue.
func (se *SQLSubqueryExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	if se.correlated {
		return se.evaluateFromPlan(ctx, cfg, st.SubqueryState(), se.plan)
	}

	var err error
	if se.cache == nil {
		// Populate cache.
		se.cache, err = se.evaluateFromPlan(ctx, cfg, st, se.plan)
		if err != nil {
			return nil, err
		}
	}
	// Read from cache.
	return se.cache, nil
}

// Exprs returns all the SQLColumnExprs associated with the columns of SQLSubqueryExpr.
func (se *SQLSubqueryExpr) Exprs() []SQLExpr {
	exprs := []SQLExpr{}
	for _, c := range se.plan.Columns() {
		exprs = append(exprs, NewSQLColumnExpr(c.SelectID,
			c.Database, c.Table, c.Name, c.EvalType, c.MongoType))
	}

	return exprs
}

func (se *SQLSubqueryExpr) String() string {
	return PrettyPrintPlan(se.plan)
}

// EvalType returns the EvalType associated with SQLSubqueryExpr.
func (se *SQLSubqueryExpr) EvalType() EvalType {
	columns := se.plan.Columns()
	if len(columns) == 1 {
		return columns[0].EvalType
	}

	return EvalTuple
}

// SQLSubtractExpr evaluates to the difference of the left expression minus the right expressions.
type SQLSubtractExpr sqlBinaryNode

var _ translatableToAggregation = (*SQLSubtractExpr)(nil)

// Evaluate evaluates a SQLSubtractExpr into a SQLValue.
func (sub *SQLSubtractExpr) Evaluate(ctx context.Context,
	cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	leftVal, err := sub.left.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	rightVal, err := sub.right.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(cfg.sqlValueKind, sub.EvalType()), nil
	}

	return doArithmetic(leftVal, rightVal, SUB)
}

func (sub *SQLSubtractExpr) String() string {
	return fmt.Sprintf("%v-%v", sub.left, sub.right)
}

// ToAggregationLanguage translates SQLSubtractExpr into something that can
// be used in an aggregation pipeline. If SQLSubtractExpr cannot be translated,
// it will return nil and error.
func (sub *SQLSubtractExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	left, err := t.ToAggregationLanguage(sub.left)
	if err != nil {
		return nil, err
	}

	right, err := t.ToAggregationLanguage(sub.right)
	if err != nil {
		return nil, err
	}

	return bson.M{bsonutil.OpSubtract: []interface{}{left, right}}, nil
}

// EvalType returns the EvalType associated with SQLSubtractExpr.
func (sub *SQLSubtractExpr) EvalType() EvalType {
	return EvalDouble
}

// SQLTupleExpr represents a tuple.
type SQLTupleExpr struct {
	Exprs []SQLExpr
}

var _ translatableToAggregation = (*SQLTupleExpr)(nil)

// Evaluate evaluates a SQLTupleExpr into a SQLValue.
func (te SQLTupleExpr) Evaluate(ctx context.Context,
	cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var values []SQLValue

	for _, v := range te.Exprs {
		value, err := v.Evaluate(ctx, cfg, st)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}

	if len(values) == 1 {
		return values[0], nil
	}

	return &SQLValues{values}, nil
}

// Normalize will attempt to change SQLTupleExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (te *SQLTupleExpr) Normalize(kind SQLValueKind) Node {
	if len(te.Exprs) == 1 {
		return te.Exprs[0]
	}

	return te
}

func (te SQLTupleExpr) String() string {
	prefix := "("
	for i, expr := range te.Exprs {
		prefix += fmt.Sprintf("%v", expr.String())
		if i != len(te.Exprs)-1 {
			prefix += ", "
		}
	}
	prefix += ")"
	return prefix
}

// ToAggregationLanguage translates SQLTupleExpr into something that can
// be used in an aggregation pipeline. If SQLTupleExpr cannot be translated,
// it will return nil and error.
func (te *SQLTupleExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	var transExprs []interface{}

	for _, expr := range te.Exprs {
		transExpr, err := t.ToAggregationLanguage(expr)
		if err != nil {
			return nil, err
		}
		transExprs = append(transExprs, transExpr)
	}

	return transExprs, nil

}

// EvalType returns the EvalType associated with SQLTupleExpr.
func (te SQLTupleExpr) EvalType() EvalType {
	if len(te.Exprs) == 1 {
		return te.Exprs[0].EvalType()
	}

	return EvalTuple
}

// SQLUnaryMinusExpr evaluates to the negation of the expression.
type SQLUnaryMinusExpr sqlUnaryNode

var _ translatableToAggregation = (*SQLUnaryMinusExpr)(nil)

// NewSQLUnaryMinusExpr is a constructor for SQLUnaryMinusExpr.
func NewSQLUnaryMinusExpr(operand SQLExpr) *SQLUnaryMinusExpr {
	return &SQLUnaryMinusExpr{operand}
}

// Evaluate evaluates a SQLUnaryMinusExpr into a SQLValue.
func (um *SQLUnaryMinusExpr) Evaluate(ctx context.Context,
	cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	if val, err := um.SQLExpr.Evaluate(ctx, cfg, st); err == nil {
		if val.IsNull() {
			return NewSQLNull(cfg.sqlValueKind, um.EvalType()), nil
		}
		difference := NewSQLFloat(cfg.sqlValueKind, -Float64(val))
		converted := ConvertTo(difference, um.EvalType())
		return converted, nil
	}
	return nil, fmt.Errorf("UnaryMinus expression does not apply to a %T", um.SQLExpr)
}

// Normalize will attempt to change SQLUnaryMinusExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (um *SQLUnaryMinusExpr) Normalize(kind SQLValueKind) Node {
	sqlVal, ok := um.SQLExpr.(SQLValue)
	if !ok {
		return um
	}

	if sqlVal.IsNull() {
		return NewSQLNull(kind, um.EvalType())
	}

	if sqlVal.EvalType() == EvalBoolean {
		if sqlVal.Value().(bool) {
			return NewSQLInt64(kind, -1)
		}
		return NewSQLInt64(kind, 0)
	}

	return um
}

func (um *SQLUnaryMinusExpr) String() string {
	return fmt.Sprintf("-%v", um.SQLExpr)
}

// ToAggregationLanguage translates SQLUnaryMinusExpr into something that can
// be used in an aggregation pipeline. If SQLUnaryMinusExpr cannot be translated,
// it will return nil and error.
func (um *SQLUnaryMinusExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	operand, err := t.ToAggregationLanguage(um.SQLExpr)
	if err != nil {
		return nil, err
	}

	letAssignment := bson.M{
		"operand": operand,
	}

	letEvaluation := bsonutil.WrapInNullCheckedCond(
		nil,
		bson.M{bsonutil.OpMultiply: []interface{}{-1, "$$operand"}},
		"$$operand",
	)

	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil

}

// EvalType returns the EvalType associated with SQLUnaryMinusExpr.
func (um *SQLUnaryMinusExpr) EvalType() EvalType {
	return um.SQLExpr.EvalType()
}

// SQLUnaryTildeExpr invert all bits in the operand
// and returns an unsigned 64-bit integer.
type SQLUnaryTildeExpr sqlUnaryNode

// NewSQLUnaryTildeExpr is a constructor for SQLUnaryTildeExpr.
func NewSQLUnaryTildeExpr(operand SQLExpr) *SQLUnaryTildeExpr {
	return &SQLUnaryTildeExpr{operand}
}

// Evaluate evaluates a SQLUnaryTildeExpr into a SQLValue.
func (td *SQLUnaryTildeExpr) Evaluate(ctx context.Context,
	cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	expr, err := td.SQLExpr.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLBool(cfg.sqlValueKind, false), err
	}

	if v, ok := expr.(SQLValue); ok {
		return NewSQLUint64(cfg.sqlValueKind, ^uint64(Int64(v))), nil
	}

	return NewSQLUint64(cfg.sqlValueKind, ^uint64(0)), nil
}

// Normalize will attempt to change SQLUnaryTildeExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (td *SQLUnaryTildeExpr) Normalize(kind SQLValueKind) Node {
	if v, ok := td.SQLExpr.(SQLValue); ok {
		return NewSQLUint64(kind, ^uint64(Int64(v)))
	}
	return td
}

func (td *SQLUnaryTildeExpr) String() string {
	return fmt.Sprintf("~%v", td.SQLExpr)
}

// EvalType returns the EvalType associated with SQLUnaryTildeExpr.
func (td *SQLUnaryTildeExpr) EvalType() EvalType {
	return td.SQLExpr.EvalType()
}

// SQLVariableExpr represents a variable lookup.
type SQLVariableExpr struct {
	Name    string
	Kind    variable.Kind
	Scope   variable.Scope
	Value   interface{}
	SQLType schema.SQLType
}

// SkipConstantFolding indicates that we should not attempt to
// constant-fold this expression.
func (v *SQLVariableExpr) SkipConstantFolding() bool {
	return true
}

// Evaluate evaluates a SQLVariableExpr into a SQLValue.
func (v *SQLVariableExpr) Evaluate(_ context.Context, cfg *ExecutionConfig, _ *ExecutionState) (SQLValue, error) {
	val := GoValueToSQLValue(cfg.sqlValueKind, v.Value)
	converted := ConvertTo(val, SQLTypeToEvalType(v.SQLType))
	return converted, nil
}

func (v *SQLVariableExpr) String() string {
	prefix := ""
	switch v.Kind {
	case variable.UserKind:
		prefix = "@"
	default:
		switch v.Scope {
		case variable.GlobalScope:
			prefix = "@@global."
		case variable.SessionScope:
			prefix = "@@session."
		}
	}

	return prefix + v.Name
}

// EvalType returns the EvalType associated with SQLVariableExpr.
func (v *SQLVariableExpr) EvalType() EvalType {
	return SQLTypeToEvalType(v.SQLType)
}

// ToAggregationLanguage translates SQLVariableExpr into something that can
// be used in an aggregation pipeline. If SQLVariableExpr cannot be translated,
// it will return nil and error.
func (v *SQLVariableExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {

	e := SQLTypeToEvalType(v.SQLType)
	if e != EvalBoolean {
		return nil, fmt.Errorf("can only pushdown boolean variable types")
	}

	return bsonutil.WrapInLiteral(v.Value), nil
}

// SQLXorExpr evaluates to true if and only if one of its children evaluates to true.
type SQLXorExpr sqlBinaryNode

var _ reconcilingSQLExpr = (*SQLXorExpr)(nil)
var _ translatableToAggregation = (*SQLXorExpr)(nil)

// Evaluate evaluates a SQLXorExpr into a SQLValue.
func (xor *SQLXorExpr) Evaluate(ctx context.Context,
	cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	left, err := xor.left.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	right, err := xor.right.Evaluate(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	if hasNullValue(left, right) {
		return NewSQLNull(cfg.sqlValueKind, xor.EvalType()), nil
	}

	if (IsFalsy(left) && Bool(right)) || (Bool(left) && IsFalsy(right)) {
		return NewSQLBool(cfg.sqlValueKind, true), nil
	}

	return NewSQLBool(cfg.sqlValueKind, false), nil
}

// Normalize will attempt to change SQLXorExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (xor *SQLXorExpr) Normalize(kind SQLValueKind) Node {
	left, leftOk := xor.left.(SQLValue)
	if leftOk {
		if Bool(left) {
			return &SQLNotExpr{xor.right}
		} else if IsFalsy(left) {
			return &SQLOrExpr{NewSQLBool(kind, false), xor.right}
		}
	}

	right, rightOk := xor.right.(SQLValue)
	if rightOk {
		if Bool(right) {
			return &SQLNotExpr{xor.left}
		} else if IsFalsy(right) {
			return &SQLOrExpr{NewSQLBool(kind, false), xor.left}
		}
	}

	return xor
}

func (xor *SQLXorExpr) String() string {
	return fmt.Sprintf("%v xor %v", xor.left, xor.right)
}

func (xor *SQLXorExpr) reconcile() (SQLExpr, error) {
	left := xor.left
	right := xor.right

	if !isBooleanComparable(left.EvalType()) {
		left = NewSQLConvertExpr(left, EvalBoolean)
	}
	if !isBooleanComparable(right.EvalType()) {
		right = NewSQLConvertExpr(right, EvalBoolean)
	}
	return &SQLXorExpr{left, right}, nil
}

// ToAggregationLanguage translates SQLXorExpr into something that can
// be used in an aggregation pipeline. If SQLXorExpr cannot be translated,
// it will return nil and error.
func (xor *SQLXorExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	left, err := t.ToAggregationLanguage(xor.left)
	if err != nil {
		return nil, err
	}

	right, err := t.ToAggregationLanguage(xor.right)
	if err != nil {
		return nil, err
	}

	letAssignment := bson.M{
		"left": left, "right": right,
	}

	letEvaluation := bson.M{
		bsonutil.OpCond: []interface{}{
			bson.M{
				bsonutil.OpOr: []interface{}{
					bson.M{
						bsonutil.OpEq: []interface{}{
							bson.M{
								bsonutil.OpIfNull: []interface{}{"$$left", nil}},
							nil,
						},
					},
					bson.M{
						bsonutil.OpEq: []interface{}{
							bson.M{
								bsonutil.OpIfNull: []interface{}{"$$right", nil}},
							nil,
						},
					},
				},
			},
			mgoNullLiteral,
			bson.M{
				bsonutil.OpAnd: []interface{}{
					bson.M{
						bsonutil.OpOr: []interface{}{"$$left", "$$right"}},
					bson.M{
						bsonutil.OpNot: bson.M{
							bsonutil.OpAnd: []interface{}{"$$left", "$$right"},
						},
					},
				},
			},
		},
	}

	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil

}

// EvalType returns the EvalType associated with SQLXorExpr.
func (*SQLXorExpr) EvalType() EvalType {
	return EvalBoolean
}

// VariableKind indicates if the variable is a system variable or a user variable.
type VariableKind string

// caseCondition holds a matcher used in evaluating case expressions and
// a value to return if a particular case is matched. If a case is matched,
// the corresponding 'then' value is evaluated and returned ('then'
// corresponds to the 'then' clause in a case expression).
type caseCondition struct {
	matcher SQLExpr
	then    SQLExpr
}

func (c *caseCondition) String() string {
	return fmt.Sprintf("when (%v) then %v", c.matcher, c.then)
}

type sqlBinaryNode struct {
	left, right SQLExpr
}

type sqlUnaryNode struct {
	SQLExpr
}

// NewSQLAddExpr is a constructor for SQLAddExpr.
func NewSQLAddExpr(left, right SQLExpr) *SQLAddExpr {
	reconciled := convertAllExprs([]SQLExpr{left, right}, EvalDouble)
	return &SQLAddExpr{reconciled[0], reconciled[1]}
}

// NewSQLColumnExpr creates a new SQLColumnExpr with its required fields.
// NewSQLColumnExpr is a constructor for SQLColumnExpr.
func NewSQLColumnExpr(
	selectID int,
	databaseName,
	tableName,
	columnName string,
	evalType EvalType,
	mongoType schema.MongoType) SQLColumnExpr {
	return SQLColumnExpr{
		selectID:     selectID,
		databaseName: databaseName,
		tableName:    tableName,
		columnName:   columnName,
		columnType: *NewColumnType(
			evalType,
			mongoType,
		),
	}
}

// NewSQLDivideExpr is a constructor for SQLDivideExpr.
func NewSQLDivideExpr(left, right SQLExpr) *SQLDivideExpr {
	reconciled := convertAllExprs([]SQLExpr{left, right}, EvalDouble)
	return &SQLDivideExpr{reconciled[0], reconciled[1]}
}

func isBooleanColumnAndNumber(left, right SQLExpr) bool {
	if _, ok := left.(SQLColumnExpr); !ok {
		return false
	}

	if left.EvalType() != EvalBoolean {
		return false
	}

	if _, ok := right.(SQLNumber); !ok {
		return false
	}

	if _, ok := right.(SQLBool); ok {
		return false
	}

	return true
}

// NewSQLIDivideExpr is a constructor for SQLIDivideExpr.
func NewSQLIDivideExpr(left, right SQLExpr) *SQLIDivideExpr {
	reconciled := convertAllExprs([]SQLExpr{left, right}, EvalDouble)
	return &SQLIDivideExpr{reconciled[0], reconciled[1]}
}

// NewSQLIsExpr is a constructor for SQLIsExpr.
func NewSQLIsExpr(left, right SQLExpr) *SQLIsExpr {
	if right.EvalType() == EvalBoolean {
		leftType := left.EvalType()
		if !(leftType.IsNumeric() || leftType == EvalBoolean) {
			reconciled := convertAllExprs([]SQLExpr{left, right}, EvalBoolean)
			return &SQLIsExpr{reconciled[0], reconciled[1]}
		}
	}
	return &SQLIsExpr{left, right}
}

// NewSQLModExpr is a constructor for SQLModExpr.
func NewSQLModExpr(left, right SQLExpr) *SQLModExpr {
	reconciled := convertAllExprs([]SQLExpr{left, right}, EvalDouble)
	return &SQLModExpr{reconciled[0], reconciled[1]}
}

// NewSQLMultiplyExpr is a constructor for SQLMultiplyExpr.
func NewSQLMultiplyExpr(left, right SQLExpr) *SQLMultiplyExpr {
	reconciled := convertAllExprs([]SQLExpr{left, right}, EvalDouble)
	return &SQLMultiplyExpr{reconciled[0], reconciled[1]}
}

// NewSQLSubtractExpr is a constructor for SQLSubtractExpr.
func NewSQLSubtractExpr(left, right SQLExpr) *SQLSubtractExpr {
	reconciled := convertAllExprs([]SQLExpr{left, right}, EvalDouble)
	return &SQLSubtractExpr{reconciled[0], reconciled[1]}
}

// NewSQLVariableExpr is a constructor for SQLVariableExpr.
func NewSQLVariableExpr(name string, kind variable.Kind,
	scope variable.Scope, sqlType schema.SQLType, value interface{}) *SQLVariableExpr {
	return &SQLVariableExpr{
		Name:    name,
		Kind:    kind,
		Scope:   scope,
		SQLType: sqlType,
		Value:   value,
	}
}
