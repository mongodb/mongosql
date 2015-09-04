package planner

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/erh/mongo-sql-temp/config"
)

type FindResults interface {
	Next(result *bson.M) bool
	Err() error
}

type FindQuery interface {
	Iter() FindResults
}

type DataSource interface {
	Find(query interface{}) FindQuery
	Insert(docs ...interface{}) error
	DropCollection() error

	GetColumns() []config.Column
}

// ------

type MgoFindResults struct {
	iter *mgo.Iter
}

func (gfr *MgoFindResults) Next(result *bson.M) bool {
	return gfr.iter.Next(result)
}

func (gfr *MgoFindResults) Err() error {
	return gfr.iter.Err()
}

// -

type MgoFindQuery struct {
	query *mgo.Query
}

func (gfq MgoFindQuery) Iter() FindResults {
	return &MgoFindResults{gfq.query.Iter()}
}

// -

type MgoDataSource struct {
	Collection *mgo.Collection
	Columns    []config.Column
}

func (gds MgoDataSource) Find(query interface{}) FindQuery {
	return MgoFindQuery{gds.Collection.Find(query)}
}

func (gds MgoDataSource) Insert(docs ...interface{}) error {
	return gds.Collection.Insert(docs...)
}

func (gds MgoDataSource) DropCollection() error {
	return gds.Collection.DropCollection()
}

func (gds MgoDataSource) GetColumns() []config.Column {
	return gds.Columns
}
