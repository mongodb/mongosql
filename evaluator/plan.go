package evaluator

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/schema"
	"github.com/mongodb/mongo-tools/common/bsonutil"
	"github.com/mongodb/mongo-tools/common/json"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Column contains information used to select data
// from a query plan. 'Table' and 'Column' define the
// source of the data while 'View' holds the display
// header representation of the data.
type Column struct {
	Table     string
	Name      string
	View      string
	SQLType   schema.SQLType
	MongoType schema.MongoType
}

type ConnectionCtx interface {
	LastInsertId() int64
	RowCount() int64
	ConnectionId() uint32
	DB() string
	Session() *mgo.Session
}

// ExecutionCtx holds exeuction context information
// used by each Iterator implemenation.
type ExecutionCtx struct {
	Depth int

	// GroupRows holds a set of rows used by each GROUP BY combination
	GroupRows []Row

	// SrcRows caches the data gotten from a table scan or join node
	SrcRows []*Row

	ConnectionCtx

	AuthProvider AuthProvider
}

func NewExecutionCtx(connCtx ConnectionCtx) *ExecutionCtx {
	return &ExecutionCtx{
		ConnectionCtx: connCtx,
		AuthProvider:  NewAuthProvider(connCtx.Session()),
	}
}

// PlanStage represents a single a node in the Plan tree.
type PlanStage interface {
	// Open returns an iterator that returns results from executing this plan stage with the given
	// ExecutionContext.
	Open(*ExecutionCtx) (Iter, error)

	// OpFields returns the set of columns that are contained in results from this plan.
	OpFields() []*Column
}

// Iter represents an object that can iterate through a set of rows.
type Iter interface {
	// Next retrieves the next row from this iterator. It returns true if it has
	// additional data and false if there is no more data or if an error occurred
	// during processing.
	//
	// When Next returns false, the Err method should be called to verify if
	// there was an error during processing.
	//
	// For example:
	//    iter, err := plan.Open(ctx);
	//
	//    if err != nil {
	//        return err
	//    }
	//
	//    for iter.Next(&row) {
	//        fmt.Printf("Row: %v\n", row)
	//    }
	//
	//    if err := iter.Close(); err != nil {
	//        return err
	//    }
	//
	//    if err := iter.Err(); err != nil {
	//        return err
	//    }
	//
	Next(*Row) bool

	// Close frees up any resources in use by this iterator. Callers should always
	// call the Close method once they are finished with an iterator.
	Close() error

	// Err returns nil if no errors happened during processing, or the actual
	// error otherwise. Callers should always call the Err method to check whether
	// any error was encountered during processing they are finished with an iterator.
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

func (se *SelectExpression) clone() *SelectExpression {
	return &SelectExpression{
		Column:     se.Column,
		Expr:       se.Expr,
		Referenced: se.Referenced,
		RefColumns: se.RefColumns,
	}
}

// AggRowCtx holds evaluated data as well as the relevant context used to evaluate the data
// used for passing data - used to process aggregation functions - between iterators.
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

func (se SelectExpressions) Unique() SelectExpressions {
	var results SelectExpressions
	for _, e := range se {
		if !results.Contains(*e.Column) {
			results = append(results, e)
		}
	}

	return results
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
		value, err := NewSQLValue(v.Value, schema.SQLNone, schema.MongoNone)
		if err != nil {
			return nil, err
		}
		values = append(values, Value{v.Name, v.Name, value})
	}
	return values, nil
}

// PlanStageVisitor is an implementation of the visitor pattern.
type PlanStageVisitor interface {
	// Visit is called with a plan stage. It returns:
	// - PlanStage, the plan used to replace the argument.
	// - error
	Visit(p PlanStage) (PlanStage, error)
}

func prettyPrintPlan(b *bytes.Buffer, p PlanStage, d int) {

	printTabs(b, d)

	switch typedE := p.(type) {

	case *DualStage:

		b.WriteString("↳ Dual(")
		for i, c := range typedE.sExprs {
			if i != 0 {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("%v as %v", c.Expr.String(), c.Name))
		}

		b.WriteString(")")

	case *EmptyStage:

		b.WriteString("↳ Empty:")

	case *FilterStage:

		b.WriteString(fmt.Sprintf("↳ Filter (%v):\n", typedE.matcher))
		prettyPrintPlan(b, typedE.source, d+1)

	case *GroupByStage:

		b.WriteString("↳ GroupBy(")

		for i, c := range typedE.keyExprs {
			if i != 0 {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("%v as %v", c.Expr.String(), c.Name))
		}

		b.WriteString("):\n")

		prettyPrintPlan(b, typedE.source, d+1)

	case *JoinStage:

		b.WriteString("↳ Join:\n")

		prettyPrintPlan(b, typedE.left, d+1)

		printTabs(b, d+1)

		b.WriteString(fmt.Sprintf("%v\n", typedE.kind))

		prettyPrintPlan(b, typedE.right, d+1)

		if typedE.matcher != nil {
			printTabs(b, d+1)

			b.WriteString(fmt.Sprintf("on %v\n", typedE.matcher.String()))
		}

	case *LimitStage:

		b.WriteString(fmt.Sprintf("↳ Limit(offset: %v, limit: %v):\n", typedE.offset, typedE.limit))
		prettyPrintPlan(b, typedE.source, d+1)

	case *OrderByStage:

		b.WriteString("↳ OrderBy(")

		for i, c := range typedE.keys {
			if i != 0 {
				b.WriteString(", ")
			}

			dir := "ASC"
			if !c.ascending {
				dir = "DESC"
			}

			b.WriteString(fmt.Sprintf("%v %v", c.expr.Expr.String(), dir))
		}

		b.WriteString("):\n")

		prettyPrintPlan(b, typedE.source, d+1)

	case *ProjectStage:

		b.WriteString("↳ Project(")

		for i, c := range typedE.sExprs {
			if i != 0 {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("%v as %v", c.Name, c.View))
		}

		b.WriteString("):\n")
		prettyPrintPlan(b, typedE.source, d+1)

	case *SchemaDataSourceStage:

		b.WriteString("↳ SchemaDataSource:")

	case *SourceAppendStage:

		b.WriteString("↳ SourceAppend:\n")
		prettyPrintPlan(b, typedE.source, d+1)

	case *SourceRemoveStage:

		b.WriteString("↳ SourceRemove:\n")
		prettyPrintPlan(b, typedE.source, d+1)

	case *MongoSourceStage:

		b.WriteString(fmt.Sprintf("↳ MongoSource: '%v' (db: '%v', collection: '%v')", typedE.tableName, typedE.dbName, typedE.collectionName))

		if typedE.aliasName != "" {
			b.WriteString(fmt.Sprintf(" as '%v'", typedE.aliasName))
		}

		b.WriteString(":\n")
		prettyPipeline, err := pipelineJSON(typedE.pipeline, d+1)
		if err != nil { // marshaling as json failed, fall back to Sprintf
			prettyPipeline = pipelineString(typedE.pipeline, d+1)
		}
		b.Write(prettyPipeline)

	case *BSONSourceStage:

		b.WriteString("↳ BSONSource:\n")

	default:

		panic(fmt.Sprintf("unsupported print operator: %T", typedE))

	}
}

func pipelineJSON(stages []bson.D, depth int) ([]byte, error) {
	buf := bytes.Buffer{}

	for i, s := range stages {
		converted, err := bsonutil.GetBSONValueAsJSON(s)
		if err != nil {
			return nil, err
		}
		b, err := json.Marshal(converted)
		if err != nil {
			return nil, err
		}
		printTabs(&buf, depth)
		buf.Write(b)
		if i != len(stages)-1 {
			buf.WriteString(",\n")
		}
	}
	return buf.Bytes(), nil
}

func pipelineString(stages []bson.D, depth int) []byte {
	buf := bytes.Buffer{}
	for i, stage := range stages {
		printTabs(&buf, depth)
		buf.WriteString(fmt.Sprintf("  stage %v: '%v'\n", i+1, stage))
	}
	return buf.Bytes()
}

func printTabs(b *bytes.Buffer, d int) {
	for i := 0; i < d; i++ {
		b.WriteString("\t")
	}
}

// PrettyPrintPlan takes a plan and recursively prints its source.
func PrettyPrintPlan(p PlanStage) string {

	b := bytes.NewBufferString("")

	prettyPrintPlan(b, p, 0)

	return b.String()

}

// walkPlanTree handles walking the children of the provided plan, calling
// v.Visit on each child which is a plan. Some visitor implementations
// may ignore this method completely, but most will use it as the default
// implementation for a majority of nodes.
func walkPlanTree(v PlanStageVisitor, p PlanStage) (PlanStage, error) {

	switch typedP := p.(type) {

	case *DualStage, *EmptyStage:
		// nothing to do
	case *FilterStage:
		source, err := v.Visit(typedP.source)
		if err != nil {
			return nil, err
		}

		if typedP.source != source {
			p = &FilterStage{
				source:      source,
				matcher:     typedP.matcher,
				hasSubquery: typedP.hasSubquery,
			}
		}
	case *GroupByStage:
		source, err := v.Visit(typedP.source)
		if err != nil {
			return nil, err
		}

		if typedP.source != source {
			p = &GroupByStage{
				source:      source,
				selectExprs: typedP.selectExprs,
				keyExprs:    typedP.keyExprs,
			}
		}
	case *JoinStage:
		left, err := v.Visit(typedP.left)
		if err != nil {
			return nil, err
		}
		right, err := v.Visit(typedP.right)
		if err != nil {
			return nil, err
		}

		if typedP.left != left || typedP.right != right {
			p = &JoinStage{
				left:     left,
				right:    right,
				kind:     typedP.kind,
				strategy: typedP.strategy,
				matcher:  typedP.matcher,
			}
		}
	case *LimitStage:
		source, err := v.Visit(typedP.source)
		if err != nil {
			return nil, err
		}

		if typedP.source != source {
			p = &LimitStage{
				source: source,
				limit:  typedP.limit,
				offset: typedP.offset,
			}
		}
	case *OrderByStage:
		source, err := v.Visit(typedP.source)
		if err != nil {
			return nil, err
		}

		if typedP.source != source {
			p = &OrderByStage{
				source: source,
				keys:   typedP.keys,
			}
		}
	case *ProjectStage:
		source, err := v.Visit(typedP.source)
		if err != nil {
			return nil, err
		}

		if typedP.source != source {
			p = &ProjectStage{
				source: source,
				sExprs: typedP.sExprs,
			}
		}
	case *SchemaDataSourceStage:
		// nothing to do
	case *SourceAppendStage:
		source, err := v.Visit(typedP.source)
		if err != nil {
			return nil, err
		}

		if typedP.source != source {
			p = &SourceAppendStage{
				source: source,
			}
		}
	case *SourceRemoveStage:
		source, err := v.Visit(typedP.source)
		if err != nil {
			return nil, err
		}

		if typedP.source != source {
			p = &SourceRemoveStage{
				source: source,
			}
		}
	case *MongoSourceStage, *BSONSourceStage:
		// nothing to do
	default:
		return nil, fmt.Errorf("unsupported plan stage: %T", typedP)
	}

	return p, nil
}
