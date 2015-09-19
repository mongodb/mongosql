package planner

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type FindResults interface {
	Next(result *bson.D) bool
	Err() error
	Close() error
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

type MgoFindResults struct {
	iter *mgo.Iter
}

func (gfr MgoFindResults) Next(result *bson.D) bool {
	return gfr.iter.Next(result)
}

func (gfr MgoFindResults) Err() error {
	return gfr.iter.Err()
}

func (gfr MgoFindResults) Close() error {
	return gfr.iter.Close()
}

// -------

type MgoFindQuery struct {
	query *mgo.Query
}

func (gfq MgoFindQuery) Iter() FindResults {
	return &MgoFindResults{gfq.query.Iter()}
}

// -------

type MgoDataSource struct {
	Collection *mgo.Collection
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

// ------

type EmptyFindResults struct {
}

func (gfr EmptyFindResults) Next(result *bson.D) bool {
	return false
}

func (gfr EmptyFindResults) Err() error {
	return nil
}

func (gfr EmptyFindResults) Close() error {
	return nil
}

