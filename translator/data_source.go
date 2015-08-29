package translator

import (
	"gopkg.in/mgo.v2"
)

type FindResults interface {
	Next(result interface{}) bool
	Err() error
}

type FindQuery interface {
	Iter() FindResults
}

type DataSource interface {
	Find(query interface{}) FindQuery
	Insert(docs ...interface{}) error
	DropCollection() error
}

// ------

type GoFindResults struct {
	iter *mgo.Iter
}

func (gfr GoFindResults) Next(result interface{}) bool {
	return gfr.iter.Next(result)
}

func (gfr GoFindResults) Err() error {
	return gfr.iter.Err()
}

// -

type GoFindQuery struct {
	query *mgo.Query
}

func (gfq GoFindQuery) Iter() FindResults {
	return GoFindResults{gfq.query.Iter()}
}

// -

type GoDataSource struct {
	collection *mgo.Collection
}

func (gds GoDataSource) Find(query interface{}) FindQuery {
	return GoFindQuery{gds.collection.Find(query)}
}

func (gds GoDataSource) Insert(docs ...interface{}) error {
	return gds.collection.Insert(docs...)
}

func (gds GoDataSource) DropCollection() error {
	return gds.collection.DropCollection()
}
