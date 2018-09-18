package evaluator

import (
	"fmt"
	"math"
	"regexp"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
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
	subqueryAll = iota
	subqueryAny
	subqueryIn
	subqueryNotIn
	subquerySome
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
	Evaluate(*EvalCtx) (SQLValue, error)
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
func (fe *MongoFilterExpr) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return nil, fmt.Errorf("could not evaluate predicate with mongo filter expression")
}

func (fe *MongoFilterExpr) String() string {
	return fmt.Sprintf("%v=%v", fe.column.String(), fe.expr.String())
}

// ToMatchLanguage translates MongoFilterExpr into something that can
// be used in an match expression. If MongoFilterExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original MongoFilterExpr.
func (fe *MongoFilterExpr) ToMatchLanguage(t *PushDownTranslator) (bson.M, SQLExpr) {
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
func (add *SQLAddExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftVal, err := add.left.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	rightVal, err := add.right.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(ctx.valueKind(), add.EvalType()), nil
	}

	return doArithmetic(leftVal, rightVal, ADD)
}

func (add *SQLAddExpr) String() string {
	return fmt.Sprintf("%v+%v", add.left, add.right)
}

// ToAggregationLanguage translates SQLAddExpr into something that can
// be used in an aggregation pipeline. If SQLAddExpr cannot be translated,
// it will return nil and error.
func (add *SQLAddExpr) ToAggregationLanguage(t *PushDownTranslator) (interface{}, error) {
	left, err := t.ToAggregationLanguage(add.left)
	if err != nil {
		return nil, err
	}

	right, err := t.ToAggregationLanguage(add.right)
	if err != nil {
		return nil, err
	}

	return bson.M{mgoOperatorAdd: eatChildren(mgoOperatorAdd, left, right)}, nil
}

// EvalType returns the EvalType associated with SQLAddExpr.
func (add *SQLAddExpr) EvalType() EvalType {
	return EvalDouble
}

// SQLAndExpr evaluates to true if and only if all its children evaluate to true.
type SQLAndExpr sqlBinaryNode

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
func (and *SQLAndExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	left, err := and.left.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	right, err := and.right.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if IsFalsy(left) || IsFalsy(right) {
		return NewSQLBool(ctx.valueKind(), false), nil
	}

	if hasNullValue(left, right) {
		return NewSQLNull(ctx.valueKind(), and.EvalType()), nil
	}

	return NewSQLBool(ctx.valueKind(), true), nil
}

// Normalize will attempt to change SQLAndExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (and *SQLAndExpr) Normalize(ctx *EvalCtx) Node {
	left, leftOk := and.left.(SQLValue)
	if leftOk && IsFalsy(left) {
		return NewSQLBool(ctx.valueKind(), false)
	} else if leftOk && Bool(left) {
		if and.right.EvalType() == EvalBoolean {
			return and.right
		}
		return NewSQLConvertExpr(and.right, EvalBoolean)
	}

	right, rightOk := and.right.(SQLValue)
	if rightOk && IsFalsy(right) {
		return NewSQLBool(ctx.valueKind(), false)
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

// ToAggregationLanguage translates SQLAndExpr into something that can
// be used in an aggregation pipeline. If SQLAndExpr cannot be translated,
// it will return nil and error.
func (and *SQLAndExpr) ToAggregationLanguage(t *PushDownTranslator) (interface{}, error) {

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
		mgoOperatorCond: []interface{}{
			bson.M{
				mgoOperatorOr: []interface{}{
					bson.M{
						mgoOperatorEq: []interface{}{
							bson.M{
								mgoOperatorIfNull: []interface{}{"$$left", nil}},
							nil,
						},
					},
					bson.M{
						mgoOperatorEq: []interface{}{
							bson.M{
								mgoOperatorIfNull: []interface{}{"$$right", nil}},
							nil,
						},
					},
				},
			},
			bson.M{
				mgoOperatorCond: []interface{}{
					bson.M{
						mgoOperatorOr: []interface{}{
							bson.M{
								mgoOperatorEq: []interface{}{"$$left", false}},
							bson.M{
								mgoOperatorEq: []interface{}{"$$right", false}},
							bson.M{
								mgoOperatorEq: []interface{}{"$$left", 0}},
							bson.M{
								mgoOperatorEq: []interface{}{"$$right", 0}},
						},
					},
					bson.M{
						mgoOperatorAnd: []interface{}{"$$left", "$$right"},
					},
					mgoNullLiteral,
				},
			},
			bson.M{
				mgoOperatorAnd: []interface{}{"$$left", "$$right"},
			},
		},
	}

	return wrapInLet(letAssignment, letEvaluation), nil

}

// ToMatchLanguage translates SQLAndExpr into something that can
// be used in an match expression. If SQLAndExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLAndExpr.
func (and *SQLAndExpr) ToMatchLanguage(t *PushDownTranslator) (bson.M, SQLExpr) {
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
		if v, ok := left[mgoOperatorAnd]; ok {
			array := v.([]interface{})
			cond = append(cond, array...)
		} else {
			cond = append(cond, left)
		}

		if v, ok := right[mgoOperatorAnd]; ok {
			array := v.([]interface{})
			cond = append(cond, array...)
		} else {
			cond = append(cond, right)
		}

		match = bson.M{mgoOperatorAnd: cond}
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
func (e *SQLAssignmentExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	value, err := e.expr.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	var literal interface{}
	if !value.IsNull() {
		literal = value.Value()
	}

	err = ctx.Variables().Set(variable.Name(e.variable.Name),
		e.variable.Scope,
		e.variable.Kind,
		literal)
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

// Evaluate evaluates a SQLBenchmarkExpr into a SQLValue.
func (e SQLBenchmarkExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	count, err := e.count.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	replaced, err := replaceMongoSourceStages(e.expr, ctx)
	if err != nil {
		return nil, err
	}

	for i := int64(0); i < Int64(count); i++ {
		_, err := replaced.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
	}

	return NewSQLInt64(ctx.valueKind(), 0), nil
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
func (e SQLCaseExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	for _, condition := range e.caseConditions {
		result, err := condition.matcher.Evaluate(ctx)
		if err != nil {
			return nil, err
		}

		if Bool(result) {
			return condition.then.Evaluate(ctx)
		}
	}

	return e.elseValue.Evaluate(ctx)
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
func (e *SQLCaseExpr) ToAggregationLanguage(t *PushDownTranslator) (interface{}, error) {
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
		cases = wrapInCond(thens[i], cases, conditions[i])
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
}

var _ translatableToAggregation = (*SQLColumnExpr)(nil)
var _ translatableToMatch = (*SQLColumnExpr)(nil)

// Evaluate evaluates a SQLColumnExpr into a SQLValue.
func (c SQLColumnExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	// first check our immediate rows
	for _, row := range ctx.Rows {
		if value, ok := row.GetField(c.selectID, c.databaseName, c.tableName, c.columnName); ok {
			return ConvertTo(value, c.EvalType()), nil
		}
	}

	// If we didn't find it there, search in the src rows, which contain parent
	// information in the case we are evaluating a correlated column.
	if ctx.ExecutionCtx != nil {
		for _, row := range ctx.ExecutionCtx.SrcRows {
			if value, ok := row.GetField(c.selectID,
				c.databaseName,
				c.tableName,
				c.columnName); ok {
				return ConvertTo(value, c.EvalType()), nil
			}
		}
	}

	return NewSQLNull(ctx.valueKind(), c.EvalType()), nil
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
func (c SQLColumnExpr) ToAggregationLanguage(t *PushDownTranslator) (interface{}, error) {

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
func (c SQLColumnExpr) ToMatchLanguage(t *PushDownTranslator) (bson.M, SQLExpr) {
	name, ok := t.LookupFieldName(c.databaseName, c.tableName, c.columnName)
	if !ok {
		return nil, c
	}

	if c.EvalType() != EvalBoolean {
		return bson.M{
			name: bson.M{
				mgoOperatorNeq: nil,
			},
		}, c
	}

	return bson.M{
		mgoOperatorAnd: []interface{}{
			bson.M{
				name: bson.M{
					mgoOperatorNeq: false,
				},
			},
			bson.M{
				name: bson.M{
					mgoOperatorNeq: nil,
				},
			},
			bson.M{
				name: bson.M{
					mgoOperatorNeq: 0,
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
func NewSQLConvertExpr(expr SQLExpr, convType EvalType) *SQLConvertExpr {
	return &SQLConvertExpr{
		expr:       expr,
		targetType: convType,
	}
}

// Evaluate evaluates a SQLConvertExpr into a SQLValue.
func (ce *SQLConvertExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	// collapse nested SQLConvertExprs
	if sce, ok := ce.expr.(*SQLConvertExpr); ok {
		ce.expr = sce.expr
	}

	v, err := ce.expr.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	mode := ctx.Variables().GetString(variable.TypeConversionMode)
	switch mode {
	case variable.MongoSQLTypeConversionMode, variable.MySQLTypeConversionMode:
		// for now, handle these cases the same way
	default:
		panic(fmt.Errorf("invalid value %q for type_conversion_mode", mode))
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

func (ce *SQLConvertExpr) translateMongoSQL(t *PushDownTranslator) (interface{}, error) {
	if !t.Ctx.VersionAtLeast(4, 0, 0) {
		return nil, fmt.Errorf("mongosql mode convert cannot be pushed" +
			" down on MongoDB versions < 4.0")
	}

	expr, err := t.ToAggregationLanguage(ce.expr)
	if err != nil {
		return nil, err
	}

	converted := wrapInConvert(expr, ce.expr.EvalType(), ce.targetType)
	return converted, nil
}

func (ce *SQLConvertExpr) translateMySQL(t *PushDownTranslator) (interface{}, error) {
	//
	// the following type conversions are pushed down:
	//
	//     any numeric type -> any numeric type
	//     any numeric type -> string
	//     datetime         -> date
	//     datetime         -> string
	//     datetime         -> any numeric type
	//     date             -> datetime
	//     date             -> string
	//     date             -> any numeric type
	//

	if !t.Ctx.VersionAtLeast(3, 6, 0) {
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
		EvalDecimal128:

		switch toType {
		case EvalInt32, EvalInt64,
			EvalUint32, EvalUint64,
			EvalDecimal128, EvalDouble,
			EvalString:
			return ce.translateMongoSQL(t)
		}

	case EvalDouble:
		switch toType {
		case EvalInt32, EvalInt64,
			EvalUint32, EvalUint64,
			EvalDecimal128:
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

		case EvalInt32, EvalInt64,
			EvalUint32, EvalUint64,
			EvalDecimal128, EvalDouble:

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

	default:
		// mysql-mode pushdown not yet implemented for conversions from other types
	}

	return nil, fmt.Errorf("mysql conversion cannot be pushdown with from type '%s'",
		EvalTypeToMongoType(fromType))
}

// ToAggregationLanguage translates SQLConvertExpr into something that can
// be used in an aggregation pipeline. At the moment, SQLConvertExpr cannot be
// translated, so this function will always return nil and error.
func (ce *SQLConvertExpr) ToAggregationLanguage(t *PushDownTranslator) (interface{}, error) {
	mode := t.Ctx.Variables().GetString(variable.TypeConversionMode)
	switch mode {
	case variable.MySQLTypeConversionMode:
		return ce.translateMySQL(t)
	case variable.MongoSQLTypeConversionMode:
		return ce.translateMongoSQL(t)
	default:
		panic(fmt.Errorf("impossible value %q for type_conversion_mode", mode))
	}
}

// SQLDivideExpr evaluates to the quotient of the left expression divided by the right.
type SQLDivideExpr sqlBinaryNode

var _ translatableToAggregation = (*SQLDivideExpr)(nil)

// Evaluate evaluates a SQLDivideExpr into a SQLValue.
func (div *SQLDivideExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftVal, err := div.left.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	rightVal, err := div.right.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	if Float64(rightVal) == 0 || hasNullValue(leftVal, rightVal) {
		return NewSQLNull(ctx.valueKind(), div.EvalType()), nil
	}

	return doArithmetic(leftVal, rightVal, DIV)
}

// Normalize will attempt to change SQLDivideExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (div *SQLDivideExpr) Normalize(ctx *EvalCtx) Node {
	if hasNullExpr(div.left, div.right) {
		return NewSQLNull(ctx.valueKind(), div.EvalType())
	}
	return div
}

func (div *SQLDivideExpr) String() string {
	return fmt.Sprintf("%v/%v", div.left, div.right)
}

// ToAggregationLanguage translates SQLDivideExpr into something that can
// be used in an aggregation pipeline. If SQLDivideExpr cannot be translated,
// it will return nil and error.
func (div *SQLDivideExpr) ToAggregationLanguage(t *PushDownTranslator) (interface{}, error) {
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

	letEvaluation := wrapInCond(
		nil,
		bson.M{mgoOperatorDivide: []interface{}{"$$left", "$$right"}},
		bson.M{mgoOperatorEq: []interface{}{"$$right", 0}},
	)

	return wrapInLet(letAssignment, letEvaluation), nil

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
func (eq *SQLEqualsExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftVal, err := eq.left.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	rightVal, err := eq.right.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(ctx.valueKind(), eq.EvalType()), nil
	}

	c, err := CompareTo(leftVal, rightVal, ctx.Collation)
	if err == nil {
		return NewSQLBool(ctx.valueKind(), c == 0), nil
	}

	return NewSQLBool(ctx.valueKind(), false), err
}

// Normalize will attempt to change SQLEqualsExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (eq *SQLEqualsExpr) Normalize(ctx *EvalCtx) Node {
	if hasNullExpr(eq.left, eq.right) {
		return NewSQLNull(ctx.valueKind(), eq.EvalType())
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
func (eq *SQLEqualsExpr) ToAggregationLanguage(t *PushDownTranslator) (interface{}, error) {
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

	letEvaluation := wrapInNullCheckedCond(
		nil,
		bson.M{
			mgoOperatorEq: []interface{}{"$$left", "$$right"},
		},
		"$$left",
		"$$right",
	)

	return wrapInLet(letAssignment, letEvaluation), nil
}

// ToMatchLanguage translates SQLEqualsExpr into something that can
// be used in an match expression. If SQLEqualsExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLEqualsExpr.
func (eq *SQLEqualsExpr) ToMatchLanguage(t *PushDownTranslator) (bson.M, SQLExpr) {
	match, ok := t.translateOperator(mgoOperatorEq, eq.left, eq.right)
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
		left, right, err = ReconcileSQLExprs(eq.left, eq.right)
	}

	return &SQLEqualsExpr{left, right}, err
}

// SQLExistsExpr evaluates to true if any result is returned from the subquery.
type SQLExistsExpr struct {
	expr *SQLSubqueryExpr
}

// NewSQLExistsExpr is a constructor for SQLExistsExpr.
func NewSQLExistsExpr(expr *SQLSubqueryExpr) *SQLExistsExpr {
	return &SQLExistsExpr{
		expr: expr,
	}
}

// Evaluate evaluates a SQLExistsExpr into a SQLValue.
func (em *SQLExistsExpr) Evaluate(ctx *EvalCtx) (value SQLValue, err error) {
	var it Iter
	var matches bool

	defer func() {
		if it != nil && err == nil {
			err = it.Err()
		}
	}()

	execCtx := ctx.ExecutionCtx

	if em.expr.correlated {
		execCtx = ctx.CreateChildExecutionCtx()
	}

	it, err = em.expr.plan.Open(execCtx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	if it.Next(&Row{}) {
		matches = true
	}

	return NewSQLBool(ctx.valueKind(), matches), it.Close()
}

func (em *SQLExistsExpr) String() string {
	return fmt.Sprintf("exists(%s)", em.expr.String())
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
func (gt *SQLGreaterThanExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftVal, err := gt.left.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	rightVal, err := gt.right.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(ctx.valueKind(), gt.EvalType()), nil
	}

	c, err := CompareTo(leftVal, rightVal, ctx.Collation)
	if err == nil {
		return NewSQLBool(ctx.valueKind(), c > 0), nil
	}
	return NewSQLBool(ctx.valueKind(), false), err
}

// Normalize will attempt to change SQLGreaterThanExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (gt *SQLGreaterThanExpr) Normalize(ctx *EvalCtx) Node {
	if hasNullExpr(gt.left, gt.right) {
		return NewSQLNull(ctx.valueKind(), gt.EvalType())
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
func (gt *SQLGreaterThanExpr) ToAggregationLanguage(t *PushDownTranslator) (interface{}, error) {
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

	letEvaluation := wrapInNullCheckedCond(
		nil,
		bson.M{
			mgoOperatorGt: []interface{}{"$$left", "$$right"},
		},
		"$$left",
		"$$right",
	)

	return wrapInLet(letAssignment, letEvaluation), nil

}

// ToMatchLanguage translates SQLGreaterThanExpr into something that can
// be used in an match expression. If SQLGreaterThanExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLGreaterThanExpr.
func (gt *SQLGreaterThanExpr) ToMatchLanguage(t *PushDownTranslator) (bson.M, SQLExpr) {
	match, ok := t.translateOperator(mgoOperatorGt, gt.left, gt.right)
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
func (gte *SQLGreaterThanOrEqualExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftVal, err := gte.left.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	rightVal, err := gte.right.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(ctx.valueKind(), gte.EvalType()), nil
	}

	c, err := CompareTo(leftVal, rightVal, ctx.Collation)
	if err == nil {
		return NewSQLBool(ctx.valueKind(), c >= 0), nil
	}

	return NewSQLBool(ctx.valueKind(), false), err
}

// Normalize will attempt to change SQLGreaterThanOrEqualExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (gte *SQLGreaterThanOrEqualExpr) Normalize(ctx *EvalCtx) Node {
	if hasNullExpr(gte.left, gte.right) {
		return NewSQLNull(ctx.valueKind(), gte.EvalType())
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
	t *PushDownTranslator) (interface{}, error) {
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

	letEvaluation := wrapInNullCheckedCond(
		nil,
		bson.M{
			mgoOperatorGte: []interface{}{"$$left", "$$right"},
		},
		"$$left",
		"$$right",
	)

	return wrapInLet(letAssignment, letEvaluation), nil

}

// ToMatchLanguage translates SQLGreaterThanOrEqualExpr into something that can
// be used in an match expression. If SQLGreaterThanOrEqualExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLGreaterThanOrEqualExpr.
func (gte *SQLGreaterThanOrEqualExpr) ToMatchLanguage(t *PushDownTranslator) (bson.M, SQLExpr) {
	match, ok := t.translateOperator(mgoOperatorGte, gte.left, gte.right)
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
func (div *SQLIDivideExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftVal, err := div.left.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	rightVal, err := div.right.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	frightVal := Float64(rightVal)
	if frightVal == 0.0 || hasNullValue(leftVal, rightVal) {
		// NOTE: this is per the mysql manual.
		return NewSQLNull(ctx.valueKind(), div.EvalType()), nil
	}

	return NewSQLInt64(ctx.valueKind(), int64(Float64(leftVal)/frightVal)), nil
}

// Normalize will attempt to change SQLIDivideExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (div *SQLIDivideExpr) Normalize(ctx *EvalCtx) Node {
	if hasNullExpr(div.left, div.right) {
		return NewSQLNull(ctx.valueKind(), div.EvalType())
	}
	return div
}

func (div *SQLIDivideExpr) String() string {
	return fmt.Sprintf("%v/%v", div.left, div.right)
}

// ToAggregationLanguage translates SQLIDivideExpr into something that can
// be used in an aggregation pipeline. If SQLIDivideExpr cannot be translated,
// it will return nil and error.
func (div *SQLIDivideExpr) ToAggregationLanguage(t *PushDownTranslator) (interface{}, error) {
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

	letEvaluation := wrapInCond(
		nil,
		bson.M{
			"$trunc": []interface{}{
				bson.M{
					mgoOperatorDivide: []interface{}{"$$left", "$$right"},
				},
			},
		},
		bson.M{mgoOperatorEq: []interface{}{"$$right", 0}},
	)

	return wrapInLet(letAssignment, letEvaluation), nil

}

// EvalType returns the EvalType associated with SQLIDivideExpr.
func (div *SQLIDivideExpr) EvalType() EvalType {
	return preferentialType(div.left, div.right)
}

// SQLInExpr evaluates to true if the left is in any of the values on the right.
type SQLInExpr sqlBinaryNode

var _ translatableToAggregation = (*SQLInExpr)(nil)

// Evaluate evaluates a SQLInExpr into a SQLValue.
func (in *SQLInExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	left, err := in.left.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	right, err := in.right.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	// right child must be of type SQLValues
	// TODO: can we not simply require this as part of the Node definition?
	rightChild, ok := right.(*SQLValues)
	if !ok {
		child, typeOk := right.(SQLValue)
		if !typeOk {
			return NewSQLBool(ctx.valueKind(), false),
				fmt.Errorf("right 'in' expression is type %T - expected tuple",
					right)
		}
		rightChild = &SQLValues{[]SQLValue{child}}
	}

	leftChild, ok := left.(*SQLValues)
	if ok {
		if len(leftChild.Values) != 1 {
			return NewSQLBool(ctx.valueKind(), false),
				fmt.Errorf("left operand should contain 1 column - got %v",
					len(leftChild.Values))
		}
		left = leftChild.Values[0]
	} else if left.IsNull() {
		return NewSQLNull(ctx.valueKind(), in.EvalType()), nil
	}

	nullInValues := false
	for _, right := range rightChild.Values {
		if right.IsNull() {
			nullInValues = true
		}
		eq := &SQLEqualsExpr{left, right}
		result, err := eq.Evaluate(ctx)
		if err != nil {
			return NewSQLBool(ctx.valueKind(), false), err
		}
		if Bool(result) {
			return NewSQLBool(ctx.valueKind(), true), nil
		}
	}

	if nullInValues {
		return NewSQLNull(ctx.valueKind(), in.EvalType()), nil
	}

	return NewSQLBool(ctx.valueKind(), false), nil
}

// Normalize will attempt to change SQLInExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (in *SQLInExpr) Normalize(ctx *EvalCtx) Node {
	if hasNullExpr(in.left) {
		return NewSQLNull(ctx.valueKind(), in.EvalType())
	}

	return in
}

func (in *SQLInExpr) String() string {
	return fmt.Sprintf("%v in %v", in.left, in.right)
}

// ToAggregationLanguage translates SQLInExpr into something that can
// be used in an aggregation pipeline. If SQLInExpr cannot be translated,
// it will return nil and error.
func (in *SQLInExpr) ToAggregationLanguage(t *PushDownTranslator) (interface{}, error) {
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

	return wrapInNullCheckedCond(
		nil,
		wrapInCond(
			true,
			wrapInCond(
				nil,
				false,
				bson.M{mgoOperatorEq: []interface{}{nullInValues, true}},
			),
			bson.M{mgoOperatorGt: []interface{}{
				bson.M{mgoOperatorSize: bson.M{mgoOperatorFilter: bson.M{"input": right,
					"as":   "item",
					"cond": bson.M{mgoOperatorEq: []interface{}{"$$item", left}},
				}}},
				wrapInLiteral(0),
			}}),
		left,
	), nil

}

// ToMatchLanguage translates SQLInExpr into something that can
// be used in an match expression. If SQLInExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLInExpr.
func (in *SQLInExpr) ToMatchLanguage(t *PushDownTranslator) (bson.M, SQLExpr) {
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

	return bson.M{name: bson.M{mgoOperatorIn: values}}, nil
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
func (is *SQLIsExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftVal, err := is.left.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	rightVal, err := is.right.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	if leftVal.IsNull() {
		if _, ok := rightVal.(SQLBool); ok {
			return NewSQLBool(ctx.valueKind(), false), nil
		}
		return NewSQLBool(ctx.valueKind(), true), nil
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLBool(ctx.valueKind(), false), nil
	}

	if Bool(leftVal) && Bool(rightVal) || IsFalsy(leftVal) && IsFalsy(rightVal) {
		return NewSQLBool(ctx.valueKind(), true), nil
	}

	return NewSQLBool(ctx.valueKind(), false), nil

}

func (is *SQLIsExpr) String() string {
	return fmt.Sprintf("%v is %v", is.left, is.right)
}

// ToAggregationLanguage translates SQLIsExpr into something that can
// be used in an aggregation pipeline. If SQLIsExpr cannot be translated,
// it will return nil and error.
func (is *SQLIsExpr) ToAggregationLanguage(t *PushDownTranslator) (interface{}, error) {
	left, err := t.ToAggregationLanguage(is.left)
	if err != nil {
		return nil, err
	}

	// if right side is {null,unknown}, it's a simple case
	sqlVal, ok := is.right.(SQLValue)
	if ok && sqlVal.IsNull() {
		return wrapInOp(mgoOperatorLte,
			left,
			wrapInLiteral(nil),
		), nil
	}

	right, err := t.ToAggregationLanguage(is.right)
	if err != nil {
		return nil, err
	}

	// if left side is a boolean, this is still simple
	if is.left.EvalType() == EvalBoolean {
		return wrapInOp(mgoOperatorEq,
			left,
			right,
		), nil
	}

	// otherwise, left side is a number type
	if is.right == NewSQLBool(t.valueKind(), true) {
		return wrapInCond(
			false,
			wrapInOp(mgoOperatorNeq,
				left,
				0,
			),
			wrapInNullCheck(left),
		), nil
	} else if is.right == NewSQLBool(t.valueKind(), false) {
		return wrapInOp(mgoOperatorEq,
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
func (is *SQLIsExpr) ToMatchLanguage(t *PushDownTranslator) (bson.M, SQLExpr) {
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
			mgoOperatorAnd: []interface{}{
				bson.M{name: bson.M{mgoOperatorNeq: 0}},
				bson.M{name: bson.M{mgoOperatorNeq: nil}},
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
func (lt *SQLLessThanExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftVal, err := lt.left.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	rightVal, err := lt.right.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(ctx.valueKind(), lt.EvalType()), nil
	}

	c, err := CompareTo(leftVal, rightVal, ctx.Collation)
	if err == nil {
		return NewSQLBool(ctx.valueKind(), c < 0), nil
	}
	return NewSQLBool(ctx.valueKind(), false), err
}

// Normalize will attempt to change SQLLessThanExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (lt *SQLLessThanExpr) Normalize(ctx *EvalCtx) Node {
	if hasNullExpr(lt.left, lt.right) {
		return NewSQLNull(ctx.valueKind(), lt.EvalType())
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
func (lt *SQLLessThanExpr) ToAggregationLanguage(t *PushDownTranslator) (interface{}, error) {
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

	letEvaluation := wrapInNullCheckedCond(
		nil,
		bson.M{
			mgoOperatorLt: []interface{}{"$$left", "$$right"},
		},
		"$$left", "$$right",
	)

	return wrapInLet(letAssignment, letEvaluation), nil

}

// ToMatchLanguage translates SQLLessThanExpr into something that can
// be used in an match expression. If SQLLessThanExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLLessThanExpr.
func (lt *SQLLessThanExpr) ToMatchLanguage(t *PushDownTranslator) (bson.M, SQLExpr) {
	match, ok := t.translateOperator(mgoOperatorLt, lt.left, lt.right)
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
func (lte *SQLLessThanOrEqualExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftVal, err := lte.left.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	rightVal, err := lte.right.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(ctx.valueKind(), lte.EvalType()), nil
	}

	c, err := CompareTo(leftVal, rightVal, ctx.Collation)
	if err == nil {
		return NewSQLBool(ctx.valueKind(), c <= 0), nil
	}
	return NewSQLBool(ctx.valueKind(), false), err
}

// Normalize will attempt to change SQLLessThanOrEqualExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (lte *SQLLessThanOrEqualExpr) Normalize(ctx *EvalCtx) Node {
	if hasNullExpr(lte.left, lte.right) {
		return NewSQLNull(ctx.valueKind(), lte.EvalType())
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
	t *PushDownTranslator) (interface{}, error) {
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

	letEvaluation := wrapInNullCheckedCond(
		nil,
		bson.M{
			mgoOperatorLte: []interface{}{"$$left", "$$right"},
		},
		"$$left", "$$right",
	)

	return wrapInLet(letAssignment, letEvaluation), nil

}

// ToMatchLanguage translates SQLLessThanOrEqualExpr into something that can
// be used in an match expression. If SQLLessThanOrEqualExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLLessThanOrEqualExpr.
func (lte *SQLLessThanOrEqualExpr) ToMatchLanguage(t *PushDownTranslator) (bson.M, SQLExpr) {
	match, ok := t.translateOperator(mgoOperatorLte, lte.left, lte.right)
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
func (l *SQLLikeExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	value, err := l.left.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if hasNullValue(value) {
		return NewSQLNull(ctx.valueKind(), l.EvalType()), nil
	}

	data := value.String()

	value, err = l.right.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if hasNullValue(value) {
		return NewSQLNull(ctx.valueKind(), l.EvalType()), nil
	}

	escape, err := l.escape.Evaluate(ctx)
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

	return NewSQLBool(ctx.valueKind(), matches), nil
}

// Normalize will attempt to change SQLLikeExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (l *SQLLikeExpr) Normalize(ctx *EvalCtx) Node {
	if right, ok := l.right.(SQLValue); ok {
		if hasNullValue(right) {
			return NewSQLNull(ctx.valueKind(), l.EvalType())
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
func (l *SQLLikeExpr) ToMatchLanguage(t *PushDownTranslator) (bson.M, SQLExpr) {
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

	return bson.M{name: bson.M{mgoOperatorRegex: bson.RegEx{Pattern: pattern, Options: opts}}}, nil
}

// EvalType returns the EvalType associated with SQLLikeExpr.
func (*SQLLikeExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLModExpr evaluates the modulus of two expressions
type SQLModExpr sqlBinaryNode

var _ translatableToAggregation = (*SQLModExpr)(nil)

// Evaluate evaluates a SQLModExpr into a SQLValue.
func (mod *SQLModExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftVal, err := mod.left.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	rightVal, err := mod.right.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	frightVal := Float64(rightVal)
	if math.Abs(frightVal) == 0.0 || hasNullValue(leftVal, rightVal) {
		return NewSQLNull(ctx.valueKind(), mod.EvalType()), nil
	}

	modVal := math.Mod(Float64(leftVal), frightVal)
	if modVal == -0 {
		modVal *= -1
	}

	return NewSQLFloat(ctx.valueKind(), modVal), nil
}

func (mod *SQLModExpr) String() string {
	return fmt.Sprintf("%v/%v", mod.left, mod.right)
}

// ToAggregationLanguage translates SQLModExpr into something that can
// be used in an aggregation pipeline. If SQLModExpr cannot be translated,
// it will return nil and error.
func (mod *SQLModExpr) ToAggregationLanguage(t *PushDownTranslator) (interface{}, error) {
	left, err := t.ToAggregationLanguage(mod.left)
	if err != nil {
		return nil, err
	}

	right, err := t.ToAggregationLanguage(mod.right)
	if err != nil {
		return nil, err
	}

	return bson.M{mgoOperatorMod: []interface{}{left, right}}, nil

}

// EvalType returns the EvalType associated with SQLModExpr.
func (mod *SQLModExpr) EvalType() EvalType {
	return preferentialType(mod.left, mod.right)
}

// SQLMultiplyExpr evaluates to the product of two expressions
type SQLMultiplyExpr sqlBinaryNode

var _ translatableToAggregation = (*SQLMultiplyExpr)(nil)

// Evaluate evaluates a SQLMultiplyExpr into a SQLValue.
func (mult *SQLMultiplyExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftVal, err := mult.left.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	rightVal, err := mult.right.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(ctx.valueKind(), mult.EvalType()), nil
	}

	return doArithmetic(leftVal, rightVal, MULT)
}

func (mult *SQLMultiplyExpr) String() string {
	return fmt.Sprintf("%v*%v", mult.left, mult.right)
}

// ToAggregationLanguage translates SQLMultiplyExpr into something that can
// be used in an aggregation pipeline. If SQLMultiplyExpr cannot be translated,
// it will return nil and error.
func (mult *SQLMultiplyExpr) ToAggregationLanguage(t *PushDownTranslator) (interface{}, error) {
	left, err := t.ToAggregationLanguage(mult.left)
	if err != nil {
		return nil, err
	}

	right, err := t.ToAggregationLanguage(mult.right)
	if err != nil {
		return nil, err
	}

	return bson.M{mgoOperatorMultiply: eatChildren(mgoOperatorMultiply, left, right)}, nil
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
func (neq *SQLNotEqualsExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftVal, err := neq.left.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	rightVal, err := neq.right.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(ctx.valueKind(), neq.EvalType()), nil
	}

	c, err := CompareTo(leftVal, rightVal, ctx.Collation)
	if err == nil {
		return NewSQLBool(ctx.valueKind(), c != 0), nil
	}

	return NewSQLBool(ctx.valueKind(), false), err
}

// Normalize will attempt to change SQLNotEqualsExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (neq *SQLNotEqualsExpr) Normalize(ctx *EvalCtx) Node {
	if hasNullExpr(neq.left, neq.right) {
		return NewSQLNull(ctx.valueKind(), neq.EvalType())
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
func (neq *SQLNotEqualsExpr) ToAggregationLanguage(t *PushDownTranslator) (interface{}, error) {
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

	letEvaluation := wrapInNullCheckedCond(
		nil,
		bson.M{
			mgoOperatorNeq: []interface{}{"$$left", "$$right"},
		},
		"$$left", "$$right",
	)

	return wrapInLet(letAssignment, letEvaluation), nil

}

// ToMatchLanguage translates SQLNotEqualsExpr into something that can
// be used in an match expression. If SQLNotEqualsExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLNotEqualsExpr.
func (neq *SQLNotEqualsExpr) ToMatchLanguage(t *PushDownTranslator) (bson.M, SQLExpr) {
	match, ok := t.translateOperator(mgoOperatorNeq, neq.left, neq.right)
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
			mgoOperatorAnd: []interface{}{match,
				bson.M{name: bson.M{mgoOperatorNeq: nil}},
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

var _ translatableToAggregation = (*SQLNotExpr)(nil)
var _ translatableToMatch = (*SQLNotExpr)(nil)

// NewSQLNotExpr is a constructor for SQLNotExpr.
func NewSQLNotExpr(operand SQLExpr) *SQLNotExpr {
	return &SQLNotExpr{operand}
}

// Evaluate evaluates a SQLNotExpr into a SQLValue.
func (not *SQLNotExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	operand, err := not.SQLExpr.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if hasNullValue(operand) {
		return NewSQLNull(ctx.valueKind(), not.EvalType()), nil
	}

	if !Bool(operand) {
		return NewSQLBool(ctx.valueKind(), true), nil
	}

	return NewSQLBool(ctx.valueKind(), false), nil
}

// Normalize will attempt to change SQLNotExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (not *SQLNotExpr) Normalize(ctx *EvalCtx) Node {
	if operand, ok := not.SQLExpr.(SQLValue); ok {
		if hasNullValue(operand) {
			return NewSQLNull(ctx.valueKind(), not.EvalType())
		}

		if Bool(operand) {
			return NewSQLBool(ctx.valueKind(), false)
		} else if IsFalsy(operand) {
			return NewSQLBool(ctx.valueKind(), true)
		}
	}

	return not
}

func (not *SQLNotExpr) String() string {
	return fmt.Sprintf("not %v", not.SQLExpr)
}

// ToAggregationLanguage translates SQLNotExpr into something that can
// be used in an aggregation pipeline. If SQLNotExpr cannot be translated,
// it will return nil and error.
func (not *SQLNotExpr) ToAggregationLanguage(t *PushDownTranslator) (interface{}, error) {
	op, err := t.ToAggregationLanguage(not.SQLExpr)
	if err != nil {
		return nil, err
	}

	return wrapInNullCheckedCond(nil, bson.M{mgoOperatorNot: op}, op), nil

}

// ToMatchLanguage translates SQLNotExpr into something that can
// be used in an match expression. If SQLNotExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLNotExpr.
func (not *SQLNotExpr) ToMatchLanguage(t *PushDownTranslator) (bson.M, SQLExpr) {
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
func (nse *SQLNullSafeEqualsExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	leftVal, err := nse.left.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	rightVal, err := nse.right.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	if leftVal.IsNull() {
		if rightVal.IsNull() {
			return NewSQLBool(ctx.valueKind(), true), nil
		}
		return NewSQLBool(ctx.valueKind(), false), nil
	}

	if rightVal.IsNull() {
		if leftVal.IsNull() {
			return NewSQLBool(ctx.valueKind(), true), nil
		}
		return NewSQLBool(ctx.valueKind(), false), nil
	}

	c, err := CompareTo(leftVal, rightVal, ctx.Collation)
	if err == nil {
		return NewSQLBool(ctx.valueKind(), c == 0), nil
	}

	return NewSQLBool(ctx.valueKind(), false), err
}

// Normalize will attempt to change SQLNullSafeEqualsExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (nse *SQLNullSafeEqualsExpr) Normalize(ctx *EvalCtx) Node {
	leftVal, leftIsVal := nse.left.(SQLValue)
	rightVal, rightIsVal := nse.right.(SQLValue)

	if !leftIsVal || !rightIsVal {
		return nse
	}

	if leftVal.IsNull() {
		if rightVal.IsNull() {
			return NewSQLBool(ctx.valueKind(), true)
		}
		return NewSQLBool(ctx.valueKind(), false)
	}

	if rightVal.IsNull() {
		if leftVal.IsNull() {
			return NewSQLBool(ctx.valueKind(), true)
		}
		return NewSQLBool(ctx.valueKind(), false)
	}

	return nse
}

func (nse *SQLNullSafeEqualsExpr) String() string {
	return fmt.Sprintf("%v <=> %v", nse.left, nse.right)
}

// ToAggregationLanguage translates SQLNullSafeEqualsExpr into something that can
// be used in an aggregation pipeline. If SQLNullSafeEqualsExpr cannot be translated,
// it will return nil and error.
func (nse *SQLNullSafeEqualsExpr) ToAggregationLanguage(t *PushDownTranslator) (interface{},
	error) {
	left, err := t.ToAggregationLanguage(nse.left)
	if err != nil {
		return nil, err
	}

	right, err := t.ToAggregationLanguage(nse.right)
	if err != nil {
		return nil, err
	}

	return bson.M{mgoOperatorEq: []interface{}{
		bson.M{mgoOperatorIfNull: []interface{}{left, nil}},
		bson.M{mgoOperatorIfNull: []interface{}{right, nil}},
	}}, nil
}

// EvalType returns the EvalType associated with SQLNullSafeEqualsExpr.
func (*SQLNullSafeEqualsExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLOrExpr evaluates to true if any of its children evaluate to true.
type SQLOrExpr sqlBinaryNode

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
func (or *SQLOrExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	left, err := or.left.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	right, err := or.right.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if Bool(left) || Bool(right) {
		return NewSQLBool(ctx.valueKind(), true), nil
	}

	if hasNullValue(left, right) {
		return NewSQLNull(ctx.valueKind(), or.EvalType()), nil
	}

	return NewSQLBool(ctx.valueKind(), false), nil
}

// Normalize will attempt to change SQLOrExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (or *SQLOrExpr) Normalize(ctx *EvalCtx) Node {
	left, leftOk := or.left.(SQLValue)

	if leftOk && Bool(left) {
		return NewSQLBool(ctx.valueKind(), true)
	} else if leftOk && IsFalsy(left) {
		if or.right.EvalType() == EvalBoolean {
			return or.right
		}
		return NewSQLConvertExpr(or.right, EvalBoolean)
	}

	right, rightOk := or.right.(SQLValue)
	if rightOk && Bool(right) {
		return NewSQLBool(ctx.valueKind(), true)
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

// ToAggregationLanguage translates SQLOrExpr into something that can
// be used in an aggregation pipeline. If SQLOrExpr cannot be translated,
// it will return nil and error.
func (or *SQLOrExpr) ToAggregationLanguage(t *PushDownTranslator) (interface{}, error) {
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

	leftIsFalse := bson.M{mgoOperatorOr: []interface{}{
		bson.M{mgoOperatorEq: []interface{}{"$$left", false}},
		bson.M{mgoOperatorEq: []interface{}{"$$left", 0}},
	}}

	leftIsTrue := bson.M{mgoOperatorOr: []interface{}{
		bson.M{mgoOperatorNeq: []interface{}{"$$left", false}},
		bson.M{mgoOperatorNeq: []interface{}{"$$left", 0}},
	}}

	rightIsFalse := bson.M{mgoOperatorOr: []interface{}{
		bson.M{mgoOperatorEq: []interface{}{"$$right", false}},
		bson.M{mgoOperatorEq: []interface{}{"$$right", 0}},
	}}

	rightIsTrue := bson.M{mgoOperatorOr: []interface{}{
		bson.M{mgoOperatorNeq: []interface{}{"$$right", false}},
		bson.M{mgoOperatorNeq: []interface{}{"$$right", 0}},
	}}

	leftIsNull := bson.M{mgoOperatorEq: []interface{}{
		bson.M{
			mgoOperatorIfNull: []interface{}{"$$left", nil}},
		nil,
	}}

	rightIsNull := bson.M{mgoOperatorEq: []interface{}{
		bson.M{
			mgoOperatorIfNull: []interface{}{"$$right", nil}},
		nil,
	}}

	nullOrFalse := bson.M{mgoOperatorOr: []interface{}{
		bson.M{mgoOperatorAnd: []interface{}{
			rightIsNull, leftIsFalse,
		}},
		bson.M{mgoOperatorAnd: []interface{}{
			leftIsNull, rightIsFalse,
		}},
	}}

	nullOrTrue := bson.M{mgoOperatorOr: []interface{}{
		bson.M{mgoOperatorAnd: []interface{}{
			rightIsNull, leftIsTrue,
		}},
		bson.M{mgoOperatorAnd: []interface{}{
			leftIsNull, rightIsTrue,
		}},
	}}

	nullOrNull := bson.M{mgoOperatorAnd: []interface{}{
		leftIsNull, rightIsNull,
	}}

	letEvaluation := bson.M{
		mgoOperatorCond: []interface{}{
			nullOrNull,
			mgoNullLiteral,
			wrapInCond(
				mgoNullLiteral,
				wrapInCond(
					true,
					wrapInNullCheckedCond(
						mgoNullLiteral,
						bson.M{
							mgoOperatorOr: []interface{}{"$$left", "$$right"},
						},
						"$$left", "$$right",
					),
					nullOrTrue,
				),
				nullOrFalse,
			),
		},
	}

	return wrapInLet(letAssignment, letEvaluation), nil

}

// ToMatchLanguage translates SQLOrExpr into something that can
// be used in an match expression. If SQLOrExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLOrExpr.
func (or *SQLOrExpr) ToMatchLanguage(t *PushDownTranslator) (bson.M, SQLExpr) {
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

	if v, ok := left[mgoOperatorOr]; ok {
		array := v.([]interface{})
		cond = append(cond, array...)
	} else {
		cond = append(cond, left)
	}

	if v, ok := right[mgoOperatorOr]; ok {
		array := v.([]interface{})
		cond = append(cond, array...)
	} else {
		cond = append(cond, right)
	}

	return bson.M{mgoOperatorOr: cond}, nil
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
func (reg *SQLRegexExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	operandVal, err := reg.operand.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	patternVal, err := reg.pattern.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	if hasNullValue(operandVal, patternVal) {
		return NewSQLNull(ctx.valueKind(), reg.EvalType()), nil
	}

	pattern, patternOK := patternVal.(SQLVarchar)
	if patternOK {
		matcher, err := regexp.CompilePOSIX(pattern.String())
		if err != nil {
			return NewSQLBool(ctx.valueKind(), false), err
		}
		match := matcher.Find([]byte(operandVal.String()))
		if match != nil {
			return NewSQLBool(ctx.valueKind(), true), nil
		}
	}
	return NewSQLBool(ctx.valueKind(), false), nil
}

func (reg *SQLRegexExpr) String() string {
	return fmt.Sprintf("%s matches %s", reg.operand.String(), reg.pattern.String())
}

// ToMatchLanguage translates SQLRegexExpr into something that can
// be used in an match expression. If SQLRegexExpr can be fully translated,
// it will return the translation and nil, otherwise it will return
// a partial translation and the original SQLRegexExpr.
func (reg *SQLRegexExpr) ToMatchLanguage(t *PushDownTranslator) (bson.M, SQLExpr) {
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
			mgoOperatorRegex: bson.RegEx{
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

// SQLSubqueryCmpExpr evaluates to true if left is in any of the
// rows returned by the SQLSubqueryExpr expression results.
type SQLSubqueryCmpExpr struct {
	subqueryOp   subqueryOp
	left         SQLExpr
	subqueryExpr *SQLSubqueryExpr
	operator     string
}

// NewSQLSubqueryCmpExpr is a constructor for SQLSubqueryCmpExpr.
func NewSQLSubqueryCmpExpr(
	subqueryOp subqueryOp,
	left SQLExpr,
	subqueryExpr *SQLSubqueryExpr,
	operator string) *SQLSubqueryCmpExpr {
	return &SQLSubqueryCmpExpr{
		subqueryOp:   subqueryOp,
		left:         left,
		subqueryExpr: subqueryExpr,
		operator:     operator,
	}
}

// Evaluate evaluates a SQLSubqueryCmpExpr into a SQLValue.
func (sc *SQLSubqueryCmpExpr) Evaluate(ctx *EvalCtx) (value SQLValue, err error) {

	left, err := sc.left.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

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

	execCtx := ctx.ExecutionCtx

	if sc.subqueryExpr.correlated {
		execCtx = ctx.CreateChildExecutionCtx()
	}

	if iter, err = sc.subqueryExpr.plan.Open(execCtx); err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	row := &Row{}

	mismatch, allMatch := true, true

	switch sc.subqueryOp {
	case subqueryAll, subqueryNotIn:
		mismatch = false
	}

	var leftLen int
	leftValues, lvsOk := left.(*SQLValues)
	if lvsOk {
		leftLen = len(leftValues.Values)
	} else {
		leftLen = 1
	}

	right := &SQLValues{}

	for iter.Next(row) {

		if err = execCtx.MemoryMonitor().Release(row.Data.Size()); err != nil {
			return nil, err
		}

		for _, value := range row.Data {
			right.Values = append(right.Values, value.Data)
		}

		// Make sure the subquery returns the same number of columns as what
		// it's being compared to.
		if leftLen != len(right.Values) {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, leftLen)
		}

		var expr SQLExpr
		var matches bool
		var result SQLValue
		switch sc.subqueryOp {
		case subqueryAll:
			expr, err = comparisonExpr(left, right, sc.operator)
			if err != nil {
				return NewSQLBool(ctx.valueKind(), false), err
			}
			result, err = expr.Evaluate(ctx)
			if err != nil {
				return NewSQLBool(ctx.valueKind(), false), err
			}
			matches = Bool(result)
			if !matches {
				allMatch = false
			}
		case subqueryAny, subquerySome:
			expr, err = comparisonExpr(left, right, sc.operator)
			if err != nil {
				return NewSQLBool(ctx.valueKind(), false), err
			}
			result, err = expr.Evaluate(ctx)
			if err != nil {
				return NewSQLBool(ctx.valueKind(), false), err
			}
			matches = Bool(result)
			if matches {
				return NewSQLBool(ctx.valueKind(), true), nil
			}
		case subqueryIn:
			eq := &SQLEqualsExpr{left, right}
			result, err = eq.Evaluate(ctx)
			if err != nil {
				return NewSQLBool(ctx.valueKind(), false), err
			}
			matches = Bool(result)
			if matches {
				return NewSQLBool(ctx.valueKind(), true), nil
			}
		case subqueryNotIn:
			neq := &SQLNotEqualsExpr{left, right}
			result, err = neq.Evaluate(ctx)
			if err != nil {
				return NewSQLBool(ctx.valueKind(), false), err
			}
			matches = Bool(result)
			if !matches {
				mismatch = true
			}
		}
		row, right = &Row{}, &SQLValues{}
	}

	return NewSQLBool(ctx.valueKind(), !mismatch && allMatch), err
}

func (sc *SQLSubqueryCmpExpr) String() string {
	switch sc.subqueryOp {
	case subqueryAll:
		return fmt.Sprintf("%v %v all %v", sc.left, sc.operator, sc.subqueryExpr)
	case subqueryAny:
		return fmt.Sprintf("%v %v any %v", sc.left, sc.operator, sc.subqueryExpr)
	case subqueryIn:
		return fmt.Sprintf("%v in %v", sc.left, sc.subqueryExpr)
	case subqueryNotIn:
		return fmt.Sprintf("%v not in %v", sc.left, sc.subqueryExpr)
	case subquerySome:
		return fmt.Sprintf("%v %v some %v", sc.left, sc.operator, sc.subqueryExpr)
	}
	return ""
}

// EvalType returns the EvalType associated with SQLSubqueryCmpExpr.
func (*SQLSubqueryCmpExpr) EvalType() EvalType {
	return EvalBoolean
}

// SQLSubqueryExpr is a wrapper around a parser.SelectStatement representing
// a subquery.
type SQLSubqueryExpr struct {
	correlated bool
	allowRows  bool
	plan       PlanStage
}

// NewSQLSubqueryExpr is a constructor for SQLSubqueryExpr.
func NewSQLSubqueryExpr(correlated, allowRows bool, plan PlanStage) *SQLSubqueryExpr {
	return &SQLSubqueryExpr{
		correlated: correlated,
		allowRows:  allowRows,
		plan:       plan,
	}
}

// Evaluate evaluates a SQLSubqueryExpr into a SQLValue.
func (se *SQLSubqueryExpr) Evaluate(evalCtx *EvalCtx) (value SQLValue, err error) {

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

	execCtx := evalCtx.ExecutionCtx
	plan := se.plan

	if se.correlated {
		execCtx = evalCtx.CreateChildExecutionCtx()
		var newPlan Node
		newPlan, err = replaceColumnWithConstant(plan, execCtx)
		if err != nil {
			return nil, err
		}
		var ok bool
		plan, ok = newPlan.(PlanStage)
		if !ok {
			return nil, fmt.Errorf("replaceColumnWithConstant returned "+
				" something that is not a PlanStage: %T", newPlan)
		}
		plan = OptimizePlan(execCtx, plan)
	}

	iter, err = plan.Open(execCtx)
	if err != nil {
		return nil, err
	}

	row := &Row{}

	hasNext := iter.Next(row)
	if hasNext {

		// release this memory here... it will be re-allocated by a consuming
		// stage
		if err = execCtx.MemoryMonitor().Release(row.Data.Size()); err != nil {
			return nil, err
		}

		// Filter has to check the entire source to return an accurate 'hasNext'
		if iter.Next(&Row{}) {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErSubqueryNoOneRow)
		}
	}

	switch len(row.Data) {
	case 0:
		return NewSQLNull(evalCtx.valueKind(), se.EvalType()), nil
	case 1:
		return row.Data[0].Data, nil
	default:
		eval := &SQLValues{}
		for _, value := range row.Data {
			eval.Values = append(eval.Values, value.Data)
		}
		return eval, nil
	}
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
func (sub *SQLSubtractExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftVal, err := sub.left.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	rightVal, err := sub.right.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	if hasNullValue(leftVal, rightVal) {
		return NewSQLNull(ctx.valueKind(), sub.EvalType()), nil
	}

	return doArithmetic(leftVal, rightVal, SUB)
}

func (sub *SQLSubtractExpr) String() string {
	return fmt.Sprintf("%v-%v", sub.left, sub.right)
}

// ToAggregationLanguage translates SQLSubtractExpr into something that can
// be used in an aggregation pipeline. If SQLSubtractExpr cannot be translated,
// it will return nil and error.
func (sub *SQLSubtractExpr) ToAggregationLanguage(t *PushDownTranslator) (interface{}, error) {
	left, err := t.ToAggregationLanguage(sub.left)
	if err != nil {
		return nil, err
	}

	right, err := t.ToAggregationLanguage(sub.right)
	if err != nil {
		return nil, err
	}

	return bson.M{mgoOperatorSubtract: []interface{}{left, right}}, nil
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
func (te SQLTupleExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	var values []SQLValue

	for _, v := range te.Exprs {
		value, err := v.Evaluate(ctx)
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
func (te *SQLTupleExpr) Normalize(ctx *EvalCtx) Node {
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
func (te *SQLTupleExpr) ToAggregationLanguage(t *PushDownTranslator) (interface{}, error) {
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
func (um *SQLUnaryMinusExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	if val, err := um.SQLExpr.Evaluate(ctx); err == nil {
		if val.IsNull() {
			return NewSQLNull(ctx.valueKind(), um.EvalType()), nil
		}
		difference := NewSQLFloat(ctx.valueKind(), -Float64(val))
		converted := ConvertTo(difference, um.EvalType())
		return converted, nil
	}
	return nil, fmt.Errorf("UnaryMinus expression does not apply to a %T", um.SQLExpr)
}

// Normalize will attempt to change SQLUnaryMinusExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (um *SQLUnaryMinusExpr) Normalize(ctx *EvalCtx) Node {
	sqlVal, ok := um.SQLExpr.(SQLValue)
	if !ok {
		return um
	}

	if sqlVal.IsNull() {
		return NewSQLNull(ctx.valueKind(), um.EvalType())
	}

	if sqlVal.EvalType() == EvalBoolean {
		if sqlVal.Value().(bool) {
			return NewSQLInt64(ctx.valueKind(), -1)
		}
		return NewSQLInt64(ctx.valueKind(), 0)
	}

	return um
}

func (um *SQLUnaryMinusExpr) String() string {
	return fmt.Sprintf("-%v", um.SQLExpr)
}

// ToAggregationLanguage translates SQLUnaryMinusExpr into something that can
// be used in an aggregation pipeline. If SQLUnaryMinusExpr cannot be translated,
// it will return nil and error.
func (um *SQLUnaryMinusExpr) ToAggregationLanguage(t *PushDownTranslator) (interface{}, error) {
	operand, err := t.ToAggregationLanguage(um.SQLExpr)
	if err != nil {
		return nil, err
	}

	letAssignment := bson.M{
		"operand": operand,
	}

	letEvaluation := wrapInNullCheckedCond(
		nil,
		bson.M{mgoOperatorMultiply: []interface{}{-1, "$$operand"}},
		"$$operand",
	)

	return wrapInLet(letAssignment, letEvaluation), nil

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
func (td *SQLUnaryTildeExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	expr, err := td.SQLExpr.Evaluate(ctx)
	if err != nil {
		return NewSQLBool(ctx.valueKind(), false), err
	}

	if v, ok := expr.(SQLValue); ok {
		return NewSQLUint64(ctx.valueKind(), ^uint64(Int64(v))), nil
	}

	return NewSQLUint64(ctx.valueKind(), ^uint64(0)), nil
}

// Normalize will attempt to change SQLUnaryTildeExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (td *SQLUnaryTildeExpr) Normalize(ctx *EvalCtx) Node {
	if v, ok := td.SQLExpr.(SQLValue); ok {
		return NewSQLUint64(ctx.valueKind(), ^uint64(Int64(v)))
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
	Name     string
	Kind     variable.Kind
	Scope    variable.Scope
	evalType EvalType
}

// Evaluate evaluates a SQLVariableExpr into a SQLValue.
func (v *SQLVariableExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	value, err := ctx.Variables().Get(variable.Name(v.Name), v.Scope, v.Kind)
	if err != nil {
		return nil, err
	}

	val := GoValueToSQLValue(ctx.valueKind(), value.Value)
	converted := ConvertTo(val, SQLTypeToEvalType(value.SQLType))
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
	return v.evalType
}

// SQLXorExpr evaluates to true if and only if one of its children evaluates to true.
type SQLXorExpr sqlBinaryNode

var _ translatableToAggregation = (*SQLXorExpr)(nil)

// Evaluate evaluates a SQLXorExpr into a SQLValue.
func (xor *SQLXorExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	left, err := xor.left.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	right, err := xor.right.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if hasNullValue(left, right) {
		return NewSQLNull(ctx.valueKind(), xor.EvalType()), nil
	}

	if (IsFalsy(left) && Bool(right)) || (Bool(left) && IsFalsy(right)) {
		return NewSQLBool(ctx.valueKind(), true), nil
	}

	return NewSQLBool(ctx.valueKind(), false), nil
}

// Normalize will attempt to change SQLXorExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (xor *SQLXorExpr) Normalize(ctx *EvalCtx) Node {
	left, leftOk := xor.left.(SQLValue)
	if leftOk {
		if Bool(left) {
			return &SQLNotExpr{xor.right}
		} else if IsFalsy(left) {
			return &SQLOrExpr{NewSQLBool(ctx.valueKind(), false), xor.right}
		}
	}

	right, rightOk := xor.right.(SQLValue)
	if rightOk {
		if Bool(right) {
			return &SQLNotExpr{xor.left}
		} else if IsFalsy(right) {
			return &SQLOrExpr{NewSQLBool(ctx.valueKind(), false), xor.left}
		}
	}

	return xor
}

func (xor *SQLXorExpr) String() string {
	return fmt.Sprintf("%v xor %v", xor.left, xor.right)
}

// ToAggregationLanguage translates SQLXorExpr into something that can
// be used in an aggregation pipeline. If SQLXorExpr cannot be translated,
// it will return nil and error.
func (xor *SQLXorExpr) ToAggregationLanguage(t *PushDownTranslator) (interface{}, error) {
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
		mgoOperatorCond: []interface{}{
			bson.M{
				mgoOperatorOr: []interface{}{
					bson.M{
						mgoOperatorEq: []interface{}{
							bson.M{
								mgoOperatorIfNull: []interface{}{"$$left", nil}},
							nil,
						},
					},
					bson.M{
						mgoOperatorEq: []interface{}{
							bson.M{
								mgoOperatorIfNull: []interface{}{"$$right", nil}},
							nil,
						},
					},
				},
			},
			mgoNullLiteral,
			bson.M{
				mgoOperatorAnd: []interface{}{
					bson.M{
						mgoOperatorOr: []interface{}{"$$left", "$$right"}},
					bson.M{
						mgoOperatorNot: bson.M{
							mgoOperatorAnd: []interface{}{"$$left", "$$right"},
						},
					},
				},
			},
		},
	}

	return wrapInLet(letAssignment, letEvaluation), nil

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

type subqueryOp int

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
		)}
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
	scope variable.Scope, evalType EvalType) *SQLVariableExpr {
	return &SQLVariableExpr{
		Name:     name,
		Kind:     kind,
		Scope:    scope,
		evalType: evalType,
	}
}
