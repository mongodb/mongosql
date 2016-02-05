package evaluator

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/schema"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Column contains information used to select data
// from an operator. 'Table' and 'Column' define the
// source of the data while 'View' holds the display
// header representation of the data.
type Column struct {
	Table string
	Name  string
	View  string
	Type  string
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
	*Column
	// RefColumns is a slice of every other column(s) referenced in the
	// select expression. For example, in the expression:
	//
	// select name, (discount * price) as discountRate from foo;
	//
	// The RefColumns slice for the second expression will contain a Column
	// entry for both the "discount" and the "price" fields.
	//
	RefColumns []*Column
	// Expr holds the transformed sqlparser expression (a SQLExpr) that can
	// subsequently be evaluated during processing.
	// For column names expressions, it is nil. For example, in the expression:
	//
	Expr SQLExpr
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
			expr.Column.View == column.View &&
			expr.Column.Table == column.Table {
			return true
		}
	}

	return false
}

func (se SelectExpressions) AggFunctions() (*SelectExpressions, error) {

	sExprs := SelectExpressions{}

	for _, sExpr := range se {
		aggFuncs, err := getAggFunctions(sExpr.Expr)
		if err != nil {
			return nil, err
		}

		if len(aggFuncs) != 0 {
			sExprs = append(sExprs, sExpr)
		}
	}

	return &sExprs, nil
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

// OperatorVisitor is an implementation of the visitor pattern.
type OperatorVisitor interface {
	// Visit is called with an operator. It returns:
	// - Operator is the operator used to replace the argument.
	// - error
	Visit(o Operator) (Operator, error)
}

func prettyPrintPlan(b *bytes.Buffer, o Operator, d int) {

	printTabs(b, d)

	switch typedE := o.(type) {

	case *Dual:

		b.WriteString("↳ Dual")

	case *Empty:

		b.WriteString("↳ Empty")

	case *Filter:

		b.WriteString("↳ Filter:\n")
		prettyPrintPlan(b, typedE.source, d+1)

	case *GroupBy:

		b.WriteString("↳ GroupBy:\n")
		prettyPrintPlan(b, typedE.source, d+1)

	case *Join:

		b.WriteString("↳ Join:\n")

		prettyPrintPlan(b, typedE.left, d+1)

		printTabs(b, d+1)

		b.WriteString(fmt.Sprintf("%v\n", typedE.kind))

		prettyPrintPlan(b, typedE.right, d+1)

	case *Limit:

		b.WriteString("↳ Limit:\n")
		prettyPrintPlan(b, typedE.source, d+1)

	case *OrderBy:

		b.WriteString("↳ OrderBy:\n")
		prettyPrintPlan(b, typedE.source, d+1)

	case *Project:

		b.WriteString("↳ Project:\n")
		prettyPrintPlan(b, typedE.source, d+1)

	case *SchemaDataSource:

		b.WriteString("↳ SchemaDataSource")

	case *SourceAppend:

		b.WriteString("↳ SourceAppend:\n")
		prettyPrintPlan(b, typedE.source, d+1)

	case *SourceRemove:

		b.WriteString("↳ SourceRemove:\n")
		prettyPrintPlan(b, typedE.source, d+1)

	case *Subquery:

		b.WriteString("↳ Subquery:\n")
		prettyPrintPlan(b, typedE.source, d+1)

	case *TableScan:

		b.WriteString(fmt.Sprintf("↳ TableScan '%v'", typedE.tableName))

		if typedE.aliasName != "" {
			b.WriteString(fmt.Sprintf(" as '%v'", typedE.aliasName))
		}

		b.WriteString("\n")

		for i, stage := range typedE.pipeline {
			printTabs(b, d+1)
			b.WriteString(fmt.Sprintf("  stage %v: '%v'\n", i+1, stage))
		}

	default:

		panic(fmt.Sprintf("unsupported print operator: %T", typedE))

	}

}

// PrettyPrintPlan takes an operator and recursively prints its source.
func PrettyPrintPlan(o Operator) string {

	b := bytes.NewBufferString("")

	prettyPrintPlan(b, o, 0)

	return b.String()

}

// walkOperatorTree handles walking the children of the provided operator, calling
// v.Visit on each child which is an operator. Some visitor implementations
// may ignore this method completely, but most will use it as the default
// implementation for a majority of nodes.
func walkOperatorTree(v OperatorVisitor, o Operator) (Operator, error) {

	switch typedO := o.(type) {

	case *Dual, *Empty:
		// nothing to do
	case *Filter:
		source, err := v.Visit(typedO.source)
		if err != nil {
			return nil, err
		}

		if typedO.source != source {
			o = &Filter{
				source:      source,
				matcher:     typedO.matcher,
				hasSubquery: typedO.hasSubquery,
			}
		}
	case *GroupBy:
		source, err := v.Visit(typedO.source)
		if err != nil {
			return nil, err
		}

		if typedO.source != source {
			o = &GroupBy{
				source:      source,
				selectExprs: typedO.selectExprs,
				keyExprs:    typedO.keyExprs,
			}
		}
	case *Join:
		left, err := v.Visit(typedO.left)
		if err != nil {
			return nil, err
		}
		right, err := v.Visit(typedO.right)
		if err != nil {
			return nil, err
		}

		if typedO.left != left || typedO.right != right {
			o = &Join{
				left:     left,
				right:    right,
				kind:     typedO.kind,
				strategy: typedO.strategy,
			}
		}
	case *Limit:
		source, err := v.Visit(typedO.source)
		if err != nil {
			return nil, err
		}

		if typedO.source != source {
			o = &Limit{
				source:   source,
				rowcount: typedO.rowcount,
				offset:   typedO.offset,
			}
		}
	case *OrderBy:
		source, err := v.Visit(typedO.source)
		if err != nil {
			return nil, err
		}

		if typedO.source != source {
			o = &OrderBy{
				source: source,
				keys:   typedO.keys,
			}
		}
	case *Project:
		source, err := v.Visit(typedO.source)
		if err != nil {
			return nil, err
		}

		if typedO.source != source {
			o = &Project{
				source: source,
				sExprs: typedO.sExprs,
			}
		}
	case *SchemaDataSource:
		// nothing to do
	case *SourceAppend:
		source, err := v.Visit(typedO.source)
		if err != nil {
			return nil, err
		}

		if typedO.source != source {
			o = &SourceAppend{
				source:      source,
				hasSubquery: typedO.hasSubquery,
			}
		}
	case *SourceRemove:
		source, err := v.Visit(typedO.source)
		if err != nil {
			return nil, err
		}

		if typedO.source != source {
			o = &SourceRemove{
				source:      source,
				hasSubquery: typedO.hasSubquery,
			}
		}
	case *Subquery:
		source, err := v.Visit(typedO.source)
		if err != nil {
			return nil, err
		}

		if typedO.source != source {
			o = &Subquery{
				tableName: typedO.tableName,
				source:    source,
			}
		}
	case *TableScan:
		// nothing to do
	default:
		return nil, fmt.Errorf("unsupported operator: %T", typedO)
	}

	return o, nil
}
