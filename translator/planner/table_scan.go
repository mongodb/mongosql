package planner

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"sync"
)

type TableScan struct {
	database   *mgo.Database
	collection string
	filter     interface{}
	sync.Mutex
	iter           *mgo.Iter
	IncludeColumns bool
	err            error
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
	ts.database = sp.GetSession().DB(ctx.Db)
	collection := ts.database.C(ts.collection)
	ts.iter = collection.Find(ts.filter).Iter()
	return nil
}

func (ts *TableScan) Next(row *Row) bool {
	data := &bson.D{}
	hasNext := ts.iter.Next(data)
	row.Data = []TableRow{{ts.collection, *data}}

	if !hasNext {
		ts.err = ts.iter.Err()
	}

	return hasNext
}

func (ts *TableScan) Close() error {
	return ts.iter.Close()
}

func (ts *TableScan) Err() error {
	var err error
	ts.Lock()
	err = ts.err
	ts.Unlock()
	return err
}
