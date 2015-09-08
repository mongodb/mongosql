package planner

import (
	"fmt"
	"github.com/erh/mongo-sql-temp/config"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"sync"
)

type TableScan struct {
	dbName         string
	tableName      string
	fullCollection string
	filter         interface{}
	filterMatcher  Matcher
	sync.Mutex
	iter        FindResults
	err         error
	tableConfig *config.TableConfig
}

// Open establishes a connection to database collection for this table.
func (ts *TableScan) Open(ctx *ExecutionCtx) error {
	return ts.init(ctx)
}

func (ts *TableScan) init(ctx *ExecutionCtx) error {
	sp, err := NewSessionProvider(ctx.Config)
	if err != nil {
		return err
	}

	if len(ts.dbName) == 0 {
		ts.dbName = ctx.Db
	}

	dbConfig := ctx.Config.Schemas[ts.dbName]
	if dbConfig == nil {
		if strings.ToLower(ts.dbName) == "information_schema" {
			var cds ConfigDataSource
			if strings.ToLower(ts.tableName) == "columns" {
				cds = ConfigDataSource{ctx.Config, true}
			} else if strings.ToLower(ts.tableName) == "tables" ||
				strings.ToLower(ts.tableName) == "txxxables" {
				cds = ConfigDataSource{ctx.Config, false}
			} else {
				return fmt.Errorf("unknown information_schema table (%s)", ts.tableName)
			}

			ts.iter = cds.Find(ts.filterMatcher).Iter()
		} else {
			return fmt.Errorf("db (%s) doesn't exist table(%s)", ts.dbName, ts.tableName)
		}
	} else {
		ts.tableConfig = dbConfig.Tables[ts.tableName]
		if ts.tableConfig == nil {
			return fmt.Errorf("table (%s) doesn't exist in db (%s)", ts.tableName, ts.dbName)
		}

		ts.fullCollection = ts.tableConfig.Collection
		pcs := strings.SplitN(ts.tableConfig.Collection, ".", 2)
		collection := sp.GetSession().DB(pcs[0]).C(pcs[1])
		ts.iter = MgoFindResults{collection.Find(ts.filter).Iter()}
	}

	return nil
}

func (ts *TableScan) Next(row *Row) bool {
	if ts.iter == nil {
		return false
	}
	data := &bson.D{}
	hasNext := ts.iter.Next(data)
	row.Data = []TableRow{{ts.tableName, *data, ts.tableConfig}}

	if !hasNext {
		ts.err = ts.iter.Err()
	}

	return hasNext
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
