package planner

import (
	"github.com/erh/mixer/sqlparser"
	"github.com/erh/mongo-sql-temp/config"
	"github.com/erh/mongo-sql-temp/translator/types"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

// Column contains information used to select data
// from an operator. 'Table' and 'Column' define the
// source of the data while 'View' holds the display
// header representation of the data.
type Column struct {
	Table string
	Name  string
	View  string
}

// ExecutionCtx holds data that is used by each operator.
type ExecutionCtx struct {
	Config *config.Config
	Db     string
	Rows   []types.Row
}

// Operator defines a set of functions that are implemented by each
// node in the query tree.
type Operator interface {
	Open(*ExecutionCtx) error
	Next(*types.Row) bool
	Close() error
	OpFields() []*Column
	Err() error
}

// SelectExpression is a panner representation of each expression in a select
// query expression. For example, in the query below, there are three select
// expressions and each will have a corresponding SelectExpression:
//
// select a, b + c, d from foo;
type SelectExpression struct {
	// Example query:
	//
	// select name, (discount * price) as discountRate from foo;
	//
	// Column holds information for this select expression - specifically,
	// the "name", "view", and "table" for the field. In the example above,
	// the Column for the first expression will hold "name", "name", and
	// "foo" respectively.
	//
	// The second expression will hold "discountRate", "discountRate", and
	// "foo" respectively. For unaliased expressions, the name and view will
	// hold a stringified version of the expression. e.g. consider the aggregate
	// function below:
	//
	// select name, sum(price) from foo;
	//
	// Column will hold "sum(price)", "sum(price)", and "".
	// Non-column name expressions always have a source table of "".
	//
	Column
	// RefColumns is a slice of every other column(s) referenced in the
	// select expression. For example, in the expression:
	//
	// select name, (discount * price) as discountRate from foo;
	//
	// The RefColumns slice for the second expression will contain a Column
	// entry for both the "discount" and the "price" fields.
	//
	RefColumns []*Column
	// Expr holds the actual expression to be evaluated during processing.
	// For column names expressions, it is nil. For example, in the expression:
	//
	// select name, (discount * price) as discountRate from foo;
	//
	// Expr will be nil for the first expression and a BinaryExpr for the second
	// expression.
	//
	Expr sqlparser.Expr
}

type SelectExpressions []SelectExpression

func (sc SelectExpressions) GetColumns() []*Column {
	columns := make([]*Column, 0)

	for _, selectExpression := range sc {
		columns = append(columns, selectExpression.RefColumns...)
	}

	return columns
}

func (sc SelectExpression) isAggFunc() bool {
	_, ok := sc.Expr.(*sqlparser.FuncExpr)
	return ok
}

func (sc SelectExpressions) AggFunctions() SelectExpressions {

	sExprs := SelectExpressions{}

	for _, selectExpression := range sc {
		if _, ok := selectExpression.Expr.(*sqlparser.FuncExpr); ok {
			sExprs = append(sExprs, selectExpression)
		}
	}

	return sExprs
}

func getKey(key string, doc bson.D) (interface{}, bool) {
	index := strings.Index(key, ".")
	if index == -1 {
		for _, entry := range doc {
			if strings.ToLower(key) == strings.ToLower(entry.Name) { // TODO optimize
				return entry.Value, true
			}
		}
		return nil, false
	}
	left := key[0:index]
	docMap := doc.Map()
	value, hasValue := docMap[left]
	if value == nil {
		return value, hasValue
	}
	subDoc, ok := docMap[left].(bson.D)
	if !ok {
		return nil, false
	}
	return getKey(key[index+1:], subDoc)
}
