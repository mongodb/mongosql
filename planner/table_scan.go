package planner

import (
	"fmt"
	"github.com/erh/mongo-sql-temp/config"
	"github.com/erh/mongo-sql-temp/evaluator"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"sync"
)

type TableScan struct {
	dbName         string
	tableName      string
	fullCollection string
	filter         interface{}
	matcher        evaluator.Matcher
	sync.Mutex
	iter        FindResults
	err         error
	dbConfig    *config.Schema
	tableConfig *config.TableConfig
}

// Open establishes a connection to database collection for this table.
func (ts *TableScan) Open(ctx *ExecutionCtx) error {
	return ts.init(ctx)
}

func (ts *TableScan) init(ctx *ExecutionCtx) error {

	if len(ts.dbName) == 0 {
		ts.dbName = ctx.Db
	}

	if err := ts.setIterator(ctx); err != nil {
		return err
	}

	return nil
}

func (ts *TableScan) setIterator(ctx *ExecutionCtx) error {
	sp, err := evaluator.NewSessionProvider(ctx.Config)
	if err != nil {
		return err
	}

	if len(ts.dbName) == 0 {
		ts.dbName = ctx.Db
	}

	ts.dbConfig = ctx.Config.Schemas[ts.dbName]
	if ts.dbConfig == nil {
		return fmt.Errorf("db (%s) doesn't exist - table (%s)", ts.dbName, ts.tableName)
	}

	ts.tableConfig = ts.dbConfig.Tables[ts.tableName]
	if ts.tableConfig == nil {
		return fmt.Errorf("table (%s) doesn't exist in db (%s)", ts.tableName, ts.dbName)
	}

	ts.fullCollection = ts.tableConfig.Collection
	pcs := strings.SplitN(ts.tableConfig.Collection, ".", 2)
	collection := sp.GetSession().DB(pcs[0]).C(pcs[1])
	ts.iter = MgoFindResults{collection.Find(ts.filter).Iter()}
	return nil
}

func (ts *TableScan) Next(row *evaluator.Row) bool {
	if ts.iter == nil {
		return false
	}

	var hasNext bool

	for {
		d := &bson.D{}
		hasNext = ts.iter.Next(d)

		values := evaluator.Values{}
		data := d.Map()

		for _, column := range ts.tableConfig.Columns {
			value := evaluator.Value{
				Name: column.Name,
				View: column.Name,
				Data: data[column.Name],
			}
			values = append(values, value)
			delete(data, column.Name)
		}

		// now add all other columns
		for key, value := range data {
			value := evaluator.Value{
				Name: key,
				View: key,
				Data: value,
			}
			values = append(values, value)
		}
		row.Data = []evaluator.TableRow{{ts.tableName, values, ts.tableConfig}}

		evalCtx := &evaluator.EvalCtx{[]evaluator.Row{*row}}

		if ts.matcher != nil {
			if ts.matcher.Matches(evalCtx) {
				break
			}
		} else {
			break
		}

		if !hasNext {
			break
		}
	}

	if !hasNext {
		ts.err = ts.iter.Err()
	}

	return hasNext
}

func (ts *TableScan) OpFields() []*Column {
	columns := []*Column{}

	// TODO: we currently only return headers from the schema
	// though the actual data is everything that comes from
	// the database
	for _, c := range ts.tableConfig.Columns {
		column := &Column{
			Table: ts.tableName,
			Name:  c.Name,
			View:  c.Name,
		}
		columns = append(columns, column)
	}

	return columns
}

func (ts *TableScan) Close() error {
	if ts.iter == nil {
		return nil
	}
	return ts.iter.Close()
}

func (ts *TableScan) Err() error {
	var err error
	ts.Lock()
	err = ts.err
	ts.Unlock()
	return err
}
