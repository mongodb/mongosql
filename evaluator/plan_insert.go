package evaluator

import (
	"context"
	"fmt"

	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// InsertCommand handles an Insert command.
type InsertCommand struct {
	dbName    string
	tableName string
	tableCols results.Columns
	// A positionMap is a mapping from a given insert column's mongo name to a position index into each row
	// of the exprListList. Any column that does not appear in the insert columns list will not appear
	// in this map unless _no_ columns are specified (in which case they will all appear with the
	// position index corresponding to the table position).
	//
	// This map is necessary because the insert columns list can be in a different order than the table
	// columns, e.g.:
	//   create table foo(x int, y int);
	//   insert into foo(y, x) values(3,4);
	positionMap  map[string]int
	exprListList [][]SQLExpr
}

// NewInsertCommand creates a new InsertCommand
func NewInsertCommand(db string, table string, tableCols results.Columns, positionMap map[string]int, exprListList [][]SQLExpr) *InsertCommand {
	return &InsertCommand{db, table, tableCols, positionMap, exprListList}
}

// Children returns a slice of all the Node children of the Node.
func (InsertCommand) Children() []Node {
	return []Node{}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (InsertCommand) ReplaceChild(i int, n Node) {
	panicWithInvalidIndex("InsertCommand", i, -1)
}

// Execute runs this command.
func (com *InsertCommand) Execute(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) error {
	docs, err := com.exprListToDocs(com.positionMap, com.exprListList)
	if err != nil {
		return err
	}
	return cfg.commandHandler.Insert(ctx, com.dbName, com.tableName, docs)
}

// exprListToDocs converts a list of []SQLExpr lists to a list of bsonutil.D. We use []interface{}
// to be consistent with what mongodb.Session.Insert expects.
func (com *InsertCommand) exprListToDocs(positionMap map[string]int, exprList [][]SQLExpr) ([]interface{}, error) {
	docs := make([]interface{}, len(exprList))
	for i := range exprList {
		var err error
		docs[i], err = com.exprListListToDoc(positionMap, exprList[i])
		if err != nil {
			return nil, err
		}
	}
	return docs, nil
}

// exprListListToDoc converts a list of []SQLExpr to a bsonutil.D with the proper keys derived from the
// tableCols.
func (com *InsertCommand) exprListListToDoc(positionMap map[string]int, exprListList []SQLExpr) (interface{}, error) {
	doc := make(bsonutil.D, len(com.tableCols))
	for i, col := range com.tableCols {
		colName := col.Name
		if idx, ok := positionMap[colName]; ok {
			// If the column name is in the positionMap, we can get the
			// associated expr from the []SQLExpr list and convert it to
			// a proper bson value.
			v, err := com.convertExpr(col, exprListList[idx])
			if err != nil {
				return nil, err
			}
			doc[i] = bsonutil.NewDocElem(colName, v)
		} else {
			// If this column name does not exist in the positionMap, it must be
			// given the default value (currently always NULL). Due to our
			// json-schema validators, we cannot simply use a missing key.
			doc[i] = bsonutil.NewDocElem(colName, nil)
		}
	}
	return doc, nil
}

// convertExpr converts a SQLExpr into a bson value.
func (com *InsertCommand) convertExpr(col *results.Column, expr SQLExpr) (interface{}, error) {
	switch typedC := expr.(type) {
	case SQLValueExpr:
		value := values.ConvertTo(typedC.Value, col.EvalType)
		// Null values must be handled specially for two reasons:
		// 1. Value() for Null Values does not do what we would prefer:
		//    it will return the 0 value for the underlying type.
		// 2. We must disallow Null Values with a useful error if the
		//    column was specified NOT NULL. The json schema validator only
		//    states that validation failed without saying *why*.
		if value.IsNull() {
			if !col.Nullable {
				return nil, fmt.Errorf("column %s cannot be NULL", col.Name)
			}
			return nil, nil
		}
		// We need to handle decimals specially because the go driver does not
		// support the decimal package that we use.
		if dec, ok := value.(values.SQLDecimal128); ok {
			v, err := primitive.ParseDecimal128(dec.String())
			if err != nil {
				return nil, err
			}
			return v, nil
		}
		return value.Value(), nil
	default:
		// This would suggest an error in both the parser and the algebrizer, which should
		// both disallow non-SQLValueExpr []SQLExpr here.
		panic(fmt.Sprintf("insert only accepts values not %T", expr))
	}
}
