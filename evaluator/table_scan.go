package evaluator

import (
	"fmt"
	"github.com/10gen/sqlproxy/schema"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

// TableScan is the primary interface for SQLProxy to a MongoDB
// installation and executes simple queries against collections.
type TableScan struct {
	dbName      string
	tableName   string
	matcher     SQLExpr
	iter        FindResults
	database    *schema.Database
	tableSchema *schema.Table
	ctx         *ExecutionCtx
	pipeline    []bson.D
	err         error
}

// Open establishes a connection to database collection for this table.
func (ts *TableScan) Open(ctx *ExecutionCtx) error {
	ts.ctx = ctx

	if len(ts.dbName) == 0 {
		ts.dbName = ctx.Db
	}

	if len(ts.dbName) == 0 {
		ts.dbName = ctx.Db
	}

	ts.database = ctx.Schema.Databases[ts.dbName]
	if ts.database == nil {
		return fmt.Errorf("db (%s) doesn't exist - table (%s)", ts.dbName, ts.tableName)
	}

	ts.tableSchema = ts.database.Tables[ts.tableName]
	if ts.tableSchema == nil {
		return fmt.Errorf("table (%s) doesn't exist in db (%s)", ts.tableName, ts.dbName)
	}

	pcs := strings.SplitN(ts.tableSchema.CollectionName, ".", 2)

	db := ctx.Session.DB(pcs[0])
	collection := db.C(pcs[1])
	ts.iter = MgoFindResults{collection.Pipe(ts.pipeline).Iter()}

	return nil
}

func (ts *TableScan) Next(row *Row) bool {
	if ts.iter == nil {
		return false
	}

	var hasNext bool

	for {
		d := &bson.D{}
		hasNext = ts.iter.Next(d)

		if !hasNext {
			break
		}

		values := Values{}
		data := d.Map()

		var err error

		for _, column := range ts.tableSchema.RawColumns {

			value := Value{
				Name: column.SqlName,
				View: column.SqlName,
			}

			if len(column.Name) != 0 {
				value.Data = extractFieldByName(column.Name, data)
			} else {
				value.Data = data[column.SqlName]
			}

			value.Data, err = NewSQLValue(value.Data, column.SqlType)
			if err != nil {
				ts.err = err
				return false
			}

			values = append(values, value)
			delete(data, column.SqlName)
		}

		row.Data = TableRows{{ts.tableName, values}}

		evalCtx := &EvalCtx{Rows{*row}, ts.ctx}

		if ts.matcher != nil {
			m, err := Matches(ts.matcher, evalCtx)
			if err != nil {
				ts.err = err
				return false
			}
			if m {
				break
			}
		} else {
			break
		}
	}

	return hasNext
}

func (ts *TableScan) OpFields() (columns []*Column) {

	for _, c := range ts.tableSchema.RawColumns {
		column := &Column{
			Table: ts.tableName,
			Name:  c.SqlName,
			View:  c.SqlName,
		}
		columns = append(columns, column)
	}

	return columns
}

func (ts *TableScan) Close() error {
	return ts.iter.Close()
}

func (ts *TableScan) Err() error {
	if err := ts.iter.Err(); err != nil {
		return err
	}
	return ts.err
}
