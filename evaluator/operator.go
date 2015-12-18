package evaluator

import (
	"bytes"
	"fmt"
	"github.com/10gen/sqlproxy/schema"
	"github.com/deafgoat/mixer/sqlparser"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

// Column contains information used to select data
// from an operator. 'Table' and 'Column' define the
// source of the data while 'View' holds the display
// header representation of the data.
type Column struct {
	Table      string
	Name       string
	View       string
	InSubquery bool
}

type ConnectionCtx interface {
	LastInsertId() int64
	RowCount() int64
	ConnectionId() uint32
	DB() string
}

// ExecutionCtx holds exeuction context information
// used by each Operator implemenation.
type ExecutionCtx struct {
	Schema   *schema.Schema
	Session  *mgo.Session
	ParseCtx *ParseCtx
	Db       string
	// GroupRows holds a set of rows used by each GROUP BY combination
	GroupRows []Row
	// SrcRows caches the data gotten from a table scan or join node
	SrcRows []*Row
	// Depth indicates what level within a WHERE expression - containing
	// a subquery = is being processed
	Depth int
	ConnectionCtx
}

// Operator defines a set of functions that are implemented by each
// node in the query pipeline.
type Operator interface {
	//
	// Open prepares the operator for query processing. Implementations of this
	// interface should perform all tasks necessary for subsequently processing
	// row data data - including opening of source nodes, setting up state, etc.
	// If successful, the Next method can be called on the operator to retrieve
	// processed row data.
	//
	Open(*ExecutionCtx) error

	//
	// Next retrieves the next row from this operator. It returns true if it has
	// additional data and false if there is no more data or if an error occurred
	// during processing.
	//
	// When Next returns false, the Err method should be called to verify if
	// there was an error during processing.
	//
	// For example:
	//
	//    if err := node.Open(ctx); err != nil {
	//        return err
	//    }
	//
	//    for node.Next(&row) {
	//        fmt.Printf("Row: %v\n", row)
	//    }
	//
	//    if err := node.Close(); err != nil {
	//        return err
	//    }
	//
	//    if err := node.Err(); err != nil {
	//        return err
	//    }
	//
	Next(*Row) bool

	//
	// OpFields returns all the column headers that this operator includes for each
	// row returned by the Next method.
	//
	// For example, in the query below:
	//
	// select a, b from foo;
	//
	// the OpFields method will return a slice with two elements - for each of the
	// select expressions ("a" and "b") in the query.
	//
	OpFields() []*Column

	//
	// Close frees up any resources in use by this operator. Callers should always
	// call the Close method once they are finished with an operator.
	//
	Close() error

	//
	// Err returns nil if no errors happened during processing, or the actual
	// error otherwise. Callers should always call the Err method to check whether
	// any error was encountered during processing they are finished with an operator.
	//
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
	// Column will hold "sum(price)", "sum(price)", and foo.
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
	// Referenced indicates if this column is part of the select expressions
	// by way of being referenced - as opposed to be explicitly requested. e.g.
	// in the expression:
	//
	// select name, (discount * price) as discountRate from foo;
	//
	// the 'discount' and 'price' columns are referenced
	Referenced bool
}

// AggRowCtx holds evaluated data as well as the relevant context used to evaluate the data
// used for passing data - used to process aggregation functions - between operators.
type AggRowCtx struct {
	// Row contains the evaluated data for each record.
	Row Row
	// Ctx contains the rows used in evaluating any aggregation
	// function used in the GROUP BY expression.
	Ctx []Row
}

type SelectExpressions []SelectExpression

func (ses SelectExpressions) String() string {

	b := bytes.NewBufferString(fmt.Sprintf("columns: \n"))

	for _, expr := range ses {
		b.WriteString(fmt.Sprintf("- %#v\n", expr.Column))
	}

	return b.String()
}

func (se SelectExpressions) GetColumns() []*Column {
	columns := make([]*Column, 0)

	for _, selectExpression := range se {
		columns = append(columns, selectExpression.RefColumns...)
	}

	return columns
}

func (se SelectExpressions) Contains(column Column) bool {

	for _, expr := range se {
		if expr.Column.Name == column.Name &&
			expr.Column.Table == column.Table &&
			expr.Column.View == column.View {
			return true
		}
	}

	return false
}

// hasSubquery returns true if any of the select expressions contains a
// subquery expression.
func (se SelectExpressions) HasSubquery() bool {

	for _, expr := range se {
		if expr.InSubquery {
			return true
		}
	}

	return false
}

func (se SelectExpression) isAggFunc() bool {
	_, ok := se.Expr.(*sqlparser.FuncExpr)
	return ok
}

func (se SelectExpressions) AggFunctions() SelectExpressions {

	sExprs := SelectExpressions{}

	for _, sExpr := range se {
		if hasAggFunctions(sExpr.Expr) {
			sExprs = append(sExprs, sExpr)
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

// bsonDToValues takes a bson.D document and returns
// the corresponding values.
func bsonDToValues(document bson.D) ([]Value, error) {
	values := []Value{}
	for _, v := range document {
		value, err := NewSQLValue(v.Value, "")
		if err != nil {
			return nil, err
		}
		values = append(values, Value{v.Name, v.Name, value})
	}
	return values, nil
}
