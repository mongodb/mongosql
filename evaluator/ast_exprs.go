package evaluator

import (
	"fmt"

	"github.com/10gen/sqlproxy/schema"
	"github.com/deafgoat/mixer/sqlparser"
)

//
// SQLCaseExpr holds a number of cases to evaluate as well as the value
// to return if any of the cases is matched. If none is matched,
// 'elseValue' is evaluated and returned.
//
type SQLCaseExpr struct {
	elseValue      SQLExpr
	caseConditions []caseCondition
}

// caseCondition holds a matcher used in evaluating case expressions and
// a value to return if a particular case is matched. If a case is matched,
// the corresponding 'then' value is evaluated and returned ('then'
// corresponds to the 'then' clause in a case expression).
type caseCondition struct {
	matcher SQLExpr
	then    SQLExpr
}

func (s SQLCaseExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	for _, condition := range s.caseConditions {

		m, err := Matches(condition.matcher, ctx)
		if err != nil {
			return nil, err
		}

		if m {
			return condition.then.Evaluate(ctx)
		}
	}

	return s.elseValue.Evaluate(ctx)
}

func (c *caseCondition) String() string {
	return fmt.Sprintf("when (%v) then %v", c.matcher, c.then)
}

func (s SQLCaseExpr) String() string {
	str := fmt.Sprintf("case ")
	for _, cond := range s.caseConditions {
		str += fmt.Sprintf("%v ", cond.String())
	}
	if s.elseValue != nil {
		str += fmt.Sprintf("else %v ", s.elseValue.String())
	}
	str += fmt.Sprintf("end")
	return str
}

func (s SQLCaseExpr) Type() schema.SQLType {
	conds := []SQLExpr{s.elseValue}
	for _, cond := range s.caseConditions {
		conds = append(conds, cond.then)
	}
	return preferentialType(conds...)
}

//
// SQLCtorExpr is a representation of a sqlparser.CtorExpr.
//
type SQLCtorExpr struct {
	Name string
	Args sqlparser.ValExprs
}

func (s SQLCtorExpr) Evaluate(_ *EvalCtx) (SQLValue, error) {

	if len(s.Args) == 0 {
		return nil, fmt.Errorf("no arguments supplied to SQLCtorExpr")
	}

	// TODO: currently only supports single argument constructors
	strVal, ok := s.Args[0].(sqlparser.StrVal)
	if !ok {
		return nil, fmt.Errorf("%v constructor requires string argument: got %T", string(s.Name), s.Args[0])
	}

	arg := string(strVal)

	switch s.Name {
	case sqlparser.AST_DATE:
		return NewSQLValue(arg, schema.SQLDate, schema.MongoNone)
	case sqlparser.AST_DATETIME:
		return NewSQLValue(arg, schema.SQLTimestamp, schema.MongoNone)
	case sqlparser.AST_TIMESTAMP:
		return NewSQLValue(arg, schema.SQLTimestamp, schema.MongoNone)
	default:
		return nil, fmt.Errorf("%v constructor is not supported", string(s.Name))
	}
}

//
// SQLColumnExpr represents a column reference.
//
type SQLColumnExpr struct {
	tableName  string
	columnName string
	columnType schema.ColumnType
}

func (c SQLColumnExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	// TODO how do we report field not existing? do we just treat is a NULL, or something else?
	for _, row := range ctx.Rows {
		for _, data := range row.Data {
			if data.Table == c.tableName {
				value, hasValue := row.GetField(c.tableName, c.columnName)
				if !hasValue || value == nil {
					return SQLNull, nil
				}
				return NewSQLValue(value, c.columnType.SQLType, c.columnType.MongoType)
			}
		}
	}
	return SQLNull, nil
}

func (c SQLColumnExpr) String() string {
	var str string
	if c.tableName != "" {
		str += c.tableName + "."
	}
	str += c.columnName
	return str
}

func (c SQLColumnExpr) Type() schema.SQLType {
	return c.columnType.SQLType
}

//
// SQLConvertExpr wraps a SQLExpr that can be
// converted to another SQLType.
//
type SQLConvertExpr struct {
	expr     SQLExpr
	convType schema.SQLType
}

func (ce SQLConvertExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	v, err := ce.expr.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	return NewSQLValue(v.Value(), ce.convType, schema.MongoNone)
}

func (ce SQLConvertExpr) String() string {
	return ce.expr.String()
}

func (ce SQLConvertExpr) Type() schema.SQLType {
	return ce.convType
}

// SQLSubqueryExpr is a wrapper around a sqlparser.SelectStatement representing
// a subquery.
type SQLSubqueryExpr struct {
	stmt  sqlparser.SelectStatement
	exprs []SQLExpr
}

func (se *SQLSubqueryExpr) Evaluate(ctx *EvalCtx) (value SQLValue, err error) {

	ctx.ExecCtx.Depth += 1

	var operator Operator

	operator, err = PlanQuery(ctx.ExecCtx, se.stmt)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err == nil {
			err = operator.Close()
		} else {
			operator.Close()
		}

		if err == nil {
			err = operator.Err()
		}

		// add context to error
		if err != nil {
			err = fmt.Errorf("SubqueryValue (%v): %v", ctx.ExecCtx.Depth, err)
		}

		ctx.ExecCtx.Depth -= 1

	}()

	err = operator.Open(ctx.ExecCtx)
	if err != nil {
		return nil, err
	}

	row := &Row{}

	hasNext := operator.Next(row)

	// Filter has to check the entire source to return an accurate 'hasNext'
	if hasNext && operator.Next(&Row{}) {
		return nil, fmt.Errorf("Subquery returns more than 1 row")
	}

	values := row.GetValues(operator.OpFields())
	if len(values) == 0 {
		return SQLNone, nil
	}

	eval := &SQLValues{}
	for _, value := range values {
		field, err := NewSQLValue(value, schema.SQLNone, schema.MongoNone)
		if err != nil {
			return nil, err
		}
		eval.Values = append(eval.Values, field)
	}
	return eval, nil
}

func (se *SQLSubqueryExpr) String() string {
	buf := sqlparser.NewTrackedBuffer(nil)
	se.stmt.Format(buf)
	return fmt.Sprintf("(%v)", buf.String())
}

func (se *SQLSubqueryExpr) Type() schema.SQLType {
	return preferentialType(se.exprs...)
}

//
// SQLTupleExpr represents a tuple.
//
type SQLTupleExpr struct {
	Exprs []SQLExpr
}

func (te SQLTupleExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	var values []SQLValue

	for _, v := range te.Exprs {
		value, err := v.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}

	return &SQLValues{values}, nil
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

func (te SQLTupleExpr) Type() schema.SQLType {
	return preferentialType(te.Exprs...)
}
